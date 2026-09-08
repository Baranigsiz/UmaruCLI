from typing import List, Dict, Optional
from fastapi import APIRouter
from pydantic import BaseModel, Field
from sse_starlette.sse import EventSourceResponse
from app.services.llm import LLMService

router = APIRouter(prefix="/chat", tags=["AI Chat"])

class Message(BaseModel):
    role: str = Field(..., description="Role: 'user', 'assistant', or 'system'")
    content: str = Field(..., description="Message text content")

class ChatRequest(BaseModel):
    messages: List[Message]
    temperature: Optional[float] = Field(0.7, ge=0.0, le=2.0)

class ChatResponse(BaseModel):
    reply: str
    provider: str

@router.post("/stream", summary="Stream chat completions in real-time via Server-Sent Events (SSE)")
async def chat_stream(request: ChatRequest):
    """
    Real-time streaming endpoint. Connect using EventSource or fetch with streaming reader.
    """
    msg_dicts = [{"role": m.role, "content": m.content} for m in request.messages]
    
    async def event_generator():
        async for chunk in LLMService.stream_chat(msg_dicts):
            yield {"data": chunk}

    return EventSourceResponse(event_generator())

@router.post("", response_model=ChatResponse, summary="Non-streaming chat completion")
async def chat_sync(request: ChatRequest):
    """
    Standard HTTP JSON request/response endpoint.
    """
    msg_dicts = [{"role": m.role, "content": m.content} for m in request.messages]
    full_text = []
    
    async for chunk_json in LLMService.stream_chat(msg_dicts):
        import json
        try:
            data = json.loads(chunk_json)
            if "content" in data:
                full_text.append(data["content"])
        except Exception:
            continue

    return ChatResponse(
        reply="".join(full_text),
        provider="configured-llm"
    )
