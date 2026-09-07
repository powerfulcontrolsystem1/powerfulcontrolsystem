# Gestión de incidentes y continuidad

Estado: Vigente. Responsable: QA/operación y seguridad. Revisión documental: 2026-09-05.

## Activación y responsables

Aplicar ante caída, degradación, pérdida de aislamiento/integridad, exposición de secretos, fallo de backup o resultado externo ambiguo. Las severidades y objetivos están en [SLO](../gobernanza_tecnica/slo_sla_operativo.md); vulnerabilidades y reporte privado en [SECURITY](../../SECURITY.md).

Nombrar coordinador del incidente, operador técnico, responsable de datos/seguridad y responsable de comunicaciones. Registrar nombres y contacto de guardia en el sistema privado de operación. Esta guía define funciones; no acredita una guardia ni atención contractual ya implementadas.

## Procedimiento

1. Registrar hora/zona, síntoma, servicios y tenants potencialmente afectados, último cambio y candidato identificado. No copiar payloads, cookies o secretos.
2. Confirmar mediante [observabilidad](../observability_runbook.md): salud, latencia, errores, saturación, colas y dependencias. Separar hechos de hipótesis.
3. Contener con alcance mínimo. Si hay riesgo de integridad, limitar nuevas mutaciones; preservar auditoría y operaciones pendientes. No reintentar cobros o documentos a ciegas.
4. Elegir mitigación y recuperación: configuración, corrección, rollback de artefacto compatible o restore autorizado. Registrar quién decide y posibles pérdidas de escrituras.
5. Ejecutar el procedimiento específico del [índice de runbooks](../gobernanza_tecnica/runbooks/README.md). Verificar entorno antes de acciones destructivas o externas.
6. Confirmar salud, permisos, aislamiento, saldos/documentos y archivos; reconciliar efectos pendientes con proveedor. Reabrir gradualmente con observación del responsable.
7. Comunicar estado y alcance verificados por el canal autorizado. No enviar comunicaciones a clientes o terceros solo por leer este documento.

## Recuperación y rollback

El [runbook de recuperación](../gobernanza_tecnica/runbooks/runbook_recuperacion_desastre_docker_vps.md) describe restore Docker/VPS. El [runbook de release](../gobernanza_tecnica/runbooks/runbook_release_profesional.md) centraliza rollback de artefactos y compatibilidad con esquema.

Registrar backup y fecha de datos recuperables, ambas bases, archivos, claves recuperables por vía privada, conteos/invariantes, tiempo de inicio/fin y RTO/RPO observados. Probar en red aislada antes de dar por disponible la recuperación. La existencia de un backup o el éxito de un comando no demuestra restauración del negocio.

## Postincidente y ejercicios

El registro de postincidente incluye línea de tiempo, impacto confirmado, causa y factores contribuyentes, decisiones, recuperación, evidencia minimizada, acciones con responsable/plazo y prueba de cierre. Evitar atribuir culpa; explicar condiciones verificables.

Revisar ejercicios después de cambios relevantes de esquema/storage/infraestructura y en la cadencia acordada por operación. Cada ejercicio documenta entorno, alcance, salidas externas bloqueadas, resultados y brechas. No asignar un resultado RTO/RPO sin medición ni prometer una frecuencia organizacional no acordada.
