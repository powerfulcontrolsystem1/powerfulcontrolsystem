# Logistica avanzada / WMS de bodega, picking, packing y despachos

Estado: Vigente. Responsable: Ingeniería del módulo. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- El código local incorpora transacciones y locks para órdenes, items, avances y despachos según el cierre de seguridad del 2026-09-05.
- WMS sigue siendo una capa operativa: el avance picking/packing no demuestra movimiento de existencias ni entrega física. La integración de inventario requiere un contrato y prueba específicos.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

Fecha: 2026-05-06

## Alcance

El modulo `logistica_wms` agrega una capa operativa profesional de bodega sobre el inventario existente. Permite gestionar ubicaciones internas, ordenes WMS, picking, packing, despachos, rutas, conteos y bitacora sin duplicar productos, bodegas, compras, produccion ni importaciones.

## Componentes

- API privada: `/api/empresa/logistica_wms`.
- Pantalla administrativa: `web/administrar_empresa/logistica_wms.html`.
- Menu principal: `linkLogisticaWMS`.
- Permiso/licencia: `logistica_wms`.
- Base de datos: `empresa_wms_ubicaciones`, `empresa_wms_ordenes`, `empresa_wms_items`, `empresa_wms_despachos`, `empresa_wms_eventos`.

## Funcionalidades

- Dashboard WMS con ubicaciones activas, ordenes abiertas, picking, packing, despachos en ruta y unidades pendientes.
- Maestro de ubicaciones por bodega, zona, pasillo, rack, nivel, posicion, tipo, capacidad, ocupacion y estado.
- Ordenes WMS para picking, packing, despacho, conteo, reabastecimiento, traslado y devolucion.
- Items de picking/packing con SKU, ubicacion origen/destino, lote, serial, cantidades solicitadas, pickeadas y empacadas.
- Registro de avance por item con inferencia de estado operativo.
- Despachos con transportadora, guia, conductor, vehiculo, ruta, fechas, costo de flete y estado.
- Bitacora de eventos por orden, item y despacho.
- Datos demo y exportacion CSV desde la pantalla.

## Integracion

El modulo se ubica dentro de Inventario y compras. La primera version opera como capa WMS independiente, preparada para conectar pedidos de venta, compras, produccion/MRP e importaciones por `origen_documento`, `producto_id`, SKU, lote, serial y ubicaciones. No modifica existencias automaticamente hasta que se definan las reglas de confirmacion de movimiento fisico por cada empresa.

## Gobierno y seguridad

Todas las tablas incluyen `empresa_id`. La API queda protegida por `WithEmpresaWMSPermissions` y por el techo de licencia `logistica_wms`. Roles operativos pueden leer; crear, actualizar y aprobar queda reservado a administradores, supervisores, inventario y compras segun la matriz empresarial.

## Pruebas

Se agregan pruebas unitarias en `backend/db/logistica_wms_test.go` para normalizacion de codigos, calculo de progreso de picking/packing e inferencia de estado de item.

## Fuentes y aceptación de la revisión

[logistica_wms.go](../backend/handlers/logistica_wms.go), [logistica_wms.go](../backend/db/logistica_wms.go), [logistica_wms_test.go](../backend/db/logistica_wms_test.go), [logistica_wms.html](../web/administrar_empresa/logistica_wms.html), [main.go](../backend/main.go), [empresa_permisos.go](../backend/handlers/empresa_permisos.go), [cierre_reparaciones_produccion_seguridad_2026-09-05.md](cierre_reparaciones_produccion_seguridad_2026-09-05.md).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).
