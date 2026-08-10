# P109-009 - CSP estricta Report-Only validada en staging

Fecha: 2026-08-08  
Ambiente: `staging`; empresa de prueba PCS; produccion excluida.

## Candidato y recuperacion

- Revision exacta: `a408bb621fdef1114c43fdb11ff0245dd3098af7`.
- Workflow inmutable: `31243348139`, aprobado con build, escaneo de cuatro
  imagenes, SBOM, publicacion por digest y validacion de Compose.
- `immutable_release_check.ps1`: PASS antes de promover.
- Se reutilizo el respaldo de PostgreSQL de staging creado inmediatamente antes
  del candidato estatico, con SHA-256 `72d29f6133be41028be7d6431083b71eb51575ad8e45c50840e19d8f5734de8f`.
- La promocion por digest activo dejo frontend `b268fc8a0081`, backend
  `cfcd54521592` y worker `bb17b8140cda`. `/ready` respondio `ready`.

## Cabeceras verificadas

- La CSP aplicada conserva `unsafe-inline` como politica de compatibilidad.
- La cabecera `Content-Security-Policy-Report-Only` no contiene
  `unsafe-inline` ni el permiso amplio `https:`.
- El backend de staging conserva `PCS_CSP_REPORT_ONLY_STRICT=1` mediante un
  override aislado de staging; no se editaron ni reiniciaron servicios de
  produccion.

## Revision visual autenticada

- Login de PCS en staging: PASS; llego al panel Super Administrador.
- La captura mostro menu, tarjetas, metricas y controles sin recortes ni
  roturas visuales.
- La consola repitio un unico `TypeError` de `MutationObserver.observe` sin
  traza expuesta por el navegador interno. El guard de `menu.js` ya estaba en
  el candidato y fue confirmado en el recurso servido; por tanto el origen del
  diagnostico no queda atribuido. Se mantiene como pendiente y no se cuenta
  como validacion limpia de consola.
- Las sesiones de prueba se cerraron por `/auth/logout` y las pestañas fueron
  finalizadas.

## Estado

La compuerta de cabeceras CSP Report-Only en staging queda aprobada. P109-009
continua parcial: falta atribuir/eliminar el diagnostico de consola y completar
DAST hostil, aislamiento A/B no global y la migracion de los scripts y estilos
embebidos antes de enforcement o produccion.
