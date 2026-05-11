<#
.SYNOPSIS
  Validacion profesional previa a commit, sync o despliegue.

.DESCRIPTION
  Ejecuta chequeos de sintaxis, auditoria de permisos/licencias/modulos,
  validacion Docker Compose y, opcionalmente, pruebas Go.
#>

param(
  [switch]$Full,
  [switch]$SkipGoTests,
  [switch]$SkipDockerConfig,
  [switch]$SkipAudit,
  [switch]$Strict,
  [string]$ReportDir = "documentos\reportes_profesionales"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$reportRoot = Join-Path $repoRoot $ReportDir
if (-not (Test-Path -LiteralPath $reportRoot)) {
  New-Item -ItemType Directory -Path $reportRoot -Force | Out-Null
}

$stamp = Get-Date -Format "yyyyMMdd_HHmmss"
$reportPath = Join-Path $reportRoot "preflight_$stamp.md"
$status = "OK"
$sections = New-Object System.Collections.Generic.List[string]
$fence = '```'

function Add-Section {
  param([string]$Title, [string]$Body)
  $script:sections.Add("## $Title`n$Body`n") | Out-Null
}

function Invoke-Captured {
  param(
    [Parameter(Mandatory=$true)][string]$Title,
    [Parameter(Mandatory=$true)][scriptblock]$Script,
    [switch]$Required
  )
  Write-Host "`n==> $Title" -ForegroundColor Cyan
  $output = @()
  $exitCode = 0
  try {
    $global:LASTEXITCODE = 0
    $output = & $Script 2>&1 | ForEach-Object { $_.ToString() }
    $exitCode = 0
  } catch {
    $output += $_.Exception.Message
    $exitCode = 1
  }
  $text = ($output -join "`n").Trim()
  if ($exitCode -ne 0) {
    if ($Required) { $script:status = "ERROR" } else { $script:status = "WARN" }
    Write-Host "[ERROR] $Title fallo con codigo $exitCode" -ForegroundColor Red
  } else {
    Write-Host "[OK] $Title" -ForegroundColor Green
  }
  Add-Section -Title "$Title (exit=$exitCode)" -Body ($fence + "text`n" + $text + "`n" + $fence)
  if ($Required -and $exitCode -ne 0) {
    throw "$Title fallo con codigo $exitCode"
  }
}

Push-Location $repoRoot
try {
  Invoke-Captured -Title "Parseo PowerShell de scripts operativos" -Required -Script {
    $files = @(
      (Join-Path "scripts" "sync_to_vps.ps1"),
      (Join-Path "scripts" "rs.ps1"),
      "rs.ps1",
      (Join-Path "scripts" "actualizar_repositorio.ps1"),
      (Join-Path "scripts" "profesional_preflight.ps1"),
      (Join-Path "scripts" "vps_backup_operacion.ps1"),
      (Join-Path "scripts" "vps_restore_validation.ps1"),
      (Join-Path "scripts" "staging_up.ps1")
    )
    foreach ($file in $files) {
      $tokens = $null
      $errors = $null
      [System.Management.Automation.Language.Parser]::ParseFile((Resolve-Path $file).Path, [ref]$tokens, [ref]$errors) | Out-Null
      if ($errors -and $errors.Count) {
        throw "$file parse errors: " + (($errors | ForEach-Object { $_.Message }) -join "; ")
      }
      "$file parse ok"
    }
  }

  Invoke-Captured -Title "Sintaxis JavaScript frontend" -Required -Script {
    $files = Get-ChildItem -Path (Join-Path "web" "js") -Filter "*.js" -File -Recurse | Sort-Object FullName
    foreach ($file in $files) {
      & node --check $file.FullName
      if ($LASTEXITCODE -ne 0) { throw "node --check fallo en $($file.FullName)" }
    }
    "JS files checked: $($files.Count)"
  }

  if (-not $SkipAudit) {
    Invoke-Captured -Title "Auditoria profesional de modulos, permisos y portal" -Required:$Strict -Script {
      & node tools\professional_audit.mjs --out $ReportDir
      $code = if ($null -ne $LASTEXITCODE) { [int]$LASTEXITCODE } else { 0 }
      if ($Strict -and $code -ne 0) { throw "auditoria profesional fallo con codigo $code" }
      "Auditoria profesional finalizo con codigo $code"
    }

    Invoke-Captured -Title "Auditoria de seguridad" -Required:$Strict -Script {
      & node tools\security_audit.mjs --out $ReportDir
      $code = if ($null -ne $LASTEXITCODE) { [int]$LASTEXITCODE } else { 0 }
      if ($Strict -and $code -ne 0) { throw "auditoria de seguridad fallo con codigo $code" }
      "Auditoria de seguridad finalizo con codigo $code"
    }

    Invoke-Captured -Title "Auditoria de permisos y licencias" -Required:$Strict -Script {
      & node tools\permissions_license_audit.mjs --out $ReportDir
      $code = if ($null -ne $LASTEXITCODE) { [int]$LASTEXITCODE } else { 0 }
      if ($Strict -and $code -ne 0) { throw "auditoria de permisos/licencias fallo con codigo $code" }
      "Auditoria de permisos/licencias finalizo con codigo $code"
    }

    Invoke-Captured -Title "Inventario OpenAPI generado" -Required -Script {
      & node tools\openapi_inventory.mjs --out documentos/api/openapi.generated.yaml
      if ($LASTEXITCODE -ne 0) { throw "generacion OpenAPI fallo con codigo $LASTEXITCODE" }
    }

    Invoke-Captured -Title "Reporte de observabilidad" -Required:$Strict -Script {
      & node tools\observability_report.mjs --out $ReportDir
      $code = if ($null -ne $LASTEXITCODE) { [int]$LASTEXITCODE } else { 0 }
      if ($Strict -and $code -ne 0) { throw "reporte de observabilidad fallo con codigo $code" }
      "Reporte de observabilidad finalizo con codigo $code"
    }
  }

  if (-not $SkipDockerConfig) {
    Invoke-Captured -Title "Validacion Docker Compose" -Required:$Strict -Script {
      if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        "Docker no disponible localmente; validacion omitida."
        return
      }
      & docker compose --env-file deploy\.env.platform.example -f deploy\docker-compose.platform.yml config --quiet
      if ($LASTEXITCODE -ne 0) { throw "docker compose config fallo con codigo $LASTEXITCODE" }
      "Docker Compose config ok"
    }
  }

  if ($Full -and -not $SkipGoTests) {
    Invoke-Captured -Title "Go test backend completo" -Required:$Strict -Script {
      Push-Location backend
      try {
        & go test ./...
        if ($LASTEXITCODE -ne 0) { throw "go test fallo con codigo $LASTEXITCODE" }
      } finally {
        Pop-Location
      }
    }
  } else {
    Add-Section -Title "Go test backend completo" -Body "Omitido en modo rapido. Ejecuta ``.\\scripts\\profesional_preflight.ps1 -Full`` para incluirlo."
  }

  Invoke-Captured -Title "git diff --check" -Required -Script {
    $previousErrorActionPreference = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
      $diffCheck = & git diff --check 2>&1
    } finally {
      $ErrorActionPreference = $previousErrorActionPreference
    }
    $problems = @($diffCheck | Where-Object {
      $_ -match ":\d+:" -or
      $_ -match "trailing whitespace" -or
      $_ -match "space before tab" -or
      $_ -match "new blank line at EOF"
    })
    $diffCheck
    if ($problems.Count -gt 0) {
      throw "git diff --check encontro errores de espacios o marcadores de conflicto"
    }
    "git diff --check ok"
  }
} finally {
  Pop-Location
}

$header = @"
# Preflight profesional

Fecha: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")
Estado: $status
Modo full: $([bool]$Full)

"@

Set-Content -LiteralPath $reportPath -Value ($header + ($sections -join "`n")) -Encoding UTF8
Write-Host "`n[INFO] Reporte: $reportPath"

if ($status -eq "ERROR") {
  exit 1
}
exit 0
