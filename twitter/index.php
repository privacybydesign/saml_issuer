<?php

define('PROVIDER', 'twitter');
define('PROVIDER_NAME', 'Twitter');
define('SERVER_NAME', 'twitter_enroll');
define('CREDENTIAL', 'pbdf.pbdf.twitter');
define('IRMA_NAME_ATTRIBUTE', 'fullname');
define('PROFILE_URL_FORMAT', 'https://twitter.com/:username:');
define('VERIFY_NAME_LABEL', 'Twitter full name');

$MAP_IRMA_SAML_ATTRIBUTES = [
    'id'          => 'twitter.id_str',
    'username'    => 'twitter.screen_name',
    'fullname'    => 'twitter.name',
    'email'       => 'twitter.email',
];

require '../index.php';
