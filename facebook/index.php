<?php

define('PROVIDER', 'facebook');
define('PROVIDER_NAME', 'Facebook');
define('SERVER_NAME', 'facebook_enroll');
define('CREDENTIAL', 'pbdf.pbdf.facebook');
define('IRMA_NAME_ATTRIBUTE', 'fullname');
define('VERIFY_NAME_LABEL', 'Facebook full name');

define('MAP_IRMA_SAML_ATTRIBUTES', [
    'id'       => 'facebook.id',
    'fullname' => 'facebook.name',
]);

define('ATTRIBUTE_HUMAN_NAMES', [
    'id'       => 'ID',
    'fullname' => 'Volledige naam',
]);

require '../index.php';
