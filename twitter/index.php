<?php

define('PROVIDER', 'twitter');
define('PROVIDER_NAME', 'Twitter');
define('SERVER_NAME', 'twitter_enroll');
define('CREDENTIAL', 'irma-demo.nijmegen.bsn');
define('IRMA_NAME_ATTRIBUTE', 'bsn');
define('PROFILE_URL_FORMAT', 'https://twitter.com/:username:');
define('VALIDITY', '+1 year');

$MAP_IRMA_SAML_ATTRIBUTES = [
    'username'    => 'twitter.screen_name',
    'fullname'    => 'twitter.name',
    'email'       => 'twitter.email',
];

require '../index.php';
