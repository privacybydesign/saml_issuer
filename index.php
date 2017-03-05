<?php

require_once __DIR__ . '/../../php/vendor/autoload.php';
require_once __DIR__ . '/../../simplesamlphp/lib/_autoload.php';

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

        $this->eduPersonPrincipalName  = $attributes['urn:mace:dir:attribute-def:eduPersonPrincipalName'][0];
        $this->fullName  = $attributes['urn:mace:dir:attribute-def:cn'][0];
        $this->givenName  = $attributes['urn:mace:dir:attribute-def:givenName'][0];
        $this->surname  = $attributes['urn:mace:dir:attribute-def:sn'][0];
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
        include $authenticatedpage;
    }
}


irma_page_requires_authentication("login.html", "issue.php");
