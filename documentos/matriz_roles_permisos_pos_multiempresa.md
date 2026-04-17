# Matriz base de roles y permisos POS multiempresa

Fecha de actualizacion: 2026-04-17
Alcance: punto 3 del plan maestro (permisos y seguridad)

## Regla de mantenimiento por modulo

- Cuando se cree un modulo nuevo o se modifique uno existente, esta matriz debe actualizarse en la misma iteracion para reflejar permisos por rol/modulo/accion y el impacto en paginas del panel.
- Esta actualizacion debe quedar sincronizada con `documentos/descripcion_de_modulos`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios` y `CHANGELOG.md`.

- Actualizacion 2026-04-17 (navegacion general: misma pestaña por defecto):
	- Se retiran aperturas automáticas en nueva ventana para navegación normal entre módulos, portales públicos, ayudas y exportes comunes del sistema.
	- Los reportes/exportes de `Clientes`, `Asistencia`, `Backups`, `Tarifas por día` y `Soporte remoto` descargan el archivo sin sacar al usuario del módulo actual.
	- Se conservan como excepción los documentos legales (`contrato`, términos de pasarela) y los popups técnicos de impresión o vista previa documental.
	- Impacto de matriz: sin cambios en roles ni permisos; solo cambia el comportamiento de navegación de rutas ya permitidas.

- Actualizacion 2026-04-17 (licencias super: valor 0 visible y editable):
	- `web/super/licencias.html` conserva el valor `0` en listado y formulario de edicion para el CRUD de licencias.
	- Impacto de matriz: sin cambios en permisos; solo corrige el comportamiento visual del modulo `Licencias` para `super_administrador`.

- Actualizacion 2026-04-17 (licencias del selector: historial y estado con vencimiento):
	- `backend/db/db.go` y `web/super/licencias.html` convierten la vista `scope=mine&con_empresa=1` en un historial de licencias pagadas/asignadas, con fecha de vencimiento, estados `activa/por vencer/vencida` y CTA de renovacion sin opciones de eliminar en esa pantalla.
	- `backend/handlers/payments_handlers_test.go` valida que el endpoint protegido siga filtrando por creador y entregue empresa + fechas al frontend.
	- Impacto de matriz: sin cambios en roles ni permisos; el modulo `Licencias` mantiene acceso de lectura/gestion para `super_administrador`, pero en el flujo del selector la misma ruta se presenta como historial restringido al alcance del administrador autenticado.

- Actualizacion 2026-04-17 (checkout publico de licencias: Epayco usa Smart Checkout v2):
	- `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `web/pagar_licencia.html` y `web/super/configuracion_avanzada.html` migran Epayco desde el flujo manual `checkout.php` al Smart Checkout oficial con `sessionId` y `checkout-v2.js`.
	- Impacto de matriz: sin cambios en permisos; `/epayco/create_transaction`, `/epayco/transaction_status`, `/epayco/webhook` y `/epayco/respuesta.html` mantienen alcance publico dentro del checkout comercial, mientras la configuracion avanzada sigue siendo exclusiva de `super_administrador`.

- Actualizacion 2026-04-17 (crear clave por correo: visibilidad de contrasena):
	- `web/registrar_contrasena_usuario_de_google.html`, `web/js/registrar_contrasena_usuario_de_google.js` y `web/estilos.css` agregan el control visual para mostrar u ocultar la contrasena antes de guardarla.
	- Impacto de matriz: sin cambios en permisos; sigue siendo el mismo flujo autenticado de autogestion posterior al login con Google.

- Actualizacion 2026-04-17 (licencias publicas: tarjetas con estilo del home):
	- `web/elegir_licencia.html` reutiliza la estructura visual de `index.html` para las tarjetas de licencias, sin cambiar rutas, protecciones ni acciones disponibles.
	- Impacto de matriz: sin cambios en permisos; la vista y el flujo de compra conservan el mismo alcance previo.

- Actualizacion 2026-04-17 (reportes globales super: una empresa o varias):
	- `web/super/reportes_globales.html` y `web/js/super_reportes_globales.js` agregan un selector de alcance para alternar entre analisis de una sola empresa o de varias empresas del mismo administrador.
	- `backend/handlers/reportes_globales_test.go` cubre el uso de `empresa_id` en la misma API protegida.
	- Impacto de matriz: sin cambios en permisos; `Reportes globales (super)` se mantiene como permiso `R` exclusivo de `super_administrador`.

- Actualizacion 2026-04-17 (autenticacion administrativa: login en una sola tarjeta visual):
	- `web/login.html` y `web/estilos.css` integran el formulario por correo dentro de la misma tarjeta del acceso con Google, retirando el recuadro secundario del flujo publico de administradores.
	- Impacto de matriz: sin cambios en permisos, roles, wrappers o visibilidad; el login administrativo sigue siendo publico y el panel super sigue reservado a `super_administrador`.

- Actualizacion 2026-04-16 (reportes globales super por administrador creador):
	- `backend/handlers/reportes_globales.go` expone `/super/api/reportes_globales` filtrando empresas por `usuario_creador = admin autenticado`.
	- `web/super/reportes_globales.html` y `web/js/super_reportes_globales.js` permiten ver datasets consolidados o separados por empresa solo dentro del panel super.
	- Impacto de matriz: el modulo `Reportes globales (super)` queda con permiso `R` exclusivo de `super_administrador`.

- Actualizacion 2026-04-17 (portal publico: arcade movil con runtime comun de poderes y premios):
	- `web/Juegos/arcade_shared.js` y `web/Juegos/arcade_window.css` pasan a ser la base comun del arcade publico para countdown, sonido, records, poderes y premios en todos los juegos activos.
	- Los nueve juegos `*_plus.html` del lobby reutilizan el mismo runtime sin ampliar rutas privadas ni introducir permisos nuevos.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni wrappers; `Portal publico - Juegos` mantiene acceso publico y de solo uso.

- Actualizacion 2026-04-17 (reportes globales super: graficos y lectura ejecutiva):
	- `web/super/reportes_globales.html`, `web/js/super_reportes_globales.js` y `web/estilos.css` agregan visualizaciones ejecutivas sobre el mismo modulo protegido de lectura.
	- Impacto de matriz: sin cambios en permisos; `Reportes globales (super)` se mantiene como `R` exclusivo de `super_administrador`.

- Actualizacion 2026-04-16 (facturacion electronica: estabilidad de pruebas automatizadas):
	- `backend/db/finanzas_test.go` fuerza el dialecto `sqlite` en `openFinanzasTestDB` para evitar que la suite del modulo herede configuracion `postgres` del entorno local y falle por compatibilidad SQL durante pruebas de esquema y documentos transaccionales.
	- Impacto de matriz: sin cambios en permisos, roles, wrappers o visibilidad del modulo `facturacion electronica`; el ajuste solo endurece la validacion automatizada del backend.

- Actualizacion 2026-04-16 (portal publico: menu flotante navegable en movil):
	- `web/menu.js` conserva la navegacion tactil de las rutas publicas al cerrar el panel solo en `click`.
	- `web/estilos.css` mejora la respuesta tactil del boton y de los items del menu.
	- Impacto de matriz: sin cambios en permisos o roles; solo se recupera la usabilidad movil de enlaces publicos ya visibles.

- Actualizacion 2026-04-16 (menu flotante publico: separacion frente a botones de cabecera):
	- `web/menu.js` marca las paginas que reciben el menu flotante con `has-floating-menu` y `web/estilos.css` reserva espacio en encabezados y barras de acciones compartidas para que el toggle no quede encima de botones cercanos.
	- Impacto de matriz: sin cambios en permisos, CRUD/A, wrappers o visibilidad por rol; la correccion es solo de layout compartido.

- Actualizacion 2026-04-16 (seleccionar empresa: formato clasico restaurado):
	- `web/js/seleccionar_empresa.js` recupera el render clasico de tarjetas del panel super sin alterar rutas ni acciones del modulo.
	- Impacto de matriz: sin cambios en permisos, roles o wrappers; solo se restituye una presentacion visual previa del selector de empresas.

- Actualizacion 2026-04-17 (portal publico de usuarios de empresa con contrato y subdominio propio):
	- `backend/handlers/usuarios_empresa.go` exige aceptacion del contrato vigente antes de login, primer password, reset o cambio de contrasena; la restriccion aplica al mismo usuario autenticandose, no a un rol administrativo nuevo.
	- `web/login_usuario.html` y `web/js/login_usuario.js` pasan a ser una puerta publica de acceso por invitacion, mientras `web/administrar_empresa.html` y `web/js/administrar_empresa.js` resuelven el enlace correcto por empresa usando `empresa_slug` o `dominio_publico` de la configuracion publica.
	- Impacto de matriz: sin cambios en CRUD/A ni en wrappers empresariales; el portal publico no amplía privilegios y la visibilidad final del panel sigue determinada por rol/permisos_contexto ya existentes.

- Actualizacion 2026-04-16 (estaciones: carrito base sincronizado desde backend):
	- `backend/handlers/empresa_estacion_prefs.go` sincroniza automaticamente los carritos enlazados al guardar `estaciones_config`, y `backend/db/empresa_estacion_prefs.go` asegura la regla `una estacion -> un carrito base` por `empresa_id`.
	- `backend/handlers/empresa_estacion_prefs_test.go` y `backend/db/empresa_estacion_prefs_test.go` cubren la creacion y actualizacion del enlace sin depender de la pagina de configuracion.
	- Impacto de matriz: sin cambios en roles ni permisos; la capacidad sigue dentro del alcance actual de configuracion/ventas empresariales y no abre nuevas rutas ni acciones a otros roles.

- Actualizacion 2026-04-16 (portal publico: contacto debajo del grid del home):
	- `web/index.html` mueve `Informacion de contacto` debajo de `portalCardsGrid` y `web/estilos.css` lo centra sin alterar accesos publicos ni rutas protegidas.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni wrappers; el ajuste es visual dentro del portal publico.

- Actualizacion 2026-04-17 (soporte remoto: limites por plan y mesa tecnica super):
	- `backend/handlers/soporte_remoto.go` mantiene el alcance empresarial en modulo `seguridad`, pero ahora devuelve `uso` y bloquea sesiones/dispositivos cuando la empresa supera los topes configurados.
	- `backend/handlers/super_soporte_remoto.go` expone `/super/api/soporte_remoto` y `web/super/soporte_remoto.html` agrega una mesa tecnica central solo para `super_administrador`.
	- Impacto de matriz: `linkSoporteRemoto` sigue requiriendo accion `A` sobre `seguridad` en panel empresa; el nuevo panel super de soporte remoto es exclusivo de `super_administrador` y no amplía permisos de roles empresariales.

- Actualizacion 2026-04-16 (deploy VPS: limpieza de procesos previos del backend):
	- `scripts/sync_to_vps.ps1` limpia procesos previos asociados a `backend/bin/server_linux_amd64` antes del arranque y genera una unidad `systemd` sin el warning de clave invalida en `Service`.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni visibilidad; el ajuste es operativo de infraestructura.

- Actualizacion 2026-04-16 (checkout publico de licencias: alias `sambox` en Epayco):
	- `backend/handlers/payments_handlers.go` normaliza `epayco.mode=sambox` como `sandbox` antes de construir el checkout publico.
	- Impacto de matriz: sin cambios en roles ni permisos; `/epayco/create_transaction` conserva el mismo alcance publico.

- Actualizacion 2026-04-16 (portal publico: arcade activo con ocho juegos):
	- `web/Juegos/menu_juegos.html` publica ocho juegos activos y fija popup uniforme `700x700` sin barras en escritorio, manteniendo apertura directa en movil.
	- `web/Juegos/arcade_window.css` y los ocho juegos `*_plus.html` mantienen una experiencia publica homogénea con pausa real, records locales y nombre de jugador compartido.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni wrappers; `Portal publico - Juegos` sigue siendo de lectura/uso para todos los roles y tambien sin autenticacion.

- Actualizacion 2026-04-17 (portal publico: nuevo Ajedrez 3D plus):
	- `web/Juegos/ajedrez_3d_plus.html` agrega una nueva ruta publica del arcade con tablero en perspectiva 3D simulada y selector de cinco dificultades.
	- `web/Juegos/menu_juegos.html` publica la nueva tarjeta del lobby y `web/img/juegos/ajedrez_3d.svg` suma la portada visual del titulo.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni wrappers; `Portal publico - Juegos` mantiene acceso publico y de solo uso.

- Actualizacion 2026-04-16 (checkout publico de licencias: metodo unico y compatibilidad Epayco legacy):
	- `web/pagar_licencia.html` omite el selector de forma de pago cuando solo hay una pasarela disponible y entra directo al panel correspondiente.
	- `backend/handlers/payments_handlers.go` añade `p_key` al checkout de Epayco cuando existe `epayco.private_key`, manteniendo el mismo alcance publico de `/epayco/*` y `/api/public/licencias/payment_methods`.
	- Impacto de matriz: sin cambios en roles ni permisos; el ajuste es funcional en checkout publico.

- Actualizacion 2026-04-16 (checkout publico de licencias: Epayco sin popup intermedio):
	- `web/pagar_licencia.html` ya no deja el pago en una pestaña emergente; ahora redirige la misma pestaña al checkout de Epayco y reutiliza `/epayco/respuesta.html` para volver con contexto a la pantalla de licencia.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad por rol; `/epayco/*` sigue siendo publico y de solo consumo para el flujo comercial.

- Actualizacion 2026-04-16 (home publico: ajuste visual de accesos superiores):
	- `index.html` mantiene exactamente la misma visibilidad publica; el cambio solo compacta y centra los botones superiores en movil.
	- Impacto de matriz: sin cambios en permisos o acceso por rol.

- Actualizacion 2026-04-16 (licencias super: persistencia del valor):
	- `super/licencias.html` y `/super/api/licencias` mantienen los mismos permisos y alcance, pero ahora exponen correctamente los errores de guardado al administrador.
	- Impacto de matriz: sin cambios en permisos, roles ni visibilidad; se corrige comportamiento del CRUD super existente.

- Actualizacion 2026-04-16 (seleccionar empresa: mejora visual de tarjetas):
	- `seleccionar_empresa.html` mantiene el mismo acceso por rol y licencia, pero las tarjetas internas ahora se adaptan mejor al contenido variable.
	- Impacto de matriz: sin cambios en permisos, CRUD/A ni visibilidad de modulo; la correccion es exclusivamente visual y de lectura.

- Actualizacion 2026-04-16 (pagina principal dinamica: correccion de render en editor super):
	- `web/super/pagina_principal.html` respeta la cantidad persistida de tarjetas al recargar el editor.
	- Impacto de matriz: sin cambios en permisos, roles ni visibilidad de modulo; la correccion solo evita un recorte incorrecto en el panel super.

- Actualizacion 2026-04-16 (pagina principal dinamica: sin parpadeo inicial del campo cantidad):
	- `web/super/pagina_principal.html` deja el campo `Cantidad de tarjetas` en estado de carga hasta recibir la configuracion real, evitando mostrar de forma transitoria el valor `5` que no corresponde a la configuracion activa.
	- El mismo editor sincroniza la cantidad con el numero real de tarjetas persistidas, manteniendo alineacion con `index.html` y `/descripcion_de_los_sistemas.ht`.
	- Impacto de matriz: sin cambios en permisos, CRUD/A, wrappers o visibilidad por rol; `Pagina principal (tarjetas index)` sigue siendo CRUA exclusivo de `super_administrador`.

- Actualizacion 2026-04-16 (infraestructura publica: wildcard HTTPS y subdominio venta digital):
	- `venta-digital.powerfulcontrolsystem.com` queda publicado como acceso publico HTTPS hacia la pagina global `venta_digital.html`.
	- La raiz generica de subdominios por empresa sigue en `venta_publica.html`; no cambia la matriz de permisos ni la visibilidad de modulos internos.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni permisos; la modificacion solo amplía una entrada publica servida por infraestructura.

- Actualizacion 2026-04-16 (autenticacion administrativa: registro con pais y ciudad):
	- `web/registrar_nuevo_usuario_administrador.html` y `web/js/registrar_nuevo_usuario_administrador.js` amplian el formulario de alta administrativa para capturar `pais` y `ciudad`.
	- `backend/handlers/auth_admin_handlers.go` mantiene la regla de confirmacion obligatoria del correo antes del ingreso, y `backend/db/db.go` agrega persistencia evolutiva de esos datos en `administradores`.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni visibilidad de modulos; el ajuste afecta solo los datos del registro y la completitud del perfil administrativo.

- Actualizacion 2026-04-16 (autenticacion administrativa: esquema `administradores` compatible con PostgreSQL):
	- `backend/db/db.go` agrega una regularizacion reusable de columnas de seguridad de `administradores` y `backend/main.go` la invoca en el arranque del backend.
	- `backend/db/administradores_auth_schema_test.go` cubre la reparacion del esquema y la escritura de password inicial sobre tablas incompletas.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni visibilidad; el ajuste es de compatibilidad de backend para el mismo flujo administrativo autenticado.

- Actualizacion 2026-04-16 (super: seguridad VPS Linux):
	- `backend/handlers/security_vps_handlers.go` expone la API protegida `/super/api/security/vps/config|run|status|history|report|compare` y `backend/main.go` registra el servicio central del modulo.
	- `web/super/seguridad.html` y `web/js/super_seguridad.js` agregan la vista operativa para configurar, ejecutar, revisar historial, comparar y exportar reportes del VPS.
	- `backend/tools/vps_security_scan/main.go` y los scripts Linux asociados permiten ejecutar el mismo modulo desde consola y programarlo por cron sin abrir acceso a otros roles.
	- Impacto de matriz: se agrega `Seguridad VPS Linux (super)` como modulo `CRUA` exclusivo de `super_administrador`; no hay acceso para roles empresariales.

- Actualizacion 2026-04-16 (portal publico: orden visual del header):
	- `web/index.html` y `web/estilos.css` reordenan el header del home para dejar `Informacion de contacto` al extremo derecho de la misma fila.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni wrappers; el ajuste es visual y no altera autorizacion.

- Actualizacion 2026-04-16 (autenticacion estable multi-host sin recordar usuario/cuenta):
	- `web/login.html` y `web/login_usuario.html` eliminan los checkboxes de `Recordar cuenta` y `Recordar usuario`, reduciendo divergencias por almacenamiento local entre `localhost`, dominio raíz y `www`.
	- `backend/handlers/auth_admin_handlers.go` deja de propagar `login_hint` en el inicio OAuth; el login Google arranca limpio y consistente.
	- `web/menu.js`, `web/js/super_administrador.js`, `web/js/seleccionar_empresa.js`, `web/super/licencias.html` y `web/super/tipos_empresas.html` retiran lógica `remember*` y conservan solo señal de sesión para navegación/autenticación visible.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni wrappers; la modificación es operativa/UX y no amplía privilegios.

- Actualizacion 2026-04-16 (autenticacion administrativa: registro separado y recuperacion guiada):
	- `web/login.html` mantiene acceso publico por Google o correo/clave, pero mueve el registro administrativo a `/registrar_nuevo_usuario_administrador.html` y deja la recuperación en formularios propios dentro del login.
	- `backend/handlers/auth_admin_handlers.go` endurece el alta y la recuperación de administradores, mientras `backend/utils/utils.go` libera `/registrar_nuevo_usuario_administrador.html` y `/auth/confirmar_admin` como rutas públicas reales.
	- `backend/handlers/auth_admin_handlers_test.go` y `backend/handlers/auth_users_carritos_test.go` cubren el alta/login/reset administrativo y la nueva superficie pública del middleware.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad por rol; el login/registro/confirmación administrativa sigue siendo público y la administración global continúa bajo `super_administrador`.

- Actualizacion 2026-04-16 (autenticacion administrativa: creación de clave local tras Google):
	- `backend/handlers/auth_admin_handlers.go` y `backend/handlers/accept_handlers.go` redirigen a `/registrar_contrasena_usuario_de_google.html` cuando la cuenta autenticada por Google todavía no tiene `password_set`.
	- `backend/handlers/account_handlers.go` expone `/api/account/set_google_password` como endpoint autenticado de solo autoservicio para el administrador en sesión.
	- `web/registrar_contrasena_usuario_de_google.html` completa el alta de contraseña local sin ampliar permisos ni abrir una nueva superficie pública.
	- Impacto de matriz: sin cambios en roles, CRUD/A o wrappers; la capacidad sigue restringida al mismo administrador autenticado sobre su propia cuenta.

- Actualizacion 2026-04-16 (Epayco: respuesta publica fija):
	- `web/epayco/respuesta.html` queda disponible como pagina publica para retorno desde la pasarela.
	- `backend/handlers/payments_handlers.go` usa `/epayco/respuesta.html` como `response` y `/epayco/webhook` como `confirmation` para licencias.
	- Impacto de matriz: sin cambios en roles ni permisos internos; ambas rutas siguen siendo publicas por integracion de pasarela.

- Actualizacion 2026-04-16 (checkout visual y seleccion de empresa):
	- `web/pagar_licencia.html` preselecciona de forma visible la unica pasarela disponible y muestra el logo de Epayco en tarjeta y panel cuando corresponde.
	- `web/js/seleccionar_empresa.js` vuelve al formato compacto previo para tarjetas de empresa, sin alterar accesos ni permisos.
	- Impacto de matriz: sin cambios en roles, CRUD/A o visibilidad funcional; las pantallas siguen con el mismo alcance de acceso que antes.

- Actualizacion 2026-04-16 (super: tamano estimado por empresa en administracion PostgreSQL):
	- `web/super/administrar_base_de_datos.html` agrega una tarjeta operativa con el boton `Cargar Empresas` para consultar consumo estimado por empresa dentro de `pcs_empresas`.
	- `backend/handlers/postgres_performance.go` extiende `/super/api/postgres/performance` con `action=empresas_storage`, manteniendo el endpoint protegido y de solo lectura.
	- `backend/handlers/postgres_performance_test.go` cubre la accion invalida del panel y utilidades asociadas.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad; `Administracion DB PostgreSQL (super)` sigue siendo lectura exclusiva de `super_administrador`.

## Roles base

| Rol | Alcance | Descripcion |
|---|---|---|
| super_administrador | global | administra configuracion, empresas, licencias, auditoria y seguridad global |
| admin_empresa | empresa | administra configuracion, catalogos, usuarios y cierres de su empresa |
| supervisor_sucursal | sucursal | supervisa operacion, aprueba cierres y movimientos criticos |
| cajero | sucursal/caja | registra ventas, cobros, devoluciones permitidas y cierre de caja |
| inventario | sucursal/bodega | gestiona productos, existencias y movimientos de bodega |
| compras | empresa/sucursal | crea ordenes de compra, recepciones y ajustes de costo |
| contabilidad | empresa | valida asientos, periodos y reportes financieros |
| auditor | empresa/global | consulta reportes, logs y trazabilidad sin modificar datos |

## Permisos por modulo

Leyenda:
- C: crear
- R: leer
- U: actualizar
- D: eliminar/anular
- A: aprobar/cerrar

| Modulo | super_administrador | admin_empresa | supervisor_sucursal | cajero | inventario | compras | contabilidad | auditor |
|---|---|---|---|---|---|---|---|---|
| Ventas POS | CRUDA | CRUA | CRUA | CRU | R | R | R | R |
| Inventarios | CRUDA | CRUA | CRUA | R | CRUDA | R | R | R |
| Clientes | CRUDA | CRUA | CRUA | CRU | R | R | R | R |
| Proveedores | CRUDA | CRUA | R | R | R | CRUA | R | R |
| Compras | CRUDA | CRUA | CRUA | R | R | CRUDA | R | R |
| Facturacion electronica | CRUDA | CRUA | R | CRU | R | R | R | R |
| Contabilidad y periodos | CRUDA | CRUA | R | R | R | R | CRUDA | R |
| Reportes financieros | CRUA | CRUA | R | R | R | R | CRUA | R |
| Cierres de caja | CRUDA | CRUA | CRUA | CRUA | R | R | R | R |
| Seguridad y permisos | CRUDA | CRUA | R | R | R | R | R | R |
| Impresoras operativas | CRUDA | CRUA | R | R | R | R | R | R |
| Seguridad VPS Linux (super) | CRUA | - | - | - | - | - | - | - |
| Administracion DB PostgreSQL (super) | R | - | - | - | - | - | - | - |
| Reportes globales (super) | R | - | - | - | - | - | - | - |
| Pagina principal (tarjetas index) | CRUA | - | - | - | - | - | - | - |
| Portal publico - Juegos | R | R | R | R | R | R | R | R |
| Contrato administrativo (super) | CRUA | - | - | - | - | - | - | - |
| Monitor de errores del sistema (super) | R | - | - | - | - | - | - | - |
| Pasarelas de licencias (Wompi/Epayco) | CRUA | - | - | - | - | - | - | - |

## Estado de implementacion tecnica inicial (2026-04-04)

- Actualizacion 2026-04-16 (super: seguridad VPS Linux):
	- `web/super/seguridad.html` amplía el monitor de seguridad del panel super para cubrir configuracion, ejecucion de escaneo, hallazgos, historial, comparacion y exportes del VPS.
	- `backend/handlers/security_vps_handlers.go` y `backend/vpssecurity/*` mantienen el modulo encapsulado y protegido solo para `super_administrador`.
	- `backend/tools/vps_security_scan/main.go` junto a los scripts Linux permiten operacion manual y por cron sin ampliar privilegios a otros roles.
	- Impacto de matriz: nuevo modulo `Seguridad VPS Linux (super)` con `CRUA` exclusivo de `super_administrador`; sin cambios para roles de empresa.

- Actualizacion 2026-04-16 (portal publico: boton de contacto al extremo derecho del home):
	- `web/index.html` y `web/estilos.css` ajustan solo la composicion visual del header comercial.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad por rol.

- Actualizacion 2026-04-16 (autenticacion administrativa: registro separado y recuperacion guiada):
	- `web/login.html` centra el acceso por correo, deja debajo `Registrarse` y `¿Olvidó su contraseña?`, y sustituye los `prompt()` por formularios reales para recuperación y restablecimiento.
	- `web/registrar_nuevo_usuario_administrador.html` agrega una superficie pública específica para alta administrativa y `backend/utils/utils.go` la libera junto con `/auth/confirmar_admin`.
	- `backend/handlers/auth_admin_handlers.go` evita sobrescribir cuentas confirmadas y exige `nombre`, `telefono` y contraseña mínima para el registro administrativo.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni wrappers; el ajuste corrige el flujo público de autenticación administrativa sin ampliar permisos.

- Actualizacion 2026-04-16 (super: tamano estimado por empresa en administracion PostgreSQL):
	- `web/super/administrar_base_de_datos.html` suma una lectura puntual del peso estimado por empresa en la base operativa compartida y la presenta ordenada de mayor a menor.
	- `backend/handlers/postgres_performance.go` mantiene el mismo modulo y endpoint protegido, agregando solo una accion de solo lectura (`action=empresas_storage`).
	- Impacto de matriz: sin cambios en roles, CRUD/A ni wrappers; la consulta sigue reservada a `super_administrador`.

- Actualizacion 2026-04-16 (portal publico: arcade movil reforzado y countdown en Patito volando):
	- `web/Juegos/arcade_shared.js` suma sonidos de cuenta regresiva reutilizables por el arcade publico.
	- `web/Juegos/patito_volando.html` agrega una cuenta regresiva previa de 5 segundos y `web/Juegos/pollitos_cataplum.html`, `web/Juegos/serpiente_pixel.html`, `web/Juegos/memoria_estelar.html` y `web/Juegos/rebote_bloques.html` ajustan su layout para celular.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni wrappers; `Portal publico - Juegos` sigue siendo de lectura/uso para todos los roles y tambien sin autenticacion.

- Actualizacion 2026-04-16 (frontend compartido: ajustes base para celular y menu flotante):
	- `web/menu.js` cierra el panel del menu flotante al seleccionar una opcion para descubrir la pagina destino de inmediato en movil.
	- `web/estilos.css` mejora el responsive compartido de tablas, navegacion administrativa, panel del menu flotante y CTA fijo de WhatsApp en `index.html`.
	- `web/login.html` vuelve a cargar la hoja compartida correcta, recuperando estilos y comportamiento responsive del login administrativo.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad por rol; el ajuste es transversal de frontend.

- Actualizacion 2026-04-16 (portal publico: botones superiores alineados al CTA de ofertas):
	- `web/estilos.css` hace que `Registrarse o iniciar sesión` e `Informacion de contacto` reutilicen el mismo tratamiento visual del boton `Explorar oferta` de las tarjetas del home.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad por rol; el ajuste es visual dentro del portal publico.

- Actualizacion 2026-04-16 (checkout de licencias: Epayco sandbox estable en PostgreSQL):
	- `backend/db/db.go` agrega `EnsurePaymentGatewaySchema(...)` para asegurar `pagos_epayco` y `pagos_wompi` en `pcs_superadministrador` cuando el backend corre sobre PostgreSQL.
	- `backend/main.go` ejecuta ese bootstrap al arrancar y `backend/handlers/payments_handlers.go` evita degradar a `ERROR` una referencia Epayco que sigue pendiente pero aun no aparece en la validacion externa.
	- `backend/handlers/payments_handlers_test.go` cubre el nuevo criterio de polling pendiente para checkout publico de licencias.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad por rol; `Pasarelas de licencias (Wompi/Epayco)` sigue siendo CRUA exclusivo de `super_administrador` y las rutas publicas mantienen uso de solo consumo.

- Actualizacion 2026-04-16 (portal publico: arcade con perfil compartido y cinco juegos):
	- `web/Juegos/arcade_shared.js` centraliza nombre de jugador, top local y control de sonido para todos los juegos publicos del portal sin introducir autenticacion obligatoria.
	- `web/Juegos/menu_juegos.html` pasa a ser un lobby cuadrado con portadas SVG, resumen de records y tarjetas por juego, y ahora enlaza cinco titulos jugables.
	- `web/Juegos/patito_volando.html` y `web/Juegos/pollitos_cataplum.html` se alinean al nuevo perfil compartido; `web/Juegos/serpiente_pixel.html`, `web/Juegos/memoria_estelar.html` y `web/Juegos/rebote_bloques.html` amplian el modulo con nuevas mecanicas.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni wrappers; `Portal publico - Juegos` sigue siendo de lectura/uso para todos los roles y tambien sin autenticacion.

- Actualizacion 2026-04-15 (alertas de reinicio del servidor: activacion configurable):
	- `web/super/configuracion_avanzada.html` agrega el switch `restart_alert_enabled` para activar o desactivar la notificacion automatica de inicio/reinicio del backend sin borrar `restart_alert_to`.
	- `backend/handlers/server_runtime_notifications.go` mantiene la bitacora y el log local aun cuando la alerta esta apagada; solo omite el envio de correo.
	- `backend/handlers/usuarios_empresa.go` y `backend/handlers/super_config_backup_handlers.go` incorporan el nuevo flag en la API y en el respaldo de configuracion critica.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni wrappers; la configuracion sigue siendo exclusiva de `super_administrador`.

- Actualizacion 2026-04-15 (selector de empresas: iconografia por tipo y tarjetas mas profesionales):
	- `web/js/seleccionar_empresa.js` asigna icono, tono visual y texto de apoyo a cada tarjeta usando `tipo_nombre`, mejorando la lectura rapida del negocio antes de abrir su panel o elegir licencia.
	- `web/estilos.css` renueva el bloque visual de `seleccionar_empresa.html` con tarjetas mas coloridas, chips de estado, metadata de acceso y CTA mas claro, sin alterar rutas ni wrappers.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni privilegios; la seleccion y administracion de empresas sigue dentro del alcance de `super_administrador`.

- Actualizacion 2026-04-15 (pagina_principal super: cantidad integrada al guardado):
	- `web/super/pagina_principal.html` elimina el paso manual `Aplicar cantidad`; la cantidad ahora se refleja en el editor al cambiar el campo y se persiste con el mismo guardado de configuracion.
	- `backend/handlers/pagina_principal_handlers_test.go` agrega cobertura para una configuracion ampliada de tarjetas y confirma que el backend conserva la cantidad solicitada al volver a cargar.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad por rol; `Pagina principal (tarjetas index)` sigue siendo CRUA exclusivo de `super_administrador`.

- Actualizacion 2026-04-15 (portal publico: segundo juego `Pollitos al cataplum`):
	- `web/Juegos/menu_juegos.html` agrega una segunda tarjeta jugable y soporta popup por `slug`, ancho y alto por juego.
	- `web/Juegos/pollitos_cataplum.html` añade un juego publico de resortera con control arrastrar/soltar y niveles cortos.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni wrappers; `Portal publico - Juegos` sigue siendo de lectura/uso para todos los roles y tambien sin autenticacion.

- Actualizacion 2026-04-15 (checkout de licencias: fallback canonico para Epayco/Wompi):
	- `backend/handlers/payments_handlers.go` ya no aborta el checkout solo porque la solicitud llega desde `localhost`; resuelve una base publica valida y cae al dominio canonico del sistema cuando no existe otra opcion publica.
	- `backend/handlers/payments_handlers_test.go` cubre el nuevo fallback canonico y la exclusion de `gmail.confirm_base_url` local.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni wrappers; el ajuste afecta solo la operacion del checkout publico de licencias.

- Actualizacion 2026-04-15 (portal publico: menu de juegos y Patito volando):
	- `web/menu.js` expone la entrada publica `Juegos` dentro del menu flotante compartido y `web/Juegos/menu_juegos.html` sirve como catalogo de tarjetas por juego.
	- `backend/utils/utils.go` deja publico el prefijo `/Juegos/` y `backend/handlers/auth_users_carritos_test.go` valida que el middleware no exija sesion para `menu_juegos.html` ni `patito_volando.html`.
	- `web/Juegos/patito_volando.html` no agrega acciones CRUD/A ni superficies administrativas; es una experiencia publica de solo lectura/uso recreativo.
	- Impacto de matriz: se agrega `Portal publico - Juegos` como lectura/uso accesible para todos los roles, con nota operativa de que tambien queda disponible sin autenticacion por formar parte del portal comercial.

- Actualizacion 2026-04-15 (contrato versionado y editor super):
	- `backend/handlers/super_contrato_handlers.go` incorpora `GET/PUT /super/api/contrato` como superficie exclusiva de `super_administrador` para administrar el contrato vigente y publicar nuevas versiones con trazabilidad.
	- `web/super/contrato.html` agrega una pagina visible solo dentro del panel super para editar el contrato, revisar historial y reutilizar una version previa como base de una nueva publicacion.
	- `backend/handlers/auth_admin_handlers.go` y `backend/handlers/accept_handlers.go` hacen que el acceso administrativo dependa de la `contrato_version_aceptada` frente a la version vigente publicada por super.
	- Impacto de matriz: se agrega el control global `Contrato administrativo (super)` como CRUA exclusivo de `super_administrador`; `/api/public/contrato` y `/contrato.html` quedan de lectura publica para sostener el flujo de aceptacion previa al panel.

- Actualizacion 2026-04-15 (monitor centralizado de errores y recovery global):
	- `backend/utils/system_errors.go` y `backend/utils/utils.go` incorporan captura transversal de errores HTTP, panicos recuperados y procesos internos protegidos, con respuesta amigable al usuario final y detalle tecnico reservado al monitor super.
	- `backend/handlers/super_error_handlers.go` publica `GET /super/api/errores` y `web/super/errores.html` entrega la vista filtrable del sistema completo.
	- Impacto de matriz: se agrega `Monitor de errores del sistema (super)` como lectura exclusiva de `super_administrador`; no se expone a administradores de empresa ni a otros roles.

- Actualizacion 2026-04-15 (checkout de licencias: Epayco visible con Public Key y rutas publicas reales):
	- `backend/handlers/payments_handlers.go` ajusta la disponibilidad de Epayco para el checkout de licencias: con `epayco.enabled=1`, basta `epayco.public_key` para publicar la pasarela y construir el checkout actual.
	- `backend/utils/utils.go` deja realmente publicas `/api/public/licencias/payment_methods`, `/wompi/*` y `/epayco/*`, alineando el middleware con la matriz y con la documentacion del portal de pagos.
	- `web/pagar_licencia.html` mejora el mensaje de indisponibilidad para distinguir entre pasarela desactivada y configuracion incompleta.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni visibilidad del panel; `Pasarelas de licencias (Wompi/Epayco)` sigue como CRUA exclusivo de `super_administrador`, mientras las rutas publicas de checkout permanecen de solo consumo para el flujo comercial.

- Actualizacion 2026-04-15 (login admin sin hint visible y Gmail SMTP editable en directo):
	- `web/login.html` elimina el bloque visible del correo recordado sin cambiar el flujo de autenticacion Google ni el alcance publico del login.
	- `web/super/configuracion_avanzada.html` habilita edicion directa de la configuracion Gmail existente sobre la misma ruta protegida `/super/api/config/gmail`.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad por rol; la configuracion global de correo sigue bajo `super_administrador` y el login administrativo permanece publico.

- Actualizacion 2026-04-15 (portal publico: tamanos configurables en home y landing desde pagina_principal):
	- `backend/handlers/pagina_principal_handlers.go` y `backend/handlers/pagina_principal_handlers_test.go` amplian el contrato de `pagina_principal` para publicar ajustes visuales globales de tamano de tarjeta y tamano de texto.
	- `web/super/pagina_principal.html` permite editar esos ajustes y `web/index.html` junto con `web/descripcion_de_los_sistemas.ht` los consumen desde la API publica para escalar el portal sin rutas adicionales.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad por rol; `Pagina principal (tarjetas index)` se mantiene como CRUA exclusivo de `super_administrador` y el portal sigue siendo publico de solo lectura.

- Actualizacion 2026-04-15 (portal publico: CTA superior de WhatsApp y botones tipo mini-tarjeta):
	- `web/index.html` mantiene las mismas rutas publicas del home, pero actualiza la presentacion del header para que los accesos comerciales principales se muestren como mini-tarjetas compactas.
	- `web/estilos.css` reposiciona el CTA flotante `Contactenos` hacia la esquina superior derecha y preserva su comportamiento responsive sin introducir acciones nuevas.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad por rol; `Pagina principal (tarjetas index)` se mantiene como CRUA exclusivo de `super_administrador` y el portal sigue siendo publico de solo lectura.

- Actualizacion 2026-04-15 (portal publico: landing descriptiva configurable desde pagina_principal):
	- `backend/handlers/pagina_principal_handlers.go` amplía la configuracion de tarjetas del portal para incluir el contenido extendido consumido por `/descripcion_de_los_sistemas.ht`.
	- `web/super/pagina_principal.html` agrega campos de edicion para etiqueta, titular ampliado, parrafos y capacidades clave; `web/descripcion_de_los_sistemas.ht` renderiza ese contenido desde la API publica y deja de depender de textos fijos por nombre de tarjeta.
	- `backend/handlers/pagina_principal_handlers_test.go` cubre la normalizacion y exposicion de esos campos ampliados.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad por rol; `Pagina principal (tarjetas index)` mantiene CRUA exclusivo de `super_administrador` y la landing descriptiva sigue siendo publica de solo lectura.

- Actualizacion 2026-04-15 (checkout de licencias: retorno recuperable tras Epayco/Wompi):
	- `backend/handlers/payments_handlers.go` devuelve a `web/pagar_licencia.html` con contexto operativo del cobro y permite lookup Wompi por `reference` para reconsultar el estado real despues del redirect.
	- `web/pagar_licencia.html` solo endurece el flujo publico de licencias: guarda el pago pendiente, reanuda polling al volver y muestra feedback claro sin crear pantallas administrativas ni acciones nuevas.
	- `backend/handlers/payments_handlers_test.go` cubre la recuperacion por referencia y la URL de retorno enriquecida del checkout.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad por rol; `Pasarelas de licencias (Wompi/Epayco)` sigue siendo CRUA exclusivo de `super_administrador` y el checkout continua siendo publico de solo consumo.

- Actualizacion 2026-04-15 (fix Epayco: llave pública correcta y callbacks con dominio público):
	- `backend/handlers/payments_handlers.go` corrige el contrato de Epayco para separar `public_key`, `private_key` y `customer_id`, además de reutilizar una base pública válida en los callbacks de Epayco/Wompi para licencias.
	- `web/super/configuracion_avanzada.html` ajusta únicamente la semántica y persistencia de la configuración global de pasarelas; no crea nuevas acciones empresariales ni altera wrappers de autorización.
	- `backend/handlers/payments_handlers_test.go` cubre el escenario de checkout público con dominio canónico y credenciales coherentes.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni visibilidad por rol; `Pasarelas de licencias (Wompi/Epayco)` permanece como CRUA exclusivo de `super_administrador`.

- Actualizacion 2026-04-15 (host canónico para login Google y carga visible en estaciones):
	- `backend/utils/utils.go` incorpora un middleware de host canónico que redirige `www.powerfulcontrolsystem.com` al dominio raíz antes de autenticación, evitando mezclar cookies y `redirect_uri` entre dos hosts públicos.
	- `backend/main.go` integra ese middleware sin crear rutas nuevas ni ampliar privilegios; el acceso administrativo conserva el mismo modelo de sesión y rol existente.
	- `web/administrar_empresa/estaciones.html` añade un estado visual `Cargando estaciones...` y mensaje de error de carga, sin modificar endpoints ni permisos del módulo estaciones.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad administrativa por rol; solo se estabiliza el acceso y la UX operativa.

- Actualizacion 2026-04-15 (portal publico: contacto visible y pagina de informacion):
	- `web/index.html` incorpora un enlace superior a `/Informacion_de_contacto.html` y un CTA flotante `Contactenos` que abre WhatsApp con el numero publico comercial.
	- El acceso principal del header se renombra a `Registrarse o iniciar sesión` y queda agrupado junto al enlace de contacto, sin alterar rutas protegidas ni permisos.
	- `AuthMiddleware` trata `index.html` y `/Informacion_de_contacto.html` como rutas publicas exactas, por lo que el portal comercial no requiere sesion.
	- `web/Informacion_de_contacto.html` expone descripcion general del sistema y datos de contacto (`powerfulcontrolsystem@hmail.com`, `3043306506`) sin requerir autenticacion.
	- `web/estilos.css` solo agrega soporte visual para el CTA flotante y la nueva landing de contacto; no se agregan endpoints protegidos ni cambios de wrappers.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni visibilidad de paginas administrativas; la nueva experiencia es completamente publica y de solo lectura.

- Actualizacion 2026-04-15 (portal publico: landing descriptiva unica por tarjetas):
	- `web/index.html` sustituye el destino directo de `Explorar oferta` por la landing publica `/descripcion_de_los_sistemas.ht#<seccion>`, conservando el catalogo en una sola pagina y el salto a la descripcion correcta.
	- `AuthMiddleware` incluye `/descripcion_de_los_sistemas.ht` en el whitelist publico para que la navegacion desde las tarjetas no pida login.
	- `web/descripcion_de_los_sistemas.ht` consume `/api/public/pagina_principal` para renderizar todas las secciones del catalogo y reutiliza el `enlace` configurado en super como CTA `Probar Gratis` por tarjeta.
	- `web/super/pagina_principal.html` solo ajusta la semantica del campo `enlace`; no hay nuevos privilegios CRUD/A ni cambios de wrappers.
	- Impacto de matriz: `Pagina principal (tarjetas index)` sigue siendo CRUA exclusivo de `super_administrador`; la nueva landing publica es de lectura y no altera permisos empresariales.

- Actualizacion 2026-04-15 (checkout de licencias: Epayco primero y Wompi gobernado por configuracion avanzada):
	- `backend/handlers/payments_handlers.go` agrega `GET /api/public/licencias/payment_methods` para exponer al checkout publico solo las pasarelas realmente disponibles y en orden operativo (`epayco`, `wompi`).
	- `web/pagar_licencia.html` consume ese endpoint para mostrar solo Epayco y Wompi, con Epayco primero y Wompi debajo; `web/super/configuracion_avanzada.html` ya permite activar o desactivar ambas pasarelas desde super sin alterar privilegios empresariales.
	- `WompiTermsHandler` y `WompiCreateNequiTransactionHandler` bloquean acceso cuando Wompi esta desactivado o no tiene llaves completas.
	- Impacto de matriz: no se agregan roles nuevos ni se amplian permisos CRUD/A; `super_administrador` conserva CRUA sobre `Pasarelas de licencias (Wompi/Epayco)` y la ruta publica nueva es exclusivamente de lectura para el portal de licencias.

- Actualizacion 2026-04-15 (responsive transversal portal/admin):
	- Se ajustan `web/index.html` y `web/estilos.css` para mejorar portabilidad entre movil y escritorio en el portal publico, panel super y panel empresa.
	- No se agregan ni retiran permisos CRUD/A; el cambio no altera rutas protegidas, wrappers ni visibilidad por rol.
	- Impacto de matriz: sin cambios funcionales de autorizacion; mejora exclusiva de presentacion y navegabilidad.

- Actualizacion 2026-04-15 (hardening login Google y recordar cuenta):
	- `backend/handlers/auth_admin_handlers.go` omite `login_hint` invalidos y la capa cliente limpia `rememberedEmail` corruptos en login/menu/paneles administrativos antes de construir `/auth/google/login`.
	- No se agregan rutas nuevas ni se altera el alcance por rol para autenticacion, super o empresa.
	- Impacto de matriz: sin cambios CRUD/A ni wrappers; mejora exclusiva de integridad del acceso administrativo.

- Actualizacion 2026-04-14 (fix login Google en VPS/local y recordar cuenta):
	- Se habilitan como rutas publicas `GET /js/login.js` y `GET /api/public/pagina_principal` dentro de `AuthMiddleware`, evitando respuestas `401` en pagina de login/portal antes de autenticacion.
	- Se robustece el callback OAuth para adaptar `redirect_uri` al host real de la solicitud (`localhost` o dominio VPS) y forzar `https` en dominio publico, sin cambiar privilegios por rol.
	- Se emite una cookie auxiliar visible `browser_session_active` junto a `session_token` para que login/menu detecten sesion activa sin leer la cookie `HttpOnly` real.
	- Se evita fetch de perfil sin sesion en login/menu para reducir errores `401` visibles en cliente antes de autenticar.
	- Impacto de matriz: no hay ampliacion de permisos CRUD/A; se mantiene el mismo modelo de acceso por sesion para rutas protegidas y contexto global para login administrativo.

- Actualizacion 2026-04-14 (checkout de licencias con Epayco):
	- Se completa `POST /epayco/create_transaction`, `GET /epayco/transaction_status` y `POST/GET /epayco/webhook` para flujo de pago/confirmacion de licencias.
	- Se mantiene gestion global de pasarelas en contexto super (`/super/api/config/epayco`), sin cambios de privilegios en wrappers `/api/empresa/*`.
	- Impacto de matriz: solo `super_administrador` conserva permisos CRUD/A en modulo de pasarelas de licencias.

- Actualizacion 2026-04-14 (pagina principal dinamica administrada por super):
	- Se agrega `backend/handlers/pagina_principal_handlers.go` con `GET/PUT /super/api/pagina_principal` y `GET /api/public/pagina_principal`.
	- `web/super/pagina_principal.html` permite configurar tarjetas del home (cantidad, imagen, titulo, descripcion, enlace) y `web/index.html` las renderiza dinamicamente con CTA `Explorar oferta`.
	- Alcance de seguridad: gestion exclusiva para `super_administrador`; sin cambios en wrappers ni privilegios de `/api/empresa/*`.

- Actualizacion 2026-04-14 (venta publica por subdominio empresarial):
	- Se habilita resolucion automatica de `empresa_slug` por `Host`/`X-Forwarded-Host` en `backend/handlers/venta_publica.go` para subdominios tipo `{slug}.powerfulcontrolsystem.com`.
	- La raiz de subdominio empresarial (`/`) redirige internamente a `venta_publica.html` en `backend/main.go` para consumo publico sin query manual.
	- No hay cambios en privilegios CRUD/A por rol: se mantiene endpoint publico `/api/public/venta_publica` y wrappers existentes de seguridad empresarial.

- Actualizacion 2026-04-14 (impresoras operativas por empresa):
	- Se incorpora `backend/handlers/empresa_impresoras.go` y `backend/db/empresa_impresoras.go` para administrar impresoras empresariales y resolver destino de impresion.
	- Se registran rutas `GET/POST/PUT/DELETE /api/empresa/impresoras` (wrapper de seguridad) y `GET /api/empresa/impresoras/resolver` (wrapper de ventas).
	- Sin cambios de privilegios globales: se mantiene politica de lectura comun para rutas con wrapper y mutacion restringida por modulo de seguridad/ventas segun rol.

- Actualizacion 2026-04-14 (super: administracion de base de datos PostgreSQL):
	- Se incorpora endpoint global `/super/api/postgres/performance` para lectura de salud y rendimiento del motor PostgreSQL.
	- Se agrega la vista `web/super/administrar_base_de_datos.html` en el panel de superadministrador para monitoreo y accion operativa.
	- No hay cambios en permisos empresariales ni wrappers de `/api/empresa/*`; el acceso permanece exclusivo para `super_administrador`.

- Actualizacion 2026-04-14 (fase 4 PostgreSQL - estabilizacion worker contable):
	- Se ajusta tecnicamente el procesamiento de eventos/asientos en backend (`backend/db/eventos_contables.go`) para compatibilidad SQL portable en PostgreSQL durante salida controlada.
	- Se restablece runtime VPS con DSN PostgreSQL activos y validacion de salud operativa.
	- No hay cambios en permisos por rol, matrices CRUD/A ni wrappers de autorizacion por modulo.

- Actualizacion 2026-04-14 (migracion PostgreSQL runtime en VPS):
	- Se completa la conmutacion de runtime backend para operar con `DB_DIALECT=postgres` y DSN por base (`DB_SUPERADMIN_DSN`, `DB_EMPRESAS_DSN`).
	- Se agregan capas de compatibilidad SQL para transicion SQLite/PostgreSQL en modulos core sin ampliar privilegios por rol.
	- No hay cambios en la matriz CRUD/A ni en wrappers de autorizacion: se preserva el mismo control por modulo y aislamiento por `empresa_id`.

- Actualizacion 2026-04-13 (estaciones, sensores y facturacion visual por estacion):
	- No hay ampliacion de privilegios ni cambios en matriz CRUD/A; se mantiene el mismo control para `/api/empresa/estacion_prefs`, `/api/empresa/configuracion_avanzada`, `/api/empresa/carritos_compra` y endpoints de sensores empresariales.
	- La reubicacion de colores de carrito a `configuracion_de_estaciones` es un cambio de UX/flujo, no de autorizacion.
	- Se valida aislamiento por `empresa_id` con prueba de handler en `empresa_estacion_prefs`, reforzando separacion de datos entre empresas.

- Actualizacion 2026-04-13 (fix persistencia `empresa_estacion_prefs`):
	- Se corrige normalizacion de estado en capa DB (`estado` vacio => `activo`) sin alterar permisos ni wrappers de autorizacion.
	- El alcance de seguridad permanece igual: controles por `empresa_id` y permisos vigentes en rutas `/api/empresa/estacion_prefs`.

- Actualizacion 2026-04-13 (login empresa, seleccion y estaciones):
	- Se mantiene el mismo esquema de permisos por rol/modulo para endpoints empresariales (`/api/empresa/usuarios/*`, `/api/empresa/estacion_prefs`, `/api/empresa/carritos_compra`).
	- Los cambios son de robustez de flujo y contexto (`empresa_id`) en frontend, sin ampliacion de privilegios ni cambio de matriz CRUD/A.
	- Se preserva aislamiento por `empresa_id` para operacion concurrente de multiples estaciones y carritos por empresa.

- Actualizacion 2026-04-12 (login admin: contrato + reCAPTCHA real):
	- Se consolida la ruta administrativa `login.html -> /auth/google/* -> /accept.html -> /accept/complete` con persistencia de aceptación por cuenta en `administradores.acepta_contrato`.
	- No cambia la matriz CRUD por rol/modulo para rutas empresariales; el ajuste aplica al acceso administrativo global y al endurecimiento de autenticación.
	- Se mantiene aislamiento por `empresa_id` en acceso posterior, ya dentro de wrappers `/api/empresa/*` existentes.

- Actualizacion 2026-04-08 (super: alertas de inicio/reinicio de servidor):
	- Se agrega en configuracion avanzada super la clave `gmail.restart_alert_to` para correo destino de alertas operativas de arranque/reinicio.
	- El cambio no altera permisos de roles empresariales en wrappers `/api/empresa/*`; aplica al ambito global de `super_administrador`.
	- Se mantiene aislamiento multiempresa en operacion: los eventos runtime se registran en contexto super y no exponen datos de empresas fuera de su alcance.

- Actualizacion 2026-04-08 (chat/tareas usuario-admin con adjuntos documentales):
	- Se mantiene el control de acceso del modulo ventas para `/api/empresa/chat_tareas/*` (sin cambios de rol/accion respecto a la matriz vigente).
	- El backend de chat/tareas deriva actor desde sesion autenticada para distinguir `usuario` y `admin`, evitando suplantacion de autor en mensajes/adjuntos.
	- Se habilita colaboracion directa usuario-admin al autoagregar admin propietario de la empresa cuando una conversacion es creada por usuario.
	- Se amplian adjuntos permitidos para colaboracion operativa con documentos de oficina (`doc/docx/xls/xlsx/ppt/pptx/rtf/odt/ods/odp`).

- Actualizacion 2026-04-08 (configuracion monetaria y numerica empresarial):
	- Se agrega en `administrar_empresa/configuracion.html` la seccion de formato monetario/numerico por empresa (`moneda_codigo`, `sistema_numerico`, `usar_decimales`, `cantidad_decimales`).
	- Se mantiene el mismo control de acceso del modulo seguridad para `/api/empresa/configuracion_avanzada` (sin cambios de rol/accion respecto a la matriz actual).

- Actualizacion 2026-04-08 (chat IA empresarial):
	- Se alinea la configuracion avanzada de super para gestionar credencial `deepseek:deepseek-chat`.
	- La pagina `chat_con_inteligencia_artificial` de empresa se actualiza a mensajes/modelo IA generico con ejecucion operativa en DeepSeek, manteniendo control de alcance por `empresa_id`.

- Se implementa middleware en `backend/handlers/empresa_permisos.go` para validar:
	- identidad administrativa activa,
	- alcance de `empresa_id`,
	- permisos por rol/accion (C/R/U/D/A) por modulo.
- Cobertura inicial aplicada en `backend/main.go` sobre rutas criticas:
	- Ventas: `/api/empresa/carritos_compra`, `/api/empresa/carritos_compra/items`.
	- Inventario: `/api/empresa/bodegas`, `/api/empresa/categorias_productos`, `/api/empresa/productos`, `/api/empresa/inventario/*`, `/api/empresa/productos/precios_historial`.
	- Finanzas: `/api/empresa/finanzas/movimientos`, `/api/empresa/finanzas/configuracion`, `/api/empresa/finanzas/periodos`, `/api/empresa/finanzas/asientos_contables`.
- Cobertura ampliada (2026-04-04):
	- Clientes: `/api/empresa/clientes`.
	- Compras/Proveedores: `/api/empresa/proveedores`.
	- Facturacion: `/api/empresa/facturacion_electronica`, `/api/empresa/facturacion_electronica/pais_detectado`, `/api/empresa/facturacion_electronica/dian`.
	- Servicios de catalogo: `/api/empresa/servicios` bajo politica de inventario.
- Cobertura adicional (2026-04-04 - cierre de rutas pendientes):
	- Seguridad/usuarios:
		- `/api/empresa/usuarios`.
		- `/api/empresa/configuracion_avanzada`.
		- `/api/empresa/roles_de_usuario`.
		- `/api/empresa/auditoria/eventos`.
	- Inventario:
		- `/api/empresa/productos/imagen`.
		- `/api/empresa/ubicacion_gps/dispositivos`.
		- `/api/empresa/ubicacion_gps/recorridos`.
	- Colaboracion operativa (politica ventas):
		- `/api/empresa/chat_tareas/conversaciones`.
		- `/api/empresa/chat_tareas/participantes`.
		- `/api/empresa/chat_tareas/mensajes`.
		- `/api/empresa/chat_tareas/mensajes/adjunto`.
		- `/api/empresa/chat_tareas/tareas`.
		- `/api/empresa/chat_tareas/citas`.
- Cobertura adicional (2026-04-05 - contexto de permisos por rol):
	- Seguridad:
		- `/api/empresa/permisos_contexto` con soporte de matriz expandida (`include_matrix=1`) para consulta de permisos efectivos por modulo/accion.
- Cobertura adicional (2026-04-05 - control visual de menu por permisos efectivos):
	- Frontend empresa:
		- `web/js/administrar_empresa.js` consume `/api/empresa/permisos_contexto` para ocultar enlaces no autorizados por rol/modulo.
		- `web/administrar_empresa.html` muestra evidencia visual (`menuPermsEvidence`) con rol y fuente de permisos activa para UAT.
- Cobertura automatizada inicial en `backend/handlers/empresa_permisos_test.go`:
	- denegacion de escritura sin permiso por rol,
	- aprobacion permitida para rol contabilidad en cierre de periodos,
	- bloqueo por fuera de alcance de empresa.
	- denegacion/escritura por rol en modulos `compras` y `facturacion`, y aprobacion de escritura en `clientes` para `cajero` segun matriz.
	- denegacion de escritura en modulo seguridad para `supervisor_sucursal`.
	- aprobacion permitida en modulo seguridad para `admin_empresa`.
	- denegacion para `cajero` al procesar asientos (`action=procesar_asientos`) en modulo finanzas.
	- aprobacion para `contabilidad` al procesar asientos (`action=procesar_asientos`) en modulo finanzas.
	- registro automatico de auditoria para acciones criticas autorizadas (`C/U/D/A`) en middleware de permisos empresariales.
	- cobertura de auditoria automatica por modulo con pruebas en `backend/handlers/auditoria_empresa_test.go` para:
		- `ventas` (`action=cerrar`),
		- `compras` (`action=emitir_orden`),
		- `facturacion` (`action=emitir`).

## Matriz UAT de cierres de caja (roles y transiciones)

Fecha de actualizacion: 2026-04-04

### Casos por rol en endpoint `/api/empresa/finanzas/cierres_caja`

| Caso | Rol | Metodo/accion | Resultado esperado |
|---|---|---|---|
| UAT-CC-R1 | cajero | `PUT action=aprobar` | `403 forbidden` |
| UAT-CC-R2 | supervisor_sucursal | `PUT action=aprobar` | `403 forbidden` |
| UAT-CC-R3 | admin_empresa | `PUT action=aprobar` | `200 ok` |

### Casos de transicion del estado de cierre

| Caso | Estado actual | Accion | Precondicion | Resultado esperado |
|---|---|---|---|---|
| UAT-CC-T1 | abierto | aprobar | ninguna | `409 conflict` (transicion invalida) |
| UAT-CC-T2 | abierto | cerrar | `caja_fisica` valida | `200 ok`, estado `cerrado` |
| UAT-CC-T3 | cerrado | aprobar | ninguna | `200 ok`, estado `aprobado` |
| UAT-CC-T4 | aprobado | reabrir | ninguna | `200 ok`, estado `abierto` |
| UAT-CC-T5 | aprobado | editar/eliminar | ninguna | bloqueo (`409`/error de negocio) |

## Matriz final endpoint/rol (implementacion vigente 2026-04-04)

Leyenda de roles:
- SA: super_administrador
- AE: admin_empresa
- SS: supervisor_sucursal
- CJ: cajero
- IN: inventario
- CO: compras
- CT: contabilidad
- AU: auditor

Regla de lectura comun (R):
- En rutas con wrapper de permisos, lectura queda habilitada para SA, AE, SS, CJ, IN, CO, CT y AU.

| Endpoint | Wrapper/modulo | C/U/A habilitado | D habilitado | Observaciones de accion |
|---|---|---|---|---|
| `/api/empresa/carritos_compra` | `WithEmpresaVentasPermissions` | SA, AE, SS, CJ | SA, AE, SS, CJ | `action=cerrar|reabrir|pagar_estacion|activar_estacion|pagar|suspender|reactivar` exige `A` |
| `/api/empresa/carritos_compra/items` | `WithEmpresaVentasPermissions` | SA, AE, SS, CJ | SA, AE, SS, CJ | mutaciones de items bajo politica de ventas |
| `/api/empresa/chat_tareas/conversaciones` | `WithEmpresaVentasPermissions` | SA, AE, SS, CJ | SA, AE, SS, CJ | colaboracion operativa bajo modulo ventas |
| `/api/empresa/chat_tareas/participantes` | `WithEmpresaVentasPermissions` | SA, AE, SS, CJ | SA, AE, SS, CJ | colaboracion operativa bajo modulo ventas |
| `/api/empresa/chat_tareas/mensajes` | `WithEmpresaVentasPermissions` | SA, AE, SS, CJ | SA, AE, SS, CJ | colaboracion operativa bajo modulo ventas |
| `/api/empresa/chat_tareas/mensajes/adjunto` | `WithEmpresaVentasPermissions` | SA, AE, SS, CJ | SA, AE, SS, CJ | multipart con `empresa_id` obligatorio |
| `/api/empresa/chat_tareas/tareas` | `WithEmpresaVentasPermissions` | SA, AE, SS, CJ | SA, AE, SS, CJ | colaboracion operativa bajo modulo ventas |
| `/api/empresa/chat_tareas/citas` | `WithEmpresaVentasPermissions` | SA, AE, SS, CJ | SA, AE, SS, CJ | agenda de citas compartida por empresa con recordatorios y estado operativo |
| `/api/empresa/bodegas` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | CRUD inventario |
| `/api/empresa/categorias_productos` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | CRUD inventario |
| `/api/empresa/productos` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | CRUD inventario |
| `/api/empresa/productos/imagen` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | upload multipart |
| `/api/empresa/servicios` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | catalogo operativo en politica inventario |
| `/api/empresa/inventario/existencias` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | lectura y mutaciones bajo modulo inventario |
| `/api/empresa/inventario/movimientos` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | lectura y mutaciones bajo modulo inventario |
| `/api/empresa/inventario/transferir` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | transferencias de bodega |
| `/api/empresa/inventario/ajustar` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | ajustes de existencias |
| `/api/empresa/inventario/cambiar_producto` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | remapeo operativo producto/bodega |
| `/api/empresa/productos/precios_historial` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | historial de precios |
| `/api/empresa/ubicacion_gps/dispositivos` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | geolocalizacion en politica inventario |
| `/api/empresa/ubicacion_gps/recorridos` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | geolocalizacion en politica inventario |
| `/api/empresa/clientes` | `WithEmpresaClientesPermissions` | SA, AE, SS, CJ | - | modulo clientes sin `D` por politica actual |
| `/api/empresa/proveedores` | `WithEmpresaComprasPermissions` | SA, AE, SS, CO | - | `action=emitir_orden|recepcionar_compra|contabilizar_compra|aprobar` exige `A` |
| `/api/empresa/facturacion_electronica` | `WithEmpresaFacturacionPermissions` | SA, AE, CJ | - | `action=emitir|nota_credito|emitir_factura|emitir_documento` exige `A` |
| `/api/empresa/facturacion_electronica/pais_detectado` | `WithEmpresaFacturacionPermissions` | SA, AE, CJ | - | consulta/actualizacion bajo politica facturacion |
| `/api/empresa/facturacion_electronica/dian` | `WithEmpresaFacturacionPermissions` | SA, AE, CJ | - | incluye `action=guia_onboarding|validar_credenciales|subir_firma|checklist|validar|generar_cufe_demo|generar_xml_demo|firmar_xml_real|enviar_documento_real|consultar_acuse_real|reconexion_dian|enviar_set_pruebas`; opera por `empresa_id` con `NIT/token/certificado` por empresa y software compartido opcional |
| `/api/empresa/finanzas/movimientos` | `WithEmpresaFinanzasPermissions` | SA, AE, CT | SA, CT | `action=cerrar|reabrir|aprobar|procesar_asientos|procesar` exige `A` |
| `/api/empresa/finanzas/configuracion` | `WithEmpresaFinanzasPermissions` | SA, AE, CT | SA, CT | configuracion financiera |
| `/api/empresa/finanzas/periodos` | `WithEmpresaFinanzasPermissions` | SA, AE, CT | SA, CT | cierre/reapertura de periodos en `A` |
| `/api/empresa/finanzas/asientos_contables` | `WithEmpresaFinanzasPermissions` | SA, AE, CT | SA, CT | `action=procesar_asientos` validado por rol |
| `/api/empresa/finanzas/cierres_caja` | `WithEmpresaFinanzasPermissions` | SA, AE, CT | SA, CT | `action=aprobar` restringido por permiso `A` |
| `/api/empresa/usuarios` | `WithEmpresaSeguridadPermissions` | SA, AE | SA, AE | seguridad/usuarios solo administracion empresa |
| `/api/empresa/configuracion_avanzada` | `WithEmpresaSeguridadPermissions` | SA, AE | SA, AE | seguridad/configuracion sensible |
| `/api/empresa/impresoras` | `WithEmpresaSeguridadPermissions` | SA, AE | SA, AE | CRUD impresoras y acciones `predeterminada|activar|desactivar|funcionalidad|producto` por empresa |
| `/api/empresa/impresoras/resolver` | `WithEmpresaVentasPermissions` | - | - | endpoint operativo de solo lectura para resolver impresora objetivo por `funcionalidad`/`producto_id` |
| `/api/empresa/roles_de_usuario` | `WithEmpresaSeguridadPermissions` | SA, AE | SA, AE | consulta catalogo de roles con control de alcance |
| `/api/empresa/permisos_contexto` | `WithEmpresaSeguridadPermissions` | - | - | endpoint `GET` para visualizar permisos efectivos por modulo/accion; `include_matrix=1` retorna matriz comparativa por rol |
| `/api/empresa/auditoria/eventos` | `WithEmpresaSeguridadPermissions` | SA, AE | SA, AE | consulta y retencion (`action=retener|purgar`) |
| `/api/empresa/backups` | `WithEmpresaSeguridadPermissions` | SA, AE | SA, AE | snapshots/restauracion y depuracion por fecha (`action=restaurar|depurar_fecha` requiere `A`) |

### Endpoints fuera de wrapper (control alterno)

| Endpoint | Control aplicado | Nota |
|---|---|---|
| `/api/empresa/usuarios/login` | validacion de alcance por usuario/empresa en handler | sin middleware de modulo |
| `/api/empresa/usuarios/establecer_password` | validacion de alcance por usuario/empresa en handler | sin middleware de modulo |
| `/api/empresa/facturacion_electronica/paises_disponibles` | catalogo global | sin `empresa_id` obligatorio |
| `/api/empresa/chat_con_inteligencia_artificial/modelos` | `ensureEmpresaAccessByAccount` | validacion por cuenta Google + `empresa_id` |
| `/api/empresa/chat_con_inteligencia_artificial/modelo_preferido` | `ensureEmpresaAccessByAccount` | validacion por cuenta Google + `empresa_id` |
| `/api/empresa/chat_con_inteligencia_artificial/consultar` | `ensureEmpresaAccessByAccount` | validacion por cuenta Google + `empresa_id` |
| `/api/empresa/chat_con_inteligencia_artificial/historial` | `ensureEmpresaAccessByAccount` | validacion por cuenta Google + `empresa_id` |

## Checklist UAT de Punto 3 (permisos y seguridad)

| ID | Verificacion | Estado | Evidencia automatizada |
|---|---|---|---|
| P3-UAT-01 | Denegar escritura inventario a `cajero` | ok | `TestWithEmpresaInventarioPermissionsDeniesCajeroWrite` |
| P3-UAT-02 | Denegar escritura GPS a `cajero` | ok | `TestWithEmpresaInventarioPermissionsDeniesCajeroWriteGPS` |
| P3-UAT-03 | Permitir chat adjunto a `cajero` autenticado | ok | `TestWithEmpresaVentasPermissionsAllowsCajeroChatAdjuntoMultipart` |
| P3-UAT-04 | Rechazar chat adjunto sin autenticacion | ok | `TestWithEmpresaVentasPermissionsRejectsChatAdjuntoWithoutAuth` |
| P3-UAT-05 | Bloquear acceso fuera de alcance de empresa | ok | `TestWithEmpresaVentasPermissionsDeniesOutOfScopeEmpresa` |
| P3-UAT-06 | Denegar `procesar_asientos` a `cajero` | ok | `TestWithEmpresaFinanzasPermissionsDeniesCajeroProcesarAsientos` |
| P3-UAT-07 | Permitir `procesar_asientos` a `contabilidad` | ok | `TestWithEmpresaFinanzasPermissionsAllowsContabilidadProcesarAsientos` |
| P3-UAT-08 | Denegar escritura seguridad a `supervisor_sucursal` | ok | `TestWithEmpresaSeguridadPermissionsDeniesSupervisorWrite` |
| P3-UAT-09 | Permitir accion de seguridad a `admin_empresa` | ok | `TestWithEmpresaSeguridadPermissionsAllowsAdminApprove` |
| P3-UAT-10 | Registrar auditoria en acciones criticas ventas/compras/facturacion | ok | `TestWithEmpresaVentasPermissionsRegistraAuditoriaAccionCritica`, `TestWithEmpresaComprasPermissionsRegistraAuditoriaAccionCritica`, `TestWithEmpresaFacturacionPermissionsRegistraAuditoriaAccionCritica` |
| P3-UAT-11 | Exponer contexto de permisos por rol/modulo en endpoint de seguridad | ok | `TestEmpresaPermisosContextoHandlerRetornaPermisosPorRol`, `TestEmpresaPermisosContextoHandlerIncluyeMatrizRoles` |
| P3-UAT-12 | Ocultar menu por permisos efectivos y mostrar evidencia visual por rol en panel empresa | ok | evidencia visual `menuPermsEvidence` + consumo `GET /api/empresa/permisos_contexto` en `web/js/administrar_empresa.js` |

Ejecucion de validacion actual (2026-04-05):
- `go test ./handlers -run "PermisosContexto|WithEmpresa.*Permissions" -count=1`.
- Resultado: validacion del bloque de permisos y endpoint de contexto (ok).

## Reglas de seguridad obligatorias

1. Todo endpoint debe validar empresa_id y, cuando aplique, sucursal_id antes de operar.
2. Ningun usuario puede actuar fuera de su alcance de empresa/sucursal.
3. Toda accion critica debe dejar auditoria con request_id, empresa_id, usuario, accion y timestamp.
4. Operaciones de cierre/aprobacion deben requerir rol con permiso A.
5. Eliminaciones funcionales deben implementarse como anulacion/inactivacion cuando aplique trazabilidad legal.

## Acciones tecnicas siguientes (cierre operativo punto 3)

1. Incorporar pruebas UAT de regresion para endpoints sin wrapper de modulo (`usuarios/login`, `establecer_password`, chat IA por cuenta Google).
2. Definir politica de aprobacion para rutas de lectura sensible en seguridad (`auditoria/eventos`) segun perfil `auditor` vs `admin_empresa`.
3. Evaluar prueba automatizada E2E del menu dinamico para evitar regresiones de visibilidad por rol.
