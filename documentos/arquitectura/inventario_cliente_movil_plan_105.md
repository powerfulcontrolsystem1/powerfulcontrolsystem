# Inventario de cliente movil - Plan 105

Fecha de corte: 2026-07-21. Revision local sin compilar, firmar ni publicar.

## Hallazgo

La ruta `mobile/powerful_control_system_app` contiene solamente directorios de
artefactos (`build` y `.dart_tool`). No contiene `pubspec.yaml`, `lib`,
`android`, `ios`, pruebas, README, lockfile, pipeline de CI ni configuracion de
firma. Por ello no existe fuente versionada suficiente para reproducir una APK
o IPA, revisar dependencias, ejecutar pruebas o validar identidad de paquete.

Los artefactos locales no son una fuente de release: pueden corresponder a otro
commit, toolchain o secreto de firma y no permiten auditoria reproducible.

## Decision requerida

P105-020 queda fuera del lanzamiento web/servidor hasta elegir expresamente
una alternativa:

1. **Incluir movil:** recuperar el proyecto fuente completo en control de
   versiones, fijar Flutter/Dart y Gradle/Xcode, versionar lockfile y pruebas,
   crear CI reproducible y configurar firma mediante secret store externo.
   Antes de publicar, generar APK/IPA desde un runner limpio y registrar hash,
   version, identificador de paquete y evidencia funcional.
2. **Excluir movil:** retirar cualquier promesa de app nativa reproducible de
   documentacion operativa y declarar la experiencia web responsive/PWA como
   unico cliente soportado para este release.

No se deben versionar `.dart_tool`, `build`, claves, keystores, APK/IPA ni
credenciales como sustituto del proyecto fuente.

## Criterio de cierre

No cerrar con este inventario. El cierre exige la decision anterior y, si se
incluye movil, una compilacion limpia firmada con pruebas y trazabilidad de
release; si se excluye, documentacion vigente sin afirmaciones incompatibles.
