<?php 

# require_once ($_SERVER['DOCUMENT_ROOT'] . '/../support/php/prelude.inc');
require_once ($_SERVER['DOCUMENT_ROOT'] . '/support/php/prelude.inc');

$authenticated_user = irma_page_requires_authentication_mockup ();


$template = new IrmaTemplate ('enroll.template.xhtml');

// fill in template parameters
$template->picture_path = "";
$template->authenticated_user = $authenticated_user;

// execute the template
try {
    echo $template->execute();
} catch (Exception $e) {
    echo $e;
}

?> 
