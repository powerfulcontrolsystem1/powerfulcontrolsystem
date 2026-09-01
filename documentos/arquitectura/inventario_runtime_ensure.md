# Inventario de llamadas Ensure fuera del migrador

Estado: generado. Actualizar con `node tools/runtime_ensure_inventory.mjs`.

Las llamadas listadas son deuda de extraccion. En produccion, API y worker deben llegar a verificar esquema versionado, no crear o alterar tablas. El guard de runtime es una defensa adicional, no una sustitucion de esta migracion de codigo.

## Resumen

- Llamadas inventariadas: 102.
- arranque; protegido por rol, requiere extraccion: 68.
- proceso de plataforma; revisar rol: 3.
- trafico HTTP; priorizar reemplazo por verificacion: 31.

## Registro

| Funcion Ensure | Llamador | Riesgo / prioridad |
| --- | --- | --- |
| `EnsurePostgresRuntimeCompat` | [backend/cmd/pcs-migrate/main.go:69](../../backend/cmd/pcs-migrate/main.go#L69) | proceso de plataforma; revisar rol |
| `EnsureRuntimeDatabaseRole` | [backend/cmd/pcs-migrate/main.go:83](../../backend/cmd/pcs-migrate/main.go#L83) | proceso de plataforma; revisar rol |
| `EnsureRuntimeDatabaseRole` | [backend/cmd/pcs-migrate/main.go:86](../../backend/cmd/pcs-migrate/main.go#L86) | proceso de plataforma; revisar rol |
| `EnsureEmpresaControlElectricoPrimaryRaspberry` | [backend/handlers/control_electrico.go:311](../../backend/handlers/control_electrico.go#L311) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaControlElectricoPrimaryRaspberry` | [backend/handlers/control_electrico.go:502](../../backend/handlers/control_electrico.go#L502) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaEventosContablesSchema` | [backend/handlers/creditos.go:981](../../backend/handlers/creditos.go#L981) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaEmailRowsForExistingEmpresas` | [backend/handlers/email_corporativo_handlers.go:752](../../backend/handlers/email_corporativo_handlers.go#L752) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaEmailRowsForExistingEmpresas` | [backend/handlers/email_corporativo_handlers.go:1397](../../backend/handlers/email_corporativo_handlers.go#L1397) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaCorporateEmailAfterCreate` | [backend/handlers/email_corporativo_handlers.go:1642](../../backend/handlers/email_corporativo_handlers.go#L1642) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaPermisosFinosSchema` | [backend/handlers/empresa_permisos.go:816](../../backend/handlers/empresa_permisos.go#L816) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureNuevasPlantillasProduccionMasivaLicencias` | [backend/handlers/empresa_plantillas_nuevas.go:82](../../backend/handlers/empresa_plantillas_nuevas.go#L82) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaProductosSchema` | [backend/handlers/empresa_preconfiguracion.go:77](../../backend/handlers/empresa_preconfiguracion.go#L77) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaUsuariosAuthSchema` | [backend/handlers/empresa_preconfiguracion.go:81](../../backend/handlers/empresa_preconfiguracion.go#L81) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaConfiguracionOperativaSchema` | [backend/handlers/empresa_preconfiguracion.go:386](../../backend/handlers/empresa_preconfiguracion.go#L386) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaComisionesServicioSchema` | [backend/handlers/empresa_preconfiguracion.go:430](../../backend/handlers/empresa_preconfiguracion.go#L430) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaTarifasPorMinutosSchema` | [backend/handlers/empresa_preconfiguracion.go:989](../../backend/handlers/empresa_preconfiguracion.go#L989) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaTarifasPorDiaSchema` | [backend/handlers/empresa_preconfiguracion.go:1021](../../backend/handlers/empresa_preconfiguracion.go#L1021) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaTarifasMotelSchema` | [backend/handlers/empresa_preconfiguracion.go:1054](../../backend/handlers/empresa_preconfiguracion.go#L1054) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaControlElectricoSchema` | [backend/handlers/empresa_preconfiguracion.go:1195](../../backend/handlers/empresa_preconfiguracion.go#L1195) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaConfiguracionOperativaSchema` | [backend/handlers/empresa_preconfiguracion.go:1347](../../backend/handlers/empresa_preconfiguracion.go#L1347) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaComisionesServicioSchema` | [backend/handlers/empresa_preconfiguracion.go:1363](../../backend/handlers/empresa_preconfiguracion.go#L1363) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaEventosContablesSchema` | [backend/handlers/modulos_faltantes.go:2326](../../backend/handlers/modulos_faltantes.go#L2326) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaNextcloudSchema` | [backend/handlers/nextcloud.go:211](../../backend/handlers/nextcloud.go#L211) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaNextcloudAssignmentsForAll` | [backend/handlers/nextcloud.go:229](../../backend/handlers/nextcloud.go#L229) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureNextcloudAssignmentsForAll` | [backend/handlers/nextcloud.go:633](../../backend/handlers/nextcloud.go#L633) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaRappiSchema` | [backend/handlers/rappi.go:54](../../backend/handlers/rappi.go#L54) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureRolesPermisosSchema` | [backend/handlers/roles_tipos_usuario.go:134](../../backend/handlers/roles_tipos_usuario.go#L134) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureSuperAlertasSchema` | [backend/handlers/super_alertas.go:539](../../backend/handlers/super_alertas.go#L539) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaUsuariosAuthSchema` | [backend/handlers/super_correos_masivos.go:302](../../backend/handlers/super_correos_masivos.go#L302) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureSuperMantenimientoAgentesSchema` | [backend/handlers/super_mantenimiento_agentes.go:56](../../backend/handlers/super_mantenimiento_agentes.go#L56) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaNextcloudAssignment` | [backend/handlers/system_empresas_handlers.go:580](../../backend/handlers/system_empresas_handlers.go#L580) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaCorporateEmailAfterCreate` | [backend/handlers/system_empresas_handlers.go:584](../../backend/handlers/system_empresas_handlers.go#L584) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureAuthToken` | [backend/handlers/voice_stream_config.go:239](../../backend/handlers/voice_stream_config.go#L239) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureAuthToken` | [backend/handlers/voice_stream_config.go:349](../../backend/handlers/voice_stream_config.go#L349) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsurePostgresRuntimeCompat` | [backend/main.go:1171](../../backend/main.go#L1171) | arranque; protegido por rol, requiere extraccion |
| `EnsurePostgresRuntimeCompat` | [backend/main.go:1176](../../backend/main.go#L1176) | arranque; protegido por rol, requiere extraccion |
| `EnsureAdministradoresAuthSchema` | [backend/main.go:1180](../../backend/main.go#L1180) | arranque; protegido por rol, requiere extraccion |
| `EnsurePaymentGatewaySchema` | [backend/main.go:1219](../../backend/main.go#L1219) | arranque; protegido por rol, requiere extraccion |
| `EnsureLicenciasSchema` | [backend/main.go:1223](../../backend/main.go#L1223) | arranque; protegido por rol, requiere extraccion |
| `EnsureLicenciasCatalogoGlobal` | [backend/main.go:1227](../../backend/main.go#L1227) | arranque; protegido por rol, requiere extraccion |
| `EnsurePowerfulSystemEmpresa` | [backend/main.go:1233](../../backend/main.go#L1233) | arranque; protegido por rol, requiere extraccion |
| `EnsureSuperAuditoriaSchema` | [backend/main.go:1239](../../backend/main.go#L1239) | arranque; protegido por rol, requiere extraccion |
| `EnsureSuperVPSSnapshotSchema` | [backend/main.go:1243](../../backend/main.go#L1243) | arranque; protegido por rol, requiere extraccion |
| `EnsureLicenciaVencimientoNotificacionesSchema` | [backend/main.go:1247](../../backend/main.go#L1247) | arranque; protegido por rol, requiere extraccion |
| `EnsureLicenciaEmpresaRetencionSchema` | [backend/main.go:1251](../../backend/main.go#L1251) | arranque; protegido por rol, requiere extraccion |
| `EnsureUsuarioConfiguracionSchema` | [backend/main.go:1255](../../backend/main.go#L1255) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaEmailCorporativoSchema` | [backend/main.go:1259](../../backend/main.go#L1259) | arranque; protegido por rol, requiere extraccion |
| `EnsureCorporateEmailConfigFromEnv` | [backend/main.go:1263](../../backend/main.go#L1263) | arranque; protegido por rol, requiere extraccion |
| `EnsureCorporateEmailRowsForExistingCompanies` | [backend/main.go:1269](../../backend/main.go#L1269) | arranque; protegido por rol, requiere extraccion |
| `EnsureCorporateEmailProvisioningForExistingCompanies` | [backend/main.go:1277](../../backend/main.go#L1277) | arranque; protegido por rol, requiere extraccion |
| `EnsureNextcloudConfigFromEnv` | [backend/main.go:1283](../../backend/main.go#L1283) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaNextcloudSchema` | [backend/main.go:1287](../../backend/main.go#L1287) | arranque; protegido por rol, requiere extraccion |
| `EnsureNextcloudAssignmentsForAll` | [backend/main.go:1291](../../backend/main.go#L1291) | arranque; protegido por rol, requiere extraccion |
| `EnsureAsesorComercialSchema` | [backend/main.go:1301](../../backend/main.go#L1301) | arranque; protegido por rol, requiere extraccion |
| `EnsureConstructoraTipoEmpresaYLicencias` | [backend/main.go:1311](../../backend/main.go#L1311) | arranque; protegido por rol, requiere extraccion |
| `EnsureAlquileresTipoEmpresaYLicencias` | [backend/main.go:1317](../../backend/main.go#L1317) | arranque; protegido por rol, requiere extraccion |
| `EnsureNuevasPlantillasTipoEmpresaYLicencias` | [backend/main.go:1323](../../backend/main.go#L1323) | arranque; protegido por rol, requiere extraccion |
| `EnsureEnergiaSolarInTipoEmpresaPreconfiguraciones` | [backend/main.go:1333](../../backend/main.go#L1333) | arranque; protegido por rol, requiere extraccion |
| `EnsurePostgresRuntimeCompat` | [backend/main.go:1347](../../backend/main.go#L1347) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaUsuariosAuthSchema` | [backend/main.go:1350](../../backend/main.go#L1350) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaBuzonSchema` | [backend/main.go:1356](../../backend/main.go#L1356) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCarritosSchema` | [backend/main.go:1360](../../backend/main.go#L1360) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaDatafonosSchema` | [backend/main.go:1364](../../backend/main.go#L1364) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaFinanzasSchema` | [backend/main.go:1380](../../backend/main.go#L1380) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaImpuestosSchema` | [backend/main.go:1384](../../backend/main.go#L1384) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaNominaSchema` | [backend/main.go:1388](../../backend/main.go#L1388) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCreditosSchema` | [backend/main.go:1406](../../backend/main.go#L1406) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaContabilidadColombiaSchema` | [backend/main.go:1410](../../backend/main.go#L1410) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaContabilidadColombiaAvanzadaSchema` | [backend/main.go:1414](../../backend/main.go#L1414) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCentrosCostoSchema` | [backend/main.go:1418](../../backend/main.go#L1418) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCierreFiscalSchema` | [backend/main.go:1422](../../backend/main.go#L1422) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaDeclaracionesTributariasSchema` | [backend/main.go:1426](../../backend/main.go#L1426) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaTesoreriaPresupuestoSchema` | [backend/main.go:1430](../../backend/main.go#L1430) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaImportacionesCosteoSchema` | [backend/main.go:1433](../../backend/main.go#L1433) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaAIUConstruccionSchema` | [backend/main.go:1436](../../backend/main.go#L1436) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCobranzaSchema` | [backend/main.go:1439](../../backend/main.go#L1439) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaPortalContadorSchema` | [backend/main.go:1442](../../backend/main.go#L1442) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaPortalTercerosCertificadosSchema` | [backend/main.go:1445](../../backend/main.go#L1445) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaSoportesComprasIASchema` | [backend/main.go:1448](../../backend/main.go#L1448) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaModulosColombiaSchema` | [backend/main.go:1451](../../backend/main.go#L1451) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaComprasAvanzadasSchema` | [backend/main.go:1454](../../backend/main.go#L1454) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaReservasHotelSchema` | [backend/main.go:1457](../../backend/main.go#L1457) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaTarifasMotelSchema` | [backend/main.go:1461](../../backend/main.go#L1461) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaAIEnterpriseSchema` | [backend/main.go:1465](../../backend/main.go#L1465) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaAIOpenAIProviderSchema` | [backend/main.go:1469](../../backend/main.go#L1469) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaSensorPuertasSchema` | [backend/main.go:1473](../../backend/main.go#L1473) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaControlElectricoSchema` | [backend/main.go:1476](../../backend/main.go#L1476) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaEnergiaSolarSchema` | [backend/main.go:1479](../../backend/main.go#L1479) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCamarasSchema` | [backend/main.go:1482](../../backend/main.go#L1482) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCarnetsSchema` | [backend/main.go:1485](../../backend/main.go#L1485) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaParqueaderoSchema` | [backend/main.go:1488](../../backend/main.go#L1488) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaProduccionMRPSchema` | [backend/main.go:1491](../../backend/main.go#L1491) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaWMSSchema` | [backend/main.go:1494](../../backend/main.go#L1494) | arranque; protegido por rol, requiere extraccion |
| `EnsureHotelTarjetasAccesoSchema` | [backend/main.go:1497](../../backend/main.go#L1497) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaProductosSchema` | [backend/main.go:1500](../../backend/main.go#L1500) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaInventarioAvanzadoSchema` | [backend/main.go:1508](../../backend/main.go#L1508) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCRMVentasAvanzadasSchema` | [backend/main.go:1511](../../backend/main.go#L1511) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaSoporteRemotoSchema` | [backend/main.go:1514](../../backend/main.go#L1514) | arranque; protegido por rol, requiere extraccion |
| `EnsureSensitiveSuperConfigEncrypted` | [backend/main.go:1524](../../backend/main.go#L1524) | arranque; protegido por rol, requiere extraccion |
| `EnsurePostgresPrimaryKeySequences` | [backend/main.go:1528](../../backend/main.go#L1528) | arranque; protegido por rol, requiere extraccion |
| `EnsurePostgresPrimaryKeySequences` | [backend/main.go:1532](../../backend/main.go#L1532) | arranque; protegido por rol, requiere extraccion |
| `EnsureSuperContextoIALogicaNegocio` | [backend/main.go:1538](../../backend/main.go#L1538) | arranque; protegido por rol, requiere extraccion |

## Gate de retiro

1. No agregar nuevas filas: el preflight exige que este inventario coincida con el codigo.
2. Reemplazar primero llamadas en handlers de pagos, facturacion, inventario, archivos y autenticacion por verificadores de esquema o migraciones catalogadas.
3. Cada extraccion debe incluir prueba de base actualizada y de esquema faltante que falle cerrado, sin DDL desde la solicitud.
4. Solo `pcs-migrate` conserva el bootstrap del ledger y las migraciones inmutables.
