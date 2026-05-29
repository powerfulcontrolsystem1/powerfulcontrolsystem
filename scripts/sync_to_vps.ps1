<#
.SYNOPSIS
  Wrapper PowerShell para sincronizar con VPS Linux.

.DESCRIPTION
  Prioriza WSL cuando está disponible y usa fallback nativo en Windows
  (OpenSSH ssh/scp para claves OpenSSH, PuTTY plink/pscp para .ppk)
  cuando WSL no está instalado o no tiene distribuciones.
  No programa tareas; se ejecuta manualmente cuando el usuario lo necesite.
  Config opcional: scripts/pcs_deployment.local.ps1 (ver pcs_deployment.local.ps1.example) para
  PcsVpsHost, PcsVpsUser, PcsVpsRemotePath, PcsVpsPort, PcsVpsHostKey, PcsVpsIdentityFile, PcsVpsServerPort, PcsVpsPublicBaseUrl
  cuando no pasas -RemoteHost, etc. en linea de comandos.
#>

param(
  [switch]$DryRun,
  [switch]$PreviewOnly,
  [switch]$SkipBuild,
  [switch]$BuildOnly,
  [string]$LocalPath = "",
  [string]$RemoteUser = "root",
  [string]$RemoteHost = "2.24.197.58",
  [string]$RemotePath = "/root/powerfulcontrolsystem",
  [int]$Port = 22,
  [string]$SshHostKey = "",
  [string]$IdentityFile = "",
  [string]$ExcludeFile = "",
  [bool]$CompressPackage = $true,
  [int]$LargeTransferWarningMB = 200,
  [string]$BuildWorkingDir = "backend",
  [string]$BuildPackage = ".",
  [string]$BuildOutput = "backend/bin/server_linux_amd64",
  [string]$BuildGoOS = "linux",
  [string]$BuildGoArch = "amd64",
  [string]$BuildCgoEnabled = "0",
  [int]$RetryCount = 3,
  [bool]$AutoInstallDependencies = $true,
  [bool]$BootstrapServer = $true,
  [string]$ServerPort = "8080",
  [string]$GoogleClientId = "",
  [string]$GoogleClientSecret = "",
  [string]$GoogleRedirectUrl = "https://powerfulcontrolsystem.com/auth/google/callback",
  [string]$PublicBaseUrl = "https://powerfulcontrolsystem.com/",
  [string]$DbDialect = "postgres",
  [string]$DbEmpresasDsn = "",
  [string]$DbSuperadminDsn = "",
  [ValidateSet("docker", "hybrid", "legacy")]
  [string]$DeploymentMode = "docker",
  [bool]$RestartRemoteServer = $true,
  [string]$RemoteBinaryPath = "backend/bin/server_linux_amd64",
  [string]$RemoteStdoutLogPath = "backend/server.log",
  [string]$RemoteStderrLogPath = "backend/server.err",
  [int]$RestartHealthTimeoutSeconds = 900,
  [bool]$RedeployDockerStack = $true,
  [int]$DockerHealthTimeoutSeconds = 900,
  [bool]$ExcludeEvidenceFromPackage = $true,
  [bool]$OpenPublicUrlAfterDeploy = $true,
  [bool]$CleanupRemoteUnusedFiles = $true,
  [int]$RemoteCleanupTempMinAgeMinutes = 60,
  [int]$RemoteCleanupDockerBuilderCacheMaxAgeHours = 0,
  [bool]$RemoteCleanupDockerDanglingImages = $true,
  [bool]$RemoteCleanupStoppedContainers = $true
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$script:SyncExitCode = 0

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
  if (-not $PSBoundParameters.ContainsKey('SshHostKey') -and (Get-Variable PcsVpsHostKey -Scope Script -ErrorAction SilentlyContinue) -and -not [string]::IsNullOrWhiteSpace($script:PcsVpsHostKey)) {
    $SshHostKey = $script:PcsVpsHostKey.Trim()
  }
  if (-not $PSBoundParameters.ContainsKey('IdentityFile') -and (Get-Variable PcsVpsIdentityFile -Scope Script -ErrorAction SilentlyContinue) -and -not [string]::IsNullOrWhiteSpace($script:PcsVpsIdentityFile)) {
    $IdentityFile = $script:PcsVpsIdentityFile.Trim()
  }
  if (-not $PSBoundParameters.ContainsKey('ServerPort') -and (Get-Variable PcsVpsServerPort -Scope Script -ErrorAction SilentlyContinue) -and -not [string]::IsNullOrWhiteSpace($script:PcsVpsServerPort)) {
    $ServerPort = $script:PcsVpsServerPort.Trim()
  }
  if (-not $PSBoundParameters.ContainsKey('PublicBaseUrl') -and (Get-Variable PcsVpsPublicBaseUrl -Scope Script -ErrorAction SilentlyContinue) -and -not [string]::IsNullOrWhiteSpace($script:PcsVpsPublicBaseUrl)) {
    $PublicBaseUrl = $script:PcsVpsPublicBaseUrl.Trim()
  }
}

function Resolve-PublicDeployUrl {
  param(
    [AllowNull()][AllowEmptyString()][string]$PublicBaseUrl = "",
    [Parameter(Mandatory=$true)][string]$RemoteHost,
    [Parameter(Mandatory=$true)][string]$ServerPort
  )

  $normalizedPublicBaseUrl = ""
  if ($null -ne $PublicBaseUrl) {
    $normalizedPublicBaseUrl = $PublicBaseUrl.Trim()
  }

  if (-not [string]::IsNullOrWhiteSpace($normalizedPublicBaseUrl)) {
    return $normalizedPublicBaseUrl
  }

  return "http://$RemoteHost`:$ServerPort/"
}

function Convert-ToBashLiteral {
  param([AllowNull()][AllowEmptyString()][string]$Value = "")
  if ($null -eq $Value) {
    $Value = ""
  }
  $escaped = $Value.Replace("'", "'\''")
  return "'" + $escaped + "'"
}

function Assert-WslReady {
  if (-not (Get-Command wsl -ErrorAction SilentlyContinue)) {
    throw "WSL no está disponible. Instálalo con: wsl --install -d Ubuntu"
  }

  $distros = & wsl -l -q 2>$null
  $wslCode = $LASTEXITCODE
  $distroText = ($distros -join "").Trim()
  if ($wslCode -ne 0 -or [string]::IsNullOrWhiteSpace($distroText)) {
    throw "WSL está instalado pero sin distribuciones Linux. Ejecuta: wsl --install -d Ubuntu ; luego abre Ubuntu una vez y reintenta."
  }
}

function Test-WslReady {
  if (-not (Get-Command wsl -ErrorAction SilentlyContinue)) {
    return $false
  }

  $distros = & wsl -l -q 2>$null
  $wslCode = $LASTEXITCODE
  $distroText = ($distros -join "").Trim()
  return ($wslCode -eq 0 -and -not [string]::IsNullOrWhiteSpace($distroText))
}

function Get-WslUnixPath {
  param(
    [Parameter(Mandatory=$true)][string]$WindowsPath,
    [Parameter(Mandatory=$true)][string]$Label
  )

  # wslpath recibe mejor rutas tipo C:/... que C:\... desde PowerShell.
  $normalizedPath = $WindowsPath -replace "\\", "/"
  $output = & wsl wslpath -a -u "$normalizedPath" 2>$null
  $exitCode = $LASTEXITCODE
  $unixPath = ($output -join "").Trim()

  if ($exitCode -ne 0 -or [string]::IsNullOrWhiteSpace($unixPath)) {
    throw "No se pudo convertir a ruta WSL ($Label): $WindowsPath. Verifica que WSL y una distro Linux estén instalados."
  }

  return $unixPath
}

function Resolve-PuttyGen {
  $cmd = Get-Command puttygen.exe -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source -ErrorAction SilentlyContinue
  if ($cmd) {
    return $cmd
  }

  $candidates = @(
    "D:\Program Files\PuTTY\puttygen.exe",
    "C:\Program Files\PuTTY\puttygen.exe",
    "C:\Program Files (x86)\PuTTY\puttygen.exe"
  )

  foreach ($candidate in $candidates) {
    if (Test-Path $candidate) {
      return $candidate
    }
  }

  return ""
}

function Resolve-Plink {
  $cmd = Get-Command plink.exe -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source -ErrorAction SilentlyContinue
  if ($cmd) {
    return $cmd
  }

  $candidates = @(
    "D:\Program Files\PuTTY\plink.exe",
    "C:\Program Files\PuTTY\plink.exe",
    "C:\Program Files (x86)\PuTTY\plink.exe"
  )

  foreach ($candidate in $candidates) {
    if (Test-Path $candidate) {
      return $candidate
    }
  }

  return ""
}

function Resolve-Pscp {
  $cmd = Get-Command pscp.exe -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source -ErrorAction SilentlyContinue
  if ($cmd) {
    return $cmd
  }

  $candidates = @(
    "D:\Program Files\PuTTY\pscp.exe",
    "C:\Program Files\PuTTY\pscp.exe",
    "C:\Program Files (x86)\PuTTY\pscp.exe"
  )

  foreach ($candidate in $candidates) {
    if (Test-Path $candidate) {
      return $candidate
    }
  }

  return ""
}

function Resolve-SshExe {
  $cmd = Get-Command ssh.exe -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source -ErrorAction SilentlyContinue
  if ($cmd) {
    return $cmd
  }

  $candidates = @(
    "C:\Windows\System32\OpenSSH\ssh.exe",
    "C:\Program Files\Git\usr\bin\ssh.exe"
  )

  foreach ($candidate in $candidates) {
    if (Test-Path $candidate) {
      return $candidate
    }
  }

  return ""
}

function Resolve-ScpExe {
  $cmd = Get-Command scp.exe -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source -ErrorAction SilentlyContinue
  if ($cmd) {
    return $cmd
  }

  $candidates = @(
    "C:\Windows\System32\OpenSSH\scp.exe",
    "C:\Program Files\Git\usr\bin\scp.exe"
  )

  foreach ($candidate in $candidates) {
    if (Test-Path $candidate) {
      return $candidate
    }
  }

  return ""
}

function Ensure-PuttyTools {
  param([bool]$AutoInstall)

  $plink = Resolve-Plink
  $pscp = Resolve-Pscp
  if ($plink -and $pscp) {
    return [pscustomobject]@{ Plink = $plink; Pscp = $pscp }
  }

  if ($AutoInstall -and (Get-Command winget -ErrorAction SilentlyContinue)) {
    Write-Host "[INFO] PuTTY no está completo. Intentando instalar dependencias con winget..."
    & winget install --id PuTTY.PuTTY -e --accept-package-agreements --accept-source-agreements
    if ($LASTEXITCODE -ne 0) {
      Write-Warning "No se pudo instalar PuTTY automáticamente con winget."
    }

    $plink = Resolve-Plink
    $pscp = Resolve-Pscp
    if ($plink -and $pscp) {
      return [pscustomobject]@{ Plink = $plink; Pscp = $pscp }
    }
  }

  throw "Faltan herramientas PuTTY (plink/pscp). Instala PuTTY: winget install --id PuTTY.PuTTY -e"
}

function Is-NetworkTimeoutMessage {
  param([AllowEmptyString()][string]$Text)
  if ([string]::IsNullOrWhiteSpace($Text)) {
    return $false
  }

  return ($Text -match "(?i)timed out|tiempo de espera|network error|No route to host|Connection reset|Connection refused")
}

function Is-AuthDeniedMessage {
  param([AllowEmptyString()][string]$Text)
  if ([string]::IsNullOrWhiteSpace($Text)) {
    return $false
  }

  return ($Text -match "(?i)Permission denied \(publickey,password\)|Auth fail")
}

function Write-TaggedExternalOutput {
  param([AllowEmptyString()][string]$Line)

  if ($null -eq $Line) {
    return
  }

  $text = "$Line".TrimEnd()
  if ([string]::IsNullOrWhiteSpace($text)) {
    return
  }

  switch -Regex ($text) {
    '^BOOTSTRAP_STEP:(.+)$' {
      Write-Host ("[INFO] Bootstrap: " + $Matches[1])
      return
    }
    '^BOOTSTRAP_OK:(.+)$' {
      Write-Host ("[OK] Bootstrap: " + $Matches[1])
      return
    }
    '^BOOTSTRAP_WARN:(.+)$' {
      Write-Warning ("Bootstrap: " + $Matches[1])
      return
    }
    '^BOOTSTRAP_HINT:(.+)$' {
      Write-Host ("[INFO] Bootstrap sugerencia: " + $Matches[1])
      return
    }
    '^BOOTSTRAP_ERROR:(.+)$' {
      Write-Host ("[ERROR] Bootstrap: " + $Matches[1]) -ForegroundColor Red
      return
    }
    '^DEPLOY_STEP:(.+)$' {
      Write-Host ("[INFO] Deploy: " + $Matches[1])
      return
    }
    '^DEPLOY_OK:(.+)$' {
      Write-Host ("[OK] Deploy: " + $Matches[1])
      return
    }
    '^DEPLOY_WARN:(.+)$' {
      Write-Warning ("Deploy: " + $Matches[1])
      return
    }
    '^DEPLOY_HINT:(.+)$' {
      Write-Host ("[INFO] Deploy sugerencia: " + $Matches[1])
      return
    }
    '^DEPLOY_ERROR:(.+)$' {
      Write-Host ("[ERROR] Deploy: " + $Matches[1]) -ForegroundColor Red
      return
    }
    '^DEPLOY_LOG:(.+)$' {
      Write-Host ("[INFO] Deploy log: " + $Matches[1])
      return
    }
    default {
      Write-Host $text
      return
    }
  }
}

function Get-FriendlyExternalFailureMessage {
  param(
    [Parameter(Mandatory=$true)][string]$Label,
    [Parameter(Mandatory=$true)][int]$ExitCode,
    [AllowEmptyString()][string]$Text
  )

  if (Is-NetworkTimeoutMessage -Text $Text) {
    return ("Timeout de red durante " + $Label + ". Verifica internet, firewall, VPN y acceso al VPS remoto.")
  }

  if (Is-AuthDeniedMessage -Text $Text) {
    return ("Autenticación SSH rechazada durante " + $Label + ". Verifica la clave configurada en IdentityFile y que su pública esté instalada en el VPS.")
  }

  $hints = New-Object System.Collections.Generic.List[string]

  if ($Text -match 'BOOTSTRAP_ERROR:POSTGRES_MISSING_DSN') {
    $hints.Add("Configura DbEmpresasDsn y DbSuperadminDsn, o deja ambos DSN válidos en backend/.env.local del VPS.")
  }
  if ($Text -match 'BOOTSTRAP_ERROR:INVALID_SERVER_PORT') {
    $hints.Add("Revisa -ServerPort y confirma que sea un puerto numérico entre 1 y 65535.")
  }
  if ($Text -match 'BOOTSTRAP_ERROR:PRIVILEGE_REQUIRED') {
    $hints.Add("Conéctate como root o habilita sudo sin contraseña para el usuario remoto.")
  }
  if ($Text -match 'BOOTSTRAP_ERROR:PACKAGE_INSTALL_FAILED') {
    $hints.Add("Revisa el gestor de paquetes del VPS y ejecuta manualmente la instalación si el repositorio está bloqueado.")
  }
  if ($Text -match 'DEPLOY_ERROR:SYSTEMD_UNAVAILABLE') {
    $hints.Add("El VPS debe ser un Linux con systemd activo para mantener el backend persistente.")
  }
  if ($Text -match 'DEPLOY_ERROR:BIN_NOT_FOUND') {
    $hints.Add("Compila el backend sin -SkipBuild o revisa -RemoteBinaryPath para que apunte al binario correcto.")
  }
  if ($Text -match 'DEPLOY_ERROR:PRIVILEGE_REQUIRED') {
    $hints.Add("La instalación y reinicio del servicio requieren root o sudo sin contraseña.")
  }
  if ($Text -match 'DEPLOY_ERROR:(SERVICE_RESTART_FAILED|SERVICE_NOT_RUNNING)') {
    $hints.Add("Revisa backend/.env.local, DB_*_DSN, CONFIG_ENC_KEY y los logs impresos del servicio remoto.")
  }
  if ($Text -match 'DEPLOY_WARN:HEALTHCHECK_TIMEOUT') {
    $hints.Add("El proceso quedó activo, pero el healthcheck no respondió a tiempo. Verifica SERVER_PORT y que GET / responda localmente en el VPS.")
  }

  $hintText = ""
  if ($hints.Count -gt 0) {
    $hintText = " " + (($hints | Select-Object -Unique) -join " ")
  }

  return ("Falló " + $Label + " (código " + $ExitCode + ")." + $hintText)
}

function Test-TcpPortReachable {
  param(
    [Parameter(Mandatory=$true)][string]$TargetHost,
    [Parameter(Mandatory=$true)][int]$Port,
    [int]$TimeoutMs = 8000
  )

  if ($TimeoutMs -lt 1000) {
    $TimeoutMs = 1000
  }

  $client = $null
  try {
    $client = New-Object System.Net.Sockets.TcpClient
    $async = $client.BeginConnect($TargetHost, $Port, $null, $null)
    $connected = $async.AsyncWaitHandle.WaitOne($TimeoutMs, $false)
    if (-not $connected) {
      return $false
    }
    $client.EndConnect($async) | Out-Null
    return $true
  }
  catch {
    return $false
  }
  finally {
    if ($client) {
      try { $client.Close() } catch {}
    }
  }
}

function Get-RemoteBootstrapCommand {
  param(
    [Parameter(Mandatory=$true)][string]$RemotePath,
    [Parameter(Mandatory=$true)][string]$ServerPort,
    [AllowEmptyString()][string]$GoogleClientId,
    [AllowEmptyString()][string]$GoogleClientSecret,
    [AllowEmptyString()][string]$GoogleRedirectUrl,
    [AllowEmptyString()][string]$DbDialect = "postgres",
    [AllowEmptyString()][string]$DbEmpresasDsn = "",
    [AllowEmptyString()][string]$DbSuperadminDsn = ""
  )

  $backendDirLit = Convert-ToBashLiteral ($RemotePath.TrimEnd('/') + "/backend")
  $googleIdLit = Convert-ToBashLiteral $GoogleClientId
  $googleSecretLit = Convert-ToBashLiteral $GoogleClientSecret
  $googleRedirectLit = Convert-ToBashLiteral $GoogleRedirectUrl
  $dbDialectNormalized = if ([string]::IsNullOrWhiteSpace($DbDialect)) { "" } else { $DbDialect.Trim().ToLowerInvariant() }
  $dbDialectLit = Convert-ToBashLiteral $dbDialectNormalized
  $dbEmpresasDsnLit = Convert-ToBashLiteral $DbEmpresasDsn
  $dbSuperadminDsnLit = Convert-ToBashLiteral $DbSuperadminDsn
  $safePort = if ([string]::IsNullOrWhiteSpace($ServerPort)) { "8080" } else { $ServerPort }
  $serverPortLit = Convert-ToBashLiteral $safePort

  $template = @'
set -e;
log(){ echo "BOOTSTRAP_STEP:$1"; };
warn(){ echo "BOOTSTRAP_WARN:$1"; };
ok(){ echo "BOOTSTRAP_OK:$1"; };
hint(){ echo "BOOTSTRAP_HINT:$1"; };
fail(){ echo "BOOTSTRAP_ERROR:$1"; exit 1; };
can_run_root(){
  if [ "$(id -u)" -eq 0 ]; then
    return 0;
  fi;
  if command -v sudo >/dev/null 2>&1 && sudo -n true >/dev/null 2>&1; then
    return 0;
  fi;
  return 1;
};
run_root(){
  if [ "$(id -u)" -eq 0 ]; then
    "$@";
    return $?;
  fi;
  sudo -n "$@";
};
backend_dir=__BACKEND_DIR__;
env_file="$backend_dir/.env.local";
server_port=__SERVER_PORT__;
case "$server_port" in
  ''|*[!0-9]*) fail "INVALID_SERVER_PORT SERVER_PORT debe ser numerico. Valor recibido: $server_port" ;;
esac;
if [ "$server_port" -lt 1 ] || [ "$server_port" -gt 65535 ]; then
  fail "INVALID_SERVER_PORT SERVER_PORT fuera de rango (1-65535). Valor recibido: $server_port";
fi;
log "preparando directorio remoto y archivo de entorno";
mkdir -p "$backend_dir" "$backend_dir/bin" "$backend_dir/tmp";
touch "$env_file";
chmod 600 "$env_file" || true;
ok "ENV_FILE listo en $env_file";
log "detectando sistema y dependencias del VPS";
os_name="$(uname -s 2>/dev/null || echo desconocido)";
os_release="$(uname -r 2>/dev/null || echo desconocido)";
os_arch="$(uname -m 2>/dev/null || echo desconocido)";
ok "SYSTEM_INFO host=$os_name arch=$os_arch kernel=$os_release";
if command -v apt-get >/dev/null 2>&1; then
  ok "PKG_MANAGER detectado apt-get";
  missing_base_deps=0;
  for cmd in curl wget lsof ps; do
    if ! command -v "$cmd" >/dev/null 2>&1; then missing_base_deps=1; fi;
  done;
  if [ "$missing_base_deps" -eq 0 ]; then
    ok "PKG_INSTALL_SKIP paquetes base ya disponibles; se omite apt-get update/install";
  elif can_run_root; then
    export DEBIAN_FRONTEND=noninteractive;
    run_root apt-get update -y >/dev/null 2>&1 || warn "APT_UPDATE apt-get update reporto incidencias; se intentara instalar de todos modos";
    if run_root apt-get install -y ca-certificates curl wget procps lsof >/dev/null 2>&1; then
      ok "PKG_INSTALL paquetes base instalados con apt-get: ca-certificates curl wget procps lsof";
    else
      fail "PACKAGE_INSTALL_FAILED fallo apt-get install de paquetes base";
    fi;
  else
    warn "PRIVILEGE_REQUIRED sin root ni sudo -n; se omite instalacion automatica de paquetes";
    hint "Conectate como root o habilita sudo sin contraseña si quieres que el script prepare dependencias del VPS";
  fi;
elif command -v dnf >/dev/null 2>&1; then
  ok "PKG_MANAGER detectado dnf";
  if can_run_root; then
    if run_root dnf install -y ca-certificates curl wget procps-ng lsof >/dev/null 2>&1; then
      ok "PKG_INSTALL paquetes base instalados con dnf: ca-certificates curl wget procps-ng lsof";
    else
      fail "PACKAGE_INSTALL_FAILED fallo dnf install de paquetes base";
    fi;
  else
    warn "PRIVILEGE_REQUIRED sin root ni sudo -n; se omite instalacion automatica de paquetes";
  fi;
elif command -v yum >/dev/null 2>&1; then
  ok "PKG_MANAGER detectado yum";
  if can_run_root; then
    if run_root yum install -y ca-certificates curl wget procps-ng lsof >/dev/null 2>&1; then
      ok "PKG_INSTALL paquetes base instalados con yum: ca-certificates curl wget procps-ng lsof";
    else
      fail "PACKAGE_INSTALL_FAILED fallo yum install de paquetes base";
    fi;
  else
    warn "PRIVILEGE_REQUIRED sin root ni sudo -n; se omite instalacion automatica de paquetes";
  fi;
elif command -v apk >/dev/null 2>&1; then
  ok "PKG_MANAGER detectado apk";
  if can_run_root; then
    if run_root apk add --no-cache ca-certificates curl wget procps lsof >/dev/null 2>&1; then
      ok "PKG_INSTALL paquetes base instalados con apk: ca-certificates curl wget procps lsof";
    else
      fail "PACKAGE_INSTALL_FAILED fallo apk add de paquetes base";
    fi;
  else
    warn "PRIVILEGE_REQUIRED sin root ni sudo -n; se omite instalacion automatica de paquetes";
  fi;
elif command -v zypper >/dev/null 2>&1; then
  ok "PKG_MANAGER detectado zypper";
  if can_run_root; then
    if run_root zypper --non-interactive install ca-certificates curl wget procps lsof >/dev/null 2>&1; then
      ok "PKG_INSTALL paquetes base instalados con zypper: ca-certificates curl wget procps lsof";
    else
      fail "PACKAGE_INSTALL_FAILED fallo zypper install de paquetes base";
    fi;
  else
    warn "PRIVILEGE_REQUIRED sin root ni sudo -n; se omite instalacion automatica de paquetes";
  fi;
else
  warn "PKG_MANAGER_UNKNOWN no se detecto apt-get, dnf, yum, apk ni zypper";
  hint "Verifica manualmente que el VPS tenga ca-certificates, curl o wget, lsof y utilidades base antes del reinicio";
fi;
if command -v systemctl >/dev/null 2>&1; then
  ok "SYSTEMD_OK systemctl disponible";
else
  warn "SYSTEMD_MISSING systemctl no esta disponible; el backend persistente requiere systemd activo";
  hint "Usa una VM Linux con systemd o ajusta manualmente el metodo de arranque del VPS";
fi;
gid=__GOOGLE_ID__;
gsec=__GOOGLE_SECRET__;
grurl=__GOOGLE_REDIRECT_URL__;
dbdialect=__DB_DIALECT__;
dbemp=__DB_EMPRESAS_DSN__;
dbsuper=__DB_SUPERADMIN_DSN__;
get_env_value(){ grep -E "^$1=" "$env_file" | tail -n1 | cut -d= -f2- || true; };
upsert_env(){
  key="$1";
  value="$2";
  grep -v "^$key=" "$env_file" > "$env_file.tmp" 2>/dev/null || true;
  mv "$env_file.tmp" "$env_file" 2>/dev/null || true;
  printf '%s=%s\n' "$key" "$value" >> "$env_file";
};
current_dbdialect="$(get_env_value DB_DIALECT)";
current_dbemp="$(get_env_value DB_EMPRESAS_DSN)";
current_dbsuper="$(get_env_value DB_SUPERADMIN_DSN)";
current_gid="$(get_env_value GOOGLE_CLIENT_ID)";
current_gsec="$(get_env_value GOOGLE_CLIENT_SECRET)";
current_grurl="$(get_env_value GOOGLE_REDIRECT_URL)";
effective_dbdialect="$dbdialect";
effective_dbemp="$dbemp";
effective_dbsuper="$dbsuper";
effective_gid="$gid";
effective_gsec="$gsec";
effective_grurl="$grurl";
if [ -z "$effective_dbdialect" ]; then effective_dbdialect="$current_dbdialect"; fi;
if [ -z "$effective_dbemp" ]; then effective_dbemp="$current_dbemp"; fi;
if [ -z "$effective_dbsuper" ]; then effective_dbsuper="$current_dbsuper"; fi;
if [ -z "$effective_gid" ]; then effective_gid="$current_gid"; fi;
if [ -z "$effective_gsec" ]; then effective_gsec="$current_gsec"; fi;
if [ -z "$effective_grurl" ]; then effective_grurl="$current_grurl"; fi;
if [ -z "$effective_dbdialect" ] && { [ -n "$effective_dbemp" ] || [ -n "$effective_dbsuper" ]; }; then
  effective_dbdialect=postgres;
fi;
if [ "$effective_dbdialect" = "postgres" ] && { [ -z "$effective_dbemp" ] || [ -z "$effective_dbsuper" ]; }; then
  echo "BOOTSTRAP_ERROR:POSTGRES_MISSING_DSN faltan DB_EMPRESAS_DSN y/o DB_SUPERADMIN_DSN para modo postgres";
  echo "BOOTSTRAP_HINT:Define DbEmpresasDsn y DbSuperadminDsn, o deja ambos DSN validos en backend/.env.local del VPS";
  exit 1;
fi;
log "sincronizando backend/.env.local remoto";
upsert_env SERVER_PORT "$server_port";
ok "SERVER_PORT actualizado a $server_port";
if [ -n "$effective_dbdialect" ]; then
  upsert_env DB_DIALECT "$effective_dbdialect";
fi;
if [ -n "$effective_dbemp" ]; then
  upsert_env DB_EMPRESAS_DSN "$effective_dbemp";
fi;
if [ -n "$effective_dbsuper" ]; then
  upsert_env DB_SUPERADMIN_DSN "$effective_dbsuper";
fi;
if [ -n "$effective_gid" ]; then upsert_env GOOGLE_CLIENT_ID "$effective_gid"; fi;
if [ -n "$effective_gsec" ]; then upsert_env GOOGLE_CLIENT_SECRET "$effective_gsec"; fi;
if [ -n "$effective_grurl" ]; then upsert_env GOOGLE_REDIRECT_URL "$effective_grurl"; fi;
for k in DB_DIALECT DB_SUPERADMIN_DSN DB_EMPRESAS_DSN GOOGLE_CLIENT_ID GOOGLE_CLIENT_SECRET GOOGLE_REDIRECT_URL SERVER_PORT CONFIG_ENC_KEY; do
  line="$(grep -E "^$k=" "$env_file" | tail -n1 || true)";
  if [ -z "$line" ]; then
    case "$k" in
      GOOGLE_CLIENT_ID|GOOGLE_CLIENT_SECRET|GOOGLE_REDIRECT_URL)
        echo "BOOTSTRAP_WARN:$k ausente (solo requerido para login Google)";
        ;;
      CONFIG_ENC_KEY)
        echo "BOOTSTRAP_WARN:CONFIG_ENC_KEY ausente (requerida para cifrado de secretos)";
        echo "BOOTSTRAP_HINT:Define CONFIG_ENC_KEY en backend/.env.local antes de guardar credenciales sensibles";
        ;;
      *)
        echo "BOOTSTRAP_WARN:$k ausente en backend/.env.local";
        ;;
    esac;
  else
    val="${line#*=}";
    if [ -z "$val" ]; then
      case "$k" in
        GOOGLE_CLIENT_ID|GOOGLE_CLIENT_SECRET|GOOGLE_REDIRECT_URL)
          echo "BOOTSTRAP_WARN:$k vacio (solo requerido para login Google)";
          ;;
        CONFIG_ENC_KEY)
          echo "BOOTSTRAP_WARN:CONFIG_ENC_KEY vacia (requerida para cifrado de secretos)";
          echo "BOOTSTRAP_HINT:Define CONFIG_ENC_KEY en backend/.env.local antes de guardar credenciales sensibles";
          ;;
        *)
          echo "BOOTSTRAP_WARN:$k vacio en backend/.env.local";
          ;;
      esac;
    else
      echo "BOOTSTRAP_OK:$k configurado";
    fi;
  fi;
done;
ok "BOOTSTRAP_COMPLETE entorno remoto preparado para el redeploy"
'@

  $cmd = $template.Replace("__BACKEND_DIR__", $backendDirLit).Replace("__SERVER_PORT__", $serverPortLit).Replace("__GOOGLE_ID__", $googleIdLit).Replace("__GOOGLE_SECRET__", $googleSecretLit).Replace("__GOOGLE_REDIRECT_URL__", $googleRedirectLit).Replace("__DB_DIALECT__", $dbDialectLit).Replace("__DB_EMPRESAS_DSN__", $dbEmpresasDsnLit).Replace("__DB_SUPERADMIN_DSN__", $dbSuperadminDsnLit)
  $cmd = $cmd -replace "`r", "" -replace "`n", " "
  return $cmd
}

function Get-RemoteRestartCommand {
  param(
    [Parameter(Mandatory=$true)][string]$RemotePath,
    [Parameter(Mandatory=$true)][string]$BinaryRelativePath,
    [Parameter(Mandatory=$true)][string]$ServerPort,
    [Parameter(Mandatory=$true)][string]$StdoutLogRelativePath,
    [Parameter(Mandatory=$true)][string]$StderrLogRelativePath,
    [Parameter(Mandatory=$true)][int]$HealthTimeoutSeconds
  )

  $repoDirLit = Convert-ToBashLiteral ($RemotePath.TrimEnd('/'))
  $binaryRelLit = Convert-ToBashLiteral (($BinaryRelativePath -replace "\\", "/").TrimStart('/'))
  $stdoutRelLit = Convert-ToBashLiteral (($StdoutLogRelativePath -replace "\\", "/").TrimStart('/'))
  $stderrRelLit = Convert-ToBashLiteral (($StderrLogRelativePath -replace "\\", "/").TrimStart('/'))
  $safePort = if ([string]::IsNullOrWhiteSpace($ServerPort)) { "8080" } else { $ServerPort }
  $portLit = Convert-ToBashLiteral $safePort
  $safeTimeout = if ($HealthTimeoutSeconds -lt 5) { 5 } else { $HealthTimeoutSeconds }

  $template = @'
set -e;
log(){ echo "DEPLOY_STEP:$1"; };
warn(){ echo "DEPLOY_WARN:$1"; };
hint(){ echo "DEPLOY_HINT:$1"; };
fail(){ echo "DEPLOY_ERROR:$1"; exit 1; };
can_run_root(){
  if [ "$(id -u)" -eq 0 ]; then
    return 0;
  fi;
  if command -v sudo >/dev/null 2>&1 && sudo -n true >/dev/null 2>&1; then
    return 0;
  fi;
  return 1;
};
run_root(){
  if [ "$(id -u)" -eq 0 ]; then
    "$@";
    return $?;
  fi;
  sudo -n "$@";
};
repo_dir=__REPO_DIR__;
backend_dir=$repo_dir/backend;
bin_rel=__BIN_REL__;
port=__PORT__;
health_timeout=__HEALTH_TIMEOUT__;
bin_path=$repo_dir/$bin_rel;
stdout_rel=__STDOUT_REL__;
stderr_rel=__STDERR_REL__;
stdout_log=$repo_dir/$stdout_rel;
stderr_log=$repo_dir/$stderr_rel;
env_file=$backend_dir/.env.local;
pid_file=$repo_dir/backend/server.pid;
service_base=$(basename "$repo_dir");
service_name=$(printf '%s' "$service_base" | tr -c 'A-Za-z0-9_.@-' '_');
service_unit=$service_name.service;
service_file=/etc/systemd/system/$service_unit;
dump_diagnostics(){
  echo "DEPLOY_LOG:systemctl status $service_unit";
  run_root systemctl status "$service_unit" --no-pager -l || true;
  if command -v journalctl >/dev/null 2>&1; then
    echo "DEPLOY_LOG:journalctl -u $service_unit -n 80";
    run_root journalctl -u "$service_unit" -n 80 --no-pager || true;
  fi;
  if command -v ss >/dev/null 2>&1; then
    echo "DEPLOY_LOG:ss -ltnp (*:$port)";
    run_root ss -ltnp 2>/dev/null | awk -v port="$port" '$4 ~ ":" port "$" { print; }' || true;
  fi;
  if [ -f "$stderr_log" ]; then
    echo "DEPLOY_LOG:tail -n 80 $stderr_log";
    tail -n 80 "$stderr_log" || true;
  fi;
  if [ -f "$stdout_log" ]; then
    echo "DEPLOY_LOG:tail -n 40 $stdout_log";
    tail -n 40 "$stdout_log" || true;
  fi;
};
pids_listening_on_port(){
  if command -v ss >/dev/null 2>&1; then
    run_root ss -ltnp 2>/dev/null | awk -v port="$port" '
      $4 ~ ":" port "$" {
        line = $0;
        while (match(line, /pid=[0-9]+/)) {
          pid = substr(line, RSTART + 4, RLENGTH - 4);
          print pid;
          line = substr(line, RSTART + RLENGTH);
        }
      }
    ' | sort -u;
    return 0;
  fi;
  if command -v lsof >/dev/null 2>&1; then
    run_root lsof -tiTCP:"$port" -sTCP:LISTEN 2>/dev/null | sort -u;
    return 0;
  fi;
  return 1;
};
pids_matching_backend_binary(){
  if command -v pgrep >/dev/null 2>&1; then
    run_root pgrep -f "$bin_path" 2>/dev/null | sort -u;
    return 0;
  fi;
  run_root ps -eo pid=,args= 2>/dev/null | awk -v target="$bin_path" 'index($0, target) > 0 { print $1; }' | sort -u;
};
skip_cleanup_pid(){
  pid="$1";
  case "$pid" in
    ''|*[!0-9]*|0|1)
      return 0;
      ;;
  esac;
  if [ "$pid" = "$$" ] || [ "$pid" = "$PPID" ]; then
    return 0;
  fi;
  return 1;
};
cleanup_stale_backend_processes(){
  stray_backend_pids=$(pids_matching_backend_binary || true);
  if [ -z "$stray_backend_pids" ]; then
    return 0;
  fi;
  echo "DEPLOY_WARN:STALE_BACKEND_PROCESSES se detectaron procesos previos del backend; se intentara detenerlos antes del nuevo arranque";
  for pid in $stray_backend_pids; do
    if skip_cleanup_pid "$pid"; then
      continue;
    fi;
    exe_path=$(readlink -f "/proc/$pid/exe" 2>/dev/null || true);
    cmdline=$(tr '\0' ' ' <"/proc/$pid/cmdline" 2>/dev/null | sed 's/[[:space:]]*$//');
    echo "DEPLOY_WARN:STALE_BACKEND_PID pid=$pid exe=${exe_path:-unknown} cmd=${cmdline:-unknown}";
    run_root kill "$pid" >/dev/null 2>&1 || true;
  done;
  sleep 2;
  stray_backend_pids=$(pids_matching_backend_binary || true);
  if [ -n "$stray_backend_pids" ]; then
    for pid in $stray_backend_pids; do
      if skip_cleanup_pid "$pid"; then
        continue;
      fi;
      echo "DEPLOY_WARN:STALE_BACKEND_PID_FORCE pid=$pid";
      run_root kill -9 "$pid" >/dev/null 2>&1 || true;
    done;
    sleep 1;
  fi;
  stray_backend_pids=$(pids_matching_backend_binary || true);
  if [ -n "$stray_backend_pids" ]; then
    echo "DEPLOY_ERROR:STALE_BACKEND_PID_PRESENT no se pudieron detener todos los procesos previos del backend (pids=$(printf '%s' "$stray_backend_pids" | tr '\n' ' ' | sed 's/[[:space:]]*$//'))";
    echo "DEPLOY_HINT:Revisa procesos manuales del backend fuera de systemd antes de reintentar el deploy";
    dump_diagnostics;
    exit 1;
  fi;
};
cleanup_port_conflicts(){
  run_root systemctl stop "$service_unit" >/dev/null 2>&1 || true;
  stray_pids=$(pids_listening_on_port || true);
  if [ -z "$stray_pids" ]; then
    return 0;
  fi;
  echo "DEPLOY_WARN:PORT_IN_USE se detectaron procesos fuera de systemd ocupando el puerto $port; se intentara liberarlo";
  for pid in $stray_pids; do
    if skip_cleanup_pid "$pid"; then
      continue;
    fi;
    exe_path=$(readlink -f "/proc/$pid/exe" 2>/dev/null || true);
    cmdline=$(tr '\0' ' ' <"/proc/$pid/cmdline" 2>/dev/null | sed 's/[[:space:]]*$//');
    echo "DEPLOY_WARN:PORT_PID pid=$pid exe=${exe_path:-unknown} cmd=${cmdline:-unknown}";
    run_root kill "$pid" >/dev/null 2>&1 || true;
  done;
  sleep 2;
  stray_pids=$(pids_listening_on_port || true);
  if [ -n "$stray_pids" ]; then
    for pid in $stray_pids; do
      if skip_cleanup_pid "$pid"; then
        continue;
      fi;
      echo "DEPLOY_WARN:PORT_PID_FORCE pid=$pid";
      run_root kill -9 "$pid" >/dev/null 2>&1 || true;
    done;
    sleep 1;
  fi;
  stray_pids=$(pids_listening_on_port || true);
  if [ -n "$stray_pids" ]; then
    echo "DEPLOY_ERROR:PORT_STILL_BUSY no se pudo liberar el puerto $port (pids=$(printf '%s' "$stray_pids" | tr '\n' ' ' | sed 's/[[:space:]]*$//'))";
    echo "DEPLOY_HINT:Deten manualmente los procesos que aun escuchan en el puerto o cambia SERVER_PORT antes de reintentar";
    dump_diagnostics;
    exit 1;
  fi;
};
log "preparando servicio systemd persistente";
if ! can_run_root; then
  fail "PRIVILEGE_REQUIRED se requiere root o sudo -n para instalar o reiniciar la unidad systemd";
fi;
if ! command -v systemctl >/dev/null 2>&1; then
  fail "SYSTEMD_UNAVAILABLE systemctl no esta disponible en el VPS";
fi;
if [ ! -f "$bin_path" ]; then
  fail "BIN_NOT_FOUND binario remoto no encontrado en $bin_path";
fi;
run_root mkdir -p "$(dirname "$stdout_log")" "$(dirname "$stderr_log")";
run_root touch "$stdout_log" "$stderr_log";
chmod +x "$bin_path" || true;
rm -f $pid_file 2>/dev/null || true;
tmp_service=$(mktemp "${TMPDIR:-/tmp}/pcs_sync_service.XXXXXX");
cat > "$tmp_service" <<EOF
[Unit]
Description=Powerful Control System backend ($service_name)
After=network-online.target
Wants=network-online.target
StartLimitIntervalSec=0

[Service]
Type=simple
User=root
WorkingDirectory=$backend_dir
EnvironmentFile=-$env_file
ExecStart=$bin_path
Restart=always
RestartSec=5
KillSignal=SIGTERM
TimeoutStopSec=30
StandardOutput=append:$stdout_log
StandardError=append:$stderr_log

[Install]
WantedBy=multi-user.target
EOF
run_root cp "$tmp_service" "$service_file";
rm -f "$tmp_service";
echo "DEPLOY_STEP:unidad systemd actualizada en $service_file";
run_root systemctl daemon-reload;
run_root systemctl enable "$service_unit" >/dev/null 2>&1;
run_root systemctl reset-failed "$service_unit" >/dev/null 2>&1 || true;
cleanup_stale_backend_processes;
cleanup_port_conflicts;
log "reiniciando $service_unit y validando arranque";
if ! run_root systemctl start "$service_unit"; then
  echo "DEPLOY_ERROR:SERVICE_RESTART_FAILED no fue posible reiniciar $service_unit en el puerto $port";
  echo "DEPLOY_HINT:Revisa backend/.env.local, los DSN PostgreSQL, CONFIG_ENC_KEY y el binario remoto antes de reintentar";
  dump_diagnostics;
  exit 1;
fi;
healthy=0;
for i in $(seq 1 $health_timeout); do
  if ! run_root systemctl is-active --quiet "$service_unit"; then
    echo "DEPLOY_ERROR:SERVICE_NOT_RUNNING el servicio $service_unit se detuvo durante el healthcheck";
    echo "DEPLOY_HINT:Revisa backend/.env.local, el puerto SERVER_PORT=$port y los logs mostrados abajo";
    dump_diagnostics;
    exit 1;
  fi;
  if command -v curl >/dev/null 2>&1; then
    http_code=$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:$port/ || true);
    if [ -n "$http_code" ] && [ "$http_code" != "000" ]; then
      healthy=1;
      break;
    fi;
  elif command -v wget >/dev/null 2>&1; then
    if wget -qO- http://127.0.0.1:$port/ >/dev/null 2>&1; then
      healthy=1;
      break;
    fi;
  else
    healthy=1;
    break;
  fi;
  sleep 1;
done;
main_pid=$(run_root systemctl show -p MainPID --value "$service_unit" 2>/dev/null || echo 0);
enabled_state=$(run_root systemctl is-enabled "$service_unit" 2>/dev/null || true);
if [ $healthy -eq 1 ]; then
  echo "DEPLOY_OK:SERVICE_READY servicio $service_unit activo (pid=$main_pid, puerto=$port, enabled=$enabled_state)";
else
  echo "DEPLOY_WARN:HEALTHCHECK_TIMEOUT el servicio $service_unit quedo activo (pid=$main_pid, enabled=$enabled_state) pero no respondio al healthcheck en $health_timeout s";
  echo "DEPLOY_HINT:Verifica que SERVER_PORT=$port coincida con backend/.env.local y que GET / responda localmente en el VPS";
fi;
'@

  $cmd = $template.Replace("__REPO_DIR__", $repoDirLit).Replace("__BIN_REL__", $binaryRelLit).Replace("__STDOUT_REL__", $stdoutRelLit).Replace("__STDERR_REL__", $stderrRelLit).Replace("__PORT__", $portLit).Replace("__HEALTH_TIMEOUT__", "$safeTimeout")
  $cmd = $cmd -replace "`r", ""
  return $cmd
}

function Invoke-ExternalWithRetry {
  param(
    [Parameter(Mandatory=$true)][string]$Label,
    [Parameter(Mandatory=$true)][string]$CommandPath,
    [Parameter(Mandatory=$true)][string[]]$Arguments,
    [int]$MaxAttempts = 3,
    [switch]$RetryOnTimeoutOnly
  )

  if ($MaxAttempts -lt 1) {
    $MaxAttempts = 1
  }

  for ($attempt = 1; $attempt -le $MaxAttempts; $attempt++) {
    Write-Host ("[INFO] " + $Label + " (intento " + $attempt + "/" + $MaxAttempts + ")...")
    $output = & $CommandPath @Arguments 2>&1
    $exitCode = $LASTEXITCODE
    if ($output) {
      $output | ForEach-Object { Write-TaggedExternalOutput -Line "$_" }
    }

    if ($exitCode -eq 0) {
      return
    }

    $text = ($output -join "`n")
    $isTimeout = Is-NetworkTimeoutMessage -Text $text
    $canRetry = $attempt -lt $MaxAttempts -and ((-not $RetryOnTimeoutOnly) -or $isTimeout)

    if ($canRetry) {
      Write-Warning ("Fallo en " + $Label + ". Reintentando...")
      continue
    }

    throw (Get-FriendlyExternalFailureMessage -Label $Label -ExitCode $exitCode -Text $text)
  }
}

function Resolve-AbsolutePath {
  param(
    [Parameter(Mandatory=$true)][string]$BasePath,
    [Parameter(Mandatory=$true)][string]$InputPath,
    [switch]$MustExist
  )

  if ([System.IO.Path]::IsPathRooted($InputPath)) {
    $pathCandidate = $InputPath
  } else {
    $pathCandidate = Join-Path $BasePath $InputPath
  }

  if ($MustExist) {
    return (Resolve-Path $pathCandidate).Path
  }

  return [System.IO.Path]::GetFullPath($pathCandidate)
}

function Resolve-DefaultIdentityFile {
  param([Parameter(Mandatory=$true)][string]$RepoRoot)

  $projectPpk = Join-Path $RepoRoot "clave privada ssh.ppk"
  $userOpenSsh = Join-Path $env:USERPROFILE ".ssh\id_rsa"
  $candidates = @($projectPpk, $userOpenSsh)

  foreach ($candidate in $candidates) {
    if ($candidate -and (Test-Path $candidate)) {
      return (Resolve-Path $candidate).Path
    }
  }

  return $userOpenSsh
}

function Get-DotEnvValue {
  param(
    [Parameter(Mandatory=$true)][string]$EnvFilePath,
    [Parameter(Mandatory=$true)][string]$Key
  )

  if (-not (Test-Path $EnvFilePath)) {
    return ""
  }

  $line = Get-Content -LiteralPath $EnvFilePath -ErrorAction SilentlyContinue |
    ForEach-Object { $_.Trim() } |
    Where-Object { $_ -and -not $_.StartsWith("#") -and $_ -like ($Key + "=*") } |
    Select-Object -Last 1

  if (-not $line) {
    return ""
  }

  $parts = $line -split "=", 2
  if ($parts.Count -lt 2) {
    return ""
  }

  return $parts[1].Trim()
}

function Parse-BoolLike {
  param([AllowEmptyString()][string]$Value)

  if ($null -eq $Value) {
    $Value = ""
  }
  $v = $Value.Trim().ToLowerInvariant()
  return ($v -in @("1", "true", "yes", "on", "si", "sí"))
}

function Parse-IntOrDefault {
  param(
    [AllowEmptyString()][string]$Value,
    [int]$DefaultValue
  )

  if ($null -eq $Value) {
    $Value = ""
  }
  $n = 0
  if ([int]::TryParse($Value.Trim(), [ref]$n)) {
    return $n
  }
  return $DefaultValue
}

if ([string]::IsNullOrWhiteSpace($SshHostKey)) {
  $envHostKey = [Environment]::GetEnvironmentVariable('PCS_VPS_SSH_HOSTKEY', 'Process')
  if (-not [string]::IsNullOrWhiteSpace($envHostKey)) {
    $SshHostKey = $envHostKey.Trim()
  }
}
$script:PlinkHostKeyArgs = @()
if (-not [string]::IsNullOrWhiteSpace($SshHostKey)) {
  $script:PlinkHostKeyArgs = @('-hostkey', $SshHostKey.Trim())
}

function Format-ByteSize {
  param([long]$Bytes)

  if ($Bytes -lt 1024) {
    return "$Bytes B"
  }

  $units = @("KB", "MB", "GB", "TB")
  $value = [double]$Bytes
  $unitIndex = -1
  do {
    $value = $value / 1024
    $unitIndex += 1
  } while ($value -ge 1024 -and $unitIndex -lt ($units.Count - 1))

  return ("{0:N1} {1}" -f $value, $units[$unitIndex])
}

function Normalize-PostgresDsnForVps {
  param(
    [AllowEmptyString()][string]$Dsn,
    [int]$TunnelLocalPort = 15432,
    [int]$RemoteDbPort = 5432
  )

  if ($null -eq $Dsn) {
    $Dsn = ""
  }
  $trimmed = $Dsn.Trim()
  if ([string]::IsNullOrWhiteSpace($trimmed)) {
    return $trimmed
  }

  $lower = $trimmed.ToLowerInvariant()
  if (-not ($lower.StartsWith("postgres://") -or $lower.StartsWith("postgresql://"))) {
    return $trimmed
  }

  try {
    $uri = [System.Uri]$trimmed
  }
  catch {
    return $trimmed
  }

  $parsedHost = $uri.Host
  if ($null -eq $parsedHost) {
    $parsedHost = ""
  }
  $parsedHost = $parsedHost.Trim().ToLowerInvariant()
  $isLoopback = ($parsedHost -eq "127.0.0.1" -or $parsedHost -eq "localhost" -or $parsedHost -eq "::1")
  if (-not $isLoopback) {
    return $trimmed
  }

  if ($uri.Port -ne $TunnelLocalPort -or $RemoteDbPort -le 0) {
    return $trimmed
  }

  $builder = [System.UriBuilder]::new($uri)
  $builder.Port = $RemoteDbPort
  return $builder.Uri.AbsoluteUri
}

function Invoke-LocalLinuxBuild {
  param(
    [Parameter(Mandatory=$true)][string]$RepoRoot,
    [Parameter(Mandatory=$true)][string]$WorkingDir,
    [Parameter(Mandatory=$true)][string]$Package,
    [Parameter(Mandatory=$true)][string]$OutputPath,
    [Parameter(Mandatory=$true)][string]$GoOS,
    [Parameter(Mandatory=$true)][string]$GoArch,
    [Parameter(Mandatory=$true)][string]$CgoEnabled
  )

  $goCmd = Get-Command go -ErrorAction SilentlyContinue
  if (-not $goCmd) {
    throw "No se encontró 'go' en PATH. Instálalo para compilar el binario Linux local."
  }

  $workDirAbs = Resolve-AbsolutePath -BasePath $RepoRoot -InputPath $WorkingDir -MustExist
  $outputAbs = Resolve-AbsolutePath -BasePath $RepoRoot -InputPath $OutputPath
  $outputDir = Split-Path -Parent $outputAbs
  if (-not (Test-Path $outputDir)) {
    New-Item -ItemType Directory -Path $outputDir -Force | Out-Null
  }

  if (-not (Test-Path (Join-Path $workDirAbs "go.mod"))) {
    throw "No se encontró go.mod en la ruta de compilación: $workDirAbs"
  }

  $prevGOOS = $env:GOOS
  $prevGOARCH = $env:GOARCH
  $prevCGO = $env:CGO_ENABLED
  $locationPushed = $false

  try {
    $env:GOOS = $GoOS
    $env:GOARCH = $GoArch
    $env:CGO_ENABLED = $CgoEnabled

    Write-Host "[INFO] Compilando binario Linux en local..."
    Write-Host ("[INFO] go build -trimpath -o '" + $outputAbs + "' " + $Package)

    Push-Location $workDirAbs
    $locationPushed = $true
    & go build -trimpath -o $outputAbs $Package
    if ($LASTEXITCODE -ne 0) {
      throw "Fallo compilación local Linux (go build, código $LASTEXITCODE)."
    }
  }
  finally {
    if ($locationPushed) {
      Pop-Location
    }
    $env:GOOS = $prevGOOS
    $env:GOARCH = $prevGOARCH
    $env:CGO_ENABLED = $prevCGO
  }

  Write-Host ("[OK] Binario Linux generado: " + $outputAbs)
  return $outputAbs
}

function Get-RelativePathIfInside {
  param(
    [Parameter(Mandatory=$true)][string]$BasePath,
    [Parameter(Mandatory=$true)][string]$TargetPath
  )

  $baseFull = [System.IO.Path]::GetFullPath($BasePath).TrimEnd([char]'\', [char]'/') + "\\"
  $targetFull = [System.IO.Path]::GetFullPath($TargetPath)
  if (-not $targetFull.StartsWith($baseFull, [System.StringComparison]::OrdinalIgnoreCase)) {
    return ""
  }

  $relative = $targetFull.Substring($baseFull.Length)
  return ($relative -replace "\\", "/")
}

function Get-SyncExcludePatterns {
  param(
    [AllowEmptyString()][string]$ExcludeFile,
    [bool]$ExcludeEvidence = $true
  )

  $patterns = @(
    ".git",
    ".git/*",
    ".gitignore",
    ".codex",
    ".codex/*",
    ".codex-gocache",
    ".codex-gocache/*",
    ".codex-tmp-go",
    ".codex-tmp-go/*",
    ".agents",
    ".agents/*",
    ".cache",
    ".cache/*",
    ".gocache",
    ".gocache/*",
    ".gotmp",
    ".gotmp/*",
    "*/.codex-gocache",
    "*/.codex-gocache/*",
    "*/.codex-tmp-go",
    "*/.codex-tmp-go/*",
    "*/.gocache",
    "*/.gocache/*",
    "*/.gotmp",
    "*/.gotmp/*",
    ".cursor",
    ".cursor/*",
    ".github",
    ".github/*",
    ".vscode",
    ".vscode/*",
    "backup",
    "backup/*",
    "descargas",
    "descargas/*",
    "node_modules",
    "*/node_modules",
    "*/*/node_modules",
    "logs",
    "logs/*",
    "scripts/logs",
    "scripts/logs/*",
    "tmp",
    "tmp/*",
    "test_runs",
    "test_runs/*",
    "documentos/evidencias_qa",
    "documentos/evidencias_qa/*",
    "coverage",
    "coverage/*",
    "dist",
    "dist/*",
    "build",
    "build/*",
    ".pytest_cache",
    ".pytest_cache/*",
    "__pycache__",
    "*/__pycache__",
    "*/*/__pycache__",
    "*.db",
    "*.log",
    "*.exe",
    "*.vsix",
    "*.tmp",
    "*.bak",
    "backend/.env.local",
    "backend/.env",
    "deploy/.env.platform",
    "deploy/.env.platform.*",
    "deploy/iredmail/.env",
    "backend/server_linux_amd64",
    "backend/tools",
    "backend/tools/*",
    "backend/tmp",
    "backend/tmp/*",
    "backend/.codex-gocache",
    "backend/.codex-gocache/*",
    "backend/.codex-tmp-go",
    "backend/.codex-tmp-go/*",
    "backend/server.log",
    "backend/server.err",
    "herramientas",
    "herramientas/*",
    "*.ppk",
    "*.pem",
    "*.key"
  )

  if (-not $ExcludeEvidence) {
    $patterns = $patterns | Where-Object {
      $_ -ne "documentos/evidencias_qa" -and $_ -ne "documentos/evidencias_qa/*"
    }
  }

  if ($ExcludeFile) {
    if (-not (Test-Path $ExcludeFile)) {
      throw "exclude-file no existe: $ExcludeFile"
    }
    $extra = Get-Content -LiteralPath $ExcludeFile -ErrorAction Stop |
      ForEach-Object { $_.Trim() } |
      Where-Object { $_ -and -not $_.StartsWith("#") }
    if ($extra) {
      $patterns += $extra
    }
  }

  return $patterns
}

function New-TempRemoteCommandFile {
  param(
    [Parameter(Mandatory=$true)][string]$Prefix,
    [Parameter(Mandatory=$true)][AllowEmptyString()][string]$Content
  )

  $tmpDir = Join-Path $env:TEMP "pcs_sync_remote_cmds"
  if (-not (Test-Path $tmpDir)) {
    New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null
  }

  $fileName = "{0}_{1}_{2}.sh" -f $Prefix, (Get-Date -Format "yyyyMMdd_HHmmss"), ([System.Guid]::NewGuid().ToString("N").Substring(0, 8))
  $filePath = Join-Path $tmpDir $fileName
  $encoding = [System.Text.UTF8Encoding]::new($false)
  [System.IO.File]::WriteAllText($filePath, $Content, $encoding)
  return $filePath
}

function Invoke-PuttySync {
  param(
    [Parameter(Mandatory=$true)][string]$LocalResolvedPath,
    [Parameter(Mandatory=$true)][string]$RemoteUser,
    [Parameter(Mandatory=$true)][string]$RemoteHost,
    [Parameter(Mandatory=$true)][string]$RemotePath,
    [Parameter(Mandatory=$true)][int]$Port,
    [Parameter(Mandatory=$true)][string]$IdentityPath,
    [Parameter(Mandatory=$true)][bool]$IsDryRun,
    [Parameter(Mandatory=$true)][bool]$IsPreviewOnly,
    [AllowEmptyString()][string]$ExcludeFile,
    [bool]$ExcludeEvidence = $true,
    [bool]$UseCompression = $true,
    [int]$LargeTransferWarningMB = 200,
    [AllowEmptyString()][string]$ExecRelativePath = "",
    [int]$Retries = 3,
    [bool]$AutoInstallDeps = $true,
    [bool]$RunBootstrap = $true,
    [string]$BootstrapServerPort = "8080",
    [AllowEmptyString()][string]$BootstrapGoogleClientId = "",
    [AllowEmptyString()][string]$BootstrapGoogleClientSecret = "",
    [AllowEmptyString()][string]$BootstrapGoogleRedirectUrl = "",
    [AllowEmptyString()][string]$BootstrapDbDialect = "postgres",
    [AllowEmptyString()][string]$BootstrapDbEmpresasDsn = "",
    [AllowEmptyString()][string]$BootstrapDbSuperadminDsn = "",
    [bool]$RestartServer = $true,
    [string]$RestartBinaryRelativePath = "backend/bin/server_linux_amd64",
    [string]$RestartStdoutLogRelativePath = "backend/server.log",
    [string]$RestartStderrLogRelativePath = "backend/server.err",
    [int]$RestartHealthTimeout = 45
  )

  if (-not (Test-Path $IdentityPath)) {
    throw "No se encontró la clave de identidad para fallback sin WSL: $IdentityPath"
  }

  if (-not (Test-TcpPortReachable -TargetHost $RemoteHost -Port $Port -TimeoutMs 8000)) {
    throw "No hay conectividad TCP con ${RemoteHost}:$Port desde este equipo. Verifica red, firewall y acceso SSH al VPS."
  }

  $remoteTarget = "$RemoteUser@$RemoteHost"
  $identityResolved = (Resolve-Path $IdentityPath).Path
  $isPpkIdentity = ([System.IO.Path]::GetExtension($identityResolved).ToLowerInvariant() -eq ".ppk")
  $transportLabel = "OpenSSH"
  $excludePatterns = Get-SyncExcludePatterns -ExcludeFile $ExcludeFile -ExcludeEvidence $ExcludeEvidence

  $tmpDir = Join-Path $LocalResolvedPath ".gotmp\pcs_sync_staging"
  if (-not (Test-Path $tmpDir)) {
    New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null
  }
  Get-ChildItem -LiteralPath $tmpDir -Filter "pcs_sync_*.tar*" -File -ErrorAction SilentlyContinue |
    Where-Object { $_.LastWriteTime -lt (Get-Date).AddDays(-2) } |
    Remove-Item -Force -ErrorAction SilentlyContinue

  $stamp = Get-Date -Format "yyyyMMdd_HHmmss"
  $archiveExtension = if ($UseCompression) { ".tar.gz" } else { ".tar" }
  $archivePath = Join-Path $tmpDir ("pcs_sync_" + $stamp + $archiveExtension)
  $remoteArchive = "/tmp/pcs_sync_$stamp$archiveExtension"

  $tarArgs = @()
  foreach ($pattern in $excludePatterns) {
    $tarArgs += "--exclude=$pattern"
  }
  if ($UseCompression) {
    $tarArgs += @("-czf", $archivePath, "-C", $LocalResolvedPath, ".")
  } else {
    $tarArgs += @("-cf", $archivePath, "-C", $LocalResolvedPath, ".")
  }

  $mkdirCmd = "mkdir -p '$RemotePath'"
  $extractTarFlag = if ($UseCompression) { "-xzf" } else { "-xf" }
  $cleanBackendSourceCmd = "if [ -d '$RemotePath/backend' ]; then find '$RemotePath/backend' -mindepth 1 -maxdepth 1 ! -name '.env' ! -name '.env.local' ! -name 'logs' ! -name 'bin' ! -name 'tmp' ! -name 'secure' -exec rm -rf {} +; fi"
  $extractCmd = "mkdir -p '$RemotePath' && $cleanBackendSourceCmd && if tar --version 2>/dev/null | grep -qi 'GNU tar'; then tar --warning=no-unknown-keyword $extractTarFlag '$remoteArchive' -C '$RemotePath'; else tar $extractTarFlag '$remoteArchive' -C '$RemotePath'; fi && rm -f '$remoteArchive'"
  if (-not [string]::IsNullOrWhiteSpace($ExecRelativePath)) {
    $remoteExecPath = ($RemotePath.TrimEnd('/') + "/" + $ExecRelativePath.TrimStart('/'))
    $extractCmd += " && if [ -f '$remoteExecPath' ]; then chmod +x '$remoteExecPath'; fi"
  }
  $verifyCommandPath = ""
  $uploadCommandPath = ""
  $extractCommandPath = ""
  $bootstrapCommandPath = ""
  $verifyArgs = @()
  $uploadArgs = @()
  $extractArgs = @()
  $bootstrapCmd = ""
  $bootstrapArgs = @()
  $restartCmd = ""
  $restartArgs = @()
  $tempCommandFiles = @()

  if ($isPpkIdentity) {
    $tools = Ensure-PuttyTools -AutoInstall $AutoInstallDeps
    $plink = $tools.Plink
    $pscp = $tools.Pscp
    $transportLabel = "PuTTY"

    $verifyCommandPath = $plink
    $uploadCommandPath = $pscp
    $extractCommandPath = $plink
    $bootstrapCommandPath = $plink

    $plinkBaseArgs = @('-batch') + $script:PlinkHostKeyArgs + @('-P', "$Port", '-i', $identityResolved)
    $verifyArgs = $plinkBaseArgs + @($remoteTarget, $mkdirCmd)
    $uploadArgs = $plinkBaseArgs + @($archivePath, "${remoteTarget}:$remoteArchive")
    $extractArgs = $plinkBaseArgs + @($remoteTarget, $extractCmd)
  } else {
    $sshExe = Resolve-SshExe
    $scpExe = Resolve-ScpExe
    if (-not $sshExe -or -not $scpExe) {
      throw "No se encontraron ssh.exe/scp.exe para usar la clave OpenSSH. Instala OpenSSH Client de Windows o usa una clave .ppk con PuTTY."
    }

    $sshCommonArgs = @(
      '-o', 'BatchMode=yes',
      '-o', 'StrictHostKeyChecking=accept-new',
      '-o', 'ConnectTimeout=15',
      '-o', 'ServerAliveInterval=30',
      '-o', 'ServerAliveCountMax=4',
      '-p', "$Port",
      '-i', $identityResolved
    )

    $verifyCommandPath = $sshExe
    $uploadCommandPath = $scpExe
    $extractCommandPath = $sshExe
    $bootstrapCommandPath = $sshExe

    $verifyArgs = $sshCommonArgs + @($remoteTarget, $mkdirCmd)
    $uploadArgs = @(
      '-o', 'BatchMode=yes',
      '-o', 'StrictHostKeyChecking=accept-new',
      '-o', 'ConnectTimeout=15',
      '-o', 'ServerAliveInterval=30',
      '-o', 'ServerAliveCountMax=4',
      '-P', "$Port",
      '-i', $identityResolved,
      $archivePath,
      "${remoteTarget}:$remoteArchive"
    )
    $extractArgs = $sshCommonArgs + @($remoteTarget, $extractCmd)
  }

  if ($RunBootstrap) {
    $bootstrapCmd = Get-RemoteBootstrapCommand -RemotePath $RemotePath -ServerPort $BootstrapServerPort -GoogleClientId $BootstrapGoogleClientId -GoogleClientSecret $BootstrapGoogleClientSecret -GoogleRedirectUrl $BootstrapGoogleRedirectUrl -DbDialect $BootstrapDbDialect -DbEmpresasDsn $BootstrapDbEmpresasDsn -DbSuperadminDsn $BootstrapDbSuperadminDsn
    if ($isPpkIdentity) {
      $bootstrapScriptPath = New-TempRemoteCommandFile -Prefix "bootstrap_remote" -Content $bootstrapCmd
      $tempCommandFiles += $bootstrapScriptPath
      $bootstrapArgs = $plinkBaseArgs + @('-m', $bootstrapScriptPath, $remoteTarget)
    } else {
      $bootstrapArgs = @(
        '-o', 'BatchMode=yes',
        '-o', 'StrictHostKeyChecking=accept-new',
        '-o', 'ConnectTimeout=15',
        '-o', 'ServerAliveInterval=30',
        '-o', 'ServerAliveCountMax=4',
        '-p', "$Port",
        '-i', $identityResolved,
        $remoteTarget,
        $bootstrapCmd
      )
    }
  }

  if ($RestartServer) {
    $restartCmd = Get-RemoteRestartCommand -RemotePath $RemotePath -BinaryRelativePath $RestartBinaryRelativePath -ServerPort $BootstrapServerPort -StdoutLogRelativePath $RestartStdoutLogRelativePath -StderrLogRelativePath $RestartStderrLogRelativePath -HealthTimeoutSeconds $RestartHealthTimeout
    if ($isPpkIdentity) {
      $restartScriptPath = New-TempRemoteCommandFile -Prefix "restart_remote" -Content $restartCmd
      $tempCommandFiles += $restartScriptPath
      $restartArgs = $plinkBaseArgs + @('-m', $restartScriptPath, $remoteTarget)
    } else {
      $restartArgs = @(
        '-o', 'BatchMode=yes',
        '-o', 'StrictHostKeyChecking=accept-new',
        '-o', 'ConnectTimeout=15',
        '-o', 'ServerAliveInterval=30',
        '-o', 'ServerAliveCountMax=4',
        '-p', "$Port",
        '-i', $identityResolved,
        $remoteTarget,
        $restartCmd
      )
    }
  }

  if ($IsPreviewOnly) {
    Write-Host ("[PREVIEW] Fallback sin WSL (" + $transportLabel + "):")
    Write-Host ("tar " + (($tarArgs | ForEach-Object { '"' + $_ + '"' }) -join " "))
    Write-Host ($verifyCommandPath + " " + (($verifyArgs | ForEach-Object { '"' + $_ + '"' }) -join " "))
    Write-Host ($uploadCommandPath + " " + (($uploadArgs | ForEach-Object { '"' + $_ + '"' }) -join " "))
    Write-Host ($extractCommandPath + " " + (($extractArgs | ForEach-Object { '"' + $_ + '"' }) -join " "))
    if ($RunBootstrap) {
      Write-Host ($bootstrapCommandPath + " " + (($bootstrapArgs | ForEach-Object { '"' + $_ + '"' }) -join " "))
    }
    if ($RestartServer) {
      Write-Host ($bootstrapCommandPath + " " + (($restartArgs | ForEach-Object { '"' + $_ + '"' }) -join " "))
    }
    return
  }

  try {
    Write-Host "[INFO] Empaquetando proyecto local para sincronización..."
    & tar @tarArgs
    if ($LASTEXITCODE -ne 0 -or -not (Test-Path $archivePath)) {
      throw "Falló la creación del paquete TAR local (código $LASTEXITCODE)."
    }

    $archiveInfo = Get-Item -LiteralPath $archivePath -ErrorAction Stop
    $archiveSizeText = Format-ByteSize -Bytes ([long]$archiveInfo.Length)
    Write-Host ("[INFO] Paquete listo para subir: " + $archivePath)
    Write-Host ("[INFO] Tamano del paquete: " + $archiveSizeText)
    if ($LargeTransferWarningMB -gt 0 -and $archiveInfo.Length -gt ([int64]$LargeTransferWarningMB * 1024 * 1024)) {
      Write-Warning ("El paquete supera " + $LargeTransferWarningMB + " MB. La subida puede tardar; revisa -DryRun o agrega exclusiones con -ExcludeFile si no esperabas tantos datos.")
    }

    if ($IsDryRun) {
      Write-Host "[INFO] Modo DryRun (sin cambios remotos)."
      $entries = & tar -tf $archivePath
      $count = ($entries | Measure-Object).Count
      Write-Host ("[INFO] Tamano que se intentaria subir: " + $archiveSizeText)
      Write-Host ("[INFO] Archivos que se transferirían: " + $count)
      $entries | Select-Object -First 40 | ForEach-Object { Write-Host ("  - " + $_) }
      if ($count -gt 40) {
        Write-Host "  ..."
      }
      return
    }

    Invoke-ExternalWithRetry -Label "verificación remota" -CommandPath $verifyCommandPath -Arguments $verifyArgs -MaxAttempts $Retries -RetryOnTimeoutOnly
    Write-Host ("[INFO] Subiendo paquete a " + $remoteTarget + ":" + $remoteArchive + " (" + $archiveSizeText + ")...")
    Invoke-ExternalWithRetry -Label "subida de paquete" -CommandPath $uploadCommandPath -Arguments $uploadArgs -MaxAttempts $Retries -RetryOnTimeoutOnly

    if (-not [string]::IsNullOrWhiteSpace($ExecRelativePath)) {
      Write-Host ("[INFO] Aplicando permiso ejecutable a: " + $ExecRelativePath)
    }
    Invoke-ExternalWithRetry -Label "extracción remota" -CommandPath $extractCommandPath -Arguments $extractArgs -MaxAttempts $Retries -RetryOnTimeoutOnly

    if ($RunBootstrap) {
      Invoke-ExternalWithRetry -Label "bootstrap remoto" -CommandPath $bootstrapCommandPath -Arguments $bootstrapArgs -MaxAttempts $Retries -RetryOnTimeoutOnly
    }

    if ($RestartServer) {
      Invoke-ExternalWithRetry -Label "redeploy remoto" -CommandPath $bootstrapCommandPath -Arguments $restartArgs -MaxAttempts $Retries -RetryOnTimeoutOnly
    }

    Write-Host ("[OK] Sincronización completada por fallback sin WSL (" + $transportLabel + ").")
  }
  finally {
    if (Test-Path $archivePath) {
      Remove-Item -LiteralPath $archivePath -Force -ErrorAction SilentlyContinue
    }
    foreach ($tempFile in $tempCommandFiles) {
      if ($tempFile -and (Test-Path $tempFile)) {
        Remove-Item -LiteralPath $tempFile -Force -ErrorAction SilentlyContinue
      }
    }
  }
}

function Resolve-IdentityContext {
  param([Parameter(Mandatory=$true)][string]$IdentityPath)

  if (-not (Test-Path $IdentityPath)) {
    throw "No se encontró la clave de identidad: $IdentityPath"
  }

  $resolvedPath = (Resolve-Path $IdentityPath).Path
  $context = [ordered]@{
    Mode = "ssh"
    IdentityPath = $resolvedPath
    PlinkExeWsl = ""
    PlinkKeyWin = ""
  }

  if ([System.IO.Path]::GetExtension($resolvedPath).ToLowerInvariant() -ne ".ppk") {
    return [pscustomobject]$context
  }

  $puttygen = Resolve-PuttyGen
  if ($puttygen) {
    $tmpDir = Join-Path $env:TEMP "pcs_sync_keys"
    if (-not (Test-Path $tmpDir)) {
      New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null
    }

    $base = [System.IO.Path]::GetFileNameWithoutExtension($resolvedPath)
    $converted = Join-Path $tmpDir ($base + "_openssh")
    & $puttygen $resolvedPath -O private-openssh -o $converted | Out-Null
    if ($LASTEXITCODE -eq 0 -and (Test-Path $converted)) {
      $context.IdentityPath = $converted
      return [pscustomobject]$context
    }
  }

  $plinkExe = Resolve-Plink
  if (-not $plinkExe) {
    throw "La clave es .ppk y no se encontró plink.exe para usarla directamente ni fue posible convertirla con puttygen."
  }

  $context.Mode = "plink"
  $context.IdentityPath = ""
  $context.PlinkKeyWin = $resolvedPath
  $context.PlinkExeWsl = Get-WslUnixPath -WindowsPath $plinkExe -Label "plink"
  return [pscustomobject]$context
}

function Invoke-RemoteCommandSimple {
  param(
    [Parameter(Mandatory=$true)][string]$RemoteUser,
    [Parameter(Mandatory=$true)][string]$RemoteHost,
    [Parameter(Mandatory=$true)][int]$Port,
    [Parameter(Mandatory=$true)][pscustomobject]$IdentityContext,
    [Parameter(Mandatory=$true)][string]$Command
  )

  if ($IdentityContext.Mode -eq "plink") {
    $plinkExe = Resolve-Plink
    if (-not $plinkExe) {
      throw "No se encontró plink.exe para ejecutar comandos remotos."
    }
    if ([string]::IsNullOrWhiteSpace($IdentityContext.PlinkKeyWin)) {
      throw "Falta PlinkKeyWin (clave .ppk) para ejecutar comando remoto."
    }
    $plinkArgs = @('-batch') + $script:PlinkHostKeyArgs + @('-P', "$Port", '-i', $IdentityContext.PlinkKeyWin, "$RemoteUser@$RemoteHost", $Command)
    & $plinkExe @plinkArgs
    if ($LASTEXITCODE -ne 0) {
      throw "Fallo comando remoto via plink (exit=$LASTEXITCODE)"
    }
    return
  }

  $sshExe = Resolve-SshExe
  if (-not $sshExe) {
    throw "No se encontró ssh.exe para ejecutar comandos remotos."
  }
  $keyPath = $IdentityContext.IdentityPath
  if ([string]::IsNullOrWhiteSpace($keyPath)) {
    throw "Falta IdentityPath para ejecutar comando remoto."
  }
  & $sshExe -o StrictHostKeyChecking=accept-new -p $Port -i $keyPath "$RemoteUser@$RemoteHost" $Command
  if ($LASTEXITCODE -ne 0) {
    throw "Fallo comando remoto via ssh (exit=$LASTEXITCODE)"
  }
}

function Invoke-RemoteDockerComposeRedeploy {
  param(
    [Parameter(Mandatory=$true)][string]$RemoteUser,
    [Parameter(Mandatory=$true)][string]$RemoteHost,
    [Parameter(Mandatory=$true)][int]$Port,
    [Parameter(Mandatory=$true)][string]$IdentityPath,
    [Parameter(Mandatory=$true)][string]$RemotePath,
    [Parameter(Mandatory=$true)][bool]$Enabled,
    [Parameter(Mandatory=$true)][bool]$IsDryRun,
    [Parameter(Mandatory=$true)][bool]$IsPreviewOnly,
    [int]$HealthTimeoutSeconds = 180
  )

  if (-not $Enabled) {
    Write-Host "[INFO] Redeploy Docker omitido por parametro -RedeployDockerStack false."
    return
  }
  if ($IsDryRun -or $IsPreviewOnly) {
    Write-Host "[INFO] Redeploy Docker omitido por DryRun/PreviewOnly."
    return
  }

  if (-not (Test-Path $IdentityPath)) {
    throw "No se encontrÃ³ la clave de identidad: $IdentityPath"
  }
  $resolvedIdentityPath = (Resolve-Path $IdentityPath).Path
  $identityContext = [pscustomobject]@{
    Mode = "ssh"
    IdentityPath = $resolvedIdentityPath
    PlinkExeWsl = ""
    PlinkKeyWin = ""
  }
  if ([System.IO.Path]::GetExtension($resolvedIdentityPath).ToLowerInvariant() -eq ".ppk") {
    $identityContext.Mode = "plink"
    $identityContext.IdentityPath = ""
    $identityContext.PlinkKeyWin = $resolvedIdentityPath
  }
  $remotePathLit = Convert-ToBashLiteral $RemotePath
  $remoteScript = @"
set -e
remote_path=$remotePathLit
if [ ! -d "`$remote_path" ]; then
  echo "[INFO] Docker redeploy omitido: no existe `$remote_path"
  exit 0
fi
cd "`$remote_path"
if ! command -v docker >/dev/null 2>&1; then
  echo "[INFO] Docker redeploy omitido: docker no esta instalado en el VPS"
  exit 0
fi
if [ ! -f deploy/docker-compose.platform.yml ] || [ ! -f deploy/.env.platform ]; then
  echo "[INFO] Docker redeploy omitido: falta deploy/docker-compose.platform.yml o deploy/.env.platform"
  exit 0
fi
echo "[INFO] Docker stack detectado. Reconstruyendo backend/frontend con Compose..."
echo "[INFO] Limpiando caches locales que no pertenecen al runtime Docker..."
rm -rf .codex-gocache .codex-tmp-go .gocache .gotmp tmp \
  backend/.codex-gocache backend/.codex-tmp-go backend/.gocache backend/.gotmp \
  backend/.tmp-go-test-* backend/tmp 2>/dev/null || true
export HEALTH_TIMEOUT_SECONDS=$HealthTimeoutSeconds
bash deploy/scripts/vps-compose-sidecar-up.sh
echo "[INFO] Revisando proxy Nginx del host para iRedMail..."
if ! bash deploy/scripts/vps-configure-iredmail-host-nginx.sh; then
  echo "[WARN] No se pudo activar el proxy host de iRedMail. Ver deploy/scripts/vps-configure-iredmail-host-nginx.sh"
fi
echo "[INFO] Estado Docker despues del redeploy:"
docker ps --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}' | grep -E 'pcs-(backend|frontend|postgres|iredmail)|NAMES' || true
"@
  $command = "bash -lc " + (Convert-ToBashLiteral $remoteScript)
  Invoke-RemoteCommandSimple -RemoteUser $RemoteUser -RemoteHost $RemoteHost -Port $Port -IdentityContext $identityContext -Command $command
}

function Invoke-RemoteUnusedFilesCleanup {
  param(
    [Parameter(Mandatory=$true)][string]$RemoteUser,
    [Parameter(Mandatory=$true)][string]$RemoteHost,
    [Parameter(Mandatory=$true)][int]$Port,
    [Parameter(Mandatory=$true)][string]$IdentityPath,
    [Parameter(Mandatory=$true)][string]$RemotePath,
    [Parameter(Mandatory=$true)][bool]$Enabled,
    [Parameter(Mandatory=$true)][bool]$IsDryRun,
    [Parameter(Mandatory=$true)][bool]$IsPreviewOnly,
    [int]$TempMinAgeMinutes = 60,
    [int]$DockerBuilderCacheMaxAgeHours = 0,
    [bool]$PruneDockerDanglingImages = $true,
    [bool]$PruneStoppedContainers = $true
  )

  if (-not $Enabled) {
    Write-Host "[INFO] Limpieza remota omitida por parametro -CleanupRemoteUnusedFiles false."
    return
  }
  if ($IsDryRun -or $IsPreviewOnly) {
    Write-Host "[INFO] Limpieza remota omitida por DryRun/PreviewOnly."
    return
  }

  if ($TempMinAgeMinutes -lt 1) {
    $TempMinAgeMinutes = 60
  }
  if ($DockerBuilderCacheMaxAgeHours -lt 0) {
    $DockerBuilderCacheMaxAgeHours = 0
  }

  if (-not (Test-Path $IdentityPath)) {
    throw "No se encontrÃ³ la clave de identidad: $IdentityPath"
  }
  $resolvedIdentityPath = (Resolve-Path $IdentityPath).Path
  $identityContext = [pscustomobject]@{
    Mode = "ssh"
    IdentityPath = $resolvedIdentityPath
    PlinkExeWsl = ""
    PlinkKeyWin = ""
  }
  if ([System.IO.Path]::GetExtension($resolvedIdentityPath).ToLowerInvariant() -eq ".ppk") {
    $identityContext.Mode = "plink"
    $identityContext.IdentityPath = ""
    $identityContext.PlinkKeyWin = $resolvedIdentityPath
  }

  $remotePathLit = Convert-ToBashLiteral $RemotePath
  $pruneImages = if ($PruneDockerDanglingImages) { "1" } else { "0" }
  $pruneContainers = if ($PruneStoppedContainers) { "1" } else { "0" }
  $remoteScript = @"
set +e
remote_path=$remotePathLit
temp_min_age=$TempMinAgeMinutes
builder_cache_max_age_hours=$DockerBuilderCacheMaxAgeHours
prune_images=$pruneImages
prune_containers=$pruneContainers

echo "[INFO] Limpieza VPS: iniciando limpieza segura de temporales y caches."
set -- `$(df -h / | tail -n 1)
echo "[INFO] Disco antes limpieza: usado=`$5 libre=`$4 total=`$2"

echo "[INFO] Limpieza VPS: eliminando paquetes temporales antiguos de sync en /tmp."
find /tmp -maxdepth 1 -type f -name 'pcs_sync_*.tar.gz' -mmin +"`$temp_min_age" -print -delete 2>/dev/null | sed 's/^/[INFO] Eliminado: /'

if [ -n "`$remote_path" ] && [ -d "`$remote_path" ]; then
  echo "[INFO] Limpieza VPS: eliminando caches locales no persistentes del proyecto."
  rm -rf "`$remote_path/.codex-gocache" "`$remote_path/.codex-tmp-go" "`$remote_path/.gocache" "`$remote_path/.gotmp" "`$remote_path/tmp" \
    "`$remote_path/backend/.codex-gocache" "`$remote_path/backend/.codex-tmp-go" "`$remote_path/backend/.gocache" "`$remote_path/backend/.gotmp" \
    "`$remote_path/backend/.tmp-go-test-"* "`$remote_path/backend/tmp" 2>/dev/null || true
fi

if command -v docker >/dev/null 2>&1; then
  echo "[INFO] Limpieza VPS: estado Docker antes de limpiar."
  docker system df 2>/dev/null || true

  if [ "`$prune_containers" = "1" ]; then
    echo "[INFO] Limpieza VPS: eliminando contenedores detenidos antiguos."
    docker container prune -f --filter "until=24h" 2>/dev/null || true
  fi

  if [ "`$prune_images" = "1" ]; then
    echo "[INFO] Limpieza VPS: eliminando imagenes dangling no usadas."
    docker image prune -f 2>/dev/null || true
  fi

  echo "[INFO] Limpieza VPS: eliminando cache BuildKit no usado."
  if [ "`$builder_cache_max_age_hours" -gt 0 ]; then
    docker builder prune -af --filter "until=`${builder_cache_max_age_hours}h" 2>/dev/null || true
  else
    docker builder prune -af 2>/dev/null || true
  fi

  echo "[INFO] Limpieza VPS: estado Docker despues de limpiar."
  docker system df 2>/dev/null || true
fi

set -- `$(df -h / | tail -n 1)
echo "[INFO] Disco despues limpieza: usado=`$5 libre=`$4 total=`$2"
echo "[OK] Limpieza VPS completada sin tocar volumenes ni bases de datos."
"@
  $command = "bash -lc " + (Convert-ToBashLiteral $remoteScript)
  Invoke-RemoteCommandSimple -RemoteUser $RemoteUser -RemoteHost $RemoteHost -Port $Port -IdentityContext $identityContext -Command $command
}

try {
  $scriptPath = Join-Path $PSScriptRoot "sync_to_vps.sh"
  if (-not (Test-Path $scriptPath)) {
    throw "No se encuentra el script base: $scriptPath"
  }

  if (-not $LocalPath) {
    $LocalPath = (Resolve-Path (Join-Path $PSScriptRoot "..\")).Path
  } else {
    $LocalPath = (Resolve-Path $LocalPath).Path
  }

  $identityWasProvided = $PSBoundParameters.ContainsKey("IdentityFile") -and -not [string]::IsNullOrWhiteSpace($IdentityFile)
  if (-not $identityWasProvided) {
    $IdentityFile = Resolve-DefaultIdentityFile -RepoRoot $LocalPath
    Write-Host ("[INFO] IdentityFile no especificado. Usando: " + $IdentityFile)
  } elseif (-not [System.IO.Path]::IsPathRooted($IdentityFile)) {
    $repoIdentity = Join-Path $LocalPath $IdentityFile
    if (Test-Path $repoIdentity) {
      $IdentityFile = (Resolve-Path $repoIdentity).Path
    }
  }

  $localBackendEnvPath = Join-Path $LocalPath "backend\.env.local"

  if ([string]::IsNullOrWhiteSpace($DbDialect)) {
    $DbDialect = Get-DotEnvValue -EnvFilePath $localBackendEnvPath -Key "DB_DIALECT"
    if ([string]::IsNullOrWhiteSpace($DbDialect)) {
      $DbDialect = [Environment]::GetEnvironmentVariable("DB_DIALECT")
    }
    if ([string]::IsNullOrWhiteSpace($DbDialect)) {
      $DbDialect = [Environment]::GetEnvironmentVariable("DB_ENGINE")
    }
    if ([string]::IsNullOrWhiteSpace($DbDialect)) {
      $DbDialect = [Environment]::GetEnvironmentVariable("PCS_DB_DIALECT")
    }
  }

  if ([string]::IsNullOrWhiteSpace($DbEmpresasDsn)) {
    $DbEmpresasDsn = Get-DotEnvValue -EnvFilePath $localBackendEnvPath -Key "DB_EMPRESAS_DSN"
    if ([string]::IsNullOrWhiteSpace($DbEmpresasDsn)) {
      $DbEmpresasDsn = [Environment]::GetEnvironmentVariable("DB_EMPRESAS_DSN")
    }
    if ([string]::IsNullOrWhiteSpace($DbEmpresasDsn)) {
      $DbEmpresasDsn = [Environment]::GetEnvironmentVariable("PCS_DB_EMPRESAS_DSN")
    }
  }

  if ([string]::IsNullOrWhiteSpace($DbSuperadminDsn)) {
    $DbSuperadminDsn = Get-DotEnvValue -EnvFilePath $localBackendEnvPath -Key "DB_SUPERADMIN_DSN"
    if ([string]::IsNullOrWhiteSpace($DbSuperadminDsn)) {
      $DbSuperadminDsn = [Environment]::GetEnvironmentVariable("DB_SUPERADMIN_DSN")
    }
    if ([string]::IsNullOrWhiteSpace($DbSuperadminDsn)) {
      $DbSuperadminDsn = [Environment]::GetEnvironmentVariable("PCS_DB_SUPERADMIN_DSN")
    }
  }

  $tunnelEnabled = Parse-BoolLike (Get-DotEnvValue -EnvFilePath $localBackendEnvPath -Key "DB_VPS_TUNNEL_ENABLED")
  if ($tunnelEnabled) {
    $tunnelLocalPort = Parse-IntOrDefault -Value (Get-DotEnvValue -EnvFilePath $localBackendEnvPath -Key "DB_VPS_LOCAL_PORT") -DefaultValue 15432
    $tunnelRemotePort = Parse-IntOrDefault -Value (Get-DotEnvValue -EnvFilePath $localBackendEnvPath -Key "DB_VPS_REMOTE_PORT") -DefaultValue 5432

    $normalizedEmpresasDsn = Normalize-PostgresDsnForVps -Dsn $DbEmpresasDsn -TunnelLocalPort $tunnelLocalPort -RemoteDbPort $tunnelRemotePort
    if ($normalizedEmpresasDsn -ne $DbEmpresasDsn) {
      Write-Host ("[INFO] Normalizando DB_EMPRESAS_DSN para despliegue VPS (" + $tunnelLocalPort + " -> " + $tunnelRemotePort + ").")
      $DbEmpresasDsn = $normalizedEmpresasDsn
    }

    $normalizedSuperDsn = Normalize-PostgresDsnForVps -Dsn $DbSuperadminDsn -TunnelLocalPort $tunnelLocalPort -RemoteDbPort $tunnelRemotePort
    if ($normalizedSuperDsn -ne $DbSuperadminDsn) {
      Write-Host ("[INFO] Normalizando DB_SUPERADMIN_DSN para despliegue VPS (" + $tunnelLocalPort + " -> " + $tunnelRemotePort + ").")
      $DbSuperadminDsn = $normalizedSuperDsn
    }
  }

  $builtBinaryAbs = ""
  $builtBinaryRel = ""
  $restartBinaryRel = (($RemoteBinaryPath -replace "\\", "/").TrimStart('/'))
  if ([string]::IsNullOrWhiteSpace($restartBinaryRel)) {
    $restartBinaryRel = "backend/bin/server_linux_amd64"
  }

  if (-not [string]::IsNullOrWhiteSpace($BuildOutput)) {
    if ([System.IO.Path]::IsPathRooted($BuildOutput)) {
      try {
        $candidateAbs = Resolve-AbsolutePath -BasePath $LocalPath -InputPath $BuildOutput
        $builtBinaryRel = Get-RelativePathIfInside -BasePath $LocalPath -TargetPath $candidateAbs
      }
      catch {
        $builtBinaryRel = ""
      }
    } else {
      $builtBinaryRel = ($BuildOutput -replace "\\", "/").TrimStart('.', '/')
    }
  }

  if ($SkipBuild -and $BuildOnly) {
    throw "No puedes usar -SkipBuild y -BuildOnly al mismo tiempo."
  }

  if ($BuildOnly -and $PreviewOnly) {
    throw "No puedes usar -BuildOnly y -PreviewOnly al mismo tiempo."
  }

  $effectiveSkipBuild = $SkipBuild.IsPresent
  $effectiveRestartRemoteServer = [bool]$RestartRemoteServer
  $effectiveRedeployDockerStack = [bool]$RedeployDockerStack
  switch ($DeploymentMode) {
    "docker" {
      $effectiveSkipBuild = $true
      $effectiveRestartRemoteServer = $false
      $effectiveRedeployDockerStack = $true
      Write-Host "[INFO] DeploymentMode=docker: Docker Compose construye y reinicia backend/frontend; se omite systemd."
    }
    "legacy" {
      $effectiveRedeployDockerStack = $false
      Write-Host "[INFO] DeploymentMode=legacy: se usa binario/systemd y se omite Docker Compose."
    }
    default {
      Write-Host "[INFO] DeploymentMode=hybrid: se actualiza binario/systemd y tambien Docker Compose."
    }
  }
  if ($BuildOnly) {
    $effectiveSkipBuild = $false
  }

  if (-not $effectiveSkipBuild) {
    if ($PreviewOnly) {
      Write-Host "[INFO] PreviewOnly: compilación Linux omitida (usa -BuildOnly para compilar sin sincronizar)."
    } else {
      $builtBinaryAbs = Invoke-LocalLinuxBuild -RepoRoot $LocalPath -WorkingDir $BuildWorkingDir -Package $BuildPackage -OutputPath $BuildOutput -GoOS $BuildGoOS -GoArch $BuildGoArch -CgoEnabled $BuildCgoEnabled
      if ([string]::IsNullOrWhiteSpace($builtBinaryRel) -and -not [string]::IsNullOrWhiteSpace($builtBinaryAbs)) {
        $builtBinaryRel = Get-RelativePathIfInside -BasePath $LocalPath -TargetPath $builtBinaryAbs
      }
      if (-not [string]::IsNullOrWhiteSpace($builtBinaryRel)) {
        $restartBinaryRel = $builtBinaryRel
      }
    }
  } else {
    if ($DeploymentMode -eq "docker" -and -not $SkipBuild.IsPresent) {
      Write-Host "[INFO] Compilación Linux omitida porque Docker construye el backend dentro del contenedor."
    } else {
      Write-Host "[INFO] Compilación Linux omitida por parámetro -SkipBuild."
    }
    if (-not [string]::IsNullOrWhiteSpace($builtBinaryRel)) {
      $restartBinaryRel = $builtBinaryRel
    }
  }

  if ($BuildOnly) {
    Write-Host "[OK] BuildOnly completado. No se ejecutó sincronización."
    return
  }

  if ($BootstrapServer) {
    $dialectDisplay = if ([string]::IsNullOrWhiteSpace($DbDialect)) { "EMPTY" } else { $DbDialect }
    $empDisplay = if ([string]::IsNullOrWhiteSpace($DbEmpresasDsn)) { "EMPTY" } else { "SET" }
    $superDisplay = if ([string]::IsNullOrWhiteSpace($DbSuperadminDsn)) { "EMPTY" } else { "SET" }
    Write-Host ("[INFO] Bootstrap DB config: DB_DIALECT=" + $dialectDisplay + " DB_EMPRESAS_DSN=" + $empDisplay + " DB_SUPERADMIN_DSN=" + $superDisplay)
  }

  if (-not (Test-WslReady)) {
    Invoke-PuttySync -LocalResolvedPath $LocalPath -RemoteUser $RemoteUser -RemoteHost $RemoteHost -RemotePath $RemotePath -Port $Port -IdentityPath $IdentityFile -IsDryRun $DryRun.IsPresent -IsPreviewOnly $PreviewOnly.IsPresent -ExcludeFile $ExcludeFile -ExcludeEvidence $ExcludeEvidenceFromPackage -UseCompression $CompressPackage -LargeTransferWarningMB $LargeTransferWarningMB -ExecRelativePath $builtBinaryRel -Retries $RetryCount -AutoInstallDeps $AutoInstallDependencies -RunBootstrap $BootstrapServer -BootstrapServerPort $ServerPort -BootstrapGoogleClientId $GoogleClientId -BootstrapGoogleClientSecret $GoogleClientSecret -BootstrapGoogleRedirectUrl $GoogleRedirectUrl -BootstrapDbDialect $DbDialect -BootstrapDbEmpresasDsn $DbEmpresasDsn -BootstrapDbSuperadminDsn $DbSuperadminDsn -RestartServer $effectiveRestartRemoteServer -RestartBinaryRelativePath $restartBinaryRel -RestartStdoutLogRelativePath $RemoteStdoutLogPath -RestartStderrLogRelativePath $RemoteStderrLogPath -RestartHealthTimeout $RestartHealthTimeoutSeconds

    Invoke-RemoteDockerComposeRedeploy -RemoteUser $RemoteUser -RemoteHost $RemoteHost -Port $Port -IdentityPath $IdentityFile -RemotePath $RemotePath -Enabled $effectiveRedeployDockerStack -IsDryRun $DryRun.IsPresent -IsPreviewOnly $PreviewOnly.IsPresent -HealthTimeoutSeconds $DockerHealthTimeoutSeconds

    Invoke-RemoteUnusedFilesCleanup -RemoteUser $RemoteUser -RemoteHost $RemoteHost -Port $Port -IdentityPath $IdentityFile -RemotePath $RemotePath -Enabled $CleanupRemoteUnusedFiles -IsDryRun $DryRun.IsPresent -IsPreviewOnly $PreviewOnly.IsPresent -TempMinAgeMinutes $RemoteCleanupTempMinAgeMinutes -DockerBuilderCacheMaxAgeHours $RemoteCleanupDockerBuilderCacheMaxAgeHours -PruneDockerDanglingImages $RemoteCleanupDockerDanglingImages -PruneStoppedContainers $RemoteCleanupStoppedContainers

    if ($OpenPublicUrlAfterDeploy -and -not $DryRun.IsPresent -and -not $PreviewOnly.IsPresent) {
      $deployUrl = Resolve-PublicDeployUrl -PublicBaseUrl $PublicBaseUrl -RemoteHost $RemoteHost -ServerPort $ServerPort
      Write-Host ("[INFO] Abriendo URL pública: " + $deployUrl)
      try {
        Start-Process $deployUrl | Out-Null
      }
      catch {
        Write-Warning ("No se pudo abrir el navegador automáticamente: " + $_.Exception.Message)
      }
    }
    return
  }

  $scriptWsl = Get-WslUnixPath -WindowsPath $scriptPath -Label "script"
  $localWsl = Get-WslUnixPath -WindowsPath $LocalPath -Label "local"

  $identityContext = [pscustomobject]@{
    Mode = "ssh"
    IdentityPath = $IdentityFile
    PlinkExeWsl = ""
    PlinkKeyWin = ""
  }
  $identityWsl = $IdentityFile
  if ($IdentityFile -match '^[A-Za-z]:\\' -or (Test-Path $IdentityFile)) {
    $identityContext = Resolve-IdentityContext -IdentityPath $IdentityFile
    if ($identityContext.Mode -eq "ssh" -and -not [string]::IsNullOrWhiteSpace($identityContext.IdentityPath)) {
      $identityWsl = Get-WslUnixPath -WindowsPath $identityContext.IdentityPath -Label "identity"
    } else {
      $identityWsl = ""
    }
  }

  $argList = @(
    "--local", $localWsl,
    "--host", $RemoteHost,
    "--user", $RemoteUser,
    "--remote", $RemotePath,
    "--port", "$Port"
  )

  if ($identityContext.Mode -eq "ssh" -and -not [string]::IsNullOrWhiteSpace($identityWsl)) {
    $argList += @("--identity", $identityWsl)
  }

  if ($DryRun) {
    $argList = @("--dry-run") + $argList
  }

  if ($ExcludeFile) {
    $excludePath = (Resolve-Path $ExcludeFile).Path
    $excludeWsl = Get-WslUnixPath -WindowsPath $excludePath -Label "exclude-file"
    $argList += @("--exclude-file", $excludeWsl)
  }

  if ($effectiveRestartRemoteServer) {
    $argList += @("--restart-server", "--server-port", "$ServerPort", "--remote-binary", $restartBinaryRel, "--stdout-log", $RemoteStdoutLogPath, "--stderr-log", $RemoteStderrLogPath, "--health-timeout", "$RestartHealthTimeoutSeconds")
  } else {
    $argList += @("--no-restart-server")
  }

  $escapedArgs = ($argList | ForEach-Object { Convert-ToBashLiteral $_ }) -join " "
  $bootstrapServerValue = if ($BootstrapServer) { "1" } else { "0" }
  $envParts = @(
    "BOOTSTRAP_SERVER=$(Convert-ToBashLiteral $bootstrapServerValue)",
    "GOOGLE_CLIENT_ID=$(Convert-ToBashLiteral $GoogleClientId)",
    "GOOGLE_CLIENT_SECRET=$(Convert-ToBashLiteral $GoogleClientSecret)",
    "GOOGLE_REDIRECT_URL=$(Convert-ToBashLiteral $GoogleRedirectUrl)",
    "DB_DIALECT=$(Convert-ToBashLiteral $DbDialect)",
    "DB_EMPRESAS_DSN=$(Convert-ToBashLiteral $DbEmpresasDsn)",
    "DB_SUPERADMIN_DSN=$(Convert-ToBashLiteral $DbSuperadminDsn)"
  )

  if ($identityContext.Mode -eq "plink") {
    $envParts += @(
      "SSH_CLIENT=$(Convert-ToBashLiteral 'plink')",
      "PLINK_EXE=$(Convert-ToBashLiteral $identityContext.PlinkExeWsl)",
      "PLINK_KEY_WIN=$(Convert-ToBashLiteral $identityContext.PlinkKeyWin)"
    )
    if (-not [string]::IsNullOrWhiteSpace($SshHostKey)) {
      $envParts += "PLINK_HOSTKEY=$(Convert-ToBashLiteral $SshHostKey)"
    }
  }

  $envPrefix = ""
  if ($envParts.Count -gt 0) {
    $envPrefix = ($envParts -join " ") + " "
  }

  $bashCmd = "$envPrefix" + "bash $(Convert-ToBashLiteral $scriptWsl) $escapedArgs"

  if ($PreviewOnly) {
    Write-Host "[PREVIEW] Comando que se ejecutaría en WSL:"
    Write-Host $bashCmd
    return
  }

  Write-Host "Ejecutando sincronización en WSL..."
  Write-Host $bashCmd
  $wslOutput = & wsl bash -lc $bashCmd 2>&1
  if ($wslOutput) {
    $wslOutput | ForEach-Object { Write-TaggedExternalOutput -Line "$_" }
  }

  if ($LASTEXITCODE -ne 0) {
    throw (Get-FriendlyExternalFailureMessage -Label "sincronización en WSL" -ExitCode $LASTEXITCODE -Text ($wslOutput -join "`n"))
  }

  Invoke-RemoteDockerComposeRedeploy -RemoteUser $RemoteUser -RemoteHost $RemoteHost -Port $Port -IdentityPath $IdentityFile -RemotePath $RemotePath -Enabled $effectiveRedeployDockerStack -IsDryRun $DryRun.IsPresent -IsPreviewOnly $PreviewOnly.IsPresent -HealthTimeoutSeconds $DockerHealthTimeoutSeconds

  Invoke-RemoteUnusedFilesCleanup -RemoteUser $RemoteUser -RemoteHost $RemoteHost -Port $Port -IdentityPath $IdentityFile -RemotePath $RemotePath -Enabled $CleanupRemoteUnusedFiles -IsDryRun $DryRun.IsPresent -IsPreviewOnly $PreviewOnly.IsPresent -TempMinAgeMinutes $RemoteCleanupTempMinAgeMinutes -DockerBuilderCacheMaxAgeHours $RemoteCleanupDockerBuilderCacheMaxAgeHours -PruneDockerDanglingImages $RemoteCleanupDockerDanglingImages -PruneStoppedContainers $RemoteCleanupStoppedContainers

  if ($OpenPublicUrlAfterDeploy -and -not $DryRun.IsPresent -and -not $PreviewOnly.IsPresent) {
    $deployUrl = Resolve-PublicDeployUrl -PublicBaseUrl $PublicBaseUrl -RemoteHost $RemoteHost -ServerPort $ServerPort
    Write-Host ("[INFO] Abriendo URL pública: " + $deployUrl)
    try {
      Start-Process $deployUrl | Out-Null
    }
    catch {
      Write-Warning ("No se pudo abrir el navegador automáticamente: " + $_.Exception.Message)
    }
  }
}
catch {
  $script:SyncExitCode = 1
  $errMsg = $_.Exception.Message
  if ([string]::IsNullOrWhiteSpace($errMsg)) {
    $errMsg = ($_ | Out-String).Trim()
  }
  if ([string]::IsNullOrWhiteSpace($errMsg)) {
    $errMsg = "Error desconocido durante la sincronizacion."
  }

  $innerMsg = ""
  if ($_.Exception -and $_.Exception.InnerException) {
    $innerMsg = $_.Exception.InnerException.Message
  }

  if (-not [string]::IsNullOrWhiteSpace($innerMsg)) {
    [Console]::Error.WriteLine("[ERROR] " + $errMsg + " | Inner: " + $innerMsg)
  } else {
    [Console]::Error.WriteLine("[ERROR] " + $errMsg)
  }
}

$global:LASTEXITCODE = $script:SyncExitCode
return
