# saml_issuer — repo notes

Go service (module root) that issues Yivi credentials from SURFconext (SAML) and
LinkedIn (OIDC). It replaced a SimpleSAMLphp/PHP service.

## Build & test

- `go vet ./... && go test ./...`. The Go version is set in `go.mod`.
- The tests run the whole service over TLS against a fake IdP, IRMA server and
  LinkedIn (`harness_test.go`); no network or Docker is needed.
- `js/jquery.min.js` and `js/yivi.js` are built by the Dockerfile's node stage
  (`build.sh`) and embedded. They are not in the repository.
- No Docker daemon in CI-adjacent sandboxes: build the image's steps by hand
  (`yarn install && ./build.sh`, then `CGO_ENABLED=0 go build`).

## Layout

- `internal/saml`: gosaml2 wrapper. `shape.go` runs before the library,
  `validate.go` after it.
- `internal/samltest`: builds signed SAML responses with generated keys. Used by
  the tests only.
- `internal/session`, `internal/credential`, `internal/irma`: sessions,
  attribute mapping, IRMA session call.
- Root: config, routes, handlers, `templates/`, `oidc.go`.

## Gotchas

- Do not loosen `internal/saml`. A mistake there fails silently: an attacker
  gets a valid `surfnet-2` credential for any institution. Every relaxation
  needs a test in `validate_test.go` or `xsw_test.go` that shows what is let in.
- gosaml2 alone rejects the XSW1 to XSW8 variants in `xsw_test.go`, but accepts a
  signed `Response` around an unsigned `Assertion`, duplicate IDs, comments, DTDs
  and deflated input. `checkShape` closes those. Keep both layers.
- gosaml2 checks `SubjectConfirmationData/@NotOnOrAfter` without skew, so the
  library runs with a clock set back by `saml.ClockSkew`. The AuthnRequest is
  built by a second library instance with the real clock. `Conditions` are checked
  in `validate.go`.
- goxmldsig rejects a pinned certificate outside its validity period.
- `etree` is pinned to v1.7.0, the version gosaml2 and goxmldsig declare. Do not
  float it to `@latest`.
- The embedded files are listed in `assetFS` (`server.go`) and in
  `static_test.go`. Add a file to both, or the test fails. The tests also request
  every file in the repository and expect it not to be served.
- The session store is in memory, hence one replica. The cookie is
  `SameSite=None` because the SAML response is a cross-site POST.
- Anyone can create a session. One without a login lives 30 minutes and is the
  first to be evicted when the store is full. `Rotate` (the login) grants the 8
  hours. Do not give a new session the full lifetime.
- The ops repo (`saml_issuer_ops`) deploys on a PR, so changes to it are listed in
  the PR description instead of opened as a PR.
