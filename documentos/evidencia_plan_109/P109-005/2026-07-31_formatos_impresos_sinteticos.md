# P109-005 - Formatos impresos sintéticos

Fecha: 2026-07-31
Entorno: render local reproducible del código del candidato, sin crear documentos empresariales.

## Ejecución

`tools/qa_print_formats.cjs` renderizó HTML, captura y PDF de 18 variantes:
factura, recibo, comprobantes de ingreso/egreso, orden, corte de caja,
parqueadero y turnos, en Carta y POS cuando aplica.

Resultado técnico: **18 PASS, 0 review, 0 fallos de impresión**. Las métricas
no detectaron desborde horizontal ni nodos fuera del ancho de sus formatos.
Los tres flujos con impresión automática instrumentada (factura POS, turno y
parqueadero) invocaron `print` una vez y no cerraron la ventana.

## Revisión visual puntual

Se inspeccionaron las capturas de factura electrónica Carta y POS generadas por
el arnés. Ambas conservan encabezado, datos legales, filas, cantidades,
unitarios, totales, nota CUFE/CUDE y pie. Carta conserva una lectura ordenada
en dos columnas; POS ajusta las filas a la anchura angosta sin sobrepasar el
formato.

## Límite de la evidencia

Es validación sintética de plantillas compartidas. No equivale a impresión
física ni a aceptación DIAN: siguen pendientes documentos reales autorizados,
impresoras del piloto y los formatos de módulos que no usen esta plantilla.
