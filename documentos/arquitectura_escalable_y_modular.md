# Arquitectura escalable y modular

Actualizacion: 2026-07-15.

PCS es un monolito modular en Go y PostgreSQL. La unidad de aislamiento es la
empresa validada por sesion, wrapper de permisos y consulta SQL; no existe una
fuente de autoridad basada solo en `empresa_id` enviado por navegador o movil.

## Dominios

El catalogo ejecutable vive en `backend/internal/platform/modules/catalog.go`.
Los dominios estables son autenticacion, empresas, usuarios, ventas,
inventario y caja. Pagos, facturacion, documentos, notificaciones, IA y soporte
remoto permanecen en piloto por depender de proveedores o pruebas externas.
Los verticales se mantienen experimentales y se habilitan por empresa.

```text
HTTP web / API v1
       |
auth + CSRF/bearer + permiso + empresa validada
       |
handler de transporte
       |
regla de negocio existente / servicio de dominio
       |
repositorios db PostgreSQL + transaccion
       |
outbox transaccional -> jobs PostgreSQL -> worker
```

## Procesos

```text
pcs-migrate -- migraciones/compatibilidad -- termina
       |
       +--> pcs-backend (/health, /ready, HTTP web y API)
       +--> pcs-worker (claim SKIP LOCKED, lease, retry, dead jobs)
```

El backend productivo inicia con `PCS_RUNTIME_SCHEMA_BOOTSTRAP=0` y el worker
no ejecuta DDL. El rol `migrate` conserva temporalmente el bootstrap historico
mientras se termina la extraccion de `Ensure...Schema` heredados desde handlers
a migraciones versionadas por dominio. Los endpoints ERP, documentos
transaccionales y contador publico ya verifican migracion en vez de crear tabla.
Las migraciones versionadas del ledger se serializan con advisory lock
PostgreSQL.

## Consistencia

Ventas, inventario, caja, pagos y facturacion conservan sus reglas vigentes.
Las nuevas mutaciones moviles usan `Idempotency-Key` hash por empresa. Los jobs
persistentes tienen hash de idempotencia opcional, expiracion, actor, correlacion
y limite de reintentos. El worker reclama filas con `FOR UPDATE SKIP LOCKED`.

## Archivos y cache

Archivos privados se almacenan por empresa fuera de la raiz publica. El
adaptador S3 compatible se mantiene como siguiente cambio de infraestructura:
su adopcion exige bucket, secretos fuera de imagen, migracion por hash y prueba
de restauracion; no se activa una ruta S3 ficticia sin proveedor operativo.
La cache nunca concede autorizacion: cambios de sesion, rol, empresa o cuenta
deben invalidarla de inmediato.

## Observabilidad

`/health` no expone dependencias. `/ready` verifica las dos conexiones
PostgreSQL. Errores operativos deben conservar request/correlation/job id sin
secretos, cuerpo de webhook, token, correo completo ni DSN.
