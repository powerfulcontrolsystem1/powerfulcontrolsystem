# Runbook de despliegue de produccion

Este runbook no autoriza ejecutar despliegues desde esta rama.

1. Exigir todos los gates verdes de `documentos/release_checklist.md`.
2. Fijar commit, imagenes por digest y variables de entorno fuera del repositorio.
3. Realizar backup verificable y prueba de restauracion vigente.
4. Desplegar primero a staging, ejecutar smoke tests y aprobar formalmente.
5. Programar ventana, responsables y criterio de rollback antes de cambiar
   produccion.
6. Desplegar de forma gradual, observar metricas y detener ante errores de
   seguridad, aislamiento, pagos o facturacion.
