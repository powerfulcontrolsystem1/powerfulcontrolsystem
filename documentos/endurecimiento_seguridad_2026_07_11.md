# Endurecimiento de seguridad 2026-07-11

Estado: preparado en la rama `security/full-hardening`; no desplegado.

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
- El workflow `Professional CI` se ejecuta tambien en ramas `security/**`,
  ademas de pull requests, ramas principales y ejecucion manual.

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

## Cambios pendientes antes de desplegar

- Validar en staging el origen CSRF para los clientes heredados y completar tokens sincronizadores solo donde una integracion no pueda enviar Origin/Referer.
- Inventariar, firmar e idempotentizar todos los webhooks y callbacks externos.
- Auditar cada handler multiempresa contra `RequireEmpresaAccess` y agregar pruebas cruzadas de lectura, escritura, exportacion y archivos.
- Verificar en staging los permisos de volumen del backend no root y los flujos administrativos que antes dependian del socket Docker, ahora retirado del contenedor de negocio.
- Configurar desde GitHub la proteccion de rama, el escaneo de secretos/SBOM y la revision requerida; son controles de plataforma que no se pueden imponer desde archivos versionados.

## Despliegue y rollback

1. Probar la rama en staging con una copia anonimizada y una clave de cifrado valida ya provisionada.
2. Configurar CIDR del reverse proxy y validar login, callback Google, recuperacion, logout, seleccion de empresa y pagos de prueba.
3. Respaldar la base y confirmar que la migracion de sesiones no deja tokens legibles. La primera ejecucion mantiene las cookies existentes validas.
4. Desplegar con ventana controlada y monitorear autenticacion, 401/403 y errores de descifrado.
5. Para rollback de aplicacion, volver al commit anterior. No revertir la migracion de tokens: los verificadores SHA-256 son deliberadamente unidireccionales y el rollback debe conservar el soporte de lectura hash.
