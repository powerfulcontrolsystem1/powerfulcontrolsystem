# Inventario de llamadas Ensure fuera del migrador

Estado: generado. Actualizar con `node tools/runtime_ensure_inventory.mjs`.

Las llamadas listadas son deuda de extraccion. En produccion, API y worker deben llegar a verificar esquema versionado, no crear o alterar tablas. El guard de runtime es una defensa adicional, no una sustitucion de esta migracion de codigo.

## Resumen

- Llamadas inventariadas: 104.
- arranque; protegido por rol, requiere extraccion: 72.
- proceso de plataforma; revisar rol: 1.
- trafico HTTP; priorizar reemplazo por verificacion: 31.

## Registro

| Funcion Ensure | Llamador | Riesgo / prioridad |
| --- | --- | --- |
| `EnsurePostgresRuntimeCompat` | [backend/cmd/pcs-migrate/main.go:69](../../backend/cmd/pcs-migrate/main.go#L69) | proceso de plataforma; revisar rol |
| `EnsureEmpresaControlElectricoPrimaryRaspberry` | [backend/handlers/control_electrico.go:224](../../backend/handlers/control_electrico.go#L224) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaControlElectricoPrimaryRaspberry` | [backend/handlers/control_electrico.go:384](../../backend/handlers/control_electrico.go#L384) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaEventosContablesSchema` | [backend/handlers/creditos.go:981](../../backend/handlers/creditos.go#L981) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaEmailRowsForExistingEmpresas` | [backend/handlers/email_corporativo_handlers.go:752](../../backend/handlers/email_corporativo_handlers.go#L752) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaEmailRowsForExistingEmpresas` | [backend/handlers/email_corporativo_handlers.go:1389](../../backend/handlers/email_corporativo_handlers.go#L1389) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaCorporateEmailAfterCreate` | [backend/handlers/email_corporativo_handlers.go:1634](../../backend/handlers/email_corporativo_handlers.go#L1634) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaPermisosFinosSchema` | [backend/handlers/empresa_permisos.go:845](../../backend/handlers/empresa_permisos.go#L845) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureNuevasPlantillasProduccionMasivaLicencias` | [backend/handlers/empresa_plantillas_nuevas.go:82](../../backend/handlers/empresa_plantillas_nuevas.go#L82) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaProductosSchema` | [backend/handlers/empresa_preconfiguracion.go:77](../../backend/handlers/empresa_preconfiguracion.go#L77) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaUsuariosAuthSchema` | [backend/handlers/empresa_preconfiguracion.go:81](../../backend/handlers/empresa_preconfiguracion.go#L81) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaConfiguracionOperativaSchema` | [backend/handlers/empresa_preconfiguracion.go:386](../../backend/handlers/empresa_preconfiguracion.go#L386) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaComisionesServicioSchema` | [backend/handlers/empresa_preconfiguracion.go:430](../../backend/handlers/empresa_preconfiguracion.go#L430) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaTarifasPorMinutosSchema` | [backend/handlers/empresa_preconfiguracion.go:989](../../backend/handlers/empresa_preconfiguracion.go#L989) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaTarifasPorDiaSchema` | [backend/handlers/empresa_preconfiguracion.go:1021](../../backend/handlers/empresa_preconfiguracion.go#L1021) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaTarifasMotelSchema` | [backend/handlers/empresa_preconfiguracion.go:1054](../../backend/handlers/empresa_preconfiguracion.go#L1054) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaControlElectricoSchema` | [backend/handlers/empresa_preconfiguracion.go:1304](../../backend/handlers/empresa_preconfiguracion.go#L1304) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaConfiguracionOperativaSchema` | [backend/handlers/empresa_preconfiguracion.go:1456](../../backend/handlers/empresa_preconfiguracion.go#L1456) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaComisionesServicioSchema` | [backend/handlers/empresa_preconfiguracion.go:1472](../../backend/handlers/empresa_preconfiguracion.go#L1472) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaEventosContablesSchema` | [backend/handlers/modulos_faltantes.go:2323](../../backend/handlers/modulos_faltantes.go#L2323) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaNextcloudSchema` | [backend/handlers/nextcloud.go:211](../../backend/handlers/nextcloud.go#L211) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaNextcloudAssignmentsForAll` | [backend/handlers/nextcloud.go:229](../../backend/handlers/nextcloud.go#L229) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureNextcloudAssignmentsForAll` | [backend/handlers/nextcloud.go:633](../../backend/handlers/nextcloud.go#L633) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaRappiSchema` | [backend/handlers/rappi.go:53](../../backend/handlers/rappi.go#L53) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureRolesPermisosSchema` | [backend/handlers/roles_tipos_usuario.go:134](../../backend/handlers/roles_tipos_usuario.go#L134) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureSuperAlertasSchema` | [backend/handlers/super_alertas.go:539](../../backend/handlers/super_alertas.go#L539) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaUsuariosAuthSchema` | [backend/handlers/super_correos_masivos.go:302](../../backend/handlers/super_correos_masivos.go#L302) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureSuperMantenimientoAgentesSchema` | [backend/handlers/super_mantenimiento_agentes.go:56](../../backend/handlers/super_mantenimiento_agentes.go#L56) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaNextcloudAssignment` | [backend/handlers/system_empresas_handlers.go:580](../../backend/handlers/system_empresas_handlers.go#L580) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaCorporateEmailAfterCreate` | [backend/handlers/system_empresas_handlers.go:584](../../backend/handlers/system_empresas_handlers.go#L584) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureAuthToken` | [backend/handlers/voice_stream_config.go:239](../../backend/handlers/voice_stream_config.go#L239) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureAuthToken` | [backend/handlers/voice_stream_config.go:349](../../backend/handlers/voice_stream_config.go#L349) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsurePostgresRuntimeCompat` | [backend/main.go:905](../../backend/main.go#L905) | arranque; protegido por rol, requiere extraccion |
| `EnsurePostgresRuntimeCompat` | [backend/main.go:910](../../backend/main.go#L910) | arranque; protegido por rol, requiere extraccion |
| `EnsureAdministradoresAuthSchema` | [backend/main.go:914](../../backend/main.go#L914) | arranque; protegido por rol, requiere extraccion |
| `EnsurePaymentGatewaySchema` | [backend/main.go:953](../../backend/main.go#L953) | arranque; protegido por rol, requiere extraccion |
| `EnsureLicenciasSchema` | [backend/main.go:957](../../backend/main.go#L957) | arranque; protegido por rol, requiere extraccion |
| `EnsureLicenciasCatalogoGlobal` | [backend/main.go:961](../../backend/main.go#L961) | arranque; protegido por rol, requiere extraccion |
| `EnsurePowerfulSystemEmpresa` | [backend/main.go:967](../../backend/main.go#L967) | arranque; protegido por rol, requiere extraccion |
| `EnsureSuperAuditoriaSchema` | [backend/main.go:973](../../backend/main.go#L973) | arranque; protegido por rol, requiere extraccion |
| `EnsureSuperVPSSnapshotSchema` | [backend/main.go:977](../../backend/main.go#L977) | arranque; protegido por rol, requiere extraccion |
| `EnsureLicenciaVencimientoNotificacionesSchema` | [backend/main.go:981](../../backend/main.go#L981) | arranque; protegido por rol, requiere extraccion |
| `EnsureLicenciaEmpresaRetencionSchema` | [backend/main.go:985](../../backend/main.go#L985) | arranque; protegido por rol, requiere extraccion |
| `EnsureUsuarioConfiguracionSchema` | [backend/main.go:989](../../backend/main.go#L989) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaEmailCorporativoSchema` | [backend/main.go:993](../../backend/main.go#L993) | arranque; protegido por rol, requiere extraccion |
| `EnsureCorporateEmailConfigFromEnv` | [backend/main.go:997](../../backend/main.go#L997) | arranque; protegido por rol, requiere extraccion |
| `EnsureCorporateEmailRowsForExistingCompanies` | [backend/main.go:1003](../../backend/main.go#L1003) | arranque; protegido por rol, requiere extraccion |
| `EnsureCorporateEmailProvisioningForExistingCompanies` | [backend/main.go:1011](../../backend/main.go#L1011) | arranque; protegido por rol, requiere extraccion |
| `EnsureNextcloudConfigFromEnv` | [backend/main.go:1017](../../backend/main.go#L1017) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaNextcloudSchema` | [backend/main.go:1021](../../backend/main.go#L1021) | arranque; protegido por rol, requiere extraccion |
| `EnsureNextcloudAssignmentsForAll` | [backend/main.go:1025](../../backend/main.go#L1025) | arranque; protegido por rol, requiere extraccion |
| `EnsureAsesorComercialSchema` | [backend/main.go:1035](../../backend/main.go#L1035) | arranque; protegido por rol, requiere extraccion |
| `EnsureConstructoraTipoEmpresaYLicencias` | [backend/main.go:1045](../../backend/main.go#L1045) | arranque; protegido por rol, requiere extraccion |
| `EnsureDrogueriaFarmaciaTipoEmpresaYLicencias` | [backend/main.go:1051](../../backend/main.go#L1051) | arranque; protegido por rol, requiere extraccion |
| `EnsureAlquileresTipoEmpresaYLicencias` | [backend/main.go:1057](../../backend/main.go#L1057) | arranque; protegido por rol, requiere extraccion |
| `EnsureNuevasPlantillasTipoEmpresaYLicencias` | [backend/main.go:1063](../../backend/main.go#L1063) | arranque; protegido por rol, requiere extraccion |
| `EnsureEnergiaSolarInTipoEmpresaPreconfiguraciones` | [backend/main.go:1073](../../backend/main.go#L1073) | arranque; protegido por rol, requiere extraccion |
| `EnsurePostgresRuntimeCompat` | [backend/main.go:1087](../../backend/main.go#L1087) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaUsuariosAuthSchema` | [backend/main.go:1090](../../backend/main.go#L1090) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaBuzonSchema` | [backend/main.go:1096](../../backend/main.go#L1096) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCarritosSchema` | [backend/main.go:1100](../../backend/main.go#L1100) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaDatafonosSchema` | [backend/main.go:1104](../../backend/main.go#L1104) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaFinanzasSchema` | [backend/main.go:1120](../../backend/main.go#L1120) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaImpuestosSchema` | [backend/main.go:1124](../../backend/main.go#L1124) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaNominaSchema` | [backend/main.go:1128](../../backend/main.go#L1128) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCreditosSchema` | [backend/main.go:1146](../../backend/main.go#L1146) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaContabilidadColombiaSchema` | [backend/main.go:1150](../../backend/main.go#L1150) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaContabilidadColombiaAvanzadaSchema` | [backend/main.go:1154](../../backend/main.go#L1154) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCentrosCostoSchema` | [backend/main.go:1158](../../backend/main.go#L1158) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCierreFiscalSchema` | [backend/main.go:1162](../../backend/main.go#L1162) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaDeclaracionesTributariasSchema` | [backend/main.go:1166](../../backend/main.go#L1166) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaTesoreriaPresupuestoSchema` | [backend/main.go:1170](../../backend/main.go#L1170) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaImportacionesCosteoSchema` | [backend/main.go:1173](../../backend/main.go#L1173) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaAIUConstruccionSchema` | [backend/main.go:1176](../../backend/main.go#L1176) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCobranzaSchema` | [backend/main.go:1179](../../backend/main.go#L1179) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaPortalContadorSchema` | [backend/main.go:1182](../../backend/main.go#L1182) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaPortalTercerosCertificadosSchema` | [backend/main.go:1185](../../backend/main.go#L1185) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaSoportesComprasIASchema` | [backend/main.go:1188](../../backend/main.go#L1188) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaModulosColombiaSchema` | [backend/main.go:1191](../../backend/main.go#L1191) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaComprasAvanzadasSchema` | [backend/main.go:1194](../../backend/main.go#L1194) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaReservasHotelSchema` | [backend/main.go:1197](../../backend/main.go#L1197) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaTarifasMotelSchema` | [backend/main.go:1201](../../backend/main.go#L1201) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaAIEnterpriseSchema` | [backend/main.go:1205](../../backend/main.go#L1205) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaAIOpenAIProviderSchema` | [backend/main.go:1209](../../backend/main.go#L1209) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaSensorPuertasSchema` | [backend/main.go:1213](../../backend/main.go#L1213) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaControlElectricoSchema` | [backend/main.go:1216](../../backend/main.go#L1216) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaEnergiaSolarSchema` | [backend/main.go:1219](../../backend/main.go#L1219) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCamarasSchema` | [backend/main.go:1222](../../backend/main.go#L1222) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaGrafologiaSchema` | [backend/main.go:1225](../../backend/main.go#L1225) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCarnetsSchema` | [backend/main.go:1228](../../backend/main.go#L1228) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaParqueaderoSchema` | [backend/main.go:1231](../../backend/main.go#L1231) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaApartamentosTuristicosSchema` | [backend/main.go:1234](../../backend/main.go#L1234) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaPropiedadHorizontalSchema` | [backend/main.go:1237](../../backend/main.go#L1237) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaProduccionMRPSchema` | [backend/main.go:1240](../../backend/main.go#L1240) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaWMSSchema` | [backend/main.go:1243](../../backend/main.go#L1243) | arranque; protegido por rol, requiere extraccion |
| `EnsureHotelTarjetasAccesoSchema` | [backend/main.go:1246](../../backend/main.go#L1246) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaProductosSchema` | [backend/main.go:1249](../../backend/main.go#L1249) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaInventarioAvanzadoSchema` | [backend/main.go:1257](../../backend/main.go#L1257) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCRMVentasAvanzadasSchema` | [backend/main.go:1260](../../backend/main.go#L1260) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaSoporteRemotoSchema` | [backend/main.go:1263](../../backend/main.go#L1263) | arranque; protegido por rol, requiere extraccion |
| `EnsureSensitiveSuperConfigEncrypted` | [backend/main.go:1273](../../backend/main.go#L1273) | arranque; protegido por rol, requiere extraccion |
| `EnsurePostgresPrimaryKeySequences` | [backend/main.go:1277](../../backend/main.go#L1277) | arranque; protegido por rol, requiere extraccion |
| `EnsurePostgresPrimaryKeySequences` | [backend/main.go:1281](../../backend/main.go#L1281) | arranque; protegido por rol, requiere extraccion |
| `EnsureSuperContextoIALogicaNegocio` | [backend/main.go:1287](../../backend/main.go#L1287) | arranque; protegido por rol, requiere extraccion |

## Gate de retiro

1. No agregar nuevas filas: el preflight exige que este inventario coincida con el codigo.
2. Reemplazar primero llamadas en handlers de pagos, facturacion, inventario, archivos y autenticacion por verificadores de esquema o migraciones catalogadas.
3. Cada extraccion debe incluir prueba de base actualizada y de esquema faltante que falle cerrado, sin DDL desde la solicitud.
4. Solo `pcs-migrate` conserva el bootstrap del ledger y las migraciones inmutables.
