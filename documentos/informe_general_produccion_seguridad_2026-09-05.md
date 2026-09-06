# Revisión general de producción, seguridad y aislamiento multiempresa

> **Actualización posterior:** las reparaciones locales derivadas de H01-H12 y su evidencia están registradas en [cierre_reparaciones_produccion_seguridad_2026-09-05.md](cierre_reparaciones_produccion_seguridad_2026-09-05.md). Este documento conserva la fotografía original que originó el trabajo.

Fecha: 5 de septiembre de 2026. Proyecto: Powerful Control System.

**Dictamen: NO-GO para abrir todos los módulos a producción general.** Hay una base técnica aprovechable y numerosas pruebas aprobadas, pero también defectos comprobados en autorización de relaciones, integridad operativa, protección ante abuso y separación de privilegios. Deben repararse antes de ofrecer los módulos afectados. Los demás necesitan completar sus pruebas de aceptación sobre un candidato reproducible.

Esta conclusión no significa que se haya demostrado una intrusión, extracción de datos de otra empresa o ejecución remota de código. Tampoco significa que todos los módulos estén averiados. Distingue reparación necesaria, validación pendiente y capacidades que todavía no deben comercializarse como completas.

## 1. Alcance y límites de la revisión

Se revisaron los contextos del proyecto, mapa de módulos, decisiones, contratos de seguridad y operación, documentación de BD y permisos, registros HTTP, autorización central, SQL de módulos representativos, almacenamiento privado, autenticación, integraciones, colas, despliegue, CI y frontend. Se ejecutaron la batería general de Go, análisis de dependencias, auditorías estáticas, comprobación de sintaxis frontend y dos pruebas de caracterización aisladas de los defectos encontrados. Se consultó el dominio público mediante seis GET sin sesión, sin crear operaciones comerciales.

El inventario cubre **los 53 módulos de permisos** y las superficies transversales de Super Administrador, pagos/licencias, IA, API móvil e infraestructura. La revisión manual profundizó en las fronteras de seguridad y los módulos señalados en los hallazgos; **no equivale a leer manualmente cada línea, probar todos los botones ni realizar un pentest completo**. Las filas marcadas «inventario» requieren una revisión funcional específica antes de obtener un GO individual.

- HEAD observado: `73cc7dfda1cbb2aaa374e99dc889c4b18dda1077`.
- Se auditó el árbol de trabajo local, que ya contenía cambios sin integrar. La captura inicial registró 61 entradas de estado Git; una captura posterior registró 62, incluida documentación añadida concurrentemente por otro trabajo. No se asumió que este árbol fuese un candidato congelado.
- Los resultados de pruebas corresponden a su momento de ejecución. No certifican modificaciones concurrentes posteriores ni la equivalencia con el binario del VPS.
- No se modificó código funcional, configuración, dependencias, BD ni despliegue. Las pruebas adicionales usaron un overlay temporal y un capturador SQL de biblioteca estándar, sin conexión a una base real.
- No se ejecutaron cobros, facturas, envíos, GPIO, restauraciones, cargas agresivas, instalaciones ni pruebas sobre empresas ajenas.
- No se verificaron en esta revisión la configuración privada del VPS, el estado de MFA por cuenta, reglas de firewall/WAF, antivirus, copias remotas, restauración, ni sesiones empresariales autenticadas.

Las referencias históricas a planes y a pruebas DIAN sirven de antecedentes. El contexto general vigente retira el Plan 110 como hoja activa; el contexto específico todavía lo presenta como vigente. Este informe no hereda porcentajes ni pendientes antiguos sin contrastarlos.

## 2. Evidencia ejecutada

| Comprobación | Resultado observado | Interpretación correcta |
| --- | --- | --- |
| `go test ./... -json -count=1 -timeout 5m` desde backend | 1.287 pruebas principales aprobadas; 1.667 contando subpruebas; 0 fallos; 18 omitidas | Valida la batería disponible, no las integraciones omitidas. |
| `go vet ./...` | Aprobado | Sin diagnósticos de vet en esa ejecución. |
| `go mod verify` | Todos los módulos verificados | Integridad del caché; no ausencia de vulnerabilidades. |
| `govulncheck` v1.1.4, reconstruido con Go 1.26.6 | Dos avisos SSH con traza de llamada; un aviso OpenPGP solo a nivel de módulo | Hallazgo de seguridad positivo, detallado en H01. La salida JSON terminó con código 0: se analizó su contenido. |
| Auditoría estricta de seguridad existente | `ok` | Comprueba principalmente presencia de patrones; no demuestra eficacia integral. |
| Auditoría estricta de permisos/licencias | 53 módulos, 55 wrappers, sin módulos faltantes en sus comprobaciones | Cobertura estructural, no autorización de cada objeto relacionado. |
| Inventario actual de `/api/empresa/` | 197 registros, 197 con wrapper, 0 duplicados | Solo esa familia y registros reconocidos por la herramienta; no es el total de endpoints del sistema. |
| `tenant_route_inventory.mjs --check` | Falló: documento desactualizado | Se generó una copia actual en temporal, sin sobrescribir la matriz del repositorio. |
| Auditoría estricta de migraciones | `ok`; 55 entradas históricas sin cambios frente a la referencia local `origin/main` | No se aplicaron migraciones ni se cotejó el ledger productivo. |
| Auditorías de contrato de módulos críticos y matriz de roles | `ok` | Son contratos estáticos, no recorridos autenticados por rol. |
| Auditoría de calidad estructural | Aprobó su línea base de no regresión | Mantiene deuda importante: 469 llamadas DB sin contexto y 736 resultados ignorados explícitamente. No todos son fallos. |
| Sintaxis frontend | 89 archivos JS y 206 bloques ejecutables inline sin errores | Compilación sintáctica; no demuestra comportamiento, accesibilidad ni ausencia de XSS. |
| Inventario CSP | Falló su no regresión: Super 198 frente a 197 permitidos | 206 scripts inline sin protección, 152 bloques de estilos y 862 atributos style; 1.220 obstáculos a CSP estricta en 190 archivos. No son 1.220 vulnerabilidades explotables. |
| Detector de carreras | No ejecutable con el entorno actual: `-race requires cgo` | No hay resultado race local. Debe ejecutarse en Linux/CI con herramientas compatibles. |
| Caracterización adicional aislada | Dos pruebas aprobadas reproduciendo el comportamiento defectuoso | CRM acepta referencia/autor/estado del cliente; una petición rechazada consume cuota antes de validar acceso. No se probó explotación HTTP productiva. |

Las 18 omisiones comprenden 14 pruebas que requieren PostgreSQL o un esquema restaurado aislado —incluyen cartera, inventario, secretos/fuentes DIAN y outbox/CxP—, dos validaciones XSD oficiales y dos pruebas de enlaces simbólicos sin privilegios locales suficientes. **No deben contarse como aprobadas.**

Evidencia resumida preservada: [JSON de evidencia](informe_general_produccion_seguridad_2026-09-05.evidencia.json). Los logs de trabajo permanecen en `.gotmp/auditoria_general_20260905/`; son temporales y no sustituyen los artefactos de una liberación.

### Observación del dominio publicado

| Petición sin sesión | Resultado |
| --- | --- |
| `/login.html` | 200, CSP y `nosniff`, caché desactivada; HSTS ausente. |
| `/api/empresa/productos?empresa_id=12` | 401 JSON. |
| `/super/api/pagos/auditoria` | 401 JSON; el middleware puede rechazar antes de resolver la ruta. |
| `/metrics` | 404. |
| `/ready` | 200 JSON, `no-store`; esto solo demuestra respuesta de disponibilidad. |
| `/.env` | 401; no se obtuvo contenido privado. |

HSTS no apareció en ninguna de estas seis respuestas HTTPS. La CSP publicada del login conserva `unsafe-inline` y `connect-src 'self' https: wss:`. Estas observaciones son actuales pero no identifican el SHA del backend desplegado.

## 3. Reparaciones prioritarias con evidencia

P1 indica alta prioridad: reparar antes de habilitar la superficie afectada. P2 indica endurecimiento o corrección necesaria según el alcance y la exposición. Estas prioridades son de liberación; no son una puntuación CVSS ni una afirmación de explotación.

### H01 — P1 — Vulnerabilidades conocidas en SSH de Domótica

`backend/go.mod` incluye `golang.org/x/crypto v0.55.0`. El análisis con Go 1.26.6 y la base de vulnerabilidades de Go, cuya fecha informada es 2026-09-02, detectó:

| Aviso | CVE | Efecto | Versión corregida informada |
| --- | --- | --- | --- |
| GO-2026-6354 | CVE-2026-78662 | Un par SSH malicioso puede bloquear la conexión durante establecimiento de canales. | `v0.56.0` |
| GO-2026-6355 | CVE-2026-56855 | Mensajes de canal establecido pueden bloquear la conexión. | `v0.56.0` |

La traza llega a `dialDomoticaSSH` en [control_electrico_ssh_install.go:296](../backend/handlers/control_electrico_ssh_install.go#L296), desde el instalador empresarial. Hay validación de huella y límites temporales: el escenario requiere alcanzar un par SSH permitido/malicioso o comprometido durante esa operación. **No se demostró un ataque anónimo directo desde Internet ni RCE.**

Reparación: actualizar la dependencia existente a una versión corregida compatible, con la autorización requerida por AGENTS.md para cambiar `go.mod`, probar conexión/huella/cancelación y volver a escanear los binarios e imágenes Linux que se desplegarán. Mientras tanto, excluir la instalación SSH del alcance liberado si no se corrige.

El aviso GO-2026-5932 sobre OpenPGP apareció solo a nivel del módulo; no se observó importación ni llamada afectada en la traza. No se contabiliza como tercera vulnerabilidad alcanzable. Las dos entradas SSH y sus trazas están preservadas en el JSON de evidencia; el visor web no pudo abrirlas durante esta revisión.

### H02 — P1 — Falta validar la empresa de registros relacionados

El wrapper valida la empresa principal, pero varios CRUD genéricos aceptan `producto_id`, `bodega_id`, `lead_id`, `cliente_id`, `bom_id`, `transportista_id` o `documento_gestion_id` sin verificar su pertenencia. Las configuraciones `cfgLotesSeries`, `cfgCRMInteracciones`, BOM, logística y firmas no asignan `ValidatePayload`; solo la configuración CxP lo hace en ese bloque.

Evidencia: [modulos_faltantes.go:231](../backend/handlers/modulos_faltantes.go#L231), [validación opcional:5942](../backend/handlers/modulos_faltantes.go#L5942) y [INSERT genérico:1414](../backend/db/modulos_faltantes.go#L1414). Las rutas están registradas y protegidas por wrappers en ese mismo archivo. El esquema inspeccionado no aporta una FK compuesta empresa/objeto para esas relaciones.

La prueba aislada ejecutó el handler CRM con empresa validada 11 y referencias externas: el primer acceso SQL fue el INSERT, sin consulta previa de pertenencia. **Confirma la omisión de validación; no demuestra lectura de los datos del supuesto padre ni cambio de filas de otra empresa en PostgreSQL.** Permite referencias inconsistentes y deja una frontera insuficiente para usos posteriores de esas relaciones.

También hay altas sin validación obligatoria de padre en MRP, WMS y partidas de tesorería: [produccion_mrp.go:475](../backend/db/produccion_mrp.go#L475), [logistica_wms.go:324](../backend/db/logistica_wms.go#L324), [tesoreria_presupuesto.go:336](../backend/db/tesoreria_presupuesto.go#L336). En MRP la consulta opcional de receta ignora el error; consumos/calidad comprueban un ID positivo, no su pertenencia antes de insertar.

Reparación: validar cada referencia con `(empresa_id, id)`, dentro de la transacción de escritura; derivar los datos del padre en servidor y añadir restricciones compuestas donde corresponda, tras conciliar datos existentes. Probar A/B con BD real tanto en POST como en PUT, importaciones y jobs.

### H03 — P1 — Campos de autoría, aprobación y estado modificables por CRUD

Las listas genéricas incluyen `usuario_creador` y estados operativos. En vacaciones/RRHH incluyen `aprobado_por`, niveles e historial de aprobación y referencias de nómina; en devoluciones, identificadores de impacto contable. El handler únicamente asigna `usuario_creador` cuando el cliente lo deja vacío. La ruta genérica PUT conserva los campos permitidos del payload y no exige pasar por la acción de transición.

Evidencia: [configuración RRHH:268](../backend/handlers/modulos_faltantes.go#L268), [autoría en POST:5928](../backend/handlers/modulos_faltantes.go#L5928), [PUT:6006](../backend/handlers/modulos_faltantes.go#L6006), [fallback de máquina de estados:6053](../backend/handlers/modulos_faltantes.go#L6053). La prueba CRM confirmó la aceptación de `usuario_creador` arbitrario y `estado_interaccion=cerrada` por creación ordinaria.

Esto **no elude por sí solo el permiso de escritura del wrapper**, pero permite saltarse reglas más finas y falsificar campos de trazabilidad. La auditoría automática puede conservar el usuario real; eso no vuelve confiable la autoría del registro de negocio.

Reparación: separar DTO de alta/edición de acciones de aprobación, contabilización y transición; retirar campos derivados de la lista escribible; obtener actor de la sesión y estado inicial del servidor. Probar que un usuario con edición pero sin aprobación no pueda simularla por PUT directo.

### H04 — P1 — Cuota de empresa consumida antes de autorizar su acceso

[empresa_permisos.go:1414](../backend/handlers/empresa_permisos.go#L1414) llama al limitador antes de comprobar identidad y `snapshot.CanAccess` (líneas 1427–1450). Un administrador autenticado puede seleccionar una empresa a la que no pertenece; para ese tipo de principal, el rechazo de pertenencia sucede después del consumo de cuota. Los usuarios operativos ligados a una empresa tienen además una comprobación previa distinta.

La prueba local confirmó que una solicitud rechazada por el wrapper ya había incrementado la cuota. No incluyó el middleware exterior de autenticación, por lo que no demuestra explotación anónima productiva.

El contador es un mapa por proceso y no elimina claves de empresas antiguas: [empresa_permisos.go:1339](../backend/handlers/empresa_permisos.go#L1339). Por ello hay riesgos de denegación de servicio entre empresas, crecimiento de memoria con IDs variados y límites diferentes por réplica.

Reparación: aplicar límite previo por identidad/IP para tráfico rechazado, comprobar pertenencia antes de consumir la cuota empresarial, acotar/purgar contadores y coordinar límites entre réplicas. Verificar que el tráfico rechazado de A no altera la disponibilidad de B.

### H05 — P1 — El worker recibe credenciales del propietario PostgreSQL

La API y el worker conectan normalmente con el rol runtime restringido. Sin embargo, [docker-compose.platform.yml:335](../deploy/docker-compose.platform.yml#L335) entrega al worker `PGUSER` y `PGPASSWORD` del propietario para snapshots. Esto contradice la separación declarada de que solo el migrador posee la credencial administrativa.

No hay exposición pública demostrada del secreto. El problema es el alcance de una intrusión en el worker: tendría a su disposición una identidad de BD más privilegiada. Además, el sidecar de disco tiene Docker socket; su compromiso posee un impacto potencial sobre el host y necesita aislamiento especial.

Reparación: separar backups del worker de negocio, usar credenciales de respaldo con los privilegios imprescindibles y revisar redes/operaciones autorizadas del sidecar. Verificar la configuración efectiva de contenedores sin imprimir variables privadas. Conservar el rol DML restringido ya implementado en [runtime_db_role.go](../backend/db/runtime_db_role.go).

### H06 — P1 — MFA disponible, pero no obligatorio y con fallo abierto de configuración

[admin_totp_handlers.go:96](../backend/handlers/admin_totp_handlers.go#L96) devuelve desactivado si falla leer la configuración. La necesidad de OTP depende de habilitación global, inscripción de la cuenta y secreto presente; no hay allí una obligación incondicional por ser Super Administrador. [auth_admin_handlers.go:505](../backend/handlers/auth_admin_handlers.go#L505) usa esa decisión para permitir continuar el login.

Esto es una debilidad verificable del diseño. No se consultó si MFA está activo en las cuentas publicadas, ni se provocaron fallos en su BD.

Reparación: exigir MFA a cuentas privilegiadas, fallar cerrado si no se puede determinar la política, completar inscripción/recuperación y revisar también OAuth, revocación y cambios de privilegios. Prueba de aceptación: ningún canal de acceso privilegiado evita el segundo factor exigido.

### H07 — P1/P2 — Endurecimiento de navegador y HTTPS incompleto

La comprobación pública confirmó HSTS ausente, aunque está configurado en código/template. Revisar el proxy TLS realmente activo y la confianza en cabeceras reenviadas; no basta editar una plantilla que no se usa. Evidencia esperada: HSTS en respuestas normales, errores y estáticos del dominio correcto.

CSP conserva scripts inline y orígenes amplios en el login. El inventario actual supera la línea base de Super. Debe corregirse la regresión y reducir la superficie de scripts con archivos propios, hashes o nonces, manteniendo Google/pagos/editor bajo orígenes explícitos.

No se demostró una inyección XSS. Retirar `unsafe-inline` de golpe sin adaptar las páginas rompería funcionalidades. La sintaxis correcta de JS no acredita saneamiento de datos; faltan pruebas de XSS almacenado/reflejado, mensajes entre iframes y recursos externos en los flujos autenticados.

### H08 — P1 — MRP/WMS pueden dejar operaciones parciales o perder consistencia concurrente

En [CreateEmpresaProduccionOrden](../backend/db/produccion_mrp.go#L475) la orden se inserta antes de materializar consumos, sin transacción envolvente: un error posterior deja la orden creada aunque la operación devuelva error. Los consumos y calidad insertan y después actualizan costos/estado ignorando el resultado de ese segundo paso (líneas 590–653).

En [WMS](../backend/db/logistica_wms.go#L324) se guardan ítems/avances/despachos y se ignoran errores de recomputar estado y registrar eventos. Varias transiciones leen estado y después actualizan sin bloqueo ni condición sobre la versión anterior. El código de orden MRP usa `COUNT(*)+1` ([línea 1063](../backend/db/produccion_mrp.go#L1063)), que puede colisionar bajo concurrencia.

Reparación: transacción por operación, idempotencia persistente, consecutivo atómico, actualización condicionada/versionada y tratamiento obligatorio de fallos de efectos que forman parte del negocio. Probar doble envío, dos operadores y fallo deliberado entre pasos en una BD desechable. No extrapolar a estos módulos las garantías ya añadidas al cobro POS.

### H09 — P1 antes de escalar — Topología todavía ligada a una sola instalación

Compose usa `container_name` fijo para API y worker; `PCS_WORKER_ID` tiene un valor común predeterminado. El almacenamiento usa volúmenes del host y rutas de filesystem. `PCS_PRIVATE_STORAGE_MODE=shared|object` es una declaración de configuración que el guard comprueba; no acredita por sí misma almacenamiento entre servidores ni demuestra un adaptador Object Storage en los handlers que usan `os.Open`.

Reparación: definir topología concreta de réplicas, balanceo, IDs únicos de worker, almacenamiento realmente compartido y despliegue de imágenes por digest. Probar pérdida de una API/worker y continuidad de archivos, sesión, jobs e idempotencia. Mantener una réplica no resuelve los demás hallazgos, pero evita prometer una capacidad aún no validada.

### H10 — P1/P2 — Consultas sin cancelación y amplificación de conexiones

El auditor AST detectó 469 llamadas DB sin contexto. Los helpers de compatibilidad todavía sustituyen el contexto por `context.Background()` en rutas de lectura/escritura: [sql_compat.go:557](../backend/db/sql_compat.go#L557). El timeout HTTP no garantiza que esas consultas se interrumpan.

En el listado de recetas MRP se consulta cada grupo de componentes mientras se siguen recorriendo las filas principales ([produccion_mrp.go:444](../backend/db/produccion_mrp.go#L444)). Este patrón N+1 retiene una conexión y solicita otras; bajo muchas peticiones simultáneas puede agotar el pool. Hay además dashboards que ignoran errores y muestran cero o listas parciales como si fueran datos completos.

Los pools sí tienen límites: 24 conexiones por BD para API y 8 por BD para worker ([postgres_pool.go:60](../backend/db/postgres_pool.go#L60)). Con dos BD, cuatro API podrían reservar hasta 192 conexiones, más workers, migrador y operación. Es un cálculo de configuración, no una medición de carga.

Reparación: propagar `r.Context()`, fijar plazos SQL por operación, cancelar consultas, cargar relaciones por lotes, paginar y medir planes SQL con volumen representativo. Definir presupuesto agregado de conexiones antes de aumentar réplicas.

### H11 — P1 para trazabilidad crítica — Auditoría automática sin entrega durable

[auditoria_empresa.go:657](../backend/handlers/auditoria_empresa.go#L657) y [auditoria_modulos_especificos.go:34](../backend/handlers/auditoria_modulos_especificos.go#L34) lanzan goroutines sin una cola durable para registrar eventos. Si el proceso cae, esos eventos pueden perderse; ante BD lenta, muchas solicitudes pueden acumular trabajo. La existencia de exportación forense y cadena de hashes no garantiza la recepción de eventos que nunca se persistieron.

Reparación: persistir los eventos críticos en la transacción de negocio/outbox y entregar con reintento; acotar concurrencia de auditoría no crítica y observar pérdida/retraso. Distinguir movimientos financieros ya trazados transaccionalmente de esta auditoría automática auxiliar: no se afirma que todo el sistema carezca de trazabilidad.

### H12 — P1 como condición de liberación — Falta evidencia integral del candidato

Hay 18 pruebas omitidas, race no ejecutado, matriz de rutas desactualizada y árbol sin congelar. No se ejecutaron A/B reales completos, restauración del candidato, carga autenticada, validación visual por rol, escaneo de imágenes Linux ni verificación privada del VPS. Las auditorías de presencia de texto pueden aprobar aun cuando fallen H02/H03; no deben constituir por sí solas la autorización de producción.

Reparación: integrar cambios, fijar SHA/digests, ejecutar CI y staging equivalentes, hacer obligatorias las pruebas PostgreSQL/XSD requeridas por el alcance y adjuntar los resultados. Corregir documentación contradictoria; no aumentar líneas base simplemente para ocultar un fallo.

## 4. Separación por empresa: evaluación

**La separación está implementada de manera importante, pero todavía no puede considerarse cerrada de punta a punta.**

| Capa | Lo comprobado | Lo que falta |
| --- | --- | --- |
| Sesión y selección de empresa | Wrappers validan pertenencia, rol, licencia y consistencia entre query/cabecera/formulario/multipart; existen pruebas negativas. | A/B autenticado sobre todos los caminos, incluyendo administradores compartidos/delegados y sesiones operativas. |
| Registros principales | CRUD genérico filtra lecturas/cambios por `empresa_id` e ID; tablas/columnas dinámicas pasan por listas cerradas. | H02: relaciones secundarias, verificaciones de filas afectadas y estados protegidos. |
| Vida personal | SQL e índices usan `empresa_id + usuario_id`; idempotencia y archivos se vinculan al propietario. | Prueba real con dos usuarios de la misma empresa y otro tenant, incluidos recibos y exportes. |
| Archivos | Raíces privadas por empresa, referencias saneadas y descargas por handlers; rutas legacy sensibles bloqueadas en Nginx. | Migración efectiva de archivos, antivirus, cuota concurrente, backup/restore y symlinks en Linux. |
| Colas y outbox | Jobs llevan empresa, claves idempotentes y reclamo con `FOR UPDATE SKIP LOCKED`; dispatcher conserva empresa. | Pruebas omitidas de durabilidad, reintento y aislamiento; payload/efectos deben permanecer en el mismo tenant. |
| Privilegios de BD | Rol runtime sin superusuario, DDL o BYPASSRLS; migrador independiente. | H05. No se encontró una política general de RLS en el código inspeccionado: el aislamiento depende de aplicación/consultas. |
| Recursos compartidos | Límites por empresa y pools configurables. | H04, H09 y H10: una empresa no debe consumir la cuota o bloquear el trabajo de otra. |

No es obligatorio crear una base distinta por empresa ni migrar a microservicios. El monolito modular puede mantenerse. Conviene evaluar RLS como defensa adicional en tablas de mayor riesgo, con contexto transaccional y pruebas específicas; **no sustituye** comprobar referencias, archivos, permisos ni jobs.

La selección de estas comprobaciones se apoya en la [guía multiempresa de OWASP](https://cheatsheetseries.owasp.org/cheatsheets/Multi_Tenant_Security_Cheat_Sheet.html), que trata por separado autorización de tenant, aislamiento de datos, recursos, archivos y tareas asíncronas.

## 5. Informe de los 53 módulos

«Reparar» señala un defecto concreto de este informe. «Validar» indica pruebas/aceptación pendientes, sin atribuir un fallo funcional nuevo. «Inventario» indica cobertura de catálogo/rutas y batería general, pero sin revisión manual completa del módulo. **Ninguna fila declara por sí sola GO productivo**: todas dependen de cerrar seguridad transversal y validar el alcance publicado.

| Módulo de permisos | Evaluación | Trabajo para producción |
| --- | --- | --- |
| `seguridad` — usuarios, roles y empresa | Reparar | H04/H06; MFA privilegiado, cuota tras pertenencia, pruebas A/B y revocación para roles personalizados/delegados. |
| `ventas` — carrito, venta directa, estaciones y caja | Validar | Conciliar cobro, stock, turno y documento; probar cuatro cajas, doble pago, cancelación y recuperación. No se detectó en esta revisión un fallo nuevo específico del cobro. |
| `inventario` — productos, bodegas, lotes/series | Reparar | H02/H03 en lotes genéricos; validar kardex, transferencias, series y stock bajo concurrencia; ejecutar migraciones PostgreSQL omitidas. |
| `clientes` | Validar | A/B de clientes/contactos, duplicados concurrentes, importación, crédito y exportación de datos personales. |
| `compras` | Reparar / validar | Restringir campos derivados de devoluciones y referencias de proveedor (H02/H03); conciliar compra, inventario, cuenta por pagar y reversos. |
| `finanzas` — caja, CxC/CxP y cartera | Validar | Fuente canónica CxP ya existe y la histórica rechaza nuevas escrituras; terminar conciliación histórica, precisión monetaria, abonos/outbox y cierres sobre PostgreSQL real. No repetir como vigente la antigua afirmación de dos fuentes nuevas escribibles. |
| `bancos_pagos` | Inventario / validar | Conciliación, referencias bancarias, no duplicar movimientos, estados definitivos y separación por cuenta/empresa; homologación de proveedores cuando aplique. |
| `tesoreria_presupuesto` | Reparar | H02 en partidas/presupuesto; controlar saldos y aprobaciones derivados y validar escenarios frente a movimientos reales. |
| `cobranza` | Inventario / validar | Cuotas, mora, reversos, destinatarios por empresa y reintentos que no dupliquen mensajes; evidencia del canal real. |
| `contabilidad_colombia` | Validar | Asientos balanceados y únicos, periodos cerrados, reversos, compras/ventas/caja y exportación contable consistente. |
| `contabilidad_colombia_avanzada` | Validar / alcance parcial | Conciliar históricos CxP; certificar documento soporte desde su fuente dedicada y mantener cerradas familias sin adaptador. |
| `centros_costo` | Inventario / validar | A/B de centros y referencias en partidas/asientos; totales reconciliados, inactivación y permisos de edición. |
| `cierre_fiscal` | Inventario / validar | Cierre/reapertura autorizado, exclusión de escrituras concurrentes y consistencia de todos los módulos que escriben al periodo. |
| `activos_fijos_niif_fiscal` | Inventario / validar | Depreciación, baja, revaluación y asientos; evitar doble corrida y cambios de periodos cerrados. |
| `declaraciones_tributarias` | Inventario / validar | Reconciliar bases y formatos con contabilidad; validación especializada del alcance normativo, periodos e historial. No se emitió una certificación tributaria. |
| `facturacion` — Colombia | Validar / alcance parcial | Factura desde venta pagada y nota crédito total tienen implementación específica. Certificar el candidato, fuentes inmutables, artefactos, numeración concurrente, reintento y acuse oficial por empresa. |
| `facturacion_ecuador` | Inventario / validar proveedor | Definir operaciones realmente implementadas; validar firma, numeración, respuesta oficial, representaciones y credenciales del emisor antes de anunciar disponibilidad. |
| `facturacion_panama` | Inventario / validar proveedor | Misma compuerta por país/proveedor; no heredar el resultado colombiano. |
| `nomina_sueldos` | Validar / alcance parcial | Nómina mensual DIAN dedicada existe; ejecutar XSD/BD omitidos, conciliar empleados/pagos y certificar aceptación. Ajustes, habilitación automática y entrega dedicada requieren cierre de su alcance. |
| `horarios_trabajadores` | Inventario / validar | Solapamientos, cambio de zona/fecha, autorización del empleado y relación con asistencia/nómina. |
| `asistencia_empleados` | Inventario / validar | Marcación repetida/offline, suplantación de empleado, turnos nocturnos y exportación aislada. |
| `hoja_vida_operativa` | Inventario / validar | Propietario de persona/adjuntos, permisos de datos personales, historial y retención. |
| `crm_unificado` | Reparar | H02/H03 reproducidos en interacciones; proteger relaciones lead/cliente, actor, estados y transiciones. |
| `produccion_mrp` | Reparar | H02/H08/H10: padres, atomicidad, consecutivos, costos, calidad y consultas N+1. Retirar o cerrar BOM legado si no forma parte del producto publicado. |
| `logistica_wms` | Reparar | H02/H08: orden de cada ítem/despacho, avance atómico, eventos y estado consistentes; concurrencia en picking/packing. |
| `importaciones_costeo` | Inventario / validar | Reparto de costos, moneda/tasa, referencias de compras, cierre de importación y asiento/inventario sin duplicados. |
| `calidad_procesos` | Inventario / validar | Separación entre quien registra y aprueba, evidencia privada, estados inmutables y relación con MRP. |
| `gestion_documental` | Reparar / validar | H02/H03 en gestión/firmas genéricas; versiones, propietario, hash, acceso/descarga y transición sin confiar en el cliente. |
| `documentos_onlyoffice` | Validar | Tokens temporales por empresa/archivo/acción, callback, SSRF, caducidad, permisos reales y colaboración concurrente; verificar servicio publicado. |
| `contratos_obligaciones` | Inventario / validar | Referencias y adjuntos por empresa, vencimientos, aprobaciones y autenticidad de estados/firma. |
| `cumplimiento_kyc` | Inventario / validar | Acceso mínimo a documentos de identidad, estados de revisión, retención, descargas y auditoría. |
| `portal_contador` | Inventario / validar | Acceso delegado solo a empresas autorizadas, revocación inmediata, exportes y ausencia de permisos implícitos globales. |
| `portal_terceros_certificados` | Inventario / validar | Tokens/enlaces vencidos, identidad del tercero y certificados propios; evitar enumeración y descarga cruzada. |
| `venta_publica` | Validar | Cálculo del precio/stock en servidor, autorización por token de seguimiento, firma/replay de pagos, callbacks y conciliación venta-caja. |
| `domicilios` | Inventario / validar | Pedido y repartidor de la empresa, privacidad de ubicación, transición/entrega y conciliación con venta/pago. |
| `reservas_hotel` | Inventario / validar | Doble reserva, fechas solapadas, cambio de habitación/estación, tarifa y cobro concurrente. |
| `alquileres` | Inventario / validar | Disponibilidad simultánea, entrega/devolución, depósito, daño y documentos; aislamiento del bien y cliente. |
| `aiu_construccion` | Inventario / validar | Bases AIU, costos/obra/cliente, aprobaciones y documentos; validación contable del alcance ofertado. |
| `parqueadero` | Inventario / validar | Entrada/salida repetida, tarifa por tiempo, tiquete, pago y cierre; comportamiento con red intermitente. |
| `vehiculos_registro` | Inventario / validar | Vehículo/tercero/estación de la empresa, imágenes privadas y entradas/salidas consistentes. |
| `turnos_atencion` | Inventario / validar | Asignación concurrente, llamado único, privacidad en pantalla pública y recuperación de conexión. |
| `control_electrico` — Domótica/Raspberry | Reparar / hardware | H01; tokens, enrolamiento/revocación y ACK por dispositivo; además ensayo físico GPIO/relés/sensores y recuperación tras desconexión. |
| `energia_solar` | Inventario / hardware | Telemetría asociada al dispositivo/tenant, unidades, lecturas obsoletas y validación con equipos físicos. |
| `camaras` | Inventario / validar | URLs/credenciales privadas, SSRF y acceso al stream; probar revocación y pertenencia del equipo. |
| `ubicacion_gps` | Inventario / validar | Dispositivo/persona del tenant, acceso a históricos, retención, caducidad y consumo de consultas. |
| `carnets` | Inventario / validar | Identidad y fotografía correctas, generación/descarga privada y verificación de QR/token. |
| `chat_tareas` | Validar | Mensajes/adjuntos por empresa y participantes, desconexión, WebSocket, historial y cambios de permisos. |
| `reportes` | Validar / rendimiento | Reconciliar cifras con fuentes, archivos por empresa, exportes grandes, paginación, tiempos SQL y permisos efectivos. |
| `auditoria` | Reparar | H11; persistencia de eventos críticos, entrega durable, retención/exporte forense y restricciones para purgar. |
| `backups` | Validar / operación | Restauración A/B de datos/archivos, privilegios mínimos, copia externa cifrada y ensayo de recuperación aislado. No ejecutar restore real para probar. |
| `soportes_compras_ia` | Validar | Cuotas/archivos, idempotencia de radicación, borrador revisado antes de contabilizar y aislamiento; ejecutar prueba PostgreSQL omitida. |
| `vida` | Validar | Conservar doble aislamiento empresa/usuario y separación de contabilidad; comprobar recibos, precios, suscripciones y cámara móvil real. |
| `bolsa` | Inventario / validar | Exactitud/antigüedad de datos, permisos y ausencia de ejecución financiera no autorizada; definir alcance informativo frente a funciones reales. |

### Superficies transversales y comerciales

| Superficie | Trabajo específico |
| --- | --- |
| Super Administrador, delegaciones y selector | MFA obligatorio; cada ruta administrativa valida rol efectivo; pruebas de revocación y agregados limitados a empresas autorizadas. Auditar especialmente disco, DB admin, snapshots y soporte remoto. |
| Pagos y licencias ePayco/Wompi | Los cambios locales incluyen idempotencia de checkout/efectos y auditoría privada. Falta congelar e integrar ese candidato y probar pago aprobado/rechazado/pendiente, firma inválida, callback duplicado/fuera de orden, monto/moneda/empresa incorrectos y renovación concurrente. Confirmar producción de cada proveedor sin asumir la antigua configuración de sandbox. |
| Bre-B, QR y datáfonos | Mostrar/generar un QR no demuestra recepción de fondos. Definir validación bancaria/homologación, conciliación, reversos y evidencia física del terminal. |
| IA empresarial, selector y Super | Herramientas cerradas, propuestas y separación de contexto tienen base y tests. Hacen falta evaluaciones adversariales por ámbito, fuga en adjuntos/historial, confirmación independiente, costo concurrente y evidencia del proveedor. Un modelo nuevo no cambia los permisos del sistema. |
| Rappi | El código actual rechaza webhook sin secreto/firma válida y comprueba antigüedad; corregir la documentación que dice que la firma es opcional. Validar tienda/empresa, replay y mapeo de venta/inventario con el proveedor. |
| Mailu, WhatsApp y notificaciones | Credenciales/plantillas por alcance correcto; demostrar entrega, reintento sin duplicado, rebote, cuotas y alarmas operativas. No inferir entrega por HTTP 200 del backend. |
| Nextcloud, OnlyOffice y soporte remoto | Verificar aprovisionamiento, cuota, revocación, aislamiento de archivos/tokens y exposición de servicios internos. No basta que la página abra. |
| Impresión | Validar cada formato con importes reales, permisos, reimpresión y estado de cola; confirmar papel físico. «Reclamado por agente» no equivale a impreso. |
| PWA, modo offline y API móvil | Aislamiento de caché/IndexedDB al cambiar usuario/empresa, sincronización idempotente, conflicto de inventario y actualización de service worker. El cliente nativo requiere fuente/build/aceptación propios. |
| Portal público, shells y experiencia móvil | Corregir CSP/HSTS, probar sesión caducada, iframes, responsive, errores visibles y accesibilidad en la aplicación autenticada. La comprobación de sintaxis es solo una parte. |
| CI, despliegue, PostgreSQL y recuperación | Resolver H05/H09/H10/H12, escanear imágenes, cerrar SHA/digests, migración/rollback, restauración y capacidad. Revisar los resultados efectivos de CI de ese candidato. |

## 6. Preparación ante ataques

PCS ya tiene defensas concretas: cookies HttpOnly/SameSite, CSRF, revocación, límites HTTP, permisos por empresa, restricciones de rutas públicas, cifrado de secretos, controles SSRF probados, descarga privada y wrappers de licencia. Hay que conservarlas y probar su efectividad conjunta.

Para poder afirmar una preparación razonable deben completarse estas verificaciones, sin prometer invulnerabilidad:

| Amenaza | Condición de aceptación |
| --- | --- |
| Robo de contraseña o sesión privilegiada | MFA exigido, sesiones revocadas, recuperación segura, alertas y controles de acceso a herramientas administrativas. |
| Acceso a otra empresa/objeto | Matriz A/B HTTP + PostgreSQL + archivos + jobs con usuarios no globales; cero lectura, mutación o filtración de metadatos ajenos. |
| Alteración de negocio | Rechazo de campos reservados; precios, estados, actor y aprobaciones derivados en servidor; transacciones e idempotencia. |
| XSS/CSRF | CSP adaptada, pruebas de entrada almacenada/reflejada y ausencia de bypass CSRF por métodos o rutas auxiliares. |
| SSRF/archivos maliciosos | Probar redirecciones, DNS/IP privados, rutas, tamaño expandido de archivos, symlinks, cuotas y antivirus de los flujos habilitados. |
| DDoS y abuso autenticado | Límites en borde y aplicación, cuota posterior a autorización, límites de jobs/consultas/archivos y prueba de vecino ruidoso. La protección volumétrica requiere capacidad externa al proceso Go. |
| Compromiso de worker/sidecar | Ninguna credencial administrativa en procesos de negocio; redes y privilegios mínimos; auditoría de operaciones de infraestructura. |
| Dependencias vulnerables | H01 corregido, análisis de símbolos y escaneo de las imágenes exactas, con gestión de avisos nuevos. |
| Borrado/ransomware | Copia independiente, protegida frente al mismo administrador comprometido, restore aislado probado y secretos recuperables por un procedimiento controlado. |
| Incidente activo | Responsable, canal de aviso probado, procedimiento para revocar sesiones/tokens, preservar evidencias, contener y recuperar con tiempos medidos. |

Se propone usar [OWASP ASVS](https://owasp.org/www-project-application-security-verification-standard/) como matriz de requisitos verificables para la siguiente fase. No se afirma que PCS esté certificado contra ese estándar. El pentest debe ejecutarse después de corregir los defectos conocidos, sobre staging equivalente y con alcance definido.

## 7. Escalabilidad y aceptación operativa

La arquitectura Go/PostgreSQL con API, migrador, worker, cola y outbox puede sostener un crecimiento ordenado. **No hay evidencia en esta revisión para prometer un número concreto de empresas, cajas o peticiones por segundo.** Se requiere medir con los datos, índices, hardware y mezcla de operaciones que se usarán.

El ensayo debe incluir login y panel, catálogo, cobro, informes/exportes, documentos, adjuntos, IA y túnel de dispositivos. Medir p50/p95/p99, errores/timeouts, conexiones activas y esperas, bloqueos SQL, CPU/memoria/disco y edad de colas. Separar una carga normal de la de un cliente que consume su máximo permitido.

Criterios: cero cobros/documentos duplicados, cero mezcla A/B, cero operaciones parciales, recuperación al perder una réplica y funcionamiento degradado controlado cuando falla un proveedor. Los objetivos de latencia, concurrencia, recuperación (RTO) y pérdida máxima de datos (RPO) deben acordarse y cumplirse con mediciones. No sustituirlos por `/health` o `/ready` en 200.

## 8. Orden de trabajo para entrar en producción

1. **Fijar alcance y candidato.** Integrar los cambios locales, resolver conflictos, elegir SHA/digests y definir exactamente qué módulos se ofrecerán. Los que no se certifiquen quedan fuera del lanzamiento.
2. **Cerrar seguridad e integridad conocidas.** H01–H06; H02/H03 en todas las familias genéricas alcanzadas; H08 en MRP/WMS y H11 para eventos críticos. Verificar HTTPS y resolver la regresión CSP.
3. **Ejecutar las pruebas que faltan.** PostgreSQL aislado/restaurado, negativos A/B, errores intermedios, concurrencia, race y XSD del alcance fiscal. Exigir que las comprobaciones críticas no se omitan silenciosamente en CI.
4. **Certificar operaciones de negocio.** POS/caja/inventario/CxP primero; después pagos/licencias/DIAN de cada empresa y familia, con hechos reales autorizados y evidencia oficial. La implementación de soporte/nómina no implica aceptación productiva automática.
5. **Validar experiencia y equipos.** Navegación por rol, móvil/PWA, impresiones, dispositivos y errores de proveedor. No crear datos fiscales ficticios para esta fase.
6. **Certificar operación y capacidad.** Imágenes exactas, privilegios efectivos, firewall/borde, alertas entregadas, copia externa/restore, rollback y carga. Resolver H09/H10 antes de varias réplicas.
7. **Emitir GO del alcance concreto.** Adjuntar resultados del mismo candidato, responsables y límites de operación. Un piloto reducido podría aprobarse antes que todos los verticales; necesita igualmente seguridad transversal, aislamiento y continuidad comprobados.

La primera reparación recomendada es la frontera de CRUD y sus relaciones, seguida de SSH/MFA/cuotas/privilegios del worker. Después, cerrar integridad de MRP/WMS y completar la certificación de los flujos comerciales prioritarios. No se recomienda una reescritura general ni agregar infraestructura para ocultar estos defectos.
