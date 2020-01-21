#!/usr/bin/env bash

set -euxo pipefail

wget -c https://github.com/simplesamlphp/simplesamlphp/releases/download/v1.17.7/simplesamlphp-1.17.7.tar.gz -O - | tar -xz
mv simplesamlphp-1.17.7 simplesamlphp
cd simplesamlphp
composer require --no-update cirrusidentity/simplesamlphp-module-authoauth2 && composer update --no-dev cirrusidentity/simplesamlphp-module-authoauth2
