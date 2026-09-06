# Marco de gobierno documental

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-05.

## Propósito y alcance

Mantener información suficiente para desarrollar, revisar, operar, transferir y evolucionar PCS sin depender de una persona o de una conversación. Aplica a documentación propia versionada, referencias técnicas, contratos API, diagramas, evidencia y guías de agentes. El [catálogo](../catalogo_documental.json) registra cada documento de texto bajo control de Git, incluidos archivos nuevos no ignorados; los binarios y adjuntos externos se describen desde su documento de evidencia.

La existencia de documentos no demuestra conformidad integral con normas, certificación, cumplimiento legal ni madurez operativa. Esta adopción se basa en descripciones públicas de las referencias; no se ha realizado una evaluación de todas sus cláusulas ni adquirido sus textos completos.

## Referencias y adaptación al proyecto

Referencias consultadas el 2026-09-05. Las versiones se fijan para trazabilidad; revisarlas al cambiar el marco, sin actualizar automáticamente contratos.

| Referencia | Aplicación práctica elegida en PCS | Evidencia documental y límite |
| --- | --- | --- |
| [ISO/IEC/IEEE 15289:2019](https://www.iso.org/standard/74909.html) | Gestionar propósito, contenido, responsables y ciclo de vida de la información | Este marco y catálogo; no equivalen a una auditoría de sus requisitos |
| [ISO/IEC/IEEE 42010:2022](https://www.iso.org/standard/74393.html) | Describir interesados, preocupaciones, vistas, decisiones y correspondencias de arquitectura | [Descripción de arquitectura](../arquitectura/descripcion_arquitectura.md) |
| [ISO/IEC/IEEE 29148:2018](https://www.iso.org/standard/72089.html) | Requisitos identificables, verificables y relacionados con diseño y aceptación | [Especificación y trazabilidad](../requisitos/especificacion_y_trazabilidad.md); cobertura por módulo aún debe completarse |
| [ISO/IEC 25010:2023](https://www.iso.org/standard/78176.html) | Organizar atributos de calidad y escenarios medibles | [Estrategia QA](../calidad/estrategia_verificacion.md) y objetivos SLO; sin resultados inferidos |
| [NIST SP 800-218, SSDF 1.1](https://csrc.nist.gov/pubs/sp/800/218/final) | Preparación, protección del software, producción segura y respuesta a vulnerabilidades | [Contribución](../../CONTRIBUTING.md), [seguridad](../../SECURITY.md), CI y release |
| [OWASP ASVS 5.0.0](https://owasp.org/www-project-application-security-verification-standard/) | Referencia para casos de verificación de aplicaciones y servicios web | [Modelo de amenazas](../seguridad/modelo_amenazas_y_privacidad.md) y checklist; pendiente mapear y ejecutar requisitos aplicables, sin declarar nivel ASVS |

Los controles de PCS son una adaptación técnica propia. No se copian textos normativos ni se presentan porcentajes de cumplimiento sin matriz de evaluación y evidencia.

## Autoridad y resolución de conflictos

1. Instrucciones vigentes del usuario y reglas del repositorio determinan el alcance autorizado; un documento no concede por sí solo permiso para operar.
2. [Decisiones permanentes](../decisiones_tecnicas.md), ADR aceptados y contrato específico definen el comportamiento requerido.
3. Código, configuración no secreta, pruebas y artefactos del candidato describen la implementación observada. Si contradicen el contrato, registrar un defecto; no cambiar el requisito para hacer desaparecer el fallo.
4. [Estado actual](../estado_actual.md) y evidencia fechada delimitan qué se ha validado y en qué entorno.
5. Mapas y catálogos facilitan descubrimiento. Informes, planes e historia no reemplazan contratos ni mediciones nuevas.

No resolver discrepancias solo por fecha: comprobar alcance, estado y fuente. Bloquear únicamente la decisión que depende de la discrepancia, documentarla y continuar el trabajo independiente.

## Fuentes principales por información

| Información | Fuente principal | Mantener en otros archivos |
| --- | --- | --- |
| Identidad, reglas y orientación | Contexto general, AGENTS, portal documental | Enlaces breves |
| Requisitos y aceptación | Especificación transversal y contrato del módulo | ID del requisito y prueba |
| Arquitectura y decisiones | Descripción de arquitectura, decisiones y ADR | Correspondencias a código |
| Rutas, payloads, errores | Contrato del módulo y OpenAPI curado | Inventario generado como descubrimiento |
| Tablas y ownership | Estructura BD, migraciones y gobierno de datos | Nombre y enlace, sin SQL duplicado |
| Roles y permisos | Matriz de roles y contrato de autorización | Enlace y prueba negativa |
| Procedimiento operativo | Runbook específico | Enlace; nunca otra receta divergente |
| Disponibilidad y release | Estado actual y evidencia del candidato | Fecha, alcance y referencia |
| Evolución | Historial y changelog | Resumen del cambio, sin nuevas reglas enterradas |

## Estados y revisión

El registro externo permite gobernar documentos antiguos sin reescribir evidencias firmadas o generadas. `responsable` identifica un rol, no acredita que una persona haya aprobado el contenido.

| Estado del catálogo | Uso | Revisión |
| --- | --- | --- |
| `vigente` | Fuente principal reconciliada en esta revisión | En cada cambio relacionado; como máximo cada 90 días |
| `referencia_por_validar` | Documento heredado útil, con posible mezcla de historia y contrato | Responsable del módulo debe cotejarlo antes de usarlo para una decisión |
| `historico` | Plan, snapshot o changelog preservado | No se reactiva por lectura; una decisión nueva requiere fuente vigente |
| `evidencia` | Resultado de una ejecución fechada | Inmutable salvo corrección trazable; no garantiza un candidato posterior |
| `generado` | Derivado de código/herramienta | Regenerar con el generador y revisar el diff; no editar manualmente |
| `control_documental` | Política JSON y matriz manual de revisión | Editar con revisión; no es un inventario generado |
| `contrato_maquina` | Contrato YAML/JSON mantenido manualmente, con guía compañera | Revisar operaciones, schemas y compatibilidad; no sobrescribirlo con inventarios generados |
| `referencia_externa` | Material de terceros | Conservar procedencia, versión/licencia y fecha de consulta |

Un documento `vigente` incluye título único, propósito, responsable, estado, revisión, fuentes, alcance y límites. Los documentos nuevos usan Markdown UTF-8; contratos procesables conservan YAML/JSON. Los históricos grandes permanecen fuera de la ruta de lectura inicial.

## Responsabilidades

| Rol | Responsabilidad |
| --- | --- |
| Responsable de producto | Prioridad, alcance funcional y criterios de aceptación comercial |
| Coordinación técnica | Autoridad documental, arquitectura, revisión integrada y asignación de responsables |
| Ingeniería del módulo | Contratos, código, datos, interfaces, pruebas y actualización de referencias |
| QA/operación | Evidencia, SLO, incidentes, release y recuperación |
| Responsable de seguridad y privacidad | Amenazas, acceso, vulnerabilidades, retención y requisitos regulatorios aplicables |

Los nombres, suplencias y contacto de guardia deben mantenerse en el directorio privado de operación. Falta verificar su asignación organizacional; no se inventan personas ni aprobaciones en Git. CODEOWNERS existente sigue gobernando las revisiones configuradas.

## Ciclo de cambio documental

1. Identificar fuente principal y requisitos afectados; editarla sin duplicar el tema.
2. Registrar una decisión nueva mediante ADR cuando cambia una restricción o frontera; usar [plantillas](plantillas_documentales.md).
3. Registrar archivos nuevos en `documentos/descripcion_de_archivos`; dejar evolución en historial y changelog.
4. Ejecutar `node tools/docs_catalog.mjs --write` desde la raíz y revisar clasificación, enlaces y hallazgos.
5. Ejecutar `node tools/docs_catalog.mjs --check` y `git diff --check`. CI verifica el mismo catálogo.
6. Revisión humana de exactitud y aceptación. El validador no interpreta reglas de negocio ni verifica URLs externas, secretos o producción.

## Consolidar, archivar o retirar

Eliminar solo documentos sin información exclusiva útil, después de consolidar contenido, buscar consumidores en todo el repositorio y reparar enlaces. Registrar origen, destino, motivo y forma de recuperación mediante Git. Los informes de ejecuciones distintas se conservan aunque su texto coincida.

No mover artefactos consumidos por scripts o el catálogo técnico del panel super sin revisar ese consumidor. Los snapshots de contextos previos viven en `documentos/historico/2026-09-05/`; su clasificación invalida instrucciones antiguas sin perder conocimiento.

## Límite de esta revisión

Se inventarió todo el corpus de texto propio versionable, se verificaron enlaces locales y se revisaron fuentes principales y contratos representativos. El detalle semántico de todos los módulos, la trazabilidad completa por endpoint y la evaluación normativa cláusula por cláusula siguen en el [registro de brechas](riesgos_y_brechas.md). Clasificar un documento no equivale a revisar todo su contenido.
