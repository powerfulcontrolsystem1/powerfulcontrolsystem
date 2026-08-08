# P109-008 - Restore, replicas y rollback coordinado de `c8094f5b`

Fecha: 2026-08-08 01:50 America/Bogota  
Entorno: VPS de staging, Docker efimero y aislado  
Alcance: snapshot `20260808_031501`, candidato exacto
`c8094f5be638bbd6262e12e191d365793ed92f6b`. Staging activo, PCS y produccion
no fueron escritos ni reiniciados.

## Controles aprobados

Se ejecuto `vps-p109-restored-app-drill.sh` desde copia temporal verificada por
SHA-256 y con limpieza automatica:

- restore de las dos bases y volumen privado, migrador/API exactos y
  health/readiness 200;
- 5 tablas y 28 filas criticas de PCS, 4 endpoints anonimos protegidos y 5
  dominios autenticados despues del restore;
- inventario privado: 2 archivos y 2 referencias, sin huerfanos, referencias
  heredadas ni symlinks;
- dos replicas sobre el mismo volumen: carga por A, descarga por B con
  SHA-256 igual y continuidad de B tras retirar A (`replica_checks=2`);
- cinco negativos: cruce de empresa, contenido activo, archivo sobredimensionado,
  ausencia de fila tras rechazo y escape por symlink (`archivos_hostiles=5`);
- rollback coordinado de bases, archivo y sesiones: 7 controles, 5 dominios,
  archivo y fila recuperados (`rollback_RTO=26 s`);
- rol runtime sin privilegios DDL, RTO total 52 s y RPO observado 12.980 s
  (3 h 36 min 20 s), dentro de los objetivos publicados de 2 h/24 h.

La comprobacion posterior encontro cero contenedores, redes, volumenes o
scripts temporales del ensayo.

## Limites

P109-008 conserva estado **parcial**: faltan cuota empresarial, retencion y
borrado/recuperacion, antivirus y aislamiento con una segunda identidad no
global. Por ello esta evidencia no autoriza promocion ni aumenta el porcentaje
del Plan 109.
