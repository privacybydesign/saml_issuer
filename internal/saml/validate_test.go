package saml_test

import (
	"bytes"
	"compress/flate"
	"crypto/x509"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/beevik/etree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/privacybydesign/saml_issuer/internal/saml"
	"github.com/privacybydesign/saml_issuer/internal/samltest"
)

const (
	spEntityID = "https://issuer.example.org/simplesaml/module.php/saml/sp/metadata/surfnet"
	acsURL     = "https://issuer.example.org/simplesaml/module.php/saml/sp/saml2-acs.php/surfnet"
	requestID  = "_request-1"

	uidAttr  = "urn:mace:dir:attribute-def:uid"
	mailAttr = "urn:mace:dir:attribute-def:mail"
)

type fixture struct {
	idp      *samltest.IdP
	attacker *samltest.IdP
	sp       *saml.ServiceProvider
}

func newFixture(t *testing.T) *fixture {
	t.Helper()

	idp := samltest.NewIdP(t)
	sp, err := saml.New(saml.Config{
		EntityID: spEntityID,
		ACSURL:   acsURL,
		Cert:     samltest.NewCertificate(t),
		IdPs: []saml.IdP{{
			EntityID: idp.EntityID,
			SSOURL:   idp.SSOURL,
			Certs:    []*x509.Certificate{idp.Cert},
		}},
		Clock: samltest.Clock(),
	})
	require.NoError(t, err)

	return &fixture{idp: idp, attacker: samltest.NewAttacker(t), sp: sp}
}

func (f *fixture) options() samltest.Options {
	return samltest.NewOptions(f.idp, spEntityID, acsURL, requestID)
}

func (f *fixture) build(opts samltest.Options) *samltest.Message {
	return f.idp.Build(opts)
}

func (f *fixture) validate(m *samltest.Message) (saml.Attributes, error) {
	return f.sp.Validate(f.idp.EntityID, m.Encode(), requestID)
}

func (f *fixture) validateRaw(raw []byte) (saml.Attributes, error) {
	return f.sp.Validate(f.idp.EntityID, base64.StdEncoding.EncodeToString(raw), requestID)
}

func TestValidResponse(t *testing.T) {
	for name, signing := range map[string]samltest.Signing{
		"signed assertion":              samltest.SignAssertion,
		"signed response and assertion": samltest.SignResponseAndAssertion,
	} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			opts := f.options()
			opts.Signing = signing
			opts.Attributes = []samltest.Attribute{
				{Name: uidAttr, Values: []string{"jdoe"}},
				{Name: mailAttr, Values: []string{"j.doe@example.org", "second@example.org"}},
			}

			attributes, err := f.validate(f.build(opts))

			require.NoError(t, err)
			assert.Equal(t, saml.Attributes{
				uidAttr:  {"jdoe"},
				mailAttr: {"j.doe@example.org", "second@example.org"},
			}, attributes)
		})
	}
}

func TestNameIDAttributeValue(t *testing.T) {
	f := newFixture(t)
	m := f.build(f.options())

	// An attribute whose value is a NameID element, as eduPersonTargetedID is.
	value := m.Assertion.FindElement("./AttributeStatement/Attribute/AttributeValue")
	value.SetText("")
	value.CreateElement("saml:NameID").SetText("targeted-id")
	samltest.RemoveSignature(m.Assertion)
	f.idp.Sign(m.Assertion)

	attributes, err := f.validate(m)

	require.NoError(t, err)
	assert.Equal(t, []string{"targeted-id"}, attributes[uidAttr])
}

func TestMissingAttributeStatement(t *testing.T) {
	f := newFixture(t)
	opts := f.options()
	opts.Attributes = nil

	attributes, err := f.validate(f.build(opts))

	require.NoError(t, err)
	assert.Empty(t, attributes)
}

func TestRejectsUnsignedAssertion(t *testing.T) {
	for name, signing := range map[string]samltest.Signing{
		"nothing signed":                  samltest.SignNothing,
		"signed response, bare assertion": samltest.SignResponseOnly,
	} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			opts := f.options()
			opts.Signing = signing

			attributes, err := f.validate(f.build(opts))

			require.ErrorIs(t, err, saml.ErrUnsigned)
			assert.Nil(t, attributes)
		})
	}
}

func TestRejectsTamperedAssertion(t *testing.T) {
	f := newFixture(t)
	m := f.build(f.options())
	samltest.SetAttribute(m.Assertion, uidAttr, "admin")

	_, err := f.validate(m)

	require.ErrorIs(t, err, saml.ErrVerification)
}

func TestRejectsTwoAssertions(t *testing.T) {
	for name, sign := range map[string]bool{"second unsigned": false, "second signed": true} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			m := f.build(f.options())
			second := m.Assertion.Copy()
			samltest.RemoveSignature(second)
			second.CreateAttr("ID", "_assertion-2")
			if sign {
				f.idp.Sign(second)
			}
			m.Response.AddChild(second)

			_, err := f.validate(m)

			require.ErrorIs(t, err, saml.ErrMalformed)
		})
	}
}

func TestAssertionElsewhere(t *testing.T) {
	f := newFixture(t)
	m := f.build(f.options())
	extensions := etree.NewElement("samlp:Extensions")
	m.Response.RemoveChild(m.Assertion)
	extensions.AddChild(m.Assertion)
	m.Response.InsertChildAt(1, extensions)

	_, err := f.validate(m)

	require.ErrorIs(t, err, saml.ErrMalformed)
}

func TestRejectsEncryptedAssertion(t *testing.T) {
	f := newFixture(t)
	m := f.build(f.options())
	m.Response.CreateElement("saml:EncryptedAssertion")

	_, err := f.validate(m)

	require.ErrorIs(t, err, saml.ErrMalformed)
}

func TestPinnedCertificate(t *testing.T) {
	tests := map[string]func(f *fixture, o *samltest.Options){
		"attacker key, attacker cert in KeyInfo": func(f *fixture, o *samltest.Options) {
			o.SignedBy = f.attacker
		},
		"attacker key, pinned cert in KeyInfo": func(f *fixture, o *samltest.Options) {
			o.SignedBy = f.attacker
			o.KeyInfo = samltest.KeyInfoOther
			o.KeyInfoCert = f.idp.Cert
		},
		"attacker key, no KeyInfo": func(f *fixture, o *samltest.Options) {
			o.SignedBy = f.attacker
			o.KeyInfo = samltest.KeyInfoOmit
		},
		"pinned key, attacker cert in KeyInfo": func(f *fixture, o *samltest.Options) {
			o.KeyInfo = samltest.KeyInfoOther
			o.KeyInfoCert = f.attacker.Cert
		},
	}

	for name, tamper := range tests {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			opts := f.options()
			tamper(f, &opts)

			attributes, err := f.validate(f.build(opts))

			require.ErrorIs(t, err, saml.ErrVerification)
			assert.Nil(t, attributes)
		})
	}
}

func TestNoKeyInfo(t *testing.T) {
	f := newFixture(t)
	opts := f.options()
	opts.KeyInfo = samltest.KeyInfoOmit

	_, err := f.validate(f.build(opts))

	require.NoError(t, err)
}

func TestRejectsWrongValues(t *testing.T) {
	tests := map[string]struct {
		change func(o *samltest.Options)
		want   error
	}{
		"audience":            {func(o *samltest.Options) { o.Audience = "https://other.example.org/sp" }, saml.ErrAudience},
		"response issuer":     {func(o *samltest.Options) { o.Issuer = "https://evil.example.org/idp" }, saml.ErrIssuer},
		"assertion issuer":    {func(o *samltest.Options) { o.AssertionIssuer = "https://evil.example.org/idp" }, saml.ErrIssuer},
		"destination":         {func(o *samltest.Options) { o.Destination = "https://evil.example.org/acs" }, saml.ErrDestination},
		"missing destination": {func(o *samltest.Options) { o.Destination = "" }, saml.ErrDestination},
		"recipient":           {func(o *samltest.Options) { o.Recipient = "https://evil.example.org/acs" }, saml.ErrDestination},
		"status":              {func(o *samltest.Options) { o.Status = "urn:oasis:names:tc:SAML:2.0:status:Requester" }, saml.ErrStatus},
		"unknown request":     {func(o *samltest.Options) { o.InResponseTo = "_someone-elses-request" }, saml.ErrInResponseTo},
		"no request":          {func(o *samltest.Options) { o.InResponseTo = "" }, saml.ErrInResponseTo},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			opts := f.options()
			tt.change(&opts)

			attributes, err := f.validate(f.build(opts))

			require.ErrorIs(t, err, tt.want)
			assert.Nil(t, attributes)
		})
	}
}

func TestRejectsMissingConditions(t *testing.T) {
	tests := map[string]func(m *samltest.Message){
		"conditions": func(m *samltest.Message) {
			m.Assertion.RemoveChild(m.Assertion.FindElement("./Conditions"))
		},
		"audience restriction": func(m *samltest.Message) {
			conditions := m.Assertion.FindElement("./Conditions")
			conditions.RemoveChild(conditions.FindElement("./AudienceRestriction"))
		},
	}

	for name, remove := range tests {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			m := f.build(f.options())
			samltest.RemoveSignature(m.Assertion)
			remove(m)
			f.idp.Sign(m.Assertion)

			_, err := f.validate(m)

			require.Error(t, err)
		})
	}
}

func TestResponseForAnotherRequest(t *testing.T) {
	f := newFixture(t)
	m := f.build(f.options())

	// The response answers _request-1, but the user's session expects _request-2.
	attributes, err := f.sp.Validate(f.idp.EntityID, m.Encode(), "_request-2")

	require.ErrorIs(t, err, saml.ErrInResponseTo)
	assert.Nil(t, attributes)
}

func TestAssertionReplay(t *testing.T) {
	f := newFixture(t)
	m := f.build(f.options())

	// An unsigned Response can be rewritten freely, so only the assertion's own
	// InResponseTo can bind it to a request.
	m.Response.CreateAttr("InResponseTo", "_request-2")
	attributes, err := f.sp.Validate(f.idp.EntityID, m.Encode(), "_request-2")

	require.ErrorIs(t, err, saml.ErrInResponseTo)
	assert.Nil(t, attributes)
}

func TestWrongResponseInResponseTo(t *testing.T) {
	f := newFixture(t)
	m := f.build(f.options())

	// The assertion answers _request-1. Only the unsigned Response says otherwise.
	m.Response.CreateAttr("InResponseTo", "_request-2")
	attributes, err := f.validate(m)

	require.ErrorIs(t, err, saml.ErrInResponseTo)
	assert.Nil(t, attributes)
}

func TestTimeWindow(t *testing.T) {
	skew := saml.ClockSkew
	tests := map[string]struct {
		notBefore    time.Time
		notOnOrAfter time.Time
		want         error
	}{
		"inside":                       {samltest.Now.Add(-time.Minute), samltest.Now.Add(time.Minute), nil},
		"not yet valid, within skew":   {samltest.Now.Add(skew - time.Second), samltest.Now.Add(time.Hour), nil},
		"not yet valid, past skew":     {samltest.Now.Add(skew + time.Second), samltest.Now.Add(time.Hour), saml.ErrNotYetValid},
		"expired, within skew":         {samltest.Now.Add(-time.Hour), samltest.Now.Add(-skew + time.Second), nil},
		"expired at the end of skew":   {samltest.Now.Add(-time.Hour), samltest.Now.Add(-skew), saml.ErrExpired},
		"expired, past skew":           {samltest.Now.Add(-time.Hour), samltest.Now.Add(-skew - time.Second), saml.ErrExpired},
		"expired long ago":             {samltest.Now.Add(-48 * time.Hour), samltest.Now.Add(-24 * time.Hour), saml.ErrExpired},
		"valid only in the far future": {samltest.Now.Add(24 * time.Hour), samltest.Now.Add(48 * time.Hour), saml.ErrNotYetValid},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			opts := f.options()
			opts.NotBefore = tt.notBefore
			opts.NotOnOrAfter = tt.notOnOrAfter
			opts.ConfirmationNotOnOrAfter = tt.notOnOrAfter

			_, err := f.validate(f.build(opts))

			if tt.want == nil {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, tt.want)
		})
	}
}

func TestExpiredConfirmation(t *testing.T) {
	f := newFixture(t)
	opts := f.options()
	opts.ConfirmationNotOnOrAfter = samltest.Now.Add(-time.Hour)

	_, err := f.validate(f.build(opts))

	require.ErrorIs(t, err, saml.ErrExpired)
}

func TestRejectsXMLComments(t *testing.T) {
	f := newFixture(t)
	opts := f.options()
	opts.Attributes = []samltest.Attribute{{Name: mailAttr, Values: []string{"admin@example.org.evil.example"}}}
	m := f.build(opts)

	// Canonicalisation drops comments, so the signature stays valid while a
	// parser that reads the first text node sees "admin@example.org".
	value := m.Assertion.FindElement("./AttributeStatement/Attribute/AttributeValue")
	value.SetText("")
	value.CreateText("admin@example.org")
	value.AddChild(etree.NewComment("injected"))
	value.CreateText(".evil.example")

	attributes, err := f.validate(m)

	require.ErrorIs(t, err, saml.ErrMalformed)
	assert.Nil(t, attributes)
}

func TestRejectsOversizedResponse(t *testing.T) {
	f := newFixture(t)
	opts := f.options()
	opts.Attributes = []samltest.Attribute{{Name: uidAttr, Values: []string{strings.Repeat("a", saml.MaxResponseBytes)}}}

	_, err := f.validate(f.build(opts))

	require.ErrorIs(t, err, saml.ErrMalformed)
}

func TestRejectsUndecodableInput(t *testing.T) {
	f := newFixture(t)
	valid := f.build(f.options()).Bytes()

	var deflated bytes.Buffer
	writer, err := flate.NewWriter(&deflated, flate.DefaultCompression)
	require.NoError(t, err)
	_, err = writer.Write(valid)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	tests := map[string][]byte{
		"empty":        nil,
		"not XML":      []byte("hello"),
		"deflated":     deflated.Bytes(),
		"truncated":    valid[:len(valid)/2],
		"wrong root":   []byte(`<Response xmlns="urn:example"/>`),
		"DTD":          append([]byte(`<!DOCTYPE r [<!ENTITY x "y">]>`), valid...),
		"deep nesting": []byte(strings.Repeat("<a>", 100) + strings.Repeat("</a>", 100)),
	}

	for name, raw := range tests {
		t.Run(name, func(t *testing.T) {
			attributes, err := f.validateRaw(raw)

			require.ErrorIs(t, err, saml.ErrMalformed)
			assert.Nil(t, attributes)
		})
	}

	t.Run("not base64", func(t *testing.T) {
		_, err := f.sp.Validate(f.idp.EntityID, "!!not base64!!", requestID)

		require.ErrorIs(t, err, saml.ErrMalformed)
	})
}

func TestRejectsUnknownIdP(t *testing.T) {
	f := newFixture(t)

	_, err := f.sp.Validate("https://other.example.org/idp", f.build(f.options()).Encode(), requestID)

	require.Error(t, err)
}

func TestRejectsEmptyRequestID(t *testing.T) {
	f := newFixture(t)
	opts := f.options()
	opts.InResponseTo = ""

	_, err := f.sp.Validate(f.idp.EntityID, f.build(opts).Encode(), "")

	require.ErrorIs(t, err, saml.ErrInResponseTo)
}

func TestSignatureReference(t *testing.T) {
	for name, uri := range map[string]string{
		"empty reference":       "",
		"reference to response": "#" + samltest.ResponseID,
		"reference to other ID": "#_elsewhere",
	} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			m := f.build(f.options())
			m.Assertion.FindElement("./Signature/SignedInfo/Reference").CreateAttr("URI", uri)

			_, err := f.validate(m)

			require.ErrorIs(t, err, saml.ErrMalformed)
		})
	}
}

func TestRejectsDuplicateIDs(t *testing.T) {
	f := newFixture(t)
	m := f.build(f.options())
	m.Response.CreateElement("saml:Marker").CreateAttr("ID", samltest.AssertionID)

	_, err := f.validate(m)

	require.ErrorIs(t, err, saml.ErrMalformed)
}
