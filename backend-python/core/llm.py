"""
AI Agent Service - Core Module
Provides LLM and embedding services using Volcano Engine (豆包)
"""

import os
from typing import Optional
from langchain_openai import ChatOpenAI
from langchain_openai import OpenAIEmbeddings
from langchain.schema import BaseChatModel
from langchain.embeddings.base import Embeddings
from dotenv import load_dotenv

load_dotenv()


class LLMService:
    """LLM Service using Doubao (Volcano Engine)"""
    
    def __init__(
        self,
        model: str = "doubao-pro-32k",
        temperature: float = 0.7,
        streaming: bool = True,
        base_url: Optional[str] = None,
        api_key: Optional[str] = None
    ):
        self.model = model
        self.temperature = temperature
        self.streaming = streaming
        self.base_url = base_url or os.getenv("VOLC_ENDPOINT", "https://ark.cn-beijing.volces.com/api/v3")
        self.api_key = api_key or os.getenv("VOLC_API_KEY")
        self._llm: Optional[BaseChatModel] = None
    
    @property
    def llm(self) -> BaseChatModel:
        """Lazy initialization of LLM"""
        if self._llm is None:
            self._llm = ChatOpenAI(
                model=self.model,
                openai_api_key=self.api_key,
                openai_api_base=self.base_url,
                temperature=self.temperature,
                streaming=self.streaming,
            )
        return self._llm
    
    def chat(self, prompt: str) -> str:
        """Simple chat without streaming"""
        return self.llm.invoke(prompt).content
    
    async def chat_stream(self, prompt: str):
        """Streaming chat"""
        async for chunk in self.llm.astream(prompt):
            if chunk.content:
                yield chunk.content
    
    def chat_with_messages(self, messages: list) -> str:
        """Chat with message history"""
        return self.llm.invoke(messages).content


class EmbeddingService:
    """Embedding Service using Doubao"""
    
    def __init__(
        self,
        model: str = "doubao-embedding-text-256",
        base_url: Optional[str] = None,
        api_key: Optional[str] = None
    ):
        self.model = model
        self.base_url = base_url or os.getenv("VOLC_ENDPOINT", "https://ark.cn-beijing.volces.com/api/v3")
        self.api_key = api_key or os.getenv("VOLC_API_KEY")
        self._embeddings: Optional[Embeddings] = None
    
    @property
    def embeddings(self) -> Embeddings:
        """Lazy initialization of embeddings"""
        if self._embeddings is None:
            self._embeddings = OpenAIEmbeddings(
                openai_api_key=self.api_key,
                openai_api_base=self.base_url,
                model=self.model,
            )
        return self._embeddings
    
    def embed_text(self, text: str) -> list:
        """Embed a single text"""
        return self.embeddings.embed_query(text)
    
    def embed_texts(self, texts: list) -> list:
        """Embed multiple texts"""
        return self.embeddings.embed_documents(texts)


# Global instances
_llm_service: Optional[LLMService] = None
_embedding_service: Optional[EmbeddingService] = None


def get_llm_service() -> LLMService:
    """Get global LLM service instance"""
    global _llm_service
    if _llm_service is None:
        _llm_service = LLMService()
    return _llm_service


def get_embedding_service() -> EmbeddingService:
    """Get global embedding service instance"""
    global _embedding_service
    if _embedding_service is None:
        _embedding_service = EmbeddingService()
    return _embedding_service
