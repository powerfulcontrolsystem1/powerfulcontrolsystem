---
name: pcs-protocolo-por-modulo
description: Aplica el protocolo de delegacion y la plantilla de trabajo por modulo del proyecto. Use when a task affects a functional module, authentication, payments, portal publico, reportes, estaciones, paneles administrativos, or any cross-cutting workflow.
---

# PCS Protocolo Por Modulo

## Fuentes canonicas

- `.github/agents/protocolo_delegacion.md`
- `.github/agents/plantilla_trabajo_por_modulo.md`

## Flujo minimo

1. Clasificar el modulo afectado.
2. Determinar si el cambio toca backend, frontend, QA o varias capas.
3. Activar los frentes obligatorios segun la matriz del protocolo.
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
