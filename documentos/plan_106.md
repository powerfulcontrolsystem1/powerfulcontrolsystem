# Plan 106 - certificación integral para entrada en producción

Estado: **ejecución controlada P0 en curso; `NO-GO` para producción**.

Veredicto inicial: **NO-GO para producción general**.

Fecha de corte: 2026-07-24.

Repositorio auditado: `D:\powerfulcontrolsystem`.

Rama y commit auditados: `main` en `3dd48d24`, alineado con `origin/main` al
inicio de esta planificación.

Empresa obligatoria para las pruebas reales controladas:
`Powerful Control System` (`empresa_id=12`).

Cuenta autorizada para iniciar las pruebas:
`powerfulcontrolsystem@gmail.com`.

Clave: **no se versiona ni se copia en este documento**. El ejecutor debe usar
la clave entregada por el usuario mediante entrada interactiva, variable de
entorno temporal o sesión ya autenticada. Nunca debe imprimirla en consola,
capturas, reportes, commits, comandos, historial ni artefactos de pruebas.

Modelo solicitado para la auditoría y revisión de alto riesgo:
**GPT-5.6 Sol con razonamiento alto**.

Modelo recomendado para ejecutar este plan:
**GPT-5.6 Terra con razonamiento alto**. Terra debe ejecutar una fase acotada
por vez; GPT-5.6 Sol alto debe revisar las decisiones P0 de arquitectura,
seguridad, dinero, migraciones y GO final. Luna solo se recomienda para tareas
repetitivas ya especificadas, nunca para decidir arquitectura, seguridad,
contabilidad, DIAN, concurrencia ni liberación.

## 1. Propósito y alcance

### Planes integrados obligatorios

El cierre de Plan 106 incorpora las compuertas de `plan_107_contador_profesional.md`
y `plan_ia.md`. Plan 107 permanece NO-GO hasta completar fixtures reversibles,
recalculo contable, UAT de contador y conciliaciones en staging. Plan IA
permanece NO-GO hasta demostrar privacidad por usuario/empresa, memoria
consentida, herramientas cerradas, auditoría y pruebas reales por rol. Ninguno
puede declararse terminado solo por pruebas locales.

Pendientes absorbidos de Plan 107:

- ejecutar fixtures reversibles de ciclo contable en staging y comprobar venta,
  caja, inventario, CxC/CxP, impuestos, bancos, nómina, activos y cierres;
- recalcular los datasets, exportaciones y estados financieros con contador
  independiente; completar comparativos, notas, firmas y UAT;
- demostrar aislamiento A/B, idempotencia, concurrencia y reversos contables.

Pendientes absorbidos de Plan IA:

- completar aislamiento de historial, preferencias y memoria por empresa y
  usuario, con consentimiento, retención y borrado verificables;
- probar el modo agente apagado por defecto, herramientas cerradas con permisos
  de dominio, propuestas, confirmación, idempotencia y auditoría unificada;
- ejecutar pruebas reales por rol, adjuntos, streaming, errores, cuotas,
  seguridad y evaluación de calidad/coste/latencia antes de habilitar escrituras.

El Plan 106 sustituye como hoja de ruta principal al Plan 105, conserva todos
sus bloqueadores abiertos y agrega cuatro exigencias:

1. consolidar profesionalmente el nuevo módulo de cartera de proveedores;
2. probar todas las funciones y módulos alcanzables del sistema;
3. probar visualmente todos los botones, incluidos todos los que usan IA;
4. ejecutar pruebas reales controladas con varias cajas simultáneas en la
   empresa `Powerful Control System`.

Este documento es un plan, no una ejecución. No autoriza `rs`, despliegue,
pagos, facturas fiscales, correos, WhatsApp, borrados, cierres contables,
movimientos bancarios, cambios de DNS ni acciones de proveedores. El agente
que redacta el plan debe detenerse al terminarlo.

Actualización de ejecución 2026-07-24: el usuario autorizó implementar el
plan por fases. Se inició solo el bloque P0 de CxP sin ejecutar migraciones,
`rs`, proveedores, pagos, facturación, correos ni mutaciones sobre datos
reales. La ADR y el cambio de código P0 no sustituyen staging ni las pruebas
controladas posteriores.

Avance de ejecución al 2026-07-25: **44%**. Este porcentaje refleja
análisis, diseño, implementación P0/P1 parcial, evidencia automática local y
una validación visual real de solo lectura;
la certificación para producción permanece en **0%** hasta completar staging,
pruebas visuales reales, impresiones, IA, roles A/B, varias cajas simultáneas,
conciliación y la corrección autorizada del incidente DIAN.

La expresión “probar todo” se tratará de forma verificable:

- cada página, ruta, módulo, permiso, job, integración y control interactivo
  debe aparecer en un inventario con propietario y estado;
- cada comportamiento alcanzable en producción debe tener al menos un caso
  positivo, uno negativo y, cuando aplique, concurrencia, idempotencia,
  aislamiento y recuperación;
- ningún botón puede quedar omitido silenciosamente por ser `submit`,
  destructivo, externo o difícil de automatizar;
- un caso no ejecutable debe figurar como `BLOQUEADO` o
  `NO APLICA CON APROBACIÓN`, con causa, riesgo y aprobador;
- aprobar validadores estáticos no equivale a aprobar el flujo visible ni el
  resultado real de negocio.

## 2. Instrucciones obligatorias para GPT-5.6 Terra alto

Antes de editar:

1. Leer completos `AGENTS.md`,
   `documentos/contexto_general_del_sistema.md`,
   `documentos/contexto_especifico_del_sistema.md`,
   `documentos/contexto_codex.md` y este plan.
2. Consultar por fase `documentos/mapa_modulos.md`,
   `documentos/flujos_operativos.md`, `documentos/comandos_codex.md`,
   `documentos/decisiones_tecnicas.md`,
   `documentos/checklist_seguridad_endpoint_multiempresa.md`,
   `documentos/estructura_bd.md`,
   `documentos/matriz_roles_permisos_pos_multiempresa.md` y la documentación
   específica del módulo.
3. Leer Planes 101 a 105 y reutilizar evidencia vigente solo después de
   reconfirmarla contra el SHA candidato.
4. Confirmar `git status --short`, rama, `HEAD`, `origin/main`, diferencias y
   procesos concurrentes antes de cada bloque.
5. Ejecutar una sola tarea P106 acotada por vez. Cada bloque debe cerrar con
   código, pruebas, evidencia, documentación, riesgo residual y rollback.
6. Mantener Go puro, PostgreSQL único, frontend HTML/CSS/JavaScript estático,
   aislamiento por `empresa_id` y cero secretos versionados.
7. No agregar dependencias ni modificar `go.mod` sin autorización expresa.
8. No editar migraciones ya aplicadas ni ejecutar DDL desde API o worker.
9. No usar datos reales de otra empresa para completar pruebas. Las pruebas
   A/B deben usar tenants QA autorizados o una copia anonimizada en staging.
10. No interpretar una contraseña compartida como autorización abierta para
    pagos, facturación, mensajes, borrados o infraestructura.
11. Después de cualquier cambio de SHA, repetir las compuertas afectadas; una
    evidencia anterior no aprueba código nuevo.
12. No ejecutar `rs` hasta P106-020 y solo con autorización expresa posterior.

Condiciones de parada inmediata:

- falta MFA, captcha, firma, llave privada, token o aprobación de un portal;
- una acción puede cobrar, emitir un documento fiscal, transferir dinero,
  enviar comunicaciones, cerrar caja/periodo o borrar datos sin autorización
  específica;
- no se puede demostrar la empresa efectiva, el rol o la licencia;
- hay divergencia entre la base, reportes, contabilidad o saldo de cartera;
- falla una migración, restore, prueba A/B, carrera, idempotencia o gate P0;
- el entorno no ejecuta exactamente el SHA/digest aprobado;
- aparecen cambios ajenos superpuestos en el árbol.

## 3. Línea base comprobada

La revisión del árbol y los validadores locales produjo esta base:

- 1.480 archivos inventariados;
- 566 archivos Go y 197 archivos Go de prueba;
- 697 funciones `Test`, `Benchmark` o `Fuzz`;
- 309 HTML, 105 JavaScript/MJS/CJS, 4 CSS, 23 PowerShell y 32 shell;
- 376 registros `http.HandleFunc` encontrados en el barrido general;
- 59 módulos de permisos backend, 62 wrappers y 136 referencias de menú
  reconocidas por la auditoría profesional;
- 327 rutas en el OpenAPI generado;
- 203 rutas empresariales en la matriz estática, todas con wrapper detectado;
- 154 funciones `Ensure*`, 122 pasos en el catálogo heredado y 104 llamadas
  runtime inventariadas;
- 310 páginas inspeccionadas estáticamente, 223 scripts embebidos y 1.925
  botones detectados por la auditoría UX;
- `go test ./... -count=1`: aprobado;
- `go vet ./...`: aprobado;
- `profesional_preflight.ps1 -Full`: aprobado;
- inventarios de bootstrap, runtime, rutas multiempresa, OpenAPI, migraciones y
  contrato de despliegue: aprobados;
- Docker no está disponible localmente, por lo que Compose no se validó en
  runtime local;
- `go test -race ./...`, carga sostenida, staging integral, todos los botones,
  proveedores reales y regresión visible completa no quedaron demostrados.

Estos resultados son locales y estáticos en parte. No certifican producción.
El Plan 105 continúa con bloqueadores abiertos de candidato inmutable, DDL
heredado, aislamiento A/B, storage compartido, proveedores, CSP/Nextcloud,
capacidad, carrera, app móvil reproducible y piloto.

## 4. Hallazgos P0 del módulo cartera de proveedores

La implementación actual no debe habilitarse como módulo financiero definitivo
sin resolver estos hallazgos:

### 4.1 Dos fuentes de verdad

Existen dos modelos paralelos:

- `empresa_cuentas_por_pagar`, usado por
  `/api/empresa/finanzas/cuentas_pagar`, `finanzas.html`, soportes de compras
  IA y reportes ejecutivos;
- `empresa_contabilidad_cartera_cxp`, usado por
  `/api/empresa/contabilidad_colombia_avanzada?action=cartera_cxp` y
  `contabilidad_colombia_avanzada.html`.

Los reportes de CxP y edades consultan `empresa_cuentas_por_pagar`; por tanto,
una obligación creada en la tabla contable avanzada puede no aparecer en los
reportes principales. El Plan 106 exige una ADR que seleccione una sola fuente
de verdad y retire la escritura paralela.

### 4.2 Dinero, fechas y modelo incompleto

- Los importes están almacenados como `REAL`; debe decidirse y migrarse de
  forma segura a precisión decimal PostgreSQL apropiada para dinero.
- Varias fechas se guardan como `TEXT`; deben normalizarse o validarse con un
  contrato único.
- No hay modelo completo de documento, cuota, asignación, anticipo, nota,
  ajuste, disputa, reverso y evidencia.
- `proveedor_id` puede quedar en cero y el nombre libre no demuestra que el
  proveedor pertenezca a la misma empresa.
- Faltan restricciones únicas e invariantes suficientes para impedir
  duplicados por empresa, proveedor, documento, origen e idempotencia.

### 4.3 Abonos y conciliación no atómicos

- El abono de contabilidad avanzada hace lectura y actualización separadas,
  sin transacción ni `SELECT ... FOR UPDATE`; dos abonos simultáneos pueden
  perder actualizaciones.
- No existe clave de idempotencia del abono y un reintento puede duplicar el
  efecto.
- Un pago superior al saldo se reduce silenciosamente; debe rechazarse o
  requerir una regla explícita y auditable.
- El flujo de Finanzas crea primero un movimiento y después actualiza la CxP
  sin una transacción común. Un fallo intermedio puede dejar movimiento sin
  aplicación o saldo desalineado.
- La conciliación relaciona pagos por documento o nombre del tercero. Esa
  heurística puede asociar movimientos incorrectos y no reemplaza una tabla
  explícita de asignaciones.
- Los eventos contables “no bloqueantes” pueden fallar después del cambio de
  saldo; la consistencia contable debe ser atómica o usar outbox durable.

### 4.4 Integraciones que también pueden duplicar

La contabilización de un soporte de compra IA inserta una CxP y luego actualiza
el soporte en operaciones separadas. Debe ser transaccional e idempotente.
También deben revisarse compras, devoluciones a proveedor, egresos, anticipos,
documentos soporte, tesorería, bancos, impuestos, contabilidad y reportes para
evitar que cada uno cree una deuda distinta.

### 4.5 Permisos y pruebas insuficientes

- Los wrappers generales de Finanzas y Contabilidad avanzada son una base, pero
  no separan registrar, aprobar, ajustar, proponer pago, pagar, conciliar,
  reversar y exportar.
- No se encontraron pruebas Go específicas para los dos flujos CxP, el abono
  concurrente, la conciliación ni la contabilización IA a CxP.
- La interfaz actual no cubre profesionalmente cuotas, saldos iniciales,
  anticipos, notas, reversos, estados de cuenta, propuestas de pago,
  aprobaciones y trazabilidad completa.

## 5. Resultado obligatorio del módulo cartera de proveedores

Antes de producción debe existir un único módulo empresarial integrado, no una
tercera implementación. Debe ofrecer:

- obligaciones por proveedor y documento;
- compras a crédito y saldos iniciales;
- cuotas y fechas de vencimiento;
- anticipos y saldos a favor;
- abonos y asignaciones parciales;
- notas crédito/débito, descuentos, ajustes y reversos;
- estados `borrador`, `aprobado`, `parcial`, `pagado`, `vencido`,
  `en_disputa` y `anulado`;
- propuesta y aprobación de pagos, sin confundir registro con transferencia
  bancaria;
- conciliación por asignaciones explícitas;
- edad de cartera, vencido, por vencer, flujo esperado y estado de cuenta;
- moneda, tasa y redondeo definidos;
- adjuntos/evidencia privada por empresa;
- vínculo con compra, soporte IA, devolución, egreso, asiento y auditoría;
- exportación coherente y aislada;
- historia inmutable de movimientos;
- idempotencia y concurrencia demostradas.

## 6. Fases de ejecución

### P106-001 - Congelar candidato y alcance [P0]

1. Actualizar `main` de forma segura y seleccionar un SHA único.
2. Crear manifiesto con commit, árbol limpio, migraciones, imágenes y digests.
3. Aprobar qué módulos entran al piloto y cuáles quedan apagados realmente en
   UI, API, worker y proveedores.
4. Definir responsables, ventana, RPO, RTO, SLO, presupuesto IA y presupuesto
   de proveedores.
5. Crear matriz de evidencia con estado, SHA, entorno, empresa, rol, dato QA,
   resultado, captura/log saneado, riesgo y aprobador.

Aceptación: existe una única fuente inmutable de liberación y el usuario aprobó
el alcance. Cualquier cambio posterior invalida las evidencias afectadas.

### P106-002 - Inventario total de módulos, funciones y botones [P0]

1. Cruzar `mapa_modulos`, `descripcion_de_modulos`, permisos, licencias,
   menús, HTML, JavaScript, rutas, OpenAPI, tablas, jobs e integraciones.
2. Generar un manifiesto versionado de cada página y control interactivo con
   selector estable, acción, endpoint, permiso, licencia y nivel de riesgo.
3. Tomar los 1.925 botones como línea base inicial, incluir botones dinámicos,
   enlaces con apariencia de botón, submits, controles por rol y elementos
   creados por JavaScript.
4. Clasificar cada control: navegación, lectura, mutación reversible, dinero,
   fiscal, comunicación, destructivo, proveedor, hardware o IA.
5. Inventariar funciones de negocio y handlers alcanzables. Cada una debe
   relacionarse con pruebas positivas/negativas; las utilidades generadas o
   fuera de runtime deben justificarse.
6. Modificar `qa_e2e_buttons.cjs` para que ningún control `unsafe` desaparezca
   de cobertura: debe quedar probado con fixture o marcado bloqueado con
   aprobación.

Aceptación: cobertura de inventario 100%, cero páginas/rutas/botones huérfanos y
cero omisiones silenciosas.

### P106-003 - ADR y reconciliación inicial de CxP [P0]

1. Medir por empresa conteos, sumas y relaciones de
   `empresa_cuentas_por_pagar` y `empresa_contabilidad_cartera_cxp`.
2. Rastrear todos los productores y consumidores: Finanzas, Contabilidad
   avanzada, compras, soportes IA, devoluciones, egresos, eventos y reportes.
3. Redactar ADR seleccionando la fuente canónica. La opción preferida debe
   reutilizar `empresa_cuentas_por_pagar` si la evidencia confirma que concentra
   integraciones y reportes; la ADR puede decidir otra opción solo con
   justificación y plan de compatibilidad.
4. Detectar duplicados y diferencias sin fusionar ni borrar automáticamente.
5. Definir puente temporal de solo lectura/escritura única y fecha de retiro de
   la tabla no canónica.

Aceptación: una sola fuente de verdad aprobada, totales reconciliados y plan de
migración reversible sin doble contabilización.

### P106-004 - Modelo, migraciones e invariantes de cartera [P0]

1. Crear migraciones inmutables desde `pcs-migrate`; nunca DDL en handlers.
2. Usar tipos PostgreSQL adecuados para dinero, fecha, moneda y tasa.
3. Crear documentos, cuotas, movimientos/asignaciones, anticipos, ajustes,
   evidencias y aprobaciones solo donde la ADR demuestre necesidad.
4. Añadir claves foráneas o validaciones equivalentes por `empresa_id`,
   restricciones de estado, saldos y unicidad, e índices por empresa,
   proveedor, vencimiento, estado y documento.
5. Preparar backfill con conteos y sumas antes/después, ensayo desde cero,
   upgrade representativo y rollback.
6. Retirar gradualmente la tabla paralela sin borrar historia.

Aceptación: migración reproducible, checksum correcto, ningún cambio de saldo,
ningún DDL en API/worker y rollback demostrado.

### P106-005 - Motor transaccional e idempotente de movimientos [P0]

1. Implementar un servicio Go único para crear, aprobar, abonar, asignar,
   ajustar, anular y reversar.
2. Ejecutar cada operación en transacción PostgreSQL con bloqueo acotado de
   obligación/asignación.
3. Exigir clave de idempotencia única por empresa, operación y origen.
4. Rechazar sobreasignación, monto inválido, moneda/tasa incompatible,
   documento cerrado, proveedor de otra empresa y periodo cerrado.
5. Escribir saldo, movimiento, asignación, auditoría, evento contable y outbox
   de forma atómica.
6. Usar reversos compensatorios; nunca editar o borrar pagos cerrados.
7. Asegurar que registrar/proponer un pago no ordena dinero.

Aceptación: dos o más abonos concurrentes no pierden actualizaciones, no
duplican movimientos y nunca dejan saldo negativo o contabilidad huérfana.

### P106-006 - Integración de compras, IA, devoluciones y contabilidad [P0]

1. Hacer transaccional e idempotente la conversión de soporte IA a CxP.
2. Conectar compras a crédito, documentos soporte, devoluciones, anticipos,
   egresos, tesorería, bancos y contabilidad con el servicio canónico.
3. Evitar asientos dobles entre evento, movimiento financiero y CxP.
4. Validar impuestos, retenciones, centros de costo y periodo contable.
5. Actualizar reportes y libros para leer exclusivamente la fuente canónica.
6. Implementar importación de saldos iniciales con vista previa, validación de
   totales, idempotencia, actor y rollback.

Aceptación: una operación de origen crea exactamente una obligación/asignación
y todos los módulos muestran el mismo saldo.

### P106-007 - Seguridad, roles y aislamiento de cartera [P0]

1. Definir permisos independientes para registrar, leer, aprobar, ajustar,
   proponer pago, registrar pago, conciliar, reversar, importar y exportar.
2. Resolver siempre empresa y usuario desde el contexto backend.
3. Validar que proveedor, documento, cuota, movimiento, archivo y cuenta
   pertenecen a la misma empresa.
4. Aplicar CSRF, límites, auditoría, errores públicos saneados y licencias.
5. Ejecutar A/B con IDs iguales o parecidos por query, path, body, header,
   cache, exportación, archivo, job y búsqueda.
6. Verificar que ocultar botones no permite invocar endpoints sin permiso.

Aceptación: cero fuga A/B y cada acción financiera rechaza rol, licencia,
empresa o estado no autorizados.

### P106-008 - UX profesional de cartera de proveedores [P1]

Crear una entrada única en Administrar empresa con:

- tablero de deuda total, vencida, por vencer y saldo a favor;
- filtros por proveedor, documento, estado, vencimiento, centro, moneda y
  origen;
- detalle por documento, cuota y movimiento;
- estado de cuenta y edad de cartera;
- anticipos, notas, ajustes y reversos;
- propuesta/aprobación/registro de pago claramente diferenciados;
- conciliación y diferencias visibles;
- importación con vista previa;
- adjuntos, actor, fecha y auditoría;
- estados vacíos, carga, error, éxito y bloqueo de periodo;
- ayuda contextual y textos que no prometan transferencia bancaria.

Validar escritorio, móvil, teclado, zoom, contraste, temas claro/oscuro,
lectores, impresión y ausencia de errores de consola/red.

### P106-009 - Suite específica de cartera [P0]

Incluir como mínimo:

- creación manual y desde compra/soporte IA;
- documento duplicado, proveedor inexistente y proveedor de otra empresa;
- cuota única y múltiples cuotas;
- saldo inicial, anticipo, saldo a favor, nota, ajuste y reverso;
- abono parcial/total y sobrepago;
- dos, cuatro y diez abonos concurrentes;
- reintento, doble clic, timeout y replay;
- periodo abierto/cerrado;
- COP sin centavos operativos y monedas con decimales;
- fechas límite, vencimientos y zonas horarias;
- importación repetida y rollback;
- caída después de cada paso transaccional;
- contabilidad, reportes, exportaciones y edades reconciliados;
- permisos por rol y A/B multiempresa;
- restore con saldos idénticos.

Objetivo: cobertura alta de las reglas críticas y cero prueba omitida en el
manifiesto. La cobertura de líneas no reemplaza invariantes de dinero.

### P106-010 - Cierre de deuda técnica heredada del Plan 105 [P0/P1]

1. Retirar todas las llamadas `Ensure*` alcanzables desde tráfico HTTP y
   completar migraciones versionadas.
2. Demostrar API y worker sin DDL.
3. Completar clasificación y corrección de goroutines, outbox y reintentos.
4. Terminar errores públicos saneados.
5. Resolver CSP sin `unsafe-inline` y la integración soportada de Nextcloud.
6. Demostrar almacenamiento privado compartido para réplicas.
7. Fijar imágenes/digests, SBOM, escaneo y mínimo privilegio.
8. Decidir app móvil reproducible o retirarla del alcance.
9. Mantener voz Python/Piper fuera del lanzamiento salvo excepción aprobada.
10. Resolver datos QA abiertos con acciones de negocio autorizadas.

Aceptación: todos los P0/P1 heredados están aprobados o explícitamente fuera de
alcance con control efectivo y aprobación.

### P106-011 - Seguridad integral [P0]

Probar autenticación, sesión, MFA de super administrador, CSRF, XSS, SQLi,
SSRF, path traversal, uploads, CORS, CSP, cookies, rate limit, brute force,
IDOR, escalamiento de rol, secretos, logs, backups, webhooks, firma, replay,
dependencias, contenedores, firewall, puertos y configuración.

Ejecutar análisis estático, pruebas dinámicas en staging, escaneo de imágenes y
host, y revisión manual de flujos financieros/fiscales. Cero vulnerabilidades
críticas o altas abiertas; las medias requieren dueño, fecha y aceptación.

### P106-012 - Regresión de todos los módulos y funciones [P0]

La matriz debe cubrir, como mínimo:

| Dominio | Flujos obligatorios |
|---|---|
| Acceso y gobierno | registro, login, recuperación, contrato, logout, sesiones, selección de empresa, super administrador, usuarios, roles, permisos y licencias |
| Empresa y configuración | creación/edición, preconfiguración, sucursales, estaciones, cajas, impresoras, preferencias, temas, país, moneda y auditoría |
| POS y ventas | productos, servicios, clientes, scanner, carrito, venta directa, estaciones, reservas, descuentos, impuestos, propinas, comisiones, pagos mixtos, crédito, devolución, factura y recibo |
| Caja | apertura, turnos, abonos, ingresos/egresos, cierres, reportes, reapertura autorizada y varias cajas simultáneas |
| Inventario y compras | catálogo, bodegas, existencias, Kardex, lotes/series, proveedores, compras, reposición, devoluciones, soportes IA y cartera CxP |
| Finanzas y contabilidad | movimientos, bancos, Bre-B, conciliación, cuentas, CxC, CxP, centros de costo, tesorería, presupuestos, NIIF, activos, impuestos, exógena, libros, cierre fiscal y portal contador |
| Facturación electrónica | Colombia/DIAN, Ecuador/SRI y Panamá/DGI según alcance; factura, notas, documento soporte, nómina y RADIAN |
| Pagos y licencias | Wompi, ePayco, checkout, webhooks, renovaciones, comprobantes, replays, rechazos y conciliación |
| Personas | empleados, nómina, asistencia, horarios, vacaciones, carnets, hoja de vida y seguridad social |
| CRM y canales | CRM, cotizaciones, pedidos, campañas, venta pública, domicilios, Rappi, WhatsApp, correo y red social |
| Documentos y colaboración | archivos, uploads, exportes, backups, Nextcloud, OnlyOffice, PDF, Word, Excel, impresión y correo |
| IA | chat, agentes, Centro IA, Renta IA, soportes, compras, ingresos/egresos, productos, DIAN, grafología, música y voz |
| Verticales | hotel, motel, apartamentos, restaurante, droguería, consultorio, gimnasio, taxi/GPS, cámaras, domótica, energía solar, control eléctrico, producción MRP y demás plantillas habilitadas |
| Operación | API, worker, migrador, jobs, outbox, cron, health/ready, métricas, alertas, soporte, VPS, Docker, backup/restore y rollback |
| Público y móvil | portal, noticias, catálogo, venta pública, privacidad, API v1 y app móvil solo si entra al alcance |

Cada fila debe ejecutarse con roles permitidos y restringidos, estados vacío,
carga, éxito y error. No se acepta una fila aprobada solo porque abre la página.

### P106-013 - Pruebas visuales y de todos los botones [P0]

1. Ejecutar todas las páginas en 1440x900, 1366x768, 1024x768, 768x1024,
   390x844 y 360x800.
2. Probar claro/oscuro, zoom 100/125/200%, teclado, foco, contraste y lector.
3. Capturar consola, errores de página, requests fallidos y respuestas >=400.
4. Probar los 1.925 botones de línea base y cualquier control nuevo.
5. Para mutaciones, usar fixtures QA y verificar estado antes/después.
6. Para acciones destructivas/externas, usar staging/sandbox o aprobación
   puntual; nunca marcarlas aprobadas por omitir el clic.
7. Probar modales, pestañas, filtros, paginación, menús, copiar, descargar,
   subir, imprimir, fullscreen, cámara, micrófono y hardware cuando aplique.
8. Validar papel real de recibos, facturas, reportes y comprobantes.
9. Guardar matriz `PASS/FAIL/BLOCKED/NA`, selector, captura y evidencia.

Aceptación: 100% de controles inventariados tiene resultado explícito; cero
error de consola/red P0, texto roto, acción sin feedback o botón sin permiso.

#### Subfase obligatoria P106-013A - Impresiones y vistas previas de todo el sistema [P0]

No basta con que el botón `Imprimir`, `PDF`, `Vista previa`, `Descargar` o
`Compartir` abra una ventana. Se debe inventariar y revisar toda salida visual
de negocio, incluyendo facturas, facturas electrónicas, notas, documentos
soporte, recibos de venta/pago/caja, comprobantes contables, egresos, ingresos,
cierres, arqueos, pedidos, cotizaciones, órdenes de compra, remisiones,
devoluciones, etiquetas, reportes, nómina, certificados, exportaciones PDF y
cualquier documento habilitado por un módulo o vertical.

Para cada tipo de documento y plantilla activa, ejecutar y evidenciar:

1. Vista previa en pantalla: filas y columnas alineadas, encabezados, totales,
   impuestos, descuentos, fechas, moneda, observaciones, QR/códigos, firmas y
   datos de empresa legibles; sin texto superpuesto, cortado, desbordado ni
   valores fuera de su columna.
2. Datos límite: razón social/dirección/producto/proveedor largos, muchos ítems,
   múltiples impuestos/descuentos/pagos, notas extensas, caracteres UTF-8 y
   documentos de varias páginas. Confirmar repetición de encabezado, salto de
   página, numeración y total final correctos.
3. Formatos: A4/Carta y térmico configurado (58 mm y 80 mm cuando aplique),
   orientación, márgenes, escala, corte, apertura de cajón solo si está
   autorizada y ausencia de hojas en blanco adicionales.
4. Presentación: claro/oscuro no altera el documento impreso, contraste y foco
   accesibles, zoom 100/125/200%, escritorio y móvil; PDF descargado coincide
   con la vista previa y la impresión física.
5. Consistencia: importes, consecutivo, cliente/proveedor, `empresa_id`, actor,
   fecha, estado y QR/código coinciden con la transacción origen. Una impresión
   no crea, duplica, modifica ni reenvía documentos.
6. Roles y errores: rol sin permiso no puede generar ni acceder a un documento
   ajeno; sin impresora, popup bloqueado, plantilla ausente o fallo de PDF
   muestran una explicación accionable y no dejan la operación en estado
   ambiguo.

La matriz debe registrar URL/selector, tipo de documento, empresa, rol,
plantilla, tamaño de papel, navegador/impresora, caso de datos, resultado,
captura de vista previa y PDF o foto de impresión saneada. Los documentos
fiscales o pagos reales se prueban primero en staging/datos QA y solo pasan a
empresa 12 con autorización puntual.

Aceptación P106-013A: 100% de salidas documentales inventariadas tiene resultado
explícito y evidencia visual; cero defectos P0 de legibilidad, columnas/filas,
totales, salto de página, aislamiento de empresa o efectos secundarios.

### P106-014 - Pruebas de todos los botones y funciones IA [P0]

El inventario IA debe incluir botones estáticos, dinámicos y lanzadores:

- chat de super, selector, empresa y portal público: abrir, minimizar, cerrar,
  ejemplos, configuración, nuevo chat, conversación, micrófono, voz, detener,
  adjuntar, quitar adjunto y enviar;
- modelos/agentes, consulta normal, streaming, cancelación, timeout y reintento;
- herramientas de agente en lectura, propuesta, confirmación y rechazo;
- Centro IA: diagnóstico ERP y todas sus funciones dinámicas;
- Renta IA;
- productos: cargar carta/precios con IA;
- compras, ingresos y egresos: analizar factura/comprobante;
- soportes de compras: radicar, extraer IA, aprobar/rechazar y contabilizar;
- DIAN: analizar PDF de numeración y aplicar valores;
- grafología GPT;
- búsqueda de música IA por estación;
- configuración IA global, conexión OpenAI, proveedor propio por empresa,
  límites, consumo, contexto y voz.

Casos obligatorios:

- IA activada, desactivada, sin clave, clave inválida, cuota agotada y límite de
  costo;
- respuesta válida, error del proveedor, 429, 5xx, desconexión, streaming
  interrumpido y botón `Stop`;
- adjuntos válidos/inválidos, tamaño, MIME, malware simulado y privacidad;
- prompt injection, intento de extraer secretos, SQL, otra empresa o cambiar
  `empresa_id`;
- rol de lectura intentando escribir;
- propuesta vencida, alterada, confirmada dos veces o repetida por red;
- salida con PII/secretos saneada;
- consumo, modelo, costo, actor, empresa y resultado auditados;
- respuesta visible, accesible y coherente en PC/móvil.

Toda escritura de IA exige propuesta temporal y confirmación independiente.
Ningún modelo puede recibir SQL libre, HTTP arbitrario ni seleccionar tenant.
Probar IA real requiere presupuesto autorizado; de lo contrario queda
`BLOQUEADO`, nunca simulado como proveedor aprobado.

### P106-015 - Varias cajas simultáneas reales [P0]

En `Powerful Control System` preparar cuatro usuarios/cajas QA independientes:
`Caja QA-01` a `Caja QA-04`. Usar productos y clientes marcados `P106-QA`,
stock controlado y datos reversibles.

Ejecutar en paralelo:

1. login y apertura de cuatro turnos;
2. cuatro carritos con productos diferentes;
3. dos cajas compitiendo por el último stock del mismo producto;
4. cantidades, descuento, impuesto, propina y comisión;
5. pago efectivo, mixto, crédito y método no monetario/sandbox según alcance;
6. doble clic, retry de red, offline/sync y cierre concurrente;
7. transferencia de estación y visibilidad entre cajas;
8. intento de operar dos sesiones sobre la misma caja;
9. abono concurrente y pago de cartera en paralelo;
10. cierre y reporte simultáneo.

Reconciliar por caja y consolidado:

- ventas y documentos únicos;
- stock inicial - vendido + reversos = stock final;
- pagos por método = movimientos y caja;
- propinas, comisiones, impuestos, descuentos y devoluciones;
- crédito y cartera creada;
- aperturas, cierres, sobrante/faltante y auditoría;
- ninguna caja ve o modifica datos no permitidos de otra;
- cero venta, pago, factura o movimiento duplicado.

Las ventas fiscales, pagos reales y cierres irreversibles requieren autorización
puntual. Primero se demuestra el escenario completo en staging y luego el
subconjunto aprobado en empresa 12.

### P106-016 - Proveedores e integraciones reales [P0]

Probar solo integraciones incluidas y autorizadas:

- DIAN hasta acuse final aceptado `StatusCode=00`, no solo TrackId;
- Wompi/ePayco: sandbox y operación real mínima autorizada, webhook, replay,
  conciliación, comisión, reverso/rechazo;
- Bre-B/QR y bancos: conciliación sin ordenar dinero accidentalmente;
- Mailu/correo: envío, recepción, SPF, DKIM, DMARC, rebote y adjunto;
- WhatsApp y Rappi: firma, webhook, replay, consentimiento y aislamiento;
- Nextcloud/OnlyOffice: login, iframe soportado, archivo A/B, cuota, edición,
  restore y CSP;
- almacenamiento, mapas/GPS, cámaras/hardware y otros proveedores incluidos.

Cada prueba debe registrar costo, dato creado, evidencia externa, conciliación
y limpieza/reverso. Un HTTP 200 no aprueba la integración.

Bloqueador observado el 2026-07-24 en la empresa 12: una factura electrónica
falló antes de contactar DIAN porque el proceso backend no pudo leer la clave
privada de firma configurada. Antes de reintentar o emitir otro documento se
debe, en staging y con aprobación operativa, verificar propietario/permisos de
la ruta privada frente al usuario efectivo del contenedor y migrar o recargar la
firma únicamente mediante el flujo seguro de PCS. No copiar la clave a consola,
chat, repositorio ni almacenamiento público. La corrección se acepta solo con
diagnóstico DIAN seguro y una emisión de habilitación autorizada que llegue a
`GetStatusZip StatusCode=00`.

### P106-017 - Rendimiento, continuidad y observabilidad [P0]

1. Ejecutar `go test -race ./...` en Linux compatible.
2. Carga sostenida y picos para login, catálogo, carrito, pago, caja, cartera,
   reportes, uploads, IA y jobs.
3. Medir P50/P95/P99, errores, pool DB, locks, CPU, RAM, disco, red y colas.
4. Probar caída/reinicio de API, worker, PostgreSQL, Redis, storage y proveedor.
5. Ejecutar backup completo autorizado y restore aislado actual.
6. Medir RPO/RTO y restaurar archivos, permisos, auditoría y saldos de cartera.
7. Ensayar rollback de imagen y migración compatible.
8. Probar alertas reales, guardias, escalamiento y runbooks.

Aceptación: objetivos aprobados, cero pérdida/duplicación, restore demostrado y
alertas accionables.

### P106-018 - Staging equivalente y ensayo general [P0]

1. Publicar el SHA/digest candidato en staging equivalente.
2. Restaurar copia anonimizada representativa.
3. Aplicar migraciones desde versión actual y desde cero.
4. Ejecutar P106-009 a P106-017 completos.
5. Corregir y repetir hasta tener cero P0/P1.
6. Hacer ensayo de despliegue, smoke, rollback y recuperación.
7. Congelar el candidato final y repetir CI/preflight.

Aceptación: matriz completa en verde sobre el mismo artefacto que se propone
para producción.

### P106-019 - Pruebas reales controladas en empresa 12 [P0]

Con autorización expresa posterior:

1. autenticar la cuenta indicada sin registrar la clave;
2. confirmar visual y técnicamente `Powerful Control System`, `empresa_id=12`;
3. crear datos `P106-QA` con responsable y caducidad;
4. ejecutar smoke de todos los módulos habilitados, pruebas IA autorizadas y
   cuatro cajas simultáneas;
5. verificar DB/operación mediante endpoints y reportes, no SQL destructivo;
6. reconciliar dinero, stock, caja, cartera, documentos, archivos y auditoría;
7. limpiar únicamente con acciones de negocio y verificar antes/después;
8. conservar evidencia saneada sin datos personales ni secretos.

Un módulo fuera del alcance aprobado debe permanecer bloqueado en UI, API,
worker e integración. No se activa para “poder probarlo” sin autorización.

### P106-020 - Release, piloto y decisión GO/NO-GO [P0]

1. Confirmar SHA, tag, digests, árbol limpio, CI, race, seguridad, SBOM y
   firmas.
2. Confirmar migraciones, backup, restore, rollback, capacidad y alertas.
3. Confirmar matrices de módulos, botones, IA, cajas, cartera y proveedores.
4. Obtener firmas del responsable técnico, QA/operación y usuario.
5. Solicitar autorización expresa para `rs`.
6. Ejecutar `rs` solo después de la autorización.
7. Validar postdespliegue salud, readiness, SHA/digests, logs, migración,
   worker, smoke visual y conciliación.
8. Abrir piloto limitado, observar métricas y cerrar o revertir según SLO.
9. Ampliar producción únicamente tras la ventana aprobada.

## 7. Comandos mínimos de verificación

Terra debe consultar primero `documentos/comandos_codex.md`.

```powershell
Set-Location D:\powerfulcontrolsystem
git status --short
git rev-parse HEAD
git rev-parse origin/main

$node = 'C:\Users\ivanm\.cache\codex-runtimes\codex-primary-runtime\dependencies\node\bin\node.exe'
& $node tools/ensure_bootstrap_inventory.mjs --check
& $node tools/runtime_ensure_inventory.mjs --check
& $node tools/tenant_route_inventory.mjs --check
& $node tools/openapi_inventory.mjs --check
& $node tools/migration_audit.mjs --strict
& $node tools/deploy_pipeline_contract.mjs
& $node tools/plan106_ui_inventory.mjs

pwsh.exe -NoProfile -File .\scripts\profesional_preflight.ps1 -Full

Set-Location D:\powerfulcontrolsystem\backend
go test ./... -count=1
go vet ./...
```

En CI/builder Linux:

```sh
go test -race ./...
```

Los comandos de navegador, staging, backup, restore, carga, release y `rs`
deben tomarse de la documentación vigente y ejecutarse únicamente en su fase.

## 8. Matriz de evidencia que debe mantener Terra

| ID | Estado | SHA/digest | Entorno | Empresa/rol | Prueba | Evidencia | Riesgo | Aprobador |
|---|---|---|---|---|---|---|---|---|
| P106-001 | En curso | árbol local P106 | decisión | - | candidato/alcance | cambios P0 sin desplegar | artefacto variable | usuario |
| P106-002 | Parcial | árbol local P106 | local | - | inventario total | manifiesto estático reproducible: 309 HTML, 6.189 controles, 2.975 acciones y 881 marcadores dinámicos; runner E2E listo y validado con el Playwright incluido por Codex, sin instalar dependencias | falta enlazar cada control dinámico con selector/endpoint/permiso y ejecutar evidencia funcional controlada | técnico |
| P106-003 | Parcial | árbol local P106 | análisis/DB | sin datos reales | ADR CxP | `adr_106_cxp_fuente_canonica.md` | históricos sin conciliar | técnico/contador |
| P106-004 | Parcial | árbol local P106 | PostgreSQL efímero/staging | A/B | migración CxP | catálogo `20260724-001` sin ejecutar | datos históricos | técnico |
| P106-005 | Parcial | árbol local P106 | tests/staging | A/B | concurrencia/idempotencia | transacción y hash implementados; staging pendiente | dinero | técnico/QA |
| P106-006 | Parcial | árbol local P106 | tests/staging | sin datos reales | soporte IA a CxP | conversión transaccional, proveedor tenant y outbox implementados | compras/devoluciones/reportes aún pendientes | contador/QA |
| P106-007 | Parcial | árbol local P106 | revisión estática/tests dirigidos | roles A/B | permisos/IDOR CxP | ruta bajo `WithEmpresaFinanzasPermissions`; consultas CxP y catálogo de proveedor activo filtran `empresa_id`; alta exige proveedor canónico; A/B real pendiente | fuga tenant | seguridad |
| P106-008 | Bloqueado despliegue | árbol local P106 | producción y staging, navegador real | Powerful Control System, administrador | UX CxP | autenticación real aprobada en ambos entornos; Finanzas publicada y staging no contienen aún CxP ni cargue IA presentes en el árbol local | cambio no desplegado; no puede certificarse el flujo ni sus impresiones | usuario/release |
| P106-009 | Parcial | árbol local P106 | tests locales/CI pendiente | A/B | suite CxP | contratos de hash, bloqueo, ledger, outbox y sobrepago; concurrencia PostgreSQL pendiente | saldos | QA |
| P106-010 | Pendiente | - | local/CI/staging | - | deuda P105 | - | plataforma | técnico |
| P106-011 | Pendiente | - | staging/VPS | roles | seguridad | - | incidente | seguridad |
| P106-012 | Parcial | producción y árbol local | Powerful Control System, administrador | empresa 12 | smoke módulos/funciones | Panel, venta directa, corte de caja, Finanzas y facturación electrónica abrieron sin login; smoke detectó un error `MutationObserver` en Facturación y la guarda está corregida localmente | cambio pendiente de despliegue; faltan todos los módulos, roles, botones y regresión completa | QA |
| P106-013 | Parcial | local y producción, navegador real | Powerful Control System, administrador | empresa 12 | botones y P106-013A: impresiones/vistas previas | batería local portable: 18 formatos Carta/POS (facturas, recibos, comprobantes, órdenes, cortes y tickets) sin desborde ni error; tres flujos autoimpresión invocaron `window.print()`; Finanzas/CxP escritorio/móvil contenido | faltan cada salida real, botones, vista previa y papel físico | QA |
| P106-014 | Parcial | producción, navegador real | Powerful Control System, administrador | empresa 12 | IA completa | El diálogo IA abrió; modo `Ayudante por pasos` y agente `Compras` se seleccionaron sin enviar consulta, archivo ni acción | faltan minimizar/cerrar en navegador compatible, consulta, adjuntos, streaming, errores, cuotas, A/B y cada botón IA | usuario/QA |
| P106-015 | Parcial | producción, navegador real | Powerful Control System, administrador | empresa 12 | 4 cajas | Cuatro aperturas `P106-QA-01..04` coexistieron con saldo cero, se verificaron desde una segunda sesión y se eliminaron; un ciclo adicional confirmó apertura y limpieza. El cierre pide caja física en `window.prompt`, diálogo no expuesto por este navegador de automatización | faltan ventas concurrentes, cierre manual compatible, arqueo, conciliación, stock y reintentos | usuario/QA |
| P106-016 | Bloqueado | incidente DIAN 2026-07-24 | staging/producción controlada | empresa 12 | E2E externo y acceso seguro a firma | lectura de clave privada fallida; sin reintento. Verificación de solo lectura 2026-07-25 confirma que la alerta histórica publicada aún expone detalle técnico; el árbol local la sanea al leerla | falta despliegue, reparación autorizada de permisos/firma y reintento controlado | fiscal/seguridad | usuario |
| P106-017 | Parcial | VPS/staging autorizado | staging | - | backup/restore/carga | backup completo, drill temporal y smoke staging: 30 solicitudes, concurrencia 5, P95 427 ms, 0% errores | faltan `-race` CI y carga sostenida autenticada | operaciones |
| P106-018 | Pendiente | - | staging | A/B | ensayo general | - | equivalencia | técnico/QA |
| P106-019 | Pendiente | - | producción controlada | empresa 12 | smoke real | - | datos reales | usuario |
| P106-020 | Pendiente | - | release/piloto | empresa 12 | GO/NO-GO | - | producción | usuario |

Estados permitidos: `Pendiente`, `En curso`, `Bloqueado`, `Aprobado`,
`No aplica con aprobación`. Nunca usar `Aprobado` sin evidencia ejecutada sobre
el SHA y entorno declarados.

## 9. Compuerta final GO/NO-GO

GO solo si:

- [ ] SHA, tag, imágenes y digests son únicos, trazables y aprobados.
- [ ] CI, Go, race, seguridad, migraciones y preflight están verdes.
- [ ] API/worker no ejecutan DDL y el migrador es reproducible.
- [ ] Todos los P0/P1 están aprobados o fuera de alcance de forma efectiva.
- [ ] Una sola fuente CxP está reconciliada y todos los movimientos son
      transaccionales e idempotentes.
- [ ] Aislamiento A/B está demostrado en datos, archivos, cache, jobs, IA,
      reportes y exportaciones.
- [ ] El manifiesto de módulos, funciones y botones tiene 100% de estados.
- [ ] Todas las impresiones y vistas previas documentales tienen evidencia de
      legibilidad, filas/columnas, totales, paginación y aislamiento empresarial.
- [ ] Todos los botones IA incluidos tienen evidencia funcional y de seguridad.
- [ ] Cuatro cajas simultáneas cierran y reconcilian sin duplicados.
- [ ] Staging equivalente completó el ensayo general.
- [ ] Proveedores incluidos tienen evidencia real reconciliada.
- [ ] Backup/restore, RPO/RTO, rollback, carga, SLO y alertas están demostrados.
- [ ] Producción controlada en empresa 12 no deja datos o saldos incoherentes.
- [ ] Documentación, runbooks, soporte y responsables están actualizados.
- [ ] El usuario autoriza explícitamente `rs`, piloto y apertura comercial.

NO-GO automático ante:

- fuga o autorización cruzada entre empresas;
- doble fuente de cartera, diferencia de saldo o contabilidad;
- pago, venta, factura, caja, stock, abono o job duplicable;
- migración no reproducible, DDL runtime o checksum divergente;
- botón/módulo P0 sin probar o error visual/de consola P0;
- IA capaz de filtrar secretos/otro tenant o escribir sin confirmación;
- proveedor sin prueba real;
- backup sin restore o rollback inviable;
- vulnerabilidad crítica/alta;
- SHA/digest diferente del aprobado;
- pruebas de varias cajas sin reconciliación;
- evidencia estática presentada como E2E real.

## 10. Cierre de esta planificación

La base local compila y supera sus validadores, pero sigue en **NO-GO**. El
módulo de cartera de proveedores presenta dos fuentes de verdad y riesgos de
concurrencia/idempotencia que son P0. La entrada en producción solo puede
considerarse después de ejecutar y aprobar todas las fases del Plan 106.

Al terminar de crear este documento, Codex debe detenerse. El siguiente turno
debe comenzar únicamente cuando el usuario cambie al modelo deseado y autorice
expresamente la ejecución del Plan 106.
