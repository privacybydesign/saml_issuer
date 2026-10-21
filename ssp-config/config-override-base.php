<?php

// Overrides for the default config.php, driven by environment variables.
$config['baseurlpath'] = getenv('SSP_BASE_URL_PATH') ?: getenv('SSP_APACHE_ALIAS');
$config['timezone'] = 'UTC';
$config['auth.adminpassword'] = getenv('SSP_ADMIN_PASSWORD');
$config['logging.level'] = (int) getenv('SSP_LOG_LEVEL');
$config['logging.handler'] = getenv('SSP_LOG_HANDLER');
$config['session.cookie.secure'] = true;
$config['usenewui'] = getenv('SSP_NEW_UI') === 'true';
$config['secretsalt'] = getenv('SSP_SECRET_SALT') ?: 'defaultsecretsalt';
$config['enable.saml20-idp'] = getenv('SSP_ENABLE_IDP') === 'true';

foreach (array_filter(explode(' ', (string) getenv('SSP_ENABLED_MODULES'))) as $moduleName) {
    $config['module.enable'][$moduleName] = true;
}

// Optional extra overrides, e.g. mounted into the container.
$overrideFile = __DIR__ . '/config-override.php';
if (file_exists($overrideFile)) {
    include $overrideFile;
}
