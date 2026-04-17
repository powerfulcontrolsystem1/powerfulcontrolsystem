# Revisión ortográfica del repositorio

Este conjunto de scripts facilita la revisión ortográfica de los textos visibles al usuario (HTML, Markdown, TXT).

Requisitos (elige uno):

- Node.js (para usar `cspell` vía `npx`) — recomendado
- Python 3 con las dependencias `pyspellchecker` y `beautifulsoup4` (fallback)

Instalación de dependencias (opcional):

Para cSpell (global):
```powershell
npm install -g cspell
```

Usar npx (sin instalación global):
```powershell
npx -y cspell@6 --help
```

Para Python (fallback):
```powershell
py -m pip install --user pyspellchecker beautifulsoup4
```

Uso:

PowerShell (Windows):
```powershell
.\scripts\spellcheck.ps1 .
```

Bash (Linux/macOS):
```bash
./scripts/spellcheck.sh .
```

Ejecutar directamente el fallback Python:
```powershell
python scripts/spellcheck.py .
```

Personalización:

- Añade términos específicos del proyecto a `scripts/spell_whitelist.txt` para que no sean marcados.
- Ajusta `.cspell.json` si prefieres cSpell y quieres reglas más estrictas.
