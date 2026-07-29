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

## Validación autenticada de comprobante financiero - 2026-07-28

- SHA desplegado en staging:
  `0a9c4fd4894c2f3f39888fc197a1da5cf37f15e2`.
- El candidato se construyó desde un worktree aislado; frontend, API, worker y
  PostgreSQL quedaron saludables con bootstrap de esquema desactivado en API y
  worker.
- La página servida contiene el contrato corregido: `window.open` ocurre dentro
  del gesto del botón y antes de `await resolvePrinterForFinanzas()`.
- El botón `Imprimir` no mostró mensaje de popup bloqueado, no añadió errores de
  consola y no duplicó el movimiento financiero QA.
- La fila permaneció con código `CXP-14-2E3B0C9FD617`, total y neto `$25`, y la
  cartera relacionada conservó original `$100`, pagado `$25`, saldo `$75`.

El navegador interno no expone la ventana emergente como una pestaña
controlable. Sigue pendiente capturar directamente esa ventana en carta/POS y
probar impresión física; por ello el estado de P108-013 continúa parcial.

## Repetición local visual - 2026-07-29

- La batería `qa_print_formats.cjs` volvió a renderizar 18 de 18 casos, con
  cero pendientes de revisión y cero fallos de llamada a impresión.
- Factura electrónica, recibo de venta, comprobantes de ingreso/egreso, orden,
  corte de caja, parqueadero y turnos se generaron en carta/POS; los tres flujos
  automáticos llamaron exactamente una vez a `window.print()`.
- La inspección visual de la factura carta mostró encabezado, datos legales,
  líneas, totales, CUFE/CUDE y firmas dentro del área imprimible, sin cortes ni
  solapamientos observables.

La evidencia permanece sintética y local: no sustituye vista previa real por
rol, impresión operativa, paginación de documentos reales ni impresora física.
