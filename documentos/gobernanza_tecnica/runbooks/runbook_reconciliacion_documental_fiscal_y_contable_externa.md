# Runbook: reconciliacion documental fiscal y contable externa

Fecha: 2026-04-18
Estado: vigente

## Sintomas cubiertos

- una compra queda recepcionada o contabilizada sin validación documental clara.
- la facturación muestra `integracion_fiscal.estado_envio=fallido` o deja elementos en `cola_reintentos`.
- el resumen de `reconciliacion` muestra documentos pendientes, en contingencia o no conciliados.
- un documento existe en repositorio o historial, pero el rol operativo no logra acceder a él.
- el cierre operativo encuentra diferencias entre documento de compra, documento fiscal y evidencia del repositorio empresarial.

## Alcance

Aplica al frente documental empresarial compuesto por:

- `/api/empresa/compras/documentos`
- `/api/empresa/facturacion_electronica?action=documentos`
- `/api/empresa/facturacion_electronica?action=reintentos`
- `/api/empresa/facturacion_electronica?action=reconciliacion`
- `/api/empresa/facturacion_electronica?action=procesar_reintentos`
- `/api/empresa/facturacion_electronica?action=reconciliar_estados`
- `/api/empresa/documentos/gestion?action=acceso`
- `/api/empresa/documentos/gestion?action=repositorio`
- `/api/empresa/documentos/gestion?action=versiones`

## Fuentes de evidencia

- `backend/handlers/compras.go`
- `backend/handlers/facturacion_electronica.go`
- `backend/handlers/modulos_faltantes.go`
- `backend/db/documentos_transaccionales.go`
- `backend/handlers/compras_documentos_test.go`
- `backend/handlers/facturacion_electronica_reintentos_test.go`
- `backend/handlers/modulos_faltantes_test.go`

## Verificaciones iniciales

1. Confirmar `empresa_id`, `documento_codigo` y si el caso nace en compras, facturación o repositorio documental.
2. Consultar el documento de compras por `empresa_id` y revisar `estado_documento`, `validacion_documental_estado`, `proveedor_documento_ref`, `factura_documento_ref` y `entrada_documento_ref`.
3. Consultar el documento de facturación y revisar `estado_documento`, `numero_legal`, `codigo_validacion`, `integracion_fiscal` y si existe `cola_reintentos` asociada.
4. Si hay versiones documentales, consultar `action=versiones` y confirmar cuál es la fila vigente y cuáles quedaron históricas.
5. Si el problema es de visibilidad, consultar `action=acceso` o `action=repositorio` con el mismo rol operativo que reporta la incidencia.
6. Si hubo exporte regulatorio o reporte usado como respaldo, consultar su dataset, formato y consistencia para verificar que no esté reemplazando indebidamente la evidencia documental.

## Causas probables

- la compra fue recepcionada parcialmente y aún tiene items pendientes.
- el flujo de validación documental no recibió todas las referencias externas obligatorias.
- la integración fiscal falló de forma no bloqueante y dejó item en reintentos.
- la reconciliación fue consultada pero no aplicada, por lo que el resumen no se reflejó aún en los documentos.
- el rol del operador no tiene acceso de actualización o visualización al módulo documental involucrado.
- la nueva versión del documento existe, pero el equipo está revisando la fila histórica y no la vigente.

## Acciones de recuperacion

1. Si la compra está en `recepcion_parcial`, revisar `recepcion_resumen` antes de contabilizar; no forzar cierre mientras haya `items_pendientes`.
2. Si la compra no tiene referencias suficientes, repetir `validar_documentos` con `proveedor_documento_ref`, `factura_documento_ref` y `entrada_documento_ref` consistentes.
3. Si facturación quedó con `estado_envio=fallido`, listar `action=reintentos` y confirmar si el item está vencido o listo para reproceso.
4. Ejecutar `procesar_reintentos` cuando la causa ya haya sido corregida; no reprocesar en bucle si la configuración o el proveedor siguen fallando.
5. Consultar `action=reconciliacion` para obtener el resumen antes de aplicar cambios automáticos.
6. Ejecutar `action=reconciliar_estados&aplicar=true` solo cuando el equipo quiera reflejar el ajuste en el estado documental persistido.
7. Si el problema es de acceso, usar `action=acceso` con `id` o `modulo` para confirmar si el bloqueo es por rol y no por ausencia del documento.
8. Si hay duda sobre la evidencia vigente, revisar `action=versiones` y confirmar que la última versión tenga `estado_documento=vigente`.
9. Si el flujo declara firma externa o exporte regulatorio, reconciliar además `hash_archivo`, `hash_firma`, fecha de firma y dataset exportado antes de dar por cerrado el incidente.

## Restricciones operativas relevantes

1. La integración fiscal es no bloqueante: una factura puede quedar emitida aunque la entrega fiscal falle.
2. Un item en `cola_reintentos` no significa que el documento fiscal esté perdido; significa que la persistencia documental sobrevivió y la entrega externa quedó pendiente.
3. `reconciliacion` consulta el resumen; `reconciliar_estados` puede además aplicarlo sobre la persistencia si se envía `aplicar=true`.
4. `versionar` no reemplaza la fila previa; crea una nueva versión y deja la anterior como `historico`.
5. `include_denegados=1` en repositorio sirve para diagnóstico de permisos, no para otorgar acceso efectivo.
6. Un exporte `pdf`, `xls`, `csv`, `txt` o `json` no reemplaza la existencia de la versión vigente del documento ni su firma cuando el flujo exige evidencia reforzada.

## Validacion posterior

- la compra queda en estado consistente con su evidencia documental y sin pendientes ocultos.
- la validación documental termina en `validada` cuando ya existen las referencias externas requeridas.
- la factura refleja `integracion_fiscal` coherente con el estado real y la cola de reintentos deja de crecer sin control.
- el resumen de reconciliación coincide con el estado persistido cuando se decidió aplicar cambios.
- el repositorio documental muestra una única versión vigente y el rol correcto obtiene `acceso_permitido=true`.
- si hubo firma o exporte regulatorio, ambos quedan reconciliados con la versión documental vigente y no como evidencia aislada.

## Indicadores de salida sana

- `validacion_documental_estado=validada`
- `estado_envio=enviado` o `no_aplica` cuando corresponda por ambiente/configuración
- `cola_reintentos` vacía o sin items vencidos críticos
- `acceso_permitido=true` para el rol esperado
- versión vigente visible en historial documental

## Contratos relacionados

- `documentos/gobernanza_tecnica/contratos/contrato_interoperabilidad_documental_contable_y_fiscal_externa.md`
- `documentos/gobernanza_tecnica/contratos/contrato_integraciones_bancarias_y_conectores_externos.md`
- `documentos/gobernanza_tecnica/contratos/contrato_repositorio_documental_y_firmas_externas.md`
- `documentos/gobernanza_tecnica/contratos/contrato_reportes_contables_financieros_y_exportacion_multiformato.md`

## Runbooks relacionados

- `documentos/gobernanza_tecnica/runbooks/runbook_contingencias_integraciones_bancarias_y_conectores.md`
- `documentos/gobernanza_tecnica/runbooks/runbook_versionado_documental_y_firmas_externas.md`