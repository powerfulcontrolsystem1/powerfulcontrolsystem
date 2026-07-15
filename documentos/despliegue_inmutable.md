# Despliegue reproducible e inmutable

Actualizacion: 2026-07-15.

## Secuencia

1. Ejecutar pruebas, `go vet`, preflight y `git diff --check`.
2. Construir la imagen sin secretos y etiquetarla con la revision liberada.
3. Ejecutar el servicio `migrate` una sola vez; debe terminar bien antes de
   iniciar API o worker.
4. Iniciar API con `PCS_RUNTIME_SCHEMA_BOOTSTRAP=0` y comprobar `/health` y
   `/ready`.
5. Iniciar worker con sus dos DSN internos y comprobar la cola/outbox.
6. Cambiar trafico solo despues de evidencia de staging y backup restaurable.

## Configuracion obligatoria

Los secretos se inyectan fuera del repositorio. Son obligatorios en produccion:
`PCS_ENV`, origenes CSRF, proxies confiables, limites HTTP, sesion, claves de
cifrado y ambos DSN. Los pools PostgreSQL se configuran con
`PCS_DB_MAX_OPEN_CONNS`, `PCS_DB_MAX_IDLE_CONNS`,
`PCS_DB_CONN_MAX_LIFETIME`, `PCS_DB_CONN_MAX_IDLE_TIME`,
`PCS_DB_CONNECT_TIMEOUT`, `PCS_DB_QUERY_TIMEOUT` y `PCS_DB_TX_TIMEOUT`.
Generar claves con `openssl rand -base64 32`.

El worker recibe ademas `PCS_WORKER_ID`, `PCS_WORKER_POLL_INTERVAL`,
`PCS_WORKER_BATCH_SIZE` y `PCS_WORKER_JOB_LEASE`. El total de conexiones debe
calcularse como API replicas + worker replicas + migrador + margen operativo,
sin superar el limite de PostgreSQL.

## Seguridad del compose

Los servicios PCS se ejecutan read-only cuando es viable, sin capacidades Linux
adicionales y con `no-new-privileges`. PostgreSQL no publica puertos al exterior
del compose. Volumenes se limitan a datos, almacenamiento privado, backups y
logs necesarios. La imagen conserva temporalmente scripts de Mailu revisados
para provisionamiento corporativo; su separacion a una imagen administrativa
dedicada es un cambio de operacion que requiere prueba Mailu antes de retirar
ese artefacto.

## Rollback

No se revierten migraciones destructivamente. Ante incidente: detener trafico,
conservar evidencia saneada, restaurar imagen anterior compatible y usar un
backup probado si se requiere retorno de datos. La confirmacion se hace con
`/ready`, logs saneados y prueba operativa contra staging.
