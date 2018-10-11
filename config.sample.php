<?php
define('LANG', 'en');

define('ROOT_DIR', __DIR__ . '/../../../');
define('LOG_FILE', ROOT_DIR . 'logs/php.log');
define('KEY_FILE', ROOT_DIR . 'saml-sk.pem');

require_once ROOT_DIR . 'php/vendor/autoload.php';
require_once ROOT_DIR . 'simplesamlphp/lib/_autoload.php';
