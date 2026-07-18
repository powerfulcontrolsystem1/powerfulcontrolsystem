# Plan 104 - Cierre verificable para produccion

Fecha de auditoria: 2026-07-16

Estado del plan: endurecimiento interno, restauracion tecnica y staging privado
parcialmente comprobados; proveedores externos y validacion funcional aislada
siguen abiertos.

Rama auditada: `codex/rs-20260716-225734`

Commit auditado: `22bb4479b9f1eb83d5ac511643879b00782be704`

## 1. Objetivo

Cerrar de forma comprobable las brechas que aun impiden declarar Powerful
Control System listo para produccion general. El plan prioriza integridad
financiera y fiscal, aislamiento multiempresa, recuperacion, seguridad,
operacion reproducible y evidencia real. No agrega funciones comerciales por
conveniencia.

Este documento no autoriza despliegues, cambios de datos reales ni el uso de
credenciales productivas. Cada fase debe generar evidencia revisable antes de
avanzar.

## 1.1 Avance interno 2026-07-16

- El rol `migrate` puede omitir explicitamente el bootstrap heredado mediante
  `PCS_RUNTIME_SCHEMA_BOOTSTRAP=0`; API y worker no pueden habilitarlo en
  produccion. Esto permite ensayar el catalogo inmutable sin cambiar el
  comportamiento compatible de instalaciones existentes.
- Los flujos HTTP de modulos ERP, conversion de cotizaciones/pedidos y pagos
  ya no invocan DDL para sus tablas transaccionales. Verifican que
  `pcs-migrate` haya preparado el esquema y fallan cerrados con un mensaje de
  disponibilidad, sin detalles SQL.
- El catalogo legado de 121 pasos tiene ahora una huella SHA-256 por paso,
  generada desde el codigo y comprobada por preflight. La migracion inmutable
  `20260717-001-legacy-schema-manifest-v1` registra la huella global sin
  alterar los checksums historicos ya aplicados.
- La matriz generada de rutas empresariales inventaria 203 endpoints
  `/api/empresa/`, todos con wrapper autoritativo detectado. El preflight
  bloquea que la matriz quede desactualizada; falta aun demostrar SQL, archivos,
  cache y jobs mediante la suite negativa A/B.
- La validacion central de tenant rechaza ahora valores repetidos o ambiguos de
  `empresa_id` en query, cabecera, formulario y multipart; estas fuentes no
  pueden elegir un tenant distinto del ya autorizado.
- El inventario de runtime registra 153 llamadas `Ensure*`: 80 desde handlers
  HTTP y 72 durante arranque. Los flujos de cobro de licencia y sincronizacion
  DIAN, carrito y corte de caja ya verifican configuracion avanzada,
  facturacion, ventas y finanzas sin emitir DDL. Esta deuda queda protegida
  contra crecimiento por preflight, pero debe extraerse por dominio antes de
  apagar el bootstrap para todos los entornos.
- La CSP activa y `Report-Only` nacen de la misma fuente y usan orígenes
  explicitos. Se eliminaron las fuentes genericas `https:`/`wss:` de
  `connect-src` y `https:` de `img-src`; la transicion de scripts inline a
  nonces sigue pendiente de inventario y validacion visual en staging.
- El middleware JSON redacciona detalles tecnicos que lleguen por respuestas
  4xx sin eliminar los errores de validacion seguros que necesita el frontend.
- El perfil opcional `voice` queda desactivado en la plantilla de plataforma.
  Su implementacion Python sigue fuera del alcance del piloto hasta una
  excepcion aprobada o reemplazo en Go.

Estos cambios reducen riesgo, pero **no cierran** P104-002, P104-004 a
P104-007, P104-009, P104-011 a P104-014 ni autorizan ejecutar `rs`.

## 1.2 Evidencia adicional 2026-07-18

- La candidata de staging corrigió el orden del migrador: prepara el ledger
  `schema_migrations` antes de consultar el baseline heredado. En staging se
  ejecutó `pcs-migrate` repetidamente con `PCS_RUNTIME_SCHEMA_BOOTSTRAP=0` sin
  re-aplicar cambios.
- El frontend de staging se corrigió para publicar el puerto interno `8080` de
  Nginx, coherente con la imagen no privilegiada. El host mantiene una reserva
  de puertos Docker que impidió el smoke publicado; se validó el mismo frontend
  dentro de la red aislada sin reiniciar el daemon compartido con produccion.
- Se detectó y corrigió un worker degradado: `METRICS_INTERVAL_SECONDS` ya no
  acepta valores menores de cinco segundos. El worker candidato terminó en
  estado `healthy` y la prueba cubre la programacion registrada.
- El rollback de aplicacion al SHA previo y el retorno posterior a la
  candidata respondieron `200` en `/health` y `/ready` desde la red privada de
  staging. La evidencia detallada vive en
  `documentos/staging_execution_report.md`.

Esto deja P104-002 parcialmente cerrado solo en su componente tecnico. No
demuestra datos anonimizados, smoke por navegador publicado, ni operaciones
funcionales A/B completas; por tanto no autoriza produccion ni `rs`.

## 2. Veredicto actual

**NO-GO para produccion general.**

El proyecto tiene una base tecnica considerable y las pruebas Go actuales
pasan, pero todavia no existe evidencia suficiente para garantizar una salida
general segura. El sistema puede prepararse primero para un piloto controlado
de una sola replica, siempre que se cierren todos los bloqueadores P0 y P1
definidos en este plan.

Estado estimado por alcance, basado solamente en la evidencia disponible:

| Alcance | Preparacion estimada | Veredicto |
| --- | ---: | --- |
| Piloto web, una replica y modulos esenciales | 72% | Requiere cerrar P0/P1 y staging |
| Produccion web general, todos los modulos | 52% | No autorizada |
| Varias replicas | 42% | No autorizada |
| Aplicacion movil publica | 35% | No autorizada; falta fuente versionada |

Los porcentajes son una medida de avance del plan, no una certificacion.

## 3. Evidencia positiva verificada

- El arbol de trabajo estaba limpio al iniciar la auditoria.
- `go test ./...` paso en `backend/`.
- `go vet ./...` paso en `backend/`.
- `git diff --check` paso.
- Existen 549 archivos Go, 181 archivos de prueba Go y 627 funciones de prueba.
- La API expone 328 registros de rutas HTTP y el inventario OpenAPI generado
  reporta 327 rutas.
- El CI incluye pruebas, detector de carreras, `govulncheck`, `gosec`,
  Gitleaks, Trivy, SBOM, licencias y validacion de Docker Compose.
- Las acciones de GitHub usadas por CI estan fijadas por commit.
- API, migrador y worker tienen roles separados.
- El migrador dispone de ledger, checksum, advisory lock y registro de
  ejecuciones.
- La API y el worker de produccion bloquean DDL de runtime y verifican
  migraciones requeridas.
- El worker dispone de leases, reintentos, DLQ, recuperacion y healthcheck.
- Los procesos API, migrador y worker tienen imagenes no root, filesystem
  `read_only`, `no-new-privileges` y `cap_drop`.
- Los pools PostgreSQL y los timeouts HTTP tienen configuracion por rol.
- CSRF, cookies seguras, origen estricto, rate limit de login, 2FA y
  revocacion de sesiones cuentan con controles y pruebas.
- Los documentos privados nuevos tienen una ruta fuera de la raiz publica y
  existen controles de traversal y symlink en areas recientes.
- Wompi y ePayco tienen contratos estaticos de creacion, consulta, retorno e
  idempotencia de webhook.
- Existe una fachada `/api/v1` con JSON uniforme, paginacion e idempotencia
  inicial para operaciones moviles.

## 4. Limites de la evidencia actual

- `documentos/staging_execution_report.md` dice expresamente
  `NO EJECUTADO EN ESTA ESTACION`.
- El preflight mas reciente fue rapido: omitio pruebas Go completas y Docker.
- La auditoria automatica de seguridad verifica patrones, no demuestra que
  cada handler, consulta, archivo y job preserve el tenant.
- El reporte de pagos verifica contratos de codigo, no una transaccion sandbox
  reciente aceptada por cada proveedor.
- Los reportes de observabilidad prueban existencia de componentes, no que
  alertas, dashboards, retencion y respuesta a incidentes funcionen en
  runtime.
- La evidencia de carga localizada es antigua y no cubre el commit auditado.
- No hay evidencia reciente de restauracion completa, rollback ni RPO/RTO
  medidos.
- La sintaxis JavaScript no pudo repetirse en esta estacion porque `node` no
  estaba disponible en el `PATH`; debe quedar cubierta por CI.

## 5. Bloqueadores de produccion

### P0 - Bloqueadores absolutos

#### P104-001 - Fuente unica de liberacion

Hallazgo:

- La rama auditada es `codex/rs-20260716-225734`.
- `main` local apunta a `f622e8b1`.
- `origin/main` apunta a `b9b5313d`.
- La rama auditada contiene trabajo posterior que no esta integrado en la
  referencia remota principal.

Riesgo:

- Construir, probar y desplegar revisiones diferentes.
- Perder correcciones o publicar una rama sin la proteccion esperada.

Aceptacion:

- Elegir una unica revision candidata.
- Integrarla mediante PR protegida.
- Exigir CI verde sobre el SHA exacto.
- Crear tag de release y manifiesto con SHA, digests y esquema esperado.
- Prohibir `rs` o sincronizacion si el SHA local no coincide con la revision
  aprobada.

#### P104-002 - Staging real y reproducible

Hallazgo:

- El stack de staging existe, pero no hay ejecucion actual documentada.

Aceptacion:

- Levantar staging aislado con una copia anonimizada.
- Ejecutar `pcs-migrate` dos veces.
- Iniciar API y worker con `PCS_RUNTIME_SCHEMA_BOOTSTRAP=0`.
- Verificar `/health`, `/ready` y health del worker.
- Ejecutar smoke desktop y movil con al menos dos empresas y roles distintos.
- Ensayar despliegue del digest candidato y rollback al digest anterior.
- Registrar SHA, digests, fechas, resultados y responsable sin secretos.

#### P104-003 - Migraciones inmutables completas

Hallazgo:

- El inventario registra 156 funciones `Ensure*` y 124 archivos que tocan
  esquema.
- El baseline heredado ejecuta funciones Go mutables bajo una huella general.
  Cambiar el cuerpo de una funcion `Ensure*` no cambia necesariamente el
  checksum de la migracion ya aplicada.
- El proceso `migrate` conserva un bootstrap historico amplio con seeds,
  normalizaciones y aprovisionamientos.

Riesgo:

- Deriva silenciosa de esquema.
- Entornos con estructuras diferentes bajo la misma version.
- Migraciones dificiles de revertir o repetir.

Aceptacion:

- Congelar el baseline heredado con una huella verificable por paso.
- Convertir toda modificacion futura a migraciones inmutables, ordenadas y
  transaccionales.
- Separar DDL, migracion de datos, seeds y aprovisionamiento externo.
- El migrador no debe contactar Mailu, Nextcloud, DIAN ni otros proveedores.
- La API y el worker deben iniciar sin ejecutar ni intentar DDL.
- Probar base vacia, base actualizada, segunda aplicacion y deriva de checksum.

#### P104-004 - Restauracion y continuidad comprobadas

Hallazgo:

- Hay scripts y runbook de backup, pero no una restauracion reciente completa
  con mediciones.

Aceptacion:

- Restaurar PostgreSQL, archivos privados, uploads permitidos y configuracion
  en un entorno desechable.
- Verificar login, empresas, licencias, productos, venta, caja, factura,
  adjuntos y auditoria.
- Medir RPO y RTO reales.
- Probar backup externo y recuperacion cuando el VPS principal no exista.
- Registrar checksum de respaldos y evidencia de restauracion.

#### P104-005 - Aislamiento multiempresa demostrable

Hallazgo:

- Existen wrappers de permisos y contexto de empresa, pero la autoridad de
  tenant esta distribuida.
- Muchos handlers siguen leyendo `empresa_id` de query o JSON y realizan
  comprobaciones locales.
- No existe evidencia de una auditoria completa por las 328 rutas, jobs,
  caches, archivos y exportaciones.
- PostgreSQL RLS no esta implementado en dominios criticos.

Aceptacion:

- Crear un registro de endpoints con permiso, fuente autoritativa del tenant,
  tablas, archivos, jobs y exportaciones.
- Centralizar un `TenantContext` obligatorio y evitar que repositorios
  empresariales acepten un ID no validado.
- Añadir pruebas A/B para lectura, alta, cambio, borrado, exportacion,
  descarga, webhook, cache y job.
- Probar IDs secuenciales de otra empresa en URL, JSON y cabeceras.
- Evaluar y aplicar RLS focal en ventas, pagos, facturas, documentos y
  archivos, sin usarlo como sustituto de la autorizacion de aplicacion.
- Cero acceso cruzado en la suite negativa.

#### P104-006 - Consistencia financiera, fiscal e idempotencia

Hallazgo:

- El outbox transaccional tiene un unico productor de negocio:
  `commerce.sale-paid`.
- Permanecen 19 goroutines iniciadas desde handlers o paquetes de negocio.
- Pagos, DIAN, correo, documentos, inventario y otros efectos externos no
  estan cubiertos de forma uniforme por outbox.

Riesgo:

- Venta cobrada sin inventario o contabilidad.
- Pago, factura o notificacion duplicada.
- Trabajo perdido cuando el proceso se reinicia.

Aceptacion:

- Inventariar todas las operaciones mutantes y su llave de idempotencia.
- Hacer atomicos carrito, venta, pago, caja, inventario y documento.
- Emitir eventos outbox dentro de la misma transaccion.
- Llevar DIAN, pagos, correo, WhatsApp, documentos y notificaciones al worker.
- Eliminar goroutines no durables de handlers criticos.
- Probar doble clic, timeout, reintento, webhook repetido, reinicio y
  concurrencia.
- Ninguna repeticion legitima puede duplicar dinero, stock o documento fiscal.

#### P104-007 - Proveedores externos con evidencia autorizada

Hallazgo:

- Existen contratos de codigo, pero no evidencia reciente integral del commit
  candidato.

Aceptacion:

- Wompi y ePayco: pago aprobado, rechazado, pendiente, expirado, retorno,
  webhook repetido, monto alterado y activacion unica de licencia.
- DIAN: emision autorizada, rechazo controlado, nota credito, correo al
  cliente, QR y consulta posterior.
- Bre-B: dejarlo manual/conciliado hasta tener confirmacion bancaria firmada;
  no marcar pago automatico por mostrar un QR.
- Correo/Mailu: entrega, rebote, SPF, DKIM, DMARC, PTR, TLS y logo remitente.
- WhatsApp: plantillas, consentimiento, rechazo, rate limit y costo.
- Nextcloud: alta, cuota, acceso por empresa, revocacion y borrado.
- OnlyOffice: JWT, apertura, guardado, callback y aislamiento de archivo.
- RustDesk/WebRTC: autorizacion, token corto, origen, revocacion y tenant.
- OpenAI: limites, proveedor propio por empresa, redaccion, timeout y costos.
- Cada proveedor debe tener circuit breaker operativo, timeout, reintento
  idempotente y alerta.

### P1 - Altos antes del piloto

#### P104-008 - Errores API sin filtracion interna

Hallazgo:

- Se detectaron aproximadamente 1261 llamadas que pueden devolver
  `err.Error()` mediante `http.Error`.
- El middleware sanitiza errores 5xx, pero mensajes 4xx pueden exponer nombres
  de tabla, validaciones internas o detalles de proveedor.

Aceptacion:

- Crear taxonomia de errores de dominio con codigo publico estable.
- Conservar causa interna y request ID solo en logs protegidos.
- Prohibir `err.Error()` directo en respuestas.
- Añadir una prueba estatica de CI y regresiones para SQL, archivos,
  proveedores y secretos.

#### P104-009 - Archivos privados y Object Storage

Hallazgo:

- Persisten volumenes locales para uploads, descargas, backups y documentos.
- Nginx bloquea varias rutas sensibles, pero monta el volumen completo de
  uploads y la clasificacion depende de listas de rutas.
- No existe un adaptador Object Storage terminado.

Aceptacion:

- Catalogar cada tipo de archivo como publico, privado o temporal.
- Mover todo archivo privado fuera del web root.
- Servir privados por handler autenticado o URL firmada corta.
- Implementar interfaz de storage con backend local para desarrollo, MinIO en
  staging y S3/R2 compatible en produccion.
- Guardar tenant, hash, MIME, tamano, retencion y propietario.
- Migracion reanudable con checksum, sin symlinks ni sobrescritura.
- Object Storage es obligatorio antes de varias replicas y adjuntos moviles.

#### P104-010 - CSP efectiva

Hallazgo:

- La politica activa conserva `unsafe-inline`, `img-src https:` y
  `connect-src https:`/`wss:` genericos.
- La politica mas estricta sigue en modo Report-Only.

Aceptacion:

- Consolidar una unica fuente de CSP para Go y Nginx.
- Inventariar scripts inline y migrarlos por lotes a archivos o nonces.
- Reemplazar origenes genericos por listas cerradas.
- Agregar endpoint y almacenamiento de reportes CSP sin datos sensibles.
- Promover la politica estricta de Report-Only a enforced despues de un
  periodo sin violaciones inesperadas.

#### P104-011 - Hardening Docker completo

Hallazgo:

- El override de release fija por digest solo API, migrador y worker.
- PostgreSQL, Mailu, OnlyOffice, RustDesk, edge y monitoreo usan tags.
- Varios servicios no tienen `read_only`, `cap_drop`, limites de CPU/RAM,
  `pids_limit` ni healthcheck.
- cAdvisor es privilegiado y monta recursos del host.

Aceptacion:

- Fijar por digest cada imagen del release o documentar excepcion temporal.
- Definir limites de memoria, CPU y procesos por servicio.
- Aplicar usuario no root y hardening donde el proveedor lo soporte.
- Aislar redes y no publicar puertos internos.
- Restringir cAdvisor o sustituirlo por una opcion con menor privilegio.
- Escanear todos los perfiles realmente habilitados, no solo API/frontend.
- Registrar excepciones de OnlyOffice, Mailu o RustDesk y su mitigacion.

#### P104-012 - Observabilidad operativa real

Aceptacion:

- Metricas de latencia, errores, conexiones DB, jobs, outbox, DLQ, DIAN,
  pagos, correo, storage y consumo IA.
- Alertas con umbrales, deduplicacion y canal probado.
- Dashboard por disponibilidad, negocio y capacidad.
- Ejecutar simulacros: DB caida, proveedor lento, job muerto, disco lleno,
  storage no disponible y certificado por vencer.
- Medir que cada alerta llega y enlaza a un runbook.

#### P104-013 - Rendimiento y concurrencia

Hallazgo:

- Existen archivos de 13.444, 6.898 y 6.675 lineas, handlers con SQL y
  dominios muy acoplados.
- No hay planes de consulta recientes con un dataset anonimo representativo.

Aceptacion:

- Medir POS, carrito, caja, productos, inventario, reportes, facturacion y
  panel con `EXPLAIN (ANALYZE, BUFFERS)`.
- Corregir N+1, indices y cargas completas en memoria.
- Probar stock, caja y carrito con concurrencia.
- Ejecutar carga sostenida y picos con presupuesto de conexiones.
- Definir SLO y capacidad de una replica antes de escalar.

#### P104-014 - Alcance modular de lanzamiento

Aceptacion:

- Clasificar cada modulo como `estable`, `piloto`, `experimental` o
  `deshabilitado`.
- Un modulo deshabilitado no debe mostrar pagina, aceptar ruta, programar job
  ni consumir proveedor.
- El primer piloto debe limitarse a modulos con pruebas completas.
- Los verticales no probados deben quedar apagados por empresa.

### P2 - Necesarios para sostenibilidad

#### P104-015 - Modularizacion progresiva

- Dividir `modulos_faltantes.go`, pagos, reportes, carrito, productos,
  permisos y facturacion por dominio.
- Sacar SQL y reglas complejas de handlers.
- Introducir servicios/repositorios solo donde reduzcan acoplamiento real.
- Mantener compatibilidad de rutas mientras se agregan pruebas de contrato.
- Prohibir nuevos archivos de dominio excesivos y nuevas funciones duplicadas.

#### P104-016 - Documentacion vigente y codificacion

Hallazgo:

- La auditoria documental reporta 343 secuencias sospechosas en cinco archivos
  principales.
- Documentos anteriores conservan estados ya superados y otros aun pendientes.

Aceptacion:

- Corregir UTF-8 sin reemplazos destructivos.
- Mantener una sola matriz de readiness vigente.
- Marcar documentos historicos como tales.
- Actualizar contexto, arquitectura, modulos, flujos, BD, OpenAPI, runbooks y
  changelog con evidencia del SHA candidato.

#### P104-017 - Servicio de voz

Hallazgo:

- `voice-stream` esta implementado en Python dentro del sistema, mientras las
  reglas vigentes del repositorio requieren Go para runtime.

Aceptacion:

- Mantener el perfil `voice` apagado para el piloto, o
- reimplementar el servicio en Go, o
- obtener una excepcion tecnica explicita, documentada y aprobada.

## 6. Aplicacion movil

Hallazgo critico independiente:

- `mobile/powerful_control_system_app` contiene solamente `.dart_tool` y
  `build`.
- No hay `pubspec.yaml`, `lib/`, `test/`, Android ni iOS versionados en la rama
  auditada.
- `mobile/` aparece completamente ignorado y el APK encontrado es de debug.

Consecuencia:

- No existe una fuente reproducible desde la cual construir, revisar, firmar y
  publicar Android o iOS.

Plan movil obligatorio:

1. Recuperar o reconstruir la fuente Flutter desde una rama revisada.
2. Versionar `pubspec.yaml`, `lib/`, `test/`, Android, iOS y configuracion de
   CI; ignorar unicamente `.dart_tool`, `build` y secretos.
3. Implementar sesiones por dispositivo, refresh rotativo, cierre remoto,
   PKCE, almacenamiento seguro y revocacion inmediata.
4. Reemplazar cursores basados en offset por cursores estables por ID/fecha.
5. Completar sync incremental, conflictos y borrados.
6. Implementar push FCM/APNs mediante outbox y preferencias por empresa.
7. Probar offline, reintentos, doble envio, cambio de empresa y perdida de red.
8. Generar builds release reproducibles y firmados en CI.
9. Ejecutar pruebas en Android e iOS reales antes de publicar.

La aplicacion movil no bloquea un piloto web si queda fuera del alcance
publicado, pero si bloquea cualquier anuncio de disponibilidad movil.

## 7. Fases de ejecucion

### Fase 0 - Congelar candidato y gobernanza

Incluye P104-001 y P104-014.

Salida:

- SHA unico, PR aprobada, CI verde, tag, manifiesto y catalogo de modulos.

### Fase 1 - Esquema determinista

Incluye P104-003.

Salida:

- Migraciones inmutables completas, API/worker sin DDL y base vacia/actual
  probadas dos veces.

### Fase 2 - Tenant, errores y seguridad web

Incluye P104-005, P104-008 y P104-010.

Salida:

- Suite A/B completa, errores publicos estables y CSP estricta operativa.

### Fase 3 - Integridad de negocio

Incluye P104-006 y P104-013.

Salida:

- Matriz de mutaciones, transacciones, outbox, idempotencia y concurrencia
  verificada.

### Fase 4 - Archivos y recuperacion

Incluye P104-004 y P104-009.

Salida:

- Storage compartido, backup externo y restauracion con RPO/RTO medidos.

### Fase 5 - Plataforma operable

Incluye P104-011, P104-012 y P104-017.

Salida:

- Digests, limites, hardening, alertas, dashboards y excepciones aprobadas.

### Fase 6 - Staging integral

Incluye P104-002.

Salida:

- Release y rollback del SHA candidato con pruebas por rol, tenant, viewport,
  impresion, carga y fallos.

### Fase 7 - Proveedores

Incluye P104-007.

Salida:

- Matriz firmada de proveedores con evidencia sandbox o autorizada.

### Fase 8 - Sostenibilidad y movil

Incluye P104-015, P104-016 y el plan movil.

Salida:

- Fuente movil reproducible, documentacion vigente y deuda modular gobernada.

### Fase 9 - Cutover controlado

Solo se ejecuta con todas las compuertas obligatorias aprobadas.

1. Congelar cambios.
2. Confirmar backup y restauracion reciente.
3. Confirmar digests y migraciones.
4. Desplegar primero migrador, despues API/worker/frontend.
5. Ejecutar smoke de solo lectura.
6. Habilitar trafico gradual.
7. Vigilar SLO, pagos, DIAN, jobs, errores y conexiones.
8. Revertir por digest ante cualquier criterio de aborto.
9. Abrir modulos por feature flag, nunca todos simultaneamente.

## 8. Matriz minima de pruebas

| Area | Pruebas obligatorias |
| --- | --- |
| Autenticacion | login, logout, reset, invitacion, 2FA, revocacion y cache |
| Multiempresa | A no lee, cambia, borra, exporta ni descarga B |
| Permisos | super, administrador, cajero, vendedor, contador y soporte |
| Ventas | efectivo, combinado, credito, descuento, anulacion y devolucion |
| Caja | apertura, concurrencia, cierre, diferencia, impresion y logout |
| Inventario | negativo permitido, reservas, bodegas, traslados y concurrencia |
| Facturacion | venta simple, FE, rechazo, reintento, nota y correo/QR |
| Licencias | cantidad 1-5, descuento, pago repetido, activacion y vencimiento |
| Pagos | aprobado, rechazado, pendiente, expirado, webhook repetido y monto falso |
| Archivos | MIME falso, traversal, symlink, tamano, tenant y descarga sin permiso |
| Worker | lease, retry, DLQ, reinicio, idempotencia y cierre gradual |
| Integraciones | timeout, credencial invalida, proveedor caido y circuit breaker |
| UI | desktop, movil, claro/oscuro, formularios, botones e impresion |
| Carga | sostenida, pico, consulta lenta, pool agotado y job pesado |
| Recuperacion | backup, restore, rollback, DNS/VPS nuevo y perdida de storage |

## 9. Compuertas GO/NO-GO

La decision solo puede ser GO si:

- [ ] El SHA candidato esta integrado en `origin/main` y tiene CI verde.
- [ ] `go test ./...`, `go test -race ./...`, `go vet ./...`,
      `govulncheck`, `gosec`, Gitleaks, Trivy, SBOM y licencias pasan.
- [ ] Docker Compose de produccion y staging valida todos los perfiles activos.
- [ ] Las migraciones se aplican dos veces y la API/worker no ejecutan DDL.
- [ ] La suite multiempresa A/B pasa para rutas, archivos, jobs y exportaciones.
- [ ] No hay respuestas API que filtren errores internos.
- [ ] Venta, pago, caja, inventario y factura son atomicos e idempotentes.
- [ ] El outbox cubre todos los efectos externos del alcance publicado.
- [ ] Backup y restauracion completos cumplen RPO/RTO.
- [ ] Release y rollback por digest fueron ensayados.
- [ ] Los proveedores activos tienen evidencia autorizada.
- [ ] Las alertas y runbooks fueron probados.
- [ ] Los modulos no validados estan apagados.
- [ ] Existe aprobacion humana de negocio, seguridad y operacion.

NO-GO automatico si:

- cambia el SHA despues de las pruebas;
- falla una migracion o aparece deriva;
- existe acceso cruzado entre empresas;
- una repeticion duplica dinero, stock o documento;
- no se puede restaurar el backup;
- un proveedor critico no tiene modo degradado;
- faltan secretos o digests requeridos;
- una alerta critica no llega;
- el rollback no funciona.

## 10. Orden recomendado de implementacion

1. P104-001 Fuente unica de liberacion.
2. P104-003 Migraciones inmutables.
3. P104-005 Aislamiento multiempresa.
4. P104-008 Errores API.
5. P104-006 Consistencia, outbox e idempotencia.
6. P104-009 Archivos y Object Storage.
7. P104-004 Restauracion.
8. P104-011 Docker y supply chain.
9. P104-010 CSP.
10. P104-012 Observabilidad.
11. P104-013 Rendimiento y concurrencia.
12. P104-014 Alcance modular.
13. P104-002 Staging integral.
14. P104-007 Proveedores.
15. P104-015/P104-016 Sostenibilidad y documentacion.
16. Fuente y contrato movil.
17. Cutover gradual.

## 11. Definicion final de terminado

Plan 104 se considera completado solo cuando:

- cada P104 tiene estado, responsable, SHA y enlace a evidencia;
- no quedan P0 ni P1 abiertos;
- el release candidate es reproducible desde un checkout limpio;
- una restauracion completa y un rollback fueron ejecutados;
- la matriz de tenant, permisos, pagos, fiscal y archivos esta en verde;
- el alcance publicado coincide con los modulos realmente habilitados;
- la documentacion describe el runtime desplegado, no un estado historico;
- el acta GO esta firmada por negocio, seguridad y operacion.

Hasta entonces, la declaracion correcta sigue siendo:

**plataforma avanzada en preproduccion, no produccion general aprobada.**
