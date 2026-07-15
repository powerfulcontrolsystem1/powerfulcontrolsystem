# Plan preproduccion

Estado: ejecutado de forma incremental el 2026-07-15. Este plan es el control
de cambio para llevar PCS a produccion sin mezclar el trafico HTTP con DDL,
aprovisionamiento ni tareas repetibles.

## Objetivo

Mantener PCS como monolito modular Go/PostgreSQL, con limites empresariales
obligatorios, procesos escalables y contratos estables para web y movil. No se
introducen microservicios, dependencias nuevas ni una segunda regla de negocio
para POS, caja, inventario, pagos o facturacion.

## Fases y resultado

| Fase | Resultado | Estado |
|---|---|---|
| Diagnostico | Inventario de acoplamientos, handlers extensos, DDL en runtime y riesgos de despliegue. | Completada |
| Roles runtime | API, migrador y worker diferenciados; el API productivo no hace bootstrap por defecto. | Completada |
| Migraciones | Contenedor de migracion ejecuta el binario principal con rol `migrate`; aplica el bootstrap historico antes de iniciar API/worker. El ledger versionado se protege con advisory lock PostgreSQL. | Completada |
| DDL HTTP | ERP generico, documentos transaccionales y contador publico dejaron de crear tablas durante una solicitud; ahora verifican esquema y devuelven 503 si falta una migracion. Los `Ensure...Schema` restantes de handlers legados quedan inventariados para extraccion por dominio. | En transicion controlada; no escalar replicas antes de completar la extraccion |
| Procesos durables | Jobs y outbox incluyen empresa, actor/correlacion, prioridad, lease, heartbeat, vencimiento, resultado saneado e idempotencia hash; worker recupera abandonados y no crea esquema. | Completada |
| Salud operativa | `/health` comprueba vida y `/ready` comprueba PostgreSQL mas la migracion requerida, sin revelar topologia. Docker usa readiness para backend. | Completada |
| Limites de modulo | Catalogo tipado en `internal/platform/modules` y arquitectura documentada para auth, empresas, usuarios, ventas, inventario, caja, pagos, facturacion, documentos, notificaciones, IA, soporte y verticales. | Completada |
| API movil | `/api/v1` sigue siendo contrato versionado, con autorizacion empresarial, idempotencia, JSON uniforme y OpenAPI existente. | Completada |

## Controles de cierre

1. `pcs-migrate` es el unico rol que debe realizar cambios de esquema en
   produccion. Los `Ensure...Schema` heredados no se usan para provisionar
   instalaciones nuevas y se retiran por dominio antes de escalar replicas.
2. `pcs-backend` inicia con `PCS_RUNTIME_SCHEMA_BOOTSTRAP=0`.
3. `pcs-worker` no ejecuta DDL; depende de que migracion termine bien.
4. Cada trabajo durable debe tener `empresa_id` validado, actor/origen cuando
   exista, clave idempotente para efectos externos, vencimiento y reintentos
   limitados.
5. Un endpoint empresarial conserva el wrapper de permisos y usa el
   `empresa_id` del contexto como autoridad, nunca uno impuesto por el cliente.
6. Los archivos privados siguen fuera de `web/`, con nombre aleatorio, limite,
   validacion de contenido, descarga autenticada y cabeceras `no-store` y
   `nosniff`.

## Gates operativos no sustituibles por codigo

Antes de habilitar trafico productivo se debe ejecutar, sobre staging
anonimizado: restauracion de backup, carga concurrente de ventas/caja,
confirmacion de webhooks firmados, DIAN, Wompi/ePayco, Mailu/WhatsApp,
Nextcloud/OnlyOffice, y compilacion/firma Android/iPhone. Son validaciones de
infraestructura y proveedores externos; no se simulan como aceptacion real.

El compose de staging conserva almacenamiento privado, volumenes y worker
propios. El contrato estatico `tools/staging_compose_contract.mjs` y el
contrato de despliegue `tools/deploy_pipeline_contract.mjs` se ejecutan en CI.
La canalizacion rechaza publicar una rama de trabajo y evita transportar
secretos de Docker desde el equipo local.

## Rollback

Si una version falla, detener API/worker, restaurar la imagen anterior y no
ejecutar migraciones inversas destructivas automaticamente. Las migraciones
son aditivas e idempotentes; toda reversa que elimine datos requiere backup
verificado y procedimiento aprobado en `preparacion_final_produccion.md`.
