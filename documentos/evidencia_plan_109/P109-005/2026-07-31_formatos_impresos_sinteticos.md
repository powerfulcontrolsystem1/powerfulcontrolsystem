# P109-005 - Formatos impresos sintéticos

Fecha: 2026-07-31
Entorno: render local reproducible del código del candidato, sin crear documentos empresariales.

## Ejecución

`tools/qa_print_formats.cjs` renderizó HTML, captura y PDF de 20 variantes:
factura, recibo, comprobantes de ingreso/egreso, orden, corte de caja,
parqueadero y turnos, en Carta y POS cuando aplica. La batería incluye una
factura electrónica y un recibo extensos de 96 renglones cada uno.

Resultado técnico: **20 PASS, 0 review, 0 fallos de impresión**. Las métricas
no detectaron desborde horizontal ni nodos fuera del ancho de sus formatos.
Los tres flujos con impresión automática instrumentada (factura POS, turno y
parqueadero) invocaron `print` una vez y no cerraron la ventana.

Los dos documentos extensos produjeron seis páginas PDF: cinco secciones de
detalle con encabezado de columnas repetido y una página final de resumen. El
arnés confirmó los 96 renglones, el número esperado de secciones y la ausencia
de imágenes rotas. La factura usa una imagen QR SVG generada localmente desde
una URL DIAN de habilitación y comprobó que el QR cargó con dimensiones reales;
esta comprobación no afirma aceptación fiscal del documento sintético.

## Revisión visual puntual

Se inspeccionaron las capturas de factura electrónica Carta y POS generadas por
el arnés. Ambas conservan encabezado, datos legales, filas, cantidades,
unitarios, totales, nota CUFE/CUDE, QR y pie. Carta conserva una lectura
ordenada en columnas; POS ajusta las filas a la anchura angosta sin sobrepasar
el formato.

También se renderizaron con Poppler las seis páginas de los documentos extensos.
La inspección detectó primero un desplazamiento lateral en fragmentos finales;
se corrigió el motor compartido con anchos de columna estables, paginación
explícita y página final independiente. La repetición posterior mostró tablas,
descripciones, importes, totales, firmas y QR completos, sin recortes.

## Límite de la evidencia

Es validación sintética de plantillas compartidas. No equivale a impresión
física ni a aceptación DIAN: siguen pendientes documentos reales autorizados,
roles permitido/denegado, tableta y las impresoras del piloto, además de los
formatos de módulos que no usen esta plantilla. Por eso P109-005 continúa
parcial.
