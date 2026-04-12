<#
. Scrip para ejecutar tests módulo a módulo y opcionalmente importar SQL de semilla.
. Uso:
.  - Ejecutar sin semilla: .\run_tests_by_module.ps1
.  - Ejecutar con semilla: .\run_tests_by_module.ps1 -Seed
.  - Cambiar archivo seed: .\run_tests_by_module.ps1 -Seed -SeedFile "C:\ruta\a\seed_data.sql"
.
. El script intenta usar `sqlite3.exe` ubicado en la raíz del repo para importar `seed_data.sql`
. al archivo `testdata/seed.db` si se activa `-Seed`.
.
#>
Param(
    [switch]$Seed,
    [string]$SeedFile = "$PSScriptRoot/seed_data.sql",
    [string]$LogDir = "$PSScriptRoot/../logs/test_runs"
)

Set-StrictMode -Version Latest

$repoRoot = Resolve-Path "$PSScriptRoot/.."
$backendDir = Join-Path $repoRoot 'backend'

if (-not (Test-Path $LogDir)) {
    New-Item -ItemType Directory -Path $LogDir -Force | Out-Null
}

$timestamp = Get-Date -Format 'yyyyMMdd-HHmmss'

# Ajustar lista de módulos que se testearán (en orden recomendando)
$modules = @('db','auth','handlers','utils','metrics')

Write-Host "Repositorio: $repoRoot"

if ($Seed) {
    Write-Host "Modo seed activado. Intentando importar: $SeedFile"
    if (Test-Path $SeedFile) {
        $sqliteExe = Join-Path $repoRoot 'sqlite3.exe'
        if (Test-Path $sqliteExe) {
            $seedDbDir = Join-Path $repoRoot 'testdata'
            if (-not (Test-Path $seedDbDir)) { New-Item -ItemType Directory -Path $seedDbDir | Out-Null }
            $seedDbPath = Join-Path $seedDbDir 'seed.db'
            if (Test-Path $seedDbPath) { Remove-Item $seedDbPath -Force }
            Write-Host "Importando SQL de semilla a $seedDbPath usando sqlite3.exe"
            & $sqliteExe $seedDbPath ".read `"$SeedFile`""
            if ($LASTEXITCODE -ne 0) { Write-Warning "Falló importación de seed (exit=$LASTEXITCODE)" } else { Write-Host "Seed importado correctamente." }
        } else {
            Write-Warning "No se encontró sqlite3.exe en la raíz del repo ($repoRoot). Omitiendo importación SQL."
        }
    } else {
        Write-Warning "Archivo seed no existe: $SeedFile"
    }
}

$results = @{}

foreach ($m in $modules) {
    Write-Host "`n=== Ejecutando tests: $m ==="
    Push-Location $backendDir
    $logFile = Join-Path $LogDir "$($timestamp)-$($m).log"
    $package = "./$m"
    Write-Host "Comando: go test $package -v | tee $logFile"
    & go test $package -v 2>&1 | Tee-Object -FilePath $logFile
    $exit = $LASTEXITCODE
    $results[$m] = $exit
    if ($exit -eq 0) { Write-Host "PASS: $m" } else { Write-Host "FAIL: $m (exit $exit)" }
    Pop-Location
}

Write-Host "`nResumen de ejecución:";
foreach ($k in $results.Keys) {
    $status = if ($results[$k] -eq 0) { 'PASS' } else { 'FAIL' }
    Write-Host "$k : $status (exit $($results[$k]))"
}

Write-Host "Logs guardados en: $LogDir"

Exit 0
