#!/usr/bin/env bash
set -euo pipefail

APP_DIR="${APP_DIR:-/opt/powerfulcontrolsystem}"
SERVICE_DIR="$APP_DIR/services/voice_stream_server"
ENV_FILE="${ENV_FILE:-/etc/powerfulcontrolsystem-voice.env}"
RUN_USER="${RUN_USER:-www-data}"
PORT="${VOICE_STREAM_PORT:-8097}"
MODEL_DIR="${VOICE_STREAM_MODEL_DIR:-/opt/piper/models}"
MODEL_NAME="${VOICE_STREAM_MODEL_NAME:-es_ES-sharvard-medium}"
MODEL_BASE_URL="${VOICE_STREAM_MODEL_BASE_URL:-https://huggingface.co/rhasspy/piper-voices/resolve/main/es/es_ES/sharvard/medium}"

if [[ ! -d "$SERVICE_DIR" ]]; then
  echo "No existe $SERVICE_DIR. Copia el proyecto al VPS antes de ejecutar este instalador." >&2
  exit 1
fi

apt-get update
apt-get install -y python3 python3-venv python3-pip curl ca-certificates espeak-ng

python3 -m venv "$SERVICE_DIR/.venv"
"$SERVICE_DIR/.venv/bin/pip" install --upgrade pip
"$SERVICE_DIR/.venv/bin/pip" install -r "$SERVICE_DIR/requirements.txt"
"$SERVICE_DIR/.venv/bin/pip" install piper-tts

mkdir -p "$MODEL_DIR"
MODEL_PATH="$MODEL_DIR/$MODEL_NAME.onnx"
MODEL_CONFIG_PATH="$MODEL_DIR/$MODEL_NAME.onnx.json"
if [[ ! -f "$MODEL_PATH" ]]; then
  curl -fL "$MODEL_BASE_URL/$MODEL_NAME.onnx" -o "$MODEL_PATH"
fi
if [[ ! -f "$MODEL_CONFIG_PATH" ]]; then
  curl -fL "$MODEL_BASE_URL/$MODEL_NAME.onnx.json" -o "$MODEL_CONFIG_PATH"
fi

if [[ ! -f "$ENV_FILE" ]]; then
  if command -v openssl >/dev/null 2>&1; then
    AUTH_TOKEN="$(openssl rand -base64 32 | tr '+/' '-_' | tr -d '=')"
  else
    AUTH_TOKEN="$(python3 - <<'PY'
import secrets
print(secrets.token_urlsafe(32))
PY
)"
  fi
  cat > "$ENV_FILE" <<EOF
VOICE_STREAM_ENABLED=true
VOICE_STREAM_PIPER_BIN=$SERVICE_DIR/.venv/bin/piper
VOICE_STREAM_TTS_MODEL=$MODEL_PATH
VOICE_STREAM_TTS_CONFIG=$MODEL_CONFIG_PATH
VOICE_STREAM_MAX_CHARS=4000
VOICE_STREAM_TTS_TIMEOUT=20
VOICE_STREAM_AUTH_HEADER=X-PCS-Voice-Token
VOICE_STREAM_AUTH_TOKEN=$AUTH_TOKEN
EOF
  chmod 640 "$ENV_FILE"
fi

cat > /etc/systemd/system/pcs-voice-stream.service <<EOF
[Unit]
Description=Powerful Control System Voice Stream Server
After=network.target

[Service]
Type=simple
User=$RUN_USER
WorkingDirectory=$SERVICE_DIR
EnvironmentFile=$ENV_FILE
ExecStart=$SERVICE_DIR/.venv/bin/uvicorn app:app --host 127.0.0.1 --port $PORT
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable pcs-voice-stream
systemctl restart pcs-voice-stream
systemctl status pcs-voice-stream --no-pager
