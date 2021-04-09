<?php

require_once '../config.php';

define('PROVIDER', 'linkedin');
define('PROVIDER_NAME', 'LinkedIn');
define('CREDENTIAL', IRMA_ISSUER_ID . '.linkedin');
define('IRMA_NAME_ATTRIBUTE', 'familyname');
define('VALIDITY', '+1 year');

$MAP_IRMA_SAML_ATTRIBUTES = [
    'firstname'   => 'linkedin.firstName',
    'familyname'  => 'linkedin.lastName',
    'email'       => 'linkedin.emailAddress',
];

require '../index.php';
