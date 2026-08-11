# P109-005 - QA sintético de impresión del candidato actual

Fecha: 2026-08-09
Ambiente: checkout aislado P109, Chrome local y PDF generado
Alcance: motor común de documentos; no se emitieron facturas, no se imprimió en
hardware y no se modificaron datos.

## Resultado

`tools/qa_print_formats.cjs` aprobó 20/20 casos, con cero revisiones y cero
fallos de autoimpresión:

- factura electrónica, recibo, comprobantes de ingreso/egreso, orden, corte de
  caja, parqueadero y turno en Carta/POS;
- factura y recibo extensos: seis páginas cada uno, con cinco páginas de detalle
  y un resumen final;
- QR de factura, imágenes, anchos, filas/columnas y paginación: PASS;
- llamadas `window.print()` comprobadas para factura POS, turno y parqueadero.

## Límite

La validación no sustituye documentos reales con todas las configuraciones de
empresa, tableta, lector de pantalla ni impresión física del piloto. P109-005
permanece **parcial**.
