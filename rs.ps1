<#
.SYNOPSIS
  Acceso corto al flujo scripts/rs.ps1 desde la raiz del proyecto.
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

$repoRoot = $PSScriptRoot
$target = Join-Path $repoRoot "scripts\rs.ps1"

if (-not (Test-Path -LiteralPath $target)) {
  throw "No se encontro el script requerido: $target"
}

$forward = @{
  Message = $Message
}
if ($SkipChangeLog) { $forward.SkipChangeLog = $true }
if ($SetOrigin) { $forward.SetOrigin = $true }
if ($ForcePush) { $forward.ForcePush = $true }
if ($DryRun) { $forward.DryRun = $true }
if ($PreviewOnly) { $forward.PreviewOnly = $true }
if ($SkipPreflight) { $forward.SkipPreflight = $true }
if ($FullPreflight) { $forward.FullPreflight = $true }
$forward.RestartHealthTimeoutSeconds = $RestartHealthTimeoutSeconds
$forward.DockerHealthTimeoutSeconds = $DockerHealthTimeoutSeconds
$forward.CleanupRemoteUnusedFiles = $CleanupRemoteUnusedFiles
$forward.RemoteCleanupTempMinAgeMinutes = $RemoteCleanupTempMinAgeMinutes
$forward.RemoteCleanupDockerBuilderCacheMaxAgeHours = $RemoteCleanupDockerBuilderCacheMaxAgeHours

& $target @forward
exit $LASTEXITCODE
