# Perfiles de trabajo del repositorio

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Se elimina la delegación automática contradictoria y se mantiene una única autoridad en AGENTS.md.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

La autoridad operativa reside en [AGENTS.md](../../AGENTS.md). Los perfiles
representan responsabilidades para la revisión integrada, no procesos que deban
iniciarse automáticamente.

- [Coordinación](agente_go.agent.md): arquitectura, integración y cierre.
- [Backend y datos](agente_backend_db.agent.md): Go, PostgreSQL, permisos y negocio.
- [Frontend y UX](agente_frontend_ux.agent.md): interfaz y flujo operativo.
- [QA y operación](agente_qa_operacion.agent.md): evidencia, runtime y recuperación.

Usar el [protocolo](protocolo_delegacion.md) y la
[plantilla](plantilla_trabajo_por_modulo.md). Solo delegar cuando el usuario pide agentes.

## Fuentes y aceptación de la revisión

[AGENTS.md](../../AGENTS.md).

Requisitos aplicables: PCS-REQ-016 ([matriz transversal](../../documentos/requisitos/especificacion_y_trazabilidad.md)).
