# Activos Fijos e Intangibles NIIF / Fiscal

Estado: Vigente. Responsable: Ingeniería del módulo. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- El handler contempla GET de activos/libro, depreciaciones y eventos; las escrituras admiten POST y PUT.
- Los valores NIIF/fiscales son registros y cálculos del módulo; la guía no certifica la corrección tributaria de una empresa ni sus políticas de depreciación.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

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
- `POST/PUT action=activo`: registra activo o intangible.
- `POST/PUT action=depreciacion`: genera depreciacion del periodo.
- `POST/PUT action=evento`: registra traslado, mantenimiento, ajuste, baja, venta o retiro.
- `POST/PUT action=seed_demo`: crea activos de ejemplo y genera depreciacion del periodo.

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

## Fuentes y aceptación de la revisión

[activos_fijos_niif_fiscal.go](../backend/handlers/activos_fijos_niif_fiscal.go), [activos_fijos_niif_fiscal.html](../web/administrar_empresa/activos_fijos_niif_fiscal.html), [main.go](../backend/main.go), [empresa_permisos.go](../backend/handlers/empresa_permisos.go).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).
