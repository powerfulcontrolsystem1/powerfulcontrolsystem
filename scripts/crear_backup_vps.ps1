<#
.SYNOPSIS
  Crea una copia de recuperacion del VPS PCS en D:\Backup vps PCS.

.DESCRIPTION
  Ejecuta un respaldo remoto no destructivo: dump logico de PostgreSQL,
  imagenes Docker locales, inventario Docker, volumenes Docker, archivos del
  proyecto y scripts de restauracion. Descarga todo a una carpeta versionada en
  D:\Backup vps PCS, manteniendo backups anteriores.

  Modo restauracion/subida:
    .\scripts\crear_backup_vps.ps1 -Restore -BackupPath "D:\Backup vps PCS\...\pcs_vps_full_backup_....tar.gz" -TargetHost "IP_NUEVO_VPS"

  Por seguridad, -Restore sube el paquete y deja preparado el restaurador. Para
  ejecutarlo remotamente se requiere -ExecuteRemoteRestore.
#>

param(
  [switch]$Restore,
  [switch]$ExecuteRemoteRestore,
  [string]$BackupPath = "",
  [string]$BackupRoot = "D:\Backup vps PCS",
  [string]$RemoteUser = "root",
  [string]$RemoteHost = "2.24.197.58",
  [int]$Port = 22,
  [string]$RemotePath = "/root/powerfulcontrolsystem",
  [string]$IdentityFile = "",
  [string]$SshHostKey = "",
  [string]$TargetUser = "root",
  [string]$TargetHost = "",
  [int]$TargetPort = 22,
  [string]$TargetIdentityFile = "",
  [string]$TargetRemotePath = "/root/powerfulcontrolsystem",
  [string]$TargetSshHostKey = "",
  [switch]$NoGui
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$localConfig = Join-Path $PSScriptRoot "pcs_deployment.local.ps1"
if (Test-Path -LiteralPath $localConfig) {
  . $localConfig
  if (-not $PSBoundParameters.ContainsKey("RemoteHost") -and (Get-Variable PcsVpsHost -Scope Script -ErrorAction SilentlyContinue) -and -not [string]::IsNullOrWhiteSpace($script:PcsVpsHost)) { $RemoteHost = $script:PcsVpsHost.Trim() }
  if (-not $PSBoundParameters.ContainsKey("RemoteUser") -and (Get-Variable PcsVpsUser -Scope Script -ErrorAction SilentlyContinue) -and -not [string]::IsNullOrWhiteSpace($script:PcsVpsUser)) { $RemoteUser = $script:PcsVpsUser.Trim() }
  if (-not $PSBoundParameters.ContainsKey("RemotePath") -and (Get-Variable PcsVpsRemotePath -Scope Script -ErrorAction SilentlyContinue) -and -not [string]::IsNullOrWhiteSpace($script:PcsVpsRemotePath)) { $RemotePath = $script:PcsVpsRemotePath.Trim() }
  if (-not $PSBoundParameters.ContainsKey("Port") -and (Get-Variable PcsVpsPort -Scope Script -ErrorAction SilentlyContinue) -and $null -ne $script:PcsVpsPort) { $Port = [int]$script:PcsVpsPort }
  if (-not $PSBoundParameters.ContainsKey("SshHostKey") -and (Get-Variable PcsVpsHostKey -Scope Script -ErrorAction SilentlyContinue) -and -not [string]::IsNullOrWhiteSpace($script:PcsVpsHostKey)) { $SshHostKey = $script:PcsVpsHostKey.Trim() }
  if (-not $PSBoundParameters.ContainsKey("IdentityFile") -and (Get-Variable PcsVpsIdentityFile -Scope Script -ErrorAction SilentlyContinue) -and -not [string]::IsNullOrWhiteSpace($script:PcsVpsIdentityFile)) { $IdentityFile = $script:PcsVpsIdentityFile.Trim() }
}

function Resolve-ToolPath {
  param(
    [Parameter(Mandatory=$true)][string]$Name,
    [Parameter(Mandatory=$true)][string[]]$Candidates
  )
  $cmd = Get-Command $Name -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source -ErrorAction SilentlyContinue
  if ($cmd) { return $cmd }
  foreach ($candidate in $Candidates) {
    if (Test-Path -LiteralPath $candidate) { return $candidate }
  }
  throw "No se encontro $Name. Instala PuTTY o agregalo al PATH."
}

function Resolve-Plink { Resolve-ToolPath -Name "plink.exe" -Candidates @("D:\Program Files\PuTTY\plink.exe", "C:\Program Files\PuTTY\plink.exe", "C:\Program Files (x86)\PuTTY\plink.exe") }
function Resolve-Pscp { Resolve-ToolPath -Name "pscp.exe" -Candidates @("D:\Program Files\PuTTY\pscp.exe", "C:\Program Files\PuTTY\pscp.exe", "C:\Program Files (x86)\PuTTY\pscp.exe") }

function Resolve-Identity {
  param([string]$Value)
  if (-not [string]::IsNullOrWhiteSpace($Value)) {
    if ([System.IO.Path]::IsPathRooted($Value) -and (Test-Path -LiteralPath $Value)) { return (Resolve-Path $Value).Path }
    $repoCandidate = Join-Path $repoRoot $Value
    if (Test-Path -LiteralPath $repoCandidate) { return (Resolve-Path $repoCandidate).Path }
  }
  $default = Join-Path $repoRoot "clave privada ssh.ppk"
  if (Test-Path -LiteralPath $default) { return (Resolve-Path $default).Path }
  throw "No se encontro IdentityFile. Configuralo en scripts/pcs_deployment.local.ps1 o usa -IdentityFile."
}

function Convert-ToBashLiteral {
  param([AllowNull()][AllowEmptyString()][string]$Value = "")
  if ($null -eq $Value) { $Value = "" }
  return "'" + $Value.Replace("'", "'\''") + "'"
}

function New-ProgressUi {
  param([switch]$Disabled)
  $ctx = [ordered]@{ Enabled = $false; Form = $null; Bar = $null; Label = $null; Log = $null }
  if ($Disabled) { return [pscustomobject]$ctx }
  try {
    Add-Type -AssemblyName System.Windows.Forms
    Add-Type -AssemblyName System.Drawing
    $form = New-Object System.Windows.Forms.Form
    $form.Text = "Backup VPS PCS"
    $form.Width = 620
    $form.Height = 300
    $form.StartPosition = "CenterScreen"
    $form.TopMost = $false

    $label = New-Object System.Windows.Forms.Label
    $label.Left = 16
    $label.Top = 16
    $label.Width = 570
    $label.Height = 26
    $label.Text = "Preparando..."

    $bar = New-Object System.Windows.Forms.ProgressBar
    $bar.Left = 16
    $bar.Top = 50
    $bar.Width = 570
    $bar.Height = 24
    $bar.Minimum = 0
    $bar.Maximum = 100

    $log = New-Object System.Windows.Forms.TextBox
    $log.Left = 16
    $log.Top = 88
    $log.Width = 570
    $log.Height = 155
    $log.Multiline = $true
    $log.ScrollBars = "Vertical"
    $log.ReadOnly = $true

    $form.Controls.Add($label)
    $form.Controls.Add($bar)
    $form.Controls.Add($log)
    $form.Show()
    [System.Windows.Forms.Application]::DoEvents()
    $ctx.Enabled = $true
    $ctx.Form = $form
    $ctx.Bar = $bar
    $ctx.Label = $label
    $ctx.Log = $log
  } catch {
    $ctx.Enabled = $false
  }
  return [pscustomobject]$ctx
}

function Set-BackupProgress {
  param(
    [Parameter(Mandatory=$true)]$Ui,
    [int]$Percent,
    [string]$Message
  )
  $Percent = [Math]::Max(0, [Math]::Min(100, $Percent))
  Write-Host ("[{0,3}%] {1}" -f $Percent, $Message)
  if ($Ui.Enabled) {
    $Ui.Bar.Value = $Percent
    $Ui.Label.Text = ("{0}% - {1}" -f $Percent, $Message)
    $Ui.Log.AppendText((Get-Date -Format "HH:mm:ss") + "  " + $Message + [Environment]::NewLine)
    [System.Windows.Forms.Application]::DoEvents()
  }
}

function Get-PlinkArgs {
  param([int]$SshPort, [string]$KeyPath, [string]$HostKey)
  $args = @("-batch")
  if (-not [string]::IsNullOrWhiteSpace($HostKey)) { $args += @("-hostkey", $HostKey) }
  $args += @("-P", "$SshPort", "-i", $KeyPath)
  return $args
}

function Invoke-RemoteCommand {
  param(
    [string]$Plink,
    [string[]]$BaseArgs,
    [string]$Target,
    [string]$Command
  )
  & $Plink @BaseArgs $Target $Command
  if ($LASTEXITCODE -ne 0) { throw "Comando remoto fallo con codigo $LASTEXITCODE." }
}

function Invoke-RemoteScript {
  param(
    [string]$Plink,
    [string[]]$BaseArgs,
    [string]$Target,
    [string]$ScriptText
  )
  $tmp = Join-Path $env:TEMP ("pcs_backup_remote_" + (Get-Date -Format "yyyyMMdd_HHmmss") + ".sh")
  $utf8NoBom = New-Object System.Text.UTF8Encoding($false)
  [System.IO.File]::WriteAllText($tmp, $ScriptText, $utf8NoBom)
  try {
    & $Plink @BaseArgs "-m" $tmp $Target
    if ($LASTEXITCODE -ne 0) { throw "Script remoto fallo con codigo $LASTEXITCODE." }
  } finally {
    Remove-Item -LiteralPath $tmp -Force -ErrorAction SilentlyContinue
  }
}

function New-RemoteBackupScript {
  param([string]$RemotePath, [string]$Stamp)
  $remotePathLit = Convert-ToBashLiteral $RemotePath
  $stampLit = Convert-ToBashLiteral $Stamp
  $template = @'
set -e
remote_path=__REMOTE_PATH__
stamp=__STAMP__
work="/tmp/pcs_vps_backup_$stamp"
archive="/tmp/pcs_vps_full_backup_$stamp.tar.gz"
restore_script="$work/restore_to_new_vps.sh"

rm -rf "$work" "$archive"
mkdir -p "$work"/{inventory,volumes,images,project,database}
chmod 700 "$work"

echo "[INFO] Inventario Docker y sistema"
date -Is > "$work/manifest.txt"
uname -a >> "$work/manifest.txt" 2>/dev/null || true
docker ps -a > "$work/inventory/docker_ps.txt" 2>&1 || true
docker images > "$work/inventory/docker_images.txt" 2>&1 || true
docker volume ls > "$work/inventory/docker_volumes.txt" 2>&1 || true
docker network ls > "$work/inventory/docker_networks.txt" 2>&1 || true
df -h > "$work/inventory/df_h.txt" 2>&1 || true
du -sh "$remote_path" > "$work/inventory/project_size.txt" 2>&1 || true
if [ -d "$remote_path" ]; then
  (cd "$remote_path" && find deploy -maxdepth 3 -type f 2>/dev/null | sort > "$work/inventory/deploy_files.txt" || true)
fi

echo "[INFO] Dump logico PostgreSQL"
if docker ps --format '{{.Names}}' | grep -qx 'pcs-postgres'; then
  docker exec pcs-postgres sh -lc 'pg_dumpall -U "$POSTGRES_USER"' > "$work/database/postgres_all.sql"
  gzip -9 "$work/database/postgres_all.sql"
else
  echo "[WARN] pcs-postgres no esta activo" > "$work/database/postgres_all.missing.txt"
fi

echo "[INFO] Guardando imagenes locales PCS"
images="$(docker images --format '{{.Repository}}:{{.Tag}}' | grep -E '^(pcs-backend|pcs-frontend):' || true)"
if [ -n "$images" ]; then
  docker save $images | gzip -9 > "$work/images/pcs_local_images.tar.gz"
else
  echo "[WARN] No se encontraron imagenes pcs-backend/pcs-frontend" > "$work/images/pcs_local_images.missing.txt"
fi

echo "[INFO] Empaquetando volumenes Docker"
docker volume ls --format '{{.Name}}' | sort > "$work/inventory/docker_volume_names.txt"
while IFS= read -r volume; do
  case "$volume" in
    powerful-control-system_*|pcs-*|pcs_*)
      safe="$(printf '%s' "$volume" | tr -c 'A-Za-z0-9_.-' '_')"
      echo "[INFO] Volumen $volume"
      docker run --rm -v "$volume:/volume:ro" -v "$work/volumes:/backup" alpine:3.20 sh -lc "apk add --no-cache tar >/dev/null && cd /volume && tar --warning=no-file-ignored --ignore-failed-read -czf /backup/$safe.tar.gz ."
      ;;
  esac
done < "$work/inventory/docker_volume_names.txt"

echo "[INFO] Empaquetando proyecto remoto"
if [ -d "$remote_path" ]; then
  parent="$(dirname "$remote_path")"
  base="$(basename "$remote_path")"
  tar -czf "$work/project/project_files.tar.gz" \
    --exclude="$base/.git" \
    --exclude="$base/backups/vps-snapshots" \
    --exclude="$base/.gotmp" \
    --exclude="$base/node_modules" \
    -C "$parent" "$base"
fi

echo "[INFO] Creando restaurador"
cat > "$restore_script" <<'RESTORE'
#!/usr/bin/env bash
set -euo pipefail
if [ "${PCS_RESTORE_CONFIRM:-}" != "RESTORE_PCS_VPS" ]; then
  echo "Set PCS_RESTORE_CONFIRM=RESTORE_PCS_VPS para ejecutar una restauracion real."
  echo "Ejemplo: PCS_RESTORE_CONFIRM=RESTORE_PCS_VPS bash restore_to_new_vps.sh"
  exit 2
fi
backup_dir="$(cd "$(dirname "$0")" && pwd)"
target_path="${PCS_TARGET_PATH:-/root/powerfulcontrolsystem}"

echo "[RESTORE] Instalando dependencias base si hacen falta"
if command -v apt-get >/dev/null 2>&1; then
  apt-get update -y
  apt-get install -y docker.io docker-compose-plugin gzip tar
fi
systemctl enable --now docker >/dev/null 2>&1 || true

echo "[RESTORE] Restaurando archivos del proyecto en $target_path"
mkdir -p "$(dirname "$target_path")"
if [ -f "$backup_dir/project/project_files.tar.gz" ]; then
  tar -xzf "$backup_dir/project/project_files.tar.gz" -C "$(dirname "$target_path")"
fi

echo "[RESTORE] Cargando imagenes PCS si existen"
if [ -f "$backup_dir/images/pcs_local_images.tar.gz" ]; then
  gunzip -c "$backup_dir/images/pcs_local_images.tar.gz" | docker load
fi

echo "[RESTORE] Restaurando volumenes Docker"
for artifact in "$backup_dir"/volumes/*.tar.gz; do
  [ -f "$artifact" ] || continue
  name="$(basename "$artifact" .tar.gz)"
  docker volume create "$name" >/dev/null
  docker run --rm -v "$name:/volume" -v "$backup_dir/volumes:/backup:ro" alpine:3.20 sh -lc "cd /volume && tar -xzf /backup/$(basename "$artifact")"
done

echo "[RESTORE] Levantando stack Docker"
cd "$target_path"
if [ -f deploy/docker-compose.platform.yml ]; then
  docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml --profile mail up -d || \
  docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml up -d
fi

echo "[RESTORE] Restaurando dump PostgreSQL si esta disponible"
if [ -f "$backup_dir/database/postgres_all.sql.gz" ]; then
  for i in $(seq 1 60); do
    if docker exec pcs-postgres pg_isready -U postgres >/dev/null 2>&1 || docker exec pcs-postgres sh -lc 'pg_isready -U "$POSTGRES_USER"' >/dev/null 2>&1; then
      break
    fi
    sleep 2
  done
  gunzip -c "$backup_dir/database/postgres_all.sql.gz" | docker exec -i pcs-postgres sh -lc 'psql -U "$POSTGRES_USER"'
fi

echo "[RESTORE] Restauracion finalizada. Revisa DNS, firewall, .env y certificados antes de produccion."
RESTORE
chmod +x "$restore_script"

echo "[INFO] Creando paquete final"
tar -czf "$archive" -C "$(dirname "$work")" "$(basename "$work")"
sha256sum "$archive" > "$archive.sha256"
du -h "$archive" | tee "$work/package_size.txt"
echo "[OK] BACKUP_ARCHIVE=$archive"
'@
  return $template.Replace("__REMOTE_PATH__", $remotePathLit).Replace("__STAMP__", $stampLit)
}

function Invoke-BackupMode {
  $ui = New-ProgressUi -Disabled:$NoGui
  $stamp = Get-Date -Format "yyyyMMdd_HHmmss"
  $identityResolved = Resolve-Identity $IdentityFile
  $plink = Resolve-Plink
  $pscp = Resolve-Pscp
  $target = "$RemoteUser@$RemoteHost"
  $baseArgs = Get-PlinkArgs -SshPort $Port -KeyPath $identityResolved -HostKey $SshHostKey
  $pscpArgs = @("-batch")
  if (-not [string]::IsNullOrWhiteSpace($SshHostKey)) { $pscpArgs += @("-hostkey", $SshHostKey) }
  $pscpArgs += @("-P", "$Port", "-i", $identityResolved)

  New-Item -ItemType Directory -Path $BackupRoot -Force | Out-Null
  $backupDir = Join-Path $BackupRoot ("PCS_VPS_backup_" + $stamp)
  New-Item -ItemType Directory -Path $backupDir -Force | Out-Null
  $logPath = Join-Path $backupDir "backup_log.txt"

  try {
    Set-BackupProgress $ui 5 "Carpeta local lista: $backupDir"
    Set-BackupProgress $ui 10 "Verificando conexion SSH al VPS"
    Invoke-RemoteCommand -Plink $plink -BaseArgs $baseArgs -Target $target -Command "echo PCS_BACKUP_SSH_OK" | Tee-Object -FilePath $logPath -Append | Out-Null

    Set-BackupProgress $ui 20 "Creando snapshot remoto Docker/PostgreSQL"
    $remoteScript = New-RemoteBackupScript -RemotePath $RemotePath -Stamp $stamp
    Invoke-RemoteScript -Plink $plink -BaseArgs $baseArgs -Target $target -ScriptText $remoteScript 2>&1 | Tee-Object -FilePath $logPath -Append

    $remoteArchive = "/tmp/pcs_vps_full_backup_$stamp.tar.gz"
    $remoteSha = "$remoteArchive.sha256"
    $localArchive = Join-Path $backupDir ("pcs_vps_full_backup_$stamp.tar.gz")
    $localSha = Join-Path $backupDir ("pcs_vps_full_backup_$stamp.tar.gz.sha256")

    Set-BackupProgress $ui 65 "Descargando paquete del VPS"
    & $pscp @pscpArgs "${target}:$remoteArchive" $localArchive 2>&1 | Tee-Object -FilePath $logPath -Append
    if ($LASTEXITCODE -ne 0) { throw "No se pudo descargar el paquete principal." }
    & $pscp @pscpArgs "${target}:$remoteSha" $localSha 2>&1 | Tee-Object -FilePath $logPath -Append
    if ($LASTEXITCODE -ne 0) { throw "No se pudo descargar el hash SHA256." }

    Set-BackupProgress $ui 80 "Verificando paquete local"
    $tarList = & tar -tzf $localArchive 2>&1
    if ($LASTEXITCODE -ne 0) { throw "El paquete descargado no abre como tar.gz: $tarList" }
    $size = (Get-Item -LiteralPath $localArchive).Length
    $manifest = [ordered]@{
      created_at = (Get-Date).ToString("s")
      remote = "$RemoteUser@$RemoteHost`:$RemotePath"
      port = $Port
      backup_dir = $backupDir
      archive = $localArchive
      archive_bytes = $size
      sha256_file = $localSha
      restore_mode = "Use -Restore -BackupPath `"$localArchive`" -TargetHost NUEVO_VPS"
      notes = "Contiene dump PostgreSQL, volumenes Docker PCS, imagenes PCS locales, inventario y restaurador Linux."
    }
    $manifestPath = Join-Path $backupDir "manifest.json"
    ($manifest | ConvertTo-Json -Depth 4) | Set-Content -LiteralPath $manifestPath -Encoding UTF8

    Set-BackupProgress $ui 92 "Limpiando temporales remotos"
    Invoke-RemoteCommand -Plink $plink -BaseArgs $baseArgs -Target $target -Command "rm -rf /tmp/pcs_vps_backup_$stamp /tmp/pcs_vps_full_backup_$stamp.tar.gz /tmp/pcs_vps_full_backup_$stamp.tar.gz.sha256" | Tee-Object -FilePath $logPath -Append | Out-Null

    Set-BackupProgress $ui 100 "Backup completado"
    Write-Host "[OK] Backup local: $localArchive"
    Write-Host "[OK] Manifest: $manifestPath"
  } finally {
    if ($ui.Enabled) {
      $ui.Label.Text = "Proceso terminado. Puedes cerrar esta ventana."
      [System.Windows.Forms.Application]::DoEvents()
    }
  }
}

function Invoke-RestoreUploadMode {
  $ui = New-ProgressUi -Disabled:$NoGui
  if ([string]::IsNullOrWhiteSpace($TargetHost)) { throw "Para -Restore debes indicar -TargetHost del VPS nuevo." }
  if ([string]::IsNullOrWhiteSpace($BackupPath)) {
    $latest = Get-ChildItem -LiteralPath $BackupRoot -Recurse -Filter "pcs_vps_full_backup_*.tar.gz" -ErrorAction SilentlyContinue | Sort-Object LastWriteTime -Descending | Select-Object -First 1
    if ($latest) { $BackupPath = $latest.FullName }
  }
  if ([string]::IsNullOrWhiteSpace($BackupPath) -or -not (Test-Path -LiteralPath $BackupPath)) { throw "No se encontro BackupPath." }
  $targetIdentity = if ([string]::IsNullOrWhiteSpace($TargetIdentityFile)) { Resolve-Identity $IdentityFile } else { Resolve-Identity $TargetIdentityFile }
  $plink = Resolve-Plink
  $pscp = Resolve-Pscp
  $target = "$TargetUser@$TargetHost"
  $baseArgs = Get-PlinkArgs -SshPort $TargetPort -KeyPath $targetIdentity -HostKey $TargetSshHostKey
  $pscpArgs = @("-batch")
  if (-not [string]::IsNullOrWhiteSpace($TargetSshHostKey)) { $pscpArgs += @("-hostkey", $TargetSshHostKey) }
  $pscpArgs += @("-P", "$TargetPort", "-i", $targetIdentity)
  $remoteIncoming = "/root/pcs_restore_incoming"
  $remoteArchive = "$remoteIncoming/" + [System.IO.Path]::GetFileName($BackupPath)

  Set-BackupProgress $ui 10 "Preparando VPS nuevo"
  Invoke-RemoteCommand -Plink $plink -BaseArgs $baseArgs -Target $target -Command "mkdir -p '$remoteIncoming' && chmod 700 '$remoteIncoming'"
  Set-BackupProgress $ui 35 "Subiendo paquete al VPS nuevo"
  & $pscp @pscpArgs $BackupPath "${target}:$remoteArchive"
  if ($LASTEXITCODE -ne 0) { throw "No se pudo subir el backup al VPS nuevo." }
  Set-BackupProgress $ui 65 "Extrayendo paquete en VPS nuevo"
  Invoke-RemoteCommand -Plink $plink -BaseArgs $baseArgs -Target $target -Command "cd '$remoteIncoming' && tar -xzf '$remoteArchive'"
  Set-BackupProgress $ui 80 "Ubicando script restaurador"
  $remoteRestoreCmd = "cd '$remoteIncoming' && restore_dir=`$(find . -maxdepth 1 -type d -name 'pcs_vps_backup_*' | sort | tail -n 1) && test -n `"`$restore_dir`" && echo `$restore_dir/restore_to_new_vps.sh"
  Invoke-RemoteCommand -Plink $plink -BaseArgs $baseArgs -Target $target -Command $remoteRestoreCmd
  if ($ExecuteRemoteRestore) {
    Set-BackupProgress $ui 90 "Ejecutando restauracion remota confirmada"
    $cmd = "cd '$remoteIncoming' && restore_dir=`$(find . -maxdepth 1 -type d -name 'pcs_vps_backup_*' | sort | tail -n 1) && PCS_RESTORE_CONFIRM=RESTORE_PCS_VPS PCS_TARGET_PATH='$TargetRemotePath' bash `"`$restore_dir/restore_to_new_vps.sh`""
    Invoke-RemoteCommand -Plink $plink -BaseArgs $baseArgs -Target $target -Command $cmd
  } else {
    Write-Host "[INFO] Paquete subido. Para ejecutar restauracion real en el VPS nuevo:"
    Write-Host "       cd $remoteIncoming"
    Write-Host "       PCS_RESTORE_CONFIRM=RESTORE_PCS_VPS PCS_TARGET_PATH='$TargetRemotePath' bash ./pcs_vps_backup_*/restore_to_new_vps.sh"
  }
  Set-BackupProgress $ui 100 "Subida/restauracion preparada"
}

if ($Restore) {
  Invoke-RestoreUploadMode
} else {
  Invoke-BackupMode
}
