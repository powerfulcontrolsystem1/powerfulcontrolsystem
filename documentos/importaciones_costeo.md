# Importaciones y costeo de nacionalizacion

Estado: Vigente. Responsable: Ingeniería del módulo. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Se mantienen importación, items, costos y distribución; el costo aterrizado es cálculo interno según TRM y entradas registradas.
- Registrar arancel, IVA o incoterm no acredita nacionalización ni validación aduanera. Verificar sumas distribuidas, redondeo y pertenencia del producto antes de aceptación.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

Fecha: 2026-05-06

## Alcance

Modulo empresarial para compras internacionales: embarques, proveedor, pais de origen, incoterm, TRM, items importados, costos de nacionalizacion y costo aterrizado por producto.

## Superficies

- Backend: `/api/empresa/importaciones_costeo`.
- Pantalla: `web/administrar_empresa/importaciones_costeo.html`.
- Modulo de licencia: `importaciones_costeo`.
- Tablas: `empresa_importaciones_costeo`, `empresa_importaciones_costeo_items`, `empresa_importaciones_costeo_costos`.

## Funciones

- Registro de importacion por empresa con incoterm, moneda, TRM, referencia y estado.
- Items importados con cantidad, peso, volumen y costo en moneda origen.
- Costos de nacionalizacion: flete, seguro, arancel, IVA, agencia de aduanas, bodegaje u otros.
- Distribucion de costos por valor, peso, volumen o cantidad.
- Calculo de costo base COP, costo distribuido y costo unitario final aterrizado.
- Dashboard con importaciones abiertas/cerradas y costo total.

## Acciones API

- `GET action=dashboard`
- `GET action=importaciones`
- `GET action=detalle&id=...`
- `POST action=importacion`
- `POST action=item`
- `POST action=costo`
- `POST action=distribuir`
- `POST action=seed_demo`

## Validacion

- Pruebas unitarias: `go test ./db -run TestNormalizeImportacion -count=1`.
- QA Calipso: crea importacion, items, costos, distribucion y dashboard.

## Fuentes y aceptación de la revisión

[importaciones_costeo.go](../backend/handlers/importaciones_costeo.go), [importaciones_costeo.go](../backend/db/importaciones_costeo.go), [importaciones_costeo_test.go](../backend/db/importaciones_costeo_test.go), [importaciones_costeo.html](../web/administrar_empresa/importaciones_costeo.html), [main.go](../backend/main.go), [empresa_permisos.go](../backend/handlers/empresa_permisos.go).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).
