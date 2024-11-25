<?php

$config = array(

	'sets' => array(

		// Metadata based on https://wiki.surfnet.nl/display/surfconextdev/Metadata+of+your+Service
		// Other settings are specified in authsources.php
		'edugain' => array(
			'cron'		=> array('hourly'),
			'sources'	=> array(
				array(
					'conditionalGET' => TRUE,
					'src' => 'https://metadata.surfconext.nl/edugain-downstream-idp.xml',
					'certificates' => array(
						'sfs.pem',
					),
					'validateFingerprint' => 'F4:EF:7D:97:F8:D9:90:F5:B2:2F:C7:57:03:C2:DC:E6:15:CE:09:9B',
					'template' => array(
						'tags'	=> array('edugain'),
						'authproc' => array(
						),
					),
				),
			),
			'expireAfter' 		=> 60*60*24*4, // Maximum 4 days cache time
			'outputDir' 	=> 'metadata/edugain/',
			'outputFormat'  => 'flatfile',
		),
	),
);
