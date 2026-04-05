# CHANGELOG

## 2026-04-05
- Facturación electrónica: envío automático del resumen de factura al correo del cliente al emitir.
	- `backend/handlers/facturacion_electronica.go` ahora intenta enviar correo en `action=emitir` de `factura_electronica`.
	- Soporta destinatario por `cliente_email` o por `cliente_id`/`entidad_id` consultando clientes.
	- La respuesta incluye bloque `factura_email` con estado de intento/envío/error sin bloquear la emisión legal.
	- `backend/db/clientes.go` agrega `GetClienteByID` para resolver destinatario desde la base de datos.
	- `backend/main.go` actualiza la inyección de `dbSuper` al handler de facturación para lectura de SMTP.
	- `web/administrar_empresa/facturacion_electronica.html` agrega campos de cliente y muestra el resultado de envío en pantalla.
	- Cobertura añadida en `backend/db/clientes_test.go` y `backend/handlers/eventos_contables_modulos_test.go`.

## 2026-04-05
- Se crea el modulo de codigos de descuento por empresa y validacion de metodos de pago en carrito de compras.
	- `backend/db/codigos_descuento.go` (nuevo) agrega la tabla `codigos_de_descuento`, generacion automatica de codigos, CRUD, validacion por vencimiento/usos y resolucion de descuento aplicable por monto.
	- `backend/handlers/codigos_descuento.go` (nuevo) expone `/api/empresa/codigos_de_descuento` con operaciones CRUD, activar/desactivar y `action=validar`.
	- `backend/db/carritos_compras.go` agrega campos `metodo_pago` y `referencia_pago`, normaliza metodos permitidos y registra consumo transaccional de codigo de descuento al cerrar venta.
	- `backend/handlers/carritos_compras.go` valida `metodo_pago` (`efectivo`, `tarjeta_credito`, `tarjeta_debito`, `codigo_descuento`) y exige referencia para pagos con tarjeta.
	- `backend/main.go` asegura esquema `codigos_de_descuento`, registra migracion `2026-04-05-012-codigos-descuento-pagos` y expone ruta protegida de codigos de descuento.
	- `web/administrar_empresa/codigos_de_descuento.html` (nuevo) incorpora modulo profesional para crear/editar/activar/eliminar codigos con valor y fecha de vencimiento.
	- `web/administrar_empresa/carrito_de_compras.html` agrega selector de metodo de pago, referencia y aplicacion de codigos de descuento con validacion operativa.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran el enlace de menu `Codigos de descuento` con permisos del modulo ventas.
	- `backend/db/codigos_descuento_test.go` y `backend/handlers/auth_users_carritos_test.go` agregan cobertura para validacion/uso de codigos y rechazo de metodo de pago invalido.

## 2026-04-05
- Se crea el modulo de combos de productos con receta de ingredientes y precio unico de venta.
	- `backend/handlers/combos_productos.go` (nuevo) expone `/api/empresa/combos_productos` con operaciones CRUD y acciones `activar/desactivar`.
	- `backend/db/productos.go` incorpora esquema y logica de combos (`combos_productos`, `combos_productos_detalle`) con controles de consistencia para carritos abiertos.
	- `backend/db/carritos_compras.go` extiende el ajuste de inventario para descontar/liberar stock por ingrediente cuando el item es `tipo_item=combo`.
	- `backend/handlers/carritos_compras.go` valida `referencia_id` obligatorio para items combo.
	- `backend/main.go` registra la nueva ruta protegida bajo permisos de inventario.
	- `web/administrar_empresa/combos_productos.html` (nuevo) agrega interfaz completa para gestionar combos y receta.
	- `web/administrar_empresa/carrito_de_compras.html` incorpora busqueda/catalogo y visualizacion de combos en carrito.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran el modulo en menu y permisos.
	- `backend/db/productos_categorias_test.go` y `backend/db/carritos_inventario_test.go` agregan cobertura de CRUD y flujo de inventario por ingredientes.

## 2026-04-05
- Se crea el modulo de graficos y estadisticas por empresa.
	- `backend/handlers/graficos_estadisticas.go` (nuevo) expone `/api/empresa/graficos_estadisticas` con acciones `panel`, `serie`, `rankings`, `distribuciones` y `catalogo`.
	- `backend/main.go` registra la nueva ruta protegida bajo permisos de finanzas.
	- `backend/handlers/graficos_estadisticas_test.go` (nuevo) agrega cobertura de contrato HTTP y validaciones de error.
	- `web/administrar_empresa/graficos_estadisticas.html` (nuevo) incorpora panel visual con series, distribuciones y rankings.
	- `web/estilos.css` agrega estilos responsivos del nuevo modulo.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran el acceso en menu con control de permisos.
	- `web/ayuda/ayuda.html` incorpora guia y API del modulo de analitica.

## 2026-04-05
- Se crea el modulo de control de asistencia de empleados por empresa.
	- `backend/db/asistencia_empleados.go` (nuevo) agrega tabla `empresa_asistencia_empleados` y operaciones CRUD con marcacion de entrada/salida.
	- `backend/handlers/asistencia_empleados.go` (nuevo) expone `/api/empresa/asistencia_empleados` con acciones operativas de asistencia.
	- `backend/main.go` incorpora esquema, migracion `2026-04-05-010-asistencia-empleados` y registro de ruta protegida.
	- `web/administrar_empresa/asistencia_empleados.html` (nuevo) agrega UI completa para gestion diaria de asistencia.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran el modulo en menu y permisos.
	- `backend/handlers/asistencia_empleados_test.go` (nuevo) valida flujo funcional del modulo.
	- Se actualizan `web/ayuda/ayuda.html`, `estructura_bd.md` y diagramas/documentacion tecnica para trazabilidad.

## 2026-04-05
- Modulo de reportes robustecido a nivel empresarial, operativo y contable con enfoque escalable por dataset.
	- `backend/handlers/reportes.go` (nuevo) implementa `/api/empresa/reportes` con acciones `catalogo`, `suite`, `dataset`, `tablero` y `export`.
	- Se habilitan exportaciones multi-formato para datasets: `JSON`, `CSV`, `TXT` y `XLS`.
	- `backend/main.go` registra la nueva ruta protegida bajo permisos de finanzas.
	- `web/administrar_empresa/reportes.html` incorpora selector de dataset, vista tabular profesional y exportes desde interfaz.
	- `backend/handlers/reportes_test.go` (nuevo) agrega cobertura de contrato HTTP y validacion de exportaciones.
	- Se actualizan diagramas de arquitectura/flujo en `documentos/diagramas/estructura_del_codigo.md` y `documentos/diagramas/diagrama_flujo_procesos.md`.

## 2026-04-04
- Centro de ayuda actualizado con tutorial por cada módulo del sistema.
	- `web/ayuda/ayuda.html` amplía el contenido con una sección de tutoriales por módulos de administración global y módulos del panel de empresa.
	- Se agregan pasos operativos por módulo y enlaces directos a cada pantalla para facilitar onboarding y uso diario.

## 2026-04-04
- Verificacion integral real de modulos + limpieza de artefactos temporales.
	- Validacion real ejecutada (sin simulaciones/mocks) sobre SQLite y capa HTTP:
		- `go test ./auth ./db ./handlers ./metrics ./utils -count=1` (ok).
		- `go test ./... -count=1` (ok).
		- `go test ./handlers -run "TestEmpresaScope|FueraDeAlcance|WithEmpresa|isol|Aisla|multiempresa|UsuariosHandlerAislaEmpresa|ConsolidaEmpresa" -count=1` (ok).
		- `go test ./handlers -run "TestEmpresaClientes|TestEmpresaProveedores|TestEmpresaFacturacion|TestEmpresaCompras|TestEmpresaInventario|TestEmpresaFinanzas|TestEmpresaAuditoria|TestEmpresaCarritos|TestEmpresaUsuarios|TestModelosHandler" -count=1` (ok).
		- `go test ./db -run "Test.*(Cliente|Proveedor|Facturacion|Compra|Inventario|Finanzas|Evento|Auditoria|Scope|Empresa)" -count=1` (ok).
	- Se eliminan artefactos temporales/no usados del repositorio:
		- `backend/tmp_api.json`.
		- `backend/tmp_config.html`.
		- `backend/server.err`.
		- `backend/server.run.err`.
		- `backend/db/empresas.db.20260326-174525.bak`.
		- `backend/db/superadministrador.db.20260326-174324.bak`.
		- `backend/db/superadministrador.db.20260326-174525.bak`.

## 2026-04-04
- Punto 14 (operacion continua) - inicio operativo con KPI y roadmap trimestral.
	- `documentos/punto_14_operacion_continua.md` (nuevo): define marco de mejora continua y cadencia de seguimiento.
	- `documentos/roadmap_trimestral_pos_multiempresa.md` (nuevo): formaliza roadmap Q2/Q3/Q4 2026.
	- `scripts/generar_reporte_operacion_continua.ps1` (nuevo): genera reporte operativo y bitacora tecnica.
	- `documentos/punto_14_operacion_continua_reporte.md` (nuevo): evidencia de la ultima corrida operativa.
	- `documentos/plan_maestro_pos_multiempresa_14_puntos.md`: punto 14 actualizado a `en curso`.
- Validacion tecnica:
	- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\generar_reporte_operacion_continua.ps1` (ok).

## 2026-04-04
- Punto 13 (calidad, UAT y despliegue) - arranque operativo con validacion integral automatizada.
	- `scripts/validar_punto_13.ps1` (nuevo): ejecuta gate tecnico y genera evidencia automatica.
	- `documentos/punto_13_calidad_uat_despliegue.md` (nuevo): formaliza flujo de calidad/UAT/salida controlada.
	- `documentos/punto_13_validacion_integral_resultado.md` (nuevo): reporte de ultima validacion tecnica.
	- `documentos/release_checklist.md`: incorpora gate del punto 13 y verificacion de evidencia.
	- `documentos/plan_maestro_pos_multiempresa_14_puntos.md`: punto 13 pasa a `en curso`.
- Validacion tecnica:
	- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\validar_punto_13.ps1` (ok).
	- `go test ./auth ./db ./handlers ./metrics ./utils -count=1` (ok).
	- `go test ./... -count=1` (ok).

## 2026-04-04
- Punto 8 (facturacion electronica) - refuerzo de cobertura en cumplimiento legal de emision.
	- `backend/db/facturacion_electronica_test.go` (nuevo) agrega pruebas unitarias para `PrepareFacturacionDocumentoLegal`:
		- `TestPrepareFacturacionDocumentoLegalSuccessAndConsecutivo`.
		- `TestPrepareFacturacionDocumentoLegalRejectsExpiredResolution`.
		- `TestPrepareFacturacionDocumentoLegalRejectsConfigInactivaAndRangoAgotado`.
	- Se valida reserva e incremento de consecutivo legal, rechazo por resolucion vencida, rechazo por configuracion FE inactiva y agotamiento de rango.
- Validacion tecnica:
	- `gofmt -w db/facturacion_electronica_test.go` (ok).
	- `go test ./db -run "TestPrepareFacturacionDocumentoLegal" -count=1` (ok).
	- `go test ./db ./handlers -run "TestPrepareFacturacionDocumentoLegal|TestEmpresaDocumentoFacturacionUpsertAndGet|TestEmpresaFacturacionTransaccional" -count=1` (ok).

## 2026-04-04
- Punto 9 (modulo de compras) - avance funcional con endpoint y vista dedicados para ciclo documental.
	- `backend/db/documentos_transaccionales.go` agrega:
		- `ListEmpresaDocumentosCompraByEmpresa`.
		- `SetEmpresaDocumentoCompraEstadoByCodigo`.
	- `backend/handlers/compras.go` (nuevo) implementa `GET/POST/PUT/DELETE /api/empresa/compras/documentos` con acciones documentales (`crear`, `emitir_orden`, `recepcionar_compra`, `contabilizar_compra`) y activar/desactivar.
	- `backend/main.go` registra la ruta protegida `/api/empresa/compras/documentos`.
	- `web/administrar_empresa/compras.html` (nuevo) incorpora interfaz dedicada de compras para crear, consultar y transicionar documentos.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran el acceso de menu `Compras` con control por permisos de modulo.
	- Cobertura agregada en:
		- `backend/db/documentos_transaccionales_test.go`.
		- `backend/handlers/compras_documentos_test.go` (nuevo).
- Validacion tecnica:
	- `gofmt -w handlers/compras.go handlers/compras_documentos_test.go main.go db/documentos_transaccionales.go db/documentos_transaccionales_test.go` (ok).
	- `go test ./db -run "TestEmpresaDocumentoCompraListAndSetEstadoByCodigo" -count=1` (ok).
	- `go test ./handlers -run "TestEmpresaComprasDocumentos" -count=1` (ok).
	- `go test ./db ./handlers -run "TestEmpresaDocumentoCompraListAndSetEstadoByCodigo|TestEmpresaComprasDocumentos" -count=1` (ok).
	- `go test ./... -run "TestEmpresaComprasDocumentos|TestEmpresaDocumentoCompraListAndSetEstadoByCodigo" -count=1` (ok).

## 2026-04-04
- Punto 8 (facturacion electronica) - avance funcional de emision legal y cumplimiento normativo inicial.
	- `backend/db/facturacion_electronica.go` agrega `PrepareFacturacionDocumentoLegal` para validar configuracion legal, vigencia de resolucion y rango de consecutivos por empresa/pais antes de emitir.
	- `backend/db/documentos_transaccionales.go` amplia `empresa_facturacion_documentos` con metadata legal persistida: `numero_legal`, `codigo_validacion`, `pais_codigo`, `ambiente_fe`.
	- `backend/handlers/facturacion_electronica.go` endurece `action=emitir` con rechazo `422` cuando no hay cumplimiento normativo y devuelve bloque `cumplimiento_normativo` en emisiones exitosas.
	- `web/administrar_empresa/facturacion_electronica.html` incorpora bloque operativo para `emitir`, `anular` y `nota_credito`, con visualizacion del resultado legal.
	- Cobertura extendida en:
		- `backend/db/documentos_transaccionales_test.go`.
		- `backend/handlers/eventos_contables_modulos_test.go`.
- Validacion tecnica:
	- `go test ./db -run "TestEmpresaDocumentoFacturacionUpsertAndGet" -count=1` (ok).
	- `go test ./handlers -run "TestEmpresaFacturacionTransaccionalEmiteEventosContables|TestEmpresaFacturacionTransaccionalEmitirRechazaSinCumplimientoLegal|TestEmpresaFacturacionTransaccionalRechazaTransicionInvalida" -count=1` (ok).
	- `go test ./db ./handlers -count=1` (ok).

## 2026-04-04
- Punto 7 (gestion de proveedores) - avance funcional de catalogo, precios y condiciones comerciales.
	- `backend/db/productos.go` amplia el modelo `Proveedor` y su migracion segura con campos:
		- `catalogo_referencia`,
		- `precio_base_referencial`,
		- `descuento_porcentaje`,
		- `plazo_pago_dias`,
		- `condicion_entrega`.
	- `backend/handlers/productos.go` agrega validacion HTTP de rango para los nuevos campos en `POST/PUT /api/empresa/proveedores` y enriquece metadata de eventos contables de compras.
	- `web/administrar_empresa/administrar_productos.html` amplia el formulario y la tabla de proveedores para gestionar y visualizar datos comerciales.
	- Cobertura nueva/extendida en:
		- `backend/db/productos_categorias_test.go`.
		- `backend/handlers/eventos_contables_modulos_test.go`.
- Validacion tecnica:
	- `gofmt -w db/productos.go db/productos_categorias_test.go handlers/productos.go handlers/eventos_contables_modulos_test.go` (ok).
	- `go test ./db -run "TestProveedorCRUDIncluyeCatalogoPreciosYCondiciones" -count=1` (ok).
	- `go test ./handlers -run "TestEmpresaProveedoresEmiteEventoContableCompras|TestEmpresaProveedoresRechazaCamposComercialesInvalidos" -count=1` (ok).
	- `get_errors` sobre backend/frontend modificado (ok).

## 2026-04-04
- Punto 6 (gestion de clientes) - avance funcional de perfil, historial y segmentacion.
	- `backend/db/clientes.go` agrega contratos analiticos (`ClientePerfilComercial`, `ClienteCompraHistorial`, `ClienteSegmentacionResumen`) y funciones de consulta por cliente/empresa.
	- `backend/handlers/clientes.go` amplia `GET /api/empresa/clientes` con `action=perfil`, `action=historial`, `action=segmentacion|segmentos`.
	- `web/administrar_empresa/administrar_clientes.html` agrega paneles de segmentacion y de perfil/historial por cliente con accion `Perfil`.
	- Cobertura nueva en:
		- `backend/db/clientes_test.go`.
		- `backend/handlers/clientes_test.go`.
- Validacion tecnica:
	- `gofmt -w db/clientes.go db/clientes_test.go handlers/clientes.go handlers/clientes_test.go` (ok).
	- `go test ./db -run "TestGetClientePerfilComercialByEmpresaAndHistorial|TestGetClientePerfilComercialByEmpresaSinComprasSegmentoNuevo" -count=1` (ok).
	- `go test ./handlers -run "TestEmpresaClientesHandlerPerfilHistorialSegmentacion" -count=1` (ok).
	- `get_errors` sobre backend/frontend modificado (ok).

## 2026-04-04
- Punto 5 (control de inventarios) — continuidad preventiva-compras ciclo documental desde reposicion.
	- `backend/db/productos.go` agrega `InventarioPlanReposicionOrdenEstadoActualizado` y `ActualizarEstadoOrdenCompraDesdeReposicion` para transiciones `recepcionar_compra` y `contabilizar_compra`.
	- `backend/handlers/productos.go` agrega endpoint `POST /api/empresa/compras/plan_reposicion/actualizar_estado`.
	- `backend/main.go` registra `/api/empresa/compras/plan_reposicion/actualizar_estado` bajo permisos de compras.
	- `web/administrar_empresa/administrar_productos.html` amplía el flujo a `fases 10-12` con acciones `Recepcionar orden` y `Contabilizar orden` y contexto de estado de OC.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`.
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./db -run "TestEmitirOrdenCompraDesdePlanReposicionBorradorPersistDoc|TestActualizarEstadoOrdenCompraDesdeReposicionCiclo"` (ok).
	- `go test ./handlers -run "TestEmpresaComprasPlanReposicionEmitirOrdenHandlerEmiteDocumento|TestEmpresaComprasPlanReposicionActualizarEstadoHandlerGestionaCiclo"` (ok).
	- `get_errors` sobre backend/frontend modificado (ok).

## 2026-04-04
- Punto 5 (control de inventarios) — continuidad preventiva-compras emitible desde borrador.
	- `backend/db/productos.go` agrega `InventarioPlanReposicionOrdenEmitida` y `EmitirOrdenCompraDesdePlanReposicionBorrador` para emitir OC desde el borrador y persistirla en documentos de compras.
	- `backend/handlers/productos.go` agrega endpoint `POST /api/empresa/compras/plan_reposicion/emitir_orden`.
	- `backend/main.go` registra `/api/empresa/compras/plan_reposicion/emitir_orden` bajo permisos de compras.
	- `web/administrar_empresa/administrar_productos.html` agrega accion `Emitir orden` en el bloque de borrador (fase 10).
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`.
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./handlers ./db -count=1` (ok).
	- `get_errors` sobre backend/frontend modificado (ok).

## 2026-04-04
- Punto 5 (control de inventarios) — continuidad preventiva-compras ordenable por proveedor.
	- `backend/db/productos.go` agrega `InventarioPlanReposicionBorradorItem`, `InventarioPlanReposicionBorradorCompra` y `GetInventarioPlanReposicionBorradorByEmpresa` para generar borradores de orden por proveedor con detalle y totales.
	- `backend/handlers/productos.go` agrega endpoint `GET /api/empresa/inventario/plan_reposicion_borrador`.
	- `backend/main.go` registra `/api/empresa/inventario/plan_reposicion_borrador` bajo permisos de inventario.
	- `web/administrar_empresa/administrar_productos.html` agrega bloque `Borrador de orden de compra por proveedor (fase 10)` y accion `Borrador OC` desde consolidado fase 9.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`.
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./handlers ./db -count=1` (ok).
	- `get_errors` sobre backend/frontend modificado (ok).

## 2026-04-04
- Punto 5 (control de inventarios) — continuidad preventiva-compras consolidada por proveedor.
	- `backend/db/productos.go` agrega `InventarioPlanReposicionProveedorResumen` y `GetInventarioPlanReposicionResumenByEmpresa` para consolidar compra preventiva por proveedor.
	- `backend/handlers/productos.go` agrega endpoint `GET /api/empresa/inventario/plan_reposicion_resumen`.
	- `backend/main.go` registra `/api/empresa/inventario/plan_reposicion_resumen` bajo permisos de inventario.
	- `web/administrar_empresa/administrar_productos.html` agrega tabla `Consolidado de compra por proveedor (fase 9)` y filtro de items del plan por proveedor.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`.
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./handlers ./db -count=1` (ok).
	- `get_errors` sobre `web/administrar_empresa/administrar_productos.html` (ok).

## 2026-04-04
- Punto 5 (control de inventarios) — continuidad preventiva-compras con plan de reposicion por proveedor.
	- `backend/db/productos.go` agrega `InventarioPlanReposicionItem` y `GetInventarioPlanReposicionByEmpresa` para consolidar sugerencias por proveedor con costo estimado.
	- `backend/handlers/productos.go` agrega endpoint `GET /api/empresa/inventario/plan_reposicion` con validaciones operativas.
	- `backend/main.go` registra `/api/empresa/inventario/plan_reposicion` bajo permisos de inventario.
	- `web/administrar_empresa/administrar_productos.html` agrega tabla `Plan de reposicion por proveedor (fase 8)` con resumen de costo estimado y accion `Preparar`.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`.
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./handlers ./db -count=1` (ok).
	- `get_errors` sobre `web/administrar_empresa/administrar_productos.html` (ok).

## 2026-04-04
- Punto 5 (control de inventarios) — continuidad preventiva con proyeccion de quiebre.
	- `backend/db/productos.go` agrega `InventarioProyeccionQuiebre` y `GetInventarioProyeccionQuiebreByEmpresa` para estimar consumo diario, cobertura y sugerido de reposicion por producto/bodega.
	- `backend/handlers/productos.go` agrega endpoint `GET /api/empresa/inventario/proyeccion_quiebre` con validacion de `dias_ventana`, `bodega_id`, `limit` y `offset`.
	- `backend/main.go` registra `/api/empresa/inventario/proyeccion_quiebre` bajo permisos de inventario.
	- `web/administrar_empresa/administrar_productos.html` agrega tabla `Proyeccion de quiebre (preventiva)` y accion `Preparar` para reposicion preventiva guiada.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`.
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./handlers ./db -count=1` (ok).
	- `get_errors` sobre `web/administrar_empresa/administrar_productos.html` (ok).

## 2026-04-04
- Punto 5 (control de inventarios) — continuidad operativa-analitica con balance por bodega.
	- `backend/db/productos.go` agrega `InventarioBalanceBodega` y `GetInventarioBalanceBodegasByEmpresa` para consolidar entradas/salidas/traslados/neto por bodega en rango.
	- `backend/handlers/productos.go` agrega endpoint `GET /api/empresa/inventario/balance_bodegas` con validacion de fechas y filtros por bodega/rango.
	- `backend/main.go` registra `/api/empresa/inventario/balance_bodegas` bajo permisos de inventario.
	- `web/administrar_empresa/administrar_productos.html` agrega tabla `Balance por bodega` y contexto de neto acumulado sincronizado con filtros del kardex.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`.
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./handlers ./db -count=1` (ok).
	- `get_errors` sobre `web/administrar_empresa/administrar_productos.html` (ok).

## 2026-04-04
- Punto 5 (control de inventarios) — continuidad analitica con tendencia diaria.
	- `backend/db/productos.go` agrega `InventarioTendenciaDia` y `GetInventarioTendenciaByEmpresa` para serie diaria por empresa con filtros por bodega/rango.
	- `backend/handlers/productos.go` agrega endpoint `GET /api/empresa/inventario/tendencia` con validacion de fechas y ventana por `dias`.
	- `backend/main.go` registra `/api/empresa/inventario/tendencia` bajo permisos de inventario.
	- `web/administrar_empresa/administrar_productos.html` agrega tabla `Tendencia diaria inventario` y contexto de neto acumulado/eventos sincronizado con filtros del kardex.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`.
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./handlers ./db -count=1` (ok).
	- `get_errors` sobre `web/administrar_empresa/administrar_productos.html` (ok).

## 2026-04-04
- Punto 5 (control de inventarios) — continuidad operacional en panel de productos.
	- `web/administrar_empresa/administrar_productos.html` agrega:
		- bloque `Top productos críticos (déficit)` alimentado desde alertas de inventario,
		- priorización de críticos por `sin_stock` y mayor déficit,
		- acción `Preparar reposición` para precargar ajuste de inventario con producto, bodega y cantidad sugerida.
- Validacion tecnica:
	- `get_errors` sobre `web/administrar_empresa/administrar_productos.html` (ok).

## 2026-04-04
- Punto 5 (control de inventarios) — continuidad KPI operativo en panel de productos.
	- `backend/db/productos.go` agrega `InventarioResumen` y `GetInventarioResumenByEmpresa` para consolidar existencias, alertas y movimientos por rango.
	- `backend/handlers/productos.go` agrega endpoint `GET /api/empresa/inventario/resumen` con validacion de fechas `YYYY-MM-DD`.
	- `backend/main.go` registra `/api/empresa/inventario/resumen` bajo permisos de inventario.
	- `web/administrar_empresa/administrar_productos.html` agrega KPI visibles de inventario e integra consumo del resumen segun rango del kardex.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`,
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./handlers ./db -count=1` (ok).
	- `get_errors` sobre `web/administrar_empresa/administrar_productos.html` (ok).

## 2026-04-04
- Punto 5 (control de inventarios) — continuidad UI operativa en panel de productos.
	- `web/administrar_empresa/administrar_productos.html` agrega:
		- filtro por bodega para alertas de quiebre,
		- filtros de kardex por bodega, tipo y rango de fechas,
		- acciones `Filtrar` y `Limpiar` en ambos bloques de consulta.
	- Se actualiza documentacion asociada en plan maestro y estructura tecnica.
- Validacion tecnica:
	- `get_errors` sobre `web/administrar_empresa/administrar_productos.html` (ok).

## 2026-04-04
- Punto 5 (control de inventarios) — inicio tecnico: kardex operativo + reglas de stock + alertas de quiebre por bodega.
	- `backend/db/productos.go`:
		- valida `stock_minimo/stock_maximo` en creacion y edicion de productos,
		- agrega `GetAlertasQuiebreByEmpresa`,
		- amplía `GetMovimientosByEmpresa` con filtros `bodega_id`, `tipo`, `desde`, `hasta`.
	- `backend/handlers/productos.go`:
		- nuevo endpoint `GET /api/empresa/inventario/alertas`,
		- compatibilidad `action=alertas|alertas_quiebre|quiebre` en existencias,
		- filtros de kardex + validacion de fechas `YYYY-MM-DD` en movimientos.
	- `backend/main.go` registra `/api/empresa/inventario/alertas` bajo permisos de inventario.
	- `web/administrar_empresa/administrar_productos.html` agrega tabla de alertas de quiebre por bodega.
	- `documentos/descripcion_del_proyecto` actualiza la descripcion de inventario con alertas de quiebre y kardex filtrable.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`,
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `runTests` en archivos de prueba modificados (ok).
	- `go test ./handlers ./db -count=1` en `backend` (ok).

## 2026-04-04
- Punto 3 (permisos y seguridad) — continuidad operativa: catalogo frontend por rol + regresion endpoints sin wrapper.
	- `web/js/administrar_empresa.js` agrega catalogo de permisos por enlace y aplica ocultamiento de opciones no autorizadas segun rol autenticado (`GET /me`).
	- Se agrega fallback de navegacion en iframe cuando la ultima pagina guardada no es visible para el rol actual.
	- `backend/handlers/auth_users_carritos_test.go` agrega regresiones de alcance por `empresa_id` para:
		- `POST /api/empresa/usuarios/login`.
		- `POST /api/empresa/usuarios/establecer_password`.
	- `backend/handlers/chat_con_inteligencia_artificial_controller_test.go` agrega regresion de alcance por cuenta Google en `ModelosHandler`.
	- Se actualiza documentacion tecnica en:
		- `documentos/diagramas/diagrama_roles_permisos.md`.
		- `documentos/diagramas/estructura_del_codigo.md`.
- Validacion tecnica:
	- `runTests` sobre `backend/handlers/auth_users_carritos_test.go` y `backend/handlers/chat_con_inteligencia_artificial_controller_test.go`.
	- resultado: 14 pruebas aprobadas, 0 fallidas.
	- `get_errors` sobre `web/js/administrar_empresa.js`: sin errores.

## 2026-04-04
- Punto 3 (permisos y seguridad) — consolidacion documental endpoint/rol y checklist UAT:
	- `documentos/matriz_roles_permisos_pos_multiempresa.md` agrega matriz final endpoint/rol alineada con wrappers reales y reglas por accion.
	- Se documentan endpoints fuera de wrapper con control alterno por handler/cuenta Google.
	- Se agrega checklist UAT de punto 3 con evidencia automatizada.
	- `documentos/plan_maestro_pos_multiempresa_14_puntos.md` agrega seccion de consolidacion con estado operativo y pendientes de cierre total.
- Validacion tecnica:
	- `runTests` sobre `backend/handlers/empresa_permisos_test.go` y `backend/handlers/auditoria_empresa_test.go`.
	- resultado: 25 pruebas aprobadas, 0 fallidas.

## 2026-04-04
- Ajuste editorial de consistencia documental (plan maestro):
	- `documentos/plan_maestro_pos_multiempresa_14_puntos.md` corrige `Backlog inmediato` para reflejar cierre real de Punto 1 y Punto 2.
	- El backlog siguiente queda enfocado en Punto 3 (permisos y seguridad) y Punto 5 (control de inventarios).
- Validacion tecnica:
	- cambio documental (sin ejecucion de pruebas automatizadas).

## 2026-04-04
- Punto 1 + Punto 2 (plan maestro) — cierre de backlog inmediato con formalizacion tecnica documental.
	- `documentos/matriz_kpi_pos_multiempresa.md` se actualiza a formato formal con:
		- formula implementada por KPI,
		- endpoint canonico de lectura/exportacion,
		- tablas fuente reales por metrica.
	- Se crea `documentos/matriz_entidades_multiempresa_aislamiento.md` con matriz de aislamiento por endpoint:
		- llave primaria `empresa_id`,
		- llaves secundarias por recurso,
		- mecanismo de control de alcance (middleware o validacion interna).
	- `documentos/plan_maestro_pos_multiempresa_14_puntos.md` marca Punto 1 y Punto 2 como `completado`.
- Validacion tecnica:
	- cambio documental (sin ejecucion de pruebas automatizadas).

## 2026-04-04
- Punto 11 (reportes financieros) — continuidad de backlog inmediato: exportacion unificada del tablero por rango.
	- `backend/handlers/finanzas.go` agrega `action=tablero_export` en `GET /api/empresa/finanzas/movimientos` con:
		- `format=json` para payload unificado del tablero,
		- `format=csv` para matriz unificada por bloque/metrica/valor.
	- La exportacion integra bloques `estado_resultados` y `balance_general` junto con KPI operativos/financieros/contables.
	- `web/administrar_empresa/reportes.html` incorpora botones:
		- `Exportar tablero CSV`,
		- `Exportar tablero JSON`.
	- `backend/handlers/eventos_contables_modulos_test.go` agrega `TestEmpresaFinanzasTableroResumenExportHandler`.
- Validacion tecnica:
	- `go test ./handlers -run "TestEmpresaFinanzasTableroResumenHandler|TestEmpresaFinanzasTableroResumenExportHandler|TestEmpresaFinanzasAsientosContablesHandlerConciliacionPeriodo" -count=1` (ok).
	- `go test ./handlers -count=1` (ok).
	- `go test ./db -count=1` (ok).

## 2026-04-04
- Punto 10 (modulo contable integrado) — continuidad de backlog inmediato: vista de conciliacion por periodo (eventos vs asientos).
	- `backend/db/eventos_contables.go` agrega modelos y funcion `GetEmpresaConciliacionContablePorPeriodo` para consolidar por periodo:
		- eventos totales/procesados/pendientes/con error,
		- asientos generados,
		- desfase de conteo y desfase de monto,
		- estado de conciliacion por periodo.
	- `backend/handlers/finanzas.go` agrega `GET /api/empresa/finanzas/asientos_contables?action=conciliacion_periodo|conciliacion`.
	- `web/administrar_empresa/finanzas.html` incorpora vista de conciliacion con filtros, KPIs y tabla comparativa por periodo.
	- `backend/db/eventos_contables_test.go` agrega prueba de conciliacion por periodo.
	- `backend/handlers/eventos_contables_modulos_test.go` agrega prueba del endpoint de conciliacion.
- Validacion tecnica:
	- `go test ./db -run "EventosContables|ConPolitica|Conciliacion" -count=1` (ok).
	- `go test ./handlers -run "AsientosContablesHandler|ConciliacionPeriodo" -count=1` (ok).
	- `go test ./db -count=1` (ok).
	- `go test ./handlers -count=1` (ok).

## 2026-04-04
- Punto 10 (modulo contable integrado) — continuidad de backlog inmediato: ejecucion automatica por lotes de asientos.
	- `backend/db/eventos_contables.go` agrega:
		- `ProcessEmpresaEventosContablesPendientesConPolitica` con soporte de `max_reintentos`,
		- `RunEmpresaAsientosContablesWorkerCycle`,
		- `StartEmpresaAsientosContablesWorker`.
	- `backend/main.go` integra worker automatico de asientos con politica configurable por entorno:
		- `ASIENTOS_WORKER_INTERVAL_MINUTES`,
		- `ASIENTOS_WORKER_BATCH_SIZE`,
		- `ASIENTOS_WORKER_MAX_RETRIES`.
	- `backend/handlers/finanzas.go` permite `max_reintentos` opcional en proceso manual de `/api/empresa/finanzas/asientos_contables?action=procesar_asientos`.
	- `backend/db/eventos_contables_test.go` agrega prueba de politica de reintentos.
	- `backend/handlers/eventos_contables_modulos_test.go` agrega validacion `400` para `max_reintentos` invalido y cobertura del parametro.
- Validacion tecnica:
	- `go test ./db -run "EventosContables|ConPolitica|Asientos" -count=1` (ok).
	- `go test ./handlers -run "AsientosContablesHandler|FinanzasAsientos" -count=1` (ok).
	- `go test ./handlers -count=1` (ok).
	- `go test ./db -count=1` (ok).

## 2026-04-04
- Punto 15 (auditoria por empresa) — continuacion de backlog inmediato 1 y 2:
	- `backend/db/auditoria_empresa.go` agrega filtros avanzados de consulta por `recurso_id` y `codigo_http` en `ListEmpresaAuditoriaEventos`.
	- `backend/handlers/auditoria_empresa.go` valida y expone nuevos filtros en `GET /api/empresa/auditoria/eventos`:
		- `recurso_id`.
		- `codigo_http`.
	- `web/administrar_empresa/auditoria.html` incorpora:
		- filtros avanzados por `codigo_http` y `recurso_id`,
		- exportacion de resultados filtrados a `CSV` y `JSON`.
	- `backend/db/auditoria_empresa_test.go` fortalece cobertura de listado con filtros avanzados.
	- `backend/handlers/auditoria_empresa_test.go` agrega `TestEmpresaAuditoriaEventosHandlerFiltrosAvanzados` para contrato HTTP y validacion de parametros invalidos.
- Validacion tecnica:
	- `go test ./db -run "Auditoria" -count=1` (ok).
	- `go test ./handlers -run "Auditoria" -count=1` (ok).
	- `go test ./handlers -count=1` (ok).
	- `go test ./db -count=1` (ok).

## 2026-04-04
- Punto 15 (auditoria por empresa) — continuacion de backlog 1, 2 y 3:
	- `backend/handlers/empresa_permisos.go` refuerza clasificacion de acciones criticas en `ventas`, `compras` y `facturacion` (alias operativos de aprobacion/eliminacion).
	- `backend/handlers/auditoria_empresa.go` amplia metadata de trazabilidad para recursos de ventas/compras/facturacion (`carrito_id`, `proveedor_id`, `entidad_id`, `documento_codigo`).
	- `backend/handlers/auditoria_empresa_test.go` agrega pruebas de registro automatico de auditoria en acciones criticas de:
		- ventas (`action=cerrar`),
		- compras (`action=emitir_orden`),
		- facturacion (`action=emitir`).
	- `web/administrar_empresa/auditoria.html` agrega vista de consulta filtrable y retencion manual para auditoria por empresa.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` agregan acceso del menu lateral a la nueva vista `Auditoria`.
	- `backend/db/auditoria_empresa.go` agrega:
		- purga automatica por expiracion (`PurgeExpiredEmpresaAuditoriaEventos`),
		- worker programado (`StartEmpresaAuditoriaRetentionWorker`),
		- calculo de `fecha_expiracion` alineado a `fecha_evento` cuando se provee.
	- `backend/main.go` arranca worker de retencion automatica de auditoria (intervalo 12h).
	- `backend/db/auditoria_empresa_test.go` agrega prueba de purga automatica por expiracion.
- Validacion tecnica:
	- `go test ./handlers -run "Auditoria|WithEmpresa(Ventas|Compras|Facturacion|Finanzas)Permissions" -count=1` (ok).
	- `go test ./db -run "Auditoria" -count=1` (ok).
	- `go test ./handlers -count=1` (ok).
	- `go test ./db -count=1` (ok).

## 2026-04-04
- Punto 15 (auditoria por empresa) — implementacion base minima:
	- `backend/db/auditoria_empresa.go` agrega tabla `empresa_auditoria_eventos`, filtros de consulta y purga por retencion.
	- `backend/handlers/auditoria_empresa.go` agrega endpoint protegido:
		- `GET /api/empresa/auditoria/eventos`.
		- `PUT/POST /api/empresa/auditoria/eventos?action=retener|purgar`.
	- `backend/handlers/empresa_permisos.go` integra registro automatico no bloqueante para acciones criticas (`C/U/D/A`).
	- `backend/main.go` integra `EnsureEmpresaAuditoriaSchema`, migracion `2026-04-04-011-auditoria-empresa` y ruta de auditoria.
	- Pruebas nuevas: `backend/db/auditoria_empresa_test.go` y `backend/handlers/auditoria_empresa_test.go`.
- Validacion tecnica:
	- `go test ./db -run "Auditoria|EventosContables|ReportesTableroResumen" -count=1` (ok).
	- `go test ./handlers -run "Auditoria|AsientosContables|WithEmpresaFinanzasPermissions" -count=1` (ok).
	- `go test ./handlers -count=1` (ok).
	- `go test ./db -count=1` (ok).

## 2026-04-04
- Plan maestro POS multiempresa:
	- `documentos/plan_maestro_pos_multiempresa_14_puntos.md` se actualiza de 14 a 15 puntos.
	- Se incorpora el nuevo `Punto 15: Modulo de auditoria por empresa` con alcance, entregables iniciales, backlog y criterio de avance.
	- `documentos/descripcion_del_proyecto` se alinea para referenciar el plan de 15 puntos.
- Validacion tecnica:
	- cambio documental (sin cambios de codigo ni ejecucion de pruebas adicionales).

## 2026-04-04
- Punto 10 + Punto 11 (continuacion de backlog 1 y 2):
	- `backend/db/eventos_contables.go` amplía `empresa_eventos_contables` con metadatos de procesamiento (`intentos_procesamiento`, `fecha_ultimo_intento`, `error_procesamiento`, `asiento_contable_id`) y crea tabla canonica `empresa_asientos_contables` con hash de idempotencia.
	- `backend/handlers/finanzas.go` agrega `EmpresaFinanzasAsientosContablesHandler`:
		- `GET /api/empresa/finanzas/asientos_contables` para consulta,
		- `POST/PUT action=procesar_asientos|procesar` para procesamiento manual por lote.
	- `backend/handlers/empresa_permisos.go` clasifica `action=procesar_asientos` como accion de aprobacion en finanzas.
	- `backend/main.go` publica `/api/empresa/finanzas/asientos_contables` y registra migracion `2026-04-04-010-asientos-canonicos`.
	- `backend/db/finanzas.go` integra en el tablero los bloques `estado_resultados` y `balance_general`, junto con KPI contables de asientos (`asientos_generados`, `asientos_monto_total`).
	- `web/administrar_empresa/reportes.html` incorpora visualizacion de utilidad operacional, activos/pasivos/patrimonio, resultado del ejercicio y cuadre.
	- `web/administrar_empresa/finanzas.html` añade accion manual `Procesar eventos contables`.
	- Cobertura de pruebas nueva/extendida en `backend/db/eventos_contables_test.go`, `backend/db/finanzas_test.go`, `backend/handlers/eventos_contables_modulos_test.go` y `backend/handlers/empresa_permisos_test.go`.
- Validacion tecnica:
	- `go test ./db -run "EventosContables|ReportesTableroResumen" -count=1` (ok).
	- `go test ./handlers -run "AsientosContables|TableroResumen|WithEmpresaFinanzasPermissions" -count=1` (ok).
	- `go test ./handlers -count=1` (ok).
	- `go test ./db -count=1` (ok).

## 2026-04-04
- Punto 12 + Punto 10 (continuacion de backlog 1 y 2):
	- `backend/handlers/empresa_permisos_test.go` agrega pruebas UAT por rol para `PUT action=aprobar` en `cierres_caja`:
		- rechazo para `cajero`,
		- rechazo para `supervisor_sucursal`,
		- aprobacion permitida para `admin_empresa`.
	- `documentos/matriz_roles_permisos_pos_multiempresa.md` agrega matriz UAT de cierres con casos por rol y transiciones de estado.
	- `documentos/plan_maestro_pos_multiempresa_14_puntos.md` define estrategia de procesamiento de asientos sobre `empresa_eventos_contables` y referencias canonicas documentales (`entidad_id`).
- Validacion tecnica:
	- `go test ./handlers -run "TestWithEmpresaFinanzasPermissions(DeniesCajeroAprobarCierreCaja|DeniesSupervisorAprobarCierreCaja|AllowsAdminAprobarCierreCaja)" -count=1` (ok).

## 2026-04-04
- Punto 12 (cierres de caja) — continuacion con UI operativa en panel empresa:
	- `web/administrar_empresa/finanzas.html` integra modulo visual de cierres de caja por sucursal con:
		- formulario de apertura/actualizacion,
		- calculo de `caja_teorica` y `diferencia_caja`,
		- filtros por sucursal/caja/estado/fecha,
		- tabla de acciones (`cerrar`, `reabrir`, `aprobar`, `anular`, `activar/desactivar`, `eliminar`).
	- La vista queda conectada al endpoint existente `GET/POST/PUT/DELETE /api/empresa/finanzas/cierres_caja`.
- Validacion tecnica:
	- `get_errors` sobre `web/administrar_empresa/finanzas.html` (ok).

## 2026-04-04
- Punto 12 (cierres de caja) — inicio de flujo operativo por sucursal:
	- `backend/db/finanzas.go` agrega `empresa_cierres_caja` con soporte de apertura, arqueo, cierre, reapertura, aprobacion y anulacion.
	- `backend/handlers/finanzas.go` incorpora `GET/POST/PUT/DELETE /api/empresa/finanzas/cierres_caja`.
	- `backend/main.go` publica la ruta de cierres de caja y registra migracion `2026-04-04-009-cierres-caja`.
	- `backend/handlers/empresa_permisos.go` trata `action=aprobar` en finanzas como accion `A`.
	- Pruebas nuevas:
		- `backend/db/finanzas_test.go`: `TestEmpresaCierresCajaFlow`.
		- `backend/handlers/eventos_contables_modulos_test.go`: `TestEmpresaFinanzasCierresCajaHandler`.
- Validacion tecnica:
	- `go test ./db -run "TestEmpresaCierresCajaFlow|TestGetEmpresaReportesTableroResumen|TestEmpresaFinanzas" -count=1` (ok).
	- `go test ./handlers -run "TestEmpresaFinanzasCierresCajaHandler|TestEmpresaFinanzasTableroResumenHandler|TestEmpresaFinanzasEmiteEventosContables" -count=1` (ok).
	- `go test ./auth ./db ./handlers ./metrics ./utils -count=1` (ok).

## 2026-04-04
- Punto 11 (reportes financieros) — inicio de tablero minimo financiero-operativo:
	- `backend/db/finanzas.go` agrega `GetEmpresaReportesTableroResumen` con KPI consolidados:
		- operativos (ventas/ticket/clientes/productos/compras),
		- financieros (ingresos/egresos/balance/periodos),
		- contables (eventos y documentos activos).
	- `backend/handlers/finanzas.go` extiende `GET /api/empresa/finanzas/movimientos` con `action=tablero|dashboard|resumen_kpi`.
	- `web/administrar_empresa/reportes.html` incorpora KPI financieros y contables en la misma vista de reportes.
	- Pruebas nuevas:
		- `backend/db/finanzas_test.go`: `TestGetEmpresaReportesTableroResumen`.
		- `backend/handlers/eventos_contables_modulos_test.go`: `TestEmpresaFinanzasTableroResumenHandler`.
- Validacion tecnica:
	- `go test ./db -run "TestGetEmpresaReportesTableroResumen|TestEmpresaFinanzas" -count=1` (ok).
	- `go test ./handlers -run "TestEmpresaFinanzasTableroResumenHandler|TestEmpresaFinanzasEmiteEventosContables" -count=1` (ok).
	- `go test ./... -count=1` en `backend` (ok).

## 2026-04-04
- Punto 8 + Punto 9 + Punto 10 (facturacion/compras) — persistencia canonica de documentos transaccionales para `entidad_id`:
	- Se agrega `backend/db/documentos_transaccionales.go` con tablas y APIs de upsert/lectura para:
		- `empresa_facturacion_documentos`.
		- `empresa_compras_documentos`.
	- `backend/main.go` integra:
		- `EnsureEmpresaDocumentosTransaccionalesSchema`.
		- migracion `2026-04-04-008-documentos-transaccionales`.
	- `backend/handlers/facturacion_electronica.go` y `backend/handlers/productos.go` ahora:
		- consultan estado documental persistido por `documento_codigo`,
		- aplican transicion de ciclo sobre estado canonico,
		- persisten el nuevo estado en tabla de negocio,
		- emiten evento contable usando `entidad_id` canonico (ID persistido en tabla documental).
	- Se agrega `backend/db/documentos_transaccionales_test.go` y se amplian aserciones en `backend/handlers/eventos_contables_modulos_test.go` para verificar estabilidad de `entidad_id` en el ciclo documental.
- Validacion tecnica:
	- `go test ./handlers -run "FacturacionTransaccionalEmiteEventosContables|ComprasTransaccionalEmiteEventosContables|FacturacionElectronicaEmiteEventoContable|ProveedoresEmiteEventoContableCompras|FinanzasEmiteEventosContables" -count=1` (ok).
	- `go test ./auth ./db ./handlers ./metrics ./utils` (ok).
	- `go test ./...` en `backend` (ok).

## 2026-04-04
- Punto 8 + Punto 9 + Punto 10 (facturacion/compras) — estandarizacion de estados en ciclo documental transaccional:
	- Se agrega `backend/handlers/documentos_lifecycle.go` con reglas de transicion por accion y estado previo para facturacion/compras.
	- `backend/handlers/facturacion_electronica.go` ahora valida `estado_actual` en `emitir/anular/nota_credito`, devuelve `409` en conflictos y responde `estado_anterior`/`estado_nuevo` cuando la transicion es valida.
	- `backend/handlers/productos.go` (`EmpresaProveedoresHandler`) aplica validacion equivalente para `emitir_orden/recepcionar_compra/contabilizar_compra`.
	- `backend/handlers/eventos_contables_modulos_test.go` amplía cobertura con pruebas de transiciones invalidas para facturacion y compras.
- Validacion tecnica:
	- `runTests` sobre `backend/handlers/eventos_contables_modulos_test.go` (ok).
	- `go test ./...` en `backend` (ok).

## 2026-04-04
- Punto 8 + Punto 9 + Punto 10 (facturacion/compras) — eventos transaccionales de factura y orden:
	- `backend/handlers/facturacion_electronica.go` agrega acciones transaccionales:
		- `action=emitir` -> `factura_emitida`.
		- `action=anular` -> `factura_anulada`.
		- `action=nota_credito|emitir_nota_credito` -> `nota_credito_emitida`.
	- `backend/handlers/productos.go` (`EmpresaProveedoresHandler`) agrega acciones transaccionales:
		- `action=emitir|emitir_orden` -> `orden_compra_emitida`.
		- `action=recepcionar|recepcionar_compra` -> `compra_recepcionada`.
		- `action=contabilizar|contabilizar_compra` -> `compra_contabilizada`.
	- `backend/handlers/empresa_permisos.go` amplía mapeo de acciones de permisos para compras/facturacion.
	- `backend/handlers/eventos_contables_modulos_test.go` agrega pruebas de emisiones transaccionales de factura/orden.
- Validacion tecnica:
	- `go test ./handlers -run "FacturacionTransaccionalEmiteEventosContables|ComprasTransaccionalEmiteEventosContables|FacturacionElectronicaEmiteEventoContable|ProveedoresEmiteEventoContableCompras|FinanzasEmiteEventosContables" -count=1` (ok).
	- `go test ./...` en `backend` (ok).

## 2026-04-04
- Punto 8 + Punto 9 + Punto 10 (facturacion/compras/finanzas) — extension de emision de eventos contables por modulo:
	- Se agrega `backend/handlers/eventos_contables.go` para registro no bloqueante y reutilizable de eventos contables en handlers.
	- Se amplia `backend/db/eventos_contables.go` con eventos operativos de:
		- `facturacion`: `configuracion_facturacion_actualizada`.
		- `compras`: `proveedor_registrado`, `proveedor_actualizado`, `proveedor_activado`, `proveedor_desactivado`, `proveedor_eliminado`.
	- Se integra emision en:
		- `backend/handlers/facturacion_electronica.go`.
		- `backend/handlers/productos.go` (proveedores).
		- `backend/handlers/finanzas.go` (movimientos y periodos).
	- `backend/handlers/carritos_compras.go` migra a helper comun para consistencia del registro contable.
	- Se agregan pruebas en `backend/handlers/eventos_contables_modulos_test.go` para validar emision en facturacion, compras y finanzas.
- Validacion tecnica:
	- `go test ./db -run "EventosContables" -count=1` (ok).
	- `go test ./handlers -run "FacturacionElectronicaEmiteEventoContable|ProveedoresEmiteEventoContableCompras|FinanzasEmiteEventosContables|CarritosCompraAndItemsFlow" -count=1` (ok).
	- `go test ./...` en `backend` (ok).

## 2026-04-04
- Punto 4 + Punto 10 (gestion de ventas + modulo contable integrado) — contrato de eventos contables por modulo:
	- Se agrega `backend/db/eventos_contables.go` con contrato base de eventos para `ventas`, `facturacion`, `compras` y `finanzas`.
	- Se crea tabla `empresa_eventos_contables` en `empresas.db` para registrar trazabilidad contable por empresa (`modulo`, `evento`, `entidad`, `documento`, `periodo_contable`, `monto`, `payload_json`, `procesado`).
	- Se integra bootstrap en `backend/main.go`:
		- `EnsureEmpresaEventosContablesSchema`.
		- migracion `2026-04-04-007-eventos-contables`.
	- Se actualiza `backend/handlers/carritos_compras.go` para emitir eventos contables en transiciones de venta de carritos (`venta_sesion_activada`, `venta_activada`, `venta_suspendida`, `venta_cerrada`, `venta_reabierta`, `venta_pagada`).
	- Se agregan pruebas:
		- `backend/db/eventos_contables_test.go`.
		- `backend/handlers/auth_users_carritos_test.go` (validacion de emision de `venta_pagada`).
- Validacion tecnica:
	- `go test ./db -run "EventosContables|CarritoEstadoVentaLifecycle|Finanzas" -count=1` (ok).
	- `go test ./handlers -run "EmpresaCarritosCompra|CarritosCompraAndItemsFlow" -count=1` (ok).
	- `go test ./...` en `backend` (ok).

## 2026-04-04
- Punto 4 (gestion de ventas) — formalizacion de transiciones del ciclo de venta en carritos:
	- `backend/handlers/carritos_compras.go` ahora valida transiciones por accion y estado actual del carrito.
	- Se agregan respuestas de control para integridad de flujo:
		- `404` para carrito inexistente,
		- `409` para transiciones no permitidas (doble pago, reabrir pagada, activar estacion pagada sin `reset_items=1`, etc.).
	- Se agregan pruebas en `backend/handlers/auth_users_carritos_test.go`:
		- `TestEmpresaCarritosCompraRejectsDoublePago`.
		- `TestEmpresaCarritosCompraRejectsReabrirVentaPagada`.
		- `TestEmpresaCarritosCompraRejectsActivarEstacionPagadaSinReset`.
- Validacion tecnica:
	- `go test ./handlers -run "Carritos|EmpresaCarritosCompra" -count=1` (ok).
	- `go test ./...` en `backend` (ok).

## 2026-04-04
- Cierre validado del punto 3 (permisos y seguridad) con pruebas de endpoints protegidos recien incorporados:
	- `backend/handlers/empresa_permisos_test.go` agrega:
		- `TestWithEmpresaInventarioPermissionsDeniesCajeroWriteGPS`.
		- `TestWithEmpresaVentasPermissionsAllowsCajeroChatAdjuntoMultipart`.
		- `TestWithEmpresaVentasPermissionsRejectsChatAdjuntoWithoutAuth`.
	- Se valida control por rol en GPS, extraccion de `empresa_id` en `multipart/form-data` para adjuntos de chat y rechazo `401` sin autenticacion.
- Inicio del punto 4 (gestion de ventas):
	- `backend/db/carritos_compras.go` incorpora `estado_venta` derivado en el modelo `CarritoCompra` para estandarizar ciclo de vida de venta:
		- `venta_abierta`,
		- `venta_cerrada`,
		- `venta_pagada`,
		- `venta_suspendida`.
	- `backend/handlers/carritos_compras.go` expone `estado_venta` en acciones operativas (`activar_estacion`, `pagar_estacion`, `activar/desactivar`, `cerrar/reabrir`).
	- Se amplian pruebas en:
		- `backend/handlers/auth_users_carritos_test.go`.
		- `backend/db/carritos_inventario_test.go`.
- Validacion tecnica de esta iteracion:
	- `runTests` sobre archivos de pruebas modificados (ok).
	- `go test ./...` en `backend` (ok).

## 2026-04-04
- Continuacion del punto 3 del plan maestro (permisos y seguridad) con cierre de rutas operativas pendientes:
	- `backend/handlers/empresa_permisos.go` agrega modulo `seguridad` y wrapper `WithEmpresaSeguridadPermissions`.
	- `backend/main.go` amplía middleware en rutas:
		- seguridad: `/api/empresa/usuarios`, `/api/empresa/configuracion_avanzada`, `/api/empresa/roles_de_usuario`.
		- inventario: `/api/empresa/productos/imagen`, `/api/empresa/ubicacion_gps/dispositivos`, `/api/empresa/ubicacion_gps/recorridos`.
		- colaboracion operativa (politica ventas): `/api/empresa/chat_tareas/conversaciones`, `/api/empresa/chat_tareas/participantes`, `/api/empresa/chat_tareas/mensajes`, `/api/empresa/chat_tareas/mensajes/adjunto`, `/api/empresa/chat_tareas/tareas`.
	- `backend/handlers/empresa_permisos_test.go` agrega cobertura para modulo seguridad:
		- `TestWithEmpresaSeguridadPermissionsDeniesSupervisorWrite`.
		- `TestWithEmpresaSeguridadPermissionsAllowsSupervisorRead`.
		- `TestWithEmpresaSeguridadPermissionsAllowsAdminApprove`.
	- Validacion tecnica: `go test ./handlers -run "WithEmpresa|ConsultarHandlerRejectsEmpresaFueraDeAlcance" -count=1` (ok) y `go test ./...` (ok).

## 2026-04-04
- Continuacion del punto 3 del plan maestro (permisos y seguridad):
	- `backend/handlers/empresa_permisos.go` amplía modulos de autorizacion para `clientes`, `compras` y `facturacion`.
	- Se agregan wrappers: `WithEmpresaClientesPermissions`, `WithEmpresaComprasPermissions`, `WithEmpresaFacturacionPermissions`.
	- `backend/main.go` aplica middleware en rutas: `/api/empresa/clientes`, `/api/empresa/proveedores`, `/api/empresa/facturacion_electronica`, `/api/empresa/facturacion_electronica/pais_detectado`, y `/api/empresa/servicios` (politica inventario).
	- Se amplian pruebas en `backend/handlers/empresa_permisos_test.go` para cobertura de los modulos nuevos.
	- Validacion tecnica: `go test ./handlers -run "WithEmpresa|ConsultarHandlerRejectsEmpresaFueraDeAlcance" -count=1` (ok) y `go test ./...` (ok).

## 2026-04-04
- Se registra nueva credencial Gemini cifrada en configuración avanzada (`ai.model.google.gemini_2_0_flash.api_key` en `superadministrador.db`).
- Se valida consumo de Gemini con la nueva credencial: respuesta del proveedor `429` por cuota excedida (sin error de credencial/servicio bloqueado).
- Se verifica la presencia de la tarjeta de Gemini en `web/super/configuracion_avanzada.html` y se corrige un bloque JavaScript en la carga de estado para mantener consistencia de la vista.
- Se agrega prueba de seguridad de alcance por empresa para chat IA en `backend/handlers/chat_con_inteligencia_artificial_controller_test.go`:
	- `TestConsultarHandlerRejectsEmpresaFueraDeAlcance`.
	- Validación: `go test ./handlers -run "TestConsultarHandlerRejectsEmpresaFueraDeAlcance|TestModelosHandlerRequiresGoogleAccount|TestModelosHandlerReturnsPreferredModelForGoogleAccount" -count=1` (ok).

## 2026-04-04
- Chat IA empresarial migrado a Gemini-only:
	- `backend/handlers/chat_con_inteligencia_artificial_controller.go` ahora integra Google Gemini (`generateContent`) y elimina dependencias de OpenAI/DeepSeek/Hugging Face para este módulo.
	- El catálogo y la configuración de credenciales IA quedan en un único modelo soportado: `google:gemini-2.0-flash` (`GEMINI_API_KEY`).
	- `web/super/configuracion_avanzada.html` simplifica la tarjeta IA a una sola credencial Gemini con trazabilidad por cuenta Google.
	- `web/administrar_empresa/chat_con_inteligencia_artificial.html` se rediseña con experiencia visual tipo Gemini, chips de contexto y flujo explícito de autenticación Google.
	- Pruebas ajustadas y validadas: `go test ./auth ./db ./handlers ./metrics ./utils` (ok) en `backend`.
- Se agrega gestión de credenciales IA en `super/configuracion_avanzada.html` para 5 modelos populares con plan gratuito limitado:
	- OpenAI GPT-4o mini,
	- OpenAI GPT-4.1 mini,
	- DeepSeek Chat,
	- DeepSeek Reasoner,
	- Meta Llama 3.1 8B Instruct (Hugging Face).
- Se crea endpoint `GET/PUT /super/api/config/ai` en backend para guardar/consultar credenciales con registro de la cuenta Google logueada que realiza cambios.
- El módulo `chat_con_inteligencia_artificial` ahora resuelve credenciales en este orden:
	- configuración guardada por modelo,
	- configuración por proveedor,
	- variable de entorno.
- Validación técnica ejecutada:
	- `go test ./handlers -run "AIModelsConfigHandler|Chat|ModelosHandler" -count=1` (ok).
	- `go test ./...` en `backend` (ok).
- Se implementa la primera fase tecnica del punto 3 (permisos y seguridad) con middleware de autorizacion por rol + alcance de empresa:
	- nuevo `backend/handlers/empresa_permisos.go`,
	- aplicacion en rutas criticas de ventas, inventario y finanzas desde `backend/main.go`,
	- pruebas nuevas en `backend/handlers/empresa_permisos_test.go` para denegacion/aprobacion por rol y empresa.
- Validacion tecnica de la fase:
	- `go test ./handlers -run WithEmpresa -count=1` (ok).
	- `go test ./...` en `backend` (ok).
- Se actualiza la documentacion del proyecto para continuar el plan maestro de 14 puntos:
	- nuevo `documentos/plan_maestro_pos_multiempresa_14_puntos.md` con estado, entregables y backlog de ejecucion,
	- nueva `documentos/matriz_kpi_pos_multiempresa.md` con formulas/frecuencia/fuentes de KPI,
	- nueva `documentos/matriz_roles_permisos_pos_multiempresa.md` para iniciar el punto 3 de permisos y seguridad,
	- actualizacion de `documentos/descripcion_del_proyecto` para referenciar estos documentos como base de seguimiento.
- Continuación de implementación en `chat_con_inteligencia_artificial`:
	- Se corrige el orden de validación de autenticación para cuenta Google en `backend/handlers/chat_con_inteligencia_artificial_controller.go`.
	- Cuando no hay cuenta Google autenticada, los endpoints del módulo IA ahora responden `401` de forma consistente (en lugar de caer en validación de alcance con `403`).
	- Se centraliza validación de alcance con `ensureEmpresaAccessByAccount` para reutilizar la cuenta ya validada.
- Se agregan pruebas automáticas del módulo IA:
	- `backend/db/chat_inteligencia_artificial_test.go` (upsert/get de modelo preferido y acumulación de uso diario).
	- `backend/handlers/chat_con_inteligencia_artificial_controller_test.go` (autorización por cuenta Google y respuesta con modelo preferido).
- Validación técnica ejecutada en esta continuación:
	- `go test ./db -run EmpresaAI -count=1` (ok).
	- `go test ./handlers -run ModelosHandler -count=1` (ok).
	- `go test ./...` en `backend` (ok).
- Se amplía el módulo `chat_con_inteligencia_artificial` para registrar el modelo preferido por cuenta Google autenticada (por empresa):
	- Nueva tabla `empresa_ai_modelo_preferido` en `empresas.db` (UNIQUE por `empresa_id + admin_email`).
	- Nuevas funciones en `backend/db/chat_inteligencia_artificial.go`: `GetEmpresaAIModeloPreferido` y `UpsertEmpresaAIModeloPreferido`.
	- Nuevo endpoint `GET/PUT /api/empresa/chat_con_inteligencia_artificial/modelo_preferido`.
	- `GET /modelos` ahora devuelve `google_account` y `modelo_preferido`.
	- `POST /consultar` ahora persiste el `model_id` usado como preferencia de la cuenta Google y devuelve confirmación en respuesta.
- Se actualiza `web/administrar_empresa/chat_con_inteligencia_artificial.html` para:
	- cargar automáticamente el modelo preferido de la cuenta Google,
	- guardar el modelo preferido al cambiar selección,
	- mostrar la cuenta Google vinculada en el bloque de uso diario.
- Validación técnica ejecutada para esta ampliación:
	- `gofmt -w backend/db/chat_inteligencia_artificial.go backend/handlers/chat_con_inteligencia_artificial_controller.go backend/handlers/chat_con_inteligencia_artificial_router.go` (ok).
	- `go test ./...` en `backend` (ok).
- Se fortalece `backend/utils/utils.go` para observabilidad profesional:
	- `LoggingMiddleware` ahora genera `request_id` por solicitud, calcula `empresa_id` (query/header/JSON body) y registra inicio/fin con latencia.
	- Se agregan logs separados por empresa en `backend/logs/empresa_<id>.log` y un fallback global en `backend/logs/empresa_global.log`.
	- `JSONErrorMiddleware` ahora normaliza errores no-JSON incluyendo `request_id` y `empresa_id` cuando aplica, y registra errores API por empresa.
- Se ajustan endpoints multipart para reforzar separación de logs por empresa:
	- `backend/handlers/chat_tareas.go` y `backend/handlers/productos.go` ahora establecen `X-Empresa-ID` tras parsear `empresa_id` del formulario.
- Se endurece `backend/handlers/usuarios_empresa.go` en autenticación/primer ingreso:
	- se reemplazan respuestas `500` que exponían detalles internos por mensajes profesionales y seguros,
	- se agrega logging servidor con contexto (`empresa_id`, `email`, `id`) para trazabilidad sin filtrar errores sensibles al cliente.
- Se endurece `scripts/iniciar_servidor.ps1` para detectar caída temprana de `server.exe`: ahora conserva el `PID`, valida salida prematura y muestra las últimas líneas de `backend/server.err` para diagnóstico inmediato.
- Validación de corrección ejecutada:
	- `gofmt -w backend/utils/utils.go` (ok).
	- `go test ./...` en `backend` (ok).
- Se corrige `scripts/iniciar_servidor.ps1` en `Resolve-GoogleOAuthCredentials`: la construccion de `envCandidates` ahora usa `Join-Path -Path/-ChildPath` por elemento, evitando el error `CannotConvertArgument` de `Join-Path`.
- Se corrige `backend/db/finanzas.go` en `EnsureEmpresaFinanzasSchema`: los indices que dependen de columnas migradas (`periodo_contable` y `estado` de periodos) se crean al final de la migracion para compatibilidad con bases antiguas.
- Validacion de correccion ejecutada:
	- `go test ./...` en `backend` (ok).
	- `go run .` en `backend` (arranque correcto en `:8080`).
- Se incorpora el modulo `chat_con_inteligencia_artificial` en el panel empresarial con interfaz tipo chat en `web/administrar_empresa/chat_con_inteligencia_artificial.html`.
- Se crean `backend/db/chat_inteligencia_artificial.go`, `backend/handlers/chat_con_inteligencia_artificial_controller.go` y `backend/handlers/chat_con_inteligencia_artificial_router.go` para arquitectura modular (DB + controller + router).
- Se publican rutas del modulo IA:
	- `GET /api/empresa/chat_con_inteligencia_artificial/modelos`
	- `POST /api/empresa/chat_con_inteligencia_artificial/consultar`
	- `GET /api/empresa/chat_con_inteligencia_artificial/historial`
- Se agregan tablas en `empresas.db` para auditoria y limites diarios:
	- `empresa_ai_consultas`
	- `empresa_ai_uso_diario`
- Se integra `EnsureEmpresaAIChatSchema` y la migracion `2026-04-03-005-chat-ia-empresa` en `backend/main.go`.
- Se implementa aislamiento estricto por `empresa_id`, validacion de alcance de usuario y control de limite free-tier por empresa/proveedor/modelo/dia con opcion de upgrade.
- Se habilitan modelos famosos de OpenAI, DeepSeek y Hugging Face usando credenciales solo en backend mediante variables de entorno (`OPENAI_API_KEY`, `DEEPSEEK_API_KEY`, `HUGGINGFACE_API_KEY`).
- Se amplía el módulo financiero con control de periodos contables por empresa:
	- tabla `empresa_finanzas_periodos`.
	- endpoint `GET/POST/PUT /api/empresa/finanzas/periodos`.
	- acciones de cierre y reapertura de periodo.
- Se aplican bloqueos de integridad contable: no se permite crear/editar/eliminar/activar/desactivar movimientos cuando su periodo está cerrado.
- Se amplía `empresa_finanzas_movimientos` con:
	- `periodo_contable`,
	- retenciones (`retencion_fuente`, `retencion_ica`, `retencion_iva`, `total_retenciones`),
	- `total_neto`.
- Se amplía `empresa_finanzas_configuracion` con `cuenta_retenciones_cobrar` y `cuenta_retenciones_pagar`.
- Se completa la UI de finanzas para:
	- gestionar periodos (cerrar/reabrir/actualizar),
	- calcular total bruto, retenciones y neto,
	- filtrar por periodo,
	- exportar `balance general`, `libro diario` y `libro mayor` en CSV.
- Se corrige el escaneo de puertos de seguridad para compatibilidad IPv6 usando `net.JoinHostPort` en `backend/handlers/system_empresas_handlers.go`.
- Se ajusta `scripts/iniciar_servidor.ps1` para usar nombre de función con verbo aprobado de PowerShell en la carga de `.env`.
- Validación técnica ejecutada: `go test ./...` en `backend` (ok).
- Se implementa el módulo financiero multiempresa con enfoque unificado de ingresos y egresos en `web/administrar_empresa/finanzas.html`.
- Se crea `backend/db/finanzas.go` con esquema, validaciones y CRUD de:
	- `empresa_finanzas_movimientos`
	- `empresa_finanzas_configuracion`
- Se crea `backend/handlers/finanzas.go` y se publican rutas:
	- `GET/POST/PUT/DELETE /api/empresa/finanzas/movimientos`
	- `GET/POST/PUT /api/empresa/finanzas/configuracion`
- Se actualiza `backend/main.go` para asegurar el esquema financiero y registrar la migración `2026-04-03-003-finanzas`.
- Se integra el acceso al módulo en `web/administrar_empresa.html` y `web/js/administrar_empresa.js`.
- Se agrega `backend/db/finanzas_test.go` con pruebas de configuración y flujo CRUD de movimientos financieros.
- Se amplía `backend/tools/seed_motel_malibu/main.go` para sembrar configuración financiera y movimientos demo de ingreso/egreso.
- Se separa visualmente el libro financiero en dos pestañas operativas dentro del módulo: `Ingresos` y `Egresos`.
- Se agrega la pestaña `Todos` para consolidar ingresos y egresos en una sola vista del libro financiero.
- Se agrega exportación del libro financiero filtrado por fechas a:
	- Excel (CSV compatible con Excel).
	- PDF (vista de impresión).
	- JSON contable para integración externa (incluye resumen, detalle y asientos recomendados).
- Se amplía la configuración financiera por empresa para contabilidad externa con parametrización de:
	- destino de integración (`generico`, `siigo`, `world_office`, `alegra`),
	- cuentas base (caja/bancos, ingresos, IVA generado, gastos, IVA descontable),
	- cuentas por categoría para ingresos y egresos.
- La exportación `JSON contable` deja de usar cuentas fijas y ahora construye asientos con la parametrización real guardada por empresa.
- El JSON exportado incorpora `accounting_profile` y `erp_projection` por movimiento para facilitar mapeo hacia software contable externo.
- Se actualiza `backend/db/finanzas_test.go` para validar persistencia de la nueva parametrización contable.
- Se amplía `web/administrar_empresa/finanzas.html` con salidas contables adicionales:
	- Plantilla dedicada SIIGO (CSV) para importación de asientos.
	- Balance de prueba (CSV).
	- Estado de resultados (CSV).
- Se crea `documentos/plantillas/siigo_plantilla_importacion_asientos.csv` como plantilla de referencia ERP.
- Se crea `documentos/informe_contable_directivo_2026-04-03.md` con revisión de cumplimiento contable/directivo, brechas y plan recomendado.
- Validación técnica ejecutada:
	- `go test ./... -count=1` (ok).
	- `go run ./tools/seed_motel_malibu` (ok, incluye creación de 4 movimientos financieros demo).
	- `runTests` global (ok: 3/3).

## 2026-04-03
- Se implementa control de inventario en carrito: al agregar items de producto se descuenta stock y al desactivar/eliminar items abiertos se revierte automáticamente.
- Se asegura que, al cerrar una venta, el descuento de inventario permanezca aplicado y no se revierta en el pago.
- Se mejoran respuestas de API para stock insuficiente en operaciones de items de carrito.
- Se agrega `backend/db/carritos_inventario_test.go` con pruebas de descuento de inventario y caso de stock insuficiente.
- Se amplía `backend/tools/seed_motel_malibu/main.go` para registrar 10 clientes y 10 usuarios de empresa.
- La semilla valida automáticamente el flujo comercial completo: venta cerrada, descuento de inventario al agregar y persistencia tras pagar.
- Se confirma en seed la validación de impresión con vista previa POS y Carta.
- Se amplía `web/administrar_empresa/reportes.html` con reporte de ventas, reporte de productos y reporte de compra de productos, todos con búsqueda por rango de fechas.
- Validación técnica ejecutada: `go test ./auth ./db ./handlers ./metrics ./utils` (ok) y `go run ./tools/seed_motel_malibu` (ok).
- Se agrega el vínculo `Ayuda` en el menú flotante global (`web/menu.js`) y se reestructura `web/ayuda/ayuda.html` como centro de ayuda con menú interno y sección de APIs.
- Se adapta `web/administrar_empresa/carrito_de_compras.html` para operación con lector de código de barras (escaneo por código/SKU, Enter para agregar y acumulación opcional de cantidad).
- Se extiende `web/administrar_empresa/configuracion.html` con configuración por empresa para el lector: habilitar, autofoco y acumulación.
- Se amplía `web/administrar_empresa/reportes.html` con KPI de productos bajo mínimo y reporte de inventario actual por bodega.
- Validación técnica ejecutada para flujo carrito/inventario multiempresa: `go test ./db -run Carrito -count=1` (ok) y `go test ./handlers -run Carritos -count=1` (ok).

## 2026-04-02
- Se crea la herramienta `backend/tools/seed_motel_malibu/main.go` para cargar datos demo comerciales en la empresa Motel Malibu.
- La semilla inserta 10 productos con precios COP, 5 clientes y crea una venta de prueba cerrada para validar el flujo comercial.
- Se valida la configuracion de impresion con vista previa de formatos POS y Carta desde la herramienta de seed.
- Se implementa la seccion `web/administrar_empresa/reportes.html` con KPIs, ventas cerradas, top productos, top clientes y resumen de impresion.
- Se reestructura `backend/tools` en subcarpetas por herramienta para eliminar conflictos de compilación por múltiples `main`.
- Se valida backend completo con `go test ./...` (ok).
- Se valida el módulo GPS con pruebas específicas:
	- `go test ./db -run TestEmpresaGPSDispositivosYRecorridosCRUD -count=1` (ok).
	- `go test ./handlers -run TestEmpresaUbicacionGPSHandlersCRUDFlow -count=1` (ok).
- Se implementa el modulo de ubicacion GPS por empresa con soporte de multiples dispositivos.
- Se agregan tablas `empresa_gps_dispositivos` y `empresa_gps_recorridos` en `empresas.db`.
- Se crean endpoints CRUD para dispositivos y recorridos GPS en `/api/empresa/ubicacion_gps/*`.
- Se agrega la pagina `web/administrar_empresa/ubicacion_gps.html` con mapa OpenStreetMap (Leaflet).
- Se habilita tracking automatico de recorridos cada 10 segundos por dispositivo.
- Se agregan pruebas en `backend/db/ubicacion_gps_test.go` y `backend/handlers/ubicacion_gps_test.go`.
