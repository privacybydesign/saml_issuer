package saml

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	saml2 "github.com/russellhaering/gosaml2"
	"github.com/russellhaering/gosaml2/types"
)

// Validate checks a base64 SAMLResponse from the HTTP-POST binding, sent by
// the given identity provider in answer to the request with ID requestID, and
// returns the attributes of its assertion.
//
// The response is verified against the identity provider's pinned
// certificates only. Everything returned, and every check made here, comes
// from the element the library verified, never from the raw document.
func (sp *ServiceProvider) Validate(idpID, encoded, requestID string) (Attributes, error) {
	state, err := sp.state(idpID)
	if err != nil {
		return nil, err
	}
	if requestID == "" {
		return nil, fmt.Errorf("%w: no request to answer", ErrInResponseTo)
	}

	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMalformed, err)
	}
	if len(raw) > MaxResponseBytes {
		return nil, fmt.Errorf("%w: %d bytes", ErrMalformed, len(raw))
	}

	shape, err := checkShape(raw)
	if err != nil {
		return nil, err
	}

	response, err := state.verify.ValidateEncodedResponse(encoded)
	if err != nil {
		return nil, classify(err)
	}

	assertion, err := sp.checkResponse(state, response, shape, requestID)
	if err != nil {
		return nil, err
	}

	return attributesOf(assertion), nil
}

// classify maps the library's typed errors onto our reasons.
func classify(err error) error {
	var invalid saml2.ErrInvalidValue
	if errors.As(err, &invalid) {
		switch {
		case invalid.Reason == saml2.ReasonExpired:
			return fmt.Errorf("%w: %v", ErrExpired, err)
		case invalid.Key == saml2.DestinationAttr, invalid.Key == saml2.RecipientAttr:
			return fmt.Errorf("%w: %v", ErrDestination, err)
		case invalid.Key == saml2.IssuerTag:
			return fmt.Errorf("%w: %v", ErrIssuer, err)
		case invalid.Key == saml2.StatusCodeTag:
			return fmt.Errorf("%w: %v", ErrStatus, err)
		}
	}

	return fmt.Errorf("%w: %v", ErrVerification, err)
}

// checkResponse repeats the checks the library makes on the parts of the
// Response it does not sign-verify, and adds the ones it leaves out.
func (sp *ServiceProvider) checkResponse(state *idpState, response *types.Response, shape *shape, requestID string) (*types.Assertion, error) {
	if len(response.Assertions) != 1 || len(response.EncryptedAssertions) != 0 {
		return nil, fmt.Errorf("%w: verified response holds %d assertions", ErrMalformed, len(response.Assertions))
	}

	assertion := &response.Assertions[0]
	if assertion.ID != shape.assertionID {
		return nil, fmt.Errorf("%w: verified assertion is not the one that was checked", ErrMalformed)
	}

	if assertion.Issuer == nil || assertion.Issuer.Value != state.idp.EntityID {
		return nil, ErrIssuer
	}

	if response.Destination != sp.cfg.ACSURL {
		return nil, fmt.Errorf("%w: destination %q", ErrDestination, response.Destination)
	}

	if response.InResponseTo != requestID {
		return nil, ErrInResponseTo
	}

	if err := checkSubject(assertion, sp.cfg.ACSURL, requestID); err != nil {
		return nil, err
	}

	if err := sp.checkConditions(assertion); err != nil {
		return nil, err
	}

	return assertion, nil
}

func checkSubject(assertion *types.Assertion, acsURL, requestID string) error {
	if assertion.Subject == nil || assertion.Subject.SubjectConfirmation == nil ||
		assertion.Subject.SubjectConfirmation.SubjectConfirmationData == nil {
		return fmt.Errorf("%w: no subject confirmation", ErrMalformed)
	}

	data := assertion.Subject.SubjectConfirmation.SubjectConfirmationData
	if data.Recipient != acsURL {
		return fmt.Errorf("%w: recipient %q", ErrDestination, data.Recipient)
	}
	if data.InResponseTo != requestID {
		return ErrInResponseTo
	}

	return nil
}

func (sp *ServiceProvider) checkConditions(assertion *types.Assertion) error {
	conditions := assertion.Conditions
	if conditions == nil {
		return fmt.Errorf("%w: no conditions", ErrMalformed)
	}

	notBefore, err := time.Parse(time.RFC3339, conditions.NotBefore)
	if err != nil {
		return fmt.Errorf("%w: NotBefore: %v", ErrMalformed, err)
	}

	notOnOrAfter, err := time.Parse(time.RFC3339, conditions.NotOnOrAfter)
	if err != nil {
		return fmt.Errorf("%w: NotOnOrAfter: %v", ErrMalformed, err)
	}

	now := sp.cfg.Clock.Now()
	if now.Add(ClockSkew).Before(notBefore) {
		return ErrNotYetValid
	}
	if !now.Add(-ClockSkew).Before(notOnOrAfter) {
		return ErrExpired
	}

	if len(conditions.AudienceRestrictions) == 0 {
		return fmt.Errorf("%w: no audience restriction", ErrAudience)
	}
	for _, restriction := range conditions.AudienceRestrictions {
		if !hasAudience(restriction, sp.cfg.EntityID) {
			return ErrAudience
		}
	}

	return nil
}

func hasAudience(restriction types.AudienceRestriction, audience string) bool {
	for _, candidate := range restriction.Audiences {
		if candidate.Value == audience {
			return true
		}
	}

	return false
}

func attributesOf(assertion *types.Assertion) Attributes {
	attributes := Attributes{}
	if assertion.AttributeStatement == nil {
		return attributes
	}

	for _, attribute := range assertion.AttributeStatement.Attributes {
		for _, value := range attribute.Values {
			attributes[attribute.Name] = append(attributes[attribute.Name], valueString(value))
		}
	}

	return attributes
}

// valueString returns the text of an attribute value, or the value of the
// NameID inside it (eduPersonTargetedID).
func valueString(value types.AttributeValue) string {
	if value.NameID != nil && value.NameID.Value != "" {
		return value.NameID.Value
	}

	return value.Value
}
