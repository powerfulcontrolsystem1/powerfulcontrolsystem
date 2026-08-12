# P110-005 — impresión virtual y revisión visual de 20 formatos

Fecha: 2026-08-12
Ámbito: candidato `945d1751`, ejecución local reproducible con Chrome instalado
en el equipo, sin desplegar ni ejecutar `rs`.

## Resultado automático

`tools/qa_print_formats.cjs` generó HTML, PDF y capturas de los 20 casos del
catálogo de impresión. El resultado fue **20/20 OK**, cero casos a revisar y
cero fallos de autoimpresión.

- Factura, recibo, comprobantes de ingreso y egreso, orden de servicio, corte
  de caja, parqueadero y turno aprobaron Carta/POS.
- La factura y el recibo extensos generaron seis páginas cada uno: cinco de
  detalle y una de resumen.
- Factura POS, ticket de turno y recibo de parqueadero llamaron
  `window.print()` exactamente una vez.
- No se agregaron dependencias ni se usaron datos personales o fiscales reales
  en estas plantillas sintéticas.

El reporte reproducible de esta corrida quedó en
`tmp/pdfs/p110_print_current/reporte.md`; los artefactos son temporales y no se
versionan.

## Revisión visual humana

Se inspeccionaron las capturas de factura Carta, factura POS, factura extensa
y recibo extenso. Se confirmó:

- filas y columnas alineadas, sin texto ni importes recortados;
- total y resumen separados del último bloque de detalle;
- encabezados de detalle repetidos después de cada salto de página;
- CUFE/CUDE y QR visibles en la factura, con zona silenciosa suficiente;
- formatos POS legibles dentro del ancho angosto;
- firmas y pie ubicados después del contenido, sin superposición.

## Límite y estado

P110-005 continúa **parcial**. Esta corrida prueba la impresora virtual PDF y
la composición visual sintética, pero no reemplaza impresión física, tableta,
zoom/lector de pantalla, asistencia humana ni recorrido de todos los formatos
con documentos reales de PCS. Tampoco certifica por sí sola el XML, CUFE o QR
de la factura real `1PCS7` y su nota crédito.
