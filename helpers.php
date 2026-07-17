<?php

/**
 * Pure output-escaping helpers.
 *
 * These are kept in their own file (with no SimpleSAMLphp dependency) so they
 * can be unit-tested in isolation — see tests/helpers_test.php.
 */

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
