# QA visual de impresion

- Fecha: 2026-08-26T13:35:54.914Z
- Motor comun: `web/js/print_documents.js`
- Casos renderizados: 20
- Casos OK: 20
- Casos a revisar: 0

## Validacion de llamada a impresion
- PCSPrint autoPrint factura POS: `window.print()` llamado 1 vez/veces.
- Ticket turnos autoPrint: `window.print()` llamado 1 vez/veces.
- Recibo parqueadero autoPrint: `window.print()` llamado 1 vez/veces.

## Evidencia
- OK | Factura electronica de venta - carta | factura | carta | paginas PDF: 1 | captura: `test_runs/qa_print_formats_2026-08-26T13-35-50/screenshots/factura_electronica_de_venta_-_carta.png` | PDF: `test_runs/qa_print_formats_2026-08-26T13-35-50/pdf/factura_electronica_de_venta_-_carta.pdf`
- OK | Factura electronica de venta - pos | factura | pos | paginas PDF: 1 | captura: `test_runs/qa_print_formats_2026-08-26T13-35-50/screenshots/factura_electronica_de_venta_-_pos.png` | PDF: `test_runs/qa_print_formats_2026-08-26T13-35-50/pdf/factura_electronica_de_venta_-_pos.pdf`
- OK | Recibo de venta - carta | recibo | carta | paginas PDF: 1 | captura: `test_runs/qa_print_formats_2026-08-26T13-35-50/screenshots/recibo_de_venta_-_carta.png` | PDF: `test_runs/qa_print_formats_2026-08-26T13-35-50/pdf/recibo_de_venta_-_carta.pdf`
- OK | Recibo de venta - pos | recibo | pos | paginas PDF: 1 | captura: `test_runs/qa_print_formats_2026-08-26T13-35-50/screenshots/recibo_de_venta_-_pos.png` | PDF: `test_runs/qa_print_formats_2026-08-26T13-35-50/pdf/recibo_de_venta_-_pos.pdf`
- OK | Comprobante de ingreso - carta | comprobante_ingreso | carta | paginas PDF: 1 | captura: `test_runs/qa_print_formats_2026-08-26T13-35-50/screenshots/comprobante_de_ingreso_-_carta.png` | PDF: `test_runs/qa_print_formats_2026-08-26T13-35-50/pdf/comprobante_de_ingreso_-_carta.pdf`
- OK | Comprobante de ingreso - pos | comprobante_ingreso | pos | paginas PDF: 1 | captura: `test_runs/qa_print_formats_2026-08-26T13-35-50/screenshots/comprobante_de_ingreso_-_pos.png` | PDF: `test_runs/qa_print_formats_2026-08-26T13-35-50/pdf/comprobante_de_ingreso_-_pos.pdf`
- OK | Comprobante de egreso - carta | comprobante_egreso | carta | paginas PDF: 1 | captura: `test_runs/qa_print_formats_2026-08-26T13-35-50/screenshots/comprobante_de_egreso_-_carta.png` | PDF: `test_runs/qa_print_formats_2026-08-26T13-35-50/pdf/comprobante_de_egreso_-_carta.pdf`
- OK | Comprobante de egreso - pos | comprobante_egreso | pos | paginas PDF: 1 | captura: `test_runs/qa_print_formats_2026-08-26T13-35-50/screenshots/comprobante_de_egreso_-_pos.png` | PDF: `test_runs/qa_print_formats_2026-08-26T13-35-50/pdf/comprobante_de_egreso_-_pos.pdf`
- OK | Orden de servicio - carta | orden | carta | paginas PDF: 1 | captura: `test_runs/qa_print_formats_2026-08-26T13-35-50/screenshots/orden_de_servicio_-_carta.png` | PDF: `test_runs/qa_print_formats_2026-08-26T13-35-50/pdf/orden_de_servicio_-_carta.pdf`
- OK | Orden de servicio - pos | orden | pos | paginas PDF: 1 | captura: `test_runs/qa_print_formats_2026-08-26T13-35-50/screenshots/orden_de_servicio_-_pos.png` | PDF: `test_runs/qa_print_formats_2026-08-26T13-35-50/pdf/orden_de_servicio_-_pos.pdf`
- OK | Corte de caja - carta | corte_caja | carta | paginas PDF: 1 | captura: `test_runs/qa_print_formats_2026-08-26T13-35-50/screenshots/corte_de_caja_-_carta.png` | PDF: `test_runs/qa_print_formats_2026-08-26T13-35-50/pdf/corte_de_caja_-_carta.pdf`
- OK | Corte de caja - pos | corte_caja | pos | paginas PDF: 1 | captura: `test_runs/qa_print_formats_2026-08-26T13-35-50/screenshots/corte_de_caja_-_pos.png` | PDF: `test_runs/qa_print_formats_2026-08-26T13-35-50/pdf/corte_de_caja_-_pos.pdf`
- OK | Recibo parqueadero - carta | parqueadero | carta | paginas PDF: 1 | captura: `test_runs/qa_print_formats_2026-08-26T13-35-50/screenshots/recibo_parqueadero_-_carta.png` | PDF: `test_runs/qa_print_formats_2026-08-26T13-35-50/pdf/recibo_parqueadero_-_carta.pdf`
- OK | Recibo parqueadero - pos | parqueadero | pos | paginas PDF: 1 | captura: `test_runs/qa_print_formats_2026-08-26T13-35-50/screenshots/recibo_parqueadero_-_pos.png` | PDF: `test_runs/qa_print_formats_2026-08-26T13-35-50/pdf/recibo_parqueadero_-_pos.pdf`
- OK | Ticket de turno - carta | turno_atencion | carta | paginas PDF: 1 | captura: `test_runs/qa_print_formats_2026-08-26T13-35-50/screenshots/ticket_de_turno_-_carta.png` | PDF: `test_runs/qa_print_formats_2026-08-26T13-35-50/pdf/ticket_de_turno_-_carta.pdf`
- OK | Ticket de turno - pos | turno_atencion | pos | paginas PDF: 1 | captura: `test_runs/qa_print_formats_2026-08-26T13-35-50/screenshots/ticket_de_turno_-_pos.png` | PDF: `test_runs/qa_print_formats_2026-08-26T13-35-50/pdf/ticket_de_turno_-_pos.pdf`
- OK | Factura electronica extensa multipagina - carta | factura_extenso | carta | paginas PDF: 6 | captura: `test_runs/qa_print_formats_2026-08-26T13-35-50/screenshots/factura_electronica_extensa_multipagina_-_carta.png` | PDF: `test_runs/qa_print_formats_2026-08-26T13-35-50/pdf/factura_electronica_extensa_multipagina_-_carta.pdf`
- OK | Recibo de venta extenso multipagina - carta | recibo_extenso | carta | paginas PDF: 6 | captura: `test_runs/qa_print_formats_2026-08-26T13-35-50/screenshots/recibo_de_venta_extenso_multipagina_-_carta.png` | PDF: `test_runs/qa_print_formats_2026-08-26T13-35-50/pdf/recibo_de_venta_extenso_multipagina_-_carta.pdf`
- OK | Ticket real turnos atencion - pos | turno_atencion_real | pos | paginas PDF: 1 | captura: `test_runs/qa_print_formats_2026-08-26T13-35-50/screenshots/ticket_real_turnos_atencion_-_pos.png` | PDF: `test_runs/qa_print_formats_2026-08-26T13-35-50/pdf/ticket_real_turnos_atencion_-_pos.pdf`
- OK | Recibo real parqueadero - pos | parqueadero_real | pos | paginas PDF: 1 | captura: `test_runs/qa_print_formats_2026-08-26T13-35-50/screenshots/recibo_real_parqueadero_-_pos.png` | PDF: `test_runs/qa_print_formats_2026-08-26T13-35-50/pdf/recibo_real_parqueadero_-_pos.pdf`
