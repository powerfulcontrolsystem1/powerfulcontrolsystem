# P109-008 - Cuota empresarial para soportes CxP/IA

Fecha: 2026-08-08 America/Bogota
Alcance: candidato inmutable `516de42e` promovido solo a staging y una
restauracion efimera separada. No hubo mutacion de PCS activa ni de produccion.

## Cambio implementado

`/api/empresa/soportes_compras_ia?action=radicar` conserva el wrapper de
permisos del modulo y el `empresa_id` validado. Antes de guardar un adjunto
privado ahora:

1. suma el uso de `/uploads/empresas/<empresa>` y de
   `private_storage/soportes_compras_ia/empresa_<id>`;
2. aplica la configuracion global existente `empresa_storage` de limite,
   maximo por archivo, bloqueo y activacion;
3. rechaza con HTTP 507 y mensaje publico saneado si la carga excede la cuota;
4. mantiene el maximo estricto de soportes de 15 MiB y usa el menor limite si
   la configuracion corporativa define uno inferior.

El cliente no controla ruta, empresa, limite ni contador. No se cambiaron
tablas, rutas, permisos, dependencias ni archivos existentes.

## Pruebas locales

- cuota superada bloquea el adjunto de la misma empresa;
- llegar exactamente al limite permanece permitido, coherente con el adjunto
  de buzon;
- switches de cuota y bloqueo se respetan;
- maximo por archivo se aplica;
- los bytes privados de empresa 12 no incluyen los de empresa 7;
- `go test ./... -count=1`, `go vet ./...` y `git diff --check`: PASS.

## Validacion HTTP aislada del candidato

- El workflow inmutable `31245235398` construyo, escaneo, genero SBOM y
  publico los cuatro digests de `516de42e`; staging recibio esos digests sin
  reconstruccion y mantuvo `/health` y `/ready` en HTTP 200.
- Para no alterar la configuracion global de staging, se restauro el snapshot
  vigente en PostgreSQL/contenedores efimeros y se inicio la API exacta del
  candidato en un puerto loopback aislado. La restauracion, migrador y API
  aprobaron antes del ensayo y el entorno se autolimpio al finalizar.
- Solo en esa copia se fijo una cuota de 1 MiB para la empresa 12. Una sesion
  oficial autorizada intento radicar un soporte PNG de 2 MiB en
  `soportes_compras_ia`. La respuesta fue HTTP 507 con mensaje saneado y la
  consulta posterior confirmo `0` filas con el numero de documento de prueba.
- La configuracion activa de staging no fue escrita: el endpoint administrativo
  protegido devolvio 403 antes de cualquier mutacion, por lo que se uso la
  copia restaurada para una prueba equivalente y reversible.
- Tras el hold controlado no quedaron contenedor, red, directorio temporal,
  script ni log del ensayo en el VPS.

## Limite

La subcompuerta de rechazo HTTP de cuota queda demostrada en el candidato
inmutable restaurado. La siguiente seccion registra la prueba de carrera que
cerro ese hueco; aun faltan retencion, borrado/recuperacion, antivirus y A/B no
global antes de considerar cerrada P109-008 o emitir un GO.

## Carrera entre replicas del candidato corregido

- El candidato posterior `44610128` incorpora un candado asesor de PostgreSQL
  por empresa alrededor de la verificacion de cuota, la escritura privada y el
  registro. El candado es de sesion y se libera en `defer`, por lo que dos APIs
  comparten la misma exclusion mutua sin depender de memoria local.
- El workflow inmutable `31245956681` aprobo build, escaneo, SBOM y Compose.
  Tras un respaldo, sus digests exactos se promovieron exclusivamente a staging;
  `/health` y `/ready` siguieron en HTTP 200.
- Sobre una nueva restauracion efimera se arrancaron dos replicas API exactas
  con el mismo PostgreSQL y volumen privado. Con cuota de 1 MiB solo en la
  copia, dos sesiones oficiales enviaron en paralelo soportes de 700 KiB.
  Una respuesta fue HTTP 200 con una fila y la otra HTTP 507 con cero filas.
- La segunda replica, ambas sesiones, PostgreSQL, red, volumen temporal, script
  y log fueron eliminados al terminar. Ninguna cuota de staging, empresa PCS
  activa ni produccion fue modificada.

## Pendientes que impiden cierre

P109-008 conserva estado parcial por retencion, borrado/recuperacion,
antivirus y una prueba A/B no global. La cuota ya cuenta con evidencia HTTP y
de concurrencia entre replicas, pero eso no equivale a certificacion productiva.
