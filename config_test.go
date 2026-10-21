package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/privacybydesign/saml_issuer/internal/samltest"
)

// validEnv is a complete environment with the certificate and IdP metadata on disk.
func validEnv(t *testing.T) map[string]string {
	t.Helper()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, spCertFile), pemCertificate(samltest.NewCertificate(t)))
	writeFile(t, filepath.Join(dir, idpMetadataFile), string(samltest.NewIdP(t).MetadataXML()))

	return map[string]string{
		"HOST_NAME":                  "saml-issuer.example.org",
		"IRMA_SERVER_URL":            "https://irma.example.org/",
		"API_TOKEN":                  "token",
		"SIMPLESAMLPHP_CERT_DIR":     dir,
		"SIMPLESAMLPHP_METADATA_DIR": dir,
	}
}

func load(env map[string]string) (config, error) {
	return loadConfig(func(name string) string { return env[name] })
}

func TestConfigDefaults(t *testing.T) {
	cfg, err := load(validEnv(t))

	require.NoError(t, err)
	assert.Equal(t, "https://saml-issuer.example.org", cfg.baseURL)
	assert.Equal(t, "https://irma.example.org", cfg.irmaServerURL)
	assert.Equal(t, "token", cfg.irmaToken)
	assert.Equal(t, "pbdf.pbdf", cfg.irmaIssuerID)
	assert.Equal(t, 200, cfg.surfnetBatch)
	assert.Empty(t, cfg.linkedInClientID)
	assert.Equal(t, "https://www.linkedin.com/oauth", cfg.linkedInIssuer)
	assert.NotNil(t, cfg.spCert)
}

func TestConfigVariables(t *testing.T) {
	env := validEnv(t)
	env["HTTP_PROTOCOL"] = "http"
	env["IRMA_ISSUER_ID"] = "irma-demo.sidn-pbdf"
	env["IRMA_SDJWT_BATCH_SIZE_SURFCONTEXT"] = "50"
	env["LINKEDIN_CLIENT_ID"] = "id"
	env["LINKEDIN_CLIENT_SECRET"] = "secret"
	env["SURFCONEXT_IDP_ENTITY_ID"] = "https://idp.example.org"

	cfg, err := load(env)

	require.NoError(t, err)
	assert.Equal(t, "http://saml-issuer.example.org", cfg.baseURL)
	assert.Equal(t, "irma-demo.sidn-pbdf", cfg.irmaIssuerID)
	assert.Equal(t, 50, cfg.surfnetBatch)
	assert.Equal(t, "id", cfg.linkedInClientID)
	assert.Equal(t, "secret", cfg.linkedInClientSecret)
	assert.Equal(t, "https://idp.example.org", cfg.idpEntityID)
	assert.Equal(t, filepath.Join(env["SIMPLESAMLPHP_METADATA_DIR"], "saml20-idp-remote.xml"), cfg.idpMetadata)
}

func TestConfigRejectsBadValues(t *testing.T) {
	tests := map[string]func(env map[string]string){
		"no host":                 func(env map[string]string) { delete(env, "HOST_NAME") },
		"no IRMA server":          func(env map[string]string) { delete(env, "IRMA_SERVER_URL") },
		"bad protocol":            func(env map[string]string) { env["HTTP_PROTOCOL"] = "ftp" },
		"bad batch size":          func(env map[string]string) { env["IRMA_SDJWT_BATCH_SIZE_SURFCONTEXT"] = "many" },
		"negative batch size":     func(env map[string]string) { env["IRMA_SDJWT_BATCH_SIZE_SURFCONTEXT"] = "-1" },
		"LinkedIn without secret": func(env map[string]string) { env["LINKEDIN_CLIENT_ID"] = "id" },
		"LinkedIn without id":     func(env map[string]string) { env["LINKEDIN_CLIENT_SECRET"] = "secret" },
		"no certificate":          func(env map[string]string) { env["SIMPLESAMLPHP_CERT_DIR"] = t.TempDir() },
	}

	for name, change := range tests {
		t.Run(name, func(t *testing.T) {
			env := validEnv(t)
			change(env)

			_, err := load(env)

			require.Error(t, err)
		})
	}
}

func TestConfigBadCertificate(t *testing.T) {
	env := validEnv(t)
	writeFile(t, filepath.Join(env["SIMPLESAMLPHP_CERT_DIR"], spCertFile), "not a certificate")

	_, err := load(env)

	require.ErrorContains(t, err, "no PEM certificate")
}

func TestServerNeedsIdPMetadata(t *testing.T) {
	env := validEnv(t)
	writeFile(t, filepath.Join(env["SIMPLESAMLPHP_METADATA_DIR"], idpMetadataFile), "<nope/>")
	cfg, err := load(env)
	require.NoError(t, err)

	_, err = newServer(cfg, samltest.Clock())

	require.Error(t, err)
}

func TestSeveralIdPs(t *testing.T) {
	env := validEnv(t)
	first, second := samltest.NewIdP(t), samltest.NewAttacker(t)
	second.EntityID = "https://second.example.org/metadata"
	writeFile(t, filepath.Join(env["SIMPLESAMLPHP_METADATA_DIR"], idpMetadataFile),
		`<md:EntitiesDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata">`+
			stripXMLHeader(first.MetadataXML())+stripXMLHeader(second.MetadataXML())+`</md:EntitiesDescriptor>`)

	cfg, err := load(env)
	require.NoError(t, err)
	_, err = newServer(cfg, samltest.Clock())
	require.ErrorContains(t, err, "SURFCONEXT_IDP_ENTITY_ID")

	env["SURFCONEXT_IDP_ENTITY_ID"] = second.EntityID
	cfg, err = load(env)
	require.NoError(t, err)
	srv, err := newServer(cfg, samltest.Clock())
	require.NoError(t, err)
	assert.Equal(t, second.EntityID, srv.idpID)

	env["SURFCONEXT_IDP_ENTITY_ID"] = "https://unknown.example.org"
	cfg, err = load(env)
	require.NoError(t, err)
	_, err = newServer(cfg, samltest.Clock())
	require.Error(t, err)
}

func stripXMLHeader(xml []byte) string {
	const header = `<?xml version="1.0"?>`

	return string(xml[len(header):])
}
