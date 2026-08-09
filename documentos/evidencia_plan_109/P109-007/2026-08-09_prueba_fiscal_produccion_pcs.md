# P109-007 - Prueba fiscal controlada y reverso oficial en PCS

Fecha: 2026-08-09, America/Bogota  
Ambiente: produccion autenticada, empresa Powerful Control System (`empresa_id=12`).

## Autorizacion y alcance

La persona titular autorizo en el mismo ciclo una venta DIAN real y su
anulacion fiscal. Se uso el flujo visible de venta directa y el articulo de
menor valor disponible: una unidad de menta por COP 100. No se usaron SQL,
endpoints directos, secretos ni la emision manual.

## Controles previos

- El Centro DIAN de la empresa mostro ambiente de produccion, rango vigente,
  certificado disponible por referencia segura y estado DIAN aceptado.
- La accion oficial `Validar credenciales` respondio HTTP 200, sin faltantes ni
  incidencias: firma RSA, software por empresa y configuracion del set fueron
  validados.
- `Probar conexion` respondio HTTP 200 y clasifico el transporte SOAP como
  `online`; el `HEAD 400` del WCF fue interpretado correctamente por el modulo
  como disponibilidad de servicio, no como una factura emitida.

## Ejecucion y conciliacion visible

1. Se agrego una menta desde el catalogo por botones y se cerro el carrito con
   pago en efectivo por COP 100.
2. La configuracion avanzo el consecutivo de produccion y el Centro DIAN dejo
   una transicion posterior de envio en estado aceptado.
3. En la vista de Facturas electronicas, el documento fiscal creado aparecio en
   filas/columnas como factura emitida, con numeracion legal nueva y total COP
   100. La interfaz mantuvo el companion interno separado del documento fiscal.
4. Se uso el boton visible `Anular` sobre esa misma factura, se escribio la
   confirmacion requerida y el motivo de prueba controlada. El sistema genero
   una nota credito electronica total; la factura quedo `anulada` y la nota
   credito `emitida`, ambas visibles en la tabla.

## Resultado y limite de certificacion

La venta fiscal y el reverso oficial completaron sin error visible y la
configuracion de la empresa termino en estado aceptado. Sin embargo, el
historial del Centro DIAN no incorporo un TrackId/ZipKey nuevo para esta
operacion, por lo que no fue posible ejecutar una reconsulta independiente que
demuestre `GetStatusZip StatusCode=00` para la factura y para la nota credito.

Estado: **parcial**. La evidencia cierra la ejecucion visible de venta y
anulacion, pero P109-007 no puede aprobarse hasta persistir y reconsultar los
dos acuses DIAN oficiales (incluyendo la nota credito) y conservar su resultado
sin exponer identificadores o secretos.
