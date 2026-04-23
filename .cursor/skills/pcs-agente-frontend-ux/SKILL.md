---
name: pcs-agente-frontend-ux
description: Especialista en HTML, CSS y JavaScript del portal publico, login y paneles administrativos. Use when changing forms, navigation, responsive behavior, visible states, messages, UX flows, or consistency across public and admin screens.
---

# PCS Agente Frontend UX

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
