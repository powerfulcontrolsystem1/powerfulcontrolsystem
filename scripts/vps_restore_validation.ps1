<#
.SYNOPSIS
  Valida restaurabilidad de backups Docker/PostgreSQL en la VPS.

.DESCRIPTION
  Por defecto no modifica datos: encuentra el ultimo snapshot, verifica dump
  PostgreSQL y tarballs esperados. Con -ExecuteDrill crea un contenedor
  PostgreSQL temporal y restaura el dump para probar que el respaldo abre.
#>

param(
  [string]$RemoteUser = "root",
  [string]$RemoteHost = "2.24.197.58",
  [int]$Port = 22,
  [string]$IdentityFile = "",
  [string]$RemotePath = "/root/powerfulcontrolsystem",
  [string]$BackupDir = "",
  [switch]$ExecuteDrill
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
if ([string]::IsNullOrWhiteSpace($IdentityFile)) {
  $candidate = Join-Path $repoRoot "clave privada ssh.ppk"
  if (Test-Path -LiteralPath $candidate) { $IdentityFile = (Resolve-Path $candidate).Path }
}
if ([string]::IsNullOrWhiteSpace($IdentityFile) -or -not (Test-Path -LiteralPath $IdentityFile)) {
  throw "No se encontro IdentityFile. Indicalo con -IdentityFile."
}

function Resolve-Plink {
  $cmd = Get-Command plink.exe -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source -ErrorAction SilentlyContinue
  if ($cmd) { return $cmd }
  foreach ($candidate in @("D:\Program Files\PuTTY\plink.exe", "C:\Program Files\PuTTY\plink.exe", "C:\Program Files (x86)\PuTTY\plink.exe")) {
    if (Test-Path -LiteralPath $candidate) { return $candidate }
  }
  throw "No se encontro plink.exe."
}

function Convert-ToBashLiteral {
  param([AllowNull()][AllowEmptyString()][string]$Value = "")
  if ($null -eq $Value) { $Value = "" }
  return "'" + $Value.Replace("'", "'\''") + "'"
}

$remotePathLit = Convert-ToBashLiteral $RemotePath
$backupDirLit = Convert-ToBashLiteral $BackupDir
$execute = if ($ExecuteDrill) { "1" } else { "0" }

$remoteScript = @"
set -e
remote_path=$remotePathLit
backup_dir=$backupDirLit
execute_drill=$execute
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

for artifact in \
  powerful-control-system_pcs_web_uploads.tar.gz \
  powerful-control-system_pcs_downloads.tar.gz \
  powerful-control-system_pcs_backend_logs.tar.gz \
  powerful-control-system_pcs_backups.tar.gz \
  powerful-control-system_pcs_postgres_data.tar.gz \
  powerful-control-system_pcs_letsencrypt.tar.gz \
  powerful-control-system_pcs_certbot_www.tar.gz; do
  if [ -f "`$backup_dir/`$artifact" ]; then
    tar -tzf "`$backup_dir/`$artifact" >/dev/null
    echo "[OK] Tarball valido: `$artifact"
  else
    echo "[WARN] Tarball ausente: `$artifact"
  fi
done

if [ "`$execute_drill" = "1" ]; then
  drill="pcs-restore-drill-`$(date +%s)"
  echo "[INFO] Ejecutando restauracion temporal en contenedor `$drill"
  docker run --name "`$drill" -e POSTGRES_PASSWORD=restore_drill -d postgres:16-alpine >/dev/null
  cleanup() { docker rm -f "`$drill" >/dev/null 2>&1 || true; }
  trap cleanup EXIT
  sleep 8
  gunzip -c "`$backup_dir/postgres_all.sql.gz" | docker exec -i "`$drill" psql -U postgres >/tmp/pcs_restore_drill.log
  docker exec "`$drill" psql -U postgres -tAc "select 1" >/dev/null
  echo "[OK] Restauracion temporal PostgreSQL completada."
else
  echo "[OK] Validacion no destructiva completada. Use -ExecuteDrill para restaurar en contenedor temporal."
fi
"@

$tmp = Join-Path $env:TEMP ("pcs_vps_restore_validation_" + (Get-Date -Format "yyyyMMdd_HHmmss") + ".sh")
$utf8NoBom = New-Object System.Text.UTF8Encoding($false)
[System.IO.File]::WriteAllText($tmp, $remoteScript, $utf8NoBom)
try {
  $plink = Resolve-Plink
  & $plink -batch -P $Port -i $IdentityFile -m $tmp "$RemoteUser@$RemoteHost"
  if ($LASTEXITCODE -ne 0) { throw "Validacion de restauracion fallo con codigo $LASTEXITCODE." }
} finally {
  Remove-Item -LiteralPath $tmp -Force -ErrorAction SilentlyContinue
}
