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
- Licencia gratis: `web/pagar_licencia.html`, `/licencias/activar_sin_pago`.
- Panel empresarial: `web/administrar_empresa.html`,
  `web/administrar_empresa/panel.html`.
- Configuracion empresa: `web/administrar_empresa/configuracion_menu.html` y
  paginas bajo `web/administrar_empresa/configuracion/`.
- Configuracion carrito: `web/administrar_empresa/configuracion_carrito_de_compra_empresa.html`,
  `/api/empresa/estacion_prefs`.
- Estaciones: `web/administrar_empresa/estaciones.html`,
  `/api/empresa/carritos_compra`.
- Carrito y venta directa: `web/administrar_empresa/carrito_de_compras.html`,
  `/api/empresa/carritos_compra`, `/api/empresa/carritos_compra/items`.
- Caja y corte: `web/administrar_empresa/corte_de_caja.html`,
  `/api/empresa/corte_caja`, `/api/empresa/corte_caja/configuracion`.
- Reportes de turnos: `web/administrar_empresa/reportes_turnos.html`,
  `/api/empresa/corte_caja?action=turnos|turno_reporte|turno_export|turno_email`.
- Facturacion electronica: `web/administrar_empresa/facturacion_electronica_menu.html`,
  `/api/empresa/facturacion_electronica`,
  `/api/empresa/facturacion_electronica/panama`,
  `/api/empresa/facturacion_electronica/ecuador`.
- Facturacion offline: `/api/empresa/offline_ventas`,
  `backend/db/offline_ventas.go`.
- Alertas sistema super administrador: `web/super/alertas_sistema.html`,
  `/super/api/alertas_sistema`.
- Email corporativo iRedMail: `web/super/email_corporativo.html`,
  `/super/api/email_corporativo`, `/api/empresa/email_corporativo`,
  `backend/handlers/email_corporativo_handlers.go`,
  `backend/db/email_corporativo.go`, `deploy/iredmail/`.
  El modulo genera email unico por empresa al crearla; la provision real por API
  requiere iRedAdmin-Pro REST API y credenciales cifradas. En Docker portable,
  `deploy/.env.platform` provee `EMAIL_CORPORATIVO_*` / `IREDMAIL_*` y el
  backend las registra en `configuraciones` al arrancar. La pagina de super
  administrador incluye diagnostico operativo y `Probar iRedAdmin` para validar
  URL, usuario y clave sin exponer secretos ni crear buzones.
- Informacion de modulos del index: `web/super/informacion_de_modulos.html`,
  `/super/api/informacion_de_modulos`,
  `/api/public/informacion_de_modulos`.
- Energia solar: `web/administrar_empresa/energia_solar.html`,
  `web/js/energia_solar.js`, `/api/empresa/energia_solar`,
  tablas `empresa_energia_solar_*`. El modulo es por empresa, usa permiso
  `energia_solar`, soporta Victron/SMA/SolarEdge/gateway local y alerta por
  correo usando SMTP configurado.
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

## Flujo de creacion de empresa

Desde `seleccionar_empresa.html`, el administrador crea una empresa eligiendo
tipo. El backend crea la empresa, aplica preconfiguracion por tipo, prepara
permisos/modulos y puede activar una licencia gratis de 15 dias si corresponde.
La licencia gratis solo puede usarse una vez por empresa. La creacion de empresa
puede disparar una alerta por correo al super administrador si el check esta
activo.

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
