from fastapi import APIRouter, HTTPException, status
from pydantic import BaseModel, Field
from typing import Optional
from sse_starlette.sse import EventSourceResponse

from app.services.vector_store import vector_store
from app.services.rag_service import rag_service

router = APIRouter()


class IngestRequest(BaseModel):
    title: str = Field(..., example="UmaruCLI Overview")
    content: str = Field(..., example="UmaruCLI is a lightning-fast scaffolding tool for Go, TypeScript, and Python.")
    source: Optional[str] = Field("manual", example="docs/readme.md")


class QueryRequest(BaseModel):
    query: str = Field(..., example="What is UmaruCLI?")
    top_k: Optional[int] = Field(3, ge=1, le=10)


@router.get("/health", tags=["Health"])
async def health_check():
    return {
        "status": "healthy",
        "service": "Production RAG Agent",
        "vector_store": vector_store.get_stats()
    }


@router.post("/api/v1/rag/ingest", tags=["RAG"], status_code=status.HTTP_201_CREATED)
async def ingest_document(req: IngestRequest):
    try:
        result = rag_service.ingest_document(
            title=req.title,
            content=req.content,
            source=req.source
        )
        return {"status": "success", "data": result}
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Failed to ingest document: {str(e)}"
        )


@router.post("/api/v1/rag/query", tags=["RAG"])
async def query_knowledge_base(req: QueryRequest):
    try:
        result = await rag_service.generate_rag_answer(
            query=req.query,
            top_k=req.top_k
        )
        return result
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"RAG query failed: {str(e)}"
        )


@router.post("/api/v1/rag/chat-stream", tags=["RAG"])
async def stream_rag_chat(req: QueryRequest):
    """Streams RAG tokens directly as Server-Sent Events (SSE)."""
    return EventSourceResponse(
        rag_service.stream_rag_answer(query=req.query, top_k=req.top_k),
        media_type="text/event-stream"
    )


@router.get("/api/v1/rag/stats", tags=["RAG"])
async def get_rag_stats():
    return vector_store.get_stats()
