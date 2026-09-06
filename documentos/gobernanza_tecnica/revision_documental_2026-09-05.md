# Revisión y reorganización documental del 2026-09-05

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-05.

## Solicitud y alcance

Organizar la documentación de PCS para trabajo empresarial, escalabilidad y transferencia a ingenieros/agentes, creando faltantes y consolidando redundancias. Se revisó la estructura completa de texto propio versionable y las fuentes principales; no se leyó semánticamente cada línea de los 11 MB de historia ni se certificó todo el producto contra normas.

El corte inicial identificó **442 documentos de texto y 10.996.260 bytes** antes de reorganizar. El catálogo final amplía la cobertura a metadatos documentales JSON/YAML, contratos procesables y diagramas textuales bajo `documentos/`. Excluye archivos ignorados/privados y binarios. No deben compararse ambos totales como si tuvieran el mismo alcance.

## Hallazgos y cambios

| Hallazgo | Tratamiento |
| --- | --- |
| Faltaba README raíz y ruta breve de incorporación | Portada, portal, incorporación y CONTRIBUTING |
| Contextos acumulaban cronologías y órdenes incompatibles | Contextos breves, estado de entrega y snapshots de versiones previas |
| Contexto específico mantenía Plan 110 activo contra el contexto general | Referencia corregida; planes generales identificados como históricos |
| Descripción del proyecto asignaba migraciones a API y delegación automática | Corrección puntual y alcance de coordinación alineado con AGENTS |
| Decisiones asumían ausencia de producción para justificar defaults/retiros | Sustitución por verificación del entorno, alcance e identificación de consumidores |
| Faltaba relación entre requisitos, diseño y aceptación | Requisitos transversales, ejemplos de trazabilidad y estrategia QA |
| Arquitectura dispersa sin vistas de interesados/escalamiento | Descripción por vistas y correspondencias con archivos existentes |
| Faltaba síntesis de amenazas, privacidad y ownership de datos | Modelo inicial de amenazas, gobierno de datos y brechas organizacionales |
| Metas operativas podían confundirse con SLA o resultados | Objetivos preservados, metodología de medición y limitaciones explícitas |
| No había catálogo de todo el corpus con autoridad y revisión | Catálogo Markdown/JSON, política de fuentes mantenidas y control CI |

## Conservación y retiros

Se preservan los cambios previos del árbol de trabajo. Los contextos general/Codex, README documental, resumen raíz y antiguo índice de gobernanza se conservan en `documentos/historico/2026-09-05/`, con origen identificado y enlaces relativos ajustados donde correspondía.

Se retiraron `documentos/production_deployment_runbook.md` y `documentos/production_rollback_runbook.md` después de incorporar sus precondiciones, pasos y restricciones al [runbook principal de release](runbooks/runbook_release_profesional.md). Se actualizó el registro que los referenciaba. El contenido original permanece recuperable en Git. La retirada afecta documentación, no scripts ni despliegue.

El duplicado exacto inicial estaba en informes pertenecientes a dos ejecuciones E2E de productos/inventario. Se conservó cada copia para no separar evidencia de su ejecución. No se borraron planes, auditorías ni artefactos por tamaño o antigüedad.

## Comprobaciones y evidencia

Comandos de validación de esta entrega desde la raíz:

```powershell
node --test tools/docs_catalog.test.mjs
node tools/docs_catalog.mjs --write
node tools/docs_catalog.mjs --check
git diff --check
```

Resultados observados en el corte de esta entrega:

| Comprobación | Resultado |
| --- | --- |
| Tests del validador | 6 aprobados, 0 fallidos, 0 omitidos |
| Catálogo | 486 documentos/artefactos textuales inventariados |
| Fuentes principales revisadas | 31 con responsable, fecha y control estricto |
| Enlaces/anclas comprobados | 0 hallazgos después de reparar 21 referencias heredadas |
| `--check` | Sin drift ni errores bloqueantes |
| `git diff --check` | Aprobado en el árbol observado |
| CI remota | Pasos incorporados; no se ejecutó GitHub Actions en esta tarea |

Los 137 documentos clasificados como `referencia_por_validar` conservan su detalle y requieren cotejo semántico por módulo; no se presentan como revisión completa de sus afirmaciones. El catálogo también diferencia 274 evidencias, 22 históricos, 20 generados, una referencia externa y un control documental. Los conteos son un corte y cambian al regenerar el catálogo con nuevas fuentes. El validador usa Git y biblioteca estándar de Node, sin dependencias nuevas. Las pruebas del validador usan repositorios temporales y comprueban fallos reales por drift, enlaces, UTF-8, metadatos y archivos ausentes.

No se ejecutaron tests funcionales Go, migraciones, arranque PCS, sesiones UI, despliegue, pagos, emisión fiscal, proveedores ni hardware por esta tarea documental. Las pruebas de auditorías anteriores mantienen su fecha y autoría; no se presentan como nuevas.

## Brechas y mantenimiento

La base nueva está orientada por referencias técnicas internacionales consultadas en fuentes oficiales, con límites en el [marco](marco_documental.md). No acredita certificación ISO, nivel ASVS, preparación productiva ni cumplimiento legal.

El [registro de riesgos](riesgos_y_brechas.md) conserva la revisión semántica por módulo, trazabilidad exhaustiva, inventarios generados, ownership nominal, privacidad, capacidad y restore pendientes. Un ingeniero debe revisar estas brechas al decidir el alcance, sin convertirlas en hechos ya resueltos por disponer de documentación.
