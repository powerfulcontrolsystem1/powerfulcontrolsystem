# Plan 110 - cierre verificable para entrada en producción

Fecha de creación: 2026-08-11

Estado inicial: **NO-GO**

Modelo ejecutor previsto: **GPT-5.6 Terra con razonamiento medio**

Empresa autorizada para pruebas: **Powerful Control System (PCS)**

Ámbito prioritario: **aplicación web responsive**

## 1. Objetivo y condición de cierre

El Plan 110 recoge únicamente el trabajo que todavía falta para convertir las
evidencias parciales del Plan 109 en una autorización verificable de entrada en
producción. No vuelve a contar pruebas antiguas como si pertenecieran al
candidato final y no permite declarar `100 %` mientras exista una compuerta P0
abierta.

El plan termina cuando ocurre una de estas dos condiciones:

1. **GO:** un único SHA y sus cuatro imágenes por digest aprobaron integración,
   staging, pruebas reales PCS, seguridad, restore, ensayo general y piloto; la
   decisión está firmada y la promoción de esos mismos digests fue verificable.
2. **NO-GO formal:** se identifica un bloqueo externo o técnico no corregible
   dentro del alcance, se conserva evidencia, se asigna responsable y no se
   promueve producción.

La aplicación móvil nativa queda fuera de este cierre. La web móvil y la PWA sí
son obligatorias. Ningún enlace, permiso o proceso nativo incompleto puede
quedar visible o habilitado en el piloto sin una exclusión comprobable.

## 2. Línea base auditada del Plan 109

### 2.1 Estado heredado

- Último porcentaje formal publicado por el Plan 109: **56,7 % de
  implementación**, **6,7 % de certificación del candidato** y **NO-GO**.
- P109-006 demostró cuatro cajas concurrentes sobre un digest concreto, pero se
  debe repetir la aceptación abreviada sobre el candidato final.
- CxP demostró aislamiento A/B, precisión monetaria, pago idempotente y rechazo
  de sobrepago; faltan aceptación contable, recuperación operativa elegible y
  decisión documentada sobre la fuente canónica de cartera.
- CxP/IA demostró carga limpia, rechazo EICAR, fail-closed, extracción real y
  edición humana. Faltan el conjunto completo de botones/escenarios IA, evals,
  aislamiento integral y cierre operativo de alertas.
- La regresión sintética de impresión aprobó 20 formatos y una factura real en
  PDF. Faltan dispositivos físicos, tableta y accesibilidad asistida.
- DIAN emitió una factura y una nota crédito reales, pero la evidencia oficial
  aplicable a cada tipo de respuesta debe quedar conciliada; no se debe inventar
  un `TrackId` cuando `SendBillSync` solo devuelve un acuse síncrono.
- Backup, restore, réplica A/B y rollback tienen ensayos positivos en recursos
  aislados. Deben repetirse contra el candidato final y recibir aceptación
  contractual de RPO/RTO.
- Prometheus cargó las reglas de antivirus y Alertmanager interno recibió una
  alerta real de indisponibilidad. Faltan resolución confirmada, canal externo,
  deduplicación, responsables y simulacro completo.
- UAT del contador, matriz mutante por rol, impresión/accesibilidad asistida,
  ensayo general, piloto y decisión GO siguen sin firma.

### 2.2 Candidato existente que no cierra el plan

El workflow inmutable `31470172572` terminó correctamente para el SHA
`9a560767517a8a6d76506eee4c7b04d7299db0db`: construyó, escaneó, generó SBOM,
publicó y validó cuatro digests. Después se corrigió el validador de restore en
otro commit, por lo que ese workflow no representa por sí solo el SHA final del
Plan 110. La promoción y el drill de restauración de ese intento no quedaron
completados.

Terra debe comenzar verificando nuevamente Git, CI, staging y los digests. No
debe suponer que un servicio sano equivale a un candidato certificado.

### 2.3 Correspondencia completa P109 -> P110

| Fase heredada | Cierre en P110 | Brecha que no se puede omitir |
|---|---|---|
| P109-000 | P110-001, P110-008 | integración limpia y digest final único |
| P109-001 | P110-002 | recuperación elegible, conciliación y fuente CxP canónica |
| P109-002 | P110-003 | todos los botones IA, evals y aislamiento completo |
| P109-003 | P110-002 | catálogo contable/fiscal y UAT firmado |
| P109-004 | P110-004 | acciones mutantes reales por rol y módulo |
| P109-005 | P110-005 | tableta, impresión física y accesibilidad asistida |
| P109-006 | P110-004, P110-009, P110-010 | repetición abreviada de cuatro cajas en digest final |
| P109-007 | P110-006 | DIAN, Mailu, pagos e integraciones oficiales |
| P109-008 | P110-007, P110-008 | antivirus/retención y restore del candidato final |
| P109-009 | P110-007 | DAST, CSP, sesión y A/B integral |
| P109-010 | P110-009 | canal externo, deduplicación, resolución y simulacro |
| P109-011 | P110-001, P110-008 | migraciones y rollback ligados al digest final |
| P109-012 | P110-011 | runbooks, capacitación, responsables y aceptación |
| P109-013 | P110-010 | ensayo general real PCS y conciliación completa |
| P109-014 | P110-011, P110-012 | piloto, GO/NO-GO, promoción y estabilización |

Esta correspondencia es exhaustiva: una fase P109 no se considera cerrada por
estar mencionada; debe cumplir el criterio de la fase P110 que la recibe.

## 3. Autorización de pruebas reales y manejo de credenciales

El usuario autoriza pruebas reales dentro de la empresa **Powerful Control
System**, incluida la cuenta de pruebas indicada en `AGENTS.md` y entregada en
el contexto seguro de esta tarea. La contraseña no se duplica en este archivo
porque no se pueden versionar secretos.

El ejecutor debe:

- leer la identidad autorizada desde `AGENTS.md` o el canal seguro vigente;
- cargarla solo en memoria mediante `P110_QA_EMAIL` y `P110_QA_PASSWORD`;
- no imprimirla, registrarla en evidencias, incorporarla a comandos visibles,
  guardarla en archivos, historial de shell, commits, capturas o artefactos;
- iniciar sesión por el flujo oficial y verificar en servidor que la empresa
  efectiva sea PCS antes de cualquier mutación;
- usar APIs/UI oficiales, nunca SQL directo para producir efectos de negocio;
- registrar cada fixture creado y limpiarlo o revertirlo mediante el flujo
  oficial cuando termine;
- cerrar sesiones, desactivar identidades temporales y retirar túneles o
  recursos efímeros al finalizar.

La autorización cubre en PCS/staging ventas, cajas, inventario controlado,
CxP/CxC, archivos, IA, correos corporativos, pruebas EICAR, factura electrónica
real y nota crédito real. En DIAN se debe comenzar con el producto de menor
costo disponible y ejecutar solo el mínimo número de documentos necesario para
obtener evidencia concluyente. Una promoción a producción sigue condicionada
a la decisión GO firmada de P110-011; las pruebas no autorizan saltarse esa
compuerta.

## 4. Reglas de ejecución para GPT-5.6 Terra medio

1. Leer completos `AGENTS.md`, `documentos/contexto_general_del_sistema.md`,
   `documentos/contexto_especifico_del_sistema.md`,
   `documentos/contexto_codex.md`, `documentos/mapa_modulos.md`,
   `documentos/flujos_operativos.md`, `documentos/comandos_codex.md`,
   `documentos/decisiones_tecnicas.md`,
   `documentos/checklist_seguridad_endpoint_multiempresa.md`, el Plan 109 y
   este plan antes de editar.
2. Inspeccionar `git status`, rama, SHA, diferencias locales, CI y ambiente.
   Preservar cualquier cambio ajeno y no desplegar desde un árbol sucio.
3. Trabajar en bloques grandes por dependencia. No abrir una PR por prueba: usar
   una sola rama de integración y, si la protección de `main` lo exige, una
   única PR consolidada con revisión y CI verde.
4. Usar Go puro y PostgreSQL. No agregar dependencias ni cambiar `go.mod` sin
   autorización expresa y documentación técnica.
5. Mantener `empresa_id`, usuario efectivo, rol, permiso y licencia validados
   en servidor para toda consulta y mutación empresarial.
6. Mantener DIAN separado por empresa: NIT, resolución, consecutivos,
   certificado, clave, firma, documentos y trazabilidad son empresariales; el
   software DIAN SaaS puede ser compartido cuando corresponda.
7. Tratar timeouts como inconclusos. Una prueba unitaria, un mock, un daemon
   sano o una captura no sustituyen la prueba del flujo HTTP/UI real.
8. Usar el navegador interno para la prueba visible y una impresora virtual PDF
   para evidencia reproducible. La impresión física y el lector de pantalla
   requieren dispositivo/persona real y no pueden simularse como PASS.
9. No promover producción mientras el estado sea NO-GO. Cada acción remota debe
   declarar ambiente, efecto, reverso y criterio de parada.
10. Cada corrección funcional debe actualizar pruebas y la documentación
    obligatoria del repositorio; cada archivo nuevo debe registrarse en
    `documentos/descripcion_de_archivos`.
11. No crear porcentajes por cantidad de commits, páginas abiertas o tiempo
    invertido. Aplicar exclusivamente la fórmula de la sección 10.
12. Si una fase requiere firma humana, destino de alertas, certificado BIMI,
    impresora física o ventana de mantenimiento, avanzar las demás fases y
    registrar el bloqueo con responsable, sin falsificar aceptación.

### Registro obligatorio antes de cada mutación real

`fecha | P110 | ambiente | SHA/digests | empresa | usuario/rol | acción |
datos/documentos | efecto externo | rollback | responsable | resultado`

No incluir secretos, payloads privados completos, certificados, tokens ni
datos personales innecesarios.

## 5. Orden de ejecución y dependencias

El orden obligatorio es:

1. P110-000: gobierno, alcance y fuente única de verdad.
2. P110-001, incluida su ampliación obligatoria P110-001A, y P110-002 a
   P110-007: correcciones funcionales, saneamiento estructural y controles P0.
3. P110-008: congelar y desplegar el candidato inmutable final en staging.
4. P110-009: carga, observabilidad y simulacro sobre ese candidato.
5. P110-010: ensayo general real PCS sobre el mismo digest.
6. P110-011: piloto limitado y decisión firmada.
7. P110-012: promoción productiva y estabilización, solo si existe GO.

Si cambia código, migración, Compose, configuración crítica o imagen después de
P110-008, se genera un nuevo candidato y se repiten todas las certificaciones
afectadas. No se mezclan resultados de distintos digests.

## 6. Fases del Plan 110

### P110-000 - Gobierno, alcance y matriz de trazabilidad [P0]

**Objetivo:** congelar qué entra al piloto y convertir todos los pendientes del
Plan 109 en una matriz sin duplicados ni vacíos.

**Acciones:**

1. Inventariar módulos web, rutas, jobs, integraciones, roles, licencias,
   documentos imprimibles y botones IA visibles.
2. Marcar cada elemento como `incluido`, `deshabilitado verificablemente` o
   `excluido con aceptación`. Todo elemento visible se presume incluido.
3. Decidir el estado del módulo de domótica: incluirlo con todos sus controles
   o deshabilitarlo por permiso/feature flag durante el piloto. No dejarlo medio
   operativo.
4. Mantener fuera del cierre la aplicación móvil nativa y comprobar que la web
   responsive no depende de ella.
5. Consolidar evidencia reutilizable P109 por SHA/digest y marcar la que debe
   repetirse en el candidato final.
6. Publicar responsables de negocio, técnico, contador, seguridad, soporte,
   DIAN y observabilidad; definir SLO, RPO, RTO, horario y criterios de reverso.
7. Crear la matriz maestra:
   `requisito | origen P109 | módulo | responsable | ambiente | digest |
   evidencia | estado | bloqueo | fecha`.
8. Crear la matriz canónica de capacidades:
   `capacidad | página | API | servicio | tabla | permiso | propietario |
   escritura canónica | alias/legado | fecha de retiro | rollback`. Una
   similitud de nombre o estilo no permite fusionar módulos sin esta decisión.

**Aceptación:** 100 % del inventario tiene alcance, responsable y criterio de
aceptación; no existen módulos visibles sin clasificación.

**Evidencia:** `documentos/evidencia_plan_110/P110-000/`.

### P110-001 - Integración limpia y modelo de datos canónico [P0]

**Objetivo:** obtener una sola base de código revisable y eliminar ambigüedades
de datos antes de congelar imágenes.

**Acciones:**

1. Consolidar los cambios vigentes en una rama limpia derivada de `main` y
   resolver conflictos, especialmente archivos de historial/evidencia, sin
   descartar cambios del usuario.
2. Ejecutar CI profesional: preflight, pruebas Go, `go vet`, auditorías,
   secretos, dependencias, contenedores, Trivy y SBOM.
3. Adoptar un ADR que declare la tabla/fuente canónica de cartera de
   proveedores y el destino de cualquier tabla histórica alternativa.
4. Verificar restricciones, claves, `empresa_id`, precisión `NUMERIC(18,2)`,
   idempotencia y trazabilidad documento-pago-evento-asiento.
5. Probar migración de base vacía, upgrade de snapshot y segunda pasada; API y
   worker deben arrancar sin privilegios DDL.
6. Simular fallo antes, durante y después de una migración; demostrar rollback
   de aplicación y datos sin pérdida fuera del RPO.
7. Resolver cualquier check obligatorio rojo antes de continuar.

**Aceptación:** árbol limpio, revisión aprobada cuando la protección de rama la
exija, CI verde, ADR canónico aprobado, migraciones deterministas y cero
conflictos o cambios sin propietario.

**Rollback:** conservar la rama anterior y snapshot; no aplicar una migración
irreversible sin estrategia probada.

**Evidencia:** `documentos/evidencia_plan_110/P110-001/`.

#### P110-001A - Consolidación de código, módulos y deuda técnica [P0/P1]

Esta ampliación forma parte de P110-001 y no crea una fase adicional ni cambia
el denominador de 13 fases. P110-001 no puede aprobarse mientras P110-001A esté
abierta.

**Línea base confirmada el 2026-08-13:**

- 638 archivos Go y 7.183 funciones de producción; 450 funciones superan 100
  líneas y 124 superan 200;
- 52 grupos de cuerpos de función idénticos;
- 155 funciones `Ensure*`, 122 pasos de catálogo legado y 106 llamadas fuera
  del migrador, 31 de ellas alcanzables durante tráfico HTTP;
- 1.841 llamadas DB sin contexto en código de producción y 777 resultados
  descartados explícitamente pendientes de clasificación;
- al menos 100 grupos de declaraciones CSS idénticas, 209 scripts inline y
  deuda CSP en 195 páginas;
- cero rutas empresariales duplicadas, cero IDs HTML realmente duplicados y
  cero archivos frontend completos idénticos en el barrido. No se autoriza una
  reescritura general basada en falsos positivos.

**Acciones obligatorias:**

1. Congelar la matriz canónica por capacidad y clasificar cada página, ruta,
   servicio y tabla como `canónica`, `fachada intencional`, `compatibilidad con
   retiro` o `eliminada`. Toda compatibilidad debe tener propietario,
   telemetría, fecha de retiro, prueba y reverso.
2. Mantener MRP como producción canónica y decidir el retiro de
   `/api/empresa/produccion/bom*`; mantener CxP canónica en
   `empresa_cuentas_por_pagar` y probar que la tabla histórica no recibe CxP ni
   abonos nuevos; clasificar cada CRUD genérico de `modulos_faltantes.go`.
3. Crear pruebas de caracterización por acción antes de dividir
   `EmpresaCarritosCompraHandler`, `EmpresaFacturacionElectronicaHandler`,
   `EmpresaNominaSueldosHandler`, `EmpresaComprasDocumentosHandler`,
   `EmpresaImpresorasHandler`, `EmpresaDIANColombiaHandler` y `main`.
4. Separar handlers por caso de uso sin duplicar middleware, permiso,
   `empresa_id`, validación, transacción, idempotencia ni auditoría. El archivo
   coordinador debe registrar dependencias y delegar; no contener reglas de
   negocio repetidas.
5. Extraer únicamente utilidades idénticas con semántica compartida: primer
   valor no vacío, límites/paginación, patrones LIKE, detección de índices,
   períodos, host y entorno. Las coincidencias de dominio no se generalizan
   sin contrato y pruebas.
6. Dar a cada tabla un único dueño de esquema en migración inmutable. Eliminar
   las definiciones múltiples de tablas IA/Nextcloud y dejar cero DDL o
   `Ensure*` de esquema en tráfico HTTP; API y worker solo verifican readiness.
7. Propagar `r.Context()` desde el handler y usar `QueryContext`,
   `ExecContext` y `BeginTx` en rutas críticas de dinero, fiscal, inventario,
   autenticación, archivos e IA. Un contexto desligado solo se permite para un
   job durable o cleanup con timeout, dueño y telemetría.
8. Clasificar los 777 descartes explícitos como `cleanup seguro`, `best effort
   observable` o `error obligatorio`. Decodificación, auditoría, evento fiscal,
   pago y mutación no pueden fallar silenciosamente.
9. Consolidar primitivas visuales compartidas para pestañas, campos, chips,
   tablas, diálogos e impresión. Migrar por lotes con evidencia responsive,
   accesible y visual; compartir CSS no convierte dos módulos en uno.
10. Inventariar `alert`, `confirm`, `prompt`, `document.write`, `innerHTML`,
    `catch` vacío y `localStorage`. Eliminar los diálogos nativos de acciones
    críticas, demostrar escape DOM, limitar almacenamiento local a preferencias
    no autoritativas y hacer observables los fallos críticos.
11. Versionar una compuerta de complejidad/duplicación y publicar cobertura Go
    por paquete y total con timeout; exigir no regresión. Las rutas críticas
    refactorizadas deben cubrir éxito, entrada inválida, permiso, tenant A/B,
    concurrencia, cancelación, rollback y error externo.
12. Actualizar los contextos vigentes para señalar Plan 110 como única hoja de
    ruta activa y conservar planes anteriores solo como historial.

**Aceptación:** cero fuentes de escritura duplicadas, cero DDL en tráfico HTTP,
un dueño por tabla, cero duplicaciones exactas sin clasificar, contextos en
operaciones críticas, descartes críticos corregidos, handlers caracterizados y
divididos sin perder cobertura, contratos históricos con retiro gobernado y
regresión completa en verde sobre el mismo candidato.

**Rollback:** refactorizar en cortes pequeños protegidos por pruebas de
caracterización; conservar adaptadores de compatibilidad con telemetría hasta
probar uso cero; no mezclar cambios destructivos de datos con la división de
código; revertir el corte completo si cambia contrato, tenant, saldo o salida
imprimible.

**Evidencia:**
`documentos/evidencia_plan_110/P110-001/2026-08-13_auditoria_integral_duplicacion_calidad.md`.

### P110-002 - Finanzas, CxP y UAT del contador [P0]

**Objetivo:** cerrar P109-001/P109-003 y el trabajo pendiente del Plan 107 con
conciliación humana independiente.

**Acciones:**

1. Crear por flujos oficiales fixtures PCS con saldos controlados para venta,
   compra, CxP, CxC, pago parcial/total, devolución, descuento, impuestos,
   retenciones, cierre y reapertura.
2. Recuperar un evento outbox elegible y demostrar exactamente un efecto
   contable natural, reintento idempotente, lease, dead-letter y auditoría.
3. Repetir pago concurrente con misma clave y claves distintas; exigir un solo
   movimiento y rechazo de sobrepago.
4. Probar aislamiento A/B de cartera, proveedores, pagos, asientos, reportes,
   outbox, jobs, exportaciones, caché y archivos.
5. Conciliar `documento -> evento -> asiento -> mayor -> reporte -> reverso` con
   diferencia cero.
6. Ejecutar el catálogo contable aplicable: resultados, situación financiera,
   patrimonio, auxiliares, comparativos, notas, impuestos, exógena y cierres.
7. Obtener UAT de un contador distinto del desarrollador, con firma, fecha,
   hallazgos y decisión.

**Aceptación:** diferencias cero, débitos iguales a créditos, sin doble pago,
aislamiento A/B, reversos oficiales y UAT firmado.

**Rollback:** usar notas, anulaciones y reversos oficiales; nunca editar
asientos o saldos directamente por SQL.

**Evidencia:** `documentos/evidencia_plan_110/P110-002/`.

### P110-003 - IA empresarial, CxP/IA y revisión humana [P0]

**Objetivo:** probar todos los botones IA incluidos y demostrar que la IA nunca
contabiliza sin autorización humana válida.

**Acciones:**

1. Verificar la clave de cifrado y proveedor IA mediante configuración segura,
   sin mostrar credenciales.
2. Cargar factura y recibo reales de prueba; validar tamaño, tipo, firma,
   antivirus, cuota, almacenamiento privado, duplicado y edición de campos.
3. Ejecutar cada botón IA visible: extraer, reintentar, cancelar, confirmar,
   rechazar, editar, regenerar, exportar y acciones del modo agente.
4. Probar doble clic, timeout, cancelación, respuesta inválida, proveedor caído,
   datos incompletos, prompt injection y reanudación.
5. Confirmar que extracción y edición no crean CxP ni asiento; solo la
   confirmación humana autorizada puede producir la obligación canónica.
6. Probar ReportSpec, reportes IA, Centro IA, memoria por usuario/empresa,
   permisos, costo, tokens, latencia y auditoría.
7. Ejecutar A/B con identidad no global para archivos, memoria, prompts,
   historial, herramientas, reportes y resultados.
8. Publicar evals de exactitud y seguridad con conjunto, umbral, resultados y
   revisión de falsos positivos/negativos.

**Aceptación:** matriz IA 100 % PASS o exclusión firmada, revisión humana
obligatoria, cero fuga A/B, cero secreto expuesto y degradación clara.

**Rollback:** deshabilitar capacidad/proveedor por permiso o configuración sin
borrar el historial auditable.

**Evidencia:** `documentos/evidencia_plan_110/P110-003/`.

### P110-004 - Regresión web completa, roles y domótica [P0]

**Objetivo:** convertir el inventario estático en pruebas reales de todas las
funciones incluidas, priorizando la aplicación web.

**Acciones:**

1. Recorrer todas las rutas incluidas en escritorio, tableta y móvil web con
   consola, red, auditoría y estado visual.
2. Ejecutar CRUD, filtros, búsqueda, importación, exportación, carga, descarga,
   cambio de estado, pagos, anulaciones, recuperación, borrado y navegación.
3. Probar cada acción mutante con rol permitido y denegado: superadministrador,
   administrador, contador, cajero, vendedor, soporte y roles del piloto.
4. Resolver o excluir formalmente timeouts, 5xx, botones sin etiqueta,
   desbordes, mutaciones al cargar página y navegaciones externas.
5. Repetir una prueba abreviada de cuatro cajas sobre el candidato final:
   inventario compartido, cuatro pagos, reintentos, consecutivos y cierres.
6. Si domótica entra al piloto, probar por empresa: estaciones, varias
   Raspberry, alta/revocación de túnel, ID único, entrada/salida GPIO, horarios,
   zona horaria, historial, fotos, consumo, tráfico, reconexión y estado seguro
   `apagado` ante pérdida. La activación física de relés debe ser supervisada.
7. Si domótica no entra, ocultar y denegar sus rutas/API por permiso o feature
   flag y probar que una URL directa tampoco la habilita.

**Aceptación:** 100 % de acciones incluidas tienen PASS o exclusión firmada;
cero 5xx, mutaciones inesperadas, accesos indebidos o efectos sin auditoría.

**Evidencia:** `documentos/evidencia_plan_110/P110-004/`.

### P110-005 - Calidad visual, accesibilidad e impresiones [P0]

**Objetivo:** cerrar la calidad visible de toda la web y de todo documento
imprimible real.

**Acciones:**

1. Revisar escritorio, 390 px, tableta y resoluciones del piloto por cada rol.
2. Probar teclado, orden de foco, zoom 200 %, contraste, mensajes, nombres
   accesibles y lector de pantalla con apoyo real.
3. Revisar factura, recibo, comprobante, pedido, orden, cotización, corte,
   reporte, nota crédito y todo formato incluido.
4. Usar documentos PCS cortos, extensos y multipágina; validar Carta, POS,
   vista previa, PDF e impresión física.
5. Confirmar filas/columnas, encabezados repetidos, totales, moneda, impuestos,
   firmas, QR/CUFE, saltos, márgenes, logos, pies y ausencia de recortes.
6. Probar estados vacíos, carga, error, permiso denegado y datos largos.

**Aceptación:** cero defecto visual crítico; todos los formatos/dispositivos
tienen firma visual; teclado y lector son operables; impresión física aprobada.

**Evidencia:** `documentos/evidencia_plan_110/P110-005/`.

### P110-006 - DIAN, correo corporativo, pagos e integraciones [P0/P1]

**Objetivo:** obtener evidencia real de cada proveedor externo incluido sin
confundir una respuesta local con aceptación del proveedor.

**Acciones DIAN:**

1. Verificar por empresa NIT, ambiente, resolución/rango, consecutivo,
   certificado, clave, vigencia, permisos del archivo de firma y trazabilidad.
2. Emitir en PCS una factura real con el producto de menor costo disponible.
3. Registrar la respuesta oficial correcta: `StatusCode=00` cuando exista un
   TrackId reconsultable, o acuse síncrono auditable cuando el método no lo
   entregue. No fabricar ni reutilizar identificadores.
4. Comprobar visualmente XML/PDF, CUFE, QR, totales, filas y estado.
5. Anular por nota crédito total, validar relación y respuesta oficial, e
   imprimir ambos documentos.
6. Probar reintento/doble clic sin duplicar factura ni consecutivo y verificar
   aislamiento con una empresa B.

**Acciones de correo e integraciones:**

1. Probar registro de administrador completo, verificación, invitación, reset,
   alerta y rebote usando Mailu propio del VPS.
2. Verificar SPF, DKIM y DMARC; comprobar que el logo oficial aparece dentro del
   cuerpo y distinguirlo del avatar externo del remitente.
3. BIMI/VMC/CMC queda como dependencia comercial: adquirir/configurar el
   certificado solo con pago y autorización vigentes. Si no es requisito de
   salida, registrar una excepción P1 con responsable y fecha.
4. Probar Wompi/Epayco, WhatsApp, Nextcloud, OnlyOffice y demás proveedores solo
   si están incluidos; exigir resultado oficial, error seguro, reintento e
   idempotencia.

**Aceptación:** factura y nota crédito oficiales conciliadas; registro/correos
entregados; DNS de correo válido; cada integración incluida tiene evidencia del
proveedor y separación por empresa.

**Rollback:** reverso fiscal/oficial, deshabilitar proveedor por empresa y
preservar auditoría. No borrar ni alterar documentos fiscales por SQL.

**Evidencia:** `documentos/evidencia_plan_110/P110-006/`.

### P110-007 - Seguridad multiempresa, CSP y antivirus [P0]

**Objetivo:** cerrar las amenazas dinámicas y el aislamiento completo del
candidato.

**Acciones:**

1. Ejecutar DAST autenticado y público contra staging, con alcance y límites
   controlados.
2. Probar XSS, CSRF, SSRF y redirecciones, IDOR, CORS, sesión/revocación,
   enumeración, rate limit, archivos hostiles, traversal, symlink, ZIP bomb y
   contenido activo.
3. Completar A/B en SQL, archivos, caché, jobs, colas, reportes, exportaciones,
   IA, DIAN, domótica y descargas.
4. Probar ClamAV por el endpoint HTTP real: limpio, EICAR antes de persistir,
   servicio caído fail-closed, recuperación y concurrencia. Confirmar cero
   bypass.
5. Completar CSP: eliminar `unsafe-inline` o documentar excepción mínima con
   responsable y vencimiento; validar cabeceras en rutas públicas/autenticadas.
6. Escanear digests finales y revisar hardening Docker/VPS, parches pendientes,
   reinicio en ventana, puertos, usuarios, claves, backups y exposición de
   métricas.
7. Ejecutar escaneo final de secretos sin imprimir valores.

**Aceptación:** cero P0/P1 explotable, aislamiento A/B completo, EICAR rechazado
antes de persistencia, sesión segura y CSP/riesgos aceptados formalmente.

**Evidencia:** `documentos/evidencia_plan_110/P110-007/`.

### P110-008 - Candidato final, migración, restore y rollback [P0]

**Objetivo:** congelar un único candidato y demostrar recuperación integral con
sus imágenes exactas.

**Acciones:**

1. Construir API, migrador, worker y frontend una sola vez desde el SHA completo
   final; escanear, generar SBOM y publicar referencias `@sha256`.
2. Fijar también el digest de ClamAV y validar Compose sin tags mutables.
3. Promover esos mismos digests solo a staging, sin recompilar.
4. Verificar migrador, backend, worker, frontend, ClamAV, `/health`, `/ready`,
   métricas, logs y producción intacta.
5. Restaurar snapshot en red, bases, volumen y puerto efímeros; validar CxP,
   contabilidad, IA, DIAN, documentos y archivos mediante login oficial.
6. Ejecutar dos réplicas, carga por A/lectura por B, pérdida de A, pérdida de
   bases/volumen y rollback coordinado de aplicación/datos.
7. Medir RPO/RTO, checksums, filas críticas, huérfanos y limpieza total de
   recursos efímeros.

**Aceptación:** mismo SHA/digests, staging sano, restore integral, réplica
intercambiable, rollback dentro de RPO/RTO firmado y cero residuo.

**Rollback:** restaurar digests previos de staging y snapshot aprobado; nunca
usar una imagen mutable o la base activa para el drill.

**Evidencia:** `documentos/evidencia_plan_110/P110-008/`.

### P110-009 - Carga, observabilidad y simulacro de incidente [P0]

**Objetivo:** demostrar que el sistema detecta, comunica y recupera fallos bajo
carga realista.

**Acciones:**

1. Ejecutar cuatro cajas y carga autenticada sostenida; medir p50/p95/p99,
   errores, locks, pool, CPU, memoria, disco, colas y backpressure.
2. Ensayar outbox/job dead, lease vencido, ClamAV caído, proveedor IA caído,
   almacenamiento privado, disco alto, base degradada y worker detenido.
3. Confirmar firing, recepción externa, deduplicación, escalamiento y resolución
   para cada alerta P0.
4. Configurar un canal externo aprobado con destinatario y responsables. El
   receptor interno por sí solo no certifica operación.
5. Publicar dashboards, SLO, error budget, umbrales de disco y política segura
   de limpieza/retención.
6. Ejecutar simulacro con cronología, comunicación, rollback, recuperación,
   postmortem y acciones.

**Aceptación:** SLO cumplido, todas las alertas P0 recibidas/resueltas dentro del
objetivo, responsables contactables y cero pérdida o doble efecto.

**Evidencia:** `documentos/evidencia_plan_110/P110-009/`.

### P110-010 - Ensayo general real en PCS [P0]

**Objetivo:** ejecutar de punta a punta el alcance completo sobre el candidato
final como ensayo de producción.

**Acciones:**

1. Congelar digests, configuración, snapshot, responsables y ventana.
2. Usar la cuenta PCS autorizada y crear por flujo oficial las identidades
   adicionales necesarias para roles/cajas; confirmar correos reales.
3. Recorrer registro, empresa, licencia, usuarios, inventario, compras,
   proveedores, CxP/IA, contabilidad, ventas, cuatro cajas, DIAN/anulación,
   reportes, documentos, correo, seguridad e integraciones incluidas.
4. Ejecutar todas las matrices funcionales, IA, roles, A/B, visual, responsive,
   impresión y accesibilidad aplicables.
5. Simular un incidente y rollback sin cambiar de digest.
6. Conciliar ventas, inventario, cajas, cartera, impuestos, documentos fiscales,
   asientos, archivos y correos; diferencia cero.
7. Limpiar o revertir fixtures por flujos oficiales, revocar sesiones y firmar
   PASS/FAIL.

**Aceptación:** cero P0/P1, conciliación completa, mismo digest, rollback
probado, evidencia visible y responsables conformes.

**Evidencia:** `documentos/evidencia_plan_110/P110-010/`.

### P110-011 - Piloto limitado y decisión GO/NO-GO [P0]

**Objetivo:** validar operación humana real antes de cualquier promoción
productiva general.

**Acciones:**

1. Firmar alcance, riesgos, datos permitidos, módulos deshabilitados, soporte,
   SLO, RPO/RTO y reverso.
2. Ejecutar un piloto limitado con observación reforzada y el mismo digest del
   ensayo general.
3. Registrar errores, conciliación, soporte, integraciones, rendimiento y
   satisfacción por rol.
4. Resolver hallazgos o aceptar únicamente riesgos P1 no críticos con dueño y
   vencimiento; ningún P0 puede aceptarse.
5. Obtener firma de negocio, contador, responsable técnico, seguridad y
   operación.
6. Publicar decisión inequívoca `GO` o `NO-GO`.

**Aceptación:** piloto estable, firmas completas, cero P0/P1 crítico y decisión
GO explícita. Sin firmas, el resultado automático es NO-GO.

**Evidencia:** `documentos/evidencia_plan_110/P110-011/`.

### P110-012 - Promoción productiva y estabilización [P0]

**Objetivo:** promover exclusivamente el candidato aprobado y demostrar que la
operación productiva queda estable o revierte por completo.

**Precondición:** P110-000 a P110-011 aprobadas y acta GO firmada.

**Acciones:**

1. Verificar backup reciente, restore probado, capacidad, guardia, responsables
   y ventana de cambio.
2. Promover exactamente los mismos cuatro digests y configuración aprobada; no
   recompilar ni cambiar migraciones durante la ventana.
3. Ejecutar migración controlada, health/readiness, smoke por rol, worker,
   correo, IA e integración fiscal no destructiva.
4. Observar SLO, errores, colas, disco, seguridad, conciliación y soporte durante
   la ventana definida.
5. Activar rollback inmediato si se supera un umbral y demostrar recuperación.
6. Publicar acta, SHA/digests, horarios, métricas, incidentes y estado final.

**Aceptación:** producción estable durante la ventana de observación, misma
huella aprobada, cero P0/P1 y conciliación correcta; o rollback completo y
estado NO-GO documentado.

**Evidencia:** `documentos/evidencia_plan_110/P110-012/`.

## 7. Matrices obligatorias

### 7.1 Funciones y roles

`módulo | ruta/API | control | acción | rol permitido | rol denegado | empresa |
HTTP | efecto | auditoría | visual | responsive | reverso | digest | resultado`

### 7.2 IA

`botón | permiso | entrada válida/inválida | doble clic | cancelación | timeout |
proveedor caído | edición humana | confirmación | empresa B | costo/tokens |
latencia | auditoría | efecto financiero | resultado`

### 7.3 Finanzas y fiscal

`documento | venta/compra | pago | caja/banco | inventario | impuesto/retención |
evento | asiento | reporte | reverso/nota crédito | respuesta DIAN | diferencia`

### 7.4 Visual e impresión

`formato | documento real | corto/extenso | Carta/POS | navegador | resolución |
rol | filas/columnas | totales | QR/CUFE | páginas | PDF | impresión física |
teclado/lector | resultado`

### 7.5 Multiempresa

`dominio | objeto empresa A | identidad empresa B | lectura | mutación |
exportación | archivo | caché/job | respuesta pública | auditoría | resultado`

### 7.6 Evidencia externa

`proveedor | ambiente | solicitud | idempotencia | identificador oficial |
respuesta oficial | conciliación | reverso | responsable | resultado`

### 7.7 Arquitectura y calidad de código

`capacidad | página | API | servicio | tabla | dueño canónico | alias/legado |
fecha de retiro | función/archivo crítico | duplicación | contexto DB |
errores descartados | cobertura antes/después | rollback | resultado`

## 8. Criterios de parada inmediata

Detener la fase afectada, conservar evidencia y aplicar rollback si ocurre:

- fuga o mutación cruzada entre empresas;
- pago, venta, factura, nota crédito, CxP o asiento duplicado;
- diferencia contable/fiscal distinta de cero;
- pérdida de datos fuera del RPO o restore no funcional;
- secreto expuesto, P0/P1 explotable o antivirus en bypass;
- digest distinto al aprobado o producción modificada antes del GO;
- alerta P0 no recibida o sin responsable;
- 5xx sostenido, sobreventa, consecutivo duplicado o cuatro cajas descuadradas;
- certificado/firma DIAN incorrecto o respuesta oficial no conciliada.
- una segunda fuente de escritura, DDL durante tráfico, pérdida de filtro
  `empresa_id` o refactorización sin prueba de caracterización y reverso.

Un hallazgo P0 mantiene el plan en NO-GO, aunque otras fases continúen.

## 9. Compuerta final GO/NO-GO

Solo se puede declarar GO cuando:

- [ ] P110-000 a P110-011 están aprobadas.
- [ ] Existe un único SHA y cuatro digests certificados.
- [ ] CI, Trivy, SBOM, migraciones y staging están verdes.
- [ ] P110-001A cerró fuentes duplicadas, DDL runtime, deuda crítica de
  contexto/errores y contratos históricos sin reducir cobertura.
- [ ] Todas las funciones y todos los botones IA incluidos tienen evidencia.
- [ ] Matriz mutante por rol y aislamiento A/B están completos.
- [ ] CxP/contabilidad tienen diferencia cero y UAT del contador firmado.
- [ ] Cuatro cajas concurrentes cuadran sobre el candidato final.
- [ ] Factura DIAN y nota crédito tienen respuesta oficial conciliada.
- [ ] Registro/correos Mailu e integraciones incluidas funcionan.
- [ ] Visual, responsive, PDF, impresión física y accesibilidad están firmados.
- [ ] ClamAV rechaza EICAR por HTTP y falla cerrado.
- [ ] Restore/rollback cumplen RPO/RTO firmado.
- [ ] Alertas P0 llegan a un canal externo y se resuelven.
- [ ] Ensayo general y piloto están firmados sin P0/P1 crítico.
- [ ] Backup, ventana, responsables y rollback productivo están preparados.

P110-012 solo puede ejecutarse después de esta lista. Si cualquier punto falta,
la decisión es **NO-GO**.

## 10. Porcentaje del Plan 110

El Plan 110 tiene 13 fases de igual peso.

- `pendiente`, `bloqueada` o `fallida`: 0 %.
- `parcial`: 50 % solo con evidencia material del bloque actual.
- `aprobada`: 100 % con todos sus criterios y firmas requeridas.

Se publican tres cifras separadas:

1. **Implementación:** crédito por fases parciales/aprobadas.
2. **Certificación del candidato:** solo fases aprobadas sobre el mismo SHA y
   digests finales.
3. **Preparación productiva:** 100 % únicamente después de P110-012 aprobada.

Estado inicial del Plan 110:

- Implementación: **0 %**; el plan todavía no se ha ejecutado.
- Certificación del candidato final: **0 %**; aún no existe el digest final
  congelado del Plan 110.
- Preparación productiva: **0 %**.
- Evidencia histórica heredada: Plan 109 **56,7 % implementación / 6,7 %
  certificación**, reutilizable como antecedente, no como aprobación automática.
- Veredicto: **NO-GO**.

## 11. Formato de cierre de cada fase

Cada evidencia debe incluir:

- identificador P110, fecha, ambiente, rama, SHA y digests;
- empresa, identidad/rol sin secreto y permisos efectivos;
- alcance, datos creados y efectos externos;
- comandos/pruebas y resultados observados;
- capturas o artefactos visuales cuando aplique;
- conciliación y consultas de solo lectura;
- aislamiento A/B, idempotencia y auditoría;
- rollback/limpieza ejecutados;
- riesgos, responsables, firmas y vencimientos;
- estado formal y porcentaje calculado.

## 12. Primera orden para Terra medio

1. No implementar desde este documento hasta que el usuario lo ordene.
2. Al recibir autorización, abrir P110-000 y auditar el SHA/ambientes vigentes.
3. Crear la matriz maestra y decidir inclusión o feature flag de domótica.
4. Consolidar una única rama/candidato; no promover aún.
5. Ejecutar P110-001 a P110-007 en bloques grandes y documentados.
6. Solo entonces congelar digests en P110-008 y repetir toda prueba que dependa
   del candidato.
7. Detenerse ante una nueva necesidad de autoridad productiva o dependencia
   externa no proporcionada; continuar frentes independientes seguros.

Este documento es el cierre de planificación solicitado. Su creación no
autoriza ni ejecuta despliegue productivo.

## Actualización 2026-08-11, inicio de implementación P110

- P110-000 queda **parcial**: se auditó `main`, la rama P109, staging y las
  evidencias; la web/PWA y domótica se clasifican como incluidas, mientras la
  aplicación móvil nativa queda fuera. Aún faltan matriz firmada, responsables
  operativos, SLO/RPO/RTO aceptados y alcance final de integraciones.
- P110-001 queda **parcial**: se consolidó el bloque de trabajo sobre `main`,
  se verificó la ADR 106 y la persistencia ahora rechaza cualquier nueva CxP o
  abono CxP en la tabla histórica, incluso fuera de HTTP. Las pruebas
  focalizadas y el preflight completo (`go test`, `go vet`, migraciones,
  auditorías y Compose) aprobaron. Faltan CI remoto, migración, upgrade,
  rollback, conciliación real y el digest final.
- No se desplegó ni se modificó producción. Implementación P110: **7,7 %**
  (dos fases parciales de 13); certificación del candidato final: **0 %**;
  preparación productiva: **0 %**; veredicto **NO-GO**.

## Actualización 2026-08-11, aislamiento de staging P110-008

- La primera promoción del candidato se detuvo de forma segura antes de crear
  el nuevo backend: el init-container de permisos heredaba el nombre y volumen
  de la plataforma. No se tocó el contenedor ni el volumen de la plataforma.
- El override de staging ahora usa nombre y volumen exclusivos y queda cubierto
  por una prueba de regresión. El runner de restore permite indicar el entorno
  privado del staging aislado y reconoce el servicio backend actual. Se requiere
  un candidato inmutable nuevo para comprobar Compose y restore; P110-008 queda
  **parcial**.
- No se desplegó ni se modificó producción. Implementación P110: **11,5 %**
  (tres fases parciales de 13); certificación del candidato final: **0 %**;
  preparación productiva: **0 %**; veredicto **NO-GO**.

## Actualización 2026-08-11, candidato final, ClamAV y restore

- P110-007 queda **parcial**: el candidato `fd6a4a8a` comprobó por HTTP
  autenticado archivo limpio, EICAR previo a persistencia, caída fail-closed y
  recuperación de ClamAV. Permanecen pendientes DAST, CSP, sesión, A/B integral,
  concurrencia y escaneo final de seguridad.
- P110-008 queda **parcial**: CI generó cuatro digests inmutables, staging los
  promovió sin recompilar y el restore efímero validó dos bases, cinco tablas
  críticas y el inventario privado (RTO observado 86 s). Faltan dos réplicas,
  pérdida de A, rollback coordinado, RPO/RTO aceptado y firmas operativas.
- No se desplegó ni se modificó producción. Implementación P110: **15,4 %**
  (cuatro fases parciales de 13); certificación del candidato final: **0 %**;
  preparación productiva: **0 %**; veredicto **NO-GO**.

## Actualización 2026-08-11, ClamAV dentro del restore A/B

- El primer intento de dos réplicas detectó correctamente que ClamAV obligatorio
  no existía dentro de la red efímera: la carga devolvió 503 y no se aceptó un
  bypass. El drill y su runner ahora incluyen el digest de ClamAV, lo levantan
  sin exposición pública y esperan su disponibilidad antes de las operaciones
  autenticadas.
- Se debe crear un candidato nuevo y repetir la matriz A/B y rollback sobre el
  mismo digest. P110-008 sigue **parcial**; no cambia el **15,4 %** de
  implementación ni el veredicto **NO-GO**.

## Actualización 2026-08-11, A/B y rollback del restore

- El candidato `9ee2fe9a` aprobó la matriz aislada con ClamAV obligatorio:
  creación en réplica A, descarga con checksum igual en B, continuidad tras
  pérdida de A y rollback coordinado de dos bases y almacenamiento privado.
  No quedaron recursos efímeros y staging permaneció sano.
- P110-008 sigue **parcial** porque RPO/RTO aún no tienen aceptación firmada y
  faltan las compuertas de carga, operación y ensayo general. El avance se
  mantiene en **15,4 %**, con certificación final y preparación productiva en
  **0 %**; veredicto **NO-GO**.

## Actualización 2026-08-11, DAST autenticado no mutante

- El candidato `9ee2fe9a` superó la comprobación dinámica autenticada de
  sesión, CSRF, Origin, CORS, límite por empresa y logout, sin crear datos.
- P110-007 permanece **parcial**: aún requiere DAST integral, CSP evaluada,
  matriz A/B de dominios y alerta externa. El avance se mantiene en **15,4 %**
  y el veredicto sigue siendo **NO-GO**.

## Actualización 2026-08-11, proveedor CxP y carrera concurrente

- El candidato `6bed2315` corrige la derivación servidor del proveedor CxP y
  aprobó en staging el alta oficial, replay idempotente y bloqueo de sobrepago
  bajo dos solicitudes simultáneas. El detalle final quedó pagado sin saldo.
- P110-002 queda **parcial** hasta la conciliación independiente, recuperación
  outbox, reportes con diferencia cero y UAT firmado del contador. Implementación
  P110: **19,2 %** (cinco fases parciales de 13); certificación y preparación
  productiva: **0 %**; veredicto **NO-GO**.

## Actualización 2026-08-11, visual de finanzas y observabilidad

- La revisión visual autenticada confirma que CxP y sus abonos se muestran en
  filas/columnas con totales y estados, sin errores de consola. La conciliación
  exhibe pendientes contables que requieren revisión independiente.
- P110-009 queda bloqueada por capacidad/alertas visibles y porque Alertmanager
  no tiene un receptor externo aprobado. El avance formal no cambia: **19,2 %**
  de implementación, **0 %** de certificación y **NO-GO**.

## Actualización 2026-08-11, capacidad y DIAN de lectura

- Se redujo el uso de disco de staging de 88 % a 60 % recuperando solamente
  caché de compilación y volúmenes anónimos huérfanos. Los servicios aislados y
  sus endpoints de salud se mantuvieron correctos.
- DIAN respondió en línea para estado de conexión, sin reintentos y con
  reconciliación disponible. Falta evidencia fiscal oficial; P110-006 continúa
  pendiente y P110-009 no cierra sin carga, SLO y alerta externa. El avance se
  mantiene en **19,2 %** y el veredicto es **NO-GO**.

## Actualización 2026-08-11, saneamiento de eventos contables

- El candidato `21a4f2cf` clasificó y procesó los siete hitos operativos que
  no requerían asiento. La conciliación de agosto quedó sin pendientes, sin
  errores y con diferencia monetaria cero.
- P110-002 continúa **parcial** hasta la aceptación independiente del contador.
  El avance formal se mantiene en **19,2 %**, certificación **0 %** y **NO-GO**.

## Actualización 2026-08-11, asistente IA no operativo

- El asistente IA respondió una consulta neutra en PCS con modo agente
  desactivado y sin efectos operativos. Se observó un error aislado de consola
  pendiente de atribución, por lo que no se presenta como prueba limpia total.
- P110-003 queda **parcial**; faltan todos los botones y escenarios IA. La
  implementación sube a **23,1 %** (seis fases parciales de 13), certificación
  **0 %** y veredicto **NO-GO**.

## Actualización 2026-08-11, paridad DIAN PCS

- La configuración DIAN completa de PCS existe en el entorno principal, pero
  la réplica aislada de staging no contiene su fila ni referencias de firma.
  La comprobación fue de solo lectura y no expuso ni transfirió secretos.
- P110-006 sigue pendiente: antes de emitir una factura de prueba se debe
  restaurar la configuración y el material de firma por una vía segura y
  acotada, confirmar que staging no apunte a un ambiente fiscal no autorizado
  y ejecutar el flujo oficial mínimo. El avance no cambia: **23,1 %**,
  certificación **0 %**, **NO-GO**.

## Actualización 2026-08-11, impresión y responsive

- La regresión reproducible aprobó 20/20 formatos Carta/POS y la factura
  extensa se inspeccionó visualmente sin recortes de filas, columnas, total ni
  QR. Finanzas PCS/staging mantuvo controles etiquetados y visibles a 390 px.
- P110-005 pasa a **parcial**; restan documentos reales, tableta, lector/zoom
  e impresión física. El avance formal sube a **26,9 %** (siete fases parciales
  de 13), certificación **0 %** y veredicto **NO-GO**.

## Actualización 2026-08-12, carga y barrido crítico

- La carga autenticada de lectura de PCS/staging aprobó con 240 solicitudes,
  concurrencia 12, p95 306 ms, p99 674 ms y cero errores. El barrido crítico
  web aprobó 10/10 vistas entre escritorio/móvil, y los contratos de domótica
  aprobaron sin ejecutar pulsos físicos.
- P110-004 y P110-009 pasan a **parcial**. Permanecen pendientes las acciones
  mutantes por rol, cuatro cajas, hardware GPIO, alerta externa y simulacro.
  El avance formal sube a **34,6 %** (nueve fases parciales de 13),
  certificación **0 %**, **NO-GO**.

## Actualización 2026-08-12, correo corporativo

- SPF, DKIM y DMARC de correo corporativo están publicados y Mailu está activo.
  BIMI conserva el logo, pero sin VMC/CMC no garantiza el avatar externo del
  remitente. No se enviaron correos durante esta comprobación.
- P110-006 pasa a **parcial** por la evidencia de correo y la validación de
  firma del principal. Sigue bloqueada la emisión DIAN en staging y faltan los
  flujos reales de entrega. El avance formal sube a **38,5 %** (diez fases
  parciales de 13), certificación **0 %**, **NO-GO**.

## Actualización 2026-08-12, réplica DIAN aislada en staging

- La configuración DIAN de PCS y su par de firma se replicaron de manera
  acotada al staging de la empresa 12. La auditoría independiente confirmó la
  paridad requerida y las referencias legibles sin revelar secretos.
- Staging quedó forzado a habilitación, sin emisión local, pendiente y con
  consecutivo cero. No se emitió factura ni se alteró producción. P110-006
  permanece **parcial** hasta diagnosticar bajo sesión y obtener acuses oficiales
  DIAN; el avance formal se mantiene en **38,5 %**, certificación **0 %** y
  veredicto **NO-GO**.
- La revisión visual autenticada confirmó esa configuración y validó las
  credenciales sin emitir. El botón de diagnóstico no devolvió un resultado
  adicional visible, por lo que no se infiere éxito fiscal ni cambia el avance.
- La compuerta integrada volvió a aprobar salud y paridad DIAN; el bloqueo
  automático restante es Alertmanager sin canal externo verificable. Se mantiene
  **38,5 %**, certificación **0 %** y **NO-GO**.
- Se emitió una única factura de habilitación controlada y DIAN informó que el
  set estaba aceptado. Se corrigió en el candidato una clasificación transitoria
  de esa respuesta positiva; hasta desplegarla y reconciliar los artefactos no
  cambia el avance ni el veredicto.
- Las regresiones separadas de permisos, tenant, CxP, IA, firma DIAN y correo
  aprobaron. Son evidencia técnica de servidor, no sustituyen usuarios reales
  por rol ni elevan el porcentaje formal.
- El candidato inmutable `1a6dc4fa` aprobó CI/SBOM y quedó activo en staging.
  La reconsulta oficial del acuse DIAN persistió como aceptada sin emisión
  adicional. P110-008 continúa parcial por el drill final y la aceptación de
  RPO/RTO; el avance formal se mantiene en **38,5 %** y **NO-GO**.
- El restore aislado del mismo candidato aprobó migrador sin DDL, health/ready,
  inventario privado sin huérfanos y limpieza total con RTO observado de 119 s.
  Siguen pendientes la réplica autenticada, rollback coordinado y firma de
  objetivos RPO/RTO; no cambia el porcentaje formal.
- El candidato actual `e308ca4b` repitió el restore con dos réplicas
  autenticadas, transferencia A/B por SHA-256, negativos de archivos y rollback
  coordinado de ambas bases más almacenamiento privado. El ensayo se limpió por
  completo y observó 99 s de rollback y 226 s de RTO total. P110-008 conserva
  estado parcial: faltan congelación final y aceptación humana de RPO/RTO; el
  avance formal sigue en **38,5 %**, certificación **0 %**, **NO-GO**.
- La recarga visual autenticada del Centro DIAN confirmó estado `aceptado` en
  habilitación y producción local sin activar. Esta confirmación no autoriza
  emisión fiscal productiva ni promoción.

## Actualización 2026-08-12, entrega externa de Alertmanager

- Alertmanager usa ahora un relay Mailu interno mediante plantilla versionada y
  configuración privada generada en el VPS. Se corrigió el permiso de lectura
  del runtime y el despliegue espera su endpoint de salud antes de finalizar.
- Un único aviso sintético autorizado fue aceptado por el relay externo y su
  alerta se resolvió y desapareció del API. Falta la confirmación visual de
  recepción/resolución, deduplicación y la guardia operativa; P110-009 queda
  **parcial**. La evidencia no sustituye pruebas de cajas, UAT contable,
  impresión física, réplica/rollback ni piloto.
- El avance formal se mantiene en **38,5 %**, la certificación en **0 %** y el
  veredicto es **NO-GO**.

## Actualización 2026-08-12, antivirus obligatorio en el candidato

- Una caída controlada reveló que el backend previo de staging no llevaba la
  configuración ClamAV obligatoria. Se promovió el candidato inmutable con su
  overlay antivirus y se repitió la prueba visual: con ClamAV detenido la carga
  limpia recibe indisponibilidad y no persiste el soporte; después se recuperó
  el escáner y el endpoint de staging.
- La sonda EICAR no alcanzó el navegador porque la protección local la eliminó
  antes de enviarla. Posteriormente se ejecutó desde el VPS de staging por la
  ruta HTTP autenticada: ClamAV la rechazó con `422` antes de persistencia. Se
  enviaron a papelera los soportes QA creados durante intentos previos, mediante
  la API auditada. Siguen pendientes recuperación, concurrencia y alertas P0.
  P110-007 permanece **parcial**; el avance formal
  continúa en **38,5 %**, certificación **0 %**, **NO-GO**.

## Actualización 2026-08-12, controles de seguridad y observabilidad

- Los contratos estáticos de cookies, sesión, CORS, scope empresarial, secretos
  runtime, SLO y observabilidad aprobaron. En staging, las rutas empresariales
  anónimas y un preflight externo devolvieron `401`; la compuerta automática
  confirmó salud, paridad DIAN saneada y Alertmanager sin alertas activas.
- La CSP conserva `unsafe-inline` como hallazgo verificable. El control SSH del
  VPS se completó con `fail2ban` activo y una cárcel conservadora para el puerto
  real, sin modificar SSH/UFW ni afectar staging. P110-007 y P110-009 continúan
  parciales; avance formal **38,5 %**, certificación **0 %**, **NO-GO**.

## Actualización 2026-08-12, migración vacía y upgrade aislado

- El migrador de `e308ca4b` aprobó base vacía, segunda pasada, checksum drift
  fallido/corregido y upgrade de una copia lógica de staging sin cambiar el
  número de tablas. La corrección quedó limitada al runner efímero, que ahora
  monta almacenamiento privado obligatorio.
- P110-001 se mantiene parcial por aceptación de ADR/rollback futuro; el
  avance formal sigue en **38,5 %**, certificación **0 %**, **NO-GO**.
- El mismo candidato aprobó además tres negativos previos a migración, cinco
  comprobaciones de rollback transaccional y cuatro de compatibilidad hacia
  atrás con una API anterior. Se conserva la condición de parcial hasta la
  aceptación de los responsables; el avance formal no cambia.

## Actualización 2026-08-12, CxP sobre el candidato actual

- El esquema atómico `empresa_cxp_pagos` quedó comprobado en staging y las
  pruebas focalizadas aprobaron la idempotencia, bloqueo por empresa,
  reconciliación solo lectura y protección de la tabla histórica.
- No se infiere cierre financiero: siguen pendientes conciliación independiente,
  recuperación elegible y UAT firmado del contador. El avance formal permanece
  en **38,5 %**, certificación **0 %**, **NO-GO**.

## Actualización 2026-08-12, contratos IA del candidato

- Aprobaron los contratos de extracción IA, edición humana, proveedor canónico,
  doble clic, estado cerrado, papelera y aislamiento de memoria; `go vet` de
  handlers/db también aprobó. La IA no puede convertir una lectura en CxP sin
  confirmación humana independiente.
- Aún falta ejecutar todos los botones por rol y comprobar visualmente timeout,
  cancelación y proveedor real. P110-003 continúa **parcial**, sin cambio de
  avance formal: **38,5 %**, certificación **0 %**, **NO-GO**.

## Actualización 2026-08-12, carga y compuerta automática

- La compuerta automática actualizada aprobó salud, paridad DIAN y entrega
  externa Alertmanager en staging. La alerta de malware activa corresponde a
  la sonda EICAR controlada y se resolvió naturalmente; ClamAV y staging se
  encuentran saludables y no se suprimió la alerta manualmente.
- La carga autenticada de 300 GET a concurrencia 15 aprobó con cero errores,
  p95 de 907 ms y p99 de 1890 ms, sin mutaciones. Faltan cajas mutantes,
  recursos/SLO, resolución y deduplicación completa de alertas. P110-009 sigue
  **parcial**; avance formal **38,5 %**, certificación **0 %**, **NO-GO**.

## Actualización 2026-08-12, candidato IA corregido y rollback completo

- La prueba autenticada descubrió que los IDs canónicos de algunas funciones
  del Centro IA se normalizaban erróneamente a diagnóstico. El candidato
  `d3d21414` corrigió el selector, aprobó Go completo, CI, escaneo, SBOM y
  promoción exclusiva a staging. Las siete funciones del catálogo se repitieron
  visualmente como recomendaciones/borradores, sin mutaciones de negocio.
- El mismo digest aprobó restore aislado con dos réplicas autenticadas, cinco
  dominios protegidos, negativos de archivos, rollback coordinado de ambas
  bases y almacenamiento privado, y cleanup total. RTO total: 223 s; RPO
  medido: 14.611 s, aún pendiente de aceptación humana.
- Alertmanager consolidó dos publicaciones sintéticas idénticas en una alerta y
  la resolución la eliminó. Quedan pendientes confirmación visible del receptor,
  guardia y simulacro integral. No se añade una fase parcial nueva: el avance
  formal permanece **38,5 %**, certificación **0 %** y **NO-GO** hasta completar
  roles/cajas, UAT contable, impresión física, piloto y firmas.

## Actualización 2026-08-12, auditorías estáticas del candidato actual

- Las auditorías profesionales, de seguridad, observabilidad y SLO aprobaron
  sobre el checkout del candidato `d3d21414`. Cubren sintaxis web, contratos de
  sesión y scope empresarial, observabilidad de worker/cola y definición de
  objetivos operativos.
- La ejecución amplia de Go se agotó localmente sin un resultado concluyente y
  no se contabiliza como aprobación. Persisten DAST integral, CSP, roles/cajas,
  simulacro y aceptaciones humanas. El avance formal continúa en **38,5 %**,
  certificación **0 %** y veredicto **NO-GO**.

## Actualización 2026-08-12, lectura operativa de outbox

- El staging del candidato `d3d21414` conserva backend, worker, frontend y
  ClamAV saludables, sin reinicios ni errores recientes del worker. Heartbeat,
  jobs en espera/proceso y leases vencidos permanecieron en cero.
- La métrica empresarial reporta dos outbox `dead`. No se reintentaron a ciegas:
  requieren previsualización por superadministración y conciliación CxP antes de
  cualquier efecto. P110-009 continúa **parcial**, avance **38,5 %**,
  certificación **0 %**, **NO-GO**.

## Actualización 2026-08-12, barrido autenticado crítico

- Finanzas/CxP, Centro IA, pruebas DIAN y domótica aprobaron 8/8 en escritorio
  y móvil bajo una sesión PCS de staging, sin errores visibles ni mutaciones.
  Se corrigió el inventario: CxP está integrado en Finanzas y no posee una
  pantalla independiente con el nombre usado inicialmente por el auditor.
- El barrido omitió acciones de negocio por diseño. P110-004 no cierra hasta
  completar roles, cuatro cajas, hardware y evidencias humanas. El avance formal
  conserva **38,5 %**, certificación **0 %**, **NO-GO**.

## Actualización 2026-08-12, barrido administrativo autenticado

- El lote inicial de 12 pantallas administrativas aprobó 24/24 vistas entre
  escritorio y móvil. Los abortos de recursos del menú de productos ocurrieron
  durante navegación interna y su pantalla final aprobó en ambos formatos.
- Es cobertura visual segura, no validación mutante ni de roles. El avance
  formal permanece **38,5 %**, certificación **0 %**, **NO-GO**.

## Actualización 2026-08-12, segundo lote administrativo

- Doce pantallas adicionales aprobaron 24/24 vistas entre escritorio y móvil,
  incluidos carritos y domótica, sin mutaciones ni errores. P110-004 conserva
  estado parcial; el avance formal sigue en **38,5 %**, **NO-GO**.

## Actualización 2026-08-12, tercer lote administrativo

- El lote de IA, tareas, fiscal, cobranza y comisiones aprobó 24/24 vistas sin
  errores ni mutaciones. P110-004 se mantiene parcial; avance formal **38,5 %**,
  certificación **0 %**, **NO-GO**.

## Actualización 2026-08-12, recuperación auditada CxP y cuarto lote

- La recuperación CxP de PCS previsualizó un único dead elegible, verificó la
  barrera CSRF y reactivó exactamente un evento por el flujo auditado. La vista
  posterior quedó en cero y el worker continuó sano. Un dead agregado de otro
  alcance permanece fuera de la empresa y no se alteró.
- Compras y configuraciones aprobaron 24/24 vistas en escritorio/móvil sin
  mutaciones. Persisten conciliación independiente, roles, cajas, simulacro y
  aprobaciones humanas; avance formal **38,5 %**, certificación **0 %**,
  **NO-GO**.

## Actualización 2026-08-12, lotes de permisos y contabilidad

- Permisos, rol cajero, correo, tributario, contabilidad, contratos, domótica,
  corte de caja y créditos aprobaron 48/48 vistas en ambos formatos sin
  mutaciones. La cobertura visual no sustituye pruebas operativas por rol.
- El avance formal permanece **38,5 %**, certificación **0 %**, **NO-GO**.

## Actualización 2026-08-12, lotes CRM y fiscal visual

- CRM, documentos, egresos, correo, estaciones, facturación electrónica y
  finanzas aprobaron 48/48 vistas en escritorio/móvil sin mutaciones. Las
  acciones fiscales se omitieron expresamente por ser efectos reales.
- El avance formal permanece **38,5 %**, certificación **0 %**, **NO-GO**.

## Actualización 2026-08-12, lotes documental e inventario

- Gestión documental, horarios, importaciones, impuestos, inventario, licencia,
  WMS, Nextcloud, NIIF y nómina aprobaron 48/48 vistas en ambos formatos sin
  mutaciones. El avance formal se mantiene en **38,5 %**, **NO-GO**.

## Actualización 2026-08-12, lotes nómina, reportes y contador

- Nómina, portal/suite contador, producción, bodegas, proveedores, reportes,
  soporte y soportes IA aprobaron 64/64 vistas en escritorio y móvil sin
  mutaciones. El avance formal conserva **38,5 %**, certificación **0 %**,
  **NO-GO**.

## Actualización 2026-08-12, barrido operativo y contratos de aislamiento

- Tarifas, tesorería, turnos, domótica, caja, venta y superadministración
  aprobaron 32/32 vistas en ambos formatos sin mutaciones. También aprobaron
  contratos focalizados de CxP, tenant, permisos, roles y archivos privados.
- No sustituye usuarios reales, cajas simultáneas ni hardware; avance formal
  **38,5 %**, certificación **0 %**, **NO-GO**.

## Actualización 2026-08-12, barrido de superadministración

- El panel superadministrador aprobó 32/32 vistas de escritorio y móvil,
  incluidas alertas, auditoría, consumo y configuración, sin mutaciones.
- El avance formal sigue en **38,5 %**, certificación **0 %**, **NO-GO**.

## Actualización 2026-08-12, compuerta rs e inventarios generados

- `rs` ejecutó el preflight y se detuvo antes de sincronizar porque la rama
  `codex/p110-execution` no es `main`; no se desplegó staging ni producción.
- El preflight reveló inventarios generados desactualizados. Se regeneraron y
  aprobaron 155 funciones `Ensure*`, 204 rutas multiempresa y 106 llamadas
  runtime; los cambios solo actualizan referencias de línea. El avance formal
  sigue en **38,5 %**, certificación **0 %**, **NO-GO**.
- La repetición de `rs` aprobó todas sus auditorías y volvió a detenerse antes
  de sincronizar por la protección de rama: `codex/p110-execution` no es
  `main`. No es un fallo de staging ni autoriza desplegar sin integración
  aprobada.

## Actualización 2026-08-12, despliegue Domótica y restore repetible

- El acceso directo opcional a Domótica y su actualización inmediata entre
  iframes fueron fusionados en `main`, desplegados con `rs` y verificados en
  PCS: el enlace `Venta directa` cambió a equipos sin recarga y volvió al
  carrito al desactivar el check. `/health` respondió 200 y Nginx validó.
- El restore aislado seleccionó automáticamente el último snapshot completo,
  aprobó health/ready, migración, dos bases, inventario privado sin huérfanos y
  limpieza total. RTO observado 78 s y RPO observado 59.854 s.
- P110-008 conserva estado **parcial**: falta repetir réplica autenticada y
  rollback sobre el próximo digest final y obtener aceptación de RPO/RTO. El
  avance formal continúa en **38,5 %**, certificación **0 %**, **NO-GO**.

## Actualización 2026-08-12, adjunto CxP/IA real con revisión humana

- El candidato `945d1751` radicó un XML sintético limpio, extrajo por IA
  proveedor, NIT, fechas, subtotal, IVA y total, permitió corregir los campos
  leídos y registró la edición en auditoría.
- El soporte `SCI-0028` se rechazó sin contabilizar y la búsqueda posterior del
  documento en CxP no encontró registros. La IA no creó obligación, asiento,
  pago ni movimiento de inventario.
- P110-003 permanece **parcial** por timeout/cancelación deliberada, proveedor
  caído, todos los botones/roles, evals y A/B. El avance formal continúa en
  **38,5 %**, certificación **0 %**, **NO-GO**.

## Actualización 2026-08-12, factura y anulación DIAN reales

- PCS cerró una venta real de `menta` por `COP 100`. DIAN aceptó la factura
  `1PCS7` con acuse sincrónico y mensaje `Procesado Correctamente.`.
- La anulación oficial generó la nota crédito total `NC12000000113`; DIAN
  también la aceptó con `Procesado Correctamente.`. La factura quedó anulada y
  la nota crédito emitida, sin reenvíos ni duplicados.
- P110-006 permanece **parcial** por XML/PDF/CUFE/QR completos, impresión
  física, empresa B, idempotencia por doble clic y demás integraciones. El
  avance formal se mantiene en **38,5 %**, certificación **0 %**, **NO-GO**.
- Por instrucción del usuario no se ejecutó `rs` en este bloque.

## Actualización 2026-08-12, impresión virtual y revisión visual

- El candidato `945d1751` generó 20/20 formatos Carta/POS sin casos a revisar
  ni fallos de autoimpresión. Factura y recibo extensos conservaron 96 filas en
  cinco páginas de detalle y una página final de resumen.
- La revisión humana confirmó filas, columnas, importes, encabezados repetidos,
  totales, firmas y QR sin recortes en factura Carta/POS y documentos extensos.
- P110-005 permanece **parcial** hasta completar impresión física, tableta,
  accesibilidad asistida y documentos PCS reales. El avance formal permanece
  en **38,5 %**, certificación **0 %**, **NO-GO**.
- Por instrucción del usuario no se ejecutó `rs` en este bloque.

## Actualización 2026-08-12, matriz IA y cancelación segura

- En PCS/staging el asistente central respondió una consulta neutra y el modo
  agente presentó una propuesta con confirmación obligatoria. Se canceló y el
  catálogo confirmó que no se creó el producto.
- Compras, Egresos e Ingresos bloquearon correctamente la IA sin adjunto. El
  inventario reproducible localizó 20 controles IA empresariales; Grafología
  quedó inconclusa por un `alert` bloqueante y no se contabiliza como aprobada
  en staging. Se corrigió localmente para mostrar guardas y errores en su panel.
- Se corrigió localmente el botón IA de Productos, que enviaba el mensaje al
  iframe en vez del shell superior, y se añadió timeout de 90 segundos con
  cancelación, reembolso de cuota y métricas separadas para soportes IA.
- Finanzas/CxP inicia ahora el borrador editable sin `window.confirm`: cambia
  visiblemente el tipo a CxP, explica la revisión pendiente y conserva el
  guardado como confirmación humana independiente.
- Aprobaron `go test ./...`, `go vet` focalizado, auditoría profesional,
  inventario IA, comprobación del diff y escaneo saneado de archivos cambiados.
- Falta desplegar y repetir estas correcciones, cubrir los 20 controles por rol,
  proveedor caído y evals A/B. P110-003 sigue **parcial**; avance formal
  **38,5 %**, certificación **0 %**, **NO-GO**. No se ejecutó `rs`.

## Actualización 2026-08-12, reconciliación Mailu y CSP de observación

- La página real de correo corporativo mostró un buzón pendiente, sin no leídos
  verificables. La corrección local solo permite marcarlo provisionado después
  de autenticar INBOX a través del front Mailu; el botón usa POST con CSRF y
  conserva el estado anterior ante fallo.
- El staging público conserva cabeceras HTTPS y el frontend ya observa una CSP
  estricta. El backend aún heredaba `unsafe-inline`; el Compose local activa por
  defecto la CSP Report-Only estricta sin modificar todavía la política aplicada.
- Ambas correcciones requieren despliegue y repetición visual. P110-006 y
  P110-007 continúan **parciales**; avance formal **38,5 %**, certificación
  **0 %**, **NO-GO**. No se ejecutó `rs`.

## Actualización 2026-08-12, compuerta CSP y primer módulo sin inline

- Se midieron 1.347 bloqueos CSP en 243 páginas. Grafología externalizó su
  bloque de estilos y once atributos, quedó en cero y redujo el total a 1.335
  bloqueos en 242 páginas; el área empresarial bajó de 975 a 963.
- La nueva línea base falla en CI ante cualquier aumento global o por área y si
  el archivo de referencia desaparece. Sus valores solo pueden disminuir.
- La revisión visual local confirmó hoja CSS cargada, dos columnas en escritorio,
  una columna/KPI 2x2 en móvil, cero errores de consola y cero overflow horizontal.
- La deuda restante impide retirar `unsafe-inline` de la política aplicada;
  P110-007 continúa **parcial**, avance formal **38,5 %**, certificación **0 %**,
  **NO-GO**. No se ejecutó `rs`.

## Actualización 2026-08-12, eventos inline eliminados globalmente

- Las nueve superficies que conservaban atributos `on*` migraron sus 19 eventos
  a listeners explicitos, delegados o al helper comun de impresion e imagenes.
  El inventario completo de `web/**/*.html` queda en cero eventos inline.
- La deuda CSP baja de 1.335 a 1.316 bloqueos: 948 empresariales, 216 de
  superadministracion y 138 publicos. La linea base de CI fue reducida para que
  cualquier reintroduccion falle automaticamente.
- Aprobaron sintaxis JavaScript/HTML, `go test ./...`, el preflight estricto,
  Docker Compose y las auditorias de seguridad, permisos, UX, migraciones y
  contratos. El inventario runtime se regenero tras detectar referencias de
  linea desactualizadas, sin cambiar logica de negocio.
- Todavia restan scripts, bloques `style` y atributos `style`; por eso no se
  aplica una CSP sin `unsafe-inline`. P110-007 sigue **parcial**, avance formal
  **38,5 %**, certificacion **0 %**, **NO-GO**. No se ejecuto `rs`.

## Actualización 2026-08-12, CSS compartido en configuración empresarial

- Seis accesos de configuración que duplicaban el mismo bloque `style` usan
  ahora una hoja externa compartida: Cobro operativo, Formato monetario,
  Pasarelas de pago, Perfil tributario Colombia, Productos y pedidos y Respaldo.
- La deuda CSP baja de 1.316 a 1.310 bloqueos y de 242 a 236 páginas afectadas;
  el área empresarial queda en 942 bloqueos sobre 140 páginas.
- La revisión visual local comprobó las seis rutas, CSS cargado, iframe presente,
  encabezado responsive y cero desborde horizontal. El contrato de recursos
  estáticos, `go test ./...`, `go vet ./...` y el preflight estricto aprobaron;
  la línea base decreciente fue actualizada.
- P110-007 continúa **parcial**, avance formal **38,5 %**, certificación **0 %**,
  **NO-GO**. No se ejecutó `rs`.

## Actualización 2026-08-12, límites de tiempo en CI profesional

- Una corrida de la PR quedó congelada después de aprobar preflight, tests
  normales y secretos, específicamente en `go test -race` y build de imágenes;
  se canceló y relanzó sin interpretar el bloqueo como fallo funcional.
- El detector de carreras tiene ahora límite de doce minutos y timeout Go de
  diez; la construcción de imágenes queda limitada a veinte minutos. El job
  conserva además su límite global de 45 minutos.
- La repetición Linux sigue siendo obligatoria porque el Go local de Windows no
  dispone de CGO para `-race`. No cambia el avance formal: **38,5 %**,
  certificación **0 %**, **NO-GO**. No se ejecutó `rs`.

## Actualización 2026-08-13, cuatro módulos empresariales sin inline

- Tesorería, Proveedores de Firma Digital, Soporte remoto y Tutorial de
  Domótica externalizaron sus bloques visuales completos y quedan con cero
  bloqueos del inventario CSP por página.
- La deuda baja de 1.310 a 1.306 bloqueos y de 236 a 232 páginas; el área
  empresarial queda en 938 bloqueos sobre 136 páginas.
- Las cuatro rutas aprobaron contrato de recursos y revisión local de escritorio
  y móvil: CSS cargado, grid conservado, cero overflow y consola limpia. No se
  ejecutaron acciones mutantes. También aprobaron `go test ./...`, `go vet ./...`
  y el preflight estricto completo.
- P110-007 continúa **parcial**, avance formal **38,5 %**, certificación **0 %**,
  **NO-GO**. No se ejecutó `rs`.

## Actualización 2026-08-13, lote CSS de 23 páginas

- El shell empresarial, CRM, WMS, MRP, Ayuda de APIs, conceptos de marca,
  auditorías, seguridad y configuración del superadministrador externalizaron
  sus únicos bloques style; los trece accesos equivalentes comparten una
  hoja y Nextcloud conserva una específica.
- La deuda baja de 1.306 a 1.283 bloqueos y de 232 a 209 páginas; los bloques
  style bajan de 177 a 154. Enlaces CSS, go test ./..., go vet ./...
  y preflight estricto completo aprobaron.
- La repetición visual desplegada continúa pendiente: el navegador interno
  bloqueó file: y el servidor auxiliar local no escuchó. No se retiró
  unsafe-inline.
- P110-007 continúa **parcial**; avance formal **38,5 %**, certificación **0 %**,
  **NO-GO**. No se ejecutó rs.

## Actualización 2026-08-13, scripts externos en 14 páginas

- Catorce páginas cuyo único bloqueo era un script inline usan ahora un
  controlador externo en el mismo punto de ejecución; no se cambiaron contratos
  de API, permisos ni datos.
- La comparación con HEAD y node --check aprobaron 14/14. La deuda baja de
  1.283 a 1.269 bloqueos y de 209 a 195 páginas; quedan 209 scripts inline.
- La repetición visual y autenticada del candidato desplegado sigue pendiente.
  P110-007 continúa **parcial**; avance formal **38,5 %**, certificación **0 %**,
  **NO-GO**. No se ejecutó rs.

## Actualización 2026-08-13, auditoría integral de duplicación y calidad

- El barrido estático cubrió 638 archivos Go y 421 recursos HTML/JS/CSS. Los
  auditores de rutas multiempresa, permisos/licencias, seguridad, plantillas y
  consistencia UX aprobaron; no aparecieron rutas empresariales duplicadas,
  archivos frontend idénticos ni IDs DOM repetidos reales.
- Se incorporó P110-001A para cerrar 106 llamadas `Ensure*` fuera del migrador,
  52 grupos de funciones idénticas, 124 funciones de más de 200 líneas, acceso
  DB sin contexto, errores descartados, repetición visual y contratos legados.
- El intento de cobertura completa superó 60 segundos y quedó inconcluso; el
  plan exige una línea base acotada antes de refactorizar handlers críticos.
- La auditoría amplía el trabajo pendiente y no certifica código ni candidato.
  El avance formal continúa en **38,5 %**, certificación **0 %**, **NO-GO**. No
  se ejecutó `rs` ni se realizaron pruebas reales PCS.

## Actualización 2026-08-13, compuerta estructural y cancelación IA

- Una herramienta Go sin dependencias fija máximos decrecientes para funciones
  extensas, duplicación exacta, contexto DB y resultados descartados. Preflight
  y CI fallan ante cualquier aumento y publican evidencia JSON.
- Doce flujos HTTP IA propagan `r.Context()` y los adjuntos/documentos también
  conservan cancelación hasta OpenAI Responses, Chat Completions y Gemini. Tres
  pruebas con servidor real de prueba confirmaron que cancelar corta el request.
- Domótica ya no sincroniza la Raspberry principal desde la lectura resumen;
  la compatibilidad se ejecuta al guardar configuración y el inventario de
  llamadas `Ensure*` fuera del migrador baja de 106 a 104.
- P110-001A continúa parcial: la compuerta evita regresión, pero la deuda base
  todavía debe reducirse y clasificarse. Avance formal **38,5 %**,
  certificación **0 %**, **NO-GO**. No se ejecutó `rs`.

## Actualización 2026-08-13, DDL retirado de preconfiguración HTTP

- La aplicación y limpieza de plantillas empresariales reemplazó diez llamadas
  `Ensure*` por verificadores de esquema ya migrado para productos, usuarios,
  configuración operativa, comisiones, tarifas por minuto/día/motel y Domótica.
  Los verificadores no crean ni alteran objetos y fallan cerrados si falta el
  contrato; toda escritura conserva su `empresa_id` recibido del flujo.
- Los errores de las escrituras de limpieza dejan de descartarse. Una prueba
  estructural impide reintroducir cualquiera de esas llamadas DDL en el handler.
  Una segunda extracción aplica readiness a eventos contables, permisos finos,
  Nextcloud, usuarios del correo masivo, Rappi y permisos por rol. El inventario
  Alertas y agentes de mantenimiento también validan el contrato migrado sin
  DDL en worker/handler. Correo empresarial deja de aprovisionar cuentas desde
  GET; Nextcloud, plantillas/licencias y voz distinguen sincronización de
  migración. El total baja de 104 a **72** y el subconjunto HTTP de 29 a **0**.
- El fallo remoto inicial de `gosec` en el nuevo auditor se corrigió reduciendo
  permisos del reporte a `0600`, del directorio a `0750` y documentando la ruta
  del baseline como argumento local confiable de CI.
- P110-001A sigue **parcial** porque restan 69 llamadas del bootstrap legado
  exclusivo de migración y 3 del migrador dedicado, además de las otras deudas
  medidas. Avance formal **38,5 %**, certificación
  **0 %**, **NO-GO**. No se usó PCS, no se desplegó y no se ejecutó `rs`.

## Actualización 2026-08-13, cero bootstrap Ensure en HTTP

- El inventario reproducible confirma **0** llamadas `Ensure*` en handlers, 69
  llamadas del bootstrap legado que producción permite solo al rol `migrate`
  mediante decisión explícita y 3 del migrador dedicado. La compuerta ahora
  falla si reaparece una llamada `Ensure*` en tráfico HTTP.
- Los aprovisionamientos idempotentes conservan su intención y aislamiento:
  alta/sincronización de correo, asignación Nextcloud, catálogo comercial y
  token de voz. Consultar correo con GET es estrictamente lectura y una prueba
  estructural protege ese contrato.
- La fase no se marca terminada: falta trasladar las 69 inicializaciones del
  puente legado al catálogo inmutable y repetir base vacía/upgrade; las 3 del
  migrador son autoridad permitida. Avance formal
  **38,5 %**, certificación **0 %**, **NO-GO**. No se ejecutó `rs`.
- La compuerta `govulncheck` había encontrado vulnerabilidades alcanzables de la
  biblioteca estándar en Go 1.25.12. Ese baseline quedó superado por la
  actualización a Go **1.26.6** registrada en la actualización P110-001A del
  2026-08-14.
- Validación local con temporales/caché en D: aprobó pruebas, vet y
  `govulncheck ./...` con **0 vulnerabilidades alcanzables**. C: continúa
  prácticamente lleno; no se eliminó contenido al bloquearse la limpieza.

## Actualización 2026-08-13, primer saneamiento de duplicación DB

- Se consolidaron tres familias idénticas: verificación de columnas, consulta
  de índices PostgreSQL y normalización de paginación. Once repositorios
  reutilizan helpers internos sin cambiar consultas de negocio, filtros
  `empresa_id`, permisos, endpoints ni autoridad DDL.
- La compuerta estructural baja de **52 a 49** grupos duplicados exactos y la
  línea base se reduce para impedir regresión. Las pruebas DB enfocadas y vet
  aprobaron.
- La comprobación visual autenticada del entorno publicado abrió PCS empresa
  12 y confirmó panel, correo corporativo activo y navegación empresarial. Los
  avisos DIAN visibles son históricos; no se reenviaron documentos fallidos.
- Una consulta IA neutra, con modo agente desactivado, respondió sin propuesta
  ni mutación y señaló dos stocks negativos. La API oficial confirmó **2**
  alertas y déficit **19**, mientras la vista principal mostraba ceros porque no
  inicializaba ese resumen. El candidato corrige el bootstrap y lo protege con
  una prueba; falta repetir visualmente después de desplegarlo.
- P110-001A continúa **parcial** porque persisten handlers monolíticos, acceso
  DB sin contexto, errores ignorados y otras duplicaciones. El avance formal
  conserva **38,5 %**, certificación **0 %**, **NO-GO**. No se ejecutó `rs`.

## Actualización 2026-08-13, segundo saneamiento de utilidades DB

- Doce repositorios reutilizan ahora tres reglas comunes para valores no vacíos,
  patrones `LIKE ... ESCAPE '!'` y paginación. Se eliminaron 104 líneas
  repetidas sin modificar SQL, filtros `empresa_id`, permisos ni endpoints.
- La compuerta baja de **49 a 47** grupos duplicados exactos y de 7.197 a 7.186
  funciones productivas. La prueba de caracteres `%`, `_` y `!`, las pruebas DB
  enfocadas y `go vet ./db` aprobaron.
- P110-001A permanece **parcial** por las deudas estructurales todavía medidas.
  El avance formal sigue en **38,5 %**, certificación **0 %**, **NO-GO**. No se
  usó PCS, no se desplegó y no se ejecutó `rs`.

## Actualización 2026-08-13, compatibilidad de utilidades en handlers

- Las superficies de auditoría, permisos y compras reutilizan los helpers
  canónicos de valores no vacíos y números positivos mediante wrappers de
  compatibilidad. No se alteraron rutas, permisos, `empresa_id`, transacciones
  ni contratos de respuesta.
- Las pruebas completas de `go test ./handlers`, `go vet ./handlers` y
  `git diff --check` aprobaron. La métrica global no cambia porque permanecen
  grupos equivalentes en otros dominios; se conserva la línea base sin
  maquillarla.
- P110-001A continúa parcial; el avance formal permanece en **38,5 %**,
  certificación **0 %**, **NO-GO**. No se ejecutó `rs`.

## Actualización 2026-08-13, códigos compartidos de repositorios

- Apartamentos turísticos, domicilios, parqueadero y taxi delegan la misma
  generación de códigos al helper `repositoryCoreCode`; se conserva el nombre
  de cada wrapper de dominio y su formato, incluyendo el fallback temporal.
- Las pruebas DB y `go vet ./db` aprobaron. La compuerta estructural reduce los
  grupos duplicados exactos de **47 a 46** sin tocar SQL, tenant, permisos ni
  contratos.
- P110-001A continúa parcial; el avance formal sigue en **38,5 %**,
  certificación **0 %**, **NO-GO**. No se ejecutó `rs`.

## Actualización 2026-08-14, bloqueo Trivy por Go stdlib

- La PR #170 expuso `CVE-2026-46600` HIGH en la imagen backend por Go
  1.25.13. El escaneo de filesystem y frontend permaneció limpio.
- Se actualizan módulo, Docker y CI a Go **1.26.6**, versión mínima indicada
  por el aviso para corregir la vulnerabilidad. La PR debe repetir Trivy,
  preflight y pruebas antes de considerarse cerrada.
- El Plan 110 continúa **NO-GO** y no se ejecutó `rs`.

## Actualización 2026-08-14, consolidación directa en main

- Con autorización temporal se desactivó únicamente `enforce_admins`; las
  revisiones y checks obligatorios permanecen configurados para los demás.
- Cinco dominios de handlers delegan la selección de texto no vacío al helper
  canónico. Seis normalizadores fiscales/contables reutilizan una única regla
  de período con fallback explícito y pruebas de casos completo, corto y vacío.
- `go test ./db ./handlers`, `go vet ./db ./handlers` y `git diff --check`
  aprobaron. La compuerta reduce duplicaciones exactas de **46 a 43**.
- P110-001A continúa parcial por deuda restante; avance formal **38,5 %**,
  certificación **0 %**, **NO-GO**. No se ejecutó `rs`.

- Segundo lote directo: carnets, contabilidad y domicilios reutilizan el helper
  DB de valor con fallback; configuración guiada y alertas reutilizan el helper
  canónico de handlers. Las pruebas DB/handlers y vet aprobaron, reduciendo la
  compuerta de **43 a 41** grupos duplicados.

## Actualización 2026-08-14, utilidades escalares y pasarelas

- Tarifas reutiliza límites enteros; venta pública/digital comparte moneda;
  descuentos y tarifas comparten día ISO; outbox/jobs comparte hash SHA-256;
  Wompi comparte el mapeo de estado y DIAN reutiliza truncado fiscal.
- Los wrappers de dominio conservan nombres y contratos. Pruebas completas de
  DB/handlers y vet aprobaron; la duplicación exacta baja de **41 a 35**.
- P110-001A continúa parcial; avance formal **38,5 %**, certificación **0 %**,
  **NO-GO**. No se ejecutó `rs`.

## Actualización 2026-08-14, normalizadores de dominios operativos

- Soporte remoto comparte su lectura, configuración IA comparte enteros no
  negativos y DB centraliza estado activo/archivado, porcentajes, moneda,
  identidad de estación y slugs de producción/tesorería.
- Los wrappers conservan contratos, permisos y aislamiento empresarial. Las
  pruebas DB/handlers y vet aprobaron; la duplicación baja de **35 a 27**.
- P110-001A sigue parcial por contextos DB, errores ignorados y handlers
  extensos. Avance formal **38,5 %**, certificación **0 %**, **NO-GO**. No se
  ejecutó `rs`.

## Actualización 2026-08-14, JSON, credenciales y contexto IA

- Finanzas/reportes comparte texto, portales públicos comparten IP, Docker/VPS
  comparte JSON y credenciales opcionales usan un único resolvedor. DB comparte
  JSON map, códigos de gimnasio/odontología, unidades y secciones de contexto IA.
- Las pruebas DB/handlers y vet aprobaron. Duplicación exacta baja de **27 a
  18** y descartes explícitos de **776 a 773** al centralizar decodificación.
- P110-001A continúa parcial; avance formal **38,5 %**, certificación **0 %**,
  **NO-GO**. No se ejecutó `rs`.

## Actualización 2026-08-13, cero duplicaciones exactas

- Se introduce `internal/platform/valueutil` para texto, JSON, fechas, hosts,
  identificadores SQL y límites numéricos; `runtimeconfig` queda como autoridad
  de entorno/DSN y `httpguard` comparte el contrato de salud API/worker.
- Los wrappers existentes preservan contratos de dominio. No cambian endpoints,
  SQL, permisos, secretos ni aislamiento por `empresa_id`.
- `go test ./...`, `go vet ./...` y el auditor estructural aprobaron. La deuda
  de cuerpos exactamente duplicados baja de **18 a 0** (y de **52 a 0** desde
  la línea inicial de P110-001A).
- P110-001A continúa parcial por **1.841** llamadas DB sin contexto, **773**
  resultados descartados explícitamente y handlers extensos. Avance formal
  **38,5 %**, certificación **0 %**, **NO-GO**. No se ejecutó `rs`.

## Actualización 2026-08-13, medición SQL sin falsos positivos HTTP

- El auditor confundía `r.URL.Query()` con `database/sql.Query()`: 1.104 usos
  visibles de parámetros HTTP inflaban la métrica y desviaban la priorización.
- La clasificación excluye únicamente receptores terminados en `.URL.Query` y
  conserva `db.Query`, `tx.Exec`, `QueryRow`, `Prepare` y `Begin` como deuda.
- La prueba del auditor y la medición completa aprueban. La deuda SQL real sin
  contexto queda en **689** llamadas; no se declara cerrada ni se reemplaza por
  `context.Background()` para maquillar la métrica.
- P110-001A continúa parcial; avance formal **38,5 %**, certificación **0 %**,
  **NO-GO**. No se ejecutó `rs`.

## Actualización 2026-08-13, cierre gosec del instalador SSH Domótica

- El instalador SSH atiende errores de `SetDeadline`, cierre del socket y
  liberación/cierre del advisory lock PostgreSQL sin exponer credenciales.
- Ante error de deadline o handshake se devuelve la causa original y se agrega
  la falla de cierre cuando existe; la liberación tardía queda trazada solo con
  `empresa_id` y `raspberry_id`.
- Pruebas SSH/Victron enfocadas y `gosec -include=G104 ./handlers` aprueban con
  **0 hallazgos**. No se ejecutó `rs`.

## Actualización 2026-08-13, cancelación SQL en reportes programados

- Las 17 operaciones SQL de plantillas, programaciones, ejecuciones y
  consistencia usan ahora `QueryContext`, `QueryRowContext` o `ExecContext` con
  el contexto real de la solicitud HTTP.
- Los helpers de búsqueda reciben `context.Context` explícito, por lo que una
  desconexión o timeout puede cancelar PostgreSQL durante todo el flujo.
- `go test ./handlers`, `go vet ./handlers` y el auditor aprueban; la deuda real
  baja de **689 a 672** llamadas sin contexto y la duplicación permanece en 0.
- P110-001A continúa parcial; avance formal **38,5 %**, certificación **0 %**,
  **NO-GO**. No se ejecutó `rs`.

## Actualización 2026-08-13, contextos RRHH e inventario trazable

- Vacaciones RRHH y lotes/series propagan el contexto HTTP a consultas,
  transacción, actualización, inserción y trazabilidad PostgreSQL.
- La inserción del movimiento usa `RETURNING id` en la misma transacción y deja
  de depender de `LastInsertId`, no soportado de forma portable por PostgreSQL.
- La actualización de vigencia y recarga del lote falla de forma explícita; la
  consulta opcional distingue `sql.ErrNoRows` de un error real de base de datos.
- `go test ./handlers`, `go vet ./handlers` y el auditor aprueban. La deuda baja
  de **672 a 662**, duplicación 0. Plan formal **38,5 %**, certificación **0 %**,
  **NO-GO**. No se ejecutó `rs`.

## Actualización 2026-08-13, contexto SQL transversal y embudo de ventas

- Plan de cuentas, cartera, gestión documental, producción y logística pasan el
  contexto real de la solicitud a sus consultas, transacciones y mutaciones.
- La cadena cotización-pedido-documento final y el embudo de conversión propagan
  cancelación hasta PostgreSQL, también cuando el dataset se construye desde
  reportes empresariales, globales, programados o solicitados mediante IA.
- `go test ./handlers`, `go vet ./handlers` y el auditor estructural aprueban; la
  deuda DB sin contexto baja de **662 a 653** y la duplicación exacta sigue en 0.
- P110-001A continúa parcial; avance formal **38,5 %**, certificación **0 %**,
  **NO-GO**. No se ejecutó `rs` ni se creó PR.

## Actualización 2026-08-13, cancelación transaccional del pago CxP

- El handler de pago canónico a proveedores propaga `r.Context()` al
  repositorio y a todas las operaciones SQL de la transacción idempotente,
  incluidos bloqueos `FOR UPDATE`, movimiento financiero, abono y outbox.
- Los helpers de compatibilidad PostgreSQL ya ofrecen variantes Context; los
  wrappers legacy se conservan para jobs sin alterar el contrato existente.
- La serialización del outbox y la recarga posterior al commit fallan de forma
  visible en vez de descartarse. Las regresiones protegen contexto, tenant e
  idempotencia.
- Pruebas DB/handlers, `vet` y auditoría aprueban: **633** llamadas DB sin
  contexto y duplicación exacta 0. P110-001A continúa parcial; avance formal
  **38,5 %**, certificación **0 %**, **NO-GO**. No se ejecutó `rs` ni se creó PR.

## Actualización 2026-08-13, contexto en gobierno de empresas y usuarios

- La vista de impacto previa a desactivar empresas pasa `r.Context()` a la
  introspección y a los conteos por empresa de usuarios, carritos, reservas y
  licencias; una desconexión ya puede cancelar ese análisis administrativo.
- Crear o actualizar un usuario empresarial conserva el contexto al resolver
  el tipo de empresa y validar el rol permitido.
- Pruebas handlers, `vet` y auditoría aprueban: **630** llamadas DB sin contexto
  y duplicación exacta 0. P110-001A continúa parcial; avance formal **38,5 %**,
  certificación **0 %**, **NO-GO**. No se ejecutó `rs` ni se creó PR.

## Actualización 2026-08-13, cero LastInsertId en producción

- Todos los repositorios productivos recuperan identificadores PostgreSQL con
  `RETURNING id`; se retiraron las veinte dependencias restantes de
  `LastInsertId` en empresa, finanzas, inventario, correo, estaciones, hoja de
  vida, venta digital, soporte remoto, GPS y sensores.
- Los UPSERT financieros devuelven su ID en la misma sentencia y ya no hacen
  una segunda consulta de fallback. Refrescos/trazas posteriores a altas dejan
  de ignorar errores.
- Una regresión global impide reintroducir la práctica. Pruebas DB/handlers,
  `vet` y auditoría aprueban: **610** llamadas DB sin contexto, **768**
  descartes explícitos y duplicación exacta 0.
- P110-001A continúa parcial; avance formal **38,5 %**, certificación **0 %**,
  **NO-GO**. No se ejecutó `rs` ni se creó PR.

## Actualización 2026-08-13, PostgreSQL nativo y GET sin mutación

- Las altas de plantillas, programaciones y ejecuciones de reportes,
  configuración/historial de Calculadora y proveedores usan `RETURNING id` y
  ya no dependen de `LastInsertId`, incompatible con el contrato PostgreSQL.
- Calculadora ofrece variantes `Context`, los handlers propagan `r.Context()` y
  una lectura de configuración ausente devuelve defaults sin crear filas.
- El paz y salvo de Créditos falla de forma explícita si no puede conciliar el
  documento o los pagos; Facturación retira su fallback runtime para SQLite.
- Pruebas de regresión y preflight completo estricto protegen PostgreSQL y la
  ausencia de mutación GET. La auditoría queda en **640** llamadas DB sin
  contexto, **771** descartes
  explícitos y 0 duplicaciones. P110-001A continúa parcial; avance formal
  **38,5 %**, certificación **0 %**, **NO-GO**. No se ejecutó `rs` ni se creó PR.

## Actualizacion 2026-08-13, facturacion electronica cancelable y deuda accionable

- Configuracion fiscal por empresa, deteccion de jurisdiccion, reserva
  transaccional de consecutivos y cola de reintentos propagan
  `context.Context` hasta PostgreSQL; se mantienen wrappers compatibles para
  jobs sin request.
- Las rutas principales de Facturacion electronica, Bolsa e Impuestos usan el
  contexto HTTP real. Los listados comprueban `rows.Err()` y los GET de Panama
  y Ecuador dejan de conservar ramas de escritura inalcanzables.
- El auditor estructural ahora ordena los 30 archivos con mayor deuda DB y de
  resultados descartados. Pruebas de DB, handlers y auditor aprueban; la deuda
  DB sin contexto baja de **610 a 597**, descartes **768** y duplicacion 0.
- P110-001A continua parcial; avance formal **38,5 %**, certificacion **0 %**,
  **NO-GO**. No se ejecuto `rs` ni se creo PR.

## Actualizacion 2026-08-13, movimientos financieros cancelables

- Configuracion financiera, CRUD de ingresos/egresos, comprobantes y periodos
  contables propagan `r.Context()` hasta PostgreSQL y conservan aislamiento por
  `empresa_id`.
- Las mutaciones verifican el error de `RowsAffected`; la carga de comprobante
  usa readiness fail-closed en vez de ejecutar `EnsureEmpresaFinanzasSchema`
  durante trafico HTTP.
- Una regresion de contrato protege el uso de variantes `Context` en el handler.
  Pruebas DB/handlers y `vet` aprueban; la deuda DB sin contexto baja de **597 a
  587**, descartes **768** y duplicacion 0.
- P110-001A y P110-002 continuan parciales; avance formal **38,5 %**,
  certificacion **0 %**, **NO-GO**. No se ejecuto `rs` ni se creo PR.

## Actualizacion 2026-08-13, cajas simultaneas cancelables

- Apertura, cupo segun licencia, seleccion de caja abierta por usuario,
  ingresos/egresos atomicos, cierre, reapertura, aprobacion, anulacion y borrado
  conservan contexto hasta PostgreSQL.
- Finanzas, Corte de caja, carrito, venta offline y datafono llaman las variantes
  cancelables sin perder `empresa_id`, usuario, caja, turno ni sucursal.
- El rollback inesperado queda observable y siete mutaciones comprueban
  `RowsAffected`. Pruebas DB/handlers, `vet` y auditor aprueban; la deuda baja de
  **587 a 576** llamadas DB sin contexto y de **768 a 767** descartes explicitos.
- P110-001A, P110-002 y la matriz de cajas de P110-005 continuan parciales hasta
  la prueba autenticada concurrente del candidato. Avance formal **38,5 %**,
  certificacion **0 %**, **NO-GO**. No se ejecuto `rs` ni se creo PR.

## Actualizacion 2026-08-13, tablero financiero sin DDL runtime

- El tablero empresarial reemplaza seis inicializaciones `Ensure*` por readiness
  fail-closed de carritos, clientes, productos, finanzas, eventos y documentos.
- KPI operativos/financieros y estados construidos desde asientos propagan el
  contexto de Finanzas o Reportes hasta PostgreSQL. Una regresion impide
  reintroducir DDL en ese flujo.
- La primera compilacion no se contabilizo porque C: quedo sin espacio. Se
  limpiaron unicamente caches Go reconstruibles, se liberaron aproximadamente
  21 GB y la repeticion con cache temporal en D: aprobo pruebas y `vet`.
- La deuda DB baja de **576 a 575**, descartes **767** y duplicacion 0.
  P110-001A/P110-002 continuan parciales; avance formal **38,5 %**,
  certificacion **0 %**, **NO-GO**. No se ejecuto `rs` ni se creo PR.

## Actualizacion 2026-08-13, alta de soporte remoto cancelable

- Lectura/guardado de configuracion y alta de dispositivos remotos propagan el
  contexto HTTP hasta PostgreSQL; el limite de dispositivos conserva
  `empresa_id` dentro del mismo flujo.
- Si no puede cargarse el plan empresarial, el alta falla cerrada y ya no omite
  silenciosamente el control. Una regresion protege el contrato del handler.
- Pruebas DB/handlers, `vet` y auditor aprueban; deuda DB **575 a 571**,
  descartes **767**, duplicacion 0. P110-001A/P110-007 continuan parciales;
  avance formal **38,5 %**, certificacion **0 %**, **NO-GO**. No se ejecuto
  `rs` ni se creo PR.

## Actualizacion 2026-08-14, soporte remoto completamente cancelable

- Uso del plan, dispositivos, sesiones, limites, tokens de visualizacion y
  credenciales WebRTC propagan el contexto de las solicitudes empresarial,
  publica y super hasta PostgreSQL, siempre filtradas por `empresa_id`.
- Heartbeat verifica que la mutacion afecte exactamente al dispositivo esperado.
  Si el plan bloquea una sesion y no puede persistirse su auditoria, el fallo es
  observable y conserva `ErrSoporteRemotoPlanLimit` para respuesta fail-closed.
- Regresiones estaticas impiden que los handlers vuelvan a los wrappers sin
  contexto. Pruebas DB/handlers, `vet` y auditor aprueban; deuda DB **571 a
  548**, descartes **767 a 766** y duplicacion exacta 0.
- P110-001A/P110-007 continuan parciales; el avance formal permanece en **38,5
  %**, certificacion **0 %**, **NO-GO**, porque no existe candidato desplegado ni
  evidencia operativa equivalente. No se ejecuto `rs` ni se creo PR.

## Actualizacion 2026-08-14, venta digital cancelable y fallos observables

- Configuracion, catalogo, ordenes, conciliacion Wompi y entrega de licencias
  conservan `r.Context()` hasta PostgreSQL. La creacion y consulta de
  transacciones Wompi usan tambien el contexto real de la solicitud.
- Asociacion de imagenes/instrucciones, registro de rechazos Wompi,
  sincronizacion y entrega dejan de descartar errores; el cliente recibe un
  fallo observable en vez de una confirmacion falsa.
- La regresion protege el contrato HTTP. Pruebas DB/handlers, `vet` y auditor
  aprueban; deuda DB **548 a 533**, descartes **766 a 762** y duplicacion 0.
- P110-001A/P110-006/P110-007 continuan parciales hasta validar el candidato
  desplegado, proveedor y correo reales. Avance formal **38,5 %**, certificacion
  **0 %**, **NO-GO**. No se ejecuto `rs` ni se creo PR.
