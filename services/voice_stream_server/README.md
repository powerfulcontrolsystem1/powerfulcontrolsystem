# Servidor de voz natural para IA

Este servicio es un microservidor abierto para convertir texto de la IA en audio. Usa FastAPI y Piper TTS, ambos componentes de codigo abierto. El backend principal no depende de este proceso: si esta apagado, desactivado o falla, el chat conserva el modo de texto y la voz del navegador como respaldo.

## Variables principales

- `VOICE_STREAM_ENABLED=true`
- `VOICE_STREAM_PIPER_BIN=/opt/powerfulcontrolsystem/services/voice_stream_server/.venv/bin/piper`
- `VOICE_STREAM_TTS_MODEL=/opt/piper/models/es_ES-sharvard-medium.onnx`
- `VOICE_STREAM_TTS_CONFIG=/opt/piper/models/es_ES-sharvard-medium.onnx.json`
- `VOICE_STREAM_MAX_CHARS=4000`
- `VOICE_STREAM_TTS_TIMEOUT=20`
- `VOICE_STREAM_AUTH_HEADER=X-PCS-Voice-Token`
- `VOICE_STREAM_AUTH_TOKEN=<token-largo-generado>`

Si `VOICE_STREAM_AUTH_TOKEN` esta definido, `/health` y `/api/voice/tts` exigen ese token en el header configurado. Guarda el mismo token en Super Administrador > Voz IA streaming; el backend lo persiste cifrado en la tabla de configuraciones.

## Ejecutar localmente en el VPS

```bash
cd /opt/powerfulcontrolsystem/services/voice_stream_server
python3 -m venv .venv
. .venv/bin/activate
pip install -r requirements.txt
uvicorn app:app --host 127.0.0.1 --port 8097
```

Configura en Super Administrador la URL `http://127.0.0.1:8097` si el backend corre en el mismo VPS. Si esta en otro servidor, publica el servicio detras de HTTPS y firewall.

El instalador `scripts/install_voice_stream_server_vps.sh` prepara FastAPI, Uvicorn, `piper-tts`, `espeak-ng`, descarga por defecto la voz abierta `es_ES-sharvard-medium` desde `rhasspy/piper-voices` y crea el servicio systemd `pcs-voice-stream`.

## Voces del chat, robot y secretaria

El frontend puede enviar el identificador de voz configurado por empresa junto con cada solicitud TTS. El robot usa la voz elegida en Administrar empresa > Configuracion > Configurar chat/robot; la secretaria IA fuerza una voz femenina efectiva (`es-CO-female`) para mantener personalidad consistente.

Reglas operativas:

- si el identificador solicitado no existe en Piper, el servicio debe responder error controlado o usar la voz por defecto, sin romper el backend principal.
- si el servicio esta apagado, desactivado o demora demasiado, el chat conserva texto y puede usar la voz del navegador como respaldo.
- no guardar tokens ni credenciales en este README; el token real va en entorno seguro y se registra cifrado desde Super Administrador > Voz IA streaming.
- para agregar voces naturales nuevas, descargar el modelo `.onnx` y `.onnx.json`, registrar su ruta en el servicio o variables de entorno y probar `/health` + `/api/voice/tts` antes de activar en produccion.
