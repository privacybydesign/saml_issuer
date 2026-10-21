// Package credential maps what an identity provider released onto the
// attributes of the Yivi credential that is issued from it.
package credential

import "time"

const (
	surfnetCredential  = "surfnet-2"
	linkedinCredential = "linkedin"

	surfnetValidMonths  = 3
	linkedinValidMonths = 12
)

// Attribute is one attribute of a credential about to be issued.
type Attribute struct {
	Key   string
	Value string
}

// mapping takes the credential attribute key from the identity provider
// attribute called source.
type mapping struct {
	key    string
	source string
}

// Spec describes a credential and where its attributes come from.
type Spec struct {
	// ID is the full credential identifier, e.g. pbdf.pbdf.surfnet-2.
	ID          string
	validMonths int
	// BatchSize is the number of SD-JWT instances to issue.
	BatchSize int
	// mappings are in the order the attributes are shown to the user.
	mappings []mapping
}

// Surfnet is the credential issued from SURFconext.
func Surfnet(issuerID string, batchSize int) Spec {
	return Spec{
		ID:          issuerID + "." + surfnetCredential,
		validMonths: surfnetValidMonths,
		BatchSize:   batchSize,
		mappings: []mapping{
			{"institute", "urn:mace:terena.org:attribute-def:schacHomeOrganization"},
			{"type", "urn:mace:dir:attribute-def:eduPersonAffiliation"},
			{"id", "urn:mace:dir:attribute-def:uid"},
			{"fullid", "urn:mace:dir:attribute-def:eduPersonPrincipalName"},
			{"fullname", "urn:mace:dir:attribute-def:displayName"},
			{"firstname", "urn:mace:dir:attribute-def:givenName"},
			{"familyname", "urn:mace:dir:attribute-def:sn"},
			{"email", "urn:mace:dir:attribute-def:mail"},
		},
	}
}

// LinkedIn is the credential issued from LinkedIn. Sources are OpenID
// Connect claims.
func LinkedIn(issuerID string) Spec {
	return Spec{
		ID:          issuerID + "." + linkedinCredential,
		validMonths: linkedinValidMonths,
		mappings: []mapping{
			{"fullname", "name"},
			{"firstname", "given_name"},
			{"familyname", "family_name"},
			{"email", "email"},
		},
	}
}

// Map returns the credential attributes for what the identity provider
// released. Only the first value of an attribute is used, and attributes that
// were not released are left out.
func (s Spec) Map(released map[string][]string) []Attribute {
	var attributes []Attribute
	for _, m := range s.mappings {
		values := released[m.source]
		if len(values) == 0 {
			continue
		}

		attributes = append(attributes, Attribute{Key: m.key, Value: values[0]})
	}

	return attributes
}

// ValidUntil returns when a credential issued at now expires.
func (s Spec) ValidUntil(now time.Time) time.Time {
	return now.AddDate(0, s.validMonths, 0)
}

// AsMap returns the attributes by key.
func AsMap(attributes []Attribute) map[string]string {
	byKey := make(map[string]string, len(attributes))
	for _, attribute := range attributes {
		byKey[attribute.Key] = attribute.Value
	}

	return byKey
}
