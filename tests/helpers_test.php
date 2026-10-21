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

echo "escape_html():\n";
check('angle brackets', '&lt;script&gt;alert(1)&lt;/script&gt;', escape_html('<script>alert(1)</script>'));
check('double quote', '&quot;', escape_html('"'));
check('single quote', '&#039;', escape_html("'"));
check('ampersand', '&amp;', escape_html('&'));
check('null -> empty string', '', escape_html(null));
// Invalid UTF-8 must not blank the whole value (e.g. the language link).
check('invalid UTF-8 substituted', "a\u{FFFD}b", escape_html("a\xffb"));

// The templates read these constants and $_SERVER directly.
define('PROVIDER', 'surfnet');
define('PROVIDER_NAME', 'SURFconext');
define('IRMA_SERVER_URL', 'https://irma.example');
define('LANG', 'en');

function render_page(string $page, array $irma_attributes): string {
    $ATTRIBUTE_HUMAN_NAMES = ['firstname' => 'First name'];
    ob_start();
    require __DIR__ . '/../' . $page;
    return ob_get_clean();
}

// Hostile request values: the language links must print them escaped.
$script = '<script>alert(1)</script>';
$_SERVER['REQUEST_URI'] = '/nl/surfconext/"' . $script . '?q="' . $script;
$_SERVER['QUERY_STRING'] = 'q="' . $script;

foreach (['en', 'nl'] as $lang) {
    echo "\nlanguage links in $lang pages:\n";
    foreach (['login', 'issue', 'done'] as $name) {
        $html = render_page("$name-$lang.php", ['firstname' => 'Alice']);
        check("$name-$lang.php: request path and query escaped", false, str_contains($html, $script));
        check("$name-$lang.php: language link kept", true, str_contains($html, 'href="/en/surfconext/&quot;&lt;script&gt;'));
    }

    echo "\nfirstname in issue-$lang.php:\n";
    $html = render_page("issue-$lang.php", ['firstname' => '<img src=x onerror="alert(1)">']);
    check('markup in firstname escaped', false, str_contains($html, '<img src=x'));
    check('escaped firstname shown', true, str_contains($html, '&lt;img src=x onerror=&quot;alert(1)&quot;&gt;'));
}

echo "\n$count checks, $failures failure(s)\n";
exit($failures === 0 ? 0 : 1);
