<?php

require_once '../config.php';

define('PROVIDER', 'twitter');
define('PROVIDER_NAME', 'Twitter');
define('CREDENTIAL', IRMA_ISSUER_ID . '.twitter');
define('IRMA_NAME_ATTRIBUTE', 'fullname');
define('PROFILE_URL_FORMAT', 'https://twitter.com/:username:');
define('VALIDITY', '+1 year');

$MAP_IRMA_SAML_ATTRIBUTES = [
    'username'    => 'twitter.screen_name',
    'fullname'    => 'twitter.name',
    'email'       => 'twitter.email',
];

require '../index.php';
