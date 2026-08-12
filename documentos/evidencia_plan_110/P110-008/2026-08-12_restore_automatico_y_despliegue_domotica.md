# P110-008 — restore automático y despliegue Domótica

Fecha: 2026-08-12 (America/Bogota)  
Empresa de validación visual: Powerful Control System (`empresa_id=12`).

## Despliegue verificable

Los cambios de acceso directo opcional a Domótica fueron integrados en `main`.
`rs` aprobó el preflight profesional, sincronizó el VPS, ejecutó migrador y
permisos, levantó backend, worker y frontend, y validó Nginx. PostgreSQL quedó
saludable y la ruta pública `/health` respondió `200` con estado `ok`.

En sesión autenticada de PCS, el check cambió `Venta directa` del carrito a la
vista consolidada de equipos sin recargar el panel. Al desactivarlo, el enlace
recuperó inmediatamente el carrito. El valor final quedó desactivado.

## Restore aislado

El primer intento contra el directorio padre falló cerrado porque no encontró
`postgres_all.sql.gz`. Se mejoró el runner para seleccionar el subdirectorio
completo más reciente y mostrar solo diagnósticos remotos saneados. La
repetición sin indicar un snapshot concreto produjo:

```text
health=200 ready=200 bases=2 tablas=5 filas_empresa_12=32
endpoints_protegidos=4 archivos_privados=6 referencias_privadas=6
huerfanos_privados=0 referencias_heredadas=0 runtime_privilegios=0
RTO=78s RPO=59854s
```

PostgreSQL, red, puertos y archivos fueron efímeros. El runner informó limpieza
automática; no se conectó a la base activa ni se modificaron volúmenes activos.

## Estado

P110-008 permanece **parcial**. Esta evidencia no incluye autenticación en dos
réplicas, pérdida del nodo A, rollback coordinado del próximo digest final ni
aceptación humana de RPO/RTO. El veredicto global continúa **NO-GO**.
