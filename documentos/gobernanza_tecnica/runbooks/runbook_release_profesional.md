# Runbook de release profesional

## Orden obligatorio

1. Rama de trabajo actualizada.
2. `.\scripts\release_gate.ps1 -SkipE2E` si no hay credenciales E2E locales.
3. `.\scripts\staging_up.ps1 -Build` o staging VPS activo.
4. Workflow E2E visual contra `https://staging.powerfulcontrolsystem.com`.
5. Backup VPS real.
6. Restauracion temporal del backup.
7. `.\rs.ps1`.
8. Verificacion postdeploy: login, panel super, administrar empresa, licencias, pagos, impresion y errores.

## Compuerta local completa

```powershell
.\scripts\release_gate.ps1
```

Si no se van a ejecutar E2E locales:

```powershell
.\scripts\release_gate.ps1 -SkipE2E
```

## Criterios para produccion

- Preflight completo OK.
- Auditorias de seguridad/permisos/observabilidad OK.
- Backup creado.
- Restauracion temporal OK.
- Staging responde.
- E2E visual sin errores criticos.
- No hay cambios sin documentar.
