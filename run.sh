# run a cron job in the background
php -d max_execution_time=120 -d memory_limit=600M ./simplesamlphp/modules/cron/bin/cron.php -t hourly

# start the apache server
apache2-foreground
