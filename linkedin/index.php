<?php

define('PROVIDER', 'linkedin');
define('PROVIDER_NAME', 'LinkedIn');
define('SERVER_NAME', 'linkedin_enroll');
define('CREDENTIAL', 'pbdf.pbdf.linkedin');
define('IRMA_NAME_ATTRIBUTE', 'fullname');
define('VERIFY_NAME_LABEL', 'LinkedIn full name');

$MAP_IRMA_SAML_ATTRIBUTES = [
    'id'          => 'linkedin.id',
    'firstname'   => 'linkedin.firstName',
    'familyname'  => 'linkedin.lastName',
    'email'       => 'linkedin.emailAddress',
];

$ATTRIBUTE_HUMAN_NAMES = [
    'id'          => 'ID',
    'firstname'   => 'Voornaam',
    'familyname'  => 'Achternaam',
    'email'       => 'E-mailadres',
];

require '../index.php';
