<#
.SYNOPSIS
  Encadena actualizar_repositorio.ps1 (push a origin) y sync_to_vps.ps1 (despliegue al VPS).

.DESCRIPTION
  Carga scripts/pcs_deployment.local.ps1 si existe (misma config que al ejecutar los scripts por separado).
  -SkipGit: solo VPS. -SkipVps: solo Git. Parametros comunes hacia el script de actualizacion de repo.
#>
param(
  [string]$Message = "Publicacion: git y VPS",
  [switch]$SkipChangeLog,
  [switch]$SetOrigin,
  [switch]$ForcePush,
  [switch]$SkipGit,
  [switch]$SkipVps
)

$ErrorActionPreference = "Stop"
$here = $PSScriptRoot

$pub = Join-Path $here "actualizar_repositorio.ps1"
$sync = Join-Path $here "sync_to_vps.ps1"

if (-not (Test-Path -LiteralPath $pub)) { throw "No se encuentra: $pub" }
if (-not (Test-Path -LiteralPath $sync)) { throw "No se encuentra: $sync" }

if (-not $SkipGit) {
  & $pub -Message $Message -SkipChangeLog:$SkipChangeLog -SetOrigin:$SetOrigin -ForcePush:$ForcePush
  if ($null -ne $LASTEXITCODE -and $LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}
if (-not $SkipVps) {
  & $sync
  if ($null -ne $LASTEXITCODE -and $LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

Write-Host "[OK] publicar_git_y_vps: completado." -ForegroundColor Green
exit 0
