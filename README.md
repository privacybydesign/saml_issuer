# saml_issuer

A Go service that issues Yivi credentials from attributes released by an
identity provider. It supports SURFconext (SAML 2.0) and LinkedIn (OpenID
Connect).

| Page | Credential |
| --- | --- |
| `/surfconext/` | `pbdf.pbdf.surfnet-2`, valid for 3 months |
| `/linkedin/` | `pbdf.pbdf.linkedin`, valid for 1 year |

Both pages also exist under `/nl/` and `/en/`. Dutch is the default.

Other endpoints:

| Path | Purpose |
| --- | --- |
| `/simplesaml/module.php/saml/sp/metadata/surfnet` | SP metadata, and the SP entity ID |
| `/simplesaml/module.php/saml/sp/saml2-acs.php/surfnet` | SAML assertion consumer service (HTTP-POST) |
| `/simplesaml/module.php/authoauth2/linkback` | LinkedIn redirect URI |
| `/` and `/healthy.html` | Health probes, 200 |

These paths are the ones SimpleSAMLphp used, so the SURFconext and LinkedIn
registrations stay as they are.

The service listens on plain HTTP on port 8080. TLS is terminated by the
ingress.

## Configuration

| Variable | Description | Required |
| --- | --- | --- |
| `HOST_NAME` | Host the issuer is served on. Builds the entity ID and the ACS URL | yes |
| `IRMA_SERVER_URL` | URL of the IRMA server this issuer talks to | yes |
| `API_TOKEN` | Token for the IRMA server, sent as the `Authorization` header | |
| `IRMA_ISSUER_ID` | Scheme manager and issuer ID, e.g. `irma-demo.sidn-pbdf`. Defaults to `pbdf.pbdf` | |
| `IRMA_SDJWT_BATCH_SIZE_SURFCONTEXT` | SD-JWT batch size for the SURFconext credential. Defaults to `200` | |
| `HTTP_PROTOCOL` | `http` or `https`. Defaults to `https` | |
| `SIMPLESAMLPHP_CERT_DIR` | Directory with `saml.crt`, the SP certificate. Defaults to `/var/simplesamlphp/cert` | |
| `SIMPLESAMLPHP_METADATA_DIR` | Directory with `saml20-idp-remote.xml`, the IdP metadata. Defaults to `/var/simplesamlphp/metadata` | |
| `SURFCONEXT_IDP_ENTITY_ID` | IdP to send users to. Only needed when the metadata holds more than one | |
| `LINKEDIN_CLIENT_ID` | LinkedIn client ID. LinkedIn is off when unset | |
| `LINKEDIN_CLIENT_SECRET` | LinkedIn client secret | |

`sample.env` lists the same variables.

### IdP metadata

`saml20-idp-remote.xml` is standard SAML metadata: one `EntityDescriptor`, or an
`EntitiesDescriptor` holding several. SURFconext publishes it. The signing
certificates of the IdP that users are sent to are the only keys a SAML response
is accepted with. When the file holds several IdPs, set `SURFCONEXT_IDP_ENTITY_ID`.
There is no metadata refresh: replace the file when SURFconext rotates its key.

The certificate must be valid at the time of login. goxmldsig checks its
validity period, which SimpleSAMLphp did not.

### SP certificate

Only `saml.crt` is read. It is published in the SP metadata. Requests are not
signed and assertions are not encrypted, so the private key (`saml.pem`) is not
needed by the service.

## Sessions

Login state is kept in memory, behind an unguessable session cookie
(`SameSite=None; Secure`, so it is sent with the cross-site POST to the ACS).
A session lasts 30 minutes until the user logs in. At login it is replaced by a
new ID and lasts 8 hours from then. The store holds at most 50,000 sessions: when
it is full, the oldest session without a login is dropped to make room, and
logins are kept. Because the state is in memory, run one replica.

## SAML validation

A response is accepted only if all of these hold. The code is in
`internal/saml`.

- It is at most 128 KiB of plain XML, without comments or a DTD.
- It holds exactly one `Assertion`, a child of the `Response`, that carries its
  own enveloped signature referencing the assertion's ID. A signed `Response`
  around an unsigned `Assertion` is rejected.
- The signature verifies with the IdP's pinned certificate. A certificate in
  `KeyInfo` is not trusted, and is rejected unless it is the pinned one.
- Issuer, audience, destination and recipient match, the status is Success, and
  `NotBefore` and `NotOnOrAfter` hold with 180 s of clock skew.
- `InResponseTo` is the ID of the AuthnRequest stored in the user's session.
  The ID is used once.

Attributes are read only from the assertion that was verified. The NameID is not
used.

## Run

Build and test:

```bash
go vet ./...
go test ./...
```

The pages need `js/jquery.min.js` and `js/yivi.js`, which the Docker build
creates with `yarn install && ./build.sh`. Without them the issue page cannot
start the Yivi popup, but the tests do not need them.

With Docker:

```bash
./setup-local.sh
# put the IdP metadata in config/metadata/saml20-idp-remote.xml
docker compose up --build
```

SURFconext sends users back to the ACS URL it has registered for the entity ID,
so a full SURFconext login needs a public HTTPS host. The `__Host-session`
cookie is `Secure`: browsers accept it on `http://localhost`, Safari does not.

## Documentation

- [SURFconext](https://servicedesk.surf.nl/wiki/spaces/IAM/pages/128909810/SURFconext+for+Service+Providers)
- [LinkedIn](https://learn.microsoft.com/en-us/linkedin/shared/authentication/authentication)
