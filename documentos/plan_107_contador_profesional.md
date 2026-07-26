# Plan 107 - certificacion profesional del contador y reportes IA

Estado: **P0 local parcial implementado; NO-GO contable y de datos reales**.

Fecha de corte: 2026-07-25.

Empresa prevista para la prueba real controlada:
`Powerful Control System` (`empresa_id=12`).

Modelo solicitado para ejecutar y revisar este plan:
**GPT-5.6 Sol con razonamiento alto**.

Este documento complementa el Plan 106. No sustituye sus compuertas de cartera
de proveedores, aislamiento, varias cajas, DIAN, impresiones, staging,
backup/restore ni liberacion. Si ambos planes difieren, se aplica la condicion
mas estricta.

## 1. Parada obligatoria y autorizacion

La ejecucion P0 local no autoriza datos reales, despliegues ni efectos externos.
Los componentes restantes del plan conservan las siguientes prohibiciones:

- `rs`, despliegue, migraciones ni cambios en VPS;
- ventas, compras, movimientos de caja, pagos, abonos o cierres reales;
- facturas o documentos electronicos DIAN;
- transferencias, movimientos bancarios o llamadas a proveedores;
- correos, WhatsApp, declaraciones o presentaciones ante autoridades;
- borrado o alteracion de datos de otras empresas.

La ejecucion comienza solo cuando el usuario cambie al modelo deseado y la
autorice expresamente. Las pruebas se hacen primero en staging con fixtures
reversibles. El subconjunto real en empresa 12 se ejecuta despues de conciliar
staging y con autorizacion puntual para cualquier efecto fiscal o externo.

Las credenciales no se copian a comandos, documentos, capturas, logs ni
artefactos. Se usa sesion autenticada o entrada segura.

## 1.1 Avance local implementado el 2026-07-24

- El rol `contador` puede leer `reportes`, usar el chat de reportes IA y el
  Centro IA empresarial, sin adquirir permisos de crear, aprobar, pagar,
  cerrar ni modificar datos.
- `admin_empresa` conserva sus consultas autorizadas dentro de su empresa; el
  alcance no se amplía por el prompt.
- La IA puede proponer un reporte nuevo mediante `ReportSpec` validado en el
  servidor. Solo permite fuentes, campos, filtros, métricas y fórmulas de un
  catálogo cerrado; no ejecuta SQL del modelo.
- La auditoría de cada reporte nuevo conserva la fuente y una huella SHA-256
  determinista de su `ReportSpec`, sin duplicar datos de negocio en el registro.
- El Centro de reportes muestra la vista previa y permite exportar la misma
  especificación en Excel, PDF o CSV. La exportación no muta datos.
- Antes de exportar, el Centro de reportes muestra el `ReportSpec` ejecutado
  para que el contador pueda revisar fuente, filtros, métricas y fórmulas.
- El contador puede guardar una nueva versión de la plantilla del reporte IA.
  La plantilla reutiliza el repositorio empresarial versionado y el servidor
  valida de nuevo que el `ReportSpec` no cambie de fuente semántica.
- Estado de resultados y balance general se marcan explícitamente como
  `preliminar_no_oficial` cuando el período no tiene asientos contables
  canónicos; los resúmenes con asientos son aptos para revisión contable, pero
  `apto_para_presentacion_oficial` permanece en `false` hasta completar notas,
  comparativos, políticas y firmas. Esto no reemplaza los estados financieros
  completos pendientes.
- El catálogo incluye auxiliares por rubro y cuenta de resultado, situación
  financiera y patrimonio, calculados exclusivamente desde asientos canónicos.
  Muestran de forma explícita que siguen requiriendo comparativos, notas y
  firmas para constituir estados financieros completos.
- P107-004 dispone de un manifiesto de fixtures `P107-QA` determinista, con
  escenarios, claves de idempotencia, huella SHA-256 y reverso. Aún no existe
  autorización ni ejecutor de datos para aplicarlo fuera de staging.
- Pruebas locales focalizadas de permisos, plantilla IA, estados financieros y
  rubros PUC (ok). La comprobación visual de staging no pudo iniciar porque
  `127.0.0.1:8082` rechazó la conexión; siguen pendientes pruebas visuales,
  roles reales, IA real, aislamiento A/B y recalculo en staging.
- Pruebas locales de CxP (ok): esquema y migración, proveedor canónico por
  empresa, idempotencia sin exponer la llave, abono atómico y validaciones del
  handler. La conciliación contra submayor, mayor y banco sigue pendiente en
  staging con fixtures reversibles.
- Regresión del catálogo contable (ok): los estados, libros, balance de
  prueba, CxC/CxP, edades, bancos, impuestos, exógena y auxiliares financieros
  no pueden desaparecer ni perder exportación JSON/CSV/XLS/PDF sin fallar la
  suite P107.
- Seguridad de `ReportSpec` (ok): se rechazan fuentes o campos inyectados,
  alias no seguros, fórmulas con referencias fuera de las métricas, operadores
  inválidos y límites excesivos antes de consultar cualquier dataset.
- Batería contable local seleccionada (ok): cartera por edades, reglas PUC y
  tercero/impuesto, asiento de venta balanceado, hitos precontables, CxP
  canónica, permisos de contador, catálogo/reportes fiscales, ReportSpec y
  auxiliares financieros. Es evidencia de lógica local, no sustituye la
  conciliación visible ni el recalculo independiente en staging.
- Regresión amplia local (ok): `go test -vet=off ./handlers -count=1` y
  `go test -vet=off ./db -count=1` pasaron el 2026-07-25. La cobertura no
  conecta una base PostgreSQL ni ejecuta procesos de caja, por lo que el plan
  conserva estado NO-GO hasta completar staging y UAT.
- Regresión transversal local (ok): `go test -vet=off ./... -count=1` pasó
  para todos los paquetes Go del backend. Mantiene la confianza de integración
  de código, pero no reemplaza la evidencia operativa de P107-QA.
- Preflight P107 implementado: `go run ./tools/plan107_preflight` impide
  iniciar fixtures fuera de staging. La comprobación remota de solo lectura
  contra `https://staging.powerfulcontrolsystem.com/health` respondió `200 OK`
  y dejó el entorno apto técnicamente para la ventana controlada; no se
  cargaron datos `P107-QA`.

## 2. Objetivo

Certificar PCS como herramienta util y confiable para un contador colombiano:

1. demostrar que cada operacion produce inventario, caja, cartera, impuestos,
   comprobantes y estados financieros coherentes;
2. probar todos los reportes y funciones que necesita un contador;
3. simular un ciclo economico completo con ventas, credito, compras,
   proveedores, bancos, nomina, activos, impuestos y cierre;
4. permitir que la IA genere reportes existentes y reportes nuevos solicitados
   en lenguaje natural sin SQL libre ni acceso entre empresas;
5. detectar faltantes, implementarlos por fases y repetir la certificacion;
6. entregar evidencia que un contador independiente pueda recalcular.

## 3. Referencia profesional y normativa

### 3.1 Funciones observadas en Siigo Colombia

Como referencia de mercado, Siigo expone:

- contabilidad integrada con documentos, impuestos y estados financieros:
  <https://www.siigo.com/siigo-contador/>;
- estado de situacion financiera con comparativos, centro de costo,
  diferencias fiscales, firmas, Excel y PDF:
  <https://siigonube.portaldeclientes.siigo.com/generar-estado-de-situacion-financiera/>;
- balance de prueba general, por tercero y por centro de costo:
  <https://siigonube.portaldeclientes.siigo.com/generar-balance-de-prueba-general/>,
  <https://siigonube.portaldeclientes.siigo.com/generar-balance-de-prueba-por-tercero/>,
  <https://siigonube.portaldeclientes.siigo.com/generar-balance-de-prueba-por-centros-de-costos/>;
- estado de cambios en el patrimonio y catalogo de informes financieros:
  <https://siigonube.portaldeclientes.siigo.com/generar-estado-de-cambios-en-el-patrimonio/>,
  <https://siigonube.portaldeclientes.siigo.com/informes-financieros-y-contables/>;
- cuentas por cobrar/pagar por edades, documento y tercero:
  <https://siigonube.portaldeclientes.siigo.com/generar-informes-de-proveedores/>,
  <https://siigonube.portaldeclientes.siigo.com/generar-reporte-cuentas-por-cobrar-detallada-por-documento/>;
- libros oficiales, auxiliares, comprobantes, cierre y reverso:
  <https://siigonube.portaldeclientes.siigo.com/generar-libro-mayor-y-balance/>,
  <https://siigonube.portaldeclientes.siigo.com/generar-cierre-de-ano/>;
- informacion exogena generada desde contabilidad y validada antes de enviar:
  <https://www.siigo.com/blog/obligaciones-fiscales/informacion-exogena/>.

PCS no debe copiar pantallas ni marcas. Estas funciones sirven como lista de
expectativas del contador y deben implementarse con los patrones propios de PCS.

### 3.2 Requisitos colombianos que gobiernan la aceptacion

- El marco compilado en Colombia exige, segun el grupo aplicable, un conjunto
  completo de estados financieros: situacion financiera, resultado y otro
  resultado integral, cambios en el patrimonio, flujos de efectivo, notas e
  informacion comparativa:
  <https://www.funcionpublica.gov.co/eva/gestornormativo/norma.php?i=76054>.
- La DIAN publica calendario tributario 2026 y vencimientos por NIT:
  <https://www.dian.gov.co/Contribuyentes-Plus/Paginas/Calendario-de-obligaciones.aspx>.
- La informacion exogena se rige por resoluciones y anexos de cada año
  gravable; no basta una tabla generica:
  <https://www.dian.gov.co/impuestos/sociedades/exogenatributaria/normatividad/paginas/default.aspx>.
- El documento soporte en adquisiciones a no obligados requiere contenido,
  numeracion, firma, formato electronico y CUDS:
  <https://micrositios.dian.gov.co/sistema-de-facturacion-electronica/soporte-adquisiciones-no-obligados/>.
- Nomina electronica, factura, notas y eventos RADIAN requieren trazabilidad y
  validacion DIAN:
  <https://micrositios.dian.gov.co/sistema-de-facturacion-electronica/solucion-gratuita-de-nomina-electronica/>,
  <https://micrositios.dian.gov.co/sistema-de-facturacion-electronica/radian/>.

Las reglas, topes, tarifas, formatos y fechas deben ser configurables por
vigencia. Ningun texto del plan se usa como parametro tributario permanente.

## 4. Linea base comprobada en el codigo actual

### 4.1 Capacidades existentes

- `suite_contador.html` coordina 16 accesos: Portal contador, contabilidad
  Colombia, NIIF, contabilidad avanzada, impuestos, declaraciones, DIAN,
  reportes, Renta IA, certificados, cierre fiscal, activos, Centro IA, nomina,
  compras IA y tesoreria.
- `/api/empresa/reportes` publica 43 datasets fijos en JSON, CSV, TXT, XLS
  compatible y PDF.
- El catalogo incluye estado de resultados, balance general, flujo de caja,
  balance de prueba, libro auxiliar, mayor, impuestos/retenciones, exogena base,
  CxC, CxP, edades, conciliacion bancaria, ventas, inventario y compras.
- Existen PUC, terceros, comprobantes de doble partida, periodos, activos,
  nomina, documento soporte, exogena configurable y motor de eventos/asientos.
- El producto de prueba documentado es `menta`, producto `id=103`, SKU `1`,
  precio de referencia COP 100. Al ejecutar se debe reconfirmar ID, SKU, precio,
  impuesto, costo, stock y estado; la documentacion no reemplaza el dato vivo.

### 4.2 Bloqueadores conocidos

1. **IA limitada a catalogo fijo.** `reportes_ia_chat.go` solo pide a la IA
   elegir un dataset y formato existentes. No genera una definicion nueva de
   columnas, filtros, agrupaciones o formulas.
2. **Rol contador inconsistente.** La pagina permite accesos contables al rol
   `contador`, pero la regla base de `permModuleReportes:R` no incluye ese rol.
   Debe existir una prueba backend que demuestre el acceso de `contador` a la IA
   y reportes, sin ampliar escritura.
3. **Estados financieros insuficientes.** Los datasets principales son
   resumenes de una fila. Faltan rubros, comparativos, firmas, notas, analisis
   vertical/horizontal y el conjunto completo exigido por el marco aplicable.
4. **Fallback no aceptable como estado oficial.** Si no hay asientos, PCS puede
   construir resultado y balance aproximados desde ingresos/egresos
   financieros. Un reporte oficial debe fallar cerrado o marcarse claramente
   como preliminar no contable; nunca inventar activos, pasivos o patrimonio.
5. **Impuestos incompletos.** El tablero actual se concentra en impuesto
   generado por ventas POS. Debe reconciliar IVA generado/descontable,
   retenciones practicadas/sufridas, ICA/ReteICA, compras, notas, anticipos,
   saldos a favor, declaraciones y cuentas contables.
6. **Exogena no formal.** `fiscal_informacion_exogena_base` declara ser base
   para revision. No prueba formatos por vigencia, conceptos, topes, XML,
   prevalidador ni correcciones.
7. **CxP doble.** El Plan 106 registra dos fuentes de verdad; ningun reporte de
   proveedores se aprueba hasta reconciliarlas.
8. **NIIF local.** Parte del diagnostico/politicas/notas vive solo en navegador.
   No constituye libro ni evidencia empresarial persistente y versionada.

## 5. Roles y autorizacion de la IA

### 5.1 Matriz requerida

| Rol | Consultar IA | Generar reporte nuevo | Guardar plantilla | Mutar contabilidad |
|---|---:|---:|---:|---:|
| `admin_empresa` | Si, cualquier consulta dentro de su empresa | Si | Si | Solo por endpoint, permiso y confirmacion propios |
| `contador` | Si, datos contables/fiscales autorizados | Si | Si, en espacio contable de la empresa | No por una consulta IA |
| `contabilidad` | Si | Si | Si | Solo con permisos contables especificos |
| `auditor` | Lectura y trazabilidad | Si, solo lectura | Opcional segun politica | No |
| `empresario` | Solo reportes ejecutivos autorizados | Limitado a su catalogo | No por defecto | No |
| otros roles | Solo si tienen `reportes:R` explicito | Segun alcance | No por defecto | No |

“Cualquier consulta” para `admin_empresa` significa cualquier pregunta sobre
datos y metadatos que su rol ya puede leer en la empresa activa. No significa
SQL arbitrario, secretos, datos personales innecesarios, archivos privados sin
permiso, infraestructura reservada ni datos de otra empresa.

### 5.2 Casos de permiso obligatorios

- `contador` abre Centro de reportes e IA, consulta y exporta.
- `contador` recibe 403 al intentar crear asiento, cambiar impuesto, pagar,
  cerrar periodo, emitir DIAN o modificar venta.
- `admin_empresa` puede consultar todos los dominios empresariales permitidos.
- `cajero` y `empresario` no pueden ampliar alcance mediante prompt.
- override de empresa, rol compartido y URL directa no amplian permisos.
- empresa A no puede inferir conteos, nombres, importes, plantillas ni errores
  de empresa B.

## 6. Motor seguro para reportes nuevos con IA

### 6.1 Contrato funcional

La IA debe convertir la solicitud en un `ReportSpec` JSON validado por servidor:

- nombre, descripcion y periodo;
- dominios/fuentes semanticas permitidas;
- dimensiones, metricas, columnas y alias;
- filtros, agrupaciones, orden y limite;
- formulas sobre metricas permitidas;
- moneda, zona horaria y criterios contables;
- formato y plantilla visual;
- nivel de detalle, totales y comparativos;
- advertencias, fuentes y fecha de corte.

La IA no produce ni ejecuta SQL. El backend traduce `ReportSpec` mediante un
catalogo semantico de fuentes, relaciones, campos y agregados autorizados. La
autoridad de `empresa_id`, usuario, rol, licencia, maximo de filas, tiempo y
costo siempre pertenece al servidor.

### 6.2 Flujo visible

1. El usuario pregunta en lenguaje natural.
2. La IA responde o pide aclaracion si periodo, base, tercero o criterio son
   ambiguos.
3. PCS muestra “Como se construira”: fuentes, filtros, columnas, formulas y
   limitaciones.
4. Se genera vista previa sin mutaciones.
5. El usuario puede ajustar, exportar o guardar como plantilla versionada.
6. PCS conserva prompt saneado, `ReportSpec`, hash, actor, rol, empresa,
   fuentes, costo, tiempo, conteo y resultado.
7. Si faltan datos o una relacion segura, la IA dice exactamente que falta y
   propone una especificacion; no inventa cifras ni crea joins libres.

### 6.3 Biblioteca minima de preguntas

Probar, como minimo:

- “Ventas de menta por dia, caja, medio de pago e impuesto.”
- “Margen bruto de menta comparado con el mes anterior.”
- “Ventas a credito, abonos y saldo por cliente.”
- “Cartera vencida por edades y probabilidad de recaudo.”
- “Facturas de proveedor pendientes, anticipos y plan semanal de pagos.”
- “Compras por proveedor con IVA descontable y retenciones.”
- “IVA generado menos IVA descontable y conciliacion contra PUC.”
- “Retenciones practicadas por tercero para generar certificados.”
- “Balance de prueba por tercero y centro de costo.”
- “Auxiliar de la cuenta indicada con documento y saldo acumulado.”
- “Estado de resultados comparativo con analisis vertical y horizontal.”
- “Estado de situacion financiera y variacion contra año anterior.”
- “Flujo de efectivo por operacion, inversion y financiacion.”
- “Cambios en patrimonio y explicacion de la variacion.”
- “Cruce de ventas, caja, bancos, DIAN y asientos con diferencias.”
- “Exogena por formato, concepto y tercero con errores de identificacion.”
- “Inventario de menta: inicial, compras, ventas, devoluciones y final.”
- “Movimientos inusuales, asientos descuadrados y documentos sin contabilizar.”
- “Proyeccion de caja de 13 semanas con supuestos visibles.”
- una solicitud completamente nueva que combine dos dominios permitidos;
- una solicitud imposible, para comprobar respuesta “dato no disponible”;
- solicitud ambigua, para comprobar pregunta aclaratoria;
- prompt injection, SQL, secretos y cambio de `empresa_id`, todos bloqueados.

Aceptacion: el contador recalcula una muestra y obtiene las mismas cifras; cero
alucinaciones, cero SQL del modelo y cero acceso cruzado.

## 7. Escenario contable integral de datos

Todos los datos usan prefijo `P107-QA`, usuario, timestamp y referencias
reversibles. Antes de crear se captura saldo/stock inicial. La limpieza usa
reversos y anulaciones auditables, no borrado silencioso.

### 7.1 Saldos de apertura

- caja, banco, capital, inventario y cuentas por cobrar/pagar iniciales;
- comprobante balanceado y trazable;
- importacion con vista previa, validacion y rechazo de descuadre;
- saldo inicial visible en balance, auxiliares y estados.

### 7.2 Ventas reales minimas con `menta`

Reconfirmar que `menta` vale COP 100 y usar el valor vivo si cambio:

1. venta de una menta en efectivo;
2. venta de una menta por transferencia/sandbox;
3. venta de una menta a credito a 30 dias;
4. venta de dos mentas con pago mixto;
5. abono parcial COP 40 al credito de COP 100 y saldo COP 60;
6. pago final, estado de cuenta y paz y salvo;
7. devolucion/nota sobre una venta autorizada;
8. doble clic, retry y repeticion con la misma idempotencia;
9. venta anulada antes de pago y rechazo de mutacion posterior;
10. venta en periodo cerrado, que debe ser rechazada.

Cada caso concilia:

`venta = recibo/documento = pago = caja/banco/CxC = salida de inventario =
costo de venta = impuestos = evento = asiento = reportes`.

La prueba fiscal DIAN de una menta se mantiene separada. No se emite hasta
resolver el bloqueador de firma del Plan 106 y recibir autorizacion puntual.

### 7.3 Impuestos de venta

Una menta puede no cubrir todos los impuestos. Crear en staging productos QA
reversibles para:

- IVA 19%, IVA 5%, exento, excluido y sin impuesto;
- descuento antes de impuesto y redondeo;
- devolucion/nota credito y correccion del impuesto;
- impuesto al consumo cuando aplique al tipo de empresa;
- operacion por debajo y por encima de bases de retencion configuradas;
- tarifa vigente vs tarifa historica.

Los importes se calculan desde parametros vigentes; no se hardcodean topes
legales en pruebas permanentes.

### 7.4 Clientes, credito y cartera

- cliente contado y cliente credito con cupo;
- intento sobre cupo, sin cupo y cliente de otra empresa;
- cuotas, mora, abono parcial, total, anticipado y en paralelo;
- refinanciacion, reverso, nota, saldo a favor y castigo autorizado;
- edades corriente, 1-30, 31-60, 61-90 y mayor de 90;
- estado de cuenta y paz y salvo;
- CxC y contabilidad reconciliadas por documento, no por nombre.

### 7.5 Proveedores, compras y CxP

- proveedor obligado y no obligado a facturar;
- compra contado y compra credito con una y varias cuotas;
- recepcion parcial/total e impacto en inventario;
- factura duplicada, proveedor de otra empresa e idempotencia;
- IVA descontable, retefuente, ReteIVA, ReteICA y descuento;
- anticipo, abono parcial, pago total, saldo a favor y reverso;
- devolucion, nota credito/debito y disputa;
- estado de cuenta, edades y propuesta de pago;
- documento soporte preparado en sandbox con CUDS/estado;
- una sola fuente CxP, saldo/movimiento/asiento atomicos.

### 7.6 Bancos, tesoreria y caja

- apertura, ingresos/egresos, venta, abono y cierre por caja;
- extracto con match exacto, tolerancia, pendiente, duplicado y desviacion;
- conciliacion de transferencia de venta y pago de proveedor;
- reimportacion idempotente;
- saldo contable vs extracto, partidas conciliatorias y corte;
- flujo de caja historico y proyeccion separados;
- periodo cerrado bloquea mutaciones.

### 7.7 Nomina, activos y otros ajustes

- empleado, devengados, deducciones, seguridad social y provisiones;
- lote de nomina electronica preparado, sin envio real no autorizado;
- compra de activo, depreciacion, deterioro, mejora, baja y venta;
- diferencia contable/fiscal e impuesto diferido cuando corresponda;
- gasto causado, anticipo, diferido, provision y ajuste;
- centro de costo y sucursal en transacciones seleccionadas.

## 8. Matriz profesional de reportes

Cada reporte se prueba vacio, con un registro, volumen, errores, permisos,
periodos abiertos/cerrados, comparativo, otra empresa, PDF/Excel/CSV/JSON/TXT,
impresion y recalculo independiente.

### 8.1 Libros y auxiliares

- comprobantes y consecutivos;
- libro diario, diario resumido, mayor y balance;
- auxiliar por cuenta, tercero, documento y centro de costo;
- balance de prueba general, por tercero y centro de costo;
- saldos anteriores, debitos, creditos y saldo final;
- cuentas de orden, naturaleza y niveles PUC;
- documentos no contabilizados, asientos pendientes y descuadrados.

### 8.2 Estados financieros completos

- estado de situacion financiera;
- estado de resultado y otro resultado integral;
- estado de cambios en el patrimonio;
- estado de flujos de efectivo por operacion, inversion y financiacion;
- notas, politicas, revelaciones y referencias a rubros;
- comparativo del periodo anterior;
- rubros configurables, materialidad y diferencias contable/fiscal;
- firmas de representante legal, contador y revisor fiscal cuando aplique;
- analisis vertical y horizontal;
- moneda, fecha de corte, empresa y estado borrador/aprobado.

Un “flujo de caja diario” operativo no se aprueba como estado de flujos de
efectivo. Un resumen de una fila no se aprueba como estado financiero completo.

### 8.3 Ventas, inventario, caja y costos

- ventas por producto, cliente, caja, sucursal, vendedor, canal y medio de pago;
- base, descuento, impuesto, total, recaudo, credito y devolucion;
- rentabilidad por producto con costo comprobable;
- Kardex valorizado, costo de ventas y existencia final;
- reporte de turno/cierre, sobrante/faltante y consolidado;
- conciliacion venta-documento-pago-caja-inventario-asiento.

### 8.4 Cartera y proveedores

- CxC/CxP general y por documento;
- edades y vencimientos;
- movimientos, abonos, asignaciones, anticipos y notas;
- estado de cuenta de cliente/proveedor;
- flujo esperado de recaudo/pago;
- diferencias entre submayor y cuenta control del PUC.

### 8.5 Fiscal y cumplimiento

- IVA generado, descontable, retenciones, ICA/ReteICA e INC si aplica;
- impuestos por tarifa, tercero, municipio, documento y periodo;
- certificados de retencion;
- borradores/formularios de trabajo de declaraciones, nunca presentacion
  automatica sin autorizacion;
- exogena por año, formato, concepto, tercero, topes y especificacion tecnica;
- XML/prevalidador, errores, reemplazos y correcciones;
- factura/nota/documento soporte/nomina y estados DIAN;
- conciliacion contable-fiscal y trazabilidad documental.

### 8.6 Gestion y auditoria

- presupuesto vs real;
- indicadores de liquidez, endeudamiento, rentabilidad y rotacion;
- activos y depreciacion;
- nomina y provisiones;
- obligaciones y calendario del contador;
- auditoria de cambios, cierres, reaperturas, exportes e IA;
- reportes programados, ejecuciones, errores y consistencia multiformato.

## 9. Fases de ejecucion

### P107-001 - Congelar alcance y datos [P0]

- confirmar SHA, entorno, empresa, roles, licencias y modulos;
- capturar inventario de reportes/endpoints/tablas vigente;
- acordar staging, datos reales permitidos, presupuesto IA y responsables.

Aceptacion: alcance firmado y cero mutacion antes de la ventana.

### P107-002 - Corregir permisos de contador e IA [P0]

- permitir `reportes:R` y uso IA a `contador`;
- mantener C/U/D/A denegados salvo permiso separado;
- confirmar consultas amplias para `admin_empresa`;
- agregar pruebas de rol, override, URL directa y tenant A/B.

Aceptacion: matriz de la seccion 5 demostrada en backend y navegador.

### P107-003 - Fuente contable canonica [P0]

- reconciliar eventos, asientos, comprobantes, PUC y submayores;
- eliminar fallback presentado como balance oficial;
- marcar preliminar/no contable cuando falten asientos;
- reconciliar fuente unica CxP del Plan 106.

Aceptacion: ningun estado oficial nace de aproximaciones ni fuentes dobles.

### P107-004 - Fabrica de fixtures contables [P0]

- crear generador idempotente `P107-QA` por empresa;
- producir saldos, terceros, productos, ventas, compras, cartera, bancos,
  nomina, activos e impuestos;
- incluir manifiesto esperado y procedimiento de reverso.

Aceptacion: dos ejecuciones no duplican datos y la limpieza queda auditada.

### P107-005 - Ciclo venta menta [P0]

- ejecutar los casos 7.2 como cajero y observar como contador;
- validar stock, costo, pagos, credito, caja, impuesto y asiento;
- repetir concurrencia e idempotencia.

Aceptacion: todas las ecuaciones de conciliacion cuadran a cero.

### P107-006 - Compras, proveedores y cartera [P0]

- ejecutar 7.4 y 7.5;
- completar la integracion canonica CxP si falta;
- probar estados de cuenta, edades, notas, anticipos y pagos.

Aceptacion: saldo por documento, submayor, mayor y banco coinciden.

### P107-007 - Impuestos y documentos Colombia [P0]

- implementar IVA descontable/retenciones/ICA faltantes;
- versionar parametros por vigencia y jurisdiccion;
- conciliar ventas, compras, notas, PUC y borradores de declaraciones;
- validar documentos electronicos primero en sandbox.

Aceptacion: contador recalcula muestras y no hay doble impuesto ni omisiones.

### P107-008 - Bancos, caja y cierre [P0]

- importar y reconciliar extractos;
- cerrar/reabrir periodo solo con evidencia en staging;
- comprobar bloqueo de mutaciones e impacto en reportes.

Aceptacion: extracto, bancos, caja y contabilidad tienen diferencias explicadas.

### P107-009 - Estados financieros completos [P0]

- implementar rubros, comparativos, notas, firmas y analisis;
- diferenciar reportes gerenciales de estados oficiales;
- validar conjunto completo segun grupo contable de la empresa.

Aceptacion: contador firma checklist de completitud y recalculo.

### P107-010 - Exogena y certificados [P0]

- implementar formatos/conceptos/topes por año;
- generar archivos compatibles con especificacion/prevalidador;
- validar identificacion de terceros, duplicados y correcciones;
- generar certificados conciliados con cuentas y terceros.

Aceptacion: cero error estructural y diferencias explicadas antes de presentar.

### P107-011 - Reportes IA dinamicos [P0]

- implementar `ReportSpec`, catalogo semantico, validador y compilador seguro;
- vista previa, aclaraciones, formulas, plantillas, versionado y auditoria;
- ejecutar la biblioteca de la seccion 6.

Aceptacion: reporte nuevo correcto sin agregar codigo ni permitir SQL del modelo.

### P107-012 - Exportacion, impresion y programacion [P1]

- probar JSON/CSV/TXT/XLS/PDF, hashes, filas y totales;
- validar A4/Carta/80 mm solo donde corresponda;
- probar programacion, zona horaria, destinatarios autorizados y reintentos.

Aceptacion: mismo dataset produce cifras iguales en todos los formatos.

### P107-013 - Seguridad, privacidad y rendimiento [P0]

- tenant A/B, prompt injection, PII, archivos, permisos y auditoria;
- limites de filas, costo, timeout, cancelacion y concurrencia;
- consultas pesadas, 429/5xx del proveedor y degradacion segura.

Aceptacion: cero fuga, cero mutacion por chat y objetivos de rendimiento aprobados.

### P107-014 - UAT del contador [P0]

Un contador distinto al implementador debe:

1. recibir acceso con rol `contador`;
2. revisar el ciclo completo sin ayuda tecnica;
3. solicitar cinco reportes existentes y cinco nuevos a la IA;
4. recalcular muestras;
5. exportar y firmar observaciones;
6. clasificar cada faltante P0/P1/P2.

Aceptacion: acta UAT `PASS`; todo P0 corregido y repetido.

### P107-015 - Produccion controlada empresa 12 [P0]

- ejecutar solo el subconjunto autorizado con menta COP 100;
- no emitir DIAN, mover dinero externo ni cerrar periodos sin aprobacion puntual;
- conciliar y revertir/limpiar mediante movimientos auditables;
- comprobar que no se afectaron otras empresas.

Aceptacion: saldo neto esperado, evidencia completa y cero residuo incoherente.

## 10. Evidencia por caso

Cada fila de la matriz debe registrar:

- ID, objetivo, requisito y riesgo;
- entorno, SHA/digest, empresa, rol y usuario saneado;
- datos iniciales y referencias `P107-QA`;
- pasos visibles, request correlacionado y resultado;
- ecuacion esperada, valor observado y diferencia;
- IDs de venta, documento, pago, movimiento, asiento y reporte;
- captura/PDF/archivo saneado y hash;
- estado `PASS`, `FAIL`, `BLOCKED` o `NA APROBADO`;
- defecto, severidad, responsable, correccion, repeticion y rollback.

Un HTTP 200, una pagina que abre o una respuesta convincente de IA no son
evidencia suficiente.

## 11. Compuerta final GO/NO-GO contable

GO contable solo si:

- [ ] rol `contador` y `admin_empresa` usan IA conforme a la matriz;
- [ ] IA genera reportes nuevos mediante `ReportSpec`, sin SQL libre;
- [ ] 43 datasets existentes y los nuevos pasan recalculo y formatos;
- [ ] venta de menta concilia venta, pago, caja, inventario, cartera, impuesto y asiento;
- [ ] CxC y CxP tienen una fuente y coinciden con cuentas control;
- [ ] impuestos incluyen ventas, compras, notas y retenciones;
- [ ] estados financieros completos tienen comparativos, notas y firmas;
- [ ] exogena/certificados cumplen vigencia y validacion estructural;
- [ ] bancos, caja, nomina, activos y cierre estan reconciliados;
- [ ] aislamiento A/B, idempotencia y concurrencia estan demostrados;
- [ ] UAT de contador esta aprobada;
- [ ] no quedan defectos P0/P1 ni datos QA incoherentes.

NO-GO automatico ante:

- cifra inventada o aproximacion presentada como contabilidad oficial;
- asiento descuadrado, evento sin contabilizar o diferencia sin explicar;
- doble fuente de cartera;
- fuga entre empresas o IA que acepte SQL/tenant del prompt;
- contador sin acceso a reportes/IA o con permisos de escritura accidentales;
- impuesto, exogena o documento fiscal declarado valido sin prueba normativa;
- PDF/Excel con totales diferentes;
- datos productivos creados sin conciliacion y reverso.

## 12. Cierre de esta planificacion

PCS tiene una base amplia, pero el estado actual es **NO-GO contable** hasta
completar las fases pendientes de estados financieros, impuestos, exógena,
CxP, validación A/B y UAT. Staging está disponible y el acceso de
Super Administrador fue comprobado el 2026-07-25. Powerful Control System
cuenta con licencia anual de 365 días activada en staging mediante un descuento
controlado de COP 600.000, sin DIAN ni pasarela. No se programan ni ejecutan
ventas hasta contar con una autorización puntual para la ventana de datos de
prueba P107-QA.
