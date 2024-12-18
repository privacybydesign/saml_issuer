# saml_issuer

A SimpleSAMLphp library for issuing attributes coming from a SAML Identity Provider.
Currently supports SURFconext, eduGAIN, LinkedIn and Twitter.

# Environment variables
See `sample.env` for a list of required environment variables. Copy the file to `.env` and fill in the secrets.

| Variable | description |
| --- | --- |
| `HOST_NAME` | Url on which this issuer runs |
| `IRMA_SERVER_URL` | Url for the IRMA server this issuer talks to |
| `API_TOKEN` | API token for the IRMA server |
| `LINKEDIN_CLIENT_ID` | LinkedIn Client ID |
| `LINKEDIN_CLIENT_SECRET` | LinkedIn Client Secret |
| `TWITTER_CLIENT_ID` | Twitter Client ID |
| `TWITTER_CLIENT_SECRET` | Twitter Client Secret |

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
docker run --env-file .env -p 8080:80 saml-issuer
```