# Capacidad de colas multiempresa

Estado: Candidato local. Responsable: Coordinacion tecnica y QA/operacion. Revision: 2026-09-06.

## Objetivo

Evitar que el volumen de una empresa degrade a las demas y hacer visible, antes
del colapso, la presion de los trabajos de impresion, las altas de producto y la
emision fiscal. Esta arquitectura no certifica por si sola 1000 empresas: esa
cifra requiere una prueba de carga en el VPS candidato, con PostgreSQL,
replicas, red y proveedores equivalentes al entorno objetivo.

## Carriles independientes

| Carril | Ejecucion | Aislamiento y control |
| --- | --- | --- |
| Impresiones | Cola PostgreSQL durable consumida por agente/estacion | Claim con `FOR UPDATE SKIP LOCKED`, propiedad por `empresa_id`, limite de trabajos activos por empresa y bloqueo asesor transaccional para impedir carreras al admitir. |
| Agregar productos | Escritura sincronica transaccional; no se difiere porque el operador necesita confirmacion inmediata | Presupuesto por minuto y empresa para crear producto o agregarlo al carrito. Una empresa limitada recibe `429` y no consume el presupuesto de otra. |
| Emision fiscal | Cola durable de reintentos DIAN/proveedor | Particiones deterministas por `empresa_id`, rotacion por ultima empresa atendida, lease documental y procesamiento que continua si otra empresa falla. |

El carril de productos se denomina cola operativa en el panel, pero no es una
cola diferida. Convertirlo en trabajo asincrono haria ambiguos el inventario,
la respuesta al cajero y los errores de validacion. La proteccion correcta es
admision por empresa, transaccion corta e indicadores agregados.

## Configuracion y observabilidad

`Super administrador > Plataforma > Capacidad de colas` consume la API
protegida `/super/api/capacidad_colas`. Permite configurar, para cada carril:

- solicitudes por minuto y empresa;
- umbral global de pendientes;
- antiguedad maxima antes de alertar;
- maximo activo por empresa cuando el carril es durable;
- activacion de alertas.

La configuracion vive en `super_queue_capacity_config`. La rotacion fiscal usa
`pcs_queue_tenant_state`; ambas tablas pertenecen al ledger de `pcs-migrate`.
La API y el worker no crean esquema en runtime.

`/metrics` publica solamente agregados por carril:

- `pcs_queue_lane_pending`;
- `pcs_queue_lane_processing`;
- `pcs_queue_lane_failed`;
- `pcs_queue_lane_oldest_seconds`;
- `pcs_queue_lane_saturation_percent`;
- `pcs_queue_lane_requests_current_minute`;
- `pcs_queue_lane_query_success`.

Prometheus dispara `PCSCarrilOperativoSaturado` cuando un carril permanece al
100% o mas durante cinco minutos. El worker evalua los mismos umbrales cada
minuto y reutiliza el destinatario y cooldown de `Alertas sistema`. No se
incluyen payloads, documentos, credenciales ni datos fiscales en metricas.

## Paralelismo del worker

`PCS_WORKER_CONCURRENCY` controla trabajos simultaneos por replica (default 4,
maximo 32). `FACTURACION_RETRY_SHARDS` define carriles fiscales disjuntos
(default 8, maximo 32). `PCS_WORKER_REPLICAS` conserva escalado horizontal; el
claim durable y los leases impiden que dos replicas posean el mismo trabajo.

Subir estos valores sin medir puede agotar las conexiones PostgreSQL o la cuota
del proveedor. Deben ajustarse junto con el pool por proceso y nunca superar el
presupuesto total de conexiones del servidor.

## Prueba de capacidad para 1000 empresas

La compuerta de aceptacion debe ejecutarse en staging aislado y sin emitir
documentos fiscales reales ni imprimir en hardware productivo:

1. Aplicar las dos migraciones `20260906-001-queue-capacity-*-v1` con
   `pcs-migrate` y verificar rollback ensayado sobre copia descartable.
2. Generar carga sintetica identificable para 1000 empresas, con mezcla de
   altas de producto, trabajos de impresion simulados y reintentos fiscales con
   proveedor controlado.
3. Confirmar que no hay lectura, escritura, claim ni conteo cruzado entre
   empresas; una empresa ruidosa debe recibir su propio `429` o acumular solo
   su backlog.
4. Medir p95 de rutas criticas contra el SLO vigente de 1200 ms, CPU promedio
   menor a 80% durante diez minutos, memoria menor a 85%, conexiones, locks,
   backlog, antiguedad y tiempo de drenaje.
5. Inducir saturacion de cada carril y verificar panel, metrica Prometheus,
   alerta, cooldown, recuperacion y ausencia de duplicados.
6. Aumentar replicas/concurrencia por pasos y registrar el primer punto que
   viola un SLO. El limite operativo se fija por debajo de ese punto con margen
   de seguridad; no se deduce solamente de CPU o del numero de empresas.

## Limites de evidencia

- Un trabajo de impresion `tomado` no significa papel entregado. No se recupera
  automaticamente porque podria producir una segunda copia; se alerta y se
  concilia con el agente fisico.
- Un HTTP exitoso del proveedor fiscal no sustituye acuse oficial. Los shards
  mejoran equidad y drenaje, pero no eliminan limites o caidas de DIAN.
- Las pruebas unitarias y visuales locales validan contratos, no capacidad del
  VPS, despliegue, impresion fisica ni una emision fiscal real.
