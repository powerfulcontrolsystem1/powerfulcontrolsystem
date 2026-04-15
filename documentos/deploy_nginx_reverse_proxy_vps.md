# Guia operativa: Nginx reverse proxy para dominio en VPS

Fecha: 2026-04-14
Alcance: Ubuntu VPS (Hostinger), app backend en puerto 8080

## Objetivo
Publicar la aplicacion backend en:
- http://powerfulcontrolsystem.com
- http://www.powerfulcontrolsystem.com
- http://empresa1.powerfulcontrolsystem.com (y subdominios por empresa)

usando Nginx como reverse proxy hacia:
- http://127.0.0.1:8080

## Comandos en orden (listos para copiar/pegar)

```bash
sudo apt update
sudo apt install -y nginx

sudo tee /etc/nginx/sites-available/powerfulcontrolsystem > /dev/null <<'EOF'
server {
    listen 80;
    listen [::]:80;
  server_name powerfulcontrolsystem.com www.powerfulcontrolsystem.com;

  location / {
    proxy_pass http://127.0.0.1:8080;
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
        proxy_pass http://127.0.0.1:8080;
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

if [ -e /etc/nginx/sites-enabled/default ]; then
  sudo rm -f /etc/nginx/sites-enabled/default
fi

sudo nginx -t
sudo systemctl restart nginx
sudo systemctl enable nginx

sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 8080/tcp
sudo ufw status

curl -I http://localhost:8080
curl -I http://powerfulcontrolsystem.com
curl -I -H "Host: empresa1.powerfulcontrolsystem.com" http://127.0.0.1/
curl -I -H "Host: empresa1.powerfulcontrolsystem.com" http://127.0.0.1/venta_publica.html
```

## HTTPS automatico (opcional recomendado)

```bash
sudo apt update
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d powerfulcontrolsystem.com -d www.powerfulcontrolsystem.com --redirect -m admin@powerfulcontrolsystem.com --agree-tos --no-eff-email
sudo certbot renew --dry-run
curl -I https://powerfulcontrolsystem.com
curl -I https://www.powerfulcontrolsystem.com
```

## Nota operativa
No se cambia el puerto 8080 del backend. Nginx publica el dominio y reenvia trafico al backend en localhost.

Si activas HTTPS con `certbot --nginx --redirect`, debes abrir `443/tcp` en UFW. Si el dominio redirige a `https://...` pero `443/tcp` queda cerrado, la aplicacion puede responder bien en `127.0.0.1:8080` y aun asi verse caída desde navegadores externos.

Si publicas `www.powerfulcontrolsystem.com`, mantenlo incluido en el certificado (`-d www.powerfulcontrolsystem.com`) y verifica que `http://www...` no quede devolviendo `404`, sino redirigiendo o resolviendo por HTTPS.

Para subdominios por empresa, crear un registro DNS wildcard `*.powerfulcontrolsystem.com` apuntando a `2.24.197.58`.

## Servicio persistente del backend

El backend del VPS no debe quedar corriendo con `nohup` manual. El flujo soportado es desplegar con `scripts/sync_to_vps.ps1` o `scripts/sync_to_vps.sh`, porque esos scripts ya crean o actualizan la unidad `systemd` del proyecto y la dejan habilitada para autoarranque.

Comandos utiles de verificacion en el VPS:

```bash
sudo systemctl status powerfulcontrolsystem.service --no-pager
sudo systemctl is-enabled powerfulcontrolsystem.service
sudo journalctl -u powerfulcontrolsystem.service -n 80 --no-pager
tail -n 80 /root/powerfulcontrolsystem/backend/server.err
tail -n 80 /root/powerfulcontrolsystem/backend/server.log
```
