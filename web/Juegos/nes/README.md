# Emulador NES integrado

Esta carpeta contiene una integración ligera de `jsnes` para pruebas locales.

- `index.html` — interfaz del emulador, orientación vertical móvil (canvas arriba, controles abajo), 5 ranuras de ROM.
- `nes-wrapper.js` — adaptador que carga ROMs, mapea teclado y touch, guarda ROMs en `localStorage` y guarda/restaura partidas si el emulador soporta snapshot JSON.
- `styles.css` — estilos responsive para mobile.

Importante:
- No se incluyen ROMs por razones de derechos. Debes subir tus propias ROMs `.nes` en las ranuras disponibles.
- El guardado de partida utiliza `nes.toJSON()` / `nes.fromJSON()` si están presentes en el build de emulador; si tu emulador no lo soporta, solo se guarda la ROM.

Uso rápido:
1. Abrir `web/juegos/nes/index.html` en un navegador.
2. En móvil, verás la pantalla (canvas) arriba y los controles debajo.
3. Subir ROMs a las ranuras y tocarlas para cargarlas.
4. Usar "Guardar partida" para intentar snapshot local.
