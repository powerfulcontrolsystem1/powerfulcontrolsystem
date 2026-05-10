# Reporte de roles, permisos y ayuda privada super

Fecha: 2026-05-10
Estado: vigente

## Resumen

Se actualiza el sistema documental y de ayuda para reflejar la separacion profesional de permisos por modulo, pagina y endpoint. La ayuda administrativa completa queda disponible desde un boton interno del panel super administrador y conserva restriccion exclusiva para el rol `super_administrador`.

## Cambios de roles y permisos

- Se agregan modulos finos al catalogo de roles: `crm_unificado`, `reservas_hotel`, `chat_tareas`, `horarios_trabajadores`, `asistencia_empleados`, `vehiculos_registro`, `hoja_vida_operativa`, `ubicacion_gps`, `nomina_sueldos`, `reportes`, `auditoria`, `backups`, `documentos_onlyoffice` y `nextcloud`.
- Las paginas principales y submenus del panel empresarial tienen regla de pagina en backend y regla espejo en `web/js/administrar_empresa.js`.
- Las rutas API empresariales usan wrappers especificos cuando existe modulo propio, evitando depender de grupos genericos como `ventas`, `seguridad`, `finanzas` o `inventario`.
- Las licencias existentes mantienen compatibilidad: una licencia que ya tenia `ventas`, `seguridad`, `finanzas`, `inventario` o `clientes` puede seguir habilitando las funciones separadas que antes colgaban de esos modulos amplios.
- La matriz de roles mantiene el modelo soportado: licencia define techo de modulos; rol define acciones R/C/U/D/A; pagina define visibilidad; API valida empresa, licencia, rol y pagina.

## Ayuda privada de super administrador

- El panel `web/super_administrador.html` incluye el boton `Ayuda super administrador`.
- El boton abre `/ayuda/ayuda.html` dentro del iframe `contentFrame`.
- `AuthMiddleware` conserva `/ayuda/ayuda.html` como ruta autenticada y exclusiva para `super_administrador`.
- Las ayudas publicas especificas, como `/ayuda/login_administradores.html` y `/ayuda/chat_ia.html`, siguen siendo publicas o de alcance reducido segun su finalidad.
- El rol `control_super_administrador` no recibe el boton de ayuda privada en la navegacion limitada.

## Verificacion recomendada

- `node --check web/js/super_administrador.js`
- `node --check web/js/administrar_empresa.js`
- `go test ./...` desde `backend/`
- Revision estatica de IDs `link...` del panel empresarial contra `permissionPagesCatalogOrdered` y `menuPermissionCatalog`.

## Archivos relacionados

- `backend/handlers/empresa_permisos.go`
- `backend/main.go`
- `backend/main_empresa_routes_security_test.go`
- `backend/main_super_help_frontend_test.go`
- `web/js/administrar_empresa.js`
- `web/js/super_administrador.js`
- `web/super_administrador.html`
- `web/ayuda/ayuda.html`
- `documentos/matriz_roles_permisos_pos_multiempresa.md`
