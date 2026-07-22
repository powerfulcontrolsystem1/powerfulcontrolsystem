# Matriz de pruebas tenant A/B - Plan 105

Fecha de corte: 2026-07-21. Esta matriz separa controles existentes de la
evidencia que falta para cerrar P105-004. No crea empresas, usuarios ni datos.

## Cobertura ya comprobada en codigo

| Control | Evidencia actual | Limite |
| --- | --- | --- |
| Registro de rutas | `tools/tenant_route_inventory.mjs --check`: 203 rutas empresariales, 203 con wrapper detectado, 0 manuales | Detecta wrapper, no filtro SQL, archivos, cache ni job. |
| Query/cabecera/JSON | `empresa_permisos_tenant_test.go` rechaza `empresa_id` cruzado y repetido | No ejecuta CRUD ni consulta una BD real. |
| Cuerpo restaurado | `empresa_permisos_licencias_test.go` verifica rechazo y preservacion del JSON para el handler | Cubre una forma de payload, no todos los endpoints. |
| API movil | `mobile_api_v1_test.go` normaliza `empresa_id` del cuerpo al tenant de query | No prueba acceso de objeto B por identificador valido. |

## Matriz obligatoria de ejecucion en staging

Crear Empresa A y Empresa B anonimizadas, un administrador y cajero equivalentes
por empresa, y marcadores unicos no sensibles. Cada caso debe conservar
`request_id`, usuario, empresa efectiva, identificador objetivo, status y una
consulta posterior que demuestre que B no cambio.

| Familia P0 | Manipulacion A contra B | Resultado requerido |
| --- | --- | --- |
| Carritos, venta, caja, productos e inventario | `empresa_id`, `carrito_id`, item, estacion, codigo y payload JSON de B | 403/404 sin nombre, total, stock ni mutacion de B. |
| Pagos, datafonos y comprobantes | pago/transaccion/webhook/replay de B | sin cobro, conciliacion ni cambio de estado en B. |
| Facturacion y documentos | venta/factura/CUFE/archivo/exportacion de B | sin lectura, descarga, emision ni metadatos de B. |
| Usuarios, roles, licencias y empresas compartidas | usuario/rol/permisos de B desde sesion A | fallo cerrado, sin elevacion ni enumeracion. |
| Archivos, Nextcloud y backups | URL, token, archivo, backup o exportacion de B | 403/404 y archivo no descargado. |
| Reportes, IA, cache y jobs/outbox | filtro, cursor, clave cache, job/evento de B | datos solo A; job no ejecuta ni notifica para B. |
| Estaciones, reservas y control electrico | estacion/reserva/dispositivo de B | sin cambio operativo ni despacho externo. |

## Reglas de implementacion para Terra

1. Empezar con una sola familia P0 y fixtures transaccionales aisladas; no usar
   una empresa productiva ni IDs que puedan coincidir con datos reales.
2. Construir helpers que reciban A, B y rol, ejecuten solicitud correcta y
   cruzada, y consulten la fila/archivo/evento posterior.
3. Probar query, path, cabecera, JSON, formulario, multipart, ID repetido,
   cursor y URL compartible cuando existan en la familia.
4. Esperar de forma acotada los jobs y verificar que no existe evento/outbox de
   B; no usar `Sleep` fijo como unica prueba.
5. Integrar el caso en CI solo despues de que el fixture PostgreSQL efimero sea
   reproducible. Guardar en staging evidencia redactada, no dumps ni PII.

P105-004 sigue abierto hasta cubrir estas familias con BD real y superficie HTTP
completa. El inventario de rutas no debe usarse como sustituto de esta matriz.
