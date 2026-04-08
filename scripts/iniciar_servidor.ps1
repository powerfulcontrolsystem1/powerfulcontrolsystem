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
Clear-Host

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

# Asegurar carpeta backend\db y mover .db existentes a esa carpeta
Write-Step "2/8 Preparando rutas de base de datos"
Write-Info "Asegurando carpeta de bases de datos: $backend\db"
New-Item -ItemType Directory -Path (Join-Path $backend 'db') -Force | Out-Null

$dbFolder = Join-Path $backend 'db'
# Mover archivos .db que estén en backend a backend\db (si existen)
$dbFiles = Get-ChildItem -Path $backend -Filter '*.db' -File -ErrorAction SilentlyContinue
if ($dbFiles) {
    foreach ($f in $dbFiles) {
        try {
            $dest = Join-Path $dbFolder $f.Name
            if (-not (Test-Path $dest)) {
                Write-Info ("Moviendo {0} -> {1}" -f $f.FullName, $dest)
                Move-Item -Path $f.FullName -Destination $dest -ErrorAction Stop
            } else {
                Write-WarnMsg ("Ya existe {0}, se omite movimiento." -f $dest)
            }
        } catch {
            Write-WarnMsg ("No se pudo mover $($f.Name): $_")
        }
    }
} else {
    Write-Info "No se encontraron archivos .db en la raiz de backend."
}

# Establecer variables de entorno para rutas de DB si no están definidas
if (-not $env:DB_EMPRESAS_PATH) { $env:DB_EMPRESAS_PATH = Join-Path $dbFolder 'empresas.db' }
if (-not $env:DB_SUPERADMIN_PATH) { $env:DB_SUPERADMIN_PATH = Join-Path $dbFolder 'superadministrador.db' }
if (-not $env:DB_POS_PATH) { $env:DB_POS_PATH = Join-Path $dbFolder 'pos.db' }

Write-Ok "Rutas de DB preparadas."
Write-Info "DB_EMPRESAS_PATH=$env:DB_EMPRESAS_PATH"
Write-Info "DB_SUPERADMIN_PATH=$env:DB_SUPERADMIN_PATH"
Write-Info "DB_POS_PATH=$env:DB_POS_PATH"

function Stop-ProcessesOnPort {
    param([int]$port)
    Write-Info ("Comprobando puerto {0}..." -f $port)
    $netstatOut = netstat -ano | findstr ":$port" 2>$null
    if (-not [string]::IsNullOrWhiteSpace($netstatOut)) {
        $pids = ($netstatOut -split "\r?\n" | ForEach-Object {
            if ($_ -match '\s+(\d+)$') { $matches[1] }
        }) | Where-Object { $_ } | Select-Object -Unique
        $pids = @($pids)
        if ($pids.Count -gt 0) {
            $joined = $pids -join ', '
            Write-WarnMsg ("Procesos detectados en puerto {0}: {1}" -f $port, $joined)
            foreach ($killPid in $pids) {
                try {
                    Write-Info ("Terminando proceso PID {0}..." -f $killPid)
                    taskkill /PID $killPid /F | Out-Null
                    Write-Ok ("PID {0} terminado." -f $killPid)
                } catch {
                    $msg = if ($_.Exception) { $_.Exception.Message } else { $_ }
                    Write-WarnMsg ("No se pudo terminar PID {0}: {1}" -f $killPid, $msg)
                }
            }
            Start-Sleep -Seconds 1
        }
    } else {
        Write-Info "No hay procesos detectados en el puerto $port"
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

# Esperar a que el puerto 8080 esté en LISTENING
$maxWait = 30  # segundos
$waited = 0
Write-Step "8/8 Esperando disponibilidad del servidor"
Write-Info "Esperando a que http://localhost:8080 responda (timeout ${maxWait}s)..."
while ($waited -lt $maxWait) {
    Start-Sleep -Seconds 1
    $waited++

    if ($serverProc.HasExited) {
        Write-ErrMsg ("server.exe finalizo antes de abrir el puerto 8080 (ExitCode={0})." -f $serverProc.ExitCode)
        $errPath = Join-Path $backend "server.err"
        if (Test-Path $errPath) {
            Write-WarnMsg "Ultimas lineas de backend/server.err:"
            Get-Content -Path $errPath -Tail 40 | ForEach-Object { Write-Host $_ }
        }
        Pop-Location
        exit 1
    }

    $listening = (netstat -ano | findstr ":8080" 2>$null) -match 'LISTENING'
    if ($listening) { break }
}

if ($listening) {
    Write-Ok "Servidor escuchando en puerto 8080."
    Write-Info "Direccion: http://localhost:8080"
    # Abrir en navegador por defecto
    try {
        Write-Info "Abriendo navegador por defecto..."
        Start-Process "http://localhost:8080"
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
