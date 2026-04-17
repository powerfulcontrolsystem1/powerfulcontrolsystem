<#
PowerShell wrapper for repository spellcheck.
Prefers `cspell` via `npx` when available, otherwise runs the Python fallback `scripts/spellcheck.py`.
#>
param(
  [string]$Path = "."
)

$patterns = @(
  "web/**/*.html",
  "web/**/*.htm",
  "web/**/*.md",
  "documentos/**/*.md",
  "web/**/*.js"
)

Write-Host "Spellcheck: path = $Path"

if (Get-Command npx -ErrorAction SilentlyContinue) {
  Write-Host "Usando cSpell (npx). Esto puede descargar temporalmente el paquete si no existe localmente."
  & npx -y cspell@6 --config .cspell.json @patterns
  exit $LASTEXITCODE
}

if (Get-Command python -ErrorAction SilentlyContinue) {
  Write-Host "cSpell no encontrado. Usando fallback Python (scripts/spellcheck.py)."
  python scripts/spellcheck.py $Path
  exit $LASTEXITCODE
}

Write-Error "Ni 'npx' ni 'python' están disponibles en PATH. Instala Node.js (cSpell) o Python para usar este script."
exit 2
