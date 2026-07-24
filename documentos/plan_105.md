# Plan 105 - cierre verificable para produccion general

Estado del plan: **ejecucion y despliegue `rs` autorizados expresamente por el
usuario para esta corrida; pagos, facturas fiscales y borrados siguen fuera de
la prueba automatica**.

Veredicto actual: **NO-GO para produccion general**.

Avance comprobado del plan: **94%** (avance de trabajo y evidencia; no
equivale a autorizacion GO para produccion general).

Fecha de corte: 2026-07-24.

Repositorio auditado: `D:\powerfulcontrolsystem`.

Rama auditada: `main`.

Commit auditado: `5bbd4e62` (merge publicado mediante `rs`; contiene el cierre
Trivy y la integracion visual de Nextcloud).

Referencias locales al corte:

- `main`: `5bbd4e62`.
- `origin/main`: `5bbd4e62`.

Modelo previsto para ejecutar este plan: **Codex Terra, esfuerzo medio**.

## 1. Proposito y limite de esta entrega

Este documento convierte la revision general del proyecto en una secuencia de
trabajo autocontenida, verificable y suficientemente explicita para que Codex
Terra medio pueda ejecutarla sin depender de la memoria de la auditoria.

La creacion de este documento **no autoriza su ejecucion**, no autoriza `rs`, no
autoriza despliegues, pagos, facturas DIAN, mensajes, correos ni cambios de datos
productivos. El agente debe detenerse al terminar de redactarlo para que el
usuario lo lea, cambie el modelo y apruebe expresamente la ejecucion.

La auditoria combino lectura de documentacion vigente, inventarios completos de
archivos y patrones, revision profunda de areas criticas, suites Go, validadores
del repositorio, comprobaciones HTTP publicas y una prueba visual real de bajo
impacto. "Revision completa" significa cobertura sistematica del arbol y sus
contratos; no significa afirmar que cada linea fue demostrada semanticamente en
produccion. Los huecos que requieren infraestructura, proveedores o mutaciones
reales se mantienen como bloqueadores, no como resultados aprobados.

## 2. Instrucciones obligatorias para Codex Terra medio

Antes de modificar codigo:

1. Leer, en este orden, `AGENTS.md`,
   `documentos/contexto_general_del_sistema.md`,
   `documentos/contexto_especifico_del_sistema.md`,
   `documentos/contexto_codex.md` y este plan completo.
2. Consultar para cada fase `documentos/mapa_modulos.md`,
   `documentos/flujos_operativos.md`, `documentos/comandos_codex.md`,
   `documentos/decisiones_tecnicas.md` y
   `documentos/checklist_seguridad_endpoint_multiempresa.md`.
3. Revisar el contexto especifico del modulo y las matrices de roles, base de
   datos, arquitectura y archivos antes de editarlo.
4. Confirmar rama, SHA, `git status --short` y diferencias contra `main` y
   `origin/main`. Si el corte cambio, actualizar primero la linea base de este
   plan y repetir los gates afectados.
5. Ejecutar **una sola fase o bloque acotado por vez**. No intentar resolver el
   plan completo en una edicion masiva.
6. Mantener Go puro, PostgreSQL unico, aislamiento por `empresa_id` y cero
   secretos en codigo, consola, documentos, capturas o commits.
7. No agregar dependencias externas ni modificar `go.mod` sin autorizacion
   explicita. Si una fase parece necesitarlas, detenerse y proponer la alternativa
   con libreria estandar.
8. No editar migraciones ya aplicadas. Toda evolucion nueva debe ser inmutable,
   registrada, ordenada, idempotente cuando corresponda y probada desde cero y
   sobre una copia representativa.
9. No usar `rs`, no publicar, no activar proveedores y no cambiar datos reales
   hasta llegar a la compuerta final y recibir autorizacion expresa.
10. Nunca reutilizar ni escribir en el plan la clave entregada para la prueba.
    Cualquier autenticacion posterior debe usar entrada segura del usuario o una
    sesion ya abierta.

Condiciones de parada inmediata:

- Hace falta captcha, MFA, firma, llave privada, token o aprobacion de un portal.
- Una prueba generaria un cobro, factura fiscal, correo, WhatsApp, alta/baja de
  cuenta, borrado, devolucion, cierre de caja o cambio irreversible.
- No se puede demostrar la empresa efectiva de una operacion multiempresa.
- El arbol contiene cambios ajenos que se superponen con la fase.
- Falla una migracion, restauracion, prueba A/B de tenant, prueba financiera o
  compuerta de seguridad.
- El entorno real no coincide con el SHA candidato.

En cualquiera de esos casos: conservar evidencia redactada, no improvisar un
atajo y solicitar decision al usuario.

## 3. Linea base comprobada el 2026-07-21

### 3.1 Alcance inventariado

- 556 archivos Go, aproximadamente 11,1 MB.
- 187 archivos Go de prueba.
- 662 funciones `Test`, `Benchmark` o `Fuzz` inventariadas.
- 309 archivos HTML, 77 JavaScript, 3 CSS, 21 PowerShell y 31 shell.
- La documentacion contiene un corpus grande de artefactos generados y
  referencias: el barrido textual incluyo `documentos`, y la lectura manual se
  concentro en fuentes vigentes, contratos operativos y reportes de gates.
- Los mayores focos de complejidad incluyen:
  `backend/handlers/modulos_faltantes.go` (13.444 lineas),
  `backend/handlers/payments_handlers.go` (6.900),
  `backend/handlers/reportes.go` (6.675),
  `backend/db/productos.go` (5.654),
  `backend/handlers/empresa_permisos.go` (3.869) y los flujos de carrito,
  facturacion, creditos y configuracion.

### 3.2 Pruebas y validadores que pasaron

- `go test ./... -count=1`: aprobado en todos los paquetes ejecutables.
- `go vet ./...`: aprobado.
- Inventario bootstrap: vigente, 153 funciones y 122 pasos heredados.
- Linea base de auditoria runtime: 132 llamadas `Ensure*` fuera del migrador:
  72 en arranque protegido, 1 en proceso de plataforma y 59 en trafico HTTP.
  Siete lotes de ejecucion del Plan 105 lo redujeron a 104 y 31 en HTTP; ver
  P105-003 y el inventario generado vigente.
- Matriz multiempresa: 203 rutas empresariales, 203 con wrapper detectado,
  sin duplicados. Esta comprobacion estatica no reemplaza pruebas A/B.
- OpenAPI, contrato de despliegue, sintaxis PowerShell/JavaScript, seguridad,
  permisos, roles, modulos criticos, UX, pagos, observabilidad, SLO, hardening,
  soporte y `git diff --check`: aprobados por el preflight estricto.
- El preflight estricto completo sin repetir Go termino correctamente con
  reporte en `.gotmp/plan105_preflight_final_20260721/`.
- El preflight con Go incluido alcanzo los mismos controles iniciales, pero la
  invocacion fue interrumpida por el limite local de 60 segundos mientras
  repetia `go test`; no se registra como fallo funcional.

### 3.3 Pruebas que no quedaron demostradas

- `go test -race ./...` no pudo ejecutarse: el equipo tiene `CGO_ENABLED=0` y
  no dispone de compilador C. Debe correr en CI o builder Linux compatible.
- `documentos/migration_validation_report.md` continua pendiente de PostgreSQL
  efimero.
- `documentos/backup_restore_report.md` continua pendiente de simulacro
  efimero actual con RPO/RTO.
- `documentos/load_test_report.md` continua pendiente de staging.
- `documentos/integration_validation_report.md` declara validacion local
  parcial; falta end to end.
- El staging registrado fue parcial, sin snapshot anonimizado funcional, sin
  publicacion HTTP completa y sin proveedores externos.

### 3.4 Comprobacion publica y prueba visual real

Superficie probada: `https://powerfulcontrolsystem.com`, empresa
**Powerful Control System**, `empresa_id=12`.

- `/`, `/health`, `/ready` y `/login.html` respondieron HTTP 200; salud y
  readiness devolvieron estado correcto en esa observacion puntual.
- Inicio de sesion real completado y empresa correcta seleccionada.
- Panel superadministrador observado con puntuacion general 72/100, PostgreSQL
  marcado visualmente como critico, prioridad OnlyOffice, indice de hardening
  0/100 y cuatro hallazgos. Estas lecturas deben reconfirmarse desde una sesion
  fresca y correlacionarse con metricas del VPS antes de corregir.
- Flujo de venta directa: busqueda por nombre de `menta`, resultado SKU 1 por
  COP 100, agregado al carrito, cantidad 1 -> 2 -> 1 y totales
  COP 100 -> 200 -> 100. No se pago, facturo, cerro, cancelo ni devolvio.
- Quedo **una unidad de `menta` en el carrito abierto**, fila observada 205,
  total COP 100. Puede existir reserva de inventario. Su limpieza es una accion
  de datos y requiere autorizacion; no debe ocultarse ni suponerse inocua.
- En vista movil no hubo desbordamiento de pagina, pero la tabla del producto
  requiere desplazamiento horizontal interno y oculta columnas/acciones.
- Se observo un error de consola de `MutationObserver.observe`: el parametro 1
  no era un `Node`. Debe reproducirse en una sesion autenticada limpia.
- Las advertencias de configuracion avanzada, permisos y cajas abiertas fueron
  vistas tambien durante una carga inicialmente no autenticada; no deben
  atribuirse al flujo autenticado hasta repetirlo desde una sesion limpia.
- El catalogo de radio mostro `Pais de emisoras: Panama`. Debe compararse con el
  pais fiscal/configuracion de la empresa antes de clasificarlo como defecto.

### 3.5 Hallazgos transversales del codigo y runtime

- La rama candidata no esta unificada con `main` ni con `origin/main`.
- Permanecen 31 llamadas `Ensure*` alcanzables desde trafico HTTP. El bootstrap
  heredado no puede considerarse retirado.
- El barrido encontro 23 lanzamientos `go func`; nueve estan concentrados en
  `backend/handlers/empresa_permisos.go`. Cada uno requiere clasificacion de
  durabilidad, cancelacion, recuperacion e idempotencia.
- El patron `err.Error()` aparece 1.602 veces. Es un inventario amplio, no prueba
  por si solo una filtracion; requiere clasificar respuestas HTTP, logs internos
  y serializaciones para impedir PII, SQL, rutas o secretos al cliente.
- No se encontraron politicas RLS en el repositorio. Los wrappers son una
  defensa valida, pero debe existir evidencia negativa A/B en todas las capas.
- La outbox esta implementada, pero la adopcion visible por productores de
  negocio sigue siendo estrecha frente al numero de integraciones criticas.
- CSP conserva tres coincidencias de `unsafe-inline`, incluidas utilidades web y
  Nginx de staging.
- El release exige digests para API, worker y migrador, pero Postgres, Redis,
  Mailu, Nginx, Certbot, OnlyOffice, RustDesk y monitoreo usan etiquetas.
- `cadvisor` usa `privileged: true`; debe justificarse y aislarse o reemplazarse
  por el minimo privilegio demostrable.
- `mobile/powerful_control_system_app` no contiene fuente versionada fuera de
  artefactos `build`/`.dart_tool`; no se puede reproducir una app movil.
- El servicio de voz incluye instalacion Python/Piper. La voz esta apagada por
  defecto para empresas existentes, pero el perfil debe quedar explicitamente
  fuera del lanzamiento o con excepcion arquitectonica aprobada.
- La normalizacion documental refinada reporta 218 secuencias sospechosas en
  `CHANGELOG.md`; los parámetros URL legítimos no cuentan como corrupción. Es
  deuda P2, no razon para falsear un
  GO operativo.

## 4. Estrategia de ejecucion

Orden obligatorio:

1. Fase 0: congelar candidato, alcance y matriz de evidencia.
2. Fase 1: migraciones y retiro de DDL del runtime.
3. Fase 2: aislamiento multiempresa y consistencia transaccional.
4. Fase 3: staging reproducible, restauracion, concurrencia y observabilidad.
5. Fase 4: proveedores reales autorizados y UX operativa.
6. Fase 5: hardening, artefactos inmutables, piloto y decision GO/NO-GO.
7. Fase 6: deuda P2 posterior al piloto, sin mezclarla con el hot path del
   lanzamiento salvo que bloquee un gate.

Cada item P105 debe producir:

- cambio acotado;
- pruebas automatizadas enfocadas y suite de regresion;
- evidencia fechada con SHA y entorno;
- documentacion y trazabilidad actualizadas;
- riesgo residual, rollback y decision `APROBADO`, `BLOQUEADO` o `NO APLICA`;
- ninguna marca aprobada basada solo en lectura de codigo.

## 5. Fase 0 - candidato, alcance y evidencia

### P105-001 - Fuente unica e inmutable de liberacion [P0]

Problema inicial: el commit auditado vivia en una rama divergente y no habia
una referencia unica demostrada para construir, probar y desplegar. La
verificacion estricta local del 2026-07-21 confirma que `origin/main` es
ancestro del candidato y que la rama tiene upstream; siguen faltando un arbol
limpio atribuible al SHA final y los tres digests inmutables.

Acciones:

1. Comparar historial y diff entre candidato, `main` y `origin/main`.
2. Clasificar commits que deben integrarse y cambios obsoletos; no hacer reset
   destructivo.
3. Integrar mediante PR revisable con CI verde.
4. Generar tag candidato firmado o protegido y registrar SHA exacto.
5. Construir API, worker y migrador una sola vez; publicar digests inmutables.
6. Usar los mismos digests en staging, piloto y produccion.

Pruebas minimas:

```powershell
git status --short
git log --oneline --decorate --graph --all -n 80
git diff --stat main...HEAD
git diff --stat origin/main...HEAD
node tools/deploy_pipeline_contract.mjs
```

Aceptacion: PR aprobada, checks requeridos verdes sobre el SHA final, tag y
tres digests registrados; staging ejecuta exactamente esos artefactos.

Parada/rollback: si una nueva subida cambia el SHA, invalidar aprobaciones y
repetir CI. Rollback de aplicacion solo a digests previamente validados y
compatible con el esquema ya aplicado.

### P105-002 - Alcance comercial y piloto controlado [P0]

Problema: el sistema contiene muchos modulos, integraciones y perfiles que no
poseen evidencia end to end equivalente. Produccion general no puede significar
"todo activo".

Acciones:

1. Crear matriz por modulo: habilitado, empresa piloto, roles, proveedor,
   datos sensibles, pruebas, observabilidad, soporte y rollback.
2. Definir el alcance inicial con el usuario. Por defecto, todo modulo sin
   evidencia real queda apagado mediante configuracion por empresa.
3. Separar `core POS/caja/carrito/inventario/usuarios` de DIAN, pagos, correo,
   WhatsApp, Rappi, IA, voz, Nextcloud, OnlyOffice y RustDesk.
4. Confirmar que flags, menus, endpoints y jobs respetan licencia, permiso y
   `empresa_id`; no basta ocultar un enlace.

Aceptacion: matriz firmada por el usuario, empresa piloto definida, modulos no
validados deshabilitados en frontend, API y worker, y runbook de reactivacion.

Avance 2026-07-21: se creo el borrador
`documentos/matriz_alcance_piloto_plan_105.md`. No equivale a aprobacion ni
activa modulos; faltan la decision del usuario y la comprobacion tecnica de
cada control efectivo.

## 6. Fase 1 - migraciones y runtime sin DDL

### P105-003 - Migraciones inmutables y retiro de `Ensure*` [P0]

Problema al iniciar el plan: 132 llamadas heredadas permanecian fuera del
migrador, 59 en trafico HTTP. Siete lotes las redujeron a 104 y 31; aun asi,
`PCS_RUNTIME_SCHEMA_BOOTSTRAP=1` sigue siendo un puente operativo.

Ejecucion local acumulada (2026-07-21): los cuatro lotes ya retiraron DDL de
preferencias por estacion, GPS/taxi, grafologia, energia solar, hoja de vida,
reservas, programacion de reportes (empresa y superadmin) y Chat/Tareas. Cada
reemplazo usa un `*SchemaReady` de solo lectura y tiene prueba de conexion nula;
el inventario generado es la fuente de verdad. No se ha probado este cambio
sobre PostgreSQL real porque Docker no esta disponible en este equipo.

Siguiente lote que debe tomar Terra: no elegir por cantidad. Priorizar solo un
handler cuyos CRUD no vuelvan a invocar `Ensure*` internamente y cuya prueba se
pueda correr contra PostgreSQL efimero. Mantener fuera de los lotes rapidos
Control Electrico (aprovisiona Raspberry), Corte de Caja/Sensor Puertas,
Creditos/Eventos Contables, permisos/roles, correo corporativo, Nextcloud,
Rappi, voz y tarifas hasta clasificar si cada `Ensure*` tambien hace seed,
backfill o efectos externos.

Archivos iniciales:

- `documentos/arquitectura/inventario_bootstrap_ensure.md`
- `documentos/arquitectura/inventario_runtime_ensure.md`
- `backend/db/legacy_schema_catalog.go`
- `backend/db/legacy_schema_catalog_manifest_generated.go`
- `backend/cmd/pcs-migrate/`
- handlers listados como trafico HTTP por el inventario

Acciones por lotes pequenos:

1. Congelar la linea base con los dos inventarios `--check`.
2. Clasificar cada `Ensure*`: DDL, seed, backfill, configuracion o verificacion.
3. Crear migraciones nuevas para DDL/seed/backfill, con checksum y lock.
4. Reemplazar llamadas HTTP y de API/worker por verificadores de readiness que
   solo consulten catalogo y fallen cerrados sin mutar.
5. Mover configuraciones de entorno al aprovisionamiento explicito.
6. Reducir el inventario a cero llamadas que muten esquema desde trafico, API o
   worker; justificar temporalmente cualquier verificador cuyo nombre quede.
7. Probar bootstrap desde base vacia y upgrade desde snapshot anonimizado.
8. Ejecutar API y worker con `PCS_RUNTIME_SCHEMA_BOOTSTRAP=0`.

Pruebas:

```powershell
node tools/ensure_bootstrap_inventory.mjs --check
node tools/runtime_ensure_inventory.mjs --check
node tools/migration_audit.mjs --strict
Set-Location backend
go test ./db ./handlers ./internal/platform/... ./cmd/... -count=1
go test ./... -count=1
```

En PostgreSQL efimero ejecutar dos caminos: instalacion desde cero y actualizacion
de snapshot. Capturar lista de migraciones, checksums, duracion, locks, esquema
final y segunda ejecucion sin cambios.

Aceptacion:

- cero DDL en API, worker o solicitudes HTTP;
- migrador unico con historial completo e inmutable;
- instalacion nueva y upgrade aprobados;
- API/worker arrancan y atienden con bootstrap 0;
- `migration_validation_report.md` deja de estar pendiente con evidencia real.

Rollback: migraciones expansivas primero; no eliminar columnas/tablas hasta que
el codigo anterior deje de necesitarlas. Cada migracion destructiva requiere
backup probado, ventana y autorizacion separada.

## 7. Fase 2 - aislamiento y consistencia de negocio

### P105-004 - Suite negativa multiempresa A/B [P0]

Problema: 203 wrappers detectados no prueban el aislamiento de SQL, archivos,
cache, jobs, exportaciones ni URLs compartibles.

Avance 2026-07-21: `documentos/arquitectura/matriz_pruebas_tenant_ab_plan_105.md`
documenta los controles unitarios ya presentes y los casos P0 que deben correr
en PostgreSQL/staging. No existen aun fixtures A/B ni evidencia de objetos de B
inmutables despues de una solicitud cruzada.

Acciones:

1. Crear fixtures Empresa A/Empresa B con usuarios equivalentes y datos
   marcadores no sensibles.
2. Para cada ruta empresarial, probar tenant correcto y manipular query,
   cabecera, path, JSON, formulario, cookies e identificadores repetidos.
3. Cubrir roles sin permiso, licencia vencida, cuenta desactivada y sesion de
   otra empresa.
4. Validar consultas, conteos, joins, reportes, PDFs, CSV/XLSX, backups,
   busquedas, archivos, cache, jobs/outbox, notificaciones y auditoria.
5. Verificar que un identificador valido de B usado desde A produzca 403/404 sin
   revelar existencia, nombre, total ni metadatos.
6. Evaluar RLS por tabla critica. Si no se adopta, documentar por que, defensa
   equivalente, ownership de conexion y pruebas obligatorias.
7. Integrar el conjunto negativo en CI; mantener el inventario de rutas como
   gate de crecimiento.

Archivos iniciales: `backend/main.go`, wrappers de `backend/handlers`, capa
`backend/db`, `tools/tenant_route_inventory.mjs`, matrices de roles y seguridad.

Aceptacion: cero fuga A/B en todas las familias P0; evidencia incluye request,
rol, empresa efectiva, respuesta redactada y comprobacion directa de que B no
cambio. Cualquier fuga mantiene `NO-GO` automatico.

### P105-005 - Transacciones, idempotencia, outbox y goroutines [P0]

Problema: pagos, venta, caja, licencias, facturacion y entregas externas deben
resistir duplicados, carreras, timeouts y reinicios. Hay 23 `go func` y adopcion
parcial de outbox.

Avance 2026-07-21: `documentos/arquitectura/inventario_goroutines_plan_105.md`
clasifica los 23 lanzamientos. Identifica como P0 el lease del worker, el
despacho de control electrico y el snapshot de permisos; es inventario estatico,
no evidencia de resiliencia ni autorizacion para cambiar efectos externos.

Acciones:

1. Inventariar cada goroutine con propietario, contexto, cancelacion, timeout,
   persistencia, reintento y destino. Ningun efecto financiero/fiscal debe
   depender de una goroutine sin estado durable.
2. Trazar transacciones de carrito -> pago -> venta -> inventario -> caja ->
   factura -> comision -> notificacion.
3. Definir claves idempotentes por operacion y proveedor; crear constraints
   unicos y estados monotonicamente validos.
4. Publicar eventos de negocio dentro de la misma transaccion en outbox.
5. Mover entregas externas al worker durable con lease, retry acotado, backoff,
   dead-letter/estado terminal y correlacion.
6. Bloquear doble pago, doble cierre, stock negativo, doble factura, doble
   comision y replay de webhook.
7. Probar caidas en cada frontera: antes/despues de commit, timeout de proveedor,
   worker muerto, respuesta perdida y reintento concurrente.

Archivos iniciales:

- `backend/handlers/carritos_compras.go`
- `backend/db/carritos_compras.go`
- `backend/handlers/payments_handlers.go`
- `backend/handlers/facturacion_electronica.go`
- `backend/handlers/modulos_faltantes.go`
- `backend/handlers/empresa_permisos.go`
- `backend/db/outbox.go`, `backend/db/async_jobs.go`
- `backend/internal/platform/worker/`

Aceptacion: pruebas deterministas de duplicado y carrera pasan, un reinicio no
pierde ni duplica efectos, y toda transicion critica conserva correlacion y
auditoria por empresa.

### P105-006 - Errores seguros y observabilidad util [P1]

Avance 2026-07-21: se agrego redaccion P0 al proxy y configuracion de
`voice_stream`; respuestas 5xx/502 devuelven codigo estable y `request_id`,
con prueba de canarios. El inventario restante sigue requiriendo clasificacion
por familia antes de cualquier reemplazo masivo.

Problema: las 1.602 coincidencias `err.Error()` pueden ser logs legitimos o
respuestas inseguras; se requiere clasificacion verificable.

Avance 2026-07-21: el inventario estatico en
`documentos/arquitectura/inventario_errores_publicos_plan_105.md` midio 320
respuestas 5xx y 75 campos JSON directos con `err.Error()`. Ocho lotes
redactan mantenimiento, alertas, correos masivos, seguridad VPS y cupo del
agente fiscal, incluido el streaming del chat publico; dejan 51 JSON directos.
El validador de propuestas IA ahora usa lista blanca y el inventario queda en
49. El lote adicional de `empresa_buzon.go` ya no concatena causas internas al
guardar configuracion, limpiar adjuntos antiguos o persistir un archivo;
conserva diagnostico tecnico en log y tiene regresion estatica. Los lotes P0
restantes son pagos,
productos, voz, identidad y multiempresa. El lote de `chat_con_ia_global_super.go`
tambien redacta errores del proveedor en respuestas HTTP y SSE, conserva el
diagnostico solo en log y agrega regresion estatica; el inventario numerico debe
regenerarse antes del cierre. En `super_chat_ia_logica.go` se redactaron 13
respuestas 500 de lectura de configuracion/consumo IA; falta agregar correlacion
estructurada y regenerar el inventario. En `payments_handlers.go` se redactaron
12 respuestas 500 de licencias, alcance y activacion; falta cubrir los demas
flujos de pagos y regenerar el inventario. El inventario no prueba aun
ausencia de fuga en todas las familias.

Acciones:

1. Inventariar todos los puntos que escriben respuestas HTTP/JSON.
2. Crear taxonomia estable de codigos publicos y mensajes genericos.
3. Conservar causa tecnica, `request_id`, `empresa_id` autorizado y operacion
   solo en logs estructurados; redactar PII, SQL, DSN, rutas y secretos.
4. Añadir tests que inyecten errores con canarios sensibles y confirmen ausencia
   en cuerpo, cabeceras y logs expuestos.
5. Activar gate que impida nuevas respuestas directas con `err.Error()` salvo
   allowlist documentada.

Aceptacion: cero canarios en respuestas o logs accesibles al usuario, soporte
puede correlacionar el codigo publico con un evento interno y no se pierde
capacidad de diagnostico.

## 8. Fase 3 - staging, continuidad y rendimiento

### P105-007 - Staging reproducible equivalente [P0]

Problema: el reporte actual solo demuestra ejecucion tecnica parcial en red
privada, no el flujo publicado completo.

Avance 2026-07-22: la comprobacion externa de solo lectura contra
`https://staging.powerfulcontrolsystem.com/health` fue bloqueada por
`ERR_CERT_DATE_INVALID`. No se acepto ni eludio el certificado. Este es un
NO-GO de P105-007: renovar/corregir la cadena TLS del dominio de staging y
repetir `/health`, `/ready` y el checklist autenticado antes de cualquier
carga, restore o matriz A/B.

Actualizacion 2026-07-22: las probes TLS de OnlyOffice y Nextcloud tambien
fallan por certificado expirado, mientras `/health` y `/ready` del nucleo PCS
responden 200. Tratar como incidente unico de certificados/edge: renovar la
cadena de los tres hostnames, validar sin `--insecure`, y solo despues reanudar
las pruebas de continuidad y proveedores.

Acciones:

1. Crear infraestructura aislada con dominios/certificados de staging,
   PostgreSQL, Redis, almacenamiento, API, worker, migrador y frontend.
2. Restaurar snapshot reciente **anonimizado**; validar que el proceso de
   anonimizacion elimine PII, secretos, tokens, firmas y correos reales.
3. Desplegar los mismos digests candidatos de P105-001.
4. Probar HTTPS externo, cookies seguras, CORS/CSP, redirects, uploads,
   descargas, worker, cron, correo sink y webhooks simulados.
5. Ejecutar matriz de dos empresas, superadmin, administrador, cajero y rol
   restringido en escritorio y movil.
6. Guardar capturas, HAR/red de bajo riesgo, consola, logs correlacionados y
   resultados de DB sin secretos.

Aceptacion: checklist de staging completo sobre URL publicada, datos
anonimizados, mismo SHA/digest y cero errores de consola/red en flujos P0.

### P105-008 - Backup, restauracion y rollback con RPO/RTO [P0]

Problema: existen scripts/runbooks y evidencia historica, pero los reportes
vigentes de este candidato siguen pendientes.

Avance 2026-07-21: el restore drill acepta `-RestoreImage` y registra la imagen
PostgreSQL usada; por defecto coincide con el Compose vigente. En staging debe
recibir la referencia inmutable candidata. Falta ejecutar el drill autorizado y
medir RPO/RTO reales.

Acciones:

1. Generar snapshot del entorno autorizado con PostgreSQL y todos los volumenes
   persistentes: archivos privados, uploads, certificados, Mailu, OnlyOffice y
   servicios externos aplicables.
2. Copiarlo a almacenamiento externo cifrado y verificar checksum/retencion.
3. Ejecutar `scripts/vps_restore_validation.ps1 -ExecuteDrill` en destino
   efimero, nunca encima de produccion.
4. Arrancar el stack restaurado, ejecutar smoke funcional y verificar conteos,
   relaciones, archivos y migraciones.
5. Medir RPO/RTO reales contra los objetivos del SLO.
6. Ensayar rollback de aplicacion y definir compatibilidad hacia atras de cada
   migracion.
7. Documentar responsables, llaves, dependencias DNS y pasos de desastre total.

Aceptacion: restauracion funcional independiente, RPO/RTO medidos y dentro del
objetivo, checksums validos, datos/archivos consistentes y rollback ensayado.

### P105-009 - Carrera, carga, capacidad y degradacion [P0]

Problema: la suite race no corrio y el reporte de carga espera staging.

Estado verificado 2026-07-21: `.github/workflows/professional-ci.yml` ya
ejecuta `go test -race ./...` sobre Linux con Go 1.25.12. El equipo local no
puede ejecutar race por ausencia de compilador/CGO; no crear un segundo job.
Antes de cerrar la fase, exigir el resultado verde sobre el SHA inmutable de
P105-001 y adjuntarlo a la evidencia de release.

Acciones:

1. Confirmar que el job Linux existente con CGO ejecuta `go test -race ./...`
   sobre el SHA inmutable y que no fue omitido por condicion de workflow.
2. Priorizar paquetes de carrito, pagos, caja, licencias, permisos, outbox,
   worker, reportes y facturacion.
3. Ejecutar carga realista en staging: login, catalogo, busqueda, carrito,
   checkout simulado, reportes y jobs.
4. Medir p50/p95/p99, errores, conexiones, locks, CPU, memoria, disco, Redis,
   goroutines, colas y saturacion.
5. Probar reinicio de worker/API, proveedor lento, DB/Redis temporalmente no
   disponibles y limites de payload.
6. Definir capacidad inicial, limites, autoservicio prohibido y alertas.

Aceptacion: race limpio, SLO sostenido con margen acordado, cero corrupcion o
duplicado y comportamiento de sobrecarga controlado (429/503, timeout y retry).

### P105-010 - Observabilidad viva y simulacros de alerta [P0]

Problema: validadores estaticos pasan, pero el panel real mostro indicadores
criticos/0 y prioridad OnlyOffice.

Avance 2026-07-22: produccion respondio `/health={"status":"ok"}` y
`/ready={"status":"ready"}`. En el panel autenticado, la lectura de alertas
reporto disco 44.9%, 22 conexiones PostgreSQL y 167 sesiones administrativas,
todos bajo sus umbrales; sin embargo, el tablero agregado seguia marcando
PostgreSQL/continuidad como criticos. Debe correlacionarse esa discrepancia
antes de cerrar observabilidad; no se ejecutaron botones de reinicio, evaluacion
ni envio de correo.

Acciones:

1. Correlacionar cada indicador del panel con Prometheus/log/consulta fuente.
2. Verificar si PostgreSQL "critico" y hardening 0/100 son fallos reales,
   ausencia de datos o errores de calculo.
3. Corregir fuente o infraestructura y añadir estados `sin datos` para evitar
   falsos ceros.
4. Probar alertas de DB, disco, memoria, latencia, 5xx, worker detenido, cola
   creciente, backup vencido, certificado, correo, almacenamiento y proveedor.
5. Ejecutar un simulacro por alerta P0 y confirmar recepcion, deduplicacion,
   runbook, escalamiento y recuperacion.

Aceptacion: dashboards reflejan metricas comprobables, alertas llegan a un
responsable, runbooks recuperan el servicio y no quedan indicadores P0 sin
explicacion.

## 9. Fase 4 - proveedores y experiencia real

### P105-011 - Evidencia real de proveedores habilitados [P0]

Problema: contratos y auditorias estaticas no demuestran aceptacion del
proveedor. El staging anterior no ejecuto integraciones externas.

Regla: probar solo proveedores incluidos en P105-002 y con autorizacion expresa.
Los demas deben permanecer deshabilitados.

Matriz minima:

| Proveedor | Evidencia obligatoria para aprobar |
|---|---|
| DIAN | Factura autorizada de bajo riesgo, XML/ZIP/firma correctos, `GetStatusZip` con `StatusCode=00`, CUFE y estado persistido; notas/eventos solo si entran al alcance. |
| Wompi | Transaccion autorizada, webhook firmado, replay idempotente, conciliacion y estado final consistente. |
| ePayco | Firma/callback valido, replay, conciliacion, retorno y estado final consistentes. |
| Mailu/SMTP | Buzon provisionado, envio y recepcion reales, DKIM/SPF/DMARC, rebote y redaccion de logs. |
| WhatsApp | Plantilla/canal autorizado, entrega y error controlado sin exponer datos. |
| Rappi | Credenciales y webhook del comercio autorizado, mapeo por empresa, replay y conciliacion. |
| Nextcloud | Cuenta/cuota de empresa, aislamiento A/B, upload/download/delete y backup/restauracion. |
| OnlyOffice | Edicion real, JWT, archivo privado, concurrencia y descarga/eliminacion seguras. |
| RustDesk | Sesion autorizada, limites, consentimiento, auditoria y cierre. |
| OpenAI/IA | Clave por empresa o politica global aprobada, aislamiento, limites, costos, redaccion y fallo seguro. |

Para cada proveedor registrar: empresa, modo, timestamp, request/correlation ID,
resultado del proveedor, filas locales, reintento, logs redactados y rollback.
Nunca incluir secretos, documentos fiscales completos ni PII en el repositorio.

Aceptacion: evidencia real y reconciliada para cada proveedor habilitado. Un
TrackId inicial, respuesta pendiente, simulador local o prueba estatica no
aprueba esta compuerta.

### P105-012 - Cierre visual y accesibilidad de flujos P0 [P1]

Problemas observados: tabla de carrito comprimida en movil, error potencial de
`MutationObserver`, avisos de contexto por confirmar y pais de radio dudoso.

Acciones:

1. Repetir desde sesion autenticada limpia el flujo exacto de venta directa y
   capturar consola/red antes y despues.
2. Identificar el `MutationObserver` que recibe un nodo inexistente; esperar el
   elemento o proteger la llamada, sin ocultar errores legitimos.
3. Convertir las filas del carrito a diseño movil legible o aplicar columnas
   prioritarias/acciones persistentes; el usuario no debe depender de scroll
   horizontal para cantidad, precio, total o accion principal.
4. Probar anchos 320, 360, 390/434, 768, 1024 y escritorio; teclado, foco,
   lector, contraste, zoom 200 % y mensajes de estado.
5. Reproducir advertencias de configuracion/permisos/caja solo despues de login;
   corregir si persisten, ignorar evidencia contaminada por carga anonima.
6. Comparar `Pais de emisoras` con la preferencia real por empresa. Corregir
   aislamiento/configuracion solo si el dato no corresponde.
7. Probar cajero, administrador y rol restringido; verificar que la UI y API
   coincidan.

Aceptacion: cero excepciones de consola, red limpia, flujo P0 usable sin columnas
ocultas, permisos coherentes y capturas desktop/movil asociadas al SHA.

### P105-013 - Resolver hallazgos del panel de ciberseguridad [P0]

Actualizacion 2026-07-24: el reporte ahora declara si audita `container` u
`host-local`, marca cobertura incompleta si una herramienta habilitada falla y
no presenta el puerto loopback como exposicion publica. Trivy omite solo los
archivos de credenciales no legibles del contenedor. Falta ejecutar y revisar
el escaneo vivo post-despliegue; una auditoria completa de Ubuntu requiere la
CLI desde el host y evidencia externa de puertos.

Acciones:

1. Extraer los cuatro hallazgos sin revelar secretos y clasificarlos por fuente,
   severidad, activo, explotabilidad y propietario.
2. Validar que no sean falsos positivos por telemetria ausente.
3. Corregir todos los criticos/altos y crear regresiones/preflight.
4. Para medios/bajos, corregir o documentar riesgo aceptado por el usuario con
   fecha de vencimiento.
5. Repetir escaneo autenticado y no autenticado, cabeceras, TLS, cookies,
   permisos, dependencias de imagen y exposicion de puertos.

Aceptacion: cero hallazgos criticos/altos, los demas tienen mitigacion o
aceptacion explicita, y el indice se deriva de evidencia viva.

### P105-014 - Politica CSP sin `unsafe-inline` [P1]

Actualizacion 2026-07-24: PCS permite el origen exacto de Nextcloud en
`frame-src` y su pagina empresarial movio estilos a un archivo externo. La
prueba autenticada posterior a `rs` confirma que Nextcloud 29 conserva
`frame-ancestors 'self'` y `X-Frame-Options: SAMEORIGIN`; un segundo encabezado
Nginx no lo amplía, lo intersecta y el navegador rechaza el iframe. Se restauro
el sitio Nginx desde su backup. El script de host ahora falla antes de escribir
si detecta esa directiva; queda pendiente una solucion soportada por Nextcloud
o un filtro de encabezados que preserve sus nonces. No cierra la retirada global
de `unsafe-inline`: el inventario de scripts, estilos y handlers inline sigue
siendo amplio.

Acciones:

1. Inventariar scripts/estilos inline y handlers HTML dinamicos.
2. Migrar a archivos propios o nonces por respuesta; no usar hashes globales
   para contenido variable sin control.
3. Aplicar primero `Content-Security-Policy-Report-Only`, revisar violaciones y
   luego hacer enforcement.
4. Cubrir login, panel, carrito, pagos, DIAN, reportes y paginas publicas.

Aceptacion: cero `unsafe-inline` en politica de produccion, cero violaciones P0
y pruebas visuales completas.

### P105-015 - Almacenamiento privado compartido [P1]

Problema: replicas, documentos y restauracion exigen storage compartido u
Object Storage; archivos locales no son suficientes para produccion escalable.

Acciones:

1. Inventariar todas las escrituras/lecturas/borrados de archivos.
2. Aplicar la interfaz de storage existente a documentos, adjuntos, audio,
   exports, backups y temporales.
3. Definir claves con `empresa_id`, nombres no adivinables, metadatos, cifrado,
   retencion, antivirus/validacion MIME, cuotas y lifecycle.
4. Usar URLs firmadas cortas o proxy autorizado; nunca exponer bucket publico.
5. Probar A/B, replica A/B, fallo de nodo, borrado, expiracion y restauracion.

Aceptacion: todas las familias privadas usan storage aprobado, aislamiento A/B
y restore demostrados; readiness impide replicas sobre storage local.

Avance 2026-07-21: inventario estatico en
`documentos/arquitectura/inventario_storage_plan_105.md`; la capa privada
actual reduce exposicion por empresa, pero no existe adaptador de Object Storage
ni evidencia de replica/restore. P105-015 permanece pendiente.

## 10. Fase 5 - supply chain, piloto y liberacion

### P105-016 - Docker, imagenes y minimo privilegio [P1]

Acciones:

1. Fijar por digest todas las imagenes efectivamente habilitadas, no solo las
   tres PCS. Registrar origen, version, licencia y fecha de actualizacion.
2. Escanear imagenes y SBOM; no aprobar vulnerabilidades criticas/altas sin
   mitigacion y aceptacion.
3. Aplicar usuario no root, `read_only`, `no-new-privileges`, `cap_drop`,
   limites CPU/RAM/PID, healthchecks y redes privadas donde aplique.
4. Justificar y aislar `cadvisor privileged`; limitar mounts y publicacion.
5. Verificar que secretos vengan de archivos/secret store con permisos, no de
   imagenes, compose versionado ni logs.
6. Probar actualizacion y rollback usando digests.

Aceptacion: manifiesto inmutable, cero tag mutable en servicios habilitados,
minimo privilegio demostrado, escaneo aprobado y rollback reproducible.

Avance 2026-07-21: la compuerta exige digest para API, migrador y worker, y
los roles PCS usan filesystem de solo lectura y capacidades reducidas. El
inventario de servicios auxiliares, SBOM, escaneo y rollback queda en
`documentos/arquitectura/inventario_supply_chain_plan_105.md`; P105-016 sigue
pendiente hasta validar todos los servicios habilitados por digest.

### P105-017 - Piloto productivo limitado y ventana de observacion [P0]

Precondiciones: P105-001 a P105-011 y P105-013 aprobados; P1 que afecte el
alcance resuelto; autorizacion explicita del usuario.

Acciones:

1. Crear backup/restauracion reciente y confirmar rollback.
2. Publicar los digests validados durante ventana acordada.
3. Ejecutar migrador una vez, revisar ledger y arrancar API/worker sin bootstrap
   runtime.
4. Smoke de salud, login, empresa, permisos, producto, carrito, caja y proveedor
   incluido; evitar datos innecesarios.
5. Operar solo empresa(s), usuarios y modulos del piloto.
6. Vigilar SLO, errores, locks, colas, jobs, recursos y alertas durante el periodo
   acordado.
7. Registrar incidentes, resolver P0/P1 y decidir ampliar, mantener o revertir.

Aceptacion: sin corrupcion, fuga, duplicados, incidentes P0/P1 ni SLO incumplido;
soporte y rollback disponibles. El piloto no autoriza automaticamente el resto
de empresas o modulos.

## 11. Fase 6 - deuda estructural posterior o paralela segura

### P105-018 - Modularizacion de archivos gigantes [P2]

Separar por dominio, sin reescritura masiva:

1. Crear caracterizacion de comportamiento antes de mover codigo.
2. Extraer un dominio por PR conservando API publica.
3. Priorizar `modulos_faltantes.go`, pagos, reportes, productos, permisos,
   carrito y `main.go`.
4. Prohibir ciclos y paquetes "utils" genericos; mantener ownership claro.
5. Medir cobertura, tiempos de build/test y complejidad despues de cada lote.

Aceptacion: archivos con responsabilidades acotadas, sin cambio funcional no
probado y suite completa verde.

### P105-019 - Normalizacion documental UTF-8 [P2]

Corregir en lotes revisables las 218 secuencias sospechosas de `CHANGELOG.md`.
No hacer reemplazos ciegos
que alteren comandos, hashes o evidencia historica.

Aceptacion: auditor de normalizacion en verde y diff revisado semanticamente.

### P105-020 - Cliente movil reproducible [P2 o fuera de alcance]

Decidir con el usuario una de dos rutas:

- versionar proyecto fuente completo, toolchain fijado, pruebas, firma externa,
  CI y politica de secretos; o
- declarar movil fuera de este lanzamiento y retirar promesas/artefactos no
  reproducibles de la documentacion operativa.

Avance 2026-07-21: `mobile/powerful_control_system_app` solo contiene
artefactos locales (`build` y `.dart_tool`), no fuente ni manifiesto
reproducible. Inventario y criterio de decision en
`documentos/arquitectura/inventario_cliente_movil_plan_105.md`; requiere
eleccion de alcance y permanece pendiente.

No reconstruir fuente desde `build`/`.dart_tool` ni publicar binarios sin origen.

### P105-021 - Voz Python/Piper [P2 o fuera de alcance]

Mantener el perfil apagado en el lanzamiento inicial. Si el usuario lo incluye,
documentar excepcion a Go puro, imagen/digest, dependencias/modelo/licencia,
aislamiento de red, limites, privacidad, disponibilidad y rollback. Alternativa:
servicio externo o implementacion compatible con la politica del proyecto.

### P105-022 - Higiene de datos reales de QA [P1 antes del piloto]

1. Con autorizacion, revisar el carrito abierto de empresa 12 y la reserva de
   `menta` fila 205.
2. Usar la accion de negocio correcta y auditable; no borrar filas directamente.
3. Verificar carrito, reserva, stock, auditoria y caja antes/despues.
4. Crear politica de datos QA: productos marcados, carrito temporal, responsable,
   caducidad y limpieza autorizada.

Aceptacion: no queda impacto de la prueba o queda expresamente aceptado y
trazado; ninguna venta, devolucion o factura fue creada por limpieza accidental.

### P105-023 - Credito y cuentas por pagar a proveedores multiempresa [P0 antes del piloto financiero]

Objetivo: consolidar un modulo de obligaciones con proveedores que permita
registrar compras a credito, facturas por pagar, anticipos, abonos, ajustes,
vencimientos, conciliaciones y reportes de cartera, sin duplicar los flujos o
tablas ya existentes de CxP. Cada lectura, escritura, exportacion, archivo,
trabajo en segundo plano y auditoria debe estar aislado por `empresa_id`.

Referencia funcional investigada: el enfoque documentado publicamente por Siigo
presenta deuda, saldo a favor, vencido y por vencer por proveedor; discrimina
documentos/cuotas/vencimientos, permite ajustes y cruces, y carga saldos
iniciales por documento, proveedor, fecha de vencimiento y moneda. Es una
referencia de comportamiento, no una copia de su implementacion ni de su UX.

1. **Descubrimiento y decision de reutilizacion.** Inventariar y probar
   `empresa_contabilidad_cartera_cxp`, `empresa_cuentas_por_pagar`,
   `empresa_proveedores`, `empresa_compras_documentos`, los soportes de compra,
   los reportes de edades y sus handlers/UI. Redactar una ADR que decida una
   sola fuente de verdad para saldo y movimientos; prohibido crear una segunda
   cartera paralela. Mapear importacion de saldos, compras existentes y puente
   contable antes de toda migracion.
2. **Modelo y reglas inmutables.** Para cada documento conservar empresa,
   proveedor validado dentro de esa empresa, origen, consecutivo externo,
   moneda/tasa, fecha de emision/vencimiento, cuotas, impuestos/retenciones,
   importe original, saldo, terminos de credito y estado. Estados minimos:
   borrador, aprobado, parcialmente_pagado, pagado, vencido, en_disputa y
   anulado. Un documento aprobado conserva historial; una anulacion o nota de
   credito revierte por movimiento trazable, nunca borra una deuda pagada.
3. **Movimientos y concurrencia.** Modelar abonos, anticipos, descuentos,
   notas credito/debito y cruces como asignaciones con identificador de
   idempotencia, actor, fecha y evidencia. En una transaccion PostgreSQL con
   bloqueo acotado, impedir que la suma asignada exceda el saldo salvo ajuste
   autorizado; actualizar saldo y auditoria atomicos. No enviar pagos bancarios
   ni integrar una pasarela en este bloque: registrar/proponer pago no equivale
   a ordenar dinero.
4. **Contabilidad, importacion y conciliacion.** Disenar saldos iniciales por
   proveedor-documento-cuota-vencimiento, incluyendo anticipos/saldos a favor y
   fecha de corte; validar totales antes de confirmar y conservar referencia al
   comprobante/puente contable sin duplicar asientos. Implementar conciliacion
   de proveedor que muestre documentos y creditos aplicables, con reverso
   auditable y sin edicion destructiva de un cruce cerrado.
5. **Permisos y seguridad.** Separar permisos por empresa: compras registra y
   recibe; contabilidad aprueba/ajusta/concili­a; tesoreria propone o registra
   pago; administrador financiero aprueba excepciones. Todos los endpoints
   derivan `empresa_id` de sesion/contexto, validan proveedor y documento en la
   misma empresa, usan CSRF/autorizacion/auditoria y no aceptan el tenant del
   cliente como autoridad. Aplicar el checklist de seguridad multiempresa y
   pruebas de enumeracion IDOR, cache, exportacion y jobs.
6. **UX y reportes.** Integrar en Administrar empresa, no como pagina global:
   tablero de deuda, credito a favor, vencido/por vencer y proximos pagos;
   estado de cuenta por proveedor; detalle por documento/cuota; edad de cartera
   (corriente, 1-30, 31-60, 61-90 y >90); propuesta de pagos; ajustes y
   conciliacion. Filtrar por fecha de corte, proveedor, estado, centro de costo
   y moneda. Mostrar origen, saldo y validaciones visibles; nunca mostrar una
   accion bancaria como completada solo por registrar un abono.
7. **Migracion y compatibilidad.** Crear migraciones versionadas e idempotentes
   solo si la ADR demuestra que hacen falta. Preparar migrador con conteos,
   sumas por empresa y rollback probado; no ejecutar DDL en handlers/worker.
   Los datos historicos deben poder relacionarse con el nuevo detalle sin
   cambiar saldos ni generar doble contabilizacion.
8. **Pruebas obligatorias.** Unitarias e integracion PostgreSQL para saldos,
   vencimientos, cuotas, monedas, redondeo, creditos, anulaciones, idempotencia,
   sobreasignacion y rollback. Ejecutar A/B real con dos empresas para proveedor
   compartido en nombre/NIT, documentos, reportes, archivos y exportaciones;
   incluir carrera de dos abonos concurrentes. Validar visualmente con rol de
   compras, contabilidad y tesoreria, escritorio/movil y estados de error. En
   staging, reconciliar conteos/saldos antes y despues de migrar una copia
   anonimizada y restaurarla.
9. **Criterio de aceptacion.** Por cada empresa, la suma de movimientos y
   asignaciones coincide con saldo/reportes y contabilidad; no existe acceso A/B
   ni saldo negativo no autorizado; cada cambio tiene actor/evidencia y los
   reversos conservan historia. Se requiere aprobacion explicita antes de
   habilitar pagos reales, importar datos productivos o ejecutar una orden
   bancaria.

## 12. Matriz minima de regresion antes del GO

| Capa | Casos obligatorios |
|---|---|
| Autenticacion | login valido/invalido, cierre, expiracion, cookies, rate limit, cambio de empresa y sesion revocada |
| Multiempresa | A/B por query/path/header/body/ID, roles, licencia, archivos, cache, jobs, exports y auditoria |
| POS | busqueda, scanner, cantidades, stock, descuentos, impuestos, pagos mixtos, caja, cancelacion autorizada y concurrencia |
| Finanzas | duplicado, replay, timeout, conciliacion, comision, redondeo, moneda y cierre |
| CxP proveedores | A/B, documento/cuota/vencimiento, saldo inicial, anticipo, abono concurrente, sobreasignacion, ajuste/reverso, conciliacion, moneda, reporte de edades, exportacion y roles |
| DIAN | preflight, firma, envio, consulta final, rechazo, reintento idempotente y persistencia por empresa |
| Documentos | upload, MIME/tamano, edicion, descarga, URL firmada, borrado, cuota, A/B y restore |
| Worker | lease, retry, reinicio, vencimiento, dead letter, concurrencia y metrica |
| UI | escritorio/movil, teclado, zoom, contraste, consola, red, estados vacio/error/carga y permisos |
| Operacion | migracion, backup, restore, rollback, alertas, saturacion, disco lleno, DB/Redis/proveedor caido |
| Seguridad | TLS/cabeceras/CSP, CSRF, XSS, inyeccion, uploads, secretos, logs, puertos e imagenes |

Comandos base, ajustados a la fase:

```powershell
Set-Location D:\powerfulcontrolsystem
git status --short
node tools/ensure_bootstrap_inventory.mjs --check
node tools/runtime_ensure_inventory.mjs --check
node tools/tenant_route_inventory.mjs --check
node tools/openapi_inventory.mjs --check
node tools/migration_audit.mjs --strict
node tools/deploy_pipeline_contract.mjs
Set-Location backend
go test ./... -count=1
go vet ./...
```

En CI Linux compatible:

```sh
go test -race ./...
```

Antes de cualquier `rs`, leer `documentos/comandos_codex.md`, ejecutar el
preflight estricto completo, revisar diff/secretos, obtener autorizacion y
validar despues salud, logs, SHA/digests, migraciones y smoke real.

## 13. Compuerta final GO/NO-GO

GO para un alcance concreto solo si:

- [ ] SHA/tag/digests unicos y aprobados.
- [ ] CI completa, race, seguridad y regresion verdes.
- [ ] API/worker sin DDL; migraciones nuevas/upgrade aprobadas.
- [ ] Aislamiento A/B demostrado en SQL, archivos, cache, jobs y exports.
- [ ] Operaciones financieras/fiscales idempotentes y durables.
- [ ] Staging equivalente publicado y matriz E2E completa.
- [ ] Backup restaurado, RPO/RTO y rollback medidos.
- [ ] Capacidad/SLO y alertas demostrados.
- [ ] Proveedores habilitados con evidencia real reconciliada.
- [ ] Cero vulnerabilidades criticas/altas y cero hallazgos P0/P1 sin dueño.
- [ ] UX P0 desktop/movil sin errores de consola/red.
- [ ] Modulos fuera de alcance apagados de forma efectiva.
- [ ] Soporte, runbooks, responsables y ventana aprobados.
- [ ] Usuario autoriza explicitamente el piloto/despliegue.

NO-GO automatico ante:

- fuga entre empresas o autorizacion inconsistente;
- migracion no reproducible, DDL desde API/worker o checksum divergente;
- backup sin restauracion funcional;
- pago, factura, stock, caja o comision duplicable/inconsistente;
- proveedor habilitado sin evidencia real;
- vulnerabilidad critica/alta, secreto expuesto o imagen no trazable;
- SHA distinto del aprobado, checks fallidos o rollback no viable;
- error P0 de consola/red/flujo operativo;
- alertas, SLO o capacidad sin demostrar.

## 14. Registro de avance que Terra debe mantener

Actualizar esta tabla al terminar cada bloque, sin marcar evidencia futura:

| ID | Estado | SHA | Entorno | Pruebas/evidencia | Riesgo residual | Aprobador |
|---|---|---|---|---|---|---|
| P105-001 | En curso | `f8f388b7` | main/CI/VPS | PR #44/#45 fusionadas, CI verde; `rs` final termino con codigo 0, `/health` y `/ready` 200, backend healthy y frontend/worker recreados | artefacto no inmutable | - |
| P105-002 | En curso | - | decision | borrador `matriz_alcance_piloto_plan_105.md` | alcance abierto | usuario |
| P105-003 | En curso | pendiente de SHA | local | inventarios bootstrap, multiempresa y runtime vigentes; `migration_audit.mjs --strict` OK. Las 21 huellas regeneradas corresponden a canonicalizacion EOL, no a DDL; falta preflight/rs sobre SHA final | DDL heredado/trazabilidad | - |
| P105-004 | En curso | - | diseno/test/staging | wrapper 203/203; matriz A/B P0 definida; fixtures y ejecucion faltantes | fuga tenant | - |
| P105-005 | En curso | - | analisis/test/staging | inventario de 23 goroutines; contratos estaticos de pasarelas/webhooks idempotentes OK; outbox parcial | duplicados | - |
| P105-006 | En curso | - | analisis/test | 320 5xx y 49 JSON directos en la ultima linea base; lotes IA super y licencias/pagos con canarios estaticos OK; `go test ./...` pasa; falta regenerar inventario y cubrir pagos restantes/productos/voz/identidad/multiempresa | fuga tecnica | - |
| P105-007 | En curso | candidato staging `6e280d3` + hotfix no consolidado | staging/edge | TLS válido, upstream `18082`, `/health` y `/ready` externos 200; frontend visible; migración aplicada y login visual autenticado correcto. Falta matriz E2E completa y consolidación SHA/CI; no se declara GO | equivalencia/rollback | - |
| P105-008 | En curso | - | VPS/staging | backup real ejecutado en `/root/powerfulcontrolsystem/backups/vps-snapshots/20260722_041759` sin secretos de entorno; `pg_dumpall` y cinco volúmenes empaquetados, hashes y `gzip/tar` íntegros; restore drill PostgreSQL aislado completado con `rto_seconds=15`, `rpo_seconds=232`; réplica externa/nube y política de alertas aún pendientes | continuidad/backup | - |
| P105-009 | En curso | - | CI/staging | suite Go completa, `go vet ./...` y builds de migrador/worker pasan nuevamente tras redacciones P105-006; smoke de carga staging: 80 requests, error rate 0%, P95 821 ms; el host local confirma `CGO_ENABLED=0`, por lo que `go test -race` no puede ejecutarse aquí; falta resultado race Linux sobre SHA candidato y carga sostenida real | carreras/capacidad | - |
| P105-010 | En curso | - | local/VPS/staging | `/health` y `/ready` core/staging 200; volumen de logs staging corregido a usuario `pcs` y el backend ya arranca sin `permission denied`; dashboard visual staging muestra continuidad 0/100, 295 eventos críticos y 161 sesiones frente a umbral 50; consulta read-only de producción registra 2034 sesiones activas (1874 con fecha_fin no vacía), lo que exige reconciliar métrica/limpieza gobernada. No se ejecutaron botones mutantes | ceguera operativa/sesiones | - |
| P105-011 | En curso | - | proveedor/TLS | TLS valido y respuestas minimas 200 de OnlyOffice/Nextcloud confirmadas; faltan E2E de proveedor y reconciliacion autorizada. Runbook: `documentos/gobernanza_tecnica/runbooks/runbook_tls_staging_y_servicios_plan_105.md` | externos/reconciliacion | usuario |
| P105-012 | Bloqueado | `5bbd4e62` | producción/navegador real | login real empresa 12: shell y cuenta/quota cargan, pero iframe muestra "rechazó la conexión". Nextcloud 29 emite `frame-ancestors 'self'`/`SAMEORIGIN`; prueba de header Nginx fue revertida desde backup tras validar config | requiere politica soportada que preserve CSP nonce; no usar cabeceras duplicadas | usuario |
| P105-013 | En curso | pendiente de SHA | local/VPS | pruebas de alcance, cobertura, Trivy y loopback pasan; falta escaneo vivo que termine con Trivy `ok` y auditoria real del host Ubuntu separada | hardening/telemetría | - |
| P105-014 | En curso | pendiente de SHA | local/produccion/staging | origen Nextcloud exacto y estilos del modulo externos; retirada total de `unsafe-inline` aun pendiente | CSP/deriva | - |
| P105-015 | Pendiente | - | staging | storage externo no demostrado | archivos/replicas | - |
| P105-016 | En curso | - | local/CI/staging | auditoria estatica de hardening OK; VPS ya expone digests de imagenes desplegadas para inventario, pero referencias por tag, SBOM y escaneo firmado siguen pendientes | supply chain | - |
| P105-017 | Pendiente | - | piloto | depende de P0 | operacion real | usuario |
| P105-018 | En curso | - | analisis | inventario actualizado y primer corte seguro por archivo; refactor separado pendiente | mantenibilidad | - |
| P105-019 | En curso | - | docs | auditoria 2026-07-22 mantiene `warning` exclusivo por 218 secuencias sospechosas en `CHANGELOG.md`; estrategia por lote pendiente | legibilidad | - |
| P105-020 | Pendiente | - | decision | sin fuente movil | no reproducible | usuario |
| P105-021 | Pendiente | - | decision | perfil Python apagado | excepcion tecnica | usuario |
| P105-022 | Pendiente | - | produccion controlada | carrito QA abierto | reserva/stock | usuario |
| P105-023 | Pendiente | - | diseno/staging | alcance agregado: ADR de fuente de verdad CxP, inventario de tablas/flujos existentes y matriz A/B aun pendientes | fuga tenant, doble cartera o doble contabilizacion | usuario |

Estados permitidos: `Pendiente`, `En curso`, `Bloqueado`, `Aprobado`,
`No aplica con aprobacion`. Nunca usar "Aprobado" si falta evidencia externa.

## 15. Cierre de esta planificacion

El proyecto tiene una base tecnica amplia y validadores locales utiles, pero el
estado comprobado al corte no permite afirmar que esta listo para produccion
general. Los bloqueadores centrales son evidencia, no cantidad de codigo:
migraciones totalmente separadas del runtime, aislamiento A/B, consistencia
financiera/fiscal, staging equivalente, restauracion, race/carga, observabilidad
viva y proveedores reales.

En esta corrida el usuario autorizo expresamente la validacion real de la
empresa Powerful Control System y el despliegue controlado mediante `rs`.
Persisten fuera de alcance los cobros, facturas DIAN, borrados y cualquier
mutacion irreversible. El porcentaje de avance se actualiza solo con la
evidencia que sobreviva al `rs` y a las sondas post-despliegue.
