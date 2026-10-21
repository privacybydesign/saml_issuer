package saml

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"time"

	"github.com/jonboulle/clockwork"
	saml2 "github.com/russellhaering/gosaml2"
	"github.com/russellhaering/gosaml2/types"
	dsig "github.com/russellhaering/goxmldsig"
	dsigtypes "github.com/russellhaering/goxmldsig/types"
)

const (
	// ClockSkew is how far an assertion's times may be off from our clock.
	ClockSkew = 180 * time.Second

	// MaxResponseBytes bounds a decoded SAML response.
	MaxResponseBytes = 128 << 10

	metadataValidity = 7 * 24 * time.Hour
	requestIDAttr    = "ID"
)

// Config describes this service provider and the identity providers it trusts.
type Config struct {
	EntityID string
	ACSURL   string
	// Cert is the signing certificate published in the SP metadata.
	Cert *x509.Certificate
	IdPs []IdP
	// Clock defaults to the system clock.
	Clock clockwork.Clock
}

// ServiceProvider sends AuthnRequests and validates the responses to them.
type ServiceProvider struct {
	cfg  Config
	idps map[string]*idpState
}

type idpState struct {
	idp IdP
	// login builds requests with the real clock; verify checks responses
	// with a clock set back by ClockSkew.
	login  *saml2.SAMLServiceProvider
	verify *saml2.SAMLServiceProvider
}

// AuthnRequest is a request ready to be sent to the identity provider.
type AuthnRequest struct {
	ID  string
	URL string
}

// Attributes maps an attribute name to its values, in document order.
type Attributes map[string][]string

// New builds a service provider for the given identity providers.
func New(cfg Config) (*ServiceProvider, error) {
	if cfg.Clock == nil {
		cfg.Clock = clockwork.NewRealClock()
	}
	if cfg.EntityID == "" || cfg.ACSURL == "" || cfg.Cert == nil {
		return nil, errors.New("entity ID, ACS URL and certificate are required")
	}

	sp := &ServiceProvider{cfg: cfg, idps: map[string]*idpState{}}
	for _, idp := range cfg.IdPs {
		sp.idps[idp.EntityID] = newIdPState(cfg, idp)
	}

	return sp, nil
}

// pastClock runs ClockSkew behind its wrapped clock.
type pastClock struct{ clockwork.Clock }

func (c pastClock) Now() time.Time { return c.Clock.Now().Add(-ClockSkew) }

func newIdPState(cfg Config, idp IdP) *idpState {
	return &idpState{
		idp:   idp,
		login: newLibrary(cfg, idp, cfg.Clock),
		// The library checks SubjectConfirmationData/@NotOnOrAfter without any
		// skew, so it runs behind our clock. Conditions are checked by us.
		verify: newLibrary(cfg, idp, pastClock{cfg.Clock}),
	}
}

func newLibrary(cfg Config, idp IdP, clock clockwork.Clock) *saml2.SAMLServiceProvider {
	return &saml2.SAMLServiceProvider{
		IdentityProviderSSOURL:      idp.SSOURL,
		IdentityProviderIssuer:      idp.EntityID,
		ServiceProviderIssuer:       cfg.EntityID,
		AssertionConsumerServiceURL: cfg.ACSURL,
		AudienceURI:                 cfg.EntityID,
		NameIdFormat:                saml2.NameIdFormatTransient,
		IDPCertificateStore:         &dsig.MemoryX509CertificateStore{Roots: idp.Certs},
		MaximumDecompressedBodySize: MaxResponseBytes,
		Clock:                       dsig.NewFakeClock(clock),
	}
}

func (sp *ServiceProvider) state(idpID string) (*idpState, error) {
	state, ok := sp.idps[idpID]
	if !ok {
		return nil, fmt.Errorf("unknown identity provider %q", idpID)
	}

	return state, nil
}

// Login builds an unsigned AuthnRequest for the HTTP-Redirect binding.
func (sp *ServiceProvider) Login(idpID string) (AuthnRequest, error) {
	state, err := sp.state(idpID)
	if err != nil {
		return AuthnRequest{}, err
	}

	doc, err := state.login.BuildAuthRequestDocumentNoSig()
	if err != nil {
		return AuthnRequest{}, err
	}

	url, err := state.login.BuildAuthURLRedirect("", doc)
	if err != nil {
		return AuthnRequest{}, err
	}

	return AuthnRequest{ID: doc.Root().SelectAttrValue(requestIDAttr, ""), URL: url}, nil
}

// Metadata returns the SP metadata: entity ID, signing certificate and the
// HTTP-POST assertion consumer service.
func (sp *ServiceProvider) Metadata() ([]byte, error) {
	desc := types.EntityDescriptor{
		ValidUntil: sp.cfg.Clock.Now().UTC().Add(metadataValidity).Truncate(time.Second),
		EntityID:   sp.cfg.EntityID,
		SPSSODescriptor: &types.SPSSODescriptor{
			WantAssertionsSigned:       true,
			ProtocolSupportEnumeration: saml2.SAMLProtocolNamespace,
			NameIDFormats:              []string{saml2.NameIdFormatTransient},
			KeyDescriptors: []types.KeyDescriptor{{
				Use: keyUseSigning,
				KeyInfo: dsigtypes.KeyInfo{X509Data: dsigtypes.X509Data{
					X509Certificates: []dsigtypes.X509Certificate{{
						Data: base64.StdEncoding.EncodeToString(sp.cfg.Cert.Raw),
					}},
				}},
			}},
			AssertionConsumerServices: []types.IndexedEndpoint{{
				Binding:  saml2.BindingHttpPost,
				Location: sp.cfg.ACSURL,
			}},
		},
	}

	body, err := xml.MarshalIndent(desc, "", "  ")
	if err != nil {
		return nil, err
	}

	return append([]byte(xml.Header), body...), nil
}
