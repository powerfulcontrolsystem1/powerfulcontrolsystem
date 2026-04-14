<#
.SYNOPSIS
  Wrapper PowerShell para sincronizar con VPS Linux.

.DESCRIPTION
  Prioriza WSL cuando está disponible y usa fallback nativo en Windows
  (OpenSSH ssh/scp para claves OpenSSH, PuTTY plink/pscp para .ppk)
  cuando WSL no está instalado o no tiene distribuciones.
  No programa tareas; se ejecuta manualmente cuando el usuario lo necesite.
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
  [string]$IdentityFile = "",
  [string]$ExcludeFile = "",
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
  [string]$GoogleRedirectUrl = "",
  [string]$DbDialect = "postgres",
  [string]$DbEmpresasDsn = "",
  [string]$DbSuperadminDsn = "",
  [bool]$RestartRemoteServer = $true,
  [string]$RemoteBinaryPath = "backend/bin/server_linux_amd64",
  [string]$RemoteStdoutLogPath = "backend/server.log",
  [string]$RemoteStderrLogPath = "backend/server.err",
  [int]$RestartHealthTimeoutSeconds = 45,
  [bool]$OpenPublicUrlAfterDeploy = $true
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$script:SyncExitCode = 0

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

  $template = @'
set -e;
if command -v apt-get >/dev/null 2>&1; then
  export DEBIAN_FRONTEND=noninteractive;
  apt-get update -y >/dev/null 2>&1 || true;
  apt-get install -y ca-certificates curl sqlite3 >/dev/null 2>&1 || true;
fi;
backend_dir=__BACKEND_DIR__;
env_file="$backend_dir/.env.local";
mkdir -p "$backend_dir";
touch "$env_file";
chmod 600 "$env_file" || true;
if ! grep -q '^SERVER_PORT=' "$env_file" 2>/dev/null; then echo SERVER_PORT=__SERVER_PORT__ >> "$env_file"; fi;
gid=__GOOGLE_ID__;
gsec=__GOOGLE_SECRET__;
grurl=__GOOGLE_REDIRECT_URL__;
dbdialect=__DB_DIALECT__;
dbemp=__DB_EMPRESAS_DSN__;
dbsuper=__DB_SUPERADMIN_DSN__;
current_dbdialect="$(grep -E '^DB_DIALECT=' "$env_file" | tail -n1 | cut -d= -f2- || true)";
current_dbemp="$(grep -E '^DB_EMPRESAS_DSN=' "$env_file" | tail -n1 | cut -d= -f2- || true)";
current_dbsuper="$(grep -E '^DB_SUPERADMIN_DSN=' "$env_file" | tail -n1 | cut -d= -f2- || true)";
effective_dbdialect="$dbdialect";
effective_dbemp="$dbemp";
effective_dbsuper="$dbsuper";
if [ -z "$effective_dbdialect" ]; then effective_dbdialect="$current_dbdialect"; fi;
if [ -z "$effective_dbemp" ]; then effective_dbemp="$current_dbemp"; fi;
if [ -z "$effective_dbsuper" ]; then effective_dbsuper="$current_dbsuper"; fi;
if [ -z "$effective_dbdialect" ] && { [ -n "$effective_dbemp" ] || [ -n "$effective_dbsuper" ]; }; then
  effective_dbdialect=postgres;
fi;
if [ "$effective_dbdialect" = "postgres" ] && { [ -z "$effective_dbemp" ] || [ -z "$effective_dbsuper" ]; }; then
  echo "BOOTSTRAP_ERROR:POSTGRES_MISSING_DSN";
  echo "BOOTSTRAP_HINT:Define DB_EMPRESAS_DSN and DB_SUPERADMIN_DSN";
  exit 1;
fi;
for key in DB_DIALECT DB_EMPRESAS_DSN DB_SUPERADMIN_DSN; do
  grep -v "^$key=" "$env_file" > "$env_file.tmp" 2>/dev/null || true;
  mv "$env_file.tmp" "$env_file" 2>/dev/null || true;
done;
if [ -n "$effective_dbdialect" ]; then echo "DB_DIALECT=$effective_dbdialect" >> "$env_file"; fi;
if [ -n "$effective_dbemp" ]; then echo "DB_EMPRESAS_DSN=$effective_dbemp" >> "$env_file"; fi;
if [ -n "$effective_dbsuper" ]; then echo "DB_SUPERADMIN_DSN=$effective_dbsuper" >> "$env_file"; fi;
if [ -n "$gid" ]; then
  grep -v '^GOOGLE_CLIENT_ID=' "$env_file" > "$env_file.tmp" 2>/dev/null || true;
  mv "$env_file.tmp" "$env_file" 2>/dev/null || true;
  echo "GOOGLE_CLIENT_ID=$gid" >> "$env_file";
fi;
if [ -n "$gsec" ]; then
  grep -v '^GOOGLE_CLIENT_SECRET=' "$env_file" > "$env_file.tmp" 2>/dev/null || true;
  mv "$env_file.tmp" "$env_file" 2>/dev/null || true;
  echo "GOOGLE_CLIENT_SECRET=$gsec" >> "$env_file";
fi;
if [ -n "$grurl" ]; then
  grep -v '^GOOGLE_REDIRECT_URL=' "$env_file" > "$env_file.tmp" 2>/dev/null || true;
  mv "$env_file.tmp" "$env_file" 2>/dev/null || true;
  echo "GOOGLE_REDIRECT_URL=$grurl" >> "$env_file";
fi;
for k in DB_DIALECT DB_SUPERADMIN_DSN DB_EMPRESAS_DSN GOOGLE_CLIENT_ID GOOGLE_CLIENT_SECRET GOOGLE_REDIRECT_URL SERVER_PORT CONFIG_ENC_KEY; do
  line="$(grep -E "^$k=" "$env_file" | tail -n1 || true)";
  if [ -z "$line" ]; then
    echo "BOOTSTRAP_WARN:$k=MISSING";
  else
    val="${line#*=}";
    if [ -z "$val" ]; then
      echo "BOOTSTRAP_WARN:$k=EMPTY";
    else
      echo "BOOTSTRAP_OK:$k=SET";
    fi;
  fi;
done
'@

  $cmd = $template.Replace("__BACKEND_DIR__", $backendDirLit).Replace("__SERVER_PORT__", $safePort).Replace("__GOOGLE_ID__", $googleIdLit).Replace("__GOOGLE_SECRET__", $googleSecretLit).Replace("__GOOGLE_REDIRECT_URL__", $googleRedirectLit).Replace("__DB_DIALECT__", $dbDialectLit).Replace("__DB_EMPRESAS_DSN__", $dbEmpresasDsnLit).Replace("__DB_SUPERADMIN_DSN__", $dbSuperadminDsnLit)
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
  $safeTimeout = if ($HealthTimeoutSeconds -lt 5) { 5 } elseif ($HealthTimeoutSeconds -gt 300) { 300 } else { $HealthTimeoutSeconds }

  $template = @'
set -e;
repo_dir=__REPO_DIR__;
bin_rel=__BIN_REL__;
stdout_rel=__STDOUT_REL__;
stderr_rel=__STDERR_REL__;
port=__PORT__;
health_timeout=__HEALTH_TIMEOUT__;
bin_path=$repo_dir/$bin_rel;
bin_name=$(basename $bin_rel);
stdout_log=$repo_dir/$stdout_rel;
stderr_log=$repo_dir/$stderr_rel;
pid_file=$repo_dir/backend/server.pid;
mkdir -p $(dirname $stdout_log) $(dirname $stderr_log);
if [ ! -f $bin_path ]; then
  echo DEPLOY_ERROR:bin_not_found path=$bin_path;
  exit 1;
fi;
chmod +x $bin_path || true;
old_pid=0;
if [ -f $pid_file ]; then
  old_pid=$(cat $pid_file 2>/dev/null || echo 0);
fi;
if [ ${old_pid:-0} -gt 0 ] 2>/dev/null && kill -0 $old_pid 2>/dev/null; then
  kill $old_pid 2>/dev/null || true;
  for i in $(seq 1 15); do
    kill -0 $old_pid 2>/dev/null || break;
    sleep 1;
  done;
  if kill -0 $old_pid 2>/dev/null; then
    kill -9 $old_pid 2>/dev/null || true;
  fi;
fi;
for pid in $(pgrep -f $bin_name 2>/dev/null || true); do
  if [ ${pid:-0} -le 0 ] 2>/dev/null; then continue; fi;
  if [ $pid -eq $$ ] 2>/dev/null || [ $pid -eq $PPID ] 2>/dev/null; then continue; fi;
  kill $pid 2>/dev/null || true;
done;
sleep 1;
for pid in $(pgrep -f $bin_name 2>/dev/null || true); do
  if [ ${pid:-0} -le 0 ] 2>/dev/null; then continue; fi;
  if [ $pid -eq $$ ] 2>/dev/null || [ $pid -eq $PPID ] 2>/dev/null; then continue; fi;
  kill -9 $pid 2>/dev/null || true;
done;
nohup $bin_path >> $stdout_log 2>> $stderr_log < /dev/null &
new_pid=$!;
echo $new_pid > $pid_file;
healthy=0;
for i in $(seq 1 $health_timeout); do
  if ! kill -0 $new_pid 2>/dev/null; then
    echo DEPLOY_ERROR:process_not_running pid=$new_pid port=$port;
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
    if kill -0 $new_pid 2>/dev/null; then
      healthy=1;
      break;
    fi;
  fi;
  sleep 1;
done;
if [ $healthy -eq 1 ]; then
  echo DEPLOY_OK:pid=$new_pid port=$port;
else
  echo DEPLOY_WARN:healthcheck_timeout pid=$new_pid port=$port;
fi;
'@

  $cmd = $template.Replace("__REPO_DIR__", $repoDirLit).Replace("__BIN_REL__", $binaryRelLit).Replace("__STDOUT_REL__", $stdoutRelLit).Replace("__STDERR_REL__", $stderrRelLit).Replace("__PORT__", $safePort).Replace("__HEALTH_TIMEOUT__", "$safeTimeout")
  $cmd = $cmd -replace "`r", "" -replace "`n", " "
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
      $output | ForEach-Object { Write-Host $_ }
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

    if ($isTimeout) {
      throw ("Timeout de red durante " + $Label + ". Verifica internet, firewall, VPN y acceso al VPS remoto.")
    }

    if (Is-AuthDeniedMessage -Text $text) {
      throw "Autenticación SSH rechazada durante $Label. Verifica la clave configurada en IdentityFile y que su pública esté instalada en el VPS."
    }

    throw ("Falló " + $Label + " (código " + $exitCode + ").")
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
  param([AllowEmptyString()][string]$ExcludeFile)

  $patterns = @(
    ".git",
    ".gitignore",
    "node_modules",
    "logs",
    "test_runs",
    "*.db",
    "*.sqlite",
    "*.exe",
    "backend/.env.local",
    "backend/server.err",
    "*.ppk",
    "*.pem",
    "*.key"
  )

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

  if (Get-Command Test-NetConnection -ErrorAction SilentlyContinue) {
    $tnc = Test-NetConnection $RemoteHost -Port $Port -WarningAction SilentlyContinue
    if (-not $tnc.TcpTestSucceeded) {
      throw "No hay conectividad TCP con ${RemoteHost}:$Port desde este equipo."
    }
  }

  $remoteTarget = "$RemoteUser@$RemoteHost"
  $identityResolved = (Resolve-Path $IdentityPath).Path
  $isPpkIdentity = ([System.IO.Path]::GetExtension($identityResolved).ToLowerInvariant() -eq ".ppk")
  $transportLabel = "OpenSSH"
  $excludePatterns = Get-SyncExcludePatterns -ExcludeFile $ExcludeFile

  $tmpDir = Join-Path $env:TEMP "pcs_sync_staging"
  if (-not (Test-Path $tmpDir)) {
    New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null
  }

  $stamp = Get-Date -Format "yyyyMMdd_HHmmss"
  $archivePath = Join-Path $tmpDir ("pcs_sync_" + $stamp + ".tar")
  $remoteArchive = "/tmp/pcs_sync_$stamp.tar"

  $tarArgs = @()
  foreach ($pattern in $excludePatterns) {
    $tarArgs += "--exclude=$pattern"
  }
  $tarArgs += @("-cf", $archivePath, "-C", $LocalResolvedPath, ".")

  $mkdirCmd = "mkdir -p '$RemotePath'"
  $extractCmd = "mkdir -p '$RemotePath' && tar -xf '$remoteArchive' -C '$RemotePath' && rm -f '$remoteArchive'"
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

  if ($isPpkIdentity) {
    $tools = Ensure-PuttyTools -AutoInstall $AutoInstallDeps
    $plink = $tools.Plink
    $pscp = $tools.Pscp
    $transportLabel = "PuTTY"

    $verifyCommandPath = $plink
    $uploadCommandPath = $pscp
    $extractCommandPath = $plink
    $bootstrapCommandPath = $plink

    $verifyArgs = @('-batch', '-P', "$Port", '-i', $identityResolved, $remoteTarget, $mkdirCmd)
    $uploadArgs = @('-batch', '-P', "$Port", '-i', $identityResolved, $archivePath, "${remoteTarget}:$remoteArchive")
    $extractArgs = @('-batch', '-P', "$Port", '-i', $identityResolved, $remoteTarget, $extractCmd)
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
      $bootstrapArgs = @('-batch', '-P', "$Port", '-i', $identityResolved, $remoteTarget, $bootstrapCmd)
    } else {
      $bootstrapArgs = @(
        '-o', 'BatchMode=yes',
        '-o', 'StrictHostKeyChecking=accept-new',
        '-o', 'ConnectTimeout=15',
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
      $restartArgs = @('-batch', '-P', "$Port", '-i', $identityResolved, $remoteTarget, $restartCmd)
    } else {
      $restartArgs = @(
        '-o', 'BatchMode=yes',
        '-o', 'StrictHostKeyChecking=accept-new',
        '-o', 'ConnectTimeout=15',
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

    if ($IsDryRun) {
      Write-Host "[INFO] Modo DryRun (sin cambios remotos)."
      $entries = & tar -tf $archivePath
      $count = ($entries | Measure-Object).Count
      Write-Host ("[INFO] Archivos que se transferirían: " + $count)
      $entries | Select-Object -First 40 | ForEach-Object { Write-Host ("  - " + $_) }
      if ($count -gt 40) {
        Write-Host "  ..."
      }
      return
    }

    Invoke-ExternalWithRetry -Label "verificación remota" -CommandPath $verifyCommandPath -Arguments $verifyArgs -MaxAttempts $Retries -RetryOnTimeoutOnly
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
    $DbDialect = [Environment]::GetEnvironmentVariable("DB_DIALECT")
    if ([string]::IsNullOrWhiteSpace($DbDialect)) {
      $DbDialect = [Environment]::GetEnvironmentVariable("DB_ENGINE")
    }
    if ([string]::IsNullOrWhiteSpace($DbDialect)) {
      $DbDialect = [Environment]::GetEnvironmentVariable("PCS_DB_DIALECT")
    }
    if ([string]::IsNullOrWhiteSpace($DbDialect)) {
      $DbDialect = Get-DotEnvValue -EnvFilePath $localBackendEnvPath -Key "DB_DIALECT"
    }
  }

  if ([string]::IsNullOrWhiteSpace($DbEmpresasDsn)) {
    $DbEmpresasDsn = [Environment]::GetEnvironmentVariable("DB_EMPRESAS_DSN")
    if ([string]::IsNullOrWhiteSpace($DbEmpresasDsn)) {
      $DbEmpresasDsn = [Environment]::GetEnvironmentVariable("PCS_DB_EMPRESAS_DSN")
    }
    if ([string]::IsNullOrWhiteSpace($DbEmpresasDsn)) {
      $DbEmpresasDsn = Get-DotEnvValue -EnvFilePath $localBackendEnvPath -Key "DB_EMPRESAS_DSN"
    }
  }

  if ([string]::IsNullOrWhiteSpace($DbSuperadminDsn)) {
    $DbSuperadminDsn = [Environment]::GetEnvironmentVariable("DB_SUPERADMIN_DSN")
    if ([string]::IsNullOrWhiteSpace($DbSuperadminDsn)) {
      $DbSuperadminDsn = [Environment]::GetEnvironmentVariable("PCS_DB_SUPERADMIN_DSN")
    }
    if ([string]::IsNullOrWhiteSpace($DbSuperadminDsn)) {
      $DbSuperadminDsn = Get-DotEnvValue -EnvFilePath $localBackendEnvPath -Key "DB_SUPERADMIN_DSN"
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

  if (-not $SkipBuild) {
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
    Write-Host "[INFO] Compilación Linux omitida por parámetro -SkipBuild."
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
    Invoke-PuttySync -LocalResolvedPath $LocalPath -RemoteUser $RemoteUser -RemoteHost $RemoteHost -RemotePath $RemotePath -Port $Port -IdentityPath $IdentityFile -IsDryRun $DryRun.IsPresent -IsPreviewOnly $PreviewOnly.IsPresent -ExcludeFile $ExcludeFile -ExecRelativePath $builtBinaryRel -Retries $RetryCount -AutoInstallDeps $AutoInstallDependencies -RunBootstrap $BootstrapServer -BootstrapServerPort $ServerPort -BootstrapGoogleClientId $GoogleClientId -BootstrapGoogleClientSecret $GoogleClientSecret -BootstrapGoogleRedirectUrl $GoogleRedirectUrl -BootstrapDbDialect $DbDialect -BootstrapDbEmpresasDsn $DbEmpresasDsn -BootstrapDbSuperadminDsn $DbSuperadminDsn -RestartServer $RestartRemoteServer -RestartBinaryRelativePath $restartBinaryRel -RestartStdoutLogRelativePath $RemoteStdoutLogPath -RestartStderrLogRelativePath $RemoteStderrLogPath -RestartHealthTimeout $RestartHealthTimeoutSeconds
    if ($OpenPublicUrlAfterDeploy -and -not $DryRun.IsPresent -and -not $PreviewOnly.IsPresent -and $RestartRemoteServer) {
      $deployUrl = "http://$RemoteHost`:$ServerPort/"
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

  if ($RestartRemoteServer) {
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
    $wslOutput | ForEach-Object { Write-Host $_ }
  }

  if ($LASTEXITCODE -ne 0) {
    throw "La sincronización terminó con código $LASTEXITCODE"
  }

  if ($OpenPublicUrlAfterDeploy -and -not $DryRun.IsPresent -and -not $PreviewOnly.IsPresent -and $RestartRemoteServer) {
    $deployUrl = "http://$RemoteHost`:$ServerPort/"
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
  Write-Error $_.Exception.Message
}

$global:LASTEXITCODE = $script:SyncExitCode
return
