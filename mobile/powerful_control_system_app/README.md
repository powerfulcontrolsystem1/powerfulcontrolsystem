# Powerful Control System App

Base Flutter única para Android e iPhone. Requiere Flutter estable con Dart 3.3
o superior. No incluya secretos en `--dart-define`, repositorio ni capturas.

En la estacion PCS de desarrollo, Flutter se instala en `D:\Herramientas\flutter`.
El SDK Android se configura en `D:\Herramientas\Android\Sdk` para no consumir la
unidad del sistema. Verifique el entorno con `flutter doctor -v` antes de compilar.

## Inicio local

```powershell
flutter pub get
flutter analyze
flutter test
flutter build apk --debug
flutter run --dart-define=PCS_ENV=development --dart-define=PCS_API_BASE_URL=https://staging.example.invalid
```

Para generar una distribucion Android, copie `android/key.properties.example` a
`android/key.properties` y complete referencias locales al keystore. Ese archivo
esta ignorado por Git. iPhone requiere macOS, Xcode, una cuenta Apple y sus
certificados; desde Windows solo se prepara y valida el codigo comun.

El entorno de producción usa `https://powerfulcontrolsystem.com` por defecto.
La app usa tokens Bearer de sesión de dispositivo, guardados mediante Keychain/
Android Keystore por `flutter_secure_storage`. La API valida empresa, rol,
permiso, licencia e idempotencia en el backend; Flutter no reproduce esas reglas.

La primera funcionalidad nativa es autenticación, segundo factor, selector de
empresa, panel y catálogo de productos. Ventas, pagos, factura y sincronización
offline quedan disponibles por contrato `/api/v1/`, pero se implementarán como
flujos nativos en etapas siguientes sin usar WebView como sustituto.
