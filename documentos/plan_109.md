# Plan 109 - cierre ejecutivo de producción de Powerful Control System

Fecha de creación: 2026-07-30  
Estado inicial: **NO-GO**  
Modelo ejecutor previsto: **GPT-5.6 Terra con razonamiento alto**  
Repositorio: `D:\powerfulcontrolsystem`  
Rama observada: `codex/p108-large-validation-block`  
SHA observado al crear el plan:
`eb1177243ca8a0f88d67d5b5d94d11aff1926a62`

## 1. Propósito y autoridad del plan

El Plan 109 es la continuación única y depurada del Plan 108. No vuelve a
contar trabajo histórico: convierte sus evidencias positivas y sus pendientes
en compuertas finales, ordenadas por dependencia, para decidir si Powerful
Control System puede iniciar un piloto productivo.

Los planes 106, 107, IA y 108 permanecen como antecedentes técnicos. Desde la
aprobación de este documento, todo trabajo nuevo de preparación para producción
debe registrarse con identificadores `P109-*`.

Este plan no transforma automáticamente una autorización de staging en
autorización de producción. Las pruebas reales autorizadas se limitan a la
empresa Powerful Control System en el entorno de pruebas indicado. Emisiones
DIAN, cobros, mensajes externos, compras, cambios DNS, borrados y despliegue
productivo requieren verificar la autorización vigente inmediatamente antes
de ejecutarlos.

## 2. Base auditada y corrección del porcentaje

### 2.1 Estado Git y candidato

- La PR `#105` contiene la recuperación auditada de eventos outbox CxP.
- El SHA observado es fusionable y sus dos jobs obligatorios de CI terminaron
  correctamente.
- La PR todavía requiere una aprobación con permisos de escritura.
- El candidato publicado actualmente en staging es anterior a esa recuperación;
  por tanto no puede certificar el código nuevo.
- Producción no fue modificada por el bloque de recuperación CxP.

### 2.2 Estado verificable heredado del Plan 108

Fases aprobadas:

- P108-000: gobierno y candidato.
- P108-001: CI y cadena de suministro.
- P108-021: aplicación nativa excluida formalmente; web responsive/PWA incluida.

Fases parciales con evidencia útil:

- P108-002 a P108-007, excepto P108-008.
- P108-009 a P108-013.
- P108-015 a P108-020.
- P108-022.

Fases pendientes sin cierre:

- P108-008: contador profesional y cierre integral del Plan 107.
- P108-014: cuatro cajas simultáneas. Existen cuatro invitaciones, pero sus
  identidades siguen sin confirmar/activar y no se ejecutó la corrida.
- P108-023: staging equivalente y ensayo general.
- P108-024: prueba real integral en Powerful Control System.
- P108-025: piloto, decisión GO/NO-GO y producción.

El corte de P108 publicó 48,1 % al contar 19 fases parciales. Su evidencia
mantiene P108-014 como pendiente, y la regla del propio plan asigna 0 % a una
fase pendiente. El cálculo conservador corregido es:

`(3 aprobadas + 18 parciales x 0,5) / 26 = 46,2 %`.

La certificación histórica del digest anterior fue 23,1 %. El Plan 109 comienza
en **0 % de certificación del nuevo candidato**, porque la recuperación CxP y
este plan aún no pertenecen a un digest desplegado y ensayado.

### 2.3 Pendientes materiales consolidados

1. Fusionar la PR aprobada, construir imágenes inmutables y promover solo el
   mismo digest.
2. Migrar y probar la recuperación explícita de los eventos CxP históricos.
3. Ejecutar aislamiento A/B real con segunda identidad empresarial controlada.
4. Reparar la configuración cifrada de IA en staging y completar CxP/IA,
   reportes IA y Centro IA.
5. Completar contador, impuestos, estados, reportes, anulaciones y matriz
   financiera del Plan 107.
6. Probar las funciones oficiales que el barrido preservó: 1.980 acciones
   mutantes/riesgosas, con prioridad por alcance del piloto y todos los botones IA.
7. Activar cuatro cajeros por el flujo normal y ejecutar cajas concurrentes.
8. Completar documentos reales, impresión, accesibilidad y roles.
9. Cerrar DIAN e integraciones reales autorizadas, incluida anulación fiscal.
10. Completar réplica A/B de archivos, restore funcional y rollback coordinado.
11. Ejecutar DAST autenticado, sesiones, CSP y pruebas hostiles.
12. Completar alertas de lease/almacenamiento y responsables de escalamiento.
13. Ensayar runbooks, capacitación, soporte y límites del piloto.
14. Ejecutar ensayo general, prueba PCS, piloto y decisión GO/NO-GO.
15. Publicar VMC/CMC BIMI si se exige eliminar la inicial del avatar en clientes
    que condicionan el logo a un certificado externo.

## 3. Instrucciones para GPT-5.6 Terra alto

La guía oficial de GPT-5.6 recomienda conservar contexto de dominio,
restricciones duras, límites de aprobación y criterios de éxito; también
recomienda usar razonamiento alto cuando exista una ganancia de calidad medida.
Referencia:
`https://developers.openai.com/api/docs/guides/model-guidance?model=gpt-5.6-terra`.

Terra debe ejecutar el plan con estas reglas:

1. Usar `gpt-5.6-terra` con razonamiento `high`.
2. Leer completos `AGENTS.md`, `contexto_general_del_sistema.md`,
   `contexto_codex.md`, `comandos_codex.md`, `decisiones_tecnicas.md`,
   `checklist_seguridad_endpoint_multiempresa.md`, este plan y la evidencia de
   la fase antes de actuar.
3. Revisar rama, SHA, `git status`, PR, CI y ambiente antes de editar o desplegar.
4. Trabajar en bloques grandes y coherentes, pero con una sola compuerta en
   riesgo a la vez. Evitar una PR por prueba; usar una PR consolidada por
   candidato.
5. No cambiar `go.mod`, dependencias, motor PostgreSQL ni arquitectura sin
   autorización expresa.
6. Mantener `empresa_id`, rol, permiso, licencia, usuario y alcance derivados y
   verificados en servidor.
7. No guardar ni imprimir contraseñas, tokens, API keys, certificados, DSN,
   enlaces de invitación o rutas privadas.
8. Usar credenciales de prueba únicamente desde el canal seguro existente.
9. Para revisar/diagnosticar, realizar lecturas y reportar. Para corregir,
   hacer cambios locales acotados y validarlos. Solicitar confirmación antes de
   producción, compras, efectos externos nuevos o acciones destructivas.
10. Un timeout es inconcluso. Un test estático no sustituye una prueba
    autenticada. Staging no equivale a producción.
11. No ejecutar `rs` desde un `main` sucio ni promover una rama sin CI/revisión.
12. No declarar una fase aprobada sin evidencia del mismo SHA/digest.
13. Si una acción cambia datos, registrar antes alcance, reverso e identidad;
    después conciliar por API/UI y consultas de solo lectura.
14. Detener el trabajo dependiente si aparece un P0, pero continuar en frentes
    independientes seguros.
15. Informar porcentaje solo con la fórmula de la sección 8.

### Formato obligatorio de cada ciclo

Antes de ejecutar:

- identificador P109;
- objetivo;
- SHA/digest y ambiente;
- empresa, rol y datos afectados;
- efectos externos;
- pruebas previstas;
- rollback;
- criterios de aceptación.

Al cerrar:

- cambio o prueba realizada;
- archivos/rutas/tablas afectados;
- comandos y resultado;
- evidencia visual/funcional;
- aislamiento y permisos;
- datos creados y su limpieza;
- riesgos abiertos;
- estado `pendiente`, `parcial`, `aprobado`, `bloqueado` o `fallido`;
- porcentaje actualizado.

## 4. Alcance obligatorio del piloto

Antes del ensayo general se debe publicar una matriz firmada con:

- módulos incluidos, visibles deshabilitados y excluidos;
- país, moneda, impuestos y facturación habilitados;
- roles y licencias;
- empresas piloto y número de cajas;
- integraciones requeridas;
- SLO, RPO y RTO;
- horarios y responsables de soporte;
- datos reales permitidos;
- criterios de reverso;
- riesgos aceptados y fecha de vencimiento.

Todo módulo visible se considera incluido hasta que servidor, permiso/licencia
y UX demuestren lo contrario.

## 5. Fases de implementación y certificación

### P109-000 - Integración, release y promoción controlada [P0]

**Origen:** P108-000, P108-001 y recuperación CxP.

**Acciones:**

1. Obtener aprobación de la PR consolidada y verificar CI verde.
2. Fusionar sin conflictos y comprobar árbol limpio en `main`.
3. Construir API, migrador, worker y frontend una sola vez desde el SHA completo.
4. Escanear las cuatro imágenes, generar SBOM y guardar digests.
5. Ejecutar migraciones sobre base efímera vacía y upgrade de snapshot.
6. Promover los cuatro digests exactos solo a staging, sin recompilar.
7. Verificar migración, `/health`, `/ready`, worker, métricas y huella de
   producción sin cambios.

**Aceptación:** PR aprobada/fusionada, CI/release verdes, cuatro digests,
migraciones idempotentes, staging sano y producción intacta.

**Rollback:** volver a los cuatro digests previos de staging y restaurar el
snapshot solo si la migración no es compatible hacia atrás.

**Evidencia:** `documentos/evidencia_plan_109/P109-000/`.

### P109-001 - Recuperación CxP y aislamiento A/B [P0]

**Origen:** P108-003, P108-004 y P108-005.

**Acciones:**

1. Abrir como super administrador la vista previa del outbox para PCS.
2. Verificar que solo aparezcan eventos `dead` del topic CxP y que el payload
   crudo no sea visible.
3. Probar rechazo sin empresa, otro topic, ID publicado, ID inexistente, ID
   repetido y evento de empresa B.
4. Reactivar únicamente los dos IDs inventariados de PCS con razón y
   confirmación.
5. Conciliar outbox, job, evento y asiento por cada pago; exigir exactamente uno.
6. Repetir la solicitud para demostrar rechazo seguro/idempotencia.
7. Comprobar que cuenta, pagos y movimientos no cambian al reprocesar el evento.
8. Ejecutar matriz A/B de CxP, proveedores, auditoría, cache y jobs con una
   segunda identidad empresarial controlada.

**Aceptación:** dos eventos históricos publicados/completados, dos eventos
contables naturales, asientos balanceados, cero pagos duplicados, auditoría por
evento y cruces A/B rechazados.

**Rollback:** pausar worker/outbox de staging, conservar auditoría y corregir el
handler; no revertir pagos ya existentes.

**Evidencia:** `documentos/evidencia_plan_109/P109-001/`.

### P109-002 - IA empresarial operativa y segura [P0]

**Origen:** P108-006, P108-009, P108-010 y Plan IA.

**Acciones:**

1. Reconfigurar `CONFIG_ENC_KEY_ID` y la credencial OpenAI mediante flujo seguro.
2. Reiniciar staging y comprobar lectura sin revelar la clave.
3. Cargar factura y recibo de prueba; validar firma/tamaño/tipo y almacenamiento.
4. Ejecutar todos los botones IA visibles: extracción CxP, edición, confirmar,
   cancelar, doble clic, error de proveedor y reintento.
5. Confirmar que IA no crea CxP antes de revisión humana.
6. Probar duplicado por hash/documento y aislamiento A/B.
7. Generar ReportSpec válido; probar campos inexistentes, inyección, rol,
   cancelación, vista previa, exportación y plantilla.
8. Habilitar Centro IA por permiso oficial; probar modo agente apagado por
   defecto, memoria por usuario/empresa, auditoría, costo, latencia y rechazo A/B.
9. Ejecutar evals de exactitud y seguridad con umbrales publicados.

**Aceptación:** botones IA incluidos en piloto probados; CxP editable y
confirmable; reportes seguros; memoria/acciones aisladas; degradación clara;
cero secretos o SQL/herramientas arbitrarias.

**Rollback:** deshabilitar capacidades por permiso/configuración sin borrar
historial auditable.

**Evidencia:** `documentos/evidencia_plan_109/P109-002/`.

### P109-003 - Contabilidad, impuestos y Plan 107 [P0]

**Origen:** P108-007 y P108-008.

**Acciones:**

1. Definir fixtures y saldos de apertura controlados.
2. Conciliar ventas, compras, CxP/CxC, pagos, devoluciones, notas crédito,
   anulaciones, impuestos, retenciones y moneda.
3. Validar débitos=créditos, periodos, consecutivos y cierres.
4. Ejecutar catálogo completo de reportes contables y sus exportaciones.
5. Probar estados de resultados, situación financiera, patrimonio, auxiliares,
   comparativos, notas y firmas según alcance.
6. Completar impuestos, exógena y cierres aplicables al país piloto.
7. Ejecutar UAT con un contador distinto del desarrollador.

**Aceptación:** diferencias en cero, trazabilidad documento-evento-asiento,
reportes reproducibles, cierre/reapertura controlados y UAT firmado.

**Rollback:** reversos contables oficiales; nunca editar asientos por SQL.

**Evidencia:** `documentos/evidencia_plan_109/P109-003/`.

### P109-004 - Regresión funcional y todos los botones [P0]

**Origen:** P108-011.

**Acciones:**

1. Congelar inventario de módulos y controles del alcance.
2. Clasificar las 1.980 acciones preservadas por riesgo y módulo.
3. Ejecutar por flujo oficial todas las funciones CRUD, filtros, exportaciones,
   cambios de estado, anulaciones, carga/descarga y navegación.
4. Probar todos los botones IA y su autorización.
5. Probar cada rol del piloto y accesos negados.
6. Repetir rutas en escritorio/móvil con consola/red y auditoría.
7. Resolver los 242 timeouts, 37 desbordes internos, 30 botones sin etiqueta,
   ocho navegaciones externas y cuatro desbordes horizontales o documentar
   exclusión verificable.

**Aceptación:** 100 % del inventario incluido tiene PASS o exclusión firmada;
cero 5xx, errores de página, controles sin resultado o mutaciones sin auditoría.

**Evidencia:** `documentos/evidencia_plan_109/P109-004/`.

### P109-005 - Visual, accesibilidad e impresiones reales [P0]

**Origen:** P108-012 y P108-013.

**Acciones:**

1. Revisar escritorio, móvil y tableta por cada rol.
2. Probar teclado, foco, lector de pantalla, contraste, zoom y mensajes.
3. Verificar factura, recibo, comprobante, pedido, orden, corte, reporte y todo
   documento imprimible del alcance.
4. Usar documentos reales cortos y extensos, varias páginas, carta y POS.
5. Confirmar filas, columnas, totales, impuestos, firmas, QR/CUFE y pies.
6. Ejecutar impresión física en los dispositivos del piloto.

**Aceptación:** cero solapamientos/desbordes críticos, teclado/lector funcionales,
permisos correctos y firma visual de cada formato/dispositivo.

**Evidencia:** `documentos/evidencia_plan_109/P109-005/`.

### P109-006 - Cuatro cajas y carga transaccional sostenida [P0]

**Origen:** P108-014 y P108-019.

**Acciones:**

1. Confirmar invitaciones y contraseñas por flujo normal para cuatro cajeros.
2. Abrir cuatro sesiones y cajas independientes.
3. Ejecutar ventas concurrentes sobre inventario compartido, medios de pago,
   devoluciones, facturación, descuentos y cierre.
4. Probar reintentos/doble clic/idempotencia, consecutivos y stock.
5. Mantener carga transaccional sostenida con observación de p50/p95/p99,
   locks, pool, CPU, memoria, disco, colas y backpressure.
6. Conciliar cada caja y el consolidado contable.
7. Desactivar las cuentas temporales al terminar.

**Aceptación:** cero doble cobro/sobreventa/consecutivo duplicado; cuatro cierres
cuadrados; SLO cumplido; degradación y backpressure controlados.

**Evidencia:** `documentos/evidencia_plan_109/P109-006/`.

### P109-007 - DIAN, correo, pagos e integraciones [P0/P1]

**Origen:** P108-015.

**Acciones:**

1. Verificar separación DIAN por empresa: NIT, certificado, clave, consecutivo,
   resolución y trazabilidad.
2. Emitir únicamente la factura real autorizada del producto de menor costo.
3. Exigir aceptación oficial DIAN y `GetStatusZip StatusCode=00`.
4. Visualizar/imprimir la factura y ejecutar anulación fiscal por nota crédito
   total; verificar aceptación y relación.
5. Probar Wompi/Epayco según alcance sin inventar aprobaciones.
6. Probar Mailu: registro de administrador, invitación, reset, alertas,
   confirmaciones, rebotes, SPF/DKIM/DMARC y logo dentro del cuerpo.
7. Probar WhatsApp, Nextcloud, OnlyOffice y proveedores incluidos.
8. Completar VMC/CMC y BIMI si el criterio comercial exige sustituir la inicial
   del avatar en Gmail/clientes compatibles. Registrar costo y dependencia
   externa; no comprar sin autorización.

**Aceptación:** evidencia oficial por proveedor, reintentos idempotentes,
errores públicos seguros, aislamiento por empresa y conciliación de efectos.

**Rollback:** desactivar proveedor/configuración por empresa y usar reverso
oficial; nunca falsificar estado externo.

**Evidencia:** `documentos/evidencia_plan_109/P109-007/`.

### P109-008 - Archivos, backup, restore y rollback [P0]

**Origen:** P108-016 y P108-020.

**Acciones:**

1. Subir en réplica A y descargar en réplica B.
2. Ejecutar matriz negativa de empresa, traversal, symlink, contenido activo,
   cuota, retención y borrado.
3. Inventariar/migrar archivos heredados y huérfanos.
4. Restaurar bases, volumen privado, metadatos y checksums.
5. Verificar CxP, contabilidad, IA, DIAN y documentos después del restore.
6. Ensayar pérdida de réplica y rollback coordinado de aplicación/base.
7. Medir RPO/RTO y limpiar recursos temporales.

**Aceptación:** restore funcional integral dentro de objetivos, datos/archivos
consistentes, réplica intercambiable y rollback demostrado.

**Evidencia:** `documentos/evidencia_plan_109/P109-008/`.

### P109-009 - Seguridad dinámica y CSP [P0]

**Origen:** P108-004 y P108-017.

**Acciones:**

1. Ejecutar DAST autenticado y no autenticado.
2. Probar XSS, CSRF, SSRF, IDOR, sesión, revocación, subida hostil, CORS,
   rate limits y enumeración.
3. Ejecutar matriz A/B para SQL, archivos, cache, jobs, reportes, IA,
   exportaciones y descargas.
4. Inventariar `unsafe-inline`; migrar a nonces/hashes o aprobar excepción
   mínima con responsable/fecha.
5. Escanear imágenes/digest final y revisar hardening VPS/Docker.

**Aceptación:** cero P0/P1 explotable, aislamiento A/B completo, sesión segura y
CSP con riesgo explícitamente cerrado.

**Evidencia:** `documentos/evidencia_plan_109/P109-009/`.

### P109-010 - Observabilidad y operación [P0]

**Origen:** P108-018.

**Acciones:**

1. Ensayar lease vencido, outbox/job dead, almacenamiento privado y disco.
2. Verificar alertas, deduplicación, recepción y resolución.
3. Configurar canal externo autorizado y responsables/escalamiento.
4. Publicar dashboards/SLO, error budget y consultas de diagnóstico.
5. Ejecutar simulacro de incidente con runbook y postmortem.

**Aceptación:** todas las señales P0 detectan y resuelven dentro del objetivo,
con dueño, canal, evidencia y sin exposición pública de métricas.

**Evidencia:** `documentos/evidencia_plan_109/P109-010/`.

### P109-011 - Migraciones y rollback de datos [P0]

**Origen:** P108-002 y P108-020.

**Acciones:**

1. Ejecutar base vacía, upgrade de snapshot y segunda pasada con el digest final.
2. Verificar API/worker sin DDL.
3. Simular fallo antes, durante y después de migración.
4. Ejecutar matriz de compatibilidad hacia atrás y rollback de datos.
5. Verificar checksums, auditoría y cero residuos.

**Aceptación:** migraciones deterministas y rollback de aplicación/datos
ensayado sin pérdida fuera del RPO.

**Evidencia:** `documentos/evidencia_plan_109/P109-011/`.

### P109-012 - Documentación, soporte y capacitación [P1]

**Origen:** P108-022.

**Acciones:**

1. Actualizar manuales por cajero, administrador, contador y soporte.
2. Ensayar instalación, deploy, rollback, restore, DIAN, IA y proveedores con
   una persona distinta del desarrollador.
3. Registrar contactos, horario, escalamiento, módulos deshabilitados y límites.
4. Ejecutar escaneo final de secretos y enlaces rotos.
5. Firmar capacitación y aceptación.

**Aceptación:** runbooks reproducibles, responsables identificados y soporte
listo para el piloto.

**Evidencia:** `documentos/evidencia_plan_109/P109-012/`.

### P109-013 - Ensayo general y prueba real PCS [P0]

**Origen:** P108-023 y P108-024.

**Acciones:**

1. Congelar digest/configuración y snapshot.
2. Ejecutar de extremo a extremo registro, empresa, licencia, usuarios,
   inventario, compra, CxP/IA, venta, caja, factura DIAN, anulación, reportes,
   contabilidad, archivos, correo e integraciones incluidas.
3. Ejecutar visual, responsive, impresión, roles, A/B, IA y cuatro cajas.
4. Simular incidente y rollback.
5. Conciliar datos de PCS y limpiar/desactivar fixtures temporales por flujos
   oficiales.
6. Firmar resultado PASS/FAIL y lista de riesgos.

**Aceptación:** mismo digest, cero P0/P1, conciliación completa, rollback probado
y responsables conformes.

**Evidencia:** `documentos/evidencia_plan_109/P109-013/`.

### P109-014 - Piloto, GO/NO-GO y producción [P0]

**Origen:** P108-025.

**Acciones:**

1. Aprobar alcance, riesgos y plan de reverso.
2. Iniciar piloto limitado con observación reforzada.
3. Medir estabilidad, errores, SLO, soporte, integraciones y conciliación.
4. Resolver o aceptar formalmente hallazgos.
5. Tomar decisión GO/NO-GO firmada.
6. Solo con GO, promover los mismos digests a producción dentro de ventana.
7. Ejecutar smoke, observación y rollback si se supera cualquier umbral.

**Aceptación:** piloto estable, decisión firmada, sin P0/P1 abiertos y
producción verificable o rollback completo.

**Evidencia:** `documentos/evidencia_plan_109/P109-014/`.

## 6. Matrices obligatorias

### 6.1 Matriz por función

Cada función incluida debe registrar:

`módulo | ruta | rol | empresa | control | acción | efecto esperado | estado
HTTP | auditoría | visual | responsive | rollback | resultado`.

### 6.2 Matriz de IA

Cada botón IA debe probar:

- permitido y denegado;
- archivo válido/inválido;
- doble clic/reintento/cancelación;
- proveedor disponible/no disponible;
- datos incompletos e inyección;
- edición humana;
- confirmación explícita;
- empresa B;
- costo/tokens/latencia;
- auditoría y degradación.

### 6.3 Matriz financiera

Debe conciliar:

`documento | venta/compra | pago | caja/banco | impuesto/retención |
evento | asiento | reporte | reverso | diferencia`.

### 6.4 Matriz de impresiones

Debe incluir carta/POS, documento corto/extenso, varias páginas, escritorio,
móvil/tableta, rol permitido/denegado, vista previa, PDF e impresión física.

## 7. Compuerta GO/NO-GO

El resultado solo puede ser GO cuando:

- [ ] P109-000 a P109-014 están aprobadas o una P1 está excluida por decisión
      firmada sin ocultar funcionalidad crítica.
- [ ] No hay P0/P1 explotable ni fallos financieros/fiscales abiertos.
- [ ] El mismo digest aprobó CI, staging, ensayo, piloto y promoción.
- [ ] Todas las funciones y botones incluidos tienen evidencia.
- [ ] Todos los botones IA incluidos tienen evidencia.
- [ ] Cuatro cajas concurrentes cuadran.
- [ ] DIAN y anulaciones tienen estado oficial aceptado.
- [ ] Aislamiento A/B y roles están demostrados.
- [ ] Restore/rollback cumplen RPO/RTO.
- [ ] Observabilidad, soporte y responsables están activos.
- [ ] La decisión GO está firmada.

Una sola compuerta P0 abierta mantiene **NO-GO**.

## 8. Porcentaje del Plan 109

El Plan 109 tiene 15 fases de igual peso.

- `pendiente`, `bloqueada` o `fallida`: 0 %.
- `parcial`: 50 % solo si existe código/prueba/evidencia del mismo bloque.
- `aprobada`: 100 % con todos sus criterios.

Se informan dos cifras:

1. **Implementación:** suma fases aprobadas y parciales.
2. **Certificación del candidato:** solo fases aprobadas sobre el mismo digest.

Estado actualizado al 2026-08-01:

- P109-000 está aprobada en staging para el SHA fusionado `ea9642dd...`: CI
  `30714951274` y release inmutable `30715195384` verdes,
  cuatro digests/SBOM, Trivy sin vulnerabilidades, base vacía, upgrade,
  idempotencia, rechazo de checksum, rollback de aplicación, salud y producción
  intacta. P109-001, P109-002, P109-003, P109-004,
  P109-005, P109-008, P109-009, P109-010, P109-011 y P109-012 tienen evidencia
  parcial. Las demás fases continúan pendientes o bloqueadas.
- P109-001 añadió rechazos runtime para empresa/topic ausentes, duplicados, ID
  inexistente y publicado sin modificar el evento histórico excluido. El pago
  mínimo de `CXP-SCI-0003` quedó conciliado 1:1 entre pago, movimiento, outbox,
  evento y asiento balanceado. La prueba detectó pérdida de un centavo por
  columnas `REAL`; la migración local a `NUMERIC(18,2)` aprobó PostgreSQL
  aislado y preflight, pero aún no está publicada ni aplicada. Falta una segunda
  empresa controlada para la matriz A/B.
- P109-002 ya demostró configuración cifrada, catálogo/preferencia por usuario,
  historial y dashboard CxP/IA. Sobre el candidato final confirmó que el filtro
  vacío lista estados activos, persistió la edición humana y rechazó por el
  botón oficial el soporte no fiscal. El documento quedó sin CxP, pago, asiento
  ni contaminación de producción. El código nuevo serializa e idempotentiza la
  aprobación/rechazo, bloquea doble extracción entre réplicas, exige
  confirmación humana y degrada de forma segura al proveedor; las evals locales
  de faltantes, JSON inválido, confianza y descuadre pasan. Sigue parcial hasta
  comprobar estos contratos sobre el nuevo digest en staging, recorrer Centro
  IA y completar aislamiento A/B con una identidad no global. La primera
  apertura autenticada de Centro IA confirmó su ocultamiento por empresa; al
  habilitarlo por el flujo oficial se detectó y corrigió la falta de invalidación
  inmediata de las cachés de permisos. Sobre `ea9642dd...`, la revocación y
  restauración inmediata, la consola y `Diagnostico ERP` aprobaron. Un overlay
  temporal escaneado corrigió el contrato `file_data` y demostró extracción real
  y edición restaurable de `SCI-0001`, sin CxP ni pago automático. Cancelaciones
  y doble clic aprobaron; la aprobación/rechazo real no convirtió el soporte.
  `SCI-0003` creó una sola CxP canónica y su reproceso fue idempotente;
  `SCI-0004` quedó duplicado sin cartera adicional. La corrección aún debe
  publicarse como candidato inmutable; faltan recibo real y aislamiento A/B.
- P109-004 completo sobre `89d6e042...` 618 vistas/309 rutas y 11.062 controles:
  600 `ok`, 18 revisiones conocidas, 97 clics seguros y 12 POST automáticos
  bloqueados por la guardia. Finanzas y CxP/IA aprobaron además 4/4 vistas
  dirigidas y 114 controles en escritorio/móvil. Faltan acciones riesgosas por
  flujo oficial, rol, empresa A/B y firma del alcance.
- P109-005 amplió su regresión sintética a 20/20 formatos: factura y recibo de
  96 renglones produjeron cinco páginas de detalle más una de resumen, con QR
  cargado y revisión visual mediante Poppler sin recortes. Continúa parcial por
  faltar documentos reales, roles, tableta e impresión física del piloto.
- P109-007 continúa bloqueada: el centro DIAN de PCS muestra ambiente, estado,
  TestSetId y rango como no configurados, avance visible de 10 % y cero
  TrackId/ZipKey. Sin esos datos no es válido emitir, anular ni afirmar un
  `GetStatusZip StatusCode=00` oficial. La repeticion autenticada del
  diagnostico en staging sobre `44610128` confirma HTTP 200, operaciones SOAP
  objetivo visibles y preparacion `sin_configuracion`; tampoco habilita emitir
  la menta de 100 COP que el catalogo identifica, pues su stock visible es 0.
- P109-009 redujo el escaneo rápido del host de dos hallazgos altos a cero altos
  mediante SSH por llave, UFW activo, VNC restringido y retiro de Avahi/CUPS.
  Quedan 30 paquetes y reinicio para una ventana de mantenimiento, además del
  DAST/CSP/A-B que esta prueba no sustituye.
- P109-010 aprobó sobre `89d6e042...` 500/500 GET autenticados con concurrencia
  10, p95 de 217 ms, cero locks, cero 5xx estructurados y servicios saludables.
  Alertmanager recibió y resolvió una alerta sintética interna; falta receptor
  externo y deduplicación. La limpieza de 19
  volúmenes anónimos bajó el disco de 87 % a 79 %; una segunda limpieza exacta
  de imagenes candidatas antiguas recuperó 41,5 GB y lo dejó en 40 %, sin
  afectar servicios ni borrar datos.
- P109-011 aprobó base vacía, upgrade, drift y rollback de aplicación sobre el
  esquema actualizado. El ensayo posterior a migración perdió y restauró las
  dos bases junto con el volumen privado en 23 segundos, recuperando fila,
  cinco dominios y SHA-256. El ensayo ampliado bloqueó un rol sin DDL antes de
  migrar, revirtió DDL y ledger ante un fallo durante la transacción, recuperó
  ambos atómicamente y arrancó la API `8847288b...` sobre el esquema nuevo sin
  mutarlo. Continúa parcial solo por aprobación contractual del RPO/RTO y la
  vinculación al digest final del piloto.
- P109-008 arrancó la API y el migrador exactos de `89d6e042...` contra el
  snapshot restaurado: salud/readiness 200, 5 tablas y 28 filas críticas de PCS,
  4 endpoints anónimos bloqueados y 5 dominios recorridos mediante login oficial.
  El rol temporal no tuvo privilegios de plataforma y la limpieza terminó con
  cero recursos. Los mismos cinco módulos aprobaron revisión visual en
  escritorio/móvil sin desbordamiento de página ni errores de consola; una
  interrupción `TERM` también dejó cero recursos. Dos réplicas de aplicación
  aprobaron carga por A, descarga por B, SHA-256 y continuidad de B tras retirar
  A. Cinco negativos dinámicos bloquearon empresa cruzada, HTML activo, exceso
  de 15 MiB y symlink sin crear filas. Un nuevo ensayo perdió las dos bases y
  todo el volumen privado efímero, y recuperó coherentemente fila, archivo,
  SHA-256, cinco dominios y readiness en 23 segundos, con cero residuos y sin
  tocar servicios activos. El inventario cruzó 2/2 archivos/referencias del
  catálogo privado sin huérfanos, faltantes ni referencias heredadas. Continúa
  parcial por cuotas/retención, antivirus, segunda identidad A/B no global y
  aprobación formal RPO/RTO.
- Implementación Plan 109: **40,0 %** (`1 aprobada + 10 parciales`, de 15 fases).
- Certificación del candidato desplegado: **6,7 %** (solo P109-000 aprobada en
  el mismo digest). Las fases funcionales, fiscales, A/B, cuatro cajas, restore,
  mantenimiento y ensayo general aún impiden promoverlo a piloto.
- Veredicto: **NO-GO**.

Actualización 2026-08-01, candidato `3ed34774`:

- P109-001 conserva estado parcial, pero ya aprobó migración inmutable
  `NUMERIC(18,2)`, saldo exacto 214199,99, respaldo previo, UI y producción
  intacta.
- P109-002 conserva estado parcial, pero ya aprobó en el digest desplegado
  archivos Base64, XML como texto no confiable, extracción real, revisión
  editable y control de duplicado sin crear CxP. También aprobó un recibo real
  `SCI-0008`: doble clic con una sola extracción, lectura exacta de subtotal
  500, IVA 95 y total 595, dos ediciones auditadas y rechazo final sin CxP,
  pago, evento o asiento. La cuota temporal quedó restaurada a 5.
- ReportSpec generó una vista CxP de tres filas, exportó XLS/PDF y rechazó dos
  intentos de campo/fuente fuera del contrato con HTTP 400. Centro IA cargó sus
  siete funciones y completó `Diagnostico ERP` sin alterar ventas, CxP, pagos,
  soportes, eventos o asientos. Ambos resultados pertenecen a `3ed34774`.
- Los porcentajes no aumentan porque ninguna fase parcial cumplió todavía todos
  sus criterios de aceptación A/B y por roles con una segunda identidad no
  global.

Actualización 2026-08-01, candidato `f9694e10`:

- P109-003 recorrió autenticadamente los 46 datasets del catálogo y sus cinco
  formatos. La primera pasada 44/46 reveló dos incompatibilidades PostgreSQL en
  fechas dinámicas y permanencia de vehículos; ambas quedaron corregidas,
  cubiertas por regresión y publicadas mediante el workflow `30730395189`.
- La pasada final obtuvo 46/46 datasets consistentes, 46/46 exportaciones en
  JSON/CSV/TXT/XLS/PDF, 225 filas, cero alertas y 21/21 reportes contables,
  fiscales y de cartera. Descargas autenticadas representativas confirmaron
  MIME y firma de PDF, XLS y CSV.
- La conciliación PCS confirmó cinco asientos balanceados (102,02=102,02), tres
  CxP sin invariantes rotas y cinco pagos enlazados 1:1 con sus movimientos. Los
  cuatro outbox posteriores a habilitar el handler están publicados; se
  conserva un `dead` histórico anterior. La vista previa oficial lo encontró,
  pero el usuario PCS recibió 403 al intentar la recuperación auditada; no se
  forzó por SQL ni se amplió su permiso.
- Producción mantuvo sus imágenes locales y salud/readiness 200. P109-003 sigue
  parcial por casos contables/fiscales no representados, recuperación del
  evento histórico, cierre/reapertura y UAT firmada por contador. Por ello los
  porcentajes permanecen en **40,0 % de implementación** y **6,7 % de
  certificación**, con veredicto **NO-GO**.

Actualización 2026-08-01, candidato `99348ff4`:

- La recepción externa de alertas detectó acumulación real de sesiones de 24
  horas: máximo 244 para una identidad en staging y 155 en producción. El nuevo
  contrato conserva como máximo las 20 sesiones recientes por identidad, poda
  expiradas dentro de la misma transacción y serializa entre réplicas PostgreSQL.
- El workflow inmutable `30731146474` aprobó build, Trivy, SBOM, publicación y
  compose. Tras respaldar las dos bases, los cuatro digests se promovieron solo
  a staging; producción conservó sus imágenes locales y sus datos.
- Veinticinco logins reales consecutivos aprobaron: staging pasó de 249 a 25
  sesiones vigentes, con máximo 20 por identidad y cero identidades fuera del
  límite. La recepción Gmail confirma un canal externo de alertas funcional.
- Las cuatro invitaciones de cajero existen en el buzón. La primera abre el
  formulario válido de staging, pero completar contraseña y contrato requiere
  intervención personal; el usuario PCS tampoco posee actualmente el permiso
  efectivo para reenviarlas. P109-006 sigue pendiente.
- Ninguna fase adicional cumple todos sus criterios, así que el estado se
  mantiene en **40,0 % de implementación**, **6,7 % de certificación** y
  **NO-GO**.

Actualización 2026-08-01, candidato `edd2bdac`:

- El barrido global protegido del candidato anterior recorrio 618 vistas,
  11.102 controles y 97 clics seguros: cero excepciones y HTTP 5xx. Aislo como
  defectos reales el 403 del reporte de aseo y dos fotos locales antiguas
  ausentes en la red social.
- El middleware de autoservicio ahora propaga el rol resuelto en servidor sin
  confiar en cabeceras del cliente. La red social deja de entregar referencias
  locales inexistentes sin modificar los datos historicos y rechaza traversal.
- `go test ./...`, pruebas enfocadas y `go vet` aprobaron. El workflow inmutable
  `30732433840` completo build, Trivy, SBOM, publicacion y Compose, y los cuatro
  digests se promovieron solo a staging tras respaldar ambas bases.
- La repeticion autenticada y visual aprobo 4/4 vistas responsive, 168
  controles, cero errores HTTP/consola/excepciones e imagenes rotas. El boton
  `Consultar` respondio 200 y mostro `Reporte actualizado.`; una discrepancia
  query/cabecera de empresa fue rechazada con 400.
- Produccion conservo sus imagenes y salud. Se retiraron 40 imagenes GHCR no
  usadas manteniendo candidato y rollback inmediato; el disco paso de 73 % a
  58 %. Staging conserva maximo 20 sesiones por identidad.
- P109-004 sigue parcial por acciones mutantes, roles, segunda identidad no
  global A/B y firma del alcance.
- P109-011 queda aprobada sobre el mismo candidato: base vacia con 337/49
  tablas, segunda pasada idempotente, drift fail-closed, fallo transaccional,
  API anterior compatible, restore de dos bases/cinco dominios/dos replicas y
  rollback coordinado de datos y archivos. El RTO total fue 48 s, el rollback
  24 s y el RPO 5.466 s, dentro de los objetivos publicados de 2 h/24 h; la
  limpieza termino con cero recursos efimeros.
- El estado pasa a **43,3 % de implementacion** (`2 aprobadas + 9 parciales`),
  **6,7 % de certificacion del candidato** y **NO-GO**. La certificacion no
  suma P109-000 para `edd2bdac` porque el usuario pidio no crear PR y esa fase
  exige integracion revisada/fusionada; el 6,7 % actual corresponde a P109-011.
- P109-012 agoto el cierre automatizable: el preflight profesional completo
  aprobo 20/20 compuertas, incluido Go, seguridad, permisos, roles, pagos,
  migraciones, UX, documentacion, observabilidad, hardening y `diff-check`.
  Permanece parcial solo por ensayo/capacitacion/firma de otra persona y por
  definir contactos y horario del piloto; por ello no altera el porcentaje.

Actualizacion 2026-08-02, candidato `7c47d4df`:

- P109-006 queda aprobada. Cuatro identidades completaron invitacion, contrasena,
  contrato y login normal; el candidato final ejecuto cuatro cajas, cinco ventas,
  efectivo, debito, Nequi, pago mixto, stock compartido, doble pago, devolucion,
  descuento y cierres conciliados. Las cuatro cajas terminaron con diferencia e
  incidencia cero y PostgreSQL con cero esperas por lock.
- La vista historica mostro el cierre mixto con 150 en efectivo, 50 en otros y
  150 esperados, sin error visible ni desbordamiento de pagina.
- La carga operativa de concurrencia 10 termino 100/100, p50 139 ms, p95 289 ms,
  p99 326 ms y maximo 333 ms. Una rafaga fuera del perfil, de 100 solicitudes a
  la vez, termino sin errores pero elevo p95 a 3.667 ms; el servicio recupero el
  SLO sin reinicio y permanecio saludable.
- Se corrigieron tres defectos descubiertos por el flujo: permiso de devolucion
  del cajero, mutacion de items cerrados y sesiones residuales tras desactivar.
  La prueba final obtuvo HTTP 200 al devolver antes del pago, HTTP 409 despues
  del pago y HTTP 401 en las cuatro sesiones tras la desactivacion. Cero cuentas
  QA activas, sesiones QA activas o cajas QA abiertas quedaron al terminar.
- El workflow `30735137007` aprobo el SHA exacto y los cuatro digests se
  promovieron solo a staging. Produccion conservo su API local anterior y salud.
- El estado pasa a **50,0 % de implementacion** (`3 aprobadas + 9 parciales`, de
  15 fases). La **certificacion del candidato exacto es 6,7 %**, correspondiente
  a P109-006; las aprobaciones historicas de otros digests no se suman al exacto.
  El veredicto general permanece **NO-GO** por las demas compuertas P0/P1.

Actualizacion 2026-08-08, seguridad dinamica del candidato `7c47d4df`:

- P109-009 repitio DAST seguro autenticado y no autenticado en staging para PCS:
  lectura empresarial sin sesion 401, mutacion autenticada sin CSRF 403, origen
  externo con token CSRF 403, payload same-origin vacio 400 sin escritura,
  lectura autenticada 200 y logout seguido de 401. El preflight desde origen
  externo no expuso `Access-Control-Allow-Origin`; las cabeceras HSTS, CSP,
  nosniff, frame, referrer, permissions y no-store estuvieron presentes.
- La vista autenticada del panel no presento errores de consola; las dos sesiones
  de prueba se cerraron por el flujo oficial. `security_audit.mjs` aprobo y el
  inventario detecto 204/204 wrappers empresariales sin revision manual.
- P109-009 sigue parcial: CSP aun usa `unsafe-inline` y falta el aislamiento A/B
  de dos identidades no globales, ademas de DAST completo para XSS, SSRF, subidas,
  archivos, cache, jobs, IA, exportaciones y descargas. No se modifico produccion.
- No se aprueba una fase adicional: el estado se conserva en **50,0 % de
  implementacion**, **6,7 % de certificacion del candidato exacto** y **NO-GO**.

Actualizacion 2026-08-08, preflight completo del candidato local:

- `scripts/profesional_preflight.ps1 -Full` aprobo las 22 compuertas locales:
  parseo PowerShell, 72 JS, modulos, seguridad, permisos/licencias, OpenAPI,
  observabilidad, migraciones, QA de modulos/roles/pagos, soporte,
  anonimizacion, SLO/SLA, hardening, UX, documentacion, Compose, Go completo y
  `git diff --check`. El inventario mantuvo 204/204 wrappers empresariales sin
  revision manual.
- Es evidencia automatizada local, no un despliegue ni una promocion. P109-012
  conserva sus requisitos humanos de capacitacion, ensayo, contactos, horario y
  firma del piloto; no se incrementa el porcentaje ni se modifica produccion.

Actualizacion 2026-08-08, CSP estricta preparada para observacion:

- Se agrego `PCS_CSP_REPORT_ONLY_STRICT`. Con valor `1`, solo la cabecera
  `Content-Security-Policy-Report-Only` omite `unsafe-inline` de scripts y
  estilos; la CSP aplicada mantiene compatibilidad. El valor por defecto es
  `0` en Compose y ejemplos, por lo que no altera staging ni produccion.
- Pruebas de cabeceras, `go vet ./utils` y `go test ./... -count=1` aprobaron.
  El siguiente paso es publicar un candidato, activarlo exclusivamente en
  staging y revisar el lote visual/funcional de violaciones antes de enforcement.
- P109-009 sigue parcial y el porcentaje no cambia: el inventario actual sigue
  encontrando 191 paginas con scripts embebidos y 173 con estilos embebidos,
  ademas de la matriz A/B y DAST hostil pendientes.

Actualizacion 2026-08-08, correccion transversal de CSP estatica:

- La primera promocion de observacion confirmo que el backend de staging tomo
  `PCS_CSP_REPORT_ONLY_STRICT=1`, pero revelo que las paginas estaticas seguian
  entregando una cabecera Report-Only heredada con `unsafe-inline` desde Nginx.
- Se corrigio el include estatico para conservar la CSP aplicada compatible y
  retirar compatibilidad inline solo de Report-Only, con origenes explicitos.
  Una prueba evita que el include vuelva a usar destinos amplios o inline.
- La correccion local aprobo prueba de recursos estaticos, pruebas de cabeceras,
  vet y diff-check. Falta su candidato inmutable y su comprobacion visual en
  staging; P109-009 continua parcial y el porcentaje no cambia.

Actualizacion 2026-08-08, hallazgo de consola en panel autenticado:

- La revision visual de PCS en staging mostro el panel estable y organizado,
  pero la consola reporto un `MutationObserver.observe` con destino no nodo.
- Se agrego una guardia de tipo antes de observar el documento. Falta validar
  esta correccion en el siguiente candidato inmutable de staging; P109-009
  sigue parcial y no se incrementa el porcentaje.

Actualizacion 2026-08-08, candidato CSP `a408bb62` en staging:

- El workflow inmutable `31243348139` aprobo build, escaneo, SBOM, digests y
  Compose. La promocion exclusiva de staging quedo sana y uso respaldo previo
  verificable de PostgreSQL; produccion no fue editada ni reiniciada.
- Las cabeceras reales ya separan compatibilidad y observacion: la aplicada
  conserva `unsafe-inline`; `Content-Security-Policy-Report-Only` no contiene
  ni `unsafe-inline` ni el esquema amplio `https:`. La compuerta de cabeceras
  CSP queda aprobada.
- El login y panel PCS se vieron correctamente, pero el navegador interno
  repitio un `MutationObserver.observe` sin traza atribuible pese a que el
  recurso servido contiene la guardia. P109-009 continua parcial y el porcentaje
  no cambia hasta atribuirlo y completar DAST hostil, A/B no global y la
  migracion de 191 scripts/173 estilos embebidos.
- La implementacion consolidada se mantiene en **50,0 %** (`3 aprobadas + 9
  parciales`, de 15). La certificacion del candidato exacto `a408bb62` es
  **0,0 %**: ninguna fase completa fue reejecutada sobre esos digests; el
  **6,7 %** historico corresponde al candidato `7c47d4df` y no se transfiere.
  El veredicto sigue siendo **NO-GO**.

Actualizacion 2026-08-08, defensa completa de observadores del menu:

- Se revisaron los tres observadores activos en el panel administrativo y se
  agregaron guardias de nodo para tema, notificaciones e iconos. Una prueba
  estatica exige las tres condiciones antes de `observe`.
- Las pruebas Go y vet enfocadas aprobaron. Falta el release inmutable y la
  repeticion visual en staging para decidir si el diagnostico restante procede
  de PCS o del navegador interno; P109-009 y los porcentajes no cambian.

Actualizacion 2026-08-08, candidato `c8094f5b` visualmente limpio:

- El workflow `31243743197` aprobo y el candidato se promovio por digest solo
  a staging despues de respaldo verificable. Salud, CSP aplicada/Report-Only y
  las cuatro imagenes activas aprobaron.
- Login PCS y panel Super Administrador mostraron menu, metricas y controles
  sin recortes; la consola termino con 0 errores y 0 advertencias. El hallazgo
  de `MutationObserver` queda resuelto para este candidato.
- P109-009 sigue parcial exclusivamente por DAST hostil, A/B no global y
  migracion de scripts/estilos embebidos. La implementacion permanece en
  **50,0 %**, la certificacion exacta no recibe una fase completa y el estado
  sigue **NO-GO**.

Actualizacion 2026-08-08, carga segura y preflight de `c8094f5b`:

- La carga autenticada de solo lectura completo 30/30 recargas con concurrencia
  5, p50 840 ms y p95 1.245 ms, sin mutaciones de negocio. El navegador interno
  emitio eventos de observador sin traza durante recargas repetidas, por lo que
  no se atribuyen a PCS ni se considera una consola limpia bajo carga.
- El preflight profesional `-Full` aprobo 22/22 compuertas del arbol actual,
  incluidas sintaxis JavaScript, seguridad, permisos, Compose y Go completo.
- P109-010 y P109-012 siguen parciales: faltan carga HTTP autenticada concluyente,
  alertas/receptor externo, capacitación, responsables y firma humana. No cambia
  el porcentaje ni el veredicto **NO-GO**.

Actualizacion 2026-08-08, invariantes CxP en staging:

- El candidato de staging confirma aplicada la migración atómica CxP y columnas
  monetarias `NUMERIC(18,2)`. PCS tiene 3 obligaciones canónicas, 0 históricas
  y 5 pagos sin relaciones huérfanas, montos no positivos ni saldos negativos.
- Las pruebas enfocadas de CxP aprobaron; la integración PostgreSQL local quedó
  explícitamente omitida por no disponer de `PCS_TEST_POSTGRES_DSN` aislado.
- P109-001 continúa parcial por A/B, carrera real reversible y conciliación con
  contador. Los porcentajes y el estado **NO-GO** no cambian.

Actualizacion 2026-08-08, contratos IA y soportes:

- Pruebas enfocadas aprobaron IA cerrada por defecto, redacción, revisión y
  confirmación humana, JSON inválido, degradación de proveedor y doble clic,
  además del proveedor canónico para CxP.
- No se invocó proveedor ni se creó cartera/pago desde interfaz. P109-002 sigue
  parcial hasta flujo autenticado reversible, proveedor disponible y A/B no
  global. No cambia el porcentaje ni el estado **NO-GO**.

Actualizacion 2026-08-08, migracion y restore exactos de `c8094f5b`:

- El restore aislado del snapshot `20260808_031501` aprobo API/migrador del
  candidato, health/readiness 200, dos bases, cinco tablas, 28 filas criticas
  PCS, cuatro endpoints protegidos, inventario privado consistente y rol runtime
  sin DDL. RTO 23 s y RPO observado 12.287 s permanecen dentro de los objetivos
  publicados de 2 h/24 h.
- La base vacia aprobo tres rechazos sin DDL, doble pasada, drift fail-closed
  con recuperacion, cinco controles de rollback transaccional y cuatro de
  compatibilidad hacia atras. Esquema final: 337/49 tablas y 19/10 migraciones;
  la limpieza termino sin recursos efimeros.
- P109-011 queda **aprobada para este candidato**. El avance verificable pasa a
  **53,3 % de implementacion** (`4 aprobadas + 8 parciales`, de 15) y a
  **6,7 % de certificacion del candidato exacto**, por una sola fase completa.
  El veredicto permanece **NO-GO**: ninguna otra compuerta P0/P1 recibe credito
  por esta evidencia.

Actualizacion 2026-08-08, revision visual autenticada de documentos:

- En PCS staging, Facturas electronicas mostro filtros, KPIs y 11 ventas en
  filas/columnas ordenadas. La accion Visualizar confirmo apertura de la vista;
  no se emitio, anulo, descargo ni imprimio un documento.
- En movil 390 x 844 no hubo overflow horizontal de pagina; la tabla de 1.257 px
  conservo scroll propio dentro de 326 px. Escritorio y movil no mostraron
  errores ni advertencias de consola durante el recorrido.
- P109-005 sigue parcial por los formatos reales restantes, tableta, accesibilidad
  asistida, ventana de previsualizacion completa e impresion fisica. No cambia
  el porcentaje ni el veredicto **NO-GO**.

Actualizacion 2026-08-08, restore con replicas y rollback de `c8094f5b`:

- El snapshot `20260808_031501` aprobo restore de bases/volumen privado, dos
  replicas, cinco negativos de archivos, inventario sin huerfanos y rollback
  coordinado de datos, archivo y sesiones. RTO total 52 s, rollback 26 s y RPO
  observado 12.980 s quedaron dentro de los objetivos publicados de 2 h/24 h;
  no quedaron recursos efimeros.
- P109-008 sigue parcial por cuota, retencion, borrado/recuperacion, antivirus
  y A/B no global. No cambia el **53,3 %** ni el veredicto **NO-GO**.

Actualizacion 2026-08-08, cuota privada de soportes CxP/IA:

- El candidato `516de42e` incorpora la cuota empresarial a los adjuntos privados de
  `soportes_compras_ia`: suma almacenamiento publico y privado por `empresa_id`,
  aplica limite/maximo/bloqueo corporativos y responde HTTP 507 saneado ante
  exceso. La ruta, permisos y persistencia existente no cambian.
- Pruebas unitarias cubren limite, switches, maximo por archivo y aislamiento
  de bytes; `go test ./...`, `go vet ./...`, `diff --check` y el preflight de
  22 compuertas aprobaron. El workflow `31245235398` publico el candidato y
  staging recibio los digests exactos tras respaldo, sin tocar produccion.
- Una restauracion efimera del candidato, con cuota de 1 MiB solo en esa copia,
  rechazo por HTTP 507 un soporte de 2 MiB y dejo cero filas de prueba; al
  cierre se verifico la limpieza de sus recursos Docker y temporales.
- El candidato posterior `44610128` serializa esa seccion critica con candado
  asesor PostgreSQL por empresa. Dos replicas restauradas enviaron en paralelo
  soportes de 700 KiB con cuota aislada de 1 MiB: una inserto una fila y la otra
  fue rechazada con HTTP 507, sin sobregiro ni segunda fila. Sus recursos se
  limpiaron y staging se mantuvo saludable con los digests exactos.
- P109-008 sigue parcial por retencion, borrado/recuperacion, antivirus y A/B
  no global. El porcentaje y **NO-GO** no cambian.

Actualizacion 2026-08-08, prueba real PCS y colision de ventas reutilizables:

- El despliegue fusionado `bad1e80d` quedo saludable en produccion; la factura
  historica `1PCS4` recupero por flujo oficial el cliente de su venta sin crear
  otro consecutivo ni reenviar correo.
- El reenvio unico autorizado de `1PCS4` recibio HTTP 200 con `Regla 90:
  Documento procesado anteriormente`. El CUFE calculado por el reintento no
  aparece en la consulta publica DIAN y la segunda consulta quedo limitada por
  CAPTCHA; por seguridad fiscal el documento no se volvio a reenviar ni se
  marco localmente como aceptado.
- Una segunda venta real autorizada de una menta por COP 100 demostro un P0: la
  caja QA-FE reutiliza el carrito 117 y la clave del comprobante seguia basada
  solo en el ID del carrito, por lo que el upsert idempotente absorbio el nuevo
  cierre en la venta anterior.
- El candidato local `22359cf9` agrega una identidad estable por `pagado_en` a
  cada cierre, conserva idempotencia del mismo pago y hereda esa identidad al
  convertir la venta en factura. Pruebas enfocadas, paquetes `handlers`/`db`,
  `go vet` y el preflight profesional completo aprobaron. La carrera local no
  pudo ejecutarse porque Windows no tiene compilador C; no se oculto como PASS.
- P109-006 y P109-007 no se aprueban: el candidato requiere promocion controlada
  y repeticion real antes de consumir otro consecutivo DIAN. El avance permanece
  en **53,3 % de implementacion**, la certificacion del nuevo candidato es
  **0 %** y el veredicto continua **NO-GO**.

La evidencia histórica P108 sirve como línea base y evita repetir pruebas
independientes del artefacto, pero no aumenta automáticamente estas cifras.

Actualización 2026-08-08, descarga privada CxP/IA y SSRF por redirección:

- La descarga real del soporte privado `SCI-0004` aprobó autenticación,
  pertenencia a `empresa_id=12`, adjunto forzado, `no-store`, `nosniff` y cero
  fuga de ruta; sin sesión respondió 401, el cruce a empresa 53 respondió 404 y
  los identificadores ausente/manipulado respondieron 400.
- La auditoría local cerró una brecha de defensa en profundidad del callback
  OnlyOffice: cada redirección vuelve a validar esquema, host y puerto contra el
  Document Server configurado. Las pruebas demuestran que el mismo origen
  continúa funcionando y que un segundo origen no recibe la solicitud.
- El arreglo permanece local sin PR ni despliegue. P109-009 sigue parcial y el
  avance permanece en **53,3 % de implementación**, **0 % de certificación del
  arreglo local** y **NO-GO**.

Actualización 2026-08-08, salidas HTTP DIAN e integraciones:

- Se cerró localmente SSRF en probes, envio/acuse/reconexion DIAN y operaciones
  SOAP, además de despacho/health de proveedor fiscal: solo HTTP(S) público,
  resolución DNS validada, cero redes privadas o especiales y redirecciones
  limitadas al mismo origen.
- Los overrides DIAN deben conservar el origen configurado para la empresa y la
  clasificación oficial exige HTTPS sobre `dian.gov.co` o subdominio exacto.
  Dominios parecidos, loopback y metadata cloud quedan rechazados.
- Las pruebas negativas, los sets DIAN de contrato, Go completo, vet y preflight
  estricto aprobaron sin emitir documentos ni modificar datos. El cambio sigue
  local sin PR ni despliegue; P109-009 permanece parcial y el Plan conserva
  **53,3 %**, **0 % de certificación del arreglo local** y **NO-GO**.

Actualización 2026-08-08, papelera recuperable de soportes CxP/IA:

- Se implementó localmente eliminación lógica y recuperación con motivo,
  actor, auditoría y bloqueo transaccional por `empresa_id`; el archivo y el
  historial no se destruyen.
- Los soportes eliminados no pueden descargarse, editarse, procesarse con IA ni
  contabilizarse. La recuperación falla si existe duplicado activo y los
  soportes contabilizados/convertidos no pueden enviarse a papelera.
- La UI separa Activos/Papelera y habilita únicamente las acciones compatibles.
  Las pruebas Go y contratos de interfaz aprobaron, pero falta desplegar el
  candidato y repetir el ciclo autenticado/visual en PCS y en empresa A/B.
- P109-002 y P109-008 continúan parciales. El avance permanece en **53,3 % de
  implementación**, **0 % de certificación del arreglo local** y **NO-GO**.

Actualización 2026-08-08, adjuntos hostiles, cancelación IA y retención:

- La admisión local ahora contrasta firma/extensión, fija MIME y rechaza XML
  activo, DTD, entidades, instrucciones y documentos mal formados antes de
  escribir. Un fallo posterior de base limpia el archivo nuevo confinado.
- `Cancelar IA` aborta navegador y transporte Responses mediante el contexto de
  la petición; una prueba con proveedor lento local confirmó `context.Canceled`.
- La vista previa de retención filtra por empresa solo eliminados antiguos no
  contabilizados y calcula bytes sin ejecutar purga.
- Esto no sustituye antivirus, carrera Linux, proveedor real, purga certificada
  ni UAT PCS/A-B. P109-002, P109-008 y P109-009 continúan parciales; el avance
  permanece **53,3 %**, la certificación del cambio local **0 %** y **NO-GO**.

Actualización 2026-08-08, protocolo antivirus para soportes:

- Se integró localmente `clamd` por INSTREAM con deadlines, respuesta limitada,
  rechazo de malware antes de escritura y modo obligatorio fail-closed.
- Un servidor TCP simulado aprobó limpio, `FOUND`, modo opcional y caída
  obligatoria. Compose y ejemplos exponen la configuración apagada por defecto.
- Falta desplegar/actualizar firmas y comprobar ClamAV real en staging, carga
  EICAR controlada, métricas y recuperación ante caída. P109-008/P109-009 siguen
  parciales; permanece **53,3 %**, certificación local **0 %** y **NO-GO**.

Actualización 2026-08-09, depuración segura de soportes CxP/IA:

- Se implementó localmente la transición final `eliminado -> purgado` para un
  soporte no contabilizado que ya cumple retención. Exige Delete, motivo,
  antigüedad y confirmación exacta del código bajo bloqueo por `empresa_id`.
- La fila y sus eventos se conservan; el archivo se procesa con cuarentena,
  rollback si la transacción falla y eliminación solo después del commit. Un
  registro purgado no se descarga, opera ni recupera.
- Se cerró la fabricación de referencias privadas por JSON y cada consumo del
  archivo exige el prefijo exacto de la empresa. Las pruebas locales cubren
  A/B, commit/rollback, fechas y contrato PostgreSQL preparado.
- No se ejecutó purga real, staging ni restore posterior; P109-008 continúa
  parcial. El avance formal permanece en **53,3 % de implementación**, **0 % de
  certificación del cambio local** y **NO-GO**.
- La depuración se endureció como saga `eliminado -> purga_pendiente -> purgado`:
  inicio y final son idempotentes y un reintento recupera caídas en cada frontera
  archivo/base. Cuarentenas múltiples se rechazan sin elegir arbitrariamente.
- `cuarentena_preview` ofrece diagnóstico Read por empresa con registros,
  archivos y bytes, sin nombres, rutas ni mutación. Esto mejora operación local,
  pero no sustituye la prueba de caída real, restore, réplica A/B ni staging.
- La saga incorpora advisory lock por empresa entre replicas, replay idempotente
  incluso después de finalizar y alerta de pendientes envejecidos. El runbook
  oficial prohíbe SQL/borrado manual y exige backup antes de una recuperación.

Actualización 2026-08-09, observabilidad de la depuración CxP/IA:

- `/metrics` publica únicamente agregados globales de sagas pendientes,
  pendientes por al menos 15 minutos y finalizadas, con estado de consulta y
  sin etiquetas de empresa, usuario, soporte, nombre o ruta privada.
- Prometheus incorpora la alerta crítica `PCSSoporteIAPurgaVencida`, Grafana
  presenta pendientes/vencidas y el runbook dirige al flujo oficial reanudable.
- Pruebas locales de render, contrato y vet aprobaron. Falta desplegar, provocar
  y resolver una saga vencida en staging y comprobar entrega externa; P109-010
  sigue parcial, el avance permanece **53,3 %** y el veredicto **NO-GO**.

Actualización 2026-08-09, telemetría del antivirus de soportes:

- El scanner clamd registra contadores atomicos agregados para limpio, malware,
  indisponible y omitido, además de modo obligatorio/configurado. Las métricas
  no contienen dimensiones empresariales, usuarios, archivos ni rutas.
- Prometheus alerta configuración fail-closed incompleta, indisponibilidad,
  omisión y malware bloqueado; Grafana muestra resultados recientes por job.
- Pruebas locales cubren los cuatro resultados y 64 llamadas simultáneas. Falta
  clamd real con firmas, EICAR controlada, recuperación y recepción externa en
  staging; P109-008/P109-010 siguen parciales, permanece **53,3 %** y **NO-GO**.

Actualización 2026-08-09, evals adversariales de extracción IA:

- El parser limita tamaño, claves, tipos, líneas, textos y rangos; rechaza NaN,
  infinito, negativos, confianza fuera de 0..1 y fechas inválidas. Faltantes o
  descuadres obligan revisión humana antes de cualquier aprobación.
- Se publican seis resultados agregados de extracción y alertas para proveedor,
  respuesta inválida y persistencia, sin empresa, documento, prompt o respuesta.
- Casos adversariales y 64 registros simultáneos aprobaron localmente. Falta el
  candidato desplegado, proveedor real y A/B; P109-002/P109-010 siguen parciales,
  el avance permanece **53,3 %** y el veredicto **NO-GO**.

Actualización 2026-08-09, integridad de archivos privados de soportes:

- Descarga y envío a IA verifican SHA-256, archivo regular y tamaño acotado. Un
  contenido alterado o hash inválido falla antes de servir o llamar proveedor.
- La respuesta fuerza attachment, MIME canónico/binario y cabeceras sandbox,
  nosniff, same-origin, no-referrer, DENY y no-store. Métrica, alerta y panel no
  contienen empresa, soporte, hash, nombre ni ruta.
- El incidente usa `FOR UPDATE` por empresa, invalida una aprobación abierta y
  registra evento minimizado en la misma transacción; terminales se preservan.
- Las pruebas locales de bytes, stream, MIME y contrato aprobaron. Falta fixture
  PostgreSQL reversible desplegado y A/B; el ensayo quedó preparado pero hizo
  SKIP sin DSN. P109-009 sigue parcial, permanece **53,3 %** y **NO-GO**.

## 9. Siguiente orden para Terra alto

P109-000 ya está aprobada para `ea9642dd...` y P109-011 para `c8094f5b`;
Terra no debe reconstruirlas ni repetirlas sin un cambio de código. Debe
continuar así:

1. confirmar que staging conserva los cuatro digests registrados y producción
   conserva sus imágenes anteriores;
2. conservar la matriz A/B ya aprobada con identidad no global y cerrar
   P109-001 únicamente después de la carrera reversible y la conciliación del
   contador autorizado;
3. desplegar el candidato aislado y repetir en P109-002/P109-008 la papelera,
   descarga bloqueada, recuperación, duplicado y botones IA en PCS y empresa
   A/B; completar además confirmación/cancelación, degradación y evals;
4. ejecutar UAT contable/fiscal de P109-003 con contador autorizado;
5. completar acciones mutantes de P109-004 por rol y flujo oficial;
6. continuar cuatro cajas, DIAN, receptor externo y piloto solo cuando
   sus credenciales, identidades, hardware o ventanas estén disponibles;
7. registrar evidencia y recalcular las dos cifras sin dar crédito completo a
   una fase parcial.

No debe saltar a producción ni llamar 100 % a un bloque con pendientes.

Actualización 2026-08-09, pruebas reales PCS y correcciones del candidato local:

- El barrido autenticado cubrió 48 variantes de 24 rutas críticas y aisló
  desbordamiento móvil DIAN, CSP de Mailu y permiso/licencia de Centro IA.
- Las impresiones reales de `1PCS5` en Carta, compacta y POS demostraron QR y
  datos fiscales, pero descubrieron recursos relativos rotos dentro de blobs;
  el candidato ya resuelve hoja, logos y QR contra el origen seguro.
- La captura CxP/IA controlada permitió extracción y edición y terminó rechazada
  sin cuenta por pagar, pago o asiento.
- El cifrado legado del buzón PCS se renovó por el flujo oficial. El candidato
  corrige la renovación Mailu para actualizar por PATCH antes de crear por POST.
- DIAN presenta ambiente/rango reales, pero el indicador de producción local se
  perdía cuando `estado_dian` era sobrescrito por un envío. El candidato ya lo
  separa en un indicador persistente, migra evidencia histórica y bloquea la
  autoactivación desde el CRUD; la compuerta continúa hasta desplegar, reconciliar
  PCS y repetir un envío/acuse controlado.
- Preflight profesional completo y pruebas enfocadas pasan. P109-005/P109-007
  siguen parciales porque falta desplegar el candidato, repetir visual/IMAP,
  comprobar empresa A/B y cerrar el estado DIAN. El avance permanece **53,3 %
  de implementación**, **0 % de certificación del arreglo local** y **NO-GO**.

Actualización 2026-08-09, candidato `17c55dd8` y prioridad web:

- El primer digest detectó correctamente que el indicador DIAN dependía del
  bootstrap heredado y quedó invalidado. El segundo candidato incorporó la
  migración inmutable `20260809-001-dian-local-production-flag-v1`; workflow,
  respaldo, promoción por cuatro digests, ledger, checksum, constraint,
  backfill transaccional, salud y aislamiento de producción aprobaron.
- El barrido autenticado del mismo candidato cubrió 48 variantes de 24 rutas:
  46 aprobaron y las dos revisiones corresponden a la URL heredada inexistente
  `inventario.html`. Centro IA, correo y DIAN ya no reproducen los hallazgos
  anteriores. La impresión sintética volvió a aprobar 20/20 y se inspeccionó
  Carta, POS, QR y 96 filas.
- Por decisión explícita del propietario se prioriza terminar la aplicación
  web. La aplicación móvil nativa continúa formalmente aplazada y su código se
  conserva. Solo se mantiene como requisito de la web la usabilidad responsive
  mínima en navegador móvil; no se agregan funciones móviles al alcance del
  piloto.
- P109-000 queda aprobado para `17c55dd8`. P109-005 y P109-007 siguen parciales
  por documentos físicos/reales, Mailu/IMAP y la prueba fiscal controlada. El
  avance permanece **53,3 % de implementación**, la certificación del candidato
  sube a **6,7 %** y el veredicto continúa **NO-GO**.

Actualización 2026-08-09, CxP/IA autenticada en el candidato `17c55dd8`:

- PCS radicó `SCI-0009`, extrajo los datos mediante IA, permitió editar la
  lectura, vinculó proveedor, aprobó y rechazó con trazabilidad completa. La
  aprobación no se contabilizó.
- La repetición exacta quedó bloqueada como `SCI-0010` duplicado. Un segundo
  soporte `SCI-0011` confirmó que **Cancelar IA** aborta la petición y conserva
  `radicado`; luego se rechazó por el flujo oficial.
- La conciliación pasó de 8 a 11 soportes y mantuvo sin cambios 3 CxP, 5 pagos y
  5 movimientos. La bandeja se comprobó visualmente con filas, columnas,
  importes, estados y confianza organizados.
- La papelera visual permanece pendiente porque el navegador interno no soporta
  `prompt()`; una sesión administrativa independiente recibió 403. La empresa
  inexistente devolvió cero datos, pero no sustituye una identidad A/B real.
- P109-002 y P109-008 siguen parciales. El avance permanece **53,3 % de
  implementación**, **6,7 % de certificación del candidato** y **NO-GO**.

Actualización 2026-08-09, dialogos verificables de papelera CxP/IA:

- Se sustituyeron `confirm`/`prompt` nativos por un dialogo propio con foco,
  cancelacion, motivo obligatorio y confirmacion fuerte por codigo.
- El payload y el endpoint no cambian; permisos, retencion, auditoria y
  aislamiento por `empresa_id` permanecen bajo control del backend.
- Sintaxis, contratos y paquetes enfocados aprobaron localmente. P109-008 no
  recibe credito adicional hasta construir el candidato y repetir papelera y
  recuperacion visualmente en staging. Permanece **53,3 %**, **6,7 %** y
  **NO-GO**.

Actualización 2026-08-09, papelera visual en el candidato `34cfd852`:

- Release inmutable `31303668393`, preflight 22/22, respaldo de ambas bases y
  promoción por cuatro digests sin build en staging aprobaron.
- El nuevo diálogo validó motivo obligatorio, envió `SCI-0009` a Papelera y
  mostró el bloqueo seguro de recuperación por el duplicado activo `SCI-0010`.
- Tras enviar también el duplicado QA a Papelera, `SCI-0009` se recuperó con
  auditoría. CxP/pagos/movimientos permanecieron 3/5/5.
- P109-008 continúa parcial por retención, caída, antivirus y A/B. El avance de
  implementación permanece **53,3 %** y el veredicto **NO-GO**. La certificación
  formal del nuevo SHA se recalculará al repetir la compuerta completa P109-000.

Actualización 2026-08-09, verificación ampliada de `34cfd852`:

- Digests efectivos, migrador 0, salud/listo, worker y producción sana fueron
  comprobados. E2E `31304164994` aprobó escritorio/móvil sin hallazgos; impresión
  20/20 y matrices de roles/pagos quedaron `ok`.
- El endpoint rechazó con 401 eliminar/restaurar sin sesión y la empresa
  inexistente. No hubo mutaciones.
- P109-000 sigue parcial: falta PR/fusión y los drills exactos de base vacía y
  upgrade. El avance permanece **53,3 %**, la certificación formal del SHA
  exacto continúa **0 %** hasta cerrar la compuerta y el veredicto es **NO-GO**.

Actualización 2026-08-09, drills exactos de migración de `34cfd852`:

- El digest inmutable del migrador aprobó base vacía, segunda pasada
  idempotente, fallo cerrado por drift de checksum y recuperación. El esquema
  terminó con 337 tablas empresariales, 49 globales y ledger 20/10.
- El upgrade de una copia lógica consistente de staging conservó 350 tablas
  empresariales y 59 globales antes/después; ambas pasadas aplicaron cero
  migraciones nuevas y el rol runtime continuó sin DDL.
- La limpieza dejó cero contenedores, redes y volúmenes efímeros. Staging
  conservó salud/listo y sus digests; PCS mantuvo CxP/pagos/movimientos 3/5/5;
  producción permaneció saludable y sin despliegue.
- La parte técnica pendiente de P109-000 queda aprobada. La fase continúa
  formalmente parcial porque el candidato no tiene PR/revisión/fusión a `main`,
  conforme a la instrucción de no crear PR. El avance permanece **53,3 % de
  implementación**, **0 % de certificación formal del SHA** y **NO-GO**.

Actualización 2026-08-09, retención real y antivirus de `34cfd852`:

- En PCS, la vista previa con retención de un día informó cero candidatos. El
  intento de depurar `SCI-0010`, con motivo y código fuerte correctos, fue
  rechazado por no cumplir la retención; visualmente conservó estado eliminado.
- PostgreSQL confirmó cero eventos de purga y mantuvo CxP/pagos/movimientos
  3/5/5. La sesión autenticada se cerró al terminar.
- Las métricas reales muestran antivirus `required=0`, `configured=0`, cero
  escaneos y ningún servicio clamd en el VPS. Se registra como bloqueo, no PASS.
- P109-008 y P109-010 siguen parciales por depuración vencida/reanudable,
  antivirus real, A/B y señales operativas restantes. El avance permanece
  **53,3 %**, la certificación formal del SHA **0 %** y el veredicto **NO-GO**.

Actualización 2026-08-09, identidad A/B no global e integración PostgreSQL:

- Una identidad empresarial temporal recuperó su acceso mediante el correo
  oficial y probó soportes IA, finanzas, contabilidad, DIAN y documentos. PCS
  cargó sus vistas; `empresa_id=7` fue rechazada fuera del alcance y no mostró
  filas ni datos de PCS.
- El intento visible de enviar `SCI-0011` a papelera fue rechazado por permiso
  efectivo; la consulta posterior conservó el soporte activo/rechazado. La
  cuenta quedó inactiva por HTTP 204 y su sesión previa respondió unauthorized.
- PostgreSQL 16 efímero aprobó papelera, restauración, duplicado, retención,
  bloqueo contable, incidente de integridad, purga reanudable/idempotente,
  precisión monetaria y bandera DIAN por empresa. El fixture usa los esquemas
  reales de migración y las tablas funcionales del ensayo son temporales.
- Túnel y contenedor efímeros fueron retirados; producción no cambió. La prueba
  A/B deja de ser pendiente, pero P109-001 conserva carrera/conciliación humana,
  P109-002 proveedor/degradación y P109-008 antivirus/purga vencida/restore.
  La compuerta profesional aprobó sus 20 controles y la regresión Go completa
  aprobó pruebas y vet.
  Ninguna fase parcial cerró todos sus criterios: el avance formal permanece
  **53,3 % de implementación**, **0 % de certificación formal del SHA** y
  **NO-GO**.

Actualización 2026-08-09, durabilidad outbox y Centro IA:

- PostgreSQL 16 efímero aprobó deduplicación outbox, reclamación concurrente
  única, lease vencido, recuperación, dead-letter y pago CxP idempotente/A-B.
  El túnel, contenedor y listener temporal fueron retirados al finalizar.
- En staging PCS, el Centro IA cargó siete funciones, ejecutó diagnóstico y
  consulta no mutante en modo agente; el interruptor inició y terminó apagado,
  sin crear efectos financieros. La bandeja CxP/IA mostró su duplicado QA
  bloqueado y no editable.
- La revisión visual detectó Markdown literal en el resultado IA. El candidato
  local ya lo muestra con títulos/listas/énfasis mediante escape previo y sin
  dependencias; exige despliegue aislado y repetición visual antes de crédito.
- P109-002 y P109-010 permanecen parciales por criterios operativos externos y
  del candidato final. El avance formal continúa **53,3 %**, la certificación
  formal **0 %** y el veredicto **NO-GO**.

Actualización 2026-08-09, candidato `bb285968` en staging:

- El candidato aislado de Centro IA se construyó desde SHA exacto fuera del
  checkout productivo y dejó health/readiness y los cuatro servicios saludables.
  La revisión visual confirmó títulos, listas y énfasis seguros, sin Markdown
  literal ni HTML del proveedor interpretado.
- Se ejecutaron las siete funciones IA visibles como recomendaciones, con modo
  agente apagado al inicio y al cierre. Las invariantes PCS siguieron en 3 CxP,
  5 pagos y 5 movimientos; únicamente quedó un soporte demo duplicado bloqueado.
- La compuerta resolvió el runtime Node y Chrome instalado; el E2E completo
  registró 363 variantes antes del timeout externo y el dirigido aprobó 2/2
  vistas, 108 botones y cero errores. Impresión 20/20 y carga p95 411 ms/error
  0 % aprobaron.
- El manifest se corrigió para exigir el digest de frontend. P109-000/P109-004
  no se aprueban: faltan cuatro imágenes inmutables, fusión/revisión y el
  inventario E2E completo. El avance permanece **53,3 %**, certificación **0 %**
  y **NO-GO**.

Actualización 2026-08-09, candidato inmutable `31915619` promovido solo a staging:

- El workflow `31308770525` construyó una sola vez API, migrador, worker y
  frontend desde el SHA completo `31915619a74227216b9590b5268e036b3e6a51b4`.
  Trivy, los cuatro SBOM, publicación en GHCR y la validación de Compose por
  digest aprobaron en la misma ejecución.
- Los cuatro digests publicados se promovieron con
  `vps-staging-digest-up.sh`; el script no reconstruyó código. Staging quedó
  `health=ok`, `ready=ready`, API/worker/frontend saludables y migrador con
  exit code `0`. El ledger registra
  `platform:20260809-001-dian-local-production-flag-v1:applied` y la columna
  DIAN existe. Las imágenes y salud de producción fueron verificadas sin cambio.
- `qa_e2e_buttons.cjs` ya puede reanudar el inventario por lotes ordenados con
  offset/tamaño explícitos, y sus reportes guardan ambos datos. Esta mejora
  aún no es una ejecución total: P109-004 conserva pendiente la matriz de
  acciones mutantes, roles e IA y no recibe crédito adicional hasta completar
  todos los lotes autenticados.
- P109-000 conserva estado **parcial** por la política vigente de no crear PR:
  el SHA no está revisado ni fusionado a `main`. El avance de implementación
  permanece **53,3 %**, la certificación formal **0 %** y el veredicto
  **NO-GO**.

Actualización 2026-08-09, SHA actual `3676cc02` en staging:

- El workflow `31319557537` volvió a aprobar construcción única, Trivy, SBOM,
  publicación y Compose para el SHA que añade la reanudación E2E por lotes.
- Sus cuatro digests exactos fueron promovidos sin build. API, worker y frontend
  alcanzaron estado saludable; migrador terminó `0`; `/health` y `/ready` de
  staging aprobaron. Las tres imágenes y la salud de producción siguen intactas.
- No se altera la fórmula: P109-000 permanece parcial sin PR/fusión y el plan
  sigue en **53,3 % de implementación**, **0 % de certificación formal**,
  **NO-GO**.

Actualización 2026-08-09, capacidad VPS y revisión visual financiera:

- El panel staging informó 87 % de disco. La inspección aisló 17,2 GB de
  imágenes Docker sin uso y 2,5 GB de caché de build; no se tocaron volúmenes,
  bases ni respaldos. La limpieza recuperable liberó 28,2 GB y redujo el uso
  de `/` a 60 %. Staging y producción conservaron salud correcta.
- Con sesión oficial PCS se comprobó visualmente Finanzas/CxP: formulario,
  tabla de 11 cierres y controles de acciones aparecen organizados. A 390 px,
  el ancho de documento fue exactamente 390 px, hubo cero botones visibles sin
  etiqueta y cero errores de consola. Esta es evidencia no mutante y no
  sustituye las pruebas de cuatro cajas, roles ni conciliación contable.
- P109-005/P109-009 permanecen parciales; el avance formal continúa **53,3 %**
  y **NO-GO**.

Actualización 2026-08-09, hallazgo del barrido E2E de Renta IA:

- Los lotes autenticados detectaron dos POST bloqueados al abrir Renta IA. La
  causa fue un cálculo automático en el arranque de la pantalla. El candidato
  local elimina ese POST: la consulta financiera y cualquier uso de IA requieren
  el clic explícito del usuario. Debe construirse y repetirse el lote antes de
  acreditar la corrección.

Actualización 2026-08-09, barrido autenticado completo por lotes:

- Cuatro ejecuciones CI recorrieron las 309 rutas en escritorio y móvil: 618
  vistas, 11.258 botones inventariados y 75 clics seguros. La carga GET de 500
  solicitudes aprobó con p95 135 ms y 0 % de errores.
- Renta IA se repitió sobre el digest corregido `e51228dd`: 2/2 vistas `ok`,
  cero mutaciones bloqueadas y cero errores.
- Persisten seis escrituras automáticas de páginas públicas y dos respuestas
  502 de ayuda móvil. La repetición de páginas públicas aprobó 10/10 al separar
  y bloquear su telemetría; la repetición aislada de ayudas móviles aprobó y no
  reprodujo los 502 transitorios. P109-004 sigue **parcial** por acciones
  mutantes, roles y botones IA. El avance
  formal continúa **53,3 %**, **NO-GO**.

Actualización 2026-08-09, restore integral del snapshot vigente:

- El snapshot `20260809_031501` aprobó restore PostgreSQL aislado con los
  nueve tarballs obligatorios, dos bases PCS, cinco tablas críticas filtradas
  por empresa y tres checksums de soportes IA privados.
- El RTO observado fue 27 s y el RPO 50.544 s; el contenedor temporal se
  eliminó automáticamente. No se modificaron staging activo ni producción.
- P109-008 continúa parcial por réplica A/B, pérdida de réplica y rollback
  coordinado. No cambia la fórmula: **53,3 %**, certificación formal **0 %**
  y **NO-GO**.

Actualización 2026-08-09, Finanzas/CxP responsive en staging:

- La revisión autenticada de PCS validó los formularios financieros y las seis
  tablas CxP/CxC en móvil (390 px), sin desborde horizontal, sin botones sin
  etiqueta y sin errores de consola. La carga CxP con IA conserva revisión
  humana antes de guardar.
- La inspección fue solo lectura; P109-005 continúa parcial por dispositivos,
  impresión física, accesibilidad completa y roles. La fórmula permanece en
  **53,3 %**, certificación formal **0 %** y **NO-GO**.

Actualización 2026-08-09, sondeo público seguro de staging:

- Salud, readiness y portada pública respondieron HTTP 200; las seis cabeceras
  de endurecimiento aplicables estuvieron presentes. La portada pública no
  declara `Cache-Control`, observación que no se extrapola a rutas autenticadas.
- P109-009 continúa parcial por DAST hostil, aislamiento A/B y recursos
  autenticados. La fórmula permanece **53,3 %**, certificación formal **0 %**
  y **NO-GO**.

Actualización 2026-08-09, auditoría de observabilidad del candidato:

- El preflight aislado encontró que `runtime_state_log` era una alerta estática
  falsa: el registro vive en su handler y no necesariamente como archivo antes
  del arranque. El auditor ahora revisa ambas fuentes reales.
- El resto de la compuerta breve aprobó. P109-010 continúa parcial por alertas
  externas, responsables y ensayo antivirus real; el avance sigue **53,3 %**,
  certificación formal **0 %** y **NO-GO**.

Actualización 2026-08-09, compilación íntegra del candidato P109:

- La compilación de todos los paquetes y herramientas Go se completó por grupos
  y `go vet ./...` aprobó. El preflight estándar volvió a aprobar 20/20;
  el timeout de la orden global se mantiene explícitamente inconcluso.
- P109-012 conserva pendientes humanos de soporte/capacitación. No cambia la
  fórmula: **53,3 %**, certificación formal **0 %** y **NO-GO**.

Actualización 2026-08-09, restore de aplicación del digest activo:

- El candidato activo de staging restauró sus dos bases y API en recursos
  efímeros: health/ready 200, cinco tablas, 31 filas críticas PCS y archivos
  privados sin huérfanos. El rol runtime permaneció sin DDL.
- P109-008 sigue parcial porque falta la variante autenticada con réplica A/B,
  hostiles y rollback coordinado sobre este mismo candidato. La fórmula se
  mantiene en **53,3 %**, certificación formal **0 %** y **NO-GO**.

Actualización 2026-08-09, QA de impresión del candidato actual:

- Veinte formatos Carta/POS aprobaron PDF, QR, imágenes, filas, columnas y
  paginación; factura y recibo extensos alcanzaron seis páginas cada uno sin
  recortes.
- P109-005 conserva impresión física, dispositivos reales y accesibilidad como
  aceptación pendiente. La fórmula sigue **53,3 %**, certificación formal
  **0 %** y **NO-GO**.

Actualización 2026-08-09, impresión virtual real y prerrequisito de usuarios:

- Chrome imprimió la factura real PCS `1PCS6` en dos páginas A4 contra las APIs
  autenticadas. El candidato recuperó el logo oficial y dejó encabezado, cinco
  columnas, importes, CUFE, URL, observaciones y QR sin recortes; la regresión
  sintética aprobó 20/20 formatos.
- La repetición de cuatro usuarios no puede declararse: en el PCS servido existe
  una sola identidad confirmada activa. El buzón propio está provisionado, pero
  IMAP rechaza la consulta y el acceso SnappyMail termina 302/302/403. El arreglo
  de host SSO queda probado localmente y requiere despliegue antes de emitir y
  confirmar tres invitaciones nuevas.
- Con P109-008 ya aprobada técnicamente en snapshot aislado, el avance formal
  vigente es **56,7 % de implementación**, certificación formal **0 %** y
  **NO-GO**. P109-005 continúa parcial y P109-006 conserva su aprobación previa
  de staging sin extrapolarla al PCS servido actual.

Actualización 2026-08-09, repetición de cuatro cajas sobre el digest `349712fb`:

- El workflow inmutable `31334057174` publicó API, migrador, worker y frontend
  para el SHA completo; los cuatro digests se promovieron solo a staging.
  Health, ready y login aprobaron, mientras las imágenes de producción se
  conservaron sin cambios.
- Los cuatro usuarios temporales confirmados se activaron, recuperaron su clave
  mediante cuatro correos reales y abrieron cuatro sesiones independientes con
  cookies y CSRF distintos. Después abrieron cuatro cajas simultáneas, crearon
  cuatro ventas de COP 100, pagaron con efectivo, débito, Nequi y transferencia
  y recibieron cuatro comprobantes no fiscales.
- Los cuatro reintentos de pago devolvieron HTTP 409. Las cajas cerraron con
  teórico igual a físico y cero incidencias. Al desactivar los usuarios, las
  cuatro sesiones quedaron revocadas con HTTP 401.
- La vista servida de staging imprimió uno de los comprobantes nuevos en una
  página A4, 74.865 bytes y cero imágenes rotas; la inspección visual confirmó
  filas, cinco columnas, importes, logo y observaciones sin recortes.
- P109-006 queda aprobada para el digest exacto. La implementación permanece en
  **56,7 %** porque no cerró una fase nueva del plan; la certificación formal del
  candidato sube a **6,7 %**. P109-005, Mailu corporativo, UAT humano, piloto y
  las demás compuertas P0/P1 mantienen **NO-GO**.

Actualización 2026-08-09, aislamiento y carrera CxP del digest `349712fb`:

- P109-001 ejecutó una matriz A/B de lectura entre PCS y una segunda empresa
  activa: CxP, proveedores, reconciliación y recuperación outbox permanecieron
  separados; una CxP de PCS consultada desde la segunda empresa devolvió 404.
- En PCS se creó una obligación de ensayo de COP 2 por el flujo oficial. Dos
  pagos simultáneos con la misma clave devolvieron un solo pago/movimiento y un
  replay idempotente; dos claves distintas por el último COP resultaron en un
  pago HTTP 200 y un rechazo HTTP 409, con saldo final COP 0 y estado pagada.
- P109-001 conserva estado **parcial** por la conciliación/aceptación humana y
  la recuperación operativa de un evento elegible del mismo candidato. No se
  cierra una fase adicional: la implementación sigue en **56,7 %**, la
  certificación exacta en **6,7 %** y el veredicto en **NO-GO**.

Actualización 2026-08-09, candidato ClamAV de staging:

- Se preparó un overlay exclusivo de staging con ClamAV oficial fijado por
  digest, red interna, volumen de firmas, healthcheck y modo obligatorio
  fail-closed para soportes CxP/IA. Los scripts de staging y de promoción por
  digest exigen y arrancan ese servicio; producción no incorpora el overlay.
- Las pruebas Go simuladas aprobaron limpio, malware/EICAR, caída y métricas.
  La estación local no dispone de Docker/WSL, así que la prueba con firmas
  reales y alertas queda pendiente del despliegue aislado del candidato.
- No se cierra fase adicional: la implementación permanece en **56,7 %**, la
  certificación exacta en **6,7 %** y el estado en **NO-GO**.

Actualización 2026-08-11, prueba real fail-closed de ClamAV en staging:

- El candidato `e70a9406` está desplegado solo en staging con `clamav` sano,
  `/ready` 200 y modo obligatorio. La interfaz autenticada de PCS aceptó un
  archivo limpio, rechazó EICAR sin crear soporte y bloqueó una carga limpia
  cuando se detuvo únicamente el servicio antivirus.
- Al recuperar el healthcheck del servicio, una nueva carga limpia fue aceptada.
  Las métricas agregadas verificaron una ocurrencia limpia, una de malware y una
  de indisponibilidad, sin etiquetas empresariales ni nombres de adjuntos.
- P109-010 sigue **parcial**: falta confirmar entrega, deduplicación y
  resolución en el receptor externo de alertas, además de las demás señales P0,
  responsables y simulacro de incidente. La implementación continúa en
  **56,7 %**, certificación formal **6,7 %**, con **NO-GO**.

Actualización 2026-08-11, extracción IA controlada en PCS staging:

- `Extraer IA` se ejecutó por el flujo visible y autenticado sobre un soporte de
  prueba. La respuesta insuficiente no creó una CxP: quedó en revisión humana y
  el backend registró el resultado agregado correspondiente.
- La revisión humana editable se guardó y quedó auditada sin aprobación ni
  contabilización. P109-002 sigue parcial por los botones y escenarios IA que
  faltan, aislamiento A/B y las evaluaciones completas; la fórmula permanece
  **56,7 %** de implementación, **6,7 %** de certificación y **NO-GO**.

Actualización 2026-08-11, carga de reglas antivirus en observabilidad:

- Prometheus de staging retenía una versión anterior del archivo montado. Tras
  recrear exclusivamente ese contenedor con su entorno de monitoreo existente,
  `promtool` validó 17 reglas y quedaron montadas las cuatro alertas de
  antivirus. No se modificaron backend, datos, ClamAV ni producción.
- Falta probar entrega/deduplicación/resolución en un receptor externo aprobado.
  P109-010 permanece parcial y la fórmula sigue **56,7 %**, **6,7 %** y
  **NO-GO**.

Actualización 2026-08-11, recepción interna de alerta antivirus:

- La caída controlada de ClamAV generó rechazo fail-closed desde la interfaz;
  Prometheus evaluó `PCSAntivirusSoportesNoDisponible` como firing y
  Alertmanager interno la recibió. ClamAV se recuperó sano.
- La resolución depende de la ventana de diez minutos y el receptor externo
  permanece sin configurar, por lo que P109-010 continúa **parcial** y el
  veredicto **NO-GO** no cambia.

Actualización 2026-08-11, preparaci\u00f3n del drill de restauraci\u00f3n por digest:

- El validador P109 reconoce la topolog\u00eda actual de staging, cuyo contenedor
  API se denomina `pcs-staging-backend`; conserva el nombre anterior como
  compatibilidad y exige en ambos casos una referencia `@sha256` v\u00e1lida.
- El candidato inmutable se construye, analiza y publica de forma separada
  antes de ejecutar el restore. Mientras no existan los cuatro digests
  publicados y no termine el drill, P109-008 no se declara aprobado y el
  veredicto sigue **NO-GO**.
