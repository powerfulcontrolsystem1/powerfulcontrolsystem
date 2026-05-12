# CRM y ventas avanzadas

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
