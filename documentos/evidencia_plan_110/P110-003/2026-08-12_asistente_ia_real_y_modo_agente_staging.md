# P110-003 — Asistente IA real y propuesta sin ejecución

Fecha: 2026-08-12 (America/Bogota)  
Empresa: Powerful Control System (`empresa_id=12`) en **staging**.  
Alcance: navegador autenticado, sin adjuntos, sin confirmaciones mutantes y sin
cambios en producción.

## Matriz ejecutada

| Control | Entrada | Resultado visible |
|---|---|---|
| Consulta normal | marca neutra que solicita una respuesta fija | El asistente respondió la marca esperada y mostró que la respuesta estaba lista. |
| Modelo activo | consulta normal | La interfaz identificó el modelo configurado sin exponer credenciales. |
| Modo agente | propuesta explícitamente «sin ejecutar» de una tarea interna | Se mostró una propuesta de revisión humana; no apareció una confirmación ejecutada, tarea creada, mensaje enviado ni otra mutación. |

La prueba de modo agente confirma el límite operativo visible: una intención de
acción se mantiene como propuesta y no amplía permisos ni se confirma de forma
automática. Los intentos de clic sobre controles fuera de la zona visible del
navegador fueron rechazados por la propia automatización; el envío se completó
por el atajo de teclado soportado por la interfaz, sin alterar el flujo.

## Límites

No se adjuntó archivo ni se aprobó/confirmó propuesta. Permanecen pendientes
las matrices de proveedor caído, timeout, cancelación, doble envío, adjuntos
específicamente autorizados y las identidades efectivas por rol. P110-003 sigue
**parcial** y no habilita automatización contable.

Como respaldo estático de la matriz, `qa_roles_matrix.mjs --strict` y
`qa_module_contracts.mjs --strict` aprobaron; sus reportes temporales no
contienen secretos ni sustituyen la evidencia autenticada por cada rol.
