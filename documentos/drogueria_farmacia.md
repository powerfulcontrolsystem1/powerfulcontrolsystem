# Drogueria y farmacia

Fecha: 2026-05-11

## Objetivo

Plantilla empresarial para operar una drogueria o farmacia con trazabilidad sanitaria por empresa: medicamentos, lotes, INVIMA, vencimientos, formulas, controlados, dispensacion, devoluciones y farmacovigilancia.

## Alcance funcional

- Expedientes sanitarios para medicamentos, lotes, formulas medicas, controlados, vencimientos, devoluciones e incidentes de farmacovigilancia.
- Seguimiento por responsable, prioridad, categoria, fechas, evidencias, aprobaciones, tareas, SLA y riesgo operativo.
- Datos guia para lote, registro INVIMA, vencimiento, formula, medicamento controlado, cadena de frio y dispensacion.

## Seguridad y permisos

- Clave de modulo/licencia: `drogueria_farmacia`.
- Pagina empresarial: `web/administrar_empresa/drogueria_farmacia.html`.
- Endpoint protegido: `/api/empresa/drogueria_farmacia`.
- Wrapper: `WithEmpresaDrogueriaFarmaciaPermissions`.
- Todas las tablas compartidas usan `empresa_id` y `modulo='drogueria_farmacia'`; no hay rutas publicas ni mezcla de informacion entre empresas.

## Tablas

No crea tablas propias de producto, inventario, venta ni pago. Usa las tablas compartidas:

- `empresa_modulos_colombia_registros`
- `empresa_modulos_colombia_eventos`
- `empresa_modulos_colombia_evidencias`
- `empresa_modulos_colombia_aprobaciones`
- `empresa_modulos_colombia_tareas`

## Integracion con el nucleo

- Productos, medicamentos, existencias, compras, ventas, pagos, clientes y facturacion pertenecen a los modulos centrales.
- `drogueria_farmacia` guarda solo el expediente sanitario y operativo que no existe en el nucleo comercial.
- Las licencias y preconfiguraciones de drogueria/farmacia habilitan `inventario`, `compras`, `ventas`, `clientes`, `facturacion`, `logistica_wms`, `soportes_compras_ia` y modulos de cumplimiento relacionados.
- La vertical queda visible como `plantilla_integrada_nucleo` porque no duplica clientes, productos, inventario, ventas ni pagos.
