# saml_issuer

A SimpleSAMLphp library for issuing attributes coming from a SAML Identity Provider.
Currently supports SURFconext and LinkedIN.

# Environment variables
See `sample.env` for a list of environment variables. Copy the file to `.env` and fill in the secrets.
A list of optional environment variables (expect for `APACHE_CERT_NAME`, which) for SimpleSAMLphp Docker image [can be found here](https://github.com/cirrusidentity/docker-simplesamlphp/tree/master?tab=readme-ov-file#environmental-variables).

| Variable | description | required |
| --- | --- | --- |
| `HOST_NAME` | Url on which this issuer runs | ✅ |
| `IRMA_SERVER_URL` | Url for the IRMA server this issuer talks to | ✅ |
| `API_TOKEN` | API token for the IRMA server | ✅ |
| `SURFCONEXT_IDP_METADATA_URL` | URL to the metadata of SURFconext IdP | ✅ |
| `LINKEDIN_CLIENT_ID` | LinkedIN Client ID | ✅ |
| `LINKEDIN_CLIENT_SECRET` | LinkedIN Client Secret | ✅ |
| `APACHE_CERT_PATH` | Path to the SSL certificate Apache will use for the host. The cert needs to be mounted inside the pod. | ✅ |
| `APACHE_KEY_PATH` | Path to the SSL key Apache will use for the host. The key needs to be mounted inside the pod. | ✅ |
| `IRMA_ISSUER_ID` | Combination of the schememanager + IssuerID (i.e. `irma-demo.pbdf`). Defaults to `pbdf.pbdf` if unspecified | |
| `HTTP_PROTOCOL` | 'http' or 'https'; defaults to 'https' if unspecified | |
| `SIMPLESAMLPHP_CONFIG_DIR` | Path to load SimpleSAMLphp config files from. Defaults to `/var/simplesamlphp/config` if unspecified | |
| `SIMPLESAMLPHP_METADATA_DIR` | Path to load SimpleSAMLphp default metadata files from. Defaults to `/var/simplesamlphp/metadata` if unspecified | |
| `SIMPLESAMLPHP_CERT_DIR` | Path to load SimpleSAMLphp certificate files from. Defaults to `/var/simplesamlphp/cert` if unspecified | |
| `ENABLE_IRMA_LOGGING` | Enable logging of SAML -> IRMA mapping data. Default: false. Note: only use for debugging purpose, as this logs personal data! | |

## Run with docker

### Build image
```bash
docker build . -t saml-issuer
```

In case you're using the wildcard testing certificate (see below), make sure to always build using `--pull` to get the latest Cirrus image (so Docker updates the cached image if needed). This way, the testing certificate will always be valid.
```bash
docker build . -t saml-issuer --pull
```

### Configure `.env` file
Create a `.env` file for the configuration of the container.
SimpleSAMLphp config and metadata template files can be found in the [SimpleSAMLphp Github repository](https://github.com/simplesamlphp/simplesamlphp).

### Obtain SSL key/cert
In order to run the image, you need a valid SSL key/cert combo for the `HOST_NAME` you want to run.
This is also true for local development. You can provide your own keys via volume-mounts, or use the local-stack key as described below.

The image is created with a wildcard certificate for testing, which can be found at the locations below.
If you use this cert, make sure the `HOST_NAME` matches `*.local.stack-dev.cirrusidentity.com`, where `*` can be any valid subdomain.
The certificate expires every 90 days. If it is expired, you can delete the 
- APACHE_CERT_PATH: /etc/ssl/certs/local-stack-dev.pem
- APACHE_KEY_PATH: /etc/ssl/private/local-stack-dev.key

### Change HOSTS file
Make sure to add an entry to your HOSTS file, pointing the `HOST_NAME` configured in the `.env` file to `127.0.0.1`.

### Run container
Run the container with volumes mounted like below (make sure to configure the correct `SIMPLESAMLPHP_*_DIR` and `APACHE_*_PATH` variables in the `.env` file).
Replace the {{HOST_NAME}} variable below, with the hostname from the `.env` file.

```bash
docker run  -d \
    --env-file .env \
    -p 443:443 saml-issuer \
    --name saml_issuer \
    --hostname "{{HOST_NAME}}" \
    -v /local/path/to/simplesamlphp/config:/var/simplesamlphp/config \
    -v /local/path/to/simplesamlphp/metadata:/var/simplesamlphp/metadata \
    -v /local/path/to/simplesamlphp/metadata_certs:/var/simplesamlphp/cert \
    -v /local/path/to/.secret/tls.pem:/etc/ssl/certs/tls.pem \
    -v /local/path/to/.secret/tls.key:/etc/ssl/private/tls.key \
    saml-issuer
```

# Documentation
Documentation regarding the supported IdPs can be found:
- [SURFconext](https://servicedesk.surf.nl/wiki/spaces/IAM/pages/128909810/SURFconext+for+Service+Providers)
- [LinkedIN](https://learn.microsoft.com/en-us/linkedin/shared/authentication/authentication)