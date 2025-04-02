####################
# Build image
####################
FROM node:23-alpine AS builder
WORKDIR /app

ADD js/ ./js/
ADD css/ ./css/
ADD yarn.lock .
ADD package.json .
ADD build.sh .

RUN yarn install

RUN chmod +x build.sh
RUN ./build.sh

####################
# Runtime image
####################
FROM cirrusid/simplesamlphp:v2.3.7 AS runtime

ENV SSP_NEW_UI=true

# Overwrite default Apache config
ADD apache.conf /etc/apache2/sites-available/ssp.conf

# Add runtime startup script
ADD run-on-start.sh /opt/simplesaml/
RUN chmod +x /opt/simplesaml/run-on-start.sh

# Install SimpleSAMLphp modules
WORKDIR ${SSP_DIR}
RUN composer config prefer-stable true \
        && composer require --update-no-dev cirrusidentity/simplesamlphp-module-authoauth2

RUN mkdir -p /var/data/simplesamlphp
RUN mkdir -p /var/log/simplesamlphp

RUN chown -R www-data:www-data /var/data/simplesamlphp && chmod -R 750 /var/data/simplesamlphp
RUN chown -R www-data:www-data /var/log/simplesamlphp && chmod -R 750 /var/log/simplesamlphp

# Add build files
COPY --from=builder /app/css/                /var/www/css/
COPY --from=builder /app/js/                 /var/www/js/

# Add project files
COPY edugain/                 /var/www/edugain/
COPY linkedin/                /var/www/linkedin/
COPY surfconext/              /var/www/surfconext/
COPY *.html *.php .htaccess   /var/www/
