# ADR-0001: `empresa_id` como frontera obligatoria multiempresa

Fecha: 2026-04-18
Estado: aceptada

## Contexto

El sistema es multiempresa y expone modulos operativos, publicos y administrativos que comparten backend, frontend e integraciones. El mayor riesgo funcional transversal es mezclar datos, permisos o efectos persistentes entre empresas.

## Decision

Se adopta `empresa_id` como frontera obligatoria de aislamiento funcional para toda operacion empresarial.

Esto implica:

- todas las rutas bajo `/api/empresa` deben validar alcance por `empresa_id` mediante wrappers o middleware equivalentes.
- todo flujo publico que termine afectando una empresa especifica debe transportar y validar contexto esperado de `empresa_id` antes de persistir o confirmar estados.
- toda documentacion tecnica de un flujo multiempresa debe declarar explicitamente donde entra, donde se valida y donde se persiste `empresa_id`.

## Consecuencias

### Positivas

- reduce cruces de datos entre empresas.
- alinea permisos, frontend, pagos, estaciones y documentos sobre una misma frontera tecnica.
- facilita pruebas de regresion con enfoque de aislamiento.

### Costos

- obliga a documentar y validar mas parametros en rutas publicas o intermodulares.
- hace visibles fallos legacy donde antes el sistema asumía contexto implicito.

## Aplicacion inmediata

Este ADR aplica como minimo a:

- pagos y licencias
- estaciones, carritos y sensores
- venta publica por empresa
- usuarios y permisos por empresa
- facturacion y documentos transaccionales
