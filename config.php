<?php

// Only set language if it is not already set in the querystring
if(isset($_GET['lang'])) {
    define('LANG', $_GET['lang']);
} else {
    define('LANG', 'nl');
}

define('ROOT_DIR', __DIR__ . '');
define('LOG_FILE', '/var/log/simplesamlphp/yivi.log');
define('ENABLE_IRMA_LOGGING', strtolower(getenv('ENABLE_IRMA_LOGGING')) === "true");
define('IRMA_SERVER_API_TOKEN', getenv('API_TOKEN'));
define('IRMA_SERVER_URL', getenv('IRMA_SERVER_URL'));

if (!getenv('IRMA_ISSUER_ID')) {
    fwrite(STDOUT, "WARNING: environment variable IRMA_ISSUER_ID not set, using default value 'pbdf.pbdf'." . PHP_EOL);
}
define('IRMA_ISSUER_ID', !getenv('IRMA_ISSUER_ID') ? 'pbdf.pbdf' : getenv('IRMA_ISSUER_ID'));

require_once getenv('SSP_DIR') . '/vendor/autoload.php';
require_once getenv('SSP_DIR') . '/lib/_autoload.php';
