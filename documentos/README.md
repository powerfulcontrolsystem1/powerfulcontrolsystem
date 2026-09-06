# Portal de documentación de PCS

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-05.


Esta es la entrada para ingenieros, operación, producto, QA y agentes. Cada tema tiene una fuente principal; el catálogo incluye también los documentos heredados y la evidencia histórica.

## Rutas de lectura

| Necesidad | Lectura |
| --- | --- |
| Entender el sistema | [Contexto general](contexto_general_del_sistema.md) → [arquitectura](arquitectura/descripcion_arquitectura.md) → [mapa de módulos](mapa_modulos.md) |
| Incorporarse al equipo | [Guía de incorporación](desarrollo/incorporacion.md) → [contribución](../CONTRIBUTING.md) → [comandos](comandos_codex.md) |
| Trabajar como agente | [AGENTS](../AGENTS.md) → [contexto general](contexto_general_del_sistema.md) → [contexto IA](contexto_codex.md) → [índice específico](contexto_especifico_del_sistema.md) |
| Definir o cambiar funciones | [Requisitos y trazabilidad](requisitos/especificacion_y_trazabilidad.md) → [flujos](flujos_operativos.md) → [contratos](gobernanza_tecnica/contratos/README.md) |
| Cambiar datos o API | [Gobierno de datos](arquitectura/gobierno_datos.md) → [estructura BD](estructura_bd.md) → [API](api/README.md) |
| Revisar seguridad | [Política privada](../SECURITY.md) → [amenazas y privacidad](seguridad/modelo_amenazas_y_privacidad.md) → [checklist multiempresa](checklist_seguridad_endpoint_multiempresa.md) |
| Probar y liberar | [Estrategia QA](calidad/estrategia_verificacion.md) → [release](gobernanza_tecnica/runbooks/runbook_release_profesional.md) → [checklist](release_checklist.md) |
| Operar o recuperar | [Observabilidad](observability_runbook.md) → [incidentes](operacion/incidentes_y_continuidad.md) → [recuperación](gobernanza_tecnica/runbooks/runbook_recuperacion_desastre_docker_vps.md) |
| Evaluar capacidad | [Arquitectura y escalamiento](arquitectura/descripcion_arquitectura.md) → [objetivos SLO/RTO/RPO](gobernanza_tecnica/slo_sla_operativo.md) |
| Revisar pendientes | [Estado actual](estado_actual.md) → [registro de riesgos y brechas](gobernanza_tecnica/riesgos_y_brechas.md) |
| Mantener documentación | [Marco documental](gobernanza_tecnica/marco_documental.md) → [plantillas](gobernanza_tecnica/plantillas_documentales.md) → [catálogo completo](catalogo_documental.md) |

## Fuentes de referencia existentes

- [Descripción funcional](descripcion_del_proyecto), [descripción de módulos](descripcion_de_modulos) y [matriz de permisos](matriz_roles_permisos_pos_multiempresa.md): entradas actuales por área; las cronologías anteriores se conservan como referencias históricas enlazadas.
- [Estructura del código](diagramas/estructura_del_codigo.md), [estructura BD](estructura_bd.md) y [paquete de diagramas](diagramas/documentacion_tecnica_completa.md): detalle técnico y derivados; su existencia no prueba sincronización con producción.
- [Decisiones permanentes](decisiones_tecnicas.md) y [ADR](gobernanza_tecnica/README.md): restricciones y motivos.
- [Changelog](CHANGELOG.md) e [historial](historial_de_cambios): evolución; no instrucciones operativas.

## Qué es vigente

Los documentos nuevos o reconciliados muestran responsable, estado y fecha. En el [catálogo](catalogo_documental.json), `vigente` describe autoridad documental, no implementación certificada; `referencia_por_validar` exige revisión contextual antes de actuar; `historico`, `evidencia`, `generado` y `referencia_externa` no dan órdenes.

Los planes 101–110 y el plan final de producción son antecedentes. No hay una orden general de ejecutar esos planes. El trabajo se decide por alcance autorizado, contrato aplicable y evidencia del candidato.

Para encontrar una ruta sin leer el historial completo:

```powershell
rg -n "nombre_del_modulo" documentos/mapa_modulos.md documentos/catalogo_documental.md
rg -n "nombre_del_contrato" documentos/gobernanza_tecnica
node tools/docs_catalog.mjs --check
```

## Revisión documental consolidada

[Informe y pendientes del 2026-09-06](gobernanza_tecnica/revision_semantica_2026-09-06.md) y [matriz por archivo](gobernanza_tecnica/matriz_revision_semantica_2026-09-06.json).
