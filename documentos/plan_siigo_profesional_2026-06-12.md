# Plan profesional PCS tipo Siigo - 2026-06-12

## Lectura documental aplicada

Este plan se construye despues de revisar la documentacion viva del proyecto:
contexto Codex, mapa de modulos, flujos operativos, comandos, decisiones
tecnicas, estructura de base de datos, descripcion del proyecto, estructura del
codigo, matriz de roles y permisos, contratos de gobernanza, runbooks operativos
y documentacion funcional de contabilidad, compras, CRM/ventas, cartera,
tesoreria, activos, nomina, impuestos, portal contador, API y ERP multiempresa.

La conclusion tecnica es que PCS ya tiene muchas superficies de ERP, pero debe
cerrar mejor los contratos transversales: asiento contable automatico, estados
comerciales, cartera/CxP, bancos, reportes oficiales, migracion inicial, API y
documentos fiscales.

## Fase 1. Nucleo contable automatico

Estado: iniciado e implementado en primer corte.

- Mantener `empresa_eventos_contables` como bitacora auditable multiempresa.
- Generar `empresa_asientos_contables` solo para eventos monetarios reales.
- Bloquear asientos sin partida doble o con diferencia debito/credito.
- Soportar ventas, facturacion, compras, finanzas, cartera, creditos, inventario,
  nomina y activos fijos con cuentas configurables o fallback PUC.
- Registrar hitos precontables sin forzar asiento para no contaminar la cola.

## Fase 2. Ventas, facturas y cartera

Estado: siguiente corte recomendado.

- Unificar estados visibles: `sin_factura`, `facturada`, `pagada`, `parcial`,
  `anulada`, `con_nota`.
- Conectar cotizacion, pedido, remision, venta, factura electronica, recibo de
  caja, devolucion y nota credito.
- Formalizar abonos, anticipos, pagos parciales, saldos por tercero y vencimiento.
- Evitar doble contabilizacion entre abonos previos y cierre final de venta.

## Fase 3. Compras, proveedores y egresos

Estado: payload contable reforzado en primer corte.

- Formalizar orden de compra, recepcion, factura de compra, gasto, documento
  soporte, cuenta por pagar y egreso.
- Mantener aprobaciones y recepciones parciales como hitos auditables.
- Causar compras contabilizadas con base, IVA, retenciones y CxP.

## Fase 4. Bancos y conciliacion

Estado: existe base tecnica; falta madurez operativa.

- Reforzar importacion CSV/Excel/OFX con hash idempotente por empresa.
- Conciliar automaticamente por valor, fecha, referencia y tercero.
- Preparar conectores Bre-B, Wompi, ePayco y bancos como integraciones
  auditadas, no como acoples directos al asiento.

## Fase 5. Reportes contables oficiales

Estado: existe infraestructura de asientos/reportes; falta salida contable fuerte.

- Balance de prueba, estado de situacion financiera, estado de resultados, flujo
  de efectivo, cambios en patrimonio y auxiliares por tercero.
- Filtros por periodo, tercero, centro de costo, bodega, caja y empresa.
- Exportacion JSON/CSV/TXT/XLS/PDF segun contratos vigentes.

## Fase 6. Migracion inicial guiada

Estado: pendiente.

- Asistente por empresa para plan de cuentas, terceros, productos, inventario
  inicial, cartera, proveedores, bancos, activos, nomina y numeracion DIAN.
- Plantillas Excel validadas antes de persistir.
- Eventos contables de saldos iniciales balanceados antes de activar empresa.

## Fase 7. API empresarial, documentos fiscales e IA operativa

Estado: pendiente.

- Portal API por empresa con llaves, permisos, OpenAPI, idempotencia, limites,
  webhooks y auditoria.
- Madurar documento soporte, nomina electronica, RADIAN, eventos DIAN,
  contingencia, acuses, estados y trazabilidad.
- Conectar IA a acciones reales: consultar datos por empresa, preparar registros,
  sugerir causaciones, explicar reportes y detectar errores fiscales.

## Primer corte implementado

- Motor contable con partida doble estricta.
- Eventos precontables auditables sin asiento obligatorio.
- Payloads fiscales reforzados en facturacion, compras, finanzas y ventas.
- Pruebas enfocadas del motor contable y compilacion de handlers.
