param(
    [string]$Message = "Actualización automática desde script: añadir/actualizar archivos",
    [string]$RepoUrl = "",
    [switch]$SkipChangeLog
)

Write-Host "Verificando estado de git..."
if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
    Write-Host "ERROR: Git no está disponible en PATH. Instala Git o ejecuta manualmente los comandos de git." -ForegroundColor Red
    exit 1
}

$root = Resolve-Path "$PSScriptRoot\.."
Set-Location $root

Write-Host "Preparando cambios para commit..."
git add -A

$status = git status --porcelain
if ([string]::IsNullOrEmpty($status)) {
    Write-Host "No hay cambios para commitear." -ForegroundColor Yellow
    exit 0
}

Write-Host "Creando commit: $Message"
git commit -m $Message
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: El commit falló. Revisa 'git status'." -ForegroundColor Red
    exit $LASTEXITCODE
}

function Try-PushOrigin {
    git push origin HEAD 2>&1
    return $LASTEXITCODE
}

Write-Host "Intentando push al remoto 'origin'..."
$pushCode = Try-PushOrigin
if ($pushCode -eq 0) {
    $pushMsg = 'OK'
    Write-Host "Éxito: cambios empujados al remoto 'origin'." -ForegroundColor Green
} else {
    Write-Warning "Push inicial falló (código $pushCode). Intentando resolver remoto 'origin'..."
    $originUrl = git remote get-url origin 2>$null
    if (-not $originUrl) {
        if (-not [string]::IsNullOrWhiteSpace($RepoUrl)) {
            Write-Host "Agregando remoto 'origin' desde parámetro: $RepoUrl"
            git remote add origin $RepoUrl
            if ($LASTEXITCODE -ne 0) {
                Write-Host "ERROR: No se pudo agregar remoto 'origin'." -ForegroundColor Red
                $pushMsg = 'FAIL_ADD_REMOTE'
            } else {
                Write-Host "Remoto 'origin' agregado. Reintentando push..."
                git push -u origin HEAD
                if ($LASTEXITCODE -eq 0) {
                    $pushMsg = 'OK'
                    Write-Host "Éxito: cambios empujados al remoto 'origin'." -ForegroundColor Green
                } else {
                    $pushMsg = 'FAIL_PUSH'
                    Write-Host "ERROR: El push falló incluso después de añadir remoto. Código: $LASTEXITCODE" -ForegroundColor Red
                }
            }
        } else {
            Write-Host "ERROR: No existe el remoto 'origin' y no se proporcionó -RepoUrl. No puedo hacer push." -ForegroundColor Red
            $pushMsg = 'NO_ORIGIN'
        }
    } else {
        Write-Host "Remoto 'origin' existe ($originUrl) pero el push falló. Código: $pushCode" -ForegroundColor Red
        $pushMsg = 'FAIL_PUSH'
    }
}

if (-not $SkipChangeLog) {
    $hist = Join-Path $root "documentos\historial_de_cambios"
    if (Test-Path $hist) {
        $files = git show --name-only --pretty="" HEAD | Where-Object { $_ -ne '' } | ForEach-Object { "- $_" } | Out-String
        $entry = "$(Get-Date -Format yyyy-MM-dd): $Message`nPushStatus: $pushMsg`nArchivos modificados:`n$files"
        Add-Content -Path $hist -Value "`n$entry"
        git add $hist
        git commit -m "Actualizar historial_de_cambios: registro automático" | Out-Null
        if ($pushMsg -eq 'OK') {
            git push origin HEAD | Out-Null
        } else {
            Write-Host "Aviso: historial actualizado localmente pero no fue posible pushearlo al remoto." -ForegroundColor Yellow
        }
    } else {
        Write-Host "Aviso: no se encontró 'documentos/historial_de_cambios' para actualizar." -ForegroundColor Yellow
    }
}

if ($pushMsg -eq 'OK') {
    Write-Host "Operación completada con éxito." -ForegroundColor Green
    exit 0
} else {
    switch ($pushMsg) {
        'NO_ORIGIN' { Write-Host "Falló: no existe el remoto 'origin'. Vuelve a ejecutar el script con -RepoUrl <url> para agregarlo automáticamente." -ForegroundColor Red }
        'FAIL_ADD_REMOTE' { Write-Host "Falló: no se pudo agregar el remoto 'origin'. Revisa la URL o permisos." -ForegroundColor Red }
        default { Write-Host "Falló: no se pudo pushear los cambios al remoto. Revisa la configuración de Git y permisos." -ForegroundColor Red }
    }
    exit 1
}
