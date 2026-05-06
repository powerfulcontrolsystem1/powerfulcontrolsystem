# Activos Fijos e Intangibles NIIF / Fiscal

Fecha: 2026-05-06

## Objetivo

El modulo `activos_fijos_niif_fiscal` formaliza la gestion empresarial de propiedad, planta y equipo e intangibles por empresa. Reutiliza el nucleo de datos de la suite contable Colombia avanzada para evitar duplicar activos, depreciaciones o eventos.

## Alcance funcional

- Libro maestro de activos e intangibles por `empresa_id`.
- Campos NIIF: costo, valor residual, vida util, metodo de depreciacion, depreciacion acumulada, deterioro, valor razonable y valor en libros.
- Campos fiscales: base fiscal, vida util fiscal, metodo fiscal, depreciacion fiscal acumulada, valor fiscal y diferencia NIIF/fiscal.
- Informacion administrativa: serial, placa, ubicacion, responsable, centro de costo, proveedor, seguro, poliza y mantenimiento.
- Eventos: traslados, mantenimientos, ajustes, bajas, ventas y retiros.
- Generacion de depreciacion mensual por periodo.
- Dashboard con costo historico, valor en libros, valor fiscal, diferencia NIIF/fiscal, deterioro y alertas.
- Exportacion CSV del libro de activos.

## Backend

- API principal: `/api/empresa/activos_fijos_niif_fiscal`.
- Handler: `backend/handlers/activos_fijos_niif_fiscal.go`.
- Permiso/licencia: `activos_fijos_niif_fiscal`.
- Wrapper: `WithEmpresaActivosFijosNIIFPermissions`.
- Base de datos reutilizada y ampliada:
  - `empresa_contabilidad_activos_fijos`
  - `empresa_contabilidad_activos_depreciacion`
  - `empresa_contabilidad_activos_eventos`

## Acciones API

- `GET action=dashboard`: resumen, libro, depreciaciones, eventos, alertas y agrupaciones.
- `GET action=activos`: libro maestro.
- `GET action=depreciaciones`: depreciaciones por periodo.
- `GET action=eventos`: bitacora por activo o general.
- `POST action=activo`: registra activo o intangible.
- `POST action=depreciacion`: genera depreciacion del periodo.
- `POST action=evento`: registra traslado, mantenimiento, ajuste, baja, venta o retiro.
- `POST action=seed_demo`: crea activos de ejemplo y genera depreciacion del periodo.

## Frontend

- Pantalla: `web/administrar_empresa/activos_fijos_niif_fiscal.html`.
- Enlace principal: `linkActivosFijosNIIF`.
- Enlace dentro del centro financiero: `linkActivosFijosNIIFMenu`.
- La pantalla se adapta a modo claro/oscuro usando variables centralizadas y `color-mix`.

## Integracion

- Licencias: checkbox en `web/super/licencias.html`.
- Roles: permisos por modulo y pagina desde `backend/handlers/empresa_permisos.go`.
- Menu principal y centro financiero enlazados desde `web/administrar_empresa.html` y `web/administrar_empresa/finanzas_menu.html`.
- Portada publica actualizada en `web/index.html`.

## Pruebas

- `cd backend; go test ./... -count=1`
- `git diff --check`
