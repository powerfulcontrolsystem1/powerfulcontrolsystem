# Nomina Colombia avanzada

Estado: Vigente. Responsable: Ingeniería del módulo. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Se conserva el alcance de nómina electrónica desde fuente mensual pagada y numeración específica. Un CUNE calculado o documento local no acredita aceptación oficial.
- Ajustes, habilitación automatizada y distribución dedicada permanecen fuera del alcance declarado; no crear fuentes laborales ficticias para probar.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

Fecha: 2026-08-26

## Alcance

El modulo conserva liquidacion, asistencia, novedades, provisiones, pagos,
PILA y desprendibles en `nomina_sueldos`. Para Colombia agrega un adaptador
dedicado de documento soporte de pago de nomina electronica
`NominaIndividual`, sin reutilizar XML, numeracion ni representacion grafica de
factura de venta.

La unidad fiscal es mensual: PCS consolida todas las liquidaciones activas y
pagadas de un trabajador dentro de un mismo mes calendario. Una frecuencia
interna semanal, quincenal o mensual no cambia esta regla. La fuente solo puede
construirse para un mes ya cerrado y queda aislada por `empresa_id`.

## Superficies

- Backend de nomina: `/api/empresa/nomina`.
- Emision fiscal dedicada:
  `POST /api/empresa/facturacion_electronica?action=emitir_nomina_electronica`.
- Pantalla: `web/administrar_empresa/nomina_sueldos.html`, seccion
  `Nomina electronica DIAN`.
- Configuracion DIAN principal:
  `web/administrar_empresa/facturacion_electronica.html`.
- Configuracion independiente de familia: registro `nomina_electronica` en
  `empresa_dian_documentos_configuracion`.
- Persistencia principal: `empresa_nomina_empleados`,
  `empresa_nomina_liquidaciones`, `empresa_nomina_pagos`,
  `empresa_contabilidad_nomina_electronica` y
  `empresa_nomina_dian_perfiles`.
- Espejo fiscal comun: `empresa_facturacion_documentos`,
  `facturacion_electronica_reintentos` y
  `empresa_facturacion_artefactos`.

## Fuente fiscal mensual

PCS construye la fuente en servidor; el navegador no puede suministrar
devengados, deducciones, partes, CUNE ni numero legal. Para cada trabajador y
mes se exige:

1. Liquidaciones activas pertenecientes a la empresa y al mismo trabajador.
2. Periodos que no se solapen, no crucen el limite del mes y cubran exactamente
   la fuente seleccionada.
3. Un pago activo real por liquidacion; se conservan todas las fechas de pago
   del mes y se rechazan pagos ambiguos o duplicados.
4. Totales de devengados, deducciones y comprobante conciliados con precision
   monetaria exacta.
5. Perfil fiscal DIAN explicito del trabajador: tipo y numero de documento,
   apellidos/nombres, tipo de trabajador, subtipo, contrato, salario integral,
   lugar de trabajo y atributos requeridos.
6. Datos del empleador y, cuando el software DIAN es compartido, identidad
   completa del proveedor del software, incluido un DV de exactamente un
   digito para cada parte. PIN, certificado y otros secretos no forman parte
   de la instantanea publica.
7. Conceptos de tiempo con intervalos reales. Si una liquidacion solo contiene
   horas o recargos agregados sin fecha/hora inicial y final, la emision se
   bloquea; PCS no inventa intervalos.

La reserva atomica genera como maximo un documento por
`empresa_id + empleado_nomina_id + periodo_reporte`, conserva el primer numero
legal asignado y sella `fuente_fiscal_json` y `configuracion_dian_json`. Un
reintento reutiliza el mismo XML firmado y no regenera fecha, CUNE ni
consecutivo.

## Configuracion

- `empresa_nomina_configuracion.periodo_nomina_dian` usa el codigo DIAN de
  periodicidad y admite valores entre 0 y 6.
- `empresa_dian_documentos_configuracion` mantiene para
  `nomina_electronica` ambiente, estado, modo de operacion, `TestSetId`, prefijo
  y consecutivo interno. No reutiliza la resolucion/rango de factura de venta.
- La configuracion DIAN principal conserva Software ID/PIN, certificado,
  endpoint y la identidad del proveedor del software compartido.
- Habilitacion y produccion deben usar sus endpoints oficiales correspondientes;
  una URL cruzada bloquea el preflight.
- La bandeja solo marca `puede_emitir=true` si el mes ya cerro, la familia esta
  activa en produccion y la integracion Colombia tambien esta activa contra
  DIAN real. Habilitacion y sandbox siguen disponibles para revision, pero no
  muestran una emision productiva como lista.

## Acciones API

- `GET action=perfil_dian&empleado_nomina_id={id}`: consulta el perfil fiscal
  del trabajador.
- `POST action=perfil_dian`: crea o actualiza el perfil fiscal dentro del tenant.
- `GET action=documentos_electronicos_colombia`: resume candidatos mensuales y
  bloqueos sin reservar ni transmitir.
- `GET action=nomina_electronica_preflight&liquidacion_id={id}`: reconstruye la
  fuente mensual y valida configuracion, origen, totales y esquema sin efectos
  fiscales.
- `POST action=preparar_nomina_electronica`: conserva compatibilidad como
  preparacion/revision; no emite.
- `POST /api/empresa/facturacion_electronica?action=emitir_nomina_electronica`:
  exige `confirmar_emision=true` y la frase exacta
  `EMITIR NOMINA ELECTRONICA DIAN`.

## Flujo de produccion

1. Configurar parametros legales, empresa, software DIAN y familia documental.
2. Completar el perfil fiscal de cada trabajador.
3. Calcular y conciliar liquidaciones; registrar sus pagos reales.
4. Esperar a que cierre el mes calendario reportado.
5. Ejecutar el preflight desde la fila mensual. Ningun preflight reserva numero
   ni transmite XML.
6. En produccion y solo con todos los bloqueos resueltos, abrir el dialogo de
   emision, revisar trabajador, periodo, pagos y total, y escribir la frase
   exacta.
7. El servidor toma un bloqueo asesor por documento, reserva el consecutivo de
   la familia, sella fuente/configuracion, genera `NominaIndividual`, calcula
   CUNE SHA-384, firma XAdES, valida el XML y usa `SendNominaSync`.
8. El XML firmado se guarda antes del transporte. El acuse se conserva como
   artefacto privado y se refleja en documento, cola y espejo de nomina.
9. Una respuesta pendiente usa la cola durable. El worker puede continuar el
   reintento ya autorizado; el procesamiento manual generico omite nomina.
10. Un reenvio individual manual exige permiso de aprobacion en Facturacion y
    Nomina y la frase exacta `REENVIAR NOMINA ELECTRONICA DIAN`.

## Permisos y privacidad

- Emision y reenvio exigen simultaneamente licencia/pagina/permiso de
  Facturacion y de Nomina; `empresa_id` lo fija el contexto autorizado.
- Listados generales, configuracion de familias, cola y artefactos omiten
  nomina si el usuario solo tiene acceso a Facturacion. Una consulta o descarga
  explicita de nomina devuelve prohibido.
- El procesamiento manual general excluye la familia de nomina en SQL antes de
  paginar y no devuelve codigos ni conteos de documentos omitidos. El worker
  solo retoma documentos que ya quedaron autorizados, numerados y sellados.
- El dashboard de la suite contable tambien evalua `nomina_sueldos:R`: sin ese
  alcance no consulta ni devuelve conteos, nombres, documentos o resumenes de
  nomina, y la interfaz oculta su KPI y pestaña.
- El identificador relacionado de un documento de nomina representa al
  trabajador y nunca se une contra `clientes`; evita mezclar identidades cuando
  ambas tablas tienen por casualidad el mismo ID numerico.
- El correo generico de facturas no admite nomina porque produciria una
  representacion grafica incorrecta. La distribucion al trabajador debe usar
  un flujo dedicado antes de declararse operativa.

## Limites vigentes

- `NominaIndividualDeAjuste` permanece bloqueada hasta tener fuente de ajuste,
  tipo de nota, referencia y ciclo DIAN propios.
- En habilitacion PCS ejecuta preflight local, pero aun no automatiza el envio
  de nomina con `SendTestSetAsync` y `TestSetId`; por eso no se debe activar
  produccion sin completar la habilitacion oficial por la via controlada.
- No se ejecuto una emision real de nomina contra DIAN en este candidato. La
  prueba real exige una liquidacion/pago/configuracion genuinos y autorizacion
  operativa posterior al despliegue.
- La representacion grafica y entrega dedicada de nomina siguen cerradas; el
  sistema falla antes de usar el PDF/correo de factura.

## Validacion

- Generacion, CUNE, orden XML, fechas de pago, bloqueos de intervalos, firma
  XAdES, preflight y contrato SOAP tienen regresiones Go.
- El XML firmado de prueba paso el XSD oficial
  `NominaIndividualElectronicaXSDV1.0.6.xsd` de la caja DIAN.
- `SendNominaSync` serializa exclusivamente `contentFile`, segun el contrato
  oficial; no incorpora `TestSetId` en produccion.
- La migracion empresarial es `20260826-005-dian-nomina-electronica-v2` y la
  prueba PostgreSQL opcional usa `PCS_TEST_POSTGRES_DSN`.
- Esta evidencia es local. PR, CI, aprobacion, merge, despliegue, migracion VPS
  y QA autenticado se registran como compuertas separadas.

## Fuentes y aceptación de la revisión

[dian_nomina_electronica.go](../backend/handlers/dian_nomina_electronica.go), [dian_nomina_electronica.go](../backend/db/dian_nomina_electronica.go), [dian_nomina_electronica_migration.go](../backend/db/dian_nomina_electronica_migration.go), [nomina_sueldos.go](../backend/handlers/nomina_sueldos.go), [empresa_permisos.go](../backend/handlers/empresa_permisos.go).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).
