# Inventario de llamadas Ensure en procesos ejecutables

Estado: generado. Actualizar con `node tools/runtime_ensure_inventory.mjs`.

Las llamadas listadas permiten vigilar la autoridad de esquema. En produccion, API y worker deben llegar a verificar esquema versionado, no crear o alterar tablas. Las llamadas de `backend/main.go` estan dentro de `LegacySchemaBootstrap`: `runtimeconfig` solo permite activarlas con rol `migrate` y una decision explicita. El binario `pcs-migrate` es la autoridad dedicada.

## Resumen

- Llamadas inventariadas: 72.
- bootstrap legado; solo autoridad migrate en produccion: 69.
- migrador dedicado; autoridad de esquema permitida: 3.

## Registro

| Funcion Ensure | Llamador | Riesgo / prioridad |
| --- | --- | --- |
| `EnsurePostgresRuntimeCompat` | [backend/cmd/pcs-migrate/main.go:69](../../backend/cmd/pcs-migrate/main.go#L69) | migrador dedicado; autoridad de esquema permitida |
| `EnsureRuntimeDatabaseRole` | [backend/cmd/pcs-migrate/main.go:83](../../backend/cmd/pcs-migrate/main.go#L83) | migrador dedicado; autoridad de esquema permitida |
| `EnsureRuntimeDatabaseRole` | [backend/cmd/pcs-migrate/main.go:86](../../backend/cmd/pcs-migrate/main.go#L86) | migrador dedicado; autoridad de esquema permitida |
| `EnsurePostgresRuntimeCompat` | [backend/main.go:1171](../../backend/main.go#L1171) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsurePostgresRuntimeCompat` | [backend/main.go:1176](../../backend/main.go#L1176) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureAdministradoresAuthSchema` | [backend/main.go:1180](../../backend/main.go#L1180) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsurePaymentGatewaySchema` | [backend/main.go:1219](../../backend/main.go#L1219) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureLicenciasSchema` | [backend/main.go:1223](../../backend/main.go#L1223) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureLicenciasCatalogoGlobal` | [backend/main.go:1227](../../backend/main.go#L1227) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsurePowerfulSystemEmpresa` | [backend/main.go:1233](../../backend/main.go#L1233) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureSuperAuditoriaSchema` | [backend/main.go:1239](../../backend/main.go#L1239) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureSuperVPSSnapshotSchema` | [backend/main.go:1243](../../backend/main.go#L1243) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureLicenciaVencimientoNotificacionesSchema` | [backend/main.go:1247](../../backend/main.go#L1247) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureLicenciaEmpresaRetencionSchema` | [backend/main.go:1251](../../backend/main.go#L1251) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureUsuarioConfiguracionSchema` | [backend/main.go:1255](../../backend/main.go#L1255) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaEmailCorporativoSchema` | [backend/main.go:1259](../../backend/main.go#L1259) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureCorporateEmailConfigFromEnv` | [backend/main.go:1263](../../backend/main.go#L1263) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureNextcloudConfigFromEnv` | [backend/main.go:1283](../../backend/main.go#L1283) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaNextcloudSchema` | [backend/main.go:1287](../../backend/main.go#L1287) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureAsesorComercialSchema` | [backend/main.go:1301](../../backend/main.go#L1301) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureConstructoraTipoEmpresaYLicencias` | [backend/main.go:1311](../../backend/main.go#L1311) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureDrogueriaFarmaciaTipoEmpresaYLicencias` | [backend/main.go:1317](../../backend/main.go#L1317) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureAlquileresTipoEmpresaYLicencias` | [backend/main.go:1323](../../backend/main.go#L1323) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureNuevasPlantillasTipoEmpresaYLicencias` | [backend/main.go:1329](../../backend/main.go#L1329) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEnergiaSolarInTipoEmpresaPreconfiguraciones` | [backend/main.go:1339](../../backend/main.go#L1339) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsurePostgresRuntimeCompat` | [backend/main.go:1353](../../backend/main.go#L1353) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaUsuariosAuthSchema` | [backend/main.go:1356](../../backend/main.go#L1356) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaBuzonSchema` | [backend/main.go:1362](../../backend/main.go#L1362) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCarritosSchema` | [backend/main.go:1366](../../backend/main.go#L1366) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaDatafonosSchema` | [backend/main.go:1370](../../backend/main.go#L1370) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaFinanzasSchema` | [backend/main.go:1386](../../backend/main.go#L1386) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaImpuestosSchema` | [backend/main.go:1390](../../backend/main.go#L1390) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaNominaSchema` | [backend/main.go:1394](../../backend/main.go#L1394) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCreditosSchema` | [backend/main.go:1412](../../backend/main.go#L1412) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaContabilidadColombiaSchema` | [backend/main.go:1416](../../backend/main.go#L1416) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaContabilidadColombiaAvanzadaSchema` | [backend/main.go:1420](../../backend/main.go#L1420) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCentrosCostoSchema` | [backend/main.go:1424](../../backend/main.go#L1424) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCierreFiscalSchema` | [backend/main.go:1428](../../backend/main.go#L1428) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaDeclaracionesTributariasSchema` | [backend/main.go:1432](../../backend/main.go#L1432) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaTesoreriaPresupuestoSchema` | [backend/main.go:1436](../../backend/main.go#L1436) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaImportacionesCosteoSchema` | [backend/main.go:1439](../../backend/main.go#L1439) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaAIUConstruccionSchema` | [backend/main.go:1442](../../backend/main.go#L1442) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCobranzaSchema` | [backend/main.go:1445](../../backend/main.go#L1445) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaPortalContadorSchema` | [backend/main.go:1448](../../backend/main.go#L1448) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaPortalTercerosCertificadosSchema` | [backend/main.go:1451](../../backend/main.go#L1451) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaSoportesComprasIASchema` | [backend/main.go:1454](../../backend/main.go#L1454) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaModulosColombiaSchema` | [backend/main.go:1457](../../backend/main.go#L1457) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaComprasAvanzadasSchema` | [backend/main.go:1460](../../backend/main.go#L1460) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaReservasHotelSchema` | [backend/main.go:1463](../../backend/main.go#L1463) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaTarifasMotelSchema` | [backend/main.go:1467](../../backend/main.go#L1467) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaAIEnterpriseSchema` | [backend/main.go:1471](../../backend/main.go#L1471) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaAIOpenAIProviderSchema` | [backend/main.go:1475](../../backend/main.go#L1475) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaSensorPuertasSchema` | [backend/main.go:1479](../../backend/main.go#L1479) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaControlElectricoSchema` | [backend/main.go:1482](../../backend/main.go#L1482) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaEnergiaSolarSchema` | [backend/main.go:1485](../../backend/main.go#L1485) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCamarasSchema` | [backend/main.go:1488](../../backend/main.go#L1488) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaGrafologiaSchema` | [backend/main.go:1491](../../backend/main.go#L1491) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCarnetsSchema` | [backend/main.go:1494](../../backend/main.go#L1494) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaParqueaderoSchema` | [backend/main.go:1497](../../backend/main.go#L1497) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaApartamentosTuristicosSchema` | [backend/main.go:1500](../../backend/main.go#L1500) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaPropiedadHorizontalSchema` | [backend/main.go:1503](../../backend/main.go#L1503) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaProduccionMRPSchema` | [backend/main.go:1506](../../backend/main.go#L1506) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaWMSSchema` | [backend/main.go:1509](../../backend/main.go#L1509) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureHotelTarjetasAccesoSchema` | [backend/main.go:1512](../../backend/main.go#L1512) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaProductosSchema` | [backend/main.go:1515](../../backend/main.go#L1515) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaInventarioAvanzadoSchema` | [backend/main.go:1523](../../backend/main.go#L1523) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCRMVentasAvanzadasSchema` | [backend/main.go:1526](../../backend/main.go#L1526) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaSoporteRemotoSchema` | [backend/main.go:1529](../../backend/main.go#L1529) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureSensitiveSuperConfigEncrypted` | [backend/main.go:1539](../../backend/main.go#L1539) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsurePostgresPrimaryKeySequences` | [backend/main.go:1543](../../backend/main.go#L1543) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsurePostgresPrimaryKeySequences` | [backend/main.go:1547](../../backend/main.go#L1547) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureSuperContextoIALogicaNegocio` | [backend/main.go:1553](../../backend/main.go#L1553) | bootstrap legado; solo autoridad migrate en produccion |

## Gate de retiro

1. No agregar nuevas filas: el preflight exige que este inventario coincida con el codigo.
2. El conteo aceptable de trafico HTTP es cero; cualquier nueva fila HTTP bloquea el preflight.
3. Cada extraccion debe incluir prueba de base actualizada y de esquema faltante que falle cerrado, sin DDL desde la solicitud.
4. Retirar gradualmente el bootstrap legado de `backend/main.go` cuando el catalogo inmutable cubra cada esquema; `pcs-migrate` conserva la autoridad de esquema.
