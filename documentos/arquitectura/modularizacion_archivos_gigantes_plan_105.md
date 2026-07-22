# Inventario de modularizacion - P105-018

Fecha de linea base: 2026-07-22. Medicion: lineas fisicas de Go bajo
`backend`.

| Prioridad | Archivo | Lineas | Primer corte seguro |
|---|---|---:|---|
| P0 | `backend/handlers/modulos_faltantes.go` | 13.444 | Extraer configuraciones y CRUD generico por dominio, con pruebas de contrato por ruta. |
| P0 | `backend/handlers/payments_handlers.go` | 6.900 | Separar Wompi, Epayco, callbacks y presentacion; conservar idempotencia. |
| P0 | `backend/handlers/reportes.go` | 6.675 | Separar programacion, exportacion y reportes ejecutivos sin cambiar permisos. |
| P0 | `backend/db/productos.go` | 5.654 | Extraer lectura, escritura y stock transaccional con pruebas de empresa. |
| P0 | `backend/db/carritos_compras.go` | 4.020 | Extraer reserva, cierre e inventario conservando locks/transacciones. |
| P0 | `backend/handlers/carritos_compras.go` | 3.879 | Separar endpoints de lectura, mutacion y caja; preservar CSRF/permisos. |
| P0 | `backend/handlers/empresa_permisos.go` | 3.869 | Extraer catálogo, roles, licencias y middleware sin modificar fallos cerrados. |
| P1 | `backend/db/creditos.go` | 3.499 | Extraer cartera, límites y aprobación luego de pruebas de duplicado. |
| P1 | `backend/db/chat_inteligencia_artificial.go` | 3.424 | Separar conversación, ejecución y auditoría, manteniendo `empresa_id`. |

## Regla de ejecución

Cada extracción requiere un PR/commit separado, prueba de caracterización antes
del movimiento, `go test ./... -count=1`, `go vet ./...` y una revisión de
permisos, transacciones y `empresa_id`. Ninguna extracción P0 forma parte del
SHA de producción si introduce cambio funcional no verificado.
