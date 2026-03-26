param(
    [string]$Message = "Actualización automática desde script: añadir/actualizar archivos",
    [string]$RepoUrl = "",
    [switch]$SkipChangeLog,
    [switch]$NoForce
)

Write-Host "Verificando estado de git..."
if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
    Write-Host "ERROR: Git no está disponible en PATH. Instala Git o ejecuta manualmente los comandos de git." -ForegroundColor Red
    exit 1
}

$root = Resolve-Path "$PSScriptRoot\.."
Set-Location $root
# Preparar logging detallado (transcripción)
$logsDir = Join-Path $root 'scripts\logs'
New-Item -ItemType Directory -Force -Path $logsDir | Out-Null
$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$logFile = Join-Path $logsDir ("actualizar_repositorio-$timestamp.log")
try {
    Start-Transcript -Path $logFile -Append -ErrorAction SilentlyContinue
} catch {
    Write-Host "Aviso: no fue posible iniciar Start-Transcript: $_" -ForegroundColor Yellow
}

Write-Host "Preparando cambios para commit..."
# Ensure a sensible .gitignore exists to avoid committing build artifacts
$gitignorePath = Join-Path $root '.gitignore'
if (-not (Test-Path $gitignorePath)) {
    Write-Host "Creando .gitignore básico..."
    @(
        "# Binarios y logs de backend"
        "backend/server.exe"
        "backend/*.log"
        "backend/*.err"
        "backend/*.exe~"
        "backend/pos.db"
        "*.sqlite"
        "*.db"
        "# VSCode"
        ".vscode/"
        "# Archivos compilados"
        "bin/"
        "obj/"
    ) | Out-File -FilePath $gitignorePath -Encoding UTF8
    git add .gitignore | Out-Null
}

# Stage all changes
git add -A

$status = git status --porcelain
if ([string]::IsNullOrEmpty($status)) {
    Write-Host "No hay cambios para commitear." -ForegroundColor Yellow
    try { Stop-Transcript -ErrorAction SilentlyContinue } catch {}
    exit 0
}

Write-Host "Creando commit: $Message"
git commit -m $Message
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: El commit falló. Revisa 'git status'." -ForegroundColor Red
    try { Stop-Transcript -ErrorAction SilentlyContinue } catch {}
    exit $LASTEXITCODE
}

function Try-PushOrigin {
    $forceArg = if ($NoForce) { "" } else { "--force" }
    git push $forceArg -u origin HEAD 2>&1
    return $LASTEXITCODE
}

Write-Host "Intentando push al remoto 'origin' (forzando por defecto)..."
$pushCode = Try-PushOrigin
if ($pushCode -eq 0) {
    $pushMsg = 'OK'
    Write-Host "Éxito: cambios empujados al remoto 'origin'." -ForegroundColor Green
} else {
    Write-Warning "Push falló (código $pushCode). Verificando remoto 'origin'..."
    $originUrl = git remote get-url origin 2>$null
    if (-not $originUrl) {
        # Try to get RepoUrl from parameter or environment
        if ([string]::IsNullOrWhiteSpace($RepoUrl)) {
            $RepoUrl = $env:REPO_URL
        }
        if (-not [string]::IsNullOrWhiteSpace($RepoUrl)) {
            Write-Host "Agregando remoto 'origin' desde parámetro/entorno: $RepoUrl"
            git remote add origin $RepoUrl
            if ($LASTEXITCODE -ne 0) {
                Write-Host "ERROR: No se pudo agregar remoto 'origin'." -ForegroundColor Red
                $pushMsg = 'FAIL_ADD_REMOTE'
            } else {
                Write-Host "Remoto 'origin' agregado. Reintentando push forzado..."
                $pushCode = Try-PushOrigin
                if ($pushCode -eq 0) {
                    $pushMsg = 'OK'
                    Write-Host "Éxito: cambios empujados al remoto 'origin'." -ForegroundColor Green
                } else {
                    $pushMsg = 'FAIL_PUSH'
                    Write-Host "ERROR: El push falló incluso después de añadir remoto. Código: $pushCode" -ForegroundColor Red
                }
            }
        } else {
            Write-Host "ERROR: No existe el remoto 'origin' y no se proporcionó -RepoUrl ni la variable de entorno REPO_URL." -ForegroundColor Red
            $pushMsg = 'NO_ORIGIN'
        }
    } else {
        Write-Host "Remoto 'origin' existe ($originUrl) pero el push falló. Intentando push forzado de todos modos..." -ForegroundColor Yellow
        $pushCode = Try-PushOrigin
        if ($pushCode -eq 0) {
            $pushMsg = 'OK'
            Write-Host "Éxito: push forzado completado." -ForegroundColor Green
        } else {
            $pushMsg = 'FAIL_PUSH'
            Write-Host "ERROR: No fue posible pushear los cambios. Código: $pushCode" -ForegroundColor Red
        }
    }
}

if (-not $SkipChangeLog) {
    $hist = Join-Path $root "documentos\historial_de_cambios"
    if (Test-Path $hist) {
        $files = git show --name-only --pretty="" HEAD | Where-Object { $_ -ne '' } | ForEach-Object { "- $_" } | Out-String
        $entry = "$(Get-Date -Format yyyy-MM-dd): $Message`nPushStatus: $pushMsg`nArchivos modificados:`n$files"
        Add-Content -Path $hist -Value "`n$entry"
        git add $hist | Out-Null
        git commit -m "Actualizar historial_de_cambios: registro automático" | Out-Null
        # Intentar pushear el historial también (forzando si corresponde)
        if ($pushMsg -eq 'OK') {
            git push origin HEAD | Out-Null
        } else {
            $forceArg = if ($NoForce) { "" } else { "--force" }
            git push $forceArg origin HEAD | Out-Null
            if ($LASTEXITCODE -ne 0) {
                Write-Host "Aviso: historial actualizado localmente pero no fue posible pushearlo al remoto." -ForegroundColor Yellow
            }
        }
    } else {
        Write-Host "Aviso: no se encontró 'documentos/historial_de_cambios' para actualizar." -ForegroundColor Yellow
    }
}

if ($pushMsg -eq 'OK') {
    Write-Host "Operación completada con éxito." -ForegroundColor Green
    try { Stop-Transcript -ErrorAction SilentlyContinue } catch {}
    exit 0
} else {
    switch ($pushMsg) {
        'NO_ORIGIN' { Write-Host "Falló: no existe el remoto 'origin'. Vuelve a ejecutar el script con -RepoUrl <url> para agregarlo automáticamente." -ForegroundColor Red }
        'FAIL_ADD_REMOTE' { Write-Host "Falló: no se pudo agregar el remoto 'origin'. Revisa la URL o permisos." -ForegroundColor Red }
        default { Write-Host "Falló: no se pudo pushear los cambios al remoto. Revisa la configuración de Git y permisos." -ForegroundColor Red }
    }
    try { Stop-Transcript -ErrorAction SilentlyContinue } catch {}
    exit 1
}
