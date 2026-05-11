# Contrato de pruebas reales de pagos y comprobantes

## Cobertura obligatoria

- Wompi sandbox: checkout aprobado, rechazado, pendiente y webhook.
- Wompi produccion: prueba controlada de bajo valor con conciliacion manual.
- Epayco sandbox: checkout aprobado, rechazado, pendiente y webhook.
- Epayco produccion: prueba controlada de bajo valor con conciliacion manual.
- Reembolsos/anulaciones: validar trazabilidad cuando el proveedor lo permita.
- Comprobantes: factura grande, factura pequena, recibo, ticket y soporte de pago.

## Evidencia

Cada corrida debe registrar:

- Fecha, ambiente, empresa usada y licencia/producto probado.
- Referencia de transaccion del proveedor.
- Estado interno guardado en base de datos.
- Capturas visuales de pantalla de checkout y comprobante final.
- Resultado de impresion grande y pequena cuando aplique.

## Reglas de seguridad

- Las credenciales reales se guardan solo en variables de entorno del VPS.
- Los webhooks deben validarse por firma/token del proveedor.
- No se deben commitear respuestas completas con datos sensibles.
