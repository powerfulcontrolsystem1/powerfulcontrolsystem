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

## Validacion de textos y codificacion

Antes de cerrar cambios que toquen textos visibles, ayudas, plantillas de correo,
mensajes backend o documentacion operativa, buscar caracteres rotos por
codificacion. El objetivo es no publicar palabras con tildes rotas, secuencias
de doble codificacion o caracteres de reemplazo en pantallas del sistema.

```powershell
$badEncodingPattern = ([char]0xFFFD) + "|" + ([char]0x00C3) + "|" + ([char]0x00D2) + "|[A-Za-zÁÉÍÓÚáéíóúÑñ]\?[A-Za-zÁÉÍÓÚáéíóúÑñ]"
rg -n $badEncodingPattern web backend scripts documentos CHANGELOG.md AGENTS.md -g "*.html" -g "*.js" -g "*.css" -g "*.go" -g "*.md" -g "*.txt" -g "*.ps1" -g "*.json" -g "*.yaml" -g "*.yml" -g "*.sql" -g "!documentos/historial_de_cambios"
```

Revisar manualmente los resultados porque las URLs con query string pueden dar
falsos positivos por `?action=...`. Si se corrige un archivo, conservarlo en
UTF-8 y volver a ejecutar el barrido.

## Preflight

```powershell
.\scripts\profesional_preflight.ps1
.\scripts\profesional_preflight.ps1 -Full
```

Usar preflight antes de sincronizaciones o cambios grandes. Si falla, corregir la
causa concreta o dejar el riesgo documentado.

## rs

El usuario suele pedir `ejecuta rs`. El script canonico vive en `scripts`:

```powershell
.\scripts\rs.ps1
```

No depender de un wrapper en la raiz del proyecto. Revisar el contenido del
script antes de asumir su alcance. Puede encadenar preflight, actualizacion,
sincronizacion y pasos operativos.

`rs.ps1` ejecuta cada script interno en un proceso PowerShell aislado, con
archivos separados de salida y error bajo `scripts/logs/rs-*.log`. Esta regla
evita que un `exit` de preflight, actualizacion o sincronizacion cierre el
orquestador antes de los pasos restantes; el codigo de salida se conserva y
detiene el flujo solo cuando el paso correspondiente falla. Cada fase tiene
timeout controlado (3600 segundos por defecto) y reporta las rutas de log si se
agota o falla. Puede ajustarse con `-StepTimeoutSeconds`.

Si GitHub protege `main` y rechaza el push directo, `actualizar_repositorio.ps1`
crea una rama `codex/rs-...`, abre la PR y solicita `auto-merge`. Nunca se
autoaprueba ni evita checks: GitHub solo la fusiona despues de una aprobacion
independiente y verificaciones verdes. Si el repositorio permite administracion
por GitHub CLI pero tiene Auto-merge desactivado, el script lo habilita sin
cambiar las reglas de proteccion. `rs` espera hasta 900 segundos por
defecto; mientras la PR siga pendiente termina sin sincronizar la VPS. Para
ajustar la espera o desactivar auto-merge:

```powershell
.\scripts\rs.ps1 -ProtectedMainPRWaitSeconds 1800
.\scripts\rs.ps1 -NoAutoMergeProtectedPR
```

Cuando GitHub fusione esa PR mediante `squash`, el commit de `main` tendra otra
identidad aunque contenga el mismo cambio. El actualizador primero intenta
fast-forward y, solo si el arbol esta limpio, reconcilia mediante rebase. Si hay
conflicto o `HEAD` no termina exactamente igual a `origin/main`, aborta y `rs`
no sincroniza la VPS.

El hijo se resuelve como `pwsh.exe` cuando el orquestador se ejecuta en
PowerShell Core y como `powershell.exe` en Windows PowerShell, con fallback a
un comando instalado. No se debe asumir que `$PSHOME` contiene ambos binarios.

Si el arbol esta limpio pero la rama contiene commits locales sin upstream o
sin publicar, `actualizar_repositorio.ps1` publica `HEAD` y configura el
upstream antes de continuar. Asi `rs` no omite una rama de trabajo solo porque
los cambios ya fueron commiteados manualmente.

Durante la extraccion remota, `sync_to_vps.ps1` borra las rutas retiradas
`web/Juegos`, `juegos` y `web/img/juegos` antes de aplicar el paquete. Esto
evita que los archivos estaticos de un modulo eliminado sobrevivan a una
sincronizacion incremental.

## Roles de plataforma

El despliegue Docker ejecuta `pcs-migrate` antes de la API y mantiene
`pcs-worker` como proceso separado. En una consola con las DSN privadas ya
cargadas, los binarios se validan sin abrir HTTP:

```powershell
Set-Location D:\powerfulcontrolsystem\backend
go build ./cmd/pcs-migrate
go build ./cmd/pcs-worker
```

No establecer `PCS_RUNTIME_SCHEMA_BOOTSTRAP=0` en una instalacion existente
hasta verificar el ledger de migraciones y los flujos de provisionamiento.

## Backup completo del VPS

El backup operativo independiente del VPS se ejecuta con:

```powershell
.\scripts\crear_backup_vps.ps1
```

El script abre una ventana pequena con progreso de 0 a 100 y guarda cada copia
en una carpeta nueva bajo:

```text
D:\Backup vps PCS
```

Los scripts operativos `vps_backup_operacion.ps1` y
`vps_restore_validation.ps1` no contactan ningun servidor por defecto. Exigen
`-AllowRemoteTarget` despues de confirmar que el destino es staging aislado o
una operacion remota expresamente autorizada. La compuerta `release_gate.ps1`
aplica la misma regla.

Cada backup incluye inventario del VPS, dump logico PostgreSQL, imagenes Docker
locales PCS, volumenes Docker, archivos del proyecto filtrados, SHA256, manifest
local y un restaurador `restore_to_new_vps.sh` dentro del paquete. No imprimir
secretos, `.env`, claves privadas, certificados ni DSN durante la ejecucion o al
reportar resultados.

Para subir una copia a un VPS nuevo y dejar preparada la restauracion:

```powershell
.\scripts\crear_backup_vps.ps1 -Restore -BackupPath "D:\Backup vps PCS\PCS_VPS_backup_YYYYMMDD_HHMMSS\pcs_vps_full_backup_YYYYMMDD_HHMMSS.tar.gz" -TargetHost "IP_NUEVO_VPS"
```

Por seguridad, el modo `-Restore` solo sube y prepara el paquete. Ejecutar la
restauracion remota destructiva requiere agregar `-ExecuteRemoteRestore` despues
de validar que el VPS destino es el correcto.

## Consola de recuperacion

Interfaz local con botones para operaciones de recuperacion y publicacion:

```powershell
.\scripts\consola_de_recuperacion.ps1
```

La ventana permite ejecutar `crear_backup_vps`, preparar restauracion de un
backup en un VPS nuevo, `sync_to_vps`, `actualizar_repositorio` y `rs`. Cada
ejecucion escribe log en `scripts\logs\consola_de_recuperacion_*.log`.

## Logo de correos en Gmail

Los correos HTML de `@powerfulcontrolsystem.com` incrustan el logo corporativo,
pero el avatar circular que muestra Gmail en celular no se toma del HTML del
mensaje. Para que Gmail reemplace la letra por el logo del dominio se debe
publicar BIMI en DNS, con DMARC alineado. Activo publico preparado:

```text
https://powerfulcontrolsystem.com/img/bimi-pcs.svg
```

Registro DNS esperado cuando el dominio ya tenga DMARC en enforcement:

```text
default._bimi.powerfulcontrolsystem.com TXT "v=BIMI1; l=https://powerfulcontrolsystem.com/img/bimi-pcs.svg; a="
```

Si se adquiere certificado VMC/CMC, completar `a=` con la URL publica del
certificado. Sin BIMI/DMARC en DNS, Gmail puede seguir mostrando una inicial.


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

## sync_to_vps2

VPS2 es el servidor local de pruebas. Su operacion esta documentada en
`documentos/vps2_operacion.md`.

```powershell
.\scripts\sync_to_vps2.ps1
```

Usos frecuentes:

```powershell
.\scripts\sync_to_vps2.ps1 -SkipDeploy
.\scripts\sync_to_vps2.ps1 -SkipDisableGui -SkipNextcloud
.\scripts\sync_to_vps2.ps1 -RestartDockerStack:$false
```

El script lee `PcsVps2Host`, `PcsVps2User`, `PcsVps2Port`,
`PcsVps2RemotePath`, `PcsVps2HostKey`, `PcsVps2IdentityFile`,
`PcsVps2RepoUrl` y, si no hay llave SSH, `PcsVps2Password` desde
`scripts/pcs_deployment.local.ps1` o variables `PCS_VPS2_*`.
No guardar claves en archivos versionados.

## Conexion SSH al VPS

La configuracion local privada vive en:

```text
scripts/pcs_deployment.local.ps1
```

Ese archivo esta ignorado por Git y puede contener host, usuario, puerto, ruta
remota, host key y llave privada. No imprimir sus valores completos en consola,
documentacion ni respuestas.

Para cargar la configuracion y abrir una sesion SSH manual desde PowerShell:

```powershell
Set-Location D:\powerfulcontrolsystem
. .\scripts\pcs_deployment.local.ps1
$ssh = "C:\Windows\System32\OpenSSH\ssh.exe"
$target = "$script:PcsVpsUser@$script:PcsVpsHost"
$args = @("-p", [string]$script:PcsVpsPort, "-o", "StrictHostKeyChecking=accept-new")
if ($script:PcsVpsIdentityFile) { $args += @("-i", [string]$script:PcsVpsIdentityFile) }
& $ssh @args $target
```

Para ejecutar un comando remoto puntual sin abrir consola interactiva:

```powershell
. .\scripts\pcs_deployment.local.ps1
$ssh = "C:\Windows\System32\OpenSSH\ssh.exe"
$target = "$script:PcsVpsUser@$script:PcsVpsHost"
$args = @("-p", [string]$script:PcsVpsPort, "-o", "StrictHostKeyChecking=accept-new")
if ($script:PcsVpsIdentityFile) { $args += @("-i", [string]$script:PcsVpsIdentityFile) }
& $ssh @args $target "cd '$script:PcsVpsRemotePath' && docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml ps"
```

Reglas de seguridad:

- Nunca imprimir `deploy/.env.platform`, DSN completos, `CONFIG_ENC_KEY`,
  `POSTGRES_PASSWORD`, certificados, PIN DIAN, tokens ni claves privadas.
- Preferir comandos de solo lectura primero: `docker ps`, `docker logs --tail`,
  `curl -I`, `git status`, `docker compose ps`.
- Antes de ejecutar SQL de escritura, confirmar que el `WHERE empresa_id = ...`
  esta presente y que el cambio no afecta otras empresas.
- Pasar SQL por archivo temporal en `/tmp` y eliminarlo al finalizar; no dejar
  secretos ni dumps en el repositorio.

## Docker y VPS

Consultar:

- `documentos/docker_vps_operacion.md`
- `documentos/manual_de_instalacion.md`
- `documentos/deploy_nginx_reverse_proxy_vps.md`
- `deploy/`

Antes de cambios Docker, validar que el proyecto pueda moverse sin incluir
`.env`, uploads privados, backups, certificados o datos runtime.

### Compilar y publicar en VPS

Flujo normal cuando el usuario pide publicar, sincronizar o `rs`:

```powershell
Set-Location D:\powerfulcontrolsystem
.\scripts\rs.ps1
```

`scripts\rs.ps1` es el orquestador preferido porque encadena las validaciones del proyecto,
sincroniza al VPS, reconstruye/recarga servicios y verifica salud publica segun
la configuracion vigente.

Flujo manual si se necesita separar pasos:

```powershell
.\scripts\profesional_preflight.ps1
.\scripts\actualizar_repositorio.ps1
.\scripts\sync_to_vps.ps1
```

Validacion remota despues de compilar/desplegar:

```powershell
. .\scripts\pcs_deployment.local.ps1
$ssh = "C:\Windows\System32\OpenSSH\ssh.exe"
$target = "$script:PcsVpsUser@$script:PcsVpsHost"
$args = @("-p", [string]$script:PcsVpsPort, "-o", "StrictHostKeyChecking=accept-new")
if ($script:PcsVpsIdentityFile) { $args += @("-i", [string]$script:PcsVpsIdentityFile) }
& $ssh @args $target "cd '$script:PcsVpsRemotePath' && docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml ps && curl -I http://127.0.0.1:8081/ && curl -I https://powerfulcontrolsystem.com/"
```

Para revisar errores del backend sin exponer secretos:

```powershell
& $ssh @args $target "docker logs --tail 160 pcs-backend"
```

Para revisar PostgreSQL por consola del contenedor:

```powershell
& $ssh @args $target "docker exec -i pcs-postgres sh -lc 'psql -U \"$POSTGRES_USER\" -d pcs_empresas -c \"select 1\"'"
```

Si se actualizan datos operativos en produccion, registrar el motivo en
`documentos/historial_de_cambios` y validar por API o pantalla publicada.

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

### Login API seguro para pruebas PCS

- Usar solo credenciales autorizadas explicitamente por el usuario en el chat.
- No escribir claves en documentacion, commits ni respuestas finales.
- Guardar cookies solo en `.gotmp` y eliminarlas al terminar si ya no se
  necesitan.

Ejemplo de flujo sin imprimir secretos:

```powershell
$cookie = ".gotmp\pcs_api_cookie.txt"
# Construir el payload en memoria con la clave autorizada por el usuario.
curl.exe --ssl-no-revoke -sS -c $cookie -b $cookie `
  -X POST "https://powerfulcontrolsystem.com/super/api/administradores/login" `
  -H "Content-Type: application/json" `
  --data-binary "@.gotmp\login_payload.json"
```

El login por correo de `login.html` usa `/super/api/administradores/login`.
Si reCAPTCHA o 2FA estan activos, preferir la sesion real del navegador interno
o Chrome autorizado por el usuario.

### Numeracion DIAN PCS 2026-06-17

PDF autorizado por el usuario:

```text
C:\Users\ivanm\Documents\18764111318575 Autorizacion numercion DIAN 17 JUNIO 2026.pdf
```

Importar PDF Formulario 1876 con IA GPT-5.5, igual que el boton visible de
`facturacion_electronica.html`:

```powershell
curl.exe --ssl-no-revoke -sS -b .gotmp\pcs_api_cookie.txt `
  -X POST "https://powerfulcontrolsystem.com/api/empresa/facturacion_electronica/dian?action=importar_numeracion_pdf_ia&empresa_id=12" `
  -F "archivo=@C:\Users\ivanm\Documents\18764111318575 Autorizacion numercion DIAN 17 JUNIO 2026.pdf;type=application/pdf" `
  -F "empresa_id=12"
```

El endpoint `action=importar_numeracion_pdf` queda como respaldo tecnico local
para pruebas automatizadas cuando IA no este disponible, pero el flujo visual
principal debe usar IA y permitir digitacion manual en los campos existentes.

### Pruebas visuales con navegador desde Codex

Cuando el usuario pida "prueba visualmente", Codex debe intentar primero la
herramienta de navegador interna o la extension de Chrome si esta disponible en
el hilo. Si esas herramientas no aparecen en `tool_search`, usar Playwright
desde el workspace contra la URL publicada o local.

Flujo recomendado con Playwright:

```powershell
# Usar una carpeta temporal ignorada por Git para capturas.
New-Item -ItemType Directory -Force .gotmp\visual | Out-Null

# Abrir paginas publicadas y guardar evidencia visual.
node .gotmp\visual_check.mjs
```

El script debe:

- Abrir `https://powerfulcontrolsystem.com/login.html` y autenticar solo si el
  usuario autorizo credenciales en el chat actual.
- Entrar a PCS con `empresa_id=12`.
- Probar escritorio y celular con `page.setViewportSize`.
- Capturar consola, errores de pagina y screenshots en `.gotmp\visual`.
- Revisar que no haya texto cortado, botones fuera de tarjeta, selects
  ilegibles, spinners numericos indeseados ni errores JavaScript.
- Para carrito, probar busqueda por nombre, botones `+`/`-`, cantidad visible,
  pagos combinados y flujo de cliente sin enviar documentos externos si el
  usuario no lo autorizo.

No guardar claves ni cookies en documentacion. Borrar cookies temporales de
`.gotmp` al terminar si contienen sesiones.

Valores esperados del PDF PCS:

```text
Formulario: 18764111318575
Prefijo: 1PCS
Rango: 1-100000
Fecha desde: 2026-06-17
Fecha hasta: 2028-06-17
Vigencia: 24 meses
```

### Venta de prueba DIAN PCS

Datos controlados:

- Empresa: PCS, `empresa_id=12`.
- Producto: `menta`, producto `id=103`, SKU `1`, precio COP 100.
- Cliente: IVAN FRANCISCO CAYON GUARNIZO, cliente `id=22`, CC `84456779`.

Flujo API equivalente al carrito:

1. Crear carrito en `/api/empresa/carritos_compra?empresa_id=12&modo=venta_directa&perm_page=linkVentaDirecta`.
2. Activarlo con `PUT action=activar_estacion`.
3. Agregar producto por `/api/empresa/carritos_compra/items` con `permitir_sin_stock=true` si la empresa lo permite.
4. Abrir caja con `PUT action=abrir_caja_cobro`.
5. Cerrar pago con `PUT action=pagar_estacion`.
6. Si no autoemite FE, llamar `/api/empresa/facturacion_electronica?action=facturar_desde_venta&empresa_id=12` con `tipo_documento=comprobante_pago`, `documento_codigo` y `cliente_id=22`.
7. Revisar `integracion_fiscal.cola_reintentos`, numero legal y reglas DIAN.

Resultado de referencia 2026-06-17: factura `FV-FE-MENTA-20260617151719`,
numero legal `1PCS1`, enviada a DIAN y rechazada por `FAK61`, `FAB05c` y
`FAD06`. El error de rango/prefijo de la resolucion anterior ya no aparecio.

Revalidacion 2026-06-18: despues de asociar la numeracion en portal DIAN
produccion y consultar `GetNumberingRange`, PCS emitio factura `1PCS2` por
producto `menta`. Luego el usuario confirmo en portal DIAN produccion que
`1PCS2` aparece como `Aprobado con notificacion`. Si un reenvio devuelve
`Regla: 90, Documento procesado anteriormente`, tratar esa regla como pendiente
de consulta del acuse original y revisar portal/CUFE antes de reenviar.

Resultado real 2026-06-18: prueba viva en VPS emitio `1PCS3` contra DIAN
produccion por SOAP/WCF `SendBillSync`; DIAN respondio HTTP 200 con
`estado_dian=aceptado`, `acuse_estado=aceptado`, CUFE registrado y notificacion
`RUT01` informativa.

Confirmacion portal DIAN 2026-06-18: el usuario encontro en produccion, consulta
de documentos recibidos, las facturas `1PCS3` del 18-06-2026 y `1PCS2` del
17-06-2026 como `Aprobado con notificacion`, valor `$ 100`. Despues de esa
prueba los contadores quedaron en siguiente consecutivo `1PCS4`.

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
- Si el rechazo incluye `FAB05c`, verificar primero la asociacion del rango en
  `https://catalogo-vpfe.dian.gov.co/User/Login`.
- Si incluye `FAD06`, volver a consultar clave tecnica DIAN y revisar CUFE,
  prefijo, consecutivo, fecha/hora, impuestos y totales.
- Si incluye `Regla 90`, no marcar como aceptado por esa regla sola; consultar
  primero portal DIAN, CUFE o acuse original. Si el documento ya aparece
  aprobado, continuar con el siguiente consecutivo y no reenviar el mismo folio.
- `Aprobado con notificacion` en DIAN cuenta como documento aprobado; documentar
  la notificacion (`RUT01`, etc.) y corregir datos maestros si aplica.

Usar navegador interno o Chrome solo para validar pantallas y flujo visible:
login, seleccionar empresa, carrito, cliente, totales, factura/impresion. Para
consultas de estado y reintentos DIAN preferir API autenticada porque conserva
la evidencia exacta y reduce errores visuales.

Despues de `.\scripts\rs.ps1`, validar siempre contra el dominio publico con parametro
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
