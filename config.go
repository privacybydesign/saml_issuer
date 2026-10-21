package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	listenAddr = ":8080"

	defaultIssuerID  = "pbdf.pbdf"
	defaultBatchSize = 200
	defaultCertDir   = "/var/simplesamlphp/cert"
	defaultMetaDir   = "/var/simplesamlphp/metadata"

	spCertFile      = "saml.crt"
	idpMetadataFile = "saml20-idp-remote.xml"

	linkedInIssuer = "https://www.linkedin.com/oauth"
)

type config struct {
	baseURL string

	irmaServerURL string
	irmaToken     string
	irmaIssuerID  string
	surfnetBatch  int

	spCert      *x509.Certificate
	idpMetadata string
	idpEntityID string

	linkedInClientID     string
	linkedInClientSecret string
	linkedInIssuer       string
}

// loadConfig reads the environment variables the PHP service used, so the
// deployment keeps working with a new image tag.
func loadConfig(getenv func(string) string) (config, error) {
	cfg := config{
		irmaServerURL:        strings.TrimRight(getenv("IRMA_SERVER_URL"), "/"),
		irmaToken:            getenv("API_TOKEN"),
		irmaIssuerID:         getenv("IRMA_ISSUER_ID"),
		idpEntityID:          getenv("SURFCONEXT_IDP_ENTITY_ID"),
		linkedInClientID:     getenv("LINKEDIN_CLIENT_ID"),
		linkedInClientSecret: getenv("LINKEDIN_CLIENT_SECRET"),
		linkedInIssuer:       linkedInIssuer,
		surfnetBatch:         defaultBatchSize,
	}

	host := getenv("HOST_NAME")
	if host == "" {
		return config{}, fmt.Errorf("HOST_NAME is not set")
	}

	protocol := getenv("HTTP_PROTOCOL")
	switch protocol {
	case "":
		protocol = "https"
	case "http", "https":
	default:
		return config{}, fmt.Errorf("HTTP_PROTOCOL must be http or https, got %q", protocol)
	}
	cfg.baseURL = protocol + "://" + host

	if cfg.irmaServerURL == "" {
		return config{}, fmt.Errorf("IRMA_SERVER_URL is not set")
	}

	if cfg.irmaIssuerID == "" {
		slog.Warn("IRMA_ISSUER_ID not set, using default", "default", defaultIssuerID)
		cfg.irmaIssuerID = defaultIssuerID
	}

	if raw := getenv("IRMA_SDJWT_BATCH_SIZE_SURFCONTEXT"); raw != "" {
		size, err := strconv.Atoi(raw)
		if err != nil || size < 0 {
			return config{}, fmt.Errorf("IRMA_SDJWT_BATCH_SIZE_SURFCONTEXT must be a number of at least 0, got %q", raw)
		}
		cfg.surfnetBatch = size
	}

	if (cfg.linkedInClientID == "") != (cfg.linkedInClientSecret == "") {
		return config{}, fmt.Errorf("LINKEDIN_CLIENT_ID and LINKEDIN_CLIENT_SECRET must be set together")
	}

	certDir := envOr(getenv, "SIMPLESAMLPHP_CERT_DIR", defaultCertDir)
	cert, err := readCertificate(filepath.Join(certDir, spCertFile))
	if err != nil {
		return config{}, err
	}
	cfg.spCert = cert

	cfg.idpMetadata = filepath.Join(envOr(getenv, "SIMPLESAMLPHP_METADATA_DIR", defaultMetaDir), idpMetadataFile)

	return cfg, nil
}

func envOr(getenv func(string) string, name, fallback string) string {
	if value := getenv(name); value != "" {
		return value
	}

	return fallback
}

func readCertificate(path string) (*x509.Certificate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("SP certificate: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("SP certificate %s: no PEM certificate found", path)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("SP certificate %s: %w", path, err)
	}

	return cert, nil
}
