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

Para desplegar un candidato P108 sin cambiar el checkout que sirve producción,
usar `deploy/scripts/vps-staging-candidate-up.sh`. Requiere explícitamente
`CANDIDATE_REF`, `CANDIDATE_SHA`, un `WORKTREE_DIR` distinto al proyecto
productivo y el archivo staging existente. El script rechaza `RESET_STAGING` y
no crea ni imprime secretos. Si el paquete productivo del VPS no conserva
`.git`, clona solo el candidato en el directorio aislado y mantiene el archivo
de entorno staging existente fuera de ese clon. Ejecutar solo después de CI verde y conservar SHA,
salida de migración, health, ready y resultados E2E.

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
