## [2026-05-19] Caja y turno por usuario
- [Backend] Cierres de caja, aperturas automaticas, pagos, abonos e ingresos/egresos se resuelven por usuario autenticado dentro de la empresa.
- [Base de datos] `empresa_cierres_caja` usa unicidad por `empresa_id`, sucursal, caja, fecha, turno y `usuario_creador`; se eliminan indices legacy sin usuario y se agrega indice para localizar cajas abiertas por usuario.
- [Operacion] Dos usuarios de la misma empresa pueden trabajar al mismo tiempo sin mezclar caja, turno ni reporte; el mismo usuario reutiliza su caja abierta si intenta abrirla de nuevo.
- [Flujo] El cajero puede abrir/reutilizar caja desde `Corte de Caja`; si cobra desde carrito sin caja abierta, el sistema intenta abrir una automaticamente para ese usuario. El cierre se hace al generar/revisar el reporte de turno, imprimir si aplica y cerrar caja/turno.
- [Estaciones] El estado de estaciones queda comun por empresa y el tablero se refresca cada 3 segundos para que varias cajas vean cambios de ocupacion/disponibilidad.
- [Alcance] No cambia permisos, rutas ni dependencias.

## [2026-05-19] Documentos imprimibles blanco y negro
- [UX] Facturas, ventas, notas, documentos imprimibles y reportes de turno/corte ya no heredan la apariencia clara u oscura del panel.
- [Impresion] Las vistas previas y ventanas de impresion usan hoja blanca, texto negro/gris, bordes simples y sin sombras para verse como saldran en papel o POS.
- [Logos] Se respeta la configuracion existente de logo empresarial y logo del sistema; ventas y facturas esperan la configuracion avanzada antes de pintar la primera vista previa y en documentos imprimibles se renderizan en escala de grises.
- [Alcance] No cambia backend, endpoints, permisos ni tablas.

## [2026-05-19] Reporte de turno POS legible
- [UX] El resumen financiero en POS 80mm muestra importes monetarios en una sola linea cuando caben en el rollo.
- [Impresion] Aplica al reporte actual de corte y al historico de reportes de turnos.
- [Alcance] No cambia backend, endpoints, permisos ni tablas.

## [2026-05-19] Prueba operativa Motel Calipso
- [Backend] Pago y anulacion de carritos usan helpers transaccionales compatibles con PostgreSQL en las rutas tocadas por cierre, anulacion, restauracion de inventario y descuentos.
- [Reportes] El pago persiste `descuento_total`, dejando correcto el total de descuentos en el reporte de turno.
- [Alcance] No cambia tablas, rutas, permisos ni dependencias.

## [2026-05-19] Login de usuarios operativos
- [Backend] Se corrige `UpsertAdminEmpresaCompartidaAcceso` para insertar `estado` y `observaciones` junto con el acceso empresarial creado al iniciar sesion.
- [Operacion] El login de cajeros y otros usuarios de empresa ya no falla por SQL de columnas/valores al asegurar el acceso compartido.
- [Alcance] No cambia tablas, rutas, permisos efectivos ni dependencias.

## [2026-05-19] Reporte de turno profesional compacto
- [UX] El reporte de turno actual e historico queda como informe operativo profesional: compacto, completo, con datos ordenados fila a fila y sin cuadriculas.
- [Impresion] Se mantiene compatibilidad con POS 80mm y papel grande.
- [Ventas] Las ventas y facturas electronicas conservan el estilo visual de factura electronica en POS y carta.
- [Alcance] No cambia backend, permisos ni endpoints.

## [2026-05-19] Titulo compacto del login
- [UX] `Powerful Control System` queda mas compacto, en una sola fila y con fuente mas cuadrada.
- [Alcance] No cambia autenticacion, permisos ni backend.

## [2026-05-19] Instalar app desde login
- [UX] El login agrega el boton `Instalar app` junto al acceso de inicio.
- [PWA] Se agregan manifest, service worker e iconos para que Chrome/Edge/Android puedan instalar la app y crear acceso.
- [Compatibilidad] En iOS o navegadores sin prompt, el boton muestra la indicacion para usar el menu del navegador.
- [Alcance] No cambia autenticacion, permisos ni backend.

## [2026-05-19] Titulo del login
- [UX] `Powerful Control System` queda mas grande en el login de administradores, con ajuste responsive para movil.
- [Alcance] No cambia autenticacion, permisos ni endpoints.

## [2026-05-19] Botones de corte de caja
- [UX] Los botones de `Lectura rapida` en `corte_de_caja.html` quedan con color uniforme y mas claro que el panel.
- [Alcance] No cambia backend, permisos, endpoints ni tablas.

## [2026-05-19] Cajas multiples por empresa
- [Configuracion] `empresa_configuracion_general` permite activar/desactivar cajas simultaneas y definir un limite interno por empresa.
- [Backend] Abrir, reabrir y abrir caja automatica para cobro validan caja activa, cupo por empresa, cupo de licencia y cajas abiertas.
- [Reportes] `Ver reporte de mi turno` mantiene caja, turno, sucursal y `cierre_caja_id` para no mezclar reportes entre cajas.

## [2026-05-19] Impresora POS default global
- [Backend] `empresa_impresoras` centraliza helpers para asegurar `POS_80MM` como impresora activa y predeterminada por empresa.
- [Operacion] `set_pos80_config -all` aplica el default a todas las empresas activas, asignando `general`, `corte_caja`, `turno_reporte` y `cajon_monedero`.
- [Alta de empresas] Las empresas nuevas intentan quedar preparadas con POS 80mm al crearse.

## [2026-05-18] POS 80mm para reporte de turno
- [Configuracion] El formato predeterminado del reporte de corte pasa a `pos` para empresas sin configuracion previa o restauradas.
- [Frontend] Corte de caja, configuracion y reportes historicos priorizan `Ticket POS 80mm`; historicos consultan la configuracion por `empresa_id` y bloquean carta/grande cuando la empresa esta en POS.
- [Operacion] Se agrega `backend/tools/set_pos80_config` para activar una impresora POS 80mm predeterminada y asignarla a funcionalidades de caja por empresa.

## [2026-05-18] Contraste de efectivo esperado en caja
- [UX] `corte_de_caja.html` corrige el bloque final `Efectivo esperado en caja` para evitar fondo blanco con texto blanco en modo oscuro.
- [Alcance] No cambia backend, permisos ni tablas.

## [2026-05-18] Contraste del reporte de turno en modo oscuro
- [UX] `corte_de_caja.html` y `reportes_turnos.html` corrigen textos oscuros del reporte cuando el panel esta en modo oscuro.
- [Impresion] La salida impresa conserva fondo blanco y texto oscuro.
- [Alcance] No cambia backend, permisos ni tablas.

## [2026-05-18] Cerrar turno, imprimir y cerrar sesion
- [Frontend] `corte_de_caja.html` agrega el boton `Cerrar turno e imprimir reporte`.
- [Operacion] La accion guarda el cierre, imprime el reporte y cierra sesion con `/auth/logout` al finalizar la impresion.
- [Alcance] No cambia backend, permisos ni tablas.

## [2026-05-18] Descuentos en reporte de turno
- [Backend] El resumen de corte acumula `descuentos_total` y `descuentos_cantidad` desde ventas cerradas.
- [Configuracion] `empresa_corte_caja_configuracion` agrega `mostrar_total_descuentos` activo por defecto.
- [Frontend] El reporte actual y los historicos muestran `Total descuentos`; las exportaciones incluyen la fila de resumen.

## [2026-05-18] Cantidad de ventas en reporte de turno
- [Backend] `empresa_corte_caja_configuracion` agrega `mostrar_cantidad_ventas` activo por defecto.
- [Frontend] El reporte de turno actual y los reportes historicos muestran `Cantidad de ventas` en el resumen.
- [Reportes] Las exportaciones historicas incluyen una fila de resumen con el conteo de ventas del turno.

## [2026-05-18] Tema del reporte de mi turno
- [UX] `corte_de_caja.html` corrige la vista previa de `Ver reporte de mi turno` para respetar el modo claro/oscuro del panel.
- [Impresion] La salida impresa conserva fondo blanco, texto oscuro y bordes legibles para carta, ejecutivo y POS.
- [Alcance] No cambia backend, endpoints, permisos ni tablas.

## [2026-05-18] Reporte de turno en papel grande y POS
- [Frontend] `corte_de_caja.html` compacta el modo POS 80mm para imprimir detalle de ventas como bloques verticales con etiquetas.
- [Frontend] `reportes_turnos.html` agrega selector de papel grande/POS para la vista imprimible historica.
- [QA] Se valida que papel grande conserve ancho carta y que POS no genere desborde horizontal.

## [2026-05-18] Reportes de turnos historicos
- [Frontend] Nueva pagina `Reportes de turnos` en el submenu de reportes, con listado de turnos antiguos, filtros, vista imprimible, imprimir, compartir, exportar y enviar por email.
- [Backend] `/api/empresa/corte_caja` agrega acciones historicas para listar cierres, reconstruir un reporte por `cierre_caja_id`, exportarlo en `json/csv/txt/xls/pdf` y enviarlo como adjunto.
- [Permisos] Se agrega `linkReportesTurnos` bajo modulo `reportes` con accion lectura, separando consulta historica del cierre operativo de caja.

## [2026-05-18] Reporte de turno configurable
- [Backend] `/api/empresa/corte_caja` incluye NIT de empresa y fechas de entrada/salida por venta para el detalle del turno.
- [Base de datos] `empresa_corte_caja_configuracion` agrega checks de encabezado, datos de empresa, usuario, consecutivo, columnas del detalle y totales de productos/servicios por `empresa_id`.
- [Frontend] `corte_de_caja.html` renderiza el reporte de turno con encabezado, detalle ordenado por fecha/hora de venta y resumenes de caja; `configuracion.html` permite activar/desactivar los bloques.
- [QA] Se agregan pruebas de default profesional y se validan scripts inline de las pantallas afectadas.

## [2026-05-18] Panel ejecutivo super administrador
- [Frontend] `web/super/licencias_resumen.html` reemplaza el centro de mando saturado por un tablero ejecutivo compacto.
- [UX] Se dejan 6 KPIs unicos, prioridades, accesos clave e incidentes recientes, retirando bloques repetidos de negocio, servicios y costos.
- [Alcance] No cambia backend, endpoints, permisos, tablas ni dependencias.

## [2026-05-18] Login administrador
- [UX] `web/login.html` cambia el titulo superior a `Powerful Control System`.
- [UX] `Acceso de administradores` queda superpuesto sobre la imagen lateral y `Ir al inicio` queda debajo de la tarjeta del login.
- [Alcance] No cambia autenticacion, reCAPTCHA, permisos, tablas ni endpoints.

## [2026-05-18] Fix formato monetario por empresa
- [Backend] `empresa_configuracion_avanzada` normaliza columnas legacy booleanas a enteros `0/1` antes de guardar configuracion avanzada.
- [Backend] El `UPSERT` de configuracion avanzada usa `sqlNowExpr()` para fechas runtime PostgreSQL.
- [Backend] `configuracion_operativa` guarda configuracion, roles, politicas e historial con `RETURNING id`, reparando snapshots y rollback en PostgreSQL.
- [QA] Se agrega regresion para el guardado de configuracion monetaria/numerica y la migracion ligera de flags.
- [QA] Se agrega regresion para impedir `LastInsertId()` en los guardados operativos de configuracion.

## [2026-05-18] Corte de caja: acciones y texto de estaciones
- [Frontend] `corte_de_caja.html` mueve `Generar corte`, `Ver reporte de mi turno`, `Corte automatico`, `Cerrar turno` e `Imprimir seleccion` dentro de `Lectura rapida` como botones verticales.
- [UX] `Ver reporte de mi turno` centra el reporte como hoja imprimible en pantalla y respeta el formato `Carta`, `Ejecutivo` o `POS`.
- [Frontend/Backend] Los textos visibles del control por sensores usan `Estaciones ocupadas sin factura`; se actualizan etiquetas visibles de reportes, tarifas y preconfiguraciones.
- [Fix] `Ver reporte de mi turno` limpia el cierre/caja automaticos previos antes de consultar el reporte del usuario actual.

## [2026-05-18] Panel sin tarjeta de mercado
- [Frontend] `web/administrar_empresa/panel.html` elimina la tarjeta `Mercado en contexto`.
- [Operacion] El panel ya no consulta indicadores externos de divisas, criptoactivos ni ETFs al cargar.
- [Alcance] No cambia backend, permisos, tablas ni endpoints.

## [2026-05-18] Configuracion del reporte de corte en Configuracion
- [Frontend] `corte_de_caja.html` ya no muestra las tarjetas de configuracion del reporte ni de reportes a imprimir.
- [Frontend] `configuracion.html` agrega la seccion `Reporte de corte` con formato, reportes a imprimir, metricas y botones de guardar/restaurar.
- [Alcance] Reutiliza `/api/empresa/corte_caja/configuracion`; no cambia tablas ni endpoints.

## [2026-05-18] Corte de Caja en operacion y ventas
- [Frontend] `web/administrar_empresa.html` muestra el boton `Corte de Caja` debajo de `Estaciones`.
- [Navegacion] El acceso abre `corte_de_caja.html` con el mismo contexto usado al hacer clic en la tarjeta Caja de estaciones.
- [Permisos] `linkCorteCaja` queda visible como acceso operativo del menu principal; no crea endpoint ni ruta nueva.

## [2026-05-18] Reporte de mi turno desde estacion Caja
- [Frontend] `web/administrar_empresa/corte_de_caja.html` agrega el boton `Ver reporte de mi turno`.
- [Backend] `/api/empresa/corte_caja?action=reporte_mi_turno` reutiliza el filtro seguro de caja abierta del usuario autenticado.
- [UX] El boton muestra e imprime el reporte del turno actual sin habilitar el guardado/cierre accidental del turno.

## [2026-05-17] Venta directa con acciones del carrito de estaciones
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` trata venta directa y estaciones como la misma vista enfocada tambien para la tarjeta de acciones.
- [UX] Los botones operativos del carrito, incluido `Abonos`, cargan en venta directa con la misma configuracion global del carrito.
- [Compatibilidad] No cambia endpoints, tablas, permisos ni el carrito canonico `VENTA-DIRECTA-{empresa_id}-0`.

## [2026-05-17] Menu flotante sin compartir WhatsApp
- [Frontend] `web/menu.js` retira `Compartir por WhatsApp` del submenu `Utilidades`.
- [UX] `Cambiar apariencia` queda como penultimo elemento del panel flotante y `Cerrar sesion`/`Iniciar sesion` queda de ultimo.
- [Alcance] No cambia backend, permisos, endpoints ni configuracion de WhatsApp del portal publico.

## [2026-05-17] Inventario debajo de operacion y ventas
- [Frontend] `web/administrar_empresa.html` mueve el grupo `Inventario y compras` justo debajo de `Operacion y ventas`.
- [Alcance] Solo cambia el orden visual del submenu empresarial; no altera rutas, permisos, endpoints ni tablas.

## [2026-05-17] Abonos operativos en carrito de estacion
- [Backend] Se agrega `carrito_compra_abonos` con aislamiento por `empresa_id` y endpoints `action=abonos|abono` en `/api/empresa/carritos_compra`.
- [Pago] `pagar_estacion` valida abonos registrados, descuenta el saldo final a cobrar y conserva el total de la cuenta para documento/venta.
- [Frontend] El boton `Abonos` del carrito de estacion abre una tarjeta para registrar/listar abonos y el desglose muestra cuenta, abonos y saldo final.
- [QA] Validado con pruebas Go enfocadas, parseo JS y prueba visual Playwright mock: abono COP 30000 sobre cuenta COP 100000 deja `total_pagado=70000`.

## [2026-05-17] Venta directa con carrito igual a estaciones
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` usa la misma vista enfocada de estaciones para venta directa.
- [UX] Items, totales, lector y controles de pago cargan en el panel operativo compartido, evitando que venta directa caiga al grid general antiguo.
- [Compatibilidad] Se conserva el carrito canonico `VENTA-DIRECTA-{empresa_id}-0`; no cambian endpoints, tablas ni permisos.

## [2026-05-17] Favoritos del panel como botones
- [Frontend] `web/administrar_empresa/panel.html` ajusta los favoritos de `Accesos rapidos` para verse y comportarse como botones de accion.
- [UX] Se agregan borde destacado, fondo accionable, cursor, estado activo y foco visible sin cambiar rutas, permisos ni almacenamiento local.

## [2026-05-17] Comunicaciones super unificadas
- [Frontend] `web/super_administrador.html` crea el modulo principal `Comunicaciones`.
- [Navegacion] El modulo agrupa mantenimiento, mensajes masivos, alertas sistema, Gmail SMTP, alertas de licencia y WhatsApp portal.
- [Compatibilidad] No cambia endpoints, permisos ni formularios; solo consolida la ubicacion visual del submenu super.

## [2026-05-17] Mantenimiento super como modulo principal
- [Frontend] `web/super_administrador.html` mueve `Mantenimiento sistema` a un grupo propio del menu principal super.
- [Frontend] `web/super/mantenimiento_sistema.html` agrega `Eliminar alertas viejas` para limpiar avisos desactivados o con fecha anterior a hoy.
- [Backend] `/super/api/config/mantenimiento?action=limpiar_viejos` elimina alertas viejas y resincroniza el aviso visible para empresas sin tocar el bloqueo real.
- [QA] Se agrega prueba de filtrado de avisos viejos e inactivos.

## [2026-05-17] Configuracion super por paginas
- [Frontend] `web/super_administrador.html` agrega el grupo `Configuracion` con paginas independientes para consumos, RustDesk, limitaciones, OnlyOffice, Voz IA, Epayco, Wompi/Nequi, Gmail, alertas de licencia, WhatsApp, reCAPTCHA, 2FA, IA global y respaldo.
- [Frontend] `web/super/configuracion/*.html` son paginas contenedoras por seccion y `web/super/configuracion_avanzada.html` soporta modo aislado `?single=1&section=...` para cargar solo la tarjeta correspondiente.
- [Compatibilidad] Los formularios, endpoints y botones de guardar existentes se conservan; no cambia base de datos ni backend.

## [2026-05-17] Mantenimiento del sistema como pagina super
- [Backend] `/super/api/config/mantenimiento` agrega gestion de avisos individuales con `action=desactivar|eliminar`, persistidos en `mantenimiento_programado.avisos_json`.
- [Frontend] `web/super/mantenimiento_sistema.html` queda como pagina independiente en el submenu principal de Super Administrador y muestra tabla de avisos programados con acciones para cargar, desactivar o eliminar.
- [Navegacion] `web/super_administrador.html` agrega el acceso `Mantenimiento`; `web/super/configuracion_avanzada.html` deja de contener esa tarjeta.
- [Seguridad] Desactivar/eliminar avisos no cambia el bloqueo real `mantenimiento_activo`.
- [QA] Se agregan pruebas enfocadas en `backend/handlers/super_mantenimiento_handlers_test.go`.

## [2026-05-17] Panel primero en administrar empresa
- [Frontend] `web/administrar_empresa.html` elimina el grupo `Inicio` y deja `Panel` como primer boton directo del menu lateral.
- [Backend] `backend/handlers/empresa_permisos.go` retira `linkInicio` del catalogo de paginas; `linkPanelEmpresa` queda como acceso general.
- [Alcance] No cambia rutas, endpoints, tablas ni dependencias.

## [2026-05-17] Carrito simplificado como default por tipo de empresa
- [Backend] `defaultEmpresaPreconfigCarritoUI()` usa el preset operativo limpio para empresas nuevas.
- [Backend] `ApplyDefaultCarritoUIToExistingEmpresaPrefs` aplica el mismo preset a empresas antiguas al arrancar; tambien crea una configuracion base para empresas sin `estaciones_config`.
- [Frontend] Los defaults de `configuracion_carrito_de_compra_empresa.html`, `configuracion_de_estaciones.html`, `carrito_de_compras.html` y `estaciones.html` quedan alineados.
- [UX] Activos por defecto: buscar productos, catalogo, lector, items, totales, acciones, valores por medio de pago, pago mixto y pagar.
- [UX] Apagados por defecto: cobro avanzado, descuentos, propina, comision, desglose de cobro y lavador/comision.
- [QA] Se agregan pruebas enfocadas del preset y de la migracion de empresas antiguas.
- [Alcance] Sin cambios de APIs, tablas, permisos ni dependencias.

## [2026-05-17] Carrito mueve Buscar Productos junto a Agregar
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` reubica el boton de productos por botones junto al boton `Agregar` y lo rotula `Buscar Productos`.
- [Frontend] `web/administrar_empresa/buscar_producto_botones.html` deja la barra `Buscar producto` con placeholder para escribir el nombre.
- [Alcance] Sin cambios de backend, APIs, tablas, permisos ni dependencias.

## [2026-05-17] Carrito sin encabezado descriptivo operativo
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` oculta el encabezado superior en venta directa o estacion.
- [UX] Se elimina el texto `Venta directa usa una sola caja abierta...` para que la pantalla empiece directamente desde los controles del carrito.
- [Alcance] Sin cambios de backend, APIs, tablas, permisos ni dependencias.

## [2026-05-17] Carrito calcula cambio por efectivo recibido
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` agrega el campo `Efectivo recibido` en la seccion de valores por medio de pago.
- [UX] El carrito calcula `Cambio a devolver` cuando el recibido cubre el efectivo esperado, o `Falta recibir` cuando no alcanza.
- [Alcance] Sin cambios de backend, APIs, tablas, permisos ni dependencias.

## [2026-05-17] Carrito oculta secciones avanzadas por check
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` agrega los checks `Mostrar opciones de cobro` y `Mostrar lavador`.
- [UX] `Cobro y estados del carrito` y `Lavador` quedan apagados por defecto y se abren solo cuando el usuario los necesita.
- [Operativo] Al ocultar opciones de cobro se limpian descuento, devolucion, referencia y propina para no aplicar valores invisibles.
- [Alcance] Sin cambios de backend, APIs, tablas, permisos ni dependencias.

## [2026-05-17] Administrar empresa abre Panel por defecto
- [Frontend] `web/js/administrar_empresa.js` usa `linkPanelEmpresa` como pagina preferida de arranque del iframe.
- [UX] El usuario entra primero al tablero de administracion, aunque el grupo `Operacion y ventas` siga arriba en el menu.
- [Alcance] Sin cambios de permisos, APIs, tablas ni dependencias.

## [2026-05-17] Administrar empresa simplifica Operacion y ventas
- [Frontend] `web/administrar_empresa.html` deja el grupo `Operacion y ventas` solo con `Venta directa` y `Estaciones`.
- [Frontend] `Venta publica`, `Red social empresarial`, `Codigos de descuento` y `Chat y tareas` pasan al grupo `Canales digitales y colaboracion`.
- [Frontend] `Reservas` se mueve a `Soluciones por negocio`.
- [Permisos] `backend/handlers/empresa_permisos.go` actualiza los grupos administrativos de esas paginas y mueve `Punto de venta / TPV` a `Permisos base de ventas`, sin cambiar acciones, rutas ni endpoints.
- [Alcance] Sin tablas ni dependencias nuevas; mantiene aislamiento por `empresa_id`.

## [2026-05-17] Facturacion electronica DIAN revisada
- [Backend] `FacturacionPaisVistaFor("CO")` deja de referenciar el modulo ERP/documental viejo y apunta a la subpagina canonica de pruebas DIAN.
- [DIAN] El set base para software propio/proveedor tecnologico queda en 60 facturas, 20 notas debito y 20 notas credito, editable segun el objetivo del portal DIAN.
- [Frontend] `facturacion_electronica_pruebas_dian.html` muestra esos valores base y advierte que deben ajustarse si DIAN asigna otro objetivo.
- [Documentacion] Se registra que la configuracion esta separada por empresa/pais, que Colombia conserva pruebas DIAN aparte y que el correo automatico actual no adjunta XML/PDF fiscal certificado.

## [2026-05-17] Estaciones con primer clic solo activa
- [Frontend] `configuracion_de_estaciones.html` agrega el check `Primer clic solo activa`.
- [Frontend] `estaciones.html` usa `solo_activar_primer_clic` para que el primer clic active la estacion sin abrir carrito y el segundo clic abra el carrito ya activo.
- [Compatibilidad] `abrir_carrito_al_activar=false` sigue funcionando como alias historico.
- [QA] Validado con parseo JS, pruebas Go enfocadas y capturas visuales del check, primer clic y segundo clic.
- [Alcance] Sin tablas, endpoints ni dependencias nuevas.

## [2026-05-17] Reporte de turno y caja verificado
- [Frontend] `web/administrar_empresa/reportes_ejecutivos.html` agrega filtros de usuario/cajero, caja, turno y cierre ID para consultar/exportar `reporte_de_turno`.
- [Backend] `backend/handlers/reportes_catalogo_test.go` protege que `reporte_de_turno` siga en el catalogo profesional de reportes.
- [QA] Se valido visualmente `Corte automatico` -> `Turno cerrado` -> impresion, el reporte automatico de ultimos movimientos por caja/usuario actual y la vista previa filtrada del reporte de turno.
- [Alcance] Sin tablas, endpoints ni dependencias nuevas; mantiene aislamiento por `empresa_id`.

## [2026-05-15] Venta directa abre Buscar por botones
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` prepara o recupera el carrito enfocado antes de abrir `Buscar (botones)`, especialmente en `modo=venta_directa`.
- [Frontend] `web/administrar_empresa/buscar_producto_botones.html` conserva `modo=venta_directa` en el retorno, lo envia en las APIs de carrito/items y acepta respuesta puntual o lista filtrada al resolver el carrito.
- [Backend] `GET /api/empresa/carritos_compra?empresa_id=...&id=...` devuelve el carrito solicitado en vez de ignorar `id` y responder siempre el listado completo.
- [Permisos] Las subrutas `/api/empresa/carritos_compra/*` tambien se resuelven contra `linkVentaDirecta` cuando traen contexto de venta directa.
- [Alcance] Sin tablas ni dependencias nuevas; conserva permisos e aislamiento por `empresa_id`.

## [2026-05-15] Administrar empresa prioriza Venta directa
- [Frontend] `web/administrar_empresa.html` mueve `Operacion y ventas` al inicio del menu; `Venta directa` queda como primer acceso y `Estaciones` como segundo.
- [Frontend] `web/js/administrar_empresa.js` prioriza `linkVentaDirecta` en el orden de enlaces y en el iframe inicial cuando el permiso esta visible.
- [Permisos] `linkCarritoCompras` queda agrupado como `Configuracion - Ventas y cobro`; no cambia ruta, accion ni aislamiento por `empresa_id`.
- [Alcance] Sin tablas ni dependencias nuevas.

## [2026-05-15] Asistencia de empleados reparada y profesionalizada
- [Backend] `backend/db/asistencia_empleados.go` deja de usar SQL runtime incompatible con PostgreSQL en configuracion, cierres, listados, CRUD y marcacion de entrada/salida.
- [Backend] Las operaciones usan `queryRowSQLCompat`, `querySQLCompat`, `execSQLCompat`, `insertSQLCompat` y `sqlNowExpr()`.
- [UX] `Agregar registro` y `Editar` cambian automaticamente a la seccion `Registro`, aunque el submenu interno este en Resumen, Consulta u otra pestaña.
- [UX] Los botones de tabla quedan como acciones compactas visibles; `Entrada ahora` y `Salida ahora` registran la hora actual sin prompts nativos.
- [QA] Validado con `go test ./db ./handlers -count=1` y prueba visual Playwright con API mock: crear, entrada, salida, editar, guardar configuracion, filtrar, limpiar, cerrar periodo y descargar reporte.
- [Alcance] Sin tablas, permisos ni dependencias nuevas; mantiene aislamiento por `empresa_id`.

## [2026-05-15] Centro de reportes sin pagina IA dedicada
- [Frontend] `web/administrar_empresa/reportes_menu.html` retira el boton `Asistente IA`.
- [Frontend] `web/administrar_empresa/reportes_ejecutivos.html` elimina el boton superior y la tarjeta `Asistente IA para reportes`.
- [Limpieza] Se elimina `web/administrar_empresa/reportes_ia_chat.html` y el permiso de pagina `linkReportesIAChat`.
- [Backend] `/api/empresa/reportes_ia_chat` se conserva como soporte tecnico del asistente global en modo reportes, autorizado por el permiso general `linkReportes`.
- [Alcance] Sin tablas ni dependencias nuevas; el centro de reportes queda como pantalla unica de consulta/exportacion.

## [2026-05-15] Submenu de facturacion electronica sin regreso a empresas
- [Frontend] `web/administrar_empresa/facturacion_electronica_menu.html` retira el enlace `Volver a empresas`.
- [UX] El submenu queda enfocado en configuracion, pruebas DIAN, proveedores de firma, facturas electronicas y AIU construccion.
- [Alcance] Sin permisos, endpoints, tablas ni dependencias nuevas.

## [2026-05-15] Nuevo credito compatible con PostgreSQL
- [Backend] `CreateEmpresaCredito` usa expresiones de fecha PostgreSQL via `sqlNowExpr()` en el insert runtime.
- [Backend] La asignacion automatica de codigo y la generacion de cuotas usan `execTxSQLCompat`, evitando `tx.Exec` directo con placeholders SQLite.
- [Causa] El boton `Nuevo credito` podia responder 500 al crear creditos con cuotas porque la transaccion mezclaba `?` y `datetime('now','localtime')`.
- [QA] Validado con prueba Go enfocada y `go test ./db ./handlers -count=1`.
- [Alcance] Sin tablas, permisos, endpoints ni dependencias nuevas; mantiene aislamiento por `empresa_id`.

## [2026-05-15] Productos avisa cuando falta bodega
- [Frontend] `Nuevo producto` valida bodegas activas antes de abrir el formulario.
- [UX] Si no existe bodega activa, se muestra un aviso claro y un boton `Crear bodega`.
- [Resguardo] El guardado de productos nuevos tambien bloquea el caso sin bodega activa.
- [QA] Validado con parseo del script inline y prueba visual Playwright en empresa sin bodegas y con una bodega activa.
- [Alcance] Sin backend, tablas, permisos ni dependencias nuevas.

## [2026-05-15] Compras con proveedores creados previamente
- [Frontend] `compras_avanzadas.html` usa selectores de proveedores activos para proveedor sugerido, cotizacion y recepcion.
- [Frontend] `compras.html` agrega acceso a gestionar proveedores y autocompleta el documento del proveedor elegido.
- [Backend] `compras_avanzadas` valida que `proveedor_id` exista activo en la misma empresa antes de guardar cotizaciones o recepciones.
- [BD] `empresa_compras_recepciones_avanzadas` agrega `proveedor_id` para trazabilidad por proveedor.
- [QA] Validado con `node --check`, `go test ./db ./handlers`, `git diff --check` y prueba visual Playwright con API mock.
- [Alcance] Sin dependencias nuevas; se mantiene PostgreSQL y aislamiento por `empresa_id`.

## [2026-05-15] Menu empresarial con Venta directa primero
- [Frontend] `web/administrar_empresa.html` mueve `Venta directa` al primer lugar de `Operacion y ventas`.
- [Frontend] `Carritos` se retira del grupo operativo principal y queda disponible en `Configuracion > Ventas y cobro`.
- [QA] Validado con Playwright revisando el orden del menu principal y la permanencia de `Carritos` en el submenu de configuracion.
- [Alcance] Sin permisos, rutas, endpoints, tablas ni dependencias nuevas.

## [2026-05-15] Alta de bodegas compatible con PostgreSQL
- [Backend] `backend/db/productos.go` usa `sqlNowExpr()` en `CreateBodega` para `fecha_creacion` y `fecha_actualizacion`, eliminando `datetime('now','localtime')` del insert runtime.
- [QA] `backend/db/productos_bodegas_test.go` blinda que el alta de bodega no vuelva a usar funciones SQLite.
- [Visual] Validado con Playwright en escritorio y movil: abrir `bodega.html`, crear bodega, ver mensaje de exito, KPI actualizado y fila en tabla.
- [Alcance] Sin tablas, permisos, rutas ni dependencias nuevas.

## [2026-05-15] Login y registro movil con controles tactiles
- [Frontend] `web/estilos.css` aumenta a minimo 44 px las areas tactiles de botones, enlaces, inputs y mostrar contrasena dentro de `login-page`.
- [UX] `registrar_nuevo_usuario_administrador.html` marca el body de registro para compactar el chat flotante en celular y evitar que tape el formulario.
- [QA] Validado con Playwright movil en login administrador, login operativo, recuperacion de invitacion, registro operativo por invitacion y registro administrador con mock de endpoints.
- [Alcance] Sin backend, tablas, permisos ni dependencias nuevas.

## [2026-05-15] Venta directa con carrito 0 automatico
- [Frontend] `carrito_de_compras.html?modo=venta_directa` crea o reutiliza `VENTA-DIRECTA-{empresa_id}-0` si no llega `carrito_codigo`.
- [Compatibilidad] Se reutilizan carritos antiguos `VENTA-DIRECTA-{empresa_id}` o `CAJA_DIRECTA` antes de crear uno nuevo.
- [UX] Venta directa deja de mostrarse como estacion y oculta el boton `Regresar a estaciones`.
- [Alcance] Sin tablas, endpoints ni dependencias nuevas.

## [2026-05-15] Empresas compartidas con alcance por rol y modulos
- [Backend] `admin_empresa_compartida` y `admin_empresa_compartida_invitaciones` agregan `nivel_acceso` y `modulos_permitidos`.
- [Permisos] El alcance compartido se aplica sobre modulos, acciones y paginas efectivas despues de licencia, vertical, rol y politicas finas de empresa.
- [UI] `editar_empresa` y el selector de empresas permiten elegir `Solo ver`, `Acceso total` o `Solo ciertos modulos`, muestran el alcance y ofrecen `Dejar de compartir`.
- [Seguridad] Revocar acceso invalida caches de empresa/permisos para retirar el acceso inmediatamente.
- [Alcance] Sin dependencias nuevas; PostgreSQL y aislamiento por `empresa_id`.

## [2026-05-15] Licencia de prueba unica por empresa y asesor comercial
- [Backend] `licencias_activaciones_gratis` agrega `asesor_id` y un indice unico activo por `empresa_id` para impedir mas de una prueba/gratis por empresa.
- [Migracion] El arranque normaliza duplicados activos historicos como `historico_duplicado` antes de crear el indice unico.
- [Checkout] La prueba de 15 dias y la activacion sin pago aceptan `asesor_id`/`codigo_asesor`, validan que el asesor exista, este aceptado y no este inactivo, y dejan trazabilidad en pagos/comisiones.
- [Comercial] La prueba de 15 dias queda con 250 documentos/ventas mensuales y la UI avisa que solo se puede activar una vez por empresa.
- [Alcance] Sin dependencias nuevas; se mantiene PostgreSQL y aislamiento por `empresa_id`.

## [2026-05-14] Nucleo de pantallas empresariales unificado
- [Frontend] `Venta directa` entra directo al carrito unico con `modo=venta_directa` y permiso `linkVentaDirecta`.
- [Frontend] `Productos` usa `administrar_productos.html?view=...` para categorias/proveedores/precios/compras sin wrappers intermedios.
- [Limpieza] Se eliminan los HTML que solo redirigian: `venta_directa.html`, `documentos_onlyoffice_menu.html`, `crm_ventas_avanzadas.html`, `categorias.html` y wrappers bajo `web/administrar_empresa/productos/`.
- [Backend] La preconfiguracion y configuracion guiada guardan la URL canonica de venta directa.
- [Alcance] Sin tablas, permisos ni dependencias nuevas; se conserva aislamiento por `empresa_id`.

## [2026-05-14] Reportes unificados y limpieza preproduccion
- [Frontend] `web/administrar_empresa/reportes_menu.html` queda con Centro de reportes y Asistente IA.
- [Frontend] `web/administrar_empresa/reportes_ejecutivos.html` agrega catalogo, vista previa y exportacion directa sin saltar a paginas antiguas.
- [Limpieza] Se eliminan `reportes.html`, `reportes_inventario.html`, `reportes_finanzas.html`, `graficos_estadisticas.html` y `backend/handlers/graficos_estadisticas.go`.
- [Datos] Se limpian usuarios/clientes/empresa smoke de la base, se normalizan productos `DEMO-*` a `BASE-*`, se retiran evidencias QA locales antiguas y se elimina el runner que generaba finanzas QA en Motel Calipso.
- [Seguridad] Sin tablas, permisos ni dependencias nuevas; los datasets siguen bajo `/api/empresa/reportes`.

## [2026-05-14] Lobby de juegos uniforme con Doon FPS
- [Frontend] `web/Juegos/menu_juegos.html` deja todas las tarjetas del catalogo con tamano uniforme, portada 16:9 y descripcion.
- [Juegos] Se retiran del lobby los juegos 3D y el emulador generico; queda visible solo `N64` como emulador.
- [Nuevo] `web/Juegos/doon_fps/index.html` agrega un shooter retro original en canvas/raycasting con dos sectores y records.
- [Assets] Portadas raster originales en `web/img/juegos`, sin capturas comerciales ni ROMs.

## [2026-05-14] Carrito con valores editables por medio de pago
- [Frontend] Los campos `Efectivo`, `Debito`, `Credito` y `Otro` del carrito ahora aceptan edicion directa.
- [Pago] Cuando hay varios medios con valor, el carrito envia `metodo_pago=mixto` con `pagos_mixtos`; `Otro` se registra como transferencia bancaria.
- [Validacion] La suma debe coincidir con el total final y las tarjetas/transferencias mantienen referencia obligatoria.
- [Alcance] Sin cambios de backend, tablas, permisos ni dependencias.

## [2026-05-14] Carrito limpia titulos y destaca acciones
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` oculta los textos repetidos de items, lector, valores de pago y acciones.
- [UX] `Carrito de compras` y `Estacion: <nombre>` quedan en una misma fila cuando el ancho lo permite.
- [Estilos] `web/estilos.css` conserva tarjetas planas, pero devuelve apariencia de boton al toolbar de acciones.
- [Alcance] Cambio solo visual; no modifica APIs, backend, permisos, tablas ni dependencias.

## [2026-05-13] Boton pagar y cerrar carrito con reintento visual
- [Backend] `pagar_estacion` y `pagar` pasan a permiso operativo `Crear / registrar` en ventas, evitando bloquear cajeros o roles personalizados que venden pero no tienen acciones gerenciales de aprobacion.
- [Frontend] Si la carga inicial de cajas abiertas falla, el boton `Pagar y cerrar carrito` queda habilitado, informa que puede reintentar y al clic vuelve a cargar cajas antes de enviar el pago.
- [QA] Validado con `go test ./handlers ./db` y prueba visual Playwright simulando 403 inicial de cajas: el boton queda clickeable, recarga `CAJA-1` y envia `PUT action=pagar_estacion` con `cierre_caja_id=11`.
- [Alcance] Sin tablas ni dependencias; mantiene validacion final de caja abierta por `empresa_id` en backend.

## [2026-05-13] Carrito cobra con caja abierta desde permisos de ventas
- [Backend] `backend/handlers/carritos_compras.go` agrega `GET /api/empresa/carritos_compra?action=cajas_abiertas` para listar cajas abiertas necesarias al cobrar sin exigir acceso al modulo financiero.
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` carga cajas abiertas desde el flujo de carritos/ventas y deja de depender de `/api/empresa/finanzas/cierres_caja`.
- [QA] Validado con `go test ./handlers ./db` desde `backend` y prueba visual Playwright mock: boton `Pagar y cerrar carrito` habilitado, caja seleccionada y `PUT action=pagar_estacion` con `cierre_caja_id`.
- [Alcance] Sin tablas, dependencias ni motores nuevos; conserva aislamiento por `empresa_id` y validacion final de caja abierta en backend.

## [2026-05-13] Carrito plano reforzado contra temas posteriores
- [Frontend] `web/estilos.css` agrega un override absoluto al final del archivo para `carrito-flat-page`.
- [UX] Ningun tema, modo tactil o estilo legacy posterior puede volver a mostrar sombras, radios o margenes en tarjetas del carrito.
- [QA] Validado con parseo JS y Playwright revisando 15 tarjetas/formularios con `box-shadow: none`, `border-radius: 0px` y margenes `0px`.
- [Alcance] Cambio solo visual; no modifica APIs, backend, permisos, tablas ni dependencias.

## [2026-05-13] Panel empresarial prioriza ciudad por IP
- [Frontend] `web/administrar_empresa/panel.html` detecta primero la ciudad por IP para clima e indicadores cuando no existe ciudad manual guardada.
- [UX] GPS queda como respaldo si falla la IP o como accion manual desde el boton `Usar GPS`; la ciudad manual guardada conserva prioridad.
- [QA] Validado con parseo JS y prueba visual Playwright mock: IP devuelve `Medellin`, el panel muestra esa ciudad y no consulta GPS.
- [Alcance] Cambio solo frontend; no modifica APIs, backend, permisos, tablas ni dependencias.

## [2026-05-13] Carrito de compras compacto tipo tabla plana
- [Frontend] `web/estilos.css` compacta el layout bajo `carrito-flat-page` quitando gaps, margenes, radios y sombras de tarjetas/contenedores del carrito.
- [UX] Las tarjetas quedan pegadas visualmente entre si, con lectura plana similar a la tabla de items.
- [QA] Validado con Playwright comprobando `gap: 0`, `margin: 0`, `border-radius: 0` y `box-shadow: none` en contenedores clave.
- [Alcance] Cambio solo visual; no modifica APIs, backend, permisos, tablas ni dependencias.

## [2026-05-13] Carrito de compras sin sombras en tarjetas
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` agrega clase scoped `carrito-flat-page` para controlar el estilo visual de esta pagina.
- [UX] `web/estilos.css` deja las tarjetas y formularios del carrito sin `box-shadow`, evitando apariencia 3D o relieve.
- [QA] Validado con parseo JS y prueba visual Playwright comprobando tarjetas del carrito con `box-shadow: none`.
- [Alcance] Cambio solo visual; no modifica APIs, backend, permisos, tablas ni dependencias.

## [2026-05-13] Carrito de estacion mueve contexto al encabezado
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` deja de mostrar `Items de estacion: <nombre>` dentro de la tarjeta del carrito.
- [UX] En modo estacion, el encabezado superior mantiene `Carrito de compras` y muestra debajo `Estacion: <nombre>`, dejando la tarjeta interna como `Items del carrito`.
- [Alcance] Cambio solo visual; no modifica APIs, backend, permisos, tablas ni dependencias.

## [2026-05-13] OnlyOffice abre editor con URL publica del Document Server
- [Backend] `backend/handlers/onlyoffice.go` corrige la carpeta temporal de documentos para no crear `/data/empresas/empresas/{empresa_id}` cuando `PCS_DATA_ROOT` ya apunta a `/data/empresas`.
- [Backend] `editor_config` entrega al navegador una URL publica de OnlyOffice cuando la configuracion global apunta a un hostname interno Docker como `onlyoffice-documentserver`; tambien acepta `ONLYOFFICE_PUBLIC_DOCUMENT_SERVER_URL` o `ONLYOFFICE_BROWSER_DOCUMENT_SERVER_URL` como override operativo.
- [Frontend] `web/administrar_empresa/documentos_onlyoffice.html` valida que `api.js` cargue `DocsAPI.DocEditor` y muestra un mensaje claro dentro del marco si el Document Server no es accesible desde el navegador.
- [QA] Se agrega `backend/handlers/onlyoffice_test.go`; validado con `go test ./handlers -run OnlyOffice -count=1` y prueba visual Playwright del flujo crear y abrir editor.
- [Alcance] Sin tablas, permisos ni dependencias nuevas; conserva aislamiento por `empresa_id` y el wrapper de permisos `documentos_onlyoffice`.

## [2026-05-13] Reportes ejecutivos profesionales
- [Frontend] `web/administrar_empresa/reportes_menu.html` reduce el submenu a entradas de alto nivel: centro ejecutivo, analitica, suite, inventario, finanzas e IA.
- [Frontend] `web/administrar_empresa/reportes_ejecutivos.html` agrega una portada de mando con accesos agrupados por direccion, ventas, inventario, finanzas, contabilidad/fiscal e IA.
- [Frontend] `web/administrar_empresa/reportes_finanzas.html` reemplaza botones placeholder por carga real del dataset `contable_movimientos_financieros` y exportacion por formato usando `/api/empresa/reportes`.
- [Alcance] Sin cambios de backend, base de datos, permisos ni dependencias.

## [2026-05-13] Facturacion electronica separa configuracion y pruebas DIAN
- [Frontend] `web/administrar_empresa/facturacion_electronica.html` elimina el boton `Abrir modulo DIAN / documental` y queda enfocada en configurar pais, firma, DIAN Colombia y parametros avanzados de la empresa.
- [Frontend] `web/administrar_empresa/facturacion_electronica_pruebas_dian.html` concentra diagnostico DIAN, set de habilitacion, conexion/cola y emision documental manual.
- [Navegacion] `web/administrar_empresa/facturacion_electronica_menu.html` agrega `Pruebas DIAN y documentos`.
- [Alcance] Sin cambios de backend, base de datos, permisos ni dependencias.

## [2026-05-13] Control de aseo por estacion
- [Backend] Se agrega `users.control_aseo_estaciones`, `empresa_estacion_aseo_eventos` y `/api/empresa/estacion_aseo` para medir el tiempo desde `estacion_estado_sucia=1` hasta el reporte de aseo terminado.
- [Frontend] `administrar_usuarios.html` permite activar/desactivar el control por usuario; `estaciones.html` permite reportar el aseo con un clic sobre la estacion sucia; `reporte_aseo_estaciones.html` muestra tiempos por estacion.
- [Seguridad] La finalizacion se atribuye al usuario autenticado y conserva aislamiento por `empresa_id`; el reporte queda para roles administrativos/supervision.
- [QA] Validado con pruebas Go dirigidas de `db`/`handlers` y parseo JS inline de las paginas tocadas.

## [2026-05-13] Factura electronica de venta no queda emitida sin acuse fiscal
- [Backend] `backend/handlers/carritos_compras.go` mantiene cerrada la venta/comprobante, pero degrada la factura electronica asociada a `pendiente_emision` cuando Colombia produccion requiere acuse fiscal y la integracion DIAN/proveedor no queda `enviado`.
- [Backend] `backend/handlers/facturacion_electronica.go` bloquea Colombia produccion con proveedor `manual`, `local` o `interno`, para no simular como exitoso un envio que debe pasar por DIAN/proveedor real.
- [Operacion] El resultado conserva `integracion_fiscal`, `cola_reintentos` y una observacion con la causa para que el usuario reprocesar o corregir configuracion sin confundir una factura pendiente con una aceptada.
- [QA] Se agrega cobertura en `backend/handlers/facturacion_documentos_electronicos_test.go` para exigir acuse solo en Colombia produccion y confirmar que `fallido` no sea tratado como enviado.
- [Alcance] No cambia tablas, permisos, endpoints publicos ni dependencias; conserva aislamiento por `empresa_id`.

## [2026-05-13] Logos empresariales activos para Calipso y Bollon
- [Activos] Se agregan `web/uploads/empresa_logos/empresa_7/motel-calipso-logo.svg` y `web/uploads/empresa_logos/empresa_32/gimnasio-el-bollon-logo.svg`.
- [Datos] En produccion se actualiza `empresa_configuracion_avanzada.logo_url` para `empresa_id=7` y `empresa_id=32`, dejando `mostrar_logo` y `mostrar_logo_empresa` activos.
- [UX] Panel empresarial, carrito, facturas, recibos y reportes usan esos logos a traves de la configuracion avanzada existente.
- [Alcance] No se crean tablas, endpoints, permisos ni dependencias nuevas; se reutiliza el almacenamiento `/uploads/empresa_logos`.

## [2026-05-13] Indicadores economicos legibles en movil
- [Frontend] `web/administrar_empresa/panel.html` elimina recortes de texto en la seccion `Indicadores economicos importantes` para pantallas moviles.
- [UX] En celulares las tarjetas usan 2 columnas legibles y pasan a 1 columna en pantallas muy estrechas o bajas; nombre, valor, referencia y variacion quedan completos.
- [Alcance] Cambio solo responsive del panel empresarial; no modifica datos, endpoints, permisos, tablas ni dependencias.
- [QA] Validado con parseo de scripts inline y prueba visual Playwright en 390x844 y 360x640 sin elementos truncados.

## [2026-05-13] Facturacion electronica enlaza proveedores de firma digital
- [Frontend] `web/administrar_empresa/facturacion_electronica.html` agrega el boton `Adquirir Firma Electronica` junto a `Cargar firma` en la tarjeta de carga DIAN Colombia.
- [Frontend] `web/administrar_empresa/proveedores_firma_digital.html` publica la pagina `Proveedores de Firma Digital` dentro del submenu de facturacion electronica.
- [Proveedor externo] La pagina agrega a Sensiyo como opcion para adquirir certificado digital/firma electronica DIAN y abre la compra en `https://sensiyo.co/certificados-digitales/`.
- [Alcance] Cambio solo de navegacion/frontend; no modifica endpoints, tablas, permisos ni dependencias.

## [2026-05-13] Corte automatico de turno en Caja
- [Frontend] `web/administrar_empresa/corte_de_caja.html` agrega el boton `Corte automatico`, que calcula el periodo desde la apertura de la caja abierta del usuario actual hasta la hora del corte.
- [Operacion] Al entrar desde `Estaciones -> Caja` con `auto_generar=1`, el corte usa `mi_caja_actual` y autocompleta fecha inicial, fecha final, usuario, caja, turno y efectivo de apertura.
- [Backend] `cerrarCorteCaja` cierra la caja abierta existente cuando recibe `cierre_caja_id`, evitando crear un cierre duplicado para el mismo turno.
- [QA] Parseo JS inline, `go test ./handlers -count=1`, `go test ./db -run "CierreCaja|Finanzas|Carritos" -count=1`, `git diff --check` y prueba visual Playwright con servidor mock.

## [2026-05-13] Ultimos movimientos de caja por usuario actual
- [Backend] `backend/handlers/corte_caja.go` agrega el alcance `mi_caja_actual` para resolver solo la caja abierta del usuario autenticado y filtrar ventas/productos/movimientos por `cierre_caja_id`, `caja_codigo` y usuario.
- [Frontend] `web/administrar_empresa/estaciones.html` envia `solo_usuario_actual=1` desde el boton `Ver ultimos movimientos`; `web/administrar_empresa/ultimos_movimientos_de_caja.html` consume el corte de caja acotado en vez de listados globales.
- [Seguridad] La URL ya no permite mezclar movimientos de cajas de otros usuarios; se conserva aislamiento por `empresa_id` sin agregar tablas, permisos ni dependencias.
- [QA] `go test ./handlers -count=1`, `go test ./db -run "CierreCaja|Finanzas|Carritos" -count=1`, parseo JS inline con Node y `git diff --check`.

## [2026-05-13] Caja recupera tarjeta compacta con boton de movimientos
- [Frontend] `web/administrar_empresa/estaciones.html` vuelve a mostrar la tarjeta `Caja` como bloque compacto con titulo, totales y boton `Ver ultimos movimientos`, sin el texto descriptivo largo.
- [UX] La tarjeta `Caja` toma el color configurado de estaciones mediante `--station-state-bg`, conservando el boton independiente para ultimos movimientos.
- [Operacion] El clic o teclado sobre la tarjeta completa sigue abriendo `corte_de_caja.html`; solo el boton interno navega a `ultimos_movimientos_de_caja.html`.

## [2026-05-13] Carritos recupera esquema incompleto antes de responder 500
- [Backend] `backend/db/carritos_compras.go` valida y completa todas las columnas requeridas por el listado de carritos, items y metricas de estacion antes de marcar el esquema como listo, con cache por base/esquema PostgreSQL.
- [Backend] El reintento reducido del listado ya no selecciona `ic.item_count` cuando se desactiva el join de items, evitando otro 500 en bases con tablas auxiliares rezagadas.
- [Operacion] Si `/api/empresa/carritos_compra` encuentra columnas o relaciones faltantes en una base empresarial rezagada, refresca migraciones ligeras y reintenta la consulta antes de mostrar error en estaciones.
- [QA] Se agrega `backend/db/carritos_compras_schema_test.go`, pasan pruebas enfocadas de `db` y `handlers`, y se valida visualmente el flujo Estaciones -> Zona 1 -> carrito sin mostrar el error inicial.

## [2026-05-13] Juegos moviles, sonido y records globales
- [Backend] `backend/main.go` registra `EnsureSuperJuegosSchema` y `/api/public/juegos/records` para guardar/consultar records en `super_juegos_records`.
- [Frontend] `web/Juegos/juegos_records.js`, `arcade_embed.js` y `open_game_embed.js` agregan ranking, reporte automatico de puntajes, panel de records, controles tactiles y sonido WebAudio.
- [UX] `web/Juegos/menu_juegos.html` muestra tarjetas uniformes con capturas PNG reales de cada juego en `web/img/juegos/`.
- [Emulador] `/emulador/` suma D-pad tactil y `/Juegos/n64/index.html` embebe el emulador real en una pantalla responsive.
- [QA] Validado con `go test ./...`, `node --check` y Playwright movil/escritorio sobre menu, juegos, N64 y emulador.

## [2026-05-13] Caja en estaciones deja de caer al flujo de carrito
- [Frontend] `web/administrar_empresa/estaciones.html` agrega una guarda explicita para la tarjeta `Caja`, de modo que nunca use la activacion generica de estaciones ni abra `carrito_de_compras.html`.
- [Operacion] Clic o teclado sobre `Caja` siempre abren `corte_de_caja.html` para cierre de turno, corte e impresion del reporte del cajero actual.

## [2026-05-13] Estaciones vuelven a abrir carritos legado por nombre
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` ya no restringe la carga inicial del modo estación al filtro `estacion_id`, para poder encontrar carritos previos de la misma empresa que todavía solo coincidían por nombre.
- [Frontend] Si el carrito rescatado pertenece a la estación pero quedó con identidad vieja, la misma apertura lo actualiza a `EST-{empresa}-{estacion}` y `ESTACION_{id}` antes de activar o recuperar la sesión.
- [Operación] Se elimina el escenario donde una estación como `Zona 1` intentaba recrear un carrito con nombre duplicado y terminaba mostrando `No se pudo abrir esta estación`.

## [2026-05-13] Administrar empresa exige sesion valida antes de abrir modulos
- [Frontend] `web/js/administrar_empresa.js` valida `/me` al arrancar el shell empresarial y redirige a `login.html` cuando la sesion ya expiro o no existe, en vez de dejar cargar menus e iframes protegidos como si hubiera acceso.
- [Frontend] `web/administrar_empresa/carrito_de_compras.html`, `web/administrar_empresa/administrar_clientes.html` y `web/administrar_empresa/bodega.html` convierten `401` en redireccion al login administrativo y `403` en mensaje claro de permiso insuficiente.
- [QA] La verificacion visual sobre `administrar_empresa.html` y modulos de `Carritos`, `Venta directa`, `Clientes` y `Bodegas` confirmo que el problema observable provenia de respuestas `401` del API, no de una caida general del frontend.

## [2026-05-13] Login con imagenes reales y glow adaptable
- [Frontend] `web/login.html` y `web/login_usuario.html` usan imagenes reales PNG en lugar de las ilustraciones previas.
- [Assets] Se agregan `web/img/login-admin-real.png` y `web/img/login-usuario-real.png` como copias estables para despliegue web sin espacios en el nombre.
- [UX] `web/estilos.css` incorpora un iluminado exterior suave que responde bien a apariencia clara y oscura en ambos logins.

## [2026-05-13] Caja en Estaciones abre cierre de turno
- [Estaciones] La tarjeta especial `Caja` en `web/administrar_empresa/estaciones.html` deja de lanzar `carrito_de_compras.html` en modo venta directa.
- [Corte de caja] Al entrar desde estaciones, `web/administrar_empresa/corte_de_caja.html` muestra contexto de caja, boton `Cerrar turno`, enlace `Regresar a estaciones`, auto-generacion del reporte y filtro por `caja_codigo` de la empresa.
- [QA] Validacion visual con Playwright en servidor mock local, confirmando navegacion `Caja -> corte_de_caja.html` y carga del reporte del cajero.

## [2026-05-13] Carritos y venta directa vuelven a cargar en Administrar empresa
- [Backend] `backend/handlers/carritos_compras.go` vuelve a asegurar el esquema antes del listado y `backend/db/carritos_compras.go` reintenta la consulta sin joins opcionales cuando una base vieja todavia no tiene `clientes` o `carrito_compra_items` alineados.
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` protege toda la secuencia de carga inicial con el mismo fallback visual, evitando iframes en blanco si el fallo ocurre antes de `loadCarritos()`.
- [QA] Verificacion visual con Playwright en `carrito_de_compras.html` y `venta_directa.html` sobre servidor mock local, mas `go test ./handlers` y `go test ./db` enfocados.

## [2026-05-13] Diagrama entidad relacion canonico
- [Diagramas] Se agrega `documentos/diagramas/diagrama_entidad_relacion.md` como DER resumido y vigente del proyecto.
- [Imagen] Se agrega `documentos/diagramas/diagrama_entidad_relacion.svg` para visualizacion rapida del modelo relacional.
- [Limpieza] `documentos/diagramas/estructura_del_codigo.md` deja de listar como vigentes diagramas historicos que ya no existen fisicamente.

## [2026-05-13] Documentacion y ayuda actualizadas
- [Docs] Se agrega `documentos/estado_documentacion_2026-05-13.md` como mapa consolidado del estado vigente del proyecto.
- [Indice] `documentos/README.md` y `documentos/descripcion_del_proyecto` quedan alineados con operacion conectada, cajas simultaneas, login global, comunicaciones, soporte, Docker/VPS y validacion.
- [Ayuda] `web/ayuda/ayuda.html` suma secciones de acceso de usuarios, operacion conectada, cajas simultaneas, soporte/comunicaciones, documentos locales y backups.

## [2026-05-13] Cajas simultaneas por licencia
- [Licencias] Se agrega `max_cajas_simultaneas` con default 2 y default 4 para planes de 4000 documentos.
- [Finanzas] Apertura y reapertura de cajas validan el cupo contra la licencia activa de la empresa.
- [Ventas] Cada pago de carrito se enlaza a una caja abierta para mantener cierres y arqueos separados.
- [Super] La pantalla de licencias permite configurar el maximo de cajas simultaneas.

## [2026-05-13] Conexion obligatoria para operacion y facturacion
- [Facturacion] `web/administrar_empresa/facturacion_electronica.html` retira controles de modo offline/contingencia DIAN y muestra conexion obligatoria.
- [Backend] `backend/handlers/facturacion_electronica.go` ignora banderas offline antiguas y bloquea la emision cuando DIAN/proveedor no esta disponible.
- [Operacion] Las ventas, cobros y facturacion quedan documentadas como flujos que requieren conexion activa con el servidor.

## [2026-05-13] Correos masivos globales
- [Super] Se agrega `web/super/correos_masivos.html` al menu del super administrador para enviar comunicados de politicas, actualizaciones, mantenimiento, seguridad o informacion general.
- [Backend] Nuevo endpoint `/super/api/correos_masivos` con validacion de rol super, vista previa con emails enmascarados, deduplicacion de destinatarios y confirmacion obligatoria.
- [DB] Nuevas tablas `super_correos_masivos` y `super_correos_masivos_destinatarios` en `pcs_superadministrador` para trazabilidad de campanas y resultado por destinatario.

## [2026-05-13] Proporcion de tarjetas en carrusel del index
- [Portal] `web/index.html` corrige la barra horizontal para que sus tarjetas no se estiren verticalmente por la tarjeta mas larga.
- [UX] Se controla el alto en escritorio y se limita la descripcion dentro del carrusel, manteniendo las tarjetas superiores sin cambios.
- [Alcance] Solo frontend; sin cambios de backend, tablas, permisos ni dependencias.

## [2026-05-13] Propinas y comisiones enlazadas a usuarios
- [Datos] `empresa_propinas_movimientos` y `empresa_comisiones_servicio_movimientos` guardan ids de `users` para origen, asignado y lavador.
- [Backend] Los movimientos automaticos/manuales resuelven usuarios por id, correo, documento, nombre o etiqueta `Nombre (correo)` dentro del mismo `empresa_id`.
- [UX] Propinas, comisiones y carrito sugieren usuarios creados y muestran el id vinculado en reportes.
- [Validacion] `go test ./...` en `backend/` y parseo de scripts inline HTML con Node.

## [2026-05-13] Backup profesional junto a Configuracion
- [Empresa] `web/administrar_empresa.html` mueve el acceso de backups al grupo `Administracion`, inmediatamente junto a `Configuracion`.
- [Permisos] Se conserva `linkBackups`, `module=backups` y la misma regla existente de permisos; no hay cambios de backend, tablas ni dependencias.

## [2026-05-13] Imagen profesional en login de usuarios
- [UX] `login_usuario.html` agrega una ilustracion de usuario iniciando sesion frente a un computador.
- [Frontend] La imagen queda en la esquina superior derecha en escritorio y se adapta como visual compacto sobre el encabezado en movil.
- [Activo] Nuevo `web/img/login-usuario-computador.svg`; no hay cambios de backend, tablas, permisos ni dependencias.

## [2026-05-13] Login unico para usuarios de empresa
- [Auth] `login_usuario.html` opera como portal global para usuarios de todas las empresas, sin subdominio empresarial obligatorio.
- [Backend] El login resuelve la empresa por email y clave con consulta global por correo; si hay ambiguedad no abre una empresa arbitraria.
- [Invitaciones] El primer password se valida por `token_invitacion`, email y documento, y los enlaces ya no agregan `empresa_id`.
- [UX] La pantalla mantiene compatibilidad con enlaces antiguos, persiste `empresa_id` al autenticar y redirige a `administrar_empresa.html?id=...`.
- [Validacion] `go test ./...` en `backend/`, `node --check web/js/login_usuario.js` y `git diff --check`.

## [2026-05-13] Aviso de mantenimiento programado
- [Super] `web/super/configuracion_avanzada.html` agrega check para publicar aviso, fecha, hora inicio, hora fin, zona horaria y mensaje publico.
- [Backend] `/super/api/config/mantenimiento` conserva `mantenimiento_activo` y persiste `mantenimiento_programado.*`; `/api/empresa/mantenimiento_programado` expone el aviso activo bajo alcance de empresa.
- [Empresa] `web/administrar_empresa/panel.html` muestra una franja de aviso cuando el check esta activo.
- [Seguridad] El aviso no activa el bloqueo global de mantenimiento; no hay tablas ni dependencias nuevas.

## [2026-05-13] OnlyOffice local editable en una sola pantalla
- [Empresa] `web/administrar_empresa/documentos_onlyoffice.html` permite elegir tipo, crear documento, abrir OnlyOffice embebido en la misma vista y guardar el resultado en el PC/celular.
- [Backend] `/api/empresa/documentos?action=create_edit_local` crea una sesion temporal editable y `/api/empresa/documentos?action=download&delete=1` descarga el archivo final eliminando la copia temporal del VPS.
- [Compatibilidad] `documentos_onlyoffice_menu.html` redirige a la pantalla unica y el menu de Administrar empresa apunta directo al nuevo flujo.
- [Seguridad] Sin tablas ni permisos nuevos; se conserva `empresa_id`, saneamiento de nombres y el wrapper de permisos de `documentos_onlyoffice`.

## [2026-05-13] Alertas de vencimiento de licencias
- [Super] `/super/configuracion_avanzada.html` agrega configuracion para activar alertas, definir dias de aviso, revisar pendientes y ejecutar envio manual.
- [Backend] Nuevo `/super/api/licencias/vencimiento_alertas`, worker periodico cada 12 horas y plantilla `licencia_expiry_warning`.
- [Datos] `licencia_vencimiento_notificaciones` registra envios/capturas y evita duplicados por licencia, empresa, correo, fecha y umbral.
- [Validacion] `go test ./...` en `backend/` OK.

## [2026-05-12] SSH VPS cambiado a 49222
- [Operacion VPS] SSH queda escuchando solo en `49222`; el puerto `22` fue cerrado despues de probar conexion externa real por el puerto nuevo.
- [Seguridad] `ssh.socket` queda desactivado y `ssh.service` gestiona el listener para evitar el bloqueo ocurrido anteriormente con socket activation.
- [Deploy] `scripts/pcs_deployment.local.ps1`, `scripts/pcs_deployment.local.ps1.example`, tuneles locales, RustDesk remoto y escaner VPS usan `49222`.

## [2026-05-13] Submenu de Configuracion super alineado
- [UX] `/super/configuracion_avanzada.html` adopta el patron visual del sidebar de `seleccionar_empresa.html` para su submenu interno.
- [Frontend] El submenu usa titulo simple, ancho estable, botones compactos, estados activo/hover coherentes y colapso movil.
- [Alcance] No cambia backend, endpoints, permisos, configuraciones guardadas, base de datos ni dependencias.

## [2026-05-13] Apariencia claro/oscuro del Explorador de Archivos super
- [Frontend] `/super/explorador_archivos.html` inicializa tema desde `localStorage`/cookie `pcs_theme`.
- [UX] Fondos, tarjetas, tabla, botones, input de ruta, estados y hover ahora usan variables globales compatibles con claro/oscuro.
- [Alcance] No cambia backend, endpoint, permisos, filesystem, base de datos ni dependencias.

## [2026-05-12] Tickets de ayuda empresariales profesionalizados
- [Empresa] El menu flotante crea tickets con contacto preferido, telefono opcional, contexto tecnico seguro de la pantalla activa y tickets recientes de la empresa.
- [Backend] `/api/empresa/tickets_ayuda` permite detalle y comentarios propios con validacion por `empresa_id`; `super_tickets_ayuda` guarda contacto y contexto tecnico.
- [Super] `/super/tickets_ayuda.html` muestra categoria, modulo, ruta, contacto y diagnostico para triage profesional.
- [Seguridad] Las notas internas no se devuelven a empresas y el contexto no incluye cookies, localStorage, tokens ni secretos.

## [2026-05-12] Retiro de mesa de ayuda legado
- [Backend] Se elimina la ruta empresarial heredada y su wrapper de permisos; el soporte oficial queda en `/api/empresa/tickets_ayuda`.
- [Frontend] Se retiran la pagina, el enlace de menu y la opcion de licencias del modulo legado.
- [Producto] Los tickets propios quedan centralizados en `super_tickets_ayuda` y `super_ticket_ayuda_mensajes`, con bandeja exclusiva del super administrador.

## [2026-05-12] Mesa central de tickets de ayuda
- [Backend] Se agregan `/api/empresa/tickets_ayuda` y `/super/api/tickets_ayuda` con almacenamiento PostgreSQL en `super_tickets_ayuda` y `super_ticket_ayuda_mensajes`.
- [Empresa] El menu flotante global suma `Crear ticket de ayuda`, detecta empresa activa, modulo y ruta para enviar la solicitud al soporte del SaaS.
- [Super] `web/super/tickets_ayuda.html` permite filtrar, responder, cambiar estado/prioridad, asignar responsable y cerrar tickets.
- [Seguridad] La creacion valida alcance por `empresa_id`; la bandeja central queda exclusiva de `super_administrador`.

## [2026-05-12] Carrusel index con tarjetas iguales
- [Frontend] `web/index.html` alinea el ancho de las tarjetas del carrusel con la grilla superior.
- [Responsive] El carrusel usa 3 tarjetas por vista en escritorio, 2 en tablet y 1 en movil, igualando la lectura visual de la grilla.
- [UX] Las flechas avanzan una tarjeta usando el gap real calculado por CSS.

## [2026-05-12] Paquete vpssecurity/logstore en Docker
- [Docker] El paquete Go de almacenamiento del escaner VPS se mueve a `backend/vpssecurity/logstore`.
- [Correccion] El backend Docker deja de confundir codigo Go con carpetas runtime llamadas `logs` ignoradas por `.dockerignore`.

## [2026-05-12] Limpieza de codigo backend remoto
- [Deploy] `scripts/sync_to_vps.ps1` borra codigo fuente backend obsoleto antes de extraer el paquete en el VPS.
- [Proteccion] Conserva `.env`, `.env.local`, `logs`, `bin`, `tmp` y `secure`.
- [Correccion] Evita builds Docker rotos por archivos eliminados localmente que seguian presentes en el VPS.

## [2026-05-12] SSH VPS restablecido temporalmente a 22
- [Operacion VPS] Hostinger quedo restablecido temporalmente para SSH en el puerto `22` antes del cambio seguro posterior a `49222`.

## [2026-05-12] Index con carrusel de sistemas
- [Frontend] `web/index.html` renderiza las primeras 6 tarjetas en la grilla principal y las demas en una hilera horizontal navegable.
- [UX] La barra de tarjetas restantes queda antes de `Modulos del sistema` y se controla con flechas izquierda/derecha.
- [Alcance] Sin cambios de backend, permisos, tablas ni dependencias.

## [2026-05-12] Fotos del index en landing descriptiva
- [Frontend] `web/descripcion_de_los_sistemas.html` ahora usa `imagen_secundaria_url` como imagen principal de cada tarjeta descriptiva, igual que las tarjetas del index.
- [UX] La landing conserva el logo pequeno de cada sistema como apoyo, pero la foto/ilustracion visible queda sincronizada con el portal principal.
- [Alcance] Sin cambios de backend, permisos, tablas ni dependencias.

## [2026-05-12] VPS portable 100% Docker
- [Deploy] `deploy/docker-compose.platform.yml` agrega perfil `edge` con Nginx publico para `80/443` y perfil `certbot` para certificados Let's Encrypt.
- [Operacion] Nuevos scripts `deploy/scripts/vps-docker-edge-up.sh` y `deploy/scripts/vps-docker-edge-renew.sh` permiten mover TLS/Nginx publico al stack Docker y renovar certificados sin depender de Nginx del host.
- [Portabilidad] Se agregan volumenes `pcs_letsencrypt` y `pcs_certbot_www`; migrar un VPS futuro requiere proyecto/imagenes, `deploy/.env.platform` y volumenes Docker.
- [Limpieza] Las plantillas `.env` y scripts de staging dejan de generar variables Nextcloud legacy.

## [2026-05-12] Explorador de Archivos super
- [Super] `web/super/explorador_archivos.html` agrega una vista tipo explorador de Windows para navegar carpetas del filesystem visible para el backend en el VPS.
- [Backend] Nuevo `GET /super/api/explorador_archivos?action=list&path=...`, protegido con `paginaPrincipalRequireSuperAdmin`, devuelve raices, ruta actual, ruta padre y metadata de archivos/carpetas sin leer contenido.
- [Seguridad] La operacion queda en modo solo lectura y no expone acciones de borrar, editar, subir, descargar ni abrir archivos; sin tablas ni dependencias nuevas.

## [2026-05-12] Retiro de Nextcloud y cuota DB por empresa
- [Backend] Se retiran rutas y handlers Nextcloud; el arranque elimina la tabla legacy `empresa_nextcloud_accounts` y las claves super `nextcloud.*`.
- [Super] `Limitaciones por empresa` cambia la cuota de almacenamiento por `db_max_gb`, usada como tamano maximo de base de datos por empresa y visible en el panel PostgreSQL por empresa.
- [Deploy] Nextcloud sale del Compose oficial; `deploy/scripts/vps-remove-nextcloud.sh` apaga contenedores legacy y puede purgar volumenes con confirmacion operativa.

## [2026-05-12] Documentos y backups locales por dispositivo
- [OnlyOffice] `web/administrar_empresa/documentos_onlyoffice.html` crea documentos en modo local por defecto; `backend/handlers/onlyoffice.go` devuelve el archivo como descarga con `action=create_local` sin guardarlo en el VPS.
- [Backups] `web/administrar_empresa/backups.html` agrega backup automatico local por navegador y descarga directa al equipo/celular.
- [Backend] `/api/empresa/backups` suma `exportar_local` y `exportar_configuracion_local`, construyendo JSON en memoria sin crear historial ni archivos en disco.

## [2026-05-12] Logos configurables en documentos empresariales
- [Backend] `empresa_configuracion_avanzada` agrega `mostrar_logo_empresa` y `mostrar_logo_sistema`, manteniendo `mostrar_logo` como compatibilidad general.
- [Frontend] Configuracion de empresa muestra checks separados para activar el logo empresarial o el logo del sistema en facturas, recibos, reportes y documentos imprimibles.
- [Impresion] `web/js/print_documents.js` soporta varios logos por documento; facturas, ventas y carrito pasan la configuracion avanzada al motor comun.

## [2026-05-12] Regreso a estaciones en ultimos movimientos de caja
- [Frontend] `web/administrar_empresa/ultimos_movimientos_de_caja.html` agrega `Regresar a estaciones` en el encabezado.
- [UX] El enlace conserva `empresa_id` para volver a `estaciones.html` dentro de la misma empresa.
- [Alcance] Sin backend, permisos, tablas ni dependencias nuevas.

## [2026-05-12] Login usuario simplificado y reenvio de invitacion
- [Frontend] `web/login_usuario.html` retira controles y enlaces no necesarios del login operativo: apariencia manual, inicio, acceso admin, completar/cambiar/volver redundantes.
- [UX] El formulario principal conserva solo `¿Olvidó su contraseña?` y `Recuperar email de invitación`.
- [Backend] Nuevo `POST /api/empresa/usuarios/recuperar_invitacion`, con reCAPTCHA y respuesta enmascarada, reenvia invitacion si el usuario ya fue creado por el administrador y no completo password.
- [Seguridad] El reenvio rota el token de invitacion por 48 horas y no revela existencia de correos.

## [2026-05-12] Registro operativo solo por invitacion
- [Backend] `EmpresaUsuarioSetPasswordHandler` ahora exige `token_invitacion` para el primer registro, valida expiracion y consume el token al guardar la contrasena.
- [Correo] Las invitaciones apuntan a `login_usuario.html?token_invitacion=...&modo=registro`; la ruta legacy `/auth/confirmar_correo` redirige al mismo flujo sin confirmar por separado.
- [Sesion] El primer ingreso confirma correo, abre sesion, devuelve `empresa_id`, `rol` y `redirect_url` hacia `administrar_empresa.html?id=...`.
- [Permisos] `administrar_empresa` conserva `/api/empresa/permisos_contexto` como fuente efectiva para ocultar/ver acciones por rol, licencia y reglas finas.
- [Alcance] Sin dependencias nuevas ni columnas nuevas; se reutilizan `email_confirm_token`, `email_confirm_expira` y `email_confirmado`.

## [2026-05-12] Tema del login usuario desde cookies
- [Frontend] `web/login_usuario.html` lee primero la cookie `pcs_theme` antes de caer a `localStorage`, aplicando el tema antes de cargar estilos.
- [Tema global] `web/menu.js` tambien prioriza `pcs_theme` para que el gestor compartido no sobrescriba el login con un valor local antiguo.
- [UX] El login de usuarios operativos conserva la apariencia clara/oscura del ultimo usuario que inicio sesion o cambio tema en el navegador.
- [Alcance] Sin cambios de API, permisos, backend, base de datos ni dependencias.

## [2026-05-12] Ayuda interna en CRM unificado
- [Menu] `web/administrar_empresa/modulo_menu.html` agrega `Ayuda` como ultimo boton del menu de `crm_unificado`.
- [Frontend] `web/administrar_empresa/crm_comercial.html` agrega una pestana `Ayuda` que primero define que es CRM y despues explica el uso del modulo.
- [UX] La ayuda cubre tablero ejecutivo, leads, seguimientos, cotizaciones, forecast, metas, embudo documental y la regla de no duplicar clientes/ventas/facturacion.
- [Alcance] Sin cambios de API, permisos, backend, base de datos ni dependencias.

## [2026-05-12] Totales del carrito por tipo de empresa
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` adapta la tarjeta `Totales y detalles` con perfiles de negocio: estadia, gimnasio, clinico, pedido, transporte, parqueadero, alquiler, copropiedad, farmacia, belleza, orden de servicio, academico, obra y verticales nuevos.
- [Contexto] La pagina consulta `/api/empresa/permisos_contexto` para leer `vertical_scope` y usa configuracion general como respaldo; si el contexto no llega, conserva el perfil POS universal.
- [UX] `web/estilos.css` agrega un indicador compacto de perfil sin cambiar el ancho de la tarjeta ni el flujo de pago.
- [Alcance] Sin cambios de API, permisos, backend, base de datos ni dependencias.

## [2026-05-12] Gimnasio empresarial
- [Frontend] `web/administrar_empresa/gimnasio.html` suma un mando ejecutivo con salud operativa, alertas, acciones rapidas y logo del vertical.
- [UX] `web/js/gimnasio.js` calcula riesgos desde socios, planes, clases, pagos, dispositivos y bitacora de acceso; agrega busqueda/filtro global y badges de estado en tablas.
- [Menu] `web/administrar_empresa/gimnasio_menu.html` organiza el modulo por Direccion, Comercial, Operacion deportiva y Acceso, usando el patron agrupado del panel empresarial.
- [Alcance] Sin cambios de API, permisos, backend, base de datos ni dependencias.

## [2026-05-12] Menu de configuracion super alineado
- [Frontend] `web/super/configuracion_avanzada.html` usa el mismo patron visual de menu lateral agrupado que Administrar empresa: `admin-sidebar`, `admin-nav-grouped`, grupos plegables, iconos y boton de colapso movil.
- [UX] El buscador de secciones se conserva y ahora abre/oculta grupos completos segun coincidencias, manteniendo scroll interno a cada tarjeta.
- [Alcance] Sin cambios de API, permisos, backend, base de datos ni dependencias.

## [2026-05-12] Scroll superior en Tipos de empresa
- [Frontend] `web/super/tipos_empresas.html` agrega un scroll horizontal arriba de la tabla y lo sincroniza con el scroll inferior.
- [Alcance] Sin cambios de API, permisos, backend, base de datos ni dependencias.

## [2026-05-12] Verticales 20 actualizado
- [Frontend] `web/super/verticales_produccion_masiva.html` muestra semaforo ejecutivo, brechas principales, tarjetas de foco comercial y KPIs de licencias/base/readiness.
- [Operacion] Se agregan filtros `Sin licencia` y `Sin preconfig` para priorizar tareas antes de publicar comercialmente un vertical.
- [Contrato] El cruce de readiness acepta payloads con `items`, `licencias` o arreglos directos y valida catalogo, metadata, preconfiguracion activa y licencia base.
- [Alcance] Sin endpoints, permisos, tablas ni dependencias nuevas.

## [2026-05-12] Limpieza visual del super administrador
- [Frontend] `web/super_administrador.html` retira los indicadores compactos `PostgreSQL`, `VPS`, `Licencias` y `Seguridad` del encabezado lateral.
- [Alcance] Sin cambios de rutas, iframes, permisos, backend, base de datos ni dependencias.

## [2026-05-12] Identidad visual empresarial
- [Frontend] `web/administrar_empresa/configuracion.html` agrega una seccion visible para cargar el logo de la empresa y sincronizarlo con la configuracion de factura/documentos.
- [Panel] `web/administrar_empresa/panel.html` muestra el logo encima de `Panel de <empresa>` con tamano fijo, sin modificar la columna ni dimensiones de la tarjeta de clima.
- [Backend] `/api/empresa/configuracion_avanzada/logo` acepta PNG/JPG/WEBP/GIF hasta 5 MB, guarda en `/uploads/empresa_logos/empresa_<id>/` y persiste `logo_url` + `mostrar_logo`.
- [Datos] No hay tablas nuevas; se reutiliza `empresa_configuracion_avanzada` como fuente unica para panel, factura y documentos.

## [2026-05-12] CRM empresarial profesional
- [Backend] `GET /api/empresa/crm_avanzado` con `action=dashboard` suma salud comercial, valor en riesgo, leads sin contacto, oportunidades estancadas, acciones priorizadas, responsables y canales.
- [Frontend] `web/administrar_empresa/crm_comercial.html` y `web/js/crm_comercial.js` muestran cockpit ejecutivo, plan de accion, responsables y canales dentro de CRM unificado.
- [Permisos] Los CRUD `/api/empresa/crm/*` quedan bajo `WithEmpresaCRMUnificadoPermissions`.
- [Alcance] Sin tablas, rutas publicas ni dependencias nuevas.

## [2026-05-12] Nucleo configurable por plantilla
- [Backend] `tipo_empresa_preconfiguraciones.config_json` normaliza `adaptacion_nucleo` para usuarios operativos, productos/servicios y estaciones como recursos del negocio.
- [Aplicacion] Al aplicar una plantilla se guarda `preconfiguracion_tipo_empresa_adaptacion_nucleo` y `estaciones_config` incluye `tipo_recurso`, `tipo_recurso_plural` y `representa_recurso_negocio`.
- [Super] `web/super/preconfiguracion_tipos_empresa.html` muestra la seccion `Nucleo configurable` y envia la metadata al guardar.
- [Alcance] Sin tablas, columnas, permisos ni dependencias nuevas.

## [2026-05-12] Matriz profesional de 30 verticales
- [Backend] `/api/*/verticales_integracion/catalogo` publica exactamente 30 verticales canonicos: 10 clasicos reales y 20 nuevos. `consultorio_odontologico` se fusiona en `odontologia`, `taxi` en `taxi_system` y `turnos_atencion`/`turnos` quedan como soporte transversal.
- [Contrato] Cada vertical calcula `professional_ready`, `readiness_score`, checks, alcance de configuracion, ingresos, egresos, tablas financieras y metadata de `fused_modules`, `support_modules` o `similar_templates` cuando aplica.
- [Frontend] `web/administrar_empresa/verticales_integracion.html` muestra perfil activo por empresa, contrato, ventas/finanzas, reportes y configuracion; cada vertical queda marcado como `Profesional` o `Brecha`.
- [Finanzas] Los 30 verticales quedan atados a `empresa_finanzas_movimientos` y a los modulos centrales de ventas, pagos, bancos/tesoreria y reportes para ingresos y egresos.
- [Configuracion] El acceso conserva `linkVerticalesIntegracion` y permiso `seguridad:R`, pero se presenta como `Adaptacion por tipo`.
- [Alcance] No hay tablas, endpoints de escritura, permisos ni dependencias nuevas.

## [2026-05-12] Licencias ocultables para clientes
- [Super] `web/super/licencias.html` expone la bandera como `Visibilidad comercial`: visible u oculta para clientes.
- [Backend] Los checkouts públicos de licencia rechazan licencias con `activo=0`, incluyendo resumen público, Wompi, Nequi, Epayco, activación sin pago y addons seleccionados manualmente.
- [Datos] Reutiliza `licencias.activo`; no agrega tablas, columnas ni dependencias.

## [2026-05-12] Indicadores economicos compactos en panel empresarial
- [Frontend] `web/administrar_empresa/panel.html` presenta los indicadores de mercado en una tabla compacta de dos indicadores por fila en escritorio.
- [Responsive] En movil conserva las tarjetas reducidas existentes para evitar desbordes horizontales.
- [Alcance] Sin cambios de API, permisos, base de datos ni dependencias.

## [2026-05-12] Enlace Probar Gratis del index
- [Frontend] `web/index.html` cambia el destino de ficha comercial a `/descripcion_de_los_sistemas.html` conservando `accion=probar_gratis`, `tipo_empresa`, modulo, secciones y ancla de la tarjeta elegida.
- [Backend] `backend/main.go` atiende la ruta legacy `/descripcion_de_los_sistemas.ht` sirviendo la version `.html` con `Content-Type: text/html`, evitando que navegadores la descarguen.
- [Seguridad] `AuthMiddleware` mantiene publicas ambas rutas descriptivas; no se abren rutas privadas ni permisos nuevos.
- [Operacion] `web/super/pagina_principal.html` y la auditoria profesional usan la ruta oficial `.html`.

## [2026-05-12] Apariencia claro/oscuro del Centro de mando
- [Frontend] `web/super/licencias_resumen.html` usa las variables globales `--bg`, `--surface`, `--surface-soft`, `--text`, `--muted`, `--border`, `--accent` y `--accent-2` para integrarse con temas claros y oscuros.
- [UX] Se corrigen fondos, bordes, botones, pills de estado, tablas, graficas SVG y aro de score para evitar contraste roto al cambiar de tema.
- [Alcance] Sin cambios de API, permisos, base de datos ni dependencias.

## [2026-05-12] Centro de mando super reconstruido
- [Frontend] `web/super/licencias_resumen.html` se reemplaza completo por una consola ejecutiva responsive con score operativo, KPIs de plataforma, PostgreSQL, seguridad, negocio SaaS, costos, SLO, riesgos, servicios e incidentes.
- [Operacion] La vista agrega controles directos para actualizar, evaluar alertas y abrir gobierno de alertas, PostgreSQL, seguridad VPS, licencias, empresas, tipos de empresa, roles, verticales, IA, configuracion y reportes.
- [Alcance] Reutiliza APIs existentes del panel super; no agrega endpoints, permisos, tablas, dependencias ni cambios en `go.mod`.
- [QA] Parseo de script inline con Node OK; `git diff --check -- web/super/licencias_resumen.html` sin errores.

## [2026-05-11] Matriz de integracion en configuracion empresarial
- [Menu] `Matriz de integracion` sale de Soluciones por negocio y queda en Administrar empresa > Configuracion > Base empresarial.
- [Permisos] `linkVerticalesIntegracion` conserva `seguridad:R`, ahora agrupado como Administracion y configuracion.

## [2026-05-11] Emisora online por empresa
- [Backend] `/api/chat_flotante/preferencias` acepta `empresa_id` y persiste `chat_flotante.*`, incluida `radio_online_enabled`, en `empresa_estacion_prefs`.
- [Frontend] `Configurar chat y robot` agrega el check `Activar emisora online`; el panel compacto del chat y `radio_player.js` sincronizan el reproductor flotante con esa preferencia.
- [QA] `node --check web/js/ai_chat_drawer.js`; `node --check web/js/radio_player.js`; `go test ./...` en `backend/`.

## [2026-05-11] Alcance vertical por licencia
- [Backend] `/api/empresa/permisos_contexto` calcula `vertical_scope` desde tipo/preconfiguracion/licencia y desactiva acciones de verticales ajenos sin tocar el nucleo universal.
- [Licencias] El checkout, activacion manual/gratuita y confirmaciones de pago validan que la licencia base corresponda al tipo de empresa elegido.
- [Frontend] `elegir_licencia.html` consulta licencias filtradas por `tipo_id` y `editar_empresa.js` conserva `tipo_id/tipo_nombre` al renovar.
- [QA] `go test ./handlers`; `go test ./db`.

## [2026-05-11] 2FA del login desde configuracion avanzada
- [Seguridad] El login de administradores oculta el campo de codigo 2FA salvo que `security.admin_2fa.enabled` este activo.
- [Backend] `/config.js` publica `ADMIN_2FA_LOGIN_ENABLED` y `AdminLoginHandler` solo exige OTP cuando el switch global y el TOTP de la cuenta estan activos.
- [Frontend] `web/super/configuracion_avanzada.html` agrega la tarjeta `2FA login` para activar/desactivar la exigencia global sin tocar secretos por cuenta.
- [QA] `go test ./handlers -run "TestAdminTOTPLoginRequiredForAdmin" -count=1`; `go test ./... -count=1`; validacion JS de `login.js` y scripts inline.

## [2026-05-11] Catalogos publicos de verticales sin sesion
- [Seguridad] `backend/utils/utils.go` agrega a la lista publica `/api/public/verticales_nuevos/catalogo` y `/api/public/verticales_integracion/catalogo`.
- [Producto] La portada publica y las fichas comerciales pueden consultar el catalogo real de verticales sin depender de una sesion administrativa.
- [QA] `backend/utils/auth_middleware_test.go` valida que ambas rutas pasen sin cookie y que las rutas privadas sigan protegidas.

## [2026-05-11] Sincronizacion idempotente de pagos verticales
- [Backend] `backend/db/odontologia.go` y `backend/db/gimnasio.go` reutilizan `carritos_compras.referencia_externa` antes de crear carritos desde pagos historicos.
- [QA] Se agregan pruebas para fijar la llave historica de pagos en odontologia y gimnasio.
- [Alcance] No hay tablas, endpoints, permisos ni dependencias nuevas.

## [2026-05-11] Correccion de cargas parciales en verticales integrados
- [Backend] `backend/db/odontologia.go` y `backend/db/gimnasio.go` crean indices de integracion solo despues de asegurar columnas nuevas en bases PostgreSQL existentes.
- [Frontend] `web/js/consultorio_odontologico.js`, `web/js/gimnasio.js` y `web/js/alquileres.js` limpian el aviso de carga parcial cuando la recarga completa no devuelve errores.
- [QA] Se agregan pruebas para impedir que los indices de `cliente_id`, `servicio_id` y `carrito_id` vuelvan a ejecutarse antes de las columnas.
- [Alcance] No hay tablas, endpoints, permisos ni dependencias nuevas.

## [2026-05-11] Fix arranque PostgreSQL parqueadero
- [Backend] `backend/db/parqueadero.go` ahora asegura columnas de integracion al nucleo antes de crear el indice por `carrito_id`.
- [Operacion] Corrige el fallo de despliegue en VPS con bases existentes que todavia no tenian `empresa_parqueadero_tickets.carrito_id`.
- [Alcance] No hay nuevas tablas, endpoints, permisos ni dependencias.

## [2026-05-11] Consistencia del panel super
- [Frontend] `web/js/super_administrador.js` permite restaurar la ayuda privada del super administrador mediante validacion explicita de ruta.
- [Seguridad] La ayuda privada no queda dentro de la lista limitada del rol `control_super_administrador`.
- [QA] Alinea el contrato esperado por `TestSuperAdminPanelExposesPrivateHelpButton`.

## [2026-05-11] 20 verticales nuevos reales
- [Backend] `backend/db/nuevos_verticales_bootstrap.go` promueve los 20 verticales nuevos a produccion masiva con ranking 1-20.
- [API] `/super/api/verticales_nuevos/catalogo` acepta `asegurar_20_licencias` y conserva `asegurar_v1_licencias` como alias compatible.
- [Frontend] `web/js/nuevos_verticales_catalogo.js`, `web/index.html` y `web/super/verticales_produccion_masiva.html` publican y gobiernan las 20 plantillas reales.
- [QA] Las pruebas actualizadas exigen 20 verticales masivos, metadata extendida y decision de produccion masiva en nuevos verticales.
- [Alcance] No hay tablas, dependencias ni circuitos paralelos de clientes, productos, ventas o pagos.

## [2026-05-11] Portada index alineada a modulos reales
- [Frontend] `web/index.html` y los defaults de `/api/public/pagina_principal` actualizan el texto de cobertura y las tarjetas publicas con nucleo unico, modulos reales y verticales clasificados.
- [Producto] Los 20 verticales nuevos siguen en catalogo y quedan publicables como tarjetas operativas de `Probar gratis`.
- [Catalogo] `web/js/nuevos_verticales_catalogo.js` agrega decision, ranking, metadata de plantilla, permisos, flujo de venta y reportes para sincronizar la portada con la matriz extendida.
- [Alcance] No hay endpoints, tablas, permisos, dependencias ni cambios en `go.mod`.

## [2026-05-11] Aseguramiento comercial de verticales
- [Backend] `POST /super/api/verticales_nuevos/catalogoaction=asegurar_20_licencias` llama `EnsureNuevosVerticalesProduccionMasivaLicencias`; `asegurar_v1_licencias` queda como alias compatible.
- [Producto] La accion asegura tipos de empresa, preconfiguraciones y cuatro planes recomendados para los 20 verticales.
- [Frontend] `web/super/verticales_produccion_masiva.html` agrega `Asegurar 20` y refresca el semaforo despues de ejecutar.
- [Alcance] No hay tablas, rutas nuevas, permisos nuevos ni dependencias.

## [2026-05-11] Semaforo listo para venta en verticales
- [Frontend] `web/super/verticales_produccion_masiva.html` cruza verticales, preconfiguraciones y licencias activas para marcar `Listo venta`.
- [Regla] Un vertical queda listo solo si tiene metadata completa, preconfiguracion activa con `integracion_vertical` y licencia activa que incluye el modulo.
- [Alcance] No hay cambios de esquema, endpoints, permisos ni dependencias.

## [2026-05-11] Acciones de gobierno para verticales 20
- [Frontend] Cada fila de `web/super/verticales_produccion_masiva.html` enlaza a tipos, preconfiguraciones y licencias del vertical.
- [UX] `web/super/tipos_empresas.html`, `web/super/preconfiguracion_tipos_empresa.html` y `web/super/licencias.html` aplican filtros iniciales desde `q`, `vertical` o `modulo`.
- [Alcance] No se agregan endpoints, tablas, permisos ni dependencias.

## [2026-05-11] Gobierno super de verticales de produccion masiva
- [Frontend] Se agrega `web/super/verticales_produccion_masiva.html` con KPIs, filtros, ranking, decision, metadata extendida y exportacion CSV.
- [Menu] `web/super_administrador.html` incorpora `Verticales 20` dentro de Licencias y `web/js/super_administrador.js` permite restaurar la pagina.
- [Seguridad] Se reutiliza `/super/api/verticales_nuevos/catalogo`; no hay endpoints, permisos, esquemas ni dependencias nuevas.

## [2026-05-11] Preconfiguraciones y verticales de produccion masiva
- [Backend] `config_json` de tipos de empresa puede incluir `integracion_vertical` con decision, prioridad, permisos, flujo de venta, tablas y reportes.
- [Catalogos] Los endpoints de verticales nuevos publican `integracion_preconfig`, `produccion_masiva`, `prioridad_produccion` y `decision_preconfig`.
- [Producto] Se priorizan los 20 verticales nuevos para produccion masiva en `documentos/plan_verticales_produccion_masiva_2026-05-11.md`.
- [QA] Las pruebas exigen metadata extendida y exactamente 20 verticales marcados como produccion masiva; no hay cambios de esquema ni dependencias.

## [2026-05-11] Matriz extendida de plantillas verticales
- [Backend] El catalogo de integracion agrega `template_activates`, `tables_touched`, `required_permissions`, `sale_flow` y `reports_produced`.
- [Frontend] La matriz empresarial muestra modulos, plantilla, tablas, permisos, flujo de venta y reportes por vertical.
- [QA] La prueba de contrato impide publicar verticales visibles sin metadata completa; no hay cambios de esquema ni dependencias.

## [2026-05-11] Sincronizacion segura de matriz vertical
- [Frontend] La matriz consulta `/api/empresa/permisos_contexto`, calcula sincronizaciones permitidas, deshabilita botones sin permiso efectivo y confirma antes de ejecutar POST.
- [Seguridad] El endpoint vertical conserva la autorizacion final por rol, licencia y `empresa_id`; no hay nuevas dependencias ni cambios de esquema.

## [2026-05-11] Sincronizacion desde matriz vertical
- [Frontend] `web/administrar_empresa/verticales_integracion.html` agrega botones `Sincronizar` por vertical y muestra resultado/resumen de la accion.
- [Seguridad] La vista conserva permiso `seguridad:R`; cada POST mantiene la autorizacion real del endpoint vertical correspondiente.

## [2026-05-11] Pantalla de matriz vertical en empresa
- [Frontend] Se agrega `web/administrar_empresa/verticales_integracion.html` para consultar KPIs, estado, nucleo, especialidad y sincronizacion por vertical.
- [Menu] `web/administrar_empresa/configuracion_menu.html` incorpora `Matriz de integración` dentro de Configuracion > Base empresarial.
- [Permisos] `linkVerticalesIntegracion` queda registrado con `seguridad:R` en backend y frontend.

## [2026-05-11] Indicador de matriz vertical en panel empresa
- [Frontend] `web/administrar_empresa.html` agrega un indicador compacto en el sidebar empresarial.
- [JS] `web/js/administrar_empresa.js` lo alimenta con el resumen de `web/js/verticales_integracion_catalogo.js`.
- [UX] El panel muestra fuente API/local y conteo de verticales visibles/ocultos sin cambiar permisos, licencias ni rutas.

## [2026-05-11] Frontend consume matriz API de verticales
- [Frontend] `web/js/administrar_empresa.js` carga `/api/empresa/verticales_integracion/catalogo` antes de aplicar permisos/licencias del menu empresarial.
- [Fallback] `web/js/verticales_integracion_catalogo.js` conserva el catalogo local y ahora permite fusionar items recibidos desde backend.
- [Gobernanza] El menu deja de depender solo de un archivo JS estatico para decidir si una vertical clasica puede mostrarse como operativa.

## [2026-05-11] Catalogo API de integracion vertical
- [Backend] Se agrega `backend/handlers/empresa_verticales_integracion.go` para exponer la matriz de verticales clasicos.
- [API] Nuevas rutas de solo lectura: `/api/public/verticales_integracion/catalogo`, `/api/empresa/verticales_integracion/catalogo` y `/super/api/verticales_integracion/catalogo`.
- [QA] `backend/handlers/empresa_verticales_integracion_test.go` bloquea verticales visibles con duplicados del nucleo.

## [2026-05-11] AIU construccion integrado al nucleo
- [Backend] `aiu_construccion` enlaza clientes de obra con clientes centrales, contratos/conceptos con servicios y facturas AIU con ventas centrales en carritos.
- [Frontend] El panel AIU incluye accion de sincronizacion y resumen de clientes, servicios y facturas conectadas.
- [Gobernanza] AIU queda visible como plantilla integrada; sus tablas propias se conservan para capitulos, calculo AIU, retenciones, anticipo, garantia, avance, riesgo y auditoria tecnica.

## [2026-05-11] Drogueria/farmacia validada al nucleo
- [Backend] `drogueria_farmacia` se mantiene sobre `empresa_modulos_colombia_*` como expediente sanitario, sin tablas paralelas de productos, inventario, ventas ni pagos.
- [Frontend] La pagina de drogueria/farmacia declara que opera sobre productos, inventario, ventas y facturacion centrales.
- [Catalogo] La vertical queda visible como `plantilla_integrada_nucleo` y sin duplicados del nucleo.
- [Gobernanza] Lotes, INVIMA, formulas, controlados, dispensacion, devoluciones y farmacovigilancia quedan como especialidad sanitaria.

## [2026-05-11] Alquileres integrado al nucleo
- [Backend] `alquileres` enlaza clientes de contratos a clientes centrales, activos/tarifas a servicios y contratos con valor a ventas centrales en carritos.
- [Frontend] El panel de alquileres incluye accion de sincronizacion y resumen de clientes, servicios y contratos conectados.
- [Gobernanza] Alquileres queda visible como plantilla integrada; sus tablas propias se conservan para activos, garantias, kilometraje, GPS, mantenimiento, entrega y devolucion.

## [2026-05-11] Propiedad horizontal integrada al nucleo
- [Backend] `propiedad_horizontal` enlaza propietarios/residentes a clientes centrales, unidades/cargos a servicios y recaudos a ventas centrales en carritos.
- [Frontend] El panel de propiedad horizontal incluye accion de sincronizacion y resumen de clientes, servicios y recaudos conectados.
- [Gobernanza] Propiedad horizontal queda visible como plantilla integrada; sus tablas propias se conservan para unidades, coeficientes, cartera, PQR y asambleas.

## [2026-05-11] Apartamentos turisticos integrado al nucleo
- [Backend] `apartamentos_turisticos` enlaza huespedes a clientes centrales, unidades a servicios y reservas cerradas a ventas centrales en carritos.
- [Frontend] El panel de apartamentos incluye accion de sincronizacion y resumen de reservas, servicios, clientes y observaciones.
- [Gobernanza] Apartamentos turisticos queda visible como plantilla integrada; sus tablas propias se conservan para unidades, tarifas, disponibilidad, codigos de acceso, limpieza y mantenimiento.

## [2026-05-11] Domicilios integrado al nucleo
- [Backend] `domicilios` enlaza clientes de pedidos a clientes centrales, productos de menu a servicios y pedidos entregados a ventas centrales en carritos.
- [Frontend] El panel de domicilios incluye accion de sincronizacion y resumen de pedidos, servicios de menu, clientes y observaciones.
- [Gobernanza] Domicilios queda visible como plantilla integrada; sus tablas propias se conservan para restaurantes, domiciliarios, ofertas, GPS, tracking y estados logisticos.

## [2026-05-11] Fases de integracion profesional de verticales
- [Gobernanza] Se agrega `documentos/matriz_integracion_verticales.md` como contrato para mantener clientes, productos/servicios, ventas, pagos, facturacion, reportes y permisos en el nucleo.
- [Frontend] `web/js/verticales_integracion_catalogo.js` clasifica verticales clasicos y oculta del menu operativo los que siguen duplicando funciones centrales.
- [Catalogo] `web/js/nuevos_verticales_catalogo.js` y los endpoints de verticales nuevos publican estado de integracion, visibilidad operativa, modulos base y duplicados detectados.

## [2026-05-11] Gimnasio integrado al nucleo
- [Backend] `gimnasio` enlaza socios a clientes, planes a servicios y pagos a ventas centrales en carritos.
- [Frontend] El dashboard de gimnasio incluye accion de sincronizacion y resumen de clientes/servicios/ventas sincronizados.
- [Gobernanza] Gimnasio queda visible como plantilla integrada; sus tablas propias se conservan para acceso, clases y asistencia.

## [2026-05-11] Odontologia integrada al nucleo
- [Backend] `odontologia` enlaza pacientes a clientes, tratamientos a servicios y pagos a ventas centrales en carritos.
- [Frontend] El panel de consultorio incluye accion de sincronizacion y resumen de pacientes/tratamientos/pagos sincronizados.
- [Gobernanza] Odontologia queda visible como plantilla integrada; sus tablas propias se conservan para historia clinica, odontograma, agenda y presupuesto clinico.

## [2026-05-11] Parqueadero integrado al nucleo
- [Backend] `parqueadero` enlaza tickets cobrados a clientes opcionales, servicios y ventas centrales en carritos.
- [Frontend] El panel de parqueadero incluye accion de sincronizacion y resumen de tickets sincronizados.
- [Gobernanza] Parqueadero queda visible como plantilla integrada; su tabla propia se conserva para placas, QR, entrada/salida, tarifas y anulaciones.

## [2026-05-11] Taxi system integrado al nucleo
- [Backend] `taxi_system` enlaza clientes registrados/invitados a clientes centrales, servicios de viaje a servicios y viajes completados a ventas centrales en carritos.
- [Frontend] El panel de taxi incluye accion de sincronizacion y resumen de viajes, clientes y pendientes.
- [Gobernanza] Taxi system queda visible como plantilla integrada; sus tablas propias se conservan para conductores, GPS, despacho, ofertas y rutas.

## [2026-05-11] Panel super profesional
- [Shell] `web/super_administrador.html` queda con un menu ejecutivo de 16 accesos necesarios.
- [UX] Se agrega cabecera compacta PCS, alcance operativo y estilo visual mas denso para trabajo diario.
- [Navegacion] `web/js/super_administrador.js` solo restaura paginas visibles del panel principal.
- [Alcance] No se eliminan modulos ni endpoints; los accesos secundarios dejan de ocupar el panel inicial.

## [2026-05-11] Limpieza PostgreSQL-only
- [Backend] Las verificaciones residuales de indices en finanzas y propinas consultan `pg_indexes` y ya no conservan ramas de motor legado.
- [Frontend] Los helpers de fecha en carrito y codigos de descuento usan nombres neutrales de backend.
- [Operacion] Scripts de sincronizacion y actualizacion dejan de contemplar extensiones de motores retirados; se eliminaron artefactos locales generados en perfiles temporales.
- [Gobernanza] Documentacion e instrucciones quedan alineadas a PostgreSQL como unico motor permitido, sin dependencias nuevas ni cambios de esquema.

## [2026-05-11] Centro de mando profesional super
- [Super] `web/super/licencias_resumen.html` ahora abre con una lectura ejecutiva del VPS y del proyecto.
- [Metricas] El panel consolida CPU, memoria, disco, trafico, historico, PostgreSQL, alertas, errores, servicios, procesos, licencias, empresas y consumo OpenAI estimado.
- [Gobernanza] Se reutilizan endpoints existentes y no se agregan dependencias, tablas ni permisos nuevos.

## [2026-05-11] Cierre implementable de pendientes 1 a 8
- [Pagos] `web/pagar_licencia.html` permite elegir manualmente el pais de pago, guarda la preferencia local y recarga disponibilidad de Wompi/Epayco por `pais_codigo`.
- [Docs] Se corrigen referencias activas a documentos historicos inexistentes y se apunta a fuentes vigentes.
- [Alcance] DIAN oficial SOAP/WSDL, proveedores/hardware reales, E2E con credenciales y normalizacion masiva de mojibake quedan documentados como pendientes externos/controlados, no como cierres locales.

## [2026-05-11] Madurez empresarial de 12 pasos
- [Staging] `deploy/scripts/vps-refresh-staging-from-production.sh` anonimiza datos por defecto con `deploy/scripts/vps-anonymize-staging.sh`.
- [Monitoreo] Se agrega stack Prometheus/Grafana/node-exporter/cAdvisor en `deploy/monitoring/` y script `deploy/scripts/vps-monitoring-up.sh`.
- [Backups] Se agrega backup externo opcional por `rclone` o `s3` con cron instalable.
- [QA] Se agregan auditorias `qa_roles_matrix`, `payment_matrix_audit`, `support_center_audit`, `docs_normalization_audit` y `load_smoke_test`.
- [Release] `scripts/release_gate.ps1` genera manifiesto con `tools/release_manifest.mjs` e integra prueba de carga smoke.
- [CI] `professional-ci.yml` incorpora auditorias nuevas y despliegue opcional a staging con `PCS_ENABLE_STAGING_DEPLOY=true`.
- [Verificacion] `.\scripts\profesional_preflight.ps1 -Full` OK y `node tools\load_smoke_test.mjs` contra staging OK.

## [2026-05-11] Operacion profesional diaria de los 12 frentes
- [VPS] Se agregan scripts para instalar Docker Buildx, activar staging Nginx, tomar snapshots de observabilidad y programar backups por cron.
- [Staging] Se agrega refresco controlado de staging desde produccion con `deploy/scripts/vps-refresh-staging-from-production.sh`.
- [Release] Se agrega `scripts/release_gate.ps1` y runbook de release profesional.
- [CI/E2E] El workflow E2E queda programado contra staging y se documentan secretos requeridos de GitHub.
- [Auditorias] Se agregan auditorias de migraciones, QA funcional de modulos criticos y consistencia UX global al preflight.

## [2026-05-11] Implementacion profesional de 12 frentes operativos
- [Staging] Se agrega `deploy/docker-compose.staging.yml`, `deploy/.env.staging.example`, `scripts/staging_up.ps1` y `deploy/scripts/vps-staging-up.sh`.
- [CI/CD] Se agregan workflows `.github/workflows/professional-ci.yml` y `.github/workflows/e2e-visual.yml`.
- [Auditorias] Se agregan `tools/security_audit.mjs`, `tools/permissions_license_audit.mjs`, `tools/observability_report.mjs` y `tools/openapi_inventory.mjs`.
- [Backups] Se agrega `scripts/vps_restore_validation.ps1` para validar snapshots y ejecutar restauracion temporal PostgreSQL.
- [API/DRP] Se genera `documentos/api/openapi.generated.yaml` y runbooks de staging, CI, E2E y recuperacion ante desastre.
- [Verificacion] Se creo snapshot real en VPS `backups/vps-snapshots/20260511_055520`, se validaron tarballs y se ejecuto restauracion temporal PostgreSQL exitosa.

## [2026-05-11] Base profesional de QA, respaldo y auditoria
- [QA] Se agrega `scripts/profesional_preflight.ps1` para validar sintaxis, Docker, auditoria de modulos/permisos y `git diff --check` antes de despliegues.
- [Auditoria] Se agrega `tools/professional_audit.mjs` para revisar catalogo de 20 verticales, permisos backend, wrappers, portal publico y documentacion obligatoria.
- [RS] `rs.ps1` ejecuta preflight por defecto antes de actualizar repositorio y sincronizar VPS; se puede omitir con `-SkipPreflight`.
- [Backups] Se agrega `scripts/vps_backup_operacion.ps1` para crear dump PostgreSQL y tarballs de volumenes persistentes en la VPS con retencion.
- [Docs] Se agrega `documentos/plan_profesionalizacion_plataforma.md` como guia de los siete frentes profesionales.

## [2026-05-11] Limpieza segura en sync_to_vps
- [Deploy] `scripts/sync_to_vps.ps1` agrega `-CleanupRemoteUnusedFiles` para limpiar temporales antiguos de sync, caches locales no persistentes, contenedores detenidos, imagenes Docker dangling y cache BuildKit no usado.
- [Seguridad] La limpieza evita `docker volume prune` y no toca PostgreSQL, uploads, descargas ni backups persistentes.
- [RS] `rs.ps1` y `scripts/rs.ps1` heredan la limpieza por defecto y permiten ajustar antiguedad de temporales/cache.

## [2026-05-11] Script rapido rs
- [Deploy] Se agrega `scripts/rs.ps1` para ejecutar en secuencia `actualizar_repositorio.ps1` y `sync_to_vps.ps1`.
- [Raiz] Se agrega `rs.ps1` como acceso corto desde la raiz del proyecto.
- [Seguridad operativa] Si falla la actualizacion del repositorio, no se ejecuta la sincronizacion al VPS.

## [2026-05-11] Reubicacion de ayuda super administrador
- [Super] El boton `Ayuda super administrador` se mueve al grupo `Infraestructura y comunicaciones`, justo al lado de `Configuracion avanzada`.
- [Seguridad] Se conserva la misma ruta privada `/ayuda/ayuda.html` y el mismo filtro de rol existente.

## [2026-05-11] Portada publica con verticales completos
- [Index] Las 20 nuevas empresas del catalogo publico usan descripciones largas, similares a las tarjetas principales de la portada.
- [Probar gratis] El enlace de cada tarjeta conserva contexto de titulo, descripcion, modulo/tipo de empresa y secciones para llegar a una ficha de detalle mas completa.
- [Detalle publico] `descripcion_de_los_sistemas.ht` reutiliza el catalogo ampliado para mostrar informacion especifica de cada vertical antes del registro de prueba.

## [2026-05-10] Preconfiguraciones inteligentes y robot no automatico
- [Preconfiguracion] La siembra por tipo de empresa completa faltantes reales por `tipo_empresa_id`, aunque existan plantillas antiguas o sobrantes.
- [Verticales] Los 20 tipos nuevos usan su plantilla inteligente como default si aun no tienen preconfiguracion guardada.
- [UX] Se retira el acceso visible a `Configuracion guiada` del submenu de configuracion empresarial.
- [Robot] Al crear una empresa ya no se pregunta ni se agenda el inicio automatico del robot de configuracion.

## [2026-05-10] Carritos en modo tactil
- [Carritos] Se agrega `modo_pantalla_tactil` a la configuracion unificada de carrito por empresa/estacion.
- [UI] `carrito_de_compras.html` adapta botones, campos, toolbar, lector, cobro y resumen para uso profesional en pantallas tactiles.
- [Productos] `buscar_producto_botones.html` agranda grilla, tarjetas y buscador cuando el carrito abre el catalogo en modo tactil.
- [Persistencia] La opcion se guarda dentro de `estaciones_config`, sin tablas nuevas ni dependencias externas.

## [2026-05-10] Alertas automaticas del sistema super
- [Super] Se agrega `web/super/alertas_sistema.html` como modulo privado con configuracion, estado actual, historial, prueba de correo y evaluacion manual.
- [Backend] Se agrega `/super/api/alertas_sistema` y el worker `super.alertas_worker` para evaluar cada minuto disco VPS, trafico, sesiones administrativas y conexiones PostgreSQL.
- [Base de datos] Se crean `super_alertas_config` y `super_alertas_eventos`; `metrics` ahora persiste `disk_total`, `disk_used` y `disk_percent`.
- [Correo] Las alertas reutilizan Gmail SMTP global y por defecto notifican a `powerfulcontrolsystem@gmail.com`, con enfriamiento configurable para no repetir correos.
- [Gobernanza] Se aplican las reglas de `copilot-instructions.md`: Go puro, sin dependencias externas, documentacion de archivos, modulos, base de datos y matriz de permisos.
- [Verificacion] `go test ./...` en `backend/`.

## [2026-05-10] Roles finos y ayuda privada super
- [Permisos] Se documentan modulos finos para CRM unificado, reservas, chat/tareas, horarios, asistencia, vehiculos, hoja de vida operativa, GPS, nomina, reportes, auditoria, backups, OnlyOffice y Nextcloud.
- [Backend] Las rutas empresariales recientes quedan asociadas a wrappers especificos y la prueba de seguridad de rutas reconoce esos wrappers.
- [Licencias] Se mantiene compatibilidad para licencias antiguas con modulos amplios (`ventas`, `seguridad`, `finanzas`, `inventario`, `clientes`) para evitar perdida de acceso tras separar modulos.
- [Super] `web/super_administrador.html` agrega el boton `Ayuda super administrador`, que abre `/ayuda/ayuda.html` dentro del panel.
- [Seguridad] `/ayuda/ayuda.html` permanece exclusiva de `super_administrador`; el rol `control_super_administrador` no recibe el boton ni la ruta en su lista limitada.
- [Docs] Se agrega `documentos/reporte_roles_ayuda_super_2026-05-10.md` y se actualizan README documental, resumen, descripcion del proyecto, descripcion de modulos, matriz de roles y ayuda web.

## [2026-05-07] QA E2E Motel Calipso y ajustes flotantes
- [QA Motel Calipso] Se ejecuta regresion autenticada sobre `empresa_id=7` con 60 modulos de Administrar empresa: escritorio con 60/60 modulos cargados y validacion dirigida posterior sin errores para activos fijos, auditoria, red social, control electrico, venta publica, radio y configuracion.
- [QA profunda] Se agrega runner `backend/tmp_tools/qa_calipso_operativo/deep_flows_calipso.mjs` para crear datos QA reales en parqueadero, WMS, centros de costo, activos fijos, red social con imagen, carta publica, venta publica y validacion QR publica. Resultado final: 6/6 pasos OK, sin errores de consola, red ni pagina.
- [QA movil] Se habilita viewport configurable en `frontend_buttons_calipso.mjs` y se recorren los 60 modulos en 390x844. Los dos hallazgos moviles se corrigen y se validan de forma dirigida en venta publica y asistencia de empleados.
- [UX flotante] El robot/secretaria separa los globos del avatar en movil, se evita que el drawer del asistente quede bajo el boton de favoritos y la radio flotante se compacta como boton circular inferior izquierdo para reducir bloqueos de botones.
- [Robustez frontend] Activos fijos exporta aunque el dashboard aun no haya cargado, auditoria tolera ausencia del rotulo de empresa, red social muestra errores controlados sin romper consola, graficos/estadisticas degradan con aviso visual.
- [Riesgos externos] Queda documentado que impresora fisica, sensores electricos reales, GPS fisico, DIAN y pasarelas en produccion requieren credenciales/dispositivos externos para prueba final fuera del entorno local.

## [2026-05-06] Modulos empresariales Colombia - fases compartidas
- [ERP Colombia] Se implementan `bancos_pagos`, `gestion_documental`, `cumplimiento_kyc`, `contratos_obligaciones` y `calidad_procesos` sobre nucleo compartido por `empresa_id`.
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
## [2026-05-18] Configuracion empresarial por paginas
- [Frontend] El submenu `Configuracion empresarial` se divide en paginas enfocadas para Productos y pedidos, Identidad visual, Formato monetario, Cobro operativo, Reporte de corte, Respaldo y Pasarelas de pago.
- [Permisos] Se registran claves de pagina para las nuevas secciones bajo el modulo `seguridad` con accion de actualizacion, conservando los endpoints empresariales existentes.
- [QA] Verificacion visual desktop y movil con emulacion CDP de 390 px sin desborde horizontal; pruebas Go enfocadas de catalogo de permisos.

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
- [Carrito] El carrito de estacion incorpora boton `Control electrico` para abrir un panel operativo y controlar manualmente salidas de la estacion sin salir de la venta.
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
- Nueva funcionalidad: MÒ³dulo Red Social Comercial con portal pÒºblico y administraciÒ³n por empresa. EliminaciÒ³n de modulo juegos y venta de licencias desde cliente.

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
- 2026-05-06: implementados modulos empresariales Colombia `bancos_pagos`, `gestion_documental`, `cumplimiento_kyc`, `contratos_obligaciones` y `calidad_procesos` con nucleo compartido, APIs privadas por empresa, paginas administrativas, permisos/licencias y documentacion.
## [2026-05-11] Contrato universal de 30 verticales
- [Backend] `backend/handlers/empresa_verticales_integracion.go` deja de publicar acciones de migracion manual antigua y declara las verticales clasicas como plantillas sobre nucleo comun.
- [Licencias] La activacion de licencia aplica la preconfiguracion idempotente del tipo de empresa sin ejecutar migraciones automaticas.
- [Frontend] La matriz empresarial queda como auditoria de plantilla, nucleo, permisos, flujo de venta y reportes; los dashboards clasicos ya no muestran botones de migracion manual.
- [QA] `go test ./...`; validacion JS de catalogos y pantallas empresariales tocadas.
## [2026-05-12] Menu empresarial ajustado
- [Frontend] `web/administrar_empresa.html` elimina el cuadro de evidencia `Verticales · conteo · API/local` del encabezado del menu lateral.
- [Navegacion] `Soluciones por negocio` queda reubicado en la parte baja del menu, inmediatamente encima de `Administracion`.
- [Alcance] Sin cambios de API, permisos, base de datos ni dependencias.
