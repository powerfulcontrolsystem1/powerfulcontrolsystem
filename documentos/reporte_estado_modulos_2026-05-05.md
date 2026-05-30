# Reporte de estado de modulos - 2026-05-05

Estado: vigente
Alcance: actualizacion documental completa posterior a portal index, carta publica QR, Motel Calipso, domicilios, Taxi System, roles/licencias y ajustes visuales.

## Resumen ejecutivo

El proyecto queda documentado como plataforma POS/ERP multiempresa con modulos plantillas activables por licencia y permisos: hotel, motel, gimnasio, odontologia, turnos, domicilios tipo Rappi, Taxi System tipo Uber, venta publica, carta QR, red social comercial, control electrico, hoja de vida de vehiculos/activos, inventario, finanzas, nomina, facturacion, CRM, auditoria e IA.

La pagina principal `web/index.html` fue actualizada para mostrar descripciones comerciales completas de los modulos y mantener fallback local cuando no cargan datos dinamicos.

## Portal publico e index

- `web/index.html` presenta los modulos principales con descripciones alineadas al alcance real del sistema.
- El fallback incluye tarjetas para hoteles/moteles, gimnasio, odontologia, domicilios, Taxi System, turnos, inventario, finanzas, roles/licencias y venta publica.
- La descripcion publica ya menciona carta QR, red social comercial, control electrico y hoja de vida de vehiculos/activos.

## Venta publica, carta QR y Motel Calipso

- La carta publica se sirve desde `visualizar_productos_y_precios_publico.html`.
- La ruta directa y la ruta por slug `/{empresa_slug}/visualizar_productos_y_precios_publico.html` quedan publicas, sin login.
- Motel Calipso queda con slug `motel-calipso`, paginas publicas activas y productos/servicios de ejemplo.
- La carta QR se administra desde empresa y permite visualizar la URL publica para imprimir o compartir.
- La red social comercial queda con publicaciones activas de Motel Calipso.

## Domicilios profesional

- El modulo queda documentado como flujo funcional para central administrativa, restaurantes, domiciliarios y cliente publico.
- Incluye pedidos, asignacion, estados, tracking GPS, codigo de entrega, evidencias, metricas y permisos/licencias independientes.
- El objetivo operativo es un alcance basico completo, listo para endurecimiento productivo segun proveedor de mapas, pagos y mensajeria.

## Taxi System profesional

- El modulo queda documentado como operacion tipo Uber para flotas, conductores, vehiculos, dispositivos GPS y mapa operativo.
- Incluye asociacion de GPS por tipo/protocolo, tracking, filtros de estado, eventos y control por roles/licencias.
- La interfaz debe mantener compatibilidad visual con modo claro/oscuro mediante los tokens centralizados.

## Roles, licencias y permisos

- Los modulos nuevos o reforzados quedan gobernados por activacion de licencia y permisos de usuario.
- Aplica a `venta_publica`, `domicilios`, `taxi_system`, `gimnasio`, `odontologia`, `turnos_atencion`, `control_electrico`, `alquileres`, `carnets` y modulos base.
- Auditoria 2026-05-05: las claves de modulo del backend, la pantalla de licencias y el catalogo central del menu empresa quedan alineados sin duplicar modulos ni funciones.
- Todos los enlaces visibles de `web/administrar_empresa.html` tienen regla de pagina en `permissionPagesCatalogOrdered` y regla equivalente en `web/js/administrar_empresa.js`.
- Las paginas publicas solo exponen lectura externa; la administracion, configuracion, precios, QR y publicaciones siguen protegidas.

## Carnets empresariales

- Se agrega modulo profesional para emitir carnets de empleados, usuarios internos, contratistas, visitantes y directivos.
- Incluye plantillas por empresa, colores, orientacion vertical/horizontal, foto, logo, QR, campos de identidad, cargo, area, nivel de acceso, vencimiento y contacto de emergencia.
- Permite vista previa, impresion, exportacion PNG/SVG, marcado de impresion y bitacora por carnet.
- API protegida: `/api/empresa/carnets`.
- Pagina: `web/administrar_empresa/carnets.html`.
- Tablas: `empresa_carnets_plantillas`, `empresa_carnets`, `empresa_carnets_eventos`.

## Aislamiento multiempresa rectificado

- Todas las rutas privadas `/api/empresa/...` revisadas en `main.go`, `RegisterEmpresaChatIARoutes` y `RegisterEmpresaModulosFaltantesRoutes` quedan protegidas por wrappers `WithEmpresa*`.
- El wrapper central rechaza peticiones con `empresa_id` contradictorio entre URL, cabecera, formulario/multipart o JSON.
- El `empresaID` validado se fija en el contexto de request y tiene prioridad sobre valores recibidos posteriormente por helpers comunes.
- Esta regla protege todos los modulos empresariales: POS, estaciones, inventario, compras, clientes, finanzas, facturacion, usuarios, hotel/motel, gimnasio, odontologia, domicilios, Taxi System, turnos, control electrico, hoja de vida, reportes, IA, backups, soporte remoto y ERP adicional.
- Las tablas empresariales operativas mantienen `empresa_id` como columna de alcance; tablas globales `super_*`, licencias, pagos globales, tipos de empresa y catalogos publicos se documentan como alcance global o publico.

## Apariencia y experiencia

- Odontologia, chat flotante/robot, botones y paginas nuevas deben respetar la apariencia centralizada.
- Los textos, botones, tarjetas y estados visuales deben adaptarse a modo claro, oscuro y temas disponibles.
- Se mantiene como regla que todo nuevo modulo use variables/tokens de tema compartidos antes que colores aislados.

## Base de datos

- La publicacion de Motel Calipso no requiere tablas nuevas.
- Usa tablas existentes: `empresa_venta_publica_configuracion`, `empresa_venta_publica_paginas`, `empresa_venta_publica_items` y `empresa_publicaciones_red_social`.
- Los cambios de exposicion publica se resolvieron en middleware y documentacion de rutas.

## Verificacion asociada

- `go test ./utils`
- `go test ./handlers -run 'TestLicenciaModulosCSVControlsModuleAccess|TestValidateEmpresaIDConsistency' -count=1`
- `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`
- Auditoria estatica: sin rutas `/api/empresa` duplicadas, sin claves de modulo duplicadas y sin enlaces visibles del menu empresa fuera del catalogo backend/frontend.
- Validacion externa en produccion:
  - `https://powerfulcontrolsystem.com/motel-calipso/venta_publica.html`
  - `https://powerfulcontrolsystem.com/motel-calipso/visualizar_productos_y_precios_publico.html`
  - `https://powerfulcontrolsystem.com/red_social_comercial.html`
  - `https://powerfulcontrolsystem.com/api/public/venta_publica?empresa_slug=motel-calipso`
  - `https://powerfulcontrolsystem.com/api/public/publicaciones?empresa_id=7`

## Documentacion actualizada

- `documentos/README.md`
- `RESUMEN_DEL_PROYECTO.md`
- `documentos/descripcion_del_proyecto`
- `documentos/estructura_bd.md`
- `documentos/diagramas/estructura_del_codigo.md`
- `documentos/descripcion_de_modulos`
- `documentos/descripcion_de_archivos`
- `documentos/matriz_roles_permisos_pos_multiempresa.md`
- `documentos/carta_publica_productos.md`
- `documentos/domicilios_profesional.md`
- `documentos/taxi_system_profesional.md`
- `documentos/CHANGELOG.md`
- `CHANGELOG.md`
- `documentos/historial_de_cambios`
