# Arquitectura de la aplicacion movil

Actualizacion: 2026-07-15.

Actualizacion de preproduccion 2026-07-15: `/api/v1` conserva el contrato
oficial para clientes Android/iPhone. El cliente debe tratar `request_id` como
trazabilidad, conservar `Idempotency-Key` hasta una respuesta definitiva y no
usar el identificador de empresa como autoridad. La base de datos y archivos
privados nunca se exponen directamente al dispositivo.

La aplicacion oficial vive en `mobile/powerful_control_system_app` y comparte
una sola base Flutter/Dart para Android e iPhone. No accede a PostgreSQL ni a
archivos empresariales desde el dispositivo: usa exclusivamente la API versionada
`/api/v1/` de PCS mediante HTTPS.

El repositorio incluye los scaffolds nativos `android/` e `ios/`, identificador
`com.powerfulcontrolsystem.pcs`, icono PCS y politica de firma: una distribucion
Android exige un `android/key.properties` local no versionado. La compilacion
iPhone se valida y firma en macOS con Xcode; Windows prepara el codigo comun y
Android, pero no puede generar un paquete iOS distribuible.

## Capas

| Capa | Responsabilidad |
|---|---|
| `config` | Entorno, URL permitida, timeout y banderas de diagnostico. |
| `core` | Cliente HTTP, sesion segura, conectividad y errores sin datos sensibles. |
| `features/auth` | Inicio, segundo factor, restauracion, renovacion y cierre de sesion. |
| `features/companies` | Selector de empresa autorizada y contexto persistente por usuario. |
| `features/dashboard` | Inicio de sesion, menu gobernado por el backend y primer modulo operativo. |
| `features/products` | Consulta paginada de productos como primer modulo movil funcional. |

Riverpod mantiene estado de sesion, conectividad, empresa activa y consultas;
`go_router` protege las rutas autenticadas. `flutter_secure_storage` persiste
solo el Bearer rotatorio y el identificador de empresa seleccionada. No se
guardan contrasenas, secretos de empresa, respuestas de IA ni documentos en
preferencias sin cifrar.

## Sesion y aislamiento

`POST /api/v1/auth/login` aplica las mismas reglas vigentes de confirmacion,
contrasena y TOTP. El token contiene 32 bytes aleatorios, se entrega una vez y
la base de datos conserva un verificador de sesion. Ante un `401`, el cliente
intenta una renovacion unica por `/api/v1/auth/refresh`; el servidor crea un
reemplazo y revoca la sesion anterior. Logout revoca la sesion de inmediato.

La empresa enviada al backend nunca es autoridad. El servidor valida relacion
usuario-empresa, rol, licencia y permiso antes de cada lectura o mutacion.

## Crecimiento sin romper contratos

Los modulos futuros se agregan como `features/<modulo>` con repositorio y modelos
propios; comparten solo el cliente API, autenticacion, permisos y observabilidad.
Las rutas heredadas web siguen siendo compatibles, pero consumidores nuevos usan
el sobre JSON de `/api/v1/`. Carrito, cobro, facturacion, cola offline y buzon
ya tienen fachada v1 para las siguientes pantallas POS.
