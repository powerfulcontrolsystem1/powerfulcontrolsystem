## [2026-05-07] QA E2E Motel Calipso y ajustes flotantes
- [QA Motel Calipso] Se ejecuta regresion autenticada sobre `empresa_id=7` con 60 modulos de Administrar empresa: escritorio con 60/60 modulos cargados y validacion dirigida posterior sin errores para activos fijos, auditoria, red social, control electrico, venta publica, radio y configuracion.
- [QA profunda] Se agrega runner `backend/tmp_tools/qa_calipso_operativo/deep_flows_calipso.mjs` para crear datos QA reales en parqueadero, WMS, centros de costo, activos fijos, red social con imagen, carta publica, venta publica y validacion QR publica. Resultado final: 6/6 pasos OK, sin errores de consola, red ni pagina.
- [QA movil] Se habilita viewport configurable en `frontend_buttons_calipso.mjs` y se recorren los 60 modulos en 390x844. Los dos hallazgos moviles se corrigen y se validan de forma dirigida en venta publica y asistencia de empleados.
- [UX flotante] El robot/secretaria separa los globos del avatar en movil, se evita que el drawer del asistente quede bajo el boton de favoritos y la radio flotante se compacta como boton circular inferior izquierdo para reducir bloqueos de botones.
- [Robustez frontend] Activos fijos exporta aunque el dashboard aun no haya cargado, auditoria tolera ausencia del rotulo de empresa, red social muestra errores controlados sin romper consola, graficos/estadisticas degradan con aviso visual.
- [Riesgos externos] Queda documentado que impresora fisica, sensores electricos reales, GPS fisico, DIAN y pasarelas en produccion requieren credenciales/dispositivos externos para prueba final fuera del entorno local.

## [2026-05-06] Modulos empresariales Colombia - fases compartidas
- [ERP Colombia] Se implementan `bancos_pagos`, `gestion_documental`, `cumplimiento_kyc`, `contratos_obligaciones`, `helpdesk` y `calidad_procesos` sobre nucleo compartido por `empresa_id`.
- [Backend] APIs privadas por modulo con acciones `dashboard`, `plantilla`, `reporte`, `registros`, `eventos`, `evidencias`, `registro`, `estado`, `evento`, `evidencia`, `importar_registros` y `seed_demo`.
- [UI] Pantallas administrativas compartidas con KPIs, reporte ejecutivo, CSV, seguimiento, cambio de estado y evidencias/soportes.
- [Workflow] Se agrega flujo de aprobaciones por nivel, destinatario, vencimiento, decision y bitacora.
- [Workflow] Se agrega gestion de tareas por registro con responsable, prioridad, vencimiento y estados operativos.
- [Operacion] Se agrega expediente 360 por registro para consolidar eventos, evidencias, aprobaciones, tareas, resumen y recomendacion.
- [Operacion] Se agrega agenda de alertas por modulo con vencidos, proximos vencimientos, tareas, aprobaciones pendientes y acceso al expediente.
- [Gobierno] Se agrega cierre controlado para impedir cierre sin evidencia o con aprobaciones/tareas abiertas.
- [Operacion] Se agrega generador de plan de accion para crear tareas desde alertas de agenda sin duplicar tareas abiertas.
- [Operacion] Se agrega tablero de responsables con carga pendiente, vencidos y recomendaciones por responsable.
- [Operacion] Se agrega tablero SLA con cumplimiento, semaforo, buckets de vencimiento y recomendaciones.
- [Operacion] Se agrega matriz de riesgo operativo con score, nivel, factores ponderados y recomendaciones.
- [Auditoria] Se agrega exportacion CSV multi-seccion con resumen, registros, agenda, SLA, riesgo, responsables, tareas, aprobaciones, evidencias y bitacora.
- [Operacion] Se agrega busqueda avanzada de backend por texto, estado, tipo, categoria, prioridad, responsable, vencidos y proximos vencimientos.
- [Operacion] Se agregan acciones masivas controladas para cambiar estado, prioridad y responsable con bitacora por registro.
- [BD] Tablas compartidas `empresa_modulos_colombia_registros`, `empresa_modulos_colombia_eventos`, `empresa_modulos_colombia_evidencias`, `empresa_modulos_colombia_aprobaciones` y `empresa_modulos_colombia_tareas`.
- [Docs/QA] Documentacion actualizada y validacion con `go test ./... -count=1`.

## [2026-05-06] Portal de Terceros y Certificados Tributarios
- [Finanzas/tributario] Nuevo modulo `portal_terceros_certificados` para proveedores, clientes, empleados, contratistas y terceros externos.
- [Backend] Nueva API administrativa `/api/empresa/portal_terceros_certificados` y API publica `/api/public/certificados_tributarios`.
- [UI] Nueva pantalla administrativa y nueva pagina publica de visualizacion/impresion de certificados por token.
- [Permisos/licencias] Nuevo modulo activable por licencia y controlado por roles financieros/contables.
- [Auditoria] Bitacora de descargas con IP, navegador, canal y fecha.

## [2026-05-06] Activos Fijos e Intangibles NIIF/Fiscal
- [Finanzas/contabilidad] Nuevo modulo formal `activos_fijos_niif_fiscal` para PPE e intangibles por empresa.
- [Backend] Nueva API `/api/empresa/activos_fijos_niif_fiscal` con dashboard, libro, depreciaciones, eventos, registro de activos y datos demo.
- [BD] Se amplio `empresa_contabilidad_activos_fijos` con campos NIIF/fiscales, deterioro, valor razonable, valor fiscal y diferencia NIIF/fiscal.
- [UI] Nueva pantalla `web/administrar_empresa/activos_fijos_niif_fiscal.html` enlazada desde el menu principal y el centro financiero.
- [Permisos/licencias] Nuevo modulo `activos_fijos_niif_fiscal` activable por licencia y roles.

## [2026-05-06] Propiedad horizontal y promocion por asesor
- [Verticales] Nuevo modulo `propiedad_horizontal` para copropiedades, conjuntos, edificios y condominios.
- [Backend] Nueva API `/api/empresa/propiedad_horizontal` con configuracion, unidades, personas, cargos, recaudos, PQR, asambleas, dashboard y datos demo.
- [Permisos] Nueva clave de modulo/licencia `propiedad_horizontal`, pagina `linkPropiedadHorizontal` y wrapper `WithEmpresaPropiedadHorizontalPermissions`.
- [Super] En `Asesor comercial` se agrega promocion activable/desactivable para descuento adicional por codigo de asesor, con porcentaje configurable.
- [Checkout] `pagar_licencia.html`, Wompi, Epayco y activacion sin pago consideran `asesor_id` para aplicar descuento si la promocion esta activa y conservar comisiones.
- [Docs/QA] Se agregan `documentos/propiedad_horizontal.md` y `documentos/promocion_asesor_licencias.md`; pruebas `go test ./... -count=1` en `backend/`.

## [2026-05-06] Cierre y bloqueo fiscal
- [Finanzas/contabilidad] Nuevo modulo `cierre_fiscal` para proteger periodos cerrados, documentos reportados y operaciones post-cierre.
- [Backend] Nueva API `/api/empresa/cierre_fiscal` con dashboard, politicas, periodos, excepciones, validacion, bitacora y datos demo.
- [Base de datos] Nuevas tablas `empresa_cierre_fiscal_politicas`, `empresa_cierre_fiscal_periodos`, `empresa_cierre_fiscal_excepciones` y `empresa_cierre_fiscal_eventos`, todas por `empresa_id`.
- [Integracion] El cierre/reapertura de `contabilidad_colombia` sincroniza el periodo fiscal para evitar doble control de cierres.
- [Permisos] Nueva clave de modulo/licencia `cierre_fiscal`, paginas `linkCierreFiscal`/`linkCierreFiscalMenu` y wrapper `WithEmpresaCierreFiscalPermissions`.
- [Frontend] Nueva pantalla `web/administrar_empresa/cierre_fiscal.html` dentro del Centro financiero y contable.
- [Docs/QA] Se crea `documentos/cierre_fiscal.md`; pruebas `go test ./... -count=1` en `backend/`.

## [2026-05-06] Centros de costo y rentabilidad
- [Finanzas/contabilidad] Nuevo modulo formal `centros_costo` para medir rentabilidad por sucursal, area, unidad de negocio o proyecto.
- [Backend] Nueva API `/api/empresa/centros_costo` con dashboard, centros, reglas, presupuestos, movimientos integrados y datos demo.
- [Base de datos] Nuevas tablas `empresa_centros_costo`, `empresa_centros_costo_reglas` y `empresa_centros_costo_presupuestos`, aisladas por `empresa_id`.
- [Integracion] El dashboard consolida movimientos existentes con `centro_costo` desde contabilidad Colombia, tesoreria, compras avanzadas, soportes OCR/IA y AIU construccion, sin duplicar modulos financieros.
- [Permisos] Nueva clave de modulo/licencia `centros_costo`, paginas `linkCentrosCosto`/`linkCentrosCostoMenu` y wrapper `WithEmpresaCentrosCostoPermissions`.
- [Frontend] Nueva pantalla `web/administrar_empresa/centros_costo.html` dentro del Centro financiero y contable, con modo claro/oscuro, exportacion CSV y formularios CRUD.
- [Docs/QA] Se crea `documentos/centros_costo.md`; pruebas `go test ./... -count=1` en `backend/`.

## [2026-05-06] QA transversal y profesionalizacion de modulos
- [Portal publico] `web/index.html` actualiza la descripcion de modulos para incluir Cobranza, Portal contador, Captura IA/OCR de compras y gastos, AIU construccion, Parqueaderos con ticket QR y Apartamentos turisticos en la seccion publica y tarjetas fallback.
- [Permisos] `web/js/administrar_empresa.js` reconoce `administrador_total` como rol de acceso total en la evaluacion local del menu, alineado con backend.
- [Rendimiento] Los dashboards de `cobranza`, `portal_contador` y `soportes_compras_ia` evitan validaciones repetidas de esquema dentro de la misma peticion.
- [Frontend] La pagina `soportes_compras_ia.html` valida enlaces dinamicos de archivos antes de renderizarlos.
- [QA Motel Calipso] Login super administrador, paginas principales y APIs de modulos recientes verificadas con HTTP 200 para `empresa_id=7`.
- [Auditoria] Revision estatica de enlaces, botones `onclick` e IDs de paginas empresariales; sin botones muertos relevantes.
- [Docs] Se agrega `documentos/reporte_qa_modulos_2026-05-06.md` con alcance, pruebas, observaciones y estado final.

## [2026-05-06] Portal contador
- [Finanzas/contabilidad] Nuevo modulo `portal_contador` como oficina virtual para contadores y firmas contables.
- [Backend] Nueva API `/api/empresa/portal_contador` con dashboard, clientes, obligaciones, solicitudes, comunicaciones y datos demo.
- [Base de datos] Nuevas tablas `empresa_portal_contador_clientes`, `empresa_portal_contador_obligaciones`, `empresa_portal_contador_solicitudes` y `empresa_portal_contador_comunicaciones`, todas aisladas por `empresa_id`.
- [Permisos] Nueva clave de modulo/licencia `portal_contador`, paginas `linkPortalContador`/`linkPortalContadorMenu` y wrapper `WithEmpresaPortalContadorPermissions`.
- [Frontend] Nueva pantalla `web/administrar_empresa/portal_contador.html` dentro del Centro financiero y contable.
- [Docs/QA] Se crea `documentos/portal_contador.md`; pruebas `go test ./db -run TestPortalContador -count=1` y `go test ./... -count=1`.

## [2026-05-06] Gestion de cobranza
- [Finanzas] Nuevo modulo profesional `cobranza` para recuperar cartera sin duplicar cuentas por cobrar.
- [Backend] Nueva API `/api/empresa/cobranza` con dashboard, cuentas, plantillas, campanas, gestiones, promesas, simulacion de envio y datos demo.
- [Base de datos] Nuevas tablas `empresa_cobranza_plantillas`, `empresa_cobranza_campanas`, `empresa_cobranza_gestiones` y `empresa_cobranza_promesas`, enlazadas por `empresa_id` y `cuenta_id`.
- [Permisos] Nueva clave de modulo/licencia `cobranza`, paginas `linkCobranza`/`linkCobranzaMenu` y wrapper `WithEmpresaCobranzaPermissions`.
- [Frontend] Nueva pantalla `web/administrar_empresa/cobranza.html` dentro del Centro financiero y contable.
- [Docs/QA] Se crea `documentos/cobranza.md`; pruebas `go test ./db -run TestCobranza -count=1` y `go test ./... -count=1`.

## [2026-05-06] AIU construccion
- [Mejora profesional] El modulo AIU ahora incluye responsable, centro de costo, modalidad contractual, riesgo, avance, retenciones, anticipo, garantia, neto a cobrar y flujo validado de estados.
- [Backend] Nuevas acciones `facturas`, `reporte` y `estado`; el dashboard suma contratos/facturas, neto, retenciones, pendiente por facturar y alertas.
- [Frontend] Pantalla reorganizada con KPIs, filtros, acciones de estado, resumen financiero, facturas recientes y exportacion CSV.
- [Facturacion/obra] Nuevo modulo `aiu_construccion` para arquitectos, constructoras, contratistas y pequenas empresas de obra.
- [Backend] Nueva API `/api/empresa/aiu_construccion` con dashboard, contratos, conceptos, calculadora AIU, datos demo y generacion de factura electronica AIU.
- [Base de datos] Nuevas tablas `empresa_aiu_contratos`, `empresa_aiu_items` y `empresa_aiu_facturas`, todas aisladas por `empresa_id`.
- [Permisos] Nueva clave de licencia/rol `aiu_construccion`, pagina `linkAIUConstruccion` y wrapper `WithEmpresaAIUConstruccionPermissions`.
- [Frontend] Nueva pantalla `web/administrar_empresa/aiu_construccion.html` enlazada desde el submenu de facturacion electronica.
- [Docs/QA] Se crea `documentos/aiu_construccion.md`; pruebas `go test ./db -run Test.*AIU -count=1` y `go test ./...`.

## [2026-05-06] Documentos electronicos DIAN/Siigo
- [Facturacion electronica] Se amplia el ciclo documental existente para cubrir factura electronica, nota credito, nota debito, documento soporte, nomina electronica y documento equivalente POS electronico por empresa.
- [Backend] `/api/empresa/facturacion_electronica` normaliza aliases Siigo/DIAN, valida tipos soportados, conserva auditoria/eventos contables y usa la misma cola DIAN/proveedor para los nuevos documentos.
- [Frontend] `web/administrar_empresa/facturacion_electronica.html` agrega selector de tipo documental y botones rapidos para emitir factura, notas, soporte, nomina y POS electronico.
- [Docs/QA] Verificado con pruebas unitarias enfocadas en normalizacion y transiciones documentales, mas prueba de defaults de facturacion por pais.

## [2026-05-06] CRM y ventas avanzadas
- [CRM] Se amplia `clientes`/CRM comercial sin duplicar modulo con metas, forecast, scoring, agenda y conversion de lead a cotizacion.
- [Backend] Nueva API `/api/empresa/crm_avanzado` con dashboard, metas, scores, demo y cotizacion desde lead.
- [Base de datos] Nueva tabla `empresa_crm_metas_comerciales`; el dashboard reutiliza `crm_leads`, `crm_interacciones`, `empresa_cotizaciones_venta` y `empresa_pedidos_venta`.
- [Permisos] Se reutiliza el modulo/licencia `clientes`, con pagina `linkCRMAvanzado`.
- [Frontend] Nueva pantalla `web/administrar_empresa/crm_ventas_avanzadas.html` en Ventas, clientes y caja.
- [Docs/QA] Se crea `documentos/crm_ventas_avanzadas.md` y el QA Calipso valida lead, meta, scoring, cotizacion y forecast.

## [2026-05-06] Inventario avanzado
- [Inventario] Se amplia el modulo existente `inventario` sin duplicarlo con lotes, seriales, reservas, vencimientos y valorizacion por bodega.
- [Backend] Nueva API `/api/empresa/inventario_avanzado` con dashboard, lotes, seriales, reservas, valorizacion, demo y confirmacion de salida.
- [Base de datos] Nuevas tablas `empresa_inventario_lotes_avanzados`, `empresa_inventario_seriales_avanzados` y `empresa_inventario_reservas_avanzadas`, todas aisladas por `empresa_id`.
- [Integracion] La entrada de lote actualiza `inventario_existencias` y `inventario_movimientos`, manteniendo el kardex existente como eje operativo.
- [Permisos] Se reutiliza el modulo/licencia `inventario`, con pagina `linkInventarioAvanzado`.
- [Frontend] Nueva pantalla `web/administrar_empresa/inventario_avanzado.html` enlazada desde el menu de Productos.
- [Docs/QA] Se crea `documentos/inventario_avanzado.md` y el QA Calipso valida lote, serial, reserva, salida y valorizacion.

## [2026-05-06] Compras avanzadas
- [Compras] Se amplia el modulo existente `compras` sin duplicarlo con requisiciones internas, cotizaciones, aprobaciones y recepcion parcial/total por empresa.
- [Backend] Nueva API `/api/empresa/compras_avanzadas` con dashboard, requisiciones, detalle, cotizaciones, aprobaciones, recepciones y datos demo.
- [Base de datos] Nuevas tablas `empresa_compras_requisiciones`, `empresa_compras_requisicion_items`, `empresa_compras_cotizaciones`, `empresa_compras_aprobaciones`, `empresa_compras_recepciones_avanzadas` y `empresa_compras_recepcion_items_avanzadas`.
- [Permisos] Se reutiliza el modulo/licencia `compras`, con pagina `linkComprasAvanzadas` dentro del submenu de Compras.
- [Frontend] Nueva pantalla `web/administrar_empresa/compras_avanzadas.html` enlazada desde `compras_menu.html`.
- [Docs/QA] Se crea `documentos/compras_avanzadas.md` y el QA Calipso valida requisicion, cotizacion, aprobacion, recepcion y dashboard.

## [2026-05-06] Importaciones y costeo de nacionalizacion
- [Compras] Se agrega el modulo `importaciones_costeo` para compras internacionales, incoterms, TRM, items importados, fletes, aranceles, aduana y costo aterrizado.
- [Backend] Nueva API `/api/empresa/importaciones_costeo` con dashboard, importaciones, detalle, items, costos, distribucion y datos demo por empresa.
- [Base de datos] Nuevas tablas `empresa_importaciones_costeo`, `empresa_importaciones_costeo_items` y `empresa_importaciones_costeo_costos`, todas aisladas por `empresa_id`.
- [Permisos] Nueva clave de licencia `importaciones_costeo`, pagina `linkImportacionesCosteo` y wrapper `WithEmpresaImportacionesCosteoPermissions`.
- [Frontend] Nueva pantalla `web/administrar_empresa/importaciones_costeo.html` dentro de Inventario y compras.
- [Docs/QA] Se crea `documentos/importaciones_costeo.md` y el QA Calipso valida embarque, items, costos, distribucion y dashboard.

## [2026-05-06] Activos fijos avanzado
- [Contabilidad] Se amplia `contabilidad_colombia_avanzada` con activos fijos avanzados sin duplicar modulo: depreciacion por periodo, eventos, mantenimiento, traslados y bajas.
- [Backend] La API `/api/empresa/contabilidad_colombia_avanzada` agrega `activos_resumen`, `activos_depreciaciones`, `activos_eventos`, `generar_depreciacion_activos` y `activo_evento`.
- [Base de datos] Se enriquecen activos fijos con serial, placa, metodo de depreciacion, centro de costo, proveedor, poliza y mantenimiento programado; se agregan `empresa_contabilidad_activos_depreciacion` y `empresa_contabilidad_activos_eventos`.
- [Frontend] La suite contable agrega pestaña `Activos avanzado` para generar depreciacion, registrar eventos y consultar inventario gerencial.
- [Docs/QA] Se crea `documentos/activos_fijos_avanzado.md` y el QA Calipso valida activo, depreciacion, mantenimiento y resumen avanzado.

## [2026-05-06] Nomina Colombia avanzada
- [Nomina] Se amplia el modulo existente `nomina_sueldos` sin duplicarlo: conceptos legales Colombia, novedades aprobables y resumen PILA por empresa.
- [Backend] La API `/api/empresa/nomina` agrega acciones `conceptos_colombia`, `novedades_colombia`, `pila_colombia`, `dashboard_colombia`, `concepto_colombia`, `novedad_colombia`, `generar_pila` y `seed_colombia`.
- [Base de datos] Nuevas tablas aisladas por `empresa_id`: `empresa_nomina_colombia_conceptos`, `empresa_nomina_colombia_novedades` y `empresa_nomina_colombia_pila_resumen`.
- [Frontend] La pantalla de nomina incluye una seccion `Nomina Colombia avanzada` con KPIs, conceptos, novedades y PILA.
- [Docs/QA] Se crea `documentos/nomina_colombia_avanzada.md` y el QA Calipso registra empleado, concepto, novedad, liquidacion y PILA.

## [2026-05-06] Tesoreria y presupuesto
- [Tesoreria] Se agrega el modulo `tesoreria_presupuesto` para cuentas banco/caja, presupuestos, partidas, ejecucion y flujo de caja proyectado por empresa.
- [Backend] Nueva API `/api/empresa/tesoreria_presupuesto`, tablas aisladas por `empresa_id` y dashboard con saldo disponible, ingresos/egresos proyectados y flujo neto.
- [Permisos] Nueva clave de licencia `tesoreria_presupuesto`, pagina `linkTesoreriaPresupuesto` y wrapper `WithEmpresaTesoreriaPresupuestoPermissions`, alineado con roles financieros/contables.
- [Frontend] Nueva pantalla `web/administrar_empresa/tesoreria_presupuesto.html` dentro del Centro financiero y contable.
- [Docs/QA] Se crea `documentos/tesoreria_presupuesto.md` y se amplian pruebas de normalizacion y QA Calipso.

## [2026-05-06] Produccion / MRP empresarial
- [Produccion] Se agrega el modulo `produccion_mrp` para recetas/BOM, componentes, ordenes de produccion, consumos, control de calidad, plan MRP y datos demo por empresa.
- [Backend] Nueva API `/api/empresa/produccion_mrp`, tablas aisladas por `empresa_id` y dashboard operativo con KPIs de ordenes, calidad y costos.
- [Permisos] Nueva clave de licencia `produccion_mrp`, pagina `linkProduccionMRP` y wrapper `WithEmpresaProduccionMRPPermissions`; roles de inventario/compras pueden operar segun matriz.
- [Frontend] Nueva pantalla `web/administrar_empresa/produccion_mrp.html` bajo Inventario y compras, con tabs de recetas, ordenes, consumos/calidad, MRP y configuracion.
- [Docs/QA] Se crea `documentos/produccion_mrp.md` y pruebas unitarias de normalizacion del modulo.

## [2026-05-05] Reportes colombianos avanzados
- [Reportes] Se agregan datasets profesionales que suelen exigir los sistemas contables/POS usados en Colombia: ventas diarias por medio de pago, rentabilidad por producto, Kardex valorizado, compras detalladas por proveedor, balance de prueba, libro auxiliar, libro mayor, impuestos/retenciones, informacion exogena base, edades de cartera CxC y edades CxP.
- [Backend] Los reportes quedan en el endpoint existente `/api/empresa/reportes`, reutilizan filtros por rango, exportacion `JSON/CSV/TXT/XLS/PDF`, programacion/plantillas y separacion estricta por `empresa_id`.
- [Frontend] El menu de reportes agrega accesos directos a los nuevos datasets sin duplicar modulos ni romper los enlaces anteriores.
- [QA] Se agrega prueba de catalogo para evitar datasets duplicados y asegurar metadatos/formats completos.
- [Administrar empresa] El menu principal queda reorganizado en categorias plegables; se separan `Operacion por tipo` y `Ventas, clientes y caja` para reducir botones visibles sin cambiar rutas ni permisos.

## [2026-05-05] Organizacion de modulos y control super
- [Administrar empresa] Se fusiona la navegacion repetida de Finanzas, Contabilidad Colombia y Suite contable Colombia bajo `Centro financiero y contable`, manteniendo rutas internas compatibles.
- [Ayuda] La ayuda administrativa principal `/ayuda/ayuda.html` queda protegida para `super_administrador`; las ayudas publicas especificas se conservan.
- [Super] Se agrega el rol `control_super_administrador` para supervision limitada de administradores, seguridad, errores, metricas y reportes globales.
- [Seguridad] El contralor super no puede eliminar ni desactivar el super administrador principal ni administrar otros contralores super.

## [2026-05-06] Captura inteligente de compras y gastos con OCR/IA
- [Compras] Se agrega el modulo `soportes_compras_ia` para radicar soportes con foto, PDF o XML, detectar duplicados y gestionar estados de revision, aprobacion y contabilizacion.
- [IA] La extraccion usa la capa existente de IA con modelo recomendado `openai:gpt-5.5`, prompt contable colombiano y registro en historial de consultas IA.
- [Backend] Nueva API protegida `/api/empresa/soportes_compras_ia`, tablas `empresa_soportes_compras_ia` y `empresa_soportes_compras_ia_eventos`, guardado de archivos por empresa y conversion a `empresa_cuentas_por_pagar`.
- [Frontend] Nueva pantalla `web/administrar_empresa/soportes_compras_ia.html` enlazada desde el menu de Compras con dashboard, carga de archivos, acciones GPT-5.5, aprobacion, rechazo, contabilizacion y exportacion CSV.
- [Permisos] Nuevo modulo de rol/licencia `soportes_compras_ia` con acceso operativo para compras, contabilidad, supervisor y administrador de empresa.
- [Docs/QA] Se crea `documentos/soportes_compras_ia.md`; validado con `go test ./db -run Test.*Soporte.*IA -count=1`, `go test ./... -count=1` y `git diff --check`.

## [2026-05-05] Suite contable Colombia avanzada
- [Contabilidad] Se agrega `contabilidad_colombia_avanzada` con informacion exogena DIAN/medios magneticos, nomina electronica, documento soporte, activos fijos, cartera/CxP y libros oficiales por empresa.
- [Backend] Nueva API `/api/empresa/contabilidad_colombia_avanzada`, tablas empresariales aisladas por `empresa_id` y generacion de exogena/libros desde comprobantes contabilizados del nucleo `contabilidad_colombia`.
- [Permisos] Nuevo modulo de licencia `contabilidad_colombia_avanzada`, pagina `linkContabilidadColombiaAvanzada` y wrapper `WithEmpresaContabilidadColombiaAvanzadaPermissions`.
- [Frontend] Nueva vista `web/administrar_empresa/contabilidad_colombia_avanzada.html` con dashboard y pestañas profesionales para cada submodulo.
- [Docs/QA] Se crea `documentos/contabilidad_colombia_avanzada.md`; pruebas Go y auditoria de rutas/permisos actualizadas.

## [2026-05-05] Portal publico, carta QR y Motel Calipso publicado
- [Permisos] Se audita el enlace por empresa de todos los modulos visibles en Administrar empresa: menu frontend, catalogo de paginas backend y licencias quedan alineados, sin claves duplicadas ni rutas `/api/empresa` duplicadas.
- [Carnets] Se agrega modulo empresarial profesional `/api/empresa/carnets` y `web/administrar_empresa/carnets.html` para emitir carnets modernos de empleados/usuarios con plantillas, QR, foto, exportacion PNG/SVG, impresion y bitacora.
- [Licencias] Se agrega modulo `carnets`, pagina `linkCarnets` y soporte en `licencias.modulos_habilitados` para activarlo/desactivarlo por empresa.
- [Aislamiento] Se rectifica el control multiempresa en wrappers `WithEmpresa*`: query, cabecera `X-Empresa-ID`, formulario/multipart y JSON no pueden declarar empresas distintas para una misma peticion.
- [Seguridad] Las rutas privadas `/api/empresa/...` registradas en `main.go`, chat IA empresarial y modulos ERP faltantes quedan revisadas bajo wrappers de empresa.
- [Portal] Se actualizan las descripciones de modulos en `web/index.html` para reflejar el alcance real del sistema: POS, estaciones, hotel/motel, gimnasio, odontologia, domicilios tipo Rappi, Taxi System tipo Uber, turnos, control electrico, inventario, finanzas, facturacion, usuarios/permisos/licencias, IA, reportes, carta publica QR, red social y hoja de vida.
- [Carta publica] `visualizar_productos_y_precios_publico.html` queda documentada y habilitada como pagina publica de solo lectura, directa o bajo `/{empresa_slug}/visualizar_productos_y_precios_publico.html`, con QR exportable desde Administrar empresa.
- [Motel Calipso] Se documenta la publicacion real del slug `motel-calipso` con venta publica, carta publica, paginas, items de ejemplo y publicaciones en red social comercial.
- [Seguridad] `AuthMiddleware` permite la carta publica sin sesion; la administracion conserva control por modulo `venta_publica`, rol, licencia y pagina `linkCartaProductosPublica`.
- [Verificacion] `go test ./utils`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`; validacion productiva HTTP 200 de venta publica, carta publica y red social.

## [2026-05-05] Roles y licencias para modulos verticales
- [Permisos] Se agregan modulos independientes para venta publica/carta, gimnasio, taxi system, domicilios, alquileres, odontologia, turnos de atencion y control electrico.
- [Licencias] La pantalla de licencias permite activar/desactivar estos modulos desde `modulos_habilitados`, con presets actualizados.
- [Backend] Los endpoints administrativos verticales usan wrappers dedicados para que licencia, rol y pagina del menu bloqueen con `403` cuando corresponda.
- [Docs] Se actualiza la matriz de roles/permisos y la documentacion de domicilios.

## [2026-05-05] Modulo profesional de domicilios
- [Domicilios] Se agrega modulo tipo marketplace para restaurantes, domiciliarios, cliente publico y central administrativa con pedidos, menu, tarifas, comisiones, autoasignacion y estados operativos.
- [Tracking] Los domiciliarios reportan ubicacion GPS desde el navegador movil; la central visualiza restaurantes, clientes y domiciliarios en mapa Leaflet.
- [Seguridad] Cada pedido genera codigo de entrega y token de seguimiento para el cliente; restaurantes y domiciliarios usan PIN operativo.
- [Docs] Se crea `documentos/domicilios_profesional.md` con superficies, flujo, endpoints, datos demo y recomendaciones de produccion.

## [2026-05-05] Taxi System profesional con mapa y GPS
- [Taxi System] El panel administrativo incorpora operacion tipo Uber con mapa, filtros por conductores disponibles/ocupados, solicitudes y GPS externos.
- [GPS] Se integra el inventario corporativo de dispositivos GPS dentro de Taxi System para registrar apps moviles, trackers, OBD2, celulares, tablets, dashcams y webhooks con protocolo/proveedor.
- [Conductores] Se agregan campos de asociacion GPS en `empresa_taxi_drivers`: `gps_dispositivo_id`, `gps_codigo`, `gps_tipo`, `gps_proveedor` y `gps_protocolo`.
- [Docs] Se crea `documentos/taxi_system_profesional.md` con alcance, superficies, endpoints y validacion.

## [2026-05-04] Control electrico Raspberry Pi por estacion
- [Control electrico] Nuevo modulo en Administrar empresa para configurar Raspberry Pi, IP/puerto/ruta API, token opcional, timeout y sincronizacion automatica.
- [Estaciones] Cada estacion puede mapearse a multiples relés GPIO con salida/carga (luces, jacuzzi, aire, puerta u otro), nombre, pin, logica activo alto, pulso opcional y prueba manual ON/OFF.
- [Carrito] El carrito de estacion incorpora boton `Control electrico` para abrir un panel operativo y controlar manualmente salidas de la habitacion sin salir de la venta.
- [Automatizacion] Al activar/recuperar/reabrir una estacion se envia `on`; al pagar/cerrar/desactivar se envia `off`. Tambien se engancha con autoactivacion por sensor de puertas.
- [Auditoria] Se agrega bitacora electrica por empresa con comando, estado objetivo, GPIO, HTTP status, respuesta/error, actor, origen y fecha.
- [Backend] Se registran tablas `empresa_control_electrico_config`, `empresa_control_electrico_reles` y `empresa_control_electrico_eventos`, mas endpoint protegido `/api/empresa/control_electrico`.

## [2026-05-02] Gimnasio, impresoras y documentacion del proyecto
- [Gimnasio] Se robustece el esquema del modulo con migraciones defensivas para tablas antiguas de empresas, evitando errores internos al abrir dashboard, acceso, credenciales o dispositivos cuando faltaban columnas historicas.
- [Gimnasio] Se agrega preconfiguracion operativa propia del modulo: sede principal, RFID/NFC/QR, planes base, clases iniciales y dispositivos de acceso, todo aplicable desde el dashboard del gimnasio.
- [Impresoras] Se corrige el guardado de configuracion avanzada para que `modo_documento_venta` gobierne correctamente la activacion o desactivacion de facturacion electronica automatica.
- [Impresoras] Se incorpora `cajon_monedero` como funcionalidad asignable de impresora dentro de `Configuracion > Impresora`, alineando la UI con la operacion real de caja.
- [Docs] Se actualiza `RESUMEN_DEL_PROYECTO.md` para reflejar configuracion guiada por IA, impresion empresarial, horarios laborales y modulos verticales ya integrados como gimnasio, odontologia, taxi system, turnos de atencion y alquileres.

## [2026-04-30] Pagos, chat IA, empresas compartidas, hoja de vida operativa y documentos dinamicos
- [Pagos/Epayco] Smart Checkout v2 conserva fallback clasico firmado por POST a `https://secure.payco.co/checkout.php`; se elimina la redireccion GET que producia XML `AccessDenied` y se documenta el requisito de `epayco.customer_id` para fallback.
- [Pagos/Epayco] El fallback clasico resuelve su modo con `epayco.customer_id` + `epayco.checkout_key`/`epayco.p_key`, separado de las llaves Smart Checkout, para no enviar cuentas reales como pruebas y evitar el error "El comercio no fue reconocido".
- [Chat IA] La secretaria IA 3D se rediseña como avatar estilo caricatura ejecutiva joven y habla siempre con voz femenina (`es-CO-female`), manteniendo el robot con voz configurable.
- [Empresas compartidas] El editor de empresa permite consultar y retirar administradores compartidos desde ambos lados del acceso, con trazabilidad del actor.
- [Administrar empresa] Se implementa la hoja de vida operativa universal para motos de taller, pacientes, vehiculos, equipos, activos o mascotas, con ficha, eventos, servicios, alertas y resumen operativo.
- [Documentos IA] Se documenta el flujo `/generate` + `/download` para generar documentos dinamicos con IA/templates y exportar PDF, DOCX, XLSX, HTML, TXT o JSON.
- Nueva funcionalidad: MÃ³dulo Red Social Comercial con portal pÃºblico y administraciÃ³n por empresa. EliminaciÃ³n de modulo juegos y venta de licencias desde cliente.

## [2026-04-23] Retiro Tipos de usuario (panel super)
- [Super/DB] Eliminación del módulo Tipos de usuario: sin API ni UI; tabla `tipos_de_usuario` removida al arranque; documentación alineada.

## [2026-04-23] reCAPTCHA, backup y manual de instalación
- [Docs/Operación] Se actualizó el manual de instalación con reCAPTCHA v2/v3/Enterprise, variables, panel super y fallos frecuentes (dominios, tipo de clave). Se documentan las claves y copias best-effort en `backup/super_administrador` y `backup/empresas/<empresa_id>`. Ajustes en `descripcion_de_archivos` e `historial_de_cambios` y alineación con `CHANGELOG.md` raíz.

## [2026-04-20] Limpieza Total Themes
- [UI/Temas] Auditoría y barrido de más de 50 páginas y scripts en web/administrar_empresa, web/super y páginas públicas para limpiar colores fijos, migrando lógicas JS a .classList.add('text-danger') y respetando las 6 paletas dinámicas. Completado barrido masivo de vistas.
- **2026-04-30 - Pagos ePayco de licencias**: el fallback estandar ahora usa `checkout.js` con `external: "true"` y `PUBLIC_KEY`; se evita el POST legacy a `secure.payco.co/checkout.php`, `P_KEY` queda solo en backend para validacion de webhooks con SHA256 y el frontend `pagar_licencia.html` soporta `checkout_type=classic_js`. Verificacion: `go test ./handlers -run Test.*Epayco -count=1` y `go test ./... -count=1`.

## [2026-05-03] Documentacion, ayuda y estado operativo de modulos
- [Docs] Se crea `documentos/reporte_estado_modulos_2026-05-03.md` con estado compacto por modulo, observaciones de calidad y dependencias pendientes de certificacion.
- [Ayuda] Se actualiza `web/ayuda/ayuda.html` con una seccion de estado operativo, estaciones/carrito, tarjetas adaptables, indicadores del panel y limites honestos de validacion.
- [Operacion] Se documentan los cambios recientes: carrito desde estacion, pago con retorno a estaciones, `USD / COP` primero y despliegue VPS correcto.
- **Logistica avanzada / WMS**: nuevo modulo `logistica_wms` con ubicaciones internas, ordenes WMS, picking, packing, despachos, rutas, bitacora, permisos/licencia, pantalla administrativa y documentacion. Verificacion prevista: pruebas unitarias del motor WMS, `go test ./... -count=1` y `git diff --check`.

- **Declaraciones Tributarias y Motor de Impuestos Colombia**: nuevo modulo `declaraciones_tributarias` con API privada, dashboard, preliquidacion, calendario editable, saldos a pagar/favor, movimientos de conciliacion, permisos/licencia, pantalla administrativa y documentacion. Verificacion prevista: pruebas unitarias del motor, `go test ./... -count=1` y `git diff --check`.
- 2026-05-06: implementados modulos empresariales Colombia `bancos_pagos`, `gestion_documental`, `cumplimiento_kyc`, `contratos_obligaciones`, `helpdesk` y `calidad_procesos` con nucleo compartido, APIs privadas por empresa, paginas administrativas, permisos/licencias y documentacion.
