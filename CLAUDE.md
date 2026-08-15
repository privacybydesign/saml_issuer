# saml_issuer — repo notes

## Docker build: composer require needs `--with-all-dependencies`

The runtime stage installs `cirrusidentity/simplesamlphp-module-authoauth2`
on top of the `cirrusid/simplesamlphp` base image, which ships its own
`composer.lock`. That lock pins transitive packages (e.g. `guzzlehttp/psr7`
at 2.9.0). A plain `composer require` is a partial update and will not move
those pins, so once a newer `guzzlehttp/guzzle` (pulled in via
`guzzle-cache-middleware`) requires `psr7 ^2.12`, the build fails with a
"fixed to 2.9.0 (lock file version) by a partial update" resolver error.

Keep the `--with-all-dependencies` flag on that `composer require` so the
transitive deps can be upgraded together. This also lets composer pick
`authoauth2 v5.2.0` + `firebase/php-jwt` v7, which are free of the php-jwt
security advisories that block the older `^5.5|^6` range.

No Docker daemon in CI-adjacent sandboxes: reproduce dependency resolution
locally by fetching the matching simplesamlphp tag's `composer.json` +
`composer.lock` and running `composer require --dry-run --ignore-platform-reqs`.
