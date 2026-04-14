<#
.SYNOPSIS
  Instala una clave publica OpenSSH en el VPS remoto dentro de ~/.ssh/authorized_keys.

.DESCRIPTION
  Script manual (sin cron) para preparar acceso SSH por clave publica.
  Incluye reintentos para errores de red intermitentes (timeout) y modo PreviewOnly.
#>

param(
  [string]$RemoteHost = "2.24.197.58",
  [string]$User = "root",
  [int]$Port = 22,
  [string]$PublicKeyFile = "",
  [string]$PublicKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCEWb73M/GiB38Z63C9pzqkSgRlIjDr8uZcMy/N/lu8CucNQik+bCXJ3NIuEbmNB0HG9TYktfjtER1zzAwT1FYYvZAgkbLC+HLuMbK2cFuqUsYrZs4Rlht9ft2pdYZ1MWbQeTEcio5/ZBkeqRWTp6LbVeZ1C+L/x89H/5Adfip5C/oalF5ufZRIxd7xZe2JK4eCMQZ+KqBGlZtp5Nd7n0Xh9CFl7WYgHleIAM+rtYk8iYt5dMqFc6GaUC/en4H2Ki42E/Ns3KrEOMd5kC8ZRm1c65ewBiBSZgYnvFVMKXdWYSvh1EyfbXk4rA1zkg1A9zk/k45tKcMEnsV/uePtTgzn rsa-key-20260413",
  [int]$RetryCount = 3,
  [switch]$PreviewOnly
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Resolve-SshExecutable {
  $cmd = Get-Command ssh.exe -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source -ErrorAction SilentlyContinue
  if ($cmd) {
    return $cmd
  }

  $fallback = "C:\Windows\System32\OpenSSH\ssh.exe"
  if (Test-Path $fallback) {
    return $fallback
  }

  throw "No se encontró ssh.exe. Instálalo con 'Add-WindowsCapability -Online -Name OpenSSH.Client~~~~0.0.1.0' o habilita OpenSSH Client en Windows Features."
}

function Get-EffectivePublicKey {
  param(
    [string]$FilePath,
    [string]$InlineKey
  )

  if ($FilePath) {
    if (-not (Test-Path $FilePath)) {
      throw "No se encontró PublicKeyFile: $FilePath"
    }
    $raw = Get-Content -LiteralPath $FilePath -Raw -ErrorAction Stop
    $value = $raw.Trim()
  } else {
    if ($null -eq $InlineKey) {
      $value = ""
    } else {
      $value = $InlineKey.Trim()
    }
  }

  if ([string]::IsNullOrWhiteSpace($value)) {
    throw "No hay clave pública para instalar. Usa -PublicKeyFile o -PublicKey."
  }

  if ($value -notmatch "^(ssh-(rsa|ed25519)|ecdsa-sha2-nistp)\S*\s+") {
    throw "La clave pública no parece formato OpenSSH válido (ej: ssh-rsa ... o ssh-ed25519 ...)."
  }

  # Forzamos una sola línea por seguridad.
  return ($value -replace "\r", "" -replace "\n", "")
}

function ConvertTo-SingleQuotedShellText {
  param([string]$Text)
  return $Text.Replace("'", "'\\''")
}

function Test-IsTimeoutText {
  param([string]$Text)
  if ([string]::IsNullOrWhiteSpace($Text)) {
    return $false
  }
  return ($Text -match "(?i)timed out|tiempo de espera|network error|No route to host|Connection reset")
}

function Invoke-SshWithRetry {
  param(
    [string]$SshExe,
    [string[]]$CmdArgs,
    [int]$Attempts
  )

  if ($Attempts -lt 1) {
    $Attempts = 1
  }

  for ($i = 1; $i -le $Attempts; $i++) {
    Write-Host ("[INFO] Conectando por SSH (intento " + $i + "/" + $Attempts + ")...")
    $out = & $SshExe @CmdArgs 2>&1
    $code = $LASTEXITCODE

    if ($out) {
      $out | ForEach-Object { Write-Host $_ }
    }

    if ($code -eq 0) {
      return
    }

    $text = ($out -join "`n")
    $isTimeout = Test-IsTimeoutText -Text $text
    if ($i -lt $Attempts -and $isTimeout) {
      Write-Warning "Timeout de red detectado. Reintentando..."
      continue
    }

    if ($isTimeout) {
      throw "La instalación de clave falló por timeout de red. Verifica internet, firewall o VPN."
    }

    throw "La instalación de clave terminó con código $code"
  }
}

$sshExe = Resolve-SshExecutable
$key = Get-EffectivePublicKey -FilePath $PublicKeyFile -InlineKey $PublicKey
$keyEscaped = ConvertTo-SingleQuotedShellText -Text $key

$remoteCommand = "umask 077; mkdir -p ~/.ssh; touch ~/.ssh/authorized_keys; if ! grep -Fqx -- '$keyEscaped' ~/.ssh/authorized_keys; then printf '%s\n' '$keyEscaped' >> ~/.ssh/authorized_keys; fi; chmod 700 ~/.ssh; chmod 600 ~/.ssh/authorized_keys; echo '[OK] clave pública instalada (o ya existente)'"
$sshArgs = @("-p", "$Port", "$User@$RemoteHost", $remoteCommand)

Write-Host ("Destino SSH: " + $User + "@" + $RemoteHost + ":" + $Port)

if ($PreviewOnly) {
  Write-Host "[PREVIEW] Comando que se ejecutaría:"
  Write-Host ($sshExe + " " + (($sshArgs | ForEach-Object { '"' + $_ + '"' }) -join " "))
  return
}

Invoke-SshWithRetry -SshExe $sshExe -CmdArgs $sshArgs -Attempts $RetryCount
