# Endurecimiento de despliegue - 2026-07-15

## Controles incorporados

- Staging usa worker, contenedores, volumenes y almacenamiento privado propios.
- `rs.ps1` y `sync_to_vps.ps1` solo publican un `HEAD` que coincide con
  `origin/main`; una rama de trabajo debe integrarse mediante la proteccion de
  GitHub.
- Las vistas previas no actualizan Git ni contactan el VPS. En Docker, el
  bootstrap remoto queda desactivado para conservar secretos en el VPS.
- Compose ejecuta `--remove-orphans` dentro del proyecto PCS, evitando workers
  antiguos que compitan por la cola.
- CI ejecuta `tools/deploy_pipeline_contract.mjs` para detectar regresiones de
  estos controles.

## Evidencia local

- `go test ./...` y `go vet ./...` desde `backend`.
- `flutter analyze` y `flutter test` sobre la aplicacion movil cuando exista
  en la rama de trabajo correspondiente.
- Validacion de sintaxis PowerShell y contratos estaticos de staging/despliegue.

## Gates antes de trafico comercial

No sustituir con mocks: restauracion de backup en staging anonimizado, prueba
concurrente de venta/caja, DIAN, Wompi/Epayco/Bre-B, Mailu/WhatsApp,
Nextcloud/OnlyOffice, y una cuenta de QA vigente para la validacion visual.
Android e iPhone requieren sus firmas reales en los entornos protegidos del
proveedor CI. Los hallazgos se registran antes de habilitar trafico comercial.
