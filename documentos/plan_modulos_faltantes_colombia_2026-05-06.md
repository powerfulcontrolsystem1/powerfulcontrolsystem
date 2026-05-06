# Plan de modulos importantes faltantes

Fecha: 2026-05-06
Estado: plan de trabajo

## Criterio

El sistema ya cubre POS, inventario, compras, produccion/MRP, logistica WMS, contabilidad Colombia, declaraciones tributarias, activos fijos, tesoreria/presupuesto, cobranza, portal contador, certificados tributarios, AIU, propiedad horizontal y verticales operativas. Los faltantes siguientes son modulos empresariales de alto impacto que suelen complementar suites tipo ERP/contable en Colombia.

## Modulos propuestos

1. Bancos y pagos masivos Colombia
- Extractos bancarios OFX/CSV/Excel.
- Conciliacion bancaria avanzada por reglas.
- Pagos masivos a proveedores/nomina con archivo bancario.
- Estados de transmision, rechazo y aprobacion.
- Integracion con tesoreria, CxP, nomina y contabilidad.

2. Gestion documental empresarial y aprobaciones
- Radicacion de documentos internos/externos.
- Versiones, expedientes, etiquetas y vencimientos.
- Flujos de aprobacion por rol, monto, area o centro de costo.
- Auditoria, adjuntos, busqueda y retencion documental.

3. Cumplimiento, KYC/KYB y riesgo LAFT
- Debida diligencia de clientes, proveedores, empleados y contratistas.
- Listas restrictivas parametrizables.
- Matriz de riesgo, señales de alerta, aprobaciones y bitacora.
- Reportes de cumplimiento y evidencia por tercero.

4. Contratos, obligaciones y firma electronica
- Contratos comerciales, laborales, arrendamiento, mantenimiento y proveedores.
- Hitos, renovaciones, polizas, vencimientos, responsables y alertas.
- Plantillas, anexos, aprobaciones y firma electronica externa o manual.

5. Mesa de ayuda / Helpdesk empresarial
- Tickets internos y externos.
- SLA, prioridades, categorias, responsables, estados y comentarios.
- Base de conocimiento, evidencias y tablero de soporte.
- Integracion con clientes, activos, usuarios, propiedad horizontal y WMS.

6. Calidad, procesos y no conformidades
- Procesos, procedimientos, checklists, auditorias internas.
- No conformidades, acciones correctivas/preventivas y responsables.
- Indicadores de calidad, hallazgos y seguimiento.

## Fases de implementacion

Fase 1: Bancos y pagos masivos
- Crear tablas por `empresa_id`.
- API `/api/empresa/bancos_pagos`.
- Pantalla en Centro financiero.
- Reglas de conciliacion y exportacion CSV base para bancos.
- Pruebas unitarias de matching y control de duplicados.

Fase 2: Gestion documental
- Crear nucleo documental reutilizable.
- API `/api/empresa/gestion_documental`.
- Pantalla con expedientes, documentos, versiones y aprobaciones.
- Conectar con compras, contabilidad, contratos y certificados.

Fase 3: Riesgo LAFT/KYC
- Maestro de evaluaciones por tercero.
- Matriz de riesgo y alertas.
- Pantalla de cumplimiento.
- Control por licencia `cumplimiento_kyc`.

Fase 4: Contratos y obligaciones
- Contratos, partes, hitos, anexos, polizas y vencimientos.
- Alertas y dashboard.
- Integracion con terceros, activos, propiedad horizontal y compras.

Fase 5: Helpdesk
- Tickets, SLA, comentarios y evidencias.
- Portal interno y publico opcional.
- Integracion con usuarios, clientes y activos.

Fase 6: Calidad y procesos
- Procesos, auditorias, checklists y no conformidades.
- Dashboard y reportes.

## Reglas transversales

- Todos los modulos deben tener `empresa_id`.
- No duplicar productos, terceros, contabilidad, inventario ni usuarios.
- Cada modulo debe tener permiso, licencia, menu, documentacion y pruebas.
- Cada pantalla debe adaptarse a `light`, `light-rose`, `light-gold`, `dark`, `dark-violet` y `dark-emerald`.
- Cada fase debe cerrar con `go test ./... -count=1`, `git diff --check`, prueba visual y prueba funcional con Motel Calipso.
