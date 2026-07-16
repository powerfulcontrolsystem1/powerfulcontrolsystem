# Sube cambios locales al remoto "origin".
# URL del repositorio (en orden de prioridad): -RepoUrl, $script:PcsGitRemoteUrl en scripts/pcs_deployment.local.ps1,
# variable PCS_REPO_URL o REPO_URL, archivo local scripts/actualizar_repositorio.repo_url
# (copia desde .repo_url.example; no se versiona).
# Flujo conjunto con VPS: scripts/publicar_git_y_vps.ps1
# Si ya tienes origin pero apunta a un repo antiguo y configuraste la URL nueva, usa -SetOrigin
# para ejecutar: git remote set-url origin <url>
# Config unificada (opcional): scripts/pcs_deployment.local.ps1 (plantilla: pcs_deployment.local.ps1.example)
# El dot-source debe ir *despues* de param (param ha de ser el primer bloque ejecutable).
param(
    [string]$Message = "Actualizacion automatica desde script: anadir/actualizar archivos",
    [string]$RepoUrl = "",
    [switch]$SkipChangeLog,
    [switch]$ForcePush,
    [switch]$SetOrigin,
    [bool]$CreateProtectedMainPR = $true,
    [int]$ProtectedMainPRWaitSeconds = 900,
    [switch]$NoAutoMergeProtectedPR
)

$__arDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$__arCfg = Join-Path $__arDir "pcs_deployment.local.ps1"
if (Test-Path -LiteralPath $__arCfg) {
    . $__arCfg
}

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

function Invoke-GitAddQuietLineEndingAdvice {
    param(
        [Parameter(Mandatory = $true)]
        [string[]]$Paths
    )

    $previousErrorActionPreference = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        $raw = @(& git add @Paths 2>&1 | ForEach-Object { $_.ToString() })
    } finally {
        $ErrorActionPreference = $previousErrorActionPreference
    }
    if ($LASTEXITCODE -ne 0) {
        if ($raw.Count -gt 0) {
            $raw | ForEach-Object { Write-Host $_ }
        }
        return $false
    }

    foreach ($line in $raw) {
        if ([string]::IsNullOrWhiteSpace($line)) {
            continue
        }
        if ($line -match 'LF will be replaced by CRLF') {
            continue
        }
        if ($line -match 'CRLF will be replaced by LF') {
            continue
        }
        Write-Host $line
    }
    return $true
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
    $previousErrorActionPreference = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        $output = (& git @args 2>&1) | ForEach-Object { $_.ToString() }
        $code = $LASTEXITCODE
    } finally {
        $ErrorActionPreference = $previousErrorActionPreference
    }

    return [pscustomobject]@{
        Code   = $code
        Output = ($output -join "`n")
    }
}

function Get-TargetRepoUrl {
    param([string]$ExplicitRepoUrl)
    if (-not [string]::IsNullOrWhiteSpace($ExplicitRepoUrl)) {
        return $ExplicitRepoUrl.Trim()
    }
    if (Get-Variable -Name PcsGitRemoteUrl -Scope Script -ErrorAction SilentlyContinue) {
        if (-not [string]::IsNullOrWhiteSpace($script:PcsGitRemoteUrl)) {
            return $script:PcsGitRemoteUrl.Trim()
        }
    }
    if (-not [string]::IsNullOrWhiteSpace($env:PCS_REPO_URL)) {
        return $env:PCS_REPO_URL.Trim()
    }
    if (-not [string]::IsNullOrWhiteSpace($env:REPO_URL)) {
        return $env:REPO_URL.Trim()
    }
    $path = Join-Path $PSScriptRoot 'actualizar_repositorio.repo_url'
    if (Test-Path -LiteralPath $path) {
        foreach ($line in Get-Content -LiteralPath $path) {
            $t = $line.Trim()
            if ([string]::IsNullOrWhiteSpace($t) -or $t.StartsWith('#')) {
                continue
            }
            return $t
        }
    }
    return ""
}

function Normalize-RepoUrlForCompare {
    param([string]$Url)
    if ([string]::IsNullOrWhiteSpace($Url)) {
        return ""
    }
    return $Url.Trim().TrimEnd('/').ToLowerInvariant()
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

function Test-ProtectedBranchRejection {
    param([string]$Output)

    return $Output -match 'Protected branch update failed|Changes must be made through a pull request|protected branch hook declined'
}

function Enable-RepositoryAutoMergeIfAllowed {
    $repo = (& gh repo view --json nameWithOwner --jq '.nameWithOwner' 2>$null | Select-Object -Last 1).ToString().Trim()
    if ([string]::IsNullOrWhiteSpace($repo) -or $LASTEXITCODE -ne 0) {
        Write-WarnMsg "No fue posible identificar el repositorio para habilitar auto-merge."
        return $false
    }
    & gh api --method PATCH "repos/$repo" -f allow_auto_merge=true *> $null
    if ($LASTEXITCODE -ne 0) {
        Write-WarnMsg "GitHub no permitio habilitar auto-merge automaticamente para este repositorio."
        return $false
    }
    Write-Ok "Auto-merge habilitado en GitHub sin modificar las reglas de aprobacion ni checks."
    return $true
}

function Sync-LocalBaseAfterProtectedMerge {
    param(
        [Parameter(Mandatory = $true)][string]$BaseBranch
    )

    & git fetch origin $BaseBranch
    if ($LASTEXITCODE -ne 0) {
        throw "La PR fue fusionada, pero no se pudo actualizar origin/$BaseBranch."
    }

    & git pull --ff-only origin $BaseBranch
    if ($LASTEXITCODE -eq 0) {
        Write-Ok "$BaseBranch local actualizado por fast-forward."
        return
    }

    # GitHub puede fusionar una PR protegida con squash. En ese caso el commit
    # local publicado y el commit de main tienen el mismo cambio, pero distinta
    # identidad, por lo que fast-forward falla. Solo con arbol limpio se intenta
    # un rebase no destructivo; luego se exige igualdad exacta con origin/main
    # para impedir que rs despliegue commits locales no publicados.
    $status = (& git status --porcelain 2>$null | Out-String).Trim()
    if (-not [string]::IsNullOrWhiteSpace($status)) {
        throw "La PR fue fusionada, pero $BaseBranch local divergió y el arbol no esta limpio. Revisa los cambios antes de reconciliarlo."
    }
    Write-WarnMsg "Fast-forward no disponible despues de la fusion protegida; reconciliando $BaseBranch mediante rebase seguro."
    & git rebase "origin/$BaseBranch"
    if ($LASTEXITCODE -ne 0) {
        & git rebase --abort *> $null
        throw "La PR fue fusionada, pero el rebase de $BaseBranch local tuvo conflictos. Se aborto sin modificar el arbol."
    }

    $localRevision = ((& git rev-parse HEAD 2>$null) | Out-String).Trim()
    $remoteRevision = ((& git rev-parse "origin/$BaseBranch" 2>$null) | Out-String).Trim()
    if ([string]::IsNullOrWhiteSpace($localRevision) -or $localRevision -ne $remoteRevision) {
        throw "$BaseBranch local se reconcilio, pero contiene commits no publicados. rs no desplegara una revision distinta de origin/$BaseBranch."
    }
    Write-Ok "$BaseBranch local reconciliado con origin/$BaseBranch despues de la fusion protegida."
}

function Invoke-ProtectedMainPullRequest {
    param(
        [Parameter(Mandatory = $true)][string]$BaseBranch,
        [Parameter(Mandatory = $true)][string]$CommitMessage,
        [int]$WaitSeconds,
        [switch]$DisableAutoMerge
    )

    if (-not (Get-Command gh -ErrorAction SilentlyContinue)) {
        throw "GitHub CLI no esta disponible para crear la PR requerida por la rama protegida."
    }
    & gh auth status *> $null
    if ($LASTEXITCODE -ne 0) {
        throw "GitHub CLI no tiene una sesion valida para crear la PR requerida por la rama protegida."
    }

    $stamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $prBranch = "codex/rs-$stamp"
    Write-Info "La rama $BaseBranch esta protegida. Creando rama de publicacion $prBranch."
    & git switch -c $prBranch
    if ($LASTEXITCODE -ne 0) {
        throw "No se pudo crear la rama temporal de publicacion $prBranch."
    }
    & git push -u origin $prBranch
    if ($LASTEXITCODE -ne 0) {
        throw "No se pudo publicar la rama temporal $prBranch."
    }

    $body = @(
        "## Publicacion automatizada por rs",
        "",
        "Esta PR fue creada porque GitHub protege `'$BaseBranch'`. No omite revisiones ni checks obligatorios.",
        "",
        "- Commit local: ``$(git rev-parse --short HEAD)``",
        "- Auto-merge solicitado: ``$(-not $DisableAutoMerge)``",
        "- No se sincroniza la VPS hasta que GitHub confirme la fusion."
    ) -join "`n"
    $prUrl = (& gh pr create --base $BaseBranch --head $prBranch --title $CommitMessage --body $body 2>&1 | Select-Object -Last 1).ToString().Trim()
    if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($prUrl)) {
        throw "No se pudo crear la PR de publicacion para la rama protegida."
    }
    Write-Info "PR creada: $prUrl"

    if (-not $DisableAutoMerge) {
        $mergeOutput = @(& gh pr merge $prUrl --auto --squash --delete-branch 2>&1 | ForEach-Object { $_.ToString() })
        if ($LASTEXITCODE -ne 0) {
            if (($mergeOutput -join "`n") -match 'Auto merge is not allowed' -and (Enable-RepositoryAutoMergeIfAllowed)) {
                $mergeOutput = @(& gh pr merge $prUrl --auto --squash --delete-branch 2>&1 | ForEach-Object { $_.ToString() })
            }
            if ($LASTEXITCODE -ne 0) {
                $mergeOutput | ForEach-Object { Write-Info $_ }
                Write-WarnMsg "No fue posible activar auto-merge. La PR conserva las protecciones de GitHub."
            } else {
                Write-Ok "Auto-merge solicitado: GitHub fusionara solo despues de aprobacion independiente y checks verdes."
            }
        } else {
            Write-Ok "Auto-merge solicitado: GitHub fusionara solo despues de aprobacion independiente y checks verdes."
        }
    }

    if ($WaitSeconds -le 0) {
        return [pscustomobject]@{ Merged = $false; Pending = $true; Url = $prUrl }
    }

    $deadline = (Get-Date).AddSeconds($WaitSeconds)
    while ((Get-Date) -lt $deadline) {
        $stateJson = & gh pr view $prUrl --json state,mergedAt --jq '{state:.state,mergedAt:.mergedAt}' 2>$null
        if ($LASTEXITCODE -eq 0 -and -not [string]::IsNullOrWhiteSpace($stateJson)) {
            $state = $stateJson | ConvertFrom-Json
            if ($state.state -eq 'MERGED') {
                & git switch $BaseBranch
                if ($LASTEXITCODE -ne 0) {
                    throw "La PR fue fusionada, pero no se pudo volver a $BaseBranch."
                }
                Sync-LocalBaseAfterProtectedMerge -BaseBranch $BaseBranch
                Write-Ok "PR fusionada por GitHub. $BaseBranch quedo actualizado para continuar con rs."
                return [pscustomobject]@{ Merged = $true; Pending = $false; Url = $prUrl }
            }
        }
        Start-Sleep -Seconds 10
    }

    return [pscustomobject]@{ Merged = $false; Pending = $true; Url = $prUrl }
}

function Resolve-ProtectedMainPush {
    param(
        [string]$Branch,
        [pscustomobject]$PushResult,
        [string]$CommitMessage
    )

    if ($Branch -ne 'main' -or -not $CreateProtectedMainPR -or -not (Test-ProtectedBranchRejection -Output $PushResult.Result.Output)) {
        return $null
    }
    $flow = Invoke-ProtectedMainPullRequest -BaseBranch 'main' -CommitMessage $CommitMessage -WaitSeconds $ProtectedMainPRWaitSeconds -DisableAutoMerge:$NoAutoMergeProtectedPR
    if ($flow.Merged) {
        return $flow
    }
    Write-WarnMsg "La PR requiere aprobacion independiente o checks pendientes: $($flow.Url)"
    return $flow
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
        "*.db"
        "# VSCode"
        ".vscode/"
        "# Archivos compilados"
        "bin/"
        "obj/"
    ) | Out-File -FilePath $gitignorePath -Encoding UTF8
    if (-not (Invoke-GitAddQuietLineEndingAdvice -Paths @('.gitignore'))) {
        Write-ErrMsg "git add .gitignore fallo."
        Exit-WithCode 1
    }
    Write-Ok ".gitignore basico creado y agregado al staging."
} else {
    Write-Info ".gitignore ya existe."
}

Write-Step "4/8 Preparando cambios para commit"
if (-not (Invoke-GitAddQuietLineEndingAdvice -Paths @('-A'))) {
    Write-ErrMsg "git add -A fallo."
    Exit-WithCode 1
}
$statusLines = @(git status --porcelain)
if ($statusLines.Count -eq 0) {
    # Un arbol limpio no implica que la rama ya este publicada. Esto sucede
    # cuando Codex u otra herramienta crea commits locales antes de ejecutar
    # rs; dejar de inmediato evitaba que sync_to_vps desplegara una rama con
    # upstream verificable.
    $originUrl = ((& git remote get-url origin 2>$null) | Out-String).Trim()
    if ([string]::IsNullOrWhiteSpace($originUrl)) {
        Write-ErrMsg "No existe remoto 'origin' para publicar la rama limpia."
        Exit-WithCode 1
    }
    $cleanPush = Push-WithPolicy -AllowForce:$ForcePush -SetUpstream -Context "rama limpia"
    if (-not $cleanPush.Ok) {
        $protectedFlow = Resolve-ProtectedMainPush -Branch $branch -PushResult $cleanPush -CommitMessage $Message
        if ($null -ne $protectedFlow) {
            if ($protectedFlow.Merged) {
                Write-Ok "Publicacion por PR protegida completada."
                Exit-WithCode 0
            }
            Write-WarnMsg "No se sincroniza la VPS hasta que GitHub fusione la PR protegida."
            Exit-WithCode 2
        }
        Write-ErrMsg "No se pudo verificar o publicar la rama limpia en origin."
        if ($cleanPush.Result.Output) {
            Write-Host $cleanPush.Result.Output
        }
        Exit-WithCode 1
    }
    Write-Ok ("Rama limpia verificada/publicada en origin (modo: {0})." -f $cleanPush.Mode)
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

Write-Step "6/8 Alineando remoto origin y subiendo commit principal"
$targetUrl = Get-TargetRepoUrl -ExplicitRepoUrl $RepoUrl
$originUrl = ((& git remote get-url origin 2>$null) | Out-String).Trim()

if ([string]::IsNullOrWhiteSpace($originUrl)) {
    if ([string]::IsNullOrWhiteSpace($targetUrl)) {
        Write-ErrMsg "No existe remoto 'origin'. Indica la URL del repositorio nuevo: parametro -RepoUrl, variable de entorno PCS_REPO_URL o REPO_URL, o crea scripts/actualizar_repositorio.repo_url (plantilla: scripts/actualizar_repositorio.repo_url.example)."
        Exit-WithCode 1
    }
    Write-Info "Agregando remoto origin..."
    & git remote add origin $targetUrl
    if ($LASTEXITCODE -ne 0) {
        Write-ErrMsg "No se pudo agregar el remoto origin. Revisa URL o permisos."
        Exit-WithCode $LASTEXITCODE
    }
    $originUrl = ((& git remote get-url origin 2>$null) | Out-String).Trim()
} elseif (-not [string]::IsNullOrWhiteSpace($targetUrl)) {
    $a = Normalize-RepoUrlForCompare -Url $originUrl
    $b = Normalize-RepoUrlForCompare -Url $targetUrl
    if ($a -ne $b) {
        if (-not $SetOrigin) {
            Write-ErrMsg "El remoto 'origin' apunta a:`n  $originUrl`n`nLa URL configurada para este clone es:`n  $targetUrl`n`nPara cambiar al repositorio nuevo, ejecuta el script con -SetOrigin`n(o manualmente: git remote set-url origin <url>)."
            Exit-WithCode 1
        }
        Write-Info "Ajustando origin a la URL configurada (-SetOrigin)..."
        & git remote set-url origin $targetUrl
        if ($LASTEXITCODE -ne 0) {
            Write-ErrMsg "git remote set-url fallo. Revisa la URL y permisos."
            Exit-WithCode $LASTEXITCODE
        }
        $originUrl = ((& git remote get-url origin 2>$null) | Out-String).Trim()
    }
}
Write-Info "Remoto origin: $originUrl"

$mainPush = Push-WithPolicy -AllowForce:$ForcePush -SetUpstream -Context "commit principal"
if (-not $mainPush.Ok) {
    $protectedFlow = Resolve-ProtectedMainPush -Branch $branch -PushResult $mainPush -CommitMessage $Message
    if ($null -ne $protectedFlow) {
        if ($protectedFlow.Merged) {
            Write-Ok "Publicacion por PR protegida completada."
            Exit-WithCode 0
        }
        Write-WarnMsg "No se sincroniza la VPS hasta que GitHub fusione la PR protegida."
        Exit-WithCode 2
    }
    Write-ErrMsg "No se pudo subir el commit principal al remoto."
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
        if (-not (Invoke-GitAddQuietLineEndingAdvice -Paths $filesToAdd)) {
            Write-ErrMsg "git add (bitacoras) fallo."
            Exit-WithCode 1
        }
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
