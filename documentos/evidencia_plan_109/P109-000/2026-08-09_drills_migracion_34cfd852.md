# P109-000 - drills exactos de migración de 34cfd852

Fecha: 2026-08-09

Ambiente: recursos efímeros aislados en el VPS y snapshot lógico de staging.

Producción: sin despliegue ni mutaciones.

## Candidato

- SHA funcional: `34cfd8526905f2a495b6794787411c66957f99da`.
- Migrador inmutable: `ghcr.io/powerfulcontrolsystem1/pcs-migrate@sha256:7dc69a3c558dc5c1b2ea35c288e8d4072ffe968cd91052194d89289c050ffa4c`.
- Los ensayos usaron el digest, no una etiqueta mutable ni un build local.

## Base vacía

El script `deploy/scripts/vps-p108-empty-migration-drill.sh` creó bases, roles,
red y volumen efímeros, ejecutó el migrador dos veces y probó además un drift
controlado de checksum.

| Control | Resultado |
| --- | ---: |
| Tablas empresariales | 337 |
| Tablas globales | 49 |
| Ledger empresarial | 20 |
| Ledger global | 10 |
| Segunda pasada | 0 migraciones nuevas |
| Drift detectado con fallo cerrado | PASS |
| Esquema sin cambios durante el drift | 337 |
| Ledger sin cambios durante el drift | 20 |
| Recuperación después del drift | PASS |

Resultado final del script: `OK`.

## Upgrade de snapshot

El script `deploy/scripts/vps-p108-upgrade-migration-drill.sh` obtuvo una copia
lógica consistente de las dos bases de staging, la restauró en PostgreSQL
aislado y ejecutó dos veces el mismo digest con bootstrap runtime desactivado.

| Control | Antes | Después |
| --- | ---: | ---: |
| Tablas empresariales | 350 | 350 |
| Tablas globales | 59 | 59 |

Ambas pasadas informaron cero migraciones nuevas y verificaron el rol runtime
sin privilegios DDL. Resultado final del script: `OK`.

## Limpieza y regresión posterior

- Contenedores temporales restantes: 0.
- Redes temporales restantes: 0.
- Volúmenes temporales restantes: 0.
- Staging: `/health=ok`, `/ready=ready`; backend, worker y frontend saludables.
- Digests efectivos de API, worker y frontend coinciden con el release
  `31303668393`.
- PCS `empresa_id=12`: CxP 3, pagos 5, movimientos 5; sin cambios.
- Producción: `https://powerfulcontrolsystem.com/health=ok`, sin promoción.

## Estado de la compuerta

Los dos drills técnicos pendientes de P109-000 quedan aprobados para el SHA
exacto. P109-000 continúa formalmente parcial porque el candidato permanece en
rama y el propietario pidió no crear PR: falta revisión/fusión a `main`. Esta
limitación de gobernanza impide otorgar crédito binario de certificación y no
autoriza producción.
