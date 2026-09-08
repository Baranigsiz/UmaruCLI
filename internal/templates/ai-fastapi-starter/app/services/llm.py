import asyncio
import json
from typing import AsyncGenerator, List, Dict
import httpx
from app.core.config import settings

class LLMService:
    @staticmethod
    async def stream_chat(messages: List[Dict[str, str]]) -> AsyncGenerator[str, None]:
        """
        Unified chat streaming router that handles OpenAI, Ollama, and a mock fallback.
        Yields JSON chunks compatible with SSE format.
        """
        provider = settings.LLM_PROVIDER.lower()

        if provider == "openai" and settings.OPENAI_API_KEY:
            async for chunk in LLMService._stream_openai(messages):
                yield chunk
        elif provider == "ollama":
            async for chunk in LLMService._stream_ollama(messages):
                yield chunk
        else:
            async for chunk in LLMService._stream_mock(messages):
                yield chunk

    @staticmethod
    async def _stream_openai(messages: List[Dict[str, str]]) -> AsyncGenerator[str, None]:
        url = "https://api.openai.com/v1/chat/completions"
        headers = {
            "Authorization": f"Bearer {settings.OPENAI_API_KEY}",
            "Content-Type": "application/json",
        }
        payload = {
            "model": settings.OPENAI_MODEL,
            "messages": messages,
            "stream": True,
        }

        try:
            async with httpx.AsyncClient(timeout=60.0) as client:
                async with client.stream("POST", url, json=payload, headers=headers) as response:
                    async for line in response.aiter_lines():
                        if line.startswith("data: "):
                            data = line[6:].strip()
                            if data == "[DONE]":
                                break
                            try:
                                parsed = json.loads(data)
                                delta = parsed["choices"][0]["delta"].get("content", "")
                                if delta:
                                    yield json.dumps({"content": delta})
                            except json.JSONDecodeError:
                                continue
        except Exception as e:
            yield json.dumps({"error": f"OpenAI error: {str(e)}"})

    @staticmethod
    async def _stream_ollama(messages: List[Dict[str, str]]) -> AsyncGenerator[str, None]:
        url = f"{settings.OLLAMA_BASE_URL}/api/chat"
        payload = {
            "model": settings.OLLAMA_MODEL,
            "messages": messages,
            "stream": True,
        }

        try:
            async with httpx.AsyncClient(timeout=60.0) as client:
                async with client.stream("POST", url, json=payload) as response:
                    async for line in response.aiter_lines():
                        if line.strip():
                            try:
                                parsed = json.loads(line)
                                content = parsed.get("message", {}).get("content", "")
                                if content:
                                    yield json.dumps({"content": content})
                                if parsed.get("done", False):
                                    break
                            except json.JSONDecodeError:
                                continue
        except Exception:
            # Fallback to simulated response if Ollama is not running locally
            async for chunk in LLMService._stream_mock(messages):
                yield chunk

    @staticmethod
    async def _stream_mock(messages: List[Dict[str, str]]) -> AsyncGenerator[str, None]:
        """Simulation fallback when no API key or local Ollama is detected."""
        last_user_msg = messages[-1].get("content", "Hello") if messages else "Hello"
        mock_reply = f"⚡ Umaru AI Agent received: '{last_user_msg}'. Set OPENAI_API_KEY or start Ollama for live responses."
        
        words = mock_reply.split(" ")
        for word in words:
            await asyncio.sleep(0.04)
            yield json.dumps({"content": word + " "})
