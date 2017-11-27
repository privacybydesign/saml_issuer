<?php

define('ROOT_DIR', __DIR__ . '/../../../../');

require_once ROOT_DIR . 'vendor/autoload.php';
require_once ROOT_DIR . 'simplesamlphp/lib/_autoload.php';

use \Firebase\JWT\JWT;

class IrmaAutenticatedUser {
    private $saml_authenticator;

    function __construct() {
        $this->saml_authenticator = new SimpleSAML_Auth_Simple('facebook');
    }

    function isAuthenticated() {
        return $this->saml_authenticator->isAuthenticated();
    }

    function login() {
        $this->saml_authenticator->login();
    }

    function logout() {
        $this->saml_authenticator->logout();
    }

    function loadAttributes() {
        if (!$this->isAuthenticated())
            return;

        $attributes = $this->saml_authenticator->getAttributes();

        $this->fullname = $attributes["facebook.name"][0];
        $this->id = $attributes["facebook.id"][0];
    }
}

function get_jwt_key() {
    $pk = openssl_pkey_get_private("file://" . ROOT_DIR . "saml-sk.pem");
    if ($pk === false)
        throw new Exception("Failed to load signing key");
    return $pk;
}

function get_verification_jwt() {
    $pk = get_jwt_key();
    $sprequest = [
        "sub" => "verification_request",
        "iss" => "Privacy by Design Foundation",
        "iat" => time(),
        "sprequest" => [
            "validity" => 60,
            "request" => [
                "content" => [
                    [
                        "label" => "Facebook full name",
                        "attributes" => ["pbdf.pbdf.facebook.fullname"]
                    ],
                ]
            ]
        ]
    ];

    return JWT::encode($sprequest, $pk, "RS256", "facebook_enroll");
}

function get_issuance_jwt($authenticated_user) {
    $pk = get_jwt_key();
    $iprequest = [
        "sub" => "issue_request",
        "iss" => "Privacy by Design Foundation",
        "iat" => time(),
        "iprequest" => [
            "timeout" => 300,
            "request" => [
                "credentials" => [
                    [
                        "credential" => "pbdf.pbdf.facebook",
                        "validity" => (new DateTime("+3 months"))->getTimestamp(),
                        "attributes" => [
                            "id" => $authenticated_user->id,
                            "fullname" => $authenticated_user->fullname,
                        ]
                    ]
                ]
            ]
        ]
    ];

    return JWT::encode($iprequest, $pk, "RS256", "facebook_enroll");
}

function handle_action($loginpage, $authenticatedpage, $donepage) {
    $authenticated_user = new IrmaAutenticatedUser();

    if (isset($_REQUEST ['action']))
        $action = $_REQUEST['action'];
    else
        $action = NULL;

    $authenticated = $authenticated_user->isAuthenticated();

    if ($action === 'login' && !$authenticated)
        $authenticated_user->login();
    if ($action === 'logout' && $authenticated)
        $authenticated_user->logout();

    if ($action === 'done') {
        $jwt = get_verification_jwt();
        include $donepage;
    } elseif (!$authenticated) {
        include $loginpage;
    } else {
        $authenticated_user->loadAttributes();
        $jwt = get_issuance_jwt($authenticated_user);
        include $authenticatedpage;
    }
}

handle_action("../login.html", "../issue.php", "../done.php");
