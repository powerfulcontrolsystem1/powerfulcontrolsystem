# Plan de implementacion de gobernanza tecnica

Fecha: 2026-04-18
Estado: en ejecucion

## Objetivo

Implementar una capa documental empresarial que convierta el proyecto en un sistema mas predecible, auditable y escalable, con menor margen de error para cambios asistidos por Copilot y para evoluciones futuras del equipo.

## Principios de implantacion

- No duplicar la documentacion funcional existente; extenderla con artefactos de gobierno tecnico.
- Priorizar los flujos criticos con mayor riesgo operativo o multiempresa.
- Mantener cada artefacto pequeno, accionable y enlazado a una fuente canonica existente.
- Exigir trazabilidad documental en la misma iteracion del cambio tecnico.

## Fases

### Fase 0. Baseline canonico

Entregables:

- `documentos/README.md`
- `documentos/gobernanza_tecnica/README.md`
- este plan de implementacion

Resultado esperado:

- cualquier desarrollador o agente sabe que leer primero y donde vive cada fuente canonica.

### Fase 1. Cambio seguro obligatorio

Entregables:

- `documentos/gobernanza_tecnica/estandares_de_cambio_seguro.md`
- reglas de lectura previa por tipo de cambio
- checklist de validacion minima por area de riesgo

Resultado esperado:

- los cambios nuevos dejan de depender solo de memoria operativa y pasan a apoyarse en reglas explicitas del repo.

### Fase 2. ADRs base del sistema

Entregables minimos:

- ADR sobre frontera multiempresa y `empresa_id`
- ADR sobre PostgreSQL como runtime canonico y VPS como entorno productivo
- ADR sobre validacion de contexto esperado en flujos publicos de pago
- ADR sobre regularizacion de esquemas legacy al vuelo cuando sea obligatorio por compatibilidad

Resultado esperado:

- las decisiones estructurales dejan de ser reabiertas implicitamente en cada tarea.

### Fase 3. Contratos tecnicos por flujos criticos

Avance 2026-04-18:

- creado `contrato_checkout_licencias_publico.md`
- creado `contrato_estaciones_sensores_ventas_simple.md`
- creado `contrato_autenticacion_administrativa_y_usuarios_empresa.md`
- creado `contrato_venta_publica_empresarial_por_empresa.md`
- creado `contrato_permisos_contexto_y_wrappers_api_empresa.md`
- creado `contrato_facturacion_electronica_y_documentos_transaccionales.md`
- creado `contrato_reportes_contables_financieros_y_exportacion_multiformato.md`
- creado `contrato_soporte_remoto_por_empresa_y_mesa_tecnica_central.md`
- creado `contrato_conciliacion_bancaria_y_cierre_periodo_contable.md`
- creado `contrato_interoperabilidad_documental_contable_y_fiscal_externa.md`
- creado `contrato_integraciones_bancarias_y_conectores_externos.md`
- creado `contrato_repositorio_documental_y_firmas_externas.md`

Prioridad de contratos:

1. checkout publico de licencias
2. estaciones, sensores y ventas simples por estacion
3. autenticacion administrativa y de usuarios de empresa
4. venta publica empresarial por empresa
5. permisos_contexto y wrappers `/api/empresa`
6. facturacion electronica y documentos transaccionales
7. reportes contables, financieros y exportacion multiformato
8. soporte remoto por empresa y mesa tecnica central
9. conciliacion bancaria y cierre de periodo contable
10. interoperabilidad documental contable y fiscal externa
11. integraciones bancarias y conectores externos
12. repositorio documental y firmas externas

Resultado esperado:

- cada flujo critico tiene limites claros de entrada, salida, estados, side effects y validaciones obligatorias.
- los contratos documentales quedan reconciliados entre si cuando un mismo flujo cruza repositorio, firmas, compras, facturacion y exportes regulatorios.

### Fase 4. Runbooks operativos

Avance 2026-04-18:

- creado `runbook_checkout_licencias.md`
- creado `runbook_estaciones_sensores_ventas_simple.md`
- creado `runbook_arranque_postgresql_tunel_local.md`
- creado `runbook_dian_set_pruebas_y_diagnostico_oficial.md`
- creado `runbook_alertas_reinicio_y_monitoreo_gmail_smtp.md`
- creado `runbook_reportes_programados_y_exportaciones_contables.md`
- creado `runbook_soporte_remoto_sesiones_y_dispositivos.md`
- creado `runbook_cierre_periodo_y_conciliacion_bancaria.md`
- creado `runbook_contingencias_integraciones_bancarias_y_conectores.md`
- creado `runbook_reconciliacion_documental_fiscal_y_contable_externa.md`
- creado `runbook_versionado_documental_y_firmas_externas.md`

Prioridad de runbooks:

1. pago aprobado pero licencia no activa o correo incorrecto
2. estacion no refleja sensor o abre carrito incorrecto
3. arranque backend PostgreSQL por tunel local
4. fallas de Gmail/configuracion avanzada
5. errores DIAN y envio de set de pruebas
6. reportes programados y exportaciones contables
7. soporte remoto, sesiones y dispositivos
8. cierres de periodo contable y conciliacion extendida
9. contingencias de integraciones bancarias y conectores externos
10. reconciliacion documental fiscal y contable externa
11. versionado documental y firmas externas

Resultado esperado:

- soporte, QA y desarrollo pueden diagnosticar y recuperar incidentes sin improvisacion.
- los runbooks documentales distinguen evidencia vigente, firmas asociadas y exportes regulatorios sin confundir salida informativa con respaldo formal.

### Fase 5. Gobernanza continua

Entregables:

- regla de actualizacion documental por tipo de cambio
- criterio de aceptacion tecnica para nuevos modulos
- criterio minimo para cerrar tareas: pruebas, trazabilidad y artefacto aplicable actualizado

Resultado esperado:

- la gobernanza pasa de ser un esfuerzo puntual a una disciplina continua del repositorio.

## Roadmap recomendado de ejecucion

1. completar Fase 1 y Fase 2.
2. documentar contratos de pagos y estaciones.
3. abrir runbooks de pagos, sensores y arranque PostgreSQL.
4. extender contratos a autenticacion, venta publica y facturacion.
5. usar cada incidente real para enriquecer ADRs, contratos o runbooks en vez de dejar la leccion solo en memoria del agente.

## Criterios de salida por fase

- Fase 1 cerrada cuando todo cambio de backend, frontend critico o DB tenga checklist minimo y lectura previa definida.
- Fase 2 cerrada cuando las decisiones base del sistema no dependan de interpretacion informal.
- Fase 3 cerrada cuando los flujos publicos y multiempresa mas sensibles tengan contrato tecnico vigente.
- Fase 4 cerrada cuando los principales incidentes productivos se puedan diagnosticar con runbooks reproducibles.
- Fase 5 cerrada cuando la gobernanza se mantenga en cada iteracion sin depender de recordatorios ad hoc.
