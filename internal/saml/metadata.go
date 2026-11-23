package saml

import (
	"encoding/base64"
	"encoding/xml"
	"time"

	saml2 "github.com/russellhaering/gosaml2"
	"github.com/russellhaering/gosaml2/types"
	dsigtypes "github.com/russellhaering/goxmldsig/types"
)

const (
	metadataValidity = 7 * 24 * time.Hour

	langEN = "en"
	langNL = "nl"

	serviceName   = "Yivi credential issuer"
	descriptionEN = "Issue Yivi attributes based on your federated identity"
	descriptionNL = "Verkrijg Yivi attributen gebaseerd op uw federatieve identiteit"

	organizationName  = "Yivi"
	organizationURLEN = "https://yivi.app/en"
	organizationURLNL = "https://yivi.app"
	privacyURLEN      = "https://yivi.app/en/privacy/"
	privacyURLNL      = "https://yivi.app/privacy/"

	logoURL  = "https://raw.githubusercontent.com/privacybydesign/pbdf-schememanager/refs/heads/master/sidn-pbdf/logo.png"
	logoSize = 512

	contactTechnical = "technical"
	contactSupport   = "support"
	// A role address. SURFconext requires the mailto: prefix.
	contactEmail = "mailto:support@yivi.app"

	attrNameFormatURI = "urn:oasis:names:tc:SAML:2.0:attrname-format:uri"

	attrMail                   = "urn:oid:0.9.2342.19200300.100.1.3"
	attrSchacHomeOrganization  = "urn:oid:1.3.6.1.4.1.25178.1.2.9"
	attrEduPersonAffiliation   = "urn:oid:1.3.6.1.4.1.5923.1.1.1.1"
	attrUID                    = "urn:oid:0.9.2342.19200300.100.1.1"
	attrDisplayName            = "urn:oid:2.16.840.1.113730.3.1.241"
	attrEduPersonPrincipalName = "urn:oid:1.3.6.1.4.1.5923.1.1.1.6"
	attrGivenName              = "urn:oid:2.5.4.42"
	attrSurname                = "urn:oid:2.5.4.4"
)

// Localized is a text or URL in one language.
type Localized struct {
	Lang  string `xml:"http://www.w3.org/XML/1998/namespace lang,attr"`
	Value string `xml:",chardata"`
}

// The SP metadata. gosaml2's types have no Organization, ContactPerson, UIInfo
// or AttributeConsumingService. Fields are in the order the schema requires.
type entityDescriptor struct {
	XMLName      xml.Name        `xml:"urn:oasis:names:tc:SAML:2.0:metadata EntityDescriptor"`
	ValidUntil   time.Time       `xml:"validUntil,attr"`
	EntityID     string          `xml:"entityID,attr"`
	SPSSO        spSSODescriptor `xml:"SPSSODescriptor"`
	Organization organization    `xml:"Organization"`
	Contacts     []contactPerson `xml:"ContactPerson"`
}

type spSSODescriptor struct {
	AuthnRequestsSigned        bool                      `xml:"AuthnRequestsSigned,attr"`
	WantAssertionsSigned       bool                      `xml:"WantAssertionsSigned,attr"`
	ProtocolSupportEnumeration string                    `xml:"protocolSupportEnumeration,attr"`
	UIInfo                     uiInfo                    `xml:"Extensions>UIInfo"`
	KeyDescriptors             []types.KeyDescriptor     `xml:"KeyDescriptor"`
	NameIDFormats              []string                  `xml:"NameIDFormat"`
	AssertionConsumerServices  []types.IndexedEndpoint   `xml:"AssertionConsumerService"`
	AttributeConsumingService  attributeConsumingService `xml:"AttributeConsumingService"`
}

type uiInfo struct {
	XMLName              xml.Name    `xml:"urn:oasis:names:tc:SAML:metadata:ui UIInfo"`
	DisplayNames         []Localized `xml:"DisplayName"`
	Descriptions         []Localized `xml:"Description"`
	InformationURLs      []Localized `xml:"InformationURL"`
	PrivacyStatementURLs []Localized `xml:"PrivacyStatementURL"`
	Logo                 logo        `xml:"Logo"`
}

type logo struct {
	Height int    `xml:"height,attr"`
	Width  int    `xml:"width,attr"`
	URL    string `xml:",chardata"`
}

type attributeConsumingService struct {
	Index        int                  `xml:"index,attr"`
	Names        []Localized          `xml:"ServiceName"`
	Descriptions []Localized          `xml:"ServiceDescription"`
	Requested    []requestedAttribute `xml:"RequestedAttribute"`
}

type requestedAttribute struct {
	Name       string `xml:"Name,attr"`
	NameFormat string `xml:"NameFormat,attr"`
	IsRequired bool   `xml:"isRequired,attr,omitempty"`
}

type organization struct {
	Names        []Localized `xml:"OrganizationName"`
	DisplayNames []Localized `xml:"OrganizationDisplayName"`
	URLs         []Localized `xml:"OrganizationURL"`
}

type contactPerson struct {
	Type  string `xml:"contactType,attr"`
	Email string `xml:"EmailAddress"`
}

func bilingual(en, nl string) []Localized {
	return []Localized{{langEN, en}, {langNL, nl}}
}

// Metadata returns the SP metadata. SURFconext reads it from the entity ID's
// URL and publishes it in eduGAIN, which requires the organization and both
// contacts: https://servicedesk.surf.nl/wiki/spaces/IAM/pages/128910147
//
// There is no encryption key: encrypted assertions are rejected.
func (sp *ServiceProvider) Metadata() ([]byte, error) {
	names := bilingual(serviceName, serviceName)
	descriptions := bilingual(descriptionEN, descriptionNL)

	// The attributes the credential is built from. Only mail is required.
	requested := []requestedAttribute{{Name: attrMail, NameFormat: attrNameFormatURI, IsRequired: true}}
	for _, name := range []string{
		attrSchacHomeOrganization,
		attrEduPersonAffiliation,
		attrUID,
		attrDisplayName,
		attrEduPersonPrincipalName,
		attrGivenName,
		attrSurname,
	} {
		requested = append(requested, requestedAttribute{Name: name, NameFormat: attrNameFormatURI})
	}

	desc := entityDescriptor{
		ValidUntil: sp.cfg.Clock.Now().UTC().Add(metadataValidity).Truncate(time.Second),
		EntityID:   sp.cfg.EntityID,
		SPSSO: spSSODescriptor{
			WantAssertionsSigned:       true,
			ProtocolSupportEnumeration: saml2.SAMLProtocolNamespace,
			UIInfo: uiInfo{
				DisplayNames:         names,
				Descriptions:         descriptions,
				InformationURLs:      sp.cfg.InfoURLs,
				PrivacyStatementURLs: bilingual(privacyURLEN, privacyURLNL),
				Logo:                 logo{Height: logoSize, Width: logoSize, URL: logoURL},
			},
			KeyDescriptors: []types.KeyDescriptor{{
				Use: keyUseSigning,
				KeyInfo: dsigtypes.KeyInfo{X509Data: dsigtypes.X509Data{
					X509Certificates: []dsigtypes.X509Certificate{{
						Data: base64.StdEncoding.EncodeToString(sp.cfg.Cert.Raw),
					}},
				}},
			}},
			NameIDFormats: []string{saml2.NameIdFormatTransient},
			AssertionConsumerServices: []types.IndexedEndpoint{{
				Binding:  saml2.BindingHttpPost,
				Location: sp.cfg.ACSURL,
			}},
			AttributeConsumingService: attributeConsumingService{
				Names:        names,
				Descriptions: descriptions,
				Requested:    requested,
			},
		},
		Organization: organization{
			Names:        bilingual(organizationName, organizationName),
			DisplayNames: bilingual(organizationName, organizationName),
			URLs:         bilingual(organizationURLEN, organizationURLNL),
		},
		Contacts: []contactPerson{
			{Type: contactTechnical, Email: contactEmail},
			{Type: contactSupport, Email: contactEmail},
		},
	}

	body, err := xml.MarshalIndent(desc, "", "  ")
	if err != nil {
		return nil, err
	}

	return append([]byte(xml.Header), body...), nil
}
