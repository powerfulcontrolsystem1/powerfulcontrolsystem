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
