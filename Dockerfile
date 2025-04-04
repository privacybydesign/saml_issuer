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
FROM cirrusid/simplesamlphp:v2.3.7 AS runtime

ENV SSP_NEW_UI=true

# Install SimpleSAMLphp modules
WORKDIR ${SSP_DIR}
RUN composer config prefer-stable true \
        && composer require --update-no-dev cirrusidentity/simplesamlphp-module-authoauth2:4.1.0

RUN mkdir -p /var/data/simplesamlphp
RUN mkdir -p /var/log/simplesamlphp

RUN chown -R www-data:www-data /var/data/simplesamlphp && chmod -R 750 /var/data/simplesamlphp
RUN chown -R www-data:www-data /var/log/simplesamlphp && chmod -R 750 /var/log/simplesamlphp

# Add build files
COPY --from=builder /app/css/                /var/www/css/
COPY --from=builder /app/js/                 /var/www/js/

# Overwrite default Apache config
COPY apache.conf /etc/apache2/sites-available/ssp.conf

# Add runtime startup script
COPY run-on-start.sh /opt/simplesaml/
RUN chmod +x /opt/simplesaml/run-on-start.sh

# Add project files
COPY linkedin/      /var/www/linkedin/
COPY surfconext/    /var/www/surfconext/
COPY *.html *.php   /var/www/
