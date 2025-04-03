# Higher memory limits are needed for EduGAIN metadata refresh.
# As we are not using EduGAIN (yet), we don't need to update the memory limits now.

# Set memory_limit in php.ini
#su -s "/bin/sh" -c "sed -i 's,^memory_limit =.*$,memory_limit = 256M,' /usr/local/etc/php/php.ini"

# run a cron job in the background
if [ -f "${SSP_DIR}/modules/cron/bin/cron.php" ]; then
    su -s "/bin/sh" -c "php -d max_execution_time=120 -d ${SSP_DIR}/modules/cron/bin/cron.php -t hourly" www-data &
    #su -s "/bin/sh" -c "php -d max_execution_time=120 -d memory_limit=256M ${SSP_DIR}/modules/cron/bin/cron.php -t hourly" www-data &
fi