# Plan final para produccion

Fecha de corte: 2026-07-16  
Estado: plan de ejecucion aprobado para preparar una liberacion controlada.  
Alcance: arquitectura, datos, procesos, seguridad operativa, despliegue y API
movil. No autoriza por si solo un despliegue general ni reemplaza las pruebas de
staging, proveedores sandbox, restauracion y aprobacion humana.

## Resumen ejecutivo

Powerful Control System ya tiene una base funcional importante: monolito Go con
PostgreSQL, aislamiento empresarial en los flujos recientes, permisos,
sesiones, APIs web y una fachada movil `/api/v1`, probes `/health` y `/ready`,
Compose separado y controles de seguridad/documentos privados. Tambien existen
las primeras piezas de una plataforma escalable: `pcs-migrate`, `pcs-worker`,
ledger de migraciones, cola PostgreSQL, outbox e idempotencia movil.

La separacion ha avanzado, pero todavia no esta completa. El bootstrap heredado
sigue disponible solo para la corrida inicial del migrador; API y worker fallan
cerrado ante DDL de runtime en produccion. Los timers de negocio ya no arrancan
desde la API: el worker registra handlers y schedules durables, y la outbox ya
publica el evento de cobro de carrito. Persisten productores por migrar
(correo, DIAN, documentos y otros dominios), la extraccion completa del legado
y la evidencia operativa de staging. Por tanto, no se habilitan replicas ni se
declara produccion general lista solo por estos cambios de codigo.

La estrategia aprobada es **monolito modular con procesos separados**, no
microservicios prematuros. Primero se estabiliza el esquema, la ejecucion
asincrona y la consistencia de ventas/pagos; despues se desacoplan los dominios
criticos y se habilitan replicas, almacenamiento de objetos y clientes moviles.

| Dimension | Estimacion actual | Meta al terminar este plan | Base de la estimacion |
| --- | ---: | ---: | --- |
| Piloto controlado web | 68% | 90% | Flujos POS existentes y controles recientes; faltan gates externos. |
| Produccion general | 45% | 92% | Faltan migracion exclusiva, worker real, outbox y release inmutable. |
| Escalamiento horizontal | 30% | 90% | API aun ejecuta cron y no configura pool compartible. |
| Backend movil MVP | 55% | 90% | `/api/v1` existe; faltan dispositivos, refresh, cursores, push y sync formal. |

Los porcentajes son una medida de gestion del riesgo, no una sustitucion de
evidencia. Cada hito define pruebas observables que deben cumplir antes de
subir la estimacion.

## Arquitectura actual verificada

```text
Navegador web / PWA / cliente movil v1
                 |
              Nginx
                 |
          pcs-backend (API HTTP)
          |       |          |
          |       |          +-- bootstrap historico: 156 Ensure* y seeds
          |       +-- 11 workers/timers locales de mantenimiento
          +-- PostgreSQL empresas y superadministrador

pcs-migrate -> ledger inicial + Ensure de cola/outbox/idempotencia
pcs-worker  -> cola PostgreSQL, pero sin handlers registrados y con DDL propio
```

Archivos que prueban este estado:

| Componente | Archivos principales | Estado | Riesgo y prioridad |
| --- | --- | --- | --- |
| API | `backend/main.go`, `backend/handlers`, `backend/db` | Funcional; sigue concentrando bootstrap y tareas de fondo. | Bloqueador para replicas. |
| Migrador | `backend/cmd/pcs-migrate/main.go`, `backend/db/migrations.go` | Ejecutable separado, pero solo registra una fundacion y llama `Ensure*`. | Critica. |
| Worker | `backend/cmd/pcs-worker/main.go`, `backend/internal/platform/worker/worker.go` | Lee cola con `SKIP LOCKED`; mapa de handlers vacio. | Bloqueador. |
| Cola | `backend/db/async_jobs.go` | Persistente, con estados y reintentos basicos. | Critica: sin lease vencible, heartbeat, prioridad o recuperacion. |
| Outbox | `backend/db/outbox.go` | Tabla e insercion transaccional disponibles. No hay consumidor ni uso de negocio. | Critica. |
| Configuracion de rol | `backend/internal/platform/runtimeconfig/config.go` | Roles API/worker/migrate y bandera de bootstrap. | Alta: configuracion parcial y bootstrap activo por compatibilidad. |
| PostgreSQL | `backend/db`, Compose | Dos conexiones y compatibilidad PostgreSQL. No hay configuracion explicita de pool. | Alta. |
| Cache | `backend/db/db.go` | Invalida caches locales de licencias/empresa. | Alta con multiples replicas: memoria no compartida. |
| Salud | `backend/handlers/runtime_health.go` | `/health` y `/ready` publicos y sin diagnostico sensible. | Parcial: falta `/metrics` con contrato operativo. |
| Docker | `deploy/docker-compose.platform.yml` | Roles separados, usuario no root y FS de solo lectura en servicios PCS. | Alta: images locales/tags y healthchecks no alineados del todo. |
| Archivos | handlers privados, `PCS_PRIVATE_STORAGE_DIR`, volumenes Compose | Hay segregacion y descargas autenticadas recientes. | Alta: falta interfaz ObjectStorage y migracion externa. |
| API movil | `backend/handlers/mobile_api_v1.go`, `documentos/api/openapi.mobile.v1.yaml` | Auth movil, catalogo, clientes, carrito, pago, sync, facturacion y notificaciones. | Alta: faltan sesiones por dispositivo completas, refresh y cursores. |
| Observabilidad | `backend/metrics`, Prometheus/Grafana Compose | Colector y stack disponibles. | Alta: no hay metricas uniformes API/worker/cola/outbox ni alertas de SLO. |

## Arquitectura objetivo

```text
Clientes web, PWA y Android/iPhone
            |
       Nginx / balanceador
            |
    +-------+--------+
    |                |
pcs-api x N       pcs-api x N
HTTP, auth,       sin DDL, sin cron,
casos de uso      sin trabajo largo
    |                |
    +------ PostgreSQL -----+
             |              |
       migraciones       outbox/jobs
             |              |
      pcs-migrate      pcs-worker x N
      lock global      leases, handlers,
      una vez          reintentos y DLQ
                             |
                proveedores: DIAN, pagos, correo,
                WhatsApp, archivos, notificaciones
```

```text
Dominio -> transaccion PostgreSQL -> cambio de negocio + evento outbox
                                             |
                                     dispatcher con lease
                                             |
                           job idempotente por empresa/proveedor
                                             |
                           handler versionado y observable
```

```text
Empresa validada -> TenantContext -> servicio -> repositorio
        |               |              |             |
        +-- cache key --+              +-- outbox ---+-- SQL con empresa_id
        +-- storage tenant/<empresa>/  +-- auditoria y permisos
```

```text
PostgreSQL
  empresas: dominios empresariales, idempotencia movil, archivos/metadatos
  super:    sesiones, plataforma, jobs/outbox operativos y configuracion
  ambos:    migraciones con lock, pool limitado, timeouts, metricas y backup
```

```text
ObjectStorage
  ObjectStorage interface
     |-- filesystem privado temporal (adaptador de compatibilidad)
     |-- MinIO staging
     |-- S3 o R2 produccion
  objetos: tenant/<empresa_id>/<tipo>/<uuid>
  metadatos, retencion y permisos: PostgreSQL
```

```text
commit firmado -> CI -> tag semantico -> imagen por digest -> registro
     -> pcs-migrate (lock) -> despliegue API/worker -> /ready -> monitoreo
     -> rollback a digest anterior solo con migracion compatible
```

```text
Aplicacion movil
  OAuth PKCE / login -> access token corto + refresh rotativo por dispositivo
        -> /api/v1 con TenantContext y Idempotency-Key
        -> cursor/sincronizacion incremental -> conflicto declarado
        -> registro FCM/APNs -> job de notificacion -> estado de entrega
```

## Hallazgos y clasificacion

| ID | Hallazgo confirmado | Clasificacion | Impacto | Accion de cierre |
| --- | --- | --- | --- | --- |
| ARC-01 | `main.go` ejecuta bootstrap historico, provisiones y DDL mediante 156 funciones `Ensure*`. | Bloqueador | Corrupcion por DDL concurrente, inicio lento y replicas no deterministas. | Extraer migraciones y desactivar bootstrap en API. |
| ARC-02 | El ledger no usa advisory lock, checksum, orden declarativo ni transaccion completa. | Critica | Dos migradores pueden competir y no se detecta deriva. | Motor versionado con lock y checksum. |
| ARC-03 | `pcs-worker` crea schema y registra mapa vacio de handlers. | Bloqueador | Jobs se reintentan como no soportados; no procesa trabajo real. | Registro versionado de handlers y worker sin DDL. |
| ARC-04 | Cola sin lease vencible, heartbeat, recuperacion, prioridad ni archivo. | Critica | Jobs quedan en `processing` tras caida. | Lease, recovery, DLQ y metricas. |
| ARC-05 | Outbox solo tiene tabla/insert; no hay dispatcher ni productores activos. | Critica | Se pueden perder efectos externos tras commit. | Dispatcher idempotente y adopcion por dominio. |
| ARC-06 | La API inicia once workers/timers locales, incluido metricas y alertas. | Bloqueador de replicas | Ejecucion duplicada al escalar. | Migrar cada timer a job programado/worker. |
| ARC-07 | No hay `SetMaxOpenConns`, limites por replica ni politica de timeout de pool. | Alta | Saturacion de PostgreSQL bajo carga. | Configuracion tipada y presupuestos por rol. |
| ARC-08 | Caches de empresa/licencia residen por proceso. | Alta | Privilegios o estado desactualizado entre replicas. | Invalidador por evento y cache compartida o lectura segura. |
| ARC-09 | Compatibilidad SQL y `Ensure*` siguen repartidos en `db`/handlers. | Alta | Deuda Postgres-only, pruebas y cambios costosos. | Inventario, retiro gradual y repositorios por dominio. |
| ARC-10 | Operaciones de pagos/facturacion/integraciones aun mezclan HTTP, goroutines y proveedores. | Critica | Duplicacion financiera y fallos parciales. | Transaccion + idempotencia + outbox. |
| ARC-11 | Aislamiento multiempresa existe en wrappers recientes, pero no hay auditoria total por query/job/cache/storage. | Critica | Riesgo de IDOR o mezcla entre empresas. | TenantContext, pruebas negativas y RLS focal. |
| ARC-12 | Archivos privados existen, pero frontend aun monta rutas de uploads/descargas y no hay ObjectStorage. | Alta | Dependencia de host y superficie de archivos. | Adaptador, metadatos y migracion progresiva. |
| ARC-13 | CSP tiene reporte controlado, pero conserva `unsafe-inline` y fuentes amplias por compatibilidad. | Alta | Menor defensa frente a XSS. | Rollout nonce/hash en report-only antes de endurecer. |
| ARC-14 | Compose separa roles pero usa builds/tags locales y no un manifiesto inmutable de release. | Alta | Rollback no reproducible. | Registro, digest, version de schema y runbook. |
| ARC-15 | API movil v1 e idempotencia iniciales existen; faltan refresh, device/session lifecycle, push y cursores. | Alta | Cliente movil no tiene contrato completo de operacion offline. | Completar MVP movil por dominios. |
| ARC-16 | La validacion SSH directa posterior a una sincronizacion exitosa no fue estable. | Media | Operacion/remediacion remota menos confiable. | Runbook de red, monitoreo y prueba de acceso fuera del despliegue. |

Los hallazgos ARC-01 a ARC-06 son la ruta critica. ARC-11 y ARC-10 se
implementan de manera transversal antes de abrir trafico financiero masivo.
ARC-12 a ARC-15 se completan antes de escalamiento general y salida movil.

## Fases de ejecucion

### Fase 1. Migraciones y retiro de bootstrap

**Objetivo.** Hacer de `pcs-migrate` el unico proceso autorizado para cambiar
el esquema o ejecutar seeds/versionamientos controlados.

- Inventariar las 156 funciones `Ensure*`: tabla, tipo de accion (DDL, indice,
  backfill, seed, compatibilidad) y base objetivo.
- Crear paquetes de migraciones ordenados por base y version, con checksum,
  `pg_advisory_lock`, transaccion cuando PostgreSQL lo permita, registro de
  inicio/fin/error y validacion de deriva.
- Separar migraciones estructurales, datos/reparacion y seeds idempotentes.
- Mover `EnsureAsyncJobsSchema`, `EnsureOutboxSchema`, idempotencia movil y
  cada DDL restante a migraciones declaradas. API, worker y handlers deben
  fallar de forma clara si falta el schema, no crearlo.
- Mantener una ventana de compatibilidad: ejecutar migracion dos veces en
  staging, tomar backup, validar esquema y solo entonces establecer
  `PCS_RUNTIME_SCHEMA_BOOTSTRAP=0` para la API de produccion.

Archivos probables: `backend/cmd/pcs-migrate/main.go`,
`backend/db/migrations.go`, nuevo `backend/internal/platform/migrations/`,
`backend/main.go`, `backend/db/*.go`, `deploy/docker-compose.platform.yml` y
pruebas de migracion efimera.

Aceptacion: una segunda ejecucion no cambia nada; dos migradores concurrentes
no se pisan; API y worker arrancan sin `CREATE`/`ALTER`; rollback de version
compatible documentado; alerta si falta una version requerida.

Commits sugeridos: `database: inventory runtime schema mutations`,
`database: add locked checksummed migration runner`,
`database: move platform schemas into migrations`,
`runtime: disable api schema bootstrap after staging gate`.

### Fase 2. Worker de produccion

**Objetivo.** Registrar handlers reales, versionados y observables sin estado
HTTP ni mutaciones de schema.

- Crear `internal/platform/jobs` con contrato `{kind, version, timeout,
  max_attempts, backoff, idempotency, enabled}` y registro central.
- Implementar por orden: correo/notificaciones, webhooks verificados,
  facturacion/DIAN, pagos, documentos/exportaciones, reportes, mantenimiento,
  retencion, contabilidad y snapshots.
- Cada handler recibe contexto con deadline, empresa validada y referencia a
  datos internos; nunca secreto o payload crudo innecesario.
- Agregar interruptor de habilitacion por tipo de job, circuit breaker de
  proveedor y auditoria redactada.

Aceptacion: worker arranca sin DDL, expone health interno, procesa un job por
cada familia inicial, mantiene idempotencia tras reintento y registra metricas.

Commits: `worker: add typed handler registry`, `worker: move email and
notifications to handlers`, `worker: add provider timeouts and job metrics`.

### Fase 3. Cola persistente confiable

**Objetivo.** Convertir `pcs_async_jobs` en una cola operable con multiples
workers.

- Agregar `lease_until`, heartbeat, owner/version, prioridad, cancelacion,
  clave idempotente, `completed_at`, `dead_at`, politica de retencion y tabla
  de archivo/DLQ.
- Recuperar jobs cuyo lease vence; usar `FOR UPDATE SKIP LOCKED`; calcular
  backoff exponencial con jitter; limitar concurrencia por tipo y empresa.
- Redactar causas almacenadas y separar detalle tecnico protegido de mensaje
  operativo. Crear metricas de pendiente, procesamiento, reintento, muerto,
  edad y tasa por tipo/proveedor.

Aceptacion: caida de worker recupera jobs vencidos, doble worker no duplica,
un job cancelado no se ejecuta y DLQ puede reintentarse controladamente.

Commits: `queue: add leases and abandoned job recovery`, `queue: add priority
cancelation and dead letter archive`, `queue: expose durable job metrics`.

### Fase 4. Transactional outbox conectada

**Objetivo.** Garantizar que un cambio confirmado y su efecto externo no se
separen ante una caida.

- Ampliar outbox con estado, lease, intentos, clave de deduplicacion, version,
  disponibilidad, error redactado y publicacion.
- Crear dispatcher que reclama eventos, los transforma en jobs idempotentes y
  marca exactamente el evento procesado; recuperar leases vencidos.
- Adoptar primero: cierre/anulacion de ventas, pagos, factura/DIAN, inventario,
  correos, WhatsApp/notificaciones, documentos, contabilidad y webhooks.
- La mutacion de negocio y el evento deben ir en la misma transaccion.

Aceptacion: simulacion de caida despues del commit entrega una sola vez el
efecto observable; reintentos no duplican factura/pago/stock/notificacion.

Commits: `outbox: add leased dispatcher contract`, `outbox: publish sales and
payment events`, `outbox: migrate fiscal and notification effects`.

### Fase 5. Separacion final API, worker y migrador

**Objetivo.** Dejar `pcs-api -> HTTP`, `pcs-worker -> trabajo asincorno`,
`pcs-migrate -> schema`.

- Migrar los once arranques actuales de `main.go`: metricas, alertas super,
  retencion, estado/vencimiento de licencias, snapshots VPS, mantenimiento de
  agentes, parametros legales, cobranza, asientos y control electrico.
- Sustituir timers locales por jobs programados con lock/logica de schedule
  central; retirar goroutines de handlers de pagos una vez que usen outbox.
- Añadir apagado gradual: dejar de reclamar, terminar job dentro de deadline,
  liberar lease si no termina y cerrar conexiones.

Aceptacion: al levantar dos APIs no se duplica ningun trabajo; dos workers
procesan la cola sin solaparse; `main.go` no inicia jobs de negocio.

Commits: `architecture: move scheduled jobs from api to worker`,
`runtime: add graceful worker shutdown`, `api: remove background schedulers`.

### Fase 6. PostgreSQL y escalabilidad

**Objetivo.** Preparar consultas, pool y transacciones para replicas y volumen.

- Crear configuracion de pool por rol: max/idle con presupuesto global,
  `ConnMaxLifetime`, `ConnMaxIdleTime`, connect/query/transaction timeout y
  `application_name` por proceso.
- Auditar N+1, filtros, paginacion, orden, consultas sin indice y transacciones
  largas en ventas, inventario, caja, pagos, facturacion, reportes y auditoria.
- Agregar indices compuestos empezando por `(empresa_id, fecha/id/estado)` y
  validar con `EXPLAIN (ANALYZE, BUFFERS)` en datos anonimizados.
- Aplicar locking optimista/versionado a inventario, carrito, caja y pagos; no
  particionar hasta medir, pero preparar archivado de auditoria, jobs y eventos.

Aceptacion: presupuesto de conexiones documentado por replica; consultas
criticas tienen plan medido; conflictos concurrentes devuelven respuesta
determinista y no dejan movimientos parciales.

### Fase 7. Modularizacion del monolito

**Objetivo.** Reducir acoplamiento sin reescritura masiva.

Estructura destino por dominio:

```text
internal/<dominio>/
  transport/    adaptadores HTTP v0 y v1
  service/      casos de uso y transacciones
  repository/   PostgreSQL y consultas tenant-aware
  model/        contratos y value objects
  events/       outbox y consumidores
  errors/       errores tipados
```

Orden: `auth`, `sessions`, `empresas`, `ventas`, `inventario`, `caja`, `pagos`,
`facturacion`, `documentos`. Los handlers existentes se conservan como
adaptadores para no romper URLs actuales. Las dependencias permitidas fluyen
`transport -> service -> repository`; eventos salen desde service; ningun
dominio importa handlers de otro dominio. Los contratos entre dominios son
interfaces pequenas y eventos, no acceso a tablas ajenas.

Aceptacion: cada dominio critico prueba casos de uso sin HTTP, no contiene SQL
extenso en handlers y no forma ciclos de importacion.

### Fase 8. Aislamiento multiempresa

**Objetivo.** Hacer obligatorio y auditable el alcance empresarial en request,
job, cache, archivo y exportacion.

- Definir `TenantContext` inmutable desde sesion/autorizacion; nunca aceptar
  `empresa_id` del cliente como autoridad.
- Repositorios obligan `empresa_id` validado, verifican filas afectadas y usan
  indices compuestos; claves de cache, objetos, jobs y eventos lo incluyen.
- Añadir suite negativa A/B para lectura, escritura, borrado, exportacion,
  descarga, webhooks, jobs y sesiones por dispositivo.
- Evaluar RLS de PostgreSQL para ventas, facturacion, pagos, documentos y
  archivos despues de normalizar transacciones/roles; sera defensa adicional,
  no reemplazo de la aplicacion.

Aceptacion: la suite de aislamiento falla al quitar un filtro y demuestra que
un usuario, job o archivo de A no puede operar B.

### Fase 9. Consistencia e idempotencia financiera

**Objetivo.** Evitar ventas, pagos, facturas, stock y documentos duplicados o
parciales.

- Catalogar operaciones mutantes: ventas, carrito, pagos, caja, inventario,
  compras, creditos, reservas, contabilidad, facturacion, webhooks y sync.
- Para cada una: transaccion unica, idempotency key/hash/resultado, versionado
  optimista, validacion de filas y evento outbox en el mismo commit.
- Completar la idempotencia movil actual con expiracion, trazabilidad y
  semantica identica para API web cuando haya doble clic/reintento.

Aceptacion: repetir request/webhook/sync no crea segundo movimiento; fallos
intermedios hacen rollback y una operacion recuperada conserva su resultado.

### Fase 10. Configuracion profesional

**Objetivo.** Un unico contrato tipado para entorno, rol y secretos.

- Extender `runtimeconfig` a HTTP, pools, sesiones, CSRF, worker, cola, outbox,
  storage, observabilidad, API movil e integraciones.
- Separar defaults seguros de desarrollo, staging y produccion; fallar al inicio
  ante secreto/URL/origen/limite obligatorio ausente en produccion.
- Mantener secretos fuera de logs, imagen y respuesta HTTP; documentar cada
  variable sin valores reales y validar configuracion por rol.

Aceptacion: API/worker/migrador muestran configuracion redactada y rechazan
configuracion critica incompleta antes de aceptar trafico.

### Fase 11. Health, readiness y observabilidad

**Objetivo.** Operar el sistema por señales, no por inspeccion manual.

- Mantener `/health` liviano y `/ready` dependiente de recursos criticos;
  alinear healthchecks Compose con esas rutas.
- Publicar `/metrics` autenticado/red interna con metricas API, latencia,
  errores, pool, PostgreSQL, worker, cola, outbox, DLQ, storage, sesiones,
  cache, WebSocket e integraciones.
- Definir dashboards, SLO y alertas: readiness, error rate, latencia, pool
  agotado, lease vencido, DLQ, proveedor degradado, almacenamiento y backup.

Aceptacion: un fallo de DB/proveedor/dead job cambia alerta y runbook; cada
proceso se distingue por etiqueta de rol/instancia/version.

### Fase 12. Almacenamiento escalable

**Objetivo.** Retirar dependencia de disco local sin romper archivos actuales.

- Introducir interfaz `ObjectStorage` con adaptador filesystem privado, MinIO
  staging y S3/R2 produccion. Guardar metadatos, hash, tenant, tipo, retencion
  y auditoria en PostgreSQL; nombre interno UUID.
- Mantener descarga autenticada, URL firmada de corto plazo cuando corresponda,
  MIME validado y separacion `tenant/<empresa_id>/`.
- Migrar por lote verificable, con checksum/rollback; no exponer buckets ni
  paths al cliente. Agregar subida reanudable para movil.

Aceptacion: un archivo privado de A no es accesible por B ni por URL directa;
migracion parcial se puede reanudar y restaurar.

### Fase 13. Docker y releases inmutables

**Objetivo.** Desplegar la misma imagen comprobada, con rollback previsible.

- Construir API/worker/migrador desde commit/tag, escanear, publicar por digest
  y desplegar manifiesto con version de aplicacion/esquema.
- Sustituir `latest`/tags flotantes por digests cuando el proveedor lo permita;
  conservar usuario no root, capabilities retiradas, limites CPU/RAM,
  networks internas, healthchecks y `.dockerignore` estricto.
- Secuencia: CI -> tag -> image -> scan/SBOM -> backup -> migrate con lock ->
  API/worker -> readiness -> smoke -> monitoreo. Rollback solo si la migracion
  es compatible o existe restauracion ensayada.

Aceptacion: release es reproducible desde un digest; version/esquema se ven en
observabilidad; rollback ensayado en staging sin perdida de datos.

### Fase 14. Gobierno de modulos

**Objetivo.** Declarar que se ofrece, a quien y bajo que dependencia.

- Crear manifiesto por modulo: estado, version, responsable, permisos, tablas,
  endpoints, jobs, eventos, configuracion, storage, integraciones y pruebas.
- Estados: `estable`, `piloto`, `experimental`, `deshabilitado`. Un modulo
  deshabilitado no carga menus, jobs ni webhooks empresariales.
- Primer lanzamiento propuesto: autenticacion, empresas, usuarios/roles,
  productos/inventario basico, ventas/carrito, caja, clientes, licencias,
  documentos/impresion, facturacion electronica solo tras aceptacion DIAN,
  reportes esenciales y correo corporativo ya validado.
- Mantener como piloto hasta evidencia: IA/agentes, soporte remoto/WebRTC,
  domotica, verticales especializados, integraciones marketplace, GPS y
  automatizaciones externas.

Aceptacion: cada modulo visible declara contrato y estado; no hay endpoint/job
activo de un modulo deshabilitado.

### Fase 15. Aplicacion movil

**Objetivo.** Completar el contrato estable para Android/iPhone sin duplicar
reglas del POS web.

- Completar autenticacion: OAuth PKCE, access corto, refresh rotativo, sesiones
por dispositivo, revocacion, metadatos de dispositivo y limites por usuario.
- Evolucionar listados a cursores estables, `fields` cerrados, errores uniformes
  y versionamiento compatible. Completar OpenAPI y pruebas contractuales.
- Formalizar sincronizacion incremental/offline: cursor de cambio, conflictos,
  idempotencia, reintento y registro de dispositivo; no replicar datos de otras
  empresas.
- Implementar registro FCM/APNs, preferencias empresariales, outbox/job de
  push y trazabilidad de entrega. Archivos usan ObjectStorage/URLs firmadas.
- MVP movil: login/empresa, catalogo, clientes, carrito, pago, ventas,
  documentos fiscales consultables, notificaciones y sincronizacion basica.

Aceptacion: cliente de referencia consume OpenAPI, revive sesion de forma
segura, sincroniza sin duplicar venta y recibe push solo de su empresa.

## Backlog ejecutable y ordenado

| ID | Titulo y detalle tecnico | Prioridad / fase | Dependencias | Aceptacion | Riesgo | Commit sugerido | Estado inicial |
| --- | --- | --- | --- | --- | --- | --- | --- |
| PROD-001 | Inventariar cada `Ensure*`, su base y clase de mutacion. | Critica / 1 | Ninguna | Inventario versionado y sin DDL sin clasificar. | Bajo | `database: inventory runtime schema mutations` | Completado 2026-07-16; quedan clases por migrar antes de apagar bootstrap. |
| PROD-002 | Rehacer ledger con advisory lock, checksum y estados. | Critica / 1 | 001 | Dos migradores no compiten; deriva detectada. | Medio | `database: complete migration ledger and advisory lock` | Completado 2026-07-16; requiere ensayo de staging con ledger historico. |
| PROD-003 | Convertir cola, outbox e idempotencia en migraciones declaradas. | Critica / 1 | 002 | API/worker no ejecutan DDL de esas tablas. | Medio | `database: move platform schemas into migrations` | Completado para esquemas plataforma 2026-07-16; DDL legado permanece en bootstrap. |
| PROD-004 | Extraer DDL/seeds de API por lotes y cubrirlos con pruebas. | Critica / 1 | 001-003 | `PCS_RUNTIME_SCHEMA_BOOTSTRAP=0` pasa staging. | Alto | `database: migrate legacy bootstrap batch 01` | Parcial 2026-07-16; Nextcloud tecnico catalogado, faltan los demas dominios y staging. |
| PROD-005 | Desactivar bootstrap de API tras gate de staging. | Bloqueador / 1 | 004 | Arranque API no emite DDL. | Alto | `runtime: disable api schema bootstrap` | Pendiente |
| PROD-006 | Definir registro tipado de handlers/versiones. | Critica / 2 | 003 | Worker tiene handlers declarados y timeout. | Medio | `worker: add production handler registry` | Parcial 2026-07-16; registro, timeout y validacion listos, faltan handlers de negocio. |
| PROD-007 | Migrar correo y notificaciones a handlers. | Critica / 2 | 006 | Reintento no duplica envio. | Medio | `worker: move notifications to durable jobs` | Pendiente |
| PROD-008 | Migrar DIAN, pagos y documentos a handlers idempotentes. | Critica / 2 | 006-007 | Repeticion no duplica efecto fiscal/financiero. | Alto | `worker: add fiscal and payment handlers` | Pendiente |
| PROD-009 | Migrar reportes, retencion y contabilidad. | Alta / 2 | 006 | Jobs largos no bloquean HTTP. | Medio | `worker: add maintenance and report handlers` | Pendiente |
| PROD-010 | Agregar lease, heartbeat y recuperacion. | Critica / 3 | 003 | Job abandonado vuelve a pendiente con seguridad. | Medio | `queue: add leases and abandoned recovery` | Completado en la base 2026-07-16; falta validar carga en staging. |
| PROD-011 | Agregar prioridad, cancelacion, DLQ y archivo. | Alta / 3 | 010 | DLQ trazable y reintento manual seguro. | Medio | `queue: add dead letter operations` | Parcial 2026-07-16; prioridad/cancelacion/DLQ listos, falta archivo y operacion UI. |
| PROD-012 | Publicar metricas de cola. | Alta / 3 | 010 | Edad/estado/tasa visibles por tipo. | Bajo | `queue: add metrics and retention` | Parcial 2026-07-16; metricas internas disponibles, falta superficie operativa y retencion. |
| PROD-013 | Crear dispatcher outbox con lease/deduplicacion. | Critica / 4 | 006,010 | Evento confirmado produce un unico job. | Alto | `outbox: implement production dispatcher` | Parcial 2026-07-16; dispatcher seguro sin productores de negocio. |
| PROD-014 | Emitir outbox desde venta, pago e inventario. | Critica / 4 | 013 | Commit y evento son atomicos. | Alto | `outbox: publish commerce events` | Pendiente |
| PROD-015 | Emitir outbox desde DIAN, documentos, correo y contabilidad. | Alta / 4 | 013 | Efectos externos recuperables. | Alto | `outbox: publish integration events` | Pendiente |
| PROD-016 | Sustituir timers de `main.go` por schedules del worker. | Bloqueador / 5 | 006-013 | Dos APIs no ejecutan cron duplicado. | Alto | `architecture: remove scheduled jobs from api` | Pendiente |
| PROD-017 | Retirar goroutines de handlers de pago/integracion. | Critica / 5 | 014-015 | Request termina sin trabajo externo directo. | Alto | `payments: route async effects through outbox` | Pendiente |
| PROD-018 | Agregar apagado gradual y health del worker. | Alta / 5 | 006,010 | SIGTERM no abandona trabajo sin lease. | Medio | `worker: add graceful shutdown and readiness` | Parcial 2026-07-16; readiness loopback y cierre por SIGTERM listos, falta evidencia de carga/staging. |
| PROD-019 | Configurar pools, timeouts y presupuesto de conexiones. | Alta / 6 | 005 | Cada rol respeta limite documentado. | Medio | `database: configure postgres pools by role` | Parcial 2026-07-16; limites por rol configurados, falta medicion de capacidad. |
| PROD-020 | Medir/indizar rutas POS, inventario, caja y reportes. | Alta / 6 | 019 | Planes de consulta aceptados en dataset anonimo. | Medio | `database: optimize critical tenant queries` | Pendiente |
| PROD-021 | Implementar control optimista de stock/caja/carrito. | Critica / 6,9 | 020 | Conflictos no duplican ni pierden saldo. | Alto | `commerce: add optimistic concurrency guards` | Pendiente |
| PROD-022 | Extraer `auth` y `sessions` a modulo interno. | Alta / 7 | 005 | Casos de uso sin HTTP; compatibilidad de rutas. | Medio | `architecture: modularize auth and sessions` | Pendiente |
| PROD-023 | Extraer empresas y TenantContext. | Critica / 7,8 | 022 | Repositorios reciben tenant validado. | Alto | `tenancy: introduce required tenant context` | Pendiente |
| PROD-024 | Extraer ventas e inventario a servicios/repositorios. | Alta / 7 | 021,023 | Handler sin SQL complejo ni acceso cross-domain. | Alto | `architecture: modularize sales and inventory` | Pendiente |
| PROD-025 | Extraer caja, pagos, facturacion y documentos. | Critica / 7,9 | 024 | Transacciones y eventos por dominio. | Alto | `architecture: modularize financial domains` | Pendiente |
| PROD-026 | Suite negativa A/B para datos, archivos, jobs y exports. | Critica / 8 | 023 | Intentos cross-tenant son rechazados. | Alto | `security: add tenant isolation contract suite` | Pendiente |
| PROD-027 | Evaluar/activar RLS focal con pruebas de rol DB. | Alta / 8 | 026 | Defensa adicional sin romper migraciones. | Alto | `database: add scoped RLS for critical data` | Pendiente |
| PROD-028 | Catalogar operaciones mutantes/idempotencia. | Critica / 9 | 023 | Matriz completa de llaves/resultados. | Medio | `consistency: inventory mutating operations` | Pendiente |
| PROD-029 | Completar transacciones/outbox de ventas, pagos y caja. | Critica / 9 | 014,021,028 | Double click/webhook no duplica. | Alto | `commerce: make sales and payments atomic` | Pendiente |
| PROD-030 | Completar compras, credito, reserva, contabilidad y DIAN. | Alta / 9 | 029 | Rollback y reintento comprobados. | Alto | `finance: extend idempotency to remaining domains` | Pendiente |
| PROD-031 | Centralizar configuracion tipada por rol. | Alta / 10 | 005,019 | Produccion falla cerrado ante variable critica. | Medio | `config: centralize typed runtime settings` | Parcial |
| PROD-032 | Alinear `/health`, `/ready`, `/metrics`, dashboards y alertas. | Alta / 11 | 018,019 | Degradacion observable y alertada. | Medio | `operations: complete health readiness metrics` | Parcial |
| PROD-033 | Crear adaptador ObjectStorage y metadatos de archivo. | Alta / 12 | 023,031 | Descarga privada por tenant sigue funcionando. | Medio | `storage: add object storage abstraction` | Pendiente |
| PROD-034 | Migrar archivos a MinIO staging y S3/R2 productivo. | Alta / 12 | 033 | Lote reanudable con checksum/rollback. | Alto | `storage: migrate tenant private objects` | Pendiente |
| PROD-035 | Pipeline de imagenes por digest/SBOM/escaneo. | Alta / 13 | 031 | Release reproducible desde tag/digest. | Medio | `deploy: publish immutable runtime images` | Parcial |
| PROD-036 | Ensayar release/rollback y restauracion en staging. | Bloqueador / 13 | 005,035 | RPO/RTO y rollback con evidencia. | Alto | `operations: validate staged release rollback` | Pendiente |
| PROD-037 | Crear manifiesto/estado por modulo y feature flags. | Alta / 14 | 023 | Modulo deshabilitado no deja jobs/rutas activas. | Medio | `platform: add module lifecycle manifest` | Pendiente |
| PROD-038 | Clasificar primer piloto y retirar exposicion experimental. | Alta / 14 | 037 | Catalogo de piloto aprobado y probado. | Medio | `release: define pilot module set` | Pendiente |
| PROD-039 | Completar sesiones moviles/dispositivo/refresh/PKCE. | Alta / 15 | 022,023 | Revocacion por dispositivo comprobada. | Alto | `mobile: complete device session lifecycle` | Parcial |
| PROD-040 | Cursor, delta sync y conflictos movil. | Alta / 15 | 021,029,039 | Offline no duplica ni mezcla empresas. | Alto | `mobile: add cursor sync and conflicts` | Parcial |
| PROD-041 | FCM/APNs por outbox y preferencias. | Media / 15 | 013,039 | Push trazable y aislado. | Medio | `mobile: add push registration and delivery` | Pendiente |
| PROD-042 | Completar OpenAPI/contract tests y cliente de referencia. | Alta / 15 | 039-041 | Contrato v1 validado automaticamente. | Medio | `mobile: finalize v1 openapi contracts` | Parcial |

## Dependencias y hitos

| Hito | Condicion objetiva | Tareas que lo cierran | Preparacion estimada |
| --- | --- | --- | ---: |
| H1. Esquema controlado | Migraciones bloqueadas/versionadas; API y worker sin DDL; bootstrap apagado. | 001-005 | 55% |
| H2. Asincronia confiable | Handlers, leases, DLQ, outbox y timers fuera de API. | 006-018 | 67% |
| H3. Escalamiento horizontal | Pools, cache coherente, storage adaptable y observabilidad. | 019-021, 031-034 | 78% |
| H4. Dominios criticos | Auth, empresas, ventas, inventario, caja, pagos, facturacion y documentos delimitados. | 022-030 | 86% |
| H5. Release candidate | Imagen inmutable, restore/rollback ensayado y modulos clasificados. | 035-038 | 92% |
| H6. Backend movil listo | Sesiones, sync, push y OpenAPI v1 completo. | 039-042 | 95% |

Dependencias principales: H2 depende de H1; H3 necesita H1/H2; H4 usa el
TenantContext de H3/H1; H5 requiere H1-H4; H6 necesita H2, H4 y ObjectStorage
cuando habilite adjuntos. Ningun release general salta H1, H2, las pruebas de
aislamiento de H4 ni la restauracion de H5.

## Ejecucion registrada 2026-07-16 - Base de Fase 1 a 4 y 6

Se implemento la primera base tecnica sin apagar el bootstrap historico:

- `schema_migrations` ahora mantiene checksum, estado, ejecutor y una corrida
  auditable. Cada migracion catalogada toma un advisory lock transaccional y se
  aplica junto con su registro; si el catalogo deriva, el migrador falla
  cerrado.
- `pcs_async_jobs`, `pcs_outbox_events` y la idempotencia movil pasan al
  catalogo de `pcs-migrate`. API y worker verifican el esquema en modo lectura,
  pero no hacen DDL para esos componentes.
- La cola incorpora lease, heartbeat, recuperacion de trabajo vencido,
  prioridad, cancelacion, reintento, estado muerto e idempotencia por hash. El
  worker valida un registro tipado con version, timeout y maximo de intentos.
- La outbox incorpora lease, reintento e idempotencia. Su dispatcher solo
  acepta tipos de job registrados; un evento no reconocido no se pierde ni se
  ejecuta como accion externa.
- Cada rol recibe limites de pool PostgreSQL independientes. Los valores de
  produccion se deben dimensionar con metricas de staging antes de habilitar
  replicas.

No son condiciones para declarar produccion general lista: quedan 156
`Ensure*` inventariados, el bootstrap legado continua activo y once timers de
la API aun deben migrarse al worker. Tampoco se han migrado los productores de
negocio de correo, DIAN, pagos, documentos o reportes. El siguiente lote debe
ser `PROD-004`, por grupos de esquema con ensayo de staging y rollback.

Actualizacion adicional: `pcs-worker` expone solo dentro del contenedor
`/health` y `/ready` en loopback. Docker lo sondea sin publicar puerto; la
readiness exige un ciclo de trabajo correcto y conexion activa a PostgreSQL.
Un fallo de batch deja el worker no listo hasta el siguiente ciclo exitoso y
`SIGTERM` apaga el servidor de salud antes de concluir el proceso.

## Checklist de entrada a produccion

### Piloto controlado

- [ ] Migraciones aplicadas dos veces en staging y API/worker sin DDL.
- [ ] Worker procesa correo y operaciones POS prioritarias con lease y DLQ.
- [ ] Venta, pago, inventario y caja son idempotentes y transaccionales.
- [ ] Suite A/B multiempresa, archivos y exportaciones aprobada.
- [ ] Backups restaurados en entorno desechable con RPO/RTO medido.
- [ ] `/health`, `/ready`, logs redactados, alertas y runbooks revisados.
- [ ] Solo modulos `estable`/`piloto` autorizados; integraciones con sandbox.

### Produccion general

- [ ] H1 a H5 cerrados con evidencia de CI/Linux, staging y rollback.
- [ ] Releases por digest y version de schema; no tags flotantes para servicios
      propios o de riesgo.
- [ ] CSP endurecida tras periodo report-only sin violaciones no resueltas.
- [ ] Proveedores de pago, DIAN, correo, WhatsApp, Nextcloud/OnlyOffice probados
      con cuentas autorizadas, firmas y reintentos idempotentes.
- [ ] Capacidad de PostgreSQL, workers y storage medida para la carga objetivo.

### Escalamiento horizontal

- [ ] Dos o mas APIs no ejecutan cron ni comparten cache insegura.
- [ ] Dos o mas workers recuperan leases sin duplicar efectos.
- [ ] Pool presupuestado, timeout aplicado y dashboards de DB/cola activos.
- [ ] ObjectStorage externo probado para archivos privados y movil.

### Publicacion movil

- [ ] OpenAPI v1 completo, versionado y contract-tested.
- [ ] Refresh rotativo, PKCE, sesiones/dispositivos y revocacion por empresa.
- [ ] Cursor/delta sync, conflictos, idempotencia y offline probados en red
      intermitente.
- [ ] Push FCM/APNs aislado por empresa y consentimiento/preferencias.

## Estrategia de releases y rollback

1. Crear rama y cambios pequenos por tarea del backlog; exigir pruebas de
   servicio/repositorio y contratos de API antes de integrar.
2. En CI: `go test`, `go test -race` Linux, `go vet`, vulnerabilidades, secretos,
   SBOM, Compose, imagen, IaC y `git diff --check`.
3. Generar tag semantico, changelog y artefactos API/worker/migrador por digest.
4. En staging: backup, `pcs-migrate` con advisory lock, segunda ejecucion,
   despliegue gradual, `/ready`, smoke multiempresa, proveedor sandbox y carga.
5. Produccion: migraciones hacia adelante compatibles, despliegue API/worker
   por digest, observacion reforzada y ventana de rollback.
6. Rollback: volver a digest anterior solo con schema compatible. Si no lo es,
   detener trafico afectado, restaurar backup probado y ejecutar el runbook;
   nunca borrar volumenes, jobs ni evidencias para aparentar exito.

## Riesgos que se mantienen hasta completar las fases

- El mayor riesgo operativo actual es activar multiples replicas antes de H2:
  la API puede duplicar cron y el worker no procesa handlers reales.
- El mayor riesgo de datos es retirar el bootstrap sin haber trasladado cada
  mutacion a una migracion compatible y ensayada.
- El mayor riesgo financiero es mover pagos/DIAN a asincronia sin una clave de
  idempotencia y la outbox en la misma transaccion.
- El mayor riesgo multiempresa es asumir que los wrappers recientes cubren
  consultas, archivos, caches, jobs y exportaciones historicas sin la suite A/B.
- El mayor riesgo de despliegue es usar una imagen diferente a la validada o
  hacer rollback sobre una migracion incompatible.

## Primera fase que debe implementarse

Se debe iniciar por **Fase 1: migraciones y retiro de bootstrap**.

Es la dependencia de todas las demas: ni worker, cola, outbox, replicas,
ObjectStorage ni modularizacion pueden ser confiables mientras API y worker
alteran tablas durante el arranque. Afectara primero
`backend/cmd/pcs-migrate/main.go`, `backend/db/migrations.go`,
`backend/internal/platform/runtimeconfig/config.go`, `backend/main.go`, las
funciones `Ensure*` inventariadas y `deploy/docker-compose.platform.yml`.

El resultado exigido no es solo un ledger: es que un despliegue aplique una
lista verificable de migraciones bajo lock, que una segunda ejecucion no cambie
nada y que API/worker no emitan DDL. Esto desbloquea H2, permite un rollback
razonado y elimina el principal obstaculo para ejecutar replicas sin introducir
parches nuevos.

## Evidencia y documentos relacionados

- `documentos/preparacion_produccion_y_app_movil.md`: fundacion ya incorporada
  y limites declarados al 2026-07-14.
- `documentos/production_readiness_final.md`: gates pendientes de staging,
  restore, carga e integraciones externas.
- `documentos/plan_101_arquitectura_modular.md`: decisiones de monolito modular
  previas, complementadas por este orden de ejecucion.
- `documentos/comandos_codex.md`: comandos operativos, preflight, `rs` y
  evidencia de despliegue.

Este documento reemplaza cualquier interpretacion optimista de que la sola
existencia de los binarios de migracion/worker equivale a produccion lista. La
base existe; el siguiente trabajo debe conectarla, probarla y retirarle los
atajos de runtime en el orden descrito.
