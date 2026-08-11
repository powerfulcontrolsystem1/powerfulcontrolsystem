<#
.SYNOPSIS
  Ejecuta el drill P109 de aplicación restaurada contra digests activos de staging.

.DESCRIPTION
  Copia el script de drill al VPS por entrada SSH, lo ejecuta exclusivamente con
  red, PostgreSQL, puertos y directorio temporales, y lo elimina al terminar.
  Nunca conecta el API restaurado a la base activa de staging o producción.
#>

param(
  [string]$RemoteUser = "",
  [string]$RemoteHost = "",
  [int]$Port = 0,
  [string]$IdentityFile = "",
  [string]$RemotePath = "",
  [string]$SourceEnv = "",
  [switch]$PullMissingImages,
  [switch]$VerifyReplica,
  [switch]$VerifyCoordinatedRollback,
  [switch]$AllowRemoteTarget
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$deploymentConfig = if (-not [string]::IsNullOrWhiteSpace($env:PCS_DEPLOYMENT_CONFIG)) {
  $env:PCS_DEPLOYMENT_CONFIG
} else {
  Join-Path $PSScriptRoot "pcs_deployment.local.ps1"
}
if (Test-Path -LiteralPath $deploymentConfig) { . $deploymentConfig }
if ([string]::IsNullOrWhiteSpace($RemoteUser) -and (Get-Variable -Name PcsVpsUser -Scope Script -ErrorAction SilentlyContinue)) { $RemoteUser = $script:PcsVpsUser }
if ([string]::IsNullOrWhiteSpace($RemoteHost) -and (Get-Variable -Name PcsVpsHost -Scope Script -ErrorAction SilentlyContinue)) { $RemoteHost = $script:PcsVpsHost }
if ($Port -le 0 -and (Get-Variable -Name PcsVpsPort -Scope Script -ErrorAction SilentlyContinue)) { $Port = [int]$script:PcsVpsPort }
if ([string]::IsNullOrWhiteSpace($RemotePath) -and (Get-Variable -Name PcsVpsRemotePath -Scope Script -ErrorAction SilentlyContinue)) { $RemotePath = $script:PcsVpsRemotePath }
if ([string]::IsNullOrWhiteSpace($IdentityFile) -and (Get-Variable -Name PcsVpsIdentityFile -Scope Script -ErrorAction SilentlyContinue)) { $IdentityFile = $script:PcsVpsIdentityFile }

if (-not $AllowRemoteTarget) { throw "Operacion remota bloqueada. Usa -AllowRemoteTarget solo en staging autorizado." }
if ([string]::IsNullOrWhiteSpace($RemoteUser) -or [string]::IsNullOrWhiteSpace($RemoteHost) -or $Port -le 0 -or [string]::IsNullOrWhiteSpace($RemotePath)) { throw "Faltan datos privados de destino." }
if ([string]::IsNullOrWhiteSpace($SourceEnv)) { $SourceEnv = "$RemotePath/deploy/.env.staging" }
if ([string]::IsNullOrWhiteSpace($IdentityFile)) {
  $configKey = Join-Path (Split-Path -Parent $deploymentConfig) "clave privada ssh.ppk"
  if (Test-Path -LiteralPath $configKey) { $IdentityFile = $configKey }
}
if ([string]::IsNullOrWhiteSpace($IdentityFile)) {
  foreach ($candidate in @((Join-Path $HOME ".ssh\id_rsa"), (Join-Path $HOME ".ssh\id_ed25519"))) {
    if (Test-Path -LiteralPath $candidate) { $IdentityFile = $candidate; break }
  }
}
if ([string]::IsNullOrWhiteSpace($IdentityFile) -or -not (Test-Path -LiteralPath $IdentityFile)) { throw "No se encontro IdentityFile configurado." }

function Resolve-Plink {
  $cmd = Get-Command plink.exe -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source -ErrorAction SilentlyContinue
  if ($cmd) { return $cmd }
  foreach ($candidate in @("D:\Program Files\PuTTY\plink.exe", "C:\Program Files\PuTTY\plink.exe", "C:\Program Files (x86)\PuTTY\plink.exe")) { if (Test-Path -LiteralPath $candidate) { return $candidate } }
  throw "No se encontro plink.exe."
}

function Resolve-OpenSSH {
  $cmd = Get-Command ssh.exe -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source -ErrorAction SilentlyContinue
  if ($cmd) { return $cmd }
  foreach ($candidate in @("C:\Windows\System32\OpenSSH\ssh.exe", "C:\Program Files\Git\usr\bin\ssh.exe")) { if (Test-Path -LiteralPath $candidate) { return $candidate } }
  throw "No se encontro ssh.exe."
}

function Convert-ToBashLiteral([string]$Value) { return "'" + $Value.Replace("'", "'\''") + "'" }

$drillSource = Join-Path $repoRoot "deploy\scripts\vps-p109-restored-app-drill.sh"
if (-not (Test-Path -LiteralPath $drillSource)) { throw "No existe el drill P109 de aplicación restaurada." }
$sourceHash = (Get-FileHash -LiteralPath $drillSource -Algorithm SHA256).Hash.ToLowerInvariant()
$sourceBase64 = [Convert]::ToBase64String([IO.File]::ReadAllBytes($drillSource))
$drillID = "p109-restore-app-" + (Get-Date -Format "yyyyMMddhhmmss")
$remotePathLit = Convert-ToBashLiteral $RemotePath
$sourceEnvLit = Convert-ToBashLiteral $SourceEnv
$drillIDLit = Convert-ToBashLiteral $drillID
$pullMissingValue = if ($PullMissingImages) { "1" } else { "0" }
$verifyReplicaValue = if ($VerifyReplica) { "1" } else { "0" }
$verifyRollbackValue = if ($VerifyCoordinatedRollback) { "1" } else { "0" }
$qaEmail = [string]$env:P109_QA_EMAIL
$qaPassword = [string]$env:P109_QA_PASSWORD
if ($VerifyCoordinatedRollback -and -not $VerifyReplica) { throw "El rollback coordinado requiere -VerifyReplica." }
if ($VerifyReplica -and ([string]::IsNullOrWhiteSpace($qaEmail) -or [string]::IsNullOrWhiteSpace($qaPassword))) { throw "La replica autenticada requiere P109_QA_EMAIL y P109_QA_PASSWORD solo en la sesion actual." }
$qaEmailLit = Convert-ToBashLiteral $qaEmail
$qaPasswordLit = Convert-ToBashLiteral $qaPassword

$remoteScript = @"
set -euo pipefail
remote_path=$remotePathLit
drill_id=$drillIDLit
tmp_script="/tmp/`$drill_id.sh"
cleanup() { rm -f "`$tmp_script"; }
trap cleanup EXIT
base64 -d > "`$tmp_script" <<'PCS_DRILL_B64'
$sourceBase64
PCS_DRILL_B64
actual_hash="`$(sha256sum "`$tmp_script" | awk '{print `$1}')"
[ "`$actual_hash" = "$sourceHash" ] || { echo "[ERROR] El script remoto no coincide con el SHA-256 local." >&2; exit 1; }
chmod 700 "`$tmp_script"
source_env=$sourceEnvLit
[ -f "`$source_env" ] || { echo "[ERROR] Falta configuracion privada de staging." >&2; exit 1; }
resolve_digest() {
  local container="`$1"
  local configured image_id digest env_name
  configured="`$(docker inspect -f '{{.Config.Image}}' "`$container" 2>/dev/null || true)"
  case "`$configured" in *@sha256:????????????????????????????????????????????????????????????????) printf '%s' "`$configured"; return 0;; esac
  image_id="`$(docker inspect -f '{{.Image}}' "`$container" 2>/dev/null || true)"
  digest="`$(docker image inspect -f '{{range .RepoDigests}}{{println .}}{{end}}' "`$image_id" 2>/dev/null | awk '/@sha256:[a-f0-9]{64}$/ {print; exit}')"
  if [ -n "`$digest" ]; then printf '%s' "`$digest"; return 0; fi
  case "`$container" in
    pcs-staging-api) env_name=PCS_API_IMAGE_DIGEST ;;
    pcs-staging-backend) env_name=PCS_API_IMAGE_DIGEST ;;
    pcs-staging-migrate) env_name=PCS_MIGRATE_IMAGE_DIGEST ;;
    *) return 0 ;;
  esac
  digest="`$(awk -F= -v key="`$env_name" '`$1 == key {sub(/^[^=]*=/, ""); print; exit}' "`$source_env")"
  printf '%s' "`$digest"
}
api_image="`$(resolve_digest pcs-staging-api)"
# La topología vigente de Compose nombra el servicio API como backend. Se
# conserva el nombre histórico como primer intento para no romper candidatos
# anteriores, pero el drill sigue exigiendo una referencia inmutable por digest.
if ! [[ "`$api_image" =~ @sha256:????????????????????????????????????????????????????????????????$ ]]; then
  api_image="`$(resolve_digest pcs-staging-backend)"
fi
migrate_image="`$(resolve_digest pcs-staging-migrate)"
case "`$api_image" in *@sha256:????????????????????????????????????????????????????????????????) ;; *) echo "[ERROR] API staging sin digest inmutable." >&2; exit 1;; esac
case "`$migrate_image" in *@sha256:????????????????????????????????????????????????????????????????) ;; *) echo "[ERROR] Migrador staging sin digest inmutable." >&2; exit 1;; esac
if ! docker image inspect "`$api_image" >/dev/null 2>&1; then
  [ "$pullMissingValue" = "1" ] || { echo "[ERROR] La imagen API exacta no esta disponible localmente." >&2; exit 1; }
  docker pull "`$api_image" >/dev/null
fi
if ! docker image inspect "`$migrate_image" >/dev/null 2>&1; then
  [ "$pullMissingValue" = "1" ] || { echo "[ERROR] La imagen migrador exacta no esta disponible localmente." >&2; exit 1; }
  docker pull "`$migrate_image" >/dev/null
fi
PROJECT_DIR="`$remote_path" SOURCE_ENV="`$source_env" P109_DRILL_ID="`$drill_id" \
  PCS_API_IMAGE_DIGEST="`$api_image" PCS_MIGRATE_IMAGE_DIGEST="`$migrate_image" \
  P109_VERIFY_PRIVATE_INVENTORY=1 P109_VERIFY_REPLICA=$verifyReplicaValue \
  P109_VERIFY_COORDINATED_ROLLBACK=$verifyRollbackValue \
  P109_QA_EMAIL=$qaEmailLit P109_QA_PASSWORD=$qaPasswordLit bash "`$tmp_script"
"@

$tmp = Join-Path $env:TEMP ("pcs_p109_restore_app_" + [guid]::NewGuid().ToString("N") + ".sh")
[IO.File]::WriteAllText($tmp, ($remoteScript -replace "`r", ""), [Text.UTF8Encoding]::new($false))
try {
  if ([IO.Path]::GetExtension($IdentityFile).ToLowerInvariant() -eq ".ppk") {
    $plink = Resolve-Plink
    & $plink -batch -P $Port -i $IdentityFile -m $tmp "$RemoteUser@$RemoteHost"
  } else {
    $ssh = Resolve-OpenSSH
    $psi = [Diagnostics.ProcessStartInfo]::new()
    $psi.FileName = $ssh; $psi.UseShellExecute = $false; $psi.RedirectStandardInput = $true; $psi.RedirectStandardOutput = $true; $psi.RedirectStandardError = $true
    foreach ($arg in @("-T", "-o", "BatchMode=yes", "-o", "IdentitiesOnly=yes", "-o", "StrictHostKeyChecking=accept-new", "-p", "$Port", "-i", "$IdentityFile", "$RemoteUser@$RemoteHost", "bash -s")) { [void]$psi.ArgumentList.Add($arg) }
    $process = [Diagnostics.Process]::new(); $process.StartInfo = $psi; [void]$process.Start()
    $process.StandardInput.Write([IO.File]::ReadAllText($tmp)); $process.StandardInput.Close()
    $stdout = $process.StandardOutput.ReadToEnd(); $stderr = $process.StandardError.ReadToEnd(); $process.WaitForExit()
    if ($stdout) { Write-Output $stdout.TrimEnd() }
    if ($process.ExitCode -ne 0 -and $stderr) { Write-Output "[ERROR] El drill remoto emitio diagnostico en stderr." }
    $global:LASTEXITCODE = $process.ExitCode
  }
  if ($LASTEXITCODE -ne 0) { throw "El drill P109 de aplicación restaurada falló." }
} finally { Remove-Item -LiteralPath $tmp -Force -ErrorAction SilentlyContinue }
