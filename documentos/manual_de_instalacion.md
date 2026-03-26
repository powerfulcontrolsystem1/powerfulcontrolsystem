# Manual de instalación

Este documento explica los pasos mínimos para configurar las credenciales de OAuth de Google y dejar la aplicación preparada en local.

## 1) Credenciales en Google Cloud Console
1. Abre: https://console.cloud.google.com/apis/credentials
2. Selecciona el ID de cliente correspondiente (por ejemplo: `powerful_oauth`).

### Authorized redirect URIs
- Añade exactamente la siguiente URI en la sección **Authorized redirect URIs**:

  http://localhost:8080/auth/google/callback

  *IMPORTANTE:* la URI debe coincidir carácter por carácter con la que envía la aplicación (esquema, host, puerto y path). Si registras `127.0.0.1` en vez de `localhost`, la aplicación debe usar exactamente esa variante.

### Authorized JavaScript origins
- Si vas a usar la web desde el navegador (CORS / JS), añade únicamente el origen (sin path):

  http://localhost:8080

  *No* incluyas paths aquí (p. ej. `/auth/google/callback`) — eso causará el error "Origen no válido".

## 2) Usuarios de prueba (si la pantalla de consentimiento está en Testing)
- Ve a **OAuth consent screen** → **Test users** y añade las cuentas que usarás para pruebas.
- Cuenta principal añadida en este proyecto (actual): powerfulcontrolsystem@gmail.com

## 3) Variables de entorno locales
- El script de arranque `scripts\iniciar_servidor.ps1` intenta leer `documentos/descripcion_del_proyecto` para extraer `GOOGLE_CLIENT_ID` y `GOOGLE_CLIENT_SECRET`.
- Alternativamente exporta manualmente antes de ejecutar el script:

  ```powershell
  $env:GOOGLE_CLIENT_ID = "tu-client-id.apps.googleusercontent.com"
  $env:GOOGLE_CLIENT_SECRET = "tu-secret"
  $env:GOOGLE_REDIRECT_URL = "http://localhost:8080/auth/google/callback"
  $env:PORT = "8080"
  .\scripts\iniciar_servidor.ps1
  ```

## 4) Reinicio y pruebas
1. Reinicia el servidor con `scripts\iniciar_servidor.ps1`.
2. Abre el navegador en: http://localhost:8080
3. Pulsa "Login con Google" y comprueba que el flujo redirige a Google y vuelve a `http://localhost:8080/auth/google/callback`.

## 5) Notas de depuración
- Si obtienes `redirect_uri_mismatch`, compara exactamente la URL en el error con la registrada en Google Cloud (incluyendo `http://` y el puerto).
- En los logs del servidor se registra la `Auth URL:` completa para revisar parámetros enviados.

---

Correo principal registrado en la documentación: powerfulcontrolsystem@gmail.com
