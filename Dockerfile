####################
# Build image
####################
FROM node:22-alpine AS builder
WORKDIR /app

COPY js/ ./js/
COPY css/ ./css/
COPY yarn.lock .
COPY package.json .
COPY build.sh .

RUN yarn install

RUN chmod +x build.sh
RUN ./build.sh

####################
# Runtime image
####################
# Adapted from https://github.com/cirrusidentity/docker-simplesamlphp
FROM php:8.4-apache AS runtime

COPY --from=mlocati/php-extension-installer /usr/bin/install-php-extensions /usr/bin/

RUN apt-get update -y && apt-get -y upgrade && apt-get install -y \
        git \
        curl \
        unzip \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

RUN install-php-extensions gmp ldap soap apcu memcached zip pdo_pgsql pdo_mysql intl bcmath @composer-2.9.4

ARG SSP_VERSION="2.5.3.1"
ARG SSP_HASH="b199ebcba6e30220e7b5df2783ca42f8947dd9bdf046e9e6302fef3d3a438d01"

ENV COMPOSER_HOME=/opt/simplesamlphp/.composer \
    SSP_APACHE_ALIAS="simplesaml/" \
    SSP_DIR=/var/simplesamlphp \
    SSP_ENABLED_MODULES="" \
    SSP_LOG_HANDLER=errorlog \
    SSP_LOG_LEVEL=6 \
    SSP_NEW_UI=true

RUN cp $PHP_INI_DIR/php.ini-production $PHP_INI_DIR/php.ini

RUN a2enmod ssl rewrite headers \
    && echo "Mutex posixsem" >> /etc/apache2/apache2.conf
EXPOSE 443

RUN mkdir -p $SSP_DIR /var/cache/simplesamlphp $COMPOSER_HOME \
    && chown www-data $SSP_DIR /var/cache/simplesamlphp $COMPOSER_HOME

USER www-data
RUN curl -fsSL -o /tmp/ssp.tar.gz "https://github.com/simplesamlphp/simplesamlphp/releases/download/v${SSP_VERSION}/simplesamlphp-${SSP_VERSION}-full.tar.gz" \
    && echo "${SSP_HASH}  /tmp/ssp.tar.gz" | sha256sum -c - \
    && tar xzf /tmp/ssp.tar.gz --strip-components 1 -C $SSP_DIR \
    && rm /tmp/ssp.tar.gz

COPY --chown=www-data ssp-config/ ${SSP_DIR}/config/
RUN cp ${SSP_DIR}/config/config.php.dist ${SSP_DIR}/config/config.php \
    && echo 'require "config-override-base.php";' >> ${SSP_DIR}/config/config.php \
    && mkdir -p ${SSP_DIR}/data ${SSP_DIR}/ticketcache
USER root

# Install SimpleSAMLphp modules
WORKDIR ${SSP_DIR}
RUN composer config prefer-stable true \
        && composer require --update-no-dev --with-all-dependencies cirrusidentity/simplesamlphp-module-authoauth2:^5.0

RUN mkdir -p /var/data/simplesamlphp
RUN mkdir -p /var/log/simplesamlphp

RUN chown -R www-data:www-data /var/data/simplesamlphp && chmod -R 750 /var/data/simplesamlphp
RUN chown -R www-data:www-data /var/log/simplesamlphp && chmod -R 750 /var/log/simplesamlphp

# Add build files
COPY --from=builder /app/css/                /var/www/css/
COPY --from=builder /app/js/                 /var/www/js/
COPY fonts/                                  /var/www/fonts/
COPY images/                                 /var/www/images/

# Overwrite default Apache config
COPY apache.conf /etc/apache2/sites-available/ssp.conf
RUN a2ensite ssp

# Add runtime startup scripts
COPY run-on-start.sh ssp-startup.sh /opt/simplesaml/
RUN chmod +x /opt/simplesaml/run-on-start.sh /opt/simplesaml/ssp-startup.sh
CMD ["/opt/simplesaml/ssp-startup.sh"]

# Add project files
COPY linkedin/      /var/www/linkedin/
COPY surfconext/    /var/www/surfconext/
COPY *.html *.php   /var/www/
