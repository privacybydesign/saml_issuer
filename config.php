<?php

// Only set language if it is not already set in the querystring
define('LANG', getenv('lang') ?: 'nl');

define('ROOT_DIR', __DIR__ . '');
define('LOG_FILE', '/var/log/simplesamlphp/yivi.log');
define('ENABLE_IRMA_LOGGING', strtolower(getenv('ENABLE_IRMA_LOGGING')) === "true");
define('IRMA_SERVER_API_TOKEN', getenv('API_TOKEN'));
define('IRMA_SERVER_URL', getenv('IRMA_SERVER_URL'));

if (!getenv('IRMA_ISSUER_ID')) {
    fwrite(STDOUT, "WARNING: environment variable IRMA_ISSUER_ID not set, using default value 'pbdf.pbdf'." . PHP_EOL);
}
define('IRMA_ISSUER_ID', getenv('IRMA_ISSUER_ID') ?: 'pbdf.pbdf');

/**
 * Escape a value for safe output in an HTML text or double-quoted attribute context.
 * Use this for every user- or IdP-controlled value rendered into a template.
 */
function h($value): string {
    return htmlspecialchars((string) ($value ?? ''), ENT_QUOTES | ENT_HTML5);
}

/**
 * Return the URL only when it uses a safe http/https scheme, otherwise an empty string.
 * Prevents javascript:/data: (and other) URIs from being rendered as clickable links.
 */
function safe_http_url($url): string {
    $url = (string) ($url ?? '');
    return preg_match('#^https?://#i', $url) === 1 ? $url : '';
}

require_once getenv('SSP_DIR') . '/vendor/autoload.php';
require_once getenv('SSP_DIR') . '/lib/_autoload.php';
