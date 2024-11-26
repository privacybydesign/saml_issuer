#!/usr/bin/env bash

set -euxo pipefail

wget -c https://github.com/simplesamlphp/simplesamlphp/releases/download/v2.3.3/simplesamlphp-2.3.3-full.tar.gz -O - | tar -xz
mv simplesamlphp-2.3.3 simplesamlphp
cd simplesamlphp
composer require --no-update cirrusidentity/simplesamlphp-module-authoauth2 && composer update --no-dev cirrusidentity/simplesamlphp-module-authoauth2

cp ../template/config.php ./config/config.php
cp ../template/authsources.php ./config/authsources.php

rm -rf metadata
mkdir metadata

cp ../template/module_cron.php config
cp ../template/config-metarefresh.php config

cp ../template/saml20-idp-remote.php metadata


mkdir -p modules/authtwitter
mkdir -p modules/cron
mkdir -p modules/metafresh
mkdir -p modules/oauth
mkdir -p modules/authoauth2

touch modules/authtwitter/enable
touch modules/cron/enable
touch modules/metafresh/enable
touch modules/oauth/enable
touch modules/authoauth2/enable

