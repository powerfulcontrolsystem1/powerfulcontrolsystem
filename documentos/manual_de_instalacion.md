# Manual de instalacion

Este documento resume los pasos minimos para dejar operativo el proyecto en entorno local y para validar el login Google en produccion sobre el dominio publico.

## 1) Credenciales en Google Cloud Console

1. Abre:
   https://console.cloud.google.com/apis/credentials
2. Selecciona el ID de cliente OAuth correspondiente al proyecto. En este repositorio se ha usado como referencia el cliente web `powerful_oauth`.

### 1.1 Authorized redirect URIs

En la seccion **Authorized redirect URIs** registra exactamente estas URLs:

- `http://localhost:8080/auth/google/callback`
- `https://powerfulcontrolsystem.com/auth/google/callback`

Si tambien vas a servir el portal por `www`, agrega ademas:

- `https://www.powerfulcontrolsystem.com/auth/google/callback`

Importante:

- La URL debe coincidir caracter por caracter con la que envia la aplicacion.
- Si registras `localhost`, la app debe usar `localhost`; si registras `127.0.0.1`, la app debe usar `127.0.0.1`.
- Si Google muestra `redirect_uri_mismatch`, la primera revision debe hacerse en este punto.

### 1.2 Authorized JavaScript origins

En **Authorized JavaScript origins** registra solo el origen, sin rutas:

- `http://localhost:8080`
- `https://powerfulcontrolsystem.com`

Opcional si usas `www`:

- `https://www.powerfulcontrolsystem.com`

No incluyas paths aqui. Ejemplo incorrecto:

- `https://powerfulcontrolsystem.com/auth/google/callback`

### 1.3 Usuarios de prueba

Si la pantalla de consentimiento esta en modo **Testing**:

1. Ve a **OAuth consent screen**.
2. En **Test users**, agrega las cuentas que utilizaras para pruebas.

Cuenta de referencia usada en la documentacion existente:

- `powerfulcontrolsystem@gmail.com`

### 1.4 Nota operativa para produccion en VPS

En la incidencia revisada el backend ya estaba emitiendo correctamente:

- `https://powerfulcontrolsystem.com/auth/google/callback`

Por tanto, si el sitio publico muestra `Error 400: invalid_request` o `redirect_uri_mismatch`, el ajuste prioritario es registrar en Google Cloud Console la URL publica exacta anterior.

## 2) Variables de entorno locales

El script `scripts/iniciar_servidor.ps1` carga credenciales desde variables de entorno del proceso o desde `backend/.env.local` y `backend/.env`.

Ejemplo minimo para desarrollo local:

```powershell
$env:GOOGLE_CLIENT_ID = "tu-client-id.apps.googleusercontent.com"
$env:GOOGLE_CLIENT_SECRET = "<definir-en-entorno-seguro>"
$env:GOOGLE_REDIRECT_URL = "http://localhost:8080/auth/google/callback"
$env:PORT = "8080"
.\scripts\iniciar_servidor.ps1
```

No guardes secretos reales en documentos versionados.

## 3) Configuracion operativa en produccion

Para produccion existen dos fuentes de configuracion validas segun el estado del entorno:

- Variables de entorno del proceso.
- Configuracion persistida en la base `pcs_superadministrador`, tabla `configuraciones`.

Claves relevantes:

- `google.client_id`
- `google.client_secret`
- `google.redirect_url`

Valor esperado para `google.redirect_url` en VPS:

- `https://powerfulcontrolsystem.com/auth/google/callback`

## 4) Reinicio y pruebas

### 4.1 Prueba local

1. Reinicia el servidor con `scripts/iniciar_servidor.ps1`.
2. Abre `http://localhost:8080`.
3. Pulsa el boton de login con Google.
4. Verifica que Google redirige a `http://localhost:8080/auth/google/callback`.

### 4.2 Prueba publica en VPS

1. Abre `https://powerfulcontrolsystem.com`.
2. Inicia el flujo de login Google.
3. Verifica que la URL enviada a Google contiene:

   `redirect_uri=https://powerfulcontrolsystem.com/auth/google/callback`

4. Si la redireccion es correcta pero Google rechaza la solicitud, revisa nuevamente la configuracion del cliente OAuth en Google Cloud Console.

### 4.3 Wildcard HTTPS para subdominios

DNS confirmado en Hostinger para produccion:

- `A @ -> 2.24.197.58`
- `A * -> 2.24.197.58`
- `CNAME www -> powerfulcontrolsystem.com`

Emision manual ejecutada en VPS:

```bash
certbot certonly --manual --preferred-challenges dns -d powerfulcontrolsystem.com -d *.powerfulcontrolsystem.com
```

Resultado operativo documentado:

- Certificado emitido en `/etc/letsencrypt/live/powerfulcontrolsystem.com-0001/fullchain.pem`
- Llave privada en `/etc/letsencrypt/live/powerfulcontrolsystem.com-0001/privkey.pem`
- Vigencia del certificado emitido el `2026-04-16`: hasta `2026-07-15`

Importante:

- La emision fue por desafio manual `DNS-01`.
- El registro `TXT _acme-challenge` usado durante la validacion puede borrarse despues de finalizar la emision.
- `nginx` debe apuntar al certificado wildcard nuevo para cubrir `powerfulcontrolsystem.com` y `*.powerfulcontrolsystem.com`.

### 4.4 Renovacion manual del wildcard

La renovacion NO es automatica en este esquema porque `certbot --manual` no instala un hook de autenticacion DNS.

Debe repetirse el mismo proceso de renovacion manual en cualquiera de estas situaciones:

- Antes del vencimiento del certificado actual.
- Si se migra el VPS y ya no existe el certificado en el nuevo servidor.
- Si se elimina o reemplaza accidentalmente el certificado activo de `nginx`.
- Si en el futuro se quieren cambiar los dominios cubiertos por el certificado.

Frecuencia recomendada:

- Revisar el vencimiento al menos una vez por mes.
- Ejecutar la renovacion manual alrededor de 30 dias antes del vencimiento.
- Para el certificado emitido el `2026-04-16`, la ventana segura de renovacion comienza aproximadamente el `2026-06-15`.

Comando de renovacion manual:

```bash
certbot certonly --manual --preferred-challenges dns -d powerfulcontrolsystem.com -d *.powerfulcontrolsystem.com
```

Flujo de renovacion:

1. Ejecutar el comando en el VPS.
2. Crear el `TXT _acme-challenge` exacto que entregue `certbot` en Hostinger.
3. Esperar propagacion DNS.
4. Continuar la emision en consola.
5. Confirmar rutas finales del certificado en `/etc/letsencrypt/live/powerfulcontrolsystem.com-0001/` o la version nueva que genere `certbot`.
6. Validar `nginx -t` y recargar `nginx` si el nombre del certificado cambia.

## 4.5 Subdominio publico de prueba para venta digital

Se dejo publicado un subdominio dedicado de prueba para la pagina publica global `venta_digital.html`:

- `https://venta-digital.powerfulcontrolsystem.com/`

Comportamiento esperado:

- La raiz del subdominio responde `302` hacia `/venta_digital.html`.
- La pagina final `https://venta-digital.powerfulcontrolsystem.com/venta_digital.html` responde `200`.
- El subdominio usa el mismo certificado wildcard del dominio principal.

Este subdominio es una prueba controlada de infraestructura y no reemplaza la raiz generica de subdominios por empresa que sigue destinada a `venta_publica.html`.

## 5) Notas de depuracion OAuth

- Si aparece `redirect_uri_mismatch`, compara exactamente la URL del error con la URL registrada en Google Cloud.
- Si aparece `Origen no valido`, revisa **Authorized JavaScript origins** y elimina cualquier path agregado por error.
- En los logs del servidor se puede revisar la `Auth URL` completa para confirmar parametros enviados.
- En dominio publico, el esquema debe ser `https`.

## 6) Webhooks y pruebas de pago con ngrok

Para pruebas end-to-end de pagos desde entorno local se puede exponer `http://localhost:8080` mediante un tunel HTTPS temporal con `ngrok`.

### 6.1 Pasos minimos

1. Registrate en:
   https://dashboard.ngrok.com/signup
2. Copia tu `authtoken`.
3. Desde la raiz del repositorio ejecuta:

```powershell
# registrar authtoken (solo la primera vez)
.\herramientas\ngrok.exe authtoken <TU_NGROK_AUTHTOKEN>

# iniciar tunel HTTPS que expone el servidor local en el puerto 8080
.\herramientas\ngrok.exe http 8080 --log=stdout
```

4. Obtén la URL publica desde `http://127.0.0.1:4040` o mediante la API local:

```powershell
Invoke-RestMethod -Uri 'http://127.0.0.1:4040/api/tunnels' | ConvertTo-Json -Depth 4
```

Busca el campo `public_url`.

### 6.2 Mercado Pago: notification_url

En Mercado Pago agrega una notificacion `POST` apuntando a:

- `<TU_NGROK_URL>/mercadopago/webhook`

Eventos recomendados como minimo:

- `payment`
- `payment.updated`

Ejemplo de payload de preferencia en servidor:

```json
{
  "items": [
    {
      "title": "Licencia",
      "quantity": 1,
      "unit_price": 1.0
    }
  ],
  "payer": {
    "email": "cliente@example.com"
  },
  "back_urls": {
    "success": "https://tu-dominio/pagos/success",
    "failure": "https://tu-dominio/pagos/failure"
  },
  "notification_url": "https://<tu-subdominio-ngrok>/mercadopago/webhook"
}
```

Notas:

- El tunel de ngrok es temporal.
- No almacenes URLs publicas reales de tuneles en documentacion versionada.
- No uses URLs de ngrok en produccion.

## 7) Alternativa de pago: Nequi con Wompi

### 7.1 Credenciales necesarias

En Wompi se requieren estas llaves:

- `public_key`
- `private_key`
- `integrity_key`

Configuracion en la UI:

- `Super administrador -> Configuracion avanzada -> Wompi`

API asociada:

- `GET /super/api/config/wompi`
- `PUT /super/api/config/wompi`

### 7.2 Sandbox y numeros de prueba

En sandbox se documentaron como referencia estos telefonos:

- `3991111111` -> `APPROVED`
- `3992222222` -> `DECLINED`

### 7.3 Flujo tecnico implementado

1. El frontend consulta terminos con `GET /wompi/terms`.
2. El usuario ingresa el celular Nequi y acepta terminos.
3. El backend crea la transaccion con `POST /wompi/create_transaction_nequi`.
4. El frontend consulta estado con `GET /wompi/transaction_status?id=...`.
5. Si el estado es `APPROVED`, la licencia se activa automaticamente.

### 7.4 Requisitos importantes

- `customer_email` debe existir.
- El celular Nequi debe tener 10 digitos e iniciar por `3`.
- No mezclar llaves de sandbox con entorno productivo.

## 8) Epayco: URL de respuesta y confirmacion

Cuando configures Epayco para el checkout de licencias debes registrar tanto la pagina publica de respuesta como la URL de confirmacion del webhook.

### 8.1 URLs que debes registrar en Epayco

Produccion:

- URL de respuesta: `https://powerfulcontrolsystem.com/epayco/respuesta.html`
- URL de confirmacion: `https://powerfulcontrolsystem.com/epayco/webhook`

### 8.2 Valores exactos del formulario de Epayco

En el formulario `URL Respuesta y Confirmacion` de Epayco registra estos valores:

- `¿URL de Respuesta?` -> `Si`
- `URL de respuesta` -> `https://powerfulcontrolsystem.com/epayco/respuesta.html`
- `¿URL de Confirmacion?` -> `Si`
- `Metodo de consulta` -> `POST`
- `URL de Confirmacion` -> `https://powerfulcontrolsystem.com/epayco/webhook`
- `Estados a confirmar` -> `Todos`
- `Numero Reintentos` -> `5`
- `Hora limite para permitir reintentos` -> `23:59`
- `Personalizar codigo de respuesta` -> `Si`
- `Respuesta` -> `200`

Si el panel de Epayco no permite `23:59`, usa la hora mas alta disponible dentro del mismo dia.

Importante:

- La URL de respuesta no activa la licencia por si sola; solo recibe la vuelta de Epayco y redirige a `pagar_licencia.html` para que el sistema consulte el estado real del pago.
- La URL de confirmacion si debe permanecer apuntando al backend porque recibe la notificacion servidor a servidor de la pasarela.
- No registres `localhost`, `127.0.0.1` ni URLs temporales de ngrok en produccion.

### 8.3 Flujo esperado

1. El usuario inicia el pago desde `pagar_licencia.html`.
2. El backend crea el checkout de Epayco con:
  - `response = https://powerfulcontrolsystem.com/epayco/respuesta.html`
  - `confirmation = https://powerfulcontrolsystem.com/epayco/webhook`
3. Epayco devuelve al navegador a `/epayco/respuesta.html`.
4. Esa pagina redirige a `/pagar_licencia.html` con el contexto del pago.
5. El frontend y el backend consultan el estado real antes de marcar la licencia como activa.

### 8.4 Verificacion recomendada despues de configurar Epayco

1. Abre `https://powerfulcontrolsystem.com/pagar_licencia.html` con una licencia lista para pago.
2. Inicia un checkout con Epayco.
3. Verifica que, al volver desde la pasarela, la navegacion pase por:
  - `https://powerfulcontrolsystem.com/epayco/respuesta.html`
  - y luego por `https://powerfulcontrolsystem.com/pagar_licencia.html`
4. Confirma que el sistema muestre el resultado final del pago y no solo el retorno visual de la pasarela.

## 9) Confirmacion de correo para usuarios de empresa con Gmail SMTP

Cuando un administrador crea un usuario de empresa, el sistema puede enviar un enlace de confirmacion por correo.

### 9.1 Configuracion en Gmail

1. Usa una cuenta Gmail remitente.
2. Activa verificacion en dos pasos.
3. Genera una contrasena de aplicacion.
4. Configura el panel:

- `Super administrador -> Configuracion avanzada -> Correo Gmail - Confirmacion de usuarios`

Campos recomendados:

- Correo remitente: `tu-cuenta@gmail.com`
- SMTP Host: `smtp.gmail.com`
- SMTP Port: `587`
- URL base de confirmacion: `http://localhost:8080` en local o tu dominio HTTPS en produccion.

### 9.2 Flujo de confirmacion

1. Crear usuario desde `Administrar empresa -> Usuarios`.
2. El backend envia un enlace:

   `/auth/confirmar_correo?token=<token>`

3. El usuario abre el enlace y el sistema marca el correo como confirmado.

### 9.3 Endpoints relacionados

- `GET /super/api/config/gmail`
- `PUT /super/api/config/gmail`
- `GET /api/empresa/roles_de_usuario?empresa_id=...`
- `GET /api/empresa/usuarios`
- `POST /api/empresa/usuarios`
- `PUT /api/empresa/usuarios`
- `DELETE /api/empresa/usuarios`
- `PUT /api/empresa/usuarios?action=reenviar_confirmacion`
- `GET /auth/confirmar_correo?token=...`

### 8.4 Nota operativa

Si Gmail rechaza el envio, valida:

- que la contrasena sea de aplicacion,
- que host y puerto sean correctos,
- que no exista un bloqueo temporal de seguridad en la cuenta.

## 10) Configuracion del Servidor de Soporte Remoto (RustDesk)

El sistema de soporte remoto nativo funciona integrando un servidor propio de RustDesk (hbbs y hbbr) en el VPS, que permite saltarse intermediarios publicos y reducir la latencia para el soporte a empresas clientes.

### 10.1 Instalacion en Ubuntu / Debian VPS

1. Descargar el script oficial de instalacion de RustDesk Server:
   \\\ash
   wget https://raw.githubusercontent.com/rustdesk/rustdesk-server/master/setup.sh
   chmod +x setup.sh
   ./setup.sh
   \\\
   
2. Obtener la Clave Publica de cifrado (Key):
   El script mostrara en pantalla la Public Key del servidor, necesaria para que los clientes se conecten. Tambien puedes verla luego en:
   \\\ash
   cat /opt/rustdesk/id_ed25519.pub
   \\\

3. Comprobar que los servicios esten en ejecucion:
   \\\ash
   systemctl is-active rustdesk-hbbs
   systemctl is-active rustdesk-hbbr
   \\\
   *(Deben responder 'active')*

### 10.2 Administracion desde el Panel Super
Una vez instalados los servicios en el VPS, el panel Super Administrador en la seccion **Configuracion avanzada** (`/super/configuracion_avanzada.html`) leerá su estado usando comandos `systemctl` sobre el host donde corre RustDesk.

Si el backend se ejecuta directamente en Linux dentro del VPS, la comprobacion es local. Si el backend se ejecuta en Windows durante desarrollo, el panel usa la configuracion SSH existente (`DB_VPS_SSH_HOST`, `DB_VPS_SSH_USER`, `DB_VPS_SSH_KEY_PATH`) para consultar y probar RustDesk remotamente en el VPS mediante `plink.exe` (o `ssh.exe` cuando la llave no sea `.ppk`).

Se proveen botones para iniciar, detener, reiniciar y probar el servicio de forma visual. Para que los botones del panel funcionen en produccion (Linux VPS), el usuario interno que ejecuta el backend (`server_linux_amd64`) o el usuario SSH usado desde Windows debera tener permisos sudo sin contrasena *exclusivamente para los servicios rustdesk*:

\\\ash
# Ejecutar en el servidor VPS con root
visudo

# Anadir al final del archivo para el usuario 'miusuario':
miusuario ALL=(ALL) NOPASSWD: /usr/bin/systemctl start rustdesk-hbbs, /usr/bin/systemctl stop rustdesk-hbbs, /usr/bin/systemctl restart rustdesk-hbbs, /usr/bin/systemctl start rustdesk-hbbr, /usr/bin/systemctl stop rustdesk-hbbr, /usr/bin/systemctl restart rustdesk-hbbr
\\\

### 10.3 Cliente y Modulo Empresa
El modulo de 'Soporte Remoto' del administrador de empresa en el ERP proveera instrucciones para descargar el cliente RustDesk e instruira introducir la IP del servidor ID (hbbs/hbbr) y opcionalmente su Clave (Key), garantizando conexiones privadas y directas desde el portal hacia el equipo del cliente.
