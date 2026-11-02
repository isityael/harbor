#!/bin/bash
set -e
chown -R 10000:10000 /var/log/docker
cron
rm -f /var/run/rsyslogd.pid
sudo -u \#10000 -E  'rsyslogd' '-n'
set +e
