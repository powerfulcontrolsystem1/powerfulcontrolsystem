# Inventario estático de interfaz - Plan 106

Generado por `node tools/plan106_ui_inventory.mjs`. No editar manualmente.

## Alcance

- Páginas HTML: **302**
- Controles detectados: **5832**
- Acciones a cubrir en E2E: **2877**
- Entradas y selectores: **2955**
- Controles con marcador dinámico: **842**
- Estado: inventario estático previo; la cobertura funcional, visual, por permisos y de IA se registra en el runner E2E y la matriz P106.

## Controles por página

### `web/accept.html` (3)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/checkbox | accion | acceptCheckbox | acceptCheckbox | no |
| 2 | button/- | accion | acceptBtn | Aceptar y continuar | no |
| 3 | a/- | accion | - | Cancelar | no |

### `web/administrar_empresa.html` (49)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | - | Operación y ventas | no |
| 2 | button/button | accion | - | Inventario y compras | no |
| 3 | button/button | accion | - | Producci&oacute;n | no |
| 4 | button/button | accion | - | Finanzas y cumplimiento | no |
| 5 | button/button | accion | - | CRM y clientes | no |
| 6 | button/button | accion | - | Usuarios, clientes y personas | no |
| 7 | button/button | accion | - | Canales digitales y colaboración | no |
| 8 | button/button | accion | - | Control de asistencia y horarios | no |
| 9 | button/button | accion | - | Análisis y control | no |
| 10 | button/button | accion | - | Domótica y Energía Solar | no |
| 11 | button/button | accion | - | Documentos, nube y soporte | no |
| 12 | button/button | accion | - | Soluciones por negocio | no |
| 13 | button/button | accion | - | Administración | no |
| 14 | button/button | accion | - | Licencia | no |
| 15 | button/button | accion | - | &#9776; Ocultar menú | sí |
| 16 | button/button | accion | adminNotificationBell | &#128276; 0 | no |
| 17 | button/button | accion | adminNotificationRefresh | Actualizar | no |
| 18 | button/button | accion | adminFavoriteBtn | &#9733; | no |
| 19 | button/button | accion | openAIDrawer | Asistente IA | sí |
| 20 | button/button | accion | openRadioDrawer | Música latina | no |
| 21 | input/checkbox | accion | radioFloatingEnabled | radioFloatingEnabled | no |
| 22 | button/button | accion | closeRadioDrawer | &times; | no |
| 23 | select/- | entrada | radioCountrySelect | Detectar automaticamente Panamá Ecuador | no |
| 24 | input/text | entrada | radioCustomName | Nombre de emisora | no |
| 25 | input/text | entrada | radioCustomGenre | Genero de emisora | no |
| 26 | input/url | entrada | radioCustomStream | URL de streaming | no |
| 27 | input/url | entrada | radioCustomSource | Sitio web de la emisora | no |
| 28 | select/- | entrada | radioCustomCountry | Personalizada Panamá Ecuador | no |
| 29 | button/submit | accion | - | Agregar | no |
| 30 | button/button | accion | closeRadioDrawerBottom | Cerrar reproductor | no |
| 31 | button/button | accion | radioMiniClose | &times; | no |
| 32 | button/button | accion | radioMiniPlayPause | Pausar | no |
| 33 | input/range | entrada | radioMiniVolume | 0.7 | no |
| 34 | button/button | accion | aiChatMinibarExpand | Abrir asistente IA | no |
| 35 | button/button | accion | aiChatHintToggle | Ver ejemplos | no |
| 36 | button/button | accion | aiChatConfigBtn | Configurar chat flotante | no |
| 37 | button/button | accion | aiChatMinimize | Minimizar chat | no |
| 38 | button/button | accion | closeAIDrawer | &times; | no |
| 39 | button/button | accion | aiChatNewBtn | Nuevo chat | no |
| 40 | button/button | accion | aiChatConvBtn | Modo conversación | no |
| 41 | button/button | accion | aiChatMicBtn | Dictar mensaje | no |
| 42 | button/button | accion | aiChatVoiceBtn | Voz del asistente | no |
| 43 | button/button | accion | aiChatStopBtn | Detener audio y respuesta | no |
| 44 | input/hidden | entrada | aiChatMode | operativo | no |
| 45 | input/file | accion | aiChatAttachment | Adjuntar archivo para IA | no |
| 46 | button/button | accion | aiChatAttachBtn | Adjuntar archivo | no |
| 47 | button/button | accion | aiChatClearAttachment | &times; | no |
| 48 | textarea/- | entrada | aiChatInput | Mensaje al asistente IA | no |
| 49 | button/submit | accion | - | Enviar | no |

### `web/administrar_empresa/activos_fijos_niif_fiscal.html` (49)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/month | entrada | periodo | Periodo | no |
| 2 | button/button | accion | btnRefresh | Actualizar | no |
| 3 | button/button | accion | btnSeed | Cargar demo | no |
| 4 | button/button | accion | btnExport | Exportar CSV | no |
| 5 | button/button | accion | - | Dashboard | sí |
| 6 | button/button | accion | - | Libro | sí |
| 7 | button/button | accion | - | Registrar | sí |
| 8 | button/button | accion | - | Depreciación | sí |
| 9 | button/button | accion | - | Eventos | sí |
| 10 | input/- | entrada | codigo | codigo | no |
| 11 | input/- | entrada | nombre | nombre | no |
| 12 | select/- | entrada | categoria | Equipo de cómputo Muebles y enseres Maquinaria Vehículo Edificación Intangible Mejora en propiedad ajena | no |
| 13 | input/- | entrada | serial | serial | no |
| 14 | input/- | entrada | placa | placa | no |
| 15 | input/date | entrada | fechaCompra | fechaCompra | no |
| 16 | input/number | entrada | costo | costo | no |
| 17 | input/number | entrada | residual | residual | no |
| 18 | input/number | entrada | deterioro | deterioro | no |
| 19 | input/number | entrada | vidaNiif | 60 | no |
| 20 | select/- | entrada | metodoNiif | Línea recta Doble saldo | no |
| 21 | input/date | entrada | inicioDep | inicioDep | no |
| 22 | input/number | entrada | baseFiscal | baseFiscal | no |
| 23 | input/number | entrada | vidaFiscal | 60 | no |
| 24 | select/- | entrada | metodoFiscal | Línea recta Doble saldo | no |
| 25 | input/- | entrada | ctaActivo | ctaActivo | no |
| 26 | input/- | entrada | ctaDep | ctaDep | no |
| 27 | input/- | entrada | ctaGasto | ctaGasto | no |
| 28 | input/- | entrada | ctaDeterioro | ctaDeterioro | no |
| 29 | input/number | entrada | valorRazonable | valorRazonable | no |
| 30 | input/date | entrada | fechaValoracion | fechaValoracion | no |
| 31 | input/- | entrada | ubicacion | ubicacion | no |
| 32 | input/- | entrada | responsable | responsable | no |
| 33 | input/- | entrada | centroCosto | centroCosto | no |
| 34 | input/- | entrada | proveedor | proveedor | no |
| 35 | input/number | entrada | valorAsegurado | valorAsegurado | no |
| 36 | input/- | entrada | poliza | poliza | no |
| 37 | input/number | entrada | mantDias | 0 | no |
| 38 | input/date | entrada | ultimoMant | ultimoMant | no |
| 39 | button/submit | accion | - | Guardar activo | no |
| 40 | input/month | entrada | depPeriodo | depPeriodo | no |
| 41 | button/submit | accion | - | Generar periodo | no |
| 42 | select/- | entrada | eventoActivo | eventoActivo | no |
| 43 | select/- | entrada | eventoTipo | Traslado Mantenimiento Ajuste valor Baja Venta Retiro | no |
| 44 | input/date | entrada | eventoFecha | eventoFecha | no |
| 45 | input/number | entrada | eventoValor | 0 | no |
| 46 | input/- | entrada | eventoUbicacion | eventoUbicacion | no |
| 47 | input/- | entrada | eventoResponsable | eventoResponsable | no |
| 48 | textarea/- | entrada | eventoDetalle | eventoDetalle | no |
| 49 | button/submit | accion | - | Registrar evento | no |

### `web/administrar_empresa/administrar_clientes.html` (29)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/- | accion | addBtn | Agregar | no |
| 2 | input/hidden | entrada | itemId | itemId | no |
| 3 | input/hidden | entrada | empresa_id | empresa_id | no |
| 4 | select/- | entrada | tipo_documento | NIT CC CE Pasaporte TI Otro | no |
| 5 | input/- | entrada | numero_documento | numero_documento | no |
| 6 | input/- | entrada | digito_verificacion | digito_verificacion | no |
| 7 | select/- | entrada | tipo_persona | Jurídica Natural | no |
| 8 | input/- | entrada | nombre_razon_social | nombre_razon_social | no |
| 9 | input/- | entrada | nombre_comercial | nombre_comercial | no |
| 10 | input/- | entrada | regimen_fiscal | regimen_fiscal | no |
| 11 | input/- | entrada | responsabilidad_tributaria | responsabilidad_tributaria | no |
| 12 | input/email | entrada | email | email | no |
| 13 | input/- | entrada | telefono | telefono | no |
| 14 | input/- | entrada | direccion | direccion | no |
| 15 | input/- | entrada | pais | CO | no |
| 16 | input/- | entrada | departamento | departamento | no |
| 17 | input/- | entrada | municipio | municipio | no |
| 18 | input/- | entrada | codigo_postal | codigo_postal | no |
| 19 | textarea/- | entrada | observaciones | observaciones | no |
| 20 | button/submit | accion | saveBtn | Guardar | no |
| 21 | button/button | accion | cancelBtn | Cancelar | no |
| 22 | input/- | entrada | buscar | buscar | no |
| 23 | button/button | accion | buscarBtn | Buscar | no |
| 24 | select/- | entrada | segmentExportFormat | Excel (XLS) PDF CSV JSON TXT | no |
| 25 | button/button | accion | segmentExportBtn | Exportar segmentos | no |
| 26 | button/- | accion | ' + item.id + ' | Perfil | sí |
| 27 | button/- | accion | ' + item.id + ' | Editar | sí |
| 28 | button/- | accion | ' + item.id + ' | Eliminar | sí |
| 29 | button/- | accion | ' + item.id + ' | ' + toggleLabel + ' | sí |

### `web/administrar_empresa/administrar_productos.html` (162)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | productoLinkProduccionMRP | Abrir Produccion / MRP | no |
| 2 | button/button | accion | btnNuevaBodega | Nueva | sí |
| 3 | input/hidden | entrada | bodegaId | bodegaId | no |
| 4 | input/- | entrada | bodegaCodigo | bodegaCodigo | no |
| 5 | input/- | entrada | bodegaNombre | bodegaNombre | no |
| 6 | input/- | entrada | bodegaUbicacion | bodegaUbicacion | no |
| 7 | input/- | entrada | bodegaResponsable | bodegaResponsable | no |
| 8 | textarea/- | entrada | bodegaObservaciones | bodegaObservaciones | no |
| 9 | button/submit | accion | - | Guardar bodega | no |
| 10 | button/button | accion | btnNuevoProveedor | Nuevo | sí |
| 11 | input/hidden | entrada | proveedorId | proveedorId | no |
| 12 | input/- | entrada | proveedorNombre | proveedorNombre | no |
| 13 | input/- | entrada | proveedorContacto | proveedorContacto | no |
| 14 | input/- | entrada | proveedorTelefono | proveedorTelefono | no |
| 15 | input/email | entrada | proveedorEmail | proveedorEmail | no |
| 16 | input/- | entrada | proveedorCatalogoReferencia | proveedorCatalogoReferencia | no |
| 17 | input/- | entrada | proveedorCondicionEntrega | proveedorCondicionEntrega | no |
| 18 | input/number | entrada | proveedorPrecioBaseReferencial | 0 | no |
| 19 | input/number | entrada | proveedorDescuentoPorcentaje | 0 | no |
| 20 | input/number | entrada | proveedorPlazoPagoDias | 0 | no |
| 21 | input/- | entrada | proveedorObservaciones | proveedorObservaciones | no |
| 22 | button/submit | accion | - | Guardar proveedor | no |
| 23 | input/- | entrada | filtroCategoriasProducto | filtroCategoriasProducto | no |
| 24 | button/button | accion | btnBuscarCategoriasProducto | Buscar | no |
| 25 | button/button | accion | btnNuevaCategoriaProducto | Nueva | sí |
| 26 | button/button | accion | btnCerrarCategoriaProductoForm | Cerrar formulario | no |
| 27 | input/hidden | entrada | categoriaProductoId | categoriaProductoId | no |
| 28 | input/- | entrada | categoriaProductoCodigo | categoriaProductoCodigo | no |
| 29 | input/- | entrada | categoriaProductoNombre | categoriaProductoNombre | no |
| 30 | input/- | entrada | categoriaProductoColor | categoriaProductoColor | no |
| 31 | input/number | entrada | categoriaProductoOrden | 0 | no |
| 32 | input/- | entrada | categoriaProductoDescripcion | categoriaProductoDescripcion | no |
| 33 | input/- | entrada | categoriaProductoObservaciones | categoriaProductoObservaciones | no |
| 34 | button/submit | accion | - | Guardar categoría | no |
| 35 | select/- | entrada | filtroCategoriaProducto | Todas las categorías | no |
| 36 | input/- | entrada | filtroProductos | filtroProductos | no |
| 37 | button/button | accion | btnBuscarProductos | Buscar | no |
| 38 | button/button | accion | btnNuevoProducto | Nuevo producto | sí |
| 39 | button/button | accion | btnCrearBodegaDesdeProductos | Crear bodega | no |
| 40 | select/- | entrada | productoExportFormato | CSV / Excel JSON Imprimible HTML | no |
| 41 | select/- | entrada | productoExportTamano | Carta POS 80 mm | no |
| 42 | button/button | accion | btnExportarProductos | Exportar | no |
| 43 | button/button | accion | btnPlantillaProductos | Plantilla | no |
| 44 | input/file | accion | productoImportArchivo | productoImportArchivo | no |
| 45 | button/button | accion | btnImportarProductos | Importar productos | no |
| 46 | button/button | accion | btnCartaPreciosIA | Cargar carta/precios con IA | sí |
| 47 | button/button | accion | btnCerrarProductoForm | Cerrar formulario | no |
| 48 | input/hidden | entrada | productoId | productoId | no |
| 49 | input/- | entrada | productoNombre | productoNombre | no |
| 50 | input/- | entrada | productoSKU | productoSKU | no |
| 51 | input/- | entrada | productoCodigoBarras | productoCodigoBarras | no |
| 52 | select/- | entrada | productoCategoria | Sin categoría | no |
| 53 | input/- | entrada | productoMarca | productoMarca | no |
| 54 | input/- | entrada | productoUnidad | productoUnidad | no |
| 55 | input/number | entrada | productoCosto | 0 | no |
| 56 | input/number | entrada | productoPrecio | 0 | no |
| 57 | select/- | entrada | productoImpuesto | Sin impuesto / exento (0%) | no |
| 58 | input/number | entrada | productoStockMin | 0 | no |
| 59 | input/number | entrada | productoStockMax | 0 | no |
| 60 | input/number | entrada | productoStockInicial | 0 | no |
| 61 | select/- | entrada | productoBodegaPrincipal | productoBodegaPrincipal | no |
| 62 | select/- | entrada | productoProveedorPrincipal | productoProveedorPrincipal | no |
| 63 | input/file | accion | productoImagen | productoImagen | no |
| 64 | input/checkbox | accion | productoManejaVencimiento | productoManejaVencimiento | no |
| 65 | input/date | entrada | productoFechaVencimiento | productoFechaVencimiento | no |
| 66 | input/number | entrada | productoDiasAlertaVencimiento | 30 | no |
| 67 | input/- | entrada | productoLoteCodigo | productoLoteCodigo | no |
| 68 | input/- | entrada | productoImagenURL | productoImagenURL | no |
| 69 | select/- | entrada | productoImpresoraPedido | productoImpresoraPedido | no |
| 70 | input/- | entrada | productoMotivoPrecio | productoMotivoPrecio | no |
| 71 | input/- | entrada | productoReferenciaPrecio | productoReferenciaPrecio | no |
| 72 | textarea/- | entrada | productoDescripcion | productoDescripcion | no |
| 73 | textarea/- | entrada | productoObservaciones | productoObservaciones | no |
| 74 | button/submit | accion | - | Guardar producto | no |
| 75 | button/button | accion | btnNuevoServicio | Nuevo | sí |
| 76 | input/hidden | entrada | servicioId | servicioId | no |
| 77 | input/- | entrada | servicioNombre | servicioNombre | no |
| 78 | input/- | entrada | servicioCodigo | servicioCodigo | no |
| 79 | input/- | entrada | servicioCategoria | servicioCategoria | no |
| 80 | input/number | entrada | servicioPrecio | 0 | no |
| 81 | input/number | entrada | servicioCostoReferencial | 0 | no |
| 82 | input/number | entrada | servicioImpuesto | 0 | no |
| 83 | input/number | entrada | servicioDuracion | 0 | no |
| 84 | textarea/- | entrada | servicioDescripcion | servicioDescripcion | no |
| 85 | textarea/- | entrada | servicioObservaciones | servicioObservaciones | no |
| 86 | button/submit | accion | - | Guardar servicio | no |
| 87 | select/- | entrada | inventarioPoliticaCosto | Promedio PEPS | no |
| 88 | input/- | entrada | inventarioConfigObservaciones | inventarioConfigObservaciones | no |
| 89 | button/button | accion | btnGuardarInventarioConfig | Guardar política | no |
| 90 | select/- | entrada | ajusteProducto | ajusteProducto | no |
| 91 | select/- | entrada | ajusteBodega | ajusteBodega | no |
| 92 | select/- | entrada | ajusteTipo | Entrada Salida Devolución Pérdida Ajuste positivo Ajuste negativo | no |
| 93 | input/number | entrada | ajusteCantidad | ajusteCantidad | no |
| 94 | input/- | entrada | ajusteReferencia | ajusteReferencia | no |
| 95 | input/- | entrada | ajusteObservaciones | ajusteObservaciones | no |
| 96 | button/submit | accion | - | Registrar ajuste | no |
| 97 | select/- | entrada | conteoProducto | conteoProducto | no |
| 98 | select/- | entrada | conteoBodega | conteoBodega | no |
| 99 | input/number | entrada | conteoCantidad | conteoCantidad | no |
| 100 | input/- | entrada | conteoReferencia | conteoReferencia | no |
| 101 | input/- | entrada | conteoObservaciones | conteoObservaciones | no |
| 102 | button/submit | accion | - | Registrar conteo cíclico | no |
| 103 | select/- | entrada | cambioProductoOrigen | cambioProductoOrigen | no |
| 104 | select/- | entrada | cambioProductoDestino | cambioProductoDestino | no |
| 105 | select/- | entrada | cambioBodega | cambioBodega | no |
| 106 | input/number | entrada | cambioCantidad | cambioCantidad | no |
| 107 | input/- | entrada | cambioReferencia | cambioReferencia | no |
| 108 | input/- | entrada | cambioObservaciones | cambioObservaciones | no |
| 109 | button/submit | accion | - | Registrar cambio | no |
| 110 | select/- | entrada | transferProducto | transferProducto | no |
| 111 | select/- | entrada | transferOrigen | transferOrigen | no |
| 112 | select/- | entrada | transferDestino | transferDestino | no |
| 113 | input/number | entrada | transferCantidad | transferCantidad | no |
| 114 | input/- | entrada | transferReferencia | transferReferencia | no |
| 115 | input/- | entrada | transferObs | transferObs | no |
| 116 | button/submit | accion | - | Mover inventario | no |
| 117 | select/- | entrada | alertaBodegaFiltro | Todas las bodegas | no |
| 118 | button/button | accion | btnAplicarAlertasFiltro | Filtrar | no |
| 119 | button/button | accion | btnLimpiarAlertasFiltro | Limpiar | no |
| 120 | select/- | entrada | kardexBodegaFiltro | Todas las bodegas | no |
| 121 | select/- | entrada | kardexTipoFiltro | Todos los tipos Entrada Salida Traslado Devolución Pérdida Ajuste positivo Ajuste negativo Cambio producto | no |
| 122 | input/date | entrada | kardexDesdeFiltro | Desde | no |
| 123 | input/date | entrada | kardexHastaFiltro | Hasta | no |
| 124 | button/button | accion | btnAplicarKardexFiltros | Filtrar | no |
| 125 | button/button | accion | btnLimpiarKardexFiltros | Limpiar | no |
| 126 | select/- | entrada | comprasKardexBodegaFiltro | Todas las bodegas | no |
| 127 | input/date | entrada | comprasKardexDesdeFiltro | Desde | no |
| 128 | input/date | entrada | comprasKardexHastaFiltro | Hasta | no |
| 129 | button/button | accion | btnAplicarComprasKardexFiltros | Filtrar | no |
| 130 | button/button | accion | btnLimpiarComprasKardexFiltros | Limpiar | no |
| 131 | button/button | accion | btnLimpiarPlanProveedor | Ver todos | no |
| 132 | button/button | accion | btnEmitirOrdenBorrador | Emitir orden | no |
| 133 | button/button | accion | btnRecepcionarOrdenBorrador | Recepcionar orden | no |
| 134 | button/button | accion | btnContabilizarOrdenBorrador | Contabilizar orden | no |
| 135 | button/button | accion | btnLimpiarPlanBorrador | Limpiar borrador | no |
| 136 | select/- | entrada | historialProducto | historialProducto | no |
| 137 | button/button | accion | btnCargarHistorialPrecios | Actualizar | no |
| 138 | button/button | accion | - | CSV | sí |
| 139 | button/button | accion | - | JSON | sí |
| 140 | button/button | accion | - | HTML | sí |
| 141 | button/button | accion | - | PDF | sí |
| 142 | button/button | accion | ' + Number(b.id) + ' | Editar | sí |
| 143 | button/button | accion | ' + Number(b.id) + ' | ' + activar + ' | sí |
| 144 | button/button | accion | ' + Number(b.id) + ' | Eliminar | sí |
| 145 | button/button | accion | ' + Number(p.id) + ' | Editar | sí |
| 146 | button/button | accion | ' + Number(p.id) + ' | ' + activar + ' | sí |
| 147 | button/button | accion | ' + Number(p.id) + ' | Eliminar | sí |
| 148 | button/button | accion | ' + Number(c.id) + ' | Editar | sí |
| 149 | button/button | accion | ' + Number(c.id) + ' | ' + activar + ' | sí |
| 150 | button/button | accion | ' + Number(c.id) + ' | Eliminar | sí |
| 151 | button/button | accion | ' + Number(p.id) + ' | Editar | sí |
| 152 | button/button | accion | ' + Number(p.id) + ' | ' + activar + ' | sí |
| 153 | button/button | accion | ' + Number(p.id) + ' | Eliminar | sí |
| 154 | button/button | accion | ' + Number(s.id) + ' | Editar | sí |
| 155 | button/button | accion | ' + Number(s.id) + ' | ' + activar + ' | sí |
| 156 | button/button | accion | ' + Number(s.id) + ' | Eliminar | sí |
| 157 | button/button | accion | ' + Number(r.producto_id \|\| 0) + ' | Preparar | sí |
| 158 | button/button | accion | ' + Number(r.proveedor_id \|\| 0) + ' | Borrador OC | sí |
| 159 | button/button | accion | ' + Number(r.producto_id \|\| 0) + ' | Preparar | sí |
| 160 | button/button | accion | ' + proveedorID + ' | Ver items | sí |
| 161 | button/button | accion | ' + proveedorID + ' | Borrador OC | sí |
| 162 | button/button | accion | ' + Number(r.producto_id \|\| 0) + ' | Preparar reposición | sí |

### `web/administrar_empresa/administrar_productos_menu.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | - | ☰ Ocultar menú | sí |

### `web/administrar_empresa/administrar_usuarios.html` (39)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/- | accion | addBtn | Agregar | no |
| 2 | input/hidden | entrada | itemId | itemId | no |
| 3 | input/hidden | entrada | empresa_id | empresa_id | no |
| 4 | input/email | entrada | email | email | no |
| 5 | input/- | entrada | nombre | nombre | no |
| 6 | input/- | entrada | documento_identidad | documento_identidad | no |
| 7 | select/- | entrada | rol_usuario_id | rol_usuario_id | no |
| 8 | textarea/- | entrada | observaciones | observaciones | no |
| 9 | input/file | accion | foto_usuario | foto_usuario | no |
| 10 | input/checkbox | accion | control_aseo_estaciones | control_aseo_estaciones | no |
| 11 | textarea/- | entrada | mensaje_invitacion | mensaje_invitacion | no |
| 12 | button/submit | accion | saveBtn | Guardar | no |
| 13 | button/button | accion | cancelBtn | Cancelar | no |
| 14 | select/- | entrada | roleEditorUser | roleEditorUser | no |
| 15 | select/- | entrada | roleEditorRole | roleEditorRole | no |
| 16 | button/button | accion | roleEditorSave | Guardar rol | no |
| 17 | input/hidden | entrada | customRoleId | customRoleId | no |
| 18 | input/- | entrada | customRoleName | customRoleName | no |
| 19 | select/- | entrada | customRoleBase | customRoleBase | no |
| 20 | button/button | accion | customRoleSave | Guardar rol | no |
| 21 | button/button | accion | customRoleCancel | Cancelar | no |
| 22 | textarea/- | entrada | customRoleDescription | customRoleDescription | no |
| 23 | input/checkbox | accion | stationAccessEnabled | stationAccessEnabled | no |
| 24 | select/- | entrada | stationAccessUser | stationAccessUser | no |
| 25 | input/checkbox | accion | stationAccessLimit | stationAccessLimit | no |
| 26 | input/checkbox | accion | stationAccessCaja | stationAccessCaja | no |
| 27 | button/button | accion | stationAccessAll | Marcar todas | no |
| 28 | button/button | accion | stationAccessNone | Quitar todas | no |
| 29 | button/button | accion | stationAccessSave | Guardar acceso | no |
| 30 | button/button | accion | - | Seleccionar | no |
| 31 | a/- | accion | - | Saber mas | no |
| 32 | button/button | accion | - | Editar | no |
| 33 | button/button | accion | - | sin etiqueta | no |
| 34 | input/checkbox | accion | - | ' + id + ' | no |
| 35 | button/- | accion | ' + item.id + ' | Editar | sí |
| 36 | button/- | accion | ' + item.id + ' | Cambiar rol | sí |
| 37 | button/- | accion | ' + item.id + ' | Eliminar | sí |
| 38 | button/- | accion | ' + item.id + ' | ' + toggleLabel + ' | sí |
| 39 | button/- | accion | ' + item.id + ' | Reenviar confirmación | sí |

### `web/administrar_empresa/aiu_construccion.html` (49)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnNuevo | Nuevo | no |
| 2 | button/button | accion | btnDemo | Demo | no |
| 3 | button/button | accion | btnExportar | Exportar CSV | no |
| 4 | input/- | entrada | filtro_q | filtro_q | no |
| 5 | select/- | entrada | filtro_estado | Todos Borrador Cotizado Aprobado En ejecucion Suspendido Facturado Cerrado Anulado | no |
| 6 | button/button | accion | btnFiltrar | Filtrar | no |
| 7 | input/hidden | entrada | contrato_id | contrato_id | no |
| 8 | input/- | entrada | codigo | codigo | no |
| 9 | input/- | entrada | nombre | nombre | no |
| 10 | select/- | entrada | estado | Borrador Cotizado Aprobado En ejecucion Suspendido Facturado Cerrado Anulado | no |
| 11 | input/- | entrada | cliente_nombre | cliente_nombre | no |
| 12 | input/- | entrada | responsable | responsable | no |
| 13 | input/- | entrada | centro_costo | centro_costo | no |
| 14 | select/- | entrada | riesgo_nivel | Bajo Medio Alto Critico | no |
| 15 | select/- | entrada | tipo_obra | Obra civil Arquitectura Remodelacion Mantenimiento Servicios | no |
| 16 | select/- | entrada | modalidad_contrato | Precio global Administracion delegada Costos reembolsables Obra por unidad Mantenimiento | no |
| 17 | select/- | entrada | modelo_aiu | AIU no sumado AIU sumado al total | no |
| 18 | select/- | entrada | base_iva_modo | Utilidad AIU total Costo + AIU | no |
| 19 | input/date | entrada | fecha_inicio | fecha_inicio | no |
| 20 | input/date | entrada | fecha_fin | fecha_fin | no |
| 21 | input/number | entrada | avance_porcentaje | 0 | no |
| 22 | input/number | entrada | porcentaje_admin | 10 | no |
| 23 | input/number | entrada | porcentaje_imprevistos | 5 | no |
| 24 | input/number | entrada | porcentaje_utilidad | 10 | no |
| 25 | input/number | entrada | porcentaje_iva | 19 | no |
| 26 | input/number | entrada | porcentaje_retencion_fuente | 0 | no |
| 27 | input/number | entrada | porcentaje_retencion_ica | 0 | no |
| 28 | input/number | entrada | porcentaje_retencion_iva | 0 | no |
| 29 | input/number | entrada | porcentaje_anticipo | 0 | no |
| 30 | input/number | entrada | porcentaje_garantia | 0 | no |
| 31 | textarea/- | entrada | observaciones | observaciones | no |
| 32 | button/submit | accion | - | Guardar | no |
| 33 | button/button | accion | btnCalcular | Calcular | no |
| 34 | button/button | accion | - | Cotizar | sí |
| 35 | button/button | accion | - | Aprobar | sí |
| 36 | button/button | accion | - | Iniciar | sí |
| 37 | button/button | accion | - | Suspender | sí |
| 38 | button/button | accion | - | Cerrar | sí |
| 39 | button/button | accion | - | Anular | sí |
| 40 | input/- | entrada | item_capitulo | item_capitulo | no |
| 41 | input/- | entrada | item_descripcion | item_descripcion | no |
| 42 | input/- | entrada | item_unidad | und | no |
| 43 | input/number | entrada | item_cantidad | 1 | no |
| 44 | input/number | entrada | item_valor_unitario | 0 | no |
| 45 | button/submit | accion | - | Agregar concepto | no |
| 46 | input/- | entrada | factura_documento_codigo | factura_documento_codigo | no |
| 47 | input/- | entrada | factura_periodo | factura_periodo | no |
| 48 | button/button | accion | btnGenerarFactura | Generar factura electronica | no |
| 49 | button/button | accion | ${x.id} | Abrir | sí |

### `web/administrar_empresa/alquileres.html` (100)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | - | Dashboard | sí |
| 2 | button/button | accion | - | Configuración | sí |
| 3 | button/button | accion | - | Activos | sí |
| 4 | button/button | accion | - | Tarifas | sí |
| 5 | button/button | accion | - | Contratos | sí |
| 6 | button/button | accion | - | Mantenimientos | sí |
| 7 | button/button | accion | - | Mapa | sí |
| 8 | input/- | entrada | rentalSystemName | rentalSystemName | no |
| 9 | input/- | entrada | rentalCurrency | COP | no |
| 10 | input/number | entrada | rentalAlertHours | 12 | no |
| 11 | input/number | entrada | rentalDepositBase | 0 | no |
| 12 | input/checkbox | accion | rentalAllowReservations | rentalAllowReservations | no |
| 13 | input/checkbox | accion | rentalAllowGPS | rentalAllowGPS | no |
| 14 | input/checkbox | accion | rentalRequireDeposit | rentalRequireDeposit | no |
| 15 | input/checkbox | accion | rentalAllowMileage | rentalAllowMileage | no |
| 16 | input/checkbox | accion | rentalRequireChecklist | rentalRequireChecklist | no |
| 17 | input/checkbox | accion | rentalAllowDelivery | rentalAllowDelivery | no |
| 18 | button/submit | accion | - | Guardar configuración | no |
| 19 | button/button | accion | rentalSeedBtn | Preconfigurar módulo | no |
| 20 | input/hidden | entrada | rentalCategoryId | rentalCategoryId | no |
| 21 | input/- | entrada | rentalCategoryCode | rentalCategoryCode | no |
| 22 | input/- | entrada | rentalCategoryName | rentalCategoryName | no |
| 23 | select/- | entrada | rentalCategoryType | Equipo Herramienta Herramienta electrica Vehiculo Moto Maquinaria Mobiliario Sonido / eventos Tecnologia Objeto | no |
| 24 | button/submit | accion | - | Guardar categoría | no |
| 25 | input/hidden | entrada | rentalAssetId | rentalAssetId | no |
| 26 | input/- | entrada | rentalAssetCode | rentalAssetCode | no |
| 27 | input/- | entrada | rentalAssetName | rentalAssetName | no |
| 28 | select/- | entrada | rentalAssetCategory | rentalAssetCategory | no |
| 29 | select/- | entrada | rentalAssetType | Equipo Herramienta Herramienta electrica Vehiculo Moto Maquinaria Mobiliario Sonido / eventos Tecnologia Objeto | no |
| 30 | input/- | entrada | rentalAssetBrand | rentalAssetBrand | no |
| 31 | input/- | entrada | rentalAssetModel | rentalAssetModel | no |
| 32 | input/- | entrada | rentalAssetSerie | rentalAssetSerie | no |
| 33 | input/- | entrada | rentalAssetPlate | rentalAssetPlate | no |
| 34 | input/- | entrada | rentalAssetSede | principal | no |
| 35 | input/number | entrada | rentalAssetReplacement | 0 | no |
| 36 | input/number | entrada | rentalAssetBaseCost | 0 | no |
| 37 | input/number | entrada | rentalAssetDeposit | 0 | no |
| 38 | select/- | entrada | rentalAssetStatus | Disponible Reservado Alquilado Mantenimiento Fuera de servicio | no |
| 39 | input/checkbox | accion | rentalAssetGPS | rentalAssetGPS | no |
| 40 | input/checkbox | accion | rentalAssetChecklist | rentalAssetChecklist | no |
| 41 | input/checkbox | accion | rentalAssetLicense | rentalAssetLicense | no |
| 42 | button/submit | accion | - | Guardar activo | no |
| 43 | input/hidden | entrada | rentalRateId | rentalRateId | no |
| 44 | input/- | entrada | rentalRateCode | rentalRateCode | no |
| 45 | input/- | entrada | rentalRateName | rentalRateName | no |
| 46 | select/- | entrada | rentalRateCategory | rentalRateCategory | no |
| 47 | select/- | entrada | rentalRateMode | Hora Día Semana Mes Kilómetro Evento | no |
| 48 | input/number | entrada | rentalRateBase | 0 | no |
| 49 | input/number | entrada | rentalRateDeposit | 0 | no |
| 50 | input/number | entrada | rentalRateHour | 0 | no |
| 51 | input/number | entrada | rentalRateDay | 0 | no |
| 52 | input/number | entrada | rentalRateWeek | 0 | no |
| 53 | input/number | entrada | rentalRateMonth | 0 | no |
| 54 | input/number | entrada | rentalRateKm | 0 | no |
| 55 | select/- | entrada | rentalRateStatus | Activa Inactiva | no |
| 56 | button/submit | accion | - | Guardar tarifa | no |
| 57 | input/hidden | entrada | rentalContractId | rentalContractId | no |
| 58 | input/- | entrada | rentalContractCode | rentalContractCode | no |
| 59 | select/- | entrada | rentalContractType | Reserva Alquiler | no |
| 60 | select/- | entrada | rentalContractAsset | rentalContractAsset | no |
| 61 | input/- | entrada | rentalContractClient | rentalContractClient | no |
| 62 | input/- | entrada | rentalContractDocument | rentalContractDocument | no |
| 63 | input/- | entrada | rentalContractPhone | rentalContractPhone | no |
| 64 | input/email | entrada | rentalContractEmail | rentalContractEmail | no |
| 65 | select/- | entrada | rentalContractTariff | rentalContractTariff | no |
| 66 | select/- | entrada | rentalContractMode | Hora Día Semana Mes Kilómetro Evento | no |
| 67 | input/datetime-local | entrada | rentalContractStart | rentalContractStart | no |
| 68 | input/datetime-local | entrada | rentalContractEnd | rentalContractEnd | no |
| 69 | select/- | entrada | rentalContractStatus | Reservado En curso Vencido Devuelto Cancelado | no |
| 70 | input/number | entrada | rentalContractDays | 0 | no |
| 71 | input/number | entrada | rentalContractHours | 0 | no |
| 72 | input/number | entrada | rentalContractKm | 0 | no |
| 73 | input/number | entrada | rentalContractBase | 0 | no |
| 74 | input/number | entrada | rentalContractDeposit | 0 | no |
| 75 | input/number | entrada | rentalContractTaxes | 0 | no |
| 76 | input/number | entrada | rentalContractDiscount | 0 | no |
| 77 | input/- | entrada | rentalContractOrigin | rentalContractOrigin | no |
| 78 | input/- | entrada | rentalContractReturn | rentalContractReturn | no |
| 79 | input/checkbox | accion | rentalContractGuarantee | rentalContractGuarantee | no |
| 80 | input/checkbox | accion | rentalContractGPS | rentalContractGPS | no |
| 81 | button/submit | accion | - | Guardar contrato | no |
| 82 | input/hidden | entrada | rentalMaintenanceId | rentalMaintenanceId | no |
| 83 | select/- | entrada | rentalMaintenanceAsset | rentalMaintenanceAsset | no |
| 84 | select/- | entrada | rentalMaintenanceType | Preventivo Correctivo Inspección | no |
| 85 | select/- | entrada | rentalMaintenancePriority | Baja Media Alta Crítica | no |
| 86 | select/- | entrada | rentalMaintenanceStatus | Abierto En proceso Cerrado | no |
| 87 | input/date | entrada | rentalMaintenanceDate | rentalMaintenanceDate | no |
| 88 | input/- | entrada | rentalMaintenanceProvider | rentalMaintenanceProvider | no |
| 89 | input/number | entrada | rentalMaintenanceEstimated | 0 | no |
| 90 | input/number | entrada | rentalMaintenanceReal | 0 | no |
| 91 | input/- | entrada | rentalMaintenanceDescription | rentalMaintenanceDescription | no |
| 92 | button/submit | accion | - | Guardar mantenimiento | no |
| 93 | select/- | entrada | rentalLocationAsset | rentalLocationAsset | no |
| 94 | select/- | entrada | rentalLocationContract | rentalLocationContract | no |
| 95 | select/- | entrada | rentalLocationSource | Manual Móvil GPS | no |
| 96 | input/number | entrada | rentalLocationLat | rentalLocationLat | no |
| 97 | input/number | entrada | rentalLocationLng | rentalLocationLng | no |
| 98 | input/- | entrada | rentalLocationRef | rentalLocationRef | no |
| 99 | button/submit | accion | - | Registrar ubicación | no |
| 100 | button/button | accion | rentalUseMyLocation | Usar mi GPS | no |

### `web/administrar_empresa/asistencia_empleados.html` (43)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | addBtn | Agregar registro | no |
| 2 | input/number | entrada | cfg_tolerancia_entrada | 10 | no |
| 3 | input/time | entrada | cfg_hora_manana | 06:00:00 | no |
| 4 | input/time | entrada | cfg_hora_tarde | 14:00:00 | no |
| 5 | input/time | entrada | cfg_hora_noche | 22:00:00 | no |
| 6 | input/checkbox | accion | cfg_permitir_nocturno | cfg_permitir_nocturno | no |
| 7 | input/checkbox | accion | cfg_permitir_cruzado | cfg_permitir_cruzado | no |
| 8 | button/button | accion | saveConfigBtn | Guardar configuración | no |
| 9 | input/date | entrada | close_periodo_desde | close_periodo_desde | no |
| 10 | input/date | entrada | close_periodo_hasta | close_periodo_hasta | no |
| 11 | input/- | entrada | close_periodo_motivo | close_periodo_motivo | no |
| 12 | button/button | accion | closePeriodoBtn | Cerrar periodo | no |
| 13 | button/button | accion | refreshPeriodosBtn | Actualizar cierres | no |
| 14 | select/- | entrada | reporteFormato | PDF Excel (XLS) CSV JSON TXT | no |
| 15 | button/button | accion | downloadReporteBtn | Descargar reporte | no |
| 16 | input/hidden | entrada | itemId | itemId | no |
| 17 | input/hidden | entrada | empresa_id | empresa_id | no |
| 18 | select/- | entrada | usuario_empresa_id | Sin vincular / registro manual | no |
| 19 | input/- | entrada | empleado_codigo | empleado_codigo | no |
| 20 | input/- | entrada | empleado_nombre | empleado_nombre | no |
| 21 | input/- | entrada | empleado_documento | empleado_documento | no |
| 22 | input/- | entrada | cargo | cargo | no |
| 23 | select/- | entrada | turno | General Manana Tarde Noche Mixto Rotativo | no |
| 24 | input/date | entrada | fecha_asistencia | fecha_asistencia | no |
| 25 | input/time | entrada | hora_entrada | hora_entrada | no |
| 26 | input/time | entrada | hora_salida | hora_salida | no |
| 27 | input/number | entrada | minutos_tarde | 0 | no |
| 28 | select/- | entrada | estado_asistencia | Pendiente Presente Tarde Ausente Permiso Incapacidad Vacaciones | no |
| 29 | input/- | entrada | novedad | novedad | no |
| 30 | textarea/- | entrada | observaciones | observaciones | no |
| 31 | button/submit | accion | saveBtn | Guardar | no |
| 32 | button/button | accion | cancelBtn | Cancelar | no |
| 33 | input/- | entrada | buscar | buscar | no |
| 34 | input/date | entrada | filtro_desde | filtro_desde | no |
| 35 | input/date | entrada | filtro_hasta | filtro_hasta | no |
| 36 | select/- | entrada | filtro_estado_asistencia | Todos Pendiente Presente Tarde Ausente Permiso Incapacidad Vacaciones | no |
| 37 | button/button | accion | filtrarBtn | Filtrar | no |
| 38 | button/button | accion | limpiarBtn | Limpiar | no |
| 39 | button/button | accion | - | Editar | sí |
| 40 | button/button | accion | ' + id + ' | Entrada ahora | sí |
| 41 | button/button | accion | ' + id + ' | Salida ahora | sí |
| 42 | button/button | accion | ' + id + ' | ' + nextLabel + ' | sí |
| 43 | button/button | accion | ' + id + ' | Eliminar | sí |

### `web/administrar_empresa/auditoria.html` (25)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnBuscar | Buscar | no |
| 2 | button/button | accion | btnLimpiar | Limpiar | no |
| 3 | button/button | accion | btnExportCSV | Exportar CSV | no |
| 4 | button/button | accion | btnExportJSON | Exportar JSON | no |
| 5 | select/- | entrada | fModulo | Todos Ventas Carritos / cierres Venta publica Compras Compras avanzadas Ordenes de compra Proveedores Soportes compras I | no |
| 6 | select/- | entrada | fMetodoHttp | Todos GET POST PUT PATCH DELETE | no |
| 7 | input/- | entrada | fAcción | fAcción | no |
| 8 | select/- | entrada | fResultado | Todos OK Error | no |
| 9 | input/- | entrada | fUsuario | fUsuario | no |
| 10 | input/- | entrada | fRequestId | fRequestId | no |
| 11 | input/- | entrada | fRecurso | fRecurso | no |
| 12 | input/- | entrada | fEndpoint | fEndpoint | no |
| 13 | input/- | entrada | fSearch | fSearch | no |
| 14 | input/date | entrada | fDesde | fDesde | no |
| 15 | input/date | entrada | fHasta | fHasta | no |
| 16 | input/number | entrada | fLimit | 200 | no |
| 17 | input/number | entrada | fCodigoHttp | fCodigoHttp | no |
| 18 | input/number | entrada | fRecursoId | fRecursoId | no |
| 19 | input/checkbox | accion | fIncluirInactivos | fIncluirInactivos | no |
| 20 | input/number | entrada | retencionDias | 180 | no |
| 21 | button/button | accion | btnPurgar | Aplicar retencion | no |
| 22 | button/button | accion | btnPrevPage | Anterior | no |
| 23 | button/button | accion | btnNextPage | Siguiente | no |
| 24 | button/button | accion | btnCerrarDetalle | Cerrar | no |
| 25 | button/button | accion | - | Ver | sí |

### `web/administrar_empresa/backups.html` (42)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/text | entrada | bkConfigNombre | bkConfigNombre | no |
| 2 | input/text | entrada | bkConfigDescripcion | bkConfigDescripcion | no |
| 3 | input/file | accion | bkConfigFile | bkConfigFile | no |
| 4 | button/button | accion | btnExportarConfiguracionEmpresa | Exportar configuración | no |
| 5 | button/button | accion | btnImportarConfiguracionEmpresa | Importar configuración | no |
| 6 | input/checkbox | accion | bkAutoEnabled | bkAutoEnabled | no |
| 7 | select/- | entrada | bkAutoScope | Datos completos de la empresa Solo configuracion | no |
| 8 | select/- | entrada | bkAutoInterval | Cada 30 minutos Cada 1 hora Cada 6 horas Cada dia | no |
| 9 | button/button | accion | btnGuardarBackupAutomatico | Guardar backup automatico | no |
| 10 | button/button | accion | btnEjecutarBackupLocalAhora | Ejecutar local ahora | no |
| 11 | input/text | entrada | bkNombre | bkNombre | no |
| 12 | input/text | entrada | bkDescripcion | bkDescripcion | no |
| 13 | input/text | entrada | bkInclude | bkInclude | no |
| 14 | input/text | entrada | bkExclude | bkExclude | no |
| 15 | button/button | accion | btnCrearBackup | Crear snapshot interno | no |
| 16 | select/- | entrada | bkDownloadFormat | JSON (completo) JSON (.gz) CSV (resumen) Excel (resumen) TXT (resumen) PDF (resumen) | no |
| 17 | button/button | accion | btnDescargarDatosEmpresa | Descargar datos al equipo | no |
| 18 | input/email | entrada | bkEmailTo | bkEmailTo | no |
| 19 | button/button | accion | btnEnviarBackupEmail | Enviar por correo | no |
| 20 | button/button | accion | btnRefrescarBackups | Refrescar | no |
| 21 | select/- | entrada | bkResetModo | Eliminar hasta una fecha Eliminar todos los tiempos | no |
| 22 | input/date | entrada | bkResetFechaCorte | bkResetFechaCorte | no |
| 23 | input/checkbox | accion | bkResetCrearBackupPrevio | bkResetCrearBackupPrevio | no |
| 24 | input/text | entrada | bkResetInclude | bkResetInclude | no |
| 25 | input/text | entrada | bkResetExclude | bkResetExclude | no |
| 26 | input/text | entrada | bkResetConfirmacion | bkResetConfirmacion | no |
| 27 | button/button | accion | btnPrevisualizarReset | Previsualizar registros | no |
| 28 | button/button | accion | btnResetOperativo | Eliminar registros operativos | no |
| 29 | input/date | entrada | bkPurgeFechaCorte | bkPurgeFechaCorte | no |
| 30 | input/checkbox | accion | bkPurgeCrearBackupPrevio | bkPurgeCrearBackupPrevio | no |
| 31 | input/text | entrada | bkPurgeInclude | bkPurgeInclude | no |
| 32 | input/text | entrada | bkPurgeExclude | bkPurgeExclude | no |
| 33 | button/button | accion | btnPurgarFecha | Eliminar información hasta la fecha | no |
| 34 | input/text | entrada | bkQ | bkQ | no |
| 35 | input/checkbox | accion | bkIncludeInactive | bkIncludeInactive | no |
| 36 | button/button | accion | ' + id + ' | Detalle | sí |
| 37 | button/button | accion | ' + id + ' | JSON | sí |
| 38 | button/button | accion | ' + id + ' | CSV | sí |
| 39 | button/button | accion | ' + id + ' | PDF | sí |
| 40 | button/button | accion | ' + id + ' | Email | sí |
| 41 | button/button | accion | ' + id + ' | Restaurar | sí |
| 42 | button/button | accion | ' + id + ' | ' + toggleLabel + ' | sí |

### `web/administrar_empresa/bodega.html` (12)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnNuevaBodega | Nueva | no |
| 2 | input/hidden | entrada | bodegaId | bodegaId | no |
| 3 | input/- | entrada | bodegaCodigo | bodegaCodigo | no |
| 4 | input/- | entrada | bodegaNombre | bodegaNombre | no |
| 5 | input/- | entrada | bodegaUbicacion | bodegaUbicacion | no |
| 6 | input/- | entrada | bodegaResponsable | bodegaResponsable | no |
| 7 | textarea/- | entrada | bodegaObservaciones | bodegaObservaciones | no |
| 8 | button/submit | accion | - | Guardar bodega | no |
| 9 | button/button | accion | btnCancelarBodega | Cancelar | no |
| 10 | button/button | accion | ' + Number(b.id) + ' | Editar | sí |
| 11 | button/button | accion | ' + Number(b.id) + ' | ' + activar + ' | sí |
| 12 | button/button | accion | ' + Number(b.id) + ' | Eliminar | sí |

### `web/administrar_empresa/bolsa.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | refreshBtn | Actualizar | no |

### `web/administrar_empresa/buscar_producto_botones.html` (4)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnAplicarSeleccion | Agregar selección al carrito | no |
| 2 | button/button | accion | btnSalir | Salir al carrito | no |
| 3 | input/- | entrada | filtroProductoInput | filtroProductoInput | no |
| 4 | button/button | accion | - | 0 ? ' is-selected' : '') + (isAdding ? ' is-adding' : '') + '" style="--product-tone:' + tone + ';" data-id="' + Number( | sí |

### `web/administrar_empresa/camaras.html` (23)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnCamarasRefresh | Actualizar | no |
| 2 | input/hidden | entrada | camaraId | camaraId | no |
| 3 | input/- | entrada | camaraNombre | camaraNombre | no |
| 4 | input/- | entrada | camaraUbicacion | camaraUbicacion | no |
| 5 | input/- | entrada | camaraDvrNombre | camaraDvrNombre | no |
| 6 | input/- | entrada | camaraDvrHost | camaraDvrHost | no |
| 7 | input/- | entrada | camaraCanal | camaraCanal | no |
| 8 | input/- | entrada | camaraFabricante | camaraFabricante | no |
| 9 | input/- | entrada | camaraModelo | camaraModelo | no |
| 10 | select/- | entrada | camaraProtocolo | camaraProtocolo | no |
| 11 | select/- | entrada | camaraVisor | camaraVisor | no |
| 12 | input/number | entrada | camaraEstacionId | 0 | no |
| 13 | input/- | entrada | camaraUrlStream | camaraUrlStream | no |
| 14 | input/- | entrada | camaraUrlEmbed | camaraUrlEmbed | no |
| 15 | input/- | entrada | camaraUrlSnapshot | camaraUrlSnapshot | no |
| 16 | input/- | entrada | camaraUsuarioRef | camaraUsuarioRef | no |
| 17 | input/- | entrada | camaraPasswordRef | camaraPasswordRef | no |
| 18 | input/number | entrada | camaraOrden | 0 | no |
| 19 | input/checkbox | accion | camaraCargarEstaciones | camaraCargarEstaciones | no |
| 20 | input/checkbox | accion | camaraActiva | camaraActiva | no |
| 21 | textarea/- | entrada | camaraObservaciones | camaraObservaciones | no |
| 22 | button/submit | accion | - | Guardar camara | no |
| 23 | button/button | accion | btnCamaraLimpiar | Limpiar | no |

### `web/administrar_empresa/carnets.html` (41)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | seedBtn | Cargar demo | no |
| 2 | button/button | accion | printBtn | Imprimir | no |
| 3 | input/- | entrada | filterQ | filterQ | no |
| 4 | select/- | entrada | filterEstado | Todos Vigente Pendiente Suspendido Vencido Revocado | no |
| 5 | button/button | accion | reloadBtn | Recargar | no |
| 6 | input/hidden | entrada | carnetId | carnetId | no |
| 7 | select/- | entrada | usuario_id | usuario_id | no |
| 8 | select/- | entrada | plantilla_id | plantilla_id | no |
| 9 | select/- | entrada | tipo_persona | Empleado Usuario Contratista Visitante Temporal Directivo | no |
| 10 | input/- | entrada | nombre_completo | nombre_completo | no |
| 11 | input/- | entrada | documento | documento | no |
| 12 | input/- | entrada | cargo | cargo | no |
| 13 | input/- | entrada | area | area | no |
| 14 | input/- | entrada | nivel_acceso | nivel_acceso | no |
| 15 | input/email | entrada | email | email | no |
| 16 | input/- | entrada | telefono | telefono | no |
| 17 | input/- | entrada | grupo_sanguineo | grupo_sanguineo | no |
| 18 | input/- | entrada | foto_url | foto_url | no |
| 19 | input/date | entrada | fecha_vencimiento | fecha_vencimiento | no |
| 20 | input/- | entrada | contacto_emergencia | contacto_emergencia | no |
| 21 | input/- | entrada | telefono_emergencia | telefono_emergencia | no |
| 22 | textarea/- | entrada | observaciones | observaciones | no |
| 23 | button/submit | accion | - | Guardar carnet | no |
| 24 | button/button | accion | newBtn | Nuevo | no |
| 25 | button/button | accion | suspendBtn | Suspender | no |
| 26 | button/button | accion | revokeBtn | Revocar | no |
| 27 | input/hidden | entrada | templateId | templateId | no |
| 28 | input/- | entrada | templateNombre | Carnet corporativo moderno | no |
| 29 | select/- | entrada | templateOrientacion | Vertical Horizontal | no |
| 30 | input/color | entrada | colorPrimario | #1f6feb | no |
| 31 | input/color | entrada | colorSecundario | #0f172a | no |
| 32 | input/checkbox | accion | mostrarLogo | mostrarLogo | no |
| 33 | input/checkbox | accion | mostrarFoto | mostrarFoto | no |
| 34 | input/checkbox | accion | mostrarQR | mostrarQR | no |
| 35 | input/checkbox | accion | mostrarCodigoBarras | mostrarCodigoBarras | no |
| 36 | button/submit | accion | - | Guardar plantilla | no |
| 37 | button/button | accion | applyTemplateBtn | Aplicar vista | no |
| 38 | button/button | accion | exportPngBtn | PNG | no |
| 39 | button/button | accion | exportSvgBtn | SVG | no |
| 40 | button/button | accion | markPrintedBtn | Marcar impreso | no |
| 41 | button/button | accion | ' + esc(c.id) + ' | ' + ' ' + esc(c.nombre_completo) + ' ' + esc(c.codigo) + ' · ' + esc(c.cargo \|\| c.tipo_persona) + ' · ' + esc(c.area \|\|  | sí |

### `web/administrar_empresa/carrito_control_electrico.html` (22)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | select/- | entrada | cartElectricStationFilter | cartElectricStationFilter | no |
| 2 | button/button | accion | cartElectricRefreshBtn | Actualizar | no |
| 3 | button/button | accion | cartElectricBackBtn | Volver | no |
| 4 | button/- | accion | - | Cerrar | no |
| 5 | input/checkbox | accion | cartElectricScheduleEnabled | cartElectricScheduleEnabled | no |
| 6 | input/datetime-local | entrada | cartElectricScheduleOn | cartElectricScheduleOn | no |
| 7 | input/datetime-local | entrada | cartElectricScheduleOff | cartElectricScheduleOff | no |
| 8 | select/- | entrada | cartElectricScheduleDays | Todos Lunes a viernes Sábado y domingo | no |
| 9 | button/- | accion | - | Cancelar | no |
| 10 | button/- | accion | cartElectricScheduleSave | Guardar programación | no |
| 11 | button/- | accion | - | Cerrar | no |
| 12 | button/button | accion | - | 15 minutos | sí |
| 13 | button/button | accion | - | 30 minutos | sí |
| 14 | button/button | accion | - | 1 hora | sí |
| 15 | input/number | entrada | cartElectricTimerHours | 0 | no |
| 16 | input/number | entrada | cartElectricTimerMinutes | 15 | no |
| 17 | input/number | entrada | cartElectricTimerSeconds | 0 | no |
| 18 | button/- | accion | - | Cancelar | no |
| 19 | button/- | accion | - | Iniciar temporizador | no |
| 20 | button/button | accion | ' + escapeHtml(rele.id \|\| 0) + ' | ' + escapeHtml(controlLabel(rele, isOn)) + ' | sí |
| 21 | button/button | accion | ' + escapeHtml(rele.id \|\| 0) + ' | Programar | sí |
| 22 | button/button | accion | ' + escapeHtml(rele.id \|\| 0) + ' | ' + timerText(rele.id) + ' | sí |

### `web/administrar_empresa/carrito_de_compras.html` (159)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/- | accion | addCarritoBtn | Agregar carrito | no |
| 2 | input/hidden | entrada | carritoId | carritoId | no |
| 3 | input/hidden | entrada | empresa_id | empresa_id | no |
| 4 | input/- | entrada | carritoNombre | carritoNombre | no |
| 5 | input/- | entrada | carritoCodigo | carritoCodigo | no |
| 6 | select/- | entrada | carritoCanal | Mostrador Domicilio Ecommerce Teléfono Otro | no |
| 7 | select/- | entrada | carritoClienteID | carritoClienteID | no |
| 8 | select/- | entrada | carritoMoneda | COP USD EUR | no |
| 9 | input/- | entrada | carritoReferencia | carritoReferencia | no |
| 10 | textarea/- | entrada | carritoObservaciones | carritoObservaciones | no |
| 11 | button/submit | accion | saveCarritoBtn | Guardar carrito | no |
| 12 | button/button | accion | cancelCarritoBtn | Cancelar | no |
| 13 | input/- | entrada | buscarCarrito | buscarCarrito | no |
| 14 | button/button | accion | buscarCarritoBtn | Buscar | no |
| 15 | button/button | accion | backToStationsHeaderBtn | Regresar a estaciones | no |
| 16 | button/button | accion | directSaleFullscreenBtn | Pantalla completa ⛶ | no |
| 17 | select/- | entrada | scannerSearchMode | Codigo de barras Codigo SKU Nombre | no |
| 18 | input/- | entrada | scannerCodigo | scannerCodigo | no |
| 19 | input/- | entrada | scannerSku | scannerSku | no |
| 20 | input/- | entrada | scannerNombre | scannerNombre | no |
| 21 | button/button | accion | scannerCantidadMenos | - | no |
| 22 | input/number | entrada | scannerCantidad | 1 | no |
| 23 | button/button | accion | scannerCantidadMas | + | no |
| 24 | button/button | accion | scannerAgregarBtn | Agregar | no |
| 25 | button/button | accion | carritoBtnBuscarBotones | Buscar Productos | no |
| 26 | select/- | entrada | scaleBaudRate | 9600 baudios 4800 baudios 19200 baudios 2400 baudios | no |
| 27 | select/- | entrada | scaleDefaultUnit | kg g lb | no |
| 28 | input/checkbox | accion | scaleApplyWeight | scaleApplyWeight | no |
| 29 | button/button | accion | scaleConnectBtn | Conectar bascula | no |
| 30 | button/button | accion | scaleDisconnectBtn | Desconectar | no |
| 31 | button/button | accion | scaleTareBtn | Tara local | no |
| 32 | select/- | entrada | stationCarritoClienteID | stationCarritoClienteID | no |
| 33 | button/button | accion | stationAssignClienteBtn | Asignar cliente | no |
| 34 | select/- | entrada | stationClienteSearchMode | Nombre NIT / cedula / identificacion | no |
| 35 | input/- | entrada | stationClienteSearchInput | stationClienteSearchInput | no |
| 36 | button/button | accion | stationNewClienteToggleBtn | Nuevo cliente | no |
| 37 | input/hidden | entrada | quickClienteEditID | quickClienteEditID | no |
| 38 | select/- | entrada | quickClienteTipoPersona | Persona natural Empresa / persona juridica | no |
| 39 | select/- | entrada | quickClienteTipoDoc | CC NIT CE PAS TI | no |
| 40 | input/- | entrada | quickClienteDocumento | quickClienteDocumento | no |
| 41 | input/- | entrada | quickClienteDV | quickClienteDV | no |
| 42 | input/- | entrada | quickClienteNombre | quickClienteNombre | no |
| 43 | input/- | entrada | quickClienteNombreComercial | quickClienteNombreComercial | no |
| 44 | select/- | entrada | quickClienteRegimenFiscal | No responsable de IVA Responsable de IVA Regimen SIMPLE Regimen ordinario Gran contribuyente Autorretenedor | no |
| 45 | select/- | entrada | quickClienteResponsabilidad | R-99-PN - No aplica / persona natural O-13 - Gran contribuyente O-15 - Autorretenedor O-23 - Agente retencion IVA O-47 - | no |
| 46 | input/- | entrada | quickClienteTelefono | quickClienteTelefono | no |
| 47 | input/email | entrada | quickClienteEmail | quickClienteEmail | no |
| 48 | input/- | entrada | quickClienteDireccion | quickClienteDireccion | no |
| 49 | input/- | entrada | quickClientePais | CO | no |
| 50 | input/- | entrada | quickClienteDepartamento | quickClienteDepartamento | no |
| 51 | input/- | entrada | quickClienteMunicipio | quickClienteMunicipio | no |
| 52 | input/- | entrada | quickClienteCodigoPostal | quickClienteCodigoPostal | no |
| 53 | input/- | entrada | quickClienteObservaciones | quickClienteObservaciones | no |
| 54 | input/checkbox | accion | quickClienteAbrirCupoCredito | quickClienteAbrirCupoCredito | no |
| 55 | input/number | entrada | quickClienteCupoCredito | 0 | no |
| 56 | input/number | entrada | quickClienteMaxCreditosActivos | 3 | no |
| 57 | button/button | accion | quickClienteGuardarBtn | Guardar cliente | no |
| 58 | button/button | accion | quickClienteEditarExistenteBtn | Editar cliente existente | no |
| 59 | select/- | entrada | carritoActionSelect | Selecciona una acción Historial de productos ☑ Descuentos ☑ Cambiar tarifa ☐ Transferir cuenta ☐ Domótica ☑ Cancelar car | no |
| 60 | button/button | accion | carritoBtnControlElectrico | ⚡ Domótica | no |
| 61 | button/button | accion | carritoBtnDescuentos | Descuentos | no |
| 62 | button/button | accion | carritoBtnCambiarTarifa | Cambiar tarifa | no |
| 63 | button/button | accion | carritoBtnTransferirCuenta | Transferir cuenta | no |
| 64 | button/button | accion | carritoBtnCancelarCarrito | Cancelar carrito | no |
| 65 | input/checkbox | accion | carritoAlerta10Enable | carritoAlerta10Enable | no |
| 66 | button/button | accion | carritoBtnClientes | Clientes | no |
| 67 | button/button | accion | carritoBtnAbonos | Abonos | no |
| 68 | button/button | accion | carritoBtnPagoQR | QR de pago | no |
| 69 | button/button | accion | carritoBtnVehiculo | Vehículo | no |
| 70 | select/- | entrada | carritoTarifaModo | Automática Motel por tiempo Hotel por día | no |
| 71 | select/- | entrada | carritoTarifaSelect | Selecciona una tarifa | no |
| 72 | button/button | accion | carritoTarifaAplicarBtn | Aplicar tarifa | no |
| 73 | select/- | entrada | carritoTransferDestinoSelect | Selecciona destino | no |
| 74 | input/- | entrada | carritoTransferMotivo | carritoTransferMotivo | no |
| 75 | button/button | accion | carritoTransferAplicarBtn | Transferir | no |
| 76 | button/button | accion | carritoTransferCerrarBtn | Cerrar | no |
| 77 | input/number | entrada | carritoAbonoMonto | 0 | no |
| 78 | select/- | entrada | carritoAbonoMetodo | Efectivo Tarjeta crédito Tarjeta débito Transferencia Bre-B Nequi Otra transferencia | no |
| 79 | input/text | entrada | carritoAbonoReferencia | carritoAbonoReferencia | no |
| 80 | button/button | accion | carritoAbonoRegistrarBtn | Registrar abono | no |
| 81 | button/button | accion | - | Cerrar | sí |
| 82 | button/button | accion | domoticaDeviceOffBtn | Apagar | no |
| 83 | button/button | accion | domoticaDeviceOnBtn | Encender | no |
| 84 | button/button | accion | domoticaDeviceSaveScheduleBtn | Guardar programacion | no |
| 85 | button/button | accion | - | Cancelar | sí |
| 86 | button/button | accion | - | Venta sola | sí |
| 87 | button/button | accion | - | Venta con factura electronica | sí |
| 88 | input/checkbox | accion | toggleStationCheckoutAdvanced | toggleStationCheckoutAdvanced | no |
| 89 | input/checkbox | accion | toggleStationWorkerCommission | toggleStationWorkerCommission | no |
| 90 | select/- | entrada | discountType | Sin descuento Porcentaje Código Valor fijo | no |
| 91 | input/number | entrada | discountValue | 0 | no |
| 92 | input/- | entrada | discountCode | discountCode | no |
| 93 | select/- | entrada | paymentMethod | Efectivo Tarjeta crédito Tarjeta débito Transferencia Bre-B Nequi Otra transferencia Crédito cliente Pago mixto Código d | no |
| 94 | select/- | entrada | activeCashRegister | Cargando cajas... | no |
| 95 | input/- | entrada | paymentReference | paymentReference | no |
| 96 | input/number | entrada | devolucionMonto | 0 | no |
| 97 | input/checkbox | accion | applyTipCheckbox | applyTipCheckbox | no |
| 98 | select/- | entrada | mixedMethod1 | Efectivo Tarjeta crédito Tarjeta débito Transferencia Bre-B Nequi Otra transferencia Crédito cliente | no |
| 99 | input/number | entrada | mixedAmount1 | 0 | no |
| 100 | input/- | entrada | mixedReference1 | mixedReference1 | no |
| 101 | select/- | entrada | mixedMethod2 | Tarjeta débito Tarjeta crédito Transferencia Bre-B Nequi Otra transferencia Efectivo Crédito cliente | no |
| 102 | input/number | entrada | mixedAmount2 | 0 | no |
| 103 | input/- | entrada | mixedReference2 | mixedReference2 | no |
| 104 | select/- | entrada | mixedMethod3 | Tarjeta crédito Tarjeta débito Transferencia Bre-B Nequi Otra transferencia Efectivo Crédito cliente | no |
| 105 | input/number | entrada | mixedAmount3 | 0 | no |
| 106 | input/- | entrada | mixedReference3 | mixedReference3 | no |
| 107 | button/button | accion | btnActivarSesionCarrito | Activar carrito | no |
| 108 | input/text | entrada | stationPayDisplayEfectivo | stationPayDisplayEfectivo | sí |
| 109 | input/text | entrada | stationPayDisplayCredito | stationPayDisplayCredito | sí |
| 110 | input/text | entrada | stationPayDisplayDebito | stationPayDisplayDebito | sí |
| 111 | input/text | entrada | stationPayDisplayBreb | stationPayDisplayBreb | sí |
| 112 | input/text | entrada | stationPayDisplayNequi | stationPayDisplayNequi | sí |
| 113 | input/text | entrada | stationPayDisplayTransferenciaOtro | stationPayDisplayTransferenciaOtro | sí |
| 114 | input/text | entrada | stationPayDisplayCreditoCliente | stationPayDisplayCreditoCliente | sí |
| 115 | input/text | entrada | stationCashReceivedDisplay | stationCashReceivedDisplay | no |
| 116 | input/checkbox | accion | paymentQrEnabledCheck | paymentQrEnabledCheck | no |
| 117 | button/button | accion | paymentQrGenerateBtn | Generar QR | no |
| 118 | button/button | accion | paymentQrUseBtn | Usar como pago | no |
| 119 | select/- | entrada | paymentQrAccountSelect | paymentQrAccountSelect | no |
| 120 | input/- | entrada | commissionWorker | commissionWorker | no |
| 121 | button/button | accion | btnVipCodigo | Generar codigo para cliente vip | no |
| 122 | input/hidden | entrada | itemId | itemId | no |
| 123 | input/- | entrada | itemBuscarCatalogo | itemBuscarCatalogo | no |
| 124 | button/button | accion | buscarCatalogoBtn | Buscar catálogo | no |
| 125 | select/- | entrada | itemBusquedaResultados | itemBusquedaResultados | no |
| 126 | button/button | accion | aplicarCatalogoBtn | Aplicar selección | no |
| 127 | button/button | accion | limpiarCatalogoBtn | Quitar referencia | no |
| 128 | select/- | entrada | itemTipo | Producto Receta Servicio Otro | no |
| 129 | input/number | entrada | itemReferenciaID | itemReferenciaID | no |
| 130 | input/- | entrada | itemCodigo | itemCodigo | no |
| 131 | input/- | entrada | itemDescripcion | itemDescripcion | no |
| 132 | input/- | entrada | itemUnidad | unidad | no |
| 133 | input/number | entrada | itemCantidad | 1 | no |
| 134 | input/number | entrada | itemPrecio | 0 | no |
| 135 | input/number | entrada | itemDescuentoPct | 0 | no |
| 136 | input/number | entrada | itemImpuestoPct | 0 | no |
| 137 | input/- | entrada | itemImpuestoCodigo | IVA | no |
| 138 | textarea/- | entrada | itemObservaciones | itemObservaciones | no |
| 139 | button/submit | accion | saveItemBtn | Guardar item | no |
| 140 | button/button | accion | cancelItemBtn | Cancelar | no |
| 141 | button/button | accion | btnPagarCarrito | Pagar y cerrar carrito | sí |
| 142 | button/button | accion | retryInitialLoadBtn | Reintentar carga | no |
| 143 | button/button | accion | backToStationsBtn | Regresar a ' + escapeHtml(stationTerm(false)) + ' | no |
| 144 | button/- | accion | ' + item.id + ' | Abrir | sí |
| 145 | button/- | accion | ' + item.id + ' | Buscar productos | sí |
| 146 | button/- | accion | ' + item.id + ' | Editar | sí |
| 147 | button/- | accion | ' + item.id + ' | Eliminar | sí |
| 148 | button/- | accion | ' + item.id + ' | ' + toggleEstadoLabel + ' | sí |
| 149 | button/- | accion | ' + item.id + ' | ' + toggleOperacionLabel + ' | sí |
| 150 | button/button | accion | ' + escapeHtml(rele.id \|\| 0) + ' | ' + ' ' + escapeHtml(controlElectricoIcon(rele.tipo_carga)) + ' ' + escapeHtml(domoticaDeviceLabel(rele)) + ' ' + ' ' +  | sí |
| 151 | input/checkbox | accion | domoticaScheduleEnabled | domoticaScheduleEnabled | no |
| 152 | input/time | entrada | domoticaScheduleOn | ' + escapeHtml(normalize(rele.hora_encendido)) + ' | no |
| 153 | input/time | entrada | domoticaScheduleOff | ' + escapeHtml(normalize(rele.hora_apagado)) + ' | no |
| 154 | select/- | entrada | domoticaScheduleDays | ' + ' Todos ' + ' Lunes a viernes ' + ' Sabado y domingo ' + ' | no |
| 155 | input/- | entrada | domoticaScheduleTimezone | ' + escapeHtml(normalize(rele.programacion_timezone) \|\| 'America/Bogota') + ' | no |
| 156 | button/button | accion | ' + Number(item.id \|\| 0) + ' | - | sí |
| 157 | input/number | entrada | ' + Number(item.id \|\| 0) + ' | ' + qtyValue + ' | sí |
| 158 | button/button | accion | ' + Number(item.id \|\| 0) + ' | + | sí |
| 159 | button/- | accion | ' + item.id + ' | Devolver | sí |

### `web/administrar_empresa/carrito_historial_productos.html` (3)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnRegresarCarrito | Regresar al carrito | no |
| 2 | button/button | accion | btnActualizar | Actualizar | no |
| 3 | button/button | accion | btnImprimir | Imprimir | no |

### `web/administrar_empresa/carta_productos_publica.html` (37)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnRefresh | Actualizar | no |
| 2 | button/button | accion | btnPreview | Visualizar carta publica | no |
| 3 | button/button | accion | btnCopyUrl | Copiar | no |
| 4 | select/- | entrada | qrExportSize | 512 px 1024 px 2048 px | no |
| 5 | input/- | entrada | qrFileName | qrFileName | no |
| 6 | button/button | accion | btnDownloadQrPng | Descargar PNG | no |
| 7 | button/button | accion | btnDownloadQrSvg | Descargar SVG | no |
| 8 | button/button | accion | btnDownloadQrPdf | Descargar PDF | no |
| 9 | button/button | accion | btnPrintQr | Imprimir QR | no |
| 10 | input/- | entrada | empresaSlug | empresaSlug | no |
| 11 | input/- | entrada | moneda | moneda | no |
| 12 | input/- | entrada | nombreTienda | nombreTienda | no |
| 13 | textarea/- | entrada | descripcionTienda | descripcionTienda | no |
| 14 | input/file | accion | perfilUpload | perfilUpload | no |
| 15 | button/button | accion | btnUploadPerfil | Subir perfil | no |
| 16 | input/- | entrada | logoUrl | logoUrl | no |
| 17 | input/file | accion | portadaUpload | portadaUpload | no |
| 18 | button/button | accion | btnUploadPortada | Subir portada | no |
| 19 | input/- | entrada | bannerUrl | bannerUrl | no |
| 20 | input/color | entrada | colorPrimario | #0f766e | no |
| 21 | input/- | entrada | dominioPublico | dominioPublico | no |
| 22 | input/checkbox | accion | mostrarStock | mostrarStock | no |
| 23 | input/checkbox | accion | contactoFormularioActivo | contactoFormularioActivo | no |
| 24 | input/checkbox | accion | whatsappFlotanteActivo | whatsappFlotanteActivo | no |
| 25 | input/- | entrada | whatsappNumero | whatsappNumero | no |
| 26 | input/- | entrada | paginaSlug | paginaSlug | no |
| 27 | input/- | entrada | paginaNombre | paginaNombre | no |
| 28 | textarea/- | entrada | paginaDescripcion | paginaDescripcion | no |
| 29 | button/button | accion | btnSaveConfig | Guardar configuracion | no |
| 30 | input/- | entrada | searchProducts | searchProducts | no |
| 31 | select/- | entrada | filterState | Todos Publicados Sin publicar Ocultos | no |
| 32 | button/- | accion | ' + product.id + ' | Actualizar | sí |
| 33 | button/- | accion | - | Ocultar | sí |
| 34 | button/- | accion | - | Activar | sí |
| 35 | button/- | accion | ' + product.id + ' | Actualizar | sí |
| 36 | button/- | accion | ' + product.id + ' | Publicar | sí |
| 37 | button/- | accion | - | Imprimir | sí |

### `web/administrar_empresa/centro_ia_empresarial.html` (8)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/date | entrada | desde | desde | no |
| 2 | input/date | entrada | hasta | hasta | no |
| 3 | button/button | accion | btnActualizar | Actualizar | no |
| 4 | input/checkbox | accion | agentMode | agentMode | no |
| 5 | textarea/- | entrada | consulta | consulta | no |
| 6 | button/button | accion | btnDiagnostico | Diagnostico ERP | sí |
| 7 | button/button | accion | btnCustom | Analizar con IA | sí |
| 8 | button/button | accion | - | Ejecutar IA | sí |

### `web/administrar_empresa/centros_costo.html` (43)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/- | entrada | ccPeriodo | ccPeriodo | no |
| 2 | button/button | accion | ccRefresh | Actualizar | no |
| 3 | button/button | accion | ccSeed | Cargar demo | no |
| 4 | button/button | accion | ccExport | Exportar CSV | no |
| 5 | button/button | accion | - | Dashboard | sí |
| 6 | button/button | accion | - | Centros | sí |
| 7 | button/button | accion | - | Presupuesto | sí |
| 8 | button/button | accion | - | Reglas | sí |
| 9 | button/button | accion | - | Movimientos | sí |
| 10 | input/hidden | entrada | centroId | centroId | no |
| 11 | input/- | entrada | centroCodigo | centroCodigo | no |
| 12 | select/- | entrada | centroTipo | Operativo Sucursal Area Unidad negocio Proyecto Administrativo | no |
| 13 | input/- | entrada | centroNombre | centroNombre | no |
| 14 | input/- | entrada | centroSucursal | centroSucursal | no |
| 15 | input/- | entrada | centroArea | centroArea | no |
| 16 | input/- | entrada | centroUnidad | centroUnidad | no |
| 17 | input/- | entrada | centroResponsable | centroResponsable | no |
| 18 | input/number | entrada | centroMeta | 0 | no |
| 19 | select/- | entrada | centroEstado | Activo Inactivo Cerrado Suspendido | no |
| 20 | textarea/- | entrada | centroObs | centroObs | no |
| 21 | button/submit | accion | - | Guardar centro | no |
| 22 | input/hidden | entrada | presId | presId | no |
| 23 | select/- | entrada | presCentro | presCentro | no |
| 24 | input/- | entrada | presPeriodo | presPeriodo | no |
| 25 | input/number | entrada | presIngresos | 0 | no |
| 26 | input/number | entrada | presEgresos | 0 | no |
| 27 | input/number | entrada | presMeta | 0 | no |
| 28 | input/- | entrada | presResponsable | presResponsable | no |
| 29 | button/submit | accion | - | Guardar presupuesto | no |
| 30 | input/hidden | entrada | reglaId | reglaId | no |
| 31 | select/- | entrada | reglaCentro | reglaCentro | no |
| 32 | select/- | entrada | reglaOrigen | General Contabilidad Tesoreria Compras IA compras AIU | no |
| 33 | input/- | entrada | reglaNombre | reglaNombre | no |
| 34 | input/- | entrada | reglaCategoria | reglaCategoria | no |
| 35 | input/- | entrada | reglaCuenta | reglaCuenta | no |
| 36 | input/- | entrada | reglaTercero | reglaTercero | no |
| 37 | input/number | entrada | reglaPorcentaje | 100 | no |
| 38 | input/number | entrada | reglaPrioridad | 100 | no |
| 39 | select/- | entrada | reglaActiva | Si No | no |
| 40 | button/submit | accion | - | Guardar regla | no |
| 41 | button/button | accion | - | Editar | sí |
| 42 | button/button | accion | - | Editar | sí |
| 43 | button/button | accion | - | Editar | sí |

### `web/administrar_empresa/chat_con_inteligencia_artificial.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | openCentralAI | Abrir asistente IA central | no |

### `web/administrar_empresa/chat_tareas_chat_usuarios.html` (3)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | chatUsuariosRefreshBtn | Actualizar | no |
| 2 | textarea/- | entrada | chatUsuariosInput | chatUsuariosInput | no |
| 3 | button/submit | accion | chatUsuariosSendBtn | Enviar mensaje | no |

### `web/administrar_empresa/chat_tareas_papelera.html` (3)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | select/- | entrada | tipo | Conversaciones Mensajes Tareas Agenda / Citas | no |
| 2 | button/button | accion | reloadBtn | Actualizar | no |
| 3 | button/- | accion | '+esc(it.id)+' | Restaurar | sí |

### `web/administrar_empresa/chat_y_tareas.html` (69)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | reloadBtn | Actualizar todo | no |
| 2 | button/button | accion | quickConversationBtn | Nueva conversación | no |
| 3 | button/button | accion | quickTaskBtn | Nueva tarea | no |
| 4 | button/button | accion | quickAppointmentBtn | Programar reunión | no |
| 5 | button/button | accion | calendarPrevBtn | Mes anterior | no |
| 6 | button/button | accion | calendarNextBtn | Mes siguiente | no |
| 7 | button/button | accion | calendarTodayBtn | Hoy | no |
| 8 | input/hidden | entrada | appointmentId | appointmentId | no |
| 9 | input/- | entrada | appointmentTitle | appointmentTitle | no |
| 10 | textarea/- | entrada | appointmentDescription | appointmentDescription | no |
| 11 | select/- | entrada | appointmentType | Reunión Llamada Seguimiento Capacitación Otro | no |
| 12 | select/- | entrada | appointmentEstado | Programada Completada Cancelada | no |
| 13 | input/datetime-local | entrada | appointmentStart | appointmentStart | no |
| 14 | input/datetime-local | entrada | appointmentEnd | appointmentEnd | no |
| 15 | input/number | entrada | appointmentReminder | 30 | no |
| 16 | select/- | entrada | appointmentConversation | Sin conversación asociada | no |
| 17 | input/- | entrada | appointmentLocation | appointmentLocation | no |
| 18 | button/submit | accion | saveAppointmentBtn | Guardar cita | no |
| 19 | button/button | accion | clearAppointmentBtn | Limpiar | no |
| 20 | select/- | entrada | appointmentEstadoFilter | Todas Programadas Completadas Canceladas | no |
| 21 | input/- | entrada | appointmentSearch | appointmentSearch | no |
| 22 | button/button | accion | appointmentSearchBtn | Filtrar | no |
| 23 | input/- | entrada | convTitulo | convTitulo | no |
| 24 | textarea/- | entrada | convDescripcion | convDescripcion | no |
| 25 | input/- | entrada | convParticipantSearch | convParticipantSearch | no |
| 26 | select/- | entrada | convPrioridad | Media Alta Urgente Baja | no |
| 27 | button/submit | accion | createConvBtn | Crear | no |
| 28 | input/- | entrada | convSearch | convSearch | no |
| 29 | button/button | accion | convSearchBtn | Buscar | no |
| 30 | input/- | entrada | participantSearch | participantSearch | no |
| 31 | button/button | accion | addParticipantBtn | Agregar seleccionados | no |
| 32 | textarea/- | entrada | messageText | messageText | no |
| 33 | input/file | accion | messageFile | messageFile | no |
| 34 | button/submit | accion | sendMessageBtn | Enviar | no |
| 35 | button/button | accion | recordMessageVoiceBtn | Grabar voz | no |
| 36 | button/button | accion | stopMessageVoiceBtn | Detener | no |
| 37 | button/button | accion | clearMessageVoiceBtn | Limpiar voz | no |
| 38 | input/hidden | entrada | taskId | taskId | no |
| 39 | input/- | entrada | taskTitulo | taskTitulo | no |
| 40 | textarea/- | entrada | taskDescripcion | taskDescripcion | no |
| 41 | select/- | entrada | taskPrioridad | Media Alta Urgente Baja | no |
| 42 | input/datetime-local | entrada | taskFechaLimite | taskFechaLimite | no |
| 43 | select/- | entrada | taskAsignado | taskAsignado | no |
| 44 | button/submit | accion | saveTaskBtn | Guardar | no |
| 45 | button/button | accion | recordTaskVoiceBtn | Grabar voz tarea | no |
| 46 | button/button | accion | stopTaskVoiceBtn | Detener | no |
| 47 | button/button | accion | clearTaskVoiceBtn | Limpiar voz | no |
| 48 | select/- | entrada | taskEstadoFilter | Todas Pendientes En progreso Bloqueadas Completadas Canceladas | no |
| 49 | input/- | entrada | taskSearch | taskSearch | no |
| 50 | button/button | accion | taskSearchBtn | Filtrar | no |
| 51 | button/button | accion | - | ' + escapeHtml(actionLabel) + ' | sí |
| 52 | input/checkbox | accion | - | ' + option.id + ' | no |
| 53 | input/checkbox | accion | - | ' + option.id + ' | no |
| 54 | button/button | accion | ' + id + ' | '; html += ' ' + escapeHtml(c.titulo \|\| "Sin titulo") + ' '; html += ' ' + escapeHtml((c.prioridad \|\| "media").toUpperCa | sí |
| 55 | button/button | accion | ' + id + ' | ' + (status === "abierta" ? "Cerrar" : "Reabrir") + ' | sí |
| 56 | button/button | accion | ' + id + ' | ' + (estado === "activo" ? "Desactivar" : "Activar") + ' | sí |
| 57 | button/button | accion | ' + id + ' | Eliminar | sí |
| 58 | button/button | accion | ' + Number(t.id) + ' | Iniciar | sí |
| 59 | button/button | accion | ' + Number(t.id) + ' | Completar | sí |
| 60 | button/button | accion | ' + Number(t.id) + ' | Reabrir | sí |
| 61 | button/button | accion | ' + Number(t.id) + ' | Cancelar | sí |
| 62 | button/button | accion | ' + Number(t.id) + ' | Eliminar | sí |
| 63 | button/button | accion | ' + Number(cita.id \|\| 0) + ' | '; html += escapeHtml(formatHourMinute(cita.fecha_inicio) + ' ' + normalize(cita.titulo \|\| 'Cita')); html += ' | sí |
| 64 | button/button | accion | ' + id + ' | Editar | sí |
| 65 | button/button | accion | ' + id + ' | Completar | sí |
| 66 | button/button | accion | ' + id + ' | Cancelar | sí |
| 67 | button/button | accion | ' + id + ' | Reprogramar | sí |
| 68 | button/button | accion | ' + id + ' | ' + (estado === "activo" ? "Desactivar" : "Activar") + ' | sí |
| 69 | button/button | accion | ' + id + ' | Eliminar | sí |

### `web/administrar_empresa/cierre_fiscal.html` (49)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | cfRefresh | Actualizar | no |
| 2 | button/button | accion | cfSeed | Cargar demo | no |
| 3 | button/button | accion | cfExport | Exportar CSV | no |
| 4 | button/button | accion | - | Dashboard | sí |
| 5 | button/button | accion | - | Periodos | sí |
| 6 | button/button | accion | - | Politicas | sí |
| 7 | button/button | accion | - | Excepciones | sí |
| 8 | button/button | accion | - | Validar | sí |
| 9 | button/button | accion | - | Bitacora | sí |
| 10 | input/hidden | entrada | periodoId | periodoId | no |
| 11 | input/- | entrada | periodoCodigo | periodoCodigo | no |
| 12 | select/- | entrada | periodoEstado | Abierto En revision Cerrado Bloqueado | no |
| 13 | input/date | entrada | periodoDesde | periodoDesde | no |
| 14 | input/date | entrada | periodoHasta | periodoHasta | no |
| 15 | select/- | entrada | periodoTipo | Mensual Bimestral Trimestral Anual Manual | no |
| 16 | input/- | entrada | periodoMotivo | periodoMotivo | no |
| 17 | input/checkbox | accion | bloqVentas | bloqVentas | no |
| 18 | input/checkbox | accion | bloqCompras | bloqCompras | no |
| 19 | input/checkbox | accion | bloqCaja | bloqCaja | no |
| 20 | input/checkbox | accion | bloqInventario | bloqInventario | no |
| 21 | input/checkbox | accion | bloqContabilidad | bloqContabilidad | no |
| 22 | input/checkbox | accion | bloqFacturacion | bloqFacturacion | no |
| 23 | button/submit | accion | - | Guardar periodo | no |
| 24 | select/- | entrada | polModulo | Ventas Compras Caja/Tesoreria Inventario Contabilidad Facturacion | no |
| 25 | input/- | entrada | polNombre | polNombre | no |
| 26 | input/number | entrada | polDias | 30 | no |
| 27 | select/- | entrada | polEstado | Activo Inactivo | no |
| 28 | input/checkbox | accion | polAuto | polAuto | no |
| 29 | input/checkbox | accion | polReapertura | polReapertura | no |
| 30 | input/checkbox | accion | polExcepciones | polExcepciones | no |
| 31 | input/checkbox | accion | polNotificar | polNotificar | no |
| 32 | button/submit | accion | - | Guardar politica | no |
| 33 | input/- | entrada | exPeriodo | exPeriodo | no |
| 34 | select/- | entrada | exModulo | Todos Ventas Compras Caja Inventario Contabilidad Facturacion | no |
| 35 | select/- | entrada | exAccion | Actualizar Crear Anular Contabilizar Emitir Todas | no |
| 36 | input/date | entrada | exExpira | exExpira | no |
| 37 | input/- | entrada | exDocTipo | exDocTipo | no |
| 38 | input/number | entrada | exDocId | 0 | no |
| 39 | textarea/- | entrada | exMotivo | exMotivo | no |
| 40 | button/submit | accion | - | Crear excepcion | no |
| 41 | input/date | entrada | valFecha | valFecha | no |
| 42 | select/- | entrada | valModulo | Ventas Compras Caja Inventario Contabilidad Facturacion | no |
| 43 | select/- | entrada | valAccion | Crear Actualizar Anular Emitir Contabilizar | no |
| 44 | input/- | entrada | valDocTipo | valDocTipo | no |
| 45 | button/submit | accion | - | Validar operacion | no |
| 46 | button/- | accion | - | Editar | sí |
| 47 | button/- | accion | - | Cerrar | sí |
| 48 | button/- | accion | - | Reabrir | sí |
| 49 | button/- | accion | - | Editar | sí |

### `web/administrar_empresa/cobranza.html` (54)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnDemo | Demo | no |
| 2 | button/button | accion | btnExportar | Exportar CSV | no |
| 3 | button/button | accion | btnRefrescar | Actualizar | no |
| 4 | input/checkbox | accion | auto_activo | auto_activo | no |
| 5 | input/checkbox | accion | email_activo | email_activo | no |
| 6 | input/checkbox | accion | whatsapp_activo | whatsapp_activo | no |
| 7 | input/number | entrada | dias_antes | 1 | no |
| 8 | input/number | entrada | frecuencia_dias | 3 | no |
| 9 | input/time | entrada | hora_local | 09:00 | no |
| 10 | input/- | entrada | auto_asunto | Recordatorio de pago | no |
| 11 | textarea/- | entrada | auto_mensaje | Hola {{cliente}}, el documento {{documento}} registra un saldo de {{saldo}} con vencimiento {{vencimiento}}. | sí |
| 12 | button/button | accion | btnGuardarAuto | Guardar configuracion | no |
| 13 | button/button | accion | btnProbarAuto | Probar sin enviar | no |
| 14 | button/button | accion | btnEjecutarAuto | Ejecutar ahora | no |
| 15 | input/- | entrada | filtro_q | filtro_q | no |
| 16 | select/- | entrada | filtro_estado | Todas Pendiente Vencida Pagada Castigada | no |
| 17 | input/number | entrada | filtro_mora | 0 | no |
| 18 | button/button | accion | btnFiltrar | Filtrar | no |
| 19 | input/hidden | entrada | cuenta_id | cuenta_id | no |
| 20 | input/- | entrada | cliente_nombre | cliente_nombre | no |
| 21 | input/- | entrada | documento_codigo | documento_codigo | no |
| 22 | select/- | entrada | canal | Llamada WhatsApp Email SMS | no |
| 23 | select/- | entrada | resultado | Contactado Sin respuesta Promesa de pago Reprogramado Escalado Fallido | no |
| 24 | input/datetime-local | entrada | fecha_proximo_contacto | fecha_proximo_contacto | no |
| 25 | input/number | entrada | valor_compromiso | 0 | no |
| 26 | input/date | entrada | promesa_fecha | promesa_fecha | no |
| 27 | input/- | entrada | contacto | contacto | no |
| 28 | textarea/- | entrada | mensaje | mensaje | no |
| 29 | textarea/- | entrada | observaciones | observaciones | no |
| 30 | button/submit | accion | - | Registrar gestion | no |
| 31 | button/button | accion | btnSimularEnvio | Simular envio | no |
| 32 | button/button | accion | - | Campanas | sí |
| 33 | button/button | accion | - | Plantillas | sí |
| 34 | button/button | accion | - | Promesas | sí |
| 35 | input/- | entrada | campana_nombre | campana_nombre | no |
| 36 | select/- | entrada | campana_tipo | Preventiva Recuperacion Juridica Masiva VIP | no |
| 37 | select/- | entrada | campana_canal | WhatsApp Email SMS Llamada | no |
| 38 | select/- | entrada | campana_estado | Borrador Activa Pausada Finalizada | no |
| 39 | input/- | entrada | campana_segmento | campana_segmento | no |
| 40 | input/number | entrada | campana_meta | 0 | no |
| 41 | input/date | entrada | campana_inicio | campana_inicio | no |
| 42 | input/date | entrada | campana_fin | campana_fin | no |
| 43 | button/submit | accion | - | Guardar campana | no |
| 44 | input/- | entrada | plantilla_nombre | plantilla_nombre | no |
| 45 | select/- | entrada | plantilla_canal | WhatsApp Email SMS Llamada | no |
| 46 | input/number | entrada | plantilla_desde | 0 | no |
| 47 | input/number | entrada | plantilla_hasta | 9999 | no |
| 48 | input/number | entrada | plantilla_prioridad | 1 | no |
| 49 | input/- | entrada | plantilla_asunto | plantilla_asunto | no |
| 50 | textarea/- | entrada | plantilla_cuerpo | plantilla_cuerpo | sí |
| 51 | button/submit | accion | - | Guardar plantilla | no |
| 52 | button/button | accion | - | Gestionar | sí |
| 53 | button/button | accion | - | Cumplida | sí |
| 54 | button/button | accion | - | Incumplida | sí |

### `web/administrar_empresa/codigos_de_descuento.html` (28)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/- | accion | addCodigoBtn | Nuevo código | no |
| 2 | input/hidden | entrada | codigoId | codigoId | no |
| 3 | input/hidden | entrada | empresaID | empresaID | no |
| 4 | input/- | entrada | codigoTexto | codigoTexto | no |
| 5 | button/button | accion | generarCodigoBtn | Generar automático | no |
| 6 | select/- | entrada | tipoDescuento | Valor fijo Porcentaje | no |
| 7 | input/number | entrada | valorDescuento | 0 | no |
| 8 | select/- | entrada | monedaCodigo | COP USD EUR | no |
| 9 | input/number | entrada | montoMinimo | 0 | no |
| 10 | input/number | entrada | usosMaximos | 1 | no |
| 11 | input/date | entrada | fechaVencimiento | fechaVencimiento | no |
| 12 | select/- | entrada | segmentoCliente | Todos Nuevo Frecuente Recurrente En riesgo VIP | no |
| 13 | select/- | entrada | canalVentaCodigo | Todos Mostrador App Web Kiosko Domicilio Teléfono | no |
| 14 | input/time | entrada | horarioDesdeCodigo | horarioDesdeCodigo | no |
| 15 | input/time | entrada | horarioHastaCodigo | horarioHastaCodigo | no |
| 16 | input/- | entrada | diasSemanaCodigo | diasSemanaCodigo | no |
| 17 | input/number | entrada | maxUsosPorCliente | 0 | no |
| 18 | input/number | entrada | ventanaHorasFraude | 24 | no |
| 19 | textarea/- | entrada | observacionesCodigo | observacionesCodigo | no |
| 20 | button/submit | accion | saveCodigoBtn | Guardar código | no |
| 21 | button/button | accion | cancelCodigoBtn | Cancelar | no |
| 22 | input/- | entrada | buscarCodigo | buscarCodigo | no |
| 23 | button/button | accion | buscarCodigoBtn | Buscar | no |
| 24 | button/- | accion | ' + Number(item.id \|\| 0) + ' | Editar | sí |
| 25 | button/- | accion | ' + Number(item.id \|\| 0) + ' | Enviar correo | sí |
| 26 | button/- | accion | ' + Number(item.id \|\| 0) + ' | WhatsApp | sí |
| 27 | button/- | accion | ' + Number(item.id \|\| 0) + ' | ' + toggleText + ' | sí |
| 28 | button/- | accion | ' + Number(item.id \|\| 0) + ' | Eliminar | sí |

### `web/administrar_empresa/comisiones.html` (54)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/checkbox | accion | cfgHabilitar | cfgHabilitar | no |
| 2 | input/number | entrada | cfgPorcentaje | 10 | no |
| 3 | input/- | entrada | cfgFiltroServicio | lavado | no |
| 4 | input/checkbox | accion | cfgAplicarAuto | cfgAplicarAuto | no |
| 5 | textarea/- | entrada | cfgObservaciones | cfgObservaciones | no |
| 6 | button/button | accion | btnGuardarCfg | Guardar configuración | no |
| 7 | button/button | accion | btnActivarCfg | Activar | no |
| 8 | button/button | accion | btnDesactivarCfg | Desactivar | no |
| 9 | input/hidden | entrada | escalaId | 0 | no |
| 10 | input/- | entrada | escalaRol | escalaRol | no |
| 11 | input/- | entrada | escalaServicio | escalaServicio | no |
| 12 | input/number | entrada | escalaPct | 0 | no |
| 13 | input/number | entrada | escalaTope | 0 | no |
| 14 | input/number | entrada | escalaPrioridad | 100 | no |
| 15 | input/- | entrada | escalaObs | escalaObs | no |
| 16 | button/button | accion | btnGuardarEscala | Guardar escala | no |
| 17 | button/button | accion | btnLimpiarEscala | Limpiar | no |
| 18 | button/button | accion | btnRefrescarEscalas | Actualizar escalas | no |
| 19 | input/- | entrada | ajUsuarioLavador | ajUsuarioLavador | no |
| 20 | input/- | entrada | ajRolOperacion | ajRolOperacion | no |
| 21 | input/number | entrada | ajMonto | ajMonto | no |
| 22 | input/number | entrada | ajBase | 0 | no |
| 23 | input/number | entrada | ajCarrito | ajCarrito | no |
| 24 | input/number | entrada | ajItem | ajItem | no |
| 25 | input/- | entrada | ajServicioCodigo | ajServicioCodigo | no |
| 26 | input/- | entrada | ajServicioNombre | ajServicioNombre | no |
| 27 | input/- | entrada | ajVentaRef | ajVentaRef | no |
| 28 | input/- | entrada | ajReferencia | ajReferencia | no |
| 29 | input/- | entrada | ajMoneda | COP | no |
| 30 | input/- | entrada | ajServicioCategoria | ajServicioCategoria | no |
| 31 | textarea/- | entrada | ajMotivo | ajMotivo | no |
| 32 | button/button | accion | btnRegistrarAjuste | Registrar ajuste | no |
| 33 | input/date | entrada | fDesde | fDesde | no |
| 34 | input/date | entrada | fHasta | fHasta | no |
| 35 | input/- | entrada | fUsuarioLavador | fUsuarioLavador | no |
| 36 | input/- | entrada | fRolOperacion | fRolOperacion | no |
| 37 | input/- | entrada | fServicioFiltro | fServicioFiltro | no |
| 38 | select/- | entrada | fOrigen | Todos Venta Ajuste manual | no |
| 39 | select/- | entrada | fAjusteEstado | Todos Pendiente Aprobado Rechazado | no |
| 40 | input/number | entrada | fLimit | 200 | no |
| 41 | input/checkbox | accion | fSoloAjustes | fSoloAjustes | no |
| 42 | input/checkbox | accion | fSoloPendientes | fSoloPendientes | no |
| 43 | input/checkbox | accion | fNoLiquidado | fNoLiquidado | no |
| 44 | input/checkbox | accion | fIncludeInactive | fIncludeInactive | no |
| 45 | button/button | accion | btnRefrescarReporte | Actualizar reporte | no |
| 46 | input/date | entrada | liqDesde | liqDesde | no |
| 47 | input/date | entrada | liqHasta | liqHasta | no |
| 48 | input/- | entrada | liqLavador | liqLavador | no |
| 49 | input/- | entrada | liqEmpleadoNombre | liqEmpleadoNombre | no |
| 50 | button/button | accion | btnResumenLiquidacion | Consultar resumen | no |
| 51 | button/button | accion | ' + toInt(item.id) + ' | Editar | sí |
| 52 | button/button | accion | ' + toInt(item.id) + ' | ' + (isActive ? 'Desactivar' : 'Activar') + ' | sí |
| 53 | button/button | accion | ' + toInt(item.id) + ' | Aprobar | sí |
| 54 | button/button | accion | ' + toInt(item.id) + ' | Rechazar | sí |

### `web/administrar_empresa/compras.html` (34)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | btnGestionarProveedores | Gestionar proveedores | no |
| 2 | select/- | entrada | proveedorSelect | Seleccione proveedor | no |
| 3 | select/- | entrada | accionCreate | Guardar como borrador Enviar a aprobacion Emitir orden de compra | no |
| 4 | input/- | entrada | documentoCodigo | documentoCodigo | no |
| 5 | input/number | entrada | montoTotal | 0 | no |
| 6 | input/- | entrada | moneda | COP | no |
| 7 | input/- | entrada | periodoContable | periodoContable | no |
| 8 | input/checkbox | accion | requiereAprobacion | requiereAprobacion | no |
| 9 | input/number | entrada | nivelesAprobacion | 2 | no |
| 10 | input/- | entrada | proveedorDocRef | proveedorDocRef | no |
| 11 | input/- | entrada | facturaDocRef | facturaDocRef | no |
| 12 | input/- | entrada | entradaDocRef | entradaDocRef | no |
| 13 | input/file | accion | comprobanteArchivo | comprobanteArchivo | no |
| 14 | button/button | accion | btnAnalizarCompraIA | Carga automática de factura de compra | no |
| 15 | textarea/- | entrada | observaciones | observaciones | no |
| 16 | button/submit | accion | - | Guardar documento | no |
| 17 | button/button | accion | btnLimpiarForm | Limpiar | no |
| 18 | select/- | entrada | filtroEstado | Todos los estados Borrador Pendiente aprobacion Emitida Recepcion parcial Recepcionada Contabilizada Rechazada | no |
| 19 | input/- | entrada | filtroBusqueda | filtroBusqueda | no |
| 20 | input/checkbox | accion | filtroInactivos | filtroInactivos | no |
| 21 | button/button | accion | btnFiltrar | Filtrar | no |
| 22 | a/- | accion | - | ' + escapeHtml(nombre) + ' | no |
| 23 | button/button | accion | - | Solicitar aprob. | sí |
| 24 | button/button | accion | - | Emitir | sí |
| 25 | button/button | accion | - | Aprobar nivel | sí |
| 26 | button/button | accion | - | Rechazar | sí |
| 27 | button/button | accion | - | Recep. parcial | sí |
| 28 | button/button | accion | - | Recepcionar | sí |
| 29 | button/button | accion | - | Validar docs | sí |
| 30 | button/button | accion | - | Contabilizar | sí |
| 31 | button/button | accion | - | Validar docs | sí |
| 32 | button/button | accion | - | Adjuntar soporte | sí |
| 33 | button/button | accion | - | Desactivar | sí |
| 34 | button/button | accion | - | Activar | sí |

### `web/administrar_empresa/compras_avanzadas.html` (53)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnRefresh | Actualizar | no |
| 2 | button/button | accion | btnSeed | Cargar demo | no |
| 3 | a/- | accion | btnProveedores | Proveedores | no |
| 4 | button/button | accion | - | Requisicion | sí |
| 5 | button/button | accion | - | Cotizacion | sí |
| 6 | button/button | accion | - | Aprobar | sí |
| 7 | button/button | accion | - | Recepcion | sí |
| 8 | input/- | entrada | reqCodigo | reqCodigo | no |
| 9 | input/- | entrada | reqSolicitante | reqSolicitante | no |
| 10 | input/- | entrada | reqArea | reqArea | no |
| 11 | select/- | entrada | reqPrioridad | media alta urgente baja | no |
| 12 | input/- | entrada | reqCentroCosto | reqCentroCosto | no |
| 13 | input/date | entrada | reqFecha | reqFecha | no |
| 14 | input/date | entrada | reqNecesidad | reqNecesidad | no |
| 15 | select/- | entrada | reqEstado | solicitada borrador cotizando | no |
| 16 | textarea/- | entrada | reqJustificacion | reqJustificacion | no |
| 17 | input/- | entrada | itemNombre1 | itemNombre1 | no |
| 18 | input/number | entrada | itemCant1 | itemCant1 | no |
| 19 | input/number | entrada | itemCosto1 | itemCosto1 | no |
| 20 | select/- | entrada | itemProv1 | Cargando proveedores... | no |
| 21 | input/- | entrada | itemNombre2 | itemNombre2 | no |
| 22 | input/number | entrada | itemCant2 | itemCant2 | no |
| 23 | input/number | entrada | itemCosto2 | itemCosto2 | no |
| 24 | select/- | entrada | itemProv2 | Cargando proveedores... | no |
| 25 | button/button | accion | btnSaveReq | Guardar requisicion | no |
| 26 | input/number | entrada | cotReqID | cotReqID | no |
| 27 | select/- | entrada | cotProveedor | Cargando proveedores... | no |
| 28 | input/- | entrada | cotNumero | cotNumero | no |
| 29 | input/date | entrada | cotFecha | cotFecha | no |
| 30 | input/number | entrada | cotSubtotal | cotSubtotal | no |
| 31 | input/number | entrada | cotImpuestos | cotImpuestos | no |
| 32 | input/number | entrada | cotEntrega | cotEntrega | no |
| 33 | input/date | entrada | cotValidez | cotValidez | no |
| 34 | input/- | entrada | cotCondiciones | cotCondiciones | no |
| 35 | button/button | accion | btnSaveCot | Guardar cotizacion | no |
| 36 | input/number | entrada | aprReqID | aprReqID | no |
| 37 | input/number | entrada | aprCotID | aprCotID | no |
| 38 | select/- | entrada | aprDecision | aprobada rechazada pendiente | no |
| 39 | input/number | entrada | aprMonto | aprMonto | no |
| 40 | textarea/- | entrada | aprComentario | aprComentario | no |
| 41 | button/button | accion | btnSaveApr | Registrar decision | no |
| 42 | input/number | entrada | recReqID | recReqID | no |
| 43 | input/number | entrada | recCotID | recCotID | no |
| 44 | input/- | entrada | recDocumento | recDocumento | no |
| 45 | select/- | entrada | recEstado | parcial total | no |
| 46 | select/- | entrada | recProveedor | Cargando proveedores... | no |
| 47 | input/date | entrada | recFecha | recFecha | no |
| 48 | input/- | entrada | recProducto | recProducto | no |
| 49 | input/number | entrada | recItemID | recItemID | no |
| 50 | input/number | entrada | recOrdenada | recOrdenada | no |
| 51 | input/number | entrada | recRecibida | recRecibida | no |
| 52 | input/number | entrada | recCosto | recCosto | no |
| 53 | button/button | accion | btnSaveRec | Guardar recepcion | no |

### `web/administrar_empresa/compras_menu.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | - | &#9776; Ocultar menú | sí |

### `web/administrar_empresa/configuracion.html` (208)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnGuardarLogoEmpresa | Guardar logo | no |
| 2 | input/checkbox | accion | empresaMostrarLogoEmpresa | empresaMostrarLogoEmpresa | no |
| 3 | input/checkbox | accion | empresaMostrarLogoSistema | empresaMostrarLogoSistema | no |
| 4 | input/checkbox | accion | empresaMostrarLogoFacturaIdentidad | empresaMostrarLogoFacturaIdentidad | no |
| 5 | input/checkbox | accion | empresaMostrarLogoCarrito | empresaMostrarLogoCarrito | no |
| 6 | select/- | entrada | empresaLogoCarritoTamano | Pequeno Mediano Grande | no |
| 7 | input/file | accion | empresaLogoFile | empresaLogoFile | no |
| 8 | button/button | accion | btnSubirLogoEmpresa | Subir Archivo | no |
| 9 | input/- | entrada | empresaLogoURL | empresaLogoURL | no |
| 10 | a/- | accion | - | Abrir configuración de impresora | no |
| 11 | button/button | accion | btnGuardarConfig | Guardar configuración | no |
| 12 | input/checkbox | accion | imprimirOrdenServicio | imprimirOrdenServicio | no |
| 13 | input/checkbox | accion | habilitarDescuentos | habilitarDescuentos | no |
| 14 | input/checkbox | accion | permitirDescuentoPorcentaje | permitirDescuentoPorcentaje | no |
| 15 | input/checkbox | accion | permitirDescuentoCodigo | permitirDescuentoCodigo | no |
| 16 | input/checkbox | accion | permitirDescuentoValor | permitirDescuentoValor | no |
| 17 | input/checkbox | accion | habilitarLectorCodigoBarras | habilitarLectorCodigoBarras | no |
| 18 | input/checkbox | accion | autofocoLectorCodigoBarras | autofocoLectorCodigoBarras | no |
| 19 | input/checkbox | accion | acumularLectorCodigoBarras | acumularLectorCodigoBarras | no |
| 20 | input/- | entrada | areaDespacho | areaDespacho | no |
| 21 | input/number | entrada | copiasOrden | 1 | no |
| 22 | textarea/- | entrada | notaOrden | notaOrden | no |
| 23 | textarea/- | entrada | codigosDescuento | codigosDescuento | no |
| 24 | button/button | accion | btnGuardarProductoCamposObligatoriosConfig | Guardar campos | no |
| 25 | input/checkbox | accion | - | sin etiqueta | no |
| 26 | input/checkbox | accion | - | sin etiqueta | sí |
| 27 | input/checkbox | accion | - | sin etiqueta | sí |
| 28 | input/checkbox | accion | - | sin etiqueta | sí |
| 29 | input/checkbox | accion | - | sin etiqueta | sí |
| 30 | input/checkbox | accion | - | sin etiqueta | sí |
| 31 | input/checkbox | accion | - | sin etiqueta | sí |
| 32 | input/checkbox | accion | - | sin etiqueta | sí |
| 33 | input/checkbox | accion | - | sin etiqueta | sí |
| 34 | input/checkbox | accion | - | sin etiqueta | sí |
| 35 | input/checkbox | accion | - | sin etiqueta | sí |
| 36 | input/checkbox | accion | - | sin etiqueta | sí |
| 37 | input/checkbox | accion | - | sin etiqueta | sí |
| 38 | input/checkbox | accion | - | sin etiqueta | sí |
| 39 | input/checkbox | accion | - | sin etiqueta | sí |
| 40 | input/checkbox | accion | - | sin etiqueta | sí |
| 41 | input/checkbox | accion | - | sin etiqueta | sí |
| 42 | input/checkbox | accion | - | sin etiqueta | sí |
| 43 | input/checkbox | accion | - | sin etiqueta | sí |
| 44 | input/checkbox | accion | - | sin etiqueta | sí |
| 45 | input/checkbox | accion | - | sin etiqueta | sí |
| 46 | button/button | accion | btnGuardarConfigOperativa | Guardar empresa | no |
| 47 | input/checkbox | accion | opMetodoEfectivo | opMetodoEfectivo | no |
| 48 | input/checkbox | accion | opMetodoTarjetaCredito | opMetodoTarjetaCredito | no |
| 49 | input/checkbox | accion | opMetodoTarjetaDebito | opMetodoTarjetaDebito | no |
| 50 | input/checkbox | accion | opMetodoTransferencia | opMetodoTransferencia | no |
| 51 | input/checkbox | accion | opMetodoMixto | opMetodoMixto | no |
| 52 | input/checkbox | accion | opMetodoCodigoDescuento | opMetodoCodigoDescuento | no |
| 53 | input/checkbox | accion | opHabilitarPropinas | opHabilitarPropinas | no |
| 54 | input/checkbox | accion | opHabilitarComisiones | opHabilitarComisiones | no |
| 55 | select/- | entrada | opRolSelect | admin_empresa supervisor_sucursal cajero inventario compras contabilidad auditor | no |
| 56 | input/checkbox | accion | opRolActivo | opRolActivo | no |
| 57 | input/checkbox | accion | opRolMetodoEfectivo | opRolMetodoEfectivo | no |
| 58 | input/checkbox | accion | opRolMetodoTarjetaCredito | opRolMetodoTarjetaCredito | no |
| 59 | input/checkbox | accion | opRolMetodoTarjetaDebito | opRolMetodoTarjetaDebito | no |
| 60 | input/checkbox | accion | opRolMetodoTransferencia | opRolMetodoTransferencia | no |
| 61 | input/checkbox | accion | opRolMetodoMixto | opRolMetodoMixto | no |
| 62 | input/checkbox | accion | opRolMetodoCodigoDescuento | opRolMetodoCodigoDescuento | no |
| 63 | input/checkbox | accion | opRolHabilitarPropinas | opRolHabilitarPropinas | no |
| 64 | input/checkbox | accion | opRolHabilitarComisiones | opRolHabilitarComisiones | no |
| 65 | button/button | accion | btnGuardarConfigOperativaRol | Guardar rol | no |
| 66 | select/- | entrada | opPoliticaCanal | Todos mostrador app estacion reserva online delivery kiosko | no |
| 67 | input/number | entrada | opPoliticaSucursalId | 0 | no |
| 68 | input/- | entrada | opPoliticaTurno | opPoliticaTurno | no |
| 69 | input/number | entrada | opPoliticaPrioridad | 100 | no |
| 70 | input/checkbox | accion | opPoliticaActivo | opPoliticaActivo | no |
| 71 | input/checkbox | accion | opPoliticaMetodoEfectivo | opPoliticaMetodoEfectivo | no |
| 72 | input/checkbox | accion | opPoliticaMetodoTarjetaCredito | opPoliticaMetodoTarjetaCredito | no |
| 73 | input/checkbox | accion | opPoliticaMetodoTarjetaDebito | opPoliticaMetodoTarjetaDebito | no |
| 74 | input/checkbox | accion | opPoliticaMetodoTransferencia | opPoliticaMetodoTransferencia | no |
| 75 | input/checkbox | accion | opPoliticaMetodoMixto | opPoliticaMetodoMixto | no |
| 76 | input/checkbox | accion | opPoliticaMetodoCodigoDescuento | opPoliticaMetodoCodigoDescuento | no |
| 77 | input/checkbox | accion | opPoliticaHabilitarPropinas | opPoliticaHabilitarPropinas | no |
| 78 | input/checkbox | accion | opPoliticaHabilitarComisiones | opPoliticaHabilitarComisiones | no |
| 79 | button/button | accion | btnGuardarConfigOperativaPolitica | Guardar politica | no |
| 80 | select/- | entrada | opSimRol | admin_empresa supervisor_sucursal cajero inventario compras contabilidad auditor | no |
| 81 | input/- | entrada | opSimCanal | opSimCanal | no |
| 82 | input/number | entrada | opSimSucursalId | 0 | no |
| 83 | input/- | entrada | opSimTurno | opSimTurno | no |
| 84 | input/- | entrada | opSimObservaciones | opSimObservaciones | no |
| 85 | button/button | accion | btnSimularConfigOperativa | Simular | no |
| 86 | button/button | accion | btnSimularGuardarConfigOperativa | Simular y guardar | no |
| 87 | input/number | entrada | opRollbackHistorialId | opRollbackHistorialId | no |
| 88 | button/button | accion | btnRefrescarOperativaHistorial | Refrescar historial | no |
| 89 | button/button | accion | btnRollbackConfigOperativa | Aplicar rollback | no |
| 90 | button/button | accion | btnGuardarConfigReporte | Guardar configuracion | no |
| 91 | button/button | accion | btnRestaurarConfigReporte | Restaurar profesional | no |
| 92 | select/- | entrada | corteFormatoImpresion | Ticket POS 80mm Carta completo Ejecutivo compacto | no |
| 93 | input/checkbox | accion | - | sin etiqueta | sí |
| 94 | input/checkbox | accion | - | sin etiqueta | sí |
| 95 | input/checkbox | accion | - | sin etiqueta | sí |
| 96 | input/checkbox | accion | - | sin etiqueta | sí |
| 97 | input/checkbox | accion | - | sin etiqueta | sí |
| 98 | input/checkbox | accion | - | sin etiqueta | sí |
| 99 | input/checkbox | accion | - | sin etiqueta | sí |
| 100 | input/checkbox | accion | - | sin etiqueta | sí |
| 101 | input/checkbox | accion | - | sin etiqueta | sí |
| 102 | input/checkbox | accion | - | sin etiqueta | sí |
| 103 | input/checkbox | accion | - | sin etiqueta | sí |
| 104 | input/checkbox | accion | - | sin etiqueta | sí |
| 105 | input/checkbox | accion | - | sin etiqueta | sí |
| 106 | input/checkbox | accion | - | sin etiqueta | sí |
| 107 | input/checkbox | accion | - | sin etiqueta | sí |
| 108 | input/checkbox | accion | - | sin etiqueta | sí |
| 109 | input/checkbox | accion | - | sin etiqueta | sí |
| 110 | input/checkbox | accion | - | sin etiqueta | sí |
| 111 | input/checkbox | accion | - | sin etiqueta | sí |
| 112 | input/checkbox | accion | - | sin etiqueta | sí |
| 113 | input/checkbox | accion | - | sin etiqueta | sí |
| 114 | input/checkbox | accion | - | sin etiqueta | sí |
| 115 | input/checkbox | accion | - | sin etiqueta | sí |
| 116 | input/checkbox | accion | - | sin etiqueta | sí |
| 117 | input/checkbox | accion | - | sin etiqueta | sí |
| 118 | input/checkbox | accion | - | sin etiqueta | sí |
| 119 | input/checkbox | accion | - | sin etiqueta | sí |
| 120 | input/checkbox | accion | - | sin etiqueta | sí |
| 121 | input/checkbox | accion | - | sin etiqueta | sí |
| 122 | input/checkbox | accion | - | sin etiqueta | sí |
| 123 | input/checkbox | accion | - | sin etiqueta | sí |
| 124 | input/checkbox | accion | - | sin etiqueta | sí |
| 125 | input/checkbox | accion | - | sin etiqueta | sí |
| 126 | input/checkbox | accion | - | sin etiqueta | sí |
| 127 | input/checkbox | accion | - | sin etiqueta | sí |
| 128 | button/button | accion | btnGuardarCajaConfig | Guardar caja | no |
| 129 | input/checkbox | accion | cajaActiva | cajaActiva | no |
| 130 | input/checkbox | accion | cajasSimultaneasHabilitadas | cajasSimultaneasHabilitadas | no |
| 131 | input/text | entrada | cajaNombre | cajaNombre | no |
| 132 | input/text | entrada | cajaCodigo | cajaCodigo | no |
| 133 | input/number | entrada | maxCajasSimultaneasEmpresa | 0 | no |
| 134 | textarea/- | entrada | cajaObservaciones | cajaObservaciones | no |
| 135 | input/checkbox | accion | cajonMonederoHabilitado | cajonMonederoHabilitado | no |
| 136 | input/checkbox | accion | abrirCajonAlPagarCarrito | abrirCajonAlPagarCarrito | no |
| 137 | input/checkbox | accion | abrirCajonAlCerrarTransaccion | abrirCajonAlCerrarTransaccion | no |
| 138 | select/- | entrada | cajonMonederoMetodo | Impresion POS / driver de impresora Manual, solo aviso operativo | no |
| 139 | input/text | entrada | cajonMonederoImpresoraFuncionalidad | cajon_monedero | no |
| 140 | select/- | entrada | cajonMonederoComando | Pulso ESC/POS por impresora Apertura automatica del driver | no |
| 141 | button/button | accion | btnGuardarFacturaConfig | Guardar documento | no |
| 142 | select/- | entrada | facturaModoDocumentoVenta | No, solo generar la venta/comprobante Sí, generar además factura electrónica | no |
| 143 | select/- | entrada | facturaFormatoImpresion | Grande / carta Pequeña POS (tirilla) | no |
| 144 | input/checkbox | accion | facturaImprimirVenta | facturaImprimirVenta | no |
| 145 | input/checkbox | accion | facturaImprimirFacturaElectronica | facturaImprimirFacturaElectronica | no |
| 146 | input/number | entrada | facturaFuenteCarta | facturaFuenteCarta | no |
| 147 | input/number | entrada | facturaFuentePOS | facturaFuentePOS | no |
| 148 | input/checkbox | accion | facturaImprimirCopia | facturaImprimirCopia | no |
| 149 | input/file | accion | facturaLogoFile | facturaLogoFile | no |
| 150 | button/button | accion | btnSubirLogoFactura | Subir Archivo | no |
| 151 | input/- | entrada | facturaLogoURL | facturaLogoURL | no |
| 152 | input/checkbox | accion | facturaMostrarLogo | facturaMostrarLogo | no |
| 153 | button/button | accion | btnGuardarFormatoNumerico | Guardar formato | no |
| 154 | input/- | entrada | formatoMonedaCodigo | formatoMonedaCodigo | no |
| 155 | select/- | entrada | formatoSistemaNumerico | Latino (1.234,56) Internacional (1,234.56) | no |
| 156 | input/checkbox | accion | formatoUsarDecimales | formatoUsarDecimales | no |
| 157 | input/number | entrada | formatoCantidadDecimales | 2 | no |
| 158 | button/button | accion | btnGuardarPerfilTributarioCO | Guardar perfil | no |
| 159 | select/- | entrada | tributarioTipoPersonaCO | Seleccionar Persona natural Persona juridica | no |
| 160 | select/- | entrada | tributarioNaturalezaCO | tributarioNaturalezaCO | no |
| 161 | select/- | entrada | tributarioRegimenCO | Seleccionar Renta regimen ordinario Regimen Simple de Tributacion - SIMPLE Regimen Tributario Especial Ingresos y patrim | no |
| 162 | select/- | entrada | tributarioIVAResponsabilidadCO | Seleccionar Responsable de IVA No responsable de IVA Persona juridica no responsable de IVA por SIMPLE Productor/exporta | no |
| 163 | select/- | entrada | tributarioINCResponsabilidadCO | No aplica Responsable de INC No responsable de consumo restaurantes y bares | no |
| 164 | select/- | entrada | tributarioRutExtraCO | 07 Retencion en la fuente a titulo de renta 09 Retencion en la fuente en IVA 13 Gran contribuyente 14 Informante de exog | no |
| 165 | button/button | accion | btnRefrescarImpresoras | Refrescar | no |
| 166 | input/hidden | entrada | impresoraId | 0 | no |
| 167 | input/- | entrada | impresoraCodigo | impresoraCodigo | no |
| 168 | input/- | entrada | impresoraNombre | impresoraNombre | no |
| 169 | select/- | entrada | impresoraTipoConexion | Red USB Windows Bluetooth | no |
| 170 | select/- | entrada | impresoraFormato | POS Carta | no |
| 171 | input/- | entrada | impresoraDireccion | impresoraDireccion | no |
| 172 | input/- | entrada | impresoraArea | impresoraArea | no |
| 173 | input/checkbox | accion | impresoraPredeterminada | impresoraPredeterminada | no |
| 174 | input/checkbox | accion | impresoraActiva | impresoraActiva | no |
| 175 | input/- | entrada | impresoraObservaciones | impresoraObservaciones | no |
| 176 | button/button | accion | btnGuardarImpresora | Guardar impresora | no |
| 177 | button/button | accion | btnLimpiarImpresora | Limpiar formulario | no |
| 178 | select/- | entrada | printerFuncionalidad | Recibo de venta / ticket de cobro Orden de servicio Factura de caja Reporte de caja Comanda cocina Comanda barra Recepci | no |
| 179 | select/- | entrada | printerFuncionalidadImpresora | printerFuncionalidadImpresora | no |
| 180 | button/button | accion | btnGuardarImpresoraFuncionalidad | Guardar funcionalidad | no |
| 181 | select/- | entrada | printerProductoAlcance | Producto especifico Categoria de producto Todos los productos | no |
| 182 | select/- | entrada | printerProducto | printerProducto | no |
| 183 | select/- | entrada | printerCategoria | printerCategoria | no |
| 184 | select/- | entrada | printerProductoImpresora | printerProductoImpresora | no |
| 185 | button/button | accion | btnGuardarImpresoraProducto | Guardar asignacion | no |
| 186 | button/button | accion | btnDescargarBackup | Descargar backup | no |
| 187 | input/file | accion | backupFile | backupFile | no |
| 188 | button/button | accion | btnRestaurarBackup | Restaurar backup | no |
| 189 | button/button | accion | btnGuardarPasarelasPago | Guardar pasarelas | no |
| 190 | input/checkbox | accion | pasarelaWompiActiva | pasarelaWompiActiva | no |
| 191 | select/- | entrada | pasarelaWompiModo | sandbox production | no |
| 192 | input/- | entrada | pasarelaWompiPublicKey | pasarelaWompiPublicKey | no |
| 193 | input/- | entrada | pasarelaWompiPrivateRef | pasarelaWompiPrivateRef | no |
| 194 | input/- | entrada | pasarelaWompiIntegrityRef | pasarelaWompiIntegrityRef | no |
| 195 | input/- | entrada | pasarelaWompiEventRef | pasarelaWompiEventRef | no |
| 196 | input/checkbox | accion | pasarelaEpaycoActiva | pasarelaEpaycoActiva | no |
| 197 | select/- | entrada | pasarelaEpaycoModo | sandbox production | no |
| 198 | input/- | entrada | pasarelaEpaycoPublicKey | pasarelaEpaycoPublicKey | no |
| 199 | input/- | entrada | pasarelaEpaycoPrivateRef | pasarelaEpaycoPrivateRef | no |
| 200 | input/- | entrada | pasarelaEpaycoCustomerId | pasarelaEpaycoCustomerId | no |
| 201 | button/button | accion | btnRefrescarPasarelasPago | Refrescar | no |
| 202 | button/button | accion | btnAbrirVentaPublicaPublica | Abrir tienda pública | no |
| 203 | button/button | accion | ' + Number(item.id \|\| 0) + ' | Editar | sí |
| 204 | button/button | accion | ' + Number(item.id \|\| 0) + ' | Predeterminada | sí |
| 205 | button/button | accion | ' + Number(item.id \|\| 0) + ' | ' + (active ? 'Desactivar' : 'Activar') + ' | sí |
| 206 | button/button | accion | - | Eliminar | sí |
| 207 | button/button | accion | ' + categoriaID + ' | Eliminar | sí |
| 208 | button/button | accion | ' + productoID + ' | Eliminar | sí |

### `web/administrar_empresa/configuracion/cobro_operativo.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Ver todas | no |

### `web/administrar_empresa/configuracion/corte_de_caja.html` (4)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnGuardarConfigReporte | Guardar configuracion | no |
| 2 | button/button | accion | btnRestaurarConfigReporte | Restaurar profesional | no |
| 3 | select/- | entrada | corteFormatoImpresion | Ticket POS 80mm Ejecutivo compacto | no |
| 4 | input/checkbox | accion | - | sin etiqueta | sí |

### `web/administrar_empresa/configuracion/email_corporativo.html` (4)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/checkbox | accion | autoOpenEmail | autoOpenEmail | no |
| 2 | input/password | entrada | emailPassword | emailPassword | no |
| 3 | input/password | entrada | emailPasswordConfirm | emailPasswordConfirm | no |
| 4 | button/submit | accion | saveEmailConfig | Guardar cambios | no |

### `web/administrar_empresa/configuracion/formato_monetario.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Ver todas | no |

### `web/administrar_empresa/configuracion/identidad_visual.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Ver todas | no |

### `web/administrar_empresa/configuracion/menu_visual.html` (5)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnGuardarMenuVisual | Guardar configuracion | no |
| 2 | input/checkbox | accion | menuVisualEnabled | menuVisualEnabled | no |
| 3 | button/button | accion | btnMostrarTodo | Mostrar todo | no |
| 4 | button/button | accion | btnOcultarNoCriticos | Ocultar soluciones especializadas | no |
| 5 | input/checkbox | accion | - | sin etiqueta | sí |

### `web/administrar_empresa/configuracion/panel_inicio.html` (6)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/checkbox | accion | favoritosEnabled | favoritosEnabled | no |
| 2 | input/checkbox | accion | emailEnabled | emailEnabled | no |
| 3 | input/checkbox | accion | noticiasEnabled | noticiasEnabled | no |
| 4 | input/checkbox | accion | buzonEnabled | buzonEnabled | no |
| 5 | input/checkbox | accion | chatEnabled | chatEnabled | no |
| 6 | button/submit | accion | panelConfigSave | Guardar configuracion | no |

### `web/administrar_empresa/configuracion/pasarelas_pago.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Ver todas | no |

### `web/administrar_empresa/configuracion/perfil_tributario_colombia.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Ver todas | no |

### `web/administrar_empresa/configuracion/productos_pedidos.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Ver todas | no |

### `web/administrar_empresa/configuracion/respaldo.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Ver todas | no |

### `web/administrar_empresa/configuracion_carrito_de_compra_empresa.html` (96)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnGuardarConfigCarrito | Guardar configuraciÃ³n | no |
| 2 | input/checkbox | accion | carritoCfgPitidoPc | carritoCfgPitidoPc | no |
| 3 | input/checkbox | accion | carritoCfgPitidoMovil | carritoCfgPitidoMovil | no |
| 4 | input/checkbox | accion | carritoCfgVibracionPc | carritoCfgVibracionPc | no |
| 5 | input/checkbox | accion | carritoCfgVibracionMovil | carritoCfgVibracionMovil | no |
| 6 | input/checkbox | accion | carritoCfgAtajosPOSEnabled | carritoCfgAtajosPOSEnabled | no |
| 7 | input/checkbox | accion | carritoCfgAtajosPOSAyuda | carritoCfgAtajosPOSAyuda | no |
| 8 | input/checkbox | accion | carritoCfgModoTactil | carritoCfgModoTactil | no |
| 9 | input/checkbox | accion | carritoCfgBasculaElectronica | carritoCfgBasculaElectronica | no |
| 10 | textarea/- | entrada | carritoCfgMetodosPagoPersonalizados | carritoCfgMetodosPagoPersonalizados | no |
| 11 | input/checkbox | accion | carritoCfgPermitirVentaSinStock | carritoCfgPermitirVentaSinStock | no |
| 12 | input/checkbox | accion | carritoCfgBuscarProductos | carritoCfgBuscarProductos | no |
| 13 | input/checkbox | accion | carritoCfgBusquedaCatalogo | carritoCfgBusquedaCatalogo | no |
| 14 | input/checkbox | accion | carritoCfgCodigoManual | carritoCfgCodigoManual | no |
| 15 | input/checkbox | accion | carritoCfgObservaciones | carritoCfgObservaciones | no |
| 16 | input/checkbox | accion | carritoCfgCliente | carritoCfgCliente | no |
| 17 | input/checkbox | accion | carritoCfgClienteObligatorio | carritoCfgClienteObligatorio | no |
| 18 | input/- | entrada | carritoCfgClienteGeneralNombre | carritoCfgClienteGeneralNombre | no |
| 19 | input/checkbox | accion | carritoCfgImpuestos | carritoCfgImpuestos | no |
| 20 | input/checkbox | accion | carritoCfgLector | carritoCfgLector | no |
| 21 | select/- | entrada | carritoCfgBusquedaPredeterminada | Codigo de barras Codigo SKU Nombre | no |
| 22 | input/checkbox | accion | carritoCfgCantidadDecimal | carritoCfgCantidadDecimal | no |
| 23 | input/checkbox | accion | carritoCfgAltoItemsContenido | carritoCfgAltoItemsContenido | no |
| 24 | input/checkbox | accion | carritoCfgMostrarTodosItems | carritoCfgMostrarTodosItems | no |
| 25 | input/checkbox | accion | carritoCfgResumenProductos | carritoCfgResumenProductos | no |
| 26 | input/checkbox | accion | carritoCfgMostrarPagar | carritoCfgMostrarPagar | no |
| 27 | input/checkbox | accion | carritoCfgPreguntarDocumentoPago | carritoCfgPreguntarDocumentoPago | no |
| 28 | input/checkbox | accion | carritoCfgMostrarLogoEmpresa | carritoCfgMostrarLogoEmpresa | no |
| 29 | select/- | entrada | carritoCfgLogoEmpresaTamano | Pequeno Mediano Grande | no |
| 30 | input/checkbox | accion | carritoCfgMostrarTarjetasPago | carritoCfgMostrarTarjetasPago | no |
| 31 | button/button | accion | btnGuardarVisibilidadPagoCarrito | Guardar cambios | no |
| 32 | input/checkbox | accion | carritoCfgTarjetaLector | carritoCfgTarjetaLector | no |
| 33 | input/checkbox | accion | carritoCfgTarjetaItems | carritoCfgTarjetaItems | no |
| 34 | input/checkbox | accion | carritoCfgTarjetaTotales | carritoCfgTarjetaTotales | no |
| 35 | input/checkbox | accion | carritoCfgTotalEntrada | carritoCfgTotalEntrada | no |
| 36 | input/checkbox | accion | carritoCfgTotalSalida | carritoCfgTotalSalida | no |
| 37 | input/checkbox | accion | carritoCfgTotalItems | carritoCfgTotalItems | no |
| 38 | input/checkbox | accion | carritoCfgTotalProductos | carritoCfgTotalProductos | no |
| 39 | input/checkbox | accion | carritoCfgTotalServicios | carritoCfgTotalServicios | no |
| 40 | input/checkbox | accion | carritoCfgTotalFacturadoProductos | carritoCfgTotalFacturadoProductos | no |
| 41 | input/checkbox | accion | carritoCfgTotalFacturadoServicios | carritoCfgTotalFacturadoServicios | no |
| 42 | input/checkbox | accion | carritoCfgTotalCliente | carritoCfgTotalCliente | no |
| 43 | input/checkbox | accion | carritoCfgTotalSubtotal | carritoCfgTotalSubtotal | no |
| 44 | input/checkbox | accion | carritoCfgTotalDescuento | carritoCfgTotalDescuento | no |
| 45 | input/checkbox | accion | carritoCfgTotalImpuesto | carritoCfgTotalImpuesto | no |
| 46 | input/checkbox | accion | carritoCfgTotalPagado | carritoCfgTotalPagado | no |
| 47 | input/checkbox | accion | carritoCfgTotalGeneral | carritoCfgTotalGeneral | no |
| 48 | input/checkbox | accion | carritoCfgTotalSaldo | carritoCfgTotalSaldo | no |
| 49 | input/checkbox | accion | carritoCfgTarjetaCobro | carritoCfgTarjetaCobro | no |
| 50 | input/checkbox | accion | carritoCfgTarjetaAcciones | carritoCfgTarjetaAcciones | no |
| 51 | input/checkbox | accion | carritoCfgTarjetaValoresPago | carritoCfgTarjetaValoresPago | no |
| 52 | input/checkbox | accion | carritoCfgTarjetaComision | carritoCfgTarjetaComision | no |
| 53 | input/checkbox | accion | carritoCfgTarjetaVip | carritoCfgTarjetaVip | no |
| 54 | input/checkbox | accion | carritoCfgControlElectrico | carritoCfgControlElectrico | no |
| 55 | input/checkbox | accion | carritoCfgTarjetaDomotica | carritoCfgTarjetaDomotica | no |
| 56 | input/checkbox | accion | carritoCfgBotonDescuentos | carritoCfgBotonDescuentos | no |
| 57 | input/checkbox | accion | carritoCfgBotonCambiarTarifa | carritoCfgBotonCambiarTarifa | no |
| 58 | input/checkbox | accion | carritoCfgBotonTransferirCuenta | carritoCfgBotonTransferirCuenta | no |
| 59 | input/checkbox | accion | carritoCfgBotonControlElectrico | carritoCfgBotonControlElectrico | no |
| 60 | input/checkbox | accion | carritoCfgBotonCancelar | carritoCfgBotonCancelar | no |
| 61 | input/checkbox | accion | carritoCfgBotonClientes | carritoCfgBotonClientes | no |
| 62 | input/checkbox | accion | carritoCfgBotonAbonos | carritoCfgBotonAbonos | no |
| 63 | input/checkbox | accion | carritoCfgBotonVehiculo | carritoCfgBotonVehiculo | no |
| 64 | input/checkbox | accion | carritoCfgBotonesSelectorEnabled | carritoCfgBotonesSelectorEnabled | no |
| 65 | select/- | entrada | carritoCfgBotonesSelector | Selecciona un boton Descuentos Cambiar tarifa Transferir cuenta Domotica Cancelar carrito Clientes Abonos Vehiculo | no |
| 66 | select/- | entrada | carritoCfgBotonesSelectorAction | Mostrar Ocultar | no |
| 67 | button/button | accion | carritoCfgBotonesSelectorApply | Aplicar al boton | no |
| 68 | input/checkbox | accion | carritoCfgMostrarAlertaTiempo | carritoCfgMostrarAlertaTiempo | no |
| 69 | input/number | entrada | carritoCfgAlertaTiempoMinutos | 10 | no |
| 70 | input/checkbox | accion | carritoCfgAlertaTiempoDefault | carritoCfgAlertaTiempoDefault | no |
| 71 | input/checkbox | accion | carritoCfgDescuentos | carritoCfgDescuentos | no |
| 72 | input/checkbox | accion | carritoCfgPropina | carritoCfgPropina | no |
| 73 | input/checkbox | accion | carritoCfgComision | carritoCfgComision | no |
| 74 | input/checkbox | accion | carritoCfgPagoMixto | carritoCfgPagoMixto | no |
| 75 | input/checkbox | accion | carritoCfgResumenTotales | carritoCfgResumenTotales | no |
| 76 | input/checkbox | accion | carritoCfgDesgloseCobro | carritoCfgDesgloseCobro | no |
| 77 | input/checkbox | accion | carritoCfgMetodoPagoEfectivo | carritoCfgMetodoPagoEfectivo | no |
| 78 | input/checkbox | accion | carritoCfgMetodoPagoCredito | carritoCfgMetodoPagoCredito | no |
| 79 | input/checkbox | accion | carritoCfgMetodoPagoDebito | carritoCfgMetodoPagoDebito | no |
| 80 | input/checkbox | accion | carritoCfgMetodoPagoBreb | carritoCfgMetodoPagoBreb | no |
| 81 | input/checkbox | accion | carritoCfgMetodoPagoNequi | carritoCfgMetodoPagoNequi | no |
| 82 | input/checkbox | accion | carritoCfgMetodoPagoOtraTransferencia | carritoCfgMetodoPagoOtraTransferencia | no |
| 83 | input/checkbox | accion | carritoCfgMetodoPagoCreditoCliente | carritoCfgMetodoPagoCreditoCliente | no |
| 84 | input/checkbox | accion | carritoCfgFacturacionOffline | carritoCfgFacturacionOffline | no |
| 85 | input/checkbox | accion | carritoCfgMarcaOffline | carritoCfgMarcaOffline | no |
| 86 | input/checkbox | accion | carritoCfgQrFacturaElectronica | carritoCfgQrFacturaElectronica | no |
| 87 | button/button | accion | btnAgregarCuentaQR | Agregar cuenta | no |
| 88 | input/checkbox | accion | carritoCfgPagoQR | carritoCfgPagoQR | no |
| 89 | input/checkbox | accion | - | sin etiqueta | sí |
| 90 | input/- | entrada | - | ' + escapeAttr(account.nombre) + ' | sí |
| 91 | select/- | entrada | - | ' + ' Bre-B ' + ' Nequi ' + ' Otro ' + ' | sí |
| 92 | input/- | entrada | - | ' + escapeAttr(account.llave) + ' | sí |
| 93 | input/- | entrada | - | ' + escapeAttr(account.comercio) + ' | sí |
| 94 | textarea/- | entrada | - | ' + escapeHtmlText(account.payload_oficial) + ' | sí |
| 95 | textarea/- | entrada | - | ' + escapeHtmlText(account.instrucciones) + ' | sí |
| 96 | button/button | accion | - | Eliminar | no |

### `web/administrar_empresa/configuracion_chat_flotante.html` (9)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/checkbox | accion | chatEnabled | chatEnabled | no |
| 2 | input/checkbox | accion | radioOnlineEnabled | radioOnlineEnabled | no |
| 3 | select/- | entrada | chatTheme | Normal en negro Corporativo: títulos rojos y datos azules Océano: azules profesionales Esmeralda: verdes profesionales V | no |
| 4 | select/- | entrada | chatTextSize | Pequeño Mediano Grande | no |
| 5 | input/checkbox | accion | chatVoiceEnabled | chatVoiceEnabled | no |
| 6 | select/- | entrada | chatRobotVoice | Colombiana natural (predeterminada) Colombiana femenina Colombiana masculina Espanol latino Espanol castellano | no |
| 7 | button/button | accion | saveChatModeBtn | Guardar configuracion | no |
| 8 | button/button | accion | resetChatModeBtn | Restablecer IA activa | no |
| 9 | a/- | accion | - | Volver al panel | no |

### `web/administrar_empresa/configuracion_de_estaciones.html` (115)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/- | entrada | estacionNombreSingular | estacionNombreSingular | no |
| 2 | input/- | entrada | estacionNombrePlural | estacionNombrePlural | no |
| 3 | button/button | accion | saveStationLabelsBtn | Guardar nombres | no |
| 4 | input/checkbox | accion | cajaToggle | cajaToggle | no |
| 5 | select/- | entrada | cajaPlacementSelect | Antes de las estaciones Después de las estaciones | no |
| 6 | input/checkbox | accion | youtubeToggle | youtubeToggle | no |
| 7 | select/- | entrada | youtubePlacementSelect | Antes de las estaciones Después de las estaciones | no |
| 8 | input/checkbox | accion | notasToggle | notasToggle | no |
| 9 | select/- | entrada | notasPlacementSelect | Antes de las estaciones Después de las estaciones | no |
| 10 | input/checkbox | accion | iaPedidosToggle | iaPedidosToggle | no |
| 11 | select/- | entrada | iaPedidosPlacementSelect | Antes de las estaciones Después de las estaciones | no |
| 12 | input/checkbox | accion | camarasToggle | camarasToggle | no |
| 13 | select/- | entrada | camarasPlacementSelect | Antes de las estaciones Despues de las estaciones | no |
| 14 | button/button | accion | btnAgregarCajaConfig | Agregar caja | no |
| 15 | button/button | accion | btnGuardarCajasConfig | Guardar cajas | no |
| 16 | input/checkbox | accion | cajaLoginAutoComputadorToggle | cajaLoginAutoComputadorToggle | no |
| 17 | input/- | entrada | youtubeReferenceInput | youtubeReferenceInput | no |
| 18 | textarea/- | entrada | notasDefaultTextInput | notasDefaultTextInput | no |
| 19 | input/number | entrada | notasDefaultMinutesInput | 5 | no |
| 20 | input/number | entrada | notasRepeatMinutesInput | 0 | no |
| 21 | input/checkbox | accion | notasBeepToggle | notasBeepToggle | no |
| 22 | input/checkbox | accion | notasBlinkToggle | notasBlinkToggle | no |
| 23 | button/button | accion | saveYoutubeReferenceBtn | Guardar referencia YouTube | no |
| 24 | button/button | accion | saveNotasConfigBtn | Guardar configuración Notas | no |
| 25 | input/checkbox | accion | stationCancelTimeEnabled | stationCancelTimeEnabled | no |
| 26 | input/number | entrada | stationCancelTimeMinutes | 5 | no |
| 27 | button/button | accion | saveStationCancelTimeBtn | Guardar tiempo para cancelar | no |
| 28 | select/- | entrada | cardSizeSelect | Pequeño Medio Grande Se adapta al texto | no |
| 29 | button/button | accion | btnGuardarApariencia | Guardar apariencia | no |
| 30 | button/button | accion | btnGuardarEstadoCarrito | Guardar colores | no |
| 31 | input/color | entrada | carritoColorActivo | #d9fbe8 | no |
| 32 | input/color | entrada | carritoColorInactivo | #fff9ef | no |
| 33 | input/color | entrada | estacionColorDisponible | #fff9ef | no |
| 34 | input/color | entrada | estacionColorOcupada | #d9fbe8 | no |
| 35 | input/color | entrada | estacionColorSucia | #ffe0e0 | no |
| 36 | input/color | entrada | estacionColorAlertaTiempo | #fff3cd | no |
| 37 | input/checkbox | accion | stShowClienteNombre | stShowClienteNombre | no |
| 38 | input/checkbox | accion | stShowTarifaResumen | stShowTarifaResumen | no |
| 39 | input/checkbox | accion | stShowHoraExtra | stShowHoraExtra | no |
| 40 | input/checkbox | accion | stAlert10Min | stAlert10Min | no |
| 41 | input/checkbox | accion | stMarkDirtyOnPay | stMarkDirtyOnPay | no |
| 42 | input/checkbox | accion | stOnlyActivateFirstClick | stOnlyActivateFirstClick | no |
| 43 | input/checkbox | accion | stShowDomoticaButton | stShowDomoticaButton | no |
| 44 | button/button | accion | btnGuardarStationCardFlags | Guardar opciones de visualización | no |
| 45 | input/number | entrada | cantidadEstaciones | 1 | no |
| 46 | button/submit | accion | - | Generar estaciones | no |
| 47 | button/button | accion | btnInactivarCarritos | Inactivar carritos de estaciones | no |
| 48 | button/button | accion | btnGuardarNombres | Guardar nombres | no |
| 49 | input/checkbox | accion | masterCarritoGlobal | Marcar/desmarcar uso de configuración global para todas | no |
| 50 | input/checkbox | accion | masterMostrarInicio | Marcar/desmarcar fecha y hora de inicio para todas | no |
| 51 | input/checkbox | accion | masterMostrarFinTarifa | Marcar/desmarcar fin de tarifa para todas | no |
| 52 | input/checkbox | accion | masterMostrarTotal | Marcar/desmarcar mostrar total para todas | no |
| 53 | button/button | accion | btnGuardarCarritoConfig | Guardar configuración seleccionada | no |
| 54 | button/button | accion | btnAplicarCarritoGlobal | Aplicar global a todas | no |
| 55 | select/- | entrada | carritoConfigTarget | carritoConfigTarget | no |
| 56 | input/checkbox | accion | carritoConfigUseGlobal | carritoConfigUseGlobal | no |
| 57 | input/checkbox | accion | carritoCfgModoTactil | carritoCfgModoTactil | no |
| 58 | input/checkbox | accion | carritoCfgBuscarProductos | carritoCfgBuscarProductos | no |
| 59 | input/checkbox | accion | carritoCfgBusquedaCatalogo | carritoCfgBusquedaCatalogo | no |
| 60 | input/checkbox | accion | carritoCfgCodigoManual | carritoCfgCodigoManual | no |
| 61 | input/checkbox | accion | carritoCfgObservaciones | carritoCfgObservaciones | no |
| 62 | input/checkbox | accion | carritoCfgCliente | carritoCfgCliente | no |
| 63 | input/checkbox | accion | carritoCfgClienteObligatorio | carritoCfgClienteObligatorio | no |
| 64 | input/- | entrada | carritoCfgClienteGeneralNombre | carritoCfgClienteGeneralNombre | no |
| 65 | input/checkbox | accion | carritoCfgImpuestos | carritoCfgImpuestos | no |
| 66 | input/checkbox | accion | carritoCfgLector | carritoCfgLector | no |
| 67 | input/checkbox | accion | carritoCfgResumenProductos | carritoCfgResumenProductos | no |
| 68 | input/checkbox | accion | carritoCfgMostrarPagar | carritoCfgMostrarPagar | no |
| 69 | input/checkbox | accion | carritoCfgMostrarTarjetasPago | carritoCfgMostrarTarjetasPago | no |
| 70 | button/button | accion | btnGuardarVisibilidadPagoCarrito | Guardar cambios | no |
| 71 | input/checkbox | accion | carritoCfgTarjetaLector | carritoCfgTarjetaLector | no |
| 72 | input/checkbox | accion | carritoCfgTarjetaItems | carritoCfgTarjetaItems | no |
| 73 | input/checkbox | accion | carritoCfgTarjetaTotales | carritoCfgTarjetaTotales | no |
| 74 | input/checkbox | accion | carritoCfgTarjetaCobro | carritoCfgTarjetaCobro | no |
| 75 | input/checkbox | accion | carritoCfgTarjetaAcciones | carritoCfgTarjetaAcciones | no |
| 76 | input/checkbox | accion | carritoCfgTarjetaValoresPago | carritoCfgTarjetaValoresPago | no |
| 77 | input/checkbox | accion | carritoCfgTarjetaComision | carritoCfgTarjetaComision | no |
| 78 | input/checkbox | accion | carritoCfgTarjetaVip | carritoCfgTarjetaVip | no |
| 79 | input/checkbox | accion | carritoCfgControlElectrico | carritoCfgControlElectrico | no |
| 80 | input/checkbox | accion | carritoCfgTarjetaDomotica | carritoCfgTarjetaDomotica | no |
| 81 | input/checkbox | accion | carritoCfgBotonDescuentos | carritoCfgBotonDescuentos | no |
| 82 | input/checkbox | accion | carritoCfgBotonCambiarTarifa | carritoCfgBotonCambiarTarifa | no |
| 83 | input/checkbox | accion | carritoCfgBotonControlElectrico | carritoCfgBotonControlElectrico | no |
| 84 | input/checkbox | accion | carritoCfgBotonCancelar | carritoCfgBotonCancelar | no |
| 85 | input/checkbox | accion | carritoCfgBotonTaxi | carritoCfgBotonTaxi | no |
| 86 | input/checkbox | accion | carritoCfgBotonClientes | carritoCfgBotonClientes | no |
| 87 | input/checkbox | accion | carritoCfgBotonAbonos | carritoCfgBotonAbonos | no |
| 88 | input/checkbox | accion | carritoCfgBotonVehiculo | carritoCfgBotonVehiculo | no |
| 89 | input/checkbox | accion | carritoCfgMostrarAlertaTiempo | carritoCfgMostrarAlertaTiempo | no |
| 90 | input/number | entrada | carritoCfgAlertaTiempoMinutos | 10 | no |
| 91 | input/checkbox | accion | carritoCfgAlertaTiempoDefault | carritoCfgAlertaTiempoDefault | no |
| 92 | input/checkbox | accion | carritoCfgDescuentos | carritoCfgDescuentos | no |
| 93 | input/checkbox | accion | carritoCfgPropina | carritoCfgPropina | no |
| 94 | input/checkbox | accion | carritoCfgComision | carritoCfgComision | no |
| 95 | input/checkbox | accion | carritoCfgPagoMixto | carritoCfgPagoMixto | no |
| 96 | input/checkbox | accion | carritoCfgResumenTotales | carritoCfgResumenTotales | no |
| 97 | input/checkbox | accion | carritoCfgDesgloseCobro | carritoCfgDesgloseCobro | no |
| 98 | input/checkbox | accion | carritoCfgQrFacturaElectronica | carritoCfgQrFacturaElectronica | no |
| 99 | input/checkbox | accion | - | sin etiqueta | sí |
| 100 | input/- | entrada | - | ' + sanitize(caja.codigo) + ' | sí |
| 101 | input/- | entrada | - | ' + sanitize(caja.nombre) + ' | sí |
| 102 | select/- | entrada | - | ' + renderBodegaOptions(caja.bodega_id) + ' | sí |
| 103 | input/- | entrada | - | ' + sanitize(caja.descripcion) + ' | sí |
| 104 | button/button | accion | - | Eliminar | sí |
| 105 | input/- | entrada | - | ' + sanitize(est.nombre) + ' | sí |
| 106 | input/checkbox | accion | - | sin etiqueta | sí |
| 107 | input/checkbox | accion | - | sin etiqueta | sí |
| 108 | input/checkbox | accion | - | sin etiqueta | sí |
| 109 | input/checkbox | accion | - | sin etiqueta | sí |
| 110 | select/- | entrada | - | ' + ' General ' + ' Motel ' + ' Hotel ' + ' Restaurante ' + ' Lavadero ' + ' | sí |
| 111 | input/- | entrada | - | ' + sanitize(est.descripcion \|\| '') + ' | sí |
| 112 | input/- | entrada | - | ' + sanitize(est.usuario_asignado \|\| '') + ' | sí |
| 113 | select/- | entrada | - | Normal Camara | sí |
| 114 | select/- | entrada | - | ' + buildCamaraOptions(est.camara_id) + ' | sí |
| 115 | button/button | accion | - | Editar checks | sí |

### `web/administrar_empresa/configuracion_guiada.html` (8)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | guidedFormBtn | Abrir formulario guiado | no |
| 2 | button/button | accion | guidedStartBtn | Abrir chat IA | no |
| 3 | button/button | accion | guidedReloadBtn | Actualizar resumen | no |
| 4 | button/submit | accion | - | Guardar configuracion guiada | no |
| 5 | select/- | entrada | - | ' + ' Si ' + ' No ' + ' | sí |
| 6 | select/- | entrada | - | ' + question.options.map(function (option) { option = String(option \|\| ''); return ' ' + escapeHtml(option) + ' '; }).jo | sí |
| 7 | textarea/- | entrada | - | ' + escapeHtml(value) + ' | sí |
| 8 | input/' + (type === 'number' ? 'number' : 'text') + ' | entrada | - | ' + escapeHtml(value) + ' | sí |

### `web/administrar_empresa/configuracion_ia_propia.html` (5)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/checkbox | accion | ownOpenAIEnabled | ownOpenAIEnabled | no |
| 2 | input/password | entrada | ownOpenAIKey | ownOpenAIKey | no |
| 3 | input/checkbox | accion | ownOpenAIClear | ownOpenAIClear | no |
| 4 | button/button | accion | saveOwnOpenAI | Guardar configuración | no |
| 5 | a/- | accion | - | Volver a configuración | no |

### `web/administrar_empresa/configuracion_impresora.html` (195)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnGuardarConfig | Guardar configuración | no |
| 2 | input/checkbox | accion | imprimirOrdenServicio | imprimirOrdenServicio | no |
| 3 | input/checkbox | accion | habilitarDescuentos | habilitarDescuentos | no |
| 4 | input/checkbox | accion | permitirDescuentoPorcentaje | permitirDescuentoPorcentaje | no |
| 5 | input/checkbox | accion | permitirDescuentoCodigo | permitirDescuentoCodigo | no |
| 6 | input/checkbox | accion | permitirDescuentoValor | permitirDescuentoValor | no |
| 7 | input/checkbox | accion | habilitarLectorCodigoBarras | habilitarLectorCodigoBarras | no |
| 8 | input/checkbox | accion | autofocoLectorCodigoBarras | autofocoLectorCodigoBarras | no |
| 9 | input/checkbox | accion | acumularLectorCodigoBarras | acumularLectorCodigoBarras | no |
| 10 | input/- | entrada | areaDespacho | areaDespacho | no |
| 11 | input/number | entrada | copiasOrden | 1 | no |
| 12 | textarea/- | entrada | notaOrden | notaOrden | no |
| 13 | textarea/- | entrada | codigosDescuento | codigosDescuento | no |
| 14 | button/button | accion | btnGuardarConfigOperativa | Guardar empresa | no |
| 15 | input/checkbox | accion | opMetodoEfectivo | opMetodoEfectivo | no |
| 16 | input/checkbox | accion | opMetodoTarjetaCredito | opMetodoTarjetaCredito | no |
| 17 | input/checkbox | accion | opMetodoTarjetaDebito | opMetodoTarjetaDebito | no |
| 18 | input/checkbox | accion | opMetodoTransferencia | opMetodoTransferencia | no |
| 19 | input/checkbox | accion | opMetodoMixto | opMetodoMixto | no |
| 20 | input/checkbox | accion | opMetodoCodigoDescuento | opMetodoCodigoDescuento | no |
| 21 | input/checkbox | accion | opHabilitarPropinas | opHabilitarPropinas | no |
| 22 | input/checkbox | accion | opHabilitarComisiones | opHabilitarComisiones | no |
| 23 | select/- | entrada | opRolSelect | admin_empresa supervisor_sucursal cajero inventario compras contabilidad auditor | no |
| 24 | input/checkbox | accion | opRolActivo | opRolActivo | no |
| 25 | input/checkbox | accion | opRolMetodoEfectivo | opRolMetodoEfectivo | no |
| 26 | input/checkbox | accion | opRolMetodoTarjetaCredito | opRolMetodoTarjetaCredito | no |
| 27 | input/checkbox | accion | opRolMetodoTarjetaDebito | opRolMetodoTarjetaDebito | no |
| 28 | input/checkbox | accion | opRolMetodoTransferencia | opRolMetodoTransferencia | no |
| 29 | input/checkbox | accion | opRolMetodoMixto | opRolMetodoMixto | no |
| 30 | input/checkbox | accion | opRolMetodoCodigoDescuento | opRolMetodoCodigoDescuento | no |
| 31 | input/checkbox | accion | opRolHabilitarPropinas | opRolHabilitarPropinas | no |
| 32 | input/checkbox | accion | opRolHabilitarComisiones | opRolHabilitarComisiones | no |
| 33 | input/checkbox | accion | opRolPermitirIngresosManuales | opRolPermitirIngresosManuales | no |
| 34 | input/checkbox | accion | opRolPermitirEgresosManuales | opRolPermitirEgresosManuales | no |
| 35 | button/button | accion | btnGuardarConfigOperativaRol | Guardar rol | no |
| 36 | select/- | entrada | opPoliticaCanal | Todos mostrador app estacion reserva online delivery kiosko | no |
| 37 | input/number | entrada | opPoliticaSucursalId | 0 | no |
| 38 | input/- | entrada | opPoliticaTurno | opPoliticaTurno | no |
| 39 | input/number | entrada | opPoliticaPrioridad | 100 | no |
| 40 | input/checkbox | accion | opPoliticaActivo | opPoliticaActivo | no |
| 41 | input/checkbox | accion | opPoliticaMetodoEfectivo | opPoliticaMetodoEfectivo | no |
| 42 | input/checkbox | accion | opPoliticaMetodoTarjetaCredito | opPoliticaMetodoTarjetaCredito | no |
| 43 | input/checkbox | accion | opPoliticaMetodoTarjetaDebito | opPoliticaMetodoTarjetaDebito | no |
| 44 | input/checkbox | accion | opPoliticaMetodoTransferencia | opPoliticaMetodoTransferencia | no |
| 45 | input/checkbox | accion | opPoliticaMetodoMixto | opPoliticaMetodoMixto | no |
| 46 | input/checkbox | accion | opPoliticaMetodoCodigoDescuento | opPoliticaMetodoCodigoDescuento | no |
| 47 | input/checkbox | accion | opPoliticaHabilitarPropinas | opPoliticaHabilitarPropinas | no |
| 48 | input/checkbox | accion | opPoliticaHabilitarComisiones | opPoliticaHabilitarComisiones | no |
| 49 | button/button | accion | btnGuardarConfigOperativaPolitica | Guardar politica | no |
| 50 | select/- | entrada | opSimRol | admin_empresa supervisor_sucursal cajero inventario compras contabilidad auditor | no |
| 51 | input/- | entrada | opSimCanal | opSimCanal | no |
| 52 | input/number | entrada | opSimSucursalId | 0 | no |
| 53 | input/- | entrada | opSimTurno | opSimTurno | no |
| 54 | input/- | entrada | opSimObservaciones | opSimObservaciones | no |
| 55 | button/button | accion | btnSimularConfigOperativa | Simular | no |
| 56 | button/button | accion | btnSimularGuardarConfigOperativa | Simular y guardar | no |
| 57 | input/number | entrada | opRollbackHistorialId | opRollbackHistorialId | no |
| 58 | button/button | accion | btnRefrescarOperativaHistorial | Refrescar historial | no |
| 59 | button/button | accion | btnRollbackConfigOperativa | Aplicar rollback | no |
| 60 | button/button | accion | btnGuardarCajaConfig | Guardar caja | no |
| 61 | input/checkbox | accion | cajaActiva | cajaActiva | no |
| 62 | input/checkbox | accion | cajasSimultaneasHabilitadas | cajasSimultaneasHabilitadas | no |
| 63 | input/text | entrada | cajaNombre | cajaNombre | no |
| 64 | input/text | entrada | cajaCodigo | cajaCodigo | no |
| 65 | input/number | entrada | maxCajasSimultaneasEmpresa | 0 | no |
| 66 | textarea/- | entrada | cajaObservaciones | cajaObservaciones | no |
| 67 | input/checkbox | accion | cajonMonederoHabilitado | cajonMonederoHabilitado | no |
| 68 | input/checkbox | accion | abrirCajonAlPagarCarrito | abrirCajonAlPagarCarrito | no |
| 69 | input/checkbox | accion | abrirCajonAlCerrarTransaccion | abrirCajonAlCerrarTransaccion | no |
| 70 | select/- | entrada | cajonMonederoMetodo | Impresión POS / driver de impresora Manual, solo aviso operativo | no |
| 71 | input/text | entrada | cajonMonederoImpresoraFuncionalidad | cajon_monedero | no |
| 72 | select/- | entrada | cajonMonederoComando | Pulso ESC/POS por impresora Apertura automática del driver | no |
| 73 | button/button | accion | btnGuardarFacturaConfig | Guardar documento | no |
| 74 | select/- | entrada | facturaModoDocumentoVenta | No, solo generar la venta/comprobante Sí, generar además factura electrónica | no |
| 75 | select/- | entrada | facturaFormatoImpresion | Grande / carta Pequeña POS (tirilla) | no |
| 76 | input/checkbox | accion | facturaImprimirVenta | facturaImprimirVenta | no |
| 77 | input/checkbox | accion | facturaImprimirFacturaElectronica | facturaImprimirFacturaElectronica | no |
| 78 | input/checkbox | accion | facturaImprimirCopia | facturaImprimirCopia | no |
| 79 | input/checkbox | accion | facturaMostrarDeducidoImpuesto | facturaMostrarDeducidoImpuesto | no |
| 80 | input/- | entrada | facturaLogoURL | facturaLogoURL | no |
| 81 | input/checkbox | accion | facturaMostrarLogo | facturaMostrarLogo | no |
| 82 | input/number | entrada | facturaFuentePOS | 11 | no |
| 83 | input/number | entrada | facturaFuenteCarta | 13 | no |
| 84 | input/number | entrada | reporteFuentePOS | 11 | no |
| 85 | input/number | entrada | reporteFuenteCarta | 13 | no |
| 86 | input/checkbox | accion | reciboItemEmpresa | reciboItemEmpresa | sí |
| 87 | input/checkbox | accion | reciboItemCarrito | reciboItemCarrito | sí |
| 88 | input/checkbox | accion | reciboItemCodigo | reciboItemCodigo | sí |
| 89 | input/checkbox | accion | reciboItemNumeroLegal | reciboItemNumeroLegal | sí |
| 90 | input/checkbox | accion | reciboItemCliente | reciboItemCliente | sí |
| 91 | input/checkbox | accion | reciboItemClienteEmail | reciboItemClienteEmail | sí |
| 92 | input/checkbox | accion | reciboItemClienteDocumento | reciboItemClienteDocumento | sí |
| 93 | input/checkbox | accion | reciboItemCajero | reciboItemCajero | sí |
| 94 | input/checkbox | accion | reciboItemMetodo | reciboItemMetodo | sí |
| 95 | input/checkbox | accion | reciboItemFecha | reciboItemFecha | sí |
| 96 | input/checkbox | accion | reciboItemEstado | reciboItemEstado | sí |
| 97 | input/checkbox | accion | reciboItemTotal | reciboItemTotal | sí |
| 98 | input/checkbox | accion | reciboItemTotalLetras | reciboItemTotalLetras | sí |
| 99 | input/checkbox | accion | reciboItemMoneda | reciboItemMoneda | sí |
| 100 | input/checkbox | accion | reciboItemPeriodo | reciboItemPeriodo | sí |
| 101 | input/checkbox | accion | reciboItemControl | reciboItemControl | sí |
| 102 | input/checkbox | accion | reciboItemTipoDocumento | reciboItemTipoDocumento | sí |
| 103 | input/checkbox | accion | reciboItemValidacion | reciboItemValidacion | sí |
| 104 | input/checkbox | accion | reciboItemPais | reciboItemPais | sí |
| 105 | input/checkbox | accion | reciboItemAmbiente | reciboItemAmbiente | sí |
| 106 | input/checkbox | accion | reciboItemObservaciones | reciboItemObservaciones | sí |
| 107 | input/checkbox | accion | reciboItemNotasLegales | reciboItemNotasLegales | sí |
| 108 | input/checkbox | accion | reciboItemQRDian | reciboItemQRDian | sí |
| 109 | input/checkbox | accion | reciboItemFormato | reciboItemFormato | sí |
| 110 | input/checkbox | accion | reciboItemImpresora | reciboItemImpresora | sí |
| 111 | input/checkbox | accion | reciboItemCopias | reciboItemCopias | sí |
| 112 | input/checkbox | accion | reciboItemCampoPersonalizado | reciboItemCampoPersonalizado | sí |
| 113 | input/- | entrada | reciboCampoPersonalizadoEtiqueta | reciboCampoPersonalizadoEtiqueta | no |
| 114 | input/- | entrada | reciboCampoPersonalizadoValor | reciboCampoPersonalizadoValor | no |
| 115 | input/checkbox | accion | reciboItemCampoPersonalizadoDescripcion | reciboItemCampoPersonalizadoDescripcion | sí |
| 116 | textarea/- | entrada | reciboCampoPersonalizadoDescripcion | reciboCampoPersonalizadoDescripcion | no |
| 117 | input/checkbox | accion | corteItemEmpresa | corteItemEmpresa | sí |
| 118 | input/checkbox | accion | corteItemFecha | corteItemFecha | sí |
| 119 | input/checkbox | accion | corteItemUsuario | corteItemUsuario | sí |
| 120 | input/checkbox | accion | corteItemCaja | corteItemCaja | sí |
| 121 | input/checkbox | accion | corteItemConsecutivo | corteItemConsecutivo | sí |
| 122 | input/checkbox | accion | corteItemEntrada | corteItemEntrada | sí |
| 123 | input/checkbox | accion | corteItemSalida | corteItemSalida | sí |
| 124 | input/checkbox | accion | corteItemVenta | corteItemVenta | sí |
| 125 | input/checkbox | accion | corteItemEstacion | corteItemEstacion | sí |
| 126 | input/checkbox | accion | corteItemCajero | corteItemCajero | sí |
| 127 | input/checkbox | accion | corteItemMedio | corteItemMedio | sí |
| 128 | input/checkbox | accion | corteItemTotal | corteItemTotal | sí |
| 129 | button/button | accion | btnGuardarFormatoNumerico | Guardar formato | no |
| 130 | input/- | entrada | formatoMonedaCodigo | formatoMonedaCodigo | no |
| 131 | select/- | entrada | formatoSistemaNumerico | Latino (1.234,56) Internacional (1,234.56) | no |
| 132 | input/checkbox | accion | formatoUsarDecimales | formatoUsarDecimales | no |
| 133 | input/number | entrada | formatoCantidadDecimales | 2 | no |
| 134 | button/button | accion | btnRefrescarImpresoras | Refrescar | no |
| 135 | input/hidden | entrada | impresoraId | 0 | no |
| 136 | input/- | entrada | impresoraCodigo | impresoraCodigo | no |
| 137 | input/- | entrada | impresoraNombre | impresoraNombre | no |
| 138 | select/- | entrada | impresoraTipoConexion | Red USB Windows Bluetooth | no |
| 139 | select/- | entrada | impresoraFormato | POS Carta | no |
| 140 | input/- | entrada | impresoraDireccion | impresoraDireccion | no |
| 141 | input/- | entrada | impresoraArea | impresoraArea | no |
| 142 | input/checkbox | accion | impresoraPredeterminada | impresoraPredeterminada | no |
| 143 | input/checkbox | accion | impresoraActiva | impresoraActiva | no |
| 144 | input/- | entrada | impresoraObservaciones | impresoraObservaciones | no |
| 145 | button/button | accion | btnGuardarImpresora | Guardar impresora | no |
| 146 | button/button | accion | btnLimpiarImpresora | Limpiar formulario | no |
| 147 | select/- | entrada | printerFuncionalidad | Recibo de venta / ticket de cobro Orden de servicio Factura de caja Cajón monedero Reporte de caja Comanda cocina Comand | no |
| 148 | select/- | entrada | printerFuncionalidadImpresora | printerFuncionalidadImpresora | no |
| 149 | button/button | accion | btnGuardarImpresoraFuncionalidad | Guardar funcionalidad | no |
| 150 | select/- | entrada | printerProductoAlcance | Producto especifico Categoria de producto Todos los productos | no |
| 151 | select/- | entrada | printerProducto | printerProducto | no |
| 152 | select/- | entrada | printerCategoria | printerCategoria | no |
| 153 | select/- | entrada | printerProductoImpresora | printerProductoImpresora | no |
| 154 | button/button | accion | btnGuardarImpresoraProducto | Guardar asignacion | no |
| 155 | input/- | entrada | printerDeviceLabel | printerDeviceLabel | no |
| 156 | input/- | entrada | printerDeviceCashCode | printerDeviceCashCode | no |
| 157 | select/- | entrada | printerDeviceFunction | General de este computador Recibo de venta / ticket Factura de caja Cajon monedero Reporte de caja | no |
| 158 | select/- | entrada | printerDevicePrinter | printerDevicePrinter | no |
| 159 | input/number | entrada | printerDeviceStation | 0 | no |
| 160 | input/- | entrada | printerDeviceNotes | printerDeviceNotes | no |
| 161 | button/button | accion | btnGuardarImpresoraDispositivo | Asociar a este computador | no |
| 162 | input/- | entrada | printerAgentId | printerAgentId | no |
| 163 | input/number | entrada | printerAgentEstacion | 0 | no |
| 164 | select/- | entrada | printerTrabajoFuncionalidad | Recibo de venta / ticket Factura Pedido / orden de servicio Comanda cocina Comanda barra Reporte o cierre de caja | no |
| 165 | select/- | entrada | printerTrabajoImpresora | printerTrabajoImpresora | no |
| 166 | input/- | entrada | printerTrabajoTitulo | Prueba de impresion PCS | no |
| 167 | textarea/- | entrada | printerTrabajoContenido | Prueba de impresion desde PCS. | no |
| 168 | button/button | accion | btnCrearTrabajoImpresion | Enviar prueba a cola | no |
| 169 | button/button | accion | btnTomarTrabajosImpresion | Tomar pendientes | no |
| 170 | button/button | accion | btnDescargarBackup | Descargar backup | no |
| 171 | input/file | accion | backupFile | backupFile | no |
| 172 | button/button | accion | btnRestaurarBackup | Restaurar backup | no |
| 173 | button/button | accion | btnGuardarPasarelasPago | Guardar pasarelas | no |
| 174 | input/checkbox | accion | pasarelaWompiActiva | pasarelaWompiActiva | no |
| 175 | select/- | entrada | pasarelaWompiModo | sandbox production | no |
| 176 | input/- | entrada | pasarelaWompiPublicKey | pasarelaWompiPublicKey | no |
| 177 | input/- | entrada | pasarelaWompiPrivateRef | pasarelaWompiPrivateRef | no |
| 178 | input/- | entrada | pasarelaWompiIntegrityRef | pasarelaWompiIntegrityRef | no |
| 179 | input/- | entrada | pasarelaWompiEventRef | pasarelaWompiEventRef | no |
| 180 | input/checkbox | accion | pasarelaEpaycoActiva | pasarelaEpaycoActiva | no |
| 181 | select/- | entrada | pasarelaEpaycoModo | sandbox production | no |
| 182 | input/- | entrada | pasarelaEpaycoPublicKey | pasarelaEpaycoPublicKey | no |
| 183 | input/- | entrada | pasarelaEpaycoPrivateRef | pasarelaEpaycoPrivateRef | no |
| 184 | input/- | entrada | pasarelaEpaycoCustomerId | pasarelaEpaycoCustomerId | no |
| 185 | button/button | accion | btnRefrescarPasarelasPago | Refrescar | no |
| 186 | button/button | accion | btnAbrirVentaPublicaPublica | Abrir tienda pública | no |
| 187 | button/button | accion | ' + Number(item.id \|\| 0) + ' | Editar | sí |
| 188 | button/button | accion | ' + Number(item.id \|\| 0) + ' | Predeterminada | sí |
| 189 | button/button | accion | ' + Number(item.id \|\| 0) + ' | ' + (active ? 'Desactivar' : 'Activar') + ' | sí |
| 190 | button/button | accion | - | Eliminar | sí |
| 191 | button/button | accion | ' + categoriaID + ' | Eliminar | sí |
| 192 | button/button | accion | ' + productoID + ' | Eliminar | sí |
| 193 | button/button | accion | ' + escapeConfigHtml(deviceID) + ' | Eliminar | sí |
| 194 | button/button | accion | ' + id + ' | Marcar impreso | sí |
| 195 | button/button | accion | ' + id + ' | Reintentar | sí |

### `web/administrar_empresa/configuracion_menu.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | - | &#9776; Ocultar menú | sí |

### `web/administrar_empresa/configuracion_permisos.html` (5)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnReloadPermisos | Actualizar | no |
| 2 | button/button | accion | btnEnableAllEmpresaPerms | Activar todo | no |
| 3 | button/button | accion | btnSaveEmpresaPerms | Guardar techo empresa | no |
| 4 | input/checkbox | accion | - | sin etiqueta | sí |
| 5 | input/checkbox | accion | - | sin etiqueta | sí |

### `web/administrar_empresa/configuracion_rol_cajero.html` (54)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnGuardarTodo | Guardar configuracion del cajero | no |
| 2 | input/hidden | entrada | customRoleId | customRoleId | no |
| 3 | input/- | entrada | customRoleName | customRoleName | no |
| 4 | textarea/- | entrada | customRoleDescription | customRoleDescription | no |
| 5 | input/checkbox | accion | customRoleActive | customRoleActive | no |
| 6 | button/button | accion | btnGuardarPerfil | Guardar perfil | no |
| 7 | a/- | accion | linkUsuarios | Asignar usuarios | no |
| 8 | input/checkbox | accion | opRolActivo | opRolActivo | no |
| 9 | input/checkbox | accion | opRolMetodoEfectivo | opRolMetodoEfectivo | no |
| 10 | input/checkbox | accion | opRolMetodoTarjetaCredito | opRolMetodoTarjetaCredito | no |
| 11 | input/checkbox | accion | opRolMetodoTarjetaDebito | opRolMetodoTarjetaDebito | no |
| 12 | input/checkbox | accion | opRolMetodoTransferencia | opRolMetodoTransferencia | no |
| 13 | input/checkbox | accion | opRolMetodoMixto | opRolMetodoMixto | no |
| 14 | input/checkbox | accion | opRolMetodoCodigoDescuento | opRolMetodoCodigoDescuento | no |
| 15 | input/checkbox | accion | opRolHabilitarPropinas | opRolHabilitarPropinas | no |
| 16 | input/checkbox | accion | opRolHabilitarComisiones | opRolHabilitarComisiones | no |
| 17 | input/checkbox | accion | opRolPermitirIngresosManuales | opRolPermitirIngresosManuales | no |
| 18 | input/checkbox | accion | opRolPermitirEgresosManuales | opRolPermitirEgresosManuales | no |
| 19 | button/button | accion | btnGuardarOperativa | Guardar cobro y caja | no |
| 20 | input/checkbox | accion | cartSearchProducts | cartSearchProducts | no |
| 21 | input/checkbox | accion | cartBarcode | cartBarcode | no |
| 22 | input/checkbox | accion | cartClientPanel | cartClientPanel | no |
| 23 | input/checkbox | accion | cartClientRequired | cartClientRequired | no |
| 24 | input/checkbox | accion | cartPayButton | cartPayButton | no |
| 25 | input/checkbox | accion | cartPaymentCards | cartPaymentCards | no |
| 26 | input/checkbox | accion | cartTouchMode | cartTouchMode | no |
| 27 | input/checkbox | accion | cartShortcuts | cartShortcuts | no |
| 28 | input/checkbox | accion | cartScale | cartScale | no |
| 29 | input/checkbox | accion | cartOffline | cartOffline | no |
| 30 | button/button | accion | btnGuardarCarrito | Guardar carrito POS | no |
| 31 | a/- | accion | linkCarritoDetalle | Abrir configuracion completa | no |
| 32 | input/checkbox | accion | cartBtnDiscounts | cartBtnDiscounts | no |
| 33 | input/checkbox | accion | cartBtnRate | cartBtnRate | no |
| 34 | input/checkbox | accion | cartBtnTransfer | cartBtnTransfer | no |
| 35 | input/checkbox | accion | cartBtnDomotics | cartBtnDomotics | no |
| 36 | input/checkbox | accion | cartBtnCancel | cartBtnCancel | no |
| 37 | input/checkbox | accion | cartBtnTaxi | cartBtnTaxi | no |
| 38 | input/checkbox | accion | cartBtnClients | cartBtnClients | no |
| 39 | input/checkbox | accion | cartBtnCredits | cartBtnCredits | no |
| 40 | input/checkbox | accion | cartBtnVehicle | cartBtnVehicle | no |
| 41 | input/checkbox | accion | cartPayCash | cartPayCash | no |
| 42 | input/checkbox | accion | cartPayCreditCard | cartPayCreditCard | no |
| 43 | input/checkbox | accion | cartPayDebitCard | cartPayDebitCard | no |
| 44 | input/checkbox | accion | cartPayBreb | cartPayBreb | no |
| 45 | input/checkbox | accion | cartPayNequi | cartPayNequi | no |
| 46 | input/checkbox | accion | cartPayOtherTransfer | cartPayOtherTransfer | no |
| 47 | input/checkbox | accion | cartPayCustomerCredit | cartPayCustomerCredit | no |
| 48 | input/checkbox | accion | cartPayQr | cartPayQr | no |
| 49 | input/checkbox | accion | stationAccessEnabled | stationAccessEnabled | no |
| 50 | input/checkbox | accion | stationAccessCaja | stationAccessCaja | no |
| 51 | input/checkbox | accion | cashierAutoDirectSale | cashierAutoDirectSale | no |
| 52 | button/button | accion | btnGuardarEstaciones | Guardar estaciones | no |
| 53 | a/- | accion | linkEstacionesConfig | Configurar estaciones | no |
| 54 | a/- | accion | linkImpresorasConfig | Impresoras y caja | no |

### `web/administrar_empresa/configuracion_sensores_raspberry.html` (14)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/checkbox | accion | sensorAutoActivarEstacion | sensorAutoActivarEstacion | no |
| 2 | input/number | entrada | margenToleranciaEntradaMinutos | margenToleranciaEntradaMinutos | no |
| 3 | input/checkbox | accion | margenDesactivacionHabilitado | margenDesactivacionHabilitado | no |
| 4 | input/number | entrada | margenDesactivacionMinutos | margenDesactivacionMinutos | no |
| 5 | button/button | accion | saveSensorRulesBtn | Guardar reglas | no |
| 6 | input/- | entrada | deviceId | deviceId | no |
| 7 | input/number | entrada | estacionId | estacionId | no |
| 8 | input/- | entrada | deviceToken | deviceToken | no |
| 9 | button/button | accion | generateToken | Generar | no |
| 10 | button/button | accion | saveBtn | Guardar dispositivo | no |
| 11 | button/button | accion | provisionBtn | Provisionar seguro | no |
| 12 | button/button | accion | reloadBtn | Recargar estado | no |
| 13 | textarea/- | entrada | provisioningOutput | provisioningOutput | no |
| 14 | textarea/- | entrada | heartbeatExample | heartbeatExample | no |

### `web/administrar_empresa/contabilidad_colombia.html` (58)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnRefresh | Actualizar | no |
| 2 | button/button | accion | btnSeed | Cargar PUC base | no |
| 3 | button/button | accion | - | Nuevo comprobante | sí |
| 4 | button/button | accion | - | Dashboard | sí |
| 5 | button/button | accion | - | PUC | sí |
| 6 | button/button | accion | - | Terceros | sí |
| 7 | button/button | accion | - | Impuestos | sí |
| 8 | button/button | accion | - | Comprobantes | sí |
| 9 | button/button | accion | - | Cierres | sí |
| 10 | button/button | accion | - | Configuración | sí |
| 11 | input/- | entrada | ctaCodigo | ctaCodigo | no |
| 12 | input/- | entrada | ctaNombre | ctaNombre | no |
| 13 | select/- | entrada | ctaNaturaleza | Débito Crédito | no |
| 14 | select/- | entrada | ctaTipo | Auxiliar Mayor | no |
| 15 | input/- | entrada | ctaPadre | ctaPadre | no |
| 16 | input/checkbox | accion | ctaMovimiento | ctaMovimiento | no |
| 17 | input/checkbox | accion | ctaTercero | ctaTercero | no |
| 18 | input/checkbox | accion | ctaImpuesto | ctaImpuesto | no |
| 19 | button/button | accion | btnSaveCuenta | Guardar cuenta | no |
| 20 | select/- | entrada | terTipoDoc | NIT CC CE PA | no |
| 21 | input/- | entrada | terDoc | terDoc | no |
| 22 | input/- | entrada | terDV | terDV | no |
| 23 | input/- | entrada | terNombre | terNombre | no |
| 24 | select/- | entrada | terTipo | Cliente y proveedor Cliente Proveedor Empleado | no |
| 25 | select/- | entrada | terRegimen | Responsable IVA No responsable IVA Régimen simple Gran contribuyente | no |
| 26 | input/- | entrada | terEmail | terEmail | no |
| 27 | input/- | entrada | terTel | terTel | no |
| 28 | button/button | accion | btnSaveTercero | Guardar tercero | no |
| 29 | input/- | entrada | impCodigo | impCodigo | no |
| 30 | input/- | entrada | impNombre | impNombre | no |
| 31 | select/- | entrada | impTipo | IVA Retefuente ReteICA ReteIVA Autorretención | no |
| 32 | input/number | entrada | impPorcentaje | impPorcentaje | no |
| 33 | input/- | entrada | impDebito | impDebito | no |
| 34 | input/- | entrada | impCredito | impCredito | no |
| 35 | button/button | accion | btnSaveImpuesto | Guardar impuesto | no |
| 36 | select/- | entrada | compTipo | Nota contable Comprobante ingreso Comprobante egreso Causación Ajuste | no |
| 37 | input/date | entrada | compFecha | compFecha | no |
| 38 | input/- | entrada | compPeriodo | compPeriodo | no |
| 39 | input/- | entrada | compConcepto | compConcepto | no |
| 40 | button/button | accion | btnAddLinea | Agregar línea | no |
| 41 | button/button | accion | btnSaveComp | Contabilizar | no |
| 42 | input/- | entrada | cierrePeriodo | cierrePeriodo | no |
| 43 | textarea/- | entrada | cierreObs | cierreObs | no |
| 44 | button/button | accion | btnCerrarPeriodo | Cerrar periodo | no |
| 45 | button/button | accion | btnReabrirPeriodo | Reabrir periodo | no |
| 46 | input/- | entrada | cfgNombre | cfgNombre | no |
| 47 | input/- | entrada | cfgMoneda | cfgMoneda | no |
| 48 | input/- | entrada | cfgPeriodo | cfgPeriodo | no |
| 49 | input/- | entrada | cfgPuc | cfgPuc | no |
| 50 | input/- | entrada | cfgNiif | cfgNiif | no |
| 51 | input/checkbox | accion | cfgBloquear | cfgBloquear | no |
| 52 | button/button | accion | btnSaveConfig | Guardar configuración | no |
| 53 | button/button | accion | - | Anular | sí |
| 54 | select/- | entrada | - | sin etiqueta | no |
| 55 | input/number | entrada | - | 0 | no |
| 56 | input/number | entrada | - | 0 | no |
| 57 | button/button | accion | - | Quitar | no |
| 58 | input/- | entrada | - | sin etiqueta | no |

### `web/administrar_empresa/contabilidad_colombia_avanzada.html` (81)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnRefresh | Actualizar | no |
| 2 | button/button | accion | btnSeed | Cargar formatos DIAN base | no |
| 3 | button/button | accion | - | Ir a exógena | sí |
| 4 | button/button | accion | - | Dashboard | sí |
| 5 | button/button | accion | - | Exógena DIAN | sí |
| 6 | button/button | accion | - | Nómina electrónica | sí |
| 7 | button/button | accion | - | Documento soporte | sí |
| 8 | button/button | accion | - | Activos fijos | sí |
| 9 | button/button | accion | - | Activos avanzado | sí |
| 10 | button/button | accion | - | Cartera y CxP | sí |
| 11 | button/button | accion | - | Libros oficiales | sí |
| 12 | input/- | entrada | exoFormato | exoFormato | no |
| 13 | input/number | entrada | exoAnio | exoAnio | no |
| 14 | input/- | entrada | exoConcepto | exoConcepto | no |
| 15 | input/- | entrada | exoVersion | DIAN configurable | no |
| 16 | textarea/- | entrada | exoDescripcion | exoDescripcion | no |
| 17 | button/button | accion | btnSaveFormato | Guardar formato | no |
| 18 | button/button | accion | btnGenerarExogena | Generar desde contabilidad | no |
| 19 | input/number | entrada | regFormato | regFormato | no |
| 20 | input/- | entrada | regDoc | regDoc | no |
| 21 | input/- | entrada | regNombre | regNombre | no |
| 22 | input/- | entrada | regCuenta | regCuenta | no |
| 23 | input/number | entrada | regBase | regBase | no |
| 24 | input/number | entrada | regIva | regIva | no |
| 25 | input/number | entrada | regRet | regRet | no |
| 26 | input/number | entrada | regTotal | regTotal | no |
| 27 | button/button | accion | btnSaveRegistro | Guardar registro | no |
| 28 | input/- | entrada | nomPeriodo | nomPeriodo | no |
| 29 | input/date | entrada | nomFecha | nomFecha | no |
| 30 | input/- | entrada | nomDoc | nomDoc | no |
| 31 | input/- | entrada | nomNombre | nomNombre | no |
| 32 | input/number | entrada | nomSalario | nomSalario | no |
| 33 | input/number | entrada | nomDev | nomDev | no |
| 34 | input/number | entrada | nomDed | nomDed | no |
| 35 | select/- | entrada | nomEstado | borrador validado enviado rechazado | no |
| 36 | button/button | accion | btnSaveNomina | Guardar nómina electrónica | no |
| 37 | input/- | entrada | dsPeriodo | dsPeriodo | no |
| 38 | input/date | entrada | dsFecha | dsFecha | no |
| 39 | input/- | entrada | dsDoc | dsDoc | no |
| 40 | input/- | entrada | dsProveedor | dsProveedor | no |
| 41 | input/- | entrada | dsConcepto | dsConcepto | no |
| 42 | input/number | entrada | dsSubtotal | dsSubtotal | no |
| 43 | input/number | entrada | dsIva | dsIva | no |
| 44 | input/number | entrada | dsRet | dsRet | no |
| 45 | select/- | entrada | dsEstado | borrador validado enviado rechazado | no |
| 46 | button/button | accion | btnSaveSoporte | Guardar documento soporte | no |
| 47 | input/- | entrada | actCodigo | actCodigo | no |
| 48 | input/- | entrada | actNombre | actNombre | no |
| 49 | input/- | entrada | actCategoria | equipo | no |
| 50 | input/date | entrada | actFecha | actFecha | no |
| 51 | input/number | entrada | actCosto | actCosto | no |
| 52 | input/number | entrada | actResidual | actResidual | no |
| 53 | input/number | entrada | actVida | 60 | no |
| 54 | input/- | entrada | actUbicacion | actUbicacion | no |
| 55 | input/- | entrada | actCuenta | 152405 | no |
| 56 | input/- | entrada | actDep | 159205 | no |
| 57 | input/- | entrada | actGasto | 516010 | no |
| 58 | input/- | entrada | actResponsable | actResponsable | no |
| 59 | button/button | accion | btnSaveActivo | Guardar activo | no |
| 60 | input/- | entrada | actAvPeriodo | actAvPeriodo | no |
| 61 | input/number | entrada | actAvEventoID | actAvEventoID | no |
| 62 | select/- | entrada | actAvEventoTipo | Mantenimiento Traslado Baja Venta Ajuste valor libros | no |
| 63 | input/number | entrada | actAvEventoValor | actAvEventoValor | no |
| 64 | input/- | entrada | actAvDestino | actAvDestino | no |
| 65 | input/- | entrada | actAvResponsable | actAvResponsable | no |
| 66 | input/- | entrada | actAvDetalle | actAvDetalle | no |
| 67 | button/button | accion | btnActAvDep | Generar depreciacion | no |
| 68 | button/button | accion | btnActAvEvento | Registrar evento | no |
| 69 | select/- | entrada | cxTipo | Cuenta por cobrar Cuenta por pagar | no |
| 70 | input/- | entrada | cxDoc | cxDoc | no |
| 71 | input/- | entrada | cxTercero | cxTercero | no |
| 72 | input/- | entrada | cxCuenta | cxCuenta | no |
| 73 | input/- | entrada | cxConcepto | cxConcepto | no |
| 74 | input/date | entrada | cxEmision | cxEmision | no |
| 75 | input/date | entrada | cxVence | cxVence | no |
| 76 | input/number | entrada | cxValor | cxValor | no |
| 77 | input/number | entrada | cxPagado | cxPagado | no |
| 78 | button/button | accion | btnSaveCartera | Guardar obligación | no |
| 79 | select/- | entrada | libTipo | Libro diario Libro mayor / auxiliar Balance de prueba | no |
| 80 | input/- | entrada | libPeriodo | libPeriodo | no |
| 81 | button/button | accion | btnLoadLibro | Generar libro | no |

### `web/administrar_empresa/control_electrico.html` (97)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Tutorial | no |
| 2 | button/button | accion | scheduleBtn | Ejecutar agenda | no |
| 3 | button/button | accion | syncBtn | Sincronizar | no |
| 4 | button/button | accion | reloadBtn | Actualizar | no |
| 5 | input/checkbox | accion | habilitado | habilitado | no |
| 6 | input/- | entrada | raspberryIp | raspberryIp | no |
| 7 | input/number | entrada | raspberryPort | 8081 | no |
| 8 | input/- | entrada | apiPath | /api/gpio/relay | no |
| 9 | input/password | entrada | apiToken | apiToken | no |
| 10 | input/number | entrada | timeoutMs | 2500 | no |
| 11 | input/number | entrada | activationDelaySeconds | 1 | no |
| 12 | input/checkbox | accion | autoSync | autoSync | no |
| 13 | input/checkbox | accion | failSafe | failSafe | no |
| 14 | input/checkbox | accion | disconnectAlertEnabled | disconnectAlertEnabled | no |
| 15 | input/email | entrada | disconnectAlertEmail | disconnectAlertEmail | no |
| 16 | input/number | entrada | disconnectGraceMinutes | 5 | no |
| 17 | textarea/- | entrada | observaciones | observaciones | no |
| 18 | button/button | accion | saveConfigBtn | Guardar conexion | no |
| 19 | input/hidden | entrada | raspberryId | 0 | no |
| 20 | input/- | entrada | raspberryCodigo | raspberryCodigo | no |
| 21 | input/- | entrada | raspberryNombre | raspberryNombre | no |
| 22 | select/- | entrada | raspberryTipoControlador | Raspberry Pi / GPIO local Home Assistant REST Siri / Apple Home via HomeKit Bridge Matter via controlador/gateway Shelly | no |
| 23 | input/- | entrada | raspberryProveedor | raspberryProveedor | no |
| 24 | input/- | entrada | raspberryBaseUrl | raspberryBaseUrl | no |
| 25 | input/- | entrada | raspberryNodeIp | raspberryNodeIp | no |
| 26 | input/number | entrada | raspberryNodePort | 8081 | no |
| 27 | input/- | entrada | raspberryNodeApiPath | /api/gpio/relay | no |
| 28 | input/password | entrada | raspberryNodeApiToken | raspberryNodeApiToken | no |
| 29 | input/number | entrada | raspberryNodeTimeout | 2500 | no |
| 30 | textarea/- | entrada | raspberryNodeObservaciones | raspberryNodeObservaciones | no |
| 31 | button/button | accion | newRaspberryBtn | Nuevo controlador | no |
| 32 | button/button | accion | saveRaspberryBtn | Guardar controlador | no |
| 33 | input/hidden | entrada | ruleId | 0 | no |
| 34 | input/- | entrada | ruleNombre | ruleNombre | no |
| 35 | input/- | entrada | ruleSensorCodigo | ruleSensorCodigo | no |
| 36 | select/- | entrada | ruleRaspberry | Sensor externo / sin GPIO | no |
| 37 | input/number | entrada | ruleGPIO | ruleGPIO | no |
| 38 | select/- | entrada | rulePull | Sin pull Pull-up Pull-down | no |
| 39 | input/number | entrada | ruleDebounce | 250 | no |
| 40 | select/- | entrada | ruleCondicion | Igual a Distinto de Mayor que Menor que Contiene | no |
| 41 | input/- | entrada | ruleValor | ruleValor | no |
| 42 | select/- | entrada | ruleAccion | Encender aparato Encender con temporizador Activar según la programación del aparato Apagar aparato Solo alarma | no |
| 43 | input/number | entrada | ruleTimerSeconds | 900 | no |
| 44 | select/- | entrada | ruleRele | ruleRele | no |
| 45 | select/- | entrada | ruleSeveridad | Info Advertencia Critica | no |
| 46 | input/checkbox | accion | ruleAlarma | ruleAlarma | no |
| 47 | textarea/- | entrada | ruleMensaje | ruleMensaje | no |
| 48 | button/button | accion | newRuleBtn | Nueva regla | no |
| 49 | button/button | accion | saveRuleBtn | Guardar regla | no |
| 50 | select/- | entrada | reportCategoryFilter | Todas | no |
| 51 | select/- | entrada | reportDeviceFilter | Todos | no |
| 52 | button/button | accion | reportRefreshBtn | Actualizar reporte | no |
| 53 | button/button | accion | reportPrintBtn | Imprimir reporte | no |
| 54 | input/- | entrada | - | ' + sanitize(data.imagen_url) + ' | no |
| 55 | input/file | accion | - | sin etiqueta | no |
| 56 | button/button | accion | - | Subir foto | no |
| 57 | input/number | entrada | - | ' + sanitize(data.gpio_pin) + ' | no |
| 58 | input/number | entrada | - | ' + sanitize(data.pulso_ms) + ' | no |
| 59 | input/- | entrada | - | ' + sanitize(data.salida_codigo) + ' | no |
| 60 | select/- | entrada | - | Lampara / luces Motobomba Jacuzzi Aire Puerta Otro | no |
| 61 | input/- | entrada | - | ' + sanitize(data.categoria \|\| '') + ' | no |
| 62 | select/- | entrada | - | ' + relayIntegrationOptions(data.integracion_tipo) + ' | no |
| 63 | select/- | entrada | - | ' + raspberryOptions(data.raspberry_id) + ' | no |
| 64 | input/- | entrada | - | ' + sanitize(data.fabricante) + ' | no |
| 65 | input/- | entrada | - | ' + sanitize(data.modelo) + ' | no |
| 66 | input/- | entrada | - | ' + sanitize(data.entity_id) + ' | no |
| 67 | input/- | entrada | - | ' + sanitize(data.device_id) + ' | no |
| 68 | input/- | entrada | - | ' + sanitize(data.capability) + ' | no |
| 69 | input/- | entrada | - | ' + sanitize(data.relay_name) + ' | no |
| 70 | textarea/- | entrada | - | ' + sanitize(data.observaciones) + ' | no |
| 71 | input/checkbox | accion | - | sin etiqueta | no |
| 72 | input/number | entrada | - | ' + sanitize(data.potencia_w) + ' | no |
| 73 | input/- | entrada | - | ' + sanitize(data.sensor_consumo_entity_id) + ' | no |
| 74 | input/- | entrada | - | ' + sanitize((data.ultimo_consumo_w \|\| 0) + ' W / ' + (data.ultimo_consumo_kwh \|\| 0) + ' kWh / ' + (data.ultimo_voltaje_ | no |
| 75 | input/checkbox | accion | - | sin etiqueta | no |
| 76 | input/checkbox | accion | - | sin etiqueta | no |
| 77 | input/checkbox | accion | - | sin etiqueta | no |
| 78 | input/time | entrada | - | ' + sanitize(data.hora_encendido) + ' | no |
| 79 | input/time | entrada | - | ' + sanitize(data.hora_apagado) + ' | no |
| 80 | select/- | entrada | - | ' + relayDaysOptions(data.programacion_dias) + ' | no |
| 81 | input/- | entrada | - | ' + sanitize(data.programacion_timezone \|\| 'America/Bogota') + ' | no |
| 82 | input/- | entrada | - | ON ' + sanitize(data.ultima_programacion_on \|\| '--') + ' / OFF ' + sanitize(data.ultima_programacion_off \|\| '--') + ' | no |
| 83 | input/checkbox | accion | - | sin etiqueta | no |
| 84 | button/button | accion | - | Apagar | no |
| 85 | button/button | accion | - | Encender | no |
| 86 | button/button | accion | - | Guardar | no |
| 87 | button/button | accion | - | GPIO ' + pin + ' | sí |
| 88 | button/button | accion | - | Editar | no |
| 89 | button/button | accion | - | Generar instalador | no |
| 90 | button/button | accion | - | Probar conexión | no |
| 91 | button/button | accion | - | Reiniciar | no |
| 92 | button/button | accion | - | Apagar | no |
| 93 | button/button | accion | - | Probar GPIO | no |
| 94 | button/button | accion | - | Principal | no |
| 95 | button/button | accion | - | Desactivar | no |
| 96 | button/button | accion | - | Editar | no |
| 97 | button/button | accion | - | Desactivar | no |

### `web/administrar_empresa/corte_de_caja.html` (19)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | btnVolverEstaciones | Regresar a estaciones | no |
| 2 | input/datetime-local | entrada | desde | desde | no |
| 3 | input/datetime-local | entrada | hasta | hasta | no |
| 4 | input/- | entrada | usuario | usuario | no |
| 5 | input/- | entrada | cajaCodigo | CAJA-PRINCIPAL | no |
| 6 | select/- | entrada | turno | General Manana Tarde Noche Madrugada | no |
| 7 | input/number | entrada | apertura | 0 | no |
| 8 | input/number | entrada | cajaFisica | 0 | no |
| 9 | select/- | entrada | modoVista | Ticket POS 80mm Carta completo Ejecutivo compacto | no |
| 10 | input/number | entrada | umbralIncidencia | 1000 | no |
| 11 | input/- | entrada | observaciones | observaciones | no |
| 12 | button/button | accion | btnGenerar | Generar corte | no |
| 13 | button/button | accion | btnReporteMiTurno | Ver reporte de mi turno | no |
| 14 | button/button | accion | btnCorteAutomatico | Corte automatico | no |
| 15 | button/button | accion | btnCerrar | Guardar cierre | no |
| 16 | button/button | accion | btnCerrarImprimirSesion | Cerrar turno e imprimir reporte | no |
| 17 | button/button | accion | btnImprimir | Imprimir seleccion | no |
| 18 | button/button | accion | btnTutorialTurno | Tutorial | no |
| 19 | input/checkbox | accion | mantenerSesionCierre | mantenerSesionCierre | no |

### `web/administrar_empresa/creditos.html` (95)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnHeroConsultar | Actualizar panel | no |
| 2 | button/button | accion | btnHeroCrear | Nuevo crédito | no |
| 3 | button/button | accion | btnHeroExportar | Exportar cartera | no |
| 4 | input/number | entrada | crClienteId | crClienteId | no |
| 5 | input/text | entrada | crClienteNombre | crClienteNombre | no |
| 6 | select/- | entrada | crTipo | Cuotas Fijo Rotativo | no |
| 7 | input/number | entrada | crMonto | crMonto | no |
| 8 | input/number | entrada | crCupo | crCupo | no |
| 9 | input/number | entrada | crTasaInteres | 0 | no |
| 10 | input/number | entrada | crTasaMora | 0 | no |
| 11 | select/- | entrada | crPeriodicidad | Mensual Diaria Semanal Quincenal | no |
| 12 | input/number | entrada | crValorCuota | crValorCuota | no |
| 13 | input/number | entrada | crPlazoDias | 0 | no |
| 14 | input/number | entrada | crPlazoCuotas | 12 | no |
| 15 | input/date | entrada | crFechaInicio | crFechaInicio | no |
| 16 | input/date | entrada | crFechaVencimiento | crFechaVencimiento | no |
| 17 | input/checkbox | accion | crBloqueoMora | crBloqueoMora | no |
| 18 | input/checkbox | accion | crOmitirDomingos | crOmitirDomingos | no |
| 19 | input/text | entrada | crObservaciones | crObservaciones | no |
| 20 | button/button | accion | btnCrearCredito | Crear crédito | no |
| 21 | button/button | accion | btnLimpiarCrear | Limpiar | no |
| 22 | input/text | entrada | fltQ | fltQ | no |
| 23 | input/number | entrada | fltClienteId | fltClienteId | no |
| 24 | select/- | entrada | fltEstadoCredito | Estado del crédito (todos) Activo Suspendido Cerrado Castigado | no |
| 25 | select/- | entrada | fltClasificacion | Clasificación (todas) Al día Vencido Castigado | no |
| 26 | input/date | entrada | fltDesde | Desde | no |
| 27 | input/date | entrada | fltHasta | Hasta | no |
| 28 | input/checkbox | accion | fltSoloVencidos | fltSoloVencidos | no |
| 29 | input/checkbox | accion | fltIncludeInactive | fltIncludeInactive | no |
| 30 | select/- | entrada | expTipo | Reporte de cartera Reporte de morosidad | no |
| 31 | select/- | entrada | expFormat | JSON CSV TXT XLS PDF | no |
| 32 | button/button | accion | btnConsultar | Consultar cartera | no |
| 33 | button/button | accion | btnExportar | Exportar reporte | no |
| 34 | button/button | accion | btnLimpiarFiltros | Limpiar filtros | no |
| 35 | input/number | entrada | limClienteId | limClienteId | no |
| 36 | input/number | entrada | limSaldoTotal | limSaldoTotal | no |
| 37 | input/number | entrada | limMaxActivos | 1 | no |
| 38 | input/checkbox | accion | limReqAprobacion | limReqAprobacion | no |
| 39 | input/text | entrada | limObservaciones | limObservaciones | no |
| 40 | button/button | accion | btnGuardarLimite | Guardar límite | no |
| 41 | button/button | accion | btnLimpiarLimite | Limpiar | no |
| 42 | input/number | entrada | alDiasProximos | 7 | no |
| 43 | input/number | entrada | alTop | 10 | no |
| 44 | input/checkbox | accion | alIncludeInactive | alIncludeInactive | no |
| 45 | button/button | accion | btnConsultarAlertas | Consultar alertas | no |
| 46 | button/button | accion | btnExportarAlertas | Exportar morosidad | no |
| 47 | input/number | entrada | wfCreditoId | wfCreditoId | no |
| 48 | select/- | entrada | wfTipo | Tipo de solicitud (todas) Reverso de abono Refinanciación | no |
| 49 | select/- | entrada | wfEstado | Estado (todos) Pendiente aprobación Aprobada Ejecutada Rechazada Cancelada | no |
| 50 | input/checkbox | accion | wfIncludeInactive | wfIncludeInactive | no |
| 51 | button/button | accion | btnConsultarWorkflows | Consultar workflows | no |
| 52 | button/button | accion | btnAprobarWorkflow | Aprobar seleccionado | no |
| 53 | button/button | accion | btnRechazarWorkflow | Rechazar seleccionado | no |
| 54 | input/text | entrada | wfApprovedBy | wfApprovedBy | no |
| 55 | input/text | entrada | wfApprovalCode | wfApprovalCode | no |
| 56 | input/text | entrada | wfApprovalReason | wfApprovalReason | no |
| 57 | input/number | entrada | abCreditoId | abCreditoId | no |
| 58 | input/number | entrada | abMonto | abMonto | no |
| 59 | select/- | entrada | abMetodo | No especificado Efectivo Transferencia bancaria Tarjeta crédito Tarjeta débito Pasarela | no |
| 60 | input/text | entrada | abReferencia | abReferencia | no |
| 61 | input/text | entrada | abComprobante | abComprobante | no |
| 62 | input/text | entrada | abObservaciones | abObservaciones | no |
| 63 | button/button | accion | btnAplicarAbono | Aplicar abono | no |
| 64 | input/number | entrada | rfCreditoId | rfCreditoId | no |
| 65 | input/number | entrada | rfNivel | 1 | no |
| 66 | input/number | entrada | rfPlazo | 12 | no |
| 67 | input/number | entrada | rfTasa | 0 | no |
| 68 | input/text | entrada | rfMotivo | rfMotivo | no |
| 69 | input/text | entrada | rfObservaciones | rfObservaciones | no |
| 70 | button/button | accion | btnSolicitarRefinanciacion | Solicitar refinanciación | no |
| 71 | button/button | accion | btnSolicitarReverso | Solicitar reverso desde movimiento | no |
| 72 | input/number | entrada | rvCreditoId | rvCreditoId | no |
| 73 | input/number | entrada | rvMovimientoId | rvMovimientoId | no |
| 74 | input/number | entrada | rvNivel | 1 | no |
| 75 | input/text | entrada | rvMotivo | rvMotivo | no |
| 76 | input/text | entrada | rvObservaciones | rvObservaciones | no |
| 77 | input/number | entrada | estadoCreditoId | estadoCreditoId | no |
| 78 | button/button | accion | btnConsultarEstadoCuenta | Consultar estado de cuenta | no |
| 79 | button/- | accion | ' + id + ' | Ver | sí |
| 80 | button/- | accion | ' + id + ' | Cerrar | sí |
| 81 | button/- | accion | - | Cerrar | no |
| 82 | button/- | accion | ' + id + ' | Reactivar | sí |
| 83 | button/- | accion | ' + id + ' | Suspender | sí |
| 84 | button/- | accion | ' + id + ' | Activar fila | sí |
| 85 | button/- | accion | ' + id + ' | Inactivar fila | sí |
| 86 | button/- | accion | ' + id + ' | Paz y salvo | sí |
| 87 | button/- | accion | - | Paz y salvo | no |
| 88 | button/- | accion | ' + id + ' | Estado cuenta | sí |
| 89 | button/- | accion | ' + id + ' | Abono | sí |
| 90 | button/- | accion | ' + id + ' | Refinanciar | sí |
| 91 | button/- | accion | ' + movID + ' | Reverso | sí |
| 92 | button/- | accion | ' + i(row.cliente_id) + ' | Editar | sí |
| 93 | button/- | accion | ' + i(row.cliente_id) + ' | Inactivar | sí |
| 94 | input/radio | accion | - | ' + id + ' | no |
| 95 | button/- | accion | ' + id + ' | Seleccionar | sí |

### `web/administrar_empresa/creditos_menu.html` (9)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | linkCreditosPanelMenu | Panel | sí |
| 2 | a/- | accion | linkCreditosCrearMenu | Nuevo credito | sí |
| 3 | a/- | accion | linkCreditosCarteraMenu | Cartera | sí |
| 4 | a/- | accion | linkCreditosMorosidadMenu | Morosos | sí |
| 5 | a/- | accion | linkCreditosLimitesMenu | Riesgo y limites | sí |
| 6 | a/- | accion | linkCreditosOperacionesMenu | Abonos y operaciones | sí |
| 7 | a/- | accion | linkCreditosAprobacionesMenu | Aprobaciones | sí |
| 8 | a/- | accion | linkCreditosEstadoMenu | Estado de cuenta | sí |
| 9 | a/- | accion | linkCreditosTutorialMenu | Tutorial | sí |

### `web/administrar_empresa/creditos_tutorial.html` (13)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Ir al panel | sí |
| 2 | a/- | accion | - | Abrir riesgo y limites | sí |
| 3 | a/- | accion | - | Crear credito | sí |
| 4 | a/- | accion | - | Abrir abonos | sí |
| 5 | a/- | accion | - | Ver estado de cuenta | sí |
| 6 | a/- | accion | - | Revisar morosos | sí |
| 7 | a/- | accion | - | Gestion de cobranza | no |
| 8 | a/- | accion | - | Abrir aprobaciones | sí |
| 9 | a/- | accion | - | Abrir cartera | sí |
| 10 | input/checkbox | accion | - | sin etiqueta | no |
| 11 | input/checkbox | accion | - | sin etiqueta | no |
| 12 | input/checkbox | accion | - | sin etiqueta | no |
| 13 | input/checkbox | accion | - | sin etiqueta | no |

### `web/administrar_empresa/crm_comercial.html` (92)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | crmRefreshBtn | Actualizar | no |
| 2 | button/button | accion | crmSeedBtn | Cargar demo | no |
| 3 | button/button | accion | crmExportBtn | Exportar vista | no |
| 4 | input/search | entrada | crmSearch | crmSearch | no |
| 5 | input/month | entrada | crmPeriodo | crmPeriodo | no |
| 6 | input/checkbox | accion | crmIncludeInactive | crmIncludeInactive | no |
| 7 | button/button | accion | crmSearchBtn | Aplicar filtros | no |
| 8 | button/button | accion | - | Tablero ejecutivo | sí |
| 9 | button/button | accion | - | Leads | sí |
| 10 | button/button | accion | - | Seguimientos | sí |
| 11 | button/button | accion | - | Cotizaciones | sí |
| 12 | button/button | accion | - | Campanas | sí |
| 13 | button/button | accion | - | Forecast | sí |
| 14 | button/button | accion | - | Metas y conversion | sí |
| 15 | button/button | accion | - | Embudo documental | sí |
| 16 | button/button | accion | - | Ayuda | sí |
| 17 | input/hidden | entrada | leadId | leadId | no |
| 18 | input/text | entrada | leadCodigo | leadCodigo | no |
| 19 | select/- | entrada | leadEstado | nuevo contactado calificado propuesta negociacion ganado perdido descalificado reactivado postventa cerrado | no |
| 20 | input/text | entrada | leadNombre | leadNombre | no |
| 21 | input/text | entrada | leadEmpresaOrigen | leadEmpresaOrigen | no |
| 22 | input/email | entrada | leadEmail | leadEmail | no |
| 23 | input/text | entrada | leadTelefono | leadTelefono | no |
| 24 | input/text | entrada | leadCanal | leadCanal | no |
| 25 | input/text | entrada | leadPropietario | leadPropietario | no |
| 26 | input/number | entrada | leadValor | leadValor | no |
| 27 | input/number | entrada | leadProbabilidad | leadProbabilidad | no |
| 28 | input/datetime-local | entrada | leadProximoContacto | leadProximoContacto | no |
| 29 | textarea/- | entrada | leadNotas | leadNotas | no |
| 30 | textarea/- | entrada | leadObservaciones | leadObservaciones | no |
| 31 | button/submit | accion | leadSaveBtn | Guardar lead | no |
| 32 | button/button | accion | leadCancelBtn | Cancelar edicion | no |
| 33 | button/reset | accion | leadResetBtn | Limpiar | no |
| 34 | input/hidden | entrada | interaccionId | interaccionId | no |
| 35 | input/text | entrada | interaccionCodigo | interaccionCodigo | no |
| 36 | select/- | entrada | interaccionEstado | abierta en_progreso cerrada reabierta cancelada | no |
| 37 | select/- | entrada | interaccionLeadId | interaccionLeadId | no |
| 38 | input/number | entrada | interaccionClienteId | interaccionClienteId | no |
| 39 | select/- | entrada | interaccionTipo | seguimiento llamada correo reunion whatsapp visita | no |
| 40 | input/datetime-local | entrada | interaccionFecha | interaccionFecha | no |
| 41 | input/text | entrada | interaccionResponsable | interaccionResponsable | no |
| 42 | input/text | entrada | interaccionResultado | interaccionResultado | no |
| 43 | textarea/- | entrada | interaccionResumen | interaccionResumen | no |
| 44 | textarea/- | entrada | interaccionProximaAccion | interaccionProximaAccion | no |
| 45 | textarea/- | entrada | interaccionObservaciones | interaccionObservaciones | no |
| 46 | button/submit | accion | interaccionSaveBtn | Guardar seguimiento | no |
| 47 | button/button | accion | interaccionCancelBtn | Cancelar edicion | no |
| 48 | button/reset | accion | interaccionResetBtn | Limpiar | no |
| 49 | input/hidden | entrada | cotizacionId | cotizacionId | no |
| 50 | input/text | entrada | cotizacionCodigo | cotizacionCodigo | no |
| 51 | select/- | entrada | cotizacionEstado | borrador emitida aprobada rechazada vencida convertida anulada | no |
| 52 | input/text | entrada | cotizacionClienteNombre | cotizacionClienteNombre | no |
| 53 | input/number | entrada | cotizacionClienteId | cotizacionClienteId | no |
| 54 | input/date | entrada | cotizacionFecha | cotizacionFecha | no |
| 55 | input/date | entrada | cotizacionVigencia | cotizacionVigencia | no |
| 56 | input/number | entrada | cotizacionSubtotal | cotizacionSubtotal | no |
| 57 | input/number | entrada | cotizacionDescuento | cotizacionDescuento | no |
| 58 | input/number | entrada | cotizacionImpuesto | cotizacionImpuesto | no |
| 59 | input/number | entrada | cotizacionTotal | cotizacionTotal | no |
| 60 | input/text | entrada | cotizacionMoneda | COP | no |
| 61 | input/text | entrada | cotizacionOrigen | cotizacionOrigen | no |
| 62 | textarea/- | entrada | cotizacionNotas | cotizacionNotas | no |
| 63 | textarea/- | entrada | cotizacionObservaciones | cotizacionObservaciones | no |
| 64 | button/submit | accion | cotizacionSaveBtn | Guardar cotizacion | no |
| 65 | button/button | accion | cotizacionCancelBtn | Cancelar edicion | no |
| 66 | button/reset | accion | cotizacionResetBtn | Limpiar | no |
| 67 | input/hidden | entrada | campanaId | campanaId | no |
| 68 | input/text | entrada | campanaCodigo | campanaCodigo | no |
| 69 | select/- | entrada | campanaEstado | planificada activa pausada finalizada archivada cancelada | no |
| 70 | input/text | entrada | campanaNombre | campanaNombre | no |
| 71 | select/- | entrada | campanaCanal | email whatsapp redes llamadas evento mixto | no |
| 72 | input/number | entrada | campanaPresupuesto | campanaPresupuesto | no |
| 73 | textarea/- | entrada | campanaObjetivo | campanaObjetivo | no |
| 74 | textarea/- | entrada | campanaAudiencia | campanaAudiencia | no |
| 75 | input/date | entrada | campanaFechaInicio | campanaFechaInicio | no |
| 76 | input/date | entrada | campanaFechaFin | campanaFechaFin | no |
| 77 | input/text | entrada | campanaKPI | campanaKPI | no |
| 78 | textarea/- | entrada | campanaResultados | campanaResultados | no |
| 79 | textarea/- | entrada | campanaObservaciones | campanaObservaciones | no |
| 80 | button/submit | accion | campanaSaveBtn | Guardar campana | no |
| 81 | button/button | accion | campanaCancelBtn | Cancelar edicion | no |
| 82 | button/reset | accion | campanaResetBtn | Limpiar | no |
| 83 | input/month | entrada | metaPeriodo | metaPeriodo | no |
| 84 | input/text | entrada | metaPropietario | metaPropietario | no |
| 85 | input/text | entrada | metaCanal | metaCanal | no |
| 86 | input/number | entrada | metaValor | metaValor | no |
| 87 | input/number | entrada | metaLeads | metaLeads | no |
| 88 | input/number | entrada | metaConv | metaConv | no |
| 89 | button/submit | accion | btnSaveMeta | Guardar meta | no |
| 90 | select/- | entrada | leadConvertir | leadConvertir | no |
| 91 | input/text | entrada | cotCodigo | cotCodigo | no |
| 92 | button/button | accion | btnConvertLead | Generar cotizacion | no |

### `web/administrar_empresa/declaraciones_tributarias.html` (56)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | dtRefresh | Actualizar | no |
| 2 | button/button | accion | dtSeed | Cargar demo | no |
| 3 | button/button | accion | dtExport | Exportar CSV | no |
| 4 | button/button | accion | - | Dashboard | sí |
| 5 | button/button | accion | - | Preliquidar | sí |
| 6 | button/button | accion | - | Declaraciones | sí |
| 7 | button/button | accion | - | Calendario | sí |
| 8 | button/button | accion | - | Movimientos | sí |
| 9 | select/- | entrada | preTipo | IVA Retención en la fuente Retención IVA ICA ReteICA Impuesto al consumo Régimen simple Renta | no |
| 10 | input/- | entrada | prePeriodo | prePeriodo | no |
| 11 | button/submit | accion | - | Generar preliquidación | no |
| 12 | input/hidden | entrada | declaId | declaId | no |
| 13 | select/- | entrada | declaTipo | IVA Retención en la fuente ReteIVA ICA ReteICA Consumo Régimen simple Renta | no |
| 14 | input/- | entrada | declaPeriodo | declaPeriodo | no |
| 15 | input/date | entrada | declaDesde | declaDesde | no |
| 16 | input/date | entrada | declaHasta | declaHasta | no |
| 17 | input/date | entrada | declaVence | declaVence | no |
| 18 | input/- | entrada | declaNit | declaNit | no |
| 19 | input/- | entrada | declaMunicipio | declaMunicipio | no |
| 20 | input/- | entrada | declaFormulario | declaFormulario | no |
| 21 | input/number | entrada | declaIngresos | 0 | no |
| 22 | input/number | entrada | declaExcluidos | 0 | no |
| 23 | input/number | entrada | declaCompras | 0 | no |
| 24 | input/number | entrada | declaIvaGen | 0 | no |
| 25 | input/number | entrada | declaIvaDesc | 0 | no |
| 26 | input/number | entrada | declaConsumo | 0 | no |
| 27 | input/number | entrada | declaReteFuente | 0 | no |
| 28 | input/number | entrada | declaReteIva | 0 | no |
| 29 | input/number | entrada | declaReteIca | 0 | no |
| 30 | input/number | entrada | declaAuto | 0 | no |
| 31 | input/number | entrada | declaAnticipo | 0 | no |
| 32 | input/number | entrada | declaSaldoAnterior | 0 | no |
| 33 | input/number | entrada | declaSanciones | 0 | no |
| 34 | input/number | entrada | declaIntereses | 0 | no |
| 35 | select/- | entrada | declaEstado | Borrador Revisada Presentada Pagada Vencida Anulada | no |
| 36 | input/date | entrada | declaPresentacion | declaPresentacion | no |
| 37 | input/date | entrada | declaPago | declaPago | no |
| 38 | input/- | entrada | declaSoporte | declaSoporte | no |
| 39 | input/- | entrada | declaRecibo | declaRecibo | no |
| 40 | textarea/- | entrada | declaObs | declaObs | no |
| 41 | button/submit | accion | - | Guardar declaración | no |
| 42 | input/hidden | entrada | calId | calId | no |
| 43 | select/- | entrada | calTipo | IVA Retención fuente ReteIVA ICA ReteICA Consumo Régimen simple Renta | no |
| 44 | input/number | entrada | calAnio | calAnio | no |
| 45 | input/- | entrada | calPeriodo | calPeriodo | no |
| 46 | select/- | entrada | calPeriodicidad | Mensual Bimestral Cuatrimestral Trimestral Semestral Anual | no |
| 47 | input/date | entrada | calDesde | calDesde | no |
| 48 | input/date | entrada | calHasta | calHasta | no |
| 49 | input/date | entrada | calVence | calVence | no |
| 50 | input/number | entrada | calNitDesde | 0 | no |
| 51 | input/number | entrada | calNitHasta | 9 | no |
| 52 | select/- | entrada | calEstado | Activo Inactivo | no |
| 53 | textarea/- | entrada | calObs | calObs | no |
| 54 | button/submit | accion | - | Guardar vencimiento | no |
| 55 | button/button | accion | - | Editar | sí |
| 56 | button/button | accion | - | Editar | sí |

### `web/administrar_empresa/documentos_onlyoffice.html` (10)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | downloadBtnTop | Guardar en este dispositivo | no |
| 2 | button/button | accion | newBtnTop | Nuevo documento | no |
| 3 | input/text | entrada | docName | docName | no |
| 4 | button/button | accion | - | Documento Word .docx | sí |
| 5 | button/button | accion | - | Hoja Excel .xlsx | sí |
| 6 | button/button | accion | - | Presentacion .pptx | sí |
| 7 | button/button | accion | createBtn | Crear y abrir editor | no |
| 8 | button/button | accion | downloadBtn | Guardar en este dispositivo | no |
| 9 | button/button | accion | closeBtn | Cerrar | no |
| 10 | button/button | accion | - | Abrir | sí |

### `web/administrar_empresa/domicilios.html` (61)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | openCustomerPortal | Cliente | no |
| 2 | a/- | accion | openRestaurantPortal | Restaurante | no |
| 3 | a/- | accion | openCourierPortal | Domiciliario | no |
| 4 | button/- | accion | - | Resumen operativo Estado, cola y accesos críticos | sí |
| 5 | button/- | accion | - | Central Pedidos, despacho y mapa | sí |
| 6 | button/- | accion | - | Restaurantes Aliados, PIN y operación | sí |
| 7 | button/- | accion | - | Domiciliarios Flota, vehículo y disponibilidad | sí |
| 8 | button/- | accion | - | Menú Productos por restaurante | sí |
| 9 | button/- | accion | - | Configuración Tarifas, radios y reglas | sí |
| 10 | button/button | accion | - | Ir a central | sí |
| 11 | button/button | accion | - | Ver restaurantes | sí |
| 12 | button/button | accion | - | Ver domiciliarios | sí |
| 13 | button/button | accion | refreshBtn | Actualizar | no |
| 14 | button/button | accion | seedBtn | Cargar demo productiva | no |
| 15 | input/hidden | entrada | restId | restId | no |
| 16 | input/- | entrada | restCodigo | restCodigo | no |
| 17 | input/- | entrada | restPin | restPin | no |
| 18 | input/- | entrada | restNombre | restNombre | no |
| 19 | input/- | entrada | restCategoria | restCategoria | no |
| 20 | input/- | entrada | restTelefono | restTelefono | no |
| 21 | input/- | entrada | restDireccion | restDireccion | no |
| 22 | input/- | entrada | restLat | restLat | no |
| 23 | input/- | entrada | restLng | restLng | no |
| 24 | input/number | entrada | restPrep | 20 | no |
| 25 | input/checkbox | accion | restAcepta | restAcepta | no |
| 26 | button/submit | accion | - | Guardar restaurante | no |
| 27 | button/button | accion | restNew | Nuevo | no |
| 28 | input/hidden | entrada | courierId | courierId | no |
| 29 | input/- | entrada | courierCodigo | courierCodigo | no |
| 30 | input/- | entrada | courierPin | courierPin | no |
| 31 | input/- | entrada | courierNombre | courierNombre | no |
| 32 | input/- | entrada | courierDocumento | courierDocumento | no |
| 33 | input/- | entrada | courierTelefono | courierTelefono | no |
| 34 | select/- | entrada | courierVehiculo | Moto Bicicleta Carro A pie | no |
| 35 | input/- | entrada | courierPlaca | courierPlaca | no |
| 36 | button/submit | accion | - | Guardar domiciliario | no |
| 37 | button/button | accion | courierNew | Nuevo | no |
| 38 | input/hidden | entrada | menuId | menuId | no |
| 39 | select/- | entrada | menuRestaurant | menuRestaurant | no |
| 40 | input/- | entrada | menuCodigo | menuCodigo | no |
| 41 | input/- | entrada | menuNombre | menuNombre | no |
| 42 | input/- | entrada | menuCategoria | menuCategoria | no |
| 43 | input/number | entrada | menuPrecio | 0 | no |
| 44 | textarea/- | entrada | menuDescripcion | menuDescripcion | no |
| 45 | input/- | entrada | menuImagen | menuImagen | no |
| 46 | input/checkbox | accion | menuDisponible | menuDisponible | no |
| 47 | button/submit | accion | - | Guardar producto | no |
| 48 | button/button | accion | menuNew | Nuevo | no |
| 49 | input/- | entrada | cfgNombreSistema | cfgNombreSistema | no |
| 50 | input/- | entrada | cfgNombrePortal | cfgNombrePortal | no |
| 51 | input/- | entrada | cfgMoneda | COP | no |
| 52 | input/number | entrada | cfgCobertura | cfgCobertura | no |
| 53 | input/number | entrada | cfgAsignacion | cfgAsignacion | no |
| 54 | input/number | entrada | cfgRonda | cfgRonda | no |
| 55 | input/number | entrada | cfgTarifaBase | cfgTarifaBase | no |
| 56 | input/number | entrada | cfgTarifaKm | cfgTarifaKm | no |
| 57 | input/number | entrada | cfgComision | cfgComision | no |
| 58 | input/checkbox | accion | cfgAuto | cfgAuto | no |
| 59 | input/checkbox | accion | cfgPublico | cfgPublico | no |
| 60 | input/checkbox | accion | cfgCodigo | cfgCodigo | no |
| 61 | button/submit | accion | - | Guardar configuración | no |

### `web/administrar_empresa/egresos.html` (30)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnNuevoMovimiento | Nuevo egreso | no |
| 2 | input/hidden | entrada | movimientoId | movimientoId | no |
| 3 | input/hidden | entrada | codigoMovimiento | codigoMovimiento | no |
| 4 | input/hidden | entrada | comprobanteUrl | comprobanteUrl | no |
| 5 | input/datetime-local | entrada | fechaMovimiento | fechaMovimiento | no |
| 6 | select/- | entrada | categoriaMovimiento | Compras Gastos operativos Servicios publicos Nomina Proveedores Mantenimiento Otro | no |
| 7 | select/- | entrada | metodoPago | Efectivo Transferencia bancaria Tarjeta debito Tarjeta credito Otro | no |
| 8 | input/- | entrada | moneda | COP | no |
| 9 | input/- | entrada | concepto | concepto | no |
| 10 | input/- | entrada | subcategoria | subcategoria | no |
| 11 | input/- | entrada | terceroNombre | terceroNombre | no |
| 12 | input/- | entrada | terceroDocumento | terceroDocumento | no |
| 13 | input/number | entrada | monto | monto | no |
| 14 | input/number | entrada | impuesto | 0 | no |
| 15 | input/number | entrada | totalRetenciones | 0 | no |
| 16 | input/number | entrada | totalNeto | totalNeto | no |
| 17 | select/- | entrada | tipoComprobante | Factura Recibo interno Soporte externo Nota contable | no |
| 18 | input/- | entrada | numeroComprobante | numeroComprobante | no |
| 19 | input/file | accion | comprobanteFile | comprobanteFile | no |
| 20 | button/button | accion | btnAnalizarComprobanteIA | Analizar con IA | no |
| 21 | input/- | entrada | referenciaExterna | referenciaExterna | no |
| 22 | textarea/- | entrada | descripcion | descripcion | no |
| 23 | textarea/- | entrada | observaciones | observaciones | no |
| 24 | input/checkbox | accion | imprimirAlGuardar | imprimirAlGuardar | no |
| 25 | button/submit | accion | btnGuardarMovimiento | Guardar egreso | no |
| 26 | button/button | accion | btnCancelarMovimiento | Limpiar | no |
| 27 | input/- | entrada | filtroQ | filtroQ | no |
| 28 | input/date | entrada | filtroDesde | Desde | no |
| 29 | input/date | entrada | filtroHasta | Hasta | no |
| 30 | button/button | accion | btnBuscarMovimientos | Buscar | no |

### `web/administrar_empresa/egresos_ingresos_menu.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | - | &#9776; Ocultar menú | sí |

### `web/administrar_empresa/email_corporativo.html` (2)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | emailRefreshBtn | Actualizar | no |
| 2 | a/- | accion | emailConfigLink | Configuraci&oacute;n | no |

### `web/administrar_empresa/empresas_compartidas.html` (3)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | refreshBtn | Actualizar | no |
| 2 | button/button | accion | ' + escapeHtml(item.id) + ' | Desactivar acceso | sí |
| 3 | button/button | accion | ' + escapeHtml(item.id) + ' | Cancelar invitacion | sí |

### `web/administrar_empresa/energia_solar.html` (45)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnSolarRefresh | Actualizar | no |
| 2 | button/button | accion | btnSolarTestEmail | Probar alerta | no |
| 3 | input/hidden | entrada | solarSistemaId | solarSistemaId | no |
| 4 | input/text | entrada | solarNombre | solarNombre | no |
| 5 | select/- | entrada | solarProveedor | solarProveedor | no |
| 6 | input/text | entrada | solarModelo | solarModelo | no |
| 7 | input/text | entrada | solarUbicacion | solarUbicacion | no |
| 8 | input/number | entrada | solarCapacidad | 0 | no |
| 9 | select/- | entrada | solarBateriaMarca | solarBateriaMarca | no |
| 10 | input/text | entrada | solarBateriaModelo | solarBateriaModelo | no |
| 11 | input/text | entrada | solarBateriaSerial | solarBateriaSerial | no |
| 12 | select/- | entrada | solarBms | Sin definir CAN-bus RS485 Modbus TCP MQTT API nube fabricante | no |
| 13 | input/number | entrada | solarCapacidadBateria | 0 | no |
| 14 | input/text | entrada | solarInstalacion | solarInstalacion | no |
| 15 | input/url | entrada | solarApiBase | solarApiBase | no |
| 16 | input/text | entrada | solarApiKeyRef | solarApiKeyRef | no |
| 17 | input/url | entrada | solarGateway | solarGateway | no |
| 18 | input/number | entrada | solarIntervalo | 300 | no |
| 19 | input/text | entrada | solarEmails | solarEmails | no |
| 20 | input/checkbox | accion | solarEmailActivo | solarEmailActivo | no |
| 21 | input/checkbox | accion | solarActivo | solarActivo | no |
| 22 | button/submit | accion | - | Guardar sistema | no |
| 23 | button/button | accion | btnSolarClear | Nuevo | no |
| 24 | input/hidden | entrada | solarAlertaId | solarAlertaId | no |
| 25 | select/- | entrada | solarAlertaSistema | solarAlertaSistema | no |
| 26 | select/- | entrada | solarAlertaTipo | solarAlertaTipo | no |
| 27 | input/text | entrada | solarAlertaNombre | solarAlertaNombre | no |
| 28 | select/- | entrada | solarAlertaOperador | &lt; &lt;= ">&gt; =">&gt;= Estado contiene error | no |
| 29 | input/number | entrada | solarAlertaUmbral | 0 | no |
| 30 | select/- | entrada | solarAlertaSeveridad | Media Alta Crítica Informativa | no |
| 31 | input/checkbox | accion | solarAlertaEmail | solarAlertaEmail | no |
| 32 | input/checkbox | accion | solarAlertaActiva | solarAlertaActiva | no |
| 33 | button/submit | accion | - | Guardar alerta | no |
| 34 | select/- | entrada | solarLecturaSistema | solarLecturaSistema | no |
| 35 | input/number | entrada | solarLecturaProduccion | 1200 | no |
| 36 | input/number | entrada | solarLecturaSoc | 75 | no |
| 37 | input/number | entrada | solarLecturaSoh | 95 | no |
| 38 | input/number | entrada | solarLecturaCarga | 450 | no |
| 39 | input/number | entrada | solarLecturaDescarga | 0 | no |
| 40 | input/number | entrada | solarLecturaTemp | 32 | no |
| 41 | input/number | entrada | solarLecturaCeldaMin | 3.30 | no |
| 42 | input/number | entrada | solarLecturaCeldaMax | 3.34 | no |
| 43 | input/text | entrada | solarLecturaEstadoBateria | solarLecturaEstadoBateria | no |
| 44 | input/text | entrada | solarLecturaEstadoInversor | solarLecturaEstadoInversor | no |
| 45 | button/submit | accion | - | Registrar lectura | no |

### `web/administrar_empresa/estacion_ia_pedidos.html` (5)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | textarea/- | entrada | txt | txt | no |
| 2 | select/- | entrada | agentSelect | agentSelect | no |
| 3 | select/- | entrada | modelSelect | modelSelect | no |
| 4 | button/button | accion | mic | Dictar | no |
| 5 | button/submit | accion | go | Interpretar y agregar | no |

### `web/administrar_empresa/estaciones.html` (23)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnReporteAseo | Reporte de aseo | no |
| 2 | button/button | accion | btnEstacionesThumbToggle | Ver miniaturas | no |
| 3 | button/button | accion | ' + sanitize(note.id) + ' | ' + ' ' + title + ' ' + ' ' + sanitize(statusLabel) + ' ' + ' | sí |
| 4 | button/button | accion | - | ' + ' ' + sanitize(label) + ' ' + (desc ? ' ' + sanitize(desc) + ' ' : '') + ' | sí |
| 5 | button/button | accion | - | Ver últimos movimientos | no |
| 6 | button/button | accion | youtubeStationPlay | ▶ | no |
| 7 | button/button | accion | youtubeStationExpand | [] | no |
| 8 | input/text | entrada | youtubeStationSourceInput | ' + sanitize(getYouTubeStationDisplayReference(cfg)) + ' | no |
| 9 | button/button | accion | youtubeStationSaveBtn | Guardar y cargar | no |
| 10 | a/- | accion | youtubeStationOpenLinkBtn | Abrir en YouTube | no |
| 11 | button/button | accion | youtubeStationIASearchBtn | Buscar música con IA | no |
| 12 | button/button | accion | youtubeStationIALoadBtn | Cargar sugerencia | no |
| 13 | textarea/- | entrada | notasStationTextarea | notasStationTextarea | no |
| 14 | input/number | entrada | notasStationMinutesInput | 5 | no |
| 15 | input/number | entrada | notasStationRepeatInput | 0 | no |
| 16 | button/button | accion | notasStationAddBtn | Nueva nota | no |
| 17 | button/button | accion | notasStationStartBtn | Iniciar | no |
| 18 | button/button | accion | notasStationStopBtn | Detener | no |
| 19 | button/button | accion | notasStationSaveBtn | Guardar nota | no |
| 20 | button/button | accion | notasStationDeleteBtn | Eliminar nota | no |
| 21 | a/- | accion | - | Abrir camara | no |
| 22 | button/button | accion | - | Minimizar | no |
| 23 | button/button | accion | - | Cerrar | no |

### `web/administrar_empresa/facturacion_electronica.html` (134)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | select/- | entrada | pais_codigo | País fiscal interno | no |
| 2 | button/button | accion | detectBtn | Detectar de nuevo | no |
| 3 | button/button | accion | openSelectedCountryBtn | Cargar pais | no |
| 4 | input/file | accion | feFirmaFile | feFirmaFile | no |
| 5 | input/password | entrada | feFirmaPassword | feFirmaPassword | no |
| 6 | button/button | accion | feFirmaUploadBtn | Cargar firma | no |
| 7 | a/- | accion | feFirmaProviderBtn | Adquirir Firma Electr&oacute;nica | no |
| 8 | button/button | accion | feFirmaExpiryBtn | Verificar vencimiento | no |
| 9 | input/hidden | entrada | dian_id | dian_id | no |
| 10 | input/hidden | entrada | dian_modo_operacion_descripcion | dian_modo_operacion_descripcion | no |
| 11 | input/hidden | entrada | dian_modo_operacion_fecha_inicio | dian_modo_operacion_fecha_inicio | no |
| 12 | input/hidden | entrada | dian_modo_operacion_fecha_termino | dian_modo_operacion_fecha_termino | no |
| 13 | input/hidden | entrada | dian_set_documentos_requeridos | dian_set_documentos_requeridos | no |
| 14 | input/hidden | entrada | dian_set_facturas_requeridas | dian_set_facturas_requeridas | no |
| 15 | input/hidden | entrada | dian_set_notas_debito_requeridas | dian_set_notas_debito_requeridas | no |
| 16 | input/hidden | entrada | dian_set_notas_credito_requeridas | dian_set_notas_credito_requeridas | no |
| 17 | input/hidden | entrada | dian_set_documentos_aceptados_requeridos | dian_set_documentos_aceptados_requeridos | no |
| 18 | input/hidden | entrada | dian_set_facturas_aceptadas_requeridas | dian_set_facturas_aceptadas_requeridas | no |
| 19 | input/hidden | entrada | dian_set_notas_debito_aceptadas_requeridas | dian_set_notas_debito_aceptadas_requeridas | no |
| 20 | input/hidden | entrada | dian_set_notas_credito_aceptadas_requeridas | dian_set_notas_credito_aceptadas_requeridas | no |
| 21 | input/file | accion | dianNumeracionPdfInput | dianNumeracionPdfInput | no |
| 22 | button/button | accion | btnDianNumeracionPdf | Cargar PDF con IA GPT-5.5 | sí |
| 23 | button/button | accion | btnDianNumeracionAplicar | Aplicar valores | no |
| 24 | input/- | entrada | dian_nit | dian_nit | no |
| 25 | input/- | entrada | dian_dv | dian_dv | no |
| 26 | input/- | entrada | dian_razon_social | dian_razon_social | no |
| 27 | select/- | entrada | dian_tipo_ambiente | Habilitación Producción | no |
| 28 | input/- | entrada | dian_url | dian_url | no |
| 29 | input/- | entrada | dian_test_set_id | dian_test_set_id | no |
| 30 | input/- | entrada | dian_software_id | dian_software_id | no |
| 31 | input/password | entrada | dian_software_pin | dian_software_pin | no |
| 32 | input/checkbox | accion | dian_usar_software_compartido | dian_usar_software_compartido | no |
| 33 | input/- | entrada | dian_prefijo | dian_prefijo | no |
| 34 | input/- | entrada | dian_resolucion_numero | dian_resolucion_numero | no |
| 35 | input/number | entrada | dian_consecutivo_actual | 1 | no |
| 36 | input/date | entrada | dian_resolucion_fecha_desde | dian_resolucion_fecha_desde | no |
| 37 | input/date | entrada | dian_resolucion_fecha_hasta | dian_resolucion_fecha_hasta | no |
| 38 | input/number | entrada | dian_rango_desde | dian_rango_desde | no |
| 39 | input/number | entrada | dian_rango_hasta | dian_rango_hasta | no |
| 40 | input/number | entrada | dian_resolucion_alerta_dias | 30 | no |
| 41 | button/button | accion | btnDianResolucionVencimiento | Verificar vencimiento resoluci&oacute;n | no |
| 42 | input/- | entrada | dian_llave_tecnica | dian_llave_tecnica | no |
| 43 | input/- | entrada | dian_token_emisor_ref | dian_token_emisor_ref | no |
| 44 | input/- | entrada | dian_certificado_url | dian_certificado_url | no |
| 45 | input/- | entrada | dian_certificado_clave_ref | dian_certificado_clave_ref | no |
| 46 | textarea/- | entrada | dian_observaciones | dian_observaciones | no |
| 47 | button/submit | accion | btnDianConfigGuardar | Guardar DIAN Colombia | no |
| 48 | button/button | accion | btnDianConfigValidar | Validar checklist | no |
| 49 | button/button | accion | btnDianConfigRecargar | Recargar | no |
| 50 | a/- | accion | btnDianAyudaNumeracion | Ayuda numeración DIAN | no |
| 51 | a/- | accion | btnDianAsociarNumeracion | Asociar numeraci&oacute;n DIAN | no |
| 52 | a/- | accion | btnDianPasarTest | Centro de habilitación DIAN | no |
| 53 | button/button | accion | feDianSelectAllDocs | Activar todos | no |
| 54 | button/button | accion | feDianApplyDocs | Aplicar seleccion al JSON | no |
| 55 | input/- | entrada | fe_ec_ruc | fe_ec_ruc | no |
| 56 | input/- | entrada | fe_ec_est | fe_ec_est | no |
| 57 | input/- | entrada | fe_ec_punto | fe_ec_punto | no |
| 58 | select/- | entrada | fe_ec_ambiente_sri | (definir en JSON o aquí) 1 — Pruebas 2 — Producción | no |
| 59 | input/- | entrada | fe_pa_ruc | fe_pa_ruc | no |
| 60 | input/- | entrada | fe_pa_dv | fe_pa_dv | no |
| 61 | input/- | entrada | fe_cr_cedula | fe_cr_cedula | no |
| 62 | input/- | entrada | fe_cr_tipo_identificacion | fe_cr_tipo_identificacion | no |
| 63 | input/- | entrada | fe_cr_actividad | fe_cr_actividad | no |
| 64 | input/- | entrada | fe_cr_sucursal | fe_cr_sucursal | no |
| 65 | input/- | entrada | fe_cr_terminal | fe_cr_terminal | no |
| 66 | input/- | entrada | fe_cr_version_xml | fe_cr_version_xml | no |
| 67 | input/- | entrada | fe_ar_cuit | fe_ar_cuit | no |
| 68 | input/- | entrada | fe_ar_punto_venta | fe_ar_punto_venta | no |
| 69 | input/- | entrada | fe_ar_condicion_iva | fe_ar_condicion_iva | no |
| 70 | input/- | entrada | fe_ar_tipo_comprobante | fe_ar_tipo_comprobante | no |
| 71 | input/- | entrada | fe_ar_ws_servicio | fe_ar_ws_servicio | no |
| 72 | input/- | entrada | fe_ve_rif | fe_ve_rif | no |
| 73 | input/- | entrada | fe_ve_serie | fe_ve_serie | no |
| 74 | input/- | entrada | fe_ve_moneda_ref | fe_ve_moneda_ref | no |
| 75 | input/- | entrada | fe_ve_imprenta | fe_ve_imprenta | no |
| 76 | input/- | entrada | fe_ve_proveedor | fe_ve_proveedor | no |
| 77 | input/hidden | entrada | empresa_id | empresa_id | no |
| 78 | input/- | entrada | proveedor | proveedor | no |
| 79 | select/- | entrada | ambiente | Sandbox / pruebas Producción | no |
| 80 | input/- | entrada | moneda_codigo | moneda_codigo | no |
| 81 | input/- | entrada | tipo_documento_emisor | tipo_documento_emisor | no |
| 82 | input/- | entrada | identificador_fiscal | identificador_fiscal | no |
| 83 | input/- | entrada | razon_social | razon_social | no |
| 84 | input/email | entrada | email_facturacion | email_facturacion | no |
| 85 | input/checkbox | accion | enviar_factura_email_cliente_auto | enviar_factura_email_cliente_auto | no |
| 86 | input/- | entrada | telefono_facturacion | telefono_facturacion | no |
| 87 | input/- | entrada | direccion_fiscal | direccion_fiscal | no |
| 88 | input/- | entrada | prefijo_factura | prefijo_factura | no |
| 89 | input/- | entrada | resolucion_numero | resolucion_numero | no |
| 90 | input/- | entrada | api_base_url | api_base_url | no |
| 91 | textarea/- | entrada | campos_pais_json | campos_pais_json | no |
| 92 | textarea/- | entrada | observaciones | observaciones | no |
| 93 | button/submit | accion | - | Guardar configuración país | no |
| 94 | button/button | accion | reloadBtn | Recargar país | no |
| 95 | input/hidden | entrada | adv_empresa_id | adv_empresa_id | no |
| 96 | input/checkbox | accion | adv_enviar_factura_electronica_venta | adv_enviar_factura_electronica_venta | no |
| 97 | input/checkbox | accion | adv_facturacion_electronica_activa | adv_facturacion_electronica_activa | no |
| 98 | input/checkbox | accion | adv_enviar_email_venta | adv_enviar_email_venta | no |
| 99 | select/- | entrada | adv_tipo_documento_emisor | NIT CC CE PAS OTRO | no |
| 100 | input/- | entrada | adv_nit | adv_nit | no |
| 101 | input/- | entrada | adv_digito_verificacion | adv_digito_verificacion | no |
| 102 | input/- | entrada | adv_razon_social | adv_razon_social | no |
| 103 | input/- | entrada | adv_nombre_comercial | adv_nombre_comercial | no |
| 104 | input/- | entrada | adv_regimen_fiscal | adv_regimen_fiscal | no |
| 105 | input/- | entrada | adv_responsabilidad_tributaria | adv_responsabilidad_tributaria | no |
| 106 | input/email | entrada | adv_email_facturacion | adv_email_facturacion | no |
| 107 | input/- | entrada | adv_telefono_facturacion | adv_telefono_facturacion | no |
| 108 | input/- | entrada | adv_direccion_fiscal | adv_direccion_fiscal | no |
| 109 | input/- | entrada | adv_departamento | adv_departamento | no |
| 110 | input/- | entrada | adv_municipio | adv_municipio | no |
| 111 | input/- | entrada | adv_pais_codigo | CO | no |
| 112 | input/- | entrada | adv_codigo_postal | adv_codigo_postal | no |
| 113 | select/- | entrada | adv_ambiente_fe | Habilitación Producción | no |
| 114 | input/- | entrada | adv_tipo_operacion | 10 | no |
| 115 | input/- | entrada | adv_prefijo_factura | adv_prefijo_factura | no |
| 116 | input/- | entrada | adv_resolucion_numero | adv_resolucion_numero | no |
| 117 | input/date | entrada | adv_resolucion_fecha_desde | adv_resolucion_fecha_desde | no |
| 118 | input/date | entrada | adv_resolucion_fecha_hasta | adv_resolucion_fecha_hasta | no |
| 119 | input/number | entrada | adv_consecutivo_desde | 1 | no |
| 120 | input/number | entrada | adv_consecutivo_hasta | 999999 | no |
| 121 | input/number | entrada | adv_proximo_consecutivo | 1 | no |
| 122 | select/- | entrada | adv_formato_impresion | Tamaño grande / carta Impresora POS (tirilla) | no |
| 123 | input/checkbox | accion | adv_imprimir_copia_factura | adv_imprimir_copia_factura | no |
| 124 | input/checkbox | accion | adv_total_en_letras | adv_total_en_letras | no |
| 125 | input/- | entrada | adv_logo_url | adv_logo_url | no |
| 126 | input/checkbox | accion | adv_mostrar_logo | adv_mostrar_logo | no |
| 127 | textarea/- | entrada | adv_pie_factura | adv_pie_factura | no |
| 128 | textarea/- | entrada | adv_notas_legales | adv_notas_legales | no |
| 129 | input/color | entrada | adv_color_carrito_activo | #d9fbe8 | no |
| 130 | input/color | entrada | adv_color_carrito_inactivo | #fff9ef | no |
| 131 | textarea/- | entrada | adv_observaciones | adv_observaciones | no |
| 132 | button/submit | accion | - | Guardar configuración avanzada | no |
| 133 | button/button | accion | advReloadBtn | Recargar configuración avanzada | no |
| 134 | button/- | accion | - | Cargar | sí |

### `web/administrar_empresa/facturacion_electronica_ecuador.html` (27)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Abrir SRI Facturacion Electronica | no |
| 2 | a/- | accion | - | Abrir comprobantes electronicos | no |
| 3 | input/hidden | entrada | empresa_id | empresa_id | no |
| 4 | input/- | entrada | identificador_fiscal | identificador_fiscal | no |
| 5 | input/- | entrada | razon_social | razon_social | no |
| 6 | input/email | entrada | email_facturacion | email_facturacion | no |
| 7 | select/- | entrada | ambiente | Pruebas Produccion | no |
| 8 | select/- | entrada | ambiente_sri | 1 - Pruebas 2 - Produccion | no |
| 9 | select/- | entrada | moneda_codigo | USD | no |
| 10 | input/- | entrada | proveedor | proveedor | no |
| 11 | select/- | entrada | integracion | Sistema propio / XML firmado Proveedor tecnologico Facturador SRI | no |
| 12 | input/- | entrada | api_base_url | api_base_url | no |
| 13 | input/- | entrada | establecimiento | establecimiento | no |
| 14 | input/- | entrada | punto_emision | punto_emision | no |
| 15 | input/- | entrada | resolucion_numero | resolucion_numero | no |
| 16 | input/- | entrada | certificado_firma_ref | certificado_firma_ref | no |
| 17 | input/- | entrada | clave_acceso | clave_acceso | no |
| 18 | input/- | entrada | numero_autorizacion | numero_autorizacion | no |
| 19 | input/checkbox | accion | certificado_firma_confirmado | certificado_firma_confirmado | no |
| 20 | input/checkbox | accion | autorizacion_produccion_sri | autorizacion_produccion_sri | no |
| 21 | input/checkbox | accion | ride | ride | no |
| 22 | input/checkbox | accion | obligado_contabilidad | obligado_contabilidad | no |
| 23 | input/checkbox | accion | enviar_factura_email_cliente_auto | enviar_factura_email_cliente_auto | no |
| 24 | textarea/- | entrada | observaciones | observaciones | no |
| 25 | button/submit | accion | - | Guardar Ecuador | no |
| 26 | button/button | accion | validarBtn | Validar checklist | no |
| 27 | button/button | accion | recargarBtn | Recargar | no |

### `web/administrar_empresa/facturacion_electronica_menu.html` (2)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | select/- | entrada | facturacionPaisSelector | facturacionPaisSelector | no |
| 2 | button/button | accion | - | &#9776; Ocultar menú | sí |

### `web/administrar_empresa/facturacion_electronica_panama.html` (27)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Abrir DGI Factura Electronica | no |
| 2 | a/- | accion | - | Abrir e-Tax2.0 | no |
| 3 | input/hidden | entrada | empresa_id | empresa_id | no |
| 4 | input/- | entrada | identificador_fiscal | identificador_fiscal | no |
| 5 | input/- | entrada | dv | dv | no |
| 6 | input/- | entrada | razon_social | razon_social | no |
| 7 | select/- | entrada | modalidad | Proveedor Autorizado Calificado (PAC) Facturador Gratuito DGI Pendiente de definir | no |
| 8 | select/- | entrada | ambiente | Pruebas Produccion | no |
| 9 | select/- | entrada | moneda_codigo | PAB USD | no |
| 10 | input/- | entrada | proveedor | proveedor | no |
| 11 | input/- | entrada | pac_nombre | pac_nombre | no |
| 12 | input/- | entrada | api_base_url | api_base_url | no |
| 13 | input/- | entrada | sucursal | sucursal | no |
| 14 | input/- | entrada | punto_expedicion | punto_expedicion | no |
| 15 | input/- | entrada | resolucion_numero | resolucion_numero | no |
| 16 | input/- | entrada | certificado_firma_ref | certificado_firma_ref | no |
| 17 | input/- | entrada | cafe | cafe | no |
| 18 | input/- | entrada | cufe | cufe | no |
| 19 | input/- | entrada | qr_url | qr_url | no |
| 20 | input/checkbox | accion | registro_sfep | registro_sfep | no |
| 21 | input/checkbox | accion | declaracion_jurada_sfep | declaracion_jurada_sfep | no |
| 22 | input/checkbox | accion | certificado_firma_confirmado | certificado_firma_confirmado | no |
| 23 | input/checkbox | accion | enviar_factura_email_cliente_auto | enviar_factura_email_cliente_auto | no |
| 24 | textarea/- | entrada | observaciones | observaciones | no |
| 25 | button/submit | accion | - | Guardar Panama | no |
| 26 | button/button | accion | validarBtn | Validar checklist | no |
| 27 | button/button | accion | recargarBtn | Recargar | no |

### `web/administrar_empresa/facturacion_electronica_pruebas_dian.html` (54)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnDianRecargarConfig | Recargar configuracion | no |
| 2 | button/button | accion | btnDianValidarCredenciales | Validar credenciales | no |
| 3 | button/button | accion | btnDianConsultarRango | Consultar clave tecnica DIAN | no |
| 4 | button/button | accion | btnDianDiagnostico | Diagnóstico DIAN | no |
| 5 | input/- | entrada | feDianModoDescripcion | feDianModoDescripcion | no |
| 6 | input/datetime-local | entrada | feDianModoFechaInicio | feDianModoFechaInicio | no |
| 7 | input/datetime-local | entrada | feDianModoFechaTermino | feDianModoFechaTermino | no |
| 8 | input/- | entrada | feDianTestSetId | feDianTestSetId | no |
| 9 | select/- | entrada | feDianSetPreset | Portal DIAN software propio/proveedor: 30 + 10 + 10 Historico si DIAN lo exige: 60 + 20 + 20 Personalizado | no |
| 10 | input/number | entrada | feDianSetFacturas | 30 | no |
| 11 | input/number | entrada | feDianSetNotasDebito | 10 | no |
| 12 | input/number | entrada | feDianSetNotasCredito | 10 | no |
| 13 | input/number | entrada | feDianSetTotalRequerido | 50 | no |
| 14 | input/number | entrada | feDianSetAceptadosTotal | 1 | no |
| 15 | input/number | entrada | feDianSetAceptadasFacturas | 1 | no |
| 16 | input/number | entrada | feDianSetAceptadasND | 0 | no |
| 17 | input/number | entrada | feDianSetAceptadasNC | 0 | no |
| 18 | input/number | entrada | feDianSetMaxEnvios | feDianSetMaxEnvios | no |
| 19 | input/- | entrada | feDianProduccionURL | feDianProduccionURL | no |
| 20 | input/checkbox | accion | feDianConfirmarHabilitado | feDianConfirmarHabilitado | no |
| 21 | button/button | accion | btnDianGuardarObjetivo | Guardar objetivo DIAN | no |
| 22 | button/button | accion | btnDianPruebas | Ejecutar set automatico | no |
| 23 | button/button | accion | btnDianEnviarFactura | Enviar factura | no |
| 24 | button/button | accion | btnDianEnviarND | Enviar nota debito | no |
| 25 | button/button | accion | btnDianEnviarNC | Enviar nota credito | no |
| 26 | button/button | accion | btnDianTrackRecargar | Recargar historial | no |
| 27 | input/checkbox | accion | feDianConsoleAutoScroll | feDianConsoleAutoScroll | no |
| 28 | button/button | accion | btnDianConsoleCopiar | Copiar consola | no |
| 29 | button/button | accion | btnDianConsoleLimpiar | Limpiar consola | no |
| 30 | button/button | accion | btnDianCheckConexion | Probar conexion | no |
| 31 | button/button | accion | btnDianProcesarCola | Procesar cola | no |
| 32 | input/- | entrada | op_documento_codigo | op_documento_codigo | no |
| 33 | select/- | entrada | op_tipo_documento | Factura electronica de venta Nota credito electronica Nota debito electronica Documento soporte electronico Nomina elect | no |
| 34 | select/- | entrada | op_estado_actual | borrador emitida anulada ajustada | no |
| 35 | input/- | entrada | op_periodo_contable | op_periodo_contable | no |
| 36 | input/number | entrada | op_monto_total | op_monto_total | no |
| 37 | input/- | entrada | op_moneda | COP | no |
| 38 | textarea/- | entrada | op_observaciones | op_observaciones | no |
| 39 | input/number | entrada | op_cliente_id | op_cliente_id | no |
| 40 | input/email | entrada | op_cliente_email | op_cliente_email | no |
| 41 | input/- | entrada | op_cliente_nombre | op_cliente_nombre | no |
| 42 | button/button | accion | btnEmitirDocumento | Emitir documento | no |
| 43 | button/button | accion | btnAnularDocumento | Anular factura con nota credito | no |
| 44 | button/button | accion | btnEmitirFactura | Emitir factura | no |
| 45 | button/button | accion | btnEmitirNC | Emitir nota credito | no |
| 46 | button/button | accion | btnEmitirND | Emitir nota debito | no |
| 47 | button/button | accion | btnEmitirSoporte | Emitir soporte | no |
| 48 | button/button | accion | btnEmitirNominaElectronica | Emitir nomina | no |
| 49 | button/button | accion | btnEmitirPOSElectronico | Emitir POS electronico | no |
| 50 | button/button | accion | btnEmitirRadian | Registrar evento RADIAN | no |
| 51 | button/button | accion | feDianConfirmCancel | Cancelar | no |
| 52 | button/button | accion | feDianConfirmAccept | Enviar a DIAN | no |
| 53 | input/checkbox | accion | feDianPortalMostrarSensibles | feDianPortalMostrarSensibles | no |
| 54 | button/button | accion | - | Reconsultar | sí |

### `web/administrar_empresa/facturacion_electronica_tutorial_dian.html` (22)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Configuracion | sí |
| 2 | a/- | accion | - | Centro de habilitación DIAN | sí |
| 3 | a/- | accion | - | PDF DIAN: registro y modo de operacion | no |
| 4 | a/- | accion | - | DIAN: instructivo de registro y habilitacion | no |
| 5 | a/- | accion | - | PDF DIAN: requisitos y habilitacion | no |
| 6 | a/- | accion | - | Portal DIAN de habilitacion | no |
| 7 | a/- | accion | - | DIAN: proceso de registro y habilitacion | no |
| 8 | a/- | accion | - | DIAN: numeracion y autorizacion | no |
| 9 | a/- | accion | - | Descargar PDF | no |
| 10 | a/- | accion | - | Ver página DIAN | no |
| 11 | a/- | accion | - | Abrir proceso | no |
| 12 | a/- | accion | - | Portal habilitación | no |
| 13 | a/- | accion | - | Ver requerimientos | no |
| 14 | a/- | accion | - | PDF relacionado | no |
| 15 | a/- | accion | - | Descargar PDF | no |
| 16 | a/- | accion | - | Ver guía | no |
| 17 | a/- | accion | - | Abrir PDF oficial | no |
| 18 | a/- | accion | - | Ver tutorial en video | no |
| 19 | a/- | accion | - | Ver micrositio DIAN | no |
| 20 | a/- | accion | - | Asociar numeracion | no |
| 21 | a/- | accion | - | PDF habilitación | no |
| 22 | a/- | accion | - | Ruta oficial DIAN | no |

### `web/administrar_empresa/facturas_electronicas.html` (25)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | select/- | entrada | tipoDocumento | Todos Ventas Facturas electrónicas Notas crédito | no |
| 2 | select/- | entrada | estadoDocumento | Todos Borrador Emitida Anulada Ajustada | no |
| 3 | input/- | entrada | clienteFiltro | clienteFiltro | no |
| 4 | input/- | entrada | documentoFiltro | documentoFiltro | no |
| 5 | input/date | entrada | fechaDesde | fechaDesde | no |
| 6 | input/date | entrada | fechaHasta | fechaHasta | no |
| 7 | input/- | entrada | busquedaGlobal | busquedaGlobal | no |
| 8 | select/- | entrada | cajeroFiltro | Todos los usuarios | no |
| 9 | input/checkbox | accion | includeInactive | includeInactive | no |
| 10 | select/- | entrada | modoVisualizacionFactura | Carta completa Compacta ejecutiva Ticket POS 80mm | no |
| 11 | button/submit | accion | btnBuscar | Buscar documentos | no |
| 12 | button/button | accion | btnLimpiar | Limpiar filtros | no |
| 13 | button/button | accion | btnExportCsv | CSV | no |
| 14 | button/button | accion | btnExportExcel | Excel | no |
| 15 | input/- | entrada | feCancelConfirmation | feCancelConfirmation | no |
| 16 | textarea/- | entrada | feCancelReason | feCancelReason | no |
| 17 | button/button | accion | feCancelDismiss | Cancelar | no |
| 18 | button/submit | accion | feCancelSubmit | Confirmar anulación | no |
| 19 | button/button | accion | - | Imprimir ahora | sí |
| 20 | button/button | accion | - | Cerrar | sí |
| 21 | button/button | accion | - | Anular | sí |
| 22 | button/button | accion | - | Reenviar DIAN | sí |
| 23 | button/button | accion | - | Visualizar | sí |
| 24 | button/button | accion | - | Correo | sí |
| 25 | button/button | accion | - | WhatsApp | sí |

### `web/administrar_empresa/finanzas.html` (168)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnGuardarConfig | Guardar configuración | no |
| 2 | input/checkbox | accion | habilitarIngresos | habilitarIngresos | no |
| 3 | input/checkbox | accion | habilitarEgresos | habilitarEgresos | no |
| 4 | input/checkbox | accion | requiereAprobacion | requiereAprobacion | no |
| 5 | input/- | entrada | moneda | COP | no |
| 6 | input/- | entrada | prefijoIngreso | ING | no |
| 7 | input/- | entrada | prefijoEgreso | EGR | no |
| 8 | select/- | entrada | formatoImpresion | Carta POS | no |
| 9 | textarea/- | entrada | categoriasIngreso | categoriasIngreso | no |
| 10 | textarea/- | entrada | categoriasEgreso | categoriasEgreso | no |
| 11 | select/- | entrada | integracionDestino | Genérico SIIGO World Office Alegra Helisa Loggro ContaPyme | no |
| 12 | input/- | entrada | cuentaCajaBancos | 110505 | no |
| 13 | input/- | entrada | cuentaIngresos | 413595 | no |
| 14 | input/- | entrada | cuentaIVAGenerado | 240805 | no |
| 15 | input/- | entrada | cuentaGastos | 519595 | no |
| 16 | input/- | entrada | cuentaIVADescontable | 240810 | no |
| 17 | input/- | entrada | cuentaRetencionesCobrar | 135595 | no |
| 18 | input/- | entrada | cuentaRetencionesPagar | 236595 | no |
| 19 | textarea/- | entrada | cuentasIngresoMap | cuentasIngresoMap | no |
| 20 | textarea/- | entrada | cuentasEgresoMap | cuentasEgresoMap | no |
| 21 | textarea/- | entrada | configObservaciones | configObservaciones | no |
| 22 | button/button | accion | btnNuevo | Nuevo | no |
| 23 | input/hidden | entrada | movId | movId | no |
| 24 | select/- | entrada | tipoMovimiento | Ingreso Egreso | no |
| 25 | input/datetime-local | entrada | fechaMovimiento | fechaMovimiento | no |
| 26 | select/- | entrada | categoriaMovimiento | categoriaMovimiento | no |
| 27 | select/- | entrada | metodoPago | Efectivo Transferencia bancaria Tarjeta Mixto Otro | no |
| 28 | input/- | entrada | concepto | concepto | no |
| 29 | input/- | entrada | subcategoria | subcategoria | no |
| 30 | input/- | entrada | terceroNombre | terceroNombre | no |
| 31 | input/- | entrada | terceroDocumento | terceroDocumento | no |
| 32 | input/number | entrada | monto | monto | no |
| 33 | input/number | entrada | impuesto | 0 | no |
| 34 | input/number | entrada | total | total | no |
| 35 | input/- | entrada | aprobadoPor | aprobadoPor | no |
| 36 | input/number | entrada | retencionFuente | 0 | no |
| 37 | input/number | entrada | retencionICA | 0 | no |
| 38 | input/number | entrada | retencionIVA | 0 | no |
| 39 | input/number | entrada | totalRetenciones | 0 | no |
| 40 | input/number | entrada | totalNeto | 0 | no |
| 41 | select/- | entrada | tipoComprobante | Recibo interno Factura Nota contable Soporte externo | no |
| 42 | input/- | entrada | numeroComprobante | numeroComprobante | no |
| 43 | input/- | entrada | comprobanteURL | comprobanteURL | no |
| 44 | input/- | entrada | referenciaExterna | referenciaExterna | no |
| 45 | input/file | accion | comprobanteArchivo | comprobanteArchivo | no |
| 46 | textarea/- | entrada | descripcion | descripcion | no |
| 47 | textarea/- | entrada | observaciones | observaciones | no |
| 48 | button/submit | accion | btnGuardarMovimiento | Guardar movimiento | no |
| 49 | button/button | accion | btnCancelarMovimiento | Cancelar | no |
| 50 | button/button | accion | btnBuscarCierresCaja | Buscar cierres | no |
| 51 | input/hidden | entrada | cierreCajaId | cierreCajaId | no |
| 52 | input/hidden | entrada | cierreCajaEstado | abierto | no |
| 53 | input/hidden | entrada | cierreCajaRegistroEstado | activo | no |
| 54 | input/number | entrada | sucursalCierre | 0 | no |
| 55 | input/- | entrada | cajaCodigoCierre | CAJA_PRINCIPAL | no |
| 56 | input/- | entrada | turnoCierre | general | no |
| 57 | input/date | entrada | fechaOperacionCierre | fechaOperacionCierre | no |
| 58 | input/- | entrada | monedaCierre | COP | no |
| 59 | input/number | entrada | aperturaMontoCierre | 0 | no |
| 60 | input/number | entrada | ingresosEfectivoCierre | 0 | no |
| 61 | input/number | entrada | egresosEfectivoCierre | 0 | no |
| 62 | input/number | entrada | retirosEfectivoCierre | 0 | no |
| 63 | input/number | entrada | umbralIncidenciaCierre | 0 | no |
| 64 | input/number | entrada | cajaTeoricaCierre | 0 | no |
| 65 | input/number | entrada | cajaFisicaCierre | 0 | no |
| 66 | input/number | entrada | diferenciaCierre | 0 | no |
| 67 | input/- | entrada | estadoCierreVisual | abierto | no |
| 68 | input/- | entrada | estadoRegistroCierreVisual | activo | no |
| 69 | textarea/- | entrada | observacionesCierre | observacionesCierre | no |
| 70 | button/submit | accion | btnGuardarCierreCaja | Registrar apertura de caja | no |
| 71 | button/button | accion | btnNuevoCierreCaja | Nuevo cierre | no |
| 72 | input/number | entrada | filtroCierreSucursal | 0 | no |
| 73 | input/- | entrada | filtroCierreCaja | filtroCierreCaja | no |
| 74 | select/- | entrada | filtroCierreEstado | Todos Abierto Cerrado Aprobado Anulado | no |
| 75 | input/date | entrada | filtroCierreDesde | filtroCierreDesde | no |
| 76 | input/date | entrada | filtroCierreHasta | filtroCierreHasta | no |
| 77 | input/checkbox | accion | filtroCierreInactivos | filtroCierreInactivos | no |
| 78 | button/button | accion | btnAplicarPlanMotel | Aplicar plantilla Motel | no |
| 79 | button/button | accion | btnBuscarPlanCuentas | Buscar cuentas | no |
| 80 | input/hidden | entrada | planCuentaId | planCuentaId | no |
| 81 | input/- | entrada | planCodigo | planCodigo | no |
| 82 | input/- | entrada | planNombre | planNombre | no |
| 83 | select/- | entrada | planTipo | Activo Pasivo Patrimonio Ingreso Gasto Costo | no |
| 84 | select/- | entrada | planNaturaleza | Débito Crédito | no |
| 85 | input/number | entrada | planNivel | 1 | no |
| 86 | input/- | entrada | planPadre | planPadre | no |
| 87 | input/checkbox | accion | planAdmiteMovimiento | planAdmiteMovimiento | no |
| 88 | input/checkbox | accion | planAplicaImpuesto | planAplicaImpuesto | no |
| 89 | input/- | entrada | planClave | planClave | no |
| 90 | input/- | entrada | planObs | planObs | no |
| 91 | button/submit | accion | btnGuardarPlanCuenta | Guardar cuenta | no |
| 92 | button/button | accion | btnNuevoPlanCuenta | Nueva cuenta | no |
| 93 | input/- | entrada | filtroPlanQ | filtroPlanQ | no |
| 94 | input/checkbox | accion | filtroPlanInactivos | filtroPlanInactivos | no |
| 95 | input/file | accion | carteraSoporteIAArchivo | carteraSoporteIAArchivo | no |
| 96 | button/button | accion | btnCargarCxPIA | Cargar factura o recibo con IA | no |
| 97 | button/button | accion | btnBuscarCartera | Buscar cartera | no |
| 98 | button/button | accion | btnConciliarCartera | Conciliar pagos | no |
| 99 | button/button | accion | btnRevisarFuentesCxP | Comparar fuente histórica CxP | no |
| 100 | input/hidden | entrada | carteraId | carteraId | no |
| 101 | input/hidden | entrada | carteraSoporteIAId | carteraSoporteIAId | no |
| 102 | select/- | entrada | carteraTipo | Cuenta por cobrar Cuenta por pagar | no |
| 103 | input/- | entrada | carteraTercero | carteraTercero | no |
| 104 | select/- | entrada | carteraProveedorId | Selecciona un proveedor registrado | no |
| 105 | input/- | entrada | carteraDocumento | carteraDocumento | no |
| 106 | input/number | entrada | carteraValor | carteraValor | no |
| 107 | input/date | entrada | carteraFechaEmision | carteraFechaEmision | no |
| 108 | input/date | entrada | carteraFechaVencimiento | carteraFechaVencimiento | no |
| 109 | input/- | entrada | carteraObs | carteraObs | no |
| 110 | button/submit | accion | btnGuardarCartera | Guardar cartera | no |
| 111 | select/- | entrada | filtroCarteraTipo | CxC CxP | no |
| 112 | input/- | entrada | filtroCarteraQ | filtroCarteraQ | no |
| 113 | input/number | entrada | carteraAbonoMonto | carteraAbonoMonto | no |
| 114 | select/- | entrada | carteraAbonoMetodo | Efectivo Transferencia bancaria Tarjeta Otro | no |
| 115 | button/button | accion | btnPrevisualizarExtracto | Previsualizar extracto | no |
| 116 | button/button | accion | btnImportarExtracto | Importar y conciliar | no |
| 117 | button/button | accion | btnBuscarExtractos | Buscar extractos | no |
| 118 | input/- | entrada | extractoBanco | extractoBanco | no |
| 119 | input/- | entrada | extractoCuenta | extractoCuenta | no |
| 120 | input/- | entrada | extractoPeriodo | extractoPeriodo | no |
| 121 | select/- | entrada | extractoEstado | Todos Pendiente Conciliado Con desviación | no |
| 122 | textarea/- | entrada | extractoCSV | extractoCSV | no |
| 123 | button/button | accion | btnBuscar | Buscar | no |
| 124 | button/button | accion | btnCargarDemo | Cargar datos demo | no |
| 125 | button/button | accion | btnProcesarAsientos | Procesar eventos contables | no |
| 126 | button/button | accion | btnExportarExcel | Exportar Excel | no |
| 127 | button/button | accion | btnExportarPDF | Exportar PDF | no |
| 128 | button/button | accion | btnExportarJSON | Exportar JSON contable | no |
| 129 | button/button | accion | btnExportarSIIGO | Plantilla SIIGO CSV | no |
| 130 | button/button | accion | btnExportarBalancePrueba | Balance de prueba CSV | no |
| 131 | button/button | accion | btnExportarEstadoResultados | Estado resultados CSV | no |
| 132 | button/button | accion | btnExportarBalanceGeneral | Balance general CSV | no |
| 133 | button/button | accion | btnExportarLibroDiario | Libro diario CSV | no |
| 134 | button/button | accion | btnExportarLibroMayor | Libro mayor CSV | no |
| 135 | input/date | entrada | filtroDesde | filtroDesde | no |
| 136 | input/date | entrada | filtroHasta | filtroHasta | no |
| 137 | input/- | entrada | filtroQ | filtroQ | no |
| 138 | input/- | entrada | filtroPeriodo | filtroPeriodo | no |
| 139 | button/button | accion | btnCerrarPeriodo | Cerrar periodo | no |
| 140 | button/button | accion | btnReabrirPeriodo | Reabrir periodo | no |
| 141 | button/button | accion | btnRefrescarPeriodos | Actualizar periodos | no |
| 142 | button/button | accion | tabTodos | Pestaña Todos | no |
| 143 | button/button | accion | tabIngresos | Pestaña Ingresos | no |
| 144 | button/button | accion | tabEgresos | Pestaña Egresos | no |
| 145 | button/button | accion | btnBuscarConciliacion | Actualizar conciliación | no |
| 146 | input/date | entrada | concDesde | concDesde | no |
| 147 | input/date | entrada | concHasta | concHasta | no |
| 148 | input/- | entrada | concPeriodo | concPeriodo | no |
| 149 | input/number | entrada | concLimit | 24 | no |
| 150 | a/- | accion | - | Ver comprobante actual | no |
| 151 | button/- | accion | ' + item.id + ' | Editar | sí |
| 152 | button/- | accion | ' + item.id + ' | Cerrar | sí |
| 153 | button/- | accion | ' + item.id + ' | Anular | sí |
| 154 | button/- | accion | ' + item.id + ' | Aprobar | sí |
| 155 | button/- | accion | ' + item.id + ' | Reabrir | sí |
| 156 | button/- | accion | ' + item.id + ' | Anular | sí |
| 157 | button/- | accion | ' + item.id + ' | Reabrir | sí |
| 158 | button/- | accion | ' + item.id + ' | ' + (estadoRegistro === 'activo' ? 'Desactivar' : 'Activar') + ' | sí |
| 159 | button/- | accion | ' + item.id + ' | Eliminar | sí |
| 160 | a/- | accion | - | Ver adjunto | no |
| 161 | button/- | accion | ' + item.id + ' | Editar | sí |
| 162 | button/- | accion | ' + item.id + ' | ' + (estado === 'activo' ? 'Desactivar' : 'Activar') + ' | sí |
| 163 | button/- | accion | ' + item.id + ' | Editar | sí |
| 164 | button/- | accion | ' + item.id + ' | Abonar/Pagar | sí |
| 165 | button/- | accion | ' + item.id + ' | Editar | sí |
| 166 | button/- | accion | ' + item.id + ' | Imprimir | sí |
| 167 | button/- | accion | ' + item.id + ' | ' + (estado === 'activo' ? 'Desactivar' : 'Activar') + ' | sí |
| 168 | button/- | accion | ' + item.id + ' | Eliminar | sí |

### `web/administrar_empresa/finanzas_breb_qr.html` (39)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | brebTutorialLink | Tutorial | no |
| 2 | button/button | accion | brebReloadBtn | Actualizar | no |
| 3 | button/button | accion | brebSaveConfigBtn | Guardar configuracion | no |
| 4 | input/checkbox | accion | cfgPagoQR | cfgPagoQR | no |
| 5 | input/checkbox | accion | cfgBrebEnabled | cfgBrebEnabled | no |
| 6 | input/checkbox | accion | cfgMixed | cfgMixed | no |
| 7 | input/checkbox | accion | cfgRequireProof | cfgRequireProof | no |
| 8 | input/checkbox | accion | cfgAutoReconcile | cfgAutoReconcile | no |
| 9 | input/checkbox | accion | cfgCuentaCaja | cfgCuentaCaja | no |
| 10 | input/- | entrada | cfgPrefix | cfgPrefix | no |
| 11 | input/number | entrada | cfgAlertMin | cfgAlertMin | no |
| 12 | input/- | entrada | cfgWebhook | cfgWebhook | no |
| 13 | textarea/- | entrada | cfgInstructions | cfgInstructions | no |
| 14 | button/button | accion | brebAddAccountBtn | Agregar cuenta | no |
| 15 | button/button | accion | brebManualBtn | Registrar pago | no |
| 16 | input/number | entrada | manualMonto | manualMonto | no |
| 17 | input/- | entrada | manualRef | manualRef | no |
| 18 | input/datetime-local | entrada | manualFecha | manualFecha | no |
| 19 | input/- | entrada | manualCuenta | manualCuenta | no |
| 20 | input/- | entrada | manualBanco | manualBanco | no |
| 21 | input/- | entrada | manualCaja | manualCaja | no |
| 22 | input/number | entrada | manualCarrito | manualCarrito | no |
| 23 | input/- | entrada | manualPagador | manualPagador | no |
| 24 | select/- | entrada | manualEstado | Pendiente Conciliado | no |
| 25 | textarea/- | entrada | manualObs | manualObs | no |
| 26 | input/checkbox | accion | - | sin etiqueta | sí |
| 27 | input/- | entrada | - | ${esc(row.nombre)} | sí |
| 28 | select/- | entrada | - | Bre-B Nequi Otro | sí |
| 29 | input/- | entrada | - | ${esc(row.tipo_llave)} | sí |
| 30 | input/- | entrada | - | ${esc(row.llave)} | sí |
| 31 | input/- | entrada | - | ${esc(row.caja_codigo)} | sí |
| 32 | select/- | entrada | - | Dinamico Estatico | sí |
| 33 | input/- | entrada | - | ${esc(row.payload_oficial)} | sí |
| 34 | button/button | accion | - | Quitar | no |
| 35 | input/hidden | entrada | - | ${esc(row.comercio)} | sí |
| 36 | input/hidden | entrada | - | ${esc(row.instrucciones)} | sí |
| 37 | input/hidden | entrada | - | ${esc(row.cuenta_contable)} | sí |
| 38 | input/hidden | entrada | - | ${esc(row.banco_receptor)} | sí |
| 39 | input/hidden | entrada | - | ${esc(row.referencia_fija)} | sí |

### `web/administrar_empresa/finanzas_breb_qr_tutorial.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | backLink | Volver a Bre-B QR | no |

### `web/administrar_empresa/finanzas_menu.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | - | &#9776; Ocultar menu | sí |

### `web/administrar_empresa/frecuencia_fe.html` (5)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/checkbox | accion | freqEnabled | freqEnabled | no |
| 2 | input/number | entrada | freqCadaNNo | 0 | no |
| 3 | input/number | entrada | freqContador | 0 | no |
| 4 | button/submit | accion | - | Guardar frecuencia | no |
| 5 | button/button | accion | btnReload | Recargar | no |

### `web/administrar_empresa/generador_codigos_barras.html` (6)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnLoad | Recargar productos | no |
| 2 | button/button | accion | btnPrint | Imprimir etiquetas | no |
| 3 | input/- | entrada | q | q | no |
| 4 | input/- | entrada | prefix | PCS | no |
| 5 | button/button | accion | btnGenerateMissing | Generar faltantes | no |
| 6 | button/button | accion | '+Number(p.id)+' | '+saved+' | sí |

### `web/administrar_empresa/gestion_documental_menu.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | - | &#9776; Ocultar menu | sí |

### `web/administrar_empresa/historial_productos.html` (5)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnRefresh | Actualizar | no |
| 2 | button/button | accion | - | CSV | sí |
| 3 | button/button | accion | - | JSON | sí |
| 4 | button/button | accion | - | HTML | sí |
| 5 | button/button | accion | - | PDF | sí |

### `web/administrar_empresa/hoja_vida_operativa.html` (42)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | newEntityBtn | Nueva hoja de vida | no |
| 2 | input/- | entrada | filterQ | filterQ | no |
| 3 | select/- | entrada | filterTipo | Todos Moto Vehiculo Paciente Equipo Mascota Maquinaria Activo Otro | no |
| 4 | button/button | accion | reloadBtn | Recargar | no |
| 5 | input/hidden | entrada | entityId | entityId | no |
| 6 | select/- | entrada | tipo_entidad | Moto Vehiculo Paciente Equipo Mascota Maquinaria Activo Otro | no |
| 7 | input/- | entrada | codigo | codigo | no |
| 8 | select/- | entrada | estado_operativo | Activo En servicio En observacion Pendiente Cerrado | no |
| 9 | input/- | entrada | nombre | nombre | no |
| 10 | input/- | entrada | cliente_nombre | cliente_nombre | no |
| 11 | input/- | entrada | identificacion | identificacion | no |
| 12 | input/- | entrada | marca | marca | no |
| 13 | input/- | entrada | modelo | modelo | no |
| 14 | input/- | entrada | serie | serie | no |
| 15 | input/- | entrada | color | color | no |
| 16 | input/date | entrada | fecha_ingreso | fecha_ingreso | no |
| 17 | textarea/- | entrada | observaciones | observaciones | no |
| 18 | button/submit | accion | - | Guardar | no |
| 19 | button/button | accion | cancelEntityBtn | Cancelar | no |
| 20 | select/- | entrada | tipo_evento | Servicio Mantenimiento Consulta Diagnostico Reparacion Alerta Reporte Evento | no |
| 21 | input/- | entrada | evento_titulo | evento_titulo | no |
| 22 | input/datetime-local | entrada | fecha_evento | fecha_evento | no |
| 23 | input/datetime-local | entrada | fecha_proxima | fecha_proxima | no |
| 24 | input/number | entrada | costo | costo | no |
| 25 | input/- | entrada | responsable | responsable | no |
| 26 | input/- | entrada | documento_referencia | documento_referencia | no |
| 27 | textarea/- | entrada | evento_descripcion | evento_descripcion | no |
| 28 | input/checkbox | accion | recurrente | recurrente | no |
| 29 | input/number | entrada | recurrencia_dias | recurrencia_dias | no |
| 30 | button/submit | accion | saveEventBtn | Agregar evento | no |
| 31 | input/- | entrada | alerta_titulo | alerta_titulo | no |
| 32 | input/datetime-local | entrada | fecha_programada | fecha_programada | no |
| 33 | select/- | entrada | prioridad | Media Alta Baja | no |
| 34 | input/- | entrada | alerta_responsable | alerta_responsable | no |
| 35 | textarea/- | entrada | alerta_descripcion | alerta_descripcion | no |
| 36 | button/submit | accion | - | Agregar alerta | no |
| 37 | button/button | accion | loadAllAlertsBtn | Ver todas | no |
| 38 | button/button | accion | ' + esc(item.id) + ' | Ver | sí |
| 39 | button/button | accion | ' + esc(item.id) + ' | Editar | sí |
| 40 | button/button | accion | ' + esc(item.id) + ' | Eliminar | sí |
| 41 | button/button | accion | ' + esc(item.id) + ' | Eliminar | sí |
| 42 | button/button | accion | - | Completar | sí |

### `web/administrar_empresa/horarios_trabajadores.html` (43)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/date | entrada | filtro_desde | filtro_desde | no |
| 2 | input/date | entrada | filtro_hasta | filtro_hasta | no |
| 3 | input/- | entrada | filtro_q | filtro_q | no |
| 4 | input/- | entrada | filtro_area | filtro_area | no |
| 5 | input/- | entrada | filtro_sede | filtro_sede | no |
| 6 | select/- | entrada | filtro_estado | Todos Programado Publicado Cubierto Incidencia Cancelado | no |
| 7 | input/checkbox | accion | filtro_publicados | filtro_publicados | no |
| 8 | button/button | accion | btnRefrescar | Actualizar vista | no |
| 9 | button/button | accion | btnPublicarRango | Publicar rango visible | no |
| 10 | input/hidden | entrada | turno_id | turno_id | no |
| 11 | select/- | entrada | usuario_id | Registro manual | no |
| 12 | input/- | entrada | nombre_empleado | nombre_empleado | no |
| 13 | input/- | entrada | cargo | cargo | no |
| 14 | input/- | entrada | area | area | no |
| 15 | input/- | entrada | sede | sede | no |
| 16 | input/date | entrada | fecha_inicio | fecha_inicio | no |
| 17 | input/date | entrada | fecha_fin | fecha_fin | no |
| 18 | input/- | entrada | turno_nombre | turno_nombre | no |
| 19 | input/time | entrada | hora_inicio | hora_inicio | no |
| 20 | input/time | entrada | hora_fin | hora_fin | no |
| 21 | input/number | entrada | descanso_minutos | 30 | no |
| 22 | select/- | entrada | tipo_turno | Operativo Administrativo Cobertura Guardia | no |
| 23 | select/- | entrada | canal | Presencial Mixto Remoto | no |
| 24 | select/- | entrada | estado_turno | Programado Publicado Cubierto Incidencia | no |
| 25 | input/color | entrada | color_turno | #2563eb | no |
| 26 | input/- | entrada | observaciones | observaciones | no |
| 27 | input/checkbox | accion | - | 1 | no |
| 28 | input/checkbox | accion | - | 2 | no |
| 29 | input/checkbox | accion | - | 3 | no |
| 30 | input/checkbox | accion | - | 4 | no |
| 31 | input/checkbox | accion | - | 5 | no |
| 32 | input/checkbox | accion | - | 6 | no |
| 33 | input/checkbox | accion | - | 0 | no |
| 34 | input/checkbox | accion | publicado | publicado | no |
| 35 | input/checkbox | accion | requiere_cobertura | requiere_cobertura | no |
| 36 | button/submit | accion | btnGuardarTurno | Guardar turno | no |
| 37 | button/button | accion | btnLimpiarTurno | Limpiar formulario | no |
| 38 | input/number | entrada | cfg_horas_dia | cfg_horas_dia | no |
| 39 | input/number | entrada | cfg_horas_semana | cfg_horas_semana | no |
| 40 | input/number | entrada | cfg_descanso | cfg_descanso | no |
| 41 | input/number | entrada | cfg_anticipacion | cfg_anticipacion | no |
| 42 | input/checkbox | accion | cfg_permitir_solapados | cfg_permitir_solapados | no |
| 43 | button/button | accion | btnGuardarConfig | Guardar reglas | no |

### `web/administrar_empresa/hotel_tarjetas_acceso.html` (18)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | reloadBtn | Actualizar | no |
| 2 | input/hidden | entrada | cardId | cardId | no |
| 3 | select/- | entrada | stationSelect | stationSelect | no |
| 4 | input/number | entrada | stationId | stationId | no |
| 5 | input/- | entrada | stationCode | stationCode | no |
| 6 | input/- | entrada | stationName | stationName | no |
| 7 | input/- | entrada | cardCode | cardCode | no |
| 8 | input/- | entrada | cardUid | cardUid | no |
| 9 | input/- | entrada | guestName | guestName | no |
| 10 | input/number | entrada | reservationId | reservationId | no |
| 11 | input/number | entrada | maxUses | 0 | no |
| 12 | input/datetime-local | entrada | validFrom | validFrom | no |
| 13 | input/datetime-local | entrada | validTo | validTo | no |
| 14 | select/- | entrada | status | Activa Bloqueada | no |
| 15 | textarea/- | entrada | notes | notes | no |
| 16 | button/submit | accion | - | Guardar y programar | no |
| 17 | button/button | accion | resetBtn | Limpiar | no |
| 18 | input/checkbox | accion | includeInactive | includeInactive | no |

### `web/administrar_empresa/importaciones_costeo.html` (45)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | impRefresh | Actualizar | no |
| 2 | button/button | accion | impSeed | Cargar demo | no |
| 3 | button/button | accion | impExport | Exportar CSV | no |
| 4 | button/button | accion | - | Tablero | sí |
| 5 | button/button | accion | - | Importaciones | sí |
| 6 | button/button | accion | - | Items | sí |
| 7 | button/button | accion | - | Costos nacionalizacion | sí |
| 8 | button/button | accion | - | Costeo aterrizado | sí |
| 9 | button/button | accion | - | Nuevo embarque | sí |
| 10 | button/button | accion | - | Cargar items | sí |
| 11 | button/button | accion | - | Registrar costos | sí |
| 12 | button/button | accion | - | Ver costeo | sí |
| 13 | input/- | entrada | impCodigo | impCodigo | no |
| 14 | input/- | entrada | impProveedor | impProveedor | no |
| 15 | input/- | entrada | impPais | impPais | no |
| 16 | select/- | entrada | impIncoterm | FOB CIF EXW FCA DAP DDP | no |
| 17 | input/- | entrada | impMoneda | USD | no |
| 18 | input/number | entrada | impTRM | 3900 | no |
| 19 | input/date | entrada | impFecha | impFecha | no |
| 20 | input/date | entrada | impEta | impEta | no |
| 21 | input/- | entrada | impRef | impRef | no |
| 22 | select/- | entrada | impEstado | Borrador En transito Costeado Cerrado Contabilizado Anulado | no |
| 23 | button/submit | accion | - | Guardar importacion | no |
| 24 | button/button | accion | impClear | Limpiar | no |
| 25 | select/- | entrada | estadoFilter | Todos Borrador En transito Costeado Cerrado Contabilizado Anulado | no |
| 26 | input/- | entrada | searchFilter | searchFilter | no |
| 27 | select/- | entrada | itemImportacion | itemImportacion | no |
| 28 | input/- | entrada | itemNombre | itemNombre | no |
| 29 | input/- | entrada | itemSKU | itemSKU | no |
| 30 | input/- | entrada | itemUnidad | und | no |
| 31 | input/number | entrada | itemCantidad | itemCantidad | no |
| 32 | input/number | entrada | itemCosto | itemCosto | no |
| 33 | input/number | entrada | itemPeso | itemPeso | no |
| 34 | input/number | entrada | itemVol | itemVol | no |
| 35 | button/submit | accion | - | Agregar item | no |
| 36 | select/- | entrada | costImportacion | costImportacion | no |
| 37 | select/- | entrada | costTipo | Flete Seguro Arancel IVA Aduana Bodega Transporte interno Otro nacionalizacion | no |
| 38 | input/- | entrada | costConcepto | costConcepto | no |
| 39 | select/- | entrada | costBase | Valor Peso Volumen Cantidad | no |
| 40 | input/number | entrada | costValor | costValor | no |
| 41 | input/- | entrada | costTercero | costTercero | no |
| 42 | input/- | entrada | costDocumento | costDocumento | no |
| 43 | input/- | entrada | costCuenta | costCuenta | no |
| 44 | button/submit | accion | - | Agregar costo | no |
| 45 | button/button | accion | btnDistribuir | Distribuir costos | no |

### `web/administrar_empresa/impuestos.html` (10)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnImpuestosRefrescar | Actualizar | no |
| 2 | button/button | accion | btnImpuestosAgenteInternet | Agente internet | no |
| 3 | button/button | accion | btnImpuestosSembrar | Cargar catálogo base | no |
| 4 | input/date | entrada | impuestosDesde | impuestosDesde | no |
| 5 | input/date | entrada | impuestosHasta | impuestosHasta | no |
| 6 | button/button | accion | btnImpuestosAplicar | Aplicar período | no |
| 7 | button/button | accion | btnImpuestosMes | Últimos 30 días | no |
| 8 | button/button | accion | btnExportCodigos | Exportar CSV | no |
| 9 | button/button | accion | btnExportAsientos | Exportar CSV | no |
| 10 | button/button | accion | btnExportDiario | Exportar CSV | no |

### `web/administrar_empresa/ingresos.html` (30)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnNuevoMovimiento | Nuevo ingreso | no |
| 2 | input/hidden | entrada | movimientoId | movimientoId | no |
| 3 | input/hidden | entrada | codigoMovimiento | codigoMovimiento | no |
| 4 | input/hidden | entrada | comprobanteUrl | comprobanteUrl | no |
| 5 | input/datetime-local | entrada | fechaMovimiento | fechaMovimiento | no |
| 6 | select/- | entrada | categoriaMovimiento | Ventas Servicios Abonos Consignacion Otros ingresos | no |
| 7 | select/- | entrada | metodoPago | Efectivo Transferencia bancaria Tarjeta debito Tarjeta credito Otro | no |
| 8 | input/- | entrada | moneda | COP | no |
| 9 | input/- | entrada | concepto | concepto | no |
| 10 | input/- | entrada | subcategoria | subcategoria | no |
| 11 | input/- | entrada | terceroNombre | terceroNombre | no |
| 12 | input/- | entrada | terceroDocumento | terceroDocumento | no |
| 13 | input/number | entrada | monto | monto | no |
| 14 | input/number | entrada | impuesto | 0 | no |
| 15 | input/number | entrada | totalRetenciones | 0 | no |
| 16 | input/number | entrada | totalNeto | totalNeto | no |
| 17 | select/- | entrada | tipoComprobante | Recibo interno Factura Soporte externo Nota contable | no |
| 18 | input/- | entrada | numeroComprobante | numeroComprobante | no |
| 19 | input/file | accion | comprobanteFile | comprobanteFile | no |
| 20 | button/button | accion | btnAnalizarComprobanteIA | Analizar con IA | no |
| 21 | input/- | entrada | referenciaExterna | referenciaExterna | no |
| 22 | textarea/- | entrada | descripcion | descripcion | no |
| 23 | textarea/- | entrada | observaciones | observaciones | no |
| 24 | input/checkbox | accion | imprimirAlGuardar | imprimirAlGuardar | no |
| 25 | button/submit | accion | btnGuardarMovimiento | Guardar ingreso | no |
| 26 | button/button | accion | btnCancelarMovimiento | Limpiar | no |
| 27 | input/- | entrada | filtroQ | filtroQ | no |
| 28 | input/date | entrada | filtroDesde | Desde | no |
| 29 | input/date | entrada | filtroHasta | Hasta | no |
| 30 | button/button | accion | btnBuscarMovimientos | Buscar | no |

### `web/administrar_empresa/inventario_avanzado.html` (37)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnRefresh | Actualizar | no |
| 2 | button/button | accion | btnSeed | Cargar demo | no |
| 3 | button/button | accion | - | Lote | sí |
| 4 | button/button | accion | - | Serial | sí |
| 5 | button/button | accion | - | Reserva | sí |
| 6 | input/number | entrada | loteProducto | loteProducto | no |
| 7 | input/number | entrada | loteBodega | loteBodega | no |
| 8 | input/- | entrada | loteCodigo | loteCodigo | no |
| 9 | input/number | entrada | loteCantidad | loteCantidad | no |
| 10 | input/number | entrada | loteCosto | loteCosto | no |
| 11 | input/date | entrada | loteFabricacion | loteFabricacion | no |
| 12 | input/date | entrada | loteVence | loteVence | no |
| 13 | select/- | entrada | loteCalidad | liberado cuarentena bloqueado rechazado | no |
| 14 | input/- | entrada | loteProveedor | loteProveedor | no |
| 15 | input/- | entrada | loteDocumento | loteDocumento | no |
| 16 | input/- | entrada | loteUbicacion | loteUbicacion | no |
| 17 | button/button | accion | btnSaveLote | Guardar lote | no |
| 18 | input/number | entrada | serialLote | serialLote | no |
| 19 | input/number | entrada | serialProducto | serialProducto | no |
| 20 | input/number | entrada | serialBodega | serialBodega | no |
| 21 | input/- | entrada | serialCodigo | serialCodigo | no |
| 22 | select/- | entrada | serialEstado | disponible reservado mantenimiento bloqueado | no |
| 23 | input/date | entrada | serialIngreso | serialIngreso | no |
| 24 | input/date | entrada | serialGarantia | serialGarantia | no |
| 25 | button/button | accion | btnSaveSerial | Guardar serial | no |
| 26 | input/number | entrada | resProducto | resProducto | no |
| 27 | input/number | entrada | resBodega | resBodega | no |
| 28 | input/number | entrada | resLote | resLote | no |
| 29 | input/number | entrada | resSerial | resSerial | no |
| 30 | input/number | entrada | resCantidad | resCantidad | no |
| 31 | input/- | entrada | resModulo | venta | no |
| 32 | input/- | entrada | resRef | resRef | no |
| 33 | input/- | entrada | resCliente | resCliente | no |
| 34 | input/date | entrada | resExpira | resExpira | no |
| 35 | input/number | entrada | confirmReservaID | confirmReservaID | no |
| 36 | button/button | accion | btnSaveReserva | Crear reserva | no |
| 37 | button/button | accion | btnConfirmReserva | Confirmar salida | no |

### `web/administrar_empresa/licencia_sistema.html` (10)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | downloadPdfLink | Descargar licencia PDF | no |
| 2 | button/button | accion | printLicenseBtn | Vista imprimible | no |
| 3 | a/- | accion | buyLicenseLink | Comprar licencias adicionales | no |
| 4 | input/number | entrada | licenseAlertDays | 1 | no |
| 5 | input/checkbox | accion | licenseAlertEmail | licenseAlertEmail | no |
| 6 | input/checkbox | accion | licenseAlertWhatsApp | licenseAlertWhatsApp | no |
| 7 | input/checkbox | accion | licenseAlertBuzon | licenseAlertBuzon | no |
| 8 | textarea/- | entrada | licenseAlertMessage | Tu licencia de Powerful Control System está próxima a vencer. Renueva para mantener el servicio activo. | no |
| 9 | button/button | accion | saveLicenseAlertsBtn | Guardar recordatorios | no |
| 10 | button/button | accion | refreshPurchaseHistoryBtn | Actualizar historial | no |

### `web/administrar_empresa/logistica_wms.html` (68)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | wmsRefresh | Actualizar | no |
| 2 | button/button | accion | wmsSeed | Cargar demo | no |
| 3 | button/button | accion | wmsExport | Exportar CSV | no |
| 4 | button/button | accion | - | Tablero | sí |
| 5 | button/button | accion | - | Ubicaciones | sí |
| 6 | button/button | accion | - | Ordenes | sí |
| 7 | button/button | accion | - | Picking/Packing | sí |
| 8 | button/button | accion | - | Despachos | sí |
| 9 | button/button | accion | - | Bitacora | sí |
| 10 | button/button | accion | - | Mapa de bodega | sí |
| 11 | button/button | accion | - | Gestionar ordenes | sí |
| 12 | button/button | accion | - | Despachos | sí |
| 13 | input/- | entrada | ubiCodigo | ubiCodigo | no |
| 14 | input/- | entrada | ubiBodega | Principal | no |
| 15 | input/- | entrada | ubiZona | ubiZona | no |
| 16 | input/- | entrada | ubiPasillo | ubiPasillo | no |
| 17 | input/- | entrada | ubiRack | ubiRack | no |
| 18 | input/- | entrada | ubiNivel | ubiNivel | no |
| 19 | input/- | entrada | ubiPosicion | ubiPosicion | no |
| 20 | select/- | entrada | ubiTipo | Almacenamiento Picking Packing Recepcion Despacho Cuarentena Conteo | no |
| 21 | input/number | entrada | ubiCapacidad | 0 | no |
| 22 | input/number | entrada | ubiOcupacion | 0 | no |
| 23 | select/- | entrada | ubiEstado | Activa Inactiva | no |
| 24 | textarea/- | entrada | ubiObs | ubiObs | no |
| 25 | button/submit | accion | - | Guardar ubicacion | no |
| 26 | button/button | accion | ubiClear | Limpiar | no |
| 27 | select/- | entrada | ubicacionFilter | Todas Activas Inactivas | no |
| 28 | input/hidden | entrada | ordenId | ordenId | no |
| 29 | input/- | entrada | ordenCodigo | ordenCodigo | no |
| 30 | select/- | entrada | ordenTipo | Picking Packing Despacho Conteo Reabastecimiento Traslado Devolucion | no |
| 31 | input/- | entrada | ordenOrigen | ordenOrigen | no |
| 32 | input/- | entrada | ordenCliente | ordenCliente | no |
| 33 | input/date | entrada | ordenFecha | ordenFecha | no |
| 34 | select/- | entrada | ordenPrioridad | Normal Alta Urgente Baja | no |
| 35 | select/- | entrada | ordenEstado | Borrador Liberada En picking En packing Lista despacho Despachada Cerrada Cancelada | no |
| 36 | input/- | entrada | ordenResponsable | ordenResponsable | no |
| 37 | input/- | entrada | ordenObs | ordenObs | no |
| 38 | button/submit | accion | - | Guardar orden | no |
| 39 | button/button | accion | ordenClear | Limpiar | no |
| 40 | select/- | entrada | ordenTipoFilter | Todos Picking Packing Despacho Conteo Reabastecimiento Traslado Devolucion | no |
| 41 | select/- | entrada | ordenEstadoFilter | Todos Borrador Liberada En picking En packing Lista despacho Despachada Cerrada Cancelada | no |
| 42 | input/- | entrada | ordenSearch | ordenSearch | no |
| 43 | select/- | entrada | itemOrdenId | itemOrdenId | no |
| 44 | input/- | entrada | itemProducto | itemProducto | no |
| 45 | input/- | entrada | itemSku | itemSku | no |
| 46 | input/- | entrada | itemOrigen | itemOrigen | no |
| 47 | input/- | entrada | itemDestino | PACK-01 | no |
| 48 | input/number | entrada | itemCantidad | 1 | no |
| 49 | input/- | entrada | itemLote | itemLote | no |
| 50 | input/- | entrada | itemSerial | itemSerial | no |
| 51 | button/submit | accion | - | Agregar item | no |
| 52 | input/number | entrada | avanceItemId | avanceItemId | no |
| 53 | input/number | entrada | avancePick | 0 | no |
| 54 | input/number | entrada | avancePack | 0 | no |
| 55 | select/- | entrada | avanceEstado | Inferir automatico En picking Pickeado En packing Empacado Completado Cancelado | no |
| 56 | button/submit | accion | - | Guardar avance | no |
| 57 | select/- | entrada | desOrdenId | desOrdenId | no |
| 58 | input/- | entrada | desCodigo | desCodigo | no |
| 59 | select/- | entrada | desEstado | Programado En ruta Entregado Devuelto Cancelado | no |
| 60 | input/- | entrada | desTransportadora | desTransportadora | no |
| 61 | input/- | entrada | desGuia | desGuia | no |
| 62 | input/- | entrada | desConductor | desConductor | no |
| 63 | input/- | entrada | desVehiculo | desVehiculo | no |
| 64 | input/- | entrada | desRuta | desRuta | no |
| 65 | input/number | entrada | desFlete | 0 | no |
| 66 | input/date | entrada | desSalida | desSalida | no |
| 67 | input/date | entrada | desEntrega | desEntrega | no |
| 68 | button/submit | accion | - | Guardar despacho | no |

### `web/administrar_empresa/mi_horario.html` (3)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/date | entrada | desde | desde | no |
| 2 | input/date | entrada | hasta | hasta | no |
| 3 | button/button | accion | btnActualizar | Actualizar | no |

### `web/administrar_empresa/modulo_menu.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | - | &#9776; Ocultar menu | sí |

### `web/administrar_empresa/nextcloud.html` (5)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | nextcloudProvision | Preparar espacio | no |
| 2 | button/button | accion | nextcloudReset | Restablecer acceso | no |
| 3 | button/button | accion | nextcloudToggle | Desactivar espacio | no |
| 4 | button/button | accion | nextcloudOpen | Abrir Nextcloud | no |
| 5 | button/button | accion | nextcloudCopy | Copiar credencial | no |

### `web/administrar_empresa/niif.html` (28)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnRefresh | Actualizar | no |
| 2 | button/button | accion | btnExportJSON | JSON | no |
| 3 | button/button | accion | btnExportTXT | Reporte | no |
| 4 | button/button | accion | - | Diagnostico | sí |
| 5 | button/button | accion | - | Politicas | sí |
| 6 | button/button | accion | - | Medicion | sí |
| 7 | button/button | accion | - | Conciliacion | sí |
| 8 | button/button | accion | - | Cierre | sí |
| 9 | button/button | accion | - | Notas | sí |
| 10 | input/number | entrada | impCarrying | 10000000 | no |
| 11 | input/number | entrada | impRecoverable | 9200000 | no |
| 12 | input/number | entrada | depCost | 48000000 | no |
| 13 | input/number | entrada | depResidual | 3000000 | no |
| 14 | input/number | entrada | depMonths | 60 | no |
| 15 | input/number | entrada | fairMarket | 55000000 | no |
| 16 | input/number | entrada | fairCosts | 1200000 | no |
| 17 | input/number | entrada | fairUse | 52000000 | no |
| 18 | input/number | entrada | taxProfit | 85000000 | no |
| 19 | input/number | entrada | taxPermanent | 0 | no |
| 20 | input/number | entrada | taxTemporary | 12000000 | no |
| 21 | input/number | entrada | taxRate | 35 | no |
| 22 | textarea/- | entrada | noteJudgement | La administracion revisa deterioro, vida util, valor residual, provisiones, ingresos y clasificacion de instrumentos fin | no |
| 23 | textarea/- | entrada | noteRisk | Riesgo de credito en cartera, liquidez operativa, concentracion de proveedores y exposicion a cambios de precios segun l | no |
| 24 | input/checkbox | accion | - | sin etiqueta | sí |
| 25 | select/- | entrada | - | Definida En revision Pendiente | sí |
| 26 | input/- | entrada | - | '+row.owner.replace(/ | sí |
| 27 | input/- | entrada | - | '+row.review.replace(/ | sí |
| 28 | input/checkbox | accion | - | sin etiqueta | sí |

### `web/administrar_empresa/nomina_menu.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | - | &#9776; Ocultar menu | sí |

### `web/administrar_empresa/nomina_sueldos.html` (100)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | nominaTutorialPage | Tutorial | no |
| 2 | a/tutorial | accion | nominaHelpTutorial | ? Ayuda | sí |
| 3 | button/button | accion | btnRecargarTodo | Actualizar | no |
| 4 | button/button | accion | btnNominaAgenteInternet | Agente internet | no |
| 5 | input/checkbox | accion | legalAutoUpdate | legalAutoUpdate | no |
| 6 | button/button | accion | btnAplicarParametrosLegales | Aplicar actualizaci&oacute;n legal | no |
| 7 | input/- | entrada | cfgPais | CO | no |
| 8 | input/- | entrada | cfgMoneda | COP | no |
| 9 | input/number | entrada | cfgSalarioMinimo | 1750905 | no |
| 10 | input/number | entrada | cfgAuxilioLegal | 249095 | no |
| 11 | input/number | entrada | cfgHorasSemana | 42 | no |
| 12 | input/number | entrada | cfgHorasDia | 8 | no |
| 13 | input/number | entrada | cfgDiasMes | 30 | no |
| 14 | input/number | entrada | cfgDivisorHora | 210 | no |
| 15 | input/time | entrada | cfgNocDesde | 19:00:00 | no |
| 16 | input/time | entrada | cfgNocHasta | 06:00:00 | no |
| 17 | input/number | entrada | cfgRecargoNocturno | 35 | no |
| 18 | input/number | entrada | cfgExtraDiurna | 25 | no |
| 19 | input/number | entrada | cfgExtraNocturna | 75 | no |
| 20 | input/number | entrada | cfgDomDiurno | 75 | no |
| 21 | input/number | entrada | cfgDomNocturno | 110 | no |
| 22 | input/number | entrada | cfgExtraDomDiurna | 100 | no |
| 23 | input/number | entrada | cfgExtraDomNocturna | 150 | no |
| 24 | input/number | entrada | cfgDedSalud | 4 | no |
| 25 | input/number | entrada | cfgDedPension | 4 | no |
| 26 | input/number | entrada | cfgDedSolidaridad | 0 | no |
| 27 | textarea/- | entrada | cfgObservaciones | cfgObservaciones | no |
| 28 | input/number | entrada | cfgAporteSaludEmp | 8.5 | no |
| 29 | input/number | entrada | cfgAportePensionEmp | 12 | no |
| 30 | input/number | entrada | cfgAporteARL | 0.522 | no |
| 31 | input/number | entrada | cfgCajaComp | 4 | no |
| 32 | input/number | entrada | cfgICBF | 3 | no |
| 33 | input/number | entrada | cfgSENA | 2 | no |
| 34 | input/number | entrada | cfgCesantias | 8.33 | no |
| 35 | input/number | entrada | cfgIntCesantias | 1 | no |
| 36 | input/number | entrada | cfgPrima | 8.33 | no |
| 37 | input/number | entrada | cfgVacaciones | 4.17 | no |
| 38 | button/submit | accion | - | Guardar configuración | no |
| 39 | input/hidden | entrada | empId | empId | no |
| 40 | input/- | entrada | empCodigo | empCodigo | no |
| 41 | input/- | entrada | empNombre | empNombre | no |
| 42 | input/- | entrada | empDocumento | empDocumento | no |
| 43 | input/- | entrada | empCargo | empCargo | no |
| 44 | input/- | entrada | empSedeCodigo | empSedeCodigo | no |
| 45 | input/- | entrada | empSedeNombre | empSedeNombre | no |
| 46 | input/- | entrada | empCentroCosto | empCentroCosto | no |
| 47 | select/- | entrada | empTipoContrato | Indefinido Fijo Obra labor Aprendizaje | no |
| 48 | input/date | entrada | empFechaIngreso | empFechaIngreso | no |
| 49 | input/number | entrada | empSalario | 0 | no |
| 50 | input/number | entrada | empAuxilio | 0 | no |
| 51 | input/number | entrada | empBonificacion | 0 | no |
| 52 | input/number | entrada | empDeduccionFija | 0 | no |
| 53 | input/number | entrada | empJornadaDia | 8 | no |
| 54 | select/- | entrada | empAuxilioFlag | Incluir No incluir | no |
| 55 | button/submit | accion | - | Guardar empleado | no |
| 56 | button/button | accion | empCancelar | Cancelar | no |
| 57 | input/- | entrada | empBuscar | empBuscar | no |
| 58 | button/button | accion | empBuscarBtn | Buscar | no |
| 59 | button/button | accion | empLimpiarBtn | Limpiar | no |
| 60 | input/date | entrada | festFecha | festFecha | no |
| 61 | input/- | entrada | festDescripcion | festDescripcion | no |
| 62 | button/submit | accion | - | Agregar festivo | no |
| 63 | input/date | entrada | festDesde | festDesde | no |
| 64 | input/date | entrada | festHasta | festHasta | no |
| 65 | button/button | accion | festBuscarBtn | Filtrar | no |
| 66 | button/button | accion | festLimpiarBtn | Limpiar | no |
| 67 | input/date | entrada | calcDesde | calcDesde | no |
| 68 | input/date | entrada | calcHasta | calcHasta | no |
| 69 | select/- | entrada | calcEmpleado | Todos los empleados | no |
| 70 | input/number | entrada | calcOtrasDeducciones | 0 | no |
| 71 | select/- | entrada | calcOverwrite | Si, reemplazar periodo No, mantener existentes | no |
| 72 | button/submit | accion | - | Calcular nomina | no |
| 73 | button/button | accion | calcListarBtn | Consultar liquidaciones | no |
| 74 | button/button | accion | calcDesprendibleBtn | Generar desprendible | no |
| 75 | select/- | entrada | calcConciliarFix | Conciliar: solo auditar Conciliar y recalcular | no |
| 76 | button/button | accion | calcConciliarBtn | Conciliar asistencia | no |
| 77 | select/- | entrada | calcExportFormat | PDF XLS (Excel) CSV JSON TXT | no |
| 78 | button/button | accion | calcExportBtn | Exportar liquidaciones | no |
| 79 | input/checkbox | accion | payConfirmControl | payConfirmControl | no |
| 80 | button/button | accion | controlValidarBtn | Validar control contable | no |
| 81 | select/- | entrada | payMetodo | Transferencia bancaria Efectivo Nequi Daviplata Cheque | no |
| 82 | input/- | entrada | payCuenta | payCuenta | no |
| 83 | button/button | accion | payGenerarBtn | Generar pagos del período | no |
| 84 | button/button | accion | payConsultarBtn | Consultar pagos | no |
| 85 | button/button | accion | provConsultarBtn | Consultar provisiones | no |
| 86 | button/button | accion | nomCoRefreshBtn | Actualizar | no |
| 87 | button/button | accion | nomCoSeedBtn | Cargar parametros demo | no |
| 88 | button/button | accion | nomCoPilaBtn | Generar PILA | no |
| 89 | button/button | accion | nomCoSeedProfesionalBtn | Crear nomina demo Motel Calipso | no |
| 90 | button/button | accion | nomCoDianVerBtn | Ver estado DIAN | no |
| 91 | button/button | accion | nomCoDianPrepararBtn | Preparar lote DIAN | no |
| 92 | button/button | accion | nomCoDianEnviarBtn | Enviar nomina electronica a DIAN | no |
| 93 | button/button | accion | ' + id + ' | Editar | sí |
| 94 | button/button | accion | ' + id + ' | ' + actionLabel + ' | sí |
| 95 | button/button | accion | ' + id + ' | Eliminar | sí |
| 96 | button/button | accion | - | Eliminar | sí |
| 97 | button/button | accion | - | Desprendible | sí |
| 98 | button/button | accion | - | Aplicar este dato | sí |
| 99 | button/button | accion | - | Aprobar | sí |
| 100 | button/button | accion | - | Rechazar | sí |

### `web/administrar_empresa/nomina_tutorial.html` (6)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Volver a nomina | sí |
| 2 | a/- | accion | - | Ver guia guiada | sí |
| 3 | a/- | accion | - | Abrir DIAN | no |
| 4 | a/- | accion | - | Ver micrositio | no |
| 5 | a/- | accion | - | Abrir tecnica | no |
| 6 | a/- | accion | - | Descargar PDF | no |

### `web/administrar_empresa/panel.html` (27)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | weatherGpsBtn | Usar GPS | no |
| 2 | button/button | accion | weatherManualToggle | Cambiar ciudad | no |
| 3 | input/text | entrada | weatherManualInput | weatherManualInput | no |
| 4 | button/submit | accion | - | Aplicar | no |
| 5 | button/button | accion | empresaNoticiasHideBtn | Ocultar noticia | no |
| 6 | select/- | entrada | empresaBuzonMode | Enviar mensaje Asignar tarea | no |
| 7 | select/- | entrada | empresaBuzonRecipient | Seleccionar usuario | no |
| 8 | input/text | entrada | empresaBuzonTitleInput | empresaBuzonTitleInput | no |
| 9 | textarea/- | entrada | empresaBuzonMessageInput | empresaBuzonMessageInput | no |
| 10 | select/- | entrada | empresaTaskPriority | Prioridad normal Prioridad alta Prioridad baja | no |
| 11 | input/datetime-local | entrada | empresaTaskDue | Fecha limite de la tarea | no |
| 12 | input/file | accion | empresaBuzonFileInput | empresaBuzonFileInput | no |
| 13 | button/button | accion | empresaBuzonRecordBtn | Grabar audio | no |
| 14 | button/submit | accion | empresaBuzonSendBtn | Enviar | no |
| 15 | textarea/- | entrada | empresaTaskCompleteDescription | empresaTaskCompleteDescription | no |
| 16 | input/file | accion | empresaTaskEvidenceInput | empresaTaskEvidenceInput | no |
| 17 | button/button | accion | empresaTaskCompleteCancel | Cancelar | no |
| 18 | button/button | accion | empresaTaskCompleteSave | Finalizar tarea | no |
| 19 | textarea/- | entrada | empresaChatMessageInput | empresaChatMessageInput | no |
| 20 | button/submit | accion | empresaChatSendBtn | Enviar al chat | no |
| 21 | select/- | entrada | - | ' + ' Si ' + ' No ' + ' | sí |
| 22 | select/- | entrada | - | ' + question.options.map(function (option) { option = String(option \|\| ""); return ' ' + escapeHTML(option) + ' '; }).jo | sí |
| 23 | textarea/- | entrada | - | ' + escapeHTML(value) + ' | sí |
| 24 | input/' + (type === | entrada | - | ' + escapeAttr(value) + ' | sí |
| 25 | button/button | accion | empresaGuidedClose | No volver a mostrar | no |
| 26 | input/checkbox | accion | empresaGuidedNoMostrarMas | empresaGuidedNoMostrarMas | no |
| 27 | button/submit | accion | empresaGuidedApply | Aplicar configuracion | no |

### `web/administrar_empresa/parqueadero.html` (34)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | - | Operacion | sí |
| 2 | button/button | accion | - | Salida por QR | sí |
| 3 | button/button | accion | - | Tarifas | sí |
| 4 | button/button | accion | - | Recibo | sí |
| 5 | input/- | entrada | placa | placa | no |
| 6 | select/- | entrada | tipoVehiculo | Carro Moto Camioneta Camion Bicicleta | no |
| 7 | input/- | entrada | clienteNombre | clienteNombre | no |
| 8 | input/- | entrada | clienteDocumento | clienteDocumento | no |
| 9 | textarea/- | entrada | observaciones | observaciones | no |
| 10 | button/button | accion | btnEmitirTicket | Emitir ticket | no |
| 11 | button/button | accion | btnRefrescar | Refrescar | no |
| 12 | input/- | entrada | salidaToken | salidaToken | no |
| 13 | button/button | accion | btnValidarQR | Calcular salida | no |
| 14 | button/button | accion | btnCobrarQR | Cobrar y cerrar | no |
| 15 | input/- | entrada | cfgNombre | cfgNombre | no |
| 16 | input/- | entrada | cfgPrefijo | cfgPrefijo | no |
| 17 | input/- | entrada | cfgMoneda | cfgMoneda | no |
| 18 | input/number | entrada | cfgTolerancia | cfgTolerancia | no |
| 19 | input/number | entrada | cfgMinBase | cfgMinBase | no |
| 20 | input/number | entrada | cfgTarifaBase | cfgTarifaBase | no |
| 21 | input/number | entrada | cfgMinFraccion | cfgMinFraccion | no |
| 22 | input/number | entrada | cfgTarifaFraccion | cfgTarifaFraccion | no |
| 23 | input/number | entrada | cfgDiaMax | cfgDiaMax | no |
| 24 | input/number | entrada | cfgIva | cfgIva | no |
| 25 | input/checkbox | accion | cfgFraccionCompleta | cfgFraccionCompleta | no |
| 26 | input/checkbox | accion | cfgIvaIncluido | cfgIvaIncluido | no |
| 27 | input/checkbox | accion | cfgSalidaQR | cfgSalidaQR | no |
| 28 | input/checkbox | accion | cfgPrintEntrada | cfgPrintEntrada | no |
| 29 | input/checkbox | accion | cfgPrintSalida | cfgPrintSalida | no |
| 30 | button/button | accion | btnGuardarConfig | Guardar tarifas | no |
| 31 | button/button | accion | btnImprimirRecibo | Imprimir recibo | no |
| 32 | button/button | accion | - | Calcular | sí |
| 33 | button/button | accion | - | Cobrar salida | sí |
| 34 | button/button | accion | - | Anular | sí |

### `web/administrar_empresa/portal_contador.html` (43)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnDemo | Demo | no |
| 2 | button/button | accion | btnExportar | Exportar CSV | no |
| 3 | button/button | accion | btnRefrescar | Actualizar | no |
| 4 | input/- | entrada | filtro_q | filtro_q | no |
| 5 | button/button | accion | btnFiltrar | Filtrar | no |
| 6 | input/hidden | entrada | cliente_id | cliente_id | no |
| 7 | input/- | entrada | razon_social | razon_social | no |
| 8 | input/- | entrada | nit | nit | no |
| 9 | select/- | entrada | regimen | Responsable IVA No responsable IVA Simple Gran contribuyente Persona natural | no |
| 10 | select/- | entrada | periodicidad | Mensual Bimestral Trimestral Semestral Anual | no |
| 11 | select/- | entrada | riesgo_nivel | Bajo Medio Alto Critico | no |
| 12 | input/- | entrada | responsable | responsable | no |
| 13 | input/- | entrada | contacto_nombre | contacto_nombre | no |
| 14 | input/email | entrada | contacto_email | contacto_email | no |
| 15 | input/- | entrada | contacto_telefono | contacto_telefono | no |
| 16 | input/date | entrada | fecha_inicio | fecha_inicio | no |
| 17 | input/number | entrada | cierre_mensual_dia | 10 | no |
| 18 | select/- | entrada | estado_cliente | Activo Pausado En revision Retirado | no |
| 19 | textarea/- | entrada | observaciones_cliente | observaciones_cliente | no |
| 20 | button/submit | accion | - | Guardar cliente | no |
| 21 | button/button | accion | btnNuevoCliente | Nuevo | no |
| 22 | button/button | accion | - | Obligaciones | sí |
| 23 | button/button | accion | - | Solicitudes | sí |
| 24 | button/button | accion | - | Comunicaciones | sí |
| 25 | select/- | entrada | ob_cliente | ob_cliente | no |
| 26 | select/- | entrada | ob_tipo | IVA Retencion fuente ICA Renta Exogena Nomina electronica Cierre contable | no |
| 27 | input/- | entrada | ob_periodo | ob_periodo | no |
| 28 | input/date | entrada | ob_vence | ob_vence | no |
| 29 | select/- | entrada | ob_prioridad | Media Alta Critica Baja | no |
| 30 | input/number | entrada | ob_valor | 0 | no |
| 31 | button/submit | accion | - | Agregar obligacion | no |
| 32 | select/- | entrada | sol_cliente | sol_cliente | no |
| 33 | input/- | entrada | sol_titulo | sol_titulo | no |
| 34 | select/- | entrada | sol_categoria | Soportes Extractos Facturas Nomina Impuestos Contratos | no |
| 35 | input/date | entrada | sol_limite | sol_limite | no |
| 36 | select/- | entrada | sol_prioridad | Media Alta Critica Baja | no |
| 37 | button/submit | accion | - | Crear solicitud | no |
| 38 | select/- | entrada | com_cliente | com_cliente | no |
| 39 | select/- | entrada | com_canal | Interno Email WhatsApp Llamada Reunion | no |
| 40 | input/- | entrada | com_asunto | com_asunto | no |
| 41 | textarea/- | entrada | com_mensaje | com_mensaje | no |
| 42 | button/submit | accion | - | Registrar comunicacion | no |
| 43 | button/button | accion | - | Editar | sí |

### `web/administrar_empresa/portal_terceros_certificados.html` (40)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnRefresh | Actualizar | no |
| 2 | button/button | accion | btnSeed | Cargar demo | no |
| 3 | button/button | accion | btnExport | Exportar CSV | no |
| 4 | button/button | accion | - | Dashboard | sí |
| 5 | button/button | accion | - | Terceros | sí |
| 6 | button/button | accion | - | Certificados | sí |
| 7 | button/button | accion | - | Descargas | sí |
| 8 | input/hidden | entrada | terceroId | terceroId | no |
| 9 | select/- | entrada | tipoTercero | Proveedor Cliente Empleado Contratista Contador Accionista Otro | no |
| 10 | input/- | entrada | documento | documento | no |
| 11 | input/- | entrada | dv | dv | no |
| 12 | input/- | entrada | razonSocial | razonSocial | no |
| 13 | input/email | entrada | email | email | no |
| 14 | input/- | entrada | telefono | telefono | no |
| 15 | input/- | entrada | ciudad | ciudad | no |
| 16 | select/- | entrada | regimen | Responsable IVA No responsable IVA Simple Gran contribuyente Persona natural No residente | no |
| 17 | select/- | entrada | estadoTercero | Activo Inactivo Bloqueado Archivado | no |
| 18 | input/- | entrada | direccion | direccion | no |
| 19 | textarea/- | entrada | obsTercero | obsTercero | no |
| 20 | button/submit | accion | - | Guardar tercero | no |
| 21 | input/hidden | entrada | certId | certId | no |
| 22 | select/- | entrada | certTercero | certTercero | no |
| 23 | select/- | entrada | tipoCert | Retención en la fuente Retención IVA Retención ICA Ingresos y retenciones Certificado proveedor Certificado cliente Auto | no |
| 24 | input/number | entrada | anio | anio | no |
| 25 | input/date | entrada | desde | desde | no |
| 26 | input/date | entrada | hasta | hasta | no |
| 27 | input/- | entrada | concepto | concepto | no |
| 28 | input/number | entrada | baseValor | 0 | no |
| 29 | input/number | entrada | retFuente | 0 | no |
| 30 | input/number | entrada | retIVA | 0 | no |
| 31 | input/number | entrada | retICA | 0 | no |
| 32 | input/number | entrada | otros | 0 | no |
| 33 | select/- | entrada | estadoCert | Emitido Borrador Enviado Anulado | no |
| 34 | input/- | entrada | firmaNombre | firmaNombre | no |
| 35 | input/- | entrada | firmaCargo | firmaCargo | no |
| 36 | textarea/- | entrada | obsCert | obsCert | no |
| 37 | button/submit | accion | - | Guardar certificado | no |
| 38 | button/- | accion | - | Editar | sí |
| 39 | button/- | accion | - | Ver | sí |
| 40 | button/- | accion | - | Copiar | sí |

### `web/administrar_empresa/produccion_mrp.html` (66)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | prodRefresh | Actualizar | no |
| 2 | button/button | accion | prodSeed | Cargar demo | no |
| 3 | button/button | accion | prodExport | Exportar CSV | no |
| 4 | button/button | accion | prodTutorial | Tutorial | no |
| 5 | button/button | accion | prodGenerateMRP | Generar MRP | no |
| 6 | button/button | accion | - | Tablero | sí |
| 7 | button/button | accion | - | Recetas / BOM | sí |
| 8 | button/button | accion | - | Ordenes | sí |
| 9 | button/button | accion | - | Consumos | sí |
| 10 | button/button | accion | - | Calidad | sí |
| 11 | button/button | accion | - | Plan MRP | sí |
| 12 | button/button | accion | - | Configuracion | sí |
| 13 | button/button | accion | - | Gestionar BOM | sí |
| 14 | button/button | accion | - | Programar orden | sí |
| 15 | button/button | accion | - | Revisar MRP | sí |
| 16 | select/- | entrada | recetaProductoSelect | recetaProductoSelect | no |
| 17 | button/button | accion | importRecetaProducto | Importar como BOM | no |
| 18 | input/hidden | entrada | recId | recId | no |
| 19 | input/- | entrada | recCodigo | recCodigo | no |
| 20 | input/- | entrada | recVersion | 1.0 | no |
| 21 | input/- | entrada | recNombre | recNombre | no |
| 22 | input/- | entrada | recProducto | recProducto | no |
| 23 | input/- | entrada | recUnidad | und | no |
| 24 | input/number | entrada | recCantidad | 1 | no |
| 25 | input/number | entrada | recCosto | 0 | no |
| 26 | input/number | entrada | recMerma | 0 | no |
| 27 | input/number | entrada | recTiempo | 0 | no |
| 28 | select/- | entrada | recEstado | Activo Borrador Inactivo | no |
| 29 | textarea/- | entrada | recComponentes | Materia prima principal\|1\|und\|0\|0\|produccion | no |
| 30 | button/submit | accion | recSubmit | Guardar receta | no |
| 31 | button/button | accion | recCancel | Limpiar | no |
| 32 | select/- | entrada | recFilter | Todos Activas Borrador Inactivas | no |
| 33 | input/hidden | entrada | ordId | ordId | no |
| 34 | select/- | entrada | ordReceta | ordReceta | no |
| 35 | input/number | entrada | ordCantidad | 1 | no |
| 36 | select/- | entrada | ordPrioridad | Normal Alta Urgente Baja | no |
| 37 | input/date | entrada | ordFecha | ordFecha | no |
| 38 | input/- | entrada | ordResponsable | ordResponsable | no |
| 39 | select/- | entrada | ordEstado | Programada Borrador En proceso | no |
| 40 | textarea/- | entrada | ordObs | ordObs | no |
| 41 | button/submit | accion | ordSubmit | Crear orden | no |
| 42 | button/button | accion | ordCancel | Limpiar | no |
| 43 | select/- | entrada | ordenFilter | Todas Borrador Programadas En proceso Calidad Cerradas Canceladas | no |
| 44 | input/- | entrada | ordenSearch | ordenSearch | no |
| 45 | select/- | entrada | consOrden | consOrden | no |
| 46 | input/- | entrada | consProducto | consProducto | no |
| 47 | input/- | entrada | consLote | consLote | no |
| 48 | input/number | entrada | consCantidad | 1 | no |
| 49 | input/number | entrada | consCosto | 0 | no |
| 50 | button/submit | accion | - | Guardar consumo | no |
| 51 | select/- | entrada | calOrden | calOrden | no |
| 52 | select/- | entrada | calResultado | Pendiente Aprobado Rechazado Reproceso | no |
| 53 | input/number | entrada | calAprobada | 0 | no |
| 54 | input/number | entrada | calRechazada | 0 | no |
| 55 | input/- | entrada | calResponsable | calResponsable | no |
| 56 | textarea/- | entrada | calObs | calObs | no |
| 57 | button/submit | accion | - | Guardar calidad | no |
| 58 | input/- | entrada | mrpPeriodo | mrpPeriodo | no |
| 59 | button/button | accion | mrpGenerate | Generar periodo | no |
| 60 | input/- | entrada | cfgNombre | cfgNombre | no |
| 61 | input/- | entrada | cfgMoneda | COP | no |
| 62 | select/- | entrada | cfgCosteo | Estandar Promedio Real | no |
| 63 | input/checkbox | accion | cfgAprobar | cfgAprobar | no |
| 64 | input/checkbox | accion | cfgConsumir | cfgConsumir | no |
| 65 | input/checkbox | accion | cfgCalidad | cfgCalidad | no |
| 66 | button/submit | accion | - | Guardar configuracion | no |

### `web/administrar_empresa/produccion_mrp_tutorial.html` (6)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Ir | sí |
| 2 | a/- | accion | - | Ir | sí |
| 3 | a/- | accion | - | Ir | sí |
| 4 | a/- | accion | - | Ir | sí |
| 5 | a/- | accion | - | Ir | sí |
| 6 | a/- | accion | - | Ir | sí |

### `web/administrar_empresa/propinas.html` (34)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/checkbox | accion | cfgHabilitar | cfgHabilitar | no |
| 2 | input/number | entrada | cfgPorcentaje | 10 | no |
| 3 | select/- | entrada | cfgModo | Por usuario Universal (repartida en usuarios activos) | no |
| 4 | input/checkbox | accion | cfgAplicarAuto | cfgAplicarAuto | no |
| 5 | input/text | entrada | cfgPaisFiscal | cfgPaisFiscal | no |
| 6 | input/text | entrada | cfgRegimenFiscal | cfgRegimenFiscal | no |
| 7 | select/- | entrada | cfgTratamientoFiscal | No gravada Gravada | no |
| 8 | input/number | entrada | cfgPctImpuestoPropina | 0 | no |
| 9 | textarea/- | entrada | cfgObservaciones | cfgObservaciones | no |
| 10 | button/button | accion | btnGuardarCfg | Guardar configuración | no |
| 11 | button/button | accion | btnActivarCfg | Activar | no |
| 12 | button/button | accion | btnDesactivarCfg | Desactivar | no |
| 13 | input/date | entrada | fDesde | fDesde | no |
| 14 | input/date | entrada | fHasta | fHasta | no |
| 15 | input/- | entrada | fUsuario | fUsuario | no |
| 16 | select/- | entrada | fModo | Todos Por usuario Universal | no |
| 17 | select/- | entrada | fOrigen | Todos Venta Ajuste manual | no |
| 18 | input/number | entrada | fCierreCaja | fCierreCaja | no |
| 19 | input/checkbox | accion | fSoloAjustes | fSoloAjustes | no |
| 20 | input/number | entrada | fLimit | 200 | no |
| 21 | button/button | accion | btnRefrescarReporte | Actualizar reporte | no |
| 22 | input/number | entrada | ajCarritoId | ajCarritoId | no |
| 23 | input/number | entrada | ajCierreCajaId | ajCierreCajaId | no |
| 24 | input/text | entrada | ajUsuario | ajUsuario | no |
| 25 | select/- | entrada | ajModo | Por usuario Universal | no |
| 26 | input/number | entrada | ajMonto | ajMonto | no |
| 27 | select/- | entrada | ajTratamientoFiscal | Usar configuración No gravada Gravada | no |
| 28 | input/number | entrada | ajPctImpuesto | ajPctImpuesto | no |
| 29 | input/text | entrada | ajReferencia | ajReferencia | no |
| 30 | input/text | entrada | ajVentaRef | ajVentaRef | no |
| 31 | textarea/- | entrada | ajMotivo | ajMotivo | no |
| 32 | button/button | accion | btnRegistrarAjuste | Registrar ajuste manual | no |
| 33 | input/number | entrada | ccCierreCajaId | ccCierreCajaId | no |
| 34 | button/button | accion | btnConciliarCierre | Ejecutar conciliacion | no |

### `web/administrar_empresa/proveedores_firma_digital.html` (2)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Comprar en Sensiyo | no |
| 2 | a/- | accion | - | Volver a cargar firma | no |

### `web/administrar_empresa/publicar_red_social.html` (10)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | visitBusinessProfileBtn | Ver perfil de la empresa | no |
| 2 | a/- | accion | visitSocialBtn | Visitar red social | no |
| 3 | input/text | entrada | pubNombre | pubNombre | no |
| 4 | textarea/- | entrada | pubDesc | pubDesc | no |
| 5 | input/file | accion | pubFotoFile | pubFotoFile | no |
| 6 | input/text | entrada | pubFoto | pubFoto | no |
| 7 | button/button | accion | btnSubirFoto | Subir foto | no |
| 8 | input/text | entrada | pubYoutube | pubYoutube | no |
| 9 | button/- | accion | - | Publicar post | sí |
| 10 | button/- | accion | - | Eliminar | sí |

### `web/administrar_empresa/radio_online.html` (8)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/checkbox | accion | radioOnlineEnabled | radioOnlineEnabled | no |
| 2 | select/- | entrada | radioCountrySelect | Detectar automaticamente Panama Ecuador | no |
| 3 | input/text | entrada | radioCustomName | Nombre de emisora | no |
| 4 | input/text | entrada | radioCustomGenre | Genero de emisora | no |
| 5 | input/url | entrada | radioCustomStream | URL de streaming | no |
| 6 | input/url | entrada | radioCustomSource | Sitio web de la emisora | no |
| 7 | select/- | entrada | radioCustomCountry | Personalizada Panama Ecuador | no |
| 8 | button/submit | accion | - | Agregar | no |

### `web/administrar_empresa/rappi.html` (21)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnRappiTest | Probar API | no |
| 2 | button/button | accion | btnRappiSave | Guardar | no |
| 3 | select/- | entrada | rappiActivo | No Si | no |
| 4 | select/- | entrada | rappiAmbiente | Desarrollo Produccion | no |
| 5 | input/- | entrada | rappiCountryDomain | rappiCountryDomain | no |
| 6 | input/- | entrada | rappiNewDomain | rappiNewDomain | no |
| 7 | input/- | entrada | rappiClientID | rappiClientID | no |
| 8 | input/- | entrada | rappiClientSecretRef | rappiClientSecretRef | no |
| 9 | input/- | entrada | rappiWebhookSecretRef | rappiWebhookSecretRef | no |
| 10 | input/- | entrada | rappiStoreIntegrationID | rappiStoreIntegrationID | no |
| 11 | input/- | entrada | rappiStoreID | rappiStoreID | no |
| 12 | input/number | entrada | rappiCookingTime | 15 | no |
| 13 | input/checkbox | accion | rappiAutoTake | rappiAutoTake | no |
| 14 | input/checkbox | accion | rappiCrearVenta | rappiCrearVenta | no |
| 15 | textarea/- | entrada | rappiObservaciones | rappiObservaciones | no |
| 16 | button/button | accion | btnRappiStores | Tiendas | no |
| 17 | button/button | accion | btnRappiOrders | Ordenes nuevas | no |
| 18 | button/button | accion | btnRappiSent | Ordenes SENT | no |
| 19 | button/- | accion | - | Tomar | sí |
| 20 | button/- | accion | - | Listo | sí |
| 21 | button/- | accion | - | Rechazar | sí |

### `web/administrar_empresa/recetas_productos.html` (22)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | nuevoRecetaBtn | Nueva receta | no |
| 2 | input/hidden | entrada | recetaId | recetaId | no |
| 3 | input/- | entrada | recetaNombre | recetaNombre | no |
| 4 | input/- | entrada | recetaCodigo | recetaCodigo | no |
| 5 | input/number | entrada | recetaPrecio | 0 | no |
| 6 | input/number | entrada | recetaImpuesto | 0 | no |
| 7 | input/- | entrada | recetaUnidad | receta | no |
| 8 | textarea/- | entrada | recetaDescripcion | recetaDescripcion | no |
| 9 | select/- | entrada | recetaImpresoraPedido | recetaImpresoraPedido | no |
| 10 | textarea/- | entrada | recetaObservaciones | recetaObservaciones | no |
| 11 | select/- | entrada | ingredienteProducto | ingredienteProducto | no |
| 12 | input/number | entrada | ingredienteCantidad | 1 | no |
| 13 | input/- | entrada | ingredienteUnidad | ingredienteUnidad | no |
| 14 | button/button | accion | agregarIngredienteBtn | Agregar ingrediente | no |
| 15 | button/submit | accion | guardarRecetaBtn | Guardar receta | no |
| 16 | button/button | accion | cancelarRecetaBtn | Cancelar | no |
| 17 | input/- | entrada | buscarReceta | buscarReceta | no |
| 18 | button/button | accion | buscarRecetaBtn | Buscar | no |
| 19 | button/- | accion | - | Quitar | sí |
| 20 | button/- | accion | ' + Number(c.id \|\| 0) + ' | Editar | sí |
| 21 | button/- | accion | ' + Number(c.id \|\| 0) + ' | ' + toggleLabel + ' | sí |
| 22 | button/- | accion | ' + Number(c.id \|\| 0) + ' | Eliminar | sí |

### `web/administrar_empresa/renta_ia.html` (17)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnCalcular | Calcular | no |
| 2 | button/button | accion | btnIA | Analizar con IA | sí |
| 3 | input/date | entrada | desde | desde | no |
| 4 | input/date | entrada | hasta | hasta | no |
| 5 | input/number | entrada | tarifaRenta | 35 | no |
| 6 | input/number | entrada | sobretasaPuntos | 0 | no |
| 7 | input/number | entrada | anticipoRenta | 0 | no |
| 8 | input/number | entrada | ingresosNoConstitutivos | 0 | no |
| 9 | input/number | entrada | rentasExentas | 0 | no |
| 10 | input/number | entrada | deduccionesAdicionales | 0 | no |
| 11 | input/number | entrada | descuentosTributarios | 0 | no |
| 12 | input/number | entrada | retencionesAdicionales | 0 | no |
| 13 | input/- | entrada | preguntaIA | preguntaIA | no |
| 14 | input/checkbox | accion | usarVentasPOS | usarVentasPOS | no |
| 15 | input/checkbox | accion | usarMovIngresos | usarMovIngresos | no |
| 16 | input/checkbox | accion | usarMovEgresos | usarMovEgresos | no |
| 17 | input/checkbox | accion | usarComprasNomina | usarComprasNomina | no |

### `web/administrar_empresa/reporte_aseo_estaciones.html` (5)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | backBtn | Volver a estaciones | no |
| 2 | input/date | entrada | desde | desde | no |
| 3 | input/date | entrada | hasta | hasta | no |
| 4 | input/number | entrada | estacion_id | estacion_id | no |
| 5 | button/button | accion | filterBtn | Consultar | no |

### `web/administrar_empresa/reportes_ejecutivos.html` (30)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | select/- | entrada | reportsPicker | Cargando reportes... | no |
| 2 | select/- | entrada | reportsArea | Todas las areas Direccion Ventas y POS Inventario y compras Finanzas y cartera Contabilidad e impuestos Operacion y audi | no |
| 3 | input/search | entrada | reportsSearch | reportsSearch | no |
| 4 | input/date | entrada | reportsDesde | Desde | no |
| 5 | input/date | entrada | reportsHasta | Hasta | no |
| 6 | input/search | entrada | reportsUsuario | reportsUsuario | no |
| 7 | input/search | entrada | reportsCaja | reportsCaja | no |
| 8 | input/search | entrada | reportsTurno | reportsTurno | no |
| 9 | input/number | entrada | reportsCierreId | reportsCierreId | no |
| 10 | select/- | entrada | reportsPaperMode | Papel POS 80mm Papel carta | no |
| 11 | button/button | accion | btnPreviewReport | Vista previa | no |
| 12 | button/button | accion | btnPrintReport | Imprimir vista | no |
| 13 | button/button | accion | btnClearFilters | Limpiar filtros | no |
| 14 | button/button | accion | - | PDF | sí |
| 15 | button/button | accion | - | Excel | sí |
| 16 | button/button | accion | - | CSV | sí |
| 17 | button/button | accion | - | JSON | sí |
| 18 | button/button | accion | - | TXT | sí |
| 19 | input/text | entrada | reportsAIQuestion | Instruccion para nuevo reporte IA | no |
| 20 | button/button | accion | btnGenerateAIReport | Generar con IA | no |
| 21 | button/button | accion | btnSaveAIReportTemplate | Guardar plantilla | no |
| 22 | button/button | accion | btnExportAIReport | Exportar Excel | sí |
| 23 | button/button | accion | - | PDF | sí |
| 24 | button/button | accion | - | CSV | sí |
| 25 | button/button | accion | btnReloadPreview | Recargar vista | no |
| 26 | input/text | entrada | reportsAITemplateCode | reportsAITemplateCode | no |
| 27 | input/text | entrada | reportsAITemplateName | reportsAITemplateName | no |
| 28 | button/button | accion | btnCancelAIReportTemplate | Cancelar | no |
| 29 | button/submit | accion | btnConfirmAIReportTemplate | Guardar nueva versión | no |
| 30 | button/button | accion | - | Seleccionar | sí |

### `web/administrar_empresa/reportes_menu.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | - | &#9776; Ocultar menu | sí |

### `web/administrar_empresa/reportes_turnos.html` (16)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnReload | Actualizar | no |
| 2 | input/search | entrada | q | q | no |
| 3 | input/date | entrada | desde | Desde | no |
| 4 | input/date | entrada | hasta | Hasta | no |
| 5 | input/search | entrada | caja | caja | no |
| 6 | select/- | entrada | estado | Cerrados y aprobados Cerrado Aprobado Todos menos abiertos | no |
| 7 | select/- | entrada | format | PDF XLS CSV JSON TXT | no |
| 8 | select/- | entrada | paperMode | Papel POS 80mm Papel grande | no |
| 9 | input/email | entrada | emailTo | Correo destino | no |
| 10 | select/- | entrada | emailFormat | PDF XLS CSV JSON TXT | no |
| 11 | button/button | accion | btnSendEmail | Enviar por email | no |
| 12 | button/button | accion | ' + escapeHtml(t.id) + ' | Ver | sí |
| 13 | button/button | accion | ' + escapeHtml(t.id) + ' | Imprimir | sí |
| 14 | button/button | accion | ' + escapeHtml(t.id) + ' | Exportar | sí |
| 15 | button/button | accion | ' + escapeHtml(t.id) + ' | Compartir | sí |
| 16 | button/button | accion | ' + escapeHtml(t.id) + ' | Email | sí |

### `web/administrar_empresa/reservas_hotel.html` (35)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | addBtn | Agregar reserva | no |
| 2 | button/button | accion | disponibilidadHeroBtn | Consultar disponibilidad | no |
| 3 | input/hidden | entrada | itemId | itemId | no |
| 4 | input/hidden | entrada | empresa_id | empresa_id | no |
| 5 | input/number | entrada | estacion_id | estacion_id | no |
| 6 | input/- | entrada | codigo_reserva | codigo_reserva | no |
| 7 | select/- | entrada | moneda | COP USD EUR | no |
| 8 | input/- | entrada | cliente_nombre | cliente_nombre | no |
| 9 | input/- | entrada | cliente_documento | cliente_documento | no |
| 10 | input/email | entrada | cliente_email | cliente_email | no |
| 11 | input/- | entrada | cliente_telefono | cliente_telefono | no |
| 12 | input/number | entrada | cantidad_huespedes | 1 | no |
| 13 | input/number | entrada | monto_total | 0 | no |
| 14 | input/datetime-local | entrada | fecha_entrada | fecha_entrada | no |
| 15 | input/datetime-local | entrada | fecha_salida | fecha_salida | no |
| 16 | input/- | entrada | canal_origen | panel_empresa | no |
| 17 | textarea/- | entrada | observaciones | observaciones | no |
| 18 | button/submit | accion | saveBtn | Guardar | no |
| 19 | button/button | accion | cancelBtn | Cancelar | no |
| 20 | input/- | entrada | buscar | buscar | no |
| 21 | select/- | entrada | filtro_estado_reserva | Todos Pendiente pago Confirmada En curso Cancelada Expirada No show | no |
| 22 | select/- | entrada | filtro_estado_pago | Todos Pendiente Confirmado Cancelado Expirado | no |
| 23 | input/number | entrada | filtro_estacion_id | filtro_estacion_id | no |
| 24 | input/datetime-local | entrada | filtro_desde | filtro_desde | no |
| 25 | input/datetime-local | entrada | filtro_hasta | filtro_hasta | no |
| 26 | button/button | accion | filtrarBtn | Filtrar | no |
| 27 | button/button | accion | limpiarBtn | Limpiar | no |
| 28 | button/button | accion | disponibilidadBtn | Ver disponibilidad | no |
| 29 | button/button | accion | politicasBtn | Aplicar politicas | no |
| 30 | button/button | accion | ' + id + ' | Editar | sí |
| 31 | button/button | accion | ' + id + ' | Confirmar pago | sí |
| 32 | button/button | accion | ' + id + ' | Reconver. carrito | sí |
| 33 | button/button | accion | ' + id + ' | Cancelar | sí |
| 34 | button/button | accion | ' + id + ' | ' + nextLabel + " | sí |
| 35 | button/button | accion | ' + id + ' | Eliminar | sí |

### `web/administrar_empresa/soporte_remoto.html` (5)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Descargar | no |
| 2 | a/- | accion | - | Descargar | no |
| 3 | a/- | accion | - | Descargar | no |
| 4 | a/- | accion | - | Guía oficial | no |
| 5 | a/- | accion | - | Guía oficial | no |

### `web/administrar_empresa/soportes_compras_ia.html` (68)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | captureRefresh | Actualizar | no |
| 2 | button/button | accion | captureSeed | Cargar demo | no |
| 3 | button/button | accion | captureExport | Exportar CSV | no |
| 4 | button/button | accion | - | Tablero | sí |
| 5 | button/button | accion | - | Radicar soporte | sí |
| 6 | button/button | accion | - | Bandeja | sí |
| 7 | button/button | accion | - | Detalle y auditoria | sí |
| 8 | input/file | accion | archivo | archivo | no |
| 9 | select/- | entrada | tipoSoporte | Gasto Compra Documento soporte Servicio Recibo | no |
| 10 | select/- | entrada | documentoTipo | Factura de compra Documento soporte Cuenta de cobro Recibo de caja Gasto Otro | no |
| 11 | input/- | entrada | proveedorNombre | proveedor_nombre | no |
| 12 | input/- | entrada | proveedorNit | proveedor_nit | no |
| 13 | input/- | entrada | documentoNumero | documento_numero | no |
| 14 | input/date | entrada | fechaDocumento | fecha_documento | no |
| 15 | input/date | entrada | fechaVencimiento | fecha_vencimiento | no |
| 16 | input/number | entrada | subtotal | 0 | no |
| 17 | input/number | entrada | impuestoIVA | 0 | no |
| 18 | input/number | entrada | total | 0 | no |
| 19 | input/number | entrada | retencionFuente | 0 | no |
| 20 | input/number | entrada | retencionICA | 0 | no |
| 21 | input/number | entrada | retencionIVA | 0 | no |
| 22 | input/- | entrada | categoriaContable | categoria_contable | no |
| 23 | input/- | entrada | centroCosto | centro_costo | no |
| 24 | input/checkbox | accion | impactaInventario | impacta_inventario | no |
| 25 | textarea/- | entrada | observaciones | observaciones | no |
| 26 | button/submit | accion | btnRadicar | Radicar soporte | no |
| 27 | button/button | accion | btnLimpiar | Limpiar | no |
| 28 | select/- | entrada | estadoFilter | Todos Radicado Extraido En revision Aprobado Contabilizado Duplicado Rechazado | no |
| 29 | select/- | entrada | registroFilter | Activos Papelera Depuracion pendiente Depurados | no |
| 30 | select/- | entrada | tipoFilter | Todos Gasto Compra Documento soporte Servicio Recibo | no |
| 31 | input/- | entrada | searchFilter | searchFilter | no |
| 32 | button/button | accion | btnExtraer | Extraer IA | no |
| 33 | button/button | accion | btnCancelarIA | Cancelar IA | no |
| 34 | button/button | accion | btnAprobar | Aprobar | no |
| 35 | button/button | accion | btnRechazar | Rechazar | no |
| 36 | button/button | accion | btnContabilizar | Contabilizar | no |
| 37 | button/button | accion | btnEliminar | Enviar a papelera | no |
| 38 | button/button | accion | btnRestaurar | Recuperar | no |
| 39 | button/button | accion | btnPurgar | Depurar archivo | no |
| 40 | input/number | entrada | retencionDias | 90 | no |
| 41 | button/button | accion | btnRetencionPreview | Vista previa de retencion | no |
| 42 | button/button | accion | btnCuarentenaPreview | Diagnostico de cuarentena | no |
| 43 | input/hidden | entrada | editSoporteId | editSoporteId | no |
| 44 | select/- | entrada | editProveedor | Selecciona antes de contabilizar | no |
| 45 | input/- | entrada | editProveedorNombre | editProveedorNombre | no |
| 46 | input/- | entrada | editProveedorNit | editProveedorNit | no |
| 47 | select/- | entrada | editTipoSoporte | Gasto Compra Documento soporte Servicio Recibo | no |
| 48 | select/- | entrada | editDocumentoTipo | Factura de compra Documento soporte Cuenta de cobro Recibo de caja Gasto Otro | no |
| 49 | input/- | entrada | editDocumentoNumero | editDocumentoNumero | no |
| 50 | input/date | entrada | editFechaDocumento | editFechaDocumento | no |
| 51 | input/date | entrada | editFechaVencimiento | editFechaVencimiento | no |
| 52 | input/- | entrada | editMoneda | COP | no |
| 53 | input/number | entrada | editSubtotal | editSubtotal | no |
| 54 | input/number | entrada | editIVA | editIVA | no |
| 55 | input/number | entrada | editTotal | editTotal | no |
| 56 | input/number | entrada | editReteFuente | editReteFuente | no |
| 57 | input/number | entrada | editReteICA | editReteICA | no |
| 58 | input/number | entrada | editReteIVA | editReteIVA | no |
| 59 | input/- | entrada | editCategoria | editCategoria | no |
| 60 | input/- | entrada | editCentroCosto | editCentroCosto | no |
| 61 | input/checkbox | accion | editImpactaInventario | editImpactaInventario | no |
| 62 | textarea/- | entrada | editObservaciones | editObservaciones | no |
| 63 | button/submit | accion | btnGuardarRevision | Guardar revision | no |
| 64 | button/button | accion | captureActionClose | × | no |
| 65 | textarea/- | entrada | captureActionMotivo | captureActionMotivo | no |
| 66 | input/- | entrada | captureActionConfirmacion | captureActionConfirmacion | no |
| 67 | button/button | accion | captureActionCancel | Cancelar | no |
| 68 | button/submit | accion | captureActionSubmit | Confirmar | no |

### `web/administrar_empresa/suite_contador.html` (5)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | - | PC Portal contador | sí |
| 2 | button/button | accion | - | DT Declaraciones | sí |
| 3 | button/button | accion | - | IA Renta IA | sí |
| 4 | input/search | entrada | accountantSearch | accountantSearch | no |
| 5 | button/button | accion | - | Abrir | sí |

### `web/administrar_empresa/tarifas_de_hotel.html` (99)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | hotelTutorialLink | Tutorial | no |
| 2 | button/button | accion | - | Nueva tarifa por noche | sí |
| 3 | button/button | accion | - | Plan motel | sí |
| 4 | button/button | accion | - | Configurar day-use | sí |
| 5 | button/button | accion | - | Reglas globales | sí |
| 6 | button/button | accion | - | Por noche | sí |
| 7 | button/button | accion | - | Day-use y fracciones | sí |
| 8 | button/button | accion | - | Motel express | sí |
| 9 | button/button | accion | - | Reglas globales | sí |
| 10 | button/button | accion | - | Simuladores | sí |
| 11 | input/hidden | entrada | nightId | nightId | no |
| 12 | input/- | entrada | nightName | tarifa_hotel_cama_doble | no |
| 13 | select/- | entrada | nightStationSelect | nightStationSelect | no |
| 14 | input/number | entrada | nightStationId | nightStationId | no |
| 15 | input/- | entrada | nightStationCode | nightStationCode | no |
| 16 | input/- | entrada | nightStationName | nightStationName | no |
| 17 | input/- | entrada | nightService | hospedaje | no |
| 18 | input/number | entrada | nightValue | nightValue | no |
| 19 | input/number | entrada | nightPeopleFrom | 1 | no |
| 20 | input/number | entrada | nightPeopleTo | 0 | no |
| 21 | select/- | entrada | nightCurrency | COP USD EUR | no |
| 22 | input/time | entrada | nightCheckIn | 15:00 | no |
| 23 | input/time | entrada | nightCheckOut | 12:00 | no |
| 24 | input/number | entrada | nightPriority | 1 | no |
| 25 | select/- | entrada | nightStatus | Activo Inactivo | no |
| 26 | input/checkbox | accion | nightAutomatic | nightAutomatic | no |
| 27 | textarea/- | entrada | nightNotes | nightNotes | no |
| 28 | button/submit | accion | - | Guardar tarifa | no |
| 29 | button/button | accion | nightApplyAll | Aplicar a todas | no |
| 30 | button/button | accion | nightReset | Limpiar | no |
| 31 | input/hidden | entrada | dayUseId | dayUseId | no |
| 32 | select/- | entrada | dayUseStationSelect | dayUseStationSelect | no |
| 33 | input/number | entrada | dayUseStationId | dayUseStationId | no |
| 34 | input/- | entrada | dayUseStationCode | dayUseStationCode | no |
| 35 | input/- | entrada | dayUseStationName | dayUseStationName | no |
| 36 | select/- | entrada | dayUseCurrency | COP USD EUR | no |
| 37 | select/- | entrada | dayUseDayFrom | dayUseDayFrom | no |
| 38 | select/- | entrada | dayUseDayTo | dayUseDayTo | no |
| 39 | input/number | entrada | dayUseBaseMinutes | 180 | no |
| 40 | input/number | entrada | dayUseBaseValue | 0 | no |
| 41 | input/number | entrada | dayUseExtraMinutes | 60 | no |
| 42 | input/number | entrada | dayUseExtraValue | 0 | no |
| 43 | input/number | entrada | dayUsePriority | 1 | no |
| 44 | select/- | entrada | dayUseStatus | Activo Inactivo | no |
| 45 | input/checkbox | accion | dayUseFraction | dayUseFraction | no |
| 46 | textarea/- | entrada | dayUseNotes | dayUseNotes | no |
| 47 | button/submit | accion | - | Guardar regla | no |
| 48 | button/button | accion | dayUseApplyAll | Aplicar a todas | no |
| 49 | button/button | accion | dayUseReset | Limpiar | no |
| 50 | input/hidden | entrada | motelId | motelId | no |
| 51 | input/- | entrada | motelPlanName | motelPlanName | no |
| 52 | select/- | entrada | motelPlanType | Express Day-use Nocturno Amanecida Suite VIP Promocion | no |
| 53 | select/- | entrada | motelStationSelect | motelStationSelect | no |
| 54 | input/number | entrada | motelStationId | motelStationId | no |
| 55 | input/- | entrada | motelStationCode | motelStationCode | no |
| 56 | input/- | entrada | motelStationName | motelStationName | no |
| 57 | input/- | entrada | motelCategory | motelCategory | no |
| 58 | select/- | entrada | motelDayFrom | motelDayFrom | no |
| 59 | select/- | entrada | motelDayTo | motelDayTo | no |
| 60 | input/time | entrada | motelStartTime | 00:00 | no |
| 61 | input/time | entrada | motelEndTime | 23:59 | no |
| 62 | input/number | entrada | motelIncludedMinutes | 180 | no |
| 63 | input/number | entrada | motelBaseValue | 0 | no |
| 64 | input/number | entrada | motelExtraMinutes | 60 | no |
| 65 | input/number | entrada | motelExtraValue | 0 | no |
| 66 | input/number | entrada | motelTolerance | 10 | no |
| 67 | select/- | entrada | motelCurrency | COP USD EUR | no |
| 68 | input/number | entrada | motelPriority | 1 | no |
| 69 | select/- | entrada | motelStatus | Activo Inactivo | no |
| 70 | input/checkbox | accion | motelFraction | motelFraction | no |
| 71 | input/checkbox | accion | motelAutomatic | motelAutomatic | no |
| 72 | textarea/- | entrada | motelNotes | motelNotes | no |
| 73 | button/submit | accion | - | Guardar plan motel | no |
| 74 | button/button | accion | motelPresetExpress | Preset express | no |
| 75 | button/button | accion | motelPresetNight | Preset amanecida | no |
| 76 | button/button | accion | motelReset | Limpiar | no |
| 77 | select/- | entrada | ruleRoundingMode | Sin redondeo Hacia arriba Hacia abajo Matematico | no |
| 78 | input/number | entrada | ruleRoundingUnit | 100 | no |
| 79 | input/number | entrada | ruleDailyMin | 0 | no |
| 80 | input/number | entrada | ruleDailyMax | 0 | no |
| 81 | input/number | entrada | ruleTolerance | 0 | no |
| 82 | input/number | entrada | ruleCancelMinutes | 0 | no |
| 83 | input/checkbox | accion | ruleSensorAuto | ruleSensorAuto | no |
| 84 | input/checkbox | accion | ruleCancelEnabled | ruleCancelEnabled | no |
| 85 | textarea/- | entrada | ruleNotes | ruleNotes | no |
| 86 | button/submit | accion | - | Guardar reglas | no |
| 87 | button/button | accion | reloadAll | Recargar datos | no |
| 88 | select/- | entrada | simNightStation | simNightStation | no |
| 89 | input/number | entrada | simNightPeople | 2 | no |
| 90 | input/datetime-local | entrada | simNightStart | simNightStart | no |
| 91 | input/datetime-local | entrada | simNightEnd | simNightEnd | no |
| 92 | button/submit | accion | - | Calcular noche | no |
| 93 | select/- | entrada | simDayUseStation | simDayUseStation | no |
| 94 | select/- | entrada | simDayUseDay | simDayUseDay | no |
| 95 | input/number | entrada | simDayUseMinutes | 180 | no |
| 96 | button/submit | accion | - | Calcular day-use | no |
| 97 | select/- | entrada | simMotelPlan | simMotelPlan | no |
| 98 | input/number | entrada | simMotelMinutes | 180 | no |
| 99 | button/submit | accion | - | Calcular motel | no |

### `web/administrar_empresa/tarifas_de_motel.html` (43)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | motelTutorialLink | Tutorial | no |
| 2 | button/button | accion | reloadRates | Actualizar | no |
| 3 | button/button | accion | quickExpress | Express 3 horas | no |
| 4 | button/button | accion | quickNight | Amanecida | no |
| 5 | input/hidden | entrada | rateId | rateId | no |
| 6 | input/- | entrada | planName | planName | no |
| 7 | select/- | entrada | planType | Express Day-use Nocturno Amanecida Suite VIP Promocion | no |
| 8 | select/- | entrada | stationSelect | stationSelect | no |
| 9 | input/number | entrada | stationId | stationId | no |
| 10 | input/- | entrada | stationCode | stationCode | no |
| 11 | input/- | entrada | stationName | stationName | no |
| 12 | input/- | entrada | roomCategory | roomCategory | no |
| 13 | select/- | entrada | dayFrom | dayFrom | no |
| 14 | select/- | entrada | dayTo | dayTo | no |
| 15 | input/time | entrada | startTime | 00:00 | no |
| 16 | input/time | entrada | endTime | 23:59 | no |
| 17 | input/number | entrada | includedMinutes | 180 | no |
| 18 | input/number | entrada | baseValue | 0 | no |
| 19 | input/number | entrada | extraMinutes | 60 | no |
| 20 | input/number | entrada | extraValue | 0 | no |
| 21 | input/number | entrada | toleranceMinutes | 10 | no |
| 22 | select/- | entrada | currency | COP USD EUR | no |
| 23 | input/number | entrada | priority | 1 | no |
| 24 | select/- | entrada | status | Activo Inactivo | no |
| 25 | input/checkbox | accion | chargeFraction | chargeFraction | no |
| 26 | input/checkbox | accion | autoApply | autoApply | no |
| 27 | textarea/- | entrada | notes | notes | no |
| 28 | button/submit | accion | - | Guardar plan | no |
| 29 | button/button | accion | applyAllRooms | Aplicar a todas | no |
| 30 | button/button | accion | resetForm | Limpiar | no |
| 31 | button/button | accion | clearFilters | Limpiar filtros | no |
| 32 | input/- | entrada | filterText | filterText | no |
| 33 | select/- | entrada | filterType | Todos Express Day-use Nocturno Amanecida Suite VIP Promocion | no |
| 34 | select/- | entrada | filterStatus | Todos Activo Inactivo | no |
| 35 | select/- | entrada | filterStation | filterStation | no |
| 36 | select/- | entrada | simPlan | simPlan | no |
| 37 | input/number | entrada | simMinutes | 210 | no |
| 38 | button/submit | accion | - | Simular | no |
| 39 | button/button | accion | - | Express 3h | sí |
| 40 | button/button | accion | - | Day-use 6h | sí |
| 41 | button/button | accion | - | Amanecida | sí |
| 42 | button/button | accion | - | VIP 4h | sí |
| 43 | button/button | accion | - | Promo lunes-jueves | sí |

### `web/administrar_empresa/tarifas_por_dia.html` (38)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | addBtn | Agregar tarifa | no |
| 2 | input/hidden | entrada | itemId | itemId | no |
| 3 | input/hidden | entrada | empresa_id | empresa_id | no |
| 4 | select/- | entrada | estacion_selector | Selecciona una estación | no |
| 5 | input/number | entrada | estacion_id | estacion_id | no |
| 6 | input/- | entrada | estacion_codigo | estacion_codigo | no |
| 7 | input/- | entrada | estacion_nombre | estacion_nombre | no |
| 8 | input/- | entrada | nombre_tarifa | nombre_tarifa | no |
| 9 | input/- | entrada | servicio_nombre | hospedaje | no |
| 10 | select/- | entrada | moneda | COP USD EUR | no |
| 11 | input/number | entrada | valor_dia | 0 | no |
| 12 | input/number | entrada | personas_desde | 1 | no |
| 13 | input/number | entrada | personas_hasta | 0 | no |
| 14 | input/time | entrada | hora_check_in | 15:00 | no |
| 15 | input/time | entrada | hora_check_out | 12:00 | no |
| 16 | select/- | entrada | aplicar_automaticamente | Sí No | no |
| 17 | input/number | entrada | prioridad | 1 | no |
| 18 | select/- | entrada | estado | Activo Inactivo | no |
| 19 | textarea/- | entrada | observaciones | observaciones | no |
| 20 | button/submit | accion | saveBtn | Guardar | no |
| 21 | button/button | accion | applyAllBtn | Aplicar a todas las estaciones | no |
| 22 | button/button | accion | cancelBtn | Cancelar | no |
| 23 | select/- | entrada | filtro_estacion_id | Todas | no |
| 24 | select/- | entrada | filtro_include_inactive | No Sí | no |
| 25 | button/button | accion | filtrarBtn | Filtrar | no |
| 26 | button/button | accion | limpiarBtn | Limpiar | no |
| 27 | select/- | entrada | sim_estacion_id | Selecciona una estación | no |
| 28 | input/datetime-local | entrada | sim_activado_en | sim_activado_en | no |
| 29 | input/datetime-local | entrada | sim_fecha_corte | sim_fecha_corte | no |
| 30 | input/number | entrada | sim_personas | 1 | no |
| 31 | button/button | accion | simularBtn | Calcular | no |
| 32 | input/date | entrada | cmp_desde | cmp_desde | no |
| 33 | input/date | entrada | cmp_hasta | cmp_hasta | no |
| 34 | select/- | entrada | cmp_formato | PDF XLS CSV JSON TXT | no |
| 35 | button/button | accion | cmp_descargar_btn | Descargar comparativo | no |
| 36 | button/button | accion | - | Editar | sí |
| 37 | button/button | accion | ' + id + ' | ' + nextLabel + ' | sí |
| 38 | button/button | accion | ' + id + ' | Eliminar | sí |

### `web/administrar_empresa/tarifas_por_minutos.html` (42)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | addBtn | Agregar tarifa | no |
| 2 | input/hidden | entrada | itemId | itemId | no |
| 3 | input/hidden | entrada | empresa_id | empresa_id | no |
| 4 | select/- | entrada | estacion_selector | Selecciona una estación | no |
| 5 | input/number | entrada | estacion_id | estacion_id | no |
| 6 | input/- | entrada | estacion_codigo | estacion_codigo | no |
| 7 | input/- | entrada | estacion_nombre | estacion_nombre | no |
| 8 | select/- | entrada | dia_semana_desde | dia_semana_desde | no |
| 9 | select/- | entrada | dia_semana_hasta | dia_semana_hasta | no |
| 10 | input/number | entrada | minutos_base | 120 | no |
| 11 | input/number | entrada | valor_base | 35000 | no |
| 12 | select/- | entrada | moneda | COP USD EUR | no |
| 13 | input/number | entrada | minutos_extra | 60 | no |
| 14 | input/number | entrada | valor_extra | 20000 | no |
| 15 | input/checkbox | accion | cobrar_por_fraccion | cobrar_por_fraccion | no |
| 16 | input/number | entrada | prioridad | 1 | no |
| 17 | select/- | entrada | estado | Activo Inactivo | no |
| 18 | textarea/- | entrada | observaciones | observaciones | no |
| 19 | button/submit | accion | saveBtn | Guardar | no |
| 20 | button/button | accion | applyAllBtn | Aplicar a todas las estaciones | no |
| 21 | button/button | accion | cancelBtn | Cancelar | no |
| 22 | select/- | entrada | filtro_estacion_id | Todas | no |
| 23 | select/- | entrada | filtro_dia_semana | filtro_dia_semana | no |
| 24 | select/- | entrada | filtro_include_inactive | No Sí | no |
| 25 | button/button | accion | filtrarBtn | Filtrar | no |
| 26 | button/button | accion | limpiarBtn | Limpiar | no |
| 27 | select/- | entrada | cfg_redondeo_modo | Sin redondeo Hacia arriba Hacia abajo Matematico (0.5) | no |
| 28 | input/number | entrada | cfg_redondeo_unidad | 100 | no |
| 29 | input/number | entrada | cfg_monto_minimo_diario | 0 | no |
| 30 | input/number | entrada | cfg_monto_maximo_diario | 0 | no |
| 31 | input/number | entrada | cfg_margen_tolerancia_entrada_minutos | 0 | no |
| 32 | input/checkbox | accion | cfg_sensor_auto_activar_estacion | cfg_sensor_auto_activar_estacion | no |
| 33 | input/checkbox | accion | cfg_margen_desactivacion_habilitado | cfg_margen_desactivacion_habilitado | no |
| 34 | input/number | entrada | cfg_margen_desactivacion_minutos | 0 | no |
| 35 | button/button | accion | guardarConfigBtn | Guardar configuracion | no |
| 36 | select/- | entrada | sim_estacion_id | Selecciona una estación | no |
| 37 | select/- | entrada | sim_dia_semana | sim_dia_semana | no |
| 38 | input/number | entrada | sim_minutos | 120 | no |
| 39 | button/button | accion | simularBtn | Calcular | no |
| 40 | button/button | accion | - | Editar | sí |
| 41 | button/button | accion | ' + id + ' | ' + nextLabel + ' | sí |
| 42 | button/button | accion | ' + id + ' | Eliminar | sí |

### `web/administrar_empresa/tesoreria_presupuesto.html` (36)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | tesRefresh | Actualizar | no |
| 2 | button/button | accion | tesSeed | Cargar demo | no |
| 3 | button/button | accion | tesGenerar | Generar flujo | no |
| 4 | button/button | accion | - | Cuentas | sí |
| 5 | button/button | accion | - | Presupuesto | sí |
| 6 | button/button | accion | - | Flujo | sí |
| 7 | button/button | accion | - | Configuración | sí |
| 8 | input/- | entrada | ctaCodigo | ctaCodigo | no |
| 9 | select/- | entrada | ctaTipo | Banco Caja Pasarela Fiducia | no |
| 10 | input/- | entrada | ctaNombre | ctaNombre | no |
| 11 | input/- | entrada | ctaEntidad | ctaEntidad | no |
| 12 | input/- | entrada | ctaNumero | ctaNumero | no |
| 13 | input/number | entrada | ctaSaldo | 0 | no |
| 14 | input/number | entrada | ctaMinimo | 0 | no |
| 15 | button/submit | accion | - | Guardar cuenta | no |
| 16 | input/- | entrada | presCodigo | presCodigo | no |
| 17 | input/- | entrada | presPeriodo | presPeriodo | no |
| 18 | input/- | entrada | presNombre | presNombre | no |
| 19 | input/number | entrada | presIngresos | 0 | no |
| 20 | input/number | entrada | presEgresos | 0 | no |
| 21 | input/- | entrada | presResp | presResp | no |
| 22 | button/submit | accion | - | Guardar presupuesto | no |
| 23 | select/- | entrada | partPres | partPres | no |
| 24 | select/- | entrada | partTipo | Ingreso Egreso | no |
| 25 | input/- | entrada | partCat | general | no |
| 26 | input/- | entrada | partConcepto | partConcepto | no |
| 27 | input/number | entrada | partValor | 0 | no |
| 28 | input/number | entrada | partEjecutado | 0 | no |
| 29 | button/submit | accion | - | Guardar partida | no |
| 30 | input/- | entrada | cfgNombre | cfgNombre | no |
| 31 | input/- | entrada | cfgMoneda | COP | no |
| 32 | input/- | entrada | cfgPeriodo | cfgPeriodo | no |
| 33 | select/- | entrada | cfgMetodo | Mensual Semanal Trimestral Manual | no |
| 34 | input/checkbox | accion | cfgAlerta | cfgAlerta | no |
| 35 | input/checkbox | accion | cfgAprobar | cfgAprobar | no |
| 36 | button/submit | accion | - | Guardar configuración | no |

### `web/administrar_empresa/turnos_atencion.html` (40)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | openPublicKiosk | Abrir kiosco público | no |
| 2 | a/- | accion | openDisplayScreen | Abrir pantalla TV | no |
| 3 | button/button | accion | - | Configuración Nombre del sistema, tiempos y reglas públicas | sí |
| 4 | button/button | accion | - | Servicios y puestos Catálogo operativo de atención | sí |
| 5 | button/button | accion | - | Emisión y llamado Tickets, siguiente turno y atención | sí |
| 6 | button/button | accion | - | Seguimiento Fila activa y llamados recientes | sí |
| 7 | button/button | accion | - | Pantalla TV Modo sala, pantalla completa y alarmas | sí |
| 8 | input/- | entrada | cfgNombreSistema | cfgNombreSistema | no |
| 9 | input/- | entrada | cfgNombrePantalla | cfgNombrePantalla | no |
| 10 | input/- | entrada | cfgPrefijo | cfgPrefijo | no |
| 11 | input/number | entrada | cfgTiempo | cfgTiempo | no |
| 12 | input/checkbox | accion | cfgEmisionPublica | cfgEmisionPublica | no |
| 13 | input/checkbox | accion | cfgMostrarCompletados | cfgMostrarCompletados | no |
| 14 | button/submit | accion | - | Guardar configuración | no |
| 15 | input/- | entrada | svcCodigo | svcCodigo | no |
| 16 | input/- | entrada | svcNombre | svcNombre | no |
| 17 | input/- | entrada | svcPrefijo | svcPrefijo | no |
| 18 | input/number | entrada | svcPrioridad | 100 | no |
| 19 | input/color | entrada | svcColor | #2563eb | no |
| 20 | input/- | entrada | svcDescripcion | svcDescripcion | no |
| 21 | button/submit | accion | - | Crear servicio | no |
| 22 | input/- | entrada | pstCodigo | pstCodigo | no |
| 23 | input/- | entrada | pstNombre | pstNombre | no |
| 24 | input/- | entrada | pstArea | pstArea | no |
| 25 | input/- | entrada | pstUbicacion | pstUbicacion | no |
| 26 | input/- | entrada | pstServicios | pstServicios | no |
| 27 | button/submit | accion | - | Crear puesto | no |
| 28 | select/- | entrada | emitServicio | emitServicio | no |
| 29 | select/- | entrada | emitPuesto | emitPuesto | no |
| 30 | input/- | entrada | emitNombre | emitNombre | no |
| 31 | input/- | entrada | emitDocumento | emitDocumento | no |
| 32 | button/button | accion | btnEmitirTicket | Emitir ticket | no |
| 33 | button/button | accion | btnLlamarSiguiente | Llamar siguiente | no |
| 34 | button/button | accion | btnPrintLastTicket | Imprimir ticket | no |
| 35 | button/button | accion | btnOpenDisplayFromOps | Ver en TV | no |
| 36 | a/- | accion | openDisplayScreenPanel | Abrir pantalla TV | no |
| 37 | a/- | accion | openPublicKioskPanel | Abrir kiosco publico | no |
| 38 | button/button | accion | btnCopyDisplayUrl | Copiar enlace TV | no |
| 39 | input/- | entrada | displayUrlText | displayUrlText | no |
| 40 | button/button | accion | btnPrintDemoTicket | Vista de impresion | no |

### `web/administrar_empresa/tutorial_domotica.html` (3)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Volver a Domotica | no |
| 2 | a/- | accion | - | Configurar equipos | no |
| 3 | a/- | accion | - | Ver Raspberrys | no |

### `web/administrar_empresa/tutorial_tarifas_hotel.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | backLink | Volver a tarifas | no |

### `web/administrar_empresa/tutorial_tarifas_motel.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | backLink | Volver a tarifas | no |

### `web/administrar_empresa/ubicacion_gps.html` (27)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | reloadBtn | Actualizar | no |
| 2 | input/hidden | entrada | editingDeviceID | editingDeviceID | no |
| 3 | input/- | entrada | deviceCodigo | deviceCodigo | no |
| 4 | input/- | entrada | deviceNombre | deviceNombre | no |
| 5 | input/- | entrada | deviceMarca | deviceMarca | no |
| 6 | input/- | entrada | deviceModelo | deviceModelo | no |
| 7 | select/- | entrada | deviceTipo | Rastreador GPS Telefono movil OBD vehicular IoT / sensor Tablet Otro | no |
| 8 | input/- | entrada | deviceProveedor | deviceProveedor | no |
| 9 | input/- | entrada | deviceHardware | deviceHardware | no |
| 10 | input/- | entrada | deviceSIM | deviceSIM | no |
| 11 | input/- | entrada | devicePlaca | devicePlaca | no |
| 12 | input/- | entrada | deviceReferencia | deviceReferencia | no |
| 13 | input/number | entrada | deviceIntervalo | 10 | no |
| 14 | select/- | entrada | deviceProtocolo | Manual / navegador HTTP Webhook MQTT TCP Traccar Flespi | no |
| 15 | textarea/- | entrada | deviceDescripcion | deviceDescripcion | no |
| 16 | button/submit | accion | saveDeviceBtn | Crear dispositivo | no |
| 17 | button/button | accion | cancelEditDeviceBtn | Cancelar edicion | no |
| 18 | input/- | entrada | deviceSearch | deviceSearch | no |
| 19 | button/button | accion | deviceSearchBtn | Buscar | no |
| 20 | select/- | entrada | activeDeviceSelect | activeDeviceSelect | no |
| 21 | button/button | accion | toggleSelectedTrackingBtn | Iniciar tracking | no |
| 22 | button/button | accion | registerPointBtn | Registrar punto ahora | no |
| 23 | button/button | accion | ' + id + ' | Seleccionar | sí |
| 24 | button/button | accion | ' + id + ' | ' + (tracking ? ('Detener ' + intervalo + 's') : ('Iniciar ' + intervalo + 's')) + ' | sí |
| 25 | button/button | accion | ' + id + ' | Editar | sí |
| 26 | button/button | accion | ' + id + ' | ' + (estado === 'activo' ? 'Desactivar' : 'Activar') + ' | sí |
| 27 | button/button | accion | ' + id + ' | Eliminar | sí |

### `web/administrar_empresa/ultimos_movimientos_de_caja.html` (2)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | backToStationsLink | Regresar a estaciones | no |
| 2 | button/- | accion | reloadBtn | Actualizar | no |

### `web/administrar_empresa/vehiculos_registro.html` (41)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | addBtn | Agregar registro | no |
| 2 | select/- | entrada | cfg_pais_codigo | CO - Colombia MX - Mexico AR - Argentina CL - Chile OT - Otro | no |
| 3 | input/- | entrada | cfg_patente_regex | cfg_patente_regex | no |
| 4 | input/- | entrada | cfg_patente_descripcion | cfg_patente_descripcion | no |
| 5 | input/checkbox | accion | cfg_evitar_duplicado | cfg_evitar_duplicado | no |
| 6 | button/button | accion | cfgGuardarBtn | Guardar configuración | no |
| 7 | button/button | accion | cfgRecargarBtn | Recargar | no |
| 8 | input/hidden | entrada | itemId | itemId | no |
| 9 | input/hidden | entrada | empresa_id | empresa_id | no |
| 10 | input/- | entrada | patente | patente | no |
| 11 | select/- | entrada | tipo_vehiculo | Automóvil Moto Camion Camioneta Bus Van Bicicleta Otro | no |
| 12 | select/- | entrada | estado_registro | En empresa Retirado | no |
| 13 | input/- | entrada | marca | marca | no |
| 14 | input/- | entrada | modelo | modelo | no |
| 15 | input/- | entrada | color | color | no |
| 16 | input/- | entrada | conductor_nombre | conductor_nombre | no |
| 17 | input/- | entrada | conductor_documento | conductor_documento | no |
| 18 | input/- | entrada | referencia_externa | referencia_externa | no |
| 19 | input/- | entrada | propietario_nombre | propietario_nombre | no |
| 20 | input/- | entrada | propietario_documento | propietario_documento | no |
| 21 | input/datetime-local | entrada | fecha_ingreso | fecha_ingreso | no |
| 22 | input/datetime-local | entrada | fecha_salida | fecha_salida | no |
| 23 | input/- | entrada | usuario_salida | usuario_salida | no |
| 24 | input/- | entrada | motivo_ingreso | motivo_ingreso | no |
| 25 | textarea/- | entrada | observaciones | observaciones | no |
| 26 | button/submit | accion | saveBtn | Guardar | no |
| 27 | button/button | accion | cancelBtn | Cancelar | no |
| 28 | input/- | entrada | buscar | buscar | no |
| 29 | input/- | entrada | filtro_patente | filtro_patente | no |
| 30 | select/- | entrada | filtro_estado_registro | Todos En empresa Retirado | no |
| 31 | input/date | entrada | filtro_desde | filtro_desde | no |
| 32 | input/date | entrada | filtro_hasta | filtro_hasta | no |
| 33 | button/button | accion | filtrarBtn | Filtrar | no |
| 34 | button/button | accion | limpiarBtn | Limpiar | no |
| 35 | select/- | entrada | reporteFormato | PDF Excel (XLS) CSV JSON TXT | no |
| 36 | button/button | accion | reporteVerBtn | Ver permanencia | no |
| 37 | button/button | accion | reporteDescargarBtn | Descargar reporte | no |
| 38 | button/button | accion | - | Editar | sí |
| 39 | button/button | accion | ' + id + ' | Marcar salida | sí |
| 40 | button/button | accion | ' + id + ' | ' + nextLabel + ' | sí |
| 41 | button/button | accion | ' + id + ' | Eliminar | sí |

### `web/administrar_empresa/venta_publica.html` (58)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | reloadBtn | Actualizar | no |
| 2 | button/button | accion | openPublicBtn | Abrir perfil público | no |
| 3 | button/button | accion | previewPublicBtn | Ver página como cliente | no |
| 4 | button/button | accion | viewPagesBtn | Ver mis páginas | no |
| 5 | input/- | entrada | storeName | storeName | no |
| 6 | input/- | entrada | storeSlug | storeSlug | no |
| 7 | input/- | entrada | storeDomain | storeDomain | no |
| 8 | input/- | entrada | storeCurrency | COP | no |
| 9 | input/- | entrada | storeLogo | storeLogo | no |
| 10 | input/- | entrada | storeBanner | storeBanner | no |
| 11 | input/color | entrada | storeColor | #0f4c81 | no |
| 12 | select/- | entrada | storeTheme | Corporativo Premium Hospitalidad | no |
| 13 | input/- | entrada | whatsappPublicNumber | whatsappPublicNumber | no |
| 14 | input/checkbox | accion | whatsappFloatActive | whatsappFloatActive | no |
| 15 | textarea/- | entrada | storeDesc | storeDesc | no |
| 16 | input/checkbox | accion | pedidosActivo | pedidosActivo | no |
| 17 | input/checkbox | accion | pedidosRegistroOpcional | pedidosRegistroOpcional | no |
| 18 | input/checkbox | accion | pedidosRecojo | pedidosRecojo | no |
| 19 | input/checkbox | accion | pedidosDomicilio | pedidosDomicilio | no |
| 20 | input/checkbox | accion | pedidosTracking | pedidosTracking | no |
| 21 | input/checkbox | accion | pedidosDespacho | pedidosDespacho | no |
| 22 | input/- | entrada | pedidosNombreSistema | pedidosNombreSistema | no |
| 23 | input/number | entrada | pedidosTiempo | pedidosTiempo | no |
| 24 | input/checkbox | accion | wompiActivo | wompiActivo | no |
| 25 | select/- | entrada | wompiMode | Sandbox Producción | no |
| 26 | input/- | entrada | wompiPublicKey | wompiPublicKey | no |
| 27 | input/- | entrada | wompiPrivateRef | wompiPrivateRef | no |
| 28 | input/- | entrada | wompiIntegrityRef | wompiIntegrityRef | no |
| 29 | input/checkbox | accion | epaycoActivo | epaycoActivo | no |
| 30 | select/- | entrada | epaycoMode | Sandbox Producción | no |
| 31 | input/- | entrada | epaycoPublicKey | epaycoPublicKey | no |
| 32 | input/- | entrada | epaycoPrivateRef | epaycoPrivateRef | no |
| 33 | input/- | entrada | epaycoCustomerId | epaycoCustomerId | no |
| 34 | button/button | accion | saveConfigBtn | Guardar perfil público | no |
| 35 | button/button | accion | openProfilePreviewBtn | Abrir perfil | no |
| 36 | button/button | accion | openSelectedPageBtn | Abrir página seleccionada | no |
| 37 | input/- | entrada | profilePublicLink | profilePublicLink | no |
| 38 | button/button | accion | copyProfileLinkBtn | Copiar | no |
| 39 | input/- | entrada | pageName | pageName | no |
| 40 | input/- | entrada | pageSlug | pageSlug | no |
| 41 | textarea/- | entrada | pageDesc | pageDesc | no |
| 42 | input/- | entrada | pageBanner | pageBanner | no |
| 43 | button/button | accion | savePageBtn | Guardar página | no |
| 44 | button/button | accion | clearPageBtn | Nueva página | no |
| 45 | input/- | entrada | productSearch | productSearch | no |
| 46 | button/button | accion | searchProductsBtn | Buscar | no |
| 47 | button/button | accion | reloadOrdersBtn | Actualizar pedidos | no |
| 48 | button/- | accion | - | Seleccionar | sí |
| 49 | button/- | accion | - | Editar | sí |
| 50 | a/- | accion | - | Ver página | no |
| 51 | button/- | accion | - | Agregar | sí |
| 52 | a/- | accion | - | Ver página | no |
| 53 | button/- | accion | - | Quitar | sí |
| 54 | button/- | accion | - | Preparando | sí |
| 55 | button/- | accion | - | Listo | sí |
| 56 | button/- | accion | - | Mensajero | sí |
| 57 | button/- | accion | - | Entregado | sí |
| 58 | a/- | accion | - | Seguimiento | no |

### `web/administrar_empresa/ventas.html` (30)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | select/- | entrada | facturarClienteSelect | Cargando clientes... | no |
| 2 | select/- | entrada | facturarClienteTipoDocumento | CC NIT CE PAS | no |
| 3 | input/- | entrada | facturarClienteNumeroDocumento | facturarClienteNumeroDocumento | no |
| 4 | input/- | entrada | facturarClienteNombre | facturarClienteNombre | no |
| 5 | input/email | entrada | facturarClienteEmail | facturarClienteEmail | no |
| 6 | input/- | entrada | facturarClienteTelefono | facturarClienteTelefono | no |
| 7 | input/- | entrada | facturarClienteDireccion | facturarClienteDireccion | no |
| 8 | input/checkbox | accion | facturarEnviarCorreo | facturarEnviarCorreo | no |
| 9 | button/button | accion | btnFacturarVentaCancelar | Cancelar | no |
| 10 | button/button | accion | btnFacturarVentaConfirmar | Generar y enviar | no |
| 11 | button/button | accion | btnExportCSV | Exportar CSV | no |
| 12 | input/date | entrada | fechaDesde | fechaDesde | no |
| 13 | input/date | entrada | fechaHasta | fechaHasta | no |
| 14 | input/- | entrada | q | q | no |
| 15 | select/- | entrada | tipoVenta | Todas Factura electrónica Comprobante | no |
| 16 | select/- | entrada | estadoDocumento | Todos Emitida Pendiente emisión Borrador Anulada | no |
| 17 | input/checkbox | accion | includeInactive | includeInactive | no |
| 18 | select/- | entrada | modoVisualizacion | Carta completa Compacta ejecutiva Ticket POS 80mm | no |
| 19 | button/submit | accion | btnBuscar | Buscar | no |
| 20 | button/button | accion | btnLimpiar | Limpiar | no |
| 21 | button/button | accion | - | Abrir factura FE | sí |
| 22 | button/button | accion | - | Reenviar DIAN | sí |
| 23 | button/button | accion | - | Hacer factura electrónica | sí |
| 24 | button/button | accion | - | Reenviar DIAN | sí |
| 25 | button/button | accion | - | Anular venta | sí |
| 26 | button/button | accion | - | Ver / imprimir | sí |
| 27 | button/button | accion | - | Correo | sí |
| 28 | button/button | accion | - | WhatsApp | sí |
| 29 | button/button | accion | - | Imprimir ahora | sí |
| 30 | button/button | accion | - | Cerrar | sí |

### `web/administrar_empresa/youtube_station_browser.html` (4)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/text | entrada | youtubeSearchInput | youtubeSearchInput | no |
| 2 | button/button | accion | youtubeSearchBtn | Cargar | no |
| 3 | button/button | accion | youtubeStationHomeBtn | Inicio | no |
| 4 | a/- | accion | youtubeStationBrowserLink | Abrir YouTube | no |

### `web/ayuda/ayuda.html` (2)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/search | entrada | helpSearch | helpSearch | no |
| 2 | button/button | accion | helpClear | Limpiar | no |

### `web/ayuda/ayuda_apis.html` (2)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Volver al centro de ayuda | no |
| 2 | a/- | accion | - | Seleccionar empresa | no |

### `web/ayuda/ayuda_contextual.html` (2)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | backToOrigin | Volver a la pantalla anterior | no |
| 2 | a/- | accion | openFullHelp | Abrir centro de ayuda | no |

### `web/ayuda/login_administradores.html` (2)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Volver al login | no |
| 2 | a/- | accion | - | Ir al registro | no |

### `web/ayuda/tutorial_nomina.html` (8)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Volver a ayuda | no |
| 2 | a/- | accion | openPayrollModule | Abrir modulo de nomina | no |
| 3 | button/button | accion | - | Reproducir | sí |
| 4 | button/button | accion | - | Reproducir | sí |
| 5 | button/button | accion | - | Reproducir | sí |
| 6 | button/button | accion | - | Reproducir | sí |
| 7 | button/button | accion | - | Reproducir | sí |
| 8 | button/button | accion | - | Reproducir | sí |

### `web/calculadora.html` (19)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | - | C | sí |
| 2 | button/button | accion | - | DEL | sí |
| 3 | button/button | accion | - | % | sí |
| 4 | button/button | accion | - | / | sí |
| 5 | button/button | accion | - | 7 | sí |
| 6 | button/button | accion | - | 8 | sí |
| 7 | button/button | accion | - | 9 | sí |
| 8 | button/button | accion | - | x | sí |
| 9 | button/button | accion | - | 4 | sí |
| 10 | button/button | accion | - | 5 | sí |
| 11 | button/button | accion | - | 6 | sí |
| 12 | button/button | accion | - | - | sí |
| 13 | button/button | accion | - | 1 | sí |
| 14 | button/button | accion | - | 2 | sí |
| 15 | button/button | accion | - | 3 | sí |
| 16 | button/button | accion | - | + | sí |
| 17 | button/button | accion | - | 0 | sí |
| 18 | button/button | accion | - | . | sí |
| 19 | button/button | accion | - | = | sí |

### `web/configuracion_de_la_cuenta.html` (8)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/email | entrada | fldEmail | fldEmail | no |
| 2 | input/text | entrada | fldName | fldName | no |
| 3 | input/text | entrada | fldPhone | fldPhone | no |
| 4 | button/- | accion | saveProfileBtn | Guardar perfil | no |
| 5 | input/password | entrada | fldCurrentPwd | fldCurrentPwd | no |
| 6 | input/password | entrada | fldNewPwd | fldNewPwd | no |
| 7 | input/password | entrada | fldNewPwdConfirm | fldNewPwdConfirm | no |
| 8 | button/- | accion | changePwdBtn | Cambiar contraseña | no |

### `web/contrato.html` (2)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Volver al login | no |
| 2 | a/- | accion | currentVersionLink | Ver version actual | no |

### `web/descargar_informacion_de_la_empresa.html` (3)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | select/- | entrada | empresaExportFormat | Backup completo (.json) Excel empresarial (.xls) PDF ejecutivo (.pdf) CSV para otros sistemas (.csv) JSON completo (.jso | no |
| 2 | button/submit | accion | empresaExportDownload | &#8595; Descargar | no |
| 3 | button/button | accion | empresaExportBack | Regresar a seleccionar empresas | no |

### `web/descripcion_de_los_sistemas.html` (2)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Volver al portal | no |
| 2 | a/- | accion | - | Iniciar sesion | no |

### `web/domicilios.html` (11)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/- | accion | locBtn | Usar mi ubicación | no |
| 2 | input/- | entrada | name | name | no |
| 3 | input/- | entrada | phone | phone | no |
| 4 | input/- | entrada | address | address | no |
| 5 | select/- | entrada | pay | Efectivo Transferencia Tarjeta al recibir | no |
| 6 | textarea/- | entrada | notes | notes | no |
| 7 | button/submit | accion | - | Confirmar pedido | no |
| 8 | button/- | accion | - | ${esc(r.nombre)} ${esc(r.categoria \|\| "Restaurante")} \| ${esc(r.direccion \|\| "")} | sí |
| 9 | button/- | accion | - | Agregar | sí |
| 10 | button/button | accion | - | Quitar | sí |
| 11 | button/- | accion | - | Agregar | sí |

### `web/domicilios_domiciliario.html` (11)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/- | entrada | doc | doc | no |
| 2 | input/- | entrada | pin | pin | no |
| 3 | button/- | accion | - | Entrar | no |
| 4 | button/- | accion | online | Online disponible | no |
| 5 | button/- | accion | busy | Ocupado | no |
| 6 | button/- | accion | refresh | Actualizar | no |
| 7 | button/- | accion | - | Aceptar | sí |
| 8 | button/- | accion | - | Rechazar | sí |
| 9 | button/- | accion | ${o.id} | Recogido | sí |
| 10 | button/- | accion | ${o.id} | En camino | sí |
| 11 | button/- | accion | - | Entregado | sí |

### `web/domicilios_restaurante.html` (7)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/- | entrada | code | code | no |
| 2 | input/- | entrada | pin | pin | no |
| 3 | button/- | accion | - | Entrar | no |
| 4 | button/- | accion | refresh | Actualizar | no |
| 5 | button/- | accion | ${o.id} | Confirmar | sí |
| 6 | button/- | accion | ${o.id} | Preparando | sí |
| 7 | button/- | accion | ${o.id} | Pedido listo | sí |

### `web/editar_empresa.html` (14)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Volver a empresas | no |
| 2 | input/- | entrada | empresaNombre | empresaNombre | no |
| 3 | input/- | entrada | empresaTipo | empresaTipo | no |
| 4 | textarea/- | entrada | empresaObservaciones | empresaObservaciones | no |
| 5 | button/submit | accion | - | Guardar cambios | no |
| 6 | input/email | entrada | empresaShareEmail | empresaShareEmail | no |
| 7 | select/- | entrada | empresaShareNivel | Solo ver Acceso total Solo ciertos modulos | no |
| 8 | textarea/- | entrada | empresaShareMessage | empresaShareMessage | no |
| 9 | button/submit | accion | - | Enviar invitación | no |
| 10 | button/button | accion | empresaDownloadBeforeDeleteBtn | Descargar información antes de eliminar | no |
| 11 | input/- | entrada | empresaDeleteConfirm | empresaDeleteConfirm | no |
| 12 | input/- | entrada | empresaDeletePhrase | empresaDeletePhrase | no |
| 13 | input/checkbox | accion | empresaDeleteAcknowledge | empresaDeleteAcknowledge | no |
| 14 | button/button | accion | empresaDeleteBtn | Eliminar empresa definitivamente | no |

### `web/elegir_licencia.html` (2)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/text | entrada | asesor_licencia_' + escapeHtml(String(licencia.id \|\| idx)) + ' | ' + escapeHtml(initialAsesorID \|\| '') + ' | no |
| 2 | button/button | accion | ' + escapeHtml(String(empresaId \|\| '')) + ' | ' + escapeHtml(actionLabel) + ' | sí |

### `web/epayco/pago_exitoso.html` (2)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Ir a seleccionar empresa | no |
| 2 | a/- | accion | epaycoSuccessPaymentLink | Ver detalle del pago | no |

### `web/epayco/respuesta.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | epaycoContinueLink | Continuar a validar mi pago | no |

### `web/index.html` (7)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Iniciar sesi&oacute;n | no |
| 2 | a/- | accion | - | Crear cuenta | no |
| 3 | button/button | accion | headerMoreMenuToggle | M&aacute;s ▾ | no |
| 4 | a/- | accion | - | WhatsApp principal | no |
| 5 | a/- | accion | - | Enviar correo | no |
| 6 | button/button | accion | portalCarouselPrev | &lsaquo; | no |
| 7 | button/button | accion | portalCarouselNext | &rsaquo; | no |

### `web/login.html` (17)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/email | entrada | adminEmail | adminEmail | no |
| 2 | input/password | entrada | adminPassword | adminPassword | no |
| 3 | button/button | accion | - | Mostrar contraseña | sí |
| 4 | input/text | entrada | adminOtpCode | adminOtpCode | no |
| 5 | input/checkbox | accion | rememberAdminEmailCheckbox | rememberAdminEmailCheckbox | no |
| 6 | button/submit | accion | emailLoginBtn | Iniciar por correo | no |
| 7 | a/- | accion | forgotLink | ¿Olvidó su contraseña? | no |
| 8 | input/email | entrada | forgotEmail | forgotEmail | no |
| 9 | button/submit | accion | forgotPasswordBtn | Enviar recuperación | no |
| 10 | button/button | accion | backToLoginLink | Volver al login | no |
| 11 | input/email | entrada | resetEmail | resetEmail | no |
| 12 | input/password | entrada | resetPassword | resetPassword | no |
| 13 | input/password | entrada | resetPasswordConfirm | resetPasswordConfirm | no |
| 14 | button/submit | accion | resetPasswordBtn | Restablecer contraseña | no |
| 15 | button/button | accion | backFromResetLink | Volver al login | no |
| 16 | button/button | accion | installPwaBtn | &#8595; Instalar app | no |
| 17 | a/- | accion | - | &#8962; Ir al inicio | no |

### `web/login_usuario.html` (37)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/hidden | entrada | empresaID | empresaID | no |
| 2 | input/email | entrada | email | email | no |
| 3 | input/password | entrada | password | password | no |
| 4 | button/button | accion | - | Mostrar contraseña | sí |
| 5 | button/submit | accion | btnIngresar | Ingresar | no |
| 6 | a/- | accion | linkGoRecovery | ¿Olvidó su contraseña? | no |
| 7 | a/- | accion | linkGoInviteRecovery | Recuperar email de invitación | no |
| 8 | input/hidden | entrada | setupInvitationToken | setupInvitationToken | no |
| 9 | input/email | entrada | setupEmail | setupEmail | no |
| 10 | input/password | entrada | setupDocumento | setupDocumento | no |
| 11 | input/password | entrada | setupPassword | setupPassword | no |
| 12 | input/password | entrada | setupPasswordConfirm | setupPasswordConfirm | no |
| 13 | input/checkbox | accion | contractAcceptCheckbox | contractAcceptCheckbox | no |
| 14 | a/- | accion | contractLink | Leer contrato completo | no |
| 15 | button/submit | accion | btnCrearPassword | Crear contraseña y entrar | no |
| 16 | button/button | accion | btnVolverLogin | Volver | no |
| 17 | input/email | entrada | recoveryEmail | recoveryEmail | no |
| 18 | button/submit | accion | btnSolicitarRecuperacion | Solicitar recuperación | no |
| 19 | button/button | accion | btnIrAReset | Ya tengo token | no |
| 20 | button/button | accion | btnVolverDesdeRecuperacion | Volver | no |
| 21 | input/email | entrada | inviteRecoveryEmail | inviteRecoveryEmail | no |
| 22 | button/submit | accion | btnRecuperarInvitacion | Reenviar invitación | no |
| 23 | button/button | accion | btnVolverDesdeInvitacion | Volver | no |
| 24 | input/email | entrada | resetEmail | resetEmail | no |
| 25 | input/text | entrada | resetToken | resetToken | no |
| 26 | input/password | entrada | resetPassword | resetPassword | no |
| 27 | input/password | entrada | resetPasswordConfirm | resetPasswordConfirm | no |
| 28 | button/submit | accion | btnRestablecerPassword | Restablecer y entrar | no |
| 29 | button/button | accion | btnVolverDesdeReset | Volver | no |
| 30 | input/email | entrada | changeEmail | changeEmail | no |
| 31 | input/password | entrada | changeCurrentPassword | changeCurrentPassword | no |
| 32 | input/password | entrada | changeNewPassword | changeNewPassword | no |
| 33 | input/password | entrada | changeNewPasswordConfirm | changeNewPasswordConfirm | no |
| 34 | button/submit | accion | btnCambiarPassword | Cambiar y entrar | no |
| 35 | button/button | accion | btnVolverDesdeCambio | Volver | no |
| 36 | button/button | accion | installPwaBtn | &#8595; Instalar app | no |
| 37 | button/button | accion | contractDialogClose | Cerrar | no |

### `web/mis_clientes.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | reloadBtn | Actualizar | no |

### `web/pagar_licencia.html` (12)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | selectEpaycoCard | '+ ' ' + escapeHtml(epaycoBadge) + ' '+ ' '+ ' '+ ' '+ ' Grupo Davivienda '+ ' '+ ' Epayco '+ ' Tarjeta, PSE y otros '+  | no |
| 2 | button/button | accion | selectWompiCard | '+ ' ' + escapeHtml(wompiBadge) + ' '+ ' '+ ' '+ ' Grupo Bancolombia '+ ' '+ ' Wompi '+ ' Nequi, PSE y tarjetas '+ ' | no |
| 3 | select/- | entrada | checkoutCountrySelect | '+ countryOptions.map(function(code){ return ' ' + escapeHtml(countryName(code)) + ' (' + escapeHtml(code) + ') '; }).jo | no |
| 4 | input/text | entrada | discountCode | ' + escapeHtml(initialDiscountCode) + ' | no |
| 5 | input/number | entrada | checkoutQuantity | ' + escapeHtml(String(checkoutQuantity)) + ' | no |
| 6 | input/text | entrada | asesorId | ' + escapeHtml(initialAsesorID) + ' | no |
| 7 | input/email | entrada | customerEmail | ' + escapeHtml(initialCustomerEmail) + ' | no |
| 8 | button/button | accion | activateFreeBtn | Activar licencia | no |
| 9 | input/checkbox | accion | epaycoTermsChk | epaycoTermsChk | no |
| 10 | button/button | accion | payEpaycoBtn | Pagar con Epayco | no |
| 11 | input/checkbox | accion | wompiTermsChk | wompiTermsChk | no |
| 12 | button/button | accion | payWompiBtn | Pagar con Wompi | no |

### `web/pagar_productos_de_venta_publica.html` (11)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | backLink | Volver a la página | no |
| 2 | a/- | accion | privateMsgLink | Enviar mensaje privado a la empresa | no |
| 3 | input/- | entrada | buyerName | buyerName | no |
| 4 | input/- | entrada | buyerEmail | buyerEmail | no |
| 5 | input/- | entrada | buyerPhone | buyerPhone | no |
| 6 | input/number | entrada | qty | 1 | no |
| 7 | input/checkbox | accion | terms | terms | no |
| 8 | button/button | accion | payBtn | Pagar | no |
| 9 | input/radio | accion | - | '+esc(x)+' | no |
| 10 | input/number | entrada | - | '+esc(it.qty\|\|1)+' | sí |
| 11 | button/button | accion | - | Quitar | sí |

### `web/pantalla_turnos.html` (2)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | enableSoundBtn | Activar sonido | no |
| 2 | button/button | accion | fullScreenBtn | Pantalla completa | no |

### `web/perfil_red_social.html` (8)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Ver feed público | no |
| 2 | button/button | accion | - | 👍 Me gusta | sí |
| 3 | button/button | accion | - | ❤️ Me encanta | sí |
| 4 | button/button | accion | - | 💬 Comentarios (' + String(comentariosTotal) + ') | sí |
| 5 | input/text | entrada | - | nombre | no |
| 6 | textarea/- | entrada | - | contenido | no |
| 7 | button/submit | accion | - | Comentar | no |
| 8 | button/button | accion | ' + escapeHtml(p.empresa_id) + ' | Mensaje privado | sí |

### `web/productos_estacion_clientes_publico.html` (8)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Volver al portal | no |
| 2 | a/- | accion | - | Ver tienda pública | no |
| 3 | input/- | entrada | vipCodigo | vipCodigo | no |
| 4 | button/button | accion | btnIngresar | Ingresar | no |
| 5 | button/button | accion | btnQR | Ver QR | no |
| 6 | input/- | entrada | pedidoNota | pedidoNota | no |
| 7 | input/- | entrada | publicLink | publicLink | no |
| 8 | button/button | accion | - | Pedir | no |

### `web/red_social_comercial.html` (11)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | socialNotificationsBtn | N 0 | no |
| 2 | a/- | accion | socialSessionBtn | Iniciar sesión | no |
| 3 | button/button | accion | - | 👍 Me gusta | sí |
| 4 | button/button | accion | - | ❤️ Me encanta | sí |
| 5 | button/button | accion | - | 💬 Comentarios (${comentariosTotal}) | sí |
| 6 | input/text | entrada | - | nombre | no |
| 7 | textarea/- | entrada | - | contenido | no |
| 8 | button/submit | accion | - | Comentar | no |
| 9 | button/button | accion | - | Seguir | sí |
| 10 | button/button | accion | ${escapeHtml(p.empresa_id)} | Mensaje privado | sí |
| 11 | button/button | accion | markFollowSeenBtn | Marcar vistas | no |

### `web/registrar_contrasena_usuario_de_google.html` (7)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Saber más | no |
| 2 | input/email | entrada | googleAccountEmail | googleAccountEmail | no |
| 3 | input/password | entrada | googlePassword | googlePassword | no |
| 4 | button/button | accion | - | Mostrar contraseña | sí |
| 5 | input/password | entrada | googlePasswordConfirm | googlePasswordConfirm | no |
| 6 | button/button | accion | - | Mostrar contraseña | sí |
| 7 | button/submit | accion | googlePasswordSetupBtn | Guardar contraseña | no |

### `web/registrar_nuevo_usuario_administrador.html` (12)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Saber más | no |
| 2 | a/- | accion | - | Ir al inicio | no |
| 3 | input/email | entrada | registerEmail | registerEmail | no |
| 4 | input/text | entrada | registerName | registerName | no |
| 5 | input/tel | entrada | registerPhone | registerPhone | no |
| 6 | select/- | entrada | registerCountry | registerCountry | no |
| 7 | select/- | entrada | registerCity | registerCity | no |
| 8 | input/text | entrada | registerCityCustom | registerCityCustom | no |
| 9 | input/password | entrada | registerPassword | registerPassword | no |
| 10 | input/password | entrada | registerPasswordConfirm | registerPasswordConfirm | no |
| 11 | button/submit | accion | adminRegisterBtn | Registrar cuenta | no |
| 12 | a/- | accion | - | Volver al login | no |

### `web/seleccionar_empresa.html` (27)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | linkAgregarEmpresa | Inicio | no |
| 2 | button/button | accion | - | ☰ Ocultar menú | sí |
| 3 | button/- | accion | addBtn | Agregar Empresa | no |
| 4 | input/hidden | entrada | itemId | itemId | no |
| 5 | select/- | entrada | tipo_id | tipo_id | no |
| 6 | input/- | entrada | nombre | nombre | no |
| 7 | input/- | entrada | nit | nit | no |
| 8 | input/- | entrada | observaciones | observaciones | no |
| 9 | button/submit | accion | saveBtn | Guardar | no |
| 10 | button/button | accion | cancelBtn | Cancelar | no |
| 11 | button/button | accion | openAIDrawer | Asistente IA | sí |
| 12 | button/button | accion | aiChatMinibarExpand | Abrir asistente IA | no |
| 13 | button/button | accion | aiChatHintToggle | Ver ejemplos | no |
| 14 | button/button | accion | aiChatConfigBtn | Configurar chat flotante | no |
| 15 | button/button | accion | aiChatMinimize | Minimizar chat | no |
| 16 | button/button | accion | closeAIDrawer | × | no |
| 17 | button/button | accion | aiChatNewBtn | Nuevo chat | no |
| 18 | button/button | accion | aiChatConvBtn | Modo conversación | no |
| 19 | button/button | accion | aiChatMicBtn | Dictar mensaje | no |
| 20 | button/button | accion | aiChatVoiceBtn | Voz del asistente | no |
| 21 | button/button | accion | aiChatStopBtn | Detener audio y respuesta | no |
| 22 | input/hidden | entrada | aiChatMode | operativo | no |
| 23 | input/file | accion | aiChatAttachment | Adjuntar archivo para IA | no |
| 24 | button/button | accion | aiChatAttachBtn | Adjuntar archivo | no |
| 25 | button/button | accion | aiChatClearAttachment | × | no |
| 26 | textarea/- | entrada | aiChatInput | Mensaje al asistente IA | no |
| 27 | button/submit | accion | - | Enviar | no |

### `web/soporte_remoto_acceso.html` (5)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Descargar | no |
| 2 | a/- | accion | - | Descargar | no |
| 3 | a/- | accion | - | Descargar | no |
| 4 | a/- | accion | - | Guía oficial | no |
| 5 | a/- | accion | - | Guía oficial | no |

### `web/super/administradores.html` (8)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/- | accion | addBtn | Invitar administrador | no |
| 2 | input/hidden | entrada | itemId | itemId | no |
| 3 | input/- | entrada | email | email | no |
| 4 | input/- | entrada | name | name | no |
| 5 | select/- | entrada | role | administrador super_administrador | no |
| 6 | button/submit | accion | saveBtn | Enviar invitación | no |
| 7 | button/button | accion | cancelBtn | Cancelar | no |
| 8 | button/- | accion | ${escapeHTML(i.id)} | Eliminar | sí |

### `web/super/administradores_frecuencia_fe.html` (5)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/email | entrada | emailNuevo | emailNuevo | no |
| 2 | button/button | accion | btnAdd | Agregar | no |
| 3 | button/button | accion | btnGuardar | Guardar | no |
| 4 | button/button | accion | btnRecargar | Recargar | no |
| 5 | button/button | accion | - | Eliminar | sí |

### `web/super/administrar_base_de_datos.html` (11)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/- | accion | refreshBtn | Actualizar | no |
| 2 | button/- | accion | toggleAutoBtn | Auto refresco: ON | no |
| 3 | button/- | accion | exportBtn | Exportar JSON | no |
| 4 | button/- | accion | loadEmpresasStorageBtn | Cargar Empresas | no |
| 5 | input/number | entrada | outboxEmpresaID | outboxEmpresaID | no |
| 6 | input/text | entrada | outboxTopic | cuentas_por_pagar.pago_registrado | no |
| 7 | button/button | accion | previewOutboxBtn | Vista previa | no |
| 8 | textarea/- | entrada | outboxRecoveryReason | outboxRecoveryReason | no |
| 9 | input/text | entrada | outboxRecoveryConfirmation | outboxRecoveryConfirmation | no |
| 10 | button/button | accion | executeOutboxRecoveryBtn | Reactivar seleccionados | no |
| 11 | input/checkbox | accion | ' + Number(item.id) + ' | Seleccionar evento ' + Number(item.id) + ' | sí |

### `web/super/administrar_disco_vps.html` (5)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | refreshBtn | Actualizar | no |
| 2 | input/- | entrada | confirmation | confirmation | no |
| 3 | button/button | accion | selectAllBtn | Seleccionar todos | no |
| 4 | button/button | accion | cleanupBtn | Liberar espacio seleccionado | no |
| 5 | input/checkbox | accion | - | ' + escapeHTML(item.id) + ' | no |

### `web/super/agentes_de_mantenimiento_qutomatico.html` (10)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | refreshBtn | Actualizar | no |
| 2 | input/checkbox | accion | enabledInput | enabledInput | no |
| 3 | input/time | entrada | hourInput | 07:00 | no |
| 4 | input/email | entrada | emailInput | emailInput | no |
| 5 | button/submit | accion | saveBtn | Guardar | no |
| 6 | button/button | accion | runBtn | Ejecutar ahora | no |
| 7 | input/number | entrada | limitSeconds | 120 | no |
| 8 | input/number | entrada | limitAdvanced | 5 | no |
| 9 | input/number | entrada | limitLight | 20 | no |
| 10 | button/submit | accion | saveLimitsBtn | Guardar limites | no |

### `web/super/alertas_sistema.html` (21)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnReload | Actualizar | no |
| 2 | button/button | accion | btnEvaluate | Evaluar ahora | no |
| 3 | button/button | accion | btnTest | Probar correo | no |
| 4 | button/button | accion | - | Configuracion | sí |
| 5 | button/button | accion | - | Estado actual | sí |
| 6 | button/button | accion | - | Historial | sí |
| 7 | input/email | entrada | recipientEmail | recipientEmail | no |
| 8 | input/number | entrada | cooldownMinutes | cooldownMinutes | no |
| 9 | input/checkbox | accion | enabled | enabled | no |
| 10 | input/checkbox | accion | diskEnabled | diskEnabled | no |
| 11 | input/number | entrada | diskThreshold | diskThreshold | no |
| 12 | input/checkbox | accion | trafficEnabled | trafficEnabled | no |
| 13 | input/number | entrada | trafficThresholdPct | trafficThresholdPct | no |
| 14 | input/number | entrada | trafficThresholdGB | trafficThresholdGB | no |
| 15 | input/checkbox | accion | sessionsEnabled | sessionsEnabled | no |
| 16 | input/number | entrada | sessionsThreshold | sessionsThreshold | no |
| 17 | input/checkbox | accion | dbConnectionsEnabled | dbConnectionsEnabled | no |
| 18 | input/number | entrada | dbConnectionsThreshold | dbConnectionsThreshold | no |
| 19 | input/checkbox | accion | adminRegisterEnabled | adminRegisterEnabled | no |
| 20 | input/checkbox | accion | empresaNuevaEnabled | empresaNuevaEnabled | no |
| 21 | button/button | accion | btnSave | Guardar configuracion | no |

### `web/super/asesor_comercial.html` (29)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | reloadBtn | Actualizar | no |
| 2 | input/checkbox | accion | advisorPromoEnabled | advisorPromoEnabled | no |
| 3 | select/- | entrada | advisorPromoPercent | 5% 10% 15% 20% 25% 30% | no |
| 4 | button/submit | accion | - | Guardar promocion | no |
| 5 | input/hidden | entrada | advisorId | advisorId | no |
| 6 | input/email | entrada | email | email | no |
| 7 | input/number | entrada | pctPrimerAnio | 40 | no |
| 8 | input/number | entrada | pctRenovacion | 30 | no |
| 9 | input/number | entrada | mesesRenovacion | 24 | no |
| 10 | input/number | entrada | meses | 36 | no |
| 11 | select/- | entrada | metodoPago | Transferencia bancaria Nequi Daviplata Billetera digital Efectivo Otro | no |
| 12 | input/- | entrada | entidadFinanciera | entidadFinanciera | no |
| 13 | select/- | entrada | tipoCuenta | Sin definir Ahorros Corriente Nequi Daviplata Billetera digital Otro | no |
| 14 | input/- | entrada | numeroCuenta | numeroCuenta | no |
| 15 | input/- | entrada | titularCuenta | titularCuenta | no |
| 16 | input/- | entrada | documentoTitular | documentoTitular | no |
| 17 | input/email | entrada | emailPagos | emailPagos | no |
| 18 | input/- | entrada | telefonoPagos | telefonoPagos | no |
| 19 | select/- | entrada | periodicidadPago | Mensual Quincenal Semanal Bajo solicitud | no |
| 20 | input/number | entrada | diaPago | 30 | no |
| 21 | input/number | entrada | pagoMinimo | 0 | no |
| 22 | input/checkbox | accion | requiereSoportePago | requiereSoportePago | no |
| 23 | textarea/- | entrada | obs | obs | no |
| 24 | button/submit | accion | saveBtn | Enviar invitación | no |
| 25 | button/button | accion | cancelEditBtn | Cancelar edicion | no |
| 26 | button/button | accion | reloadCommissionsBtn | Refrescar pagos | no |
| 27 | button/button | accion | - | Editar configuración | sí |
| 28 | button/button | accion | - | Desactivar | sí |
| 29 | button/button | accion | - | Gestionar pago | sí |

### `web/super/auditoria_global.html` (15)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnAgregarAdmin | Agregar administrador | no |
| 2 | button/button | accion | btnActualizar | Actualizar | no |
| 3 | input/date | entrada | desde | desde | no |
| 4 | input/date | entrada | hasta | hasta | no |
| 5 | select/- | entrada | modulo | Todos Empresas Interacciones UI Administradores Empresas compartidas Licencias Reportes globales Tipos de empresas | no |
| 6 | select/- | entrada | resultado | Todos OK Error Rechazado | no |
| 7 | input/- | entrada | usuario | usuario | no |
| 8 | input/- | entrada | empresaId | empresaId | no |
| 9 | input/- | entrada | search | search | no |
| 10 | button/submit | accion | - | Ver auditoría | no |
| 11 | button/button | accion | btnCSV | Exportar CSV | no |
| 12 | button/button | accion | btnJSON | Exportar JSON | no |
| 13 | button/button | accion | prevPage | Anterior | no |
| 14 | button/button | accion | nextPage | Siguiente | no |
| 15 | button/button | accion | cerrarDetalle | Cerrar | no |

### `web/super/auditoria_super_admin.html` (14)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnActualizar | Actualizar | no |
| 2 | input/date | entrada | desde | desde | no |
| 3 | input/date | entrada | hasta | hasta | no |
| 4 | select/- | entrada | modulo | Todos Panel super UI Gmail SMTP Wompi / Nequi Epayco reCAPTCHA IA global OnlyOffice RustDesk VPS Voz IA Respaldo Segurid | no |
| 5 | select/- | entrada | resultado | Todos OK Error Rechazado | no |
| 6 | input/- | entrada | usuario | usuario | no |
| 7 | input/- | entrada | empresaId | empresaId | no |
| 8 | input/- | entrada | search | search | no |
| 9 | button/submit | accion | - | Ver auditoría | no |
| 10 | button/button | accion | btnCSV | Exportar CSV | no |
| 11 | button/button | accion | btnJSON | Exportar JSON | no |
| 12 | button/button | accion | prevPage | Anterior | no |
| 13 | button/button | accion | nextPage | Siguiente | no |
| 14 | button/button | accion | cerrarDetalle | Cerrar | no |

### `web/super/chat_con_ia_global.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | openCentralAI | Abrir asistente IA central | no |

### `web/super/configuracion/alertas_licencia.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Ver todas | no |

### `web/super/configuracion/consumos.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Ver todas | no |

### `web/super/configuracion/epayco.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Ver todas | no |

### `web/super/configuracion/gmail_smtp.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Ver todas | no |

### `web/super/configuracion/ia_global.html` (5)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Proveedor | no |
| 2 | a/- | accion | - | Reglas | no |
| 3 | a/- | accion | - | Contexto | no |
| 4 | a/- | accion | - | Chat | no |
| 5 | a/- | accion | - | Voz | no |

### `web/super/configuracion/limitaciones.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Ver todas | no |

### `web/super/configuracion/login_2fa.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Ver todas | no |

### `web/super/configuracion/nextcloud.html` (13)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/checkbox | accion | nextcloudEnabled | nextcloudEnabled | no |
| 2 | input/url | entrada | nextcloudBaseURL | nextcloudBaseURL | no |
| 3 | input/text | entrada | nextcloudAdminUser | nextcloudAdminUser | no |
| 4 | input/number | entrada | nextcloudQuota | 1024 | no |
| 5 | input/password | entrada | nextcloudAdminSecret | nextcloudAdminSecret | no |
| 6 | button/button | accion | nextcloudSave | Guardar configuracion | no |
| 7 | button/button | accion | nextcloudTest | Probar conexion | no |
| 8 | input/text | entrada | nextcloudAccountUser | nextcloudAccountUser | no |
| 9 | input/text | entrada | nextcloudAccountName | nextcloudAccountName | no |
| 10 | input/number | entrada | nextcloudAccountQuota | 5120 | no |
| 11 | button/button | accion | nextcloudCreateAccount | Crear cuenta personal | no |
| 12 | a/- | accion | nextcloudOpenService | Abrir Nextcloud | no |
| 13 | button/button | accion | nextcloudCopyAccountPassword | Copiar contraseña | no |

### `web/super/configuracion/onlyoffice.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Ver todas | no |

### `web/super/configuracion/recaptcha.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Ver todas | no |

### `web/super/configuracion/respaldo.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Ver todas | no |

### `web/super/configuracion/rustdesk_vps.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Ver todas | no |

### `web/super/configuracion/voz_ia.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Ver todas | no |

### `web/super/configuracion/whatsapp_notificaciones.html` (12)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/checkbox | accion | enabled | enabled | no |
| 2 | input/checkbox | accion | testMode | testMode | no |
| 3 | select/- | entrada | provider | Meta Cloud API | no |
| 4 | input/- | entrada | apiVersion | apiVersion | no |
| 5 | input/- | entrada | phoneNumberId | phoneNumberId | no |
| 6 | input/password | entrada | accessToken | accessToken | no |
| 7 | button/button | accion | saveBtn | Guardar configuración | no |
| 8 | input/- | entrada | testPhone | testPhone | no |
| 9 | textarea/- | entrada | testMessage | Prueba de WhatsApp desde Powerful Control System. | no |
| 10 | button/button | accion | testBtn | Enviar prueba | no |
| 11 | input/checkbox | accion | - | sin etiqueta | sí |
| 12 | input/checkbox | accion | - | sin etiqueta | sí |

### `web/super/configuracion/whatsapp_portal.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Ver todas | no |

### `web/super/configuracion/wompi_nequi.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Ver todas | no |

### `web/super/configuracion_avanzada.html` (171)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/search | entrada | advancedSectionSearch | advancedSectionSearch | no |
| 2 | button/button | accion | advancedSectionClear | Limpiar | no |
| 3 | button/button | accion | - | Consumos Costos externos | sí |
| 4 | button/button | accion | - | RustDesk VPS Soporte remoto | sí |
| 5 | button/button | accion | - | Limitaciones API, DB, cuotas | sí |
| 6 | button/button | accion | - | Almacenamiento Archivos por empresa | sí |
| 7 | button/button | accion | - | OnlyOffice Documentos | sí |
| 8 | button/button | accion | - | URLs visibles Barra del navegador | sí |
| 9 | button/button | accion | - | Voz IA Streaming | sí |
| 10 | button/button | accion | - | Epayco Checkout | sí |
| 11 | button/button | accion | - | Wompi / Nequi Pagos | sí |
| 12 | button/button | accion | - | Alertas licencia Vencimiento | sí |
| 13 | button/button | accion | - | WhatsApp portal Contacto | sí |
| 14 | button/button | accion | - | reCAPTCHA Seguridad | sí |
| 15 | button/button | accion | - | 2FA login Acceso | sí |
| 16 | button/button | accion | - | IA global Modelos | sí |
| 17 | button/button | accion | - | Respaldo Backup | sí |
| 18 | button/button | accion | - | &#9776; Ocultar menu | sí |
| 19 | button/button | accion | btnReloadConsumos | Actualizar consumos | no |
| 20 | button/button | accion | btnSaveConsumos | Guardar | no |
| 21 | button/button | accion | btnEditConsumos | Editar | no |
| 22 | button/button | accion | btnCancelConsumos | Cancelar | no |
| 23 | input/- | entrada | openaiCostPer1M | openaiCostPer1M | no |
| 24 | input/checkbox | accion | hostingerEnabled | hostingerEnabled | no |
| 25 | input/- | entrada | hostBwUsed | hostBwUsed | no |
| 26 | input/- | entrada | hostBwLimit | hostBwLimit | no |
| 27 | input/- | entrada | hostDiskUsed | hostDiskUsed | no |
| 28 | input/- | entrada | hostDiskLimit | hostDiskLimit | no |
| 29 | input/- | entrada | hostCostMonth | hostCostMonth | no |
| 30 | input/- | entrada | hostApiToken | hostApiToken | no |
| 31 | input/checkbox | accion | cursorEnabled | cursorEnabled | no |
| 32 | input/- | entrada | cursorCostMonth | cursorCostMonth | no |
| 33 | input/- | entrada | cursorApiKey | cursorApiKey | no |
| 34 | input/- | entrada | cursorNotes | cursorNotes | no |
| 35 | input/text | entrada | rustdeskServerHost | rustdeskServerHost | no |
| 36 | input/text | entrada | rustdeskServerKey | rustdeskServerKey | no |
| 37 | input/checkbox | accion | rustdeskActiveToggle | Activar RustDesk | no |
| 38 | input/checkbox | accion | rustdeskSshEnabledToggle | Usar SSH para gestionar RustDesk en el VPS | no |
| 39 | input/text | entrada | rustdeskSshHost | rustdeskSshHost | no |
| 40 | input/text | entrada | rustdeskSshUser | rustdeskSshUser | no |
| 41 | input/text | entrada | rustdeskSshKeyPath | rustdeskSshKeyPath | no |
| 42 | button/button | accion | rustdeskSaveSshBtn | Guardar SSH | no |
| 43 | button/button | accion | rustdeskEditSshBtn | Editar SSH | no |
| 44 | button/button | accion | rustdeskCancelSshBtn | Cancelar | no |
| 45 | button/button | accion | rustdeskStartBtn | Iniciar | no |
| 46 | button/button | accion | rustdeskRestartBtn | Reiniciar | no |
| 47 | button/button | accion | rustdeskStopBtn | Detener | no |
| 48 | button/button | accion | rustdeskProbeBtn | Probar funcionamiento | no |
| 49 | a/- | accion | - | Abrir | no |
| 50 | a/- | accion | - | Abrir | no |
| 51 | a/- | accion | - | Abrir | no |
| 52 | a/- | accion | - | Abrir | no |
| 53 | a/- | accion | - | Abrir | no |
| 54 | input/number | entrada | empresaLimitAPIRequests | empresaLimitAPIRequests | no |
| 55 | input/number | entrada | empresaLimitDBQueries | empresaLimitDBQueries | no |
| 56 | input/number | entrada | empresaLimitRustDeskMinutes | empresaLimitRustDeskMinutes | no |
| 57 | input/number | entrada | empresaLimitAIConsultas | empresaLimitAIConsultas | no |
| 58 | input/number | entrada | empresaLimitGPSDispositivos | empresaLimitGPSDispositivos | no |
| 59 | input/number | entrada | empresaLimitDBMaxGB | empresaLimitDBMaxGB | no |
| 60 | button/button | accion | saveEmpresaLimitacionesBtn | Guardar limitaciones | no |
| 61 | button/button | accion | editEmpresaLimitacionesBtn | Editar | no |
| 62 | button/button | accion | cancelEmpresaLimitacionesBtn | Cancelar | no |
| 63 | input/checkbox | accion | empresaStorageQuotaEnabled | Activar cuota de almacenamiento | no |
| 64 | input/number | entrada | empresaStorageDefaultLimitMB | empresaStorageDefaultLimitMB | no |
| 65 | input/number | entrada | empresaStorageWarnPercent | empresaStorageWarnPercent | no |
| 66 | input/number | entrada | empresaStorageMaxUploadMB | empresaStorageMaxUploadMB | no |
| 67 | input/checkbox | accion | empresaStorageBlockUploads | Bloquear cargas al superar limite | no |
| 68 | button/button | accion | empresaStorageSaveBtn | Guardar almacenamiento | no |
| 69 | button/button | accion | empresaStorageReloadBtn | Actualizar uso | no |
| 70 | input/checkbox | accion | onlyofficeEnabledToggle | Activar OnlyOffice | no |
| 71 | input/text | entrada | onlyofficeDSUrl | onlyofficeDSUrl | no |
| 72 | input/password | entrada | onlyofficeJWTSecret | onlyofficeJWTSecret | no |
| 73 | button/button | accion | onlyofficeSaveBtn | Guardar OnlyOffice | no |
| 74 | button/button | accion | onlyofficeEditBtn | Editar | no |
| 75 | button/button | accion | onlyofficeCancelBtn | Cancelar | no |
| 76 | button/button | accion | onlyofficeTestBtn | Probar conexión | no |
| 77 | input/checkbox | accion | adminPageURLsEnabledToggle | Mostrar URL real de subpaginas empresariales | no |
| 78 | button/button | accion | saveAdminPageURLsBtn | Guardar URLs visibles | no |
| 79 | button/button | accion | editAdminPageURLsBtn | Editar | no |
| 80 | button/button | accion | cancelAdminPageURLsBtn | Cancelar | no |
| 81 | a/- | accion | - | Abrir configuracion de voz IA | no |
| 82 | button/button | accion | voiceStreamQuickActivateTestBtn | Activar y probar | no |
| 83 | button/button | accion | voiceStreamQuickTestBtn | Probar servicio | no |
| 84 | input/checkbox | accion | epaycoEnabledToggle | Activar Epayco | no |
| 85 | input/text | entrada | epaycoPublicKey | epaycoPublicKey | no |
| 86 | input/password | entrada | epaycoPrivateKey | epaycoPrivateKey | no |
| 87 | input/text | entrada | epaycoCustomerId | epaycoCustomerId | no |
| 88 | input/password | entrada | epaycoCheckoutKey | epaycoCheckoutKey | no |
| 89 | input/checkbox | accion | epaycoCountryCO | epaycoCountryCO | no |
| 90 | input/checkbox | accion | epaycoCountryEC | epaycoCountryEC | no |
| 91 | input/checkbox | accion | epaycoCountryPA | epaycoCountryPA | no |
| 92 | input/checkbox | accion | epaycoCountryMX | epaycoCountryMX | no |
| 93 | input/checkbox | accion | epaycoCountryUS | epaycoCountryUS | no |
| 94 | input/checkbox | accion | epaycoCountryES | epaycoCountryES | no |
| 95 | button/button | accion | saveEpaycoBtn | Guardar cambios | no |
| 96 | button/button | accion | editEpaycoBtn | Editar | no |
| 97 | button/button | accion | cancelEpaycoBtn | Cancelar | no |
| 98 | button/button | accion | testEpaycoBtn | Probar Epayco | no |
| 99 | input/checkbox | accion | wompiModeToggle | Cambiar entre sandbox y real | no |
| 100 | input/checkbox | accion | wompiEnabledToggle | Activar Wompi | no |
| 101 | input/text | entrada | wompiPublicKey | wompiPublicKey | no |
| 102 | input/password | entrada | wompiPrivateKey | wompiPrivateKey | no |
| 103 | input/password | entrada | wompiIntegrityKey | wompiIntegrityKey | no |
| 104 | input/checkbox | accion | wompiCountryCO | wompiCountryCO | no |
| 105 | input/checkbox | accion | wompiCountryEC | wompiCountryEC | no |
| 106 | input/checkbox | accion | wompiCountryPA | wompiCountryPA | no |
| 107 | input/checkbox | accion | wompiCountryMX | wompiCountryMX | no |
| 108 | input/checkbox | accion | wompiCountryUS | wompiCountryUS | no |
| 109 | input/checkbox | accion | wompiCountryES | wompiCountryES | no |
| 110 | button/button | accion | saveWompiBtn | Guardar cambios | no |
| 111 | button/button | accion | editWompiBtn | Editar | no |
| 112 | button/button | accion | cancelWompiBtn | Cancelar | no |
| 113 | button/button | accion | testWompiBtn | Probar Nequi | no |
| 114 | input/email | entrada | gmailSMTPEmail | gmailSMTPEmail | no |
| 115 | input/password | entrada | gmailSMTPAppPass | gmailSMTPAppPass | no |
| 116 | input/text | entrada | gmailFromName | gmailFromName | no |
| 117 | input/text | entrada | gmailHost | gmailHost | no |
| 118 | input/text | entrada | gmailPort | gmailPort | no |
| 119 | input/text | entrada | gmailBaseURL | gmailBaseURL | no |
| 120 | input/email | entrada | gmailRestartAlertTo | gmailRestartAlertTo | no |
| 121 | input/checkbox | accion | gmailRestartAlertEnabled | Activar alertas de reinicio del servidor | no |
| 122 | button/- | accion | saveGmailBtn | Guardar Gmail SMTP | no |
| 123 | button/button | accion | editGmailBtn | Editar | no |
| 124 | button/button | accion | cancelGmailBtn | Cancelar | no |
| 125 | button/button | accion | testGmailBtn | Probar Gmail | no |
| 126 | input/checkbox | accion | licenciaVencimientoEnabled | Activar alertas de vencimiento de licencias | no |
| 127 | input/text | entrada | licenciaVencimientoDias | licenciaVencimientoDias | no |
| 128 | input/number | entrada | licenciaVencimientoMax | licenciaVencimientoMax | no |
| 129 | input/checkbox | accion | licenciaRetencionEnabled | Activar eliminacion programada de empresas con licencia vencida | no |
| 130 | input/number | entrada | licenciaRetencionDiasEspera | licenciaRetencionDiasEspera | no |
| 131 | input/number | entrada | licenciaRetencionDiasPreaviso | licenciaRetencionDiasPreaviso | no |
| 132 | input/number | entrada | licenciaRetencionMax | licenciaRetencionMax | no |
| 133 | button/button | accion | previewLicenciaRetencionBtn | Vista previa retencion | no |
| 134 | button/button | accion | runLicenciaRetencionBtn | Procesar retencion | no |
| 135 | button/button | accion | saveLicenciaVencimientoBtn | Guardar alertas | no |
| 136 | button/button | accion | previewLicenciaVencimientoBtn | Vista previa | no |
| 137 | button/button | accion | runLicenciaVencimientoBtn | Enviar ahora | no |
| 138 | input/text | entrada | portalWhatsAppNumber | portalWhatsAppNumber | no |
| 139 | button/- | accion | savePortalWhatsAppBtn | Guardar WhatsApp del portal | no |
| 140 | button/button | accion | editPortalWhatsAppBtn | Editar | no |
| 141 | button/button | accion | cancelPortalWhatsAppBtn | Cancelar | no |
| 142 | input/checkbox | accion | recaptchaEnabledToggle | Activar Google reCAPTCHA | no |
| 143 | input/text | entrada | recaptchaSiteKey | recaptchaSiteKey | no |
| 144 | input/password | entrada | recaptchaSecretKey | recaptchaSecretKey | no |
| 145 | select/- | entrada | recaptchaProvider | v2 (Checkbox visible) v3 (Invisible / token por acción) Enterprise (según tus llaves) | no |
| 146 | button/button | accion | saveRecaptchaBtn | Guardar reCAPTCHA | no |
| 147 | button/button | accion | editRecaptchaBtn | Editar | no |
| 148 | button/button | accion | cancelRecaptchaBtn | Cancelar | no |
| 149 | button/button | accion | recaptchaTestBtn | Probar reCAPTCHA | no |
| 150 | button/button | accion | recaptchaGetTokenBtn | Obtener token | no |
| 151 | button/button | accion | recaptchaResetBtn | Reset widget | no |
| 152 | input/checkbox | accion | admin2FAEnabledToggle | Activar 2FA en login de administradores | no |
| 153 | button/button | accion | saveAdmin2FABtn | Guardar 2FA login | no |
| 154 | button/button | accion | editAdmin2FABtn | Editar | no |
| 155 | button/button | accion | cancelAdmin2FABtn | Cancelar | no |
| 156 | input/checkbox | accion | aiProviderOpenAIEnabled | Activar OpenAI en los chats | no |
| 157 | input/password | entrada | aiKeyOpenAI | aiKeyOpenAI | no |
| 158 | select/- | entrada | aiOperationModel | aiOperationModel | no |
| 159 | select/- | entrada | aiAttachmentModel | aiAttachmentModel | no |
| 160 | input/checkbox | accion | aiEnabledToggle | Activar servicio global de IA | no |
| 161 | button/- | accion | saveAiConfigBtn | Guardar gobierno IA | no |
| 162 | button/button | accion | editAiConfigBtn | Editar | no |
| 163 | button/button | accion | cancelAiConfigBtn | Cancelar | no |
| 164 | button/button | accion | testAiBtn | Probar OpenAI | no |
| 165 | a/- | accion | - | Limites y contexto | no |
| 166 | a/- | accion | - | Conocimiento IA | no |
| 167 | a/- | accion | - | Voz IA | no |
| 168 | button/button | accion | downloadConfigBackupBtn | Descargar respaldo JSON | no |
| 169 | input/file | accion | restoreConfigFile | restoreConfigFile | no |
| 170 | button/button | accion | restoreConfigBtn | Restaurar respaldo | no |
| 171 | button/button | accion | testBackupBtn | Probar respaldo | no |

### `web/super/configuracion_logica_del_chat_con_ia.html` (15)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/checkbox | accion | empresaChatEnabled | Habilitar chat IA de empresas | no |
| 2 | input/number | entrada | empresaMaxConsultas | empresaMaxConsultas | no |
| 3 | input/checkbox | accion | empresaStreamingEnabled | Habilitar streaming en chat IA de empresas | no |
| 4 | input/checkbox | accion | empresaDBQueryEnabled | Permitir consultas de lectura a la base de datos para el chat IA empresarial | no |
| 5 | input/number | entrada | empresaDBQueryMaxTables | empresaDBQueryMaxTables | no |
| 6 | input/number | entrada | empresaDBQueryRows | empresaDBQueryRows | no |
| 7 | input/number | entrada | empresaMaxGPT55Consultas | empresaMaxGPT55Consultas | no |
| 8 | button/button | accion | saveEmpresaChatBtn | Guardar chat IA de empresas | no |
| 9 | input/checkbox | accion | portalChatEnabled | Habilitar chat público del portal | no |
| 10 | button/button | accion | savePortalChatBtn | Guardar chat público | no |
| 11 | input/checkbox | accion | superChatEnabled | Habilitar chat IA global super | no |
| 12 | input/number | entrada | superMaxConsultas | superMaxConsultas | no |
| 13 | input/checkbox | accion | superStreamingEnabled | Habilitar streaming en chat IA global | no |
| 14 | input/checkbox | accion | empresaSoloLectura | Habilitar en el chat global fragmentos de solo lectura de la base de empresas | no |
| 15 | button/button | accion | saveSuperChatBtn | Guardar chat IA global | no |

### `web/super/contexto_ia_logica_negocio.html` (3)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnReload | Recargar | no |
| 2 | button/button | accion | btnSave | Guardar contexto | no |
| 3 | textarea/- | entrada | contextText | contextText | no |

### `web/super/contrato.html` (11)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Ver contrato público | no |
| 2 | button/button | accion | contractReloadBtn | Recargar | no |
| 3 | button/button | accion | contractSaveBtn | Guardar nueva versión | no |
| 4 | input/text | entrada | contractTitle | contractTitle | no |
| 5 | input/text | entrada | contractChangeSummary | contractChangeSummary | no |
| 6 | textarea/- | entrada | contractSummary | contractSummary | no |
| 7 | textarea/- | entrada | contractContent | contractContent | no |
| 8 | textarea/- | entrada | contractAcceptanceNote | contractAcceptanceNote | no |
| 9 | button/button | accion | contractSaveBtnBottom | Guardar nueva versión | no |
| 10 | button/button | accion | - | Usar como base | sí |
| 11 | button/button | accion | - | Ver público | sí |

### `web/super/correos_masivos.html` (9)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnPreview | Previsualizar | no |
| 2 | button/button | accion | btnReload | Actualizar | no |
| 3 | select/- | entrada | alcance | Administradores y usuarios Solo administradores Solo usuarios de empresa | no |
| 4 | select/- | entrada | categoria | Informacion importante Politicas Actualizaciones Mantenimiento Seguridad Otro | no |
| 5 | input/text | entrada | asunto | asunto | no |
| 6 | textarea/- | entrada | cuerpoTexto | cuerpoTexto | no |
| 7 | input/text | entrada | observaciones | observaciones | no |
| 8 | input/checkbox | accion | confirmar | confirmar | no |
| 9 | button/submit | accion | btnSend | Enviar correo masivo | no |

### `web/super/docker_portabilidad.html` (19)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | refreshBtn | Actualizar | no |
| 2 | input/checkbox | accion | snapshotEnabled | snapshotEnabled | no |
| 3 | input/checkbox | accion | snapshotAutoEnabled | snapshotAutoEnabled | no |
| 4 | input/checkbox | accion | snapshotDeleteOldLocal | snapshotDeleteOldLocal | no |
| 5 | input/number | entrada | snapshotIntervalHours | snapshotIntervalHours | no |
| 6 | input/time | entrada | snapshotDailyTime | snapshotDailyTime | no |
| 7 | input/number | entrada | snapshotRetentionDays | snapshotRetentionDays | no |
| 8 | input/checkbox | accion | snapshotIncludeImages | snapshotIncludeImages | no |
| 9 | input/checkbox | accion | snapshotCloudEnabled | snapshotCloudEnabled | no |
| 10 | select/- | entrada | snapshotCloudProvider | rclone generico Google Drive Mega OneDrive S3 compatible | no |
| 11 | input/text | entrada | snapshotRcloneRemotePath | snapshotRcloneRemotePath | no |
| 12 | input/checkbox | accion | snapshotDeleteOldCloud | snapshotDeleteOldCloud | no |
| 13 | select/- | entrada | snapshotScope | Completa: proyecto, bases y volúmenes Solo bases PostgreSQL | no |
| 14 | input/checkbox | accion | snapshotIncludePostgres | snapshotIncludePostgres | no |
| 15 | input/checkbox | accion | snapshotIncludeVolumes | snapshotIncludeVolumes | no |
| 16 | button/button | accion | saveSnapshotConfigBtn | Guardar configuracion snapshot | no |
| 17 | button/button | accion | createSnapshotBtn | Crear y descargar snapshot | no |
| 18 | button/button | accion | createUploadSnapshotBtn | Crear y subir a nube | no |
| 19 | button/button | accion | refreshSnapshotBtn | Refrescar historial | no |

### `web/super/domotica_raspberry_trafico.html` (5)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | reloadBtn | Actualizar | no |
| 2 | input/number | entrada | - | '+esc(policy.limite_mensual_mb\|\|2048)+' | no |
| 3 | input/number | entrada | - | '+esc(policy.alerta_porcentaje\|\|80)+' | no |
| 4 | input/checkbox | accion | - | sin etiqueta | no |
| 5 | button/button | accion | - | Guardar | no |

### `web/super/domotica_storage.html` (5)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | reloadBtn | Actualizar | no |
| 2 | input/number | entrada | defaultMaxKb | 2048 | no |
| 3 | button/button | accion | saveDefaultBtn | Guardar limite general | no |
| 4 | input/number | entrada | - | ' + esc(e.max_image_kb \|\| 2048) + ' | no |
| 5 | button/button | accion | - | Guardar | no |

### `web/super/email_corporativo.html` (20)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Panel super | no |
| 2 | button/button | accion | testMailuBtn | Probar Mailu | no |
| 3 | button/button | accion | provisionSystemBtn | Provisionar ventas/soporte | no |
| 4 | input/email | entrada | testRecipient | testRecipient | no |
| 5 | button/button | accion | testSendBtn | Probar envio | no |
| 6 | input/checkbox | accion | enabled | enabled | no |
| 7 | input/checkbox | accion | autoCreate | autoCreate | no |
| 8 | input/- | entrada | domain | domain | no |
| 9 | input/- | entrada | webmailURL | webmailURL | no |
| 10 | input/- | entrada | logoURL | logoURL | no |
| 11 | select/- | entrada | provisionMode | Manual / pendiente Mailu directo VPS | no |
| 12 | input/number | entrada | quotaMB | quotaMB | no |
| 13 | input/number | entrada | maxAccountsPerEmpresa | maxAccountsPerEmpresa | no |
| 14 | input/- | entrada | apiBaseURL | apiBaseURL | no |
| 15 | input/- | entrada | apiAdmin | apiAdmin | no |
| 16 | input/password | entrada | apiPassword | apiPassword | no |
| 17 | input/- | entrada | directCommand | directCommand | no |
| 18 | button/submit | accion | - | Guardar configuracion | no |
| 19 | button/button | accion | syncBtn | Sincronizar empresas existentes | no |
| 20 | button/- | accion | - | Provisionar | no |

### `web/super/empresas.html` (8)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | refreshBtn | Actualizar | no |
| 2 | input/number | entrada | retentionMonths | 6 | no |
| 3 | button/button | accion | previewCleanupBtn | Previsualizar | no |
| 4 | button/button | accion | deleteCleanupBtn | Eliminar candidatas | no |
| 5 | input/search | entrada | searchInput | searchInput | no |
| 6 | select/- | entrada | licenseFilter | Todas Con licencia activa Sin licencia activa Con licencia de 15 dias Con licencia vencida | no |
| 7 | button/submit | accion | - | Buscar | no |
| 8 | button/button | accion | clearBtn | Limpiar | no |

### `web/super/explorador_archivos.html` (4)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | fileExplorerUp | Arriba | no |
| 2 | input/- | entrada | fileExplorerPath | Ruta actual | no |
| 3 | button/submit | accion | - | Abrir | no |
| 4 | button/button | accion | fileExplorerRefresh | Recargar | no |

### `web/super/formato_para_emviar_email.html` (30)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | emailTemplatesSaveTop | Guardar cambios | no |
| 2 | button/button | accion | - | Confirmación de correo | sí |
| 3 | button/button | accion | - | Administración | sí |
| 4 | button/button | accion | - | Pago de licencia | sí |
| 5 | button/button | accion | - | Recomendados | sí |
| 6 | input/text | entrada | - | sin etiqueta | sí |
| 7 | textarea/- | entrada | - | sin etiqueta | sí |
| 8 | textarea/- | entrada | - | sin etiqueta | sí |
| 9 | input/text | entrada | - | sin etiqueta | sí |
| 10 | textarea/- | entrada | - | sin etiqueta | sí |
| 11 | textarea/- | entrada | - | sin etiqueta | sí |
| 12 | input/text | entrada | - | sin etiqueta | sí |
| 13 | textarea/- | entrada | - | sin etiqueta | sí |
| 14 | textarea/- | entrada | - | sin etiqueta | sí |
| 15 | input/text | entrada | - | sin etiqueta | sí |
| 16 | textarea/- | entrada | - | sin etiqueta | sí |
| 17 | textarea/- | entrada | - | sin etiqueta | sí |
| 18 | input/text | entrada | - | sin etiqueta | sí |
| 19 | textarea/- | entrada | - | sin etiqueta | sí |
| 20 | textarea/- | entrada | - | sin etiqueta | sí |
| 21 | input/text | entrada | - | sin etiqueta | sí |
| 22 | textarea/- | entrada | - | sin etiqueta | sí |
| 23 | textarea/- | entrada | - | sin etiqueta | sí |
| 24 | input/text | entrada | - | sin etiqueta | sí |
| 25 | textarea/- | entrada | - | sin etiqueta | sí |
| 26 | textarea/- | entrada | - | sin etiqueta | sí |
| 27 | input/text | entrada | - | sin etiqueta | sí |
| 28 | textarea/- | entrada | - | sin etiqueta | sí |
| 29 | textarea/- | entrada | - | sin etiqueta | sí |
| 30 | button/button | accion | emailTemplatesSaveBottom | Guardar cambios | no |

### `web/super/informacion_de_la_empresa_y_de_los_sistemas_para_ia.html` (3)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnReload | Recargar | no |
| 2 | button/button | accion | btnSave | Guardar | no |
| 3 | textarea/- | entrada | infoText | infoText | no |

### `web/super/informacion_de_modulos.html` (9)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnReload | Recargar | no |
| 2 | button/button | accion | btnDefaults | Cargar predeterminado | no |
| 3 | button/button | accion | btnAddModule | Agregar modulo | no |
| 4 | button/button | accion | btnSave | Guardar cambios | no |
| 5 | input/text | entrada | modulesTitle | modulesTitle | no |
| 6 | button/button | accion | - | Eliminar | sí |
| 7 | input/text | entrada | - | ' + esc(mod.titulo) + ' | sí |
| 8 | input/text | entrada | - | ' + esc(mod.icono_url) + ' | sí |
| 9 | textarea/- | entrada | - | ' + esc(mod.caracteristicas.join('\n')) + ' | sí |

### `web/super/integracion_ia.html` (6)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/checkbox | accion | aiGlobalEnabled | Habilitar IA global | no |
| 2 | button/button | accion | aiGlobalSaveBtn | Guardar configuración | no |
| 3 | button/button | accion | aiGlobalTestBtn | Probar conexión | no |
| 4 | button/button | accion | aiProvidersSaveBtn | Guardar credenciales y proveedores | no |
| 5 | input/password | entrada | ' + inputId + ' | ' + inputId + ' | no |
| 6 | input/checkbox | accion | providerToggle_' + escapeHtml(model.provider) + ' | Habilitar proveedor ' + escapeHtml(model.provider) + ' | no |

### `web/super/licencias.html` (104)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | historyPageBtn | Historial de licencias | no |
| 2 | button/- | accion | addBtn | Agregar | no |
| 3 | select/- | entrada | historyEmpresaFilter | historyEmpresaFilter | no |
| 4 | button/button | accion | quickBaseBtn | Nuevo plan base | no |
| 5 | button/button | accion | quickAddonBtn | Nuevo addon | no |
| 6 | input/- | entrada | licSearch | licSearch | no |
| 7 | select/- | entrada | licTypeFilter | Todos | no |
| 8 | select/- | entrada | licCountryFilter | Todos | no |
| 9 | select/- | entrada | licClassFilter | Todas Base Adicional | no |
| 10 | select/- | entrada | licStatusFilter | Todos Visibles para clientes Ocultas para clientes Asignadas Catálogo | no |
| 11 | button/button | accion | clearFiltersBtn | Limpiar filtros | no |
| 12 | input/number | entrada | maxAdvanceSameLicense | 2 | no |
| 13 | input/number | entrada | maxLicenseUnitsPerCheckout | 5 | no |
| 14 | input/checkbox | accion | licenseWelcomeEmailEnabled | licenseWelcomeEmailEnabled | no |
| 15 | input/checkbox | accion | licenseAutoInvoiceEnabled | licenseAutoInvoiceEnabled | no |
| 16 | input/checkbox | accion | licenseAttachInvoicePdfEnabled | licenseAttachInvoicePdfEnabled | no |
| 17 | button/button | accion | saveLicenseConfigBtn | Guardar reglas | no |
| 18 | input/hidden | entrada | itemId | itemId | no |
| 19 | select/- | entrada | tipo_id | tipo_id | no |
| 20 | select/- | entrada | pais_codigo | Global / todos los paises Colombia Perú Ecuador México Chile Argentina | no |
| 21 | input/- | entrada | nombre | nombre | no |
| 22 | input/number | entrada | valor | valor | no |
| 23 | input/number | entrada | duracion_dias | duracion_dias | no |
| 24 | input/number | entrada | max_documentos_mensuales | max_documentos_mensuales | no |
| 25 | input/checkbox | accion | licencia_activa | licencia_activa | no |
| 26 | textarea/- | entrada | descripcion | descripcion | no |
| 27 | button/button | accion | - | Todos sin restricción | sí |
| 28 | button/button | accion | - | Operativo | sí |
| 29 | button/button | accion | - | Financiero | sí |
| 30 | button/button | accion | - | Comercial | sí |
| 31 | button/button | accion | - | Enterprise | sí |
| 32 | input/checkbox | accion | - | ventas | no |
| 33 | input/checkbox | accion | - | inventario | no |
| 34 | input/checkbox | accion | - | finanzas | no |
| 35 | input/checkbox | accion | - | contabilidad_colombia | no |
| 36 | input/checkbox | accion | - | contabilidad_colombia_avanzada | no |
| 37 | input/checkbox | accion | - | centros_costo | no |
| 38 | input/checkbox | accion | - | cierre_fiscal | no |
| 39 | input/checkbox | accion | - | declaraciones_tributarias | no |
| 40 | input/checkbox | accion | - | activos_fijos_niif_fiscal | no |
| 41 | input/checkbox | accion | - | portal_terceros_certificados | no |
| 42 | input/checkbox | accion | - | bancos_pagos | no |
| 43 | input/checkbox | accion | - | cumplimiento_kyc | no |
| 44 | input/checkbox | accion | - | calidad_procesos | no |
| 45 | input/checkbox | accion | - | clientes | no |
| 46 | input/checkbox | accion | - | crm_unificado | no |
| 47 | input/checkbox | accion | - | compras | no |
| 48 | input/checkbox | accion | - | soportes_compras_ia | no |
| 49 | input/checkbox | accion | - | gestion_documental | no |
| 50 | input/checkbox | accion | - | contratos_obligaciones | no |
| 51 | input/checkbox | accion | - | facturacion | no |
| 52 | input/checkbox | accion | - | facturacion_ecuador | no |
| 53 | input/checkbox | accion | - | facturacion_panama | no |
| 54 | input/checkbox | accion | - | seguridad | no |
| 55 | input/checkbox | accion | - | auditoria | no |
| 56 | input/checkbox | accion | - | backups | no |
| 57 | input/checkbox | accion | - | documentos_onlyoffice | no |
| 58 | input/checkbox | accion | - | venta_publica | no |
| 59 | input/checkbox | accion | - | reservas_hotel | no |
| 60 | input/checkbox | accion | - | chat_tareas | no |
| 61 | input/checkbox | accion | - | domicilios | no |
| 62 | input/checkbox | accion | - | produccion_mrp | no |
| 63 | input/checkbox | accion | - | logistica_wms | no |
| 64 | input/checkbox | accion | - | importaciones_costeo | no |
| 65 | input/checkbox | accion | - | aiu_construccion | no |
| 66 | input/checkbox | accion | - | tesoreria_presupuesto | no |
| 67 | input/checkbox | accion | - | cobranza | no |
| 68 | input/checkbox | accion | - | reportes | no |
| 69 | input/checkbox | accion | - | bolsa | no |
| 70 | input/checkbox | accion | - | portal_contador | no |
| 71 | input/checkbox | accion | - | nomina_sueldos | no |
| 72 | input/checkbox | accion | - | parqueadero | no |
| 73 | input/checkbox | accion | - | alquileres | no |
| 74 | input/checkbox | accion | - | turnos_atencion | no |
| 75 | input/checkbox | accion | - | control_electrico | no |
| 76 | input/checkbox | accion | - | energia_solar | no |
| 77 | input/checkbox | accion | - | camaras | no |
| 78 | input/checkbox | accion | - | carnets | no |
| 79 | input/checkbox | accion | - | horarios_trabajadores | no |
| 80 | input/checkbox | accion | - | asistencia_empleados | no |
| 81 | input/checkbox | accion | - | vehiculos_registro | no |
| 82 | input/checkbox | accion | - | hoja_vida_operativa | no |
| 83 | input/checkbox | accion | - | ubicacion_gps | no |
| 84 | input/checkbox | accion | super_rol_habilitado | super_rol_habilitado | no |
| 85 | input/checkbox | accion | es_adicional | es_adicional | no |
| 86 | input/- | entrada | codigo_funcion | codigo_funcion | no |
| 87 | button/submit | accion | saveBtn | Guardar | no |
| 88 | button/button | accion | cancelBtn | Cancelar | no |
| 89 | input/checkbox | accion | - | ' + escapeHTML(item.module) + ' | no |
| 90 | a/- | accion | - | Ver otras licencias | no |
| 91 | a/- | accion | - | Renovar licencia | no |
| 92 | a/- | accion | - | Cambiar licencia | no |
| 93 | a/- | accion | + encodeURIComponent(item.empresa_id) + | Abrir empresa | no |
| 94 | a/- | accion | - | Elegir licencia | no |
| 95 | a/- | accion | + encodeURIComponent(empresaID) + | Editar empresa | no |
| 96 | a/- | accion | + encodeURIComponent(empresaID) + | Abrir empresa | no |
| 97 | a/- | accion | - | Ver otras licencias | no |
| 98 | a/- | accion | - | Renovar licencia | no |
| 99 | a/- | accion | - | Cambiar licencia | no |
| 100 | a/- | accion | + encodeURIComponent(empresaID) + | Editar empresa | no |
| 101 | a/- | accion | + encodeURIComponent(empresaID) + | Abrir empresa | no |
| 102 | button/button | accion | - | Anterior | sí |
| 103 | button/button | accion | - | = totalPages ? 'disabled' : '') + '>Siguiente | sí |
| 104 | input/checkbox | accion | ' + escapeHTML(i.id) + ' | ' + escapeHTML(i.id) + ' | sí |

### `web/super/licencias_codigos_descuento.html` (15)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | reloadBtn | Actualizar | no |
| 2 | input/hidden | entrada | oldCode | oldCode | no |
| 3 | input/- | entrada | name | name | no |
| 4 | input/- | entrada | description | description | no |
| 5 | select/- | entrada | type | Porcentaje Valor fijo Gratis | no |
| 6 | input/number | entrada | value | 10 | no |
| 7 | input/email | entrada | email | email | no |
| 8 | input/date | entrada | expires | expires | no |
| 9 | input/checkbox | accion | active | active | no |
| 10 | button/submit | accion | saveBtn | Crear y enviar | no |
| 11 | button/button | accion | cancelBtn | Cancelar | no |
| 12 | button/button | accion | - | Editar | sí |
| 13 | button/button | accion | - | Reenviar | sí |
| 14 | button/button | accion | - | ' + (item.activo ? 'Desactivar' : 'Activar') + ' | sí |
| 15 | button/button | accion | - | Eliminar | sí |

### `web/super/licencias_resumen.html` (4)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | cmdRefresh | Actualizar | no |
| 2 | button/button | accion | cmdResetMetrics | Reiniciar metricas | no |
| 3 | button/button | accion | cmdResetErrors | Reiniciar errores | no |
| 4 | button/button | accion | cmdRunAlerts | Evaluar alertas | no |

### `web/super/mantenimiento_sistema.html` (16)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | reloadMantenimientoBtn | Actualizar | no |
| 2 | button/button | accion | cleanOldMantenimientoBtn | Eliminar alertas viejas | no |
| 3 | input/checkbox | accion | mantenimientoAvisoToggle | Mostrar aviso de mantenimiento programado | no |
| 4 | input/checkbox | accion | mantenimientoToggle | Activar modo mantenimiento | no |
| 5 | input/date | entrada | mantenimientoFecha | mantenimientoFecha | no |
| 6 | input/time | entrada | mantenimientoHoraInicio | mantenimientoHoraInicio | no |
| 7 | input/time | entrada | mantenimientoHoraFin | mantenimientoHoraFin | no |
| 8 | input/- | entrada | mantenimientoZonaHoraria | mantenimientoZonaHoraria | no |
| 9 | input/hidden | entrada | mantenimientoAvisoId | mantenimientoAvisoId | no |
| 10 | textarea/- | entrada | mantenimientoMensaje | mantenimientoMensaje | no |
| 11 | button/button | accion | saveMantenimientoBtn | Guardar mantenimiento | no |
| 12 | button/button | accion | editMantenimientoBtn | Editar | no |
| 13 | button/button | accion | cancelMantenimientoBtn | Cancelar | no |
| 14 | button/button | accion | ' + id + ' | Cargar | sí |
| 15 | button/button | accion | ' + id + ' | Desactivar | sí |
| 16 | button/button | accion | ' + id + ' | Eliminar | sí |

### `web/super/noticias.html` (24)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Ver pagina publica | no |
| 2 | button/button | accion | btnReload | Recargar | no |
| 3 | button/button | accion | btnDefaults | Cargar predeterminado | no |
| 4 | button/button | accion | btnAddNews | Agregar noticia | no |
| 5 | button/button | accion | btnSave | Guardar cambios | no |
| 6 | input/text | entrada | pageTitle | pageTitle | no |
| 7 | input/text | entrada | pageSubtitle | pageSubtitle | no |
| 8 | input/text | entrada | profileName | profileName | no |
| 9 | input/text | entrada | profileUser | profileUser | no |
| 10 | input/text | entrada | coverUrl | coverUrl | no |
| 11 | input/text | entrada | profileUrl | profileUrl | no |
| 12 | textarea/- | entrada | pageDescription | pageDescription | no |
| 13 | button/button | accion | - | Eliminar | sí |
| 14 | input/checkbox | accion | - | sin etiqueta | sí |
| 15 | input/checkbox | accion | - | sin etiqueta | sí |
| 16 | input/text | entrada | - | ' + esc(item.titulo) + ' | sí |
| 17 | input/text | entrada | - | ' + esc(item.categoria) + ' | sí |
| 18 | input/date | entrada | - | ' + esc(item.fecha) + ' | sí |
| 19 | input/text | entrada | - | ' + esc(item.imagen_url) + ' | sí |
| 20 | textarea/- | entrada | - | ' + esc(item.resumen) + ' | sí |
| 21 | textarea/- | entrada | - | ' + esc(item.contenido) + ' | sí |
| 22 | input/text | entrada | - | ' + esc(item.fuente_nombre) + ' | sí |
| 23 | input/url | entrada | - | ' + esc(item.fuente_url) + ' | sí |
| 24 | textarea/- | entrada | - | ' + esc((item.etiquetas \|\| []).join('\n')) + ' | sí |

### `web/super/pagina_principal.html` (22)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Ver landing descriptiva | no |
| 2 | button/button | accion | ppBtnReload | Recargar | no |
| 3 | input/number | entrada | ppCantidad | ppCantidad | no |
| 4 | button/button | accion | ppBtnSave | Guardar cantidad y configuración | no |
| 5 | select/- | entrada | ppIndexCardSize | ppIndexCardSize | sí |
| 6 | select/- | entrada | ppIndexTextSize | ppIndexTextSize | sí |
| 7 | select/- | entrada | ppLandingCardSize | ppLandingCardSize | sí |
| 8 | select/- | entrada | ppLandingTextSize | ppLandingTextSize | sí |
| 9 | button/button | accion | ppBtnSaveBottom | Guardar cambios | no |
| 10 | select/- | entrada | ppCardType' + idx + ' | ' + cardTypeOptionsHTML(cardType) + ' | sí |
| 11 | select/- | entrada | ppImage' + idx + ' | ' + imageOptionsHTML(image) + ' | sí |
| 12 | input/file | accion | - | sin etiqueta | sí |
| 13 | select/- | entrada | ppSecondaryImage' + idx + ' | ' + imageOptionsHTML(secondaryImage) + ' | sí |
| 14 | input/text | entrada | ppLink' + idx + ' | ' + escapeHTML(link) + ' | sí |
| 15 | input/url | entrada | ppYoutube' + idx + ' | ' + escapeHTML(youtubeURL) + ' | sí |
| 16 | input/text | entrada | ppTitle' + idx + ' | ' + escapeHTML(title) + ' | sí |
| 17 | textarea/- | entrada | ppDesc' + idx + ' | ' + escapeHTML(description) + ' | sí |
| 18 | input/text | entrada | ppDetailTag' + idx + ' | ' + escapeHTML(detailTag) + ' | sí |
| 19 | input/text | entrada | ppDetailHeadline' + idx + ' | ' + escapeHTML(detailHeadline) + ' | sí |
| 20 | textarea/- | entrada | ppDetailParagraphOne' + idx + ' | ' + escapeHTML(detailParagraphOne) + ' | sí |
| 21 | textarea/- | entrada | ppDetailParagraphTwo' + idx + ' | ' + escapeHTML(detailParagraphTwo) + ' | sí |
| 22 | textarea/- | entrada | ppDetailPoints' + idx + ' | ' + escapeHTML(detailPoints) + ' | sí |

### `web/super/permisos_rol.html` (14)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | select/- | entrada | tipo_empresa_id | tipo_empresa_id | no |
| 2 | select/- | entrada | rol_id | rol_id | no |
| 3 | input/search | entrada | prSearch | prSearch | no |
| 4 | input/- | entrada | newRolNombre | newRolNombre | no |
| 5 | input/- | entrada | newRolDescripcion | newRolDescripcion | no |
| 6 | button/button | accion | createRolBtn | Crear rol | no |
| 7 | button/button | accion | prExpandAll | Desplegar grupos | no |
| 8 | button/button | accion | prCollapseAll | Plegar grupos | no |
| 9 | button/button | accion | saveBtn | Guardar cambios | no |
| 10 | button/button | accion | saveBtn2 | Guardar cambios | no |
| 11 | input/checkbox | accion | - | sin etiqueta | sí |
| 12 | button/button | accion | - | Activar todo | sí |
| 13 | button/button | accion | - | Desactivar todo | sí |
| 14 | input/checkbox | accion | - | sin etiqueta | sí |

### `web/super/plantillas_produccion_masiva.html` (17)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Tipos de empresa | no |
| 2 | a/- | accion | - | Preconfiguraciones | no |
| 3 | a/- | accion | - | Licencias | no |
| 4 | button/button | accion | ensureBtn | Asegurar 20 | no |
| 5 | button/button | accion | exportBtn | Exportar CSV | no |
| 6 | input/search | entrada | searchInput | searchInput | no |
| 7 | button/button | accion | - | Todos | sí |
| 8 | button/button | accion | - | Listos | sí |
| 9 | button/button | accion | - | Produccion | sí |
| 10 | button/button | accion | - | No listos | sí |
| 11 | button/button | accion | - | Sin licencia | sí |
| 12 | button/button | accion | - | Sin preconfig | sí |
| 13 | button/button | accion | refreshBtn | Actualizar | no |
| 14 | button/button | accion | - | ' + ' ' + escapeHTML(row.title) + ' ' + escapeHTML(row.count) + ' ' + ' | sí |
| 15 | a/- | accion | - | Tipo | no |
| 16 | a/- | accion | - | Preconfig | no |
| 17 | a/- | accion | - | Licencias | no |

### `web/super/preconfiguracion_tipos_empresa.html` (30)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/search | entrada | typeSearch | typeSearch | no |
| 2 | input/hidden | entrada | tipoEmpresaID | tipoEmpresaID | no |
| 3 | input/checkbox | accion | enabled | enabled | no |
| 4 | input/- | entrada | nombre | nombre | no |
| 5 | textarea/- | entrada | descripcion | descripcion | no |
| 6 | input/checkbox | accion | estacionesEnabled | estacionesEnabled | no |
| 7 | input/number | entrada | estacionesCantidad | 5 | no |
| 8 | input/- | entrada | estacionesPrefijo | Mesa | no |
| 9 | select/- | entrada | cardSize | Pequena Media Grande | no |
| 10 | input/checkbox | accion | cajaEnabled | cajaEnabled | no |
| 11 | input/- | entrada | operacionTipoNegocio | operacionTipoNegocio | no |
| 12 | input/- | entrada | operacionSingular | operacionSingular | no |
| 13 | input/- | entrada | operacionPlural | operacionPlural | no |
| 14 | input/checkbox | accion | operacionUsaEstaciones | operacionUsaEstaciones | no |
| 15 | input/checkbox | accion | ventaDirectaEnabled | ventaDirectaEnabled | no |
| 16 | input/- | entrada | ventaDirectaNombre | Venta directa | no |
| 17 | input/checkbox | accion | comisionesEnabled | comisionesEnabled | no |
| 18 | input/- | entrada | comisionRol | comisionRol | no |
| 19 | input/number | entrada | comisionPorcentaje | 0 | no |
| 20 | input/- | entrada | comisionFiltro | comisionFiltro | no |
| 21 | textarea/- | entrada | rolesOperativosText | rolesOperativosText | no |
| 22 | textarea/- | entrada | productosText | productosText | no |
| 23 | textarea/- | entrada | usuariosText | usuariosText | no |
| 24 | input/checkbox | accion | asistenteEnabled | asistenteEnabled | no |
| 25 | input/- | entrada | asistenteRol | asistenteRol | no |
| 26 | textarea/- | entrada | instruccionesText | instruccionesText | no |
| 27 | textarea/- | entrada | tareasText | tareasText | no |
| 28 | button/submit | accion | - | Guardar preconfiguración | no |
| 29 | button/button | accion | resetDefaultBtn | Restaurar sugerida | no |
| 30 | button/button | accion | - | ' + ' ' + escapeHtml(tipo.nombre \|\| ('Tipo #' + tipo.id)) + ' ' + ' ' + escapeHtml(status + (item.es_default ? ' - suger | sí |

### `web/super/recordatorios_infraestructura.html` (13)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | select/- | entrada | tipo | Dominio Hosting VPS Certificado SSL Otro | no |
| 2 | input/- | entrada | nombre | nombre | no |
| 3 | input/- | entrada | proveedor | proveedor | no |
| 4 | input/date | entrada | fecha | fecha | no |
| 5 | input/number | entrada | dias | 30 | no |
| 6 | button/button | accion | addBtn | Agregar | no |
| 7 | button/button | accion | saveBtn | Guardar cambios | no |
| 8 | input/date | entrada | - | '+esc(it.fecha_vencimiento)+' | sí |
| 9 | input/number | entrada | - | '+esc(it.dias_aviso\|\|30)+' | sí |
| 10 | input/checkbox | accion | - | sin etiqueta | sí |
| 11 | input/checkbox | accion | - | sin etiqueta | sí |
| 12 | input/checkbox | accion | - | sin etiqueta | sí |
| 13 | button/button | accion | - | Quitar | sí |

### `web/super/reportes_globales.html` (16)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/date | entrada | rgFechaDesde | rgFechaDesde | no |
| 2 | input/date | entrada | rgFechaHasta | rgFechaHasta | no |
| 3 | select/- | entrada | rgDataset | rgDataset | no |
| 4 | select/- | entrada | rgFormato | PDF XLS / Excel CSV TXT JSON | no |
| 5 | input/hidden | entrada | rgModo | consolidado | no |
| 6 | input/hidden | entrada | rgTipoSeleccion | multiple | no |
| 7 | select/- | entrada | rgEmpresaUnica | rgEmpresaUnica | no |
| 8 | button/button | accion | rgVerDataset | Ver | no |
| 9 | button/button | accion | rgExportar | Exportar | no |
| 10 | button/button | accion | rgImprimir | Imprimir | no |
| 11 | button/button | accion | rgActualizar | Actualizar | no |
| 12 | input/email | entrada | rgEmailTo | rgEmailTo | no |
| 13 | button/button | accion | rgEnviarEmail | Enviar por email | no |
| 14 | button/button | accion | rgSeleccionarTodas | Todas | no |
| 15 | button/button | accion | rgSoloActivas | Activas | no |
| 16 | button/button | accion | rgLimpiarEmpresas | Limpiar | no |

### `web/super/roles_de_usuario.html` (16)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | a/- | accion | - | Roles y licencias | no |
| 2 | button/button | accion | addBtn | Nuevo rol | no |
| 3 | input/search | entrada | filterSearch | filterSearch | no |
| 4 | select/- | entrada | filterTipo | filterTipo | no |
| 5 | select/- | entrada | filterEstado | Todos Activos Inactivos | no |
| 6 | button/button | accion | clearFiltersBtn | Limpiar | no |
| 7 | input/hidden | entrada | itemId | itemId | no |
| 8 | select/- | entrada | tipo_empresa_id | tipo_empresa_id | no |
| 9 | input/- | entrada | nombre | nombre | no |
| 10 | textarea/- | entrada | descripcion | descripcion | no |
| 11 | button/submit | accion | saveBtn | Guardar rol | no |
| 12 | button/button | accion | cancelBtn | Cancelar | no |
| 13 | a/- | accion | - | Permisos | no |
| 14 | button/button | accion | ' + escapeHtml(item.id) + ' | Editar | sí |
| 15 | button/button | accion | ' + escapeHtml(item.id) + ' | ' + toggleLabel + ' | sí |
| 16 | button/button | accion | ' + escapeHtml(item.id) + ' | Eliminar | sí |

### `web/super/seguridad.html` (23)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/- | accion | runFullScanBtn | Ejecutar escaneo | no |
| 2 | button/- | accion | refreshHistoryBtn | Actualizar historial | no |
| 3 | input/text | entrada | scopeInput | Detectando | no |
| 4 | input/text | entrada | targetHostInput | targetHostInput | no |
| 5 | input/text | entrada | portListInput | portListInput | no |
| 6 | select/- | entrada | profileSelect | quick full | no |
| 7 | input/text | entrada | cronInput | cronInput | no |
| 8 | input/number | entrada | maxHistoryInput | maxHistoryInput | no |
| 9 | input/number | entrada | maxFindingsInput | maxFindingsInput | no |
| 10 | input/checkbox | accion | lynisEnabledInput | lynisEnabledInput | no |
| 11 | input/text | entrada | lynisCommandInput | lynisCommandInput | no |
| 12 | input/checkbox | accion | nmapEnabledInput | nmapEnabledInput | no |
| 13 | input/text | entrada | nmapCommandInput | nmapCommandInput | no |
| 14 | input/checkbox | accion | vulnEnabledInput | vulnEnabledInput | no |
| 15 | select/- | entrada | vulnProviderInput | trivy | no |
| 16 | input/text | entrada | vulnCommandInput | vulnCommandInput | no |
| 17 | input/text | entrada | vulnTargetInput | vulnTargetInput | no |
| 18 | button/submit | accion | - | Guardar configuración | no |
| 19 | button/button | accion | runQuickScanBtn | Ejecutar con este perfil | no |
| 20 | button/- | accion | scanPortsBtn | Escanear puertos | no |
| 21 | input/text | entrada | ipInput | 127.0.0.1 | no |
| 22 | input/text | entrada | portsInput | 22,80,443,3306,5432,8080,8443 | no |
| 23 | button/- | accion | refreshProcessesBtn | Actualizar procesos | no |

### `web/super/tickets_ayuda.html` (13)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnReload | Actualizar | no |
| 2 | input/search | entrada | filterQ | filterQ | no |
| 3 | select/- | entrada | filterEstado | Todos Nuevo En revision Respondido Cerrado | no |
| 4 | select/- | entrada | filterPrioridad | Todas Critica Alta Media Baja | no |
| 5 | input/number | entrada | filterLimit | 150 | no |
| 6 | button/button | accion | btnApplyFilters | Filtrar | no |
| 7 | select/- | entrada | detailEstado | Nuevo En revision Respondido Cerrado | no |
| 8 | select/- | entrada | detailPrioridad | Baja Media Alta Critica | no |
| 9 | input/text | entrada | detailAsignado | detailAsignado | no |
| 10 | textarea/- | entrada | responseMessage | responseMessage | no |
| 11 | input/checkbox | accion | responseInternal | responseInternal | no |
| 12 | button/submit | accion | btnSaveTicket | Guardar respuesta | no |
| 13 | button/button | accion | btnCloseTicket | Cerrar ticket | no |

### `web/super/tipos_empresas.html` (10)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/- | accion | addBtn | Agregar | no |
| 2 | input/hidden | entrada | itemId | itemId | no |
| 3 | input/- | entrada | nombre | nombre | no |
| 4 | textarea/- | entrada | observaciones | observaciones | no |
| 5 | button/submit | accion | saveBtn | Guardar | no |
| 6 | button/button | accion | cancelBtn | Cancelar | no |
| 7 | input/search | entrada | tipoSearch | tipoSearch | no |
| 8 | button/- | accion | ' + escapeHTML(i.id) + ' | Editar | sí |
| 9 | button/- | accion | ' + escapeHTML(i.id) + ' | Eliminar | sí |
| 10 | button/- | accion | ' + escapeHTML(i.id) + ' | ' + toggleLabel + ' | sí |

### `web/super/voz_streaming_ia.html` (16)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/checkbox | accion | voiceComputerMode | Usar voz rápida del computador | no |
| 2 | input/checkbox | accion | voiceNaturalMode | Usar voz natural del VPS | no |
| 3 | input/checkbox | accion | voiceStreamEnabled | Activar servidor de voz IA | no |
| 4 | select/- | entrada | voiceComputerGender | Femenina rapida Masculina rapida | no |
| 5 | input/text | entrada | voiceStreamProvider | piper | no |
| 6 | input/url | entrada | voiceStreamBaseURL | voiceStreamBaseURL | no |
| 7 | input/text | entrada | voiceStreamVoice | voiceStreamVoice | no |
| 8 | input/number | entrada | voiceStreamTimeoutMS | 12000 | no |
| 9 | input/text | entrada | voiceStreamAuthHeader | X-PCS-Voice-Token | no |
| 10 | input/password | entrada | voiceStreamAuthToken | voiceStreamAuthToken | no |
| 11 | input/checkbox | accion | voiceStreamGenerateToken | voiceStreamGenerateToken | no |
| 12 | button/button | accion | editVoiceStreamConfig | Editar | no |
| 13 | button/button | accion | saveVoiceStreamConfig | Guardar | no |
| 14 | button/button | accion | cancelVoiceStreamConfig | Cancelar | no |
| 15 | button/button | accion | activateTestVoiceStreamConfig | Activar y probar | no |
| 16 | button/button | accion | testVoiceStreamConfig | Probar servicio | no |

### `web/super/vps2.html` (12)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | refreshBtn | Actualizar | no |
| 2 | button/button | accion | restartCloudBtn | Reiniciar Nextcloud | no |
| 3 | button/button | accion | rebootBtn | Reiniciar VPS2 | no |
| 4 | button/button | accion | shutdownBtn | Apagar VPS2 | no |
| 5 | input/text | entrada | cfgHost | cfgHost | no |
| 6 | input/number | entrada | cfgPort | cfgPort | no |
| 7 | input/text | entrada | cfgUser | cfgUser | no |
| 8 | input/text | entrada | cfgRemotePath | cfgRemotePath | no |
| 9 | input/text | entrada | cfgNextcloudPath | cfgNextcloudPath | no |
| 10 | button/button | accion | saveConfigBtn | Guardar | no |
| 11 | button/button | accion | filesUpBtn | Subir | no |
| 12 | button/button | accion | filesRefreshBtn | Actualizar | no |

### `web/super_administrador.html` (19)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | - | ☰ Ocultar menú | sí |
| 2 | button/button | accion | superFavoriteBtn | &#9733; | no |
| 3 | button/button | accion | openAIDrawer | Asistente IA | sí |
| 4 | button/button | accion | aiChatMinibarExpand | Abrir asistente IA | no |
| 5 | button/button | accion | aiChatHintToggle | Ver ejemplos | no |
| 6 | button/button | accion | aiChatConfigBtn | Configurar chat flotante | no |
| 7 | button/button | accion | aiChatMinimize | Minimizar chat | no |
| 8 | button/button | accion | closeAIDrawer | × | no |
| 9 | button/button | accion | aiChatNewBtn | Nuevo chat | no |
| 10 | button/button | accion | aiChatConvBtn | Modo conversación | no |
| 11 | button/button | accion | aiChatMicBtn | Dictar mensaje | no |
| 12 | button/button | accion | aiChatVoiceBtn | Voz del asistente | no |
| 13 | button/button | accion | aiChatStopBtn | Detener audio y respuesta | no |
| 14 | input/hidden | entrada | aiChatMode | operativo | no |
| 15 | input/file | accion | aiChatAttachment | Adjuntar archivo para IA | no |
| 16 | button/button | accion | aiChatAttachBtn | Adjuntar archivo | no |
| 17 | button/button | accion | aiChatClearAttachment | × | no |
| 18 | textarea/- | entrada | aiChatInput | Mensaje al asistente IA | no |
| 19 | button/submit | accion | - | Enviar | no |

### `web/turnos_publicos.html` (1)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | printTicketBtn | Imprimir ticket | no |

### `web/venta_publica.html` (39)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/- | entrada | q | q | no |
| 2 | select/- | entrada | sort | Relevancia Precio: menor a mayor Precio: mayor a menor Nuevos | no |
| 3 | button/button | accion | searchBtn | Buscar | no |
| 4 | button/button | accion | cartBtn | Carrito ( 0 ) | no |
| 5 | input/- | entrada | orderName | orderName | no |
| 6 | input/- | entrada | orderPhone | orderPhone | no |
| 7 | input/- | entrada | orderEmail | orderEmail | no |
| 8 | select/- | entrada | orderChannel | Domicilio Recoger en tienda | no |
| 9 | input/- | entrada | orderAddress | orderAddress | no |
| 10 | textarea/- | entrada | orderNotes | orderNotes | no |
| 11 | input/checkbox | accion | orderShareLocation | orderShareLocation | no |
| 12 | button/button | accion | useCustomerLocationBtn | Usar mi ubicación | no |
| 13 | button/submit | accion | - | Enviar pedido | no |
| 14 | button/button | accion | closeCartBtn | Cerrar | no |
| 15 | button/button | accion | checkoutBtn | Ir a pagar | no |
| 16 | button/button | accion | clearCartBtn | Vaciar carrito | no |
| 17 | button/button | accion | openAIDrawer | Asistente IA | sí |
| 18 | button/button | accion | aiChatMinibarExpand | Abrir asistente IA | no |
| 19 | button/button | accion | aiChatHintToggle | Ver ejemplos | no |
| 20 | button/button | accion | aiChatConfigBtn | Configurar chat flotante | no |
| 21 | button/button | accion | aiChatMinimize | Minimizar chat | no |
| 22 | button/button | accion | closeAIDrawer | &times; | no |
| 23 | button/button | accion | aiChatNewBtn | Nuevo chat | no |
| 24 | button/button | accion | aiChatConvBtn | Modo conversación | no |
| 25 | button/button | accion | aiChatMicBtn | Dictar mensaje | no |
| 26 | button/button | accion | aiChatVoiceBtn | Voz del asistente | no |
| 27 | button/button | accion | aiChatStopBtn | Detener audio y respuesta | no |
| 28 | input/hidden | entrada | aiChatMode | operativo | no |
| 29 | input/file | accion | aiChatAttachment | Adjuntar archivo para IA | no |
| 30 | button/button | accion | aiChatAttachBtn | Adjuntar archivo | no |
| 31 | button/button | accion | aiChatClearAttachment | &times; | no |
| 32 | textarea/- | entrada | aiChatInput | Mensaje al asistente IA | no |
| 33 | button/submit | accion | - | Enviar | no |
| 34 | a/- | accion | whatsappFloat | WA | no |
| 35 | input/number | entrada | - | '+esc(it.qty\|\|1)+' | sí |
| 36 | button/button | accion | - | Quitar | sí |
| 37 | button/button | accion | - | Agregar al carrito | sí |
| 38 | a/- | accion | - | Pagar | no |
| 39 | a/- | accion | - | Mensaje | sí |

### `web/visualizar_certificado_tributario_publico.html` (2)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | button/button | accion | btnDownload | Registrar descarga | no |
| 2 | button/button | accion | - | Imprimir / PDF | sí |

### `web/visualizar_productos_y_precios_publico.html` (10)

| # | Tipo | Clase | ID | Etiqueta | Dinámico |
| ---: | --- | --- | --- | --- | --- |
| 1 | input/- | entrada | search | search | no |
| 2 | select/- | entrada | sort | Relevancia Menor precio Mayor precio Nombre A-Z | no |
| 3 | input/- | entrada | contactName | contactName | no |
| 4 | input/email | entrada | contactEmail | contactEmail | no |
| 5 | input/- | entrada | contactPhone | contactPhone | no |
| 6 | textarea/- | entrada | contactMessage | contactMessage | no |
| 7 | button/submit | accion | contactSubmit | Enviar mensaje | no |
| 8 | a/- | accion | whatsappFloat | WA | no |
| 9 | button/- | accion | - | Todos | sí |
| 10 | button/- | accion | - | ' + escapeHtml(page.nombre \|\| page.slug \|\| 'Seccion') + ' | sí |
