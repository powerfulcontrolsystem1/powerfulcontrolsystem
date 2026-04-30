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
  empresas/
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
- Partidas guardadas por empresa: `juegos/empresas/`
- En el VPS recomendado:
  - Emulador: `/opt/juegos/public/index.html`
  - EmulatorJS: `/opt/juegos/emulator/data/loader.js`
  - Juegos/ROMs: `/opt/juegos/roms/`
  - Partidas guardadas: `/opt/juegos/empresas/`
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
/opt/juegos/emulator/data/emulator.min.js
/opt/juegos/emulator/data/cores/snes9x-wasm.data
```

El script tambien descarga cores locales basicos para las ROMs incluidas en este proyecto: SNES (`snes9x`), NES (`fceumm`), Game Boy/Game Boy Color (`gambatte`), Game Boy Advance (`mgba`) y Mega Drive (`genesis_plus_gx`). Puedes ajustar la lista con la variable `PCS_EMULATORJS_CORES`.

## Agregar ROMs

Copia tus ROMs legales o homebrew:

```bash
sudo cp mi-juego.sfc /opt/juegos/roms/
sudo chown -R www-data:www-data /opt/juegos
```

Formatos aceptados: `.sfc`, `.smc`, `.fig`, `.swc`, `.nes`, `.gb`, `.gbc`, `.gba`, `.gen`, `.bin`, `.zip`.

Para pruebas locales se pueden colocar ROMs homebrew en `juegos/roms/`. Esa carpeta esta ignorada por git, salvo este README, para evitar redistribuir ROMs accidentalmente. En esta estacion de trabajo quedaron instaladas ROMs de prueba homebrew de Retrobrews:

- `blt.sfc`
- `bucket.smc`
- `hilda.sfc`
- `horizontal-shooter.sfc`
- `rockfall.smc`
- `nes-31-in-1-real-game.nes`
- `nes-assimilate.nes`
- `nes-babel-blox.nes`
- `nes-black-box-challenge.nes`
- `nes-debris-dodger.nes`
- `nes-invaders.nes`
- `nes-lunar-limit.nes`
- `gb-2048.gb`
- `gb-8-bitty-games-collection.gb`
- `gb-1d-marathon.gb`
- `gbc-2560-colors-demo.gbc`
- `gba-apotris.gba`
- `gba-bloxorz.gba`
- `gba-butano-fighter.gba`
- `gen-cave-story-md-es.gen`

Fuentes de referencia:

- SNES/NES homebrew: `https://github.com/retrobrews/snes-games` y `https://github.com/retrobrews/nes-games`
- GB/GBC homebrew: `https://github.com/gbdev/database`
- GBA homebrew: `https://github.com/gbadev-org/games`
- Mega Drive homebrew: `https://github.com/andwn/cave-story-md`

El backend detecta automaticamente el core de EmulatorJS por extension:

- SNES: `.sfc`, `.smc`, `.fig`, `.swc`
- NES: `.nes`
- Game Boy / Game Boy Color: `.gb`, `.gbc`
- Game Boy Advance: `.gba`
- Sega Mega Drive / Genesis: `.gen`, `.bin`

## Compilar Go

```bash
cd /opt/juegos
go build -o juegos-server ./app
```

## Ejecutar manualmente

```bash
./juegos-server -addr 127.0.0.1:8099 -public ./public -emulator ./emulator -roms ./roms -saves ./empresas -core snes
```

Prueba:

```bash
curl http://127.0.0.1:8099/health
curl http://127.0.0.1:8099/api/roms
curl "http://127.0.0.1:8099/api/saves/latest?empresa_id=7&rom=blt.sfc"
```

## Partidas guardadas por empresa

El emulador conserva partidas separadas por empresa. La empresa se resuelve en este orden:

1. Parametro `empresa_id` en la URL del emulador.
2. `sessionStorage` o `localStorage`: `active_empresa_id`, `empresa_id`, `admin_empresa_id`.
3. Carpeta `empresa_publico` si no hay contexto de empresa.

Cuando EmulatorJS emite `EJS_onSaveState`, el frontend sube el estado a:

```text
POST /api/saves/state
```

Cuando EmulatorJS emite `EJS_onSaveUpdate`, el frontend sube el guardado interno del juego a:

```text
POST /api/saves/file
```

El backend guarda los archivos en:

```text
empresas/
  empresa_<ID>/
    emulador/
      <rom-saneada>/
        latest.state
        latest.save
        latest.png
        meta.json
```

Al abrir un juego, el frontend consulta `/api/saves/latest` y, si existe `latest.state`, configura `EJS_loadStateURL` para retomar la partida automaticamente.

> Nota tecnica: EmulatorJS documenta `EJS_loadStateURL`, `EJS_onSaveState`, `EJS_onSaveUpdate` y `EJS_fixedSaveInterval`. El boton `Guardar avance` intenta usar `EJS_emulator.gameManager.getState()` cuando la version instalada lo expone; si no existe, el juego sigue funcionando y el usuario puede guardar desde el menu nativo del emulador.

## Servicio systemd

```bash
sudo cp deploy/juegos.service /etc/systemd/system/juegos.service
sudo systemctl daemon-reload
sudo systemctl enable --now juegos
sudo systemctl status juegos
```

Antes de iniciar el servicio, crea la carpeta de partidas:

```bash
sudo mkdir -p /opt/juegos/empresas
sudo chown -R www-data:www-data /opt/juegos/empresas
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
- `/api/saves/*` solo guarda y lee archivos saneados por `empresa_id` y ROM permitida.
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
EJS_core = "snes"; // o el core detectado para cada ROM
EJS_gameUrl = "/roms/archivo.sfc";
```

Para ZIP se usa el core de respaldo definido con el flag `-core`.
