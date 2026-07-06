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

Write-Step "Proceso terminado."
