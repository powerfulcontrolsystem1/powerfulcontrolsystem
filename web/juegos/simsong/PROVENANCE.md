# Recursos de GTA SIMSong

Three.js **0.186.0 / r186**, MIT, fue autorizado expresamente por el usuario para
renderizado, GLTF y animaciones el 2026-09-14. Se aloja en
[`../vendor/three/`](../vendor/three/LICENSE). Los únicos cambios al código del
proveedor son las rutas relativas de los imports de GLTFLoader. No usa CDN.

Los modelos de Kenney se distribuyen bajo **CC0**, con el texto de licencia en
cada carpeta de recursos. Se incorporan selecciones de estos paquetes oficiales:

- [Animated Characters Protagonists](https://kenney.nl/assets/animated-characters-protagonists):
  tres personajes, rig y clips idle/run/jump. FBX convertido a GLB con Blender
  4.5.13; texturas originales separadas a PNG para respetar el CSP de PCS.
  Los clips activos se exportan por separado y se reúnen por nombre de hueso;
  un ajuste de postura de hombros acompaña la mezcla de clips en el navegador.
- [Car Kit](https://kenney.nl/assets/car-kit): automóviles, taxi, furgón y policía.
- [City Kit Suburban](https://kenney.nl/assets/city-kit-suburban): viviendas de fondo.
- [Furniture Kit](https://kenney.nl/assets/furniture-kit): muebles de ambos interiores.

[`assets/manifest.json`](assets/manifest.json) registra SHA-256, tamaños y las
rutas distribuidas. Los personajes son modelos genéricos de Kenney; no se
presentan como modelos oficiales de los personajes de Los Simpson.

## Referencias de la reconstrucción original

La distribución es una interpretación compacta propia, no un mapa canónico ni
una copia de un mapa de un videojuego. Se contrastaron las ubicaciones de
[The Simpsons Archive](https://www.simpsonsarchive.com/guides/city.profile.html),
el [mapa del nivel 1 de Hit & Run](https://gamefaqs.gamespot.com/ps2/914690-the-simpsons-hit-and-run/map/1209-level-1-map),
Virtual Springfield y recreaciones comunitarias. El mapa publicado como
[Springfield de The Simpsons Game](https://sketchfab.com/3d-models/springfield-full-map-the-simpsons-game-7e1bac6328d24f988a24f08cced47f95)
no se incorpora: la publicación describe recursos extraídos de un juego comercial.

La casa y sus dos plantas toman referencias espaciales de la
[reconstrucción de RoomSketcher](https://www.roomsketcher.com/blog/the-simpsons-floor-plan/).
Sus imágenes/modelos no se redistribuyen. Los edificios característicos,
señalización, distribución, cúpula y geometría de interiores son código propio.

Los peatones combinan A* sobre una malla de ocupación con llegada, separación y
previsión de tráfico inspiradas en los
[comportamientos de dirección de Craig Reynolds](https://www.red3d.com/cwr/steer/).
La planificación pertenece a NPC; no genera flechas ni una ruta para el jugador.

## Combate y referencia de GTA

Se revisó [la explicación de Rockstar sobre primera persona y controles](https://blog.playstation.com/archive/2014/11/04/grand-theft-auto-v-ps4-introducing-new-first-person-mode/) y [sus consejos de perspectiva](https://www.rockstargames.com/es/newswire/article/25o2411812oa29/rockstar-game-tips-playing-with-perspective-in-gtav). Se aplican cambio de perspectiva con un botón, cámara al hombro, movimiento relativo a cámara, vehículo utilizable, misión elegible y minimapa. No se incorpora código, audio ni recursos de GTA.

Los personajes de Springfield, equipo militar, extraterrestres y ovnis de [characters.js](characters.js) son geometría original estilizada. Para Kang y Kodos se consultaron las [figuras oficiales de Super7](https://super7.com/blogs/news/the-simpsons-ultimates-wave-3-figures) y [la imagen de referencia de Kodos](https://simpsonswiki.com/wiki/File:Kodos.png): ojo único, tentáculos, dientes y casco transparente. Las imágenes no se redistribuyen. El militar y los peatones reutilizan el rig y las animaciones CC0 de Kenney; no se afirma que los personajes propios sean modelos oficiales ni que pertenezcan a la licencia de Kenney.

La IA emplea A* y separación para perseguir sin cruzar edificios, rayos contra cajas para visibilidad y obstáculos, enfriamiento entre ataques y proyectiles con colisión continua. No usa servicios externos ni aprendizaje automático.
