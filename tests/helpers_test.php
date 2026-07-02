<?php

/**
 * Unit tests for the pure output-escaping helpers in helpers.php.
 *
 * These cover the security fix for the reflected XSS / attribute-injection
 * findings: h() must neutralise HTML/attribute metacharacters, and
 * safe_http_url() must reject non-http(s) schemes (javascript:, data:, ...).
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

echo "h() — rejection paths (must escape dangerous metacharacters):\n";
// Reflected XSS payload must not survive as live markup.
check('script tag is escaped',
    '&lt;script&gt;alert(1)&lt;/script&gt;',
    h('<script>alert(1)</script>'));
// Attribute-injection payload: double quote must be encoded.
check('double quote is escaped',
    '&quot; onmouseover=&quot;alert(1)',
    h('" onmouseover="alert(1)'));
// Single quote must be encoded too (ENT_QUOTES); ENT_HTML5 renders it &apos;.
check('single quote is escaped',
    '&apos; onmouseover=&apos;x',
    h("' onmouseover='x"));
check('ampersand is escaped', 'a &amp; b', h('a & b'));

echo "h() — happy paths (safe input passes through unchanged):\n";
check('plain text unchanged', 'Alice Bakker', h('Alice Bakker'));
check('null becomes empty string', '', h(null));
check('empty string stays empty', '', h(''));

echo "safe_http_url() — rejection paths (dangerous schemes dropped):\n";
check('javascript: URI rejected', '', safe_http_url('javascript:alert(1)'));
check('data: URI rejected', '', safe_http_url('data:text/html,<script>alert(1)</script>'));
check('scheme-relative URL rejected', '', safe_http_url('//evil.example/path'));
check('relative path rejected', '', safe_http_url('/local/path'));
check('null rejected', '', safe_http_url(null));
check('empty string rejected', '', safe_http_url(''));

echo "safe_http_url() — happy paths (valid http(s) URLs preserved):\n";
check('http URL preserved',
    'http://example.com/profile',
    safe_http_url('http://example.com/profile'));
check('https URL preserved',
    'https://example.com/profile?x=1',
    safe_http_url('https://example.com/profile?x=1'));
check('uppercase scheme accepted (case-insensitive)',
    'HTTPS://example.com',
    safe_http_url('HTTPS://example.com'));

echo "\n";
if ($failures === 0) {
    echo "OK: all $count assertions passed\n";
    exit(0);
}
echo "FAILED: $failures of $count assertions failed\n";
exit(1);
