# Comandos para Codex

Estado: Vigente. Responsable: QA/operación. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Se retiraron recetas con datos personales/folios reales y se enlazan contratos vigentes.
- Los comandos son interfaces de herramientas, no aprobación de sus efectos. Inventariar target SSH, nombres Compose y dependencias antes de ejecución; Skip no equivale a PASS.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

## Control documental

Desde la raíz del repositorio, con Node disponible:

```powershell
node --test tools/docs_catalog.test.mjs
node tools/docs_catalog.mjs --write
node tools/docs_catalog.mjs --check
git diff --check
```

`--write` regenera únicamente `documentos/catalogo_documental.md` y su JSON;
revisar su diff antes de entregar. `--check` es de solo lectura y falla ante
drift, fuentes mantenidas ausentes, metadatos incompletos o enlaces locales
rotos de esas fuentes. La política de responsables/fechas se edita en
`documentos/gobernanza_tecnica/politica_catalogo.json`, no se autogenera.

El inventario usa archivos Git y nuevos no ignorados, con hashes normalizados a
LF. No inspecciona archivos privados ignorados ni ejecuta PCS. Verifica enlaces
Markdown explícitos, referencias completas y anclas; no interpreta rutas en
backticks como enlaces, URLs externas, todos los dialectos Markdown ni contenido
semántico. Los documentos heredados conservan estado de revisión pendiente y
sus hallazgos visibles. Este control no sustituye seguridad ni aceptación funcional.


Comandos confirmados para operar y validar este repositorio desde PowerShell.
No imprimir secretos ni variables privadas completas.

## Plan 108: ensayo destructivo aislado de migración vacía

Este ensayo crea recursos Docker efímeros con prefijo `p108-empty-`, ejecuta el
migrador exacto dos veces, altera un checksum solo en el ledger temporal,
comprueba fallo cerrado sin cambio de esquema, restaura el checksum controlado
y verifica recuperación. Finalmente elimina todos los recursos mediante
`trap`. Nunca se debe apuntar al volumen de staging o producción.

```bash
PROJECT_DIR=/ruta/checkout-candidato \
SOURCE_ENV=/ruta/privada/.env.staging \
PCS_MIGRATE_IMAGE_DIGEST=ghcr.io/organizacion/imagen@sha256:<64-hex> \
P108_DRILL_ID=p108-empty-<sha-corto> \
bash deploy/scripts/vps-p108-empty-migration-drill.sh
```

El script rechaza recursos preexistentes, un ID fuera del prefijo autorizado,
una imagen sin digest o un fallo de migración que no quede auditado y sin
cambios de esquema/ledger.

Para ampliar el ensayo a los fallos P109 antes/durante migración y a la
compatibilidad hacia atrás, añadir:

```bash
PCS_PREVIOUS_API_IMAGE_DIGEST=ghcr.io/organizacion/pcs-api@sha256:<digest-anterior> \
P109_VERIFY_MIGRATION_FAILURES=1 \
P109_BACKWARD_HOST_PORT=<puerto-loopback-libre>
```

El modo ampliado primero niega DDL a un rol migrador restringido y exige que el
esquema quede intacto. Después provoca un fallo real entre el DDL de una
migración y su inserción en el ledger, verifica rollback y auditoría, recupera
índice/ledger atómicamente y arranca la API anterior contra el esquema nuevo.
Todo ocurre en base, volumen, red y puerto efímeros; no usar imágenes sin digest
ni apuntar las DSN a staging o producción.

Para ensayar un upgrade desde una copia lógica de staging sin escribir en el
origen:

```bash
PROJECT_DIR=/ruta/checkout-candidato \
SOURCE_ENV=/ruta/privada/.env.staging \
SOURCE_POSTGRES_CONTAINER=pcs-staging-postgres \
PCS_MIGRATE_IMAGE_DIGEST=ghcr.io/organizacion/imagen@sha256:<64-hex> \
P108_DRILL_ID=p108-upgrade-<sha-corto> \
bash deploy/scripts/vps-p108-upgrade-migration-drill.sh
```

El script rechaza un contenedor origen que no use el prefijo
`pcs-staging-`, restaura en un volumen efímero y compara el número de tablas
antes/después.

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

### Barrido E2E autenticado por lotes

El auditor `tools\qa_e2e_buttons.cjs` permite dividir un inventario grande sin
perder el orden determinista. Definir el destino, empresa y credenciales solo
en el canal seguro; no copiarlas en comandos, evidencia ni commits. Cada lote
debe usar un directorio de salida distinto y conservar `results.json` y
`reporte.md`.

```powershell
$env:PCS_QA_ROUTE_OFFSET = "0"       # 0, 50, 100, ...
$env:PCS_QA_ROUTE_BATCH_SIZE = "50"  # rutas, antes de aplicar escritorio/movil
$env:PCS_QA_OUT_DIR = "test_runs\qa_e2e_lote_000"
& $node tools\qa_e2e_buttons.cjs
```

`PCS_QA_ROUTE_OFFSET` y `PCS_QA_ROUTE_BATCH_SIZE` se aplican al inventario
ordenado antes de duplicarlo por viewport. No sustituyen la matriz manual de
acciones mutantes, roles ni botones IA: la guardia de red bloquea mutaciones.
El contador público de visitas se bloquea y queda registrado por separado como
telemetría no operativa; no se permite ninguna otra excepción de escritura.

El workflow manual `E2E Visual QA` recibe los mismos dos valores. Para una
secuencia completa, usar lotes consecutivos con el mismo SHA y evidencia
separada (`0/80`, `80/80`, `160/80`, `240/80`); no ejecutar los lotes contra
distintos candidatos ni interpretarlos como prueba de acciones mutantes.

## Validacion de textos y codificacion

Antes de cerrar cambios que toquen textos visibles, ayudas, plantillas de correo,
mensajes backend o documentacion operativa, buscar caracteres rotos por
codificacion. El objetivo es no publicar palabras con tildes rotas, secuencias
de doble codificacion o caracteres de reemplazo en pantallas del sistema.

```powershell
$badEncodingPattern = ([char]0xFFFD) + "|" + ([char]0x00C3) + "|" + ([char]0x00D2) + "|[A-Za-zÁÉÍÓÚáéíóúÑñ]\?[A-Za-zÁÉÍÓÚáéíóúÑñ]"
rg -n $badEncodingPattern web backend scripts documentos AGENTS.md -g "*.html" -g "*.js" -g "*.css" -g "*.go" -g "*.md" -g "*.txt" -g "*.ps1" -g "*.json" -g "*.yaml" -g "*.yml" -g "*.sql"
```

Revisar manualmente los resultados porque las URLs con query string pueden dar
falsos positivos por `?action=...`. Si se corrige un archivo, conservarlo en
UTF-8 y volver a ejecutar el barrido.

## Preflight

```powershell
.\scripts\profesional_preflight.ps1
.\scripts\profesional_preflight.ps1 -Full
.\scripts\profesional_preflight.ps1 -RequireMigrationAudit -MigrationBaseRef origin/main
```

Usar preflight antes de sincronizaciones o cambios grandes. Si falla, corregir la
causa concreta o dejar el riesgo documentado.

## rs

El usuario suele pedir `ejecuta rs`. El script canonico vive en `scripts`:

```powershell
.\scripts\rs.ps1
```

El drill de aplicación restaurada acepta el directorio padre de snapshots y
selecciona automáticamente el subdirectorio completo más reciente. Siempre
opera en PostgreSQL, red, puertos y almacenamiento efímeros:

```powershell
.\scripts\vps_p109_restore_app_validation.ps1 -AllowRemoteTarget
```

Para fijar un snapshot concreto, usar `-BackupDir` con la ruta remota del
snapshot que contiene `postgres_all.sql.gz`. Los errores remotos se limitan a
diagnósticos saneados `[ERROR]`/`[WARN]`; no agregar impresión de variables ni
del stderr completo.

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

Un `rs` real exige siempre el gate de migraciones, aunque no se use
`-FullPreflight`: valida los inventarios `Ensure`, compara las entradas
historicas del catalogo contra el ancestro comun de una rama de trabajo,
rechaza migraciones eliminadas o mutadas y ejecuta las pruebas Go enfocadas del
migrador. Despues de fusionar una PR, repite el gate contra el SHA que tenia
`main` antes de la integracion para comprobar que la fusion conservo sus
migraciones. `-SkipPreflight` solo
se admite con `-DryRun` o `-PreviewOnly`. Antes de reintentar tras un inventario
desactualizado, regenerar y revisar los artefactos:

Si `main` ya esta asignada a otro worktree, `rs` no intenta duplicarla ni
desmontarla. Exige que ese worktree este limpio, lo actualiza exclusivamente con
`fetch` y `merge --ff-only`, y ejecuta alli el preflight post-integracion y la
sincronizacion. Si contiene cambios locales, el flujo se detiene antes del VPS.

```powershell
& "C:\Users\ivanm\.cache\codex-runtimes\codex-primary-runtime\dependencies\node\bin\node.exe" tools\ensure_bootstrap_inventory.mjs
& "C:\Users\ivanm\.cache\codex-runtimes\codex-primary-runtime\dependencies\node\bin\node.exe" tools\runtime_ensure_inventory.mjs
```

Si `rs` parte de una rama distinta de `main`, primero publica esa rama, crea o
reutiliza una PR abierta hacia `main`, solicita auto-merge sin omitir checks y
espera la fusion. Solo entonces cambia a `main`, aplica fast-forward y exige
igualdad exacta con `origin/main` antes del VPS. Una PR fusionada antigua de la
misma rama no cuenta como abierta para commits nuevos. Para bloquear en lugar de
integrar automaticamente, usar `-NoIntegrateWorkBranch`; para otro nombre de
rama productiva, usar `-ProductionBranch`.

`-FullPreflight` hace bloqueante la bateria `go test ./...`. Si GitHub informa
que la PR esta `DIRTY`, `rs` termina de inmediato y exige reconciliacion manual;
no elige automaticamente entre cambios en conflicto ni espera inutilmente el
timeout de la PR.

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

Los certificados DIAN heredados bajo `web/uploads` deben migrarse al volumen
privado por empresa. Después de desplegar una imagen que incluya
`pcs-migrate-private-uploads`, ejecutar en el VPS:

```bash
cd /root/powerfulcontrolsystem
EMPRESA_ID=12 bash deploy/scripts/vps-migrate-private-dian.sh
```

El script restringe la reparación temporal a la carpeta de firma de esa
empresa, ejecuta primero una simulación y luego la migración confirmada de las
dos referencias DIAN. No debe reemplazarse por un `chown -R` general sobre los
uploads.

Antes de un cambio de migraciones, cola, outbox o bootstrap, verificar el
inventario y el contrato de catalogo sin tocar datos reales:

```powershell
Set-Location D:\powerfulcontrolsystem
$node = $env:PCS_NODE_PATH
if ([string]::IsNullOrWhiteSpace($node)) {
  $node = Join-Path $env:USERPROFILE ".cache\codex-runtimes\codex-primary-runtime\dependencies\node\bin\node.exe"
}
if (-not (Test-Path -LiteralPath $node)) { $node = (Get-Command node -ErrorAction Stop).Source }
& $node tools\ensure_bootstrap_inventory.mjs --check
& $node tools\migration_audit.mjs --strict
Set-Location backend
go test ./db ./internal/platform/worker ./internal/platform/outbox
go vet ./db ./internal/platform/worker ./internal/platform/outbox
go build ./cmd/pcs-migrate
go build ./cmd/pcs-worker
```

Antes del barrido visual o E2E del Plan 106, generar el inventario estático de
interfaz. No autentica ni hace clics; su salida enumera los controles que luego
deben recibir evidencia funcional, visual y de permisos:

```powershell
Set-Location D:\powerfulcontrolsystem
$node = 'C:\Users\ivanm\.cache\codex-runtimes\codex-primary-runtime\dependencies\node\bin\node.exe'
& $node tools\plan106_ui_inventory.mjs
```

El barrido autenticado `tools\qa_e2e_buttons.cjs` no tiene URL ni empresa por
defecto. Antes de ejecutarlo se deben declarar explícitamente el entorno aislado
y la empresa autorizada mediante `PCS_QA_BASE_URL` y `PCS_QA_EMPRESA_ID`; cargar
las credenciales solo mediante variables de sesión, nunca en scripts, comandos
guardados, documentos o reportes.

`pcs-migrate` es el unico rol que aplica las migraciones de plataforma. La API
y el worker solo verifican cola, outbox e idempotencia movil. Mantener
`PCS_RUNTIME_SCHEMA_BOOTSTRAP=1` hasta que el inventario legado se haya movido
por lotes, el staging pase con valor `0` y se documente el rollback.

En `production`, el rol `migrate` debe recibir esa variable de forma explicita;
si falta, el proceso termina sin ejecutar el bootstrap historico. El Compose
oficial usa el requisito `${PCS_RUNTIME_SCHEMA_BOOTSTRAP:?Defina PCS_RUNTIME_SCHEMA_BOOTSTRAP}`
para evitar despliegues ambiguos. `sync_to_vps` asegura el valor operativo
`1` en el archivo privado remoto solo cuando la clave no existe; no imprime ni
sobrescribe secretos del archivo.

En `DeploymentMode=docker`, el bootstrap remoto se ejecuta de forma limitada
para asegurar variables operativas necesarias para Compose. Conserva los
secretos privados existentes en el VPS y no los imprime ni los reemplaza.

No establecer `PCS_RUNTIME_SCHEMA_BOOTSTRAP=0` en una instalacion existente
hasta verificar el ledger de migraciones y los flujos de provisionamiento.

## Candidato inmutable por digest

El candidato de release se construye una sola vez mediante el workflow manual
`Immutable release candidate`. Debe recibir un SHA completo ya aprobado:

```powershell
gh workflow run release-candidate.yml -f commit_sha="$(git rev-parse HEAD)"
```

El workflow construye las cuatro imágenes PCS (`api`, `migrate`, `worker` y
`frontend`), escanea exactamente sus archivos con Trivy, genera SBOM CycloneDX,
publica en GHCR y entrega `release-images.env` con referencias
`repositorio@sha256`. Ese archivo no contiene secretos, pero debe asociarse al
SHA que lo produjo y no editarse manualmente.

Antes de promover, cargar las cuatro variables `PCS_*_IMAGE_DIGEST` desde el
artefacto y ejecutar:

```powershell
.\scripts\immutable_release_check.ps1
```

Staging promueve esas mismas imágenes sin reconstruir:

```bash
env \
  PCS_API_IMAGE_DIGEST='repositorio@sha256:...' \
  PCS_MIGRATE_IMAGE_DIGEST='repositorio@sha256:...' \
  PCS_WORKER_IMAGE_DIGEST='repositorio@sha256:...' \
  PCS_FRONTEND_IMAGE_DIGEST='repositorio@sha256:...' \
  /bin/bash deploy/scripts/vps-staging-digest-up.sh
```

No guardar tokens de GHCR en Git. Si el paquete es privado, autenticar Docker
en el VPS mediante un token de lectura obtenido del canal seguro.

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

En el propio VPS, el generador operativo admite tres alcances cerrados:

```bash
BACKUP_SCOPE=database bash deploy/scripts/vps-backup-operacion.sh
BACKUP_SCOPE=system bash deploy/scripts/vps-backup-operacion.sh
BACKUP_SCOPE=vps bash deploy/scripts/vps-backup-operacion.sh
```

`database` crea dumps lógicos; `system` añade proyecto y archivos persistentes
PCS; `vps` añade los volúmenes permitidos de los contenedores activos y el dump
lógico de Nextcloud cuando está disponible. No se archivan en caliente los
volúmenes físicos de PostgreSQL/MariaDB. Cada ejecución genera `MANIFEST.txt` y
`SHA256SUMS`. La retención local predeterminada es de dos días: la conservación
larga debe ocurrir fuera del VPS.

Para nube se usa `rclone` (Google Drive, OneDrive, Mega, Dropbox, Box, pCloud,
Backblaze B2, S3 y proveedores compatibles) o AWS CLI para S3. Configurar OAuth
fuera del repositorio, preferiblemente con un remoto `crypt`, y guardar solo la
referencia en `/etc/pcs-external-backup.env` con permisos `600`:

```text
EXTERNAL_BACKUP_TARGET=rclone
RCLONE_REMOTE=nombre_crypt:PCS/backups
BACKUP_SCOPE=vps
KEEP_LOCAL_DAYS=2
```

Después se valida manualmente y solo entonces se activa el cron:

```bash
set -a; . /etc/pcs-external-backup.env; set +a
bash deploy/scripts/vps-external-backup.sh
bash deploy/scripts/vps-install-external-backup-cron.sh
```

La subida no se considera correcta hasta que `rclone check --checksum` termina
sin diferencias. No activar el cron con destino `none` ni conservar tokens en
Git, logs o la base de datos.

Para arrancar la API y el migrador exactos contra un snapshot restaurado sin
tocar los servicios activos, ejecutar en la VPS:

```bash
PROJECT_DIR=/ruta/checkout \
SOURCE_ENV=/ruta/privada/.env.platform \
PCS_API_IMAGE_DIGEST=ghcr.io/organizacion/pcs-api@sha256:<64-hex> \
PCS_MIGRATE_IMAGE_DIGEST=ghcr.io/organizacion/pcs-migrate@sha256:<64-hex> \
P109_DRILL_ID=p109-restore-app-<sha-corto> \
bash deploy/scripts/vps-p109-restored-app-drill.sh
```

Las credenciales QA son opcionales y se pasan solo mediante variables de
sesión `P109_QA_EMAIL` y `P109_QA_PASSWORD`; nunca se escriben en el script ni
en la evidencia. Si se definen, el ensayo usa el login oficial y recorre CxP,
contabilidad, CxP/IA, DIAN y documentos. El script exige digests, usa red y
puerto loopback efímeros, bloquea acceso anónimo y limpia todos los recursos.
Para una inspección visual por túnel local puede definirse
`P109_HOLD_SECONDS`, con máximo de 900 segundos; al expirar se conserva la
limpieza automática.
Para comprobar dos réplicas de aplicación sobre la copia restaurada, definir
`P109_VERIFY_REPLICA=1`, un `P109_REPLICA_HOST_PORT` loopback distinto y las
credenciales QA por variables de sesión. El ensayo carga por A con CSRF,
descarga por B, compara SHA-256, retira A y repite readiness/descarga en B.
En ese modo también exige negativos de empresa cruzada, contenido HTML, tamaño
superior a 15 MiB y symlink, y comprueba que los rechazos no creen filas.
Para ensayar pérdida y recuperación conjunta de datos/archivos, añadir
`P109_VERIFY_COORDINATED_ROLLBACK=1`. Este modo exige réplicas y credenciales
QA: detiene ambas APIs efímeras, crea un checkpoint de las dos bases y del
volumen privado, elimina exclusivamente esas copias temporales, las restaura,
reinicia una réplica y vuelve a comprobar login, cinco dominios, fila y SHA-256.
Nunca se debe apuntar este modo a las bases o volúmenes activos.
Para auditar el snapshot antes del arranque, definir
`P109_VERIFY_PRIVATE_INVENTORY=1`. La prueba cruza el catálogo cerrado de chat,
buzón, DIAN, finanzas, grafología y soportes CxP/IA contra los archivos del
volumen; bloquea referencias faltantes, symlinks y rutas fuera de
`<categoria>/empresa_<id>/<archivo>`, y reporta sin borrar huérfanos o
referencias heredadas que deban migrarse.

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

Cuando se trabaja desde un worktree limpio que no contiene la configuración
privada local de despliegue, indicar el puerto SSH de forma explícita. No copiar
archivos privados ni secretos al worktree:

```powershell
.\scripts\rs.ps1 -VpsPort 49222
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
& $ssh @args $target "docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml logs --tail 160 backend"
```

Para revisar PostgreSQL por consola del contenedor:

```powershell
& $ssh @args $target "docker exec -i pcs-postgres sh -lc 'psql -U \"$POSTGRES_USER\" -d pcs_empresas -c \"select 1\"'"
```

Si se actualizan datos operativos en produccion, registrar el motivo y la
referencia minimizada en la PR o bitácora operativa externa y validar por API o
pantalla publicada.

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

Para ejecutar la batería sintética de impresiones con el navegador ya instalado,
definir `PCS_QA_CHROME_EXECUTABLE` con la ruta de Chrome for Testing antes de
llamar `tools/qa_print_formats.cjs` o `tools/qa_e2e_buttons.cjs`. La variable
evita descargar navegadores; el resultado no sustituye la prueba de impresora
física ni documentos reales.

El runner de botones localiza automáticamente el Playwright incluido por Codex.
Antes de una sesión real se puede validar solo el runtime, sin autenticarse ni
navegar:

```powershell
$env:PCS_QA_VALIDATE_RUNTIME = '1'
& $node tools\qa_e2e_buttons.cjs
Remove-Item Env:PCS_QA_VALIDATE_RUNTIME
```

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

## Validación de flujos con efectos reales

Usar los contratos de [pagos](gobernanza_tecnica/contratos/contrato_checkout_licencias_publico.md),
[fiscal](gobernanza_tecnica/contratos/contrato_facturacion_electronica_y_documentos_transaccionales.md)
y [domótica](domotica_raspberry_tunnel.md). No reutilizar clientes, documentos,
tarifas, folios ni identidad de ejemplos históricos como datos de prueba.
Una consulta documental no autoriza cobros, emisión, correo, GPIO ni despliegue.
Registrar alcance autorizado, fuente genuina, candidato, entorno y evidencia.

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

## Fuentes y aceptación de la revisión

[rs.ps1](../scripts/rs.ps1), [sync_to_vps.ps1](../scripts/sync_to_vps.ps1), [profesional_preflight.ps1](../scripts/profesional_preflight.ps1), [docs_catalog.mjs](../tools/docs_catalog.mjs).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).
