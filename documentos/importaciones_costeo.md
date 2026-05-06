# Importaciones y costeo de nacionalizacion

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
