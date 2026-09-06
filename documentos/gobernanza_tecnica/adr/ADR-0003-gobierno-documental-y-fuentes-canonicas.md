# ADR-0003 — Gobierno documental y fuentes principales

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-05.

Estado de decisión: aceptada para la reorganización documental solicitada por el usuario. No aprueba cambios funcionales ni certificaciones.

## Contexto

El inventario inicial de texto propio contiene 442 documentos y cerca de 11 MB. Las entradas de contexto crecieron como historiales y mezclaron instrucciones incompatibles sobre planes, bootstrap y disponibilidad fiscal. Existen contratos, runbooks, diagramas y consumidores de rutas documentales que conviene preservar.

## Alternativas

| Alternativa | Ventaja | Consecuencia |
| --- | --- | --- |
| Reescribir y mover todo a una carpeta nueva | Uniformidad inmediata aparente | Rompe consumidores y puede perder historia o atribuir revisión no realizada |
| Añadir otro resumen sin reglas de mantenimiento | Cambio pequeño | Conserva contradicciones y vuelve a crecer la duplicación |
| Entradas breves, autoridad por tema, catálogo y control automático | Orientación clara y mantenimiento incremental verificable | Requiere revisar semántica heredada por módulo y conservar brechas visibles |

## Decisión

Adoptar la tercera alternativa. Mantener rutas existentes que son consumidas por agentes, scripts o paneles. Separar historia de contexto vigente; incorporar requisitos, vistas de arquitectura, QA, seguridad/datos e incorporación. Gobernar todo el corpus mediante catálogo y estados explícitos, sin declarar revisados los documentos heredados por el mero hecho de inventariarlos.

Consolidar las recetas breves de deploy/rollback en el runbook de release existente. Conservar evidencia por ejecución aunque repita contenido. No agregar dependencias ni cambiar runtime de aplicación.

## Consecuencias y validación

El catálogo es un derivado reproducible; la política de estados y fuentes se revisa como código. Un control CI detecta drift y enlaces rotos en fuentes mantenidas; los hallazgos heredados se conservan visibles. La revisión humana sigue siendo necesaria para exactitud funcional y conformidad normativa.

Los contextos previos se preservan como snapshots y los archivos retirados se recuperan por Git. Una reversión documental debe mantener cambios previos del usuario y no reactivar planes históricos como órdenes. El [marco](../marco_documental.md), [informe](../revision_documental_2026-09-05.md) y [brechas](../riesgos_y_brechas.md) registran alcance y evidencia.
