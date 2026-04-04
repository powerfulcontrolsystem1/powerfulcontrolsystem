param(
    [string]$Quarter = "",
    [int]$Year = 0
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$docsDir = Join-Path $repoRoot "documentos"
$logsDir = Join-Path $repoRoot "scripts\logs"
$planPath = Join-Path $docsDir "plan_maestro_pos_multiempresa_14_puntos.md"
$point13Path = Join-Path $docsDir "punto_13_validacion_integral_resultado.md"
$roadmapPath = Join-Path $docsDir "roadmap_trimestral_pos_multiempresa.md"
$reportPath = Join-Path $docsDir "punto_14_operacion_continua_reporte.md"

New-Item -ItemType Directory -Path $logsDir -Force | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$logPath = Join-Path $logsDir ("operacion-continua-" + $timestamp + ".log")

function Add-Log {
    param([string]$Text)
    Add-Content -Path $logPath -Value $Text
}

if (-not (Test-Path $planPath)) {
    throw "No se encontro el plan maestro en: $planPath"
}

$now = Get-Date
if ($Year -le 0) {
    $Year = $now.Year
}
if ([string]::IsNullOrWhiteSpace($Quarter)) {
    $quarterNum = [Math]::Ceiling($now.Month / 3)
    $Quarter = "Q" + $quarterNum
}

$planLines = Get-Content -Path $planPath -ErrorAction Stop
$pattern = '^\|\s*(\d+)\s*\|\s*([^|]+)\|\s*([^|]+)\|\s*([^|]+)\|'
$points = @()

foreach ($line in $planLines) {
    if ($line -match $pattern) {
        $pointId = [int]$matches[1]
        if ($pointId -ge 1 -and $pointId -le 15) {
            $points += [PSCustomObject]@{
                Point = $pointId
                Module = $matches[2].Trim()
                Status = $matches[3].Trim().ToLower()
                Deliverable = $matches[4].Trim()
            }
        }
    }
}

$totalPoints = $points.Count
$completedCount = @($points | Where-Object { $_.Status -eq "completado" }).Count
$inProgressCount = @($points | Where-Object { $_.Status -eq "en curso" }).Count
$pendingCount = @($points | Where-Object { $_.Status -eq "pendiente" }).Count

$completedPct = 0.0
if ($totalPoints -gt 0) {
    $completedPct = [Math]::Round(($completedCount * 100.0 / $totalPoints), 2)
}

$pendingNames = @($points |
    Where-Object { $_.Status -eq "pendiente" } |
    Sort-Object Point |
    ForEach-Object { "Punto " + $_.Point + " - " + $_.Module })

$point13Gate = "sin evidencia"
if (Test-Path $point13Path) {
    $point13Text = Get-Content -Path $point13Path -Raw -ErrorAction SilentlyContinue
    if ($point13Text -match "Gate tecnico: aprobado") {
        $point13Gate = "aprobado"
    } elseif ($point13Text -match "Gate tecnico: rechazado") {
        $point13Gate = "rechazado"
    }
}

$roadmapState = if (Test-Path $roadmapPath) { "vigente" } else { "no disponible" }

$summary = "Total puntos=" + $totalPoints + "; completados=" + $completedCount + "; en_curso=" + $inProgressCount + "; pendientes=" + $pendingCount + "; gate_tecnico=" + $point13Gate
Add-Log "# Reporte de operacion continua"
Add-Log ("Fecha: " + (Get-Date -Format "yyyy-MM-dd HH:mm:ss"))
Add-Log $summary

$lines = @()
$lines += "# Punto 14 - Reporte de operacion continua"
$lines += ""
$lines += "Fecha: " + (Get-Date -Format "yyyy-MM-dd HH:mm:ss")
$lines += "Periodo evaluado: " + $Quarter + " " + $Year
$lines += "Bitacora tecnica: scripts/logs/" + [IO.Path]::GetFileName($logPath)
$lines += ""
$lines += "## Resumen ejecutivo"
$lines += ""
$lines += "- Puntos del plan maestro: " + $totalPoints
$lines += "- Puntos completados: " + $completedCount + " (" + $completedPct + "%)"
$lines += "- Puntos en curso: " + $inProgressCount
$lines += "- Puntos pendientes: " + $pendingCount
$lines += "- Gate tecnico de referencia (punto 13): " + $point13Gate
$lines += "- Roadmap trimestral: " + $roadmapState
$lines += ""
$lines += "## KPI de gobierno operativo"
$lines += ""
$lines += "| KPI | Valor actual | Meta inicial | Estado |"
$lines += "|---|---|---|---|"
$lines += "| puntos_completados_pct | " + $completedPct + "% | >= 75% | " + ($(if ($completedPct -ge 75) { "ok" } else { "riesgo" })) + " |"
$lines += "| puntos_pendientes | " + $pendingCount + " | <= 1 | " + ($(if ($pendingCount -le 1) { "ok" } else { "riesgo" })) + " |"
$lines += "| gate_tecnico_vigente | " + $point13Gate + " | aprobado | " + ($(if ($point13Gate -eq "aprobado") { "ok" } else { "riesgo" })) + " |"
$lines += "| trazabilidad_actualizada | en curso | 100% | en seguimiento |"
$lines += ""
$lines += "## Pendientes detectados"
$lines += ""

if ($pendingNames.Count -eq 0) {
    $lines += "- No hay puntos pendientes en el plan maestro."
} else {
    foreach ($pending in $pendingNames) {
        $lines += "- " + $pending
    }
}

$lines += ""
$lines += "## Acciones 30/60/90 dias"
$lines += ""
$lines += "- 30 dias: ejecutar validacion tecnica mensual y cerrar gate UAT operativo de punto 13."
$lines += "- 60 dias: consolidar seguimiento de puntos en curso con reporte comparativo mensual."
$lines += "- 90 dias: evaluar cumplimiento trimestral de KPI y ajustar roadmap del siguiente trimestre."
$lines += ""
$lines += "## Fuentes"
$lines += ""
$lines += "- documentos/plan_maestro_pos_multiempresa_14_puntos.md"
$lines += "- documentos/punto_13_validacion_integral_resultado.md"
$lines += "- documentos/roadmap_trimestral_pos_multiempresa.md"

Set-Content -Path $reportPath -Value ($lines -join "`r`n") -Encoding UTF8

Write-Host "Reporte de operacion continua generado" -ForegroundColor Green
Write-Host ("Archivo: " + $reportPath)
Write-Host ("Log: " + $logPath)
