<?php

/**
 * Pure, SimpleSAMLphp-free helpers.
 *
 * These are kept in their own file (with no SimpleSAMLphp dependency) so they
 * can be unit-tested in isolation — see tests/helpers_test.php.
 */

/**
 * Normalise a SAML `dateofbirth` value for the IRMA attribute map.
 *
 * - absent / null            -> ' ' (dateofbirth is not (yet) optional, so a
 *                               single space keeps the attribute present)
 * - matching `dd/mm/yyyy`     -> reformatted `mm-dd-yyyy`
 * - any other (non-null) value -> returned unchanged
 *
 * Guarding null here also keeps preg_match() from receiving null, which is a
 * deprecation on PHP 8.1+.
 */
function normalize_dateofbirth($value): string {
    if ($value === null) {
        return ' ';
    }
    if (preg_match('#^[0-9]{2}/[0-9]{2}/[0-9]{4}$#', $value)) {
        $parts = explode('/', $value);
        return "{$parts[1]}-{$parts[0]}-{$parts[2]}";
    }
    return $value;
}
