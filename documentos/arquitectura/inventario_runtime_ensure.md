# Inventario de llamadas Ensure en procesos ejecutables

Estado: generado. Actualizar con `node tools/runtime_ensure_inventory.mjs`.

Las llamadas listadas permiten vigilar la autoridad de esquema. En produccion, API y worker deben llegar a verificar esquema versionado, no crear o alterar tablas. Las llamadas de `backend/main.go` estan dentro de `LegacySchemaBootstrap`: `runtimeconfig` solo permite activarlas con rol `migrate` y una decision explicita. El binario `pcs-migrate` es la autoridad dedicada.

## Resumen

- Llamadas inventariadas: 69.
- bootstrap legado; solo autoridad migrate en produccion: 64.
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
| `EnsurePaymentGatewaySchema` | [backend/main.go:1212](../../backend/main.go#L1212) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureLicenciasSchema` | [backend/main.go:1216](../../backend/main.go#L1216) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureLicenciasCatalogoGlobal` | [backend/main.go:1220](../../backend/main.go#L1220) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsurePowerfulSystemEmpresa` | [backend/main.go:1226](../../backend/main.go#L1226) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureSuperAuditoriaSchema` | [backend/main.go:1232](../../backend/main.go#L1232) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureSuperVPSSnapshotSchema` | [backend/main.go:1236](../../backend/main.go#L1236) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureLicenciaVencimientoNotificacionesSchema` | [backend/main.go:1240](../../backend/main.go#L1240) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureLicenciaEmpresaRetencionSchema` | [backend/main.go:1244](../../backend/main.go#L1244) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureUsuarioConfiguracionSchema` | [backend/main.go:1248](../../backend/main.go#L1248) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaEmailCorporativoSchema` | [backend/main.go:1252](../../backend/main.go#L1252) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureCorporateEmailConfigFromEnv` | [backend/main.go:1256](../../backend/main.go#L1256) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureNextcloudConfigFromEnv` | [backend/main.go:1276](../../backend/main.go#L1276) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaNextcloudSchema` | [backend/main.go:1280](../../backend/main.go#L1280) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureAsesorComercialSchema` | [backend/main.go:1294](../../backend/main.go#L1294) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureConstructoraTipoEmpresaYLicencias` | [backend/main.go:1304](../../backend/main.go#L1304) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureAlquileresTipoEmpresaYLicencias` | [backend/main.go:1310](../../backend/main.go#L1310) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEnergiaSolarInTipoEmpresaPreconfiguraciones` | [backend/main.go:1320](../../backend/main.go#L1320) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsurePostgresRuntimeCompat` | [backend/main.go:1334](../../backend/main.go#L1334) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaUsuariosAuthSchema` | [backend/main.go:1337](../../backend/main.go#L1337) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaBuzonSchema` | [backend/main.go:1343](../../backend/main.go#L1343) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCarritosSchema` | [backend/main.go:1347](../../backend/main.go#L1347) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaDatafonosSchema` | [backend/main.go:1351](../../backend/main.go#L1351) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaFinanzasSchema` | [backend/main.go:1367](../../backend/main.go#L1367) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaImpuestosSchema` | [backend/main.go:1371](../../backend/main.go#L1371) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaNominaSchema` | [backend/main.go:1375](../../backend/main.go#L1375) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCreditosSchema` | [backend/main.go:1393](../../backend/main.go#L1393) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaContabilidadColombiaSchema` | [backend/main.go:1397](../../backend/main.go#L1397) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaContabilidadColombiaAvanzadaSchema` | [backend/main.go:1401](../../backend/main.go#L1401) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCentrosCostoSchema` | [backend/main.go:1405](../../backend/main.go#L1405) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCierreFiscalSchema` | [backend/main.go:1409](../../backend/main.go#L1409) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaDeclaracionesTributariasSchema` | [backend/main.go:1413](../../backend/main.go#L1413) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaTesoreriaPresupuestoSchema` | [backend/main.go:1417](../../backend/main.go#L1417) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaImportacionesCosteoSchema` | [backend/main.go:1420](../../backend/main.go#L1420) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaAIUConstruccionSchema` | [backend/main.go:1423](../../backend/main.go#L1423) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCobranzaSchema` | [backend/main.go:1426](../../backend/main.go#L1426) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaPortalContadorSchema` | [backend/main.go:1429](../../backend/main.go#L1429) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaPortalTercerosCertificadosSchema` | [backend/main.go:1432](../../backend/main.go#L1432) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaSoportesComprasIASchema` | [backend/main.go:1435](../../backend/main.go#L1435) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaModulosColombiaSchema` | [backend/main.go:1438](../../backend/main.go#L1438) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaComprasAvanzadasSchema` | [backend/main.go:1441](../../backend/main.go#L1441) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaReservasHotelSchema` | [backend/main.go:1444](../../backend/main.go#L1444) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaTarifasMotelSchema` | [backend/main.go:1448](../../backend/main.go#L1448) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaAIEnterpriseSchema` | [backend/main.go:1452](../../backend/main.go#L1452) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaAIOpenAIProviderSchema` | [backend/main.go:1456](../../backend/main.go#L1456) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaSensorPuertasSchema` | [backend/main.go:1460](../../backend/main.go#L1460) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaControlElectricoSchema` | [backend/main.go:1463](../../backend/main.go#L1463) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaEnergiaSolarSchema` | [backend/main.go:1466](../../backend/main.go#L1466) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCamarasSchema` | [backend/main.go:1469](../../backend/main.go#L1469) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCarnetsSchema` | [backend/main.go:1472](../../backend/main.go#L1472) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaParqueaderoSchema` | [backend/main.go:1475](../../backend/main.go#L1475) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaProduccionMRPSchema` | [backend/main.go:1478](../../backend/main.go#L1478) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaWMSSchema` | [backend/main.go:1481](../../backend/main.go#L1481) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureHotelTarjetasAccesoSchema` | [backend/main.go:1484](../../backend/main.go#L1484) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaProductosSchema` | [backend/main.go:1487](../../backend/main.go#L1487) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaInventarioAvanzadoSchema` | [backend/main.go:1495](../../backend/main.go#L1495) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaCRMVentasAvanzadasSchema` | [backend/main.go:1498](../../backend/main.go#L1498) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureEmpresaSoporteRemotoSchema` | [backend/main.go:1501](../../backend/main.go#L1501) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureSensitiveSuperConfigEncrypted` | [backend/main.go:1511](../../backend/main.go#L1511) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsurePostgresPrimaryKeySequences` | [backend/main.go:1515](../../backend/main.go#L1515) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsurePostgresPrimaryKeySequences` | [backend/main.go:1519](../../backend/main.go#L1519) | bootstrap legado; solo autoridad migrate en produccion |
| `EnsureSuperContextoIALogicaNegocio` | [backend/main.go:1525](../../backend/main.go#L1525) | bootstrap legado; solo autoridad migrate en produccion |

## Gate de retiro

1. No agregar nuevas filas: el preflight exige que este inventario coincida con el codigo.
2. El conteo aceptable de trafico HTTP es cero; cualquier nueva fila HTTP bloquea el preflight.
3. Cada extraccion debe incluir prueba de base actualizada y de esquema faltante que falle cerrado, sin DDL desde la solicitud.
4. Retirar gradualmente el bootstrap legado de `backend/main.go` cuando el catalogo inmutable cubra cada esquema; `pcs-migrate` conserva la autoridad de esquema.
