# P109-010 - Recuperacion de capacidad por imagenes antiguas

Fecha: 2026-08-01  
Entorno: VPS principal

El disco alcanzo 81 % despues de publicar el candidato final. Docker acumulaba
137 imagenes y 59,29 GB. Se inventariaron contenedores e IDs antes de limpiar.

Se eliminaron exclusivamente 104 referencias antiguas de las imagenes GHCR
`pcs-api`, `pcs-migrate`, `pcs-worker` y `pcs-frontend` que no pertenecian a un
contenedor. Se conservaron por digest las cuatro imagenes activas del candidato
`5f1b0692...`, las cuatro del rollback inmediato `6da6c134...`, todas las
imagenes locales de produccion y toda imagen de infraestructura en ejecucion.

Resultado:

- 41.556.353.024 bytes recuperados;
- uso raiz de 81 % a 40 %, con 58 GB disponibles;
- 33 imagenes restantes, 25 activas;
- produccion y staging saludables; bases, volumenes, backups y archivos no se
  tocaron.

Estado: evidencia adicional de **P109-010 parcial**. Faltan simulacro de alertas,
deduplicacion, recepcion y escalamiento.
