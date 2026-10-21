package main

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jonboulle/clockwork"

	"github.com/privacybydesign/saml_issuer/internal/credential"
	"github.com/privacybydesign/saml_issuer/internal/irma"
	"github.com/privacybydesign/saml_issuer/internal/saml"
	"github.com/privacybydesign/saml_issuer/internal/session"
)

// Only these files are ever served; nothing else in the repository is
// reachable. enroll.js, jquery.min.js and yivi.js all live in js/.
//
//go:embed css/style.css js fonts images healthy.html
var assetFS embed.FS

const (
	spMetadataPath = "/simplesaml/module.php/saml/sp/metadata/surfnet"
	acsPath        = "/simplesaml/module.php/saml/sp/saml2-acs.php/surfnet"
	linkbackPath   = "/simplesaml/module.php/authoauth2/linkback"

	sessionCookie = "__Host-session"

	queryAction  = "action"
	queryOutput  = "output"
	actionLogin  = "login"
	actionLogout = "logout"
	actionDone   = "done"
	outputIssue  = "irma-session"

	headerCacheControl = "Cache-Control"
	cacheNoStore       = "no-store"

	// firstNameKey is the credential attribute the pages greet the user with.
	firstNameKey = "firstname"

	healthyPage = "healthy.html"
)

type server struct {
	cfg       config
	clock     clockwork.Clock
	sessions  *session.Store
	saml      *saml.ServiceProvider
	idpID     string
	irma      *irma.Client
	oidc      *oidcClient
	templates *template.Template
	surfnet   provider
	providers []provider
	// pages are the pages of every provider, by path.
	pages map[string]page
}

func newServer(cfg config, clock clockwork.Clock) (*server, error) {
	idp, err := selectIdP(cfg)
	if err != nil {
		return nil, err
	}

	sp, err := saml.New(saml.Config{
		EntityID: cfg.baseURL + spMetadataPath,
		ACSURL:   cfg.baseURL + acsPath,
		Cert:     cfg.spCert,
		IdPs:     []saml.IdP{idp},
		Clock:    clock,
	})
	if err != nil {
		return nil, err
	}

	templates, err := parseTemplates()
	if err != nil {
		return nil, err
	}

	s := &server{
		cfg:       cfg,
		clock:     clock,
		sessions:  session.NewStore(clock.Now),
		saml:      sp,
		idpID:     idp.EntityID,
		irma:      irma.NewClient(cfg.irmaServerURL, cfg.irmaToken),
		templates: templates,
		surfnet:   surfnetProvider(cfg),
		pages:     map[string]page{},
	}
	s.providers = []provider{s.surfnet}
	if cfg.linkedInClientID != "" {
		s.oidc = newOIDCClient(cfg)
		s.providers = append(s.providers, linkedInProvider(cfg))
	}

	for _, prefix := range []string{"", prefixNL, prefixEN} {
		for _, p := range s.providers {
			pg := page{provider: p, prefix: prefix}
			s.pages[pg.path()] = pg
		}
	}

	return s, nil
}

// selectIdP picks the identity provider that users are sent to. It is the
// only one whose responses are accepted.
func selectIdP(cfg config) (saml.IdP, error) {
	idps, err := saml.LoadIdPs(cfg.idpMetadata)
	if err != nil {
		return saml.IdP{}, fmt.Errorf("IdP metadata %s: %w", cfg.idpMetadata, err)
	}

	if cfg.idpEntityID == "" && len(idps) == 1 {
		return idps[0], nil
	}

	for _, idp := range idps {
		if idp.EntityID == cfg.idpEntityID {
			return idp, nil
		}
	}

	var ids []string
	for _, idp := range idps {
		ids = append(ids, idp.EntityID)
	}

	return saml.IdP{}, fmt.Errorf("set SURFCONEXT_IDP_ENTITY_ID to one of: %s", strings.Join(ids, ", "))
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", healthy)
	mux.HandleFunc("GET /"+healthyPage, healthy)

	for _, dir := range []string{"css", "js", "fonts", "images"} {
		mux.Handle("GET /"+dir+"/", noListing(http.FileServerFS(assetFS)))
	}

	for path, pg := range s.pages {
		mux.HandleFunc("GET "+path+"{$}", func(w http.ResponseWriter, r *http.Request) { s.page(w, r, pg) })
	}
	for _, prefix := range []string{prefixNL, prefixEN} {
		mux.HandleFunc("GET "+prefix+"/{$}", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, prefix+s.surfnet.path, http.StatusMovedPermanently)
		})
	}

	mux.HandleFunc("GET "+spMetadataPath, s.spMetadata)
	mux.HandleFunc("POST "+acsPath, s.acs)
	if s.oidc != nil {
		mux.HandleFunc("GET "+linkbackPath, s.linkedInCallback)
	}

	return nosniff(mux)
}

func healthy(w http.ResponseWriter, r *http.Request) {
	http.ServeFileFS(w, r, assetFS, healthyPage)
}

func nosniff(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

// noListing hides directory listings of the embedded assets.
func noListing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// sessionID returns the ID in the session cookie.
func sessionID(r *http.Request) string {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		return ""
	}

	return cookie.Value
}

// setSessionCookie sends the cookie. It must survive the cross-site POST of
// the SAML response to the ACS, hence SameSite=None.
func setSessionCookie(w http.ResponseWriter, id string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    id,
		Path:     "/",
		MaxAge:   int(session.MaxAge.Seconds()),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	})
}

// ensureSession returns the ID of the user's session, starting one if needed.
func (s *server) ensureSession(w http.ResponseWriter, r *http.Request) (string, error) {
	if id := sessionID(r); id != "" {
		if _, ok := s.sessions.Get(id); ok {
			return id, nil
		}
	}

	id, err := s.sessions.Create()
	if err != nil {
		return "", err
	}
	setSessionCookie(w, id)

	return id, nil
}

// finishLogin records what the provider released, moves the session to a new
// ID and sends the user back to the page they started from. The pending login
// must already have been consumed.
func (s *server) finishLogin(w http.ResponseWriter, r *http.Request, id string, p provider, pending session.Pending, attributes []credential.Attribute) {
	updated := s.sessions.Update(id, func(d *session.Data) { d.Login[p.id] = attributes })
	rotated, rotatedOK := s.sessions.Rotate(id)
	if !updated || !rotatedOK {
		s.renderFailure(w, http.StatusBadRequest, s.pageFor(pending, p))
		return
	}

	setSessionCookie(w, rotated)
	http.Redirect(w, r, pending.ReturnTo, http.StatusSeeOther)
}

// pageFor is the page a login was started from.
func (s *server) pageFor(pending session.Pending, p provider) page {
	if pg, ok := s.pages[pending.ReturnTo]; ok {
		return pg
	}

	return page{provider: p}
}

var errNoLogin = errors.New("no login in progress")

// consumePending removes the pending login with the given token. A token that
// does not match leaves the session alone.
func (s *server) consumePending(id string, p provider, token string) error {
	var found bool

	s.sessions.Update(id, func(d *session.Data) {
		candidate, ok := d.Pending[p.id]
		if !ok || candidate.Token != token {
			return
		}
		found = true
		delete(d.Pending, p.id)
	})
	if !found {
		slog.Warn("login answer does not match a pending login", "provider", p.id)
		return errNoLogin
	}

	return nil
}
