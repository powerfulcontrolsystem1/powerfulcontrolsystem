# Reporte de ejecucion de staging

Estado: **NO EJECUTADO EN ESTA ESTACION**.

No se levantó Docker ni se contactó un VPS, coherente con la restriccion de no
desplegar ni modificar infraestructura real. La estacion no tiene Docker
disponible.

## Ejecucion requerida

1. Crear una red y volumenes con el prefijo `pcs_staging_` definido en
   `deploy/docker-compose.staging.yml`.
2. Cargar solamente una copia anonima o una base vacia de prueba.
3. Ejecutar `docker compose -f deploy/docker-compose.staging.yml config` y
   levantar el perfil aislado.
4. Aplicar migraciones dos veces, comprobar healthchecks y repetir la prueba
   con dos empresas, dos usuarios y roles distintos.
5. Adjuntar aqui IDs de ejecucion de CI, hash de imagen y resultado de cada
   smoke test. No registrar secretos, URLs privadas ni datos empresariales.

El resultado requerido es una evidencia reproducible, no una declaracion de
que staging fue probado sin haberlo sido.
