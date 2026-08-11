# P109-010 - Observabilidad local de depuracion de soportes IA

Fecha: 2026-08-09
Entorno: candidato local, sin despliegue ni datos PCS
Rama: `codex/p109-batch-no-pr`

## Alcance implementado

- `/metrics` consulta en forma agregada las sagas de depuracion pendientes,
  pendientes por al menos 15 minutos y finalizadas.
- La consulta tiene timeout y cache existentes; ante tabla o base no disponible
  devuelve cero con `pcs_observability_query_success{source="support_purge"} 0`.
- No utiliza etiquetas de empresa, soporte, usuario, archivo o ruta privada.
- Prometheus alerta con `PCSSoporteIAPurgaVencida` cuando existe una saga vencida
  durante dos minutos.
- Grafana muestra pendientes y vencidas por job. El runbook dirige la
  recuperacion al endpoint empresarial oficial y prohibe SQL/borrado manual.

## Verificacion local

- `go test . -run TestPrometheus -count=1`: PASS.
- `go vet .`: PASS.
- El render de metricas cubre valores, estado de consulta y ausencia de campos
  empresariales sensibles.
- El JSON del dashboard se parseó y el contrato de alerta fue validado por una
  prueba Go enfocada; el preflight Full/Strict aprobó sus 22 compuertas.

## Limites

No se desplego Prometheus/Grafana, no se inyecto una saga vencida real ni se
comprobo entrega a receptor externo. Este bloque mejora P109-010, pero no cambia
su estado parcial, el 53,3 % de implementacion ni el veredicto NO-GO.
