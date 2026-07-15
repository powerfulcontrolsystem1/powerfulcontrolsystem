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
  [switch]$BuildAndroid,
  [switch]$BuildIOS,
  [switch]$BuildMobile,
  [switch]$SkipMobile,
  [switch]$MobileDebug,
  [switch]$MobileRelease,
  [switch]$TriggerIOSWorkflow,
  [int]$ProtectedMainPRWaitSeconds = 900,
  [switch]$NoAutoMergeProtectedPR,
  [int]$RestartHealthTimeoutSeconds = 900,
  [int]$DockerHealthTimeoutSeconds = 900,
  [int]$StepTimeoutSeconds = 3600,
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
$androidBuildScript = Join-Path $scriptDir "generar_aplicacion_android.ps1"
$iosBuildScript = Join-Path $scriptDir "generar_aplicacion_ios.ps1"
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
if ($BuildMobile) { $BuildAndroid = $true; $BuildIOS = $true }
if ($SkipMobile -and ($BuildAndroid -or $BuildIOS)) { throw "SkipMobile no se puede combinar con BuildAndroid, BuildIOS o BuildMobile." }
if ($MobileDebug -and $MobileRelease) { throw "Usa solo MobileDebug o MobileRelease." }
if (($BuildAndroid -or $BuildIOS) -and -not $SkipMobile) {
  if ($BuildAndroid -and -not (Test-Path -LiteralPath $androidBuildScript)) { throw "No se encontro el script Android: $androidBuildScript" }
  if ($BuildIOS -and -not (Test-Path -LiteralPath $iosBuildScript)) { throw "No se encontro el script iOS: $iosBuildScript" }
}

function Invoke-Step {
  param(
    [Parameter(Mandatory = $true)][string]$Name,
    [Parameter(Mandatory = $true)][string]$Path,
    [hashtable]$Arguments = @{}
  )

  if ($StepTimeoutSeconds -lt 60) {
    throw "StepTimeoutSeconds debe ser de al menos 60 segundos"
  }
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
  $safeStepName = ($Name -replace '[^A-Za-z0-9_-]', '_')
  $stepStamp = Get-Date -Format 'yyyyMMdd-HHmmss'
  $stepLogDir = Join-Path $scriptDir 'logs'
  New-Item -ItemType Directory -Force -Path $stepLogDir | Out-Null
  $stdoutPath = Join-Path $stepLogDir ("rs-{0}-{1}.out.log" -f $stepStamp, $safeStepName)
  $stderrPath = Join-Path $stepLogDir ("rs-{0}-{1}.err.log" -f $stepStamp, $safeStepName)
  Write-Host ("[INFO] Iniciando paso aislado; salida: {0}" -f $stdoutPath) -ForegroundColor DarkGray
  $process = Start-Process -FilePath $childPowerShell -ArgumentList $childArgs -NoNewWindow -PassThru -RedirectStandardOutput $stdoutPath -RedirectStandardError $stderrPath
  $deadline = (Get-Date).AddSeconds($StepTimeoutSeconds)
  while (-not $process.HasExited -and (Get-Date) -lt $deadline) {
    Start-Sleep -Seconds 1
    $process.Refresh()
  }
  if (-not $process.HasExited) {
    try { $process.Kill($true) } catch { }
    throw ("{0} supero el limite de {1} segundos. Revisa {2} y {3}." -f $Name, $StepTimeoutSeconds, $stdoutPath, $stderrPath)
  }
  $exitCode = [int]$process.ExitCode
  foreach ($logPath in @($stdoutPath, $stderrPath)) {
    if (Test-Path -LiteralPath $logPath) {
      Get-Content -LiteralPath $logPath | ForEach-Object { Write-Host $_ }
    }
  }
  if ($exitCode -ne 0) {
    Write-Host "[ERROR] $Name fallo con codigo $exitCode. Revisa $stdoutPath y $stderrPath." -ForegroundColor Red
    exit $exitCode
  }
  Write-Host "[OK] $Name completado." -ForegroundColor Green
}

function Assert-ProductionRevision {
  if ($DryRun -or $PreviewOnly) {
    Write-Host "[INFO] Verificacion de rama productiva omitida por DryRun/PreviewOnly."
    return
  }

  $branch = (& git branch --show-current 2>$null | Select-Object -Last 1).ToString().Trim()
  if ($branch -ne "main") {
    throw "rs no sincroniza ramas de trabajo al VPS. Integra la revision aprobada en main y usa staging para validar $branch."
  }

  & git fetch origin main --quiet
  if ($LASTEXITCODE -ne 0) {
    throw "No se pudo actualizar origin/main antes del despliegue."
  }
  $localRevision = (& git rev-parse HEAD 2>$null | Select-Object -Last 1).ToString().Trim()
  $remoteRevision = (& git rev-parse origin/main 2>$null | Select-Object -Last 1).ToString().Trim()
  if ([string]::IsNullOrWhiteSpace($localRevision) -or [string]::IsNullOrWhiteSpace($remoteRevision) -or $localRevision -ne $remoteRevision) {
    throw "La copia local no coincide exactamente con origin/main. Actualiza main antes de sincronizar el VPS."
  }
  Write-Host "[OK] Revision productiva confirmada en origin/main." -ForegroundColor Green
}

$updateArgs = @{
  Message = $Message
  ProtectedMainPRWaitSeconds = $ProtectedMainPRWaitSeconds
}
if ($SkipChangeLog) { $updateArgs.SkipChangeLog = $true }
if ($SetOrigin) { $updateArgs.SetOrigin = $true }
if ($ForcePush) { $updateArgs.ForcePush = $true }
if ($NoAutoMergeProtectedPR) { $updateArgs.NoAutoMergeProtectedPR = $true }

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

if ($DryRun -or $PreviewOnly) {
  Write-Host "[INFO] Actualizar repositorio omitido por DryRun/PreviewOnly."
} else {
  Invoke-Step -Name "Actualizar repositorio" -Path $updateScript -Arguments $updateArgs
  Assert-ProductionRevision
}
if ($BuildAndroid -and -not $SkipMobile) {
  $androidArgs = @{}
  if ($DryRun) { $androidArgs.DryRun = $true }
  if ($MobileDebug) { $androidArgs.Debug = $true } else { $androidArgs.Release = $true }
  Invoke-Step -Name "Generar aplicacion Android" -Path $androidBuildScript -Arguments $androidArgs
}
if ($BuildIOS -and -not $SkipMobile) {
  $iosArgs = @{}
  if ($DryRun) { $iosArgs.DryRun = $true }
  if ($TriggerIOSWorkflow) { $iosArgs.TriggerIOSWorkflow = $true }
  Invoke-Step -Name "Validar o generar aplicacion iPhone" -Path $iosBuildScript -Arguments $iosArgs
}
Invoke-Step -Name "Sincronizar VPS" -Path $syncScript -Arguments $syncArgs

Write-Host ""
if ($DryRun -or $PreviewOnly) {
  Write-Host "[OK] Previsualizacion rs completada sin actualizar repositorio ni VPS." -ForegroundColor Green
} else {
  Write-Host "[OK] Flujo rs completado: repositorio actualizado y VPS sincronizado." -ForegroundColor Green
}
