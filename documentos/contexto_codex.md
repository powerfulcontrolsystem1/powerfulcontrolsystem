# Contexto rapido para Codex

Este archivo es la primera lectura operativa antes de tocar el proyecto. Resume
lo que Codex debe tener en memoria para evitar redescubrir rutas, flujos y
decisiones en cada tarea.

## Resumen del sistema

Powerful Control System es un POS/ERP SaaS multiempresa. El backend esta escrito
en Go, la persistencia oficial es PostgreSQL y el frontend usa HTML, CSS y
JavaScript tradicional servido como archivos estaticos. El sistema cubre login
administrativo, creacion de empresas, licencias, carritos, estaciones, caja,
inventario, compras, creditos, facturacion electronica por pais, reportes,
offline, super administrador y despliegue Docker/VPS.

No se deben agregar dependencias externas ni cambiar `go.mod` sin autorizacion
explicita. No se deben documentar secretos, claves, tokens ni contrasenas.

## Arranque y estructura base

- Raiz del repo: `D:\powerfulcontrolsystem`.
- Backend Go: `backend`.
- Frontend estatico: `web`.
- Paginas empresariales: `web/administrar_empresa.html` y
  `web/administrar_empresa/`.
- Paginas super administrador: `web/super_administrador.html` y `web/super/`.
- Documentacion operativa: `documentos/`.
- Scripts de operacion: `rs.ps1`, `scripts/rs.ps1`,
  `scripts/sync_to_vps.ps1`, `scripts/profesional_preflight.ps1`.
- Docker/VPS: `deploy/`, `docker-compose*.yml`, `documentos/docker_vps_operacion.md`.

El servidor registra rutas principales en `backend/main.go`. El frontend se
sirve desde `web` y algunas paginas reciben inyecciones globales, como iconos de
botones, desde handlers estaticos del backend.

## Bases de datos

- `pcs_empresas`: datos operativos por empresa.
- `pcs_superadministrador`: configuracion global, licencias, portal publico,
  alertas super administrador y catalogos globales.
- Todo cambio multiempresa debe filtrar y validar `empresa_id` en backend. Nunca
  se debe confiar solamente en URL, localStorage, cache o datos enviados por el
  navegador.
- La configuracion empresarial flexible se guarda con frecuencia en
  `empresa_estacion_prefs`, especialmente `estaciones_config` y
  `carrito_ui_global`.

## Rutas y endpoints de referencia

- Login y registro admin: `web/login.html`,
  `web/registrar_nuevo_usuario_administrador.html`,
  `backend/handlers/auth_admin_handlers.go`.
- Seleccion y creacion de empresas: `web/seleccionar_empresa.html`,
  `web/js/seleccionar_empresa.js`, `/super/api/empresas`.
- Administradores delegados: `web/super/administradores.html`,
  `/super/api/administradores`, `backend/handlers/auth_admin_handlers.go`.
  Si el correo no existe se registra por invitacion con token; si ya existe y
  esta confirmado se usa `admin_principal_delegaciones` para que vea sus empresas
  propias mas las empresas compartidas, sin cambiar `usuario_creador`.
  El selector de empresas debe resolver alcance efectivo por cuatro caminos:
  propietario por `usuario_creador`, delegado del principal, empresa compartida
  con el administrador y empresa que el administrador compartio con otro usuario;
  esta ultima conserva `access_source=owner` para que el propietario no pierda la
  empresa despues de compartirla.
  Desde `seleccionar_empresa.html` siempre se abre con `scope=principal` y solo
  debe mostrar invitados del administrador autenticado; sin ese parametro el
  panel super mantiene la vista global. Los nuevos `super_administrador` tambien
  se crean por invitacion con token.
- Auditoria global del selector: `web/super/auditoria_global.html`,
  `/super/api/auditoria`, `backend/handlers/auditoria_super.go`,
  `pcs_superadministrador.super_auditoria_eventos`. Desde el selector se abre
  con `scope=principal`: un administrador normal ve solo su alcance y el super
  administrador puede ver global desde el panel super.
- Auditoria especial super administrador: `web/super/auditoria_super_admin.html`
  en `web/super_administrador.html > Acceso`. Usa
  `/super/api/auditoria?scope=super_panel`, reservado a roles super, para revisar
  navegacion, botones `Editar`, guardados/pruebas y endpoints sensibles del panel
  super. Nunca se deben guardar secretos en metadata.
- Licencias: `web/elegir_licencia.html`, `web/pagar_licencia.html`,
  `web/super/licencias.html`, `web/super/formato_para_emviar_email.html`,
  `web/super/licencias_codigos_descuento.html`,
  `web/administrar_empresa/licencia_sistema.html`, `/super/api/licencias`,
  `/super/api/licencias/codigos_descuento`, `/licencias/activar_sin_pago` y
  `/api/empresa/licencia_sistema/pdf`.
  El catalogo base vigente es global para todos
  los tipos de empresa (`tipo_id=0`, `pais_codigo=GLOBAL`) con cuatro planes:
  prueba gratis 15 dias, COP 60000, COP 100000 y COP 150000. La
  prueba gratis solo se puede activar una vez por empresa, incluso cuando la
  prueba anterior ya vencio, quedo inactiva o viene de datos antiguos; las
  licencias base antiguas por tipo y addons de catalogo sin empresa asignada se
  eliminan del catalogo comercial. Al activarse una licencia por pago o por
  flujo de valor cero permitido, `backend/handlers/payments_handlers.go` envia
  correo al administrador de la empresa y adjunta un PDF de licencia de software
  generado en Go puro. Ese mismo PDF se descarga desde Administrar empresa >
  Licencia > Licencia del sistema y su texto se edita con la plantilla
  `licencia_software_pdf` de Super administrador > Formatos de email.
  Si una empresa paga una licencia comercial antes de que venza la licencia
  actual, la nueva vigencia no reemplaza ni acorta la anterior: se programa
  desde el vencimiento acumulado mas lejano de esa empresa y queda lista para
  iniciar automaticamente al terminar la licencia vigente. Las tablas
  `pagos_epayco` y `pagos_wompi` guardan `licencia_activation_status`,
  `licencia_activada_id` y `licencia_activada_en` para que una consulta o
  webhook repetido no sume dias dos veces.
  Los codigos de descuento de licencias se administran desde Super
  administrador > Comercial y licencias > Codigos descuento; se guardan en
  `configuraciones.licencias.discount_codes` con formato `CODIGO=10%`,
  `CODIGO=50000` o `CODIGO=gratis`, y el checkout registra su uso en pagos o
  activaciones sin pago para bloquear reutilizacion por la misma empresa.
  El checkout publico de licencia debe mostrar Epayco y Wompi cuando sus
  credenciales reales estan configuradas; `*.enabled` solo se usa como override
  explicito para apagar una pasarela lista.
- Menu super administrador: `web/super_administrador.html` debe enlazar solo
  las paginas activas del panel super; `web/js/super_administrador.js` debe
  permitir restaurar cada enlace con `target="contentFrame"`. `Reportes globales`
  no va en el menu super y se conserva como vista del selector de empresas.
  `Metricas de trafico` no existe como pagina independiente; sus datos viven en
  `Centro de mando`. El acceso `Asesores de ventas` vive al inicio de
  `Comercial y licencias`.
- Panel empresarial: `web/administrar_empresa.html`,
  `web/administrar_empresa/panel.html`.
- Domotica: boton principal en `web/administrar_empresa.html`, submenu
  `web/administrar_empresa/modulo_menu.html?module=control_electrico`,
  consola `web/administrar_empresa/control_electrico.html` con vistas por
  `pagina=resumen|conexion|raspberry|reles|automatizaciones|reportes|bitacora`
  para que cada boton del submenu abra una pagina/vista independiente,
  endpoint `/api/empresa/control_electrico` y storage super en
  `web/super/domotica_storage.html` con `/super/api/domotica_storage`.
  Conserva la clave tecnica `control_electrico`; la carpeta empresarial de
  imagenes es `web/uploads/empresas/empresa_{id}_{slug}/imagenes/`, con
  subcarpetas como `domotica` y `usuarios`.
- Configuracion empresa: `web/administrar_empresa/configuracion_menu.html` y
  paginas bajo `web/administrar_empresa/configuracion/`. La configuracion de
  `Campos obligatorios para productos` vive en Configuracion > Productos y
  pedidos y guarda en `/api/empresa/inventario/configuracion`; el modulo de
  productos solo la consume para validar y marcar campos del formulario.
- Configuracion carrito: `web/administrar_empresa/configuracion_carrito_de_compra_empresa.html`,
  `/api/empresa/estacion_prefs`; la visibilidad automatica de la tarjeta
  Domotica se guarda como `mostrar_tarjeta_domotica_carrito` dentro de
  `carrito_ui_global` o del override por estacion.
- Estaciones: `web/administrar_empresa/estaciones.html`,
  `/api/empresa/carritos_compra`.
- Carrito y venta directa: `web/administrar_empresa/carrito_de_compras.html`,
  `/api/empresa/carritos_compra`, `/api/empresa/carritos_compra/items`.
  Venta directa usa el carrito canonico `VENTA-DIRECTA-{empresa_id}-0`,
  comparte la UI unificada de estaciones y tiene boton de pantalla completa.
  Los productos y recetas descuentan/reservan inventario en tiempo real al
  agregarse al carrito mediante `carrito_compra_items`; el pago no debe volver a
  descontar stock. El cierre `action=pagar_estacion` debe ser idempotente en
  backend: solo una solicitud puede pasar el carrito de abierto a cerrado, y los
  reintentos concurrentes no duplican documento, caja, metricas ni kardex.
  La apariencia plana del carrito se controla desde `web/estilos.css` con
  `body.carrito-flat-page`: no debe recuperar sombras, relieves ni tarjetas con
  apariencia 3D. El fondo estructural usa `--carrito-page-bg`, mas oscuro que
  las tarjetas `--carrito-card-bg`, para diferenciar zonas en todas las
  apariencias. Si se abre dentro de `web/administrar_empresa.html`, el iframe
  debe conservar `allow="geolocation; fullscreen"` y `allowfullscreen`.
- Caja y corte: `web/administrar_empresa/corte_de_caja.html`,
  `/api/empresa/corte_caja`, `/api/empresa/corte_caja/configuracion`.
- Reportes de turnos: `web/administrar_empresa/reportes_turnos.html`,
  `/api/empresa/corte_caja?action=turnos|turno_reporte|turno_export|turno_email`.
- Facturacion electronica: `web/administrar_empresa/facturacion_electronica_menu.html`,
  `/api/empresa/facturacion_electronica`,
  `/api/empresa/facturacion_electronica/panama`,
  `/api/empresa/facturacion_electronica/ecuador`.
- Facturacion offline: `/api/empresa/offline_ventas`,
  `backend/db/offline_ventas.go`. El carrito guarda la cola local por
  `empresa_id + usuario/cajero`, exige caja abierta cargada antes de vender sin
  internet y sincroniza con `sync_key` idempotente que incluye empresa, cajero,
  caja y carrito.
- Alertas sistema super administrador: `web/super/alertas_sistema.html`,
  `/super/api/alertas_sistema`.
- Mensajeria y alertas en super administrador: el menu lateral agrupa
  `web/super/alertas_sistema.html`, `web/super/configuracion/alertas_licencia.html`,
  `web/super/formato_para_emviar_email.html`, `web/super/correos_masivos.html`,
  `web/super/mantenimiento_sistema.html`, `web/super/configuracion/gmail_smtp.html`
  y `web/super/email_corporativo.html`. Los mensajes de compra/pago de licencia
  se editan desde `Formatos de email`; en esa misma pagina tambien se configura
  el texto del PDF `licencia_software_pdf` que se adjunta al correo de licencia
  activada y que cada empresa puede descargar desde Administrar empresa >
  Licencia > Licencia del sistema.
- Email corporativo Mailu: `web/super/email_corporativo.html`,
  `/super/api/email_corporativo`, `/api/empresa/email_corporativo`,
  `backend/handlers/email_corporativo_handlers.go`,
  `backend/db/email_corporativo.go`,
  `deploy/scripts/vps-provision-mailu-mailbox.sh` y
  `deploy/scripts/vps-configure-mailu-host-nginx.sh`.
  El modulo genera email unico por empresa al crearla. El proveedor activo es
  Mailu con webmail SnappyMail en el perfil Docker `mail`. En la VPS se usa
  `mailu_direct`, que ejecuta el script directo contra `pcs-mailu-admin` y crea
  o actualiza buzones con `flask mailu user` y `flask mailu password`. El mismo
  script crea la identidad principal en SnappyMail para evitar el modal inicial.
  Los servicios Mailu tienen IPs fijas en `pcs_mailu_internal` para que IMAP,
  SMTP y webmail se hablen por la red confiable de Mailu. El archivo
  `deploy/mailu/snappymail-application.ini` conserva `secfetch_allow` para que
  el webmail pueda abrir en iframe `same-site` dentro del panel empresarial y
  usa `PCSLight@custom` como tema base. Los temas `PCSLight@custom` y
  `PCSDark@custom` se montan desde `deploy/mailu/themes`; el panel envia
  `theme=light|dark` al endpoint empresarial y al autologin. No imprimir claves: la clave
  inicial del buzon se guarda cifrada con `CONFIG_ENC_KEY` cuando aplica. En
  Docker portable, `deploy/.env.platform` provee `EMAIL_CORPORATIVO_*` /
  `MAILU_*`; `EMAIL_CORPORATIVO_AUTOLOGIN_SECRET` firma tokens HMAC de 2
  minutos para entrar al webmail sin exponer contrasenas. El proxy del host
  limpia cabeceras publicas y solo el backend inyecta cabeceras SSO hacia
  SnappyMail. La pagina de super administrador incluye diagnostico operativo y boton
  `Probar Mailu`; el panel empresarial abre la bandeja automaticamente cuando el
  buzon esta asignado, salvo que `empresa_estacion_prefs` tenga
  `email_corporativo_config.auto_open=false`. La pagina
  `web/administrar_empresa/configuracion/email_corporativo.html` permite cambiar
  esa preferencia y actualizar la contrasena interna del buzon; la clave siempre
  se guarda cifrada y no se muestra al usuario. La configuracion global del
  servidor de email define tambien `max_accounts_per_empresa`, con default 5,
  para limitar desde backend cuantas cuentas corporativas puede tener una misma
  empresa.
- Informacion de modulos del index: `web/super/informacion_de_modulos.html`,
  `/super/api/informacion_de_modulos`,
  `/api/public/informacion_de_modulos`.
- Noticias del portal: `web/noticias.html`, editor
  `web/super/noticias.html`, `/super/api/noticias` y
  `/api/public/noticias`. Se guarda en
  `pcs_superadministrador.configuraciones` con la clave
  `super.noticias_portal.v1`; la pagina publica se abre desde el menu flotante
  y tiene portada, foto de perfil y publicaciones tipo red social.
- Portal publico e index en super administrador: el menu lateral de
  `web/super_administrador.html` agrupa tarjetas del index
  (`web/super/pagina_principal.html`), modulos del index
  (`web/super/informacion_de_modulos.html`), noticias
  (`web/super/noticias.html`), descripcion de sistemas para IA y portal
  (`web/super/informacion_de_la_empresa_y_de_los_sistemas_para_ia.html`),
  WhatsApp del portal (`web/super/configuracion/whatsapp_portal.html`) y accesos
  de lectura a `web/index.html` y `web/descripcion_de_los_sistemas.html`.
- Energia solar: `web/administrar_empresa/energia_solar.html`,
  `web/js/energia_solar.js`, `/api/empresa/energia_solar`,
  tablas `empresa_energia_solar_*`. El modulo es por empresa, usa permiso
  `energia_solar`, soporta Victron/SMA/SolarEdge/gateway local y alerta por
  correo usando SMTP configurado. Las preconfiguraciones por tipo incluyen
  `modulos.energia_solar` apagado por defecto, con catalogo de proveedores,
  baterias y alertas; el rol `tecnico_solar` solo recibe lectura.
- Analitica publica por pais: `/api/public/portal_visitas`,
  `web/js/portal_visits.js`.
- Chat/robot/emisora flotante: `web/js/ai_chat_drawer.js`,
  `web/js/radio_player.js`, `web/js/radio_online.js`,
  `/api/chat_flotante/preferencias`. En contexto empresarial, robot/secretaria
  IA 3D y emisora online deben iniciar apagados salvo preferencia explicita por
  `empresa_id`; no deben prenderse por configuracion global ni por
  `localStorage` viejo. Mientras el proyecto siga en preproduccion, el arranque
  puede limpiar preferencias antiguas encendidas para dejar el default en cero.
  Las preconfiguraciones por tipo tambien deben guardar/aplicar
  `asistente_ia.robot_enabled=false` y `asistente_ia.radio_online_enabled=false`.

## Flujo de login

El usuario entra por `login.html`. El backend valida credenciales u OAuth Google
en handlers de autenticacion administrativa. El registro de administradores usa
la pagina de registro y crea cuentas administrativas con confirmacion segun la
configuracion vigente. Las alertas super administrador pueden enviar correo
cuando se registra un administrador, sin incluir contrasenas ni tokens.

Los usuarios operativos entran por `login_usuario.html`. El acceso por correo,
contrasena o Google siempre debe resolver un usuario ya creado/invitado por una
empresa; no existe alta publica operativa. Para Google se usa
`/auth/google/usuario/login`, que marca el flujo como usuario y vuelve por el
callback canonico `/auth/google/callback`; el callback solo abre sesion si el
correo verificado por Google coincide con una invitacion vigente o con un
usuario empresarial ya confirmado. La sesion redirige a
`administrar_empresa.html?id={empresa_id}` para que el panel cargue roles y
permisos efectivos de esa empresa.

## Flujo de creacion de empresa

Desde `seleccionar_empresa.html`, el administrador crea una empresa eligiendo
tipo. El backend crea la empresa, aplica preconfiguracion por tipo, prepara
permisos/modulos y puede activar una licencia gratis de 15 dias si corresponde.
La licencia gratis solo puede usarse una vez por empresa y pertenece al catalogo
global compartido por todos los tipos de empresa. El bloqueo usa historial de
activaciones y licencias gratis antiguas, no solo licencias vigentes. La
creacion de empresa puede disparar una alerta por correo al super administrador
si el check esta activo.

## Flujo de administradores delegados

El administrador principal normal puede abrir `Administradores` desde
`seleccionar_empresa.html`. Esa pagina no es global para el: `/super/api/administradores`
filtra por el principal resuelto, excluye al propio principal y permite gestionar
solo cuentas con `administradores.usuario_creador` dentro de su alcance. El alta
se hace por invitacion: se crea cuenta pendiente, se envia correo con
`invitation_token`, el invitado completa `registrar_nuevo_usuario_administrador.html`
y solo despues de validar el token queda confirmado para login. Los delegados
heredan acceso a las empresas creadas por el principal como `access_source=delegated`,
pero no pueden compartirlas ni administrar otros administradores. La validacion
real vive en backend y en `CanAdminAccessEmpresaIA`.

## Flujo de carrito, venta, caja y facturacion

- Venta directa usa el carrito canonico de venta directa de la empresa.
- Estaciones usan carritos asociados a cada estacion.
- El carrito agrega productos, servicios o recetas, calcula totales, abonos,
  pagos mixtos y medio de pago.
- Cada usuario/caja debe operar de forma independiente dentro de la misma
  empresa.
- Las cajas fisicas configurables de la empresa se guardan en
  `empresa_estacion_prefs.estaciones_config.cajas_config` con `codigo`,
  `nombre`, `descripcion` y `activa`; la estacion Caja y el carrito muestran el
  nombre operativo, por ejemplo `CAJA-1 - FRUTERA`.
- El pago cierra el carrito, actualiza inventario/caja, genera documento
  imprimible y, si aplica, documento electronico.
- Caja y corte usan `corte_de_caja.html`; el reporte de turno se calcula por
  usuario/caja/turno y se imprime por defecto en POS 80mm.
- Los documentos imprimibles, facturas, recibos, notas y reportes fiscales deben
  verse como papel real en blanco y negro, sin depender de tema claro u oscuro.
- Si esta activo, el QR DIAN al final de factura/recibo se genera desde CUFE,
  CUDE o codigo de validacion.

## Donde se guardan configuraciones frecuentes

- Configuracion visual/operativa del carrito:
  `empresa_estacion_prefs.estaciones_config.carrito_ui_global`.
- Configuracion chat flotante/robot/emisora:
  claves `chat_flotante.*` en `empresa_estacion_prefs` con `estacion_id=0`.
  Robot/secretaria y emisora son opt-in por empresa.
- Overrides por estacion:
  `empresa_estacion_prefs.estaciones_config.estaciones[].carrito.configuracion`.
- Configuracion de estaciones y nombres singular/plural:
  `empresa_estacion_prefs.estaciones_config`.
- Catalogo de cajas fisicas simultaneas:
  `empresa_estacion_prefs.estaciones_config.cajas_config`.
- Configuracion de corte/reporte de caja:
  `empresa_corte_caja_configuracion`.
- Impresoras por empresa y POS 80mm:
  `empresa_impresoras*`.
- Facturacion electronica por pais:
  `facturacion_electronica_pais`.
- Reintentos/cola documental electronica:
  `facturacion_electronica_reintentos`.
- Configuracion super global:
  `pcs_superadministrador.configuraciones` o tablas super dedicadas.
- Alertas sistema:
  `super_alertas_config` y `super_alertas_eventos`.

## Scripts reales

- `.\rs.ps1`: wrapper operativo principal solicitado por el usuario para
  preflight, actualizacion/sincronizacion y tareas de runtime segun el script.
- `.\scripts\rs.ps1`: script base relacionado con el flujo `rs`.
- `.\scripts\sync_to_vps.ps1`: sincroniza hacia VPS.
- `.\scripts\sync_to_vps.sh`: alternativa shell para sincronizacion.
- `.\scripts\profesional_preflight.ps1`: validaciones previas.
- `.\scripts\actualizar_repositorio.ps1`: actualizacion de repositorio.
- `.\scripts\publicar_git_y_vps.ps1`: publicacion coordinada Git/VPS.

Antes de ejecutar scripts operativos revisar `documentos/comandos_codex.md`.

## Datos de prueba permitidos

- Empresa de prueba para motel/POS/estaciones/caja: `Motel Calipso`.
- Empresa de prueba para creditos de motos: `Venta Moto`.
- Usuario administrativo de prueba autorizado por el usuario:
  `powerfulcontrolsystem@gmail.com`. No repetir ni guardar claves en
  documentacion, consola o commits.
- Cuando se creen datos de prueba, dejar claro si son demo/preproduccion y no
  mezclar empresas sin validar `empresa_id`.

## Seguridad que siempre debe conservarse

- Validar `empresa_id` en backend y en consultas SQL.
- No permitir que una empresa lea, edite o borre datos de otra.
- Antes de crear, modificar o revisar endpoints empresariales, aplicar
  `documentos/checklist_seguridad_endpoint_multiempresa.md`.
- No imprimir secretos ni credenciales.
- Mantener auditoria en operaciones criticas: caja, pagos, facturacion,
  licencias, usuarios, backups, conectividad y cambios de configuracion.
- En tareas de limpieza, backup o reinicio de datos, conservar configuracion,
  usuarios, permisos e integraciones salvo instruccion explicita.
