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

### Navegador interno de Codex

Cuando el plugin Browser este disponible, preferir el navegador interno para
validar PCS visualmente. En un chat nuevo:

1. Leer el skill `control-in-app-browser` instalado en
   `%USERPROFILE%\.codex\plugins\cache\openai-bundled\browser\*\skills\control-in-app-browser\SKILL.md`.
2. Inicializar el runtime con `scripts/browser-client.mjs` del mismo plugin y
   seleccionar `iab`.
3. Emitir y leer completa la documentacion de `browser.documentation()`.
4. Reutilizar una pestana existente si ya esta en PCS; si no, crear una nueva.
5. Para acciones visibles, usar locators estables (`id`, `data-*`, labels) y
   confirmar que apuntan a un unico elemento antes de hacer clic o escribir.
6. Para responsive, usar la capacidad `viewport` del navegador si esta
   disponible; si no, validar con dimensiones de ventana equivalentes y una
   lectura DOM de `documentElement.scrollWidth <= innerWidth`.

No usar el navegador para enviar formularios destructivos, cerrar ventas reales,
cancelar carritos, enviar correos o cambiar permisos sin autorizacion explicita.

### Prueba visual rapida del carrito PCS

URL base:

```text
https://powerfulcontrolsystem.com/administrar_empresa/carrito_de_compras.html?modo=venta_directa&perm_page=linkVentaDirecta&empresa_id=12&qa={timestamp}
```

Checklist:

- Confirmar sesion activa y que no aparezca login.
- Buscar por nombre, por ejemplo `menta`.
- Esperar resultados visibles y seleccionar uno con mouse, o usar el primer
  resultado resaltado.
- Presionar `Agregar` y comprobar que el item aparece en el detalle o que sube
  su cantidad.
- Usar los botones `+` y `-` de cantidad del item y confirmar que el numero se
  ve y los totales cambian.
- Revisar que nombres de producto, cantidad, precios, descuento, impuesto,
  total y acciones esten alineados y legibles.
- Probar campos de medios de pago combinados escribiendo, borrando y cambiando
  entre efectivo, credito, debito y transferencias sin que el foco salte al
  buscador.
- En celular, confirmar que no haya scroll horizontal y que las tarjetas queden
  apiladas: buscador, cliente, productos, pago, acciones y totales.
- No presionar `Pagar y cerrar carrito`, `Cancelar carrito` ni acciones de
  devolucion/cierre si el usuario no lo autorizo para datos reales.

## Pruebas reales en produccion PCS

Cuando el usuario pida probar `powerfulcontrolsystem.com`, DIAN, carrito o una
venta real de la empresa Powerful Control System, no iniciar probando en local
salvo que la tarea diga explicitamente localhost. Usar:

```text
https://powerfulcontrolsystem.com
empresa_id=12
```

Si el usuario entrego credenciales para esa prueba, autenticar por navegador real
o por API con cookie temporal bajo `.gotmp`, sin imprimir la clave. Ejemplo de
patron API:

```powershell
$base = "https://powerfulcontrolsystem.com"
$cookie = "D:\powerfulcontrolsystem\.gotmp\pcs_cookie.txt"
curl.exe --ssl-no-revoke -sS -c $cookie -H "Content-Type: application/json" -X POST "$base/super/api/administradores/login" --data-binary "@D:\powerfulcontrolsystem\.gotmp\login_payload.json"
curl.exe --ssl-no-revoke -sS -b $cookie "$base/api/empresa/facturacion_electronica/dian?action=diagnostico_oficial&empresa_id=12"
```

Para facturacion electronica DIAN de PCS, el cierre minimo es:

- Verificar diagnostico oficial y configuracion DIAN sin mostrar secretos.
- Crear o reutilizar venta de producto `menta` y cliente natural/empresa segun
  lo pedido.
- Emitir con `/api/empresa/facturacion_electronica?action=emitir&empresa_id=12`
  o reintentar con `action=reenviar_dian`.
- Revisar `integracion_fiscal.estado_envio`, `numero_legal`, `cola_reintentos`
  y reglas DIAN (`FAB05c`, `FAD06`, etc.).
- Si DIAN responde HTTP 200 con `StatusCode=99`, la conexion funciono y el
  rechazo es normativo/configuracion/XML, no caida de red.

Usar navegador interno o Chrome solo para validar pantallas y flujo visible:
login, seleccionar empresa, carrito, cliente, totales, factura/impresion. Para
consultas de estado y reintentos DIAN preferir API autenticada porque conserva
la evidencia exacta y reduce errores visuales.

Despues de `.\rs.ps1`, validar siempre contra el dominio publico con parametro
`qa={timestamp}`. Si el navegador conserva cache anterior, recargar la pestana
o cambiar el parametro `qa`.

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
