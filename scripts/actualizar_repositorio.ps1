param(
    [string]$Message = "Actualizacion automatica desde script: anadir/actualizar archivos",
    [string]$RepoUrl = "",
    [switch]$SkipChangeLog,
    [switch]$ForcePush
)

$script:TranscriptStarted = $false
$script:ForceAlreadyConfirmed = $false

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

function Close-TranscriptSafe {
    if ($script:TranscriptStarted) {
        try {
            Stop-Transcript -ErrorAction SilentlyContinue | Out-Null
        } catch {
            # no-op
        }
    }
}

function Exit-WithCode {
    param([int]$Code)
    Close-TranscriptSafe
    exit $Code
}

function Request-ForcePushConfirmation {
    if ($script:ForceAlreadyConfirmed) {
        return $true
    }
    $answer = Read-Host "Confirmacion requerida: escribe SI para ejecutar push forzado"
    if ($answer -ceq 'SI') {
        $script:ForceAlreadyConfirmed = $true
        return $true
    }
    return $false
}

function Test-PushSuccess {
    param([pscustomobject]$PushResult)
    if ($null -eq $PushResult) {
        return $false
    }
    return ($PushResult.Code -eq 0 -or ($PushResult.Output -match 'Everything up-?to-?date'))
}

function Invoke-PushOrigin {
    param(
        [switch]$UseForce,
        [switch]$SetUpstream,
        [string]$Context = "push"
    )

    $args = @('push')
    if ($UseForce) {
        $args += '--force'
    }
    if ($SetUpstream) {
        $args += '-u'
    }
    $args += @('origin', 'HEAD')

    Write-Info ("[$Context] Ejecutando: git {0}" -f ($args -join ' '))
    # Convertir a cadena para evitar NativeCommandError en PowerShell al leer stderr
    $output = (& git @args 2>&1) | ForEach-Object { $_.ToString() }
    $code = $LASTEXITCODE

    return [pscustomobject]@{
        Code   = $code
        Output = ($output -join "`n")
    }
}

function Push-WithPolicy {
    param(
        [switch]$AllowForce,
        [switch]$SetUpstream,
        [string]$Context = "push"
    )

    $firstTry = Invoke-PushOrigin -SetUpstream:$SetUpstream -Context $Context
    if (Test-PushSuccess -PushResult $firstTry) {
        return [pscustomobject]@{
            Ok     = $true
            Mode   = 'normal'
            Result = $firstTry
        }
    }

    Write-WarnMsg ("[$Context] Push normal fallo (codigo {0})." -f $firstTry.Code)

    if (-not $AllowForce) {
        return [pscustomobject]@{
            Ok     = $false
            Mode   = 'normal'
            Result = $firstTry
        }
    }

    if (-not (Request-ForcePushConfirmation)) {
        return [pscustomobject]@{
            Ok     = $false
            Mode   = 'cancelado_por_usuario'
            Result = $firstTry
        }
    }

    Write-WarnMsg "[$Context] Reintentando con --force por confirmacion explicita..."
    $forceTry = Invoke-PushOrigin -UseForce -SetUpstream:$SetUpstream -Context "$Context(force)"

    if (Test-PushSuccess -PushResult $forceTry) {
        return [pscustomobject]@{
            Ok     = $true
            Mode   = 'force'
            Result = $forceTry
        }
    }

    return [pscustomobject]@{
        Ok     = $false
        Mode   = 'force'
        Result = $forceTry
    }
}

Write-Step "1/8 Verificando herramientas y repositorio"
if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
    Write-ErrMsg "Git no esta disponible en PATH. Instala Git o ejecuta manualmente los comandos de git."
    Exit-WithCode 1
}

$root = Resolve-Path "$PSScriptRoot\.."
Set-Location $root

$insideRepo = ((& git rev-parse --is-inside-work-tree 2>$null) | Out-String).Trim()
if ($LASTEXITCODE -ne 0 -or $insideRepo -ne 'true') {
    Write-ErrMsg "La ruta $root no es un repositorio Git valido."
    Exit-WithCode 1
}

$branch = ((& git rev-parse --abbrev-ref HEAD 2>$null) | Out-String).Trim()
Write-Info "Repositorio: $root"
Write-Info "Rama actual: $branch"

Write-Step "2/8 Iniciando bitacora de ejecucion"
$logsDir = Join-Path $root 'scripts\logs'
New-Item -ItemType Directory -Force -Path $logsDir | Out-Null
$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$logFile = Join-Path $logsDir ("actualizar_repositorio-$timestamp.log")
try {
    Start-Transcript -Path $logFile -Append -ErrorAction Stop | Out-Null
    $script:TranscriptStarted = $true
    Write-Info "Log de ejecucion: $logFile"
} catch {
    Write-WarnMsg "No fue posible iniciar Start-Transcript: $_"
}

Write-Step "3/8 Revisando .gitignore"
$gitignorePath = Join-Path $root '.gitignore'
if (-not (Test-Path $gitignorePath)) {
    Write-WarnMsg "No existe .gitignore. Se creara una version basica para evitar subir artefactos locales."
    @(
        "# Binarios y logs de backend"
        "backend/server.exe"
        "backend/*.log"
        "backend/*.err"
        "backend/*.exe~"
        "backend/pos.db"
        "*.db"
        "# VSCode"
        ".vscode/"
        "# Archivos compilados"
        "bin/"
        "obj/"
    ) | Out-File -FilePath $gitignorePath -Encoding UTF8
    git add .gitignore | Out-Null
    Write-Ok ".gitignore basico creado y agregado al staging."
} else {
    Write-Info ".gitignore ya existe."
}

Write-Step "4/8 Preparando cambios para commit"
git add -A
$statusLines = @(git status --porcelain)
if ($statusLines.Count -eq 0) {
    Write-WarnMsg "No hay cambios para commitear. No se realizo push."
    Exit-WithCode 0
}
Write-Info ("Archivos detectados para commit: {0}" -f $statusLines.Count)

Write-Step "5/8 Creando commit principal"
$commitOutput = & git commit -m $Message 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-ErrMsg "El commit fallo. Revisa git status y corrige conflictos o hooks."
    if ($commitOutput) {
        Write-Host $commitOutput
    }
    Exit-WithCode $LASTEXITCODE
}

$mainCommit = ((& git rev-parse --short HEAD 2>$null) | Out-String).Trim()
Write-Ok ("Commit principal creado: {0}" -f $mainCommit)

$mainFiles = @(
    & git show --name-only --pretty="" HEAD |
        Where-Object { -not [string]::IsNullOrWhiteSpace($_) }
)
$filesAsBullets = if ($mainFiles.Count -gt 0) {
    ($mainFiles | ForEach-Object { "- $_" }) -join "`n"
} else {
    "- (sin archivos detectados)"
}

Write-Step "6/8 Subiendo commit principal a GitHub"
$originUrl = ((& git remote get-url origin 2>$null) | Out-String).Trim()
if ([string]::IsNullOrWhiteSpace($originUrl)) {
    if ([string]::IsNullOrWhiteSpace($RepoUrl)) {
        $RepoUrl = $env:REPO_URL
    }
    if ([string]::IsNullOrWhiteSpace($RepoUrl)) {
        Write-ErrMsg "No existe remoto origin y no se proporciono -RepoUrl ni REPO_URL."
        Exit-WithCode 1
    }

    Write-Info "Agregando remoto origin desde parametro/entorno..."
    git remote add origin $RepoUrl
    if ($LASTEXITCODE -ne 0) {
        Write-ErrMsg "No se pudo agregar el remoto origin. Revisa URL o permisos."
        Exit-WithCode $LASTEXITCODE
    }
    $originUrl = $RepoUrl
}
Write-Info "Remoto origin: $originUrl"

$mainPush = Push-WithPolicy -AllowForce:$ForcePush -SetUpstream -Context "commit principal"
if (-not $mainPush.Ok) {
    Write-ErrMsg "No se pudo subir el commit principal a GitHub."
    if ($mainPush.Result.Output) {
        Write-Host $mainPush.Result.Output
    }
    Exit-WithCode 1
}
Write-Ok ("Push principal completado (modo: {0})." -f $mainPush.Mode)

$docsPushOk = $true
$docsCommit = ""

if (-not $SkipChangeLog) {
    Write-Step "7/8 Actualizando bitacoras del repositorio"
    $filesToAdd = @()
    $hist = Join-Path $root "documentos\historial_de_cambios"
    if (Test-Path $hist) {
        $entry = "$(Get-Date -Format yyyy-MM-dd): $Message`nPushStatus: OK`nArchivos modificados:`n$filesAsBullets"
        Add-Content -Path $hist -Value "`n$entry"
        $filesToAdd += $hist
        Write-Info "Se actualizo documentos/historial_de_cambios"
    } else {
        Write-WarnMsg "No se encontro documentos/historial_de_cambios para actualizar."
    }

    $actual = Join-Path $root "documentos\actualizaciones_del_repositorio.md"
    if (-not (Test-Path $actual)) {
        @(
            "# Actualizaciones del repositorio",
            "",
            "Este documento registra las actualizaciones automaticas realizadas por el script scripts/actualizar_repositorio.ps1.",
            ""
        ) | Out-File -FilePath $actual -Encoding UTF8
        Write-Info "Se creo documentos/actualizaciones_del_repositorio.md"
    }

    $entry2 = "$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss') - Mensaje: $Message; Commit: $mainCommit; PushStatus: OK`nArchivos modificados:`n$filesAsBullets"
    Add-Content -Path $actual -Value "`n$entry2"
    $filesToAdd += $actual
    Write-Info "Se actualizo documentos/actualizaciones_del_repositorio.md"

    if ($filesToAdd.Count -gt 0) {
        git add @filesToAdd | Out-Null
        $docsStatus = @(git status --porcelain)
        if ($docsStatus.Count -gt 0) {
            $docsCommitOutput = & git commit -m "Actualizar registros automaticos de repositorio" 2>&1
            if ($LASTEXITCODE -ne 0) {
                Write-ErrMsg "No se pudo crear el commit de bitacoras automaticas."
                if ($docsCommitOutput) {
                    Write-Host $docsCommitOutput
                }
                $docsPushOk = $false
            } else {
                $docsCommit = ((& git rev-parse --short HEAD 2>$null) | Out-String).Trim()
                Write-Ok ("Commit de bitacoras creado: {0}" -f $docsCommit)

                $docsPush = Push-WithPolicy -AllowForce:$ForcePush -Context "bitacoras automaticas"
                if (-not $docsPush.Ok) {
                    Write-ErrMsg "No se pudo subir el commit de bitacoras automaticas."
                    if ($docsPush.Result.Output) {
                        Write-Host $docsPush.Result.Output
                    }
                    $docsPushOk = $false
                } else {
                    Write-Ok ("Push de bitacoras completado (modo: {0})." -f $docsPush.Mode)
                }
            }
        } else {
            Write-Info "No hubo cambios adicionales en bitacoras para commitear."
        }
    }
} else {
    Write-Step "7/8 Omitiendo bitacoras por parametro -SkipChangeLog"
}

Write-Step "8/8 Resumen final"
Write-Ok ("Commit principal en remoto: {0}" -f $mainCommit)
if (-not [string]::IsNullOrWhiteSpace($docsCommit)) {
    Write-Ok ("Commit de bitacoras en remoto: {0}" -f $docsCommit)
}
Write-Info ("Mensaje usado: {0}" -f $Message)
Write-Info ("Rama: {0}" -f $branch)
Write-Info ("Remoto: {0}" -f $originUrl)

if ($docsPushOk) {
    Write-Ok "Operacion completada con exito."
    Exit-WithCode 0
}

Write-ErrMsg "El commit principal se subio, pero hubo fallos al subir las bitacoras automaticas."
Exit-WithCode 1
