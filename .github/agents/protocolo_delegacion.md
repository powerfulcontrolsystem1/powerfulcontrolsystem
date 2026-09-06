# Protocolo de coordinación y delegación

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Se elimina la delegación automática contradictoria y se mantiene una única autoridad en AGENTS.md.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

## Regla y alcance

`agente_go` clasifica el cambio y aplica los frentes siguientes como checklist.
La delegación solo procede cuando el usuario pide agentes/subagentes.
No depende del color, tamaño o criticidad del módulo.

| Cambio | Revisión necesaria |
| --- | --- |
| Backend/BD | Reglas, tenant, transacciones, migraciones y pruebas; UI si cambia contrato |
| Frontend | Estados visibles, permisos efectivos, accesibilidad y responsive; backend si cambian datos |
| Operación | Entorno, candidato, comandos, recuperación y evidencia real |
| Pagos/licencias/venta pública/estaciones/carritos | Revisar conjuntamente negocio, interfaz y operación |
| Autenticación/permisos | Identidad, sesión, alcance, interfaz y pruebas negativas |
| Fiscal/reportes | Fuente, estado, formatos, privacidad y aceptación por familia |

## Cuando hay delegación solicitada

Asignar una tarea acotada y rutas de edición sin conflicto. Compartir fuentes,
restricciones y evidencia esperada; integrar resultados sin revertir cambios
ajenos. No delegar autoridad de compra, emisión, despliegue ni borrado por el
mero hecho de crear un agente.

## Cierre integrado

Backend entrega causa, rutas/tablas y riesgos. Frontend entrega comportamiento
visible y dependencia de APIs. QA entrega comandos, entorno, resultados y
omisiones. El coordinador reconcilia el conjunto con contratos/documentación.
Una limitación del entorno se informa como pendiente, no como prueba aprobada.

## Fuentes y aceptación de la revisión

[AGENTS.md](../../AGENTS.md).

Requisitos aplicables: PCS-REQ-016 ([matriz transversal](../../documentos/requisitos/especificacion_y_trazabilidad.md)).
