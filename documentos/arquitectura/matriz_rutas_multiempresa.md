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
| `/api/empresa/activos_fijos_niif_fiscal` | [backend/main.go:1799](../../backend/main.go#L1799) | `WithEmpresaActivosFijosNIIFPermissions` | protegida |
| `/api/empresa/ai/enterprise` | [backend/handlers/chat_con_inteligencia_artificial_router.go:30](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L30) | `WithEmpresaAIEnterprisePermissions` | protegida |
| `/api/empresa/aiu_construccion` | [backend/main.go:1675](../../backend/main.go#L1675) | `WithEmpresaAIUConstruccionPermissions` | protegida |
| `/api/empresa/alquileres` | [backend/main.go:1699](../../backend/main.go#L1699) | `WithEmpresaAlquileresPermissions` | protegida |
| `/api/empresa/apartamentos_turisticos` | [backend/main.go:1697](../../backend/main.go#L1697) | `WithEmpresaApartamentosTuristicosPermissions` | protegida |
| `/api/empresa/asistencia_empleados` | [backend/main.go:1688](../../backend/main.go#L1688) | `WithEmpresaAsistenciaEmpleadosPermissions` | protegida |
| `/api/empresa/auditoria/eventos` | [backend/main.go:1830](../../backend/main.go#L1830) | `WithEmpresaAuditoriaPermissions` | protegida |
| `/api/empresa/backups` | [backend/main.go:1819](../../backend/main.go#L1819) | `WithEmpresaBackupsPermissions` | protegida |
| `/api/empresa/bancos_pagos` | [backend/main.go:1806](../../backend/main.go#L1806) | `WithEmpresaBancosPagosPermissions` | protegida |
| `/api/empresa/bodegas` | [backend/main.go:1638](../../backend/main.go#L1638) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/bolsa` | [backend/main.go:1772](../../backend/main.go#L1772) | `WithEmpresaBolsaPermissions` | protegida |
| `/api/empresa/buzon` | [backend/main.go:1669](../../backend/main.go#L1669) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/buzon/archivo` | [backend/main.go:1670](../../backend/main.go#L1670) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/calculadora` | [backend/main.go:1813](../../backend/main.go#L1813) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/calidad_procesos` | [backend/main.go:1808](../../backend/main.go#L1808) | `WithEmpresaCalidadProcesosPermissions` | protegida |
| `/api/empresa/camaras` | [backend/main.go:1769](../../backend/main.go#L1769) | `WithEmpresaCamarasPermissions` | protegida |
| `/api/empresa/carnets` | [backend/main.go:1692](../../backend/main.go#L1692) | `WithEmpresaCarnetsPermissions` | protegida |
| `/api/empresa/carritos_compra` | [backend/main.go:1709](../../backend/main.go#L1709) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/carritos_compra/historial_productos` | [backend/main.go:1711](../../backend/main.go#L1711) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/carritos_compra/items` | [backend/main.go:1710](../../backend/main.go#L1710) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/categorias_productos` | [backend/main.go:1639](../../backend/main.go#L1639) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/centros_costo` | [backend/main.go:1800](../../backend/main.go#L1800) | `WithEmpresaCentrosCostoPermissions` | protegida |
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
| `/api/empresa/chat_tareas/archivo` | [backend/main.go:1775](../../backend/main.go#L1775) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/citas` | [backend/main.go:1780](../../backend/main.go#L1780) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/conversaciones` | [backend/main.go:1774](../../backend/main.go#L1774) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/mensajes` | [backend/main.go:1777](../../backend/main.go#L1777) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/mensajes/adjunto` | [backend/main.go:1778](../../backend/main.go#L1778) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/papelera` | [backend/main.go:1782](../../backend/main.go#L1782) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/participantes` | [backend/main.go:1776](../../backend/main.go#L1776) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/tareas` | [backend/main.go:1779](../../backend/main.go#L1779) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/tareas/nota_voz` | [backend/main.go:1781](../../backend/main.go#L1781) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/cierre_fiscal` | [backend/main.go:1801](../../backend/main.go#L1801) | `WithEmpresaCierreFiscalPermissions` | protegida |
| `/api/empresa/clientes` | [backend/main.go:1707](../../backend/main.go#L1707) | `WithEmpresaClientesPermissions` | protegida |
| `/api/empresa/cobranza` | [backend/main.go:1815](../../backend/main.go#L1815) | `WithEmpresaCobranzaPermissions` | protegida |
| `/api/empresa/codigos_de_descuento` | [backend/main.go:1742](../../backend/main.go#L1742) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/comisiones` | [backend/main.go:1744](../../backend/main.go#L1744) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/compras_avanzadas` | [backend/main.go:1664](../../backend/main.go#L1664) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/compras/devoluciones_proveedor` | [backend/handlers/modulos_faltantes.go:639](../../backend/handlers/modulos_faltantes.go#L639) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/compras/documentos` | [backend/main.go:1662](../../backend/main.go#L1662) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/compras/documentos/comprobante` | [backend/main.go:1663](../../backend/main.go#L1663) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/compras/plan_reposicion/actualizar_estado` | [backend/main.go:1661](../../backend/main.go#L1661) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/compras/plan_reposicion/emitir_orden` | [backend/main.go:1660](../../backend/main.go#L1660) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/configuracion_avanzada` | [backend/main.go:1747](../../backend/main.go#L1747) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/configuracion_avanzada/logo` | [backend/main.go:1748](../../backend/main.go#L1748) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/configuracion_general` | [backend/main.go:1745](../../backend/main.go#L1745) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/configuracion_guiada` | [backend/main.go:1749](../../backend/main.go#L1749) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/configuracion_ia_propia` | [backend/main.go:1750](../../backend/main.go#L1750) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/configuracion_operativa` | [backend/main.go:1746](../../backend/main.go#L1746) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/contabilidad_colombia` | [backend/main.go:1797](../../backend/main.go#L1797) | `WithEmpresaContabilidadColombiaPermissions` | protegida |
| `/api/empresa/contabilidad_colombia_avanzada` | [backend/main.go:1798](../../backend/main.go#L1798) | `WithEmpresaContabilidadColombiaAvanzadaPermissions` | protegida |
| `/api/empresa/contratos_obligaciones` | [backend/main.go:1667](../../backend/main.go#L1667) | `WithEmpresaContratosObligacionesPermissions` | protegida |
| `/api/empresa/control_electrico` | [backend/main.go:1844](../../backend/main.go#L1844) | `WithEmpresaControlElectricoPermissions` | protegida |
| `/api/empresa/corte_caja` | [backend/main.go:1789](../../backend/main.go#L1789) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/corte_caja/configuracion` | [backend/main.go:1790](../../backend/main.go#L1790) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/creditos` | [backend/main.go:1814](../../backend/main.go#L1814) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/crm_avanzado` | [backend/main.go:1708](../../backend/main.go#L1708) | `WithEmpresaCRMUnificadoPermissions` | protegida |
| `/api/empresa/crm/campanas` | [backend/handlers/modulos_faltantes.go:644](../../backend/handlers/modulos_faltantes.go#L644) | `WithEmpresaCRMUnificadoPermissions` | protegida |
| `/api/empresa/crm/interacciones` | [backend/handlers/modulos_faltantes.go:643](../../backend/handlers/modulos_faltantes.go#L643) | `WithEmpresaCRMUnificadoPermissions` | protegida |
| `/api/empresa/crm/leads` | [backend/handlers/modulos_faltantes.go:642](../../backend/handlers/modulos_faltantes.go#L642) | `WithEmpresaCRMUnificadoPermissions` | protegida |
| `/api/empresa/cumplimiento_kyc` | [backend/main.go:1807](../../backend/main.go#L1807) | `WithEmpresaCumplimientoKYCPermissions` | protegida |
| `/api/empresa/datafonos` | [backend/main.go:1713](../../backend/main.go#L1713) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/db_admin` | [backend/main.go:1755](../../backend/main.go#L1755) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/declaraciones_tributarias` | [backend/main.go:1802](../../backend/main.go#L1802) | `WithEmpresaDeclaracionesTributariasPermissions` | protegida |
| `/api/empresa/documentos` | [backend/main.go:1820](../../backend/main.go#L1820) | `WithEmpresaDocumentosOnlyOfficePermissions` | protegida |
| `/api/empresa/documentos/firmas` | [backend/handlers/modulos_faltantes.go:655](../../backend/handlers/modulos_faltantes.go#L655) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/documentos/gestion` | [backend/handlers/modulos_faltantes.go:654](../../backend/handlers/modulos_faltantes.go#L654) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/domicilios` | [backend/main.go:1695](../../backend/main.go#L1695) | `WithEmpresaDomiciliosPermissions` | protegida |
| `/api/empresa/drogueria_farmacia` | [backend/main.go:1672](../../backend/main.go#L1672) | `WithEmpresaDrogueriaFarmaciaPermissions` | protegida |
| `/api/empresa/email_corporativo` | [backend/main.go:1754](../../backend/main.go#L1754) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/energia_solar` | [backend/main.go:1768](../../backend/main.go#L1768) | `WithEmpresaEnergiaSolarPermissions` | protegida |
| `/api/empresa/estacion_aseo` | [backend/main.go:1760](../../backend/main.go#L1760) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/estacion_prefs` | [backend/main.go:1759](../../backend/main.go#L1759) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/facturacion_electronica` | [backend/main.go:1761](../../backend/main.go#L1761) | `WithEmpresaFacturacionPermissions` | protegida |
| `/api/empresa/facturacion_electronica/dian` | [backend/handlers/modulos_faltantes.go:660](../../backend/handlers/modulos_faltantes.go#L660) | `WithEmpresaFacturacionPermissions` | protegida |
| `/api/empresa/facturacion_electronica/ecuador` | [backend/main.go:1762](../../backend/main.go#L1762) | `WithEmpresaFacturacionEcuadorPermissions` | protegida |
| `/api/empresa/facturacion_electronica/pais_detectado` | [backend/main.go:1764](../../backend/main.go#L1764) | `WithEmpresaFacturacionPermissions` | protegida |
| `/api/empresa/facturacion_electronica/paises_disponibles` | [backend/main.go:1765](../../backend/main.go#L1765) | `WithEmpresaFacturacionPermissions` | protegida |
| `/api/empresa/facturacion_electronica/panama` | [backend/main.go:1763](../../backend/main.go#L1763) | `WithEmpresaFacturacionPanamaPermissions` | protegida |
| `/api/empresa/finanzas/archivo` | [backend/main.go:1787](../../backend/main.go#L1787) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/asientos_contables` | [backend/main.go:1795](../../backend/main.go#L1795) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/breb_qr` | [backend/main.go:1793](../../backend/main.go#L1793) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/cierres_caja` | [backend/main.go:1796](../../backend/main.go#L1796) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/configuracion` | [backend/main.go:1791](../../backend/main.go#L1791) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/cuentas_cobrar` | [backend/handlers/modulos_faltantes.go:635](../../backend/handlers/modulos_faltantes.go#L635) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/cuentas_pagar` | [backend/handlers/modulos_faltantes.go:636](../../backend/handlers/modulos_faltantes.go#L636) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/movimientos` | [backend/main.go:1786](../../backend/main.go#L1786) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/movimientos/comprobante` | [backend/main.go:1788](../../backend/main.go#L1788) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/periodos` | [backend/main.go:1792](../../backend/main.go#L1792) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/plan_cuentas` | [backend/handlers/modulos_faltantes.go:634](../../backend/handlers/modulos_faltantes.go#L634) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/renta_ia` | [backend/main.go:1794](../../backend/main.go#L1794) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/frecuencia_fp/permitido` | [backend/main.go:1832](../../backend/main.go#L1832) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/gestion_documental` | [backend/main.go:1666](../../backend/main.go#L1666) | `WithEmpresaGestionDocumentalPermissions` | protegida |
| `/api/empresa/gimnasio` | [backend/main.go:1693](../../backend/main.go#L1693) | `WithEmpresaGimnasioPermissions` | protegida |
| `/api/empresa/grafologia` | [backend/main.go:1770](../../backend/main.go#L1770) | `WithEmpresaGrafologiaPermissions` | protegida |
| `/api/empresa/grafologia/archivo` | [backend/main.go:1771](../../backend/main.go#L1771) | `WithEmpresaGrafologiaPermissions` | protegida |
| `/api/empresa/hoja_vida_operativa` | [backend/main.go:1785](../../backend/main.go#L1785) | `WithEmpresaHojaVidaOperativaPermissions` | protegida |
| `/api/empresa/horarios_trabajadores` | [backend/main.go:1686](../../backend/main.go#L1686) | `WithEmpresaHorariosTrabajadoresPermissions` | protegida |
| `/api/empresa/hotel_tarjetas_acceso` | [backend/main.go:1740](../../backend/main.go#L1740) | `WithEmpresaReservasHotelPermissions` | protegida |
| `/api/empresa/ia_empresarial` | [backend/main.go:1773](../../backend/main.go#L1773) | `WithEmpresaReportesPermissions` | protegida |
| `/api/empresa/ia_pedidos_estacion/ejecutar` | [backend/handlers/chat_con_inteligencia_artificial_router.go:25](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L25) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/ia_radio/activar` | [backend/handlers/chat_con_inteligencia_artificial_router.go:26](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L26) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/ia/importar_desde_foto` | [backend/handlers/chat_con_inteligencia_artificial_router.go:24](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L24) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/importaciones_costeo` | [backend/main.go:1674](../../backend/main.go#L1674) | `WithEmpresaImportacionesCosteoPermissions` | protegida |
| `/api/empresa/impresoras` | [backend/main.go:1756](../../backend/main.go#L1756) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/impresoras/agente` | [backend/main.go:1757](../../backend/main.go#L1757) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/impresoras/resolver` | [backend/main.go:1758](../../backend/main.go#L1758) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/impuestos` | [backend/main.go:1766](../../backend/main.go#L1766) | `WithEmpresaFacturacionPermissions` | protegida |
| `/api/empresa/impuestos/agente_internet` | [backend/main.go:1767](../../backend/main.go#L1767) | `WithEmpresaFacturacionPermissions` | protegida |
| `/api/empresa/integraciones/apis` | [backend/handlers/modulos_faltantes.go:657](../../backend/handlers/modulos_faltantes.go#L657) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/integraciones/bancos` | [backend/handlers/modulos_faltantes.go:658](../../backend/handlers/modulos_faltantes.go#L658) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/inventario_avanzado` | [backend/main.go:1658](../../backend/main.go#L1658) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/ajustar` | [backend/main.go:1656](../../backend/main.go#L1656) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/alertas` | [backend/main.go:1645](../../backend/main.go#L1645) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/balance_bodegas` | [backend/main.go:1649](../../backend/main.go#L1649) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/cambiar_producto` | [backend/main.go:1657](../../backend/main.go#L1657) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/configuracion` | [backend/main.go:1644](../../backend/main.go#L1644) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/conteo_ciclico` | [backend/main.go:1646](../../backend/main.go#L1646) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/existencias` | [backend/main.go:1643](../../backend/main.go#L1643) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/lotes_series` | [backend/handlers/modulos_faltantes.go:638](../../backend/handlers/modulos_faltantes.go#L638) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/movimientos` | [backend/main.go:1654](../../backend/main.go#L1654) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/plan_reposicion` | [backend/main.go:1651](../../backend/main.go#L1651) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/plan_reposicion_borrador` | [backend/main.go:1653](../../backend/main.go#L1653) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/plan_reposicion_resumen` | [backend/main.go:1652](../../backend/main.go#L1652) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/proyeccion_quiebre` | [backend/main.go:1650](../../backend/main.go#L1650) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/resumen` | [backend/main.go:1647](../../backend/main.go#L1647) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/tendencia` | [backend/main.go:1648](../../backend/main.go#L1648) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/transferir` | [backend/main.go:1655](../../backend/main.go#L1655) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/licencia_sistema/pdf` | [backend/main.go:1752](../../backend/main.go#L1752) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/licencias/comprobantes` | [backend/main.go:1753](../../backend/main.go#L1753) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/logistica_wms` | [backend/main.go:1677](../../backend/main.go#L1677) | `WithEmpresaWMSPermissions` | protegida |
| `/api/empresa/logistica/envios` | [backend/handlers/modulos_faltantes.go:652](../../backend/handlers/modulos_faltantes.go#L652) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/logistica/rutas` | [backend/handlers/modulos_faltantes.go:651](../../backend/handlers/modulos_faltantes.go#L651) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/logistica/transportistas` | [backend/handlers/modulos_faltantes.go:650](../../backend/handlers/modulos_faltantes.go#L650) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/mantenimiento_programado` | [backend/main.go:1926](../../backend/main.go#L1926) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/mi_horario` | [backend/main.go:1687](../../backend/main.go#L1687) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/nextcloud` | [backend/main.go:1821](../../backend/main.go#L1821) | `WithEmpresaGestionDocumentalPermissions` | protegida |
| `/api/empresa/nomina` | [backend/main.go:1689](../../backend/main.go#L1689) | `WithEmpresaNominaSueldosPermissions` | protegida |
| `/api/empresa/nomina/agente_internet` | [backend/main.go:1690](../../backend/main.go#L1690) | `WithEmpresaNominaSueldosPermissions` | protegida |
| `/api/empresa/noticias` | [backend/main.go:1671](../../backend/main.go#L1671) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/odontologia` | [backend/main.go:1700](../../backend/main.go#L1700) | `WithEmpresaOdontologiaPermissions` | protegida |
| `/api/empresa/offline_ventas` | [backend/main.go:1712](../../backend/main.go#L1712) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/panel_configuracion` | [backend/main.go:1751](../../backend/main.go#L1751) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/parqueadero` | [backend/main.go:1696](../../backend/main.go#L1696) | `WithEmpresaParqueaderoPermissions` | protegida |
| `/api/empresa/permisos_contexto` | [backend/main.go:1846](../../backend/main.go#L1846) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/permisos_empresa` | [backend/main.go:1847](../../backend/main.go#L1847) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/plantillas_integracion/catalogo` | [backend/main.go:1805](../../backend/main.go#L1805) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/plantillas_nuevas/catalogo` | [backend/main.go:1804](../../backend/main.go#L1804) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/portal_contador` | [backend/main.go:1816](../../backend/main.go#L1816) | `WithEmpresaPortalContadorPermissions` | protegida |
| `/api/empresa/portal_terceros_certificados` | [backend/main.go:1817](../../backend/main.go#L1817) | `WithEmpresaPortalTercerosPermissions` | protegida |
| `/api/empresa/produccion_mrp` | [backend/main.go:1676](../../backend/main.go#L1676) | `WithEmpresaProduccionMRPPermissions` | protegida |
| `/api/empresa/produccion/bom` | [backend/handlers/modulos_faltantes.go:646](../../backend/handlers/modulos_faltantes.go#L646) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/produccion/bom_detalle` | [backend/handlers/modulos_faltantes.go:647](../../backend/handlers/modulos_faltantes.go#L647) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/produccion/ordenes` | [backend/handlers/modulos_faltantes.go:648](../../backend/handlers/modulos_faltantes.go#L648) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/productos` | [backend/main.go:1640](../../backend/main.go#L1640) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/productos/imagen` | [backend/main.go:1642](../../backend/main.go#L1642) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/productos/precios_historial` | [backend/main.go:1659](../../backend/main.go#L1659) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/propiedad_horizontal` | [backend/main.go:1698](../../backend/main.go#L1698) | `WithEmpresaPropiedadHorizontalPermissions` | protegida |
| `/api/empresa/propinas` | [backend/main.go:1743](../../backend/main.go#L1743) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/proveedores` | [backend/main.go:1673](../../backend/main.go#L1673) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/publicaciones` | [backend/main.go:1702](../../backend/main.go#L1702) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/publicaciones/` | [backend/main.go:1703](../../backend/main.go#L1703) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/rappi` | [backend/main.go:1715](../../backend/main.go#L1715) | `WithEmpresaVentaPublicaPermissions` | protegida |
| `/api/empresa/recetas_productos` | [backend/main.go:1641](../../backend/main.go#L1641) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/reportes` | [backend/main.go:1828](../../backend/main.go#L1828) | `WithEmpresaReportesPermissions` | protegida |
| `/api/empresa/reportes_ia_chat` | [backend/main.go:1829](../../backend/main.go#L1829) | `WithEmpresaReportesPermissions` | protegida |
| `/api/empresa/reservas_hotel` | [backend/main.go:1736](../../backend/main.go#L1736) | `WithEmpresaReservasHotelPermissions` | protegida |
| `/api/empresa/roles_de_usuario` | [backend/main.go:1845](../../backend/main.go#L1845) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/rrhh/vacaciones_licencias` | [backend/handlers/modulos_faltantes.go:640](../../backend/handlers/modulos_faltantes.go#L640) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/sensor_puertas` | [backend/main.go:1839](../../backend/main.go#L1839) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/sensor_puertas/messages` | [backend/main.go:1843](../../backend/main.go#L1843) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/servicios` | [backend/main.go:1678](../../backend/main.go#L1678) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/soporte_remoto` | [backend/main.go:1827](../../backend/main.go#L1827) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/soportes_compras_ia` | [backend/main.go:1665](../../backend/main.go#L1665) | `WithEmpresaSoportesComprasIAPermissions` | protegida |
| `/api/empresa/tarifas_motel` | [backend/main.go:1739](../../backend/main.go#L1739) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/tarifas_por_dia` | [backend/main.go:1738](../../backend/main.go#L1738) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/tarifas_por_minutos` | [backend/main.go:1737](../../backend/main.go#L1737) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/taxi_system` | [backend/main.go:1694](../../backend/main.go#L1694) | `WithEmpresaTaxiSystemPermissions` | protegida |
| `/api/empresa/tesoreria_presupuesto` | [backend/main.go:1803](../../backend/main.go#L1803) | `WithEmpresaTesoreriaPresupuestoPermissions` | protegida |
| `/api/empresa/tickets_ayuda` | [backend/main.go:1668](../../backend/main.go#L1668) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/turnos_atencion` | [backend/main.go:1701](../../backend/main.go#L1701) | `WithEmpresaTurnosAtencionPermissions` | protegida |
| `/api/empresa/ubicacion_gps/dispositivos` | [backend/main.go:1783](../../backend/main.go#L1783) | `WithEmpresaUbicacionGPSPermissions` | protegida |
| `/api/empresa/ubicacion_gps/recorridos` | [backend/main.go:1784](../../backend/main.go#L1784) | `WithEmpresaUbicacionGPSPermissions` | protegida |
| `/api/empresa/usuarios` | [backend/main.go:1685](../../backend/main.go#L1685) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/usuarios/cambiar_password` | [backend/main.go:1684](../../backend/main.go#L1684) | `WithEmpresaPublicScope` | protegida |
| `/api/empresa/usuarios/establecer_password` | [backend/main.go:1680](../../backend/main.go#L1680) | `WithEmpresaPublicScope` | protegida |
| `/api/empresa/usuarios/login` | [backend/main.go:1679](../../backend/main.go#L1679) | `WithEmpresaPublicScope` | protegida |
| `/api/empresa/usuarios/recuperar_invitacion` | [backend/main.go:1681](../../backend/main.go#L1681) | `WithEmpresaPublicScope` | protegida |
| `/api/empresa/usuarios/restablecer_password` | [backend/main.go:1683](../../backend/main.go#L1683) | `WithEmpresaPublicScope` | protegida |
| `/api/empresa/usuarios/solicitar_recuperacion_password` | [backend/main.go:1682](../../backend/main.go#L1682) | `WithEmpresaPublicScope` | protegida |
| `/api/empresa/vehiculos_registro` | [backend/main.go:1691](../../backend/main.go#L1691) | `WithEmpresaVehiculosRegistroPermissions` | protegida |
| `/api/empresa/venta_publica` | [backend/main.go:1714](../../backend/main.go#L1714) | `WithEmpresaVentaPublicaPermissions` | protegida |
| `/api/empresa/ventas/cotizaciones` | [backend/handlers/modulos_faltantes.go:630](../../backend/handlers/modulos_faltantes.go#L630) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/ventas/devoluciones` | [backend/handlers/modulos_faltantes.go:632](../../backend/handlers/modulos_faltantes.go#L632) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/ventas/pedidos` | [backend/handlers/modulos_faltantes.go:631](../../backend/handlers/modulos_faltantes.go#L631) | `WithEmpresaVentasPermissions` | protegida |

## Gate de cambios

1. Una ruta nueva bajo `/api/empresa/` debe usar un wrapper que cree `TenantContext` despues de validar sesion, pertenencia, rol y permiso.
2. Una fila `requiere revision manual` bloquea declarar cobertura completa hasta documentar su excepcion o corregirla.
3. El handler debe tomar el `empresa_id` desde `TenantContext`; parametros de URL, JSON o cabecera nunca son fuente de autoridad.
4. Los cambios de lectura, escritura, exportacion, descarga, cache o job requieren prueba negativa entre empresa A y empresa B.
