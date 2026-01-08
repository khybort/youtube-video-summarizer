from fastapi import FastAPI, UploadFile, File, HTTPException, Request
from faster_whisper import WhisperModel
import os
import tempfile
from typing import Optional
import requests
import asyncio
from concurrent.futures import ThreadPoolExecutor, TimeoutError as FutureTimeoutError
import threading

app = FastAPI(title="Whisper Transcription Service")

model = None
model_size = os.getenv("WHISPER_MODEL", "small")  # Default changed from base to small for better accuracy
# Limit concurrent transcriptions to prevent CPU overload
executor = ThreadPoolExecutor(max_workers=1)  # Only one transcription at a time
# Maximum transcription time: 10 minutes (600 seconds)
MAX_TRANSCRIPTION_TIME = 600
# Lock to track if transcription is in progress
transcription_lock = threading.Lock()
transcription_in_progress = False

@app.on_event("startup")
async def load_model():
    global model
    print(f"Loading Whisper model: {model_size}")
    # Optimized for better accuracy while maintaining reasonable speed
    # Use int8 for speed, but increase threads and model size for better results
    model = WhisperModel(
        model_size, 
        device="cpu", 
        compute_type="int8",  # Keep int8 for speed
        cpu_threads=4,  # Increased from 2 to 4 for faster processing
        num_workers=1   # Single worker for transcription
    )
    print("Model loaded successfully")

@app.get("/health")
async def health():
    return {"status": "healthy", "model": model_size}

@app.post("/transcribe")
async def transcribe(
    request: Request,
    file: UploadFile = File(...),
    language: Optional[str] = None,
    task: str = "transcribe"
):
    global transcription_in_progress
    
    if model is None:
        raise HTTPException(status_code=503, detail="Model not loaded")
    
    # Check if another transcription is in progress
    with transcription_lock:
        if transcription_in_progress:
            raise HTTPException(
                status_code=503,
                detail="Service busy: Another transcription is in progress. Please try again later."
            )
        transcription_in_progress = True
    
    # Save uploaded file temporarily
    with tempfile.NamedTemporaryFile(delete=False, suffix=os.path.splitext(file.filename)[1]) as tmp_file:
        content = await file.read()
        tmp_file.write(content)
        tmp_path = tmp_file.name
    
    try:
        # Run transcription in thread pool with timeout to prevent hanging
        loop = asyncio.get_event_loop()
        
        def transcribe_with_timeout():
            return model.transcribe(
                tmp_path,
                language=language,
                task=task,
                word_timestamps=True,
                beam_size=2,  # Increased from 1 to 2 for better accuracy
                best_of=2,   # Increased from 1 to 2 for better quality
            )
        
        try:
            # Use asyncio.wait_for to add timeout
            # This will raise TimeoutError if transcription takes too long
            segments, info = await asyncio.wait_for(
                loop.run_in_executor(executor, transcribe_with_timeout),
                timeout=MAX_TRANSCRIPTION_TIME
            )
        except asyncio.TimeoutError:
            with transcription_lock:
                transcription_in_progress = False
            raise HTTPException(
                status_code=504,
                detail=f"Transcription timeout: exceeded {MAX_TRANSCRIPTION_TIME} seconds"
            )
        finally:
            # Always release the lock, even on error (if not already released)
            with transcription_lock:
                if transcription_in_progress:
                    transcription_in_progress = False
        
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
    except HTTPException:
        # Re-raise HTTP exceptions
        with transcription_lock:
            transcription_in_progress = False
        raise
    except Exception as e:
        # Release lock on any other error
        with transcription_lock:
            transcription_in_progress = False
        raise HTTPException(status_code=500, detail=f"Transcription failed: {str(e)}")
    finally:
        # Cleanup
        if os.path.exists(tmp_path):
            os.unlink(tmp_path)

