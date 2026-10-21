package main

import (
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// publicAssets is everything that may be reachable over HTTP besides the
// pages and endpoints.
var publicAssets = map[string]bool{
	"healthy.html":                            true,
	"css/style.css":                           true,
	"js/enroll.js":                            true,
	"js/jquery.min.js":                        true,
	"js/yivi.js":                              true,
	"images/button-primary.svg":               true,
	"images/done.png":                         true,
	"fonts/alexandria-v1-latin-600.woff":      true,
	"fonts/alexandria-v1-latin-600.woff2":     true,
	"fonts/alexandria-v1-latin-regular.woff":  true,
	"fonts/alexandria-v1-latin-regular.woff2": true,
	"fonts/open-sans-v34-latin-700.woff":      true,
	"fonts/open-sans-v34-latin-700.woff2":     true,
	"fonts/open-sans-v34-latin-regular.woff":  true,
	"fonts/open-sans-v34-latin-regular.woff2": true,
}

func TestEmbeddedAssets(t *testing.T) {
	err := fs.WalkDir(assetFS, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		assert.True(t, publicAssets[name], "%s is embedded but not on the list of public assets", name)

		return nil
	})

	require.NoError(t, err)
}

func TestPublicAssetsAreServed(t *testing.T) {
	h := newHarness(t, harnessOptions{})

	for name := range publicAssets {
		if strings.HasPrefix(name, "js/") && name != "js/enroll.js" {
			continue // built by the Docker image, not part of the repository
		}

		response := h.get("/" + name)

		assert.Equal(t, http.StatusOK, response.StatusCode, name)
	}
}

func TestRepoFilesNotServed(t *testing.T) {
	h := newHarness(t, harnessOptions{})

	var targets []string
	for _, root := range []string{".", "internal", "templates", ".github"} {
		entries, err := os.ReadDir(root)
		if os.IsNotExist(err) {
			continue
		}
		require.NoError(t, err)

		for _, entry := range entries {
			targets = append(targets, path.Join(root, entry.Name()))
		}
	}
	require.NotEmpty(t, targets)

	for _, target := range targets {
		if publicAssets[target] {
			continue
		}

		for _, prefix := range []string{"/", "/css/", "/js/", "/nl/", "/en/"} {
			response := h.get(prefix + strings.TrimPrefix(target, "./"))

			assert.NotEqual(t, http.StatusOK, response.StatusCode, prefix+target)
		}
	}
}

func TestSensitivePaths(t *testing.T) {
	h := newHarness(t, harnessOptions{})

	for _, target := range []string{
		"/.git/config", "/.git/HEAD", "/.env", "/go.mod", "/go.sum", "/Dockerfile", "/Dockerfile.test",
		"/docker-compose.yml", "/README.md", "/CLAUDE.md", "/main.go", "/server.go", "/config.go",
		"/server_test.go", "/templates/login-nl.html", "/templates/", "/internal/saml/validate.go",
		"/node_modules/jquery/package.json", "/vendor/modules.txt", "/package.json", "/yarn.lock",
		"/css/", "/js/", "/fonts/", "/images/", "/css", "/js", "/css/bootstrap.min.css", "/js/verify.js",
		"/css/../go.mod", "/css/..%2fgo.mod", "/css/%2e%2e/go.mod", "/js/../main.go", "//etc/passwd",
		"/saml.crt", "/saml.pem", "/saml20-idp-remote.xml", "/cert/saml.pem", "/config/",
		"/simplesaml/", "/simplesaml/admin/", "/simplesaml/module.php/core/frontpage_welcome.php",
		"/simplesaml/module.php/cron/cron.php", "/simplesaml/saml2/idp/metadata.php",
		"/edugain/", "/index.php", "/config.php",
	} {
		response := h.get(target)

		assert.GreaterOrEqual(t, response.StatusCode, http.StatusMultipleChoices, target)
		assert.NotEqual(t, http.StatusOK, response.StatusCode, target)
		assert.NotContains(t, body(t, response), "module github.com", target)
	}
}

func TestEmbeddedMatchDisk(t *testing.T) {
	for name := range publicAssets {
		if strings.HasPrefix(name, "js/") && name != "js/enroll.js" {
			continue
		}

		embedded, err := assetFS.ReadFile(name)
		require.NoError(t, err, name)
		onDisk, err := os.ReadFile(name)
		require.NoError(t, err, name)

		assert.Equal(t, onDisk, embedded, name)
	}
}
