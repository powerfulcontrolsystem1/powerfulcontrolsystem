# P109-007 - Diagnóstico DIAN de staging PCS

Fecha: 2026-07-31
Entorno: staging autenticado, empresa Powerful Control System (`empresa_id=12`).

## Hallazgo

El checklist DIAN respondió correctamente, pero sin configuración persistida.
Faltan los datos legales, resolución, rango/consecutivo, llave técnica, URL,
referencias de certificado, `test_set_id`, software ID y PIN. No se revelaron
secretos ni se cambió configuración.

La pantalla intentaba consultar vencimiento de certificado y resolución aun sin
un registro DIAN; el backend respondió 400 de forma correcta y el navegador
mostraba esos errores de configuración como recursos fallidos.

## Corrección propuesta

La pantalla solo consulta vencimientos cuando `state.dianConfig` representa un
registro persistido. Sin él muestra una guía para configurar y guardar los datos
DIAN primero, sin invocar endpoints de vencimiento.

La suite focalizada de backend aprobó los estados de vencimiento de certificado
y los contratos de anulación por nota crédito, sin hacer llamadas externas.

## Límite de cierre

P109-007 sigue bloqueada hasta que se carguen por el flujo seguro los datos y
credenciales reales de PCS. Luego se requiere validación, prueba DIAN oficial,
estado `GetStatusZip=00`, impresión y anulación fiscal autorizadas.
