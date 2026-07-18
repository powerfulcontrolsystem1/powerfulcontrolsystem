# Runbook staging, CI y E2E visual

## Staging local

```powershell
.\scripts\staging_up.ps1 -ConfigOnly
.\scripts\staging_up.ps1 -Build
```

Abre `http://127.0.0.1:8082`.

## Staging VPS

```bash
bash deploy/scripts/vps-staging-up.sh
```

El override de staging fija `PCS_RUNTIME_SCHEMA_BOOTSTRAP=0` para migrador,
API y worker. Antes de iniciar el smoke, ejecutar el migrador dos veces contra
la misma copia anonimizada y conservar ambos resultados; una segunda corrida no
debe aplicar DDL adicional ni alterar datos.

Configurar Nginx para `staging.powerfulcontrolsystem.com` apuntando a `127.0.0.1:8082` solo despues de validar secretos de `deploy/.env.staging`.

## CI profesional

Workflow: `.github/workflows/professional-ci.yml`.

Valida:

- PowerShell.
- JavaScript.
- Go tests.
- Auditoria profesional.
- Auditoria de seguridad.
- Auditoria permisos/licencias.
- OpenAPI.
- Docker production/staging.
- Observabilidad.

## E2E visual manual

Workflow: `.github/workflows/e2e-visual.yml`.

Secretos requeridos:

- `PCS_QA_EMAIL`
- `PCS_QA_PASSWORD`

Inputs recomendados:

- `base_url`: `https://staging.powerfulcontrolsystem.com`.
- `empresa_id`: empresa de pruebas.
- `max_pages`: `0` para barrido completo.
