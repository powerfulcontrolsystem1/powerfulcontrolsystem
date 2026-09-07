# Guía de contribución

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-05.

## Antes de cambiar

Leer [AGENTS](AGENTS.md), [contexto general](documentos/contexto_general_del_sistema.md) y [guía de incorporación](documentos/desarrollo/incorporacion.md). Identificar módulo, contrato, datos, permisos, requisitos y entorno. Revisar `git status --short` y conservar cambios ajenos.

Trabajar en una rama acotada, por defecto `codex/descripcion-del-cambio`, sin sobrescribir trabajo existente. No añadir dependencias ni modificar `go.mod` sin autorización explícita. PostgreSQL es el único motor; usar biblioteca estándar de Go cuando aplique.

## Implementación y revisión

1. Describir el problema, resultado esperado y requisitos afectados. Para decisiones transversales, registrar ADR.
2. Reutilizar el patrón del módulo; validar seguridad en backend, incluidas referencias secundarias del tenant.
3. Añadir pruebas proporcionales al riesgo y casos negativos/concurrentes para flujos críticos. No ejecutar efectos externos a partir de un ejemplo documental.
4. Actualizar las fuentes afectadas: contrato, datos, permisos, arquitectura, ayuda y runbook. Enlazar archivos nuevos desde la fuente o mapa que los consume.
5. Revisar diff, resultados y omisiones. Pasar [QA](documentos/calidad/estrategia_verificacion.md) y controles documentales.
6. Abrir revisión con problema, comportamiento final, alcance, pruebas, migración/compatibilidad y riesgos residuales. Aplicar CODEOWNERS configurado y revisión humana apropiada.

## Comprobación documental

Desde la raíz, con Node disponible:

```powershell
node tools/docs_catalog.mjs --write
node tools/docs_catalog.mjs --check
git diff --check
```

Revisar el catálogo generado antes de entregar; regenerar no certifica el contenido. Una nueva fuente principal se añade a la política del catálogo con responsable y fecha de revisión real. La evidencia pesada se conserva fuera de Git y el resultado mínimo se referencia desde la PR o el sistema de CI.

## Cierre y liberación

El cambio está listo para revisión cuando código y documentación coinciden y las pruebas necesarias tienen resultado o impedimento explícito. No contar pruebas omitidas como aprobadas. No cerrar un módulo crítico sin evidencia de las capas afectadas.

La integración no equivale a despliegue. Seguir el [runbook de release](documentos/gobernanza_tecnica/runbooks/runbook_release_profesional.md), con candidato inmutable, staging, recuperación y autorización aplicables. Reportar vulnerabilidades por [SECURITY](SECURITY.md).
