package saml

import (
	"testing"

	"github.com/russellhaering/gosaml2/types"
	"github.com/stretchr/testify/require"
)

// No document makes the library verify another assertion than the one checkShape
// saw, so the guard is tested on its own.
func TestVerifiedAssertionIsTheCheckedOne(t *testing.T) {
	sp := &ServiceProvider{}
	response := &types.Response{Assertions: []types.Assertion{{ID: "_verified"}}}

	_, err := sp.checkResponse(nil, response, &shape{assertionID: "_checked"}, "_request-1")

	require.ErrorIs(t, err, ErrMalformed)
}
