# Portal de Terceros y Certificados Tributarios

Fecha: 2026-05-06

## Objetivo

El modulo `portal_terceros_certificados` permite administrar terceros y emitir certificados tributarios consultables por enlace publico seguro. Esta pensado para proveedores, clientes, empleados, contratistas, accionistas y contadores externos.

## Alcance funcional

- Maestro de terceros por `empresa_id`.
- Certificados de retencion en la fuente, retencion IVA, retencion ICA, ingresos y retenciones, certificados de proveedor/cliente y otros soportes tributarios.
- Enlace publico seguro por token para visualizacion e impresion.
- Registro de descargas con IP, navegador, canal y fecha.
- Dashboard con certificados emitidos, borradores, anulados, descargas del mes y total de retenciones certificadas.
- Exportacion CSV desde la pantalla administrativa.
- Datos demo para validar el flujo.

## Backend

- API administrativa: `/api/empresa/portal_terceros_certificados`.
- API publica: `/api/public/certificados_tributarios?token=...`.
- Handler: `backend/handlers/portal_terceros_certificados.go`.
- Modelo DB: `backend/db/portal_terceros_certificados.go`.
- Permiso/licencia: `portal_terceros_certificados`.
- Wrapper: `WithEmpresaPortalTercerosPermissions`.

## Tablas

- `empresa_portal_terceros`: tercero, documento, contacto, regimen, estado y token de acceso.
- `empresa_certificados_tributarios`: certificado, periodo, valores, retenciones, estado, firma, token publico y auditoria.
- `empresa_certificados_tributarios_descargas`: bitacora de consulta/impresion publica.

## Frontend

- Administracion: `web/administrar_empresa/portal_terceros_certificados.html`.
- Pagina publica: `web/visualizar_certificado_tributario_publico.html`.
- Enlaces:
  - `linkPortalTercerosCertificados`
  - `linkPortalTercerosCertificadosMenu`

## Seguridad

- El panel administrativo usa permisos por empresa, roles y licencia.
- La consulta externa solo expone certificados con estado `emitido` o `enviado`.
- La URL publica usa `public_token` aleatorio; no permite listar otros certificados.
- Cada descarga se registra en bitacora para trazabilidad.

## Pruebas

- `cd backend; go test ./... -count=1`
- Verificacion estatica de HTML y rutas principales.
