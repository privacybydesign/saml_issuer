package saml

import "errors"

// Reasons a response is rejected. The ACS handler logs them and shows a
// generic page; tests match on them.
var (
	ErrMalformed    = errors.New("malformed response")
	ErrUnsigned     = errors.New("assertion is not signed")
	ErrVerification = errors.New("signature or schema verification failed")
	ErrStatus       = errors.New("status is not Success")
	ErrIssuer       = errors.New("unexpected issuer")
	ErrAudience     = errors.New("not addressed to this service provider")
	ErrDestination  = errors.New("unexpected destination or recipient")
	ErrInResponseTo = errors.New("response does not answer our request")
	ErrExpired      = errors.New("assertion expired")
	ErrNotYetValid  = errors.New("assertion not yet valid")
)
