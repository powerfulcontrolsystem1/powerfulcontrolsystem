# Contrato de pruebas reales de pagos y comprobantes

Estado: Vigente. Responsable: Ingeniería backend y QA. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Matriz de aceptación a ejecutar sobre un candidato identificado. Un cobro/reembolso de producción requiere autorización específica de importe, cuenta y conciliación; no se ejecuta por leer este documento.
- Credenciales se resuelven por configuración privada autorizada; no se limita su custodia a una única forma de variable de entorno.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

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

- Las credenciales reales se custodian en la configuración privada autorizada y nunca en el repositorio ni en la evidencia compartida.
- Los webhooks deben validarse por firma/token del proveedor.
- No se deben commitear respuestas completas con datos sensibles.

## Fuentes y aceptación de la revisión

[contrato_checkout_licencias_publico.md](contrato_checkout_licencias_publico.md), [main.go](../../../backend/main.go).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](../../requisitos/especificacion_y_trazabilidad.md)).
