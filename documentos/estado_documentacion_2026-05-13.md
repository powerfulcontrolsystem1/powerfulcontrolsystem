# Estado documental vigente - 2026-05-13

Este documento consolida el estado actual del proyecto despues de las actualizaciones de seguridad, licencias, operacion conectada, soporte, documentos, backups, portal publico y modulos empresariales. No reemplaza los documentos canonicos; sirve como mapa rapido para desarrollo, QA, soporte y despliegue.

## Fuentes canonicas

- Vision funcional: `documentos/descripcion_del_proyecto`.
- Arquitectura y rutas: `documentos/diagramas/estructura_del_codigo.md`.
- Base de datos: `documentos/estructura_bd.md`.
- Catalogo de modulos: `documentos/descripcion_de_modulos`.
- Roles, permisos y wrappers: `documentos/matriz_roles_permisos_pos_multiempresa.md`.
- Historial detallado: `documentos/historial_de_cambios`.
- Changelog operativo: `documentos/CHANGELOG.md` y `CHANGELOG.md`.
- Ayuda funcional: `web/ayuda/ayuda.html`.

## Reglas vigentes de producto

- El runtime de datos es PostgreSQL.
- Todo dato operativo de empresa debe conservar aislamiento por `empresa_id`.
- La operacion y la facturacion requieren conexion activa con el backend; no se mantiene modo offline para clientes.
- El despliegue objetivo del VPS es portable bajo Docker, con PostgreSQL, backend, frontend, Nginx edge, TLS y servicios opcionales por perfil.
- Nextcloud queda retirado del producto; la cuota antes asociada a almacenamiento se interpreta como limite maximo de base de datos por empresa.
- Los documentos OnlyOffice se crean en una pantalla unica, se editan en una sesion temporal y se descargan al PC/celular del usuario cuando se guardan.

## Acceso y autenticacion

- `login.html` es el acceso administrativo.
- `login_usuario.html` es el acceso global para usuarios operativos de todas las empresas, sin subdominio obligatorio.
- Los usuarios operativos nuevos se registran por invitacion enviada por el administrador; no hay registro abierto.
- El tema visual del login operativo se toma primero desde cookie `pcs_theme`.
- El panel empresarial carga rol y visibilidad efectiva desde permisos por empresa.

## Licencias y cajas simultaneas

- Cada licencia define `max_cajas_simultaneas`.
- Valor por defecto: 2 cajas abiertas simultaneas por empresa.
- Licencias de 4000 documentos: 4 cajas abiertas simultaneas.
- Abrir o reabrir caja valida el cupo contra la licencia activa.
- Cada pago de carrito debe asociarse a una caja abierta para mantener cierres, arqueos y efectivo separados por `cierre_caja_id`.

## Finanzas, ventas y documentos

- El nucleo operativo de ventas, carritos, ingresos, egresos, facturas, reportes y auditoria funciona conectado al servidor.
- Ventas, facturas, comprobantes, reportes y documentos incorporan acciones para compartir por WhatsApp o correo cuando la pantalla aplica.
- Propinas y comisiones quedan enlazadas a usuarios creados mediante ids internos, conservando etiquetas historicas de compatibilidad.
- Los logos de empresa y sistema son configurables por separado para facturas, recibos, reportes y documentos imprimibles.

## Soporte, comunicaciones y mantenimiento

- La mesa oficial de ayuda es el sistema propio de tickets: empresas crean tickets desde el menu flotante y super administrador los atiende en `/super/tickets_ayuda.html`.
- Super administrador puede enviar correos masivos globales a administradores y usuarios del sistema con previsualizacion, confirmacion y trazabilidad.
- Super administrador puede configurar alertas de vencimiento de licencias por correo.
- Super administrador puede publicar avisos de mantenimiento programado en el panel de empresa sin activar el bloqueo real del sistema.

## Portal publico y experiencia visual

- `index.html` muestra 6 tarjetas principales y las demas en un carrusel horizontal con flechas.
- Las tarjetas del carrusel conservan proporcion visual equivalente a las tarjetas superiores.
- `descripcion_de_los_sistemas.html` reutiliza la misma imagen principal configurada para cada tarjeta del index.
- `login.html` y `login_usuario.html` incorporan ilustraciones profesionales locales sin depender de servicios externos.

## Super administrador y operacion VPS

- El panel super conserva gobierno de licencias, roles/permisos, configuracion avanzada, seguridad, PostgreSQL, alertas, archivos, correos y tickets.
- El Explorador de Archivos super es solo lectura y respeta tema claro/oscuro.
- SSH operativo documentado: puerto `49222` para despliegue y seguridad VPS.
- La seguridad VPS se documenta en runbooks y manuales; cualquier cambio de puerto o firewall debe validar primero conexion externa antes de cerrar accesos antiguos.

## Validacion recomendada antes de produccion

- Ejecutar `go test ./...` en `backend/`.
- Validar scripts JS modificados con `node --check` o parseo de scripts inline cuando aplique.
- Probar por rol: super administrador, administrador de empresa, usuario operativo, cajero/supervisor y roles restringidos.
- Probar flujo minimo: login, seleccionar empresa, abrir caja, vender, pagar, imprimir/compartir, cerrar caja, reporte, auditoria y backup local.
- Probar dependencias reales: SMTP, pasarelas, DIAN/proveedor fiscal, impresoras, sensores/Raspberry, OnlyOffice y servicios Docker del VPS.
