# P110-005 — QA de impresión y responsive

Fecha: 2026-08-11  
Ámbito: candidato local cuyo código funcional coincide con el promovido a
staging; revisión visual autenticada sobre PCS/staging.

## Impresión reproducible

Se repitió el 2026-08-12 `tools/qa_print_formats.cjs` usando Chrome instalado,
sin instalar dependencias nuevas. Resultado: **20/20 OK**, cero casos a
revisar y cero fallos de autoimpresión. La corrida dejó HTML, PDF y capturas
reproducibles bajo `test_runs/qa_print_formats_2026-08-12T05-23-04/` (artefacto
local no versionado).

- Factura y recibo extensos Carta: seis páginas cada uno (cinco de detalle y
  una de resumen).
- Carta/POS: factura, recibo, comprobantes de ingreso/egreso, orden, corte,
  parqueadero y turno.
- QR: presente y legible en los formatos de factura previstos.
- Autoimpresión: una invocación de `window.print()` para factura POS, ticket de
  turno y recibo de parqueadero.

La captura compuesta de la factura extensa fue revisada visualmente: conserva
filas/columnas, encabezados de detalle repetidos, importes alineados, total
separado y QR final sin recortes. La auditoría automática confirmó además seis
páginas PDF, cinco páginas de detalle, una de resumen, 96 filas, cero imágenes
rotas, cero errores de consola y ningún desbordamiento horizontal.

## Revisión visual móvil autenticada

En PCS/staging, `Finanzas empresariales` se revisó a 390 × 844 px. Los campos
principales conservan etiqueta accesible, controles visibles y una jerarquía
de una columna sin desbordamiento en la sección recorrida. La consola de esa
vista no produjo errores ni advertencias.

## Límite

P110-005 queda **parcial**, no certificada: faltan recorrido de todos los
formatos con documentos PCS reales, tableta, zoom/lector de pantalla y firma de
impresión física. Esta evidencia no declara aceptación DIAN ni reemplaza esos
pasos.
