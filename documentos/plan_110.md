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
2. P110-001 a P110-007: correcciones funcionales y controles P0.
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

Un hallazgo P0 mantiene el plan en NO-GO, aunque otras fases continúen.

## 9. Compuerta final GO/NO-GO

Solo se puede declarar GO cuando:

- [ ] P110-000 a P110-011 están aprobadas.
- [ ] Existe un único SHA y cuatro digests certificados.
- [ ] CI, Trivy, SBOM, migraciones y staging están verdes.
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
  antes de enviarla. Se mantiene pendiente el rechazo HTTP EICAR, recuperación,
  concurrencia y alertas P0. P110-007 permanece **parcial**; el avance formal
  continúa en **38,5 %**, certificación **0 %**, **NO-GO**.
