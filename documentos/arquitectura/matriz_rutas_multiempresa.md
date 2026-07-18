# Matriz de rutas multiempresa

Estado: generado. Actualizar con `node tools/tenant_route_inventory.mjs`.

Este inventario detecta registros HTTP bajo `/api/empresa/` y exige que cada uno tenga una evidencia de wrapper autoritativo. Es un control de cobertura, no sustituye las pruebas negativas A/B ni el filtro `empresa_id` en SQL, archivos, cache y jobs.

## Resumen

- Rutas empresariales inventariadas: 203.
- Con wrapper autoritativo detectado: 203.
- Requieren revision manual: 0.
- Duplicados de ruta detectados: 0.

## Registro

| Ruta | Archivo | Wrapper detectado | Estado |
| --- | --- | --- | --- |
| `/api/empresa/activos_fijos_niif_fiscal` | [backend/main.go:1533](../../backend/main.go#L1533) | `WithEmpresaActivosFijosNIIFPermissions` | protegida |
| `/api/empresa/ai/enterprise` | [backend/handlers/chat_con_inteligencia_artificial_router.go:26](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L26) | `WithEmpresaAIEnterprisePermissions` | protegida |
| `/api/empresa/aiu_construccion` | [backend/main.go:1409](../../backend/main.go#L1409) | `WithEmpresaAIUConstruccionPermissions` | protegida |
| `/api/empresa/alquileres` | [backend/main.go:1433](../../backend/main.go#L1433) | `WithEmpresaAlquileresPermissions` | protegida |
| `/api/empresa/apartamentos_turisticos` | [backend/main.go:1431](../../backend/main.go#L1431) | `WithEmpresaApartamentosTuristicosPermissions` | protegida |
| `/api/empresa/asistencia_empleados` | [backend/main.go:1422](../../backend/main.go#L1422) | `WithEmpresaAsistenciaEmpleadosPermissions` | protegida |
| `/api/empresa/auditoria/eventos` | [backend/main.go:1564](../../backend/main.go#L1564) | `WithEmpresaAuditoriaPermissions` | protegida |
| `/api/empresa/backups` | [backend/main.go:1553](../../backend/main.go#L1553) | `WithEmpresaBackupsPermissions` | protegida |
| `/api/empresa/bancos_pagos` | [backend/main.go:1540](../../backend/main.go#L1540) | `WithEmpresaBancosPagosPermissions` | protegida |
| `/api/empresa/bodegas` | [backend/main.go:1372](../../backend/main.go#L1372) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/bolsa` | [backend/main.go:1506](../../backend/main.go#L1506) | `WithEmpresaBolsaPermissions` | protegida |
| `/api/empresa/buzon` | [backend/main.go:1403](../../backend/main.go#L1403) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/buzon/archivo` | [backend/main.go:1404](../../backend/main.go#L1404) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/calculadora` | [backend/main.go:1547](../../backend/main.go#L1547) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/calidad_procesos` | [backend/main.go:1542](../../backend/main.go#L1542) | `WithEmpresaCalidadProcesosPermissions` | protegida |
| `/api/empresa/camaras` | [backend/main.go:1503](../../backend/main.go#L1503) | `WithEmpresaCamarasPermissions` | protegida |
| `/api/empresa/carnets` | [backend/main.go:1426](../../backend/main.go#L1426) | `WithEmpresaCarnetsPermissions` | protegida |
| `/api/empresa/carritos_compra` | [backend/main.go:1443](../../backend/main.go#L1443) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/carritos_compra/historial_productos` | [backend/main.go:1445](../../backend/main.go#L1445) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/carritos_compra/items` | [backend/main.go:1444](../../backend/main.go#L1444) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/categorias_productos` | [backend/main.go:1373](../../backend/main.go#L1373) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/centros_costo` | [backend/main.go:1534](../../backend/main.go#L1534) | `WithEmpresaCentrosCostoPermissions` | protegida |
| `/api/empresa/chat_con_inteligencia_artificial/consultar` | [backend/handlers/chat_con_inteligencia_artificial_router.go:13](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L13) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/chat_con_inteligencia_artificial/consultar_con_adjunto` | [backend/handlers/chat_con_inteligencia_artificial_router.go:14](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L14) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/chat_con_inteligencia_artificial/consultar_stream` | [backend/handlers/chat_con_inteligencia_artificial_router.go:15](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L15) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/chat_con_inteligencia_artificial/historial` | [backend/handlers/chat_con_inteligencia_artificial_router.go:16](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L16) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/chat_con_inteligencia_artificial/modelo_preferido` | [backend/handlers/chat_con_inteligencia_artificial_router.go:12](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L12) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/chat_con_inteligencia_artificial/modelos` | [backend/handlers/chat_con_inteligencia_artificial_router.go:11](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L11) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/chat_documentos/compartir_email` | [backend/handlers/chat_con_inteligencia_artificial_router.go:19](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L19) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/chat_documentos/exportar` | [backend/handlers/chat_con_inteligencia_artificial_router.go:18](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L18) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/chat_documentos/generar` | [backend/handlers/chat_con_inteligencia_artificial_router.go:17](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L17) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/chat_tareas/archivo` | [backend/main.go:1509](../../backend/main.go#L1509) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/citas` | [backend/main.go:1514](../../backend/main.go#L1514) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/conversaciones` | [backend/main.go:1508](../../backend/main.go#L1508) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/mensajes` | [backend/main.go:1511](../../backend/main.go#L1511) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/mensajes/adjunto` | [backend/main.go:1512](../../backend/main.go#L1512) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/papelera` | [backend/main.go:1516](../../backend/main.go#L1516) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/participantes` | [backend/main.go:1510](../../backend/main.go#L1510) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/tareas` | [backend/main.go:1513](../../backend/main.go#L1513) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/tareas/nota_voz` | [backend/main.go:1515](../../backend/main.go#L1515) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/cierre_fiscal` | [backend/main.go:1535](../../backend/main.go#L1535) | `WithEmpresaCierreFiscalPermissions` | protegida |
| `/api/empresa/clientes` | [backend/main.go:1441](../../backend/main.go#L1441) | `WithEmpresaClientesPermissions` | protegida |
| `/api/empresa/cobranza` | [backend/main.go:1549](../../backend/main.go#L1549) | `WithEmpresaCobranzaPermissions` | protegida |
| `/api/empresa/codigos_de_descuento` | [backend/main.go:1476](../../backend/main.go#L1476) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/comisiones` | [backend/main.go:1478](../../backend/main.go#L1478) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/compras_avanzadas` | [backend/main.go:1398](../../backend/main.go#L1398) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/compras/devoluciones_proveedor` | [backend/handlers/modulos_faltantes.go:632](../../backend/handlers/modulos_faltantes.go#L632) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/compras/documentos` | [backend/main.go:1396](../../backend/main.go#L1396) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/compras/documentos/comprobante` | [backend/main.go:1397](../../backend/main.go#L1397) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/compras/plan_reposicion/actualizar_estado` | [backend/main.go:1395](../../backend/main.go#L1395) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/compras/plan_reposicion/emitir_orden` | [backend/main.go:1394](../../backend/main.go#L1394) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/configuracion_avanzada` | [backend/main.go:1481](../../backend/main.go#L1481) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/configuracion_avanzada/logo` | [backend/main.go:1482](../../backend/main.go#L1482) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/configuracion_general` | [backend/main.go:1479](../../backend/main.go#L1479) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/configuracion_guiada` | [backend/main.go:1483](../../backend/main.go#L1483) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/configuracion_ia_propia` | [backend/main.go:1484](../../backend/main.go#L1484) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/configuracion_operativa` | [backend/main.go:1480](../../backend/main.go#L1480) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/contabilidad_colombia` | [backend/main.go:1531](../../backend/main.go#L1531) | `WithEmpresaContabilidadColombiaPermissions` | protegida |
| `/api/empresa/contabilidad_colombia_avanzada` | [backend/main.go:1532](../../backend/main.go#L1532) | `WithEmpresaContabilidadColombiaAvanzadaPermissions` | protegida |
| `/api/empresa/contratos_obligaciones` | [backend/main.go:1401](../../backend/main.go#L1401) | `WithEmpresaContratosObligacionesPermissions` | protegida |
| `/api/empresa/control_electrico` | [backend/main.go:1577](../../backend/main.go#L1577) | `WithEmpresaControlElectricoPermissions` | protegida |
| `/api/empresa/corte_caja` | [backend/main.go:1523](../../backend/main.go#L1523) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/corte_caja/configuracion` | [backend/main.go:1524](../../backend/main.go#L1524) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/creditos` | [backend/main.go:1548](../../backend/main.go#L1548) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/crm_avanzado` | [backend/main.go:1442](../../backend/main.go#L1442) | `WithEmpresaCRMUnificadoPermissions` | protegida |
| `/api/empresa/crm/campanas` | [backend/handlers/modulos_faltantes.go:637](../../backend/handlers/modulos_faltantes.go#L637) | `WithEmpresaCRMUnificadoPermissions` | protegida |
| `/api/empresa/crm/interacciones` | [backend/handlers/modulos_faltantes.go:636](../../backend/handlers/modulos_faltantes.go#L636) | `WithEmpresaCRMUnificadoPermissions` | protegida |
| `/api/empresa/crm/leads` | [backend/handlers/modulos_faltantes.go:635](../../backend/handlers/modulos_faltantes.go#L635) | `WithEmpresaCRMUnificadoPermissions` | protegida |
| `/api/empresa/cumplimiento_kyc` | [backend/main.go:1541](../../backend/main.go#L1541) | `WithEmpresaCumplimientoKYCPermissions` | protegida |
| `/api/empresa/datafonos` | [backend/main.go:1447](../../backend/main.go#L1447) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/db_admin` | [backend/main.go:1489](../../backend/main.go#L1489) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/declaraciones_tributarias` | [backend/main.go:1536](../../backend/main.go#L1536) | `WithEmpresaDeclaracionesTributariasPermissions` | protegida |
| `/api/empresa/documentos` | [backend/main.go:1554](../../backend/main.go#L1554) | `WithEmpresaDocumentosOnlyOfficePermissions` | protegida |
| `/api/empresa/documentos/firmas` | [backend/handlers/modulos_faltantes.go:648](../../backend/handlers/modulos_faltantes.go#L648) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/documentos/gestion` | [backend/handlers/modulos_faltantes.go:647](../../backend/handlers/modulos_faltantes.go#L647) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/domicilios` | [backend/main.go:1429](../../backend/main.go#L1429) | `WithEmpresaDomiciliosPermissions` | protegida |
| `/api/empresa/drogueria_farmacia` | [backend/main.go:1406](../../backend/main.go#L1406) | `WithEmpresaDrogueriaFarmaciaPermissions` | protegida |
| `/api/empresa/email_corporativo` | [backend/main.go:1488](../../backend/main.go#L1488) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/energia_solar` | [backend/main.go:1502](../../backend/main.go#L1502) | `WithEmpresaEnergiaSolarPermissions` | protegida |
| `/api/empresa/estacion_aseo` | [backend/main.go:1494](../../backend/main.go#L1494) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/estacion_prefs` | [backend/main.go:1493](../../backend/main.go#L1493) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/facturacion_electronica` | [backend/main.go:1495](../../backend/main.go#L1495) | `WithEmpresaFacturacionPermissions` | protegida |
| `/api/empresa/facturacion_electronica/dian` | [backend/handlers/modulos_faltantes.go:653](../../backend/handlers/modulos_faltantes.go#L653) | `WithEmpresaFacturacionPermissions` | protegida |
| `/api/empresa/facturacion_electronica/ecuador` | [backend/main.go:1496](../../backend/main.go#L1496) | `WithEmpresaFacturacionEcuadorPermissions` | protegida |
| `/api/empresa/facturacion_electronica/pais_detectado` | [backend/main.go:1498](../../backend/main.go#L1498) | `WithEmpresaFacturacionPermissions` | protegida |
| `/api/empresa/facturacion_electronica/paises_disponibles` | [backend/main.go:1499](../../backend/main.go#L1499) | `WithEmpresaFacturacionPermissions` | protegida |
| `/api/empresa/facturacion_electronica/panama` | [backend/main.go:1497](../../backend/main.go#L1497) | `WithEmpresaFacturacionPanamaPermissions` | protegida |
| `/api/empresa/finanzas/archivo` | [backend/main.go:1521](../../backend/main.go#L1521) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/asientos_contables` | [backend/main.go:1529](../../backend/main.go#L1529) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/breb_qr` | [backend/main.go:1527](../../backend/main.go#L1527) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/cierres_caja` | [backend/main.go:1530](../../backend/main.go#L1530) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/configuracion` | [backend/main.go:1525](../../backend/main.go#L1525) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/cuentas_cobrar` | [backend/handlers/modulos_faltantes.go:628](../../backend/handlers/modulos_faltantes.go#L628) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/cuentas_pagar` | [backend/handlers/modulos_faltantes.go:629](../../backend/handlers/modulos_faltantes.go#L629) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/movimientos` | [backend/main.go:1520](../../backend/main.go#L1520) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/movimientos/comprobante` | [backend/main.go:1522](../../backend/main.go#L1522) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/periodos` | [backend/main.go:1526](../../backend/main.go#L1526) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/plan_cuentas` | [backend/handlers/modulos_faltantes.go:627](../../backend/handlers/modulos_faltantes.go#L627) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/renta_ia` | [backend/main.go:1528](../../backend/main.go#L1528) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/frecuencia_fp/permitido` | [backend/main.go:1566](../../backend/main.go#L1566) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/gestion_documental` | [backend/main.go:1400](../../backend/main.go#L1400) | `WithEmpresaGestionDocumentalPermissions` | protegida |
| `/api/empresa/gimnasio` | [backend/main.go:1427](../../backend/main.go#L1427) | `WithEmpresaGimnasioPermissions` | protegida |
| `/api/empresa/grafologia` | [backend/main.go:1504](../../backend/main.go#L1504) | `WithEmpresaGrafologiaPermissions` | protegida |
| `/api/empresa/grafologia/archivo` | [backend/main.go:1505](../../backend/main.go#L1505) | `WithEmpresaGrafologiaPermissions` | protegida |
| `/api/empresa/hoja_vida_operativa` | [backend/main.go:1519](../../backend/main.go#L1519) | `WithEmpresaHojaVidaOperativaPermissions` | protegida |
| `/api/empresa/horarios_trabajadores` | [backend/main.go:1420](../../backend/main.go#L1420) | `WithEmpresaHorariosTrabajadoresPermissions` | protegida |
| `/api/empresa/hotel_tarjetas_acceso` | [backend/main.go:1474](../../backend/main.go#L1474) | `WithEmpresaReservasHotelPermissions` | protegida |
| `/api/empresa/ia_empresarial` | [backend/main.go:1507](../../backend/main.go#L1507) | `WithEmpresaReportesPermissions` | protegida |
| `/api/empresa/ia_pedidos_estacion/ejecutar` | [backend/handlers/chat_con_inteligencia_artificial_router.go:21](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L21) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/ia_radio/activar` | [backend/handlers/chat_con_inteligencia_artificial_router.go:22](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L22) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/ia/importar_desde_foto` | [backend/handlers/chat_con_inteligencia_artificial_router.go:20](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L20) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/importaciones_costeo` | [backend/main.go:1408](../../backend/main.go#L1408) | `WithEmpresaImportacionesCosteoPermissions` | protegida |
| `/api/empresa/impresoras` | [backend/main.go:1490](../../backend/main.go#L1490) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/impresoras/agente` | [backend/main.go:1491](../../backend/main.go#L1491) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/impresoras/resolver` | [backend/main.go:1492](../../backend/main.go#L1492) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/impuestos` | [backend/main.go:1500](../../backend/main.go#L1500) | `WithEmpresaFacturacionPermissions` | protegida |
| `/api/empresa/impuestos/agente_internet` | [backend/main.go:1501](../../backend/main.go#L1501) | `WithEmpresaFacturacionPermissions` | protegida |
| `/api/empresa/integraciones/apis` | [backend/handlers/modulos_faltantes.go:650](../../backend/handlers/modulos_faltantes.go#L650) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/integraciones/bancos` | [backend/handlers/modulos_faltantes.go:651](../../backend/handlers/modulos_faltantes.go#L651) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/inventario_avanzado` | [backend/main.go:1392](../../backend/main.go#L1392) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/ajustar` | [backend/main.go:1390](../../backend/main.go#L1390) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/alertas` | [backend/main.go:1379](../../backend/main.go#L1379) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/balance_bodegas` | [backend/main.go:1383](../../backend/main.go#L1383) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/cambiar_producto` | [backend/main.go:1391](../../backend/main.go#L1391) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/configuracion` | [backend/main.go:1378](../../backend/main.go#L1378) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/conteo_ciclico` | [backend/main.go:1380](../../backend/main.go#L1380) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/existencias` | [backend/main.go:1377](../../backend/main.go#L1377) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/lotes_series` | [backend/handlers/modulos_faltantes.go:631](../../backend/handlers/modulos_faltantes.go#L631) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/movimientos` | [backend/main.go:1388](../../backend/main.go#L1388) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/plan_reposicion` | [backend/main.go:1385](../../backend/main.go#L1385) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/plan_reposicion_borrador` | [backend/main.go:1387](../../backend/main.go#L1387) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/plan_reposicion_resumen` | [backend/main.go:1386](../../backend/main.go#L1386) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/proyeccion_quiebre` | [backend/main.go:1384](../../backend/main.go#L1384) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/resumen` | [backend/main.go:1381](../../backend/main.go#L1381) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/tendencia` | [backend/main.go:1382](../../backend/main.go#L1382) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/transferir` | [backend/main.go:1389](../../backend/main.go#L1389) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/licencia_sistema/pdf` | [backend/main.go:1486](../../backend/main.go#L1486) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/licencias/comprobantes` | [backend/main.go:1487](../../backend/main.go#L1487) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/logistica_wms` | [backend/main.go:1411](../../backend/main.go#L1411) | `WithEmpresaWMSPermissions` | protegida |
| `/api/empresa/logistica/envios` | [backend/handlers/modulos_faltantes.go:645](../../backend/handlers/modulos_faltantes.go#L645) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/logistica/rutas` | [backend/handlers/modulos_faltantes.go:644](../../backend/handlers/modulos_faltantes.go#L644) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/logistica/transportistas` | [backend/handlers/modulos_faltantes.go:643](../../backend/handlers/modulos_faltantes.go#L643) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/mantenimiento_programado` | [backend/main.go:1659](../../backend/main.go#L1659) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/mi_horario` | [backend/main.go:1421](../../backend/main.go#L1421) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/nextcloud` | [backend/main.go:1555](../../backend/main.go#L1555) | `WithEmpresaGestionDocumentalPermissions` | protegida |
| `/api/empresa/nomina` | [backend/main.go:1423](../../backend/main.go#L1423) | `WithEmpresaNominaSueldosPermissions` | protegida |
| `/api/empresa/nomina/agente_internet` | [backend/main.go:1424](../../backend/main.go#L1424) | `WithEmpresaNominaSueldosPermissions` | protegida |
| `/api/empresa/noticias` | [backend/main.go:1405](../../backend/main.go#L1405) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/odontologia` | [backend/main.go:1434](../../backend/main.go#L1434) | `WithEmpresaOdontologiaPermissions` | protegida |
| `/api/empresa/offline_ventas` | [backend/main.go:1446](../../backend/main.go#L1446) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/panel_configuracion` | [backend/main.go:1485](../../backend/main.go#L1485) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/parqueadero` | [backend/main.go:1430](../../backend/main.go#L1430) | `WithEmpresaParqueaderoPermissions` | protegida |
| `/api/empresa/permisos_contexto` | [backend/main.go:1579](../../backend/main.go#L1579) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/permisos_empresa` | [backend/main.go:1580](../../backend/main.go#L1580) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/plantillas_integracion/catalogo` | [backend/main.go:1539](../../backend/main.go#L1539) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/plantillas_nuevas/catalogo` | [backend/main.go:1538](../../backend/main.go#L1538) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/portal_contador` | [backend/main.go:1550](../../backend/main.go#L1550) | `WithEmpresaPortalContadorPermissions` | protegida |
| `/api/empresa/portal_terceros_certificados` | [backend/main.go:1551](../../backend/main.go#L1551) | `WithEmpresaPortalTercerosPermissions` | protegida |
| `/api/empresa/produccion_mrp` | [backend/main.go:1410](../../backend/main.go#L1410) | `WithEmpresaProduccionMRPPermissions` | protegida |
| `/api/empresa/produccion/bom` | [backend/handlers/modulos_faltantes.go:639](../../backend/handlers/modulos_faltantes.go#L639) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/produccion/bom_detalle` | [backend/handlers/modulos_faltantes.go:640](../../backend/handlers/modulos_faltantes.go#L640) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/produccion/ordenes` | [backend/handlers/modulos_faltantes.go:641](../../backend/handlers/modulos_faltantes.go#L641) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/productos` | [backend/main.go:1374](../../backend/main.go#L1374) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/productos/imagen` | [backend/main.go:1376](../../backend/main.go#L1376) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/productos/precios_historial` | [backend/main.go:1393](../../backend/main.go#L1393) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/propiedad_horizontal` | [backend/main.go:1432](../../backend/main.go#L1432) | `WithEmpresaPropiedadHorizontalPermissions` | protegida |
| `/api/empresa/propinas` | [backend/main.go:1477](../../backend/main.go#L1477) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/proveedores` | [backend/main.go:1407](../../backend/main.go#L1407) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/publicaciones` | [backend/main.go:1436](../../backend/main.go#L1436) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/publicaciones/` | [backend/main.go:1437](../../backend/main.go#L1437) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/rappi` | [backend/main.go:1449](../../backend/main.go#L1449) | `WithEmpresaVentaPublicaPermissions` | protegida |
| `/api/empresa/recetas_productos` | [backend/main.go:1375](../../backend/main.go#L1375) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/reportes` | [backend/main.go:1562](../../backend/main.go#L1562) | `WithEmpresaReportesPermissions` | protegida |
| `/api/empresa/reportes_ia_chat` | [backend/main.go:1563](../../backend/main.go#L1563) | `WithEmpresaReportesPermissions` | protegida |
| `/api/empresa/reservas_hotel` | [backend/main.go:1470](../../backend/main.go#L1470) | `WithEmpresaReservasHotelPermissions` | protegida |
| `/api/empresa/roles_de_usuario` | [backend/main.go:1578](../../backend/main.go#L1578) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/rrhh/vacaciones_licencias` | [backend/handlers/modulos_faltantes.go:633](../../backend/handlers/modulos_faltantes.go#L633) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/sensor_puertas` | [backend/main.go:1573](../../backend/main.go#L1573) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/sensor_puertas/messages` | [backend/main.go:1576](../../backend/main.go#L1576) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/servicios` | [backend/main.go:1412](../../backend/main.go#L1412) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/soporte_remoto` | [backend/main.go:1561](../../backend/main.go#L1561) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/soportes_compras_ia` | [backend/main.go:1399](../../backend/main.go#L1399) | `WithEmpresaSoportesComprasIAPermissions` | protegida |
| `/api/empresa/tarifas_motel` | [backend/main.go:1473](../../backend/main.go#L1473) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/tarifas_por_dia` | [backend/main.go:1472](../../backend/main.go#L1472) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/tarifas_por_minutos` | [backend/main.go:1471](../../backend/main.go#L1471) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/taxi_system` | [backend/main.go:1428](../../backend/main.go#L1428) | `WithEmpresaTaxiSystemPermissions` | protegida |
| `/api/empresa/tesoreria_presupuesto` | [backend/main.go:1537](../../backend/main.go#L1537) | `WithEmpresaTesoreriaPresupuestoPermissions` | protegida |
| `/api/empresa/tickets_ayuda` | [backend/main.go:1402](../../backend/main.go#L1402) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/turnos_atencion` | [backend/main.go:1435](../../backend/main.go#L1435) | `WithEmpresaTurnosAtencionPermissions` | protegida |
| `/api/empresa/ubicacion_gps/dispositivos` | [backend/main.go:1517](../../backend/main.go#L1517) | `WithEmpresaUbicacionGPSPermissions` | protegida |
| `/api/empresa/ubicacion_gps/recorridos` | [backend/main.go:1518](../../backend/main.go#L1518) | `WithEmpresaUbicacionGPSPermissions` | protegida |
| `/api/empresa/usuarios` | [backend/main.go:1419](../../backend/main.go#L1419) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/usuarios/cambiar_password` | [backend/main.go:1418](../../backend/main.go#L1418) | `WithEmpresaPublicScope` | protegida |
| `/api/empresa/usuarios/establecer_password` | [backend/main.go:1414](../../backend/main.go#L1414) | `WithEmpresaPublicScope` | protegida |
| `/api/empresa/usuarios/login` | [backend/main.go:1413](../../backend/main.go#L1413) | `WithEmpresaPublicScope` | protegida |
| `/api/empresa/usuarios/recuperar_invitacion` | [backend/main.go:1415](../../backend/main.go#L1415) | `WithEmpresaPublicScope` | protegida |
| `/api/empresa/usuarios/restablecer_password` | [backend/main.go:1417](../../backend/main.go#L1417) | `WithEmpresaPublicScope` | protegida |
| `/api/empresa/usuarios/solicitar_recuperacion_password` | [backend/main.go:1416](../../backend/main.go#L1416) | `WithEmpresaPublicScope` | protegida |
| `/api/empresa/vehiculos_registro` | [backend/main.go:1425](../../backend/main.go#L1425) | `WithEmpresaVehiculosRegistroPermissions` | protegida |
| `/api/empresa/venta_publica` | [backend/main.go:1448](../../backend/main.go#L1448) | `WithEmpresaVentaPublicaPermissions` | protegida |
| `/api/empresa/ventas/cotizaciones` | [backend/handlers/modulos_faltantes.go:623](../../backend/handlers/modulos_faltantes.go#L623) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/ventas/devoluciones` | [backend/handlers/modulos_faltantes.go:625](../../backend/handlers/modulos_faltantes.go#L625) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/ventas/pedidos` | [backend/handlers/modulos_faltantes.go:624](../../backend/handlers/modulos_faltantes.go#L624) | `WithEmpresaVentasPermissions` | protegida |

## Gate de cambios

1. Una ruta nueva bajo `/api/empresa/` debe usar un wrapper que cree `TenantContext` despues de validar sesion, pertenencia, rol y permiso.
2. Una fila `requiere revision manual` bloquea declarar cobertura completa hasta documentar su excepcion o corregirla.
3. El handler debe tomar el `empresa_id` desde `TenantContext`; parametros de URL, JSON o cabecera nunca son fuente de autoridad.
4. Los cambios de lectura, escritura, exportacion, descarga, cache o job requieren prueba negativa entre empresa A y empresa B.
