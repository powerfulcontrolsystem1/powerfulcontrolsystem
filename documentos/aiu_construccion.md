# AIU construccion y contratos de obra

Actualizacion: 2026-05-06

El modulo AIU construccion agrega una capa empresarial para arquitectos, constructoras, contratistas de obra civil, remodelaciones y pequenas empresas que facturan contratos con Administracion, Imprevistos y Utilidad.

## Alcance funcional

- Contratos de obra por `empresa_id` con cliente, responsable, centro de costo, modalidad contractual, tipo de obra, riesgo, fechas, avance, estado y observaciones.
- Capitulos y conceptos de obra con cantidad, unidad, valor unitario y costo directo acumulado.
- Calculo automatico de Administracion, Imprevistos y Utilidad.
- Dos modelos compatibles con la referencia de Siigo:
  - `base_aiu_no_sumada`: la base AIU se informa/controla, pero no se suma completa al total de la factura.
  - `base_aiu_sumada`: Administracion + Imprevistos + Utilidad se suman al total.
- Base de impuestos configurable: solo utilidad, AIU total o costo directo + AIU.
- Retenciones operativas: retencion en la fuente sobre base IVA, retencion ICA sobre total, retencion IVA sobre IVA, amortizacion de anticipo, garantia y neto a cobrar.
- Flujo de estados controlado: borrador, cotizado, aprobado, en ejecucion, suspendido, facturado, cerrado y anulado, con validacion de transiciones y trazabilidad de aprobacion.
- Tablero profesional con totales, facturado, neto a cobrar, pendiente por facturar, alertas de avance alto sin facturar y contratos por estado.
- Filtros por estado/busqueda, facturas recientes y exportacion CSV desde la vista.
- Generacion de factura electronica AIU en `empresa_facturacion_documentos`, reutilizando el modulo de facturacion electronica, auditoria y permisos.
- Control independiente por licencia/rol mediante `aiu_construccion`.

## Rutas y archivos

- API empresarial: `/api/empresa/aiu_construccion`.
- Vista: `web/administrar_empresa/aiu_construccion.html`.
- Datos: `backend/db/aiu_construccion.go`.
- Handler: `backend/handlers/aiu_construccion.go`.
- Menu: `web/administrar_empresa/facturacion_electronica_menu.html`.

## Acciones API

- `GET action=dashboard`: KPI y ultimos contratos/facturas.
- `GET action=contratos`: lista contratos, con filtros `estado` y `q`.
- `GET action=facturas`: lista facturas AIU, opcionalmente por `contrato_id`.
- `GET action=detalle&id=...`: contrato con items y facturas.
- `GET action=reporte`: entrega contratos y facturas para reporterias/exportaciones.
- `POST action=calcular`: calcula un payload sin guardar.
- `POST action=contrato`: crea o actualiza contrato.
- `POST action=item`: agrega concepto/capitulo al contrato.
- `POST action=generar_factura`: registra factura AIU y documento electronico.
- `POST action=estado`: cambia estado con validacion de flujo y aprobacion.
- `POST action=seed_demo`: crea un ejemplo operativo.

## Verificacion

- `go test ./db -run Test.*AIU -count=1`
- `go test ./handlers -run "TestNormalizeFacturacionDocumentoElectronicoTipo|TestResolveFacturacionTransitionForDocumentosElectronicosNuevos" -count=1`
- `go test ./...`
