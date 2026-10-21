package main

import (
	"log/slog"
	"net/http"

	"github.com/privacybydesign/saml_issuer/internal/credential"
	"github.com/privacybydesign/saml_issuer/internal/irma"
	"github.com/privacybydesign/saml_issuer/internal/session"
)

// page serves a provider's page: log in, the attributes to issue, or done.
func (s *server) page(w http.ResponseWriter, r *http.Request, pg page) {
	w.Header().Set(headerCacheControl, cacheNoStore)

	query := r.URL.Query()
	action, output := query.Get(queryAction), query.Get(queryOutput)
	id := sessionID(r)
	data, _ := s.sessions.Get(id)
	attributes, loggedIn := data.Login[pg.provider.id]

	switch {
	case action == actionLogin && !loggedIn:
		s.startLogin(w, r, pg)
	case action == actionLogout && loggedIn:
		s.logout(w, r, pg, id)
	case action == actionDone:
		s.render(w, http.StatusOK, "done", pg, s.pageData(pg, r.URL.RawQuery))
	case loggedIn && output == outputIssue:
		s.startIssuance(w, r, pg, attributes)
	case loggedIn:
		s.renderIssue(w, r, pg, attributes)
	case output == outputIssue:
		http.Error(w, "not logged in", http.StatusUnauthorized)
	default:
		s.render(w, http.StatusOK, "login", pg, s.pageData(pg, r.URL.RawQuery))
	}
}

func (s *server) renderIssue(w http.ResponseWriter, r *http.Request, pg page, attributes []credential.Attribute) {
	data := s.pageData(pg, r.URL.RawQuery)
	data.Attributes = attributeRows(pg.lang(), attributes)
	for _, attribute := range attributes {
		if attribute.Key == firstNameKey {
			data.FirstName = attribute.Value
		}
	}

	s.render(w, http.StatusOK, "issue", pg, data)
}

func (s *server) startLogin(w http.ResponseWriter, r *http.Request, pg page) {
	id, err := s.ensureSession(w, r)
	if err != nil {
		slog.Error("start session", "error", err)
		http.Error(w, "cannot start a session", http.StatusServiceUnavailable)
		return
	}

	switch pg.provider.id {
	case providerSurfnet:
		s.startSAML(w, r, pg, id)
	case providerLinkedIn:
		s.startLinkedIn(w, r, pg, id)
	}
}

// logout only ends the local login: there is no single logout.
func (s *server) logout(w http.ResponseWriter, r *http.Request, pg page, id string) {
	s.sessions.Update(id, func(d *session.Data) {
		delete(d.Login, pg.provider.id)
		delete(d.Pending, pg.provider.id)
	})

	http.Redirect(w, r, pg.path(), http.StatusSeeOther)
}

// startIssuance starts the issuance session at the IRMA server and hands its
// answer to the Yivi frontend.
func (s *server) startIssuance(w http.ResponseWriter, r *http.Request, pg page, attributes []credential.Attribute) {
	if len(attributes) == 0 {
		http.Error(w, "no attributes to issue", http.StatusConflict)
		return
	}

	spec := pg.provider.spec
	reply, err := s.irma.StartIssuance(r.Context(), irma.Issuance{
		Credential:     spec.ID,
		Attributes:     credential.AsMap(attributes),
		Validity:       spec.ValidUntil(s.clock.Now()),
		SDJWTBatchSize: spec.BatchSize,
	})
	if err != nil {
		slog.Error("start issuance", "provider", pg.provider.id, "error", err)
		http.Error(w, "cannot start the issuance session", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write(reply)
}
