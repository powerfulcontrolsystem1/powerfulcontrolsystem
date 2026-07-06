#!/usr/bin/env bash
set -uo pipefail

echo "---SYSTEM---"
lsb_release -ds
uname -r

echo "---SERVICES---"
systemctl is-enabled ssh docker vncserver@1.service
systemctl is-active ssh docker vncserver@1.service
systemctl is-enabled sleep.target suspend.target hibernate.target hybrid-sleep.target

echo "---DOCKER---"
docker ps --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}'

echo "---DISK---"
df -hT / /srv/data

echo "---URLS---"
curl -I --max-time 10 http://127.0.0.1:8081/ | head -5
curl -fsS --max-time 10 http://127.0.0.1:8082/status.php
