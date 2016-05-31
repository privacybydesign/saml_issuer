<?php 

# require_once ($_SERVER['DOCUMENT_ROOT'] . '/../support/php/prelude.inc');
require_once ($_SERVER['DOCUMENT_ROOT'] . '/support/php/prelude.inc');

$authenticated_user = irma_page_requires_authentication ();

$template = new IrmaTemplate ('index.template.xhtml');

// fill in template parameters
$template->authenticated_user = $authenticated_user;

// execute the template
try {
    echo $template->execute();
} catch (Exception $e) {
    echo $e;
}

?> 
