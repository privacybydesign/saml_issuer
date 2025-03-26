# saml_issuer

A SimpleSAMLphp library for issuing attributes coming from a SAML Identity Provider.
Currently supports SURFconext, eduGAIN and LinkedIn.

# Environment variables
See `sample.env` for a list of environment variables. Copy the file to `.env` and fill in the secrets.
A list of optional environment variables (expect for `APACHE_CERT_NAME`, which) for SimpleSAMLphp Docker image [can be found here](https://github.com/cirrusidentity/docker-simplesamlphp/tree/master?tab=readme-ov-file#environmental-variables).

| Variable | description | required |
| --- | --- | --- |
| `HOST_NAME` | Url on which this issuer runs | ✅ |
| `IRMA_SERVER_URL` | Url for the IRMA server this issuer talks to | ✅ |
| `API_TOKEN` | API token for the IRMA server | ✅ |
| `LINKEDIN_CLIENT_ID` | LinkedIn Client ID | ✅ |
| `LINKEDIN_CLIENT_SECRET` | LinkedIn Client Secret | ✅ |
| `APACHE_CERT_PATH` | Path to the SSL certificate Apache will use for the host. The cert needs to be mounted inside the pod. | ✅ |
| `APACHE_KEY_PATH` | Path to the SSL key Apache will use for the host. The key needs to be mounted inside the pod. | ✅ |
| `HTTP_PROTOCOL` | 'http' or 'https'; defaults to 'https' if unspecified | |
| `SIMPLESAMLPHP_CONFIG_DIR` | Path to load SimpleSAMLphp config files from. Defaults to `/var/simplesamlphp/config` if unspecified | |
| `SIMPLESAMLPHP_METADATA_DIR` | Path to load SimpleSAMLphp default metadata files from. Defaults to `/var/simplesamlphp/metadata` if unspecified | |
| `SIMPLESAMLPHP_CERT_DIR` | Path to load SimpleSAMLphp certificate files from. Defaults to `/var/simplesamlphp/cert` if unspecified | |

## Build
 * `composer install`
 * `yarn install`
 * `./build.sh`
 * `./install-simplesamlphp`
 * Move the simplesamlphp directory to your php root and set php root in `config.php` of saml_issuer
 * Add the API keys in simplesamlphp: `config/authsources.php`
 * Host both the `simplesamlphp/www` and the `saml_issuer` directory on a web server
   (the url of `simplesamlphp/www` must be set as basepathurl in `config/config.php` of simplesamlphp)

## Run with docker
```bash
docker build . -t saml-issuer
```

```bash
docker run --env-file .env -p 8080:8080 saml-issuer
```