<?php
define('LANG', 'en');

define('ROOT_DIR', __DIR__ . '');
define('LOG_FILE', ROOT_DIR . 'logs/php.log');
define('IRMA_SERVER_API_TOKEN', getenv('API_TOKEN'));
define('IRMA_SERVER_URL', getenv('IRMA_SERVER_URL'));

define('IRMA_ISSUER_ID', 'pbdf-staging.sidn-pbdf');

require_once ROOT_DIR . '/vendor/autoload.php';
require_once ROOT_DIR . '/simplesamlphp/lib/_autoload.php';
