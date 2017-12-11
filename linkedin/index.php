<?php

define('PROVIDER', 'linkedin');
define('PROVIDER_NAME', 'LinkedIn');
define('SERVER_NAME', 'linkedin_enroll');
define('CREDENTIAL', 'pbdf.pbdf.linkedin');
define('IRMA_NAME_ATTRIBUTE', 'fullname');
define('PROFILE_URL_REGEX', '%^https://www.linkedin.com/in/([a-zA-Z0-9]+)$%');
define('VERIFY_NAME_LABEL', 'LinkedIn full name');

$MAP_IRMA_SAML_ATTRIBUTES = [
    'fullname'    => 'linkedin.formattedName',
    'firstname'   => 'linkedin.firstName',
    'familyname'  => 'linkedin.lastName',
    'email'       => 'linkedin.emailAddress',
    'profileurl'  => 'linkedin.publicProfileUrl',
];

require '../index.php';
