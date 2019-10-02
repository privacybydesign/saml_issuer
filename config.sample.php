<?php
define('LANG', 'en');

define('ROOT_DIR', __DIR__ . '/../../../');
define('LOG_FILE', ROOT_DIR . 'logs/php.log');
define('IRMA_SERVER_API_TOKEN', '');
define('IRMA_SERVER_URL', 'http://localhost:8088');

require_once ROOT_DIR . 'php/vendor/autoload.php';
require_once ROOT_DIR . 'simplesamlphp/lib/_autoload.php';
