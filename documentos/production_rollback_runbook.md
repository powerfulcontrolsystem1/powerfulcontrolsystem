# Runbook de rollback de produccion

1. Declarar incidente y detener nuevas mutaciones cuando exista riesgo de datos.
2. Preservar auditoria y metricas sin copiar secretos a tickets o chat.
3. Volver al ultimo artefacto inmutable aprobado; no usar una rama mutable como
   sustituto de release.
4. Restaurar datos solo desde backup probado y siguiendo el RPO/RTO aprobado.
5. Invalidar sesiones, tokens o credenciales afectadas cuando el incidente lo
   requiera.
6. Validar salud, aislamiento por empresa y conteos contables antes de reabrir.
