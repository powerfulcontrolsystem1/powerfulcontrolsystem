# Email corporativo con iRedMail

Este directorio deja preparada la integracion Docker para un servidor de correo
separado del runtime principal de Powerful Control System.

## Flujo soportado

1. El super administrador activa el modulo en `/super/email_corporativo.html`.
2. Al crear una empresa se genera un correo unico con el dominio configurado,
   por ejemplo `motel.calipso@powerfulcontrolsystem.com`.
3. Si iRedAdmin-Pro REST API esta configurado, el sistema intenta crear el buzon.
4. Si el API no esta listo, la cuenta queda pendiente y se puede provisionar luego.
5. En el panel de administrar empresa aparece la tarjeta para abrir el webmail.

## Nota tecnica importante

iRedMail Open Source no publica un flujo Docker comunitario unico dentro de su
documentacion principal. Para aprovisionamiento automatico por API se requiere
iRedAdmin-Pro con `ENABLE_RESTFUL_API = True`.

Por eso este compose es intencionalmente aislado y exige definir
`IREDMAIL_IMAGE` con una imagen soportada por el proveedor o por la estrategia de
operacion aprobada para la VPS. No se conecta al compose principal hasta validar:

- DNS: A, MX, SPF, DKIM, DMARC y PTR/reverse DNS.
- Puertos: 25, 80, 110, 143, 443, 465, 587, 993 y 995.
- TLS: certificado valido para `mail.powerfulcontrolsystem.com`.
- Backups: volumen de correo y base interna de iRedMail.
- Seguridad: API REST restringida a la IP interna de la aplicacion.

## Arranque manual

```powershell
Copy-Item deploy\iredmail\.env.example deploy\iredmail\.env
# Editar deploy\iredmail\.env con valores reales.
docker compose --env-file deploy\iredmail\.env -f deploy\iredmail\docker-compose.iredmail.yml --profile iredmail up -d
```

Despues de levantar iRedMail, en el panel super se configura:

- Dominio: `powerfulcontrolsystem.com`
- Webmail: `https://mail.powerfulcontrolsystem.com/mail/` o la ruta final real
- Modo de provision: `iRedAdmin-Pro API`
- URL API: `https://mail.powerfulcontrolsystem.com/iredadmin`
- Admin y clave iRedAdmin-Pro
