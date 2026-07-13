FROM python:3.12-slim

WORKDIR /app

RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates curl espeak-ng \
  && rm -rf /var/lib/apt/lists/*

COPY services/voice_stream_server/requirements.txt /app/requirements.txt
RUN pip install --no-cache-dir -r /app/requirements.txt piper-tts

COPY services/voice_stream_server /app
COPY deploy/docker/voice-stream-entrypoint.sh /entrypoint.sh
RUN useradd --system --uid 10001 --create-home pcs \
  && mkdir -p /models /tmp/pcs_voice \
  && chmod +x /entrypoint.sh \
  && chown -R pcs:pcs /app /models /tmp/pcs_voice /entrypoint.sh

EXPOSE 8000
USER 10001:10001
ENTRYPOINT ["/entrypoint.sh"]
CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8000"]
