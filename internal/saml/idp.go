package saml

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"os"
	"strings"

	saml2 "github.com/russellhaering/gosaml2"
	"github.com/russellhaering/gosaml2/types"
)

const (
	keyUseSigning = "signing"
	entitiesTag   = "EntitiesDescriptor"
	entityTag     = "EntityDescriptor"
	metadataNS    = "urn:oasis:names:tc:SAML:2.0:metadata"
)

// IdP is an identity provider read from local metadata. Certs are the only
// keys its responses are verified with.
type IdP struct {
	EntityID string
	SSOURL   string
	Certs    []*x509.Certificate
}

type entitiesDescriptor struct {
	XMLName  xml.Name
	Entities []types.EntityDescriptor `xml:"urn:oasis:names:tc:SAML:2.0:metadata EntityDescriptor"`
}

// LoadIdPs reads SAML metadata (one EntityDescriptor or an EntitiesDescriptor)
// and returns every identity provider in it.
func LoadIdPs(path string) ([]IdP, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return parseIdPs(data)
}

func parseIdPs(data []byte) ([]IdP, error) {
	var root entitiesDescriptor
	if err := xml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("parse metadata: %w", err)
	}

	var entities []types.EntityDescriptor
	switch {
	case root.XMLName.Space == metadataNS && root.XMLName.Local == entitiesTag:
		entities = root.Entities
	case root.XMLName.Space == metadataNS && root.XMLName.Local == entityTag:
		var single types.EntityDescriptor
		if err := xml.Unmarshal(data, &single); err != nil {
			return nil, fmt.Errorf("parse metadata: %w", err)
		}
		entities = []types.EntityDescriptor{single}
	default:
		return nil, fmt.Errorf("metadata root is %s, want %s or %s", root.XMLName.Local, entityTag, entitiesTag)
	}

	var idps []IdP
	for _, entity := range entities {
		if entity.IDPSSODescriptor == nil {
			continue
		}

		idp, err := idpFromEntity(entity)
		if err != nil {
			return nil, fmt.Errorf("entity %s: %w", entity.EntityID, err)
		}
		idps = append(idps, idp)
	}

	if len(idps) == 0 {
		return nil, fmt.Errorf("metadata contains no identity provider")
	}

	return idps, nil
}

func idpFromEntity(entity types.EntityDescriptor) (IdP, error) {
	idp := IdP{EntityID: entity.EntityID}

	for _, sso := range entity.IDPSSODescriptor.SingleSignOnServices {
		if sso.Binding == saml2.BindingHttpRedirect {
			idp.SSOURL = sso.Location
			break
		}
	}
	if idp.SSOURL == "" {
		return IdP{}, fmt.Errorf("no HTTP-Redirect single sign-on service")
	}

	for _, key := range entity.IDPSSODescriptor.KeyDescriptors {
		if key.Use != "" && key.Use != keyUseSigning {
			continue
		}

		for _, encoded := range key.KeyInfo.X509Data.X509Certificates {
			der, err := base64.StdEncoding.DecodeString(strings.Join(strings.Fields(encoded.Data), ""))
			if err != nil {
				return IdP{}, fmt.Errorf("decode certificate: %w", err)
			}

			cert, err := x509.ParseCertificate(der)
			if err != nil {
				return IdP{}, fmt.Errorf("parse certificate: %w", err)
			}
			idp.Certs = append(idp.Certs, cert)
		}
	}
	if len(idp.Certs) == 0 {
		return IdP{}, fmt.Errorf("no signing certificate")
	}

	return idp, nil
}
