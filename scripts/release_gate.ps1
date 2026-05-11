<#
.SYNOPSIS
  Ejecuta la compuerta profesional previa a release.
#>

param(
  [switch]$SkipRemoteBackup,
  [switch]$SkipRestoreDrill,
  [switch]$SkipE2E,
  [switch]$SkipLoadSmoke,
  [string]$E2EBaseUrl = "https://staging.powerfulcontrolsystem.com",
  [string]$EmpresaId = "7"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
Push-Location $repoRoot
try {
  & .\scripts\profesional_preflight.ps1 -Full
  if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

  if (-not $SkipRemoteBackup) {
    & .\scripts\vps_backup_operacion.ps1
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    if ($SkipRestoreDrill) {
      & .\scripts\vps_restore_validation.ps1
    } else {
      & .\scripts\vps_restore_validation.ps1 -ExecuteDrill
    }
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
  }

  if (-not $SkipE2E) {
    if (-not $env:PCS_QA_EMAIL -or -not $env:PCS_QA_PASSWORD) {
      throw "Para E2E define PCS_QA_EMAIL y PCS_QA_PASSWORD, o usa -SkipE2E."
    }
    $env:PCS_QA_BASE_URL = $E2EBaseUrl
    $env:PCS_QA_EMPRESA_ID = $EmpresaId
    $env:PCS_QA_VIEWPORTS = "desktop,mobile"
    & node tools\qa_e2e_buttons.cjs
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    & node tools\qa_print_formats.cjs
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
  }

  if (-not $SkipLoadSmoke) {
    $env:PCS_LOAD_BASE_URL = $E2EBaseUrl
    & node tools\load_smoke_test.mjs
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
  }

  & node tools\release_manifest.mjs
  if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

  Write-Host "[OK] Release gate completado." -ForegroundColor Green
} finally {
  Pop-Location
}
