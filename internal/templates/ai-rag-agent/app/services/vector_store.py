import logging
from typing import List, Dict, Any, Optional
import chromadb
from chromadb.config import Settings as ChromaSettings
from app.core.config import settings

logger = logging.getLogger(__name__)


class VectorStoreService:
    def __init__(self):
        self._client: Optional[chromadb.ClientAPI] = None
        self._collection = None

    def get_client(self) -> chromadb.ClientAPI:
        if self._client is None:
            if settings.CHROMA_HOST:
                logger.info(f"Connecting to remote ChromaDB at {settings.CHROMA_HOST}:{settings.CHROMA_PORT}")
                self._client = chromadb.HttpClient(
                    host=settings.CHROMA_HOST,
                    port=settings.CHROMA_PORT,
                    settings=ChromaSettings(anonymized_telemetry=False),
                )
            else:
                logger.info(f"Using local Persistent ChromaDB at {settings.CHROMA_PERSIST_DIR}")
                self._client = chromadb.PersistentClient(
                    path=settings.CHROMA_PERSIST_DIR,
                    settings=ChromaSettings(anonymized_telemetry=False),
                )
        return self._client

    def get_collection(self):
        if self._collection is None:
            client = self.get_client()
            self._collection = client.get_or_create_collection(
                name=settings.CHROMA_COLLECTION_NAME,
                metadata={"description": "UmaruCLI Production RAG Knowledge Base"}
            )
        return self._collection

    def add_documents(
        self,
        texts: List[str],
        metadatas: List[Dict[str, Any]],
        ids: List[str]
    ) -> int:
        collection = self.get_collection()
        collection.add(
            documents=texts,
            metadatas=metadatas,
            ids=ids
        )
        return len(ids)

    def query(self, query_text: str, n_results: int = 4) -> List[Dict[str, Any]]:
        collection = self.get_collection()
        results = collection.query(
            query_texts=[query_text],
            n_results=n_results
        )

        documents = results.get("documents", [[]])[0]
        metadatas = results.get("metadatas", [[]])[0]
        distances = results.get("distances", [[]])[0] if "distances" in results else []

        formatted = []
        for i, doc in enumerate(documents):
            formatted.append({
                "content": doc,
                "metadata": metadatas[i] if i < len(metadatas) else {},
                "distance": distances[i] if i < len(distances) else None,
            })
        return formatted

    def get_stats(self) -> Dict[str, Any]:
        collection = self.get_collection()
        return {
            "collection_name": settings.CHROMA_COLLECTION_NAME,
            "total_documents": collection.count(),
            "mode": "remote" if settings.CHROMA_HOST else "local_embedded",
        }


vector_store = VectorStoreService()
