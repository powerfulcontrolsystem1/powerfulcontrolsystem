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
  [switch]$VerifyCriticalData,
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
if ($VerifyCriticalData -and -not $ExecuteDrill) {
  throw "VerifyCriticalData requiere ExecuteDrill porque consulta una restauracion temporal."
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
$verifyCritical = if ($VerifyCriticalData) { "1" } else { "0" }

$remoteScript = @"
set -e
validation_started_at=`$(date +%s)
remote_path=$remotePathLit
backup_dir=$backupDirLit
execute_drill=$execute
verify_critical=$verifyCritical
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
  if [ "`$verify_critical" = "1" ]; then
    database_count="`$(docker exec "`$drill" psql -U postgres -tAc "SELECT count(*) FROM pg_database WHERE datname IN ('pcs_empresas','pcs_superadministrador')")"
    if [ "`$database_count" != "2" ]; then
      echo "[ERROR] Restore critico: se esperaban dos bases y se encontraron `$database_count."
      exit 1
    fi

    critical_tables="empresa_cuentas_por_pagar empresa_asientos_contables empresa_ai_memoria empresa_dian_configuracion empresa_documentos_gestion"
    checked_tables=0
    for table in `$critical_tables; do
      exists="`$(docker exec "`$drill" psql -U postgres -d pcs_empresas -tAc "SELECT to_regclass('public.`$table') IS NOT NULL")"
      if [ "`$exists" != "t" ]; then
        echo "[ERROR] Restore critico: tabla obligatoria ausente: `$table"
        exit 1
      fi
      docker exec "`$drill" psql -U postgres -d pcs_empresas -tAc "SELECT count(*) FROM `$table WHERE empresa_id=12" >/dev/null
      checked_tables=`$((checked_tables+1))
    done

    private_tar="`$backup_dir/powerful-control-system_pcs_private_storage.tar.gz"
    private_member_count="`$(tar -tzf "`$private_tar" | awk '!/\/$/ {count++} END {print count+0}')"
    support_total="`$(docker exec "`$drill" psql -U postgres -d pcs_empresas -tAc "SELECT count(*) FROM empresa_soportes_compras_ia WHERE archivo_url LIKE 'private://soportes_compras_ia/%'")"
    support_hashed="`$(docker exec "`$drill" psql -U postgres -d pcs_empresas -tAc "SELECT count(*) FROM empresa_soportes_compras_ia WHERE archivo_url LIKE 'private://soportes_compras_ia/%' AND btrim(COALESCE(archivo_hash,'')) <> ''")"
    if [ "`$support_total" != "`$support_hashed" ]; then
      echo "[ERROR] Restore critico: existen soportes IA privados sin checksum."
      exit 1
    fi

    checked_hashes=0
    while IFS='|' read -r empresa_id archivo_url archivo_hash; do
      if [ -z "`$archivo_url" ]; then
        continue
      fi
      expected_prefix="private://soportes_compras_ia/empresa_`$empresa_id/"
      case "`$archivo_url" in
        "`$expected_prefix"*) ;;
        *)
          echo "[ERROR] Restore critico: referencia de soporte IA fuera de su empresa."
          exit 1
          ;;
      esac
      relative="`${archivo_url#private://}"
      member="./`$relative"
      if ! tar -tzf "`$private_tar" "`$member" >/dev/null 2>&1; then
        echo "[ERROR] Restore critico: falta un soporte IA referenciado en el volumen privado."
        exit 1
      fi
      actual_hash="`$(tar -xOzf "`$private_tar" "`$member" | sha256sum | awk '{print `$1}')"
      if [ "`$actual_hash" != "`$archivo_hash" ]; then
        echo "[ERROR] Restore critico: checksum de soporte IA no coincide."
        exit 1
      fi
      checked_hashes=`$((checked_hashes+1))
    done <<EOF
`$(docker exec "`$drill" psql -U postgres -d pcs_empresas -AtF '|' -c "SELECT empresa_id,archivo_url,lower(archivo_hash) FROM empresa_soportes_compras_ia WHERE archivo_url LIKE 'private://soportes_compras_ia/%' ORDER BY empresa_id,id")
EOF

    echo "[OK] Restore critico: bases=2 tablas=`$checked_tables filtros_empresa=5 archivos_privados=`$private_member_count checksums_soportes_ia=`$checked_hashes"
  fi
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
