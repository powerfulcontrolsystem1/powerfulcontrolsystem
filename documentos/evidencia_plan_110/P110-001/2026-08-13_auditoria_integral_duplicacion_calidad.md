# P110-001 - Auditoría integral de duplicación y calidad

Fecha: 2026-08-13

Ambiente: revisión estática local del árbol del candidato; sin despliegue, sin
mutaciones de negocio, sin uso de credenciales y sin ejecutar `rs`.

Estado: **hallazgos abiertos; P110-001 permanece parcial y el sistema continúa
en NO-GO**.

## Objetivo

Revisar el proyecto completo para distinguir duplicación accidental, módulos
con responsabilidades superpuestas, deuda estructural y malas prácticas que
puedan impedir una entrada segura en producción. Esta evidencia no convierte
una coincidencia de texto en defecto: cada hallazgo se clasifica como
confirmado, intencional o pendiente de caracterización.

## Alcance medido

| Superficie | Archivos | Líneas |
|---|---:|---:|
| Go | 638 | 286.823 |
| HTML | 311 | 128.856 |
| JavaScript | 91 | 35.557 |
| CSS | 19 | 21.186 |

La revisión combinó AST de Go, inventario de rutas, esquema y migraciones,
comparación normalizada de frontend, búsqueda semántica de módulos y los
auditores profesionales ya versionados. Los archivos de prueba fueron
separados de la medición de producción cuando el hallazgo dependía de código
ejecutable.

## Método y límites

- Se ejecutaron en modo satisfactorio `professional_audit.mjs`,
  `ux_consistency_audit.mjs`, `permissions_license_audit.mjs` y
  `security_audit.mjs` con salida temporal ignorada por Git.
- `ensure_bootstrap_inventory.mjs --check`,
  `runtime_ensure_inventory.mjs --check` y
  `tenant_route_inventory.mjs --check` confirmaron que los inventarios
  versionados están sincronizados.
- `qa_csp_inline_inventory.mjs` confirmó 1.269 bloqueos en 195 páginas, sin
  regresión contra la línea base.
- El preflight profesional rápido aprobó sintaxis frontend, auditorías,
  migraciones, seguridad, permisos, módulos críticos, roles, pagos,
  observabilidad, UX, Docker y `git diff --check`.
- Un analizador AST local, temporal y sin dependencias externas midió funciones,
  tamaño, duplicados exactos y patrones de contexto/DB. P110-001A exige
  convertir las reglas aceptadas en una compuerta versionada antes del cierre.
- El frontend se comparó eliminando scripts y estilos antes de buscar IDs DOM;
  esto evitó contar interpolaciones JavaScript como nodos HTML.
- No se inició servidor, no se navegó el sistema y no se consultó VPS, staging
  ni producción. Los hallazgos son estáticos; cada refactorización debe añadir
  prueba funcional, visual y multiempresa antes de considerarse corregida.

## Controles que no mostraron duplicación funcional

- Las 204 rutas `/api/empresa/` inventariadas tienen wrapper autoritativo; no
  se detectaron rutas empresariales duplicadas ni pendientes de clasificación.
- Los 59 módulos de permiso están cubiertos por licencias y frontend según el
  auditor vigente.
- El catálogo profesional de plantillas no presenta módulos duplicados.
- No se encontraron archivos frontend completos idénticos ni IDs HTML
  realmente duplicados. Los resultados iniciales de IDs repetidos eran cadenas
  JavaScript y se descartaron al excluir `script` y `style` del análisis DOM.
- `auditoria_global.html` y `auditoria_super_admin.html` son vistas distintas
  por alcance y rol; no deben fusionarse sin rediseñar su contrato de acceso.
- La configuración empresarial de venta pública, la tienda pública y el pago
  público comparten APIs por diseño, pero cumplen etapas distintas del flujo.
- La API móvil v1 es una fachada deliberada sobre servicios web y permanece
  fuera del piloto nativo; no se considera un segundo motor de negocio.

Estos resultados evitan una refactorización masiva basada solo en nombres o
similitud visual.

## Hallazgos confirmados

### P0 - Autoridad de esquema todavía distribuida

- Hay 155 funciones `Ensure*`; 122 forman parte del catálogo legado.
- Persisten 106 llamadas `Ensure*` fuera de la autoridad final del migrador:
  72 en arranque, 3 de plataforma y 31 alcanzables durante tráfico HTTP.
- Existen tablas con más de una definición `CREATE TABLE`, entre ellas
  `empresa_nextcloud_accounts`, `empresa_ai_memoria` y
  `empresa_ai_usuario_preferencias`.
- `main()` mide 1.109 líneas y todavía coordina gran parte del bootstrap y del
  registro de rutas.

Riesgo: deriva entre migraciones y runtime, bloqueos DDL durante tráfico,
permisos excesivos y dificultad para demostrar una migración determinista.

Decisión propuesta: una tabla debe tener un único propietario de esquema en
una migración inmutable. API y worker solo verifican versión/readiness; ningún
handler crea, altera o completa esquema.

### P0 - Acceso a datos sin cancelación uniforme

El AST encontró 1.841 llamadas de base de datos sin variante con contexto en
código de producción. Los focos más altos incluyen
`handlers/modulos_faltantes.go`, `handlers/finanzas.go`,
`handlers/facturacion_electronica.go`, `handlers/nomina_sueldos.go`,
`handlers/productos.go`, `db/productos.go` y `db/creditos.go`.

También se localizaron 25 usos de `context.Background()` en handlers. La
mayoría está acotada con `WithTimeout` para comandos o liberación de locks, pero
el adaptador de IA conserva rutas que pierden la cancelación de la solicitud,
aunque el cliente HTTP tenga timeout.

Riesgo: consultas que siguen ocupando conexiones después de cancelar el
request, presión del pool y operaciones externas que sobreviven al usuario.

Decisión propuesta: propagar `r.Context()` por contratos de servicio y DB,
usar `QueryContext`, `ExecContext` y `BeginTx`, y reservar contexto desligado
solo para jobs durables o cleanup acotado y documentado.

### P0/P1 - Errores ignorados sin política común

Se encontraron 748 asignaciones explícitas `_ =` en Go de producción. No todas
son defectos: rollback, close y telemetría de mejor esfuerzo pueden ser
intencionales. Sin embargo, el inventario incluye 41 decodificaciones JSON,
168 lecturas `Scan`, 56 operaciones de evento/auditoría y 7 respuestas JSON
cuyo error se descarta.

Riesgo: entradas parciales aceptadas, métricas incorrectas, pérdida silenciosa
de auditoría y respuestas truncadas.

Decisión propuesta: clasificar cada descarte como `cleanup seguro`, `best
effort observable` o `error obligatorio`. Todo error obligatorio se propaga;
todo best effort emite telemetría saneada y tiene prueba de fallo. Se prohíbe
agregar nuevos descartes sin justificación local.

### P1 - Handlers y funciones con demasiadas responsabilidades

De 7.183 funciones de producción, 450 superan 100 líneas y 124 superan 200.
Los casos principales son:

| Función | Líneas |
|---|---:|
| `EmpresaCarritosCompraHandler` | 2.052 |
| `main` | 1.109 |
| `EmpresaFacturacionElectronicaHandler` | 1.095 |
| `EnsureEmpresaModulosFaltantesSchema` | 951 |
| `EmpresaNominaSueldosHandler` | 861 |
| `EmpresaComprasDocumentosHandler` | 775 |
| `EmpresaImpresorasHandler` | 723 |
| `EnsureEmpresaProductosSchema` | 688 |
| `EmpresaDIANColombiaHandler` | 569 |

Riesgo: cambios con amplio radio de impacto, difícil caracterización por acción
y mayor probabilidad de omitir permiso, tenant, transacción o rollback.

Decisión propuesta: separar por acción/caso de uso y mantener un handler
orquestador pequeño. La división debe conservar middleware, `empresa_id`,
transacciones e idempotencia; no se autoriza copiar bloques a nuevos archivos.

### P1 - Utilidades idénticas dispersas

El AST confirmó 52 grupos de cuerpos de función idénticos en archivos de
producción. Los grupos principales repiten selección del primer valor no vacío,
patrones `LIKE`, normalización de límite/offset, detección de índices, períodos,
hosts y variables de entorno.

Riesgo: correcciones parciales y variaciones futuras de una misma regla.

Decisión propuesta: extraer solo utilidades con semántica realmente común a
paquetes internos de bajo nivel. Las normalizaciones de dominio que coincidan
accidentalmente permanecen en su módulo hasta contar con pruebas de contrato.

### P1 - Repetición de sistema visual y deuda CSP

El analizador encontró al menos 100 grupos de declaraciones CSS idénticas. Las
pantallas financieras repiten pestañas, campos, chips y tablas con prefijos
distintos; otros diez módulos repiten el mismo contenedor de tabla. Además, el
inventario vigente conserva 209 scripts inline y deuda CSP en 195 páginas.

Riesgo: correcciones responsive y de accesibilidad inconsistentes, archivos
monolíticos (`web/estilos.css` supera 21.000 líneas) y retraso para aplicar una
CSP estricta.

Decisión propuesta: ampliar el sistema visual compartido por primitivas
(`tabs`, `field`, `chip`, `table-wrap`, diálogos y documentos) y migrar por
lotes con comparación visual. No se fusionan módulos de negocio por compartir
CSS.

### P1/P2 - Patrones frontend que requieren clasificación

La búsqueda heurística contabilizó 51 `alert`, 95 `confirm`, 54 `prompt`, 13
`document.write`, 1.304 asignaciones `innerHTML`, 668 `catch` vacíos y 287
accesos a `localStorage`. No se detectó `eval` ni `new Function`. Los seis
`empresa_id` literales están en ejemplos de Ayuda de APIs, no en decisiones de
tenant del runtime.

Estos conteos son inventario, no 2.472 vulnerabilidades. La compuerta debe:

1. eliminar diálogos nativos de flujos críticos;
2. demostrar escape o construcción DOM segura para datos dinámicos;
3. limitar `localStorage` a preferencias no autoritativas;
4. reemplazar `document.write` por el motor imprimible compartido;
5. hacer observables los `catch` de acciones críticas.

### P1 - Solapamientos históricos que necesitan fecha de retiro

- Los endpoints genéricos `/api/empresa/produccion/bom*` siguen registrados
  aunque MRP es el módulo canónico de producción.
- La cartera histórica `empresa_contabilidad_cartera_cxp` conserva CxC y lectura
  CxP. Las guardas actuales rechazan nuevas CxP y abonos CxP, pero debe existir
  conciliación, dueño y fecha de retiro de la rama histórica.
- `modulos_faltantes.go` conserva CRUD genéricos junto a módulos profesionales;
  cada configuración debe clasificarse como canónica, compatibilidad temporal
  o retirada.

Riesgo: dos caminos visibles o programables para la misma capacidad y soporte
indefinido de contratos que ya no deben recibir escrituras.

### P1 - Cobertura global no tiene línea base ejecutable acotada

El intento `go test ./... -coverprofile=...` excedió 60 segundos y se cerró sin
archivo de cobertura válido. Un timeout es inconcluso, no un fallo funcional;
sí demuestra que hoy no hay una línea base de cobertura completa, rápida y
reproducible para guiar una refactorización transversal.

Decisión propuesta: CI debe publicar cobertura por paquetes y total, con
timeout explícito, tendencia y umbral de no regresión. Antes de dividir un god
handler se crean pruebas de caracterización por acción, permiso, tenant,
transacción y error.

### P2 - Contexto documental contradictorio

Los contextos general y rápido todavía presentan Planes 106/108 como hojas de
ruta vigentes. Esto contradice el Plan 110 y puede dirigir a otro ejecutor a
usar una fuente obsoleta.

Decisión propuesta: declarar Plan 110 como única hoja de ruta activa y conservar
las entradas anteriores únicamente como historial.

## Propuesta de ejecución incorporada a P110-001A

1. Congelar una matriz `capacidad -> página -> API -> servicio -> tabla ->
   permiso -> dueño -> estado` y retirar cualquier escritura secundaria.
2. Crear pruebas de caracterización de los nueve handlers más grandes antes de
   dividirlos por acción.
3. Mover toda autoridad DDL a migraciones inmutables y dejar cero llamadas de
   esquema en tráfico HTTP.
4. Propagar contextos en rutas críticas empezando por DIAN, pagos, carrito,
   CxP, inventario, nómina e IA.
5. Clasificar los 748 descartes; corregir primero entrada, auditoría, fiscal,
   dinero y seguridad.
6. Consolidar utilidades idénticas con pruebas, sin crear un paquete genérico
   de reglas de negocio.
7. Extraer primitivas visuales compartidas y completar CSP por lotes con
   regresión responsive, accesibilidad e impresión.
8. Retirar o redirigir contratos históricos solo después de telemetría de uso,
   compatibilidad documentada y rollback.
9. Publicar cobertura por paquete y bloquear regresiones de complejidad,
   duplicación y cobertura en CI.

## Criterio de cierre

P110-001A no se aprueba hasta demostrar simultáneamente:

- cero DDL alcanzable desde tráfico HTTP y un dueño por tabla;
- cero fuentes de escritura duplicadas para una capacidad;
- cero duplicaciones exactas sin clasificar y ninguna nueva;
- contexto/cancelación en todas las operaciones críticas de DB/proveedor;
- errores descartados clasificados y los críticos observables o propagados;
- handlers críticos divididos con caracterización y sin degradar cobertura;
- contratos históricos con dueño, telemetría, fecha de retiro y reverso;
- regresión Go, seguridad, permisos, multiempresa, visual, impresión y CSP en
  verde sobre un mismo candidato.

Esta auditoría no aumenta el porcentaje del Plan 110: identifica trabajo
faltante dentro de P110-001 y mantiene implementación en **38,5 %**,
certificación del candidato en **0 %** y veredicto **NO-GO**.
