# Matriz KPI POS multiempresa

Fecha de actualizacion: 2026-04-04
Alcance: punto 1 del plan maestro (matriz formal de KPI con formula, endpoint y tabla fuente)

## Fuente canonica de consulta

- Endpoint de lectura: `GET /api/empresa/finanzas/movimientos?action=tablero|dashboard|resumen_kpi&empresa_id={id}&desde={yyyy-mm-dd}&hasta={yyyy-mm-dd}`.
- Endpoint de exportacion: `GET /api/empresa/finanzas/movimientos?action=tablero_export&format=json|csv&empresa_id={id}&desde={yyyy-mm-dd}&hasta={yyyy-mm-dd}`.
- Funcion DB fuente: `GetEmpresaReportesTableroResumen` en `backend/db/finanzas.go`.

## KPI operativos

| KPI | Formula implementada | Endpoint de lectura | Tablas fuente |
|---|---|---|---|
| ventas_cerradas | `COUNT(*)` de `carritos_compras` con `estado_carrito='cerrado'` y `estado='activo'` en rango | `/api/empresa/finanzas/movimientos?action=tablero` | `carritos_compras` |
| ventas_hoy | `SUM(CASE WHEN date(fecha_pago_o_actualizacion)=date('now','localtime') THEN 1 END)` sobre ventas cerradas activas | `/api/empresa/finanzas/movimientos?action=tablero` | `carritos_compras` |
| ingresos_ventas | `SUM(CASE WHEN total_pagado>0 THEN total_pagado ELSE total END)` sobre ventas cerradas activas en rango | `/api/empresa/finanzas/movimientos?action=tablero` | `carritos_compras` |
| ticket_promedio | `ingresos_ventas / ventas_cerradas` (si `ventas_cerradas > 0`) | `/api/empresa/finanzas/movimientos?action=tablero` | derivado de `carritos_compras` |
| clientes_activos | `COUNT(*)` de clientes con `estado='activo'` | `/api/empresa/finanzas/movimientos?action=tablero` | `clientes` |
| productos_activos | `COUNT(*)` de productos con `estado='activo'` | `/api/empresa/finanzas/movimientos?action=tablero` | `productos` |
| productos_bajo_minimo | `SUM(CASE WHEN stock_total <= stock_minimo THEN 1 END)` con `stock_total` agregado desde existencias activas | `/api/empresa/finanzas/movimientos?action=tablero` | `productos`, `inventario_existencias` |
| compras_movimientos | `COUNT(*)` de movimientos inventario activos tipo `entrada/ajuste_entrada/ajuste_positivo/compra` en rango | `/api/empresa/finanzas/movimientos?action=tablero` | `inventario_movimientos` |
| compras_costo | `SUM(cantidad * costo_unitario)` para los movimientos de compra/entrada en rango | `/api/empresa/finanzas/movimientos?action=tablero` | `inventario_movimientos` |

## KPI financieros

| KPI | Formula implementada | Endpoint de lectura | Tablas fuente |
|---|---|---|---|
| movimientos_ingresos | `SUM(CASE WHEN tipo_movimiento='ingreso' THEN 1 END)` en movimientos activos por rango | `/api/empresa/finanzas/movimientos?action=tablero` | `empresa_finanzas_movimientos` |
| movimientos_egresos | `SUM(CASE WHEN tipo_movimiento='egreso' THEN 1 END)` en movimientos activos por rango | `/api/empresa/finanzas/movimientos?action=tablero` | `empresa_finanzas_movimientos` |
| ingresos | `SUM(CASE ingreso THEN COALESCE(NULLIF(total_neto,0), NULLIF(total,0), monto,0) END)` | `/api/empresa/finanzas/movimientos?action=tablero` | `empresa_finanzas_movimientos` |
| egresos | `SUM(CASE egreso THEN COALESCE(NULLIF(total_neto,0), NULLIF(total,0), monto,0) END)` | `/api/empresa/finanzas/movimientos?action=tablero` | `empresa_finanzas_movimientos` |
| balance | `ingresos - egresos` | `/api/empresa/finanzas/movimientos?action=tablero` | derivado de `empresa_finanzas_movimientos` |
| periodos_abiertos | `SUM(CASE WHEN estado='abierto' THEN 1 END)` excluyendo `estado='inactivo'` | `/api/empresa/finanzas/movimientos?action=tablero` | `empresa_finanzas_periodos` |
| periodos_cerrados | `SUM(CASE WHEN estado='cerrado' THEN 1 END)` excluyendo `estado='inactivo'` | `/api/empresa/finanzas/movimientos?action=tablero` | `empresa_finanzas_periodos` |

## KPI contables

| KPI | Formula implementada | Endpoint de lectura | Tablas fuente |
|---|---|---|---|
| eventos_pendientes | `SUM(CASE WHEN procesado=0 THEN 1 END)` en eventos activos por rango | `/api/empresa/finanzas/movimientos?action=tablero` | `empresa_eventos_contables` |
| eventos_procesados | `SUM(CASE WHEN procesado=1 THEN 1 END)` en eventos activos por rango | `/api/empresa/finanzas/movimientos?action=tablero` | `empresa_eventos_contables` |
| eventos_total | `COUNT(*)` de eventos contables activos por rango | `/api/empresa/finanzas/movimientos?action=tablero` | `empresa_eventos_contables` |
| eventos_monto_total | `SUM(monto_total)` de eventos contables activos por rango | `/api/empresa/finanzas/movimientos?action=tablero` | `empresa_eventos_contables` |
| asientos_generados | `COUNT(*)` de asientos activos en rango | `/api/empresa/finanzas/movimientos?action=tablero` | `empresa_asientos_contables` |
| asientos_monto_total | `SUM(max(total_debito,total_credito))` por asiento activo en rango | `/api/empresa/finanzas/movimientos?action=tablero` | `empresa_asientos_contables` |
| documentos_facturacion_activos | `COUNT(*)` con `estado='activo'` | `/api/empresa/finanzas/movimientos?action=tablero` | `empresa_facturacion_documentos` |
| documentos_compras_activos | `COUNT(*)` con `estado='activo'` | `/api/empresa/finanzas/movimientos?action=tablero` | `empresa_compras_documentos` |

## Estado de resultados y balance general

| KPI | Formula implementada | Endpoint de lectura | Tablas fuente |
|---|---|---|---|
| estado_resultados.ingresos | suma de lineas contables clase `4` (`ingresos += -delta`) | `/api/empresa/finanzas/movimientos?action=tablero` | `empresa_asientos_contables` (`lineas_json`) |
| estado_resultados.gastos | suma de lineas clase `5/6/7` (`gastos += delta`) | `/api/empresa/finanzas/movimientos?action=tablero` | `empresa_asientos_contables` (`lineas_json`) |
| estado_resultados.utilidad_operacional | `ingresos - gastos` | `/api/empresa/finanzas/movimientos?action=tablero` | derivado de asientos |
| balance_general.activos | suma de lineas clase `1` (`activos += delta`) | `/api/empresa/finanzas/movimientos?action=tablero` | `empresa_asientos_contables` (`lineas_json`) |
| balance_general.pasivos | suma de lineas clase `2` (`pasivos += -delta`) | `/api/empresa/finanzas/movimientos?action=tablero` | `empresa_asientos_contables` (`lineas_json`) |
| balance_general.patrimonio | `patrimonio_base + utilidad` con `patrimonio_base` en clase `3` (`patrimonio += -delta`) | `/api/empresa/finanzas/movimientos?action=tablero` | `empresa_asientos_contables` (`lineas_json`) |
| balance_general.resultado_ejercicio | `utilidad_operacional` | `/api/empresa/finanzas/movimientos?action=tablero` | derivado de asientos |
| balance_general.cuadre | `activos - (pasivos + patrimonio)` | `/api/empresa/finanzas/movimientos?action=tablero` | derivado de asientos |

## Regla de fallback contable (implementada)

- Si `asientos_generados = 0` en el rango solicitado:
	- `estado_resultados.ingresos = financiero.ingresos`.
	- `estado_resultados.gastos = financiero.egresos`.
	- `estado_resultados.utilidad_operacional = ingresos - egresos`.
	- El balance se construye con la utilidad como aproximacion transitoria hasta tener asientos canonicos.

## Notas operativas

- Todas las metricas del tablero se calculan por `empresa_id` y aceptan rango `desde/hasta`.
- La exportacion `tablero_export` usa exactamente la misma fuente que `action=tablero` para evitar desviaciones entre UI y descarga.
