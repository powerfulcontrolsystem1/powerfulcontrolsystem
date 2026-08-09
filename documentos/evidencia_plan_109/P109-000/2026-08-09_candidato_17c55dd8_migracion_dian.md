# P109-000/P109-005/P109-007 - candidato `17c55dd8` en staging

Fecha: 2026-08-09.  
Ambiente modificado: staging aislado.  
Empresa de prueba autenticada: Powerful Control System, `empresa_id=12`.

## Candidato inmutable

- SHA: `17c55dd86030c47dc9e40d9abc99d2447b9091d9`.
- Workflow `Immutable release candidate`: ejecución `31301293281`, aprobada.
- El artefacto entregó cuatro referencias `repositorio@sha256` y el manifiesto
  tuvo SHA-256
  `00639fda0e99d2b773901439e6a7d7e7ddd1907e46957fff125ba96b7ace5fe5`.
- `immutable_release_check.ps1` aprobó y staging promovió esas mismas cuatro
  imágenes sin reconstruir.
- Respaldo previo verificado:
  `/root/pcs-staging-backups/p109-pre-dian-migration-20260809T073448Z`,
  79.806.669 bytes entre ambas bases.

## Migración DIAN

- `pcs-staging-migrate` terminó con código `0`.
- El ledger registró
  `20260809-001-dian-local-production-flag-v1|applied` con checksum
  `5d15ad4766d3593c04d2b4140b56894cb8a558954971e7c35272963ef55035c5`,
  igual al catálogo fuente.
- La columna pública quedó `integer|NO|0` y el constraint permite únicamente
  `0/1`. Staging no tenía filas DIAN preexistentes, por lo que no se inventó
  una activación.
- Un ensayo PostgreSQL transaccional y reversible reprodujo cinco casos. El
  backfill produjo `1,1,0,0,0`: conservó las dos activaciones con evidencia y
  dejó inactivas producción sin traza, habilitación y valores vacíos. El
  `ROLLBACK` dejó cero filas públicas.

## Salud, aislamiento y seguridad

- Staging respondió `health=ok` y `ready=ready`.
- Producción respondió `health=ok`; la huella combinada de API, worker y
  frontend fue idéntica antes y después:
  `2ff0fde9651c94047059858e0a86d9522171c653c238022b9cb1cccdc36a063c`.
- CSP enforced y report-only incluyen el origen exacto
  `https://mail.powerfulcontrolsystem.com`; no se agregó comodín.
- El manifiesto temporal se retiró del VPS.
- El segundo juego de imágenes elevó temporalmente el disco raíz a 80 %. Se
  verificó que los cuatro digests del primer candidato invalidado no estaban
  referenciados por ningún contenedor y se eliminaron de forma explícita. No se
  podaron volúmenes, bases, respaldos ni imágenes activas; el disco bajó a 78 %
  con 22.283.648 KiB libres y ambas saludes permanecieron `ok`. Los digests
  eliminados siguen siendo recuperables desde GHCR.

## Pruebas autenticadas y visuales

- El navegador interno abrió la sesión autorizada y `empresa_id=12`. Panel,
  centro DIAN y correo corporativo cargaron sin el error anterior del buzón.
- El centro DIAN mostró explícitamente `Produccion local sin activar` y cero
  historial, coherente con la base aislada; no se emitió ni anuló un documento
  fiscal sin configuración de staging.
- El barrido reproducible recorrió 48 variantes de 24 rutas críticas en
  escritorio y navegador móvil: 46 `ok`, 2 `review`, 1.226 botones detectados,
  12 clics seguros y 363 acciones riesgosas omitidas. Centro IA, correo y las
  dos páginas DIAN aprobaron; el único `review` fue la ruta heredada inexistente
  `/administrar_empresa/inventario.html` en ambas dimensiones.
- La revisión visual móvil mínima de la web confirmó tarjetas, textos y
  controles DIAN en una sola columna, sin el desbordamiento previo. La app
  móvil nativa permanece fuera del alcance de salida a producción por decisión
  del propietario; no se eliminó su código.
- `qa_print_formats.cjs` con Chrome aprobó 20/20 formatos, cero casos a revisar
  y cero fallos de impresión. Carta, POS, QR y el documento de 96 filas fueron
  inspeccionados visualmente con columnas, totales y cierre sin recortes.

## Límites y estado

- La cuenta de correo copiada en staging está `pendiente_provision`; por eso el
  fix PATCH/POST de Mailu no puede certificar IMAP ni autologin en este entorno.
- Staging no contiene configuración/certificado DIAN de PCS. Una factura o
  anulación fiscal real sigue pendiente de promoción controlada y validación de
  la configuración productiva; este ensayo no autoriza desplegar producción.
- P109-000 queda aprobado para este digest. P109-005 y P109-007 continúan
  parciales. El Plan 109 conserva **53,3 % de implementación**, alcanza **6,7 %
  de certificación del candidato `17c55dd8`** y mantiene veredicto **NO-GO**.
