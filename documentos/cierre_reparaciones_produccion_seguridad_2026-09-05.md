# Cierre de reparaciones del informe general de producción y seguridad

Fecha: 5 de septiembre de 2026. Alcance: correcciones locales derivadas de H01-H12 del informe general de la misma fecha.

## Dictamen actualizado

El código queda sustancialmente endurecido y las pruebas locales pasan, pero el candidato continúa en **NO-GO productivo** hasta completar las compuertas externas y operativas descritas al final. Este cierre no equivale a despliegue, pentest, prueba de carga, restauración ni homologación con proveedores o hardware.

## Seguridad, autenticación y separación por empresa

Estado local: **reparado en las fronteras identificadas**.

- H01: `golang.org/x/crypto` se actualizó a `v0.56.0`. `govulncheck ./...` informa cero vulnerabilidades alcanzables.
- H02: los CRUD genéricos validan referencias secundarias con `(empresa_id,id)` antes de escribir. La cobertura incluye lotes/series, devoluciones, RRHH, CRM, BOM, logística, firmas, CxC y CxP. Producción/MRP, WMS y Tesorería validan también padres, productos, ubicaciones y cuentas dentro del tenant.
- H03: los DTO genéricos descartan `empresa_id`, `id`, autoría, estados y campos calculados o de aprobación controlados por servidor. La creación obtiene el actor de la sesión y fija el estado inicial. Producción y WMS impiden crear órdenes o ítems con avances y estados finales enviados por el cliente.
- H04: la cuota empresarial se consume después de comprobar identidad y acceso. El contador compartido se persiste en `empresa_api_rate_limits` mediante UPSERT atómico por empresa, ámbito y ventana; un error de la base cierra la petición con 503.
- H06: el acceso global privilegiado exige TOTP confirmado, la lectura de política falla cerrada y esos roles no pueden desactivar su segundo factor desde el flujo ordinario.

Validación todavía necesaria: ejecutar pruebas negativas A/B con PostgreSQL restaurado y sesiones reales de dos empresas, incluido Super delegado, importaciones y jobs. Antes de desplegar la política MFA deben estar enroladas todas las cuentas privilegiadas y debe probarse el procedimiento de recuperación para evitar bloqueo operativo.

## Producción/MRP, WMS y Tesorería

Estado local: **reparado; falta prueba concurrente con PostgreSQL real**.

- H08: creación de órdenes, materialización de consumos, registro de consumos/calidad, cambios de estado, ítems WMS, avances, despachos, eventos y recálculos usan transacciones y bloqueos de fila.
- Los consecutivos MRP se calculan bajo bloqueo asesor por empresa, sin `COUNT(*)+1` concurrente.
- La regeneración del plan MRP reemplaza e inserta todo en una sola transacción.
- Las recetas cargan componentes por lote y el generador reutiliza esa agrupación, eliminando el N+1 señalado.
- Los errores que antes se ignoraban en efectos críticos ahora cancelan la transacción.

Validación todavía necesaria: doble envío, dos operadores simultáneos, fallo deliberado entre pasos, planes SQL y volumen representativo en una base desechable.

## Auditoría

Estado local: **endurecimiento parcial**.

- H11: se retiraron las goroutines sin seguimiento; el intento de persistencia termina antes de que finalice la solicitud y ya no acumula trabajo fuera del ciclo HTTP.
- Los movimientos MRP/WMS incorporados en esta reparación guardan cambio y evento en la misma transacción.

Pendiente antes de declarar trazabilidad crítica integral: conectar todas las operaciones críticas restantes a su transacción de negocio o a una outbox durable con reintento. La auditoría automática general continúa siendo auxiliar y registra el error si su escritura falla; no puede usarse como única evidencia financiera o forense.

## Dependencias, navegador y borde HTTPS

Estado local: **vulnerabilidad corregida y configuración endurecida; CSP heredada pendiente**.

- H01 queda resuelto en el módulo Go y el escáner no encuentra llamadas vulnerables.
- H07: la plantilla Nginx añade HSTS y limita `connect-src` e `img-src` a orígenes explícitos.
- La nueva vista de auditoría de pagos movió su CSS a un archivo externo. Los dos portales públicos de login trasladan su bootstrap a JavaScript externo y eliminan sus scripts y estilos estáticos inline. El inventario baja de 1.219 a 1.203 bloqueadores, sin regresiones contra la línea base.

Pendiente: el frontend heredado conserva 1.203 bloqueadores de CSP estricta (203 scripts inline, 151 bloques de estilo y 849 atributos `style`). Por compatibilidad se mantiene `unsafe-inline` hasta migrar esas páginas por módulos. Debe verificarse en el proxy TLS publicado que HSTS y CSP aparezcan también en errores y archivos estáticos.

## CRM y Domótica

Estado local: **paridad de estados corregida; falta desplegar y repetir la UI**.

- Los formularios generales de leads, seguimientos, cotizaciones y campañas muestran el estado actual como solo lectura. Las mutaciones de ciclo de vida pasan únicamente por `action=transicionar`, que aplica la máquina de estados del backend.
- La prueba autenticada de Domótica reprodujo `isDoorSensor is not defined` en el runtime publicado. El árbol local inicializa esa variable antes de construir cada tarjeta y conserva una prueba de contrato para evitar la regresión.
- Las APIs autenticadas de CRM y Domótica respondieron 200; no se enviaron mensajes, transiciones, comandos, pruebas de GPIO ni acciones sobre Raspberry.

## Base de datos, backups e infraestructura

Estado local: **separación de privilegios y escalado de una instalación preparados; validación operativa pendiente**.

- H05: `pcs-migrate` crea o rota un rol de respaldo dedicado, de solo lectura, sin superusuario, creación de roles/BD ni `BYPASSRLS`. API y worker no reciben la clave del propietario para snapshots.
- Los snapshots usan `pg_dump` de las dos bases con el rol dedicado, sin propietarios ni privilegios embebidos.
- El alcance `vps` del backup exige y empaqueta los volúmenes privados de archivos PCS, certificados Mailu y datos/librerías/logs de OnlyOffice aunque el contenedor asociado esté detenido; si falta un artefacto, el backup falla de forma visible.
- H09: se retiraron nombres fijos de API/worker, las réplicas son configurables, el worker usa el hostname como identidad y Nginx balancea el servicio con `least_conn`.
- El arranque multirréplica acepta únicamente almacenamiento privado realmente compartido. El modo `object` declarativo se rechaza hasta que exista un adaptador implementado.
- Los scripts operativos localizan contenedores por etiquetas de Compose y el chequeo de arranque contempla todas las réplicas de API.

Pendiente: crear las credenciales secretas del rol de backup en el entorno, aplicar migraciones, validar `docker compose config` y `bash -n` en Linux, probar restauración aislada, pérdida de réplica, almacenamiento compartido, rotación de secretos y límites de conexiones. El sidecar con socket Docker conserva privilegio sobre el host y debe ejecutarse solo en la red/host administrativos previstos.

## Rendimiento y cancelación

Estado local: **reparación dirigida, deuda transversal pendiente**.

- H10: los CRUD genéricos afectados ahora propagan contexto a consultas y escrituras; MRP eliminó el N+1 identificado.
- Los límites de pool existentes continúan aplicándose.

Pendiente: el inventario global original registró cientos de llamadas históricas sin contexto. No se modificaron masivamente porque cada contrato necesita plazos y semántica de cancelación propios. Antes de aumentar réplicas deben medirse carga, conexiones agregadas, consultas lentas y endpoints de reportes/exportación, y migrarse primero las rutas de mayor tráfico.

## Liberación y evidencia

Estado local: **compuertas reparadas; candidato aún no certificado**.

- H12: se regeneraron OpenAPI (325 rutas), matriz multiempresa (197 rutas, 0 revisiones manuales), inventario Ensure (142 funciones/117 pasos) e inventario runtime Ensure (70 llamadas).
- El preflight completo ahora detecta cualquier prueba Go omitida; el modo estricto exige `-race` en Linux y `govulncheck` disponible.
- La auditoría estricta de migraciones pasa sin alterar cuerpos históricos.

Evidencia local ejecutada:

- `go test ./... -count=1 -timeout=12m`: aprobado en todos los paquetes.
- `go vet ./...`: aprobado.
- `go mod verify`: aprobado.
- `govulncheck ./...`: 0 vulnerabilidades alcanzables; existe un aviso solo a nivel de módulo sin llamada alcanzable.
- `migration_audit.mjs --base-ref origin/main --strict`: aprobado.
- `profesional_preflight.ps1 -SkipDockerConfig -RequireMigrationAudit`: aprobado, reporte `preflight_20260905_164052.md`.
- Inventario CSP: 1.219 bloqueadores heredados y 0 regresiones.
- Sintaxis PowerShell y `git diff --check`: aprobados.

## Compuertas obligatorias que mantienen el NO-GO

1. Integrar un árbol limpio y fijar el SHA y los digests exactos de imágenes.
2. Aplicar `20260905-002-empresa-rate-limit-v1` con `pcs-migrate` y verificar el ledger en ambas bases.
3. Ejecutar en Linux `docker compose config`, validación `bash -n`, `go test -race ./...`, pruebas PostgreSQL/XSD sin omisiones y escaneo de las imágenes finales.
4. Probar aislamiento A/B autenticado, concurrencia MRP/WMS, carga y caída de réplicas sobre el candidato exacto.
5. Enrolar MFA privilegiado, provisionar el rol de backup y demostrar una restauración aislada con credenciales de mínimo privilegio.
6. Verificar HSTS/CSP en el dominio desplegado y completar por módulos la retirada de `unsafe-inline`.
7. Ejecutar aceptación real de DIAN, pagos, correo/WhatsApp, Raspberry y demás proveedores o hardware dentro del alcance que se vaya a publicar, sin fabricar registros comerciales.

Hasta completar esas siete compuertas, los cambios pueden pasar a integración y staging, pero no deben anunciarse como producción general preparada contra ataques.
