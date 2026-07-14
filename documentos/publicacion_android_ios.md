# Publicacion Android e iPhone

Actualizacion: 2026-07-14.

## Android

Ejecutar en una maquina con Flutter estable:

```powershell
pwsh -File scripts/generar_aplicacion_android.ps1 -Release
```

El script valida, analiza, prueba, genera iconos nativos desde
`web/img/pwa-icon-512.png`, y deja APK/AAB junto con SHA-256 bajo
`artifacts/mobile/android/`. La publicacion real requiere firma Android y una
cuenta de Google Play; ni claves ni AAB se versionan.

## iPhone

La compilacion IPA requiere macOS, Xcode y firma Apple:

```powershell
pwsh -File scripts/generar_aplicacion_ios.ps1 -Release
```

En Windows el script falla de forma explicita o puede activar el workflow iOS
con `-TriggerIOSWorkflow`. No se distribuyen IPA ni perfiles de firma desde la
web. El workflow produce candidatos sin publicar hasta configurar secretos de
firma en GitHub.

## CI y descarga publica

`mobile-ci.yml` valida Android e iOS en cada cambio movil; `mobile-release.yml`
se activa manualmente para candidatos. La pagina publica `index.html?pagina=aplicaciones`
lee `web/assets/data/mobile-releases.json`: solo muestra URLs HTTPS permitidas,
version, fecha, checksum y estado. Actualmente no entrega binarios hasta que se
complete la firma y la publicacion oficial.
