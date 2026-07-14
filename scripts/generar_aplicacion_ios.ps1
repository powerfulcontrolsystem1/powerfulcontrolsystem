<#
.SYNOPSIS
  Valida y genera el archivo IPA de la aplicacion Flutter PCS desde macOS.
.DESCRIPTION
  iOS solo se compila en macOS con Xcode y una firma valida. En Windows esta
  herramienta valida el proyecto si Flutter existe; no intenta fabricar un IPA.
#>
param(
  [switch]$DryRun,
  [switch]$ValidateOnly,
  [switch]$Debug,
  [switch]$Release,
  [switch]$TriggerIOSWorkflow
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$appDir = Join-Path $repoRoot 'mobile\powerful_control_system_app'
$artifactsDir = Join-Path $repoRoot 'artifacts\mobile\ios'

if ($Debug -and $Release) { throw 'Usa solo uno de -Debug o -Release.' }
if (-not $Debug -and -not $Release) { $Release = $true }
if (-not (Test-Path -LiteralPath (Join-Path $appDir 'pubspec.yaml'))) { throw "No se encontro el proyecto Flutter en $appDir" }

function Invoke-Flutter([string[]]$Arguments) {
  Write-Host ('[INFO] flutter ' + ($Arguments -join ' ')) -ForegroundColor DarkGray
  & flutter @Arguments
  if ($LASTEXITCODE -ne 0) { throw "Flutter fallo: flutter $($Arguments -join ' ')" }
}

if ($DryRun) {
  Write-Host '[DRY RUN] Se validaria Flutter y se generaria IPA solo desde macOS/Xcode; no se publica.' -ForegroundColor Yellow
  exit 0
}
if (-not (Get-Command flutter -ErrorAction SilentlyContinue)) {
  throw 'Flutter no esta instalado o no esta en PATH. Instala Flutter estable y ejecuta flutter doctor.'
}
if (-not $IsMacOS) {
  if ($TriggerIOSWorkflow) {
    if (-not (Get-Command gh -ErrorAction SilentlyContinue)) { throw 'GitHub CLI no esta disponible para activar el workflow iOS.' }
    & gh workflow run mobile-release.yml --ref (git branch --show-current)
    if ($LASTEXITCODE -ne 0) { throw 'No fue posible activar el workflow iOS.' }
    exit 0
  }
  throw 'La compilacion IPA requiere macOS, Xcode y una firma Apple. Use -TriggerIOSWorkflow o ejecute este script en macOS.'
}

Push-Location $appDir
try {
  Invoke-Flutter @('--version')
  Invoke-Flutter @('doctor', '-v')
  if (-not (Test-Path -LiteralPath (Join-Path $appDir 'ios'))) { Invoke-Flutter @('create', '--platforms=android,ios', '.') }
  Invoke-Flutter @('pub', 'get')
  & dart run flutter_launcher_icons
  if ($LASTEXITCODE -ne 0) { throw 'No fue posible generar los iconos nativos de la aplicacion.' }
  Invoke-Flutter @('analyze')
  Invoke-Flutter @('test')
  if ($ValidateOnly) { Write-Host '[OK] Validacion iOS completada.' -ForegroundColor Green; exit 0 }
  New-Item -ItemType Directory -Force -Path $artifactsDir | Out-Null
  if ($Debug) {
    Invoke-Flutter @('build', 'ios', '--debug', '--no-codesign')
  } else {
    Invoke-Flutter @('build', 'ipa', '--release')
    $ipa = Get-ChildItem -LiteralPath (Join-Path $appDir 'build\ios\ipa') -Filter '*.ipa' -File | Select-Object -First 1
    if ($null -eq $ipa) { throw 'Flutter no genero el IPA esperado.' }
    Copy-Item -LiteralPath $ipa.FullName -Destination (Join-Path $artifactsDir 'powerful-control-system-release.ipa') -Force
  }
  Get-ChildItem -LiteralPath $artifactsDir -File | ForEach-Object {
    (Get-FileHash -Algorithm SHA256 $_.FullName).Hash + '  ' + $_.Name | Set-Content -LiteralPath ($_.FullName + '.sha256') -NoNewline
  }
  Write-Host "[OK] Artefactos iOS: $artifactsDir" -ForegroundColor Green
} finally { Pop-Location }
