# P110-001 — Invariantes CxP en candidato e308ca4b

Fecha: 2026-08-12 (America/Bogota)  
Entorno: candidato inmutable de staging; producción no intervenida.

## Comprobaciones

- El workflow de candidato del SHA `e308ca4b` aprobó construcción, escaneo,
  SBOM, publicación y validación de Compose.
- El PostgreSQL de staging contiene `empresa_cxp_pagos`, confirmando que el
  migrador del candidato dejó disponible el esquema atómico esperado.
- La suite focalizada de `backend/db` aprobó las invariantes de ledger,
  idempotencia, bloqueo por empresa, reconciliación solo lectura y rechazo de
  escrituras CxP históricas.

## Límite

La evidencia es técnica y no reemplaza la conciliación de datos reales ni la
aceptación independiente del contador. P110-001/P110-002 permanecen
**parciales** y el resultado global continúa **NO-GO**.
