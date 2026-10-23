package main

import (
	"github.com/privacybydesign/saml_issuer/internal/credential"
)

const (
	providerSurfnet  = "surfnet"
	providerLinkedIn = "linkedin"
)

// provider is one place to log in at and the credential that comes out of it.
type provider struct {
	id   string
	name string
	// path is where its pages live, without a language prefix.
	path string
	spec credential.Spec
}

func surfnetProvider(cfg config) provider {
	return provider{
		id:   providerSurfnet,
		name: "SURFconext",
		path: "/surfconext/",
		spec: credential.Surfnet(cfg.irmaIssuerID, cfg.surfnetBatch),
	}
}

func linkedInProvider(cfg config) provider {
	return provider{
		id:   providerLinkedIn,
		name: "LinkedIn",
		path: "/linkedin/",
		spec: credential.LinkedIn(cfg.irmaIssuerID),
	}
}
