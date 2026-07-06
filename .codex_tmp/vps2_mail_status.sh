#!/usr/bin/env bash
set -u

echo "---MAIL LOG---"
tail -160 /tmp/pcs_mail_profile.log 2>/dev/null || true

echo "---RUNNING---"
pgrep -af 'docker compose.*profile mail' || true

echo "---DOCKER---"
docker ps --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}'

echo "---MAIL PORTS---"
ss -ltnp | grep -E ':2525|:8143|:8465|:8587|:8993|:8089|:8091' || true
