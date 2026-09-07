# Inventario inicial de impresiones y vistas previas - Plan 106

Estado: parcial, generado por revisión estática el 2026-07-24. No constituye
prueba visual, PDF ni impresión física aprobada.

## Alcance y regla de ejecución

Este inventario es la línea base de P106-013A. Cada entrada debe probarse con
empresa, rol, datos y plantilla declarados; en pantalla, vista previa/PDF y
papel real cuando tenga salida física. La prueba debe revisar filas, columnas,
encabezados, totales, saltos de página y que la acción no altere la transacción.

Los hallazgos `pendiente de descubrimiento dinámico` cubren documentos creados
por configuración, backend, verticales o botones cargados después de abrir una
página. No se pueden marcar como `NA` sin evidencia de que no están habilitados
para la empresa y rol probados.

## Superficie directa identificada

| Familia | Punto de entrada | Salida esperada | Estado P106-013A |
|---|---|---|---|
| Venta/POS | `administrar_empresa/carrito_de_compras.html` | ticket, factura, recibo y formatos de impresora | Pendiente |
| Ventas | `administrar_empresa/ventas.html` | factura/recibo de venta con QR cuando aplique | Pendiente |
| Facturación electrónica | `administrar_empresa/facturas_electronicas.html` | factura, notas y comprobantes electrónicos | Pendiente |
| Caja | `administrar_empresa/corte_de_caja.html` | arqueo, cierre y reporte de sesión | Pendiente |
| Ingresos/Egresos | `web/js/egresos_ingresos.js` | comprobante con soporte y firmas | Pendiente |
| Finanzas | `administrar_empresa/finanzas.html` | movimientos, CxC/CxP, reportes y comprobantes dinámicos | Pendiente de descubrimiento dinámico |
| Compras/CxP IA | `administrar_empresa/finanzas.html`, soportes IA | soporte, CxP y estado de cuenta de proveedor | Pendiente |
| Pedidos/CRM | ventas, cotizaciones, pedidos y venta pública | pedido, cotización, remisión y devolución | Pendiente de descubrimiento dinámico |
| Turnos | `web/js/turnos_atencion.js`, `web/js/turnos_publicos.js`, `administrar_empresa/reportes_turnos.html` | ticket térmico y reporte | Pendiente |
| Parqueadero | `administrar_empresa/parqueadero.html` | ticket de entrada y recibo de salida | Pendiente |
| Carnets | `administrar_empresa/carnets.html` | carnet individual y lote | Pendiente |
| Código de barras | `administrar_empresa/generador_codigos_barras.html` | etiquetas/códigos | Pendiente |
| Carta pública | `administrar_empresa/carta_productos_publica.html` | QR de carta | Pendiente |
| Reportes | `administrar_empresa/reportes_ejecutivos.html`, `super_reportes_globales.js` | reporte tabular/gráfico | Pendiente |
| Inventario | historial de carrito y reportes | Kardex, existencias, movimientos y etiquetas | Pendiente de descubrimiento dinámico |
| Licencias | `administrar_empresa/licencia_sistema.html` | licencia, compra y comprobante | Pendiente |
| Control eléctrico | `administrar_empresa/control_electrico.html` | reporte operativo | Pendiente |
| Certificados | `visualizar_certificado_tributario_publico.html` | certificado tributario | Pendiente |
| Documentos comunes | `web/js/print_documents.js` | documento generado por función compartida | Pendiente: probar cada consumidor |

## Casos de diseño obligatorios por entrada

1. Documento breve y documento con suficientes ítems para más de una página.
2. Nombres de empresa/cliente/proveedor/producto extensos y caracteres UTF-8.
3. Montos altos, decimales si la moneda los permite, descuentos, impuestos,
   retenciones, pagos mixtos y observaciones extensas.
4. A4, Carta, térmico 58 mm y 80 mm cuando la plantilla lo soporte; confirmar
   márgenes, corte, escala, encabezado repetido y sin páginas vacías.
5. Móvil y escritorio; zoom 100%, 125% y 200%; tema claro/oscuro sin afectar
   el resultado impreso.
6. Rol autorizado, rol restringido y empresa A/B. Confirmar que el documento
   no consulta ni muestra datos de otra empresa.
7. Popup bloqueado, impresora ausente, PDF fallido y datos incompletos: mensaje
   visible, operación sin duplicados y estado recuperable.

## Evidencia mínima por fila

`PASS`, `FAIL`, `BLOCKED` o `NA aprobada`; URL/selector, empresa, rol,
plantilla, datos QA, navegador, tamaño de papel, captura de vista previa,
PDF/foto saneada, resultado de consola/red y verificación de importes contra
la operación origen.

## Pendientes de inventario dinámico

- botones generados en tablas, modales, menús, verticales y configuración de
  impresoras por empresa;
- PDFs/Word/Excel de handlers y documentos descargables no invocados por
  `window.print`;
- notas, documentos soporte, nómina, comprobantes contables, pedidos,
  cotizaciones, remisiones, devoluciones y plantillas activadas por licencia;
- impresoras de red, cajón y hardware: se prueban solo en staging o con
  autorización específica, sin disparar efectos físicos por accidente.
