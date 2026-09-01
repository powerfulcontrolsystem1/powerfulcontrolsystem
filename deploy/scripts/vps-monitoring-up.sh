#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/root/powerfulcontrolsystem}"
MONITORING_ENV="${MONITORING_ENV:-$PROJECT_DIR/deploy/monitoring/.env.monitoring}"
ALERTMANAGER_TEMPLATE="$PROJECT_DIR/deploy/monitoring/alertmanager.yml.tmpl"
ALERTMANAGER_RUNTIME="$PROJECT_DIR/deploy/monitoring/alertmanager.runtime.yml"

cd "$PROJECT_DIR"

docker network inspect pcs_internal >/dev/null 2>&1 || docker network create pcs_internal >/dev/null
docker network inspect pcs_staging_internal >/dev/null 2>&1 || docker network create pcs_staging_internal >/dev/null
docker network inspect pcs_mailu_internal >/dev/null 2>&1 || {
  echo "[ERROR] No existe la red privada pcs_mailu_internal de Mailu." >&2
  exit 1
}

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
ALERTMANAGER_BIND=127.0.0.1
ALERTMANAGER_PORT=9093
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

[ -s "$ALERTMANAGER_TEMPLATE" ] || {
  echo "[ERROR] No existe la plantilla de Alertmanager." >&2
  exit 1
}

alert_env_value() {
  local key="$1"
  sed -n "s/^${key}=//p" "$MONITORING_ENV" | tail -n 1 | tr -d '\r'
}

alert_email_enabled="$(alert_env_value ALERTMANAGER_EMAIL_ENABLED)"
alert_smtp_host="$(alert_env_value ALERTMANAGER_SMTP_SMARTHOST)"
alert_from="$(alert_env_value ALERTMANAGER_EMAIL_FROM)"
alert_to="$(alert_env_value ALERTMANAGER_EMAIL_TO)"
case "${alert_email_enabled:-0}" in
  0|false|FALSE|no|NO|'') alert_email_enabled=0 ;;
  1|true|TRUE|yes|YES) alert_email_enabled=1 ;;
  *) echo "[ERROR] ALERTMANAGER_EMAIL_ENABLED debe ser 0 o 1." >&2; exit 1 ;;
esac

if [ "$alert_email_enabled" = "1" ]; then
  [[ "$alert_smtp_host" =~ ^[A-Za-z0-9.-]+:[0-9]{1,5}$ ]] || {
    echo "[ERROR] ALERTMANAGER_SMTP_SMARTHOST invalido." >&2; exit 1;
  }
  [[ "$alert_from" =~ ^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+$ ]] || {
    echo "[ERROR] ALERTMANAGER_EMAIL_FROM invalido." >&2; exit 1;
  }
  [[ "$alert_to" =~ ^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+$ ]] || {
    echo "[ERROR] ALERTMANAGER_EMAIL_TO invalido." >&2; exit 1;
  }
fi

awk -v enabled="$alert_email_enabled" -v smtp="$alert_smtp_host" -v from="$alert_from" -v to="$alert_to" '
  /^__SMTP_GLOBAL__$/ {
    if (enabled == "1") {
      print "  smtp_smarthost: \"" smtp "\""
      print "  smtp_from: \"" from "\""
      print "  smtp_require_tls: false"
    }
    next
  }
  /^  receiver: __PRIMARY_RECEIVER__$/ {
    print "  receiver: " (enabled == "1" ? "observabilidad-correo" : "observabilidad-interna")
    next
  }
  /^__EMAIL_RECEIVER__$/ {
    if (enabled == "1") {
      print "  - name: observabilidad-correo"
      print "    email_configs:"
      print "      - to: \"" to "\""
      print "        send_resolved: true"
    }
    next
  }
  { print }
' "$ALERTMANAGER_TEMPLATE" > "$ALERTMANAGER_RUNTIME"
# Alertmanager runs without root privileges. The runtime file contains no
# credentials (the Mailu relay is on an internal Docker network), but it must
# be readable by the container user.
chmod 644 "$ALERTMANAGER_RUNTIME"

docker compose --env-file "$MONITORING_ENV" -f deploy/monitoring/docker-compose.monitoring.yml up -d
docker compose --env-file "$MONITORING_ENV" -f deploy/monitoring/docker-compose.monitoring.yml ps

for attempt in $(seq 1 12); do
  if curl -fsS --max-time 5 "http://127.0.0.1:${ALERTMANAGER_PORT:-9093}/-/ready" >/dev/null; then
    break
  fi
  if [ "$attempt" -eq 12 ]; then
    echo "[ERROR] Alertmanager no respondio en loopback." >&2
    exit 1
  fi
  sleep 2
done

echo "[OK] Monitoreo levantado. Prometheus, Grafana y Alertmanager quedan ligados a 127.0.0.1 por defecto."
