# Equipo de agentes del repositorio

Este repositorio opera con un equipo base de cuatro agentes coordinados.

- `agente_go`: agente principal, seleccionado por defecto para el trabajo diario y responsable de dirigir al resto del equipo.
- `agente_backend_db`: agente especialista en backend Go, PostgreSQL, modelos, handlers, seguridad, migraciones y rendimiento.
- `agente_frontend_ux`: agente especialista en interfaces HTML/CSS/JavaScript, experiencia operativa y consistencia visual.
- `agente_qa_operacion`: agente especialista en pruebas, validacion end to end, runbooks, despliegue, incidentes y verificacion operativa.

## Regla de coordinacion

- Toda tarea entra primero por `agente_go`.
- `agente_go` decide si resuelve directamente o si divide el trabajo entre especialistas.
- Los agentes especialistas no redefinen arquitectura por su cuenta; devuelven hallazgos, cambios o validaciones a `agente_go`.
- `agente_go` integra resultados, valida conflictos entre areas y cierra con trazabilidad documental completa.

## Regla de mejora continua

- Si una tarea toca backend, frontend y operacion, `agente_go` debe pedir colaboracion cruzada y consolidar una salida unica.
- Si una tarea solo afecta un area, `agente_go` puede delegar la ejecucion tecnica y luego validar documentacion, pruebas y riesgos.
- Toda mejora debe respetar `empresa_id`, trazabilidad operativa, documentacion vigente y reglas de exportacion/interoperabilidad del proyecto.

## Protocolo operativo

- El protocolo detallado de delegacion vive en `.github/agents/protocolo_delegacion.md`.
- La plantilla base de ejecucion por modulo vive en `.github/agents/plantilla_trabajo_por_modulo.md`.
- `agente_go` debe usar ambos documentos para repartir trabajo y exigir la misma disciplina de salida a todos los especialistas.

## Uso rapido

- Si necesitas decidir rapido a quien activar, usa primero la tabla rapida por modulo del protocolo.
- Si necesitas decidir todavia mas rapido, usa el semaforo ejecutivo del protocolo para saber si el modulo exige uno, dos o tres frentes.
- Si el cambio es critico en `pagos`, `licencias`, `venta_publica`, `estaciones`, `ventas_simple` o `carritos`, `agente_go` debe activar a los tres especialistas sin excepcion ordinaria.
- Si dudas sobre el flujo, usa los ejemplos reales del protocolo y la plantilla por modulo antes de cerrar la tarea.
- Si un especialista no devuelve evidencia minima suficiente, `agente_go` no debe aceptar ese frente como cerrado.