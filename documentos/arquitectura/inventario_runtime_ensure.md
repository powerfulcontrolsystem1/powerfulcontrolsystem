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
| `EnsurePostgresRuntimeCompat` | [backend/main.go:1147](../../backend/main.go#L1147) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsurePostgresRuntimeCompat` | [backend/main.go:1152](../../backend/main.go#L1152) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureAdministradoresAuthSchema` | [backend/main.go:1156](../../backend/main.go#L1156) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsurePaymentGatewaySchema` | [backend/main.go:1195](../../backend/main.go#L1195) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureLicenciasSchema` | [backend/main.go:1199](../../backend/main.go#L1199) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureLicenciasCatalogoGlobal` | [backend/main.go:1203](../../backend/main.go#L1203) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsurePowerfulSystemEmpresa` | [backend/main.go:1209](../../backend/main.go#L1209) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureSuperAuditoriaSchema` | [backend/main.go:1215](../../backend/main.go#L1215) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureSuperVPSSnapshotSchema` | [backend/main.go:1219](../../backend/main.go#L1219) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureLicenciaVencimientoNotificacionesSchema` | [backend/main.go:1223](../../backend/main.go#L1223) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureLicenciaEmpresaRetencionSchema` | [backend/main.go:1227](../../backend/main.go#L1227) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureUsuarioConfiguracionSchema` | [backend/main.go:1231](../../backend/main.go#L1231) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaEmailCorporativoSchema` | [backend/main.go:1235](../../backend/main.go#L1235) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureCorporateEmailConfigFromEnv` | [backend/main.go:1239](../../backend/main.go#L1239) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureNextcloudConfigFromEnv` | [backend/main.go:1259](../../backend/main.go#L1259) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaNextcloudSchema` | [backend/main.go:1263](../../backend/main.go#L1263) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureAsesorComercialSchema` | [backend/main.go:1277](../../backend/main.go#L1277) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureConstructoraTipoEmpresaYLicencias` | [backend/main.go:1287](../../backend/main.go#L1287) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureDrogueriaFarmaciaTipoEmpresaYLicencias` | [backend/main.go:1293](../../backend/main.go#L1293) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureAlquileresTipoEmpresaYLicencias` | [backend/main.go:1299](../../backend/main.go#L1299) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureNuevasPlantillasTipoEmpresaYLicencias` | [backend/main.go:1305](../../backend/main.go#L1305) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEnergiaSolarInTipoEmpresaPreconfiguraciones` | [backend/main.go:1315](../../backend/main.go#L1315) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsurePostgresRuntimeCompat` | [backend/main.go:1329](../../backend/main.go#L1329) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaUsuariosAuthSchema` | [backend/main.go:1332](../../backend/main.go#L1332) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaBuzonSchema` | [backend/main.go:1338](../../backend/main.go#L1338) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCarritosSchema` | [backend/main.go:1342](../../backend/main.go#L1342) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaDatafonosSchema` | [backend/main.go:1346](../../backend/main.go#L1346) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaFinanzasSchema` | [backend/main.go:1362](../../backend/main.go#L1362) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaImpuestosSchema` | [backend/main.go:1366](../../backend/main.go#L1366) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaNominaSchema` | [backend/main.go:1370](../../backend/main.go#L1370) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCreditosSchema` | [backend/main.go:1388](../../backend/main.go#L1388) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaContabilidadColombiaSchema` | [backend/main.go:1392](../../backend/main.go#L1392) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaContabilidadColombiaAvanzadaSchema` | [backend/main.go:1396](../../backend/main.go#L1396) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCentrosCostoSchema` | [backend/main.go:1400](../../backend/main.go#L1400) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCierreFiscalSchema` | [backend/main.go:1404](../../backend/main.go#L1404) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaDeclaracionesTributariasSchema` | [backend/main.go:1408](../../backend/main.go#L1408) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaTesoreriaPresupuestoSchema` | [backend/main.go:1412](../../backend/main.go#L1412) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaImportacionesCosteoSchema` | [backend/main.go:1415](../../backend/main.go#L1415) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaAIUConstruccionSchema` | [backend/main.go:1418](../../backend/main.go#L1418) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCobranzaSchema` | [backend/main.go:1421](../../backend/main.go#L1421) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaPortalContadorSchema` | [backend/main.go:1424](../../backend/main.go#L1424) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaPortalTercerosCertificadosSchema` | [backend/main.go:1427](../../backend/main.go#L1427) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaSoportesComprasIASchema` | [backend/main.go:1430](../../backend/main.go#L1430) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaModulosColombiaSchema` | [backend/main.go:1433](../../backend/main.go#L1433) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaComprasAvanzadasSchema` | [backend/main.go:1436](../../backend/main.go#L1436) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaReservasHotelSchema` | [backend/main.go:1439](../../backend/main.go#L1439) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaTarifasMotelSchema` | [backend/main.go:1443](../../backend/main.go#L1443) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaAIEnterpriseSchema` | [backend/main.go:1447](../../backend/main.go#L1447) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaAIOpenAIProviderSchema` | [backend/main.go:1451](../../backend/main.go#L1451) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaSensorPuertasSchema` | [backend/main.go:1455](../../backend/main.go#L1455) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaControlElectricoSchema` | [backend/main.go:1458](../../backend/main.go#L1458) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaEnergiaSolarSchema` | [backend/main.go:1461](../../backend/main.go#L1461) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCamarasSchema` | [backend/main.go:1464](../../backend/main.go#L1464) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaGrafologiaSchema` | [backend/main.go:1467](../../backend/main.go#L1467) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCarnetsSchema` | [backend/main.go:1470](../../backend/main.go#L1470) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaParqueaderoSchema` | [backend/main.go:1473](../../backend/main.go#L1473) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaApartamentosTuristicosSchema` | [backend/main.go:1476](../../backend/main.go#L1476) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaPropiedadHorizontalSchema` | [backend/main.go:1479](../../backend/main.go#L1479) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaProduccionMRPSchema` | [backend/main.go:1482](../../backend/main.go#L1482) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaWMSSchema` | [backend/main.go:1485](../../backend/main.go#L1485) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureHotelTarjetasAccesoSchema` | [backend/main.go:1488](../../backend/main.go#L1488) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaProductosSchema` | [backend/main.go:1491](../../backend/main.go#L1491) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaInventarioAvanzadoSchema` | [backend/main.go:1499](../../backend/main.go#L1499) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCRMVentasAvanzadasSchema` | [backend/main.go:1502](../../backend/main.go#L1502) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaSoporteRemotoSchema` | [backend/main.go:1505](../../backend/main.go#L1505) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureSensitiveSuperConfigEncrypted` | [backend/main.go:1515](../../backend/main.go#L1515) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsurePostgresPrimaryKeySequences` | [backend/main.go:1519](../../backend/main.go#L1519) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsurePostgresPrimaryKeySequences` | [backend/main.go:1523](../../backend/main.go#L1523) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureSuperContextoIALogicaNegocio` | [backend/main.go:1529](../../backend/main.go#L1529) | bootstrap legado; solo autoridad migrate en produccion |

## Gate de retiro

1. No agregar nuevas filas: el preflight exige que este inventario coincida con el codigo.
2. El conteo aceptable de trafico HTTP es cero; cualquier nueva fila HTTP bloquea el preflight.
3. Cada extraccion debe incluir prueba de base actualizada y de esquema faltante que falle cerrado, sin DDL desde la solicitud.
4. Retirar gradualmente el bootstrap legado de `backend/main.go` cuando el catalogo inmutable cubra cada esquema; `pcs-migrate` conserva la autoridad de esquema.
