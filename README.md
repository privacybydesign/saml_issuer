# saml_issuer

A SimpleSAMLphp library for issuing attributes coming from a SAML Identity Provider.
Currently supports SURFconext, eduGAIN, LinkedIn and Twitter.

# Environment variables
| Variable | description |
| --- | --- |
| `HOST_NAME` | Url on which this issuer runs |
| `IRMA_SERVER_URL` | Url for the IRMA server this issuer talks to |
| `IRMA_SERVER_API_TOKEN` | API token for the IRMA server |

## Build
 * `composer install`
 * `yarn install`
 * `./build.sh`
 * `./install-simplesamlphp`
 * Move the simplesamlphp directory to your php root and set php root in `config.php` of saml_issuer
 * Add the API keys in simplesamlphp: `config/authsources.php`
 * Host both the `simplesamlphp/www` and the `saml_issuer` directory on a web server
   (the url of `simplesamlphp/www` must be set as basepathurl in `config/config.php` of simplesamlphp)
