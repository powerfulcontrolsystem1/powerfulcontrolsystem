# P108-015 - Preflight DIAN autorizado de PCS

Fecha: 2026-07-26  
Empresa: Powerful Control System (`empresa_id=12`)  
Entorno: producción, sesión administrativa autorizada por el usuario.

## Alcance ejecutado

Se abrió visualmente el Centro de habilitación DIAN y se ejecutó la acción
oficial `Validar credenciales`. No se enviaron documentos, no se ejecutó el set
automático, no se reintentó cola y no se cargaron, descargaron ni modificaron
certificados.

## Resultado observado

La pantalla muestra ambiente de producción, rango activo y consecutivo actual,
pero estado DIAN rechazado. La validación respondió `ok=false` con los problemas
`certificado_clave_ref invalido` y `certificado_url invalido`.

La emisión real quedó bloqueada antes de crear una venta o factura. Por tanto:

- no se seleccionó ni vendió producto alguno;
- no se consumió consecutivo;
- no se transmitió una factura a DIAN;
- no existe representación gráfica de una nueva factura para validar.

## Decisión

No es seguro ni correcto forzar un reintento mientras la referencia de clave y
el certificado no superen el preflight. La siguiente prueba real debe hacerse
solo después de corregir el candidato, desplegarlo de forma controlada y volver
a validar credenciales. Una aceptación requiere el acuse oficial `GetStatusZip`
con resultado aceptado, además de revisar visualmente Carta, compacta y POS.

## Diagnóstico técnico y candidato de reparación

La inspección del VPS confirmó que la API corre como el usuario sin privilegios
`pcs` (`uid/gid 10001`), mientras las llaves PEM heredadas de DIAN y su carpeta
eran propiedad de `root` con modo `0700`. La configuración de PCS seguía
apuntando al volumen correcto, pero el proceso no podía atravesar o leer esos
archivos.

El candidato incorpora una migración operativa restringida por `empresa_id` y
categoría `dian`. Primero corrige temporalmente solo la carpeta de firma de la
empresa seleccionada; después mueve las referencias de clave y certificado al
volumen privado, guarda archivos con modo `0600`, actualiza las referencias de
esa misma fila empresarial y retira los PEM heredados del árbol público.

También se corrigió el flujo de anulación: una factura electrónica colombiana
ya no admite la transición local genérica `anular`. La interfaz exige motivo y
confirmación, crea la nota crédito y conserva la factura como emitida mientras
DIAN no acepte la nota. Reintentos y conciliación finalizan la anulación solo
cuando el estado fiscal sea `aceptado`.

La emisión automática desde una venta aplica la misma compuerta: `enviado` es
un estado intermedio y no se presenta como confirmación fiscal; el documento
solo queda confirmado con `aceptado`.

## Aislamiento multiempresa verificado en candidato

- las rutas DIAN y de facturación electrónica permanecen detrás del middleware
  empresarial de licencia, rol y alcance;
- configuración, numeración, documentos, cola y acuses se consultan con
  `empresa_id`;
- la migración operativa exige una empresa positiva y agrega
  `AND empresa_id = $1` a cada consulta;
- el script de VPS solo coincide con la carpeta heredada de la empresa
  indicada y rechaza una carpeta simbólica;
- las pruebas de contrato y las suites enfocadas de handlers/DB aprobaron antes
  del despliegue.

Estos resultados describen el candidato local. La prueba real, el `rs`, la
migración de PCS, la venta controlada, el acuse `GetStatusZip` y las
visualizaciones se registrarán únicamente después de observarlos.
