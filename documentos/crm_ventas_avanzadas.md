# CRM y ventas avanzadas

Estado: Vigente. Responsable: Ingeniería del módulo. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- La superficie crm_avanzado reutiliza CRM unificado y consultas comerciales; scoring/forecast son indicadores calculados, no ventas ni recaudo confirmados.
- Las referencias secundarias y campos controlados por servidor tuvieron correcciones locales posteriores al QA de mayo. La aceptación A/B sigue siendo necesaria.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

Ampliacion del modulo `clientes`/CRM comercial. No reemplaza clientes, leads ni ventas: construye una capa gerencial encima de `crm_leads`, `crm_interacciones`, `crm_campanas`, `empresa_cotizaciones_venta` y `empresa_pedidos_venta`.

## Alcance

- Metas comerciales por periodo, responsable y canal.
- Dashboard de pipeline, forecast ponderado, cotizaciones abiertas y pedidos abiertos.
- Embudo por estado de lead con valor, probabilidad y forecast.
- Scoring de leads con recomendacion de accion comercial.
- Agenda de proximos contactos e interacciones.
- Salud comercial, valor en riesgo, leads sin contacto, oportunidades estancadas y plan de accion priorizado.
- Rendimiento por responsable y canal de adquisicion para revision gerencial.
- Conversion de lead a cotizacion de venta.

## Rutas

- `GET /api/empresa/crm_avanzado?action=dashboard&empresa_id=ID&periodo=YYYY-MM`
- `GET /api/empresa/crm_avanzado?action=metas&empresa_id=ID`
- `GET /api/empresa/crm_avanzado?action=scores&empresa_id=ID`
- `POST /api/empresa/crm_avanzado?action=meta`
- `POST /api/empresa/crm_avanzado?action=cotizacion_desde_lead`
- `POST /api/empresa/crm_avanzado?action=seed_demo`

## Seguridad

Usa `WithEmpresaCRMUnificadoPermissions`, pagina `linkCRMComercial` y el modulo/licencia `crm_unificado`. Todas las tablas y consultas se filtran por `empresa_id`.

## QA

La prueba de Motel Calipso crea un lead, interaccion, meta comercial, cotizacion desde lead y valida dashboard/scoring/forecast. La cobertura unitaria valida normalizacion de metas, scoring, alertas, salud comercial y acciones priorizadas.

## Fuentes y aceptación de la revisión

[crm_ventas_avanzadas.go](../backend/handlers/crm_ventas_avanzadas.go), [crm_ventas_avanzadas.go](../backend/db/crm_ventas_avanzadas.go), [crm_ventas_avanzadas_test.go](../backend/db/crm_ventas_avanzadas_test.go), [main.go](../backend/main.go), [empresa_permisos.go](../backend/handlers/empresa_permisos.go), [cierre_reparaciones_produccion_seguridad_2026-09-05.md](cierre_reparaciones_produccion_seguridad_2026-09-05.md).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).
