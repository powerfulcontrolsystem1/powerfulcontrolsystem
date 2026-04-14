#!/bin/bash
set -euo pipefail
CFG="/etc/nginx/sites-available/powerfulcontrolsystem"
if [ ! -f "$CFG" ]; then
  echo "ERROR: no existe $CFG" >&2
  exit 1
fi

if grep -q "~^([a-z0-9-]+)\\.powerfulcontrolsystem\\.com$" "$CFG"; then
  echo "INFO: wildcard ya configurado"
else
  sed -i "s/server_name powerfulcontrolsystem.com www.powerfulcontrolsystem.com;/server_name powerfulcontrolsystem.com www.powerfulcontrolsystem.com ~^([a-z0-9-]+)\\.powerfulcontrolsystem\\.com$;/" "$CFG"
  echo "INFO: wildcard agregado en server_name"
fi

nginx -t
systemctl reload nginx

echo "--- server_name line ---"
grep -n "server_name" "$CFG"

echo "--- host header test ---"
curl -I --max-time 15 -H "Host: empresa1.powerfulcontrolsystem.com" http://127.0.0.1 | head -n 8
