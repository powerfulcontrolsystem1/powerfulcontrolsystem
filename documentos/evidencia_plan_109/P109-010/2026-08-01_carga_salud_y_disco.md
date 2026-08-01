# P109-010 - Carga, colas, salud y capacidad de disco

Fecha: 2026-08-01
Entorno: staging y VPS
Candidato: `6da6c13453a40b2d84e23285fa83f255f34788da`

La cuenta PCS ejecutó 500 GET autenticados de solo lectura sobre `/me` y la
configuración operativa de `empresa_id=12`, con concurrencia 10:

| Métrica | Resultado |
| --- | ---: |
| Solicitudes | 500/500 |
| Fallos / error rate | 0 / 0 % |
| p50 / p95 / p99 | 99 / 119 / 278 ms |
| Umbral p95 | 2.500 ms |

Después de la carga, los cuatro servicios de staging quedaron saludables,
PostgreSQL registró 30 sesiones, cero esperas de lock y cero leases vencidos.
Los jobs administrativos estaban completados y sin leases vencidos. El outbox
conservó cuatro eventos publicados y el único evento histórico `dead` excluido
de P109-001.

El disco estaba en 87 % por 19 volúmenes Docker anónimos sin contenedor, unos
8 GB. Se verificó individualmente que todos eran `dangling`, con nombre hash y
sin montajes; se eliminaron solo esos volúmenes. No se tocaron volúmenes con
nombre, bases, backups, archivos, imágenes actuales ni digests de rollback.
El uso bajó a 79 % y producción/staging mantuvieron salud y readiness.

Estado: **P109-010 parcial**. Faltan inyecciones de señales, alertas recibidas y
deduplicadas, SLO/error budget y simulacro con responsables.

## Observabilidad correlacionada del host

Prometheus tiene sanos sus targets de staging, node exporter y cAdvisor. Las
reglas PostgreSQL y lease durable estan cargadas, saludables e inactivas. Se
encontro una alerta critica `PCSBackendCaido` activa para el backend productivo:
el contenedor responde, pero su version anterior devuelve 401 al scrape privado
de `/metrics`; staging con el candidato responde correctamente. El receptor
`observabilidad-interna` no tiene un canal externo configurado.

No se silenció ni ocultó la alerta. Debe resolverse al promover el mismo
candidato que expone métricas solo en la red interna y, antes del GO, configurar
y ensayar un receptor autorizado con deduplicacion y resolución observada.
