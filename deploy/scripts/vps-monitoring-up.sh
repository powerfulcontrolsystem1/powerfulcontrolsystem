#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/root/powerfulcontrolsystem}"
MONITORING_ENV="${MONITORING_ENV:-$PROJECT_DIR/deploy/monitoring/.env.monitoring}"

cd "$PROJECT_DIR"

if [ ! -f "$MONITORING_ENV" ]; then
  cat > "$MONITORING_ENV" <<'EOF'
TZ=America/Bogota
PROMETHEUS_BIND=127.0.0.1
PROMETHEUS_PORT=9090
GRAFANA_BIND=127.0.0.1
GRAFANA_PORT=3001
GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=change-me-grafana-admin
EOF
  chmod 600 "$MONITORING_ENV" 2>/dev/null || true
  echo "[WARN] Se creo $MONITORING_ENV. Cambia GRAFANA_ADMIN_PASSWORD antes de exponer Grafana."
fi

docker compose --env-file "$MONITORING_ENV" -f deploy/monitoring/docker-compose.monitoring.yml up -d
docker compose --env-file "$MONITORING_ENV" -f deploy/monitoring/docker-compose.monitoring.yml ps

echo "[OK] Monitoreo levantado. Prometheus y Grafana quedan ligados a 127.0.0.1 por defecto."
