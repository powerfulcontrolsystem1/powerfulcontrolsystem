# Accesorios y juegos personales

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-14.

## Navegación y alcance

El [menú flotante](../web/menu.js) agrupa Calculadora y Juegos en Accesorios.
[Juegos](../web/juegos.html) muestra exclusivamente un menú de seis botones con
capturas de los juegos. El [contenedor flotante](../web/juegos/ventana.js) y sus
[estilos](../web/juegos/ventana.css) abren un iframe pequeño, movible, redimensionable
en PC y minimizable, sin bloquear la aplicación del fondo. Cada juego ocupa ese
mismo iframe; volver al menú pausa y guarda. Cerrar oculta el iframe y conserva
su estado en memoria. No se interceptan los atajos de la aplicación anfitriona.
El catálogo presenta primero GTA SIMSong, seguido de Pac-Man Canvas, Tetris,
Buscaminas, Solitario Klondike y la aventura submarina sorpresa. DOON se retira.
La calculadora conserva su ventana compacta. En móvil el mando semitransparente
se superpone al juego; no se abre una pestaña nueva.

La [sala](../web/juegos/sala.js), sus [estilos](../web/juegos/sala.css) y el
[núcleo de audio/estado](../web/juegos/core.js) coordinan una sola partida activa,
pausa automática al ocultar la página, teclado y botones táctiles, pantalla
completa y guardado automático cada 20 segundos, al pausar y al terminar.
Guardar requiere sesión y red; un fallo permanece visible y no se presenta como
guardado exitoso. Cerrar antes del próximo guardado puede perder ese intervalo.
Cargar y Nueva solicitan confirmación para reemplazar avances.

## Juegos y procedencia

- [Pac-Man](../web/juegos/pacman/engine.js): adaptación de
  [platzhersh/pacman-canvas](https://github.com/platzhersh/pacman-canvas/tree/531064a1c906b80abb046f918396e0f21d3ce5b2),
  diez niveles, cuatro fantasmas, persecución/dispersión, pastillas de poder y
  tres vidas. Conserva [CC0](../web/juegos/pacman/CC0.txt),
  [mapa](../web/juegos/pacman/map.json) y los SVG locales de fantasmas.
  Se retiraron jQuery, PHP, AppCache y las grabaciones externas. El reloj,
  sonido sintético y serialización son propios de PCS.
- [Clásicos](../web/juegos/clasicos.js): siete piezas, bolsa equilibrada,
  reserva, vista previa, pieza fantasma, giros y niveles para Tetris;
  primera apertura segura, banderas, apertura vecina y expansión de vacíos para
  Buscaminas; robo de una carta, secuencias alternadas, bases por palo,
  auto a bases y deshacer para Solitario.
- [GTA SIMSong](../web/juegos/simsong/game.js): ciudad de exploración libre,
  caminar/correr, salto y gravedad, vehículos utilizables, peatones animados,
  cúpula transparente con límite físico, casa de Homero de dos plantas y Moe
  accesibles, más la oficina de Burns en la central. Las tres misiones de descubrimiento se completan en cualquier
  orden y permiten seguir explorando al concluir. No hay flechas ni camino guiado.
  [Datos del mundo](../web/juegos/simsong/world-data.js),
  [ciudad](../web/juegos/simsong/city.js),
  [interiores](../web/juegos/simsong/interiors.js),
  [física y planificación de NPC](../web/juegos/simsong/simulation.js) y
  [carga de recursos](../web/juegos/simsong/assets.js) separan escena, reglas y
  estado. A* evita edificios; separación y previsión de tráfico complementan
  las rutas de los peatones. Se guardan posición, interior, vehículo y misiones.
  El militar usa láser (F/clic), salto, carrera y tres cámaras (V): primera persona,
  tercera media y tercera lejana. El mando móvil añade Láser y Cámara.
  [Reglas de combate](../web/juegos/simsong/combat.js) y
  [simulación de la invasión](../web/juegos/simsong/invasion.js) coordinan doce
  extraterrestres (90 de salud, 250 puntos) y tres ovnis (300 de salud, 1000 puntos).
  El láser inflige 30 de daño. A* y separación guían la persecución terrestre;
  línea de visión, enfriamiento y colisión continua limitan los disparos enemigos.
  Los ovnis patrullan en altura y sus pilotos son visibles bajo cabinas transparentes.
  Hay tres vidas, caída y alma ascendente durante cuatro segundos; después se
  reaparece con protección temporal, o termina la partida al agotar las vidas.
  Eliminar la invasión concede 2000 puntos una sola vez y mantiene exploración libre.
  El guardado privado conserva salud, vidas, cámara, enemigos, proyectiles y progreso;
  las partidas anteriores reciben valores de combate por defecto sin perder avance.
  [Personajes](../web/juegos/simsong/characters.js) contiene adaptaciones estilizadas
  originales: Homero, Marge, Bart, Lisa y Maggie en casa; Moe y Barney en la taberna;
  Burns y Smithers en la central; Apu, Flanders, Skinner, Krusty y Milhouse en la
  ciudad, y Gorgory sentado en la patrulla. Este es el elenco inicial, ampliable.
  El equipo militar se monta sobre el rig CC0 y recolorea ropa en el shader.
- [Mundos](../web/juegos/mundos.js) y [renderizador](../web/juegos/render3d.js):
  WebGL nativo para la aventura sorpresa, que
  incluye exploración, oxígeno recargable en superficie, diez tesoros, fauna
  animada y pulso localizador. El código de combate y selva retirado ya no forma
  parte de este motor.
- [Créditos visibles](../web/juegos/creditos.html) delimitan autoría y licencias.
  Se investigaron también [Freedoom](https://freedoom.github.io/about.html) y
  [Dwasm](https://github.com/GMH-Code/Dwasm); no se distribuyen sus recursos.

GTA SIMSong usa Three.js r186/0.186.0, GLTFLoader, SkeletonUtils y
BufferGeometryUtils, autorizados expresamente por el usuario. La
[procedencia](../web/juegos/simsong/PROVENANCE.md) detalla versiones, licencias,
modelos CC0 de Kenney, conversiones y referencias de Springfield. El
[manifiesto](../web/juegos/simsong/assets/manifest.json) permite verificar los
recursos mediante [el comprobador](../tools/simsong_assets.mjs). La dependencia
es exclusivamente web y local, sin cambios en go.mod ni peticiones a CDN.
La alternativa sin terceros sería extender el renderizador WebGL existente;
se adopta Three.js para cargar GLTF con rigs y mezclas de animación mantenibles,
capacidades que Go estándar no proporciona en el navegador. Impacto: descarga
adicional del motor/modelos al seleccionar GTA, memoria GPU y mantenimiento de
la versión vendorizada; los clásicos no cargan ese motor.

La ciudad es una primera interpretación compacta con geometría original y
modelos CC0 y personajes originales estilizados; no replica exactamente un mapa
oficial ni contiene modelos oficiales de los personajes. La estética es estilizada, no fotorealista.
WebGL y Web
Audio deben estar disponibles. Pantalla completa depende de la API del navegador;
su ausencia se informa y se conserva la vista adaptable. La emulación móvil de
Chromium no certifica rendimiento ni compatibilidad en cada teléfono físico.

## Identidad, datos y API

[JuegosHandler](../backend/handlers/juegos.go) expone `/api/juegos` con sesión
activa. Es un accesorio personal gratuito, sin permiso operativo ni licencia
empresarial: no accede a caja, ventas ni datos comerciales. Esta excepción al
alcance empresarial es explícita y no concede capacidades a otros endpoints.

La clave privada se calcula en servidor: `admin:<id>` para administradores,
`empresa:<empresa_id>:usuario:<id>` para usuarios operativos activos y confirmados
de empresa activa. Un correo coincidente no une estas identidades. El middleware
permite que la ruta personal valide al usuario operativo sin exigir un
administrador homónimo. Sesiones expiradas, usuarios desactivados y empresas
inactivas no pueden leer ni guardar partidas mediante el handler.

| Petición | Contrato |
| --- | --- |
| `GET ?juego=<id>` | Partida privada, versión y nombre de la sesión |
| `GET ?juego=<id>&action=records` | Primeros 20 récords globales del juego; solo nombre, puntaje y fecha |
| `PUT ?juego=<id>` | JSON `version`, `puntaje`, `estado`; mismo origen y CSRF |

Solo se aceptan los seis IDs del catálogo. Se rechazan parámetros de identidad,
empresa o campos desconocidos, JSON adicional, estados incompatibles, puntajes
negativos/fuera del máximo y cargas de más de 512 KiB. El puntaje declarado debe
coincidir con el estado. El juego cliente calcula el resultado: **la clasificación
es recreativa y no tiene validación antitrampas por replay en el servidor**.
No usar estos resultados para premios, pagos ni decisiones laborales.

La migración aditiva `20260914-001-juegos-personales-v1` del catálogo super crea
`juegos_partidas` y `juegos_records`; implementación en
[juegos.go](../backend/db/juegos.go). La nueva
`20260914-002-simsong-v1`, en [juegos_simsong.go](../backend/db/juegos_simsong.go),
amplía el CHECK para `simsong` conservando filas antiguas de `selva`; ese ID ya
no está admitido por la API ni el catálogo. No se reescribe el checksum anterior.
Solo el migrador aplica DDL en runtime.
El guardado y el récord se actualizan en una transacción. Una versión antigua
produce 409 sin sobrescribir. Un puntaje inferior o empatado conserva al titular
y su fecha original; desempate global por fecha ascendente y clave estable.
Las fechas nacen en PostgreSQL (TIMESTAMPTZ) y se muestran en la zona del navegador.
La proyección compartida no expone correo, propietario interno, empresa ni JSON
de partidas. Nombres que parecen correos se sustituyen por Jugador PCS.

## Verificación

[Pruebas de handler](../backend/handlers/juegos_test.go) verifican sesión,
payload e identidad visible. [Pruebas PostgreSQL](../backend/db/juegos_test.go)
requieren `PCS_JUEGOS_TEST_DSN` apuntando exclusivamente a una base efímera;
ejercitan la migración nueva, privacidad, ranking compartido, conservación de
fecha y conflicto de ocho escritores concurrentes. Nunca apuntar esa variable
a producción. [Pruebas de reglas](../tools/juegos.test.mjs) cubren las transiciones
de los clásicos y serialización sin depender de puntuaciones de cuentas reales.
[Pruebas de Springfield](../tools/simsong.test.mjs) cubren colisiones, límite de
la cúpula, salto/gravedad, A*, evitación de tráfico y accesibilidad de entradas.
[Pruebas de combate](../tools/simsong_combat.test.mjs) comprueban impactos múltiples,
puntos únicos, tres vidas y animación completa, restauración y oclusión por paredes.
[La prueba de migración de Springfield](../backend/db/juegos_simsong_test.go) comprueba que los guardados retirados se conservan y que el CHECK admite el nuevo juego.
La aceptación visual incluye controles, pausa, guardado/carga, audio activado y
pantalla completa en escritorio y viewport móvil. La evidencia de cada candidato
se conserva fuera de Git; no se deduce despliegue de la mera existencia del código.
