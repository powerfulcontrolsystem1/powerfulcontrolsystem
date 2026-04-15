# Sincronizar el proyecto con un VPS (rsync + SSH)

Archivos añadidos:

- `scripts/sync_to_vps.sh` — script Bash para sincronizar el repo local con el VPS usando `rsync` sobre SSH. Soporta opciones para ruta local, host, usuario, puerto, identidad SSH y `--dry-run`.
- `scripts/sync_to_vps.ps1` — wrapper PowerShell que usa `rsync` nativo si está disponible o invoca `scripts/sync_to_vps.sh` vía WSL (Windows).
- `scripts/instalar_clave_publica_vps.ps1` — instala automáticamente en el VPS una clave pública en formato PuTTYgen (RFC4716) dentro de `~/.ssh/authorized_keys`.

Instalar clave pública (1 comando)

Antes de sincronizar por primera vez, ejecuta:

```powershell
.\scripts\instalar_clave_publica_vps.ps1
```

El script:
- Conecta por SSH a `root@2.24.197.58` (puerto 22 por defecto).
- Convierte tu clave pública de PuTTYgen a formato OpenSSH en el VPS.
- La agrega en `~/.ssh/authorized_keys` si no existe.
- No duplica la clave si ya estaba instalada.

Opciones útiles:

```powershell
# Solo mostrar lo que haría (sin ejecutar conexión)
.\scripts\instalar_clave_publica_vps.ps1 -PreviewOnly

# Cambiar host/usuario/puerto
.\scripts\instalar_clave_publica_vps.ps1 -RemoteHost 2.24.197.58 -User root -Port 22

# Usar un archivo de clave pública diferente
.\scripts\instalar_clave_publica_vps.ps1 -PublicKeyFile C:\ruta\mi_clave_publica.putty
```

Antes de usar
1. Configure acceso SSH sin contraseña (clave pública) desde su máquina local al VPS (recomendado):

```bash
ssh-copy-id -i ~/.ssh/id_rsa.pub root@2.24.197.58
```

si `ssh-copy-id` no está disponible, copie el contenido de `~/.ssh/id_rsa.pub` a `/root/.ssh/authorized_keys` en el VPS.

2. Pruebe conexión SSH:

```bash
ssh -i ~/.ssh/id_rsa -p 22 root@2.24.197.58
```

Uso recomendado (pruebas primero)

- Ejecutar en modo `dry-run` para ver qué cambios se harían sin aplicar nada:

```bash
chmod +x scripts/sync_to_vps.sh
./scripts/sync_to_vps.sh --dry-run
```

- Ejecución real (por defecto sincroniza la raíz del repositorio actual al directorio `/root/powerfulcontrolsystem` en el VPS):

```bash
./scripts/sync_to_vps.sh --host 2.24.197.58 --user root --remote /root/powerfulcontrolsystem
```

Opciones útiles

- `--local PATH` : cambiar carpeta local a sincronizar
- `--host HOST`  : host remoto (por defecto: 2.24.197.58)
- `--user USER`  : usuario SSH remoto (por defecto: root)
- `--remote PATH`: ruta destino en el VPS
- `--port PORT`  : puerto SSH
- `--identity FILE`: clave privada SSH

Windows (PowerShell)

El wrapper funciona en dos modos:

- Modo WSL: usa `sync_to_vps.sh` + `rsync` dentro de una distro Linux.
- Fallback sin WSL: empaqueta el proyecto en `.tar` (con exclusiones) y selecciona transporte según clave:
	- Clave OpenSSH (por ejemplo `id_rsa`): usa `ssh.exe` + `scp.exe`.
	- Clave PuTTY `.ppk`: usa `plink.exe` + `pscp.exe`.
- En fallback, al finalizar se aplica `chmod +x` al binario configurado en `BuildOutput`.
- Al finalizar la sincronización (WSL y fallback), el script hace redeploy remoto del backend mediante `systemd`:
	- Instala o actualiza una unidad `systemd` derivada del directorio remoto (por defecto `powerfulcontrolsystem.service`).
	- Detiene listeners huérfanos que sigan ocupando `SERVER_PORT` fuera de `systemd` antes de iniciar la nueva versión, registrando `DEPLOY_WARN:PORT_*` con PID, ejecutable y comando cuando detecta procesos manuales o binarios borrados.
	- Reinicia la nueva versión con `Restart=always` y `systemctl enable`, dejando autoarranque tras reinicio del VPS.
	- Conserva `backend/server.log` y `backend/server.err` como destino de salida del servicio.
	- Ejecuta healthcheck HTTP en `SERVER_PORT` y, si falla, devuelve `systemctl status`, `journalctl` y colas recientes de logs.
- El flujo ahora muestra mensajes por etapa y errores accionables (`BOOTSTRAP_*`, `DEPLOY_*`) para distinguir con claridad si falló red, autenticación, variables de entorno, instalación base del VPS o arranque del servicio.
- Por defecto abre automáticamente la URL pública del despliegue (`http://<host>:<puerto>/`) al terminar sin errores (`-OpenPublicUrlAfterDeploy:$true`).

Si no hay distribuciones WSL instaladas, el script cambia automáticamente a fallback nativo de Windows.

Si no se especifica `-IdentityFile`, el script usa este orden automático:

1. `D:\powerfulcontrolsystem\clave privada ssh.ppk` (si existe en la raíz del proyecto).
2. `$env:USERPROFILE\.ssh\id_rsa`.

Por defecto, antes de sincronizar, el script compila en local un binario Linux de Go:

- Working dir: `backend`
- Package: `.`
- Output: `backend/bin/server_linux_amd64`
- Entorno: `GOOS=linux`, `GOARCH=amd64`, `CGO_ENABLED=0`

Además, al terminar la sincronización (modo WSL y fallback), ejecuta bootstrap remoto para servidor nuevo:

- Detecta el gestor de paquetes remoto (`apt-get`, `dnf`, `yum`, `apk` o `zypper`) e intenta instalar dependencias base del VPS (`ca-certificates`, `curl`, `wget`, `procps`/`procps-ng`, `lsof`) cuando el usuario remoto tiene `root` o `sudo -n`.
- Garantiza `backend/.env.local`, actualiza `SERVER_PORT` en cada despliegue y valida que el puerto enviado sea numérico y válido.
- Conserva configuración DB remota existente y permite inyectar/actualizar `DB_DIALECT`, `DB_EMPRESAS_DSN` y `DB_SUPERADMIN_DSN`.
- Si el modo efectivo queda en `postgres` y falta alguno de los DSN, el bootstrap falla con `BOOTSTRAP_ERROR:POSTGRES_MISSING_DSN` antes del redeploy (evita tumbar el proceso por arranque inválido).
- Reporta estado de variables críticas (`DB_DIALECT`, `DB_SUPERADMIN_DSN`, `DB_EMPRESAS_DSN`, `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `CONFIG_ENC_KEY`) y emite sugerencias explícitas cuando falta cifrado (`CONFIG_ENC_KEY`) o privilegios para preparar el VPS.
- La sincronización excluye por defecto `backend/.env.local` para no sobrescribir secretos/DSN existentes del VPS.

Validaciones operativas añadidas:

- `sync_to_vps.sh` valida localmente `REMOTE_PORT`, `SERVER_PORT` y `HEALTH_TIMEOUT` antes de abrir conexión SSH.
- Si el usuario remoto no tiene `root` ni `sudo -n`, el bootstrap no queda silencioso: emite advertencia y explica por qué no pudo instalar dependencias o registrar el servicio.
- Si el servicio no arranca, el redeploy devuelve hint + `systemctl status` + `journalctl` + colas de `backend/server.log` y `backend/server.err`.
- Si encuentra procesos fuera de `systemd` ocupando el puerto del backend, intenta detenerlos antes del arranque y aborta con `PORT_STILL_BUSY` si no logra liberar el listener.

`sync_to_vps.ps1` también intenta completar variables DB desde el entorno local (`DB_*`, `PCS_DB_*`) y desde `backend/.env.local` local cuando no se pasan por parámetro.

```powershell
# dry run desde PowerShell
.\scripts\sync_to_vps.ps1 -DryRun

# preview del comando interno sin ejecutar conexión remota
.\scripts\sync_to_vps.ps1 -PreviewOnly

# ejecutar sincronización real
.\scripts\sync_to_vps.ps1 -RemoteHost 2.24.197.58 -RemoteUser root -RemotePath /root/powerfulcontrolsystem

# usar clave PuTTY .ppk explícita (sin WSL)
.\scripts\sync_to_vps.ps1 -RemoteHost 2.24.197.58 -IdentityFile "D:\powerfulcontrolsystem\clave privada ssh.ppk"

# reforzar red inestable con reintentos automáticos
.\scripts\sync_to_vps.ps1 -RemoteHost 2.24.197.58 -IdentityFile "D:\powerfulcontrolsystem\clave privada ssh.ppk" -RetryCount 3

# desactivar auto-instalación de dependencias (por defecto está activa)
.\scripts\sync_to_vps.ps1 -RemoteHost 2.24.197.58 -IdentityFile "D:\powerfulcontrolsystem\clave privada ssh.ppk" -AutoInstallDependencies:$false

# compilar solo binario Linux local (sin sincronizar)
.\scripts\sync_to_vps.ps1 -BuildOnly -LocalPath "D:\powerfulcontrolsystem"

# omitir compilación y solo sincronizar
.\scripts\sync_to_vps.ps1 -SkipBuild -RemoteHost 2.24.197.58 -IdentityFile "D:\powerfulcontrolsystem\clave privada ssh.ppk"

# personalizar build Linux
.\scripts\sync_to_vps.ps1 -BuildWorkingDir backend -BuildPackage . -BuildOutput backend/bin/server_linux_amd64 -BuildGoOS linux -BuildGoArch amd64 -BuildCgoEnabled 0

# bootstrap remoto desactivado (si no lo quieres en una corrida específica)
.\scripts\sync_to_vps.ps1 -BootstrapServer:$false

# desactivar reinicio remoto automático del backend (solo sincronizar archivos)
.\scripts\sync_to_vps.ps1 -RestartRemoteServer:$false

# desactivar apertura automática de la URL pública al finalizar
.\scripts\sync_to_vps.ps1 -OpenPublicUrlAfterDeploy:$false

# personalizar binario/logs/timeout del redeploy remoto
.\scripts\sync_to_vps.ps1 -RemoteBinaryPath "backend/bin/server_linux_amd64" -RemoteStdoutLogPath "backend/server.log" -RemoteStderrLogPath "backend/server.err" -RestartHealthTimeoutSeconds 60

# configurar Google OAuth desde el deploy (opcional)
.\scripts\sync_to_vps.ps1 -GoogleClientId "TU_CLIENT_ID" -GoogleClientSecret "TU_CLIENT_SECRET"

# forzar runtime PostgreSQL en VPS e inyectar DSN (recomendado para migración)
.\scripts\sync_to_vps.ps1 -DbDialect "postgres" -DbEmpresasDsn "postgres://USUARIO:PASS@HOST:5432/pcs_empresas?sslmode=disable" -DbSuperadminDsn "postgres://USUARIO:PASS@HOST:5432/pcs_superadministrador?sslmode=disable"
```

Nota sobre `-DryRun` en fallback PuTTY:

- El script genera un paquete temporal y muestra el listado de archivos que se transferirían (sin cambiar el VPS).
- Se excluyen por defecto `.git`, `node_modules`, `logs`, `test_runs`, `*.db`, `*.exe`, `backend/.env.local`, `*.ppk`, `*.pem`, `*.key`.
- Ante errores de red tipo timeout, el script reintenta automáticamente (`-RetryCount`) y muestra diagnósticos claros por etapa.

Advertencias y buenas prácticas

- El script usa `rsync --delete` por defecto: los archivos que existan en el remoto y no en local serán eliminados. Use `--dry-run` antes de ejecutar para evitar borrados accidentales.
- Ajuste la lista `EXCLUDES` dentro de `scripts/sync_to_vps.sh` para evitar subir archivos grandes, bases de datos locales, binarios o temporales.
- Asegúrese de que la clave SSH tiene permisos seguros (`chmod 600 ~/.ssh/id_rsa`) y que el VPS está configurado para aceptar su clave.

Soporte y personalización

Si quieres que el script haga además: instalar dependencias remotas, reiniciar servicios, o desplegar en un path específico (por ejemplo `/var/www/` y ajustar permisos), dime qué pasos quieres automatizar y lo extiendo.

Nota operativa actual:

- El backend del VPS ya no debe levantarse con `nohup` manual. La ruta soportada es `scripts/sync_to_vps.ps1` o `scripts/sync_to_vps.sh`, que dejan el servicio persistente en `systemd` y lo reinician solo durante el despliegue.
- Si un despliegue anterior dejó un binario huérfano escuchando en `:8080`, el siguiente `sync_to_vps` lo limpiará automáticamente antes de volver a arrancar el servicio.
