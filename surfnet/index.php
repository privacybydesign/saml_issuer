<?php

define('ROOT_DIR', __DIR__ . '/../../../../');

require_once ROOT_DIR . 'php/vendor/autoload.php';
require_once ROOT_DIR . 'simplesamlphp/lib/_autoload.php';

define('PROVIDER',  'surfnet');
define('PROVIDER_NAME', 'Surfnet');
define('SERVER_NAME', 'surfnet_enroll');
define('CREDENTIAL', 'pbdf.pbdf.surfnet');
define('IRMA_NAME_ATTRIBUTE', 'fullname');
define('VERIFY_NAME_LABEL', 'Surfnet full name');

define('MAP_IRMA_SAML_ATTRIBUTES', [
    'institute'  => 'urn:mace:terena.org:attribute-def:schacHomeOrganization';
    'type'       => 'urn:mace:dir:attribute-def:eduPersonAffiliation';
    'id'         => 'urn:mace:dir:attribute-def:uid';
    'fullname'   => 'urn:mace:dir:attribute-def:cn';
    'firstname'  => 'urn:mace:dir:attribute-def:givenName';
    'familyname' => 'urn:mace:dir:attribute-def:sn';
    'email'      => 'urn:mace:dir:attribute-def:mail';
]);

define('ATTRIBUTE_HUMAN_NAMES', [
    'institute'  => 'Instituut',
    'type'       => 'Type',
    'id'         => 'ID',
    'fullname'   => 'Volledige naam',
    'firstname'  => 'Voornaam',
    'familyname' => 'Achternaam',
    'email'      => 'Emailadres',
]);

require '../index.php';
