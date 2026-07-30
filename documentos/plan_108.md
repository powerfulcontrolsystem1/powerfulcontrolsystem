# Plan 108 - cierre integral y certificación para producción

Fecha de corte local: 2026-07-25
Estado inicial: **NO-GO**
Modelo ejecutor previsto: **GPT-5.6 Terra con razonamiento medio**
Repositorio auditado: `D:\powerfulcontrolsystem`
Commit base observado: `3dd48d243527bf51581d66c9b396029a0dbd43bb`

## 1. Propósito

El Plan 108 es la hoja de ruta única para llevar Powerful Control System (PCS)
desde el estado local actual hasta una decisión verificable de entrada en
producción. Consolida, sin declarar terminados de forma artificial, los
pendientes de:

- Plan 106: certificación integral, cartera de proveedores, interfaz,
  impresiones, IA, cajas simultáneas, proveedores y ensayo general.
- Plan 107: contador profesional, consistencia contable, estados financieros,
  impuestos, reportes dinámicos y UAT contable.
- Plan IA: privacidad por usuario, permisos efectivos, memoria, herramientas,
  auditoría, evaluaciones y despliegue gradual.
- Planes 104 y 105 y `plan_final_para_produccion.md`: migraciones inmutables,
  runtime sin DDL, aislamiento multiempresa, worker/outbox, seguridad,
  almacenamiento, observabilidad, continuidad, capacidad y lanzamiento.

Desde la aprobación de este documento, los planes anteriores quedan como
antecedentes y fuentes de requisitos. El avance operativo se registra solamente
contra identificadores `P108-*` para impedir dobles conteos o porcentajes
inflados.

Este documento es un plan. Su creación no autoriza despliegues, emisiones
fiscales, cobros, mensajes externos, migraciones sobre datos reales ni apertura
de producción.

## 2. Veredicto de la auditoría

### 2.1 Evidencia positiva actual

El preflight completo ejecutado el 2026-07-25 finalizó correctamente y dejó el
reporte:

`documentos/reportes_profesionales/preflight_20260725_225755.md`

La evidencia local aprobó, entre otros controles:

- compilación y pruebas Go completas;
- sintaxis JavaScript y PowerShell;
- contratos de módulos, roles, permisos y licencias;
- contrato OpenAPI y controles de seguridad estáticos;
- migraciones y observabilidad a nivel de validadores;
- contratos de pagos, soporte, staging y SLO/SLA;
- auditoría UX estática y validaciones de documentación;
- validación de Docker Compose.

También existe evidencia previa de restauración aislada y smoke público de
staging. Esa evidencia es útil, pero no certifica el candidato actual ni
reemplaza pruebas autenticadas, sostenidas y transaccionales.

### 2.2 Razones del NO-GO

PCS no está certificado para producción general por estas razones P0:

1. El árbol de trabajo contiene cambios mezclados y archivos nuevos de los
   planes 106, 107 e IA; no existe todavía un candidato limpio, inmutable,
   etiquetado y reproducible.
2. La certificación productiva del Plan 106 continúa en cero: hay pruebas
   locales parciales, pero no una pasada integral sobre el mismo artefacto en
   staging equivalente.
3. Cartera de proveedores conserva dos modelos históricos que deben
   reconciliarse contra una fuente canónica antes de aceptar saldos reales.
4. La migración definitiva de CxP y varias pruebas PostgreSQL de atomicidad,
   idempotencia y concurrencia no están demostradas en staging.
5. El Plan 107 tiene base local útil, pero todavía no demuestra el ciclo
   contable completo, los 46 reportes observados, impuestos, estados
   financieros, cierres, exógena ni UAT independiente.
6. El Plan IA tiene cambios locales no desplegados; la versión publicada
   observada todavía conserva la experiencia antigua. Faltan evaluación de
   herramientas, seguridad, auditoría, costo, latencia y permisos por rol.
7. El runtime conserva una deuda amplia de creación/verificación de esquema.
   El inventario previo identificó 154 funciones `Ensure*` y 122 pasos
   históricos. Debe probarse que API y worker no ejecutan DDL en producción.
8. No existe una suite negativa completa que demuestre aislamiento entre dos
   empresas en SQL, archivos, caché, trabajos, exportaciones, reportes e IA.
9. Faltan pruebas E2E visuales y funcionales de todas las acciones, documentos
   imprimibles y botones IA, con consola y red limpias.
10. Faltan pruebas sostenidas y autenticadas de varias cajas concurrentes,
    incluyendo inventario, consecutivos, caja, pagos, cartera y contabilidad.
11. El incidente de firma DIAN requiere corrección operacional segura y una
    aceptación oficial nueva. Una respuesta HTTP exitosa no sustituye
    `GetStatusZip` con estado aceptado.
12. Falta evidencia actual del candidato para pagos, correo, WhatsApp,
    Nextcloud/OnlyOffice, OpenAI y cualquier integración incluida en el piloto.
13. El almacenamiento privado compartido y su recuperación no están cerrados
    para un escenario con más de una réplica.
14. No se ha ejecutado un ensayo general completo, rollback incluido, con el
    mismo artefacto que se pretende liberar.
15. La normalización documental terminó con advertencia: `CHANGELOG.md`
    conserva 218 secuencias sospechosas que deben revisarse.

## 3. Principios obligatorios de ejecución

GPT-5.6 Terra medio debe aplicar estas reglas en cada fase:

La estructura sigue la guía oficial de modelos GPT-5.6: entregar contexto de
dominio, restricciones duras, límites de autorización y criterios de éxito
explícitos, conservando el nivel de razonamiento solicitado y validándolo con
evaluaciones:
`https://developers.openai.com/api/docs/guides/model-guidance?model=gpt-5.6-terra`.

1. Leer primero `AGENTS.md`, `documentos/contexto_general_del_sistema.md`, este
   plan y la documentación específica indicada por la fase.
2. Revisar `git status`, commit, rama, diferencias y archivos concurrentes antes
   de editar. Nunca borrar ni sobrescribir trabajo ajeno.
3. Ejecutar una sola fase acotada a la vez. No mezclar una corrección P0 con
   refactorizaciones P2 no necesarias.
4. Mantener Go y PostgreSQL. No agregar dependencias ni modificar `go.mod` sin
   autorización expresa y trazabilidad.
5. Derivar en servidor `empresa_id`, usuario, rol, permisos y licencia. Nunca
   confiar en identificadores empresariales enviados por el navegador o por un
   modelo IA.
6. No permitir SQL libre, HTTP arbitrario, secretos, selección de tenant ni
   confirmación de escrituras dentro de prompts o herramientas IA.
7. Toda mutación financiera debe ser transaccional, idempotente, auditable y
   segura frente a concurrencia.
8. No ejecutar DDL desde API o worker en el perfil de producción. El esquema se
   cambia mediante el migrador versionado.
9. No registrar contraseñas, tokens, claves privadas, certificados, datos
   sensibles completos ni rutas privadas en código, documentos, capturas o
   reportes.
10. La cuenta de prueba autorizada para Powerful Control System debe obtenerse
    del canal seguro ya proporcionado por el usuario. Su contraseña no debe
    escribirse en el repositorio ni en evidencias.
11. Las pruebas con efectos reales se limitan a la empresa autorizada y al
    entorno aprobado. Antes de enviar DIAN, cobrar, enviar correo/WhatsApp,
    confirmar una acción IA o alterar datos reales, verificar alcance,
    reversibilidad y autorización vigente.
12. Un timeout es resultado inconcluso, no aprobado. Un test estático no
    sustituye una prueba autenticada. Staging no equivale a producción.
13. Cada fase termina con código, pruebas, evidencia y documentación alineados.
14. Si aparece un P0 nuevo, detener el avance dependiente, registrarlo y
    corregirlo antes de continuar.

### 3.1 Formato obligatorio de cada ciclo de Terra

Antes de trabajar:

- objetivo exacto e identificador `P108-*`;
- archivos y módulos previstos;
- datos, permisos, tenant y efectos externos involucrados;
- criterios de aceptación que se van a demostrar;
- estado del árbol Git y evidencia base.

Al terminar:

- causa o necesidad atendida;
- archivos modificados;
- pruebas ejecutadas con resultado;
- evidencia guardada;
- riesgos restantes;
- rollback o forma de deshacer;
- estado: `pendiente`, `en curso`, `bloqueado`, `aprobado` o `fallido`.

Terra no debe marcar una fase `aprobada` si falta una evidencia obligatoria.

## 4. Alcance de lanzamiento

Antes de corregir indiscriminadamente todo el sistema se debe definir el
alcance comercial del primer lanzamiento:

- módulos incluidos en el piloto;
- módulos visibles pero desactivados;
- países, monedas y reglas fiscales habilitadas;
- proveedores externos requeridos;
- roles autorizados;
- número de empresas, usuarios y cajas esperados;
- objetivos de disponibilidad, latencia, RPO y RTO;
- soporte, horario de piloto y responsables de decisión.

Todo módulo visible para el usuario se considera parte del alcance funcional y
debe probarse. Si un módulo no estará listo, debe quedar oculto o bloqueado en
servidor por permiso/licencia/configuración, con una explicación visible; no
basta con no probarlo.

## 5. Fases de ejecución

### P108-000 - Gobierno, inventario y candidato limpio [P0]

**Objetivo:** convertir el estado mezclado actual en una línea base controlada.

**Acciones:**

1. Inventariar cambios locales y atribuirlos a Plan 106, 107, IA u otra tarea.
2. Resolver archivos duplicados, generados, temporales o sin propietario.
3. Revisar los cambios por backend/BD, frontend/UX y QA/operación.
4. Confirmar qué cambios pertenecen al candidato y cuáles se difieren.
5. Ejecutar pruebas enfocadas antes de integrar cada bloque.
6. Crear un commit candidato limpio en una rama `codex/*`, sin secretos.
7. Registrar commit, árbol limpio, fecha, imagen, digest, migraciones y
   configuración no secreta.
8. Construir una sola vez y promover el mismo artefacto entre ambientes.

**Aceptación:**

- árbol limpio;
- revisión de diferencias terminada;
- commit y digest inequívocos;
- artefactos reproducibles;
- manifiesto de alcance y configuración;
- ninguna credencial versionada.

**Evidencia:** `evidencia_plan_108/P108-000/manifest.md`, SHA, digests, salida de
build, SBOM y estado Git.

### P108-001 - CI, calidad y cadena de suministro [P0]

**Objetivo:** demostrar que el candidato compila y pasa controles repetibles en
el entorno objetivo Linux.

**Acciones:**

1. Ejecutar preflight completo desde el commit candidato.
2. Ejecutar `go test ./... -count=1`, pruebas enfocadas y `go vet ./...`.
3. Ejecutar `go test -race` con CGO habilitado en Linux para paquetes de
   concurrencia, caja, ventas, pagos, CxP, jobs e IA.
4. Ejecutar escaneo de vulnerabilidades Go, secretos, SAST y dependencias.
5. Analizar imágenes finales con Trivy y registrar imagen/digest escaneado.
6. Generar SBOM y revisar licencias.
7. Fijar imágenes por versión o digest, usuario sin privilegios, filesystem y
   capacidades mínimas.
8. Corregir o aceptar formalmente cada hallazgo con dueño y vencimiento. Ningún
   crítico o alto explotable puede quedar abierto para el piloto.

**Aceptación:** CI verde sobre el SHA candidato, race verde, cero secretos,
cero vulnerabilidades críticas explotables y evidencia asociada al digest.

### P108-002 - Migraciones inmutables y runtime sin DDL [P0]

**Objetivo:** que el esquema sea determinista y no dependa del arranque de API
o worker.

**Acciones:**

1. Generar el inventario canónico de todos los `Ensure*`, `CREATE`, `ALTER`,
   índices, triggers, extensiones y datos semilla.
2. Clasificar cada operación: migración, comprobación de compatibilidad o
   bootstrap histórico.
3. Crear migraciones versionadas, ordenadas, transaccionales cuando aplique y
   con checksum.
4. Incluir la migración canónica de CxP y las migraciones pendientes de IA y
   reportes.
5. Probar base vacía, base actualizada desde snapshot y segunda aplicación sin
   cambios.
6. Desactivar `PCS_RUNTIME_SCHEMA_BOOTSTRAP` en el perfil candidato.
7. Fallar rápido si el esquema requerido no existe; no intentar repararlo desde
   API/worker.
8. Verificar permisos PostgreSQL: migrador con DDL; API y worker sin DDL.

**Aceptación:** una base vacía y una restaurada llegan al mismo esquema; segunda
aplicación es vacía; API/worker arrancan sin DDL y fallan de forma segura ante
esquema incompatible.

### P108-003 - Worker, jobs, outbox e idempotencia transversal [P0]

**Objetivo:** eliminar efectos duplicados y trabajos frágiles dentro del proceso
web.

**Acciones:**

1. Inventariar goroutines, timers, cron, colas y reintentos del API.
2. Trasladar al worker los trabajos de notificación, DIAN, pagos, documentos,
   reportes y limpieza que deban sobrevivir reinicios.
3. Conectar transactional outbox a las mutaciones de negocio críticas.
4. Definir clave idempotente, estados, intentos, backoff, timeout, lease,
   dead-letter y reanudación.
5. Probar doble entrega, worker duplicado, caída antes/después de commit,
   reinicio y recuperación.
6. Exponer métricas por cola, edad, reintentos, fallos y dead-letter.

**Aceptación:** ninguna repetición duplica venta, pago, documento, asiento,
inventario, mensaje o movimiento de cartera.

### P108-004 - Aislamiento multiempresa y autorización efectiva [P0]

**Objetivo:** demostrar que una empresa o rol no puede observar ni mutar datos
ajenos.

**Acciones:**

1. Construir matriz de endpoints, tablas, archivos, cachés, jobs,
   exportaciones, reportes, integraciones e IA.
2. Probar empresa A/B con IDs válidos ajenos, parámetros manipulados, rutas,
   paginación, filtros, descargas y referencias indirectas.
3. Verificar roles `super_administrador`, `admin_empresa`, `contador`, `cajero`
   y roles operativos relevantes.
4. Verificar permisos efectivos y licencia en servidor para GET y mutaciones.
5. Probar CSRF, IDOR, sesión expirada, revocación, cambio de empresa, carrera de
   permisos y acceso directo por URL.
6. Revisar consultas sin filtro empresarial, joins incompletos y claves de caché
   sin tenant.
7. Revisar URLs privadas y jobs para que el tenant se derive de un contexto
   confiable.

**Aceptación:** toda prueba negativa devuelve denegación segura o conjunto vacío
sin filtrar existencia, PII, rutas ni detalles internos.

### P108-005 - Cartera de proveedores canónica [P0]

**Objetivo:** cerrar completamente P106-003 a P106-009.

**Acciones:**

1. Confirmar mediante ADR `empresa_cuentas_por_pagar` como fuente canónica o
   documentar otra decisión antes de tocar datos.
2. Comparar la fuente canónica con
   `empresa_contabilidad_cartera_cxp`, compras, documentos y asientos.
3. Crear conciliador de solo lectura que reporte faltantes, duplicados y
   diferencias sin corregir silenciosamente.
4. Definir invariantes: moneda/unidad, saldo no negativo, fecha, vencimiento,
   estado, proveedor, documento, empresa y referencia externa.
5. Migrar con respaldo, métricas antes/después y rollback ensayado.
6. Implementar movimientos inmutables para causación, abono, ajuste,
   devolución, nota y anulación.
7. Hacer atómico cada abono con caja/banco, asiento, saldo y auditoría.
8. Aplicar idempotencia y bloqueo seguro a pagos concurrentes.
9. Integrar compras, devoluciones, contabilidad e importación IA sin duplicar.
10. Probar tenant A/B, permisos, doble clic, reintento, pago simultáneo y
    vencimientos.

**Aceptación:** un mismo hecho económico aparece una sola vez, el saldo se
reconstruye desde movimientos, contabilidad concilia y no existen pagos dobles
ni cruces de empresa.

### P108-006 - Carga de facturas y recibos de proveedor con IA [P0/P1]

**Objetivo:** completar el flujo solicitado de carga documental asistida.

**Acciones:**

1. Permitir PDF e imágenes con límites de tipo, tamaño, páginas y frecuencia.
2. Guardar el archivo en almacenamiento privado, con nombre opaco, hash,
   empresa, usuario, antivirus/validación y política de retención.
3. Extraer mediante contrato estructurado: proveedor, identificación, número,
   fecha, vencimiento, moneda, subtotal, impuestos, retenciones, total,
   conceptos y confianza por campo.
4. Nunca aceptar la lectura IA como asiento o CxP final automática.
5. Mostrar una vista de revisión editable, documento original y valores
   detectados en filas y columnas legibles.
6. Marcar campos de baja confianza, inconsistencias aritméticas, duplicados y
   proveedor no encontrado.
7. Exigir confirmación humana y permiso efectivo antes de crear el borrador o
   la obligación.
8. Aplicar idempotencia por empresa, hash y referencia documental.
9. Auditar archivo, extracción, edición, confirmación y resultado, sin guardar
   el contenido sensible en logs o prompts innecesarios.
10. Probar documento válido, borroso, rotado, multipágina, duplicado,
    manipulado, tipo inválido, importe discordante y documento de empresa B.

**Aceptación:** el usuario puede cargar, revisar y editar; ninguna lectura se
contabiliza sin confirmación; duplicados y cruces de tenant quedan bloqueados.

### P108-007 - Consistencia financiera y contable transversal [P0]

**Objetivo:** demostrar integridad entre operación, cartera y contabilidad.

**Acciones:**

1. Definir fuentes canónicas para ventas, pagos, inventario, CxC, CxP,
   impuestos y asientos.
2. Probar venta de contado, crédito, mixta, devolución, nota, anulación,
   descuento, impuesto, retención y redondeo.
3. Conciliar documento, pago, caja/banco, inventario, costo, cartera, impuesto
   y asiento después de cada evento.
4. Probar moneda y precisión con enteros/unidad menor o decimal PostgreSQL
   definido; evitar `float` para dinero.
5. Probar doble clic, reintento, timeout ambiguo y dos actores simultáneos.
6. Crear reportes de conciliación que no modifiquen datos automáticamente.

**Aceptación:** no hay descuadres ni efectos parciales; cada movimiento se puede
trazar hasta su origen y reconstruir de manera determinista.

### P108-008 - Cierre del contador profesional y Plan 107 [P0]

**Objetivo:** certificar el sistema para trabajo contable profesional.

**Acciones:**

1. Verificar en navegador permisos de contador y administrador sin elevación.
2. Consolidar el catálogo real de reportes y eliminar discrepancias entre los
   inventarios de 43 y 46 datasets.
3. Ejecutar fixture contable determinista solo en staging.
4. Probar ciclo integral: apertura, menta, contado, crédito, compras,
   proveedores, CxC/CxP, bancos, caja, nómina, activos, ajustes e impuestos.
5. Completar balance de prueba, mayores, auxiliares, estado de resultados,
   situación financiera, patrimonio, flujo de efectivo cuando aplique y notas.
6. Diferenciar claramente preliminar/no oficial de apto para presentación.
7. Validar IVA, retenciones, notas, compras, ventas y reglas Colombia por
   versión y período.
8. Validar exógena y certificados con catálogo versionado.
9. Validar cierres, reaperturas, períodos bloqueados y trazabilidad.
10. Probar exportación, impresión y programación.
11. Ejecutar UAT independiente con un contador y conservar acta de hallazgos.

**Aceptación:** reportes concilian con asientos canónicos, formatos declaran
alcance y período, no hay SQL generado por IA y el UAT no conserva P0/P1.

### P108-009 - Reportes dinámicos seguros con IA [P0]

**Objetivo:** cerrar P107-011 y el fallo de contrato observado en staging.

**Acciones:**

1. Publicar el contrato actual de `ReportSpec`.
2. Corregir el rechazo del modo nuevo y mantener lista cerrada de dimensiones,
   métricas, filtros, orden y límites.
3. Validar el `ReportSpec` exclusivamente en servidor.
4. Resolver datos mediante consultas preparadas autorizadas; nunca SQL libre.
5. Aplicar empresa, rol, permisos, período, privacidad y máximo de filas.
6. Mostrar vista previa antes de exportar o guardar plantilla.
7. Versionar plantillas y auditar la huella del spec.
8. Probar prompt injection, campos inventados, consultas costosas, acceso de
   empresa B y rol sin permiso.

**Aceptación:** solicitudes permitidas producen resultados reproducibles;
solicitudes fuera del contrato se rechazan de forma segura y comprensible.

### P108-010 - Plan IA profesional, privacidad y herramientas [P0]

**Objetivo:** cerrar PIA-000 a PIA-010 sin ampliar escrituras inseguras.

**Acciones:**

1. Desplegar en staging la experiencia de un solo interruptor `Modo agente`,
   apagado por defecto y sin selector de agente.
2. Verificar preferencias, conversación y memoria aisladas por empresa y
   usuario, con consentimiento, caducidad y borrado por el dueño.
3. Derivar contexto, rol y permisos en servidor.
4. Mantener catálogo estricto de herramientas por dominio, sin parámetros de
   tenant, rol, usuario o confirmación controlables por el modelo.
5. Separar consultar, proponer y confirmar. Toda escritura requiere
   confirmación independiente y nueva comprobación de permisos.
6. Añadir idempotencia, correlación y auditoría común.
7. Crear evaluaciones deterministas de precisión, rechazo, fuga entre tenants,
   prompt injection, tool injection, archivos hostiles y cambios de rol.
8. Medir latencia, tokens, costo, errores, reintentos y tasa de confirmación.
9. Definir presupuesto, límites, timeout, fallback y degradación sin IA.
10. Probar cada botón o acción IA visible con resultados válidos, inválidos,
    cancelación y doble clic.

**Aceptación:** IA no puede saltarse permisos ni confirmar sus propias
propuestas; memoria es privada; todas las acciones son trazables; UX publicada
coincide con la implementada.

### P108-011 - Regresión funcional de todos los módulos [P0]

**Objetivo:** comprobar que cada módulo visible y cada función crítica opera.

**Acciones:**

1. Construir inventario canónico uniendo rutas, páginas, controles, permisos,
   APIs y módulos. Reconciliar las distintas métricas de inventario UX.
2. Asignar caso de prueba, rol, empresa, datos y resultado esperado a cada
   acción.
3. Probar navegación, alta, lectura, edición, eliminación/anulación,
   búsqueda, filtros, paginación, importación, exportación y errores.
4. Cubrir ventas, POS, caja, facturación, inventario, compras, usuarios,
   clientes, finanzas, cumplimiento, asistencia, producción, CRM, canales,
   análisis, documentos, soporte, administración, licencia y noticias.
5. Verificar consola sin errores no controlados y red sin 4xx/5xx inesperados.
6. Registrar explícitamente acciones no aplicables, ocultas o fuera del piloto.

**Aceptación:** 100 % del inventario incluido tiene evidencia aprobada o una
exclusión firmada; ningún botón queda “no probado” por omisión.

### P108-012 - QA visual, responsive y accesibilidad [P0/P1]

**Objetivo:** certificar la experiencia real en PC y celular.

**Acciones:**

1. Probar anchos representativos de escritorio, tableta y móvil.
2. Verificar filas, columnas, tablas, formularios, modales, menús, estados
   vacíos, carga, error, éxito y textos largos.
3. Probar teclado, foco, etiquetas, contraste, zoom, lector y objetivos táctiles.
4. Capturar antes/después y conservar resolución, navegador, rol y ruta.
5. Validar temas claro/oscuro sin que afecten impresión.
6. Probar todos los botones inventariados, incluidos los generados
   dinámicamente.

**Aceptación:** cero bloqueos visuales P0/P1, sin contenido cortado o controles
inaccesibles en tamaños soportados.

### P108-013 - Impresiones, vistas previas y documentos [P0]

**Objetivo:** revisar todas las representaciones imprimibles del sistema.

**Alcance mínimo:**

- recibos y comprobantes;
- facturas electrónicas y no electrónicas;
- ventas y cotizaciones;
- pedidos, remisiones y devoluciones;
- compras y cuentas por pagar;
- caja, cierres y movimientos;
- reportes contables, fiscales y ejecutivos;
- nómina, asistencia, inventario y etiquetas cuando apliquen.

**Acciones:**

1. Inventariar HTML, PDF, vista previa, descarga y envío asociado.
2. Probar A4, carta y formatos térmicos soportados.
3. Verificar encabezado, empresa, NIT, cliente/proveedor, fechas, consecutivo,
   filas, columnas, cantidades, impuestos, totales, moneda, firmas y pie.
4. Probar cero, una y muchas líneas; textos largos; saltos de página; tildes,
   símbolos y códigos QR/barras.
5. Comparar valores impresos con API y datos canónicos.
6. Verificar que botones de impresión no dupliquen operaciones.
7. Confirmar que no se filtren rutas, secretos ni datos de otra empresa.

**Aceptación:** cada documento es legible, consistente, paginado y conciliado;
las capturas y PDF se adjuntan a su caso.

### P108-014 - POS y varias cajas simultáneas [P0]

**Objetivo:** demostrar operación concurrente realista.

**Escenario mínimo:**

- cuatro sesiones de cajero independientes;
- dos productos compartidos y uno con existencia limitada;
- ventas simultáneas de contado, crédito y pago mixto;
- impresión, devolución y cierre de caja;
- un reintento o timeout controlado;
- administrador observando y conciliando.

**Comprobaciones:**

1. No vender existencia negativa salvo regla explícita.
2. Consecutivos únicos y orden correcto.
3. Cada caja conserva apertura, movimientos y cierre propios.
4. Pagos, inventario, documento, cartera, impuestos y asiento coinciden.
5. Doble clic o reconexión no duplica.
6. No hay datos de otra empresa.
7. Medir latencia y errores durante la concurrencia.

**Aceptación:** cuatro cajas terminan conciliadas, sin duplicados, bloqueos
indefinidos, saldos negativos indebidos ni pérdida de eventos.

### P108-015 - Proveedores e integraciones externas [P0]

**Objetivo:** demostrar los proveedores necesarios para el alcance de piloto.

**Matriz obligatoria por proveedor:**

- entorno y configuración;
- secreto almacenado de forma segura;
- prueba feliz;
- error, timeout y reintento;
- idempotencia;
- webhook autenticado cuando aplique;
- auditoría y métricas;
- degradación y runbook;
- evidencia sin secretos.

**Cobertura:**

1. DIAN: certificado por empresa, permisos de archivo seguros, vigencia, NIT,
   numeración, firma, envío y consulta final. Para aprobar se requiere evidencia
   oficial del documento de prueba; `GetStatusZip` debe informar aceptación.
2. Wompi y Epayco: sandbox o monto controlado, firma/webhook, duplicado,
   conciliación y reverso.
3. Correo y WhatsApp: entrega real autorizada, error y trazabilidad. Para todos
   los correos enviados por el dominio propio, verificar por separado:
   - el logo oficial del computador PCS embebido mediante `cid:` dentro del
     cuerpo HTML, usando el constructor corporativo común;
   - SPF y DKIM alineados con el dominio visible `From`;
   - DMARC en enforcement (`p=quarantine` o `p=reject`, `pct=100`);
   - SVG Tiny-PS cuadrado, público por HTTPS como `image/svg+xml`, y registro
     `default._bimi.powerfulcontrolsystem.com`;
   - preferencia BIMI de marca y certificado VMC o CMC vigente cuando el
     receptor lo exija. Gmail no sustituye de forma garantizada la inicial del
     remitente con un BIMI autoafirmado sin ese certificado;
   - prueba visual real en Gmail y al menos otro cliente compatible, guardando
     captura del avatar, cuerpo, remitente y resultados SPF/DKIM/DMARC sin
     exponer cabeceras sensibles.
   La aceptación no puede prometer el mismo avatar en clientes que no soportan
   BIMI; sí debe demostrar que PCS publica y envía correctamente la identidad
   oficial en todos los mensajes del VPS.
4. Nextcloud/OnlyOffice: autenticación, CSP final después de redirecciones,
   archivo privado, edición y permisos.
5. OpenAI: modelo/configuración, herramientas, límites, timeout, costo y
   comportamiento sin proveedor.
6. Rappi u otra integración: solo si forma parte del piloto firmado.

**Aceptación:** cada proveedor incluido tiene evidencia vigente del mismo
candidato; los no incluidos están desactivados de forma segura.

### P108-016 - Almacenamiento privado y ciclo de archivos [P0/P1]

**Objetivo:** que archivos y documentos sobrevivan despliegues, réplicas y
restauraciones sin exposición.

**Acciones:**

1. Inventariar uploads, firmas, adjuntos, exportaciones y archivos temporales.
2. Definir almacenamiento compartido u Object Storage para todas las réplicas.
3. Usar claves opacas, metadatos con tenant y URLs firmadas de corta duración.
4. Validar tipo, tamaño, contenido, nombre y antivirus cuando corresponda.
5. Probar cuota, borrado, retención, huérfanos y recuperación.
6. Migrar archivos existentes con checksum y reconciliación.
7. Probar que credenciales y firma DIAN sean legibles solo por el proceso
   autorizado, sin permisos excesivos ni mensajes que revelen rutas internas.

**Aceptación:** archivo creado en una réplica se lee desde otra; backup/restore
recupera contenido y metadatos; tenant B nunca obtiene acceso.

### P108-017 - Seguridad dinámica y hardening [P0]

**Objetivo:** complementar los validadores estáticos con pruebas del sistema en
ejecución.

**Acciones:**

1. Revisar autenticación, sesiones, cookies, CSRF, CORS, CSP, cabeceras, rate
   limits, uploads y manejo de errores.
2. Ejecutar DAST autenticado y no autenticado sobre staging autorizado.
3. Probar inyección SQL, XSS almacenado/reflejado/DOM, SSRF, path traversal,
   upload hostil, deserialización y redirecciones.
4. Verificar redacción de PII, tokens, cabeceras, errores PostgreSQL y rutas.
5. Revisar secretos, rotación, privilegios de servicios, redes y puertos VPS.
6. Eliminar dependencia de `unsafe-inline` o registrar excepción acotada con
   fecha de cierre.
7. Repetir panel y escaneo de ciberseguridad sobre imagen final.

**Aceptación:** cero P0/P1 abiertos y ninguna filtración de secretos, datos
interempresa o detalles internos.

### P108-018 - Observabilidad, SLO y respuesta a incidentes [P0]

**Objetivo:** detectar y operar fallos antes de afectar de forma prolongada a
usuarios.

**Acciones:**

1. Confirmar logs estructurados con correlación, empresa segura y redacción.
2. Instrumentar latencia, errores y volumen en HTTP, PostgreSQL, jobs, colas,
   IA y proveedores.
3. Crear tableros y alertas accionables para salud, saturación, fallos DIAN,
   pagos, backups, almacenamiento y dead-letter.
4. Establecer SLI/SLO y presupuesto de error del piloto.
5. Simular API caída, worker detenido, PostgreSQL lento, proveedor caído,
   almacenamiento lleno y cola acumulada.
6. Validar runbooks, escalamiento y tiempo de respuesta del responsable.

**Aceptación:** cada simulacro genera alerta, diagnóstico y recuperación
medibles sin exponer información sensible.

### P108-019 - Capacidad, carga, concurrencia y degradación [P0]

**Objetivo:** demostrar capacidad para el volumen de lanzamiento.

**Acciones:**

1. Definir carga esperada y límite por API, cajas, empresas, jobs e IA.
2. Ejecutar pruebas autenticadas progresivas, sostenidas y de pico.
3. Medir p50/p95/p99, errores, CPU, RAM, conexiones, locks, disco y colas.
4. Analizar consultas lentas y planes PostgreSQL de rutas críticas.
5. Probar pool agotado, timeout, retry storm y proveedor lento.
6. Verificar backpressure y degradación: ventas esenciales no deben depender de
   radio, IA u otros servicios no críticos.
7. Documentar capacidad segura y criterio de escalamiento.

**Aceptación:** se cumplen los SLO acordados durante la duración definida, sin
inconsistencia ni agotamiento no controlado.

### P108-020 - Backup, restauración, continuidad y rollback [P0]

**Objetivo:** poder recuperar el mismo sistema y sus datos dentro del RPO/RTO.

**Acciones:**

1. Respaldar PostgreSQL, archivos privados, configuración cifrada y metadatos.
2. Verificar checksums, retención, cifrado, acceso y copia fuera del VPS.
3. Restaurar en un entorno aislado desde cero.
4. Validar autenticación, datos, documentos, CxP, contabilidad, IA y archivos.
5. Medir RPO/RTO completos, no solo restauración de base.
6. Ensayar rollback de aplicación y estrategia compatible de base de datos.
7. Probar una migración fallida y una pérdida de réplica.

**Aceptación:** recuperación y rollback cumplen objetivos y son ejecutables por
el runbook sin conocimiento implícito.

### P108-021 - Aplicación móvil y API móvil [P1 o fuera de alcance]

**Objetivo:** decidir de forma explícita si móvil entra al lanzamiento.

Si entra:

1. Cerrar rotación/revocación de tokens, PKCE y sesiones por dispositivo.
2. Probar sincronización incremental, conflictos, offline y reintentos.
3. Probar idempotencia móvil transaccional y permisos.
4. Probar push, deep links, almacenamiento seguro y cierre de sesión.
5. Publicar contrato OpenAPI y cliente reproducible.

Si no entra:

- ocultar enlaces;
- desactivar endpoints no necesarios o limitar su exposición;
- documentar fecha y condición para una fase posterior.

### P108-022 - Higiene documental, soporte y entrenamiento [P1]

**Objetivo:** que operación y soporte no dependan del conocimiento del
desarrollador.

**Acciones:**

1. Corregir las 218 secuencias sospechosas de `CHANGELOG.md` y repetir el gate.
2. Actualizar arquitectura, módulos, BD, permisos, flujos y archivos.
3. Consolidar runbooks de despliegue, rollback, backup, restauración,
   proveedores, IA, DIAN, pagos e incidentes.
4. Preparar manuales por rol y entrenamiento de cajero, administrador,
   contador y soporte.
5. Registrar límites conocidos, módulos deshabilitados y contactos.
6. Verificar que documentación y evidencias no contengan secretos.

**Aceptación:** documentación sin advertencias, runbooks ensayados y usuarios
del piloto capacitados.

### P108-023 - Staging equivalente y ensayo general [P0]

**Objetivo:** ejecutar una certificación completa sobre el candidato inmutable.

**Acciones:**

1. Desplegar por digest el candidato en staging equivalente.
2. Aplicar migraciones mediante el migrador y verificar API/worker sin DDL.
3. Restaurar datos anonimizados o fixtures aprobados.
4. Ejecutar P108-004 a P108-020 contra el mismo digest.
5. Ejecutar rollback y volver a desplegar el candidato.
6. Corregir fallos mediante un nuevo SHA/digest y repetir las pruebas afectadas;
   no parchear contenedores manualmente.
7. Firmar matriz de resultados por backend/BD, frontend/UX y QA/operación.

**Aceptación:** todas las compuertas P0 están verdes sobre un único digest y no
hay diferencias no documentadas con el entorno objetivo.

### P108-024 - Prueba real controlada en Powerful Control System [P0]

**Objetivo:** validar como usuario real la empresa autorizada antes del GO.

**Preparación:**

- backup vigente y restauración comprobada;
- ventana de prueba y responsables;
- datos marcados y plan de limpieza/reverso;
- proveedores y efectos externos expresamente autorizados;
- cuenta obtenida del canal seguro, sin exponer la contraseña.

**Flujo mínimo:**

1. Inicio de sesión, empresa, rol, licencia y navegación.
2. Venta menta controlada y conciliación integral.
3. Cuatro cajas simultáneas.
4. Compra, proveedor, CxP, carga de factura/recibo con IA, edición y
   confirmación.
5. Reportes del contador y reporte dinámico seguro.
6. Chat IA, memoria y propuestas sin confirmación automática.
7. Vistas previas e impresiones representativas.
8. Proveedores externos autorizados; DIAN solo con la autorización operacional
   vigente y evidencia oficial.
9. Limpieza o reverso mediante flujos oficiales, nunca SQL directo.
10. Conciliación final de inventario, caja, cartera, documentos, impuestos y
    contabilidad.

**Aceptación:** acta firmada, sin P0/P1, datos de prueba conciliados y sin
efectos externos inesperados.

### P108-025 - Piloto, GO/NO-GO y producción general [P0]

**Objetivo:** liberar gradualmente y con salida segura.

**Acciones:**

1. Reunión formal de GO/NO-GO con matriz y riesgos residuales.
2. Congelar digest, configuración y migraciones.
3. Verificar backup, rollback, responsables y canales de soporte.
4. Desplegar primero a un piloto limitado.
5. Monitorear métricas técnicas y de negocio durante la ventana acordada.
6. Suspender o revertir ante cualquier umbral de abortar.
7. Ampliar producción solo tras el cierre del piloto.
8. Registrar decisión, fecha, responsables y evidencia.

**Aceptación:** no quedan P0/P1, el piloto cumple SLO y conciliaciones, y existe
aprobación explícita de negocio, técnica y operación.

## 6. Orden obligatorio y dependencias

```text
P108-000 candidato
  -> P108-001 calidad
  -> P108-002 migraciones
  -> P108-003 worker/outbox
  -> P108-004 aislamiento
  -> P108-005..010 datos, contador e IA
  -> P108-011..017 funcional, visual, documentos, cajas, proveedores y seguridad
  -> P108-018..020 observabilidad, carga y continuidad
  -> P108-021..022 alcance móvil y documentación
  -> P108-023 ensayo general
  -> P108-024 prueba real controlada
  -> P108-025 piloto y GO
```

Se permite trabajo paralelo solamente cuando el usuario lo autorice y las fases
no compartan esquema, contratos o archivos. Nunca se paralelizan migración
canónica, conciliación financiera y pruebas reales sobre los mismos datos.

## 7. Matriz mínima de evidencia

| ID | Entregable | Evidencia obligatoria | Estado inicial |
|---|---|---|---|
| P108-000 | Candidato inmutable | SHA, digest, árbol limpio, manifiesto | Pendiente |
| P108-001 | CI y supply chain | tests, race, scans, SBOM | Parcial local |
| P108-002 | Esquema determinista | vacía, upgrade, segunda pasada, no DDL | Pendiente |
| P108-003 | Worker/outbox | reinicio, duplicado, dead-letter | Parcial aislado |
| P108-004 | Tenant/roles | matriz A/B negativa | Parcial estático |
| P108-005 | CxP canónica | ADR, migración, conciliación, concurrencia | Parcial local |
| P108-006 | Documentos proveedor IA | carga, edición, confirmación, duplicado | Parcial local |
| P108-007 | Consistencia financiera | conciliación por evento | Pendiente |
| P108-008 | Contador profesional | ciclo, reportes, impuestos, UAT | Parcial local |
| P108-009 | ReportSpec IA | staging, rechazo seguro, exportación | Parcial local |
| P108-010 | IA profesional | roles, memoria, tools, evals, costos | Parcial local |
| P108-011 | Módulos y funciones | inventario completo E2E | Parcial estático |
| P108-012 | Visual/responsive | capturas y accesibilidad | Pendiente |
| P108-013 | Impresiones | PDFs/capturas conciliados | Pendiente |
| P108-014 | Cuatro cajas | sesiones y conciliación concurrente | Pendiente |
| P108-015 | Proveedores | evidencia real autorizada | Parcial histórica |
| P108-016 | Archivos privados | réplica, tenant, backup/restore | Pendiente |
| P108-017 | Seguridad dinámica | DAST y hardening | Parcial estático |
| P108-018 | Observabilidad | alertas y simulacros | Parcial local |
| P108-019 | Capacidad | carga autenticada sostenida | Parcial pública |
| P108-020 | Continuidad | restore y rollback completos | Parcial |
| P108-021 | Móvil | certificación o exclusión formal | Aprobada: nativa fuera del primer lanzamiento; PWA incluida |
| P108-022 | Documentación | gate limpio y runbooks | Parcial |
| P108-023 | Ensayo general | un digest, matriz verde | Pendiente |
| P108-024 | Empresa autorizada | acta y conciliación | Pendiente |
| P108-025 | Piloto/GO | aprobación y ventana estable | Pendiente |

## 8. Estructura de evidencia

Cada caso debe crear o actualizar:

```text
documentos/evidencia_plan_108/
  P108-NNN/
    manifest.md
    comandos.txt
    resultados.md
    capturas/
    reportes/
    conciliacion/
    riesgos.md
```

`manifest.md` debe contener SHA, digest, ambiente, fecha, rol, empresa
seudonimizada, datos de prueba, criterio esperado y resultado. No debe contener
contraseñas, tokens, firmas, cookies ni PII innecesaria.

## 9. Comandos base

Ejecutar desde las rutas indicadas y adaptar solo cuando la fase lo documente:

```powershell
Set-Location D:\powerfulcontrolsystem
git status --short
git rev-parse HEAD
pwsh.exe -NoProfile -File .\scripts\profesional_preflight.ps1 -Full

Set-Location D:\powerfulcontrolsystem\backend
go test ./... -count=1
go vet ./...
```

La prueba `-race`, escaneos de imagen, migraciones, backup, restore, carga y
runtime se ejecutan en el entorno Linux autorizado. No se deben presentar como
aprobados a partir de una simulación incompatible en Windows.

Antes de pruebas visuales o despliegue se debe leer
`documentos/comandos_codex.md`. Antes de cualquier endpoint o dato empresarial,
se debe aplicar
`documentos/checklist_seguridad_endpoint_multiempresa.md`.

## 10. Regla de avance y porcentaje

El porcentaje se calcula por compuertas, no por archivos editados ni por tiempo:

- P108-000 a P108-025 tienen el mismo valor base.
- Una fase `aprobada` vale 100 % de su valor.
- Una fase `parcial` vale como máximo 50 % y debe listar qué evidencia falta.
- `pendiente`, `bloqueada` o `fallida` valen 0 %.
- Ninguna evidencia histórica cuenta si no corresponde al SHA/digest candidato,
  salvo que la fase demuestre que es independiente del artefacto.
- La certificación productiva permanece en 0 % hasta existir candidato
  inmutable. Después solo suma fases aprobadas sobre ese candidato.

Siempre se deben informar dos cifras:

1. **avance de implementación**, que puede incluir trabajo local parcial;
2. **certificación del candidato para producción**, que solo incluye evidencia
   del artefacto inmutable.

### Corte verificable 2026-07-28

- Candidato inmutable:
  `7ca5fb1be10d1f02fe3e0a7c5009f559c9d6f853`.
- P108-000 y P108-001: aprobadas.
- P108-002 a P108-013, P108-015, P108-017 a P108-020 y P108-022:
  parciales. P108-019 solo tiene smoke público; no cuenta como carga
  autenticada ni certificación operativa.
- Las demás fases: pendientes.
- Avance de implementación por compuertas:
  `(2 aprobadas + 17 parciales x 0,5) / 26 = 40,4 %`.
- Certificación sobre el digest actual:
  P108-000 y P108-001 aprobadas; P108-002, P108-011 y P108-012 parciales.
  `(2 + 3 x 0,5) / 26 = 13,5 %`.

Este porcentaje no equivale a cantidad de código construido. Mide evidencia de
aceptación para producción; no se eleva por pruebas históricas, locales o de un
SHA diferente.

### Corte verificable 2026-07-30

- El candidato inmutable activo en staging es
  `f9396da5e41562968996b05136fffca9991b56f9`; GitHub Actions aprobó build,
  `go test -race`, seguridad, secretos, escaneo, SBOM, publicación por digest
  y Compose.
- P108-018 ganó métricas privadas para PostgreSQL, worker, outbox y trabajos
  durables. El scrape real del digest aprueba y la exposición pública devuelve
  404; faltan los simulacros específicos y la configuración permanente.
- P108-019 atendió 500 lecturas autenticadas con concurrencia 10, cero fallos y
  p95 de 134 ms sobre este mismo digest; aún falta carga sostenida y
  transaccional.
- P108-020 ya tiene snapshot completo y restore PostgreSQL aislado, pero aún
  falta recuperación funcional integral y rollback del mismo candidato.
- P108-014 continúa bloqueada: la empresa autorizada no dispone de cuatro
  usuarios cajeros activos. Ya existen cuatro invitaciones temporales creadas
  por el flujo oficial y entregadas por Mailu; falta confirmarlas y abrir sus
  cajas.
- P108-006 continúa bloqueada por una credencial IA cifrada con una llave
  incompatible en staging.
- El barrido ampliado cubrió 48 rutas en escritorio y móvil: 96 combinaciones,
  2.148 botones, 104 clics seguros y 253 mutaciones preservadas. Tras corregir
  el contexto de empresa del runner, la repetición dirigida pasó 14/16; solo
  Crédito conserva un 500 en el digest anterior.
- El candidato corrige el agregado vacío de Créditos, hace idempotente la
  siembra concurrente de Contabilidad Colombia, instala CSRF en 19 páginas
  empresariales mutantes y corrige el desborde móvil de Cobranza. La repetición
  desplegada terminó 16/16 `ok`; el botón real de reenvío de confirmación
  respondió correctamente y la captura móvil no presenta desborde.
- La carga del mismo digest terminó 500/500, p95 134 ms, cero fallos, cero
  sesiones esperando lock y señales saludables de PostgreSQL, worker, outbox y
  trabajos durables.
- La batería imprimible volvió a aprobar 18/18 y la revisión visual de factura
  carta/POS y corte de caja confirmó filas, columnas y totales ordenados. La
  fase sigue parcial por documentos reales extensos e impresión física.
- P108-021 queda aprobada por exclusión formal: la web responsive/PWA entra al
  lanzamiento, el cliente nativo queda para una fase posterior y la API v1 se
  conserva autenticada sin enlaces de descarga nativa.
- Avance de implementación: **46,2 %**, aplicando tres fases aprobadas y
  dieciocho parciales; no se suman bloqueos ni pruebas incompletas.
- Certificación del candidato para producción: **21,2 %** y **NO-GO**:
  P108-000/P108-001/P108-021 aprobadas y P108-002, P108-011, P108-012,
  P108-018 y P108-019 parciales sobre el mismo artefacto.

## 11. Compuerta final GO/NO-GO

La decisión solo puede ser GO cuando todo lo siguiente está comprobado:

- [ ] alcance de lanzamiento firmado;
- [ ] candidato limpio, inmutable, reproducible y escaneado;
- [ ] CI, race, seguridad y supply chain verdes;
- [ ] migraciones deterministas y runtime sin DDL;
- [ ] worker, outbox, jobs e idempotencia aprobados;
- [ ] aislamiento A/B y permisos efectivos aprobados;
- [ ] CxP canónica migrada, conciliada y concurrente;
- [ ] carga IA de facturas/recibos editable y confirmable;
- [ ] consistencia financiera y contable integral;
- [ ] contador, reportes, impuestos, estados y UAT aprobados;
- [ ] IA privada, autorizada, auditable y evaluada;
- [ ] todos los módulos, funciones y botones probados;
- [ ] todas las acciones IA probadas;
- [ ] visual, responsive y accesibilidad aprobados;
- [ ] impresiones y vistas previas de todo el alcance aprobadas;
- [ ] cuatro cajas simultáneas conciliadas;
- [ ] proveedores incluidos con evidencia real autorizada;
- [ ] correos del VPS con logo oficial inline y avatar BIMI certificado,
      autenticación alineada y evidencia visual en clientes compatibles;
- [ ] DIAN aceptada oficialmente para la prueba autorizada;
- [ ] almacenamiento privado compartido y recuperable;
- [ ] DAST y hardening sin P0/P1;
- [ ] observabilidad y simulacros operativos aprobados;
- [ ] carga sostenida cumple SLO;
- [ ] backup, restauración y rollback cumplen RPO/RTO;
- [ ] móvil aprobado o excluido formalmente;
- [ ] documentación, soporte y entrenamiento cerrados;
- [ ] ensayo general del mismo digest aprobado;
- [ ] prueba controlada en Powerful Control System conciliada;
- [ ] piloto estable y decisión explícita de responsables.

Si una sola casilla P0 permanece abierta, el resultado es **NO-GO**.

## 12. Primera orden para GPT-5.6 Terra medio

La primera ejecución de este plan debe limitarse a `P108-000`:

1. leer la documentación obligatoria;
2. auditar el árbol sucio sin borrar cambios;
3. clasificar cada diferencia por plan y módulo;
4. detectar conflictos, duplicados, secretos y archivos generados;
5. proponer el contenido exacto del candidato;
6. ejecutar pruebas enfocadas sin desplegar;
7. dejar el manifiesto y el estado real de la fase;
8. detenerse si necesita autorización para integrar, desplegar o producir
   efectos externos.

No se debe comenzar P108-002, staging ni pruebas reales mientras P108-000 no
produzca un candidato limpio e inequívoco.

## 13. Cierre de la planificación

El estado inicial del Plan 108 es **NO-GO**. El preflight local demuestra una
base técnica valiosa, pero no demuestra por sí solo preparación para producción.
La prioridad inmediata no es agregar más funciones: es estabilizar el candidato,
cerrar las fuentes canónicas y demostrar de extremo a extremo, sobre el mismo
artefacto, que PCS conserva seguridad, consistencia, usabilidad y capacidad de
recuperación.
