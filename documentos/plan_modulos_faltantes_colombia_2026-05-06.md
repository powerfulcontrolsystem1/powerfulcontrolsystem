# Plan de modulos importantes faltantes

Fecha: 2026-05-06
Estado: implementado como nucleo compartido inicial

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

5. Mesa de ayuda empresarial (retirada luego en favor de tickets propios)
- Tickets internos y externos.
- SLA, prioridades, categorias, responsables, estados y comentarios.
- Base de conocimiento, evidencias y tablero de soporte.
- Integracion con clientes, activos, usuarios, propiedad horizontal y WMS.

6. Calidad, procesos y no conformidades
- Procesos, procedimientos, checklists, auditorias internas.
- No conformidades, acciones correctivas/preventivas y responsables.
- Indicadores de calidad, hallazgos y seguimiento.

## Fases de implementacion

Implementacion 2026-05-06:
- Se creo un nucleo comun para los seis modulos en `backend/db/modulos_empresariales_colombia.go`, evitando duplicar tablas, handlers y UI.
- Se agregaron APIs privadas por empresa: `/api/empresa/bancos_pagos`, `/api/empresa/gestion_documental`, `/api/empresa/cumplimiento_kyc`, `/api/empresa/contratos_obligaciones` y `/api/empresa/calidad_procesos`. La mesa de ayuda heredada fue reemplazada posteriormente por `/api/empresa/tickets_ayuda`.
- Se agregaron pantallas administrativas, permisos, licencias, menu empresarial, datos demo y documentacion.
- Todas las entidades quedan aisladas por `empresa_id` y discriminadas por `modulo`.
- Continuacion de fases: se agregan plantillas por modulo, seguimiento profesional, cambio rapido de estado, bitacora enriquecida y filtros por estado, todo sobre el mismo nucleo compartido.
- Continuacion de fases: se agrega reporte ejecutivo compartido con vencimientos, criticidad, responsables, valor pendiente y recomendaciones automaticas por empresa.
- Continuacion de fases: se agrega importacion masiva CSV/JSON de registros con validacion y bitacora de importacion.
- Continuacion de fases: se agrega gestion compartida de evidencias y soportes por registro, con validacion por empresa/modulo y bitacora automatica.
- Continuacion de fases: se agrega flujo compartido de aprobaciones por nivel, responsable y decision, con actualizacion de estado y bitacora.
- Continuacion de fases: se agrega gestion compartida de tareas y compromisos por registro, con responsable, prioridad, vencimiento y estados operativos.
- Continuacion de fases: se agrega expediente 360 por registro, consolidando eventos, evidencias, aprobaciones y tareas con resumen y recomendacion operativa.
- Continuacion de fases: se agrega agenda compartida de vencimientos y alertas, con severidad y enlace directo al expediente.
- Continuacion de fases: se agrega cierre controlado de registros, bloqueando cierre sin evidencias, con aprobaciones pendientes o tareas abiertas.
- Continuacion de fases: se agrega generador de plan de accion que convierte alertas de agenda en tareas, evitando duplicados abiertos.
- Continuacion de fases: se agrega tablero de responsables y carga, con registros, tareas, aprobaciones y recomendaciones por responsable.
- Continuacion de fases: se agrega medicion SLA y cumplimiento con semaforo, buckets de vencimiento y recomendaciones.
- Continuacion de fases: se agrega matriz de riesgo operativo con score, factores ponderados y recomendaciones.
- Continuacion de fases: se agrega exportacion profesional de auditoria en CSV multi-seccion con resumen, agenda, SLA, riesgo, responsables, tareas, aprobaciones, evidencias y bitacora.
- Continuacion de fases: se agrega busqueda avanzada compartida desde backend por texto, estado, tipo, categoria, prioridad, responsable, vencidos y proximos vencimientos.
- Continuacion de fases: se agregan acciones masivas controladas para cambiar estado, prioridad y responsable con bitacora por registro.

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

Fase 5: Mesa de ayuda empresarial
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
