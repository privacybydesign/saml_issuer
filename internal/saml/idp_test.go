package saml_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/privacybydesign/saml_issuer/internal/saml"
	"github.com/privacybydesign/saml_issuer/internal/samltest"
)

func writeMetadata(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "idp.xml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	return path
}

func TestLoadSurfconextMetadata(t *testing.T) {
	// Metadata as SURFconext publishes it at https://metadata.surfconext.nl/.
	idps, err := saml.LoadIdPs(filepath.Join("testdata", "surfconext-idp-metadata.xml"))

	require.NoError(t, err)
	require.Len(t, idps, 1)
	assert.Equal(t, "https://engine.surfconext.nl/authentication/idp/metadata", idps[0].EntityID)
	assert.Equal(t, "https://engine.surfconext.nl/authentication/idp/single-sign-on/key:20230503", idps[0].SSOURL)
	require.Len(t, idps[0].Certs, 1)
	assert.Contains(t, idps[0].Certs[0].Subject.CommonName, "engine.surfconext.nl")
}

func TestLoadSingleEntity(t *testing.T) {
	idp := samltest.NewIdP(t)

	idps, err := saml.LoadIdPs(writeMetadata(t, string(idp.MetadataXML())))

	require.NoError(t, err)
	require.Len(t, idps, 1)
	assert.Equal(t, idp.EntityID, idps[0].EntityID)
	assert.Equal(t, idp.SSOURL, idps[0].SSOURL)
	assert.True(t, idp.Cert.Equal(idps[0].Certs[0]))
}

func TestLoadSeveralEntities(t *testing.T) {
	first, second := samltest.NewIdP(t), samltest.NewAttacker(t)
	second.EntityID = "https://second.example.org/metadata"
	// A service provider in the file is not an identity provider.
	sp := `<md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata" entityID="https://sp.example.org"><md:SPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol"/></md:EntityDescriptor>`
	content := `<md:EntitiesDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata">` +
		stripHeader(first.MetadataXML()) + sp + stripHeader(second.MetadataXML()) + `</md:EntitiesDescriptor>`

	idps, err := saml.LoadIdPs(writeMetadata(t, content))

	require.NoError(t, err)
	require.Len(t, idps, 2)
	assert.Equal(t, first.EntityID, idps[0].EntityID)
	assert.Equal(t, second.EntityID, idps[1].EntityID)
}

func TestLoadIgnoresEncryptionKeys(t *testing.T) {
	idp := samltest.NewIdP(t)
	metadata := strings.Replace(string(idp.MetadataXML()), `use="signing"`, `use="encryption"`, 1)

	_, err := saml.LoadIdPs(writeMetadata(t, metadata))

	require.ErrorContains(t, err, "no signing certificate")
}

func TestLoadRejectsBadMetadata(t *testing.T) {
	idp := string(samltest.NewIdP(t).MetadataXML())
	tests := map[string]struct {
		content string
		want    string
	}{
		"not XML":        {"hello", "parse metadata"},
		"wrong root":     {`<md:Other xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata"/>`, "metadata root"},
		"no IdP":         {`<md:EntitiesDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata"/>`, "no identity provider"},
		"no redirect":    {strings.Replace(idp, "HTTP-Redirect", "HTTP-POST", 1), "HTTP-Redirect"},
		"no certificate": {removeBetween(idp, "<ds:X509Certificate>", "</ds:X509Certificate>"), "no signing certificate"},
		"bad base64":     {strings.Replace(idp, "<ds:X509Certificate>", "<ds:X509Certificate>!!", 1), "decode certificate"},
		"not a cert":     {strings.Replace(idp, "<ds:X509Certificate>MII", "<ds:X509Certificate>AAA", 1), "parse certificate"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := saml.LoadIdPs(writeMetadata(t, tt.content))

			require.ErrorContains(t, err, tt.want)
		})
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := saml.LoadIdPs(filepath.Join(t.TempDir(), "nope.xml"))

	require.ErrorIs(t, err, os.ErrNotExist)
}

func stripHeader(xml []byte) string {
	return strings.TrimPrefix(string(xml), `<?xml version="1.0"?>`)
}

// removeBetween returns s without the text from start up to and including end.
func removeBetween(s, start, end string) string {
	from := strings.Index(s, start)
	to := strings.Index(s, end) + len(end)

	return s[:from] + s[to:]
}
