package irma_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/privacybydesign/saml_issuer/internal/irma"
)

type recorded struct {
	method        string
	path          string
	contentType   string
	authorization string
	body          map[string]any
}

func fakeServer(t *testing.T, status int, reply string) (*httptest.Server, *recorded) {
	t.Helper()

	seen := &recorded{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &seen.body))

		seen.method = r.Method
		seen.path = r.URL.Path
		seen.contentType = r.Header.Get("Content-Type")
		seen.authorization = r.Header.Get("Authorization")

		w.WriteHeader(status)
		_, _ = io.WriteString(w, reply)
	}))
	t.Cleanup(server.Close)

	return server, seen
}

func issuance() irma.Issuance {
	return irma.Issuance{
		Credential:     "pbdf.pbdf.surfnet-2",
		Attributes:     map[string]string{"id": "jdoe", "email": "jane@example.org"},
		Validity:       time.Date(2027, 1, 21, 12, 0, 0, 0, time.UTC),
		SDJWTBatchSize: 200,
	}
}

func TestIssuanceRequest(t *testing.T) {
	server, seen := fakeServer(t, http.StatusOK, `{"token":"abc","sessionPtr":{"u":"https://irma.example/irma/session/abc","irmaqr":"issuing"}}`)

	reply, err := irma.NewClient(server.URL, "secret-token").StartIssuance(context.Background(), issuance())

	require.NoError(t, err)
	assert.JSONEq(t, `{"token":"abc","sessionPtr":{"u":"https://irma.example/irma/session/abc","irmaqr":"issuing"}}`, string(reply))
	assert.Equal(t, http.MethodPost, seen.method)
	assert.Equal(t, "/session", seen.path)
	assert.Equal(t, "application/json", seen.contentType)
	assert.Equal(t, "secret-token", seen.authorization)
	assert.Equal(t, map[string]any{
		"@context": "https://irma.app/ld/request/issuance/v2",
		"credentials": []any{map[string]any{
			"credential":     "pbdf.pbdf.surfnet-2",
			"validity":       float64(time.Date(2027, 1, 21, 12, 0, 0, 0, time.UTC).Unix()),
			"attributes":     map[string]any{"id": "jdoe", "email": "jane@example.org"},
			"sdJwtBatchSize": float64(200),
		}},
	}, seen.body)
}

func TestStartIssuanceWithoutToken(t *testing.T) {
	server, seen := fakeServer(t, http.StatusOK, `{}`)

	_, err := irma.NewClient(server.URL+"/", "").StartIssuance(context.Background(), issuance())

	require.NoError(t, err)
	assert.Equal(t, "/session", seen.path)
	assert.Empty(t, seen.authorization)
}

func TestIssuanceRejected(t *testing.T) {
	for _, status := range []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusInternalServerError} {
		server, _ := fakeServer(t, status, `{"error":"nope"}`)

		reply, err := irma.NewClient(server.URL, "token").StartIssuance(context.Background(), issuance())

		require.Error(t, err)
		assert.Nil(t, reply)
		assert.NotContains(t, err.Error(), "nope", "the server's reply must not leak into the error")
	}
}

func TestIssuanceUnreachable(t *testing.T) {
	server, _ := fakeServer(t, http.StatusOK, `{}`)
	server.Close()

	_, err := irma.NewClient(server.URL, "token").StartIssuance(context.Background(), issuance())

	require.Error(t, err)
}

func TestIssuanceContext(t *testing.T) {
	server, _ := fakeServer(t, http.StatusOK, `{}`)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := irma.NewClient(server.URL, "token").StartIssuance(ctx, issuance())

	require.ErrorIs(t, err, context.Canceled)
}
