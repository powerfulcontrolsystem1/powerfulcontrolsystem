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
  [string]$ProductionBranch = "main",
  [switch]$NoIntegrateWorkBranch,
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
$repoRoot = (Resolve-Path (Join-Path $scriptDir "..")).Path
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
  if ($DryRun -or $PreviewOnly) { return }
  $branch = (& git branch --show-current 2>$null | Select-Object -Last 1).ToString().Trim()
  if ($branch -ne $ProductionBranch) { throw "rs no sincroniza ramas de trabajo al VPS. La revision debe estar integrada en $ProductionBranch." }
  & git fetch origin $ProductionBranch --quiet
  if ($LASTEXITCODE -ne 0) { throw "No se pudo actualizar origin/$ProductionBranch antes del despliegue." }
  $localRevision = (& git rev-parse HEAD 2>$null | Select-Object -Last 1).ToString().Trim()
  $remoteRevision = (& git rev-parse "origin/$ProductionBranch" 2>$null | Select-Object -Last 1).ToString().Trim()
  if ([string]::IsNullOrWhiteSpace($localRevision) -or $localRevision -ne $remoteRevision) { throw "La copia local no coincide exactamente con origin/$ProductionBranch." }
}

function Assert-ReleaseBranchContext {
  if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
    throw "Git no esta disponible para validar la rama de publicacion."
  }
  Set-Location -LiteralPath $repoRoot
  $insideRepo = ((& git rev-parse --is-inside-work-tree 2>$null) | Out-String).Trim()
  if ($LASTEXITCODE -ne 0 -or $insideRepo -ne "true") {
    throw "$repoRoot no es un repositorio Git valido."
  }
  $branch = ((& git branch --show-current 2>$null) | Out-String).Trim()
  if ([string]::IsNullOrWhiteSpace($branch)) {
    throw "rs no publica desde detached HEAD. Cambia a una rama identificable."
  }
  foreach ($stateRef in @("MERGE_HEAD", "REBASE_HEAD", "CHERRY_PICK_HEAD")) {
    & git rev-parse -q --verify $stateRef *> $null
    if ($LASTEXITCODE -eq 0) {
      throw "Hay una operacion Git incompleta ($stateRef). Resuelvela antes de ejecutar rs."
    }
  }
  & git fetch origin $ProductionBranch --quiet
  if ($LASTEXITCODE -ne 0) {
    throw "No se pudo leer origin/$ProductionBranch para preparar la publicacion."
  }
  & git rev-parse --verify "origin/$ProductionBranch" *> $null
  if ($LASTEXITCODE -ne 0) {
    throw "No existe la rama de produccion origin/$ProductionBranch."
  }
  $script:ProductionRevisionBefore = ((& git rev-parse "origin/$ProductionBranch" 2>$null) | Out-String).Trim()
  $script:MigrationBaseRef = "origin/$ProductionBranch"
  if ($branch -ne $ProductionBranch -and $NoIntegrateWorkBranch) {
    throw "La rama actual es $branch. Quita -NoIntegrateWorkBranch para integrarla por PR o cambia a $ProductionBranch."
  }
  if ($branch -eq $ProductionBranch) {
    Write-Host "[OK] Candidato sobre rama de produccion $ProductionBranch." -ForegroundColor Green
  } else {
    $mergeBase = ((& git merge-base HEAD "origin/$ProductionBranch" 2>$null) | Out-String).Trim()
    if ([string]::IsNullOrWhiteSpace($mergeBase) -or $LASTEXITCODE -ne 0) {
      throw "No se pudo resolver el ancestro comun entre $branch y origin/$ProductionBranch."
    }
    $script:MigrationBaseRef = $mergeBase
    $counts = ((& git rev-list --left-right --count "origin/$ProductionBranch...HEAD" 2>$null) | Out-String).Trim()
    Write-Host "[INFO] Rama de trabajo detectada: $branch (diferencia frente a origin/${ProductionBranch}: $counts)." -ForegroundColor Gray
    Write-Host "[INFO] Se publicara la rama y se integrara por PR antes de sincronizar el VPS." -ForegroundColor Gray
  }
  return $branch
}

function Complete-WorkBranchIntegration {
  param([Parameter(Mandatory = $true)][string]$Branch)

  if ($Branch -eq $ProductionBranch) { return }
  if (-not (Get-Command gh -ErrorAction SilentlyContinue)) {
    throw "GitHub CLI no esta disponible para integrar $Branch en $ProductionBranch."
  }
  & gh auth status *> $null
  if ($LASTEXITCODE -ne 0) {
    throw "GitHub CLI no tiene una sesion valida para integrar la rama de trabajo."
  }
  $status = ((& git status --porcelain 2>$null) | Out-String).Trim()
  if (-not [string]::IsNullOrWhiteSpace($status)) {
    throw "La actualizacion dejo cambios sin commit. No se puede crear una PR de publicacion."
  }
  & git fetch origin $Branch $ProductionBranch --quiet
  if ($LASTEXITCODE -ne 0) {
    throw "No se pudieron actualizar las referencias remotas de $Branch y $ProductionBranch."
  }
  $localRevision = ((& git rev-parse HEAD 2>$null) | Out-String).Trim()
  $remoteWorkRevision = ((& git rev-parse "origin/$Branch" 2>$null) | Out-String).Trim()
  if ([string]::IsNullOrWhiteSpace($localRevision) -or $localRevision -ne $remoteWorkRevision) {
    throw "La rama local $Branch no coincide exactamente con origin/$Branch."
  }
  $ahead = [int](((& git rev-list --count "origin/$ProductionBranch..HEAD" 2>$null) | Out-String).Trim())
  if ($ahead -eq 0) {
    Write-Host "[INFO] $Branch no contiene commits pendientes frente a $ProductionBranch; se usara la rama de produccion." -ForegroundColor Gray
  } else {
    $prUrl = ((& gh pr list --base $ProductionBranch --head $Branch --state open --json url --jq '.[0].url' 2>$null) | Out-String).Trim()
    if ([string]::IsNullOrWhiteSpace($prUrl)) {
      $body = @(
        "## Publicacion automatizada por rs",
        "",
        "La rama ``$Branch`` fue validada localmente y debe integrarse en ``$ProductionBranch`` antes del despliegue.",
        "",
        "- Revision candidata: ``$($localRevision.Substring(0, 12))``",
        "- Migraciones e inventarios: gate obligatorio",
        "- VPS: bloqueado hasta confirmar la fusion"
      ) -join "`n"
      $created = @(& gh pr create --base $ProductionBranch --head $Branch --title $Message --body $body 2>&1 | ForEach-Object { $_.ToString().Trim() })
      $createCode = $LASTEXITCODE
      $prUrl = @($created | Where-Object { $_ -match '^https://github\.com/[^/]+/[^/]+/pull/\d+$' } | Select-Object -Last 1)[0]
      if ($createCode -ne 0 -or [string]::IsNullOrWhiteSpace($prUrl)) {
        $created | ForEach-Object { Write-Host $_ }
        throw "No se pudo crear la PR de $Branch hacia $ProductionBranch."
      }
      Write-Host "[OK] PR de publicacion creada: $prUrl" -ForegroundColor Green
    } else {
      Write-Host "[INFO] Reutilizando PR abierta de publicacion: $prUrl" -ForegroundColor Gray
    }

    if (-not $NoAutoMergeProtectedPR) {
      $mergeOutput = @(& gh pr merge $prUrl --auto --squash --delete-branch 2>&1 | ForEach-Object { $_.ToString() })
      if ($LASTEXITCODE -ne 0) {
        $mergeOutput | ForEach-Object { Write-Host $_ }
        Write-Host "[AVISO] GitHub no activo auto-merge; rs esperara una fusion valida sin omitir checks." -ForegroundColor Yellow
      } else {
        Write-Host "[OK] Auto-merge solicitado; GitHub conserva checks y aprobaciones obligatorias." -ForegroundColor Green
      }
    }

    $deadline = (Get-Date).AddSeconds($ProtectedMainPRWaitSeconds)
    do {
      $stateJson = ((& gh pr view $prUrl --json state,mergedAt,mergeStateStatus 2>$null) | Out-String).Trim()
      if ($LASTEXITCODE -eq 0 -and -not [string]::IsNullOrWhiteSpace($stateJson)) {
        $prState = $stateJson | ConvertFrom-Json
        if ($prState.state -eq "MERGED") { break }
        if ($prState.state -eq "CLOSED") { throw "La PR se cerro sin fusionar. El VPS permanece sin cambios." }
      }
      if ((Get-Date) -ge $deadline) {
        throw "La PR sigue pendiente despues de $ProtectedMainPRWaitSeconds segundos: $prUrl. El VPS permanece sin cambios."
      }
      Start-Sleep -Seconds 10
    } while ($true)
    Write-Host "[OK] GitHub confirmo la fusion de la PR." -ForegroundColor Green
  }

  & git switch $ProductionBranch
  if ($LASTEXITCODE -ne 0) {
    throw "La integracion termino, pero no se pudo cambiar a $ProductionBranch."
  }
  & git fetch origin $ProductionBranch --quiet
  if ($LASTEXITCODE -ne 0) {
    throw "No se pudo actualizar origin/$ProductionBranch despues de la fusion."
  }
  & git merge --ff-only "origin/$ProductionBranch"
  if ($LASTEXITCODE -ne 0) {
    throw "$ProductionBranch local no admite fast-forward seguro hacia origin/$ProductionBranch."
  }
  Write-Host "[OK] Rama local $ProductionBranch alineada con la revision fusionada." -ForegroundColor Green
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

$startingBranch = Assert-ReleaseBranchContext

if ($SkipPreflight -and -not ($DryRun -or $PreviewOnly)) {
  throw "Un rs real no permite -SkipPreflight: el gate de migraciones y rama es obligatorio."
}

if (-not $SkipPreflight) {
  $preflightArgs = @{ RequireMigrationAudit = $true; MigrationBaseRef = $script:MigrationBaseRef }
  if ($FullPreflight) { $preflightArgs.Full = $true }
  Invoke-Step -Name "Preflight profesional" -Path $preflightScript -Arguments $preflightArgs
}

if ($DryRun -or $PreviewOnly) {
  Write-Host "[INFO] Actualizar repositorio omitido por DryRun/PreviewOnly."
} else {
  Invoke-Step -Name "Actualizar repositorio" -Path $updateScript -Arguments $updateArgs
  Complete-WorkBranchIntegration -Branch $startingBranch
  if ($startingBranch -ne $ProductionBranch) {
    $postIntegrationArgs = @{ RequireMigrationAudit = $true; MigrationBaseRef = $script:ProductionRevisionBefore }
    Invoke-Step -Name "Preflight post-integracion" -Path $preflightScript -Arguments $postIntegrationArgs
  }
  Assert-ProductionRevision
}
Invoke-Step -Name "Sincronizar VPS" -Path $syncScript -Arguments $syncArgs

Write-Host ""
if ($DryRun -or $PreviewOnly) { Write-Host "[OK] Previsualizacion rs completada sin actualizar repositorio ni VPS." -ForegroundColor Green } else { Write-Host "[OK] Flujo rs completado: repositorio actualizado y VPS sincronizado." -ForegroundColor Green }
