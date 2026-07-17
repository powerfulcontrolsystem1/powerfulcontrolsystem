# Plan 102 - Cierre tecnico de preproduccion

Fecha de corte: 2026-07-16  
Estado: implementacion tecnica cerrada para despliegue controlado; no equivale
a autorizacion de produccion general.

## Entregado

| Area | Resultado verificable |
| --- | --- |
| Migraciones | `pcs-migrate` aplica el ledger con lock/checksum y registra el baseline heredado; API y worker verifican, no crean esquema. |
| Runtime | Un guard bloquea DDL en API/worker de produccion y la configuracion rechaza bootstrap de runtime para esos roles. |
| Worker | Registro tipado de handlers reales, schedules persistentes, dispatcher de outbox y probes internos. |
| Consistencia | El cobro del carrito publica `commerce.sale-paid` en la misma transaccion; cierre de caja usa bloqueo transaccional. |
| Multiempresa | `TenantContext` tipado conserva empresa y rol ya validados y evita volver a confiar IDs recibidos del cliente. |
| Movil | Paginacion con cursor opaco en v1, idempotencia con expiracion/reclamo seguro y OpenAPI actualizado. |
| Readiness | `/ready` exige bases, migraciones y almacenamiento privado utilizable. Con replicas, storage local queda rechazado. |
| Release | Imagenes API/worker/migrador separadas y override de release que exige tres referencias inmutables por digest. |

## Gates que no se pueden simular desde codigo

1. Ejecutar staging con imagenes publicadas por digest, dos corridas de
   migracion y rollback compatible.
2. Restaurar un backup en un entorno desechable y medir RPO/RTO.
3. Configurar y validar Object Storage antes de replicas o adjuntos moviles.
4. Probar proveedores autorizados: DIAN, Wompi/ePayco, correo, WhatsApp,
   Nextcloud y OnlyOffice, incluyendo reintentos e idempotencia.
5. Completar para movil sesiones por dispositivo/refresh, PKCE, sincronizacion
   incremental/conflictos, push FCM/APNs y URLs firmadas de archivos.

## Regla de despliegue

Para una liberacion reproducible se usa `deploy/docker-compose.release.yml` y
se definen `PCS_API_IMAGE_DIGEST`, `PCS_WORKER_IMAGE_DIGEST` y
`PCS_MIGRATE_IMAGE_DIGEST` con formato `repositorio@sha256:<digest>`. No se
usan tags flotantes en un release aprobado.

## Evidencia requerida al cerrar un gate

- Pruebas Go, `go vet`, detector de carreras y preflight completos.
- Manifiesto de versiones/digests y salida de `pcs-migrate` sin deriva.
- Smoke multiempresa, prueba de recuperacion y validacion de `/ready`.
- Bitacora sin secretos ni payloads privados.
