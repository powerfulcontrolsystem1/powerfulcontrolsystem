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
- El script de arranque `scripts\iniciar_servidor.ps1` carga credenciales solo desde variables de entorno del proceso o desde `backend/.env.local` y `backend/.env`.
- Exporta manualmente antes de ejecutar el script (no guardes valores reales en documentos):

  ```powershell
  $env:GOOGLE_CLIENT_ID = "tu-client-id.apps.googleusercontent.com"
  $env:GOOGLE_CLIENT_SECRET = "<definir-en-entorno-seguro>"
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

### 6.1) Sandbox de Mercado Pago (evitar botón "Pagar" deshabilitado)

Si el checkout sandbox muestra el botón de pagar deshabilitado, valida estos puntos:

1. Usa usuarios de prueba diferentes para vendedor y comprador en Mercado Pago.
  - Vendedor de prueba: es la cuenta dueña del `TEST-...access_token` guardado en el sistema.
  - Comprador de prueba: debe ser otra cuenta de prueba distinta (no la del vendedor).
2. El sistema ahora selecciona modo automáticamente según la credencial del vendedor:
  - `TEST-...` -> usa checkout sandbox (`sandbox_init_point`).
  - credencial productiva -> usa checkout estándar (`init_point`).
  - Si necesitas forzar sandbox explícito desde la URL: `pagar_licencia.html?...&mp_mode=sandbox`.
  - Si necesitas forzar estándar explícito: `pagar_licencia.html?...&mp_mode=prod`.
  - Protección adicional: si detecta credencial `TEST-...` con checkout estándar, el frontend redirige automáticamente a sandbox para evitar el error "Una de las partes es de prueba".
3. El prefill automático del pagador está desactivado por defecto para evitar bloquear el checkout.
  - Solo habilitar si realmente se requiere: `pagar_licencia.html?...&prefill_payer=1`.
  - Con credenciales `TEST-...` el backend ignora el prefill de payer para evitar la mezcla test/productivo.
4. Si usas `webhook_secret`, la firma debe coincidir exactamente con la configurada en Mercado Pago.
  - Sin firma válida, el webhook será rechazado (401) por seguridad.

Nota: si usas credenciales de prueba (`TEST-...`), realiza el pago entrando al checkout con comprador sandbox, no con la cuenta vendedora.

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

### URL pública generada (plantilla)

No almacenes URLs públicas reales de túneles en documentación versionada. Usa siempre una plantilla:

- `https://<tu-subdominio-ngrok>/mercadopago/webhook`

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

## 7) Alternativa de pago: Nequi (Wompi Colombia)

Se agregó una alternativa de pago con Nequi usando Wompi (además de Mercado Pago).

### 7.1 Credenciales necesarias

En Wompi (dashboard de comercios) obtén estas llaves:

- `public_key` (`pub_test_...` o `pub_prod_...`)
- `private_key` (`prv_test_...` o `prv_prod_...`)
- `integrity_key` (`test_integrity_...` o `prod_integrity_...`)

Configúralas en la UI de super administrador:

- `Super administrador -> Configuración avanzada -> Wompi (Nequi)`
- En la tarjeta de Nequi usa el switch de modo:
  - `Sandbox` para pruebas (con teléfonos como `3991111111`).
  - `Real` para cobros productivos.

API asociada:

- `GET/PUT /super/api/config/wompi`

### 7.2 Sandbox y números de prueba Nequi

Según documentación de Wompi para Colombia, en sandbox puedes usar:

- `3991111111` -> transacción `APPROVED`
- `3992222222` -> transacción `DECLINED`

### 7.3 Flujo técnico implementado

1. El frontend solicita términos con `GET /wompi/terms`.
2. El usuario ingresa celular Nequi y acepta términos.
3. El backend crea transacción con `POST /wompi/create_transaction_nequi`.
4. El frontend consulta estado con `GET /wompi/transaction_status?id=...`.
5. Si el estado queda `APPROVED`, se activa la licencia automáticamente.

### 7.4 Requisitos importantes

- `customer_email` debe existir (se toma del usuario autenticado).
- El celular Nequi debe tener formato colombiano de 10 dígitos (inicia por `3`).
- En sandbox usa llaves `*_test_*`; en producción usa `*_prod_*`.
- No mezclar llaves de un ambiente con URL/base del otro.

### 7.5 Método temporal: Activar licencia sin pagar

Para avanzar en desarrollo sin bloquearse por pasarelas:

- En `pagar_licencia.html` existe el método `Activar licencia sin pagar`.
- Ese flujo llama `POST /licencias/activar_sin_pago` con `licencia_id` y `empresa_id`.
- Si todo sale bien, activa la licencia y redirige a `administrar_empresa.html`.

Importante:

- Este método es para avance/prototipo interno y no reemplaza los cobros reales de Mercado Pago o Nequi.

## 8) Confirmación de correo para usuarios de empresa (Gmail SMTP)

Se añadió el submódulo de usuarios por empresa con confirmación de correo. Cuando un administrador crea un usuario, el sistema envía un enlace de confirmación al correo indicado.

### 8.1 Configuración en Gmail (cuenta remitente)

1. Usa una cuenta Gmail que será el remitente (ejemplo: `tu-cuenta@gmail.com`).
2. Activa verificación en dos pasos en esa cuenta.
3. Genera una contraseña de aplicación (App Password) para "Correo".
4. Guarda esa contraseña en el panel:
  - `Super administrador -> Configuración avanzada -> Correo Gmail — Confirmación de usuarios`.

Campos recomendados:

- Correo remitente: `tu-cuenta@gmail.com`
- Contraseña de aplicación: (16 caracteres generados por Google)
- SMTP Host: `smtp.gmail.com`
- SMTP Port: `587`
- URL base de confirmación: `http://localhost:8080` (o tu dominio HTTPS en producción)

Opcional de seguridad:

- Activar `Cifrar contraseña al guardar` si el backend tiene `CONFIG_ENC_KEY` configurada.

### 8.2 Flujo de confirmación

1. Crear usuario desde `Administrar empresa -> Usuarios`.
2. El backend envía un correo con enlace:

  `/auth/confirmar_correo?token=<token>`

3. El usuario abre el enlace y el sistema marca el correo como confirmado.

### 8.3 Endpoints relacionados

- `GET/PUT /super/api/config/gmail` -> configuración SMTP Gmail.
- `GET /api/empresa/roles_de_usuario?empresa_id=...` -> roles válidos para la empresa.
- `GET/POST/PUT/DELETE /api/empresa/usuarios` -> CRUD de usuarios por empresa.
- `PUT /api/empresa/usuarios?action=reenviar_confirmacion` -> reenvío del correo.
- `GET /auth/confirmar_correo?token=...` -> confirma correo (endpoint público).

### 8.4 Nota operativa

Si Gmail rechaza el envío, valida:

- que la contraseña sea de aplicación (no la contraseña normal de la cuenta),
- que host/puerto sean correctos,
- que no exista bloqueo temporal de seguridad en la cuenta de Gmail.

