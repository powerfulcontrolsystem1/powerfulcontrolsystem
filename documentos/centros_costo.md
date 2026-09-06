# Centros de costo y rentabilidad

Estado: Vigente. Responsable: Ingeniería del módulo. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- El handler expone dashboard, centros, reglas, presupuestos y movimientos; las escrituras de maestro/reglas/presupuesto admiten POST/PUT.
- Un maestro de reglas de imputación no acredita que todas las operaciones origen las apliquen. La rentabilidad depende de los movimientos y centro_costo realmente cargados.
- Los roles efectivos se consultan en empresa_permisos.go y sus overrides; los ejemplos de la matriz no garantizan acceso de una cuenta concreta.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

Fecha: 2026-05-06

## Objetivo

El modulo `centros_costo` formaliza el control de rentabilidad por empresa, sucursal, area, unidad de negocio o proyecto. No duplica Finanzas ni Contabilidad: actua como capa gerencial que consolida movimientos ya registrados en contabilidad Colombia, tesoreria, compras avanzadas, captura OCR/IA de compras y AIU construccion.

## Alcance funcional

- Maestro de centros de costo con codigo, tipo, responsable, sucursal, area, unidad de negocio, estado y meta de margen.
- Reglas de imputacion por modulo, categoria, cuenta, tercero, porcentaje y prioridad.
- Presupuesto por centro, periodo y escenario base, con ingresos, egresos y meta de margen.
- Dashboard comparativo con ingresos, egresos, margen, margen porcentual, ejecucion presupuestal y alertas.
- Movimientos integrados por `empresa_id` desde modulos existentes con campo `centro_costo`.
- Centros inferidos desde movimientos cuando aun no exista el maestro, para evitar perder informacion historica.
- Exportacion CSV desde la pantalla administrativa.
- Datos demo para Motel Calipso, operaciones, ventas, administracion y obras AIU.

## Backend

- API privada: `/api/empresa/centros_costo`
- Wrapper: `WithEmpresaCentrosCostoPermissions`
- Modulo/licencia: `centros_costo`
- Paginas de permiso: `linkCentrosCosto` y `linkCentrosCostoMenu`

Acciones:

- `GET dashboard`: resumen ejecutivo del periodo.
- `GET centros`: maestro de centros.
- `GET reglas`: reglas de imputacion.
- `GET presupuestos`: presupuesto por periodo/escenario.
- `GET movimientos`: movimientos integrados.
- `POST/PUT centro`: crear o actualizar centro.
- `POST/PUT regla`: crear o actualizar regla.
- `POST/PUT presupuesto`: crear o actualizar presupuesto.
- `POST seed_demo`: precarga ejemplo profesional.

## Base de datos

- `empresa_centros_costo`: maestro por empresa.
- `empresa_centros_costo_reglas`: reglas de imputacion.
- `empresa_centros_costo_presupuestos`: presupuesto por centro/periodo.

Todas las tablas incluyen `empresa_id` y se consultan con el contexto validado por el middleware empresarial.

## Frontend

Pantalla: `web/administrar_empresa/centros_costo.html`

Ubicacion:

- Administrar empresa > Finanzas y cumplimiento > Centros de costo.
- Centro financiero y contable > Centros de costo.

La interfaz usa variables visuales centralizadas para adaptarse a modo claro/oscuro y contiene dashboard, maestro, presupuesto, reglas, movimientos y exportacion.

## Permisos

La matriz base lo trata como modulo financiero/contable:

- Lectura: roles operativos con lectura permitida.
- Crear/actualizar/aprobar: `admin_empresa` y `contabilidad`.
- Eliminar: politica financiera restringida a `contabilidad` si se agrega accion destructiva futura.
- `super_administrador` y `administrador_total` conservan acceso total.

## Verificacion

- `go test ./... -count=1` ejecutado en `backend/`.
- Pruebas unitarias nuevas para normalizacion y agregacion de dashboard en `backend/db/centros_costo_test.go`.

## Fuentes y aceptación de la revisión

[centros_costo.go](../backend/handlers/centros_costo.go), [centros_costo.go](../backend/db/centros_costo.go), [centros_costo_test.go](../backend/db/centros_costo_test.go), [centros_costo.html](../web/administrar_empresa/centros_costo.html), [main.go](../backend/main.go), [empresa_permisos.go](../backend/handlers/empresa_permisos.go).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).
