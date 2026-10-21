package credential_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/privacybydesign/saml_issuer/internal/credential"
)

func TestSurfnetMapsAllAttributes(t *testing.T) {
	released := map[string][]string{
		"urn:mace:terena.org:attribute-def:schacHomeOrganization": {"example.org"},
		"urn:mace:dir:attribute-def:eduPersonAffiliation":         {"student", "member"},
		"urn:mace:dir:attribute-def:uid":                          {"jdoe"},
		"urn:mace:dir:attribute-def:eduPersonPrincipalName":       {"jdoe@example.org"},
		"urn:mace:dir:attribute-def:displayName":                  {"Jane Doe"},
		"urn:mace:dir:attribute-def:givenName":                    {"Jane"},
		"urn:mace:dir:attribute-def:sn":                           {"Doe"},
		"urn:mace:dir:attribute-def:mail":                         {"jane@example.org"},
	}

	attributes := credential.Surfnet("pbdf.pbdf", 200).Map(released)

	assert.Equal(t, []credential.Attribute{
		{Key: "institute", Value: "example.org"},
		{Key: "type", Value: "student"},
		{Key: "id", Value: "jdoe"},
		{Key: "fullid", Value: "jdoe@example.org"},
		{Key: "fullname", Value: "Jane Doe"},
		{Key: "firstname", Value: "Jane"},
		{Key: "familyname", Value: "Doe"},
		{Key: "email", Value: "jane@example.org"},
	}, attributes)
}

func TestSurfnetSkipsMissing(t *testing.T) {
	released := map[string][]string{
		"urn:mace:dir:attribute-def:uid":         {"jdoe"},
		"urn:mace:dir:attribute-def:displayName": {},
		"urn:mace:dir:attribute-def:mail":        {""},
		"urn:oid:0.9.2342.19200300.100.1.3":      {"not@used.example"},
	}

	attributes := credential.Surfnet("pbdf.pbdf", 200).Map(released)

	assert.Equal(t, []credential.Attribute{
		{Key: "id", Value: "jdoe"},
		{Key: "email", Value: ""},
	}, attributes)
}

func TestLinkedInMapsClaims(t *testing.T) {
	released := map[string][]string{
		"name":        {"Jane Doe"},
		"given_name":  {"Jane"},
		"family_name": {"Doe"},
		"email":       {"jane@example.org"},
		"picture":     {"https://example.org/p.jpg"},
	}

	attributes := credential.LinkedIn("pbdf.pbdf").Map(released)

	assert.Equal(t, []credential.Attribute{
		{Key: "fullname", Value: "Jane Doe"},
		{Key: "firstname", Value: "Jane"},
		{Key: "familyname", Value: "Doe"},
		{Key: "email", Value: "jane@example.org"},
	}, attributes)
}

func TestMapOfNothingIsEmpty(t *testing.T) {
	assert.Empty(t, credential.Surfnet("pbdf.pbdf", 200).Map(nil))
	assert.Empty(t, credential.LinkedIn("pbdf.pbdf").Map(nil))
}

func TestCredentialIdentifiers(t *testing.T) {
	assert.Equal(t, "pbdf.pbdf.surfnet-2", credential.Surfnet("pbdf.pbdf", 200).ID)
	assert.Equal(t, "irma-demo.sidn-pbdf.surfnet-2", credential.Surfnet("irma-demo.sidn-pbdf", 200).ID)
	assert.Equal(t, "pbdf.pbdf.linkedin", credential.LinkedIn("pbdf.pbdf").ID)
}

func TestValidity(t *testing.T) {
	now := time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)

	assert.Equal(t, time.Date(2027, 1, 21, 12, 0, 0, 0, time.UTC), credential.Surfnet("pbdf.pbdf", 200).ValidUntil(now))
	assert.Equal(t, time.Date(2027, 10, 21, 12, 0, 0, 0, time.UTC), credential.LinkedIn("pbdf.pbdf").ValidUntil(now))
}

func TestBatchSize(t *testing.T) {
	assert.Equal(t, 200, credential.Surfnet("pbdf.pbdf", 200).BatchSize)
	assert.Zero(t, credential.LinkedIn("pbdf.pbdf").BatchSize)
}

func TestAsMap(t *testing.T) {
	attributes := []credential.Attribute{{Key: "id", Value: "jdoe"}, {Key: "email", Value: "j@example.org"}}

	assert.Equal(t, map[string]string{"id": "jdoe", "email": "j@example.org"}, credential.AsMap(attributes))
}
