<?php

define('PROVIDER', 'linkedin');
define('PROVIDER_NAME', 'LinkedIn');
define('SERVER_NAME', 'linkedin_enroll');
define('CREDENTIAL', 'pbdf.pbdf.linkedin');
define('IRMA_NAME_ATTRIBUTE', 'familyname');
define('VALIDITY', '+1 year');

$MAP_IRMA_SAML_ATTRIBUTES = [
    'firstname'   => 'linkedin.firstName',
    'familyname'  => 'linkedin.lastName',
    'email'       => 'linkedin.emailAddress',
];

require '../index.php';
