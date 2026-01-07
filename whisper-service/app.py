from fastapi import FastAPI, UploadFile, File, HTTPException
from faster_whisper import WhisperModel
import os
import tempfile
from typing import Optional
import requests

app = FastAPI(title="Whisper Transcription Service")

model = None
model_size = os.getenv("WHISPER_MODEL", "base")

@app.on_event("startup")
async def load_model():
    global model
    print(f"Loading Whisper model: {model_size}")
    model = WhisperModel(model_size, device="cpu", compute_type="int8")
    print("Model loaded successfully")

@app.get("/health")
async def health():
    return {"status": "healthy", "model": model_size}

@app.post("/transcribe")
async def transcribe(
    file: UploadFile = File(...),
    language: Optional[str] = None,
    task: str = "transcribe"
):
    if model is None:
        raise HTTPException(status_code=503, detail="Model not loaded")
    
    # Save uploaded file temporarily
    with tempfile.NamedTemporaryFile(delete=False, suffix=os.path.splitext(file.filename)[1]) as tmp_file:
        content = await file.read()
        tmp_file.write(content)
        tmp_path = tmp_file.name
    
    try:
        # Transcribe
        segments, info = model.transcribe(
            tmp_path,
            language=language,
            task=task,
            word_timestamps=True
        )
        
        # Format response
        transcript_segments = []
        full_text = []
        
        for segment in segments:
            segment_data = {
                "start": segment.start,
                "end": segment.end,
                "text": segment.text
            }
            transcript_segments.append(segment_data)
            full_text.append(segment.text)
        
        return {
            "text": " ".join(full_text),
            "language": info.language,
            "duration": info.duration,
            "segments": transcript_segments
        }
    finally:
        # Cleanup
        if os.path.exists(tmp_path):
            os.unlink(tmp_path)

