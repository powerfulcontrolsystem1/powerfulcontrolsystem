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
| `EnsurePostgresRuntimeCompat` | [backend/main.go:1166](../../backend/main.go#L1166) | arranque; protegido por rol, requiere extraccion |
| `EnsurePostgresRuntimeCompat` | [backend/main.go:1171](../../backend/main.go#L1171) | arranque; protegido por rol, requiere extraccion |
| `EnsureAdministradoresAuthSchema` | [backend/main.go:1175](../../backend/main.go#L1175) | arranque; protegido por rol, requiere extraccion |
| `EnsurePaymentGatewaySchema` | [backend/main.go:1214](../../backend/main.go#L1214) | arranque; protegido por rol, requiere extraccion |
| `EnsureLicenciasSchema` | [backend/main.go:1218](../../backend/main.go#L1218) | arranque; protegido por rol, requiere extraccion |
| `EnsureLicenciasCatalogoGlobal` | [backend/main.go:1222](../../backend/main.go#L1222) | arranque; protegido por rol, requiere extraccion |
| `EnsurePowerfulSystemEmpresa` | [backend/main.go:1228](../../backend/main.go#L1228) | arranque; protegido por rol, requiere extraccion |
| `EnsureSuperAuditoriaSchema` | [backend/main.go:1234](../../backend/main.go#L1234) | arranque; protegido por rol, requiere extraccion |
| `EnsureSuperVPSSnapshotSchema` | [backend/main.go:1238](../../backend/main.go#L1238) | arranque; protegido por rol, requiere extraccion |
| `EnsureLicenciaVencimientoNotificacionesSchema` | [backend/main.go:1242](../../backend/main.go#L1242) | arranque; protegido por rol, requiere extraccion |
| `EnsureLicenciaEmpresaRetencionSchema` | [backend/main.go:1246](../../backend/main.go#L1246) | arranque; protegido por rol, requiere extraccion |
| `EnsureUsuarioConfiguracionSchema` | [backend/main.go:1250](../../backend/main.go#L1250) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaEmailCorporativoSchema` | [backend/main.go:1254](../../backend/main.go#L1254) | arranque; protegido por rol, requiere extraccion |
| `EnsureCorporateEmailConfigFromEnv` | [backend/main.go:1258](../../backend/main.go#L1258) | arranque; protegido por rol, requiere extraccion |
| `EnsureCorporateEmailRowsForExistingCompanies` | [backend/main.go:1264](../../backend/main.go#L1264) | arranque; protegido por rol, requiere extraccion |
| `EnsureCorporateEmailProvisioningForExistingCompanies` | [backend/main.go:1272](../../backend/main.go#L1272) | arranque; protegido por rol, requiere extraccion |
| `EnsureNextcloudConfigFromEnv` | [backend/main.go:1278](../../backend/main.go#L1278) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaNextcloudSchema` | [backend/main.go:1282](../../backend/main.go#L1282) | arranque; protegido por rol, requiere extraccion |
| `EnsureNextcloudAssignmentsForAll` | [backend/main.go:1286](../../backend/main.go#L1286) | arranque; protegido por rol, requiere extraccion |
| `EnsureAsesorComercialSchema` | [backend/main.go:1296](../../backend/main.go#L1296) | arranque; protegido por rol, requiere extraccion |
| `EnsureConstructoraTipoEmpresaYLicencias` | [backend/main.go:1306](../../backend/main.go#L1306) | arranque; protegido por rol, requiere extraccion |
| `EnsureDrogueriaFarmaciaTipoEmpresaYLicencias` | [backend/main.go:1312](../../backend/main.go#L1312) | arranque; protegido por rol, requiere extraccion |
| `EnsureAlquileresTipoEmpresaYLicencias` | [backend/main.go:1318](../../backend/main.go#L1318) | arranque; protegido por rol, requiere extraccion |
| `EnsureNuevasPlantillasTipoEmpresaYLicencias` | [backend/main.go:1324](../../backend/main.go#L1324) | arranque; protegido por rol, requiere extraccion |
| `EnsureEnergiaSolarInTipoEmpresaPreconfiguraciones` | [backend/main.go:1334](../../backend/main.go#L1334) | arranque; protegido por rol, requiere extraccion |
| `EnsurePostgresRuntimeCompat` | [backend/main.go:1348](../../backend/main.go#L1348) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaUsuariosAuthSchema` | [backend/main.go:1351](../../backend/main.go#L1351) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaBuzonSchema` | [backend/main.go:1357](../../backend/main.go#L1357) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCarritosSchema` | [backend/main.go:1361](../../backend/main.go#L1361) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaDatafonosSchema` | [backend/main.go:1365](../../backend/main.go#L1365) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaFinanzasSchema` | [backend/main.go:1381](../../backend/main.go#L1381) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaImpuestosSchema` | [backend/main.go:1385](../../backend/main.go#L1385) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaNominaSchema` | [backend/main.go:1389](../../backend/main.go#L1389) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCreditosSchema` | [backend/main.go:1407](../../backend/main.go#L1407) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaContabilidadColombiaSchema` | [backend/main.go:1411](../../backend/main.go#L1411) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaContabilidadColombiaAvanzadaSchema` | [backend/main.go:1415](../../backend/main.go#L1415) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCentrosCostoSchema` | [backend/main.go:1419](../../backend/main.go#L1419) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCierreFiscalSchema` | [backend/main.go:1423](../../backend/main.go#L1423) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaDeclaracionesTributariasSchema` | [backend/main.go:1427](../../backend/main.go#L1427) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaTesoreriaPresupuestoSchema` | [backend/main.go:1431](../../backend/main.go#L1431) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaImportacionesCosteoSchema` | [backend/main.go:1434](../../backend/main.go#L1434) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaAIUConstruccionSchema` | [backend/main.go:1437](../../backend/main.go#L1437) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCobranzaSchema` | [backend/main.go:1440](../../backend/main.go#L1440) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaPortalContadorSchema` | [backend/main.go:1443](../../backend/main.go#L1443) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaPortalTercerosCertificadosSchema` | [backend/main.go:1446](../../backend/main.go#L1446) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaSoportesComprasIASchema` | [backend/main.go:1449](../../backend/main.go#L1449) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaModulosColombiaSchema` | [backend/main.go:1452](../../backend/main.go#L1452) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaComprasAvanzadasSchema` | [backend/main.go:1455](../../backend/main.go#L1455) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaReservasHotelSchema` | [backend/main.go:1458](../../backend/main.go#L1458) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaTarifasMotelSchema` | [backend/main.go:1462](../../backend/main.go#L1462) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaAIEnterpriseSchema` | [backend/main.go:1466](../../backend/main.go#L1466) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaAIOpenAIProviderSchema` | [backend/main.go:1470](../../backend/main.go#L1470) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaSensorPuertasSchema` | [backend/main.go:1474](../../backend/main.go#L1474) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaControlElectricoSchema` | [backend/main.go:1477](../../backend/main.go#L1477) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaEnergiaSolarSchema` | [backend/main.go:1480](../../backend/main.go#L1480) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCamarasSchema` | [backend/main.go:1483](../../backend/main.go#L1483) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaGrafologiaSchema` | [backend/main.go:1486](../../backend/main.go#L1486) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCarnetsSchema` | [backend/main.go:1489](../../backend/main.go#L1489) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaParqueaderoSchema` | [backend/main.go:1492](../../backend/main.go#L1492) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaApartamentosTuristicosSchema` | [backend/main.go:1495](../../backend/main.go#L1495) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaPropiedadHorizontalSchema` | [backend/main.go:1498](../../backend/main.go#L1498) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaProduccionMRPSchema` | [backend/main.go:1501](../../backend/main.go#L1501) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaWMSSchema` | [backend/main.go:1504](../../backend/main.go#L1504) | arranque; protegido por rol, requiere extraccion |
| `EnsureHotelTarjetasAccesoSchema` | [backend/main.go:1507](../../backend/main.go#L1507) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaProductosSchema` | [backend/main.go:1510](../../backend/main.go#L1510) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaInventarioAvanzadoSchema` | [backend/main.go:1518](../../backend/main.go#L1518) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaCRMVentasAvanzadasSchema` | [backend/main.go:1521](../../backend/main.go#L1521) | arranque; protegido por rol, requiere extraccion |
| `EnsureEmpresaSoporteRemotoSchema` | [backend/main.go:1524](../../backend/main.go#L1524) | arranque; protegido por rol, requiere extraccion |
| `EnsureSensitiveSuperConfigEncrypted` | [backend/main.go:1534](../../backend/main.go#L1534) | arranque; protegido por rol, requiere extraccion |
| `EnsurePostgresPrimaryKeySequences` | [backend/main.go:1538](../../backend/main.go#L1538) | arranque; protegido por rol, requiere extraccion |
| `EnsurePostgresPrimaryKeySequences` | [backend/main.go:1542](../../backend/main.go#L1542) | arranque; protegido por rol, requiere extraccion |
| `EnsureSuperContextoIALogicaNegocio` | [backend/main.go:1548](../../backend/main.go#L1548) | arranque; protegido por rol, requiere extraccion |

## Gate de retiro

1. No agregar nuevas filas: el preflight exige que este inventario coincida con el codigo.
2. Reemplazar primero llamadas en handlers de pagos, facturacion, inventario, archivos y autenticacion por verificadores de esquema o migraciones catalogadas.
3. Cada extraccion debe incluir prueba de base actualizada y de esquema faltante que falle cerrado, sin DDL desde la solicitud.
4. Solo `pcs-migrate` conserva el bootstrap del ledger y las migraciones inmutables.
