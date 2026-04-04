# Plan maestro POS multiempresa (14 puntos)

Fecha de actualizacion: 2026-04-03
Estado global: en ejecucion

## Objetivo
Implementar y consolidar un sistema POS multiempresa con contabilidad integrada, trazabilidad completa por empresa/sucursal y control de acceso por roles.

## Estado por punto

| Punto | Modulo | Estado | Entregable principal |
|---|---|---|---|
| 1 | Alcance funcional y KPI | en curso | matriz de procesos y KPI operativos/financieros |
| 2 | Arquitectura multiempresa | en curso | lineamientos de aislamiento por empresa/sucursal |
| 3 | Permisos y seguridad | en curso | matriz de roles/permisos por empresa/sucursal |
| 4 | Gestion de ventas | pendiente | flujo de venta/factura/descuento/inventario |
| 5 | Control de inventarios | pendiente | stock, alertas y movimientos de bodega |
| 6 | Gestion de clientes | pendiente | perfil, historial y segmentacion |
| 7 | Gestion de proveedores | pendiente | catalogo, precios y condiciones |
| 8 | Modulo de facturacion electronica | pendiente | emision legal y cumplimiento normativo |
| 9 | Modulo de compras | pendiente | orden, recepcion y contabilizacion |
| 10 | Modulo contable integrado | pendiente | asientos automaticos por evento |
| 11 | Reportes financieros | pendiente | balance, estado de resultados, flujo de caja |
| 12 | Cierres de caja | pendiente | arqueo y cierre por sucursal/empresa |
| 13 | Calidad, UAT y despliegue | pendiente | validacion integral y salida controlada |
| 14 | Operacion continua | pendiente | mejora continua con KPI y roadmap trimestral |

## Continuacion ejecutada ahora

### Punto 1. Alcance funcional y KPI (avance)

Se define alcance minimo obligatorio para cada modulo:
- Ventas: registro, descuentos, impuestos, devoluciones, factura y salida de inventario.
- Inventarios: existencia por bodega, stock minimo/maximo, alertas y kardex.
- Clientes: datos base, historial, frecuencia de compra y segmentacion.
- Proveedores: condiciones comerciales, precios, plazos y cumplimiento.
- Facturacion electronica: emision, estado del documento, reintentos y auditoria.
- Compras: solicitud, orden, recepcion, diferencias y costo final.
- Contabilidad: asientos por evento, integridad de periodos y trazabilidad por documento.
- Reportes financieros: balance general, estado de resultados y flujo de caja.
- Cierre de caja: apertura/cierre por sucursal, arqueo, diferencias y aprobacion.
- Permisos: rol por empresa/sucursal/usuario con principio de minimo privilegio.

KPI iniciales del sistema:
- Ventas: ventas_diarias, ticket_promedio, margen_bruto.
- Inventario: rotacion, dias_inventario, quiebres_stock.
- Clientes: recompra_30d, cliente_activo, valor_vida_cliente.
- Compras/proveedores: cumplimiento_entrega, variacion_costo, lead_time_promedio.
- Contabilidad/finanzas: utilidad_operativa, razon_corriente, flujo_caja_neto.
- Caja: diferencia_caja, tiempo_cierre, cierres_con_incidencia.
- Seguridad: eventos_denegados_por_rol, acciones_criticas_auditadas.

### Punto 2. Arquitectura multiempresa (avance)

Reglas tecnicas de aislamiento:
- Toda tabla transaccional debe incluir empresa_id; cuando aplique operacion fisica, incluir sucursal_id y bodega_id.
- Toda operacion API debe validar alcance por empresa/sucursal antes de consultar o mutar datos.
- Todo log funcional debe incluir request_id, empresa_id y usuario.
- Todo documento financiero/fiscal debe poder rastrearse hasta su transaccion origen.

Reglas de integridad contable:
- Cada evento de negocio debe mapear a un asiento contable verificable.
- Ningun cierre de periodo debe permitir mutaciones posteriores sin reapertura autorizada.
- Toda anulacion debe conservar rastro y contrapartida contable.

### Punto 3. Permisos y seguridad (inicio)

Entregable generado:
- `documentos/matriz_roles_permisos_pos_multiempresa.md` con:
	- roles base por alcance,
	- permisos por modulo (C/R/U/D/A),
	- reglas obligatorias de autorizacion por empresa/sucursal,
	- siguientes acciones tecnicas para llevar la matriz a middleware y pruebas.

Implementacion tecnica inicial completada:
- Se crea `backend/handlers/empresa_permisos.go` con middleware de autorizacion por rol y alcance de empresa.
- Se aplica el middleware en rutas criticas de:
	- ventas (`/api/empresa/carritos_compra`, `/api/empresa/carritos_compra/items`),
	- inventario (`/api/empresa/bodegas`, `categorias_productos`, `productos`, `inventario/*`, `productos/precios_historial`),
	- finanzas (`/api/empresa/finanzas/movimientos`, `configuracion`, `periodos`).
- Se agregan pruebas de permiso/denegacion en `backend/handlers/empresa_permisos_test.go`.

### Punto 3. Permisos y seguridad (continuacion ejecutada)

Implementacion tecnica completada en Chat IA empresarial:
- El modulo `chat_con_inteligencia_artificial` queda en modo Gemini-only para reducir superficie operativa y simplificar control de credenciales.
- Se mantiene autenticacion obligatoria por cuenta Google administradora y validacion de alcance por `empresa_id` en todos los endpoints IA (`modelos`, `modelo_preferido`, `consultar`, `historial`).
- Se conserva registro automatico de cuenta Google administradora en callback OAuth para primer acceso valido.
- Se rediseña la UI del chat para hacer visible el alcance por empresa, la cuenta Google activa y el estado de sesion/autenticacion.
- Se mantiene trazabilidad de uso y auditoria por empresa/modelo/cuenta administradora en tablas de IA.

## Backlog inmediato (siguiente iteracion)

1. Extender la cobertura del middleware de permisos a rutas operativas restantes (punto 3).
2. Definir contrato de eventos contables por modulo (puntos 4, 8, 9, 10).
3. Estandarizar estados de ciclo de vida para ventas, compras y facturas.
4. Definir tablero minimo de reportes financieros y operativos.
5. Definir flujo operativo de cierre de caja por sucursal.

## Criterios de avance para la siguiente fase

- Punto 1 queda en estado completo cuando exista una matriz formal de KPI con formulas y fuente de datos por endpoint/tabla.
- Punto 2 queda en estado completo cuando exista matriz de entidades con llaves de aislamiento (empresa/sucursal/bodega) y reglas de validacion por endpoint.
- Punto 3 iniciara con entregable de roles base: super_admin, admin_empresa, cajero, inventario, contabilidad, auditor.
