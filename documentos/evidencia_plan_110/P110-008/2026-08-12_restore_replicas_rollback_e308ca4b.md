# P110-008 — réplica autenticada y rollback coordinado de `e308ca4b`

Fecha: 2026-08-12  
Ambiente: VPS aislado; red, PostgreSQL, volumen privado y puertos loopback
efímeros. No se escribieron datos ni se reiniciaron servicios activos de
staging o producción.

## Ejecución verificable

El runner restauró el último snapshot en PostgreSQL efímero, aplicó el
migrador por digest y levantó dos réplicas API del mismo candidato, con ClamAV
por digest. La autenticación se realizó por el login oficial de PCS y las
credenciales solo existieron como variables de la sesión remota.

- Ambos nodos superaron `health` y `ready`; los cuatro endpoints protegidos
  rechazaron anónimo y cinco dominios empresariales respondieron autenticados.
- Una carga controlada en A se descargó en B con el mismo SHA-256. Tras retirar
  A, B conservó readiness y la descarga; se aprobaron dos comprobaciones de
  réplica.
- Una petición de otra empresa, HTML activo, archivo superior al límite y
  symlink fueron rechazados sin crear filas; se aprobaron cinco negativos de
  archivo.
- Se creó un checkpoint coherente de las dos bases y del volumen privado, se
  eliminó exclusivamente esa copia temporal y se restauró. El login, los cinco
  dominios, la fila controlada y su archivo SHA-256 se recuperaron; se
  aprobaron siete comprobaciones de rollback y cinco dominios posteriores.

Resultado saneado:

```text
health=200 ready=200 bases=2 tablas=5 filas_empresa_12=32
endpoints_protegidos=4 dominios_autenticados=5 replica_checks=2
archivos_hostiles=5 archivos_privados=6 referencias_privadas=6
huerfanos_privados=0 referencias_heredadas=0 rollback_checks=7
rollback_dominios=5 rollback_RTO=99s runtime_privilegios=0 RTO=226s
```

Al finalizar se comprobó que quedaban cero contenedores del drill. Los valores
observados de RTO/RPO describen este ensayo, no un compromiso firmado de
servicio.

## Límite

P110-008 sigue parcial hasta congelar el SHA final sin cambios posteriores y
obtener la aceptación operativa de RPO/RTO. Esta réplica no autoriza promoción
a producción ni sustituye el piloto, UAT, impresión física o la decisión GO.
