<#
.SYNOPSIS
  Crea respaldos operativos de datos persistentes Docker en la VPS.

.DESCRIPTION
  Ejecuta en la VPS un respaldo consistente de PostgreSQL via pg_dumpall y
  empaqueta volumenes persistentes de uploads, descargas, logs y backups.
  No detiene contenedores. No trae archivos al equipo local por defecto.
#>

param(
  [string]$RemoteUser = "",
  [string]$RemoteHost = "",
  [int]$Port = 0,
  [string]$IdentityFile = "",
  [string]$RemotePath = "",
  [int]$RetentionDays = 14,
  [switch]$IncludeEnvSecrets,
  [switch]$DryRun,
  [switch]$AllowRemoteTarget
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$deploymentConfig = Join-Path $PSScriptRoot "pcs_deployment.local.ps1"
if (Test-Path -LiteralPath $deploymentConfig) {
  . $deploymentConfig
}
if ([string]::IsNullOrWhiteSpace($RemoteUser) -and (Get-Variable -Name PcsVpsUser -Scope Script -ErrorAction SilentlyContinue)) { $RemoteUser = $script:PcsVpsUser }
if ([string]::IsNullOrWhiteSpace($RemoteHost) -and (Get-Variable -Name PcsVpsHost -Scope Script -ErrorAction SilentlyContinue)) { $RemoteHost = $script:PcsVpsHost }
if ($Port -le 0 -and (Get-Variable -Name PcsVpsPort -Scope Script -ErrorAction SilentlyContinue)) { $Port = [int]$script:PcsVpsPort }
if ([string]::IsNullOrWhiteSpace($RemotePath) -and (Get-Variable -Name PcsVpsRemotePath -Scope Script -ErrorAction SilentlyContinue)) { $RemotePath = $script:PcsVpsRemotePath }
if ([string]::IsNullOrWhiteSpace($IdentityFile) -and (Get-Variable -Name PcsVpsIdentityFile -Scope Script -ErrorAction SilentlyContinue)) { $IdentityFile = $script:PcsVpsIdentityFile }
if (-not $AllowRemoteTarget) {
  throw "Operacion remota bloqueada por seguridad. Usa -AllowRemoteTarget solo despues de verificar que el destino es aislado o esta expresamente autorizado."
}
if ([string]::IsNullOrWhiteSpace($RemoteUser) -or [string]::IsNullOrWhiteSpace($RemoteHost) -or $Port -le 0 -or [string]::IsNullOrWhiteSpace($RemotePath)) {
  throw "Faltan parametros de destino remoto. Configuralos localmente o indicalos de forma explicita."
}
if ([string]::IsNullOrWhiteSpace($IdentityFile)) {
  $candidate = Join-Path $repoRoot "clave privada ssh.ppk"
  if (Test-Path -LiteralPath $candidate) {
    $IdentityFile = (Resolve-Path $candidate).Path
  }
}
if ([string]::IsNullOrWhiteSpace($IdentityFile)) {
  foreach ($candidate in @(
    (Join-Path $HOME ".ssh\id_rsa"),
    (Join-Path $HOME ".ssh\id_ed25519")
  )) {
    if (Test-Path -LiteralPath $candidate) {
      $IdentityFile = (Resolve-Path $candidate).Path
      break
    }
  }
}
if ([string]::IsNullOrWhiteSpace($IdentityFile) -or -not (Test-Path -LiteralPath $IdentityFile)) {
  throw "No se encontro IdentityFile. Indicalo con -IdentityFile."
}
if ($RetentionDays -lt 1) {
  $RetentionDays = 14
}

function Resolve-Plink {
  $cmd = Get-Command plink.exe -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source -ErrorAction SilentlyContinue
  if ($cmd) { return $cmd }
  foreach ($candidate in @("D:\Program Files\PuTTY\plink.exe", "C:\Program Files\PuTTY\plink.exe", "C:\Program Files (x86)\PuTTY\plink.exe")) {
    if (Test-Path -LiteralPath $candidate) { return $candidate }
  }
  throw "No se encontro plink.exe. Instala PuTTY o agrega plink al PATH."
}

function Resolve-OpenSSH {
  $cmd = Get-Command ssh.exe -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source -ErrorAction SilentlyContinue
  if ($cmd) { return $cmd }
  foreach ($candidate in @("C:\Windows\System32\OpenSSH\ssh.exe", "C:\Program Files\Git\usr\bin\ssh.exe")) {
    if (Test-Path -LiteralPath $candidate) { return $candidate }
  }
  throw "No se encontro ssh.exe."
}

function Convert-ToBashLiteral {
  param([AllowNull()][AllowEmptyString()][string]$Value = "")
  if ($null -eq $Value) { $Value = "" }
  return "'" + $Value.Replace("'", "'\''") + "'"
}

$remotePathLit = Convert-ToBashLiteral $RemotePath
$includeSecrets = if ($IncludeEnvSecrets) { "1" } else { "0" }
$dry = if ($DryRun) { "1" } else { "0" }

$remoteScript = @"
set -e
remote_path=$remotePathLit
retention_days=$RetentionDays
include_env_secrets=$includeSecrets
dry_run=$dry

if [ "`$dry_run" = "1" ]; then
  echo "[PREVIEW] Backup VPS: no se escribiran respaldos."
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "[ERROR] Docker no esta disponible en la VPS."
  exit 1
fi

backup_root="`$remote_path/backups/vps-snapshots"
stamp="`$(date +%Y%m%d_%H%M%S)"
backup_dir="`$backup_root/`$stamp"

echo "[INFO] Backup VPS: destino `$backup_dir"
if [ "`$dry_run" != "1" ]; then
  mkdir -p "`$backup_dir"
  chmod 700 "`$backup_root" "`$backup_dir" 2>/dev/null || true
fi

if docker ps --format '{{.Names}}' | grep -qx 'pcs-postgres'; then
  echo "[INFO] Backup VPS: generando pg_dumpall desde pcs-postgres."
  if [ "`$dry_run" != "1" ]; then
    docker exec pcs-postgres sh -lc 'pg_dumpall -U "`$POSTGRES_USER"' > "`$backup_dir/postgres_all.sql"
    gzip -9 "`$backup_dir/postgres_all.sql"
  fi
else
  echo "[WARN] Backup VPS: pcs-postgres no esta activo; se omite dump PostgreSQL Docker."
fi

volumes="
powerful-control-system_pcs_web_uploads
powerful-control-system_pcs_downloads
powerful-control-system_pcs_backend_logs
powerful-control-system_pcs_backups
powerful-control-system_pcs_private_storage
powerful-control-system_mailu_certs
powerful-control-system_pcs_onlyoffice_data
powerful-control-system_pcs_onlyoffice_lib
powerful-control-system_pcs_onlyoffice_logs
powerful-control-system_pcs_letsencrypt
powerful-control-system_pcs_certbot_www
"

for volume in `$volumes; do
  if docker volume inspect "`$volume" >/dev/null 2>&1; then
    echo "[INFO] Backup VPS: empaquetando volumen `$volume"
    if [ "`$dry_run" != "1" ]; then
      docker run --rm -v "`$volume:/volume:ro" -v "`$backup_dir:/backup" alpine:3.20 sh -lc "cd /volume && tar -czf /backup/`$volume.tar.gz ."
    fi
  else
    echo "[INFO] Backup VPS: volumen no encontrado, omitido: `$volume"
  fi
done

if [ "`$include_env_secrets" = "1" ] && [ -f "`$remote_path/deploy/.env.platform" ]; then
  echo "[INFO] Backup VPS: copiando deploy/.env.platform dentro del respaldo privado."
  if [ "`$dry_run" != "1" ]; then
    cp "`$remote_path/deploy/.env.platform" "`$backup_dir/env.platform.backup"
    chmod 600 "`$backup_dir/env.platform.backup" 2>/dev/null || true
  fi
fi

if [ "`$dry_run" != "1" ]; then
  echo "[INFO] Backup VPS: aplicando retencion de `$retention_days dias."
  find "`$backup_root" -mindepth 1 -maxdepth 1 -type d -mtime +"`$retention_days" -print -exec rm -rf {} \; 2>/dev/null || true
  du -sh "`$backup_dir" 2>/dev/null || true
  echo "[OK] Backup VPS completado: `$backup_dir"
else
  echo "[OK] Preview de backup VPS completado."
fi
"@

$tmp = Join-Path $env:TEMP ("pcs_vps_backup_" + (Get-Date -Format "yyyyMMdd_HHmmss") + ".sh")
$utf8NoBom = New-Object System.Text.UTF8Encoding($false)
[System.IO.File]::WriteAllText($tmp, $remoteScript, $utf8NoBom)
try {
  if ([System.IO.Path]::GetExtension($IdentityFile).ToLowerInvariant() -eq ".ppk") {
    $plink = Resolve-Plink
    & $plink -batch -P $Port -i $IdentityFile -m $tmp "$RemoteUser@$RemoteHost"
  } else {
    $ssh = Resolve-OpenSSH
    $remoteScriptUnix = ((Get-Content -LiteralPath $tmp -Raw) -replace "`r", "").TrimEnd("`n") + "`n"
    $psi = [System.Diagnostics.ProcessStartInfo]::new()
    $psi.FileName = $ssh
    $psi.UseShellExecute = $false
    $psi.RedirectStandardInput = $true
    $psi.RedirectStandardOutput = $true
    $psi.RedirectStandardError = $true
    foreach ($arg in @("-T", "-o", "BatchMode=yes", "-o", "IdentitiesOnly=yes", "-o", "StrictHostKeyChecking=accept-new", "-p", "$Port", "-i", "$IdentityFile", "$RemoteUser@$RemoteHost", "bash -s")) {
      [void]$psi.ArgumentList.Add($arg)
    }
    $process = [System.Diagnostics.Process]::new()
    $process.StartInfo = $psi
    [void]$process.Start()
    $process.StandardInput.Write($remoteScriptUnix)
    $process.StandardInput.Close()
    $stdout = $process.StandardOutput.ReadToEnd()
    $stderr = $process.StandardError.ReadToEnd()
    $process.WaitForExit()
    if (-not [string]::IsNullOrWhiteSpace($stdout)) { Write-Output $stdout.TrimEnd() }
    if (-not [string]::IsNullOrWhiteSpace($stderr)) { Write-Error $stderr.TrimEnd() }
    $global:LASTEXITCODE = $process.ExitCode
  }
  if ($LASTEXITCODE -ne 0) {
    throw "Backup VPS fallo con codigo $LASTEXITCODE."
  }
} finally {
  Remove-Item -LiteralPath $tmp -Force -ErrorAction SilentlyContinue
}
