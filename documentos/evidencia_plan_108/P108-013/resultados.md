# P108-013 - Impresiones y vistas previas

Fecha: 2026-07-26  
SHA local evaluado: `74f91956d35e829178050be9127a1fc14fee065c`

## Batería automatizada aprobada

El runner `tools/qa_print_formats.cjs` generó capturas y PDF locales de 18
formatos sintéticos mediante el motor común `web/js/print_documents.js`:

- 18 casos renderizados, 18 aprobados y 0 pendientes de revisión.
- Factura electrónica, recibo de venta, comprobantes de ingreso/egreso, orden,
  corte de caja, parqueadero y turnos en formatos carta/POS.
- Las tres rutas con impresión automática verificaron exactamente una llamada a
  `window.print()` sin cierre prematuro.
- Artefactos locales: `test_runs/qa_print_formats_2026-07-26T05-46-31/`.

## Límites pendientes

La evidencia es sintética y local. Aún falta generar y revisar documentos
operativos reales controlados en staging, vistas previas por rol, impresora
física, paginación multipágina, aislamiento de empresa y todos los formatos que
no usan el motor común.

Estado: **parcial; no certifica P108-013**.
