<?php

require_once '../config.php';

define('PROVIDER', 'linkedin');
define('PROVIDER_NAME', 'LinkedIn');
define('CREDENTIAL', IRMA_ISSUER_ID . '.linkedin');
define('IRMA_NAME_ATTRIBUTE', 'name');
define('VALIDITY', '+1 year');

$MAP_IRMA_SAML_ATTRIBUTES = [
    'firstname'   => 'given_name',
    'familyname'  => 'family_name',
    'email'       => 'email',
];

require '../index.php';
