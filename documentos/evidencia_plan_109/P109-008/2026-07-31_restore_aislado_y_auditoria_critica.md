# P109-008 - Restore aislado y auditoría crítica

Fecha: 2026-07-31  
Entorno: VPS autorizada, contenedores PostgreSQL efímeros y aislados.  
Alcance: snapshot más reciente disponible; producción no fue destino del
restore y no se modificaron contenedores ni datos de la aplicación.

## Compuerta previa

- Se usó el procedimiento documentado
  `scripts/vps_restore_validation.ps1` con la confirmación remota explícita.
- El snapshot seleccionado fue el más reciente bajo el repositorio de backups
  de la VPS.
- El restore se ejecutó en un contenedor temporal con PostgreSQL 16.14.
- El contenedor tuvo limpieza automática mediante `trap`, incluso ante error.

## Integridad de artefactos

El dump global PostgreSQL superó `gzip -t`. También abrieron correctamente los
tarballs obligatorios de:

- uploads y downloads;
- logs del backend y backups internos;
- almacenamiento privado;
- certificados Mailu;
- datos, librerías y logs de OnlyOffice.

Los tarballs de Let's Encrypt y `certbot_www` no están configurados en este
snapshot y el runbook los clasifica como opcionales; esto no se interpreta como
evidencia de recuperación TLS.

## Restore y verificación funcional de datos

El primer ensayo restauró el dump completo en un PostgreSQL temporal y confirmó
conectividad SQL. Un segundo ensayo independiente repitió el restore y exigió:

- presencia de las dos bases esperadas;
- consulta de `empresa_cuentas_por_pagar`;
- consulta de `empresa_asientos_contables`;
- consulta de `empresa_ai_memoria`;
- consulta de `empresa_dian_configuracion`;
- consulta de `empresa_documentos_gestion`;
- ejecución correcta del filtro `empresa_id = 12` en las cinco tablas.

Resultado del segundo ensayo:

```text
[OK] Bases restauradas: 2/2
[OK] Tablas criticas consultables: 5/5
[OK] Filtros empresa_id=12 ejecutables: 5/5
[OK] Restore audit RTO=16s RPO=57011s
```

El ensayo documentado previo terminó con RTO de 23 segundos y RPO de 56.918
segundos. La diferencia corresponde al momento de ejecución; el RPO observado
es aproximadamente 15,8 horas y debe compararse con el objetivo que se firme
para el piloto.

## Limpieza y límites

Después de ambos ensayos se verificó que no quedaran contenedores
`pcs-restore-drill-*` ni `pcs-restore-audit-*`. No se imprimieron filas,
credenciales ni secretos.

P109-008 queda **parcial**. Se demostró restaurabilidad del dump, apertura de
volúmenes y consulta multiempresa de cinco dominios críticos. Todavía faltan:

1. levantar la aplicación contra el conjunto restaurado y recorrer CxP,
   contabilidad, IA, DIAN y documentos por API/UI;
2. comprobar archivos y checksums cruzados entre metadatos y volumen privado;
3. ejecutar subida en réplica A y descarga en réplica B;
4. ensayar pérdida de réplica y rollback coordinado de aplicación/base;
5. aprobar formalmente objetivos RPO/RTO del piloto.

El procedimiento quedó incorporado al runbook mediante
`-ExecuteDrill -VerifyCriticalData`: además del restore exige las dos bases,
las cinco tablas, sus filtros empresariales y la correspondencia entre cada
soporte IA privado persistido, su empresa, el miembro del tarball y su SHA-256.
Si cualquiera de esas invariantes falla, el ensayo termina de forma cerrada.

La ejecución reproducible del runbook nuevo aprobó con:

```text
[OK] Restore critico: bases=2 tablas=5 filtros_empresa=5 archivos_privados=2 checksums_soportes_ia=0
[OK] Restauracion temporal PostgreSQL completada. imagen=postgres:16.14-alpine RTO=24s RPO=57722s
```

El snapshot contiene dos archivos privados, pero ninguna fila vigente de
soportes IA con referencia `private://soportes_compras_ia/`; por ello el valor
`checksums_soportes_ia=0` es correcto y no se presenta como una conciliación de
soportes inexistentes. La compuerta queda preparada para exigir automáticamente
el hash cuando el flujo real cree esos soportes. La limpieza posterior volvió a
confirmar cero contenedores temporales.
