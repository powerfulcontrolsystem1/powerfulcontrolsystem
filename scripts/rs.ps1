<#
.SYNOPSIS
  Ejecuta el flujo rapido: actualizar repositorio y sincronizar al VPS.

.DESCRIPTION
  Orquestador corto para el uso diario. Ejecuta, en orden:
  1. scripts/actualizar_repositorio.ps1
  2. scripts/sync_to_vps.ps1

  Si la actualizacion del repositorio falla, no intenta sincronizar al VPS.
#>

param(
  [string]$Message = "rs: actualizar repositorio y sincronizar VPS",
  [switch]$SkipChangeLog,
  [switch]$SetOrigin,
  [switch]$ForcePush,
  [switch]$DryRun,
  [switch]$PreviewOnly,
  [switch]$SkipPreflight,
  [switch]$FullPreflight,
  [int]$RestartHealthTimeoutSeconds = 900,
  [int]$DockerHealthTimeoutSeconds = 900,
  [bool]$CleanupRemoteUnusedFiles = $true,
  [int]$RemoteCleanupTempMinAgeMinutes = 60,
  [int]$RemoteCleanupDockerBuilderCacheMaxAgeHours = 0
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$scriptDir = $PSScriptRoot
$updateScript = Join-Path $scriptDir "actualizar_repositorio.ps1"
$syncScript = Join-Path $scriptDir "sync_to_vps.ps1"
$preflightScript = Join-Path $scriptDir "profesional_preflight.ps1"
$childPowerShell = if ($PSVersionTable.PSEdition -eq "Core") {
  Join-Path $PSHOME "pwsh.exe"
} else {
  Join-Path $PSHOME "powershell.exe"
}
if (-not (Test-Path -LiteralPath $childPowerShell)) {
  $fallbackShell = Get-Command pwsh, powershell -ErrorAction SilentlyContinue | Select-Object -First 1
  if ($null -eq $fallbackShell) {
    throw "No se encontro un ejecutable PowerShell para ejecutar los pasos de rs"
  }
  $childPowerShell = $fallbackShell.Source
}

if (-not (Test-Path -LiteralPath $updateScript)) {
  throw "No se encontro el script requerido: $updateScript"
}
if (-not (Test-Path -LiteralPath $syncScript)) {
  throw "No se encontro el script requerido: $syncScript"
}
if (-not $SkipPreflight -and -not (Test-Path -LiteralPath $preflightScript)) {
  throw "No se encontro el script requerido: $preflightScript"
}

function Invoke-Step {
  param(
    [Parameter(Mandatory = $true)][string]$Name,
    [Parameter(Mandatory = $true)][string]$Path,
    [hashtable]$Arguments = @{}
  )

  Write-Host ""
  Write-Host "==> $Name" -ForegroundColor Cyan
  # Cada script operativo se ejecuta en un proceso hijo. Varios scripts
  # historicos usan `exit` para devolver su resultado; invocarlos en el mismo
  # proceso cerraba `rs` antes de ejecutar los pasos siguientes.
  $commandParts = @("&", ("'{0}'" -f $Path.Replace("'", "''")))
  foreach ($key in $Arguments.Keys) {
    $value = $Arguments[$key]
    if ($value -is [System.Management.Automation.SwitchParameter]) {
      if ($value.IsPresent) {
        $commandParts += "-$key"
      }
      continue
    }
    if ($value -is [bool]) {
      $boolLiteral = if ($value) { '$true' } else { '$false' }
      $commandParts += "-$key`:$boolLiteral"
      continue
    }
    $commandParts += "-$key"
    $commandParts += ("'{0}'" -f ([string]$value).Replace("'", "''"))
  }
  $childArgs = @("-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", ($commandParts -join " "))
  $global:LASTEXITCODE = 0
  & $childPowerShell @childArgs
  $exitCode = if ($null -ne $LASTEXITCODE) { [int]$LASTEXITCODE } else { 0 }
  if ($exitCode -ne 0) {
    Write-Host "[ERROR] $Name fallo con codigo $exitCode." -ForegroundColor Red
    exit $exitCode
  }
  Write-Host "[OK] $Name completado." -ForegroundColor Green
}

$updateArgs = @{
  Message = $Message
}
if ($SkipChangeLog) { $updateArgs.SkipChangeLog = $true }
if ($SetOrigin) { $updateArgs.SetOrigin = $true }
if ($ForcePush) { $updateArgs.ForcePush = $true }

$syncArgs = @{}
if ($DryRun) { $syncArgs.DryRun = $true }
if ($PreviewOnly) { $syncArgs.PreviewOnly = $true }
$syncArgs.RestartHealthTimeoutSeconds = $RestartHealthTimeoutSeconds
$syncArgs.DockerHealthTimeoutSeconds = $DockerHealthTimeoutSeconds
$syncArgs.CleanupRemoteUnusedFiles = $CleanupRemoteUnusedFiles
$syncArgs.RemoteCleanupTempMinAgeMinutes = $RemoteCleanupTempMinAgeMinutes
$syncArgs.RemoteCleanupDockerBuilderCacheMaxAgeHours = $RemoteCleanupDockerBuilderCacheMaxAgeHours

if (-not $SkipPreflight) {
  $preflightArgs = @{}
  if ($FullPreflight) { $preflightArgs.Full = $true }
  Invoke-Step -Name "Preflight profesional" -Path $preflightScript -Arguments $preflightArgs
}

Invoke-Step -Name "Actualizar repositorio" -Path $updateScript -Arguments $updateArgs
Invoke-Step -Name "Sincronizar VPS" -Path $syncScript -Arguments $syncArgs

Write-Host ""
Write-Host "[OK] Flujo rs completado: repositorio actualizado y VPS sincronizado." -ForegroundColor Green
