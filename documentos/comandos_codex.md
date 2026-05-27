# Comandos para Codex

Comandos confirmados para operar y validar este repositorio desde PowerShell.
No imprimir secretos ni variables privadas completas.

## Ubicacion

```powershell
Set-Location D:\powerfulcontrolsystem
```

## Pruebas Go

Ejecutar desde `backend`:

```powershell
Set-Location D:\powerfulcontrolsystem\backend
go test ./handlers -run NombreDePrueba -count=1
go test ./db ./handlers -run "Patron1|Patron2" -count=1
go test ./... -run "^$" -count=1
```

Usar pruebas dirigidas primero. `go test ./...` puede tardar mas y debe
reservarse para cambios transversales.

## Validaciones de PowerShell

```powershell
[System.Management.Automation.Language.Parser]::ParseFile("scripts\sync_to_vps.ps1",[ref]$null,[ref]$null)
[System.Management.Automation.Language.Parser]::ParseFile("scripts\rs.ps1",[ref]$null,[ref]$null)
```

## Validaciones HTML y JavaScript

Node disponible en este entorno:

```powershell
C:\Users\ivanm\.cache\codex-runtimes\codex-primary-runtime\dependencies\node\bin\node.exe --version
```

Para sintaxis de JS externo:

```powershell
C:\Users\ivanm\.cache\codex-runtimes\codex-primary-runtime\dependencies\node\bin\node.exe --check web\js\archivo.js
```

Para paginas HTML con scripts embebidos, preferir helpers existentes si los hay
o usar Node para extraer scripts y validarlos sin ejecutar llamadas reales.

## Preflight

```powershell
.\scripts\profesional_preflight.ps1
.\scripts\profesional_preflight.ps1 -Full
```

Usar preflight antes de sincronizaciones o cambios grandes. Si falla, corregir la
causa concreta o dejar el riesgo documentado.

## rs

El usuario suele pedir `ejecuta rs`. El wrapper raiz es:

```powershell
.\rs.ps1
```

El script interno relacionado es:

```powershell
.\scripts\rs.ps1
```

Revisar el contenido del script antes de asumir su alcance. Puede encadenar
preflight, actualizacion, sincronizacion y pasos operativos.

## sync_to_vps

```powershell
.\scripts\sync_to_vps.ps1
```

Modos utiles segun necesidad:

```powershell
.\scripts\sync_to_vps.ps1 -PreviewOnly
.\scripts\sync_to_vps.ps1 -DryRun
.\scripts\sync_to_vps.ps1 -DeploymentMode docker
.\scripts\sync_to_vps.ps1 -CleanupRemoteUnusedFiles:$false
```

No mostrar credenciales, llaves ni hosts privados sensibles en respuestas.

## Docker y VPS

Consultar:

- `documentos/docker_vps_operacion.md`
- `documentos/manual_de_instalacion.md`
- `documentos/deploy_nginx_reverse_proxy_vps.md`
- `deploy/`

Antes de cambios Docker, validar que el proyecto pueda moverse sin incluir
`.env`, uploads privados, backups, certificados o datos runtime.

## Validacion visual

Chrome for Testing instalado para pruebas locales:

```text
C:\Users\ivanm\AppData\Local\CodexBrowserTools\chrome-for-testing\149.0.7827.22\chrome-win64\chrome.exe
```

Herramientas auxiliares:

```text
C:\Users\ivanm\AppData\Local\CodexBrowserTools\capture-url.ps1
C:\Users\ivanm\AppData\Local\CodexBrowserTools\browser-config.json
```

Playwright disponible por runtime Node:

```text
C:\Users\ivanm\.cache\codex-runtimes\codex-primary-runtime\dependencies\node\node_modules\.pnpm\playwright@1.60.0\node_modules
```

Para frontend, hacer prueba visual cuando el cambio afecte pantallas, botones,
formularios, impresion o responsive. En impresiones POS/carta, revisar captura o
HTML imprimible en blanco y negro.

## Validacion de diff

```powershell
git diff --check
```

Ejecutar al final de cambios de texto/codigo para detectar espacios invalidos o
conflictos. Las advertencias de fin de linea CRLF pueden aparecer en archivos
Windows; distinguirlas de errores reales.

## Regla sobre Python

El proyecto no usa Python como runtime. Para tareas del repositorio preferir Go,
PowerShell o Node segun corresponda. Python solo seria una herramienta local
temporal si no hay alternativa razonable y nunca debe introducirse como
dependencia del proyecto.

