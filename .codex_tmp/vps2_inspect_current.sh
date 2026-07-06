#!/usr/bin/env bash
set -u

echo "---DOCKER---"
docker ps --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}'

echo "---SERVICES---"
systemctl is-enabled ssh docker vncserver@1.service
systemctl is-active ssh docker vncserver@1.service

echo "---PORTS---"
ss -ltnp | grep -E ':22|:5901|:8081|:8082|:25|:80|:443|:587|:993' || true

echo "---NEXTCLOUD---"
curl -fsS --max-time 10 http://127.0.0.1:8082/status.php || true

echo "---PCS---"
curl -I --max-time 10 http://127.0.0.1:8081/ | head -5 || true
