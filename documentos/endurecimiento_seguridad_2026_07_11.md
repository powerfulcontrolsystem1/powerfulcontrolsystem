# Endurecimiento de seguridad 2026-07-11

Estado: validación final en `security/full-hardening-clean-20260712`; no desplegado.

## Linea base

- `go test ./...` paso antes de los cambios.
- No se modificaron servidores, contenedores, datos empresariales, secretos ni configuraciones activas.

## Corregido en la rama

- OAuth Google usa state aleatorio de 32 bytes, cookie HttpOnly de 10 minutos, comparacion constante, consumo unico y PKCE S256. Rechaza correo no verificado y ya no crea sesiones con datos previsibles.
- Sesiones nuevas almacenan solo un verificador SHA-256 versionado. Al iniciar, la migracion idempotente reemplaza tokens heredados sin texto plano.
- El middleware no eleva ni reescribe roles a partir de una direccion de correo. El rol persistido es la fuente de autorizacion.
- Cabeceras de seguridad, HSTS condicionado a HTTPS, no-store en rutas sensibles, limites del servidor HTTP y confianza explicita en proxies.
- Cifrado exige `CONFIG_ENC_KEY` Base64 canonico de 32 bytes, usa envoltura versionada AES-GCM con AAD y permite claves anteriores identificadas durante rotacion.
- Los logs dejan de registrar query strings, user-agent y tenant no validado; neutralizan saltos de linea y se crean con permisos restringidos.
- Los secretos TOTP nuevos se cifran con AES-GCM en un dominio criptografico
  exclusivo `totp`, derivado de la clave maestra. El valor persistido incluye
  version, finalidad e identificador de clave; no se guarda ni se vuelve a
  exponer el secreto al terminar la activacion.
- La migracion de TOTP heredado acepta modo simulacion y cifra los valores
  antiguos de manera idempotente. La activacion y desactivacion exige
  reautenticacion, rota la sesion y revoca las restantes sesiones activas.
- Los codigos de recuperacion TOTP se generan con 32 bytes aleatorios, se
  almacenan solo como SHA-256, son de uso unico y se invalidan al regenerarse.
- Los nuevos tokens administrativos de recuperacion de contraseña se guardan
  solo como verificadores SHA-256; la migracion conserva la validez de los
  enlaces heredados sin retener su texto plano.
- Las invitaciones, confirmaciones y recuperaciones de usuarios empresariales
  usan el mismo verificador SHA-256 y se migran en el arranque sin exponer los
  valores originales. Las consultas siguen delimitadas por `empresa_id`.
- La cache de autenticacion tiene invalidacion explicita por token y por
  administrador, ademas del TTL cero, para que logout, reset y cambio de 2FA
  surtan efecto inmediatamente.
- Las mutaciones autenticadas por `session_token` requieren ahora el token
  sincronizador `pcs_csrf` en la cabecera `X-CSRF-Token`, ademas de Origin o
  Referer de esquema, host y puerto exactos. `menu.js` lo incorpora a los
  `fetch` same-origin; los clientes Bearer y los webhooks firmados quedan fuera
  de este modelo porque no dependen de la cookie de sesion.
- El workflow `Professional CI` se ejecuta tambien en ramas `security/**`,
  ademas de pull requests, ramas principales y ejecucion manual.
- La CSP estricta se publica inicialmente mediante
  `Content-Security-Policy-Report-Only`: elimina destinos comodin de conexiones
  y agrega `form-action 'self'` sin bloquear aun el frontend heredado.
- El workflow y el modulo usan Go 1.25 para que `govulncheck` analice una
  biblioteca estandar con los parches vigentes; `pgx/v5` se actualiza a 5.9.2,
  que corrige la vulnerabilidad de confusion de placeholders SQL reportada por
  el escaner.
- La version del runner queda fijada en Go 1.25.12. El escaneo previo identifico
  GO-2026-5856 y GO-2026-4970 en Go 1.25.11, junto con vulnerabilidades
  alcanzables en `x/net`, `x/crypto` y `x/sys`; se actualizan respectivamente
  a 0.55.0, 0.52.0 y 0.45.0, las versiones corregidas reportadas por
  `govulncheck`.
- El módulo y la imagen de compilación de producción fijan también el toolchain
  Go 1.25.12. Con esa versión, `govulncheck v1.1.4` informa cero
  vulnerabilidades alcanzables; la advertencia de módulo sobre `openpgp` no es
  alcanzable por el código del proyecto.
- El CI no despliega por eventos `push`: cualquier despliegue debe vivir en un
  workflow manual y protegido. Las verificaciones de vulnerabilidades y analisis
  estatico se ejecutan como pasos independientes, junto con `go mod verify` y
  `git diff --check`.

## Variables nuevas o modificadas

- `PCS_TRUSTED_PROXY_CIDRS`: CIDR de los proxies que pueden aportar `X-Forwarded-*`. Es obligatorio configurarlo correctamente antes de produccion.
- `CONFIG_ENC_KEY`: ahora requiere Base64 canonico de exactamente 32 bytes.
- `CONFIG_ENC_KEY_PREVIOUS`: lista temporal `id:base64` para descifrar datos durante una rotacion; no debe permanecer indefinidamente.
- `CONFIG_ENC_KEY_ID`: identificador estable de la clave activa, por ejemplo
  `key-2026-07`. Debe cambiarse al rotar `CONFIG_ENC_KEY`; los valores nuevos
  quedan etiquetados con este id y las claves anteriores se declaran en
  `CONFIG_ENC_KEY_PREVIOUS`.
- `PCS_TOTP_MIGRATION_DRY_RUN=true`: cuenta secretos TOTP y tokens de
  recuperacion administrativos heredados que se migrarian al arrancar, sin
  modificar filas. Debe ejecutarse primero en staging y retirarse para la
  migracion real.
- `PCS_ENV=production` o `APP_ENV=production`: bloquea el arranque si falta una clave de cifrado valida.
- `PCS_PRIVATE_STORAGE_DIR`: raíz no pública para documentos empresariales;
  producción no inicia si falta y Nginx no monta este volumen.
- `SESSION_TIMEOUT`, `MAX_REQUEST_BODY_BYTES`, `HTTP_READ_TIMEOUT`,
  `HTTP_WRITE_TIMEOUT` y `HTTP_IDLE_TIMEOUT`: límites obligatorios de runtime.

## Evidencia consolidada

| Hallazgo | Severidad | Evidencia inicial | Corrección | Prueba | Resultado | Riesgo residual | Estado |
|---|---|---|---|---|---|---|---|
| Secretos TOTP legibles | Alta | columna heredada sin envoltura | AES-GCM versionado, finalidad `totp`, rotación y migración simulable | pruebas de cifrado, rotación, activación y recuperación | pasa | migración real solo en staging | corregido |
| Tokens y sesiones reutilizables | Alta | valores heredados y caché por TTL | verificadores SHA-256, consumo único, revocación e invalidación inmediata | pruebas de vencimiento, reemplazo y revocación | pasa | ninguno conocido | corregido |
| CSRF en cookies | Alta | mutaciones sin sincronizador global | token separado, Origin/Referer exactos y transición frontend | suite CSRF y login | pasa | CSP continúa en report-only durante transición | mitigado |
| WebRTC sin credencial fuerte | Crítica | señalización por empresa/rol del cliente | token+nonce de uso único, permiso, Origin, expiración, límites y revocación | `soporte_remoto_webrtc_test.go` | pasa | validación real solo en staging | corregido |
| Adjuntos empresariales bajo raíz web | Alta | chat, buzón, finanzas, grafología, capturas y firma DIAN usaban rutas de `uploads` | almacenamiento por empresa en `PCS_PRIVATE_STORAGE_DIR`, nombres aleatorios, MIME real, descarga autenticada y bloqueo Nginx de rutas heredadas | `private_files_test.go`, pruebas DIAN y suite Go | pasa | ejecutar primero la migración simulada en staging | corregido |
| Aislamiento multiempresa manipulable | Crítica | múltiples fuentes podían aportar `empresa_id` | middleware central valida sesión, relación, empresa, rol y permiso; el contexto conserva el único tenant autorizado | `empresa_permisos_tenant_test.go` y pruebas de archivos/soporte | pasa | ampliar casos operativos al agregar nuevos módulos | corregido |
| Operaciones dinámicas sin evidencia | Media | `gosec` G202/G204/G304 | revisión por sitio, parámetros, raíces controladas y eliminación de shells | `gosec` local y job `Preflight, audits and Docker config` 29215120284 | pasa | ninguno aceptado sin justificación concreta | corregido |
| Binarios compilados vulnerables | Crítica | Trivy detectó ejecutables Go antiguos dentro de `project_export` y Grype encontró `ngrok.exe` compilado con Go 1.16.2 y dependencias de 2020 | retiro de cuatro binarios generados/locales, reglas Git/Docker y reconstrucción desde fuente vigente | Trivy de imagen y Grype sobre SBOM en GitHub Actions | pendiente del run posterior al commit | bases de vulnerabilidades cambian | en validación |
| Base Nginx obsoleta | Alta | imagen 1.27 incluía bibliotecas Alpine vulnerables y el primer rebuild conservó `c-ares 1.34.6-r0` | Nginx unprivileged 1.30.3 sobre Alpine 3.23 y `c-ares 1.34.8-r0`, versiones fijas | construcción y Trivy de imagen | pendiente del run posterior al commit | fijar digest tras validar la arquitectura de producción | en validación |
| Cadena de suministro incompleta | Alta | CI sin secretos/IaC/SBOM/imágenes | jobs separados con herramientas versionadas, gates explícitos, SARIF y artefactos | GitHub Actions | escaneo de secretos, filesystem e IaC pasan; imágenes en revalidación | bases de vulnerabilidades cambian | en validación |

## Contenedores y excepciones técnicas

- Backend, frontend y voz ejecutan como usuarios no root, con
  `no-new-privileges`, capacidades retiradas y filesystem de solo lectura donde
  es compatible. Frontend usa Nginx unprivileged en el puerto interno 8080.
- Redis de Mailu ejecuta como `redis`, con volumen de datos separado, raíz de
  solo lectura y healthcheck.
- PostgreSQL conserva el entrypoint oficial, que requiere privilegios iniciales
  para preparar un volumen nuevo; queda aislado en red interna, sin puerto
  público y con healthcheck. Forzar `USER` o `cap_drop: ALL` rompería la
  inicialización oficial del volumen.
- Mailu, OnlyOffice y RustDesk conservan sus usuarios/entrypoints soportados por
  los proveedores. Se mitigan con perfiles opt-in, versiones fijas, redes,
  volúmenes mínimos y puertos explícitos. Cambiar su usuario sin soporte del
  proveedor se valida primero en staging.

## Dependencia Excel

- `github.com/xuri/excelize/v2` se actualiza de 2.8.1 a 2.11.0 por
  `CVE-2026-54063`. La alternativa en Go puro sería reimplementar ZIP, XML,
  estilos, fórmulas y compatibilidad XLSX, con mayor superficie de error y sin
  beneficio de seguridad. El cambio afecta `backend/go.mod`, `backend/go.sum`
  y los flujos existentes de importación/exportación Excel.

## Validaciones obligatorias de staging

Estas actividades no son correcciones de código pendientes y no se ejecutan en
esta tarea porque requieren entorno aislado, copia anonimizada o controles de
GitHub:

- Ejecutar `go run ./tools/migrate_private_uploads --web-root=<RAIZ_WEB>` desde
  `backend` para obtener el inventario sin modificar archivos ni base. Aplicar
  solo después del respaldo de staging con `--apply
  --confirm=MIGRATE_PRIVATE_UPLOADS`; la actualización exige coincidencia de
  `id`, `empresa_id` y referencia anterior antes de retirar el archivo legado.
- Ejecutar las migraciones TOTP/tokens primero en modo simulación y verificar
  conteos antes de aplicar.
- Validar login, CSRF, pagos sandbox, descarga privada, WebRTC y soporte remoto
  con datos ficticios y orígenes reales de staging.
- Verificar UID/permisos de los volúmenes para backend, frontend y voz no root.
- Activar protección de `main` y checks obligatorios después de fusionar un PR
  completamente verde.

## Despliegue y rollback

1. Probar la rama en staging con una copia anonimizada y una clave de cifrado valida ya provisionada.
2. Configurar CIDR del reverse proxy y validar login, callback Google, recuperacion, logout, seleccion de empresa y pagos de prueba.
3. Respaldar la base y confirmar que la migracion de sesiones no deja tokens legibles. La primera ejecucion mantiene las cookies existentes validas.
4. Desplegar con ventana controlada y monitorear autenticacion, 401/403 y errores de descifrado.
5. Para rollback de aplicacion, volver al commit anterior. No revertir la migracion de tokens: los verificadores SHA-256 son deliberadamente unidireccionales y el rollback debe conservar el soporte de lectura hash.
