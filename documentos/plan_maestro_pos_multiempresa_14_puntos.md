# Plan maestro POS multiempresa (14 puntos)

Fecha de actualizacion: 2026-04-04
Estado global: en ejecucion

## Objetivo
Implementar y consolidar un sistema POS multiempresa con contabilidad integrada, trazabilidad completa por empresa/sucursal y control de acceso por roles.

## Estado por punto

| Punto | Modulo | Estado | Entregable principal |
|---|---|---|---|
| 1 | Alcance funcional y KPI | en curso | matriz de procesos y KPI operativos/financieros |
| 2 | Arquitectura multiempresa | en curso | lineamientos de aislamiento por empresa/sucursal |
| 3 | Permisos y seguridad | en curso | matriz de roles/permisos por empresa/sucursal |
| 4 | Gestion de ventas | en curso | flujo de venta/factura/descuento/inventario |
| 5 | Control de inventarios | pendiente | stock, alertas y movimientos de bodega |
| 6 | Gestion de clientes | pendiente | perfil, historial y segmentacion |
| 7 | Gestion de proveedores | pendiente | catalogo, precios y condiciones |
| 8 | Modulo de facturacion electronica | pendiente | emision legal y cumplimiento normativo |
| 9 | Modulo de compras | pendiente | orden, recepcion y contabilizacion |
| 10 | Modulo contable integrado | pendiente | asientos automaticos por evento |
| 11 | Reportes financieros | en curso | balance, estado de resultados, flujo de caja |
| 12 | Cierres de caja | en curso | arqueo y cierre por sucursal/empresa |
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

### Punto 3. Permisos y seguridad (avance adicional 2026-04-04)

Implementacion tecnica adicional completada:
- Se amplía el middleware de autorizacion por rol/empresa en `backend/handlers/empresa_permisos.go` con nuevos modulos:
	- `clientes`,
	- `compras` (aplicado a proveedores),
	- `facturacion`.
- Se agregan wrappers dedicados:
	- `WithEmpresaClientesPermissions`,
	- `WithEmpresaComprasPermissions`,
	- `WithEmpresaFacturacionPermissions`.
- Se extiende cobertura de rutas en `backend/main.go`:
	- `clientes`: `/api/empresa/clientes`,
	- `compras/proveedores`: `/api/empresa/proveedores`,
	- `facturacion`: `/api/empresa/facturacion_electronica`, `/api/empresa/facturacion_electronica/pais_detectado`.
- Se incorpora control para `servicios` bajo politica de inventario:
	- `/api/empresa/servicios`.
- Se agregan pruebas de autorizacion por rol para nuevos modulos en `backend/handlers/empresa_permisos_test.go`.

Validacion ejecutada:
- `go test ./handlers -run "WithEmpresa|ConsultarHandlerRejectsEmpresaFueraDeAlcance" -count=1` (ok).
- `go test ./...` en `backend` (ok).

### Punto 3. Permisos y seguridad (avance adicional 2026-04-04 - cierre de rutas pendientes)

Implementacion tecnica adicional completada:
- Se agrega modulo `seguridad` en `backend/handlers/empresa_permisos.go` con wrapper:
	- `WithEmpresaSeguridadPermissions`.
- Se amplian rutas protegidas en `backend/main.go`:
	- seguridad: `/api/empresa/usuarios`, `/api/empresa/configuracion_avanzada`, `/api/empresa/roles_de_usuario`.
	- inventario: `/api/empresa/productos/imagen`, `/api/empresa/ubicacion_gps/dispositivos`, `/api/empresa/ubicacion_gps/recorridos`.
	- colaboracion operativa (politica ventas):
		- `/api/empresa/chat_tareas/conversaciones`,
		- `/api/empresa/chat_tareas/participantes`,
		- `/api/empresa/chat_tareas/mensajes`,
		- `/api/empresa/chat_tareas/mensajes/adjunto`,
		- `/api/empresa/chat_tareas/tareas`.
- Se agregan pruebas del modulo seguridad en `backend/handlers/empresa_permisos_test.go`.

Validacion ejecutada:
- `go test ./handlers -run "WithEmpresa|ConsultarHandlerRejectsEmpresaFueraDeAlcance" -count=1` (ok).
- `go test ./...` en `backend` (ok).

### Punto 3. Permisos y seguridad (validacion de cierre en endpoints protegidos)

Validacion tecnica adicional completada:
- Se agregan pruebas de middleware para rutas protegidas recientemente incorporadas:
	- `TestWithEmpresaInventarioPermissionsDeniesCajeroWriteGPS` para `/api/empresa/ubicacion_gps/dispositivos`.
	- `TestWithEmpresaVentasPermissionsAllowsCajeroChatAdjuntoMultipart` para `/api/empresa/chat_tareas/mensajes/adjunto` con `multipart/form-data`.
	- `TestWithEmpresaVentasPermissionsRejectsChatAdjuntoWithoutAuth` para validar `401` sin autenticacion.
- Se confirma extraccion de `empresa_id` desde payload multipart y validacion de cabecera `X-Empresa-ID` en respuesta del middleware.

Validacion ejecutada:
- `runTests` sobre `backend/handlers/empresa_permisos_test.go` (ok).
- `go test ./...` en `backend` (ok).

### Punto 4. Gestion de ventas (inicio 2026-04-04)

Implementacion tecnica inicial completada:
- Se estandariza el ciclo de vida de venta en carritos con nuevo campo de salida `estado_venta`.
- `backend/db/carritos_compras.go` ahora calcula `estado_venta` en lectura (`GetCarritosCompraByEmpresa`, `GetCarritoCompraByID`) con estados:
	- `venta_abierta`,
	- `venta_cerrada`,
	- `venta_pagada`,
	- `venta_suspendida`.
- `backend/handlers/carritos_compras.go` devuelve `estado_venta` en acciones operativas:
	- `activar_estacion`,
	- `pagar_estacion`,
	- `activar/desactivar`,
	- `cerrar/reabrir`.
- Se agrega cobertura de pruebas para ciclo de vida en:
	- `backend/handlers/auth_users_carritos_test.go`.
	- `backend/db/carritos_inventario_test.go`.

Validacion ejecutada:
- `runTests` sobre `backend/handlers/auth_users_carritos_test.go` y `backend/db/carritos_inventario_test.go` (ok).
- `go test ./...` en `backend` (ok).

### Punto 4. Gestion de ventas (avance adicional 2026-04-04 - transiciones permitidas)

Implementacion tecnica adicional completada:
- Se formalizan transiciones de ciclo de venta en `backend/handlers/carritos_compras.go` con validaciones de negocio por accion:
	- `pagar_estacion`: bloquea doble pago y exige venta activa.
	- `cerrar/reabrir`: bloquea reabrir ventas pagadas y evita transiciones incoherentes.
	- `activar/desactivar`: evita activar ventas ya activas y bloquear activacion directa de ventas pagadas.
	- `activar_estacion`: para ventas pagadas exige `reset_items=1` para iniciar nueva sesion.
- Se agregan respuestas HTTP de control de estado:
	- `404` cuando el carrito no existe.
	- `409` cuando la transicion solicitada no es valida para el estado actual.
- Se amplian pruebas en `backend/handlers/auth_users_carritos_test.go` con escenarios de conflicto:
	- `TestEmpresaCarritosCompraRejectsDoublePago`.
	- `TestEmpresaCarritosCompraRejectsReabrirVentaPagada`.
	- `TestEmpresaCarritosCompraRejectsActivarEstacionPagadaSinReset`.

Validacion ejecutada:
- `go test ./handlers -run "Carritos|EmpresaCarritosCompra" -count=1` (ok).
- `go test ./...` en `backend` (ok).

### Punto 4 + Punto 10. Gestion de ventas con contrato de eventos contables (avance 2026-04-04)

Implementacion tecnica adicional completada:
- Se define contrato base de eventos contables por modulo en `backend/db/eventos_contables.go` para:
	- `ventas`,
	- `facturacion`,
	- `compras`,
	- `finanzas`.
- Se crea tabla empresarial `empresa_eventos_contables` con trazabilidad para integracion contable:
	- modulo, evento, entidad, documento, periodo contable, monto, payload JSON y estado de procesamiento.
- Se agrega registro operativo del contrato en flujo de ventas (carritos) desde `backend/handlers/carritos_compras.go` para eventos:
	- `venta_sesion_activada`,
	- `venta_activada`,
	- `venta_suspendida`,
	- `venta_cerrada`,
	- `venta_reabierta`,
	- `venta_pagada`.
- Se integra esquema en bootstrap de servidor (`backend/main.go`):
	- `EnsureEmpresaEventosContablesSchema`.
	- migracion `2026-04-04-007-eventos-contables`.

Validacion ejecutada:
- `go test ./db -run "EventosContables|CarritoEstadoVentaLifecycle|Finanzas" -count=1` (ok).
- `go test ./handlers -run "EmpresaCarritosCompra|CarritosCompraAndItemsFlow" -count=1` (ok).
- `go test ./...` en `backend` (ok).

### Punto 8 + Punto 9 + Punto 10. Extension de eventos contables a facturacion/compras/finanzas (avance 2026-04-04)

Implementacion tecnica adicional completada:
- Se extiende el contrato contable en `backend/db/eventos_contables.go` para soportar eventos operativos actuales de:
	- `facturacion`: `configuracion_facturacion_actualizada`.
	- `compras`: `proveedor_registrado`, `proveedor_actualizado`, `proveedor_activado`, `proveedor_desactivado`, `proveedor_eliminado`.
	- `finanzas`: `movimiento_ingreso_registrado`, `movimiento_egreso_registrado`, `periodo_contable_cerrado`, `periodo_contable_reabierto`.
- Se agrega helper reutilizable no bloqueante `backend/handlers/eventos_contables.go` para centralizar serializacion de payload, normalizacion y registro seguro de eventos.
- Se integra emision en handlers por modulo:
	- `backend/handlers/facturacion_electronica.go`: emite `configuracion_facturacion_actualizada` tras guardar configuracion FE por pais.
	- `backend/handlers/productos.go` (proveedores/compras): emite eventos en alta, actualizacion, activacion/desactivacion y eliminacion de proveedor.
	- `backend/handlers/finanzas.go`: emite eventos al crear movimientos (`ingreso`/`egreso`) y al cerrar/reabrir periodos contables.
- Se mantiene emision de ventas en `backend/handlers/carritos_compras.go` ahora usando helper comun.
- Se agrega cobertura de pruebas en `backend/handlers/eventos_contables_modulos_test.go` para validar emision en facturacion, compras y finanzas.

Validacion ejecutada:
- `go test ./db -run "EventosContables" -count=1` (ok).
- `go test ./handlers -run "FacturacionElectronicaEmiteEventoContable|ProveedoresEmiteEventoContableCompras|FinanzasEmiteEventosContables|CarritosCompraAndItemsFlow" -count=1` (ok).
- `go test ./...` en `backend` (ok).

### Punto 8 + Punto 9 + Punto 10. Eventos transaccionales de factura/orden (avance 2026-04-04)

Implementacion tecnica adicional completada:
- `backend/handlers/facturacion_electronica.go` incorpora acciones transaccionales via `action`:
	- `emitir` -> evento `factura_emitida`.
	- `anular` -> evento `factura_anulada`.
	- `nota_credito` / `emitir_nota_credito` -> evento `nota_credito_emitida`.
- `backend/handlers/productos.go` (`EmpresaProveedoresHandler`) incorpora acciones transaccionales via `action`:
	- `emitir` / `emitir_orden` -> evento `orden_compra_emitida`.
	- `recepcionar` / `recepcionar_compra` -> evento `compra_recepcionada`.
	- `contabilizar` / `contabilizar_compra` -> evento `compra_contabilizada`.
- `backend/handlers/empresa_permisos.go` actualiza resolucion de acciones para compras/facturacion y clasifica correctamente operaciones de aprobacion/anulacion.
- `backend/handlers/eventos_contables_modulos_test.go` agrega pruebas especificas de eventos transaccionales para factura y orden de compra.

Validacion ejecutada:
- `go test ./handlers -run "FacturacionTransaccionalEmiteEventosContables|ComprasTransaccionalEmiteEventosContables|FacturacionElectronicaEmiteEventoContable|ProveedoresEmiteEventoContableCompras|FinanzasEmiteEventosContables" -count=1` (ok).
- `go test ./...` en `backend` (ok).

### Punto 8 + Punto 9 + Punto 10. Estandarizacion de estados en ciclo documental (avance 2026-04-04)

Implementacion tecnica adicional completada:
- Se agrega `backend/handlers/documentos_lifecycle.go` con reglas de transicion para:
	- facturacion: `emitir`, `anular`, `nota_credito`.
	- compras: `emitir_orden`, `recepcionar_compra`, `contabilizar_compra`.
- `backend/handlers/facturacion_electronica.go` valida `estado_actual` antes de emitir eventos transaccionales:
	- retorna `409` cuando la transicion no es valida,
	- responde `estado_anterior` y `estado_nuevo` en operaciones exitosas,
	- persiste dichos estados en el `payload_json` del evento contable.
- `backend/handlers/productos.go` (`EmpresaProveedoresHandler`) aplica la misma validacion de ciclo para compras.
- Se amplian pruebas en `backend/handlers/eventos_contables_modulos_test.go`:
	- `TestEmpresaFacturacionTransaccionalRechazaTransicionInvalida`.
	- `TestEmpresaComprasTransaccionalRechazaTransicionInvalida`.

Validacion ejecutada:
- `go test ./handlers -run "FacturacionTransaccional|ComprasTransaccional|FacturacionElectronicaEmiteEventoContable|ProveedoresEmiteEventoContableCompras|FinanzasEmiteEventosContables" -count=1` (ok).
- `go test ./...` en `backend` (ok).

### Punto 8 + Punto 9 + Punto 10. Persistencia formal de documentos de factura/orden (avance 2026-04-04)

Implementacion tecnica adicional completada:
- Se agrega `backend/db/documentos_transaccionales.go` para persistencia canonica de documentos de negocio en tablas dedicadas:
	- `empresa_facturacion_documentos`.
	- `empresa_compras_documentos`.
- Se integra en `backend/main.go`:
	- `EnsureEmpresaDocumentosTransaccionalesSchema`.
	- migracion `2026-04-04-008-documentos-transaccionales`.
- `backend/handlers/facturacion_electronica.go` y `backend/handlers/productos.go` ahora:
	- consultan estado desde documento persistido por `documento_codigo`,
	- aplican/guardan transicion sobre el documento canonico,
	- emiten eventos en `empresa_eventos_contables` con `entidad_id` canonico (ID persistido del documento).
- Se agrega cobertura en `backend/db/documentos_transaccionales_test.go` y se amplian verificaciones en `backend/handlers/eventos_contables_modulos_test.go` para asegurar estabilidad de `entidad_id` durante el ciclo documental.

Validacion ejecutada:
- `go test ./handlers -run "FacturacionTransaccionalEmiteEventosContables|ComprasTransaccionalEmiteEventosContables|FacturacionElectronicaEmiteEventoContable|ProveedoresEmiteEventoContableCompras|FinanzasEmiteEventosContables" -count=1` (ok).
- `go test ./auth ./db ./handlers ./metrics ./utils` (ok).
- `go test ./...` en `backend` (ok).

### Punto 11. Reportes financieros (inicio 2026-04-04 - tablero minimo financiero-operativo)

Implementacion tecnica adicional completada:
- `backend/db/finanzas.go` incorpora resumen consolidado de KPI con `GetEmpresaReportesTableroResumen` para tablero minimo por empresa:
	- bloque `operativo`: ventas cerradas/hoy, ingresos ventas, ticket promedio, clientes activos, productos activos, productos bajo minimo, compras por movimientos/costo.
	- bloque `financiero`: ingresos, egresos, balance, movimientos por tipo, periodos abiertos/cerrados.
	- bloque `contable`: eventos pendientes/procesados/total, monto de eventos, documentos activos de facturacion y compras.
	- soporte de filtros de fecha (`desde`, `hasta`) para rangos operativos y financieros.
- `backend/handlers/finanzas.go` extiende `GET /api/empresa/finanzas/movimientos` con `action=tablero|dashboard|resumen_kpi` para exponer el tablero en API.
- `web/administrar_empresa/reportes.html` integra una segunda franja de KPI financieros y contables consumiendo el endpoint anterior, manteniendo fallback `N/D` cuando la API no esta disponible para el rol.

Validacion ejecutada:
- `go test ./db -run "TestGetEmpresaReportesTableroResumen|TestEmpresaFinanzas" -count=1` (ok).
- `go test ./handlers -run "TestEmpresaFinanzasTableroResumenHandler|TestEmpresaFinanzasEmiteEventosContables" -count=1` (ok).
- `go test ./... -count=1` en `backend` (ok).

### Punto 12. Cierres de caja (inicio 2026-04-04 - flujo operativo por sucursal)

Implementacion tecnica adicional completada:
- `backend/db/finanzas.go` incorpora tabla y dominio `empresa_cierres_caja` con flujo de:
	- apertura de caja,
	- arqueo con `caja_fisica`,
	- cierre con calculo de `caja_teorica` y `diferencia_caja`,
	- aprobacion/reapertura/anulacion con reglas de transicion.
- `backend/handlers/finanzas.go` agrega endpoint `GET/POST/PUT/DELETE /api/empresa/finanzas/cierres_caja` con acciones:
	- `action=cerrar`,
	- `action=reabrir`,
	- `action=aprobar`,
	- `action=anular`,
	- `action=activar|desactivar`.
- `backend/handlers/empresa_permisos.go` clasifica `action=aprobar` en finanzas como accion de aprobacion (`A`).
- `backend/main.go` publica la ruta `"/api/empresa/finanzas/cierres_caja"` y registra migracion `2026-04-04-009-cierres-caja`.
- Cobertura de pruebas:
	- `backend/db/finanzas_test.go`: `TestEmpresaCierresCajaFlow`.
	- `backend/handlers/eventos_contables_modulos_test.go`: `TestEmpresaFinanzasCierresCajaHandler`.

Validacion ejecutada:
- `go test ./db -run "TestEmpresaCierresCajaFlow|TestGetEmpresaReportesTableroResumen|TestEmpresaFinanzas" -count=1` (ok).
- `go test ./handlers -run "TestEmpresaFinanzasCierresCajaHandler|TestEmpresaFinanzasTableroResumenHandler|TestEmpresaFinanzasEmiteEventosContables" -count=1` (ok).
- `go test ./auth ./db ./handlers ./metrics ./utils -count=1` (ok).

## Backlog inmediato (siguiente iteracion)

1. Integrar UI operativa de cierres de caja (apertura/arqueo/cierre/aprobacion) en panel empresa.
2. Definir estrategia de procesamiento de asientos (consumo de `empresa_eventos_contables` y documentos canonicos).
3. Integrar estado de resultados/balance general del modulo finanzas con el tablero minimo en reportes.

## Criterios de avance para la siguiente fase

- Punto 1 queda en estado completo cuando exista una matriz formal de KPI con formulas y fuente de datos por endpoint/tabla.
- Punto 2 queda en estado completo cuando exista matriz de entidades con llaves de aislamiento (empresa/sucursal/bodega) y reglas de validacion por endpoint.
- Punto 3 queda en estado completo cuando exista proceso documentado y probado para convertir eventos en asientos con referencia canonica de documento (`entidad_id`).
