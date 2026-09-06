# Gobernanza técnica de PCS

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-05.

Este paquete reúne decisiones, contratos y procedimientos. El punto de entrada
general es el [portal documental](../README.md); leer primero el contexto general.

## Fuentes de gobierno

- [Marco documental](marco_documental.md): autoridad, estados, responsables y ciclo de revisión.
- [Requisitos](../requisitos/especificacion_y_trazabilidad.md): obligaciones y aceptación transversal.
- [Estándares de cambio seguro](estandares_de_cambio_seguro.md): reglas de implementación por capa.
- [Contratos](contratos/README.md): detalle por flujo crítico.
- [Runbooks](runbooks/README.md): procedimientos de diagnóstico y operación.
- [SLO/RTO/RPO](slo_sla_operativo.md): objetivos internos, separados de medición y SLA contractual.
- [Riesgos y brechas](riesgos_y_brechas.md): pendientes y criterios de cierre.
- [Plantillas](plantillas_documentales.md): estructuras reutilizables.

## Decisiones de arquitectura

- [ADR-0001: frontera multiempresa](adr/ADR-0001-frontera-multiempresa-empresa-id.md).
- [ADR-0002: PostgreSQL runtime](adr/ADR-0002-postgresql-runtime-canonico-vps.md).
- [ADR-0003: gobierno documental](adr/ADR-0003-gobierno-documental-y-fuentes-canonicas.md).
- [Decisión CxP](../arquitectura/adr_106_cxp_fuente_canonica.md): verificar su estado y evidencia antes de migrar datos.
- [Decisiones permanentes](../decisiones_tecnicas.md): restricciones técnicas del proyecto.

Los planes de adopción anteriores se conservan como antecedentes, sin autorización
de ejecución implícita. El [índice anterior](../historico/2026-09-05/gobernanza_tecnica_README.md)
preserva la evolución de este paquete. No inferir disponibilidad actual de sus
descripciones históricas de IA, firmas o integraciones.
