# Manifiesto de evidencia PSI-000

Empresa autorizada: Powerful Control System (`empresa_id=12`).

Este directorio conserva evidencia sintetica y visual de la preparacion de
Productos, Servicios, Compras e Inventario. No contiene contrasenas, cookies,
tokens, cabeceras de autorizacion ni datos comerciales de una factura real.

## Grupos

- `productos_staging_*_5b4d9ac0.png`: catalogo de Productos en escritorio y
  movil sobre el candidato `5b4d9ac0`.
- `inventario_avanzado_staging_*_efbde84d*.png`: flujo real de lote, serial,
  reserva, confirmacion y valorizacion antes/despues, incluido movil, sobre
  `efbde84d`.
- `compras_avanzadas_staging_*_efbde84d*.png`: requisicion, cotizacion,
  aprobacion y recepciones reales de QA antes/despues sobre `efbde84d`.
- `github-e2e-32940533286/test_runs/qa_e2e_buttons_*`: barrido autenticado de
  14 paginas, 622 botones inventariados, ocho clics seguros y cero hallazgos.
- `github-e2e-32940533286/test_runs/qa_print_formats_*`: render de 20 formatos
  imprimibles y sus resultados estructurados.
- `github-e2e-32940533286/documentos/reportes_profesionales/`: matrices de
  roles y comprobantes generadas por el mismo workflow.

Las capturas y el reporte del SHA final se agregan al cerrar CI y la repeticion
desplegada. La observacion de papel fisico se registra por separado porque una
cola tomada no demuestra por si sola que la impresora produjo papel.
