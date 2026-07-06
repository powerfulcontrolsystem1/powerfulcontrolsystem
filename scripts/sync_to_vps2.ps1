<#
.SYNOPSIS
  Actualiza y asegura servicios basicos del VPS2.

.DESCRIPTION
  Script especifico para el VPS2 local de pruebas. Lee valores desde
  scripts/pcs_deployment.local.ps1 o variables de entorno PCS_VPS2_*.
  No guarda credenciales en el repositorio.

  Operaciones:
  - prueba conectividad SSH;
  - opcionalmente actualiza el repositorio remoto con git pull --ff-only;
  - opcionalmente reinicia el stack Docker si encuentra compose;
  - deshabilita el modo grafico dejando multi-user.target;
  - deja Nextcloud con restart unless-stopped cuando detecta contenedores.
#>

param(
  [switch]$DryRun,
  [switch]$SkipDeploy,
  [switch]$SkipDisableGui,
  [switch]$SkipNextcloud,
  [string]$RemoteHost = "",
  [string]$RemoteUser = "",
  [int]$Port = 0,
  [string]$RemotePath = "",
  [string]$Branch = "",
  [string]$RepoUrl = "",
  [string]$SshHostKey = "",
  [string]$IdentityFile = "",
  [string]$Password = "",
  [switch]$SkipPublishSnapshot,
  [bool]$RestartDockerStack = $true
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Write-Step {
  param([Parameter(Mandatory=$true)][string]$Message)
  Write-Host "[sync_to_vps2] $Message"
}

function Write-Warn {
  param([Parameter(Mandatory=$true)][string]$Message)
  Write-Host "[sync_to_vps2][aviso] $Message" -ForegroundColor Yellow
}

function Convert-ToBashLiteral {
  param([AllowNull()][AllowEmptyString()][string]$Value = "")
  if ($null -eq $Value) { $Value = "" }
  return "'" + $Value.Replace("'", "'\''") + "'"
}

function Resolve-Plink {
  $cmd = Get-Command plink.exe -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source -ErrorAction SilentlyContinue
  if ($cmd) { return $cmd }

  $candidates = @(
    "D:\Program Files\PuTTY\plink.exe",
    "C:\Program Files\PuTTY\plink.exe",
    "C:\Program Files (x86)\PuTTY\plink.exe"
  )

  foreach ($candidate in $candidates) {
    if (Test-Path -LiteralPath $candidate) { return $candidate }
  }

  throw "No se encontro plink.exe. Instala PuTTY o configura OpenSSH con llave y adapta el script."
}

function Resolve-Pscp {
  $cmd = Get-Command pscp.exe -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source -ErrorAction SilentlyContinue
  if ($cmd) { return $cmd }

  $candidates = @(
    "D:\Program Files\PuTTY\pscp.exe",
    "C:\Program Files\PuTTY\pscp.exe",
    "C:\Program Files (x86)\PuTTY\pscp.exe"
  )

  foreach ($candidate in $candidates) {
    if (Test-Path -LiteralPath $candidate) { return $candidate }
  }

  throw "No se encontro pscp.exe. Instala PuTTY o configura publicacion manual del snapshot."
}

function Resolve-ConfigValue {
  param(
    [Parameter(Mandatory=$true)][string]$Name,
    [AllowNull()][AllowEmptyString()][string]$CurrentValue,
    [AllowNull()][AllowEmptyString()][string]$ScriptVariableName,
    [AllowNull()][AllowEmptyString()][string]$EnvironmentName,
    [AllowNull()][AllowEmptyString()][string]$DefaultValue = ""
  )

  if (-not [string]::IsNullOrWhiteSpace($CurrentValue)) { return $CurrentValue.Trim() }
  if (-not [string]::IsNullOrWhiteSpace($ScriptVariableName)) {
    $localVar = Get-Variable -Name $ScriptVariableName -Scope Script -ErrorAction SilentlyContinue
    if ($localVar -and -not [string]::IsNullOrWhiteSpace([string]$localVar.Value)) {
      return ([string]$localVar.Value).Trim()
    }
  }
  if (-not [string]::IsNullOrWhiteSpace($EnvironmentName)) {
    $envValue = [Environment]::GetEnvironmentVariable($EnvironmentName)
    if (-not [string]::IsNullOrWhiteSpace($envValue)) { return $envValue.Trim() }
  }
  return $DefaultValue
}

function Resolve-ConfigInt {
  param(
    [int]$CurrentValue,
    [AllowNull()][AllowEmptyString()][string]$ScriptVariableName,
    [AllowNull()][AllowEmptyString()][string]$EnvironmentName,
    [int]$DefaultValue
  )

  if ($CurrentValue -gt 0) { return $CurrentValue }
  if (-not [string]::IsNullOrWhiteSpace($ScriptVariableName)) {
    $localVar = Get-Variable -Name $ScriptVariableName -Scope Script -ErrorAction SilentlyContinue
    if ($localVar -and $null -ne $localVar.Value) { return [int]$localVar.Value }
  }
  if (-not [string]::IsNullOrWhiteSpace($EnvironmentName)) {
    $envValue = [Environment]::GetEnvironmentVariable($EnvironmentName)
    if (-not [string]::IsNullOrWhiteSpace($envValue)) { return [int]$envValue }
  }
  return $DefaultValue
}

function Invoke-Vps2Ssh {
  param(
    [Parameter(Mandatory=$true)][string]$Command,
    [switch]$AllowFailure
  )

  $plink = Resolve-Plink
  $target = "$script:ResolvedRemoteUser@$script:ResolvedRemoteHost"
  $args = @("-batch", "-ssh", "-P", [string]$script:ResolvedPort)
  if (-not [string]::IsNullOrWhiteSpace($script:ResolvedSshHostKey)) {
    $args += @("-hostkey", $script:ResolvedSshHostKey)
  }
  if (-not [string]::IsNullOrWhiteSpace($script:ResolvedIdentityFile)) {
    $args += @("-i", $script:ResolvedIdentityFile)
  } elseif (-not [string]::IsNullOrWhiteSpace($script:ResolvedPassword)) {
    $args += @("-pw", $script:ResolvedPassword)
  } else {
    throw "No hay IdentityFile ni Password para autenticar en VPS2. Configura scripts/pcs_deployment.local.ps1 o PCS_VPS2_PASSWORD."
  }

  $args += @($target, $Command)
  if ($DryRun) {
    Write-Step "DryRun: se omitio SSH remoto a $target."
    return
  }

  & $plink @args
  $code = $LASTEXITCODE
  if ($code -ne 0 -and -not $AllowFailure) {
    throw "Comando SSH en VPS2 fallo con codigo $code."
  }
}

function Get-CurrentGitBranch {
  $branch = ""
  try {
    $branch = (& git rev-parse --abbrev-ref HEAD 2>$null | Select-Object -First 1).Trim()
  } catch {
    $branch = ""
  }
  if ([string]::IsNullOrWhiteSpace($branch) -or $branch -eq "HEAD") {
    return "main"
  }
  return $branch
}

function New-VPS2StatusSnapshot {
  $statusCommand = @'
printf 'hostname=%s\n' "$(hostname)";
printf 'uptime=%s\n' "$(uptime -p 2>/dev/null || uptime)";
printf 'kernel=%s\n' "$(uname -srmo)";
printf 'default_target=%s\n' "$(systemctl get-default 2>/dev/null || true)";
printf 'load=%s\n' "$(cat /proc/loadavg 2>/dev/null | awk '{print $1" "$2" "$3}')";
printf 'cpu_cores=%s\n' "$(nproc 2>/dev/null || echo 0)";
temp="$(for f in /sys/class/thermal/thermal_zone*/temp; do [ -r "$f" ] && awk '{printf "%.1f", $1/1000}' "$f" && break; done 2>/dev/null)";
printf 'temperature_c=%s\n' "$temp";
printf 'memory=%s\n' "$(free -b 2>/dev/null | awk '/Mem:/ {print $2" "$3" "$7}')";
printf 'disk_root=%s\n' "$(df -PB1 / 2>/dev/null | awk 'NR==2 {print $2" "$3" "$4" "$5}')";
printf 'service_docker=%s\n' "$(systemctl is-active docker 2>/dev/null || true)";
printf 'service_ssh=%s\n' "$(systemctl is-active ssh 2>/dev/null || systemctl is-active sshd 2>/dev/null || true)";
printf 'docker_version=%s\n' "$(docker --version 2>/dev/null || true)";
printf 'docker_total=%s\n' "$(docker ps -a -q 2>/dev/null | wc -l | tr -d " ")";
printf 'docker_running=%s\n' "$(docker ps -q 2>/dev/null | wc -l | tr -d " ")";
ip -4 -o addr show scope global 2>/dev/null | awk '{print "ip="$2" "$4}';
docker ps -a --format '{{.Names}}|{{.Status}}|{{.Ports}}' 2>/dev/null | grep -Ei '(^|[-_])(nextcloud|cloud)([-_]|$)|nextcloud' | sed 's/^/nextcloud=/' || true
'@
  $out = Invoke-Vps2Ssh -Command $statusCommand
  $kv = @{}
  foreach ($line in ($out -split "`r?`n")) {
    if ($line -notmatch '=') { continue }
    $parts = $line.Split('=', 2)
    $key = $parts[0].Trim()
    $value = $parts[1].Trim()
    if (-not $kv.ContainsKey($key)) { $kv[$key] = @() }
    $kv[$key] += $value
  }
  function FirstValue([string]$key) {
    if ($kv.ContainsKey($key) -and $kv[$key].Count -gt 0) { return [string]$kv[$key][0] }
    return ""
  }
  function ParseBytesTriple([string]$raw) {
    $p = @($raw -split '\s+' | Where-Object { $_ })
    return @{
      total = if ($p.Count -gt 0) { [int64]$p[0] } else { 0 }
      used = if ($p.Count -gt 1) { [int64]$p[1] } else { 0 }
      available = if ($p.Count -gt 2) { [int64]$p[2] } else { 0 }
    }
  }
  function ParseDisk([string]$raw) {
    $p = @($raw -split '\s+' | Where-Object { $_ })
    return @{
      total = if ($p.Count -gt 0) { [int64]$p[0] } else { 0 }
      used = if ($p.Count -gt 1) { [int64]$p[1] } else { 0 }
      available = if ($p.Count -gt 2) { [int64]$p[2] } else { 0 }
      percent = if ($p.Count -gt 3) { [string]$p[3] } else { "" }
    }
  }
  $cloud = @()
  if ($kv.ContainsKey("nextcloud")) {
    foreach ($row in $kv["nextcloud"]) {
      $p = $row.Split('|', 3)
      $cloud += @{
        name = if ($p.Count -gt 0) { $p[0] } else { "" }
        status = if ($p.Count -gt 1) { $p[1] } else { "" }
        ports = if ($p.Count -gt 2) { $p[2] } else { "" }
      }
    }
  }
  return @{
    ok = $true
    checked_at = (Get-Date).ToString("o")
    config = @{
      host = $script:ResolvedRemoteHost
      port = $script:ResolvedPort
      user = $script:ResolvedRemoteUser
      remote_path = $script:ResolvedRemotePath
      host_key = $script:ResolvedSshHostKey
      has_password = -not [string]::IsNullOrWhiteSpace($script:ResolvedPassword)
    }
    reachable = @{ ssh = $true; vnc = $false }
    system = @{
      hostname = FirstValue "hostname"
      uptime = FirstValue "uptime"
      kernel = FirstValue "kernel"
      default_target = FirstValue "default_target"
      ip = (($kv["ip"] | Where-Object { $_ }) -join ", ")
    }
    resources = @{
      load = FirstValue "load"
      cpu_cores = FirstValue "cpu_cores"
      temperature_c = FirstValue "temperature_c"
      memory = ParseBytesTriple (FirstValue "memory")
      disk_root = ParseDisk (FirstValue "disk_root")
    }
    docker = @{
      version = FirstValue "docker_version"
      containers_total = FirstValue "docker_total"
      containers_running = FirstValue "docker_running"
    }
    services = @{
      docker = FirstValue "service_docker"
      ssh = FirstValue "service_ssh"
    }
    nextcloud = $cloud
    last_message = "Snapshot publicado por sync_to_vps2."
  }
}

function Publish-VPS2SnapshotToMainVPS {
  param([Parameter(Mandatory=$true)]$Snapshot)
  $mainHost = Resolve-ConfigValue -Name "PcsVpsHost" -CurrentValue "" -ScriptVariableName "PcsVpsHost" -EnvironmentName "PCS_VPS_HOST" -DefaultValue ""
  $mainUser = Resolve-ConfigValue -Name "PcsVpsUser" -CurrentValue "" -ScriptVariableName "PcsVpsUser" -EnvironmentName "PCS_VPS_USER" -DefaultValue "root"
  $mainPort = Resolve-ConfigInt -CurrentValue 0 -ScriptVariableName "PcsVpsPort" -EnvironmentName "PCS_VPS_PORT" -DefaultValue 22
  $mainHostKey = Resolve-ConfigValue -Name "PcsVpsHostKey" -CurrentValue "" -ScriptVariableName "PcsVpsHostKey" -EnvironmentName "PCS_VPS_HOSTKEY" -DefaultValue ""
  $mainIdentity = Resolve-ConfigValue -Name "PcsVpsIdentityFile" -CurrentValue "" -ScriptVariableName "PcsVpsIdentityFile" -EnvironmentName "PCS_VPS_IDENTITY_FILE" -DefaultValue ""
  if ([string]::IsNullOrWhiteSpace($mainIdentity)) {
    $candidate = Join-Path (Split-Path -Parent $PSScriptRoot) "clave privada ssh.ppk"
    if (Test-Path -LiteralPath $candidate) { $mainIdentity = $candidate }
  }
  if ([string]::IsNullOrWhiteSpace($mainHost) -or [string]::IsNullOrWhiteSpace($mainIdentity)) {
    Write-Warn "Snapshot VPS2 no publicado: falta host o IdentityFile del VPS principal."
    return
  }
  $tmp = Join-Path ([System.IO.Path]::GetTempPath()) "pcs_vps2_status.json"
  $Snapshot | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $tmp -Encoding UTF8
  $plink = Resolve-Plink
  $pscp = Resolve-Pscp
  $target = "$mainUser@$mainHost"
  $baseArgs = @("-batch")
  if (-not [string]::IsNullOrWhiteSpace($mainHostKey)) { $baseArgs += @("-hostkey", $mainHostKey) }
  $baseArgs += @("-P", [string]$mainPort, "-i", $mainIdentity)
  & $pscp @($baseArgs + @($tmp, "${target}:/tmp/pcs_vps2_status.json"))
  if ($LASTEXITCODE -ne 0) { throw "No se pudo subir snapshot VPS2 al VPS principal." }
  $remoteCmd = "mkdir -p /root/powerfulcontrolsystem/backup && cp /tmp/pcs_vps2_status.json /root/powerfulcontrolsystem/backup/vps2_status.json && if docker ps --format '{{.Names}}' | grep -qx pcs-backend; then docker cp /tmp/pcs_vps2_status.json pcs-backend:/app/backup/vps2_status.json; fi && rm -f /tmp/pcs_vps2_status.json"
  & $plink @($baseArgs + @($target, $remoteCmd))
  if ($LASTEXITCODE -ne 0) { throw "No se pudo instalar snapshot VPS2 en el VPS principal." }
  Write-Step "Snapshot VPS2 publicado en el VPS principal."
}

$localConfig = Join-Path $PSScriptRoot "pcs_deployment.local.ps1"
if (Test-Path -LiteralPath $localConfig) {
  . $localConfig
}

$script:ResolvedRemoteHost = Resolve-ConfigValue -Name "RemoteHost" -CurrentValue $RemoteHost -ScriptVariableName "PcsVps2Host" -EnvironmentName "PCS_VPS2_HOST" -DefaultValue "192.168.1.188"
$script:ResolvedRemoteUser = Resolve-ConfigValue -Name "RemoteUser" -CurrentValue $RemoteUser -ScriptVariableName "PcsVps2User" -EnvironmentName "PCS_VPS2_USER" -DefaultValue "admin1"
$script:ResolvedPort = Resolve-ConfigInt -CurrentValue $Port -ScriptVariableName "PcsVps2Port" -EnvironmentName "PCS_VPS2_PORT" -DefaultValue 22
$script:ResolvedRemotePath = Resolve-ConfigValue -Name "RemotePath" -CurrentValue $RemotePath -ScriptVariableName "PcsVps2RemotePath" -EnvironmentName "PCS_VPS2_REMOTE_PATH" -DefaultValue "/home/admin1/powerfulcontrolsystem"
$script:ResolvedRepoUrl = Resolve-ConfigValue -Name "RepoUrl" -CurrentValue $RepoUrl -ScriptVariableName "PcsVps2RepoUrl" -EnvironmentName "PCS_VPS2_REPO_URL" -DefaultValue ""
if ([string]::IsNullOrWhiteSpace($script:ResolvedRepoUrl)) {
  $script:ResolvedRepoUrl = Resolve-ConfigValue -Name "RepoUrl" -CurrentValue "" -ScriptVariableName "PcsGitRemoteUrl" -EnvironmentName "PCS_REPO_URL" -DefaultValue ""
}
$script:ResolvedSshHostKey = Resolve-ConfigValue -Name "SshHostKey" -CurrentValue $SshHostKey -ScriptVariableName "PcsVps2HostKey" -EnvironmentName "PCS_VPS2_HOSTKEY" -DefaultValue "SHA256:QQmT0ZjCVNNxw7ICwV7FKwrzzzfWrOrtZ9zTrEGkwH0"
$script:ResolvedIdentityFile = Resolve-ConfigValue -Name "IdentityFile" -CurrentValue $IdentityFile -ScriptVariableName "PcsVps2IdentityFile" -EnvironmentName "PCS_VPS2_IDENTITY_FILE" -DefaultValue ""
$script:ResolvedPassword = Resolve-ConfigValue -Name "Password" -CurrentValue $Password -ScriptVariableName "PcsVps2Password" -EnvironmentName "PCS_VPS2_PASSWORD" -DefaultValue ""
$resolvedBranch = Resolve-ConfigValue -Name "Branch" -CurrentValue $Branch -ScriptVariableName "PcsVps2Branch" -EnvironmentName "PCS_VPS2_BRANCH" -DefaultValue (Get-CurrentGitBranch)

Write-Step "Destino VPS2: $script:ResolvedRemoteUser@$script:ResolvedRemoteHost`:$script:ResolvedPort ($script:ResolvedRemotePath)."

$tcp = Test-NetConnection -ComputerName $script:ResolvedRemoteHost -Port $script:ResolvedPort -WarningAction SilentlyContinue
if (-not $tcp.TcpTestSucceeded) {
  throw "El puerto SSH $script:ResolvedPort no responde en $script:ResolvedRemoteHost."
}
Write-Step "SSH responde por red."

Invoke-Vps2Ssh -Command "printf 'ssh_ok '; hostname"

if (-not $SkipDeploy) {
  $remotePathLiteral = Convert-ToBashLiteral $script:ResolvedRemotePath
  $branchLiteral = Convert-ToBashLiteral $resolvedBranch
  $repoUrlLiteral = Convert-ToBashLiteral $script:ResolvedRepoUrl
  $restartFlag = if ($RestartDockerStack) { "1" } else { "0" }
  $deployScript = @"
set -eu
REMOTE_PATH=$remotePathLiteral
BRANCH=$branchLiteral
REPO_URL=$repoUrlLiteral
RESTART_DOCKER=$restartFlag
if [ ! -d "`$REMOTE_PATH/.git" ]; then
  if [ -z "`$REPO_URL" ]; then
    echo "repo_no_encontrado=`$REMOTE_PATH"
    echo "repo_url_no_configurada"
    exit 0
  fi
  mkdir -p "`$(dirname "`$REMOTE_PATH")"
  git clone --branch "`$BRANCH" "`$REPO_URL" "`$REMOTE_PATH"
fi
cd "`$REMOTE_PATH"
echo "git_branch=`$BRANCH"
git fetch origin "`$BRANCH" --prune
git pull --ff-only origin "`$BRANCH"
if [ "`$RESTART_DOCKER" = "1" ]; then
  COMPOSE_FILE=""
  for candidate in deploy/docker-compose.platform.yml deploy/docker-compose.yml docker-compose.yml compose.yml; do
    if [ -f "`$candidate" ]; then COMPOSE_FILE="`$candidate"; break; fi
  done
  if [ -n "`$COMPOSE_FILE" ]; then
    if [ -f deploy/.env.platform ]; then
      docker compose --env-file deploy/.env.platform -f "`$COMPOSE_FILE" up -d --build
    else
      docker compose -f "`$COMPOSE_FILE" up -d --build
    fi
    docker ps --format '{{.Names}} {{.Status}}' | head -20
  else
    echo "compose_no_encontrado"
  fi
fi
"@
  $encodedDeploy = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($deployScript))
  Write-Step "Actualizando repositorio remoto por git pull --ff-only."
  Invoke-Vps2Ssh -Command "echo $encodedDeploy | base64 -d | bash"
}

if (-not $SkipDisableGui -or -not $SkipNextcloud) {
  $sudoPrefix = ""
  if (-not [string]::IsNullOrWhiteSpace($script:ResolvedPassword)) {
    $sudoPrefix = "export PCS_SUDO_PASSWORD=$(Convert-ToBashLiteral $script:ResolvedPassword); "
  }

  $maintenanceScript = @"
set -u
run_sudo() {
  if sudo -n true >/dev/null 2>&1; then
    sudo "`$@"
  elif [ -n "`${PCS_SUDO_PASSWORD:-}" ]; then
    printf '%s\n' "`$PCS_SUDO_PASSWORD" | sudo -S -p '' "`$@"
  else
    echo "sudo_no_interactivo_no_disponible"
    return 1
  fi
}

if [ "$([int](-not $SkipDisableGui))" = "1" ]; then
  echo "deshabilitando_modo_grafico"
  run_sudo systemctl set-default multi-user.target
  for svc in display-manager gdm3 lightdm sddm; do
    if systemctl list-unit-files "`$svc.service" >/dev/null 2>&1; then
      run_sudo systemctl disable --now "`$svc.service" >/dev/null 2>&1 || true
    fi
  done
  systemctl get-default
fi

if [ "$([int](-not $SkipNextcloud))" = "1" ]; then
  echo "asegurando_nextcloud"
  nextcloud_names="`$(docker ps -a --format '{{.Names}}' 2>/dev/null | grep -Ei '(^|[-_])(nextcloud|cloud)([-_]|`$)|nextcloud' || true)"
  if [ -n "`$nextcloud_names" ]; then
    echo "`$nextcloud_names" | xargs -r docker update --restart unless-stopped >/dev/null
    echo "`$nextcloud_names" | xargs -r docker start >/dev/null 2>&1 || true
    echo "contenedores_nextcloud=`$(echo "`$nextcloud_names" | tr '\n' ' ')"
  else
    echo "sin_contenedores_nextcloud_detectados"
  fi
  compose_file=""
  for candidate in \
    "`$HOME/powerfulcontrolsystem/deploy/nextcloud/docker-compose.yml" \
    "`$HOME/powerfulcontrolsystem/deploy/nextcloud/compose.yml" \
    "`$HOME/nextcloud/docker-compose.yml" \
    "/opt/nextcloud/docker-compose.yml" \
    "/srv/nextcloud/docker-compose.yml"; do
    if [ -f "`$candidate" ]; then compose_file="`$candidate"; break; fi
  done
  if [ -n "`$compose_file" ]; then
    docker compose -f "`$compose_file" up -d
    echo "compose_nextcloud=`$compose_file"
  else
    echo "sin_compose_nextcloud_detectado"
  fi
  docker ps --format '{{.Names}} {{.Status}}' | grep -Ei 'nextcloud|cloud' || true
fi
"@
  $encodedMaintenance = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($maintenanceScript))
  Write-Step "Aplicando mantenimiento de VPS2."
  Invoke-Vps2Ssh -Command "$sudoPrefix echo $encodedMaintenance | base64 -d | bash"
}

if (-not $SkipPublishSnapshot) {
  Write-Step "Generando snapshot de estado VPS2 para el panel super."
  $snapshot = New-VPS2StatusSnapshot
  Publish-VPS2SnapshotToMainVPS -Snapshot $snapshot
}

Write-Step "Proceso terminado."
