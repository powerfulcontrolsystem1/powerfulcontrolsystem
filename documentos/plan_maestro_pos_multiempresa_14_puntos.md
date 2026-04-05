# Plan maestro POS multiempresa (15 puntos)

Fecha de actualizacion: 2026-04-05
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
| 5 | Control de inventarios | en curso | stock, alertas y movimientos de bodega |
| 6 | Gestion de clientes | en curso | perfil, historial y segmentacion |
| 7 | Gestion de proveedores | en curso | catalogo, precios y condiciones |
| 8 | Modulo de facturacion electronica | en curso | emision legal y cumplimiento normativo |
| 9 | Modulo de compras | en curso | orden, recepcion y contabilizacion |
| 10 | Modulo contable integrado | en curso | asientos automaticos por evento |
| 11 | Reportes financieros | en curso | balance, estado de resultados, flujo de caja |
| 12 | Cierres de caja | en curso | arqueo y cierre por sucursal/empresa |
| 13 | Calidad, UAT y despliegue | en curso | validacion integral y salida controlada |
| 14 | Operacion continua | en curso | mejora continua con KPI y roadmap trimestral |
| 15 | Modulo de auditoria por empresa | en curso | trazabilidad por usuario/accion/recurso con consulta por empresa |

### Punto 14. Operacion continua (avance 2026-04-04)

Implementacion tecnica inicial completada:
- Se crea guia de operacion continua: `documentos/punto_14_operacion_continua.md`.
	- Define cadencia diaria/semanal/mensual/trimestral.
	- Define KPI de gobierno operativo y flujo minimo obligatorio.
- Se crea roadmap trimestral: `documentos/roadmap_trimestral_pos_multiempresa.md`.
	- Establece focos Q2, Q3 y Q4 de 2026 con indicadores de salida.
- Se crea script de reporte operativo: `scripts/generar_reporte_operacion_continua.ps1`.
	- Consolida estado del plan maestro.
	- Reutiliza evidencia tecnica del punto 13.
	- Genera `documentos/punto_14_operacion_continua_reporte.md` y bitacora en `scripts/logs/`.

Validacion tecnica ejecutada en esta iteracion:
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\generar_reporte_operacion_continua.ps1` (ok).

Estado actual del punto 14:
- Framework operativo de mejora continua: activo.
- Seguimiento trimestral KPI/roadmap: inicializado.

### Punto 13. Calidad, UAT y despliegue (avance 2026-04-04)

Implementacion tecnica inicial completada:
- Se crea guia operativa: `documentos/punto_13_calidad_uat_despliegue.md`.
	- Define flujo de validacion tecnica, matriz UAT minima, gates de salida y criterio de rollback.
- Se crea script de ejecucion repetible: `scripts/validar_punto_13.ps1`.
	- Ejecuta suite productiva y suite completa del backend.
	- Genera log tecnico en `scripts/logs/` y reporte consolidado en `documentos/punto_13_validacion_integral_resultado.md`.
- Se actualiza `documentos/release_checklist.md`.
	- Incorpora gate explicito del punto 13 con comando operativo y verificacion de evidencia.

Validacion tecnica ejecutada en esta iteracion:
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\validar_punto_13.ps1` (ok).
- Resultado reportado:
	- `go test ./auth ./db ./handlers ./metrics ./utils -count=1` (ok).
	- `go test ./... -count=1` (ok).

Estado actual del punto 13:
- Gate tecnico: aprobado.
- Gate UAT manual: pendiente de ejecucion operativa por modulo antes de salida productiva.

### Punto 6. Gestion de clientes (avance 2026-04-04)

Implementacion tecnica completada (fase funcional inicial):
- `backend/db/clientes.go`:
	- agrega contratos analiticos `ClientePerfilComercial`, `ClienteCompraHistorial` y `ClienteSegmentacionResumen`.
	- agrega funciones:
		- `GetClientePerfilComercialByEmpresa`,
		- `GetClienteHistorialComprasByEmpresa`,
		- `GetClientesSegmentacionByEmpresa`.
	- calcula metricas por cliente sobre `carritos_compras` (compras, monto, ticket, dias sin compra, segmento).
- `backend/handlers/clientes.go`:
	- amplia `GET /api/empresa/clientes` con acciones:
		- `action=perfil`,
		- `action=historial`,
		- `action=segmentacion|segmentos`.
	- agrega parseo robusto de `cliente_id`/`id`.
- `web/administrar_empresa/administrar_clientes.html`:
	- agrega bloque `Segmentacion de clientes (punto 6)`.
	- agrega bloque `Perfil e historial del cliente`.
	- agrega accion `Perfil` por fila y refresco integral de datos analiticos tras cambios CRUD.

Cobertura de pruebas agregada/extendida:
- `backend/db/clientes_test.go`:
	- `TestGetClientePerfilComercialByEmpresaAndHistorial`.
	- `TestGetClientePerfilComercialByEmpresaSinComprasSegmentoNuevo`.
- `backend/handlers/clientes_test.go`:
	- `TestEmpresaClientesHandlerPerfilHistorialSegmentacion`.

Validacion ejecutada:
- `gofmt -w db/clientes.go db/clientes_test.go handlers/clientes.go handlers/clientes_test.go`.
- `go test ./db -run "TestGetClientePerfilComercialByEmpresaAndHistorial|TestGetClientePerfilComercialByEmpresaSinComprasSegmentoNuevo" -count=1` (ok).
- `go test ./handlers -run "TestEmpresaClientesHandlerPerfilHistorialSegmentacion" -count=1` (ok).
- `get_errors` en backend/frontend modificado (ok).

### Punto 7. Gestion de proveedores (avance 2026-04-04)

Implementacion tecnica completada (fase funcional inicial):
- `backend/db/productos.go`:
	- amplia el contrato `Proveedor` con datos comerciales:
		- `catalogo_referencia`,
		- `precio_base_referencial`,
		- `descuento_porcentaje`,
		- `plazo_pago_dias`,
		- `condicion_entrega`.
	- actualiza `EnsureEmpresaProductosSchema` para crear/migrar columnas nuevas en `proveedores` sin romper instalaciones existentes.
	- agrega validaciones de negocio para proveedores:
		- precio base >= 0,
		- descuento entre 0 y 100,
		- plazo de pago >= 0.
	- actualiza `CreateProveedor`, `GetProveedoresByEmpresa` y `UpdateProveedor` para persistir/consultar la nueva informacion comercial.
- `backend/handlers/productos.go`:
	- valida en API los nuevos campos comerciales en `POST/PUT /api/empresa/proveedores`.
	- amplia el `payload` de eventos contables de compras para registrar condiciones y precios de referencia del proveedor.
- `web/administrar_empresa/administrar_productos.html`:
	- amplia formulario de proveedores con catalogo referencia, condicion de entrega, precio base, descuento y plazo de pago.
	- amplia la grilla de proveedores para visualizar precio base y condiciones comerciales.
	- agrega validaciones frontend para rangos de precio/descuento/plazo antes de enviar al backend.

Cobertura de pruebas agregada/extendida:
- `backend/db/productos_categorias_test.go`:
	- `TestProveedorCRUDIncluyeCatalogoPreciosYCondiciones`.
- `backend/handlers/eventos_contables_modulos_test.go`:
	- `TestEmpresaProveedoresEmiteEventoContableCompras` (extendido con nuevos campos),
	- `TestEmpresaProveedoresRechazaCamposComercialesInvalidos`.

Validacion ejecutada:
- `gofmt -w db/productos.go db/productos_categorias_test.go handlers/productos.go handlers/eventos_contables_modulos_test.go`.
- `go test ./db -run "TestProveedorCRUDIncluyeCatalogoPreciosYCondiciones" -count=1` (ok).
- `go test ./handlers -run "TestEmpresaProveedoresEmiteEventoContableCompras|TestEmpresaProveedoresRechazaCamposComercialesInvalidos" -count=1` (ok).
- `get_errors` en backend/frontend modificado (ok).

### Punto 8. Modulo de facturacion electronica (avance 2026-04-04 - cumplimiento normativo en emision)

Implementacion tecnica completada (fase de cumplimiento legal inicial):
- `backend/db/facturacion_electronica.go`:
	- agrega `PrepareFacturacionDocumentoLegal` para validar configuracion legal por empresa/pais y reservar consecutivo.
	- valida vigencia de resolucion, rango de consecutivos y datos fiscales minimos antes de emitir.
	- genera `numero_legal` y `codigo_validacion` (hash de trazabilidad) para la factura emitida.
- `backend/db/documentos_transaccionales.go`:
	- amplia `empresa_facturacion_documentos` y su contrato con:
		- `numero_legal`,
		- `codigo_validacion`,
		- `pais_codigo`,
		- `ambiente_fe`.
	- actualiza lectura/upsert para persistir metadata legal de emision y mantener compatibilidad con documentos existentes.
- `backend/handlers/facturacion_electronica.go`:
	- `action=emitir` exige cumplimiento normativo previo; retorna `422` cuando falta configuracion legal.
	- persiste metadata legal en documento transaccional y la propaga a payload/evento de facturacion.
	- responde `cumplimiento_normativo` en emisiones exitosas para auditoria operativa.
- `web/administrar_empresa/facturacion_electronica.html`:
	- agrega bloque `Emision documental (punto 8)` con acciones `Emitir factura`, `Anular factura` y `Emitir nota credito`.
	- muestra resultado estructurado incluyendo datos legales cuando aplica.

Cobertura de pruebas agregada/extendida:
- `backend/db/documentos_transaccionales_test.go`:
	- `TestEmpresaDocumentoFacturacionUpsertAndGet` valida persistencia de `numero_legal`, `codigo_validacion`, `pais_codigo` y `ambiente_fe`.
- `backend/db/facturacion_electronica_test.go` (nuevo):
	- `TestPrepareFacturacionDocumentoLegalSuccessAndConsecutivo` valida emision legal e incremento de consecutivo.
	- `TestPrepareFacturacionDocumentoLegalRejectsExpiredResolution` valida rechazo por resolucion vencida.
	- `TestPrepareFacturacionDocumentoLegalRejectsConfigInactivaAndRangoAgotado` valida rechazo por configuracion FE inactiva y por rango agotado.
- `backend/handlers/eventos_contables_modulos_test.go`:
	- `TestEmpresaFacturacionTransaccionalEmiteEventosContables` valida emision legal y persistencia de metadata.
	- `TestEmpresaFacturacionTransaccionalEmitirRechazaSinCumplimientoLegal` valida rechazo `422` sin configuracion legal minima.

Validacion ejecutada:
- `gofmt -w db/facturacion_electronica.go db/documentos_transaccionales.go db/documentos_transaccionales_test.go handlers/facturacion_electronica.go handlers/eventos_contables_modulos_test.go`.
- `gofmt -w db/facturacion_electronica_test.go`.
- `go test ./db -run "TestPrepareFacturacionDocumentoLegal" -count=1` (ok).
- `go test ./db -run "TestEmpresaDocumentoFacturacionUpsertAndGet" -count=1` (ok).
- `go test ./handlers -run "TestEmpresaFacturacionTransaccionalEmiteEventosContables|TestEmpresaFacturacionTransaccionalEmitirRechazaSinCumplimientoLegal|TestEmpresaFacturacionTransaccionalRechazaTransicionInvalida" -count=1` (ok).
- `go test ./db ./handlers -count=1` (ok).

### Punto 5. Control de inventarios (inicio ejecutado 2026-04-04)

Implementacion tecnica completada (fase inicial):
- `backend/db/productos.go`:
	- valida reglas de stock en alta/edicion de productos (`stock_minimo <= stock_maximo`, sin negativos),
	- agrega `GetAlertasQuiebreByEmpresa` para alertas por empresa/producto/bodega,
	- formaliza kardex operativo en `GetMovimientosByEmpresa` con filtros `bodega_id`, `tipo`, `desde`, `hasta`.
- `backend/handlers/productos.go`:
	- agrega endpoint `GET /api/empresa/inventario/alertas`,
	- agrega modo de compatibilidad `GET /api/empresa/inventario/existencias?action=alertas|alertas_quiebre|quiebre`,
	- amplía `GET /api/empresa/inventario/movimientos` con filtros de kardex por bodega/tipo/rango y validacion de fechas (`YYYY-MM-DD`).
- `backend/main.go`:
	- registra ruta protegida `/api/empresa/inventario/alertas` bajo `WithEmpresaInventarioPermissions`.
- `web/administrar_empresa/administrar_productos.html`:
	- incorpora tabla `Alertas de quiebre por bodega` en el modulo de inventario.

Cobertura de pruebas agregada/extendida:
- `backend/handlers/productos_categorias_test.go`:
	- `TestEmpresaInventarioAlertasHandlerDevuelveQuiebrePorBodega`.
	- `TestEmpresaInventarioMovimientosHandlerFiltraPorBodegaTipoYRango`.
- `backend/db/productos_categorias_test.go`:
	- `TestCreateAndUpdateProductoValidanStockMinMax`.

Validacion ejecutada:
- `go test ./handlers ./db -count=1` (ok).

### Punto 5. Control de inventarios (continuacion UI operativa 2026-04-04)

Implementacion tecnica completada (fase 2):
- `web/administrar_empresa/administrar_productos.html`:
	- agrega filtros operativos para `Alertas de quiebre` por bodega,
	- agrega filtros de kardex en `Movimientos recientes` por:
		- `bodega_id`,
		- `tipo`,
		- `desde` y `hasta`.
	- integra acciones de `Filtrar` y `Limpiar` para ambos bloques, reutilizando endpoints existentes del backend.

Validacion ejecutada:
- diagnostico de archivo (`get_errors`) en `web/administrar_empresa/administrar_productos.html` (ok).

### Punto 5. Control de inventarios (continuacion KPI operativo 2026-04-04)

Implementacion tecnica completada (fase 3):
- `backend/db/productos.go`:
	- agrega `InventarioResumen` con metricas operativas del modulo,
	- agrega `GetInventarioResumenByEmpresa` para consolidar:
		- existencias totales,
		- alertas (`sin_stock`, `bajo_minimo`, `deficit_total`),
		- movimientos por rango (`entrada`, `salida`, `traslado`, `ajuste`, `total`) y ultimo movimiento.
- `backend/handlers/productos.go`:
	- agrega endpoint `GET /api/empresa/inventario/resumen` con validacion de fechas `YYYY-MM-DD`.
- `backend/main.go`:
	- registra ruta protegida `/api/empresa/inventario/resumen` bajo `WithEmpresaInventarioPermissions`.
- `web/administrar_empresa/administrar_productos.html`:
	- agrega KPI de inventario visibles en cabecera (`alertas`, `sin_stock`, `movimientos del periodo`, `deficit total`),
	- integra consumo de `GET /api/empresa/inventario/resumen`,
	- sincroniza resumen con filtros de rango del kardex y con refrescos de existencias.

Cobertura de pruebas agregada/extendida:
- `backend/handlers/productos_categorias_test.go`:
	- `TestEmpresaInventarioResumenHandlerDevuelveKPIsPorRango`.
- `backend/db/productos_categorias_test.go`:
	- `TestGetInventarioResumenByEmpresaCalculaIndicadores`.

Validacion ejecutada:
- `gofmt -w db/productos.go handlers/productos.go handlers/productos_categorias_test.go db/productos_categorias_test.go main.go`.
- `go test ./handlers ./db -count=1` (ok).
- `get_errors` en `web/administrar_empresa/administrar_productos.html` (ok).

### Punto 5. Control de inventarios (continuacion operacional 2026-04-04)

Implementacion tecnica completada (fase 4):
- `web/administrar_empresa/administrar_productos.html`:
	- agrega bloque `Top productos críticos (déficit)` basado en alertas de quiebre,
	- prioriza visualmente productos `sin_stock` y mayor déficit,
	- agrega accion `Preparar reposición` que preconfigura el formulario de ajuste (`tipo=entrada`, producto, bodega, cantidad sugerida y referencia).

Validacion ejecutada:
- `get_errors` en `web/administrar_empresa/administrar_productos.html` (ok).

### Punto 5. Control de inventarios (continuacion analitica 2026-04-04)

Implementacion tecnica completada (fase 5):
- `backend/db/productos.go`:
	- agrega `InventarioTendenciaDia`,
	- agrega `GetInventarioTendenciaByEmpresa` con serie diaria de `entradas`, `salidas`, `traslados`, `neto` y `eventos`,
	- soporta filtros por `bodega_id`, `desde`, `hasta` y ventana por `dias`.
- `backend/handlers/productos.go`:
	- agrega endpoint `GET /api/empresa/inventario/tendencia` con validacion de fechas (`YYYY-MM-DD`) y parametros operativos.
- `backend/main.go`:
	- registra ruta protegida `/api/empresa/inventario/tendencia` bajo `WithEmpresaInventarioPermissions`.
- `web/administrar_empresa/administrar_productos.html`:
	- agrega tabla `Tendencia diaria inventario` sincronizada con filtros del kardex,
	- muestra neto acumulado y eventos del rango para seguimiento operativo.

Cobertura de pruebas agregada/extendida:
- `backend/handlers/productos_categorias_test.go`:
	- `TestEmpresaInventarioTendenciaHandlerDevuelveSeriePorRango`.
- `backend/db/productos_categorias_test.go`:
	- `TestGetInventarioTendenciaByEmpresaDevuelveSerieDiaria`.

Validacion ejecutada:
- `gofmt -w db/productos.go handlers/productos.go handlers/productos_categorias_test.go db/productos_categorias_test.go main.go`.
- `go test ./handlers ./db -count=1` (ok).
- `get_errors` en `web/administrar_empresa/administrar_productos.html` (ok).

### Punto 5. Control de inventarios (continuacion operativa-analitica 2026-04-04)

Implementacion tecnica completada (fase 6):
- `backend/db/productos.go`:
	- agrega `InventarioBalanceBodega`,
	- agrega `GetInventarioBalanceBodegasByEmpresa` para consolidar por bodega:
		- `entradas`,
		- `salidas`,
		- `traslados_entrada`, `traslados_salida`, `traslado_neto`,
		- `neto` y `eventos` en rango.
- `backend/handlers/productos.go`:
	- agrega endpoint `GET /api/empresa/inventario/balance_bodegas` con validacion de fechas (`YYYY-MM-DD`) y soporte por `bodega_id`.
- `backend/main.go`:
	- registra ruta protegida `/api/empresa/inventario/balance_bodegas` bajo `WithEmpresaInventarioPermissions`.
- `web/administrar_empresa/administrar_productos.html`:
	- agrega bloque `Balance por bodega` sincronizado con filtros de kardex,
	- muestra neto por bodega y neto acumulado del rango consultado.

Cobertura de pruebas agregada/extendida:
- `backend/handlers/productos_categorias_test.go`:
	- `TestEmpresaInventarioBalanceBodegasHandlerDevuelveResumenPorBodega`.
- `backend/db/productos_categorias_test.go`:
	- `TestGetInventarioBalanceBodegasByEmpresaConsolidaMovimientos`.

Validacion ejecutada:
- `gofmt -w db/productos.go handlers/productos.go handlers/productos_categorias_test.go db/productos_categorias_test.go main.go`.
- `go test ./handlers ./db -count=1` (ok).
- `get_errors` en `web/administrar_empresa/administrar_productos.html` (ok).

### Punto 5. Control de inventarios (continuacion preventiva 2026-04-04)

Implementacion tecnica completada (fase 7):
- `backend/db/productos.go`:
	- agrega `InventarioProyeccionQuiebre`,
	- agrega `GetInventarioProyeccionQuiebreByEmpresa` para estimar por producto/bodega:
		- `salida_promedio_diaria`,
		- `dias_cobertura`,
		- `estado_proyeccion`,
		- `sugerido_reposicion`,
		- priorizacion por severidad de riesgo.
- `backend/handlers/productos.go`:
	- agrega endpoint `GET /api/empresa/inventario/proyeccion_quiebre` con validacion de `dias_ventana`, `bodega_id`, `limit` y `offset`.
- `backend/main.go`:
	- registra ruta protegida `/api/empresa/inventario/proyeccion_quiebre` bajo `WithEmpresaInventarioPermissions`.
- `web/administrar_empresa/administrar_productos.html`:
	- agrega bloque `Proyeccion de quiebre (preventiva)` sincronizado con filtros del kardex,
	- agrega accion `Preparar` para preconfigurar el ajuste de inventario con reposicion preventiva sugerida.

Cobertura de pruebas agregada/extendida:
- `backend/handlers/productos_categorias_test.go`:
	- `TestEmpresaInventarioProyeccionQuiebreHandlerDevuelveRiesgo`.
- `backend/db/productos_categorias_test.go`:
	- `TestGetInventarioProyeccionQuiebreByEmpresaPriorizaRiesgo`.

Validacion ejecutada:
- `gofmt -w db/productos.go handlers/productos.go handlers/productos_categorias_test.go db/productos_categorias_test.go main.go`.
- `go test ./handlers ./db -count=1` (ok).
- `get_errors` en `web/administrar_empresa/administrar_productos.html` (ok).

### Punto 5. Control de inventarios (continuacion preventiva-compras 2026-04-04)

Implementacion tecnica completada (fase 8):
- `backend/db/productos.go`:
	- agrega `InventarioPlanReposicionItem`,
	- agrega `GetInventarioPlanReposicionByEmpresa` para consolidar por proveedor:
		- items sugeridos por riesgo,
		- cantidad recomendada,
		- costo unitario de referencia,
		- costo estimado por item.
- `backend/handlers/productos.go`:
	- agrega endpoint `GET /api/empresa/inventario/plan_reposicion` con validacion de `dias_ventana`, `solo_riesgo`, `bodega_id`, `limit` y `offset`.
- `backend/main.go`:
	- registra ruta protegida `/api/empresa/inventario/plan_reposicion` bajo `WithEmpresaInventarioPermissions`.
- `web/administrar_empresa/administrar_productos.html`:
	- agrega bloque `Plan de reposicion por proveedor (fase 8)`,
	- muestra proveedor, producto, estado, cantidad sugerida y costo estimado,
	- agrega accion `Preparar` para preconfigurar ajuste de inventario desde el plan.

Cobertura de pruebas agregada/extendida:
- `backend/handlers/productos_categorias_test.go`:
	- `TestEmpresaInventarioPlanReposicionHandlerDevuelveCostoEstimado`.
- `backend/db/productos_categorias_test.go`:
	- `TestGetInventarioPlanReposicionByEmpresaConsolidaProveedorYCosto`.

Validacion ejecutada:
- `gofmt -w db/productos.go handlers/productos.go handlers/productos_categorias_test.go db/productos_categorias_test.go main.go`.
- `go test ./handlers ./db -count=1` (ok).
- `get_errors` en `web/administrar_empresa/administrar_productos.html` (ok).

### Punto 5. Control de inventarios (continuacion preventiva-compras consolidada 2026-04-04)

Implementacion tecnica completada (fase 9):
- `backend/db/productos.go`:
	- agrega `InventarioPlanReposicionProveedorResumen`,
	- agrega `GetInventarioPlanReposicionResumenByEmpresa` para consolidar por proveedor:
		- items y productos unicos,
		- cantidad total sugerida,
		- costo total estimado,
		- conteo de urgencia (`quiebre_inminente`, `riesgo_alto`).
- `backend/handlers/productos.go`:
	- agrega endpoint `GET /api/empresa/inventario/plan_reposicion_resumen` con validacion de `dias_ventana`, `solo_riesgo`, `bodega_id`, `limit` y `offset`.
- `backend/main.go`:
	- registra ruta protegida `/api/empresa/inventario/plan_reposicion_resumen` bajo `WithEmpresaInventarioPermissions`.
- `web/administrar_empresa/administrar_productos.html`:
	- agrega bloque `Consolidado de compra por proveedor (fase 9)`,
	- permite filtrar la tabla del plan (fase 8) por proveedor desde el consolidado,
	- agrega control `Ver todos` para limpiar filtro y volver a la vista global.

Cobertura de pruebas agregada/extendida:
- `backend/handlers/productos_categorias_test.go`:
	- `TestEmpresaInventarioPlanReposicionResumenHandlerAgrupaProveedor`.
- `backend/db/productos_categorias_test.go`:
	- `TestGetInventarioPlanReposicionResumenByEmpresaAgrupaProveedor`.

Validacion ejecutada:
- `gofmt -w db/productos.go handlers/productos.go handlers/productos_categorias_test.go db/productos_categorias_test.go main.go`.
- `go test ./handlers ./db -count=1` (ok).
- `get_errors` en `web/administrar_empresa/administrar_productos.html` (ok).

### Punto 5. Control de inventarios (continuacion preventiva-compras ordenable 2026-04-04)

Implementacion tecnica completada (fase 10):
- `backend/db/productos.go`:
	- agrega `InventarioPlanReposicionBorradorItem` y `InventarioPlanReposicionBorradorCompra`,
	- agrega `GetInventarioPlanReposicionBorradorByEmpresa` para construir borrador de orden por proveedor con:
		- codigo sugerido de borrador,
		- lineas por producto/bodega,
		- agregados de cantidad/costo,
		- conteo de severidad (`quiebre_inminente`, `bajo_minimo`, `riesgo_alto`, `riesgo_medio`).
- `backend/handlers/productos.go`:
	- agrega endpoint `GET /api/empresa/inventario/plan_reposicion_borrador` con validacion de `proveedor_id`, `dias_ventana`, `solo_riesgo` y `bodega_id`.
- `backend/main.go`:
	- registra ruta protegida `/api/empresa/inventario/plan_reposicion_borrador` bajo `WithEmpresaInventarioPermissions`.
- `web/administrar_empresa/administrar_productos.html`:
	- agrega bloque `Borrador de orden de compra por proveedor (fase 10)`,
	- agrega accion `Borrador OC` desde el consolidado de proveedores (fase 9),
	- agrega control `Limpiar borrador` para reiniciar vista.

Cobertura de pruebas agregada/extendida:
- `backend/handlers/productos_categorias_test.go`:
	- `TestEmpresaInventarioPlanReposicionBorradorHandlerConstruyeDocumento`.
- `backend/db/productos_categorias_test.go`:
	- `TestGetInventarioPlanReposicionBorradorByEmpresaProveedor`.

Validacion ejecutada:
- `gofmt -w db/productos.go handlers/productos.go main.go db/productos_categorias_test.go handlers/productos_categorias_test.go`.
- `go test ./handlers ./db -count=1` (ok).
- `get_errors` en backend/frontend modificado (ok).

### Punto 5. Control de inventarios (continuacion preventiva-compras emitible 2026-04-04)

Implementacion tecnica completada (fase 11):
- `backend/db/productos.go`:
	- agrega `InventarioPlanReposicionOrdenEmitida`,
	- agrega `EmitirOrdenCompraDesdePlanReposicionBorrador` para emitir una OC desde el borrador preventivo y persistirla en `empresa_compras_documentos`.
- `backend/handlers/productos.go`:
	- agrega endpoint `POST /api/empresa/compras/plan_reposicion/emitir_orden` con validacion de `empresa_id`, `proveedor_id`, `bodega_id`, `dias_ventana`, `solo_riesgo` y datos documentales.
	- registra evento contable `orden_compra_emitida` al confirmar la emision del documento.
- `backend/main.go`:
	- registra ruta protegida `/api/empresa/compras/plan_reposicion/emitir_orden` bajo `WithEmpresaComprasPermissions`.
- `web/administrar_empresa/administrar_productos.html`:
	- agrega accion `Emitir orden` en bloque de borrador (fase 10),
	- envía el borrador al nuevo endpoint de compras,
	- refresca plan y consolidado tras la emision.

Cobertura de pruebas agregada/extendida:
- `backend/handlers/productos_categorias_test.go`:
	- `TestEmpresaComprasPlanReposicionEmitirOrdenHandlerEmiteDocumento`.
- `backend/db/productos_categorias_test.go`:
	- `TestEmitirOrdenCompraDesdePlanReposicionBorradorPersistDoc`.

Validacion ejecutada:
- `gofmt -w db/productos.go handlers/productos.go main.go db/productos_categorias_test.go handlers/productos_categorias_test.go`.
- `go test ./handlers ./db -count=1` (ok).
- `get_errors` en backend/frontend modificado (ok).

### Punto 5. Control de inventarios (continuacion preventiva-compras ciclo documental 2026-04-04)

Implementacion tecnica completada (fase 12):
- `backend/db/productos.go`:
	- agrega `InventarioPlanReposicionOrdenEstadoActualizado`,
	- agrega `ActualizarEstadoOrdenCompraDesdeReposicion` para transicionar la OC emitida por reposicion con acciones `recepcionar_compra` y `contabilizar_compra`.
- `backend/handlers/productos.go`:
	- agrega endpoint `POST /api/empresa/compras/plan_reposicion/actualizar_estado` con validacion de `empresa_id`, `proveedor_id`, `documento_codigo` y `accion`.
	- registra eventos contables `compra_recepcionada` y `compra_contabilizada` segun transicion.
- `backend/main.go`:
	- registra ruta protegida `/api/empresa/compras/plan_reposicion/actualizar_estado` bajo `WithEmpresaComprasPermissions`.
- `web/administrar_empresa/administrar_productos.html`:
	- amplía el bloque de borrador a `fases 10-12`,
	- agrega acciones `Recepcionar orden` y `Contabilizar orden`,
	- muestra contexto de estado de la OC emitida en el ciclo documental.

Cobertura de pruebas agregada/extendida:
- `backend/handlers/productos_categorias_test.go`:
	- `TestEmpresaComprasPlanReposicionActualizarEstadoHandlerGestionaCiclo`.
- `backend/db/productos_categorias_test.go`:
	- `TestActualizarEstadoOrdenCompraDesdeReposicionCiclo`.

Validacion ejecutada:
- `gofmt -w backend/db/productos.go backend/db/productos_categorias_test.go backend/handlers/productos.go backend/handlers/productos_categorias_test.go backend/main.go`.
- `go test ./db -run "TestEmitirOrdenCompraDesdePlanReposicionBorradorPersistDoc|TestActualizarEstadoOrdenCompraDesdeReposicionCiclo"` (ok).
- `go test ./handlers -run "TestEmpresaComprasPlanReposicionEmitirOrdenHandlerEmiteDocumento|TestEmpresaComprasPlanReposicionActualizarEstadoHandlerGestionaCiclo"` (ok).
- `get_errors` en backend/frontend modificado (ok).

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

### Punto 9. Modulo de compras (avance 2026-04-04 - modulo dedicado de documentos)

Implementacion tecnica adicional completada:
- `backend/db/documentos_transaccionales.go`:
	- agrega `ListEmpresaDocumentosCompraByEmpresa` para consulta operativa de documentos por empresa con filtros.
	- agrega `SetEmpresaDocumentoCompraEstadoByCodigo` para activacion/desactivacion logica por codigo documental.
- `backend/handlers/compras.go` (nuevo):
	- incorpora `EmpresaComprasDocumentosHandler` con operaciones `GET/POST/PUT/DELETE` en `/api/empresa/compras/documentos`.
	- habilita acciones de ciclo documental:
		- `crear`,
		- `emitir_orden`,
		- `recepcionar_compra`,
		- `contabilizar_compra`,
		- `activar/desactivar`.
	- registra eventos contables del modulo compras para trazabilidad de cada transicion.
- `backend/main.go`:
	- registra la nueva ruta protegida `/api/empresa/compras/documentos` con middleware de permisos de compras.
- `web/administrar_empresa/compras.html` (nuevo):
	- agrega modulo dedicado de compras en frontend con formulario, filtros y acciones por documento.
- `web/administrar_empresa.html` y `web/js/administrar_empresa.js`:
	- integran acceso de menu `Compras` con control de visibilidad por permisos de modulo.

Cobertura de pruebas agregada/extendida:
- `backend/db/documentos_transaccionales_test.go`:
	- `TestEmpresaDocumentoCompraListAndSetEstadoByCodigo`.
- `backend/handlers/compras_documentos_test.go` (nuevo):
	- `TestEmpresaComprasDocumentosCicloCompleto`.
	- `TestEmpresaComprasDocumentosActivarYFiltrarInactivos`.
	- `TestEmpresaComprasDocumentosTransicionInvalida`.

Validacion ejecutada:
- `gofmt -w handlers/compras.go handlers/compras_documentos_test.go main.go db/documentos_transaccionales.go db/documentos_transaccionales_test.go`.
- `go test ./db -run "TestEmpresaDocumentoCompraListAndSetEstadoByCodigo" -count=1` (ok).
- `go test ./handlers -run "TestEmpresaComprasDocumentos" -count=1` (ok).
- `go test ./db ./handlers -run "TestEmpresaDocumentoCompraListAndSetEstadoByCodigo|TestEmpresaComprasDocumentos" -count=1` (ok).
- `go test ./... -run "TestEmpresaComprasDocumentos|TestEmpresaDocumentoCompraListAndSetEstadoByCodigo" -count=1` (ok).

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

### Punto 11. Reportes financieros (continuacion 2026-04-05 - dataset de flujo de caja diario)

Implementacion tecnica completada:
- Se extiende el endpoint central `GET /api/empresa/reportes?action=dataset` con el dataset `contable_flujo_caja`:
	- `backend/handlers/reportes.go` agrega `contable_flujo_caja` al catalogo de datasets empresariales.
	- El dataset consolida movimientos financieros por fecha y calcula:
		- `ingresos`,
		- `egresos`,
		- `neto_dia`,
		- `saldo_acumulado`,
		- `movimientos`.
	- El resumen del periodo incluye `total_ingresos`, `total_egresos`, `neto_periodo`, `saldo_final`, promedio diario y cantidad de dias con movimiento.
	- Mantiene exportacion homologada del modulo en `pdf`, `xls`, `csv`, `json` y `txt`.

Cobertura de pruebas agregada/extendida:
- `backend/handlers/reportes_test.go` agrega `TestEmpresaReportesHandlerDatasetContableFlujoCaja` para validar:
	- agregacion diaria,
	- saldo acumulado por dia,
	- totales del resumen en el periodo.

Validacion ejecutada:
- `runTests` en `backend/handlers/reportes_test.go` (5 passed, 0 failed).

### Punto 11. Reportes financieros (continuacion 2026-04-05 - filtros contables en flujo de caja diario)

Implementacion tecnica completada:
- Se extiende el dataset `contable_flujo_caja` para analitica segmentada por atributos contables:
	- `backend/handlers/reportes.go` incorpora filtros `categoria` y `metodo_pago` en el endpoint `GET /api/empresa/reportes?action=dataset`.
	- El flujo de consolidacion diaria aplica filtros antes de calcular `ingresos`, `egresos`, `neto_dia` y `saldo_acumulado`.
	- El resumen del dataset agrega `filtro_categoria` y `filtro_metodo_pago` para trazabilidad inter-formato en exportaciones.
- `web/administrar_empresa/reportes.html` agrega controles de filtro contable y envia estos parametros en consultas y exportaciones del modulo.

Cobertura de pruebas agregada/extendida:
- `backend/handlers/reportes_test.go` agrega `TestEmpresaReportesHandlerDatasetContableFlujoCajaFiltros` para validar:
	- agregacion diaria con filtro por categoria,
	- filtro por metodo de pago,
	- consistencia de totales y resumen.

Validacion ejecutada:
- `runTests` en `backend/handlers/reportes_test.go` (6 passed, 0 failed).

### Punto 3. Permisos y seguridad (continuacion 2026-04-05 - endpoint de contexto de permisos)

Implementacion tecnica completada:
- Se agrega endpoint de lectura para inspeccionar permisos efectivos por rol en la operacion empresarial:
	- `backend/handlers/empresa_permisos.go` incorpora `GET /api/empresa/permisos_contexto`.
	- La respuesta incluye permisos por modulo/accion (`R/C/U/D/A`) y resumen de capacidad efectiva del rol autenticado.
	- Se soporta `include_matrix=1` para exponer matriz comparativa por roles canonicos (`super_administrador`, `admin_empresa`, `supervisor_sucursal`, `cajero`, `inventario`, `compras`, `contabilidad`, `auditor`).
- `backend/main.go` registra la ruta bajo `WithEmpresaSeguridadPermissions` para conservar aislamiento por `empresa_id` y control de seguridad.

Cobertura de pruebas agregada/extendida:
- `backend/handlers/empresa_permisos_test.go` agrega:
	- `TestEmpresaPermisosContextoHandlerRetornaPermisosPorRol`.
	- `TestEmpresaPermisosContextoHandlerIncluyeMatrizRoles`.

Validacion ejecutada:
- `go test ./handlers -run "PermisosContexto|WithEmpresa.*Permissions" -count=1` (ok).

### Punto 3. Permisos y seguridad (continuacion 2026-04-05 - menu dinamico por permisos efectivos)

Implementacion tecnica completada:
- Se completa el consumo frontend del contexto de permisos para cierre operativo del Punto 3:
	- `web/js/administrar_empresa.js` consume `GET /api/empresa/permisos_contexto?empresa_id={id}` para resolver visibilidad real de enlaces del menu lateral por modulo/accion.
	- Se mantiene fallback local por rol para continuidad cuando no hay respuesta del endpoint.
	- `web/administrar_empresa.html` agrega `menuPermsEvidence` para evidencia UAT visual de rol y fuente de permisos aplicada.

Validacion ejecutada:
- `get_errors` sobre `web/js/administrar_empresa.js` y `web/administrar_empresa.html` (sin errores).

### Punto 3. Permisos y seguridad (continuacion 2026-04-05 - regresion UAT en endpoints sin wrapper)

Implementacion tecnica completada:
- Se refuerza cobertura de regresion en endpoints sin wrapper de modulo para validar aislamiento por `empresa_id` y cuenta Google autenticada:
	- `backend/handlers/auth_users_carritos_test.go` agrega:
		- `TestEmpresaUsuarioLoginHandlerRejectsWrongEmpresaScopeFromQuery`.
		- `TestEmpresaUsuarioSetPasswordHandlerRejectsWrongEmpresaScopeFromQuery`.
	- `backend/handlers/chat_con_inteligencia_artificial_controller_test.go` agrega:
		- `TestModeloPreferidoHandlerGetRejectsEmpresaFueraDeAlcanceByGoogleAccount`.
		- `TestModeloPreferidoHandlerPutRejectsEmpresaFueraDeAlcanceByGoogleAccount`.
		- `TestHistorialHandlerRejectsEmpresaFueraDeAlcanceByGoogleAccount`.

Validacion ejecutada:
- `go test ./handlers -run "EmpresaUsuario(LoginHandlerRejectsWrongEmpresaScope|SetPasswordHandlerRejectsWrongEmpresaScope)|ModelosHandlerRejectsEmpresaFueraDeAlcanceByGoogleAccount|ConsultarHandlerRejectsEmpresaFueraDeAlcance|ModeloPreferidoHandler(Get|Put)RejectsEmpresaFueraDeAlcanceByGoogleAccount|HistorialHandlerRejectsEmpresaFueraDeAlcanceByGoogleAccount" -count=1` (ok).

## Backlog inmediato (siguiente iteracion)

1. Cerrar Punto 3 (permisos y seguridad): definir e implementar politica final de lectura sensible en auditoria (`/api/empresa/auditoria/eventos`) por rol (`auditor`, `admin_empresa`, `super_administrador`) y su evidencia UAT automatizada.
2. Iniciar Punto 5 (control de inventarios): formalizar kardex operativo, reglas de stock min/max y alertas de quiebre por bodega.

## Criterios de avance para la siguiente fase

- Punto 1 queda en estado completo cuando exista una matriz formal de KPI con formulas y fuente de datos por endpoint/tabla.
- Punto 2 queda en estado completo cuando exista matriz de entidades con llaves de aislamiento (empresa/sucursal/bodega) y reglas de validacion por endpoint.
- Punto 10 queda en estado completo cuando exista proceso documentado y probado para convertir eventos en asientos con referencia canonica de documento (`entidad_id`) y ejecucion automatica controlada.
- Punto 15 queda en estado completo cuando el registro de auditoria por empresa cubra acciones criticas, tenga consulta segura por rol y cuente con pruebas automatizadas de trazabilidad.
