# Mapa de modulos

Tabla de ubicacion rapida para no buscar desde cero cada modulo. Si una fila
queda incompleta al implementar una mejora, actualizarla en el mismo cambio.

Actualizacion 2026-07-09: `Administrar empresa > Facturacion electronica`
compacta la bandeja `Facturas electronicas`: filtros en una sola grilla,
visualizacion imprimible con tabla de documento/totales y tipo visible
`Factura electronica` cuando una venta `CP-*` tiene factura asociada `FV-*`.
El tutorial DIAN queda como acordeon de pasos con `Software propio` como unica
modalidad PCS y enlaces oficiales DIAN por paso. En super administrador,
`Diagramas tecnicos` queda debajo de `Configuracion`.

Actualizacion 2026-07-09: `Super administrador > Panel` queda como primer
acceso independiente del menu, sin grupo `Gobierno`; `Auditoria global` se mueve
a Plataforma. El panel `web/super/licencias_resumen.html` compacta sus tarjetas,
desplaza favoritos para no chocar con el menu flotante, agrega consumo WhatsApp
API desde `/super/api/consumos` y agrega ventas mensuales de licencias desde
`/super/api/licencias/ventas_resumen`, consolidando pagos aprobados de Epayco y
Wompi relacionados con la empresa interna `Powerful Control System`.

Actualizacion 2026-07-08: `Administrar empresa > Panel` corrige la
`Configuracion guiada inicial` para que `No volver a mostrar` se guarde de forma
real en `/api/empresa/configuracion_guiada` con `action=no_mostrar_mas`. El
estado queda persistido en `empresa_estacion_prefs` con clave
`configuracion_guiada_resumen`, filtrado por `empresa_id`, y el GET del endpoint
devuelve `auto_abrir=false` y `oculta_por_usuario=true` cuando la empresa ya la
oculto o la pospuso. El frontend ya no depende de cookies/localStorage para
decidir la autoapertura y solo cierra el modal despues de un POST correcto.

Actualizacion 2026-07-07: `Super administrador > Diagramas tecnicos` agrega
15 paginas estaticas bajo `web/super/diagramas/` para modulos, ERD,
multiempresa, arquitectura, ventas POS, facturacion DIAN, inventario,
roles/permisos, API/endpoints, despliegue, seguridad, auditoria/logs, reportes,
integraciones y agentes automaticos. El visor usa `web/js/super_diagramas.js`
y la fuente estructurada `web/js/super_diagramas_data.js`; para Codex quedan
fuentes Mermaid/JSON en `documentos/diagramas/diagramas_sistema_pcs.md` y
`documentos/diagramas/diagramas_sistema_pcs_manifest.json`. No agrega APIs,
tablas, permisos empresariales ni dependencias externas.

Actualizacion 2026-07-07: la navegacion del super administrador queda
homologada para favoritos y retorno al panel. `web/js/super_administrador.js`
permite marcar como favorito todas las rutas del menu super, y
`web/js/super_page_tools.js` se carga en las paginas `web/super/**/*.html` para
mostrar `Favorito` y `Panel super` dentro de cada pagina, tanto en iframe como
en apertura directa.

Actualizacion 2026-07-07: `Administrar empresa > Canales digitales y
colaboracion > Rappi` queda en `web/administrar_empresa/rappi.html` y usa
`/api/empresa/rappi` con permiso `venta_publica:C`. El modulo guarda
configuracion por `empresa_id` en `empresa_rappi_configuracion`, registra
ordenes/webhooks en `empresa_rappi_ordenes` y recibe eventos en
`/api/public/rappi/webhook?empresa_id=...` verificando `Rappi-Signature` con
HMAC-SHA256 cuando la empresa configure `webhook_secret_ref`. La API oficial de
Rappi requiere onboarding/credenciales del aliado; PCS no hardcodea secretos.

Actualizacion 2026-06-24: `Administrar empresa > Panel` mantiene la tarjeta
`Noticias` configurable por empresa, pero el boton `Ocultar noticia` solo marca
como leida la noticia actual en el navegador para esa empresa; cuando el sistema
publique una noticia distinta, la tarjeta vuelve a mostrarse. `Red social
empresarial` usa `web/img/red-social-empresarial-portada.png` como portada,
retira el subtitulo del feed y centraliza la inyeccion del chat IA para evitar
doble icono. `Radio online` sigue apagada por defecto y limpia emisoras viejas
del navegador cuando la preferencia empresarial esta apagada.

Actualizacion 2026-06-24: `Configuracion carrito` agrega
`preguntar_tipo_documento_al_pagar` en
`web/administrar_empresa/configuracion_carrito_de_compra_empresa.html`. Cuando
esta activo, `web/administrar_empresa/carrito_de_compras.html` abre un modal al
presionar Pagar para elegir `Venta sola` o `Venta con factura electronica`. El
endpoint `/api/empresa/carritos_compra?action=pagar_estacion` recibe
`modo_documento_venta`; `backend/handlers/carritos_compras.go` fuerza la factura
electronica solo para esa venta o la omite si se eligio venta sola, manteniendo
el modo automatico/frecuencia si la opcion no se usa.

Actualizacion 2026-06-25: `Configuracion carrito` e `Identidad visual`
controlan `mostrar_logo_empresa_carrito` y `logo_empresa_carrito_tamano`
(`pequeno`, `mediano`, `grande`) desde la misma preferencia
`estaciones_config.carrito_ui_global`. El carrito ya no tiene checks internos ni
guarda esa decision en `localStorage`; solo renderiza una imagen encima de
`Totales y detalles`, en la columna activa del carrito, cuando la preferencia
esta activa y la empresa tiene logo real cargado.

Actualizacion 2026-06-26: `Configuracion carrito` retira la prueba de vista
rapida y su boton en venta directa. La venta directa conserva un unico carrito
operativo; el acceso superior queda dedicado a `Pantalla completa`, con el
mismo flujo de pagos, inventario, descuentos, credito, facturacion e impresion.

Actualizacion 2026-06-25: `Administrar empresa` ubica `Productos` dentro del
grupo `Inventario y compras`. `Configuracion empresarial > Fiscal y
automatizacion > Reporte de corte` carga una pagina propia para activar o
desactivar los campos del reporte de turno/caja por empresa, usando
`/api/empresa/corte_caja/configuracion` y la tabla
`empresa_corte_caja_configuracion`; incluye `Caja`, encabezado, resumen, pagos,
detalle de ventas, movimientos, items, sensores y auditoria.

Actualizacion 2026-06-22: `Portal publico` abre desde `index.html` las secciones
`Contacto`, `Privacidad y datos`, `Quienes somos` y `Sistema Hotel / Motel`
debajo de la barra superior, con rutas de respaldo que redirigen al mismo
portal. `Administrar empresa > Panel` deja el chat interno apagado por defecto
en `/api/empresa/panel_configuracion`; el buzon permanece activo y centrado,
recupera la tarjeta de Email corporativo junto a Favoritos y la tarjeta
Noticias enlaza a `/noticias.html`.

Actualizacion 2026-06-21: `Chat IA` queda en modo operativo/ayudante sin
autoconvertir respuestas a documentos ni mostrar exportaciones no solicitadas;
la voz se migra apagada por defecto con
`DisableFloatingChatVoicePrefs`. `Ingresos`, `Egresos` y `Compras` usan botones
con logo IA que cargan el archivo al chat para revisar calidad de captura y
novedades, ademas de precargar datos cuando el endpoint de soportes IA responde.
`Radio online` simplifica la lista a boton `Reproducir`, activa la emisora si
hace falta y minimiza al reproductor compacto. `Super administrador >
Informacion de la empresa y de los sistemas para IA` carga con textos UTF-8 y
textarea amplio dentro del panel.

Actualizacion 2026-06-20: `Super administrador` renombra `Centro de mando` a
`Panel`, reemplaza la tarjeta superior de `Email corporativo` por
`Seleccionar empresa` y envia `Analitica publica / Visitas al portal por pais`
al final del panel. `Administrar empresa > Panel` agrega la tarjeta `Noticias`
alimentada por `/api/public/noticias`, visible solo cuando hay noticias activas.
La nueva pagina `Configuracion empresarial > Panel de inicio`
(`web/administrar_empresa/configuracion/panel_inicio.html`) permite desactivar
por `empresa_id` las tarjetas `Noticias`, `Buzon de usuario` y `Chat de la
empresa` mediante `/api/empresa/panel_configuracion`; si no hay preferencia
guardada, las tres quedan activas para empresas nuevas y existentes.
`Carrito de compras` agrega la accion `Cupo de credito`, que selecciona el pago
`credito_cliente`, conserva validacion de cliente/cupo y usa sombreado de
titulos tambien en apariencias claras. `Facturacion electronica` envia correo
al cliente con cuerpo HTML/texto, adjuntos HTML/TXT y enlace QR DIAN cuando hay
codigo de validacion.

Actualizacion 2026-06-20: `Administrar empresa` mueve
`Importaciones y costeo`, `Produccion / MRP` y `Logistica WMS` al grupo
`Produccion, logistica e importaciones`; `Panel` queda sin etiqueta Beta y
`Login de usuarios` abre en ventana nueva sin cerrar la consola. El `Centro
financiero y contable` (`web/administrar_empresa/finanzas_menu.html`) ya no
duplica Nomina y usa altura minima real para que sus paginas no se vean mochas.
`Configuracion guiada inicial` refuerza el ocultamiento en celular: el boton y
el check `No volver a mostrar` guardan marcas locales compatibles, limpian el
pendiente y persisten `no_mostrar_mas` en el API; el panel se carga con version
nueva para evitar cache viejo.

Actualizacion 2026-06-20: `Carrito de compras` conserva el catalogo base de
medios de pago por permisos y agrega `metodos_pago_personalizados` en
`web/administrar_empresa/configuracion_carrito_de_compra_empresa.html`. Las
formas nuevas se muestran al cajero como opciones del selector y se registran
internamente sobre `transferencia_otro` para no romper reportes, caja ni
validaciones de backend. `Venta publica` y `Carta publica` agregan WhatsApp
flotante configurable por empresa (`whatsapp_flotante_activo`,
`whatsapp_numero`) en `empresa_venta_publica_configuracion`, visible en
`web/venta_publica.html` y `web/visualizar_productos_y_precios_publico.html`.
En super administrador queda una sola auditoria visible, `Auditoria global`, y
los codigos de descuento nuevos vencen por defecto a 30 dias.

Actualizacion 2026-06-19: `Configuracion guiada inicial` en
`web/administrar_empresa/panel.html` solo se autoabre una vez. El check y el
boton `No volver a mostrar` cierran el modal sin guardar formulario, limpian el
pendiente local y persisten `no_mostrar_mas` en
`/api/empresa/configuracion_guiada`. La reapertura manual sigue disponible en
`web/administrar_empresa/configuracion_guiada.html`.

Actualizacion 2026-06-19: `Carrito de compras` centraliza en
`web/estilos.css` el sombreado de etiquetas inspirado en `Detalle del pago`.
Labels del carrito y controles de cantidad comparten el mismo fondo adaptable a
tema claro/oscuro; `Efectivo recibido` solo aumenta tamano y conserva el color
normal del tema.
Actualizacion 2026-06-19: el mismo modulo refuerza el modo tabla plana: todo el
carrito usa fuente unificada, esquinas cuadradas, sin separaciones entre
tarjetas, botones con estilo unico, selector de acciones sin texto visible
`Accion del carrito`, encabezados de items sombreados y `Totales y detalles`
con etiquetas sombreadas en modo claro y oscuro.

Actualizacion 2026-06-19: `Captura inteligente de compras y gastos` ya no queda
clasificada como pagina IA oculta por defecto en permisos. El endpoint
`/api/empresa/soportes_compras_ia` conserva validacion de empresa, licencia,
rol y accion con `WithEmpresaSoportesComprasIAPermissions`, pero permite abrir
la pantalla cuando el rol tiene acceso al modulo operativo `soportes_compras_ia`.

Actualizacion 2026-06-19: `Super administrador` mantiene `Seleccionar empresa`
en `web/super_administrador.html` como boton integrado del toolbar, con estilo
azul claro en `web/estilos.css` y navegacion directa a `/seleccionar_empresa.html`.
`Administrar empresa` crea el submenu `Usuarios, clientes y personas` en
`web/administrar_empresa.html` para Usuarios, Clientes y Login de usuarios. El
menu Beta se controla desde `web/js/administrar_empresa.js`; Estaciones/Puntos
de venta, Corte de caja, Compras, Centro financiero y contable y Clientes quedan
como modulos sin etiqueta Beta. `Carrito de compras` mantiene la venta sin stock
habilitada por defecto en frontend y backend (`backend/db/carritos_compras.go`)
y usa un estilo unico basado en la tarjeta `Detalle del pago`, definido entre
`web/administrar_empresa/carrito_de_compras.html` y `web/estilos.css`.

Actualizacion 2026-06-19: `Correo corporativo Mailu` usa logo inline en todos
los correos del motor corporativo. `backend/handlers/usuarios_empresa.go` arma
`multipart/related` con `Content-ID` y carga el logo desde
`web/img/Logo pcs 1.png` o desde `email_corporativo.logo_url` cuando apunta a
un recurso local permitido (`img/` o `uploads/`). `mail_utils.go`,
`licencias_vencimiento_alertas.go` y `super_correos_masivos.go` reutilizan ese
constructor para evitar correos sin marca.

Actualizacion 2026-06-19: `Super administrador > Codigos descuento` queda en
`web/super/licencias_codigos_descuento.html` y
`/super/api/licencias/codigos_descuento`. Los codigos se generan
automaticamente, guardan nombre, descripcion, email y vencimiento, permiten
reenvio y el checkout de licencias ignora codigos vencidos. `Super
administrador > Empresas` agrega limpieza controlada de empresas sin licencia
activa con minimo 6 meses desde `/super/api/empresas_estado`.

Actualizacion 2026-06-19: el menu de `web/super_administrador.html` ya no
incluye `Seleccionar empresa` ni `2FA super`. `Seleccionar empresa` queda como
boton en el toolbar superior derecho y la pagina legacy
`web/super/seguridad_2fa.html` fue retirada; el control vigente de login 2FA
permanece en `web/super/configuracion_avanzada.html` y
`web/super/configuracion/login_2fa.html`.

Actualizacion 2026-06-19: `Super administrador > Empresas` queda disponible en
`web/super/empresas.html` y consume `/super/api/empresas_estado` para listar en
modo solo lectura empresas por nombre/NIT/tipo/licencia, con filtros de
licencia activa, sin licencia activa, 15 dias y vencida. `Centro de mando`
(`web/super/licencias_resumen.html`) muestra los conteos de empresas con y sin
licencia activa y enlaza con la lista filtrada.

Actualizacion 2026-06-19: `Email corporativo Mailu` agrega configuracion
`email_corporativo.logo_url` desde `web/super/email_corporativo.html`, con
fallback a `/img/Logo pcs 1.png`. Los correos HTML enviados por el motor
corporativo (`backend/handlers/usuarios_empresa.go`) insertan el logo como
cabecera visual y la accion `test_send` valida el render multipart con Mailu.

Actualizacion 2026-06-18: `Configuracion guiada inicial` ahora se abre de forma
interactiva desde `web/administrar_empresa/panel.html` cuando una empresa nueva
con preconfiguracion entra por primera vez. Backend: `/api/empresa/configuracion_guiada`
GET/POST con `empresa_id`; guarda `estaciones_config`,
`preconfiguracion_tipo_empresa_operacion`, `configuracion_guiada_resumen` y
`configuracion_guiada_interactiva`. El chat flotante agrega selector de agente y
`agente_configuracion_de_empresa` para guiar productos, tarifas, estaciones,
impresoras, caja y parametros, con acciones UI confirmadas.
Actualizacion 2026-06-19: si el usuario pulsa `Despues`, el panel llama
`POST /api/empresa/configuracion_guiada?action=posponer` y guarda
`configuracion_guiada_resumen.estado=pospuesta`; por tanto el asistente no se
abre de nuevo en el siguiente arranque o login de esa empresa. El modal tambien
incluye el check `No mostrar mas`, que guarda
`configuracion_guiada_resumen.estado=no_mostrar_mas`. Desde
`web/administrar_empresa/configuracion_guiada.html` la empresa puede abrir el
formulario guiado manualmente o iniciar el chat IA/agente de configuracion.

Actualizacion 2026-06-18: captura IA documental por empresa. Ingresos,
Egresos y Compras llaman `/api/empresa/soportes_compras_ia?action=radicar` y
`action=extraer_ia` para foto/PDF/XML; la extraccion usa GPT-5.5, descuenta
cuota avanzada en `empresa_agentes_uso_diario` y solo precarga formularios.
Productos agrega acceso "Cargar carta/precios con IA" desde
`web/administrar_empresa/administrar_productos.html`, abriendo el chat con
`agente_configuracion_de_empresa` para revision antes de guardar.

Actualizacion 2026-06-17: `Facturacion electronica Colombia` permite cargar el
PDF de Autorizacion de Numeracion DIAN Formulario 1876 desde
`web/administrar_empresa/facturacion_electronica.html`. El unico boton visible
del PDF llama `POST /api/empresa/facturacion_electronica/dian?action=importar_numeracion_pdf_ia`
con `empresa_id` y usa IA `openai:gpt-5.5` para detectar formulario, NIT/DV,
razon social, prefijo, rango y vigencia. El formulario conserva los campos
manuales para que el usuario pueda digitar o corregir datos antes de guardar. El
endpoint local `importar_numeracion_pdf` queda como respaldo tecnico/test. No se
guardan secretos ni se sobrescribe la configuracion sin confirmacion del usuario.

Actualizacion 2026-06-18: la prueba real PCS con resolucion DIAN nueva quedo
aceptada. La factura de producto `menta` emitida como `1PCS2` y la prueba
posterior `1PCS3` aparecen en portal DIAN produccion como `Aprobado con
notificacion`; `1PCS3` tambien fue aceptada por SOAP/WCF con acuse `aceptado` y
notificacion `RUT01`. PCS debe tratar `Regla: 90, Documento procesado
anteriormente` como pendiente de consulta del acuse original, no como aceptacion
automatica. El flujo
correcto es: asociar la numeracion en
`https://catalogo-vpfe.dian.gov.co/User/Login`, cargar resolucion/prefijo/rango
en PCS, consultar clave tecnica con `GetNumberingRange`, emitir y revisar
acuse/TrackId en `facturacion_electronica_pruebas_dian.html`.

Actualizacion 2026-06-18: `Agentes de mantenimiento automatico` queda disponible
en super administrador desde `web/super/agentes_de_mantenimiento_qutomatico.html`
y `/super/api/agentes_mantenimiento`. El primer agente, `dian_noticias`, se
puede habilitar/deshabilitar, guardar con hora diaria y correo de notificacion,
ejecutar manualmente y registra hallazgos relevantes de fuentes oficiales DIAN
en `super_mantenimiento_agente_hallazgos`. La clasificacion usa OpenAI cuando
esta habilitado y registra consumo en las tablas IA existentes; si OpenAI no
esta disponible, conserva lectura por palabras clave sin bloquear el panel.

Actualizacion 2026-06-17: la consola DIAN muestra errores de rechazo en rojo con
ayuda operativa para `FAB05c`, `FAD06`, `FAD05`, `FAD10`, `FAK61`, `ZE02`,
`RUT01`, vencimientos y `Regla 90`. Si un envio queda `fallido` o pendiente de
acuse original, el backend crea
una alerta de buzon para el administrador/creador de la empresa con documento,
estado, error y enlace al centro DIAN.

Actualizacion 2026-06-17: el submenu `IA` del super administrador queda reducido
a una sola pagina, `web/super/configuracion/ia_global.html`. Esa pagina consolida
proveedor/credencial, reglas, contexto, chat global y voz mediante secciones
internas; no debe volver a repetirse como cinco opciones de menu lateral.

Actualizacion 2026-06-17: el parser del Formulario 1876 soporta prefijos
alfanumericos que empiezan por numero, como `1PCS`, y vigencias que el extractor
del PDF separa visualmente como `2 4`. La autorizacion PCS cargada en produccion
quedo con resolucion `18764111318575`, prefijo `1PCS`, rango `1-100000`,
vigencia `2026-06-17` a `2028-06-17`. Una prueba real de carrito genero
`FV-FE-MENTA-20260617151719` con numero legal `1PCS1`; DIAN recibio el envio y
lo rechazo por configuracion/validacion (`FAK61`, `FAB05c`, `FAD06`), no por
prefijo ni rango.

Actualizacion 2026-06-11: `Chat IA flotante operativo` permite lectura real
controlada de la base de datos de la empresa activa solo para
`super_administrador`, `administrador_total` y `admin_empresa`. La respuesta
directa para usuarios consulta `users` filtrado por `empresa_id`; el contexto
amplio de tablas omite secretos y queda desactivado para roles operativos. La
IA no ejecuta SQL libre ni modifica datos directamente: toda escritura debe
pasar por endpoints/funciones PCS con permisos, aislamiento multiempresa y
confirmacion cuando aplique.

Actualizacion 2026-06-11: `Buzon, tareas y chat empresarial` carga el selector
de destinatarios desde `/api/empresa/buzon?action=usuarios` con nombre, rol y
email. El directorio combina usuarios activos de `users`, administrador
propietario, administradores compartidos y actor actual, siempre por
`empresa_id`. Al enviar mensajes o adjuntos a administradores se valida alcance
real de empresa antes de aceptar el destinatario.

Actualizacion 2026-06-11: `Creditos y cartera` queda activado como tarjeta de
acceso rapido dentro del `Centro financiero y contable`. La tarjeta abre el
submenu real `creditos_menu.html` conservando `empresa_id`; desde alli se
administran cupos por cliente, creditos, abonos, cartera, morosidad,
aprobaciones y estados de cuenta sin duplicar el modulo en el menu principal.

Actualizacion 2026-06-11: `Empresas compartidas` permite definir por invitacion
si el administrador receptor tambien puede compartir esa misma empresa. El dato
vive en `admin_empresa_compartida_invitaciones.puede_compartir` y pasa a
`admin_empresa_compartida.puede_compartir` al aceptar. El selector recibe
`shared_puede_compartir` y solo habilita el icono de compartir para propietario,
super administrador o acceso compartido autorizado. Si el correo invitado no
existe, el backend crea una cuenta administrativa pendiente y envia el enlace a
`registrar_nuevo_usuario_administrador.html`; despues del registro la invitacion
queda visible como pendiente en seleccionar empresa.

Actualizacion 2026-06-11: `Produccion / MRP` se unifica con `Recetas de
productos` sin mezclar responsabilidades. `recetas_productos` sigue siendo la
receta vendible del POS, inventario, impresoras y carrito; `produccion_mrp`
queda como dueño de BOM productiva, ordenes, consumos, calidad y plan de
materiales. La pagina MRP lista recetas vendibles activas y permite importarlas
como BOM mediante `/api/empresa/produccion_mrp?action=import_receta_producto`,
actualizando por codigo `POS-*` para evitar doble digitacion o duplicados. Los
endpoints genericos antiguos `/api/empresa/produccion/bom` quedan como legado
tecnico de compatibilidad, no como flujo operativo principal.

Actualizacion 2026-06-11: `Configuracion operativa de cobro` agrega permisos
por rol para ingresos y egresos manuales. Los checks viven en
`empresa_configuracion_operativa_roles.permitir_ingresos_manuales` y
`permitir_egresos_manuales`, se administran desde
`web/administrar_empresa/configuracion_impresora.html` y el backend los exige
en `/api/empresa/finanzas/movimientos` cuando el rol efectivo es `cajero`. Si
un cajero tiene alguno activo, el contexto de permisos muestra solo los enlaces
necesarios de ingresos/egresos sin abrir el resto del centro financiero.

Actualizacion 2026-06-10: `Licencias` incorpora retencion configurable de
empresas vencidas desde `web/super/configuracion_avanzada.html`. La ruta
`/super/api/licencias/vencimiento_alertas` expone `retencion_empresas`,
`retencion_preview` y `retencion_run_now`; el worker de vencimientos primero
preavisa al administrador y solo despues elimina empresas no operativas sin
licencia base vigente. El reporte vive en `licencia_empresa_retencion_log`
con `empresa_ref_id` para sobrevivir al borrado total.

Actualizacion 2026-06-10: `Docker y VPS` incorpora snapshots completos desde
`web/super/docker_portabilidad.html`. La ruta `/super/api/vps_snapshots` permite
crear/descargar copias restaurables con proyecto portable, PostgreSQL, volumenes
Docker, manifiesto y subida opcional por `rclone` a Google Drive, Mega, OneDrive
o S3 sin guardar tokens de nube en PCS.

Actualizacion 2026-06-10: `Nomina` queda como modulo principal dentro de
`Finanzas y cumplimiento`. El acceso directo de `web/administrar_empresa.html`
y el centro `web/administrar_empresa/finanzas_menu.html` abren
`web/administrar_empresa/nomina_menu.html`, que contiene submenu propio para
centro, empleados, configuracion legal, liquidaciones, pagos/PILA, nomina
electronica DIAN y tutorial. Las subpaginas reutilizan
`nomina_sueldos.html?seccion=...`; el boton de DIAN prepara el lote y luego usa
`/api/empresa/facturacion_electronica?action=nomina_electronica` con
`empresa_id`.

Actualizacion 2026-06-10: `Carrito y venta directa` permite crear cliente
rapido con datos fiscales completos para facturacion electronica: persona
natural o juridica, tipo/documento/DV, razon social, nombre comercial, regimen
IVA, responsabilidad tributaria, correo, telefono y direccion fiscal con pais,
departamento, municipio y codigo postal. La tabla de productos agregados permite
editar cantidad en linea; el frontend llama `PUT /api/empresa/carritos_compra/items`
para que el backend recalcule totales y reservas de inventario. `Devolver
producto` elimina la linea con `DELETE` del mismo endpoint y libera inventario
reservado por `empresa_id`.

Actualizacion 2026-06-10: `Carrito y venta directa` permite venta a credito de
cliente usando el modulo real `Creditos y cartera`. El medio `credito_cliente`
se habilita por empresa desde `Configuracion carrito`, exige cliente registrado
con cupo activo en `empresa_creditos_clientes_limites`, valida disponible antes
de cerrar el carrito y crea una fila en `empresa_creditos` ligada a la venta
por `venta_origen_id`/`documento_origen`. El recibo/factura visual informa el
monto de credito, codigo de cartera y vencimiento cuando aplica.

Actualizacion 2026-06-10: `Carrito y venta directa` conserva como estructura
base las tarjetas de cliente, productos agregados, detalle del pago y acciones.
Las preferencias antiguas de `carrito_ui_global` ya no pueden dejar oculta ni
reordenada esa estructura minima en venta directa o estaciones; si ya existe
carrito seleccionado y falla una sincronizacion secundaria, la pantalla se
conserva y muestra advertencia en lugar de desmontar la venta.

Actualizacion 2026-06-11: `Carrito y venta directa` refuerza el buscador del
catalogo inteligente para productos por codigo, SKU, codigo de barras o nombre.
El endpoint de productos conserva `empresa_id` y prioriza coincidencias exactas
por codigo/SKU antes de resultados parciales; la UI muestra ambos codigos en la
lista para evitar ambiguedad en caja.

Actualizacion 2026-06-11: `Carrito y venta directa` separa la busqueda por
nombre del campo de codigo/SKU y permite operar la lista con teclado. Flecha
abajo/arriba cambia el resultado seleccionado, `Enter` agrega el seleccionado y
si la busqueda devuelve una sola coincidencia se toma el primer resultado. El
boton de pantalla completa de venta directa queda solo como icono con etiqueta
accesible.

Actualizacion 2026-06-11: `Chat IA` agrega boton `Stop` en el recuadro normal
para detener audio y abortar la respuesta activa. El frontend usa
`AbortController` cuando esta disponible, invalida respuestas antiguas por
secuencia y deja el mensaje parcial marcado como detenido sin cambiar endpoints
ni permisos.

Actualizacion 2026-06-10: `Chat IA` puede usar `OPENAI_API_KEY` de entorno como
respaldo operativo cuando una credencial cifrada antigua no descifra con la llave
actual. `rs`/`sync_to_vps` propagan ese fallback al VPS sin imprimir secretos.

Actualizacion 2026-06-09: la configuracion visual separa `logo_url` como logo
corporativo general y `logo_factura_url` como logo de factura. El endpoint
`/api/empresa/configuracion_avanzada/logo` recibe `tipo_logo=empresa|factura` y
guarda archivos bajo `web/uploads/empresas/empresa_{id}_{slug}/imagenes/logos/`
por empresa. Las facturas usan el logo de factura con respaldo al corporativo;
panel, reportes y documentos generales conservan el logo corporativo.

Actualizacion 2026-06-09: el modo POS tactil se configura por empresa desde
`Configuracion carrito` en `estaciones_config.carrito_ui_global.modo_pantalla_tactil`.
Cuando esta activo, carrito, catalogo por botones, estaciones y corte de caja
amplian botones/campos sin cambiar endpoints, tablas ni permisos.

Actualizacion 2026-06-09: el recibo operativo de venta puede incluir un campo
adicional configurable por empresa desde `Configuracion de impresora`. El dato
vive en `impresion_recibo_items_json`, se imprime como informacion operativa y
no modifica venta, pagos, totales, contabilidad, inventario ni factura DIAN.

Actualizacion 2026-06-09: la bandeja `Ventas` permite generar factura
electronica desde un comprobante emitido seleccionando cliente existente o
creando cliente rapido. El backend valida `empresa_id`, asocia el cliente al
documento origen y prepara una factura con fecha/hora fiscal nueva.

Actualizacion 2026-06-11: la bandeja `Ventas` consolida ventas internas y
facturas electronicas en una sola pagina para el rol cajero y roles superiores.
Cada fila muestra la relacion `Solo venta`, `Venta con factura electronica` o
`Factura electronica`; se deja una sola accion `Ver / imprimir` y se retira el
salto redundante a la pagina separada de facturas desde el listado.

Actualizacion 2026-06-11: `Configuracion de impresora` controla por empresa los
campos de recibos y representaciones impresas, incluido `Total en letras`.
Ventas/comprobantes pueden ocultar campos visuales segun checks; facturas
electronicas y documentos electronicos conservan impresos los campos legales
obligatorios aunque el check este apagado. La regla no modifica XML, CUFE/CUDE,
envio DIAN ni totales.

Actualizacion 2026-06-09: el carrito permite transferir una cuenta abierta entre
mesas, habitaciones o estaciones cuando la empresa activa el check en
`Configuracion carrito`. El endpoint `transferir_estacion` mueve items/abonos en
transaccion y bloquea destinos ocupados o tarifas de motel/hotel incompatibles.

Actualizacion 2026-06-10: el carrito separa los medios de pago por empresa en
efectivo, tarjeta credito, tarjeta debito, transferencia Bre-B, Nequi y otra
transferencia. `Configuracion carrito` guarda checks en
`estaciones_config.carrito_ui_global` y el backend bloquea pagos/abonos con
metodos deshabilitados por empresa o rol. La tarjeta `Detalle del pago` permite
pagos combinados con esos medios y exige referencia para tarjetas y
transferencias. El medio `Credito cliente` opera contra cupos de cartera, no
suma efectivo a caja y tambien puede combinarse con otros medios. Bre-B queda
preparado para conciliacion automatica futura por webhook/API bancaria; mientras
tanto se registra la referencia confirmada.

Actualizacion 2026-06-11: `Finanzas y cumplimiento` incorpora `Pagos Bre-B QR`
en `web/administrar_empresa/finanzas_breb_qr.html` y tutorial en
`web/administrar_empresa/finanzas_breb_qr_tutorial.html`. La ruta protegida
`/api/empresa/finanzas/breb_qr` lee y guarda la configuracion Bre-B dentro de
`empresa_estacion_prefs.estaciones_config.carrito_ui_global`, lista ventas y
abonos reales pagados con `transferencia_bre_b` y permite registrar pagos
bancarios manuales en `empresa_finanzas_bancos_movimientos` para conciliacion.
Para varias cajas simultaneas se recomienda QR dinamico o referencia unica por
empresa/caja/carrito; la confirmacion automatica requiere webhook/API bancaria.

Actualizacion 2026-06-12: `Motor contable automatico` endurece
`empresa_eventos_contables` -> `empresa_asientos_contables`. El procesador de
asientos acepta eventos de ventas, facturacion, compras, finanzas, cartera,
inventario, nomina y activos fijos, genera lineas de partida doble con cuentas
base/configurables e impuestos/retenciones cuando vienen en el payload, y
bloquea el evento si no produce minimo dos lineas o si debito/credito no
cuadran. El endpoint operativo sigue siendo
`/api/empresa/finanzas/asientos_contables?action=procesar_asientos`, protegido
por permisos empresariales y siempre filtrado por `empresa_id`.

| Modulo | Pagina | Handler/API | Tablas o configuracion | Permisos | Pruebas sugeridas |
| --- | --- | --- | --- | --- | --- |
| Autenticacion administrador | `web/login.html`, `web/registrar_nuevo_usuario_administrador.html` | `backend/handlers/auth_admin_handlers.go`, `/auth/*`, `/manifest.webmanifest`, `/sw.js` | usuarios/admin en `pcs_superadministrador`, PWA publica | sesion admin, super admin | Login correo, registro, OAuth Google, instalar app, celular y PC |
| Administradores delegados | `web/super/administradores.html`, acceso desde `web/seleccionar_empresa.html?scope=principal` | `/super/api/administradores`, `backend/handlers/auth_admin_handlers.go` | `administradores.usuario_creador`, `admin_principal_delegaciones`, `email_confirm_token`, `email_confirmado`, sesiones admin | administrador principal, super admin | Selector sin invitados debe listar vacio, invitar correo nuevo, compartir a administrador existente, agregar desde panel super un correo ya confirmado sin 409, crear super por invitacion, verificar que vea empresas propias mas compartidas y que revocar no borre su cuenta |
| Auditoria super administrador | `web/super/auditoria_super_admin.html`, acceso desde `web/super_administrador.html > Acceso` | `/super/api/auditoria?scope=super_panel`, `WithSuperAuditoria`, `backend/handlers/auditoria_super.go` | `super_auditoria_eventos`, metadata saneada sin secretos | solo rol super administrador | Abrir desde menu super, filtrar movimientos, exportar CSV/JSON, editar configuracion sensible y verificar evento UI/backend |
| Gobierno IA super administrador | `web/super/configuracion/ia_global.html`, card `aiConfigCard` en `web/super/configuracion_avanzada.html`, `web/super/configuracion_logica_del_chat_con_ia.html`, `web/super/contexto_ia_logica_negocio.html`, `web/super/configuracion/voz_ia.html` | `/super/api/config/ai`, `/super/api/config/chat_ia_logica`, `/super/api/config/contexto_ia_logica_negocio`, endpoints de voz IA | `pcs_superadministrador.configuraciones` con credencial OpenAI cifrada, switches globales/proveedor, limites diarios, streaming y contexto IA; sin secretos en pantalla | solo rol super administrador | Abrir IA global, validar estado/credencial sin revelar key, probar OpenAI, guardar switch global/proveedor, abrir limites/contexto/voz, confirmar que la IA no muestra robot/secretaria y que los chats respetan servicio apagado |
| Centro de mando super | `web/super/licencias_resumen.html`, pagina inicial de `web/super_administrador.html` | Endpoints existentes de metricas, licencias, empresas, alertas, servicios, PostgreSQL y eventos UI de auditoria para favoritos; `/super/api/panel_control/reset?action=metricas|errores` para reiniciar indicadores; `/super/api/consumos` para OpenAI/WhatsApp; `/super/api/licencias/ventas_resumen` para ventas de licencias | `localStorage:super_admin:favorites` para favoritos visuales; `metrics` para historico de panel; `super_errores_sistema` para monitor de errores; `configuraciones.whatsapp.usage.*`; `pagos_epayco`, `pagos_wompi` y `licencias` para ventas aprobadas | super administrador | Abrir panel super, verificar orden Favoritos/Seleccionar empresa/OpenAI/WhatsApp/licencias, marcar paginas con la estrella sin choque con menu flotante, actualizar metricas, reiniciar metricas/errores con confirmacion, revisar barras de ventas de licencias y contraste claro/oscuro |
| Auditoria global del selector | `web/super/auditoria_global.html`, acceso desde `web/seleccionar_empresa.html` | `/super/api/auditoria`, `WithSuperAuditoria` | `super_auditoria_eventos` | administrador principal: solo su alcance; super administrador: global | Ver boton Auditoria en selector, filtrar movimientos, exportar CSV/JSON, abrir Agregar administrador, comprobar eventos de empresas/administradores/licencias/reportes |
| Empresas compartidas | `web/administrar_empresa/empresas_compartidas.html`, icono compartir en selector | `/super/api/empresas/compartidos`, `backend/handlers/empresa_compartida_handlers.go` | `admin_empresa_compartida`, `admin_empresa_compartida_invitaciones` | seguridad:U, propietario/quien comparte/receptor; super administrador puede gestionar comparticiones de cualquier empresa | Ver accesos activos, cancelar invitacion, desactivar acceso, compartir como super administrador una empresa ajena, validar que no borre cuentas y confirmar que quien compartio conserva la empresa en el selector |
| Seleccion y creacion de empresas | `web/seleccionar_empresa.html`, `web/js/seleccionar_empresa.js` | `/super/api/empresas`, `/api/user/configuracion`, `backend/handlers/system_empresas_handlers.go`, `backend/handlers/user_config.go` | empresas, tipos, preconfiguracion, licencias, defaults Colombia `CO-2026-06` en `empresa_impuestos_config`, `empresa_nomina_configuracion`, bodega base `Bodega 1` en `bodegas` y marcador `empresa_estacion_prefs.preconfiguracion_colombia_fiscal_nomina`; `admin_empresa_compartida` para alcance efectivo, `usuario_configuracion.selector_empresas_orden_json` para orden visual por usuario; carpeta base `web/uploads/empresas/empresa_{id}_{slug}/` | administrador autenticado | Crear empresa, confirmar preconfiguracion Colombia de impuestos/nomina y `Bodega 1` sin productos ni stock simulados, aplicar tipo, entrar al panel, ordenar tarjetas visibles con clic sostenido o dedo en activas/inactivas, restablecer orden, confirmar carpeta empresarial base, abrir descarga de informacion, doble clic/POST repetido sin duplicar, verificar empresas propias/delegadas/compartidas y empresas compartidas por el propietario; al eliminar, confirmar que se limpian comparticiones, licencias/datos con `empresa_id`, orden del selector y caches para que no aparezca a otros usuarios |
| Descarga informacion empresa | `web/descargar_informacion_de_la_empresa.html`, `web/js/descargar_informacion_de_la_empresa.js`, panel derecho de `web/seleccionar_empresa.html` | `/super/api/empresas?action=resumen_descarga`, `/super/api/empresas?action=exportar_informacion`, `backend/handlers/system_empresas_export.go` | snapshot exportable por `empresa_id` | administrador autenticado con empresa visible | Abrir desde selector en el panel derecho, respetar apariencia, regresar al listado, elegir formato y descargar Backup/JSON/PDF/XLS/CSV/TXT |
| Licencias | `web/elegir_licencia.html`, `web/pagar_licencia.html`, `web/super/licencias.html`, `web/super/historial_licencias.html`, `web/super/formato_para_emviar_email.html`, `web/administrar_empresa/licencia_sistema.html`, paginas super de licencias | `/licencias/activar_sin_pago`, `/api/public/licencias/payment_methods`, `/epayco/create_transaction`, `/wompi/create_checkout`, `/api/empresa/licencia_sistema/pdf`, `/super/api/licencias`, `/super/api/licencias/configuracion`, `/super/api/config/email_templates`, handlers de pagos/licencias | `licencias` con catalogo global `tipo_id=0`/`pais_codigo=GLOBAL`, limite por documentos/ventas emitidas y sin cupo comercial de cajas, configuracion `licencias.max_compras_adelantadas_misma_licencia` default 2, switches `licencias.email_bienvenida_compra.enabled`, `licencias.factura_electronica_automatica.enabled` y `licencias.email_adjuntar_factura_pdf.enabled`; la empresa emisora `Powerful Control System` se resuelve para facturar compras de licencias, pero la licencia tecnica heredada `PCS_SYSTEM_INTERNAL_PERPETUAL` se desactiva y no otorga acceso perpetuo; `licencias_activaciones_gratis`, `pagos_epayco`, `pagos_wompi` con idempotencia de activacion/correo/factura, documentos en `empresa_facturacion_documentos`, plantillas `licencia_activation_payment` y `licencia_software_pdf`; Epayco y Wompi conservan disponibilidad segun credenciales y pais configurado de la empresa; Colombia carga ePayco/Wompi por defecto y otros paises requieren habilitacion explicita; roles/modulos se gobiernan por permisos efectivos, preconfiguracion y fallback de licencias antiguas | licencia/super admin, alcance efectivo de empresas, `linkLicenciaSistema` por empresa | Ver ocho planes globales para cualquier tipo de empresa, incluyendo prueba gratis 15 dias y prueba pagada 1 dia COP 1000, activar/desactivar cada plan global desde Super administrador, confirmar que cada plan solo limita documentos/ventas y no cajas, configurar compras adelantadas y switches de comunicacion, probar checkout/pasarelas, recibir correo de bienvenida sin PDF de licencia y, si el pago real es mayor a cero y la configuracion lo permite, PDF de factura electronica adjunto en el mismo mensaje, confirmar que descuento total/valor cero no emite factura, descargar PDF de licencia desde Administrar empresa > Licencia, editar correo de bienvenida y PDF descargable desde Formatos de email, ver Historial de licencias, bloquear renovaciones fuera de regla y validar que roles respetan licencia/modulo |
| Asesor en ventas | `web/super/asesor_comercial.html`, primer acceso en `web/super_administrador.html > Comercial y licencias` | `/super/api/asesor_comercial`, `/api/asesor_comercial/aceptar`, `/api/asesor_comercial/mis_clientes`, `backend/handlers/asesor_comercial.go` | `asesores_comerciales`, `asesor_comercial_comisiones`, configuracion `licencias.asesor_promo.*` | super administrador | Abrir desde menu super, configurar promocion, invitar asesor con porcentaje de primer ano, porcentaje de renovacion anual y meses de renovacion, aceptar invitacion, validar comisiones y descuentos por codigo |
| Codigos descuento licencias | `web/super/licencias_codigos_descuento.html`, `web/pagar_licencia.html` | `/super/api/licencias/codigos_descuento`, `/api/public/licencias/checkout_summary`, `/epayco/create_transaction`, `/wompi/create_checkout`, `/licencias/activar_sin_pago`, `backend/handlers/licencias_codigos_descuento.go`, `backend/handlers/payments_handlers.go` | `configuraciones.licencias.discount_codes`, `pagos_epayco.discount_code`, `pagos_wompi.discount_code`, `licencias_activaciones_gratis.discount_code` | super administrador; checkout publico valida disponibilidad y un uso por empresa | Crear, editar, activar/desactivar y eliminar codigos globales; probar `CODIGO=10%`, `CODIGO=50000` y `CODIGO=gratis`; abrir checkout con `discount_code` por URL o escrito manualmente; validar que una empresa no pueda reutilizar el mismo codigo |
| Panel administrar empresa | `web/administrar_empresa.html`, `web/administrar_empresa/panel.html` | contexto/permisos empresa | preferencias, permisos, modulos | rol efectivo por empresa | Abrir panel por defecto, ver grupo principal `Licencia` con `Licencia del sistema`, responsive movil |
| Usuarios empresa | `web/administrar_empresa/administrar_usuarios.html`, `web/login_usuario.html` | `/api/empresa/usuarios`, `/api/empresa/roles_de_usuario`, `/api/empresa/usuarios/login`, `/api/empresa/usuarios/establecer_password`, `/auth/google/usuario/login`, callback canonico `/auth/google/callback`, `backend/handlers/usuarios_empresa.go`, `backend/handlers/auth_admin_handlers.go`, `/api/empresa/estacion_prefs` | `users`, `email_confirm_token`, `email_confirmado`, `foto_url`, fotos en `/uploads/empresas/empresa_{id}_{slug}/imagenes/usuarios/`, `empresa_estacion_prefs.estaciones_config.acceso_estaciones_cajeros`, `empresa_estacion_prefs.estaciones_config.cajas_config`, `empresa_estacion_prefs.estaciones_config.caja_login_auto_por_computador`, `localStorage.pcs_dispositivo_id`, catalogo global deduplicado de `roles_de_usuario` mas roles personalizados con `empresa_id`, `origen=empresa` y `rol_base_id` para asignacion aislada por empresa | `seguridad:C/R/U/D` por empresa; login operativo solo con invitacion/usuario confirmado; el primer ingreso exige token de invitacion, correo, documento, contrato vigente y confirmacion de contrasena; roles base: supervisor, vendedor, recepcion, cajero, portero, servicio_limpieza, jefe_bodega, responsable_bodega, recursos_humanos, tecnico_solar, contador, empresario, compras, inventario, contabilidad y auditor; alias `Caja` normaliza a `cajero`; un rol personalizado hereda permisos del rol base global y solo se puede asignar dentro de su `empresa_id`; el cajero solo ve menu operativo y `Facturas electronicas` como bandeja documental unica, conservando APIs auxiliares del carrito para productos/clientes/pagos y permisos documentales internos para consultar/reimprimir/reenviar facturas | Crear usuario, subir foto, editar, cambiar rol desde la tarjeta `Editar rol de usuario`, crear/editar/desactivar roles propios desde `Roles personalizados de esta empresa`, asignar un rol personalizado, reenviar invitacion, intentar crear un correo ya pendiente y confirmar que la lista recarga/resalta el registro con `Reenviar confirmacion`, abrir enlace de invitacion, ingresar documento y contrasena confirmada, login correo, login Google con invitacion, rechazar Google sin invitacion, validar que el login del cajero no muestra selector manual de caja, detectar computador con ID local y entrar automaticamente con caja asignada, redirigir a `administrar_empresa.html?id={empresa_id}` con `caja_codigo` cuando exista asociacion, ver roles globales sin repetidos con enlace `Saber mas`, validar que `Caja/Cajero` no se repiten y que cajero solo ve Venta directa/Estaciones/Corte de Caja/Facturas electronicas, que puede consultar/reimprimir/reenviar documentos y que carga carrito completo, asignar estaciones por cajero, crear usuario `portero`, `servicio_limpieza`, `jefe_bodega`, `responsable_bodega`, `tecnico_solar`, `contador` y `empresario`, validar que usuario_id de otra empresa no sube foto |
| Email corporativo Mailu | `web/super/email_corporativo.html`, pagina empresarial `web/administrar_empresa/email_corporativo.html`, notificaciones en `web/administrar_empresa/panel.html`, `web/administrar_empresa/configuracion/email_corporativo.html` | `/super/api/email_corporativo`, `/api/empresa/email_corporativo` con `check_unread=1`, `/pcs-mail-autologin`, `backend/handlers/email_corporativo_handlers.go`, envios comunes en `backend/handlers/usuarios_empresa.go` y `backend/handlers/mail_utils.go` | `empresa_email_corporativo`, `empresa_estacion_prefs.email_corporativo_config`, `configuraciones.email_corporativo.*`, perfil Docker `mail`, `deploy/mailu/themes`, scripts `vps-provision-mailu-mailbox.sh` y `vps-configure-mailu-host-nginx.sh`; `email_corporativo.max_accounts_per_empresa` default 5; buzones de sistema `ventas@powerfulcontrolsystem.com` para licencias/comercial y `soporte@powerfulcontrolsystem.com` para alertas, usuarios, facturacion, reportes, masivos, agente DIAN y pruebas; conteo de no leidos se consulta por IMAP `STATUS INBOX` usando clave cifrada sin exponerla | super administrador para configurar global; seguridad empresa para consultar/guardar configuracion por empresa; `linkEmailCorporativo` usa seguridad:R en frontend | Activar/desactivar, guardar dominio/webmail/cupo de cuentas por empresa, probar Mailu, provisionar ventas/soporte, probar envio real, crear empresa y verificar correo unico, sincronizar existentes, abrir `E-mail Corporativo` desde el menu con autologin en tema claro/oscuro, revisar notificacion bajo Favoritos, validar conteo de no leidos o error saneado, desactivar autoapertura y cambiar clave sin exponerla |
| Documentos OnlyOffice | `web/administrar_empresa/documentos_onlyoffice.html`, configuracion super `web/super/configuracion/onlyoffice.html` | `/api/empresa/documentos`, `/api/onlyoffice/file`, `/api/onlyoffice/callback`, `/super/api/config/onlyoffice`, `backend/handlers/onlyoffice.go`, `backend/handlers/onlyoffice_super_config.go` | archivos temporales por `empresa_id` en `PCS_DATA_ROOT`/`/data/empresas`, `configuraciones.onlyoffice.*`; `editor_config` usa JWT HS256 solo en `config.token` y tokens temporales separados para archivo/callback | `documentos_onlyoffice:R`, wrapper `WithEmpresaDocumentosOnlyOfficePermissions`; archivo/callback publicos por token temporal | Crear Word/Excel/PowerPoint, verificar que `api.js` carga desde URL publica, abrir editor sin error de token, guardar en dispositivo, callback devuelve `{"error":0}`, otro `empresa_id` no accede al archivo |
| Configuracion empresa | `web/administrar_empresa/configuracion_menu.html`, `web/administrar_empresa/configuracion/*.html`, `web/administrar_empresa/configuracion_de_estaciones.html`, `web/administrar_empresa/configuracion_impresora.html` | `/api/empresa/configuracion_*`, `/api/empresa/configuracion_avanzada`, `/api/empresa/estacion_prefs`, `/api/empresa/inventario/configuracion` | configuracion avanzada, operativa, estacion prefs, `estaciones_config.cajas_config`, `menu_visual_config.hidden_links` para ocultamiento visual del menu por empresa, perfil tributario Colombia en `empresa_configuracion_avanzada.tipo_persona_fiscal`, `naturaleza_juridica`, `regimen_tributario_colombia`, `iva_responsabilidad`, `inc_responsabilidad`, `responsabilidades_rut_json`, `obligaciones_fiscales_json`, `empresa_configuracion_avanzada.mostrar_deducido_impuesto_factura`, `empresa_configuracion_avanzada.impresion_recibo_items_json` con campos imprimibles y `total_en_letras`, `empresa_configuracion_avanzada.impresion_corte_items_json`, tamanos `impresion_factura_fuente_pos/carta` e `impresion_reporte_fuente_pos/carta`, `empresa_inventario_configuracion.producto_campos_obligatorios_json` | configuracion/seguridad | Guardar seccion independiente, configurar Menu visible sin cambiar permisos backend, configurar perfil tributario Colombia por persona natural/juridica, regimen ordinario/SIMPLE/especial/ingresos y patrimonio/no declarante PN, IVA/INC y responsabilidades RUT sugeridas, configurar varias cajas fisicas con nombre/descripcion, activar deducido de impuesto en impresion de factura, configurar checks de campos imprimibles para recibos/facturas/cortes, activar total en letras, ajustar tamano de fuente POS/carta de facturas y reportes, verificar que factura electronica conserva campos obligatorios, guardar campos obligatorios de productos desde Productos y pedidos y recargar |
| Configuracion rol cajero | `web/administrar_empresa/configuracion_rol_cajero.html`, enlazada desde `web/administrar_empresa/configuracion_menu.html` con `linkConfiguracionRolCajero` | `/api/empresa/configuracion_operativa?action=rol`, `/api/empresa/estacion_prefs`, `/api/empresa/roles_de_usuario`, `/api/empresa/usuarios` | `empresa_configuracion_operativa_roles` para regla `cajero`, `empresa_estacion_prefs.estaciones_config.carrito_ui_global`, `empresa_estacion_prefs.estaciones_config.acceso_estaciones_cajeros`, `roles_de_usuario` personalizado con `empresa_id` y `rol_base_id` | `seguridad:U` en menu; cambios siempre por empresa activa | Cargar rol base cajero, crear/actualizar perfil personalizado de cajero, activar/desactivar cobro efectivo/tarjetas/transferencias/pago mixto/propinas/comisiones/ingresos/egresos, activar checks principales de carrito POS, medios de pago, botones visibles y control por estaciones; comprobar que no altera el rol global ni datos de otra empresa |
| Configuracion carrito | `web/administrar_empresa/configuracion_carrito_de_compra_empresa.html` | `/api/empresa/estacion_prefs` | `estaciones_config.carrito_ui_global`, incluido `modo_pantalla_tactil`, `habilitar_bascula_electronica`, medios de pago, `metodo_pago_credito_cliente`, `mostrar_boton_transferir_cuenta_carrito`, `permitir_transferir_cuenta_carrito` y `atajos_pos` | ventas/configuracion | Checks visibles, pagar, cliente obligatorio, medios de pago habilitados/deshabilitados, pagos combinados, credito cliente con cupo, bascula electronica apagada por defecto, QR, offline, tarjeta Domotica, transferencia de cuenta, feedback sonoro/tactil PC/celular, modo POS tactil y atajos F1-F12/combos configurables por empresa |
| Estaciones | `web/administrar_empresa/estaciones.html` | `/api/empresa/carritos_compra`, `/api/empresa/carritos_compra/items`, `/api/empresa/estacion_prefs`, `/api/empresa/estacion_aseo` | estaciones en `empresa_estacion_prefs`, `acceso_estaciones_cajeros`, carritos, `empresa_estacion_aseo_eventos`, lectura de `carrito_ui_global.modo_pantalla_tactil` | ventas/estaciones; `portero` con `ventas:R/A` restringido a `linkEstaciones` y `action=activar_estacion`; `servicio_limpieza` con `ventas:R` y finalizacion de aseo | Activar estacion, doble clic, estados entre usuarios, modo POS tactil con tarjetas grandes, cajero con estaciones filtradas no puede operar estaciones no asignadas por URL/API, portero ve solo estaciones y no abre carrito ni items, Servicio de limpieza solo limpia estaciones sucias |
| Carrito y venta directa | `web/administrar_empresa/carrito_de_compras.html`; estilos finales en `web/estilos.css` bajo `body.carrito-flat-page` y `body.pos-touch-mode`; iframe de `web/administrar_empresa.html` permite fullscreen | `/api/empresa/carritos_compra`, `/api/empresa/carritos_compra/items`, `action=transferir_estacion`, auxiliares `/api/empresa/productos`, `/api/empresa/servicios`, `/api/empresa/recetas_productos`, `/api/empresa/clientes`, `/api/empresa/creditos?action=disponibilidad_cliente|limites_cliente`, `/api/empresa/codigos_de_descuento`, `/api/empresa/propinas`, `/api/empresa/comisiones` | `carritos_compras`, `carrito_compra_items`, abonos, `inventario_existencias`, `inventario_movimientos`, `clientes` con datos fiscales DIAN por empresa, `empresa_creditos`, `empresa_creditos_clientes_limites`, `estaciones_config.carrito_ui_global`, `empresa_configuracion_avanzada.mostrar_deducido_impuesto_factura`, `empresa_configuracion_avanzada.impresion_recibo_items_json`; tarjeta independiente de bascula usa Web Serial del navegador solo si `habilitar_bascula_electronica` esta activo, sin tabla nueva y con cantidad decimal solo para unidades de peso; atajos POS se leen desde `carrito_ui_global.atajos_pos` | ventas/carritos; cajero con APIs auxiliares sin paginas administrativas; transferencia bloqueada si el check empresarial esta apagado | Agregar item desde dos cajeros, crear/asignar cliente natural o empresa con NIT/DV, regimen IVA, responsabilidad tributaria, direccion fiscal y cupo, buscar productos por codigo/SKU desde el campo de escaner y por nombre desde el campo dedicado a la derecha, agregandolos ahi mismo, buscar productos por nombre/SKU/codigo de barras o atajo, editar cantidad en linea validando enteros o peso y verificando cantidad visible en la tabla, devolver producto y confirmar liberacion de inventario, vender a credito total o mixto validando cupo y cartera creada, validar pago combinado con valores legibles por medio, transferir cuenta abierta a mesa/habitacion disponible, validar bloqueo por destino ocupado y tarifa motel/hotel incompatible, activar check de bascula, conectar bascula USB/serial real desde Chrome/Edge local/HTTPS, aplicar peso al escaneo, validar atajos F1-F12/CTRL/ESC/ENTER, descuentos/taxi/pagos en modo POS tactil, validar Colombia/COP, reserva de stock, cliente, abono, pago idempotente, impresion, fullscreen y carga con rol cajero |
| Domotica | `Administrar empresa > Analisis y control > Domotica` en `web/administrar_empresa.html`, submenu `web/administrar_empresa/modulo_menu.html?module=control_electrico`, vistas `web/administrar_empresa/control_electrico.html?pagina=*` para resumen, conexion, controladores, estaciones/aparatos, automatizaciones, reportes y bitacora; `web/administrar_empresa/configuracion_sensores_raspberry.html`, tarjeta `Domotica` en `web/administrar_empresa/carrito_de_compras.html`; storage en `web/super/domotica_storage.html` | `/api/empresa/control_electrico`, `/super/api/domotica_storage`, `backend/handlers/control_electrico.go`, `backend/handlers/super_domotica_storage.go` | `empresa_control_electrico_config`, `empresa_control_electrico_raspberry_pis`, `empresa_control_electrico_reles`, `empresa_control_electrico_eventos`, `empresa_control_electrico_lecturas`, `empresa_control_electrico_reglas`, configs `domotica.storage.*` | clave tecnica `control_electrico`, permisos `linkControlElectrico` y `linkConfiguracionSensoresRaspberry`, configuracion/ventas segun pagina; storage solo super admin | Abrir Domotica desde Analisis y control, confirmar que cada boton del submenu carga solo su vista, activar modulo, crear lampara/ motobomba/aire, subir foto, programar horarios, registrar regla de sensor, revisar reporte, abrir Storage Domotica y cambiar limite |
| Caja y corte de turno | `web/administrar_empresa/corte_de_caja.html`, tarjeta Caja en `web/administrar_empresa/estaciones.html` | `/api/empresa/corte_caja`, `/api/empresa/corte_caja/configuracion`, `/api/empresa/carritos_compra?action=totales_pago`, lectura de `/api/empresa/estacion_prefs` para modo tactil | `empresa_cierres_caja`, movimientos, corte config, `usuario_creador` por caja, catalogo `estaciones_config.cajas_config`, deteccion `estaciones_config.caja_login_auto_por_computador` activa por defecto, `empresa_configuracion_avanzada.impresion_corte_items_json` para campos impresos, `carrito_ui_global.modo_pantalla_tactil` para UX | finanzas/corte caja | Corte automatico, reporte mi turno, cerrar e imprimir, modo POS tactil con acciones grandes, alternar checks de campos de corte, dos cajeros con caja y totales independientes, abrir `CAJA-1 - FRUTERA` desde estaciones, login de cajero sin selector manual y caja detectada por computador |
| Suite contador | `web/administrar_empresa/suite_contador.html`, acceso desde `web/administrar_empresa/finanzas_menu.html` | Sin endpoint nuevo; consulta `/api/empresa/permisos_contexto` y enlaza modulos existentes | Sin tablas nuevas; coordina `portal_contador`, `contabilidad_colombia`, `contabilidad_colombia_avanzada`, `impuestos`, `facturacion_electronica`, `declaraciones_tributarias`, `portal_terceros_certificados`, `cierre_fiscal`, `activos_fijos_niif_fiscal`, `nomina_sueldos`, `tesoreria_presupuesto`, `reportes`; los accesos IA se omiten si no estan habilitados por empresa | `finanzas:R`, pagina `linkSuiteContador`; rol `contador` ve la suite y accesos contables clave, pero cada endpoint mantiene su wrapper y licencia; IA no se muestra por defecto | Abrir desde Centro financiero, validar que conserva `empresa_id`, muestra modulos disponibles/bloqueados segun `/api/empresa/permisos_contexto`, filtra busqueda, abre modulos con `empresa_id` y no crea datos; confirmar que IA no aparece sin regla fina explicita |
| NIIF | `web/administrar_empresa/niif.html`, accesos desde `finanzas_menu.html` y `suite_contador.html` | Sin endpoint nuevo propio; lee `/api/empresa/contabilidad_colombia?action=dashboard` si esta disponible | Sin tablas nuevas; guarda diagnostico local por navegador/empresa y enlaza contabilidad, activos NIIF/fiscal y cierre | `finanzas:R`, pagina `linkNIIF`; rol `contador` puede consultar sin escritura | Abrir desde Centro financiero, validar metricas reales de contabilidad cuando hay permiso, marcar diagnostico, cambiar politicas, calcular deterioro/depreciacion/conciliacion, exportar JSON/TXT y confirmar que conserva `empresa_id` |
| Renta IA | `web/administrar_empresa/renta_ia.html`, acceso desde `web/administrar_empresa/finanzas_menu.html` solo si la empresa lo habilita | `/api/empresa/finanzas/renta_ia`, `backend/handlers/finanzas_renta_ia.go`, `backend/db/finanzas_renta.go` | Sin tablas nuevas; lee `carritos_compras`, `empresa_finanzas_movimientos`, `inventario_movimientos`, `empresa_nomina_liquidaciones` y uso IA diario por empresa | `finanzas:R`, pagina `linkRentaIA`, wrapper `WithEmpresaFinanzasPermissions`; oculto por defecto salvo regla fina explicita de pagina | Abrir desde Centro financiero solo cuando este habilitado, calcular periodo actual, ajustar tarifa/retenedores/deducciones, validar alerta de doble conteo, generar analisis IA si esta activa, confirmar que otro `empresa_id` no omite wrapper |
| Centro IA empresarial | `web/administrar_empresa/centro_ia_empresarial.html`, acceso desde `web/administrar_empresa.html` y `web/administrar_empresa/finanzas_menu.html` segun permisos/licencia | `/api/empresa/ia_empresarial`, `backend/handlers/empresa_ia_empresarial.go`, `web/js/ai_button_icons.js` | Sin tablas nuevas; lee snapshot real de `empresas`, `carritos_compras`, `empresa_finanzas_movimientos`, `clientes`, `productos`, `servicios`, `inventario_existencias` y registra uso IA diario por empresa | `reportes:R`, pagina `linkCentroIAEmpresarial`, wrapper `WithEmpresaReportesPermissions`; no concede escritura, aprobacion ni emision por si solo | Abrir Centro IA empresarial con rol permitido, validar KPIs reales por periodo, ejecutar diagnostico, borrador factura, cobranza, inventario, conciliacion, compras y cumplimiento DIAN; confirmar que los botones IA tienen icono y que otro `empresa_id` no omite wrapper |
| Chat IA empresarial con agentes | Chat flotante/normal, `web/administrar_empresa/centro_ia_empresarial.html`, `web/administrar_empresa/estacion_ia_pedidos.html`, tarjetas IA de `web/administrar_empresa/estaciones.html` | `/api/empresa/chat_con_inteligencia_artificial/modelos`, `/consultar`, `/consultar_con_adjunto`, `/consultar_stream`, `/api/empresa/ia_pedidos_estacion/ejecutar`, `backend/handlers/chat_con_inteligencia_artificial_controller.go`, `backend/handlers/ia_pedidos_estacion.go` | `empresa_ai_uso_diario`, `empresa_ai_consultas` y `empresa_agentes_uso_diario`; auditoria de IA guarda `agente=...`; no guarda secretos ni SQL libre | wrappers de ventas para chat/pedidos, reportes para Centro IA; cada endpoint revalida sesion, `empresa_id`, rol/licencia y permisos efectivos | Cargar catalogo de modelos/agentes, elegir agente general/ventas/inventario/compras/nomina/impuestos/agente_internet, probar consulta normal y streaming, verificar limite diario de agente, confirmar que pedidos con IA solo agregan productos por endpoint autorizado y que Centro IA solo entrega borradores/recomendaciones |
| Ayuda con caja IA | `web/js/help_ai_bridge.js`, `web/js/ai_chat_drawer.js`, ayudas `web/ayuda/*.html`, tutoriales empresariales | Usa `PCSAIChatHelp`, compatibilidad `PCSAIChatRobot.showMessage` redirigida al recuadro normal, mensaje frontend `pcs-help-ai-open`, chat `/api/empresa/chat_con_inteligencia_artificial/*`, pedidos `/api/empresa/ia_pedidos_estacion/ejecutar` y radio `/api/empresa/ia_radio/activar` | Preferencias `chat_flotante.*` en `empresa_estacion_prefs`; `chat_flotante.chat_enabled` queda activo para empresas y se reaplica en arranque; `ia_radio` solo cambia emisora por empresa; los enlaces de ayuda conservan su URL HTML como respaldo y el puente conserva `empresa_id` en URLs internas; el boton `Stop` cancela audio y respuesta activa con `AbortController`/secuencia local sin modificar datos | El asistente se muestra como recuadro normal con icono circular flotante visible en empresas; robot/secretaria quedan retirados; la voz se conserva; cajero puede pedir a estacion/mesa/habitacion y activar radio; productos, nomina y tarifas dependen de sus wrappers y permisos | Abrir Nomina, presionar `Ayuda`, confirmar que dentro del panel empresarial aparece el recuadro IA con ayuda contextual, que el icono circular no desaparece por cache viejo, que `Stop` corta audio/respuesta en curso y que una accion `PCS_ACTION` requiere confirmacion antes de ejecutar |
| Reportes de turnos | `web/administrar_empresa/reportes_turnos.html` | `/api/empresa/corte_caja?action=turnos|turno_reporte|turno_export|turno_email`, lectura de `/api/empresa/configuracion_avanzada` | `empresa_cierres_caja`, `empresa_configuracion_avanzada.impresion_corte_items_json` | finanzas/reportes | Vista POS 80mm, carta, exportar, email, verificar que encabezado y detalle respetan checks de corte por empresa |
| Facturacion electronica Colombia | `web/administrar_empresa/facturacion_electronica.html`, `facturacion_electronica_pruebas_dian.html` como `Centro de habilitación DIAN`, `facturacion_electronica_tutorial_dian.html` como guia operativa; menu multi-pais con selector `Pais fiscal` que carga Colombia/DIAN en la pagina principal y abre las paginas existentes de Ecuador/SRI o Panama/DGI segun la seleccion, sin enlaces repetidos por pais | `/api/empresa/facturacion_electronica?action=config_pais`, `/api/empresa/facturacion_electronica`, `/api/empresa/facturacion_electronica/dian?action=pruebas_dian`, `historial_tracks`, `consultar_acuse_real`, `validar_credenciales`, `diagnostico_oficial`, `estado_conexion_dian`, `vencimiento_certificado`, `analizar_captura_dian`, `uploadDIANCompanySignature` | `facturacion_electronica_pais`, `facturacion_electronica_reintentos`; `empresa_dian_configuracion` guarda software DIAN, prefijo, resolucion, rango, `llave_tecnica`, firma, vencimiento, alerta, modo de operacion, fechas del set, documentos requeridos y minimos aceptados; `empresa_dian_track_historial` guarda TrackId/ZipKey por empresa con ultimo acuse GetStatusZip saneado sin XML crudo ni secretos; `facturacion_electronica.html` muestra selector superior de pais detectado/seleccionado y una barra 0-100% de configuracion DIAN que calcula identidad fiscal, ambiente/TestSetId, software, numeracion, llave tecnica y firma/certificado con lista de faltantes; el resumen de cobertura esencial muestra factura, notas, documento soporte, nomina y eventos RADIAN como operativos cuando el catalogo backend los marca listos; la tarjeta `Asistente DIAN con capturas` queda retirada de la pagina principal y la configuracion DIAN se realiza desde formularios, ayuda contextual y tutorial; los campos de firma, DIAN Colombia, configuracion por pais y avanzada tienen ayuda contextual `?` mediante `web/js/form_field_help.js` sin exponer secretos; el guardado de perfil pais usa `action=config_pais` para no mezclarse con acciones fiscales por `documento_codigo`; los resultados de diagnostico se muestran como resumen legible con JSON tecnico plegado y saneado; `generateDIANUBLBase` genera factura, nota credito y nota debito con estructura DIAN UBL 2.1, `ProfileID` por tipo, `CUFE/CUDE-SHA384`, `DianExtensions`, `SoftwareSecurityCode`, QR, parties, `PaymentMeans`, totales y lineas correctas; `validateDIANDocumentPreflight` bloquea XML incompleto o con tipo de linea equivocado antes de enviar; `ErrorMessageList` de DIAN se conserva como lista saneada; las notas requieren factura aceptada por DIAN como referencia y se bloquean si solo existe factura generada/pendiente; los envios manuales de prueba usan confirmacion integrada antes de llamar DIAN; el preset principal del portal para software propio/proveedor queda en 30 facturas, 10 notas debito y 10 notas credito, pero el boton automatico usa el objetivo guardado por empresa; el boton rapido 2+2+2 fue retirado para evitar pruebas parciales accidentales; `test_set_id` es obligatorio para habilitacion real; `token_emisor_ref` solo se exige en proveedor/API bearer, no en endpoint oficial SOAP/WCF DIAN; el transporte oficial usa WS-Security con `BinarySecurityToken`, firma de `wsa:To` e `InclusiveNamespaces`; firma electronica privada en `web/uploads/empresas/empresa_{id}_{slug}/facturacion_electronica/firma_electronica/` y capturas historicas en `facturacion_electronica/capturas_dian/`, ambas con referencias internas y carpeta empresarial | facturacion_electronica | Abrir `Tutorial DIAN`, cambiar pais desde selector y verificar que Colombia conserva DIAN mientras Ecuador/Panama cargan su pagina especifica con `empresa_id`, configurar empresa, revisar ayudas `?` por campo sin secretos, cargar firma, guardar configuracion pais, revisar barra 0-100% de configuracion DIAN y faltantes, guardar objetivo del portal DIAN, revisar barra 0-100% de avance operativo, validar credenciales, enviar primero una factura diagnostica real con TestSetId/firma, consultar `GetStatusZip` hasta `StatusCode=00` o rechazo, solo despues enviar nota debito y nota credito referenciando esa factura aceptada, ver historial TrackId/ZipKey, registrar evento RADIAN sobre factura electronica existente cuando aplique, reconciliar aceptados minimos por tipo, verificar modal integrado, mensajes de recepcion, correo de alerta al administrador y aislamiento por `empresa_id` |
| Facturacion electronica Panama | `web/administrar_empresa/facturacion_electronica_panama.html` | `/api/empresa/facturacion_electronica/panama` | `facturacion_electronica_pais` | facturacion_panama | Detectar pais PA, guardar DGI, licencia |
| Facturacion electronica Ecuador | `web/administrar_empresa/facturacion_electronica_ecuador.html` | `/api/empresa/facturacion_electronica/ecuador` | `facturacion_electronica_pais` | facturacion_ecuador | Detectar pais EC, checklist SRI, guardar |
| Facturacion offline | carrito y service worker/conectividad | `/api/empresa/offline_ventas` | `empresa_ventas_offline_sync`, cola local navegador por empresa+cajero | ventas/offline | Perder internet con caja abierta, imprimir provisional, validar que dos cajeros no compartan cola, sincronizar una sola vez por `sync_key` |
| Inventario y productos | paginas de productos, bodegas, categorias, recetas | APIs `/api/empresa/productos*`, `/api/empresa/inventario/configuracion`, `/api/empresa/impuestos?action=context`, `/api/empresa/recetas_productos`; `/api/empresa/productos?action=exportar|plantilla_importacion|importar` | productos, bodegas con default `Bodega 1` por empresa, kardex, recetas, empresa_inventario_configuracion, `empresa_impuestos_config` para selector tributario; `administrar_productos.html` usa ayuda contextual `?` en los 25 campos del formulario `Nuevo producto` y tarjeta superior para exportar CSV/JSON/HTML carta/POS o importar CSV con plantilla | inventario; roles `jefe_bodega` y `responsable_bodega` administran bodega sin eliminar | Crear empresa y confirmar `Bodega 1`; crear producto, verificar ayudas `?`, exportar CSV/JSON/HTML carta y POS, descargar plantilla, importar CSV con duplicados/errores, que los campos obligatorios configurados se apliquen al formulario/importador, elegir impuesto configurado activo de ventas, bodega, receta, traslado, validar menu y bloqueo de ventas/caja para jefe_bodega/responsable_bodega |
| Produccion / MRP | `web/administrar_empresa/produccion_mrp.html`, tutorial `web/administrar_empresa/produccion_mrp_tutorial.html`, acceso desde `modulo_menu.html?module=produccion_mrp` | `/api/empresa/produccion_mrp?action=dashboard|recetas|ordenes|consumos|calidad|mrp_plan|generar_mrp|seed_demo|catalogo_recetas_vendibles|import_receta_producto`, `backend/handlers/produccion_mrp.go`, `backend/db/produccion_mrp.go` | `empresa_produccion_mrp_config`, `empresa_produccion_recetas`, `empresa_produccion_receta_componentes`, `empresa_produccion_ordenes`, `empresa_produccion_consumos`, `empresa_produccion_calidad`, `empresa_produccion_mrp_plan`; importa recetas POS desde `recetas_productos` como BOM `POS-*` sin duplicar el rol de Productos | `produccion_mrp:C`, wrapper `WithEmpresaProduccionMRPPermissions`, siempre filtrado por `empresa_id` | Cargar demo PCS idempotente, importar una receta vendible como BOM, crear orden, iniciar, registrar consumo, enviar a calidad, registrar revision, generar MRP y verificar resumen/alertas; confirmar que Productos conserva catalogo/POS y MRP conserva planeacion/produccion |
| Compras y proveedores | paginas de compras/proveedores | APIs de compras/proveedores | compras, proveedores, movimientos inventario | compras/inventario | Compra con proveedor creado y stock |
| Creditos y cartera | `web/administrar_empresa/finanzas_menu.html` tarjeta `linkCreditosTarjeta`, `web/administrar_empresa/creditos_menu.html`, `creditos.html` | `/api/empresa/creditos` | creditos, cuotas, abonos, clientes, cupos por cliente | finanzas | Abrir desde Centro financiero, nuevo credito, cupo cliente, abonos, morosidad, estado de cuenta |
| Reportes ejecutivos y auditoria de modulos | `web/administrar_empresa/reportes_ejecutivos.html`, filtro ampliado en `web/administrar_empresa/auditoria.html` | `/api/empresa/reportes`, dataset `operativo_modulos_resumen`; `/api/empresa/auditoria/eventos` para eventos | datos operativos por modulo; inventario de tablas empresariales, hubs sin tabla y actividad por `empresa_id`; eventos especificos para Bre-B QR, buzon/tareas/chat, impresoras, menu visible y atajos POS | reportes; rol `empresario` con `reportes:R`; auditoria segun wrapper existente y helper especifico por modulo | Select de reporte, abrir `Auditoria de modulos`, filtrar area `Operacion y auditoria`, exportar, validar DIAN/Bolsa/Suite contador/NIIF/Centro IA/OnlyOffice/Bre-B/Buzon/Impresoras/Menu visible/Atajos POS en filtros y que `empresario` no vea reportes de turnos/caja |
| Reportes globales del selector | `web/seleccionar_empresa.html`, `web/super/reportes_globales.html`, `web/js/super_reportes_globales.js` | `/super/api/reportes_globales`, `backend/handlers/reportes_globales.go` | catalogo con el mismo alcance efectivo del selector: empresas propias, delegadas o compartidas | administrador autenticado con empresas visibles | Abrir desde seleccionar empresa, elegir fechas, seleccionar varias empresas visibles, rechazar empresa_id fuera de alcance, ver graficos, exportar PDF/XLS/CSV/TXT/JSON e imprimir |
| Energia solar | `web/administrar_empresa/energia_solar.html`, `web/js/energia_solar.js` | `/api/empresa/energia_solar`, `backend/handlers/energia_solar.go`, `backend/db/energia_solar.go` | `empresa_energia_solar_sistemas`, `empresa_energia_solar_alertas`, `empresa_energia_solar_lecturas`, `empresa_energia_solar_eventos`, `tipo_empresa_preconfiguraciones.config_json.modulos.energia_solar` | `energia_solar:C/R/U`, fallback licencia a `control_electrico` o `seguridad`; rol `tecnico_solar` solo `R`; preconfiguracion apagada por defecto | Registrar sistema Victron/SMA/SolarEdge/gateway, configurar bateria, crear alerta, registrar lectura, validar correo/evento, tecnico_solar solo consulta dashboard/eventos, crear tipo de empresa y confirmar catalogo solar opcional |
| Camaras y DVR | `web/administrar_empresa/camaras.html`, `web/js/camaras.js`, visor en `web/administrar_empresa/estaciones.html`, configuracion en `web/administrar_empresa/configuracion_de_estaciones.html` | `/api/empresa/camaras`, `backend/handlers/camaras.go`, `backend/db/camaras.go` | `empresa_camaras`, `estaciones_config.camaras_enabled`, `estaciones_config.camaras_placement`, `estaciones_config.estaciones[].tipo_estacion`, `estaciones_config.estaciones[].camara_id` | `camaras:C/R/U`, pagina `linkCamaras`, fallback licencia a `control_electrico` o `seguridad` | Registrar camara RTSP/ONVIF/HLS/WebRTC/MJPEG/iframe, validar que el visor web solo use URL segura, asociar a estacion, marcar estacion tipo camara, cambiar orden antes/despues, confirmar que otro `empresa_id` no puede editarla |
| GRAFOLOGIX grafologia IA | `web/administrar_empresa/grafologia.html`, `web/js/grafologia.js` | `/api/empresa/grafologia`, acciones `analizar` y `analizar_ia`, `/api/empresa/clientes`, `backend/handlers/grafologia.go`, `backend/internal/grafologia`, `backend/db/grafologia.go`, Chat IA `openai:gpt-5.5` | `empresa_grafologia_analisis` con `cliente_id`, snapshot de cliente y descripcion/caracteristicas de persona; `metricas_json` guarda `details` con medidas tecnicas de inclinacion, espacios, margenes, continuidad, lineas, palabras y forma; imagenes en `/uploads/empresas/empresa_{id}/imagenes/grafologia/`, artefactos en `/procesado/`; la transcripcion opcional se conserva para compatibilidad de datos; consultas GPT-5.5 registradas en uso diario IA por empresa | `grafologia:C/R/U`, fallback licencia a `reportes` o `seguridad`; endpoint con `WithEmpresaGrafologiaPermissions`; clientes usa permisos de clientes; GPT-5.5 respeta activacion y limites del Chat IA empresarial | Buscar/crear cliente central, asociarlo al manuscrito, subir foto, usar camara, ajustar brillo/contraste/zoom, analizar con motor Go, analizar con GPT-5.5, verificar persona en resultado e historial, revisar que el reporte muestre angulos, separaciones y margenes, exportar HTML/PDF multipagina/Word/JSON/CSV/TXT, confirmar que otro `empresa_id` no carga el reporte, revisar modo claro y oscuro |
| OCR documental retirado | Paginas eliminadas: `web/administrar_empresa/ocr.html` y `web/super/configuracion/ocr.html` | Rutas retiradas: `/api/empresa/ocr` y `/super/api/config/ocr` | La tabla historica `empresa_ocr_documentos` deja de crearse en runtime; los datos previos solo quedan como evidencia historica si existen en una base antigua | Modulo `ocr`, pagina `linkOCR` y wrapper `WithEmpresaOCRPermissions` retirados del catalogo activo | Confirmar que el menu empresarial y super ya no muestran OCR; usar `soportes_compras_ia` con IA GPT-5.5 para captura de soportes |
| Bolsa | `web/administrar_empresa/bolsa.html` | `/api/empresa/bolsa`, `backend/handlers/bolsa.go` | Sin tablas nuevas; detecta pais con configuracion empresarial/facturacion, consulta proveedor de mercado desde backend y usa cache en memoria de 60 segundos | `bolsa:R`, pagina `linkBolsa`, wrapper `WithEmpresaBolsaPermissions`, fallback licencia a `reportes` o `finanzas` | Abrir desde Administrar empresa > Analisis y control, comprobar pais detectado, ver indicadores internacionales y locales, refrescar, validar errores saneados por indicador y que otro `empresa_id` no omite el wrapper |
| Impuestos | paginas de impuestos en Finanzas | `/api/empresa/impuestos?action=context|dashboard|upsert` | `empresa_impuestos_config` por `empresa_id`; catalogo base Colombia centralizado y prellenado en empresas nuevas/existentes de preproduccion; productos consumen el contexto para selector de impuesto; la pantalla calcula barra 0-100 de preparacion tributaria con faltantes visibles | facturacion:R para rol `contador`; administracion por roles superiores | Crear impuesto con admin, verificar defaults Colombia, confirmar que aparece en Productos si esta activo y aplica a ventas/ambos, aplicar a venta, reporte, revisar barra de preparacion y exportables; contador solo consulta dashboard/contexto y recibe 403 al guardar |
| Impresoras | `web/administrar_empresa/configuracion_impresora.html` | `/api/empresa/impresoras`, `/api/empresa/impresoras/resolver`, `/api/empresa/impresoras/agente`, `backend/handlers/empresa_impresoras.go` | `empresa_impresoras`, `empresa_impresoras_funcionalidades`, `empresa_impresoras_productos`, `empresa_impresoras_productos_reglas`, `empresa_impresoras_recetas`, `empresa_impresoras_dispositivos`, `empresa_impresoras_cola`; computador detectado por `localStorage.pcs_dispositivo_id` y agente local por `agente_id` | configuracion/seguridad para administrar; ventas para resolver y agente local | POS 80mm default, impresora por funcionalidad/producto/categoria/computador, crear prueba de cola con impresora automatica por computador, tomar pendientes con agente, marcar impreso/error, validar `empresa_id` |
| Buzon, tareas y chat empresarial | `web/administrar_empresa.html` campana, `web/administrar_empresa/panel.html` tarjetas de buzon/chat, `web/super/configuracion_avanzada.html` tarjeta Almacenamiento | `/api/empresa/buzon`, `/super/api/config/empresa_storage`, notificacion desde `/api/empresa/inventario/transferir` | `empresa_buzon_mensajes`, `empresa_buzon_adjuntos`, `empresa_chat_mensajes`; archivos en `/uploads/empresas/{carpeta}/mensajeria/buzon/...`; configs `empresa_storage.quota_enabled`, `empresa_storage.default_limit_mb`, `empresa_storage.warn_percent`, `empresa_storage.block_uploads_over_limit`, `empresa_storage.max_upload_mb` | `WithEmpresaSelfServicePermissions` por empresa; super storage con auditoria; adjuntar a mensaje existente exige destinatario o creador del mensaje | Enviar mensaje, asignar tarea, finalizar tarea con descripcion/evidencia, grabar audio, adjuntar foto/archivo, recibir notificacion por traslado de bodega, ver campana con contador, limpiar archivos antiguos desde super y validar bloqueo por cuota |
| Pagos Bre-B QR y datafonos | `web/administrar_empresa/finanzas_breb_qr.html`, `web/administrar_empresa/finanzas_breb_qr_tutorial.html`, carrito y configuracion pagos | `/api/empresa/finanzas/breb_qr`, `/api/empresa/carritos_compra`, `/api/empresa/estacion_prefs` | `empresa_estacion_prefs.estaciones_config.carrito_ui_global.pago_qr_*`, `carritos_compras`, `carrito_compra_abonos`, `empresa_finanzas_bancos_movimientos` | finanzas/ventas | Configurar cuentas receptoras Bre-B por caja, guardar payload oficial/plantilla, generar QR desde carrito, registrar pago manual recibido, listar ventas/abonos Bre-B y validar que dos cajas usen referencias diferentes |
| Nomina | `web/administrar_empresa/nomina_sueldos.html`, tutorial operativo `web/administrar_empresa/nomina_tutorial.html`, ayuda guiada `web/ayuda/tutorial_nomina.html` | `/api/empresa/nomina` acciones `config`, `parametros_legales`, `parametros_legales_aplicar`, `parametros_legales_auto`, `empleados`, `liquidaciones`, `dashboard`, `dashboard_colombia`, `seed_motel_calipso`, `documentos_electronicos_colombia`, `preparar_nomina_electronica` | `empresa_nomina_configuracion` con salario minimo y auxilio legal, `catalogo_legal_pais_versiones`, `catalogo_legal_pais_parametros`, `empresa_parametros_legales_aplicados`, `empresa_nomina_empleados`, `empresa_nomina_liquidaciones`, `empresa_nomina_colombia_*`; la pantalla calcula barra 0-100 con faltantes de empleados, documentos, salarios, parametros, liquidaciones, pagos, provisiones y lote DIAN | nomina/finanzas, siempre por `empresa_id` | Abrir parametros legales, validar version aplicada/disponible, aplicar manualmente, activar/desactivar autoactualizacion, abrir configuracion y validar salario minimo/auxilio legal preconfigurados, crear empleado y confirmar sugeridos, crear demo Motel Calipso, validar sedes, liquidar, PILA, pagos, desprendible, barra de preparacion y lote DIAN; abrir `Tutorial` desde nomina y confirmar que conserva `empresa_id`; abrir icono Ayuda desde nomina |
| Asistencia empleados | modulo asistencia | APIs de asistencia | asistencia, empleados, turnos | rrhh/asistencia | Entrada/salida, reporte, botones |
| Backup empresarial | backup/restablecimiento empresa | APIs backup empresa | datos operativos por empresa | super/seguridad | Previsualizar, backup previo, eliminar por fecha |
| Navegacion super | `web/super_administrador.html`, `web/js/super_administrador.js`, paginas activas del menu super | Rutas estaticas y APIs propias de cada pagina | Sin persistencia propia; contrato de navegacion | super_administrador | Confirmar que Panel aparece primero fuera de grupos, Auditoria global queda en Plataforma, Diagramas tecnicos queda debajo de Configuracion sin iniciar desplegado, Seleccionar empresa vive dentro del panel y el shell abre siempre `licencias_resumen.html`; Reportes globales queda fuera del menu super y Metricas de trafico no existe como pagina independiente |
| Mensajeria y alertas super | Grupo `Mensajeria y alertas` en `web/super_administrador.html`; `web/super/alertas_sistema.html`, `web/super/configuracion/alertas_licencia.html`, `web/super/formato_para_emviar_email.html`, `web/super/correos_masivos.html`, `web/super/agentes_de_mantenimiento_qutomatico.html`, `web/super/mantenimiento_sistema.html`, `web/super/email_corporativo.html` | `/super/api/alertas_sistema`, `/super/api/agentes_mantenimiento`, APIs de correos masivos, formatos de email y email corporativo Mailu | `super_alertas_config`, `super_alertas_eventos`, `super_mantenimiento_agentes`, `super_mantenimiento_agente_hallazgos`, `pcs_superadministrador.configuraciones`, `empresa_email_corporativo` | super_administrador | Abrir cada pagina desde el grupo, guardar formatos de email, configurar agente DIAN, ejecutar revision manual, probar alerta/licencia por Mailu, verificar que el formato de pago de licencia este accesible |
| Agente internet fiscal/nomina | Botones `Agente internet` en `web/administrar_empresa/nomina_sueldos.html` y `web/administrar_empresa/impuestos.html`; cuotas en `web/super/agentes_de_mantenimiento_qutomatico.html` | `/api/empresa/nomina/agente_internet`, `/api/empresa/impuestos/agente_internet`, `/super/api/agentes_mantenimiento?action=limits`, `backend/handlers/agente_internet_fiscal.go`, `backend/db/empresa_agentes_uso.go` | `empresa_agentes_uso_diario`, `configuraciones.agentes.empresa.*`; no aplica cambios automaticamente | nomina usa permisos de nomina; impuestos usa permisos de facturacion/impuestos; super configura cuotas | Guardar limites por empresa, ejecutar boton de nomina/impuestos, revisar propuesta actual vs sugerido y confirmar que no se aplican cambios sin aprobacion humana |
| Alertas sistema super | `web/super/alertas_sistema.html` | `/super/api/alertas_sistema` | `super_alertas_config`, `super_alertas_eventos` | super_administrador | Guardar checks, enviar prueba, historial |
| Portal publico e index super | Grupo `Portal publico e index` en `web/super_administrador.html`; editores `web/super/pagina_principal.html`, `web/super/informacion_de_modulos.html`, `web/super/informacion_de_la_empresa_y_de_los_sistemas_para_ia.html`, `web/super/configuracion/whatsapp_portal.html`; lectura `web/index.html` y `web/descripcion_de_los_sistemas.html` | `/super/api/pagina_principal`, `/super/api/informacion_de_modulos`, `/super/api/config/portal_chat_ia_info`, configuracion WhatsApp via `configuracion_avanzada` | `pcs_superadministrador.configuraciones` | super_administrador para editar; publico para leer portal | Abrir menu super, verificar grupo unico, cargar cada editor en iframe, abrir index y descripcion publica en pestaña nueva |
| Noticias portal | `web/noticias.html`, editor `web/super/noticias.html`, enlace `Noticias` en `web/menu.js` y menu `Portal publico e index` del super | `/api/public/noticias`, `/super/api/noticias`, `backend/handlers/pagina_principal_handlers.go` | `pcs_superadministrador.configuraciones` claves `super.noticias_portal.v1` y `super.noticias_portal.v1.updated_by` | publico lectura; super_administrador edicion | Abrir Noticias desde menu flotante, verificar portada/foto/feed en PC y celular, editar noticia DIAN en super, guardar, publicar/desactivar |
| Informacion modulos index | `web/super/informacion_de_modulos.html`, `web/index.html` | `/super/api/informacion_de_modulos`, `/api/public/informacion_de_modulos` | `pcs_superadministrador.configuraciones` | super_administrador | Editar vineta, ver index con fallback |
| Portal publico | `web/index.html`, rutas de respaldo `web/Informacion_de_contacto.html`, `web/privacidad_y_datos.html`, `web/quienes_somos.html`, `web/hotel_motel_domotica.html`, `web/js/portal_visits.js`, assets del carrusel en `web/img/portal-systems/` y fotos realistas en `web/img/portal-systems/realistic/` | `/api/public/*` | portal, visitas agregadas | publico/super lectura | Modo claro/oscuro, movil, abrir `Mas > Contacto/Privacidad y datos/Quienes somos/Sistema Hotel Motel` dentro de `index.html`, verificar rutas de respaldo, contacto publico, mapa en super y que `Mas sistemas` muestre foto realista distinta por tarjeta |
| Rappi | `web/administrar_empresa/rappi.html`, acceso `linkRappi` en Canales digitales y colaboracion | `/api/empresa/rappi?action=config|stores|orders|orders_sent|events|test|take|reject|ready`, `/api/public/rappi/webhook`, `backend/handlers/rappi.go` | `empresa_rappi_configuracion`, `empresa_rappi_ordenes` | `venta_publica:C`, wrapper `WithEmpresaVentaPublicaPermissions`; webhook publico exige `empresa_id` y firma HMAC si hay secreto | Guardar credenciales por referencia, probar token/stores, listar tiendas, traer ordenes READY/SENT, tomar/rechazar/listo, recibir webhook firmado, confirmar bitacora por `empresa_id`; venta interna real queda pendiente de mapeo de productos/caja por empresa |
| Docker y VPS | `deploy/`, `web/super/docker_portabilidad.html`, `scripts/crear_backup_vps.ps1` | `/super/api/docker_portabilidad`, `/super/api/vps_snapshots`, worker `super.vps_snapshot_worker`; backup externo por PuTTY/SSH desde PowerShell | paquete portable sin secretos; `super_vps_snapshots`; configuraciones `super.vps_snapshot.*`; copias fisicas en `backup/vps_snapshots`; nube via `rclone` externo; backups completos versionados en `D:\Backup vps PCS` con dump PostgreSQL, volumenes Docker, imagenes PCS, inventario, SHA256 y restaurador | super_administrador / operacion local autorizada | Exportar paquete portable, crear snapshot manual, descargar `.tar.gz`, verificar que no incluya `.env.platform` por defecto, configurar ruta `rclone`, revisar historial, retencion local/remota, ejecutar `.\scripts\crear_backup_vps.ps1`, validar manifest/tar y probar restauracion en VPS desechable antes de produccion |
| Apariencias / menu flotante | `web/menu.js`, paginas con selector de tema, calculadora, campana global de notificaciones y tickets | `/api/user/configuracion`; consume `/api/empresa/buzon?action=resumen` y marca leido con `action=leer` por `empresa_id` | `usuario_configuracion.apariencia`; notificaciones de buzon por `empresa_id` | Sesion autenticada cuando guarda en servidor; lectura visual de notificaciones empresariales; Juegos queda accesible desde Super administrador y el emulador dentro de Juegos | Cambiar tema en escritorio y movil; validar iframes con `dark-corporate`, `dark-absolute`, `dark-obsidian` y `light`; verificar que Calculadora este en la lista principal, que el menu no muestre compartir por correo/juegos/emulador, que la campana sea el primer item del panel y que el clic navegue al enlace relacionado |
## 2026-07-05 - WhatsApp interno, licencias y recordatorios

- Super administrador / Mensajeria y alertas:
  - UI: `web/super/configuracion/whatsapp_notificaciones.html`
  - API: `GET|PUT|POST /super/api/config/whatsapp_notificaciones`
  - Backend: `backend/handlers/whatsapp_notifications.go`
  - Alcance: notificaciones internas del sistema por evento; no reemplaza los enlaces publicos de WhatsApp para clientes.
- Super administrador / Vencimientos externos:
  - UI: `web/super/recordatorios_infraestructura.html`
  - API: `GET|PUT /super/api/recordatorios_infraestructura`
  - Backend: `backend/handlers/super_recordatorios_infraestructura.go`
- Administrar empresa / Licencia del sistema:
  - UI: `web/administrar_empresa/licencia_sistema.html`
  - Preferencia: `empresa_estacion_prefs.clave='licencia_notificaciones'`
- Administrar empresa / Rol cajero:
  - UI: `web/administrar_empresa/configuracion_rol_cajero.html`
  - Shell: `web/js/administrar_empresa.js`
  - Preferencia: `empresa_estacion_prefs.clave='estaciones_config'`, campo `cajero_auto_venta_directa`
- Inventario y compras / Bodegas y traslados:
  - Vista actual: `web/administrar_empresa/administrar_productos.html?view=bodegas`
  - Compatibilidad: `web/administrar_empresa/productos/bodegas.html`
