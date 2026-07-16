# Preparacion para produccion y aplicacion movil

Estado: base operativa implementada el 2026-07-14. Este documento describe la
fundacion inicial. El plan de ejecucion vigente, hallazgos confirmados, backlog
ordenado y gates de liberacion estan en `plan_final_para_produccion.md`; ambos
deben leerse antes de declarar produccion o replicas listas.

## Arquitectura acordada

PCS conserva un monolito modular Go/PostgreSQL. Los modulos se comunican por
contratos HTTP, servicios y persistencia ya existentes; no se introducen
microservicios ni dependencias nuevas sin una necesidad demostrable. La empresa
activa procede siempre de sesion y permiso validado, nunca de un identificador
que el cliente pueda imponer.

Los procesos se separan en tres roles:

- `pcs-backend`: API HTTP y frontend actual.
- `pcs-migrate`: proceso corto, repetible y sin HTTP que registra/aplica la
  fundacion versionada de esquema antes de arrancar la API.
- `pcs-worker`: proceso sin HTTP para trabajo duradero en PostgreSQL.

La API conserva por defecto el bootstrap legado mientras se extraen sus tareas
de provisionamiento sin riesgo de regresion. Cuando el ledger y los flujos de
una instalacion ya esten verificados, produccion puede desactivarlo de forma
explicita con `PCS_RUNTIME_SCHEMA_BOOTSTRAP=0`; el rol de migracion siempre
puede aplicar la compatibilidad. Esa excepcion debe quedar registrada en la
bitacora de despliegue.

## Migraciones, cola y outbox

`schema_migrations` conserva las versiones aplicadas. La migracion de fundacion
crea la cola `pcs_async_jobs`, con leasing `FOR UPDATE SKIP LOCKED`, reintentos,
backoff, limite de intentos y estado terminal `dead`; tambien crea
`pcs_outbox_events` para que una mutacion pueda registrar un evento dentro de la
misma transaccion antes de solicitar un proveedor externo.

No se deben escribir secretos, tokens, datos completos de pago o adjuntos en
payloads. Cada tarea empresarial incluye `empresa_id` validado y el consumidor
debe volver a comprobar permiso/estado antes de ejecutar una accion externa.
Los workers nuevos se registran por tipo de tarea y deben ser idempotentes.

## API movil

La superficie estable actual es `/api/v1/`; su contrato vive en
`api/openapi.mobile.v1.yaml`. Incluye identidad, catalogo, clientes, carritos,
items, ventas, pagos, sincronizacion offline, documentos fiscales y
notificaciones. Las mutaciones sensibles exigen `Idempotency-Key`, se guardan
por hash y empresa, y las consultas tienen limite, offset, filtros y seleccion
cerrada de campos.

Los endpoints web antiguos permanecen por compatibilidad. Cualquier endpoint
movil nuevo debe usar el sobre JSON, `request_id`, permiso empresarial, errores
genericos y el checklist multiempresa. No se publican endpoints administrativos
internos como contrato movil.

## Checklist de liberacion

1. Ejecutar `pcs-migrate` y confirmar el ledger de ambas bases.
2. Verificar `pcs-worker` saludable sin tareas en estado `processing` vencidas
   ni tareas `dead` sin revision.
3. Ejecutar preflight completo, pruebas Go, `go vet`, detector de carreras y
   validacion de Compose en un host con Docker.
4. Restaurar una copia en ambiente desechable y probar login, venta, pago,
   factura, adjunto, empresa aislada y webhook firmado.
5. Ejecutar carga controlada sobre API v1 y revisar pool PostgreSQL, latencia y
   errores antes de habilitar trafico movil.
6. Confirmar backups, secretos de produccion, dominios, TLS, SPF/DKIM/DMARC y
   monitoreo. No se declara produccion lista solo por pasar pruebas locales.

## Limites actuales y siguiente adopcion

La fundacion de cola/outbox ya es persistente, pero los workers legados de
correo, DIAN, reportes e integraciones deben migrarse gradualmente a eventos
outbox, uno por uno, con prueba de idempotencia del proveedor. La API v1 usa
paginacion por offset por compatibilidad; catalogos grandes deben adoptar cursor
estable antes de consumo masivo movil. El almacenamiento privado actual mantiene
validacion y segregacion por empresa; la adopcion de S3 compatible se hace por
adaptador y migracion de objetos, nunca exponiendo rutas de proveedor al
frontend.
