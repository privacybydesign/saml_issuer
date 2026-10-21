package main

import (
	"bytes"
	"embed"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/privacybydesign/saml_issuer/internal/credential"
)

//go:embed templates/*.html
var templateFS embed.FS

const (
	langNL = "nl"
	langEN = "en"

	prefixNL = "/" + langNL
	prefixEN = "/" + langEN
)

var attributeNames = map[string]map[string]string{
	langNL: {
		"id":         "Gebruiker ID",
		"fullid":     "Volledig ID",
		"fullname":   "Volledige naam",
		"firstname":  "Voornaam",
		"familyname": "Achternaam",
		"email":      "E-mailadres",
		"institute":  "Instituut",
		"type":       "Type",
	},
	langEN: {
		"id":         "User ID",
		"fullid":     "Full ID",
		"fullname":   "Full name",
		"firstname":  "First name",
		"familyname": "Family name",
		"email":      "Email address",
		"institute":  "Institute",
		"type":       "Type",
	},
}

var failureMessages = map[string]string{
	langNL: "Het inloggen kon niet worden voltooid. Probeer het opnieuw.",
	langEN: "The login could not be completed. Please try again.",
}

// page is a provider's page under a language prefix. An empty prefix is Dutch.
type page struct {
	provider provider
	prefix   string
}

func (p page) lang() string {
	if p.prefix == prefixEN {
		return langEN
	}

	return langNL
}

func (p page) path() string { return p.prefix + p.provider.path }

type attributeRow struct {
	Label string
	Value string
}

type pageData struct {
	Lang          string
	Provider      string
	ProviderName  string
	NLURL         string
	ENURL         string
	IrmaServerURL string
	Attributes    []attributeRow
	FirstName     string
	Message       string
	RetryURL      string
}

func parseTemplates() (*template.Template, error) {
	return template.ParseFS(templateFS, "templates/*.html")
}

func (s *server) pageData(pg page, rawQuery string) pageData {
	suffix := ""
	if rawQuery != "" {
		suffix = "?" + rawQuery
	}

	return pageData{
		Lang:          pg.lang(),
		Provider:      pg.provider.id,
		ProviderName:  pg.provider.name,
		NLURL:         prefixNL + pg.provider.path + suffix,
		ENURL:         prefixEN + pg.provider.path + suffix,
		IrmaServerURL: s.cfg.irmaServerURL,
	}
}

// render writes the named template for the page's language.
func (s *server) render(w http.ResponseWriter, status int, name string, pg page, data pageData) {
	var body bytes.Buffer
	if err := s.templates.ExecuteTemplate(&body, name+"-"+pg.lang()+".html", data); err != nil {
		slog.Error("render page", "template", name, "error", err)
		http.Error(w, msgInternalError, http.StatusInternalServerError)
		return
	}

	w.Header().Set(headerContentType, contentTypeHTML)
	w.WriteHeader(status)
	_, _ = w.Write(body.Bytes())
}

// renderFailure shows a generic error page, without saying why it failed.
func (s *server) renderFailure(w http.ResponseWriter, status int, pg page) {
	data := s.pageData(pg, "")
	data.Message = failureMessages[pg.lang()]
	data.RetryURL = pg.path() + loginQuery

	w.Header().Set(headerCacheControl, cacheNoStore)
	s.render(w, status, "error", pg, data)
}

func attributeRows(lang string, attributes []credential.Attribute) []attributeRow {
	rows := make([]attributeRow, 0, len(attributes))
	for _, attribute := range attributes {
		rows = append(rows, attributeRow{Label: attributeNames[lang][attribute.Key], Value: attribute.Value})
	}

	return rows
}
