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

## 6) Webhooks y pruebas de pago (ngrok)

Para pruebas end-to-end con Mercado Pago desde un entorno local es necesario exponer `http://localhost:8080` mediante un túnel HTTPS. Recomendamos usar `ngrok` (binario incluido en `herramientas/ngrok.exe`). Pasos mínimos:

1. Regístrate en https://dashboard.ngrok.com/signup y copia tu `authtoken` (necesario para sesiones estables).
2. Desde la raíz del repositorio ejecuta (PowerShell):

```powershell
# registrar authtoken (solo la primera vez)
.\herramientas\ngrok.exe authtoken <TU_NGROK_AUTHTOKEN>

# iniciar túnel HTTPS que expone el servidor local en el puerto 8080
.\herramientas\ngrok.exe http 8080 --log=stdout
```

3. Obtén la URL pública desde el dashboard local de ngrok `http://127.0.0.1:4040` o consultando la API local:

```powershell
Invoke-RestMethod -Uri 'http://127.0.0.1:4040/api/tunnels' | ConvertTo-Json -Depth 4
# busca el campo `public_url` (ej. https://abcd-1234.ngrok.io)
```

4. En Mercado Pago (Dashboard → Desarrolladores → Webhooks / Notificaciones) añade una nueva URL de notificación con método `POST` y apunta a:

```
<TU_NGROK_URL>/mercadopago/webhook
```

5. Selecciona los eventos relacionados a pagos. Recomendado mínimo:
- `payment` (notificaciones de pago)
- `payment.updated` (actualizaciones de estado de pago)

6. Alternativamente, al crear una preferencia en el servidor, incluye en el body el campo `notification_url` apuntando a la URL pública de ngrok para asegurar que Mercado Pago notifique ese pago en particular.

Ejemplo (servidor):

```json
{
  "items":[{"title":"Licencia","quantity":1,"unit_price":1.00}],
  "payer":{"email":"cliente@example.com"},
  "back_urls":{"success":"https://tu-dominio/pagos/success","failure":"https://tu-dominio/pagos/failure"},
  "notification_url":"https://abcd-1234.ngrok.io/mercadopago/webhook"
}
```

Notas de seguridad y limpieza:
- El túnel de ngrok es temporario; la URL cambia cada vez que se inicia ngrok salvo que uses un subdominio reservado en tu cuenta. No uses la URL pública en producción.
- Revoca o deja de exponer el túnel una vez finalizadas las pruebas.

### URL pública generada (ejemplo)

En la sesión actual se generó la siguiente URL pública mediante ngrok:

- https://betsey-sinistrous-bluffly.ngrok-free.dev

Para que Mercado Pago envíe notificaciones (webhooks) a tu servidor, añade la siguiente URL en el dashboard de Mercado Pago (Desarrolladores → Webhooks / Notificaciones) o úsala como `notification_url` al crear la preferencia:

```
https://betsey-sinistrous-bluffly.ngrok-free.dev/mercadopago/webhook
```

Eventos recomendados para suscribir en Mercado Pago (mínimo):
- `payment`
- `payment.updated`

Importante: la API local de ngrok indica que el túnel público está actualmente dirigido a `http://localhost:80`. Asegúrate de que el backend esté escuchando en ese puerto; si tu servidor corre en `:8080`, crea un túnel directo a `8080` como se muestra a continuación.

Comandos rápidos para crear/ajustar el túnel (PowerShell, desde la raíz del repo):

```powershell
# registrar authtoken (si no está registrado)
.\herramientas\ngrok.exe authtoken <TU_NGROK_AUTHTOKEN>

# iniciar túnel apuntando a 8080 (si tu servidor escucha en 8080)
.\herramientas\ngrok.exe http 8080 --log=stdout

# o, usando el ngrok del sistema (si está en PATH)
ngrok http 8080 --log=stdout

# obtener la URL pública desde la API local de ngrok
Invoke-RestMethod -Uri 'http://127.0.0.1:4040/api/tunnels' | ConvertTo-Json -Depth 4
```

Verifica que la `public_url` apunte al puerto correcto antes de configurar Mercado Pago. Si la URL pública cambia (ngrok free), actualiza la URL de notificación en el dashboard de Mercado Pago.

