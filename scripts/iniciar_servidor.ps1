param(
    [switch]$Background
)

$clearMsg = "Limpiando la terminal..."
Write-Host $clearMsg -ForegroundColor DarkGray
Clear-Host

$backend = Join-Path $PSScriptRoot "..\backend"
Push-Location $backend

Write-Host "Buscando procesos que usan el puerto 8080..."
$port = 8080
$netstatOut = & netstat -ano | findstr ":$port" 2>$null
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
                Write-Host ("Terminando proceso PID {0}..." -f $killPid)
                taskkill /PID $killPid /F | Out-Null
                Write-Host ("PID {0} terminado." -f $killPid) -ForegroundColor Green
            } catch {
                $msg = if ($_.Exception) { $_.Exception.Message } else { $_ }
                Write-Host ("No se pudo terminar PID {0}: {1}" -f $killPid, $msg) -ForegroundColor Yellow
            }
        }
        Start-Sleep -Seconds 1
    }
}

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
# Intentar extraer credenciales de OAuth desde documentos/descripcion_del_proyecto
$descPath = Join-Path (Resolve-Path "$PSScriptRoot\..") "documentos\descripcion_del_proyecto"
$clientId = $null
$clientSecret = $null
if (Test-Path $descPath) {
    $content = Get-Content $descPath -Raw -ErrorAction SilentlyContinue
    if ($content) {
        # Buscar client id (contiene apps.googleusercontent.com)
        $m = [regex]::Match($content, "([0-9]+[-][A-Za-z0-9._-]+apps\.googleusercontent\.com)")
        if ($m.Success) { $clientId = $m.Groups[1].Value }
        # Buscar secreto: línea que contenga 'Secreto' y luego token
        $m2 = [regex]::Match($content, "Secreto[^:\n\r]*[:\-\s]*([A-Za-z0-9_\-]+)", 'IgnoreCase')
        if ($m2.Success) { $clientSecret = $m2.Groups[1].Value }
    }
}

# Exportar variables de entorno si se encontraron (no las logueamos)
if ($clientId) { $env:GOOGLE_CLIENT_ID = $clientId }
if ($clientSecret) { $env:GOOGLE_CLIENT_SECRET = $clientSecret }
$env:GOOGLE_REDIRECT_URL = "http://localhost:8080/auth/google/callback"

# Forzar puerto 8080 (usuario solicitó usar solo 8080)
$env:PORT = "8080"

if ($clientId) { Write-Host "Encontrado GOOGLE_CLIENT_ID" -ForegroundColor Green } else { Write-Host "GOOGLE_CLIENT_ID no encontrado en documentos/descripcion_del_proyecto" -ForegroundColor Yellow }
if ($clientSecret) { Write-Host "Encontrado GOOGLE_CLIENT_SECRET" -ForegroundColor Green } else { Write-Host "GOOGLE_CLIENT_SECRET no encontrado en documentos/descripcion_del_proyecto" -ForegroundColor Yellow }

# Compilar el binario para ejecutar con un entorno controlado
Write-Host "Compilando el servidor (go build -o server.exe .)" -ForegroundColor DarkGray
& go build -o server.exe .
if ($LASTEXITCODE -ne 0) {
    Write-Error "Fallo en go build (código $LASTEXITCODE)"
    Pop-Location
    exit $LASTEXITCODE
}

# Preparar mapa de entorno para Start-Process (no incluimos el secreto en logs)
$envMap = @{
    PORT = $env:PORT
    GOOGLE_CLIENT_ID = $clientId
    GOOGLE_CLIENT_SECRET = $clientSecret
    GOOGLE_REDIRECT_URL = $env:GOOGLE_REDIRECT_URL
}

Write-Host "Lanzando servidor.exe con entorno controlado..."
Start-Process -FilePath (Join-Path $backend "server.exe") -WorkingDirectory $backend -ArgumentList @() -NoNewWindow -PassThru -Environment $envMap | Out-Null
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
