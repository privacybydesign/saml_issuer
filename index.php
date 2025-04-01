<?php

require_once 'config.php';

if (!defined('PROVIDER')) {
    // Trying to load this PHP script directly (old style), redirect to surfconext issuer.
    header('Location: surfconext/', true, 301);
    exit('redirect to surfconext/');
}

define('PAGE_LOGIN', '../login-' . LANG . '.php');
define('PAGE_ISSUE', '../issue-' . LANG . '.php');
define('PAGE_DONE',  '../done-' . LANG . '.php');

if (LANG == 'nl') {
    $ATTRIBUTE_HUMAN_NAMES = [
        'id'          => 'Gebruiker ID',
        'fullid'      => 'Volledig ID',
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

} else { // en, or fallback
    $ATTRIBUTE_HUMAN_NAMES = [
        'id'          => 'User ID',
        'fullid'      => 'Full ID',
        'username'    => 'Username',
        'profileurl'  => 'Profile',
        'fullname'    => 'Full name',
        'firstname'   => 'First name',
        'familyname'  => 'Family name',
        'email'       => 'Email address',
        'dateofbirth' => 'Birthdate',
        'institute'   => 'Institute',
        'type'        => 'Type',
    ];
}

function map_saml_attributes($saml_attributes) {
    global $MAP_IRMA_SAML_ATTRIBUTES;
    $irma_attributes = [];
    foreach ($MAP_IRMA_SAML_ATTRIBUTES as $irma_key => $saml_key) {
        writeLog("50: " . json_encode($saml_key, JSON_PRETTY_PRINT));

        $value = NULL;
        if (isset($saml_attributes[$saml_key]) &&
                count($saml_attributes[$saml_key]) > 0 &&
                $saml_attributes[$saml_key] !== NULL) {
            $value = $saml_attributes[$saml_key][0];
            writeLog("57: " . $irma_key . ", " . $value);
        }
        if ($irma_key === 'dateofbirth') {
            if (preg_match('#^[0-9]{2}/[0-9]{2}/[0-9]{4}$#', $value)) {
                $parts = explode('/', $value);
                $value = "{$parts[1]}-{$parts[0]}-{$parts[2]}";
            }
            if ($value === NULL)
                $value = ' '; // Fallback to old solution for absent attributes: dateofbirth is not (yet) optional :(
        }
        if ($value !== NULL) {
            $irma_attributes[$irma_key] = $value;
            writeLog("69: " . $irma_key . ", " . $value);
        } else {
            writeLog("71: " . $irma_key . ", " . $value . " = NULL");
        }
    }

    // Add profileurl from username
    if (array_key_exists('username', $irma_attributes) && !array_key_exists('profileurl', $irma_attributes) && defined('PROFILE_URL_FORMAT')) {
        $irma_attributes['profileurl'] = str_replace(':username:', $irma_attributes['username'], PROFILE_URL_FORMAT);
    }

    return $irma_attributes;
}

function start_session($sessionrequest) {
    $jsonsr = json_encode($sessionrequest);
    $api_call = array(
        'http' => array(
            'method' => 'POST',
            'header' => "Content-type: application/json\r\n"
                . "Content-Length: " . strlen($jsonsr) . "\r\n"
                . "Authorization: " . IRMA_SERVER_API_TOKEN . "\r\n",
            'content' => $jsonsr
        )
    );
    $resp = file_get_contents(IRMA_SERVER_URL . '/session', false, stream_context_create($api_call));
    if (! $resp) {
        die();
    }
    return $resp;
}

function start_issuance_session($irma_attributes, $validity) {
    $iprequest = [
        "@context" => "https://irma.app/ld/request/issuance/v2",
        "credentials" => [[
            "credential" => CREDENTIAL,
            "validity" => $validity,
            "attributes" => $irma_attributes
        ]]
    ];

    return start_session($iprequest);
}

function start_verification_session($attributes) {
    $sprequest = [
        "@context" => "https://irma.app/ld/request/disclosure/v2",
        "disclose" => $attributes
    ];

    return start_session($sprequest);
}

function handle_request() {
    global $ATTRIBUTE_HUMAN_NAMES, $MAP_IRMA_SAML_ATTRIBUTES;

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
        if (isset($_GET['output']) && $_GET['output'] == 'irma-session') {
            header('Content-Type: text/plain');
            $fullname_attribute = CREDENTIAL . '.' . IRMA_NAME_ATTRIBUTE;
            echo start_verification_session([[[$fullname_attribute]]]);
        } else {
            require PAGE_DONE;
        }
    } elseif ($saml_authenticator->isAuthenticated()) {
        $attrs = $saml_authenticator->getAttributes();

        writeLog(json_encode($attrs, JSON_PRETTY_PRINT));
        
        $irma_attributes = map_saml_attributes($attrs);
        $validity = (new DateTime(VALIDITY))->getTimestamp();

        if (isset($_GET['output']) && $_GET['output'] == 'irma-session') {
            header('Content-Type: text/plain');
            echo start_issuance_session($irma_attributes, $validity);
        } else {
            require PAGE_ISSUE;
        }

    } else {
        require PAGE_LOGIN;
    }
}

function writeLog($message) {
    if (ENABLE_IRMA_LOGGING) {
        $timestamp = date('Y-m-d H:i:s');
        $logMessage = "[$timestamp] $message" . PHP_EOL;
        file_put_contents(LOG_FILE, $logMessage, FILE_APPEND);
    }
}

handle_request();
