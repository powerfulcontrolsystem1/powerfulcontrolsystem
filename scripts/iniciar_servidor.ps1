param(
    [switch]$Background
)

function Write-Step {
    param([string]$Text)
    Write-Host "`n==> $Text" -ForegroundColor Cyan
}

function Write-Info {
    param([string]$Text)
    Write-Host "[INFO] $Text" -ForegroundColor Gray
}

function Write-Ok {
    param([string]$Text)
    Write-Host "[OK] $Text" -ForegroundColor Green
}

function Write-WarnMsg {
    param([string]$Text)
    Write-Host "[AVISO] $Text" -ForegroundColor Yellow
}

function Write-ErrMsg {
    param([string]$Text)
    Write-Host "[ERROR] $Text" -ForegroundColor Red
}

Write-Step "1/8 Preparando entorno"


$backend = Join-Path $PSScriptRoot "..\backend"
if (-not (Test-Path $backend)) {
    Write-ErrMsg "No se encontro la carpeta backend en: $backend"
    exit 1
}
Write-Info "Directorio backend detectado: $backend"
if ($Background) {
    Write-Info "Modo background activo: no se abrira navegador automaticamente."
}
Push-Location $backend

function Import-DotEnvValues {
    param([string]$Path)
    $map = @{}
    if (-not (Test-Path $Path)) { return $map }

    $lines = Get-Content -Path $Path -ErrorAction SilentlyContinue
    foreach ($line in $lines) {
        $raw = [string]$line
        if ([string]::IsNullOrWhiteSpace($raw)) { continue }
        $trimmed = $raw.Trim()
        if ($trimmed.StartsWith('#')) { continue }
        $idx = $trimmed.IndexOf('=')
        if ($idx -lt 1) { continue }

        $key = $trimmed.Substring(0, $idx).Trim()
        $value = $trimmed.Substring($idx + 1).Trim()
        if ($value.StartsWith('"') -and $value.EndsWith('"') -and $value.Length -ge 2) {
            $value = $value.Substring(1, $value.Length - 2)
        }
        if ($value.StartsWith("'") -and $value.EndsWith("'") -and $value.Length -ge 2) {
            $value = $value.Substring(1, $value.Length - 2)
        }
        $map[$key] = $value
    }
    return $map
}

function Test-ConfigEncKeyFormat {
    param([string]$KeyValue)

    if ([string]::IsNullOrWhiteSpace($KeyValue)) {
        return $false
    }

    $candidate = [string]$KeyValue
    try {
        $decoded = [Convert]::FromBase64String($candidate)
        if ($decoded.Length -ge 16) {
            return $true
        }
    } catch {
        # Si no es base64 válido, evaluar modo literal.
    }

    return ($candidate.Length -ge 32)
}

function Save-ConfigEncKeyToEnvLocal {
    param(
        [string]$BackendDir,
        [string]$ConfigEncKey
    )

    $envLocalPath = Join-Path $BackendDir '.env.local'
    $prefix = 'CONFIG_ENC_KEY='

    if (Test-Path $envLocalPath) {
        $lines = @(Get-Content -Path $envLocalPath -ErrorAction SilentlyContinue)
        $updated = $false
        for ($i = 0; $i -lt $lines.Count; $i++) {
            $trimmed = [string]$lines[$i]
            if ($trimmed.Trim().StartsWith($prefix)) {
                $lines[$i] = "$prefix$ConfigEncKey"
                $updated = $true
                break
            }
        }
        if (-not $updated) {
            $lines += "$prefix$ConfigEncKey"
        }
        Set-Content -Path $envLocalPath -Value $lines -Encoding UTF8
        return $envLocalPath
    }

    @(
        '# Archivo local de entorno (secrets de desarrollo; no versionar)'
        "$prefix$ConfigEncKey"
    ) | Set-Content -Path $envLocalPath -Encoding UTF8

    return $envLocalPath
}

function Resolve-ConfigEncryptionKey {
    param([string]$BackendDir)

    $processValue = [Environment]::GetEnvironmentVariable('CONFIG_ENC_KEY', 'Process')
    if (-not [string]::IsNullOrWhiteSpace($processValue)) {
        if (-not (Test-ConfigEncKeyFormat -KeyValue $processValue)) {
            throw 'CONFIG_ENC_KEY en entorno del proceso es inválida. Use base64 válido o >=32 caracteres.'
        }
        return @{
            Value = $processValue
            Source = 'variable de entorno del proceso'
            Generated = $false
        }
    }

    $envCandidates = @(
        (Join-Path -Path $BackendDir -ChildPath '.env.local')
        (Join-Path -Path $BackendDir -ChildPath '.env')
    )

    foreach ($envPath in $envCandidates) {
        if (-not (Test-Path $envPath)) { continue }
        $vals = Import-DotEnvValues -Path $envPath
        if (-not $vals.ContainsKey('CONFIG_ENC_KEY')) { continue }

        $candidate = [string]$vals['CONFIG_ENC_KEY']
        if (-not (Test-ConfigEncKeyFormat -KeyValue $candidate)) {
            throw ("CONFIG_ENC_KEY inválida en {0}. Corrija el valor o elimínelo para autogenerar una nueva." -f $envPath)
        }

        return @{
            Value = $candidate
            Source = $envPath
            Generated = $false
        }
    }

    $bytes = New-Object byte[] 32
    $rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
    try {
        $rng.GetBytes($bytes)
    } finally {
        if ($null -ne $rng) {
            $rng.Dispose()
        }
    }

    $generated = [Convert]::ToBase64String($bytes)
    $savedPath = Save-ConfigEncKeyToEnvLocal -BackendDir $BackendDir -ConfigEncKey $generated

    return @{
        Value = $generated
        Source = $savedPath
        Generated = $true
    }
}

function Resolve-GoogleOAuthCredentials {
    param([string]$BackendDir)

    $result = @{
        ClientId = [Environment]::GetEnvironmentVariable('GOOGLE_CLIENT_ID', 'Process')
        ClientSecret = [Environment]::GetEnvironmentVariable('GOOGLE_CLIENT_SECRET', 'Process')
        Source = 'variables de entorno del proceso'
    }

    $envCandidates = @(
        (Join-Path -Path $BackendDir -ChildPath '.env.local')
        (Join-Path -Path $BackendDir -ChildPath '.env')
    )

    foreach ($envPath in $envCandidates) {
        if (-not (Test-Path $envPath)) { continue }
        $vals = Import-DotEnvValues -Path $envPath
        if ($vals.ContainsKey('GOOGLE_CLIENT_ID') -and -not [string]::IsNullOrWhiteSpace($vals['GOOGLE_CLIENT_ID'])) {
            $result.ClientId = $vals['GOOGLE_CLIENT_ID']
            $result.Source = $envPath
        }
        if ($vals.ContainsKey('GOOGLE_CLIENT_SECRET') -and -not [string]::IsNullOrWhiteSpace($vals['GOOGLE_CLIENT_SECRET'])) {
            $result.ClientSecret = $vals['GOOGLE_CLIENT_SECRET']
            $result.Source = $envPath
        }
    }

    return $result
}

function Test-GoogleOAuthCredentials {
    param(
        [string]$ClientId,
        [string]$ClientSecret,
        [string]$RedirectURL
    )

    if ([string]::IsNullOrWhiteSpace($ClientId) -or [string]::IsNullOrWhiteSpace($ClientSecret)) {
        return $false
    }

    $body = @{
        client_id = $ClientId
        client_secret = $ClientSecret
        code = 'dummy-verification-code'
        grant_type = 'authorization_code'
        redirect_uri = $RedirectURL
    }

    try {
        Invoke-WebRequest -Method Post -Uri 'https://oauth2.googleapis.com/token' -Body $body -ContentType 'application/x-www-form-urlencoded' -UseBasicParsing -ErrorAction Stop | Out-Null
        return $true
    } catch {
        $errorText = ''
        if ($_.ErrorDetails -and $_.ErrorDetails.Message) {
            $errorText = [string]$_.ErrorDetails.Message
        } else {
            $errorText = [string]$_.Exception.Message
        }

        if ($errorText -match 'invalid_client') {
            Write-Host 'Las credenciales OAuth son invalidas (invalid_client).' -ForegroundColor Red
            return $false
        }

        if ($errorText -match 'invalid_grant') {
            # invalid_grant con codigo dummy implica que client_id/client_secret son validos.
            return $true
        }

        Write-Host ("No se pudo validar OAuth de forma concluyente: {0}" -f $errorText) -ForegroundColor Yellow
        return $true
    }
}

function Load-PostgresEnvFromFiles {
    param([string]$BackendDir)

    $envCandidates = @(
        (Join-Path -Path $BackendDir -ChildPath '.env.local')
        (Join-Path -Path $BackendDir -ChildPath '.env')
    )

    foreach ($envPath in $envCandidates) {
        if (-not (Test-Path $envPath)) { continue }
        $vals = Import-DotEnvValues -Path $envPath

        foreach ($key in @(
            'DB_DIALECT',
            'DB_EMPRESAS_DSN',
            'DB_SUPERADMIN_DSN',
            'DB_VPS_TUNNEL_ENABLED',
            'DB_VPS_SSH_HOST',
            'DB_VPS_SSH_USER',
            'DB_VPS_SSH_KEY_PATH',
            'DB_VPS_LOCAL_PORT',
            'DB_VPS_REMOTE_HOST',
            'DB_VPS_REMOTE_PORT',
            'OLLAMA_BASE_URL',
            'OLLAMA_VPS_TUNNEL_ENABLED',
            'OLLAMA_VPS_SSH_HOST',
            'OLLAMA_VPS_SSH_USER',
            'OLLAMA_VPS_SSH_KEY_PATH',
            'OLLAMA_VPS_LOCAL_PORT',
            'OLLAMA_VPS_REMOTE_HOST',
            'OLLAMA_VPS_REMOTE_PORT'
        )) {
            if (-not $vals.ContainsKey($key)) { continue }
            $candidate = [string]$vals[$key]
            if ([string]::IsNullOrWhiteSpace($candidate)) { continue }

            $current = [Environment]::GetEnvironmentVariable($key, 'Process')
            if ([string]::IsNullOrWhiteSpace($current)) {
                [Environment]::SetEnvironmentVariable($key, $candidate, 'Process')
                Set-Item -Path ("Env:" + $key) -Value $candidate
            }
        }
    }
}

function Test-TruthyValue {
    param([string]$Value)

    $normalized = ''
    if (-not [string]::IsNullOrWhiteSpace($Value)) {
        $normalized = $Value.Trim().ToLowerInvariant()
    }
    return ($normalized -in @('1', 'true', 'yes', 'si', 'on'))
}

function Rewrite-PostgresDSNForTunnel {
    param(
        [string]$Dsn,
        [int]$LocalPort
    )

    if ([string]::IsNullOrWhiteSpace($Dsn)) {
        return $Dsn
    }

    $rewritten = $Dsn
    $rewritten = $rewritten -replace '@127\.0\.0\.1:5432/', ("@127.0.0.1:{0}/" -f $LocalPort)
    $rewritten = $rewritten -replace '@localhost:5432/', ("@127.0.0.1:{0}/" -f $LocalPort)
    return $rewritten
}

function Rewrite-OllamaBaseUrlForTunnel {
    param(
        [string]$BaseUrl,
        [int]$LocalPort
    )

    if ([string]::IsNullOrWhiteSpace($BaseUrl)) {
        return ("http://127.0.0.1:{0}" -f $LocalPort)
    }

    $rewritten = $BaseUrl.Trim()
    $rewritten = $rewritten -replace 'http://127\.0\.0\.1:11434', ("http://127.0.0.1:{0}" -f $LocalPort)
    $rewritten = $rewritten -replace 'http://localhost:11434', ("http://127.0.0.1:{0}" -f $LocalPort)
    return $rewritten
}

function Ensure-VpsSshTunnel {
    param(
        [string]$BackendDir,
        [string]$SshHost,
        [string]$SshUser,
        [string]$SshKeyPath,
        [int]$LocalPort,
        [string]$RemoteHost,
        [int]$RemotePort,
        [string]$TunnelLabel
    )

    if ([string]::IsNullOrWhiteSpace($TunnelLabel)) {
        $TunnelLabel = 'servicio'
    }

    if ([string]::IsNullOrWhiteSpace($SshHost)) {
        throw ("{0}: SSH host es obligatorio cuando el tunel está habilitado." -f $TunnelLabel)
    }

    if ([string]::IsNullOrWhiteSpace($SshUser)) {
        throw ("{0}: SSH user es obligatorio cuando el tunel está habilitado." -f $TunnelLabel)
    }

    if ([string]::IsNullOrWhiteSpace($SshKeyPath)) {
        $SshKeyPath = '..\clave privada ssh.ppk'
    }

    if ([System.IO.Path]::IsPathRooted($SshKeyPath)) {
        $resolvedKeyPath = $SshKeyPath
    } else {
        $resolvedKeyPath = Join-Path $BackendDir $SshKeyPath
    }

    $resolvedKeyPath = [System.IO.Path]::GetFullPath($resolvedKeyPath)
    if (-not (Test-Path $resolvedKeyPath)) {
        throw ("{0}: no se encontro la llave SSH: {1}" -f $TunnelLabel, $resolvedKeyPath)
    }

    $plink = Get-Command plink.exe -ErrorAction SilentlyContinue
    if ($null -eq $plink) {
        throw ("{0}: no se encontro plink.exe. Instala PuTTY para habilitar el tunel." -f $TunnelLabel)
    }

    $listening = Get-NetTCPConnection -LocalPort $LocalPort -State Listen -ErrorAction SilentlyContinue
    if ($listening) {
        Write-Info ("Tunel {0} detectado en localhost:{1}. Se reutiliza." -f $TunnelLabel, $LocalPort)
        return
    }

    $forwardSpec = "{0}:{1}:{2}" -f $LocalPort, $RemoteHost, $RemotePort
    $target = "{0}@{1}" -f $SshUser, $SshHost

    $keyArg = $resolvedKeyPath
    if ($keyArg -match '\s') {
        $keyArg = '"' + ($keyArg -replace '"', '\"') + '"'
    }
    $plinkArgs = @('-batch', '-N', '-i', $keyArg, '-L', $forwardSpec, $target)

    $tmpDir = Join-Path $BackendDir 'tmp'
    if (-not (Test-Path $tmpDir)) {
        New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null
    }

    $safeLabel = ($TunnelLabel -replace '[^a-zA-Z0-9_-]', '_').ToLowerInvariant()
    $plinkStdOut = Join-Path $tmpDir ("plink_tunnel_{0}_{1}.out.log" -f $safeLabel, $LocalPort)
    $plinkStdErr = Join-Path $tmpDir ("plink_tunnel_{0}_{1}.err.log" -f $safeLabel, $LocalPort)
    Remove-Item -Path $plinkStdOut -Force -ErrorAction SilentlyContinue
    Remove-Item -Path $plinkStdErr -Force -ErrorAction SilentlyContinue

    $proc = Start-Process -FilePath $plink.Source -ArgumentList $plinkArgs -WindowStyle Hidden -RedirectStandardOutput $plinkStdOut -RedirectStandardError $plinkStdErr -PassThru
    if ($null -eq $proc -or $proc.HasExited) {
        throw ("{0}: no se pudo iniciar tunel SSH a {1} ({2})." -f $TunnelLabel, $target, $forwardSpec)
    }

    $listenerReady = $false
    $maxAttempts = 16
    for ($attempt = 1; $attempt -le $maxAttempts; $attempt++) {
        $listenerCheck = Get-NetTCPConnection -LocalPort $LocalPort -State Listen -ErrorAction SilentlyContinue
        if ($listenerCheck) {
            $listenerReady = $true
            break
        }

        if ($proc.HasExited) {
            break
        }

        Start-Sleep -Milliseconds 500
    }

    $diagnosticDetail = ''
    $stderrTail = @()
    if (Test-Path $plinkStdErr) {
        $stderrTail = @(Get-Content -Path $plinkStdErr -Tail 10 -ErrorAction SilentlyContinue)
    }
    if ($stderrTail.Count -gt 0) {
        $diagnosticDetail = ' stderr=' + (($stderrTail -join ' | ').Trim())
    } else {
        $stdoutTail = @()
        if (Test-Path $plinkStdOut) {
            $stdoutTail = @(Get-Content -Path $plinkStdOut -Tail 10 -ErrorAction SilentlyContinue)
        }
        if ($stdoutTail.Count -gt 0) {
            $diagnosticDetail = ' stdout=' + (($stdoutTail -join ' | ').Trim())
        }
    }

    if (-not $listenerReady) {
        if ($proc.HasExited) {
            throw ("{0}: el tunel SSH se cerro al iniciar (PID={1}, ExitCode={2}) para {3} ({4}).{5}" -f $TunnelLabel, $proc.Id, $proc.ExitCode, $target, $forwardSpec, $diagnosticDetail)
        }
        throw ("{0}: no se detecto listener en localhost:{1} tras iniciar tunel SSH (PID={2}) hacia {3} ({4}).{5}" -f $TunnelLabel, $LocalPort, $proc.Id, $target, $forwardSpec, $diagnosticDetail)
    }

    Write-Info ("Tunel {0} iniciado: localhost:{1} -> {2}:{3} (PID={4})" -f $TunnelLabel, $LocalPort, $RemoteHost, $RemotePort, $proc.Id)
}

function Ensure-VpsPostgresTunnel {
    param(
        [string]$BackendDir,
        [string]$SshHost,
        [string]$SshUser,
        [string]$SshKeyPath,
        [int]$LocalPort,
        [string]$RemoteHost,
        [int]$RemotePort
    )

    Ensure-VpsSshTunnel -BackendDir $BackendDir -SshHost $SshHost -SshUser $SshUser -SshKeyPath $SshKeyPath -LocalPort $LocalPort -RemoteHost $RemoteHost -RemotePort $RemotePort -TunnelLabel 'DB'
}

Load-PostgresEnvFromFiles -BackendDir $backend

# Validar modo de base de datos PostgreSQL-only
Write-Step "2/8 Validando configuracion de base de datos (PostgreSQL)"

if (-not $env:DB_DIALECT) {
    $env:DB_DIALECT = 'postgres'
}

$tunnelEnabled = Test-TruthyValue -Value ([Environment]::GetEnvironmentVariable('DB_VPS_TUNNEL_ENABLED', 'Process'))
if ($tunnelEnabled) {
    $sshHost = [Environment]::GetEnvironmentVariable('DB_VPS_SSH_HOST', 'Process')
    $sshUser = [Environment]::GetEnvironmentVariable('DB_VPS_SSH_USER', 'Process')
    $sshKeyPath = [Environment]::GetEnvironmentVariable('DB_VPS_SSH_KEY_PATH', 'Process')
    $remoteHost = [Environment]::GetEnvironmentVariable('DB_VPS_REMOTE_HOST', 'Process')
    if ([string]::IsNullOrWhiteSpace($remoteHost)) {
        $remoteHost = '127.0.0.1'
    }

    $localPort = 15432
    $rawLocalPort = [Environment]::GetEnvironmentVariable('DB_VPS_LOCAL_PORT', 'Process')
    if (-not [string]::IsNullOrWhiteSpace($rawLocalPort)) {
        $parsedLocal = 0
        if ([int]::TryParse($rawLocalPort, [ref]$parsedLocal) -and $parsedLocal -gt 0) {
            $localPort = $parsedLocal
        }
    }

    $remotePort = 5432
    $rawRemotePort = [Environment]::GetEnvironmentVariable('DB_VPS_REMOTE_PORT', 'Process')
    if (-not [string]::IsNullOrWhiteSpace($rawRemotePort)) {
        $parsedRemote = 0
        if ([int]::TryParse($rawRemotePort, [ref]$parsedRemote) -and $parsedRemote -gt 0) {
            $remotePort = $parsedRemote
        }
    }

    Ensure-VpsPostgresTunnel -BackendDir $backend -SshHost $sshHost -SshUser $sshUser -SshKeyPath $sshKeyPath -LocalPort $localPort -RemoteHost $remoteHost -RemotePort $remotePort

    $env:DB_EMPRESAS_DSN = Rewrite-PostgresDSNForTunnel -Dsn $env:DB_EMPRESAS_DSN -LocalPort $localPort
    $env:DB_SUPERADMIN_DSN = Rewrite-PostgresDSNForTunnel -Dsn $env:DB_SUPERADMIN_DSN -LocalPort $localPort
    [Environment]::SetEnvironmentVariable('DB_EMPRESAS_DSN', $env:DB_EMPRESAS_DSN, 'Process')
    [Environment]::SetEnvironmentVariable('DB_SUPERADMIN_DSN', $env:DB_SUPERADMIN_DSN, 'Process')
}

$ollamaTunnelEnabled = Test-TruthyValue -Value ([Environment]::GetEnvironmentVariable('OLLAMA_VPS_TUNNEL_ENABLED', 'Process'))
if ($ollamaTunnelEnabled) {
    $ollamaSshHost = [Environment]::GetEnvironmentVariable('OLLAMA_VPS_SSH_HOST', 'Process')
    if ([string]::IsNullOrWhiteSpace($ollamaSshHost)) {
        $ollamaSshHost = [Environment]::GetEnvironmentVariable('DB_VPS_SSH_HOST', 'Process')
    }

    $ollamaSshUser = [Environment]::GetEnvironmentVariable('OLLAMA_VPS_SSH_USER', 'Process')
    if ([string]::IsNullOrWhiteSpace($ollamaSshUser)) {
        $ollamaSshUser = [Environment]::GetEnvironmentVariable('DB_VPS_SSH_USER', 'Process')
    }

    $ollamaSshKeyPath = [Environment]::GetEnvironmentVariable('OLLAMA_VPS_SSH_KEY_PATH', 'Process')
    if ([string]::IsNullOrWhiteSpace($ollamaSshKeyPath)) {
        $ollamaSshKeyPath = [Environment]::GetEnvironmentVariable('DB_VPS_SSH_KEY_PATH', 'Process')
    }

    $ollamaRemoteHost = [Environment]::GetEnvironmentVariable('OLLAMA_VPS_REMOTE_HOST', 'Process')
    if ([string]::IsNullOrWhiteSpace($ollamaRemoteHost)) {
        $ollamaRemoteHost = '127.0.0.1'
    }

    $ollamaLocalPort = 11435
    $rawOllamaLocalPort = [Environment]::GetEnvironmentVariable('OLLAMA_VPS_LOCAL_PORT', 'Process')
    if (-not [string]::IsNullOrWhiteSpace($rawOllamaLocalPort)) {
        $parsedOllamaLocal = 0
        if ([int]::TryParse($rawOllamaLocalPort, [ref]$parsedOllamaLocal) -and $parsedOllamaLocal -gt 0) {
            $ollamaLocalPort = $parsedOllamaLocal
        }
    }

    $ollamaRemotePort = 11434
    $rawOllamaRemotePort = [Environment]::GetEnvironmentVariable('OLLAMA_VPS_REMOTE_PORT', 'Process')
    if (-not [string]::IsNullOrWhiteSpace($rawOllamaRemotePort)) {
        $parsedOllamaRemote = 0
        if ([int]::TryParse($rawOllamaRemotePort, [ref]$parsedOllamaRemote) -and $parsedOllamaRemote -gt 0) {
            $ollamaRemotePort = $parsedOllamaRemote
        }
    }

    Ensure-VpsSshTunnel -BackendDir $backend -SshHost $ollamaSshHost -SshUser $ollamaSshUser -SshKeyPath $ollamaSshKeyPath -LocalPort $ollamaLocalPort -RemoteHost $ollamaRemoteHost -RemotePort $ollamaRemotePort -TunnelLabel 'OLLAMA'

    $env:OLLAMA_BASE_URL = Rewrite-OllamaBaseUrlForTunnel -BaseUrl ([Environment]::GetEnvironmentVariable('OLLAMA_BASE_URL', 'Process')) -LocalPort $ollamaLocalPort
    [Environment]::SetEnvironmentVariable('OLLAMA_BASE_URL', $env:OLLAMA_BASE_URL, 'Process')
    Write-Info ("OLLAMA_BASE_URL configurado para backend local: {0}" -f $env:OLLAMA_BASE_URL)
}

if ($env:DB_DIALECT -ne 'postgres') {
    Write-Host "DB_DIALECT=$($env:DB_DIALECT) no es valido. Este proyecto opera solo con PostgreSQL." -ForegroundColor Red
    exit 1
}

if (-not $env:DB_EMPRESAS_DSN -or -not $env:DB_SUPERADMIN_DSN) {
    Write-Host "Faltan DSN de PostgreSQL. Define DB_EMPRESAS_DSN y DB_SUPERADMIN_DSN en backend/.env.local o en el entorno." -ForegroundColor Red
    exit 1
}

Write-Ok "Configuracion PostgreSQL validada."
Write-Info "DB_DIALECT=$env:DB_DIALECT"
Write-Info "DB_EMPRESAS_DSN configurado: $([string]::IsNullOrWhiteSpace($env:DB_EMPRESAS_DSN) -eq $false)"
Write-Info "DB_SUPERADMIN_DSN configurado: $([string]::IsNullOrWhiteSpace($env:DB_SUPERADMIN_DSN) -eq $false)"
Write-Info "OLLAMA_BASE_URL configurado: $([string]::IsNullOrWhiteSpace($env:OLLAMA_BASE_URL) -eq $false)"

function Stop-ProcessesOnPort {
    param([int]$port)

    function Get-ListeningPidsOnPort {
        param([int]$TargetPort)

        $detected = @()

        # Preferir API nativa cuando este disponible (mas confiable que parsear netstat).
        try {
            $connections = Get-NetTCPConnection -LocalPort $TargetPort -ErrorAction Stop |
                Where-Object { $_.State -eq 'Listen' }
            if ($connections) {
                $detected = $connections | Select-Object -ExpandProperty OwningProcess -Unique
            }
        } catch {
            # Fallback a netstat para entornos donde Get-NetTCPConnection no este disponible.
        }

        $detected = @($detected)

        if ($detected.Count -eq 0) {
            $netstatOut = netstat -ano -p tcp | findstr "LISTENING" 2>$null
            if (-not [string]::IsNullOrWhiteSpace($netstatOut)) {
                $detected = ($netstatOut -split "\r?\n" | ForEach-Object {
                    $line = [string]$_
                    if ([string]::IsNullOrWhiteSpace($line)) { return }

                    if ($line -match '^\s*TCP\s+(\S+):(\d+)\s+\S+\s+LISTENING\s+(\d+)\s*$') {
                        $localPort = 0
                        if ([int]::TryParse($matches[2], [ref]$localPort) -and $localPort -eq $TargetPort) {
                            $matches[3]
                        }
                    }
                }) | Where-Object { $_ } | Select-Object -Unique
            }
        }

        $detected = @($detected)

        return @($detected | Where-Object {
            $pidCandidate = 0
            [int]::TryParse([string]$_, [ref]$pidCandidate) -and $pidCandidate -gt 0
        })
    }

    function Get-ProcessMetadata {
        param([int]$TargetPid)

        $meta = @{
            PID = $TargetPid
            Name = 'desconocido'
            CommandLine = ''
        }

        try {
            $proc = Get-CimInstance Win32_Process -Filter ("ProcessId = {0}" -f $TargetPid) -ErrorAction Stop
            if ($proc) {
                $meta.Name = [string]$proc.Name
                $meta.CommandLine = [string]$proc.CommandLine
                return $meta
            }
        } catch {
            # fallback a Get-Process
        }

        try {
            $p = Get-Process -Id $TargetPid -ErrorAction Stop
            if ($p) {
                $meta.Name = [string]$p.ProcessName
            }
        } catch {
            # conservar valores por defecto
        }

        return $meta
    }

    function Is-ManagedServerProcess {
        param([hashtable]$Meta)

        $name = ([string]$Meta.Name).ToLowerInvariant()
        $cmd = [string]$Meta.CommandLine

        $managedNames = @('server', 'server.exe', 'pos-backend', 'pos-backend.exe')
        if ($managedNames -contains $name) {
            return $true
        }

        if (($name -eq 'go' -or $name -eq 'go.exe') -and -not [string]::IsNullOrWhiteSpace($cmd)) {
            $cmdLower = $cmd.ToLowerInvariant()
            if ($cmdLower.Contains(' go run ') -or $cmdLower.Contains('go.exe run')) {
                if ($cmdLower.Contains('powerfulcontrolsystem') -or $cmdLower.Contains('backend')) {
                    return $true
                }
            }
        }

        if (-not [string]::IsNullOrWhiteSpace($cmd)) {
            $cmdLower = $cmd.ToLowerInvariant()
            if ($cmdLower.Contains('server.exe') -or $cmdLower.Contains('pos-backend')) {
                return $true
            }
        }

        return $false
    }

    Write-Info ("Comprobando puerto {0}..." -f $port)
    $pids = Get-ListeningPidsOnPort -TargetPort $port
    $pids = @($pids)

    $blockedPids = @()
    if ($pids.Count -gt 0) {
        $joined = $pids -join ', '
        Write-WarnMsg ("Procesos detectados en puerto {0}: {1}" -f $port, $joined)
        foreach ($killPid in $pids) {
            $pidNum = 0
            if (-not [int]::TryParse([string]$killPid, [ref]$pidNum)) {
                Write-WarnMsg ("PID inválido detectado en puerto {0}: {1}" -f $port, $killPid)
                continue
            }

            if ($pidNum -eq $PID) {
                Write-WarnMsg ("Se omite el proceso actual (PID {0}) para evitar cortar la terminal." -f $pidNum)
                continue
            }

            $meta = Get-ProcessMetadata -TargetPid $pidNum
            if (-not (Is-ManagedServerProcess -Meta $meta)) {
                Write-WarnMsg ("PID {0} ({1}) usa el puerto {2}, pero no es un proceso gestionado del backend. No se finalizara automaticamente." -f $pidNum, $meta.Name, $port)
                $blockedPids += $meta
                continue
            }

            try {
                Write-Info ("Terminando proceso backend PID {0} ({1})..." -f $pidNum, $meta.Name)
                Stop-Process -Id $pidNum -Force -ErrorAction Stop
                Write-Ok ("PID {0} terminado." -f $pidNum)
            } catch {
                $msg = if ($_.Exception) { $_.Exception.Message } else { $_ }
                Write-WarnMsg ("No se pudo terminar PID {0}: {1}" -f $pidNum, $msg)
            }
        }

        if ($blockedPids.Count -gt 0) {
            $blockedSummary = ($blockedPids | ForEach-Object { "{0}({1})" -f $_.PID, $_.Name }) -join ', '
            throw ("Puerto {0} ocupado por procesos no gestionados: {1}. Cierra esos procesos manualmente y vuelve a ejecutar el script." -f $port, $blockedSummary)
        }

        Start-Sleep -Seconds 1
    } else {
        Write-Info "No hay procesos escuchando en el puerto $port"
    }

    # Además intentar cerrar procesos por nombre comunes del servidor
    $names = @('server','server.exe','pos-backend','pos-backend.exe')
    foreach ($name in $names) {
        try {
            $procFilter = $name -replace '\.exe$',''
            $procs = Get-Process -Name $procFilter -ErrorAction SilentlyContinue
            if ($procs) {
                foreach ($p in $procs) {
                    try {
                        Write-Info ("Terminando proceso por nombre {0} (PID {1})..." -f $p.ProcessName, $p.Id)
                        Stop-Process -Id $p.Id -Force -ErrorAction Stop
                        Write-Ok "Proceso terminado."
                    } catch {
                        Write-WarnMsg ("No se pudo terminar proceso {0}: {1}" -f $p.Id, $_)
                    }
                }
            }
        } catch {
            # ignorar errores de Get-Process
        }
    }
}

function Test-ServerAvailability {
    param(
        [int]$Port,
        [string]$BaseUrl
    )

    $tcpListening = $false
    try {
        $tcpListening = @(Get-NetTCPConnection -LocalPort $Port -State Listen -ErrorAction Stop).Count -gt 0
    } catch {
        $tcpListening = ((netstat -ano | findstr (":" + $Port) 2>$null) -match 'LISTENING').Count -gt 0
    }

    if ($tcpListening) {
        return $true
    }

    try {
        $response = Invoke-WebRequest -Uri $BaseUrl -Method Get -UseBasicParsing -MaximumRedirection 0 -TimeoutSec 5 -ErrorAction Stop
        if ($null -ne $response -and $response.StatusCode -ge 200 -and $response.StatusCode -lt 500) {
            return $true
        }
    } catch {
        if ($_.Exception -and $_.Exception.Response) {
            try {
                $statusCode = [int]$_.Exception.Response.StatusCode
                if ($statusCode -ge 200 -and $statusCode -lt 500) {
                    return $true
                }
            } catch {
            }
        }
    }

    return $false
}

Write-Step "3/8 Liberando puerto 8080"
Stop-ProcessesOnPort -port 8080

Write-Step "4/8 Verificando dependencias Go"
Write-Info "Ejecutando: go mod tidy"
go mod tidy
if ($LASTEXITCODE -ne 0) {
    Write-ErrMsg "Fallo en 'go mod tidy' (codigo $LASTEXITCODE)."
    Pop-Location
    exit $LASTEXITCODE
}

Write-Ok "Dependencias verificadas."

Write-Step "5/8 Resolviendo credenciales OAuth y variables de entorno"
# Resolver credenciales desde entorno o backend/.env(.local)
$oauthCreds = Resolve-GoogleOAuthCredentials -BackendDir $backend
$clientId = $oauthCreds.ClientId
$clientSecret = $oauthCreds.ClientSecret

if ([string]::IsNullOrWhiteSpace($clientId) -or [string]::IsNullOrWhiteSpace($clientSecret)) {
    Write-WarnMsg "No se encontraron GOOGLE_CLIENT_ID/GOOGLE_CLIENT_SECRET en entorno o .env; el backend intentara resolverlos desde la DB."
}

$env:GOOGLE_REDIRECT_URL = "http://localhost:8080/auth/google/callback"

if (-not [string]::IsNullOrWhiteSpace($clientId)) {
    $env:GOOGLE_CLIENT_ID = $clientId
}
if (-not [string]::IsNullOrWhiteSpace($clientSecret)) {
    $env:GOOGLE_CLIENT_SECRET = $clientSecret
}

# Forzar puerto 8080 (usuario solicitó usar solo 8080)
$env:PORT = "8080"

try {
    $encKeyResult = Resolve-ConfigEncryptionKey -BackendDir $backend
    $env:CONFIG_ENC_KEY = [string]$encKeyResult.Value
    if ($encKeyResult.Generated) {
        Write-Ok ("CONFIG_ENC_KEY autogenerada y persistida en: {0}" -f $encKeyResult.Source)
    } else {
        Write-Info ("CONFIG_ENC_KEY cargada desde: {0}" -f $encKeyResult.Source)
    }
} catch {
    $msg = if ($_.Exception) { $_.Exception.Message } else { $_ }
    Write-ErrMsg $msg
    Pop-Location
    exit 1
}

if (-not [string]::IsNullOrWhiteSpace($clientId) -and -not [string]::IsNullOrWhiteSpace($clientSecret)) {
    Write-Info ("Credenciales OAuth cargadas desde: {0}" -f $oauthCreds.Source)
    if (-not (Test-GoogleOAuthCredentials -ClientId $clientId -ClientSecret $clientSecret -RedirectURL $env:GOOGLE_REDIRECT_URL)) {
        Write-WarnMsg "OAuth de entorno/.env no valido; se omitira para que el backend intente resolver credenciales desde DB."
        $clientId = ""
        $clientSecret = ""
        Remove-Item Env:GOOGLE_CLIENT_ID -ErrorAction SilentlyContinue
        Remove-Item Env:GOOGLE_CLIENT_SECRET -ErrorAction SilentlyContinue
    }
} else {
    Write-WarnMsg "Se continua sin credenciales OAuth de entorno; backend resolvera desde DB si existen."
}

# Compilar el binario para ejecutar con un entorno controlado
Write-Step "6/8 Compilando backend"
Write-Info "Compilando el servidor (go build -o server.exe .)"
& go build -o server.exe .
if ($LASTEXITCODE -ne 0) {
    Write-ErrMsg "Fallo en go build (codigo $LASTEXITCODE)."
    Pop-Location
    exit $LASTEXITCODE
}
Write-Ok "Compilacion completada."

Write-Step "7/8 Lanzando servidor"
Write-Info "Lanzando server.exe con entorno controlado (logs en backend/server.log y backend/server.err)."
$serverPath = Join-Path $backend "server.exe"
# Asegurar que las variables de entorno estén en el proceso actual (Start-Process heredará estas en Windows PowerShell)
if ($env:PORT) { $env:PORT = $env:PORT }
if ($clientId) { $env:GOOGLE_CLIENT_ID = $clientId }
if ($clientSecret) { $env:GOOGLE_CLIENT_SECRET = $clientSecret }
if ($env:GOOGLE_REDIRECT_URL) { $env:GOOGLE_REDIRECT_URL = $env:GOOGLE_REDIRECT_URL }
if ($env:CONFIG_ENC_KEY) { $env:CONFIG_ENC_KEY = $env:CONFIG_ENC_KEY }
$env:PCS_SERVER_START_REASON = "inicio_script_iniciar_servidor"
Write-Info "Motivo de arranque enviado al backend: $env:PCS_SERVER_START_REASON"

# Iniciar sin -Environment para compatibilidad con Windows PowerShell 5.1
$serverProc = Start-Process -FilePath $serverPath -WorkingDirectory $backend -RedirectStandardOutput (Join-Path $backend "server.log") -RedirectStandardError (Join-Path $backend "server.err") -PassThru
if (-not $serverProc) {
    Write-ErrMsg "No se pudo iniciar server.exe"
    Pop-Location
    exit 1
}
Write-Ok ("Proceso del servidor lanzado. PID={0}" -f $serverProc.Id)

if ($Background) {
    Write-Ok "Servidor iniciado en modo background."
    Write-Info ("URL esperada: http://localhost:{0}" -f $env:PORT)
    Write-Info ("Log stdout: {0}" -f (Join-Path $backend "server.log"))
    Write-Info ("Log stderr: {0}" -f (Join-Path $backend "server.err"))
    Pop-Location
    exit 0
}

# Esperar a que el backend abra el puerto y/o responda HTTP
$maxWait = 30  # segundos
$waited = 0
$listenPort = 8080
if (-not [string]::IsNullOrWhiteSpace($env:PORT)) {
    $parsedPort = 0
    if ([int]::TryParse($env:PORT, [ref]$parsedPort) -and $parsedPort -gt 0) {
        $listenPort = $parsedPort
    }
}
$baseUrl = "http://localhost:{0}" -f $listenPort
Write-Step "8/8 Esperando disponibilidad del servidor"
Write-Info ("Esperando a que {0} responda (timeout {1}s)..." -f $baseUrl, $maxWait)
while ($waited -lt $maxWait) {
    Start-Sleep -Seconds 1
    $waited++

    if ($serverProc.HasExited) {
        Write-ErrMsg ("server.exe finalizo antes de quedar disponible en el puerto {0} (ExitCode={1})." -f $listenPort, $serverProc.ExitCode)
        $errPath = Join-Path $backend "server.err"
        if (Test-Path $errPath) {
            Write-WarnMsg "Ultimas lineas de backend/server.err:"
            Get-Content -Path $errPath -Tail 40 | ForEach-Object { Write-Host $_ }
        }
        Pop-Location
        exit 1
    }

    $listening = Test-ServerAvailability -Port $listenPort -BaseUrl $baseUrl
    if ($listening) { break }
}

if ($listening) {
    Write-Ok ("Servidor disponible en puerto {0}." -f $listenPort)
    Write-Info ("Direccion: {0}" -f $baseUrl)
    # Abrir en navegador por defecto
    try {
        Write-Info "Abriendo navegador por defecto..."
        Start-Process $baseUrl
    } catch {
        Write-WarnMsg "No se pudo abrir el navegador automaticamente: $_"
    }
    Write-Info ("Log stdout: {0}" -f (Join-Path $backend "server.log"))
    Write-Info ("Log stderr: {0}" -f (Join-Path $backend "server.err"))
} else {
    Write-WarnMsg "El servidor no respondio en ${maxWait}s. Verifica logs."
    $errPath = Join-Path $backend "server.err"
    if (Test-Path $errPath) {
        Write-WarnMsg "Ultimas lineas de backend/server.err:"
        Get-Content -Path $errPath -Tail 40 | ForEach-Object { Write-Host $_ }
    }
}

Pop-Location
