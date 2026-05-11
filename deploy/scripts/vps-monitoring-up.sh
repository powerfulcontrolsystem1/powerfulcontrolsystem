#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/root/powerfulcontrolsystem}"
MONITORING_ENV="${MONITORING_ENV:-$PROJECT_DIR/deploy/monitoring/.env.monitoring}"

cd "$PROJECT_DIR"

docker network inspect pcs_internal >/dev/null 2>&1 || docker network create pcs_internal >/dev/null
docker network inspect pcs_staging_internal >/dev/null 2>&1 || docker network create pcs_staging_internal >/dev/null

if [ ! -f "$MONITORING_ENV" ]; then
  GRAFANA_PASSWORD="$(LC_ALL=C tr -dc 'A-Za-z0-9_#%@+=' </dev/urandom | head -c 32 || true)"
  if [ -z "$GRAFANA_PASSWORD" ]; then
    GRAFANA_PASSWORD="change-me-grafana-admin"
  fi
  cat > "$MONITORING_ENV" <<'EOF'
TZ=America/Bogota
PROMETHEUS_BIND=127.0.0.1
PROMETHEUS_PORT=9090
GRAFANA_BIND=127.0.0.1
GRAFANA_PORT=3001
GRAFANA_ADMIN_USER=admin
EOF
  printf 'GRAFANA_ADMIN_PASSWORD=%s\n' "$GRAFANA_PASSWORD" >> "$MONITORING_ENV"
  chmod 600 "$MONITORING_ENV" 2>/dev/null || true
  echo "[OK] Se creo $MONITORING_ENV con una clave Grafana local. Guarda esa clave antes de exponer Grafana."
elif grep -q 'GRAFANA_ADMIN_PASSWORD=change-me-grafana-admin' "$MONITORING_ENV"; then
  GRAFANA_PASSWORD="$(LC_ALL=C tr -dc 'A-Za-z0-9_#%@+=' </dev/urandom | head -c 32 || true)"
  if [ -n "$GRAFANA_PASSWORD" ]; then
    sed -i "s/GRAFANA_ADMIN_PASSWORD=change-me-grafana-admin/GRAFANA_ADMIN_PASSWORD=$GRAFANA_PASSWORD/" "$MONITORING_ENV"
    echo "[OK] Se reemplazo la clave placeholder de Grafana por una clave fuerte."
  else
    echo "[WARN] No se pudo generar clave fuerte; revisa GRAFANA_ADMIN_PASSWORD en $MONITORING_ENV."
  fi
fi

docker compose --env-file "$MONITORING_ENV" -f deploy/monitoring/docker-compose.monitoring.yml up -d
docker compose --env-file "$MONITORING_ENV" -f deploy/monitoring/docker-compose.monitoring.yml ps

echo "[OK] Monitoreo levantado. Prometheus y Grafana quedan ligados a 127.0.0.1 por defecto."
