# Activos fijos avanzado

Fecha: 2026-05-06

## Alcance

La fase amplia el submodulo de activos fijos dentro de `contabilidad_colombia_avanzada`, sin crear un modulo duplicado. El objetivo es controlar inventario contable de activos, depreciacion, mantenimientos, traslados y bajas por empresa.

## Superficies

- Backend: `/api/empresa/contabilidad_colombia_avanzada`.
- Pantalla: `web/administrar_empresa/contabilidad_colombia_avanzada.html`.
- Tablas base: `empresa_contabilidad_activos_fijos`.
- Tablas nuevas: `empresa_contabilidad_activos_depreciacion` y `empresa_contabilidad_activos_eventos`.
- Aislamiento: todas las operaciones usan `empresa_id`.

## Funciones

- Registro enriquecido: serial, placa interna, centro de costo, proveedor, poliza, estado operativo y mantenimiento programado.
- Depreciacion por periodo: genera registros por activo y actualiza depreciacion acumulada y valor en libros.
- Metodos soportados: linea recta y saldos decrecientes.
- Eventos: mantenimiento, traslado, baja, venta, retiro y ajuste.
- Resumen gerencial: costo total, valor en libros, depreciacion del periodo, activos dados de baja y mantenimientos pendientes.

## Acciones API

- `GET action=activos_resumen&periodo=YYYY-MM`
- `GET action=activos_depreciaciones&periodo=YYYY-MM`
- `GET action=activos_eventos`
- `POST action=generar_depreciacion_activos`
- `POST action=activo_evento`

## Validacion

- Pruebas unitarias: `go test ./db -run 'TestCalcularEmpresaActivo|TestNormalizeActivoEvento' -count=1`.
- QA Calipso: registra activo, genera depreciacion del periodo, registra mantenimiento y valida resumen avanzado.
