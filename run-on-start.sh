# run a cron job in the background
su -s "/bin/sh" -c "php -d max_execution_time=120 -d memory_limit=400M ${SSP_DIR}/modules/cron/bin/cron.php -t hourly" www-data