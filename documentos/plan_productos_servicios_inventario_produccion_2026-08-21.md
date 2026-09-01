# Plan de preparación para producción: Productos, Servicios e Inventario

Fecha de creación: 2026-08-21  
Estado inicial: **NO-GO / 0 % certificado**  
Modelo planificador: **GPT-5.6 Sol, razonamiento alto**  
Modelo ejecutor solicitado: **GPT-5.6 Terra, razonamiento medio**  
Repositorio: `D:\powerfulcontrolsystem`

## 1. Objetivo y límite de autorización

Preparar y certificar para producción el núcleo multiempresa de productos,
servicios e inventario de PCS, incluyendo sus fronteras con ventas, compras,
reportes, IA, API móvil, catálogo público, MRP y almacenamiento de imágenes.

Este documento es únicamente el plan. Su creación no autoriza implementar,
desplegar ni modificar datos. La futura ejecución puede usar la cuenta de prueba
autorizada por el propietario únicamente desde el canal seguro y en la empresa
`Powerful Control System`; la contraseña no se copia a comandos, evidencias,
documentación, logs ni commits.

La autorización de pruebas no equivale a autorización para:

- ejecutar `rs` o promover producción;
- emitir compras, pagos, documentos fiscales o mensajes externos;
- borrar datos reales, purgar imágenes o aplicar migraciones en producción;
- usar SQL directo para crear efectos de negocio.

Esas acciones requieren verificar el ambiente y la autorización concreta justo
antes de ejecutarlas. Los efectos de negocio de QA se realizan por UI o API
oficial y se concilian después por las mismas superficies; SQL se limita a
validaciones de solo lectura o bases efímeras.

## 2. Base observada al crear el plan

- Rama observada: `codex/domotica-activation-queue`.
- SHA observado: `76a43b971ca53cb6dfe6ccf642faa6e41781c7ef`.
- El árbol contiene cambios ajenos a este módulo en pantallas administrativas y
  archivos CSS. No deben mezclarse, descartarse ni recibir crédito en este plan.
- El Plan 109 global continúa en `NO-GO`; este plan es una compuerta modular y
  no reemplaza la decisión global de producción.
- El inventario estático previo registra 162 controles en
  `administrar_productos.html`, 37 en Inventario avanzado, 22 en Recetas, 12 en
  Bodega, cinco en Historial y seis en Generador de códigos. Es una línea base,
  no evidencia funcional.
- Existen pruebas unitarias y estáticas puntuales de importación/exportación,
  vencimientos, bodega, inventario avanzado, permisos y API móvil. No se observó
  una batería PostgreSQL integral que demuestre todos los CRUD, invariantes,
  carreras, reintentos, restauración y flujos visuales del módulo.

### Hallazgos estáticos P0 que Terra debe confirmar primero

Estos puntos son hipótesis basadas en lectura del candidato observado. No se
declaran fallos productivos hasta reproducirlos en una base PostgreSQL aislada:

1. Las reservas avanzadas separan la consulta de disponibilidad de la creación;
   confirmar sobre-reserva concurrente, bloqueo e idempotencia.
2. La confirmación de reserva, las salidas, los traslados y los cambios de
   producto deben demostrar que dos transacciones no dejan stock negativo ni
   descuentan dos veces. Revisar `FOR UPDATE`, updates condicionales,
   `RowsAffected`, orden estable de locks y claves idempotentes.
3. El conteo cíclico registra el movimiento y el acta en operaciones separadas;
   confirmar atomicidad, reintento y competencia con una venta o traslado.
4. Repetir una entrada de lote o serial puede actuar como `UPSERT`; definir qué
   repetición es actualización válida y cuál debe ser conflicto o replay
   idempotente.
5. Las escrituras de servicios requieren validar rangos, duplicados, referencias
   en carritos/documentos y política de borrado lógico antes de permitir
   eliminación física.
6. Varias respuestas de `productos.go`, inventario avanzado y servicios aún
   concatenan `err.Error()`; los errores 5xx no deben exponer SQL, rutas ni PII.
7. `EnsureEmpresaProductosSchema` e inventario avanzado pertenecen al inventario
   de DDL heredado; API y worker no deben modificar esquema en producción.
8. La carga pública de imágenes acepta extensiones, incluido SVG. Confirmar
   contenido real, XSS, límites, nombres, aislamiento, reemplazo atómico y
   comportamiento con dos réplicas.
9. Las imágenes siguen en almacenamiento local público. No aprobar réplicas
   hasta demostrar almacenamiento compartido/objeto, migración, checksum,
   restore y lectura consistente.
10. La resolución genérica de permisos de inventario trata los POST como `C`.
    Confirmar que confirmar salida, cargar demo, ajustes y demás acciones
    críticas exijan la acción efectiva correcta, no solo visibilidad de página.
11. Revisar uso de `REAL` en precios, costos y cantidades. Conservar cantidades
    fraccionarias donde la unidad lo permita, pero definir tipos y reglas de
    redondeo que no produzcan deriva monetaria o de existencias.
12. Confirmar claves foráneas, unicidad, filas huérfanas, duplicados históricos,
    stock negativo, diferencias existencias/Kardex/costos/lotes y referencias
    cruzadas entre empresas antes de migrar o endurecer constraints.

## 3. Alcance funcional obligatorio

### 3.1 Núcleo incluido

- productos: alta, lectura, edición, activación/inactivación, duplicados,
  vencimiento, impuestos, unidades, costo, precio, imagen y campos obligatorios;
- servicios: CRUD, estado, precio, costo referencial, duración, impuestos y uso
  desde el carrito;
- categorías, proveedores como dependencia del producto y bodega principal;
- bodegas, existencias, Kardex, ajustes, devoluciones, pérdidas, traslados,
  cambio de producto, conteos cíclicos y políticas promedio/PEPS;
- alertas, resumen, tendencias, balance por bodega, proyección de quiebre y plan
  de reposición hasta la frontera con órdenes de compra;
- lotes, seriales, reservas, confirmación, vencimientos, calidad y valorización;
- recetas vendibles, ingredientes, versionado, costos y descuento de existencias;
- historial de precios, importación/exportación, etiquetas/códigos de barras y
  salidas Carta/POS;
- menús, tutorial, permisos, licencias, auditoría, accesibilidad y responsive.

### 3.2 Fronteras incluidas como regresión

- carrito de venta directa y estaciones, pagos concurrentes y devolución de
  ítems antes de pagar;
- ventas offline e idempotencia de sincronización;
- compras, recepciones y devoluciones de proveedor que afecten existencias;
- MRP/WMS sin crear un inventario paralelo;
- catálogo/carta pública y `venta_publica` sin filtrar costo o stock privado;
- `/api/v1/empresa/productos` y consumidores móviles existentes;
- Centro IA, chat y propuesta confirmable de alta de producto sin SQL libre;
- reportes de inventario/Kardex y eventos contables asociados;
- imágenes públicas y cualquier job, cache o archivo relacionado.

### 3.3 Fuera de alcance, salvo regresión de frontera

- certificar integralmente CxP, DIAN, pagos, nómina, logística o MRP;
- aplicación móvil nativa; se mantiene web responsive/PWA según Plan 109;
- rediseñar todo el ERP o agregar dependencias externas;
- crear una segunda fuente de existencias, productos o servicios.

## 4. Reglas obligatorias para GPT-5.6 Terra medio

1. Usar exactamente `gpt-5.6-terra` con razonamiento `medium`.
2. Antes de actuar, leer completos `AGENTS.md`, contexto general, contexto
   específico, contexto Codex, comandos, decisiones técnicas, checklist
   multiempresa, mapa de módulos, flujos operativos, estructura BD, matriz de
   roles, este plan y la evidencia de la fase.
3. Revalidar rama, SHA, `git status`, ambiente y versión desplegada. Crear un
   worktree/rama limpia `codex/` desde el candidato aprobado; no trabajar sobre
   los cambios ajenos observados.
4. No agregar dependencias, no cambiar `go.mod`, no usar otro motor y no
   introducir Python como runtime.
5. Todo ID empresarial y secundario se deriva o valida contra el
   `TenantContext`; toda consulta, join, cache, archivo y job conserva
   `empresa_id`.
6. Usar migraciones catalogadas, checksum y migrador propietario. API y worker
   solo verifican esquema y fallan cerrado.
7. Corregir por causa, agregar prueba de regresión y actualizar documentación en
   el mismo bloque. No hacer reemplazos globales de errores o SQL.
8. Un timeout es inconcluso. Una prueba estática no aprueba runtime, una UI
   visible no aprueba autorización y staging no aprueba producción.
9. No declarar porcentaje por archivos editados. Solo cuentan compuertas con
   evidencia del mismo SHA/digest.
10. Al hallar un P0, bloquear las fases dependientes y continuar únicamente con
    frentes independientes seguros.

## 5. Convención de evidencia y estados

Cada fase crea evidencia bajo:

`documentos/evidencia_productos_inventario/<ID>/<fecha>_<descripcion>.md`

Antes de cada ciclo registrar: ID, objetivo, SHA/digest, ambiente, empresa,
rol, datos, efectos externos, pruebas, rollback y aceptación. Al cerrar registrar:
archivos/rutas/tablas, comandos y resultados, capturas saneadas, aislamiento,
datos QA creados/limpiados, riesgos y estado.

Estados permitidos: `pendiente`, `parcial`, `aprobado`, `bloqueado`, `fallido`.
Solo `aprobado` abre la fase dependiente.

## 6. Fases de ejecución

### PSI-000 — Gobierno, candidato e inventario trazable [P0]

Acciones:

1. Partir de un worktree limpio y fijar SHA, rama, CI y alcance de release.
2. Generar matrices ruta/método/action → handler → función DB → tablas → rol →
   licencia → página/control → efecto → auditoría.
3. Inventariar controles dinámicos, fetches, impresiones, archivos, IA, jobs y
   consumidores transversales; reconciliar el inventario estático vigente.
4. Clasificar cada entrada como incluida, frontera o exclusión firmada.
5. Crear un catálogo de datos QA con prefijo único y reverso por flujo oficial.

Aceptación: árbol limpio, candidato inmutable, inventario completo sin rutas o
controles huérfanos y matriz de alcance aprobada. Sin esto, 0 % certificado.

### PSI-001 — Modelo PostgreSQL, migraciones e integridad [P0]

Acciones:

1. Levantar PostgreSQL efímero con esquema vacío, upgrade desde snapshot
   anonimizado y catálogo de migraciones exacto.
2. Auditar tipos, defaults, nullability, checks, índices, unicidad y relaciones
   compuestas por `empresa_id` para todos los IDs secundarios.
3. Crear reportes de solo lectura de duplicados, huérfanos, stock negativo y
   reconciliación entre existencias, movimientos, costos, lotes y reservas.
4. Extraer DDL de productos/inventario del runtime hacia migraciones nuevas,
   idempotentes y con rollback ensayado; no reescribir migraciones aplicadas.
5. Definir precisión monetaria y de cantidades por unidad. Probar COP, otras
   monedas, peso decimal, límites grandes y redondeo acumulado.

Aceptación: base vacía y upgrade pasan; API/worker sin DDL; cero anomalías P0 o
plan de conciliación aprobado; migración, rollback y checksum demostrados.

### PSI-002 — Seguridad, permisos, licencia y errores [P0]

Acciones:

1. Aplicar la checklist a todas las rutas de productos, servicios, recetas,
   bodegas, inventario, inventario avanzado, imágenes y reposición.
2. Probar sesión ausente, expirada, CSRF, rol insuficiente, módulo sin licencia,
   empresa inactiva y discrepancias de `empresa_id` en query/header/body/form.
3. Probar IDs de empresa B en producto, categoría, proveedor, bodega, lote,
   serial, reserva, receta, ingrediente, movimiento, conteo e imagen.
4. Separar `R/C/U/D/A` por riesgo real; confirmar que cajero solo consulte el
   catálogo auxiliar y que roles de bodega no eliminen ni accedan a ventas/caja.
5. Sanear todos los 5xx con `request_id`; mantener mensajes de validación solo
   desde taxonomía local segura y probar canarios de SQL, ruta, token y correo.
6. Auditar exportaciones, reportes, cachés, archivos y jobs contra cruce A/B.

Aceptación: matriz A/B completa con 403/404 estable, cero fuga de datos/errores,
permiso efectivo correcto y cero mutación crítica sin autorización servidor.

### PSI-003 — Atomicidad, idempotencia y concurrencia [P0]

Acciones:

1. Definir invariantes: stock nunca negativo, una sola existencia por
   empresa/producto/bodega, Kardex 1:1, suma de lotes/reservas conciliable y una
   transición válida por estado.
2. Serializar salidas, traslados, cambios, conteos y confirmaciones con locks y
   updates condicionales; validar `RowsAffected` y evitar deadlocks con orden de
   bloqueo estable.
3. Definir claves idempotentes/replay para altas, importación, lote, serial,
   reserva, confirmación, ajuste, conteo y orden de reposición.
4. Ejecutar carreras PostgreSQL reales: 20 solicitudes sobre el último stock,
   transferencias cruzadas, venta vs. conteo, reserva vs. pago, confirmación
   duplicada, importación repetida y replay offline.
5. Interrumpir proceso/worker/DB entre pasos y demostrar rollback completo.

Aceptación: cero sobreventa/stock negativo/doble movimiento, replay devuelve el
mismo resultado, deadlocks controlados y cada operación es atómica y auditable.

### PSI-004 — Productos, Servicios y catálogos [P0]

Acciones:

1. Cubrir CRUD/estado de productos, servicios, categorías y dependencias de
   proveedor/bodega con validaciones de bordes, duplicados y `RowsAffected`.
2. Preferir inactivación cuando haya historial; bloquear borrado físico de
   entidades referenciadas y documentar retención.
3. Probar campos obligatorios por empresa, impuestos activos, unidades, precios,
   costos, vencimiento, SKU/código de barras y búsquedas/paginación.
4. Probar importación CSV con BOM/UTF-8, delimitadores, archivos grandes,
   duplicados internos/externos, fila parcial, reintento y resumen por fila.
5. Probar exportación CSV/JSON/HTML sin costo interno para roles o superficies
   que no deban verlo y sin fórmulas peligrosas al abrir en hoja de cálculo.
6. Endurecer imágenes por contenido real, dimensiones/tamaño, nombres aleatorios,
   SVG/HTML activo, reemplazo, limpieza diferida y acceso A/B.

Aceptación: CRUD e importación/exportación pasan por rol/tenant; ningún borrado
rompe históricos; archivos no ejecutan contenido activo ni se pierden al fallar.

### PSI-005 — Inventario operativo y avanzado [P0]

Acciones:

1. Probar Bodega 1 idempotente, alta/estado/borrado de bodegas y responsable de
   bodega asignado, incluida restricción a su bodega real.
2. Cubrir entrada, salida, devolución, pérdida, ajuste, traslado, cambio y conteo
   con Kardex, existencia y costo reconciliados.
3. Validar promedio y PEPS con entradas a costos distintos, consumo parcial,
   transferencia y devolución; comparar resultados esperados con precisión.
4. Cubrir alertas, KPIs, tendencias, balance, quiebre y reposición en bordes de
   fecha/zona horaria, filtros, paginación y volumen.
5. Cubrir lote, serial, calidad, garantía, vencimiento, reserva, expiración,
   cancelación/confirmación y valorización. Si faltan transiciones operativas,
   implementarlas antes de certificar.
6. Conciliar cada resultado con reportes y evento contable aplicable.

Aceptación: todas las invariantes contables/de stock cuadran, trazabilidad
completa por actor/referencia y cero diferencia sin explicación aprobada.

### PSI-006 — Recetas e integraciones del núcleo [P0]

Acciones:

1. Probar recetas con ingredientes de la misma empresa, versiones, costos,
   activación, edición concurrente y bloqueo de borrado con carritos abiertos.
2. Probar carrito con producto, servicio y receta: agregar, cambiar cantidad,
   devolver, cancelar, pagar, reintentar y verificar reserva/descuento exactos.
3. Ejecutar cuatro cajas y al menos dos carritos compitiendo por stock limitado;
   validar estados, caja e impresión sin doble descuento.
4. Repetir sincronización offline y API v1 con idempotencia y contrato estable.
5. Validar compras/recepciones/devoluciones, MRP y WMS solo como productores o
   consumidores del mismo inventario canónico.
6. Validar IA: lectura autorizada, propuesta de producto, confirmación humana,
   hash/idempotencia, flags de escritura y rechazo de `empresa_id` inyectado.

Aceptación: una sola regla canónica de existencias; POS/offline/móvil/IA no
duplican efectos; servicios no descuentan stock salvo contrato explícito.

### PSI-007 — UX, responsive, accesibilidad y todos los controles [P1]

Acciones:

1. Recorrer todos los controles estáticos y dinámicos en escritorio, tableta y
   móvil para admin, inventario, jefe/responsable de bodega, compras, cajero,
   auditor y rol negado.
2. Cubrir teclado, foco, etiquetas, lector, contraste, zoom 200 %, errores,
   carga, vacío, offline, sesión expirada y doble clic.
3. Validar formularios largos, tablas, filtros y modales sin scroll horizontal
   ni pérdida de contexto; mantener tema claro/oscuro y UTF-8.
4. Probar `Cargar carta/precios con IA`, tutorial, enlaces MRP, carta pública y
   navegación del menú con licencia/permiso.

Aceptación: 100 % de controles incluidos en PASS o exclusión firmada; cero 5xx,
errores de consola, botones sin resultado, desbordes críticos o mutaciones sin
confirmación/auditoría.

### PSI-008 — Reportes, exportes, etiquetas e impresión [P1]

Acciones:

1. Validar Kardex valorizado, existencias, historial de precios y reportes de
   reposición contra el origen transaccional.
2. Probar CSV/JSON/HTML/PDF y nombres largos, UTF-8, 0/1/muchas filas, montos y
   cantidades extremas, paginación y filtros A/B.
3. Probar Carta, POS 80 mm y etiquetas/códigos en vista previa; confirmar blanco
   y negro, márgenes, escala, corte, código legible y sin páginas vacías.
4. Ejecutar impresión física solo con autorización y dispositivo del piloto.

Aceptación: cifras y filas coinciden 1:1, ningún archivo mezcla empresas y cada
formato incluido tiene evidencia visual saneada.

### PSI-009 — Rendimiento, observabilidad y recuperación [P0]

Acciones:

1. Crear dataset representativo por empresa y medir p50/p95/p99, 5xx,
   PostgreSQL, CPU, memoria y bloqueos en lecturas, escrituras y reportes.
2. Cumplir SLO vigente: API crítica p95 <= 1200 ms, 5xx < 1 % por 15 minutos y
   recursos del VPS bajo sus umbrales documentados.
3. Revisar planes de consulta e índices sin ocultar consultas N+1; probar límites
   y paginación con catálogo grande.
4. Añadir métricas/alertas de stock negativo bloqueado, conflictos, replays,
   latencia, fallos de importación y divergencia de Kardex.
5. Respaldar y restaurar PostgreSQL más imágenes en ambiente aislado; medir RTO
   y RPO y verificar checksums, permisos, tenant A/B y lectura con dos réplicas.

Aceptación: SLO, alertas, restore y rollback aprobados; ninguna réplica depende
de almacenamiento local no compartido para datos durables.

### PSI-010 — Staging equivalente y prueba real controlada PCS [P0]

Acciones:

1. Desplegar en staging únicamente el digest que aprobó CI, migraciones, SBOM y
   escaneo. Confirmar producción intacta.
2. Usar la empresa PCS y resolver su ID actual por flujo autenticado, sin confiar
   en un número histórico. Crear datos con prefijo único y reverso documentado.
3. Ejecutar el recorrido completo: categoría, bodega, producto, servicio,
   receta, imagen, import/export, lote/serial/reserva, ajuste/conteo/traslado,
   venta concurrente, devolución, Kardex, reporte y etiqueta.
4. Ejecutar tenant A/B con una segunda identidad empresarial no global.
5. Verificar auditoría, logs saneados, consola/red, métricas, limpieza oficial y
   reconciliación final; no dejar carritos, reservas o stock QA ambiguos.

Aceptación: recorrido real PASS en el mismo digest, A/B PASS, datos conciliados
y reverso completo. Cualquier efecto externo no autorizado queda `BLOCKED`, no
se simula como aprobado.

### PSI-011 — Release, piloto y decisión GO/NO-GO [P0]

Acciones:

1. Ejecutar preflight completo, CI, `go test`, `go vet`, build, race en Linux,
   validaciones JS/HTML, seguridad, migración y rollback sobre el SHA final.
2. Publicar matriz firmada de funciones, roles, licencias, riesgos aceptados,
   responsables, SLO/RPO/RTO, soporte y rollback.
3. Promover el mismo digest por el proceso canónico solo tras autorización
   explícita; ejecutar smoke autenticado y observar la ventana acordada.
4. Revertir ante fuga multiempresa, stock negativo, doble movimiento, pérdida de
   imagen, error de migración, 5xx sostenido o incumplimiento de SLO.

Aceptación: todas las fases aprobadas, cero P0/P1 abierto y firma GO. Una sola
compuerta P0 abierta mantiene **NO-GO**.

## 7. Matrices mínimas que debe producir Terra

1. **Superficie:** ruta, método, action, handler, función DB, tablas, control UI.
2. **Seguridad:** rol, R/C/U/D/A, licencia, página, empresa A/B, ID secundario.
3. **Invariantes:** operación, stock antes/después, Kardex, costo, lote, reserva,
   evento contable, auditoría.
4. **Concurrencia:** escenario, trabajadores, replay key, resultado esperado,
   resultado observado, locks/deadlocks.
5. **UI:** página/control, rol, viewport, tema, teclado, consola/red, resultado.
6. **Archivos/impresión:** formato, filas, tenant, checksum, vista previa, papel.
7. **Integraciones:** POS, offline, móvil, compras, MRP/WMS, IA, reportes y público.
8. **Release:** SHA, digests, CI, migraciones, staging, restore, rollback y smoke.

## 8. Comandos previstos, no ejecutados al crear este plan

Terra debe empezar con pruebas enfocadas y usar el runtime Node documentado:

```powershell
Set-Location D:\powerfulcontrolsystem\backend
go test ./db ./handlers -run "Producto|Servicio|Inventario|Bodega|Receta|Permiso|Tenant|Mobile" -count=1
go test ./... -run "^$" -count=1
go vet ./db ./handlers

Set-Location D:\powerfulcontrolsystem
& $node tools\ensure_bootstrap_inventory.mjs --check
& $node tools\runtime_ensure_inventory.mjs --check
& $node tools\migration_audit.mjs --strict
& $node tools\qa_module_contracts.mjs
git diff --check
```

Las carreras de datos se ejecutan contra PostgreSQL real aislado y `-race` en
CI/Linux. El barrido autenticado declara URL, empresa y credenciales en variables
de sesión seguras y conserva resultados por SHA. `rs` no forma parte de estos
comandos hasta la compuerta PSI-011 autorizada.

## 9. Cálculo de avance

Hay 12 fases de igual peso.

- `pendiente`, `bloqueado` o `fallido`: 0 %;
- `parcial`: 50 % solo con evidencia reproducible del mismo SHA;
- `aprobado`: 100 % con todos los criterios cumplidos.

Se informan dos cifras:

1. implementación: aprobadas + parciales ponderadas;
2. certificación: solo aprobadas sobre el mismo SHA/digest.

Estado al crear el plan: **0 % implementación ejecutada por este plan, 0 %
certificación, NO-GO**. Las pruebas históricas sirven para priorizar y evitar
repetición inútil, pero deben vincularse al candidato actual antes de recibir
crédito.

## 10. Orden de inicio para Terra medio

1. PSI-000 y PSI-001.
2. PSI-002 y PSI-003 antes de ampliar funcionalidad o ejecutar datos reales.
3. PSI-004 a PSI-006.
4. PSI-007 y PSI-008.
5. PSI-009, PSI-010 y finalmente PSI-011.

## 11. Ejecución técnica 2026-08-21

Estado de esta ejecución: **parcial, sin autorización de despliegue**. Se
implementaron los P0 que podían verificarse de forma segura en el candidato
local, sin mutar datos de negocio ni ejecutar `rs`.

| Fase | Estado | Evidencia de ejecución |
| --- | --- | --- |
| PSI-000 | aprobado local | Candidato `76a43b971ca53cb6dfe6ccf642faa6e41781c7ef`; se preservaron cambios ajenos del árbol. |
| PSI-001 | parcial | El bootstrap legado permanece catalogado y bloqueado en runtime; no se alteró su manifiesto inmutable. La extracción total a migraciones requiere una migración versionada y ensayo PostgreSQL equivalente. |
| PSI-002 | parcial | Reservas, seriales y lotes validan producto, bodega, lote y serial con `empresa_id`; el handler avanzado ya no expone detalles internos. La carga de imágenes de productos valida firma MIME y deja de aceptar SVG público. |
| PSI-003 | aprobado local | Transferencia, movimiento, cambio de producto y confirmación de reserva descuentan con condición `cantidad >= ?` y verifican filas afectadas. Reservas avanzadas usan transacción, bloqueos `FOR UPDATE`, serial reservado atómico e idempotencia por referencia de origen activa. |
| PSI-004 a PSI-008 | pendiente de validación operativa | No se cambiaron contratos funcionales ya existentes; falta matriz visual, accesibilidad, impresión y regresión por rol en un entorno autenticado. |
| PSI-009 | parcial | Se añadieron pruebas unitarias de no filtración de errores y de la invariante de descuento condicional. Faltan métricas, restore drill y carga PostgreSQL. |
| PSI-010 y PSI-011 | bloqueado por compuerta externa | Requieren staging equivalente, backup/restore probado, pruebas A/B autenticadas y autorización explícita de despliegue/piloto. El estado de producción sigue **NO-GO**. |

Pruebas locales aprobadas durante esta ejecución:

```text
go test ./db ./handlers -count=1
go test ./... -run '^$' -count=1
```

La suite transversal completa `go test ./... -count=1` devolvió código 1 sin
diagnóstico del proceso tras varios minutos. Por tanto no se contabiliza como
aprobada; la compilación global sin casos y las suites completas de `db` y
`handlers` sí finalizaron correctamente.

No saltar a staging, prueba real o despliegue mientras exista una brecha P0
dependiente. Al terminar cada fase, actualizar evidencia, documentación y las
dos cifras sin convertir resultados parciales en certificación.
