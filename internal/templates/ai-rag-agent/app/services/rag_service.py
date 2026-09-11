import uuid
import json
import logging
import asyncio
from typing import List, Dict, Any, AsyncGenerator
import httpx

from app.core.config import settings
from app.services.vector_store import vector_store

logger = logging.getLogger(__name__)


def chunk_text(text: str, chunk_size: int = 500, chunk_overlap: int = 50) -> List[str]:
    """Simple recursive-like text chunker without heavy dependencies."""
    if not text:
        return []
    
    chunks = []
    start = 0
    while start < len(text):
        end = start + chunk_size
        chunk = text[start:end]
        chunks.append(chunk.strip())
        start += chunk_size - chunk_overlap
    return [c for c in chunks if c]


class RAGService:
    @staticmethod
    def ingest_document(title: str, content: str, source: str = "manual") -> Dict[str, Any]:
        chunks = chunk_text(content)
        if not chunks:
            raise ValueError("Content is empty or cannot be chunked")

        ids = [f"{uuid.uuid4().hex[:12]}_{i}" for i in range(len(chunks))]
        metadatas = [
            {"title": title, "source": source, "chunk_index": i, "total_chunks": len(chunks)}
            for i in range(len(chunks))
        ]

        vector_store.add_documents(texts=chunks, metadatas=metadatas, ids=ids)

        return {
            "title": title,
            "chunks_created": len(chunks),
            "source": source,
            "status": "indexed",
        }

    @staticmethod
    async def generate_rag_answer(query: str, top_k: int = 3) -> Dict[str, Any]:
        """Retrieves relevant context and generates an augmented answer."""
        retrieved_docs = vector_store.query(query, n_results=top_k)

        context_texts = [f"[{i+1}] {doc['content']}" for i, doc in enumerate(retrieved_docs)]
        context_block = "\n\n".join(context_texts) if context_texts else "No relevant context found."

        system_prompt = (
            "You are a helpful and precise AI Assistant. Use the provided context below to answer the user's question. "
            "If the answer cannot be found in the context, state that you don't know based on the provided documents.\n\n"
            f"Context:\n{context_block}"
        )

        # Check if OpenAI API Key is configured
        if settings.OPENAI_API_KEY and settings.OPENAI_API_KEY != "your-openai-api-key-here":
            try:
                async with httpx.AsyncClient(timeout=30.0) as client:
                    resp = await client.post(
                        "https://api.openai.com/v1/chat/completions",
                        headers={
                            "Authorization": f"Bearer {settings.OPENAI_API_KEY}",
                            "Content-Type": "application/json",
                        },
                        json={
                            "model": settings.LLM_MODEL,
                            "messages": [
                                {"role": "system", "content": system_prompt},
                                {"role": "user", "content": query},
                            ],
                            "temperature": 0.2,
                        },
                    )
                    resp.raise_for_status()
                    data = resp.json()
                    answer = data["choices"][0]["message"]["content"]
            except Exception as e:
                logger.error(f"OpenAI API call failed: {e}")
                answer = f"Error calling OpenAI API: {str(e)}"
        else:
            # Informative fallback simulation mode
            answer = (
                f"[SIMULATION MODE - Set OPENAI_API_KEY in .env for live LLM responses]\n\n"
                f"Retrieved {len(retrieved_docs)} relevant context pieces for query: '{query}'.\n\n"
                f"Top context summary:\n{context_block[:400]}..."
            )

        citations = [
            {
                "title": doc.get("metadata", {}).get("title", "Unknown"),
                "source": doc.get("metadata", {}).get("source", "Unknown"),
                "chunk_snippet": doc["content"][:150] + "...",
            }
            for doc in retrieved_docs
        ]

        return {
            "query": query,
            "answer": answer,
            "citations": citations,
            "context_count": len(retrieved_docs),
        }

    @staticmethod
    async def stream_rag_answer(query: str, top_k: int = 3) -> AsyncGenerator[str, None]:
        """Streams answer tokens as Server-Sent Events (SSE)."""
        retrieved_docs = vector_store.query(query, n_results=top_k)
        context_texts = [doc['content'] for doc in retrieved_docs]
        context_block = "\n".join(context_texts)

        # Initial event: sending citations
        yield json.dumps({
            "event": "citations",
            "count": len(retrieved_docs),
            "sources": [d.get("metadata", {}).get("source", "doc") for d in retrieved_docs]
        })

        if settings.OPENAI_API_KEY and settings.OPENAI_API_KEY != "your-openai-api-key-here":
            try:
                async with httpx.AsyncClient(timeout=60.0) as client:
                    async with client.stream(
                        "POST",
                        "https://api.openai.com/v1/chat/completions",
                        headers={
                            "Authorization": f"Bearer {settings.OPENAI_API_KEY}",
                            "Content-Type": "application/json",
                        },
                        json={
                            "model": settings.LLM_MODEL,
                            "messages": [
                                {
                                    "role": "system",
                                    "content": f"Answer based on context:\n{context_block}",
                                },
                                {"role": "user", "content": query},
                            ],
                            "stream": True,
                        },
                    ) as response:
                        async for line in response.aiter_lines():
                            if line.startswith("data: ") and line != "data: [DONE]":
                                try:
                                    payload = json.loads(line[6:])
                                    delta = payload["choices"][0]["delta"].get("content", "")
                                    if delta:
                                        yield json.dumps({"event": "token", "token": delta})
                                except Exception:
                                    pass
            except Exception as e:
                yield json.dumps({"event": "error", "message": str(e)})
        else:
            simulated_text = (
                f"Based on your knowledge base containing {len(retrieved_docs)} matching documents, "
                f"here is the synthesized answer for: '{query}'."
            )
            for word in simulated_text.split(" "):
                yield json.dumps({"event": "token", "token": word + " "})
                await asyncio.sleep(0.04)

        yield json.dumps({"event": "done"})


rag_service = RAGService()
