# Jeugos - Emulador Web con EmulatorJS

Emulador web multiplataforma para navegador PC y movil usando Go nativo, Nginx y EmulatorJS.

## Estructura

```text
jeugos/
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
    jeugos.service
    fetch_emulatorjs.sh
  go.mod
```

## Donde encontrar el emulador y los juegos

- Emulador web: `jeugos/public/index.html`
- Servidor Go del emulador: `jeugos/app/main.go`
- Archivos oficiales de EmulatorJS: `jeugos/emulator/data/`
- Juegos/ROMs: `jeugos/roms/`
- En el VPS recomendado:
  - Emulador: `/opt/jeugos/public/index.html`
  - EmulatorJS: `/opt/jeugos/emulator/data/loader.js`
  - Juegos/ROMs: `/opt/jeugos/roms/`
- URL de acceso despues de configurar Nginx: `http://juegos.tu-dominio.com/` o `https://juegos.tu-dominio.com/`
- API para listar juegos disponibles: `/api/roms`

## Instalar en Linux/VPS

```bash
sudo apt update
sudo apt install -y golang nginx curl unzip
sudo mkdir -p /opt/jeugos
sudo rsync -a ./jeugos/ /opt/jeugos/
cd /opt/jeugos
```

## Instalar EmulatorJS

```bash
sh deploy/fetch_emulatorjs.sh
```

Debe quedar disponible:

```text
/opt/jeugos/emulator/data/loader.js
```

## Agregar ROMs

Copia tus ROMs legales de SNES:

```bash
sudo cp mi-juego.sfc /opt/jeugos/roms/
sudo chown -R www-data:www-data /opt/jeugos
```

Formatos aceptados: `.sfc`, `.smc`, `.fig`, `.swc`, `.zip`.

## Compilar Go

```bash
cd /opt/jeugos
go build -o jeugos-server ./app
```

## Ejecutar manualmente

```bash
./jeugos-server -addr 127.0.0.1:8099 -public ./public -emulator ./emulator -roms ./roms -core snes
```

Prueba:

```bash
curl http://127.0.0.1:8099/health
curl http://127.0.0.1:8099/api/roms
```

## Servicio systemd

```bash
sudo cp deploy/jeugos.service /etc/systemd/system/jeugos.service
sudo systemctl daemon-reload
sudo systemctl enable --now jeugos
sudo systemctl status jeugos
```

## Nginx

```bash
sudo cp deploy/nginx.conf /etc/nginx/sites-available/jeugos.conf
sudo ln -s /etc/nginx/sites-available/jeugos.conf /etc/nginx/sites-enabled/jeugos.conf
sudo nginx -t
sudo systemctl reload nginx
```

Edita `server_name` y rutas si no usas `/opt/jeugos`.

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
