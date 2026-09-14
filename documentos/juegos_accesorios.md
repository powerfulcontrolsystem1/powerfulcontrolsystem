# Accesorios y juegos personales

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-14.

## Navegación y alcance

El [menú flotante](../web/menu.js) agrupa Calculadora y Juegos en Accesorios.
[Juegos](../web/juegos.html) abre una pestaña nueva con `noopener`. Sus seis
secciones se apilan en PC y móvil: Pac-Man Canvas, Tetris, Buscaminas, Solitario
Klondike, DOON Selva viva y una aventura submarina sorpresa en 3D.
La calculadora conserva su ventana compacta.

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
- [Mundos](../web/juegos/mundos.js) y [renderizador](../web/juegos/render3d.js):
  WebGL nativo, sin bibliotecas ni binarios externos. Selva incluye tres armas,
  tres oleadas, vidas, suministros, escudos, poder radial, búsqueda de caminos
  alrededor de obstáculos y enemigos cuerpo a cuerpo/a distancia. La sorpresa
  incluye exploración, oxígeno recargable en superficie, diez tesoros, fauna
  animada y pulso localizador. Son juegos originales, no un port de DOOM.
- [Créditos visibles](../web/juegos/creditos.html) delimitan autoría y licencias.
  Se investigaron también [Freedoom](https://freedoom.github.io/about.html) y
  [Dwasm](https://github.com/GMH-Code/Dwasm); no se distribuyen sus recursos.

La estética de los mundos es 3D procedural con modelos articulados. No se
presenta como fotorealismo ni como animación capturada de actores. WebGL y Web
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
[juegos.go](../backend/db/juegos.go). Solo el migrador aplica DDL en runtime.
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
La aceptación visual incluye controles, pausa, guardado/carga, audio activado y
pantalla completa en escritorio y viewport móvil. La evidencia de cada candidato
se conserva fuera de Git; no se deduce despliegue de la mera existencia del código.
