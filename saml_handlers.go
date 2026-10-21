package main

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/privacybydesign/saml_issuer/internal/saml"
	"github.com/privacybydesign/saml_issuer/internal/session"
)

const (
	// maxBodyBytes bounds the ACS POST. A response is at most
	// saml.MaxResponseBytes of XML, which base64 and form encoding inflate.
	maxBodyBytes = 4 * saml.MaxResponseBytes

	samlResponseField = "SAMLResponse"
)

// spMetadata publishes the entity ID, certificate and ACS for SURFconext.
func (s *server) spMetadata(w http.ResponseWriter, r *http.Request) {
	body, err := s.saml.Metadata()
	if err != nil {
		slog.Error("build SP metadata", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/samlmetadata+xml")
	_, _ = w.Write(body)
}

// startSAML sends the user to the identity provider with an AuthnRequest whose
// ID is kept in the session. Only a response that answers it is accepted.
func (s *server) startSAML(w http.ResponseWriter, r *http.Request, pg page, id string) {
	request, err := s.saml.Login(s.idpID)
	if err != nil {
		slog.Error("build AuthnRequest", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	s.sessions.Update(id, func(d *session.Data) {
		d.Pending[pg.provider.id] = session.Pending{Token: request.ID, ReturnTo: pg.path()}
	})

	http.Redirect(w, r, request.URL, http.StatusFound)
}

// acs receives the SAML response, posted by the user's browser.
func (s *server) acs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(headerCacheControl, cacheNoStore)
	p := s.surfnet

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := r.ParseForm(); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			s.renderFailure(w, http.StatusRequestEntityTooLarge, page{provider: p})
			return
		}

		s.renderFailure(w, http.StatusBadRequest, page{provider: p})
		return
	}

	id := sessionID(r)
	data, _ := s.sessions.Get(id)
	pending, ok := data.Pending[p.id]
	if !ok {
		slog.Warn("SAML response without a login in progress")
		s.renderFailure(w, http.StatusBadRequest, page{provider: p})
		return
	}
	pg := s.pageFor(pending, p)

	released, err := s.saml.Validate(s.idpID, r.PostForm.Get(samlResponseField), pending.Token)
	if err != nil {
		slog.Warn("SAML response rejected", "error", err)
		s.renderFailure(w, http.StatusForbidden, pg)
		return
	}

	// Accepted once: a replay finds the pending login gone.
	if err := s.consumePending(id, p, pending.Token); err != nil {
		s.renderFailure(w, http.StatusForbidden, pg)
		return
	}

	s.finishLogin(w, r, id, p, pending, p.spec.Map(released))
}
