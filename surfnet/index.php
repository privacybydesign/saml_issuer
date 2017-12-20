<?php

define('PROVIDER',  'surfnet');
define('PROVIDER_NAME', 'Surfnet');
define('SERVER_NAME', 'surfnet_enroll');
define('CREDENTIAL', 'pbdf.pbdf.surfnet');
define('IRMA_NAME_ATTRIBUTE', 'fullname');

$MAP_IRMA_SAML_ATTRIBUTES = [
    'institute'  => 'urn:mace:terena.org:attribute-def:schacHomeOrganization',
    'type'       => 'urn:mace:dir:attribute-def:eduPersonAffiliation',
    'id'         => 'urn:mace:dir:attribute-def:uid',
    'fullname'   => 'urn:mace:dir:attribute-def:cn',
    'firstname'  => 'urn:mace:dir:attribute-def:givenName',
    'familyname' => 'urn:mace:dir:attribute-def:sn',
    'email'      => 'urn:mace:dir:attribute-def:mail',
];

require '../index.php';
