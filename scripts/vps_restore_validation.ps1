<#
.SYNOPSIS
  Valida restaurabilidad de backups Docker/PostgreSQL en la VPS.

.DESCRIPTION
  Por defecto no modifica datos: encuentra el ultimo snapshot, verifica dump
  PostgreSQL y tarballs esperados. Con -ExecuteDrill crea un contenedor
  PostgreSQL temporal y restaura el dump para probar que el respaldo abre.
#>

param(
  [string]$RemoteUser = "",
  [string]$RemoteHost = "",
  [int]$Port = 0,
  [string]$IdentityFile = "",
  [string]$RemotePath = "",
  [string]$BackupDir = "",
  [string]$RestoreImage = "postgres:16.14-alpine",
  [switch]$ExecuteDrill,
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
  if (Test-Path -LiteralPath $candidate) { $IdentityFile = (Resolve-Path $candidate).Path }
}
if ([string]::IsNullOrWhiteSpace($IdentityFile)) {
  foreach ($candidate in @(
    # La VPS principal usa la clave RSA publicada para esta estacion. Se
    # conserva ED25519 como alternativa para destinos nuevos o aislados.
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
if ([string]::IsNullOrWhiteSpace($RestoreImage) -or $RestoreImage -match '[\r\n]') {
  throw "RestoreImage debe ser una referencia Docker no vacia en una sola linea."
}

function Resolve-Plink {
  $cmd = Get-Command plink.exe -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source -ErrorAction SilentlyContinue
  if ($cmd) { return $cmd }
  foreach ($candidate in @("D:\Program Files\PuTTY\plink.exe", "C:\Program Files\PuTTY\plink.exe", "C:\Program Files (x86)\PuTTY\plink.exe")) {
    if (Test-Path -LiteralPath $candidate) { return $candidate }
  }
  throw "No se encontro plink.exe."
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
$backupDirLit = Convert-ToBashLiteral $BackupDir
$restoreImageLit = Convert-ToBashLiteral $RestoreImage
$execute = if ($ExecuteDrill) { "1" } else { "0" }

$remoteScript = @"
set -e
validation_started_at=`$(date +%s)
remote_path=$remotePathLit
backup_dir=$backupDirLit
execute_drill=$execute
restore_image=$restoreImageLit
backup_root="`$remote_path/backups/vps-snapshots"

if [ -z "`$backup_dir" ]; then
  backup_dir="`$(find "`$backup_root" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | sort | tail -n 1)"
fi

if [ -z "`$backup_dir" ] || [ ! -d "`$backup_dir" ]; then
  echo "[ERROR] No se encontro snapshot de backup en `$backup_root"
  exit 1
fi

echo "[INFO] Validando snapshot: `$backup_dir"
test -s "`$backup_dir/postgres_all.sql.gz"
gzip -t "`$backup_dir/postgres_all.sql.gz"
snapshot_epoch="`$(stat -c %Y "`$backup_dir/postgres_all.sql.gz")"

for artifact in \
  powerful-control-system_pcs_web_uploads.tar.gz \
  powerful-control-system_pcs_downloads.tar.gz \
  powerful-control-system_pcs_backend_logs.tar.gz \
  powerful-control-system_pcs_backups.tar.gz \
  powerful-control-system_pcs_private_storage.tar.gz \
  powerful-control-system_mailu_certs.tar.gz \
  powerful-control-system_pcs_onlyoffice_data.tar.gz \
  powerful-control-system_pcs_onlyoffice_lib.tar.gz \
  powerful-control-system_pcs_onlyoffice_logs.tar.gz; do
  if [ ! -f "`$backup_dir/`$artifact" ]; then
    echo "[ERROR] Tarball obligatorio ausente: `$artifact"
    exit 1
  fi
  tar -tzf "`$backup_dir/`$artifact" >/dev/null
  echo "[OK] Tarball valido: `$artifact"
done

for artifact in \
  powerful-control-system_pcs_letsencrypt.tar.gz \
  powerful-control-system_pcs_certbot_www.tar.gz; do
  if [ -f "`$backup_dir/`$artifact" ]; then
    tar -tzf "`$backup_dir/`$artifact" >/dev/null
    echo "[OK] Tarball valido: `$artifact"
  else
    echo "[INFO] Tarball opcional no configurado: `$artifact"
  fi
done

if [ "`$execute_drill" = "1" ]; then
  drill="pcs-restore-drill-`$(date +%s)"
  echo "[INFO] Ejecutando restauracion temporal en contenedor `$drill con imagen `$restore_image"
  docker run --name "`$drill" -e POSTGRES_PASSWORD=restore_drill -d "`$restore_image" >/dev/null
  cleanup() { docker rm -f "`$drill" >/dev/null 2>&1 || true; }
  trap cleanup EXIT
  sleep 8
  gunzip -c "`$backup_dir/postgres_all.sql.gz" | docker exec -i "`$drill" psql -U postgres >/tmp/pcs_restore_drill.log
  docker exec "`$drill" psql -U postgres -tAc "select 1" >/dev/null
  validation_finished_at="`$(date +%s)"
  echo "[OK] Restauracion temporal PostgreSQL completada. imagen=`$restore_image RTO=`$((validation_finished_at-validation_started_at))s RPO=`$((validation_finished_at-snapshot_epoch))s"
else
  validation_finished_at="`$(date +%s)"
  echo "[OK] Validacion no destructiva completada. RTO=`$((validation_finished_at-validation_started_at))s RPO=`$((validation_finished_at-snapshot_epoch))s. Use -ExecuteDrill para restaurar en contenedor temporal."
fi
"@

$tmp = Join-Path $env:TEMP ("pcs_vps_restore_validation_" + (Get-Date -Format "yyyyMMdd_HHmmss") + ".sh")
$utf8NoBom = New-Object System.Text.UTF8Encoding($false)
[System.IO.File]::WriteAllText($tmp, $remoteScript, $utf8NoBom)
try {
  if ([System.IO.Path]::GetExtension($IdentityFile).ToLowerInvariant() -eq ".ppk") {
    $plink = Resolve-Plink
    & $plink -batch -P $Port -i $IdentityFile -m $tmp "$RemoteUser@$RemoteHost"
  } else {
    $ssh = Resolve-OpenSSH
    # bash remoto debe recibir LF. El archivo temporal se crea en Windows y
    # CRLF hace que terminadores como `fi` se interpreten incorrectamente.
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
  if ($LASTEXITCODE -ne 0) { throw "Validacion de restauracion fallo con codigo $LASTEXITCODE." }
} finally {
  Remove-Item -LiteralPath $tmp -Force -ErrorAction SilentlyContinue
}
