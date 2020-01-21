#!/usr/bin/env bash

wget -c https://github.com/simplesamlphp/simplesamlphp/releases/download/v1.18.3/simplesamlphp-1.18.3.tar.gz -O - | tar -xz
mv simplesamlphp-1.18.3 simplesamlphp
cd simplesamlphp
composer require --no-update cirrusidentity/simplesamlphp-module-authoauth2 && composer update --no-dev cirrusidentity/simplesamlphp-module-authoauth2
