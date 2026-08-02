# P109-004 - Barrido global 99348ff4 y cierre dirigido edd2bdac

Fecha: 2026-08-01  
Entorno: staging aislado  
Empresa: Powerful Control System (`empresa_id=12`)  
Produccion: no modificada

## Barrido global protegido

El candidato `99348ff4e44283d461723afe87800e8d58bb2700` recorrio 618
vistas (309 rutas en escritorio y movil) mediante login real y guardia de no
mutacion:

- 602 vistas `ok` y 16 `review`;
- 11.102 controles inventariados, 97 clics seguros y 2.097 acciones riesgosas
  preservadas para su flujo oficial;
- 12 POST automaticos bloqueados por la guardia: renta IA y telemetria publica;
- cero excepciones de pagina, dialogos, bloqueos de seguridad y HTTP 5xx;
- 23 abortos de navegacion esperados al recargar despues de clics seguros;
- dos HTTP 403 del reporte de aseo y cuatro HTTP 404 correspondientes a dos
  imagenes antiguas repetidas en escritorio y movil.

Los unicos defectos visibles reproducibles fueron el reporte de aseo negado a
un administrador valido y dos referencias locales de publicaciones cuyo
archivo ya no existia.

## Correccion

El SHA `edd2bdac64f760d18289249d6c2f4720aba6f925`:

- propaga al handler el rol administrativo resuelto en servidor dentro del
  middleware de autoservicio; no confia en `X-Admin-Role` del cliente ni cambia
  el alcance validado por `empresa_id`;
- oculta en la respuesta las fotos locales de red social que no existen en el
  volumen, conserva URLs externas y el dato historico, y rechaza traversal;
- agrega regresiones para archivos presentes, ausentes, externos y traversal,
  y para los roles autorizados/no autorizados del reporte de aseo.

`go test ./...`, `go test ./handlers ./db` y `go vet ./handlers ./db`
terminaron correctamente.

## Candidato inmutable y staging

El workflow `30732433840` aprobo build, Trivy HIGH/CRITICAL, cuatro SBOM,
publicacion y validacion Compose. Los digests promovidos sin recompilar fueron:

- API: `sha256:99477e71beb99eb98620dad6785b113154edd7c85e5e84006f6323af1f079c9a`;
- migrador: `sha256:a190b4b8e24af9931d2db82b9b7f237a3697cc3add1ea65a15add0a3a1bf3c2a`;
- worker: `sha256:e9d515e19701410033dfa5914fab4fea2d669075748fe4ff7b3c85b804827daf`;
- frontend: `sha256:67cedcafd2aae5b35e83c4cc5be7d3031e886cd7a84f3e3849bab2f24b8c6126`.

Antes de la promocion se respaldaron las dos bases de staging y se calcularon
sus SHA-256. `/health` y `/ready` quedaron en 200.

La repeticion dirigida obtuvo 4/4 vistas `ok`, 168 controles, cero respuestas
HTTP fallidas, errores de consola, excepciones o mutaciones bloqueadas. La
revision visual en 1366x900 y 390x844 confirmo filas, columnas, tarjetas,
controles y fallback de publicaciones sin recortes ni imagen rota. El boton
real `Consultar` produjo HTTP 200 y el estado visible `Reporte actualizado.`.
La API publica ya entrega cero referencias a los dos archivos ausentes.

La inconsistencia entre `empresa_id=12` y `X-Empresa-ID=7` fue rechazada con
HTTP 400. La identidad PCS usada es global y puede acceder a otras empresas por
diseño; por tanto esta prueba no se cuenta como aislamiento A/B. Sigue siendo
obligatoria una segunda identidad no global.

Staging conserva 20 sesiones vigentes, maximo 20 por identidad y cero
identidades excedidas. Produccion conservo sus imagenes locales y respondio
`/health` y `/ready` en 200. Se retiraron 40 imagenes de candidatos GHCR no
usadas, conservando el candidato actual y el rollback inmediato; el disco bajo
de 73 % a 58 %. Los logs posteriores de backend, worker y frontend registraron
cero coincidencias de panic, fatal o HTTP 5xx.

Estado: **P109-004 parcial**. El inventario no mutante y estos defectos quedan
cerrados sobre candidato inmutable, pero faltan acciones riesgosas por flujo
oficial, roles, segunda identidad A/B y firma del alcance. No cambia el
porcentaje general: **40,0 % de implementacion**, **6,7 % de certificacion** y
**NO-GO**.
