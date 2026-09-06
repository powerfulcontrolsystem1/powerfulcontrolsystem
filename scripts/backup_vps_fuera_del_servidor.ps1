<#
.SYNOPSIS
  Crea un backup PCS en el VPS, lo descarga a D: y verifica todos sus SHA-256.

.DESCRIPTION
  La copia remota solo se elimina después de comprobar MANIFEST.txt y cada
  entrada de SHA256SUMS en el almacenamiento local. No imprime ni copia la
  configuración privada de despliegue.
#>

[CmdletBinding()]
param(
  [ValidateSet("database", "system", "vps")]
  [string]$Scope = "vps",
  [string]$BackupRoot = "D:\Backup vps PCS",
  [ValidateRange(1, 3650)]
  [int]$LocalRetentionDays = 30,
  [ValidateRange(1, 30)]
  [int]$RemoteRetentionDays = 2,
  [switch]$KeepRemoteCopy
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$privateConfig = Join-Path $PSScriptRoot "pcs_deployment.local.ps1"
if (-not (Test-Path -LiteralPath $privateConfig)) {
  throw "Falta la configuración privada de despliegue."
}
. $privateConfig

$ssh = "C:\Windows\System32\OpenSSH\ssh.exe"
$scp = "C:\Windows\System32\OpenSSH\scp.exe"
if (-not (Test-Path -LiteralPath $ssh) -or -not (Test-Path -LiteralPath $scp)) {
  throw "OpenSSH de Windows no está disponible."
}

$target = "$script:PcsVpsUser@$script:PcsVpsHost"
$sshArgs = @("-p", [string]$script:PcsVpsPort, "-o", "BatchMode=yes", "-o", "ConnectTimeout=20", "-o", "StrictHostKeyChecking=accept-new", "-o", "ServerAliveInterval=15")
$scpArgs = @("-P", [string]$script:PcsVpsPort, "-o", "BatchMode=yes", "-o", "ConnectTimeout=20", "-o", "StrictHostKeyChecking=accept-new", "-o", "ServerAliveInterval=15")
$identityVariable = Get-Variable PcsVpsIdentityFile -Scope Script -ErrorAction SilentlyContinue
if ($identityVariable -and -not [string]::IsNullOrWhiteSpace([string]$identityVariable.Value)) {
  $sshArgs += @("-i", [string]$identityVariable.Value)
  $scpArgs += @("-i", [string]$identityVariable.Value)
}

$stamp = Get-Date -Format "yyyyMMdd_HHmmss"
$remoteRoot = "$($script:PcsVpsRemotePath.TrimEnd('/'))/backups/vps-snapshots"
$remoteDir = "$remoteRoot/$stamp"
$localRootFull = [IO.Path]::GetFullPath($BackupRoot)
[IO.Directory]::CreateDirectory($localRootFull) | Out-Null
$localDir = Join-Path $localRootFull ("PCS_{0}_backup_structured_{1}" -f $Scope.ToUpperInvariant(), $stamp)

Write-Host "[INFO] Creando backup $Scope en el VPS."
$remoteCreate = "cd '$($script:PcsVpsRemotePath)' && BACKUP_SCOPE='$Scope' BACKUP_STAMP='$stamp' RETENTION_DAYS='$RemoteRetentionDays' bash deploy/scripts/vps-backup-operacion.sh"
& $ssh @sshArgs $target $remoteCreate
if ($LASTEXITCODE -ne 0) { throw "No se pudo crear el backup remoto." }

[IO.Directory]::CreateDirectory($localDir) | Out-Null
try {
  Write-Host "[INFO] Descargando backup fuera del VPS: $localDir"
  & $scp @scpArgs -r "${target}:$remoteDir/." $localDir
  if ($LASTEXITCODE -ne 0) { throw "No se pudo descargar el backup." }

  $manifestPath = Join-Path $localDir "MANIFEST.txt"
  $checksumsPath = Join-Path $localDir "SHA256SUMS"
  if (-not (Test-Path -LiteralPath $manifestPath) -or -not (Test-Path -LiteralPath $checksumsPath)) {
    throw "La copia descargada no contiene manifiesto y checksums."
  }
  if (-not (Select-String -LiteralPath $manifestPath -SimpleMatch "scope=$Scope" -Quiet)) {
    throw "El alcance del manifiesto no coincide con la solicitud."
  }

  $verified = 0
  foreach ($line in Get-Content -LiteralPath $checksumsPath) {
    if ($line -notmatch '^([0-9a-f]{64})\s+(.+)$') {
      throw "SHA256SUMS contiene una línea inválida."
    }
    $expected = $Matches[1]
    $name = $Matches[2].TrimStart("*")
    $candidate = [IO.Path]::GetFullPath((Join-Path $localDir $name))
    $localPrefix = [IO.Path]::GetFullPath($localDir) + [IO.Path]::DirectorySeparatorChar
    if (-not $candidate.StartsWith($localPrefix, [StringComparison]::OrdinalIgnoreCase)) {
      throw "SHA256SUMS contiene una ruta fuera del backup."
    }
    if (-not (Test-Path -LiteralPath $candidate)) {
      throw "Falta un archivo declarado en SHA256SUMS."
    }
    $actual = (Get-FileHash -LiteralPath $candidate -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($actual -ne $expected.ToLowerInvariant()) {
      throw "Falló la verificación SHA-256 de un archivo."
    }
    $verified++
  }
  if ($verified -lt 2) { throw "La copia no contiene suficientes artefactos verificables." }

  if (-not $KeepRemoteCopy) {
    $remoteDelete = "case '$remoteDir' in '$remoteRoot'/[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]_[0-9][0-9][0-9][0-9][0-9][0-9]) rm -rf --one-file-system '$remoteDir' ;; *) exit 9 ;; esac"
    & $ssh @sshArgs $target $remoteDelete
    if ($LASTEXITCODE -ne 0) { throw "La copia local quedó verificada, pero no se pudo retirar su temporal remoto." }
  }

  $cutoff = (Get-Date).AddDays(-$LocalRetentionDays)
  $rootPrefix = $localRootFull.TrimEnd([IO.Path]::DirectorySeparatorChar) + [IO.Path]::DirectorySeparatorChar
  Get-ChildItem -LiteralPath $localRootFull -Directory -Filter "PCS_*_backup_structured_*" | Where-Object {
    $_.LastWriteTime -lt $cutoff -and $_.FullName -ne $localDir
  } | ForEach-Object {
    $oldPath = [IO.Path]::GetFullPath($_.FullName)
    if ($oldPath.StartsWith($rootPrefix, [StringComparison]::OrdinalIgnoreCase)) {
      [IO.Directory]::Delete($oldPath, $true)
    }
  }

  $bytes = (Get-ChildItem -LiteralPath $localDir -File | Measure-Object -Property Length -Sum).Sum
  Write-Host "[OK] Backup externo verificado: scope=$Scope files=$verified bytes=$bytes path=$localDir"
} catch {
  Write-Error $_
  throw
}
