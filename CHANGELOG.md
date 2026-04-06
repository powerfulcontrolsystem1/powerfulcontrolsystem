# CHANGELOG

## 2026-04-06
- Hotfix de compatibilidad de migraciones legacy en startup:
	- Se corrige el orden de migracion en `EnsureEmpresaPropinasSchema` para crear indices despues de asegurar columnas faltantes (`cierre_caja_id` y relacionadas), evitando fallos en bases antiguas.
	- Se corrige el orden de migracion en `EnsureEmpresaComisionesServicioSchema` para crear indices despues de asegurar columnas faltantes (`ajuste_manual` y relacionadas), evitando fallos en bases antiguas.
	- Resultado operativo: el script `scripts/iniciar_servidor.ps1` vuelve a iniciar correctamente y el backend queda escuchando en `:8080`.
	- Validaciones ejecutadas:
		- `go test ./db -run "Propina|Comision" -count=1` (OK).
		- `go test ./handlers -run "Propina|Comision" -count=1` (OK).
		- `go test ./... -run TestDoesNotExist -count=1` (compilacion global backend OK).
- Cierre tecnico del modulo 15 (Comisiones por servicio):
	- Se amplía el modelo de comisiones con escalas por rol/servicio y tope por item:
		- nueva tabla `empresa_comisiones_servicio_escalas` (`rol_operacion`, `servicio_filtro`, `porcentaje_comision`, `tope_comision`, `prioridad`).
	- Se amplía `empresa_comisiones_servicio_movimientos` con trazabilidad operativa:
		- `rol_operacion`, `escala_id`, `monto_comision_bruto`, `tope_comision_aplicado`,
		- `origen_movimiento`, `ajuste_manual`, `referencia_ajuste`, `ajuste_estado`, `aprobado_por`, `aprobado_en`,
		- `liquidacion_nomina_id`, `periodo_liquidacion_desde`, `periodo_liquidacion_hasta`, `liquidado_en`, `liquidado_por`.
	- Se incorporan endpoints/acciones de comisiones para operacion avanzada:
		- escalas (`escalas`, `escala`, `activar_escala`, `desactivar_escala`),
		- ajuste manual y aprobacion (`ajuste_manual`, `aprobar_ajuste`, `rechazar_ajuste`),
		- resumen para nomina (`resumen_liquidacion`).
	- Se integra nomina con comisiones:
		- `empresa_nomina_liquidaciones` incorpora `comisiones_servicio_total`, `comisiones_servicio_movimientos`, `comisiones_servicio_ajustes`.
		- el calculo de liquidacion integra comisiones y enlaza movimientos al periodo liquidado.
	- Se amplia `web/administrar_empresa/comisiones.html` para operacion completa del modulo 15:
		- gestion de escalas/topes,
		- registro de ajuste manual,
		- aprobacion/rechazo de ajustes pendientes,
		- filtros avanzados de reporte y consulta de `resumen_liquidacion`.
	- Validaciones ejecutadas:
		- `go test ./db -run "TestEmpresaComisionesServicio|TestEmpresaNominaLiquidacionIntegraComisionesServicio" -count=1` (OK).
		- `go test ./handlers -run "TestEmpresaComisionesServicioHandler" -count=1` (OK).
		- `go test ./... -run TestDoesNotExist -count=1` (compilacion global backend OK).
- Avance y cierre tecnico del modulo 14 (Propinas):
	- Se amplía la configuracion empresarial de propinas con reglas fiscales:
		- `pais_fiscal`, `regimen_fiscal`, `tratamiento_fiscal` (`no_gravada`/`gravada`) y `porcentaje_impuesto_propina`.
	- Se amplía el libro de movimientos de propinas con:
		- `origen_movimiento` (`venta`/`ajuste_manual`),
		- `ajuste_manual`, `referencia_ajuste`, `cierre_caja_id`, `conciliado_en`,
		- snapshot fiscal por movimiento (`fiscal_*`).
	- Se incorpora conciliacion de propinas contra cierre de caja:
		- accion manual `action=conciliacion_cierre` en propinas,
		- integracion automatica durante transiciones `cerrar/aprobar` de cierre de caja,
		- persistencia de resumen en `empresa_cierres_caja` (`propinas_movimientos`, `propinas_total`, `propinas_ajustes`, `propinas_impuesto`, `propinas_neto`, `propinas_conciliado_*`).
	- Se incorpora ajuste manual auditado de propinas:
		- accion `action=ajuste_manual`,
		- registro no bloqueante en `empresa_auditoria_eventos`.
	- Se actualiza frontend `web/administrar_empresa/propinas.html` con:
		- configuracion fiscal,
		- formulario de ajuste manual,
		- accion de conciliacion por cierre,
		- filtros y columnas extendidas en el reporte.
	- Se agrega cobertura de pruebas para flujo de ajuste y conciliacion:
		- `backend/handlers/propinas_test.go`.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "Propinas|Cierre" -count=1` (OK).
		- `go test ./db -run "Propina|CierreCaja|Finanzas" -count=1` (OK).
		- `go test ./... -run "^$" -count=1` (compilacion global backend OK).
- Avance del modulo 13 (Codigos de descuento avanzados):
	- Se amplía `codigos_de_descuento` con reglas contextuales:
		- `segmento_cliente`, `canal_venta`, `horario_desde`, `horario_hasta`, `dias_semana`.
	- Se incorpora antifraude por cliente:
		- `max_usos_por_cliente`, `ventana_horas_fraude`.
	- Se agrega trazabilidad de redenciones en nueva tabla `codigos_descuento_redenciones` con estados:
		- `aplicada`, `revertida`, `anulada`.
	- Se integra ciclo de redencion con carritos:
		- aplica al cerrar carrito,
		- revierte al reabrir,
		- anula al eliminar carrito.
	- Se extiende API de codigos:
		- validacion contextual (`action=validar` con `carrito_id`, `cliente_id`, `canal_venta`),
		- consulta de trazabilidad (`action=redenciones`).
	- Se actualiza `web/administrar_empresa/codigos_de_descuento.html` para administrar reglas avanzadas y antifraude.
	- Validaciones ejecutadas:
		- `go test ./db -run "TestCodigoDescuento" -count=1` (OK).
		- `go test ./handlers -run "TestNoExiste" -count=1` (OK, compilacion handlers).
		- `go test ./db -run "TestCarritoProductoDescuentaInventarioYVentaMantieneStock|TestCarritoStockNoSeDuplicaAlReactivarSesionCerrada" -count=1` (OK).
		- `go test ./...` (falla en prueba no relacionada: `TestEmpresaGraficosEstadisticasHandlerPanelYAcciones`).
- Validacion final de continuidad tecnica y documental (post-ediciones recientes):
	- Se revalida compilacion global de backend tras ajustes recientes en inventario/combos.
	- Resultado: `go test ./... -run TestDoesNotExist -count=1` en `backend` -> OK (sin errores de compilacion).
	- Se confirma sincronizacion de cierre de modulos 1-12 en documentacion operativa y tecnica.
- Cierre del modulo 12 (Combos de productos):
	- Se implementa versionado de receta por combo:
		- nuevas columnas en `combos_productos`: `receta_version`, `costo_teorico`, `costo_real`, `variacion_costo`, `variacion_costo_porcentaje`.
		- nueva tabla `combos_productos_versiones` para snapshots historicos de ingredientes por version.
	- Se incorpora validacion de costo teorico vs costo real de ingredientes en create/update de combos:
		- bloqueo si la variacion porcentual supera el umbral operativo.
		- bloqueo si el precio del combo no cubre el costo real calculado.
	- Se endurece concurrencia de inventario en carritos:
		- reserva de stock con `UPDATE` atomico condicionado (`cantidad >= requerida`) para evitar sobreventa en ventas simultaneas.
	- Se actualiza frontend `web/administrar_empresa/combos_productos.html` para mostrar version de receta y metricas de costo.
	- Validaciones ejecutadas:
		- `runTests` sobre `backend/db/productos_categorias_test.go` y `backend/db/carritos_inventario_test.go`.
		- `go test ./... -run TestDoesNotExist -count=1`.
- Verificacion final de continuidad del modulo 11 (Inventario):
	- Se ejecuta compilacion global posterior a cambios recientes en archivos de inventario con `go test ./... -run TestDoesNotExist -count=1` (OK).
	- Se confirma cierre operativo completo del checklist de modulo 11 (schema, costos, conteo ciclico, alertas proactivas y documentacion).
- Cierre del modulo 11 (Inventario) de Fase 3:
	- Se implementa configuracion de politica de costo por empresa:
		- `GET/PUT /api/empresa/inventario/configuracion`.
		- Politicas soportadas: `promedio` y `peps`.
	- Se incorpora soporte de capas/lotes de costo para trazabilidad de salidas y transferencias:
		- tabla `inventario_costos_lotes`.
		- salida con PEPS por capas y recalculo de costo promedio por bodega/producto.
	- Se implementa conteo ciclico con ajuste auditado:
		- `GET/POST /api/empresa/inventario/conteo_ciclico`.
		- tabla `inventario_conteos_ciclicos` y movimiento automatico `ajuste_positivo/ajuste_negativo` cuando hay variacion.
	- Se cierran alertas operativas proactivas de inventario:
		- `GET /api/empresa/inventario/alertas?action=proactivas`.
		- incorpora `sobrestock`, `deficit`, `exceso` y `accion_sugerida`.
	- Se actualiza frontend `web/administrar_empresa/administrar_productos.html` con:
		- selector/guardado de politica de costo,
		- formulario y tabla de conteo ciclico,
		- visualizacion de alertas proactivas (quiebre/sobrestock).
	- Validaciones ejecutadas:
		- `go test ./db -run "TestInventarioPoliticaCostoPromedioYPEPS|TestRegistrarConteoCiclicoInventarioAjustaYAudita|TestGetAlertasOperativasByEmpresaIncluyeSobrestock" -count=1`.
		- `go test ./handlers -run "TestEmpresaInventarioConfiguracionYConteoCiclicoHandler|TestEmpresaInventarioAlertasHandlerProactivasIncluyeSobrestock" -count=1`.
		- `go test ./... -run TestDoesNotExist -count=1`.
- Cierre del modulo 10 (Clientes) de Fase 3:
	- Se implementa deduplicacion por `documento`, `correo` y `telefono` en `create/update` de clientes por `empresa_id`.
	- El endpoint `POST/PUT /api/empresa/clientes` responde `409` cuando detecta conflicto de deduplicacion, con mensaje de campo duplicado.
	- Se agrega dataset operativo para exportacion masiva comercial:
		- `operativo_clientes_segmentacion_comercial` en `/api/empresa/reportes`.
		- Incluye segmento, metricas de compra y `accion_comercial_sugerida` por cliente.
		- Exportacion disponible en `pdf/xls/csv/json/txt`.
	- Se actualiza frontend `web/administrar_empresa/administrar_clientes.html` con panel de exportacion masiva por segmento.
	- Validaciones ejecutadas:
		- `go test ./db -run "Test(CreateClienteDeduplicacionDocumentoCorreoTelefono|UpdateClienteDeduplicacionCorreoTelefono|GetClientePerfilComercialByEmpresaAndHistorial|GetClientePerfilComercialByEmpresaSinComprasSegmentoNuevo|GetClienteByID)$" -count=1`.
		- `go test ./handlers -run "Test(EmpresaClientesHandlerPerfilHistorialSegmentacion|EmpresaClientesHandlerConflictosDeduplicacion|EmpresaReportesHandlerDatasetOperativoClientesSegmentacionComercial)$" -count=1`.
		- `go test ./... -run "^$" -count=1`.
- Cierre del modulo 9 (Tarifas por dia) de Fase 3:
	- Se implementa prorrateo de tarifa diaria por ventana de `check-in/check-out` para entrada/salida fuera de ventana.
	- Se extiende simulador de `GET /api/empresa/tarifas_por_dia?action=calcular` con detalle de:
		- `dias_completos`, `dias_equivalentes`,
		- `monto_dias_completos`, `monto_prorrateo_(entrada|intermedio|salida)`,
		- `minutos_prorrateo_fuera_ventana`.
	- Se agrega aplicacion masiva de una misma tarifa diaria a todas las estaciones detectadas:
		- `PUT /api/empresa/tarifas_por_dia?action=aplicar_todas_estaciones`.
	- Se agrega reporte operativo comparativo por estacion:
		- dataset `operativo_tarifas_comparativo_estaciones` en `/api/empresa/reportes`,
		- comparativo de ingreso esperado (motor prorrateado) vs ingreso real cobrado,
		- exportacion en `pdf/xls/csv/json/txt`.
	- Se actualiza frontend `web/administrar_empresa/tarifas_por_dia.html` con:
		- boton `Aplicar a todas las estaciones`,
		- simulador con desglose de prorrateo,
		- panel de descarga del comparativo esperado vs real.
	- Validaciones ejecutadas:
		- `go test ./db -run "TarifaPorDia|ApplyEmpresaTarifaPorDiaToAllStations|EmpresaTarifasPorDia"`.
		- `go test ./handlers -run "TarifasPorDia|CarritosCompraListIncluyeTarifaPorDiaAutomatica|OperativoTarifasIngresos|OperativoTarifasComparativoEstaciones"`.
		- `go test ./... -run "^$"`.
- Cierre del modulo 8 (Tarifas por minutos) de Fase 3:
	- Se agrega configuracion empresarial avanzada de calculo:
		- `redondeo_modo` (`ninguno`, `arriba`, `abajo`, `matematico`),
		- `redondeo_unidad`,
		- `monto_minimo_diario`,
		- `monto_maximo_diario`.
	- Se extiende simulador de cobro por minutos con detalle de:
		- monto base, monto extra, subtotal, monto redondeado y ajuste,
		- aplicacion de minimo/maximo diario,
		- soporte de minutos fraccionarios (`minutos_consumidos` decimal).
	- Se cierra trazabilidad contable del calculo por minutos:
		- registro de evento `finanzas.tarifa_por_minutos_calculada` en `empresa_eventos_contables`,
		- respuesta de simulacion con `trazabilidad_contable_id`, `documento_codigo` y `periodo_contable`.
	- Se agrega aplicacion masiva de una misma regla de tarifa a todas las estaciones detectadas:
		- `PUT /api/empresa/tarifas_por_minutos?action=aplicar_todas_estaciones`.
	- Se actualiza frontend `web/administrar_empresa/tarifas_por_minutos.html` con:
		- panel de configuracion avanzada de redondeo y topes,
		- boton `Aplicar a todas las estaciones`,
		- simulador con detalle de calculo y referencia contable.
	- Validaciones ejecutadas:
		- `go test ./db -run "TestEmpresaTarifasPorMinutos|TestApplyEmpresaTarifaPorMinutosToAllStations|TestRegisterTarifaPorMinutosCalculoContable|TestEmpresaEventosContables" -count=1`.
		- `go test ./handlers -run "TestEmpresaTarifasPorMinutosHandler" -count=1`.
		- `go test ./... -run "^$" -count=1`.
- Cierre del modulo 7 (Reservas por estacion/habitacion) de Fase 3:
	- Se refuerza control de concurrencia anti-overbooking por estacion en ventanas solapadas:
		- validacion de conflicto por `estacion_id` y `carrito_id` asociado,
		- bloqueo para estados operativos `pendiente_pago`, `confirmada` y `en_curso`.
	- Se implementa politica automatica avanzada de reservas:
		- expiracion de pendientes por `fecha_expiracion` y fallback por antiguedad de creacion,
		- marcacion automatica de `no_show` sobre reservas confirmadas fuera de tolerancia operativa,
		- accion de sincronizacion: `GET /api/empresa/reservas_hotel?action=aplicar_politicas`.
	- Se incorpora reconversion operativa de reserva a carrito:
		- `PUT /api/empresa/reservas_hotel?action=convertir_carrito`.
		- transicion de reserva a estado `en_curso` y activacion de carrito asociado.
	- Se actualiza frontend `web/administrar_empresa/reservas_hotel.html` con:
		- accion `Aplicar politicas`,
		- accion `Reconver. carrito`,
		- filtros extendidos para estados `en_curso` y `no_show`.
	- Validaciones ejecutadas:
		- `go test ./db -run "TestReservaHotel(FlowCRUDAndDisponibilidad|MultiEstacionNoOverbookingYReconversion|PoliticaNoShowYExpiracionAvanzada)$" -count=1`.
		- `go test ./handlers -run "TestEmpresaReservasHotelHandler(CRUDAndDisponibilidad|PoliticasYReconversion)$" -count=1`.
		- `go test ./... -run "^$" -count=1`.
- Cierre del modulo 6 (Registro de vehiculos) de Fase 3:
	- Se agrega configuracion de validacion de placa/patente por empresa y pais:
		- `GET/PUT /api/empresa/vehiculos_registro?action=config`.
		- Tabla `empresa_vehiculos_configuracion` con `pais_codigo`, `patente_regex`, `patente_descripcion`, `evitar_duplicado_activo`.
	- Se implementa bloqueo de duplicidad activa por patente canonica en patio/empresa:
		- validado en crear, editar y activar registros de vehiculos.
		- respuesta HTTP `409` ante conflicto de duplicidad activa.
	- Se agrega reporte operativo de permanencia y tiempos de estancia:
		- `GET /api/empresa/vehiculos_registro?action=permanencia`.
		- dataset `operativo_vehiculos_permanencia` en `/api/empresa/reportes` con exportacion `pdf/xls/csv/json/txt`.
	- Se integra frontend en `web/administrar_empresa/vehiculos_registro.html`:
		- panel de configuracion de formato de placa por pais,
		- consulta visual de permanencia,
		- exportacion de reporte en formatos estandar.
	- Validaciones ejecutadas:
		- `go test ./db -run TestEmpresaVehiculoRegistroConfigValidacionDuplicidadYPermanencia -count=1`.
		- `go test ./handlers -run TestEmpresaVehiculosRegistroHandlerConfigYReportePermanencia -count=1`.
		- `go test ./handlers -run TestEmpresaReportesHandlerDatasetOperativoVehiculosPermanencia -count=1`.
		- `go test ./... -run "^$" -count=1`.
- Cierre del modulo 5 (Nomina de sueldos) de Fase 3:
	- Se agregan operaciones nuevas en nomina:
		- `GET /api/empresa/nomina?action=desprendible&empleado_nomina_id={id}&periodo_desde=YYYY-MM-DD&periodo_hasta=YYYY-MM-DD`.
		- `GET /api/empresa/nomina?action=conciliacion_asistencia` (auditoria sin cambios).
		- `POST /api/empresa/nomina?action=conciliar_asistencia` (auditoria con opcion de auto-recalculo).
	- Se implementa desprendible estandar por empleado y periodo con detalle de horas, devengados, deducciones y neto a pagar.
	- Se implementa conciliacion automatica entre asistencia y liquidacion final:
		- detecta diferencias de registros/horas,
		- identifica asistencias sin liquidacion,
		- permite recalcular/crear liquidaciones inconsistentes cuando `auto_recalcular=true`.
	- Se integra frontend en `web/administrar_empresa/nomina_sueldos.html`:
		- boton de conciliacion con modo auditoria o auto-recalculo,
		- generacion/visualizacion de desprendible por empleado-periodo,
		- accion de desprendible desde tabla de liquidaciones.
	- Se documentan y validan casos de formula por pais/empresa (CO/MX + override por empresa) con pruebas automatizadas.
	- Validaciones ejecutadas:
		- `go test ./db -run "Test(EmpresaNominaGenerateLiquidacionesFromAsistencia|EmpresaNominaCalculoPorPaisYEmpresa|EmpresaNominaDesprendibleYConciliacionAsistencia)$" -count=1`.
		- `go test ./handlers -run "TestEmpresaNominaSueldosHandlerFlow$" -count=1`.
		- `go test ./... -run "^$" -count=1`.
- Cierre del modulo 4 (Asistencia de empleados) de Fase 5:
	- Se implementa cierre de periodo con bloqueo operativo de edicion posterior:
		- `POST /api/empresa/asistencia_empleados?action=cerrar_periodo`.
		- `GET /api/empresa/asistencia_empleados?action=periodos_cerrados`.
	- Se agrega configuracion por empresa para tolerancias y reglas de turno:
		- `GET/PUT /api/empresa/asistencia_empleados?action=config`.
		- `tolerancia_entrada_minutos`, `hora_inicio_turno_(manana|tarde|noche)`, `permitir_turno_nocturno`, `permitir_turno_cruzado`.
	- Se incorporan validaciones de negocio en asistencia:
		- bloqueo de create/update/delete/activar/desactivar/marcar_entrada/marcar_salida cuando la fecha pertenece a periodo cerrado,
		- rechazo de turno nocturno o cruzado cuando la configuracion empresarial lo deshabilita,
		- calculo de tardanza con tolerancia configurable.
	- Se publica reporte operativo de auditoria para nomina:
		- dataset `operativo_asistencia_nomina_auditoria` en `/api/empresa/reportes` con exportacion `pdf/xls/csv/json/txt`.
	- Se integra frontend en `web/administrar_empresa/asistencia_empleados.html`:
		- panel de configuracion,
		- cierre de periodo y listado de cierres,
		- descarga del reporte de auditoria de nomina.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "Test(EmpresaAsistenciaEmpleadosHandlerCRUDFlow|EmpresaAsistenciaEmpleadosHandlerConfigTurnosYTolerancia|EmpresaAsistenciaEmpleadosHandlerCierrePeriodoBloqueaEdicion|EmpresaReportesHandlerDatasetOperativoAsistenciaNominaAuditoria)$" -count=1`.
		- `go test ./... -run "^$" -count=1`.
- Cierre del modulo 3 (Usuarios de empresa) de Fase 1:
	- Se agrega cambio autogestionado de contraseña para usuario empresa:
		- `POST /api/empresa/usuarios/cambiar_password`.
	- Se implementan politicas de contraseña configurables desde `configuraciones`:
		- `usuarios.password_min_length`
		- `usuarios.password_require_uppercase`
		- `usuarios.password_require_lowercase`
		- `usuarios.password_require_digit`
		- `usuarios.password_require_symbol`
		- `usuarios.password_rotation_days`.
	- El login de usuario empresa ahora devuelve `password_rotation_required` cuando aplica rotacion obligatoria.
	- Se incorpora captura de notificaciones de confirmacion/restablecimiento en entorno de pruebas de correo:
		- tabla `super_correo_notificaciones_prueba` en `superadministrador.db`.
		- activacion por `PCS_MAIL_TEST_MODE=1` o `gmail.smtp_test_mode=1`.
	- Se integra frontend de autogestion en `web/login_usuario.html` y `web/js/login_usuario.js`.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "Test(EmpresaUsuarioChangePasswordFlow|EmpresaUsuarioChangePasswordPolicyRejectsWeakPassword|EmpresaUsuarioLoginRequiresRotationWhenPolicyEnabled|EmpresaUsuarioNotificationsCaptureInMailTestMode|EmpresaUsuarioPasswordRecoveryFlow)" -count=1`.
		- `go test ./... -run "^$" -count=1`.
- Cierre del modulo 2 (Administracion global super) de Fase 1:
	- Se implementa desactivacion/rehabilitacion de empresa con validaciones de impacto operativo y confirmacion forzada cuando existen bloqueos:
		- `GET /super/api/empresas?id={id}&action=impacto_desactivacion`.
		- `PUT /super/api/empresas?id={id}&action=desactivar[&force=1]`.
		- `PUT /super/api/empresas?id={id}&action=activar&activo=1`.
	- Se agrega respaldo/restauracion de configuracion critica super:
		- `GET /super/api/config/backup` (exporta JSON).
		- `PUT /super/api/config/backup` (restaura JSON).
	- Se integra operacion desde frontend:
		- `web/js/seleccionar_empresa.js` para desactivar/reactivar con consulta de impacto.
		- `web/super/configuracion_avanzada.html` con descarga y restauracion de respaldo.
	- Se agregan pruebas de permisos y flujo super en `backend/handlers/system_empresas_handlers_test.go`.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "Test(EmpresasHandlerDesactivarConImpactoYForce|EmpresasHandlerImpactoDesactivacion|SuperConfigBackupHandlerExportYRestore|SuperEndpointsPermisosPorRol)" -count=1`.
		- `go test ./... -run "^$" -count=1`.
- Cierre del modulo 1 (Autenticacion y sesiones) de Fase 1:
	- Se implementa bloqueo temporal por intentos fallidos en login de usuario empresa.
	- Se agrega recuperacion de contrasena para usuario empresa con token temporal:
		- `POST /api/empresa/usuarios/solicitar_recuperacion_password`
		- `POST /api/empresa/usuarios/restablecer_password`
	- Se endurece seguridad de sesion:
		- sesiones nuevas con `fecha_fin` por expiracion (24h),
		- revocacion de token en logout,
		- middleware bloquea tokens expirados o revocados.
	- Se habilita flujo frontend de recuperacion en `web/login_usuario.html` y `web/js/login_usuario.js`.
	- Validaciones ejecutadas:
		- `runTests` sobre `backend/handlers/auth_users_carritos_test.go` -> 24/24.
		- `go test ./... -run "^$" -count=1` (compilacion global OK).
- Cierre tecnico backend de pasarela unica Wompi:
	- Se elimina remanente de Mercado Pago en backend:
		- `backend/handlers/payments_handlers.go`: retiro de handlers/utilidades Mercado Pago.
		- `backend/db/db.go`: retiro de tipo/funciones de persistencia `pagos_mercadopago`.
		- `backend/main.go`: retiro de bootstrap/migracion de `pagos_mercadopago`.
		- `backend/utils/utils.go`: retiro del prefijo `/mercadopago/` en manejo JSON API.
		- `backend/tools/query_users/main.go`: migracion de inspeccion local hacia `wompi.*` y `pagos_wompi`.
	- Se sincroniza documentacion tecnica con el estado real:
		- `documentos/estructura_bd.md` y `estructura_bd.md`.
		- `documentos/diagramas/estructura_del_codigo.md`.
		- `documentos/descripcion_de_archivos`.
	- Validacion tecnica ejecutada: `go test ./... -run "^$" -count=1` (compilacion global OK).
- Cierre de pendientes de modulos:
	- Se valida la matriz de estado de modulos/reportes y no quedan modulos marcados como incompletos (`Pendiente` o `Parcial`) en `documentos/modulos del proyecto.md`.
	- Se actualiza `Pendiente Notas` marcando como completado el pendiente de pasarela unica Wompi.
- Pasarela de pago unificada en Wompi:
	- Se retira la configuración de Mercado Pago de `web/super/configuracion_avanzada.html` y se deja únicamente la sección de credenciales de Wompi en configuración avanzada del panel super administrador.
	- Se simplifica `web/pagar_licencia.html` eliminando selector/panel/flujo de Mercado Pago para operar solo con Nequi (Wompi) y activación manual interna.
	- Se desregistran rutas de Mercado Pago en `backend/main.go` (`/super/api/config/mercadopago`, `/mercadopago/create_preference`, `/mercadopago/webhook`, `/mercadopago/reconcile`, `/mercadopago/test_preference`).
	- Validación técnica ejecutada: `go test ./... -run "^$" -count=1` (compilación global OK).
- Cierre de trazabilidad y validacion final del plan de reportes:
	- Se revalida la presencia de los datasets operativos de cierre (`operativo_propinas_acumulado`, `operativo_comisiones_lavador`, `operativo_facturacion_trazabilidad`, `operativo_auditoria_acciones`) en `backend/handlers/reportes.go`.
	- Se ejecuta validacion completa de `backend/handlers/reportes_test.go` con resultado `16/16` pruebas aprobadas.
	- Se confirma consistencia de estado documental en `documentos/modulos del proyecto.md`, `CHANGELOG.md` y `documentos/historial_de_cambios`.
	- Se deja cerrado el pendiente de trazabilidad del plan secuencial.
- Plan secuencial de cierre de modulos incompletos - bloques 6 a 9 (Propinas, Comisiones, Facturacion y Auditoria):
	- Se agregan en `backend/handlers/reportes.go` cuatro datasets operativos nuevos:
		- `operativo_propinas_acumulado` (acumulado por usuario, distribucion directa/universal y participacion),
		- `operativo_comisiones_lavador` (acumulado por lavador con base de servicios y ticket de comision),
		- `operativo_facturacion_trazabilidad` (emitidas/anuladas/pendientes y trazabilidad legal por tipo documental),
		- `operativo_auditoria_acciones` (eventos por modulo/usuario con errores HTTP y acciones criticas).
	- Se actualiza catalogo y switch de construccion de datasets para incluir estos cuatro reportes en suite/export.
	- Se amplia `backend/handlers/reportes_test.go` con pruebas dedicadas:
		- `TestEmpresaReportesHandlerDatasetOperativoPropinasAcumulado`.
		- `TestEmpresaReportesHandlerDatasetOperativoComisionesLavador`.
		- `TestEmpresaReportesHandlerDatasetOperativoFacturacionTrazabilidad`.
		- `TestEmpresaReportesHandlerDatasetOperativoAuditoriaAcciones`.
	- Se refuerza `ensureEmpresaReportesSchema` con `EnsureEmpresaPropinasSchema`, `EnsureEmpresaComisionesServicioSchema` y `EnsureEmpresaAuditoriaSchema` para cobertura de suite completa.
	- Se actualiza la matriz en `documentos/modulos del proyecto.md` marcando Propinas, Comisiones, Facturacion y Auditoria como activos en reportes.
	- Validacion tecnica ejecutada:
		- `runTests` focalizado en 4 pruebas nuevas (ok).
		- `runTests` completo sobre `backend/handlers/reportes_test.go` (16/16 ok).
- Plan secuencial de cierre de modulos incompletos - bloque 5 (Compras):
	- Se rediseña el dataset `operativo_compras_movimientos` en `backend/handlers/reportes.go` para consolidar compras por proveedor, dejando de depender solo de movimientos de inventario.
	- El dataset ahora expone KPI de ciclo documental: `ordenes_emitidas`, `recepciones`, `contabilizaciones`, `monto_ordenado`, `monto_recepcionado`, `monto_contabilizado`, `brecha_monto` y cumplimiento de recepcion/monto.
	- Se actualiza el catalogo del reporte con nuevo titulo y descripcion orientados a `costo por proveedor y recepcion vs orden`.
	- Se amplia `backend/handlers/reportes_test.go` con `TestEmpresaReportesHandlerDatasetOperativoComprasMovimientos` para validar consolidado por proveedor, totales de resumen y porcentajes de cumplimiento.
	- Se actualiza la matriz en `documentos/modulos del proyecto.md` marcando Compras como activo en reportes.
	- Validacion tecnica ejecutada:
		- `runTests` sobre `backend/handlers/reportes_test.go` con 8 pruebas objetivo (ok).
- Plan secuencial de cierre de modulos incompletos - bloque 4 (Inventario):
	- Se extiende el dataset `operativo_inventario_bodega` en `backend/handlers/reportes.go` con metricas de:
		- rotacion estimada y cobertura (`salida_promedio_diaria`, `dias_cobertura`, `indice_rotacion_30d`),
		- riesgo de quiebre proyectado (`estado_proyeccion`, `sugerido_reposicion`),
		- valorizacion por producto/bodega (`valorizacion_costo`, `valorizacion_venta`).
	- Se agregan KPI de resumen de inventario (`alertas`, `deficit`, `movimientos`, cobertura y rotacion promedio).
	- Se amplia `backend/handlers/reportes_test.go` con `TestEmpresaReportesHandlerDatasetOperativoInventarioBodega`.
	- Se actualiza matriz de estado en `documentos/modulos del proyecto.md` marcando Inventario como activo en reportes.
	- Se marca el bloque 4 como completado en `Pendiente Notas`.
	- Validacion tecnica ejecutada:
		- `go test ./handlers -run "TestEmpresaReportesHandler(DatasetOperativoInventarioBodega|DatasetOperativoCadenaCumplimiento|DatasetOperativoTarifasIngresos|DatasetOperativoReservasOcupacion|DatasetOperativoModulosResumen|CatalogoSuiteDataset|Exportes)" -count=1` (ok).
- Plan secuencial de cierre de modulos incompletos - bloque 3 (CRM/Produccion/Logistica):
	- Se agrega el dataset `operativo_cadena_cumplimiento` en `backend/handlers/reportes.go` para consolidar conversion comercial y cumplimiento operativo.
	- El dataset resume por modulo (`crm_leads`, `produccion_ordenes`, `logistica_envios`) registros de rango, estados finalizados/en proceso y monto de referencia.
	- Se amplia `backend/handlers/reportes_test.go` con `TestEmpresaReportesHandlerDatasetOperativoCadenaCumplimiento`.
	- Se actualiza matriz de estado en `documentos/modulos del proyecto.md` marcando CRM/Produccion/Logistica como activo en reportes.
	- Se marca el bloque 3 como completado en `Pendiente Notas`.
	- Validacion tecnica ejecutada:
		- `go test ./handlers -run "TestEmpresaReportesHandler(DatasetOperativoCadenaCumplimiento|DatasetOperativoTarifasIngresos|DatasetOperativoReservasOcupacion|DatasetOperativoModulosResumen|CatalogoSuiteDataset|Exportes)" -count=1` (ok).
- Plan secuencial de cierre de modulos incompletos - bloque 2 (tarifas):
	- Se consolida el dataset `operativo_tarifas_ingresos` para ingresos por modelo de tarifa (`tarifa_por_dia`, `tarifa_por_minutos`, `sin_modelo`).
	- Se amplia `backend/handlers/reportes_test.go` con `TestEmpresaReportesHandlerDatasetOperativoTarifasIngresos` y bootstrap de esquemas de tarifas (`EnsureEmpresaTarifasPorDiaSchema`, `EnsureEmpresaTarifasPorMinutosSchema`).
	- Se actualiza matriz de estado en `documentos/modulos del proyecto.md` marcando tarifas como activo en reportes.
	- Se marca el bloque de tarifas como completado en `Pendiente Notas`.
	- Validacion tecnica ejecutada:
		- `go test ./handlers -run "TestEmpresaReportesHandler(DatasetOperativoTarifasIngresos|DatasetOperativoReservasOcupacion|DatasetOperativoModulosResumen|CatalogoSuiteDataset|Exportes)" -count=1` (ok).
- Plan secuencial de cierre de modulos incompletos - bloque 1 (reservas):
	- Se agrega el dataset `operativo_reservas_ocupacion` en `backend/handlers/reportes.go` para consolidar ocupacion y cumplimiento por estacion.
	- Se amplian pruebas en `backend/handlers/reportes_test.go` con `TestEmpresaReportesHandlerDatasetOperativoReservasOcupacion` y bootstrap de `EnsureEmpresaReservasHotelSchema`.
	- Se actualiza matriz de estado en `documentos/modulos del proyecto.md` marcando reservas como activo en reportes.
	- Se documenta plan secuencial de cierre en `Pendiente Notas` y se marca reservas como primer modulo completado.
	- Validacion tecnica ejecutada:
		- `go test ./handlers -run "TestEmpresaReportesHandler" -count=1` (ok).
- Continuidad de plan en reportes por modulos:
	- Se valida y consolida el dataset `operativo_modulos_resumen` en `backend/handlers/reportes.go`.
	- Se corrige una llamada interna a `reportesCountByEmpresa` en el builder de resumen por modulos para compatibilidad con la firma actual de la funcion.
	- Se amplia `backend/handlers/reportes_test.go` con:
		- bootstrap de esquema para modulos ERP extendidos,
		- prueba `TestEmpresaReportesHandlerDatasetOperativoModulosResumen` con verificacion de conteos por modulo y consistencia de `summary`.
	- Validacion tecnica ejecutada:
		- `go test ./handlers -run "TestEmpresaReportesHandler" -count=1` (ok, 7/7).
- Continuidad de plan en frontend y ayuda:
	- Se corrige doble scrollbar en panel empresa eliminando la definicion duplicada de `.admin-empresa-frame` con altura fija en `web/estilos.css`.
	- Se actualiza `web/administrar_empresa/propinas.html` para recuperar consistencia visual/operativa (layout empresa, tablas estandar e integracion de menu flotante).
	- Se amplia `web/ayuda/ayuda.html` con guias de modulos pendientes: propinas, comisiones, ERP extendido y calculadora por empresa.
- Se continua el plan con dos faltantes operativos:
	- Utilidad nueva `web/administrar_empresa/calculadora.html` con contexto por empresa (`empresa_id`), memoria/historial aislados por empresa y exportacion JSON del historial.
	- Documento nuevo `documentos/modulos del proyecto.md` con inventario de modulos, conteo total y matriz base modulo -> reportes recomendados.
- Se integra la calculadora en navegacion:
	- `web/menu.js` agrega enlace `Calculadora` en menu flotante y propaga `empresa_id`.
	- `web/administrar_empresa.html` agrega enlace lateral `Calculadora`.
	- `web/js/administrar_empresa.js` incorpora `linkCalculadora` en navegacion y permisos (`finanzas/read`).
	- `web/estilos.css` agrega estilos `calc-*` para la nueva pantalla.
- Se actualiza documentacion tecnica:
	- `documentos/diagramas/estructura_del_codigo.md`.
	- `documentos/descripcion_del_proyecto`.
	- `documentos/descripcion_de_archivos`.
- Se completan faltantes de cobertura para la maquina de estados documental en ventas y CRM:
	- `backend/handlers/modulos_faltantes_test.go` amplía pruebas para:
		- ventas: `pedidos` y `devoluciones` (transiciones validas e invalidas),
		- CRM: `interacciones` y `campanas` (pipeline basico con validacion de transiciones).
	- Validacion tecnica ejecutada:
		- `go test ./handlers -run "TestEmpresa(IntegracionesAPIsHandlerHealthAndSync|IntegracionesBancosHandlerSyncAndEstado|VentasCotizacionesStateMachine|CRMLeadsStateMachine|VentasPedidosStateMachine|VentasDevolucionesStateMachine|CRMInteraccionesStateMachine|CRMCampanasStateMachine)" -count=1` (ok).
		- `go test ./... -run "^$" -count=1` (ok).
- Se implementan integraciones API/Bancos ejecutables y maquina de estados documental en CRM/Ventas:
	- `backend/handlers/modulos_faltantes.go` agrega handlers especializados sobre CRUD base:
		- Integraciones: `action=health_check`, `action=sync_manual`, `action=estado` en `/api/empresa/integraciones/apis` y `/api/empresa/integraciones/bancos`.
		- CRM/Ventas documentales: `action=estado`, `action=transiciones`, `action=transicionar` en rutas `/api/empresa/crm/*` y `/api/empresa/ventas/*`.
	- `backend/handlers/modulos_faltantes_test.go` (nuevo) cubre:
		- health/sync de integraciones,
		- transiciones validas/inválidas de cotizaciones y leads.
	- `web/js/modulos_erp_extendido.js` agrega botones operativos por fila:
		- Integraciones: `Health`, `Sync`, `Estado`.
		- CRM/Ventas documentales: `Transiciones`, `Transicionar`.
	- Validacion tecnica ejecutada:
		- `go test ./handlers -run "TestEmpresa(IntegracionesAPIsHandlerHealthAndSync|IntegracionesBancosHandlerSyncAndEstado|VentasCotizacionesStateMachine|CRMLeadsStateMachine)" -count=1` (ok).
		- `go test ./... -run "^$" -count=1` (ok).
- Inicio de implementacion del bloque de integraciones y pagos (fase de robustecimiento):
	- `backend/handlers/payments_handlers.go` agrega:
		- `MercadoPagoReconcileHandler` para conciliacion manual de pagos pendientes contra API Mercado Pago (`/mercadopago/reconcile`, requiere sesion admin).
		- `WompiWebhookHandler` para notificaciones servidor-servidor (`/wompi/webhook`) con validacion de firma y activacion automatica de licencia cuando aplica.
		- helpers compartidos para token MP, parseo de `external_reference`, estatus aprobados y activacion idempotente de licencia.
	- `backend/db/db.go` agrega helpers de persistencia para conciliacion:
		- listado de pendientes MP (`ListMPPaymentsForReconciliation`),
		- actualizacion por `id` y `payment_id`,
		- actualizacion Wompi por `reference`,
		- resolucion de contexto licencia/empresa para Wompi.
	- `backend/main.go` registra rutas nuevas:
		- `/mercadopago/reconcile`
		- `/wompi/webhook`
	- Validacion tecnica ejecutada:
		- `go test ./auth ./db ./handlers ./metrics ./utils` (ok).
- Se divide la interfaz de ERP extendido en submodulos por dominio, manteniendo el mismo backend.
	- `web/administrar_empresa/modulos_erp_extendido.html` pasa a ser hub de dominios (ventas, finanzas, inventario/compras/rrhh, crm, produccion, logistica, documental/integraciones/dian).
	- `web/administrar_empresa/modulos_erp_dominio.html` (nuevo) concentra la operacion CRUD del dominio seleccionado sin cambiar endpoints backend.
	- `web/js/modulos_erp_extendido.js` (nuevo) centraliza la logica operativa reutilizable de submodulos por dominio.
	- `web/estilos.css` agrega estilos de navegacion y tarjetas para `erp-domain-*`.
- Se completa la operacion frontend de los modulos ERP extendidos en panel de empresa.
	- `web/administrar_empresa/modulos_erp_extendido.html` (nuevo) centraliza el uso de todos los endpoints ERP faltantes con:
		- listado con filtros (`q`, `limit`, `offset`, `include_inactive`),
		- detalle por ID,
		- crear/actualizar por payload JSON,
		- activar/desactivar y eliminacion logica por registro,
		- herramientas DIAN (`checklist`, `validar`, `generar_cufe_demo`, `generar_xml_demo`).
	- `web/administrar_empresa.html` agrega acceso lateral `ERP extendido`.
	- `web/js/administrar_empresa.js` integra `linkERPExtendido` en navegacion y permisos (modulo `seguridad`, accion `update`).
	- `web/estilos.css` agrega estilos dedicados del nuevo modulo (`erp-*`) para formularios, salida, tabla y estado visual.
- Se implementa base de modulos ERP faltantes en backend con esquema multiempresa, migracion y rutas nuevas:
	- `backend/db/modulos_faltantes.go` (tablas y CRUD generico por `empresa_id`).
	- `backend/handlers/modulos_faltantes.go` (rutas ERP adicionales y handler DIAN Colombia).
	- `backend/main.go` integra `EnsureEmpresaModulosFaltantesSchema`, migracion `2026-04-06-021-modulos-faltantes-erp` y `RegisterEmpresaModulosFaltantesRoutes`.
- Se agrega soporte DIAN Colombia operativo en endpoint `/api/empresa/facturacion_electronica/dian` con acciones:
	- `checklist` y `validar`.
	- `generar_cufe_demo` y `generar_xml_demo`.
- Se amplía `web/ayuda/ayuda.html` con seccion detallada para configurar facturacion DIAN desde cero.
- Se sincroniza documentacion tecnica y de BD:
	- `documentos/diagramas/estructura_del_codigo.md`.
	- `documentos/estructura_bd.md` y `estructura_bd.md`.
	- `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`.
- Validacion tecnica ejecutada:
	- `go test ./... -run "^$" -count=1` (ok).

## 2026-04-05
- Se continua con todos los bloques y pruebas en una corrida adicional de verificacion.
	- Validaciones ejecutadas:
		- `runTests` global (150 passed, 0 failed).
		- `powershell -File ..\\scripts\\validar_punto_13.ps1` (ok, suite productiva + suite completa backend).
	- Evidencia actualizada:
		- `documentos/punto_13_validacion_integral_resultado.md`.
		- `scripts/logs/punto13-validacion-20260405-182345.log`.

## 2026-04-05
- Se continua con todos los bloques y pruebas en una nueva corrida completa.
	- Validaciones ejecutadas:
		- `runTests` global (150 passed, 0 failed).
		- `powershell -File ..\\scripts\\validar_punto_13.ps1` (ok, suite productiva + suite completa backend).
	- Evidencia actualizada:
		- `documentos/punto_13_validacion_integral_resultado.md`.
		- `scripts/logs/punto13-validacion-20260405-182133.log`.

## 2026-04-05
- Se continua ejecucion de todos los bloques y pruebas con validacion ampliada por modulos criticos.
	- Validaciones ejecutadas:
		- `powershell -File ..\\scripts\\validar_punto_13.ps1` (ok, suite productiva + suite completa backend).
		- `go test ./handlers -run "TestEmpresa(Usuario|Clientes|Inventario|Compras|Facturacion|Finanzas|Auditoria|Permisos|Carritos)" -count=1` (ok).
		- `go test ./db -run "Test(Cliente|Proveedor|Inventario|Finanzas|Facturacion|Reserva|Vehiculo|Nomina|Tarifa|CodigoDescuento|Comision|Propina)" -count=1` (ok).
		- `go test ./... -count=1` (ok).
	- Evidencia actualizada:
		- `documentos/punto_13_validacion_integral_resultado.md`.
		- `scripts/logs/punto13-validacion-20260405-181807.log`.

## 2026-04-05
- Se reejecuta la validacion integral y bloques adicionales de pruebas para cierre tecnico.
	- Validaciones ejecutadas:
		- `powershell -File .\\scripts\\validar_punto_13.ps1` (ok, suite productiva + suite completa backend).
		- `go test ./handlers -run "TestEmpresa(CarritosCompraListIncluyeTarifaPorDiaAutomatica|TarifasPorMinutosHandlerCRUDAndCalculo|TarifasPorDiaHandlerCRUDAndCalculo|ReservasHotelHandlerCRUDAndDisponibilidad|VehiculosRegistroHandlerCRUDFlow|NominaSueldosHandlerFlow|PropinasHandlerConfigAndReporte|ComisionesServicioHandlerConfigAndReporte|ConfiguracionOperativaHandlerConfigAndRole)" -count=1` (ok).
		- `go test ./... -count=1` (ok).
	- Evidencia actualizada:
		- `documentos/punto_13_validacion_integral_resultado.md`.
		- `scripts/logs/punto13-validacion-20260405-181423.log`.

## 2026-04-05
- Se corrige regresion de pruebas de carritos ante el recalculo de totales al pagar estacion.
	- `backend/handlers/auth_users_carritos_test.go` actualiza el sembrado de datos en:
		- `TestEmpresaCarritosCompraAplicaPropinaSegunConfiguracion`.
		- `TestEmpresaCarritosCompraCodigoDescuentoConsumeUso`.
		- `TestEmpresaCarritosCompraRejectsMetodoPagoInvalido`.
	- Las pruebas ahora crean items reales en `carrito_compra_items` en lugar de forzar `subtotal/total` por SQL directo, quedando alineadas con `RefreshCarritoTotalConTarifaPorDia`.
	- Validacion ejecutada:
		- `go test ./handlers -run "TestEmpresaCarritosCompraAplicaPropinaSegunConfiguracion|TestEmpresaCarritosCompraCodigoDescuentoConsumeUso|TestEmpresaCarritosCompraRejectsMetodoPagoInvalido" -count=1` (ok).
		- `powershell -File .\\scripts\\validar_punto_13.ps1` (ok, incluye suite productiva y suite completa backend).

## 2026-04-05
- Se corrige el flujo de login de usuario de empresa para mantener alcance por `empresa_id` en endpoints publicos de autenticacion.
	- `backend/handlers/usuarios_empresa.go` ahora propaga `empresa_id` en enlaces de correo y confirmacion hacia `/login_usuario.html?empresa_id=...`.
	- `ConfirmarCorreoUsuarioHandler` usa el `empresa_id` confirmado (o de query) para construir el enlace de retorno al login de usuario.
	- `web/js/login_usuario.js` toma `empresa_id` desde querystring y lo envia en query + body a `/api/empresa/usuarios/login` y `/api/empresa/usuarios/establecer_password`.
	- Validacion ejecutada:
		- `go test ./handlers -run "EmpresaUsuario(LoginHandlerSuccess|SetPasswordHandlerSuccess|LoginHandlerRejectsWrongEmpresaScopeFromQuery|SetPasswordHandlerRejectsWrongEmpresaScopeFromQuery)" -count=1` (ok).
		- `get_errors` sobre archivos modificados (sin errores).

## 2026-04-05
- Se completa auditoria integral de rutas `/api/empresa` y se cierra cobertura de wrappers por empresa al 100% en el registro de rutas.
	- `backend/handlers/empresa_permisos.go` agrega `WithEmpresaPublicScope` para endpoints publicos que requieren alcance por `empresa_id` sin autenticacion previa de admin.
	- `backend/main.go` envuelve `/api/empresa/usuarios/login` y `/api/empresa/usuarios/establecer_password` con `WithEmpresaPublicScope`.
	- `backend/main.go` envuelve `/api/empresa/facturacion_electronica/paises_disponibles` con `WithEmpresaFacturacionPermissions`.
	- `backend/handlers/chat_con_inteligencia_artificial_router.go` envuelve rutas del modulo IA (`modelos`, `modelo_preferido`, `consultar`, `historial`) con `WithEmpresaSeguridadPermissions`.
	- `web/administrar_empresa/facturacion_electronica.html` envia `empresa_id` al consultar `paises_disponibles` para compatibilidad con el wrapper de facturacion.
	- Validacion ejecutada:
		- `go test ./handlers -run "WithEmpresaVentasPermissionsInjectsEmpresaIDContextForParsers|EmpresaPermisosContextoHandlerRetornaPermisosPorRol|EmpresaUsuario(LoginHandlerRejectsWrongEmpresaScope|SetPasswordHandlerRejectsWrongEmpresaScope)|ModeloPreferidoHandler(Get|Put)RejectsEmpresaFueraDeAlcanceByGoogleAccount|HistorialHandlerRejectsEmpresaFueraDeAlcanceByGoogleAccount" -count=1` (ok).
		- `go test ./... -run "^$"` (ok).

## 2026-04-05
- Se refuerza la integracion multiempresa para que el alcance autorizado por `empresa_id` viaje en el contexto de request y sea reutilizable por handlers.
	- `backend/handlers/empresa_permisos.go` ahora inyecta `empresaID` en `context.Context` dentro de `WithEmpresa*Permissions`.
	- `backend/handlers/productos.go` actualiza `parseEmpresaIDQuery`, `parseInt64Query` y `parseInt64QueryOptional` para priorizar `empresaID` desde contexto cuando existe.
	- `backend/handlers/empresa_permisos_test.go` agrega `TestWithEmpresaVentasPermissionsInjectsEmpresaIDContextForParsers` para validar la propagacion de scope multiempresa sin dependencia estricta de querystring.
	- Validacion ejecutada: `go test ./handlers -run "WithEmpresaVentasPermissionsInjectsEmpresaIDContextForParsers|EmpresaPermisosContextoHandlerRetornaPermisosPorRol|EmpresaUsuario(LoginHandlerRejectsWrongEmpresaScope|SetPasswordHandlerRejectsWrongEmpresaScope)|ModeloPreferidoHandler(Get|Put)RejectsEmpresaFueraDeAlcanceByGoogleAccount|HistorialHandlerRejectsEmpresaFueraDeAlcanceByGoogleAccount" -count=1` (ok).

## 2026-04-05
- Se completa el subbloque de regresion UAT para endpoints sin wrapper de modulo (continuidad Punto 3).
	- `backend/handlers/auth_users_carritos_test.go` agrega cobertura de alcance por `empresa_id` enviado por querystring en:
		- `TestEmpresaUsuarioLoginHandlerRejectsWrongEmpresaScopeFromQuery`.
		- `TestEmpresaUsuarioSetPasswordHandlerRejectsWrongEmpresaScopeFromQuery`.
	- `backend/handlers/chat_con_inteligencia_artificial_controller_test.go` agrega cobertura de aislamiento por cuenta Google autenticada en:
		- `TestModeloPreferidoHandlerGetRejectsEmpresaFueraDeAlcanceByGoogleAccount`.
		- `TestModeloPreferidoHandlerPutRejectsEmpresaFueraDeAlcanceByGoogleAccount`.
		- `TestHistorialHandlerRejectsEmpresaFueraDeAlcanceByGoogleAccount`.
	- Validacion ejecutada: `go test ./handlers -run "EmpresaUsuario(LoginHandlerRejectsWrongEmpresaScope|SetPasswordHandlerRejectsWrongEmpresaScope)|ModelosHandlerRejectsEmpresaFueraDeAlcanceByGoogleAccount|ConsultarHandlerRejectsEmpresaFueraDeAlcance|ModeloPreferidoHandler(Get|Put)RejectsEmpresaFueraDeAlcanceByGoogleAccount|HistorialHandlerRejectsEmpresaFueraDeAlcanceByGoogleAccount" -count=1` (ok).

## 2026-04-05
- Se completa el subbloque de consumo frontend del contexto de permisos (cierre operativo Punto 3).
	- `web/js/administrar_empresa.js` ahora consume `GET /api/empresa/permisos_contexto?empresa_id={id}` para resolver visibilidad real de enlaces por modulo/accion en el menu lateral.
	- El panel empresa ahora admite `id` o `empresa_id` en querystring para resolver el contexto de permisos sin ambiguedad.
	- Se mantiene fallback local por rol cuando el endpoint no esta disponible, evitando bloqueos de navegacion.
	- `web/administrar_empresa.html` agrega indicador visual `menuPermsEvidence` para evidencia UAT del rol/fuente de permisos aplicado en pantalla.
	- Validacion ejecutada: `get_errors` sobre frontend modificado (sin errores).

## 2026-04-05
- Se agrega endpoint de contexto de permisos por empresa para reforzar el cierre del Punto 3 (permisos y seguridad).
	- `backend/handlers/empresa_permisos.go` incorpora `GET /api/empresa/permisos_contexto` con respuesta de permisos efectivos por modulo/accion para el rol autenticado.
	- El endpoint soporta `include_matrix=1` para retornar matriz completa por rol (`super_administrador`, `admin_empresa`, `supervisor_sucursal`, `cajero`, `inventario`, `compras`, `contabilidad`, `auditor`).
	- `backend/main.go` registra la ruta bajo `WithEmpresaSeguridadPermissions` para mantener aislamiento por `empresa_id`.
	- `backend/handlers/empresa_permisos_test.go` agrega `TestEmpresaPermisosContextoHandlerRetornaPermisosPorRol` y `TestEmpresaPermisosContextoHandlerIncluyeMatrizRoles`.
	- Validacion ejecutada: `go test ./handlers -run "PermisosContexto|WithEmpresa.*Permissions" -count=1` (ok).

## 2026-04-05
- Se amplian los reportes contables de flujo de caja con filtros por categoria y metodo de pago.
	- `backend/handlers/reportes.go` incorpora filtros `categoria` y `metodo_pago` en `contable_flujo_caja` para segmentar ingresos/egresos diarios.
	- El resumen del dataset ahora expone `filtro_categoria` y `filtro_metodo_pago` para trazabilidad del reporte exportado.
	- `backend/handlers/reportes_test.go` agrega `TestEmpresaReportesHandlerDatasetContableFlujoCajaFiltros` para validar segmentacion por categoria/metodo.
	- `web/administrar_empresa/reportes.html` agrega campos de filtro contable y los propaga en consultas/exportaciones del endpoint `/api/empresa/reportes`.

## 2026-04-05
- Se extiende el modulo de reportes con dataset contable de flujo de caja diario.
	- `backend/handlers/reportes.go` agrega dataset `contable_flujo_caja` en `/api/empresa/reportes` y consolida ingresos, egresos, neto del dia, saldo acumulado y conteo de movimientos por fecha.
	- El dataset mantiene paridad de exportacion en `pdf`, `xls`, `csv`, `json` y `txt` desde el catalogo central de reportes empresariales.
	- `backend/handlers/reportes_test.go` agrega `TestEmpresaReportesHandlerDatasetContableFlujoCaja` para validar filas diarias y resumen acumulado del periodo.

## 2026-04-05
- Se extiende el modulo de reportes con dataset contable de liquidaciones de nomina y exportacion PDF.
	- `backend/handlers/reportes.go` agrega dataset `contable_nomina_liquidaciones` con filtros por periodo y `empleado_nomina_id`.
	- `backend/handlers/reportes.go` habilita formato `pdf` en la exportacion de datasets de `/api/empresa/reportes`.
	- `web/administrar_empresa/reportes.html` agrega opcion `PDF` en el selector de formato.
	- `web/administrar_empresa/nomina_sueldos.html` incorpora accion `Exportar liquidaciones` usando `/api/empresa/reportes?action=export`.
	- `backend/handlers/reportes_test.go` agrega cobertura de dataset de nomina y validacion de export PDF.

## 2026-04-05
- Se integra operativamente el modulo de nomina de sueldos con asistencia en backend y panel de empresa.
	- `backend/main.go` incorpora `EnsureEmpresaNominaSchema`, migracion `2026-04-05-020-nomina-sueldos` y ruta `/api/empresa/nomina`.
	- `web/administrar_empresa/nomina_sueldos.html` (nuevo) agrega configuracion legal, empleados, festivos, calculo y consulta de liquidaciones.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran `linkNominaSueldos` en menu y permisos.
	- `web/estilos.css` agrega estilos dedicados del modulo.
	- `documentos/estructura_bd.md` y `estructura_bd.md` incluyen tablas y relaciones de nomina.

## 2026-04-05
- Se agrega `ventas_simple.html` como carrito alterno por estación (modo supermercado) con activación/desactivación por estación.
	- `web/administrar_empresa/ventas_simple.html` (nuevo) incorpora flujo rápido para buscar productos, agregarlos al carrito, ajustar cantidades y visualizar total consolidado por estación.
	- Se corrige la visibilidad del campo de referencia de pago para métodos que la requieren (`tarjeta_credito`, `tarjeta_debito`, `transferencia_bancaria`).
	- El cobro se ejecuta con flujo simplificado usando `action=pagar_estacion` y permite iniciar nueva venta con `action=activar_estacion`.
	- `web/administrar_empresa/configuracion_de_estaciones.html` agrega la bandera local `venta_simple_habilitada` por estación.
	- `web/administrar_empresa/estaciones.html` enruta automáticamente cada estación al carrito completo (`carrito_de_compras.html`) o al carrito simple (`ventas_simple.html`) según su configuración.
	- `web/estilos.css` integra estilos responsive para el nuevo módulo y etiqueta visual del modo por estación.

## 2026-04-05
- Se actualiza la configuracion de `agente_go` para reforzar reportes e interoperabilidad contable.
	- `.github/agents/agente_go.agent.md` agrega regla obligatoria para que todos los reportes puedan exportarse, como minimo, en `PDF` y `Excel`, y tambien en formatos de uso comun (`CSV`, `JSON`, `TXT`).
	- Se incorpora regla de compatibilidad con software contable externo mediante formatos estandar de intercambio y trazabilidad por `empresa_id`, documento y periodo.

## 2026-04-05
- Se agrega el dataset `reporte_de_turno` al modulo empresarial de reportes para control operativo de caja por turno.
	- `backend/handlers/reportes.go` incorpora `reporte_de_turno` en `/api/empresa/reportes` con filtros por `usuario`, `caja_codigo`, `turno` y `cierre_id`.
	- El dataset incluye detalle por carrito con `activado_en`, `pagado_en`, metodo de pago y acumulados de ventas por `producto` y `servicio`.
	- El resumen del reporte calcula gastos de turno y efectivo esperado (`efectivo_deberia_haber`) combinando cierres de caja y movimientos financieros.
	- `web/administrar_empresa/reportes.html` agrega campos de filtro de turno/caja/usuario/cierre y envia estos parametros en consultas y exportes del dataset.
	- `backend/handlers/reportes_test.go` agrega `TestEmpresaReportesHandlerDatasetReporteTurno` para validar filtros y consistencia del resumen financiero del turno.

## 2026-04-05
- Se crea el modulo de tarifas por dia por estacion con recálculo automático de deuda en carritos hotel activos.
	- `backend/db/tarifas_por_dia.go` (nuevo) agrega esquema `empresa_tarifas_por_dia`, CRUD, horarios `hora_check_in`/`hora_check_out` y calculo de dias/monto.
	- `backend/db/carritos_tarifa_dia.go` (nuevo) integra calculo automático de deuda diaria en carritos de estación y refresco masivo para listados.
	- `backend/db/carritos_compras.go` ajusta `RecalculateCarritoCompraTotals` para incluir tarifa diaria cuando aplique.
	- `backend/handlers/tarifas_por_dia.go` (nuevo) expone `/api/empresa/tarifas_por_dia` con acciones `listar`, `detalle`, `aplicable`, `calcular`, `activar` y `desactivar`.
	- `backend/handlers/carritos_compras.go` recalcula tarifa diaria al listar carritos y antes de `action=pagar_estacion`.
	- `backend/main.go` integra `EnsureEmpresaTarifasPorDiaSchema`, migracion `2026-04-05-019-tarifas-por-dia` y ruta protegida del modulo.
	- `web/administrar_empresa/tarifas_por_dia.html` (nuevo) agrega UI de configuracion, filtros y simulador por rango de fechas.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran `linkTarifasPorDia` en menu lateral y permisos.
	- Cobertura agregada en `backend/db/tarifas_por_dia_test.go`, `backend/handlers/tarifas_por_dia_test.go` y `backend/handlers/carritos_tarifa_por_dia_test.go`.

## 2026-04-05
- Se crea el modulo de tarifas por minutos por estacion con reglas por dia de semana y calculo de bloques extra.
	- `backend/db/tarifas_por_minutos.go` (nuevo) agrega esquema `empresa_tarifas_por_minutos`, CRUD, resolucion por dia (`dia_semana_desde/hasta`) y calculo de monto por minutos consumidos.
	- `backend/handlers/tarifas_por_minutos.go` (nuevo) expone `/api/empresa/tarifas_por_minutos` con acciones `listar`, `detalle`, `aplicable`, `calcular`, `activar` y `desactivar`.
	- `backend/main.go` integra `EnsureEmpresaTarifasPorMinutosSchema`, migracion `2026-04-05-018-tarifas-por-minutos` y ruta protegida del modulo.
	- `web/administrar_empresa/tarifas_por_minutos.html` (nuevo) agrega formulario de tarifas, filtros y simulador de cobro por minutos.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran `linkTarifasPorMinutos` en menu lateral y permisos por rol.
	- Cobertura agregada en `backend/db/tarifas_por_minutos_test.go` y `backend/handlers/tarifas_por_minutos_test.go`.
	- Se actualiza documentacion y diagramas: `documentos/estructura_bd.md`, `estructura_bd.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md` y `documentos/diagramas/diagrama_flujo_procesos.md`.

## 2026-04-05
- Se ubica el documento de base de datos dentro de la carpeta `documentos` y se alinea `agente_go` con esa ruta.
	- `documentos/estructura_bd.md` se incorpora como ubicacion requerida de la estructura de base de datos.
	- `.github/agents/agente_go.agent.md` ahora exige revisar `documentos/estructura_bd.md` antes de cambios en tablas, consultas, migraciones o datos.
	- `estructura_bd.md` en raiz se mantiene sincronizado como copia de compatibilidad documental.

## 2026-04-05
- Se actualiza la configuracion de `agente_go` para forzar lectura previa de documentacion de base de datos en tareas de datos.
	- `.github/agents/agente_go.agent.md` agrega regla para revisar `estructura_bd.md` antes de cambios en tablas, consultas, migraciones o datos operativos.

## 2026-04-05
- Se agrega modulo de busqueda y gestion de facturas electronicas por empresa.
	- `web/administrar_empresa/facturas_electronicas.html` (nuevo) permite buscar por cliente, documento y rango de fechas; ver detalle; reenviar por correo; e imprimir.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran el acceso lateral `Facturas electrónicas` con permisos de lectura del modulo `facturacion`.
	- `backend/handlers/facturacion_electronica.go` incorpora:
		- `GET action=documentos` para consulta de documentos de facturacion por filtros (`cliente`, `documento`, `fecha_desde`, `fecha_hasta`, `estado_documento`, `tipo_documento`, `q`).
		- `PUT/POST action=reenviar_correo` para reintento manual de envio al correo del cliente.
	- `backend/db/documentos_transaccionales.go` agrega consulta enriquecida con cliente (`nombre`, `email`, `documento`) para listado filtrado.
	- Cobertura agregada en `backend/db/documentos_transaccionales_test.go` para filtros de cliente/fecha/documento.

## 2026-04-05
- Se normaliza la documentacion de base de datos para eliminar duplicidad entre documentos.
	- `estructura_bd.md` queda como fuente canonica del esquema fisico SQLite.
	- `documentos/descripcion_de_las_bases_De_datos` se redefine como guia complementaria funcional y reglas operativas de mantenimiento.
	- Se evita repetir listados tabla-por-tabla en dos archivos distintos.

## 2026-04-05
- Se consolida Configuración avanzada dentro de Facturación electrónica en el panel de empresa.
	- `web/administrar_empresa/facturacion_electronica.html` integra el formulario completo de configuración avanzada fiscal/impresión y su persistencia mediante `/api/empresa/configuracion_avanzada`.
	- `web/administrar_empresa.html` elimina el enlace lateral independiente `Configuración avanzada` para dejar una única entrada funcional en `Facturación electrónica`.
	- `web/js/administrar_empresa.js` retira `linkConfigAvanzada` del catálogo de enlaces/permisos del menú.
	- `web/ayuda/ayuda.html` actualiza el tutorial para indicar que la configuración avanzada ahora se gestiona desde `facturacion_electronica.html`.
	- `web/administrar_empresa/configuracion_avanzada.html` se elimina del repositorio por consolidación funcional.

## 2026-04-05
- Se agrega configuracion operativa de cobro por empresa y por rol de usuario.
	- `backend/db/configuracion_operativa.go` (nuevo) agrega tablas `empresa_configuracion_operativa` y `empresa_configuracion_operativa_roles`, con resolucion efectiva de permisos por rol.
	- `backend/handlers/configuracion_operativa.go` (nuevo) expone `/api/empresa/configuracion_operativa` para consultar y actualizar reglas base y overrides por rol (`action=rol`).
	- `backend/handlers/empresa_permisos.go` y `backend/handlers/productos.go` propagan/normalizan rol admin en request para enforcement transversal.
	- `backend/handlers/carritos_compras.go` aplica enforcement en `action=pagar_estacion`: bloquea metodos de pago no permitidos y desactiva propina/comision segun politica operativa efectiva por rol.
	- `backend/main.go` registra `EnsureEmpresaConfiguracionOperativaSchema`, migracion `2026-04-05-017-configuracion-operativa-cobro` y ruta protegida `/api/empresa/configuracion_operativa`.
	- `web/administrar_empresa/configuracion.html` incorpora tarjeta de checks para metodos de pago, propinas y comisiones por empresa y por rol.
	- `web/administrar_empresa/carrito_de_compras.html` consume la politica operativa efectiva y refleja en UI los metodos permitidos, con bloqueo visual y validacion previa al pago.
	- Cobertura agregada en `backend/db/configuracion_operativa_test.go`, `backend/handlers/configuracion_operativa_test.go` y `backend/handlers/auth_users_carritos_test.go`.
	- Validacion ejecutada: pruebas dirigidas en DB/handlers/carritos (ok) y verificacion sin errores en frontend actualizado.

## 2026-04-05
- Se crea el modulo de comisiones por servicio por empresa con reporte por lavador.
	- `backend/db/comisiones_servicio.go` (nuevo) agrega tablas de configuracion y movimientos (`empresa_comisiones_servicio_configuracion`, `empresa_comisiones_servicio_movimientos`) con calculo/reporte por lavador.
	- `backend/handlers/comisiones.go` (nuevo) expone `/api/empresa/comisiones` con acciones `config`, `reporte` y `movimientos`.
	- `backend/handlers/carritos_compras.go` integra `usuario_lavador` en `action=pagar_estacion` y registra comisiones automaticas de servicios de lavado al cerrar venta.
	- `backend/main.go` asegura esquema de comisiones, registra migracion `2026-04-05-016-comisiones-servicio` y publica ruta protegida de comisiones bajo permisos de finanzas.
	- `web/administrar_empresa/comisiones.html` (nuevo) incorpora configuracion y reporte de comisiones por lavador.
	- `web/administrar_empresa/carrito_de_compras.html` agrega captura de lavador para comision, carga de configuracion de comisiones y visualizacion de comision estimada/registrada en cobro.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran acceso lateral `Comisiones` (`linkComisiones`) con permisos del modulo finanzas.
	- Cobertura agregada en `backend/db/comisiones_servicio_test.go`, `backend/handlers/comisiones_test.go` y `backend/handlers/auth_users_carritos_test.go`.

## 2026-04-05
- Se actualiza la configuracion de `agente_go` para definir semantica y concurrencia de estaciones.
	- `.github/agents/agente_go.agent.md` agrega que una estacion puede representar mesa de restaurante, habitacion de hotel, habitacion de motel, punto de caja u otro punto operativo equivalente.
	- Se establece que estaciones deben soportar multiples carritos/sesiones y multiples clientes en simultaneo, con aislamiento por `empresa_id` y trazabilidad operativa.

## 2026-04-05
- Se completa el modulo de reservas por estacion/habitacion para operacion empresarial.
	- `backend/db/reservas_hotel.go` (nuevo) implementa esquema y logica de reservas con disponibilidad por rango, conflicto de solapamiento, expiracion de pendientes, confirmacion de pago, cancelacion, activacion/desactivacion y eliminacion.
	- `backend/handlers/reservas_hotel.go` (nuevo) expone `/api/empresa/reservas_hotel` con acciones `listar`, `detalle`, `disponibilidad`, `confirmar_pago`, `cancelar`, `activar`, `desactivar` y CRUD operativo.
	- `backend/main.go` asegura esquema de reservas, registra migracion `2026-04-05-015-reservas-hotel` y publica ruta protegida bajo permisos de ventas.
	- `web/administrar_empresa/reservas_hotel.html` (nuevo) agrega interfaz para crear/editar reservas, consultar disponibilidad y ejecutar acciones de ciclo de vida.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran el acceso lateral `Reservas` (`linkReservasHotel`) con control de permisos por rol.
	- Cobertura de pruebas en:
		- `backend/db/reservas_hotel_test.go` (flujo DB end-to-end con disponibilidad y estados).
		- `backend/handlers/reservas_hotel_test.go` (nuevo, flujo HTTP completo del endpoint).
	- Validaciones ejecutadas:
		- `go test ./db -run ReservaHotel -count=1` (ok).
		- `go test ./handlers -run ReservasHotel -count=1` (ok).

## 2026-04-05
- Se crea el modulo de registro de vehiculos por empresa para controlar ingreso y salida por patente.
	- `backend/db/vehiculos_registro.go` (nuevo) agrega esquema y operaciones CRUD del registro vehicular, con estado operativo (`en_empresa`/`retirado`) y marcacion de salida.
	- `backend/handlers/vehiculos_registro.go` (nuevo) expone `/api/empresa/vehiculos_registro` con acciones de consulta, alta, edicion, activar/desactivar, marcar salida y eliminacion.
	- `backend/main.go` asegura esquema del modulo, registra migracion `2026-04-05-014-vehiculos-registro` y publica ruta protegida bajo permisos de seguridad.
	- `web/administrar_empresa/vehiculos_registro.html` (nuevo) incorpora UI de registro de vehiculos con patente, conductor, propietario, fechas de ingreso/salida y filtros operativos.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran el acceso lateral `Registro de vehiculos` con permisos por rol.
	- Cobertura agregada en `backend/db/vehiculos_registro_test.go` y `backend/handlers/vehiculos_registro_test.go`.
	- Validaciones ejecutadas:
		- `go test ./db -run Vehiculo -count=1` (ok).
		- `go test ./handlers -run VehiculosRegistro -count=1` (ok).

## 2026-04-05
- Se crea el modulo de propinas por empresa con configuracion operativa y reporte por usuario o universal.
	- `backend/db/propinas.go` (nuevo) agrega tablas de configuracion y movimientos de propinas, con soporte de reporte acumulado por usuario y reparto universal entre usuarios activos.
	- `backend/handlers/propinas.go` (nuevo) expone `/api/empresa/propinas` con acciones de configuracion, reporte y consulta de movimientos.
	- `backend/handlers/carritos_compras.go` integra propina en `action=pagar_estacion`, valida `total_pagado` contra total final con propina y registra movimiento de propina al cerrar venta.
	- `backend/main.go` asegura esquema de propinas, registra migracion `2026-04-05-013-propinas` y publica ruta protegida `/api/empresa/propinas` bajo permisos de finanzas.
	- `web/administrar_empresa/propinas.html` (nuevo) incorpora modulo de configuracion de propinas y reporte por rango, usuario y modo.
	- `web/administrar_empresa/carrito_de_compras.html` agrega control de aplicar propina en cobro de estacion, carga de configuracion y desglose de total final con propina.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran acceso de menu `Propinas` con permisos del modulo finanzas.
	- Cobertura agregada en `backend/db/propinas_test.go`, `backend/handlers/propinas_test.go` y `backend/handlers/auth_users_carritos_test.go`.
	- Validaciones ejecutadas:
		- `go test ./db -run Propina -count=1` (ok).
		- `go test ./handlers -run "Propinas|CarritosCompraAplicaPropinaSegunConfiguracion" -count=1` (ok).
		- `go test ./db ./handlers -count=1` (ok).

## 2026-04-05
- Se agrega `transferencia_bancaria` como forma de pago transversal en flujo de carritos y finanzas.
	- `backend/db/carritos_compras.go` normaliza y acepta alias de transferencia bancaria (`transferencia`, `transferencia_bancaria`).
	- `backend/handlers/carritos_compras.go` habilita transferencia bancaria en pago directo y mixto, y exige referencia minima para tarjeta/transferencia.
	- `backend/handlers/auth_users_carritos_test.go` agrega cobertura de pago exitoso por transferencia bancaria y rechazo cuando falta referencia valida.
	- `web/administrar_empresa/carrito_de_compras.html` incorpora transferencia bancaria en selectores de pago, habilita validacion de pago mixto con transferencia y envía `pagos_mixtos` al backend.
	- `web/administrar_empresa/finanzas.html` estandariza opcion de `transferencia_bancaria` y mantiene compatibilidad con registros legacy `transferencia`.
	- `web/ayuda/ayuda.html` actualiza descripcion de metodos soportados en cierre de carrito.
	- `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md` y `documentos/diagramas/diagrama_flujo_procesos.md` reflejan el nuevo flujo de pago.

## 2026-04-05
- Robustecimiento del modulo de auditoria empresarial con foco en trazabilidad, seguridad operativa y analisis forense.
	- `backend/db/auditoria_empresa.go` amplía filtros (`metodo_http`, `recurso`, `endpoint`, `search`), agrega `offset`, agrega conteo filtrado (`CountEmpresaAuditoriaEventos`) y refuerza indices de rendimiento.
	- `backend/handlers/auditoria_empresa.go` valida fechas/parametros, publica metadata de paginacion por headers y soporta consulta avanzada de eventos.
	- `backend/handlers/empresa_permisos.go` registra intentos criticos denegados (401/403/500) como eventos de auditoria no bloqueantes.
	- `backend/utils/utils.go` expone `RequestIDFromContext` para correlacion real entre logs de request y eventos de auditoria.
	- `web/administrar_empresa/auditoria.html` agrega filtros avanzados, paginador y panel de detalle JSON por evento.
	- `web/estilos.css` agrega estilos centralizados para paginacion y detalle del modulo de auditoria.
	- Se amplian pruebas en `backend/db/auditoria_empresa_test.go` y `backend/handlers/auditoria_empresa_test.go`.
	- Validaciones ejecutadas:
		- `go test ./db -run Auditoria -count=1` (ok).
		- `go test ./handlers -run Auditoria -count=1` (ok).
		- `go test ./auth ./db ./handlers ./metrics ./utils -count=1` (ok).

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
