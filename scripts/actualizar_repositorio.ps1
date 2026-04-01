param(
    [string]$Message = "Actualización automática desde script: añadir/actualizar archivos",
    [string]$RepoUrl = "",
    [switch]$SkipChangeLog,
    [switch]$ForcePush
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

function Request-ForcePushConfirmation {
    $answer = Read-Host "Confirmación requerida: escribe SI para ejecutar push forzado"
    return ($answer -ceq 'SI')
}

function Invoke-PushOrigin {
    param([switch]$UseForce)

    $forceArg = if ($UseForce) { "--force" } else { "" }
    $cmd = "git push $forceArg -u origin HEAD"
    Write-Host "Ejecutando: $cmd"
    $pushOutput = & git push $forceArg -u origin HEAD 2>&1
    $pushCode = $LASTEXITCODE
    # Guardar último output globalmente para análisis posterior
    Set-Variable -Name LastPushOutput -Value ($pushOutput -join "`n") -Scope Global
    return $pushCode
}

Write-Host "Intentando push al remoto 'origin' (modo normal)..."
$pushCode = Invoke-PushOrigin
$pushOutput = $Global:LastPushOutput
# Tratar como éxito si el código es 0 o la salida contiene "Everything up-to-date"
if ($pushCode -eq 0 -or ($pushOutput -match 'Everything up-?to-?date')) {
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
                Write-Host "Remoto 'origin' agregado. Reintentando push normal..."
                $pushCode = Invoke-PushOrigin
                $pushOutput = $Global:LastPushOutput
                if ($pushCode -eq 0 -or ($pushOutput -match 'Everything up-?to-?date')) {
                    $pushMsg = 'OK'
                    Write-Host "Éxito: cambios empujados al remoto 'origin'." -ForegroundColor Green
                } else {
                    if ($ForcePush) {
                        if (Request-ForcePushConfirmation) {
                            Write-Host "Reintentando push forzado por confirmación explícita..." -ForegroundColor Yellow
                            $pushCode = Invoke-PushOrigin -UseForce
                            $pushOutput = $Global:LastPushOutput
                            if ($pushCode -eq 0 -or ($pushOutput -match 'Everything up-?to-?date')) {
                                $pushMsg = 'OK'
                                Write-Host "Éxito: push forzado completado." -ForegroundColor Green
                            } else {
                                $pushMsg = 'FAIL_PUSH'
                                Write-Host "ERROR: El push forzado también falló. Código: $pushCode" -ForegroundColor Red
                            }
                        } else {
                            $pushMsg = 'FAIL_PUSH'
                            Write-Host "Push forzado cancelado por el usuario." -ForegroundColor Yellow
                        }
                    } else {
                        $pushMsg = 'FAIL_PUSH'
                        Write-Host "ERROR: El push falló incluso después de añadir remoto. Usa -ForcePush para habilitar reintento forzado con confirmación." -ForegroundColor Red
                    }
                }
            }
        } else {
            Write-Host "ERROR: No existe el remoto 'origin' y no se proporcionó -RepoUrl ni la variable de entorno REPO_URL." -ForegroundColor Red
            $pushMsg = 'NO_ORIGIN'
        }
    } else {
        if ($ForcePush) {
            if (Request-ForcePushConfirmation) {
                Write-Host "Remoto 'origin' existe ($originUrl). Reintentando push forzado por confirmación explícita..." -ForegroundColor Yellow
                $pushCode = Invoke-PushOrigin -UseForce
                $pushOutput = $Global:LastPushOutput
                if ($pushCode -eq 0 -or ($pushOutput -match 'Everything up-?to-?date')) {
                    $pushMsg = 'OK'
                    Write-Host "Éxito: push forzado completado." -ForegroundColor Green
                } else {
                    $pushMsg = 'FAIL_PUSH'
                    Write-Host "ERROR: No fue posible pushear los cambios aun con force. Código: $pushCode" -ForegroundColor Red
                }
            } else {
                $pushMsg = 'FAIL_PUSH'
                Write-Host "Push forzado cancelado por el usuario." -ForegroundColor Yellow
            }
        } else {
            $pushMsg = 'FAIL_PUSH'
            Write-Host "ERROR: El push normal falló. Reintenta con -ForcePush para habilitar force con confirmación explícita." -ForegroundColor Red
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
            if ($ForcePush -and (Request-ForcePushConfirmation)) {
                git push --force origin HEAD | Out-Null
            } else {
                git push origin HEAD | Out-Null
            }
            if ($LASTEXITCODE -ne 0) {
                Write-Host "Aviso: historial actualizado localmente pero no fue posible pushearlo al remoto." -ForegroundColor Yellow
            }
        }

        # También actualizar el documento `documentos/actualizaciones_del_repositorio.md`
        $actual = Join-Path $root "documentos\actualizaciones_del_repositorio.md"
        if (-not (Test-Path $actual)) {
            # Crear plantilla si no existe
            @(
                "# Actualizaciones del repositorio",
                "",
                "Este documento registra las actualizaciones automáticas realizadas por el script scripts/actualizar_repositorio.ps1.",
                ""
            ) | Out-File -FilePath $actual -Encoding UTF8
            git add $actual | Out-Null
            git commit -m "Crear actualizaciones_del_repositorio: registro automático" | Out-Null
        }

        # Obtener hash corto del commit actual
        $commitHash = (& git rev-parse --short HEAD 2>$null) | Out-String
        $commitHash = $commitHash.Trim()

        $entry2 = "$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss') - Mensaje: $Message; Commit: $commitHash; PushStatus: $pushMsg`nArchivos modificados:`n$files"
        Add-Content -Path $actual -Value "`n$entry2"
        git add $actual | Out-Null
        git commit -m "Actualizar actualizaciones_del_repositorio: registro automático" | Out-Null
        # Intentar pushear el registro de actualizaciones
        if ($pushMsg -eq 'OK') {
            git push origin HEAD | Out-Null
        } else {
            if ($ForcePush -and (Request-ForcePushConfirmation)) {
                git push --force origin HEAD | Out-Null
            } else {
                git push origin HEAD | Out-Null
            }
            if ($LASTEXITCODE -ne 0) {
                Write-Host "Aviso: actualizaciones_del_repositorio actualizado localmente pero no fue posible pushearlo al remoto." -ForegroundColor Yellow
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
