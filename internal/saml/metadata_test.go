package saml_test

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/privacybydesign/saml_issuer/internal/saml"
	"github.com/privacybydesign/saml_issuer/internal/samltest"
)

const certPlaceholder = "CERTIFICATE"

// testdata/sp-metadata.xml is the document SURFconext reads and republishes in
// eduGAIN. It validates against the SAML metadata and MDUI schemas. eduGAIN
// requires the Organization and the technical and support ContactPerson:
// https://servicedesk.surf.nl/wiki/spaces/IAM/pages/128910147
func TestMetadata(t *testing.T) {
	cert := samltest.NewCertificate(t)
	sp, err := saml.New(saml.Config{
		EntityID: spEntityID,
		ACSURL:   acsURL,
		InfoURLs: []saml.Localized{
			{Lang: "en", Value: "https://issuer.example.org/en/surfconext/"},
			{Lang: "nl", Value: "https://issuer.example.org/nl/surfconext/"},
		},
		Cert:  cert,
		Clock: samltest.Clock(),
	})
	require.NoError(t, err)

	metadata, err := sp.Metadata()
	require.NoError(t, err)

	want, err := os.ReadFile(filepath.Join("testdata", "sp-metadata.xml"))
	require.NoError(t, err)

	encoded := base64.StdEncoding.EncodeToString(cert.Raw)
	require.Contains(t, string(metadata), encoded)
	assert.Equal(t, strings.TrimSpace(string(want)), strings.Replace(string(metadata), encoded, certPlaceholder, 1))
}
