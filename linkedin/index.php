<?php

require_once '../config.php';

define('PROVIDER', 'linkedin');
define('PROVIDER_NAME', 'LinkedIn');
define('CREDENTIAL', IRMA_ISSUER_ID . '.linkedin');
define('IRMA_NAME_ATTRIBUTE', 'fullname');
define('VALIDITY', '+1 year');

$MAP_IRMA_SAML_ATTRIBUTES = [
    'fullname'    => 'oid.name',
    'firstname'   => 'oidc.given_name',
    'familyname'  => 'oidc.family_name',
    'email'       => 'oidc.email',
];

require '../index.php';
