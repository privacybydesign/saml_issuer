<?php
/*
 * Configuration for the Cron module.
 */

$config = array (

	'key' => '{{ gateway_simplesamlphp_cronkey }}',
	'allowed_tags' => array('daily', 'hourly', 'frequent'),
	'debug_message' => TRUE,
	'sendemail' => TRUE,

);
