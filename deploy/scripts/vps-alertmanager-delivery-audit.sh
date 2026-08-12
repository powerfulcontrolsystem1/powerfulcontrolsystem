#!/usr/bin/env bash
# Verifica si Alertmanager puede certificar entrega a un responsable externo.
# No imprime URLs, credenciales, receptores ni contenido de alertas.
set -euo pipefail

ALERTMANAGER_CONTAINER="${ALERTMANAGER_CONTAINER:-pcs-alertmanager}"
ALERTMANAGER_PORT="${ALERTMANAGER_PORT:-9093}"

fail() {
  echo "[ERROR] $*" >&2
  exit 1
}

docker inspect "$ALERTMANAGER_CONTAINER" >/dev/null 2>&1 || fail "Alertmanager no esta disponible."
curl -fsS --max-time 10 "http://127.0.0.1:${ALERTMANAGER_PORT}/-/ready" >/dev/null || fail "Alertmanager no responde por loopback."

config="$(docker exec "$ALERTMANAGER_CONTAINER" cat /etc/alertmanager/alertmanager.yml 2>/dev/null)" || fail "No se pudo leer la configuracion activa."

external_count="$(printf '%s\n' "$config" | grep -Ec '^[[:space:]]*(webhook_configs|email_configs|slack_configs|opsgenie_configs|pagerduty_configs):' || true)"
receiver_count="$(printf '%s\n' "$config" | grep -Ec '^[[:space:]]*-[[:space:]]+name:[[:space:]]*' || true)"
active_alerts="$(curl -fsS --max-time 10 "http://127.0.0.1:${ALERTMANAGER_PORT}/api/v2/alerts" | { grep -o '"status"' || true; } | wc -l | tr -d ' ')"

echo "[INFO] Alertmanager listo por loopback."
echo "[INFO] receptores_configurados=${receiver_count} alertas_activas=${active_alerts}"

if [[ "$external_count" -lt 1 ]]; then
  echo "[BLOCKED] No hay integracion externa de entrega configurada."
  echo "[BLOCKED] Las alertas internas no certifican P110-009 ni autorizan un GO."
  exit 2
fi

echo "[OK] Existe al menos una clase de entrega externa configurada."
echo "[INFO] Falta aun el simulacro autorizado de firing, recepcion, deduplicacion y resolucion."
