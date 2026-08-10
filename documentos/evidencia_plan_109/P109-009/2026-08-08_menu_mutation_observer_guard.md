# P109-009 - guardia para observador dinamico del menu

Fecha: 2026-08-08  
Alcance: candidato de staging Plan 109; sin cambio en produccion.

## Hallazgo visual autenticado

El panel Super Administrador de PCS se cargo correctamente tras autenticar en
staging y la captura no mostro fallo de interfaz. La consola, sin embargo,
registro un `TypeError` de `MutationObserver.observe` por recibir un destino
que no era un nodo. La condicion estaba protegida por `try/catch`, pero el
evento se conserva como error y no debe normalizarse para el cierre productivo.

## Correccion

El observador que sincroniza temas en `web/menu.js` valida que el destino exista
y sea un nodo elemento antes de invocar `observe`. El cambio no altera el menu,
las sesiones, datos empresariales ni la CSP; solo evita registrar el error en
ciclos tempranos de carga.

## Verificacion local

- `go test . -run 'TestFrontendStaticResourcesExist|TestNextcloudFramePolicyUsesExactOrigins' -count=1`: PASS.
- `git diff --check`: PASS.
- El ejecutable `node` no esta instalado en esta estacion, por lo que la
  sintaxis JavaScript queda cubierta por el workflow Linux inmutable antes de
  la siguiente promocion de staging.

## Siguiente compuerta

Construir el candidato inmutable, promoverlo solo a staging y repetir login,
captura y consola autenticados. La ausencia de error no reemplaza las demas
pruebas de roles, A/B, DIAN, UAT ni piloto.
