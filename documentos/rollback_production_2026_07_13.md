# Procedimiento de rollback para futura produccion - 2026-07-13

Este documento no autoriza un despliegue. Es la guia que debe validarse primero
en staging.

1. Detener la promocion y conservar el identificador de release, hora, logs
   saneados y conteos agregados. No copiar secretos ni payloads completos.
2. Confirmar que existe backup verificable previo al cambio y que su manifiesto
   no contiene credenciales. No restaurar sobre el destino equivocado.
3. Cambiar la aplicacion al artefacto anterior aprobado por CI, manteniendo la
   misma configuracion externa de secretos.
4. Si hay migracion reversible, ejecutarla solo con su procedimiento probado.
   Si no lo es, restaurar la base efimera o productiva exclusivamente mediante
   el runbook autorizado y un operador responsable.
5. Verificar salud HTTP, sesiones, aislamiento empresarial, almacenamiento
   privado, colas y conteos de documentos antes de reabrir trafico.
6. Registrar incidente sin PII, abrir analisis de causa y bloquear una nueva
   promocion hasta contar con prueba de regresion.

## Criterio de exito

El rollback solo se considera terminado si las pruebas de salud y autorizacion
no muestran acceso cruzado entre empresas y el inventario de datos coincide con
el backup o el punto de recuperacion aprobado.
