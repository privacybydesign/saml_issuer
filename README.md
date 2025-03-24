# saml_issuer

A SimpleSAMLphp library for issuing attributes coming from a SAML Identity Provider.
Currently supports SURFconext, eduGAIN and LinkedIn.

# Environment variables
See `sample.env` for a list of required environment variables. Copy the file to `.env` and fill in the secrets.
A list of optional environment variables for SimpleSAMLphp Docker image [can be found here](https://github.com/cirrusidentity/docker-simplesamlphp/tree/master?tab=readme-ov-file#environmental-variables).

| Variable | description |
| --- | --- |
| `HTTP_PROTOCOL` | 'http' or 'https' |
| `HOST_NAME` | Url on which this issuer runs |
| `IRMA_SERVER_URL` | Url for the IRMA server this issuer talks to |
| `API_TOKEN` | API token for the IRMA server |
| `LINKEDIN_CLIENT_ID` | LinkedIn Client ID |
| `LINKEDIN_CLIENT_SECRET` | LinkedIn Client Secret |

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