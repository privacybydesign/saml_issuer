<?php

/**
 * Pure, SimpleSAMLphp-free helpers.
 *
 * These are kept in their own file (with no SimpleSAMLphp dependency) so they
 * can be unit-tested in isolation — see tests/helpers_test.php.
 */

/**
 * Escape a value for HTML text and quoted attribute contexts.
 */
function escape_html($value): string {
    return htmlspecialchars((string) $value, ENT_QUOTES | ENT_SUBSTITUTE);
}
