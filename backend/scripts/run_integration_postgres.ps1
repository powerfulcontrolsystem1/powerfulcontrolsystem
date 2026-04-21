<#
.SYNOPSIS
  Aplica seeds SQL a bases Postgres y ejecuta tests Go.

.NOTES
  - Requiere `psql` o `docker` para aplicar los archivos SQL.
  - Define las variables de entorno `DB_SUPERADMIN_DSN` y `DB_EMPRESAS_DSN`.
  - Uso típico:
      Set-Location 'D:\powerfulcontrolsystem\backend'
      $env:DB_SUPERADMIN_DSN = 'postgres://user:pass@host:5432/super'
      $env:DB_EMPRESAS_DSN = 'postgres://user:pass@host:5432/empresas'
      .\scripts\run_integration_postgres.ps1 -Packages './handlers' -RunAll
#>

[CmdletBinding()]
param(
  [switch]$UseDocker,
  [string]$Packages = "./handlers",
  [switch]$RunAll
)

function Apply-SqlFile {
  param(
    [string]$Dsn,
    [string]$FilePath
  )

  Write-Host "Aplicando $FilePath en DSN: $Dsn"
  if (-not $Dsn) {
    Write-Error "DSN vacío. Define DB_SUPERADMIN_DSN y DB_EMPRESAS_DSN."
    return 1
  }

  $psqlCmd = Get-Command psql -ErrorAction SilentlyContinue
  if ($psqlCmd -and -not $UseDocker) {
    & $psqlCmd.Path $Dsn -f $FilePath
    return $LASTEXITCODE
  }

  $dockerCmd = Get-Command docker -ErrorAction SilentlyContinue
  if ($dockerCmd) {
    # Usar docker: requiere que la cadena DSN sea accesible desde el contenedor
    $pwd = (Get-Location).Path -replace '\\','/'
    Write-Host "Usando docker para ejecutar psql en imagen 'postgres:15' (asegúrate que el host en la DSN sea accesible desde el contenedor)."
    & $dockerCmd.Path 'run' '--rm' '-v' "$pwd:/work" '-w' '/work' 'postgres:15' 'psql' "$Dsn" '-f' "/work/$FilePath"
    return $LASTEXITCODE
  }

  Write-Error "No se encontró 'psql' ni 'docker'. Instala uno de los dos o usa -UseDocker con docker disponible."
  return 2
}

# Obtener / solicitar DSNs si no están en entorno
if (-not $env:DB_SUPERADMIN_DSN -or $env:DB_SUPERADMIN_DSN.Trim() -eq '') {
  $env:DB_SUPERADMIN_DSN = Read-Host "Define DB_SUPERADMIN_DSN (ej: postgres://user:pass@host:5432/super)"
}
if (-not $env:DB_EMPRESAS_DSN -or $env:DB_EMPRESAS_DSN.Trim() -eq '') {
  $env:DB_EMPRESAS_DSN = Read-Host "Define DB_EMPRESAS_DSN (ej: postgres://user:pass@host:5432/empresas)"
}

$superFile = 'scripts/seed_postgres_super.sql'
$empFile = 'scripts/seed_postgres_empresas.sql'

Write-Host "Aplicando seed en super..."
$rc = Apply-SqlFile $env:DB_SUPERADMIN_DSN $superFile
if ($rc -ne 0) { Write-Error "Fallo aplicando seed super (code $rc)" ; exit $rc }

Write-Host "Aplicando seed en empresas..."
$rc = Apply-SqlFile $env:DB_EMPRESAS_DSN $empFile
if ($rc -ne 0) { Write-Error "Fallo aplicando seed empresas (code $rc)" ; exit $rc }

if ($RunAll) {
  Write-Host "Ejecutando toda la suite de pruebas: go test ./..."
  go test ./... -count=1
} else {
  Write-Host "Ejecutando pruebas para paquetes: $Packages"
  go test $Packages -v -count=1
}

Write-Host "Proceso finalizado. Revisa salida para detalles."
