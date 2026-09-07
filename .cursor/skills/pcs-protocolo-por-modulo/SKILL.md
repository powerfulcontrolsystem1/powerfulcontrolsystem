---
name: pcs-protocolo-por-modulo
description: Aplica el protocolo de delegacion y la plantilla de trabajo por modulo del proyecto. Use when a task affects a functional module, authentication, payments, portal publico, reportes, estaciones, paneles administrativos, or any cross-cutting workflow.
---

# PCS Protocolo Por Modulo

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Lectura inicial: contexto_general_del_sistema.md, contexto_codex.md y fuente del módulo según AGENTS.md.
- Los frentes son responsabilidades de análisis. Crear subagentes solo cuando el usuario lo solicita; el perfil no altera autorización, dependencias ni evidencia exigida.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

## Fuentes canonicas

- `.github/agents/protocolo_delegacion.md`
- `.github/agents/plantilla_trabajo_por_modulo.md`

## Flujo minimo

1. Clasificar el modulo afectado.
2. Determinar si el cambio toca backend, frontend, QA o varias capas.
3. Aplicar los frentes como checklist interno. Delegar solo si el usuario pide agentes.
4. Implementar por frente sin romper contratos entre capas.
5. Validar con evidencia tecnica y operativa.
6. Cerrar con trazabilidad documental.

## Regla rapida

- `Rojo`: backend + frontend + QA obligatorios.
- `Amarillo`: al menos dos frentes; el tercero depende de seguridad, impacto visible o runtime.
- `Verde`: puede iniciar con un solo frente, pero si aparece impacto contractual o de datos se escala de inmediato.

## Entregables minimos

- backend: causa tecnica, decision, tablas/rutas/archivos, riesgo residual
- frontend: pantallas, cambio visible, dependencias de API/permisos, riesgo visual
- QA: comandos, resultados, cobertura, huecos o runbook pendiente

## Fuentes y aceptación de la revisión

[AGENTS.md](../../../AGENTS.md), [contexto_general_del_sistema.md](../../../documentos/contexto_general_del_sistema.md), [contexto_codex.md](../../../documentos/contexto_codex.md).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](../../../documentos/requisitos/especificacion_y_trazabilidad.md)).
