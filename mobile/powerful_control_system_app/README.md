# Powerful Control System App

Base Flutter única para Android e iPhone. Requiere Flutter estable con Dart 3.3
o superior. No incluya secretos en `--dart-define`, repositorio ni capturas.

## Inicio local

```powershell
flutter pub get
flutter analyze
flutter test
flutter run --dart-define=PCS_ENV=development --dart-define=PCS_API_BASE_URL=https://staging.example.invalid
```

El entorno de producción usa `https://powerfulcontrolsystem.com` por defecto.
La app usa tokens Bearer de sesión de dispositivo, guardados mediante Keychain/
Android Keystore por `flutter_secure_storage`. La API valida empresa, rol,
permiso, licencia e idempotencia en el backend; Flutter no reproduce esas reglas.

La primera funcionalidad nativa es autenticación, segundo factor, selector de
empresa, panel y catálogo de productos. Ventas, pagos, factura y sincronización
offline quedan disponibles por contrato `/api/v1/`, pero se implementarán como
flujos nativos en etapas siguientes sin usar WebView como sustituto.
