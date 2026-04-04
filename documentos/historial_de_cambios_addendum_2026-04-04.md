2026-04-04: Registro y prueba de clave API Gemini (herramientas de soporte)

- Descripción: Se añadieron dos herramientas temporales en `backend/tools/` para facilitar el registro cifrado y la verificación de la clave API de Google Gemini sin exponer la clave en claro.
- Archivos añadidos:
    - `backend/tools/register_gemini_key/main.go` — lee la clave desde stdin, la cifra usando `CONFIG_ENC_KEY` y la persiste en la tabla `configuraciones` bajo la clave `ai.model.google.gemini_2_0_flash.api_key`.
    - `backend/tools/test_gemini_key/main.go` — lee la clave cifrada desde la BD, la descifra (usando `CONFIG_ENC_KEY`) y realiza una solicitud de prueba a la API de Google Gemini, mostrando la respuesta truncada para verificación.
- Nota: Las herramientas son de uso local y temporal; no persisten la clave en archivos de texto en claro ni imprimen la clave en la salida.
