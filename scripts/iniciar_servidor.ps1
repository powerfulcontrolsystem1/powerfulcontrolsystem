param(
    [switch]$Background
)

$clearMsg = "Limpiando la terminal..."
Write-Host $clearMsg -ForegroundColor DarkGray
Clear-Host

$backend = Join-Path $PSScriptRoot "..\backend"
Push-Location $backend

function Load-DotEnvValues {
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
        Join-Path $BackendDir '.env.local',
        Join-Path $BackendDir '.env'
    )

    foreach ($envPath in $envCandidates) {
        if (-not (Test-Path $envPath)) { continue }
        $vals = Load-DotEnvValues -Path $envPath
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
Write-Host "Asegurando carpeta de bases de datos: $backend\db" -ForegroundColor DarkGray
New-Item -ItemType Directory -Path (Join-Path $backend 'db') -Force | Out-Null

$dbFolder = Join-Path $backend 'db'
# Mover archivos .db que estén en backend a backend\db (si existen)
$dbFiles = Get-ChildItem -Path $backend -Filter '*.db' -File -ErrorAction SilentlyContinue
if ($dbFiles) {
    foreach ($f in $dbFiles) {
        try {
            $dest = Join-Path $dbFolder $f.Name
            if (-not (Test-Path $dest)) {
                Write-Host ("Moviendo {0} -> {1}" -f $f.FullName, $dest)
                Move-Item -Path $f.FullName -Destination $dest -ErrorAction Stop
            } else {
                Write-Host ("Ya existe {0}, omitiendo movimiento." -f $dest) -ForegroundColor Yellow
            }
        } catch {
            Write-Host ("No se pudo mover $($f.Name): $_") -ForegroundColor Yellow
        }
    }
} else {
    Write-Host "No se encontraron archivos .db en $backend" -ForegroundColor DarkGray
}

# Establecer variables de entorno para rutas de DB si no están definidas
if (-not $env:DB_EMPRESAS_PATH) { $env:DB_EMPRESAS_PATH = Join-Path $dbFolder 'empresas.db' }
if (-not $env:DB_SUPERADMIN_PATH) { $env:DB_SUPERADMIN_PATH = Join-Path $dbFolder 'superadministrador.db' }
if (-not $env:DB_POS_PATH) { $env:DB_POS_PATH = Join-Path $dbFolder 'pos.db' }

Write-Host "Rutas DB: $env:DB_EMPRESAS_PATH, $env:DB_SUPERADMIN_PATH, $env:DB_POS_PATH" -ForegroundColor Cyan

function Stop-ProcessesOnPort {
    param([int]$port)
    Write-Host ("Comprobando puerto {0}..." -f $port)
    $netstatOut = netstat -ano | findstr ":$port" 2>$null
    if (-not [string]::IsNullOrWhiteSpace($netstatOut)) {
        $pids = ($netstatOut -split "\r?\n" | ForEach-Object {
            if ($_ -match '\s+(\d+)$') { $matches[1] }
        }) | Where-Object { $_ } | Select-Object -Unique
        $pids = @($pids)
        if ($pids.Count -gt 0) {
            $joined = $pids -join ', '
            Write-Host ("Procesos detectados en puerto {0}: {1}" -f $port, $joined)
            foreach ($killPid in $pids) {
                try {
                    Write-Host (("Terminando proceso PID {0}..." -f $killPid))
                    taskkill /PID $killPid /F | Out-Null
                    Write-Host (("PID {0} terminado." -f $killPid)) -ForegroundColor Green
                } catch {
                    $msg = if ($_.Exception) { $_.Exception.Message } else { $_ }
                    Write-Host (("No se pudo terminar PID {0}: {1}" -f $killPid, $msg)) -ForegroundColor Yellow
                }
            }
            Start-Sleep -Seconds 1
        }
    } else {
        Write-Host "No hay procesos detectados en el puerto $port"
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
                        Write-Host (("Terminando proceso por nombre {0} (PID {1})..." -f $p.ProcessName, $p.Id))
                        Stop-Process -Id $p.Id -Force -ErrorAction Stop
                        Write-Host ("Proceso terminado.") -ForegroundColor Green
                    } catch {
                        Write-Host (("No se pudo terminar proceso {0}: {1}" -f $p.Id, $_)) -ForegroundColor Yellow
                    }
                }
            }
        } catch {
            # ignorar errores de Get-Process
        }
    }
}

Write-Host "Buscando procesos que usan el puerto 8080..."
Stop-ProcessesOnPort -port 8080

Write-Host "Ejecutando: go mod tidy"
go mod tidy
if ($LASTEXITCODE -ne 0) {
    Write-Error "Fallo en 'go mod tidy' (código $LASTEXITCODE)"
    Pop-Location
    exit $LASTEXITCODE
}

Write-Host "Iniciando servidor (carpeta: $backend)"

# Start server in a separate process so the script can continue and open the browser
$goArgs = @('run', '.')
Write-Host "Arrancando proceso: go $($goArgs -join ' ')"
# Resolver credenciales desde entorno o backend/.env(.local)
$oauthCreds = Resolve-GoogleOAuthCredentials -BackendDir $backend
$clientId = $oauthCreds.ClientId
$clientSecret = $oauthCreds.ClientSecret

if ([string]::IsNullOrWhiteSpace($clientId) -or [string]::IsNullOrWhiteSpace($clientSecret)) {
    Write-Host "No se encontraron GOOGLE_CLIENT_ID/GOOGLE_CLIENT_SECRET en entorno o .env; el backend intentará resolverlos desde la DB." -ForegroundColor Yellow
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
    Write-Host ("Credenciales OAuth cargadas desde: {0}" -f $oauthCreds.Source) -ForegroundColor Cyan
    if (-not (Test-GoogleOAuthCredentials -ClientId $clientId -ClientSecret $clientSecret -RedirectURL $env:GOOGLE_REDIRECT_URL)) {
        Write-Host "OAuth de entorno/.env no valido; se omitirá para que el backend intente resolver credenciales desde DB." -ForegroundColor Yellow
        $clientId = ""
        $clientSecret = ""
        Remove-Item Env:GOOGLE_CLIENT_ID -ErrorAction SilentlyContinue
        Remove-Item Env:GOOGLE_CLIENT_SECRET -ErrorAction SilentlyContinue
    }
} else {
    Write-Host "Se continúa sin credenciales OAuth de entorno; backend resolverá desde DB si existen." -ForegroundColor Yellow
}

# Compilar el binario para ejecutar con un entorno controlado
Write-Host "Compilando el servidor (go build -o server.exe .)" -ForegroundColor DarkGray
& go build -o server.exe .
if ($LASTEXITCODE -ne 0) {
    Write-Error "Fallo en go build (código $LASTEXITCODE)"
    Pop-Location
    exit $LASTEXITCODE
}

Write-Host "Lanzando servidor.exe con entorno controlado (logs en backend/server.log, backend/server.err)..."
$serverPath = Join-Path $backend "server.exe"
# Asegurar que las variables de entorno estén en el proceso actual (Start-Process heredará estas en Windows PowerShell)
if ($env:PORT) { $env:PORT = $env:PORT }
if ($clientId) { $env:GOOGLE_CLIENT_ID = $clientId }
if ($clientSecret) { $env:GOOGLE_CLIENT_SECRET = $clientSecret }
if ($env:GOOGLE_REDIRECT_URL) { $env:GOOGLE_REDIRECT_URL = $env:GOOGLE_REDIRECT_URL }

# Iniciar sin -Environment para compatibilidad con Windows PowerShell 5.1
Start-Process -FilePath $serverPath -WorkingDirectory $backend -RedirectStandardOutput (Join-Path $backend "server.log") -RedirectStandardError (Join-Path $backend "server.err") -PassThru | Out-Null
Write-Host "Proceso del servidor lanzado." -ForegroundColor Green

# Esperar a que el puerto 8080 esté en LISTENING
$maxWait = 30  # segundos
$waited = 0
Write-Host "Esperando a que http://localhost:8080 responda (timeout ${maxWait}s)..."
while ($waited -lt $maxWait) {
    Start-Sleep -Seconds 1
    $waited++
    $listening = (netstat -ano | findstr ":8080" 2>$null) -match 'LISTENING'
    if ($listening) { break }
}

if ($listening) {
    Write-Host "Servidor escuchando en puerto 8080." -ForegroundColor Green
    Write-Host "Dirección: http://localhost:8080" -ForegroundColor Cyan
    # Abrir en navegador por defecto
    try {
        Start-Process "http://localhost:8080"
    } catch {
        Write-Host "No se pudo abrir el navegador automáticamente: $_" -ForegroundColor Yellow
    }
} else {
    Write-Host "AVISO: el servidor no respondió en ${maxWait}s. Verifica logs." -ForegroundColor Yellow
}

Pop-Location
