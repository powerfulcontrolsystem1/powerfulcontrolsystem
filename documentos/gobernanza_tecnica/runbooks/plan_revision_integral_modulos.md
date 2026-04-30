# Plan de revision integral de modulos

Fecha de inicio: 2026-04-29

## Objetivo

Revisar el sistema de principio a fin, modulo por modulo, probando cada funcionalidad antes de pasar a la siguiente. Si un modulo tiene errores, se corrige y se repite la prueba. Si esta incompleto, se documenta el faltante y se completa cuando el alcance este definido por la funcionalidad esperada del proyecto.

## Regla de avance

Un modulo solo pasa a estado revisado cuando cumple estos puntos:

1. La pagina carga sin error de JavaScript.
2. Los endpoints que usa responden con `empresa_id` o contexto correcto.
3. Crear, listar, editar, eliminar o activar/desactivar funciona cuando aplique.
4. Los mensajes de error son claros para el usuario.
5. No mezcla datos entre empresas.
6. La vista se adapta a PC y movil.
7. Se ejecuta una prueba posterior a la correccion.

## Linea base inicial

- Backend: `go test ./...` aprobado.
- Frontend HTML: 119 scripts embebidos verificados con `node --check`.
- Primeros errores detectados y corregidos:
  - `web/administrar_empresa/administrar_usuarios.html`: error de sintaxis en payload por coma faltante.
  - `web/administrar_empresa/ventas.html`: string con `</script>` que rompia el script y bloque HTML duplicado al final.
  - `web/administrar_empresa/productos/administrar_productos_menu.html`: enlace roto a `productos/proveedores.html`; corregido a `productos/administrar_proveedores.html`.
- Navegacion base: enlaces `.html` de `index`, `seleccionar_empresa`, `administrar_empresa` y `super_administrador` aprobados.
- Activos locales: quedan pendientes los instaladores RustDesk esperados en `/descargas/` para soporte remoto. No se reemplazan con archivos vacios; deben publicarse los binarios reales validados para evitar descargas corruptas.

## Orden de trabajo

### Fase 1. Nucleo de empresa

- Panel de empresa.
- Configuracion menu.
- Configuracion general.
- Configuracion avanzada.
- Configuracion de estaciones.
- Estaciones.
- Carrito de compras.
- Usuarios, roles y permisos por empresa.
- Auditoria.

### Fase 2. Operacion comercial

- Productos, categorias, bodegas y precios.
- Clientes.
- Ventas.
- Facturas electronicas.
- Facturacion electronica DIAN.
- Codigos de descuento.
- Comisiones y propinas.
- Compras.
- Finanzas.

### Fase 3. Operacion especializada

- Reservas hotel.
- Tarifas por dia y por minutos.
- Hoja de vida operativa.
- Vehiculos.
- Asistencia y nomina.
- Sensores Raspberry.
- Soporte remoto.
- Nextcloud y OnlyOffice.

### Fase 4. IA, chat y productividad

- Chat IA flotante.
- Robot y secretaria.
- Voz IA streaming.
- Chat, tareas y agenda.
- Reportes IA.
- Exportacion de documentos desde chat.
- Pedidos con IA por estacion.

### Fase 5. Publico y ventas externas

- Red social empresarial.
- Perfil publico y feed comercial.
- Venta publica.
- Pagos de venta publica.
- Elegir licencia.
- Pagar licencia.
- Pasarelas Epayco, Wompi y Nequi.

### Fase 6. Super administrador

- Panel super.
- Licencias y resumen.
- Tipos de empresa.
- Preconfiguracion tipos.
- Administradores.
- Roles y permisos.
- Reportes globales.
- Configuracion avanzada.
- Integracion IA.
- Voz streaming IA.
- Seguridad, errores, trafico, backups y base de datos.

## Evidencia por modulo

Para cada modulo se debe registrar:

- Archivo o ruta revisada.
- Endpoints usados.
- Pruebas ejecutadas.
- Error encontrado.
- Correccion aplicada.
- Resultado posterior.
- Riesgo pendiente, si existe.

## Estado actual

En progreso. Linea base automatica aprobada despues de corregir los dos errores iniciales de JavaScript.

## Evidencia inicial registrada

| Area | Prueba | Resultado | Accion |
| --- | --- | --- | --- |
| Backend Go | `go test ./...` en `backend` | Aprobado | Sin cambios requeridos en esta pasada |
| Frontend HTML | Extraccion de scripts embebidos y `node --check` | 119 scripts aprobados | Corregidos errores de sintaxis detectados |
| Navegacion principal | Verificacion de enlaces `.html` en paginas raiz | Aprobado | Sin enlaces faltantes en menus raiz |
| Productos | Verificacion de submenu | Error encontrado | Corregido enlace de proveedores |
| Soporte remoto | Verificacion de assets locales | Pendiente operativo | Faltan binarios reales RustDesk en `web/descargas/` |

## Fase 1 - avance 2026-04-29

| Modulo | Prueba | Resultado | Accion |
| --- | --- | --- | --- |
| Chat IA flotante | Carga de `administrar_empresa.html` con preferencia `personality_mode=robot` y API simulada | Aprobado: el avatar inicia oculto, `robotShowBtn` queda visible y el panel inicia oculto | `web/js/ai_chat_drawer.js` ahora conserva el robot oculto al cargar y solo lo muestra por accion del usuario |
| Chat IA flotante | Click en `Mostrar robot IA 3D` | Aprobado: aparece el avatar, se muestra el panel inline y se oculta el boton compacto | Se ajusto `showRobotAssistant` y la sincronizacion de visibilidad |
| Estaciones PC | Medicion automatica de alturas en 1366px | Aprobado: 7 tarjetas con altura uniforme de 134px | `web/estilos.css` fija filas y altura de tarjetas de estaciones |
| Estaciones PC miniatura | Medicion automatica con vista miniatura en 1366px | Aprobado: 7 tarjetas con altura uniforme de 86px y boton visible | La vista miniatura queda disponible tambien en PC |
| Estaciones movil | Medicion automatica de alturas en 390px | Aprobado: 7 tarjetas con altura uniforme de 134px | Se mantuvo el layout movil sin crecimiento desigual |
| Estaciones movil miniatura | Medicion automatica con vista miniatura en 390px | Aprobado: 7 tarjetas con altura uniforme de 86px y boton visible | Se normalizo la altura de miniaturas en movil |

## Fase 2 - avance 2026-04-29

| Modulo | Prueba | Resultado | Accion |
| --- | --- | --- | --- |
| Operacion comercial | Carga automatizada PC y movil de 20 paginas: productos, bodegas, precios, categorias, clientes, ventas, facturacion electronica, codigos de descuento, comisiones, propinas, compras, finanzas y reportes financieros | Aprobado con APIs simuladas; no se detectaron errores JavaScript bloqueantes en los modulos principales | Evidencia generada localmente durante la revision |
| Historial de productos | Carga PC y movil de `/administrar_empresa/historial_productos.html?empresa_id=1` | Error real: la pagina importaba `reqBuilder` desde `/js/administrar_empresa.js`, que no exporta modulos ES | Se reemplazo por un constructor local de URL con resolucion de `empresa_id`; prueba posterior aprobada con datos simulados |
| Seleccionar empresa | Tarjeta de empresa sin licencia | Aprobado: queda un solo boton rojo, el lapiz de editar; se elimino el indicador rojo extra de licencia inactiva en la tarjeta | `web/js/seleccionar_empresa.js` usa un espaciador en la celda de licencia cuando no hay licencia activa |

## Revision transversal - avance 2026-04-29

| Area | Prueba | Resultado | Accion |
| --- | --- | --- | --- |
| Todas las paginas HTML | Carga automatizada de 136 paginas en PC y movil, 272 vistas en total, con API simulada | Aprobado para errores de JavaScript despues de corregir historial de productos; sin desbordes horizontales moviles detectados | Se separo el fallo esperado de WebSocket en `pantalla_publica.html` causado por el servidor estatico de prueba |
| Frontend HTML | Extraccion de scripts embebidos y `node --check` | 119 scripts aprobados | Sin cambios adicionales requeridos |
| Backend Go | `go test ./...` en `backend` | Aprobado | Sin cambios requeridos en esta pasada |

## Fases 3 a 6 - avance 2026-04-29

| Fase | Prueba | Resultado | Accion |
| --- | --- | --- | --- |
| Fase 3 - Operacion especializada | Carga automatizada PC y movil de reservas, tarifas, hoja de vida operativa, vehiculos, asistencia, nomina, sensores Raspberry, soporte remoto, Nextcloud y OnlyOffice | Aprobado con APIs simuladas; sin errores JavaScript ni desbordes horizontales moviles | Se mantiene pendiente operativo de publicar binarios reales RustDesk en `web/descargas/` |
| Fase 4 - IA, chat y productividad | Carga automatizada PC y movil de chat IA, configuracion de chat flotante, robot/secretaria, chat/tareas/agenda, reportes IA y pedidos IA por estacion | Aprobado con APIs simuladas; sin errores JavaScript ni desbordes horizontales moviles | Sin correcciones adicionales en esta pasada |
| Fase 5 - Publico y ventas externas | Carga automatizada PC y movil de red social, perfil publico, venta publica, pago de venta publica, elegir licencia y pagar licencia | Aprobado con APIs simuladas; sin errores JavaScript ni desbordes horizontales moviles | Sin correcciones adicionales en esta pasada |
| Fase 6 - Super administrador | Carga automatizada PC y movil de 25 paginas super administrador: licencias, tipos, preconfiguracion, administradores, roles, reportes globales, configuracion avanzada, IA, voz, seguridad, trafico, errores, backups y base de datos | Aprobado con APIs simuladas despues de corregir recurso publico `/config.js`; sin desbordes horizontales moviles | Se creo `web/config.js` con configuracion publica segura por defecto para evitar 404 y no publicar secretos |
| Recursos locales fases 3 a 6 | Verificacion de `href` y `src` locales | Aprobado salvo instaladores RustDesk pendientes | No se crean instaladores vacios; deben publicarse binarios reales validados |

## Capa backend y contratos - avance 2026-04-29

| Area | Prueba | Resultado | Accion |
| --- | --- | --- | --- |
| Inventario de rutas backend | Extraccion de registros `http.HandleFunc` en `backend/main.go` | 200 rutas registradas: 60 super administrador, 36 empresa/configuracion base, 29 venta publica/redes, 19 operacion/finanzas/reportes, 18 IA/chat/documentos, 11 modulos operativos avanzados, 10 licencias/pagos y 17 publico/sistema | Inventario usado como base para validar cobertura funcional por fases |
| Seguridad de rutas por empresa | Verificacion estatica de rutas `/api/empresa/` contra wrappers de permisos | Aprobado: 0 rutas empresariales sin middleware de permisos, salvo endpoints publicos permitidos de autenticacion | La prueba existente `TestEmpresaRoutesUsePermissionWrappers` protege regresiones futuras |
| Contrato frontend-backend | Cruce automatico de endpoints usados por HTML/JS contra rutas Go registradas | Aprobado: 272 endpoints referenciados por frontend, 0 endpoints faltantes detectados | Se confirma que las paginas revisadas no apuntan a rutas inexistentes |
| Regresion permanente frontend-backend | Nueva prueba `TestFrontendEndpointsAreRegisteredInBackend` | Aprobado: el cruce de endpoints queda integrado a `go test ./...` | Se agrego `backend/main_frontend_routes_contract_test.go` para bloquear futuras paginas con endpoints inexistentes |
| Recursos estaticos y navegacion | Revision de `href`, `src`, `action` y `url(...)` reales, excluyendo scripts dinamicos | Aprobado: 480 referencias estaticas revisadas, 0 faltantes no permitidos | Se agrego `backend/main_frontend_static_resources_test.go`; los unicos faltantes permitidos siguen siendo instaladores RustDesk reales pendientes |
| Seguridad Super Administrador | Prueba directa de `AuthMiddleware` sobre rutas publicas y rutas `/super/api/...` sensibles | Aprobado: login/registro/recuperacion pasan sin sesion; empresas, licencias, Epayco y reportes globales exigen sesion | Se agrego `backend/utils/auth_middleware_test.go` para evitar que rutas super queden publicas por accidente |
| Backend Go | `go test ./...` en `backend` | Aprobado | Sin correcciones backend requeridas en esta capa |
| Limite de la prueba | CRUD real destructivo o con datos productivos | No ejecutado sobre datos reales para no alterar informacion de empresas | La siguiente validacion profunda debe usar base semilla o entorno sandbox con datos controlados |

## Regresion final de esta pasada - 2026-04-29

| Prueba | Resultado |
| --- | --- |
| `go test ./...` en `backend`, incluyendo las nuevas pruebas permanentes | Aprobado |
| Sintaxis frontend de scripts embebidos y JS no modulares | 136 revisiones aprobadas, 0 errores |
| Contrato frontend-backend | Aprobado dentro de `go test`: ningun endpoint frontend apunta a ruta Go inexistente |
| Recursos estaticos locales | Aprobado dentro de `go test`: ningun recurso local faltante fuera de los instaladores RustDesk pendientes |
| Seguridad de rutas super | Aprobado dentro de `go test`: rutas sensibles `/super/api/...` no quedan publicas sin sesion |

## Continuacion de fases 1 a 6 - 2026-04-30

| Area | Prueba | Resultado | Accion |
| --- | --- | --- | --- |
| Linea base backend | `go test ./... -count=1` en `backend` despues de los cambios recientes | Aprobado | Sin correcciones backend adicionales |
| Frontend completo | Validacion de sintaxis de scripts embebidos y JS no modulares | Aprobado: 137 scripts revisados, 0 errores | Sin correcciones frontend por sintaxis |
| Fases 1 a 6 | Carga automatizada con Chrome de 137 paginas HTML en PC y movil, 274 vistas en total, con APIs simuladas y contexto `empresa_id=1` | Aprobado: sin errores JavaScript de pagina y sin desbordes horizontales moviles detectados | La simulacion se ajusto para que `/super/api/tipos_empresas` responda como arreglo, igual que el endpoint real |
| Menu flotante y juegos | Carga del enlace `/Juegos/menu_juegos.html`, apertura de Pacman avanzado, inicio de partida y control por teclado/movil | Aprobado | Se mantiene el nuevo acceso a juegos desde el menu flotante |

## Nota operativa vigente

La revision automatizada cubre carga, contratos, recursos, seguridad de rutas y comportamiento visual basico. Las pruebas CRUD reales destructivas o que alteren datos productivos deben ejecutarse en una base semilla o entorno sandbox antes de marcarlas como pruebas de negocio completas.

## Continuacion de fases 1 a 6 - Pacman Arcade completo - 2026-04-30

| Area | Prueba | Resultado | Accion |
| --- | --- | --- | --- |
| Linea base backend | `go test ./... -count=1` en `backend` | Aprobado | Sin correcciones backend |
| Frontend completo | Validacion de sintaxis de HTML y JS despues de reemplazar Pacman por `daleharvey/pacman` | Aprobado: 142 scripts revisados, 0 errores | Sin correcciones por sintaxis |
| Assets Pacman | Verificacion de HTML, CSS, JS, licencia WTFPL, fuente y audio `ogg/mp3` en `web/Juegos/pacman_arcade/` | Aprobado | El juego queda local, sin depender de CDN externo ni de assets propietarios |
| Fases 1 a 6 | Carga automatizada con Chrome de 137 paginas HTML en PC y movil, 274 vistas en total, con APIs simuladas y contexto `empresa_id=1` | Aprobado: sin errores JavaScript de pagina y sin desbordes horizontales moviles detectados | Sin cambios adicionales |
| Menu flotante y juegos | Apertura desde `menu_juegos.html` hacia `pacman_arcade/index.html`; inicio de partida en PC y movil | Aprobado | El enlace estable queda publicado desde el menu de juegos |
| Pacman Arcade completo | Render de canvas, controles debajo de pantalla, fuentes de audio locales y reproduccion simulada de audio al iniciar partida | Aprobado: canvas no blanco, controles debajo, 6 fuentes de audio operativas y reproduccion de intro detectada | Version anterior eliminada; se conserva licencia WTFPL del proyecto fuente |

## Continuacion de fases 1 a 6 - seis juegos arcade HTML5 - 2026-04-30

| Area | Prueba | Resultado | Accion |
| --- | --- | --- | --- |
| Juegos arcade abiertos | Integracion local de Tetris, Space Invaders, Asteroids, Breakout, Snake y Pong desde repositorios HTML5/JavaScript con licencia MIT | Aprobado | Cada juego queda en su carpeta propia bajo `web/Juegos/`, con codigo fuente original en `source/` y licencia conservada |
| Menu de juegos | Verificacion de `/Juegos/menu_juegos.html` en movil con 9 tarjetas y enlaces a los 6 nuevos juegos, Pacman, Patito Volando y N64 | Aprobado: sin overflow horizontal y todos los enlaces visibles | Se agregaron tarjetas de acceso por juego |
| Envoltorio arcade | Carga PC y movil de los 6 wrappers, iframe local, controles tactiles y canvas renderizado | Aprobado: 12 vistas revisadas, canvas no blanco, controles disponibles y sin errores JavaScript de pagina | Se agregaron `arcade_embed.css` y `arcade_embed.js` para controles y layout comun |
| Recursos externos | Revision de fuentes y favicons remotos en Space Invaders y enlaces absolutos internos de Pong | Aprobado | Space Invaders queda sin dependencia externa de Google Fonts/favicon; Pong usa enlaces relativos dentro de su carpeta |
| Regresion backend | `go test ./... -count=1` en `backend` | Aprobado | Las pruebas permanentes de recursos estaticos y contratos siguen pasando |

## Continuacion de fases 1 a 6 - cinco juegos abiertos adicionales - 2026-04-30

| Area | Prueba | Resultado | Accion |
| --- | --- | --- | --- |
| Juegos abiertos nuevos | Integracion local de Buscaminas Arcade, Memoria Arcade, Frogger Calle, Cube Runner 3D y Space Dock 3D | Aprobado | Los cinco juegos quedan publicados en `web/Juegos/` con implementacion original MIT y sin assets copiados de terceros |
| Juegos 3D reales | Render visual de Cube Runner 3D y Space Dock 3D en Chrome | Aprobado: canvas WebGL no blanco, objetos 3D visibles y HUD activo | Se agrego `web/Juegos/webgl_arcade_engine.js` como motor WebGL sencillo sin CDN |
| Movil | Capturas de menu de juegos, Frogger y Cube Runner en 390px | Aprobado despues de ajuste: textos sin corte horizontal y controles tactiles compactos | Se ajustaron `menu_juegos.html`, `arcade_embed.css` y estilos comunes para una columna movil estable |
| Recursos y enlaces | Solicitudes HTTP 200 de wrappers y `source/index.html` de los cinco juegos | Aprobado | Todos los enlaces del menu resuelven a archivos locales |
| Regresion backend | `go test ./... -count=1` en `backend` | Aprobado | Contratos y recursos estaticos siguen protegidos por pruebas permanentes |

## Continuacion de fases - configuracion total por empresa - 2026-04-30

| Area | Prueba | Resultado | Accion |
| --- | --- | --- | --- |
| Exportar/importar configuracion por empresa | Revision de `/api/empresa/backups?action=exportar_configuracion` y pantalla `administrar_empresa/backups.html` | Implementado y reforzado | El catalogo canonico de configuracion ahora incluye permisos finos, IA preferida, finanzas, inventario, nomina, tarifas, integraciones, Nextcloud, pagos publicos, sensores Raspberry, reportes programados, venta publica y empresas compartidas |
| Seguridad de importacion | Prueba unitaria de normalizacion de tablas permitidas | Aprobado | La importacion de configuracion solo acepta tablas incluidas en la lista blanca empresarial y descarta nombres inseguros o tablas internas |
| Regresion focalizada | `go test ./db -run TestEmpresaConfigBackup -count=1` | Aprobado | Se agrego cobertura permanente para evitar que el catalogo de configuracion quede incompleto otra vez |

## Continuacion de fases - control de consumo por empresa - 2026-04-30

| Area | Prueba | Resultado | Accion |
| --- | --- | --- | --- |
| Rate limit por empresa | Prueba unitaria de ventana por minuto, bloqueo 429 y reinicio de ventana | Aprobado | El wrapper comun de `/api/empresa/...` aplica limite configurable por `empresa_id` antes de ejecutar el modulo protegido |
| Consultas DB/IA | Prueba de deteccion de scope para `/api/empresa/db_admin` | Aprobado | Las consultas administrativas de base de datos tienen limite independiente por minuto para proteger acceso usado por IA y administradores |
| Super administrador | Revision de `Configuracion avanzada > Limitaciones por empresa` | Implementado | Se agregaron campos para solicitudes API por minuto y consultas DB por minuto; ambos se guardan en `pcs_superadministrador` |
| Trazabilidad | Bloqueo por limite excedido | Implementado | El servidor responde 429 con `Retry-After`, headers `X-Empresa-RateLimit-*` y registra auditoria no bloqueante del intento |

## Continuacion de fases - documentos dinamicos desde chat IA - 2026-04-30

| Area | Prueba | Resultado | Accion |
| --- | --- | --- | --- |
| Exportacion desde chat IA | Revision de `ai_chat_drawer.js` y endpoint `/api/empresa/chat_documentos/exportar` | Implementado y reforzado | El chat muestra botones PDF, Word/DOCX, Excel/XLSX, TXT y JSON en respuestas exportables, tambien cuando la respuesta es una tabla corta |
| Backend de documentos | `go test ./handlers -run TestDynamicDocument -count=1` | Aprobado | Se valido generacion PDF, DOCX, XLSX, TXT, JSON y HTML, nombres profesionales, redaccion de secretos y preservacion de tablas para Excel |
| Seguridad de contenido | Prueba de contenido HTML generado desde chat | Aprobado | Se eliminan `script`, manejadores `on*` y URLs `javascript:` antes de renderizar el documento |
| Nombres de descarga | Prueba de fallback sin `download_filename` | Aprobado | Se corrigio el caso que podia producir `nil.docx`; ahora usa el titulo o nombre profesional |
| Frontend | `node --check web/js/ai_chat_drawer.js` con runtime empaquetado | Aprobado | La sintaxis del chat flotante queda validada despues del ajuste de exportacion |
| Regresion backend | `go test ./... -count=1` en `backend` | Aprobado | Contratos, seguridad y rutas se mantienen estables |

## Continuacion de fases - Raspberry Pi y sensores - 2026-04-30

| Area | Prueba | Resultado | Accion |
| --- | --- | --- | --- |
| Provisionamiento seguro | Pruebas unitarias de `device_id`, token y payload de instalacion | Aprobado | El panel de empresa puede provisionar un sensor generando `device_id`, token y ejemplos curl/Python para instalar en la Raspberry Pi |
| Seguridad de heartbeat | Revision de endpoint publico `/api/public/sensor_puertas` | Implementado | Si un dispositivo tiene token configurado, el servidor exige `X-Device-Token` y rechaza heartbeats o mensajes enviados solo por `device_id` |
| UI de configuracion | Validacion de script embebido de `configuracion_sensores_raspberry.html` | Aprobado | La pagina muestra estado de seguridad por dispositivo y un bloque de datos de provisionamiento de un solo uso |
| Regresion focalizada | `go test ./db -run 'TestNormalizeEmpresaSensor|TestGenerateEmpresaSensor' -count=1` y `go test ./handlers -run TestBuildEmpresaSensorProvisioningPayload -count=1` | Aprobado | Cobertura permanente para normalizacion, generacion y contrato operativo de sensores |
| Regresion backend | `go test ./... -count=1` en `backend` | Aprobado | El cambio no rompe rutas ni contratos existentes |

## Continuacion de fases - pagos ePayco licencias - 2026-04-30

| Area | Prueba | Resultado | Accion |
| --- | --- | --- | --- |
| Credenciales ePayco | Registro en `pcs_superadministrador.configuraciones` usando `CONFIG_ENC_KEY` del VPS | Aprobado | `PUBLIC_KEY`, `P_CUST_ID_CLIENTE`, `P_KEY`, `PRIVATE_KEY`, `epayco.enabled=1` y `epayco.mode=production` quedaron configurados; las llaves sensibles quedaron cifradas |
| Integracion oficial | Revision contra documentacion ePayco de Checkout personalizado y paginas de respuesta/confirmacion | Implementado | El fallback estandar ya no usa POST legacy a `secure.payco.co/checkout.php`; ahora devuelve `checkout_type=classic_js`, carga `https://checkout.epayco.co/checkout.js` y abre `external: "true"` con `PUBLIC_KEY` |
| Proteccion de secretos | Prueba unitaria del payload clasico | Aprobado | `P_KEY` no se expone al navegador; queda solo en backend para validar `x_signature` de confirmacion/webhook con SHA256 |
| Frontend pagar licencia | Revision estatica de `web/pagar_licencia.html` | Aprobado parcial | La pagina maneja Smart Checkout v2 y fallback `classic_js`; se retiro la precarga automatica de `checkout-v2.js` para no bloquear `checkout.js`. La validacion Node de sintaxis no pudo ejecutarse por bloqueo de permisos del runtime local |
| Disponibilidad publica | `GET https://powerfulcontrolsystem.com/api/public/licencias/payment_methods?pais_codigo=CO` | Aprobado | Produccion reporta ePayco `enabled=true`, `configured=true`, `available=true`; el nuevo codigo queda pendiente de despliegue al VPS |
| Regresion backend | `go test ./handlers -run Test.*Epayco -count=1` y `go test ./... -count=1` en `backend` | Aprobado | Contratos, rutas, recursos y seguridad siguen pasando |
