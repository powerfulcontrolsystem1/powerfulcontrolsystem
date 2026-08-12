# P110-006 — Envío único DIAN de habilitación y clasificación de acuse

Fecha: 2026-08-12  
Ámbito: sesión autorizada de PCS en staging, empresa 12. Producción local
permaneció desactivada.

## Ejecución limitada

- Se eligió **Enviar factura** en lugar del set automático, que estaba
  configurado para varios documentos. La confirmación visible indicó una sola
  factura hacia habilitación.
- El historial registró primero la recepción SOAP con TrackId y después la
  consulta `GetStatusZip`. DIAN informó que el set de prueba se encontraba
  aceptado. El consecutivo de staging avanzó de 1 a 2, por lo que no se repitió
  el envío.
- No se activó producción local, no se emitió factura de venta de producción, no
  se usó SQL para crear efectos de negocio y no se enviaron documentos
  adicionales.

## Hallazgo y corrección preparada

La respuesta oficial positiva llegó con `IsValid=false` y descripción de set
aceptado. El resolvedor la transitó temporalmente por `rechazado` antes de
`habilitacion_observada`, lo que podía inducir una lectura fiscal incorrecta.

Se corrigió `resolveDIANAcuseFromResponse` para que una descripción oficial
explícitamente aceptada prevalezca antes de la rama genérica de rechazo. La
regresión focalizada aprobó, incluyendo pendiente, documento ya procesado,
acuse SOAP y este caso. El proceso Go reportó `ok`; posteriormente Windows
impidió borrar un ejecutable temporal que otro proceso retenía. Es un límite de
limpieza local, no un fallo de aserción.

## Límite

La corrección queda en el candidato de código y todavía debe desplegarse a
staging antes de reconsultar el TrackId. La evidencia DIAN externa confirma la
aceptación del set, pero P110-006 sigue parcial: faltan la factura/nota de
crédito fiscal de producción prevista por el plan y sus artefactos/conciliación
oficiales.
