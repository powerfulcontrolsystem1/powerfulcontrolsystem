# Plan 103 - Cierre verificable de preproduccion

Fecha: 2026-07-16  
Estado: implementacion de controles de codigo cerrada; los gates de proveedores,
staging y recuperacion requieren evidencia operativa separada.

## Correccion del veredicto inicial

La revision de la rama vigente confirma que varios hallazgos del veredicto de
partida ya estaban resueltos antes de este plan:

- `pcs-migrate` ejecuta el catalogo versionado con checksum y
  `pg_advisory_xact_lock` transaccional.
- API y worker arrancan con `PCS_RUNTIME_SCHEMA_BOOTSTRAP=0` en Compose y el
  guard de runtime bloquea DDL por la capa de compatibilidad PostgreSQL.
- `pcs-worker` registra handlers de negocio, schedules persistentes, leases,
  recuperacion de jobs y dispatcher outbox. La venta cobrada publica
  `commerce.sale-paid` dentro de su transaccion.
- Docker usa `/ready` para API y worker; el proceso HTTP ya no se declara sano
  solo porque la portada responda.
- Licencias ya gobiernan modulos y permisos efectivos por empresa.

## Cierre incorporado en este plan

1. La recoleccion de metricas del sistema deja de vivir en `main.go` como
   goroutine de la API. `pcs-worker` la ejecuta como
   `maintenance.system-metrics`, con timeout, reintento y planificador durable.
2. El esquema `metrics` pasa a la migracion versionada
   `20260716-004-system-metrics-v1` de la base superadministradora. El worker
   solo lo verifica y falla con una instruccion clara de ejecutar
   `pcs-migrate` si falta.
3. El catalogo heredado conserva `InitMetricsTable` exclusivamente como puente
   para instalaciones nuevas durante el baseline del migrador; no se ejecuta
   desde la API ni el worker.

## Controles que siguen siendo obligatorios

No se debe declarar produccion general ni replicas solo por aprobar codigo.
Antes de ese paso se debe registrar evidencia de:

- staging con dos ejecuciones de `pcs-migrate`, verificacion de `/ready` y
  rollback compatible por digest;
- restauracion real de backup en entorno desechable con RPO/RTO medidos;
- proveedores sandbox autorizados: DIAN, Wompi/ePayco, correo, WhatsApp,
  Nextcloud y OnlyOffice;
- Object Storage compartido y URLs firmadas antes de replicas o adjuntos
  moviles;
- sesion movil por dispositivo, refresh rotativo, PKCE, push y sincronizacion
  incremental antes de publicar Android/iPhone.

## Criterio operativo actual

El codigo queda apto para continuar un piloto controlado en un host, siempre
que las migraciones pasen y los proveedores requeridos se hayan probado con
credenciales autorizadas. La salida multi-replica y la aplicacion movil publica
siguen condicionadas a los gates anteriores.
