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

- P109-000 está aprobada en staging para el SHA fusionado `89d6e042...`: PR 119,
  CI y release `30700866694` verdes,
  cuatro digests/SBOM, Trivy sin vulnerabilidades, base vacía, upgrade,
  idempotencia, rechazo de checksum, rollback de aplicación, salud y producción
  intacta. P109-001, P109-002, P109-003, P109-004,
  P109-005, P109-008, P109-009, P109-010, P109-011 y P109-012 tienen evidencia
  parcial. Las demás fases continúan pendientes o bloqueadas.
- P109-001 añadió rechazos runtime para empresa/topic ausentes, duplicados, ID
  inexistente y publicado sin modificar el evento histórico excluido. Falta una
  segunda empresa controlada para la matriz A/B.
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
  inmediata de las cachés de permisos. Falta repetirlo sobre el nuevo digest.
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
  `GetStatusZip StatusCode=00` oficial.
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

La evidencia histórica P108 sirve como línea base y evita repetir pruebas
independientes del artefacto, pero no aumenta automáticamente estas cifras.

## 9. Siguiente orden para Terra alto

P109-000 ya está aprobada para `89d6e042...`; Terra no debe reconstruirla ni
repetirla sin un cambio de código. Debe continuar así:

1. confirmar que staging conserva los cuatro digests registrados y producción
   conserva sus imágenes anteriores;
2. cerrar P109-001 únicamente cuando exista una segunda identidad/empresa A/B
   autorizada, sin usar la sesión global como prueba de aislamiento;
3. completar en P109-002 confirmación/cancelación, doble clic, degradación del
   proveedor y evals, sin convertir automáticamente el borrador en pago;
4. ejecutar UAT contable/fiscal de P109-003 con contador autorizado;
5. completar acciones mutantes de P109-004 por rol y flujo oficial;
6. continuar cuatro cajas, DIAN, receptor externo, restore y piloto solo cuando
   sus credenciales, identidades, hardware o ventanas estén disponibles;
7. registrar evidencia y recalcular las dos cifras sin dar crédito completo a
   una fase parcial.

No debe saltar a producción ni llamar 100 % a un bloque con pendientes.
