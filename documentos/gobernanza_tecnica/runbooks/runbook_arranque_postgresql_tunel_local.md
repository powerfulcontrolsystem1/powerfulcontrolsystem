# Runbook: arranque backend PostgreSQL por tunel local

Fecha: 2026-04-18
Estado: vigente

## Sintomas cubiertos

- `go run .` no logra levantar el backend cuando `DB_VPS_TUNNEL_ENABLED=1`.
- el backend intenta conectarse a `127.0.0.1:5432` aunque el tunel local usa otro puerto.
- `scripts/iniciar_servidor.ps1` falla antes de abrir el tunel o deja el tunel sin DSN reescrito.
- las pruebas o verificaciones manuales funcionan con el helper temporal pero el arranque normal falla.

## Alcance

Aplica al arranque local del backend cuando la base canonica productiva es PostgreSQL del VPS y se accede mediante tunel SSH local.

## Fuentes de evidencia

- `backend/.env.local`
- variables de entorno del proceso: `DB_DIALECT`, `DB_EMPRESAS_DSN`, `DB_SUPERADMIN_DSN`, `DB_VPS_TUNNEL_ENABLED`, `DB_VPS_LOCAL_PORT`, `DB_VPS_SSH_HOST`, `DB_VPS_SSH_USER`, `DB_VPS_REMOTE_HOST`, `DB_VPS_REMOTE_PORT`
- `backend/main.go` en la funcion `resolveRuntimePostgresDSN`
- `scripts/iniciar_servidor.ps1`
- salida del backend al arrancar y errores de conexion a PostgreSQL

## Verificaciones iniciales

1. confirmar que el proyecto este operando con `DB_DIALECT=postgres`.
2. verificar que `DB_EMPRESAS_DSN` y `DB_SUPERADMIN_DSN` existan y apunten inicialmente a `127.0.0.1` o `localhost` si dependen del tunel.
3. confirmar si `DB_VPS_TUNNEL_ENABLED=1` y si `DB_VPS_LOCAL_PORT` contiene el puerto real esperado.
4. validar que existan `DB_VPS_SSH_HOST` y `DB_VPS_SSH_USER` cuando el tunel se abre desde `scripts/iniciar_servidor.ps1`.
5. comprobar si ya existe un listener local en el puerto del tunel antes de culpar al backend.

## Causas probables

- `DB_VPS_TUNNEL_ENABLED` no esta activado en el proceso real del backend.
- `DB_VPS_LOCAL_PORT` no coincide con el puerto donde el tunel quedo escuchando.
- el DSN apunta a un host no reescribible o a `127.0.0.1:5432` mientras el tunel real usa otro puerto.
- el tunel SSH no se abrio, quedo en otro puerto o fallo por credenciales SSH incompletas.
- el backend arranco sin pasar por `scripts/iniciar_servidor.ps1` y sin heredar variables necesarias del entorno.

## Acciones de recuperacion

1. revisar `backend/.env.local` y confirmar `DB_DIALECT=postgres`, `DB_VPS_TUNNEL_ENABLED=1` y `DB_VPS_LOCAL_PORT` correcto.
2. si el tunel se gestiona por script, ejecutar el flujo que lo abre desde `scripts/iniciar_servidor.ps1` y verificar que anuncie `Tunel DB iniciado` sobre el puerto local esperado.
3. confirmar que `resolveRuntimePostgresDSN` del backend reescribe solo DSN cuyo host sea `127.0.0.1` o `localhost`; si el DSN apunta a otro host, no habra reescritura.
4. validar que el proceso del backend haya heredado las mismas variables del entorno que usó el script para abrir el tunel.
5. si el helper de verificacion temporal conecta pero `go run .` falla, comparar el puerto reescrito por el helper con el puerto que realmente ve el backend en su entorno de proceso.
6. si el problema es SSH, corregir primero `DB_VPS_SSH_HOST`, `DB_VPS_SSH_USER`, llave privada o puerto remoto antes de reintentar el arranque.
7. una vez restablecido el tunel, volver a iniciar el backend y confirmar que las dos bases `pcs_empresas` y `pcs_superadministrador` abren sin error.

## Validacion posterior

- el backend arranca sin error de conexion a PostgreSQL.
- `resolveRuntimePostgresDSN` aplica el puerto de `DB_VPS_LOCAL_PORT` cuando corresponde.
- el mismo `DB_VPS_LOCAL_PORT` funciona tanto para helpers de verificacion como para `go run .` o el arranque por script.
- las operaciones de lectura simples contra ambas bases responden sin reconfiguracion manual adicional del DSN.

## Contratos relacionados

- `documentos/gobernanza_tecnica/contratos/contrato_autenticacion_administrativa_y_usuarios_empresa.md`

## ADRs relacionados

- `ADR-0002-postgresql-runtime-canonico-vps.md`