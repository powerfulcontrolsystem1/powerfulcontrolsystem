<#
.SYNOPSIS
  Despliega OnlyOffice Document Server en la VPS (Docker).

.DESCRIPTION
  - Copia deploy/onlyoffice/docker-compose.yml al VPS.
  - Crea un archivo .env con ONLYOFFICE_JWT_SECRET.
  - Instala dependencias Docker básicas (apt-get) si faltan.
  - Levanta el servicio con docker compose up -d.

  Usa los mismos parámetros/convenciones que sync_to_vps.ps1:
  scripts/pcs_deployment.local.ps1 puede definir:
    PcsVpsHost, PcsVpsUser, PcsVpsRemotePath, PcsVpsPort, PcsVpsIdentityFile
#>

param(
  [string]$RemoteUser = "root",
  [string]$RemoteHost = "2.24.197.58",
  [string]$RemotePath = "/root/powerfulcontrolsystem",
  [int]$Port = 22,
  [string]$IdentityFile = "",
  [string]$JwtSecret = "",
  [switch]$DryRun
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$pcsDeployVps = Join-Path $PSScriptRoot "pcs_deployment.local.ps1"
if (Test-Path -LiteralPath $pcsDeployVps) {
  . $pcsDeployVps
  if (-not $PSBoundParameters.ContainsKey('RemoteHost') -and (Get-Variable PcsVpsHost -Scope Script -ErrorAction SilentlyContinue) -and -not [string]::IsNullOrWhiteSpace($script:PcsVpsHost)) {
    $RemoteHost = $script:PcsVpsHost.Trim()
  }
  if (-not $PSBoundParameters.ContainsKey('RemoteUser') -and (Get-Variable PcsVpsUser -Scope Script -ErrorAction SilentlyContinue) -and -not [string]::IsNullOrWhiteSpace($script:PcsVpsUser)) {
    $RemoteUser = $script:PcsVpsUser.Trim()
  }
  if (-not $PSBoundParameters.ContainsKey('RemotePath') -and (Get-Variable PcsVpsRemotePath -Scope Script -ErrorAction SilentlyContinue) -and -not [string]::IsNullOrWhiteSpace($script:PcsVpsRemotePath)) {
    $RemotePath = $script:PcsVpsRemotePath.Trim()
  }
  if (-not $PSBoundParameters.ContainsKey('Port') -and (Get-Variable PcsVpsPort -Scope Script -ErrorAction SilentlyContinue) -and $null -ne $script:PcsVpsPort) {
    $Port = [int]$script:PcsVpsPort
  }
  if (-not $PSBoundParameters.ContainsKey('IdentityFile') -and (Get-Variable PcsVpsIdentityFile -Scope Script -ErrorAction SilentlyContinue) -and -not [string]::IsNullOrWhiteSpace($script:PcsVpsIdentityFile)) {
    $IdentityFile = $script:PcsVpsIdentityFile.Trim()
  }
}

function Resolve-Plink {
  $cmd = Get-Command plink.exe -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source -ErrorAction SilentlyContinue
  if ($cmd) { return $cmd }
  $candidates = @(
    "D:\Program Files\PuTTY\plink.exe",
    "C:\Program Files\PuTTY\plink.exe",
    "C:\Program Files (x86)\PuTTY\plink.exe"
  )
  foreach ($c in $candidates) { if (Test-Path $c) { return $c } }
  return ""
}

function Resolve-Pscp {
  $cmd = Get-Command pscp.exe -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source -ErrorAction SilentlyContinue
  if ($cmd) { return $cmd }
  $candidates = @(
    "D:\Program Files\PuTTY\pscp.exe",
    "C:\Program Files\PuTTY\pscp.exe",
    "C:\Program Files (x86)\PuTTY\pscp.exe"
  )
  foreach ($c in $candidates) { if (Test-Path $c) { return $c } }
  return ""
}

if ([string]::IsNullOrWhiteSpace($IdentityFile)) {
  throw "IdentityFile requerido. Configúralo en scripts/pcs_deployment.local.ps1 (PcsVpsIdentityFile) o pasa -IdentityFile."
}

if ([string]::IsNullOrWhiteSpace($JwtSecret)) {
  $JwtSecret = [Environment]::GetEnvironmentVariable("ONLYOFFICE_JWT_SECRET", "Process")
}
if ([string]::IsNullOrWhiteSpace($JwtSecret)) {
  # Generar secreto automáticamente (no imprimirlo).
  $bytes = New-Object byte[] 32
  [System.Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($bytes)
  $JwtSecret = [Convert]::ToBase64String($bytes)
}

$plink = Resolve-Plink
$pscp = Resolve-Pscp
if ([string]::IsNullOrWhiteSpace($plink) -or [string]::IsNullOrWhiteSpace($pscp)) {
  throw "No se encontró PuTTY (plink/pscp). Instala PuTTY o ponlo en PATH."
}

$remoteOnlyOfficeDir = ($RemotePath.TrimEnd("/") + "/deploy/onlyoffice")
$localCompose = Join-Path (Split-Path -Parent $PSScriptRoot) "deploy\onlyoffice\docker-compose.yml"
if (-not (Test-Path -LiteralPath $localCompose)) {
  throw "No existe docker-compose.yml local en: $localCompose"
}

$envTmp = [System.IO.Path]::GetTempFileName()
try {
  # No imprimir el secreto; solo escribirlo al archivo temporal.
  Set-Content -LiteralPath $envTmp -Value ("ONLYOFFICE_JWT_SECRET=" + $JwtSecret) -Encoding ASCII

  $sshBase = @("-P", "$Port", "-i", "$IdentityFile", "$RemoteUser@$RemoteHost")
  $scpBase = @("-P", "$Port", "-i", "$IdentityFile")

  $remoteBackendEnv = ($RemotePath.TrimEnd("/") + "/backend/.env.local")
  $remoteOnlyOfficeEnv = ($remoteOnlyOfficeDir + "/.env")

  if ($DryRun) {
    Write-Host "[DRYRUN] subir compose -> $remoteOnlyOfficeDir/docker-compose.yml"
    Write-Host "[DRYRUN] subir env -> $remoteOnlyOfficeDir/.env"
    Write-Host "[DRYRUN] ejecutar: instalar docker si falta, compose up -d, healthcheck"
    exit 0
  }

  & $plink @sshBase "mkdir -p $remoteOnlyOfficeDir" | Out-Null
  & $pscp @scpBase $localCompose ("${RemoteUser}@${RemoteHost}:$remoteOnlyOfficeDir/docker-compose.yml") | Out-Null
  & $pscp @scpBase $envTmp ("${RemoteUser}@${RemoteHost}:$remoteOnlyOfficeDir/.env") | Out-Null

  # Ejecutar en pasos separados para evitar problemas de quoting y para no imprimir el secreto.
  & $plink @sshBase "bash -lc 'set -e; if ! command -v docker >/dev/null 2>&1; then apt-get update -y; apt-get install -y docker.io docker-compose curl; fi'" | Out-Null
  & $plink @sshBase "bash -lc 'systemctl enable --now docker >/dev/null 2>&1 || true'" | Out-Null
  $remoteBackendDir = ($RemotePath.TrimEnd("/") + "/backend")
  & $plink @sshBase "bash -lc 'mkdir -p $remoteBackendDir; touch $remoteBackendEnv'" | Out-Null
  & $plink @sshBase "bash -lc 'if ! grep -q \"^ONLYOFFICE_JWT_SECRET=\" $remoteBackendEnv; then cat $remoteOnlyOfficeEnv >> $remoteBackendEnv; fi'" | Out-Null
  & $plink @sshBase "bash -lc 'if ! grep -q \"^ONLYOFFICE_DOCUMENT_SERVER_URL=\" $remoteBackendEnv; then echo ONLYOFFICE_DOCUMENT_SERVER_URL=https://onlyoffice.powerfulcontrolsystem.com >> $remoteBackendEnv; fi'" | Out-Null
  & $plink @sshBase "bash -lc 'set -e; cd $remoteOnlyOfficeDir; if docker compose version >/dev/null 2>&1; then docker compose up -d; docker compose ps; else docker-compose up -d; docker-compose ps; fi'"
  & $plink @sshBase "bash -lc 'curl -fsS http://127.0.0.1:8088/healthcheck || true'"
} finally {
  try { Remove-Item -LiteralPath $envTmp -Force -ErrorAction SilentlyContinue } catch {}
}

