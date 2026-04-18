# Protocolo de delegacion del equipo de agentes

## Regla base

- Toda tarea entra por `agente_go`.
- `agente_go` clasifica el tipo de impacto antes de asignar trabajo.
- Ningun agente especialista cierra por separado una tarea que afecte mas de una capa.

## Matriz de decision de `agente_go`

### 1. Tarea solo backend o base de datos

- Delegar primero a `agente_backend_db`.
- Pedir a `agente_qa_operacion` validacion si hay riesgo en runtime, migraciones, pagos, permisos o datos.
- Mantener a `agente_frontend_ux` fuera salvo que cambien payloads, rutas, mensajes o comportamiento visible.

### 2. Tarea solo frontend o UX

- Delegar primero a `agente_frontend_ux`.
- Involucrar a `agente_backend_db` si el frontend requiere endpoint nuevo, cambio contractual o persistencia real.
- Pedir a `agente_qa_operacion` verificacion responsive o end to end cuando el flujo sea operativo.

### 3. Tarea de validacion, despliegue o incidente

- Delegar primero a `agente_qa_operacion`.
- Involucrar a `agente_backend_db` si el incidente apunta a consultas, handlers, autenticacion, pagos o esquemas.
- Involucrar a `agente_frontend_ux` si el fallo solo se reproduce en navegacion, formularios o responsive.

### 4. Tarea transversal de modulo funcional

- `agente_go` divide la tarea en tres frentes cuando afecte API, interfaz y operacion.
- `agente_backend_db` resuelve contratos, reglas de negocio, datos y seguridad.
- `agente_frontend_ux` resuelve interaccion, estados de UI, mensajes, responsive y consistencia visual.
- `agente_qa_operacion` valida regresiones, arranque, flujo real, runbook y riesgos residuales.
- `agente_go` integra la salida final y decide si la tarea queda cerrada.

## Criterios de activacion por modulo

- `pagos y licencias`: activar a los tres especialistas.
- `facturacion electronica y DIAN`: activar a `agente_backend_db` y `agente_qa_operacion`; sumar `agente_frontend_ux` si cambia panel o flujo visible.
- `estaciones`, `ventas_simple`, `carritos`: activar a los tres especialistas.
- `reportes`, `exportaciones`, `interoperabilidad contable`: activar a `agente_backend_db` y `agente_qa_operacion`; sumar `agente_frontend_ux` si cambia la experiencia de consulta/exporte.
- `autenticacion`, `permisos`, `portal publico`, `paneles administrativos`: activar backend y frontend; sumar QA cuando cambien sesiones, OAuth, correo, reset o acceso por rol.

## Tabla rapida por modulo

| Modulo | Backend DB | Frontend UX | QA Operacion | Regla |
| --- | --- | --- | --- | --- |
| `pagos` | obligatorio | obligatorio | obligatorio | cierre conjunto obligatorio |
| `licencias` | obligatorio | obligatorio | obligatorio | cierre conjunto obligatorio |
| `venta_publica` | obligatorio | obligatorio | obligatorio | cierre conjunto obligatorio |
| `estaciones` | obligatorio | obligatorio | obligatorio | cierre conjunto obligatorio |
| `ventas_simple` | obligatorio | obligatorio | obligatorio | cierre conjunto obligatorio |
| `carritos` | obligatorio | obligatorio | obligatorio | cierre conjunto obligatorio |
| `facturacion electronica` | obligatorio | condicional | obligatorio | frontend entra si cambia panel o flujo visible |
| `DIAN` | obligatorio | condicional | obligatorio | frontend entra si cambia panel o evidencia visible |
| `reportes` | obligatorio | condicional | obligatorio | frontend entra si cambia filtros, tablas o exportes visibles |
| `autenticacion` | obligatorio | obligatorio | obligatorio | obligatorio si cambia sesion, OAuth, reset o permisos |
| `portal publico` | condicional | obligatorio | condicional | backend/QA entran si hay integracion real o rutas publicas |
| `super` | obligatorio | obligatorio | condicional | QA entra si cambia runtime, permisos o flujo critico |
| `administrar_empresa` | obligatorio | obligatorio | condicional | QA entra si cambia persistencia, permisos o flujo operativo |
| `vpssecurity` | obligatorio | condicional | obligatorio | frontend entra solo si hay impacto visible |
| `deploy`, `tuneles`, `arranque` | obligatorio | no | obligatorio | cierre tecnico con evidencia runtime |

## Semaforo ejecutivo por modulo

- `Rojo`:
	- `pagos`
	- `licencias`
	- `venta_publica`
	- `estaciones`
	- `ventas_simple`
	- `carritos`
	- Regla: `agente_go` debe activar a backend, frontend y QA sin excepcion ordinaria.

- `Amarillo`:
	- `autenticacion`
	- `permisos`
	- `facturacion electronica`
	- `DIAN`
	- `reportes`
	- `administrar_empresa`
	- `super`
	- Regla: siempre entran al menos dos frentes; `agente_go` decide el tercero segun impacto visible, runtime o seguridad.

- `Verde`:
	- `portal publico` sin integracion real
	- `contenido visual`
	- `ajustes de estilo`
	- `textos`
	- Regla: puede arrancar con un solo especialista, pero si aparece impacto contractual, de datos o runtime se escala inmediatamente.

## Ejemplos reales de delegacion

### Ejemplo 1. Ajuste en checkout de licencias con Epayco

- `agente_go` clasifica el caso como `pagos y licencias`.
- Activa a `agente_backend_db` para checkout, webhook, conciliacion, correo y persistencia.
- Activa a `agente_frontend_ux` para `pagar_licencia.html`, estados visibles, retorno y mensajes de error.
- Activa a `agente_qa_operacion` para validar aprobacion, rechazo, retorno web, polling y correo final.
- `agente_go` solo cierra cuando los tres frentes entregan evidencia compatible.

### Ejemplo 2. Cambio en estaciones con documento de venta

- `agente_go` clasifica el caso como `estaciones`, `ventas_simple` y `carritos`.
- `agente_backend_db` corrige cierre, metricas, inventario y emision documental.
- `agente_frontend_ux` ajusta estados de estacion, flujo de cobro y feedback operativo.
- `agente_qa_operacion` valida la venta real, el descuento de inventario, el documento emitido y el arranque runtime.
- Sin evidencia de inventario y documento, `agente_go` no puede cerrar la tarea.

### Ejemplo 3. Cambio en login y permisos empresariales

- `agente_go` clasifica el caso como `autenticacion` y `permisos`.
- `agente_backend_db` trabaja sesiones, wrappers, roles, alcance por `empresa_id` y rutas publicas/protegidas.
- `agente_frontend_ux` trabaja formularios, mensajes de error, redirecciones y experiencia de primer ingreso o reset.
- `agente_qa_operacion` valida login, reset, primer password, rol efectivo y acceso a paginas protegidas.
- `agente_go` exige backend, frontend y QA porque aqui hay impacto funcional completo.

### Ejemplo 4. Caso completo extremo a extremo: venta por estacion con pago y documento

#### Solicitud

- El usuario reporta que una venta en `estacion 4` cobra bien, pero a veces no genera documento y la interfaz no deja claro el estado final.

#### Paso 1. Clasificacion de `agente_go`

- Modulos: `estaciones`, `ventas_simple`, `carritos`, `documentos transaccionales`.
- Nivel semaforo: `Rojo`.
- Decision: activar obligatoriamente a `agente_backend_db`, `agente_frontend_ux` y `agente_qa_operacion`.

#### Paso 2. Encargo por frente

- `agente_backend_db`:
	- revisar `pagar_estacion`, configuracion avanzada, documento de venta y persistencia de metricas.
	- identificar causa tecnica y proponer correccion.
- `agente_frontend_ux`:
	- revisar feedback de cobro, mensaje final, estado visible de la estacion y evidencia del documento generado o del error real.
	- proponer ajuste de experiencia si el usuario queda sin confirmacion clara.
- `agente_qa_operacion`:
	- reproducir la venta de punta a punta.
	- validar inventario, documento, metricas, arranque backend y comportamiento real en runtime.

#### Paso 3. Integracion de `agente_go`

- reunir causa tecnica, impacto visible y evidencia de validacion.
- confirmar que documentacion, diagramas y trazabilidad reflejan el cambio.
- rechazar el cierre si falta alguno de estos puntos:
	- evidencia de inventario o documento
	- evidencia de UI final
	- evidencia de prueba o runtime

#### Paso 4. Cierre valido

- `agente_go` solo cierra cuando los tres frentes devuelven evidencia compatible y el modulo queda coherente en backend, UX y operacion.

## Entregables minimos al volver a `agente_go`

- `agente_backend_db`: causa tecnica, archivos tocados, riesgos de datos/seguridad, pruebas necesarias.
- `agente_frontend_ux`: pantallas afectadas, dependencias de API, impacto visual/operativo, estados no cubiertos.
- `agente_qa_operacion`: pruebas ejecutadas, resultado, evidencia, huecos de verificacion y runbook pendiente.

## Regla de cierre

- Solo `agente_go` puede declarar una tarea integrada como completada.
- El cierre exige consistencia entre codigo, pruebas, documentacion, diagramas y trazabilidad.
- En modulos criticos no se acepta cierre parcial con solo un especialista activo si la tabla rapida marca participacion obligatoria multiple.
- Si el semaforo del modulo es `Rojo`, cualquier intento de cierre sin backend, frontend y QA debe considerarse invalido salvo justificacion excepcional explicita.