# Cierre PSI-000: Productos, Servicios, Compras e Inventario

- Fecha: 2026-08-26
- Empresa autorizada de QA: Powerful Control System (`empresa_id=12`)
- Candidato funcional probado: `9eba6b12729baa1735f25dc0c38659292a88d19d`
- PR de cierre: `#215` (continuacion del endurecimiento integrado por `#212`)

## Veredicto

El modulo queda **preparado como candidato productivo de software**. El codigo,
las migraciones, el aislamiento multiempresa, los flujos reales de inventario,
la carga de factura XML con IA, la cola logica de impresion y la interfaz en
escritorio/movil aprobaron sus controles en staging.

Este cierre no afirma que ya exista un despliegue en produccion. Tampoco afirma
haber visto papel salir de impresoras fisicas: esa comprobacion requiere los
equipos de cocina y caja conectados. Son los dos controles operativos externos
que quedan antes del GO definitivo: autorizacion de merge/despliegue y prueba de
papel en cada dispositivo real.

## Resultado por frente

| Frente | Evidencia | Resultado |
| --- | --- | --- |
| Productos, categorias y proveedores | CRUD real, validacion economica y conflictos por codigo/nombre; nombres equivalentes por espacios o mayusculas devuelven 409 | OK |
| Servicios | Codigo opcional, rangos, ownership y ocho altas concurrentes del mismo nombre: una creada y siete 409 | OK |
| Bodegas | CRUD unificado, compatibilidad de ruta historica, stock, traslados y referencias ajenas como 404 | OK |
| Inventario | Entradas, salidas, ajustes, traslados, lotes, seriales, reservas, valorizacion, alertas y Kardex | OK |
| Compras | Recepcion multiitem atomica e idempotente, stock/costo/lote/Kardex y comprobante privado | OK |
| Factura con IA | XML sintetico procesado con GPT-5.5; tres duplicados reutilizaron la extraccion original sin otra llamada al modelo; la semilla demo fue retirada de UI y API | OK |
| Impresion | Dos impresoras, dispositivo, reglas de cocina/caja, productos, dos trabajos tomados y 20 formatos renderizados | OK logico/visual; papel fisico pendiente |
| Multiempresa | Bodega de otra empresa no puede leerse ni mutarse desde la empresa QA; origen y destino permanecen sin cambios | OK |
| Rendimiento y navegacion | Carga inicial diferida, una sola navegacion del catalogo y estado separado entre los submenus de Productos y Compras | OK |
| UX visual | 16 rutas, escritorio y movil; 32 paginas, 1.404 botones, 28 clics seguros y cero paginas con hallazgos | OK |

## Datos operativos finales de staging

- Catalogos visibles: 9 productos, 4 servicios, 4 categorias, 4 bodegas y 3
  proveedores.
- Inventario: 62 unidades, 41 movimientos de Kardex, 8 lotes activos, 2
  seriales registrados, sin reservas activas ni deficit.
- Impresion: funciones `cocina`, `factura_caja` y `ticket_cobro` asociadas; dos
  trabajos de QA quedaron en estado logico `impreso`, un intento y sin alerta.
- Compras/IA: el soporte original conserva extraccion y archivo privado; los
  soportes duplicados apuntan al original y no consumen una segunda extraccion.
- Semilla IA: el conteo de soportes permanecio en 23 y la accion retirada
  `seed_demo` devolvio 400 sin insertar datos.

Los valores detallados, sin credenciales, cookies, tokens, identificadores
fiscales, nombres de proveedores ni hashes privados, estan en
`validacion_operativa_final_9eba6b12.json`.

## Correcciones de produccion incluidas

- Stock, reservas y recepciones se serializan dentro de transacciones; la
  recepcion de compra agrupa hasta 500 productos y conserva idempotencia.
- Toda referencia hija se valida contra el `empresa_id` canonico. Un recurso
  inexistente o ajeno produce el mismo 404 seguro y no filtra su existencia.
- Categorias, productos, proveedores y servicios traducen carreras de unicidad
  a 409 estable. Categorias, proveedores y servicios tienen indices por nombre
  normalizado dentro de cada empresa y migraciones que fallan cerradas si hay
  colisiones historicas.
- Los errores de `RowsAffected` y `rows.Err()` dejaron de descartarse en los
  recorridos auditados.
- El XML privado exige contenido bien formado, rechaza DTD y se entrega como
  attachment. Una cadena duplicada se resuelve solo dentro del tenant y con
  limite/ciclo controlados.
- Soportes de Compras IA ya no ofrece ni acepta `seed_demo`: la empresa solo
  recibe archivos radicados por un usuario y operaciones auditables sobre ellos.
- La implementacion duplicada de bodegas fue retirada: la ruta historica
  redirige a la vista canonica y el carrito usa esa vista directamente.
- Los iframes de Productos y Compras difieren su primera carga hasta conocer la
  empresa; cada submenu conserva su ultima pagina en una clave propia y no
  restaura vistas del otro modulo.
- Soportes de Compras IA permite que paneles y celdas encojan dentro de los dos
  niveles de iframe. En el flujo operativo anidado, el documento paso de
  `966/791 px` a `791/791 px`; la tabla ancha conserva su desplazamiento interno
  (`920/728 px`) sin ensanchar la aplicacion.
- Las tablas del centro de impresion conservan scroll interno y no ensanchan el
  documento en escritorio ni movil.

## Gates ejecutados

- Preflight local profesional `-Full -Strict`: aprobado; pruebas Go completas,
  auditorias, sintaxis de 74 JavaScript y duplicados estructurales en cero.
- Prueba de la migracion normalizada contra PostgreSQL real: `PASS` dentro de
  transaccion temporal, sin alterar datos empresariales.
- Staging exacto `9eba6b12`: cinco servicios saludables, `/health=200`,
  `/ready=200`, cinco migraciones aplicadas, tres indices normalizados y cero
  grupos historicos duplicados.
- GitHub Professional CI `32974327264`: `success` para el SHA funcional exacto;
  aprobaron preflight/auditorias y seguridad, dependencias, contenedores y SBOM.
- GitHub E2E Visual QA `32974846684`: `success`; el reporte descargado confirma
  32 paginas, 1.404 botones, 28 clics seguros, cero hallazgos y 20/20 formatos
  imprimibles correctos.
- Release Candidate inmutable `32975347367`: `success` para `9eba6b12`; las
  imagenes exactas aprobaron construccion, escaneo, SBOM, publicacion por digest
  y validacion del compose de entrega.
- Matriz automatizada de roles: seis roles sin paginas o APIs faltantes.

## Alcance visual conservado

El subdirectorio `github-e2e-32974846684` conserva el reporte, las 32 capturas y
las matrices del SHA funcional exacto. Incluye Productos, categorias,
proveedores, bodegas, Kardex/historial, Inventario avanzado, Compras, Compras
avanzadas, soportes de compra con IA, configuracion de impresoras, recetas,
codigos de barras y carta publica, en escritorio y movil.

Las capturas manuales complementarias documentan el contenido inferior de los
flujos extensos, las reglas por funcion/producto, la ausencia de semilla demo y
la vista unificada de bodegas. `compras_ia_anidada_sin_overflow_desktop_9eba6b12.jpg`
documenta el flujo real completo dentro de Administrar empresa y Compras.

## Condiciones del GO

1. Promover exclusivamente un SHA descendiente del candidato probado, despues
   de que CI y el workflow de Release Candidate aprueben ese SHA exacto.
2. Obtener autorizacion explicita para merge y despliegue productivo.
3. En cada estacion fisica, imprimir un trabajo de cocina y uno de caja,
   confirmar papel, contenido, ancho y corte, y registrar el dispositivo.
4. Despues del despliegue, repetir `/health`, `/ready`, migraciones y un smoke
   autenticado de Productos, bodegas, Kardex, Compras IA e impresoras.

No se usaron credenciales en archivos, commits ni reportes. Produccion no fue
modificada durante PSI-000.
