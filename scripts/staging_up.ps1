<#
.SYNOPSIS
  Valida o levanta el stack Docker de staging.

.DESCRIPTION
  Usa deploy/docker-compose.platform.yml con el override de staging para
  ejecutar la plataforma aislada en puerto 8082 y volumenes separados.
#>

param(
  [switch]$ConfigOnly,
  [switch]$Build,
  [switch]$Down,
  [string]$EnvFile = "deploy\.env.staging.example"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
Push-Location $repoRoot
try {
  if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    throw "Docker no esta disponible en este equipo."
  }
  $composeArgs = @(
    "compose",
    "--env-file", $EnvFile,
    "-f", "deploy/docker-compose.platform.yml",
    "-f", "deploy/docker-compose.staging.yml"
  )

  if ($ConfigOnly) {
    & docker @composeArgs config --quiet
    exit $LASTEXITCODE
  }

  if ($Down) {
    & docker @composeArgs down
    exit $LASTEXITCODE
  }

  $upArgs = @("up", "-d")
  if ($Build) { $upArgs += "--build" }
  & docker @composeArgs @upArgs
  if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

  Write-Host "[OK] Staging listo en http://127.0.0.1:8082" -ForegroundColor Green
} finally {
  Pop-Location
}
