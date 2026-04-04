# Matriz KPI POS multiempresa

Fecha de actualizacion: 2026-04-03
Alcance: punto 1 del plan maestro (definicion funcional y KPI)

## KPI de ventas

| KPI | Formula | Frecuencia | Fuente esperada |
|---|---|---|---|
| ventas_diarias | suma(total_ventas_cerradas_dia) | diaria | carritos_compras, carrito_compra_items |
| ticket_promedio | ventas_diarias / numero_ventas_cerradas | diaria | carritos_compras |
| margen_bruto | (ventas - costo_ventas) / ventas | diaria/semanal | carritos_compras, inventario_movimientos |
| devoluciones_porcentaje | valor_devoluciones / ventas_brutas | diaria | carritos_compras |

## KPI de inventario

| KPI | Formula | Frecuencia | Fuente esperada |
|---|---|---|---|
| rotacion_inventario | costo_ventas_periodo / inventario_promedio | semanal/mensual | inventario_existencias, inventario_movimientos |
| dias_inventario | inventario_promedio / costo_ventas_diario | semanal/mensual | inventario_existencias, inventario_movimientos |
| quiebres_stock | conteo_skus_bajo_minimo | diaria | productos, inventario_existencias |
| precision_inventario | 1 - (diferencias_conteo / stock_teorico) | mensual | inventario_existencias, ajustes |

## KPI de clientes

| KPI | Formula | Frecuencia | Fuente esperada |
|---|---|---|---|
| clientes_activos_30d | conteo_clientes_con_compra_30d | diaria | clientes, carritos_compras |
| tasa_recompra_30d | clientes_con_2_mas_compras_30d / clientes_activos_30d | semanal | clientes, carritos_compras |
| valor_vida_cliente | ingreso_neto_cliente_historico / clientes_totales | mensual | clientes, carritos_compras |
| tiempo_promedio_recompra | promedio(dias_entre_compras_por_cliente) | mensual | carritos_compras |

## KPI de proveedores y compras

| KPI | Formula | Frecuencia | Fuente esperada |
|---|---|---|---|
| cumplimiento_entrega | ordenes_entregadas_a_tiempo / ordenes_totales | semanal/mensual | compras, recepciones |
| lead_time_promedio | promedio(fecha_recepcion - fecha_orden) | semanal/mensual | compras |
| variacion_costo_compra | (costo_actual - costo_anterior) / costo_anterior | semanal | compras, productos |
| pedidos_con_diferencia | recepciones_con_diferencias / recepciones_totales | semanal | recepciones |

## KPI contables y financieros

| KPI | Formula | Frecuencia | Fuente esperada |
|---|---|---|---|
| utilidad_operativa | ingresos_operativos - egresos_operativos | diaria/mensual | empresa_finanzas_movimientos |
| razon_corriente | activos_corrientes / pasivos_corrientes | mensual | asientos contables, balance |
| flujo_caja_neto | entradas_caja - salidas_caja | diaria/mensual | caja, empresa_finanzas_movimientos |
| indice_endudamiento | pasivos_totales / activos_totales | mensual | reportes contables |

## KPI de caja y operacion

| KPI | Formula | Frecuencia | Fuente esperada |
|---|---|---|---|
| diferencia_caja | caja_teorica - caja_fisica | por cierre | cierres_caja |
| cierres_con_incidencia | cierres_con_diferencia_mayor_umbral / cierres_totales | diaria | cierres_caja |
| tiempo_promedio_cierre | promedio(minutos_cierre) | diaria | cierres_caja |
| porcentaje_cierres_a_tiempo | cierres_en_horario / cierres_totales | diaria | cierres_caja |

## KPI de seguridad y permisos

| KPI | Formula | Frecuencia | Fuente esperada |
|---|---|---|---|
| acciones_denegadas | total_eventos_autorizacion_denegada | diaria | logs auditoria |
| acciones_criticas_auditadas | acciones_criticas_con_log / acciones_criticas_totales | diaria | logs auditoria |
| sesiones_activas_por_rol | conteo_sesiones_activas_por_rol | diaria | sesiones |
| intentos_fallidos_login | total_login_fallido | diaria | logs auth |

## Notas de implementacion

- Todos los KPI deben ser filtrables por empresa_id y sucursal_id.
- El tablero debe permitir corte por fecha desde/hasta y comparativo contra periodo anterior.
- Los KPI financieros deben cuadrar contra reportes contables oficiales del mismo periodo.
