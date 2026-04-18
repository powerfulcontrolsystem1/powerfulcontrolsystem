# Contrato tecnico: conciliacion bancaria y cierre de periodo contable

Fecha: 2026-04-18
Estado: vigente

## Alcance

Este contrato cubre el flujo empresarial de periodos contables, importacion de extractos bancarios, conciliacion bancaria automatica por periodo, consulta de desviaciones financieras y la restriccion operativa que impide mutar movimientos cuando el periodo contable ya esta cerrado.

## Endpoints cubiertos

### Periodos contables

- `GET /api/empresa/finanzas/periodos?empresa_id={id}`
- `POST /api/empresa/finanzas/periodos`
- `PUT /api/empresa/finanzas/periodos`
- `PUT /api/empresa/finanzas/periodos?empresa_id={id}&action=cerrar&periodo={yyyy-mm}`
- `PUT /api/empresa/finanzas/periodos?empresa_id={id}&action=reabrir&periodo={yyyy-mm}`

### Movimientos financieros y tablero financiero

- `GET /api/empresa/finanzas/movimientos?empresa_id={id}`
- `POST /api/empresa/finanzas/movimientos`
- `PUT /api/empresa/finanzas/movimientos`
- `DELETE /api/empresa/finanzas/movimientos?empresa_id={id}&id={movimiento_id}`
- `GET /api/empresa/finanzas/movimientos?empresa_id={id}&action=tablero`
- `GET /api/empresa/finanzas/movimientos?empresa_id={id}&action=tablero_export&format={json|csv}`

### Extractos bancarios y conciliacion bancaria

- `GET /api/empresa/finanzas/movimientos?empresa_id={id}&action=extractos_bancarios`
- `POST /api/empresa/finanzas/movimientos?action=importar_extractos_bancarios`
- `POST /api/empresa/finanzas/movimientos?action=importar_bancario`
- `POST /api/empresa/finanzas/movimientos?action=conciliacion_bancaria_importar`
- `PUT /api/empresa/finanzas/movimientos?empresa_id={id}&action=conciliar_bancaria_auto`
- `PUT /api/empresa/finanzas/movimientos?empresa_id={id}&action=conciliar_bancos`
- `PUT /api/empresa/finanzas/movimientos?empresa_id={id}&action=conciliar_bancaria_automatica`
- `GET /api/empresa/finanzas/movimientos?empresa_id={id}&action=conciliacion_bancaria`
- `GET /api/empresa/finanzas/movimientos?empresa_id={id}&action=desviaciones_financieras`
- `GET /api/empresa/finanzas/movimientos?empresa_id={id}&action=desviaciones_periodo`
- `GET /api/empresa/finanzas/movimientos?empresa_id={id}&action=conciliacion_bancaria_export&format={json|csv|txt|xls|pdf}`
- `GET /api/empresa/finanzas/movimientos?empresa_id={id}&action=desviaciones_financieras_export&format={json|csv|txt|xls|pdf}`

### Conciliacion contable por periodo

- `GET /api/empresa/finanzas/asientos?empresa_id={id}&action=conciliacion_periodo`
- `GET /api/empresa/finanzas/asientos?empresa_id={id}&action=conciliacion`
- `GET /api/empresa/finanzas/asientos?empresa_id={id}&action=conciliar`
- `POST|PUT /api/empresa/finanzas/asientos?empresa_id={id}&action=procesar_asientos`

## Entradas obligatorias

### Cierre o reapertura de periodo

- `empresa_id`
- `periodo`
- `autorizado_por`
- `motivo_autorizacion`
- `evidencia_autorizacion`

### Importacion de extractos

- `empresa_id`
- `movimientos[]`

### Conciliacion bancaria automatica

- `empresa_id`

### Conciliacion contable por periodo

- `empresa_id`

## Entradas opcionales relevantes

- `codigo_autorizacion`
- `cuenta_bancaria`
- `banco_nombre`
- `origen`
- `auto_conciliar`
- `tolerancia_dias`
- `tolerancia_monto`
- `desde`
- `hasta`
- `periodo`
- `limit`
- `include_inactive`
- `estado_conciliacion`
- `max_reintentos`

## Persistencia canonica

- `empresa_finanzas_periodos`
- `empresa_finanzas_movimientos`
- `empresa_finanzas_bancos_movimientos`
- eventos contables y asientos asociados al modulo financiero

## Estados y normalizaciones canonicas

### Estado de periodo

- `abierto`
- `cerrado`
- `inactivo`

### Estado de conciliacion bancaria

- `pendiente`
- `conciliado`
- `con_desviacion`

### Estado resumen por periodo

- `conciliado`
- `con_pendientes`
- `con_descuadre`
- `sin_movimientos`

## Invariantes

1. Todo periodo, movimiento y extracto bancario queda aislado por `empresa_id`.
2. El cierre o reapertura de periodo exige evidencia explicita de autorizacion; sin ella el backend debe rechazar la operacion.
3. `evidencia_autorizacion` es obligatoria para cerrar o reabrir periodos.
4. El backend sanitiza y registra `autorizado_por`, `motivo_autorizacion`, `evidencia_autorizacion` y `codigo_autorizacion` dentro de observaciones y del evento contable correspondiente.
5. Un movimiento financiero no puede crearse, actualizarse ni eliminarse si su periodo contable esta cerrado.
6. La importacion de extractos bancarios es idempotente por `empresa_id + hash_movimiento`.
7. Si un extracto ya fue conciliado con un movimiento interno, una reimportacion no debe romper esa conciliacion.
8. La conciliacion bancaria automatica solo intenta emparejar extractos aun no conciliados.
9. El matching automatico usa tipo de movimiento, monto, periodo, cercania de fecha y referencias externas segun tolerancias configuradas.
10. Si no hay match valido, el extracto pasa a `con_desviacion` y no a `conciliado`.
11. La conciliacion bancaria automatica devuelve metricas resumidas de revisados, conciliados, pendientes, desviados y monto conciliado.
12. El resumen por periodo debe exponer total de extractos, conciliados, pendientes, con desviacion, montos y estado agregado.
13. La exportacion del resumen de conciliacion bancaria reutiliza el writer canonico de reportes para `json`, `csv`, `txt`, `xls` y `pdf`.
14. El tablero financiero resumido solo exporta `json` o `csv`.
15. El cierre de periodo emite evento contable `periodo_contable_cerrado`; la reapertura emite `periodo_contable_reabierto`.

## Salidas y errores esperados

### Periodos contables

- `200` en alta, actualizacion, cierre o reapertura exitosa.
- `400` si faltan `empresa_id`, `periodo` o la evidencia obligatoria.
- `400` si el payload de autorizacion es invalido.

Mensajes observables relevantes:

- `autorizado_por es obligatorio para cerrar o reabrir periodos`
- `motivo_autorizacion es obligatorio para cerrar o reabrir periodos`
- `evidencia_autorizacion es obligatoria para cerrar o reabrir periodos`

### Movimientos financieros

- `201` en creacion exitosa.
- `200` en actualizacion o eliminacion exitosa.
- `404` si el movimiento no existe.
- `409` cuando el periodo contable del movimiento esta cerrado.

### Importacion y conciliacion bancaria

- `201` para importacion exitosa, con `importacion` y opcionalmente `conciliacion_automatica`.
- `200` para consulta de extractos, resumen o conciliacion automatica manual.
- `400` por `empresa_id`, `movimientos`, `limit`, `tolerancia_dias` o `tolerancia_monto` invalidos.
- `500` si falla la construccion del resumen o la conciliacion automatica no controlada.

Mensajes observables relevantes:

- `No se pudieron importar los extractos bancarios`
- `No se pudo ejecutar la conciliacion bancaria automatica`
- `No se pudo construir la conciliacion bancaria`

### Exportaciones

- conciliacion bancaria: `json`, `csv`, `txt`, `xls`, `pdf`.
- tablero financiero: `json`, `csv`.

## Side effects obligatorios

- emision de eventos contables no bloqueantes para movimientos financieros y cierre/reapertura de periodos.
- persistencia de evidencia de autorizacion en observaciones del periodo y en el payload del evento contable.
- actualizacion de `movimiento_finanzas_id`, `estado_conciliacion`, `conciliado_en` y `conciliado_por` para extractos conciliados.
- marcacion explicita de extractos sin match como `con_desviacion`.

## Evidencia tecnica minima

- `backend/handlers/finanzas.go`
- `backend/db/finanzas.go`
- `backend/db/finanzas_conciliacion_bancaria.go`
- `backend/handlers/eventos_contables_modulos_test.go`
- `backend/db/finanzas_test.go`

## Runbooks relacionados

- `documentos/gobernanza_tecnica/runbooks/runbook_cierre_periodo_y_conciliacion_bancaria.md`

## ADRs relacionados

- `ADR-0001-frontera-multiempresa-empresa-id.md`
- `ADR-0002-postgresql-runtime-canonico-vps.md`