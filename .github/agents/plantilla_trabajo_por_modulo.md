# Plantilla de trabajo por módulo

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Se elimina la delegación automática contradictoria y se mantiene una única autoridad en AGENTS.md.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

## Datos de aislamiento cuando intervienen agentes

- Worktree: `D:\powerfulcontrolsystem.worktrees\...`
- Rama: `codex/<fecha>-<tarea>-<id>`
- SHA base: `<origin/main>`
- Rutas/globs exclusivos asignados: `<lista>`
- Hotspots reservados al integrador requeridos: `<ninguno o lista>`

El worktree se crea con `scripts\agent_worktree.ps1`; no editar el checkout de
integración ni ejecutar `rs.ps1` o `actualizar_repositorio.ps1`.

1. Definir objetivo, alcance, requisito y criterio de aceptación.
2. Leer contexto general/Codex, fuente del módulo y comandos aplicables.
3. Identificar rutas, tablas, permisos, UI, integraciones y riesgo de datos.
4. Aplicar los frentes backend, frontend y QA según impacto. Delegar únicamente
   cuando el usuario haya pedido agentes.
5. Implementar respetando cambios previos, tenant y autoridad del servidor.
6. Verificar con evidencia proporcional: comando, candidato, entorno, fecha,
   resultado y omisiones. Identificar qué requiere PostgreSQL, UI o proveedor.
7. Actualizar contrato/fuentes, registro de archivos, historial y catálogo.
8. Cerrar con resultado y riesgos concretos; no presentar código como evidencia
   de producción ni una propuesta como decisión aprobada.

## Entrega al coordinador

- SHA final y estado `git status --short` limpio.
- Archivos tocados y cualquier solapamiento detectado.
- Comandos, resultado, entorno y omisiones.
- Efectos externos ejecutados o explícitamente no ejecutados.
- PR o siguiente paso de integración; no retirar el worktree por cuenta propia.

Los módulos críticos exigen revisión de todas las capas afectadas. Ejemplo:
un pago aprobado sin correo requiere revisar persistencia/idempotencia,
feedback visible y entrega, sin crear un nuevo pago para reenviar la notificación.

## Fuentes y aceptación de la revisión

[AGENTS.md](../../AGENTS.md).

Requisitos aplicables: PCS-REQ-016 ([matriz transversal](../../documentos/requisitos/especificacion_y_trazabilidad.md)).
