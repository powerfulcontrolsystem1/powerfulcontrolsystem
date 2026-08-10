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

## Límite de sesiones y recepción externa (2026-08-01)

El buzón externo autorizado recibió una alerta real de sesiones administrativas
por encima del umbral. La conciliación agregada, sin exponer identidades ni
tokens, encontró 249 sesiones vigentes en staging y 160 en producción; una sola
identidad acumulaba 244 y 155 respectivamente. Las cargas autenticadas abrían
sesiones de 24 horas, pero `CreateSession` no limitaba cuántas podían coexistir
para la misma identidad.

El candidato `99348ff4e44283d461723afe87800e8d58bb2700` incorpora un límite de
20 sesiones por identidad. La inserción, limpieza de expiradas y poda de las más
antiguas ocurren en una transacción; PostgreSQL usa un advisory lock derivado de
la identidad para serializar el contrato entre réplicas. Una prueba PostgreSQL
efímera partió de 25 sesiones, agregó una nueva y terminó con 20 activas, la
nueva conservada y las seis más antiguas revocadas; la transacción se revirtió y
no dejó datos.

El workflow inmutable `30731146474` aprobó build, Trivy, SBOM, publicación por
digest y compose. Tras respaldar ambas bases se desplegaron exclusivamente en
staging:

- API: `sha256:18dedb17a62b6733d36eb58cf2d097b57e5097024c3fe687fa70aec727ac11f4`;
- migrador: `sha256:8ddeef5b9a3fda84baa35a770488b0b4443af1b9d6d80861c4e6609f62602c9b`;
- worker: `sha256:dfb62a2ec17dca780f6ef69c571c6c2a79f940963d84202b9510fec141cd39f0`;
- frontend: `sha256:ca0ef5008710338214dd5e01ed96e7b2889f978f64b7d64c749de905c4d8a5e3`.

Veinticinco inicios de sesión reales y secuenciales aprobaron. Staging terminó
con cinco identidades, 25 sesiones vigentes, máximo 20 por identidad y cero
identidades sobre el límite. Producción conservó 160 sesiones, máximo 155, y sus
imágenes locales: no recibió código ni limpieza. Los tres servicios de staging
quedaron saludables.

La recepción por Gmail demuestra que existe canal externo funcional. P109-010
permanece parcial hasta demostrar deduplicación/resolución completa, responsables
de escalamiento, SLO/error budget y simulacro firmado.
