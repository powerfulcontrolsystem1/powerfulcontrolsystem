#!/bin/sh
set -eu

MODEL_DIR="${VOICE_STREAM_MODEL_DIR:-/models}"
MODEL_NAME="${VOICE_STREAM_MODEL_NAME:-es_ES-sharvard-medium}"
MODEL_BASE_URL="${VOICE_STREAM_MODEL_BASE_URL:-https://huggingface.co/rhasspy/piper-voices/resolve/main/es/es_ES/sharvard/medium}"

mkdir -p "$MODEL_DIR"

if [ -z "${VOICE_STREAM_TTS_MODEL:-}" ]; then
  export VOICE_STREAM_TTS_MODEL="$MODEL_DIR/$MODEL_NAME.onnx"
fi

if [ -z "${VOICE_STREAM_TTS_CONFIG:-}" ]; then
  export VOICE_STREAM_TTS_CONFIG="$MODEL_DIR/$MODEL_NAME.onnx.json"
fi

if [ ! -f "$VOICE_STREAM_TTS_MODEL" ]; then
  curl -fL "$MODEL_BASE_URL/$MODEL_NAME.onnx" -o "$VOICE_STREAM_TTS_MODEL"
fi

if [ ! -f "$VOICE_STREAM_TTS_CONFIG" ]; then
  curl -fL "$MODEL_BASE_URL/$MODEL_NAME.onnx.json" -o "$VOICE_STREAM_TTS_CONFIG"
fi

exec "$@"
