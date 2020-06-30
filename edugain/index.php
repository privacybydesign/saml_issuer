<?php

require_once '../config.php';

define('PROVIDER',  'surfnet');
define('PROVIDER_NAME', 'eduGAIN');
define('SERVER_NAME', 'surfnet_enroll');
define('CREDENTIAL', IRMA_ISSUER_ID . '.surfnet-2');
define('IRMA_NAME_ATTRIBUTE', 'fullname');
define('VALIDITY', '+3 months');

$MAP_IRMA_SAML_ATTRIBUTES = [
    'institute'  => 'urn:oid:1.3.6.1.4.1.25178.1.2.9',
    'type'       => 'urn:oid:1.3.6.1.4.1.5923.1.1.1.1',
    'id'         => 'urn:oid:0.9.2342.19200300.100.1.1',
    'fullid'     => 'urn:oid:1.3.6.1.4.1.5923.1.1.1.6',
    'fullname'   => 'urn:oid:2.16.840.1.113730.3.1.241',
    'firstname'  => 'urn:oid:2.5.4.42',
    'familyname' => 'urn:oid:2.5.4.4',
    'email'      => 'urn:oid:0.9.2342.19200300.100.1.3',
];

require '../index.php';
