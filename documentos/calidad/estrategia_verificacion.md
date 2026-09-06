# Estrategia de verificación y aceptación

Estado: Vigente. Responsable: QA/operación e ingeniería del módulo. Revisión documental: 2026-09-05.

## Criterio

Probar el riesgo y el contrato del cambio. Un conteo de tests, un audit de patrones o una página visible no demuestra corrección de negocio. Relacionar cada evidencia con los [requisitos](../requisitos/especificacion_y_trazabilidad.md), entorno y candidato.

## Niveles de evidencia

| Nivel | Qué demuestra | Qué falta por sí solo |
| --- | --- | --- |
| Documental/estático | Referencias, sintaxis, contratos y ciertos patrones | Comportamiento, seguridad efectiva y runtime |
| Unitario Go/JS | Regla local en condiciones de test | SQL real, concurrencia e integración |
| PostgreSQL aislado | Migración, consultas, restricciones, locks y atomicidad bajo el caso | Equivalencia con esquema/datos productivos |
| UI autenticada | Flujo real por rol/tenant en runtime identificado | Proveedor, papel o hardware si no se ejecutaron |
| Staging | Integración del artefacto, configuración y datos controlados | Identidad con producción sin despliegue verificado |
| Proveedor/hardware | Resultado externo de la operación específica | Otras familias, dispositivos, tenants o candidatos |
| Producción | Comprobación autorizada del artefacto publicado | Cobertura fuera del alcance y fecha de esa prueba |

## Matriz por tipo de cambio

| Cambio | Casos mínimos según riesgo |
| --- | --- |
| Endpoint/permiso/dato empresarial | Sesión inválida, rol insuficiente, tenant B, hijo B, licencia, payload inválido y positivo A |
| Pago/caja/inventario | Doble submit y carrera, timeout, replay, rollback transaccional y conciliación de movimientos |
| Fiscal | Fuente inmutable, empresa/país/familia, numeración, firma/esquema del adaptador y aceptación oficial específica |
| Migración | Vacía, actualización de clon anonimizado, reejecución, checksum alterado, rol DDL y compatibilidad de binario |
| Worker/outbox | Lease vencido, caída antes/después del efecto, reintento, duplicado y recuperación |
| Archivos/IA | Acceso ajeno, ruta directa/traversal, tamaño/tipo, respuesta inválida, confirmación y privacidad |
| UI/impresión | Teclado, foco, errores, vacío/carga, móvil/escritorio, permisos; papel POS/carta según alcance |
| Release/infraestructura | CI, imágenes/digests, esquema, storage, staging, restore, observación y rollback |
| Solo documentación | Exactitud humana, catálogo, enlaces, UTF-8 y diff; no requiere ejecutar negocio |

## Ejecución reproducible

Leer [comandos](../comandos_codex.md). Desde `backend/`, seleccionar primero paquetes y pruebas afectadas; ampliar para una entrega transversal cuando proceda:

```powershell
go test ./... -count=1 -timeout 5m
go vet ./...
```

La batería puede omitir integración si faltan DSN/esquemas/XSD u otros requisitos. Leer los resultados y conservar las omisiones. `-race` necesita plataforma/toolchain compatibles, incluido soporte CGO cuando corresponda; un fallo por entorno no prueba ausencia de carreras. CI Linux debe aportar esa evidencia si el cambio lo exige.

No poner secretos en línea de comandos ni logs. Usar recursos aislados y evitar que tests o workers alcancen proveedores reales sin alcance autorizado. Una prueba de cobro/fiscal no se vuelve inocua por llamarse «test».

## Carga y recuperación

Para [SLO/RTO/RPO](../gobernanza_tecnica/slo_sla_operativo.md), registrar hardware, réplicas, dataset por tenant, conexiones, mezcla de rutas, duración, p95, error rate, colas y saturación. Ejecutar carga en entorno controlado con criterio de parada. Una prueba breve no acredita disponibilidad mensual.

Medir restore con ambas bases y archivos privados, verificar descifrado, integridad y aislamiento; documentar tiempo efectivo y punto de datos recuperado. Conservar evidencia privada apropiada, con resumen seguro en repositorio.

## Defectos y aceptación

Defecto: ID, requisito, severidad/impacto, reproducción segura, esperado/observado, archivos, entorno/candidato, evidencia, responsable, mitigación y prueba de cierre. No borrar hallazgos al cambiar el candidato; cerrarlos con evidencia nueva.

El reporte final enumera pruebas ejecutadas, fallidas y omitidas con motivo. La aceptación funcional requiere responsable y alcance. La autorización productiva sigue la [checklist de release](../release_checklist.md). Los resultados históricos de la auditoría del 2026-09-05 no se vuelven a presentar como pruebas ejecutadas por una tarea documental.
