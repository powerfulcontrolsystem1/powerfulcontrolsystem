# Matriz de rutas multiempresa

Estado: generado. Actualizar con `node tools/tenant_route_inventory.mjs`.

Este inventario detecta registros HTTP bajo `/api/empresa/` y exige que cada uno tenga una evidencia de wrapper autoritativo. Es un control de cobertura, no sustituye las pruebas negativas A/B ni el filtro `empresa_id` en SQL, archivos, cache y jobs.

## Resumen

- Rutas empresariales inventariadas: 204.
- Con wrapper autoritativo detectado: 204.
- Requieren revision manual: 0.
- Duplicados de ruta detectados: 0.

## Registro

| Ruta | Archivo | Wrapper detectado | Estado |
| --- | --- | --- | --- |
| `/api/empresa/activos_fijos_niif_fiscal` | [backend/main.go:1726](../../backend/main.go#L1726) | `WithEmpresaActivosFijosNIIFPermissions` | protegida |
| `/api/empresa/ai/enterprise` | [backend/handlers/chat_con_inteligencia_artificial_router.go:30](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L30) | `WithEmpresaAIEnterprisePermissions` | protegida |
| `/api/empresa/aiu_construccion` | [backend/main.go:1602](../../backend/main.go#L1602) | `WithEmpresaAIUConstruccionPermissions` | protegida |
| `/api/empresa/alquileres` | [backend/main.go:1626](../../backend/main.go#L1626) | `WithEmpresaAlquileresPermissions` | protegida |
| `/api/empresa/apartamentos_turisticos` | [backend/main.go:1624](../../backend/main.go#L1624) | `WithEmpresaApartamentosTuristicosPermissions` | protegida |
| `/api/empresa/asistencia_empleados` | [backend/main.go:1615](../../backend/main.go#L1615) | `WithEmpresaAsistenciaEmpleadosPermissions` | protegida |
| `/api/empresa/auditoria/eventos` | [backend/main.go:1757](../../backend/main.go#L1757) | `WithEmpresaAuditoriaPermissions` | protegida |
| `/api/empresa/backups` | [backend/main.go:1746](../../backend/main.go#L1746) | `WithEmpresaBackupsPermissions` | protegida |
| `/api/empresa/bancos_pagos` | [backend/main.go:1733](../../backend/main.go#L1733) | `WithEmpresaBancosPagosPermissions` | protegida |
| `/api/empresa/bodegas` | [backend/main.go:1565](../../backend/main.go#L1565) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/bolsa` | [backend/main.go:1699](../../backend/main.go#L1699) | `WithEmpresaBolsaPermissions` | protegida |
| `/api/empresa/buzon` | [backend/main.go:1596](../../backend/main.go#L1596) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/buzon/archivo` | [backend/main.go:1597](../../backend/main.go#L1597) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/calculadora` | [backend/main.go:1740](../../backend/main.go#L1740) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/calidad_procesos` | [backend/main.go:1735](../../backend/main.go#L1735) | `WithEmpresaCalidadProcesosPermissions` | protegida |
| `/api/empresa/camaras` | [backend/main.go:1696](../../backend/main.go#L1696) | `WithEmpresaCamarasPermissions` | protegida |
| `/api/empresa/carnets` | [backend/main.go:1619](../../backend/main.go#L1619) | `WithEmpresaCarnetsPermissions` | protegida |
| `/api/empresa/carritos_compra` | [backend/main.go:1636](../../backend/main.go#L1636) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/carritos_compra/historial_productos` | [backend/main.go:1638](../../backend/main.go#L1638) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/carritos_compra/items` | [backend/main.go:1637](../../backend/main.go#L1637) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/categorias_productos` | [backend/main.go:1566](../../backend/main.go#L1566) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/centros_costo` | [backend/main.go:1727](../../backend/main.go#L1727) | `WithEmpresaCentrosCostoPermissions` | protegida |
| `/api/empresa/chat_con_inteligencia_artificial/consultar` | [backend/handlers/chat_con_inteligencia_artificial_router.go:16](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L16) | `WithEmpresaAIEnterprisePermissions` | protegida |
| `/api/empresa/chat_con_inteligencia_artificial/consultar_con_adjunto` | [backend/handlers/chat_con_inteligencia_artificial_router.go:17](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L17) | `WithEmpresaAIEnterprisePermissions` | protegida |
| `/api/empresa/chat_con_inteligencia_artificial/consultar_stream` | [backend/handlers/chat_con_inteligencia_artificial_router.go:18](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L18) | `WithEmpresaAIEnterprisePermissions` | protegida |
| `/api/empresa/chat_con_inteligencia_artificial/historial` | [backend/handlers/chat_con_inteligencia_artificial_router.go:19](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L19) | `WithEmpresaAIEnterprisePermissions` | protegida |
| `/api/empresa/chat_con_inteligencia_artificial/memoria` | [backend/handlers/chat_con_inteligencia_artificial_router.go:20](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L20) | `WithEmpresaAIEnterprisePermissions` | protegida |
| `/api/empresa/chat_con_inteligencia_artificial/modelo_preferido` | [backend/handlers/chat_con_inteligencia_artificial_router.go:15](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L15) | `WithEmpresaAIEnterprisePermissions` | protegida |
| `/api/empresa/chat_con_inteligencia_artificial/modelos` | [backend/handlers/chat_con_inteligencia_artificial_router.go:14](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L14) | `WithEmpresaAIEnterprisePermissions` | protegida |
| `/api/empresa/chat_documentos/compartir_email` | [backend/handlers/chat_con_inteligencia_artificial_router.go:23](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L23) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/chat_documentos/exportar` | [backend/handlers/chat_con_inteligencia_artificial_router.go:22](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L22) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/chat_documentos/generar` | [backend/handlers/chat_con_inteligencia_artificial_router.go:21](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L21) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/chat_tareas/archivo` | [backend/main.go:1702](../../backend/main.go#L1702) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/citas` | [backend/main.go:1707](../../backend/main.go#L1707) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/conversaciones` | [backend/main.go:1701](../../backend/main.go#L1701) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/mensajes` | [backend/main.go:1704](../../backend/main.go#L1704) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/mensajes/adjunto` | [backend/main.go:1705](../../backend/main.go#L1705) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/papelera` | [backend/main.go:1709](../../backend/main.go#L1709) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/participantes` | [backend/main.go:1703](../../backend/main.go#L1703) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/tareas` | [backend/main.go:1706](../../backend/main.go#L1706) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/tareas/nota_voz` | [backend/main.go:1708](../../backend/main.go#L1708) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/cierre_fiscal` | [backend/main.go:1728](../../backend/main.go#L1728) | `WithEmpresaCierreFiscalPermissions` | protegida |
| `/api/empresa/clientes` | [backend/main.go:1634](../../backend/main.go#L1634) | `WithEmpresaClientesPermissions` | protegida |
| `/api/empresa/cobranza` | [backend/main.go:1742](../../backend/main.go#L1742) | `WithEmpresaCobranzaPermissions` | protegida |
| `/api/empresa/codigos_de_descuento` | [backend/main.go:1669](../../backend/main.go#L1669) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/comisiones` | [backend/main.go:1671](../../backend/main.go#L1671) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/compras_avanzadas` | [backend/main.go:1591](../../backend/main.go#L1591) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/compras/devoluciones_proveedor` | [backend/handlers/modulos_faltantes.go:634](../../backend/handlers/modulos_faltantes.go#L634) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/compras/documentos` | [backend/main.go:1589](../../backend/main.go#L1589) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/compras/documentos/comprobante` | [backend/main.go:1590](../../backend/main.go#L1590) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/compras/plan_reposicion/actualizar_estado` | [backend/main.go:1588](../../backend/main.go#L1588) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/compras/plan_reposicion/emitir_orden` | [backend/main.go:1587](../../backend/main.go#L1587) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/configuracion_avanzada` | [backend/main.go:1674](../../backend/main.go#L1674) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/configuracion_avanzada/logo` | [backend/main.go:1675](../../backend/main.go#L1675) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/configuracion_general` | [backend/main.go:1672](../../backend/main.go#L1672) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/configuracion_guiada` | [backend/main.go:1676](../../backend/main.go#L1676) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/configuracion_ia_propia` | [backend/main.go:1677](../../backend/main.go#L1677) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/configuracion_operativa` | [backend/main.go:1673](../../backend/main.go#L1673) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/contabilidad_colombia` | [backend/main.go:1724](../../backend/main.go#L1724) | `WithEmpresaContabilidadColombiaPermissions` | protegida |
| `/api/empresa/contabilidad_colombia_avanzada` | [backend/main.go:1725](../../backend/main.go#L1725) | `WithEmpresaContabilidadColombiaAvanzadaPermissions` | protegida |
| `/api/empresa/contratos_obligaciones` | [backend/main.go:1594](../../backend/main.go#L1594) | `WithEmpresaContratosObligacionesPermissions` | protegida |
| `/api/empresa/control_electrico` | [backend/main.go:1771](../../backend/main.go#L1771) | `WithEmpresaControlElectricoPermissions` | protegida |
| `/api/empresa/corte_caja` | [backend/main.go:1716](../../backend/main.go#L1716) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/corte_caja/configuracion` | [backend/main.go:1717](../../backend/main.go#L1717) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/creditos` | [backend/main.go:1741](../../backend/main.go#L1741) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/crm_avanzado` | [backend/main.go:1635](../../backend/main.go#L1635) | `WithEmpresaCRMUnificadoPermissions` | protegida |
| `/api/empresa/crm/campanas` | [backend/handlers/modulos_faltantes.go:639](../../backend/handlers/modulos_faltantes.go#L639) | `WithEmpresaCRMUnificadoPermissions` | protegida |
| `/api/empresa/crm/interacciones` | [backend/handlers/modulos_faltantes.go:638](../../backend/handlers/modulos_faltantes.go#L638) | `WithEmpresaCRMUnificadoPermissions` | protegida |
| `/api/empresa/crm/leads` | [backend/handlers/modulos_faltantes.go:637](../../backend/handlers/modulos_faltantes.go#L637) | `WithEmpresaCRMUnificadoPermissions` | protegida |
| `/api/empresa/cumplimiento_kyc` | [backend/main.go:1734](../../backend/main.go#L1734) | `WithEmpresaCumplimientoKYCPermissions` | protegida |
| `/api/empresa/datafonos` | [backend/main.go:1640](../../backend/main.go#L1640) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/db_admin` | [backend/main.go:1682](../../backend/main.go#L1682) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/declaraciones_tributarias` | [backend/main.go:1729](../../backend/main.go#L1729) | `WithEmpresaDeclaracionesTributariasPermissions` | protegida |
| `/api/empresa/documentos` | [backend/main.go:1747](../../backend/main.go#L1747) | `WithEmpresaDocumentosOnlyOfficePermissions` | protegida |
| `/api/empresa/documentos/firmas` | [backend/handlers/modulos_faltantes.go:650](../../backend/handlers/modulos_faltantes.go#L650) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/documentos/gestion` | [backend/handlers/modulos_faltantes.go:649](../../backend/handlers/modulos_faltantes.go#L649) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/domicilios` | [backend/main.go:1622](../../backend/main.go#L1622) | `WithEmpresaDomiciliosPermissions` | protegida |
| `/api/empresa/drogueria_farmacia` | [backend/main.go:1599](../../backend/main.go#L1599) | `WithEmpresaDrogueriaFarmaciaPermissions` | protegida |
| `/api/empresa/email_corporativo` | [backend/main.go:1681](../../backend/main.go#L1681) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/energia_solar` | [backend/main.go:1695](../../backend/main.go#L1695) | `WithEmpresaEnergiaSolarPermissions` | protegida |
| `/api/empresa/estacion_aseo` | [backend/main.go:1687](../../backend/main.go#L1687) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/estacion_prefs` | [backend/main.go:1686](../../backend/main.go#L1686) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/facturacion_electronica` | [backend/main.go:1688](../../backend/main.go#L1688) | `WithEmpresaFacturacionPermissions` | protegida |
| `/api/empresa/facturacion_electronica/dian` | [backend/handlers/modulos_faltantes.go:655](../../backend/handlers/modulos_faltantes.go#L655) | `WithEmpresaFacturacionPermissions` | protegida |
| `/api/empresa/facturacion_electronica/ecuador` | [backend/main.go:1689](../../backend/main.go#L1689) | `WithEmpresaFacturacionEcuadorPermissions` | protegida |
| `/api/empresa/facturacion_electronica/pais_detectado` | [backend/main.go:1691](../../backend/main.go#L1691) | `WithEmpresaFacturacionPermissions` | protegida |
| `/api/empresa/facturacion_electronica/paises_disponibles` | [backend/main.go:1692](../../backend/main.go#L1692) | `WithEmpresaFacturacionPermissions` | protegida |
| `/api/empresa/facturacion_electronica/panama` | [backend/main.go:1690](../../backend/main.go#L1690) | `WithEmpresaFacturacionPanamaPermissions` | protegida |
| `/api/empresa/finanzas/archivo` | [backend/main.go:1714](../../backend/main.go#L1714) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/asientos_contables` | [backend/main.go:1722](../../backend/main.go#L1722) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/breb_qr` | [backend/main.go:1720](../../backend/main.go#L1720) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/cierres_caja` | [backend/main.go:1723](../../backend/main.go#L1723) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/configuracion` | [backend/main.go:1718](../../backend/main.go#L1718) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/cuentas_cobrar` | [backend/handlers/modulos_faltantes.go:630](../../backend/handlers/modulos_faltantes.go#L630) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/cuentas_pagar` | [backend/handlers/modulos_faltantes.go:631](../../backend/handlers/modulos_faltantes.go#L631) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/movimientos` | [backend/main.go:1713](../../backend/main.go#L1713) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/movimientos/comprobante` | [backend/main.go:1715](../../backend/main.go#L1715) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/periodos` | [backend/main.go:1719](../../backend/main.go#L1719) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/plan_cuentas` | [backend/handlers/modulos_faltantes.go:629](../../backend/handlers/modulos_faltantes.go#L629) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/renta_ia` | [backend/main.go:1721](../../backend/main.go#L1721) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/frecuencia_fp/permitido` | [backend/main.go:1759](../../backend/main.go#L1759) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/gestion_documental` | [backend/main.go:1593](../../backend/main.go#L1593) | `WithEmpresaGestionDocumentalPermissions` | protegida |
| `/api/empresa/gimnasio` | [backend/main.go:1620](../../backend/main.go#L1620) | `WithEmpresaGimnasioPermissions` | protegida |
| `/api/empresa/grafologia` | [backend/main.go:1697](../../backend/main.go#L1697) | `WithEmpresaGrafologiaPermissions` | protegida |
| `/api/empresa/grafologia/archivo` | [backend/main.go:1698](../../backend/main.go#L1698) | `WithEmpresaGrafologiaPermissions` | protegida |
| `/api/empresa/hoja_vida_operativa` | [backend/main.go:1712](../../backend/main.go#L1712) | `WithEmpresaHojaVidaOperativaPermissions` | protegida |
| `/api/empresa/horarios_trabajadores` | [backend/main.go:1613](../../backend/main.go#L1613) | `WithEmpresaHorariosTrabajadoresPermissions` | protegida |
| `/api/empresa/hotel_tarjetas_acceso` | [backend/main.go:1667](../../backend/main.go#L1667) | `WithEmpresaReservasHotelPermissions` | protegida |
| `/api/empresa/ia_empresarial` | [backend/main.go:1700](../../backend/main.go#L1700) | `WithEmpresaReportesPermissions` | protegida |
| `/api/empresa/ia_pedidos_estacion/ejecutar` | [backend/handlers/chat_con_inteligencia_artificial_router.go:25](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L25) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/ia_radio/activar` | [backend/handlers/chat_con_inteligencia_artificial_router.go:26](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L26) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/ia/importar_desde_foto` | [backend/handlers/chat_con_inteligencia_artificial_router.go:24](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L24) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/importaciones_costeo` | [backend/main.go:1601](../../backend/main.go#L1601) | `WithEmpresaImportacionesCosteoPermissions` | protegida |
| `/api/empresa/impresoras` | [backend/main.go:1683](../../backend/main.go#L1683) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/impresoras/agente` | [backend/main.go:1684](../../backend/main.go#L1684) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/impresoras/resolver` | [backend/main.go:1685](../../backend/main.go#L1685) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/impuestos` | [backend/main.go:1693](../../backend/main.go#L1693) | `WithEmpresaFacturacionPermissions` | protegida |
| `/api/empresa/impuestos/agente_internet` | [backend/main.go:1694](../../backend/main.go#L1694) | `WithEmpresaFacturacionPermissions` | protegida |
| `/api/empresa/integraciones/apis` | [backend/handlers/modulos_faltantes.go:652](../../backend/handlers/modulos_faltantes.go#L652) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/integraciones/bancos` | [backend/handlers/modulos_faltantes.go:653](../../backend/handlers/modulos_faltantes.go#L653) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/inventario_avanzado` | [backend/main.go:1585](../../backend/main.go#L1585) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/ajustar` | [backend/main.go:1583](../../backend/main.go#L1583) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/alertas` | [backend/main.go:1572](../../backend/main.go#L1572) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/balance_bodegas` | [backend/main.go:1576](../../backend/main.go#L1576) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/cambiar_producto` | [backend/main.go:1584](../../backend/main.go#L1584) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/configuracion` | [backend/main.go:1571](../../backend/main.go#L1571) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/conteo_ciclico` | [backend/main.go:1573](../../backend/main.go#L1573) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/existencias` | [backend/main.go:1570](../../backend/main.go#L1570) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/lotes_series` | [backend/handlers/modulos_faltantes.go:633](../../backend/handlers/modulos_faltantes.go#L633) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/movimientos` | [backend/main.go:1581](../../backend/main.go#L1581) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/plan_reposicion` | [backend/main.go:1578](../../backend/main.go#L1578) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/plan_reposicion_borrador` | [backend/main.go:1580](../../backend/main.go#L1580) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/plan_reposicion_resumen` | [backend/main.go:1579](../../backend/main.go#L1579) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/proyeccion_quiebre` | [backend/main.go:1577](../../backend/main.go#L1577) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/resumen` | [backend/main.go:1574](../../backend/main.go#L1574) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/tendencia` | [backend/main.go:1575](../../backend/main.go#L1575) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/transferir` | [backend/main.go:1582](../../backend/main.go#L1582) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/licencia_sistema/pdf` | [backend/main.go:1679](../../backend/main.go#L1679) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/licencias/comprobantes` | [backend/main.go:1680](../../backend/main.go#L1680) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/logistica_wms` | [backend/main.go:1604](../../backend/main.go#L1604) | `WithEmpresaWMSPermissions` | protegida |
| `/api/empresa/logistica/envios` | [backend/handlers/modulos_faltantes.go:647](../../backend/handlers/modulos_faltantes.go#L647) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/logistica/rutas` | [backend/handlers/modulos_faltantes.go:646](../../backend/handlers/modulos_faltantes.go#L646) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/logistica/transportistas` | [backend/handlers/modulos_faltantes.go:645](../../backend/handlers/modulos_faltantes.go#L645) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/mantenimiento_programado` | [backend/main.go:1853](../../backend/main.go#L1853) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/mi_horario` | [backend/main.go:1614](../../backend/main.go#L1614) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/nextcloud` | [backend/main.go:1748](../../backend/main.go#L1748) | `WithEmpresaGestionDocumentalPermissions` | protegida |
| `/api/empresa/nomina` | [backend/main.go:1616](../../backend/main.go#L1616) | `WithEmpresaNominaSueldosPermissions` | protegida |
| `/api/empresa/nomina/agente_internet` | [backend/main.go:1617](../../backend/main.go#L1617) | `WithEmpresaNominaSueldosPermissions` | protegida |
| `/api/empresa/noticias` | [backend/main.go:1598](../../backend/main.go#L1598) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/odontologia` | [backend/main.go:1627](../../backend/main.go#L1627) | `WithEmpresaOdontologiaPermissions` | protegida |
| `/api/empresa/offline_ventas` | [backend/main.go:1639](../../backend/main.go#L1639) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/panel_configuracion` | [backend/main.go:1678](../../backend/main.go#L1678) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/parqueadero` | [backend/main.go:1623](../../backend/main.go#L1623) | `WithEmpresaParqueaderoPermissions` | protegida |
| `/api/empresa/permisos_contexto` | [backend/main.go:1773](../../backend/main.go#L1773) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/permisos_empresa` | [backend/main.go:1774](../../backend/main.go#L1774) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/plantillas_integracion/catalogo` | [backend/main.go:1732](../../backend/main.go#L1732) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/plantillas_nuevas/catalogo` | [backend/main.go:1731](../../backend/main.go#L1731) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/portal_contador` | [backend/main.go:1743](../../backend/main.go#L1743) | `WithEmpresaPortalContadorPermissions` | protegida |
| `/api/empresa/portal_terceros_certificados` | [backend/main.go:1744](../../backend/main.go#L1744) | `WithEmpresaPortalTercerosPermissions` | protegida |
| `/api/empresa/produccion_mrp` | [backend/main.go:1603](../../backend/main.go#L1603) | `WithEmpresaProduccionMRPPermissions` | protegida |
| `/api/empresa/produccion/bom` | [backend/handlers/modulos_faltantes.go:641](../../backend/handlers/modulos_faltantes.go#L641) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/produccion/bom_detalle` | [backend/handlers/modulos_faltantes.go:642](../../backend/handlers/modulos_faltantes.go#L642) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/produccion/ordenes` | [backend/handlers/modulos_faltantes.go:643](../../backend/handlers/modulos_faltantes.go#L643) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/productos` | [backend/main.go:1567](../../backend/main.go#L1567) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/productos/imagen` | [backend/main.go:1569](../../backend/main.go#L1569) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/productos/precios_historial` | [backend/main.go:1586](../../backend/main.go#L1586) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/propiedad_horizontal` | [backend/main.go:1625](../../backend/main.go#L1625) | `WithEmpresaPropiedadHorizontalPermissions` | protegida |
| `/api/empresa/propinas` | [backend/main.go:1670](../../backend/main.go#L1670) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/proveedores` | [backend/main.go:1600](../../backend/main.go#L1600) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/publicaciones` | [backend/main.go:1629](../../backend/main.go#L1629) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/publicaciones/` | [backend/main.go:1630](../../backend/main.go#L1630) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/rappi` | [backend/main.go:1642](../../backend/main.go#L1642) | `WithEmpresaVentaPublicaPermissions` | protegida |
| `/api/empresa/recetas_productos` | [backend/main.go:1568](../../backend/main.go#L1568) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/reportes` | [backend/main.go:1755](../../backend/main.go#L1755) | `WithEmpresaReportesPermissions` | protegida |
| `/api/empresa/reportes_ia_chat` | [backend/main.go:1756](../../backend/main.go#L1756) | `WithEmpresaReportesPermissions` | protegida |
| `/api/empresa/reservas_hotel` | [backend/main.go:1663](../../backend/main.go#L1663) | `WithEmpresaReservasHotelPermissions` | protegida |
| `/api/empresa/roles_de_usuario` | [backend/main.go:1772](../../backend/main.go#L1772) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/rrhh/vacaciones_licencias` | [backend/handlers/modulos_faltantes.go:635](../../backend/handlers/modulos_faltantes.go#L635) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/sensor_puertas` | [backend/main.go:1766](../../backend/main.go#L1766) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/sensor_puertas/messages` | [backend/main.go:1770](../../backend/main.go#L1770) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/servicios` | [backend/main.go:1605](../../backend/main.go#L1605) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/soporte_remoto` | [backend/main.go:1754](../../backend/main.go#L1754) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/soportes_compras_ia` | [backend/main.go:1592](../../backend/main.go#L1592) | `WithEmpresaSoportesComprasIAPermissions` | protegida |
| `/api/empresa/tarifas_motel` | [backend/main.go:1666](../../backend/main.go#L1666) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/tarifas_por_dia` | [backend/main.go:1665](../../backend/main.go#L1665) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/tarifas_por_minutos` | [backend/main.go:1664](../../backend/main.go#L1664) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/taxi_system` | [backend/main.go:1621](../../backend/main.go#L1621) | `WithEmpresaTaxiSystemPermissions` | protegida |
| `/api/empresa/tesoreria_presupuesto` | [backend/main.go:1730](../../backend/main.go#L1730) | `WithEmpresaTesoreriaPresupuestoPermissions` | protegida |
| `/api/empresa/tickets_ayuda` | [backend/main.go:1595](../../backend/main.go#L1595) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/turnos_atencion` | [backend/main.go:1628](../../backend/main.go#L1628) | `WithEmpresaTurnosAtencionPermissions` | protegida |
| `/api/empresa/ubicacion_gps/dispositivos` | [backend/main.go:1710](../../backend/main.go#L1710) | `WithEmpresaUbicacionGPSPermissions` | protegida |
| `/api/empresa/ubicacion_gps/recorridos` | [backend/main.go:1711](../../backend/main.go#L1711) | `WithEmpresaUbicacionGPSPermissions` | protegida |
| `/api/empresa/usuarios` | [backend/main.go:1612](../../backend/main.go#L1612) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/usuarios/cambiar_password` | [backend/main.go:1611](../../backend/main.go#L1611) | `WithEmpresaPublicScope` | protegida |
| `/api/empresa/usuarios/establecer_password` | [backend/main.go:1607](../../backend/main.go#L1607) | `WithEmpresaPublicScope` | protegida |
| `/api/empresa/usuarios/login` | [backend/main.go:1606](../../backend/main.go#L1606) | `WithEmpresaPublicScope` | protegida |
| `/api/empresa/usuarios/recuperar_invitacion` | [backend/main.go:1608](../../backend/main.go#L1608) | `WithEmpresaPublicScope` | protegida |
| `/api/empresa/usuarios/restablecer_password` | [backend/main.go:1610](../../backend/main.go#L1610) | `WithEmpresaPublicScope` | protegida |
| `/api/empresa/usuarios/solicitar_recuperacion_password` | [backend/main.go:1609](../../backend/main.go#L1609) | `WithEmpresaPublicScope` | protegida |
| `/api/empresa/vehiculos_registro` | [backend/main.go:1618](../../backend/main.go#L1618) | `WithEmpresaVehiculosRegistroPermissions` | protegida |
| `/api/empresa/venta_publica` | [backend/main.go:1641](../../backend/main.go#L1641) | `WithEmpresaVentaPublicaPermissions` | protegida |
| `/api/empresa/ventas/cotizaciones` | [backend/handlers/modulos_faltantes.go:625](../../backend/handlers/modulos_faltantes.go#L625) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/ventas/devoluciones` | [backend/handlers/modulos_faltantes.go:627](../../backend/handlers/modulos_faltantes.go#L627) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/ventas/pedidos` | [backend/handlers/modulos_faltantes.go:626](../../backend/handlers/modulos_faltantes.go#L626) | `WithEmpresaVentasPermissions` | protegida |

## Gate de cambios

1. Una ruta nueva bajo `/api/empresa/` debe usar un wrapper que cree `TenantContext` despues de validar sesion, pertenencia, rol y permiso.
2. Una fila `requiere revision manual` bloquea declarar cobertura completa hasta documentar su excepcion o corregirla.
3. El handler debe tomar el `empresa_id` desde `TenantContext`; parametros de URL, JSON o cabecera nunca son fuente de autoridad.
4. Los cambios de lectura, escritura, exportacion, descarga, cache o job requieren prueba negativa entre empresa A y empresa B.
