# Instrucciones del agente (copilot-instructions)

Estas instrucciones son de aplicación para el agente que trabaja en este repositorio y refuerzan las reglas almacenadas en `/memories/Agente Go.md`.

Reglas obligatorias:

- Regla principal: Siempre usar lenguaje Go puro sin librerías externas cuando aplique el desarrollo del proyecto.
  - Antes de proponer o añadir cualquier dependencia externa, el agente debe solicitar autorización explícita al usuario.
  - La propuesta de dependencia debe incluir: motivo técnico, alternativa en Go puro, impacto estimado y referencia a `documentos/historial_de_cambios`.
  - Si el usuario autoriza la dependencia, el agente debe documentar la decisión en `documentos/historial_de_cambios` con fecha, archivos afectados y la justificación.

- Regla de aplicación automática: Siempre que el agente vaya a ejecutar una acción que implique añadir/importar una librería externa (modificar `go.mod`, agregar paquetes importados, o instalar binarios), debe primero confirmar con el usuario y registrar la autorización en `documentos/historial_de_cambios`.

Comportamiento del agente:

- Leer `documentos/descripcion_del_proyecto` antes de comenzar tareas relacionadas con el proyecto y alinearse con sus restricciones.
- Actualizar `/memories/Agente Go.md` si las reglas cambian y notificar al usuario.
- No imprimir secretos ni valores sensibles en la consola o en los commits.

- Regla de trazabilidad automática: Cada vez que el agente marque una tarea como completada usando la herramienta `task_complete`, debe añadir una entrada en `documentos/historial_de_cambios` con fecha, archivos afectados y una breve descripción.

- Regla de documentación de archivos: Cada vez que el agente cree un archivo nuevo en el repositorio, debe añadir y describir el archivo en `documentos/descripcion_de_archivos` con ruta y propósito breve. Esta acción debe ocurrir inmediatamente después de crear el archivo.

- Regla específica para bases de datos: El agente debe leer y, cuando corresponda, actualizar `documentos/descripcion_de_las_bases_De_datos` antes de realizar cualquier cambio en esquemas, migraciones o manipulaciones masivas de datos. Además debe registrar la operación en `documentos/historial_de_cambios` y pedir confirmación del usuario si hay riesgo de pérdida de datos.

- Regla de diagramas y estructura del código: Antes de cambios estructurales, el agente debe revisar `documentos/diagramas/estructura_del_codigo.md` y usar `documentos/diagramas/` como carpeta oficial para diagramas técnicos. Si cambia arquitectura, flujos o integraciones, debe actualizar esos diagramas y registrar la trazabilidad en `documentos/descripcion_de_archivos` y `documentos/historial_de_cambios`.

Implementación práctica:

- Si el agente detecta una dependencia externa en el código (por ejemplo, import fuera de la stdlib o `require` en `go.mod`), debe detenerse y preguntar: "¿Autorizas añadir la dependencia X?".
- Tras la autorización, el agente aplica el cambio y añade una entrada en `documentos/historial_de_cambios` con la justificación técnica.

---
Fecha: 2026-03-26
