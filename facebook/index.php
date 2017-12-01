<?php

define('PROVIDER', 'facebook');
define('PROVIDER_NAME', 'Facebook');
define('SERVER_NAME', 'facebook_enroll');
define('CREDENTIAL', 'pbdf.pbdf.facebook');
define('IRMA_NAME_ATTRIBUTE', 'fullname');
define('VERIFY_NAME_LABEL', 'Facebook full name');

define('MAP_IRMA_SAML_ATTRIBUTES', [
    'id'          => 'facebook.id',
    'fullname'    => 'facebook.name',
    'firstname'   => 'facebook.first_name',
    'familyname'  => 'facebook.last_name',
    'email'       => 'facebook.email',
    'dateofbirth' => 'facebook.birthday',
]);

define('ATTRIBUTE_HUMAN_NAMES', [
    'id'          => 'ID',
    'fullname'    => 'Volledige naam',
    'firstname'   => 'Voornaam',
    'familyname'  => 'Achternaam',
    'email'       => 'E-mailadres',
    'dateofbirth' => 'Geboortedatum',
]);

require '../index.php';
