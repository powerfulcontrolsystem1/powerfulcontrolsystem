# P110-008 — Restore aislado del candidato `1a6dc4fa`

Fecha: 2026-08-12  
Ambiente: VPS, red, PostgreSQL, almacenamiento y puerto efímeros. Staging y
producción no recibieron escrituras.

## Ejecución

El runner restauró un snapshot en PostgreSQL temporal y aplicó el migrador por
digest del candidato `1a6dc4fa`. La API exacta se inició contra esa copia y
ClamAV se ejecutó por digest en la red aislada.

Resultado saneado observado:

```text
health=200 ready=200 bases=2 tablas=5 filas_empresa_12=32
archivos_privados=6 referencias_privadas=6 huerfanos_privados=0
runtime_privilegios=0 RTO=119s RPO=84351s
```

El migrador verificó que el rol de runtime no tenía privilegios DDL. Tras el
resultado se comprobó que no quedaban contenedores, red ni directorio temporal,
y staging continuó con `health` y `ready` correctos.

## Límite

No se ejecutó réplica autenticada ni rollback coordinado en este candidato, y
los objetivos RPO/RTO no tienen aceptación firmada. El valor observado de RPO
depende de la antigüedad del snapshot y no constituye un compromiso operacional.
P110-008 permanece parcial.
