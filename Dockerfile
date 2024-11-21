FROM node:18 AS builder

RUN apt-get update && apt-get install -y \
    php \
    php-cli \
    php-zip \
    php-xml \
    php-mbstring \
    php-curl \
    php-sqlite3 \
    unzip

RUN curl -sS https://getcomposer.org/installer | php -- --install-dir=/usr/local/bin --filename=composer

WORKDIR /app

COPY . .

RUN composer install
RUN yarn install

RUN chmod +x build.sh
RUN ./build.sh


RUN chmod +x install-simplesamlphp.sh
RUN ./install-simplesamlphp.sh

FROM php:8.0-apache

COPY --from=builder /app/ /var/www/html/

RUN chown -R www-data:www-data /var/www/html \
    && chmod -R 755 /var/www/html

RUN echo "Listen 8080" >> /etc/apache2/ports.conf

EXPOSE 8080
