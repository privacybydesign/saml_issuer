<?php

if (!defined('PROVIDER')) {
    // Trying to load this PHP script directly (old style), redirect to surfnet
    // issuer.
    header('Location: surfnet/', true, 301);
    exit('redirect to surfnet/');
}

define('ROOT_DIR', __DIR__ . '/../../../');

require_once ROOT_DIR . 'php/vendor/autoload.php';
require_once ROOT_DIR . 'simplesamlphp/lib/_autoload.php';

use \Firebase\JWT\JWT;

define('PAGE_LOGIN', '../login.php');
define('PAGE_ISSUE', '../issue.php');
define('PAGE_DONE',  '../done.php');

$ATTRIBUTE_HUMAN_NAMES = [
    'id'          => 'ID',
    'username'    => 'Gebruikersnaam',
    'profileurl'  => 'Profiel',
    'fullname'    => 'Volledige naam',
    'firstname'   => 'Voornaam',
    'familyname'  => 'Achternaam',
    'email'       => 'E-mailadres',
    'dateofbirth' => 'Geboortedatum',
    'institute'   => 'Instituut',
    'type'        => 'Type',
];

function map_saml_attributes($saml_attributes) {
    global $MAP_IRMA_SAML_ATTRIBUTES;
    $irma_attributes = [];
    foreach ($MAP_IRMA_SAML_ATTRIBUTES as $irma_key => $saml_key) {
        $value = ' ';
        if (isset($saml_attributes[$saml_key]) &&
                count($saml_attributes[$saml_key]) > 0 &&
                $saml_attributes[$saml_key] !== NULL) {
            $value = $saml_attributes[$saml_key][0];
        }
        if ($irma_key === 'dateofbirth' && preg_match('#^[0-9]{2}/[0-9]{2}/[0-9]{4}$#', $value)) {
            $parts = explode('/', $value);
            $value = "{$parts[1]}-{$parts[0]}-{$parts[2]}";
        }
        $irma_attributes[$irma_key] = $value;
    }

    return $irma_attributes;
}

function get_jwt_key() {
    $pk = openssl_pkey_get_private('file://' . ROOT_DIR . 'saml-sk.pem');
    if ($pk === false)
        throw new Exception('get_jwt_key: failed to load signing key');
    return $pk;
}

function get_issuance_jwt($irma_attributes) {
    $iprequest = [
        'sub' => 'issue_request',
        'iss' => 'Privacy by Design Foundation',
        'iat' => time(),
        'iprequest' => [
            'timeout' => 300,
            'request' => [
                'credentials' => [
                    [
                        'credential' => CREDENTIAL,
                        'validity' => (new DateTime('+3 months'))->getTimestamp(),
                        'attributes' => $irma_attributes,
                    ]
                ]
            ]
        ]
    ];

    $pk = get_jwt_key();
    return JWT::encode($iprequest, $pk, 'RS256', SERVER_NAME);
}

function get_verification_jwt($label, $attributes) {
    $sprequest = [
        'sub' => 'verification_request',
        'iss' => 'Privacy by Design Foundation',
        'iat' => time(),
        'sprequest' => [
            'validity' => 60,
            'request' => [
                'content' => [
                    [
                        'label' => $label,
                        'attributes' => $attributes,
                    ],
                ]
            ]
        ]
    ];

    $pk = get_jwt_key();
    return JWT::encode($sprequest, $pk, 'RS256', SERVER_NAME);
}

function handle_request() {
    global $ATTRIBUTE_HUMAN_NAMES;

    $saml_authenticator = new \SimpleSAML\Auth\Simple(PROVIDER);

    if (isset($_REQUEST ['action'])) {
        $action = $_REQUEST['action'];
    } else {
        $action = NULL;
    }

    if ($action === 'login' && !$saml_authenticator->isAuthenticated()) {
        $saml_authenticator->login();
    } else if ($action === 'logout' && $saml_authenticator->isAuthenticated()) {
        $saml_authenticator->logout();
    } else if ($action === 'done') {
        $fullname_attribute = CREDENTIAL . '.' . IRMA_NAME_ATTRIBUTE;
        $jwt = get_verification_jwt(VERIFY_NAME_LABEL, [$fullname_attribute]);
        include PAGE_DONE;
    } elseif ($saml_authenticator->isAuthenticated()) {
        $irma_attributes = map_saml_attributes($saml_authenticator->getAttributes());
        $jwt = get_issuance_jwt($irma_attributes);
        include PAGE_ISSUE;

    } else {
        include PAGE_LOGIN;
    }
}

handle_request();
