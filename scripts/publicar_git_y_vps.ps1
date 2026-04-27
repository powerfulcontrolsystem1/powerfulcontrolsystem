<#
.SYNOPSIS
  Encadena actualizar_repositorio.ps1 (push a origin) y sync_to_vps.ps1 (despliegue al VPS).

.DESCRIPTION
  Carga scripts/pcs_deployment.local.ps1 si existe (misma config que al ejecutar los scripts por separado).
  -SkipGit: solo VPS. -SkipVps: solo Git.
  Reexpone parametros comunes del flujo Git y los controles principales del despliegue VPS.
#>
param(
  [string]$Message = "Publicacion: git y VPS",
  [string]$RepoUrl = "",
  [switch]$SkipChangeLog,
  [switch]$SetOrigin,
  [switch]$ForcePush,
  [switch]$SkipGit,
  [switch]$SkipVps,
  [switch]$DryRun,
  [switch]$PreviewOnly,
  [switch]$SkipBuild,
  [switch]$BuildOnly,
  [string]$LocalPath = "",
  [string]$RemoteUser = "",
  [string]$RemoteHost = "",
  [string]$RemotePath = "",
  [int]$Port = 0,
  [string]$IdentityFile = "",
  [string]$ExcludeFile = "",
  [string]$ServerPort = "",
  [string]$GoogleClientId = "",
  [string]$GoogleClientSecret = "",
  [string]$GoogleRedirectUrl = "",
  [string]$PublicBaseUrl = "",
  [string]$DbDialect = "",
  [string]$DbEmpresasDsn = "",
  [string]$DbSuperadminDsn = "",
  [string]$BootstrapServer = "",
  [string]$RestartRemoteServer = "",
  [string]$OpenPublicUrlAfterDeploy = ""
)

$ErrorActionPreference = "Stop"
$here = $PSScriptRoot

$pub = Join-Path $here "actualizar_repositorio.ps1"
$sync = Join-Path $here "sync_to_vps.ps1"

if (-not (Test-Path -LiteralPath $pub)) { throw "No se encuentra: $pub" }
if (-not (Test-Path -LiteralPath $sync)) { throw "No se encuentra: $sync" }

function ConvertTo-OptionalBoolean {
  param(
    [string]$Value,
    [string]$ParameterName
  )

  if ([string]::IsNullOrWhiteSpace($Value)) {
    return $null
  }

  switch ($Value.Trim().ToLowerInvariant()) {
    '1' { return $true }
    '0' { return $false }
    'true' { return $true }
    'false' { return $false }
    'yes' { return $true }
    'no' { return $false }
    'si' { return $true }
    default { throw "Valor no valido para -${ParameterName}: $Value. Usa true/false, 1/0, yes/no o si." }
  }
}

if (-not $SkipGit) {
  $gitArgs = @{
    Message       = $Message
    SkipChangeLog = $SkipChangeLog
    SetOrigin     = $SetOrigin
    ForcePush     = $ForcePush
  }
  if (-not [string]::IsNullOrWhiteSpace($RepoUrl)) {
    $gitArgs.RepoUrl = $RepoUrl
  }

  & $pub @gitArgs
  if ($null -ne $LASTEXITCODE -and $LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}
if (-not $SkipVps) {
  $syncArgs = @{}

  foreach ($name in @(
    'DryRun',
    'PreviewOnly',
    'SkipBuild',
    'BuildOnly'
  )) {
    if ($PSBoundParameters.ContainsKey($name)) {
      $syncArgs[$name] = $PSBoundParameters[$name]
    }
  }

  foreach ($name in @(
    'LocalPath',
    'RemoteUser',
    'RemoteHost',
    'RemotePath',
    'IdentityFile',
    'ExcludeFile',
    'ServerPort',
    'GoogleClientId',
    'GoogleClientSecret',
    'GoogleRedirectUrl',
    'PublicBaseUrl',
    'DbDialect',
    'DbEmpresasDsn',
    'DbSuperadminDsn'
  )) {
    if ($PSBoundParameters.ContainsKey($name) -and -not [string]::IsNullOrWhiteSpace([string]$PSBoundParameters[$name])) {
      $syncArgs[$name] = $PSBoundParameters[$name]
    }
  }

  if ($PSBoundParameters.ContainsKey('Port') -and $Port -gt 0) {
    $syncArgs.Port = $Port
  }

  foreach ($name in @(
    'BootstrapServer',
    'RestartRemoteServer',
    'OpenPublicUrlAfterDeploy'
  )) {
    if ($PSBoundParameters.ContainsKey($name)) {
      $parsedBool = ConvertTo-OptionalBoolean -Value ([string]$PSBoundParameters[$name]) -ParameterName $name
      if ($null -ne $parsedBool) {
        $syncArgs[$name] = $parsedBool
      }
    }
  }

  & $sync @syncArgs
  if ($null -ne $LASTEXITCODE -and $LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

Write-Host "[OK] publicar_git_y_vps: completado." -ForegroundColor Green
exit 0
