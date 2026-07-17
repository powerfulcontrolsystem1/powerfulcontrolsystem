param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$required = @(
  "PCS_API_IMAGE_DIGEST",
  "PCS_MIGRATE_IMAGE_DIGEST",
  "PCS_WORKER_IMAGE_DIGEST"
)
$pattern = '^[^@\s]+@sha256:[a-fA-F0-9]{64}$'

foreach ($key in $required) {
  $value = [Environment]::GetEnvironmentVariable($key)
  if ([string]::IsNullOrWhiteSpace($value)) {
    throw "$key es obligatoria para un release inmutable."
  }
  if ($value -notmatch $pattern) {
    throw "$key debe usar una referencia completa repositorio@sha256:<64 hex>."
  }
}

Write-Host "[OK] Imagenes API, migrador y worker fijadas por digest." -ForegroundColor Green
