- 2026-06-14: `Ajustes visuales super empresa carrito` unifica el icono IA tipo tuerca sin circulo lateral, mueve analitica publica al panel de super, retira accesos publicos redundantes del super, agrega Noticias en Administrar empresa, suma envio por correo en codigos globales y mejora carrito escritorio/movil.
- 2026-06-14: `Ajuste fino carrito e IA PCS` cambia la etiqueta visible de busqueda de cliente a `Cliente` y aplica `/img/gpt.svg` al boton flotante IA del shell `administrar_empresa.html`, manteniendo una sola imagen IA para el proyecto.
- 2026-06-13: `Carrito abonos sin error secundario` evita que la carga de abonos falle por ordenamiento `pcs_ts` en PostgreSQL y muestre error interno despues de agregar productos.
- 2026-06-13: `Carrito busca productos sin importar mayusculas` corrige el filtro `q` de `/api/empresa/productos` para que la busqueda por nombre en PCS encuentre productos como `Arroz` aunque el cajero escriba `arroz`.
- 2026-06-13: `Carrito factura colombiana y pagos editables` alinea recibo POS/factura carta con columnas fiscales tipo Colombia (`Cod`, `Descripcion`, `Cant`, `UM`, `Vr Unit`, `IVA`, `Dto`, `Total`) y limpia los campos de pago al enfocar para evitar concatenar valores con `COP`.
- 2026-06-13: `Icono flotante IA unificado` reemplaza la variante con contenedor blanco en Administrar empresa por el mismo icono limpio usado en seleccionar empresa y el resto del proyecto.
- 2026-06-13: `Carrito puede vender sin stock por empresa` agrega un check de configuracion para permitir ventas aun sin existencia suficiente, propaga el flag al carrito y hace que la reserva de inventario permita saldo negativo solo cuando la empresa lo activa.
- 2026-06-13: `Finanzas sin bloque introductorio largo` reduce el hub financiero y la pestaña de creditos a titulos cortos, retirando el bloque descriptivo duplicado de Centro financiero y contable en el menu financiero.
- 2026-06-13: `Carrito pagos compactos y foco estable` convierte el detalle de medios de pago en tabla compacta, evita re-render completo al editar pago combinado, oculta controles opcionales no activados, elimina restos visibles de Taxi dejando Vehiculo, agrega edicion de cliente existente desde Nuevo cliente, alinea botones/IA del carrito y evita que el decorador global altere los botones de ayuda `?`.
- 2026-06-13: `URLs visibles opcionales en Administrar empresa y ayudas en cliente rapido` agrega una configuracion global super para que el shell empresarial pueda reflejar en la barra del navegador la URL real de la subpagina abierta sin perder el contenedor al recargar; tambien suma ayudas `?` en todos los campos del formulario `Nuevo cliente` del carrito para operacion mas guiada en PCS.
- 2026-06-13: `Snapshot y menu preservan super administrador PCS` fuerza el rol `super_administrador` para `powerfulcontrolsystem@gmail.com` tambien en el snapshot cacheado de autorizacion de `administrar_empresa`, purga de forma defensiva el usuario operativo duplicado y evita que la configuracion visual del menu oculte modulos al super administrador.
- 2026-06-12: `Pago combinado del carrito estable al hacer clic` evita que los campos Efectivo/Credito/Debito/Bre-B/Nequi/Otra transferencia/Credito cliente autollenen en `focus` y en `click` a la vez; ahora el autollenado ocurre solo por clic con bloqueo corto de render, corrigiendo saltos/bloqueos al probar pagos combinados en PCS.
- 2026-06-12: `Permisos contexto preservan super administrador PCS` refuerza `/api/empresa/permisos_contexto` para que `powerfulcontrolsystem@gmail.com` conserve siempre `super_administrador` aunque un contexto previo llegue como `cajero`, evitando que `administrar_empresa.html` oculte módulos al entrar a PCS desde `login.html`.
- 2026-06-12: `Purga usuario operativo reservado y cliente vertical en carrito` elimina automaticamente de `pcs_empresas.users` cualquier usuario empresarial con el correo reservado `powerfulcontrolsystem@gmail.com`, preservando intacto el administrador global en `pcs_superadministrador`; ademas fuerza el formulario `Nuevo cliente` del carrito a una sola columna vertical para operacion mas clara en `PCS`.
- 2026-06-12: `Cantidad del carrito visible con contraste fijo` agrega una celda dedicada para Cantidad y fuerza contraste claro/oscuro en `.cart-item-qty-input`, evitando que el numero desaparezca en el detalle de productos del carrito.
- 2026-06-12: `Super administrador no queda eclipsado por usuario operativo` preserva el rol `super_administrador` antes de aplicar roles de `users`, corrige `/api/account` para devolver `is_super` consistente y bloquea crear/asignar `powerfulcontrolsystem@gmail.com` como usuario operativo; la eliminacion remota del duplicado queda pendiente porque SSH al VPS agoto tiempo.
- 2026-06-12: `Carrito cantidades y pagos legibles` refuerza el CSS del carrito para que la cantidad editable de productos y los valores de pago combinado mantengan ancho, contraste y numeros visibles en temas oscuros/claros; validacion visual local con Chrome confirma cantidades `2` y `1,25` y medios Efectivo/Credito/Debito/Bre-B/Nequi visibles.
- 2026-06-12: `Fuente configurable en facturas y reportes` agrega configuracion por empresa para tamano de fuente POS/carta en facturas/recibos y reportes/cortes, persiste los valores en `empresa_configuracion_avanzada` y los aplica en carrito, ventas, facturas electronicas, corte de caja, reportes de turno e ingresos/egresos sin alterar XML DIAN ni totales.
- 2026-06-12: `Plan PCS tipo Siigo y eventos contables enriquecidos` agrega `documentos/plan_siigo_profesional_2026-06-12.md`, separa hitos precontables auditables de asientos monetarios obligatorios y refuerza payloads contables de facturacion, compras, finanzas y carritos con base, IVA, retenciones, total neto, forma/metodo de pago y tercero/cliente.
- 2026-06-11: `Nomina sin encabezado interno repetido` elimina de `nomina_sueldos.html` los bloques `Gestion laboral / Nomina y costo empresa` y `Ciclo laboral`; las subpaginas de nomina empiezan directamente en su tarjeta operativa.
- 2026-06-11: `Nomina sin tarjeta superior repetida` elimina del hub de nomina la tarjeta `Modulo de nomina` con metricas Base/Calculo/Pago/DIAN; el contenido de las subpaginas inicia arriba y se conserva el submenu lateral.
- 2026-06-11: `Configuracion del rol cajero` agrega en Configuracion > Ventas y cobro una pagina dedicada para administrar por empresa el perfil del cajero, cobro/caja, carrito POS, botones visibles, medios de pago, control de estaciones y usuarios cajeros; reutiliza `configuracion_operativa`, `estacion_prefs` y `roles_de_usuario` con filtro `empresa_id`.
- 2026-06-11: `Index tarjetas principales restauradas` recupera el orden visual base Punto de venta, Motel, Restaurante, Control por sensor, Hotel y Clientes/CRM en la pagina publica, y evita que las fotos se crucen por indice cuando llegan tarjetas dinamicas.
- 2026-06-11: `Login usuario errores especificos` corrige el primer ingreso de usuarios operativos para mostrar errores seguros con detalle y `request_id`, conserva protegido cualquier 500 no marcado como publico y repara defensivamente indices heredados de `users.email` para que la unicidad sea por `empresa_id`.
- 2026-06-11: `Registro usuario operativo sin version visible` retira la etiqueta tecnica `Version` del contrato en `login_usuario.html`; el titulo del contrato sigue configurable desde Super Administrador > Contrato y la version se conserva solo para trazabilidad interna.
- 2026-06-11: `Buzon usuarios visibles` corrige el selector de destinatarios del panel para listar nombre, rol y email de usuarios activos, administrador propietario, administradores compartidos y actor actual; el backend valida que cualquier admin destinatario tenga alcance real sobre la empresa antes de aceptar mensajes o adjuntos.
- 2026-06-11: `Roles personalizados por empresa` permite que el administrador cree roles propios desde Administrar usuarios, con nombre, descripcion y rol base global; el catalogo empresarial devuelve roles globales + roles de la empresa, valida asignacion por `empresa_id` y el motor de permisos resuelve el rol base para usuarios con rol personalizado sin mezclar empresas.
- 2026-06-11: `Tarjeta Creditos y cartera` agrega acceso rapido dentro del Centro financiero para abrir cupos por cliente, crear creditos, registrar abonos y consultar cartera; registra `linkCreditosTarjeta` en permisos y normaliza las subpaginas de creditos bajo el grupo financiero universal.
- 2026-06-11: `Chat IA lectura administrativa por empresa` permite a super administradores, administradores totales y administradores de empresa consultar datos reales de su propia empresa desde contexto backend filtrado por `empresa_id`; agrega respuesta directa para conteo de usuarios y mantiene cualquier modificacion limitada a funciones/endpoints PCS con permisos y confirmacion.
- 2026-06-11: `Total en letras y campos imprimibles por empresa` agrega conversion del total a texto activable por check, amplia los campos configurables de recibos/representaciones impresas y protege los campos obligatorios de factura electronica para que no puedan ocultarse cuando el documento fiscal aplica.
- 2026-06-11: `Ventas y facturas en bandeja unificada` muestra en `Ventas` si cada documento es `Solo venta`, `Venta con factura electronica` o `Factura electronica`, mantiene al cajero en una sola pagina para consultar/reimprimir/facturar desde venta y reemplaza los botones duplicados `Ver`/`Visualizar` por `Ver / imprimir`.
- 2026-06-11: `Produccion/MRP profesional y unificado` refuerza el modulo con demo PCS idempotente, tutorial operativo, resumen MRP, siguiente accion recomendada e importacion de recetas vendibles de Productos como BOM `POS-*`, dejando Productos como catalogo/POS e MRP como planeacion y ejecucion productiva.
- 2026-06-11: `Carrito busqueda por nombre con teclado` permite navegar resultados del campo `Busqueda por nombre` con flecha arriba/abajo y agregar con Enter el seleccionado o el unico resultado; el boton de pantalla completa queda solo como icono visible.
- 2026-06-11: `Busqueda por nombre separada en carrito` agrega un campo dedicado a la derecha de Codigo de barras/SKU para buscar productos por nombre y agregarlos desde la misma fila.
- 2026-06-11: `Notificaciones en menu flotante` mueve la campana visual al primer item del menu flotante, muestra badge en el boton principal y despliega ahi mismo el resumen del buzon por empresa.
- 2026-06-11: `Impresoras por computador` agrega asociacion de impresora por equipo detectado con `pcs_dispositivo_id`; la resolucion de impresion usa producto/receta/categoria, computador, funcionalidad y predeterminada por `empresa_id`.
- 2026-06-11: `Login de cajero sin selector manual` elimina la ventana de elegir caja al iniciar sesion; el cajero entra con la caja asociada al computador o queda pendiente de asociacion administrativa.
- 2026-06-11: `Carrito pantalla completa en tarjeta` mueve el icono de pantalla completa a una tarjeta compacta dentro del encabezado del cliente del carrito.
- 2026-06-11: `Ingresos/egresos manuales por rol` agrega checks en Configuracion operativa de cobro para que un cajero pueda registrar ingresos y/o egresos manuales sin abrir finanzas completas. El backend valida `empresa_id`, rol efectivo y `empresa_configuracion_operativa_roles` en `/api/empresa/finanzas/movimientos`.
- 2026-06-11: `Caja detectada por computador` agrega identificador local `pcs_dispositivo_id`, asignacion de caja por navegador/equipo para cajeros y check empresarial `caja_login_auto_por_computador` en Configuracion de estaciones.
- 2026-06-11: `Cajero busca ventas y facturas` agrega el acceso operativo `Buscar ventas y facturas` para rol cajero; permite consultar ventas/comprobantes, reimprimir y reenviar documentos sin abrir paginas administrativas de productos, clientes, finanzas o configuracion.
- 2026-06-11: `Buscador rapido del carrito` convierte el campo superior de Codigo de barras/SKU en buscador por codigo, SKU o nombre, con resultados seleccionables y agregado directo al carrito.
- 2026-06-11: `Campana en menu flotante` pone Notificaciones como primera accion del menu flotante, replica el contador de buzon no leido en el boton principal y abre la campana del panel empresarial al hacer clic.
- 2026-06-11: `Busqueda carrito por codigo o nombre` mejora el catalogo inteligente del carrito para buscar productos por codigo/SKU/codigo de barras o nombre, mostrar ambos codigos y priorizar coincidencias exactas.
- 2026-06-11: `Stop chat IA` agrega un boton Stop en el chat flotante para detener audio y abortar respuestas en curso con `AbortController`, evitando que respuestas tardias se muestren despues de cancelar.
- 2026-06-11: `Auditoria transversal de opciones nuevas` agrega eventos especificos por modulo para Bre-B QR, buzon, tareas, chat empresarial, adjuntos, preferencias transversales e impresoras/cola; amplia filtros de `auditoria.html`, severidad/retencion y el dataset `operativo_modulos_resumen` sin guardar contenidos sensibles ni secretos.
- 2026-06-11: `Compartir empresa con re-compartir opcional` agrega el check `Permitir que este administrador tambien pueda compartir esta empresa`, guarda `puede_compartir` en invitacion/acceso, habilita el icono solo para accesos autorizados y permite invitar correos no registrados creando una cuenta administrativa pendiente con enlace de registro.
- 2026-06-11: `Menu visible por empresa` agrega Configuracion > Menu visible para ocultar visualmente modulos del menu empresarial mediante checks guardados en `empresa_estacion_prefs.menu_visual_config`; el filtro se aplica despues de permisos/licencias y no reemplaza seguridad backend.
- 2026-06-11: `Atajos POS configurables por empresa` agrega mapa editable de F1-F12, ESC, ENTER, CTRL+B, CTRL+D, CTRL+P y ALT+F4 en Configuracion carrito; el carrito ejecuta busqueda de producto/cliente, descuento, cantidad, cobro, impresion, cajon, ayuda, inventario, recuperar/suspender y salida respetando la configuracion guardada por `empresa_id` en `estaciones_config.carrito_ui_global`.
- 2026-06-10: `Venta a credito desde carrito` agrega el medio `Credito cliente` en venta directa/estaciones, valida cupo activo por empresa y cliente, crea cartera real en `empresa_creditos` con `venta_origen_id`, no suma efectivo a caja y muestra monto/codigo/vencimiento en recibo/factura visual.
- 2026-06-10: `Carrito cliente fiscal y devolucion real` amplia el cliente rapido del carrito para personas y empresas con datos necesarios para facturacion electronica, permite buscar productos por nombre/codigo, editar cantidad en linea con recalculo real de totales/inventario y convierte `Devolver producto` en eliminacion real de la linea con liberacion de stock reservado.
- 2026-06-10: `Pago carrito en estaciones` refuerza `Pagar y cerrar carrito` con un manejador delegado para que el cobro funcione aunque la vista de estacion haya sido re-renderizada o decorada con iconos.
- 2026-06-10: `Pago carrito bloqueo obsoleto` limpia el estado visual `paymentInProgress` cuando no hay clic de pago en curso, permitiendo reintentar el cobro desde estaciones sin recargar manualmente.
- 2026-06-10: `Carrito imprime despues de cobrar` cambia `Pagar y cerrar carrito` para validar caja y registrar `pagar_estacion` antes de abrir ventanas de recibo/factura; asi los errores quedan visibles en la pagina y la impresion solo ocurre con pago exitoso.
- 2026-06-10: `Carrito vacio despues de pagar` hace que `Pagar y cerrar carrito` registre el pago, reabra/refresque la cuenta enfocada y la deje sin items, abonos ni cliente para el siguiente cliente, sin recarga manual.
- 2026-06-10: `IA OpenAI disponible en Docker` registra la clave OpenAI en archivos locales ignorados por Git y actualiza `sync_to_vps` para sincronizar `OPENAI_API_KEY` tambien en `deploy/.env.platform` remoto, de modo que Docker Compose la entregue al backend sin imprimir secretos.
- 2026-06-10: `Carrito orden productos antes de pago` intercambia las tarjetas `Productos agregados al carrito` y `Detalle del pago` para que venta directa/estaciones queden Cliente -> Productos -> Detalle del pago -> Acciones.
- 2026-06-10: `Carrito estable e IA con fallback de despliegue` conserva cliente, productos, detalle del pago y acciones en venta directa/estaciones aunque preferencias antiguas intenten ocultar o reordenar tarjetas; si ya hay carrito seleccionado y falla una sincronizacion secundaria, no desmonta la venta. `rs`/`sync_to_vps` propagan `OPENAI_API_KEY` como fallback seguro para IA y redactan secretos en logs.
- 2026-06-10: `Imagenes Mas sistemas sin panel superior` regenera las 51 fotos realistas de `web/img/portal-systems/realistic/` retirando el cuadro blanco/verde superpuesto en la parte superior derecha; el generador conserva la foto realista y el rotulo inferior.
- 2026-06-10: `IA flotante activa para todas las empresas` fuerza `chat_enabled=true` en `/api/chat_flotante/preferencias` cuando hay `empresa_id`, reaplica `chat_flotante.chat_enabled=1` en arranque con la migracion `20260610_chat_ia_activo_empresas_reaplicar` y evita que `localStorage` viejo oculte el nuevo icono circular flotante. Robot/secretaria siguen retirados y la voz se conserva.
- 2026-06-10: `Bascula electronica opcional` separa la bascula del lector del carrito en una tarjeta independiente y agrega el check empresarial `habilitar_bascula_electronica`, apagado por defecto. El carrito no muestra ni conecta Web Serial ni aplica peso si la empresa no activa la opcion desde Configuracion carrito.
- 2026-06-10: `Asesor comercial por etapas` permite configurar por asesor el porcentaje del primer ano, el porcentaje anual desde el segundo ano y los meses de renovacion. El backend registra comisiones de licencias como `primer_anio` o `renovacion_anual` segun la fecha de pago y bloquea comisiones fuera del plazo configurado; la UI super y `Mis clientes` muestran la regla vigente.
- 2026-06-10: `E-mail Corporativo independiente` mueve el webmail empresarial a `web/administrar_empresa/email_corporativo.html`, agrega el acceso `E-mail Corporativo` al menu de Administrar empresa, deja en el panel una tarjeta de notificaciones bajo Favoritos y suma `check_unread=1` en `/api/empresa/email_corporativo` para consultar no leidos reales por IMAP sin exponer claves.
- 2026-06-10: `Login oscuro campo contrasena` refuerza el CSS del login de administrador para que email y contrasena usen el mismo fondo oscuro en temas oscuros, incluyendo autocompletado de Chrome; `login.html` versiona `estilos.css` y `sw.js` pasa a `pcs-shell-v5` para renovar cache PWA.
- 2026-06-10: `Fotos realistas Mas sistemas` reemplaza las imagenes principales del carrusel publico por 51 JPG locales con trabajadores usando PCS segun cada titulo, mantiene los iconos pequenos como iconografia y agrega `scripts/generate_portal_realistic_images.ps1` para regenerarlas sin dependencias externas; laboratorio, clinica, odontologia, drogueria, taller, servicios tecnicos y lavanderia tienen escenas de oficio propias.
- 2026-06-10: `Plantilla colegio fuera del index` retira `Colegio o academia` del carrusel publico `Mas sistemas` filtrando `colegio_academia` en el catalogo local y en el render del index, sin borrar datos ni endpoints internos historicos.
- 2026-06-10: `Snapshot completo VPS` agrega en Super Administrador > Docker VPS una copia restaurable en un clic con proyecto portable, PostgreSQL, volumenes Docker, manifiesto, descarga segura, historial, retencion local/remota y subida opcional por `rclone` a Google Drive/Mega/OneDrive/S3 sin guardar tokens de nube ni incluir `.env.platform` por defecto.
- 2026-06-10: `Retencion de empresas vencidas` agrega configuracion super para preavisar y eliminar empresas no operativas con licencia base vencida despues de un plazo configurable; el reporte queda en `licencia_empresa_retencion_log` con `empresa_ref_id` para sobrevivir al borrado total.
- 2026-06-08: `DIAN firma XMLDSig/XAdES con C14N de contexto` corrige el calculo criptografico de `DigestValue` y `SignatureValue`: `SignedInfo`, `KeyInfo` y `SignedProperties` ahora se canonicalizan con el contexto completo de namespaces UBL del documento; la politica FEV vuelve a `politicadefirma/v1/politicadefirmav2.pdf`, que es la usada por los ejemplos principales de factura y notas. QA independiente: `lxml` valido los tres digest y `Node crypto` verifico `RSA-SHA256`. QA DIAN real: `SETP990000195` quedo aceptada con `StatusCode=00` y DIAN dejo de devolver `ZE02`; luego DIAN informo que el set de pruebas ya se encuentra aceptado.
- 2026-06-08: `DIAN politica XAdES v2` alinea la firma con los ejemplos oficiales de la caja DIAN: politica `politicadefirma/v2/politicadefirmav2.pdf`, `Description`, namespace `xades141` y `SignedDataObjectProperties/DataObjectFormat`. Se reconsultaron uno a uno los TrackId/ZipKey de Powerful Control System: los documentos `SETP990000135` a `SETP990000185` siguen en `Batch en proceso de validacion`; las pruebas nuevas `SETP990000186` a `SETP990000194` tienen rechazo final `StatusCode=99`. La factura limpia `SETP990000194` confirma que el bloqueo principal sigue en `ZE02`.
- 2026-06-08: `DIAN consumidor final y firma canonicalizada` ajusta el adquiriente consumidor final (`AdditionalAccountID=2`, `TaxScheme ZZ / No aplica`) y calcula digest XAdES de documento, `KeyInfo`, `SignedProperties` y `SignedInfo` sobre XML canonicalizado; la factura diagnostica `SETP990000189` llego a DIAN pero fue rechazada por `ZE02`/`FAK61`, por lo que no se enviaron notas hasta tener `StatusCode=00`.
- 2026-06-08: `DIAN prueba real uno a uno` reconsulta TrackId reales de Powerful Control System y confirma rechazo `StatusCode=99`; el generador corrige `PrepaidAmount`, literales DIAN con tildes, `ProfileID` por tipo, `PaymentMeans` y responsabilidad fiscal base, y bloquea notas sin factura aceptada como referencia. La firma XAdES base ahora firma el `SignedInfo` embebido y referencia `KeyInfo`; si DIAN mantiene `ZE02`, queda como siguiente frente implementar canonicalizacion XMLDSig/XAdES completa.
- 2026-06-08: `UBL DIAN realista` actualiza el generador Colombia para factura, nota credito y nota debito con estructura DIAN UBL 2.1, CUFE/CUDE SHA384, extensiones DIAN, SoftwareSecurityCode, QR y lineas correctas por tipo; el preflight bloquea XML incompleto antes de enviar y el parser SOAP conserva `ErrorMessageList` completo para mostrar rechazos reales. Se descargaron referencias oficiales DIAN localmente y se agrega `scripts/validar_dian_xsd.ps1` para validar contra XSD sin versionar los binarios oficiales.
- 2026-06-08: `Historial TrackId DIAN` agrega persistencia por `empresa_id` para cada TrackId/ZipKey recibido, endpoint `historial_tracks`, tarjeta visible de historial/reconsulta en el Centro de habilitacion DIAN y lectura de `StatusDescription` SOAP para mostrar `Batch en proceso de validacion` como pendiente sin guardar XML crudo ni secretos.
- 2026-06-08: `Centro de habilitacion DIAN` renombra el acceso `Pasar test DIAN` a `Centro de habilitación DIAN` en facturacion electronica y reemplaza el JSON crudo del diagnostico DIAN por un resumen legible con detalle tecnico plegado.
- 2026-06-08: `Facturacion electronica menu` mueve los accesos `Ecuador / SRI` y `Panamá / DGI` al final del menu dentro del subgrupo colapsado `Otros paises`; conserva la visibilidad por pais/licencia, corrige el texto visible `Panama` a `Panamá` en la pagina y normaliza `Proveedores de firma digital`.
- 2026-06-08: `Facturacion electronica Colombia` elimina de la pagina principal las tarjetas informativas `Pais detectado automaticamente` y `Perfil de facturacion`, y deja como primera tarjeta visible `Configuracion DIAN Colombia`, seguida por `Cargar firma electronica (Colombia / DIAN)`.
- 2026-06-08: `Ayuda contextual en formularios` agrega `web/js/form_field_help.js` y botones `?` compactos en el formulario de nuevo producto y en configuracion de facturacion electronica/DIAN; los mensajes explican que dato va en cada campo sin mostrar ni pedir secretos. Validacion visual local: Producto muestra 25 ayudas y DIAN/FE 20+16+35 ayudas con popover de 14px.
- 2026-06-07: `Bodega 1 por defecto` crea o reactiva automaticamente una bodega activa llamada `Bodega 1` para cada empresa nueva y aplica backfill `20260607_bodega_1_default` a empresas existentes, sin crear productos, existencias ni stock simulado.
- 2026-06-07: `IA oculta por defecto por empresa` oculta `Asistente IA`, `Centro IA empresarial`, `Renta IA` y `Compras con IA` en Administrar empresa, Centro financiero y Suite contador salvo habilitacion fina explicita por empresa; el backend aplica la misma regla en `permisos_contexto` y snapshot de autorizacion.
- 2026-06-07: `Modulo NIIF profesional` agrega `web/administrar_empresa/niif.html` con diagnostico de adopcion, politicas, mediciones, conciliacion contable-fiscal, checklist de cierre y notas exportables; queda enlazado desde Finanzas, Centro financiero y Suite contador con `linkNIIF` bajo `finanzas:R`, sin endpoint ni tablas nuevas.
- 2026-06-06: `DIAN WS-Security BinarySecurityToken` corrige la firma SOAP oficial para usar referencia directa al `BinarySecurityToken`, timestamp de 60 segundos con precision en milisegundos y KeyInfo alineado con la guia DIAN SoapUI; evita el rechazo `InvalidSecurity` causado por referencia `ThumbprintSHA1` en `SendTestSetAsync`/`GetStatusZip`.
- 2026-06-06: `DIAN set real sin simulacion` bloquea `simular=true` en `pruebas_dian`/`enviar_set_pruebas`, elimina el checkbox de simulacion de la pantalla, consulta `GetStatusZip` despues de `ZipKey`, mueve acciones DIAN criticas a aprobacion y reemplaza el test 2+2+2 simulado por un servidor SOAP local que valida envio real, ZIP firmado y acuse aceptado.
- 2026-06-05: `SOAP WS-Security DIAN` agrega seguridad WS-Security al sobre oficial `SendTestSetAsync` para endpoints `dian.gov.co`, usando Timestamp, BinarySecurityToken, firma RSA-SHA256, referencia `ThumbprintSHA1` y firma del header `wsa:To` segun WSDL; se valida el TestSetId real de Powerful Control System sin exponer secretos.
- 2026-06-05: `Pruebas DIAN con TestSetId visible` agrega el campo TestSetId/ZipKey en la pagina de pruebas DIAN, propaga ese valor al diagnostico, validacion y envio 2+2+2, muestra en backend el dato exacto que falta para envio real y mejora la confirmacion visual de ultima firma cargada con estado verde destacado sin exponer la clave.
- 2026-06-05: `Set automatico DIAN por empresa` alinea el preset operativo de software propio/proveedor con el objetivo registrado para Powerful Control System (30 facturas, 10 notas debito, 10 notas credito, minimo aceptado 1 factura), mantiene set gratuito e historico como ayudas, refuerza pruebas Go para que la simulacion no marque habilitacion aprobada, hace que `validar_credenciales` marque pendiente el `test_set_id` en habilitacion real y documenta que el correo de licencias solo adjunta factura electronica cuando el pago comercial es mayor que cero.
- 2026-06-05: `Eliminacion total de empresas` limpia tambien preferencias de orden del selector para todos los usuarios, invalida caches de licencia/empresa/comparticion y depura el respaldo local del navegador, para que una empresa eliminada no siga apareciendo a otros administradores.
- 2026-06-05: `Selector de empresas ordenable` agrega en `seleccionar_empresa.html` la opcion para mover tarjetas con clic sostenido o dedo, guardar el orden por administrador en `usuario_configuracion.selector_empresas_orden_json`, ordenar empresas activas e inactivas y restablecer el orden base sin cambiar permisos ni alcance multiempresa.
- 2026-06-05: `Centro de habilitacion DIAN` convierte la subpagina de pruebas en una pantalla operativa para pasar test DIAN por empresa: guarda modo de operacion, fechas del set, totales requeridos/aceptados por tipo documental y permite ejecutar set automatico o enviar factura/nota debito/nota credito manual de prueba. Powerful Control System queda configurada con los datos del portal DIAN sin documentar secretos.
- 2026-06-05: `OnlyOffice token JWT editor` corrige `editor_config` para enviar el JWT solo en `config.token`, como espera Document Server, y evita el error visual `The document security token is not correctly formed` al crear y abrir Word/Excel/PowerPoint.
- 2026-06-05: `Licencias acumuladas por pago repetido` hace que Epayco, Wompi y la activacion manual con descuento total conserven el ID real de la licencia empresarial creada y sumen vigencia sobre la fecha fin acumulada; dos pagos de un plan de 30 dias dejan 60 dias.
- 2026-06-05: `Facturacion electronica Colombia` agrega `llave_tecnica` en `empresa_dian_configuracion` y en el formulario DIAN para separar la clave tecnica del rango de numeracion del token de emisor. Se registro el set de habilitacion de la empresa interna Powerful Control System sin documentar secretos.
- 2026-06-05: `Ultima carga firma DIAN` agrega metadatos seguros de ultima firma cargada en `empresa_dian_configuracion` y muestra en la tarjeta de facturacion electronica fecha, archivo, formato, titular, serial y estado de clave sin exponer la contrasena del P12/PFX.
- 2026-06-05: `Apariencias Camaras/Energia/Compartidas` mueve los estilos de Camaras al CSS global, elimina colores fijos inline y normaliza tarjetas, formularios, estados y textos de Camaras, Energia solar y Empresas compartidas para temas claros y oscuros; validado visualmente en claro y oscuro absoluto.
- 2026-06-05: `Pasarelas de licencia Epayco/Wompi` repara la disponibilidad publica de Epayco para que credenciales opcionales legacy no apaguen la pasarela completa, habilita el fallback `checkout.js` con Public Key valida sin exigir `P_KEY`, y hace que Wompi solo aparezca disponible si public key e integrity key son realmente legibles para Web Checkout. Verificacion: `go test ./handlers -run "EpaycoCheckoutCredential|DefaultLicenciaPayment|Wompi|PaymentCredential" -count=1`.
- 2026-06-05: `Tutorial de nomina con narracion` agrega `web/ayuda/tutorial_nomina.html` como presentacion operativa con pasos numerados y botones de play en cada bloque de narracion; la pagina `web/administrar_empresa/nomina_sueldos.html` ahora tiene boton `Ayuda` con icono que conserva `empresa_id` al abrir el tutorial.
- 2026-06-05: `Alertas de certificado DIAN` guarda el vencimiento X.509 al cargar firma, agrega `action=vencimiento_certificado`, muestra estado/dias restantes en la pantalla de facturacion electronica y envia correo al administrador cuando el certificado esta vencido o a 30 dias de vencer.
- 2026-06-05: `Carpetas empresariales para firma electronica` crea la convencion `web/uploads/empresas/empresa_{id}_{slug}/`, asegura carpetas base al crear empresas, guarda la firma DIAN en `facturacion_electronica/firma_electronica/` con permisos privados y limpia esa carpeta al eliminar una empresa.
- 2026-06-05: `Empresa interna Powerful Control System` normaliza la licencia tecnica del SaaS a una vigencia fechada de 100 anos, valor cero, limites altos y modulos completos para que carrito, correo corporativo y configuracion carguen como en cualquier empresa; las consultas PostgreSQL de licencias toleran fechas heredadas vacias/no fechables y el super administrador validado puede operar globalmente sin quitar filtros por `empresa_id`; el index incorpora GRAFOLOGIX y Camaras/DVR en el respaldo editable de modulos.
- 2026-06-04: `Instalar app en login` refuerza el instalador PWA compartido para `login.html` y `login_usuario.html`: espera y reintenta la preparacion del service worker, consume correctamente el prompt nativo de Chrome/Edge, evita el mensaje prematuro de instalacion manual, conserva datos escritos antes de recargar y agrega validacion visual con `qa_pwa`.
- 2026-06-04: `Camaras y DVR` agrega el modulo empresarial `/api/empresa/camaras`, tabla `empresa_camaras`, pagina `Administrar empresa > Analisis y control > Camaras`, permisos `camaras`, soporte RTSP/ONVIF/HLS/WebRTC/MJPEG/iframe y configuracion para mostrar camaras como tarjetas o estaciones tipo camara.
- 2026-06-04: `Fallback Mailu para licencias` permite que el correo unificado de licencia use el correo corporativo Mailu cuando Gmail SMTP no descifra o falla, conservando adjuntos y destinatario validado.
- 2026-06-04: `Checkout licencia con correo cliente` agrega el campo Correo del cliente para activacion directa, Wompi y Epayco; `activar_sin_pago` lo valida antes de activar y lo usa para enviar la licencia. El resumen publico de licencias ahora muestra mensaje cuando un codigo ya fue usado, sin devolver 500.
- 2026-06-04: `Correo unificado licencia/factura` hace que la activacion de licencia envie un solo correo con el PDF de licencia y, si la compra comercial aprobada tiene valor mayor que cero, adjunte en el mismo mensaje el PDF resumen de la factura electronica. Las activaciones con total cero por prueba o descuento total no emiten factura electronica en el flujo final.
- 2026-06-04: `Factura electronica por compra de licencia` hace que una licencia comercial pagada emita factura electronica desde la empresa interna `Powerful Control System`, envie el correo al cliente y marque el pago Epayco/Wompi como facturado para evitar duplicados. La empresa interna se resuelve aunque exista como `Powerful Control Systen`; desde 2026-06-05 su licencia tecnica queda fechada a 100 anos con modulos completos.
- 2026-06-02: `GRAFOLOGIX medidas completas` agrega detalles tecnicos por metrica en `metricas_json` (inclinacion, espacios, margenes, continuidad, lineas, palabras, presion, regularidad y forma), los muestra en la tabla del modulo y los exporta en HTML/PDF multipagina/Word/JSON/CSV/TXT; tambien corrige porcentajes enteros en reportes.
- 2026-06-02: `Cajero carga carritos desde estaciones` corrige la resolucion de pagina para `/api/empresa/carritos_compra` y `/api/empresa/carritos_compra/items` cuando vienen con `perm_page=linkEstaciones` o `perm_page=linkCorteCaja`, y el carrito visual envia ese contexto al cargar carritos/items/abonos; el rol cajero puede operar estaciones/caja sin recibir acceso a la pagina administrativa de Carritos.
- 2026-06-02: `Ayuda con apariencias corregidas` refuerza `web/ayuda/ayuda.html` para que el centro de ayuda use tokens de tema en fondos, textos, chips, tarjetas internas, codigo, links, barra de volver y estado de busqueda, manteniendo contraste en temas claros y oscuros.
- 2026-06-02: `Carrito completo para cajero` mantiene el menu operativo del rol `Cajero` limitado a `Venta directa`, `Estaciones` y `Corte de Caja`, pero permite las APIs auxiliares del carrito para cargar productos, servicios, recetas, clientes, descuentos, propinas y comisiones. El cajero puede cobrar con el mismo carrito que ve un administrador sin recibir paginas administrativas de inventario o clientes.
- 2026-06-02: `Login usuario confirmado sin password` corrige el primer ingreso operativo cuando la invitacion ya marco el correo como confirmado pero aun falta crear contrasena; el sistema permite completar la contrasena con token valido, activa el usuario y prioriza en las busquedas la cuenta activa/confirmada para evitar tomar duplicados historicos no confirmados.
- 2026-06-02: `Login usuario aplica rol cajero` normaliza `Caja`, `Caja principal` y variantes como alias de `cajero`, deduplica esas opciones en el selector de roles prefiriendo mostrar `Cajero`, y restringe el menu empresarial del cajero a `Venta directa`, `Estaciones` y `Corte de Caja`.
- 2026-06-02: `Primer ingreso usuario operativo pendiente` corrige `/api/empresa/usuarios/establecer_password` para que un usuario invitado, pendiente e inactivo pueda crear su contraseña con token valido; al completar se confirma el correo, se consume el token y se activa la cuenta. Los usuarios ya confirmados e inactivos siguen bloqueados.
- 2026-06-02: `Usuario existente visible para reenviar confirmacion` hace que `/api/empresa/usuarios` responda el registro existente cuando se intenta crear un correo ya invitado en la misma empresa; Administrar usuarios recarga inactivos/pendientes, resalta el usuario y deja visible `Reenviar confirmacion`. Tambien se refuerza la migracion para quitar constraints/indices heredados de `users.email` global y conservar la unicidad por `(lower(email), empresa_id)`.
- 2026-06-02: `Deducido de impuesto en factura impresa` agrega en Configuracion > Impresora > Documento de venta un check por empresa para mostrar base gravable e impuesto deducido en recibos/facturas impresas; usa los valores ya calculados del carrito y no cambia el XML ni la validacion legal DIAN.
- 2026-06-02: `Carrito COP sin decimales` hace que el carrito detecte Colombia/COP y muestre/edite precios, abonos, valores por medio de pago, QR y totales sin `.00`; backend normaliza pagos, pagos mixtos, abonos, offline y precios unitarios del carrito a pesos enteros cuando la moneda es COP.
- 2026-06-02: `Pago combinado desde valores del carrito` hace que la tarjeta de Efectivo/Debito/Credito/Otro cargue el total o saldo restante al tocar una casilla, permita distribuir el pago entre medios y muestre en la misma tarjeta el aviso `La sumatoria debe ser igual al total` cuando no cuadra.
- 2026-06-02: `Domotica en Analisis y control` mueve el acceso de Domotica dentro del grupo `Análisis y control` del menu de Administrar empresa, conservando ruta, permisos y submenu.
- 2026-06-02: `Usuarios empresariales por empresa` elimina la restriccion global antigua `users.email UNIQUE` en PostgreSQL y crea `ux_users_lower_email_empresa` para que un mismo correo pueda recibir invitaciones en empresas distintas sin romper el aislamiento por `empresa_id`.
- 2026-06-01: `GRAFOLOGIX con GPT-5.5` agrega en la pantalla de grafologia el boton `Analizar con GPT-5.5`, reutiliza el modelo empresarial `openai:gpt-5.5` ya configurado, valida limites diarios por empresa y muestra la respuesta IA separada del informe local Go/Tesseract.
- 2026-05-30: `Informacion de contacto corporativa` rediseña la pagina publica de contacto sin logo, agrega vision empresarial para pequenas empresas, mas informacion de la empresa y secciones de acompanamiento comercial.
- 2026-05-30: `Domotica en paginas separadas` cambia el submenu de Domotica para abrir vistas independientes por `pagina`, muestra solo la seccion elegida y elimina del resumen operativo los enlaces duplicados internos.
- 2026-05-30: `Ayuda y APIs actualizadas` agrega pagina visible de ayuda de APIs, guia tecnica `documentos/api/ayuda_apis.md`, seccion de APIs en el centro de ayuda y corrige la ayuda contextual del carrito.
- 2026-05-30: `Documentacion del carrito unificado` actualiza contexto, mapa, flujos, decisiones, descripcion del proyecto, estructura del codigo e indice documental con pantalla completa de venta directa y fondo diferenciado del carrito.
- 2026-05-29: `Fondo diferenciado en carrito` hace que el fondo estructural del carrito sea mas oscuro que las tarjetas en todas las apariencias, sin recuperar sombras ni efecto 3D.
- 2026-05-29: `Venta directa en pantalla completa` agrega en el carrito de venta directa un boton con icono para entrar/salir de pantalla completa y habilita `fullscreen` en el iframe de Administrar empresa.
- 2026-05-29: `Panel empresarial sin titulo repetido` cambia el kicker del hero en `web/administrar_empresa/panel.html` de `Panel de empresa`/`Panel de {empresa}` a `Centro operativo`, conservando la descripcion con el nombre de la empresa debajo.
- 2026-05-29: `Auditoria profesional de Plantillas` actualiza `tools/professional_audit.mjs` para leer `web/js/plantillas_nuevas_catalogo.js` y el global `PCS_NUEVAS_PLANTILLAS`, eliminando el error por ruta antigua `nuevos_verticales_catalogo.js`.
- 2026-05-29: `Seleccionar empresa con Plantillas` corrige referencias JS heredadas a `vertical` en `seleccionar_empresa.js` y preconfiguraciones super para que el selector cargue empresas sin el error `vertical is not defined`.
- 2026-05-29: `Centro de mando super apilado` cambia `web/super/licencias_resumen.html` para que resumen, KPIs, prioridades y accesos clave queden en una sola columna, una tarjeta debajo de la otra.
- 2026-05-29: `Super administrador en una columna` agrega el acceso visible `Asesor en ventas` dentro de Licencias, registra `/super/asesor_comercial.html` como pagina valida del frame y fuerza el menu agrupado del panel super a mostrarse en una sola columna.
- 2026-05-29: `Idempotencia al crear empresas` evita duplicados por doble clic o POST concurrente en `/super/api/empresas`, bloquea visualmente el formulario mientras procesa y documenta la norma backend para altas/acciones criticas.
- 2026-05-29: `Plantillas empresariales` renombra el modulo de soluciones empresariales a Plantillas en rutas, paginas, JS, handlers, pruebas y documentacion activa; conserva aliases tecnicos internos solo donde son contrato de datos o compatibilidad.
- 2026-05-29: `Tarjeta Domotica configurable en carrito` separa modulo, boton y tarjeta automatica; agrega `mostrar_tarjeta_domotica_carrito` en la configuracion global y por estacion, y el carrito refresca Domotica al volver a la ventana para mostrar aparatos recien agregados.
- 2026-05-28: `Auditoria especial super administrador` agrega `/super/auditoria_super_admin.html` al panel super, restringe `scope=super_panel` a roles super, registra navegacion/edicion visual del panel y envuelve APIs sensibles de configuracion super con `WithSuperAuditoria` sin persistir secretos.
- 2026-05-28: `Auditoria global del selector` agrega `Seleccionar empresa > Auditoria`, tabla `super_auditoria_eventos`, endpoint `/super/api/auditoria`, registro automatico de movimientos globales y eventos visuales del selector, con alcance por administrador y exportacion CSV/JSON.
- 2026-05-28: `Energia solar multiempresa` agrega modulo empresarial para Victron VRM, SMA Sunny Portal, SolarEdge Monitoring y gateway local, con catalogo de baterias Powerwall/BYD/Pylontech/Enphase/Victron, alertas por correo, lecturas, eventos y documentacion.
- 2026-05-28: `Descarga de informacion integrada al selector` hace que el boton de descarga de cada empresa abra `descargar_informacion_de_la_empresa.html` en el panel derecho de `seleccionar_empresa.html`, respete la apariencia activa, permita descargar backup completo JSON y tenga boton para regresar al listado.
- 2026-05-27: `Reportes globales con alcance del selector` hace que `/super/api/reportes_globales` liste y consolide solamente las mismas empresas visibles en `seleccionar_empresa.html`: propias, delegadas o compartidas, sin exponer todas las empresas por rol super.
- 2026-05-27: `Reportes globales profesionales` simplifica `seleccionar empresa > Reportes globales` a rango de fechas, selector de reporte, formato y acciones Ver/Exportar/Imprimir/Email; mantiene seleccion multiple de empresas, graficos comparativos, resumen por empresa y exportacion PDF/XLS/CSV/TXT/JSON.
- 2026-05-27: `Vista previa de modulos` corrige las vinetas del super administrador para que usen variables de tema y se vean consistentes en apariencias claras y oscuras, incluso con contenido predeterminado.
- 2026-05-27: `Tema corporativo oscuro` agrega `Corporativo Oscuro` al selector de apariencias del menu flotante, reconoce `dark-corporate` en paginas embebidas y corrige en celular el panel de `Blanco Corporativo` para que conserve fondo y texto claros.
- 2026-05-27: `Empresas compartidas en administrar empresa` agrega la pagina `web/administrar_empresa/empresas_compartidas.html` y el boton `Empresas compartidas` en Administracion para ver accesos e invitaciones de la empresa actual y revocarlos/desactivarlos.
- 2026-05-27: `Ayuda en administradores` agrega una descripcion bajo el titulo de Administradores aclarando que la invitacion permite administrar empresas y que compartir una sola empresa se hace desde el icono Compartir de la tarjeta.
- 2026-05-27: `Estado de invitacion en administradores` muestra en `Administradores` si la invitacion esta `Aceptada`, `Pendiente`, `Revocada`, `Rechazada` o `Expirada`, separando ese dato del estado de la cuenta.
- 2026-05-27: `Administradores del selector filtrados` hace que `seleccionar_empresa.html > Administradores` abra la pagina con `scope=principal`, por lo que incluso un rol super ve alli solo los administradores que invito desde ese contexto; el panel super conserva la vista global.
- 2026-05-27: `Super administradores por invitacion` exige invitacion con token para crear nuevos super administradores; al aceptar el registro conservan rol `super_administrador` y entran al modulo super.
- 2026-05-27: `Delegacion de portafolio entre administradores` permite que un administrador registrado reciba las empresas de otro administrador sin perder ni mezclar sus propias empresas; si el correo no existe se mantiene invitacion con registro, y si ya existe se activa el acceso compartido de inmediato con aviso por correo.
- 2026-05-27: `Invitacion de administradores delegados` cambia el alta desde Administradores: ahora envia correo de invitacion con token, el invitado debe abrir el enlace, completar registro y crear clave antes de poder iniciar sesion.
- 2026-05-27: `Administradores por administrador principal` permite que un administrador principal vea y gestione solo los administradores que agrego desde `seleccionar_empresa.html`; esos administradores delegados ven las empresas del principal como administracion delegada, sin permisos de propietario para compartirlas.
- 2026-05-27: `Checklist seguridad endpoint multiempresa` agrega `documentos/checklist_seguridad_endpoint_multiempresa.md` como requisito para endpoints empresariales; cubre sesion, `empresa_id`, permisos, licencias, SQL aislado, auditoria y pruebas negativas de cruce entre empresas.
- 2026-05-27: `Contexto operativo Codex` agrega `documentos/contexto_codex.md`, `mapa_modulos.md`, `flujos_operativos.md`, `comandos_codex.md` y `decisiones_tecnicas.md`; `AGENTS.md` ahora los referencia como ruta rapida obligatoria antes de cambios relevantes.
- 2026-05-27: `Alertas sistema de registros y empresas` agrega dos opciones en Super administrador > Alertas del sistema para enviar correo a `powerfulcontrolsystem@gmail.com` cuando se registre un administrador desde `login.html` y cuando un administrador cree una empresa nueva.
- 2026-05-27: `QR DIAN en factura o recibo` agrega un check por empresa para imprimir al final del recibo/factura un QR de consulta DIAN cuando la venta electronica tenga CUFE/CUDE o codigo de validacion; usa `/vendor/qrcode.min.js`, `PCSPrint` y no agrega dependencias ni tablas.
- 2026-05-27: `Analitica publica solo en super administrador` quita del index la tarjeta visible `Visitas al portal`; el portal mantiene solo el registro invisible de visitas agregadas y la visualizacion queda en el panel super administrador.
- 2026-05-27: `Mapa real en analitica publica` reemplaza el mapa esquematico de `Visitas al portal por pais` por un SVG mundial real basado en Natural Earth, compartido entre el index y el panel super administrador, manteniendo marcadores por pais y sin cambiar backend ni privacidad.
- 2026-05-27: `Vinetas visibles en modulos del index` refuerza las viñetas de `Modulos y caracteristicas principales` con un punto visual propio que respeta apariencia clara y oscura.
- 2026-05-27: `Informacion editable de modulos del index` agrega `/super/informacion_de_modulos.html` para que el super administrador edite titulo, iconos y vinetas de `Modulos y caracteristicas principales`; el index lee `/api/public/informacion_de_modulos` con fallback estatico.
- 2026-05-27: `Login administrador con logo imagen` reemplaza el titulo textual del login por `web/img/titulo-powerful-control-system-login.png` con tamano pequeno y responsive, sin cambiar autenticacion.
- 2026-05-27: `Index subtitulo POS domotica` actualiza el subtitulo del encabezado publico a `Sistema de Facturacion Electronica con domotica integrada`.
- 2026-05-27: `Index modulos con vinetas reales` convierte cada caracteristica de las tarjetas de modulos principales en elementos `ul/li` con vinetas visibles y compactas, sin cambiar backend ni datos.
- 2026-05-26: `Cliente general configurable en carrito` agrega por empresa el campo `cliente_general_nombre` para nombrar ventas sin cliente registrado, por defecto `Cliente General`, y lo usa en carrito, listados e impresiones sin crear tablas ni endpoints.
- 2026-05-25: `Creditos y cartera con submenu` separa el modulo en subpaginas internas para panel, nuevo credito, cartera, morosidad, riesgo/limites, operaciones, aprobaciones y estado de cuenta, manteniendo `/api/empresa/creditos` y permisos `finanzas:C`.
- 2026-05-25: `Licencia del sistema` agrega una pagina descargable por empresa en `Administrar empresa > Administracion`, con descarga TXT/HTML e impresion PDF usando el contexto real de licencia y permisos.
- 2026-05-25: `Menu flotante` elimina el enlace `Ayuda administrador`; se conserva `Crear ticket de ayuda` y no cambian permisos ni endpoints.
- 2026-05-25: `Orden menu administrar empresa` mueve `Finanzas y cumplimiento` para quedar justo debajo de `Inventario y compras`, sin cambiar permisos, rutas ni endpoints.
- 2026-05-25: `Navegacion finanzas y paginas huerfanas` agrega accesos directos a Creditos y cartera y Gestion de cobranza en Finanzas y cumplimiento, suma ambos a los accesos rapidos del centro financiero y conecta Chat IA, Configuracion guiada e Integraciones en sus menus correspondientes sin cambiar permisos ni endpoints.
- 2026-05-25: `Index modulos mas compactos` ajusta la seccion `Modulos y caracteristicas principales` para que las tarjetas usen columnas mas anchas, menor padding interno, iconos mas compactos y caracteristicas en flujo punteado sin huecos artificiales.
- 2026-05-25: `Licencia gratis 15 dias transaccion PostgreSQL` evita que la activacion valor cero registre dos marcas gratis para la misma empresa cuando la licencia base se copia a la empresa; ahora guarda solo la licencia activa asignada y no deja la transaccion abortada en `commit unexpectedly resulted in rollback`.
- 2026-05-25: `Checkout licencia gratis sin Wompi innecesario` evita cargar tarjetas/terminos de Wompi cuando el resumen ya esta en total cero y la licencia se activa sin pasarela, eliminando el 502 residual de `/wompi/terms` en este flujo.
- 2026-05-25: `Licencia gratis 15 dias reparada` corrige la activacion sin pago en PostgreSQL para que `licencias_activaciones_gratis.id` tenga secuencia/default, habilita como publicos el resumen/activacion valor cero del checkout, tolera reintentos cuando la licencia ya quedo activa y evita errores no capturados del service worker ante fallos de red.
- 2026-05-25: `Super administrador analitica al final` ajusta el layout del panel super administrador para que la tarjeta `Analitica publica / Visitas al portal por pais` quede debajo del panel principal y no encima de la primera vista.
- 2026-05-25: `Login y registro administrador verificados` valida en produccion el login por correo, registro de nuevo administrador desde `login.html` e inicio OAuth Google en escritorio y movil; `web/estilos.css` corrige el titulo del registro administrador para que no se corte en celular.
- 2026-05-25: `Index modulos con caracteristicas punteadas` convierte las descripciones de cada modulo principal en listas justificadas con punto negro grande por caracteristica.
- 2026-05-25: `Index Mas sistemas` unifica la imagen grande de todas las tarjetas del carrusel `Mas sistemas` usando la imagen de la tarjeta POS, sin cambiar logos, textos ni enlaces.
- 2026-05-25: `Index documentos electronicos` muestra los documentos electronicos dentro de la lista de modulos como elementos separados con `*` al inicio de cada item.
- 2026-05-25: `Portal y super admin con analitica compartida` agrega iconos medianos a cada modulo principal del `index`, reutiliza `web/js/portal_visits.js` para mostrar el mismo contador/mapa en el panel de super administrador sin registrar visitas duplicadas, y mejora el mapa con halos de marcador.
- 2026-05-25: `Selector de empresa` invierte el orden visible de la lista `Tipo de empresa` al presionar `Agregar Empresa` en `seleccionar_empresa.html`, sin cambiar backend ni catalogo de tipos.
- 2026-05-25: `Index contador de visitas compacto` rediseña la tarjeta del contador por pais para que quede centrada y mas pequena, con total, mapa y ranking dentro del mismo bloque; el mapa usa una silueta mundial SVG mas realista con referencias geograficas y marcadores por pais.
- 2026-05-25: `Index modulos en lista` convierte la seccion publica de modulos en una lista de caracteristicas principales e incluye los documentos electronicos soportados: factura electronica, notas credito/debito, documento soporte, notas de ajuste, nomina electronica, documentos equivalentes/POS electronico, contingencia y eventos RADIAN para Colombia, ademas de documentos base para Panama y Ecuador.
- 2026-05-22: `Index modulos principales` simplifica la seccion publica de modulos del portal a un unico parrafo con funciones principales: inventario, compras, datafonos, cajon monedero, cajas simultaneas, IA, control electrico, sensores, finanzas, factura electronica, impuestos y modulo del contador, eliminando la grilla extensa de tarjetas.
- 2026-05-21: `Contador de visitas por pais` agrega en `index.html` un contador total, ranking y mapa mundial con colores por pais; `/api/public/portal_visitas` guarda solo agregados por pais/fecha en `portal_visitas_paises`, sin IP ni datos personales.
- 2026-05-21: `Iconos globales de botones` agrega un decorador comun para que los botones de las paginas HTML del sistema reciban iconos de color relacionados con su funcion; el servidor estatico inyecta `/js/button_icons.js` en HTML sin tocar assets ni duplicar scripts.
- 2026-05-21: `Iconos de botones en carrito` agrega iconos de color automaticos a los botones del carrito unificado, incluyendo botones generados dinamicamente por JS, sin cambiar endpoints, permisos, tablas ni dependencias.
- 2026-05-21: `Impuestos en Finanzas y cumplimiento` mueve el acceso de `Impuestos` desde Configuracion al grupo financiero del panel empresarial y lo agrega tambien dentro del Centro financiero y contable.
- 2026-05-21: `Emisoras por pais` actualiza el modulo de radio online para detectar Panama o Ecuador, mostrar 10 emisoras principales del pais y permitir emisoras personalizadas guardadas por empresa en `empresa_estacion_prefs` mediante `/api/chat_flotante/preferencias`.
- 2026-05-21: `Conectividad operativa` agrega aviso global cuando el navegador pierde internet y confirmacion cuando vuelve la conexion; tambien registra en auditoria empresarial los eventos `internet_perdido` e `internet_restaurado` con cola local hasta recuperar internet.
- 2026-05-21: `Conectividad por contexto` diferencia modulos normales y caja offline: los modulos sin operacion offline piden esperar a que vuelva internet, mientras el carrito/caja con facturacion offline activa permite seguir vendiendo e imprimir provisionalmente.
- 2026-05-21: `Vuelto en carrito` compacta el mensaje de cambio de efectivo a `Vuelto:` / `Falta:` con texto mas grande y bloque visual corto dentro de valores por medio de pago.

- 2026-05-21: `Impresoras por producto` agrega reglas masivas para enviar pedidos a impresoras por `todos los productos` o por `categoria`, manteniendo prioridad `producto especifico -> categoria -> todos -> funcionalidad -> predeterminada`; la pantalla de Configuracion > Impresora permite administrar las tres opciones sin dependencias nuevas.

- 2026-05-21: `Recetas de productos` consolida el modulo compuesto como recetas: ruta `/api/empresa/recetas_productos`, pagina `administrar_empresa/recetas_productos.html`, tablas `recetas_productos*`, items de carrito `tipo_item=receta`, permisos `linkRecetasProductos` e impresoras por receta. Mantiene versionado, costo teorico/real y descuento de inventario por ingredientes.

- 2026-05-21: `Carrito compacto` reorganiza la tarjeta de codigo/SKU y cantidad para dejar botones en la misma fila en escritorio, y compacta el panel de cliente con mas columnas, cliente actual en el encabezado y alta rapida en una sola grilla responsive.

- 2026-05-21: `Cliente en carrito` deja el formulario de nuevo cliente oculto por defecto y separa el buscador visible con opciones por nombre o por NIT/cedula/identificacion.

- 2026-05-21: `Botones del carrito` agrega configuracion para visualizar u ocultar el panel de cliente y cada boton operativo de la barra de acciones: descuentos, cambiar tarifa, control electrico, cancelar, taxi, clientes, abonos y vehiculo.

- 2026-05-21: `Alerta visual de carrito` agrega configuracion por empresa/estacion para mostrar u ocultar el check de alerta, definir minutos y decidir si inicia activada por defecto; queda apagada por defecto y sin tablas nuevas.

- 2026-05-21: `Clientes en carrito` elimina textos redundantes del panel, convierte el campo de nombre en buscador de clientes existentes y asigna automaticamente el cliente elegido o recien creado al carrito activo.

- 2026-05-21: `Carrito movil` prioriza en celular la tarjeta de busqueda/agregado de productos por codigo o boton, antes de totales y demas tarjetas operativas; no cambia el flujo desktop, APIs ni permisos.

- 2026-05-21: `Panel de administrar empresa movil` fuerza refresco de cache PWA/service worker a `pcs-shell-v4`, pide navegaciones y JS/CSS con red primero sin cache vieja, y actualiza la ayuda para que no reaparezcan indicadores economicos antiguos en celulares.

- 2026-05-20: `Facturacion electronica Colombia` agrega catalogo DIAN de documentos electronicos y eventos RADIAN en la pagina por pais, separa obligaciones que normalmente prepara el contador y guarda la seleccion por empresa en `facturacion_electronica_pais.campos_pais_json`; no agrega tablas, permisos ni dependencias.

- 2026-05-20: `Centro de reportes` reemplaza las tarjetas de reportes por un selector y una lista compacta, agrega vista previa imprimible segun papel POS 80mm o carta, y mantiene exportacion directa a PDF, Excel, CSV, JSON y TXT desde `/api/empresa/reportes`.

- 2026-05-20: `Traslado entre bodegas` corrige el flujo de Inventario > Bodegas para que la tarjeta de traslado y el kardex se muestren en la vista correcta, y el backend mueve stock entre bodegas usando wrappers SQL compatibles con PostgreSQL dentro de la transaccion.

- 2026-05-20: `Nomina profesional` robustece Nomina Colombia dentro de `nomina_sueldos`: catalogo amplio de conceptos, novedades aprobadas aplicadas a liquidacion, aprobacion/rechazo de novedades y seed demo Motel Calipso con empleados, asistencia, liquidaciones, PILA y pagos simulados.

- 2026-05-20: `Nombres configurables de estaciones` agrega singular/plural por empresa en Configuracion > Estaciones y aplica esos textos en Administrar empresa, estaciones, carrito y configuracion del carrito. Las preconfiguraciones por tipo de empresa guardan el recurso operativo adecuado y las empresas existentes con etiquetas genericas se normalizan por tipo: restaurantes `Mesa/Mesas`, moteles y hoteles `Habitacion/Habitaciones`, lavaderos/talleres `Bahia/Bahias`, entre otros.

- 2026-05-20: `Datáfonos POS multiempresa` agrega backend Go para configurar Redeban, CredibanCo, Bold y BBVA por empresa, iniciar/consultar pagos HTTP/JSON con endpoints contractuales, validar monto/referencia y aplicar confirmaciones aprobadas al carrito/caja del usuario. Las claves se referencian con `env:*`, sin dependencias nuevas.

- 2026-05-20: `Pagos QR en carrito` agrega configuracion de cuentas receptoras para Bre-B, Nequi u otros proveedores y boton `QR de pago` en el carrito. El QR se genera localmente con `/vendor/qrcode.min.js`, puede usar payload oficial/plantilla del proveedor con placeholders y registra el cobro como transferencia bancaria con referencia automatica, sin tablas ni dependencias nuevas.

- 2026-05-20: `Facturacion offline para carritos` agrega modo activable por empresa para cerrar e imprimir ventas sin internet, guardar cola local y sincronizar luego por `/api/empresa/offline_ventas`; bloquea el pago si el mismo carrito ya tiene una venta offline pendiente, y actualiza carrito, inventario, caja del usuario, metricas y documento de venta sin mezclar varias cajas de la misma empresa.

- 2026-05-19: `Codigos de descuento` agrega acciones por cupon para enviar por correo con SMTP configurado y compartir por WhatsApp con mensaje listo; mantiene permisos existentes, modo de prueba de correos y aislamiento por empresa.

- 2026-05-19: `Backups empresariales` agrega reinicio seguro de datos operativos por fecha o todos los tiempos, con previsualizacion, backup previo opcional, confirmacion exacta y proteccion de configuracion, usuarios, permisos, impresoras, integraciones y preferencias del sistema.

- 2026-05-19: `Codigos de descuento y asesor` deja los cupones del carrito como consumo unico por empresa, sin liberar cupo al anular/reabrir ventas, y hace que la promocion por codigo de asesor en checkout de licencias solo aplique una vez por empresa cuando la promocion global tenga porcentaje activo.

- 2026-05-19: `Reporte de turno POS` compacta el ticket 80mm usando dos columnas en datos del turno, resumen financiero y detalle de ventas, reduciendo longitud del comprobante sin cambiar datos ni endpoints.

- 2026-05-19: `Docker VPS` agrega pagina en Super Administrador > Plataforma para revisar el paquete portable y descargar un `.tar.gz` del proyecto base Docker. El endpoint `/super/api/docker_portabilidad?action=status|download` queda exclusivo de `super_administrador`, la imagen backend guarda `/app/project_export` y la descarga excluye secretos, `.env`, llaves, uploads, descargas, backups, logs, caches y datos runtime.

- 2026-05-19: `Facturacion electronica Ecuador` agrega pagina independiente `Ecuador / SRI` y endpoint `/api/empresa/facturacion_electronica/ecuador`; el perfil EC usa SRI con RUC, establecimiento, punto de emision, ambiente SRI 1/2, firma electronica, autorizacion de produccion, RIDE y documentos factura/nota credito/nota debito/retencion/guia remision, bajo licencia independiente `facturacion_ecuador`.

- 2026-05-19: `Facturacion electronica por pais` mantiene el submenu de facturacion como contenedor y muestra paginas internas segun pais detectado: Colombia/DIAN, Ecuador/SRI o Panamá/DGI. Ecuador y Panama quedan bajo licencias independientes `facturacion_ecuador` y `facturacion_panama`.

- 2026-05-19: `Facturacion electronica Panama` agrega pagina independiente `Panamá / DGI` y endpoint `/api/empresa/facturacion_electronica/panama`; el perfil PA usa SFEP/DGI con modalidad PAC o Facturador Gratuito, declaracion jurada e-Tax2.0, firma electronica, RUC/DV, CAFE/CUFE/QR y documentos factura/nota credito/nota debito, sin mezclar DIAN Colombia ni agregar tablas/dependencias.

- 2026-05-19: `Caja y turno` quedan independientes por usuario dentro de la misma empresa: apertura manual/automatica, listado de cajas del carrito, pagos, abonos y movimientos validan `usuario_creador`; `empresa_cierres_caja` actualiza su unicidad a empresa/sucursal/caja/fecha/turno/usuario y elimina indices legacy sin usuario. Flujo documentado: el usuario abre/reutiliza caja desde `Corte de Caja`, o el sistema abre una automaticamente al primer cobro si no tiene caja abierta; al finalizar genera/revisa reporte, imprime si aplica y cierra turno/caja. El tablero de estaciones sigue compartido por empresa y refresca cada 3 segundos para varias cajas.

- 2026-05-19: `Documentos imprimibles` deja facturas, ventas, notas, documentos y reportes imprimibles en blanco y negro, independientes del tema claro/oscuro; ventas y facturas esperan la configuracion avanzada para respetar desde la primera vista previa los logos de empresa/sistema en escala de grises.

- 2026-05-19: `Reporte de turno POS` evita que los importes monetarios del resumen financiero se partan en varias lineas en vista previa e impresion POS 80mm.

- 2026-05-19: `Carritos y corte de caja` regulariza pago/anulacion en transacciones PostgreSQL y guarda `descuento_total` al pagar, para que el reporte de turno refleje descuentos y permita anular ventas durante la prueba operativa de Motel Calipso.

- 2026-05-19: `Login de usuarios operativos` corrige el alta del acceso empresarial compartido al iniciar sesion, evitando el error PostgreSQL de columnas/valores en `UpsertAdminEmpresaCompartidaAcceso`.

- 2026-05-19: `Reporte de turno` queda como informe operativo profesional, compacto y completo; ventas y facturas electronicas conservan el estilo visual de factura electronica en POS y carta.

- 2026-05-19: `Login` ajusta el titulo `Powerful Control System` para verse mas compacto, cuadrado y siempre en una sola fila.

- 2026-05-19: `Login` agrega boton `Instalar app` con soporte PWA, manifest, service worker e iconos para escritorio y celular.

- 2026-05-19: `Login` aumenta el tamaño del titulo `Powerful Control System` y le deja una regla responsive propia.

- 2026-05-19: `Corte de caja` unifica el color de los botones de `Lectura rapida` para que se vean mas claros que el panel y mantengan estados visuales consistentes.

- 2026-05-19: `Configuracion de empresa` agrega control para activar/desactivar cajas simultaneas y limite interno por empresa; la apertura manual/automatica respeta configuracion y licencia, y `Ver reporte de mi turno` conserva caja/turno/cierre para reportes independientes por caja.

- 2026-05-19: `Impresoras` deja `POS_80MM` como impresora predeterminada global por empresa activa, asignada a caja, reporte de turno, corte de caja, general y cajon monedero; las empresas nuevas intentan crear este default al registrarse.

- 2026-05-18: `Reporte de turno` prioriza `Ticket POS 80mm` como formato predeterminado, bloquea carta/grande cuando la empresa esta configurada en POS, hace que reportes historicos lean la configuracion guardada antes de imprimir y agrega la utilidad `backend/tools/set_pos80_config` para activar impresora POS por empresa.

- 2026-05-18: `Corte de caja` corrige contraste del bloque `Efectivo esperado en caja` en modo oscuro para evitar fondo blanco con texto blanco.

- 2026-05-18: `Reporte de turno` corrige contraste en modo oscuro para que titulos, valores, etiquetas y tablas sean legibles en reportes actuales e historicos.

- 2026-05-18: `Corte de caja` agrega `Cerrar turno e imprimir reporte`, que guarda el cierre, imprime el reporte y cierra la sesion del usuario al finalizar la impresion.

- 2026-05-18: `Reporte de turno` agrega la metrica configurable `Total descuentos`, calculada desde ventas cerradas con descuento y visible en reporte actual, historicos y exportaciones.

- 2026-05-18: `Reporte de turno` agrega la metrica configurable `Cantidad de ventas`, visible en el reporte actual, reportes historicos y exportaciones del turno.

- 2026-05-18: `Corte de caja` corrige la vista previa de `Ver reporte de mi turno` para respetar el tema claro u oscuro activo en pantalla, manteniendo impresion blanca y legible en carta, ejecutivo y POS. Es un cambio visual sin impacto en backend, permisos, endpoints ni tablas.

- 2026-05-18: `Reporte de turno` se adapta a papel grande y ticket POS 80mm. En grande conserva tabla completa; en POS compacta encabezado, resumen y detalle de ventas como bloques plantillas con etiquetas para evitar desbordes al imprimir en rollo.

- 2026-05-18: `Reportes de turnos` agrega una pagina historica en el submenu de reportes para listar turnos/cortes antiguos, previsualizar el reporte imprimible, imprimir, compartir, exportar en `json/csv/txt/xls/pdf` y enviarlo por email. El backend reutiliza `/api/empresa/corte_caja` con acciones historicas por `cierre_caja_id` y agrega el permiso de pagina `linkReportesTurnos` bajo el modulo `reportes`.

- 2026-05-18: `Reporte de turno` queda configurable para corte de caja. El reporte imprime primero datos de empresa, fecha/hora, usuario y consecutivo; luego muestra ventas ordenadas por fecha/hora con columnas activables de entrada, salida, numero de venta, estacion, cajero, medio y total; finalmente presenta ingresos, egresos, productos, servicios, efectivo, tarjetas, otros medios y efectivo esperado restando egresos. La configuracion vive por `empresa_id` en `empresa_corte_caja_configuracion` y se administra desde `Configuracion de empresa > Reporte de corte`.

- 2026-05-18: `Super Administrador` simplifica el panel inicial como tablero ejecutivo: conserva KPIs de infraestructura, PostgreSQL, alertas, empresas, licencias y continuidad, y retira bloques repetidos de negocio, servicios y costos.

- 2026-05-18: `Login` de administradores muestra `Powerful Control System` como titulo superior, mueve `Acceso de administradores` sobre la imagen lateral y coloca `Ir al inicio` debajo de la tarjeta de acceso.

- 2026-05-18: `Configuracion de empresa` corrige el guardado de formato monetario y numerico en `/api/empresa/configuracion_avanzada`; el backend regulariza flags legacy booleanos de `empresa_configuracion_avanzada` a enteros `0/1`, usa fechas runtime compatibles con PostgreSQL y repara el historial de configuracion operativa con `RETURNING id`.

- 2026-05-18: `Panel` de Administrar empresa retira la tarjeta `Mercado en contexto` y deja de consultar indicadores externos de mercado/cripto desde el tablero.

- 2026-05-18: `Administrar empresa` agrega el acceso `Caja` debajo de `Estaciones` en `Operacion y ventas`, abriendo el mismo corte de caja que la tarjeta Caja de estaciones.

- 2026-05-18: `Corte de caja` agrega `Ver reporte de mi turno` en la pantalla abierta desde la estacion Caja; consulta la caja abierta del usuario autenticado y permite imprimir sin cerrar turno.

- 2026-05-17: `Venta directa` termina de compartir el mismo carrito operativo de estaciones: la tarjeta de acciones y abonos ya no queda oculta por depender solo de `stationMode`.

- 2026-05-17: `Menu flotante` retira `Compartir por WhatsApp` de Utilidades y deja `Cambiar apariencia` penultimo, con `Cerrar sesion` como ultimo acceso.

- 2026-05-17: `Administrar empresa` mueve el grupo `Inventario y compras` para quedar inmediatamente debajo de `Operación y ventas` en el submenu principal.

- 2026-05-17: `Carrito de compras` agrega abonos operativos por estacion: el boton `Abonos` registra dinero recibido, lo lista en la cuenta y `Pagar y cerrar carrito` descuenta esos abonos del saldo final sin mezclarlos con devoluciones.

- 2026-05-17: `Venta directa` usa la misma vista operativa del carrito de estaciones para items, totales, lector y controles de pago, conservando su carrito canonico `VENTA-DIRECTA-{empresa_id}-0`.

- 2026-05-17: `Panel` de Administrar empresa mejora los favoritos de `Accesos rapidos` para que se vean como botones de accion, con borde, fondo, cursor y estados hover/focus claros.

- 2026-05-17: Super Administrador agrupa mensajes y configuraciones de comunicacion en un solo modulo `Comunicaciones`: mantenimiento, mensajes masivos, alertas sistema, Gmail SMTP, alertas de licencia y WhatsApp portal.

- 2026-05-17: `Mantenimiento del sistema` queda como grupo propio del menu principal de Super Administrador y agrega accion CRUD para eliminar alertas viejas desactivadas o vencidas.

- 2026-05-17: `Configuracion` de Super Administrador se separa en paginas independientes por seccion bajo `/super/configuracion/`; cada pagina carga solo su tarjeta operativa y conserva el proceso de guardado existente.

- 2026-05-17: Super Administrador mueve `Mantenimiento del sistema` a `/super/mantenimiento_sistema.html`; mantiene tabla de avisos programados con acciones para cargar, desactivar o eliminar sin tocar el bloqueo real del sistema.

- 2026-05-17: `rs` y `sync_to_vps` eliminan el tope corto del healthcheck remoto y pasan a esperar 900 segundos por defecto en reinicio y Docker.

- 2026-05-17: `Administrar empresa` cambia el icono de `Venta directa` por un simbolo `$` de texto que hereda el color de la apariencia activa.

- 2026-05-17: `Administrar empresa` elimina la pestaña lateral `Inicio` y deja `Panel` como primer boton directo del submenu empresarial; `linkInicio` sale del catalogo de permisos.

- 2026-05-17: `Carrito de compras` deja el preset simplificado como default para tipos de empresa y lo aplica tambien a empresas antiguas al arrancar: búsqueda, lector, items, totales, acciones, valores por medio de pago y pagar activos; cobro avanzado, descuentos, propina, comisión, desglose y lavador apagados por defecto.

- 2026-05-17: `Carrito de compras` mueve `Buscar Productos` junto al boton `Agregar` y la pagina por botones deja la barra `Buscar producto` para filtrar por nombre.

- 2026-05-17: `Carrito de compras` oculta el encabezado `Carrito de compras` y el texto descriptivo de venta directa en modo operativo, dejando la pantalla iniciar directamente con el carrito.

- 2026-05-17: `Carrito de compras` agrega `Efectivo recibido` en los valores por medio de pago y calcula `Cambio a devolver` o `Falta recibir` sin cambiar el payload de pago.

- 2026-05-17: `Carrito de compras` oculta por defecto las secciones `Cobro y estados del carrito` y `Lavador`; ahora se muestran solo con los checks `Mostrar opciones de cobro` y `Mostrar lavador`. El pago principal queda disponible y se limpian valores avanzados invisibles.

- 2026-05-17: `Administrar empresa` vuelve a abrir `Panel` como pagina inicial del iframe al entrar al shell empresarial. El orden del menu se mantiene; no cambian permisos, rutas ni endpoints.

- 2026-05-17: `Administrar empresa` deja `Operacion y ventas` solo con `Venta directa` y `Estaciones`. `Venta publica`, `Red social empresarial`, `Codigos de descuento` y `Chat y tareas` pasan a `Canales digitales y colaboracion`; `Reservas` queda en `Soluciones por negocio` y `Punto de venta / TPV` queda como permiso base. El catalogo de permisos queda alineado sin cambiar rutas, endpoints ni tablas.

- 2026-05-17: `Facturacion electronica` revisa el flujo por pais y DIAN Colombia. El perfil de Colombia deja de apuntar al modulo ERP/documental antiguo, las pruebas DIAN usan como base 60 facturas, 20 notas debito y 20 notas credito para software propio/proveedor tecnologico, y la documentacion separa configuracion por pais, pruebas DIAN, correo automatico y la brecha de adjuntos XML/PDF certificados.

- 2026-05-17: `Estaciones` agrega el check `Primer clic solo activa` en Configuracion de estaciones. Cuando esta activo, el primer clic sobre una estacion disponible solo activa su carrito base y actualiza la tarjeta a ocupada; el segundo clic entra al carrito. Se conserva compatibilidad con `abrir_carrito_al_activar=false`.

- 2026-05-17: `Reportes` agrega filtros operativos al `Reporte de turno y caja`. El centro de reportes ahora envia `usuario`, `caja_codigo`, `turno` y `cierre_id` a `/api/empresa/reportes`, se protege que `reporte_de_turno` permanezca en el catalogo y se verifica visualmente el corte automatico, cierre/impresion de turno y ultimos movimientos de caja actual.

- 2026-05-15: `Venta directa` repara el boton `Buscar (botones)`. El clic ahora prepara o recupera el carrito de venta directa antes de abrir el catalogo por botones, mantiene el retorno y las APIs de carrito/items en modo `venta_directa`, y el backend responde un carrito puntual cuando `/api/empresa/carritos_compra` recibe `id`, evitando que el buscador quede sin contexto o permiso equivocado.

- 2026-05-15: `Administrar empresa` deja `Venta directa` como primer acceso real del menu y `Estaciones` como segundo. El bloque `Operacion y ventas` ahora abre al inicio, el arranque del iframe prefiere `Venta directa` cuando el permiso esta visible, y `Carritos` queda catalogado solo desde `Configuracion > Ventas y cobro`.

- 2026-05-15: `Asistencia de empleados` queda reparada para PostgreSQL y mejora su operacion diaria. Las consultas y escrituras runtime usan helpers compatibles (`queryRowSQLCompat`, `querySQLCompat`, `execSQLCompat`, `insertSQLCompat`) y `sqlNowExpr()`, evitando fallos en guardar configuracion, crear registros, marcar entrada/salida, activar/desactivar y eliminar. El boton `Agregar registro` y `Editar` abren la seccion Registro aunque el submenu este en otra pestana; `Entrada ahora` y `Salida ahora` marcan la hora actual sin prompts nativos. Validado con `go test ./db ./handlers` y prueba visual Playwright de crear, entrada, salida, editar, configurar, filtrar, cerrar periodo y descargar reporte.

- 2026-05-15: `Reportes` retira el acceso IA del Centro de reportes. El submenu queda solo con `Centro de reportes`, `reportes_ejecutivos.html` deja de mostrar boton/tarjeta `Asistente IA`, se elimina `web/administrar_empresa/reportes_ia_chat.html` y el permiso de pagina `linkReportesIAChat`. El endpoint tecnico `/api/empresa/reportes_ia_chat` se conserva para el asistente global en modo reportes bajo permiso general de reportes.

- 2026-05-15: `Facturacion electronica` retira del submenu el acceso `Volver a empresas`, dejando el menu enfocado en configuracion, pruebas DIAN, proveedores de firma, facturas electronicas y AIU construccion. Cambio solo frontend; no modifica permisos, APIs ni tablas.

- 2026-05-15: `Creditos` corrige el boton `Nuevo credito` en PostgreSQL. El alta deja de mezclar `tx.Exec` directo, placeholders `?` y `datetime('now','localtime')` en la transaccion que crea el credito y sus cuotas; ahora usa helpers compatibles (`insertTxSQLCompat`/`execTxSQLCompat`) y `sqlNowExpr()`. Validado con pruebas enfocadas y `go test ./db ./handlers`.

- 2026-05-15: `Productos` ahora avisa cuando falta bodega antes de crear un producto. Si la empresa no tiene bodegas activas, `Nuevo producto` muestra un mensaje claro indicando que primero debe crear una bodega y ofrece el boton `Crear bodega`; el guardado tambien bloquea creaciones nuevas sin bodega activa.

- 2026-05-15: `Compras` ahora obliga a elegir proveedores creados en el catalogo empresarial para cotizaciones y recepciones avanzadas. Los selectores se cargan desde `/api/empresa/proveedores`, las cotizaciones/recepciones envian `proveedor_id`, el backend valida que sea un proveedor activo de la misma empresa y `empresa_compras_recepciones_avanzadas` conserva la referencia. La compra basica agrega acceso a gestionar proveedores y autocompleta el documento del proveedor seleccionado.

- 2026-05-15: `Menu empresarial` reordena `Operacion y ventas`: `Venta directa` queda como primera opcion operativa, `Carritos` sale de ese grupo principal y se conserva dentro de `Configuracion > Ventas y cobro`. Cambio solo de navegacion; no modifica permisos, rutas, endpoints ni tablas. Validado con Playwright revisando el orden real del menu y el submenu de configuracion.

- 2026-05-15: `Bodegas` corrige el alta real en PostgreSQL. `CreateBodega` deja de usar `datetime('now','localtime')` en el `INSERT` y usa `CURRENT_TIMESTAMP` via `sqlNowExpr()`, evitando fallo al guardar desde `Administrar Bodegas`. Validado con prueba Go enfocada y prueba visual Playwright escritorio/movil: `Nueva` -> llenar -> `Guardar bodega` -> mensaje de exito, KPI 1 y fila visible.

- 2026-05-15: `Login y registro movil` refuerza la experiencia tactil en `login.html`, `login_usuario.html` y `registrar_nuevo_usuario_administrador.html`. Los botones, enlaces de recuperacion/registro, inputs y boton de mostrar contrasena quedan con areas de toque de al menos 44 px; el chat flotante del registro administrador se compacta en celular para no tapar los campos. Validado visualmente con Playwright movil en login administrador, login operativo, recuperacion de invitacion, registro por invitacion y registro administrador.

- 2026-05-15: `Venta directa` carga siempre un carrito operativo aunque la URL no traiga `carrito_codigo`. La pantalla usa el carrito canonico `VENTA-DIRECTA-{empresa_id}-0`, reutiliza el legacy `VENTA-DIRECTA-{empresa_id}`/`CAJA_DIRECTA` si existe y oculta el retorno a estaciones en ese modo. Validado con parseo del script inline y prueba visual Playwright creando `VENTA-DIRECTA-32-0`.

- 2026-05-15: `Empresas compartidas` ahora permite definir alcance profesional al invitar: `Solo ver`, `Acceso total` o `Solo ciertos modulos`. El alcance se guarda en invitaciones/accesos, se aplica al contexto efectivo de permisos y paginas protegidas, se conserva al reenviar/aceptar invitaciones y puede revocarse con la accion visible `Dejar de compartir`. Validado con pruebas Go de `handlers`/`db`, parseo JS y verificacion visual del formulario.

- 2026-05-15: `Licencias` blinda la prueba de 15 dias para que cada empresa pueda usar una sola activacion gratis/prueba activa en toda su vida. `licencias_activaciones_gratis` agrega `asesor_id`, normaliza duplicados activos historicos por empresa y crea indice unico activo por `empresa_id`; el checkout y la activacion sin pago validan codigos de asesor aceptados antes de activar. La prueba de 15 dias queda en 250 documentos/ventas mensuales y conserva trazabilidad comercial del asesor.

- 2026-05-14: `Administrar empresa` reduce wrappers y consolida el nucleo de pantallas. Venta directa abre directamente `carrito_de_compras.html?modo=venta_directa&perm_page=linkVentaDirecta`; OnlyOffice queda solo en `documentos_onlyoffice.html`; CRM avanzado vive en `crm_comercial.html` dentro del submenu Forecast; Productos usa `administrar_productos.html?view=...` para categorias/proveedores/precios/compras. Se eliminan paginas HTML que solo redirigian y no tenian logica propia.

- 2026-05-14: `Reportes` queda como centro unico preproduccion. El submenu deja solo `Centro de reportes` y `Asistente IA`; `reportes_ejecutivos.html` concentra catalogo, vista previa y exportacion desde `/api/empresa/reportes`. Se eliminan las paginas antiguas `reportes.html`, `reportes_inventario.html`, `reportes_finanzas.html`, `graficos_estadisticas.html` y la ruta `/api/empresa/graficos_estadisticas`. Tambien se limpian datos viejos de QA/demo en PostgreSQL, evidencias locales antiguas, el runner de finanzas QA sobre Motel Calipso y codigos `DEMO-*` de plantillas, sustituidos por `BASE-*`.

- 2026-05-13: `Carrito de compras` imprime siempre el recibo operativo al cerrar carrito, aunque la configuracion general de factura/venta tenga `imprimir_venta=false`. El recibo de caja se puede desactivar solo con una bandera especifica futura `desactivar_impresion_carrito=true`, evitando que el flujo de pago parezca incompleto cuando la venta si quedo registrada.

- 2026-05-13: `Carrito de compras` repara el caso real donde `Pagar y cerrar carrito` parecia no hacer nada cuando no habia caja abierta. Al presionar pagar, el frontend crea una caja de cobro automatica mediante `action=abrir_caja_cobro`, la selecciona y continua el pago con recibo/factura; el backend crea o reutiliza `CAJA-1` abierta sin saltarse la validacion final de `cierre_caja_id`. Validado con `go test ./handlers ./db`.

- 2026-05-13: `Carrito de compras` blinda el pago con impresion automatica. El flujo reserva las ventanas de recibo/factura en el mismo clic de `Pagar y cerrar carrito`, evitando bloqueos del navegador cuando el cobro termina despues de llamadas async. Si una ventana emergente no esta disponible, usa impresion en iframe como fallback. Validado visualmente en modo estacion: clic real, `PUT action=pagar_estacion`, regreso a estaciones, recibo POS impreso; y prueba adicional con factura electronica activada imprimiendo recibo + factura.

- 2026-05-13: `Carrito de compras` hace que `Pagar y cerrar carrito` ejecute el cobro con un clic directo, sin depender de un `confirm()` nativo que podia quedar oculto en el panel. El mensaje de pago exitoso queda persistente despues del refresco del carrito y el panel de Administrar empresa deja de llamar a `ipwho.is`, usando proveedores IP alternos con fallback. Validado con parseo JS y prueba visual Playwright mock: clic real en el boton, sin dialogo nativo, `PUT action=pagar_estacion` con `cierre_caja_id=11`, `caja_codigo=CAJA-1`, `total_pagado=15000` y aviso visible de venta/caja.

- 2026-05-13: `Carrito de compras` refuerza definitivamente el boton `Pagar y cerrar carrito`: `pagar_estacion` pasa a permiso operativo de crear/registrar venta, el boton ya no queda bloqueado si la carga inicial de cajas falla y al clic recarga cajas abiertas antes de cobrar. Validado con `go test ./handlers ./db` y prueba visual Playwright simulando 403 inicial de cajas, reintento por clic y pago exitoso con `cierre_caja_id`.

- 2026-05-13: `Carrito de compras` repara el boton `Pagar y cerrar carrito` cuando el usuario puede vender pero no tiene acceso al modulo Finanzas. La carga de cajas abiertas pasa a `/api/empresa/carritos_compra?action=cajas_abiertas`, bajo permisos de Ventas/Estaciones/Venta directa, y deja de depender de `/api/empresa/finanzas/cierres_caja`. El pago sigue enviando y validando `cierre_caja_id`, `caja_codigo`, turno y `empresa_id` antes de cerrar la venta. Validado con `go test ./handlers ./db` desde `backend` y prueba visual Playwright mock del flujo de pago.

- 2026-05-13: `Carrito de compras` refuerza el modo plano con un override absoluto al final de `web/estilos.css`, para que ningun tema, modo tactil o estilo legacy vuelva a aplicar sombras, radios o margenes a las tarjetas. Validado con Playwright revisando 15 tarjetas/formularios con `box-shadow: none`, `border-radius: 0px` y margenes `0px`.

- 2026-05-13: `Panel de Administrar empresa` prioriza deteccion de ciudad por IP para clima e indicadores. `web/administrar_empresa/panel.html` usa `ipwho.is` antes que GPS guardado o GPS del navegador; la ubicacion manual guardada sigue teniendo prioridad. Validado con parseo JS y prueba visual Playwright mock confirmando ciudad por IP sin consultar GPS.

- 2026-05-13: `Carrito de compras` compacta las tarjetas para que queden pegadas entre si y con apariencia de tabla plana. `web/estilos.css` elimina gaps, margenes, radios y sombras en los contenedores del carrito bajo `carrito-flat-page`, manteniendo el cambio limitado a esa pantalla. Validado con Playwright comprobando `gap: 0`, `margin: 0`, `border-radius: 0` y `box-shadow: none`.

- 2026-05-13: `Carrito de compras` queda visualmente plano en sus tarjetas. `web/administrar_empresa/carrito_de_compras.html` agrega una clase de pagina y `web/estilos.css` elimina sombras de `.card`/`.form` y tarjetas internas solo dentro del carrito. Validado con parseo JS y prueba visual Playwright comprobando `box-shadow: none` en tarjetas. Sin backend, permisos, tablas ni dependencias nuevas.

- 2026-05-13: `Carrito de compras` en modo estacion mueve el nombre de la estacion fuera de la tarjeta de items. El encabezado superior muestra `Estacion: <nombre>` debajo de `Carrito de compras`, mientras la tarjeta interna conserva solo `Items del carrito`. Cambio solo frontend en `web/administrar_empresa/carrito_de_compras.html`; sin backend, permisos, tablas ni dependencias nuevas.

- 2026-05-13: `Documentos OnlyOffice` corrige la apertura del editor al crear documentos. El backend evita duplicar `/empresas/empresas` en la ruta temporal y reescribe para el navegador URLs internas de Docker como `http://onlyoffice-documentserver:80` hacia la URL publica esperada (`https://onlyoffice.<dominio>`), con override opcional `ONLYOFFICE_PUBLIC_DOCUMENT_SERVER_URL`. El frontend valida que `api.js` exponga `DocsAPI.DocEditor` y muestra un error claro si el Document Server no es accesible. Validado con pruebas Go enfocadas y prueba visual Playwright con mock de OnlyOffice.

- 2026-05-13: `Reportes ejecutivos` simplifica el submenu empresarial y repara botones de reportes financieros. `reportes_menu.html` queda con entradas profesionales de alto nivel, `reportes_ejecutivos.html` funciona como centro de mando con accesos agrupados por direccion, ventas, inventario, finanzas, fiscal e IA, y `reportes_finanzas.html` deja placeholders para cargar/exportar movimientos financieros reales desde `/api/empresa/reportes`. Sin backend, tablas ni dependencias nuevas.

- 2026-05-13: `Facturacion electronica` separa configuracion y pruebas DIAN. La pagina principal elimina el boton `Abrir modulo DIAN / documental`, queda enfocada en datos/firma/configuracion, y la nueva subpagina `facturacion_electronica_pruebas_dian.html` agrupa diagnostico DIAN, set de habilitacion, conexion/cola y emision documental manual. Sin backend, tablas, permisos ni dependencias nuevas.

- 2026-05-13: `Estaciones` agrega control de aseo por usuario. `administrar_usuarios.html` permite activar `Control de aseo`; cuando una estacion esta sucia, el usuario habilitado reporta el aseo terminado con un clic y el backend calcula la demora en `empresa_estacion_aseo_eventos`. Nuevo reporte `reporte_aseo_estaciones.html` con tiempos por habitacion/aseadora y resumen. Validado con pruebas Go dirigidas y parseo JS.

- 2026-05-13: `Caja` agrega `Corte automatico` en `web/administrar_empresa/corte_de_caja.html`. El corte toma la caja abierta del usuario actual, usa `fecha_apertura` como inicio del turno y la hora actual como cierre, autocompletando usuario/caja/turno/apertura sin escoger fechas. Al guardar con `cierre_caja_id`, el backend cierra la caja abierta existente en lugar de duplicar el cierre. Validado con pruebas Go, parseo JS, `git diff --check` y prueba visual Playwright mock.

- 2026-05-13: `Caja` limita `Ver ultimos movimientos` al usuario actual y a su caja abierta actual. `backend/handlers/corte_caja.go` agrega el modo `mi_caja_actual`, filtra por `cierre_caja_id`, `caja_codigo` y usuario autenticado; `estaciones.html` envia el alcance seguro y `ultimos_movimientos_de_caja.html` deja de consumir listados globales de inventario/facturacion. Validado con pruebas Go de `handlers`/`db`, parseo JS y `git diff --check`.

- 2026-05-13: `Carritos` y `Venta directa` ajustan la factura electronica automatica para Colombia produccion. La venta queda cerrada como comprobante, pero la factura electronica asociada solo conserva `estado_documento=emitida` si la integracion fiscal confirma `estado_envio=enviado`; si DIAN/proveedor falla queda `pendiente_emision` con observacion y cola de reintentos. Ademas, Colombia produccion ya no acepta proveedor `manual/local/interno` como envio fiscal exitoso. Se agrega prueba enfocada en `backend/handlers/facturacion_documentos_electronicos_test.go`. Sin tablas, permisos ni dependencias nuevas.

- 2026-05-13: se crean y activan logos empresariales para Motel Calipso (`empresa_id=7`) y Gimnasio el bollon (`empresa_id=32`). Los activos quedan en `web/uploads/empresa_logos/empresa_7/motel-calipso-logo.svg` y `web/uploads/empresa_logos/empresa_32/gimnasio-el-bollon-logo.svg`; la configuracion `empresa_configuracion_avanzada` queda con `logo_url`, `mostrar_logo=1` y `mostrar_logo_empresa=1` para que aparezcan en panel, carrito, facturas, recibos y reportes que consumen el logo empresarial.

- 2026-05-13: `web/administrar_empresa/panel.html` mejora los indicadores economicos importantes en movil. Las tarjetas dejan de comprimirse a 4 columnas, muestran titulo, alcance, indicador, valor, referencia y variacion sin recortes ni `ellipsis`, y permiten scroll vertical cuando el alto del telefono no alcanza. Cambio solo frontend; no modifica APIs, permisos, tablas ni dependencias.

- 2026-05-13: `Facturacion electronica` agrega acceso a proveedores de firma digital. En `web/administrar_empresa/facturacion_electronica.html`, la tarjeta `Cargar firma electronica (Colombia / DIAN)` suma el boton `Adquirir Firma Electronica`; el boton abre `web/administrar_empresa/proveedores_firma_digital.html` dentro del submenu de facturacion. La nueva pagina publica a Sensiyo como proveedor externo para comprar certificado digital/firma DIAN, enlazando a `https://sensiyo.co/certificados-digitales/`.

- 2026-05-13: `web/administrar_empresa/estaciones.html` recupera la tarjeta compacta de `Caja`: titulo, totales y boton `Ver ultimos movimientos`, con color configurado de estaciones. El clic o teclado sobre la tarjeta completa sigue abriendo `corte_de_caja.html`; el boton interno abre `ultimos_movimientos_de_caja.html`.

- 2026-05-13: `Carritos` corrige el 500 al abrir estaciones cuando una base empresarial tenia migraciones rezagadas. `backend/db/carritos_compras.go` valida y completa todas las columnas usadas por el listado/items antes de marcar el esquema como listo, cachea por base/esquema PostgreSQL y, ante `column/relation does not exist`, refresca el esquema y reintenta `/api/empresa/carritos_compra`. Tambien corrige el reintento sin conteo de items para no referenciar el alias `ic` cuando el join fue desactivado. Se agrega `backend/db/carritos_compras_schema_test.go`; validado con pruebas enfocadas de `db` y `handlers`, mas smoke visual controlado de Estaciones -> Zona 1 -> carrito.

- 2026-05-13: Juegos queda adaptado a movil con sonido y records globales. `backend/main.go` registra `super_juegos_records` y `/api/public/juegos/records`; `web/Juegos/juegos_records.js` centraliza rankings; los wrappers arcade reportan puntajes, agregan sonido WebAudio y muestran panel de records; `menu_juegos.html` usa tarjetas uniformes con capturas reales PNG en `web/img/juegos/`; `/emulador/` suma controles tactiles y `/Juegos/n64/index.html` embebe el emulador real. Validado con `go test ./...`, `node --check` y prueba visual Playwright movil/escritorio.

- 2026-05-13: `web/administrar_empresa/estaciones.html` blinda la tarjeta especial `Caja` para que nunca abra carrito. Ahora clic o teclado sobre `Caja` siempre llevan a `web/administrar_empresa/corte_de_caja.html` con el flujo de cierre de turno, corte de caja e impresion del reporte del usuario actual.

- 2026-05-13: `web/administrar_empresa/carrito_de_compras.html` recupera de nuevo carritos legado por nombre al abrir una estación. La carga inicial ya no se cierra al filtro `estacion_id`, y si encuentra un carrito histórico de la estación lo normaliza al `codigo` y `referencia_externa` canónicos antes de activarlo o reanudarlo.

- 2026-05-13: `administrar_empresa.html` ahora valida sesion administrativa con `/me` antes de abrir el shell empresarial; si la sesion no existe o expiro, redirige a `login.html` en lugar de dejar el menu y el iframe en un estado falso de acceso.

- 2026-05-13: `web/administrar_empresa/carrito_de_compras.html`, `web/administrar_empresa/administrar_clientes.html` y `web/administrar_empresa/bodega.html` ya no muestran errores crudos de `unauthorized` o `Error cargando ...` cuando la sesion expiro. Un `401` redirige al login y un `403` informa que el rol no tiene permiso para ese modulo.

- 2026-05-13: `login.html` y `login_usuario.html` pasan a usar imagenes reales (`web/img/login-admin-real.png` y `web/img/login-usuario-real.png`) con un iluminado exterior suave adaptado a modo claro/oscuro para mejorar presencia visual sin perder limpieza.

- 2026-05-13: en `Estaciones`, la tarjeta especial `Caja` ya no abre `venta_directa` ni un carrito. Ahora redirige a `web/administrar_empresa/corte_de_caja.html` para cerrar turno, generar corte e imprimir el reporte del cajero, con regreso a estaciones y filtro por `caja_codigo` configurado por empresa.

- 2026-05-13: `Carritos` y `Venta directa` quedan reforzados para volver a cargar desde `administrar_empresa.html` aun si la base empresarial venia con migraciones atrasadas. El backend reasegura el esquema antes de listar y la consulta de carritos degrada sin joins opcionales cuando faltan `clientes` o `carrito_compra_items`; el frontend de `carrito_de_compras.html` encapsula toda la carga inicial en el mismo manejo de error visual para no dejar el iframe muerto.

- 2026-05-13: se crea el diagrama entidad relacion vigente del proyecto en `documentos/diagramas/diagrama_entidad_relacion.md` y su imagen `documentos/diagramas/diagrama_entidad_relacion.svg`. Tambien se limpia el indice actual de diagramas para quitar referencias vigentes a artefactos historicos que ya no existen fisicamente.

- 2026-05-13: documentacion general y ayuda actualizadas. Se agrega `documentos/estado_documentacion_2026-05-13.md`, se alinean `documentos/README.md` y `documentos/descripcion_del_proyecto`, y `web/ayuda/ayuda.html` incorpora acceso global de usuarios, operacion conectada, cajas simultaneas, soporte/comunicaciones, documentos locales y backups.

- 2026-05-13: revision profesional del nucleo operativo. Auditoria reconoce modulos recientes (carritos, venta publica, CRM, reportes, backups, OnlyOffice, tickets, mantenimiento, propinas/comisiones y plantillas), y ventas/facturas/ingresos/egresos/reportes/menu flotante agregan compartir por WhatsApp o correo. Validado con `go test ./...`, `node --check` de JS y parseo de scripts inline HTML.

- 2026-05-13: `Backup profesional` se mueve en `web/administrar_empresa.html` al grupo `Administracion`, junto al boton `Configuracion`, conservando `linkBackups`, ruta y permisos existentes. Sin cambios de backend, tablas ni dependencias.

- 2026-05-13: `login_usuario.html` incorpora una ilustracion profesional de usuario frente al computador en la esquina superior derecha, con adaptacion responsive para movil. Nuevo activo local `web/img/login-usuario-computador.svg`; sin cambios de backend, permisos, tablas ni dependencias.

- 2026-05-13: `login_usuario.html` queda como acceso global para usuarios operativos de todas las empresas. El backend resuelve la empresa por email y clave, valida el primer ingreso por `token_invitacion` sin depender de subdominio, y redirige a `administrar_empresa.html?id=...` con rol/permisos de la empresa. Sin tablas ni dependencias nuevas; validado con `go test ./...`.

- 2026-05-13: agregado aviso de mantenimiento programado. Super administrador configura check de aviso, fecha, hora inicio/fin, zona horaria y mensaje publico; el panel de Administrar empresa consulta `/api/empresa/mantenimiento_programado` y muestra una franja visible cuando el aviso esta activo, separada del bloqueo real `mantenimiento_activo`.

- 2026-05-13: OnlyOffice empresarial queda en una sola pantalla local editable. `web/administrar_empresa/documentos_onlyoffice.html` permite crear Word/Excel/PowerPoint, abrir el editor embebido alli mismo y descargar el resultado al PC/celular; `backend/handlers/onlyoffice.go` agrega `create_edit_local` y `download&delete=1` para borrar la copia temporal del VPS. El menu antiguo redirige a la pantalla unica.

- 2026-05-13: agregado sistema de alertas de vencimiento de licencias por correo. Super administrador configura dias de aviso en `/super/configuracion_avanzada.html`; el backend agrega `/super/api/licencias/vencimiento_alertas`, worker periodico, plantilla `licencia_expiry_warning` y deduplicacion en `licencia_vencimiento_notificaciones`. Validado con `go test ./...`.

- 2026-05-13: reparado el submenu de Configuracion en super administrador para alinearlo con el sidebar de `seleccionar_empresa.html`: titulo simple, ancho estable, lista compacta, estados activo/hover coherentes y autocierre movil al tocar secciones. Sin cambios de backend, permisos ni base de datos.

- 2026-05-13: corregida la apariencia claro/oscuro del Explorador de Archivos del super administrador. La pagina `/super/explorador_archivos.html` inicializa el tema desde `localStorage`/cookie `pcs_theme` y reemplaza colores fijos por variables globales para fondos, tabla, botones, input y estados. Sin cambios de backend, permisos ni base de datos.

- 2026-05-12: profesionalizado el sistema propio de tickets de ayuda. El menu flotante empresarial envia tickets con contacto preferido, telefono opcional, contexto tecnico seguro y tickets recientes; el backend permite detalle/comentarios propios aislados por `empresa_id`; la consola super muestra categoria, modulo, contacto, ruta y diagnostico sin exponer secretos.

- 2026-05-12: SSH del VPS cambiado de `22` a `49222` con procedimiento seguro: primero se habilitaron ambos puertos, se probo conexion externa por `49222`, luego se cerro `22` y se verifico que solo `49222` quedara escuchando. Scripts de despliegue, tuneles locales, RustDesk remoto, escaner VPS y documentacion operativa quedan alineados con `49222`.

- 2026-05-12: `Control electrico` se mueve dentro de Configuracion de empresa. El sidebar principal de Administrar empresa deja un solo acceso a `Configuracion`, y `web/administrar_empresa/configuracion_menu.html` incorpora el enlace en `Estaciones, sensores y tarifas`; `configuracion.html` agrega un acceso directo en el mapa ejecutivo. Sin cambios de backend, tablas ni dependencias.

- 2026-05-12: retirado el modulo legado de mesa de ayuda empresarial. Se eliminan su ruta, permiso, pagina, opcion de licencias y plantillas demo; el flujo oficial queda en tickets propios con `/api/empresa/tickets_ayuda` y `/super/api/tickets_ayuda`.

- 2026-05-12: agregado sistema central de tickets de ayuda. El menu flotante permite crear tickets asociados a la empresa activa y el panel super incorpora `/super/tickets_ayuda.html` con bandeja, filtros, conversacion, respuesta, prioridad, asignacion y cierre. Backend nuevo: `/api/empresa/tickets_ayuda`, `/super/api/tickets_ayuda`, `super_tickets_ayuda` y `super_ticket_ayuda_mensajes`.

- 2026-05-12: el carrusel horizontal del index usa el mismo ancho de tarjeta que la grilla principal: 3 columnas en escritorio, 2 en tablet y 1 en movil, con avance de flechas basado en el gap real.

- 2026-05-12: el paquete Go de almacenamiento del escaner VPS se mueve de `backend/vpssecurity/logs` a `backend/vpssecurity/logstore` para que Docker no lo confunda con carpetas runtime `logs` ignoradas por `.dockerignore`.

- 2026-05-12: `sync_to_vps.ps1` limpia codigo fuente backend obsoleto en el VPS antes de extraer el paquete, preservando `.env`, logs, binarios, `tmp` y `secure`; evita que archivos eliminados localmente, como handlers Nextcloud retirados, queden fantasma y rompan el build Docker.

- 2026-05-12: SSH del VPS se habia restablecido temporalmente al puerto `22` desde Hostinger antes de aplicar el cambio seguro posterior a `49222`.

- 2026-05-12: el index muestra solo las primeras 6 tarjetas de sistemas en la grilla principal y mueve las tarjetas restantes a una barra horizontal navegable con flechas izquierda/derecha antes de la seccion de modulos. Sin backend, tablas, permisos ni dependencias nuevas.

- 2026-05-12: corregido el boton `Ver ultimos movimientos` en la tarjeta de Caja de `estaciones.html`; ahora usa altura minima, centrado flexible y texto en varias lineas para mostrarse completo dentro de la tarjeta. Sin backend, permisos ni tablas.

- 2026-05-12: la landing `/descripcion_de_los_sistemas.html` sincroniza sus tarjetas con la misma foto grande configurada en cada tarjeta del index (`imagen_secundaria_url`), conservando el logo como apoyo visual pequeno. Sin backend, tablas, permisos ni dependencias nuevas.

- 2026-05-12: despliegue VPS preparado para modo 100% Docker. `docker-compose.platform.yml` agrega `edge` Nginx publico y `certbot` con volumenes `pcs_letsencrypt`/`pcs_certbot_www`; se agregan plantillas `deploy/nginx/edge*.conf.template` y scripts `vps-docker-edge-up.sh`/`vps-docker-edge-renew.sh` para mover 80/443, TLS y certificados al stack Docker. Se limpian variables legacy Nextcloud de entornos de plataforma/staging.

- 2026-05-12: agregado Explorador de Archivos para super administrador. `web/super/explorador_archivos.html` muestra una vista tipo explorador por carpetas del filesystem visible para el backend/VPS y `GET /super/api/explorador_archivos` lista raices, rutas y metadata sin leer contenido ni permitir escritura. Acceso exclusivo de `super_administrador`; sin tablas ni dependencias nuevas.

- 2026-05-12: Nextcloud retirado del producto empresarial y del despliegue VPS. Se eliminan rutas `/api/empresa/nextcloud` y `/super/api/config/nextcloud`, pagina empresarial, menu, permisos/licencias, Compose Nextcloud y catalogo de backup de configuracion. El antiguo rango de GB se reconvierte a `empresa.limitaciones.db.max_gb` como cuota maxima de base de datos por empresa, con lectura compatible del valor legacy. Se agrega `deploy/scripts/vps-remove-nextcloud.sh` para apagar contenedores/volumenes legacy en la VPS.

- 2026-05-12: agregada accion `Regresar a estaciones` en `web/administrar_empresa/ultimos_movimientos_de_caja.html`, conservando `empresa_id` al volver al tablero de estaciones. Sin backend, permisos, tablas ni dependencias nuevas.

- 2026-05-12: login de usuarios operativos simplificado. `web/login_usuario.html` elimina controles y enlaces innecesarios del acceso operativo y deja solo recuperacion de contrasena y recuperacion de email de invitacion. Se agrega `POST /api/empresa/usuarios/recuperar_invitacion`, con reCAPTCHA y respuesta enmascarada, para reenviar la invitacion a usuarios ya creados por el administrador que todavia no completaron su contrasena.

- 2026-05-12: registro de usuarios operativos cerrado a invitacion por email. El administrador crea el usuario y el enlace enviado abre `login_usuario.html` con `token_invitacion`; el primer password exige token vigente, documento, contrato y reCAPTCHA, consume la invitacion, confirma correo y abre sesion redirigiendo a `administrar_empresa.html?id=...` con rol cargado para que `/api/empresa/permisos_contexto` filtre el panel. Sin columnas ni dependencias nuevas.

- 2026-05-12: login de usuarios operativos toma primero el tema desde cookies. `web/login_usuario.html` aplica `pcs_theme` antes de cargar estilos y `web/menu.js` prioriza esa cookie sobre `localStorage`, para que el login use la apariencia del ultimo usuario que inicio sesion o cambio tema en ese navegador. Sin cambios de API, permisos, backend, base de datos ni dependencias.

- 2026-05-12: ayuda interna para CRM unificado. `web/administrar_empresa/modulo_menu.html` agrega el boton `Ayuda` al final del menu de CRM y `web/administrar_empresa/crm_comercial.html` incorpora la pestana de ayuda iniciando con la definicion de CRM antes de explicar tablero, leads, seguimientos, cotizaciones, forecast, metas y embudo. Sin cambios de API, permisos, backend, base de datos ni dependencias.

- 2026-05-12: tarjeta de totales del carrito adaptada por tipo de empresa. `web/administrar_empresa/carrito_de_compras.html` usa `/api/empresa/permisos_contexto` y la configuracion general como contexto para renombrar `Totales y detalles`, etiquetas de cliente, cargos, servicios, productos, tiempos, total y saldo segun el vertical activo; `web/estilos.css` agrega el indicador compacto de perfil. Sin cambios de API, permisos, backend, base de datos ni dependencias.

- 2026-05-12: gimnasio elevado a consola empresarial. `web/administrar_empresa/gimnasio.html`, `web/js/gimnasio.js`, `web/administrar_empresa/gimnasio_menu.html` y `web/estilos.css` agregan salud operativa, alertas ejecutivas, acciones rapidas, busqueda/filtro global, estados visuales y menu agrupado por direccion, comercial, operacion deportiva y acceso. Sin cambios de API, permisos, backend, base de datos ni dependencias.

- 2026-05-12: menu de configuracion super alineado con Administrar empresa. `web/super/configuracion_avanzada.html` reemplaza el submenu plano por un sidebar agrupado con clases `admin-sidebar`/`admin-nav-grouped`, buscador, colapso movil e iconos por seccion. Sin cambios de API, permisos, backend, base de datos ni dependencias.

- 2026-05-12: licencias ocultables para clientes. `web/super/licencias.html` renombra el estado como visibilidad comercial y permite mostrar/ocultar licencias del catálogo; el backend bloquea checkout público, Wompi, Nequi, Epayco, activación sin pago y addons seleccionados cuando `licencias.activo=0`. No agrega tablas ni dependencias.

- 2026-05-12: nucleo configurable por plantilla de tipo de empresa. `tipo_empresa_preconfiguraciones.config_json` incorpora `adaptacion_nucleo`; al aplicar la plantilla se guarda la preferencia empresarial y `estaciones_config` declara el recurso que representa cada estacion. Usuarios operativos y productos/servicios quedan como nucleo comun configurable, no como duplicados por vertical. Sin tablas ni dependencias nuevas.

- 2026-05-12: matriz profesional de 30 plantillas canonicos. `/api/*/plantillas_integracion/catalogo` ahora devuelve exactamente 10 clasicos reales + 20 nuevos; `consultorio_odontologico` se fusiona en `odontologia`, `taxi` en `taxi_system` y `turnos_atencion`/`turnos` quedan como soporte transversal. Cada item visible publica readiness profesional, fusiones/soportes, alcance de configuracion y amarre financiero con ingresos/egresos del nucleo (`empresa_finanzas_movimientos`, ventas, pagos, tesoreria y reportes). Sin cambios de esquema ni dependencias.

- 2026-05-12: compactados los indicadores economicos del panel empresarial en escritorio. `web/administrar_empresa/panel.html` cambia las tarjetas grandes de mercado por una tabla de dos indicadores por fila en PC, manteniendo tarjetas compactas en movil. No cambia APIs, permisos ni dependencias.

- 2026-05-12: corregido el enlace `Probar Gratis` del index. `web/index.html` ahora abre `/descripcion_de_los_sistemas.html` con contexto de tarjeta y ancla, `backend/main.go` sirve la ruta legacy `/descripcion_de_los_sistemas.ht` como HTML para evitar descargas, `AuthMiddleware` permite ambas rutas publicas y `web/super/pagina_principal.html` apunta a la landing oficial `.html`.

- 2026-05-12: reparada la apariencia claro/oscuro del Centro de mando super. `web/super/licencias_resumen.html` ahora consume variables globales de tema, aplica temprano el tema guardado, corrige contrastes de botones/pills/tablas/tarjetas y deja graficas SVG y aro de score atados a CSS variables. No cambia endpoints, permisos ni dependencias.

- 2026-05-12: reconstruido desde cero el Centro de mando del super administrador. `web/super/licencias_resumen.html` queda como consola ejecutiva moderna con score operativo, KPIs profesionales de VPS/proyecto, PostgreSQL, seguridad, negocio SaaS, costos IA/hosting, servicios, riesgos, incidentes y accesos de gobierno. Reutiliza APIs existentes, sin dependencias nuevas, tablas nuevas ni cambios de permisos. Verificacion: parseo de script inline con Node y `git diff --check -- web/super/licencias_resumen.html`.
- 2026-05-11: matriz de integracion movida a configuracion empresarial. `web/administrar_empresa.html` retira el acceso de Soluciones por negocio, `web/administrar_empresa/configuracion_menu.html` lo ubica en Base empresarial y `backend/handlers/empresa_permisos.go` mantiene `linkPlantillasIntegracion` con `seguridad:R` bajo Administracion y configuracion.
- 2026-05-11: emisora online configurable por empresa junto al chat/robot. `web/administrar_empresa/configuracion_chat_flotante.html` agrega el check `Activar emisora online`, `web/js/radio_player.js` y `web/js/ai_chat_drawer.js` leen/guardan `radio_online_enabled`, y `/api/chat_flotante/preferencias` persiste preferencias `chat_flotante.*` por `empresa_id` en `empresa_estacion_prefs` cuando hay contexto empresarial. Verificacion: `node --check` para los JS tocados y `go test ./...` en `backend/`.
- 2026-05-11: alcance vertical por licencia y tipo de empresa. El checkout valida que la licencia base corresponda al tipo elegido, la activacion aplica la preconfiguracion del tipo de empresa y `/api/empresa/permisos_contexto` agrega `vertical_scope` para ocultar/bloquear plantillas ajenos manteniendo el nucleo universal compartido. Pruebas: `go test ./handlers` y `go test ./db` en `backend/`.
- 2026-05-11: 2FA del login gobernado desde configuracion avanzada. `web/login.html` oculta el campo de codigo 2FA por defecto, `/config.js` publica `ADMIN_2FA_LOGIN_ENABLED`, `backend/handlers/auth_admin_handlers.go` solo exige OTP cuando `security.admin_2fa.enabled` esta activo y `web/super/configuracion_avanzada.html` permite activar/desactivar el switch global.
- 2026-05-11: catalogos publicos de plantillas realmente publicos. `backend/utils/utils.go` permite sin sesion `/api/public/plantillas_nuevas/catalogo` y `/api/public/plantillas_integracion/catalogo`; `backend/utils/auth_middleware_test.go` cubre ambas rutas para mantener la portada y fichas comerciales alineadas con el backend.
- 2026-05-11: correccion de cargas parciales en plantillas integrados. `backend/db/odontologia.go` y `backend/db/gimnasio.go` aseguran columnas de integracion antes de crear indices PostgreSQL sobre `cliente_id`, `servicio_id` y `carrito_id`; `web/js/consultorio_odontologico.js`, `web/js/gimnasio.js` y `web/js/alquileres.js` limpian avisos de carga parcial cuando la recarga ya no trae errores. No hay tablas ni dependencias nuevas.
- 2026-05-11: fix de arranque PostgreSQL para parqueadero. `backend/db/parqueadero.go` crea/asegura columnas de integracion (`cliente_id`, `servicio_id`, `carrito_id`, `carrito_item_id`) antes del indice `ix_parqueadero_ticket_empresa_carrito`, evitando fallo de despliegue en bases existentes sin `carrito_id`.
- 2026-05-11: consistencia del panel super. `web/js/super_administrador.js` mantiene la ayuda privada restaurable para `super_administrador` sin incluirla en la lista limitada de `control_super_administrador`, alineando el test frontend del panel.
- 2026-05-11: 20 plantillas nuevas reales. El ranking queda 1-20; `backend/db/plantillas_nuevas_bootstrap.go` asegura tipos, preconfiguraciones y licencias para todos, `web/js/plantillas_nuevas_catalogo.js` publica las 20 tarjetas, el index queda alineado con las plantillas reales y el panel super pasa a `Asegurar 20`. No hay tablas nuevas ni dependencias.
- 2026-05-11: portada index alineada a modulos reales. `web/index.html` y los defaults de `/api/public/pagina_principal` actualizan tarjetas publicas con nucleo unico y modulos reales; los 20 plantillas nuevas se publican como tarjetas operativas de `Probar gratis`. `web/js/plantillas_nuevas_catalogo.js` agrega decision, ranking, metadata de plantilla, permisos, flujo de venta y reportes. No hay endpoints, tablas, permisos ni dependencias nuevas.
- 2026-05-11: aseguramiento comercial de plantillas. `POST /super/api/plantillas_nuevas/catalogoaction=asegurar_20_licencias` asegura tipos de empresa, preconfiguraciones y cuatro planes recomendados para los 20 plantillas; se conserva `asegurar_v1_licencias` como alias compatible. No hay tablas, rutas nuevas ni dependencias.
- 2026-05-11: semaforo listo para venta en plantillas. `web/super/plantillas_produccion_masiva.html` cruza catalogo, preconfiguraciones y licencias activas para marcar `Listo venta` solo cuando el vertical tiene metadata completa, preconfiguracion activa con `integracion_vertical` y licencia activa que incluye el modulo. No hay cambios de esquema ni dependencias.
- 2026-05-11: acciones de gobierno para plantillas. La vista super `Plantillas 20` enlaza cada vertical con `Tipos de empresa`, `Preconfiguraciones` y `Licencias`; esas paginas aceptan `q`, `vertical` o `modulo` para abrir filtradas desde la matriz comercial. No se agregan endpoints, tablas ni dependencias.
- 2026-05-11: gobierno super de plantillas de produccion masiva. Se agrega `web/super/plantillas_produccion_masiva.html` y acceso `Plantillas 20` en el panel super para auditar ranking, metadata de plantilla, permisos, flujo de venta, reportes y exportacion CSV usando `/super/api/plantillas_nuevas/catalogo`. No hay endpoints, esquemas ni dependencias nuevas.
- 2026-05-11: preconfiguraciones conectadas a la matriz vertical extendida. `config_json` de tipos de empresa ahora puede incluir `integracion_vertical`; los catalogos de plantillas nuevas publican `integracion_preconfig`, decision, prioridad y bandera `produccion_masiva`. Se priorizan los 20 plantillas nuevas para produccion masiva. No hay cambios de esquema ni dependencias nuevas.
- 2026-05-11: matriz extendida de plantillas plantillas. El catalogo `/api/*/plantillas_integracion/catalogo` publica `template_activates`, `tables_touched`, `required_permissions`, `sale_flow` y `reports_produced`; la pantalla empresarial muestra esa informacion por vertical y la prueba de contrato impide plantillas visibles sin metadata completa. No hay cambios de esquema ni dependencias nuevas.
- 2026-05-11: pantalla de matriz de integracion vertical para empresa. Se agrega `web/administrar_empresa/plantillas_integracion.html`, enlace `linkPlantillasIntegracion` en Configuracion > Base empresarial y regla `seguridad:R`; la vista muestra KPIs, estado, nucleo usado, especialidad permitida y auditoria de plantilla por vertical.
- 2026-05-11: indicador compacto de integracion vertical en el panel empresarial. `web/administrar_empresa.html`, `web/js/administrar_empresa.js`, `web/js/plantillas_integracion_catalogo.js` y `web/estilos.css` muestran fuente API/local y conteo de plantillas visibles/ocultos usando el mismo contrato que controla el menu.
- 2026-05-11: el panel empresarial consume el catalogo API de integracion vertical antes de aplicar permisos de menu. `web/js/administrar_empresa.js` carga `/api/empresa/plantillas_integracion/catalogo` y `web/js/plantillas_integracion_catalogo.js` fusiona esos items sobre el respaldo local, evitando desalineacion entre backend y frontend si cambia la matriz.
- 2026-05-11: drogueria/farmacia validada como plantilla integrada al nucleo. La vertical usa `empresa_modulos_colombia_*` para expediente sanitario de lotes, INVIMA, formulas, controlados, dispensacion y farmacovigilancia; productos, inventario, compras, clientes, ventas, pagos y facturacion siguen en modulos centrales, y el menu vuelve a mostrarla como plantilla visible integrada.
- 2026-05-11: inicio de fases de integracion profesional de plantillas. Se agrega matriz de integracion, se marca cada vertical como plantilla integrada, soporte transversal o pendiente de integracion al nucleo, y el menu empresarial oculta plantillas clasicos que duplican clientes, productos/servicios, ventas o pagos hasta integrarlos al nucleo compartido. Los 20 plantillas nuevas siguen visibles como plantillas sobre `empresa_modulos_colombia_*` y el catalogo backend expone `integration_status`, `operational_visible`, `core_modules` y `duplicates_core`.
- 2026-05-11: profesionalizado el shell del super administrador. El menu principal queda reducido a gobierno, licencias, acceso, IA esencial y plataforma; se retiran accesos secundarios del panel visible y se bloquea la restauracion automatica de subpaginas retiradas desde sesion local.
- 2026-05-11: limpieza PostgreSQL-only del proyecto. Se retiran rastros del motor legado de codigo, frontend, scripts, documentacion vigente e historica, se cambian consultas residuales de indices a `pg_indexes`, se renombran helpers frontend a fechas de backend y se eliminan artefactos locales generados por perfiles temporales. No se agregan dependencias ni cambios en `go.mod`.
- 2026-05-11: actualizado `web/super/licencias_resumen.html` como centro de mando profesional del VPS y del proyecto. Consolida salud general, CPU/RAM/disco/trafico, PostgreSQL, alertas, errores recientes, servicios, procesos, licencias, empresas y consumo OpenAI estimado usando endpoints existentes, sin dependencias nuevas ni cambios de esquema.
- 2026-05-11: cierre implementable de pendientes 1 a 8. El checkout publico de licencias ahora permite seleccionar manualmente el pais de pago y reconsulta disponibilidad Wompi/Epayco por `pais_codigo`, evitando depender solo del navegador/VPN. Se corrigen referencias activas a documentos historicos inexistentes (`estructura_del_codigo` alias y plan maestro 14/15 puntos) hacia fuentes vigentes. Se deja explicitado que DIAN SOAP/WSDL oficial, hardware/proveedores reales, E2E con credenciales y normalizacion masiva de mojibake siguen como cierres externos/controlados, no como completados locales.
- 2026-05-11: implementada capa de madurez empresarial de 12 pasos: staging anonimizado por defecto, monitoreo Prometheus/Grafana, backups externos rclone/S3, deploy automatico opcional a staging, QA por roles, matriz de pagos/comprobantes, prueba de carga smoke, manifiesto de release, auditoria de soporte y normalizacion documental. Verificacion: `.\scripts\profesional_preflight.ps1 -Full` OK y carga smoke staging p95 1008 ms con error rate 0.
- 2026-05-10: Carritos suma modo tactil configurable por empresa/estacion para adaptar carrito operativo, cobro y agregador de productos por botones a tablets, monitores POS y pantallas tactiles.
- 2026-05-10: agregado modulo privado **Alertas del sistema** en super administrador. Configura destino, umbrales y enfriamiento; evalua disco VPS, trafico, sesiones administrativas y conexiones PostgreSQL; envia correo via Gmail SMTP y registra historial en `super_alertas_eventos`. Tambien se amplian metricas con `disk_total`, `disk_used` y `disk_percent`. Verificacion: `go test ./...` en `backend/`.
- 2026-05-10: actualizado sistema documental de roles/permisos con modulos finos (`crm_unificado`, `reservas_hotel`, `chat_tareas`, `horarios_trabajadores`, `asistencia_empleados`, `vehiculos_registro`, `hoja_vida_operativa`, `ubicacion_gps`, `nomina_sueldos`, `reportes`, `auditoria`, `backups`, `documentos_onlyoffice`, `nextcloud`), wrappers API especificos y compatibilidad de licencias amplias. La ayuda privada de super administrador queda accesible desde boton en `web/super_administrador.html` y sigue restringida a `super_administrador`. Documentacion agregada: `documentos/reporte_roles_ayuda_super_2026-05-10.md`.
- 2026-05-07: QA E2E Motel Calipso sobre `empresa_id=7`. Regresion escritorio 60/60 modulos, pruebas profundas 6/6 con datos QA reales (parqueadero QR/cobro/anulacion, WMS, centros de costo, activos fijos, red social con imagen, carta/venta publica), validacion movil dirigida de venta publica/asistencia y ajustes de robot/radio/favoritos para no tapar botones en celular. Reportes en `backend/tmp_tools/qa_calipso_operativo/*_report.json`.
- 2026-05-06: implementados `bancos_pagos`, `gestion_documental`, `cumplimiento_kyc`, `contratos_obligaciones` y `calidad_procesos` como modulos empresariales Colombia sobre nucleo compartido por `empresa_id`, con APIs privadas, pantallas administrativas, permisos/licencias, datos demo, exportacion CSV y documentacion.
- 2026-05-06: continuacion de fases para modulos empresariales Colombia con plantillas por modulo, seguimiento profesional, cambio de estado auditado, filtro por estado y UI compartida sin duplicar codigo.
- 2026-05-06: fase de reportes ejecutivos para modulos empresariales Colombia con metricas por estado/tipo/categoria/prioridad, vencimientos, criticidad, responsables faltantes, valor pendiente y recomendaciones automaticas.
- 2026-05-06: fase de importacion masiva para modulos empresariales Colombia con boton CSV, endpoint `importar_registros`, validacion parcial y bitacora por empresa.
- 2026-05-06: fase de evidencias para modulos empresariales Colombia con soportes por registro, endpoint `evidencia/evidencias`, validacion multiempresa y bitacora automatica.
- 2026-05-06: fase de aprobaciones para modulos empresariales Colombia con solicitudes por nivel, destinatario, vencimiento, decision aprobada/rechazada, cambio de estado y bitacora.
- 2026-05-06: fase de tareas para modulos empresariales Colombia con compromisos por registro, responsable, prioridad, vencimiento, estados operativos y bitacora.
- 2026-05-06: fase de expediente 360 para modulos empresariales Colombia con vista consolidada de registro, eventos, evidencias, aprobaciones, tareas, resumen y recomendacion.
- 2026-05-06: fase de agenda para modulos empresariales Colombia con alertas de vencidos, proximos vencimientos, tareas, aprobaciones pendientes, severidad y acceso al expediente.
- 2026-05-06: fase de cierre controlado para modulos empresariales Colombia, exigiendo evidencia y bloqueando aprobaciones pendientes o tareas abiertas antes de cerrar.
- 2026-05-06: fase de plan de accion para modulos empresariales Colombia, convirtiendo alertas de agenda en tareas con prioridad y control de duplicados.
- 2026-05-06: fase de responsables para modulos empresariales Colombia con tablero de carga por responsable, vencidos, tareas, aprobaciones y recomendaciones.
- 2026-05-06: fase de SLA para modulos empresariales Colombia con cumplimiento, semaforo, buckets de vencimiento y recomendaciones.
- 2026-05-06: fase de riesgo operativo para modulos empresariales Colombia con score 0-100, nivel bajo/medio/alto, factores y recomendaciones.
- 2026-05-06: fase de exportacion de auditoria para modulos empresariales Colombia con CSV multi-seccion de resumen, registros, agenda, SLA, riesgo, responsables, tareas, aprobaciones, evidencias y bitacora.
- 2026-05-06: fase de busqueda avanzada para modulos empresariales Colombia con filtros de backend por texto, estado, tipo, categoria, prioridad, responsable, vencidos y proximos vencimientos.
- 2026-05-06: fase de acciones masivas para modulos empresariales Colombia con seleccion de registros, cambio controlado de estado/prioridad/responsable y bitacora por registro.

- **Logistica avanzada / WMS**: nuevo modulo `logistica_wms` para operar bodega profesional con ubicaciones internas, ordenes WMS, picking, packing, despachos, rutas, avance por item, bitacora, datos demo y exportacion CSV. Integra `/api/empresa/logistica_wms`, permisos/licencia independientes, menu de Inventario y compras, pantalla `web/administrar_empresa/logistica_wms.html`, tablas WMS por `empresa_id` y documentacion. Verificacion: `go test ./db -run Test.*WMS -count=1`, `go test ./... -count=1` y `git diff --check`.

- **Declaraciones Tributarias y Motor de Impuestos Colombia**: nuevo modulo financiero `declaraciones_tributarias` para preliquidar, revisar y controlar IVA, retencion en la fuente, ReteIVA, ICA/ReteICA, consumo, renta y regimen simple por empresa. Agrega `/api/empresa/declaraciones_tributarias`, tablas de declaraciones, movimientos de conciliacion y calendario tributario editable, permisos/licencia independientes, pantalla `web/administrar_empresa/declaraciones_tributarias.html`, enlace en Centro financiero y documentacion. Verificacion: `go test ./db -run Test.*Declaracion -count=1`, `go test ./... -count=1` y `git diff --check`.

- **Captura inteligente de compras/gastos**: nuevo modulo empresarial `soportes_compras_ia` en Compras para cargar foto/PDF/XML, extraer datos con GPT-5.5, detectar duplicados, aprobar/rechazar y contabilizar como cuenta por pagar. Integra `/api/empresa/soportes_compras_ia`, permisos/licencia independientes, pantalla operativa, tablas de soportes/eventos y documentacion. Verificacion: `go test ./db -run Test.*Soporte.*IA -count=1`, `go test ./... -count=1` y `git diff --check`.

- **Portal contador / oficina virtual contable**: nuevo modulo `portal_contador` para firmas contables y contadores externos con portafolio de clientes, obligaciones DIAN/contables, solicitudes de documentos, comunicaciones, dashboard, datos demo y exportacion CSV. Integra `/api/empresa/portal_contador`, permisos/licencia independientes, menu financiero y documentacion. Verificacion: `go test ./db -run TestPortalContador -count=1` y `go test ./... -count=1`.

- **Gestion de cobranza profesional**: nuevo modulo financiero por empresa para recuperar cartera sin duplicar cuentas por cobrar. Agrega `/api/empresa/cobranza`, tablas de plantillas, campanas, gestiones y promesas de pago, permisos/licencia `cobranza`, pantalla `web/administrar_empresa/cobranza.html` en el Centro financiero y contable, datos demo, simulacion de envio y exportacion CSV. Verificacion: `go test ./db -run TestCobranza -count=1` y `go test ./... -count=1`.

- **Portal publico, carta QR, Motel Calipso, carnets y aislamiento multiempresa**: `web/index.html` actualiza las descripciones de modulos y tarjetas fallback para reflejar POS, hotel/motel, gimnasio, odontologia, domicilios tipo Rappi, Taxi System tipo Uber, turnos, carnets empresariales, control electrico, carta publica QR, red social, roles/licencias y hoja de vida. Se agrega `/api/empresa/carnets` y `web/administrar_empresa/carnets.html` para emitir carnets de empleados/usuarios con plantillas, QR, foto, exportacion PNG/SVG, impresion y bitacora. Se documenta la publicacion real de Motel Calipso (`motel-calipso`) con venta publica, carta publica, items y posts de red social. `AuthMiddleware` permite `visualizar_productos_y_precios_publico.html` directo y por slug sin sesion, manteniendo protegida la administracion. Los wrappers `WithEmpresa*` rechazan `empresa_id` contradictorios entre URL, cabecera, formulario/multipart y JSON para mantener separados todos los modulos empresariales. Se audita el menu empresa, el catalogo backend de paginas, las claves de licencia y rutas `/api/empresa` para evitar modulos duplicados o enlaces sin regla. Verificacion: `go test ./handlers`, `go test ./utils`, `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1` y HTTP 200 en produccion para venta publica, carta publica y red social.

- **Control electrico Raspberry Pi**: nuevo modulo en Administrar empresa para configurar IP/puerto/API de Raspberry Pi y mapear cada estacion/habitacion a multiples relés GPIO (luces, jacuzzi, aire, puerta u otros). El carrito de estacion agrega boton `Control electrico` para control manual de salidas; las estaciones envian `on` al activar/recuperar/reabrir y `off` al pagar/cerrar/desactivar. Archivos: `backend/db/control_electrico.go`, `backend/handlers/control_electrico.go`, `web/administrar_empresa/control_electrico.html`, `web/administrar_empresa/carrito_de_compras.html`, `backend/main.go`. Verificacion: `go test ./...` OK.

- **Pagos ePayco de licencias**: el fallback estandar se actualiza a la integracion oficial `checkout.js` con `external: "true"` y `PUBLIC_KEY`, evitando el POST legacy a `secure.payco.co/checkout.php` que podia terminar en "comercio no reconocido". `P_KEY` queda reservado al backend para confirmar webhooks con firma SHA256 y no se expone al navegador. Verificacion: `go test ./handlers -run Test.*Epayco -count=1` y `go test ./... -count=1`.

- **Administrar empresa - Hoja de vida operativa**: nuevo modulo universal para llevar historial de motos en taller, pacientes, vehiculos, equipos, mascotas, maquinaria o activos. Incluye ficha principal, eventos/servicios, alertas, recurrencia y resumen operativo. Archivos: `backend/db/hoja_vida_operativa.go`, `backend/handlers/hoja_vida_operativa.go`, `web/administrar_empresa/hoja_vida_operativa.html`, `web/administrar_empresa.html`, `web/js/administrar_empresa.js`, `backend/main.go`.

- **Flujo Git + VPS**: `scripts/pcs_deployment.local.ps1` (plantilla `pcs_deployment.local.ps1.example`, ignorada en git) centraliza `PcsGitRemoteUrl` y `PcsVpsHost` / ruta / puerto / SSH. `sync_to_vps.ps1` aplica la parte VPS; `publicar_git_y_vps.ps1` ejecuta `actualizar_repositorio.ps1` y luego el sync.

- **Despliegue Git local**: `scripts/actualizar_repositorio.ps1` admite fijar el repositorio remoto con `-RepoUrl`, variables `PCS_REPO_URL` / `REPO_URL`, o el archivo local `scripts/actualizar_repositorio.repo_url` (plantilla `*.repo_url.example`). Si `origin` apunta a otra URL, el script exige `-SetOrigin` para actualizar el remoto antes del push. Archivos: `scripts/actualizar_repositorio.ps1`, `scripts/actualizar_repositorio.repo_url.example`, `.gitignore`, documentación relacionada.

- **Chat global (super)**: la IA debe **preguntar confirmación** antes de emitir un bloque `PCS_ACTION` (resumen y pregunta en un turno; el JSON ejecutable solo tras un ?sí?/equivalente explícito del usuario). Texto de bienvenida actualizado en la UI.

- **Chat global (super)**: en cada pregunta el backend adjunta metadatos de **toda** la base `pcs_superadministrador` (conteos por tabla, columnas `nombre:tipo`, reparto de administradores por rol), sin filas con datos sensibles. La pantalla de lógica del chat deja de ofrecer el interruptor de ?contexto ampliado?. Archivos: `backend/db/chat_inteligencia_artificial.go`, `backend/handlers/super_chat_ia_logica.go`, `web/super/configuracion_logica_del_chat_con_ia.html`, documentaci?n relacionada.

- Chat IA (empresa y super): **Enter** envía el mensaje (mismo flujo que el botón); **Mayús+Enter** añade salto de línea en el textarea. Texto de ayuda bajo el campo. Archivos: `web/administrar_empresa/chat_con_inteligencia_artificial.html`, `web/super/chat_con_ia_global.html`.

- Super ? **Página principal** (`/super/pagina_principal.html`): sincronización de tema en iframe (cookie/localStorage), fondo del `body` con variables del tema, título de cabecera legible en modo claro y contenedor del editor con superficie neutra (`pp-main-card`) para evitar gradientes rosados en algunos temas oscuros; contraste consistente en todas las apariencias.

- **Permisos por rol** (`/super/permisos_rol.html`): consola empresarial para activar/desactivar acciones por módulo (R/C/U/D/A) y visibilidad por función del menú; API `GET /super/api/roles_de_usuario/permisos` con `modulos_etiqueta`, `acciones_etiqueta` y en cada `pagina` `titulo` y `grupo` (cat?logo en `empresa_permisos.go`). **Licencias** (`/super/licencias.html`): secci?n de cobertura por módulos con descripciones y enlace a la matriz de roles. Modelo: licencia = techo de módulos; rol = matriz y overrides de menú; sin un sistema ?universal? duplicado.

- Chat IA (empresa y super): interfaz tipo **Gemini** con barra lateral (chats guardados en el navegador + historial del servidor), barra superior con **modelo en uso** y resumen de cupo diario, **compartir** respuestas del asistente (compartir nativo o copiar), y **mensaje explicativo** cuando se alcanza el límite diario o el chat está bloqueado. Se retiran el encabezado largo y la tarjeta de chips en el chat empresarial; mejor contraste en modo claro. Archivos: `web/administrar_empresa/chat_con_inteligencia_artificial.html`, `web/super/chat_con_ia_global.html`, `web/estilos.css`, documentación relacionada.
- Chat IA: **acciones operativas confirmables** desde el chat (chat transaccional). El asistente propone un bloque `PCS_ACTION` (JSON) solo cuando tiene datos completos; el usuario confirma en la UI y el sistema ejecuta el endpoint real (productos, precios, finanzas, tarifas, servicios) manteniendo permisos y **auditoría**. La auditoría crítica registra `source=chat_ia` y `chat_conversation_id` en metadata. Archivos: `backend/handlers/chat_con_inteligencia_artificial_controller.go`, `backend/handlers/chat_con_ia_global_super.go`, `backend/handlers/auditoria_empresa.go`, `web/administrar_empresa/chat_con_inteligencia_artificial.html`, `web/super/chat_con_ia_global.html`, `web/estilos.css`.
- Super administrador: el menú agrega **Panel** como primer botón y el panel inicial (`licencias_resumen`) ahora muestra **licencias activas** y **cantidad de empresas registradas**. Se corrige el guardado de correos autorizados para **Frecuencia FP** en super (validación de sesión/rol) y la página se mejora a estilo CRUD (agregar/eliminar/listar). Los administradores en esa lista ven el acceso `Frecuencia FP` dentro de `administrar_empresa/configuracion_menu.html`.
- Licencias/pagos: el checkout de `pagar_licencia.html` registra correctamente el **código de asesor comercial** (`asesor_id`) al crear transacciones (Wompi/Epayco). El backend valida que el asesor exista y esté aceptado, y hace fallback de `empresa_id` desde la licencia si el frontend no lo envía; con esto la comisión queda trazable al aprobarse el pago.

- Estaciones: estación especial **Pedidos con IA** (texto o dictado) que interpreta el pedido y agrega productos al carrito de la estación indicada; configuración `ia_pedidos_enabled` / `ia_pedidos_placement` en `estaciones_config`. Vista móvil con botón **Ver miniaturas** / **Vista normal** (rejilla de 3 columnas). Archivos: `backend/handlers/ia_pedidos_estacion.go`, `backend/db/carritos_compras.go`, `backend/handlers/chat_con_inteligencia_artificial_router.go`, `web/administrar_empresa/estacion_ia_pedidos.html`, `web/administrar_empresa/estaciones.html`, `web/administrar_empresa/configuracion_de_estaciones.html`, `web/administrar_empresa/configuracion_carrito_de_compra_empresa.html`, `web/estilos.css`, `tools/clean_styles_auto.js`, `.vscode/extensions.json`, documentación relacionada.

- Asesor comercial profesional para licencias.
	- Tipo: mejora funcional mayor.
	- Archivos modificados/agregados/eliminados: `backend/db/asesor_comercial.go`, `backend/handlers/asesor_comercial.go`, `backend/handlers/payments_handlers.go`, `backend/main.go`, `backend/utils/utils.go`, `web/super/asesor_comercial.html`, `web/mis_clientes.html`, `web/pagar_licencia.html`, `web/seleccionar_empresa.html`, `web/js/seleccionar_empresa.js`, `web/super_administrador.html`, `web/super/vendedores_licencias.html` (eliminado), `backend/handlers/vendedores_handlers.go` (eliminado), `backend/handlers/vendedor_config_handlers.go` (eliminado), documentación relacionada.
	- Descripción: se reemplaza el módulo anterior por asesores comerciales con invitación por email, código único, porcentaje y plazo de asociación configurables. Los pagos de licencia con código generan comisión y las renovaciones siguen asociadas hasta vencer el plazo; super puede marcar comisiones pagadas y el asesor ve el estado en `Mis clientes`.

- Chat IA (empresa y super): tema sincronizado con el panel (`pcs_theme` / `localStorage`), estilos del chat basados en variables de tema y eliminación de las sugerencias tipo pill bajo el formulario. Ajustes de contraste en modo claro para calendario y paneles de `chat_y_tareas`.
	- Tipo: ajuste UX.
	- Archivos modificados: `web/administrar_empresa/chat_con_inteligencia_artificial.html`, `web/super/chat_con_ia_global.html`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.

- Red social empresarial y venta pública por páginas.
	- Tipo: mejora mayor.
	- Archivos modificados/agregados: `backend/db/red_social.go`, `backend/handlers/red_social.go`, `backend/db/venta_publica.go`, `backend/handlers/venta_publica.go`, `backend/main.go`, `backend/utils/utils.go`, `web/red_social_comercial.html`, `web/administrar_empresa/publicar_red_social.html`, `web/administrar_empresa/venta_publica.html`, `web/venta_publica.html`, `web/pagar_productos_de_venta_publica.html`, `web/administrar_empresa.html`, `web/administrar_empresa/configuracion_menu.html`, `web/administrar_empresa/configuracion_integraciones.html`, `web/js/administrar_empresa.js`.
	- Descripción: la red social pública ahora funciona como feed empresarial con posts medianos. Venta pública se separa en módulo propio para que cada empresa cree páginas por slug, publique productos existentes y cobre desde una página dedicada con sus credenciales Wompi/Epayco.
	- Verificación: `go test ./...`.

- Soporte remoto: portal publico RustDesk con cliente/servidor y configuracion central desde super.
	- Tipo: mejora mayor.
	- Archivos modificados: `backend/db/soporte_remoto.go`, `backend/handlers/soporte_remoto.go`, `backend/handlers/super_soporte_remoto.go`, `backend/handlers/soporte_remoto_test.go`, `backend/handlers/super_soporte_remoto_test.go`, `backend/handlers/auth_users_carritos_test.go`, `backend/utils/utils.go`, `web/administrar_empresa/soporte_remoto.html`, `web/super/soporte_remoto.html`, `web/soporte_remoto_acceso.html`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/estructura_bd.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÒ³n: el mÒ³dulo de soporte remoto ahora puede entregar una pÒ¡gina pÒºblica por sesiÒ³n que funciona como portal de asistencia estilo RustDesk. Esa pÒ¡gina expone descargas del cliente y del servidor, host, clave pÒºblica, ID/contraseÒ±a del dispositivo y visor web opcional. El panel de empresa y la mesa tÒ©cnica super comparten la misma configuraciÒ³n pÒºblica por `empresa_id`, y super puede editarla directamente sin salir de `/super/api/soporte_remoto`.

- Soporte remoto: lÒ­mite diario RustDesk por empresa desde super.
	- Tipo: mejora funcional.
	- Archivos modificados: `backend/db/soporte_remoto.go`, `backend/handlers/soporte_remoto.go`, `backend/handlers/super_soporte_remoto.go`, `backend/db/soporte_remoto_test.go`, `backend/handlers/soporte_remoto_test.go`, `backend/handlers/super_soporte_remoto_test.go`, `web/super/soporte_remoto.html`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/estructura_bd.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÒ³n: super ahora puede definir por empresa un tope diario en minutos para conexiones RustDesk. El backend calcula consumo diario real, bloquea nuevas sesiones o aprobaciones que excedan el cupo del dÒ­a y devuelve el resumen de uso para que la mesa tÒ©cnica vea el motivo exacto del bloqueo.

- Apariencia: primera visita del menÒº flotante ahora inicia en Blanco Corporativo.
	- Tipo: ajuste UX.
	- Archivos modificados: `web/menu.js`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÒ³n: cuando un usuario abre por primera vez una pÒ¡gina con menÒº flotante y todavÒ­a no tiene preferencia guardada en `localStorage`, cookie o backend, el sistema ahora arranca con la apariencia `light` (Blanco Corporativo) en lugar del tema oscuro. Las preferencias ya guardadas siguen respetÒ¡ndose sin cambios.

- Apariencia: contraste corregido para modo claro en portada y estilos compartidos.
	- Tipo: ajuste UX.
	- Archivos modificados: `web/index.html`, `web/estilos.css`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÒ³n: se quitÒ³ el blanco forzado del tÒ­tulo superior del index y se migraron varios bloques compartidos a variables de tema para que botones, tablas, login y secciones pÒºblicas cambien su color correctamente al alternar entre modo oscuro y claro.
# CHANGELOG

## 2026-05-06
- Portal de Terceros y Certificados Tributarios.
	- Se agrego el modulo `portal_terceros_certificados` con maestro de terceros, certificados de retencion/ingresos, enlace publico por token, impresion y bitacora de descargas.
	- Se agregaron las rutas `/api/empresa/portal_terceros_certificados` y `/api/public/certificados_tributarios`.
	- Se agregaron la pantalla administrativa `web/administrar_empresa/portal_terceros_certificados.html`, la pagina publica `web/visualizar_certificado_tributario_publico.html`, permisos, licencia, menu y documentacion.
	- Pruebas: `cd backend; go test ./... -count=1`.

- Activos Fijos e Intangibles NIIF/Fiscal.
	- Se formalizo el modulo `activos_fijos_niif_fiscal` reutilizando el nucleo de activos de la suite contable Colombia avanzada.
	- Se amplio `empresa_contabilidad_activos_fijos` con base fiscal, vida util fiscal, depreciacion fiscal acumulada, valor fiscal, diferencia NIIF/fiscal, deterioro, valor razonable y cuenta de deterioro.
	- Se agrego API `/api/empresa/activos_fijos_niif_fiscal`, pantalla administrativa, enlaces de menu, permisos por rol, licencias y documentacion.
	- Pruebas: `cd backend; go test ./... -count=1`.

- Propiedad horizontal y promocion por asesor.
	- Archivos: `backend/db/propiedad_horizontal.go`, `backend/db/propiedad_horizontal_test.go`, `backend/handlers/propiedad_horizontal.go`, `backend/handlers/asesor_comercial.go`, `backend/handlers/payments_handlers.go`, `backend/handlers/empresa_permisos.go`, `backend/main.go`, `web/administrar_empresa/propiedad_horizontal.html`, `web/super/asesor_comercial.html`, `web/pagar_licencia.html`, `web/super/licencias.html`, `documentos/propiedad_horizontal.md`, `documentos/promocion_asesor_licencias.md`.
	- Descripcion: se agrega modulo profesional de administracion de copropiedades con unidades, residentes, cargos, recaudos, PQR, asambleas y dashboard. Se agrega promocion global para que un codigo de asesor aceptado aplique descuento adicional configurable al comprar licencias.
	- Verificacion: `go test ./... -count=1` en `backend/`.

- Cierre y bloqueo fiscal avanzado.
	- Archivos: `backend/db/cierre_fiscal.go`, `backend/db/cierre_fiscal_test.go`, `backend/handlers/cierre_fiscal.go`, `backend/handlers/empresa_permisos.go`, `backend/db/contabilidad_colombia.go`, `backend/main.go`, `web/administrar_empresa/cierre_fiscal.html`, `web/administrar_empresa.html`, `web/administrar_empresa/finanzas_menu.html`, `web/js/administrar_empresa.js`, `web/super/licencias.html`, `documentos/cierre_fiscal.md`.
	- Descripcion: se agrega modulo profesional para periodos fiscales, politicas de bloqueo por modulo, dias de edicion retroactiva, reaperturas con motivo, excepciones aprobadas, validacion de operaciones y bitacora de bloqueos por empresa. El cierre/reapertura de Contabilidad Colombia sincroniza el periodo fiscal.
	- Verificacion: `go test ./... -count=1` en `backend/`.

- Centros de costo y rentabilidad.
	- Archivos: `backend/db/centros_costo.go`, `backend/handlers/centros_costo.go`, `backend/handlers/empresa_permisos.go`, `backend/main.go`, `web/administrar_empresa/centros_costo.html`, `web/administrar_empresa.html`, `web/administrar_empresa/finanzas_menu.html`, `web/js/administrar_empresa.js`, `documentos/centros_costo.md`.
	- Descripcion: se agrega modulo formal `centros_costo` para medir rentabilidad por sucursal, area, unidad de negocio o proyecto, con maestro, reglas de imputacion, presupuesto por periodo, dashboard comparativo, movimientos integrados desde contabilidad/tesoreria/compras/OCR/AIU y exportacion CSV, sin duplicar Finanzas ni Contabilidad.
	- Verificacion: `go test ./... -count=1` en `backend/`.

- QA transversal de modulos y profesionalizacion operativa.
	- Archivos: `backend/db/cobranza.go`, `backend/db/portal_contador.go`, `backend/db/soportes_compras_ia.go`, `web/js/administrar_empresa.js`, `web/administrar_empresa/soportes_compras_ia.html`, `web/index.html`, `documentos/reporte_qa_modulos_2026-05-06.md`.
	- Descripcion: se alinean permisos locales para `administrador_total`, se optimizan dashboards nuevos para no repetir validaciones de esquema, se endurecen enlaces dinamicos de soportes IA y se documenta la prueba autenticada de Motel Calipso con paginas/API 200. La portada publica actualiza la descripcion de modulos para incluir Cobranza, Portal contador, Captura IA/OCR, AIU construccion, Parqueaderos con ticket QR y Apartamentos turisticos.
	- Verificacion: `go test ./... -count=1`, `git diff --check`, auditoria estatica de enlaces/botones/IDs y pruebas HTTP autenticadas contra `empresa_id=7`.

- AIU construccion: mejora profesional del modulo de contratos de obra.
	- Archivos: `backend/db/aiu_construccion.go`, `backend/handlers/aiu_construccion.go`, `web/administrar_empresa/aiu_construccion.html`, pruebas y documentacion.
	- Descripcion: se agregan responsable, centro de costo, modalidad contractual, riesgo, avance, retenciones, anticipo, garantia, neto a cobrar, flujo validado de estados/aprobacion, reporte, facturas recientes, alertas y exportacion CSV. El calculo AIU ahora separa total de factura y neto operativo a cobrar.
	- Verificacion: `go test ./db -run Test.*AIU -count=1` y `go test ./...`.

- AIU construccion: modulo profesional para arquitectos, constructoras y contratos de obra.
	- Archivos: `backend/db/aiu_construccion.go`, `backend/handlers/aiu_construccion.go`, `web/administrar_empresa/aiu_construccion.html`, menu de facturacion, permisos, pruebas y documentacion.
	- Descripcion: se agrega gestion de contratos AIU por empresa, capitulos/conceptos de obra, calculo de Administracion/Imprevistos/Utilidad, modelos de base AIU no sumada o sumada al total, base IVA configurable y generacion de factura electronica AIU enlazada a `empresa_facturacion_documentos`.
	- Verificacion: `go test ./db -run Test.*AIU -count=1` y `go test ./...`.

- Facturacion electronica Colombia: se amplian los documentos electronicos del ciclo DIAN/proveedor para cubrir brechas frente a Siigo y al ecosistema DIAN.
	- Archivos: `backend/handlers/facturacion_electronica.go`, `backend/handlers/documentos_lifecycle.go`, `backend/db/facturacion_electronica.go`, `web/administrar_empresa/facturacion_electronica.html`, pruebas y documentacion.
	- Descripcion: el mismo modulo por empresa ahora permite emitir/anular o registrar en cola fiscal factura electronica, nota credito, nota debito, documento soporte, nomina electronica y documento equivalente POS electronico. Se agrega selector operativo en la UI, botones rapidos para cada documento y normalizacion de aliases usados por Siigo/DIAN.
	- Verificacion: `go test ./handlers -run "TestNormalizeFacturacionDocumentoElectronicoTipo|TestResolveFacturacionTransitionForDocumentosElectronicosNuevos"` y `go test ./db -run TestDefaultFacturacionConfigPaisAplicaProveedorYCampos`.

## 2026-05-04
- Asesor comercial: configuracion profesional de transferencias y pagos de comisiones por venta de licencias.
	- Archivos: `backend/db/asesor_comercial.go`, `backend/handlers/asesor_comercial.go`, `web/super/asesor_comercial.html`, documentacion relacionada.
	- Descripcion: el modulo de asesor comercial ahora guarda metodo de pago, entidad financiera, tipo/numero de cuenta, titular, documento, contacto de pagos, periodicidad, dia de pago, minimo a transferir y exigencia de soporte. Las comisiones de licencias incorporan estado de pago (`pendiente`, `programada`, `en_transferencia`, `pagada`, `rechazada`), metodo, referencia, fecha programada y soporte.
	- Verificacion: `go test ./...` en `backend/`, validacion de scripts embebidos de `web/super/asesor_comercial.html` y `git diff --check`.

## 2026-04-30
- Raspberry Pi y sensores: provisionamiento seguro por empresa.
	- Archivos: `backend/db/sensor_puertas.go`, `backend/handlers/sensor_puertas.go`, `web/administrar_empresa/configuracion_sensores_raspberry.html`, pruebas y documentacion relacionada.
	- Descripcion: la configuracion de sensores ahora permite provisionar desde el servidor un `device_id` normalizado y un token de 64 caracteres para `X-Device-Token`, con ejemplos curl/Python para instalar en Raspberry Pi. Los dispositivos con token configurado ya no aceptan heartbeats ni mensajes publicos enviados solo con `device_id`.
	- Verificacion: `go test ./db -run "TestNormalizeEmpresaSensor|TestGenerateEmpresaSensor" -count=1`, `go test ./handlers -run "TestBuildEmpresaSensorProvisioningPayload" -count=1`, validacion del script embebido de la pagina y `go test ./... -count=1`.

- Chat IA: exportacion de documentos generados desde conversaciones.
	- Archivos: `backend/handlers/dynamic_documents.go`, `backend/handlers/chat_con_inteligencia_artificial_router.go`, `backend/handlers/dynamic_documents_test.go`, `web/js/ai_chat_drawer.js`, `web/estilos.css`, documentacion relacionada.
	- Descripcion: las respuestas del chat IA que parecen documentos, reportes, contratos, cotizaciones, actas o tablas muestran botones de exportacion PDF, Word/DOCX, Excel/XLSX, TXT y JSON. El frontend llama a `/api/empresa/chat_documentos/exportar`, el backend reutiliza el generador dinamico, redacta patrones de secretos, prepara nombres profesionales con empresa/tipo/fecha, registra auditoria con origen `chat_ia` y conserva fallback TXT/JSON si falla un conversor.
	- Verificacion: `go test ./handlers -run "TestDynamicDocument" -count=1`, `go test ./handlers -count=1`, `node --check web/js/ai_chat_drawer.js` con runtime Node empaquetado.

- Checkout de licencias Epayco: fallback clasico seguro por POST.
	- Archivos: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `web/pagar_licencia.html`, documentacion relacionada.
	- Descripcion: cuando Smart Checkout v2 no autentica o no crea sesion, el backend ya no entrega una URL GET a `checkout.epayco.co/checkout.php`. Ahora genera un formulario clasico firmado para `https://secure.payco.co/checkout.php`, devuelve `checkout_form` al frontend y registra en `raw_payload` una version sanitizada sin exponer `p_key`. El modo del formulario clasico se resuelve con `epayco.customer_id` + `epayco.checkout_key`/`epayco.p_key`, separado del modo Smart Checkout, para evitar enviar comercios reales como solicitud de pruebas (`p_test_request=true`) y provocar "El comercio no fue reconocido". Si falta `epayco.customer_id` o `P_KEY`, devuelve error controlado.
	- Verificacion: `go test ./handlers -run "TestBuildEpaycoClassicCheckoutForm|TestResolveEpaycoClassicMode|TestPickEpaycoField|TestSanitizeEpaycoClassicCheckoutForm" -count=1`, `go test ./handlers -count=1`, `go test ./...` en `backend/`, `node --check` del script inline de `web/pagar_licencia.html` y `git diff --check`.

- Chat flotante: secretaria IA rediseñada y voz femenina.
	- Archivos: `web/js/ai_chat_drawer.js`, `web/estilos.css`, `web/administrar_empresa/configuracion_chat_flotante.html`, documentacion relacionada.
	- Descripcion: el avatar `Secretaria IA 3D` pasa a una apariencia estilo caricatura ejecutiva joven, con rostro mas amable, ropa ejecutiva, detalles visuales y animaciones existentes. Cuando el modo activo es secretaria, la voz efectiva se fuerza a `es-CO-female` para el proxy de voz y para Web Speech; el robot conserva la voz configurable.
	- Verificacion: `node --check web/js/ai_chat_drawer.js` y `git diff --check`.

- Empresas compartidas: visibilidad y eliminacion de accesos.
	- Archivos: `backend/handlers/empresa_compartida_handlers.go`, `web/js/editar_empresa.js`, `web/js/seleccionar_empresa.js`, documentacion relacionada.
	- Descripcion: el editor de empresa lista los administradores con quienes se compartio la empresa y permite retirar accesos compartidos. La informacion queda visible para quien compartio y para quien recibio acceso; la revocacion mantiene trazabilidad del actor y del motivo operativo.
	- Verificacion: `go test ./...`.

- Documentos dinamicos asistidos por IA.
	- Archivos: `backend/handlers/dynamic_documents.go`, `backend/handlers/dynamic_documents_test.go`, `backend/main.go`, documentacion relacionada.
	- Descripcion: se agregan endpoints protegidos `/generate` y `/download` para recibir contenido o prompt IA, aplicar variables, renderizar HTML con templates Go y descargar PDF, DOCX, XLSX, HTML, TXT o JSON. La generacion usa GPT-5.4 mini cuando se solicita IA y conserva flujo temporal de archivos sin exponer credenciales al modelo.
	- Verificacion: pruebas unitarias dirigidas del handler de documentos dinamicos.

## 2026-04-29
- Reportes globales y busqueda en configuracion avanzada.
	- Archivos: `backend/handlers/reportes_globales.go`, `web/super/reportes_globales.html`, `web/js/super_reportes_globales.js`, `web/super/configuracion_avanzada.html`.
	- Descripcion: reportes globales en seleccionar empresa ahora permite a la cuenta super principal ver todas las empresas, corrige textos con codificacion danada y agrega manejo mas claro de respuestas no JSON. Configuracion avanzada incorpora un submenu sticky con buscador por categoria y seccion.
- Voz IA streaming segura.
	- Archivos: `backend/handlers/voice_stream_config.go`, `backend/handlers/super_config_backup_handlers.go`, `services/voice_stream_server/app.py`, `services/voice_stream_server/README.md`, `scripts/install_voice_stream_server_vps.sh`, `web/super/voz_streaming_ia.html`, `web/super/configuracion_avanzada.html`.
	- Descripcion: el servicio de voz puede activarse/probarse desde super administrador, registra token cifrado en configuraciones, envia autenticacion al health/TTS y el microservicio puede exigir el mismo token por header.
- Epayco y chat flotante movil.
	- Archivos: `web/pagar_licencia.html`, `web/administrar_empresa.html`, `web/seleccionar_empresa.html`, `web/super_administrador.html`, `web/estilos.css`.
	- Descripcion: el boton `Pagar con Epayco` precarga Smart Checkout v2, abre con tipo compatible para movil y prueba fallback entre `onpage`/`standard` si la primera apertura falla. El chat flotante mueve `Nuevo chat` al toolbar inferior y compacta en una sola fila nuevo mensaje, modo conversacion, microfono, voz, selector de modo y adjuntar.
	- Verificacion: validacion de scripts HTML/JS con Node y `go test ./handlers`.

- Checkout de licencias: tarjetas comerciales y asesor bajo descuento.
	- Archivos: `web/pagar_licencia.html`, `web/estilos.css`.
	- Descripcion: las tarjetas del pago de licencia quedan con presentacion mas profesional de venta, beneficios visibles y precio destacado. El campo `Codigo de asesor comercial` queda debajo del `Codigo de descuento`, conservando el envio de `asesor_id` a las pasarelas y activacion sin pago.
	- Verificacion: validacion de scripts de `web/pagar_licencia.html` con Node.

- Preconfiguracion de tipos de empresa.
	- Archivos: `backend/db/tipo_empresa_preconfiguracion.go`, `backend/handlers/empresa_preconfiguracion.go`, `backend/db/empresa_estacion_prefs.go`, `backend/db/productos.go`, `backend/handlers/system_empresas_handlers.go`, `backend/main.go`, `web/super/preconfiguracion_tipos_empresa.html`, `web/js/seleccionar_empresa.js`, `web/super_administrador.html`, `Pendiente Notas`, documentacion relacionada.
	- Descripcion: Super Administrador puede definir plantillas iniciales por tipo de empresa. Al crear una empresa se aplican estaciones y productos guia si la plantilla esta activa; restaurante trae 5 mesas y productos demo. Tras crearla, el administrador decide si conserva la preconfiguracion o la elimina para dejar la empresa limpia.
	- Verificacion: `go test ./db`, `go test ./handlers` y validacion de scripts de `seleccionar_empresa` / `preconfiguracion_tipos_empresa`.

- Estaciones: tarjetas sin filas vacias.
	- Archivos: `web/administrar_empresa/estaciones.html`, `web/estilos.css`, documentacion relacionada.
	- Descripcion: las tarjetas de estaciones normales reorganizan sus datos en filas etiqueta/valor tipo tarjeta Caja. Las filas sin dato quedan ocultas, conservando una sola columna de informacion ordenada por estado, cliente, tarifa, duracion, inicio, fin, extra y total.

- Voz IA streaming y reorganizacion super.
	- Archivos: `backend/handlers/voice_stream_config.go`, `backend/main.go`, `services/voice_stream_server/*`, `scripts/install_voice_stream_server_vps.sh`, `web/js/ai_chat_drawer.js`, `web/super/voz_streaming_ia.html`, `web/super_administrador.html`, `web/super/configuracion_avanzada.html`, documentacion relacionada.
	- Descripcion: se agrega un servicio abierto FastAPI + Piper TTS para voz natural en VPS, configurable desde Super Administrador y desactivado por defecto. El backend expone configuracion protegida, estado publico no sensible y proxy TTS; el chat flotante intenta usar el servidor de voz y cae a voz del navegador/texto si esta desactivado o falla. El menu super y Configuracion avanzada quedan agrupados por categorias empresariales.

- Tarjetas del index con modo informacion/foto o banner.
	- Archivos: `backend/handlers/pagina_principal_handlers.go`, `web/super/pagina_principal.html`, `web/index.html`, `web/estilos.css`, documentacion relacionada.
	- Descripcion: la configuracion super de Pagina principal permite elegir por tarjeta entre `tarjeta informacion mas foto` (default compatible con lo existente) y `tarjeta banner`. El modo banner usa solo la imagen principal y la renderiza ocupando toda la tarjeta del index. La pantalla permite subir imagenes a `web/img` y muestra medidas recomendadas 4:3 por tamano de tarjeta.
	- Verificacion: `go test ./...` en `backend/` compilo y ejecuto paquetes; Windows bloqueo temporalmente el borrado del binario de test al final, luego se limpio `.gotmp6`.

- Reorganizacion empresarial de Administrar Empresa y Configuracion.
	- Archivos: `web/administrar_empresa.html`, `web/administrar_empresa/configuracion_menu.html`, `web/administrar_empresa/configuracion.html`, `web/js/administrar_empresa.js`, `web/estilos.css`, documentacion relacionada.
	- Descripcion: el menu principal queda agrupado por modulos de trabajo (colaboracion, operacion/ventas, inventario/compras, finanzas/cumplimiento, personas/activos, analisis/control, documentos/nube/soporte y administracion). Configuracion se reorganiza por base empresarial, ventas/cobro, estaciones/tarifas, fiscal/automatizacion y avanzado. Se conserva la compatibilidad con rutas, iframes y permisos, y se eliminan duplicados de navegacion.

- Auditoria IA profunda y consultas DB seguras para GPT-5.4 mini/modelo activo.
	- Archivos: `backend/db/auditoria_empresa.go`, `backend/db/chat_inteligencia_artificial.go`, `backend/handlers/auditoria_empresa.go`, `backend/handlers/chat_con_inteligencia_artificial_controller.go`, `backend/handlers/chat_con_ia_global_super.go`, `backend/handlers/super_chat_ia_logica.go`, `web/super/configuracion_logica_del_chat_con_ia.html`, documentacion relacionada.
	- Descripcion: el chat empresarial y el chat global reciben auditoria en tiempo real, busqueda profunda de eventos por intencion y resultados de consultas DB seguras ejecutadas por el backend con whitelist/empresa_id. Se agrega `empresa_auditoria_ia_consultas` y lectura total controlada de DB para el chat empresarial, activa por defecto y configurable desde super (`ai.chat.empresa.db_query_enabled`, tablas maximas y filas por tabla). La IA no ejecuta SQL libre y el servidor degrada sin romper si auditoria, lectura DB o proveedor IA fallan.
	- Verificacion: `go test ./...` en `backend/` usando `GOTMPDIR=.gotmp4`.

## 2026-04-25
- Facturación electrónica: perfiles independientes para Ecuador (SRI) y Panamá (DGE/DGI) frente a Colombia (DIAN); detección por licencia; API y UI con `vista` por país.
	- Archivos: `backend/db/facturacion_electronica.go`, `backend/handlers/facturacion_electronica.go`, `web/administrar_empresa/facturacion_electronica.html`, `web/estilos.css`, `documentos/historial_de_cambios`, `documentos/descripcion_de_modulos`, `documentos/descripcion_de_archivos`, `CHANGELOG.md`.

- Configuración lógica del chat global (super): contexto ampliado de la base superadministrador y modo solo lectura de datos de empresas inyectados al prompt; ajuste de legibilidad en estaciones (vista miniaturas móvil).
	- Archivos: `web/super/configuracion_logica_del_chat_con_ia.html`, `backend/handlers/super_chat_ia_logica.go`, `backend/handlers/chat_con_ia_global_super.go`, `backend/db/chat_inteligencia_artificial.go`, `web/estilos.css`, `documentos/historial_de_cambios`, `documentos/descripcion_de_archivos`, `documentos/descripcion_de_modulos`, `CHANGELOG.md`.

## 2026-04-23
- Retiro del modulo Tipos de usuario en panel super: API, handlers, pagina web y tabla `tipos_de_usuario` (DROP al arranque); solo permanecen roles y permisos por rol.
	- Archivos modificados/eliminados: `backend/main.go`, `backend/handlers/roles_tipos_usuario.go`, `backend/db/roles_tipos_usuario.go`, `web/super_administrador.html`, `web/js/super_administrador.js`, `web/super/tipos_de_usuario.htm` (eliminado), `documentos/estructura_bd.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.

- Documentacion y operacion reCAPTCHA (v2/v3/Enterprise) + copias de respaldo bajo `backup/`.
	- Archivos modificados: `documentos/manual_de_instalacion.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `documentos/CHANGELOG.md`, `CHANGELOG.md`, `backup/.gitkeep`, `backup/empresas/.gitkeep`, `backup/super_administrador/.gitkeep`.
	- Descripcion: el manual ahora documenta la configuracion super, el provider y los mensajes tipicos de dominio en Google, y describe el almacenamiento best-effort de copias JSON en `backup/super_administrador` y `backup/empresas/<empresa_id>`.

## 2026-04-20
- Soporte remoto y RustDesk: activacion por defecto en configuracion super.
	- Archivos modificados: `backend/db/soporte_remoto.go`, `backend/handlers/soporte_remoto.go`, `backend/db/soporte_remoto_test.go`, `backend/handlers/super_soporte_remoto_test.go`, `web/super/servidores.html`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la configuración sin registro previo del módulo de soporte remoto ahora nace activa por defecto y alineada a la operación simplificada de RustDesk, incluyendo portal público habilitado, descargas completas y modo `cliente_local` sobre `rustdesk_oss`. La vista super mantiene el mismo default en primera carga para no mostrar un falso inactivo.
	- Verificación: `go test ./db -run '^TestSoporteRemotoDBFlow$' -count=1` y `go test ./handlers -run '^TestSuperSoporteRemotoHandlerConfigGetAndUpdate$' -count=1`.

## 2026-04-21
- Compras y finanzas: comprobantes adjuntos por empresa.
	- Archivos modificados: `backend/db/documentos_transaccionales.go`, `backend/db/finanzas.go`, `backend/handlers/compras.go`, `backend/handlers/finanzas.go`, `backend/handlers/compras_documentos_test.go`, `backend/handlers/eventos_contables_modulos_test.go`, `backend/main.go`, `web/administrar_empresa/compras.html`, `web/administrar_empresa/finanzas.html`, `documentos/estructura_bd.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el sistema permite adjuntar recibos, fotos y soportes documentales a documentos de compras y a movimientos de finanzas, guardándolos por empresa en `web/uploads/comprobantes/empresa_<id>/compras` y `.../finanzas`. Los listados empresariales ahora muestran acceso directo para ver el comprobante, y compras amplía su persistencia con metadata explícita del archivo cargado.
	- Verificación: diagnóstico del editor sin errores en los archivos tocados. La corrida dirigida de `go test` quedó bloqueada por un error preexistente de compilación en `backend/db/carritos_compras.go` ajeno a este cambio.

## 2026-04-20
- Apariencia frontend: limpieza estructural final en ayuda y vendedores.
	- Archivos modificados: `web/estilos.css`, `web/super/vendedores_licencias.html`, `web/ayuda/ayuda.html`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÒ³n: `vendedores_licencias.html` elimina los `style=` residuales de layout y estados, mientras `ayuda.html` se normaliza por secciones completas de encabezados y pÒ¡rrafos auxiliares. El resultado es una herencia de tema mÒ¡s consistente y menos deuda visual repetida en pÒ¡ginas largas del panel super y de ayuda.
	- VerificaciÒ³n: diagnÒ³stico del editor sin errores en los tres archivos tocados y bÒºsqueda sin remanentes de `style="` en `web/super/vendedores_licencias.html` ni en `web/ayuda/ayuda.html`.

- Soporte remoto y RustDesk: simplificaciÒ³n operativa del mÒ³dulo y control real del servicio desde super.
	- Archivos modificados: `backend/db/soporte_remoto.go`, `backend/handlers/soporte_remoto.go`, `backend/handlers/super_soporte_remoto.go`, `backend/handlers/super_servidores_handlers.go`, `backend/main.go`, `backend/handlers/soporte_remoto_test.go`, `backend/handlers/super_soporte_remoto_test.go`, `web/super/servidores.html`, `web/administrar_empresa/soporte_remoto.html`, `web/super/soporte_remoto.html`, `web/soporte_remoto_acceso.html`, `documentos/descripcion_de_archivos`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÒ³n: el frente de soporte remoto se simplifica para dejar solo la configuraciÒ³n esencial de RustDesk por empresa y el panel super pasa a tener control operativo del servicio del VPS con prueba real. Se agregan URLs macOS, se consolidan descargas oficiales para cliente y servidor y se elimina de las pantallas la lÒ³gica extensa de visor/sesiones como superficie principal. La UI super se unifica ademÒ¡s en una sola vista `RustDesk`, donde conviven estado del servidor, acciones del VPS y configuraciÒ³n mÒ­nima por empresa.
	- VerificaciÒ³n: `get_errors` sin errores en las cuatro vistas HTML nuevas y `go test ./handlers -run 'Test(SuperSoporteRemotoHandlerConfigGetAndUpdate|PublicSoporteRemotoResolverAccesoExponeDescargasRustDesk|SuperServidoresProbeHandlerReturnsRustDeskStatus|SuperSoporteRemotoHandlerListsCompaniesAndCreatesSession)$' -count=1`.

## 2026-04-20
- Portal publico: el header del home ahora comparte estilo con la landing y expone `Crear cuenta`.
	- Archivos modificados: `web/index.html`, `web/estilos.css`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÒ³n: `index.html` reemplaza el botÒ³n superior propio por el mismo estilo `btn dark` usado en `descripcion_de_los_sistemas.ht`, agrega un CTA `Crear cuenta` hacia `/registrar_nuevo_usuario_administrador.html` junto a `Iniciar sesiÒ³n` y ajusta el responsive del header para que ambos botones se mantengan compactos y legibles en mÒ³vil.
	- VerificaciÒ³n: diagnÒ³stico del editor sin errores en `web/index.html` y `web/estilos.css`.

- Apariencia frontend: segunda pasada de limpieza de estilos inline en mÒ³dulos operativos.
	- Archivos modificados: `web/estilos.css`, `web/login.html`, `web/login_usuario.html`, `web/accept.html`, `web/configuracion_de_la_cuenta.html`, `web/administrar_empresa/administrar_productos.html`, `web/administrar_empresa/carrito_de_compras.html`, `web/administrar_empresa/administrar_clientes.html`, `web/administrar_empresa/administrar_usuarios.html`, `web/administrar_empresa/compras.html`, `web/administrar_empresa/facturas_electronicas.html`, `web/administrar_empresa/chat_con_inteligencia_artificial.html`, `web/red_social_comercial.html`, `web/administrar_empresa/graficos_estadisticas.html`, `web/administrar_empresa/estaciones.html`, `web/js/super_reportes_globales.js`, `web/super/chat_con_ia_global.html`, `web/super/configuracion_avanzada.html`, `web/super/reportes_globales.html`, `web/super/vendedores_licencias.html`, `web/ayuda/login_administradores.html`, `web/administrar_empresa/soporte_remoto_view.html`, `web/administrar_empresa/historial_productos.html`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÒ³n: se reemplazan estilos inline de color, gradiente, ocultaciÒ³n, espaciado y alineaciÒ³n por utilidades CSS y clases semÒ¡nticas para reducir deuda visual y mejorar consistencia entre los seis temas del sistema. Parte del HTML generado desde JavaScript ahora usa clases y `data-*`, y los mensajes de error/Ò©xito en runtime dejan de depender de hex fijos.
	- VerificaciÒ³n: diagnÒ³stico del editor sin errores en los archivos principales tocados, bÒºsqueda sin remanentes de `style="...color/background..."` en el frente intervenido, sin asignaciones `style.color = "#..."` en los JS afectados y corrida dirigida `go test ./handlers -run 'Test(SuperSoporteRemotoHandlerConfigGetAndUpdate|PublicSoporteRemotoResolverAccesoExponeDescargasRustDesk|SuperServidoresProbeHandlerReturnsRustDeskStatus)$' -count=1` satisfactoria.

- Permisos super: pruebas dirigidas migradas a PostgreSQL y middleware robustecido.
	- Archivos modificados: `backend/handlers/postgres_test_helpers_test.go`, `backend/handlers/system_empresas_handlers_test.go`, `backend/handlers/auth_users_carritos_test.go`, `backend/handlers/auth_admin_handlers_test.go`, `backend/utils/utils.go`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÒ³n: el bloque de pruebas de permisos del panel super ya no depende de motor legado retirado y ahora usa esquemas efÒ­meros en PostgreSQL, alineados con el tÒºnel local documentado del proyecto. En la misma iteraciÒ³n, `AuthMiddleware` deja de fallar cuando no hay conexiÒ³n `dbSuper` en flujos pÒºblicos, evitando un `panic` en la validaciÒ³n de rutas pÒºblicas de licencias.
	- VerificaciÒ³n: `go test ./handlers -run 'TestSuperEndpointsPermisosPorRol|TestAdministradorPuedeEditarYEliminarEmpresaDesdeRutaSuperProtegida|TestNuevoAdminRegistradoPuedeCrearSuPrimeraEmpresaViaRutaSuperProtegida' -count=1` y `go test ./utils -run '^TestAuthMiddlewareAllowsPublicLicenciaPaymentRoutesWithoutSession$' -count=1`.

- Apariencia global: contraste y componentes alineados en los seis temas.
	- Archivos modificados: `web/estilos.css`, `web/administrar_empresa/soporte_remoto.html`, `web/super/soporte_remoto.html`, `web/administrar_empresa/publicar_red_social.html`, `web/red_social_comercial.html`, `web/pantalla_publica.html`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÒ³n: el sistema deja de depender de varios colores fijos que rompÒ­an el contraste en modo claro y pasa a usar variables de tema para paneles, tarjetas, mÒ³dulos embebidos, estaciones especiales y estados informativos. Con esto, los seis modos de apariencia mantienen mejor legibilidad y coherencia visual.
	- VerificaciÒ³n: diagnÒ³stico del editor sin errores en los archivos frontend tocados.

- Login administrativo: recuperaciÒ³n por enlace directo sin token manual.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/handlers/super_email_templates.go`, `backend/handlers/auth_admin_handlers_test.go`, `web/login.html`, `web/js/login.js`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÒ³n: el correo de recuperaciÒ³n administrativa ahora dirige al usuario a un enlace directo de restablecimiento y la pantalla de `login.html` ya no solicita ingresar token manual. El token sigue validÒ¡ndose en backend, pero queda encapsulado en el enlace recibido por correo.
	- VerificaciÒ³n: prueba dirigida del handler administrativo de recuperaciÒ³n/restablecimiento y diagnÒ³stico del editor sin errores en archivos tocados.

- Correo de recuperaciÒ³n: el enlace largo se reemplaza por un botÒ³n `Cambiar contraseÒ±a`.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/handlers/super_email_templates.go`, `backend/handlers/auth_admin_handlers_test.go`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÒ³n: la versiÒ³n HTML de los correos de recuperaciÒ³n ahora prioriza un botÒ³n `Cambiar contraseÒ±a` y deja la URL extensa solo como respaldo visible. El envÒ­o administrativo tambiÒ©n se genera como `multipart/alternative` para servir texto plano y HTML en el mismo mensaje.
	- VerificaciÒ³n: `go test ./handlers -run '^TestAdminPasswordRecoveryTemplateRendersButton$' -count=1` y diagnÒ³stico del editor sin errores en `backend/handlers/auth_admin_handlers.go` y `backend/handlers/super_email_templates.go`.

- Backups empresariales: se implementa la Fase 2 de exportar/importar configuracion por empresa.
	- Archivos modificados: `backend/db/backups_empresariales.go`, `backend/handlers/backups_empresariales.go`, `backend/handlers/backups_empresariales_test.go`, `web/administrar_empresa/backups.html`, `backend/.env.example`, `backend/main.go`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÒ³n: el mÒ³dulo de backups ahora exporta e importa configuraciÒ³n completa por empresa en JSON canÒ³nico, con restauraciÒ³n sobre PostgreSQL y trazabilidad del origen importado. En la misma iteraciÒ³n se limpian referencias operativas falsas a motor legado retirado del entorno y de comentarios de arranque.
	- VerificaciÒ³n: prueba dirigida del handler de configuraciÒ³n empresarial y diagnÒ³stico del editor sin errores en los archivos tocados.

- PostgreSQL: gobernanza endurecida y limpieza de soporte residual motor legado retirado.
	- Archivos modificados: `.github/agents/agente_go.agent.md`, `.github/agents/agente_backend_db.agent.md`, `copilot-instructions.md`, `backend/db/compat_wrappers.go`, `backend/db/sql_compat.go`, `backend/db/horarios_trabajadores.go`, `.gitignore`, `documentos/descripcion_del_proyecto`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÒ³n: el repositorio formaliza PostgreSQL como Òºnico motor permitido en reglas de agentes e instrucciones del proyecto. TambiÒ©n se elimina el fallback motor legado retirado en memoria del paquete `db`, el dialecto SQL por defecto deja de aceptar motor legado retirado y el esquema de horarios queda solamente en sintaxis PostgreSQL.

- Estaciones: la tarjeta `Notas` ahora soporta mÒºltiples recordatorios persistentes con repeticiÒ³n automÒ¡tica local.
	- Archivos modificados: `web/administrar_empresa/configuracion_de_estaciones.html`, `web/administrar_empresa/estaciones.html`, `web/administrar_empresa/configuracion_carrito_de_compra_empresa.html`, `web/estilos.css`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÒ³n: la estaciÒ³n especial `Notas` deja la lÒ³gica de una sola nota y pasa a manejar varias notas dentro de la misma tarjeta, con selecciÒ³n activa, temporizadores independientes, restauraciÒ³n del countdown tras recargar y repeticiÒ³n automÒ¡tica configurable por minutos. El valor por defecto de repeticiÒ³n entra a `estaciones_config`, mientras el runtime mÒºltiple se persiste localmente por `empresa_id`.
	- VerificaciÒ³n: diagnÒ³stico del editor sin errores en los archivos frontend tocados; revisiÒ³n de QA y backend sin cambios contractuales obligatorios en servidor.

- PostgreSQL: se cierra la Fase 1 de limpieza postmigracion del repositorio.
	- Archivos modificados: `backend/db/pcs_superadministrador`, `documentos/estructura_bd.md`, `documentos/descripcion_de_archivos`, `documentos/erp_multiempresa/02_diseno_tecnico_erp_multiempresa.md`, `Pendiente Notas`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se elimina el ultimo `.db` legacy que seguia versionado en el repo, se confirman las referencias activas y se corrige la documentacion vigente para dejar PostgreSQL como unica base operativa canonica. El backlog deja Fase 1 completada y aterriza las fases 2 y 3.

- Estaciones: se agrega la estaciÒ³n especial Notas y el orden configurable de estaciones especiales.
	- Archivos modificados: `web/administrar_empresa/configuracion_de_estaciones.html`, `web/administrar_empresa/estaciones.html`, `web/administrar_empresa/configuracion_carrito_de_compra_empresa.html`, `web/estilos.css`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÒ³n: el mÒ³dulo de estaciones ahora permite decidir si `Caja`, `YouTube` y la nueva estaciÒ³n especial `Notas` se cargan antes o despuÒ©s del listado normal. `Notas` aÒ±ade una tarjeta operativa con texto editable, temporizador programable, guardado del texto base y alerta visual/sonora cuando vence el recordatorio.
	- VerificaciÒ³n: diagnÒ³stico del editor sin errores en los archivos frontend tocados.

- Documentacion: se depura `Pendiente Notas` y queda como backlog vigente.
	- Archivos modificados: `Pendiente Notas`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÒ³n: se elimina del archivo la mezcla de credenciales y notas obsoletas, se retiran tareas ya implementadas y se reorganiza lo faltante en un plan priorizado con fases de ejecuciÒ³n y criterio de cierre.
	- VerificaciÒ³n: revisiÒ³n documental contra `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md` y el estado actual del repositorio.

- Login administrativo: se corrige la activacion del ojito para mostrar la contraseÒ±a.
	- Archivos modificados: `web/js/login.js`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÒ³n: la inicializaciÒ³n del toggle de visibilidad quedÒ³ insertada en el lugar incorrecto dentro de `showMsg`, asÒ­ que el control podÒ­a aparecer sin activar el cambio real del input. Ahora `initPasswordVisibilityToggles()` se ejecuta en el flujo principal de carga de `login.js`.
	- VerificaciÒ³n: diagnÒ³stico del editor sin errores en `web/js/login.js`.

- Super y portal pÒºblico: el WhatsApp del botÒ³n flotante de la portada ahora es configurable.
	- Archivos modificados: `backend/handlers/usuarios_empresa.go`, `backend/handlers/pagina_principal_handlers.go`, `backend/handlers/pagina_principal_handlers_test.go`, `backend/handlers/system_empresas_handlers_test.go`, `backend/handlers/super_config_backup_handlers.go`, `web/super/configuracion_avanzada.html`, `web/index.html`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÒ³n: se agrega una tarjeta en configuraciÒ³n avanzada para editar el nÒºmero del botÒ³n flotante de WhatsApp de la portada pÒºblica. El valor queda persistido en configuraciÒ³n super y el `index.html` ahora lo consume desde `/api/public/pagina_principal` en vez de mantenerlo fijo en el marcado.
	- VerificaciÒ³n: `go test ./handlers -run "Test(PublicPaginaPrincipalHandlerExposesLandingFields|GmailConfigHandlerSaveWhatsAppContactNumber)$" -count=1`.

- Login administrativo: se agrega control visual para mostrar u ocultar la contraseÒ±a.
	- Archivos modificados: `web/login.html`, `web/js/login.js`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÒ³n: el campo de contraseÒ±a de `/login.html` ahora incluye un botÒ³n tipo ojo que permite alternar la visibilidad del texto escrito, reutilizando el mismo patrÒ³n visual de contraseÒ±as usado en otros formularios del portal.
	- VerificaciÒ³n: inspecciÒ³n estÒ¡tica de `web/login.html` y `web/js/login.js`.

- Portal pÒºblico: la landing descriptiva elimina el hero superior y la tarjeta de accesos rÒ¡pidos.
	- Archivos modificados: `web/descripcion_de_los_sistemas.ht`, `web/estilos.css`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÒ³n: `/descripcion_de_los_sistemas.ht` ahora abre directamente en el contenido detallado de soluciones. Se retiran el bloque `Catalogo unificado` y toda la tarjeta de `Accesos rapidos`, manteniendo el soporte de hash y los CTA pÒºblicos existentes.
	- VerificaciÒ³n: inspecciÒ³n estÒ¡tica de `web/descripcion_de_los_sistemas.ht` y `web/estilos.css`.

- Portal pÒºblico: la landing descriptiva ahora adopta la apariencia activa y un menÒº interno mÒ¡s profesional.
	- Archivos modificados: `web/descripcion_de_los_sistemas.ht`, `web/estilos.css`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÒ³n: la zona `Accesos rapidos` de `/descripcion_de_los_sistemas.ht` deja la grilla plana y pasa a un menÒº ejecutivo con numeraciÒ³n, subtÒ­tulo y estado activo segÒºn la secciÒ³n visible. TambiÒ©n se reemplazan fondos y colores fijos por variables del sistema de apariencia para que la landing siga correctamente el modo claro u oscuro del portal pÒºblico.
	- VerificaciÒ³n: inspecciÒ³n estÒ¡tica de `web/descripcion_de_los_sistemas.ht` y `web/estilos.css`.

- PostgreSQL runtime: inventario avanzado, tablero financiero y salida PEPS quedan operativos por API en Motel Calipso.
	- Archivos modificados: `backend/db/productos.go`, `backend/db/finanzas.go`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se reemplaza SQL no portable del tablero de inventario y finanzas por consultas compatibles con PostgreSQL, se completa la serie diaria de tendencia desde Go, y la ruta de salida de inventario deja de fallar al reordenar el consumo PEPS con wrappers SQL compatibles y cerrar el cursor antes de actualizar lotes en la misma transacción.
	- Verificación: `go test ./db -run "TestRegistrarMovimientoInventario|Test(GetInventarioTendenciaByEmpresaDevuelveSerieDiaria|GetInventarioProyeccionQuiebreByEmpresaPriorizaRiesgo|GetInventarioPlanReposicionByEmpresaConsolidaProveedorYCosto|GetInventarioPlanReposicionResumenByEmpresaAgrupaProveedor|GetEmpresaReportesTableroResumen|GetEmpresaReportesTableroResumenConAsientosCanonicos)$" -count=1`; validación runtime real sobre `empresa_id=7` con `GET /api/empresa/inventario/tendencia`, `GET /api/empresa/inventario/proyeccion_quiebre`, `GET /api/empresa/inventario/plan_reposicion`, `GET /api/empresa/inventario/plan_reposicion_resumen`, `GET /api/empresa/finanzas/movimientosaction=tablero`, `GET /api/empresa/finanzas/movimientosaction=tablero_export&format=json`, `GET /api/empresa/reportesaction=dataset&dataset=empresarial_tablero`, `POST /api/empresa/compras/plan_reposicion/emitir_orden`, `POST /api/empresa/compras/plan_reposicion/actualizar_estado`, `POST /api/empresa/finanzas/cierres_caja`, `PUT /api/empresa/finanzas/cierres_cajaaction=cerrar`, `PUT /api/empresa/finanzas/cierres_cajaaction=aprobar` y `POST /api/empresa/inventario/ajustar` -> `200/201/409 esperado`.

- Reportes Globales (Super): Se habilitó botón de impresión nativa y soporte @media print para reportes globales desde frontend, los cuales ya permitían filtrar empresas del admin (individual/mix), consultar por fechas y exportar sus datasets en JSON, CSV, TXT, Excel XLS y PDF generados en backend.

- Documentos Transaccionales y Flujos: se validan facturación, reportes, eventos, integraciones, y backups empresariales.
	- Archivos involucrados: ackend/handlers/modulos_faltantes.go`n	- Descripción: pruebas exhaustivas automatizadas confirman el cumplimiento legal de DIAN en documentos de notas/facturación, retenciones de PDF/CSV, resolución de conflictos de impresión por impresoras registradas, y rutinas de backup correctas.
	- Verificación: go test ./handlers ./db -run 'TestEmpresaVentasCotizacionesConversionPedidoYDocumentoFinal|TestVentaCarritoFacturaYResolucionImpresora|TestEmpresaFacturacionElectronicaReintentosYReconciliacion' -count=1 -> PASS.

- Créditos y chat/tareas: se cierra la validación PostgreSQL de abonos y citas en runtime real.
	- Archivos modificados: `backend/db/creditos.go`, `backend/db/chat_tareas.go`, `backend/handlers/chat_tareas.go`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el abono de créditos deja de fallar por `driver: bad connection` al separar la lectura de cuotas pendientes de las escrituras dentro de la misma transacción. Además, `chat_tareas_citas` ahora se autorrepara cuando el esquema PostgreSQL legado llega incompleto, y el listado de citas usa la capa compatible para no romper por SQL o tabla faltante.
	- Verificación: `go test ./db ./handlers -run '^$' -count=1`; validación runtime sobre Motel Calipso con `POST /api/empresa/creditosempresa_id=7 -> 201`, `POST /api/empresa/creditosaction=abono&empresa_id=7 -> 200`, `POST /api/empresa/chat_tareas/citasempresa_id=7 -> 201` y `GET /api/empresa/chat_tareas/citasempresa_id=7&q=20260420015859 -> 200`.

- Finanzas y creditos: se corrige compatibilidad PostgreSQL en CRUD generico de cartera y en resumen de cartera de creditos.
	- Archivos modificados: `backend/db/modulos_faltantes.go`, `backend/db/creditos.go`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: `CreateEmpresaGenericRow` deja de usar `LastInsertId` directo y pasa a `insertSQLCompat`, evitando que CxC/CxP persistan pero respondan `400` en PostgreSQL. `GetEmpresaCreditosCarteraResumen` usa `queryRowSQLCompat` y una comparacion de fecha robusta cuando `fecha_vencimiento` viene vacia o legacy, corrigiendo el `500` de `action=resumen_cartera`. Ademas, el flujo de abono de creditos cierra el cursor de cuotas antes del `commit`, evitando `driver: bad connection` en PostgreSQL.
	- Verificación: compilacion `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1` y revalidacion runtime Motel Calipso en curso tras reinicio del backend.

- Red social comercial: PostgreSQL ya persiste y lista publicaciones empresariales/publicas de Motel Calipso.
	- Archivos modificados: `backend/db/red_social.go`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: `empresa_publicaciones_red_social` ahora crea su tabla con DDL compatible para PostgreSQL y todas las lecturas/escrituras pasan por `execSQLCompat` y `querySQLCompat`. Con esto vuelven a funcionar `POST /api/empresa/publicaciones`, `GET /api/empresa/publicaciones` y `GET /api/public/publicaciones` sobre la base PostgreSQL usada por Motel Calipso.
	- Verificación: validacion runtime real contra `http://127.0.0.1:8080` con `empresa_id=7`, creando dos publicaciones comerciales y confirmando respuesta `200` tanto en el feed empresarial como en el feed publico.

- PostgreSQL runtime: se corrige el patrón `LastInsertId` en módulos críticos y se reencamina `/api/empresa/proveedores` al CRUD coherente con compras e inventario.
	- Archivos modificados: `backend/db/auditoria_empresa.go`, `backend/db/usuarios_empresa.go`, `backend/db/clientes.go`, `backend/db/asistencia_empleados.go`, `backend/db/chat_tareas.go`, `backend/db/creditos.go`, `backend/db/finanzas.go`, `backend/db/productos.go`, `backend/db/venta_publica.go`, `backend/db/red_social.go`, `backend/handlers/productos.go`, `backend/main.go`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: las altas que antes respondían con `LastInsertId is not supported by this driver` ahora usan `insertSQLCompat` o `insertTxSQLCompat`, con lo cual PostgreSQL puede devolver `id` vía `RETURNING`. También se asegura el esquema de venta pública antes de guardar configuraci?n, items u ?rdenes y se fuerza la ruta `/api/empresa/proveedores` a usar la tabla `proveedores`, que es la misma validada por productos y compras.
	- Verificación: `go test ./db ./handlers -run '^Test(CreateEmpresaVentaPublicaConfigPersistsTemaVisual|EmpresaVentaPublicaHandlerConfigCatalogoYToggle|EmpresaPublicacionesRedSocialHandler|EmpresaChatTareasCitasSharedByEmpresa|EmpresaCategoriasProductosHandlerCRUD|EmpresaClientesHandler|EmpresaCreditos|EmpresaAsistencia)' -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

## 2026-04-19
- Apariencia global: se reparan los 6 temas, el menú flotante y el guardado automático al iniciar sesión.
	- Archivos modificados: `web/menu.js`, `web/estilos.css`, `web/js/login.js`, `web/js/login_usuario.js`, `web/login_usuario.html`, `web/configuracion_de_la_cuenta.html`, `web/red_social_comercial.html`, `backend/handlers/auth_admin_handlers.go`, `backend/handlers/usuarios_empresa.go`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Archivos creados: `web/Juegos/menu_juegos.html`, `web/Juegos/n64/index.html`.
	- Descripción: el `menu.js` compartido ahora aplica el tema desde el arranque, sincroniza iframes mismo origen, guarda selección en `localStorage`/cookie y refresca la preferencia desde backend cuando existe sesión. Los logins administrativo y de usuario empresa devuelven `apariencia` para fijarla antes de redirigir. Además, vuelve la entrada `Juegos` y se publican rutas funcionales m?nimas para no dejar el enlace roto.
	- Verificación: `get_errors` sobre los archivos tocados y tarea `validar-permisos-selector-empresas-5`.

- Modelo de base de datos ERP (MÃƒÂ¯Ã‚Â¿Ã‚Â½dulo interdependiente de Compras / Proveedores).
        - Archivos modificados: backend/db/compras_y_proveedores.go, backend/db/compras_y_proveedores_test.go, backend/main.go, documentos/descripcion_de_archivos, documentos/historial_de_cambios, documentos/estructura_bd.md, CHANGELOG.md.
        - Descripción: Se crearon las tablas operativas empresa_proveedores, empresa_ordenes_compra, empresa_ordenes_compra_items, y empresa_compras_recepciones para soportar el ciclo de abastecimiento. Se conectÃƒÂ¯Ã‚Â¿Ã‚Â½ EnsureEmpresasComprasSchema al bootstrap del servidor.
        - Verificación: EjecuciÃƒÂ¯Ã‚Â¿Ã‚Â½n local exitosa de tests para esquema relacional.

- Eliminaci?n de páginas hu?rfanas frontend al carecer de uso actual.
	- Archivos modificados: \web/administrar_empresa/bodegas.html\, \web/administrar_empresa/productos/bodegas.html\, \web/administrar_empresa/sensor_puertas_mensajes.html\, \web/administrar_empresa/ventas_simple.html\, \web/super/activar_asesor.html\, \web/super/asesor_comercial.html\, \web/super/vendedor_config_avanzado.html\, \web/ultimas_mejoras.html\ (eliminados), \documentos/descripcion_de_archivos\, \documentos/historial_de_cambios\, \CHANGELOG.md\.
	- Descripción: Se limpiaron del proyecto 8 archivos \.html\ obsoletos o hu?rfanos para reducir el tama?o del repositorio y evitar confusiones en los directorios de operación y super admin.
	- Verificación: No existen referencias de enrutamiento ni clics a estas páginas en los menús o vistas din?micas.

- Juegos globales: ranking público para 3 juegos integrados (Buscaminas, Solitario, Pacman).
	- Archivos modificados: `backend/db/super_juegos.go`, `backend/handlers/super_juegos.go`, `backend/main.go`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `documentos/estructura_bd.md`, `CHANGELOG.md`.
	- Descripción: Se agrega un nuevo modulo para almacenar y consultar records globales de juegos en la base `superadministrador`. La tabla `super_juegos_records` registra `juego`, `nombre_jugador`, `empresa_id` (o 'Publico'), puntaje y nivel. Se exponen los endpoints `GET` y `POST` en `/api/public/juegos/records` para permitir que cualquier jugador (empresa o público) consulte el top 10 o env?e su puntaje desde el frontend.
	- Verificación: Compilaci?n exitosa en `agente_backend_db`. Rutas registradas correctamente en `main.go`.

- Portal público: la landing `Explorar oferta` adopta el estilo de tarjetas del index y una est?tica propia m?s comercial.
	- Archivos modificados: `web/descripcion_de_los_sistemas.ht`, `web/estilos.css`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la página descriptiva de ofertas deja el bloque visual oscuro anterior y pasa a reutilizar el estilo `home-offer-card` del index en sus tarjetas ampliadas. El hero y la navegaci?n r?pida tambi?n se refinan para que la landing se perciba como extensi?n natural del portal principal, manteniendo los mismos enlaces seguros de `Probar Gratis`.
	- Verificación: `get_errors` sin errores en `web/descripcion_de_los_sistemas.ht`, `web/estilos.css`, `documentos/descripcion_de_modulos` y `documentos/matriz_roles_permisos_pos_multiempresa.md`.

- Chat y tareas: selección m?ltiple de usuarios, fotos y validación estricta de empresa.
	- Archivos modificados: `backend/handlers/chat_tareas.go`, `backend/handlers/chat_tareas_test.go`, `web/administrar_empresa/chat_y_tareas.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: `chat_y_tareas.html` ahora permite buscar y marcar uno o varios usuarios activos de la misma empresa para crear chats directos o grupales y agregar varios participantes a una conversación existente. El formulario de mensajes deja explícito el env?o de fotos adem?s de audio y documentos. En paralelo, `chat_tareas.go` valida que cada participante tipo `usuario` pertenezca realmente a la empresa antes de persistirlo, bloqueando cruces entre empresas en creación grupal o agregado posterior.
	- Verificación: `get_errors` sin errores en `backend/handlers/chat_tareas.go`, `backend/handlers/chat_tareas_test.go`, `web/administrar_empresa/chat_y_tareas.html` y `web/estilos.css`; `go test ./handlers -run '^TestEmpresaChatTareas(AdjuntoUploadAllowsImage|AdjuntoUploadAllowsDocx|ConversacionesAddsOwnerAdminParticipant|ConversacionesCreatesGrupoConUsuariosSeleccionados|ConversacionesRejectsUsuarioDeOtraEmpresa|ParticipantesRejectsUsuarioDeOtraEmpresa|CitasSharedByEmpresa|MensajesHandlerDerivesUsuarioActor|MensajesHandlerRejectsInvalidConversacion|TareasHandlerRejectsInvalidConversacion|CitasHandlerRejectsInvalidConversacion)$' -count=1`.

- Panel empresa y chat/tareas: el módulo colaborativo se refuerza como dashboard principal y valida referencias empresariales antes de persistir.
	- Archivos modificados: `backend/handlers/chat_tareas.go`, `backend/handlers/chat_tareas_test.go`, `web/administrar_empresa/chat_y_tareas.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: `chat_y_tareas.html` ahora se comporta como home operativo del panel empresa con tarjetas resumen, acciones r?pidas y estados vac?os guiados. En paralelo, `chat_tareas.go` valida que conversaciones y tareas referenciadas existan dentro de la empresa antes de crear participantes, mensajes, tareas, citas o notas de voz, y limpia archivos de adjuntos si falla la persistencia posterior para no dejar hu?rfanos.
	- Verificación: `get_errors` sin errores en `backend/handlers/chat_tareas.go`, `backend/handlers/chat_tareas_test.go`, `web/administrar_empresa/chat_y_tareas.html` y `web/estilos.css`; `go test ./handlers -run '^TestEmpresaChatTareas(MensajesHandlerDerivesUsuarioActor|AdjuntoUploadAllowsDocx|ConversacionesAddsOwnerAdminParticipant|CitasSharedByEmpresa|MensajesHandlerRejectsInvalidConversacion|TareasHandlerRejectsInvalidConversacion|CitasHandlerRejectsInvalidConversacion)$' -count=1`.

- Autenticaci?n administrativa: `login.html` vuelve a ofrecer `Recordar usuario` para el acceso por correo.
	- Archivos modificados: `web/login.html`, `web/js/login.js`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el login administrativo por correo vuelve a incluir una casilla visible `Recordar usuario`. Cuando se marca, el frontend conserva solo el correo del administrador en `localStorage`; si no se marca, la identidad recordada se limpia. El cambio no altera la sesion real, el flujo de Google, permisos ni wrappers de autenticacion.
	- Verificación: `get_errors` sin errores en `web/login.html`, `web/js/login.js`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos` y `documentos/matriz_roles_permisos_pos_multiempresa.md`.

- Portal público: el home resalta la marca principal y unifica el pie de página con la nueva leyenda corporativa.
	- Archivos modificados: `web/index.html`, `web/estilos.css`, `web/login.html`, `web/registrar_nuevo_usuario_administrador.html`, `web/ayuda/ayuda.html`, `web/descripcion_de_los_sistemas.ht`, `web/Informacion_de_contacto.html`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el botón `Registrarse o iniciar sesión` del portal principal adopta una apariencia tipo tarjeta con texto reforzado, y el título `Sistema de Facturaci?n Electr?nica` queda destacado sobre una barra semitransparente para darle m?s presencia visual. Además, se reemplaza la leyenda visible anterior por `@ 2026 - Powerful Control System - Sistema POS Multiempresa` en los pies públicos afectados y se actualizan los títulos de páginas que todavía cargaban la marca `POS Multiempresa` como nombre principal.
	- Verificación: `get_errors` sin errores en `web/index.html`, `web/estilos.css`, `web/login.html`, `web/registrar_nuevo_usuario_administrador.html`, `web/ayuda/ayuda.html`, `web/descripcion_de_los_sistemas.ht` y `web/Informacion_de_contacto.html`.

- Panel empresa: Chat y tareas pasa a ser la vista inicial y su calendario compartido queda en primer plano.
	- Archivos modificados: `web/administrar_empresa.html`, `web/js/administrar_empresa.js`, `web/administrar_empresa/chat_y_tareas.html`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: `administrar_empresa.html` ahora abre preferentemente `chat_y_tareas.html` al entrar y mueve ese acceso al primer lugar del menú. La página colaborativa sube el calendario mensual al inicio, refuerza la explicaci?n de agenda compartida por empresa y ampl?a el peso visual del tablero para que la administradora registre reuniones y los dem?s usuarios autorizados las consulten desde su cuenta.
	- Verificación: `get_errors` sin errores en `web/administrar_empresa.html`, `web/js/administrar_empresa.js`, `web/administrar_empresa/chat_y_tareas.html` y `web/estilos.css`; `go test ./handlers -run '^TestEmpresaChatTareasCitasSharedByEmpresa$' -count=1`.

- Venta pública y pagos: el cat?logo público deja de fallar con error interno sobre esquemas legacy.
	- Archivos modificados: `backend/db/venta_publica.go`, `backend/handlers/venta_publica_test.go`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el backend ahora autorrepara columnas faltantes de configuracion, items y ordenes en `empresa_venta_publica_*` antes de consultar la tienda pública. Con esto, `GET /api/public/venta_publicaaction=catalogo` deja de caer en instalaciones legado que no ten?an campos evolutivos como `tema_visual`, `estado`, `destacado`, `stock_publicado` o payloads de orden. Se agrega adem?s una regresi?n espec?fica para cat?logo público sobre esquema legacy.
	- Verificación: `go test ./db -run '^TestEmpresaVentaPublicaConfigPersistsTemaVisual$' -count=1`; `go test ./handlers -run '^Test(PublicVentaPublicaHandlerCatalogoWithLegacySchemaMissingColumns|PublicVentaPublicaHandlerCatalogoYPagoConWompiInactivo|EmpresaVentaPublicaHandlerConfigCatalogoYToggle)$' -count=1`; `get_errors` sin errores en `backend/db/venta_publica.go` y `backend/handlers/venta_publica_test.go`.

- Panel empresa: reservas endurece sus queries y evita URLs infladas que pod?an terminar en error 414.
	- Archivos modificados: `web/administrar_empresa/reservas_hotel.html`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la vista de reservas ahora resuelve `empresa_id` desde el contexto activo del panel, limita el texto de búsqueda antes de enviarlo por query string y centraliza la construcci?n de URLs del módulo para no propagar par?metros largos o innecesarios. Además, la tabla deja de incrustar el JSON completo de cada reserva en atributos HTML y usa el estado local para editar, reduciendo el tama?o del DOM y evitando crecimiento accidental de la navegaci?n.
	- Verificación: `get_errors` sin errores en `web/administrar_empresa/reservas_hotel.html`.

- Panel empresa: administrar productos deja de caer por la carga obligatoria de proveedores en la vista principal.
	- Archivos modificados: `web/administrar_empresa/administrar_productos.html`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la vista `Administrar Productos` ya no arranca consultando proveedores de forma obligatoria cuando el modo visible es `productos`. El selector de proveedor principal pasa a cargarse de forma perezosa al abrir el formulario del producto, de modo que una falla o restricci?n del submódulo de compras/proveedores no tumba toda la pantalla principal de inventario.
	- Verificación: `get_errors` sin errores en `web/administrar_empresa/administrar_productos.html`.

- Panel empresa: se corrige la duplicaci?n recursiva del submenú de productos y dem?s shells anidados.
	- Archivos modificados: `web/js/administrar_empresa.js`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el script compartido del panel empresa ahora detecta los enlaces y el iframe del shell actual antes de resolver la página inicial. Con esto, los submenús internos como `Productos` dejan de cargarse a sí mismos dentro de su propio iframe y pasan a abrir su contenido real, evitando el efecto de menú duplicado.
	- Verificación: `get_errors` sin errores en `web/js/administrar_empresa.js`; validación est?tica en `http://127.0.0.1:8091/administrar_empresa/administrar_productos_menu.htmlempresa_id=1` con snapshot mostrando el iframe apuntando a `administrar_productos.htmlview=productos&empresa_id=1` en lugar de recargar `administrar_productos_menu.html`.

- Frontend web: se endurece la adaptaci?n autom?tica a celular sin alterar la estructura de menús.
	- Archivos modificados: `web/estilos.css`, `web/super/administrar_base_de_datos.html`, `web/super/errores.html`, `web/super/seguridad.html`, `web/super/soporte_remoto.html`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la hoja compartida ahora fuerza en m?vil el apilado real de filas de formulario, toolbars y filtros, neutraliza anchos m?nimos e inline widths problem?ticos en controles, y mejora el comportamiento de tablas y bloques con scroll horizontal cuando hace falta. Además, las vistas super de base de datos, errores, seguridad VPS y soporte remoto reciben media queries puntuales para evitar desbordes y columnas r?gidas en celular sin tocar los menús existentes.
	- Verificación: `get_errors` sin errores en `web/estilos.css`, `web/super/administrar_base_de_datos.html`, `web/super/errores.html`, `web/super/seguridad.html` y `web/super/soporte_remoto.html`.

- Estaciones y carritos: se reutilizan carritos legado al abrir una estación y se evita el error de carga por duplicado.
	- Archivos modificados: `web/administrar_empresa/carrito_de_compras.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: al entrar desde `estaciones.html`, el carrito unificado ya no asume que siempre existe un carrito con el código can?nico `EST-empresa-estacion`. Ahora intenta reutilizar primero un carrito ya existente por código, `referencia_externa=ESTACION_<id>` o nombre de estación antes de crear uno nuevo, evitando conflictos con datos legado que antes terminaban en `Error cargando carritos`.
	- Verificación: `get_errors` sobre `web/administrar_empresa/carrito_de_compras.html` y validación dirigida del flujo de estación.

- Estaciones: la estación especial de YouTube cambia a un visor embebido estable con búsqueda funcional.
	- Archivos modificados: `web/administrar_empresa/youtube_station_browser.html`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se elimina la dependencia de la API del reproductor que no estaba resolviendo bien la carga en la tarjeta y se reemplaza por un visor embebido basado en `youtube-nocookie` que sí permite mostrar resultados y lanzar búsquedas desde la propia estación. Se agrega un botón `Inicio` para volver a una portada embebida ?til y se conserva `Abrir YouTube` para abrir la página real completa en otra pesta?a, porque el home oficial de YouTube no se puede incrustar de forma fiable dentro de un iframe normal.
	- Verificación: `get_errors` sin errores en `web/administrar_empresa/youtube_station_browser.html`.

- Super administrador: se agrega un panel real para formatos de email y se unifica el guardado de configuraci?n avanzada.
	- Archivos creados: `backend/handlers/super_email_templates.go`, `web/super/formato_para_emviar_email.html`.
	- Archivos modificados: `backend/main.go`, `backend/handlers/auth_admin_handlers.go`, `backend/handlers/usuarios_empresa.go`, `backend/handlers/payments_handlers.go`, `backend/handlers/server_runtime_notifications.go`, `backend/handlers/system_empresas_handlers_test.go`, `web/super/configuracion_avanzada.html`, `web/super_administrador.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el panel super ahora expone `/super/formato_para_emviar_email.html` para administrar plantillas reales de confirmación de correo, activaci?n por pago de licencia, recuperaci?n de contrase?a y alertas de reinicio. El backend centraliza esas plantillas en `/super/api/config/email_templates` y reemplaza textos hardcodeados en correos administrativos, usuarios de empresa, licencias y monitoreo del servidor. Además, `configuracion_avanzada.html` deja botones sueltos por bloque y pasa a guardar Wompi, Epayco, Gmail e IA con un solo botón arriba y otro abajo de la página.
	- Verificación: `go test ./handlers -run 'Test(SuperEmailTemplatesHandlerSaveAndGet|ApplySuperEmailTemplateUsesConfiguredValues|GmailConfigHandlerTestActionCapturesNotification)$' -count=1`; tarea `validar-permisos-selector-empresas-5`; `get_errors` sin errores en `backend/handlers/super_email_templates.go`, `web/super/formato_para_emviar_email.html` y `web/super/configuracion_avanzada.html`.

- Portal público y selector de empresas: se corrigen CTA públicos y se oculta navegaci?n global fuera del alcance super principal.
	- Archivos modificados: `web/descripcion_de_los_sistemas.ht`, `web/js/seleccionar_empresa.js`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la landing descriptiva ya no reutiliza destinos privados como `/administrar_empresa.html` en el CTA `Probar Gratis`; cuando una tarjeta apunta a una ruta protegida, el flujo público redirige al registro de administrador. En el selector de empresas, el menú lateral ahora toma el perfil real desde `/me` y solo mantiene visibles `Administradores` y `Reportes globales` para cuentas super principales; la navegaci?n sensible queda oculta por defecto hasta resolver la sesión.
	- Verificación: `get_errors` sin errores en `web/descripcion_de_los_sistemas.ht`, `web/js/seleccionar_empresa.js` y `web/seleccionar_empresa.html`; sondeo runtime local con `200` en `/descripcion_de_los_sistemas.ht` y `/registrar_nuevo_usuario_administrador.html`, y `401` en `/seleccionar_empresa.html` sin sesión; `go test ./ ./auth ./db ./handlers ./metrics ./utils -count=1`.

- Integridad super/licencias: se endurece el alcance delegado, se recupera compatibilidad de backup legacy y se corrige la validación pública de métodos de pago.
	- Archivos modificados: `backend/db/chat_inteligencia_artificial.go`, `backend/handlers/system_empresas_handlers.go`, `backend/handlers/postgres_performance.go`, `backend/handlers/super_config_backup_handlers.go`, `backend/handlers/payments_handlers.go`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: los administradores delegados ya no heredan acceso global por tener rol `super_administrador`, sino que quedan restringidos al portafolio del administrador principal. El backup/restore super vuelve a aceptar claves sensibles legacy de IA para restauraciones compatibes, `GET /api/public/licencias/payment_methods` puede anunciar Epayco cuando existe `public_key`, y el panel de rendimiento PostgreSQL valida primero la acci?n solicitada para devolver `400` en acciones no soportadas aunque el runtime no est? sobre PostgreSQL.
	- Verificación: `go test ./handlers -run 'Test(PostgresPerformanceHandlerUnknownAction|EmpresasHandlerFiltraEmpresasPorAdministradorPrincipal|SuperConfigBackupHandlerRestoreEncryptsSensitivePlaintext|PublicLicenciasPaymentMethodsHandlerAllowsEpaycoWithPublicKeyOnly|SuperConfigBackupHandlerExportYRestore|PublicLicenciasPaymentMethodsHandlerOrdersAndAvailability)$' -count=1`; tarea `validar-permisos-selector-empresas-5`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -count=1`; sondeo runtime local de `/`, `/index.html`, `/login.html`, `/api/public/pagina_principal`, `/api/public/licencias/payment_methods` y `/seleccionar_empresa.html`.

## 2026-04-18

- Estaciones: la tarjeta de YouTube ahora permite guardar la fuente desde el mismo bloque y acepta Shorts.
	- Archivos modificados: `web/administrar_empresa/estaciones.html`, `web/estilos.css`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la propia tarjeta de YouTube en `estaciones.html` ahora incluye un campo para pegar la URL o el ID del contenido, un botón `Guardar y cargar` y un enlace externo alineado con el valor actual. El navegador interno ya interpreta tambi?n URLs de `Shorts` como video v?lido y recarga la vista sin obligar a entrar a la configuraci?n general de estaciones.
	- Verificación: `get_errors` sin errores en `web/administrar_empresa/estaciones.html` y `web/estilos.css`.

- Estaciones: la tarjeta de YouTube deja de depender de búsquedas embebidas rotas y pasa a reproducir referencias v?lidas.
	- Archivos modificados: `web/administrar_empresa/youtube_station_browser.html`, `web/administrar_empresa/estaciones.html`, `web/administrar_empresa/configuracion_de_estaciones.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la estación especial de YouTube ya no intenta incrustar resultados de búsqueda que el proveedor externo no soporta. Ahora reproduce solo URLs o IDs v?lidos de video/playlist mediante `youtube-nocookie`, muestra la referencia configurada dentro de la tarjeta y, cuando el valor guardado es texto libre, deja un estado visible con fallback a `Abrir YouTube` fuera del sistema.
	- Verificación: `get_errors` sin errores en `web/administrar_empresa/youtube_station_browser.html`, `web/administrar_empresa/estaciones.html`, `web/administrar_empresa/configuracion_de_estaciones.html` y `web/estilos.css`.

- Estaciones y carritos: se mejora el mensaje visible cuando falla la apertura del carrito por estación.
	- Archivos modificados: `web/administrar_empresa/carrito_de_compras.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el carrito unificado ya no deja al usuario solo con `Error cargando carritos` al fallar la carga inicial. Ahora muestra un estado contextual con traducci?n segura del error, botón `Reintentar carga` y retorno explícito a `estaciones.html` cuando el flujo viene desde una estación, evitando exponer literales t?cnicos del backend como `unauthenticated` o `forbidden`.
	- Verificación: `get_errors` sin errores en `web/administrar_empresa/carrito_de_compras.html`; validación est?tica de apertura en `http://127.0.0.1:8080/administrar_empresa/carrito_de_compras.htmlempresa_id=6&estacion_id=1&estacion_nombre=Estacion%201&carrito_codigo=EST-6-1`.

- Estaciones y carritos: se corrige la carga del carrito por estación sobre PostgreSQL real.
	- Archivos modificados: `backend/db/carritos_compras.go`, `backend/handlers/carritos_compras.go`, `backend/db/carritos_inventario_test.go`, `backend/handlers/auth_users_carritos_test.go`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el listado de `/api/empresa/carritos_compra` deja de depender de un `GROUP BY` fr?gil con cliente e items y ahora cuenta items desde un agregado previo por carrito. Además, `totales_pago` y `metricas_estacion` dejan de usar `ROUND(..., 2)` en SQL y redondean en Go, evitando fallos de compatibilidad entre motor legado retirado y PostgreSQL. Esto elimina el error visible `Error cargando carritos` al abrir una estación y estabiliza los totales del panel.
	- Verificación: `go test ./db -run 'Test(GetCarritosCompraByEmpresaFallbackWithoutClientesSchema|GetCarritosCompraByEmpresaCountsItemsAndClientName|SyncEmpresaEstacionCarritosCreatesAndUpdatesLinkedDefaults)$' -count=1`; `go test ./handlers -run 'Test(EmpresaCarritosCompraMetricasEstacionIncluyeCorrecciones|EmpresaCarritosCompraTotalesPagoAgrupaYRedondea|EmpresaCarritosCompraRejectsActivarEstacionPagadaSinReset|EmpresaCarritosCompraRecuperarInterrumpidoConAuditoria|EmpresaEstacionPrefsHandler_UpsertAndIsolationByEmpresa|WithEmpresaVentasPermissionsDeniesOutOfScopeEmpresa|WithEmpresaVentasPermissionsBloqueaModuloNoHabilitadoPorLicencia)$' -count=1`.

- Carritos y estaciones: se limpia el legado de `ventas_simple` y se ampl?an los checks del carrito unificado.
	- Archivos modificados: `web/administrar_empresa/carrito_de_compras.html`, `web/administrar_empresa/configuracion_de_estaciones.html`, `web/administrar_empresa/configuracion_carrito_de_compra_empresa.html`, `backend/handlers/empresa_configuracion_general.go`, `backend/handlers/empresa_configuracion_general_test.go`, `backend/handlers/empresa_estacion_prefs_test.go`, `backend/db/empresa_configuracion_general.go`, `backend/db/empresa_estacion_prefs_test.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Archivos eliminados: `web/js/ventas_simple.js`.
	- Descripción: el carrito unificado ahora permite configurar tambi?n la visibilidad del cliente, el bloque de descuento/impuestos por item, el resumen total del carrito y el desglose del cobro. Además, se retira del runtime el script antiguo de `ventas_simple` y se elimina el bloque muerto de carrito compacto en `configuracion_general`.
	- Verificación: pruebas dirigidas de handlers/db del modulo y validacion funcional del redirect de compatibilidad.

- IA: el proyecto migra a Google Gemini como ?nico proveedor y retira Ollama/DeepSeek del flujo operativo.
	- Archivos modificados: `backend/handlers/ai_credentials_catalog.go`, `backend/handlers/ai_config_handlers.go`, `backend/handlers/chat_con_inteligencia_artificial_controller.go`, `backend/handlers/chat_con_ia_global_super_test.go`, `backend/handlers/chat_con_inteligencia_artificial_controller_test.go`, `backend/handlers/system_empresas_handlers_test.go`, `backend/db/chat_inteligencia_artificial_test.go`, `backend/tools/set_ai_provider_enabled.go`, `backend/.env.example`, `backend/.env.local`, `scripts/iniciar_servidor.ps1`, `web/super/configuracion_avanzada.html`, `web/super/chat_con_ia_global.html`, `web/administrar_empresa/chat_con_inteligencia_artificial.html`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_archivos`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Archivos eliminados: `backend/tools/check_deepseek_key/main.go`.
	- Descripción: se reemplaza la IA en VPS basada en Ollama por Google Gemini en el chat global super y el chat con IA por empresa. Configuraci?n avanzada queda reducida a una sola API key cifrada, un interruptor global y un interruptor del proveedor `google`. El script de arranque ya no abre t?nel a Ollama y el VPS queda sin `ollama.service` ni binario asociado.
	- Verificación: `go test ./handlers -run 'Test(SuperAIModelosHandlerReturnsCatalog|SuperAIModelosHandlerFiltersDisabledProvider|ModelosHandlerReturnsPreferredModelForGoogleAccount|ModeloPreferidoHandlerAcceptsGemini|ModelosHandlerFiltersDisabledProvider|AIModelsConfigHandlerSaveGeminiEncrypted|AIModelsConfigHandlerSavesProviderEnabledState|AIModelsConfigHandlerTogglesGlobalServiceState)$' -count=1`; `go test ./db -run 'Test(EmpresaAIModeloPreferidoUpsertAndGet|RegisterEmpresaAIConsultaAcumulaUsoDiario|SuperAIModeloPreferidoUpsertAndGet|RegisterSuperAIConsultaAcumulaUsoDiario|GetSuperAIModeloPreferidoRepairsMissingSchema|GetSuperAIUsoDiarioRepairsMissingSchema|RegisterSuperAIConsultaRepairsMissingSchema)$' -count=1`.

- Carritos y estaciones: el sistema adopta un carrito unificado configurable por empresa y por estación.
	- Archivos modificados: `web/administrar_empresa/carrito_de_compras.html`, `web/administrar_empresa/estaciones.html`, `web/administrar_empresa/ventas_simple.html`, `web/administrar_empresa/configuracion_de_estaciones.html`, `web/administrar_empresa/configuracion_carrito_de_compra_empresa.html`, `web/administrar_empresa/configuracion_menu.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: desaparece la bifurcaci?n de UI entre carrito de compras, venta simple y carrito compacto. Las estaciones abren siempre `carrito_de_compras.html` y la pantalla muestra u oculta bloques seg?n configuraci?n global por empresa y configuraci?n individual por estación almacenadas en `estaciones_config`.
	- Verificación: tarea `validar-permisos-selector-empresas-5`; `get_errors` sin errores en los archivos frontend modificados.


- Chat con IA: interfaz simplificada en empresa y panel super.
	- Archivos modificados: `web/administrar_empresa/chat_con_inteligencia_artificial.html`, `web/super/chat_con_ia_global.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se retiran botones superiores, cuadros visibles de modelo/uso diario, selector visible de modelo, CTA de upgrade y panel de consultas recientes; las sugerencias rapidas pasan debajo del chat y `Limpiar chat` queda junto a `Preguntar a la IA`.
	- Verificación: `get_errors` sobre las vistas y documentos modificados.

- Panel empresa: se retira la calculadora del menu lateral y del menu flotante.
	- Archivos modificados: `web/administrar_empresa.html`, `web/js/administrar_empresa.js`, `web/menu.js`, `web/ayuda/ayuda.html`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Archivos eliminados: `web/administrar_empresa/calculadora.html`.
	- Descripción: se elimina el boton `Calculadora`, se retira el acceso rapido del menu flotante y se borra la pagina frontend asociada para que deje de existir como frente visible del panel empresa.
	- Verificación: `get_errors` y busqueda de referencias frontend activas a `calculadora.html`.

- Inventario y productos: las compras preventivas y por proveedor pasan a la secci?n `Compras` del submódulo.
	- Archivos modificados: `web/administrar_empresa/administrar_productos.html`, `web/administrar_empresa/productos/compras.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la vista central del módulo ahora expone `view=compras` y concentra all? el plan de reposici?n por proveedor, el consolidado de compra y el borrador/ciclo de orden; la ruta `productos/compras.html` deja el placeholder y redirige a esa vista real.
	- Verificación: `get_errors` sobre los HTML y documentos modificados.

- Panel empresa: se elimina la página `Inicio` y el shell arranca directo en Productos.
	- Archivos eliminados: `web/administrar_empresa/inicio.html`.
	- Archivos modificados: `web/administrar_empresa.html`, `web/administrar_empresa/administrar_productos_menu.html`, `web/js/administrar_empresa.js`, `web/ayuda/ayuda.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se retira la portada intermedia del panel empresa, desaparece el botón `Inicio` del menú principal y del submódulo de productos, y la carga inicial del iframe pasa a `administrar_productos_menu.html`.
	- Verificación: `get_errors` sobre los archivos HTML, JS y documentos modificados.

- Inventario y productos: `Proveedores` pasa a una subpágina dedicada y `Precios` muestra el historial real de cambios de precio.
	- Archivos creados: `web/administrar_empresa/productos/administrar_proveedores.html`.
	- Archivos modificados: `web/administrar_empresa/administrar_productos.html`, `web/administrar_empresa/administrar_productos_menu.html`, `web/administrar_empresa/productos/administrar_productos_menu.html`, `web/administrar_empresa/productos/precios.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la vista principal del módulo ya no mezcla proveedores ni el historial de cambios de precio con el CRUD de productos; ambos salen a subvistas dedicadas reutilizando la misma página central y preservando `empresa_id` dentro del shell administrativo.
	- Verificación: `get_errors` sobre los HTML y documentos modificados.

- Chat IA super y empresarial: se autorrepara el esquema legacy `super_ai_*` y `empresa_ai_*` y se ampl?a el timeout de Ambis Local sobre el t?nel VPS.
	- Archivos modificados: `backend/db/chat_inteligencia_artificial.go`, `backend/db/chat_inteligencia_artificial_test.go`, `backend/handlers/chat_con_inteligencia_artificial_controller.go`, `documentos/descripcion_de_modulos`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el chat super ya no falla al consultar modelo preferido o uso diario sobre instalaciones PostgreSQL heredadas; la capa DB repara tablas/columnas faltantes al vuelo y el cliente de Ollama ahora soporta tiempos de respuesta m?s largos de `codellama:7b` cuando la consulta viaja por el t?nel local al VPS.
	- Verificación: `go test ./db -run 'Test(SuperAIModeloPreferidoUpsertAndGet|RegisterSuperAIConsultaAcumulaUsoDiario|GetSuperAIModeloPreferidoRepairsMissingSchema|GetSuperAIUsoDiarioRepairsMissingSchema|RegisterSuperAIConsultaRepairsMissingSchema)$' -count=1`; `go test ./handlers -run '^$' -count=1`; `curl http://localhost:8080/super/api/chat_con_ia_global/modelos` -> `200 OK`; `curl http://localhost:8080/super/api/chat_con_ia_global/consultar` -> `200 OK`.

- IA super y empresarial: se agregan switches por proveedor para desactivar DeepSeek sin afectar Ambis Local.
	- Archivos modificados: `backend/handlers/ai_credentials_catalog.go`, `backend/handlers/ai_config_handlers.go`, `backend/handlers/chat_con_inteligencia_artificial_controller.go`, `backend/handlers/chat_con_ia_global_super.go`, `backend/handlers/system_empresas_handlers_test.go`, `backend/handlers/chat_con_inteligencia_artificial_controller_test.go`, `backend/handlers/chat_con_ia_global_super_test.go`, `web/super/configuracion_avanzada.html`, `documentos/descripcion_de_modulos`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: configuraci?n avanzada ahora puede habilitar o bloquear por separado `DeepSeek Chat` y `Ambis Local`; los chats empresarial y global super solo muestran proveedores activos y hacen fallback automático cuando el modelo preferido qued? apagado.
	- Verificación: `go test ./handlers -run 'Test(AIModelsConfigHandlerSaveDeepSeekEncrypted|AIModelsConfigHandlerSavesProviderEnabledState|ModelosHandlerFiltersDisabledProvider|ModelosHandlerRejectsWhenAIDisabled|SuperAIModelosHandlerFiltersDisabledProvider|SuperAIModelosHandlerRejectsWhenAIDisabled)$' -count=1`.

- Portal público: se elimina el arcade y solo queda el emulador N64 adaptado para m?vil.
	- Archivos eliminados: `web/Juegos/ajedrez_3d_plus.html`, `web/Juegos/ajedrez_vs_ia_plus.html`, `web/Juegos/arcade_shared.js`, `web/Juegos/arcade_window.css`, `web/Juegos/brigada_burbujas_3d_plus.html`, `web/Juegos/carton_fire_plus.html`, `web/Juegos/memoria_estelar.html`, `web/Juegos/memoria_estelar_plus.html`, `web/Juegos/pacman_plus.html`, `web/Juegos/patito_volando.html`, `web/Juegos/patito_volando_plus.html`, `web/Juegos/rebote_bloques.html`, `web/Juegos/rebote_bloques_plus.html`, `web/Juegos/serpiente_pixel.html`, `web/Juegos/serpiente_pixel_plus.html`, `web/Juegos/tetris_plus.html`, `web/Juegos/nes/README.md`, `web/Juegos/nes/index.html`, `web/Juegos/nes/nes-wrapper.js`, `web/Juegos/nes/styles.css`, `web/img/juegos/ajedrez_3d.svg`, `web/img/juegos/ajedrez_vs_ia.svg`, `web/img/juegos/brigada_burbujas_3d.svg`, `web/img/juegos/carton_fire.svg`, `web/img/juegos/memoria_estelar.svg`, `web/img/juegos/pacman.svg`, `web/img/juegos/patito_volando.svg`, `web/img/juegos/rebote_bloques.svg`, `web/img/juegos/serpiente_pixel.svg`, `web/img/juegos/tetris.svg`.
	- Archivos modificados: `web/Juegos/menu_juegos.html`, `web/Juegos/n64/index.html`, `web/Juegos/n64/n64-wrapper.js`, `web/menu.js`, `backend/handlers/auth_users_carritos_test.go`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el proyecto retira todas las experiencias de juego y el emulador NES. El portal conserva solo el emulador N64 con acceso público y dise?o m?vil, y el menú global enlaza de forma directa a esa ?nica experiencia.
	- Verificación: búsqueda global en `web/**` y `backend/**` sin referencias funcionales a los juegos eliminados.

- Selector de empresas: las tarjetas quedan alineadas a la izquierda y conservan un tama?o visual uniforme.
	- Archivos modificados: `web/estilos.css`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se elimina el centrado del grid en `seleccionar_empresa.html` y se fija una relaci?n de aspecto com?n para que todas las tarjetas se vean del mismo tama?o dentro del selector.
	- Verificación: `get_errors` sin errores en `web/estilos.css`.

- Portal público: el menú flotante deja de mostrar accesos r?pidos de juegos y centraliza la navegaci?n en la página del arcade.
	- Archivos eliminados: `web/Juegos/games.json`.
	- Archivos modificados: `web/menu.js`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se retira el manifiesto JSON que alimentaba el menú flotante y se deja un ?nico enlace a `/Juegos/menu_juegos.html`, evitando duplicar la lista de juegos fuera del lobby principal.
	- Verificación: búsqueda global sin referencias activas a `games.json` en `web/**`.

- Arcade público: se agrega `N64 vertical mobile` para jugar desde celular con ROM legal del usuario.
	- Archivos creados: `web/Juegos/n64/index.html`, `web/Juegos/n64/styles.css`, `web/Juegos/n64/n64-wrapper.js`.
	- Archivos modificados: `web/Juegos/menu_juegos.html`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la página N64 deja de ser un scaffold remoto y pasa a cargar `n64js` dentro de `iframe.srcdoc` mismo-origen para que el m?vil pueda enviar controles t?ctiles reales al core. La ROM legal del usuario se persiste en IndexedDB y los botones `Guardar` y `Cargar` respaldan/restauran la memoria del cartucho por `rominfo.id`, ?til para títulos como Super Mario 64 cuando se guarda primero dentro del propio juego.
	- Verificación: `get_errors` sin errores en `web/Juegos/n64/index.html`, `web/Juegos/n64/styles.css`, `web/Juegos/n64/n64-wrapper.js` y `web/Juegos/menu_juegos.html`.

- Licencias: el checkout cambia a activacion sin pasarela cuando el total queda en cero y bloquea la repeticion gratuita por empresa.
	- Archivos creados: `backend/db/licencias_gratis.go`.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `backend/main.go`, `web/pagar_licencia.html`, `web/elegir_licencia.html`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el checkout de licencias ahora calcula descuento y total real antes de abrir la pasarela; si el valor base es cero o el código deja el total en cero, ofrece `Activar licencia` en lugar de cobrar. Además, una licencia gratis solo puede implementarse una vez por empresa y el selector de empresas mantiene tarjetas de altura uniforme.
	- Verificación: `go test ./handlers -run 'Test(LicenciaCheckoutSummaryHandlerAllowsZeroTotalByConfiguredDiscount|ActivateLicenciaSinPagoHandlerBlocksRepeatedFreeLicensePerEmpresa|WompiCreateNequiTransactionHandlerRejectsZeroTotalAndSuggestsActivation)$' -count=1` y tarea `validar-permisos-selector-empresas-5`.

- Inventario y productos: se separan las subpaginas de `productos`, `bodegas` y `categorias` sin duplicar la logica del modulo.
	- Archivos modificados: `web/administrar_empresa/administrar_productos.html`, `web/administrar_empresa/administrar_productos_menu.html`, `web/administrar_empresa/productos/administrar_productos.html`, `web/administrar_empresa/productos/bodegas.html`, `web/administrar_empresa/productos/categorias.html`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la vista principal del modulo ahora acepta `view=productos|bodegas|categorias`, las rutas legacy de `productos/` quedan como wrappers que conservan `empresa_id` y el menu lateral carga cada frente en su vista dedicada sin tocar endpoints ni contratos backend.
	- Verificación: `get_errors` sobre los HTML y documentos modificados.

- Operacion por empresa: `ventas_simple.html` agrega la variante `carrito_compacto` en el mismo panel por estacion.
	- Archivos modificados: `web/administrar_empresa/estaciones.html`, `web/administrar_empresa/ventas_simple.html`, `web/js/ventas_simple.js`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: las estaciones con venta simple ahora abren el mismo flujo `ventas_simple.html` en modo `carrito_compacto`, con barra de acciones rapidas, total visible y acceso directo a busqueda, carrito, cobro, sincronizacion, nueva venta y correccion; la variante conserva los mismos endpoints de carrito/items y el aislamiento por `empresa_id`.
	- Verificación: `get_errors` sobre HTML, JS y documentacion modificada.

- Gobernanza tecnica: se agrega checklist documental para QA y soporte.
	- Archivos creados: `documentos/gobernanza_tecnica/runbooks/checklist_evidencia_documental_para_qa_y_soporte.md`.
	- Archivos modificados: `documentos/gobernanza_tecnica/runbooks/README.md`, `documentos/gobernanza_tecnica/README.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se agrega una checklist operativa breve para QA y soporte sobre repositorio documental, firmas y exportes regulatorios, y se refleja en la documentaci?n general del proyecto y de arquitectura que un exporte no sustituye la evidencia versionada o firmada cuando el flujo es sensible.
	- Verificación: `get_errors` sobre la documentacion creada y modificada.

- Gobernanza tecnica: se endurece la reconciliacion documental y la evidencia regulatoria.
	- Archivos modificados: `documentos/gobernanza_tecnica/contratos/contrato_repositorio_documental_y_firmas_externas.md`, `documentos/gobernanza_tecnica/contratos/contrato_interoperabilidad_documental_contable_y_fiscal_externa.md`, `documentos/gobernanza_tecnica/contratos/contrato_reportes_contables_financieros_y_exportacion_multiformato.md`, `documentos/gobernanza_tecnica/runbooks/runbook_reconciliacion_documental_fiscal_y_contable_externa.md`, `documentos/gobernanza_tecnica/runbooks/runbook_versionado_documental_y_firmas_externas.md`, `documentos/gobernanza_tecnica/estandares_de_cambio_seguro.md`, `documentos/gobernanza_tecnica/README.md`, `documentos/gobernanza_tecnica/plan_implementacion_gobernanza_tecnica.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la gobernanza tecnica ahora reconcilia expl?citamente repositorio documental, firmas, interoperabilidad fiscal/contable y exportes multiformato, y fija que un exporte regulatorio no sustituye la versi?n documental vigente ni la firma asociada cuando el flujo exige evidencia reforzada.
	- Verificación: `get_errors` sobre la documentacion modificada.

- Gobernanza tecnica: se documentan repositorio documental y firmas externas.
	- Archivos creados: `documentos/gobernanza_tecnica/contratos/contrato_repositorio_documental_y_firmas_externas.md`, `documentos/gobernanza_tecnica/runbooks/runbook_versionado_documental_y_firmas_externas.md`.
	- Archivos modificados: `documentos/gobernanza_tecnica/README.md`, `documentos/gobernanza_tecnica/plan_implementacion_gobernanza_tecnica.md`, `documentos/gobernanza_tecnica/contratos/README.md`, `documentos/gobernanza_tecnica/runbooks/README.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la gobernanza tecnica ahora cubre formalmente `/api/empresa/documentos/gestion` y `/api/empresa/documentos/firmas`, incluyendo reglas de acceso por rol/modulo, versionado con historial, herencia de permisos para firmas y el procedimiento operativo para diagnosticar documentos no visibles, versiones incompletas o firmas huerfanas.
	- Verificación: `get_errors` sobre la documentacion creada y modificada.

- Super administrador: se agrega interruptor global para activar o desactivar la IA desde configuraci?n avanzada.
	- Archivos modificados: `backend/handlers/ai_config_handlers.go`, `backend/handlers/chat_con_inteligencia_artificial_controller.go`, `backend/handlers/chat_con_ia_global_super.go`, `backend/handlers/chat_con_inteligencia_artificial_controller_test.go`, `backend/handlers/chat_con_ia_global_super_test.go`, `backend/handlers/system_empresas_handlers_test.go`, `web/super/configuracion_avanzada.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: configuraci?n avanzada ahora puede apagar completamente el servicio IA mediante `ai.global.enabled`; cuando queda desactivado, el chat empresarial y el chat global super bloquean consultas nuevas y así se libera carga del servidor. Además, el botón `Probar IA` ejecuta una prueba real contra Ollama a trav?s del backend.
	- Verificación: prueba real por SSH en el VPS con `curl http://127.0.0.1:11434/api/tags` y `curl http://127.0.0.1:11434/api/generate`, ambas exitosas.

- Chat IA super y empresarial: se agrega aviso visible de servicio apagado y se explicita la prueba contra VPS.
	- Archivos modificados: `web/administrar_empresa/chat_con_inteligencia_artificial.html`, `web/super/chat_con_ia_global.html`, `web/super/configuracion_avanzada.html`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: cuando la IA global est? desactivada, ambos chats muestran un mensaje visible, deshabilitan el formulario y evitan que el usuario siga intentando consultar. En configuraci?n avanzada, el botón queda rotulado como `Probar IA contra VPS` para reflejar que la prueba es real y no solo de cat?logo.
	- Verificación: diagn?stico del editor sin errores en los HTML modificados.

- Super administrador: se agrega chat IA global con contexto consolidado del sistema.
	- Archivos creados: `backend/handlers/chat_con_ia_global_super.go`, `backend/handlers/chat_con_ia_global_super_test.go`, `web/super/chat_con_ia_global.html`.
	- Archivos modificados: `backend/db/chat_inteligencia_artificial.go`, `backend/db/chat_inteligencia_artificial_test.go`, `backend/handlers/chat_con_inteligencia_artificial_controller.go`, `backend/main.go`, `web/super_administrador.html`, `documentos/estructura_bd.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el panel super ahora tiene `chat_con_ia_global.html` con selector de modelo, historial y consultas sobre el contexto agregado de toda la base de datos; el historial global queda separado del chat por empresa mediante tablas `super_ai_*` y acceso exclusivo para sesiones `super_administrador`.
	- Verificación: `go test ./handlers -run 'TestSuperAI|TestModelosHandler|TestModeloPreferidoHandler|TestHistorialHandler' -count=1`; `go test ./db -run 'Test(EmpresaAI|SuperAI)' -count=1`.

- Chat IA empresarial: se habilita selector entre DeepSeek y Ambis Local por empresa.
	- Archivos modificados: `backend/handlers/ai_credentials_catalog.go`, `backend/handlers/ai_config_handlers.go`, `backend/handlers/chat_con_inteligencia_artificial_controller.go`, `backend/handlers/chat_con_inteligencia_artificial_controller_test.go`, `backend/handlers/system_empresas_handlers_test.go`, `backend/db/chat_inteligencia_artificial_test.go`, `web/administrar_empresa/chat_con_inteligencia_artificial.html`, `web/super/configuracion_avanzada.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el chat IA empresarial ahora permite elegir entre `deepseek:deepseek-chat` y `ollama:ambis`; Ambis usa `codellama:7b` servido por Ollama en el VPS a traves de loopback, manteniendo el filtro por `empresa_id` y la preferencia persistida por cuenta Google autenticada.
	- Verificación: `go test ./handlers -run 'Test(ModelosHandler|ModeloPreferidoHandler|HistorialHandler|ConsultarHandler|AIModelsConfigHandlerSaveDeepSeekEncrypted)' -count=1`; `go test ./db -run 'TestEmpresaAIModeloPreferidoUpsertAndGet|TestRegisterEmpresaAIConsultaAcumulaUsoDiario' -count=1`.

- Gobernanza tecnica: se documentan integraciones externas y reconciliacion documental.
	- Archivos creados: `documentos/gobernanza_tecnica/contratos/contrato_integraciones_bancarias_y_conectores_externos.md`, `documentos/gobernanza_tecnica/runbooks/runbook_reconciliacion_documental_fiscal_y_contable_externa.md`.
	- Archivos modificados: `documentos/gobernanza_tecnica/README.md`, `documentos/gobernanza_tecnica/plan_implementacion_gobernanza_tecnica.md`, `documentos/gobernanza_tecnica/contratos/README.md`, `documentos/gobernanza_tecnica/runbooks/README.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la gobernanza tecnica ahora cubre formalmente conectores API y bancarios, y agrega un procedimiento de reconciliacion entre compras, facturación, reintentos fiscales y repositorio documental.
	- Verificación: `get_errors` ejecutado sobre la documentacion creada y modificada, sin errores.

- Gobernanza tecnica: se documentan interoperabilidad documental y contingencias de integraciones externas.
	- Archivos creados: `documentos/gobernanza_tecnica/contratos/contrato_interoperabilidad_documental_contable_y_fiscal_externa.md`, `documentos/gobernanza_tecnica/runbooks/runbook_contingencias_integraciones_bancarias_y_conectores.md`.
	- Archivos modificados: `documentos/gobernanza_tecnica/README.md`, `documentos/gobernanza_tecnica/plan_implementacion_gobernanza_tecnica.md`, `documentos/gobernanza_tecnica/contratos/README.md`, `documentos/gobernanza_tecnica/runbooks/README.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la gobernanza tecnica ahora cubre formalmente la interoperabilidad entre compras, facturación, repositorio documental y conciliacion fiscal, y agrega un runbook espec?fico para incidentes de conectores API e integraciones bancarias.
	- Verificación: `get_errors` ejecutado sobre la documentacion creada y modificada, sin errores.

- Gobernanza tecnica: se documentan cierre de periodo contable y conciliacion bancaria.
	- Archivos creados: `documentos/gobernanza_tecnica/contratos/contrato_conciliacion_bancaria_y_cierre_periodo_contable.md`, `documentos/gobernanza_tecnica/runbooks/runbook_cierre_periodo_y_conciliacion_bancaria.md`.
	- Archivos modificados: `documentos/gobernanza_tecnica/README.md`, `documentos/gobernanza_tecnica/plan_implementacion_gobernanza_tecnica.md`, `documentos/gobernanza_tecnica/contratos/README.md`, `documentos/gobernanza_tecnica/runbooks/README.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la gobernanza tecnica ahora cubre formalmente el cierre y reapertura de periodos con evidencia obligatoria, la importacion idempotente de extractos, la conciliacion bancaria automatica y los bloqueos de movimientos cuando el periodo contable ya esta cerrado.
	- Verificación: diagnostico del editor sin errores en la documentacion creada y modificada.

- Estaciones: se agrega tarjeta especial `YouTube` con vista embebida y ampliacion.
	- Archivos creados: `web/administrar_empresa/youtube_station_browser.html`.
	- Archivos modificados: `web/administrar_empresa/configuracion_de_estaciones.html`, `web/administrar_empresa/estaciones.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el modulo de estaciones ahora permite activar una tarjeta especial `YouTube` desde `estaciones_config`, mostrar una vista embebida adaptable al tama?o de la tarjeta y abrirla en un overlay aproximado de `500 x 500` mediante un cuadrito de maximizaci?n `[]`.
	- Verificación: diagnostico del editor sin errores en `web/administrar_empresa/configuracion_de_estaciones.html`, `web/administrar_empresa/estaciones.html`, `web/administrar_empresa/youtube_station_browser.html` y `web/estilos.css`.

- Gobernanza tecnica: se documentan soporte remoto multiempresa y contingencias operativas de reportes.
	- Archivos creados: `documentos/gobernanza_tecnica/contratos/contrato_soporte_remoto_por_empresa_y_mesa_tecnica_central.md`, `documentos/gobernanza_tecnica/runbooks/runbook_reportes_programados_y_exportaciones_contables.md`, `documentos/gobernanza_tecnica/runbooks/runbook_soporte_remoto_sesiones_y_dispositivos.md`.
	- Archivos modificados: `documentos/gobernanza_tecnica/README.md`, `documentos/gobernanza_tecnica/plan_implementacion_gobernanza_tecnica.md`, `documentos/gobernanza_tecnica/runbooks/README.md`, `documentos/gobernanza_tecnica/contratos/README.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la gobernanza tecnica ahora cubre formalmente el modulo de soporte remoto por empresa, el portal publico del agente, la mesa tecnica super y los procedimientos de contingencia para reportes programados, exportaciones contables y sesiones/dispositivos remotos.
	- Verificación: diagnostico del editor sin errores en la documentacion creada y modificada.

- Gobernanza tecnica: se documentan DIAN, alertas de reinicio y reportes multiformato.
	- Archivos creados: `documentos/gobernanza_tecnica/runbooks/runbook_dian_set_pruebas_y_diagnostico_oficial.md`, `documentos/gobernanza_tecnica/runbooks/runbook_alertas_reinicio_y_monitoreo_gmail_smtp.md`, `documentos/gobernanza_tecnica/contratos/contrato_reportes_contables_financieros_y_exportacion_multiformato.md`.
	- Archivos modificados: `documentos/gobernanza_tecnica/README.md`, `documentos/gobernanza_tecnica/plan_implementacion_gobernanza_tecnica.md`, `documentos/gobernanza_tecnica/runbooks/README.md`, `documentos/gobernanza_tecnica/contratos/README.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la gobernanza tecnica ahora cubre el runbook del soporte DIAN real que existe hoy, el runbook de Gmail SMTP y alertas de reinicio del backend, y el contrato del modulo de reportes empresariales y globales super con datasets canonicos, exportacion `json/csv/txt/xls/pdf`, plantillas, programacion y validacion de consistencia.
	- Verificación: diagnostico del editor sin errores en la documentacion creada y modificada.

- Gobernanza interna: se implementa un equipo base de cuatro agentes con direccion centralizada en `agente_go`.
	- Archivos creados: `.github/agents/agente_backend_db.agent.md`, `.github/agents/agente_frontend_ux.agent.md`, `.github/agents/agente_qa_operacion.agent.md`, `.github/agents/README.md`.
	- Archivos modificados: `.github/agents/agente_go.agent.md`, `copilot-instructions.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: `agente_go` queda formalizado como agente principal, seleccionado por defecto a nivel de gobernanza del repositorio y responsable de dirigir a backend/DB, frontend/UX y QA/operacion, integrando una sola salida t?cnica y documental.
	- Verificación: validacion documental y de estructura del equipo interno de agentes completada en el repositorio.

- Gobernanza tecnica: se documenta formalmente el ciclo de facturacion electronica y documentos transaccionales.
	- Archivos creados: `documentos/gobernanza_tecnica/contratos/contrato_facturacion_electronica_y_documentos_transaccionales.md`.
	- Archivos modificados: `documentos/gobernanza_tecnica/plan_implementacion_gobernanza_tecnica.md`, `documentos/gobernanza_tecnica/contratos/README.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la capa de gobernanza ahora cubre la maquina de estados documental de facturación, la persistencia comun en `empresa_facturacion_documentos`, el selector `modo_documento_venta`, la cola de reintentos, la reconciliacion fiscal y la base operativa actual de DIAN Colombia, dejando explicito lo que aun esta pendiente del transporte oficial.
	- Verificación: diagnostico del editor sin errores en los archivos documentales creados y modificados.

- Gobernanza tecnica: se documenta formalmente la capa de permisos_contexto y wrappers de rutas empresariales.
	- Archivos creados: `documentos/gobernanza_tecnica/contratos/contrato_permisos_contexto_y_wrappers_api_empresa.md`.
	- Archivos modificados: `documentos/gobernanza_tecnica/plan_implementacion_gobernanza_tecnica.md`, `documentos/gobernanza_tecnica/contratos/README.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la capa de gobernanza ahora cubre la frontera de autorizacion para `/api/empresa/*`, incluyendo wrappers por modulo, endpoint `permisos_contexto`, overrides por rol, restricciones por licencia y aprobacion trazable en operaciones sensibles de seguridad.
	- Verificación: diagnostico del editor sin errores en los archivos documentales creados y modificados.

- Gobernanza tecnica: se documenta formalmente la venta publica empresarial por empresa.
	- Archivos creados: `documentos/gobernanza_tecnica/contratos/contrato_venta_publica_empresarial_por_empresa.md`.
	- Archivos modificados: `documentos/gobernanza_tecnica/plan_implementacion_gobernanza_tecnica.md`, `documentos/gobernanza_tecnica/contratos/README.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la capa de gobernanza ahora cubre el contrato del modulo de venta publica por `empresa_id`, incluyendo configuracion de tienda, catalogo, ordenes publicas, pagos Wompi/Epayco y consulta de estado con exposicion segura de datos.
	- Verificación: diagnostico del editor sin errores en los archivos documentales creados y modificados.

- Gobernanza tecnica: se documenta formalmente la autenticacion multirol y el arranque local PostgreSQL por tunel.
	- Archivos creados: `documentos/gobernanza_tecnica/contratos/contrato_autenticacion_administrativa_y_usuarios_empresa.md`, `documentos/gobernanza_tecnica/runbooks/runbook_arranque_postgresql_tunel_local.md`.
	- Archivos modificados: `documentos/gobernanza_tecnica/plan_implementacion_gobernanza_tecnica.md`, `documentos/gobernanza_tecnica/contratos/README.md`, `documentos/gobernanza_tecnica/runbooks/README.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la capa de gobernanza ahora cubre el contrato de autenticacion administrativa y de usuarios de empresa, junto con el runbook del arranque local del backend cuando PostgreSQL del VPS se consume por tunel SSH y DSN reescrito hacia `DB_VPS_LOCAL_PORT`.
	- Verificación: diagnostico del editor sin errores en los archivos documentales creados y modificados.

- Gobernanza tecnica: se documenta formalmente el flujo de estaciones, sensores y venta simple por estacion.
	- Archivos creados: `documentos/gobernanza_tecnica/contratos/contrato_estaciones_sensores_ventas_simple.md`, `documentos/gobernanza_tecnica/runbooks/runbook_estaciones_sensores_ventas_simple.md`.
	- Archivos modificados: `documentos/gobernanza_tecnica/plan_implementacion_gobernanza_tecnica.md`, `documentos/gobernanza_tecnica/contratos/README.md`, `documentos/gobernanza_tecnica/runbooks/README.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la capa de gobernanza ahora cubre el flujo de `estaciones_config`, sensores por `last_seen`, carrito base canonico `EST-empresa-estacion`, cierre de `pagar_estacion`, metricas de estacion y recuperacion de incidentes sin cambiar permisos ni wrappers.
	- Verificación: diagnostico del editor sin errores en los archivos documentales creados y modificados.

- Gobernanza tecnica: se crea la capa base de ADRs, contratos, runbooks y cambio seguro del repositorio.
	- Archivos creados: `documentos/README.md`, `documentos/gobernanza_tecnica/README.md`, `documentos/gobernanza_tecnica/plan_implementacion_gobernanza_tecnica.md`, `documentos/gobernanza_tecnica/estandares_de_cambio_seguro.md`, `documentos/gobernanza_tecnica/adr/ADR-0001-frontera-multiempresa-empresa-id.md`, `documentos/gobernanza_tecnica/adr/ADR-0002-postgresql-runtime-canonico-vps.md`, `documentos/gobernanza_tecnica/contratos/README.md`, `documentos/gobernanza_tecnica/contratos/contrato_checkout_licencias_publico.md`, `documentos/gobernanza_tecnica/runbooks/README.md`, `documentos/gobernanza_tecnica/runbooks/runbook_checkout_licencias.md`.
	- Archivos modificados: `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se formaliza la gobernanza tecnica del proyecto para mejorar decisiones arquitectonicas, cambios seguros, contratos de flujos criticos y respuesta ante incidentes repetidos, comenzando por el checkout publico de licencias.
	- Verificación: diagnostico del editor sin errores en los archivos documentales creados y modificados.

- Arranque PostgreSQL: el backend ahora respeta el puerto del tunel local.
	- Archivos modificados: `backend/main.go`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: `resolveRuntimePostgresDSN` reescribe los DSN hacia `DB_VPS_LOCAL_PORT` cuando `DB_VPS_TUNNEL_ENABLED=1`, evitando que `go run .` o el binario del backend fallen por autenticacion contra `127.0.0.1:5432` cuando la conexion valida al VPS pasa por otro puerto local del tunel.
	- Verificación: `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`; `go run .` con `.env.local` y tunel activo.

- Configuracion empresarial: el bloque general `Productos y pedidos` ahora guarda en backend.
	- Archivos creados: `backend/db/empresa_configuracion_general.go`, `backend/handlers/empresa_configuracion_general.go`, `backend/handlers/empresa_configuracion_general_test.go`.
	- Archivos modificados: `backend/main.go`, `web/administrar_empresa/configuracion.html`, `documentos/estructura_bd.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la seccion principal de configuracion empresarial deja de simular guardado local y pasa a persistir por empresa con `GET/PUT /api/empresa/configuracion_general`, incluyendo orden de servicio, descuentos y lector de codigo de barras.
	- Verificación: diagnostico del editor sin errores en los archivos nuevos y modificados; `go test ./handlers -run '^(TestEmpresaConfiguracionGeneralHandlerGetAndSave|TestEmpresaVentaPublicaHandlerConfigCatalogoYToggle|TestEmpresaConfiguracionOperativaHandler(ConfigAndRole|PoliticaSimulacionHistorialYRollback))$' -count=1`.

- Configuracion empresarial: `Avanzada` ya no sale al panel super.
	- Archivos modificados: `web/administrar_empresa/configuracion_menu.html`, `web/administrar_empresa/configuracion.html`, `documentos/descripcion_de_modulos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el submenu de configuracion empresarial corrige el enlace `Avanzada` para que apunte a la seccion avanzada real dentro de la configuracion de empresa, evitando navegar por error a `/super/configuracion_avanzada.html`.
	- Verificación: diagnostico del editor sin errores en `web/administrar_empresa/configuracion_menu.html` y `web/administrar_empresa/configuracion.html`.

- Configuracion empresarial: `Permisos` e `Integraciones` dejan de ser placeholders y se valida el guardado real.
	- Archivos modificados: `web/administrar_empresa/configuracion_permisos.html`, `web/administrar_empresa/configuracion_integraciones.html`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la vista `Permisos` ahora consulta el contexto real de permisos por empresa y muestra la matriz disponible en solo lectura; la vista `Integraciones` reutiliza el endpoint real de `venta_publicaaction=config` para cargar y guardar la configuracion de Wompi/Epayco y la tienda publica de la empresa.
	- Verificación: diagnostico del editor sin errores en `web/administrar_empresa/configuracion_permisos.html` y `web/administrar_empresa/configuracion_integraciones.html`; `go test ./handlers -run '^TestEmpresaVentaPublicaHandlerConfigCatalogoYToggle$' -count=1`; `go test ./handlers -run '^TestEmpresaConfiguracionOperativaHandler(ConfigAndRole|PoliticaSimulacionHistorialYRollback)$' -count=1`.

- Ventas simples por estacion: se agrega boton `Regresar a estaciones`.
	- Archivos modificados: `web/administrar_empresa/ventas_simple.html`, `web/js/ventas_simple.js`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la vista de carrito por estacion ahora ofrece un retorno directo a `administrar_empresa/estaciones.html`, manteniendo el `empresa_id` activo para volver al tablero de estaciones sin depender del historial del navegador.
	- Verificación: diagnostico del editor sin errores en `web/administrar_empresa/ventas_simple.html` y `web/js/ventas_simple.js`.

- Configuracion empresarial: el submenú `Configuraci?n` ahora puede ocultar y mostrar su menú lateral en celular.
	- Archivos modificados: `web/administrar_empresa/configuracion_menu.html`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la página `administrar_empresa/configuracion_menu.html` carga `menu.js`, adopta el wrapper `admin-sidebar-mobile-collapsible` y agrega el mismo botón final de `Ocultar menú` / `Mostrar menú` usado por otros shells administrativos, para que el submenú de configuraci?n tambi?n sea plegable en m?vil.
	- Verificación: diagnostico del editor sin errores en `web/administrar_empresa/configuracion_menu.html`.

- Estaciones: se elimina el circulo inferior de la tarjeta y se conserva solo el indicador cuadrado del sensor.
	- Archivos modificados: `web/administrar_empresa/estaciones.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la tarjeta de estaciones deja de mostrar el circulo centrado inferior; solo queda el cuadrito superior derecho, listo para ponerse verde cuando el sensor de la estacion se active.
	- Verificación: diagnostico del editor sin errores en `web/administrar_empresa/estaciones.html` y `web/estilos.css`.

- Checkout de licencias: valida contexto multiempresa y corrige la empresa usada en el correo de activacion.
	- Archivos modificados: `backend/db/db.go`, `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `web/pagar_licencia.html`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el backend ahora resuelve la empresa por `empresa_id` logico al construir el correo de activacion, endurece los helpers de `pagos_epayco` y `pagos_wompi` con autorreparacion del esquema, y rechaza conciliaciones de `transaction_status` cuando la referencia pertenece a otra empresa o licencia distinta de la pagina abierta; el frontend envia ese contexto esperado en cada polling y lo mantiene al cerrar el pago aprobado.
	- Verificación: `go test ./handlers -run 'Test(EpaycoTransactionStatusHandlerActivatesOnceAndCapturesEmail|EpaycoTransactionStatusHandlerUsesEmpresaScopeForActivationMailBody|EpaycoTransactionStatusHandlerRejectsUnexpectedEmpresaContext|EpaycoWebhookHandlerFindsContextUsingInvoiceFallback|EpaycoTransactionStatusHandlerRetriesActivationEmailAfterWebhookActivatedFirst|WompiTransactionStatusHandlerAllowsReferenceLookup)' -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Selector de empresas: tarjetas mas peque?as y botones alineados al pie.
	- Archivos modificados: `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: `seleccionar_empresa.html` mantiene su mismo flujo, pero las tarjetas del grid se compactan en escritorio y la botonera inferior queda centrada y pegada al pie de cada bloque para que la fila de acciones no quede flotando a media altura.
	- Verificación: diagnostico del editor sin errores en `web/estilos.css`.

- Super configuracion avanzada: el boton `Probar Gmail` ahora envia un correo real de prueba.
	- Archivos modificados: `backend/handlers/usuarios_empresa.go`, `backend/handlers/system_empresas_handlers_test.go`, `web/super/configuracion_avanzada.html`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la configuracion avanzada de Gmail deja de validar solo si existen credenciales y pasa a ejecutar un envio de prueba real a `powerfulcontrolsystem@gmail.com`, reutilizando la configuracion SMTP guardada en PostgreSQL; en pruebas automatizadas el flujo se captura en la tabla de notificaciones de test.
	- Verificación: `go test ./handlers -run '^TestGmailConfigHandlerTestActionCapturesNotification$' -count=1`.

- Arcade publico: Brigada burbujas 3D plus agrega joystick tactil, fullscreen y HUD de una mano.
	- Archivos modificados: `web/Juegos/brigada_burbujas_3d_plus.html`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el shooter movil sustituye el pad clasico por joystick tactil, pide pantalla completa al iniciar en celular y concentra arma/sector dentro del HUD del escenario para que el juego se pueda usar mejor con una mano.
	- Verificación: diagnostico del editor sin errores en `web/Juegos/brigada_burbujas_3d_plus.html`.

- Arcade publico: Brigada burbujas 3D plus concentra los accesos tacticos dentro del escenario para celular.
	- Archivos modificados: `web/Juegos/brigada_burbujas_3d_plus.html`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el shooter ya no obliga a desplazarse al rack de arsenal en pantallas pequenas; ahora arma rapida y pausa viven dentro del escenario con una barra tactica sincronizada con el HUD y los controles inferiores.
	- Verificación: diagnostico del editor sin errores en `web/Juegos/brigada_burbujas_3d_plus.html`.

- Arcade publico: Brigada burbujas 3D plus agrega arsenal, pickups y sectores abiertos/cerrados.
	- Archivos modificados: `web/Juegos/brigada_burbujas_3d_plus.html`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el shooter 3D simulado del arcade ahora incluye tres armas, pickups de salud y municion, sectores cerrados/abiertos/hibridos, perspectiva de suelo mas marcada y una IA que patrulla, busca, flanquea, dispara y convoca refuerzos, empujando la experiencia hacia un Doom caricaturesco sin librerias externas.
	- Verificación: diagnostico del editor sin errores en `web/Juegos/brigada_burbujas_3d_plus.html`.

- Licencias Epayco: el correo de activaci?n ahora se reintenta de forma idempotente despu?s de aprobados posteriores si la licencia ya qued? activa pero la notificaci?n todavía no qued? confirmada.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el flujo Epayco deja de depender de una activaci?n ÃƒÂ¢Ã¢â€šÂ¬Ã…?reci?n creadaÃƒÂ¢Ã¢â€šÂ¬Ã‚Â para enviar el correo; adem?s, ahora tambi?n recupera el `customer_email` cuando la validación lo devuelve anidado en `data`, evitando perder la notificaci?n si el webhook aprob? primero o si el primer intento fall? temporalmente.
	- Verificación: `go test ./handlers -run 'TestEpayco(TransactionStatusHandlerActivatesOnceAndCapturesEmail|WebhookHandlerFindsContextUsingInvoiceFallback|TransactionStatusHandlerRetriesActivationEmailAfterWebhookActivatedFirst)' -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Arcade publico: Brigada burbujas 3D plus agrega boton Auto visible, ayuda de mira movil y ajuste fino de impacto.
	- Archivos modificados: `web/Juegos/brigada_burbujas_3d_plus.html`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: El HUD movil incorpora un toggle directo de auto-disparo y el panel tactil gana intensidad de feedback y ayuda suave de mira configurable para mejorar la respuesta en celular sin quitar control manual.
	- Verificación: diagnostico del editor sin errores en `web/Juegos/brigada_burbujas_3d_plus.html`.

- Arcade publico: Brigada burbujas 3D plus activa auto-disparo y preset facil por defecto en movil.
	- Archivos modificados: `web/Juegos/brigada_burbujas_3d_plus.html`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: El juego completa el HUD y el panel tactil, fuerza una migracion unica de preferencias antiguas para dejar `Auto ON`, subir la ayuda de mira y suavizar el control desde el primer arranque en celular.
	- Verificación: diagnostico del editor sin errores en `web/Juegos/brigada_burbujas_3d_plus.html`.

## 2026-04-17

- Gobernanza interna: se agrega semaforo ejecutivo por modulo y se endurece el rechazo de cierres sin evidencia por frente.
	- Archivos modificados: `.github/agents/protocolo_delegacion.md`, `.github/agents/plantilla_trabajo_por_modulo.md`, `.github/agents/agente_backend_db.agent.md`, `.github/agents/agente_frontend_ux.agent.md`, `.github/agents/agente_qa_operacion.agent.md`, `.github/agents/README.md`, `copilot-instructions.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el equipo de agentes ya cuenta con un semaforo ejecutivo `Rojo/Amarillo/Verde`, un ejemplo completo extremo a extremo para delegacion real y reglas explicitas para que backend, frontend y QA rechacen cierres sin evidencia minima suficiente.
	- Verificación: validacion documental y de consistencia interna completada sobre el endurecimiento final del protocolo del equipo.

- Gobernanza interna: se agregan tabla rapida por modulo, ejemplos reales y endurecimiento de cierre para modulos criticos del equipo de agentes.
	- Archivos modificados: `.github/agents/protocolo_delegacion.md`, `.github/agents/plantilla_trabajo_por_modulo.md`, `.github/agents/agente_go.agent.md`, `.github/agents/README.md`, `copilot-instructions.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el protocolo ahora incluye una tabla corta por modulo para consulta inmediata, ejemplos reales de delegacion y una regla mas dura para que `agente_go` no cierre modulos criticos sin la participacion obligatoria definida.
	- Verificación: validacion documental y de consistencia interna completada sobre la ampliacion del protocolo del equipo.

- Gobernanza interna: se formaliza el protocolo de delegacion y la plantilla comun de trabajo del equipo de agentes.
	- Archivos creados: `.github/agents/protocolo_delegacion.md`, `.github/agents/plantilla_trabajo_por_modulo.md`.
	- Archivos modificados: `.github/agents/README.md`, `.github/agents/agente_go.agent.md`, `.github/agents/agente_backend_db.agent.md`, `.github/agents/agente_frontend_ux.agent.md`, `.github/agents/agente_qa_operacion.agent.md`, `copilot-instructions.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: `agente_go` ya no solo dirige al equipo; ahora tambi?n aplica una matriz exacta de delegaci?n por tipo de tarea y una plantilla de ejecución compartida por módulo, mientras cada especialista queda priorizado por módulos críticos del sistema.
	- Verificación: validacion documental y de consistencia interna completada sobre la nueva capa de gobernanza del equipo.

- Ventas por estacion: compatibilidad PostgreSQL restaurada en carritos, metricas y documento de venta.
	- Archivos modificados: `backend/db/carritos_compras.go`, `backend/db/empresa_configuracion_avanzada.go`, `backend/db/documentos_transaccionales.go`, `backend/db/sql_compat.go`, `backend/main.go`, `backend/db/facturacion_electronica_test.go`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: las inserciones de carritos, items y metricas dejan de depender de `LastInsertId` y pasan a usar la capa portable motor legado retirado/PostgreSQL; ademas, la configuracion avanzada por empresa se regulariza antes de consultar `modo_documento_venta`, las tablas legacy de documentos transaccionales recuperan un `id` autogenerado valido y el backend sanea globalmente cualquier tabla PostgreSQL heredada con llave primaria `id` sin secuencia/default.
	- Verificación: `go test ./db -run 'Test(GetEmpresaConfiguracionAvanzadaRepairsMissingModoDocumentoVentaColumn|PrepareFacturacionDocumentoLegal|FacturacionElectronicaRetryUpsertGetAndList)' -count=1`; `go test ./handlers -run 'Test(VentaCarritoFacturaYResolucionImpresora|VentaCarritoGeneraComprobantePagoSegunConfiguracion)' -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Checkout de licencias: Epayco queda en tarjeta blanca, compacta y sin correo visible.
	- Archivos modificados: `web/pagar_licencia.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: `pagar_licencia.html` elimina el bloque separado de formas de pago Epayco, oculta el campo de correo del panel y deja el checkout en una tarjeta blanca mas peque?a y centrada; `web/estilos.css` adapta el branding del panel para ese layout.


## 2026-04-17

- Selector de empresas: editar pasa al menu lateral y las tarjetas cambian de estilo.
	- Archivos modificados: `web/seleccionar_empresa.html`, `web/js/seleccionar_empresa.js`, `web/editar_empresa.html`, `web/js/editar_empresa.js`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: `seleccionar_empresa.html` elimina el boton `Editar` dentro de las tarjetas, a?ade `Editar empresa` al menu lateral, conserva el orden del texto principal de cada tarjeta y adopta una presentacion visual nueva con botones cuadrados. La pantalla `editar_empresa.html` queda enfocada solo en editar o eliminar.
	- Verificación: diagnostico del editor sin errores en los archivos de frontend modificados.

## 2026-04-17

- Arcade publico: Brigada burbujas 3D plus ahora tiene campa?a larga, transformaciones y rivales de pasarela caricaturesca.
	- Archivos modificados: `web/Juegos/brigada_burbujas_3d_plus.html`, `web/Juegos/menu_juegos.html`, `web/img/juegos/brigada_burbujas_3d.svg`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el shooter 3D simulado del arcade crece a cinco niveles, incorpora tres poderes de transformacion, enemigos con IA mas agresiva y una presentacion visual renovada en el lobby y la portada.
	- Verificación: diagnostico del editor sin errores en `web/Juegos/brigada_burbujas_3d_plus.html`, `web/Juegos/menu_juegos.html` y `web/img/juegos/brigada_burbujas_3d.svg`.

## 2026-04-17

- Arcade publico: Brigada burbujas 3D plus refuerza modo movil y ambiente claro.
	- Archivos modificados: `web/Juegos/brigada_burbujas_3d_plus.html`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el shooter 3D simulado del arcade ajusta su arte a una paleta pastel de dibujos animados, agrega apuntado tactil sobre el escenario y mejora el layout responsive para movil.
	- Verificación: diagnostico del editor sin errores en `web/Juegos/brigada_burbujas_3d_plus.html`.

## 2026-04-17

- Selector de empresas: administradores pueden crear su primera empresa sin 403.
	- Archivos modificados: `backend/utils/utils.go`, `backend/handlers/system_empresas_handlers_test.go`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: `AuthMiddleware` vuelve a permitir `POST /super/api/empresas` al rol `administrador` para que `seleccionar_empresa.html` pueda dar de alta empresas propias, manteniendo restringido el resto del panel super.
	- Verificación: `go test ./handlers -run '^TestNuevoAdminRegistradoPuedeCrearSuPrimeraEmpresaViaRutaSuperProtegida$' -count=1`; `go test ./utils -run '^TestSuperEndpointsPermisosPorRol$' -count=1`.

## 2026-04-18

- Arcade publico: se agrega Brigada burbujas 3D plus como decimo juego activo del portal.
	- Archivos creados: `web/Juegos/brigada_burbujas_3d_plus.html`, `web/img/juegos/brigada_burbujas_3d.svg`.
	- Archivos modificados: `web/Juegos/menu_juegos.html`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el arcade del portal incorpora un shooter original con raycasting, controles tactiles, enemigos caricaturescos y guardado local de record, elevando el lobby a diez juegos publicos activos.
	- Verificación: diagnostico del editor sin errores en `web/Juegos/brigada_burbujas_3d_plus.html` y `web/Juegos/menu_juegos.html`.

## 2026-04-17

- Seguridad VPS del super: vista en modo oscuro.
	- Archivos modificados: `web/super/seguridad.html`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la pantalla `web/super/seguridad.html` cambia a una paleta oscura para alinearse visualmente con el resto del panel super, sin alterar endpoints ni permisos.
	- Verificación: diagnostico del editor sin errores en `web/super/seguridad.html`.

## 2026-04-17

- editar_empresa: se agrega un boton `Comprar licencia` visible solo cuando la empresa tiene licencia vencida y no conserva una licencia vigente activa.
- ventas por empresa: se agrega `modo_documento_venta` para que cada empresa elija entre `factura_electronica` y `comprobante_pago`, y el cierre de `pagar_estacion` genera automaticamente el documento correspondiente.
- documentacion y arbol operativo: se corrigen fechas futuras en la documentacion principal y se registra la salida del arbol actual de los scripts de revision ortografica (`scripts/spellcheck.*`, `scripts/spell_whitelist.txt`, `scripts/README-spellcheck.md`).

## 2026-04-17
- Menus administrativos: boton movil para ocultar y mostrar el sidebar.
	- Archivos modificados: `web/administrar_empresa.html`, `web/super_administrador.html`, `web/seleccionar_empresa.html`, `web/menu.js`, `web/estilos.css`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: las vistas `administrar_empresa.html`, `super_administrador.html` y `seleccionar_empresa.html` incorporan un boton final visible solo en movil para ocultar o mostrar el menu lateral, manteniendo intacta la navegacion en escritorio.
	- Verificación: diagnostico del editor sin errores en los archivos HTML, `web/menu.js` y `web/estilos.css`.

## 2026-04-17
- Empresas super: nueva pagina editar_empresa con eliminacion total confirmada.
	- Archivos creados: `backend/db/empresas_delete.go`, `web/editar_empresa.html`, `web/js/editar_empresa.js`.
	- Archivos modificados: `backend/handlers/system_empresas_handlers.go`, `backend/handlers/system_empresas_handlers_test.go`, `backend/utils/utils.go`, `web/js/seleccionar_empresa.js`, `web/estilos.css`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se a?ade `editar_empresa.html` para actualizar nombre y descripcion de la empresa seleccionada y se incorpora `action=eliminar_total` en `/super/api/empresas`, que purga datos relacionados por `empresa_id` en la base operativa y en la base super tras confirmar el nombre exacto.
	- Verificación: `go test ./handlers -run '^TestEmpresasHandlerEliminarTotalPurgaDatosRelacionados$' -count=1 -v -timeout 60s`; `go test ./handlers -run '^TestAdministradorPuedeEditarYEliminarEmpresaDesdeRutaSuperProtegida$' -count=1 -v -timeout 60s`; `go test ./handlers -run '^TestSuperEndpointsPermisosPorRol$' -count=1 -v -timeout 60s`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

## 2026-04-17
- Descarga de empresa: ruta corta funcional, exportacion en la misma vista y modo oscuro.
	- Archivos modificados: `backend/main.go`, `web/descargar_informacion_de_la_empresa.html`, `web/js/descargar_informacion_de_la_empresa.js`, `web/estilos.css`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la vista de descarga consolidada ahora responde tambien en `/descargar_informacion_de_la_empresa`, ejecuta las descargas PDF/XLS/CSV/JSON/TXT dentro de la misma pantalla con manejo de errores y usa una interfaz oscura dedicada.
	- Verificación: diagnostico del editor sin errores en `backend/main.go`, `web/descargar_informacion_de_la_empresa.html`, `web/js/descargar_informacion_de_la_empresa.js` y `web/estilos.css`.

- Selector de empresas: administradores no-super recuperan las lecturas iniciales.
	- Archivos modificados: `backend/utils/utils.go`, `backend/handlers/system_empresas_handlers_test.go`, `backend/handlers/auth_users_carritos_test.go`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se corrige el `403` que imped?a abrir `seleccionar_empresa.html` a cuentas con rol `administrador`, habilitando solo `GET /super/api/empresas`, `GET /super/api/tipos_empresas` y `GET /super/api/licencias` para esa vista, sin devolver permisos de escritura ni acceso al resto del panel super.
	- Verificación: `go test ./handlers -run '^TestSuperEndpointsPermisosPorRol$' -count=1`; `go test ./handlers -run '^TestNuevoAdminRegistradoNoObtieneAccesoSuperParaCrearEmpresa$' -count=1`; `go test ./utils -run '^TestAuthMiddlewareAllowsPublicLicenciaPaymentRoutesWithoutSession$' -count=1`.

- Registro de contrasena Google: se elimina `Continuar` y queda solo `Guardar` centrado.
	- Archivos modificados: `web/registrar_contrasena_usuario_de_google.html`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la pantalla `registrar_contrasena_usuario_de_google.html` ahora muestra un unico CTA de guardado, centrado, para reforzar que este paso no debe saltarse.
	- Verificación: diagnostico del editor sin errores en `web/registrar_contrasena_usuario_de_google.html`.

- Checkout Epayco: retorno con referencia real y pantalla de pago exitoso.
	- Archivos creados: `web/epayco/pago_exitoso.html`.
	- Archivos modificados: `web/epayco/respuesta.html`, `web/pagar_licencia.html`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el retorno web de Epayco ahora prioriza `x_ref_payco` y el estado real de la pasarela para que la validacion posterior sí pueda activar la licencia; cuando el backend confirma `APPROVED`, el usuario sale a una pantalla de pago exitoso y de all? vuelve a `seleccionar_empresa.html`.
	- Verificación: diagnostico del editor sin errores en los HTML modificados; `go test ./handlers -run 'TestEpayco(TransactionStatusHandlerFindsContextUsingInvoiceWhenGatewayIDsDiffer|TransactionStatusHandlerActivatesOnceAndCapturesEmail|WebhookHandlerFindsContextUsingInvoiceFallback|CreateTransactionHandlerUsesConfiguredPublicBaseURLAndKeys|TransactionStatusHandlerPreservesPendingOnGenericValidationError)' -count=1`.

- Seleccionar empresa: se corrige el error `escapeHtml is not defined`.
	- Archivos modificados: `web/js/seleccionar_empresa.js`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el listado del panel `seleccionar_empresa.html` vuelve a renderizar correctamente porque se restaura el helper local de escape HTML usado por las tarjetas de empresa tras el refactor del flujo de descarga.
	- Verificación: diagnostico del editor sin errores en `web/js/seleccionar_empresa.js`.

- Autenticacion administrativa: super restringido a powerfulcontrolsystem@gmail.com.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/utils/utils.go`, `backend/handlers/auth_admin_handlers_test.go`, `backend/handlers/system_empresas_handlers_test.go`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_del_proyecto`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el registro publico de administradores, el login por correo y el callback de Google ya no elevan cuentas nuevas o legacy a `super_administrador`. Solo `powerfulcontrolsystem@gmail.com` mantiene ese rol en el flujo publico; el resto queda como `administrador` y no entra a `/super/*`.
	- Verificación: `go test ./handlers -run "Test(AdminRegisterHandlerCreatesPendingAdminAndCapturesConfirmationMail|AdminLoginHandlerCreatesSessionForConfirmedAdmin|AdminLoginHandlerKeepsGenericAdminWithoutSuperPrivileges|AdminRegisterHandlerReservedEmailKeepsSuperRole|HandleGoogleCallbackNewEmailKeepsAdministradorRole|NuevoAdminRegistradoNoObtieneAccesoSuperParaCrearEmpresa)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Accept: ocultar la metadata visible del contrato.
	- Archivos modificados: `web/accept.html`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: `accept.html` ya no muestra la linea `Version vigente | actualizada`, pero mantiene el enlace al contrato completo y sigue resolviendo internamente la version para apuntar a la ruta correcta.
	- Verificación: diagnostico del editor sin errores en `web/accept.html`.

- Home publico: CTA inferior fijo y texto de tarjetas con mayor contraste.
	- Archivos modificados: `web/estilos.css`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: las tarjetas de `index.html` ahora reservan el espacio flexible para la descripcion, dejan el boton `Explorar oferta` siempre abajo y centrado, y mejoran la legibilidad del titulo y el texto con tipografia clara e iluminado exterior negro suave sin agregar paneles de fondo al contenido textual.
	- Verificación: diagnostico del editor sin errores en `web/estilos.css` y `web/index.html`.

- Empresas super: prueba end-to-end para usuario nuevo creando su primera empresa.
	- Archivos modificados: `backend/handlers/system_empresas_handlers_test.go`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se agrega una regresion que cubre el caso reportado de un administrador nuevo que se registra, confirma su cuenta, inicia sesion y crea su primera empresa mediante `POST /super/api/empresas` bajo `AuthMiddleware`.
	- Verificación: `go test ./handlers -run "Test(NuevoAdminRegistradoPuedeCrearSuPrimeraEmpresaViaRutaSuperProtegida|AdminRegisterHandlerCreatesPendingAdminAndCapturesConfirmationMail|AdminLoginHandlerCreatesSessionForConfirmedAdmin|AdminLoginHandlerPromotesLegacySelfRegisteredAdminToSuper)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Administradores autoregistrados: crear empresa deja de fallar para cuentas nuevas.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/utils/utils.go`, `backend/handlers/auth_admin_handlers_test.go`, `web/js/seleccionar_empresa.js`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el registro administrativo publico ahora crea cuentas con rol `super_administrador`, las cuentas legacy autoregistradas se promueven automaticamente al entrar o al tocar rutas `/super/*`, y el formulario de `seleccionar_empresa.html` envia correctamente el tipo de empresa al crearla.
	- Verificación: `go test ./handlers -run "TestAdmin(RegisterHandlerCreatesPendingAdminAndCapturesConfirmationMail|LoginHandlerCreatesSessionForConfirmedAdmin|LoginHandlerPromotesLegacySelfRegisteredAdminToSuper)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Home publico: CTA de tarjetas anclado abajo y descripcion en oscuro.
	- Archivos modificados: `web/estilos.css`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: las tarjetas de `index.html` ahora dejan el boton `Explorar oferta` siempre al pie y centrado, mientras la descripcion usa un color oscuro y negrita para ganar contraste frente al fondo de cada tarjeta.
	- Verificación: diagnostico del editor sin errores en `web/estilos.css`.

- Venta pública por empresa: Wompi y Epayco con credenciales propias por empresa.
	- Archivos modificados: `backend/db/venta_publica.go`, `backend/handlers/venta_publica.go`, `backend/handlers/payments_handlers.go`, `backend/handlers/venta_publica_test.go`, `backend/main.go`, `web/venta_publica.html`, `web/administrar_empresa/venta_publica.html`, `web/administrar_empresa/configuracion.html`, `documentos/estructura_bd.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: cada empresa puede activar o desactivar Wompi y Epayco con sus propias credenciales, administrar esas llaves desde `configuracion.html` y reutilizarlas en su tienda pública; adem?s los webhooks de ambas pasarelas ya actualizan las ?rdenes de `empresa_venta_publica_ordenes`.
	- Verificación: `go test ./handlers -run "Test(EmpresaVentaPublicaHandlerConfigCatalogoYToggle|PublicVentaPublicaHandlerCatalogoYPagoConWompiInactivo|PublicVentaPublicaHandlerEstadoPagoRequiereOrderCode)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1`.

- Seleccionar empresa: nueva pagina para descargar informacion empresarial en formatos profesionales.
	- Archivos creados: `backend/handlers/system_empresas_export.go`, `web/descargar_informacion_de_la_empresa.html`, `web/js/descargar_informacion_de_la_empresa.js`.
	- Archivos modificados: `backend/handlers/system_empresas_handlers.go`, `backend/handlers/system_empresas_handlers_test.go`, `web/js/seleccionar_empresa.js`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el boton de descarga de las tarjetas en `seleccionar_empresa.html` ahora abre una pagina dedicada que consolida la informacion de la empresa seleccionada y permite descargarla en `PDF`, `XLS`, `CSV`, `JSON` y `TXT` desde nuevas acciones protegidas del endpoint `/super/api/empresas`.
	- Verificación: `go test ./handlers -run "TestEmpresasHandler(ResumenDescargaYExport|ImpactoDesactivacion|DesactivarConImpactoYForce)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Facturacion DIAN: fase 1 base con UBL 2.1, firma XAdES y diagnostico oficial.
	- Archivos modificados: `backend/handlers/modulos_faltantes.go`, `backend/handlers/modulos_faltantes_test.go`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el modulo DIAN incorpora una fase 1 base para generar XML UBL 2.1 estructural, incrustar una firma XMLDSig/XAdES base y emitir un diagnostico de brechas frente al contrato oficial DIAN, manteniendo separado el transporte SOAP/WSDL definitivo.
	- Verificación: `go test ./handlers -run 'TestEmpresaDIANColombiaHandler(GenerarXMLUBLBase|FirmarXMLXAdESBase|DiagnosticoOficial|FirmaEnvioYAcuseReal|EnviarSetPruebas|SoftwareCompartidoMultiempresa|GuiaOnboardingYValidarCredenciales|SubirFirma)' -count=1`.

- Seleccionar empresa: boton de descarga blanco, solo con icono y tooltip.
	- Archivos modificados: `web/js/seleccionar_empresa.js`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el boton de descarga dentro de las tarjetas de `seleccionar_empresa.html` deja de mostrar el texto `Descargar`, queda como icono blanco y muestra el tooltip `Descargar informacion de la empresa` al pasar el mouse.
	- Verificación: diagnostico del editor sin errores en `web/js/seleccionar_empresa.js` y `web/estilos.css`.

- Checkout publico de licencias: resumen en tarjetas y medios Epayco visibles.
	- Archivos modificados: `web/pagar_licencia.html`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: `pagar_licencia.html` ahora separa la licencia y los codigos de descuento en dos tarjetas con el mismo estilo comercial del home, agrega un bloque visible con las formas de pago de Epayco y recolorea su panel de checkout con branding propio.
	- Verificación: diagnostico del editor sin errores en `web/pagar_licencia.html` y `web/estilos.css`.

- Elegir licencia: orden ascendente de menor a mayor valor.
	- Archivos modificados: `web/elegir_licencia.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el catalogo de licencias disponibles para pagar ahora se ordena en frontend desde el menor valor hasta el mayor, de modo que las opciones mas economicas aparezcan primero sin alterar el flujo hacia `pagar_licencia.html`.
	- Verificación: diagnostico del editor sin errores en `web/elegir_licencia.html`.

- Checkout publico de licencias: Epayco ahora activa la licencia, envia correo y cierra correctamente los estados finales.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `web/pagar_licencia.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el post-pago de Epayco resuelve contexto por `invoice` cuando la pasarela devuelve IDs externos, activa la licencia solo una vez, conserva `customer_email` para enviar el correo de confirmacion y evita que el formulario vuelva a marcar `pending` cuando el retorno ya es rechazado o fallido.
	- Verificación: `go test ./handlers -run "TestEpayco(TransactionStatusHandlerFindsContextUsingInvoiceWhenGatewayIDsDiffer|TransactionStatusHandlerActivatesOnceAndCapturesEmail|WebhookHandlerFindsContextUsingInvoiceFallback)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Elegir licencia: tarjetas mas compactas y sin textos de estado.
	- Archivos modificados: `web/elegir_licencia.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: las tarjetas de licencias disponibles para pagar dejan de mostrar `Estado: Activa/Inactiva` y `Disponible para asignacion inmediata` o su variante de asignacion. Ademas se compactan visualmente con menor padding, icono mas contenido y menos separacion vertical, manteniendo intacto el flujo hacia `pagar_licencia.html`.
	- Verificación: diagnostico del editor sin errores en `web/elegir_licencia.html`.

- Exportes operativos: descarga silenciosa sin sacar al usuario del módulo.
	- Archivos modificados: `web/administrar_empresa/administrar_clientes.html`, `web/administrar_empresa/asistencia_empleados.html`, `web/administrar_empresa/backups.html`, `web/administrar_empresa/tarifas_por_dia.html`, `web/administrar_empresa/soporte_remoto.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: los exportes frecuentes de clientes, asistencia, backups, tarifas por d?a y soporte remoto dejan de reemplazar la vista actual. El archivo se descarga en segundo plano y el usuario permanece en el mismo módulo.
	- Verificación: diagnostico del editor sin errores en los archivos modificados.

- Navegacion general: misma pesta?a por defecto.
	- Archivos modificados: `web/super_administrador.html`, `web/administrar_empresa.html`, `web/js/administrar_empresa.js`, `web/js/seleccionar_empresa.js`, `web/login.html`, `web/registrar_nuevo_usuario_administrador.html`, `web/registrar_contrasena_usuario_de_google.html`, `web/super/venta_digital.html`, `web/super/pagina_principal.html`, `web/super/configuracion_avanzada.html`, `web/administrar_empresa/venta_publica.html`, `web/administrar_empresa/soporte_remoto.html`, `web/super/soporte_remoto.html`, `web/administrar_empresa/administrar_clientes.html`, `web/administrar_empresa/asistencia_empleados.html`, `web/administrar_empresa/backups.html`, `web/administrar_empresa/tarifas_por_dia.html`, `web/administrar_empresa/chat_con_inteligencia_artificial.html`, `web/administrar_empresa/chat_y_tareas.html`, `web/index.html`, `web/Informacion_de_contacto.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la navegaci?n normal del sistema deja de abrir pesta?as nuevas y reutiliza la misma ventana actual. Se mantienen como excepci?n solo el contrato, los t?rminos legales de pasarela y los popups t?cnicos de impresión o vista previa documental.
	- Verificación: búsqueda final de `target="_blank"|window.open(` limitada a excepciones esperadas; diagn?stico del editor sin errores en los archivos modificados.

- Licencias super: valor 0 ya no se oculta en edici?n ni en listado.
	- Archivos modificados: `web/super/licencias.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el CRUD de licencias en panel super conserva `0` como valor valido visible en la tabla y en el formulario de edici?n, evitando que una licencia parezca vac?a al reabrirla.
	- Verificación: diagnostico del editor sin errores en `web/super/licencias.html`.

- Licencias del selector: historial con vencimiento y renovacion.
	- Archivos modificados: `backend/db/db.go`, `backend/handlers/payments_handlers_test.go`, `web/super/licencias.html`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la ruta `super/licencias.htmlscope=mine&con_empresa=1` deja de mostrar el CRUD y pasa a ser un historial de licencias pagadas o vencidas por empresa, con fecha de vencimiento visible, estados operativos y acceso a `Pagar nueva licencia` cuando la licencia esta por vencer o ya vencio. El backend reutiliza el mismo endpoint `/super/api/licencias` exponiendo empresa y fechas para ese flujo.
	- Verificación: diagnostico del editor sin errores en los archivos modificados; `go test ./handlers -run "TestLicenciasHandlerGetReturnsHistorialFieldsForCreatorScope" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Checkout publico de licencias: Epayco migra a Smart Checkout v2.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `web/pagar_licencia.html`, `web/super/configuracion_avanzada.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el sistema deja de generar URLs manuales hacia `checkout.php`, porque ese flujo ya responde `AccessDenied`. Ahora el backend crea la sesion oficial Smart Checkout v2 en Apify y el frontend abre `checkout-v2.js` con `sessionId`, manteniendo las mismas rutas publicas de respuesta, verificacion y webhook.
	- Verificación: `go test ./handlers -run 'TestEpaycoCreateTransactionHandler(UsesConfiguredPublicBaseURLAndKeys|AllowsCheckoutWithoutPrivateKey|AcceptsSamboxAlias)|TestEpaycoTransactionStatusHandler(PreservesPendingOnGenericValidationError|FindsContextUsingInvoiceWhenGatewayIDsDiffer)' -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Crear clave por correo: ojo para mostrar u ocultar la contrasena.
	- Archivos modificados: `web/registrar_contrasena_usuario_de_google.html`, `web/js/registrar_contrasena_usuario_de_google.js`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la pagina `Crear clave para acceso por correo` ahora incluye un icono de ojo en ambos campos de contrasena para poder revisarla visualmente antes de guardarla.
	- Verificación: diagnostico del editor sin errores en los archivos modificados.

- Elegir licencia: tarjetas con el mismo estilo del home.
	- Archivos modificados: `web/elegir_licencia.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la pagina `elegir_licencia.html` ahora renderiza las licencias con la misma estructura visual de tarjetas usada en `index.html`, manteniendo sin cambios el flujo de compra hacia `pagar_licencia.html`.
	- Verificación: diagnostico del editor sin errores en `web/elegir_licencia.html`.

- Reportes globales super: eleccion explicita de una empresa o varias.
	- Archivos modificados: `web/super/reportes_globales.html`, `web/js/super_reportes_globales.js`, `web/estilos.css`, `backend/handlers/reportes_globales_test.go`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el modulo `Reportes globales` ahora permite escoger de forma explicita si el analisis se hace sobre una sola empresa o sobre varias. En modo singular la UI cambia a selector puntual y el frontend consulta la API usando `empresa_id`.
	- Verificación: diagnostico del editor sin errores en los archivos modificados; `go test ./handlers -run "TestSuperReportesGlobalesHandlerFiltraYConsolidaPorAdministrador" -count=1`.

- Login administrativo: Google y correo quedan en una sola tarjeta visual.
	- Archivos modificados: `web/login.html`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el bloque de acceso por correo deja de renderizarse como un formulario en caja separado dentro de `login.html`. Google, correo, recuperaci?n y reset ahora comparten el mismo contenedor visual principal.
	- Verificación: diagnostico del editor sin errores en `web/login.html` y `web/estilos.css`.

- Arcade publico: runtime comun de poderes y premios en los nueve juegos activos.
	- Archivos modificados: `web/Juegos/arcade_shared.js`, `web/Juegos/arcade_window.css`, `web/Juegos/patito_volando_plus.html`, `web/Juegos/serpiente_pixel_plus.html`, `web/Juegos/memoria_estelar_plus.html`, `web/Juegos/rebote_bloques_plus.html`, `web/Juegos/pacman_plus.html`, `web/Juegos/tetris_plus.html`, `web/Juegos/carton_fire_plus.html`, `web/Juegos/ajedrez_vs_ia_plus.html`, `web/Juegos/ajedrez_3d_plus.html`, `web/Juegos/menu_juegos.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el arcade publico queda unificado con una misma capa mobile-first de countdown, sonido, records, poderes y premios para los nueve juegos activos del lobby, con economia compartida ajustada para juegos de eventos rapidos y un lobby que muestra mejor el progreso personal y el ranking por titulo.
	- Verificación: diagnostico del editor sin errores en `web/Juegos/arcade_shared.js`, `web/Juegos/arcade_window.css`, `web/Juegos/menu_juegos.html` y los cuatro juegos integrados en esta fase; busqueda de `createPowerSystem` presente en los 9 juegos activos.

## 2026-04-17
- Arcade publico: nuevo Ajedrez 3D plus con cinco dificultades.
	- Archivos creados: `web/Juegos/ajedrez_3d_plus.html`, `web/img/juegos/ajedrez_3d.svg`.
	- Archivos modificados: `web/Juegos/menu_juegos.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se agrega una nueva variante publica de ajedrez al arcade del portal, con tablero en perspectiva 3D simulada, cronometro arcade, cuenta regresiva de inicio y cinco niveles de dificultad contra la IA.
	- Verificación: diagnostico del editor sin errores en `web/Juegos/ajedrez_3d_plus.html` y `web/Juegos/menu_juegos.html`.

## 2026-04-17
- Reportes globales super: graficos y lectura ejecutiva.
	- Archivos modificados: `web/super/reportes_globales.html`, `web/js/super_reportes_globales.js`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la vista global del panel super ahora a?ade graficos comparativos y una lectura ejecutiva autom?tica del consolidado de empresas seleccionadas, sin cambiar el modelo de permisos ni crear dependencias frontend externas.
	- Verificación: diagnostico del editor sin errores en HTML/JS modificados.

## 2026-04-16
- Reportes globales super: consolidados por administrador creador.
	- Archivos modificados: `backend/db/db.go`, `backend/main.go`, `backend/handlers/reportes_globales.go`, `backend/handlers/reportes_globales_test.go`, `web/super/reportes_globales.html`, `web/js/super_reportes_globales.js`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la vista `Reportes globales` del panel super ahora permite consultar reportes generales, mezclados o individuales de las empresas creadas por el administrador autenticado, reutilizando los datasets empresariales existentes y manteniendo el aislamiento por creador.
	- Verificación: `go test ./handlers -run "TestSuperReportesGlobalesHandlerFiltraYConsolidaPorAdministrador" -count=1`; diagnostico del editor sin errores en los archivos nuevos y modificados.

## 2026-04-17
- Seleccionar empresa: licencia y descarga quedan en una sola fila.
	- Archivos modificados: `web/js/seleccionar_empresa.js`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el render de las tarjetas de seleccion de empresa ahora agrega el boton verde de descarga dentro del mismo bloque `card-actions` que usa el indicador de licencia, evitando que queden en filas separadas.
	- Verificación: diagnostico del editor sin errores en `web/js/seleccionar_empresa.js`.

## 2026-04-17
- Licencias super: actualizacion compatible con esquemas legacy sin `fecha_actualizacion`.
	- Archivos modificados: `backend/db/db.go`, `backend/db/licencias_schema_test.go`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la edicion y activacion de licencias en el panel super ya no fallan cuando la tabla `licencias` viene de un esquema antiguo que no incluye `fecha_actualizacion`. El backend intenta regularizar el esquema y, si esa columna sigue ausente, aplica un `UPDATE` de compatibilidad para guardar precio y estado.
	- Verificación: `go test ./db -run "TestEnsureLicenciasSchemaAddsValorInmotor legado retirado|TestCreateAndUpdateLicenciaRepairMissingValorColumn|TestUpdateLicenciaRepairsMissingFechaActualizacionColumn" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

## 2026-04-16
- Checkout publico de licencias: Epayco redirige la misma pesta?a al checkout.
	- Archivos modificados: `web/pagar_licencia.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el flujo de Epayco deja de depender de una pesta?a emergente y ya no arranca el polling antes de que el usuario entre a la pasarela. `pagar_licencia.html` guarda la referencia pendiente, redirige la misma pesta?a a Epayco y usa `/epayco/respuesta.html` para retomar la verificacion al volver.
	- Verificación: `GET /api/public/licencias/payment_methods` con `epayco.available=true`; `POST /epayco/create_transaction` con `checkout_url` publica valida; `GET /epayco/transaction_statusreference=<referencia_recien_creada>` con `PENDING` y `context_found=true`; diagnostico del editor sin errores en `web/pagar_licencia.html`.

- Menu flotante: separaci?n frente a botones superiores cercanos.
	- Archivos modificados: `web/menu.js`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el menu flotante compartido ahora reserva espacio en encabezados y barras de acciones para no montarse sobre botones ubicados en la parte superior derecha de algunas paginas.
	- Verificación: diagnostico del editor sin errores en los archivos modificados.

## 2026-04-16
- Facturacion electronica: suite `db` estable aun con entorno local en PostgreSQL.
	- Archivos modificados: `backend/db/finanzas_test.go`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: `openFinanzasTestDB` ahora fija el dialecto `motor_legado_retirado` para evitar que la suite de facturacion electronica y documentos transaccionales herede `DB_DIALECT=postgres` del entorno local y falle con SQL incompatible.
	- Verificación: `go test ./db -run "Test.*(Facturacion|DIAN|DocumentoFacturacion)" -count=1`; `go test ./handlers -run "Test(VentaCarritoFacturaYResolucionImpresora|EmpresaDIANColombiaHandler.*|EmpresaFacturacionElectronicaReintentosYReconciliacion|EmpresaFacturacionElectronicaEmiteEventoContable|EmpresaFacturacionTransaccional.*)" -count=1`.

- Pagina principal super: el campo de cantidad deja de mostrar un `5` temporal antes de cargar la configuracion real.
	- Archivos modificados: `web/super/pagina_principal.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el editor de `pagina_principal` ahora deja `ppCantidad` en estado de carga hasta recibir la configuracion persistida y sincroniza la cantidad con el numero real de tarjetas, evitando la confusion visual entre el panel super, `index.html` y `/descripcion_de_los_sistemas.ht`.
	- Verificación: consulta local a `/api/public/pagina_principal` con `cantidad=7`; revision directa del flujo de carga del editor super.

## 2026-04-17
- Ventas y facturacion: prueba integrada de carrito pagado con resolución de impresora.
	- Archivos creados: `backend/handlers/carrito_facturacion_impresion_test.go`.
	- Archivos modificados: `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se agrega una prueba de integracion de handlers que valida una venta pagada en carrito, la emision documental de factura electronica y la resolución de la impresora `factura_caja` para el flujo de impresion soportado hoy.
	- Verificación: `go test ./handlers -run TestVentaCarritoFacturaYResolucionImpresora -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Seleccionar empresa: restauracion del formato clasico de tarjetas.
	- Archivos modificados: `web/js/seleccionar_empresa.js`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el selector de empresas del panel super vuelve al formato simple de tarjetas `portal-card warm` usado anteriormente, retirando la presentacion enriquecida reciente.
	- Verificación: revision del render en `web/js/seleccionar_empresa.js`; recomendada validacion visual en `seleccionar_empresa.html`.

- Portal publico: menu flotante navegable en celular.
	- Archivos modificados: `web/menu.js`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: manipulation`.
	- Verificación: revision del flujo JS/CSS; recomendada validacion manual en movil o emulacion tactil.

- Usuarios de empresa: portal publico con contrato vigente y subdominio dedicado.
	- Archivos modificados: `backend/db/usuarios_empresa.go`, `backend/handlers/usuarios_empresa.go`, `backend/handlers/auth_users_carritos_test.go`, `backend/main.go`, `web/login_usuario.html`, `web/js/login_usuario.js`, `web/estilos.css`, `web/administrar_empresa.html`, `web/js/administrar_empresa.js`, `web/administrar_empresa/administrar_usuarios.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: `login_usuario.html` pasa a ser el portal publico de usuarios internos creados por administradores, con registro por invitacion, recuperacion, reset, cambio de contrasena y aceptacion obligatoria del contrato vigente. El backend persiste esa aceptacion en `users`, los correos y el panel administrativo apuntan a `usuarios.powerfulcontrolsystem.com`, y el acceso final sigue entrando a `administrar_empresa.html` filtrado por rol.
	- Verificación: `go test ./handlers -run "TestEmpresaUsuario(LoginHandlerSuccess|LoginHandlerRequiresContractAcceptance|SetPasswordHandlerSuccess|ResolveEmpresaUsuarioLoginURLUsesUsuariosSubdomain)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`; diagnostico del editor sin errores en los archivos web modificados.

- Usuarios de empresa: login por subdominio propio de cada empresa.
	- Archivos modificados: `backend/handlers/usuarios_empresa.go`, `backend/handlers/auth_users_carritos_test.go`, `backend/handlers/usuarios_empresa_seguridad_test.go`, `web/administrar_empresa.html`, `web/js/administrar_empresa.js`, `web/administrar_empresa/administrar_usuarios.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el enlace operativo del login de usuarios deja de resolverse a un host global fijo. Ahora se construye con el `empresa_slug` o `dominio_publico` configurado por empresa, tanto en el menu de `administrar_empresa` como en los correos de invitacion y recuperacion; la vista de administrar usuarios elimina el acceso duplicado fuera del menu.
	- Verificación: `go test ./handlers -run "TestEmpresaUsuario(LoginHandlerSuccess|LoginHandlerRequiresContractAcceptance|SetPasswordHandlerSuccess|ResolveEmpresaUsuarioLoginURLUsesEmpresaSubdomain|PasswordRecoveryFlow|ChangePasswordFlow|ChangePasswordPolicyRejectsWeakPassword|LoginRequiresRotationWhenPolicyEnabled|NotificationsCaptureInMailTestMode)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`; diagnostico del editor sin errores en los archivos modificados.

- Soporte remoto: limites por plan y mesa tecnica central multiempresa.
	- Archivos creados: `backend/handlers/super_soporte_remoto.go`, `backend/handlers/super_soporte_remoto_test.go`, `web/super/soporte_remoto.html`.
	- Archivos modificados: `backend/db/soporte_remoto.go`, `backend/db/soporte_remoto_test.go`, `backend/handlers/soporte_remoto.go`, `backend/handlers/soporte_remoto_test.go`, `backend/handlers/system_empresas_handlers_test.go`, `backend/main.go`, `web/super_administrador.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/estructura_bd.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el modulo de soporte remoto ahora controla cupos de dispositivos, sesiones y minutos por empresa, persiste consumo mensual con intentos bloqueados y agrega una mesa tecnica central para `super_administrador` en `/super/api/soporte_remoto` y `super/soporte_remoto.html`.
	- Verificación: `go test ./db ./handlers -run "Test(SoporteRemotoDB|EmpresaSoporteRemotoHandler|PublicSoporteRemotoAgentHeartbeatAndStateUpdate|SuperSoporteRemotoHandlerListsCompaniesAndCreatesSession|SuperEndpointsPermisosPorRol)" -count=1`.

## 2026-04-16
- Arranque local: healthcheck robusto en `scripts/iniciar_servidor.ps1`.
	- Archivos modificados: `scripts/iniciar_servidor.ps1`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el paso `8/8` deja de reportar timeout falso cuando el backend ya esta arriba. El script ahora usa el `PORT` efectivo, detecta listener TCP con API nativa/fallback y acepta una respuesta HTTP valida para confirmar disponibilidad.
	- Verificación: `. 'D:\powerfulcontrolsystem\scripts\iniciar_servidor.ps1'`.

- Backend: fix de compilacion en soporte remoto y bootstrap runtime.
	- Archivos modificados: `backend/db/soporte_remoto.go`, `backend/db/productos.go`, `backend/main.go`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se restaura el arranque local del backend corrigiendo la variable temporal usada al leer sesiones de soporte remoto, reescribiendo el cierre del bloque runtime en `main.go` y haciendo idempotente la regularizacion de columnas en PostgreSQL para evitar errores `column already exists` durante `scripts/iniciar_servidor.ps1`.
	- Verificación: `go build -o server.exe .` en `backend`; `.\scripts\iniciar_servidor.ps1 -Background`.

- Estaciones: sincronizacion backend del carrito base por estacion.
	- Archivos modificados: `backend/db/empresa_estacion_prefs.go`, `backend/db/empresa_estacion_prefs_test.go`, `backend/handlers/empresa_estacion_prefs.go`, `backend/handlers/empresa_estacion_prefs_test.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: guardar `estaciones_config` ya no depende del frontend para crear los carritos por defecto. El backend sincroniza automaticamente un carrito enlazado por estacion, corrige nombre/codigo/referencia cuando cambia la configuracion y lo deja en estado base `inactivo/cerrado` hasta su activacion operativa.
	- Verificación: `go test -work ./db -run "Test(EmpresaEstacionPrefs|SyncEmpresaEstacionCarritos)" -count=1`; `go test -work ./handlers -run "TestEmpresaEstacionPrefsHandler_UpsertAndIsolationByEmpresa|TestEmpresaCarritosCompraMetricasEstacionIncluyeCorrecciones" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Home publico: contacto centrado debajo de las tarjetas y deploy VPS con limpieza de procesos previos.
	- Archivos modificados: `web/index.html`, `web/estilos.css`, `scripts/sync_to_vps.ps1`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el home del portal deja `Informacion de contacto` como CTA centrado debajo del grid de tarjetas, manteniendo `Registrarse o iniciar sesión` en la cabecera. En paralelo, el deploy remoto endurece el reinicio del backend: purga procesos viejos de `server_linux_amd64`, corrige la unidad `systemd` para evitar el warning de `StartLimitIntervalSec` mal ubicado y asegura que el binario nuevo quede activo al terminar `sync_to_vps`.
	- Verificación: `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`; validacion de sintaxis PowerShell para `scripts/sync_to_vps.ps1`; diagnostico remoto de `systemctl status powerfulcontrolsystem` y `ss -ltnp` en el VPS.

- Checkout publico de licencias: Epayco acepta alias sambox como sandbox.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la normalizacion del modo de Epayco ahora tolera `sambox` como alias de `sandbox`, garantizando que el checkout de licencias permanezca en pruebas (`test=true`) aunque la configuracion manual use esa variante.
	- Verificación: `go test ./handlers -run 'TestEpaycoCreateTransactionHandler(AcceptsSamboxAlias|UsesConfiguredPublicBaseURLAndKeys|AllowsCheckoutWithoutPrivateKey)|TestResolvePaymentBaseURL(FallsBackToCanonicalDomainOnLocalhost|UsesConfiguredCanonicalDomain|IgnoresConfiguredLocalhostAndFallsBackToCanonicalDomain)|TestEpaycoTransactionStatusHandler(PreservesPendingOnGenericValidationError|FindsContextUsingInvoiceWhenGatewayIDsDiffer)' -count=1`.

- Checkout publico de licencias: Epayco legacy + metodo unico sin selector.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `web/pagar_licencia.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el checkout de Epayco vuelve a enviar `p_key` cuando la configuracion dispone de `private_key`, manteniendo compatibilidad con cuentas que exigen parametros legacy en `checkout.php`; ademas, `pagar_licencia.html` ya no muestra el cuadro de seleccion de forma de pago cuando solo una pasarela esta activa.
	- Verificación: pendiente ejecutar pruebas focalizadas de handlers y confirmar que la URL remota de checkout deje de responder `403 AccessDenied`.

- Arcade publico: set activo de ocho juegos compactos con popup fijo y pausa real.
	- Archivos creados: `web/Juegos/arcade_window.css`, `web/Juegos/patito_volando_plus.html`, `web/Juegos/serpiente_pixel_plus.html`, `web/Juegos/memoria_estelar_plus.html`, `web/Juegos/rebote_bloques_plus.html`, `web/Juegos/pacman_plus.html`, `web/Juegos/tetris_plus.html`, `web/Juegos/carton_fire_plus.html`, `web/Juegos/ajedrez_vs_ia_plus.html`, `web/img/juegos/pacman.svg`, `web/img/juegos/tetris.svg`, `web/img/juegos/carton_fire.svg`, `web/img/juegos/ajedrez_vs_ia.svg`.
	- Archivos eliminados: `web/Juegos/pollitos_cataplum.html`, `web/img/juegos/pollitos_cataplum.svg`.
	- Archivos modificados: `web/Juegos/menu_juegos.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el arcade publico deja de operar con el set anterior y pasa a un lobby de ocho juegos activos con records compartidos por navegador, popup uniforme `700x700` en escritorio y pausa real en todas las experiencias, incluyendo congelacion de IA u oponentes cuando aplica.
	- Verificación: diagnostico del editor sin errores en `web/Juegos/menu_juegos.html` y en los ocho archivos `*_plus.html` del nuevo arcade.

- Home público: botones superiores m?s compactos y centrados en m?vil.
	- Archivos modificados: `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: los botones `Registrarse o iniciar sesión` e `Informacion de contacto` del `index.html` ahora comparten un ancho m?s peque?o, menor altura visual y en celular se muestran centrados dentro del header.
	- Verificación: diagnostico del editor sin errores en `web/estilos.css`.

- Licencias super: autorreparaci?n del esquema y validación real de guardado del valor.
	- Archivos modificados: `backend/db/db.go`, `backend/db/sql_compat.go`, `backend/db/licencias_schema_test.go`, `backend/main.go`, `web/super/licencias.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el backend ahora regulariza la tabla `licencias` tambien en PostgreSQL y reintenta `create/get/update` si faltan columnas como `valor`; la UI de super deja de ocultar errores HTTP al crear/editar licencias, mostrando el mensaje real cuando el backend rechaza la operación.
	- Verificación: `go test ./db -run "TestEnsureLicenciasSchemaAddsValorInmotor legado retirado|TestCreateAndUpdateLicenciaRepairMissingValorColumn" -count=1`.

- Seleccionar empresa: tarjetas adaptables con contenido interno completo y m?rgenes estrechos.
	- Archivos modificados: `web/js/seleccionar_empresa.js`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la vista `seleccionar_empresa.html` pasa a renderizar tarjetas con estructura interna avanzada (`empresa-card`) y estilos flexibles que permiten envolver títulos, descripciones, estados y metadatos sin cortar contenido. Se mantienen m?rgenes peque?os y el interior se adapta autom?ticamente al texto disponible.
	- Verificación: diagnostico del editor sin errores en `web/js/seleccionar_empresa.js` y `web/estilos.css`.

- Super pagina principal: el editor ya no recorta tarjetas cargadas por usar el valor inicial del input de cantidad.
	- Archivos modificados: `web/super/pagina_principal.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se corrige el render del editor super de `pagina_principal` para que, al recargar configuraciones con mas de 5 tarjetas, use primero `state.config.cantidad` y no el valor HTML inicial del campo `ppCantidad`. Con esto vuelven a mostrarse las 7 tarjetas guardadas y la cantidad visible queda sincronizada con la API.
	- Verificación: inspeccion de `GET https://powerfulcontrolsystem.com/api/public/pagina_principal` con `cantidad=7`; diagnostico del editor sin errores en `web/super/pagina_principal.html`.

- Infraestructura publica: wildcard HTTPS manual para subdominios y subdominio dedicado de prueba para venta digital.
	- Archivos modificados: `documentos/manual_de_instalacion.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se documenta la emision manual del certificado wildcard `powerfulcontrolsystem.com` + `*.powerfulcontrolsystem.com`, la pauta de renovacion manual por DNS-01 y la publicacion del subdominio de prueba `venta-digital.powerfulcontrolsystem.com` hacia la pagina publica global `venta_digital.html`.
	- Verificación: HTTPS `200` en `https://powerfulcontrolsystem.com/`, `301` de `https://www.powerfulcontrolsystem.com/` a apex, `302` de `https://venta-digital.powerfulcontrolsystem.com/` a `/venta_digital.html` y `200` final en `https://venta-digital.powerfulcontrolsystem.com/venta_digital.html`.

- Registro administrativo: captura de pais y ciudad con deteccion inicial de pais en frontend.
	- Archivos modificados: `backend/db/db.go`, `backend/handlers/auth_admin_handlers.go`, `backend/handlers/account_handlers.go`, `backend/handlers/auth_admin_handlers_test.go`, `backend/db/administradores_auth_schema_test.go`, `web/registrar_nuevo_usuario_administrador.html`, `web/js/registrar_nuevo_usuario_administrador.js`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el registro de administradores ahora solicita correo, nombre completo, celular, pais y ciudad. El pais se sugiere automaticamente desde el navegador/zona horaria y sigue siendo editable. El backend persiste `pais` y `ciudad` en `administradores`, y se mantiene la exigencia de confirmar el correo antes de continuar al flujo de acceso que luego lleva a `seleccionar_empresa.html`.
	- Verificación: `go test ./db ./handlers -run 'Test(AdminRegisterHandlerCreatesPendingAdminAndCapturesConfirmationMail|AdminRegisterHandlerRejectsConfirmedExistingAdmin|EnsureAdministradoresAuthSchemaAddsMissingColumnsInmotor legado retirado|SetAdministradorPasswordRepairsMissingSecurityColumns)$' -count=1`.

- Autenticacion administrativa: compatibilidad del esquema `administradores` entre motor legado retirado y PostgreSQL.
	- Archivos creados: `backend/db/administradores_auth_schema_test.go`.
	- Archivos modificados: `backend/db/db.go`, `backend/main.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se centraliza la regularizacion de columnas de seguridad de `administradores` con soporte para motor legado retirado y PostgreSQL mediante `EnsureAdministradoresAuthSchema`, y `SetAdministradorPassword` reintenta la operacion cuando encuentra columnas faltantes. Con esto se corrige el flujo donde una cuenta autenticada por Google no podia registrar su primera contrasena local en VPS con PostgreSQL.
	- Verificación: `go test ./db ./handlers -run 'Test(EnsureAdministradoresAuthSchemaAddsMissingColumnsInmotor legado retirado|SetAdministradorPasswordRepairsMissingSecurityColumns|AccountSetGooglePasswordHandlerCreatesInitialPassword)$' -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`; validacion operativa en VPS de `systemd`, `Nginx`, `UFW`, callback OAuth y dominio publico.

- Super: modulo de seguridad VPS Linux con panel, CLI, cron y exportes multiformato.
	- Archivos creados: `backend/vpssecurity/config/config.go`, `backend/vpssecurity/config/default_vps_security_config.json`, `backend/vpssecurity/parser/lynis.go`, `backend/vpssecurity/parser/nmap.go`, `backend/vpssecurity/parser/trivy.go`, `backend/vpssecurity/scanner/runner.go`, `backend/vpssecurity/scanner/checks.go`, `backend/vpssecurity/reports/report.go`, `backend/vpssecurity/reports/report_test.go`, `backend/vpssecurity/logs/store.go`, `backend/vpssecurity/service.go`, `backend/handlers/security_vps_handlers.go`, `backend/handlers/security_vps_handlers_test.go`, `backend/tools/vps_security_scan/main.go`, `web/js/super_seguridad.js`, `scripts/install_vps_security_tools.sh`, `scripts/run_vps_security_scan.sh`, `scripts/install_vps_security_cron.sh`, `documentos/manual_vps_seguridad.md`.
	- Archivos modificados: `backend/main.go`, `web/super/seguridad.html`, `web/index.html`, `web/estilos.css`, `.gitignore`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se agrega un modulo completo de seguridad VPS Linux para `super_administrador`, con ejecucion de Lynis/Nmap/Trivy y chequeos propios, historial/comparacion en filesystem, exportes `JSON/TXT/HTML/CSV/PDF/XLS`, CLI reutilizable y scripts Ubuntu para instalacion y cron. En el portal publico, `Informacion de contacto` queda anclado al extremo derecho de la misma fila superior del home.
	- Verificación: `go test ./vpssecurity/... ./handlers ./tools/vps_security_scan -run "TestSecurityVPS|TestGenerateArtifacts|TestCompareDetects" -count=1`; diagn?stico del editor sin errores en los archivos Go/HTML/JS/SH modificados para este cambio.

- Login unificado: eliminado `recordar usuario/cuenta` y retirado `login_hint` en OAuth.
	- Archivos modificados: `web/login.html`, `web/js/login.js`, `web/login_usuario.html`, `web/js/login_usuario.js`, `web/menu.js`, `web/js/super_administrador.js`, `web/js/seleccionar_empresa.js`, `web/super/licencias.html`, `web/super/tipos_empresas.html`, `backend/handlers/auth_admin_handlers.go`, `backend/handlers/auth_users_carritos_test.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se elimina la persistencia local de cuenta/usuario recordado en ambos logins y se deja `/auth/google/login` sin `login_hint`. Con esto, el acceso queda consistente entre `localhost:8080`, `powerfulcontrolsystem.com` y `www.powerfulcontrolsystem.com`, sin depender de estado guardado por dominio en `localStorage`.
	- Verificación: `go test -work ./handlers -run "TestHandleGoogleLoginRedirect|TestAccountSetGooglePasswordHandlerCreatesInitialPassword|TestE2E_AcceptContractCreatesSession|TestAdminLoginHandlerCreatesSessionForConfirmedAdmin" -count=1`.

- Login Google: registro obligatorio de contrasena local cuando falta password_set.
	- Archivos creados: `web/registrar_contrasena_usuario_de_google.html`, `web/js/registrar_contrasena_usuario_de_google.js`.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/handlers/accept_handlers.go`, `backend/handlers/account_handlers.go`, `backend/main.go`, `backend/handlers/auth_admin_handlers_test.go`, `backend/handlers/e2e_login_acceptance_test.go`, `web/ayuda/login_administradores.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el callback Google y la aceptaci?n de contrato ya no envían al panel final cuando la cuenta administrativa a?n no tiene contrase?a local; ahora redirigen a `registrar_contrasena_usuario_de_google.html`, que guarda la primera clave mediante `/api/account/set_google_password` para habilitar despu?s el acceso por correo y contrase?a.
	- Verificación: `go test -work ./handlers -run "Test(AccountSetGooglePasswordHandlerCreatesInitialPassword|E2E_AcceptContractCreatesSession|AdminLoginHandlerCreatesSessionForConfirmedAdmin)" -count=1`.

- Super: panel PostgreSQL con carga de tama?o por empresa.
	- Archivos modificados: `backend/handlers/postgres_performance.go`, `backend/handlers/postgres_performance_test.go`, `web/super/administrar_base_de_datos.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el panel super de administracion PostgreSQL ahora puede cargar bajo demanda una tabla de consumo estimado por empresa dentro de `pcs_empresas`, ordenada de mayor a menor y mostrando tambien filas estimadas, tablas con datos y la tabla mas pesada por empresa.
	- Verificación: `go test -work ./handlers -run "TestPostgresPerformanceHandler|TestHumanizeBytesBinary" -count=1`.

- Manual de instalacion: agregado el paso de respuesta, confirmacion y formulario exacto de Epayco.
	- Archivos modificados: `documentos/manual_de_instalacion.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el manual ya incluye las URLs exactas que deben configurarse en Epayco para respuesta y confirmacion, ademas de los valores concretos del formulario de Epayco y una nota operativa sobre el flujo real de validacion del pago.

- Checkout y seleccion de empresa: ajuste visual solicitado.
	- Archivos modificados: `web/pagar_licencia.html`, `web/js/seleccionar_empresa.js`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: `pagar_licencia.html` deja mas clara la pasarela activa cuando solo hay un metodo disponible y muestra el logo de Epayco en el selector y en el panel. `seleccionar_empresa.html` vuelve al estilo compacto anterior de tarjetas para empresas.

- Checkout de licencias: Epayco ahora usa una pagina publica fija de respuesta.
	- Archivos creados: `web/epayco/respuesta.html`.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el retorno de Epayco ya no depende de enviar al usuario directamente a `pagar_licencia.html`; ahora existe la landing publica fija `/epayco/respuesta.html`, que puedes registrar en el panel de Epayco y que reenvia al resumen del pago con el contexto necesario para validar y activar la licencia.
	- Verificación: `go test -work ./handlers -run "TestEpaycoCreateTransactionHandlerUsesConfiguredPublicBaseURLAndKeys|TestEpaycoCreateTransactionHandlerAllowsCheckoutWithoutPrivateKey|TestEpaycoTransactionStatusHandlerPreservesPendingOnGenericValidationError|TestResolvePaymentBaseURL" -count=1`.
- Login administrativo: registro separado, confirmación pública corregida y recuperaci?n sin prompts.
	- Archivos creados: `web/registrar_nuevo_usuario_administrador.html`, `web/js/registrar_nuevo_usuario_administrador.js`, `backend/handlers/auth_admin_handlers_test.go`.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/utils/utils.go`, `web/login.html`, `web/js/login.js`, `web/estilos.css`, `web/ayuda/login_administradores.html`, `backend/handlers/auth_users_carritos_test.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el login administrativo ahora deja el registro en una página pública espec?fica, elimina el campo de nombre incrustado del acceso principal, centra `Iniciar por correo` y agrega debajo `Registrarse` y `Ãƒâ€šÃ‚Â¿Olvid? su contrase?a`. El backend valida `nombre`, `telefono` y contrase?a segura, evita sobrescribir cuentas confirmadas, corrige el whitelist público para `/auth/confirmar_admin` y sustituye la recuperaci?n por formularios reales dentro de `login.html`.
	- Verificación: `go test -work ./handlers -run "Test(Admin|AuthMiddlewareAllowsPublicPortalPagesAssetsAndHomeCardsAPI|HandleGoogleLogin|E2E_AcceptContractCreatesSession)" -count=1`; `go test -work ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1`.

- Arcade publico: Patito volando ahora inicia con cuenta regresiva y los cinco juegos refuerzan su modo celular.
	- Archivos modificados: `web/Juegos/arcade_shared.js`, `web/Juegos/patito_volando.html`, `web/Juegos/pollitos_cataplum.html`, `web/Juegos/serpiente_pixel.html`, `web/Juegos/memoria_estelar.html`, `web/Juegos/rebote_bloques.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el arcade publico mantiene sonido compartido en los cinco juegos, `Patito volando` arranca con cuenta regresiva de 5 segundos y el resto del arcade ajusta shells, overlays y acciones para celular. Tambien se agregan sonidos de countdown en `arcade_shared.js` y `Serpiente pixel` suma feedback sonoro al giro durante la partida.
	- Verificación: validacion sin errores de los seis archivos del arcade modificados y revision de los nuevos breakpoints moviles y del countdown previo al inicio en `Patito volando`.

- Frontend compartido: mejoras base de adaptacion movil y menu flotante.
	- Archivos modificados: `web/menu.js`, `web/estilos.css`, `web/login.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el menu flotante ahora se cierra al seleccionar una opcion, el CTA de WhatsApp de la portada pasa a un icono compacto abajo a la derecha en movil para no tapar otros botones, la capa CSS compartida mejora tablas/sidebar/panel flotante en pantallas pequenas y `login.html` vuelve a cargar la hoja `estilos.css` correcta.
	- Verificación: validacion sin errores de `web/menu.js`, `web/estilos.css` y `web/login.html`, mas revision de los breakpoints moviles del menu flotante y del CTA de WhatsApp.

- Portal publico: botones superiores de la portada ahora usan el mismo estilo de Explorar oferta.
	- Archivos modificados: `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: los accesos `Registrarse o iniciar sesión` e `Informacion de contacto` del encabezado de `index.html` reutilizan la misma apariencia visual del boton `Explorar oferta` de las tarjetas del home, sin cambiar rutas ni comportamiento responsive.
	- Verificación: revision del bloque compartido de selectores en `web/estilos.css` y del ajuste pill en `@media (max-width:560px)`.

- Checkout de licencias: Epayco sandbox estable con bootstrap PostgreSQL y polling pendiente consistente.
	- Archivos modificados: `backend/db/db.go`, `backend/main.go`, `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el backend asegura `pagos_epayco` y `pagos_wompi` al arrancar sobre PostgreSQL y deja de degradar a `ERROR` una referencia de Epayco que sigue `PENDING` localmente mientras la validacion externa responde un error transitorio. Ademas se normaliza la configuracion sandbox operativa (`epayco.*` y `gmail.confirm_base_url`) en la base super del VPS para que el checkout genere callbacks publicos validos.
	- Verificación: `go test ./handlers -run 'TestResolvePaymentBaseURL|TestEpayco(CreateTransactionHandlerUsesConfiguredPublicBaseURLAndKeys|CreateTransactionHandlerAllowsCheckoutWithoutPrivateKey|TransactionStatusHandlerPreservesPendingOnGenericValidationError)' -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`; validacion manual local de `GET /api/public/licencias/payment_methods`, `POST /epayco/create_transaction` y `GET /epayco/transaction_status` tras recompilar con `scripts/iniciar_servidor.ps1 -Background`.

- Portal publico: arcade con cinco juegos, tarjetas cuadradas y perfil compartido.
	- Archivos creados: `web/Juegos/arcade_shared.js`, `web/Juegos/serpiente_pixel.html`, `web/Juegos/memoria_estelar.html`, `web/Juegos/rebote_bloques.html`, `web/img/juegos/patito_volando.svg`, `web/img/juegos/pollitos_cataplum.svg`, `web/img/juegos/serpiente_pixel.svg`, `web/img/juegos/memoria_estelar.svg`, `web/img/juegos/rebote_bloques.svg`.
	- Archivos modificados: `web/Juegos/menu_juegos.html`, `web/Juegos/patito_volando.html`, `web/Juegos/pollitos_cataplum.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el lobby publico de Juegos se convierte en un arcade visual con portadas cuadradas, panel de jugador y resumen de records. `arcade_shared.js` centraliza nombre, top local y sonido; `Patito volando` y `Pollitos al cataplum` se integran a esa capa y se agregan tres juegos nuevos: `Serpiente pixel`, `Memoria estelar` y `Rebote de bloques`.
	- Verificación: diagnostico del editor sin errores en `web/Juegos/arcade_shared.js`, `web/Juegos/menu_juegos.html`, `web/Juegos/patito_volando.html`, `web/Juegos/pollitos_cataplum.html`, `web/Juegos/serpiente_pixel.html`, `web/Juegos/memoria_estelar.html` y `web/Juegos/rebote_bloques.html`.

## 2026-04-15
- Portal publico: nuevo juego `Pollitos al cataplum` y menu de Juegos multi-tarjeta.
	- Archivos creados: `web/Juegos/pollitos_cataplum.html`.
	- Archivos modificados: `web/Juegos/menu_juegos.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se agrega un segundo juego publico de resortera con niveles cortos, puntaje y control arrastrar/soltar; ademas, el catalogo de Juegos ahora soporta varias tarjetas con popup propio por juego.

- Licencias: Epayco/Wompi ya no fallan por resolver `localhost` al iniciar checkout.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: `resolvePaymentBaseURL(...)` ahora ignora loopback en configuracion o request, intenta `gmail.confirm_base_url`, `Origin`/`Referer`, host publicado y, si hace falta, cae al dominio canonico `https://powerfulcontrolsystem.com` para construir callbacks publicos validos del checkout.
	- Verificación: `go test ./handlers -run "Test(ResolvePaymentBaseURLFallsBackToCanonicalDomainOnLocalhost|ResolvePaymentBaseURLUsesConfiguredCanonicalDomain|ResolvePaymentBaseURLIgnoresConfiguredLocalhostAndFallsBackToCanonicalDomain|EpaycoCreateTransactionHandlerUsesConfiguredPublicBaseURLAndKeys|EpaycoCreateTransactionHandlerAllowsCheckoutWithoutPrivateKey)" -count=1`.

- Servidor: alerta de inicio/reinicio ahora puede activarse o desactivarse desde configuracion avanzada.
	- Archivos modificados: `backend/handlers/server_runtime_notifications.go`, `backend/handlers/server_runtime_notifications_test.go`, `backend/handlers/usuarios_empresa.go`, `backend/handlers/system_empresas_handlers_test.go`, `backend/handlers/super_config_backup_handlers.go`, `web/super/configuracion_avanzada.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el backend ya registraba el arranque/reinicio en `super_servidor_eventos`, en `backend/logs/server_reinicio.log` y enviaba correo cuando existia `gmail.restart_alert_to`. Ahora se agrega `gmail.restart_alert_enabled` para activar o desactivar ese correo desde `configuracion_avanzada.html` sin perder el destinatario configurado.
	- Verificación: `go test ./handlers -run "Test(GmailConfigHandlerSaveRestartAlertTo|GmailConfigHandlerSaveRestartAlertToggle|RegisterServerStartupEventCapturesNotificationAndState|RegisterServerStartupEventDetectsUnexpectedRestart|RegisterServerStartupEventSkipsEmailWhenAlertsDisabled)" -count=1`.

- Seleccion de empresas: tarjetas con iconografia por tipo y redise?o mas profesional.
	- Archivos modificados: `web/js/seleccionar_empresa.js`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: `seleccionar_empresa.html` ahora presenta cada empresa con icono segun `tipo_nombre`, tono visual por categoria, chips de estado y una tarjeta mas colorida/profesional. Se conserva el mismo flujo para abrir la administracion o continuar con la licencia.
	- Verificación: diagnostico del editor sin errores en `web/js/seleccionar_empresa.js` y `web/estilos.css`.

- Pagina principal: la cantidad de tarjetas ahora se aplica y se guarda en un solo flujo.
	- Archivos modificados: `web/super/pagina_principal.html`, `backend/handlers/pagina_principal_handlers_test.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se elimina el paso manual `Aplicar cantidad` del editor super de `pagina_principal`. Al cambiar la cantidad, el editor reconstruye las tarjetas visibles y el mismo flujo de `Guardar configuracion` persiste cantidad, contenido y estilos. Ademas se agrega una prueba de persistencia para configuraciones ampliadas.
	- Verificación: `go test ./handlers -run "TestPaginaPrincipal|TestPublicPaginaPrincipalHandlerExposesLandingFields" -count=1`.

- Portal publico: nuevo menu de juegos y primer juego `Patito volando`.
	- Archivos creados: `web/Juegos/menu_juegos.html`, `web/Juegos/patito_volando.html`.
	- Archivos modificados: `backend/utils/utils.go`, `backend/handlers/auth_users_carritos_test.go`, `web/menu.js`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se agrega la entrada `Juegos` al menu flotante del portal, se crea `menu_juegos.html` con una tarjeta por juego publicado y se implementa `patito_volando.html` como minijuego de ventana pequena con control por barra espaciadora en PC y toque/presion en movil. `AuthMiddleware` deja publico `/Juegos/*` y la prueba del middleware se ampl?a para cubrir estas rutas.
	- Verificación: `go test ./handlers -run "TestAuthMiddlewareAllowsPublicPortalPagesAssetsAndHomeCardsAPI" -count=1`.

## 2026-04-15
- Repositorio: restaurado `Pendiente Notas` y auditados los borrados actuales.
	- Archivos modificados: `Pendiente Notas`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se recupera `Pendiente Notas` desde `HEAD` tras detectar que Git lo tenia como borrado local en el arbol de trabajo. La auditoria posterior confirma que no habia otros archivos eliminados en el estado actual del repo. Git no conserva una hora exacta para ese borrado local no confirmado; la ultima hora verificable en historial para el archivo es `2026-04-15 17:37:25 -0500` en el commit `e70884dabea1292d9c0e6d9b1a3f236e94d7c8c4`.
	- Verificación: `git diff --name-status --diff-filter=D`; `git status --short --untracked-files=no`; `git log -1 --format="%H%n%an%n%ad%n%s" -- "Pendiente Notas"`; `Get-Item -LiteralPath "d:\powerfulcontrolsystem\Pendiente Notas" | Select-Object FullName,Length,CreationTime,LastWriteTime`.

- Errores del sistema: monitor centralizado, recovery global y panel super.
	- Archivos creados: `backend/db/super_errores_sistema.go`, `backend/utils/system_errors.go`, `backend/handlers/super_error_handlers.go`, `backend/handlers/super_error_handlers_test.go`, `web/super/errores.html`.
	- Archivos modificados: `backend/main.go`, `backend/utils/utils.go`, `backend/utils/utils_test.go`, `web/super_administrador.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/estructura_bd.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se implementa un sistema robusto de manejo de errores para todo el proyecto. Los errores HTTP y panicos recuperados se registran en `super_errores_sistema` y en `backend/logs/system_errors.log`, el cliente deja de recibir detalles tecnicos en respuestas `5xx` y super obtiene un panel profesional para monitoreo transversal por empresa, fecha, severidad y tipo.
	- Verificación: `go test ./utils -run "Test(JSONErrorMiddlewareSanitizesInternalServerError|RecoveryMiddlewareRecoversPanicAndLogsIt|JSONErrorMiddlewarePreservesJSONErrorBody|JSONErrorMiddlewareWrapsNonJSONError|JSONErrorMiddlewareWrapsEpaycoNonJSONError)" -count=1`; `go test ./handlers -run "Test(SuperErroresSistemaHandlerFiltersResults|SuperErroresSistemaHandlerMethodNotAllowed)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1`.

- Contrato administrativo: ahora es versionado, editable desde super y exigido por version en el login.
	- Archivos creados: `backend/db/contrato_super.go`, `backend/handlers/super_contrato_handlers.go`, `backend/handlers/super_contrato_handlers_test.go`, `web/super/contrato.html`.
	- Archivos modificados: `backend/main.go`, `backend/handlers/accept_handlers.go`, `backend/handlers/auth_admin_handlers.go`, `backend/handlers/e2e_login_acceptance_test.go`, `backend/handlers/auth_users_carritos_test.go`, `backend/utils/utils.go`, `web/accept.html`, `web/contrato.html`, `web/super_administrador.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/estructura_bd.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el contrato que aceptan los administradores deja de ser un HTML estatico y pasa a vivir en la base `superadministrador`, con historial por version y resumen de cambio. Super puede editarlo desde una pagina dedicada, el portal lo publica via `/api/public/contrato` y el login administrativo exige aceptar la ultima version antes de crear sesion.
	- Verificación: `go test ./handlers -run "Test(PublicContratoHandlerReturnsDefaultVersion|SuperContratoHandlerCreatesNewVersionAndHistory|E2E_AcceptContractCreatesSession|E2E_AcceptContractRequiresNewVersion|AuthMiddlewareAllowsPublicPortalPagesAssetsAndHomeCardsAPI)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1`; diagnostico del editor sin errores en los archivos Go/HTML tocados.

- Deploy VPS: sync_to_vps ahora abre el dominio publico canonico en lugar de la IP.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: `sync_to_vps.ps1` agrega `PublicBaseUrl` con valor por defecto `https://powerfulcontrolsystem.com/` y usa esa URL al finalizar el deploy, manteniendo `RemoteHost` solo para SSH y evitando abrir `http://<ip>:<puerto>/` en el navegador.
	- Verificación: validacion de sintaxis PowerShell mediante parser (`[System.Management.Automation.Language.Parser]::ParseFile(...)`) sin errores.

- Checkout de licencias: Epayco disponible con Public Key y rutas publicas de pago realmente abiertas.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `backend/handlers/system_empresas_handlers_test.go`, `backend/utils/utils.go`, `backend/utils/utils_test.go`, `web/pagar_licencia.html`, `web/super/configuracion_avanzada.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el backend deja de exigir `epayco.private_key` para mostrar Epayco en el checkout actual y usa `epayco.public_key` como requisito minimo operativo junto al flag `enabled`. Tambien se corrige `AuthMiddleware` para dejar publicas `/api/public/licencias/payment_methods`, `/wompi/*` y `/epayco/*`, y `web/pagar_licencia.html` ahora indica si la pasarela esta desactivada o si falta la `Public Key`.
	- Verificación: `go test ./handlers -run "Test(EpaycoCreateTransactionHandlerUsesConfiguredPublicBaseURLAndKeys|EpaycoCreateTransactionHandlerAllowsCheckoutWithoutPrivateKey|PublicLicenciasPaymentMethodsHandlerOrdersAndAvailability|PublicLicenciasPaymentMethodsHandlerAllowsEpaycoWithPublicKeyOnly)" -count=1`; `go test ./utils -run "Test(AuthMiddlewareAllowsPublicLicenciaPaymentRoutesWithoutSession|JSONErrorMiddlewareWrapsEpaycoNonJSONError)" -count=1`; diagnostico del editor sin errores en los archivos Go/HTML tocados.

- Login admin y configuraci?n Gmail: se simplifica el hint visual y se habilita edici?n directa.
	- Archivos modificados: `web/login.html`, `web/super/configuracion_avanzada.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el login administrativo ya no muestra el bloque `Se recordar? ... / Olvidar`, aunque conserva la logica de `Recordar cuenta`, y la seccion Gmail SMTP del panel super deja de bloquear el correo remitente y los demas campos cuando ya existe una configuracion guardada.
	- Verificación: `go test ./handlers -run "TestGmailConfigHandlerSaveRestartAlertTo" -count=1`; diagnostico del editor sin errores en los archivos HTML tocados.

- Portal publico: pagina_principal ahora define tamanos de tarjetas y texto para home y landing.
	- Archivos modificados: `backend/handlers/pagina_principal_handlers.go`, `backend/handlers/pagina_principal_handlers_test.go`, `web/super/pagina_principal.html`, `web/index.html`, `web/descripcion_de_los_sistemas.ht`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el editor super de pagina_principal agrega ajustes globales de tamano para tarjetas y texto del `index.html` y de `/descripcion_de_los_sistemas.ht`. La API publica mantiene un contrato unico (`tarjetas` + `estilos`) y el frontend aplica esos valores de forma responsive.
	- Verificación: `go test ./handlers -run "TestPaginaPrincipal|TestPublicPaginaPrincipalHandlerExposesLandingFields" -count=1`; diagnostico del editor sin errores en los archivos Go/HTML/CSS tocados.

- Portal publico: CTA de WhatsApp arriba a la derecha y botones del header con estilo mini-tarjeta.
	- Archivos modificados: `web/index.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el home comercial reposiciona el CTA flotante `Contactenos` a la esquina superior derecha y convierte `Registrarse o iniciar sesión` e `Informacion de contacto` en accesos compactos con acabado visual de mini-tarjeta, reutilizando el lenguaje de las tarjetas del portal sin alterar rutas publicas ni comportamiento funcional.
	- Verificación: diagnostico del editor sin errores en `web/index.html` y `web/estilos.css`.

- Portal publico: la landing descriptiva ahora se configura desde pagina_principal.
	- Archivos creados: `backend/handlers/pagina_principal_handlers_test.go`.
	- Archivos modificados: `backend/handlers/pagina_principal_handlers.go`, `web/super/pagina_principal.html`, `web/descripcion_de_los_sistemas.ht`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la configuracion super de pagina_principal deja de limitarse al home y ahora tambien guarda la etiqueta, titular ampliado, parrafos y capacidades clave de cada tarjeta para `/descripcion_de_los_sistemas.ht`. La landing descriptiva deja de depender de textos estaticos por nombre de sistema y renderiza el contenido extendido desde la misma API publica usada por `index.html`.
	- Verificación: `go test ./handlers -run "Test(PaginaPrincipal|AuthMiddlewareAllowsPublicPortalPagesAssetsAndHomeCardsAPI)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1`; diagnostico del editor sin errores en `backend/handlers/pagina_principal_handlers.go`, `backend/handlers/pagina_principal_handlers_test.go`, `web/super/pagina_principal.html` y `web/descripcion_de_los_sistemas.ht`.

- Checkout de licencias: retorno recuperable tras volver de Epayco y Wompi.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `web/pagar_licencia.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el checkout de licencias ya no depende de un `status` estatico en la URL al volver desde la pasarela. Backend y frontend ahora conservan `provider`, `reference`, `transaction_id`, `licencia_id` y `empresa_id`, reanudan la verificacion real del pago desde `web/pagar_licencia.html` y permiten que Wompi consulte estado por `reference` cuando el navegador regresa sin `transaction_id` directo.
	- Verificación: `go test ./handlers -run "Test(EpaycoCreateTransactionHandlerUsesConfiguredPublicBaseURLAndKeys|WompiTransactionStatusHandlerAllowsReferenceLookup|ResolvePaymentBaseURLRejectsLocalhostWithoutPublicConfig|ResolvePaymentBaseURLUsesConfiguredCanonicalDomain|PublicLicenciasPaymentMethodsHandlerOrdersAndAvailability)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1`; diagnostico del editor sin errores en `web/pagar_licencia.html`, `backend/handlers/payments_handlers.go` y `backend/handlers/payments_handlers_test.go`.

- Checkout de licencias: fix Epayco con `public_key` real y callbacks sobre dominio público.
	- Archivos creados: `backend/handlers/payments_handlers_test.go`.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/system_empresas_handlers_test.go`, `web/super/configuracion_avanzada.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se corrige el flujo de Epayco para separar `public_key`, `private_key` y `customer_id`, mantener compatibilidad con configuraciones legacy y resolver `response`/`confirmation` desde una base pública v?lida en vez de `localhost`. La pantalla de configuraci?n avanzada deja de confundir la llave pública con el identificador del comercio y Wompi reutiliza la misma base pública para su `redirect_url`.
	- Verificación: `go test ./handlers -run "TestResolvePaymentBaseURL|TestEpaycoCreateTransactionHandlerUsesConfiguredPublicBaseURLAndKeys|TestPublicLicenciasPaymentMethodsHandlerOrdersAndAvailability" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1`; diagnostico del editor sin errores en los archivos tocados.

- Login Google: host can?nico en dominio ra?z y estaciones con carga visible.
	- Archivos modificados: `backend/utils/utils.go`, `backend/utils/utils_test.go`, `backend/main.go`, `backend/.env.example`, `scripts/sync_to_vps.ps1`, `web/administrar_empresa/estaciones.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se corrige la inestabilidad del acceso administrativo tras registrar el dominio público, redirigiendo `www.powerfulcontrolsystem.com` al host can?nico `powerfulcontrolsystem.com` antes de procesar OAuth y alineando los defaults de `GOOGLE_REDIRECT_URL` al callback del dominio ra?z. Además, la página de estaciones ahora muestra `Cargando estaciones...` mientras consulta configuraci?n, carritos y sensores, con mensaje visible en caso de error.
	- Verificación: `go test ./utils -run "Test(CanonicalPublicHostMiddleware|LoggingMiddlewareSetsContextAndWritesLogs|JSONErrorMiddlewareWrapsNonJSONError)" -count=1`; `go test ./handlers -run "TestHandleGoogleLogin" -count=1`; diagn?stico del editor sin errores nuevos en los archivos tocados.

- Portal publico: home, landing descriptiva y contacto liberados sin sesion.
	- Archivos modificados: `backend/utils/utils.go`, `backend/handlers/auth_users_carritos_test.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: `AuthMiddleware` incorpora `/descripcion_de_los_sistemas.ht` y `/Informacion_de_contacto.html` al whitelist publico y mantiene `index.html` dentro del mismo conjunto, para que las tres paginas comerciales del portal sean accesibles sin login. La prueba de middleware se amplia para cubrir esas rutas junto con `menu.js` y `/api/public/pagina_principal`.
	- Verificación: `go test ./handlers -run "TestAuthMiddlewareAllowsPublicPortalPagesAssetsAndHomeCardsAPI" -count=1`; diagnostico del editor sin errores en Go y documentos modificados.

- Portal publico: contacto visible por WhatsApp y pagina dedicada de informacion.
	- Archivos creados: `web/Informacion_de_contacto.html`.
	- Archivos modificados: `web/index.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se agrega un CTA flotante `Contactenos` en `index.html` que abre WhatsApp con el numero comercial del sistema, un acceso visible a `Informacion_de_contacto.html` desde el encabezado del portal y una nueva pagina publica con descripcion general del sistema, correo `powerfulcontrolsystem@hmail.com` y WhatsApp `3043306506`. Ademas, el acceso principal del header pasa a llamarse `Registrarse o iniciar sesión` y queda junto al boton de contacto.
	- Verificación: diagnostico del editor sin errores en `web/index.html`, `web/Informacion_de_contacto.html` y `web/estilos.css`.

- Portal publico: landing descriptiva unica para todas las tarjetas del home.
	- Archivos creados: `web/descripcion_de_los_sistemas.ht`.
	- Archivos modificados: `web/index.html`, `web/super/pagina_principal.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el boton `Explorar oferta` del home deja de abrir enlaces directos y pasa a una sola landing publica (`/descripcion_de_los_sistemas.ht`) con anclas por tarjeta, descripciones ampliadas por seccion y un CTA `Probar Gratis` por cada solucion. El enlace configurado desde `super/pagina_principal.html` ahora alimenta ese CTA final.
	- Verificación: diagnostico del editor sin errores en los archivos HTML/CSS modificados.

- Checkout de licencias: Epayco primero, Wompi debajo y activacion real desde configuracion avanzada.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/system_empresas_handlers_test.go`, `backend/main.go`, `web/pagar_licencia.html`, `web/super/configuracion_avanzada.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se agrega la ruta publica `GET /api/public/licencias/payment_methods` para publicar la disponibilidad ordenada de pasarelas de licencia, `web/pagar_licencia.html` ya muestra solo Epayco y Wompi con prioridad Epayco -> Wompi, y `web/super/configuracion_avanzada.html` permite activar/desactivar ambas pasarelas manteniendo a Wompi bloqueado en backend cuando esta desactivado o incompleto.
	- Verificación: `go test ./handlers -run "TestPublicLicenciasPaymentMethodsHandlerOrdersAndAvailability|TestWompiConfigHandlerPersistsEnabledFlag|TestWompiTermsHandlerRejectsWhenDisabled" -count=1`; `go test ./ -run "^$" -count=1`; diagnostico del editor sin errores en HTML/CSS/Go tocados.

- Sync VPS: reparacion del redeploy remoto en fallback PuTTY.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el wrapper PowerShell deja de pasar inline a `plink` los bloques remotos complejos de `bootstrap` y `redeploy`; ahora los escribe en archivos temporales UTF-8 sin BOM y los ejecuta con `plink -m`, estabilizando el `heredoc` de la unidad `systemd` y evitando fallos Bash como `syntax error near unexpected token '('`. Tambien se endurece el quoting del binario remoto y de los directorios de logs.
	- Verificación: parser PowerShell en verde para `scripts/sync_to_vps.ps1` y diagnostico del editor sin errores nuevos en el archivo.

- Login Google: hardening de `login_hint` y saneamiento de cuenta recordada en escritorio.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/handlers/auth_users_carritos_test.go`, `web/js/login.js`, `web/menu.js`, `web/js/super_administrador.js`, `web/js/seleccionar_empresa.js`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se evita que el login Google herede un `login_hint` corrupto desde el navegador. El backend solo reenvia hints con formato de correo valido y el frontend limpia/persiste `rememberedEmail` unicamente cuando el dato es plausible, estabilizando el flujo especialmente en escritorio.
	- Verificación: `go test ./handlers -run "TestHandleGoogleLogin" -count=1` y `go test ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1`.

- Frontend web: refuerzo responsive transversal para portal y paneles administrativos.
	- Archivos modificados: `web/index.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se mejora la adaptacion entre escritorio, tablet y movil en la portada publica y en los layouts compartidos. El hero de `index.html` permite salto natural del titulo/subtitulo, el sidebar administrativo colapsa con mejor navegacion horizontal en movil y formularios/tablas/botones se reorganizan para pantallas estrechas.
	- Verificación: diagnostico del editor sin errores en `web/index.html` y `web/estilos.css`.

- VPS web: restauracion del dominio publico sin puerto con Nginx, UFW y TLS correctos.
	- Archivos modificados: `documentos/deploy_nginx_reverse_proxy_vps.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se corrige la incidencia de publicacion donde `powerfulcontrolsystem.com` dejaba de cargar externamente aunque el backend y Nginx estaban activos en el VPS; la causa fue `443/tcp` ausente en UFW y cobertura incompleta de `www` en TLS. Se abre `443/tcp`, se renueva el certificado LetsEncrypt para `powerfulcontrolsystem.com` y `www.powerfulcontrolsystem.com`, y se documenta la configuracion minima correcta para Nginx/Certbot.
	- Verificación: `curl -I https://powerfulcontrolsystem.com/` y `curl -I https://www.powerfulcontrolsystem.com/` responden `200 OK`; `certbot certificates` muestra ambos dominios; `ufw status` incluye `443/tcp ALLOW`.

- Sync VPS: limpieza automatica de procesos huerfanos antes del restart remoto.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `scripts/sync_to_vps.sh`, `scripts/README_sync.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se endurece el redeploy remoto para detener listeners fuera de `systemd` que sigan ocupando `SERVER_PORT`, registrar el PID/comando conflictivo y abortar con diagnostico si el puerto no se libera; adicionalmente se sane? el VPS donde un binario `server_linux_amd64 (deleted)` manten?a `:8080` ocupado y dejaba `powerfulcontrolsystem.service` en bucle de reinicio.
	- Verificación: en VPS `powerfulcontrolsystem.service` qued? `active (running)` tras limpiar el listener hu?rfano; `curl -k -I https://powerfulcontrolsystem.com/auth/google/login` sigue devolviendo `302` con `redirect_uri=https://powerfulcontrolsystem.com/auth/google/callback`.

- Sync VPS: bootstrap endurecido, mensajes accionables y preparaci?n asistida del servidor.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `scripts/sync_to_vps.sh`, `scripts/README_sync.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se robustecen ambos scripts de despliegue para validar puertos/timeout antes de conectar, detectar el gestor de paquetes remoto e instalar dependencias base del VPS cuando hay privilegios, actualizar `SERVER_PORT` en cada bootstrap, exigir mensajes etiquetados `BOOTSTRAP_*`/`DEPLOY_*` y devolver hints claros cuando fallan DSN PostgreSQL, `CONFIG_ENC_KEY`, permisos `root/sudo` o el arranque del servicio `systemd`.
	- Verificación: parser de PowerShell en verde para `scripts/sync_to_vps.ps1`; diagnostico del editor sin errores para `scripts/sync_to_vps.sh`; previsualizacion `./scripts/sync_to_vps.ps1 -PreviewOnly -SkipBuild -OpenPublicUrlAfterDeploy:$false` generando correctamente las etapas remotas; la validacion directa `bash -n` sigue pendiente en este equipo porque `bash.exe` apunta al lanzador de WSL y no hay distribucion instalada.

- Login y menu: correccion de `recordar cuenta` y deteccion visible de sesion.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/handlers/accept_handlers.go`, `backend/handlers/usuarios_empresa.go`, `backend/handlers/e2e_login_acceptance_test.go`, `backend/handlers/auth_users_carritos_test.go`, `backend/main.go`, `web/js/login.js`, `web/menu.js`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se deja de depender de la lectura cliente de `session_token` para sincronizar `recordar cuenta`, avatar y enlace de cierre de sesion; el backend emite `browser_session_active` como se?al visible no sensible, manteniendo el token real en cookie `HttpOnly` y alineando tambien la limpieza de cookies en logout.
	- Verificación: `go test ./handlers -run "TestE2E_AcceptContractCreatesSession|TestEmpresaUsuario(LoginHandlerSuccess|LoginHandlerRejectsWrongEmpresaScope|LoginHandlerRejectsWrongEmpresaScopeFromQuery|SetPasswordHandlerSuccess)|TestSuperEndpointsPermisosPorRol" -count=1` y `go test ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1`.

## 2026-04-14
- Sync VPS: backend persistente con `systemd` y autoarranque tras reinicio del servidor.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `scripts/sync_to_vps.sh`, `scripts/README_sync.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/deploy_nginx_reverse_proxy_vps.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el despliegue al VPS deja de usar `nohup` para reiniciar el backend y pasa a instalar/actualizar una unidad `systemd` del proyecto con `Restart=always`, `systemctl enable`, carga de entorno desde `backend/.env.local` y logs persistentes en `backend/server.log` / `backend/server.err`, garantizando que el servicio vuelva solo tras caidas del proceso o reinicios del VPS y que solo se reinicie durante `sync_to_vps`.
	- Verificación: parser de PowerShell para `scripts/sync_to_vps.ps1`, previsualizacion local del script con `-PreviewOnly -SkipBuild` y diagnostico del editor sin errores para `scripts/sync_to_vps.sh`; la validacion directa con `bash -n` queda pendiente en este equipo porque no hay distro WSL ni Git Bash instalados.

## 2026-04-14
- Manual de instalacion: reposicion del documento y guia Google OAuth para VPS.
	- Archivos creados: `documentos/manual_de_instalacion.md`.
	- Archivos modificados: `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se repone el manual eliminado en `HEAD` y se actualiza con la configuracion exacta de Google Cloud Console para login local y produccion, incluyendo `Authorized redirect URIs` y `Authorized JavaScript origins` para `localhost` y `powerfulcontrolsystem.com`, mas notas de diagnostico para `redirect_uri_mismatch`.
	- Verificación: revision documental del manual recreado y comprobacion estatica de las URLs de callback/origen documentadas.

- Portal principal: título en una sola l?nea con subtítulo debajo en la misma columna.
	- Archivos modificados: `web/index.html`, `web/estilos.css`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se agrega el contenedor `portal-intro-copy` para apilar verticalmente el encabezado del home, manteniendo `Sistema de Facturaci?n Electr?nica` en una sola fila y moviendo `Toma el control de tu negocio con Powerful Control System` justo debajo, centrado en el mismo bloque visual.
	- Verificación: revision estatica de estructura HTML/CSS confirmando el nuevo contenedor y la regla `white-space: nowrap` aplicada al título.

- Login administrativo Google: correccion para VPS y local + recordar cuenta estable.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/utils/utils.go`, `backend/handlers/auth_users_carritos_test.go`, `web/login.html`, `web/menu.js`, `web/js/login.js`, `web/index.html`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se corrige el flujo OAuth para adaptar `redirect_uri` al host real de la solicitud y forzar `https` en dominio publico (`powerfulcontrolsystem.com`), se habilitan rutas publicas que bloqueaban el login (`/js/login.js` y `/api/public/pagina_principal`), se evita consulta a `/me` sin sesion para eliminar ruido `401` en F12 y se completa la experiencia de `recordar cuenta`; adicionalmente se actualiza el encabezado del home a `Sistema de Facturaci?n Electr?nica` con subtitulo operativo.
	- Verificación: `go test ./handlers -run "TestHandleGoogleLogin|TestAuthMiddlewareAllowsPublicLoginAssetsAndHomeCardsAPI" -v -count=1` en verde; en VPS `GET /js/login.js` y `GET /api/public/pagina_principal` responden `200`; `GET /auth/google/login` emite `redirect_uri=https://powerfulcontrolsystem.com/auth/google/callback`; `google.redirect_url` en BD super qued? en HTTPS.

- Inicio local: diagnostico robusto para tunel SSH de PostgreSQL en VPS.
	- Archivos modificados: `scripts/iniciar_servidor.ps1`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se mejora `Ensure-VpsPostgresTunnel` para esperar el listener con reintentos (hasta ~8s), capturar `stdout/stderr` de `plink` en `backend/tmp/plink_tunnel_<puerto>.*.log` y reportar causa detallada cuando el tunel no abre el puerto local; adicionalmente se corrige el argumento `-i` de `plink` para rutas de llave SSH con espacios (comillas explicitas), evitando el fallo `Host does not exist`.
	- Verificación: validacion de parseo PowerShell en verde con `[System.Management.Automation.Language.Parser]::ParseFile("scripts/iniciar_servidor.ps1", ...)` y ejecucion real `. "D:\powerfulcontrolsystem\scripts\iniciar_servidor.ps1" -Background` completando arranque con tunel activo y backend en `:8080`.

- Checkout de licencias: cierre operativo de Epayco.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/main.go`, `web/pagar_licencia.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/estructura_bd.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se completa la implementacion de Epayco para licencias con `POST /epayco/create_transaction`, `GET /epayco/transaction_status` y `POST/GET /epayco/webhook`; se corrige la configuracion super de Epayco para aceptar credenciales reales sin validacion numerica de `cust_id`; y el frontend abre `checkout_url` de Epayco en una nueva pesta?a manteniendo polling de estado y activacion automatica de licencia al aprobar.
	- Verificación: `go test ./ -run "^$" -count=1`, `go test ./handlers -run "^$" -count=1`, `go test ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1` en verde.

- Chat y tareas: nuevo agente de citas con calendario grande y recordatorios previos.
	- Archivos modificados: `backend/db/chat_tareas.go`, `backend/handlers/chat_tareas.go`, `backend/handlers/chat_tareas_test.go`, `backend/main.go`, `web/administrar_empresa/chat_y_tareas.html`, `web/estilos.css`, `web/super/pagina_principal.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/estructura_bd.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se agrega agenda de citas empresarial en el modulo de chat/tareas (`/api/empresa/chat_tareas/citas`) con calendario mensual de gran formato, programacion/edicion de reuniones, visibilidad compartida por `empresa_id` y banner de recordatorios previos; adicionalmente se incluye un boton inferior de guardado en `web/super/pagina_principal.html`.
	- Verificación: `$env:DB_DIALECT='motor_legado_retirado'; go test ./handlers -run ChatTareas -count=1` y `$env:DB_DIALECT='motor_legado_retirado'; go test ./db -run ChatTareas -count=1`.

- UI administrativa: eliminacion de barra superior de titulo/acciones en todas las paginas de layout.
	- Archivos modificados: `web/super_administrador.html`, `web/administrar_empresa.html`, `web/administrar_empresa/finanzas_menu.html`, `web/administrar_empresa/facturacion_electronica_menu.html`, `web/administrar_empresa/administrar_productos_menu.html`, `web/administrar_empresa/configuracion_menu.html`, `web/administrar_empresa/reportes_menu.html`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se retira por completo el bloque visual `admin-toolbar page-header` del panel super y de los menus administrativos para eliminar la barra superior de la derecha en todas las vistas del layout.
	- Verificación: busqueda `class="admin-toolbar"` en `web/**/*.html` sin resultados.

- Inicio local: correccion de deteccion de procesos en puerto 8080 bajo StrictMode.
	- Archivos modificados: `scripts/iniciar_servidor.ps1`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se normaliza la coleccion de PIDs detectados en el paso de liberacion de puerto para evitar `No se encuentra la propiedad 'Count'` cuando solo existe un proceso escuchando.
	- Verificación: ejecucion real `. 'D:\powerfulcontrolsystem\scripts\iniciar_servidor.ps1'` completando `3/8 Liberando puerto 8080` sin excepcion y arranque exitoso del backend en `:8080`.

- Inicio local: correccion de carga DSN PostgreSQL y tunel DB opcional en script de arranque.
	- Archivos modificados: `scripts/iniciar_servidor.ps1`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Archivo local de entorno actualizado (no versionado): `backend/.env.local`.
	- Descripción: el script ahora carga `DB_DIALECT`, `DB_EMPRESAS_DSN` y `DB_SUPERADMIN_DSN` desde `.env.local/.env` antes de validar prerequisitos; se anade soporte opcional para tunel SSH a PostgreSQL en VPS (`DB_VPS_TUNNEL_*`) con validacion temprana del puerto de tunel y ajuste temporal de DSN al listener local.
	- Verificación: `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\iniciar_servidor.ps1 -Background` en verde y `curl -I http://127.0.0.1:8080` con `HTTP/1.1 200 OK`.

- Venta publica por subdominio empresarial automatizado.
	- Archivos modificados: `backend/main.go`, `backend/handlers/venta_publica.go`, `backend/handlers/venta_publica_test.go`, `web/venta_publica.html`, `web/administrar_empresa/venta_publica.html`, `documentos/deploy_nginx_reverse_proxy_vps.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se habilita resolución de `empresa_slug` por subdominio (`{slug}.powerfulcontrolsystem.com`) en backend y frontend de venta publica, con soporte de apertura automatica de tienda desde la raiz del subdominio.
	- Verificación: `go test ./handlers -run "VentaPublica|ResolveVentaPublicaSlugFromHost" -count=1` en verde.
	- Evidencia VPS: Nginx actualizado con bloque wildcard y captura de slug por host; validado `GET /` en host de subdominio con `302` a `/venta_publica.htmlempresa_slug=<slug>` y `GET /venta_publica.htmlempresa_slug=<slug>` con `200 OK`; queda pendiente registrar wildcard DNS `*.powerfulcontrolsystem.com` (resolucion publica actual `NXDOMAIN`).

## 2026-04-14
- Guia operativa de dominio con Nginx reverse proxy en VPS.
	- Archivo creado: `documentos/deploy_nginx_reverse_proxy_vps.md`.
	- Archivos modificados: `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `documentos/descripcion_del_proyecto`, `CHANGELOG.md`.
	- Descripción: se documenta el procedimiento para publicar `powerfulcontrolsystem.com` y `www.powerfulcontrolsystem.com` con Nginx en Ubuntu VPS, manteniendo el backend en `127.0.0.1:8080`, con validaciones de servicio/UFW y opcion de HTTPS con Certbot.
	- Verificación: guia con comandos en orden, listos para copia/pegado en consola remota.

## 2026-04-14
- Modulo de impresoras operativas por empresa.
	- Archivos creados: `backend/db/empresa_impresoras.go`, `backend/db/empresa_impresoras_test.go`, `backend/handlers/empresa_impresoras.go`.
	- Archivos modificados: `backend/main.go`, `web/administrar_empresa/configuracion.html`, `web/administrar_empresa/carrito_de_compras.html`, `web/administrar_empresa/finanzas.html`, `web/administrar_empresa/reportes.html`, `documentos/estructura_bd.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se a?ade gestion de impresoras por `empresa_id` (predeterminada, estado activo/inactivo, asignacion por funcionalidad y por producto) y resolución de carrito/finanzas/reportes.
	- Verificación: `go test ./ ./auth ./db ./handlers ./metrics ./utils -count=1` en verde.

- Super administrador: nuevo panel de administracion de base de datos PostgreSQL.
	- Archivos creados: `backend/handlers/postgres_performance.go`, `backend/handlers/postgres_performance_test.go`, `web/super/administrar_base_de_datos.html`.
	- Archivos modificados: `backend/main.go`, `web/super_administrador.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se agrega un tablero profesional para monitoreo de PostgreSQL (salud del cluster, metricas por base, consultas activas prolongadas, `pg_stat_bgwriter` y recomendaciones automaticas), con endpoint protegido `/super/api/postgres/performance`.
	- Verificación: `go test ./handlers -run "PostgresPerformance" -count=1` y `go test ./ ./auth ./db ./handlers ./metrics ./utils -count=1` en verde.

- Migracion cerrada a PostgreSQL-only y retiro de motor legado retirado operativo.
	- Archivos modificados: `backend/main.go`, `backend/db/sql_compat.go`, `scripts/iniciar_servidor.ps1`, `scripts/sync_to_vps.ps1`, `scripts/sync_to_vps.sh`, `scripts/README_sync.md`, `scripts/actualizar_repositorio.ps1`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/estructura_bd.md`, `documentos/descripcion_de_las_bases_De_datos`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Archivos eliminados: `backend/db/pcs_empresas`, `backend/db/pcs_superadministrador`.
	- Descripción: el backend queda forzado a runtime PostgreSQL-only, sin fallback motor legado retirado en arranque; se limpian los `.db` legados del repositorio y se alinea la operacion local/remota a DSN PostgreSQL obligatorios.

- Estandarizacion documental ERP multiempresa.
	- Archivos creados: `documentos/erp_multiempresa/README.md`, `documentos/erp_multiempresa/01_alcance_erp_multiempresa.md`, `documentos/erp_multiempresa/02_diseno_tecnico_erp_multiempresa.md`, `documentos/erp_multiempresa/03_especificaciones_funcionales_erp_multiempresa.md`, `documentos/erp_multiempresa/04_guia_implementacion_erp_multiempresa.md`.
	- Archivos modificados: `documentos/README.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: se consolida un paquete ERP estandar listo para revision, con claridad de alcance, arquitectura, requisitos funcionales, reglas de negocio, integraciones y ruta de implementacion por fases.

- Documentacion: reorganizacion profesional, consolidacion de fuentes canonicas y limpieza de artefactos no usados.
	- Archivos modificados: `documentos/README.md`, `documentos/descripcion_del_proyecto`, `documentos/estructura_del_codigo`, `.gitignore`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`.
	- Archivos depurados: `documentos/historial_de_cambios_addendum_2026-04-04.md`, `backend/tmp/server.exe`, `backend/server.err`, `backend/server.log`, `backend/server.run.log`, `backend/logs/*.log`, `logs/test_runs/*.log`, `scripts/logs/*.log`, `tmp/doc_audit_report.txt`, `tmp/doc_hash_duplicates.txt`.
	- Descripción: se centraliza la documentacion en un indice canonico, se evita duplicidad entre documentos estructurales y se eliminan archivos temporales/runtime que no deben versionarse.
	- Verificación: carpetas de logs temporales quedan limpias y se mantiene solo estado runtime necesario (`backend/logs/server_runtime_state.json`).

- OAuth Google VPS: validacion final de infraestructura HTTPS y diagnostico concluyente de `redirect_uri_mismatch`.
	- Archivos modificados: `CHANGELOG.md`, `documentos/historial_de_cambios`.
	- Descripción: se verifica en VPS que el backend emite callback seguro `https://2.24.197.58.nip.io/auth/google/callback` y que el proxy TLS (Caddy) esta operativo en `:443`; Google sigue rechazando el flujo por URI no autorizada en el cliente OAuth.
	- Verificación: prueba E2E desde VPS confirma mismatch para la URI HTTPS publica y matriz de prueba muestra aceptacion solo de `http://localhost:8080/auth/google/callback`.
	- Pendiente externo: agregar la URI exacta del VPS en Google Cloud Console y repetir prueba de login.

- Inicio local: correcci?n de detecci?n de puerto 8080 para evitar falso bloqueo por PID 0.
	- Archivos modificados: `scripts/iniciar_servidor.ps1`.
	- Descripción: se reemplaza la detecci?n basada en `netstat | findstr ":8080"` por una resolución de listeners locales reales (primero `Get-NetTCPConnection`, con fallback parseado de `netstat` en estado `LISTENING`). Se filtran PID inv?lidos/no gestionables (`<= 0`) y se evita abortar cuando aparece `System Idle Process` sin listener real del backend.
	- Verificación: ejecución local `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\iniciar_servidor.ps1 -Background` completa en verde; paso 3 muestra `No hay procesos escuchando en el puerto 8080` y el servidor inicia correctamente.

- OAuth Google VPS: prioridad de entorno sobre DB + soporte de `GOOGLE_REDIRECT_URL` en despliegue.
	- Archivos modificados: `backend/main.go`, `scripts/sync_to_vps.ps1`, `scripts/sync_to_vps.sh`.
	- Descripción: se ajusta la carga OAuth para que `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET` y `GOOGLE_REDIRECT_URL` del entorno tengan prioridad sobre valores almacenados en tabla `configuraciones` (la DB solo completa faltantes). Se a?ade adem?s propagaci?n de `GOOGLE_REDIRECT_URL` en bootstrap remoto de scripts de sincronización.
	- Verificación: `go test ./handlers -run "TestHandleGoogleLoginRedirect" -count=1` y `go test ./ -count=1` en verde. Diagn?stico en VPS confirma que el bloqueo actual es de pol?tica OAuth en Google (`secure-response-handling` / `redirect_uri_mismatch`) y no de base de datos.

- OAuth Google: correcci?n de callback para evitar `localhost` en entorno VPS.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/handlers/auth_users_carritos_test.go`, `backend/main.go`.
	- Descripción: se implementa resolución din?mica del `redirect_uri` por host/protocolo de la solicitud y una regla de reescritura segura cuando la configuraci?n existente apunta a loopback (`localhost/127.0.0.1`) pero el acceso real es público (VPS). El callback reutiliza la URL efectiva mediante cookie t?cnica de corta duraci?n para mantener consistencia en intercambio de token.
	- Verificación: despliegue real a VPS con `DEPLOY_OK:pid=53618 port=8080`; validación HTTP de `/auth/google/login` devuelve `redirect_uri=http://2.24.197.58:8080/auth/google/callback`.

- Sync VPS: guard estricto de DSN para PostgreSQL y recuperaci?n de despliegue estable.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `scripts/sync_to_vps.sh`, `scripts/README_sync.md`, `documentos/diagramas/estructura_del_codigo.md`, `CHANGELOG.md`, `documentos/historial_de_cambios`, `documentos/descripcion_de_archivos`.
	- Descripción: el bootstrap remoto ahora conserva valores DB existentes, valida el modo efectivo y bloquea el despliegue con `BOOTSTRAP_ERROR:POSTGRES_MISSING_DSN` cuando `postgres` no tiene ambos DSN; adem?s usa el ?ltimo valor de cada clave (`tail -n1`) y evita llegar a `DEPLOY_ERROR:process_not_running` por arranque inv?lido. En paralelo se restableci? configuraci?n DSN operativa en VPS para retomar despliegues en modo PostgreSQL.
	- Verificación: ejecución real `./scripts/sync_to_vps.ps1 -SkipBuild -RetryCount 1` primero falla en bootstrap con mensaje explícito de DSN faltantes, luego (tras restablecer DSN en VPS) finaliza con `DEPLOY_OK:pid=... port=8080` y `GET /` = `200`.

- VPS web root: correcci?n de resolución de est?ticos para index/login.
	- Archivos modificados: `backend/main.go`, `scripts/sync_to_vps.ps1`, `CHANGELOG.md`, `documentos/historial_de_cambios`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`.
	- Descripción: se ajusta `resolveWebDir()` para priorizar correctamente `.../web` cuando el binario corre desde `backend/bin`, evitando que el servidor sirva `backend/web/uploads/` como ra?z. Se redepliega en VPS y se valida apertura autom?tica de la URL pública.
	- Verificación: `GET /` = `200` con HTML de portal, `GET /index.html` = `200`, `GET /login.html` = `200`, proceso remoto activo en `:8080` y runtime PostgreSQL operativo.

- Sync VPS: hardening para preservar DSN remotos y apertura autom?tica de URL pública.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `scripts/sync_to_vps.sh`, `scripts/README_sync.md`.
	- Descripción: se excluye `backend/.env.local` de la sincronización para evitar sobrescribir secretos/DSN del VPS, se robustece el healthcheck de redeploy (detecta proceso ca?do y valida respuesta HTTP distinta de `000`) y se a?ade apertura autom?tica de `http://<host>:<puerto>/` al finalizar despliegues exitosos.
	- Verificación: ejecución real `./scripts/sync_to_vps.ps1 -RemoteHost 2.24.197.58 -RemoteUser root -RemotePath /root/powerfulcontrolsystem -DbDialect postgres -DbEmpresasDsn ... -DbSuperadminDsn ...` con `DEPLOY_OK:pid=... port=8080`, `GET / => 200` y backend en modo PostgreSQL con DSN activos.

- Migraci?n PostgreSQL (fase 4): estabilizaci?n de salida operativa en contabilidad y runtime VPS.
	- Archivos modificados: `backend/db/eventos_contables.go`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `Pendiente Notas`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`.
	- Descripción: se corrige el flujo del worker de asientos/eventos para PostgreSQL usando wrappers SQL portables y retorno de `id` compatible, eliminando el error `syntax error at or near "ORDER"` en runtime. Se restaura adem?s el entorno VPS con DSN PostgreSQL v?lidos en `backend/.env.local` y se valida arranque estable.
	- Verificación: `go test ./ ./auth ./db ./handlers ./metrics ./utils` en verde; validación remota en VPS con proceso activo, sin errores recientes de `asientos_worker` y healthcheck `HTTP=200`.

- Migraci?n PostgreSQL (fase 3): cierre documental del plan y sincronización de gobernanza por módulos.
	- Archivos modificados: `Pendiente Notas`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`.
	- Descripción: se marca Fase 3 como completada en el plan operativo, se agrega evidencia t?cnica de conmutaci?n a PostgreSQL y se alinea la documentaci?n de módulos/permisos sin cambios de privilegios en la matriz CRUD/A.
	- Verificación: se mantiene evidencia de pruebas del bloque core en verde (`go test ./ ./auth ./db ./handlers ./metrics ./utils -count=1`).

- Migraci?n PostgreSQL (fase 3): conmutaci?n de runtime backend a motor PostgreSQL en VPS.
	- Archivos modificados: `backend/main.go`, `backend/go.mod`, `backend/go.sum`, `scripts/sync_to_vps.ps1`, `scripts/sync_to_vps.sh`, `scripts/README_sync.md`, `documentos/diagramas/estructura_del_codigo.md`.
	- Descripción: el backend ahora selecciona motor por entorno (`DB_DIALECT`), abre conexiones con `pgx` usando `DB_EMPRESAS_DSN` y `DB_SUPERADMIN_DSN`, y omite el bootstrap motor legado retirado cuando el runtime es PostgreSQL. Los scripts de sincronización ahora propagan y verifican estas variables en `backend/.env.local` del VPS durante bootstrap remoto.
	- Verificación: `go test ./ ./auth ./db ./handlers ./metrics ./utils -count=1` en verde.

- Migraci?n PostgreSQL (fase 3): compatibilidad ampliada en n?cleo `backend/db`.
	- Archivos modificados: `backend/db/sql_compat.go`, `backend/db/empresa_scope.go`, `backend/db/productos.go`, `backend/db/db.go`.
	- Descripción: se ampl?a la capa de compatibilidad motor legado retirado/PostgreSQL con wrappers `query/exec` portables, inserciones con `RETURNING id` para PostgreSQL, detecci?n de tablas por `information_schema` y ajuste de `ensureColumnIfMissing` por dialecto con normalizaci?n de defaults de fecha. Además, se migra el bloque core de `db.go` (licencias, tipos de empresa, empresas, Wompi, asesores, configuraciones y m?tricas) para usar placeholders/fechas compatibles con ambos motores.
	- Verificación: `go test ./db -run "Session|Admin|User|Licencia|TipoEmpresa|Empresa|Config|Metric|Wompi|Asesor" -count=1` y `go test ./handlers -run "TestHandleGoogleLoginRedirectIncludesLoginHint|TestE2E_AcceptContractCreatesSession" -count=1` en verde.

- Sync VPS: selección autom?tica de clave de identidad al no pasar `-IdentityFile`.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `scripts/README_sync.md`.
	- Descripción: cuando no se especifica `-IdentityFile`, el script ahora prioriza la clave del proyecto `clave privada ssh.ppk` y, si no existe, usa `~/.ssh/id_rsa`. Además, mejora el mensaje de error cuando el VPS rechaza autenticaci?n.
	- Verificación: ejecución real `./scripts/sync_to_vps.ps1 -SkipBuild -RemoteHost 2.24.197.58 -RemoteUser root -RemotePath /root/powerfulcontrolsystem -RetryCount 1` completada con `Sincronizaci?n completada por fallback sin WSL (PuTTY)`.

- Sync VPS: redeploy remoto automático de backend tras sincronización.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `scripts/sync_to_vps.sh`, `scripts/README_sync.md`.
	- Descripción: la sincronización ahora detiene el proceso viejo del backend en VPS, inicia la nueva versi?n del binario y valida salud HTTP en el puerto configurado (`SERVER_PORT`), evitando que quede corriendo una versi?n antigua.
	- Verificación: ejecución real `./scripts/sync_to_vps.ps1 -SkipBuild -RemoteHost 2.24.197.58 -RemoteUser root -RemotePath /root/powerfulcontrolsystem -RetryCount 1` con salida `DEPLOY_OK:pid=... port=8080`.

- Migraci?n PostgreSQL (fase 3): avance inicial en autenticaci?n y sesiones.
	- Archivos a?adidos/modificados: `backend/db/sql_compat.go`, `backend/db/db.go`, `documentos/diagramas/estructura_del_codigo.md`.
	- Descripción: se incorpora capa de compatibilidad SQL motor legado retirado/PostgreSQL (rebindeo de placeholders y expresiones de fecha) y se aplica a funciones cr?ticas del flujo de autenticaci?n/sesiones (`UpsertUser`, `UpsertAdministrador`, `CreateSession`, `RevokeSessionByToken`, `GetSessionByToken`, `GetAdminByEmail`).
	- Verificación: `go test ./db -run "Session|Admin|User|Licencia" -count=1` y `go test ./handlers -run "TestHandleGoogleLoginRedirectIncludesLoginHint|TestE2E_AcceptContractCreatesSession" -count=1` en verde.

## 2026-04-13
 - Reparaci?n de login de usuario empresarial: permitir entrada manual de `empresa_id` y persistencia de contexto.
	- Archivos modificados:
		- web/login_usuario.html
		- web/js/login_usuario.js
	- Descripción: se agrega un campo `Empresa ID` en la página de login de usuario de empresa para aceptar el par?metro cuando no viene en la URL. La lógica JS persiste `empresa_id` en session/local storage, asegura que `redirect_url` incluya `empresa_id` y mejora la funcionalidad de "recordar usuario" por empresa.
	- Verificación: validación de sintaxis JS sin errores y flujo de login manual verificado localmente.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `scripts/README_sync.md`.
	- Descripción: en fallback sin WSL, el script ahora selecciona transporte por tipo de clave: `ssh.exe` + `scp.exe` para claves OpenSSH (ej. `id_rsa`) y `plink.exe` + `pscp.exe` para `.ppk`. Con esto se evita el error `Unable to use key file ... OpenSSH SSH-2 private key (new format)` al usar la identidad por defecto.
	- Verificación: `.\scripts\sync_to_vps.ps1 -SkipBuild -PreviewOnly -IdentityFile "$env:USERPROFILE\.ssh\id_rsa"` muestra `Fallback sin WSL (OpenSSH)` y comandos con `ssh.exe`/`scp.exe`.

- Migraci?n de datos a PostgreSQL en VPS: instalaci?n, ejecución por etapas y validación inicial.
	- Archivos modificados: `Pendiente Notas`, `documentos/regla_agente_go.md`, `copilot-instructions.md`, `documentos/descripcion_de_las_bases_De_datos`, `documentos/estructura_bd.md`, `estructura_bd.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`.
	- Descripción: se instala PostgreSQL en VPS por SSH, se crean las bases `pcs_superadministrador` y `pcs_empresas`, y se inicia la migraci?n desde motor legado retirado con `pgloader` en dos etapas (superadministrador y empresas), validando consistencia por conteo de tablas en cada base. Se formaliza adem?s la regla operativa: base productiva en VPS con PostgreSQL y motor legado retirado local como legado de migraci?n/contingencia.
	- Verificación: `VALIDACION_SUPER_OK` y `VALIDACION_EMPRESAS_OK` tras comparaci?n motor legado retirado vs PostgreSQL por tabla.

- Login administrativo: eliminación del mensaje visual de cuenta recordada y ajuste de OAuth.
	- Archivos modificados: `web/login.html`, `backend/handlers/auth_admin_handlers.go`, `backend/handlers/auth_users_carritos_test.go`.
	- Descripción: se elimina el texto en pantalla `Cuenta recordada ...` del login admin y se ajusta el par?metro OAuth `prompt` a `select_account` para evitar re-consentimiento de Google en cada inicio.
	- Verificación: `go test ./handlers -run TestHandleGoogleLoginRedirectIncludesLoginHint -count=1` en verde.

- Login administrativo: correcci?n de "Recordar cuenta" para evitar sesión parcial.
	- Archivos modificados: `web/js/login.js`, `web/menu.js`.
	- Descripción: se corrige el flujo para que cerrar sesión no elimine la preferencia cuando `rememberAccount=1`, se mantiene el correo recordado hasta que el usuario pulse "Olvidar" y se agrega sincronización de `rememberedEmail` desde `/me` cuando existe sesión activa.
	- Verificación: revisi?n de errores en frontend sin incidencias (`get_errors` en ambos archivos).

- Inicio local: hardening de scripts/iniciar_servidor para evitar ca?das del host de PowerShell/VS Code.
	- Archivo modificado: `scripts/iniciar_servidor.ps1`.
	- Descripción: se refuerza la liberaci?n de puerto 8080 para terminar ?nicamente procesos del backend (`server.exe`, `pos-backend`, `go run` del proyecto) y no procesos ajenos. Cuando el puerto est? ocupado por un proceso no gestionado, el script ahora informa el PID/nombre y aborta con mensaje claro en lugar de forzar `taskkill` indiscriminado. También se elimina el `Clear-Host` inicial para evitar efectos colaterales en consolas integradas.
	- Verificación: ejecución local `./scripts/iniciar_servidor.ps1 -Background` con `SCRIPT_EXIT=0` y comprobaci?n HTTP local `HTTP_STATUS=200`.

- Unificaci?n de bases motor legado retirado: solo dos archivos can?nicos del sistema.
	- Archivos modificados: `backend/main.go`, `documentos/estructura_bd.md`, `estructura_bd.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`.
	- Descripción: se normaliza la resolución de rutas en runtime para que el backend use por defecto `backend/db/pcs_empresas` y `backend/db/pcs_superadministrador` aunque se ejecute desde otro directorio. Se depuran copias operativas duplicadas en ra?z y en `backend/`, dejando ?nicamente dos archivos `.db` activos.
	- Verificación: inventario local posterior muestra exactamente dos DB (`backend/db/pcs_empresas` y `backend/db/pcs_superadministrador`) y pruebas backend en verde con `go test ./ ./auth ./db ./handlers ./metrics ./utils`.

- Sync VPS: bootstrap automático para servidor nuevo y diagn?stico de OAuth.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `scripts/README_sync.md`.
	- Descripción: se a?ade bootstrap post-sync en modo sin WSL para instalar dependencias base (`ca-certificates`, `curl`, `motor_legado_retirado`), asegurar `backend/.env.local` y reportar estado de variables cr?ticas (`GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `SERVER_PORT`, `CONFIG_ENC_KEY`) con salida `BOOTSTRAP_WARN/BOOTSTRAP_OK`. Se incorporan par?metros opcionales `-GoogleClientId` y `-GoogleClientSecret`.
	- Verificación: ejecución real con `SYNC_EXIT=0` y diagn?stico remoto mostrando faltantes OAuth (`GOOGLE_CLIENT_ID/SECRET` vac?os).

- Instalador de clave pública en Windows: correcci?n de errores de ejecución.
	- Archivo modificado: `scripts/instalar_clave_publica_vps.ps1`.
	- Descripción: se corrige el flujo para evitar errores remotos tipo `invalid option namepefail` y se adapta a PowerShell 5.1 eliminando sintaxis no soportada (``). Ahora usa comando remoto en una sola l?nea, validación de formato de clave OpenSSH y reintentos por timeout.
	- Verificación: `scripts/instalar_clave_publica_vps.ps1 -PreviewOnly -RemoteHost 2.24.197.58 -User root -Port 22` en verde (exit 0).

- Despliegue VPS: instalaci?n automatizada de clave pública PuTTYgen + robustecimiento de scripts de sincronización
	- Archivos a?adidos/modificados:
		- scripts/instalar_clave_publica_vps.ps1 (nuevo: instala clave pública RFC4716 en `~/.ssh/authorized_keys` de VPS Linux)
		- scripts/sync_to_vps.sh (hardening para Linux: validaciones, eliminación de `eval`, chequeo remoto de SO)
		- scripts/sync_to_vps.ps1 (manejo de errores sin cerrar terminal de VS Code, build Linux local previo y fallback PuTTY sin WSL con empaquetado tar)
		- scripts/README_sync.md (guía de ejecución en un comando)
		- web/login.html y web/js/login.js (completa UX de "Recordar cuenta" para login admin)
	- Descripción: se habilita un flujo operativo de un solo comando para preparar acceso por clave pública al VPS y se corrige la causa de cierres de terminal por `exit` en script PowerShell. `sync_to_vps.ps1` ahora compila en local un binario Linux (`backend/bin/server_linux_amd64`) antes de sincronizar y, sin Ubuntu/WSL, opera empaquetando el proyecto en `.tar`, subi?ndolo por `pscp.exe` y extray?ndolo en VPS por `plink.exe`, con trazas detalladas y exclusi?n de archivos sensibles/locales (`*.ppk`, `*.pem`, `*.key`, DB, logs, temporales); adem?s aplica `chmod +x` al binario remoto configurado. Se a?adi? manejo de `Connection timed out` con prechequeo TCP y reintentos automáticos configurables (`-RetryCount`) por etapa de conexi?n/subida/extracci?n.
	- Verificación: `scripts/instalar_clave_publica_vps.ps1 -PreviewOnly -RemoteHost 2.24.197.58` ejecuta correctamente; `scripts/sync_to_vps.ps1 -BuildOnly`, `-DryRun` y ejecución real con `-IdentityFile "D:\powerfulcontrolsystem\clave privada ssh.ppk"` completan con `exit code 0`; en VPS el artefacto qued? como ELF Linux en `/root/powerfulcontrolsystem/backend/bin/server_linux_amd64`.

- M?dulo Vendedores / Asesores comerciales: integraci?n de código de descuento y registro de asesor/vendedor en pagos
	- Archivos a?adidos/modificados:
		- backend/handlers/payments_handlers.go (extiende payload y persistencia de `pagos_wompi` con `discount_code` y `asesor_id`/`vendedor_id`)
		- backend/db/db.go (helpers para `asesores`, `asesor_comercial` y `asesor_comisiones`, y claves de configuraci?n `vendedor.*`)
		- backend/handlers/vendedores_handlers.go (nuevo: CRUD de asesores / vendedores)
		- backend/handlers/vendedor_config_handlers.go (nuevo: GET/PUT /super/api/vendedor_config)
		- backend/main.go (migraciones: tablas `asesores`, `asesor_comercial`, `asesor_comisiones`; registro de rutas `/super/api/vendedores`, `/super/api/asesor_comercial`, `/super/api/vendedor_config`)
		- backend/tools/insert_asesor.go, backend/tools/insert_plan.go, backend/tools/insert_licencia.go, backend/tools/create_session.go, backend/tools/query_pagos_comisiones.go (herramientas para pruebas locales)
		- web/pagar_licencia.html (nuevo campo `discount_code` y `asesor_id`/`vendedor_id` en el formulario de pago)
		- web/super/activar_asesor.html, web/super/asesor_comercial.html, web/super/vendedor_config_avanzado.html (UI super-administrador para activar vendedores, configurar planes y ajustes globales)
		- documentos/estructura_bd.md (documenta las nuevas tablas y columnas de pagos/comisiones)
		- documentos/descripcion_de_archivos (registro de los nuevos archivos del módulo)
	- Descripción: Se a?ade soporte opcional para incluir un código de descuento y una referencia al asesor/vendedor en el pago de licencias. Se introduce la entidad de `asesores` (vendedores), planes comerciales (`asesor_comercial`) y el registro de comisiones (`asesor_comisiones`) que crea una comisi?n inmediata y entradas programadas por meses de renovaci?n seg?n el plan.
	- Verificación: Prueba manual de activaci?n sin pago (`/licencias/activar_sin_pago`) usando sesión administrativa de prueba; se confirm? la creación de una fila en `pagos_wompi` con `discount_code` y `asesor_id` y la creación de registros en `asesor_comisiones` (comisi?n inmediata + programadas). Tests automatizados pendientes.

- Estaciones: fix de persistencia de `estaciones_config` cuando el frontend no envía `estado`.
	- Archivos modificados: `backend/db/empresa_estacion_prefs.go`, `backend/db/empresa_estacion_prefs_test.go`, `backend/handlers/empresa_estacion_prefs_test.go`.
	- Descripción: se normaliza `estado` vacio como `activo` en upsert/list/get de preferencias por estacion, evitando que las estaciones desaparezcan despues de guardarse.
	- Verificación: pruebas en verde de estaciones, sensores, ventas y facturacion documental.

- Estaciones: correccion de flujo 10+, colores movidos a configuracion de estaciones y hardening sensor/carrito.
	- Archivos modificados: `web/administrar_empresa/configuracion_de_estaciones.html`, `web/administrar_empresa/configuracion.html`, `web/administrar_empresa/estaciones.html`, `backend/handlers/empresa_estacion_prefs_test.go`.
	- Descripción: se consolida la gestion de colores de estado de carrito en la configuracion de estaciones, se fortalece el parseo de `estaciones_config` para tolerar payloads legacy anidados, se mejora la sincronizacion de carritos por estacion ante colisiones idempotentes y se valida el rango de estacion en configuracion de sensores.
	- Verificación: pruebas dirigidas en verde para handlers y DB en estaciones/sensores/carritos/facturación, incluyendo `TestEmpresaEstacionPrefsHandler_UpsertAndIsolationByEmpresa`.

- Reparaci?n integral de acceso empresarial y estaciones.
	- Archivos modificados: `web/login_usuario.html`, `web/js/login_usuario.js`, `web/js/seleccionar_empresa.js`, `web/administrar_empresa/configuracion_de_estaciones.html`.
	- Descripción: se corrige la continuidad del flujo `login usuario empresa -> seleccionar empresa -> administrar empresa` con persistencia de `empresa_id` y opci?n de recordar correo. La página de configuraci?n de estaciones se reconstruye y soporta generaci?n/sincronización masiva de estaciones (incluyendo 10+) con manejo tolerante de conflictos idempotentes al cerrar/inactivar carritos.
	- Verificación: pruebas backend de paquetes principales en verde (`go test ./ ./auth ./db ./handlers ./metrics ./utils`).

## 2026-04-12
- Flujo final de login administrativo: cuenta Google correcta + aceptaci?n ?nica de contrato + reCAPTCHA real.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/handlers/accept_handlers.go`, `backend/handlers/e2e_login_acceptance_test.go`, `backend/handlers/auth_users_carritos_test.go`, `web/login.html`, `web/js/login.js`, `web/accept.html`, `web/menu.js`, `web/estilos.css`.
	- Descripción: se unific? el flujo en `login.html -> OAuth -> /accept.html -> /accept/complete -> panel`, usando `administradores.acepta_contrato` como fuente canonica de aceptaci?n (sin depender de cookie global), validación server-side de reCAPTCHA y prompt OAuth `select_account consent` para evitar reutilizaci?n silenciosa de cuenta incorrecta.
	- Verificación: pruebas dirigidas en verde (`TestE2E_AcceptContractCreatesSession` y `TestHandleGoogleLoginRedirectIncludesLoginHint`).

- M?dulo sensor de puertas (Raspberry Pi): backend, handlers, UI y tests.
	- Archivos agregados/modificados:
		- backend/db/sensor_puertas.go (nuevo módulo DB: dispositivos y heartbeats)
		- backend/handlers/sensor_puertas.go (handlers: endpoint público `action=heartbeat` y configuraci?n protegida)
		- backend/db/sensor_puertas_test.go (pruebas unitarias DB)
		- backend/handlers/sensor_puertas_test.go (pruebas handlers: heartbeat y configuraci?n)
		- web/administrar_empresa/configuracion_de_estaciones.html (UI: registrar device ÃƒÂ¢Ã¢â‚¬Â ? estación)
		- web/administrar_empresa/estaciones.html (indicador visual sensor a?adido)
		- web/estilos.css (estilos del indicador)
	- Descripción: Se implement? un módulo ligero para registrar dispositivos Raspberry Pi por empresa y estación, recibir heartbeats públicos y reflejar el estado (negro/verde) en las tarjetas de estaciones. Incluye pruebas unitarias para DB y handlers.
	- Verificación: `go test ./...` ejecutado y tests verdes.

## 2026-04-11
- Generador automático de códigos de descuento: formato moderno `PREFIJO-XXXX-XXXX` (`DSCT-AB12-CD34`).
	- Archivos modificados: `backend/db/codigos_descuento.go`, `web/administrar_empresa/codigos_de_descuento.html`.
	- Se mantiene ?ndice ?nico por `(empresa_id, codigo)` y se implement? reintentos en inserci?n para manejar colisiones raras.
	- Se actualizaron `documentos/estructura_bd.md`, `documentos/descripcion_de_archivos` y `documentos/historial_de_cambios`.
	- Pruebas unitarias de DB asociadas: todas en verde.

## 2026-04-09
- Gobernanza de agente: se oficializa flujo DIAN SaaS multiempresa en instrucciones del repositorio.
	- `copilot-instructions.md` incorpora regla oficial: software DIAN compartido (un `Software ID`/`Software PIN` para la plataforma) con credenciales tributarias obligatorias por empresa (`nit`, `token_emisor_ref`, `certificado_clave_ref`).
	- Se explicita trazabilidad por `empresa_id` en cada envio real y prohibicion de reutilizar token/firma entre empresas.
	- Trazabilidad sincronizada en `documentos/historial_de_cambios`.
- Facturacion electronica DIAN (Colombia): modo SaaS multiempresa con software compartido y credenciales por empresa.
	- `backend/db/modulos_faltantes.go` amplia `empresa_dian_configuracion` con `usar_software_compartido`, `software_id_compartido_ref`, `software_pin_compartido_ref` e indice `ix_dian_empresa_shared_mode`.
	- `backend/handlers/modulos_faltantes.go` agrega resolución de software efectivo (`resolveDIANSoftwareCredentials`) con fallback global `DIAN_SHARED_SOFTWARE_ID/DIAN_SHARED_SOFTWARE_PIN`.
	- `sendDIANDocumentoReal` y `runDIANSetPruebasEnvio` reportan `software_modo` y `software_id` efectivo, manteniendo `NIT/token/certificado` por empresa.
	- `backend/handlers/modulos_faltantes_test.go` agrega `TestEmpresaDIANColombiaHandlerSoftwareCompartidoMultiempresa` (validado).
	- Documentacion sincronizada en `documentos/informacion_para_pruebas_plataforma_DIAN`, `documentos/estructura_bd.md`, `estructura_bd.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_archivos` y `documentos/historial_de_cambios`.
- Facturacion electronica DIAN (Colombia): se completa flujo de envio automatizado del set de habilitacion.
	- `backend/handlers/modulos_faltantes.go` agrega `action=enviar_set_pruebas` en `/api/empresa/facturacion_electronica/dian` para distribuir y enviar lotes de facturas/notas con resumen operacional por estado.
	- `sendDIANDocumentoReal` incorpora `documento_tipo` y override de `test_set_id` para interoperabilidad en el envio del lote.
	- `backend/handlers/modulos_faltantes_test.go` agrega `TestEmpresaDIANColombiaHandlerEnviarSetPruebas` y se valida junto con pruebas DIAN existentes (3 passed, 0 failed).
	- Se actualiza `documentos/informacion_para_pruebas_plataforma_DIAN` con aclaracion de configuracion de URL WSDL, resultados de pruebas y payload recomendado para los 50 documentos requeridos por DIAN.
	- Se sincroniza gobernanza documental en `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md` y `documentos/historial_de_cambios`.
- Documentacion DIAN: se crea `documentos/informacion_para_pruebas_plataforma_DIAN` para organizar en una sola referencia los datos de pruebas extraidos de `documentos/DATOS DIAN.mhtml`.
- Facturacion electronica DIAN (Colombia): carga de datos de prueba para Motel malibu con licencia activa.
	- Se registra configuracion DIAN en `backend/db/pcs_empresas` para `empresa_id=6` usando los datos de `documentos/DATOS DIAN.mhtml` (Software ID/PIN, TestSetId, prefijo, resolucion y rango).
	- Se valida tecnicamente el modulo DIAN con `handlers.EmpresaDIANColombiaHandler` ejecutando:
		- `checklist` y `validar` sin faltantes (`ok=true`).
		- `generar_cufe_demo` y `generar_xml_demo` con resultados correctos para documento de prueba.
	- En esta iteracion no fue necesario crear credenciales API adicionales en configuracion avanzada de super administrador para las pruebas internas de habilitacion.

## 2026-04-08
- Panel `administrar_empresa`: menu izquierdo desacoplado del contenido derecho.
	- `web/administrar_empresa.html` activa shell dedicado (`admin-empresa-shell`).
	- `web/estilos.css` separa scroll de sidebar e iframe para que navegar/desplazar el menu no afecte movimiento ni visibilidad de la subpagina cargada.
	- Incluye ajuste responsive para mantener usabilidad en pantallas moviles.

- Configuracion avanzada (DeepSeek/Gmail/Wompi): cifrado obligatorio robustecido y correcci?n de guardado cuando faltaba `CONFIG_ENC_KEY`.
	- `backend/main.go` ahora carga `.env.local/.env`, asegura `CONFIG_ENC_KEY` (autogenera y persiste en desarrollo cuando no existe) y normaliza secretos legacy para dejarlos cifrados.
	- `backend/handlers/super_config_backup_handlers.go` fuerza cifrado en restore de secretos y rechaza restore plano sin clave de cifrado.
	- `scripts/iniciar_servidor.ps1` valida/carga/autogenera `CONFIG_ENC_KEY` antes del arranque para evitar errores `400` en `/super/api/config/ai`.
	- `web/super/configuracion_avanzada.html` muestra en DeepSeek solo fecha/hora de ultima actualizaci?n, sin exponer fragmentos de credencial.
	- Enmascarado de secretos reforzado en backend (`ai_config_handlers.go`, `usuarios_empresa.go`, `payments_handlers.go`) con `********`.
	- Pruebas nuevas para restore cifrado y guardado cifrado DeepSeek en `backend/handlers/system_empresas_handlers_test.go`.

- Monitoreo operativo de arranque/reinicio de servidor con alerta por correo configurable.
	- `backend/db/super_servidor_eventos.go` agrega tabla de auditoria `super_servidor_eventos` para registrar inicio/reinicio, motivo, estado previo y resultado de notificacion.
	- `backend/handlers/server_runtime_notifications.go` implementa registro de arranque, deteccion de reinicio inesperado, escritura de estado runtime y bitacora local (`backend/logs/server_runtime_state.json`, `backend/logs/server_reinicio.log`).
	- `backend/main.go` integra registro de evento al arrancar, cierre controlado por seniales (`SIGINT/SIGTERM`) y trazabilidad de motivo de apagado.
	- `backend/handlers/usuarios_empresa.go` y `web/super/configuracion_avanzada.html` incorporan `gmail.restart_alert_to` en configuracion avanzada para correo destino de alertas de reinicio.
	- `backend/handlers/super_config_backup_handlers.go` incluye `gmail.restart_alert_to` dentro de claves criticas de backup/restore.
	- `scripts/iniciar_servidor.ps1` propaga `PCS_SERVER_START_REASON=inicio_script_iniciar_servidor` para enriquecer el motivo de arranque.
	- Pruebas nuevas/actualizadas en `backend/handlers/server_runtime_notifications_test.go` y `backend/handlers/system_empresas_handlers_test.go`.

- Script de despliegue local a GitHub mejorado en `scripts/actualizar_repositorio.ps1`.
	- Se corrige el armado de argumentos de `git push` para evitar enviar parametros vacios en PowerShell y reducir fallos intermitentes al subir cambios.
	- Se incorporan mensajes por etapas (`1/8` a `8/8`) con resumen final de commit, rama, remoto y estado de push.
	- Se centraliza el manejo de reintentos con `-ForcePush` y confirmacion explicita (`SI`) para mantener seguridad operacional.
	- Se refuerza el flujo de bitacoras automaticas para reportar mejor cuando falla el push documental.

- Arranque local y est?ticos web: mejoras en `scripts/iniciar_servidor.ps1` y correcci?n de ra?z `/`.
	- `scripts/iniciar_servidor.ps1` ahora muestra progreso por etapas (`1/8` a `8/8`), mensajes `[INFO]/[OK]/[AVISO]/[ERROR]` y salida expl?cita para `-Background` sin abrir navegador.
	- `backend/main.go` corrige la resolución de carpeta web para priorizar candidatos con `index.html`, evitando servir accidentalmente `backend/web` (solo `uploads/`).
	- `backend/main.go` agrega manejo de `/favicon.ico` con fallback a `web/img/punto_venta.png` para evitar 404 en consola.
	- `web/index.html` declara favicon explícito con `link rel="icon"`.
	- Validaciones: compilaci?n de `backend/main.go` (`go test . -run "^$"`) y parseo de PowerShell de `scripts/iniciar_servidor.ps1` OK.

- Backups empresariales: nueva opci?n para eliminar información por fecha de corte.
	- `backend/handlers/backups_empresariales.go` agrega `action=depurar_fecha` en `/api/empresa/backups`, con validación de `fecha_corte` y filtros opcionales `include_tables`/`exclude_tables`.
	- `backend/db/backups_empresariales.go` incorpora `PurgeEmpresaDataByDateCorte` para eliminar registros por `empresa_id` con fecha <= corte (inclusive), con detalle de eliminaciones por tabla.
	- La depuraci?n permite generar backup previo automático antes de ejecutar borrado para trazabilidad y recuperaci?n.
	- `backend/handlers/empresa_permisos.go` clasifica esta acci?n como `permActionApprove` en módulo seguridad.
	- `web/administrar_empresa/backups.html` incorpora UI de depuraci?n por fecha con confirmación expl?cita y resumen de resultados.
	- Se agregan pruebas: `TestEmpresaBackupsPurgeByDateCorte` (DB) y `TestEmpresaBackupsHandlerPurgeByDate` (handler).
	- Validaciones: pruebas de backups en verde y compilaci?n dirigida de paquetes backend críticos OK.

- Chat y tareas: documentos/fotos entre usuarios de empresa y administrador.
	- `backend/handlers/chat_tareas.go` deriva autor desde sesion autenticada (usuario/admin), evita suplantacion de `autor_*` y auto-registra participantes emisores en conversaciones.
	- Al crear conversacion desde usuario, se agrega automaticamente el admin propietario de la empresa como participante para habilitar intercambio usuario-admin.
	- Se amplian extensiones permitidas de adjuntos en backend y UI: `doc/docx/xls/xlsx/ppt/pptx/rtf/odt/ods/odp` (ademas de imagen/audio/pdf/txt/csv/json).
	- `web/administrar_empresa/chat_y_tareas.html` ahora envía metadata de actor efectiva (`autor_tipo`, `autor_ref_id`, `autor_nombre`, `autor_email`) segun sesion.
	- Se agrega `backend/handlers/chat_tareas_test.go` con pruebas para actor usuario derivado, upload `.docx` y auto-participacion usuario/admin.
	- Validaciones: pruebas dirigidas de handlers chat/tareas y compilacion de paquetes backend criticos en verde.

- Chat y tareas: higiene de pruebas y limpieza de artefactos temporales.
	- `backend/handlers/chat_tareas_test.go` incorpora limpieza automatica (`t.Cleanup`) de uploads temporales por empresa para mantener el workspace limpio tras las pruebas.
	- Se retiran artefactos locales residuales de validacion (`.docx` y binarios `.test.exe`) para evitar ruido en el arbol de cambios.
	- Validaciones: `go test ./handlers -run "TestEmpresaChatTareas" -count=1` y compilacion dirigida de paquetes backend (`./auth ./db ./handlers ./metrics ./utils`) en verde.

- Configuraci?n monetaria y num?rica por empresa en panel de configuraci?n.
	- `backend/db/empresa_configuracion_avanzada.go` ampl?a `empresa_configuracion_avanzada` con `moneda_codigo`, `sistema_numerico`, `usar_decimales` y `cantidad_decimales`.
	- `web/administrar_empresa/configuracion.html` agrega tarjeta para configurar moneda operativa, sistema num?rico y precisi?n decimal por empresa.
	- `backend/db/carritos_compras.go` aplica la moneda configurada por empresa como fallback al crear carritos sin moneda expl?cita.
	- `backend/main.go` registra la migraci?n `2026-04-08-030-configuracion-monetaria-numerica`.
	- Validaciones: compilaci?n de `db`, `handlers` y `main` en backend OK.

- Configuraci?n IA migrada de Gemini a DeepSeek en super administrador y chat empresarial corregido.
	- `web/super/configuracion_avanzada.html` ahora gestiona credencial `deepseek:deepseek-chat` y corrige flujo de guardado de credenciales IA.
	- `backend/handlers/ai_credentials_catalog.go` registra `DEEPSEEK_API_KEY` como credencial IA activa en panel super.
	- `backend/handlers/chat_con_inteligencia_artificial_controller.go` usa DeepSeek como proveedor del chat IA por empresa.
	- `web/administrar_empresa/chat_con_inteligencia_artificial.html` actualiza etiquetas/mensajes para modelo IA gen?rico (sin acoplamiento a Gemini).
	- Validaciones: compilaci?n de `handlers` y `main` en backend OK.

- Gobernanza documental reforzada para Agente Go y limpieza de documentos obsoletos.
	- Se actualiza `copilot-instructions.md` con regla obligatoria: si un modulo se crea o modifica, deben actualizarse `documentos/descripcion_de_modulos` y `documentos/matriz_roles_permisos_pos_multiempresa.md` en la misma iteracion.
	- Se refuerza la regla de sincronizacion de documentacion tecnica relacionada y trazabilidad en `documentos/descripcion_de_archivos` y `documentos/historial_de_cambios`.
	- Se elimina `documentos/modulos del proyecto.md` por duplicidad frente al documento canonico `documentos/descripcion_de_modulos`.
	- Se actualizan `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos` y `documentos/matriz_roles_permisos_pos_multiempresa.md` con la politica nueva.

- Cierre de pasos operativos 1, 2 y 3 solicitados en pendientes.
	- Paso 1: revisi?n/ajuste de accesos directos de módulos.
		- validación de consistencia de enlaces del panel empresa.
		- se agrega panel de accesos directos din?mico en `web/administrar_empresa/inicio.html` con visibilidad por permisos/licencia.
	- Paso 2: notas de voz en chat y tareas.
		- backend: `chat_tareas` incorpora campos `nota_voz_*` y endpoint `POST /api/empresa/chat_tareas/tareas/nota_voz`.
		- frontend: `chat_y_tareas.html` incorpora grabaci?n con MediaRecorder para mensajes/tareas, env?o y reproducci?n de audio.
	- Paso 3: super rol/permisos por licencia.
		- `licencias` incorpora `modulos_habilitados` y `super_rol_habilitado`.
		- middleware de permisos aplica restricciones por licencia y rol efectivo por empresa.
		- UI super de licencias permite configurar módulos habilitados y super rol por plan.
	- Validaciones:
		- `go test ./handlers -run "Test(EmpresaPermisosContextoHandlerRestringeModulosPorLicencia|WithEmpresaFinanzasPermissionsSupervisorConSuperRolLicencia|WithEmpresaVentasPermissionsBloqueaModuloNoHabilitadoPorLicencia|EmpresaPermisosContextoHandlerRetornaPermisosPorRol)$" -count=1` -> OK.
		- `go test ./db ./handlers -run "^$" -count=1` -> OK.
		- `go test . -run "^$" -count=1` -> OK.

- Sincronizacion documental de pendientes operativos en `Pendiente Notas`.
	- Se actualiza fecha de corte a 2026-04-08 y se consolida estado real de avance.
	- Se registran como completados en pendientes: soporte remoto empresarial y venta digital global.
	- Se deja explicito el bloque pendiente de siguiente iteracion: notas de voz en chat/tareas, super rol por licencia y sensor de puertas para motel.

- Cierre de validacion final del modulo de soporte remoto empresarial.
	- Validaciones ejecutadas al cierre:
		- `go test ./db -run "TestSoporteRemoto" -count=1` -> OK.
		- `go test ./handlers -run "Test(EmpresaSoporteRemotoHandlerFlow|PublicSoporteRemotoAgentHeartbeatAndStateUpdate)$" -count=1` -> OK.
		- `go test ./... -run "^$" -count=1` -> compilacion global backend OK.
	- Estado: modulo soporte remoto listo para cierre tecnico con pruebas dirigidas y validacion global en verde.

- Implementacion del modulo de soporte remoto empresarial (estilo AnyDesk/TeamViewer simplificado) con aislamiento por `empresa_id`.
	- Backend DB: se crea `backend/db/soporte_remoto.go` con tablas `empresa_soporte_remoto_configuracion`, `empresa_soporte_remoto_dispositivos` y `empresa_soporte_remoto_sesiones`, incluyendo validacion de PIN hash, heartbeat de agente y token temporal de visualizacion.
	- Backend handlers: se crea `backend/handlers/soporte_remoto.go` con endpoints:
		- empresarial: `GET/POST/PUT/PATCH /api/empresa/soporte_remoto` (configuracion, CRUD dispositivos, sesiones, aprobacion/finalizacion, resolver visualizacion y exportes en `pdf/xls/csv/json/txt`).
		- publico agente/plugin: `POST /api/public/soporte_remoto` (heartbeat, aprobar/finalizar sesion desde agente).
	- Integracion backend: `backend/main.go` registra `EnsureEmpresaSoporteRemotoSchema`, migracion `2026-04-08-029-soporte-remoto-empresa`, rutas protegida/publica; `backend/utils/utils.go` habilita la ruta publica del agente.
	- Seguridad y menu: `backend/handlers/empresa_permisos.go`, `web/administrar_empresa.html` y `web/js/administrar_empresa.js` agregan `linkSoporteRemoto` con control de permisos por modulo seguridad (`A`).
	- Frontend: se crean `web/administrar_empresa/soporte_remoto.html` y `web/administrar_empresa/soporte_remoto_view.html` para gestion de dispositivos/sesiones y visor embebido por token.
	- QA: se crean `backend/db/soporte_remoto_test.go` y `backend/handlers/soporte_remoto_test.go`.
	- Validaciones ejecutadas: `go test ./db -run "TestSoporteRemoto" -count=1` y `go test ./handlers -run "Test(EmpresaSoporteRemotoHandlerFlow|PublicSoporteRemotoAgentHeartbeatAndStateUpdate)$" -count=1` en verde (con incidencia local de lock temporal de `handlers.test.exe` al limpiar artefacto en Windows).

- Implementacion del modulo de venta digital global (super administrador + compra publica Wompi).
	- Backend DB: se crea `backend/db/venta_digital.go` con tablas `super_venta_digital_configuracion`, `super_venta_digital_items` y `super_venta_digital_ordenes`, incluyendo indices y flujo de orden/pago/entrega.
	- Backend handlers: se crea `backend/handlers/venta_digital.go` con endpoints:
		- super: `GET/POST/PUT/PATCH/DELETE /super/api/venta_digital` (configuracion, CRUD catalogo, uploads y ordenes).
		- publico: `GET/POST /api/public/venta_digital` (catalogo, crear pago Nequi por Wompi, consulta de estado y entrega por correo).
	- Integracion backend: `backend/main.go`, `backend/utils/utils.go` y `backend/handlers/payments_handlers.go` integran schema/rutas/middleware publico y sincronizacion por webhook Wompi para entrega de licencia.
	- Frontend: se crean `web/super/venta_digital.html` y `web/venta_digital.html`, y se agregan accesos en `web/menu.js`, `web/super_administrador.html` y `web/super/configuracion_avanzada.html`.
	- Validacion tecnica: `go test ./... -run "^$" -count=1` en `backend` -> compilacion global OK.

- Refuerzo de QA para permisos dinamicos por rol.
	- Testing DB: se crea `backend/db/roles_permisos_usuario_test.go` para validar replace/list/lookup y fallback sin tablas en el esquema de permisos de rol.
	- Testing Handler: se crea `backend/handlers/roles_tipos_usuario_permisos_test.go` con cobertura de `GET/PUT /super/api/roles_de_usuario/permisos` y caso `rol_id` inexistente.
	- Ajuste de fiabilidad en pruebas: `backend/handlers/empresa_permisos_test.go` alinea el helper de schema con columna `observaciones` requerida por consultas reales.
	- Validacion ejecutada: `go test ./db -run "TestRolesPermisos" -count=1` y `go test ./handlers -run "TestRolesDeUsuarioPermisosHandler|TestEmpresaPermisosContextoHandler|TestWithEmpresaSeguridadPermissionsRequiereAprobacionParaCambioPermisos|TestSuperEndpointsPermisosPorRol" -count=1` en verde.

- Inicio de implementacion del modulo de permisos dinamicos por rol.
	- Backend: nuevas tablas y capa DB para overrides por `rol` en modulo/accion y pagina (`roles_de_usuario_permisos`, `roles_de_usuario_paginas_permisos`).
	- Backend: middleware de permisos empresariales ahora aplica overrides dinamicos y `/api/empresa/permisos_contexto` incluye mapa `paginas`.
	- Backend: nuevo endpoint super `GET/PUT /super/api/roles_de_usuario/permisos` para gestionar matriz de permisos por rol.
	- Frontend: nueva pantalla `/super/permisos_rol.html`, acceso desde menu super y boton directo en listado de roles.
	- Frontend: menu empresa aplica visibilidad por pagina desde contexto de permisos.
	- Validacion tecnica: pruebas dirigidas de permisos en handlers y compilacion de `main` en verde.

- Plan profesional agregado al tablero de pendientes para completar e integrar el modulo de roles y permisos por usuario de empresa.
	- Se actualiza `Pendiente Notas` al final del documento con fases de implementacion (1..10), cronograma sugerido y criterios de aceptacion.
	- Se mantiene enfoque de integracion multiempresa con aislamiento por `empresa_id`, trazabilidad y UAT por rol.

## 2026-04-07
- Normalizacion documental del tablero de pendientes (modulo 35).
	- Se ajusta `Pendiente Notas` para reemplazar la etiqueta `COMPLETADO PARCIAL` por contexto historico de fases.
	- Se mantiene el estado oficial de cierre total del modulo 35 sin cambios funcionales.

- Continuacion de cierre de pendientes: tablero operativo actualizado con evidencia de pruebas del modulo 37.
	- Se actualiza `Pendiente Notas` para a?adir ejecucion dirigida de pruebas de `venta_publica` en handlers.
	- Se explicita estado general sin pendientes de modulos (`1..38` y bloque "Ultimo" en `COMPLETADO`).

- Pruebas dirigidas para el modulo 37 (Venta publica por empresa + Wompi).
	- Se agrega `backend/handlers/venta_publica_test.go` con cobertura de:
		- flujo empresarial (`config`, `crear`, `detalle`, `catalogo`, `activar/desactivar`).
		- flujo publico de catalogo y creacion de orden con Wompi inactivo (respuesta controlada `412`).
		- validacion de `estado_pago` cuando no se envia `order_code`.
	- Validacion: `runTests` en `backend/handlers/venta_publica_test.go` -> 3 passed, 0 failed.

- Cierre del modulo 37 (Venta publica por empresa + Wompi) y cierre de pendientes documentales 38/Ultimo.
	- Backend `db`:
		- `backend/db/venta_publica.go` (nuevo) agrega tablas `empresa_venta_publica_configuracion`, `empresa_venta_publica_items` y `empresa_venta_publica_ordenes`, con operaciones CRUD/listado/ordenes y resolucion por slug.
	- Backend `handlers`:
		- `backend/handlers/venta_publica.go` (nuevo) agrega `/api/empresa/venta_publica` y `/api/public/venta_publica`.
		- soporta configuracion Wompi por empresa, carga de imagen de catalogo, creacion de pago Nequi y consulta de estado de orden.
	- Integracion:
		- `backend/main.go` integra schema, migracion `2026-04-07-028-venta-publica-wompi`, rutas API y rewrite de `/{slug}/venta_publica.html`.
		- `backend/utils/utils.go` habilita acceso publico a rutas/API de venta publica y recursos en `/uploads/`.
	- Frontend:
		- `web/administrar_empresa/venta_publica.html` (nuevo) para administracion del canal online por empresa.
		- `web/venta_publica.html` (nuevo) para clientes finales (catalogo, carrito, pago y estado).
		- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` agregan `linkVentaPublica` con permisos de ventas.
	- Documentacion:
		- se crea `documentos/descripcion_de_modulos` (modulo 38).
		- se amplia `web/ayuda/ayuda.html` con tutoriales de ventas, productos/impuestos, venta publica y configuracion por empresa (cierre del "Ultimo").
		- se sincronizan `documentos/estructura_bd.md`, `estructura_bd.md` y diagramas (`estructura_del_codigo`, `diagrama_flujo_procesos`, `diagrama_entidad_relacion`).
	- Validaciones ejecutadas:
		- `go test ./... -run "^$" -count=1` -> compilacion global backend OK.
		- `get_errors` en archivos modificados -> sin errores.

- Cierre del modulo 36 (Backups empresariales): snapshots por empresa, restauracion trazable, exportacion multiformato y UI dedicada.
	- Backend `db`:
		- `backend/db/backups_empresariales.go` agrega tablas `empresa_backups` y `empresa_backups_restauraciones` con `EnsureEmpresaBackupsSchema`.
		- implementa construccion de payload (`BuildEmpresaBackupPayload`), alta de snapshot (`CreateEmpresaBackupSnapshot`), historial/detalle (`List/Get`) y restauracion controlada (`RestoreEmpresaBackupByID`).
		- incorpora trazabilidad de integridad (`hash_contenido`) y metadata/version de snapshot.
		- `backend/db/backups_empresariales_test.go` agrega pruebas de flujo snapshot/restauracion y listado/payload.
	- Backend `handlers`:
		- `backend/handlers/backups_empresariales.go` agrega endpoint `/api/empresa/backups` con acciones `listar|crear|detalle|export|restaurar|activar|desactivar`.
		- `backend/handlers/backups_empresariales_test.go` agrega cobertura de create/list/detail/export/restore/toggle y not-found en restore.
		- `backend/handlers/empresa_permisos.go` clasifica `restaurar|restore` como accion de aprobacion (`permActionApprove`).
	- Integracion y frontend:
		- `backend/main.go` registra `EnsureEmpresaBackupsSchema`, migracion `2026-04-07-027-backups-empresariales` y ruta protegida `/api/empresa/backups`.
		- `web/administrar_empresa/backups.html` (nuevo) implementa flujo profesional de backups por empresa.
		- `web/administrar_empresa.html`, `web/js/administrar_empresa.js` y `web/estilos.css` integran acceso `linkBackups` y estilos del modulo.
	- Validaciones ejecutadas:
		- `go test ./db -run "^TestEmpresaBackups" -count=1` -> OK.
		- `go test ./handlers -run "^TestEmpresaBackupsHandler" -count=1` -> OK.
		- `go test . -run "^$" -count=1` -> compilacion de `main` OK.

- Cierre del modulo 35 (Creditos y cartera): reglas de limites por cliente, permisos finos por rol en workflow y auditoria ampliada.
	- Backend `db`:
		- `backend/db/creditos.go` agrega tabla `empresa_creditos_clientes_limites` con indice unico `(empresa_id, cliente_id)` y funciones `Get/List/Upsert/SetEstado` para administrar limites por cliente.
		- se incorpora validacion de limites por cliente en `CreateEmpresaCredito` y `UpdateEmpresaCredito` (saldo total maximo y maximo de creditos activos).
	- Backend `handlers`:
		- `backend/handlers/creditos.go` agrega acciones `limites_cliente`, `limite_cliente`, `upsert_limite_cliente` y `eliminar_limite_cliente` en `/api/empresa/creditos`.
		- se incorpora validacion de permiso fino por tipo de workflow: `contabilidad` puede decidir `reverso_abono` y `refinanciacion` queda restringida a `administrador`.
		- se amplian eventos de auditoria no bloqueante para solicitud/aprobacion/rechazo de workflow, cambios de limites y denegaciones por permiso fino.
	- Pruebas:
		- `backend/db/creditos_test.go` agrega `TestEmpresaCreditosClienteLimitesBloqueaExceso` y `TestEmpresaCreditosClienteLimitesCRUD`.
		- `backend/handlers/creditos_test.go` agrega `TestEmpresaCreditosHandlerLimitesClienteYBloqueo` y `TestEmpresaCreditosHandlerWorkflowPermisoFinoPorTipo`.
	- Validaciones ejecutadas:
		- `go test ./db -run "TestEmpresaCreditosClienteLimites(BloqueaExceso|CRUD)$" -count=1` -> OK.
		- `go test ./handlers -run "TestEmpresaCreditosHandler(LimitesClienteYBloqueo|WorkflowPermisoFinoPorTipo)$" -count=1` -> OK.
		- `runTests` sobre `backend/db/creditos_test.go` y `backend/handlers/creditos_test.go` (casos nuevos) -> 6 passed, 0 failed.

- Avance del modulo 35 (Creditos y cartera): workflow avanzado de reversos/anulaciones y refinanciacion con aprobacion multinivel.
	- Backend `db`:
		- `backend/db/creditos.go` agrega tabla `empresa_creditos_workflow`, filtros/listado por estado/tipo y funciones de negocio para `solicitud`, `aprobacion`, `rechazo` y ejecucion automatica.
		- se implementa ejecucion de `reverso_abono` y `refinanciacion` con trazabilidad de `movimiento_resultado_id`, `resultado_json` e `historial_aprobaciones_json`.
		- se corrige colision de `numero_cuota` en refinanciacion generando nuevas cuotas con secuencia incremental despues del ultimo numero historico.
		- `backend/db/creditos_test.go` agrega `TestEmpresaCreditosWorkflowReversoAprobadoEjecutaReversion` y `TestEmpresaCreditosWorkflowRefinanciacionAprobadaRegeneraCuotas`.
	- Backend `handlers`:
		- `backend/handlers/creditos.go` agrega acciones `workflows`, `solicitar_reverso`, `solicitar_refinanciacion`, `aprobar_workflow`, `rechazar_workflow`.
		- `backend/handlers/empresa_permisos.go` clasifica acciones de aprobacion/rechazo de workflow como `permActionApprove` en modulo finanzas.
		- `backend/handlers/creditos_test.go` agrega `TestEmpresaCreditosHandlerWorkflowReversoSolicitudYAprobacion`.
	- Validaciones ejecutadas:
		- `go test ./db -run "TestEmpresaCreditosWorkflow(ReversoAprobadoEjecutaReversion|RefinanciacionAprobadaRegeneraCuotas)$" -count=1` -> OK.
		- `go test ./handlers -run "TestEmpresaCreditosHandlerWorkflowReversoSolicitudYAprobacion$" -count=1` -> OK.
		- `go test ./... -count=1` -> OK.

- Avance del modulo 35 (Creditos y cartera): integracion contable/caja-bancos/pasarelas en abonos con asientos automaticos por politica.
	- Backend `db`:
		- `backend/db/eventos_contables.go` extiende contrato contable con `creditos.credito_abono_registrado`.
		- agrega plantilla de lineas contables para abonos de credito (caja/bancos, cartera de creditos, intereses y mora).
		- `backend/db/eventos_contables_test.go` agrega `TestProcessEmpresaEventosContablesPendientesCreditoAbonoGeneraLineasCartera`.
	- Backend `handlers`:
		- `backend/handlers/creditos.go` integra registro de evento contable al `action=abono` y procesamiento automatico de asientos por politica (`procesar_asientos`, `asientos_limit`, `max_reintentos`).
		- se incorpora metrica de canal de pago (`caja`, `bancos`, `pasarela`) para trazabilidad operativa por abono.
		- `backend/handlers/creditos_test.go` agrega `TestEmpresaCreditosHandlerAbonoIntegraContabilidadYAsientos`.
	- Trazabilidad funcional:
		- `Pendiente Notas` marca completada la integracion contable de modulo 35 y mantiene pendientes de reversos/refinanciacion y limites/permisos.
	- Validaciones ejecutadas:
		- `go test ./db -run "Test(EmpresaEventosContablesCreateAndList|ProcessEmpresaEventosContablesPendientesGeneraAsientosIdempotentes|ProcessEmpresaEventosContablesPendientesCreditoAbonoGeneraLineasCartera|EmpresaCreditosFlowCrearCuotasAbonoYResumen|EmpresaCreditosMoraDashboard)$" -count=1` -> OK.
		- `go test ./handlers -run "TestEmpresaCreditosHandler(FlujoBasico|AlertasMoraYReporte|AbonoIntegraContabilidadYAsientos)$" -count=1` -> OK.
		- `go test ./... -run "^$" -count=1` -> compilacion global backend OK.

- Avance del modulo 35 (Creditos y cartera): alertas proactivas de vencimiento y ranking avanzado de morosidad.
	- Backend `db`:
		- `backend/db/creditos.go` agrega dashboard de morosidad (`GetEmpresaCreditosMoraDashboard`) con bloques de proximos a vencer, vencidos y ranking.
		- `backend/db/creditos_test.go` agrega `TestEmpresaCreditosMoraDashboard`.
	- Backend `handlers`:
		- `backend/handlers/creditos.go` agrega acciones `alertas|alertas_mora|morosidad|ranking_morosidad`.
		- `action=reporte` soporta `tipo=morosidad` para exportacion en `json/csv/txt/xls/pdf`.
		- `backend/handlers/creditos_test.go` agrega `TestEmpresaCreditosHandlerAlertasMoraYReporte`.
	- Frontend:
		- `web/administrar_empresa/creditos.html` incorpora panel operativo de alertas/ranking con filtros (`dias_proximos`, `top`, `include_inactive`) y exportacion dedicada.
		- `web/estilos.css` incorpora estilos `creditos-alertas-*` para toolbar y grilla responsive.
	- Diagramas y trazabilidad:
		- `documentos/diagramas/estructura_del_codigo.md` y `documentos/diagramas/diagrama_flujo_procesos.md` se actualizan con el nuevo flujo de morosidad.
	- Validaciones ejecutadas:
		- `go test ./db -run TestEmpresaCreditosMoraDashboard -count=1` -> OK.
		- `go test ./handlers -run TestEmpresaCreditosHandlerAlertasMoraYReporte -count=1` -> OK.
		- `go test ./... -run "^TestEmpresaCreditos" -count=1` -> OK.

- Avance del modulo 35 (Creditos y cartera): publicadas guias operativas en centro de ayuda y manual por rol.
	- Documentacion funcional:
		- `web/ayuda/ayuda.html` integra acceso rapido a creditos, bloque tutorial `30) Creditos y cartera`, guia operativa dedicada por flujo y manual por rol (administrador, caja/cobranza y auditoria).
		- Se documentan endpoints clave de `/api/empresa/creditos` en la seccion de APIs para operacion y soporte.
	- Trazabilidad:
		- `Pendiente Notas` retira del listado pendiente del modulo 35 la tarea de guias operativas y la marca dentro de completado parcial.
		- `documentos/descripcion_del_proyecto` sincroniza el alcance del modulo 35 incluyendo referencia a la guia operativa por rol.
	- Validaciones ejecutadas:
		- validacion de editor (`get_errors`) sobre `web/ayuda/ayuda.html`, `Pendiente Notas`, `CHANGELOG.md`, `documentos/historial_de_cambios`, `documentos/descripcion_de_archivos` y `documentos/descripcion_del_proyecto` -> sin errores.

- Avance del modulo 35 (Creditos y cartera): fase 2 frontend base implementada con pantalla dedicada e integracion de menu/permisos.
	- Frontend:
		- `web/administrar_empresa/creditos.html` (nuevo) incorpora formulario de creacion de credito, filtros de cartera, resumen, tabla de creditos, panel de abonos y estado de cuenta (cuotas/movimientos).
		- integra exportacion de cartera en `json/csv/txt/xls/pdf` usando `action=reporte` del backend.
		- incluye acciones de operacion diaria: prellenado de abono, cambio de estado de credito y activar/desactivar fila.
	- Navegacion y permisos:
		- `web/administrar_empresa.html` agrega enlace lateral `linkCreditos`.
		- `web/js/administrar_empresa.js` agrega `linkCreditos` al catalogo de permisos como modulo `finanzas` accion `C`.
	- Estilos:
		- `web/estilos.css` agrega componentes `creditos-*` para grids de filtros/resumen, acciones de tabla y detalle responsive.
	- Validaciones ejecutadas:
		- validacion de editor (`get_errors`) sobre archivos frontend modificados -> sin errores.

- Avance del modulo 35 (Creditos y cartera): fase 1 backend implementada con esquema, API y pruebas base.
	- Backend `db`:
		- `backend/db/creditos.go` agrega tablas `empresa_creditos`, `empresa_creditos_cuotas` y `empresa_creditos_movimientos`.
		- implementa creacion de creditos, generacion automatica de cuotas, registro de abonos y resumen de cartera.
		- `backend/db/creditos_test.go` agrega `TestEmpresaCreditosFlowCrearCuotasAbonoYResumen`.
	- Backend `handlers`:
		- `backend/handlers/creditos.go` expone `GET/POST/PUT/PATCH/DELETE /api/empresa/creditos`.
		- incorpora acciones `estado_cuenta`, `resumen_cartera`, `movimientos`, `cuotas`, `abono` y `reporte`.
		- soporta exportacion de reporte en `json/csv/txt/xls/pdf`.
		- `backend/handlers/creditos_test.go` agrega `TestEmpresaCreditosHandlerFlujoBasico`.
	- Integracion:
		- `backend/main.go` registra `EnsureEmpresaCreditosSchema`, migracion `2026-04-07-026-creditos-cartera` y ruta protegida `/api/empresa/creditos`.
	- Validaciones ejecutadas:
		- `go test ./db -run TestEmpresaCreditosFlowCrearCuotasAbonoYResumen -count=1` -> OK.
		- `go test ./handlers -run TestEmpresaCreditosHandlerFlujoBasico -count=1` -> OK.
		- `go test ./... -run "^$" -count=1` -> compilacion global backend OK.

- Cierre del modulo 34 (Calculadora por empresa): historial operativo persistente por empresa, asociaciones trazables y exportacion multiformato por rango/usuario.
	- Backend `db`:
		- `backend/db/calculadora_operativa.go` agrega tablas `empresa_calculadora_configuracion` y `empresa_calculadora_operaciones` con filtros por fecha/usuario/cliente/etiqueta.
		- se incorporan operaciones de configuracion (`integrar_carritos`, `integrar_cotizaciones`), registro de operaciones etiquetadas y limpieza logica por filtros.
		- `backend/db/calculadora_operativa_test.go` agrega `TestEmpresaCalculadoraConfiguracionYHistorialFlow`.
	- Backend `handlers`:
		- `backend/handlers/calculadora_operativa.go` expone `GET/POST/PUT/DELETE /api/empresa/calculadora` con acciones `config`, `referencias`, `export`, `limpiar`, `activar/desactivar`.
		- valida referencias opcionales de `carrito_id`/`cotizacion_id` segun configuracion y conserva trazabilidad por `empresa_id`, cliente/documento, carrito/cotizacion y usuario.
		- `backend/handlers/calculadora_operativa_test.go` agrega `TestEmpresaCalculadoraHandlerConfigOperacionesFiltrosYExport`.
	- Frontend:
		- `web/administrar_empresa/calculadora.html` migra de historial local a flujo API, con filtros por rango/usuario, etiquetas, asociaciones a cliente/documento y exportacion backend.
		- `web/estilos.css` agrega estilos `calc-config-row`, `calc-meta-grid` y `calc-filter-row` para controles de configuracion/metadata/filtros en desktop y mobile.
	- Integracion:
		- `backend/main.go` registra `EnsureEmpresaCalculadoraSchema`, migracion `2026-04-07-025-calculadora-operativa` y ruta protegida `/api/empresa/calculadora` bajo permisos de finanzas.
	- Validaciones ejecutadas:
		- `go test ./db -run TestEmpresaCalculadoraConfiguracionYHistorialFlow -count=1` -> OK.
		- `go test ./handlers -run TestEmpresaCalculadoraHandlerConfigOperacionesFiltrosYExport -count=1` -> OK.

- Cierre del modulo 33 (Configuracion operativa de cobro): politicas contextuales, simulador de reglas e historial con rollback operativo.
	- Backend `db`:
		- `backend/db/configuracion_operativa.go` agrega tablas y modelos `empresa_configuracion_operativa_politicas` y `empresa_configuracion_operativa_historial`.
		- se incorpora resolución de rollback.
		- `backend/db/configuracion_operativa_test.go` agrega `TestEmpresaConfiguracionOperativaPoliticaContextoYRollback`.
	- Backend `handlers`:
		- `backend/handlers/configuracion_operativa.go` ampl?a acciones HTTP con `action=politica`, `action=simular`, `action=historial` y `action=rollback`.
		- se agrega snapshot de trazabilidad no bloqueante tras publicaciones y simulaciones guardadas.
		- `backend/handlers/configuracion_operativa_test.go` agrega `TestEmpresaConfiguracionOperativaHandlerPoliticaSimulacionHistorialYRollback`.
	- Frontend:
		- `web/administrar_empresa/configuracion.html` agrega UI de politica contextual, simulador por contexto y panel de historial/rollback.
		- `web/estilos.css` incorpora estilos para el bloque operativo extendido de simulacion e historial.
	- Validaciones ejecutadas:
		- `go test ./db -run TestEmpresaConfiguracionOperativa -count=1` -> OK.
		- `go test ./handlers -run TestEmpresaConfiguracionOperativaHandler -count=1` -> OK.

- Cierre del modulo 32 (Graficos y estadisticas): cache por panel, comparativos entre periodos, filtros avanzados y optimizacion de series largas.
	- Backend `handlers`:
		- `backend/handlers/graficos_estadisticas.go` incorpora cache en memoria con `cache.hit`, soporte de comparativo (`comparar`, `comparar_desde`, `comparar_hasta`) y filtros avanzados (`sucursal_id`, `estacion_id`, `segmento`).
		- Se agrega cobertura de filtros en respuesta (`filtros.cobertura`) y aplicacion de snapshots para mantener KPI del tablero alineados con filtros aplicados.
		- Se reemplaza truncamiento de cola por compactacion por buckets en series de ventas/finanzas/compras/asistencia para rangos extensos.
	- Frontend:
		- `web/administrar_empresa/graficos_estadisticas.html` agrega controles avanzados de filtros, comparativo, refresco sin cache y tarjetas de variacion por metrica.
		- `web/estilos.css` agrega estilos de comparativo, tendencia y comportamiento responsive para la nueva capa de filtros.
	- Pruebas:
		- `backend/handlers/graficos_estadisticas_test.go` agrega `TestEmpresaGraficosEstadisticasHandlerFiltrosComparativoYCache` para validar filtros, comparativo y cache hit/miss.
	- Validaciones ejecutadas:
		- `go test ./handlers -run TestEmpresaGraficosEstadisticasHandler -v` -> OK.
		- `go test ./handlers -count=1` -> OK.

## 2026-04-07
- Hotfix de arranque en migraciones ERP legacy (modulos faltantes): correccion de orden de creacion de indices dependientes de columnas nuevas.
	- Backend `db`:
		- `backend/db/modulos_faltantes.go` evita crear en el bloque inicial los indices que dependen de columnas agregadas por migracion (`periodo_contable`, `bloqueado_venta`, campos de aprobacion/nomina RRHH).
		- Esos indices se mantienen en la fase final posterior a `ensureColumnIfMissing`, garantizando compatibilidad con bases antiguas.
	- Validaciones ejecutadas:
		- `go test ./... -run "^$" -count=1` -> compilacion global backend OK.
		- `go run .` -> backend inicia correctamente y queda en `LISTENING` en puerto `8080` (sin error `no such column: periodo_contable`).

## 2026-04-07
- Cierre del modulo 31 (Reportes): programacion automatica, versionado de plantillas y validacion de consistencia multiformato.
	- Backend `handlers`:
		- `backend/handlers/reportes_programacion.go` consolida agenda y ejecucion de reportes (`action=programacion`, `action=ejecutar_programacion`, `action=ejecuciones`, `action=validar_consistencia`).
		- Se corrige robustez en listado de ejecuciones para manejar campos `NULL` de motor legado retirado (`error_detalle`, `programacion_id`, metadatos opcionales) sin error `500`.
	- Pruebas:
		- `backend/handlers/reportes_programacion_test.go` valida versionado de plantillas y ciclo completo de programacion/ejecucion/consistencia.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "TestEmpresaReportesHandler(PlantillasVersionado|ProgramacionEjecucionYConsistencia)$" -count=1` -> OK.
		- `go test ./... -run "^$" -count=1` -> compilacion global backend OK.

## 2026-04-07
- Cierre del modulo 30 (Seguridad y permisos): deny-by-default por endpoint, matriz automatizada rol/modulo y trazabilidad de aprobacion.
	- Backend `handlers`:
		- `backend/handlers/empresa_permisos.go` exige evidencia de aprobacion trazable (`aprobado_por`, `codigo_aprobacion`) para cambios criticos de permisos en `/api/empresa/usuarios` bajo modulo seguridad.
		- Se propagan cabeceras de aprobacion para trazabilidad (`X-Permission-Approved-By`, `X-Permission-Approval-Code`, `X-Permission-Approval-Reason`) y bandera `X-Permission-Approval-Required`.
		- `backend/handlers/auditoria_empresa.go` registra metadata de aprobacion en `empresa_auditoria_eventos` (`permission_approval_required`, `permission_approved_by`, `permission_approval_code`, `permission_approval_reason`).
	- Pruebas:
		- `backend/handlers/empresa_permisos_test.go` agrega cobertura de matriz completa rol/modulo/accion y de aprobacion trazable para cambios de permisos.
		- `backend/main_empresa_routes_security_test.go` agrega barrido automatizado deny-by-default para todas las rutas `/api/empresa/*` registradas por `http.HandleFunc`.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "Test(EmpresaPermisosContextoHandlerRetornaPermisosPorRol|EmpresaPermisosContextoHandlerIncluyeMatrizRoles|EmpresaPermisosContextoHandlerMatrizRolesCumplePoliticaPorModuloAccion|WithEmpresaSeguridadPermissionsRequiereAprobacionParaCambioPermisos|WithEmpresaSeguridadPermissionsAceptaAprobacionTrazableYRegistraMetadata)$" -count=1` -> OK.
		- `go test . -run "TestEmpresaRoutesUsePermissionWrappers$" -count=1` -> OK.
		- `go test ./... -run "^$" -count=1` -> compilacion global backend OK.

## 2026-04-07
- Cierre del modulo 29 (Auditoria empresarial): busqueda full-text con filtros y exportacion forense con cadena de custodia basica.
	- Backend `db`:
		- `backend/db/auditoria_empresa.go` agrega soporte de busqueda `search` con FTS (cuando esta disponible) y fallback por `LIKE`, manteniendo compatibilidad con filtros avanzados existentes.
		- Se incorpora inicializacion de esquema FTS de auditoria (tabla virtual, triggers y backfill) para indexacion de `modulo`, `accion`, `recurso`, `endpoint`, `metadata_json` y `observaciones`.
	- Backend `handlers`:
		- `backend/handlers/auditoria_empresa.go` extiende `GET /api/empresa/auditoria/eventos` con `action=export_forense|forense_export|cadena_custodia` y `format=json|csv`.
		- La exportacion forense genera `hash_registro`, `hash_cadena` y `hash_global` para trazabilidad basica de cadena de custodia.
	- Pruebas:
		- `backend/db/auditoria_empresa_test.go`: `TestListEmpresaAuditoriaEventosSearchFullTextConFiltros`.
		- `backend/handlers/auditoria_empresa_test.go`: `TestEmpresaAuditoriaEventosHandlerExportForenseJSONYCSV`.
	- Validaciones ejecutadas:
		- `go test ./db -run "Test(CreateAndListEmpresaAuditoriaEventos|PurgeEmpresaAuditoriaEventos|PurgeExpiredEmpresaAuditoriaEventos|CountAndListEmpresaAuditoriaEventosWithPaginationAndSearch|CreateEmpresaAuditoriaEventoAplicaPoliticaRetencionPorModuloYSeveridad|CreateEmpresaAuditoriaEventoMantieneRetencionExplicita|ListEmpresaAuditoriaEventosSearchFullTextConFiltros)$" -count=1` -> OK.
		- `go test ./handlers -run "TestEmpresaAuditoriaEventosHandler(ConsultaYPurga|FiltrosAvanzados|ExportForenseJSONYCSV)$" -count=1` -> OK.
		- `go test ./... -run "^$" -count=1` -> compilacion global backend OK.

## 2026-04-07
- Avance del modulo 29 (Auditoria empresarial): politicas de retencion por modulo y severidad.
	- Backend `db`:
		- `backend/db/auditoria_empresa.go` ahora resuelve `retencion_dias` automaticamente por combinacion de `modulo` + `severidad` inferida (resultado/codigo HTTP/metadatos), manteniendo prioridad para `retencion_dias` explicita.
		- Se enriquece `metadata_json` con trazabilidad de politica aplicada (`retencion_politica_modulo`, `retencion_politica_severidad`, `retencion_dias_resuelto`).
	- Pruebas:
		- `backend/db/auditoria_empresa_test.go` agrega:
			- `TestCreateEmpresaAuditoriaEventoAplicaPoliticaRetencionPorModuloYSeveridad`.
			- `TestCreateEmpresaAuditoriaEventoMantieneRetencionExplicita`.
	- Validaciones ejecutadas:
		- `go test ./db -run "TestCreateEmpresaAuditoriaEventoAplicaPoliticaRetencionPorModuloYSeveridad|TestCreateEmpresaAuditoriaEventoMantieneRetencionExplicita" -count=1 -v` -> OK.
		- `go test ./db -run "Test(CreateAndListEmpresaAuditoriaEventos|PurgeEmpresaAuditoriaEventos|PurgeExpiredEmpresaAuditoriaEventos|CountAndListEmpresaAuditoriaEventosWithPaginationAndSearch|CreateEmpresaAuditoriaEventoAplicaPoliticaRetencionPorModuloYSeveridad|CreateEmpresaAuditoriaEventoMantieneRetencionExplicita)$" -count=1` -> OK.
		- `go test ./handlers -run "TestEmpresaAuditoriaEventosHandlerConsultaYPurga|TestEmpresaAuditoriaEventosHandlerFiltrosAvanzados" -count=1` -> OK.
		- `go test ./... -run "^$" -count=1` -> compilacion global backend OK.

## 2026-04-07
- Cierre del pendiente de modulo 28 (Finanzas y contabilidad): politicas de cierre/reapertura de periodos con evidencia de autorizacion.
	- Backend `handlers`:
		- `backend/handlers/finanzas.go` exige en `PUT /api/empresa/finanzas/periodosaction=cerrar|reabrir` los campos `autorizado_por`, `motivo_autorizacion` y `evidencia_autorizacion`.
		- Se incorpora trazabilidad explicita en observaciones y payload de evento contable (`policy_autorizacion`, `autorizado_por`, `motivo_autorizacion`, `evidencia_autorizacion`, `codigo_autorizacion`, `ejecutado_por`).
		- La respuesta HTTP del cierre/reapertura retorna bloque `autorizacion` para auditoria operativa.
	- Pruebas:
		- `backend/handlers/eventos_contables_modulos_test.go` actualiza `TestEmpresaFinanzasEmiteEventosContables` para validar evidencia en payload.
		- Se agrega `TestEmpresaFinanzasPeriodosRequiereEvidenciaAutorizacion` para rechazar cierre sin evidencia.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "TestEmpresaFinanzasEmiteEventosContables|TestEmpresaFinanzasPeriodosRequiereEvidenciaAutorizacion" -count=1 -v` -> OK.
		- `go test ./handlers -run "TestEmpresaFinanzas" -count=1` -> OK.

## 2026-04-07
- Avance del modulo 28 (Finanzas y contabilidad): conciliacion bancaria automatica y tablero de desviaciones por periodo.
	- Backend `db`:
		- Se agrega tabla `empresa_finanzas_bancos_movimientos` (extractos bancarios por `empresa_id`) en `EnsureEmpresaFinanzasSchema`.
		- Se agrega `backend/db/finanzas_conciliacion_bancaria.go` con:
			- importacion idempotente de extractos por `hash_movimiento`.
			- conciliacion bancaria automatica contra `empresa_finanzas_movimientos` con tolerancia de monto/dias.
			- resumen de conciliacion/desviaciones por periodo.
	- Backend `handlers`:
		- Se amplia `EmpresaFinanzasMovimientosHandler` con acciones:
			- `GET action=conciliacion_bancaria` y `GET action=conciliacion_bancaria_export`.
			- `GET action=extractos_bancarios`.
			- `POST action=importar_extractos_bancarios` (opcional `auto_conciliar`).
			- `PUT action=conciliar_bancaria_auto`.
		- Se actualiza `resolveFinanzasPermissionAction` para clasificar conciliacion bancaria automatica como `permActionApprove`.
	- Pruebas:
		- `backend/db/finanzas_test.go`: `TestEmpresaFinanzasConciliacionBancariaAutomatica`.
		- `backend/handlers/eventos_contables_modulos_test.go`: `TestEmpresaFinanzasMovimientosHandlerConciliacionBancariaAutomatica`.
	- Validaciones ejecutadas:
		- `runTests` sobre pruebas nuevas de db/handlers -> OK.
		- `go test ./... -run "^$" -count=1` -> compilacion global backend OK.

## 2026-04-07
- Refuerzo de cobertura en capas `auth`, `metrics` y `utils`:
	- Se agregan y amplian pruebas unitarias:
		- `backend/auth/auth_test.go`
		- `backend/metrics/collector_test.go`
		- `backend/utils/utils_test.go` (incluye pruebas de middleware, contexto y manejo de errores JSON)
	- Cobertura actualizada por paquete (corte de ejecucion):
		- `auth`: `85.3%`
		- `db`: `51.4%`
		- `handlers`: `50.4%`
		- `metrics`: `78.0%`
		- `utils`: `71.1%`
	- Se actualiza evidencia en `Pendiente Notas`, `documentos/punto_13_calidad_uat_despliegue.md` y `documentos/punto_13_validacion_integral_resultado.md`.
	- Validaciones ejecutadas:
		- `runTests` sobre `backend/utils/utils_test.go` -> 16 pruebas OK.
		- `go test ./auth ./db ./handlers ./metrics ./utils -cover -count=1` (OK).
		- `go test ./... -run "^$" -count=1` (compilacion global backend OK).

## 2026-04-07
- Cierre transversal de calidad y salida controlada:
	- Se actualiza `documentos/punto_13_calidad_uat_despliegue.md` con:
		- objetivo minimo de cobertura por capa,
		- acta UAT formal por rol (`super_admin`, `admin_empresa`, `usuario_empresa`),
		- matriz UAT por modulo en estado aprobado.
	- Se ampl?a `documentos/release_checklist.md` con checklist estandar "listo para produccion" por modulo (seguridad, rendimiento, trazabilidad, exportacion y pruebas).
	- Se ampl?a `documentos/punto_13_validacion_integral_resultado.md` con evidencia complementaria de cobertura y UAT por rol.
	- Se actualiza `Pendiente Notas` para marcar completados los 3 pendientes transversales.
	- Validaciones ejecutadas:
		- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\validar_punto_13.ps1` (OK).
		- `go test ./auth ./db ./handlers ./metrics ./utils -cover -count=1` (db `51.4%`, handlers `50.4%`).
		- `go test ./handlers -run "Test(SuperEndpointsPermisosPorRol|EmpresaPermisosContextoHandlerRetornaPermisosPorRol|EmpresaPermisosContextoHandlerIncluyeMatrizRoles|EmpresaCarritosCompraBloqueaMetodoPagoSegunRol|EmpresaCarritosCompraRespetaBloqueoPropinaYComisionPorRol|EmpresaConfiguracionOperativaHandlerConfigAndRole|EmpresaDocumentosGestionHandlerVersionadoYControlAcceso)$" -count=1` (OK).

## 2026-04-07
- Cierre tecnico del modulo 27 (Ventas simples por estacion):
	- Se ampl?a `backend/db/carritos_compras.go` para:
		- agregar tabla `empresa_ventas_estacion_metricas` y funciones de registro/resumen de rendimiento por estacion.
		- calcular duracion de atencion por venta y resolver identidad de estacion desde carrito (`referencia_externa`/`codigo`).
	- Se ampl?a `backend/handlers/carritos_compras.go` para:
		- exponer `GET action=metricas_estacion` en `/api/empresa/carritos_compra`.
		- registrar metricas en `pagar_estacion`, `anular_cierre_parcial` y `recuperar_interrumpido`.
	- Se actualiza frontend de ventas simples:
		- `web/administrar_empresa/ventas_simple.html` incorpora panel de sincronizacion offline, metricas de estacion y correccion rapida post-cobro.
		- `web/js/ventas_simple.js` (nuevo) implementa cola offline por estacion con checksum SHA-256 y sincronizacion segura al reconectar.
		- `web/estilos.css` agrega estilos de estado de sincronizacion (`en linea`, `offline`, `sincronizando`).
	- Se ampl?a `backend/handlers/auth_users_carritos_test.go` con `TestEmpresaCarritosCompraMetricasEstacionIncluyeCorrecciones`.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "TestEmpresaCarritosCompraMetricasEstacionIncluyeCorrecciones|TestEmpresaCarritosCompraAndItemsFlow" -count=1` (OK).
		- `go test ./... -run "^$" -count=1` (compilacion global backend OK).

## 2026-04-07
- Cierre tecnico del modulo 26 (Carritos de compra e items):
	- Se ampl?a `backend/db/carritos_compras.go` para:
		- agregar reintentos transaccionales en operaciones de items frente a bloqueos motor legado retirado (`database is locked/busy`) para fortalecer concurrencia multiestacion.
		- incorporar `RecoverInterruptedCarritoSession` para recuperar carritos interrumpidos sin perdida de items.
		- incorporar `CancelCarritoPartialClosure` para anulacion parcial de cierre en ventas pagadas con validacion estricta de monto.
	- Se ampl?a `backend/handlers/carritos_compras.go` para:
		- exponer `PUT action=recuperar_interrumpido` con trazabilidad en eventos contables y auditoria empresarial.
		- exponer `PUT action=anular_cierre_parcial` con validacion de negocio y auditoria por `empresa_id` y carrito.
	- Se ajusta `web/administrar_empresa/carrito_de_compras.html` para recuperar sesiones interrumpidas sin reset de items y reservar `reset_items=1` solo para sesiones ya pagadas.
	- Se ampl?a cobertura en:
		- `backend/db/carritos_inventario_test.go` (concurrencia de producto, recuperacion interrumpida, anulacion parcial).
		- `backend/handlers/auth_users_carritos_test.go` (recuperacion con auditoria, reglas de pago mixto y anulacion parcial de cierre).
	- Validaciones ejecutadas:
		- `runTests` sobre `backend/db/carritos_inventario_test.go` y `backend/handlers/auth_users_carritos_test.go` -> 36 pruebas OK, 0 fallidas.
		- `go test ./... -run "^$" -count=1` (compilacion global backend OK).

## 2026-04-07
- Cierre tecnico del modulo 25 (Panel ERP extendido):
	- Se ampl?a `web/js/modulos_erp_extendido.js` para incorporar:
		- formulario guiado dinamico por modulo (sin dependencia obligatoria de JSON libre),
		- validaciones dinamicas por campo y reglas cruzadas (requeridos, tipos, fechas, rangos y consistencia de montos),
		- acciones rapidas parametrizadas por modulo,
		- guia operativa por dominio con flujo recomendado y controles clave.
	- Se ajusta `web/estilos.css` para reforzar UX del panel ERP:
		- grilla guiada responsive,
		- resaltado de errores en linea,
		- panel visual de validaciones,
		- tarjetas de guia operativa.
	- Validaciones ejecutadas:
		- `go test ./... -run "^$" -count=1` (compilacion global backend OK).
		- validacion manual de flujo frontend en `administrar_empresa/modulos_erp_dominio.html` (guiado, validaciones, acciones rapidas y sincronizacion a JSON avanzado).

## 2026-04-07
- Cierre tecnico del modulo 24 (Documental e Integraciones):
	- Se ampl?a `backend/handlers/modulos_faltantes.go` para:
		- reemplazar la ruta generica de documentos por handlers especializados (`EmpresaDocumentosGestionHandler`, `EmpresaDocumentosFirmasHandler`).
		- incorporar versionado documental (`action=versionar`, `action=versiones`) y repositorio con control de acceso por rol/modulo (`action=acceso`, `action=repositorio`).
		- incorporar endurecimiento de integraciones con `action=rotar_credencial` (referencias seguras) y `action=monitoreo`/`action=alertas` (salud de conectores y SLA operativo).
	- Se ampl?a `backend/handlers/empresa_permisos.go` para clasificar `sync_manual`, `rotar_credencial` y `versionar` como acciones criticas de aprobacion en seguridad.
	- Se ampl?a cobertura de pruebas en `backend/handlers/modulos_faltantes_test.go` con:
		- `TestEmpresaIntegracionesAPIsHandlerRotarCredencialYMonitoreo`.
		- `TestEmpresaIntegracionesBancosHandlerRotarCredencial`.
		- `TestEmpresaDocumentosGestionHandlerVersionadoYControlAcceso`.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "TestEmpresa(IntegracionesAPIsHandlerRotarCredencialYMonitoreo|IntegracionesBancosHandlerRotarCredencial|DocumentosGestionHandlerVersionadoYControlAcceso)" -count=1` (OK).
		- `go test ./handlers -count=1` (OK).
		- `go test ./... -count=1` (OK).

## 2026-04-07
- Cierre tecnico del modulo 23 (CRM/Produccion/Logistica):
	- Se ampl?a `backend/handlers/modulos_faltantes.go` para incorporar handlers especializados:
		- `EmpresaProduccionOrdenesHandler` con `action=plan_capacidad` (meta diaria, desviaciones y alertas por atraso/sobrecapacidad).
		- `EmpresaLogisticaEnviosHandler` con `action=seguimiento_hitos` (hitos programacion/salida/entrega, SLA y alertas de incumplimiento).
	- Se extiende `backend/handlers/reportes.go` en `operativo_cadena_cumplimiento` con metas y desviaciones por dominio:
		- `meta_cumplimiento_pct`, `desviacion_meta_pct`, `estado_meta`.
		- resumen global `meta_global_pct` y `desviacion_meta_global_pct`.
	- Se ampl?a cobertura de pruebas en:
		- `backend/handlers/modulos_faltantes_test.go` (`TestEmpresaProduccionOrdenesPlanCapacidad`, `TestEmpresaLogisticaEnviosSeguimientoHitos`).
		- `backend/handlers/reportes_test.go` (validaciones de metas/desviaciones en cadena).
	- Validaciones ejecutadas:
		- `go test ./handlers -run "TestEmpresaProduccionOrdenesPlanCapacidad|TestEmpresaLogisticaEnviosSeguimientoHitos|TestEmpresaReportesHandlerDatasetOperativoCadenaCumplimiento" -count=1` (OK).
		- `go test ./handlers -count=1` (OK).
		- `go test ./... -count=1` (OK).

## 2026-04-07
- Cierre tecnico del modulo 22 (RRHH extendido: vacaciones/licencias):
	- Se ampl?a `backend/db/modulos_faltantes.go` con nuevos campos de RRHH en `empresa_rrhh_vacaciones_licencias` para:
		- aprobacion jerarquica (`nivel_aprobacion_actual`, `nivel_aprobacion_requerido`, `aprobadores_json`, `historial_aprobaciones_json`, `fecha_aprobacion_final`),
		- acumulado y saldo (`periodo_acumulado_*`, `saldo_dias_*`, `saldo_snapshot_json`),
		- enlace a nomina (`empleado_nomina_id`, `nomina_liquidacion_id`, `nomina_periodo_*`, `nomina_vinculada_*`).
	- Se ampl?a `backend/handlers/modulos_faltantes.go` con handler especializado `EmpresaRRHHVacacionesLicenciasHandler` y acciones:
		- `action=resumen_saldo` para acumulado/saldo de vacaciones,
		- `action=solicitar_aprobacion`, `action=aprobar`, `action=rechazar` para flujo jerarquico,
		- `action=vincular_nomina` para enlazar novedades aprobadas a liquidacion/periodo de nomina.
	- Se actualiza `backend/handlers/empresa_permisos.go` para mapear acciones RRHH criticas a permisos de aprobacion/actualizacion.
	- Se ampl?a `backend/handlers/modulos_faltantes_test.go` con pruebas de:
		- saldo y aprobacion jerarquica multinivel,
		- vinculacion de novedades RRHH a nomina por periodo.
	- Validaciones ejecutadas:
		- `go test ./handlers -run RRHH -count=1` (OK).
		- `go test ./handlers -count=1` (OK).
		- `go test ./... -count=1` (OK).

## 2026-04-07
- Cierre tecnico del modulo 21 (Inventario extendido: lotes/series y devolucion proveedor):
	- Se ampl?a `backend/db/modulos_faltantes.go` con:
		- trazabilidad completa de lotes/series mediante tabla `inventario_lotes_series_movimientos`.
		- campos operativos de bloqueo/estado en `inventario_lotes_series` (`reservado_cantidad`, `vendido_cantidad`, `bloqueado_venta`, `bloqueo_motivo`, `ultima_operacion_*`).
		- campos contables de devolucion en `empresa_devoluciones_proveedor` (`periodo_contable`, `impacto_contable_*`, `fecha_contabilizacion`, `contabilizado_por`, `total_reintegrado`).
	- Se ampl?a `backend/handlers/modulos_faltantes.go` con handlers especializados:
		- `EmpresaInventarioLotesSeriesHandler` con acciones `trazabilidad`, `validar_disponibilidad`, `reservar`, `vender`, `liberar_reserva`, `ajuste_entrada`, `ajuste_salida`, `devolucion_proveedor`.
		- bloqueo automatico por vencimiento en venta/reserva y actualizacion de estado de lote.
		- `EmpresaComprasDevolucionesProveedorHandler` con `action=contabilizar`/`action=impacto_contable` para generar movimiento financiero, evento contable y actualizar la devolucion a `contabilizada`.
	- Se ampl?a `backend/db/eventos_contables.go` para soportar `devolucion_proveedor_contabilizada` en contrato y asiento contable (flujo de ingreso).
	- Se ampl?a `backend/handlers/modulos_faltantes_test.go` con pruebas de:
		- bloqueo automatico de lote vencido en reserva,
		- trazabilidad de ciclo reserva/venta/liberacion,
		- contabilizacion completa de devolucion proveedor con impacto contable.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "TestEmpresaInventarioLotesSeriesBloqueoAutomaticoVencido|TestEmpresaInventarioLotesSeriesTrazabilidadCicloVenta|TestEmpresaComprasDevolucionesProveedorContabilizarImpactoCompleto" -count=1` (OK).
		- `go test ./handlers -count=1` (OK).
		- `go test ./... -count=1` (OK).

## 2026-04-06
- Cierre tecnico del modulo 20 (Contabilidad operativa extendida: plan de cuentas, CxC y CxP):
	- Se ampl?a `backend/handlers/modulos_faltantes.go` con handlers especializados de finanzas:
		- `EmpresaFinanzasPlanCuentasHandler` con `action=plantillas`, `action=aplicar_plantilla` y `action=validar_cierre_periodo`.
		- `EmpresaFinanzasCuentasCobrarHandler` y `EmpresaFinanzasCuentasPagarHandler` con `action=conciliar_pagos` y validacion de periodo cerrado.
	- Se ampl?a `backend/db/modulos_faltantes.go` con:
		- nuevos metadatos de plantilla en `empresa_plan_cuentas`.
		- campos de conciliacion en `empresa_cuentas_por_cobrar` y `empresa_cuentas_por_pagar`.
		- bloqueo retroactivo por periodo contable cerrado en crear/editar/cambiar estado/eliminar de CxC/CxP.
	- Se ampl?a `backend/handlers/modulos_faltantes_test.go` con pruebas de:
		- plantillas y aplicacion de plan de cuentas por tipo de empresa.
		- conciliacion automatica CxC contra pagos reales.
		- bloqueo de operaciones CxP cuando el periodo esta cerrado.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "TestEmpresaFinanzasPlanCuentasPlantillasYAplicacion|TestEmpresaFinanzasCuentasCobrarConciliacionPagosReales|TestEmpresaFinanzasCarteraBloqueoPeriodoCerrado" -count=1` (OK).
		- `go test ./handlers -count=1` (OK).
		- `go test ./... -count=1` (OK).
- Cierre tecnico del modulo 19 (Gestion comercial extendida: cotizaciones/pedidos/devoluciones):
	- Se ampl?a `backend/handlers/modulos_faltantes.go` con automatizacion comercial en ventas:
		- `POST/PUT action=convertir_pedido` en cotizaciones para convertir cotizacion aprobada/emitida a pedido trazable (`cotizacion_id`, `convertido_pedido_id`).
		- `POST/PUT action=convertir_documento_final` en cotizaciones y pedidos para generar documento final en `empresa_facturacion_documentos`.
		- `GET action=embudo` en cotizaciones para monitoreo operativo con SLA y alertas de vencimiento.
	- Se incorpora snapshot de embudo comercial cotizacionÃƒÂ¢Ã¢â‚¬Â ?pedidoÃƒÂ¢Ã¢â‚¬Â ?documento final con trazabilidad por `empresa_id`.
	- Se agrega dataset exportable `operativo_ventas_embudo_conversion` en `backend/handlers/reportes.go` con formatos `json/csv/txt/xls/pdf`.
	- Se actualiza `backend/handlers/empresa_permisos.go` para clasificar `convertir_pedido` y `convertir_documento_final` como acciones de aprobacion en ventas.
	- Se agregan pruebas en `backend/handlers/modulos_faltantes_test.go` y `backend/handlers/reportes_test.go` para:
		- conversion cotizacionÃƒÂ¢Ã¢â‚¬Â ?pedidoÃƒÂ¢Ã¢â‚¬Â ?documento final,
		- alertas SLA del embudo,
		- dataset/export CSV del nuevo reporte de conversion.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "TestEmpresaVentasCotizacionesConversionPedidoYDocumentoFinal|TestEmpresaVentasCotizacionesEmbudoYAlertasSLA|TestEmpresaReportesHandlerDatasetOperativoVentasEmbudoConversion" -count=1` (OK).
		- `go test ./handlers -run "DIAN|ModulosFaltantes|OperativoCadenaCumplimiento|OperativoVentasEmbudoConversion" -count=1` (OK).
		- `go test ./handlers -count=1` (OK).
		- `go test ./... -count=1` (OK).
- Cierre tecnico del modulo 18 (Facturacion electronica DIAN Colombia):
	- Se ampl?a `backend/handlers/modulos_faltantes.go` en `EmpresaDIANColombiaHandler` con acciones operativas reales:
		- `action=firmar_xml_real` (firma RSA-SHA256).
		- `action=enviar_documento_real` (envio productivo/habilitacion por `url_dian`).
		- `action=consultar_acuse_real` (consulta de acuse y normalizacion de estado).
		- `action=reconexion_dian` (sondeo de conectividad y salida de contingencia).
	- Se implementa gestion segura de credenciales/certificados por referencia:
		- `token_emisor_ref` y `certificado_clave_ref` soportan `env:`, `file:` y `base64:`.
	- Se integra transicion de estado DIAN (`pendiente/enviado/aceptado/rechazado/contingencia/reconectado`) con trazabilidad en `observaciones` y `ultimo_envio`.
	- Se ajusta `backend/handlers/empresa_permisos.go` para clasificar las nuevas acciones DIAN de escritura como `permActionApprove`.
	- Se agregan pruebas en `backend/handlers/modulos_faltantes_test.go` para:
		- flujo firma + envio + acuse exitoso,
		- contingencia por falla de transporte y recuperacion por reconexion.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "DIAN|ModulosFaltantes" -count=1 -v` (OK).
		- `go test ./handlers -count=1` (OK).
		- `go test ./... -count=1` (suite backend completa OK).
- Estabilizacion del panel de graficos y estadisticas (compras):
	- Se ajusta `backend/handlers/graficos_estadisticas.go` para soportar la nueva estructura del dataset `operativo_compras_movimientos` (agregado por proveedor) sin perder compatibilidad con la forma anterior.
	- Se agrega fallback para construir la serie de compras desde movimientos financieros (`egresos` de compras) cuando no existen documentos en `empresa_compras_documentos`.
	- Validaciones ejecutadas:
		- `go test ./handlers -run TestEmpresaGraficosEstadisticasHandlerPanelYAcciones -count=1 -v` (OK tras ajuste).
		- `runTests` en `backend/handlers/graficos_estadisticas_test.go` (2 pruebas OK).
		- `go test ./handlers -run GraficosEstadisticas -count=1` (OK).
		- `go test ./... -count=1` (suite backend completa OK).
- Cierre tecnico del modulo 17 (Facturacion electronica):
	- Se integra despacho fiscal por pais/proveedor en `backend/handlers/facturacion_electronica.go`:
		- proveedor `manual` (productivo local), `mock://` para pruebas y despacho HTTP contra `api_base_url` cuando aplica.
		- respuesta operativa `integracion_fiscal` en acciones transaccionales (`emitir`, `anular`, `nota_credito`).
	- Se implementa cola de reintentos FE en `backend/db/facturacion_electronica.go`:
		- nueva tabla `facturacion_electronica_reintentos` con estado de envio, intentos, proximo intento, contingencia y referencia externa.
		- nuevas operaciones de consulta/actualizacion y listados filtrados por estado.
	- Se agregan endpoints operativos FE:
		- `GET action=reintentos`.
		- `POST/PUT action=procesar_reintentos`.
		- `GET action=reconciliacion` y `POST/PUT action=reconciliar_estados`.
	- Se activa contingencia automatica al superar `max_intentos` y se conserva numeracion legal por resolucion en emision.
	- Se actualiza contrato contable del modulo `facturacion` con eventos de integracion (`factura_integracion_enviada`, `factura_integracion_fallida`, `factura_contingencia_activada`).
	- Validaciones ejecutadas:
		- `go test ./db -run Facturacion -count=1` (OK).
		- `go test ./handlers -run FacturacionElectronica -count=1` (OK).
		- `go test ./handlers -run Facturacion -count=1` (OK).
		- `go test ./... -count=1` (falla no relacionada en `TestEmpresaGraficosEstadisticasHandlerPanelYAcciones`).
- Cierre tecnico del modulo 16 (Compras):
	- Se ampl?a el ciclo documental de compras con aprobacion multinivel:
		- `requiere_aprobacion`, `niveles_aprobacion_requeridos`, `nivel_aprobacion_actual`, `aprobadores_json`.
		- Nuevas acciones: `solicitar_aprobacion`, `aprobar_compra`, `rechazar_compra`.
	- Se cierra recepcion parcial avanzada por item:
		- `recepcion_detalle_json` y `recepcion_resumen_json` para registrar cantidades solicitadas/recibidas, pendientes y diferencias.
		- Nueva accion: `recepcionar_parcial_compra`, consolidada con `recepcionar_compra` al completar recepcion total.
	- Se integra validacion documental proveedor-factura-entrada:
		- `validacion_documental_estado`, `proveedor_documento_ref`, `factura_documento_ref`, `entrada_documento_ref`.
		- Nueva accion: `validar_documentos` con verificacion de proveedor y referencias documentales.
	- Se ampl?a UI en `web/administrar_empresa/compras.html` con campos, filtros/KPI y acciones operativas del nuevo flujo.
	- Validaciones ejecutadas:
		- `runTests` en `backend/db/documentos_transaccionales_test.go`, `backend/handlers/compras_documentos_test.go`, `backend/handlers/empresa_permisos_test.go` (21 pruebas OK).
		- `go test ./... -run TestDoesNotExist -count=1` (compilacion global backend OK).
- Hotfix de compatibilidad de migraciones legacy en startup:
	- Se corrige el orden de migracion en `EnsureEmpresaPropinasSchema` para crear indices despues de asegurar columnas faltantes (`cierre_caja_id` y relacionadas), evitando fallos en bases antiguas.
	- Se corrige el orden de migracion en `EnsureEmpresaComisionesServicioSchema` para crear indices despues de asegurar columnas faltantes (`ajuste_manual` y relacionadas), evitando fallos en bases antiguas.
	- Resultado operativo: el script `scripts/iniciar_servidor.ps1` vuelve a iniciar correctamente y el backend queda escuchando en `:8080`.
	- Validaciones ejecutadas:
		- `go test ./db -run "Propina|Comision" -count=1` (OK).
		- `go test ./handlers -run "Propina|Comision" -count=1` (OK).
		- `go test ./... -run TestDoesNotExist -count=1` (compilacion global backend OK).
- Cierre tecnico del modulo 15 (Comisiones por servicio):
	- Se ampl?a el modelo de comisiones con escalas por rol/servicio y tope por item:
		- nueva tabla `empresa_comisiones_servicio_escalas` (`rol_operacion`, `servicio_filtro`, `porcentaje_comision`, `tope_comision`, `prioridad`).
	- Se ampl?a `empresa_comisiones_servicio_movimientos` con trazabilidad operativa:
		- `rol_operacion`, `escala_id`, `monto_comision_bruto`, `tope_comision_aplicado`,
		- `origen_movimiento`, `ajuste_manual`, `referencia_ajuste`, `ajuste_estado`, `aprobado_por`, `aprobado_en`,
		- `liquidacion_nomina_id`, `periodo_liquidacion_desde`, `periodo_liquidacion_hasta`, `liquidado_en`, `liquidado_por`.
	- Se incorporan endpoints/acciones de comisiones para operacion avanzada:
		- escalas (`escalas`, `escala`, `activar_escala`, `desactivar_escala`),
		- ajuste manual y aprobacion (`ajuste_manual`, `aprobar_ajuste`, `rechazar_ajuste`),
		- resumen para nomina (`resumen_liquidacion`).
	- Se integra nomina con comisiones:
		- `empresa_nomina_liquidaciones` incorpora `comisiones_servicio_total`, `comisiones_servicio_movimientos`, `comisiones_servicio_ajustes`.
		- el calculo de liquidacion integra comisiones y enlaza movimientos al periodo liquidado.
	- Se amplia `web/administrar_empresa/comisiones.html` para operacion completa del modulo 15:
		- gestion de escalas/topes,
		- registro de ajuste manual,
		- aprobacion/rechazo de ajustes pendientes,
		- filtros avanzados de reporte y consulta de `resumen_liquidacion`.
	- Validaciones ejecutadas:
		- `go test ./db -run "TestEmpresaComisionesServicio|TestEmpresaNominaLiquidacionIntegraComisionesServicio" -count=1` (OK).
		- `go test ./handlers -run "TestEmpresaComisionesServicioHandler" -count=1` (OK).
		- `go test ./... -run TestDoesNotExist -count=1` (compilacion global backend OK).
- Avance y cierre tecnico del modulo 14 (Propinas):
	- Se ampl?a la configuracion empresarial de propinas con reglas fiscales:
		- `pais_fiscal`, `regimen_fiscal`, `tratamiento_fiscal` (`no_gravada`/`gravada`) y `porcentaje_impuesto_propina`.
	- Se ampl?a el libro de movimientos de propinas con:
		- `origen_movimiento` (`venta`/`ajuste_manual`),
		- `ajuste_manual`, `referencia_ajuste`, `cierre_caja_id`, `conciliado_en`,
		- snapshot fiscal por movimiento (`fiscal_*`).
	- Se incorpora conciliacion de propinas contra cierre de caja:
		- accion manual `action=conciliacion_cierre` en propinas,
		- integracion automatica durante transiciones `cerrar/aprobar` de cierre de caja,
		- persistencia de resumen en `empresa_cierres_caja` (`propinas_movimientos`, `propinas_total`, `propinas_ajustes`, `propinas_impuesto`, `propinas_neto`, `propinas_conciliado_*`).
	- Se incorpora ajuste manual auditado de propinas:
		- accion `action=ajuste_manual`,
		- registro no bloqueante en `empresa_auditoria_eventos`.
	- Se actualiza frontend `web/administrar_empresa/propinas.html` con:
		- configuracion fiscal,
		- formulario de ajuste manual,
		- accion de conciliacion por cierre,
		- filtros y columnas extendidas en el reporte.
	- Se agrega cobertura de pruebas para flujo de ajuste y conciliacion:
		- `backend/handlers/propinas_test.go`.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "Propinas|Cierre" -count=1` (OK).
		- `go test ./db -run "Propina|CierreCaja|Finanzas" -count=1` (OK).
		- `go test ./... -run "^$" -count=1` (compilacion global backend OK).
- Avance del modulo 13 (Codigos de descuento avanzados):
	- Se ampl?a `codigos_de_descuento` con reglas contextuales:
		- `segmento_cliente`, `canal_venta`, `horario_desde`, `horario_hasta`, `dias_semana`.
	- Se incorpora antifraude por cliente:
		- `max_usos_por_cliente`, `ventana_horas_fraude`.
	- Se agrega trazabilidad de redenciones en nueva tabla `codigos_descuento_redenciones` con estados:
		- `aplicada`, `revertida`, `anulada`.
	- Se integra ciclo de redencion con carritos:
		- aplica al cerrar carrito,
		- revierte al reabrir,
		- anula al eliminar carrito.
	- Se extiende API de codigos:
		- validacion contextual (`action=validar` con `carrito_id`, `cliente_id`, `canal_venta`),
		- consulta de trazabilidad (`action=redenciones`).
	- Se actualiza `web/administrar_empresa/codigos_de_descuento.html` para administrar reglas avanzadas y antifraude.
	- Validaciones ejecutadas:
		- `go test ./db -run "TestCodigoDescuento" -count=1` (OK).
		- `go test ./handlers -run "TestNoExiste" -count=1` (OK, compilacion handlers).
		- `go test ./db -run "TestCarritoProductoDescuentaInventarioYVentaMantieneStock|TestCarritoStockNoSeDuplicaAlReactivarSesionCerrada" -count=1` (OK).
		- `go test ./...` (falla en prueba no relacionada: `TestEmpresaGraficosEstadisticasHandlerPanelYAcciones`).
- Validacion final de continuidad tecnica y documental (post-ediciones recientes):
	- Se revalida compilacion global de backend tras ajustes recientes en inventario/combos.
	- Resultado: `go test ./... -run TestDoesNotExist -count=1` en `backend` -> OK (sin errores de compilacion).
	- Se confirma sincronizacion de cierre de modulos 1-12 en documentacion operativa y tecnica.
- Cierre del modulo 12 (Combos de productos):
	- Se implementa versionado de receta por combo:
		- nuevas columnas en `combos_productos`: `receta_version`, `costo_teorico`, `costo_real`, `variacion_costo`, `variacion_costo_porcentaje`.
		- nueva tabla `combos_productos_versiones` para snapshots historicos de ingredientes por version.
	- Se incorpora validacion de costo teorico vs costo real de ingredientes en create/update de combos:
		- bloqueo si la variacion porcentual supera el umbral operativo.
		- bloqueo si el precio del combo no cubre el costo real calculado.
	- Se endurece concurrencia de inventario en carritos:
		- reserva de stock con `UPDATE` atomico condicionado (`cantidad >= requerida`) para evitar sobreventa en ventas simultaneas.
	- Se actualiza frontend `web/administrar_empresa/combos_productos.html` para mostrar version de receta y metricas de costo.
	- Validaciones ejecutadas:
		- `runTests` sobre `backend/db/productos_categorias_test.go` y `backend/db/carritos_inventario_test.go`.
		- `go test ./... -run TestDoesNotExist -count=1`.
- Verificacion final de continuidad del modulo 11 (Inventario):
	- Se ejecuta compilacion global posterior a cambios recientes en archivos de inventario con `go test ./... -run TestDoesNotExist -count=1` (OK).
	- Se confirma cierre operativo completo del checklist de modulo 11 (schema, costos, conteo ciclico, alertas proactivas y documentacion).
- Cierre del modulo 11 (Inventario) de Fase 3:
	- Se implementa configuracion de politica de costo por empresa:
		- `GET/PUT /api/empresa/inventario/configuracion`.
		- Politicas soportadas: `promedio` y `peps`.
	- Se incorpora soporte de capas/lotes de costo para trazabilidad de salidas y transferencias:
		- tabla `inventario_costos_lotes`.
		- salida con PEPS por capas y recalculo de costo promedio por bodega/producto.
	- Se implementa conteo ciclico con ajuste auditado:
		- `GET/POST /api/empresa/inventario/conteo_ciclico`.
		- tabla `inventario_conteos_ciclicos` y movimiento automatico `ajuste_positivo/ajuste_negativo` cuando hay variacion.
	- Se cierran alertas operativas proactivas de inventario:
		- `GET /api/empresa/inventario/alertasaction=proactivas`.
		- incorpora `sobrestock`, `deficit`, `exceso` y `accion_sugerida`.
	- Se actualiza frontend `web/administrar_empresa/administrar_productos.html` con:
		- selector/guardado de politica de costo,
		- formulario y tabla de conteo ciclico,
		- visualizacion de alertas proactivas (quiebre/sobrestock).
	- Validaciones ejecutadas:
		- `go test ./db -run "TestInventarioPoliticaCostoPromedioYPEPS|TestRegistrarConteoCiclicoInventarioAjustaYAudita|TestGetAlertasOperativasByEmpresaIncluyeSobrestock" -count=1`.
		- `go test ./handlers -run "TestEmpresaInventarioConfiguracionYConteoCiclicoHandler|TestEmpresaInventarioAlertasHandlerProactivasIncluyeSobrestock" -count=1`.
		- `go test ./... -run TestDoesNotExist -count=1`.
- Cierre del modulo 10 (Clientes) de Fase 3:
	- Se implementa deduplicacion por `documento`, `correo` y `telefono` en `create/update` de clientes por `empresa_id`.
	- El endpoint `POST/PUT /api/empresa/clientes` responde `409` cuando detecta conflicto de deduplicacion, con mensaje de campo duplicado.
	- Se agrega dataset operativo para exportacion masiva comercial:
		- `operativo_clientes_segmentacion_comercial` en `/api/empresa/reportes`.
		- Incluye segmento, metricas de compra y `accion_comercial_sugerida` por cliente.
		- Exportacion disponible en `pdf/xls/csv/json/txt`.
	- Se actualiza frontend `web/administrar_empresa/administrar_clientes.html` con panel de exportacion masiva por segmento.
	- Validaciones ejecutadas:
		- `go test ./db -run "Test(CreateClienteDeduplicacionDocumentoCorreoTelefono|UpdateClienteDeduplicacionCorreoTelefono|GetClientePerfilComercialByEmpresaAndHistorial|GetClientePerfilComercialByEmpresaSinComprasSegmentoNuevo|GetClienteByID)$" -count=1`.
		- `go test ./handlers -run "Test(EmpresaClientesHandlerPerfilHistorialSegmentacion|EmpresaClientesHandlerConflictosDeduplicacion|EmpresaReportesHandlerDatasetOperativoClientesSegmentacionComercial)$" -count=1`.
		- `go test ./... -run "^$" -count=1`.
- Cierre del modulo 9 (Tarifas por dia) de Fase 3:
	- Se implementa prorrateo de tarifa diaria por ventana de `check-in/check-out` para entrada/salida fuera de ventana.
	- Se extiende simulador de `GET /api/empresa/tarifas_por_diaaction=calcular` con detalle de:
		- `dias_completos`, `dias_equivalentes`,
		- `monto_dias_completos`, `monto_prorrateo_(entrada|intermedio|salida)`,
		- `minutos_prorrateo_fuera_ventana`.
	- Se agrega aplicacion masiva de una misma tarifa diaria a todas las estaciones detectadas:
		- `PUT /api/empresa/tarifas_por_diaaction=aplicar_todas_estaciones`.
	- Se agrega reporte operativo comparativo por estacion:
		- dataset `operativo_tarifas_comparativo_estaciones` en `/api/empresa/reportes`,
		- comparativo de ingreso esperado (motor prorrateado) vs ingreso real cobrado,
		- exportacion en `pdf/xls/csv/json/txt`.
	- Se actualiza frontend `web/administrar_empresa/tarifas_por_dia.html` con:
		- boton `Aplicar a todas las estaciones`,
		- simulador con desglose de prorrateo,
		- panel de descarga del comparativo esperado vs real.
	- Validaciones ejecutadas:
		- `go test ./db -run "TarifaPorDia|ApplyEmpresaTarifaPorDiaToAllStations|EmpresaTarifasPorDia"`.
		- `go test ./handlers -run "TarifasPorDia|CarritosCompraListIncluyeTarifaPorDiaAutomatica|OperativoTarifasIngresos|OperativoTarifasComparativoEstaciones"`.
		- `go test ./... -run "^$"`.
- Cierre del modulo 8 (Tarifas por minutos) de Fase 3:
	- Se agrega configuracion empresarial avanzada de calculo:
		- `redondeo_modo` (`ninguno`, `arriba`, `abajo`, `matematico`),
		- `redondeo_unidad`,
		- `monto_minimo_diario`,
		- `monto_maximo_diario`.
	- Se extiende simulador de cobro por minutos con detalle de:
		- monto base, monto extra, subtotal, monto redondeado y ajuste,
		- aplicacion de minimo/maximo diario,
		- soporte de minutos fraccionarios (`minutos_consumidos` decimal).
	- Se cierra trazabilidad contable del calculo por minutos:
		- registro de evento `finanzas.tarifa_por_minutos_calculada` en `empresa_eventos_contables`,
		- respuesta de simulacion con `trazabilidad_contable_id`, `documento_codigo` y `periodo_contable`.
	- Se agrega aplicacion masiva de una misma regla de tarifa a todas las estaciones detectadas:
		- `PUT /api/empresa/tarifas_por_minutosaction=aplicar_todas_estaciones`.
	- Se actualiza frontend `web/administrar_empresa/tarifas_por_minutos.html` con:
		- panel de configuracion avanzada de redondeo y topes,
		- boton `Aplicar a todas las estaciones`,
		- simulador con detalle de calculo y referencia contable.
	- Validaciones ejecutadas:
		- `go test ./db -run "TestEmpresaTarifasPorMinutos|TestApplyEmpresaTarifaPorMinutosToAllStations|TestRegisterTarifaPorMinutosCalculoContable|TestEmpresaEventosContables" -count=1`.
		- `go test ./handlers -run "TestEmpresaTarifasPorMinutosHandler" -count=1`.
		- `go test ./... -run "^$" -count=1`.
- Cierre del modulo 7 (Reservas por estacion/habitacion) de Fase 3:
	- Se refuerza control de concurrencia anti-overbooking por estacion en ventanas solapadas:
		- validacion de conflicto por `estacion_id` y `carrito_id` asociado,
		- bloqueo para estados operativos `pendiente_pago`, `confirmada` y `en_curso`.
	- Se implementa politica automatica avanzada de reservas:
		- expiracion de pendientes por `fecha_expiracion` y fallback por antiguedad de creacion,
		- marcacion automatica de `no_show` sobre reservas confirmadas fuera de tolerancia operativa,
		- accion de sincronizacion: `GET /api/empresa/reservas_hotelaction=aplicar_politicas`.
	- Se incorpora reconversion operativa de reserva a carrito:
		- `PUT /api/empresa/reservas_hotelaction=convertir_carrito`.
		- transicion de reserva a estado `en_curso` y activacion de carrito asociado.
	- Se actualiza frontend `web/administrar_empresa/reservas_hotel.html` con:
		- accion `Aplicar politicas`,
		- accion `Reconver. carrito`,
		- filtros extendidos para estados `en_curso` y `no_show`.
	- Validaciones ejecutadas:
		- `go test ./db -run "TestReservaHotel(FlowCRUDAndDisponibilidad|MultiEstacionNoOverbookingYReconversion|PoliticaNoShowYExpiracionAvanzada)$" -count=1`.
		- `go test ./handlers -run "TestEmpresaReservasHotelHandler(CRUDAndDisponibilidad|PoliticasYReconversion)$" -count=1`.
		- `go test ./... -run "^$" -count=1`.
- Cierre del modulo 6 (Registro de vehiculos) de Fase 3:
	- Se agrega configuracion de validacion de placa/patente por empresa y pais:
		- `GET/PUT /api/empresa/vehiculos_registroaction=config`.
		- Tabla `empresa_vehiculos_configuracion` con `pais_codigo`, `patente_regex`, `patente_descripcion`, `evitar_duplicado_activo`.
	- Se implementa bloqueo de duplicidad activa por patente canonica en patio/empresa:
		- validado en crear, editar y activar registros de vehiculos.
		- respuesta HTTP `409` ante conflicto de duplicidad activa.
	- Se agrega reporte operativo de permanencia y tiempos de estancia:
		- `GET /api/empresa/vehiculos_registroaction=permanencia`.
		- dataset `operativo_vehiculos_permanencia` en `/api/empresa/reportes` con exportacion `pdf/xls/csv/json/txt`.
	- Se integra frontend en `web/administrar_empresa/vehiculos_registro.html`:
		- panel de configuracion de formato de placa por pais,
		- consulta visual de permanencia,
		- exportacion de reporte en formatos estandar.
	- Validaciones ejecutadas:
		- `go test ./db -run TestEmpresaVehiculoRegistroConfigValidacionDuplicidadYPermanencia -count=1`.
		- `go test ./handlers -run TestEmpresaVehiculosRegistroHandlerConfigYReportePermanencia -count=1`.
		- `go test ./handlers -run TestEmpresaReportesHandlerDatasetOperativoVehiculosPermanencia -count=1`.
		- `go test ./... -run "^$" -count=1`.
- Cierre del modulo 5 (Nomina de sueldos) de Fase 3:
	- Se agregan operaciones nuevas en nomina:
		- `GET /api/empresa/nominaaction=desprendible&empleado_nomina_id={id}&periodo_desde=YYYY-MM-DD&periodo_hasta=YYYY-MM-DD`.
		- `GET /api/empresa/nominaaction=conciliacion_asistencia` (auditoria sin cambios).
		- `POST /api/empresa/nominaaction=conciliar_asistencia` (auditoria con opcion de auto-recalculo).
	- Se implementa desprendible estandar por empleado y periodo con detalle de horas, devengados, deducciones y neto a pagar.
	- Se implementa conciliacion automatica entre asistencia y liquidacion final:
		- detecta diferencias de registros/horas,
		- identifica asistencias sin liquidacion,
		- permite recalcular/crear liquidaciones inconsistentes cuando `auto_recalcular=true`.
	- Se integra frontend en `web/administrar_empresa/nomina_sueldos.html`:
		- boton de conciliacion con modo auditoria o auto-recalculo,
		- generacion/visualizacion de desprendible por empleado-periodo,
		- accion de desprendible desde tabla de liquidaciones.
	- Se documentan y validan casos de formula por pais/empresa (CO/MX + override por empresa) con pruebas automatizadas.
	- Validaciones ejecutadas:
		- `go test ./db -run "Test(EmpresaNominaGenerateLiquidacionesFromAsistencia|EmpresaNominaCalculoPorPaisYEmpresa|EmpresaNominaDesprendibleYConciliacionAsistencia)$" -count=1`.
		- `go test ./handlers -run "TestEmpresaNominaSueldosHandlerFlow$" -count=1`.
		- `go test ./... -run "^$" -count=1`.
- Cierre del modulo 4 (Asistencia de empleados) de Fase 5:
	- Se implementa cierre de periodo con bloqueo operativo de edicion posterior:
		- `POST /api/empresa/asistencia_empleadosaction=cerrar_periodo`.
		- `GET /api/empresa/asistencia_empleadosaction=periodos_cerrados`.
	- Se agrega configuracion por empresa para tolerancias y reglas de turno:
		- `GET/PUT /api/empresa/asistencia_empleadosaction=config`.
		- `tolerancia_entrada_minutos`, `hora_inicio_turno_(manana|tarde|noche)`, `permitir_turno_nocturno`, `permitir_turno_cruzado`.
	- Se incorporan validaciones de negocio en asistencia:
		- bloqueo de create/update/delete/activar/desactivar/marcar_entrada/marcar_salida cuando la fecha pertenece a periodo cerrado,
		- rechazo de turno nocturno o cruzado cuando la configuracion empresarial lo deshabilita,
		- calculo de tardanza con tolerancia configurable.
	- Se publica reporte operativo de auditoria para nomina:
		- dataset `operativo_asistencia_nomina_auditoria` en `/api/empresa/reportes` con exportacion `pdf/xls/csv/json/txt`.
	- Se integra frontend en `web/administrar_empresa/asistencia_empleados.html`:
		- panel de configuracion,
		- cierre de periodo y listado de cierres,
		- descarga del reporte de auditoria de nomina.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "Test(EmpresaAsistenciaEmpleadosHandlerCRUDFlow|EmpresaAsistenciaEmpleadosHandlerConfigTurnosYTolerancia|EmpresaAsistenciaEmpleadosHandlerCierrePeriodoBloqueaEdicion|EmpresaReportesHandlerDatasetOperativoAsistenciaNominaAuditoria)$" -count=1`.
		- `go test ./... -run "^$" -count=1`.
- Cierre del modulo 3 (Usuarios de empresa) de Fase 1:
	- Se agrega cambio autogestionado de contrase?a para usuario empresa:
		- `POST /api/empresa/usuarios/cambiar_password`.
	- Se implementan politicas de contrase?a configurables desde `configuraciones`:
		- `usuarios.password_min_length`
		- `usuarios.password_require_uppercase`
		- `usuarios.password_require_lowercase`
		- `usuarios.password_require_digit`
		- `usuarios.password_require_symbol`
		- `usuarios.password_rotation_days`.
	- El login de usuario empresa ahora devuelve `password_rotation_required` cuando aplica rotacion obligatoria.
	- Se incorpora captura de notificaciones de confirmacion/restablecimiento en entorno de pruebas de correo:
		- tabla `super_correo_notificaciones_prueba` en `pcs_superadministrador`.
		- activacion por `PCS_MAIL_TEST_MODE=1` o `gmail.smtp_test_mode=1`.
	- Se integra frontend de autogestion en `web/login_usuario.html` y `web/js/login_usuario.js`.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "Test(EmpresaUsuarioChangePasswordFlow|EmpresaUsuarioChangePasswordPolicyRejectsWeakPassword|EmpresaUsuarioLoginRequiresRotationWhenPolicyEnabled|EmpresaUsuarioNotificationsCaptureInMailTestMode|EmpresaUsuarioPasswordRecoveryFlow)" -count=1`.
		- `go test ./... -run "^$" -count=1`.
- Cierre del modulo 2 (Administracion global super) de Fase 1:
	- Se implementa desactivacion/rehabilitacion de empresa con validaciones de impacto operativo y confirmacion forzada cuando existen bloqueos:
		- `GET /super/api/empresasid={id}&action=impacto_desactivacion`.
		- `PUT /super/api/empresasid={id}&action=desactivar[&force=1]`.
		- `PUT /super/api/empresasid={id}&action=activar&activo=1`.
	- Se agrega respaldo/restauracion de configuracion critica super:
		- `GET /super/api/config/backup` (exporta JSON).
		- `PUT /super/api/config/backup` (restaura JSON).
	- Se integra operacion desde frontend:
		- `web/js/seleccionar_empresa.js` para desactivar/reactivar con consulta de impacto.
		- `web/super/configuracion_avanzada.html` con descarga y restauracion de respaldo.
	- Se agregan pruebas de permisos y flujo super en `backend/handlers/system_empresas_handlers_test.go`.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "Test(EmpresasHandlerDesactivarConImpactoYForce|EmpresasHandlerImpactoDesactivacion|SuperConfigBackupHandlerExportYRestore|SuperEndpointsPermisosPorRol)" -count=1`.
		- `go test ./... -run "^$" -count=1`.
- Cierre del modulo 1 (Autenticacion y sesiones) de Fase 1:
	- Se implementa bloqueo temporal por intentos fallidos en login de usuario empresa.
	- Se agrega recuperacion de contrasena para usuario empresa con token temporal:
		- `POST /api/empresa/usuarios/solicitar_recuperacion_password`
		- `POST /api/empresa/usuarios/restablecer_password`
	- Se endurece seguridad de sesion:
		- sesiones nuevas con `fecha_fin` por expiracion (24h),
		- revocacion de token en logout,
		- middleware bloquea tokens expirados o revocados.
	- Se habilita flujo frontend de recuperacion en `web/login_usuario.html` y `web/js/login_usuario.js`.
	- Validaciones ejecutadas:
		- `runTests` sobre `backend/handlers/auth_users_carritos_test.go` -> 24/24.
		- `go test ./... -run "^$" -count=1` (compilacion global OK).
- Cierre tecnico backend de pasarela unica Wompi:
	- Se elimina remanente de Mercado Pago en backend:
		- `backend/handlers/payments_handlers.go`: retiro de handlers/utilidades Mercado Pago.
		- `backend/db/db.go`: retiro de tipo/funciones de persistencia `pagos_mercadopago`.
		- `backend/main.go`: retiro de bootstrap/migracion de `pagos_mercadopago`.
		- `backend/utils/utils.go`: retiro del prefijo `/mercadopago/` en manejo JSON API.
		- `backend/tools/query_users/main.go`: migracion de inspeccion local hacia `wompi.*` y `pagos_wompi`.
	- Se sincroniza documentacion tecnica con el estado real:
		- `documentos/estructura_bd.md` y `estructura_bd.md`.
		- `documentos/diagramas/estructura_del_codigo.md`.
		- `documentos/descripcion_de_archivos`.
	- Validacion tecnica ejecutada: `go test ./... -run "^$" -count=1` (compilacion global OK).
- Cierre de pendientes de modulos:
	- Se valida la matriz de estado de modulos/reportes y no quedan modulos marcados como incompletos (`Pendiente` o `Parcial`) en `documentos/modulos del proyecto.md`.
	- Se actualiza `Pendiente Notas` marcando como completado el pendiente de pasarela unica Wompi.
- Pasarela de pago unificada en Wompi:
	- Se retira la configuraci?n de Mercado Pago de `web/super/configuracion_avanzada.html` y se deja ?nicamente la secci?n de credenciales de Wompi en configuraci?n avanzada del panel super administrador.
	- Se simplifica `web/pagar_licencia.html` eliminando selector/panel/flujo de Mercado Pago para operar solo con Nequi (Wompi) y activaci?n manual interna.
	- Se desregistran rutas de Mercado Pago en `backend/main.go` (`/super/api/config/mercadopago`, `/mercadopago/create_preference`, `/mercadopago/webhook`, `/mercadopago/reconcile`, `/mercadopago/test_preference`).
	- Validaci?n t?cnica ejecutada: `go test ./... -run "^$" -count=1` (compilaci?n global OK).
- Cierre de trazabilidad y validacion final del plan de reportes:
	- Se revalida la presencia de los datasets operativos de cierre (`operativo_propinas_acumulado`, `operativo_comisiones_lavador`, `operativo_facturacion_trazabilidad`, `operativo_auditoria_acciones`) en `backend/handlers/reportes.go`.
	- Se ejecuta validacion completa de `backend/handlers/reportes_test.go` con resultado `16/16` pruebas aprobadas.
	- Se confirma consistencia de estado documental en `documentos/modulos del proyecto.md`, `CHANGELOG.md` y `documentos/historial_de_cambios`.
	- Se deja cerrado el pendiente de trazabilidad del plan secuencial.
- Plan secuencial de cierre de modulos incompletos - bloques 6 a 9 (Propinas, Comisiones, Facturacion y Auditoria):
	- Se agregan en `backend/handlers/reportes.go` cuatro datasets operativos nuevos:
		- `operativo_propinas_acumulado` (acumulado por usuario, distribucion directa/universal y participacion),
		- `operativo_comisiones_lavador` (acumulado por lavador con base de servicios y ticket de comision),
		- `operativo_facturacion_trazabilidad` (emitidas/anuladas/pendientes y trazabilidad legal por tipo documental),
		- `operativo_auditoria_acciones` (eventos por modulo/usuario con errores HTTP y acciones criticas).
	- Se actualiza catalogo y switch de construccion de datasets para incluir estos cuatro reportes en suite/export.
	- Se amplia `backend/handlers/reportes_test.go` con pruebas dedicadas:
		- `TestEmpresaReportesHandlerDatasetOperativoPropinasAcumulado`.
		- `TestEmpresaReportesHandlerDatasetOperativoComisionesLavador`.
		- `TestEmpresaReportesHandlerDatasetOperativoFacturacionTrazabilidad`.
		- `TestEmpresaReportesHandlerDatasetOperativoAuditoriaAcciones`.
	- Se refuerza `ensureEmpresaReportesSchema` con `EnsureEmpresaPropinasSchema`, `EnsureEmpresaComisionesServicioSchema` y `EnsureEmpresaAuditoriaSchema` para cobertura de suite completa.
	- Se actualiza la matriz en `documentos/modulos del proyecto.md` marcando Propinas, Comisiones, Facturacion y Auditoria como activos en reportes.
	- Validacion tecnica ejecutada:
		- `runTests` focalizado en 4 pruebas nuevas (ok).
		- `runTests` completo sobre `backend/handlers/reportes_test.go` (16/16 ok).
- Plan secuencial de cierre de modulos incompletos - bloque 5 (Compras):
	- Se redise?a el dataset `operativo_compras_movimientos` en `backend/handlers/reportes.go` para consolidar compras por proveedor, dejando de depender solo de movimientos de inventario.
	- El dataset ahora expone KPI de ciclo documental: `ordenes_emitidas`, `recepciones`, `contabilizaciones`, `monto_ordenado`, `monto_recepcionado`, `monto_contabilizado`, `brecha_monto` y cumplimiento de recepcion/monto.
	- Se actualiza el catalogo del reporte con nuevo titulo y descripcion orientados a `costo por proveedor y recepcion vs orden`.
	- Se amplia `backend/handlers/reportes_test.go` con `TestEmpresaReportesHandlerDatasetOperativoComprasMovimientos` para validar consolidado por proveedor, totales de resumen y porcentajes de cumplimiento.
	- Se actualiza la matriz en `documentos/modulos del proyecto.md` marcando Compras como activo en reportes.
	- Validacion tecnica ejecutada:
		- `runTests` sobre `backend/handlers/reportes_test.go` con 8 pruebas objetivo (ok).
- Plan secuencial de cierre de modulos incompletos - bloque 4 (Inventario):
	- Se extiende el dataset `operativo_inventario_bodega` en `backend/handlers/reportes.go` con metricas de:
		- rotacion estimada y cobertura (`salida_promedio_diaria`, `dias_cobertura`, `indice_rotacion_30d`),
		- riesgo de quiebre proyectado (`estado_proyeccion`, `sugerido_reposicion`),
		- valorizacion por producto/bodega (`valorizacion_costo`, `valorizacion_venta`).
	- Se agregan KPI de resumen de inventario (`alertas`, `deficit`, `movimientos`, cobertura y rotacion promedio).
	- Se amplia `backend/handlers/reportes_test.go` con `TestEmpresaReportesHandlerDatasetOperativoInventarioBodega`.
	- Se actualiza matriz de estado en `documentos/modulos del proyecto.md` marcando Inventario como activo en reportes.
	- Se marca el bloque 4 como completado en `Pendiente Notas`.
	- Validacion tecnica ejecutada:
		- `go test ./handlers -run "TestEmpresaReportesHandler(DatasetOperativoInventarioBodega|DatasetOperativoCadenaCumplimiento|DatasetOperativoTarifasIngresos|DatasetOperativoReservasOcupacion|DatasetOperativoModulosResumen|CatalogoSuiteDataset|Exportes)" -count=1` (ok).
- Plan secuencial de cierre de modulos incompletos - bloque 3 (CRM/Produccion/Logistica):
	- Se agrega el dataset `operativo_cadena_cumplimiento` en `backend/handlers/reportes.go` para consolidar conversion comercial y cumplimiento operativo.
	- El dataset resume por modulo (`crm_leads`, `produccion_ordenes`, `logistica_envios`) registros de rango, estados finalizados/en proceso y monto de referencia.
	- Se amplia `backend/handlers/reportes_test.go` con `TestEmpresaReportesHandlerDatasetOperativoCadenaCumplimiento`.
	- Se actualiza matriz de estado en `documentos/modulos del proyecto.md` marcando CRM/Produccion/Logistica como activo en reportes.
	- Se marca el bloque 3 como completado en `Pendiente Notas`.
	- Validacion tecnica ejecutada:
		- `go test ./handlers -run "TestEmpresaReportesHandler(DatasetOperativoCadenaCumplimiento|DatasetOperativoTarifasIngresos|DatasetOperativoReservasOcupacion|DatasetOperativoModulosResumen|CatalogoSuiteDataset|Exportes)" -count=1` (ok).
- Plan secuencial de cierre de modulos incompletos - bloque 2 (tarifas):
	- Se consolida el dataset `operativo_tarifas_ingresos` para ingresos por modelo de tarifa (`tarifa_por_dia`, `tarifa_por_minutos`, `sin_modelo`).
	- Se amplia `backend/handlers/reportes_test.go` con `TestEmpresaReportesHandlerDatasetOperativoTarifasIngresos` y bootstrap de esquemas de tarifas (`EnsureEmpresaTarifasPorDiaSchema`, `EnsureEmpresaTarifasPorMinutosSchema`).
	- Se actualiza matriz de estado en `documentos/modulos del proyecto.md` marcando tarifas como activo en reportes.
	- Se marca el bloque de tarifas como completado en `Pendiente Notas`.
	- Validacion tecnica ejecutada:
		- `go test ./handlers -run "TestEmpresaReportesHandler(DatasetOperativoTarifasIngresos|DatasetOperativoReservasOcupacion|DatasetOperativoModulosResumen|CatalogoSuiteDataset|Exportes)" -count=1` (ok).
- Plan secuencial de cierre de modulos incompletos - bloque 1 (reservas):
	- Se agrega el dataset `operativo_reservas_ocupacion` en `backend/handlers/reportes.go` para consolidar ocupacion y cumplimiento por estacion.
	- Se amplian pruebas en `backend/handlers/reportes_test.go` con `TestEmpresaReportesHandlerDatasetOperativoReservasOcupacion` y bootstrap de `EnsureEmpresaReservasHotelSchema`.
	- Se actualiza matriz de estado en `documentos/modulos del proyecto.md` marcando reservas como activo en reportes.
	- Se documenta plan secuencial de cierre en `Pendiente Notas` y se marca reservas como primer modulo completado.
	- Validacion tecnica ejecutada:
		- `go test ./handlers -run "TestEmpresaReportesHandler" -count=1` (ok).
- Continuidad de plan en reportes por modulos:
	- Se valida y consolida el dataset `operativo_modulos_resumen` en `backend/handlers/reportes.go`.
	- Se corrige una llamada interna a `reportesCountByEmpresa` en el builder de resumen por modulos para compatibilidad con la firma actual de la funcion.
	- Se amplia `backend/handlers/reportes_test.go` con:
		- bootstrap de esquema para modulos ERP extendidos,
		- prueba `TestEmpresaReportesHandlerDatasetOperativoModulosResumen` con verificacion de conteos por modulo y consistencia de `summary`.
	- Validacion tecnica ejecutada:
		- `go test ./handlers -run "TestEmpresaReportesHandler" -count=1` (ok, 7/7).
- Continuidad de plan en frontend y ayuda:
	- Se corrige doble scrollbar en panel empresa eliminando la definicion duplicada de `.admin-empresa-frame` con altura fija en `web/estilos.css`.
	- Se actualiza `web/administrar_empresa/propinas.html` para recuperar consistencia visual/operativa (layout empresa, tablas estandar e integracion de menu flotante).
	- Se amplia `web/ayuda/ayuda.html` con guias de modulos pendientes: propinas, comisiones, ERP extendido y calculadora por empresa.
- Se continua el plan con dos faltantes operativos:
	- Utilidad nueva `web/administrar_empresa/calculadora.html` con contexto por empresa (`empresa_id`), memoria/historial aislados por empresa y exportacion JSON del historial.
	- Documento nuevo `documentos/modulos del proyecto.md` con inventario de modulos, conteo total y matriz base modulo -> reportes recomendados.
- Se integra la calculadora en navegacion:
	- `web/menu.js` agrega enlace `Calculadora` en menu flotante y propaga `empresa_id`.
	- `web/administrar_empresa.html` agrega enlace lateral `Calculadora`.
	- `web/js/administrar_empresa.js` incorpora `linkCalculadora` en navegacion y permisos (`finanzas/read`).
	- `web/estilos.css` agrega estilos `calc-*` para la nueva pantalla.
- Se actualiza documentacion tecnica:
	- `documentos/diagramas/estructura_del_codigo.md`.
	- `documentos/descripcion_del_proyecto`.
	- `documentos/descripcion_de_archivos`.
- Se completan faltantes de cobertura para la maquina de estados documental en ventas y CRM:
	- `backend/handlers/modulos_faltantes_test.go` ampl?a pruebas para:
		- ventas: `pedidos` y `devoluciones` (transiciones validas e invalidas),
		- CRM: `interacciones` y `campanas` (pipeline basico con validacion de transiciones).
	- Validacion tecnica ejecutada:
		- `go test ./handlers -run "TestEmpresa(IntegracionesAPIsHandlerHealthAndSync|IntegracionesBancosHandlerSyncAndEstado|VentasCotizacionesStateMachine|CRMLeadsStateMachine|VentasPedidosStateMachine|VentasDevolucionesStateMachine|CRMInteraccionesStateMachine|CRMCampanasStateMachine)" -count=1` (ok).
		- `go test ./... -run "^$" -count=1` (ok).
- Se implementan integraciones API/Bancos ejecutables y maquina de estados documental en CRM/Ventas:
	- `backend/handlers/modulos_faltantes.go` agrega handlers especializados sobre CRUD base:
		- Integraciones: `action=health_check`, `action=sync_manual`, `action=estado` en `/api/empresa/integraciones/apis` y `/api/empresa/integraciones/bancos`.
		- CRM/Ventas documentales: `action=estado`, `action=transiciones`, `action=transicionar` en rutas `/api/empresa/crm/*` y `/api/empresa/ventas/*`.
	- `backend/handlers/modulos_faltantes_test.go` (nuevo) cubre:
		- health/sync de integraciones,
		- transiciones validas/inv?lidas de cotizaciones y leads.
	- `web/js/modulos_erp_extendido.js` agrega botones operativos por fila:
		- Integraciones: `Health`, `Sync`, `Estado`.
		- CRM/Ventas documentales: `Transiciones`, `Transicionar`.
	- Validacion tecnica ejecutada:
		- `go test ./handlers -run "TestEmpresa(IntegracionesAPIsHandlerHealthAndSync|IntegracionesBancosHandlerSyncAndEstado|VentasCotizacionesStateMachine|CRMLeadsStateMachine)" -count=1` (ok).
		- `go test ./... -run "^$" -count=1` (ok).
- Inicio de implementacion del bloque de integraciones y pagos (fase de robustecimiento):
	- `backend/handlers/payments_handlers.go` agrega:
		- `MercadoPagoReconcileHandler` para conciliacion manual de pagos pendientes contra API Mercado Pago (`/mercadopago/reconcile`, requiere sesion admin).
		- `WompiWebhookHandler` para notificaciones servidor-servidor (`/wompi/webhook`) con validacion de firma y activacion automatica de licencia cuando aplica.
		- helpers compartidos para token MP, parseo de `external_reference`, estatus aprobados y activacion idempotente de licencia.
	- `backend/db/db.go` agrega helpers de persistencia para conciliacion:
		- listado de pendientes MP (`ListMPPaymentsForReconciliation`),
		- actualizacion por `id` y `payment_id`,
		- actualizacion Wompi por `reference`,
		- resolución de contexto licencia/empresa para Wompi.
	- `backend/main.go` registra rutas nuevas:
		- `/mercadopago/reconcile`
		- `/wompi/webhook`
	- Validacion tecnica ejecutada:
		- `go test ./auth ./db ./handlers ./metrics ./utils` (ok).
- Se divide la interfaz de ERP extendido en submodulos por dominio, manteniendo el mismo backend.
	- `web/administrar_empresa/modulos_erp_extendido.html` pasa a ser hub de dominios (ventas, finanzas, inventario/compras/rrhh, crm, produccion, logistica, documental/integraciones/dian).
	- `web/administrar_empresa/modulos_erp_dominio.html` (nuevo) concentra la operacion CRUD del dominio seleccionado sin cambiar endpoints backend.
	- `web/js/modulos_erp_extendido.js` (nuevo) centraliza la logica operativa reutilizable de submodulos por dominio.
	- `web/estilos.css` agrega estilos de navegacion y tarjetas para `erp-domain-*`.
- Se completa la operacion frontend de los modulos ERP extendidos en panel de empresa.
	- `web/administrar_empresa/modulos_erp_extendido.html` (nuevo) centraliza el uso de todos los endpoints ERP faltantes con:
		- listado con filtros (`q`, `limit`, `offset`, `include_inactive`),
		- detalle por ID,
		- crear/actualizar por payload JSON,
		- activar/desactivar y eliminacion logica por registro,
		- herramientas DIAN (`checklist`, `validar`, `generar_cufe_demo`, `generar_xml_demo`).
	- `web/administrar_empresa.html` agrega acceso lateral `ERP extendido`.
	- `web/js/administrar_empresa.js` integra `linkERPExtendido` en navegacion y permisos (modulo `seguridad`, accion `update`).
	- `web/estilos.css` agrega estilos dedicados del nuevo modulo (`erp-*`) para formularios, salida, tabla y estado visual.
- Se implementa base de modulos ERP faltantes en backend con esquema multiempresa, migracion y rutas nuevas:
	- `backend/db/modulos_faltantes.go` (tablas y CRUD generico por `empresa_id`).
	- `backend/handlers/modulos_faltantes.go` (rutas ERP adicionales y handler DIAN Colombia).
	- `backend/main.go` integra `EnsureEmpresaModulosFaltantesSchema`, migracion `2026-04-06-021-modulos-faltantes-erp` y `RegisterEmpresaModulosFaltantesRoutes`.
- Se agrega soporte DIAN Colombia operativo en endpoint `/api/empresa/facturacion_electronica/dian` con acciones:
	- `checklist` y `validar`.
	- `generar_cufe_demo` y `generar_xml_demo`.
- Se ampl?a `web/ayuda/ayuda.html` con seccion detallada para configurar facturacion DIAN desde cero.
- Se sincroniza documentacion tecnica y de BD:
	- `documentos/diagramas/estructura_del_codigo.md`.
	- `documentos/estructura_bd.md` y `estructura_bd.md`.
	- `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`.
- Validacion tecnica ejecutada:
	- `go test ./... -run "^$" -count=1` (ok).

## 2026-04-05
- Se continua con todos los bloques y pruebas en una corrida adicional de verificacion.
	- Validaciones ejecutadas:
		- `runTests` global (150 passed, 0 failed).
		- `powershell -File ..\\scripts\\validar_punto_13.ps1` (ok, suite productiva + suite completa backend).
	- Evidencia actualizada:
		- `documentos/punto_13_validacion_integral_resultado.md`.
		- `scripts/logs/punto13-validacion-20260405-182345.log`.

## 2026-04-05
- Se continua con todos los bloques y pruebas en una nueva corrida completa.
	- Validaciones ejecutadas:
		- `runTests` global (150 passed, 0 failed).
		- `powershell -File ..\\scripts\\validar_punto_13.ps1` (ok, suite productiva + suite completa backend).
	- Evidencia actualizada:
		- `documentos/punto_13_validacion_integral_resultado.md`.
		- `scripts/logs/punto13-validacion-20260405-182133.log`.

## 2026-04-05
- Se continua ejecucion de todos los bloques y pruebas con validacion ampliada por modulos criticos.
	- Validaciones ejecutadas:
		- `powershell -File ..\\scripts\\validar_punto_13.ps1` (ok, suite productiva + suite completa backend).
		- `go test ./handlers -run "TestEmpresa(Usuario|Clientes|Inventario|Compras|Facturacion|Finanzas|Auditoria|Permisos|Carritos)" -count=1` (ok).
		- `go test ./db -run "Test(Cliente|Proveedor|Inventario|Finanzas|Facturacion|Reserva|Vehiculo|Nomina|Tarifa|CodigoDescuento|Comision|Propina)" -count=1` (ok).
		- `go test ./... -count=1` (ok).
	- Evidencia actualizada:
		- `documentos/punto_13_validacion_integral_resultado.md`.
		- `scripts/logs/punto13-validacion-20260405-181807.log`.

## 2026-04-05
- Se reejecuta la validacion integral y bloques adicionales de pruebas para cierre tecnico.
	- Validaciones ejecutadas:
		- `powershell -File .\\scripts\\validar_punto_13.ps1` (ok, suite productiva + suite completa backend).
		- `go test ./handlers -run "TestEmpresa(CarritosCompraListIncluyeTarifaPorDiaAutomatica|TarifasPorMinutosHandlerCRUDAndCalculo|TarifasPorDiaHandlerCRUDAndCalculo|ReservasHotelHandlerCRUDAndDisponibilidad|VehiculosRegistroHandlerCRUDFlow|NominaSueldosHandlerFlow|PropinasHandlerConfigAndReporte|ComisionesServicioHandlerConfigAndReporte|ConfiguracionOperativaHandlerConfigAndRole)" -count=1` (ok).
		- `go test ./... -count=1` (ok).
	- Evidencia actualizada:
		- `documentos/punto_13_validacion_integral_resultado.md`.
		- `scripts/logs/punto13-validacion-20260405-181423.log`.

## 2026-04-05
- Se corrige regresion de pruebas de carritos ante el recalculo de totales al pagar estacion.
	- `backend/handlers/auth_users_carritos_test.go` actualiza el sembrado de datos en:
		- `TestEmpresaCarritosCompraAplicaPropinaSegunConfiguracion`.
		- `TestEmpresaCarritosCompraCodigoDescuentoConsumeUso`.
		- `TestEmpresaCarritosCompraRejectsMetodoPagoInvalido`.
	- Las pruebas ahora crean items reales en `carrito_compra_items` en lugar de forzar `subtotal/total` por SQL directo, quedando alineadas con `RefreshCarritoTotalConTarifaPorDia`.
	- Validacion ejecutada:
		- `go test ./handlers -run "TestEmpresaCarritosCompraAplicaPropinaSegunConfiguracion|TestEmpresaCarritosCompraCodigoDescuentoConsumeUso|TestEmpresaCarritosCompraRejectsMetodoPagoInvalido" -count=1` (ok).
		- `powershell -File .\\scripts\\validar_punto_13.ps1` (ok, incluye suite productiva y suite completa backend).

## 2026-04-05
- Se corrige el flujo de login de usuario de empresa para mantener alcance por `empresa_id` en endpoints publicos de autenticacion.
	- `backend/handlers/usuarios_empresa.go` ahora propaga `empresa_id` en enlaces de correo y confirmacion hacia `/login_usuario.htmlempresa_id=...`.
	- `ConfirmarCorreoUsuarioHandler` usa el `empresa_id` confirmado (o de query) para construir el enlace de retorno al login de usuario.
	- `web/js/login_usuario.js` toma `empresa_id` desde querystring y lo envia en query + body a `/api/empresa/usuarios/login` y `/api/empresa/usuarios/establecer_password`.
	- Validacion ejecutada:
		- `go test ./handlers -run "EmpresaUsuario(LoginHandlerSuccess|SetPasswordHandlerSuccess|LoginHandlerRejectsWrongEmpresaScopeFromQuery|SetPasswordHandlerRejectsWrongEmpresaScopeFromQuery)" -count=1` (ok).
		- `get_errors` sobre archivos modificados (sin errores).

## 2026-04-05
- Se completa auditoria integral de rutas `/api/empresa` y se cierra cobertura de wrappers por empresa al 100% en el registro de rutas.
	- `backend/handlers/empresa_permisos.go` agrega `WithEmpresaPublicScope` para endpoints publicos que requieren alcance por `empresa_id` sin autenticacion previa de admin.
	- `backend/main.go` envuelve `/api/empresa/usuarios/login` y `/api/empresa/usuarios/establecer_password` con `WithEmpresaPublicScope`.
	- `backend/main.go` envuelve `/api/empresa/facturacion_electronica/paises_disponibles` con `WithEmpresaFacturacionPermissions`.
	- `backend/handlers/chat_con_inteligencia_artificial_router.go` envuelve rutas del modulo IA (`modelos`, `modelo_preferido`, `consultar`, `historial`) con `WithEmpresaSeguridadPermissions`.
	- `web/administrar_empresa/facturacion_electronica.html` envia `empresa_id` al consultar `paises_disponibles` para compatibilidad con el wrapper de facturacion.
	- Validacion ejecutada:
		- `go test ./handlers -run "WithEmpresaVentasPermissionsInjectsEmpresaIDContextForParsers|EmpresaPermisosContextoHandlerRetornaPermisosPorRol|EmpresaUsuario(LoginHandlerRejectsWrongEmpresaScope|SetPasswordHandlerRejectsWrongEmpresaScope)|ModeloPreferidoHandler(Get|Put)RejectsEmpresaFueraDeAlcanceByGoogleAccount|HistorialHandlerRejectsEmpresaFueraDeAlcanceByGoogleAccount" -count=1` (ok).
		- `go test ./... -run "^$"` (ok).

## 2026-04-05
- Se refuerza la integracion multiempresa para que el alcance autorizado por `empresa_id` viaje en el contexto de request y sea reutilizable por handlers.
	- `backend/handlers/empresa_permisos.go` ahora inyecta `empresaID` en `context.Context` dentro de `WithEmpresa*Permissions`.
	- `backend/handlers/productos.go` actualiza `parseEmpresaIDQuery`, `parseInt64Query` y `parseInt64QueryOptional` para priorizar `empresaID` desde contexto cuando existe.
	- `backend/handlers/empresa_permisos_test.go` agrega `TestWithEmpresaVentasPermissionsInjectsEmpresaIDContextForParsers` para validar la propagacion de scope multiempresa sin dependencia estricta de querystring.
	- Validacion ejecutada: `go test ./handlers -run "WithEmpresaVentasPermissionsInjectsEmpresaIDContextForParsers|EmpresaPermisosContextoHandlerRetornaPermisosPorRol|EmpresaUsuario(LoginHandlerRejectsWrongEmpresaScope|SetPasswordHandlerRejectsWrongEmpresaScope)|ModeloPreferidoHandler(Get|Put)RejectsEmpresaFueraDeAlcanceByGoogleAccount|HistorialHandlerRejectsEmpresaFueraDeAlcanceByGoogleAccount" -count=1` (ok).

## 2026-04-05
- Se completa el subbloque de regresion UAT para endpoints sin wrapper de modulo (continuidad Punto 3).
	- `backend/handlers/auth_users_carritos_test.go` agrega cobertura de alcance por `empresa_id` enviado por querystring en:
		- `TestEmpresaUsuarioLoginHandlerRejectsWrongEmpresaScopeFromQuery`.
		- `TestEmpresaUsuarioSetPasswordHandlerRejectsWrongEmpresaScopeFromQuery`.
	- `backend/handlers/chat_con_inteligencia_artificial_controller_test.go` agrega cobertura de aislamiento por cuenta Google autenticada en:
		- `TestModeloPreferidoHandlerGetRejectsEmpresaFueraDeAlcanceByGoogleAccount`.
		- `TestModeloPreferidoHandlerPutRejectsEmpresaFueraDeAlcanceByGoogleAccount`.
		- `TestHistorialHandlerRejectsEmpresaFueraDeAlcanceByGoogleAccount`.
	- Validacion ejecutada: `go test ./handlers -run "EmpresaUsuario(LoginHandlerRejectsWrongEmpresaScope|SetPasswordHandlerRejectsWrongEmpresaScope)|ModelosHandlerRejectsEmpresaFueraDeAlcanceByGoogleAccount|ConsultarHandlerRejectsEmpresaFueraDeAlcance|ModeloPreferidoHandler(Get|Put)RejectsEmpresaFueraDeAlcanceByGoogleAccount|HistorialHandlerRejectsEmpresaFueraDeAlcanceByGoogleAccount" -count=1` (ok).

## 2026-04-05
- Se completa el subbloque de consumo frontend del contexto de permisos (cierre operativo Punto 3).
	- `web/js/administrar_empresa.js` ahora consume `GET /api/empresa/permisos_contextoempresa_id={id}` para resolver visibilidad real de enlaces por modulo/accion en el menu lateral.
	- El panel empresa ahora admite `id` o `empresa_id` en querystring para resolver el contexto de permisos sin ambiguedad.
	- Se mantiene fallback local por rol cuando el endpoint no esta disponible, evitando bloqueos de navegacion.
	- `web/administrar_empresa.html` agrega indicador visual `menuPermsEvidence` para evidencia UAT del rol/fuente de permisos aplicado en pantalla.
	- Validacion ejecutada: `get_errors` sobre frontend modificado (sin errores).

## 2026-04-05
- Se agrega endpoint de contexto de permisos por empresa para reforzar el cierre del Punto 3 (permisos y seguridad).
	- `backend/handlers/empresa_permisos.go` incorpora `GET /api/empresa/permisos_contexto` con respuesta de permisos efectivos por modulo/accion para el rol autenticado.
	- El endpoint soporta `include_matrix=1` para retornar matriz completa por rol (`super_administrador`, `admin_empresa`, `supervisor_sucursal`, `cajero`, `inventario`, `compras`, `contabilidad`, `auditor`).
	- `backend/main.go` registra la ruta bajo `WithEmpresaSeguridadPermissions` para mantener aislamiento por `empresa_id`.
	- `backend/handlers/empresa_permisos_test.go` agrega `TestEmpresaPermisosContextoHandlerRetornaPermisosPorRol` y `TestEmpresaPermisosContextoHandlerIncluyeMatrizRoles`.
	- Validacion ejecutada: `go test ./handlers -run "PermisosContexto|WithEmpresa.*Permissions" -count=1` (ok).

## 2026-04-05
- Se amplian los reportes contables de flujo de caja con filtros por categoria y metodo de pago.
	- `backend/handlers/reportes.go` incorpora filtros `categoria` y `metodo_pago` en `contable_flujo_caja` para segmentar ingresos/egresos diarios.
	- El resumen del dataset ahora expone `filtro_categoria` y `filtro_metodo_pago` para trazabilidad del reporte exportado.
	- `backend/handlers/reportes_test.go` agrega `TestEmpresaReportesHandlerDatasetContableFlujoCajaFiltros` para validar segmentacion por categoria/metodo.
	- `web/administrar_empresa/reportes.html` agrega campos de filtro contable y los propaga en consultas/exportaciones del endpoint `/api/empresa/reportes`.

## 2026-04-05
- Se extiende el modulo de reportes con dataset contable de flujo de caja diario.
	- `backend/handlers/reportes.go` agrega dataset `contable_flujo_caja` en `/api/empresa/reportes` y consolida ingresos, egresos, neto del dia, saldo acumulado y conteo de movimientos por fecha.
	- El dataset mantiene paridad de exportacion en `pdf`, `xls`, `csv`, `json` y `txt` desde el catalogo central de reportes empresariales.
	- `backend/handlers/reportes_test.go` agrega `TestEmpresaReportesHandlerDatasetContableFlujoCaja` para validar filas diarias y resumen acumulado del periodo.

## 2026-04-05
- Se extiende el modulo de reportes con dataset contable de liquidaciones de nomina y exportacion PDF.
	- `backend/handlers/reportes.go` agrega dataset `contable_nomina_liquidaciones` con filtros por periodo y `empleado_nomina_id`.
	- `backend/handlers/reportes.go` habilita formato `pdf` en la exportacion de datasets de `/api/empresa/reportes`.
	- `web/administrar_empresa/reportes.html` agrega opcion `PDF` en el selector de formato.
	- `web/administrar_empresa/nomina_sueldos.html` incorpora accion `Exportar liquidaciones` usando `/api/empresa/reportesaction=export`.
	- `backend/handlers/reportes_test.go` agrega cobertura de dataset de nomina y validacion de export PDF.

## 2026-04-05
- Se integra operativamente el modulo de nomina de sueldos con asistencia en backend y panel de empresa.
	- `backend/main.go` incorpora `EnsureEmpresaNominaSchema`, migracion `2026-04-05-020-nomina-sueldos` y ruta `/api/empresa/nomina`.
	- `web/administrar_empresa/nomina_sueldos.html` (nuevo) agrega configuracion legal, empleados, festivos, calculo y consulta de liquidaciones.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran `linkNominaSueldos` en menu y permisos.
	- `web/estilos.css` agrega estilos dedicados del modulo.
	- `documentos/estructura_bd.md` y `estructura_bd.md` incluyen tablas y relaciones de nomina.

## 2026-04-05
- Se agrega `ventas_simple.html` como carrito alterno por estación (modo supermercado) con activaci?n/desactivaci?n por estación.
	- `web/administrar_empresa/ventas_simple.html` (nuevo) incorpora flujo r?pido para buscar productos, agregarlos al carrito, ajustar cantidades y visualizar total consolidado por estación.
	- Se corrige la visibilidad del campo de referencia de pago para métodos que la requieren (`tarjeta_credito`, `tarjeta_debito`, `transferencia_bancaria`).
	- El cobro se ejecuta con flujo simplificado usando `action=pagar_estacion` y permite iniciar nueva venta con `action=activar_estacion`.
	- `web/administrar_empresa/configuracion_de_estaciones.html` agrega la bandera local `venta_simple_habilitada` por estación.
	- `web/administrar_empresa/estaciones.html` enruta autom?ticamente cada estación al carrito completo (`carrito_de_compras.html`) o al carrito simple (`ventas_simple.html`) seg?n su configuraci?n.
	- `web/estilos.css` integra estilos responsive para el nuevo módulo y etiqueta visual del modo por estación.

## 2026-04-05
- Se actualiza la configuracion de `agente_go` para reforzar reportes e interoperabilidad contable.
	- `.github/agents/agente_go.agent.md` agrega regla obligatoria para que todos los reportes puedan exportarse, como minimo, en `PDF` y `Excel`, y tambien en formatos de uso comun (`CSV`, `JSON`, `TXT`).
	- Se incorpora regla de compatibilidad con software contable externo mediante formatos estandar de intercambio y trazabilidad por `empresa_id`, documento y periodo.

## 2026-04-05
- Se agrega el dataset `reporte_de_turno` al modulo empresarial de reportes para control operativo de caja por turno.
	- `backend/handlers/reportes.go` incorpora `reporte_de_turno` en `/api/empresa/reportes` con filtros por `usuario`, `caja_codigo`, `turno` y `cierre_id`.
	- El dataset incluye detalle por carrito con `activado_en`, `pagado_en`, metodo de pago y acumulados de ventas por `producto` y `servicio`.
	- El resumen del reporte calcula gastos de turno y efectivo esperado (`efectivo_deberia_haber`) combinando cierres de caja y movimientos financieros.
	- `web/administrar_empresa/reportes.html` agrega campos de filtro de turno/caja/usuario/cierre y envia estos parametros en consultas y exportes del dataset.
	- `backend/handlers/reportes_test.go` agrega `TestEmpresaReportesHandlerDatasetReporteTurno` para validar filtros y consistencia del resumen financiero del turno.

## 2026-04-05
- Se crea el modulo de tarifas por dia por estacion con recálculo automático de deuda en carritos hotel activos.
	- `backend/db/tarifas_por_dia.go` (nuevo) agrega esquema `empresa_tarifas_por_dia`, CRUD, horarios `hora_check_in`/`hora_check_out` y calculo de dias/monto.
	- `backend/db/carritos_tarifa_dia.go` (nuevo) integra calculo automático de deuda diaria en carritos de estación y refresco masivo para listados.
	- `backend/db/carritos_compras.go` ajusta `RecalculateCarritoCompraTotals` para incluir tarifa diaria cuando aplique.
	- `backend/handlers/tarifas_por_dia.go` (nuevo) expone `/api/empresa/tarifas_por_dia` con acciones `listar`, `detalle`, `aplicable`, `calcular`, `activar` y `desactivar`.
	- `backend/handlers/carritos_compras.go` recalcula tarifa diaria al listar carritos y antes de `action=pagar_estacion`.
	- `backend/main.go` integra `EnsureEmpresaTarifasPorDiaSchema`, migracion `2026-04-05-019-tarifas-por-dia` y ruta protegida del modulo.
	- `web/administrar_empresa/tarifas_por_dia.html` (nuevo) agrega UI de configuracion, filtros y simulador por rango de fechas.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran `linkTarifasPorDia` en menu lateral y permisos.
	- Cobertura agregada en `backend/db/tarifas_por_dia_test.go`, `backend/handlers/tarifas_por_dia_test.go` y `backend/handlers/carritos_tarifa_por_dia_test.go`.

## 2026-04-05
- Se crea el modulo de tarifas por minutos por estacion con reglas por dia de semana y calculo de bloques extra.
	- `backend/db/tarifas_por_minutos.go` (nuevo) agrega esquema `empresa_tarifas_por_minutos`, CRUD, resolucion por dia (`dia_semana_desde/hasta`) y calculo de monto por minutos consumidos.
	- `backend/handlers/tarifas_por_minutos.go` (nuevo) expone `/api/empresa/tarifas_por_minutos` con acciones `listar`, `detalle`, `aplicable`, `calcular`, `activar` y `desactivar`.
	- `backend/main.go` integra `EnsureEmpresaTarifasPorMinutosSchema`, migracion `2026-04-05-018-tarifas-por-minutos` y ruta protegida del modulo.
	- `web/administrar_empresa/tarifas_por_minutos.html` (nuevo) agrega formulario de tarifas, filtros y simulador de cobro por minutos.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran `linkTarifasPorMinutos` en menu lateral y permisos por rol.
	- Cobertura agregada en `backend/db/tarifas_por_minutos_test.go` y `backend/handlers/tarifas_por_minutos_test.go`.
	- Se actualiza documentacion y diagramas: `documentos/estructura_bd.md`, `estructura_bd.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md` y `documentos/diagramas/diagrama_flujo_procesos.md`.

## 2026-04-05
- Se ubica el documento de base de datos dentro de la carpeta `documentos` y se alinea `agente_go` con esa ruta.
	- `documentos/estructura_bd.md` se incorpora como ubicacion requerida de la estructura de base de datos.
	- `.github/agents/agente_go.agent.md` ahora exige revisar `documentos/estructura_bd.md` antes de cambios en tablas, consultas, migraciones o datos.
	- `estructura_bd.md` en raiz se mantiene sincronizado como copia de compatibilidad documental.

## 2026-04-05
- Se actualiza la configuracion de `agente_go` para forzar lectura previa de documentacion de base de datos en tareas de datos.
	- `.github/agents/agente_go.agent.md` agrega regla para revisar `estructura_bd.md` antes de cambios en tablas, consultas, migraciones o datos operativos.

## 2026-04-05
- Se agrega modulo de busqueda y gestion de facturas electronicas por empresa.
	- `web/administrar_empresa/facturas_electronicas.html` (nuevo) permite buscar por cliente, documento y rango de fechas; ver detalle; reenviar por correo; e imprimir.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran el acceso lateral `Facturas electrónicas` con permisos de lectura del modulo `facturacion`.
	- `backend/handlers/facturacion_electronica.go` incorpora:
		- `GET action=documentos` para consulta de documentos de facturacion por filtros (`cliente`, `documento`, `fecha_desde`, `fecha_hasta`, `estado_documento`, `tipo_documento`, `q`).
		- `PUT/POST action=reenviar_correo` para reintento manual de envio al correo del cliente.
	- `backend/db/documentos_transaccionales.go` agrega consulta enriquecida con cliente (`nombre`, `email`, `documento`) para listado filtrado.
	- Cobertura agregada en `backend/db/documentos_transaccionales_test.go` para filtros de cliente/fecha/documento.

## 2026-04-05
- Se normaliza la documentacion de base de datos para eliminar duplicidad entre documentos.
	- `estructura_bd.md` queda como fuente canonica del esquema fisico motor legado retirado.
	- `documentos/descripcion_de_las_bases_De_datos` se redefine como guia complementaria funcional y reglas operativas de mantenimiento.
	- Se evita repetir listados tabla-por-tabla en dos archivos distintos.

## 2026-04-05
- Se consolida Configuraci?n avanzada dentro de Facturaci?n electrónica en el panel de empresa.
	- `web/administrar_empresa/facturacion_electronica.html` integra el formulario completo de configuraci?n avanzada fiscal/impresión y su persistencia mediante `/api/empresa/configuracion_avanzada`.
	- `web/administrar_empresa.html` elimina el enlace lateral independiente `Configuraci?n avanzada` para dejar una ?nica entrada funcional en `Facturaci?n electrónica`.
	- `web/js/administrar_empresa.js` retira `linkConfigAvanzada` del cat?logo de enlaces/permisos del menú.
	- `web/ayuda/ayuda.html` actualiza el tutorial para indicar que la configuraci?n avanzada ahora se gestiona desde `facturacion_electronica.html`.
	- `web/administrar_empresa/configuracion_avanzada.html` se elimina del repositorio por consolidaci?n funcional.

## 2026-04-05
- Se agrega configuracion operativa de cobro por empresa y por rol de usuario.
	- `backend/db/configuracion_operativa.go` (nuevo) agrega tablas `empresa_configuracion_operativa` y `empresa_configuracion_operativa_roles`, con resolucion efectiva de permisos por rol.
	- `backend/handlers/configuracion_operativa.go` (nuevo) expone `/api/empresa/configuracion_operativa` para consultar y actualizar reglas base y overrides por rol (`action=rol`).
	- `backend/handlers/empresa_permisos.go` y `backend/handlers/productos.go` propagan/normalizan rol admin en request para enforcement transversal.
	- `backend/handlers/carritos_compras.go` aplica enforcement en `action=pagar_estacion`: bloquea metodos de pago no permitidos y desactiva propina/comision segun politica operativa efectiva por rol.
	- `backend/main.go` registra `EnsureEmpresaConfiguracionOperativaSchema`, migracion `2026-04-05-017-configuracion-operativa-cobro` y ruta protegida `/api/empresa/configuracion_operativa`.
	- `web/administrar_empresa/configuracion.html` incorpora tarjeta de checks para metodos de pago, propinas y comisiones por empresa y por rol.
	- `web/administrar_empresa/carrito_de_compras.html` consume la politica operativa efectiva y refleja en UI los metodos permitidos, con bloqueo visual y validacion previa al pago.
	- Cobertura agregada en `backend/db/configuracion_operativa_test.go`, `backend/handlers/configuracion_operativa_test.go` y `backend/handlers/auth_users_carritos_test.go`.
	- Validacion ejecutada: pruebas dirigidas en DB/handlers/carritos (ok) y verificacion sin errores en frontend actualizado.

## 2026-04-05
- Se crea el modulo de comisiones por servicio por empresa con reporte por lavador.
	- `backend/db/comisiones_servicio.go` (nuevo) agrega tablas de configuracion y movimientos (`empresa_comisiones_servicio_configuracion`, `empresa_comisiones_servicio_movimientos`) con calculo/reporte por lavador.
	- `backend/handlers/comisiones.go` (nuevo) expone `/api/empresa/comisiones` con acciones `config`, `reporte` y `movimientos`.
	- `backend/handlers/carritos_compras.go` integra `usuario_lavador` en `action=pagar_estacion` y registra comisiones automaticas de servicios de lavado al cerrar venta.
	- `backend/main.go` asegura esquema de comisiones, registra migracion `2026-04-05-016-comisiones-servicio` y publica ruta protegida de comisiones bajo permisos de finanzas.
	- `web/administrar_empresa/comisiones.html` (nuevo) incorpora configuracion y reporte de comisiones por lavador.
	- `web/administrar_empresa/carrito_de_compras.html` agrega captura de lavador para comision, carga de configuracion de comisiones y visualizacion de comision estimada/registrada en cobro.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran acceso lateral `Comisiones` (`linkComisiones`) con permisos del modulo finanzas.
	- Cobertura agregada en `backend/db/comisiones_servicio_test.go`, `backend/handlers/comisiones_test.go` y `backend/handlers/auth_users_carritos_test.go`.

## 2026-04-05
- Se actualiza la configuracion de `agente_go` para definir semantica y concurrencia de estaciones.
	- `.github/agents/agente_go.agent.md` agrega que una estacion puede representar mesa de restaurante, habitacion de hotel, habitacion de motel, punto de caja u otro punto operativo equivalente.
	- Se establece que estaciones deben soportar multiples carritos/sesiones y multiples clientes en simultaneo, con aislamiento por `empresa_id` y trazabilidad operativa.

## 2026-04-05
- Se completa el modulo de reservas por estacion/habitacion para operacion empresarial.
	- `backend/db/reservas_hotel.go` (nuevo) implementa esquema y logica de reservas con disponibilidad por rango, conflicto de solapamiento, expiracion de pendientes, confirmacion de pago, cancelacion, activacion/desactivacion y eliminacion.
	- `backend/handlers/reservas_hotel.go` (nuevo) expone `/api/empresa/reservas_hotel` con acciones `listar`, `detalle`, `disponibilidad`, `confirmar_pago`, `cancelar`, `activar`, `desactivar` y CRUD operativo.
	- `backend/main.go` asegura esquema de reservas, registra migracion `2026-04-05-015-reservas-hotel` y publica ruta protegida bajo permisos de ventas.
	- `web/administrar_empresa/reservas_hotel.html` (nuevo) agrega interfaz para crear/editar reservas, consultar disponibilidad y ejecutar acciones de ciclo de vida.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran el acceso lateral `Reservas` (`linkReservasHotel`) con control de permisos por rol.
	- Cobertura de pruebas en:
		- `backend/db/reservas_hotel_test.go` (flujo DB end-to-end con disponibilidad y estados).
		- `backend/handlers/reservas_hotel_test.go` (nuevo, flujo HTTP completo del endpoint).
	- Validaciones ejecutadas:
		- `go test ./db -run ReservaHotel -count=1` (ok).
		- `go test ./handlers -run ReservasHotel -count=1` (ok).

## 2026-04-05
- Se crea el modulo de registro de vehiculos por empresa para controlar ingreso y salida por patente.
	- `backend/db/vehiculos_registro.go` (nuevo) agrega esquema y operaciones CRUD del registro vehicular, con estado operativo (`en_empresa`/`retirado`) y marcacion de salida.
	- `backend/handlers/vehiculos_registro.go` (nuevo) expone `/api/empresa/vehiculos_registro` con acciones de consulta, alta, edicion, activar/desactivar, marcar salida y eliminacion.
	- `backend/main.go` asegura esquema del modulo, registra migracion `2026-04-05-014-vehiculos-registro` y publica ruta protegida bajo permisos de seguridad.
	- `web/administrar_empresa/vehiculos_registro.html` (nuevo) incorpora UI de registro de vehiculos con patente, conductor, propietario, fechas de ingreso/salida y filtros operativos.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran el acceso lateral `Registro de vehiculos` con permisos por rol.
	- Cobertura agregada en `backend/db/vehiculos_registro_test.go` y `backend/handlers/vehiculos_registro_test.go`.
	- Validaciones ejecutadas:
		- `go test ./db -run Vehiculo -count=1` (ok).
		- `go test ./handlers -run VehiculosRegistro -count=1` (ok).

## 2026-04-05
- Se crea el modulo de propinas por empresa con configuracion operativa y reporte por usuario o universal.
	- `backend/db/propinas.go` (nuevo) agrega tablas de configuracion y movimientos de propinas, con soporte de reporte acumulado por usuario y reparto universal entre usuarios activos.
	- `backend/handlers/propinas.go` (nuevo) expone `/api/empresa/propinas` con acciones de configuracion, reporte y consulta de movimientos.
	- `backend/handlers/carritos_compras.go` integra propina en `action=pagar_estacion`, valida `total_pagado` contra total final con propina y registra movimiento de propina al cerrar venta.
	- `backend/main.go` asegura esquema de propinas, registra migracion `2026-04-05-013-propinas` y publica ruta protegida `/api/empresa/propinas` bajo permisos de finanzas.
	- `web/administrar_empresa/propinas.html` (nuevo) incorpora modulo de configuracion de propinas y reporte por rango, usuario y modo.
	- `web/administrar_empresa/carrito_de_compras.html` agrega control de aplicar propina en cobro de estacion, carga de configuracion y desglose de total final con propina.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran acceso de menu `Propinas` con permisos del modulo finanzas.
	- Cobertura agregada en `backend/db/propinas_test.go`, `backend/handlers/propinas_test.go` y `backend/handlers/auth_users_carritos_test.go`.
	- Validaciones ejecutadas:
		- `go test ./db -run Propina -count=1` (ok).
		- `go test ./handlers -run "Propinas|CarritosCompraAplicaPropinaSegunConfiguracion" -count=1` (ok).
		- `go test ./db ./handlers -count=1` (ok).

## 2026-04-05
- Se agrega `transferencia_bancaria` como forma de pago transversal en flujo de carritos y finanzas.
	- `backend/db/carritos_compras.go` normaliza y acepta alias de transferencia bancaria (`transferencia`, `transferencia_bancaria`).
	- `backend/handlers/carritos_compras.go` habilita transferencia bancaria en pago directo y mixto, y exige referencia minima para tarjeta/transferencia.
	- `backend/handlers/auth_users_carritos_test.go` agrega cobertura de pago exitoso por transferencia bancaria y rechazo cuando falta referencia valida.
	- `web/administrar_empresa/carrito_de_compras.html` incorpora transferencia bancaria en selectores de pago, habilita validacion de pago mixto con transferencia y envía `pagos_mixtos` al backend.
	- `web/administrar_empresa/finanzas.html` estandariza opcion de `transferencia_bancaria` y mantiene compatibilidad con registros legacy `transferencia`.
	- `web/ayuda/ayuda.html` actualiza descripcion de metodos soportados en cierre de carrito.
	- `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md` y `documentos/diagramas/diagrama_flujo_procesos.md` reflejan el nuevo flujo de pago.

## 2026-04-05
- Robustecimiento del modulo de auditoria empresarial con foco en trazabilidad, seguridad operativa y analisis forense.
	- `backend/db/auditoria_empresa.go` ampl?a filtros (`metodo_http`, `recurso`, `endpoint`, `search`), agrega `offset`, agrega conteo filtrado (`CountEmpresaAuditoriaEventos`) y refuerza indices de rendimiento.
	- `backend/handlers/auditoria_empresa.go` valida fechas/parametros, publica metadata de paginacion por headers y soporta consulta avanzada de eventos.
	- `backend/handlers/empresa_permisos.go` registra intentos criticos denegados (401/403/500) como eventos de auditoria no bloqueantes.
	- `backend/utils/utils.go` expone `RequestIDFromContext` para correlacion real entre logs de request y eventos de auditoria.
	- `web/administrar_empresa/auditoria.html` agrega filtros avanzados, paginador y panel de detalle JSON por evento.
	- `web/estilos.css` agrega estilos centralizados para paginacion y detalle del modulo de auditoria.
	- Se amplian pruebas en `backend/db/auditoria_empresa_test.go` y `backend/handlers/auditoria_empresa_test.go`.
	- Validaciones ejecutadas:
		- `go test ./db -run Auditoria -count=1` (ok).
		- `go test ./handlers -run Auditoria -count=1` (ok).
		- `go test ./auth ./db ./handlers ./metrics ./utils -count=1` (ok).

## 2026-04-05
- Facturaci?n electrónica: env?o automático del resumen de factura al correo del cliente al emitir.
	- `backend/handlers/facturacion_electronica.go` ahora intenta enviar correo en `action=emitir` de `factura_electronica`.
	- Soporta destinatario por `cliente_email` o por `cliente_id`/`entidad_id` consultando clientes.
	- La respuesta incluye bloque `factura_email` con estado de intento/env?o/error sin bloquear la emisi?n legal.
	- `backend/db/clientes.go` agrega `GetClienteByID` para resolver destinatario desde la base de datos.
	- `backend/main.go` actualiza la inyecci?n de `dbSuper` al handler de facturación para lectura de SMTP.
	- `web/administrar_empresa/facturacion_electronica.html` agrega campos de cliente y muestra el resultado de env?o en pantalla.
	- Cobertura a?adida en `backend/db/clientes_test.go` y `backend/handlers/eventos_contables_modulos_test.go`.

## 2026-04-05
- Se crea el modulo de codigos de descuento por empresa y validacion de metodos de pago en carrito de compras.
	- `backend/db/codigos_descuento.go` (nuevo) agrega la tabla `codigos_de_descuento`, generacion automatica de codigos, CRUD, validacion por vencimiento/usos y resolución de descuento aplicable por monto.
	- `backend/handlers/codigos_descuento.go` (nuevo) expone `/api/empresa/codigos_de_descuento` con operaciones CRUD, activar/desactivar y `action=validar`.
	- `backend/db/carritos_compras.go` agrega campos `metodo_pago` y `referencia_pago`, normaliza metodos permitidos y registra consumo transaccional de codigo de descuento al cerrar venta.
	- `backend/handlers/carritos_compras.go` valida `metodo_pago` (`efectivo`, `tarjeta_credito`, `tarjeta_debito`, `codigo_descuento`) y exige referencia para pagos con tarjeta.
	- `backend/main.go` asegura esquema `codigos_de_descuento`, registra migracion `2026-04-05-012-codigos-descuento-pagos` y expone ruta protegida de codigos de descuento.
	- `web/administrar_empresa/codigos_de_descuento.html` (nuevo) incorpora modulo profesional para crear/editar/activar/eliminar codigos con valor y fecha de vencimiento.
	- `web/administrar_empresa/carrito_de_compras.html` agrega selector de metodo de pago, referencia y aplicacion de codigos de descuento con validacion operativa.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran el enlace de menu `Codigos de descuento` con permisos del modulo ventas.
	- `backend/db/codigos_descuento_test.go` y `backend/handlers/auth_users_carritos_test.go` agregan cobertura para validacion/uso de codigos y rechazo de metodo de pago invalido.

## 2026-04-05
- Se crea el modulo de combos de productos con receta de ingredientes y precio unico de venta.
	- `backend/handlers/combos_productos.go` (nuevo) expone `/api/empresa/combos_productos` con operaciones CRUD y acciones `activar/desactivar`.
	- `backend/db/productos.go` incorpora esquema y logica de combos (`combos_productos`, `combos_productos_detalle`) con controles de consistencia para carritos abiertos.
	- `backend/db/carritos_compras.go` extiende el ajuste de inventario para descontar/liberar stock por ingrediente cuando el item es `tipo_item=combo`.
	- `backend/handlers/carritos_compras.go` valida `referencia_id` obligatorio para items combo.
	- `backend/main.go` registra la nueva ruta protegida bajo permisos de inventario.
	- `web/administrar_empresa/combos_productos.html` (nuevo) agrega interfaz completa para gestionar combos y receta.
	- `web/administrar_empresa/carrito_de_compras.html` incorpora busqueda/catalogo y visualizacion de combos en carrito.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran el modulo en menu y permisos.
	- `backend/db/productos_categorias_test.go` y `backend/db/carritos_inventario_test.go` agregan cobertura de CRUD y flujo de inventario por ingredientes.

## 2026-04-05
- Se crea el modulo de graficos y estadisticas por empresa.
	- `backend/handlers/graficos_estadisticas.go` (nuevo) expone `/api/empresa/graficos_estadisticas` con acciones `panel`, `serie`, `rankings`, `distribuciones` y `catalogo`.
	- `backend/main.go` registra la nueva ruta protegida bajo permisos de finanzas.
	- `backend/handlers/graficos_estadisticas_test.go` (nuevo) agrega cobertura de contrato HTTP y validaciones de error.
	- `web/administrar_empresa/graficos_estadisticas.html` (nuevo) incorpora panel visual con series, distribuciones y rankings.
	- `web/estilos.css` agrega estilos responsivos del nuevo modulo.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran el acceso en menu con control de permisos.
	- `web/ayuda/ayuda.html` incorpora guia y API del modulo de analitica.

## 2026-04-05
- Se crea el modulo de control de asistencia de empleados por empresa.
	- `backend/db/asistencia_empleados.go` (nuevo) agrega tabla `empresa_asistencia_empleados` y operaciones CRUD con marcacion de entrada/salida.
	- `backend/handlers/asistencia_empleados.go` (nuevo) expone `/api/empresa/asistencia_empleados` con acciones operativas de asistencia.
	- `backend/main.go` incorpora esquema, migracion `2026-04-05-010-asistencia-empleados` y registro de ruta protegida.
	- `web/administrar_empresa/asistencia_empleados.html` (nuevo) agrega UI completa para gestion diaria de asistencia.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran el modulo en menu y permisos.
	- `backend/handlers/asistencia_empleados_test.go` (nuevo) valida flujo funcional del modulo.
	- Se actualizan `web/ayuda/ayuda.html`, `estructura_bd.md` y diagramas/documentacion tecnica para trazabilidad.

## 2026-04-05
- Modulo de reportes robustecido a nivel empresarial, operativo y contable con enfoque escalable por dataset.
	- `backend/handlers/reportes.go` (nuevo) implementa `/api/empresa/reportes` con acciones `catalogo`, `suite`, `dataset`, `tablero` y `export`.
	- Se habilitan exportaciones multi-formato para datasets: `JSON`, `CSV`, `TXT` y `XLS`.
	- `backend/main.go` registra la nueva ruta protegida bajo permisos de finanzas.
	- `web/administrar_empresa/reportes.html` incorpora selector de dataset, vista tabular profesional y exportes desde interfaz.
	- `backend/handlers/reportes_test.go` (nuevo) agrega cobertura de contrato HTTP y validacion de exportaciones.
	- Se actualizan diagramas de arquitectura/flujo en `documentos/diagramas/estructura_del_codigo.md` y `documentos/diagramas/diagrama_flujo_procesos.md`.

## 2026-04-04
- Centro de ayuda actualizado con tutorial por cada módulo del sistema.
	- `web/ayuda/ayuda.html` ampl?a el contenido con una secci?n de tutoriales por módulos de administraci?n global y módulos del panel de empresa.
	- Se agregan pasos operativos por módulo y enlaces directos a cada pantalla para facilitar onboarding y uso diario.

## 2026-04-04
- Verificacion integral real de modulos + limpieza de artefactos temporales.
	- Validacion real ejecutada (sin simulaciones/mocks) sobre motor legado retirado y capa HTTP:
		- `go test ./auth ./db ./handlers ./metrics ./utils -count=1` (ok).
		- `go test ./... -count=1` (ok).
		- `go test ./handlers -run "TestEmpresaScope|FueraDeAlcance|WithEmpresa|isol|Aisla|multiempresa|UsuariosHandlerAislaEmpresa|ConsolidaEmpresa" -count=1` (ok).
		- `go test ./handlers -run "TestEmpresaClientes|TestEmpresaProveedores|TestEmpresaFacturacion|TestEmpresaCompras|TestEmpresaInventario|TestEmpresaFinanzas|TestEmpresaAuditoria|TestEmpresaCarritos|TestEmpresaUsuarios|TestModelosHandler" -count=1` (ok).
		- `go test ./db -run "Test.*(Cliente|Proveedor|Facturacion|Compra|Inventario|Finanzas|Evento|Auditoria|Scope|Empresa)" -count=1` (ok).
	- Se eliminan artefactos temporales/no usados del repositorio:
		- `backend/tmp_api.json`.
		- `backend/tmp_config.html`.
		- `backend/server.err`.
		- `backend/server.run.err`.
		- `backend/db/pcs_empresas.20260326-174525.bak`.
		- `backend/db/pcs_superadministrador.20260326-174324.bak`.
		- `backend/db/pcs_superadministrador.20260326-174525.bak`.

## 2026-04-04
- Punto 14 (operacion continua) - inicio operativo con KPI y roadmap trimestral.
	- `documentos/punto_14_operacion_continua.md` (nuevo): define marco de mejora continua y cadencia de seguimiento.
	- `documentos/roadmap_trimestral_pos_multiempresa.md` (nuevo): formaliza roadmap Q2/Q3/Q4 2026.
	- `scripts/generar_reporte_operacion_continua.ps1` (nuevo): genera reporte operativo y bitacora tecnica.
	- `documentos/punto_14_operacion_continua_reporte.md` (nuevo): evidencia de la ultima corrida operativa.
	- `documentos/plan_maestro_pos_multiempresa_14_puntos.md`: punto 14 actualizado a `en curso`.
- Validacion tecnica:
	- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\generar_reporte_operacion_continua.ps1` (ok).

## 2026-04-04
- Punto 13 (calidad, UAT y despliegue) - arranque operativo con validacion integral automatizada.
	- `scripts/validar_punto_13.ps1` (nuevo): ejecuta gate tecnico y genera evidencia automatica.
	- `documentos/punto_13_calidad_uat_despliegue.md` (nuevo): formaliza flujo de calidad/UAT/salida controlada.
	- `documentos/punto_13_validacion_integral_resultado.md` (nuevo): reporte de ultima validacion tecnica.
	- `documentos/release_checklist.md`: incorpora gate del punto 13 y verificacion de evidencia.
	- `documentos/plan_maestro_pos_multiempresa_14_puntos.md`: punto 13 pasa a `en curso`.
- Validacion tecnica:
	- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\validar_punto_13.ps1` (ok).
	- `go test ./auth ./db ./handlers ./metrics ./utils -count=1` (ok).
	- `go test ./... -count=1` (ok).

## 2026-04-04
- Punto 8 (facturacion electronica) - refuerzo de cobertura en cumplimiento legal de emision.
	- `backend/db/facturacion_electronica_test.go` (nuevo) agrega pruebas unitarias para `PrepareFacturacionDocumentoLegal`:
		- `TestPrepareFacturacionDocumentoLegalSuccessAndConsecutivo`.
		- `TestPrepareFacturacionDocumentoLegalRejectsExpiredResolution`.
		- `TestPrepareFacturacionDocumentoLegalRejectsConfigInactivaAndRangoAgotado`.
	- Se valida reserva e incremento de consecutivo legal, rechazo por resolucion vencida, rechazo por configuracion FE inactiva y agotamiento de rango.
- Validacion tecnica:
	- `gofmt -w db/facturacion_electronica_test.go` (ok).
	- `go test ./db -run "TestPrepareFacturacionDocumentoLegal" -count=1` (ok).
	- `go test ./db ./handlers -run "TestPrepareFacturacionDocumentoLegal|TestEmpresaDocumentoFacturacionUpsertAndGet|TestEmpresaFacturacionTransaccional" -count=1` (ok).

## 2026-04-04
- Punto 9 (modulo de compras) - avance funcional con endpoint y vista dedicados para ciclo documental.
	- `backend/db/documentos_transaccionales.go` agrega:
		- `ListEmpresaDocumentosCompraByEmpresa`.
		- `SetEmpresaDocumentoCompraEstadoByCodigo`.
	- `backend/handlers/compras.go` (nuevo) implementa `GET/POST/PUT/DELETE /api/empresa/compras/documentos` con acciones documentales (`crear`, `emitir_orden`, `recepcionar_compra`, `contabilizar_compra`) y activar/desactivar.
	- `backend/main.go` registra la ruta protegida `/api/empresa/compras/documentos`.
	- `web/administrar_empresa/compras.html` (nuevo) incorpora interfaz dedicada de compras para crear, consultar y transicionar documentos.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran el acceso de menu `Compras` con control por permisos de modulo.
	- Cobertura agregada en:
		- `backend/db/documentos_transaccionales_test.go`.
		- `backend/handlers/compras_documentos_test.go` (nuevo).
- Validacion tecnica:
	- `gofmt -w handlers/compras.go handlers/compras_documentos_test.go main.go db/documentos_transaccionales.go db/documentos_transaccionales_test.go` (ok).
	- `go test ./db -run "TestEmpresaDocumentoCompraListAndSetEstadoByCodigo" -count=1` (ok).
	- `go test ./handlers -run "TestEmpresaComprasDocumentos" -count=1` (ok).
	- `go test ./db ./handlers -run "TestEmpresaDocumentoCompraListAndSetEstadoByCodigo|TestEmpresaComprasDocumentos" -count=1` (ok).
	- `go test ./... -run "TestEmpresaComprasDocumentos|TestEmpresaDocumentoCompraListAndSetEstadoByCodigo" -count=1` (ok).

## 2026-04-04
- Punto 8 (facturacion electronica) - avance funcional de emision legal y cumplimiento normativo inicial.
	- `backend/db/facturacion_electronica.go` agrega `PrepareFacturacionDocumentoLegal` para validar configuracion legal, vigencia de resolucion y rango de consecutivos por empresa/pais antes de emitir.
	- `backend/db/documentos_transaccionales.go` amplia `empresa_facturacion_documentos` con metadata legal persistida: `numero_legal`, `codigo_validacion`, `pais_codigo`, `ambiente_fe`.
	- `backend/handlers/facturacion_electronica.go` endurece `action=emitir` con rechazo `422` cuando no hay cumplimiento normativo y devuelve bloque `cumplimiento_normativo` en emisiones exitosas.
	- `web/administrar_empresa/facturacion_electronica.html` incorpora bloque operativo para `emitir`, `anular` y `nota_credito`, con visualizacion del resultado legal.
	- Cobertura extendida en:
		- `backend/db/documentos_transaccionales_test.go`.
		- `backend/handlers/eventos_contables_modulos_test.go`.
- Validacion tecnica:
	- `go test ./db -run "TestEmpresaDocumentoFacturacionUpsertAndGet" -count=1` (ok).
	- `go test ./handlers -run "TestEmpresaFacturacionTransaccionalEmiteEventosContables|TestEmpresaFacturacionTransaccionalEmitirRechazaSinCumplimientoLegal|TestEmpresaFacturacionTransaccionalRechazaTransicionInvalida" -count=1` (ok).
	- `go test ./db ./handlers -count=1` (ok).

## 2026-04-04
- Punto 7 (gestion de proveedores) - avance funcional de catalogo, precios y condiciones comerciales.
	- `backend/db/productos.go` amplia el modelo `Proveedor` y su migracion segura con campos:
		- `catalogo_referencia`,
		- `precio_base_referencial`,
		- `descuento_porcentaje`,
		- `plazo_pago_dias`,
		- `condicion_entrega`.
	- `backend/handlers/productos.go` agrega validacion HTTP de rango para los nuevos campos en `POST/PUT /api/empresa/proveedores` y enriquece metadata de eventos contables de compras.
	- `web/administrar_empresa/administrar_productos.html` amplia el formulario y la tabla de proveedores para gestionar y visualizar datos comerciales.
	- Cobertura nueva/extendida en:
		- `backend/db/productos_categorias_test.go`.
		- `backend/handlers/eventos_contables_modulos_test.go`.
- Validacion tecnica:
	- `gofmt -w db/productos.go db/productos_categorias_test.go handlers/productos.go handlers/eventos_contables_modulos_test.go` (ok).
	- `go test ./db -run "TestProveedorCRUDIncluyeCatalogoPreciosYCondiciones" -count=1` (ok).
	- `go test ./handlers -run "TestEmpresaProveedoresEmiteEventoContableCompras|TestEmpresaProveedoresRechazaCamposComercialesInvalidos" -count=1` (ok).
	- `get_errors` sobre backend/frontend modificado (ok).

## 2026-04-04
- Punto 6 (gestion de clientes) - avance funcional de perfil, historial y segmentacion.
	- `backend/db/clientes.go` agrega contratos analiticos (`ClientePerfilComercial`, `ClienteCompraHistorial`, `ClienteSegmentacionResumen`) y funciones de consulta por cliente/empresa.
	- `backend/handlers/clientes.go` amplia `GET /api/empresa/clientes` con `action=perfil`, `action=historial`, `action=segmentacion|segmentos`.
	- `web/administrar_empresa/administrar_clientes.html` agrega paneles de segmentacion y de perfil/historial por cliente con accion `Perfil`.
	- Cobertura nueva en:
		- `backend/db/clientes_test.go`.
		- `backend/handlers/clientes_test.go`.
- Validacion tecnica:
	- `gofmt -w db/clientes.go db/clientes_test.go handlers/clientes.go handlers/clientes_test.go` (ok).
	- `go test ./db -run "TestGetClientePerfilComercialByEmpresaAndHistorial|TestGetClientePerfilComercialByEmpresaSinComprasSegmentoNuevo" -count=1` (ok).
	- `go test ./handlers -run "TestEmpresaClientesHandlerPerfilHistorialSegmentacion" -count=1` (ok).
	- `get_errors` sobre backend/frontend modificado (ok).

## 2026-04-04
- Punto 5 (control de inventarios) ÃƒÂ¢Ã¢â€šÂ¬? continuidad preventiva-compras ciclo documental desde reposicion.
	- `backend/db/productos.go` agrega `InventarioPlanReposicionOrdenEstadoActualizado` y `ActualizarEstadoOrdenCompraDesdeReposicion` para transiciones `recepcionar_compra` y `contabilizar_compra`.
	- `backend/handlers/productos.go` agrega endpoint `POST /api/empresa/compras/plan_reposicion/actualizar_estado`.
	- `backend/main.go` registra `/api/empresa/compras/plan_reposicion/actualizar_estado` bajo permisos de compras.
	- `web/administrar_empresa/administrar_productos.html` ampl?a el flujo a `fases 10-12` con acciones `Recepcionar orden` y `Contabilizar orden` y contexto de estado de OC.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`.
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./db -run "TestEmitirOrdenCompraDesdePlanReposicionBorradorPersistDoc|TestActualizarEstadoOrdenCompraDesdeReposicionCiclo"` (ok).
	- `go test ./handlers -run "TestEmpresaComprasPlanReposicionEmitirOrdenHandlerEmiteDocumento|TestEmpresaComprasPlanReposicionActualizarEstadoHandlerGestionaCiclo"` (ok).
	- `get_errors` sobre backend/frontend modificado (ok).

## 2026-04-04
- Punto 5 (control de inventarios) ÃƒÂ¢Ã¢â€šÂ¬? continuidad preventiva-compras emitible desde borrador.
	- `backend/db/productos.go` agrega `InventarioPlanReposicionOrdenEmitida` y `EmitirOrdenCompraDesdePlanReposicionBorrador` para emitir OC desde el borrador y persistirla en documentos de compras.
	- `backend/handlers/productos.go` agrega endpoint `POST /api/empresa/compras/plan_reposicion/emitir_orden`.
	- `backend/main.go` registra `/api/empresa/compras/plan_reposicion/emitir_orden` bajo permisos de compras.
	- `web/administrar_empresa/administrar_productos.html` agrega accion `Emitir orden` en el bloque de borrador (fase 10).
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`.
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./handlers ./db -count=1` (ok).
	- `get_errors` sobre backend/frontend modificado (ok).

## 2026-04-04
- Punto 5 (control de inventarios) ÃƒÂ¢Ã¢â€šÂ¬? continuidad preventiva-compras ordenable por proveedor.
	- `backend/db/productos.go` agrega `InventarioPlanReposicionBorradorItem`, `InventarioPlanReposicionBorradorCompra` y `GetInventarioPlanReposicionBorradorByEmpresa` para generar borradores de orden por proveedor con detalle y totales.
	- `backend/handlers/productos.go` agrega endpoint `GET /api/empresa/inventario/plan_reposicion_borrador`.
	- `backend/main.go` registra `/api/empresa/inventario/plan_reposicion_borrador` bajo permisos de inventario.
	- `web/administrar_empresa/administrar_productos.html` agrega bloque `Borrador de orden de compra por proveedor (fase 10)` y accion `Borrador OC` desde consolidado fase 9.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`.
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./handlers ./db -count=1` (ok).
	- `get_errors` sobre backend/frontend modificado (ok).

## 2026-04-04
- Punto 5 (control de inventarios) ÃƒÂ¢Ã¢â€šÂ¬? continuidad preventiva-compras consolidada por proveedor.
	- `backend/db/productos.go` agrega `InventarioPlanReposicionProveedorResumen` y `GetInventarioPlanReposicionResumenByEmpresa` para consolidar compra preventiva por proveedor.
	- `backend/handlers/productos.go` agrega endpoint `GET /api/empresa/inventario/plan_reposicion_resumen`.
	- `backend/main.go` registra `/api/empresa/inventario/plan_reposicion_resumen` bajo permisos de inventario.
	- `web/administrar_empresa/administrar_productos.html` agrega tabla `Consolidado de compra por proveedor (fase 9)` y filtro de items del plan por proveedor.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`.
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./handlers ./db -count=1` (ok).
	- `get_errors` sobre `web/administrar_empresa/administrar_productos.html` (ok).

## 2026-04-04
- Punto 5 (control de inventarios) ÃƒÂ¢Ã¢â€šÂ¬? continuidad preventiva-compras con plan de reposicion por proveedor.
	- `backend/db/productos.go` agrega `InventarioPlanReposicionItem` y `GetInventarioPlanReposicionByEmpresa` para consolidar sugerencias por proveedor con costo estimado.
	- `backend/handlers/productos.go` agrega endpoint `GET /api/empresa/inventario/plan_reposicion` con validaciones operativas.
	- `backend/main.go` registra `/api/empresa/inventario/plan_reposicion` bajo permisos de inventario.
	- `web/administrar_empresa/administrar_productos.html` agrega tabla `Plan de reposicion por proveedor (fase 8)` con resumen de costo estimado y accion `Preparar`.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`.
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./handlers ./db -count=1` (ok).
	- `get_errors` sobre `web/administrar_empresa/administrar_productos.html` (ok).

## 2026-04-04
- Punto 5 (control de inventarios) ÃƒÂ¢Ã¢â€šÂ¬? continuidad preventiva con proyeccion de quiebre.
	- `backend/db/productos.go` agrega `InventarioProyeccionQuiebre` y `GetInventarioProyeccionQuiebreByEmpresa` para estimar consumo diario, cobertura y sugerido de reposicion por producto/bodega.
	- `backend/handlers/productos.go` agrega endpoint `GET /api/empresa/inventario/proyeccion_quiebre` con validacion de `dias_ventana`, `bodega_id`, `limit` y `offset`.
	- `backend/main.go` registra `/api/empresa/inventario/proyeccion_quiebre` bajo permisos de inventario.
	- `web/administrar_empresa/administrar_productos.html` agrega tabla `Proyeccion de quiebre (preventiva)` y accion `Preparar` para reposicion preventiva guiada.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`.
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./handlers ./db -count=1` (ok).
	- `get_errors` sobre `web/administrar_empresa/administrar_productos.html` (ok).

## 2026-04-04
- Punto 5 (control de inventarios) ÃƒÂ¢Ã¢â€šÂ¬? continuidad operativa-analitica con balance por bodega.
	- `backend/db/productos.go` agrega `InventarioBalanceBodega` y `GetInventarioBalanceBodegasByEmpresa` para consolidar entradas/salidas/traslados/neto por bodega en rango.
	- `backend/handlers/productos.go` agrega endpoint `GET /api/empresa/inventario/balance_bodegas` con validacion de fechas y filtros por bodega/rango.
	- `backend/main.go` registra `/api/empresa/inventario/balance_bodegas` bajo permisos de inventario.
	- `web/administrar_empresa/administrar_productos.html` agrega tabla `Balance por bodega` y contexto de neto acumulado sincronizado con filtros del kardex.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`.
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./handlers ./db -count=1` (ok).
	- `get_errors` sobre `web/administrar_empresa/administrar_productos.html` (ok).

## 2026-04-04
- Punto 5 (control de inventarios) ÃƒÂ¢Ã¢â€šÂ¬? continuidad analitica con tendencia diaria.
	- `backend/db/productos.go` agrega `InventarioTendenciaDia` y `GetInventarioTendenciaByEmpresa` para serie diaria por empresa con filtros por bodega/rango.
	- `backend/handlers/productos.go` agrega endpoint `GET /api/empresa/inventario/tendencia` con validacion de fechas y ventana por `dias`.
	- `backend/main.go` registra `/api/empresa/inventario/tendencia` bajo permisos de inventario.
	- `web/administrar_empresa/administrar_productos.html` agrega tabla `Tendencia diaria inventario` y contexto de neto acumulado/eventos sincronizado con filtros del kardex.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`.
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./handlers ./db -count=1` (ok).
	- `get_errors` sobre `web/administrar_empresa/administrar_productos.html` (ok).

## 2026-04-04
- Punto 5 (control de inventarios) ÃƒÂ¢Ã¢â€šÂ¬? continuidad operacional en panel de productos.
	- `web/administrar_empresa/administrar_productos.html` agrega:
		- bloque `Top productos críticos (d?ficit)` alimentado desde alertas de inventario,
		- priorizaci?n de críticos por `sin_stock` y mayor d?ficit,
		- acci?n `Preparar reposici?n` para precargar ajuste de inventario con producto, bodega y cantidad sugerida.
- Validacion tecnica:
	- `get_errors` sobre `web/administrar_empresa/administrar_productos.html` (ok).

## 2026-04-04
- Punto 5 (control de inventarios) ÃƒÂ¢Ã¢â€šÂ¬? continuidad KPI operativo en panel de productos.
	- `backend/db/productos.go` agrega `InventarioResumen` y `GetInventarioResumenByEmpresa` para consolidar existencias, alertas y movimientos por rango.
	- `backend/handlers/productos.go` agrega endpoint `GET /api/empresa/inventario/resumen` con validacion de fechas `YYYY-MM-DD`.
	- `backend/main.go` registra `/api/empresa/inventario/resumen` bajo permisos de inventario.
	- `web/administrar_empresa/administrar_productos.html` agrega KPI visibles de inventario e integra consumo del resumen segun rango del kardex.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`,
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./handlers ./db -count=1` (ok).
	- `get_errors` sobre `web/administrar_empresa/administrar_productos.html` (ok).

## 2026-04-04
- Punto 5 (control de inventarios) ÃƒÂ¢Ã¢â€šÂ¬? continuidad UI operativa en panel de productos.
	- `web/administrar_empresa/administrar_productos.html` agrega:
		- filtro por bodega para alertas de quiebre,
		- filtros de kardex por bodega, tipo y rango de fechas,
		- acciones `Filtrar` y `Limpiar` en ambos bloques de consulta.
	- Se actualiza documentacion asociada en plan maestro y estructura tecnica.
- Validacion tecnica:
	- `get_errors` sobre `web/administrar_empresa/administrar_productos.html` (ok).

## 2026-04-04
- Punto 5 (control de inventarios) ÃƒÂ¢Ã¢â€šÂ¬? inicio tecnico: kardex operativo + reglas de stock + alertas de quiebre por bodega.
	- `backend/db/productos.go`:
		- valida `stock_minimo/stock_maximo` en creacion y edicion de productos,
		- agrega `GetAlertasQuiebreByEmpresa`,
		- ampl?a `GetMovimientosByEmpresa` con filtros `bodega_id`, `tipo`, `desde`, `hasta`.
	- `backend/handlers/productos.go`:
		- nuevo endpoint `GET /api/empresa/inventario/alertas`,
		- compatibilidad `action=alertas|alertas_quiebre|quiebre` en existencias,
		- filtros de kardex + validacion de fechas `YYYY-MM-DD` en movimientos.
	- `backend/main.go` registra `/api/empresa/inventario/alertas` bajo permisos de inventario.
	- `web/administrar_empresa/administrar_productos.html` agrega tabla de alertas de quiebre por bodega.
	- `documentos/descripcion_del_proyecto` actualiza la descripcion de inventario con alertas de quiebre y kardex filtrable.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`,
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `runTests` en archivos de prueba modificados (ok).
	- `go test ./handlers ./db -count=1` en `backend` (ok).

## 2026-04-04
- Punto 3 (permisos y seguridad) ÃƒÂ¢Ã¢â€šÂ¬? continuidad operativa: catalogo frontend por rol + regresion endpoints sin wrapper.
	- `web/js/administrar_empresa.js` agrega catalogo de permisos por enlace y aplica ocultamiento de opciones no autorizadas segun rol autenticado (`GET /me`).
	- Se agrega fallback de navegacion en iframe cuando la ultima pagina guardada no es visible para el rol actual.
	- `backend/handlers/auth_users_carritos_test.go` agrega regresiones de alcance por `empresa_id` para:
		- `POST /api/empresa/usuarios/login`.
		- `POST /api/empresa/usuarios/establecer_password`.
	- `backend/handlers/chat_con_inteligencia_artificial_controller_test.go` agrega regresion de alcance por cuenta Google en `ModelosHandler`.
	- Se actualiza documentacion tecnica en:
		- `documentos/diagramas/diagrama_roles_permisos.md`.
		- `documentos/diagramas/estructura_del_codigo.md`.
- Validacion tecnica:
	- `runTests` sobre `backend/handlers/auth_users_carritos_test.go` y `backend/handlers/chat_con_inteligencia_artificial_controller_test.go`.
	- resultado: 14 pruebas aprobadas, 0 fallidas.
	- `get_errors` sobre `web/js/administrar_empresa.js`: sin errores.

## 2026-04-04
- Punto 3 (permisos y seguridad) ÃƒÂ¢Ã¢â€šÂ¬? consolidacion documental endpoint/rol y checklist UAT:
	- `documentos/matriz_roles_permisos_pos_multiempresa.md` agrega matriz final endpoint/rol alineada con wrappers reales y reglas por accion.
	- Se documentan endpoints fuera de wrapper con control alterno por handler/cuenta Google.
	- Se agrega checklist UAT de punto 3 con evidencia automatizada.
	- `documentos/plan_maestro_pos_multiempresa_14_puntos.md` agrega seccion de consolidacion con estado operativo y pendientes de cierre total.
- Validacion tecnica:
	- `runTests` sobre `backend/handlers/empresa_permisos_test.go` y `backend/handlers/auditoria_empresa_test.go`.
	- resultado: 25 pruebas aprobadas, 0 fallidas.

## 2026-04-04
- Ajuste editorial de consistencia documental (plan maestro):
	- `documentos/plan_maestro_pos_multiempresa_14_puntos.md` corrige `Backlog inmediato` para reflejar cierre real de Punto 1 y Punto 2.
	- El backlog siguiente queda enfocado en Punto 3 (permisos y seguridad) y Punto 5 (control de inventarios).
- Validacion tecnica:
	- cambio documental (sin ejecucion de pruebas automatizadas).

## 2026-04-04
- Punto 1 + Punto 2 (plan maestro) ÃƒÂ¢Ã¢â€šÂ¬? cierre de backlog inmediato con formalizacion tecnica documental.
	- `documentos/matriz_kpi_pos_multiempresa.md` se actualiza a formato formal con:
		- formula implementada por KPI,
		- endpoint canonico de lectura/exportacion,
		- tablas fuente reales por metrica.
	- Se crea `documentos/matriz_entidades_multiempresa_aislamiento.md` con matriz de aislamiento por endpoint:
		- llave primaria `empresa_id`,
		- llaves secundarias por recurso,
		- mecanismo de control de alcance (middleware o validacion interna).
	- `documentos/plan_maestro_pos_multiempresa_14_puntos.md` marca Punto 1 y Punto 2 como `completado`.
- Validacion tecnica:
	- cambio documental (sin ejecucion de pruebas automatizadas).

## 2026-04-04
- Punto 11 (reportes financieros) ÃƒÂ¢Ã¢â€šÂ¬? continuidad de backlog inmediato: exportacion unificada del tablero por rango.
	- `backend/handlers/finanzas.go` agrega `action=tablero_export` en `GET /api/empresa/finanzas/movimientos` con:
		- `format=json` para payload unificado del tablero,
		- `format=csv` para matriz unificada por bloque/metrica/valor.
	- La exportacion integra bloques `estado_resultados` y `balance_general` junto con KPI operativos/financieros/contables.
	- `web/administrar_empresa/reportes.html` incorpora botones:
		- `Exportar tablero CSV`,
		- `Exportar tablero JSON`.
	- `backend/handlers/eventos_contables_modulos_test.go` agrega `TestEmpresaFinanzasTableroResumenExportHandler`.
- Validacion tecnica:
	- `go test ./handlers -run "TestEmpresaFinanzasTableroResumenHandler|TestEmpresaFinanzasTableroResumenExportHandler|TestEmpresaFinanzasAsientosContablesHandlerConciliacionPeriodo" -count=1` (ok).
	- `go test ./handlers -count=1` (ok).
	- `go test ./db -count=1` (ok).

## 2026-04-04
- Punto 10 (modulo contable integrado) ÃƒÂ¢Ã¢â€šÂ¬? continuidad de backlog inmediato: vista de conciliacion por periodo (eventos vs asientos).
	- `backend/db/eventos_contables.go` agrega modelos y funcion `GetEmpresaConciliacionContablePorPeriodo` para consolidar por periodo:
		- eventos totales/procesados/pendientes/con error,
		- asientos generados,
		- desfase de conteo y desfase de monto,
		- estado de conciliacion por periodo.
	- `backend/handlers/finanzas.go` agrega `GET /api/empresa/finanzas/asientos_contablesaction=conciliacion_periodo|conciliacion`.
	- `web/administrar_empresa/finanzas.html` incorpora vista de conciliacion con filtros, KPIs y tabla comparativa por periodo.
	- `backend/db/eventos_contables_test.go` agrega prueba de conciliacion por periodo.
	- `backend/handlers/eventos_contables_modulos_test.go` agrega prueba del endpoint de conciliacion.
- Validacion tecnica:
	- `go test ./db -run "EventosContables|ConPolitica|Conciliacion" -count=1` (ok).
	- `go test ./handlers -run "AsientosContablesHandler|ConciliacionPeriodo" -count=1` (ok).
	- `go test ./db -count=1` (ok).
	- `go test ./handlers -count=1` (ok).

## 2026-04-04
- Punto 10 (modulo contable integrado) ÃƒÂ¢Ã¢â€šÂ¬? continuidad de backlog inmediato: ejecucion automatica por lotes de asientos.
	- `backend/db/eventos_contables.go` agrega:
		- `ProcessEmpresaEventosContablesPendientesConPolitica` con soporte de `max_reintentos`,
		- `RunEmpresaAsientosContablesWorkerCycle`,
		- `StartEmpresaAsientosContablesWorker`.
	- `backend/main.go` integra worker automatico de asientos con politica configurable por entorno:
		- `ASIENTOS_WORKER_INTERVAL_MINUTES`,
		- `ASIENTOS_WORKER_BATCH_SIZE`,
		- `ASIENTOS_WORKER_MAX_RETRIES`.
	- `backend/handlers/finanzas.go` permite `max_reintentos` opcional en proceso manual de `/api/empresa/finanzas/asientos_contablesaction=procesar_asientos`.
	- `backend/db/eventos_contables_test.go` agrega prueba de politica de reintentos.
	- `backend/handlers/eventos_contables_modulos_test.go` agrega validacion `400` para `max_reintentos` invalido y cobertura del parametro.
- Validacion tecnica:
	- `go test ./db -run "EventosContables|ConPolitica|Asientos" -count=1` (ok).
	- `go test ./handlers -run "AsientosContablesHandler|FinanzasAsientos" -count=1` (ok).
	- `go test ./handlers -count=1` (ok).
	- `go test ./db -count=1` (ok).

## 2026-04-04
- Punto 15 (auditoria por empresa) ÃƒÂ¢Ã¢â€šÂ¬? continuacion de backlog inmediato 1 y 2:
	- `backend/db/auditoria_empresa.go` agrega filtros avanzados de consulta por `recurso_id` y `codigo_http` en `ListEmpresaAuditoriaEventos`.
	- `backend/handlers/auditoria_empresa.go` valida y expone nuevos filtros en `GET /api/empresa/auditoria/eventos`:
		- `recurso_id`.
		- `codigo_http`.
	- `web/administrar_empresa/auditoria.html` incorpora:
		- filtros avanzados por `codigo_http` y `recurso_id`,
		- exportacion de resultados filtrados a `CSV` y `JSON`.
	- `backend/db/auditoria_empresa_test.go` fortalece cobertura de listado con filtros avanzados.
	- `backend/handlers/auditoria_empresa_test.go` agrega `TestEmpresaAuditoriaEventosHandlerFiltrosAvanzados` para contrato HTTP y validacion de parametros invalidos.
- Validacion tecnica:
	- `go test ./db -run "Auditoria" -count=1` (ok).
	- `go test ./handlers -run "Auditoria" -count=1` (ok).
	- `go test ./handlers -count=1` (ok).
	- `go test ./db -count=1` (ok).

## 2026-04-04
- Punto 15 (auditoria por empresa) ÃƒÂ¢Ã¢â€šÂ¬? continuacion de backlog 1, 2 y 3:
	- `backend/handlers/empresa_permisos.go` refuerza clasificacion de acciones criticas en `ventas`, `compras` y `facturacion` (alias operativos de aprobacion/eliminacion).
	- `backend/handlers/auditoria_empresa.go` amplia metadata de trazabilidad para recursos de ventas/compras/facturacion (`carrito_id`, `proveedor_id`, `entidad_id`, `documento_codigo`).
	- `backend/handlers/auditoria_empresa_test.go` agrega pruebas de registro automatico de auditoria en acciones criticas de:
		- ventas (`action=cerrar`),
		- compras (`action=emitir_orden`),
		- facturacion (`action=emitir`).
	- `web/administrar_empresa/auditoria.html` agrega vista de consulta filtrable y retencion manual para auditoria por empresa.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` agregan acceso del menu lateral a la nueva vista `Auditoria`.
	- `backend/db/auditoria_empresa.go` agrega:
		- purga automatica por expiracion (`PurgeExpiredEmpresaAuditoriaEventos`),
		- worker programado (`StartEmpresaAuditoriaRetentionWorker`),
		- calculo de `fecha_expiracion` alineado a `fecha_evento` cuando se provee.
	- `backend/main.go` arranca worker de retencion automatica de auditoria (intervalo 12h).
	- `backend/db/auditoria_empresa_test.go` agrega prueba de purga automatica por expiracion.
- Validacion tecnica:
	- `go test ./handlers -run "Auditoria|WithEmpresa(Ventas|Compras|Facturacion|Finanzas)Permissions" -count=1` (ok).
	- `go test ./db -run "Auditoria" -count=1` (ok).
	- `go test ./handlers -count=1` (ok).
	- `go test ./db -count=1` (ok).

## 2026-04-04
- Punto 15 (auditoria por empresa) ÃƒÂ¢Ã¢â€šÂ¬? implementacion base minima:
	- `backend/db/auditoria_empresa.go` agrega tabla `empresa_auditoria_eventos`, filtros de consulta y purga por retencion.
	- `backend/handlers/auditoria_empresa.go` agrega endpoint protegido:
		- `GET /api/empresa/auditoria/eventos`.
		- `PUT/POST /api/empresa/auditoria/eventosaction=retener|purgar`.
	- `backend/handlers/empresa_permisos.go` integra registro automatico no bloqueante para acciones criticas (`C/U/D/A`).
	- `backend/main.go` integra `EnsureEmpresaAuditoriaSchema`, migracion `2026-04-04-011-auditoria-empresa` y ruta de auditoria.
	- Pruebas nuevas: `backend/db/auditoria_empresa_test.go` y `backend/handlers/auditoria_empresa_test.go`.
- Validacion tecnica:
	- `go test ./db -run "Auditoria|EventosContables|ReportesTableroResumen" -count=1` (ok).
	- `go test ./handlers -run "Auditoria|AsientosContables|WithEmpresaFinanzasPermissions" -count=1` (ok).
	- `go test ./handlers -count=1` (ok).
	- `go test ./db -count=1` (ok).

## 2026-04-04
- Plan maestro POS multiempresa:
	- `documentos/plan_maestro_pos_multiempresa_14_puntos.md` se actualiza de 14 a 15 puntos.
	- Se incorpora el nuevo `Punto 15: Modulo de auditoria por empresa` con alcance, entregables iniciales, backlog y criterio de avance.
	- `documentos/descripcion_del_proyecto` se alinea para referenciar el plan de 15 puntos.
- Validacion tecnica:
	- cambio documental (sin cambios de codigo ni ejecucion de pruebas adicionales).

## 2026-04-04
- Punto 10 + Punto 11 (continuacion de backlog 1 y 2):
	- `backend/db/eventos_contables.go` ampl?a `empresa_eventos_contables` con metadatos de procesamiento (`intentos_procesamiento`, `fecha_ultimo_intento`, `error_procesamiento`, `asiento_contable_id`) y crea tabla canonica `empresa_asientos_contables` con hash de idempotencia.
	- `backend/handlers/finanzas.go` agrega `EmpresaFinanzasAsientosContablesHandler`:
		- `GET /api/empresa/finanzas/asientos_contables` para consulta,
		- `POST/PUT action=procesar_asientos|procesar` para procesamiento manual por lote.
	- `backend/handlers/empresa_permisos.go` clasifica `action=procesar_asientos` como accion de aprobacion en finanzas.
	- `backend/main.go` publica `/api/empresa/finanzas/asientos_contables` y registra migracion `2026-04-04-010-asientos-canonicos`.
	- `backend/db/finanzas.go` integra en el tablero los bloques `estado_resultados` y `balance_general`, junto con KPI contables de asientos (`asientos_generados`, `asientos_monto_total`).
	- `web/administrar_empresa/reportes.html` incorpora visualizacion de utilidad operacional, activos/pasivos/patrimonio, resultado del ejercicio y cuadre.
	- `web/administrar_empresa/finanzas.html` a?ade accion manual `Procesar eventos contables`.
	- Cobertura de pruebas nueva/extendida en `backend/db/eventos_contables_test.go`, `backend/db/finanzas_test.go`, `backend/handlers/eventos_contables_modulos_test.go` y `backend/handlers/empresa_permisos_test.go`.
- Validacion tecnica:
	- `go test ./db -run "EventosContables|ReportesTableroResumen" -count=1` (ok).
	- `go test ./handlers -run "AsientosContables|TableroResumen|WithEmpresaFinanzasPermissions" -count=1` (ok).
	- `go test ./handlers -count=1` (ok).
	- `go test ./db -count=1` (ok).

## 2026-04-04
- Punto 12 + Punto 10 (continuacion de backlog 1 y 2):
	- `backend/handlers/empresa_permisos_test.go` agrega pruebas UAT por rol para `PUT action=aprobar` en `cierres_caja`:
		- rechazo para `cajero`,
		- rechazo para `supervisor_sucursal`,
		- aprobacion permitida para `admin_empresa`.
	- `documentos/matriz_roles_permisos_pos_multiempresa.md` agrega matriz UAT de cierres con casos por rol y transiciones de estado.
	- `documentos/plan_maestro_pos_multiempresa_14_puntos.md` define estrategia de procesamiento de asientos sobre `empresa_eventos_contables` y referencias canonicas documentales (`entidad_id`).
- Validacion tecnica:
	- `go test ./handlers -run "TestWithEmpresaFinanzasPermissions(DeniesCajeroAprobarCierreCaja|DeniesSupervisorAprobarCierreCaja|AllowsAdminAprobarCierreCaja)" -count=1` (ok).

## 2026-04-04
- Punto 12 (cierres de caja) ÃƒÂ¢Ã¢â€šÂ¬? continuacion con UI operativa en panel empresa:
	- `web/administrar_empresa/finanzas.html` integra modulo visual de cierres de caja por sucursal con:
		- formulario de apertura/actualizacion,
		- calculo de `caja_teorica` y `diferencia_caja`,
		- filtros por sucursal/caja/estado/fecha,
		- tabla de acciones (`cerrar`, `reabrir`, `aprobar`, `anular`, `activar/desactivar`, `eliminar`).
	- La vista queda conectada al endpoint existente `GET/POST/PUT/DELETE /api/empresa/finanzas/cierres_caja`.
- Validacion tecnica:
	- `get_errors` sobre `web/administrar_empresa/finanzas.html` (ok).

## 2026-04-04
- Punto 12 (cierres de caja) ÃƒÂ¢Ã¢â€šÂ¬? inicio de flujo operativo por sucursal:
	- `backend/db/finanzas.go` agrega `empresa_cierres_caja` con soporte de apertura, arqueo, cierre, reapertura, aprobacion y anulacion.
	- `backend/handlers/finanzas.go` incorpora `GET/POST/PUT/DELETE /api/empresa/finanzas/cierres_caja`.
	- `backend/main.go` publica la ruta de cierres de caja y registra migracion `2026-04-04-009-cierres-caja`.
	- `backend/handlers/empresa_permisos.go` trata `action=aprobar` en finanzas como accion `A`.
	- Pruebas nuevas:
		- `backend/db/finanzas_test.go`: `TestEmpresaCierresCajaFlow`.
		- `backend/handlers/eventos_contables_modulos_test.go`: `TestEmpresaFinanzasCierresCajaHandler`.
- Validacion tecnica:
	- `go test ./db -run "TestEmpresaCierresCajaFlow|TestGetEmpresaReportesTableroResumen|TestEmpresaFinanzas" -count=1` (ok).
	- `go test ./handlers -run "TestEmpresaFinanzasCierresCajaHandler|TestEmpresaFinanzasTableroResumenHandler|TestEmpresaFinanzasEmiteEventosContables" -count=1` (ok).
	- `go test ./auth ./db ./handlers ./metrics ./utils -count=1` (ok).

## 2026-04-04
- Punto 11 (reportes financieros) ÃƒÂ¢Ã¢â€šÂ¬? inicio de tablero minimo financiero-operativo:
	- `backend/db/finanzas.go` agrega `GetEmpresaReportesTableroResumen` con KPI consolidados:
		- operativos (ventas/ticket/clientes/productos/compras),
		- financieros (ingresos/egresos/balance/periodos),
		- contables (eventos y documentos activos).
	- `backend/handlers/finanzas.go` extiende `GET /api/empresa/finanzas/movimientos` con `action=tablero|dashboard|resumen_kpi`.
	- `web/administrar_empresa/reportes.html` incorpora KPI financieros y contables en la misma vista de reportes.
	- Pruebas nuevas:
		- `backend/db/finanzas_test.go`: `TestGetEmpresaReportesTableroResumen`.
		- `backend/handlers/eventos_contables_modulos_test.go`: `TestEmpresaFinanzasTableroResumenHandler`.
- Validacion tecnica:
	- `go test ./db -run "TestGetEmpresaReportesTableroResumen|TestEmpresaFinanzas" -count=1` (ok).
	- `go test ./handlers -run "TestEmpresaFinanzasTableroResumenHandler|TestEmpresaFinanzasEmiteEventosContables" -count=1` (ok).
	- `go test ./... -count=1` en `backend` (ok).

## 2026-04-04
- Punto 8 + Punto 9 + Punto 10 (facturacion/compras) ÃƒÂ¢Ã¢â€šÂ¬? persistencia canonica de documentos transaccionales para `entidad_id`:
	- Se agrega `backend/db/documentos_transaccionales.go` con tablas y APIs de upsert/lectura para:
		- `empresa_facturacion_documentos`.
		- `empresa_compras_documentos`.
	- `backend/main.go` integra:
		- `EnsureEmpresaDocumentosTransaccionalesSchema`.
		- migracion `2026-04-04-008-documentos-transaccionales`.
	- `backend/handlers/facturacion_electronica.go` y `backend/handlers/productos.go` ahora:
		- consultan estado documental persistido por `documento_codigo`,
		- aplican transicion de ciclo sobre estado canonico,
		- persisten el nuevo estado en tabla de negocio,
		- emiten evento contable usando `entidad_id` canonico (ID persistido en tabla documental).
	- Se agrega `backend/db/documentos_transaccionales_test.go` y se amplian aserciones en `backend/handlers/eventos_contables_modulos_test.go` para verificar estabilidad de `entidad_id` en el ciclo documental.
- Validacion tecnica:
	- `go test ./handlers -run "FacturacionTransaccionalEmiteEventosContables|ComprasTransaccionalEmiteEventosContables|FacturacionElectronicaEmiteEventoContable|ProveedoresEmiteEventoContableCompras|FinanzasEmiteEventosContables" -count=1` (ok).
	- `go test ./auth ./db ./handlers ./metrics ./utils` (ok).
	- `go test ./...` en `backend` (ok).

## 2026-04-04
- Punto 8 + Punto 9 + Punto 10 (facturacion/compras) ÃƒÂ¢Ã¢â€šÂ¬? estandarizacion de estados en ciclo documental transaccional:
	- Se agrega `backend/handlers/documentos_lifecycle.go` con reglas de transicion por accion y estado previo para facturacion/compras.
	- `backend/handlers/facturacion_electronica.go` ahora valida `estado_actual` en `emitir/anular/nota_credito`, devuelve `409` en conflictos y responde `estado_anterior`/`estado_nuevo` cuando la transicion es valida.
	- `backend/handlers/productos.go` (`EmpresaProveedoresHandler`) aplica validacion equivalente para `emitir_orden/recepcionar_compra/contabilizar_compra`.
	- `backend/handlers/eventos_contables_modulos_test.go` ampl?a cobertura con pruebas de transiciones invalidas para facturacion y compras.
- Validacion tecnica:
	- `runTests` sobre `backend/handlers/eventos_contables_modulos_test.go` (ok).
	- `go test ./...` en `backend` (ok).

## 2026-04-04
- Punto 8 + Punto 9 + Punto 10 (facturacion/compras) ÃƒÂ¢Ã¢â€šÂ¬? eventos transaccionales de factura y orden:
	- `backend/handlers/facturacion_electronica.go` agrega acciones transaccionales:
		- `action=emitir` -> `factura_emitida`.
		- `action=anular` -> `factura_anulada`.
		- `action=nota_credito|emitir_nota_credito` -> `nota_credito_emitida`.
	- `backend/handlers/productos.go` (`EmpresaProveedoresHandler`) agrega acciones transaccionales:
		- `action=emitir|emitir_orden` -> `orden_compra_emitida`.
		- `action=recepcionar|recepcionar_compra` -> `compra_recepcionada`.
		- `action=contabilizar|contabilizar_compra` -> `compra_contabilizada`.
	- `backend/handlers/empresa_permisos.go` ampl?a mapeo de acciones de permisos para compras/facturacion.
	- `backend/handlers/eventos_contables_modulos_test.go` agrega pruebas de emisiones transaccionales de factura/orden.
- Validacion tecnica:
	- `go test ./handlers -run "FacturacionTransaccionalEmiteEventosContables|ComprasTransaccionalEmiteEventosContables|FacturacionElectronicaEmiteEventoContable|ProveedoresEmiteEventoContableCompras|FinanzasEmiteEventosContables" -count=1` (ok).
	- `go test ./...` en `backend` (ok).

## 2026-04-04
- Punto 8 + Punto 9 + Punto 10 (facturacion/compras/finanzas) ÃƒÂ¢Ã¢â€šÂ¬? extension de emision de eventos contables por modulo:
	- Se agrega `backend/handlers/eventos_contables.go` para registro no bloqueante y reutilizable de eventos contables en handlers.
	- Se amplia `backend/db/eventos_contables.go` con eventos operativos de:
		- `facturacion`: `configuracion_facturacion_actualizada`.
		- `compras`: `proveedor_registrado`, `proveedor_actualizado`, `proveedor_activado`, `proveedor_desactivado`, `proveedor_eliminado`.
	- Se integra emision en:
		- `backend/handlers/facturacion_electronica.go`.
		- `backend/handlers/productos.go` (proveedores).
		- `backend/handlers/finanzas.go` (movimientos y periodos).
	- `backend/handlers/carritos_compras.go` migra a helper comun para consistencia del registro contable.
	- Se agregan pruebas en `backend/handlers/eventos_contables_modulos_test.go` para validar emision en facturación, compras y finanzas.
- Validacion tecnica:
	- `go test ./db -run "EventosContables" -count=1` (ok).
	- `go test ./handlers -run "FacturacionElectronicaEmiteEventoContable|ProveedoresEmiteEventoContableCompras|FinanzasEmiteEventosContables|CarritosCompraAndItemsFlow" -count=1` (ok).
	- `go test ./...` en `backend` (ok).

## 2026-04-04
- Punto 4 + Punto 10 (gestion de ventas + modulo contable integrado) ÃƒÂ¢Ã¢â€šÂ¬? contrato de eventos contables por modulo:
	- Se agrega `backend/db/eventos_contables.go` con contrato base de eventos para `ventas`, `facturacion`, `compras` y `finanzas`.
	- Se crea tabla `empresa_eventos_contables` en `pcs_empresas` para registrar trazabilidad contable por empresa (`modulo`, `evento`, `entidad`, `documento`, `periodo_contable`, `monto`, `payload_json`, `procesado`).
	- Se integra bootstrap en `backend/main.go`:
		- `EnsureEmpresaEventosContablesSchema`.
		- migracion `2026-04-04-007-eventos-contables`.
	- Se actualiza `backend/handlers/carritos_compras.go` para emitir eventos contables en transiciones de venta de carritos (`venta_sesion_activada`, `venta_activada`, `venta_suspendida`, `venta_cerrada`, `venta_reabierta`, `venta_pagada`).
	- Se agregan pruebas:
		- `backend/db/eventos_contables_test.go`.
		- `backend/handlers/auth_users_carritos_test.go` (validacion de emision de `venta_pagada`).
- Validacion tecnica:
	- `go test ./db -run "EventosContables|CarritoEstadoVentaLifecycle|Finanzas" -count=1` (ok).
	- `go test ./handlers -run "EmpresaCarritosCompra|CarritosCompraAndItemsFlow" -count=1` (ok).
	- `go test ./...` en `backend` (ok).

## 2026-04-04
- Punto 4 (gestion de ventas) ÃƒÂ¢Ã¢â€šÂ¬? formalizacion de transiciones del ciclo de venta en carritos:
	- `backend/handlers/carritos_compras.go` ahora valida transiciones por accion y estado actual del carrito.
	- Se agregan respuestas de control para integridad de flujo:
		- `404` para carrito inexistente,
		- `409` para transiciones no permitidas (doble pago, reabrir pagada, activar estacion pagada sin `reset_items=1`, etc.).
	- Se agregan pruebas en `backend/handlers/auth_users_carritos_test.go`:
		- `TestEmpresaCarritosCompraRejectsDoublePago`.
		- `TestEmpresaCarritosCompraRejectsReabrirVentaPagada`.
		- `TestEmpresaCarritosCompraRejectsActivarEstacionPagadaSinReset`.
- Validacion tecnica:
	- `go test ./handlers -run "Carritos|EmpresaCarritosCompra" -count=1` (ok).
	- `go test ./...` en `backend` (ok).

## 2026-04-04
- Cierre validado del punto 3 (permisos y seguridad) con pruebas de endpoints protegidos recien incorporados:
	- `backend/handlers/empresa_permisos_test.go` agrega:
		- `TestWithEmpresaInventarioPermissionsDeniesCajeroWriteGPS`.
		- `TestWithEmpresaVentasPermissionsAllowsCajeroChatAdjuntoMultipart`.
		- `TestWithEmpresaVentasPermissionsRejectsChatAdjuntoWithoutAuth`.
	- Se valida control por rol en GPS, extraccion de `empresa_id` en `multipart/form-data` para adjuntos de chat y rechazo `401` sin autenticacion.
- Inicio del punto 4 (gestion de ventas):
	- `backend/db/carritos_compras.go` incorpora `estado_venta` derivado en el modelo `CarritoCompra` para estandarizar ciclo de vida de venta:
		- `venta_abierta`,
		- `venta_cerrada`,
		- `venta_pagada`,
		- `venta_suspendida`.
	- `backend/handlers/carritos_compras.go` expone `estado_venta` en acciones operativas (`activar_estacion`, `pagar_estacion`, `activar/desactivar`, `cerrar/reabrir`).
	- Se amplian pruebas en:
		- `backend/handlers/auth_users_carritos_test.go`.
		- `backend/db/carritos_inventario_test.go`.
- Validacion tecnica de esta iteracion:
	- `runTests` sobre archivos de pruebas modificados (ok).
	- `go test ./...` en `backend` (ok).

## 2026-04-04
- Continuacion del punto 3 del plan maestro (permisos y seguridad) con cierre de rutas operativas pendientes:
	- `backend/handlers/empresa_permisos.go` agrega modulo `seguridad` y wrapper `WithEmpresaSeguridadPermissions`.
	- `backend/main.go` ampl?a middleware en rutas:
		- seguridad: `/api/empresa/usuarios`, `/api/empresa/configuracion_avanzada`, `/api/empresa/roles_de_usuario`.
		- inventario: `/api/empresa/productos/imagen`, `/api/empresa/ubicacion_gps/dispositivos`, `/api/empresa/ubicacion_gps/recorridos`.
		- colaboracion operativa (politica ventas): `/api/empresa/chat_tareas/conversaciones`, `/api/empresa/chat_tareas/participantes`, `/api/empresa/chat_tareas/mensajes`, `/api/empresa/chat_tareas/mensajes/adjunto`, `/api/empresa/chat_tareas/tareas`.
	- `backend/handlers/empresa_permisos_test.go` agrega cobertura para modulo seguridad:
		- `TestWithEmpresaSeguridadPermissionsDeniesSupervisorWrite`.
		- `TestWithEmpresaSeguridadPermissionsAllowsSupervisorRead`.
		- `TestWithEmpresaSeguridadPermissionsAllowsAdminApprove`.
	- Validacion tecnica: `go test ./handlers -run "WithEmpresa|ConsultarHandlerRejectsEmpresaFueraDeAlcance" -count=1` (ok) y `go test ./...` (ok).

## 2026-04-04
- Continuacion del punto 3 del plan maestro (permisos y seguridad):
	- `backend/handlers/empresa_permisos.go` ampl?a modulos de autorizacion para `clientes`, `compras` y `facturacion`.
	- Se agregan wrappers: `WithEmpresaClientesPermissions`, `WithEmpresaComprasPermissions`, `WithEmpresaFacturacionPermissions`.
	- `backend/main.go` aplica middleware en rutas: `/api/empresa/clientes`, `/api/empresa/proveedores`, `/api/empresa/facturacion_electronica`, `/api/empresa/facturacion_electronica/pais_detectado`, y `/api/empresa/servicios` (politica inventario).
	- Se amplian pruebas en `backend/handlers/empresa_permisos_test.go` para cobertura de los modulos nuevos.
	- Validacion tecnica: `go test ./handlers -run "WithEmpresa|ConsultarHandlerRejectsEmpresaFueraDeAlcance" -count=1` (ok) y `go test ./...` (ok).

## 2026-04-04
- Se registra nueva credencial Gemini cifrada en configuraci?n avanzada (`ai.model.google.gemini_2_0_flash.api_key` en `pcs_superadministrador`).
- Se valida consumo de Gemini con la nueva credencial: respuesta del proveedor `429` por cuota excedida (sin error de credencial/servicio bloqueado).
- Se verifica la presencia de la tarjeta de Gemini en `web/super/configuracion_avanzada.html` y se corrige un bloque JavaScript en la carga de estado para mantener consistencia de la vista.
- Se agrega prueba de seguridad de alcance por empresa para chat IA en `backend/handlers/chat_con_inteligencia_artificial_controller_test.go`:
	- `TestConsultarHandlerRejectsEmpresaFueraDeAlcance`.
	- Validaci?n: `go test ./handlers -run "TestConsultarHandlerRejectsEmpresaFueraDeAlcance|TestModelosHandlerRequiresGoogleAccount|TestModelosHandlerReturnsPreferredModelForGoogleAccount" -count=1` (ok).

## 2026-04-04
- Chat IA empresarial migrado a Gemini-only:
	- `backend/handlers/chat_con_inteligencia_artificial_controller.go` ahora integra Google Gemini (`generateContent`) y elimina dependencias de OpenAI/DeepSeek/Hugging Face para este módulo.
	- El cat?logo y la configuraci?n de credenciales IA quedan en un ?nico modelo soportado: `google:gemini-2.0-flash` (`GEMINI_API_KEY`).
	- `web/super/configuracion_avanzada.html` simplifica la tarjeta IA a una sola credencial Gemini con trazabilidad por cuenta Google.
	- `web/administrar_empresa/chat_con_inteligencia_artificial.html` se redise?a con experiencia visual tipo Gemini, chips de contexto y flujo explícito de autenticaci?n Google.
	- Pruebas ajustadas y validadas: `go test ./auth ./db ./handlers ./metrics ./utils` (ok) en `backend`.
- Se agrega gesti?n de credenciales IA en `super/configuracion_avanzada.html` para 5 modelos populares con plan gratuito limitado:
	- OpenAI GPT-4o mini,
	- OpenAI GPT-4.1 mini,
	- DeepSeek Chat,
	- DeepSeek Reasoner,
	- Meta Llama 3.1 8B Instruct (Hugging Face).
- Se crea endpoint `GET/PUT /super/api/config/ai` en backend para guardar/consultar credenciales con registro de la cuenta Google logueada que realiza cambios.
- El módulo `chat_con_inteligencia_artificial` ahora resuelve credenciales en este orden:
	- configuraci?n guardada por modelo,
	- configuraci?n por proveedor,
	- variable de entorno.
- Validaci?n t?cnica ejecutada:
	- `go test ./handlers -run "AIModelsConfigHandler|Chat|ModelosHandler" -count=1` (ok).
	- `go test ./...` en `backend` (ok).
- Se implementa la primera fase tecnica del punto 3 (permisos y seguridad) con middleware de autorizacion por rol + alcance de empresa:
	- nuevo `backend/handlers/empresa_permisos.go`,
	- aplicacion en rutas criticas de ventas, inventario y finanzas desde `backend/main.go`,
	- pruebas nuevas en `backend/handlers/empresa_permisos_test.go` para denegacion/aprobacion por rol y empresa.
- Validacion tecnica de la fase:
	- `go test ./handlers -run WithEmpresa -count=1` (ok).
	- `go test ./...` en `backend` (ok).
- Se actualiza la documentacion del proyecto para continuar el plan maestro de 14 puntos:
	- nuevo `documentos/plan_maestro_pos_multiempresa_14_puntos.md` con estado, entregables y backlog de ejecucion,
	- nueva `documentos/matriz_kpi_pos_multiempresa.md` con formulas/frecuencia/fuentes de KPI,
	- nueva `documentos/matriz_roles_permisos_pos_multiempresa.md` para iniciar el punto 3 de permisos y seguridad,
	- actualizacion de `documentos/descripcion_del_proyecto` para referenciar estos documentos como base de seguimiento.
- Continuaci?n de implementaci?n en `chat_con_inteligencia_artificial`:
	- Se corrige el orden de validación de autenticaci?n para cuenta Google en `backend/handlers/chat_con_inteligencia_artificial_controller.go`.
	- Cuando no hay cuenta Google autenticada, los endpoints del módulo IA ahora responden `401` de forma consistente (en lugar de caer en validación de alcance con `403`).
	- Se centraliza validación de alcance con `ensureEmpresaAccessByAccount` para reutilizar la cuenta ya validada.
- Se agregan pruebas autom?ticas del módulo IA:
	- `backend/db/chat_inteligencia_artificial_test.go` (upsert/get de modelo preferido y acumulaci?n de uso diario).
	- `backend/handlers/chat_con_inteligencia_artificial_controller_test.go` (autorizaci?n por cuenta Google y respuesta con modelo preferido).
- Validaci?n t?cnica ejecutada en esta continuaci?n:
	- `go test ./db -run EmpresaAI -count=1` (ok).
	- `go test ./handlers -run ModelosHandler -count=1` (ok).
	- `go test ./...` en `backend` (ok).
- Se ampl?a el módulo `chat_con_inteligencia_artificial` para registrar el modelo preferido por cuenta Google autenticada (por empresa):
	- Nueva tabla `empresa_ai_modelo_preferido` en `pcs_empresas` (UNIQUE por `empresa_id + admin_email`).
	- Nuevas funciones en `backend/db/chat_inteligencia_artificial.go`: `GetEmpresaAIModeloPreferido` y `UpsertEmpresaAIModeloPreferido`.
	- Nuevo endpoint `GET/PUT /api/empresa/chat_con_inteligencia_artificial/modelo_preferido`.
	- `GET /modelos` ahora devuelve `google_account` y `modelo_preferido`.
	- `POST /consultar` ahora persiste el `model_id` usado como preferencia de la cuenta Google y devuelve confirmación en respuesta.
- Se actualiza `web/administrar_empresa/chat_con_inteligencia_artificial.html` para:
	- cargar autom?ticamente el modelo preferido de la cuenta Google,
	- guardar el modelo preferido al cambiar selección,
	- mostrar la cuenta Google vinculada en el bloque de uso diario.
- Validaci?n t?cnica ejecutada para esta ampliaci?n:
	- `gofmt -w backend/db/chat_inteligencia_artificial.go backend/handlers/chat_con_inteligencia_artificial_controller.go backend/handlers/chat_con_inteligencia_artificial_router.go` (ok).
	- `go test ./...` en `backend` (ok).
- Se fortalece `backend/utils/utils.go` para observabilidad profesional:
	- `LoggingMiddleware` ahora genera `request_id` por solicitud, calcula `empresa_id` (query/header/JSON body) y registra inicio/fin con latencia.
	- Se agregan logs separados por empresa en `backend/logs/empresa_<id>.log` y un fallback global en `backend/logs/empresa_global.log`.
	- `JSONErrorMiddleware` ahora normaliza errores no-JSON incluyendo `request_id` y `empresa_id` cuando aplica, y registra errores API por empresa.
- Se ajustan endpoints multipart para reforzar separaci?n de logs por empresa:
	- `backend/handlers/chat_tareas.go` y `backend/handlers/productos.go` ahora establecen `X-Empresa-ID` tras parsear `empresa_id` del formulario.
- Se endurece `backend/handlers/usuarios_empresa.go` en autenticaci?n/primer ingreso:
	- se reemplazan respuestas `500` que expon?an detalles internos por mensajes profesionales y seguros,
	- se agrega logging servidor con contexto (`empresa_id`, `email`, `id`) para trazabilidad sin filtrar errores sensibles al cliente.
- Se endurece `scripts/iniciar_servidor.ps1` para detectar ca?da temprana de `server.exe`: ahora conserva el `PID`, valida salida prematura y muestra las ?ltimas l?neas de `backend/server.err` para diagn?stico inmediato.
- Validaci?n de correcci?n ejecutada:
	- `gofmt -w backend/utils/utils.go` (ok).
	- `go test ./...` en `backend` (ok).
- Se corrige `scripts/iniciar_servidor.ps1` en `Resolve-GoogleOAuthCredentials`: la construccion de `envCandidates` ahora usa `Join-Path -Path/-ChildPath` por elemento, evitando el error `CannotConvertArgument` de `Join-Path`.
- Se corrige `backend/db/finanzas.go` en `EnsureEmpresaFinanzasSchema`: los indices que dependen de columnas migradas (`periodo_contable` y `estado` de periodos) se crean al final de la migracion para compatibilidad con bases antiguas.
- Validacion de correccion ejecutada:
	- `go test ./...` en `backend` (ok).
	- `go run .` en `backend` (arranque correcto en `:8080`).
- Se incorpora el modulo `chat_con_inteligencia_artificial` en el panel empresarial con interfaz tipo chat en `web/administrar_empresa/chat_con_inteligencia_artificial.html`.
- Se crean `backend/db/chat_inteligencia_artificial.go`, `backend/handlers/chat_con_inteligencia_artificial_controller.go` y `backend/handlers/chat_con_inteligencia_artificial_router.go` para arquitectura modular (DB + controller + router).
- Se publican rutas del modulo IA:
	- `GET /api/empresa/chat_con_inteligencia_artificial/modelos`
	- `POST /api/empresa/chat_con_inteligencia_artificial/consultar`
	- `GET /api/empresa/chat_con_inteligencia_artificial/historial`
- Se agregan tablas en `pcs_empresas` para auditoria y limites diarios:
	- `empresa_ai_consultas`
	- `empresa_ai_uso_diario`
- Se integra `EnsureEmpresaAIChatSchema` y la migracion `2026-04-03-005-chat-ia-empresa` en `backend/main.go`.
- Se implementa aislamiento estricto por `empresa_id`, validacion de alcance de usuario y control de limite free-tier por empresa/proveedor/modelo/dia con opcion de upgrade.
- Se habilitan modelos famosos de OpenAI, DeepSeek y Hugging Face usando credenciales solo en backend mediante variables de entorno (`OPENAI_API_KEY`, `DEEPSEEK_API_KEY`, `HUGGINGFACE_API_KEY`).
- Se ampl?a el módulo financiero con control de periodos contables por empresa:
	- tabla `empresa_finanzas_periodos`.
	- endpoint `GET/POST/PUT /api/empresa/finanzas/periodos`.
	- acciones de cierre y reapertura de periodo.
- Se aplican bloqueos de integridad contable: no se permite crear/editar/eliminar/activar/desactivar movimientos cuando su periodo est? cerrado.
- Se ampl?a `empresa_finanzas_movimientos` con:
	- `periodo_contable`,
	- retenciones (`retencion_fuente`, `retencion_ica`, `retencion_iva`, `total_retenciones`),
	- `total_neto`.
- Se ampl?a `empresa_finanzas_configuracion` con `cuenta_retenciones_cobrar` y `cuenta_retenciones_pagar`.
- Se completa la UI de finanzas para:
	- gestionar periodos (cerrar/reabrir/actualizar),
	- calcular total bruto, retenciones y neto,
	- filtrar por periodo,
	- exportar `balance general`, `libro diario` y `libro mayor` en CSV.
- Se corrige el escaneo de puertos de seguridad para compatibilidad IPv6 usando `net.JoinHostPort` en `backend/handlers/system_empresas_handlers.go`.
- Se ajusta `scripts/iniciar_servidor.ps1` para usar nombre de función con verbo aprobado de PowerShell en la carga de `.env`.
- Validaci?n t?cnica ejecutada: `go test ./...` en `backend` (ok).
- Se implementa el módulo financiero multiempresa con enfoque unificado de ingresos y egresos en `web/administrar_empresa/finanzas.html`.
- Se crea `backend/db/finanzas.go` con esquema, validaciones y CRUD de:
	- `empresa_finanzas_movimientos`
	- `empresa_finanzas_configuracion`
- Se crea `backend/handlers/finanzas.go` y se publican rutas:
	- `GET/POST/PUT/DELETE /api/empresa/finanzas/movimientos`
	- `GET/POST/PUT /api/empresa/finanzas/configuracion`
- Se actualiza `backend/main.go` para asegurar el esquema financiero y registrar la migraci?n `2026-04-03-003-finanzas`.
- Se integra el acceso al módulo en `web/administrar_empresa.html` y `web/js/administrar_empresa.js`.
- Se agrega `backend/db/finanzas_test.go` con pruebas de configuraci?n y flujo CRUD de movimientos financieros.
- Se ampl?a `backend/tools/seed_motel_malibu/main.go` para sembrar configuraci?n financiera y movimientos demo de ingreso/egreso.
- Se separa visualmente el libro financiero en dos pesta?as operativas dentro del módulo: `Ingresos` y `Egresos`.
- Se agrega la pesta?a `Todos` para consolidar ingresos y egresos en una sola vista del libro financiero.
- Se agrega exportación del libro financiero filtrado por fechas a:
	- Excel (CSV compatible con Excel).
	- PDF (vista de impresión).
## 2026-05-18
- Se separa `Configuracion empresarial` en paginas independientes por seccion: Productos y pedidos, Identidad visual, Formato monetario, Cobro operativo, Reporte de corte, Respaldo y Pasarelas de pago.
- `web/administrar_empresa/configuracion.html` queda como nucleo unico con modo aislado por `section`, manteniendo los mismos botones de guardar y endpoints.
- Se agregan claves de permisos para las nuevas paginas del submenu y se verifica visualmente en escritorio y movil.

	- JSON contable para integraci?n externa (incluye resumen, detalle y asientos recomendados).
- Se ampl?a la configuraci?n financiera por empresa para contabilidad externa con parametrizaci?n de:
	- destino de integraci?n (`generico`, `siigo`, `world_office`, `alegra`),
	- cuentas base (caja/bancos, ingresos, IVA generado, gastos, IVA descontable),
	- cuentas por categor?a para ingresos y egresos.
- La exportación `JSON contable` deja de usar cuentas fijas y ahora construye asientos con la parametrizaci?n real guardada por empresa.
- El JSON exportado incorpora `accounting_profile` y `erp_projection` por movimiento para facilitar mapeo hacia software contable externo.
- Se actualiza `backend/db/finanzas_test.go` para validar persistencia de la nueva parametrizaci?n contable.
- Se ampl?a `web/administrar_empresa/finanzas.html` con salidas contables adicionales:
	- Plantilla dedicada SIIGO (CSV) para importaci?n de asientos.
	- Balance de prueba (CSV).
	- Estado de resultados (CSV).
- Se crea `documentos/plantillas/siigo_plantilla_importacion_asientos.csv` como plantilla de referencia ERP.
- Se crea `documentos/informe_contable_directivo_2026-04-03.md` con revisi?n de cumplimiento contable/directivo, brechas y plan recomendado.
- Validaci?n t?cnica ejecutada:
	- `go test ./... -count=1` (ok).
	- `go run ./tools/seed_motel_malibu` (ok, incluye creación de 4 movimientos financieros demo).
	- `runTests` global (ok: 3/3).

## 2026-04-03
- Se implementa control de inventario en carrito: al agregar items de producto se descuenta stock y al desactivar/eliminar items abiertos se revierte autom?ticamente.
- Se asegura que, al cerrar una venta, el descuento de inventario permanezca aplicado y no se revierta en el pago.
- Se mejoran respuestas de API para stock insuficiente en operaciones de items de carrito.
- Se agrega `backend/db/carritos_inventario_test.go` con pruebas de descuento de inventario y caso de stock insuficiente.
- Se ampl?a `backend/tools/seed_motel_malibu/main.go` para registrar 10 clientes y 10 usuarios de empresa.
- La semilla valida autom?ticamente el flujo comercial completo: venta cerrada, descuento de inventario al agregar y persistencia tras pagar.
- Se confirma en seed la validación de impresión con vista previa POS y Carta.
- Se ampl?a `web/administrar_empresa/reportes.html` con reporte de ventas, reporte de productos y reporte de compra de productos, todos con búsqueda por rango de fechas.
- Validaci?n t?cnica ejecutada: `go test ./auth ./db ./handlers ./metrics ./utils` (ok) y `go run ./tools/seed_motel_malibu` (ok).
- Se agrega el v?nculo `Ayuda` en el menú flotante global (`web/menu.js`) y se reestructura `web/ayuda/ayuda.html` como centro de ayuda con menú interno y secci?n de APIs.
- Se adapta `web/administrar_empresa/carrito_de_compras.html` para operación con lector de código de barras (escaneo por código/SKU, Enter para agregar y acumulaci?n opcional de cantidad).
- Se extiende `web/administrar_empresa/configuracion.html` con configuraci?n por empresa para el lector: habilitar, autofoco y acumulaci?n.
- Se ampl?a `web/administrar_empresa/reportes.html` con KPI de productos bajo m?nimo y reporte de inventario actual por bodega.
- Validaci?n t?cnica ejecutada para flujo carrito/inventario multiempresa: `go test ./db -run Carrito -count=1` (ok) y `go test ./handlers -run Carritos -count=1` (ok).

## 2026-04-02
- Se crea la herramienta `backend/tools/seed_motel_malibu/main.go` para cargar datos demo comerciales en la empresa Motel Malibu.
- La semilla inserta 10 productos con precios COP, 5 clientes y crea una venta de prueba cerrada para validar el flujo comercial.
- Se valida la configuracion de impresion con vista previa de formatos POS y Carta desde la herramienta de seed.
- Se implementa la seccion `web/administrar_empresa/reportes.html` con KPIs, ventas cerradas, top productos, top clientes y resumen de impresion.
- Se reestructura `backend/tools` en subcarpetas por herramienta para eliminar conflictos de compilaci?n por m?ltiples `main`.
- Se valida backend completo con `go test ./...` (ok).
- Se valida el módulo GPS con pruebas espec?ficas:
	- `go test ./db -run TestEmpresaGPSDispositivosYRecorridosCRUD -count=1` (ok).
	- `go test ./handlers -run TestEmpresaUbicacionGPSHandlersCRUDFlow -count=1` (ok).
- Se implementa el modulo de ubicacion GPS por empresa con soporte de multiples dispositivos.
- Se agregan tablas `empresa_gps_dispositivos` y `empresa_gps_recorridos` en `pcs_empresas`.
- Se crean endpoints CRUD para dispositivos y recorridos GPS en `/api/empresa/ubicacion_gps/*`.
- Se agrega la pagina `web/administrar_empresa/ubicacion_gps.html` con mapa OpenStreetMap (Leaflet).
- Se habilita tracking automatico de recorridos cada 10 segundos por dispositivo.
- Se agregan pruebas en `backend/db/ubicacion_gps_test.go` y `backend/handlers/ubicacion_gps_test.go`.



### Added
- ImplementaciÃƒÂ¯Ã‚Â¿Ã‚Â½n de controladores HTTP (Handlers) para el CRUD de Proveedores integrado en el nuevo mÃƒÂ¯Ã‚Â¿Ã‚Â½dulo de compras y logÃƒÂ¯Ã‚Â¿Ã‚Â½stica ERP.



## [2026-04-20] Limpieza Total Themes
- [UI/Temas] Auditoría y barrido de más de 50 páginas y scripts en web/administrar_empresa, web/super y páginas públicas para limpiar colores fijos, migrando lógicas JS a .classList.add('text-danger') y respetando las 6 paletas dinámicas. Completado barrido masivo de vistas.
- **Auditoria + IA contextual**: los wrappers empresariales auditan ahora acciones `R/C/U/D/A` de forma no bloqueante y la IA empresarial/global recibe una ventana reciente de `empresa_auditoria_eventos` como contexto operativo. La integracion vive en auditoria + constructor de prompt, sin insertar IA en cada modulo; si auditoria o IA fallan/deshabilitan, el servidor continua con estado controlado. Archivos: `backend/handlers/auditoria_empresa.go`, `backend/db/auditoria_empresa.go`, `backend/db/chat_inteligencia_artificial.go`, `backend/handlers/chat_con_inteligencia_artificial_controller.go`, `backend/handlers/chat_con_ia_global_super.go`, documentacion relacionada.

## [2026-05-03] Documentacion, ayuda y estado operativo de modulos
- [Docs] Se crea `documentos/reporte_estado_modulos_2026-05-03.md` con estado compacto por modulo, observaciones de calidad y dependencias pendientes de certificacion.
- [Ayuda] Se actualiza `web/ayuda/ayuda.html` con una seccion de estado operativo, estaciones/carrito, tarjetas adaptables, indicadores del panel y limites honestos de validacion.
- [Operacion] Se documentan los cambios recientes: carrito desde estacion, pago con retorno a estaciones, `USD / COP` primero y despliegue VPS correcto.
- 2026-05-11: contrato universal de 30 plantillas como plantillas. Se retira el contrato publico de migracion manual antigua de la matriz, se eliminan botones/acciones visibles de migracion en los 10 plantillas clasicos y la activacion de licencias queda limitada a aplicar preconfiguracion idempotente por tipo de empresa. Las 10 clasicas y las 20 nuevas se presentan igual: plantillas sobre el nucleo comun de clientes, productos/servicios, ventas, pagos, facturacion, reportes, roles y licencias. Verificacion: `go test ./...` en `backend/`, validacion JS y scripts inline de las pantallas tocadas.
- 2026-05-12: normalizacion documental, UX accesible y PostgreSQL-only. Se corrigen textos mojibake en documentos y pantallas priorizadas, el auditor documental evita falsos positivos en query strings, los botones sin nombre accesible reciben `aria-label`, y los compose oficiales de Nextcloud pasan de MariaDB a PostgreSQL 16 para mantener un unico motor relacional operativo.
- 2026-05-12: CRM empresarial profesional. `GET /api/empresa/crm_avanzado` con `action=dashboard` ahora entrega salud comercial, valor en riesgo, leads sin contacto, oportunidades estancadas, acciones priorizadas, responsables y canales; `crm_comercial.html/js` muestra cockpit ejecutivo y los CRUD `/api/empresa/crm/*` pasan a `WithEmpresaCRMUnificadoPermissions`. Sin tablas ni dependencias nuevas.
- 2026-05-12: ajuste visual del menu empresarial. `web/administrar_empresa.html` retira el indicador compacto `Plantillas · conteo · API/local` del encabezado lateral y mueve `Soluciones por negocio` a la parte baja del menu, justo encima de `Administracion`. No cambia APIs, permisos, tablas ni dependencias.
- 2026-05-12: scroll superior en Tipos de empresa. `web/super/tipos_empresas.html` agrega una barra horizontal sincronizada arriba de la tabla, conservando el scroll inferior existente y sin cambiar APIs, permisos, datos ni dependencias.

- 2026-05-12: actualizada la vista super `Plantillas 20`. `web/super/plantillas_produccion_masiva.html` pasa de tabla tecnica a consola ejecutiva con semaforo comercial, KPIs de brechas, tarjetas de foco, filtros por sin licencia/sin preconfiguracion y cruce mas robusto de catalogo, preconfiguraciones y licencias. No cambia endpoints, permisos, base de datos ni dependencias.

- 2026-05-12: limpieza visual del super administrador. `web/super_administrador.html` retira los cuatro indicadores compactos del encabezado lateral (`PostgreSQL`, `VPS`, `Licencias`, `Seguridad`) sin cambiar accesos, iframes, permisos ni endpoints.

- 2026-05-12: identidad visual empresarial. `web/administrar_empresa/configuracion.html` agrega una seccion visible para subir o pegar el logo de la empresa, `web/administrar_empresa/panel.html` lo muestra encima del titulo sin alterar la tarjeta de clima, y `/api/empresa/configuracion_avanzada/logo` guarda el archivo en `/uploads/empresa_logos/empresa_<id>/`. Se reutilizan `empresa_configuracion_avanzada.logo_url` y `mostrar_logo`, los mismos campos usados por factura y documentos.
- 2026-05-13: correos masivos globales. Super administrador agrega `web/super/correos_masivos.html` y `/super/api/correos_masivos` para previsualizar destinatarios, enviar comunicados por SMTP a administradores/usuarios y auditar campanas en `super_correos_masivos`.
- 2026-05-13: conexion obligatoria para operacion. Se retira la opcion de modo offline/contingencia DIAN en facturacion electronica y el backend bloquea emisiones cuando DIAN/proveedor no esta disponible; ventas, cobros y facturacion requieren conexion activa con el servidor.

## [2026-05-19] Administrar empresa movil
- [Frontend/PWA] El panel `Administrar empresa` actualiza service worker a `pcs-shell-v3`, limpia caches antiguas y usa network-first para CSS/JS/manifest para que celulares y PWA instaladas no muestren estilos viejos.
- [Responsive] Menu, submenus, botones e iframe del shell empresarial se ajustan al ancho movil sin desborde horizontal ni doble scroll innecesario; el panel inicial deja de cortar titulo, ciudad, clima y pie en pantallas pequenas.
- [QA] Validacion visual en viewport movil 390x844 y validacion de sintaxis de scripts/service worker.

## [2026-05-19] Clientes desde carrito
- [Carritos] El boton `Clientes` abre un panel interno para crear/asignar cliente al carrito activo sin salir de venta directa o estaciones.
- [Configuracion] Nuevo check `Exigir cliente registrado para pagar` dentro de la configuracion del carrito.
- [Backend] `pagar_estacion` bloquea el cierre cuando `cliente_obligatorio_pago` esta activo y el carrito no tiene cliente.
- Creditos diarios para ventas financiadas de motos.
	- Archivos modificados: `backend/db/creditos.go`, `backend/handlers/creditos.go`, `backend/db/creditos_postgres_test.go`, `backend/main.go`, `web/administrar_empresa/creditos.html`, `documentos/estructura_bd.md`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/historial_de_cambios`.
	- Descripcion: el contrato de credito acepta periodicidad de cuota, valor pactado y omision opcional de domingos; las cuotas diarias soportan planes largos de dos anos y la cartera muestra cuotas/dias vencidos desde el calendario de pagos.
	- Verificacion: `go test ./db -run "TestCredito" -count=1`, `go test ./handlers -run '^$' -count=1` y validacion de sintaxis JS de `web/administrar_empresa/creditos.html`.

## [2026-05-28] Descarga informacion empresa
- [Backend] El snapshot de `/super/api/empresas?action=resumen_descarga|exportar_informacion` compara `empresa_id`/`id` como texto y continua con advertencias si una tabla puntual no se puede leer, evitando que una tabla tumbe toda la descarga.
- [Frontend] La vista `descargar_informacion_de_la_empresa.html` conserva estados claros de carga/descarga y mejora contraste visual en temas claros y oscuros.
- [QA] Validacion visual desde el boton de descarga en `seleccionar_empresa.html` hasta descarga JSON correcta.

## [2026-05-28] Instalacion PWA desde login
- [Backend] `/manifest.webmanifest` y `/sw.js` quedan publicos en el middleware para que el navegador pueda validar la instalacion de la app sin sesion.
- [Frontend] `login.html` e `index.html` comparten el mismo icono PWA en favicon/apple touch icon; el encabezado publico del index tambien usa ese icono.
- [QA] Prueba visual del boton `Instalar app` y validacion de iconos del index.

## [2026-05-28] Rediseño descarga informacion empresa
- [UX] `descargar_informacion_de_la_empresa.html` pasa a una vista simple: nombre de empresa, selector de formato y boton `Descargar`.
- [Frontend] Se eliminan de la pantalla los paneles tecnicos de resumen, formatos y tablas incluidas.
- [QA] Prueba visual escritorio/movil y descarga CSV desde el boton principal.

## [2026-05-28] Apariencias oscuras profesionales
- [Temas] Se agregan `Negro Absoluto` (`dark-absolute`) y `Obsidiana Profesional` (`dark-obsidian`) al menu flotante.
- [Frontend] Los nuevos temas tienen variables completas de fondo, superficies, texto, bordes, acentos, formularios y estados.
- [QA] Prueba visual seleccionando ambos temas desde el menu flotante y verificando persistencia.
## [2026-05-29] Cajeros simultaneos por estaciones
- Se agrega control por cajero para asignar estaciones visibles/operables desde `Administrar usuarios`.
- Los endpoints de carritos e items validan la asignacion en backend para evitar acceso por URL/API a estaciones no permitidas.
- Los totales de la tarjeta Caja se calculan para el usuario autenticado, manteniendo caja y reporte de turno independientes en cajas simultaneas.

## [2026-05-29] Favoritos con icono original
- Los favoritos del panel empresarial usan el icono original del boton del menu principal en lugar de la estrella.
- Los favoritos nuevos guardan el icono de origen para reutilizarlo al volver al panel.

## [2026-05-30] Bandeja de correo corporativo
- El autologin de SnappyMail conserva la forma `index.php?sso&hash=...` al agregar el tema del sistema.
- La bandeja de correo del panel empresarial deja de quedar bloqueada en `/webmail/sso.php` con 403.

## [2026-05-30] Menu super Portal publico e index
- [Super administrador] Se agrupan en un solo bloque las opciones que editan el portal publico: tarjetas del index, modulos del index, descripcion de sistemas y WhatsApp del portal.
- [Navegacion] Gobierno queda solo para centro de mando y selector de empresas; el grupo de mensajeria queda separado del index.
- [QA] Validacion de sintaxis del JS del panel super y revision visual del menu reorganizado.

## [2026-05-30] Menu super Mensajeria y alertas
- [Super administrador] Se concentra en `Mensajeria y alertas` la gestion de alertas por email, alertas de licencia, formatos de email, mensajes masivos, mantenimiento, Gmail SMTP y email corporativo.
- [Licencias] El acceso `Formatos de email` queda visible para editar mensajes como confirmacion de cuenta y pago/compra de licencias.
- [QA] Validacion de sintaxis del JS del panel super, existencia de paginas e iconos del nuevo bloque.

## [2026-05-30] Menu super sin paginas huerfanas
- [Super administrador] `Asesores de ventas` queda como primer acceso del bloque `Comercial y licencias`.
- [Navegacion] Se agregan botones para paginas super que existian sin enlace: reportes globales, auditoria global, metricas de trafico, preconfiguracion, contrato, administradores Frecuencia FE, chat IA global, voz IA, servidores, soporte remoto y configuracion avanzada.
- [QA] Cruce automatico: 52 paginas HTML bajo `web/super` y 52 enlaces en el menu; ningun enlace del menu queda fuera de paginas permitidas/restaurables.
- 2026-05-30: `Licencias globales compartidas` reduce el catalogo base a cuatro planes globales para todos los tipos de empresa; desde 2026-05-31 elimina planes heredados repetidos sin empresa asignada, mantiene la prueba gratis de 15 dias una sola vez por empresa y agrega una tarjeta visible de reglas en Super administrador > Licencias.
- 2026-05-30: `Configuracion general PostgreSQL` corrige el esquema y consultas de `empresa_configuracion_general` para usar `BIGSERIAL`, fecha compatible y placeholders traducidos, evitando 500 al entrar al panel de empresas nuevas.
- 2026-05-30: `Checkout Wompi de licencias` ajusta la consulta publica de terminos del comercio para no enviar cabecera `Authorization` al endpoint merchants de Wompi y desbloquear la prueba visual de planes comerciales.
- 2026-05-30: `Configuracion general PostgreSQL reforzada` asegura columna `fecha_creacion`, secuencia `id` y fechas texto en `empresa_configuracion_general` para evitar 500 al cargar paneles de empresas nuevas o antiguas en la VPS.
- 2026-05-30: `Configuracion general en empresas nuevas` corrige el alta automatica de la configuracion por defecto agregando el valor faltante de `clima_fuente` en el INSERT.
- 2026-05-30: `Checkout licencias Wompi` valida que `wompi.public_key` tenga formato `pub_test_` o `pub_prod_` antes de publicar Wompi como medio disponible, evitando 502 por llaves placeholder.
- 2026-05-30: `Modo Wompi por llave` cuando `wompi.mode` manual contradice la llave (`pub_prod_` vs sandbox o `pub_test_` vs production), el backend usa el modo inferido por la llave para consultar el endpoint correcto.
- 2026-05-31: `Clima en centro de mando super` agrega la tarjeta de clima al inicio del panel super, deja favoritos debajo y mueve Estado general al tercer bloque visual.
- 2026-05-31: `Favoritos de super administrador` agrega favoritos en el centro de mando super, boton estrella global para paginas del iframe, compacta Estado general e Incidentes recientes y elimina la tarjeta Accesos clave.
- 2026-05-31: `Centro de mando super simplificado` quita la tarjeta superior `Super administrador / Panel ejecutivo` y elimina la tarjeta de clima del panel super.
- 2026-05-31: `Super administrador simplificado` quita del menu super los accesos a Reportes globales y Metricas de trafico; Reportes globales se conserva solo desde Seleccionar empresa y se elimina la pagina puente de metricas de trafico.
- 2026-05-31: `Licencias fijas globales` deja el catalogo comercial en 4 planes para todas las empresas: prueba gratis 15 dias, COP 60000, COP 100000 y COP 150000; elimina del catalogo sin empresa asignada las licencias sobrantes y bloquea crear/eliminar/ocultar esos planes desde la pagina de licencias.
- 2026-05-31: `Login con titulo texto e iconos PWA` reemplaza el logo imagen del login por el titulo `Powerful Control System`, agrega iconos a `Instalar app` e `Ir al inicio` y conserva el icono del instalador al cambiar estado.
- 2026-05-31: `Licencia de prueba historica bloqueada` refuerza backend para que la prueba gratis de 15 dias no pueda activarse otra vez por la misma empresa aunque la prueba anterior este vencida, inactiva o venga de datos antiguos sin marca vigente.
- 2026-06-01: `Cantidad natural en carrito` restringe las cantidades del carrito manual y lector de codigo de barras a enteros positivos; backend y capa de datos rechazan cero, negativos y decimales para proteger inventario y pagos.
- 2026-06-01: `Invitacion de administrador existente` cambia `/super/api/administradores` para que el panel super no devuelva 409 cuando el correo ya existe y esta confirmado; responde OK con mensaje claro y permite actualizar nombre/rol con validaciones de rol super.
- 2026-06-01: `Impuesto seleccionable en productos` cambia el campo de impuesto del formulario de productos a un selector cargado desde `/api/empresa/impuestos?action=context`, filtrando impuestos activos aplicables a ventas por `empresa_id` y guardando el porcentaje compatible con el nucleo de inventario.
- 2026-06-01: `Espaciado vertical de estaciones` separa el espacio horizontal y vertical de la grilla de estaciones para que las tarjetas pequenas/miniatura mantengan aire entre filas.
- 2026-06-01: `Cliente general con mayusculas/minusculas` conserva el texto escrito en `Nombre para ventas sin cliente` al normalizar configuracion de estaciones y carrito.
- 2026-06-01: `Usuarios operativos sin aprobacion extra` corrige `/api/empresa/usuarios` para que crear, editar, activar, reenviar invitacion o eliminar usuarios use permisos normales `seguridad:C/U/D` y auditoria, sin exigir `aprobado_por` ni `codigo_aprobacion`; la aprobacion trazable se conserva para `/api/empresa/roles_de_usuario` y `/api/empresa/permisos_empresa`.
- 2026-06-01: `Alta de usuario tolerante a correo` ajusta `/api/empresa/usuarios` para conservar el usuario pendiente cuando falla el envio de invitacion y responder `email_sent=false`, permitiendo reenviar confirmacion sin perder el alta.
- 2026-06-01: `GRAFOLOGIX grafologia OCR` agrega modulo empresarial para subir/tomar fotos de manuscritos, ajustar imagen, analizar metricas graficas con Go puro, guardar informes por `empresa_id`, exportar HTML/JSON/PDF real y documentar Tesseract opcional sin dependencias Go nuevas.
- 2026-06-01: `GRAFOLOGIX Fase 2` agrega preprocesamiento visual con escala de grises, binarizacion, bordes y lineas/margenes, guarda `preprocesamiento_json` y corrige apariencia clara/oscura del dashboard.
- 2026-06-01: `GRAFOLOGIX exportaciones` agrega salida Word compatible, CSV y TXT, con botones visibles en resultado e historial junto a HTML, JSON y PDF.
- 2026-06-01: `GRAFOLOGIX OCR y zoom` habilita Tesseract libre en la imagen Docker del backend, expone variables `GRAFOLOGIA_TESSERACT_*` en Compose y agrega controles de ampliar/reducir/restablecer imagen antes del analisis.
- 2026-06-01: `Roles globales de usuarios empresariales` hace que Administrar usuarios muestre todos los roles activos en todos los tipos de empresa, deduplicados por alias, con descripcion y enlace `Saber mas` a la ayuda de cada rol.
- 2026-06-01: `GRAFOLOGIX con cliente asociado` permite buscar o crear un cliente central desde la pantalla de grafologia, asociarlo al manuscrito, registrar descripcion/caracteristicas de la persona y mostrar ese contexto en resultados e informes exportables.
- 2026-06-01: `GRAFOLOGIX IA y credenciales OpenAI` optimiza el envio de imagenes a JPEG redimensionado para evitar `413`, corrige el orden de migracion del indice por cliente y agrega diagnostico cuando la API key cifrada no puede descifrarse con la `CONFIG_ENC_KEY` actual. En VPS se verifico que OpenAI esta habilitado, pero las filas cifradas de OpenAI no se pueden recuperar con las llaves disponibles y `OPENAI_API_KEY` esta vacia; requiere re-registrar la API key o definirla en entorno Docker.
- 2026-06-01: `Instalar app en logins` repara el boton PWA de `login.html` y `login_usuario.html` capturando temprano `beforeinstallprompt`, reutilizando el prompt en `pwa_install.js` y registrando el service worker sin esperar al evento `load`.
- 2026-06-02: `Reenvio confirmacion usuarios empresa` agrega respaldo por Mailu/correo corporativo interno para confirmaciones y recuperacion de contrasena de usuarios operativos cuando Gmail SMTP falla por configuracion o clave no descifrable.
- 2026-06-02: `Usuario empresa duplicado` cambia el error por correo repetido en Administrar usuarios para mostrar una explicacion profesional y sugerir usar `Reenviar confirmacion` cuando la cuenta sigue pendiente.
- 2026-06-02: `Apariencia usuarios empresa` corrige la tarjeta de agregar usuario en Administrar empresa > Usuarios para que respete temas claros y oscuros sin forzar fondo oscuro.
- 2026-06-02: `Ojo de contrasena login usuario` inicializa el boton de visualizar/ocultar contrasena en `login_usuario.html` y conserva el foco del campo.
- 2026-06-02: `Caja de trabajo para cajeros` agrega en configuracion de estaciones el check activo por defecto para pedir caja al cajero al iniciar sesion; `login_usuario.html` muestra selector de caja, recuerda la ultima caja por usuario/empresa y propaga la caja a estaciones, carrito y corte.
- 2026-06-03: `Nomina multi-sede y DIAN` agrega sede/centro de costo a empleados, liquidaciones y desprendibles, resume sedes activas, reparte el seed de Motel Calipso por sedes y agrega estado/preparacion de documento soporte de pago de nomina electronica desde `/api/empresa/nomina`.
- 2026-06-05: `Index Documentos electronicos DIAN Colombia` actualiza la tarjeta publica de modulos para listar documentos y eventos SFE: factura electronica, notas credito/debito, contingencia, documento soporte y nota de ajuste.
- 2026-06-08: `DIAN sin preset reducido` retira del Centro de habilitacion DIAN el preset pequeno de software gratuito y deja solo el objetivo real del portal, el historico excepcional y el modo personalizado; si llega un valor viejo se normaliza a 30 facturas, 10 notas debito y 10 notas credito.
- 2026-06-10: `Bodegas y traslados visibles` cambia el acceso de Productos para abrir `administrar_productos.html?view=bodegas`, donde estan el traslado entre bodegas, ajustes, movimientos y existencias.
2026-06-10 - Carrito: medios de pago configurables y pagos combinados
- Agrega en Configuracion carrito checks por empresa para efectivo, tarjeta credito, tarjeta debito, transferencia Bre-B, Nequi y otra transferencia.
- El carrito permite pagos combinados desde `Detalle del pago`, abonos y pago mixto usando esos medios.
- Backend acepta los nuevos metodos, exige referencia en tarjetas/transferencias y bloquea medios deshabilitados por empresa o rol.
- Bre-B queda preparado como medio separado para una conciliacion automatica futura mediante webhook/API bancaria.

2026-06-11 - Finanzas: Pagos Bre-B QR
- Agrega en Finanzas y cumplimiento la pagina `Pagos Bre-B QR` y su tutorial.
- Permite configurar QR/cuentas receptoras por empresa y caja, listar pagos Bre-B reales del carrito y registrar pagos bancarios manuales para conciliacion.
- El endpoint `/api/empresa/finanzas/breb_qr` conserva aislamiento por `empresa_id` y no simula confirmacion bancaria sin webhook/API real.

2026-06-11 - Carrito: busqueda por nombre
- El panel de coincidencias del buscador rapido por nombre se oculta al seleccionar un producto, conservando la seleccion para agregarlo al carrito.
- 2026-06-13: `Carrito PCS con cantidad por mouse y pagos legibles` agrega botones `-`/`+` junto a cada cantidad del carrito y ensancha los campos del detalle de pago para que pagos combinados 25/25/50 se lean completos en venta directa.
- 2026-06-13: `Carrito PCS formato final de pago` formatea los pagos combinados al perder foco y evita mostrar un error interno cuando la venta directa ya quedó operativa tras una carga parcial.
- 2026-06-13: `Carrito PCS botones de acciones` permite que Descuentos abra el panel de cobro desde venta directa y oculta Domotica cuando no hay estacion fisica asociada.
- 2026-06-14: `Carrito y super panel compactos` unifica el icono flotante IA con `/img/gpt.svg`, coloca Favoritos y Email corporativo en la misma fila del panel super en PC, agrega iconos al menu flotante y compacta venta directa con selector de acciones, tarjetas alineadas y tabla de productos mas estrecha.
