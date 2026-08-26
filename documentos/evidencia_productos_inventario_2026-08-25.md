# Evidencia de preparación de Productos, Servicios e Inventario

Fecha: 2026-08-25  
Empresa operativa: Powerful Control System (`empresa_id=12`)  
Rama local: `codex/domotica-activation-queue`  
Commit base observado: `76a43b971ca53cb6dfe6ccf642faa6e41781c7ef`

## Alcance revisado

- Productos, categorías, servicios, bodegas, existencias, ajustes, traslados,
  conteos cíclicos, Kardex, costos, precios, lotes, seriales y reservas.
- Requisición, cotización, aprobación y recepción de compras.
- Carga automática de factura de compra y soportes de compras con IA.
- Recetas e impresión lógica de caja y cocina.
- Aislamiento por empresa, autorización, errores públicos, reintentos e
  idempotencia de mutaciones críticas.
- Revisión estática de 302 páginas empresariales: 5.832 controles, 2.877
  acciones, 2.955 entradas y 842 controles dinámicos. El inventario detallado
  está en `documentos/arquitectura/inventario_ui_plan_106.md`.

## Pruebas reales en PCS

Se usaron únicamente pantallas y APIs oficiales; no se alteraron datos mediante
SQL directo.

- Producto QA creado: ID 193, SKU `QA-PROD-20260825-421008`.
- Categoría QA creada: ID 26703, código `QA-CAT-421008`.
- Bodega QA creada: ID 49, código `QA-BOD-421008`.
- Servicio QA creado: ID 19, código `QA-SVC-20260825-421008`.
- Lote ID 4, serial ID 4 y reserva ID 4 creados y recorridos.
- Salida superior al stock y traslado superior al disponible fueron rechazados
  sin alterar existencias.
- Traslado válido de tres unidades entre bodegas generó sus dos movimientos de
  Kardex.
- Conteo cíclico actualizó la diferencia en backend; se corrigió localmente el
  campo usado por el mensaje visible.
- Confirmar dos veces una reserva fue rechazado en el segundo intento y no
  descontó inventario otra vez.
- Los dos productos que tenían existencias negativas (IDs 103 y 104) fueron
  regularizados a cero mediante conteo cíclico auditable. No quedaron
  existencias negativas en PCS durante la comprobación.
- El aislamiento se probó con una empresa inexistente: no se expusieron
  productos, servicios ni categorías de PCS.
- Receta QA creada con ingrediente real e impresora de cocina asignada.
- Dos trabajos de prueba fueron encolados y tomados: caja a `POS_80MM` y cocina
  a `AnyDesk Printer`. No se marcaron como impresos porque no hubo observación
  física del papel.
- El selector y su lista permitida se validaron visualmente, pero el fixture XML
  sintético no se transmitió ni procesó en la aplicación real durante este
  cierre; permanece como gate explícito de staging.

Los datos QA permanecen identificados por el sufijo `20260825-421008`; su
eliminación no forma parte de este cierre porque sería una acción destructiva.

## Defectos confirmados y reparaciones locales

1. La misma salida válida podía enviarse dos veces y descontaba stock dos veces.
   Las mutaciones críticas ahora exigen una llave durable de idempotencia,
   separada por empresa y operación.
2. La recepción avanzada de compras registraba documentos, pero no aumentaba
   existencias ni generaba Kardex. Ahora producto, item y bodega se validan por
   `empresa_id`; la recepción, costo, lote opcional, existencia y movimiento se
   escriben en una sola transacción.
3. La sobre-recepción se bloquea bajo `FOR UPDATE` y con actualización
   condicional; un documento repetido para la misma requisición se rechaza.
4. Las entradas de inventario usan un `INSERT ... ON CONFLICT DO UPDATE`
   atómico sobre la clave empresa/producto/bodega.
5. Requisiciones usan productos activos del catálogo; la recepción usa el item
   pendiente y una bodega activa, en vez de depender de nombres libres.
6. Cotización y aprobación validan pertenencia de requisición, cotización y
   proveedor dentro de la misma empresa y actualizan su estado
   transaccionalmente.
7. Se retiró de la interfaz de producción el botón que sembraba datos demo.
8. Las vistas Compras y Precios estaban anidadas dentro de una sección oculta
   por una etiqueta sin cerrar; la estructura quedó balanceada y se verificó
   localmente.
9. Errores HTML/crudos ya no se muestran al operador; inventario avanzado y
   compras presentan mensajes públicos saneados.
10. El resumen de inventario incluye totales de productos, productos por vencer,
    bodegas, servicios y categorías mediante agregados acotados por empresa.
11. La imagen de producto dejó de aceptar SVG para evitar contenido activo.
12. La toma de trabajos de impresión dejó de hacer selección y actualización
    separadas: ahora usa un CTE PostgreSQL con `FOR UPDATE SKIP LOCKED`, actualiza
    e incrementa intentos en una sola sentencia y devuelve exactamente las filas
    reclamadas por el agente.
13. Solo el agente que tomó un trabajo puede cerrarlo como impreso o error; la
    repetición del mismo cierre es idempotente y el reintento manual reinicia el
    contador para que un trabajo agotado vuelva a ser elegible.
14. Crear un trabajo de cola exige `Idempotency-Key` durable por empresa. La UI
    conserva la misma llave ante una respuesta ambigua, bloquea el botón durante
    el envío y no muestra HTML, SQL ni errores internos al operador.
15. El carrito tenía tres implementaciones equivalentes para resolver impresora
    de orden, cajón y ticket. Se consolidaron en un único resolver compartido y
    se retiraron funciones nombradas repetidas en las pantallas críticas.

## Matriz visual

Se recorrieron en escritorio y vista responsive Productos, Categorías,
Servicios, Bodegas, Precios, Compras, Inventario avanzado, Soportes de compras
IA, Configuración de impresoras y Recetas. Se comprobaron encabezados,
formularios, pestañas, tablas, botones, mensajes, estados vacíos, datos reales y
errores. La recepción reparada fue revalidada localmente: muestra producto
autoritativo, item pendiente, bodega destino, lote opcional y el botón
`Recibir y actualizar inventario`. Las reglas responsive conservan una columna
por campo bajo 620 px y tablas desplazables dentro de su contenedor.
La regresión visual final de Compras confirmó que el selector de factura es
visible y acepta solo PNG/JPEG/WebP/PDF/XML; un 404 HTML del servidor estático
se presenta como `La información de compras todavía no está disponible.` sin
exponer marcado ni detalles internos.

La repetición final sobre el código actual comprobó:

- Productos a 480 px: `innerWidth=480`, `scrollWidth=480`, sin IDs duplicados;
  resumen, Productos/Servicios/Abastecimiento y accesos a categorías, bodegas y
  MRP permanecen legibles.
- Inventario avanzado a 1440 x 900: `scrollWidth=1409`, sin desbordamiento de
  página ni IDs duplicados; quedaron visibles Operación, Lotes, Reservas y
  Valorización.
- Compras a 1440 x 900: `scrollWidth=1409`, selector de comprobante visible y
  botón `Carga automática de factura de compra` accesible.
- Configuración de impresoras: cola, caja/cajón, factura, asignación por
  funcionalidad, producto, categoría y computador permanecen accesibles; la
  vista móvil observada no presentó desbordamiento horizontal.

Capturas locales del candidato actual:

- `documentos/evidencia_productos_inventario_2026-08-25_visual/productos_mobile.jpg`
- `documentos/evidencia_productos_inventario_2026-08-25_visual/inventario_avanzado_desktop.jpg`
- `documentos/evidencia_productos_inventario_2026-08-25_visual/compras_carga_factura_desktop.jpg`
- `documentos/evidencia_productos_inventario_2026-08-25_visual/configuracion_impresora_mobile.jpg`

Estas capturas validan render e interacción visible del candidato local. No
sustituyen la prueba autenticada posterior al despliegue ni la observación del
papel físico.

## Validación técnica

- `go test ./... -count=1`: aprobado en la repetición final del árbol actual.
- `go test ./... -run '^$' -count=1`: compilación global aprobada.
- `go test ./db ./handlers -count=1`: aprobado.
- Pruebas enfocadas de productos, inventario, compras IA, impresoras,
  multiempresa, atomicidad e idempotencia: aprobadas.
- `go vet ./db ./handlers`: aprobado.
- Sintaxis de `web/js/compras_avanzadas.js` y
  `web/js/inventario_avanzado.js`: aprobada con Node.
- Los scripts inline de Productos, Compras, Compras avanzadas, Configuración de
  impresoras y Carrito aprobaron parseo con Node; también aprobaron los dos
  archivos JavaScript separados de inventario/compras avanzadas.
- Auditoría de duplicados: 322 rutas literales registradas y cero duplicadas;
  ocho páginas críticas con 889 IDs de marcado y cero repetidos; 789 funciones
  nombradas y cero repetidas.
- Los contratos de cola comprueban `SKIP LOCKED`, cierre por propietario,
  reintento elegible, idempotencia durable y saneamiento de errores.
- `cmd/pcs-disk-manager/TestParseDockerBytes` fue corregido para interpretar
  formatos Docker compactos o separados, bytes y unidades binarias; su paquete
  y `go vet` aprobaron.
- `git diff --check`: sin errores de espacios; solo avisos de normalización
  CRLF del árbol existente.
- `go test -race` no está disponible en este host porque Go fue construido sin
  CGO.
- No hay `psql`, Docker ni Podman en este host y el puerto local 5432 está
  cerrado; por ello no se ejecutó una prueba nueva de concurrencia contra
  PostgreSQL desechable.
- La primera suite global agotó C: al enlazar binarios. Se eliminó únicamente la
  caché recompilable de Go (19,3 GB) y la repetición completa dejó de fallar por
  espacio.

## Gate de producción

Estado actual: **candidato local aprobado para staging; NO-GO de producción**.

El código local y las pruebas enfocadas quedan preparados, pero no se puede
afirmar 100% en producción hasta completar todos estos gates sobre el mismo
candidato:

1. fijar un commit candidato sobre un árbol estable y ejecutar CI/preflight sin
   mezclar los numerosos cambios concurrentes presentes en el workspace;
2. ejecutar migraciones, rollback y pruebas de concurrencia en PostgreSQL de
   staging;
3. desplegar el candidato con aprobación explícita y repetir el recorrido real
   de recepción compra → existencia → Kardex → costo/lote;
4. repetir la carga real del fixture sintético
   `backend/handlers/testdata/qa_factura_compra.xml` en la aplicación desplegada
   y verificar extracción, borrador, rechazo de duplicado e impacto final;
5. ejecutar el agente local en caja y cocina, observar papel físico, confirmar
   cierre/reintento de la cola y decidir explícitamente si cada flujo usa agente
   o diálogo del navegador para evitar impresión duplicada;
6. medir catálogo, Kardex y recepción con volumen representativo y dos o más
   réplicas antes del GO;
7. autorizar y ejecutar la limpieza controlada de los registros QA identificados
   en este documento, verificando primero sus dependencias e historial auditable.

Este documento no afirma merge, CI, despliegue ni producción actualizados.
