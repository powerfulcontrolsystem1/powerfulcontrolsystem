# Compras avanzadas

Modulo empresarial integrado al permiso `compras`. No duplica compras/proveedores: amplifica el ciclo existente con requisiciones internas, cotizaciones, aprobaciones y recepcion.

## Alcance

- Requisiciones por empresa con solicitante, area, centro de costo, prioridad,
  fechas, justificacion y entre 1 y 500 productos unicos preparados en el
  navegador antes de guardar.
- Cotizaciones por requisicion, proveedor, validez, plazo de entrega, subtotal, impuestos, total y condiciones de pago.
- Aprobacion o rechazo con nivel, aprobador, comentario y monto autorizado.
- Recepcion parcial o total con documento, proveedor, responsable, uno o varios items, lote por producto y estado de calidad.
- Dashboard con requisiciones abiertas, pendientes de aprobacion, cotizaciones en evaluacion, recepciones pendientes y valor pendiente.

## Rutas

- `GET /api/empresa/compras_avanzadas?action=dashboard&empresa_id=ID`
- `GET /api/empresa/compras_avanzadas?action=requisiciones&empresa_id=ID`
- `GET /api/empresa/compras_avanzadas?action=detalle&empresa_id=ID&id=REQ_ID`
- `POST /api/empresa/compras_avanzadas?action=requisicion`
- `POST /api/empresa/compras_avanzadas?action=cotizacion`
- `POST /api/empresa/compras_avanzadas?action=aprobar`
- `POST /api/empresa/compras_avanzadas?action=recepcion`

## Separacion por empresa

Todas las tablas incluyen `empresa_id` y los handlers usan `WithEmpresaComprasPermissions`, por lo que el acceso se valida con el mismo modulo/licencia `compras`.

La recepcion agrupa varios items pendientes en una sola solicitud atomica y en
un solo documento. El backend bloquea la requisicion y cada item, rechaza
mas de 500 items, items repetidos o cantidades por encima del pendiente y solo
confirma stock, costo, lote, Kardex y estados si todo el documento puede
completarse. Las referencias de producto, requisicion, cotizacion, proveedor y
bodega se validan por `empresa_id`; una referencia ajena se responde como 404
seguro y no cambia ningun dato.

Las facturas de compra procesadas por IA admiten PNG, JPEG, WebP, PDF y XML
tanto en la radicacion/extraccion como en el adjunto privado del documento. El
contenido se verifica contra su extension y se descarga como attachment con
`nosniff`. En XML se exige un unico documento bien formado y se rechazan DTD,
texto libre y estructuras mal cerradas.

La superficie productiva no publica acciones ni helpers para generar datos de
demostracion.

## QA

La prueba operativa en la empresa PCS crea una requisicion real de QA, registra
una cotizacion y aprobacion, recibe uno o varios items y valida stock, lotes,
Kardex e indicadores hasta el estado `recibida_total`.
