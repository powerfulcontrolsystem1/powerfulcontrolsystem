# Email corporativo iRedMail

## Objetivo

Permitir que cada empresa tenga un correo corporativo propio bajo el dominio
configurado, por ejemplo `motel.calipso@powerfulcontrolsystem.com`, con
activacion controlada desde super administrador.

## Flujo operativo

1. El super administrador entra a `web/super/email_corporativo.html`.
2. Activa o desactiva el modulo global.
3. Define dominio, URL de webmail, cuota y modo de provision.
4. Al crear una empresa desde `/super/api/empresas`, el sistema genera un correo
   unico basado en el nombre de la empresa.
5. Si el correo ya existe, se usa un sufijo numerico: `empresa2@dominio`.
6. Si iRedAdmin-Pro API esta configurado, se intenta crear el buzon.
7. Si no esta configurado, el correo queda pendiente para provision manual.
8. En `administrar_empresa/panel.html` aparece una tarjeta debajo de Favoritos
   para abrir el webmail cuando el modulo esta activo.

## Seguridad

- La clave inicial del buzon se guarda cifrada si `CONFIG_ENC_KEY` esta
  disponible.
- Las credenciales de iRedAdmin-Pro se guardan cifradas.
- En Docker portable, `EMAIL_CORPORATIVO_IREDADMIN_PASSWORD` e
  `IREDMAIL_ADMIN_PASSWORD` viven en `deploy/.env.platform`; al arrancar el
  backend se registran en `pcs_superadministrador.configuraciones` y la clave se
  cifra antes de guardarse.
- La creacion de correo no bloquea la creacion de empresa si iRedMail no esta
  disponible.
- La consulta de empresa usa `/api/empresa/email_corporativo` y queda protegida
  por el wrapper de seguridad multiempresa.
- Las acciones super usan `WithSuperAuditoria` con accion
  `super_email_corporativo`.

## Tablas y configuracion

- Tabla: `empresa_email_corporativo` en `pcs_superadministrador`.
- Configuracion:
  - `email_corporativo.enabled`
  - `email_corporativo.auto_create`
  - `email_corporativo.domain`
  - `email_corporativo.webmail_url`
  - `email_corporativo.provision_mode`
  - `email_corporativo.iredadmin_api_base_url`
  - `email_corporativo.iredadmin_admin`
  - `email_corporativo.iredadmin_password`
  - `email_corporativo.quota_mb`

## Despliegue VPS

El despliegue Docker queda integrado al compose portable como perfil `mail` y
tambien documentado en `deploy/iredmail/`. El perfil queda apagado por defecto
para evitar abrir puertos de correo sin DNS, TLS, PTR y politica antispam listos.

Cuando `EMAIL_CORPORATIVO_ENABLED` esta activo, el script
`deploy/scripts/vps-configure-iredmail-host-nginx.sh` publica
`mail.powerfulcontrolsystem.com` en Nginx del host. El script valida que el
certificado Let's Encrypt cubra el subdominio `mail`; si no existe, prepara el
challenge HTTP y emite un certificado especifico con certbot antes de activar el
proxy HTTPS hacia iRedMail. Esto evita que el backend desactive la validacion
TLS para probar iRedAdmin.

Variables principales en `deploy/.env.platform`:

- `EMAIL_CORPORATIVO_ENABLED`
- `EMAIL_CORPORATIVO_AUTO_CREATE`
- `EMAIL_CORPORATIVO_DOMAIN`
- `EMAIL_CORPORATIVO_WEBMAIL_URL`
- `EMAIL_CORPORATIVO_PROVISION_MODE`
- `EMAIL_CORPORATIVO_IREDADMIN_API_BASE_URL`
- `EMAIL_CORPORATIVO_IREDADMIN_ADMIN`
- `EMAIL_CORPORATIVO_IREDADMIN_PASSWORD`
- `IREDMAIL_IMAGE`
- `IREDMAIL_HOSTNAME`
- `IREDMAIL_ADMIN_EMAIL`
- `IREDMAIL_ADMIN_PASSWORD`
- `IREDMAIL_MLMMJADMIN_API_TOKEN`
- `IREDMAIL_ROUNDCUBE_DES_KEY`
- `IREDMAIL_MYSQL_ROOT_PASSWORD`

Verificacion y arranque cuando el correo este listo:

```bash
deploy/scripts/vps-register-iredmail-secrets.sh
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml --profile mail up -d iredmail
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml --profile edge up -d edge
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml up -d --build backend
```

Antes de activar provision automatica:

- Validar registros DNS A, MX, SPF, DKIM, DMARC y PTR.
- Validar puertos 25, 80, 110, 143, 443, 465, 587, 993 y 995.
- Habilitar REST API en iRedAdmin-Pro.
- Restringir el API a la IP de la aplicacion.
- Probar login de iRedAdmin-Pro y creacion de usuario en entorno controlado.

## Registro de secretos

El 2026-05-28 se generaron secretos reales para la VPS actual y se guardaron en
`deploy/.env.platform` con permisos restringidos en el servidor. Luego se ejecuto
el registrador `backend/tools/register_iredmail_secrets` como binario Linux en la
VPS para persistir la configuracion en PostgreSQL. El password de iRedAdmin queda
guardado en `configuraciones` con cifrado de aplicacion; los valores sensibles no
deben imprimirse en consola, documentacion ni commits.
