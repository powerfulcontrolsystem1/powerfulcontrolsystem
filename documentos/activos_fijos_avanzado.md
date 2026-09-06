# Activos fijos avanzado

Estado: Vigente. Responsable: Ingeniería del módulo. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- La vista avanzada y el módulo NIIF/fiscal reutilizan las tablas de activos, depreciaciones y eventos; no son dos libros independientes.
- La generación mensual y los eventos requieren verificar duplicados por activo/período, precisión monetaria y pertenencia empresarial antes de aceptar operación real.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

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

## Fuentes y aceptación de la revisión

[contabilidad_colombia_avanzada.go](../backend/handlers/contabilidad_colombia_avanzada.go), [contabilidad_colombia_avanzada.go](../backend/db/contabilidad_colombia_avanzada.go), [contabilidad_colombia_avanzada_test.go](../backend/db/contabilidad_colombia_avanzada_test.go), [contabilidad_colombia_avanzada.html](../web/administrar_empresa/contabilidad_colombia_avanzada.html), [main.go](../backend/main.go), [empresa_permisos.go](../backend/handlers/empresa_permisos.go).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).

La prueba sobre otra empresa citada en la guía es histórica; las nuevas pruebas deben usar el entorno y empresa autorizados por el usuario.
