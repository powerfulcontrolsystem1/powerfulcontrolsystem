# P109-003 - Centro de reportes en staging

Fecha: 2026-07-31
Entorno: staging autenticado, empresa Powerful Control System (`empresa_id=12`).

## Resultado

Se revisaron `reportes_menu.html` y `reportes_ejecutivos.html` en escritorio y
móvil. El catálogo cargó 46 reportes; filtros, selector, campos y tabla se
mantuvieron ordenados en ambas vistas. Se ejecutaron 12 clics limitados de
navegación/selección, sin confirmar, exportar, imprimir ni generar reportes.

No se observaron errores de página, consola ni respuestas HTTP 4xx/5xx en el
Centro de reportes. La captura visual confirmó que los controles móviles se
apilan sin desborde horizontal visible.

## Exclusión de URL histórica

`/administrar_empresa/reportes_finanzas.html` respondió 404, pero la ruta no
existe en el candidato ni en el árbol actual: fue retirada al unificar reportes
en `reportes_ejecutivos.html`. Se excluye como URL heredada, no como regresión.

## Límite de cierre

Siguen pendientes la conciliación contable completa, exportaciones oficiales,
ReportSpec IA, autorización IA y UAT por contador. P109-003 permanece parcial.

## Ampliación autenticada del 2026-08-01

Se ejecutó el catálogo completo contra la empresa PCS sobre un candidato
inmutable. La primera pasada obtuvo 44/46 reportes consistentes y aisló dos
incompatibilidades PostgreSQL concretas:

- `operativo_modulos_resumen` mezclaba timestamps con texto vacío dentro de
  `COALESCE`;
- `operativo_vehiculos_permanencia` mezclaba timestamps/texto y aplicaba
  `ROUND(double precision, 0)`, firma que PostgreSQL no ofrece.

Las fechas dinámicas y de permanencia se convierten ahora explícitamente a
texto; el valor calculado de permanencia se convierte a `NUMERIC` antes de
redondear. Las regresiones enfocadas de `handlers` y `db`, `go vet` y el
workflow de publicación aprobaron. El candidato final corresponde al commit
`f9694e104f5deabb336f25b6c885477d952e123a`, workflow `30730395189`, con estos
digests desplegados únicamente en staging:

- API: `sha256:695af344bd936e0e45a4b348b855c4fa60df41ed53b7e43edefb2f5d8ccf139c`;
- migrador: `sha256:5978aa7794c182143c24459c90b4c180a423d23a3272729fd25de444847de7f9`;
- worker: `sha256:f28a89ec7471b389e32077df3adbecf820f26dfa4b485538f122aa176742b96c`;
- frontend: `sha256:f20c87ca7db6cd7fb69e2ecc82ed9cdeee5eacf46677af145773b84a24b1d445`.

La repetición final recorrió los 46 datasets: **46/46 consistentes**, **46/46
con JSON, CSV, TXT, XLS y PDF**, 225 filas totales, cero alertas y **21/21
datasets contables, fiscales y de cartera** aprobados. Además se descargaron
por el endpoint oficial autenticado PDF reales de balance de prueba e impuestos,
XLS de libro auxiliar y CxP, y CSV de asientos: todos respondieron HTTP 200 con
MIME y firma esperados. `xlsx` respondió 400 porque el contrato publicado usa
el formato `xls`; al usar el formato del catálogo ambas exportaciones aprobaron.

## Conciliación directa de solo lectura

La verificación SQL de soporte, sin modificar asientos ni saldos, obtuvo:

- 7 eventos contables activos: 5 procesados y 2 eventos informativos sin
  partida (`venta_sesion_activada` y `proveedor_registrado`);
- 5 asientos, ninguno desbalanceado, huérfano o duplicado por evento;
- débitos 102,02 y créditos 102,02;
- 3 CxP activas por 214.302,00; pagos 27,02 y saldo 214.274,98;
- cero invariantes monetarias rotas, diferencias contra la suma de pagos,
  pagos huérfanos, movimientos faltantes o importes pago/movimiento distintos;
- 5 pagos y 5 eventos outbox sin hashes duplicados. Los cuatro eventos creados
  desde el 30 de julio están publicados. Permanece un `dead` histórico del 26
  de julio, generado antes de habilitar el handler del topic; su pago,
  movimiento y asiento sí existen. Debe recuperarse por el flujo administrativo
  auditado, no mediante edición SQL.

Staging y producción conservaron `health` y `ready` en HTTP 200. Producción
continuó con sus imágenes locales y no recibió este candidato.

## Estado de P109-003

El catálogo y la coherencia del subconjunto de datos PCS quedan demostrados,
pero la fase continúa **parcial**. Faltan fixtures empresariales suficientes
para devoluciones, notas crédito, anulaciones, cierres, moneda, impuestos y
retenciones; cierre/reapertura controlados; recuperación auditada del evento
histórico; y UAT firmada por un contador distinto del desarrollador.
