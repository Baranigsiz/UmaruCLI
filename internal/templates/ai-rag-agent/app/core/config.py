from pydantic_settings import BaseSettings
from typing import Optional


class Settings(BaseSettings):
    PROJECT_NAME: str = "Production RAG Agent"
    VERSION: str = "0.1.0"
    APP_PORT: int = 8000
    APP_DEBUG: bool = True

    # OpenAI / LLM
    OPENAI_API_KEY: Optional[str] = None
    LLM_MODEL: str = "gpt-4o-mini"
    EMBEDDING_MODEL: str = "text-embedding-3-small"

    # Ollama / Local LLM
    OLLAMA_BASE_URL: str = "http://localhost:11434"

    # ChromaDB
    CHROMA_HOST: Optional[str] = None
    CHROMA_PORT: int = 8001
    CHROMA_COLLECTION_NAME: str = "rag_knowledge_base"
    CHROMA_PERSIST_DIR: str = "./chroma_data"

    class Config:
        env_file = ".env"
        extra = "ignore"


settings = Settings()
