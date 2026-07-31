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
