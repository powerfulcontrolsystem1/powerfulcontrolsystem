# Modulo de carnets empresariales

Fecha: 2026-05-05
Estado: implementado como modulo empresarial profesional

## Alcance

El modulo permite emitir carnets modernos para empleados, usuarios internos, contratistas, visitantes, temporales y directivos. Opera por `empresa_id`, con plantillas visuales propias, QR, foto, niveles de acceso, vencimiento, estados y bitacora.

## Superficies

- Panel: `web/administrar_empresa/carnets.html`.
- API privada: `/api/empresa/carnets`.
- Menu: `Administrar empresa > Personas y activos > Carnets`.
- Permiso de pagina: `linkCarnets`.
- Modulo de licencia/rol: `carnets`.

## Funciones principales

- Dashboard con total de carnets, vigentes, suspendidos, vencidos, revocados y plantillas activas.
- Plantillas por empresa con orientacion vertical/horizontal, colores, logo, foto, QR y codigo visual.
- Emision manual o vinculada a usuarios internos de la empresa.
- Campos profesionales: documento, cargo, area, email, telefono, nivel de acceso, grupo sanguineo, contacto de emergencia, vencimiento y observaciones.
- Vista previa del carnet en tiempo real.
- Exportacion a PNG y SVG desde navegador.
- Impresion del carnet y marcado de impresion/exportacion en bitacora.
- Estados: `vigente`, `pendiente`, `suspendido`, `vencido`, `revocado`.
- Datos demo para arranque rapido por empresa.

## Aislamiento multiempresa

Todas las tablas usan `empresa_id` y la ruta administrativa pasa por `WithEmpresaCarnetsPermissions`. El wrapper central valida que el `empresa_id` de URL, cabecera, formulario/multipart y JSON no se contradiga.

## Base de datos

- `empresa_carnets_plantillas`: diseno visual, colores, orientacion y campos visibles por empresa.
- `empresa_carnets`: carnets emitidos por empresa y persona.
- `empresa_carnets_eventos`: bitacora de emision, actualizacion, cambio de estado e impresion.

## Produccion

Para produccion se recomienda:

- Definir plantilla corporativa por empresa antes de emitir carnets masivos.
- Cargar fotos por URL interna segura o repositorio de archivos empresarial.
- Usar vencimiento obligatorio para contratistas y visitantes.
- Revisar permisos de `carnets` por rol en super administrador.
- Incluir `carnets` en `licencias.modulos_habilitados` cuando el plan debe controlar este modulo explicitamente.
