# Regla para agente_go sobre diagramas

Fecha de creacion: 2026-04-01

## Regla obligatoria

El agente_go debe considerar la carpeta documentos/diagramas/ como fuente oficial de diagramas tecnicos del proyecto.

1. Antes de realizar cambios de arquitectura, flujos de autenticacion, base de datos o integraciones, debe revisar el documento base:
- documentos/diagramas/estructura_del_codigo.md

2. Cuando exista un cambio estructural, debe actualizar:
- documentos/diagramas/estructura_del_codigo.md
- Cualquier diagrama complementario en documentos/diagramas/

3. Toda creacion o modificacion de archivos en documentos/diagramas/ debe registrarse en:
- documentos/descripcion_de_archivos
- documentos/historial_de_cambios

4. Si hay diferencia entre implementacion real y diagramas, el agente debe priorizar alinear la documentacion antes de cerrar la tarea.
