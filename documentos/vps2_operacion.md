# VPS2 operacion

Documento operativo para el servidor VPS2 local de pruebas.

## Conexion conocida

- Host SSH: `192.168.1.188`
- Puerto SSH: `22`
- Usuario Ubuntu: `admin1`
- VNC: `192.168.1.188:5901`
- Host key SSH verificada: `SHA256:QQmT0ZjCVNNxw7ICwV7FKwrzzzfWrOrtZ9zTrEGkwH0`

Las claves y contrasenas no se guardan en documentacion ni archivos versionados.
Para automatizar, usar `scripts/pcs_deployment.local.ps1` o variables de entorno
locales `PCS_VPS2_*`.

## Configuracion local privada

Copiar `scripts/pcs_deployment.local.ps1.example` a
`scripts/pcs_deployment.local.ps1` y ajustar solo en el archivo local ignorado por
Git:

```powershell
$script:PcsVps2Host = "192.168.1.188"
$script:PcsVps2User = "admin1"
$script:PcsVps2RemotePath = "/home/admin1/powerfulcontrolsystem"
$script:PcsVps2Port = 22
$script:PcsVps2HostKey = "SHA256:QQmT0ZjCVNNxw7ICwV7FKwrzzzfWrOrtZ9zTrEGkwH0"
# $script:PcsVps2IdentityFile = "ruta a clave privada .ppk"
# $script:PcsVps2Password = "guardar solo en archivo local privado si es inevitable"
# $script:PcsVps2RepoUrl = $script:PcsGitRemoteUrl
```

## Sincronizacion y mantenimiento

Comando principal:

```powershell
.\scripts\sync_to_vps2.ps1
```

Comandos utiles:

```powershell
.\scripts\sync_to_vps2.ps1 -SkipDeploy
.\scripts\sync_to_vps2.ps1 -SkipDisableGui -SkipNextcloud
.\scripts\sync_to_vps2.ps1 -RestartDockerStack:$false
```

El script:

- valida que SSH responda;
- actualiza el repositorio remoto con `git pull --ff-only` cuando existe;
- clona el repositorio si no existe y `PcsVps2RepoUrl` esta configurado;
- reconstruye el stack Docker si encuentra un archivo compose compatible;
- deja el VPS2 en `multi-user.target` para no abrir modo grafico al reiniciar;
- aplica `restart unless-stopped` a contenedores Nextcloud detectados.

## Estado aplicado el 2026-07-06

- SSH respondio en `192.168.1.188:22`.
- VNC respondio en `192.168.1.188:5901`.
- El host remoto se identifico como `vps2`.
- El modo grafico quedo deshabilitado por defecto con `multi-user.target`.
- Se detectaron contenedores `nextcloud-app`, `nextcloud-redis` y
  `nextcloud-db`; quedaron activos y con reinicio automatico
  `unless-stopped`.
- No se encontro repositorio en `/home/admin1/powerfulcontrolsystem`; para el
  primer despliegue se debe configurar `PcsVps2RepoUrl` o clonar el proyecto en
  esa ruta.
