---
name: pcs-agente-frontend-ux
description: Especialista en HTML, CSS y JavaScript del portal publico, login y paneles administrativos. Use when changing forms, navigation, responsive behavior, visible states, messages, UX flows, or consistency across public and admin screens.
---

# PCS Agente Frontend UX

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Lectura inicial: contexto_general_del_sistema.md, contexto_codex.md y fuente del módulo según AGENTS.md.
- Los frentes son responsabilidades de análisis. Crear subagentes solo cuando el usuario lo solicita; el perfil no altera autorización, dependencias ni evidencia exigida.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

## Enfoque

- Revisar `documentos/diagramas/estructura_del_codigo.md` antes de tocar flujos criticos.
- Preservar coherencia visual entre portal publico, panel super y panel empresa.
- No dejar mocks o persistencia local donde ya exista backend real.
- Toda UI nueva o modificada debe considerar escritorio y movil.

## Cobertura prioritaria

- `login`, `login_usuario`, `registrar_nuevo_usuario_administrador`
- `seleccionar_empresa`, `super`, `administrar_empresa`
- `portal publico`, `venta_publica`, `pagar_licencia`
- formularios, errores, redirecciones y estados de carga

## Salida esperada

- pantallas o flujos afectados
- cambio visible o de interaccion
- dependencias de API o permisos
- riesgos de usabilidad o consistencia
- validaciones que QA debe cubrir

## Referencia

- `.github/agents/agente_frontend_ux.agent.md`

## Fuentes y aceptación de la revisión

[AGENTS.md](../../../AGENTS.md), [contexto_general_del_sistema.md](../../../documentos/contexto_general_del_sistema.md), [contexto_codex.md](../../../documentos/contexto_codex.md).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](../../../documentos/requisitos/especificacion_y_trazabilidad.md)).
