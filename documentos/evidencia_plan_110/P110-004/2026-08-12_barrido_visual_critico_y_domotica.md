# P110-004 — barrido visual crítico y domótica

Fecha: 2026-08-12  
Ámbito: PCS/staging autenticado; escritorio y móvil. Las acciones mutantes se
omitieron deliberadamente.

## Barrido web

Se revisaron Panel, Finanzas/CxP, Facturas electrónicas, Centro DIAN y
Domótica en 1366 px y 390 px. Resultado: **10/10 vistas OK**, 610 botones
inventariados, dos clics inequívocamente seguros, 251 acciones mutantes
omitidas, cero mutaciones automáticas y cero errores de página.

La ruta histórica `cuentas_por_pagar.html` se retiró de la configuración del
auditor porque CxP vive dentro de `finanzas.html`; no se cambió una ruta de
producto ni se ocultó un error de la aplicación.

## Contratos de domótica

Las pruebas Go focalizadas aprobaron el túnel HTTPS, secretos opacos,
aislamiento por empresa/estación, ID único, diagnóstico GPIO seguro,
programación, reinicio y la interfaz de varias Raspberry/equipos.

## Límite

P110-004 queda **parcial**. Permanecen pendientes acciones permitidas/denegadas
con identidades activas por rol, cuatro cajas y validación física supervisada de
GPIO; PCS no tiene hardware registrado para certificar relés reales.
