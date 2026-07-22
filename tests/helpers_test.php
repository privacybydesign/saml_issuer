<?php

/**
 * Unit tests for the pure helpers in helpers.php.
 *
 * No PHPUnit dependency so it runs in a bare `php` CLI:
 *     php tests/helpers_test.php
 */

require_once __DIR__ . '/../helpers.php';

$failures = 0;
$count = 0;

function check(string $label, $expected, $actual): void {
    global $failures, $count;
    $count++;
    if ($expected === $actual) {
        echo "  PASS: $label\n";
    } else {
        $failures++;
        echo "  FAIL: $label\n";
        echo "        expected: " . var_export($expected, true) . "\n";
        echo "        actual:   " . var_export($actual, true) . "\n";
    }
}

echo "normalize_dateofbirth():\n";
// Absent attribute -> single-space fallback (dateofbirth is not optional).
check('null -> single space', ' ', normalize_dateofbirth(null));
// dd/mm/yyyy -> mm-dd-yyyy.
check('dd/mm/yyyy reformatted to mm-dd-yyyy', '11-25-1990', normalize_dateofbirth('25/11/1990'));
check('keeps leading zeros', '02-01-2000', normalize_dateofbirth('01/02/2000'));
// Non-matching values pass through unchanged (no null to preg_match).
check('already-normalised value unchanged', '11-25-1990', normalize_dateofbirth('11-25-1990'));
check('ISO date unchanged', '1990-11-25', normalize_dateofbirth('1990-11-25'));
check('empty string unchanged', '', normalize_dateofbirth(''));
check('non-date string unchanged', 'not-a-date', normalize_dateofbirth('not-a-date'));
// Too-short / malformed digit groups do not match the anchored pattern.
check('single-digit groups unchanged', '1/2/1990', normalize_dateofbirth('1/2/1990'));

echo "\n$count checks, $failures failure(s)\n";
exit($failures === 0 ? 0 : 1);
