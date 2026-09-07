# Runbook de release, despliegue y rollback

Estado: Vigente. Responsable: QA/operación y coordinación técnica. Revisión documental: 2026-09-05.

## Alcance y precondiciones

Procedimiento principal para preparar y ejecutar una liberación autorizada de PCS.
Consolida los antiguos runbooks breves de despliegue y rollback. No autoriza
ejecución por sí mismo. Leer [comandos](../../comandos_codex.md),
[checklist](../../release_checklist.md) y [estado actual](../../estado_actual.md).

Identificar responsable del cambio, observación, datos y decisión de rollback;
entorno destino, SHA/digests y esquema; ventana, impacto y criterio de parada.
Los secretos y configuración efectiva permanecen fuera del repositorio.

## Preparación sin alterar producción

1. Identificar rama, cambios locales y candidato revisado/integrado según las
   compuertas Git/PR del orquestador. No desplegar una rama mutable como evidencia
   de un artefacto aprobado. Revisar todos los cambios que integran el candidato.
2. Ejecutar la compuerta local del alcance y conservar resultados. Los comandos
   existentes se documentan en `documentos/comandos_codex.md`:

   ```powershell
   .\scripts\release_gate.ps1
   ```

   `-SkipE2E` permite una comprobación parcial cuando falta entorno local; no
   exonera aceptación E2E ni convierte el release en aprobado.
3. Verificar CI y revisiones: tests, race/vet cuando apliquen, vulnerabilidades,
   secretos, dependencias, Compose, imágenes, SBOM e IaC según checklist.
4. Fijar imágenes por digest y compatibilidad del esquema. Verificar migraciones
   sobre PostgreSQL vacío y clon anonimizado, checksum y roles runtime sin DDL.
5. Preparar staging aislado mediante sus scripts/overrides, por ejemplo
   `.\scripts\staging_up.ps1 -Build`, únicamente con entorno autorizado.
   Validar volúmenes, bases y destinos externos antes de arrancar workers.
6. Ejecutar E2E autenticado y contratos del candidato. No confundir un servidor
   estático, una página de otro runtime o un test omitido con aceptación PCS.
7. Comprobar backup y ensayo de restore pertinentes al candidato y datos,
   con RPO/RTO medidos. Definir recuperación de archivos privados y claves.
8. Completar checklist, responsables, criterios de rollback y aprobación real.
   Cualquier defecto bloqueante o evidencia obligatoria ausente detiene el release.

## Ejecución autorizada

1. Confirmar nuevamente entorno, candidato, ventana y backup. Si cambió el SHA,
   esquema o configuración material desde staging, repetir los gates afectados.
2. Usar el entrypoint canónico `scripts/rs.ps1` con las opciones del procedimiento
   vigente. `-FullPreflight` requiere la batería completa; no sustituirlo por
   invocar manualmente un subscript para eludir una compuerta.
3. Registrar resultado, digests realmente desplegados y migraciones. No imprimir
   variables privadas ni afirmar éxito cuando el orquestador falla.
4. Verificar salud y readiness, después login, autorización multiempresa y flujos
   del alcance: paneles, licencias/pagos, impresión y errores. Operaciones
   fiscales, de dinero, envíos o hardware requieren autorización específica.
5. Observar latencia, 5xx, colas, saturación e integraciones durante la ventana
   acordada. Detener ante regresión de seguridad, aislamiento o integridad.

## Rollback y recuperación

1. Declarar incidente y limitar nuevas mutaciones si existe riesgo de datos.
   Preservar auditoría y métricas minimizadas.
2. Comparar compatibilidad del último artefacto aprobado con el esquema actual.
   Si es compatible, volver a ese digest mediante el procedimiento del entorno;
   una rama mutable no es un artefacto de rollback.
3. Si el esquema impide revertir binario, elegir corrección hacia adelante o
   restauración autorizada. No ejecutar migraciones destructivas inversas como
   respuesta automática. Restaurar un backup puede perder escrituras posteriores:
   documentar impacto y conciliar antes de actuar.
4. Aplicar [recuperación Docker/VPS](runbook_recuperacion_desastre_docker_vps.md)
   para datos/archivos. Invalidar sesiones, tokens o credenciales afectadas
   cuando la causa lo exija.
5. Validar salud, aislamiento, conteos/invariantes contables y documentos antes
   de reabrir. Reconciliar operaciones externas ambiguas sin duplicarlas.

## Cierre

Registrar SHA/digests, migraciones, entorno, aprobación, pruebas/resultados,
omisiones, observación, incidente/rollback si ocurrió y responsable del cierre.
Actualizar estado de entrega solo con evidencia del artefacto publicado.

Último ensayo de este procedimiento consolidado: pendiente. Esta tarea revisó
documentación y scripts de referencia; no desplegó ni ejecutó restore.
