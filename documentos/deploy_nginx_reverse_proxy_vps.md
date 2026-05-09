# Guia operativa: Nginx reverse proxy para dominio en VPS

Fecha: 2026-05-09
Alcance: Ubuntu VPS (Hostinger), Nginx publico hacia Docker en puerto interno 8081

## Estado actual Docker

La VPS actual ya fue conmutada a Docker para el nucleo de la plataforma. Nginx del host sigue publicando `80/443`, pero el upstream ya no debe apuntar al backend systemd en `127.0.0.1:8080`; ahora apunta al frontend Docker en:

```text
http://127.0.0.1:8081
```

El servicio anterior `powerfulcontrolsystem.service` queda disponible como rollback temporal. El backup de configuracion creado durante la conmutacion es:

```bash
/etc/nginx/sites-available/powerfulcontrolsystem.bak.20260509-193744
```

Verificacion actual:

```bash
cd /root/powerfulcontrolsystem
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml ps
curl -I http://127.0.0.1:8081/
curl -I https://powerfulcontrolsystem.com
```

## Objetivo

Publicar la aplicacion en:

- `https://powerfulcontrolsystem.com`
- `https://www.powerfulcontrolsystem.com`
- `https://empresa1.powerfulcontrolsystem.com` y subdominios por empresa

usando Nginx del host como reverse proxy hacia el frontend Docker:

- `http://127.0.0.1:8081`

## Configuracion base

Ejemplo de bloque HTTP. En produccion puede coexistir con los bloques HTTPS generados por Certbot, manteniendo el mismo upstream `127.0.0.1:8081`.

```bash
sudo tee /etc/nginx/sites-available/powerfulcontrolsystem > /dev/null <<'EOF'
server {
  listen 80;
  listen [::]:80;
  server_name powerfulcontrolsystem.com www.powerfulcontrolsystem.com;

  location / {
    proxy_pass http://127.0.0.1:8081;
    proxy_http_version 1.1;

    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;

    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";

    proxy_read_timeout 300;
    proxy_send_timeout 300;
  }
}

server {
  listen 80;
  listen [::]:80;
  server_name ~^(?<empresa_slug>[a-z0-9-]+)\.powerfulcontrolsystem\.com$;

  location = / {
    return 302 /venta_publica.html?empresa_slug=$empresa_slug;
  }

  location / {
    proxy_pass http://127.0.0.1:8081;
    proxy_http_version 1.1;

    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;

    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";

    proxy_read_timeout 300;
    proxy_send_timeout 300;
  }
}
EOF

sudo ln -sfn /etc/nginx/sites-available/powerfulcontrolsystem /etc/nginx/sites-enabled/powerfulcontrolsystem
sudo nginx -t
sudo systemctl reload nginx
```

## Verificacion

```bash
curl -I http://127.0.0.1:8081/
curl -I https://powerfulcontrolsystem.com
curl -I -H "Host: empresa1.powerfulcontrolsystem.com" http://127.0.0.1/
curl -I -H "Host: empresa1.powerfulcontrolsystem.com" http://127.0.0.1/venta_publica.html
```

Si `https://powerfulcontrolsystem.com` responde `200 OK` pero Docker no esta saludable, revisa:

```bash
cd /root/powerfulcontrolsystem
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml ps
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml logs --tail=120 backend
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml logs --tail=120 frontend
```

## HTTPS y DNS

Mantener abiertos `80/tcp` y `443/tcp`. El puerto `8081` debe permanecer local en `127.0.0.1`, no expuesto publicamente.

Para subdominios por empresa, mantener DNS wildcard:

```text
*.powerfulcontrolsystem.com -> 2.24.197.58
```

El certificado wildcard documentado en el manual de instalacion cubre `powerfulcontrolsystem.com` y `*.powerfulcontrolsystem.com`.

## Servicio legacy y rollback

Con Docker activo, el backend legacy por systemd no es el upstream principal. Puede quedar activo temporalmente para rollback:

```bash
sudo systemctl status powerfulcontrolsystem.service --no-pager
sudo journalctl -u powerfulcontrolsystem.service -n 80 --no-pager
```

Rollback rapido al upstream anterior:

```bash
sudo cp /etc/nginx/sites-available/powerfulcontrolsystem.bak.20260509-193744 /etc/nginx/sites-available/powerfulcontrolsystem
sudo nginx -t
sudo systemctl reload nginx
sudo systemctl status powerfulcontrolsystem.service --no-pager
```

Despues del rollback, validar:

```bash
curl -I http://127.0.0.1:8080/
curl -I https://powerfulcontrolsystem.com
```
