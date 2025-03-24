<?php
define('LANG', 'en');

define('ROOT_DIR', __DIR__ . '');
define('LOG_FILE', ROOT_DIR . 'logs/php.log');
define('IRMA_SERVER_API_TOKEN', getenv('API_TOKEN'));
define('IRMA_SERVER_URL', getenv('IRMA_SERVER_URL'));

define('IRMA_ISSUER_ID', 'pbdf-staging.pbdf');      # TODO: differentiate between staging and production

require_once getenv('SSP_DIR') . '/vendor/autoload.php';
require_once getenv('SSP_DIR') . '/lib/_autoload.php';
