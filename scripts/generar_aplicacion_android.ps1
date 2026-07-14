<#
.SYNOPSIS
  Valida y genera artefactos Android de la aplicacion Flutter PCS.
#>
param(
  [switch]$DryRun,
  [switch]$ValidateOnly,
  [switch]$Debug,
  [switch]$Release
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$appDir = Join-Path $repoRoot 'mobile\powerful_control_system_app'
$artifactsDir = Join-Path $repoRoot 'artifacts\mobile\android'

if (-not (Test-Path -LiteralPath (Join-Path $appDir 'pubspec.yaml'))) { throw "No se encontro el proyecto Flutter en $appDir" }
if ($Debug -and $Release) { throw 'Usa solo uno de -Debug o -Release.' }
if (-not $Debug -and -not $Release) { $Release = $true }

function Invoke-Flutter([string[]]$Arguments) {
  Write-Host ('[INFO] flutter ' + ($Arguments -join ' ')) -ForegroundColor DarkGray
  & flutter @Arguments
  if ($LASTEXITCODE -ne 0) { throw "Flutter fallo: flutter $($Arguments -join ' ')" }
}

if ($DryRun) {
  Write-Host '[DRY RUN] Se validaria Flutter, analizaria, ejecutaria pruebas y generaria APK/AAB sin publicar.' -ForegroundColor Yellow
  exit 0
}
if (-not (Get-Command flutter -ErrorAction SilentlyContinue)) {
  throw 'Flutter no esta instalado o no esta en PATH. Instala Flutter estable y ejecuta flutter doctor antes de generar Android.'
}

Push-Location $appDir
try {
  Invoke-Flutter @('--version')
  Invoke-Flutter @('doctor', '-v')
  if (-not (Test-Path -LiteralPath (Join-Path $appDir 'android'))) {
    Invoke-Flutter @('create', '--platforms=android,ios', '.')
  }
  Invoke-Flutter @('pub', 'get')
  & dart run flutter_launcher_icons
  if ($LASTEXITCODE -ne 0) { throw 'No fue posible generar los iconos nativos de la aplicacion.' }
  Invoke-Flutter @('analyze')
  Invoke-Flutter @('test')
  if ($ValidateOnly) { Write-Host '[OK] Validacion Android completada.' -ForegroundColor Green; exit 0 }
  New-Item -ItemType Directory -Force -Path $artifactsDir | Out-Null
  if ($Debug) {
    Invoke-Flutter @('build', 'apk', '--debug')
    $source = Join-Path $appDir 'build\app\outputs\flutter-apk\app-debug.apk'
  } else {
    Invoke-Flutter @('build', 'apk', '--release')
    Invoke-Flutter @('build', 'appbundle', '--release')
    $source = Join-Path $appDir 'build\app\outputs\flutter-apk\app-release.apk'
    $bundle = Join-Path $appDir 'build\app\outputs\bundle\release\app-release.aab'
    Copy-Item -LiteralPath $bundle -Destination (Join-Path $artifactsDir 'powerful-control-system-release.aab') -Force
  }
  Copy-Item -LiteralPath $source -Destination (Join-Path $artifactsDir ('powerful-control-system-' + $(if ($Debug) { 'debug' } else { 'release' }) + '.apk')) -Force
  Get-ChildItem $artifactsDir -File | ForEach-Object { (Get-FileHash -Algorithm SHA256 $_.FullName).Hash + '  ' + $_.Name | Set-Content -LiteralPath ($_.FullName + '.sha256') -NoNewline }
  Write-Host "[OK] Artefactos Android: $artifactsDir" -ForegroundColor Green
} finally { Pop-Location }
