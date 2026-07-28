# P108-009 - ReportSpec IA: prueba autenticada en staging

Fecha: 2026-07-28
Ambiente: staging autorizado
Empresa: 12 (Powerful Control System)
Rol: super_administrador autenticado
Ruta: `administrar_empresa/reportes_ejecutivos.html?empresa_id=12`

## Caso ejecutado

Se verifico visualmente el catalogo publicado de 46 reportes. Luego se envio
una solicitud de solo lectura para generar una vista previa de cuentas por
pagar por proveedor, con saldo y vencimiento, ordenada de mayor a menor.
No se exporto, guardo plantilla ni modifico dato de negocio.

## Resultado observado

La pantalla mostro el estado controlado:

`No se pudo generar el reporte: Ocurrio un problema interno. Intenta de nuevo en unos segundos.`

El backend respondio HTTP 502. El registro de staging confirma que no pudo
resolver la credencial OpenAI porque `CONFIG_ENC_KEY_ID` no cumple el formato
seguro requerido. La inspeccion sin secretos confirmo que el valor activo
incluye los delimitadores del marcador de ejemplo y que no existen credenciales
OpenAI cifradas disponibles en la configuracion de staging.

## Decision

Resultado: **bloqueado de forma segura**. No se genero `ReportSpec`, vista
previa, exportacion ni plantilla; tampoco se reintento para consumir cuota.
No se cambia la llave, el identificador ni la API key por inferencia.

## Requisito para reanudar

El responsable de staging debe configurar, mediante el canal seguro:

1. un `CONFIG_ENC_KEY_ID` real que solo use letras, numeros, guion o guion
   bajo y que corresponda a la llave de cifrado activa;
2. una credencial OpenAI valida, preferiblemente registrada desde Super
   administrador > IA despues de validar la llave;
3. reinicio controlado de API y worker de staging y prueba de lectura de la
   configuracion sin exponerla.

Despues se repetiran: caso valido, rechazo de inyeccion, campo inexistente,
empresa B, rol sin permiso, cancelacion, doble clic, vista previa y exportacion
del mismo `ReportSpec` validado.

Estado de fase: **parcial; no aprobada**.
