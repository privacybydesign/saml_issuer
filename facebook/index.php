<?php

require_once '../config.php';

define('PROVIDER', 'facebook');
define('PROVIDER_NAME', 'Facebook');
define('SERVER_NAME', 'facebook_enroll');
define('CREDENTIAL', IRMA_ISSUER_ID . '.facebook');
define('IRMA_NAME_ATTRIBUTE', 'fullname');
define('VALIDITY', '+1 year');

$MAP_IRMA_SAML_ATTRIBUTES = [
    'fullname'    => 'facebook.name',
    'firstname'   => 'facebook.first_name',
    'familyname'  => 'facebook.last_name',
    'email'       => 'facebook.email',
];

require '../index.php';
