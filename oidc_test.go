package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Two logins that arrive while LinkedIn is slow must both reach it. Holding the
// lock across the discovery request queues the second behind the first.
func TestDiscoveryRunsConcurrently(t *testing.T) {
	release := make(chan struct{})
	releaseAll := sync.OnceFunc(func() { close(release) })
	arrived := make(chan struct{}, 2)

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		arrived <- struct{}{}
		<-release

		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                 server.URL,
			"authorization_endpoint": server.URL + "/authorize",
			"token_endpoint":         server.URL + "/token",
			"jwks_uri":               server.URL + "/jwks",
		})
	}))
	t.Cleanup(server.Close)
	t.Cleanup(releaseAll)

	client := newOIDCClient(config{linkedInIssuer: server.URL})
	done := make(chan error, 2)
	for range 2 {
		go func() {
			_, _, err := client.discover(t.Context())
			done <- err
		}()
	}

	for range 2 {
		select {
		case <-arrived:
		case <-time.After(2 * time.Second):
			t.Fatal("the second discovery request waited behind the first")
		}
	}

	releaseAll()
	for range 2 {
		require.NoError(t, <-done)
	}
}
