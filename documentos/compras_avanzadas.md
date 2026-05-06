# Compras avanzadas

Modulo empresarial integrado al permiso `compras`. No duplica compras/proveedores: amplifica el ciclo existente con requisiciones internas, cotizaciones, aprobaciones y recepcion.

## Alcance

- Requisiciones por empresa con solicitante, area, centro de costo, prioridad, fechas, justificacion e items.
- Cotizaciones por requisicion, proveedor, validez, plazo de entrega, subtotal, impuestos, total y condiciones de pago.
- Aprobacion o rechazo con nivel, aprobador, comentario y monto autorizado.
- Recepcion parcial o total con documento, proveedor, responsable, detalle por item, lote y estado de calidad.
- Dashboard con requisiciones abiertas, pendientes de aprobacion, cotizaciones en evaluacion, recepciones pendientes y valor pendiente.

## Rutas

- `GET /api/empresa/compras_avanzadas?action=dashboard&empresa_id=ID`
- `GET /api/empresa/compras_avanzadas?action=requisiciones&empresa_id=ID`
- `GET /api/empresa/compras_avanzadas?action=detalle&empresa_id=ID&id=REQ_ID`
- `POST /api/empresa/compras_avanzadas?action=requisicion`
- `POST /api/empresa/compras_avanzadas?action=cotizacion`
- `POST /api/empresa/compras_avanzadas?action=aprobar`
- `POST /api/empresa/compras_avanzadas?action=recepcion`
- `POST /api/empresa/compras_avanzadas?action=seed_demo`

## Separacion por empresa

Todas las tablas incluyen `empresa_id` y los handlers usan `WithEmpresaComprasPermissions`, por lo que el acceso se valida con el mismo modulo/licencia `compras`.

## QA

La prueba operativa de Motel Calipso crea una requisicion, registra una cotizacion, aprueba la seleccion, recibe los items y valida que el estado final sea `recibida_total`.
