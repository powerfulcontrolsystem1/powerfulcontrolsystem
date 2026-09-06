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

### Aislamiento Git obligatorio

El checkout `D:\powerfulcontrolsystem` pertenece al integrador y se mantiene en
`main` limpio. Cada agente trabaja exclusivamente en un worktree propio bajo
`D:\powerfulcontrolsystem.worktrees\` y una rama
`codex/<fecha>-<tarea>-<id>`, creada con:

```powershell
.\scripts\agent_worktree.ps1 -Action Create -Task modulo-acotado
```

El coordinador asigna globs sin solapamiento. Quedan reservados para integración:
`AGENTS.md`, `CONTRIBUTING.md`, `.github/`, `backend/main.go`, migraciones y su
catálogo, `go.mod`, `go.sum`, estilos globales, historial/changelog, registro de
archivos y catálogos generados. Si dos tareas necesitan uno de esos archivos, el
coordinador secuencia la edición; no se resuelve compartiendo un mismo checkout.

El worktree se bloquea al crearse. El agente entrega su SHA, rutas, pruebas,
estado limpio y efectos externos; solo el coordinador lo desbloquea y retira con
`agent_worktree.ps1 -Action Remove`. La retirada exige árbol limpio y evidencia
de que el HEAD está integrado (ancestro de `origin/main` o head exacto de una PR
fusionada). Los squash merges se verifican por head de PR, no solo por
`merge-base`.

Ningún subagente ejecuta `rs.ps1`, `actualizar_repositorio.ps1`, despliegues,
limpieza masiva, merge ni push de integración. Esas acciones se realizan una vez
por el integrador desde el checkout limpio; el staging automático de
`actualizar_repositorio.ps1` no es seguro en un árbol compartido.

## Cierre integrado

Backend entrega causa, rutas/tablas y riesgos. Frontend entrega comportamiento
visible y dependencia de APIs. QA entrega comandos, entorno, resultados y
omisiones. El coordinador reconcilia el conjunto con contratos/documentación.
Una limitación del entorno se informa como pendiente, no como prueba aprobada.

## Fuentes y aceptación de la revisión

[AGENTS.md](../../AGENTS.md).

Requisitos aplicables: PCS-REQ-016 ([matriz transversal](../../documentos/requisitos/especificacion_y_trazabilidad.md)).
