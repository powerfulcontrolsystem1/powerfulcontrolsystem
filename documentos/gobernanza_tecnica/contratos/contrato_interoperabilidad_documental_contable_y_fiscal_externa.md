# Contrato tecnico: interoperabilidad documental contable y fiscal externa

Fecha: 2026-04-18
Estado: vigente

## Alcance

Este contrato cubre la interoperabilidad entre documentos transaccionales de compras y facturacion, su trazabilidad contable y fiscal, la validacion documental previa a contabilizacion, la cola de reintentos y conciliacion fiscal, y el repositorio documental empresarial con control de acceso y versionado.

## Endpoints cubiertos

### Repositorio documental empresarial

- `GET /api/empresa/documentos/gestion?empresa_id={id}`
- `POST /api/empresa/documentos/gestion`
- `PUT /api/empresa/documentos/gestion`
- `DELETE /api/empresa/documentos/gestion?empresa_id={id}&id={documento_id}`
- `GET /api/empresa/documentos/gestion?empresa_id={id}&action=acceso&id={documento_id}`
- `GET /api/empresa/documentos/gestion?empresa_id={id}&action=repositorio`
- `GET /api/empresa/documentos/gestion?empresa_id={id}&action=versiones&documento_codigo={codigo}`
- `POST|PUT|PATCH /api/empresa/documentos/gestion?action=versionar`

### Documentos de compras con impacto contable

- `GET /api/empresa/compras/documentos?empresa_id={id}`
- `POST /api/empresa/compras/documentos`
- `PUT /api/empresa/compras/documentos`
- `DELETE /api/empresa/compras/documentos?empresa_id={id}&documento_codigo={codigo}`

Acciones operativas cubiertas:

- `crear`
- `emitir`
- `emitir_orden`
- `solicitar_aprobacion`
- `aprobar_compra`
- `rechazar_compra`
- `recepcionar_parcial_compra`
- `recepcionar_compra`
- `contabilizar_compra`
- `validar_documentos`
- `activar`

### Facturacion documental con integracion fiscal

- `GET /api/empresa/facturacion_electronica?empresa_id={id}&action=documentos`
- `GET /api/empresa/facturacion_electronica?empresa_id={id}&action=reintentos`
- `GET /api/empresa/facturacion_electronica?empresa_id={id}&action=reconciliacion`
- `POST|PUT /api/empresa/facturacion_electronica?action=emitir`
- `POST|PUT /api/empresa/facturacion_electronica?action=anular`
- `POST|PUT /api/empresa/facturacion_electronica?action=nota_credito`
- `POST|PUT /api/empresa/facturacion_electronica?action=emitir_nota_credito`
- `POST /api/empresa/facturacion_electronica?action=procesar_reintentos`
- `POST /api/empresa/facturacion_electronica?action=reconciliar_estados`

## Persistencia canonica

- `empresa_facturacion_documentos`
- `empresa_compras_documentos`
- repositorio documental del modulo `documentos/gestion`
- cola de reintentos de facturacion electronica
- eventos contables asociados a `compras` y `facturacion`

## Entradas obligatorias

### Versionado documental

- `empresa_id`
- `id`
- al menos uno entre `url_archivo` o `hash_archivo` como nueva evidencia de version

### Validacion de acceso a documento

- `empresa_id`
- `id` o `modulo`

### Compras documentales

- `empresa_id`
- `proveedor_id` al crear
- `documento_codigo`

### Facturacion documental

- `empresa_id`
- `documento_codigo`

## Entradas opcionales relevantes

- `tipo_documento`
- `periodo_contable`
- `monto_total`
- `moneda`
- `fecha_documento`
- `observaciones`
- `estado_actual`
- `recepcion_items[]`
- `requiere_aprobacion`
- `niveles_aprobacion_requeridos`
- `proveedor_documento_ref`
- `factura_documento_ref`
- `entrada_documento_ref`
- `cliente_id`
- `cliente_email`
- `cliente_nombre`
- `pais_codigo`
- `limit`
- `offset`
- `include_inactive`
- `q`
- `fecha_desde`
- `fecha_hasta`
- `permiso`
- `include_denegados`

## Estados y salidas canonicas

### Estado documental de compras

- `borrador`
- `pendiente_aprobacion`
- `emitida`
- `rechazada`
- `recepcion_parcial`
- `recepcionada`
- `contabilizada`

### Estado de validacion documental de compras

- `no_aplica`
- `validada`
- `inconsistente`

### Estado documental de facturacion

- `borrador`
- `pendiente_emision`
- `emitida`
- `anulada`
- `ajustada`

### Estado de integracion fiscal

- `enviado`
- `fallido`
- `contingencia`
- `reconciliado`
- `no_aplica`

### Estado de documento en repositorio

- `vigente`
- `historico`

## Invariantes

1. Todo documento y toda validacion queda aislada por `empresa_id`.
2. `empresa_facturacion_documentos` y `empresa_compras_documentos` son las fuentes canonicas del estado documental operativo para facturacion y compras.
3. La unicidad de negocio se preserva por `empresa_id + tipo_documento + documento_codigo`.
4. El versionado documental no sobrescribe la version vigente en la misma fila; debe promover una nueva version y dejar la anterior en estado `historico`.
5. La consulta `action=versiones` debe devolver el historial por `documento_codigo`.
6. `action=acceso` y `action=repositorio` evalúan visibilidad documental según módulo, permiso requerido y rol efectivo del admin.
7. Un rol puede ver un documento en repositorio con `include_denegados=1`, pero el resultado debe exponer `acceso_permitido=false` cuando corresponda.
8. La solicitud de aprobación en compras mueve el documento a `pendiente_aprobacion` y fija `requiere_aprobacion=true`.
9. Una aprobación multinivel no puede saltarse niveles; solo la última aprobación habilita `emitida`.
10. `recepcionar_parcial_compra` debe persistir detalle y resumen de recepción, incluyendo diferencias e items pendientes.
11. `recepcionar_compra` no debe cerrar la recepción como completa si aún existen pendientes en el resumen.
12. `validar_documentos` debe persistir referencias externas de proveedor, factura y entrada, y clasificar el estado como `validada` o `inconsistente`.
13. Toda transición documental relevante en compras y facturación debe emitir evento contable no bloqueante.
14. La integración fiscal posterior a emitir, anular o generar nota crédito no revierte la persistencia documental aunque falle; el resultado vive en `integracion_fiscal` y en la cola de reintentos.
15. La reconciliación fiscal debe distinguir entre documentos conciliados, pendientes, en contingencia y no aplicables.
16. Un ambiente no productivo o una configuración inactiva puede llevar el estado de integración a `no_aplica` sin tratarse como incidente técnico.
17. La ausencia del documento transaccional al reprocesar una cola fiscal no debe romper el lote completo; el backend marca el item con error y actualiza el retry.
18. Si existe repositorio documental asociado, la versión vigente del documento y su `hash_archivo` deben ser la referencia canónica de evidencia antes de tratar un exporte como respaldo regulatorio.
19. Si existe firma documental asociada, la reconciliación operativa debe poder enlazar el documento transaccional con `documento_gestion_id` o con el `documento_codigo` que lo respalda.
20. Ningún exporte regulatorio debe presentarse como evidencia suficiente si no puede reconciliarse con el estado transaccional y con la versión documental vigente.

## Errores y respuestas esperadas

### Repositorio documental

- `200` en acceso, repositorio y versiones.
- `201` en `versionar` exitoso.
- `400` si faltan `empresa_id`, `id` o filtros mínimos.
- `404` si el documento solicitado no existe.

### Compras documentales

- `200` en creación y transiciones exitosas.
- `204` en activación o desactivación lógica exitosa.
- `400` por payload inválido o referencias incompletas.
- `404` si el documento no existe.
- `409` por transición inválida o recepción incompatible con el estado actual.

Mensajes observables relevantes:

- `transicion invalida`
- `la orden requiere aprobacion multinivel antes de emitir`
- `la recepcion aun tiene items pendientes`

### Facturacion documental

- `200` con `integracion_fiscal`, `cola_reintentos` y, cuando aplique, `cumplimiento_normativo`.
- `409` por transición documental inválida.
- `422` cuando falla cumplimiento normativo antes de emitir factura electrónica.
- `500` si falla persistencia o reconciliación no controlada.

## Side effects obligatorios

- emisión de eventos contables de compras y facturación.
- actualización de cola de reintentos y estados de integración fiscal.
- envío no bloqueante de correo documental al cliente cuando aplica.
- conservación de referencias externas y evidencia documental para compras.
- promoción de versiones históricas en el repositorio documental.
- preservación de vínculo operativo entre documento transaccional, repositorio documental, firma asociada y exporte regulatorio cuando exista.

## Evidencia tecnica minima

- `backend/handlers/compras.go`
- `backend/handlers/facturacion_electronica.go`
- `backend/handlers/modulos_faltantes.go`
- `backend/db/documentos_transaccionales.go`
- `backend/handlers/compras_documentos_test.go`
- `backend/handlers/facturacion_electronica_reintentos_test.go`
- `backend/handlers/modulos_faltantes_test.go`

## Runbooks relacionados

- `documentos/gobernanza_tecnica/runbooks/runbook_contingencias_integraciones_bancarias_y_conectores.md`
- `documentos/gobernanza_tecnica/runbooks/runbook_reportes_programados_y_exportaciones_contables.md`
- `documentos/gobernanza_tecnica/runbooks/runbook_reconciliacion_documental_fiscal_y_contable_externa.md`
- `documentos/gobernanza_tecnica/runbooks/runbook_versionado_documental_y_firmas_externas.md`

## ADRs relacionados

- `ADR-0001-frontera-multiempresa-empresa-id.md`
- `ADR-0002-postgresql-runtime-canonico-vps.md`