package main

import (
	"bytes"
	"compress/flate"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io"
	"math/big"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/beevik/etree"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"

	"github.com/privacybydesign/saml_issuer/internal/samltest"
)

const (
	irmaToken    = "test-irma-token"
	linkedInID   = "linkedin-client-id"
	linkedInPass = "linkedin-client-secret"
)

// harness runs the whole service over TLS, with a fake identity provider, IRMA
// server and (optionally) LinkedIn.
type harness struct {
	t      *testing.T
	server *httptest.Server
	srv    *server
	client *http.Client
	idp    *samltest.IdP
	clock  *clockwork.FakeClock
	irma   *fakeIRMA
	oidc   *fakeOIDC
	env    map[string]string
}

type harnessOptions struct {
	linkedIn bool
}

func newHarness(t *testing.T, opts harnessOptions) *harness {
	t.Helper()

	h := &harness{
		t:     t,
		idp:   samltest.NewIdP(t),
		clock: clockwork.NewFakeClockAt(samltest.Now),
		irma:  newFakeIRMA(t),
	}

	h.server = httptest.NewUnstartedServer(nil)
	host := h.server.Listener.Addr().String()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, spCertFile), pemCertificate(samltest.NewCertificate(t)))
	writeFile(t, filepath.Join(dir, idpMetadataFile), string(h.idp.MetadataXML()))

	h.env = map[string]string{
		"HOST_NAME":                  host,
		"IRMA_SERVER_URL":            h.irma.server.URL,
		"API_TOKEN":                  irmaToken,
		"IRMA_ISSUER_ID":             "pbdf.pbdf",
		"SIMPLESAMLPHP_CERT_DIR":     dir,
		"SIMPLESAMLPHP_METADATA_DIR": dir,
	}
	if opts.linkedIn {
		h.oidc = newFakeOIDC(t)
		h.env["LINKEDIN_CLIENT_ID"] = linkedInID
		h.env["LINKEDIN_CLIENT_SECRET"] = linkedInPass
	}

	cfg, err := loadConfig(func(name string) string { return h.env[name] })
	require.NoError(t, err)
	if h.oidc != nil {
		cfg.linkedInIssuer = h.oidc.server.URL
	}

	h.srv, err = newServer(cfg, h.clock)
	require.NoError(t, err)

	h.server.Config.Handler = h.srv.routes()
	h.server.StartTLS()
	t.Cleanup(h.server.Close)

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	h.client = h.server.Client()
	h.client.Jar = jar
	h.client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }

	return h
}

func (h *harness) baseURL() string { return "https://" + h.env["HOST_NAME"] }

func (h *harness) entityID() string { return h.baseURL() + spMetadataPath }

func (h *harness) acsURL() string { return h.baseURL() + acsPath }

func (h *harness) do(method, target string, body io.Reader, contentType string) *http.Response {
	h.t.Helper()

	request, err := http.NewRequest(method, h.baseURL()+target, body)
	require.NoError(h.t, err)
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}

	response, err := h.client.Do(request)
	require.NoError(h.t, err)
	h.t.Cleanup(func() { _ = response.Body.Close() })

	return response
}

func (h *harness) get(target string) *http.Response { return h.do(http.MethodGet, target, nil, "") }

func (h *harness) postForm(target string, values url.Values) *http.Response {
	return h.do(http.MethodPost, target, strings.NewReader(values.Encode()), "application/x-www-form-urlencoded")
}

func body(t *testing.T, response *http.Response) string {
	t.Helper()

	data, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	return string(data)
}

// startSAMLLogin begins a SURFconext login and returns the AuthnRequest ID.
func (h *harness) startSAMLLogin(page string) string {
	h.t.Helper()

	response := h.get(page + "?action=login")
	require.Equal(h.t, http.StatusFound, response.StatusCode)

	target, err := url.Parse(response.Header.Get("Location"))
	require.NoError(h.t, err)
	require.Equal(h.t, h.idp.SSOURL, target.Scheme+"://"+target.Host+target.Path)

	return authnRequestID(h.t, target.Query().Get("SAMLRequest"))
}

// samlResponse returns a signed response for the request, as the identity
// provider's POST binding would carry it.
func (h *harness) samlResponse(requestID string, change func(*samltest.Options)) string {
	opts := samltest.NewOptions(h.idp, h.entityID(), h.acsURL(), requestID)
	opts.Attributes = surfnetAttributes()
	if change != nil {
		change(&opts)
	}

	return h.idp.Build(opts).Encode()
}

func (h *harness) postACS(samlResponse string) *http.Response {
	return h.postForm(acsPath, url.Values{"SAMLResponse": {samlResponse}})
}

// loginSurfnet runs a complete SURFconext login on the page.
func (h *harness) loginSurfnet(page string) {
	h.t.Helper()

	requestID := h.startSAMLLogin(page)
	response := h.postACS(h.samlResponse(requestID, nil))
	require.Equal(h.t, http.StatusSeeOther, response.StatusCode, body(h.t, response))
	require.Equal(h.t, page, response.Header.Get("Location"))
}

func surfnetAttributes() []samltest.Attribute {
	return []samltest.Attribute{
		{Name: "urn:mace:terena.org:attribute-def:schacHomeOrganization", Values: []string{"example.org"}},
		{Name: "urn:mace:dir:attribute-def:eduPersonAffiliation", Values: []string{"student", "member"}},
		{Name: "urn:mace:dir:attribute-def:uid", Values: []string{"jdoe"}},
		{Name: "urn:mace:dir:attribute-def:eduPersonPrincipalName", Values: []string{"jdoe@example.org"}},
		{Name: "urn:mace:dir:attribute-def:displayName", Values: []string{"Jane Doe"}},
		{Name: "urn:mace:dir:attribute-def:givenName", Values: []string{"Jane"}},
		{Name: "urn:mace:dir:attribute-def:sn", Values: []string{"Doe"}},
		{Name: "urn:mace:dir:attribute-def:mail", Values: []string{"jane@example.org"}},
	}
}

// authnRequestID reads the ID from a deflated, base64 encoded AuthnRequest.
func authnRequestID(t *testing.T, encoded string) string {
	t.Helper()

	compressed, err := base64.StdEncoding.DecodeString(encoded)
	require.NoError(t, err)
	raw, err := io.ReadAll(flate.NewReader(bytes.NewReader(compressed)))
	require.NoError(t, err)

	doc := etree.NewDocument()
	require.NoError(t, doc.ReadFromBytes(raw))
	require.Equal(t, "AuthnRequest", doc.Root().Tag)

	return doc.Root().SelectAttrValue("ID", "")
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
}

func pemCertificate(cert *x509.Certificate) string {
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}))
}

// fakeIRMA records the requests the service makes to the IRMA server.
type fakeIRMA struct {
	server *httptest.Server

	mu       sync.Mutex
	requests []irmaRequest
	status   int
}

type irmaRequest struct {
	method        string
	path          string
	authorization string
	contentType   string
	body          map[string]any
}

const irmaReply = `{"token":"session-token","sessionPtr":{"u":"https://irma.example/irma/session/abc","irmaqr":"issuing"}}`

func newFakeIRMA(t *testing.T) *fakeIRMA {
	t.Helper()

	f := &fakeIRMA{status: http.StatusOK}
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var parsed map[string]any
		data, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(data, &parsed)

		f.mu.Lock()
		defer f.mu.Unlock()
		f.requests = append(f.requests, irmaRequest{
			method:        r.Method,
			path:          r.URL.Path,
			authorization: r.Header.Get("Authorization"),
			contentType:   r.Header.Get("Content-Type"),
			body:          parsed,
		})
		w.WriteHeader(f.status)
		_, _ = io.WriteString(w, irmaReply)
	}))
	t.Cleanup(f.server.Close)

	return f
}

func (f *fakeIRMA) seen() []irmaRequest {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]irmaRequest(nil), f.requests...)
}

func (f *fakeIRMA) fail(status int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.status = status
}

// fakeOIDC is an OpenID Connect provider that signs ID tokens with a
// generated key.
type fakeOIDC struct {
	server *httptest.Server
	// key is published in the JWKS; signingKey signs the ID tokens.
	key        *rsa.PrivateKey
	signingKey *rsa.PrivateKey

	mu       sync.Mutex
	userinfo map[string]any
	idToken  jwt.MapClaims
	codes    map[string]bool
	tokenErr bool
}

func newFakeOIDC(t *testing.T) *fakeOIDC {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	f := &fakeOIDC{key: key, signingKey: key, codes: map[string]bool{"good-code": true}}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /.well-known/openid-configuration", f.discovery)
	mux.HandleFunc("GET /jwks", f.jwks)
	mux.HandleFunc("POST /token", f.token)
	mux.HandleFunc("GET /userinfo", f.userinfoHandler)
	f.server = httptest.NewServer(mux)
	t.Cleanup(f.server.Close)
	f.reset()

	return f
}

// reset restores the claims of a normal LinkedIn user.
func (f *fakeOIDC) reset() {
	f.mu.Lock()
	defer f.mu.Unlock()

	issuer := ""
	if f.server != nil {
		issuer = f.server.URL
	}

	f.tokenErr = false
	f.idToken = jwt.MapClaims{
		"iss": issuer,
		"sub": "linkedin-user-1",
		"aud": linkedInID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	f.userinfo = map[string]any{
		"sub":            "linkedin-user-1",
		"name":           "Jane Doe",
		"given_name":     "Jane",
		"family_name":    "Doe",
		"email":          "jane@example.org",
		"email_verified": true,
		"picture":        "https://media.example/jane.jpg",
	}
}

func (f *fakeOIDC) discovery(w http.ResponseWriter, r *http.Request) {
	_ = json.NewEncoder(w).Encode(map[string]any{
		"issuer":                                f.server.URL,
		"authorization_endpoint":                f.server.URL + "/authorize",
		"token_endpoint":                        f.server.URL + "/token",
		"userinfo_endpoint":                     f.server.URL + "/userinfo",
		"jwks_uri":                              f.server.URL + "/jwks",
		"id_token_signing_alg_values_supported": []string{"RS256"},
	})
}

func (f *fakeOIDC) jwks(w http.ResponseWriter, r *http.Request) {
	_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{
		"kty": "RSA", "use": "sig", "alg": "RS256", "kid": "test-key",
		"n": base64.RawURLEncoding.EncodeToString(f.key.N.Bytes()),
		"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(f.key.E)).Bytes()),
	}}})
}

func (f *fakeOIDC) token(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := r.ParseForm(); err != nil || f.tokenErr || !f.codes[r.PostForm.Get("code")] {
		http.Error(w, `{"error":"invalid_grant"}`, http.StatusBadRequest)
		return
	}

	signed := jwt.NewWithClaims(jwt.SigningMethodRS256, f.idToken)
	signed.Header["kid"] = "test-key"
	idToken, err := signed.SignedString(f.signingKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access_token": "access-token", "token_type": "Bearer", "expires_in": 3600, "id_token": idToken,
	})
}

func (f *fakeOIDC) userinfoHandler(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if r.Header.Get("Authorization") != "Bearer access-token" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(f.userinfo)
}

func (f *fakeOIDC) setIDToken(name string, value any) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.idToken[name] = value
}

func (f *fakeOIDC) setUserinfo(name string, value any) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.userinfo[name] = value
}

func (f *fakeOIDC) failTokenEndpoint() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.tokenErr = true
}

// otherBrowser is a second user: its own cookie jar, the same service.
type otherBrowser struct {
	*harness
}

func newHarnessClient(t *testing.T, h *harness) otherBrowser {
	t.Helper()

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	copied := *h
	client := *h.client
	client.Jar = jar
	copied.client = &client

	return otherBrowser{harness: &copied}
}

func (o otherBrowser) startLogin() string { return o.startSAMLLogin("/surfconext/") }

func mustDecode(t *testing.T, encoded string) []byte {
	t.Helper()

	decoded, err := base64.StdEncoding.DecodeString(strings.Join(strings.Fields(encoded), ""))
	require.NoError(t, err)

	return decoded
}
