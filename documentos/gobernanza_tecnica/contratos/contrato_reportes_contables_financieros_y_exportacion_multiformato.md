# Contrato tecnico: reportes contables, financieros y exportacion multiformato

Fecha: 2026-04-18
Estado: vigente

## Alcance

Este contrato cubre el modulo empresarial de reportes y su agregado global super para datasets operativos, contables y financieros, tablero ejecutivo, suite de reportes, exportacion multiformato, plantillas, programacion, ejecuciones y validacion de consistencia.

## Endpoints cubiertos

### Reportes empresariales

- `GET /api/empresa/reportes?empresa_id={id}&action=catalogo`
- `GET /api/empresa/reportes?empresa_id={id}&action=tablero`
- `GET /api/empresa/reportes?empresa_id={id}&action=dataset&dataset={key}`
- `GET /api/empresa/reportes?empresa_id={id}&action=export&dataset={key}&format={json|csv|txt|xls|pdf}`
- `GET /api/empresa/reportes?empresa_id={id}&action=suite`
- `GET /api/empresa/reportes?empresa_id={id}&action=export&format=json`
- `GET|POST|PUT|DELETE /api/empresa/reportes?empresa_id={id}&action=plantillas`
- `GET|POST|PUT|DELETE /api/empresa/reportes?empresa_id={id}&action=programacion`
- `GET /api/empresa/reportes?empresa_id={id}&action=ejecuciones`
- `POST /api/empresa/reportes?empresa_id={id}&action=ejecutar_programacion`
- `GET|POST /api/empresa/reportes?empresa_id={id}&action=validar_consistencia`

### Reportes globales super

- `GET /super/api/reportes_globales?action=catalogo`
- `GET /super/api/reportes_globales?action=tablero`
- `GET /super/api/reportes_globales?action=dataset&dataset={key}&modo={consolidado|individual}`
- `GET /super/api/reportes_globales?action=export&dataset={key}&modo={consolidado|individual}&format={json|csv|txt|xls|pdf}`

## Catalogo canonico de datasets empresariales

### Ejecutivos y contables

- `empresarial_tablero`
- `contable_estado_resultados`
- `contable_balance_general`
- `contable_flujo_caja`
- `contable_movimientos_financieros`
- `contable_eventos`
- `contable_asientos`
- `contable_nomina_liquidaciones`

### Operativos comerciales y de conversion

- `operativo_modulos_resumen`
- `operativo_ventas_embudo_conversion`
- `operativo_ventas_detalle`
- `reporte_de_turno`
- `operativo_top_productos`
- `operativo_top_clientes`
- `operativo_clientes_segmentacion_comercial`

### Operativos por reservas, tarifas y estaciones

- `operativo_reservas_ocupacion`
- `operativo_tarifas_ingresos`
- `operativo_tarifas_comparativo_estaciones`
- `operativo_cadena_cumplimiento`

### Operativos de inventario, compras y personal

- `operativo_inventario_bodega`
- `operativo_compras_movimientos`
- `operativo_propinas_acumulado`
- `operativo_comisiones_lavador`
- `operativo_asistencia_nomina_auditoria`

### Operativos de trazabilidad y auditoria

- `operativo_facturacion_trazabilidad`
- `operativo_auditoria_acciones`
- `operativo_vehiculos_permanencia`

## Entradas obligatorias

### Dataset o export puntual

- `empresa_id` en reportes empresariales
- `dataset` cuando se consulta `action=dataset`
- `dataset` cuando se exporta a `csv`, `txt`, `xls` o `pdf`

### Agregado global super

- administrador autenticado distinto de `sistema`
- `dataset` para `action=dataset` o `action=export`

## Entradas opcionales relevantes

- `desde`
- `hasta`
- `max_rows`
- `cierre_id`
- `empleado_nomina_id`
- `include_inactive`
- `template_codigo`
- `template_version`
- `format`
- `formatos`
- `modo`
- `empresa_id` o `empresa_ids` en agregado global

## Invariantes funcionales

1. Todo reporte empresarial opera bajo aislamiento obligatorio por `empresa_id`.
2. Todo reporte global super se limita a empresas creadas por el administrador autenticado.
3. El catalogo de datasets es fijo en codigo y debe consultarse por clave canonica.
4. `action=dataset` exige `dataset` valido y responde error si la clave no existe.
5. `action=export` sin `dataset` solo puede exportar la suite completa en `json`.
6. `action=export` sin `dataset` debe rechazar `csv`, `txt`, `xls` y `pdf`.
7. Los formatos canonicos de salida son `json`, `csv`, `txt`, `xls` y `pdf`; los alias `excel` y `tsv` se normalizan a `xls`.
8. Toda exportacion de un mismo dataset debe conservar columnas clave, resumen y conteo coherente entre formatos, salvo truncamiento informativo explicito en PDF.
9. La programacion y las plantillas no alteran el dataset base fuera del alcance de la empresa autenticada.
10. La validacion de consistencia usa hashes de estructura y contenido para detectar discrepancias entre formatos.
11. `max_rows` se normaliza y limita para evitar respuestas descontroladas.
12. El modo global solo admite `consolidado` o `individual`.
13. Un exporte regulatorio no sustituye por sí mismo la evidencia documental del repositorio ni la firma documental asociada cuando estas existan.
14. Cuando el dataset represente documentos fiscales, contables o de compras con respaldo documental, el exporte debe poder reconciliarse con `documento_codigo`, estado transaccional y versión vigente del repositorio.
15. PDF puede ser evidencia de lectura o presentación, pero para conciliación regulatoria el formato fuente debe seguir siendo trazable hacia `json`, `csv`, `txt` o `xls` y hacia el documento/version firmado si existe.

## Side effects y persistencia

### Exportacion

- no altera datos operativos
- genera descarga con nombre derivado de dataset, empresa y timestamp

### Plantillas

- persisten configuraciones de presentacion y adaptacion del dataset por empresa

### Programacion

- persiste frecuencia, formatos, hora y siguiente ejecucion
- permite frecuencia `diario`, `semanal`, `mensual` o `manual`

### Ejecuciones y consistencia

- registran evidencia de corrida
- `validar_consistencia` genera hashes, tamanos y conteos detectados por formato
- para exportes regulatorios, la corrida debe dejar claro si existe o no soporte documental/firma asociada fuera del exporte

## Comportamiento de exportacion

### `json`

- exporta el dataset completo o la suite completa si no se especifica `dataset`

### `csv`

- exporta columnas + filas delimitadas
- falla si el dataset no se especifica

### `txt`

- exporta cabecera descriptiva, resumen y contenido tabular legible
- falla si el dataset no se especifica

### `xls`

- exporta TSV con BOM y `Content-Type` de Excel compatible
- falla si el dataset no se especifica

### `pdf`

- exporta una representacion liviana en PDF de una pagina
- puede truncar el detalle y recomendar `csv` o `json` para volumen alto
- falla si el dataset no se especifica

## Salidas y errores esperados

### Reportes empresariales

- `200` con catalogo, tablero, dataset, suite, exportacion, plantillas, programacion, ejecuciones o resultado de consistencia
- `400` si faltan parametros, el dataset es invalido, el formato no es soportado o el rango de entradas es incorrecto
- `405` si el metodo HTTP no coincide con la accion
- `500` si falla la construccion interna del tablero o de la suite

Mensajes relevantes:

- `dataset es obligatorio`
- `dataset es obligatorio para formatos tabulares (csv, txt, xls o pdf)`
- `format invalido (use json, csv, txt, xls o pdf)`
- `action invalida (use catalogo, suite, dataset, tablero, export, plantillas, programacion, ejecutar_programacion, ejecuciones o validar_consistencia)`

### Reportes globales super

- `401` si no hay administrador autenticado valido
- `400` si `modo`, `dataset` o `empresa_ids` son invalidos o quedan fuera de alcance
- `405` si el metodo no coincide con la accion
- `500` si falla el tablero global

## Evidencia tecnica minima

- `backend/handlers/reportes.go`
- `backend/handlers/reportes_globales.go`
- `backend/handlers/reportes_programacion.go`
- `writeReportesDatasetExport`
- `reportesValidateDatasetConsistency`

## Evidencia operativa endurecida para exportes regulatorios

- conservar dataset, filtros, formato, timestamp y empresa origen del exporte
- conservar hash o huella de consistencia generado por `validar_consistencia` cuando el flujo lo requiera
- si el exporte resume documentos operativos sensibles, enlazar el `documento_codigo` o la referencia documental que permita reconciliar contra repositorio y firma
- no tratar un exporte aislado como sustituto de documento gestionado, firma externa o estado transaccional persistido

## Runbooks y ADRs relacionados

- `documentos/gobernanza_tecnica/runbooks/runbook_arranque_postgresql_tunel_local.md`
- `documentos/gobernanza_tecnica/runbooks/runbook_reportes_programados_y_exportaciones_contables.md`
- `documentos/gobernanza_tecnica/runbooks/runbook_reconciliacion_documental_fiscal_y_contable_externa.md`
- `documentos/gobernanza_tecnica/runbooks/runbook_versionado_documental_y_firmas_externas.md`
- `ADR-0001-frontera-multiempresa-empresa-id.md`
- `ADR-0002-postgresql-runtime-canonico-vps.md`