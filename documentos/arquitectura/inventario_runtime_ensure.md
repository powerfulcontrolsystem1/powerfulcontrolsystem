# Inventario de llamadas Ensure fuera del migrador

Estado: generado. Actualizar con `node tools/runtime_ensure_inventory.mjs`.

Las llamadas listadas son deuda de extraccion. En produccion, API y worker deben llegar a verificar esquema versionado, no crear o alterar tablas. El guard de runtime es una defensa adicional, no una sustitucion de esta migracion de codigo.

## Resumen

- Llamadas inventariadas: 106.
- arranque; protegido por rol, requiere extraccion: 72.
- proceso de plataforma; revisar rol: 3.
- trafico HTTP; priorizar reemplazo por verificacion: 31.

## Registro

| Funcion Ensure | Llamador | Riesgo / prioridad |
| --- | --- | --- |
| `EnsurePostgresRuntimeCompat` | [backend/cmd/pcs-migrate/main.go:69](../../backend/cmd/pcs-migrate/main.go#L69) | proceso de plataforma; revisar rol |
| `EnsureRuntimeDatabaseRole` | [backend/cmd/pcs-migrate/main.go:83](../../backend/cmd/pcs-migrate/main.go#L83) | proceso de plataforma; revisar rol |
| `EnsureRuntimeDatabaseRole` | [backend/cmd/pcs-migrate/main.go:86](../../backend/cmd/pcs-migrate/main.go#L86) | proceso de plataforma; revisar rol |
| `EnsureEmpresaControlElectricoPrimaryRaspberry` | [backend/handlers/control_electrico.go:224](../../backend/handlers/control_electrico.go#L224) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaControlElectricoPrimaryRaspberry` | [backend/handlers/control_electrico.go:384](../../backend/handlers/control_electrico.go#L384) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaEventosContablesSchema` | [backend/handlers/creditos.go:981](../../backend/handlers/creditos.go#L981) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaEmailRowsForExistingEmpresas` | [backend/handlers/email_corporativo_handlers.go:752](../../backend/handlers/email_corporativo_handlers.go#L752) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaEmailRowsForExistingEmpresas` | [backend/handlers/email_corporativo_handlers.go:1389](../../backend/handlers/email_corporativo_handlers.go#L1389) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaCorporateEmailAfterCreate` | [backend/handlers/email_corporativo_handlers.go:1634](../../backend/handlers/email_corporativo_handlers.go#L1634) | trafico HTTP; priorizar reemplazo por verificacion |
| `EnsureEmpresaPermisosFinosSchema` | [backend/handlers/empresa_permisos.go:846](../../backend/handlers/empresa_permisos.go#L846) | trafico HTTP; priorizar reemplazo por verificacion |
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
| `EnsureEmpresaEventosContablesSchema` | [backend/handlers/modulos_faltantes.go:2326](../../backend/handlers/modulos_faltantes.go#L2326) | trafico HTTP; priorizar reemplazo por verificacion |
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
| `EnsurePostgresRuntimeCompat` | [backend/main.go:1098](../../backend/main.go#L1098) | arranque; protegido por rol, requiere extraccion |
| `EnsurePostgresRuntimeCompat` | [backend/main.go:1103](../../backend/main.go#L1103) | arranque; protegido por rol, requiere extraccion |
| `EnsureAdministradoresAuthSchema` | [backend/main.go:1107](../../backend/main.go#L1107) | arranque; protegido por rol, requiere extraccion |
| `EnsurePaymentGatewaySchema` | [backend/main.go:1146](../../backend/main.go#L1146) | arranque; protegido por rol, requiere extraccion |
| `EnsureLicenciasSchema` | [backend/main.go:1150](../../backend/main.go#L1150) | arranque; protegido por rol, requiere extraccion |
| `EnsureLicenciasCatalogoGlobal` | [backend/main.go:1154](../../backend/main.go#L1154) | arranque; protegido por rol, requiere extraccion |
| `EnsurePowerfulSystemEmpresa` | [backend/main.go:1160](../../backend/main.go#L1160) | arranque; protegido por rol, requiere extraccion |
| `EnsureSuperAuditoriaSchema` | [backend/main.go:1166](../../backend/main.go#L1166) | arranque; protegido por rol, requiere extraccion |
| `EnsureSuperVPSSnapshotSchema` | [backend/main.go:1170](../../backend/main.go#L1170) | arranque; protegido por rol, requiere extraccion |
| `EnsureLicenciaVencimientoNotificacionesSchema` | [backend/main.go:1174](../../backend/main.go#L1174) | arranque; protegido por rol, requiere extraccion |
| `EnsureLicenciaEmpresaRetencionSchema` | [backend/main.go:1178](../../backend/main.go#L1178) | arranque; protegido por rol, requiere extraccion |
| `EnsureUsuarioConfiguracionSchema` | [backend/main.go:1182](../../backend/main.go#L1182) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaEmailCorporativoSchema` | [backend/main.go:1186](../../backend/main.go#L1186) | arranque; protegido por rol, requiere extraccion |
| `EnsureCorporateEmailConfigFromEnv` | [backend/main.go:1190](../../backend/main.go#L1190) | arranque; protegido por rol, requiere extraccion |
| `EnsureCorporateEmailRowsForExistingCompanies` | [backend/main.go:1196](../../backend/main.go#L1196) | arranque; protegido por rol, requiere extraccion |
| `EnsureCorporateEmailProvisioningForExistingCompanies` | [backend/main.go:1204](../../backend/main.go#L1204) | arranque; protegido por rol, requiere extraccion |
| `EnsureNextcloudConfigFromEnv` | [backend/main.go:1210](../../backend/main.go#L1210) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaNextcloudSchema` | [backend/main.go:1214](../../backend/main.go#L1214) | arranque; protegido por rol, requiere extraccion |
| `EnsureNextcloudAssignmentsForAll` | [backend/main.go:1218](../../backend/main.go#L1218) | arranque; protegido por rol, requiere extraccion |
| `EnsureAsesorComercialSchema` | [backend/main.go:1228](../../backend/main.go#L1228) | arranque; protegido por rol, requiere extraccion |
| `EnsureConstructoraTipoEmpresaYLicencias` | [backend/main.go:1238](../../backend/main.go#L1238) | arranque; protegido por rol, requiere extraccion |
| `EnsureDrogueriaFarmaciaTipoEmpresaYLicencias` | [backend/main.go:1244](../../backend/main.go#L1244) | arranque; protegido por rol, requiere extraccion |
| `EnsureAlquileresTipoEmpresaYLicencias` | [backend/main.go:1250](../../backend/main.go#L1250) | arranque; protegido por rol, requiere extraccion |
| `EnsureNuevasPlantillasTipoEmpresaYLicencias` | [backend/main.go:1256](../../backend/main.go#L1256) | arranque; protegido por rol, requiere extraccion |
| `EnsureEnergiaSolarInTipoEmpresaPreconfiguraciones` | [backend/main.go:1266](../../backend/main.go#L1266) | arranque; protegido por rol, requiere extraccion |
| `EnsurePostgresRuntimeCompat` | [backend/main.go:1280](../../backend/main.go#L1280) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaUsuariosAuthSchema` | [backend/main.go:1283](../../backend/main.go#L1283) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaBuzonSchema` | [backend/main.go:1289](../../backend/main.go#L1289) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCarritosSchema` | [backend/main.go:1293](../../backend/main.go#L1293) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaDatafonosSchema` | [backend/main.go:1297](../../backend/main.go#L1297) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaFinanzasSchema` | [backend/main.go:1313](../../backend/main.go#L1313) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaImpuestosSchema` | [backend/main.go:1317](../../backend/main.go#L1317) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaNominaSchema` | [backend/main.go:1321](../../backend/main.go#L1321) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCreditosSchema` | [backend/main.go:1339](../../backend/main.go#L1339) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaContabilidadColombiaSchema` | [backend/main.go:1343](../../backend/main.go#L1343) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaContabilidadColombiaAvanzadaSchema` | [backend/main.go:1347](../../backend/main.go#L1347) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCentrosCostoSchema` | [backend/main.go:1351](../../backend/main.go#L1351) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCierreFiscalSchema` | [backend/main.go:1355](../../backend/main.go#L1355) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaDeclaracionesTributariasSchema` | [backend/main.go:1359](../../backend/main.go#L1359) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaTesoreriaPresupuestoSchema` | [backend/main.go:1363](../../backend/main.go#L1363) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaImportacionesCosteoSchema` | [backend/main.go:1366](../../backend/main.go#L1366) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaAIUConstruccionSchema` | [backend/main.go:1369](../../backend/main.go#L1369) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCobranzaSchema` | [backend/main.go:1372](../../backend/main.go#L1372) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaPortalContadorSchema` | [backend/main.go:1375](../../backend/main.go#L1375) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaPortalTercerosCertificadosSchema` | [backend/main.go:1378](../../backend/main.go#L1378) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaSoportesComprasIASchema` | [backend/main.go:1381](../../backend/main.go#L1381) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaModulosColombiaSchema` | [backend/main.go:1384](../../backend/main.go#L1384) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaComprasAvanzadasSchema` | [backend/main.go:1387](../../backend/main.go#L1387) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaReservasHotelSchema` | [backend/main.go:1390](../../backend/main.go#L1390) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaTarifasMotelSchema` | [backend/main.go:1394](../../backend/main.go#L1394) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaAIEnterpriseSchema` | [backend/main.go:1398](../../backend/main.go#L1398) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaAIOpenAIProviderSchema` | [backend/main.go:1402](../../backend/main.go#L1402) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaSensorPuertasSchema` | [backend/main.go:1406](../../backend/main.go#L1406) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaControlElectricoSchema` | [backend/main.go:1409](../../backend/main.go#L1409) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaEnergiaSolarSchema` | [backend/main.go:1412](../../backend/main.go#L1412) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCamarasSchema` | [backend/main.go:1415](../../backend/main.go#L1415) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaGrafologiaSchema` | [backend/main.go:1418](../../backend/main.go#L1418) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCarnetsSchema` | [backend/main.go:1421](../../backend/main.go#L1421) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaParqueaderoSchema` | [backend/main.go:1424](../../backend/main.go#L1424) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaApartamentosTuristicosSchema` | [backend/main.go:1427](../../backend/main.go#L1427) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaPropiedadHorizontalSchema` | [backend/main.go:1430](../../backend/main.go#L1430) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaProduccionMRPSchema` | [backend/main.go:1433](../../backend/main.go#L1433) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaWMSSchema` | [backend/main.go:1436](../../backend/main.go#L1436) | arranque; protegido por rol, requiere extraccion |
| `EnsureHotelTarjetasAccesoSchema` | [backend/main.go:1439](../../backend/main.go#L1439) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaProductosSchema` | [backend/main.go:1442](../../backend/main.go#L1442) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaInventarioAvanzadoSchema` | [backend/main.go:1450](../../backend/main.go#L1450) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCRMVentasAvanzadasSchema` | [backend/main.go:1453](../../backend/main.go#L1453) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaSoporteRemotoSchema` | [backend/main.go:1456](../../backend/main.go#L1456) | arranque; protegido por rol, requiere extraccion |
| `EnsureSensitiveSuperConfigEncrypted` | [backend/main.go:1466](../../backend/main.go#L1466) | arranque; protegido por rol, requiere extraccion |
| `EnsurePostgresPrimaryKeySequences` | [backend/main.go:1470](../../backend/main.go#L1470) | arranque; protegido por rol, requiere extraccion |
| `EnsurePostgresPrimaryKeySequences` | [backend/main.go:1474](../../backend/main.go#L1474) | arranque; protegido por rol, requiere extraccion |
| `EnsureSuperContextoIALogicaNegocio` | [backend/main.go:1480](../../backend/main.go#L1480) | arranque; protegido por rol, requiere extraccion |

## Gate de retiro

1. No agregar nuevas filas: el preflight exige que este inventario coincida con el codigo.
2. Reemplazar primero llamadas en handlers de pagos, facturacion, inventario, archivos y autenticacion por verificadores de esquema o migraciones catalogadas.
3. Cada extraccion debe incluir prueba de base actualizada y de esquema faltante que falle cerrado, sin DDL desde la solicitud.
4. Solo `pcs-migrate` conserva el bootstrap del ledger y las migraciones inmutables.
