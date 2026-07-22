# Inventario de errores publicos - Plan 105

Fecha de corte: 2026-07-21. Revision estatica; no sustituye pruebas de respuesta
HTTP ni inspeccion de logs.

## Linea base reproducible

```powershell
rg -n 'err\.Error\(\)' backend -g '*.go'
rg -n 'http\.Error\([^\r\n]*err\.Error\(' backend/handlers -g '*.go'
rg -n '"(error|message|detail|missing_or_invalid)"\s*:\s*[^\r\n]*err\.Error\(' backend/handlers -g '*.go'
```

- 1.603 usos totales de `err.Error()` en backend; el numero incluye logs y
  retornos internos, por lo que no equivale por si solo a una fuga.
- 1.266 usos en la misma linea de `http.Error(...)` en handlers, distribuidos
  en 133 archivos.
- 320 de esos usos devuelven un estado 5xx (`500`, `502`, `503` o `504`): son
  P0 porque el error puede contener SQL, rutas, proveedores, PII o secretos.
- Linea base inicial: 75 campos JSON identificados exponian `err.Error()`
  directamente. Ocho lotes de redaccion redujeron el inventario vigente a 49;
  revisar tambien funciones que devuelven mapas y los serializan fuera de la
  linea.

## Primeras familias P0

| Archivo | 5xx directos | Riesgo | Orden de correccion |
| --- | ---: | --- | --- |
| `backend/handlers/productos.go` | 59 | SQL/inventario y datos empresariales | 1 |
| `backend/handlers/payments_handlers.go` | 46 | proveedor, conciliacion y pagos | 1 |
| `backend/handlers/super_chat_ia_logica.go` | 29 | proveedor IA y configuracion global | 2 |
| `backend/handlers/system_empresas_handlers.go` | 21 | alta/configuracion multiempresa | 2 |
| `backend/handlers/usuarios_empresa.go` | 19 | identidad y permisos | 2 |
| `backend/handlers/pagina_principal_handlers.go` | 16 | superficie publica | 3 |
| `backend/handlers/voice_stream_config.go` | 15 | credenciales/configuracion de voz | 1 |

Los JSON directos requieren priorizar adicionalmente `empresa_db_admin.go`,
`security_vps_handlers.go`, `super_mantenimiento_agentes.go`,
`super_correos_masivos.go`, `datafonos.go` y controladores IA. No asumir que
todos son seguros por usar un error de validacion: clasificar el origen antes de
incluirlo en una allowlist.

## Patron obligatorio de correccion

1. Distinguir error de validacion estable (400/409) de fallo interno o externo.
2. Para 5xx, responder solo `{ok:false, code:"...", request_id:"..."}` o un
   mensaje generico equivalente; nunca concatenar la causa.
3. Registrar internamente causa, operacion, `request_id` y `empresa_id` ya
   autorizado mediante redaccion existente. No registrar payload ni secretos.
4. Mantener los mensajes de validacion solo si provienen de una taxonomia
   cerrada local, no de SQL, proveedor, filesystem ni `error` envuelto.
5. Agregar por familia una prueba con canarios (`postgres://`, ruta privada,
   token y correo) que confirme ausencia en cuerpo, cabeceras y salida JSON.
6. Tras cada lote, disminuir el conteo P0 y mantener una allowlist con motivo,
   propietario y fecha de revision. El gate final debe impedir regresiones.

## Criterio de cierre

No cerrar P105-006 con un reemplazo global. Deben quedar cero 5xx/JSON que
serialicen causas no clasificadas, pruebas de canarios en las familias P0 y una
correlacion interna que permita soporte sin revelar detalles al usuario.
