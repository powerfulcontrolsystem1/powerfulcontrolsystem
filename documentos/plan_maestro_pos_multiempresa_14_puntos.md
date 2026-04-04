# Plan maestro POS multiempresa (15 puntos)

Fecha de actualizacion: 2026-04-04
Estado global: en ejecucion

## Objetivo
Implementar y consolidar un sistema POS multiempresa con contabilidad integrada, trazabilidad completa por empresa/sucursal y control de acceso por roles.

## Estado por punto

| Punto | Modulo | Estado | Entregable principal |
|---|---|---|---|
| 1 | Alcance funcional y KPI | completado | matriz KPI formal con formula, endpoint y tablas fuente |
| 2 | Arquitectura multiempresa | completado | matriz de entidades y llaves de aislamiento por endpoint |
| 3 | Permisos y seguridad | en curso | matriz de roles/permisos por empresa/sucursal |
| 4 | Gestion de ventas | en curso | flujo de venta/factura/descuento/inventario |
| 5 | Control de inventarios | pendiente | stock, alertas y movimientos de bodega |
| 6 | Gestion de clientes | pendiente | perfil, historial y segmentacion |
| 7 | Gestion de proveedores | pendiente | catalogo, precios y condiciones |
| 8 | Modulo de facturacion electronica | pendiente | emision legal y cumplimiento normativo |
| 9 | Modulo de compras | pendiente | orden, recepcion y contabilizacion |
| 10 | Modulo contable integrado | en curso | asientos automaticos por evento |
| 11 | Reportes financieros | en curso | balance, estado de resultados, flujo de caja |
| 12 | Cierres de caja | en curso | arqueo y cierre por sucursal/empresa |
| 13 | Calidad, UAT y despliegue | pendiente | validacion integral y salida controlada |
| 14 | Operacion continua | pendiente | mejora continua con KPI y roadmap trimestral |
| 15 | Modulo de auditoria por empresa | en curso | trazabilidad por usuario/accion/recurso con consulta por empresa |

### Punto 1 + Punto 2. Cierre de backlog inmediato (2026-04-04)

Entregables completados:
- Punto 1:
	- `documentos/matriz_kpi_pos_multiempresa.md` queda formalizada con:
		- formula implementada (nivel SQL/logica),
		- endpoint canonico de consumo,
		- tablas fuente reales por KPI.
- Punto 2:
	- se crea `documentos/matriz_entidades_multiempresa_aislamiento.md` con:
		- inventario completo de endpoints `/api/empresa/*`,
		- llave primaria de aislamiento (`empresa_id`),
		- llaves secundarias por recurso/modulo,
		- tipo de control de alcance (middleware o validacion interna).

Criterio de cierre aplicado:
- Se valida trazabilidad endpoint -> handler -> fuente de datos real.
- Se registran excepciones de aislamiento de forma explicita (catalogos globales y rutas de autenticacion).

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

### Punto 3. Permisos y seguridad (consolidacion endpoint/rol + checklist UAT, 2026-04-04)

Consolidacion documental completada:
- `documentos/matriz_roles_permisos_pos_multiempresa.md` queda ampliada con:
	- matriz final endpoint/rol alineada a wrappers reales de `backend/main.go` y reglas de `backend/handlers/empresa_permisos.go`,
	- excepciones fuera de wrapper (login/establecer_password/catalogos/chat IA con validacion por cuenta Google),
	- checklist UAT de punto 3 con evidencia por prueba automatizada.

Validacion ejecutada (corte actual):
- `runTests` sobre:
	- `backend/handlers/empresa_permisos_test.go`,
	- `backend/handlers/auditoria_empresa_test.go`.
- Resultado: 25 pruebas aprobadas, 0 fallidas.

Estado operativo del punto 3 tras esta consolidacion:
- Definicion de permisos por endpoint y accion: consolidada.
- Evidencia automatizada de denegacion/aprobacion por rol: consolidada.
- Pendiente para cierre total del punto: exposicion de catalogo de permisos en frontend y regresion especifica de endpoints sin wrapper de modulo.

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

### Punto 12. Cierres de caja (continuacion 2026-04-04 - UI operativa en panel empresa)

Implementacion tecnica adicional completada:
- `web/administrar_empresa/finanzas.html` integra interfaz operativa de cierres de caja con:
	- formulario de apertura/actualizacion por `sucursal_id`, `caja_codigo`, `turno` y fecha,
	- calculo visual de `caja_teorica` y `diferencia_caja`,
	- filtros por sucursal/caja/estado/rango de fechas e inclusion de inactivos,
	- tabla de ejecucion con acciones de ciclo (`cerrar`, `reabrir`, `aprobar`, `anular`) y estado de registro (`activar/desactivar`, `eliminar`).
- La UI consume `GET/POST/PUT/DELETE /api/empresa/finanzas/cierres_caja` reutilizando el contrato ya implementado en backend.
- Se agregan KPI visuales de apoyo operativo para seguimiento rapido de:
	- cajas abiertas,
	- cierres cerrados/aprobados,
	- cierres con incidencia.

Validacion ejecutada:
- `get_errors` sobre `web/administrar_empresa/finanzas.html` (ok).

### Punto 12. Cierres de caja (continuacion 2026-04-04 - UAT por rol y matriz de transiciones)

Implementacion tecnica adicional completada:
- Se amplian pruebas en `backend/handlers/empresa_permisos_test.go` para UAT de autorizacion en `/api/empresa/finanzas/cierres_caja`:
	- `TestWithEmpresaFinanzasPermissionsDeniesCajeroAprobarCierreCaja`.
	- `TestWithEmpresaFinanzasPermissionsDeniesSupervisorAprobarCierreCaja`.
	- `TestWithEmpresaFinanzasPermissionsAllowsAdminAprobarCierreCaja`.
- Se agrega en `documentos/matriz_roles_permisos_pos_multiempresa.md` una matriz UAT de cierres con:
	- casos por rol,
	- casos de transicion de estados (`abierto`, `cerrado`, `aprobado`, `anulado`) y resultado HTTP esperado.

Validacion ejecutada:
- `go test ./handlers -run "TestWithEmpresaFinanzasPermissions(DeniesCajeroAprobarCierreCaja|DeniesSupervisorAprobarCierreCaja|AllowsAdminAprobarCierreCaja)" -count=1` (ok).

### Punto 10. Modulo contable integrado (definicion de estrategia 2026-04-04)

Estrategia de procesamiento de asientos definida:
- Fuente unica de eventos: consumir pendientes desde `empresa_eventos_contables` por `empresa_id` y `procesado=0`.
- Resolucion canonica documental:
	- facturacion desde `empresa_facturacion_documentos`,
	- compras desde `empresa_compras_documentos`,
	- usar `entidad_id` como referencia estable para idempotencia.
- Regla de idempotencia: una combinacion (`empresa_id`, `modulo`, `evento`, `entidad_id`, `documento_codigo`) no debe generar asientos duplicados.
- Pipeline propuesto de ejecucion:
	1) seleccionar lote ordenado por `id` asc,
	2) validar contrato de evento y estado documental,
	3) mapear cuentas (configuracion financiera por empresa),
	4) persistir asientos en transaccion,
	5) marcar evento como procesado con trazabilidad de fecha/resultado.
- Manejo de errores y reintentos:
	- errores funcionales: marcar observacion y dejar pendiente para correccion,
	- errores transitorios: reintentar por lote con backoff y tope de intentos.
- Entregables de implementacion siguientes:
	- tabla canonica de asientos contables por empresa,
	- worker o endpoint de procesamiento por lote,
	- pruebas de idempotencia y consistencia debito/haber.

### Punto 10 + Punto 11. Continuacion ejecutada (2026-04-04 - backlog 1 y 2)

Implementacion tecnica completada:
- `backend/db/eventos_contables.go` amplia `empresa_eventos_contables` con trazabilidad de procesamiento:
	- `intentos_procesamiento`,
	- `fecha_ultimo_intento`,
	- `error_procesamiento`,
	- `asiento_contable_id`.
- Se crea tabla canonica `empresa_asientos_contables` con:
	- referencia al evento (`evento_contable_id`),
	- lineas contables serializadas,
	- hash de idempotencia (`hash_idempotencia`) con restriccion unica,
	- control de debito/credito y diferencia.
- Se implementa proceso por lotes en DB para convertir eventos pendientes en asientos:
	- seleccion por `empresa_id` y `procesado=0`,
	- persistencia idempotente,
	- marcacion de exito/fallo por evento con contador de intentos.
- `backend/handlers/finanzas.go` agrega endpoint:
	- `GET /api/empresa/finanzas/asientos_contables` (consulta),
	- `POST/PUT action=procesar_asientos|procesar` (lote manual).
- `backend/handlers/empresa_permisos.go` clasifica `action=procesar_asientos` como accion de aprobacion (`A`).
- `backend/main.go` publica ruta de asientos y registra migracion:
	- `2026-04-04-010-asientos-canonicos`.
- `backend/db/finanzas.go` integra en tablero minimo:
	- `estado_resultados`,
	- `balance_general`,
	- KPI contables de asientos (`asientos_generados`, `asientos_monto_total`).
- `web/administrar_empresa/reportes.html` renderiza nuevos KPI de estado de resultados y balance general.
- `web/administrar_empresa/finanzas.html` agrega accion manual `Procesar eventos contables`.

Cobertura de pruebas agregada/actualizada:
- `backend/db/eventos_contables_test.go`:
	- `TestProcessEmpresaEventosContablesPendientesGeneraAsientosIdempotentes`.
- `backend/db/finanzas_test.go`:
	- `TestGetEmpresaReportesTableroResumenConAsientosCanonicos`.
- `backend/handlers/eventos_contables_modulos_test.go`:
	- `TestEmpresaFinanzasAsientosContablesHandlerProcesaPendientes`.
- `backend/handlers/empresa_permisos_test.go`:
	- `TestWithEmpresaFinanzasPermissionsDeniesCajeroProcesarAsientos`.
	- `TestWithEmpresaFinanzasPermissionsAllowsContabilidadProcesarAsientos`.

Validacion ejecutada:
- `go test ./db -run "EventosContables|ReportesTableroResumen" -count=1` (ok).
- `go test ./handlers -run "AsientosContables|TableroResumen|WithEmpresaFinanzasPermissions" -count=1` (ok).
- `go test ./handlers -count=1` (ok).
- `go test ./db -count=1` (ok).

### Punto 15. Modulo de auditoria por empresa (nuevo)

Definicion funcional incorporada al plan:
- Alcance minimo del modulo:
	- registrar eventos de auditoria por `empresa_id`, usuario y accion,
	- guardar recurso afectado (modulo, entidad, entidad_id, endpoint/metodo),
	- persistir resultado (`ok`/`error`) con metadatos de trazabilidad (`request_id`, timestamp, ip cuando aplique),
	- exponer consulta filtrable por empresa, rango de fechas, modulo y usuario.
- Criterios de uso:
	- toda accion critica (creacion, actualizacion, eliminacion, aprobacion y procesos por lote) debe emitir evento de auditoria,
	- las consultas de auditoria deben respetar permisos por rol y alcance de empresa.
- Entregables iniciales del punto 15:
	- esquema canonico `empresa_auditoria_eventos`,
	- helper/middleware de registro no bloqueante,
	- endpoint de consulta para panel empresa,
	- pruebas de integridad y alcance por permisos.

### Punto 15. Modulo de auditoria por empresa (continuacion 2026-04-04 - base minima implementada)

Implementacion tecnica completada:
- Se agrega `backend/db/auditoria_empresa.go` con:
	- esquema `empresa_auditoria_eventos`,
	- filtros de consulta por `empresa_id`, modulo, accion, resultado, usuario, `request_id` y rango de fechas,
	- politica de retencion configurable por registro (`retencion_dias`) y funcion de purga por empresa.
- Se integra middleware de auditoria no bloqueante en `backend/handlers/empresa_permisos.go`:
	- toda accion critica autorizada (`C/U/D/A`) registra automaticamente evento de auditoria,
	- se persiste modulo, accion, recurso, metodo, endpoint, resultado (`ok/error`), codigo HTTP, IP, metadatos y usuario.
- Se agrega `backend/handlers/auditoria_empresa.go` con endpoint:
	- `GET /api/empresa/auditoria/eventos` para consulta filtrable,
	- `PUT/POST /api/empresa/auditoria/eventos?action=retener|purgar` para aplicar retencion manual.
- Se actualiza `backend/main.go`:
	- bootstrap de esquema con `EnsureEmpresaAuditoriaSchema`,
	- migracion `2026-04-04-011-auditoria-empresa`,
	- ruta protegida `/api/empresa/auditoria/eventos` bajo `WithEmpresaSeguridadPermissions`.
- Cobertura de pruebas nueva:
	- `backend/db/auditoria_empresa_test.go`.
	- `backend/handlers/auditoria_empresa_test.go`.

Validacion ejecutada:
- `go test ./db -run "Auditoria|EventosContables|ReportesTableroResumen" -count=1` (ok).
- `go test ./handlers -run "Auditoria|AsientosContables|WithEmpresaFinanzasPermissions" -count=1` (ok).
- `go test ./handlers -count=1` (ok).
- `go test ./db -count=1` (ok).

### Punto 15. Modulo de auditoria por empresa (continuacion 2026-04-04 - cierre de backlog 1, 2 y 3)

Implementacion tecnica completada:
- Cobertura automatica de acciones criticas en modulos transaccionales:
	- `backend/handlers/empresa_permisos.go` amplia alias operativos de clasificacion en `ventas`, `compras` y `facturacion` para asegurar registro de auditoria en acciones criticas.
	- `backend/handlers/auditoria_empresa.go` enriquece metadata de trazabilidad por dominio (`carrito_id`, `proveedor_id`, `entidad_id`, `documento_codigo`).
- Vista de auditoria en panel empresa:
	- Nuevo `web/administrar_empresa/auditoria.html` con consulta filtrable por modulo/accion/resultado/usuario/request_id/rango.
	- Soporte de accion manual de retencion (`retencion_dias`) desde UI.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran nuevo acceso de menu `Auditoria`.
- Purga automatica programada:
	- `backend/db/auditoria_empresa.go` agrega `PurgeExpiredEmpresaAuditoriaEventos` y `StartEmpresaAuditoriaRetentionWorker`.
	- `backend/main.go` arranca worker de purga automatica cada 12 horas.
	- La limpieza usa `fecha_expiracion` y fallback por `retencion_dias` para registros legacy.

Cobertura de pruebas agregada:
- `backend/handlers/auditoria_empresa_test.go`:
	- `TestWithEmpresaVentasPermissionsRegistraAuditoriaAccionCritica`.
	- `TestWithEmpresaComprasPermissionsRegistraAuditoriaAccionCritica`.
	- `TestWithEmpresaFacturacionPermissionsRegistraAuditoriaAccionCritica`.
- `backend/db/auditoria_empresa_test.go`:
	- `TestPurgeExpiredEmpresaAuditoriaEventos`.

Validacion ejecutada:
- `go test ./handlers -run "Auditoria|WithEmpresa(Ventas|Compras|Facturacion|Finanzas)Permissions" -count=1` (ok).
- `go test ./db -run "Auditoria" -count=1` (ok).
- `go test ./handlers -count=1` (ok).
- `go test ./db -count=1` (ok).

### Punto 15. Modulo de auditoria por empresa (continuacion 2026-04-04 - cierre de backlog inmediato 1 y 2)

Implementacion tecnica completada:
- Exportacion CSV/JSON en UI de auditoria:
	- `web/administrar_empresa/auditoria.html` agrega botones de exportacion directa de resultados filtrados.
	- La exportacion soporta trazabilidad directiva por rango/modulo segun filtros activos.
- Filtros avanzados por `codigo_http` y `recurso_id` en endpoint/UI:
	- `backend/db/auditoria_empresa.go` amplia `EmpresaAuditoriaEventoFilter` y consulta SQL en `ListEmpresaAuditoriaEventos`.
	- `backend/handlers/auditoria_empresa.go` valida parametros y expone filtros en `GET /api/empresa/auditoria/eventos`.
	- `web/administrar_empresa/auditoria.html` incorpora controles de filtro para ambos campos.

Cobertura de pruebas agregada/extendida:
- `backend/db/auditoria_empresa_test.go` aplica filtros combinados (`recurso_id` + `codigo_http`) en listado.
- `backend/handlers/auditoria_empresa_test.go` agrega `TestEmpresaAuditoriaEventosHandlerFiltrosAvanzados` (filtro combinado + validacion `400` para parametro invalido).

Validacion ejecutada:
- `go test ./db -run "Auditoria" -count=1` (ok).
- `go test ./handlers -run "Auditoria" -count=1` (ok).
- `go test ./handlers -count=1` (ok).
- `go test ./db -count=1` (ok).

### Punto 10. Modulo contable integrado (continuacion 2026-04-04 - automatizacion por lotes controlada)

Implementacion tecnica completada:
- Ejecucion automatica por lotes para `procesar_asientos` con politica configurable:
	- `backend/db/eventos_contables.go` agrega:
		- `ProcessEmpresaEventosContablesPendientesConPolitica` (soporte de `max_reintentos`),
		- `RunEmpresaAsientosContablesWorkerCycle` (corrida global por empresas pendientes),
		- `StartEmpresaAsientosContablesWorker` (worker periodico de asientos).
	- La seleccion de pendientes ahora puede filtrar por limite de reintentos (`intentos_procesamiento < max_reintentos`).
- Integracion en arranque de servidor:
	- `backend/main.go` arranca worker automatico de asientos y carga politica por variables de entorno:
		- `ASIENTOS_WORKER_INTERVAL_MINUTES`,
		- `ASIENTOS_WORKER_BATCH_SIZE`,
		- `ASIENTOS_WORKER_MAX_RETRIES`.
- Endpoints manuales alineados a la politica:
	- `backend/handlers/finanzas.go` permite `max_reintentos` opcional en `POST/PUT /api/empresa/finanzas/asientos_contables?action=procesar_asientos`.

Cobertura de pruebas agregada/extendida:
- `backend/db/eventos_contables_test.go` agrega `TestProcessEmpresaEventosContablesPendientesConPoliticaRespetaMaxReintentos`.
- `backend/handlers/eventos_contables_modulos_test.go`:
	- amplía prueba de proceso manual con `max_reintentos`,
	- agrega validacion `400` para `max_reintentos` invalido.

Validacion ejecutada:
- `go test ./db -run "EventosContables|ConPolitica|Asientos" -count=1` (ok).
- `go test ./handlers -run "AsientosContablesHandler|FinanzasAsientos" -count=1` (ok).
- `go test ./handlers -count=1` (ok).
- `go test ./db -count=1` (ok).

### Punto 10. Modulo contable integrado (continuacion 2026-04-04 - vista de conciliacion por periodo)

Implementacion tecnica completada:
- Vista de conciliacion contable por periodo (eventos vs asientos):
	- `backend/db/eventos_contables.go` agrega:
		- `EmpresaConciliacionContableFilter`,
		- `EmpresaConciliacionContablePeriodo`,
		- `EmpresaConciliacionContableResumen`,
		- `GetEmpresaConciliacionContablePorPeriodo` para consolidar por periodo los totales de eventos, procesados, pendientes, errores, asientos y desfases.
	- `backend/handlers/finanzas.go` amplía `GET /api/empresa/finanzas/asientos_contables` con `action=conciliacion_periodo|conciliacion`.
	- `web/administrar_empresa/finanzas.html` agrega tarjeta de conciliacion por periodo con:
		- filtros por rango, periodo y limite,
		- KPIs de periodos conciliados/pendientes/descuadre,
		- tabla de comparativo eventos vs asientos.

Cobertura de pruebas agregada/extendida:
- `backend/db/eventos_contables_test.go` agrega `TestGetEmpresaConciliacionContablePorPeriodo`.
- `backend/handlers/eventos_contables_modulos_test.go` agrega `TestEmpresaFinanzasAsientosContablesHandlerConciliacionPeriodo`.

Validacion ejecutada:
- `go test ./db -run "EventosContables|ConPolitica|Conciliacion" -count=1` (ok).
- `go test ./handlers -run "AsientosContablesHandler|ConciliacionPeriodo" -count=1` (ok).
- `go test ./db -count=1` (ok).
- `go test ./handlers -count=1` (ok).

### Punto 11. Reportes financieros (continuacion 2026-04-04 - exportacion unificada del tablero)

Implementacion tecnica completada:
- Exportacion unificada del tablero por rango (`estado_resultados` + `balance_general`):
	- `backend/handlers/finanzas.go` amplía `GET /api/empresa/finanzas/movimientos` con `action=tablero_export` para descarga en:
		- `format=json` (payload unificado del tablero),
		- `format=csv` (matriz unificada por bloque/metrica/valor).
	- La exportacion CSV incluye bloques:
		- `operativo`,
		- `financiero`,
		- `contable`,
		- `estado_resultados`,
		- `balance_general`.
	- `web/administrar_empresa/reportes.html` incorpora botones:
		- `Exportar tablero CSV`,
		- `Exportar tablero JSON`,
		con descarga por rango activo (`desde`, `hasta`).

Cobertura de pruebas agregada/extendida:
- `backend/handlers/eventos_contables_modulos_test.go` agrega `TestEmpresaFinanzasTableroResumenExportHandler` para validar:
	- descarga JSON con bloques `estado_resultados` y `balance_general`,
	- descarga CSV unificada con filas de ambos bloques,
	- error `400` para `format` invalido.

Validacion ejecutada:
- `go test ./handlers -run "TestEmpresaFinanzasTableroResumenHandler|TestEmpresaFinanzasTableroResumenExportHandler|TestEmpresaFinanzasAsientosContablesHandlerConciliacionPeriodo" -count=1` (ok).
- `go test ./handlers -count=1` (ok).
- `go test ./db -count=1` (ok).

## Backlog inmediato (siguiente iteracion)

1. Cerrar Punto 3 (permisos y seguridad): consolidar matriz final por endpoint/rol y completar evidencia UAT en acciones criticas.
2. Iniciar Punto 5 (control de inventarios): formalizar kardex operativo, reglas de stock min/max y alertas de quiebre por bodega.

## Criterios de avance para la siguiente fase

- Punto 1 queda en estado completo cuando exista una matriz formal de KPI con formulas y fuente de datos por endpoint/tabla.
- Punto 2 queda en estado completo cuando exista matriz de entidades con llaves de aislamiento (empresa/sucursal/bodega) y reglas de validacion por endpoint.
- Punto 10 queda en estado completo cuando exista proceso documentado y probado para convertir eventos en asientos con referencia canonica de documento (`entidad_id`) y ejecucion automatica controlada.
- Punto 15 queda en estado completo cuando el registro de auditoria por empresa cubra acciones criticas, tenga consulta segura por rol y cuente con pruebas automatizadas de trazabilidad.
