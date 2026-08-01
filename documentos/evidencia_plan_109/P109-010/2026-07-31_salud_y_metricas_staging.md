# P109-010 - Salud y métricas de staging

Fecha: 2026-07-31
Entorno: staging y suite local de contratos.

## Resultado de staging

- `/health`: `200`.
- `/ready`: `200`.
- `/metrics` público: `404`, por diseño; Prometheus debe consultar el backend
  solo dentro de la red privada Docker.

## Contratos repetidos

El binario de prueba del backend aprobó los contratos de señales operativas,
etiquetas agregadas/acotadas, rechazo de mutación en métricas y no exposición
por el frontend público.

## Límite de cierre

P109-010 continúa parcial. No se han completado la inyección de fallas de
lease/outbox/disco, alertas con recepción y deduplicación, dashboard/SLO/error
budget, canal de escalamiento ni simulacro de incidente.

## Capacidad y colas del candidato `eb853788` (2026-08-01)

La cuenta PCS ejecutó 500 solicitudes GET autenticadas de solo lectura sobre
`/me` y configuración operativa de `empresa_id=12`, con concurrencia 10:

| Métrica | Resultado |
| --- | ---: |
| Solicitudes | 500/500 |
| Fallos / error rate | 0 / 0 % |
| p50 / p95 / p99 | 98 / 117 / 307 ms |
| Umbral p95 | 2.500 ms |

Después de la carga, backend, worker, PostgreSQL y frontend continuaron
saludables; PostgreSQL tenía 26 sesiones y cero esperas de lock. Los jobs
administrativos estaban completados, sin leases vencidos. Outbox conservó
cuatro eventos publicados, cero leases vencidos y un evento CxP histórico
`dead` de PCS que ya estaba inventariado y explícitamente fuera de la
recuperación P109-001. No se creó ni modificó información de negocio.

Esta evidencia aprueba la carga autenticada de lectura para el digest, pero
P109-010 continúa parcial por las señales/inyecciones y el simulacro pendientes;
tampoco sustituye la carga transaccional de cuatro cajas de P109-006.
