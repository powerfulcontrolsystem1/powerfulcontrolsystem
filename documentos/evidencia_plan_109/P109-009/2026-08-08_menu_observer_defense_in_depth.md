# P109-009 - defensa completa de observadores del menu

Fecha: 2026-08-08  
Alcance: candidato de staging Plan 109; produccion excluida.

## Motivo

Tras validar la primera guardia de `MutationObserver`, el navegador interno
siguio informando un error sin traza durante la carga autenticada. Antes de
atribuirlo al navegador se revisaron los tres observadores que el menu registra
en el panel administrativo.

## Cambio

- Los observadores de campana y contador solo observan elementos DOM reales.
- El observador de iconos valida `document.body`/`documentElement` antes de
  invocar `observe`.
- La prueba estatica exige las tres guardias para impedir regresiones.

## Verificacion local

- `go test . -run 'TestFrontendStaticResourcesExist|TestNextcloudFramePolicyUsesExactOrigins|TestMenuThemeObserverGuardsItsTarget' -count=1`: PASS.
- `go test ./utils -count=1`: PASS.
- `go vet ./utils`: PASS.
- `git diff --check`: PASS.

La sintaxis JavaScript se validara en el workflow Linux inmutable antes de
promover el candidato exclusivamente a staging. Este cambio no crea usuarios,
ventas, documentos, CxP ni movimientos empresariales.
