# Matriz de rutas multiempresa

Estado: generado. Actualizar con `node tools/tenant_route_inventory.mjs`.

Este inventario detecta registros HTTP bajo `/api/empresa/` y exige que cada uno tenga una evidencia de wrapper autoritativo. Es un control de cobertura, no sustituye las pruebas negativas A/B ni el filtro `empresa_id` en SQL, archivos, cache y jobs.

## Resumen

- Rutas empresariales inventariadas: 197.
- Con wrapper autoritativo detectado: 197.
- Requieren revision manual: 0.
- Duplicados de ruta detectados: 0.

## Registro

| Ruta | Archivo | Wrapper detectado | Estado |
| --- | --- | --- | --- |
| `/api/empresa/activos_fijos_niif_fiscal` | [backend/main.go:1779](../../backend/main.go#L1779) | `WithEmpresaActivosFijosNIIFPermissions` | protegida |
| `/api/empresa/ai/enterprise` | [backend/handlers/chat_con_inteligencia_artificial_router.go:30](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L30) | `WithEmpresaAIEnterprisePermissions` | protegida |
| `/api/empresa/aiu_construccion` | [backend/main.go:1663](../../backend/main.go#L1663) | `WithEmpresaAIUConstruccionPermissions` | protegida |
| `/api/empresa/alquileres` | [backend/main.go:1683](../../backend/main.go#L1683) | `WithEmpresaAlquileresPermissions` | protegida |
| `/api/empresa/asistencia_empleados` | [backend/main.go:1676](../../backend/main.go#L1676) | `WithEmpresaAsistenciaEmpleadosPermissions` | protegida |
| `/api/empresa/auditoria/eventos` | [backend/main.go:1810](../../backend/main.go#L1810) | `WithEmpresaAuditoriaPermissions` | protegida |
| `/api/empresa/backups` | [backend/main.go:1799](../../backend/main.go#L1799) | `WithEmpresaBackupsPermissions` | protegida |
| `/api/empresa/bancos_pagos` | [backend/main.go:1786](../../backend/main.go#L1786) | `WithEmpresaBancosPagosPermissions` | protegida |
| `/api/empresa/bodegas` | [backend/main.go:1626](../../backend/main.go#L1626) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/bolsa` | [backend/main.go:1752](../../backend/main.go#L1752) | `WithEmpresaBolsaPermissions` | protegida |
| `/api/empresa/buzon` | [backend/main.go:1657](../../backend/main.go#L1657) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/buzon/archivo` | [backend/main.go:1658](../../backend/main.go#L1658) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/calculadora` | [backend/main.go:1793](../../backend/main.go#L1793) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/calidad_procesos` | [backend/main.go:1788](../../backend/main.go#L1788) | `WithEmpresaCalidadProcesosPermissions` | protegida |
| `/api/empresa/camaras` | [backend/main.go:1751](../../backend/main.go#L1751) | `WithEmpresaCamarasPermissions` | protegida |
| `/api/empresa/carnets` | [backend/main.go:1680](../../backend/main.go#L1680) | `WithEmpresaCarnetsPermissions` | protegida |
| `/api/empresa/carritos_compra` | [backend/main.go:1692](../../backend/main.go#L1692) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/carritos_compra/historial_productos` | [backend/main.go:1694](../../backend/main.go#L1694) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/carritos_compra/items` | [backend/main.go:1693](../../backend/main.go#L1693) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/categorias_productos` | [backend/main.go:1627](../../backend/main.go#L1627) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/centros_costo` | [backend/main.go:1780](../../backend/main.go#L1780) | `WithEmpresaCentrosCostoPermissions` | protegida |
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
| `/api/empresa/chat_tareas/archivo` | [backend/main.go:1755](../../backend/main.go#L1755) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/citas` | [backend/main.go:1760](../../backend/main.go#L1760) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/conversaciones` | [backend/main.go:1754](../../backend/main.go#L1754) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/mensajes` | [backend/main.go:1757](../../backend/main.go#L1757) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/mensajes/adjunto` | [backend/main.go:1758](../../backend/main.go#L1758) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/papelera` | [backend/main.go:1762](../../backend/main.go#L1762) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/participantes` | [backend/main.go:1756](../../backend/main.go#L1756) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/tareas` | [backend/main.go:1759](../../backend/main.go#L1759) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/chat_tareas/tareas/nota_voz` | [backend/main.go:1761](../../backend/main.go#L1761) | `WithEmpresaChatTareasPermissions` | protegida |
| `/api/empresa/cierre_fiscal` | [backend/main.go:1781](../../backend/main.go#L1781) | `WithEmpresaCierreFiscalPermissions` | protegida |
| `/api/empresa/clientes` | [backend/main.go:1690](../../backend/main.go#L1690) | `WithEmpresaClientesPermissions` | protegida |
| `/api/empresa/cobranza` | [backend/main.go:1795](../../backend/main.go#L1795) | `WithEmpresaCobranzaPermissions` | protegida |
| `/api/empresa/codigos_de_descuento` | [backend/main.go:1724](../../backend/main.go#L1724) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/comisiones` | [backend/main.go:1726](../../backend/main.go#L1726) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/compras/devoluciones_proveedor` | [backend/handlers/modulos_faltantes.go:670](../../backend/handlers/modulos_faltantes.go#L670) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/compras/documentos` | [backend/main.go:1650](../../backend/main.go#L1650) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/compras/documentos/comprobante` | [backend/main.go:1651](../../backend/main.go#L1651) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/compras/plan_reposicion/actualizar_estado` | [backend/main.go:1649](../../backend/main.go#L1649) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/compras/plan_reposicion/emitir_orden` | [backend/main.go:1648](../../backend/main.go#L1648) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/compras_avanzadas` | [backend/main.go:1652](../../backend/main.go#L1652) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/configuracion_avanzada` | [backend/main.go:1729](../../backend/main.go#L1729) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/configuracion_avanzada/logo` | [backend/main.go:1730](../../backend/main.go#L1730) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/configuracion_general` | [backend/main.go:1727](../../backend/main.go#L1727) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/configuracion_guiada` | [backend/main.go:1731](../../backend/main.go#L1731) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/configuracion_ia_propia` | [backend/main.go:1732](../../backend/main.go#L1732) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/configuracion_operativa` | [backend/main.go:1728](../../backend/main.go#L1728) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/contabilidad_colombia` | [backend/main.go:1777](../../backend/main.go#L1777) | `WithEmpresaContabilidadColombiaPermissions` | protegida |
| `/api/empresa/contabilidad_colombia_avanzada` | [backend/main.go:1778](../../backend/main.go#L1778) | `WithEmpresaContabilidadColombiaAvanzadaPermissions` | protegida |
| `/api/empresa/contratos_obligaciones` | [backend/main.go:1655](../../backend/main.go#L1655) | `WithEmpresaContratosObligacionesPermissions` | protegida |
| `/api/empresa/control_electrico` | [backend/main.go:1824](../../backend/main.go#L1824) | `WithEmpresaControlElectricoPermissions` | protegida |
| `/api/empresa/corte_caja` | [backend/main.go:1769](../../backend/main.go#L1769) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/corte_caja/configuracion` | [backend/main.go:1770](../../backend/main.go#L1770) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/creditos` | [backend/main.go:1794](../../backend/main.go#L1794) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/crm/campanas` | [backend/handlers/modulos_faltantes.go:675](../../backend/handlers/modulos_faltantes.go#L675) | `WithEmpresaCRMUnificadoPermissions` | protegida |
| `/api/empresa/crm/interacciones` | [backend/handlers/modulos_faltantes.go:674](../../backend/handlers/modulos_faltantes.go#L674) | `WithEmpresaCRMUnificadoPermissions` | protegida |
| `/api/empresa/crm/leads` | [backend/handlers/modulos_faltantes.go:673](../../backend/handlers/modulos_faltantes.go#L673) | `WithEmpresaCRMUnificadoPermissions` | protegida |
| `/api/empresa/crm_avanzado` | [backend/main.go:1691](../../backend/main.go#L1691) | `WithEmpresaCRMUnificadoPermissions` | protegida |
| `/api/empresa/cumplimiento_kyc` | [backend/main.go:1787](../../backend/main.go#L1787) | `WithEmpresaCumplimientoKYCPermissions` | protegida |
| `/api/empresa/datafonos` | [backend/main.go:1696](../../backend/main.go#L1696) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/db_admin` | [backend/main.go:1737](../../backend/main.go#L1737) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/declaraciones_tributarias` | [backend/main.go:1782](../../backend/main.go#L1782) | `WithEmpresaDeclaracionesTributariasPermissions` | protegida |
| `/api/empresa/documentos` | [backend/main.go:1800](../../backend/main.go#L1800) | `WithEmpresaDocumentosOnlyOfficePermissions` | protegida |
| `/api/empresa/documentos/firmas` | [backend/handlers/modulos_faltantes.go:686](../../backend/handlers/modulos_faltantes.go#L686) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/documentos/gestion` | [backend/handlers/modulos_faltantes.go:685](../../backend/handlers/modulos_faltantes.go#L685) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/domicilios` | [backend/main.go:1681](../../backend/main.go#L1681) | `WithEmpresaDomiciliosPermissions` | protegida |
| `/api/empresa/email_corporativo` | [backend/main.go:1736](../../backend/main.go#L1736) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/energia_solar` | [backend/main.go:1750](../../backend/main.go#L1750) | `WithEmpresaEnergiaSolarPermissions` | protegida |
| `/api/empresa/estacion_aseo` | [backend/main.go:1742](../../backend/main.go#L1742) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/estacion_prefs` | [backend/main.go:1741](../../backend/main.go#L1741) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/facturacion_electronica` | [backend/main.go:1743](../../backend/main.go#L1743) | `WithEmpresaFacturacionPermissions` | protegida |
| `/api/empresa/facturacion_electronica/dian` | [backend/handlers/modulos_faltantes.go:691](../../backend/handlers/modulos_faltantes.go#L691) | `WithEmpresaFacturacionPermissions` | protegida |
| `/api/empresa/facturacion_electronica/ecuador` | [backend/main.go:1744](../../backend/main.go#L1744) | `WithEmpresaFacturacionEcuadorPermissions` | protegida |
| `/api/empresa/facturacion_electronica/pais_detectado` | [backend/main.go:1746](../../backend/main.go#L1746) | `WithEmpresaFacturacionPermissions` | protegida |
| `/api/empresa/facturacion_electronica/paises_disponibles` | [backend/main.go:1747](../../backend/main.go#L1747) | `WithEmpresaFacturacionPermissions` | protegida |
| `/api/empresa/facturacion_electronica/panama` | [backend/main.go:1745](../../backend/main.go#L1745) | `WithEmpresaFacturacionPanamaPermissions` | protegida |
| `/api/empresa/finanzas/archivo` | [backend/main.go:1767](../../backend/main.go#L1767) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/asientos_contables` | [backend/main.go:1775](../../backend/main.go#L1775) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/breb_qr` | [backend/main.go:1773](../../backend/main.go#L1773) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/cierres_caja` | [backend/main.go:1776](../../backend/main.go#L1776) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/configuracion` | [backend/main.go:1771](../../backend/main.go#L1771) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/cuentas_cobrar` | [backend/handlers/modulos_faltantes.go:666](../../backend/handlers/modulos_faltantes.go#L666) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/cuentas_pagar` | [backend/handlers/modulos_faltantes.go:667](../../backend/handlers/modulos_faltantes.go#L667) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/movimientos` | [backend/main.go:1766](../../backend/main.go#L1766) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/movimientos/comprobante` | [backend/main.go:1768](../../backend/main.go#L1768) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/periodos` | [backend/main.go:1772](../../backend/main.go#L1772) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/plan_cuentas` | [backend/handlers/modulos_faltantes.go:665](../../backend/handlers/modulos_faltantes.go#L665) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/finanzas/renta_ia` | [backend/main.go:1774](../../backend/main.go#L1774) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/frecuencia_fp/permitido` | [backend/main.go:1812](../../backend/main.go#L1812) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/gestion_documental` | [backend/main.go:1654](../../backend/main.go#L1654) | `WithEmpresaGestionDocumentalPermissions` | protegida |
| `/api/empresa/hoja_vida_operativa` | [backend/main.go:1765](../../backend/main.go#L1765) | `WithEmpresaHojaVidaOperativaPermissions` | protegida |
| `/api/empresa/horarios_trabajadores` | [backend/main.go:1674](../../backend/main.go#L1674) | `WithEmpresaHorariosTrabajadoresPermissions` | protegida |
| `/api/empresa/hotel_tarjetas_acceso` | [backend/main.go:1722](../../backend/main.go#L1722) | `WithEmpresaReservasHotelPermissions` | protegida |
| `/api/empresa/ia/importar_desde_foto` | [backend/handlers/chat_con_inteligencia_artificial_router.go:24](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L24) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/ia_empresarial` | [backend/main.go:1753](../../backend/main.go#L1753) | `WithEmpresaReportesPermissions` | protegida |
| `/api/empresa/ia_pedidos_estacion/ejecutar` | [backend/handlers/chat_con_inteligencia_artificial_router.go:25](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L25) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/ia_radio/activar` | [backend/handlers/chat_con_inteligencia_artificial_router.go:26](../../backend/handlers/chat_con_inteligencia_artificial_router.go#L26) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/importaciones_costeo` | [backend/main.go:1662](../../backend/main.go#L1662) | `WithEmpresaImportacionesCosteoPermissions` | protegida |
| `/api/empresa/impresoras` | [backend/main.go:1738](../../backend/main.go#L1738) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/impresoras/agente` | [backend/main.go:1739](../../backend/main.go#L1739) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/impresoras/resolver` | [backend/main.go:1740](../../backend/main.go#L1740) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/impuestos` | [backend/main.go:1748](../../backend/main.go#L1748) | `WithEmpresaFacturacionPermissions` | protegida |
| `/api/empresa/impuestos/agente_internet` | [backend/main.go:1749](../../backend/main.go#L1749) | `WithEmpresaFacturacionPermissions` | protegida |
| `/api/empresa/integraciones/apis` | [backend/handlers/modulos_faltantes.go:688](../../backend/handlers/modulos_faltantes.go#L688) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/integraciones/bancos` | [backend/handlers/modulos_faltantes.go:689](../../backend/handlers/modulos_faltantes.go#L689) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/inventario/ajustar` | [backend/main.go:1644](../../backend/main.go#L1644) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/alertas` | [backend/main.go:1633](../../backend/main.go#L1633) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/balance_bodegas` | [backend/main.go:1637](../../backend/main.go#L1637) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/cambiar_producto` | [backend/main.go:1645](../../backend/main.go#L1645) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/configuracion` | [backend/main.go:1632](../../backend/main.go#L1632) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/conteo_ciclico` | [backend/main.go:1634](../../backend/main.go#L1634) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/existencias` | [backend/main.go:1631](../../backend/main.go#L1631) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/lotes_series` | [backend/handlers/modulos_faltantes.go:669](../../backend/handlers/modulos_faltantes.go#L669) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/movimientos` | [backend/main.go:1642](../../backend/main.go#L1642) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/plan_reposicion` | [backend/main.go:1639](../../backend/main.go#L1639) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/plan_reposicion_borrador` | [backend/main.go:1641](../../backend/main.go#L1641) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/plan_reposicion_resumen` | [backend/main.go:1640](../../backend/main.go#L1640) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/proyeccion_quiebre` | [backend/main.go:1638](../../backend/main.go#L1638) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/resumen` | [backend/main.go:1635](../../backend/main.go#L1635) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/tendencia` | [backend/main.go:1636](../../backend/main.go#L1636) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario/transferir` | [backend/main.go:1643](../../backend/main.go#L1643) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/inventario_avanzado` | [backend/main.go:1646](../../backend/main.go#L1646) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/licencia_sistema/pdf` | [backend/main.go:1734](../../backend/main.go#L1734) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/licencias/comprobantes` | [backend/main.go:1735](../../backend/main.go#L1735) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/logistica/envios` | [backend/handlers/modulos_faltantes.go:683](../../backend/handlers/modulos_faltantes.go#L683) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/logistica/rutas` | [backend/handlers/modulos_faltantes.go:682](../../backend/handlers/modulos_faltantes.go#L682) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/logistica/transportistas` | [backend/handlers/modulos_faltantes.go:681](../../backend/handlers/modulos_faltantes.go#L681) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/logistica_wms` | [backend/main.go:1665](../../backend/main.go#L1665) | `WithEmpresaWMSPermissions` | protegida |
| `/api/empresa/mantenimiento_programado` | [backend/main.go:1907](../../backend/main.go#L1907) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/mi_horario` | [backend/main.go:1675](../../backend/main.go#L1675) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/nextcloud` | [backend/main.go:1801](../../backend/main.go#L1801) | `WithEmpresaGestionDocumentalPermissions` | protegida |
| `/api/empresa/nomina` | [backend/main.go:1677](../../backend/main.go#L1677) | `WithEmpresaNominaSueldosPermissions` | protegida |
| `/api/empresa/nomina/agente_internet` | [backend/main.go:1678](../../backend/main.go#L1678) | `WithEmpresaNominaSueldosPermissions` | protegida |
| `/api/empresa/noticias` | [backend/main.go:1660](../../backend/main.go#L1660) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/offline_ventas` | [backend/main.go:1695](../../backend/main.go#L1695) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/panel_configuracion` | [backend/main.go:1733](../../backend/main.go#L1733) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/parqueadero` | [backend/main.go:1682](../../backend/main.go#L1682) | `WithEmpresaParqueaderoPermissions` | protegida |
| `/api/empresa/permisos_contexto` | [backend/main.go:1826](../../backend/main.go#L1826) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/permisos_empresa` | [backend/main.go:1827](../../backend/main.go#L1827) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/plantillas_integracion/catalogo` | [backend/main.go:1785](../../backend/main.go#L1785) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/plantillas_nuevas/catalogo` | [backend/main.go:1784](../../backend/main.go#L1784) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/portal_contador` | [backend/main.go:1796](../../backend/main.go#L1796) | `WithEmpresaPortalContadorPermissions` | protegida |
| `/api/empresa/portal_terceros_certificados` | [backend/main.go:1797](../../backend/main.go#L1797) | `WithEmpresaPortalTercerosPermissions` | protegida |
| `/api/empresa/produccion/bom` | [backend/handlers/modulos_faltantes.go:677](../../backend/handlers/modulos_faltantes.go#L677) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/produccion/bom_detalle` | [backend/handlers/modulos_faltantes.go:678](../../backend/handlers/modulos_faltantes.go#L678) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/produccion/ordenes` | [backend/handlers/modulos_faltantes.go:679](../../backend/handlers/modulos_faltantes.go#L679) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/produccion_mrp` | [backend/main.go:1664](../../backend/main.go#L1664) | `WithEmpresaProduccionMRPPermissions` | protegida |
| `/api/empresa/productos` | [backend/main.go:1628](../../backend/main.go#L1628) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/productos/imagen` | [backend/main.go:1630](../../backend/main.go#L1630) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/productos/precios_historial` | [backend/main.go:1647](../../backend/main.go#L1647) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/propinas` | [backend/main.go:1725](../../backend/main.go#L1725) | `WithEmpresaFinanzasPermissions` | protegida |
| `/api/empresa/proveedores` | [backend/main.go:1661](../../backend/main.go#L1661) | `WithEmpresaComprasPermissions` | protegida |
| `/api/empresa/publicaciones` | [backend/main.go:1685](../../backend/main.go#L1685) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/publicaciones/` | [backend/main.go:1686](../../backend/main.go#L1686) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/rappi` | [backend/main.go:1698](../../backend/main.go#L1698) | `WithEmpresaVentaPublicaPermissions` | protegida |
| `/api/empresa/recetas_productos` | [backend/main.go:1629](../../backend/main.go#L1629) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/reportes` | [backend/main.go:1808](../../backend/main.go#L1808) | `WithEmpresaReportesPermissions` | protegida |
| `/api/empresa/reportes_ia_chat` | [backend/main.go:1809](../../backend/main.go#L1809) | `WithEmpresaReportesPermissions` | protegida |
| `/api/empresa/reservas_hotel` | [backend/main.go:1718](../../backend/main.go#L1718) | `WithEmpresaReservasHotelPermissions` | protegida |
| `/api/empresa/roles_de_usuario` | [backend/main.go:1825](../../backend/main.go#L1825) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/rrhh/vacaciones_licencias` | [backend/handlers/modulos_faltantes.go:671](../../backend/handlers/modulos_faltantes.go#L671) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/sensor_puertas` | [backend/main.go:1819](../../backend/main.go#L1819) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/sensor_puertas/messages` | [backend/main.go:1823](../../backend/main.go#L1823) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/servicios` | [backend/main.go:1666](../../backend/main.go#L1666) | `WithEmpresaInventarioPermissions` | protegida |
| `/api/empresa/soporte_remoto` | [backend/main.go:1807](../../backend/main.go#L1807) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/soportes_compras_ia` | [backend/main.go:1653](../../backend/main.go#L1653) | `WithEmpresaSoportesComprasIAPermissions` | protegida |
| `/api/empresa/tarifas_motel` | [backend/main.go:1721](../../backend/main.go#L1721) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/tarifas_por_dia` | [backend/main.go:1720](../../backend/main.go#L1720) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/tarifas_por_minutos` | [backend/main.go:1719](../../backend/main.go#L1719) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/tesoreria_presupuesto` | [backend/main.go:1783](../../backend/main.go#L1783) | `WithEmpresaTesoreriaPresupuestoPermissions` | protegida |
| `/api/empresa/tickets_ayuda` | [backend/main.go:1656](../../backend/main.go#L1656) | `WithEmpresaSelfServicePermissions` | protegida |
| `/api/empresa/turnos_atencion` | [backend/main.go:1684](../../backend/main.go#L1684) | `WithEmpresaTurnosAtencionPermissions` | protegida |
| `/api/empresa/ubicacion_gps/dispositivos` | [backend/main.go:1763](../../backend/main.go#L1763) | `WithEmpresaUbicacionGPSPermissions` | protegida |
| `/api/empresa/ubicacion_gps/recorridos` | [backend/main.go:1764](../../backend/main.go#L1764) | `WithEmpresaUbicacionGPSPermissions` | protegida |
| `/api/empresa/usuarios` | [backend/main.go:1673](../../backend/main.go#L1673) | `WithEmpresaSeguridadPermissions` | protegida |
| `/api/empresa/usuarios/cambiar_password` | [backend/main.go:1672](../../backend/main.go#L1672) | `WithEmpresaPublicScope` | protegida |
| `/api/empresa/usuarios/establecer_password` | [backend/main.go:1668](../../backend/main.go#L1668) | `WithEmpresaPublicScope` | protegida |
| `/api/empresa/usuarios/login` | [backend/main.go:1667](../../backend/main.go#L1667) | `WithEmpresaPublicScope` | protegida |
| `/api/empresa/usuarios/recuperar_invitacion` | [backend/main.go:1669](../../backend/main.go#L1669) | `WithEmpresaPublicScope` | protegida |
| `/api/empresa/usuarios/restablecer_password` | [backend/main.go:1671](../../backend/main.go#L1671) | `WithEmpresaPublicScope` | protegida |
| `/api/empresa/usuarios/solicitar_recuperacion_password` | [backend/main.go:1670](../../backend/main.go#L1670) | `WithEmpresaPublicScope` | protegida |
| `/api/empresa/vehiculos_registro` | [backend/main.go:1679](../../backend/main.go#L1679) | `WithEmpresaVehiculosRegistroPermissions` | protegida |
| `/api/empresa/venta_publica` | [backend/main.go:1697](../../backend/main.go#L1697) | `WithEmpresaVentaPublicaPermissions` | protegida |
| `/api/empresa/ventas/cotizaciones` | [backend/handlers/modulos_faltantes.go:661](../../backend/handlers/modulos_faltantes.go#L661) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/ventas/devoluciones` | [backend/handlers/modulos_faltantes.go:663](../../backend/handlers/modulos_faltantes.go#L663) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/ventas/pedidos` | [backend/handlers/modulos_faltantes.go:662](../../backend/handlers/modulos_faltantes.go#L662) | `WithEmpresaVentasPermissions` | protegida |
| `/api/empresa/vida` | [backend/main.go:1659](../../backend/main.go#L1659) | `WithEmpresaVidaPermissions` | protegida |

## Gate de cambios

1. Una ruta nueva bajo `/api/empresa/` debe usar un wrapper que cree `TenantContext` despues de validar sesion, pertenencia, rol y permiso.
2. Una fila `requiere revision manual` bloquea declarar cobertura completa hasta documentar su excepcion o corregirla.
3. El handler debe tomar el `empresa_id` desde `TenantContext`; parametros de URL, JSON o cabecera nunca son fuente de autoridad.
4. Los cambios de lectura, escritura, exportacion, descarga, cache o job requieren prueba negativa entre empresa A y empresa B.
