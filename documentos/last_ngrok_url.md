Registro de URL ngrok

Por seguridad, este archivo no debe almacenar URL pública real.

Plantilla:
- Public URL: `https://<tu-subdominio-ngrok>`
- Forwarding (destino local): `http://localhost:<puerto>`
- Webhook Mercado Pago: `https://<tu-subdominio-ngrok>/mercadopago/webhook`

Notas:
- Verifica el túnel activo en `http://127.0.0.1:4040/api/tunnels` antes de cada prueba.
- Si la URL cambia al reiniciar ngrok, actualiza la configuración de webhook en el proveedor de pagos.
