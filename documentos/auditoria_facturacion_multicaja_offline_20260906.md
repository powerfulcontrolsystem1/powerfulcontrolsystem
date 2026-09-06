# Auditoría de facturación: multicaja, numeración y offline

Estado: Evidencia de candidato, no certificación de producción.
Responsable: Coordinación técnica. Revisión: 2026-09-06.
Base: `f4a732c00fb78f6e941f692a8c4e42da2a7c0c7c`.
Rama: `codex/invoicing-audit-20260906`.
Workspace aislado: `D:\pcs-invoicing-audit-20260906`.

## Resultado

NO-GO para afirmar que todas las familias, países y contingencias están listas.
Se corrigieron fallos reproducidos de numeración, recuperación pospago y cola
offline, y se implementó la base operativa de contingencia Colombia. La
transcripción electrónica tipo 03 aún no está implementada y permanece
bloqueada. En producción solo se concilió el contador avanzado de empresa 12 de
9 a 12, con precondiciones exactas y auditoría; no se emitieron, reenviaron,
pagaron ni anularon documentos fiscales durante esta revisión. Los cambios
concurrentes del checkout principal se preservaron.

## Correcciones y pruebas

| Hallazgo | Corrección | Evidencia |
| --- | --- | --- |
| Guardar un formulario viejo reducía el próximo consecutivo | UPSERT conserva el máximo para la misma empresa/país/ambiente/prefijo, incluso al renovar resolución | PostgreSQL reprodujo reutilización antes del cambio; después pasa |
| Caída entre reserva y persistencia del documento consumía otro número al reintentar | Reserva persistente empresa+documento, insertada en la misma transacción que avanza el contador | Antes devolvía QA1 y luego QA2; después conserva número y fecha |
| Riesgo de reutilizar numeración histórica | Se comprueban documentos y reintentos; colisión revierte toda la reserva sin saltar un folio silenciosamente | Caso QA-1 histórico contra QA1, rollback confirmado |
| Configuración fiscal mezclada durante cambios | Se contrasta país/ambiente/prefijo/resolución con la fila bloqueada; replay con moneda/importe distinto se rechaza | Pruebas de conflicto, aislamiento y configuración |
| Cuota o JSON inválido de localStorage se ocultaban | Fallo visible, sin sobrescribir cola ilegible ni imprimir como guardada | Pruebas JS sobre funciones reales extraídas de HTML |
| Otra pestaña podía perder una venta capturada durante HTTP | Web Locks por empresa y mezcla con cola actual al finalizar; preservación de filas ajenas en claves heredadas | 32 inserciones concurrentes y venta B capturada mientras se sincroniza A |
| Venta offline sin identidad/key podía reclamar una operación | Validación completa de empresa, operador autenticado, caja y sync_key antes del claim; máximo 100 ventas y 2 MiB | Regresiones Go y 32 claims PostgreSQL con un solo ganador |
| Recibo offline podía interpretarse como factura válida | Se denomina comprobante provisional y aclara ausencia de validación DIAN/contingencia autorizada | Regresión de orden guardar-antes-de-imprimir |
| Centro DIAN convertía fechas civiles ISO a UTC y mostraba un día menos | Se conserva la fecha civil en visualización y campos datetime-local | 2026-06-17 sigue siendo 17/6/2026 |
| Checklist completo ocultaba disparidad de contadores | Aviso de concordancia en configuración; no corrige ni reduce números automáticamente | Prueba visual lógica del aviso 12 versus 9 |
| Dos cajas podían decidir la misma posición de frecuencia automática | La decisión factura/comprobante se congela dentro de la transacción de pago y el contador se bloquea por empresa | 32 decisiones concurrentes: 8 facturas, 24 comprobantes, contador final 0 para frecuencia cada 3 |
| Una caída después de cobrar podía dejar el documento ausente | `commerce.sale-paid` conserva la decisión; el worker recupera contabilidad y luego el documento exacto, idempotentemente | Contratos Go, ruta real de worker y pruebas enfocadas verdes |
| La conectividad fallida se confundía con contingencia fiscal | Solo se usa estado contingencia cuando existe incidente DIAN explícito, activo y del mismo tenant | Guard backend y endpoint tras permisos de facturación |
| No existía ledger de contingencia | Historial de autorizaciones, incidentes, documentos, plazo operativo y cierre sin pendientes separados por empresa | Lifecycle y aislamiento sobre PostgreSQL 16; talonario avanza solo su tenant y una renovación no reemplaza la serie durante un incidente activo |

La reserva no reemplaza la fuente fiscal inmutable ni el XML firmado. No
modifica históricos ni acredita aceptación DIAN. Un borrador sin número puede
reservar; un documento histórico numerado requiere conciliación, no renumeración.
Por prudencia la colisión histórica bloquea también si su ambiente es ambiguo.

## Evidencia ejecutada

- `go test ./... -count=1` desde backend: PASS en todos los paquetes.
- `go vet ./...` desde backend: PASS.
- `node --test tools/qa_offline_queue.test.cjs tools/qa_fiscal_ui.test.cjs`:
  13 pruebas PASS, ninguna omitida. Incluye sintaxis inline de los tres HTML y
  contrato visual fail-closed de contingencia Colombia.
- Binario Linux de pruebas DB sobre PostgreSQL 16 aislado: doce pruebas de
  integración PASS. El contenedor usó tmpfs, límites de recursos y
  exclusivamente datos sintéticos; no conectó DIAN ni las bases productivas.
  Al terminar se eliminó exclusivamente ese contenedor efímero y su binario
  temporal; backend, frontend y worker productivos permanecieron healthy.
- Concurrencia: 32 reservas distintas sin duplicados; 32 replay del mismo
  documento conservan QA1 aun agotado el rango; dos empresas pueden usar QA1
  independientemente; 32 claims offline producen un único ganador; 32 cajas
  serializan la frecuencia automática sin perder ni duplicar posiciones.
- Contingencia: incidente DIAN aislado, cierre bloqueado con pendientes, venta
  de papel sintética respaldada por carrito pagado y fuente inmutable, avance
  de CTG1 a CTG2 solo para su empresa y cero fuga al segundo tenant.
- Inventario Ensure regenerado y comprobado: 141 funciones, 117 pasos.
- Preflight profesional `-Full -RequireMigrationAudit`: estado OK, incluido
  Go completo y auditoría de migraciones. Reporte
  `documentos/reportes_profesionales/preflight_20260906_130658.md`. La compuerta
  Docker Compose del preflight quedó OK; no equivale a un despliegue del candidato.

Las pruebas PostgreSQL optativas requieren `PCS_TEST_POSTGRES_DSN`: un `go test`
sin esa variable las omite. Aquí sí se ejecutaron en PostgreSQL. No representan
32 sesiones web de pago reales ni una prueba de carga contra DIAN.

## QA autenticado en PCS, empresa 12

Se inspeccionaron renderizados y estados accesibles del navegador real.

- Configuración DIAN/RUT: carga disponible y secretos enmascarados. No se subió
  RUT ni se envió a IA. La pantalla muestra rango 1PCS, 1–100000, vigencia
  2026-06-17 a 2028-06-17; DIAN tiene consecutivo 12 y avanzada próximo 9.
  La revisión posterior de base productiva confirmó máximo canónico 1PCS8, sin
  números 9–11, y trazabilidad de pruebas DIAN previas. Se concilió únicamente
  el contador avanzado de 9 a 12, igualando el próximo DIAN sin retrocederlo.
- Centro DIAN: producción local activa, progreso 85 %, historial con acuses
  aceptados y pendientes. Emisión manual, NC/ND libres, POS y RADIAN bloqueados.
  No se ejecutó set, procesamiento de cola ni reconsulta de pendientes.
- Bandeja de agosto: 1PCS8 emitida por 119 COP y controles de artefactos
  visibles. NC12000000113 y facturas ya anuladas con Anular deshabilitado.
- Diálogo 1PCS8: deshabilitado al abrir vacío y con solo ANULAR; habilitado
  con ANULAR y motivo válido; bloqueado con confirmación en blanco y motivo.
  Se canceló sin confirmar. La herramienta no vació el campo con fill vacío;
  se comprobó el caso en blanco con un espacio, sin atribuir ese límite al UI.
- Nómina: configuración separada inexistente, estado configurando/habilitación;
  preflight ejecutado sin reservar, `pendiente_liquidaciones`, 0 documentos y
  neto 0; emisión por empleado deshabilitada.
- Documento soporte: pestaña y configuración independiente visibles, estado
  configurando/habilitación, 0 documentos y mensaje sin compra real para emitir.
  No se creó compra, proveedor, numeración ni borrador ficticio.

El candidato local no está desplegado: los nuevos avisos y correcciones se
validaron con pruebas de código, no se presentan como ya visibles en producción.
No se cortó la conexión del navegador productivo para provocar pagos/reenvíos
automáticos. La simulación de fallos offline se hizo en un entorno aislado.

## Contraste normativo Colombia

La documentación oficial mantiene el anexo FE 1.9. Falta validar cada variante
real de XML, firma y reglas de negocio, no solo un fixture o un acuse antiguo.
[DIAN, documentación técnica](https://micrositios.dian.gov.co/sistema-de-facturacion-electronica/documentacion-tecnica/).

La Resolución 227 distingue caída del facturador y caída DIAN. La primera
requiere factura de papel/talonario y posterior transcripción tipo 3; la segunda
tiene condiciones propias para transmisión posterior. Los plazos de 48 horas
no se cuentan igual. Un ticket local y una cola de ventas no implementan ese
contrato. La nota crédito de anulación no permite reutilizar el número original.
La entrega electrónica requiere validar el contenedor y el documento de
validación correspondiente, no asumir que PDF+XML bastan en todos los casos.
[DIAN, Resolución 227, arts. 1.5.1.5.5–1.5.1.5.7](https://normograma.dian.gov.co/dian/compilacion/docs/resolucion_dian_0227_2025.htm).

La Resolución 202 limita los datos exigibles al comprador: identidad, nombre y
correo cuando corresponda la entrega electrónica. No debe imponerse dirección
o RUT como requisito comercial general. El preflight actual de fuente fiscal
exige dirección, municipio, códigos DANE y responsabilidad al cliente identificado
(`facturacion_fuente_fiscal.go`): brecha pendiente de ajustar junto con el UBL,
sin inventar datos ni simplemente eliminar validaciones técnicas.
[DIAN, Resolución 202, arts. 3–4](https://normograma.dian.gov.co/dian/compilacion/docs/resolucion_dian_0202_2025.htm).

La numeración autorizada y su prefijo deben corresponder a la operación;
varias cajas no pueden decidir el siguiente número mediante estado del navegador.
La arquitectura centralizada permite una secuencia compartida serializada;
rangos/prefijos independientes por establecimiento necesitan asignación explícita.
[DIAN, preguntas de numeración](https://micrositios.dian.gov.co/sistema-de-facturacion-electronica/numeracion-de-facturacion-preguntas-frecuentes/).

## Pendientes concretos antes de cierre integral

1. Publicar/revisar el candidato, CI/aprobación, migración y despliegue normal;
   repetir pruebas autenticadas de recuperación y contingencia ya desplegadas.
2. Completar el contrato tipo 03: CUDE, `InvoiceTypeCode=03`, referencia exacta
   al papel y fecha, validación XSD, firma, transporte, acuse y reintento dentro
   del plazo aplicable. No activar offline como sustituto fiscal mientras falte.
3. Carga end-to-end de pagos y SOAP: las conexiones retenidas por advisory locks
   y los contadores de frecuencia necesitan medir saturación/equidad; las 32
   reservas no prueban esos caminos ni rendimiento sostenido.
4. Offline depende de sesión, datos/caja precargados, navegador con Web Locks
   y almacenamiento disponible. No cubre arranque en frío sin red, cierre/borrado
   del perfil, dispositivo compartido, revocación offline ni resistencia a XSS.
   localStorage no es un almacén cifrado; el filtro de operador no sustituye
   separación de perfiles del navegador y controles de acceso en servidor.
5. Pruebas negativas con roles mínimos y segundo tenant autorizado, restauración
   de artefactos/acuse, retención, rotación de claves, observabilidad de cola y
   comprobación del paquete de entrega al comprador. La auditoría estática no
   equivale a pentest ni certificación jurídica.
6. Completar requisitos reales de nómina/soporte y sus acuses; NC parcial, ND,
   ajustes de nómina/soporte, POS y RADIAN siguen fuera del alcance operativo
   integral. Un país nuevo es un perfil de configuración, no un adaptador fiscal.

## Despliegue y recuperación

Las migraciones nuevas `20260906-001-facturacion-reservas-v1` y
`20260906-002-facturacion-contingencias-v1` solo las ejecuta pcs-migrate. Crean
reservas/índices y el dominio de contingencia; no cambian checksums previos, no
renumeran y no transmiten. Evaluar tamaño/locks y respaldo antes de migrar.
No volver a un binario que ignore reservas después de asignar números nuevos;
ante incidente bloquear emisión y conservar ledger, fuente y XML para recuperar.
No borrar reservas para desbloquear un consecutivo.

El catálogo documental nuevo no existe en este SHA base. No se copió ni se
ejecutó la versión aún modificada en el checkout concurrente. Se registran todos
los archivos en el inventario textual; regenerar catálogo al integrar esa
gobernanza sin arrastrar cambios ajenos.
