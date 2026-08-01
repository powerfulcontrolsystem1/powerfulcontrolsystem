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
volúmenes, arranque de la aplicación y consulta multiempresa de cinco dominios
críticos. Todavía faltan:

1. completar el recorrido visual UI posterior al restore;
2. ejecutar subida en réplica A y descarga en réplica B;
3. ensayar pérdida de réplica y rollback coordinado de aplicación/base;
4. aprobar formalmente objetivos RPO/RTO del piloto.

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

## Arranque real del candidato sobre el restore - 2026-08-01

Se agregó `deploy/scripts/vps-p109-restored-app-drill.sh` y se ejecutó en la VPS
autorizada con las imágenes inmutables API y migrador del candidato
`89d6e042...`. El procedimiento:

- restauró el snapshot más reciente en PostgreSQL 16.14 efímero;
- creó un rol runtime temporal sin superusuario, `CREATEDB`, `CREATEROLE` ni
  `BYPASSRLS`;
- aplicó el migrador exacto con bootstrap legado desactivado: cinco migraciones
  empresariales y una administrativa;
- montó una copia temporal del almacenamiento privado y arrancó la API exacta;
- comprobó `/health` y `/ready` con HTTP 200;
- consultó las cinco tablas críticas con `empresa_id=12`, sumando 28 filas;
- rechazó sin sesión cuatro rutas empresariales con HTTP 401/403;
- autenticó por el login oficial la cuenta autorizada, sin guardar ni imprimir
  la clave, y recibió HTTP 200 en CxP, contabilidad, CxP/IA, diagnóstico DIAN y
  gestión documental;
- eliminó sesión, base, API, red y archivos temporales junto con el contenedor.

Resultado concluyente:

```text
[OK] app_restore_smoke health=200 ready=200 bases=2 tablas=5 filas_empresa_12=28 endpoints_protegidos=4 dominios_autenticados=5 runtime_privilegios=0 RTO=21s RPO=45418s
[OK] residual_containers=0 residual_networks=0 residual_tmp=0
```

Los intentos de desarrollo previos también fallaron de forma cerrada y limpiaron
sus recursos: detectaron stdin ausente al crear el rol, almacenamiento privado
de solo lectura y dos rutas HTTP obsoletas en la matriz. No se modificaron los
contenedores activos ni los datos de staging o producción.

### Inspección visual del restore

La API restaurada se expuso únicamente mediante un túnel SSH local y se abrió
con el navegador interno. El login oficial llevó al panel administrativo y se
recorrieron CxP/IA, suite contable, Finanzas, centro DIAN y gestión documental.

- Escritorio: cinco páginas sin desbordamiento horizontal de documento, cero
  errores de consola y datos/tablas visibles en filas y columnas.
- Móvil: cinco páginas sin desbordamiento horizontal de documento ni botones
  principales recortados. Las tablas anchas conservaron regiones desplazables.
- En DIAN, los botones `Reconsultar` quedan dentro de la tabla desplazable de
  403 px y no ensanchan la página. El snapshot restaurado mostró ambiente
  producción, estado rechazado y avance 50 %; no se llamó a DIAN ni se presenta
  ese estado histórico como aceptación fiscal.
- CxP/IA mostró un soporte extraído, proveedor, documento, fechas, total y
  confianza organizados en columnas, sin pérdida visual de los controles.

Al interrumpir la primera ventana SSH se comprobó que un `SIGHUP` abrupto podía
dejar recursos efímeros. Se eliminaron exactamente los dos contenedores, la red
y `/tmp/p109-restore-app-visual-89d6e042`, manteniendo el snapshot fuente. El
script se corrigió para atender `HUP`, `INT` y `TERM`; un ensayo deliberado con
`timeout -s TERM 40s` confirmó:

```text
signal_exit=124 residual_containers=0 residual_networks=0 residual_tmp=0
```

Esta evidencia completa el arranque y los recorridos API/UI del punto 5. La
siguiente sección añade el ensayo A/B sin presentarlo aún como cierre total.

### Réplicas A/B y pérdida de una réplica de aplicación

El mismo runbook arrancó dos contenedores API desde el digest exacto, conectados
a la base y copia de almacenamiento privado restauradas. Con dos sesiones
creadas por el login oficial:

1. la réplica A radicó por el endpoint oficial un PNG controlado para
   `empresa_id=12`, incluyendo token CSRF y origen same-origin;
2. la réplica B descargó el soporte y su SHA-256 coincidió con el archivo fuente;
3. se retiró por completo la réplica A;
4. la réplica B conservó `/ready=200` y volvió a descargar el mismo soporte con
   SHA-256 idéntico;
5. la base, archivo, sesiones y contenedores se destruyeron con el entorno
   efímero, sin crear datos en staging o producción.

La primera llamada de desarrollo omitió CSRF y fue rechazada con HTTP 403 sin
crear archivo ni fila; la repetición usó el contrato real y aprobó:

```text
[OK] app_restore_smoke health=200 ready=200 bases=2 tablas=5 filas_empresa_12=28 endpoints_protegidos=4 dominios_autenticados=5 replica_checks=2 runtime_privilegios=0 RTO=24s RPO=46310s
[OK] residual_containers=0 residual_networks=0 residual_tmp=0
```

Queda demostrada la conmutación entre réplicas de aplicación con volumen
compartido, no la pérdida de la capa de almacenamiento subyacente. P109-008
continúa parcial por cuotas/retención, inventario de heredados/huérfanos, fallo
del almacenamiento, rollback coordinado de datos y aprobación formal de RPO/RTO.

### Matriz hostil dinámica de archivos

Sobre el mismo entorno efímero se añadieron cinco negativos dinámicos:

- descargar el soporte de PCS usando `empresa_id=7` devolvió 403/404;
- un archivo `.html` con contenido activo devolvió HTTP 400;
- un archivo PNG declarado de 16 MiB, por encima del límite de 15 MiB, devolvió
  HTTP 400/413;
- el conteo de soportes de `empresa_id=12` permaneció idéntico después de los
  dos uploads rechazados;
- al sustituir únicamente el archivo efímero controlado por un symlink fuera de
  la carpeta empresarial, la descarga devolvió HTTP 404.

La repetición concluyó:

```text
[OK] app_restore_smoke health=200 ready=200 bases=2 tablas=5 filas_empresa_12=28 endpoints_protegidos=4 dominios_autenticados=5 replica_checks=2 archivos_hostiles=5 runtime_privilegios=0 RTO=25s RPO=46501s
[OK] residual_containers=0 residual_networks=0 residual_tmp=0
```

La matriz cubre identidad empresarial, contenido activo, tamaño y symlink. No
cubre todavía cuotas por empresa, retención/borrado/recuperación, antivirus ni
una segunda identidad A/B no global; esos puntos conservan el estado parcial.

### Pérdida y rollback coordinado - 2026-08-01

El runbook incorporó `P109_VERIFY_COORDINATED_ROLLBACK=1`, condicionado a dos
réplicas y al login QA oficial. Sobre el mismo candidato inmutable de staging:

1. radicó y descargó el soporte controlado, con SHA-256 coincidente;
2. retiró ambas APIs efímeras para congelar una frontera coherente;
3. creó dumps lógicos independientes de las dos bases, un tarball del volumen
   privado y un inventario SHA-256 de todos sus archivos;
4. eliminó deliberadamente solo las dos bases y el almacenamiento ubicados en
   el entorno temporal validado;
5. recreó/restauró ambas bases y el volumen, y comparó el inventario completo;
6. reinició una réplica, autenticó nuevamente por el login oficial y verificó
   la fila del soporte, su descarga, cinco tablas críticas y el hash original.

Resultado observado:

```text
[OK] app_restore_smoke health=200 ready=200 bases=2 tablas=5 filas_empresa_12=28 endpoints_protegidos=4 dominios_autenticados=5 replica_checks=2 archivos_hostiles=5 archivos_privados=2 referencias_privadas=2 huerfanos_privados=0 referencias_heredadas=0 rollback_checks=7 rollback_dominios=5 rollback_RTO=23s runtime_privilegios=0 RTO=47s RPO=51894s
```

Antes del arranque, el modo `P109_VERIFY_PRIVATE_INVENTORY=1` cruzó las
referencias persistidas de chat, buzón, DIAN, finanzas, grafología y soportes
CxP/IA contra el volumen. Encontró dos archivos y dos referencias exactas, sin
faltantes, huérfanos, referencias heredadas, symlinks ni rutas fuera del patrón
`<categoria>/empresa_<id>/<archivo>`. No hubo nada que migrar o eliminar en el
snapshot actual; el control queda reproducible para snapshots posteriores.

La limpieza posterior confirmó cero contenedores, redes y directorios
temporales del ensayo. Staging conservó su API por digest y producción mantuvo
su imagen previa, ambas en ejecución; no se promovió ni reinició ningún servicio
activo. Esto demuestra el rollback coordinado posterior a migración y la
recuperación del almacenamiento, pero P109-008 continúa **parcial** por cuota,
retención/borrado, antivirus, segunda identidad A/B no global y aprobación
formal del objetivo RPO/RTO.
