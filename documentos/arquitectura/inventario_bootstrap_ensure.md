# Inventario de bootstrap Ensure

Estado: generado. Ultima actualizacion: 2026-07-16.

Este archivo se genera con `node tools/ensure_bootstrap_inventory.mjs`. Inventaria las funciones `Ensure*` de backend y es la base obligatoria para retirar el bootstrap historico. Una clasificacion `por confirmar` no autoriza desactivar `PCS_RUNTIME_SCHEMA_BOOTSTRAP`; debe convertirse en una migracion catalogada, seed programado o verificacion sin DDL.

## Resumen

- Funciones inventariadas: 142.
- Huella del catalogo legado: `b327dcba705f35df6d432082d7c6474e1b3d443536fda266e9f06c693650582e` (117 pasos).
- DDL / indice / funcion: 113.
- DDL catalogado de plataforma: 4.
- compatibilidad PostgreSQL: 2.
- provisionamiento de integracion: 4.
- regla auxiliar o verificacion: 7.
- seed o provisionamiento idempotente: 12.
- Fuente: `backend/db`, `backend/handlers` y `backend/main.go`; excluye pruebas.

## Registro

| Funcion | Archivo | Clase inferida | Base objetivo inferida |
| --- | --- | --- | --- |
| `EnsureAdminPrincipalDelegacionesSchema` | [backend/db/admin_principal_delegaciones.go:31](../../backend/db/admin_principal_delegaciones.go#L31) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaAIOpenAIProviderSchema` | [backend/db/ai_empresa_proveedor.go:23](../../backend/db/ai_empresa_proveedor.go#L23) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaAIEnterpriseSchema` | [backend/db/ai_enterprise.go:98](../../backend/db/ai_enterprise.go#L98) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaAIUConstruccionSchema` | [backend/db/aiu_construccion.go:139](../../backend/db/aiu_construccion.go#L139) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaAlquileresSchema` | [backend/db/alquileres.go:210](../../backend/db/alquileres.go#L210) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureAlquileresTipoEmpresaYLicencias` | [backend/db/alquileres_bootstrap.go:64](../../backend/db/alquileres_bootstrap.go#L64) | seed o provisionamiento idempotente | superadministrador o por confirmar |
| `EnsureAsesorComercialSchema` | [backend/db/asesor_comercial.go:81](../../backend/db/asesor_comercial.go#L81) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaAsistenciaSchema` | [backend/db/asistencia_empleados.go:76](../../backend/db/asistencia_empleados.go#L76) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureAsyncJobsSchema` | [backend/db/async_jobs.go:81](../../backend/db/async_jobs.go#L81) | DDL catalogado de plataforma | superadministrador |
| `EnsureEmpresaAuditoriaSchema` | [backend/db/auditoria_empresa.go:69](../../backend/db/auditoria_empresa.go#L69) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureSuperAuditoriaSchema` | [backend/db/auditoria_super.go:62](../../backend/db/auditoria_super.go#L62) | DDL / indice / funcion | superadministrador o por confirmar |
| `EnsureEmpresaBackupsSchema` | [backend/db/backups_empresariales.go:560](../../backend/db/backups_empresariales.go#L560) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaCalculadoraSchema` | [backend/db/calculadora_operativa.go:118](../../backend/db/calculadora_operativa.go#L118) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaCamarasSchema` | [backend/db/camaras.go:47](../../backend/db/camaras.go#L47) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaCarnetsSchema` | [backend/db/carnets_empresa.go:99](../../backend/db/carnets_empresa.go#L99) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaCarnetDefaultTemplate` | [backend/db/carnets_empresa.go:271](../../backend/db/carnets_empresa.go#L271) | seed o provisionamiento idempotente | empresas o por confirmar |
| `EnsureEmpresaCarritosSchema` | [backend/db/carritos_compras.go:316](../../backend/db/carritos_compras.go#L316) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaCentrosCostoSchema` | [backend/db/centros_costo.go:135](../../backend/db/centros_costo.go#L135) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaAIChatSchema` | [backend/db/chat_inteligencia_artificial.go:336](../../backend/db/chat_inteligencia_artificial.go#L336) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureSuperAIChatSchema` | [backend/db/chat_inteligencia_artificial.go:627](../../backend/db/chat_inteligencia_artificial.go#L627) | DDL / indice / funcion | superadministrador o por confirmar |
| `EnsureEmpresaChatTareasSchema` | [backend/db/chat_tareas.go:149](../../backend/db/chat_tareas.go#L149) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaCierreFiscalSchema` | [backend/db/cierre_fiscal.go:115](../../backend/db/cierre_fiscal.go#L115) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaClientesSchema` | [backend/db/clientes.go:253](../../backend/db/clientes.go#L253) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaCobranzaSchema` | [backend/db/cobranza.go:151](../../backend/db/cobranza.go#L151) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaCodigosDescuentoSchema` | [backend/db/codigos_descuento.go:97](../../backend/db/codigos_descuento.go#L97) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaComisionesServicioSchema` | [backend/db/comisiones_servicio.go:188](../../backend/db/comisiones_servicio.go#L188) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaComprasAvanzadasSchema` | [backend/db/compras_avanzadas.go:123](../../backend/db/compras_avanzadas.go#L123) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresasComprasSchema` | [backend/db/compras_y_proveedores.go:87](../../backend/db/compras_y_proveedores.go#L87) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaConfiguracionOperativaSchema` | [backend/db/configuracion_operativa.go:232](../../backend/db/configuracion_operativa.go#L232) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureConstructoraTipoEmpresaYLicencias` | [backend/db/constructora_bootstrap.go:70](../../backend/db/constructora_bootstrap.go#L70) | seed o provisionamiento idempotente | superadministrador o por confirmar |
| `EnsureEmpresaContabilidadColombiaSchema` | [backend/db/contabilidad_colombia.go:164](../../backend/db/contabilidad_colombia.go#L164) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaContabilidadColombiaAvanzadaSchema` | [backend/db/contabilidad_colombia_avanzada.go:364](../../backend/db/contabilidad_colombia_avanzada.go#L364) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureSuperContractSchema` | [backend/db/contrato_super.go:133](../../backend/db/contrato_super.go#L133) | DDL / indice / funcion | superadministrador o por confirmar |
| `EnsureDefaultSuperContract` | [backend/db/contrato_super.go:231](../../backend/db/contrato_super.go#L231) | seed o provisionamiento idempotente | superadministrador o por confirmar |
| `EnsureEmpresaControlElectricoSchema` | [backend/db/control_electrico.go:241](../../backend/db/control_electrico.go#L241) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureSuperCorreoNotificacionesPruebaSchema` | [backend/db/correo_notificaciones_prueba.go:44](../../backend/db/correo_notificaciones_prueba.go#L44) | DDL / indice / funcion | superadministrador o por confirmar |
| `EnsureEmpresaCorteCajaConfiguracionSchema` | [backend/db/corte_caja_configuracion.go:102](../../backend/db/corte_caja_configuracion.go#L102) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaCreditosSchema` | [backend/db/creditos.go:622](../../backend/db/creditos.go#L622) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaCRMVentasAvanzadasSchema` | [backend/db/crm_ventas_avanzadas.go:122](../../backend/db/crm_ventas_avanzadas.go#L122) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaDatafonosSchema` | [backend/db/datafonos.go:110](../../backend/db/datafonos.go#L110) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureAdministradoresAuthSchema` | [backend/db/db.go:146](../../backend/db/db.go#L146) | DDL / indice / funcion | superadministrador o por confirmar |
| `EnsurePaymentGatewaySchema` | [backend/db/db.go:197](../../backend/db/db.go#L197) | DDL / indice / funcion | superadministrador o por confirmar |
| `EnsureLicenciasSchema` | [backend/db/db.go:273](../../backend/db/db.go#L273) | DDL / indice / funcion | superadministrador o por confirmar |
| `EnsureUserEmpresa` | [backend/db/db.go:414](../../backend/db/db.go#L414) | regla auxiliar o verificacion | empresas o por confirmar |
| `EnsureEmpresaDeclaracionesTributariasSchema` | [backend/db/declaraciones_tributarias.go:104](../../backend/db/declaraciones_tributarias.go#L104) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaDocumentosTransaccionalesSchema` | [backend/db/documentos_transaccionales.go:120](../../backend/db/documentos_transaccionales.go#L120) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaDomiciliosSchema` | [backend/db/domicilios.go:209](../../backend/db/domicilios.go#L209) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaEmailCorporativoSchema` | [backend/db/email_corporativo.go:32](../../backend/db/email_corporativo.go#L32) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureAdminEmpresaCompartidaSchema` | [backend/db/empresa_admin_compartida.go:97](../../backend/db/empresa_admin_compartida.go#L97) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaAgentesUsoSchema` | [backend/db/empresa_agentes_uso.go:19](../../backend/db/empresa_agentes_uso.go#L19) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaBuzonSchema` | [backend/db/empresa_buzon.go:97](../../backend/db/empresa_buzon.go#L97) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureCatalogoLegalPaisSchema` | [backend/db/empresa_colombia_defaults.go:103](../../backend/db/empresa_colombia_defaults.go#L103) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaConfiguracionAvanzadaSchema` | [backend/db/empresa_configuracion_avanzada.go:124](../../backend/db/empresa_configuracion_avanzada.go#L124) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaConfiguracionGeneralSchema` | [backend/db/empresa_configuracion_general.go:52](../../backend/db/empresa_configuracion_general.go#L52) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaEstacionAseoSchema` | [backend/db/empresa_estacion_aseo.go:55](../../backend/db/empresa_estacion_aseo.go#L55) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaEstacionPrefsSchema` | [backend/db/empresa_estacion_prefs.go:44](../../backend/db/empresa_estacion_prefs.go#L44) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaImpresorasSchema` | [backend/db/empresa_impresoras.go:171](../../backend/db/empresa_impresoras.go#L171) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaPOS80Defaults` | [backend/db/empresa_impresoras.go:905](../../backend/db/empresa_impresoras.go#L905) | seed o provisionamiento idempotente | empresas o por confirmar |
| `EnsureAllEmpresasPOS80Defaults` | [backend/db/empresa_impresoras.go:980](../../backend/db/empresa_impresoras.go#L980) | seed o provisionamiento idempotente | empresas o por confirmar |
| `EnsureEmpresaPermisosFinosSchema` | [backend/db/empresa_permisos_finos.go:24](../../backend/db/empresa_permisos_finos.go#L24) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresasScopeReferences` | [backend/db/empresa_scope.go:10](../../backend/db/empresa_scope.go#L10) | regla auxiliar o verificacion | empresas o por confirmar |
| `EnsureEmpresaEnergiaSolarSchema` | [backend/db/energia_solar.go:108](../../backend/db/energia_solar.go#L108) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaEstacionColumnPreferencesSchema` | [backend/db/estacion_columnas_pref.go:24](../../backend/db/estacion_columnas_pref.go#L24) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEstacionVIPCodigosSchema` | [backend/db/estacion_vip_codigos.go:26](../../backend/db/estacion_vip_codigos.go#L26) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaEventosContablesSchema` | [backend/db/eventos_contables.go:263](../../backend/db/eventos_contables.go#L263) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaFacturacionElectronicaSchema` | [backend/db/facturacion_electronica.go:392](../../backend/db/facturacion_electronica.go#L392) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaFinanzasSchema` | [backend/db/finanzas.go:187](../../backend/db/finanzas.go#L187) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaHojaVidaOperativaSchema` | [backend/db/hoja_vida_operativa.go:99](../../backend/db/hoja_vida_operativa.go#L99) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureHorariosTrabajadoresSchema` | [backend/db/horarios_trabajadores.go:140](../../backend/db/horarios_trabajadores.go#L140) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureHotelTarjetasAccesoSchema` | [backend/db/hotel_tarjetas.go:57](../../backend/db/hotel_tarjetas.go#L57) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaImportacionesCosteoSchema` | [backend/db/importaciones_costeo.go:92](../../backend/db/importaciones_costeo.go#L92) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaImpuestosSchema` | [backend/db/impuestos.go:103](../../backend/db/impuestos.go#L103) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaInventarioAvanzadoSchema` | [backend/db/inventario_avanzado.go:106](../../backend/db/inventario_avanzado.go#L106) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaLicenciasAdicionalesSchema` | [backend/db/licencias_adicionales.go:58](../../backend/db/licencias_adicionales.go#L58) | DDL / indice / funcion | superadministrador o por confirmar |
| `EnsurePowerfulSystemEmpresa` | [backend/db/licencias_empresa_sistema.go:47](../../backend/db/licencias_empresa_sistema.go#L47) | seed o provisionamiento idempotente | superadministrador o por confirmar |
| `EnsurePowerfulSystemEmpresaDefaultLogo` | [backend/db/licencias_empresa_sistema.go:104](../../backend/db/licencias_empresa_sistema.go#L104) | seed o provisionamiento idempotente | superadministrador o por confirmar |
| `EnsureLicenciasCatalogoGlobal` | [backend/db/licencias_globales.go:158](../../backend/db/licencias_globales.go#L158) | seed o provisionamiento idempotente | superadministrador o por confirmar |
| `EnsureLicenciasGratisActivacionesSchema` | [backend/db/licencias_gratis.go:11](../../backend/db/licencias_gratis.go#L11) | DDL / indice / funcion | superadministrador o por confirmar |
| `EnsureLicenciaEmpresaRetencionSchema` | [backend/db/licencias_retencion_empresas.go:55](../../backend/db/licencias_retencion_empresas.go#L55) | DDL / indice / funcion | superadministrador o por confirmar |
| `EnsureLicenciaVencimientoNotificacionesSchema` | [backend/db/licencias_vencimiento_notificaciones.go:44](../../backend/db/licencias_vencimiento_notificaciones.go#L44) | DDL / indice / funcion | superadministrador o por confirmar |
| `EnsureEmpresaWMSSchema` | [backend/db/logistica_wms.go:129](../../backend/db/logistica_wms.go#L129) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureSchemaMigrationsTable` | [backend/db/migrations.go:40](../../backend/db/migrations.go#L40) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureMobileAPIIdempotencySchema` | [backend/db/mobile_api_idempotency.go:39](../../backend/db/mobile_api_idempotency.go#L39) | DDL catalogado de plataforma | empresas |
| `EnsureEmpresaModulosColombiaSchema` | [backend/db/modulos_empresariales_colombia.go:468](../../backend/db/modulos_empresariales_colombia.go#L468) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaModulosFaltantesSchema` | [backend/db/modulos_faltantes.go:106](../../backend/db/modulos_faltantes.go#L106) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaNextcloudSchema` | [backend/db/nextcloud.go:14](../../backend/db/nextcloud.go#L14) | DDL catalogado de plataforma | empresas |
| `EnsureEmpresaNominaColombiaAvanzadaSchema` | [backend/db/nomina_colombia_avanzada.go:100](../../backend/db/nomina_colombia_avanzada.go#L100) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaNominaSchema` | [backend/db/nomina_sueldos.go:388](../../backend/db/nomina_sueldos.go#L388) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaVentasOfflineSchema` | [backend/db/offline_ventas.go:59](../../backend/db/offline_ventas.go#L59) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureOutboxSchema` | [backend/db/outbox.go:42](../../backend/db/outbox.go#L42) | DDL catalogado de plataforma | superadministrador |
| `EnsureEmpresaParqueaderoSchema` | [backend/db/parqueadero.go:89](../../backend/db/parqueadero.go#L89) | DDL / indice / funcion | empresas o por confirmar |
| `EnsurePaymentCheckoutIdempotencySchema` | [backend/db/payment_checkout_idempotency.go:61](../../backend/db/payment_checkout_idempotency.go#L61) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureNuevasPlantillasTipoEmpresaYLicencias` | [backend/db/plantillas_nuevas_bootstrap.go:250](../../backend/db/plantillas_nuevas_bootstrap.go#L250) | seed o provisionamiento idempotente | superadministrador o por confirmar |
| `EnsureNuevasPlantillasProduccionMasivaLicencias` | [backend/db/plantillas_nuevas_bootstrap.go:280](../../backend/db/plantillas_nuevas_bootstrap.go#L280) | regla auxiliar o verificacion | superadministrador o por confirmar |
| `EnsureEmpresaPortalContadorSchema` | [backend/db/portal_contador.go:109](../../backend/db/portal_contador.go#L109) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaPortalTercerosCertificadosSchema` | [backend/db/portal_terceros_certificados.go:98](../../backend/db/portal_terceros_certificados.go#L98) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaProduccionMRPSchema` | [backend/db/produccion_mrp.go:159](../../backend/db/produccion_mrp.go#L159) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaProductosSchema` | [backend/db/productos.go:515](../../backend/db/productos.go#L515) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaBodega1` | [backend/db/productos.go:1224](../../backend/db/productos.go#L1224) | regla auxiliar o verificacion | empresas o por confirmar |
| `EnsureEmpresaPropinasSchema` | [backend/db/propinas.go:152](../../backend/db/propinas.go#L152) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaRappiSchema` | [backend/db/rappi.go:52](../../backend/db/rappi.go#L52) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaPublicacionesRedSocialSchema` | [backend/db/red_social.go:28](../../backend/db/red_social.go#L28) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaRedSocialInteraccionesSchema` | [backend/db/red_social.go:92](../../backend/db/red_social.go#L92) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaReportesProgramacionSchema` | [backend/db/reportes_programacion.go:41](../../backend/db/reportes_programacion.go#L41) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaReservasHotelSchema` | [backend/db/reservas_hotel.go:89](../../backend/db/reservas_hotel.go#L89) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureRolesPermisosSchema` | [backend/db/roles_permisos_usuario.go:25](../../backend/db/roles_permisos_usuario.go#L25) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureRolesDeUsuarioSchema` | [backend/db/roles_tipos_usuario.go:29](../../backend/db/roles_tipos_usuario.go#L29) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureRuntimeDatabaseRole` | [backend/db/runtime_db_role.go:17](../../backend/db/runtime_db_role.go#L17) | regla auxiliar o verificacion | empresas o por confirmar |
| `EnsureBackupDatabaseRole` | [backend/db/runtime_db_role.go:95](../../backend/db/runtime_db_role.go#L95) | regla auxiliar o verificacion | empresas o por confirmar |
| `EnsureEmpresaSensorPuertasSchema` | [backend/db/sensor_puertas.go:77](../../backend/db/sensor_puertas.go#L77) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaSoporteRemotoSchema` | [backend/db/soporte_remoto.go:401](../../backend/db/soporte_remoto.go#L401) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaSoportesComprasIASchema` | [backend/db/soportes_compras_ia.go:131](../../backend/db/soportes_compras_ia.go#L131) | DDL / indice / funcion | empresas o por confirmar |
| `EnsurePostgresRuntimeCompat` | [backend/db/sql_compat.go:31](../../backend/db/sql_compat.go#L31) | compatibilidad PostgreSQL | empresas o por confirmar |
| `EnsurePostgresPrimaryKeySequences` | [backend/db/sql_compat.go:162](../../backend/db/sql_compat.go#L162) | compatibilidad PostgreSQL | empresas o por confirmar |
| `EnsureSuperAlertasSchema` | [backend/db/super_alertas.go:82](../../backend/db/super_alertas.go#L82) | DDL / indice / funcion | superadministrador o por confirmar |
| `EnsureSuperCorreosMasivosSchema` | [backend/db/super_correos_masivos.go:56](../../backend/db/super_correos_masivos.go#L56) | DDL / indice / funcion | superadministrador o por confirmar |
| `EnsureSuperErroresSistemaSchema` | [backend/db/super_errores_sistema.go:122](../../backend/db/super_errores_sistema.go#L122) | DDL / indice / funcion | superadministrador o por confirmar |
| `EnsureSuperMantenimientoAgentesSchema` | [backend/db/super_mantenimiento_agentes.go:52](../../backend/db/super_mantenimiento_agentes.go#L52) | DDL / indice / funcion | superadministrador o por confirmar |
| `EnsureSuperServidorEventosSchema` | [backend/db/super_servidor_eventos.go:42](../../backend/db/super_servidor_eventos.go#L42) | DDL / indice / funcion | superadministrador o por confirmar |
| `EnsureSuperVPSSnapshotSchema` | [backend/db/super_vps_snapshots.go:32](../../backend/db/super_vps_snapshots.go#L32) | DDL / indice / funcion | superadministrador o por confirmar |
| `EnsureEmpresaTarifasMotelSchema` | [backend/db/tarifas_motel.go:75](../../backend/db/tarifas_motel.go#L75) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaTarifasPorDiaSchema` | [backend/db/tarifas_por_dia.go:110](../../backend/db/tarifas_por_dia.go#L110) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaTarifasPorMinutosSchema` | [backend/db/tarifas_por_minutos.go:109](../../backend/db/tarifas_por_minutos.go#L109) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaTarifasPorMinutosConfiguracionSchema` | [backend/db/tarifas_por_minutos.go:214](../../backend/db/tarifas_por_minutos.go#L214) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaTesoreriaPresupuestoSchema` | [backend/db/tesoreria_presupuesto.go:114](../../backend/db/tesoreria_presupuesto.go#L114) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureAyudaTicketsSchema` | [backend/db/tickets_ayuda.go:85](../../backend/db/tickets_ayuda.go#L85) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureTipoEmpresaPreconfiguracionSchema` | [backend/db/tipo_empresa_preconfiguracion.go:375](../../backend/db/tipo_empresa_preconfiguracion.go#L375) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureCanonicalTiposEmpresaPreconfigurables` | [backend/db/tipo_empresa_preconfiguracion.go:688](../../backend/db/tipo_empresa_preconfiguracion.go#L688) | regla auxiliar o verificacion | empresas o por confirmar |
| `EnsureDefaultRolesForTipoEmpresaPreconfiguraciones` | [backend/db/tipo_empresa_preconfiguracion.go:1368](../../backend/db/tipo_empresa_preconfiguracion.go#L1368) | seed o provisionamiento idempotente | empresas o por confirmar |
| `EnsureEnergiaSolarInTipoEmpresaPreconfiguraciones` | [backend/db/tipo_empresa_preconfiguracion.go:1770](../../backend/db/tipo_empresa_preconfiguracion.go#L1770) | seed o provisionamiento idempotente | empresas o por confirmar |
| `EnsureEmpresaTurnosAtencionSchema` | [backend/db/turnos_atencion.go:109](../../backend/db/turnos_atencion.go#L109) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaUbicacionGPSSchema` | [backend/db/ubicacion_gps.go:67](../../backend/db/ubicacion_gps.go#L67) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureUsuarioConfiguracionSchema` | [backend/db/usuario_config_schema.go:12](../../backend/db/usuario_config_schema.go#L12) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaUsuariosAuthSchema` | [backend/db/usuarios_empresa.go:56](../../backend/db/usuarios_empresa.go#L56) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaVehiculosRegistroSchema` | [backend/db/vehiculos_registro.go:73](../../backend/db/vehiculos_registro.go#L73) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureSuperVentaDigitalSchema` | [backend/db/venta_digital.go:170](../../backend/db/venta_digital.go#L170) | DDL / indice / funcion | superadministrador o por confirmar |
| `EnsureVentaPublicaSchema` | [backend/db/venta_publica.go:21](../../backend/db/venta_publica.go#L21) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureEmpresaVentaPublicaSchema` | [backend/db/venta_publica.go:563](../../backend/db/venta_publica.go#L563) | DDL / indice / funcion | empresas o por confirmar |
| `EnsureCorporateEmailConfigFromEnv` | [backend/handlers/email_corporativo_handlers.go:279](../../backend/handlers/email_corporativo_handlers.go#L279) | provisionamiento de integracion | empresas o por confirmar |
| `EnsureNextcloudConfigFromEnv` | [backend/handlers/nextcloud.go:271](../../backend/handlers/nextcloud.go#L271) | provisionamiento de integracion | empresas o por confirmar |
| `EnsureSensitiveSuperConfigEncrypted` | [backend/handlers/super_config_backup_handlers.go:59](../../backend/handlers/super_config_backup_handlers.go#L59) | provisionamiento de integracion | superadministrador o por confirmar |
| `EnsureSuperContextoIALogicaNegocio` | [backend/handlers/super_portal_chat_ia_info.go:182](../../backend/handlers/super_portal_chat_ia_info.go#L182) | provisionamiento de integracion | superadministrador o por confirmar |

## Gate de retiro

1. Catalogar cada fila DDL en `db.PlatformMigrations` o una migracion de dominio equivalente con checksum.
2. Mover seeds/provisionamientos a jobs versionados y explicitos, no al arranque de API.
3. Repetir migraciones en staging, comparar esquema y ejecutar pruebas operativas antes de cambiar `PCS_RUNTIME_SCHEMA_BOOTSTRAP=0`.
4. Mantener este inventario sincronizado mediante el preflight.
