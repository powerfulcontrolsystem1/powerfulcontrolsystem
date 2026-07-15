# Arquitectura de workers y outbox

Actualizacion: 2026-07-15.

## Roles de proceso

- `migrate`: es el unico rol que ejecuta bootstrap o DDL. En produccion la
  imagen inicia este rol antes que API y worker.
- `api`: atiende HTTP y no inicia temporizadores ni altera el esquema con
  `PCS_RUNTIME_SCHEMA_BOOTSTRAP=0`.
- `worker`: ejecuta tareas periodicas y consume las tablas PostgreSQL de jobs y
  outbox. Puede ejecutarse en varias replicas sin que las replicas API dupliquen
  los temporizadores historicos.

## Cola durable

`pcs_async_jobs` conserva empresa, usuario originador, tipo, version del
payload, prioridad, estado, intento, maximo de intentos, fecha disponible,
inicio, finalizacion, worker, heartbeat, resultado saneado, correlacion y hash
de idempotencia. El reclamo usa `FOR UPDATE SKIP LOCKED`; un lease vencido se
recupera y un trabajo que agotó sus intentos queda en `dead` para revision.

Los handlers se declaran mediante `worker.HandlerSpec`: tipo, version,
habilitacion, timeout, reintentos/backoff y funcion. Un handler debe recibir
solo IDs y metadatos seguros, volver a validar `empresa_id` y permisos de
dominio, y ser idempotente. Nunca debe llevar tokens, contrasenas, XML completo
ni documentos en el payload o resultado.

## Outbox transaccional

`pcs_outbox_events` se inserta dentro de la misma transaccion PostgreSQL de la
operacion local. El dispatcher reclama eventos concurrentemente, mantiene
heartbeat, aplica backoff, publica una sola vez por clave de idempotencia y
deja fallos agotados en `dead`. Un evento se convierte en trabajo externo solo
desde un handler registrado; los tipos no registrados no se descartan.

## Tareas periodicas actuales

El proceso `pcs-worker` aloja colector de metricas, alertas super, retencion de
auditoria, estado y alertas de licencias, snapshots VPS, agentes de
mantenimiento, parametros legales, cobranza, asientos contables y control
electrico. La API ya no las inicia.

## Operacion

1. Ejecutar primero `migrate` y comprobar su salida.
2. Validar `/health` y `/ready`; `ready` exige las conexiones y la migracion
`platform/20260715-worker-outbox` sin exponer detalles. El contenedor publica
un heartbeat privado en tmpfs; su healthcheck comprueba que el proceso sigue
actualizando la señal sin abrir otro puerto de red.
3. Escalar `worker` solo tras revisar pendientes, `processing` y `dead`.
4. Reprocesar un trabajo muerto mediante una accion administrativa auditada;
   no editar estados directamente en PostgreSQL.

## Transicion heredada

Los `Ensure...Schema` historicos siguen concentrados en el rol de migracion.
Una barrera central los convierte en no-op durante API/worker productivos, de
modo que una solicitud no puede adquirir locks DDL por una ruta heredada. La
extraccion de cada dominio a migraciones versionadas permanece incremental:
cada dominio nuevo debe registrar su migracion, prueba de compatibilidad y
requisito en readiness antes de eliminar una validacion heredada.
