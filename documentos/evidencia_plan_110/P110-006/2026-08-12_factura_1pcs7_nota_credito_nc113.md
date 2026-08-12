# P110-006 - Factura real `1PCS7` y nota crédito `NC12000000113`

Fecha: 2026-08-12 (America/Bogota)
Empresa: Powerful Control System (`empresa_id=12`)
Ambiente fiscal: producción DIAN configurada por empresa.

## Emisión autorizada

- Se cerró por la interfaz oficial una venta real del producto `menta`, una
  unidad, pago en efectivo exacto y total `COP 100`.
- El cierre produjo una factura electrónica independiente del comprobante
  comercial, con número legal `1PCS7`.
- La consola DIAN de la empresa registró a las `15:47:54`:
  `factura_electronica`, estado `aceptado`, mensaje
  `Procesado Correctamente.` y `Acuse sincrónico (sin TrackId)`.
- La bandeja fiscal mostró la factura por `COP 100`, estado documental
  `emitida`, y el botón **Visualizar** informó que la vista se abrió
  correctamente.

## Anulación fiscal

- Se abrió el diálogo seguro de anulación, se confirmó con el código requerido
  y un motivo de QA mayor de diez caracteres.
- El sistema generó una nota crédito electrónica total por `COP 100`, número
  legal `NC12000000113`.
- La consola DIAN registró a las `15:50:52`:
  `nota_credito`, estado `aceptado`, mensaje `Procesado Correctamente.` y
  `Acuse sincrónico (sin TrackId)`.
- La factura `1PCS7` quedó `anulada`; la nota crédito quedó `emitida` y activa.
- Se ejecutó una sola emisión y una sola anulación; no se usaron reenvíos ni se
  fabricaron TrackId.

## Separación por empresa

Configuración, prefijo, resolución, rango, firma, historial DIAN, factura y nota
crédito se consultaron dentro del contexto autenticado `empresa_id=12`. No se
modificó ni consultó documentación fiscal de otra empresa durante las acciones.

## Estado

P110-006 continúa **parcial**. Queda evidencia oficial positiva de factura y
nota crédito reales, pero faltan inspección visual completa de XML/PDF, CUFE y
QR, impresión física, doble clic/reintento idempotente, empresa B y cierre de
las demás integraciones incluidas.

No se ejecutó `rs` en este bloque.
