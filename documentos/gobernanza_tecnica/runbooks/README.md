# Catalogo de runbooks operativos

Fecha: 2026-04-30
Estado: vigente

Los runbooks convierten fallas repetibles o diagnosticos delicados en procedimientos concretos para desarrollo, QA y soporte.

## Estructura obligatoria de un runbook

1. sintoma
2. alcance del incidente
3. fuentes de evidencia
4. verificaciones iniciales
5. causas probables
6. acciones de recuperacion
7. validacion posterior
8. contratos o ADRs relacionados

## Catalogo inicial

- `runbook_checkout_licencias.md`: creado y actualizado para Smart Checkout, fallback clasico Epayco por POST y XML `AccessDenied`.
- `runbook_estaciones_sensores_ventas_simple.md`: creado.
- `runbook_arranque_postgresql_tunel_local.md`: creado.
- `runbook_dian_set_pruebas_y_diagnostico_oficial.md`: creado.
- `runbook_alertas_reinicio_y_monitoreo_gmail_smtp.md`: creado.
- `runbook_reportes_programados_y_exportaciones_contables.md`: creado.
- `runbook_soporte_remoto_sesiones_y_dispositivos.md`: creado.
- `runbook_cierre_periodo_y_conciliacion_bancaria.md`: creado.
- `runbook_contingencias_integraciones_bancarias_y_conectores.md`: creado.
- `runbook_reconciliacion_documental_fiscal_y_contable_externa.md`: creado.
- `runbook_versionado_documental_y_firmas_externas.md`: creado.
- `checklist_evidencia_documental_para_qa_y_soporte.md`: creada.
- `runbook_tls_staging_y_servicios_plan_105.md`: recuperacion segura de TLS
  vencido en staging, OnlyOffice y Nextcloud, con validacion externa obligatoria.

## Runbooks prioritarios siguientes

1. checklist operativa transversal para QA y soporte en flujos documentales
2. cobertura adicional para evidencia regulatoria automatizable
