# Agentes (Cursor) — Coordinador por defecto: `pcs-agente-go`

Este repositorio ya tiene un sistema de agentes para GitHub Copilot en:

- `copilot-instructions.md`
- `.github/agents/`

Esta carpeta `.cursor/` agrega una **capa nativa para Cursor** sin reemplazar ni borrar lo anterior.

## Coordinador principal

- **Principal por defecto**: `pcs-agente-go`
- Objetivo: clasificar módulo/impacto, activar especialistas cuando aplique y consolidar un cierre único con trazabilidad.

## Especialistas disponibles

- `pcs-agente-backend-db`: backend Go + PostgreSQL (handlers, seguridad, permisos, DB, rendimiento).
- `pcs-agente-frontend-ux`: HTML/CSS/JS y UX (formularios, mensajes, responsive, consistencia visual).
- `pcs-agente-qa-operacion`: pruebas, runtime, túneles, arranque, deploy, verificación end-to-end.
- `pcs-protocolo-por-modulo`: aplica la matriz/plantilla de trabajo por módulo del repo.

## Protocolo operativo (fuentes canónicas)

- Matriz de delegación: `.github/agents/protocolo_delegacion.md`
- Plantilla por módulo: `.github/agents/plantilla_trabajo_por_modulo.md`

## Regla rápida para activar frentes

- **Módulos “rojo”** (`pagos`, `licencias`, `venta_publica`, `estaciones`, `ventas_simple`, `carritos`):
  - backend + frontend + QA obligatorios.
- **Autenticación y permisos**:
  - backend + frontend obligatorios; QA obligatorio si cambia sesión/OAuth/reset/correo/runtime.
- **Cambios solo visuales o textos**:
  - puede iniciar solo frontend, pero si toca contratos o datos, escalar a backend/QA.

## Cierre mínimo esperado

Antes de declarar cerrado:

- causa técnica y archivos/rutas/tablas afectadas
- validación ejecutada (o huecos/riesgos explicitados)
- actualización de documentación/trazabilidad cuando aplique (`documentos/*`)

## Skills y reglas (Cursor)

- Skills: `.cursor/skills/*/SKILL.md`
- (Opcional) Rules: `.cursor/rules/*.mdc` para aplicar gobernanza persistentemente.

