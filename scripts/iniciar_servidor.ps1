param(
    [switch]$Background,
    [int]$Port = 8080,
    [switch]$NoKillLocalPort,
    [switch]$UseVpsTunnel,
    [switch]$NoVpsTunnel
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

if ($Port -lt 1 -or $Port -gt 65535) {
    Write-ErrMsg "Puerto invalido: $Port. Usa un valor entre 1 y 65535."
    exit 1
}
if ($UseVpsTunnel -and $NoVpsTunnel) {
    Write-ErrMsg "UseVpsTunnel y NoVpsTunnel son opciones opuestas. Usa solo una."
    exit 1
}
$localServerPort = $Port

Write-Step "1/8 Preparando entorno"


$backend = Join-Path $PSScriptRoot "..\backend"
if (-not (Test-Path $backend)) {
    Write-ErrMsg "No se encontro la carpeta backend en: $backend"
    exit 1
}
Write-Info "Directorio backend detectado: $backend"
Write-Info "Modo local seguro: este script no apaga ni reinicia servicios de la VPS."
Write-Info ("Puerto local seleccionado: {0}" -f $localServerPort)
if ($Background) {
    Write-Info "Modo background activo: no se abrira navegador automaticamente."
}
if ($NoKillLocalPort) {
    Write-Info "NoKillLocalPort activo: no se cerraran procesos locales aunque el puerto este ocupado."
}
if ($NoVpsTunnel) {
    Write-Info "NoVpsTunnel activo: no se abrira tunel SSH hacia la VPS."
}
if ($UseVpsTunnel) {
    Write-Info "UseVpsTunnel activo: se intentara abrir tunel SSH local hacia PostgreSQL del VPS."
}
Push-Location $backend

function Import-DotEnvValues {
    param([string]$Path)
    if (-not (Test-Path $Path)) { return @{} }

    $lines = @(Get-Content -Path $Path -ErrorAction SilentlyContinue)
    return Import-DotEnvLines -Lines $lines
}

function Import-DotEnvLines {
    param([string[]]$Lines)
    $map = @{}

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
            'DB_VPS_SSH_PORT',
            'DB_VPS_SSH_KEY_PATH',
            'DB_VPS_SSH_HOSTKEY',
            'DB_VPS_REMOTE_APP_PATH',
            'DB_VPS_POSTGRES_CONTAINER',
            'DB_VPS_LOCAL_PORT',
            'DB_VPS_REMOTE_HOST',
            'DB_VPS_REMOTE_PORT',
            'GEMINI_API_KEY'
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

function Get-ProjectRootFromBackend {
    param([string]$BackendDir)

    $backendFull = [System.IO.Path]::GetFullPath($BackendDir)
    $root = Split-Path -Parent $backendFull
    return $root
}

function Set-ProcessEnvIfEmpty {
    param(
        [string]$Key,
        [string]$Value
    )

    if ([string]::IsNullOrWhiteSpace($Key) -or [string]::IsNullOrWhiteSpace($Value)) {
        return
    }

    $current = [Environment]::GetEnvironmentVariable($Key, 'Process')
    if ([string]::IsNullOrWhiteSpace($current)) {
        [Environment]::SetEnvironmentVariable($Key, $Value, 'Process')
        Set-Item -Path ("Env:" + $Key) -Value $Value
    }
}

function ConvertTo-PostgresDsnComponent {
    param([string]$Value)

    if ($null -eq $Value) {
        return ''
    }
    return [System.Uri]::EscapeDataString([string]$Value)
}

function New-LocalPostgresDSN {
    param(
        [string]$User,
        [string]$Password,
        [string]$Database,
        [string]$HostName = '127.0.0.1',
        [int]$Port = 5432
    )

    $safeUser = ConvertTo-PostgresDsnComponent -Value $User
    $safePassword = ConvertTo-PostgresDsnComponent -Value $Password
    $safeDatabase = ConvertTo-PostgresDsnComponent -Value $Database
    return "postgres://${safeUser}:${safePassword}@${HostName}:${Port}/${safeDatabase}?sslmode=disable"
}

function Load-PostgresEnvFromPlatformFallback {
    param([string]$BackendDir)

    if (-not ([string]::IsNullOrWhiteSpace($env:DB_EMPRESAS_DSN)) -and -not ([string]::IsNullOrWhiteSpace($env:DB_SUPERADMIN_DSN))) {
        return
    }

    $root = Get-ProjectRootFromBackend -BackendDir $BackendDir
    $platformEnvPath = Join-Path $root 'deploy\.env.platform'
    if (-not (Test-Path $platformEnvPath)) {
        return
    }

    $vals = Import-DotEnvValues -Path $platformEnvPath
    if (-not $vals.ContainsKey('POSTGRES_PASSWORD') -or [string]::IsNullOrWhiteSpace([string]$vals['POSTGRES_PASSWORD'])) {
        return
    }

    $user = 'pcs'
    if ($vals.ContainsKey('POSTGRES_USER') -and -not [string]::IsNullOrWhiteSpace([string]$vals['POSTGRES_USER'])) {
        $user = [string]$vals['POSTGRES_USER']
    }

    Set-ProcessEnvIfEmpty -Key 'DB_EMPRESAS_DSN' -Value (New-LocalPostgresDSN -User $user -Password ([string]$vals['POSTGRES_PASSWORD']) -Database 'pcs_empresas')
    Set-ProcessEnvIfEmpty -Key 'DB_SUPERADMIN_DSN' -Value (New-LocalPostgresDSN -User $user -Password ([string]$vals['POSTGRES_PASSWORD']) -Database 'pcs_superadministrador')
    Write-Info "DSN PostgreSQL local derivados desde deploy/.env.platform (valores sensibles no se imprimen)."
}

function Get-DeploymentScriptStringValue {
    param(
        [string[]]$Lines,
        [string]$VariableName
    )

    $pattern = '^\s*\$script:' + [regex]::Escape($VariableName) + '\s*=\s*[''"]([^''"]+)[''"]'
    foreach ($line in $Lines) {
        $match = [regex]::Match([string]$line, $pattern)
        if ($match.Success) {
            return $match.Groups[1].Value
        }
    }
    return ''
}

function Get-DeploymentScriptIntValue {
    param(
        [string[]]$Lines,
        [string]$VariableName
    )

    $pattern = '^\s*\$script:' + [regex]::Escape($VariableName) + '\s*=\s*(\d+)'
    foreach ($line in $Lines) {
        $match = [regex]::Match([string]$line, $pattern)
        if ($match.Success) {
            return $match.Groups[1].Value
        }
    }
    return ''
}

function Load-VpsTunnelEnvFromDeploymentConfig {
    param([string]$BackendDir)

    $root = Get-ProjectRootFromBackend -BackendDir $BackendDir
    $deploymentConfig = Join-Path $root 'scripts\pcs_deployment.local.ps1'
    if (-not (Test-Path $deploymentConfig)) {
        return
    }

    $lines = @(Get-Content -Path $deploymentConfig -ErrorAction SilentlyContinue)
    if ($lines.Count -eq 0) {
        return
    }

    $loaded = 0
    $sshHost = Get-DeploymentScriptStringValue -Lines $lines -VariableName 'PcsVpsHost'
    $sshUser = Get-DeploymentScriptStringValue -Lines $lines -VariableName 'PcsVpsUser'
    $sshPort = Get-DeploymentScriptIntValue -Lines $lines -VariableName 'PcsVpsPort'
    $sshKey = Get-DeploymentScriptStringValue -Lines $lines -VariableName 'PcsVpsIdentityFile'
    $sshHostKey = Get-DeploymentScriptStringValue -Lines $lines -VariableName 'PcsVpsHostKey'
    $remotePath = Get-DeploymentScriptStringValue -Lines $lines -VariableName 'PcsVpsRemotePath'

    if (-not [string]::IsNullOrWhiteSpace($sshHost)) {
        Set-ProcessEnvIfEmpty -Key 'DB_VPS_SSH_HOST' -Value $sshHost
        $loaded++
    }
    if (-not [string]::IsNullOrWhiteSpace($sshUser)) {
        Set-ProcessEnvIfEmpty -Key 'DB_VPS_SSH_USER' -Value $sshUser
        $loaded++
    }
    if (-not [string]::IsNullOrWhiteSpace($sshPort)) {
        Set-ProcessEnvIfEmpty -Key 'DB_VPS_SSH_PORT' -Value $sshPort
        $loaded++
    }
    if (-not [string]::IsNullOrWhiteSpace($sshKey)) {
        Set-ProcessEnvIfEmpty -Key 'DB_VPS_SSH_KEY_PATH' -Value $sshKey
        $loaded++
    }
    if (-not [string]::IsNullOrWhiteSpace($sshHostKey)) {
        Set-ProcessEnvIfEmpty -Key 'DB_VPS_SSH_HOSTKEY' -Value $sshHostKey
        $loaded++
    }
    if (-not [string]::IsNullOrWhiteSpace($remotePath)) {
        Set-ProcessEnvIfEmpty -Key 'DB_VPS_REMOTE_APP_PATH' -Value $remotePath
        $loaded++
    }

    if ($loaded -gt 0) {
        Write-Info "Configuracion SSH VPS cargada desde scripts/pcs_deployment.local.ps1 (valores no se imprimen)."
    }
}

function Resolve-VpsSshKeyPath {
    param(
        [string]$BackendDir,
        [string]$SshKeyPath
    )

    if ([string]::IsNullOrWhiteSpace($SshKeyPath)) {
        $SshKeyPath = '..\clave privada ssh.ppk'
    }

    if ([System.IO.Path]::IsPathRooted($SshKeyPath)) {
        return [System.IO.Path]::GetFullPath($SshKeyPath)
    }

    return [System.IO.Path]::GetFullPath((Join-Path $BackendDir $SshKeyPath))
}

function Invoke-VpsSshText {
    param(
        [string]$BackendDir,
        [string]$SshHost,
        [string]$SshUser,
        [string]$SshKeyPath,
        [string]$SshHostKey,
        [int]$SshPort,
        [string]$Command
    )

    if ([string]::IsNullOrWhiteSpace($SshHost) -or [string]::IsNullOrWhiteSpace($SshUser)) {
        return @()
    }

    $resolvedKeyPath = Resolve-VpsSshKeyPath -BackendDir $BackendDir -SshKeyPath $SshKeyPath
    if (-not (Test-Path $resolvedKeyPath)) {
        Write-WarnMsg "No se encontro la llave SSH local para leer configuracion remota del VPS."
        return @()
    }

    $plink = Get-Command plink.exe -ErrorAction SilentlyContinue
    if ($null -eq $plink) {
        Write-WarnMsg "No se encontro plink.exe; no se puede leer configuracion remota del VPS."
        return @()
    }

    $hostKeyArgs = @()
    if (-not [string]::IsNullOrWhiteSpace($SshHostKey)) {
        $hostKeyArgs = @('-hostkey', $SshHostKey.Trim())
    }

    $target = "{0}@{1}" -f $SshUser, $SshHost
    $plinkArgs = @('-batch') + $hostKeyArgs + @('-P', "$SshPort", '-i', $resolvedKeyPath, $target, $Command)
    $output = @(& $plink.Source @plinkArgs 2>&1)
    if ($LASTEXITCODE -ne 0) {
        Write-WarnMsg "No se pudo leer la configuracion remota de PostgreSQL por SSH; se usara la configuracion local si existe."
        return @()
    }

    return $output
}

function Load-PostgresEnvFromVpsPlatformFallback {
    param([string]$BackendDir)

    if (-not ([string]::IsNullOrWhiteSpace($env:DB_EMPRESAS_DSN)) -and -not ([string]::IsNullOrWhiteSpace($env:DB_SUPERADMIN_DSN))) {
        return
    }

    $sshHost = [Environment]::GetEnvironmentVariable('DB_VPS_SSH_HOST', 'Process')
    $sshUser = [Environment]::GetEnvironmentVariable('DB_VPS_SSH_USER', 'Process')
    $sshKeyPath = [Environment]::GetEnvironmentVariable('DB_VPS_SSH_KEY_PATH', 'Process')
    $sshHostKey = [Environment]::GetEnvironmentVariable('DB_VPS_SSH_HOSTKEY', 'Process')
    $remotePath = [Environment]::GetEnvironmentVariable('DB_VPS_REMOTE_APP_PATH', 'Process')
    if ([string]::IsNullOrWhiteSpace($remotePath)) {
        $remotePath = '/root/powerfulcontrolsystem'
    }
    if ($remotePath -notmatch '^[a-zA-Z0-9_./-]+$') {
        Write-WarnMsg "Ruta remota VPS no tiene formato seguro para lectura automatica; configura DSN localmente."
        return
    }

    $sshPort = 49222
    $rawSshPort = [Environment]::GetEnvironmentVariable('DB_VPS_SSH_PORT', 'Process')
    if (-not [string]::IsNullOrWhiteSpace($rawSshPort)) {
        $parsedSshPort = 0
        if ([int]::TryParse($rawSshPort, [ref]$parsedSshPort) -and $parsedSshPort -gt 0) {
            $sshPort = $parsedSshPort
        }
    }

    $command = "cd $remotePath && if [ -f deploy/.env.platform ]; then cat deploy/.env.platform; fi"
    $remoteEnvLines = @(Invoke-VpsSshText -BackendDir $BackendDir -SshHost $sshHost -SshUser $sshUser -SshKeyPath $sshKeyPath -SshHostKey $sshHostKey -SshPort $sshPort -Command $command)
    if ($remoteEnvLines.Count -eq 0) {
        return
    }

    $vals = Import-DotEnvLines -Lines $remoteEnvLines
    if (-not $vals.ContainsKey('POSTGRES_PASSWORD') -or [string]::IsNullOrWhiteSpace([string]$vals['POSTGRES_PASSWORD'])) {
        Write-WarnMsg "La configuracion remota no contiene POSTGRES_PASSWORD usable."
        return
    }

    $user = 'pcs'
    if ($vals.ContainsKey('POSTGRES_USER') -and -not [string]::IsNullOrWhiteSpace([string]$vals['POSTGRES_USER'])) {
        $user = [string]$vals['POSTGRES_USER']
    }

    Set-ProcessEnvIfEmpty -Key 'DB_EMPRESAS_DSN' -Value (New-LocalPostgresDSN -User $user -Password ([string]$vals['POSTGRES_PASSWORD']) -Database 'pcs_empresas')
    Set-ProcessEnvIfEmpty -Key 'DB_SUPERADMIN_DSN' -Value (New-LocalPostgresDSN -User $user -Password ([string]$vals['POSTGRES_PASSWORD']) -Database 'pcs_superadministrador')
    Write-Info "DSN PostgreSQL derivados desde configuracion remota del VPS por SSH (valores sensibles no se imprimen)."
}

function Resolve-VpsPostgresContainerHost {
    param(
        [string]$BackendDir,
        [string]$SshHost,
        [string]$SshUser,
        [string]$SshKeyPath,
        [string]$SshHostKey,
        [int]$SshPort
    )

    $containerName = [Environment]::GetEnvironmentVariable('DB_VPS_POSTGRES_CONTAINER', 'Process')
    if ([string]::IsNullOrWhiteSpace($containerName)) {
        $containerName = 'pcs-postgres'
    }
    if ($containerName -notmatch '^[a-zA-Z0-9_.-]+$') {
        Write-WarnMsg "Nombre de contenedor PostgreSQL VPS no tiene formato seguro; se usara 127.0.0.1."
        return ''
    }

    $command = "docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $containerName 2>/dev/null"
    $lines = @(Invoke-VpsSshText -BackendDir $BackendDir -SshHost $SshHost -SshUser $SshUser -SshKeyPath $SshKeyPath -SshHostKey $SshHostKey -SshPort $SshPort -Command $command)
    foreach ($line in $lines) {
        $candidate = ([string]$line).Trim()
        if ($candidate -match '^\d{1,3}(\.\d{1,3}){3}$') {
            Write-Info "PostgreSQL VPS resuelto desde contenedor Docker configurado (direccion no se imprime)."
            return $candidate
        }
    }

    Write-WarnMsg "No se pudo resolver IP del contenedor PostgreSQL VPS; se usara 127.0.0.1."
    return ''
}

function Get-PostgresDSNEndpoint {
    param([string]$Dsn)

    if ([string]::IsNullOrWhiteSpace($Dsn)) {
        return $null
    }

    try {
        $uri = [System.Uri]$Dsn
        if ([string]::IsNullOrWhiteSpace($uri.Host)) {
            return $null
        }
        $port = $uri.Port
        if ($port -le 0) {
            $port = 5432
        }
        return @{
            Host = $uri.Host
            Port = [int]$port
        }
    } catch {
        return $null
    }
}

function Test-TcpListener {
    param(
        [string]$HostName,
        [int]$Port,
        [int]$TimeoutMs = 1500
    )

    if ([string]::IsNullOrWhiteSpace($HostName) -or $Port -le 0) {
        return $false
    }

    $client = New-Object System.Net.Sockets.TcpClient
    try {
        $iar = $client.BeginConnect($HostName, $Port, $null, $null)
        if (-not $iar.AsyncWaitHandle.WaitOne($TimeoutMs, $false)) {
            return $false
        }
        $client.EndConnect($iar)
        return $true
    } catch {
        return $false
    } finally {
        $client.Close()
    }
}

function Assert-PostgresEndpointReachable {
    param([string]$Dsn)

    $endpoint = Get-PostgresDSNEndpoint -Dsn $Dsn
    if ($null -eq $endpoint) {
        Write-WarnMsg "No se pudo interpretar el endpoint del DSN PostgreSQL; se continuara hasta la validacion del backend."
        return
    }

    $hostName = [string]$endpoint.Host
    $port = [int]$endpoint.Port
    if (-not (Test-TcpListener -HostName $hostName -Port $port)) {
        Write-ErrMsg ("No hay PostgreSQL escuchando en {0}:{1}." -f $hostName, $port)
        Write-Host "Para pruebas locales tienes tres opciones:" -ForegroundColor Yellow
        Write-Host "  1. Levantar PostgreSQL local con las bases pcs_empresas y pcs_superadministrador." -ForegroundColor Yellow
        Write-Host "  2. Activar tunel VPS con DB_VPS_TUNNEL_ENABLED=1 y DSN apuntando a localhost." -ForegroundColor Yellow
        Write-Host "  3. Definir DB_EMPRESAS_DSN y DB_SUPERADMIN_DSN hacia una instancia PostgreSQL accesible." -ForegroundColor Yellow
        exit 1
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

function Ensure-VpsSshTunnel {
    param(
        [string]$BackendDir,
        [string]$SshHost,
        [string]$SshUser,
        [string]$SshKeyPath,
        [string]$SshHostKey,
        [int]$SshPort = 49222,
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

    $forwardSpec = "{0}:{1}:{2}" -f $LocalPort, $RemoteHost, $RemotePort
    $listening = Get-NetTCPConnection -LocalPort $LocalPort -State Listen -ErrorAction SilentlyContinue
    if ($listening) {
        $pids = @($listening | Select-Object -ExpandProperty OwningProcess -Unique)
        $matchingTunnel = $false
        $stoppedStaleTunnel = $false

        foreach ($pidNum in $pids) {
            $procInfo = Get-CimInstance Win32_Process -Filter ("ProcessId={0}" -f [int]$pidNum) -ErrorAction SilentlyContinue
            $cmdLine = ''
            $procName = ''
            if ($procInfo) {
                $cmdLine = [string]$procInfo.CommandLine
                $procName = [string]$procInfo.Name
            }

            if ($cmdLine.Contains($forwardSpec)) {
                $matchingTunnel = $true
                break
            }

            if ($procName -ieq 'plink.exe' -and $cmdLine.Contains('-L') -and $cmdLine.Contains(("{0}:" -f $LocalPort))) {
                Stop-Process -Id ([int]$pidNum) -Force -ErrorAction SilentlyContinue
                $stoppedStaleTunnel = $true
            }
        }

        if ($matchingTunnel) {
            Write-Info ("Tunel {0} detectado en localhost:{1}. Se reutiliza." -f $TunnelLabel, $LocalPort)
            return
        }

        if ($stoppedStaleTunnel) {
            Write-Info ("Tunel {0} previo en localhost:{1} apuntaba a otro destino; se recreara." -f $TunnelLabel, $LocalPort)
            Start-Sleep -Milliseconds 800
        } else {
            throw ("{0}: localhost:{1} esta ocupado por otro proceso y no se puede abrir el tunel." -f $TunnelLabel, $LocalPort)
        }
    }

    $target = "{0}@{1}" -f $SshUser, $SshHost
    $displayTarget = 'servidor VPS configurado'

    $keyArg = $resolvedKeyPath
    if ($keyArg -match '\s') {
        $keyArg = '"' + ($keyArg -replace '"', '\"') + '"'
    }
    $hostKeyArgs = @()
    if (-not [string]::IsNullOrWhiteSpace($SshHostKey)) {
        $hostKeyArgs = @('-hostkey', $SshHostKey.Trim())
    }
    $plinkArgs = @('-batch') + $hostKeyArgs + @('-N', '-P', "$SshPort", '-i', $keyArg, '-L', $forwardSpec, $target)

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
        throw ("{0}: no se pudo iniciar tunel SSH al {1}:{2} ({3})." -f $TunnelLabel, $displayTarget, $SshPort, $forwardSpec)
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
            throw ("{0}: el tunel SSH se cerro al iniciar (PID={1}, ExitCode={2}) para el {3}:{4} ({5}).{6}" -f $TunnelLabel, $proc.Id, $proc.ExitCode, $displayTarget, $SshPort, $forwardSpec, $diagnosticDetail)
        }
        throw ("{0}: no se detecto listener en localhost:{1} tras iniciar tunel SSH (PID={2}) hacia el {3}:{4} ({5}).{6}" -f $TunnelLabel, $LocalPort, $proc.Id, $displayTarget, $SshPort, $forwardSpec, $diagnosticDetail)
    }

    Write-Info ("Tunel {0} iniciado: localhost:{1} -> {2}:{3} (PID={4})" -f $TunnelLabel, $LocalPort, $RemoteHost, $RemotePort, $proc.Id)
}

function Ensure-VpsPostgresTunnel {
    param(
        [string]$BackendDir,
        [string]$SshHost,
        [string]$SshUser,
        [string]$SshKeyPath,
        [string]$SshHostKey,
        [int]$SshPort = 49222,
        [int]$LocalPort,
        [string]$RemoteHost,
        [int]$RemotePort
    )

    Ensure-VpsSshTunnel -BackendDir $BackendDir -SshHost $SshHost -SshUser $SshUser -SshKeyPath $SshKeyPath -SshHostKey $SshHostKey -SshPort $SshPort -LocalPort $LocalPort -RemoteHost $RemoteHost -RemotePort $RemotePort -TunnelLabel 'DB'
}

Load-PostgresEnvFromFiles -BackendDir $backend
Load-VpsTunnelEnvFromDeploymentConfig -BackendDir $backend
if ($UseVpsTunnel) {
    [Environment]::SetEnvironmentVariable('DB_VPS_TUNNEL_ENABLED', '1', 'Process')
    Set-Item -Path 'Env:DB_VPS_TUNNEL_ENABLED' -Value '1'
}
$initialTunnelEnabled = Test-TruthyValue -Value ([Environment]::GetEnvironmentVariable('DB_VPS_TUNNEL_ENABLED', 'Process'))
if ($initialTunnelEnabled -and -not $NoVpsTunnel) {
    Load-PostgresEnvFromVpsPlatformFallback -BackendDir $backend
}
Load-PostgresEnvFromPlatformFallback -BackendDir $backend

# Validar modo de base de datos PostgreSQL-only
Write-Step "2/8 Validando configuracion de base de datos (PostgreSQL)"

if (-not $env:DB_DIALECT) {
    $env:DB_DIALECT = 'postgres'
}

$tunnelEnabled = Test-TruthyValue -Value ([Environment]::GetEnvironmentVariable('DB_VPS_TUNNEL_ENABLED', 'Process'))
if ($NoVpsTunnel -and $tunnelEnabled) {
    Write-WarnMsg "DB_VPS_TUNNEL_ENABLED esta activo, pero se omitira por -NoVpsTunnel para una prueba 100% local."
    $tunnelEnabled = $false
    [Environment]::SetEnvironmentVariable('DB_VPS_TUNNEL_ENABLED', '0', 'Process')
}
if ($tunnelEnabled) {
    $sshHost = [Environment]::GetEnvironmentVariable('DB_VPS_SSH_HOST', 'Process')
    $sshUser = [Environment]::GetEnvironmentVariable('DB_VPS_SSH_USER', 'Process')
    $sshKeyPath = [Environment]::GetEnvironmentVariable('DB_VPS_SSH_KEY_PATH', 'Process')
    $sshHostKey = [Environment]::GetEnvironmentVariable('DB_VPS_SSH_HOSTKEY', 'Process')
    if ([string]::IsNullOrWhiteSpace($sshHostKey)) {
        $sshHostKey = [Environment]::GetEnvironmentVariable('PCS_VPS_SSH_HOSTKEY', 'Process')
    }
    $sshPort = 49222
    $rawSshPort = [Environment]::GetEnvironmentVariable('DB_VPS_SSH_PORT', 'Process')
    if (-not [string]::IsNullOrWhiteSpace($rawSshPort)) {
        $parsedSshPort = 0
        if ([int]::TryParse($rawSshPort, [ref]$parsedSshPort) -and $parsedSshPort -gt 0) {
            $sshPort = $parsedSshPort
        }
    }
    $remoteHost = [Environment]::GetEnvironmentVariable('DB_VPS_REMOTE_HOST', 'Process')
    $remotePort = 5432
    $rawRemotePort = [Environment]::GetEnvironmentVariable('DB_VPS_REMOTE_PORT', 'Process')
    if (-not [string]::IsNullOrWhiteSpace($rawRemotePort)) {
        $parsedRemote = 0
        if ([int]::TryParse($rawRemotePort, [ref]$parsedRemote) -and $parsedRemote -gt 0) {
            $remotePort = $parsedRemote
        }
    }
    if ([string]::IsNullOrWhiteSpace($remoteHost)) {
        $remoteHost = Resolve-VpsPostgresContainerHost -BackendDir $backend -SshHost $sshHost -SshUser $sshUser -SshKeyPath $sshKeyPath -SshHostKey $sshHostKey -SshPort $sshPort
    }
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

    Ensure-VpsPostgresTunnel -BackendDir $backend -SshHost $sshHost -SshUser $sshUser -SshKeyPath $sshKeyPath -SshHostKey $sshHostKey -SshPort $sshPort -LocalPort $localPort -RemoteHost $remoteHost -RemotePort $remotePort

    $env:DB_EMPRESAS_DSN = Rewrite-PostgresDSNForTunnel -Dsn $env:DB_EMPRESAS_DSN -LocalPort $localPort
    $env:DB_SUPERADMIN_DSN = Rewrite-PostgresDSNForTunnel -Dsn $env:DB_SUPERADMIN_DSN -LocalPort $localPort
    [Environment]::SetEnvironmentVariable('DB_EMPRESAS_DSN', $env:DB_EMPRESAS_DSN, 'Process')
    [Environment]::SetEnvironmentVariable('DB_SUPERADMIN_DSN', $env:DB_SUPERADMIN_DSN, 'Process')
}

if ($env:DB_DIALECT -ne 'postgres') {
    Write-Host "DB_DIALECT=$($env:DB_DIALECT) no es valido. Este proyecto opera solo con PostgreSQL." -ForegroundColor Red
    exit 1
}

if (-not $env:DB_EMPRESAS_DSN -or -not $env:DB_SUPERADMIN_DSN) {
    Write-Host "Faltan DSN de PostgreSQL. Define DB_EMPRESAS_DSN y DB_SUPERADMIN_DSN en backend/.env.local o en el entorno." -ForegroundColor Red
    Write-Host "Para pruebas locales, configura DB_VPS_TUNNEL_ENABLED=1 si vas a usar PostgreSQL del VPS por tunel SSH." -ForegroundColor Yellow
    exit 1
}

Assert-PostgresEndpointReachable -Dsn $env:DB_EMPRESAS_DSN
Assert-PostgresEndpointReachable -Dsn $env:DB_SUPERADMIN_DSN

Write-Ok "Configuracion PostgreSQL validada."
Write-Info "DB_DIALECT=$env:DB_DIALECT"
Write-Info "DB_EMPRESAS_DSN configurado: $([string]::IsNullOrWhiteSpace($env:DB_EMPRESAS_DSN) -eq $false)"
Write-Info "DB_SUPERADMIN_DSN configurado: $([string]::IsNullOrWhiteSpace($env:DB_SUPERADMIN_DSN) -eq $false)"
Write-Info "GEMINI_API_KEY configurado: $([string]::IsNullOrWhiteSpace($env:GEMINI_API_KEY) -eq $false)"

function Stop-ProcessesOnPort {
    param(
        [int]$port,
        [switch]$AlsoStopManagedServerNames
    )

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
    if (-not $AlsoStopManagedServerNames) {
        return
    }

    # Solo se usa cuando se solicita expresamente; para pruebas locales no debe
    # tocar servidores que esten escuchando en otros puertos.
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

Write-Step ("3/8 Revisando puerto local {0}" -f $localServerPort)
if ($NoKillLocalPort) {
    Write-WarnMsg ("Se omitio liberar el puerto local {0}. Si ya esta ocupado, el backend puede fallar al iniciar." -f $localServerPort)
} else {
    Stop-ProcessesOnPort -port $localServerPort
}

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

$env:GOOGLE_REDIRECT_URL = "http://localhost:{0}/auth/google/callback" -f $localServerPort

if (-not [string]::IsNullOrWhiteSpace($clientId)) {
    $env:GOOGLE_CLIENT_ID = $clientId
}
if (-not [string]::IsNullOrWhiteSpace($clientSecret)) {
    $env:GOOGLE_CLIENT_SECRET = $clientSecret
}

# Puerto local de pruebas. Por defecto es 8080, pero puede cambiarse con -Port.
$env:PORT = [string]$localServerPort

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
if ([string]::IsNullOrWhiteSpace($env:PCS_SKIP_CORPORATE_EMAIL_STARTUP_SYNC)) {
    $env:PCS_SKIP_CORPORATE_EMAIL_STARTUP_SYNC = "1"
}
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
$listenPort = $localServerPort
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
