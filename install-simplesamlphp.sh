#!/usr/bin/env bash

set -euxo pipefail

wget -c https://github.com/simplesamlphp/simplesamlphp/releases/download/v2.0.14/simplesamlphp-2.0.14.tar.gz -O - | tar -xz
mv simplesamlphp-2.0.14 simplesamlphp
cd simplesamlphp
composer require --no-update cirrusidentity/simplesamlphp-module-authoauth2 && composer update --no-dev cirrusidentity/simplesamlphp-module-authoauth2

cp ../config.php.dist ./config/config.php
cp ../authsources.php.dist ./config/authsources.php
