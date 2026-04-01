Descripción de la carpeta herramientas

Este documento describe los utilitarios colocados en la carpeta `herramientas` dentro del repositorio y su propósito.

- Ruta: herramientas/
- Propósito: contener utilitarios y binarios destinados a facilitar pruebas locales (túneles HTTP, herramientas de diagnóstico, scripts auxiliares). Los binarios aquí se dejaron por conveniencia para pruebas en entornos locales; el usuario decide si versionarlos o mantenerlos fuera del control de versiones.

Archivo incluido (2026-03-29):

- herramientas/ngrok.exe
  - Descripción: Ejecutable de ngrok (stable, Windows amd64). Se emplea para exponer el servidor local (`http://localhost:8080`) a una URL pública HTTPS y así permitir que servicios externos (ej. Mercado Pago) envíen webhooks durante pruebas.
  - Uso básico desde PowerShell (desde la raíz del repo):
    - `.	ools\ngrok.exe http 8080`  (o `.	ools\ngrok.exe http 8080 --log=stdout`)
    - Registrar authtoken (opcional): `.	ools\ngrok.exe authtoken <TU_AUTHTOKEN>`
  - Nota de seguridad: el binario no contiene secretos, pero exponer servicios al público conlleva riesgos. Usar ngrok sólo en entornos de prueba y asegurarse de revocar/rotar credenciales cuando sea necesario.

  - Requisito importante: la versión moderna de ngrok requiere una cuenta verificada y un `authtoken` para crear túneles persistentes. Si al iniciar ngrok obtienes un error "failed to auth" o `ERR_NGROK_4018`, registra una cuenta en https://dashboard.ngrok.com/signup y ejecuta:

    ```powershell
    .\herramientas\ngrok.exe authtoken <TU_NGROK_AUTHTOKEN>
    .\herramientas\ngrok.exe http 8080 --log=stdout
    ```

  - Obtener la URL pública (ejemplo): abrir `http://127.0.0.1:4040` o:

    ```powershell
    Invoke-RestMethod -Uri 'http://127.0.0.1:4040/api/tunnels' | ConvertTo-Json -Depth 4
    ```

  - Nota: si prefieres no usar ngrok, puedes exponer el servidor con servicios alternativos (Cloudflare Tunnel, localtunnel, o un servidor público), pero ajusta la `notification_url` en Mercado Pago a la URL resultante.

Política de trazabilidad:
- La creación/descarga de este binario fue registrada en `documentos/historial_de_cambios` con fecha 2026-03-29.
- Si se desea eliminar o reemplazar el binario, actualizar `documentos/descripcion_de_archivos` y añadir una entrada en `documentos/historial_de_cambios` mencionando la razón.

Fecha: 2026-03-29
Responsable: agente local (descarga a petición del usuario)

No almacenar URL pública activa de túneles en este documento.

Estado del túnel: verificar siempre en tiempo real con `http://127.0.0.1:4040/api/tunnels` y confirmar el puerto local de destino antes de pruebas.

Plantilla para notificaciones de Mercado Pago:

```
https://<tu-subdominio-ngrok>/mercadopago/webhook
```

Si necesitas que el túnel apunte a `:8080` (si tu backend escucha allí), crea el túnel con `ngrok http 8080`.
