package main

import (
	"encoding/xml"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/privacybydesign/saml_issuer/internal/samltest"
)

func TestSurfconextLogin(t *testing.T) {
	h := newHarness(t, harnessOptions{})

	login := h.get("/surfconext/")
	require.Equal(t, http.StatusOK, login.StatusCode)
	assert.Contains(t, body(t, login), "Log in om attributen te laden")

	h.loginSurfnet("/surfconext/")

	issue := body(t, h.get("/surfconext/"))
	for _, want := range []string{
		"Beschikbare SURFconext attributen", "Instituut", "example.org", "Type", "student",
		"Gebruiker ID", "jdoe", "Volledig ID", "jdoe@example.org", "Volledige naam", "Jane Doe",
		"Voornaam", "Jane", "Achternaam", "Doe", "E-mailadres", "jane@example.org",
		"U bent ingelogd als Jane",
	} {
		assert.Contains(t, issue, want)
	}
	assert.NotContains(t, issue, "member", "only the first affiliation is issued")

	reply := h.get("/surfconext/?output=irma-session")
	assert.Equal(t, http.StatusOK, reply.StatusCode)
	assert.Equal(t, "text/plain", reply.Header.Get("Content-Type"))
	assert.Equal(t, irmaReply, body(t, reply))

	require.Len(t, h.irma.seen(), 1)
	request := h.irma.seen()[0]
	assert.Equal(t, http.MethodPost, request.method)
	assert.Equal(t, "/session", request.path)
	assert.Equal(t, irmaToken, request.authorization)
	assert.Equal(t, "application/json", request.contentType)
	assert.Equal(t, map[string]any{
		"@context": "https://irma.app/ld/request/issuance/v2",
		"credentials": []any{map[string]any{
			"credential": "pbdf.pbdf.surfnet-2",
			"validity":   float64(samltest.Now.AddDate(0, 3, 0).Unix()),
			"attributes": map[string]any{
				"institute": "example.org", "type": "student", "id": "jdoe", "fullid": "jdoe@example.org",
				"fullname": "Jane Doe", "firstname": "Jane", "familyname": "Doe", "email": "jane@example.org",
			},
			"sdJwtBatchSize": float64(200),
		}},
	}, request.body)

	done := h.get("/surfconext/?action=done")
	assert.Equal(t, http.StatusOK, done.StatusCode)
	assert.Contains(t, body(t, done), "SURFconext attributen zijn geladen")
}

func TestEnglishLogin(t *testing.T) {
	h := newHarness(t, harnessOptions{})

	assert.Contains(t, body(t, h.get("/en/surfconext/")), "Load attributes from SURFconext")
	h.loginSurfnet("/en/surfconext/")

	issue := body(t, h.get("/en/surfconext/"))
	assert.Contains(t, issue, "Available SURFconext attributes")
	assert.Contains(t, issue, "You are logged in as Jane")
}

func TestLanguageSwitch(t *testing.T) {
	h := newHarness(t, harnessOptions{})

	page := body(t, h.get("/nl/surfconext/?foo=bar"))

	assert.Contains(t, page, `<a href="/nl/surfconext/?foo=bar" class="active">NL</a>`)
	assert.Contains(t, page, `<a href="/en/surfconext/?foo=bar">EN</a>`)
}

func TestQueryIsEscaped(t *testing.T) {
	h := newHarness(t, harnessOptions{})

	page := body(t, h.get("/nl/surfconext/?x=%22%3E%3Cscript%3Ealert(1)%3C/script%3E"))

	assert.NotContains(t, page, "<script>alert")
	assert.NotContains(t, page, `"><script`)
}

func TestAuthnRequest(t *testing.T) {
	h := newHarness(t, harnessOptions{})

	response := h.get("/surfconext/?action=login")
	target, err := url.Parse(response.Header.Get("Location"))
	require.NoError(t, err)

	query := target.Query()
	assert.Equal(t, http.StatusFound, response.StatusCode)
	assert.NotEmpty(t, query.Get("SAMLRequest"))
	assert.Empty(t, query.Get("Signature"))
	assert.Empty(t, query.Get("SigAlg"))
	assert.Regexp(t, `^_[0-9a-f-]{36}$`, authnRequestID(t, query.Get("SAMLRequest")))
}

func TestSessionCookieFlags(t *testing.T) {
	h := newHarness(t, harnessOptions{})

	response := h.get("/surfconext/?action=login")

	cookies := response.Cookies()
	require.Len(t, cookies, 1)
	cookie := cookies[0]
	assert.Equal(t, "__Host-session", cookie.Name)
	assert.Equal(t, http.SameSiteNoneMode, cookie.SameSite)
	assert.True(t, cookie.Secure)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, "/", cookie.Path)
	assert.Equal(t, 8*60*60, cookie.MaxAge)
	assert.GreaterOrEqual(t, len(cookie.Value), 43)
}

func TestSessionIDChangesAtLogin(t *testing.T) {
	h := newHarness(t, harnessOptions{})
	before := h.get("/surfconext/?action=login").Cookies()[0].Value
	requestID := authnRequestIDFromJar(t, h)

	response := h.postACS(h.samlResponse(requestID, nil))

	require.Equal(t, http.StatusSeeOther, response.StatusCode)
	after := response.Cookies()
	require.Len(t, after, 1)
	assert.NotEqual(t, before, after[0].Value)
}

// authnRequestIDFromJar starts a login again on the existing session and
// returns the new request ID.
func authnRequestIDFromJar(t *testing.T, h *harness) string {
	t.Helper()

	return h.startSAMLLogin("/surfconext/")
}

func TestResponseCanOnlyBeUsedOnce(t *testing.T) {
	h := newHarness(t, harnessOptions{})
	requestID := h.startSAMLLogin("/surfconext/")
	samlResponse := h.samlResponse(requestID, nil)

	first := h.postACS(samlResponse)
	require.Equal(t, http.StatusSeeOther, first.StatusCode)

	replay := h.postACS(samlResponse)
	assert.Equal(t, http.StatusBadRequest, replay.StatusCode)
	assert.NotContains(t, body(t, replay), "jdoe")
}

func TestReplayOtherSession(t *testing.T) {
	h := newHarness(t, harnessOptions{})
	requestID := h.startSAMLLogin("/surfconext/")
	samlResponse := h.samlResponse(requestID, nil)
	require.Equal(t, http.StatusSeeOther, h.postACS(samlResponse).StatusCode)

	// An attacker with their own session, and their own pending request.
	attacker := newHarnessClient(t, h)
	attackerRequest := attacker.startLogin()

	replay := attacker.postACS(samlResponse)
	assert.Equal(t, http.StatusForbidden, replay.StatusCode)

	// The attacker's pending login is still theirs to complete.
	assert.NotEmpty(t, attackerRequest)
}

func TestUnknownRequest(t *testing.T) {
	h := newHarness(t, harnessOptions{})
	h.startSAMLLogin("/surfconext/")

	response := h.postACS(h.samlResponse("_request-nobody-sent", nil))

	assert.Equal(t, http.StatusForbidden, response.StatusCode)
	assert.NotContains(t, body(t, h.get("/surfconext/")), "jdoe")
}

func TestFailedAttempt(t *testing.T) {
	h := newHarness(t, harnessOptions{})
	requestID := h.startSAMLLogin("/surfconext/")

	junk := h.postACS("not a response")
	require.Equal(t, http.StatusForbidden, junk.StatusCode)

	good := h.postACS(h.samlResponse(requestID, nil))
	assert.Equal(t, http.StatusSeeOther, good.StatusCode)
}

func TestACSWithoutLogin(t *testing.T) {
	h := newHarness(t, harnessOptions{})

	noSession := h.postACS(h.samlResponse("_request-1", nil))
	assert.Equal(t, http.StatusBadRequest, noSession.StatusCode)

	h.get("/surfconext/?action=login")
	h.loginSurfnet("/surfconext/")
	again := h.postACS(h.samlResponse("_request-1", nil))
	assert.Equal(t, http.StatusBadRequest, again.StatusCode)
}

func TestInvalidResponses(t *testing.T) {
	attacker := samltest.NewAttacker(t)
	tests := map[string]func(o *samltest.Options){
		"unsigned":        func(o *samltest.Options) { o.Signing = samltest.SignNothing },
		"signed response": func(o *samltest.Options) { o.Signing = samltest.SignResponseOnly },
		"wrong key":       func(o *samltest.Options) { o.SignedBy = attacker },
		"wrong audience":  func(o *samltest.Options) { o.Audience = "https://other.example.org/sp" },
		"expired":         func(o *samltest.Options) { o.NotOnOrAfter = samltest.Now.Add(-time.Hour) },
		"wrong issuer":    func(o *samltest.Options) { o.Issuer = "https://evil.example.org/idp" },
		"failed status":   func(o *samltest.Options) { o.Status = "urn:oasis:names:tc:SAML:2.0:status:AuthnFailed" },
	}

	for name, change := range tests {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t, harnessOptions{})
			requestID := h.startSAMLLogin("/surfconext/")

			response := h.postACS(h.samlResponse(requestID, change))

			assert.Equal(t, http.StatusForbidden, response.StatusCode)
			page := body(t, response)
			assert.Contains(t, page, "Inloggen met SURFconext mislukt")
			assert.NotContains(t, page, "jdoe")
			assert.Contains(t, body(t, h.get("/surfconext/")), "Log in om attributen te laden")
		})
	}
}

func TestOversizedBodyIsRejected(t *testing.T) {
	h := newHarness(t, harnessOptions{})
	h.startSAMLLogin("/surfconext/")

	response := h.postForm(acsPath, url.Values{"SAMLResponse": {strings.Repeat("A", 2<<20)}})

	assert.Equal(t, http.StatusRequestEntityTooLarge, response.StatusCode)
}

func TestACSOnlyAcceptsPost(t *testing.T) {
	h := newHarness(t, harnessOptions{})

	assert.Equal(t, http.StatusMethodNotAllowed, h.get(acsPath).StatusCode)
}

func TestLogoutEndsTheLocalSession(t *testing.T) {
	h := newHarness(t, harnessOptions{})
	h.loginSurfnet("/surfconext/")

	logout := h.get("/surfconext/?action=logout")
	assert.Equal(t, http.StatusSeeOther, logout.StatusCode)
	assert.Equal(t, "/surfconext/", logout.Header.Get("Location"))

	assert.Contains(t, body(t, h.get("/surfconext/")), "Log in om attributen te laden")
	assert.Equal(t, http.StatusUnauthorized, h.get("/surfconext/?output=irma-session").StatusCode)
	assert.Empty(t, h.irma.seen())
}

func TestLoginWhenLoggedIn(t *testing.T) {
	h := newHarness(t, harnessOptions{})
	h.loginSurfnet("/surfconext/")

	response := h.get("/surfconext/?action=login")

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Contains(t, body(t, response), "Beschikbare SURFconext attributen")
}

func TestSessionEnds(t *testing.T) {
	h := newHarness(t, harnessOptions{})
	h.loginSurfnet("/surfconext/")

	h.clock.Advance(8*time.Hour - time.Second)
	assert.Contains(t, body(t, h.get("/surfconext/")), "Beschikbare SURFconext attributen")

	h.clock.Advance(time.Second)
	assert.Contains(t, body(t, h.get("/surfconext/")), "Log in om attributen te laden")
	assert.Equal(t, http.StatusUnauthorized, h.get("/surfconext/?output=irma-session").StatusCode)
}

func TestAttributesAreEscaped(t *testing.T) {
	h := newHarness(t, harnessOptions{})
	requestID := h.startSAMLLogin("/surfconext/")
	response := h.postACS(h.samlResponse(requestID, func(o *samltest.Options) {
		o.Attributes = []samltest.Attribute{
			{Name: "urn:mace:dir:attribute-def:givenName", Values: []string{`<script>alert("x")</script>`}},
			{Name: "urn:mace:dir:attribute-def:mail", Values: []string{`"><img src=x onerror=alert(1)>`}},
		}
	}))
	require.Equal(t, http.StatusSeeOther, response.StatusCode)

	page := body(t, h.get("/surfconext/"))

	assert.NotContains(t, page, "<script>alert")
	assert.NotContains(t, page, "<img src=x")
	assert.Contains(t, page, "&lt;script&gt;alert(&#34;x&#34;)&lt;/script&gt;")
}

func TestNoAttributesPage(t *testing.T) {
	h := newHarness(t, harnessOptions{})
	requestID := h.startSAMLLogin("/surfconext/")
	require.Equal(t, http.StatusSeeOther, h.postACS(h.samlResponse(requestID, func(o *samltest.Options) {
		o.Attributes = nil
	})).StatusCode)

	assert.Contains(t, body(t, h.get("/surfconext/")), "geen attributen zijn ontvangen")
	assert.Equal(t, http.StatusConflict, h.get("/surfconext/?output=irma-session").StatusCode)
	assert.Empty(t, h.irma.seen())
}

func TestIRMAFailure(t *testing.T) {
	h := newHarness(t, harnessOptions{})
	h.loginSurfnet("/surfconext/")
	h.irma.fail(http.StatusInternalServerError)

	response := h.get("/surfconext/?output=irma-session")

	assert.Equal(t, http.StatusBadGateway, response.StatusCode)
	assert.NotContains(t, body(t, response), irmaToken)
}

func TestIRMASessionNeedsALogin(t *testing.T) {
	h := newHarness(t, harnessOptions{})

	assert.Equal(t, http.StatusUnauthorized, h.get("/surfconext/?output=irma-session").StatusCode)
	assert.Empty(t, h.irma.seen())
}

func TestSPMetadata(t *testing.T) {
	h := newHarness(t, harnessOptions{})

	response := h.get("/simplesaml/module.php/saml/sp/metadata/surfnet")

	require.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "application/samlmetadata+xml", response.Header.Get("Content-Type"))

	var metadata struct {
		EntityID string `xml:"entityID,attr"`
		SPSSO    struct {
			Keys []struct {
				Use  string `xml:"use,attr"`
				Cert string `xml:"KeyInfo>X509Data>X509Certificate"`
			} `xml:"KeyDescriptor"`
			ACS []struct {
				Binding  string `xml:"Binding,attr"`
				Location string `xml:"Location,attr"`
			} `xml:"AssertionConsumerService"`
			SLO []struct{} `xml:"SingleLogoutService"`
		} `xml:"SPSSODescriptor"`
	}
	require.NoError(t, xml.Unmarshal([]byte(body(t, response)), &metadata))

	assert.Equal(t, h.baseURL()+"/simplesaml/module.php/saml/sp/metadata/surfnet", metadata.EntityID)
	require.Len(t, metadata.SPSSO.ACS, 1)
	assert.Equal(t, "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST", metadata.SPSSO.ACS[0].Binding)
	assert.Equal(t, h.baseURL()+"/simplesaml/module.php/saml/sp/saml2-acs.php/surfnet", metadata.SPSSO.ACS[0].Location)
	require.Len(t, metadata.SPSSO.Keys, 1)
	assert.Equal(t, "signing", metadata.SPSSO.Keys[0].Use)
	assert.Equal(t, h.srv.cfg.spCert.Raw, mustDecode(t, metadata.SPSSO.Keys[0].Cert))
	assert.Empty(t, metadata.SPSSO.SLO, "logout is local only")
}

func TestOldPathsStillWork(t *testing.T) {
	h := newHarness(t, harnessOptions{})

	for _, path := range []string{"/surfconext/", "/nl/surfconext/", "/en/surfconext/", "/healthy.html", "/"} {
		assert.Equal(t, http.StatusOK, h.get(path).StatusCode, path)
	}

	for _, prefix := range []string{"/nl/", "/en/"} {
		response := h.get(prefix)
		assert.Equal(t, http.StatusMovedPermanently, response.StatusCode)
		assert.Equal(t, prefix+"surfconext/", response.Header.Get("Location"))
	}

	slashless := h.get("/surfconext")
	assert.Contains(t, []int{http.StatusMovedPermanently, http.StatusTemporaryRedirect}, slashless.StatusCode)
	assert.Equal(t, "/surfconext/", slashless.Header.Get("Location"))
}

func TestProbesGetA200(t *testing.T) {
	h := newHarness(t, harnessOptions{})

	for _, method := range []string{http.MethodGet, http.MethodHead} {
		response := h.do(method, "/", nil, "")
		assert.Equal(t, http.StatusOK, response.StatusCode, method)
	}
}

func TestNoLinkedInWhenNotSet(t *testing.T) {
	h := newHarness(t, harnessOptions{})

	assert.Equal(t, http.StatusNotFound, h.get("/linkedin/").StatusCode)
	assert.Equal(t, http.StatusNotFound, h.get(linkbackPath).StatusCode)
}

func TestNosniffOnEveryResponse(t *testing.T) {
	h := newHarness(t, harnessOptions{})

	for _, path := range []string{"/", "/surfconext/", "/css/style.css", "/nope"} {
		assert.Equal(t, "nosniff", h.get(path).Header.Get("X-Content-Type-Options"), path)
	}
}

func TestDynamicPagesAreNotCached(t *testing.T) {
	h := newHarness(t, harnessOptions{})
	h.loginSurfnet("/surfconext/")

	for _, path := range []string{"/surfconext/", "/surfconext/?output=irma-session"} {
		assert.Equal(t, "no-store", h.get(path).Header.Get("Cache-Control"), path)
	}
}
