# Email corporativo Mailu

## Objetivo

Permitir que cada empresa tenga un correo corporativo propio bajo el dominio
configurado, por ejemplo `motel.calipso@powerfulcontrolsystem.com`, usando una
plataforma abierta y portable en Docker.

## Decision tecnica

El proveedor activo es Mailu con webmail SnappyMail. Se elimina el flujo anterior
basado en iRedMail/iRedAdmin/Roundcube porque era mas pesado de operar y
Roundcube no cerraba bien el autologin sin contrasena en la VPS.

Mailu queda integrado en `deploy/docker-compose.platform.yml` con perfil
`mail`, apagado por defecto. El stack usa contenedores separados para front,
admin, IMAP, SMTP, antispam, webmail SnappyMail y Redis. El proxy publico lo administra
`deploy/scripts/vps-configure-mailu-host-nginx.sh`.

## Flujo operativo

1. El super administrador entra a `web/super/email_corporativo.html`.
2. Activa o desactiva el modulo global.
3. Define dominio, URL de webmail, cuota y modo de provision.
4. Al crear una empresa desde `/super/api/empresas`, el sistema genera un correo
   unico basado en el nombre de la empresa.
5. Si el correo ya existe, se usa un sufijo numerico: `empresa2@dominio`.
6. En modo `mailu_api`, el backend llama la API REST interna de Mailu para crear
   o actualizar el buzon. Usa `POST /api/v1/user` y, si ya existe, `PATCH` al
   mismo recurso con bearer token. La API no se publica a Internet y el backend
   no necesita Docker CLI ni socket Docker.
7. Si el modulo esta en modo manual o Mailu no esta listo, la creacion de la
   empresa no falla; el correo queda pendiente de provision.
8. En `web/administrar_empresa/panel.html` aparece la tarjeta de bandeja
   corporativa debajo de Favoritos cuando el modulo esta activo y la empresa
   tiene cuenta asignada.
9. El panel solicita una URL temporal de autologin. Esa URL apunta al subdominio
   del correo (`/pcs-mail-autologin`) y el proxy del host la envia al backend.
10. El backend valida el token HMAC, provisiona el buzon si aun esta pendiente y
    entra a SnappyMail por `/webmail/sso.php` con las cabeceras internas
    `X-Remote-User` y `X-Remote-User-Token`.
11. SnappyMail devuelve un hash temporal de inicio de sesion; el backend
    preserva la forma `/webmail/index.php?sso&hash=...` al agregar parametros de
    tema y, si el webmail entrega cookies de sesion, las traslada al navegador.
    El usuario no ve ni escribe contrasenas del buzon.
12. En la VPS, Nginx publica `/webmail/` contra el contenedor
    `pcs-mailu-webmail` en loopback y deja el resto de rutas de correo contra
    `pcs-mailu-front`.
13. El script de provision crea la identidad principal del usuario en SnappyMail
    para que la bandeja abra directamente y no muestre el modal inicial de
    "Actualizar identidad".
14. `deploy/mailu/snappymail-application.ini` se monta sobre
    `/defaults/application.ini` para conservar esa configuracion tras reinicios
    y permitir el iframe `same-site` del panel con `secfetch_allow`.
15. El panel empresarial envia `theme=light|dark` al consultar el buzon y al
    abrir `/pcs-mail-autologin`. El backend propaga esa preferencia hacia el
    script de provision con `PCS_MAILU_THEME_MODE` y `PCS_MAILU_THEME`.
16. Los temas `PCSLight@custom` y `PCSDark@custom` viven en
    `deploy/mailu/themes` y se copian al contenedor SnappyMail durante el
    arranque del perfil `mail`.
17. `Configuracion > Email corporativo` guarda `auto_open` en
    `empresa_estacion_prefs` con clave `email_corporativo_config`. El valor por
    defecto es `true`.
18. La misma pagina permite cambiar la contrasena interna del buzon. El backend
    valida longitud, cifra la clave con `CONFIG_ENC_KEY`, actualiza
    `empresa_email_corporativo.initial_password_enc` y reprovisiona Mailu por la
    API interna cuando el modo automatico esta activo.
19. Los correos automaticos del sistema salen por Mailu como dominio propio. Los
    correos comerciales y de compra/activacion de licencias usan
    `ventas@powerfulcontrolsystem.com`; alertas, recuperaciones, invitaciones,
    agente DIAN y pruebas usan `soporte@powerfulcontrolsystem.com`.
20. La pagina super permite `Provisionar ventas/soporte` para crear o actualizar
    esos buzones de sistema por la API interna Mailu, sin mostrar la clave
    generada.
21. La accion `Probar envio` envia un correo real desde el motor Mailu al
    destinatario indicado para validar el camino util de salida.
22. Los flujos legacy que antes usaban `gmail.smtp_*` para enviar directamente
    ahora construyen el mismo mensaje pero lo entregan por Mailu: confirmacion y
    recuperacion de administradores, invitaciones empresariales, asesor comercial,
    codigos de descuento, venta digital, pagos rechazados, vencimiento de
    licencias, reportes con adjunto, correos masivos y notificaciones de
    facturacion electronica.
23. Backend y worker se conectan además a `pcs_mailu_internal` y envían por
    `PCS_MAILU_SMTP_ADDR`, que apunta a la IP privada de `mailu-smtp`. Esta red
    acotada permite el relay de correos transaccionales sin confiar toda
    `pcs_internal` ni publicar un relay hacia Internet.
24. El cuerpo HTML usa como predeterminado `web/img/Logo pcs 1.png`, el logo del
    computador, embebido mediante `cid:`. El avatar externo del remitente se
    publica por separado con `web/img/bimi-pcs.svg`; requiere DMARC en
    enforcement, el registro DNS BIMI y, para Gmail, VMC o CMC.
25. Las acciones mutables de `web/super/email_corporativo.html` envían el token
    sincronizador `pcs_csrf` como `X-CSRF-Token`. Sin este encabezado el
    middleware debe rechazar guardar, probar, sincronizar o provisionar con
    HTTP 403.

## Seguridad

- La clave inicial del buzon se guarda cifrada si `CONFIG_ENC_KEY` esta
  disponible.
- El navegador no recibe claves ni tokens privados de correo.
- El navegador puede recibir cookies de sesion emitidas por SnappyMail para
  abrir la bandeja, pero no recibe la contrasena interna del buzon.
- La pagina empresarial de configuracion puede enviar una nueva contrasena, pero
  el backend nunca la devuelve y no la registra en logs ni documentacion.
- El token de autologin dura 2 minutos, va firmado con HMAC SHA-256 y se usa una
  sola vez desde el subdominio de correo.
- SnappyMail solo permite el iframe desde el mismo sitio (`same-site`), que cubre
  `powerfulcontrolsystem.com` y `mail.powerfulcontrolsystem.com`; no se permite
  embed abierto desde dominios externos.
- El proxy publico limpia `X-Auth-Email`, `X-Remote-User` y
  `X-Remote-User-Token` antes de llegar a Mailu; solo el backend puede inyectar
  cabeceras SSO desde la red interna Docker.
- No se deben imprimir claves iniciales, tokens Mailu, claves de administrador
  ni valores de `deploy/.env.platform` en consola, documentacion o respuestas.
- La consulta empresarial usa `/api/empresa/email_corporativo`, protegida por el
  wrapper multiempresa y por `empresa_id`.
- Las acciones super usan auditoria con accion `super_email_corporativo`.
- El backend no tiene Docker CLI ni socket Docker. El token de Mailu vive solo
  en el entorno del VPS, se transmite mediante `Authorization: Bearer` dentro
  de la red Docker y nunca se registra en logs o auditoria.
- El proxy publico rechaza `/api/` del dominio de correo; solo el backend usa la
  ruta interna `mailu-front` para la API de provisionamiento.
- Los contenedores Mailu usan IPs fijas dentro de `pcs_mailu_internal`. SnappyMail
  y el front resuelven `mailu-front`, `mailu-imap` y demas servicios hacia
  `192.168.203.x`; asi Dovecot recibe los accesos desde la red que Mailu marca
  como confiable y no desde la red general `pcs_internal`.

## Apariencia

- SnappyMail permite temas personalizados por CSS en una carpeta `themes`.
- El sistema instala dos temas propios: `PCSLight@custom` para apariencias
  claras y `PCSDark@custom` para apariencias oscuras.
- `web/administrar_empresa/panel.html` detecta el `data-theme` activo del
  sistema; si empieza por `dark`, usa el tema oscuro del correo, de lo contrario
  usa el tema claro.
- `deploy/scripts/vps-compose-sidecar-up.sh` deja `PCSLight@custom` como tema
  global base para buzones nuevos y copia ambos temas al contenedor
  `pcs-mailu-webmail`.
- `deploy/scripts/vps-provision-mailu-mailbox.sh` escribe una preferencia local
  por buzon en `settings/settings_local` cuando recibe `PCS_MAILU_THEME_MODE`.

## Tablas y configuracion

- Tabla: `empresa_email_corporativo` en `pcs_superadministrador`.
- Configuracion:
  - `email_corporativo.enabled`
  - `email_corporativo.auto_create`
  - `email_corporativo.domain`
  - `email_corporativo.webmail_url`
  - `email_corporativo.provision_mode`
  - `email_corporativo.mailu_api_base_url`
  - `email_corporativo.mailu_admin`
  - `email_corporativo.mailu_api_token`
  - `email_corporativo.quota_mb`
- `email_corporativo.direct_provision_command`
- `email_corporativo.restart_alert_to`
- `email_corporativo.restart_alert_enabled`

Las claves historicas `gmail.*` pueden existir en bases antiguas por
compatibilidad de configuracion y restauracion, pero el menu operativo y los
envios reales del sistema deben usar Email corporativo Mailu.

## Variables VPS

- `EMAIL_CORPORATIVO_ENABLED`
- `EMAIL_CORPORATIVO_AUTO_CREATE`
- `EMAIL_CORPORATIVO_DOMAIN`
- `EMAIL_CORPORATIVO_WEBMAIL_URL`
- `EMAIL_CORPORATIVO_PROVISION_MODE`
- `EMAIL_CORPORATIVO_MAILU_API_BASE_URL`
- `EMAIL_CORPORATIVO_INTERNAL_MAILU_API_BASE_URL`
- `EMAIL_CORPORATIVO_INTERNAL_WEBMAIL_URL`
- `EMAIL_CORPORATIVO_INTERNAL_SNAPPYMAIL_URL`
- `EMAIL_CORPORATIVO_MAILU_ADMIN`
- `MAILU_API_TOKEN` (generado con aleatoriedad criptografica por el arranque si
  falta; no copiar ni imprimir). El backend lo recibe como
  `EMAIL_CORPORATIVO_MAILU_API_TOKEN`.
- `EMAIL_CORPORATIVO_QUOTA_MB`
- `EMAIL_CORPORATIVO_DIRECT_PROVISION_COMMAND` (compatibilidad heredada: si se
  define, debe coincidir con el script interno revisado; produccion usa
  `mailu_api` y no ejecuta comandos configurables).
- `EMAIL_CORPORATIVO_AUTOLOGIN_SECRET`
- `MAILU_ENABLED`
- `MAILU_VERSION`
- `MAILU_HOSTNAME`
- `MAILU_SECRET_KEY`
- `MAILU_API_ENABLED`
- `MAILU_RESOLVER_IP`
- `MAILU_REDIS_IP`
- `MAILU_SMTP_IP`
- `PCS_MAILU_SMTP_ADDR` (inyectada por Compose desde `MAILU_SMTP_IP`; no debe
  apuntar al puerto SMTP público).
- `MAILU_ANTISPAM_IP`
- `MAILU_WEBMAIL_IP`
- `MAILU_IMAP_IP`
- `MAILU_ADMIN_IP`
- `MAILU_FRONT_IP`
- `MAILU_MESSAGE_SIZE_LIMIT`
- `MAILU_WEB_WEBMAIL`
- `MAILU_WEB_ADMIN`
- `MAILU_WEBROOT_REDIRECT`
- `MAILU_PROXY_AUTH_WHITELIST`
- `MAILU_PROXY_AUTH_HEADER`
- `MAILU_PROXY_AUTH_CREATE`
- `MAILU_REAL_IP_FROM`
- `MAILU_ADMIN_PASSWORD`
- `MAILU_WEBMAIL`
- `MAILU_TLS_FLAVOR`
- `MAILU_HTTP_PORT`
- `MAILU_WEBMAIL_PORT`

## Despliegue

```bash
deploy/scripts/vps-docker-preflight.sh
deploy/scripts/vps-compose-sidecar-up.sh
deploy/scripts/vps-configure-mailu-host-nginx.sh
```

Antes de activar correo real:

- Validar DNS A y MX para `mail.powerfulcontrolsystem.com`.
- Validar SPF, DKIM, DMARC y PTR/rDNS.
- Revisar puertos SMTP/IMAP segun la politica del proveedor VPS.
- Mantener Mailu inicialmente en loopback y publicar webmail por Nginx del host.
- Probar creacion de buzon desde `web/super/email_corporativo.html` con
  `Probar Mailu`.
- Probar `Provisionar ventas/soporte` y luego `Probar envio`.

## Diagnostico 2026-07-26

- `powerfulcontrolsystem.com` y `mail.powerfulcontrolsystem.com` resuelven a
  `2.24.197.58`.
- SPF publicado: `v=spf1 mx a:mail.powerfulcontrolsystem.com ~all`.
- Mailu genero la clave DKIM RSA del dominio con selector `dkim`; su parte
  publica esta en `dkim._domainkey.powerfulcontrolsystem.com`. La clave privada
  permanece exclusivamente dentro de Mailu.
- DMARC esta en enforcement gradual con `p=quarantine`, `pct=100`, alineacion
  estricta SPF/DKIM y reportes a `postmaster@powerfulcontrolsystem.com`.
- BIMI esta publicado en `default._bimi.powerfulcontrolsystem.com` y apunta a
  `https://powerfulcontrolsystem.com/img/bimi-pcs.svg`.
- El MX prioridad 10 apunta a `mail.powerfulcontrolsystem.com`.
- PTR/rDNS IPv4 confirmado: `2.24.197.58` resuelve en reverso a
  `mail.powerfulcontrolsystem.com` y ese hostname vuelve a `2.24.197.58`
  (FCrDNS valido).
- Prueba real hacia Gmail: entrega aceptada en uno a cuatro segundos, con
  `SPF PASS`, `DKIM PASS` para `powerfulcontrolsystem.com` y `DMARC PASS`.
  Gmail mostro correctamente el PNG del computador embebido por CID en el
  cuerpo del mensaje.
- El avatar circular del remitente depende ademas de la cache y politica del
  cliente receptor. Gmail requiere un certificado VMC o CMC para mostrar el
  logo BIMI verificado; el TXT BIMI sin certificado no garantiza ese avatar.
- La ausencia de VMC/CMC es un bloqueo externo abierto del Plan 108: el logo
  inline ya usa el archivo oficial, pero la inicial generada por Gmail no se
  puede eliminar de forma garantizada hasta publicar el PEM del certificado en
  el campo `a=` del registro BIMI y repetir la prueba visual.

## Validacion BIMI 2026-08-25

- Se revalidaron publicamente SPF, DKIM, DMARC, MX, PTR y el TXT BIMI. Todos
  permanecen publicados; el campo `a=` continua vacio hasta adquirir VMC/CMC.
- Los ocho contenedores Mailu del VPS estaban activos y saludables durante la
  comprobacion de solo lectura.
- `web/img/bimi-pcs.svg` se corrigio para eliminar atributos y valores no
  admitidos por el esquema SVG Tiny P/S. El archivo paso sin errores el esquema
  oficial `SVG_PS-latest.rnc` del BIMI Group.
- La fuente del VPS se actualizo con respaldo y hash verificado. El contenedor
  activo usa filesystem de solo lectura y aun entrega la copia anterior; la
  publicacion efectiva queda pendiente del siguiente despliegue normal.
- Antes de comprar el certificado debe fijarse el logo exacto que se presentara
  a la autoridad. Un CMC exige demostrar al menos 12 meses de uso previo; un VMC
  exige que el logo este registrado como marca. El certificado queda ligado a
  ese logo, por lo que no se debe sustituir despues sin reemitirlo.
