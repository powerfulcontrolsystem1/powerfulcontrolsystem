# P109-008 - Restore integral aislado del snapshot vigente

Fecha: 2026-08-09
Ambiente: VPS de staging, contenedor PostgreSQL temporal
Snapshot: `20260809_031501`
Alcance: validación aislada; no se modificaron bases, volúmenes ni servicios
activos de staging o producción.

## Ejecución

Se ejecutó el validador controlado con restauración temporal y comprobación de
datos críticos. El recurso temporal se elimina automáticamente al finalizar.

```powershell
.\scripts\vps_restore_validation.ps1 -ExecuteDrill -VerifyCriticalData -AllowRemoteTarget
```

## Resultado observado

- El dump `postgres_all.sql.gz` pasó su prueba de integridad.
- Los nueve tarballs obligatorios pasaron lectura completa; los dos tarballs de
  certificados opcionales no están configurados en este snapshot.
- La restauración temporal confirmó las dos bases PCS esperadas.
- Se verificaron cinco tablas críticas con filtro `empresa_id=12`:
  CxP, asientos contables, memoria IA, configuración DIAN y documentos de
  gestión.
- El inventario privado contenía cinco archivos y los tres soportes IA con
  checksum tuvieron coincidencia exacta con su referencia restaurada.
- RTO observado: 27 s. RPO observado: 50.544 s (14 h 2 min 24 s), dentro de
  los objetivos publicados de 2 h para RPO y 24 h para RTO.
- El contenedor de restore se retiró por el `trap` del validador al cerrar.

## Conclusión y límite

La recuperación de base, metadatos y soportes privados queda aprobada para el
snapshot indicado. P109-008 continúa **parcial**: aún exige la prueba de subida
en réplica A/descarga en B, pérdida de réplica y rollback coordinado de la
aplicación/base antes de acreditar la compuerta completa.
