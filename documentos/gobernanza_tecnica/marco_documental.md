# Marco de gobierno documental

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-05.

## Propósito

Mantener la información necesaria para desarrollar, revisar, operar y transferir
PCS sin mezclarla con cronologías, volcados o resultados de candidatos retirados.
El [catálogo](../catalogo_documental.json) enumera el corpus actual; no certifica
producción ni sustituye una revisión funcional.

## Autoridad

1. La instrucción vigente del usuario y `AGENTS.md` determinan alcance y efectos.
2. Decisiones técnicas, ADR y contrato del módulo definen el comportamiento.
3. Código, migraciones, pruebas y configuración no secreta describen el snapshot.
4. [Estado actual](../estado_actual.md) y requisitos delimitan aceptación y brechas.
5. Mapas e inventarios facilitan descubrir; no reemplazan contratos ni medición.

Ante una contradicción, registrar el defecto y bloquear solo la decisión afectada.
No cambiar el requisito para hacer coincidir una implementación incorrecta.

## Fuentes principales

| Información | Fuente |
| --- | --- |
| Reglas para agentes | `AGENTS.md`, contexto general y contexto Codex |
| Alcance y navegación | Descripción del proyecto y mapa de módulos |
| Requisitos y aceptación | Especificación transversal y contrato del módulo |
| Arquitectura y datos | Arquitectura, decisiones/ADR, estructura BD y migraciones |
| API | Contrato del módulo y OpenAPI curado; inventario generado como apoyo |
| Roles y permisos | Matriz de roles y autorización efectiva del backend |
| Operación | Runbook específico, checklist de release y configuración privada del entorno |
| Estado | `estado_actual.md` y registro vigente de riesgos |
| Evolución | Commits y PR; no changelogs acumulativos dentro del snapshot |

## Estados del catálogo

| Estado | Uso |
| --- | --- |
| `vigente` | Fuente mantenida con responsable y fecha de revisión |
| `referencia_por_validar` | Fuente útil que debe cotejarse antes de decidir |
| `generado` | Derivado reproducible; regenerar y revisar, no editar a mano |
| `control_documental` | Política del catálogo |
| `contrato_maquina` | Contrato procesable mantenido manualmente |
| `referencia_externa` | Material de terceros con procedencia y versión |

Los estados `historico` y `evidencia` no se guardan en la rama operativa. La
evidencia pesada o sensible vive en CI o almacenamiento externo con retención y
acceso adecuados.

## Ciclo de cambio

1. Identificar la fuente principal, requisitos, consumidores y entorno.
2. Actualizar contrato, código, prueba y documentación sin duplicar el tema.
3. Enlazar archivos nuevos desde el mapa o la fuente que los consume.
4. Cambiar `estado_actual.md` solo si cambia una condición vigente.
5. Ejecutar `node tools/docs_catalog.mjs --write`, revisar el diff y luego
   `node tools/docs_catalog.mjs --check` y `git diff --check`.
6. Registrar en la PR pruebas, límites y referencia minimizada a evidencia externa.

No almacenar secretos, dumps, sesiones, datos reales, capturas autenticadas ni
salidas masivas en Git. Si una fuente deja de ser vigente, consolidar lo necesario
en su reemplazo y retirarla; el respaldo externo previo a la reescritura es la vía
de recuperación del corpus antiguo.

## Referencias de calidad

PCS adapta principios de ISO/IEC/IEEE 15289 y 42010 para información y
arquitectura, ISO/IEC/IEEE 29148 para requisitos, ISO/IEC 25010 para calidad,
NIST SSDF para desarrollo seguro y OWASP ASVS para casos de verificación. Estas
referencias no equivalen a certificación ni a evaluación completa de sus normas.
