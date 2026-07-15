# Preparacion final de produccion

Actualizacion: 2026-07-15.

## Cierre tecnico de plataforma

La API se limita a HTTP, conexiones, middlewares, rutas, probes y apagado
controlado. Sus temporizadores historicos se trasladaron a `pcs-worker`. La
cola PostgreSQL y el outbox incluyen lease, heartbeat, recuperacion, prioridad,
idempotencia hash, reintentos y estado `dead`. `/ready` exige conexion a ambas
bases y la migracion de plataforma del release. Consultar
`arquitectura_workers_outbox.md` y `despliegue_inmutable.md`.

## Procedimiento de actualizacion

1. Ejecutar preflight y pruebas Go desde el repositorio.
2. Construir imagenes reproducibles sin secretos.
3. Iniciar `pcs-migrate`; debe terminar correctamente y registrar/aplicar el
   esquema antes de API y worker.
4. Iniciar `pcs-backend` y validar `GET /health` y `GET /ready`.
5. Iniciar `pcs-worker`; revisar trabajos `processing`, `dead`, reintentos y
   outbox antes de aumentar replicas.
6. Verificar visualmente login, empresa, venta, caja, documento, adjunto y API
   movil contra staging anonimo.

## Variables criticas

`PCS_ENV=production`, `PCS_TRUSTED_PROXY_CIDRS`,
`PCS_CSRF_ALLOWED_ORIGINS`, `SESSION_TIMEOUT`, `MAX_REQUEST_BODY_BYTES`,
`HTTP_READ_TIMEOUT`, `HTTP_WRITE_TIMEOUT`, `HTTP_IDLE_TIMEOUT`,
`PCS_PRIVATE_STORAGE_DIR`, `CONFIG_ENC_KEY`, `CONFIG_ENC_KEY_ID`,
`DB_EMPRESAS_DSN` y `DB_SUPERADMIN_DSN` deben estar definidos fuera del
repositorio. `CONFIG_ENC_KEY_PREVIOUS` se usa solo durante rotacion. Generar
claves con `openssl rand -base64 32`; nunca registrar el valor resultante.

Los limites del pool se inyectan como `PCS_DB_MAX_OPEN_CONNS`,
`PCS_DB_MAX_IDLE_CONNS`, `PCS_DB_CONN_MAX_LIFETIME`,
`PCS_DB_CONN_MAX_IDLE_TIME`, `PCS_DB_CONNECT_TIMEOUT`,
`PCS_DB_QUERY_TIMEOUT` y `PCS_DB_TX_TIMEOUT`. El worker necesita los dos DSN
internos, `PCS_WORKER_ID`, `PCS_WORKER_POLL_INTERVAL`,
`PCS_WORKER_BATCH_SIZE` y `PCS_WORKER_JOB_LEASE`.

## Verificacion y rollback

No se aprueba produccion solo por pruebas locales. Se exige backup restaurable,
staging anonimo, carga concurrente, validacion de proveedores y monitoreo.
Ante fallo: detener nuevas replicas, conservar logs saneados, restaurar imagen
anterior, validar `ready` y decidir reversa de datos unicamente desde backup.
No borrar tablas ni ejecutar rollback destructivo automatico.
