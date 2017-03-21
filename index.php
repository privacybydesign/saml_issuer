<?php

define('ROOT_DIR', __DIR__ . '/../../../');

require_once ROOT_DIR . 'php/vendor/autoload.php';
require_once ROOT_DIR . 'simplesamlphp/lib/_autoload.php';

use \Firebase\JWT\JWT;

class IrmaAutenticatedUser {
    private $saml_authenticator;

    public $eduPersonPrincipalName = NULL;
    public $fullName = NULL;
    public $givenName = NULL;
    public $surname = NULL;

    function __construct() {
        $this->saml_authenticator = new SimpleSAML_Auth_Simple('surfnet');
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

        $this->institute = $attributes["urn:mace:terena.org:attribute-def:schacHomeOrganization"][0];
        $this->type = $attributes["urn:mace:dir:attribute-def:eduPersonAffiliation"][0];
        $this->id = $attributes["urn:mace:dir:attribute-def:uid"][0];
        $this->fullname = $attributes["urn:mace:dir:attribute-def:cn"][0];
        $this->firstname = $attributes["urn:mace:dir:attribute-def:givenName"][0];
        $this->familyname = $attributes["urn:mace:dir:attribute-def:sn"][0];
        $this->email = $attributes["urn:mace:dir:attribute-def:mail"][0];
    }
}

function irma_page_requires_authentication($loginpage, $authenticatedpage) {
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

    if (!$authenticated) {
        include $loginpage;
    } else {
        $authenticated_user->loadAttributes();
        $jwt = get_jwt($authenticated_user);
        include $authenticatedpage;
    }
}

function get_jwt($authenticated_user) {
    $pk = openssl_pkey_get_private("file://" . ROOT_DIR . "sk.pem");
    if ($pk === false) {
        throw new Exception("Failed to load signing key");
    }

    $iprequest = [
        "sub" => "issue_request",
        "iss" => "Privacy by Design Foundation",
        "iat" => time(),
        "iprequest" => [
            "timeout" => 300,
            "request" => [
                "credentials" => [
                    [
                        "credential" => "pbdf.pbdf.surfnet",
                        "attributes" => [
                            "institute" => $authenticated_user->institute,
                            "type" => $authenticated_user->type,
                            "id" => $authenticated_user->id,
                            "fullname" => $authenticated_user->fullname,
                            "firstname" => $authenticated_user->firstname,
                            "familyname" => $authenticated_user->familyname,
                            "email" => $authenticated_user->email
                        ]
                    ]
                ]
            ]
        ]
    ];

    return JWT::encode($iprequest, $pk, "RS256", "surfnet_enroll");
}

irma_page_requires_authentication("login.html", "issue.php");
