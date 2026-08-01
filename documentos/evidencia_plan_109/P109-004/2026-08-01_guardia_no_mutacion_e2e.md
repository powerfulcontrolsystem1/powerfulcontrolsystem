# P109-004 - Guardia de no mutacion del barrido E2E

Fecha: 2026-08-01  
Entorno: staging  
Runtime probado: `5f1b0692a7b03207e0357298421efa185db2b48d`

El barrido completo `30689223529` se cancelo al observar un `PUT 200` en
`/super/api/licencias/vencimiento_alertas`. La auditoria global confirmo que
ejecuciones antiguas de HeadlessChrome habian escrito configuracion. La causa
fue doble: cualquier `aria-label` o `title` podia convertir un control en clic
seguro y, tras recargar la pagina, se reutilizaba un indice DOM que podia
apuntar a otro boton.

La correccion aplica tres fronteras:

- solo ejecuta controles con texto o dataset expresamente permitido;
- usa `id`, `data-target` o `href` unicos y estables, nunca indices despues de
  recargar;
- bloquea en red `POST`, `PUT`, `PATCH` y `DELETE`, deshabilita service workers
  en el contexto de auditoria y registra cualquier intento como hallazgo.

La primera prueba protegida (`30689762462`) bloqueo diez mutaciones y demostro
la colision exacta sin crear nuevas filas de auditoria. Tras reparar el selector,
la repeticion `30689926040` aprobo las dos vistas de Configuracion avanzada:
272 controles, cinco clics de navegacion, cero mutaciones, cero hallazgos y cero
errores. El conteo de mutaciones HeadlessChrome en PostgreSQL permanecio
invariable antes y despues (`101`, maximo ID `15519`).

El barrido global protegido `30689993317` recorrio 618 vistas/309 rutas y
11.062 controles sin aumentar ese conteo. Registro 596 vistas `ok`, 22 en
revision, 103 clics de identidad estable, 2.081 acciones riesgosas omitidas y
12 POST bloqueados. Estos POST eran la consulta de renta IA y el contador de
visitas del portal; no llegaron al servidor. Los unicos HTTP de aplicacion
observados fueron cuatro 403 por permisos cerrados y cuatro 404 de dos imagenes
historicas de Red Social. No hubo `pageerror`.

Los logs correlacionados encontraron un 502 del explorador VPS2 al no existir
servidor ni snapshot disponible. La rama de cierre lo representa como estado
degradado HTTP 200 con `ok=false`. La guardia tambien deja de contar como error
las cancelaciones `ERR_ABORTED` causadas por sus propias recargas y los avisos
esperados de service worker/mutacion bloqueada.

Estado: **P109-004 parcial**. Falta repetir el barrido sobre el digest que
incluya VPS2 degradado y probar las acciones mutantes por flujos oficiales.
