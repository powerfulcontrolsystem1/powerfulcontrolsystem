# Plantilla de trabajo por modulo

Esta plantilla define el ciclo minimo que debe seguir el equipo cuando `agente_go` coordina un cambio por modulo.

## Fase 1. Clasificacion

Responsable: `agente_go`

- Identificar modulo afectado.
- Determinar si el cambio toca backend, frontend, operacion o varias capas.
- Confirmar documentacion obligatoria a revisar antes del cambio.

## Fase 2. Analisis tecnico

Responsables segun impacto:

- `agente_backend_db`: rutas, handlers, modelos, consultas, tablas, seguridad, permisos, concurrencia, trazabilidad por `empresa_id`.
- `agente_frontend_ux`: paginas, componentes, formularios, estados de carga, errores, responsive, coherencia visual.
- `agente_qa_operacion`: pruebas existentes, comandos de validacion, riesgos de runtime, runbooks y dependencias externas.

Entregable:

- resumen breve por frente con hallazgos, riesgos y cambios necesarios.

## Fase 3. Implementacion

Responsables segun frente:

- `agente_backend_db` implementa persistencia, negocio, seguridad y contratos.
- `agente_frontend_ux` implementa interfaz, UX y sincronizacion con backend real.
- `agente_go` arbitra decisiones de arquitectura y evita divergencias entre capas.

## Fase 4. Validacion

Responsable principal: `agente_qa_operacion`

- Ejecutar pruebas dirigidas.
- Ejecutar verificaciones runtime cuando aplique.
- Confirmar que el flujo funciona y que no deja huecos obvios de regresion.

## Fase 5. Trazabilidad y cierre

Responsable final: `agente_go`

- Actualizar documentacion obligatoria del modulo.
- Actualizar diagramas si hubo cambio estructural o de flujo.
- Registrar archivos creados o modificados en la trazabilidad del proyecto.
- Consolidar resultado final, limites y siguientes pasos.

## Checklist por modulo

`agente_go` debe verificar como minimo:

- objetivo del cambio
- modulo afectado
- impacto en `empresa_id`
- impacto en permisos o seguridad
- impacto en frontend
- impacto en runtime/VPS
- pruebas ejecutadas
- documentacion sincronizada

## Modulos que exigen colaboracion de los tres especialistas

- pagos y licencias
- estaciones y ventas por estacion
- carritos con documento de venta
- portal publico con integraciones reales
- paneles administrativos con persistencia o permisos reales

## Ejemplos minimos de aplicacion

### Caso A. Pago aprobado pero correo no enviado

- `agente_go` activa a los tres especialistas.
- `agente_backend_db` revisa persistencia del pago, idempotencia y disparo del correo.
- `agente_frontend_ux` valida el estado visible al usuario tras aprobarse el pago.
- `agente_qa_operacion` ejecuta aprobacion end to end y confirma que el correo se envia una sola vez.

### Caso B. Estacion cobra pero no genera documento

- `agente_go` activa a los tres especialistas.
- `agente_backend_db` corrige cierre, configuracion avanzada y documento transaccional.
- `agente_frontend_ux` verifica feedback de cobro y evidencia visible del documento o error real.
- `agente_qa_operacion` confirma venta, inventario, metricas y documento final.

### Caso C. Login funciona pero el rol entra a una vista indebida

- `agente_go` activa a backend, frontend y QA por impacto de autenticacion y permisos.
- `agente_backend_db` revisa wrappers, middleware y contexto por rol.
- `agente_frontend_ux` revisa menu, redirecciones y enlaces visibles.
- `agente_qa_operacion` valida acceso permitido y denegado en rutas reales.

## Regla de salida por frente

- Ningun frente debe devolver un cierre vacio o ambiguo.
- Cada especialista debe regresar a `agente_go` con evidencia minima suficiente para decidir.

### Evidencia minima esperada

- `agente_backend_db`:
	- causa tecnica concreta
	- decision implementada
	- archivos/rutas/tablas afectadas
	- riesgo residual

- `agente_frontend_ux`:
	- pantallas o flujos afectados
	- cambio visible introducido
	- dependencia de API o permisos
	- riesgo visual u operativo restante

- `agente_qa_operacion`:
	- pruebas o comandos ejecutados
	- resultado observado
	- alcance cubierto
	- hueco de validacion o riesgo residual