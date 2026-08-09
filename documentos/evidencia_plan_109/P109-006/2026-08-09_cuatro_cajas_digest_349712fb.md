# P109-006 - Repeticion de cuatro cajas sobre el digest 349712fb

Fecha: 2026-08-09

Empresa: Powerful Control System (`empresa_id=12`)

Entorno modificado: staging aislado

Produccion: sin despliegue ni cambio de imagen

## Candidato exacto

El workflow inmutable `31334057174` construyo, escaneo, publico y valido
Compose para el commit completo
`349712fb32d029b997109edd4a2b18a6bb299331`. Staging recibio exclusivamente:

- API: `sha256:4c56f63b4e8ca61c1f80d39e4a99be4894a4e36b4594f1a7ab3c293907ff414c`;
- migrador: `sha256:35ac903ba7c873b178a10e0b941bfc4745cbe8c7ec5f4f7564b03f744ec32844`;
- worker: `sha256:94e91fff3824221ce0daf349a4841e3be1ea8a595ac0df6388a1ab1caed792ca`;
- frontend: `sha256:d98b31edbdc36df9686bb4983c075d88e2d3f03013eb51c61147155c0fdc4852`.

`/health`, `/ready` y el login publico respondieron HTTP 200. Antes y despues
se compararon las imagenes de API, worker y frontend de produccion y no
cambiaron.

## Cuatro identidades por flujo oficial

Los usuarios temporales `P108-C1` a `P108-C4`, IDs 648 a 651, estaban
confirmados e inactivos. Se activaron por el endpoint autenticado oficial y se
solicito recuperacion de clave individual desde el login de staging. El buzon
autorizado recibio los cuatro correos reales y cada token de un solo uso se
consumio desde la pantalla `Restablecer contrasena`.

Las cuatro credenciales abrieron sesiones independientes: 4/4 HTTP 200, cuatro
cookies distintas y cuatro tokens CSRF distintos. No se documentan correos,
claves, tokens, cookies ni enlaces de recuperacion.

## Ensayo simultaneo

Se abrieron en paralelo cuatro cajas nuevas, IDs 20, 23, 21 y 22. Cada sesion
creo un carrito independiente, agrego un servicio controlado de COP 100 y pago
una venta real de staging. Los medios fueron efectivo, debito, Nequi y
transferencia.

| Control | Resultado |
| --- | --- |
| Cajas abiertas | 4/4 HTTP 200 |
| Carritos creados | 4/4 HTTP 201 |
| Items agregados | 4/4 HTTP 201 |
| Ventas pagadas | 4/4 HTTP 200 |
| Comprobantes emitidos | 4/4, sin envio DIAN |
| Reintento del mismo pago | 4/4 rechazados HTTP 409 |
| Cajas cerradas | 4/4 HTTP 200 |
| Conciliacion | teorica = fisica, cero incidencias |

La caja de efectivo cerro en COP 100 teoricos y fisicos. Las otras tres
cerraron en COP 0 de efectivo, porque sus pagos fueron no efectivos. No se
emitio ni anulo una factura fiscal durante esta corrida.

## Impresion y cierre limpio

Uno de los comprobantes nuevos se abrio desde la vista servida de staging y se
imprimio con Chrome a PDF A4. Resultado: 74.865 bytes, una pagina, cero imagenes
rotas. La inspeccion visual confirma logo oficial, encabezado, metadatos, cliente,
resumen, cinco columnas, importes, control documental y observaciones sin
recortes. Archivo local no versionado:
`output/pdf/P109_staging_cuatro_cajas_349712fb.pdf`.

Al finalizar se desactivaron las cuatro cuentas mediante el endpoint oficial.
Las cuatro sesiones previas pasaron inmediatamente a HTTP 401 y no quedo caja
QA abierta.

## Veredicto

P109-006 queda **aprobada para el digest exacto 349712fb**. Esta evidencia
acredita una de quince fases de certificacion del candidato (6,7 %), pero no
autoriza produccion. P109-005 conserva pendientes de accesibilidad asistida,
tableta e impresion fisica; Mailu corporativo, UAT humano, piloto y las demas
compuertas P0/P1 mantienen el veredicto general **NO-GO**.
