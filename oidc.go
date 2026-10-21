package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/privacybydesign/saml_issuer/internal/session"
)

const (
	oidcTimeout = 15 * time.Second
	stateBytes  = 32
)

// oidcClient logs users in with an OpenID Connect provider (LinkedIn).
type oidcClient struct {
	oauth  oauth2.Config
	issuer string
	http   *http.Client

	mu       sync.Mutex
	provider *oidc.Provider
}

func newOIDCClient(cfg config) *oidcClient {
	return &oidcClient{
		issuer: cfg.linkedInIssuer,
		http:   &http.Client{Timeout: oidcTimeout},
		oauth: oauth2.Config{
			ClientID:     cfg.linkedInClientID,
			ClientSecret: cfg.linkedInClientSecret,
			RedirectURL:  cfg.baseURL + linkbackPath,
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		},
	}
}

// discover fetches the provider's configuration on first use, so that the
// service starts while LinkedIn is unreachable.
func (c *oidcClient) discover(ctx context.Context) (*oidc.Provider, oauth2.Config, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.provider == nil {
		provider, err := oidc.NewProvider(oidc.ClientContext(ctx, c.http), c.issuer)
		if err != nil {
			return nil, oauth2.Config{}, err
		}
		c.provider = provider
	}

	conf := c.oauth
	conf.Endpoint = c.provider.Endpoint()
	// LinkedIn expects the client credentials in the body of the token request.
	conf.Endpoint.AuthStyle = oauth2.AuthStyleInParams

	return c.provider, conf, nil
}

func (c *oidcClient) authURL(ctx context.Context, state string) (string, error) {
	_, conf, err := c.discover(ctx)
	if err != nil {
		return "", err
	}

	return conf.AuthCodeURL(state), nil
}

// claims exchanges the authorization code and returns the user's claims, as
// released by the provider's userinfo endpoint. The ID token is verified and
// must belong to the same subject.
func (c *oidcClient) claims(ctx context.Context, code string) (map[string][]string, error) {
	provider, conf, err := c.discover(ctx)
	if err != nil {
		return nil, err
	}
	ctx = oidc.ClientContext(ctx, c.http)

	token, err := conf.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no ID token in the token response")
	}

	idToken, err := provider.Verifier(&oidc.Config{ClientID: conf.ClientID, SupportedSigningAlgs: []string{oidc.RS256}}).Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("verify ID token: %w", err)
	}

	info, err := provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
	if err != nil {
		return nil, fmt.Errorf("userinfo: %w", err)
	}
	if info.Subject != idToken.Subject {
		return nil, errors.New("userinfo is about another subject than the ID token")
	}

	var all map[string]any
	if err := info.Claims(&all); err != nil {
		return nil, fmt.Errorf("userinfo claims: %w", err)
	}

	released := map[string][]string{}
	for name, value := range all {
		if text, ok := value.(string); ok {
			released[name] = []string{text}
		}
	}

	return released, nil
}

func newState() (string, error) {
	raw := make([]byte, stateBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// startLinkedIn sends the user to LinkedIn with a state kept in the session.
func (s *server) startLinkedIn(w http.ResponseWriter, r *http.Request, pg page, id string) {
	state, err := newState()
	if err != nil {
		slog.Error("create OAuth state", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	target, err := s.oidc.authURL(r.Context(), state)
	if err != nil {
		slog.Error("LinkedIn discovery", "error", err)
		http.Error(w, "LinkedIn is not available", http.StatusBadGateway)
		return
	}

	s.sessions.Update(id, func(d *session.Data) {
		d.Pending[pg.provider.id] = session.Pending{Token: state, ReturnTo: pg.path()}
	})

	http.Redirect(w, r, target, http.StatusFound)
}

// linkedInCallback is the redirect URI LinkedIn sends the user back to.
func (s *server) linkedInCallback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(headerCacheControl, cacheNoStore)
	p := s.pages[linkedInProvider(s.cfg).path].provider

	query := r.URL.Query()
	id := sessionID(r)
	data, _ := s.sessions.Get(id)
	pending, ok := data.Pending[p.id]
	pg := s.pageFor(pending, p)

	state := query.Get("state")
	if !ok || state == "" || subtle.ConstantTimeCompare([]byte(pending.Token), []byte(state)) != 1 {
		slog.Warn("LinkedIn answer without a matching state")
		s.renderFailure(w, http.StatusBadRequest, pg)
		return
	}

	// The state is single use, whatever LinkedIn answered.
	if err := s.consumePending(id, p, state); err != nil {
		s.renderFailure(w, http.StatusBadRequest, pg)
		return
	}

	if reason := query.Get("error"); reason != "" {
		slog.Warn("LinkedIn login refused", "error", reason)
		s.renderFailure(w, http.StatusForbidden, pg)
		return
	}

	released, err := s.oidc.claims(r.Context(), query.Get("code"))
	if err != nil {
		slog.Warn("LinkedIn login failed", "error", err)
		s.renderFailure(w, http.StatusForbidden, pg)
		return
	}

	s.finishLogin(w, r, id, p, pending, p.spec.Map(released))
}
