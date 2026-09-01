# P109-008 - Restore de aplicación del candidato activo de staging

Fecha: 2026-08-09
Ambiente: VPS de staging, Docker efímero y aislado
Alcance: API y migrador por digest inmutable activo; staging y producción no se
reiniciaron ni recibieron escrituras.

## Ejecución

El ejecutor `vps_p109_restore_app_validation.ps1` resolvió los digests activos,
recuperó únicamente las imágenes faltantes por digest y ejecutó
`vps-p109-restored-app-drill.sh` contra el último snapshot en PostgreSQL y red
temporales.

## Resultado

- restauración del snapshot y migración exacta: PASS;
- API restaurada: `/health` y `/ready` HTTP 200;
- dos bases PCS y cinco tablas críticas verificadas;
- 31 filas críticas de `empresa_id=12` comprobadas en la copia restaurada;
- cuatro endpoints anónimos protegidos rechazados;
- cinco archivos privados, cinco referencias, cero huérfanos y cero referencias
  heredadas;
- rol runtime sin privilegios DDL;
- RTO observado 24 s y RPO observado 54.834 s;
- los contenedores, red y script temporal quedaron programados para limpieza
  automática por el drill.

## Límites

Esta corrida no recibió credenciales de prueba dentro del entorno restaurado;
por ello no ejecutó réplica autenticada, carga A/B, negativos de archivo ni
rollback coordinado. P109-008 permanece **parcial** hasta completar esos
criterios con el mismo candidato.
