// Package samltest builds signed SAML responses with generated keys, for
// tests only.
package samltest

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/beevik/etree"
	"github.com/jonboulle/clockwork"
	dsig "github.com/russellhaering/goxmldsig"
)

const (
	protocolNS  = "urn:oasis:names:tc:SAML:2.0:protocol"
	assertionNS = "urn:oasis:names:tc:SAML:2.0:assertion"
	dsigNS      = "http://www.w3.org/2000/09/xmldsig#"

	StatusSuccess = "urn:oasis:names:tc:SAML:2.0:status:Success"
	ResponseID    = "_response-1"
	AssertionID   = "_assertion-1"

	keyBits      = 2048
	attrNameForm = "urn:oasis:names:tc:SAML:2.0:attrname-format:uri"
)

// Now is the instant the fixtures are built around.
var Now = time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)

// Clock returns a clock fixed at Now.
func Clock() clockwork.Clock { return clockwork.NewFakeClockAt(Now) }

// Generating RSA keys is slow, so each slot is generated once per process.
var keys [3]struct {
	once sync.Once
	key  *rsa.PrivateKey
}

func slotKey(t testing.TB, slot int) *rsa.PrivateKey {
	t.Helper()

	keys[slot].once.Do(func() {
		key, err := rsa.GenerateKey(rand.Reader, keyBits)
		if err != nil {
			panic(err)
		}
		keys[slot].key = key
	})

	return keys[slot].key
}

// IdP is a fake identity provider.
type IdP struct {
	EntityID string
	SSOURL   string
	Cert     *x509.Certificate
	key      *rsa.PrivateKey
}

// NewIdP returns the identity provider the service provider trusts.
func NewIdP(t testing.TB) *IdP { return newIdP(t, 0) }

// NewAttacker returns an identity provider with another key that claims the
// same entity ID as NewIdP.
func NewAttacker(t testing.TB) *IdP { return newIdP(t, 1) }

// NewCertificate returns a certificate nobody trusts, for the SP metadata.
func NewCertificate(t testing.TB) *x509.Certificate { return newIdP(t, 2).Cert }

func newIdP(t testing.TB, slot int) *IdP {
	t.Helper()

	key := slotKey(t, slot)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(int64(slot) + 1),
		Subject:      pkix.Name{CommonName: "idp.example.org"},
		NotBefore:    Now.Add(-24 * time.Hour),
		NotAfter:     Now.Add(10 * 365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}

	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}

	return &IdP{
		EntityID: "https://idp.example.org/metadata",
		SSOURL:   "https://idp.example.org/sso",
		Cert:     cert,
		key:      key,
	}
}

// MetadataXML returns IdP metadata as a service provider would be given it.
func (i *IdP) MetadataXML() []byte {
	return fmt.Appendf(nil, `<?xml version="1.0"?>
<md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata" xmlns:ds="%s" entityID="%s">
  <md:IDPSSODescriptor protocolSupportEnumeration="%s">
    <md:KeyDescriptor use="signing">
      <ds:KeyInfo><ds:X509Data><ds:X509Certificate>%s</ds:X509Certificate></ds:X509Data></ds:KeyInfo>
    </md:KeyDescriptor>
    <md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="%s"/>
  </md:IDPSSODescriptor>
</md:EntityDescriptor>
`, dsigNS, i.EntityID, protocolNS, base64.StdEncoding.EncodeToString(i.Cert.Raw), i.SSOURL)
}

type keyStore struct {
	key  *rsa.PrivateKey
	cert []byte
}

func (k keyStore) GetKeyPair() (*rsa.PrivateKey, []byte, error) { return k.key, k.cert, nil }

// Sign adds an enveloped signature, made with the IdP's key, to el.
func (i *IdP) Sign(el *etree.Element) {
	i.signWith(el, i.Cert)
}

func (i *IdP) signWith(el *etree.Element, keyInfoCert *x509.Certificate) {
	ctx := dsig.NewDefaultSigningContext(keyStore{key: i.key, cert: i.Cert.Raw})
	ctx.Canonicalizer = dsig.MakeC14N10ExclusiveCanonicalizerWithPrefixList("")

	sig, err := ctx.ConstructSignature(el, true)
	if err != nil {
		panic(err)
	}

	keyInfo := sig.FindElement("./KeyInfo")
	if keyInfoCert == nil {
		sig.RemoveChild(keyInfo)
	} else {
		keyInfo.FindElement("./X509Data/X509Certificate").SetText(base64.StdEncoding.EncodeToString(keyInfoCert.Raw))
	}

	// The schema puts the signature right after the issuer.
	el.InsertChildAt(1, sig)
}
