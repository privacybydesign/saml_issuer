package main

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/privacybydesign/saml_issuer/internal/samltest"
)

// startLinkedInLogin begins a LinkedIn login and returns the state sent to
// LinkedIn along with the authorization URL.
func (h *harness) startLinkedInLogin(page string) (string, *url.URL) {
	h.t.Helper()

	response := h.get(page + "?action=login")
	require.Equal(h.t, http.StatusFound, response.StatusCode)

	target, err := url.Parse(response.Header.Get("Location"))
	require.NoError(h.t, err)

	return target.Query().Get("state"), target
}

func (h *harness) linkback(values url.Values) *http.Response {
	return h.get(linkbackPath + "?" + values.Encode())
}

func (h *harness) loginLinkedIn(page string) {
	h.t.Helper()

	state, _ := h.startLinkedInLogin(page)
	response := h.linkback(url.Values{"code": {"good-code"}, "state": {state}})
	require.Equal(h.t, http.StatusSeeOther, response.StatusCode, body(h.t, response))
	require.Equal(h.t, page, response.Header.Get("Location"))
}

func TestLinkedInLogin(t *testing.T) {
	h := newHarness(t, harnessOptions{linkedIn: true})
	assert.Contains(t, body(t, h.get("/linkedin/")), "Laad attributen via LinkedIn")

	h.loginLinkedIn("/linkedin/")

	issue := body(t, h.get("/linkedin/"))
	for _, want := range []string{"Volledige naam", "Jane Doe", "Voornaam", "Jane", "Achternaam", "Doe", "E-mailadres", "jane@example.org"} {
		assert.Contains(t, issue, want)
	}
	assert.NotContains(t, issue, "media.example", "claims not issued are not shown")

	require.Equal(t, http.StatusOK, h.get("/linkedin/?output=irma-session").StatusCode)
	require.Len(t, h.irma.seen(), 1)
	assert.Equal(t, map[string]any{
		"@context": "https://irma.app/ld/request/issuance/v2",
		"credentials": []any{map[string]any{
			"credential": "pbdf.pbdf.linkedin",
			"validity":   float64(samltest.Now.AddDate(1, 0, 0).Unix()),
			"attributes": map[string]any{
				"fullname": "Jane Doe", "firstname": "Jane", "familyname": "Doe", "email": "jane@example.org",
			},
			"sdJwtBatchSize": float64(0),
		}},
	}, h.irma.seen()[0].body)
}

func TestLinkedInRedirect(t *testing.T) {
	h := newHarness(t, harnessOptions{linkedIn: true})

	state, target := h.startLinkedInLogin("/en/linkedin/")

	assert.Equal(t, h.oidc.server.URL+"/authorize", target.Scheme+"://"+target.Host+target.Path)
	query := target.Query()
	assert.Equal(t, "code", query.Get("response_type"))
	assert.Equal(t, linkedInID, query.Get("client_id"))
	assert.Equal(t, h.baseURL()+"/simplesaml/module.php/authoauth2/linkback", query.Get("redirect_uri"))
	assert.Equal(t, "openid profile email", query.Get("scope"))
	assert.GreaterOrEqual(t, len(state), 43)
}

func TestLinkedInReturnTo(t *testing.T) {
	h := newHarness(t, harnessOptions{linkedIn: true})

	h.loginLinkedIn("/en/linkedin/")

	assert.Contains(t, body(t, h.get("/en/linkedin/")), "Available LinkedIn attributes")
}

func TestLinkedInBadState(t *testing.T) {
	h := newHarness(t, harnessOptions{linkedIn: true})
	state, _ := h.startLinkedInLogin("/linkedin/")

	for name, values := range map[string]url.Values{
		"other state":   {"code": {"good-code"}, "state": {state + "x"}},
		"no state":      {"code": {"good-code"}},
		"empty state":   {"code": {"good-code"}, "state": {""}},
		"swapped state": {"code": {"good-code"}, "state": {"someone-elses-state"}},
	} {
		response := h.linkback(values)
		assert.Equal(t, http.StatusBadRequest, response.StatusCode, name)
	}

	// None of that burned the real state.
	good := h.linkback(url.Values{"code": {"good-code"}, "state": {state}})
	assert.Equal(t, http.StatusSeeOther, good.StatusCode)
}

func TestLinkedInStateIsSingleUse(t *testing.T) {
	h := newHarness(t, harnessOptions{linkedIn: true})
	state, _ := h.startLinkedInLogin("/linkedin/")
	values := url.Values{"code": {"good-code"}, "state": {state}}
	require.Equal(t, http.StatusSeeOther, h.linkback(values).StatusCode)

	assert.Equal(t, http.StatusBadRequest, h.linkback(values).StatusCode)
}

func TestLinkedInNoSession(t *testing.T) {
	h := newHarness(t, harnessOptions{linkedIn: true})

	response := h.linkback(url.Values{"code": {"good-code"}, "state": {"anything"}})

	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestLinkedInRefused(t *testing.T) {
	h := newHarness(t, harnessOptions{linkedIn: true})
	state, _ := h.startLinkedInLogin("/linkedin/")

	response := h.linkback(url.Values{"error": {"user_cancelled_authorize"}, "state": {state}})

	assert.Equal(t, http.StatusForbidden, response.StatusCode)
	assert.Contains(t, body(t, response), "Inloggen met LinkedIn mislukt")
	assert.Contains(t, body(t, h.get("/linkedin/")), "Laad attributen via LinkedIn")
}

func TestLinkedInRejectsBadTokens(t *testing.T) {
	tests := map[string]func(o *fakeOIDC){
		"unknown code":   func(o *fakeOIDC) { o.codes = map[string]bool{} },
		"token endpoint": func(o *fakeOIDC) { o.failTokenEndpoint() },
		"wrong audience": func(o *fakeOIDC) { o.setIDToken("aud", "someone-else") },
		"wrong issuer":   func(o *fakeOIDC) { o.setIDToken("iss", "https://evil.example.org") },
		"lookalike":      func(o *fakeOIDC) { o.setIDToken("iss", "https://www.linkedin.com.evil.example") },
		"no issuer":      func(o *fakeOIDC) { o.setIDToken("iss", "") },
		"expired":        func(o *fakeOIDC) { o.setIDToken("exp", time.Now().Add(-time.Hour).Unix()) },
		"other subject":  func(o *fakeOIDC) { o.setUserinfo("sub", "linkedin-user-2") },
		"no subject":     func(o *fakeOIDC) { o.setIDToken("sub", "") },
	}

	for name, break_ := range tests {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t, harnessOptions{linkedIn: true})
			break_(h.oidc)
			state, _ := h.startLinkedInLogin("/linkedin/")

			response := h.linkback(url.Values{"code": {"good-code"}, "state": {state}})

			assert.Equal(t, http.StatusForbidden, response.StatusCode)
			assert.Contains(t, body(t, h.get("/linkedin/")), "Laad attributen via LinkedIn")
			assert.Equal(t, http.StatusUnauthorized, h.get("/linkedin/?output=irma-session").StatusCode)
		})
	}
}

// LinkedIn's documentation gives the bare domain as the issuer of its ID
// tokens, while its discovery document says https://www.linkedin.com/oauth.
func TestLinkedInBareDomainIssuer(t *testing.T) {
	h := newHarness(t, harnessOptions{linkedIn: true})
	h.oidc.setIDToken("iss", "https://www.linkedin.com")

	h.loginLinkedIn("/linkedin/")

	assert.Contains(t, body(t, h.get("/linkedin/")), "Beschikbare LinkedIn attributen")
}

func TestLinkedInWrongKey(t *testing.T) {
	h := newHarness(t, harnessOptions{linkedIn: true})
	other := newFakeOIDC(t)
	h.oidc.signingKey = other.key
	state, _ := h.startLinkedInLogin("/linkedin/")

	response := h.linkback(url.Values{"code": {"good-code"}, "state": {state}})

	// The ID token is signed with a key that is not in the provider's JWKS.
	assert.Equal(t, http.StatusForbidden, response.StatusCode)
}

func TestProvidersAreSeparate(t *testing.T) {
	h := newHarness(t, harnessOptions{linkedIn: true})

	h.loginSurfnet("/surfconext/")

	assert.Contains(t, body(t, h.get("/linkedin/")), "Laad attributen via LinkedIn")
	assert.Equal(t, http.StatusUnauthorized, h.get("/linkedin/?output=irma-session").StatusCode)

	h.loginLinkedIn("/linkedin/")
	h.get("/surfconext/?action=logout")
	assert.Contains(t, body(t, h.get("/linkedin/")), "Beschikbare LinkedIn attributen")
	assert.Contains(t, body(t, h.get("/surfconext/")), "Log in om attributen te laden")
}

func TestLinkedInDiscovery(t *testing.T) {
	h := newHarness(t, harnessOptions{linkedIn: true})
	h.oidc.server.Close()

	response := h.get("/linkedin/?action=login")

	assert.Equal(t, http.StatusBadGateway, response.StatusCode)
	assert.Contains(t, body(t, h.get("/surfconext/")), "Log in om attributen te laden", "SURFconext is unaffected")
}
