# Estado actual del sistema

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-07.

## Snapshot operativo

La rama `main` representa una única línea vigente del sistema. Código, contratos,
migraciones, pruebas, configuración no secreta y runbooks del repositorio forman
el snapshot actual. Planes cerrados, changelogs acumulativos, capturas, informes y
resultados de candidatos anteriores se conservan fuera de Git.

El snapshot contiene backend Go, PostgreSQL, frontend estático, despliegue y los
módulos descritos en el [mapa](mapa_modulos.md). La presencia de una capacidad en
código no acredita por sí sola migración aplicada, proveedor, hardware, staging o
producción.

## Condiciones vigentes de aceptación

La sala personal de [Juegos y Accesorios](juegos_accesorios.md) incorpora seis
juegos adaptables, guardado por identidad de sesión y récords recreativos
compartidos entre empresas. Su aceptación requiere migración, aislamiento de
partidas, recuperación de guardados y controles visuales en escritorio y móvil;
GTAS usa Three.js autorizado y modelos CC0 con una reconstrucción compacta de Springfield. Se abre en una ventana flotante y sustituye DOON; la exploración es libre, sin ruta guiada. No se presenta como una réplica exacta ni como fotorealismo.

| Área | Criterio actual |
| --- | --- |
| Multiempresa | Toda lectura y mutación empresarial valida `empresa_id`, usuario, licencia, permiso y ownership de IDs secundarios en backend |
| Datos | `pcs-migrate` es el único dueño de DDL; PostgreSQL es el único runtime permitido; migraciones aplicadas no se reescriben |
| Pagos y licencias | Firma, ambiente, importe, moneda, idempotencia y callback se validan en backend; una operación real exige autorización separada |
| Facturación electrónica | Credenciales, NIT, numeración, firma y trazabilidad son por empresa; la aceptación debe venir del proveedor/autoridad para la familia documental concreta |
| Seguridad | Secretos no se versionan; autenticación, proxy, headers, rate limit y permisos se prueban en el entorno que se pretende liberar. PCS no exige ni expone 2FA administrativo |
| Operación | Healthchecks, migración, backup/restore, observabilidad y rollback se verifican sobre el candidato inmutable |
| UI y hardware | Una validación estática no sustituye sesión autenticada, navegador real, impresora, Raspberry ni confirmación del dispositivo |

## Límites abiertos

- Las integraciones fiscales, pagos, correo, IA y hardware requieren comprobación
  externa por candidato; no se heredan resultados de una ejecución anterior.
- Capacidad, recuperación y seguridad del VPS se vuelven a medir para cada
  release. El repositorio no contiene inventarios privados ni secretos del host.
- Las brechas transversales se mantienen en el [registro vigente](gobernanza_tecnica/riesgos_y_brechas.md)
  y los requisitos en la [matriz](requisitos/especificacion_y_trazabilidad.md).

## Registro de una entrega

La PR debe identificar requisito, SHA/digests, entorno, comandos, resultado,
omisiones, efectos externos y riesgos. La evidencia pesada o sensible se guarda
en CI o almacenamiento operativo con retención definida; Git recibe solo la
referencia minimizada necesaria para reproducir o auditar el resultado.
