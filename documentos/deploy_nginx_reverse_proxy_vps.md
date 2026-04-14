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
```

## Nota operativa
No se cambia el puerto 8080 del backend. Nginx publica el dominio y reenvia trafico al backend en localhost.

Para subdominios por empresa, crear un registro DNS wildcard `*.powerfulcontrolsystem.com` apuntando a `2.24.197.58`.
