<?php
require_once($_SERVER['DOCUMENT_ROOT'] . '/support/prelude.inc');

$authenticated_user = irma_page_requires_authentication("login.html");

include("issue.php");
