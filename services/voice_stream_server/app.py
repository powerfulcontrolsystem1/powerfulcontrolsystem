import os
import hmac
import shutil
import subprocess
import tempfile
import uuid
from pathlib import Path
from typing import Iterator

from fastapi import FastAPI, HTTPException, Request
from fastapi.responses import StreamingResponse
from pydantic import BaseModel


app = FastAPI(title="PCS Voice Stream Server", version="1.0.0")


def env_bool(name: str, default: bool = True) -> bool:
    raw = os.getenv(name, "")
    if not raw:
        return default
    return raw.strip().lower() in {"1", "true", "on", "yes", "activo", "enabled"}


def piper_bin() -> str:
    return os.getenv("VOICE_STREAM_PIPER_BIN", "piper").strip() or "piper"


def piper_model() -> str:
    return os.getenv("VOICE_STREAM_TTS_MODEL", "").strip()


def piper_config() -> str:
    return os.getenv("VOICE_STREAM_TTS_CONFIG", "").strip()


def max_chars() -> int:
    try:
        return max(200, min(6000, int(os.getenv("VOICE_STREAM_MAX_CHARS", "4000"))))
    except ValueError:
        return 4000


def auth_header_name() -> str:
    return os.getenv("VOICE_STREAM_AUTH_HEADER", "X-PCS-Voice-Token").strip() or "X-PCS-Voice-Token"


def auth_token() -> str:
    return os.getenv("VOICE_STREAM_AUTH_TOKEN", "").strip()


def verify_auth(request: Request) -> None:
    expected = auth_token()
    if not expected:
        return
    received = request.headers.get(auth_header_name(), "").strip()
    if not hmac.compare_digest(received, expected):
        raise HTTPException(status_code=401, detail="invalid voice stream token")


class TTSRequest(BaseModel):
    text: str
    voice: str | None = None


@app.get("/health")
def health(request: Request) -> dict:
    verify_auth(request)
    bin_path = shutil.which(piper_bin()) or piper_bin()
    model = piper_model()
    return {
        "ok": env_bool("VOICE_STREAM_ENABLED", True)
        and bool(bin_path)
        and bool(model)
        and Path(model).exists(),
        "enabled": env_bool("VOICE_STREAM_ENABLED", True),
        "engine": "piper",
        "piper_bin": bin_path,
        "model_configured": bool(model),
        "model_exists": bool(model) and Path(model).exists(),
        "auth_required": bool(auth_token()),
    }


@app.post("/api/voice/tts")
def text_to_speech(request: Request, payload: TTSRequest) -> StreamingResponse:
    verify_auth(request)
    if not env_bool("VOICE_STREAM_ENABLED", True):
        raise HTTPException(status_code=503, detail="voice service disabled")

    text = (payload.text or "").strip()
    if not text:
        raise HTTPException(status_code=400, detail="text is required")
    if len(text) > max_chars():
        raise HTTPException(status_code=413, detail=f"text is longer than {max_chars()} chars")

    model = piper_model()
    if not model or not Path(model).exists():
        raise HTTPException(status_code=503, detail="VOICE_STREAM_TTS_MODEL is not configured or does not exist")

    output_dir = Path(tempfile.gettempdir()) / "pcs_voice_stream"
    output_dir.mkdir(parents=True, exist_ok=True)
    output_path = output_dir / f"{uuid.uuid4().hex}.wav"

    cmd = [piper_bin(), "--model", model, "--output_file", str(output_path)]
    config = piper_config()
    if config:
        cmd.extend(["--config", config])

    try:
        subprocess.run(
            cmd,
            input=text.encode("utf-8"),
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            timeout=float(os.getenv("VOICE_STREAM_TTS_TIMEOUT", "20")),
            check=True,
        )
    except FileNotFoundError as exc:
        raise HTTPException(status_code=503, detail=f"piper binary not found: {exc}") from exc
    except subprocess.TimeoutExpired as exc:
        raise HTTPException(status_code=504, detail="piper synthesis timeout") from exc
    except subprocess.CalledProcessError as exc:
        error = (exc.stderr or b"").decode("utf-8", errors="ignore").strip()
        raise HTTPException(status_code=502, detail=error or "piper synthesis failed") from exc

    if not output_path.exists() or output_path.stat().st_size <= 0:
        raise HTTPException(status_code=502, detail="empty audio generated")

    def stream_file() -> Iterator[bytes]:
        try:
            with output_path.open("rb") as audio:
                while True:
                    chunk = audio.read(64 * 1024)
                    if not chunk:
                        break
                    yield chunk
        finally:
            try:
                output_path.unlink(missing_ok=True)
            except OSError:
                pass

    return StreamingResponse(stream_file(), media_type="audio/wav")
