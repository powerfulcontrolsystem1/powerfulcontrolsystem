param(
    [switch]$SkipFullSuite
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$backendDir = Join-Path $repoRoot "backend"
$logsDir = Join-Path $repoRoot "scripts\logs"
$reportPath = Join-Path $repoRoot "documentos\punto_13_validacion_integral_resultado.md"

New-Item -ItemType Directory -Path $logsDir -Force | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$logPath = Join-Path $logsDir ("punto13-validacion-" + $timestamp + ".log")

$script:results = @()
$overallOk = $true
$fatalMessage = ""

function Add-Log {
    param([string]$Text)
    Add-Content -Path $logPath -Value $Text
}

function Run-GoCommand {
    param(
        [string]$Name,
        [string[]]$CmdArgs,
        [switch]$AllowFailure
    )

    $commandText = "go " + ($CmdArgs -join " ")
    Write-Host ">> $commandText" -ForegroundColor Cyan

    $output = & go @CmdArgs 2>&1 | Out-String
    $exitCode = $LASTEXITCODE
    $status = if ($exitCode -eq 0) { "ok" } else { "fail" }

    Add-Log "### $Name"
    Add-Log ("PS> " + $commandText)
    Add-Log $output
    Add-Log ("ExitCode: " + $exitCode)
    Add-Log ""

    $script:results += [PSCustomObject]@{
        Name = $Name
        Command = $commandText
        Status = $status
        ExitCode = $exitCode
    }

    if (-not $AllowFailure -and $exitCode -ne 0) {
        throw "Fallo en comando obligatorio: $commandText"
    }
}

function Build-Report {
    param(
        [bool]$Ok,
        [string]$FailureMessage
    )

    $lines = @()
    $lines += "# Punto 13 - Validacion integral (ultima ejecucion)"
    $lines += ""
    $lines += "Fecha: " + (Get-Date -Format "yyyy-MM-dd HH:mm:ss")
    $lines += "Log tecnico: " + ("scripts/logs/" + [IO.Path]::GetFileName($logPath))
    $lines += ""
    $lines += "## Resultado de comandos"
    $lines += ""
    $lines += "| Paso | Comando | Estado | Exit code |"
    $lines += "|---|---|---|---|"

    foreach ($r in $script:results) {
        $lines += ('| {0} | `{1}` | {2} | {3} |' -f $r.Name, $r.Command, $r.Status, $r.ExitCode)
    }

    $lines += ""
    $lines += "## Estado final"
    $lines += ""

    if ($Ok) {
        $lines += "- Gate tecnico: aprobado"
        $lines += "- Observacion: ejecutar y registrar UAT manual antes de despliegue productivo"
    } else {
        $lines += "- Gate tecnico: rechazado"
        if (-not [string]::IsNullOrWhiteSpace($FailureMessage)) {
            $lines += "- Motivo: " + $FailureMessage
        }
    }

    $lines += ""
    $lines += "## Pendientes manuales"
    $lines += ""
    $lines += "- Ejecutar smoke/UAT en modulos criticos (auth, clientes, inventario, compras, facturacion, finanzas y auditoria)."
    $lines += "- Validar checklist de rollback antes de salida controlada."

    Set-Content -Path $reportPath -Value ($lines -join "`r`n") -Encoding UTF8
}

try {
    Add-Log "# Punto 13 - bitacora de validacion"
    Add-Log ("Inicio: " + (Get-Date -Format "yyyy-MM-dd HH:mm:ss"))
    Add-Log ""

    Push-Location $backendDir

    Run-GoCommand -Name "Suite productiva" -CmdArgs @("test", "./auth", "./db", "./handlers", "./metrics", "./utils", "-count=1")

    if (-not $SkipFullSuite) {
        Run-GoCommand -Name "Suite completa backend" -CmdArgs @("test", "./...", "-count=1")
    } else {
        $script:results += [PSCustomObject]@{
            Name = "Suite completa backend"
            Command = "go test ./... -count=1"
            Status = "skipped"
            ExitCode = 0
        }
    }
} catch {
    $overallOk = $false
    if ($_.Exception -and $_.Exception.Message) {
        $fatalMessage = $_.Exception.Message
    } else {
        $fatalMessage = "Fallo desconocido durante la validacion"
    }
} finally {
    Pop-Location
}

Build-Report -Ok:$overallOk -FailureMessage:$fatalMessage

if ($overallOk) {
    Write-Host "Validacion punto 13 completada: OK" -ForegroundColor Green
    Write-Host ("Reporte: " + $reportPath)
    exit 0
}

Write-Host "Validacion punto 13 finalizo con errores" -ForegroundColor Red
Write-Host ("Detalle: " + $fatalMessage) -ForegroundColor Yellow
Write-Host ("Reporte: " + $reportPath)
exit 1
