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

## Variables nuevas o modificadas

- `PCS_TRUSTED_PROXY_CIDRS`: CIDR de los proxies que pueden aportar `X-Forwarded-*`. Es obligatorio configurarlo correctamente antes de produccion.
- `CONFIG_ENC_KEY`: ahora requiere Base64 canonico de exactamente 32 bytes.
- `CONFIG_ENC_KEY_PREVIOUS`: lista temporal `id:base64` para descifrar datos durante una rotacion; no debe permanecer indefinidamente.
- `PCS_ENV=production` o `APP_ENV=production`: bloquea el arranque si falta una clave de cifrado valida.

## Cambios pendientes antes de desplegar

- Completar CSRF con token sincronizador en todas las mutaciones autenticadas y actualizar los clientes que no envien el encabezado.
- Inventariar, firmar e idempotentizar todos los webhooks y callbacks externos.
- Auditar cada handler multiempresa contra `RequireEmpresaAccess` y agregar pruebas cruzadas de lectura, escritura, exportacion y archivos.
- Reemplazar el WebRTC publico actual por autenticacion, tenant validado y autorizacion por sesion; no habilitarlo en produccion hasta entonces.
- Endurecer Docker: eliminar el montaje de Docker socket del backend, fijar tags/digests, ejecutar sin root, usar filesystem readonly y capability drop.
- Fijar acciones GitHub por SHA, agregar `gosec`, `govulncheck`, escaneo de secretos/SBOM y proteger la rama principal desde GitHub.

## Despliegue y rollback

1. Probar la rama en staging con una copia anonimizada y una clave de cifrado valida ya provisionada.
2. Configurar CIDR del reverse proxy y validar login, callback Google, recuperacion, logout, seleccion de empresa y pagos de prueba.
3. Respaldar la base y confirmar que la migracion de sesiones no deja tokens legibles. La primera ejecucion mantiene las cookies existentes validas.
4. Desplegar con ventana controlada y monitorear autenticacion, 401/403 y errores de descifrado.
5. Para rollback de aplicacion, volver al commit anterior. No revertir la migracion de tokens: los verificadores SHA-256 son deliberadamente unidireccionales y el rollback debe conservar el soporte de lectura hash.
