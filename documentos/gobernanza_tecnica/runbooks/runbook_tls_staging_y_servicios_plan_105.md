# Runbook: recuperacion TLS de staging y servicios externos - Plan 105

Fecha: 2026-07-21  
Estado: TLS y migracion de sesiones recuperados el 2026-07-22; vigilancia y
matriz E2E de staging aun pendientes.

## 1. Sintoma y alcance

Las comprobaciones externas sin omitir la validacion TLS detectaron certificado
expirado en `staging.powerfulcontrolsystem.com`,
`onlyoffice.powerfulcontrolsystem.com` y
`nextcloud.powerfulcontrolsystem.com`. El nucleo publico de PCS responde
correctamente en `/health` y `/ready`; por tanto no se debe diagnosticar como
caida general de la aplicacion.

El incidente bloqueo P105-007 (staging equivalente) y P105-011 (pruebas E2E
con proveedores). El 2026-07-22 se emitio el certificado SAN
`pcs-services-202607` para los tres nombres y se recargaron los vhosts. La
validacion externa sin `--insecure` confirmo OnlyOffice y Nextcloud. Staging ya
no falla TLS; el desajuste inicial de upstream y la migracion de sesiones fueron
corregidos de forma trazable. El smoke visual autenticado ya carga el panel
superadministrador; la matriz E2E completa aun no esta cerrada.

## 2. Precondiciones y protecciones

1. Obtener aprobacion para intervenir el VPS/edge y una ventana de mantenimiento
   si el cambio afecta los puertos 80/443.
2. Tomar evidencia antes de cambiar: fecha/hora UTC, hostname, emisor, fecha de
   expiracion, IP resuelta, estado de `nginx`/contenedores y ultimo log de ACME.
   No guardar llaves privadas, tokens ni archivos `.env` en tickets, consola o git.
3. Confirmar con DNS que cada hostname llega al edge esperado y que el puerto 80
   permite `/.well-known/acme-challenge/`. Un certificado no puede renovarse si
   el challenge llega a otro host.
4. Identificar el propietario real de cada terminacion TLS. El compose de PCS
   cubre el edge principal; OnlyOffice y Nextcloud pueden tener su propio proxy,
   compose o panel. No ejecutar una renovacion del edge principal suponiendo que
   cubre esos dos servicios.

## 3. Diagnostico reproducible

Desde una maquina externa, sin credenciales, ejecutar para cada hostname:

```powershell
curl.exe -Iv https://staging.powerfulcontrolsystem.com/health
curl.exe -Iv https://onlyoffice.powerfulcontrolsystem.com/healthcheck
curl.exe -Iv https://nextcloud.powerfulcontrolsystem.com/status.php
Resolve-DnsName staging.powerfulcontrolsystem.com
Resolve-DnsName onlyoffice.powerfulcontrolsystem.com
Resolve-DnsName nextcloud.powerfulcontrolsystem.com
```

En el VPS, inspeccionar antes de actuar:

```sh
sudo systemctl status nginx --no-pager
sudo docker ps --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}'
sudo certbot certificates
sudo find /etc/letsencrypt/live -maxdepth 2 -name fullchain.pem -print
sudo crontab -l
sudo systemctl list-timers --all | grep -Ei 'certbot|acme|renew'
```

Si la terminacion es Docker, revisar tambien el nombre de proyecto, los mounts
de `/etc/letsencrypt` y el log del contenedor/proxy; no asumir que existe
`pcs-edge` ni reiniciar contenedores ajenos al hostname afectado.

## 4. Recuperacion segun terminacion

### 4.1 Edge Docker principal de PCS

Los scripts versionados son `deploy/scripts/vps-docker-edge-up.sh` y
`deploy/scripts/vps-docker-edge-renew.sh`. El segundo ejecuta `certbot renew`
en el perfil `certbot` y recarga `pcs-edge` solo si existe. Usar la renovacion
primero; `edge-up` puede detener Nginx del host y solo corresponde a una
conmutacion previamente aprobada.

```sh
cd /root/powerfulcontrolsystem
sudo PROJECT_DIR=/root/powerfulcontrolsystem \
  COMPOSE_FILE=/root/powerfulcontrolsystem/deploy/docker-compose.platform.yml \
  ENV_FILE=/root/powerfulcontrolsystem/deploy/.env.platform \
  bash deploy/scripts/vps-docker-edge-renew.sh
sudo docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml --profile edge ps
```

Si la renovacion informa que no encuentra un certificado, falla el challenge o
el certificado no contiene el hostname, detenerse. Verificar DNS, `EDGE_DOMAIN`,
`EDGE_CERT_EXTRA_DOMAINS`, el webroot y la propiedad del proxy antes de volver a
solicitar uno. No ejecutar `vps-docker-edge-up.sh` sin `CONFIRM_DOCKER_EDGE=YES`
y una autorizacion que cubra expresamente la conmutacion 80/443.

### 4.2 Staging con Nginx del host

`deploy/scripts/vps-configure-staging-nginx.sh` referencia actualmente un
certificado de Let's Encrypt concreto y tiene `UPSTREAM` configurable. Antes de
recargar Nginx, comprobar que esa ruta existe, no esta vencida y contiene el
dominio staging. Confirmar tambien el puerto publicado por el compose; en el
VPS verificado el contenedor expone `127.0.0.1:18082`, por lo que el vhost debe
usar `UPSTREAM=http://127.0.0.1:18082`. Renovar con el mecanismo que administra
ese certificado (normalmente `certbot renew`) y luego:

```sh
sudo nginx -t
sudo systemctl reload nginx
curl -fsSI https://staging.powerfulcontrolsystem.com/health
curl -fsS https://staging.powerfulcontrolsystem.com/ready
```

No sobrescribir el vhost para "arreglar" una ruta de certificado sin respaldar
la configuracion existente y sin confirmar el nombre de certificado correcto.

### 4.3 OnlyOffice y Nextcloud

Determinar por separado su proxy/compose/panel y renovar el certificado en ese
propietario. Recargar solamente el proxy que sirve cada hostname y conservar la
cadena completa (`fullchain`) y clave emparejada. Si comparten certificado SAN
con PCS, validar los tres nombres antes de cerrar; si son certificados distintos,
cerrar cada evidencia por separado.

## 5. Validacion posterior obligatoria

1. Desde fuera del VPS, repetir los tres `curl -Iv` sin `-k`/`--insecure` y
   comprobar cadena valida, hostname coincidente y fecha futura.
2. Confirmar respuesta funcional minima: staging `/health` y `/ready`,
   OnlyOffice `/healthcheck`, Nextcloud `/status.php`.
3. Abrir visualmente staging autenticado y ejecutar la matriz P105-007; solo
   despues retomar P105-011 con los proveedores autorizados.
4. Verificar que existe una programacion de renovacion y una alerta previa al
   vencimiento. Registrar responsable, mecanismo, proxima ejecucion y evidencia
   sin secretos en `documentos/plan_105.md`.
5. Repetir las sondas 15 minutos despues de la recarga para detectar un proxy
   secundario o balanceador con certificado anterior.

## 6. Rollback y escalamiento

Si una recarga deja 5xx, restaurar el archivo de configuracion respaldado,
ejecutar `nginx -t` y recargar el proxy afectado. No restaurar certificados
vencidos: escalar al responsable DNS/edge con la salida saneada de `certbot` y
el hostname que fallo. Si el challenge no puede llegar al servidor correcto,
la recuperacion requiere una correccion de DNS o del balanceador, no un reintento
ciego de ACME.

## 7. Relacion con Plan 105

- P105-007 queda en curso: TLS, migracion de sesiones, health/ready y login
  visual ya estan comprobados; falta completar la matriz E2E y consolidar el
  SHA formal.
- P105-011 puede retomar sus pruebas E2E autorizadas porque OnlyOffice y
  Nextcloud tienen TLS valido.
- P105-010 debe recibir una alerta de expiracion para que el incidente no vuelva
  a descubrirse durante una liberacion.
