# Juegos - Emulador Web con EmulatorJS

Emulador web multiplataforma para navegador PC y movil usando Go nativo, Nginx y EmulatorJS.

## Estructura

```text
juegos/
  app/
    main.go
  public/
    index.html
    app.js
    styles.css
  emulator/
    data/
      loader.js
  roms/
  deploy/
    nginx.conf
    juegos.service
    fetch_emulatorjs.sh
  go.mod
```

## Donde encontrar el emulador y los juegos

- Emulador web: `juegos/public/index.html`
- Servidor Go del emulador: `juegos/app/main.go`
- Archivos oficiales de EmulatorJS: `juegos/emulator/data/`
- Juegos/ROMs: `juegos/roms/`
- En el VPS recomendado:
  - Emulador: `/opt/juegos/public/index.html`
  - EmulatorJS: `/opt/juegos/emulator/data/loader.js`
  - Juegos/ROMs: `/opt/juegos/roms/`
- URL de acceso despues de configurar Nginx: `http://juegos.tu-dominio.com/` o `https://juegos.tu-dominio.com/`
- URL integrada desde el menu de juegos del sistema: `https://tu-dominio.com/emulador/`
- Entrada visual del sistema principal: `web/Juegos/menu_juegos.html` -> tarjeta `Emulador SNES`
- API para listar juegos disponibles: `/api/roms`

## Instalar en Linux/VPS

```bash
sudo apt update
sudo apt install -y golang nginx curl unzip
sudo mkdir -p /opt/juegos
sudo rsync -a ./juegos/ /opt/juegos/
cd /opt/juegos
```

## Instalar EmulatorJS

```bash
sh deploy/fetch_emulatorjs.sh
```

Debe quedar disponible:

```text
/opt/juegos/emulator/data/loader.js
```

## Agregar ROMs

Copia tus ROMs legales de SNES:

```bash
sudo cp mi-juego.sfc /opt/juegos/roms/
sudo chown -R www-data:www-data /opt/juegos
```

Formatos aceptados: `.sfc`, `.smc`, `.fig`, `.swc`, `.zip`.

Para pruebas locales se pueden colocar ROMs homebrew en `juegos/roms/`. Esa carpeta esta ignorada por git, salvo este README, para evitar redistribuir ROMs accidentalmente. En esta estacion de trabajo quedaron instaladas ROMs de prueba homebrew de Retrobrews:

- `blt.sfc`
- `bucket.smc`
- `hilda.sfc`
- `horizontal-shooter.sfc`
- `rockfall.smc`

Fuente de referencia: `https://github.com/retrobrews/snes-games`.

## Compilar Go

```bash
cd /opt/juegos
go build -o juegos-server ./app
```

## Ejecutar manualmente

```bash
./juegos-server -addr 127.0.0.1:8099 -public ./public -emulator ./emulator -roms ./roms -core snes
```

Prueba:

```bash
curl http://127.0.0.1:8099/health
curl http://127.0.0.1:8099/api/roms
```

## Servicio systemd

```bash
sudo cp deploy/juegos.service /etc/systemd/system/juegos.service
sudo systemctl daemon-reload
sudo systemctl enable --now juegos
sudo systemctl status juegos
```

## Nginx

```bash
sudo cp deploy/nginx.conf /etc/nginx/sites-available/juegos.conf
sudo ln -s /etc/nginx/sites-available/juegos.conf /etc/nginx/sites-enabled/juegos.conf
sudo nginx -t
sudo systemctl reload nginx
```

Edita `server_name` y rutas si no usas `/opt/juegos`.

Para que la tarjeta `Emulador SNES` del menu de juegos abra el emulador dentro del dominio principal, agrega al `server` de Powerful Control System el bloque documentado en `deploy/nginx.conf`:

```nginx
location = /emulador {
    return 301 /emulador/;
}

location ^~ /emulador/ {
    rewrite ^/emulador/(.*)$ /$1 break;
    proxy_pass http://127.0.0.1:8099;
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

## Seguridad

- No hay subida de archivos desde frontend.
- `/api/roms` solo lista archivos permitidos.
- `/roms/*` sanitiza rutas y bloquea traversal.
- Nginx sirve `public`, `emulator` y `roms` como archivos estaticos.
- El backend Go queda escuchando en `127.0.0.1:8099` detras de Nginx.

## Extras incluidos

- Seleccion de ROM recordada en `localStorage`.
- Indicador de gamepad usando Gamepad API.
- Carga diferida del loader de EmulatorJS al presionar `Jugar`.
- Boton de pantalla completa.
- UI responsive para PC y movil.

## Notas

El core por defecto es `snes`, configurado en frontend con:

```js
EJS_player = "#game";
EJS_core = "snes";
EJS_gameUrl = "/roms/archivo.sfc";
```

Para otro sistema cambia el flag `-core` y ajusta las extensiones permitidas en `app/main.go`.
