# Inventario de llamadas Ensure en procesos ejecutables

Estado: generado. Actualizar con `node tools/runtime_ensure_inventory.mjs`.

Las llamadas listadas permiten vigilar la autoridad de esquema. En produccion, API y worker deben llegar a verificar esquema versionado, no crear o alterar tablas. Las llamadas de `backend/main.go` estan dentro de `LegacySchemaBootstrap`: `runtimeconfig` solo permite activarlas con rol `migrate` y una decision explicita. El binario `pcs-migrate` es la autoridad dedicada.

## Resumen

- Llamadas inventariadas: 70.
- bootstrap legado; solo autoridad migrate en produccion: 65.
- migrador dedicado; autoridad de esquema permitida: 5.

## Registro

| Funcion Ensure | Llamador | Riesgo / prioridad |
| --- | --- | --- |
| `EnsurePostgresRuntimeCompat` | [backend/cmd/pcs-migrate/main.go:69](../../backend/cmd/pcs-migrate/main.go#L69) | migrador dedicado; autoridad de esquema permitida |
| `EnsureRuntimeDatabaseRole` | [backend/cmd/pcs-migrate/main.go:83](../../backend/cmd/pcs-migrate/main.go#L83) | migrador dedicado; autoridad de esquema permitida |
| `EnsureRuntimeDatabaseRole` | [backend/cmd/pcs-migrate/main.go:86](../../backend/cmd/pcs-migrate/main.go#L86) | migrador dedicado; autoridad de esquema permitida |
| `EnsureBackupDatabaseRole` | [backend/cmd/pcs-migrate/main.go:97](../../backend/cmd/pcs-migrate/main.go#L97) | migrador dedicado; autoridad de esquema permitida |
| `EnsureBackupDatabaseRole` | [backend/cmd/pcs-migrate/main.go:100](../../backend/cmd/pcs-migrate/main.go#L100) | migrador dedicado; autoridad de esquema permitida |
| `EnsurePostgresRuntimeCompat` | [backend/main.go:1174](../../backend/main.go#L1174) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsurePostgresRuntimeCompat` | [backend/main.go:1179](../../backend/main.go#L1179) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureAdministradoresAuthSchema` | [backend/main.go:1183](../../backend/main.go#L1183) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsurePaymentGatewaySchema` | [backend/main.go:1222](../../backend/main.go#L1222) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureLicenciasSchema` | [backend/main.go:1226](../../backend/main.go#L1226) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureLicenciasCatalogoGlobal` | [backend/main.go:1230](../../backend/main.go#L1230) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsurePowerfulSystemEmpresa` | [backend/main.go:1236](../../backend/main.go#L1236) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureSuperAuditoriaSchema` | [backend/main.go:1242](../../backend/main.go#L1242) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureSuperVPSSnapshotSchema` | [backend/main.go:1246](../../backend/main.go#L1246) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureLicenciaVencimientoNotificacionesSchema` | [backend/main.go:1250](../../backend/main.go#L1250) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureLicenciaEmpresaRetencionSchema` | [backend/main.go:1254](../../backend/main.go#L1254) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureUsuarioConfiguracionSchema` | [backend/main.go:1258](../../backend/main.go#L1258) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaEmailCorporativoSchema` | [backend/main.go:1262](../../backend/main.go#L1262) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureCorporateEmailConfigFromEnv` | [backend/main.go:1266](../../backend/main.go#L1266) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureNextcloudConfigFromEnv` | [backend/main.go:1286](../../backend/main.go#L1286) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaNextcloudSchema` | [backend/main.go:1290](../../backend/main.go#L1290) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureAsesorComercialSchema` | [backend/main.go:1304](../../backend/main.go#L1304) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureConstructoraTipoEmpresaYLicencias` | [backend/main.go:1314](../../backend/main.go#L1314) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureAlquileresTipoEmpresaYLicencias` | [backend/main.go:1320](../../backend/main.go#L1320) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureNuevasPlantillasTipoEmpresaYLicencias` | [backend/main.go:1326](../../backend/main.go#L1326) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEnergiaSolarInTipoEmpresaPreconfiguraciones` | [backend/main.go:1336](../../backend/main.go#L1336) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsurePostgresRuntimeCompat` | [backend/main.go:1350](../../backend/main.go#L1350) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaUsuariosAuthSchema` | [backend/main.go:1353](../../backend/main.go#L1353) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaBuzonSchema` | [backend/main.go:1359](../../backend/main.go#L1359) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCarritosSchema` | [backend/main.go:1363](../../backend/main.go#L1363) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaDatafonosSchema` | [backend/main.go:1367](../../backend/main.go#L1367) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaFinanzasSchema` | [backend/main.go:1383](../../backend/main.go#L1383) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaImpuestosSchema` | [backend/main.go:1387](../../backend/main.go#L1387) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaNominaSchema` | [backend/main.go:1391](../../backend/main.go#L1391) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCreditosSchema` | [backend/main.go:1409](../../backend/main.go#L1409) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaContabilidadColombiaSchema` | [backend/main.go:1413](../../backend/main.go#L1413) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaContabilidadColombiaAvanzadaSchema` | [backend/main.go:1417](../../backend/main.go#L1417) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCentrosCostoSchema` | [backend/main.go:1421](../../backend/main.go#L1421) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCierreFiscalSchema` | [backend/main.go:1425](../../backend/main.go#L1425) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaDeclaracionesTributariasSchema` | [backend/main.go:1429](../../backend/main.go#L1429) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaTesoreriaPresupuestoSchema` | [backend/main.go:1433](../../backend/main.go#L1433) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaImportacionesCosteoSchema` | [backend/main.go:1436](../../backend/main.go#L1436) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaAIUConstruccionSchema` | [backend/main.go:1439](../../backend/main.go#L1439) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCobranzaSchema` | [backend/main.go:1442](../../backend/main.go#L1442) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaPortalContadorSchema` | [backend/main.go:1445](../../backend/main.go#L1445) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaPortalTercerosCertificadosSchema` | [backend/main.go:1448](../../backend/main.go#L1448) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaSoportesComprasIASchema` | [backend/main.go:1451](../../backend/main.go#L1451) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaModulosColombiaSchema` | [backend/main.go:1454](../../backend/main.go#L1454) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaComprasAvanzadasSchema` | [backend/main.go:1457](../../backend/main.go#L1457) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaReservasHotelSchema` | [backend/main.go:1460](../../backend/main.go#L1460) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaTarifasMotelSchema` | [backend/main.go:1464](../../backend/main.go#L1464) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaAIEnterpriseSchema` | [backend/main.go:1468](../../backend/main.go#L1468) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaAIOpenAIProviderSchema` | [backend/main.go:1472](../../backend/main.go#L1472) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaSensorPuertasSchema` | [backend/main.go:1476](../../backend/main.go#L1476) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaControlElectricoSchema` | [backend/main.go:1479](../../backend/main.go#L1479) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaEnergiaSolarSchema` | [backend/main.go:1482](../../backend/main.go#L1482) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCamarasSchema` | [backend/main.go:1485](../../backend/main.go#L1485) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCarnetsSchema` | [backend/main.go:1488](../../backend/main.go#L1488) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaParqueaderoSchema` | [backend/main.go:1491](../../backend/main.go#L1491) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaProduccionMRPSchema` | [backend/main.go:1494](../../backend/main.go#L1494) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaWMSSchema` | [backend/main.go:1497](../../backend/main.go#L1497) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureHotelTarjetasAccesoSchema` | [backend/main.go:1500](../../backend/main.go#L1500) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaProductosSchema` | [backend/main.go:1503](../../backend/main.go#L1503) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaInventarioAvanzadoSchema` | [backend/main.go:1511](../../backend/main.go#L1511) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCRMVentasAvanzadasSchema` | [backend/main.go:1514](../../backend/main.go#L1514) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaSoporteRemotoSchema` | [backend/main.go:1517](../../backend/main.go#L1517) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureSensitiveSuperConfigEncrypted` | [backend/main.go:1527](../../backend/main.go#L1527) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsurePostgresPrimaryKeySequences` | [backend/main.go:1531](../../backend/main.go#L1531) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsurePostgresPrimaryKeySequences` | [backend/main.go:1535](../../backend/main.go#L1535) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureSuperContextoIALogicaNegocio` | [backend/main.go:1541](../../backend/main.go#L1541) | bootstrap legado; solo autoridad migrate en produccion |

## Gate de retiro

1. No agregar nuevas filas: el preflight exige que este inventario coincida con el codigo.
2. El conteo aceptable de trafico HTTP es cero; cualquier nueva fila HTTP bloquea el preflight.
3. Cada extraccion debe incluir prueba de base actualizada y de esquema faltante que falle cerrado, sin DDL desde la solicitud.
4. Retirar gradualmente el bootstrap legado de `backend/main.go` cuando el catalogo inmutable cubra cada esquema; `pcs-migrate` conserva la autoridad de esquema.
