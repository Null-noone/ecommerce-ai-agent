"""
Vector Store Module - Milvus Integration
Provides semantic search for product queries
"""

import os
from typing import Optional, List, Tuple
from langchain.schema import Document
from langchain_community.vectorstores import Milvus
from langchain_community.vectorstores.base import VectorStore
from core.llm import get_embedding_service


class VectorStoreService:
    """Vector Store Service using Milvus"""
    
    def __init__(
        self,
        collection_name: str = "product_embeddings",
        host: str = None,
        port: str = "19530",
        connection_args: dict = None
    ):
        self.collection_name = collection_name
        self.host = host or os.getenv("MILVUS_HOST", "milvus")
        self.port = port
        self.connection_args = connection_args or {
            "host": self.host,
            "port": self.port
        }
        self._vector_store: Optional[VectorStore] = None
    
    @property
    def vector_store(self) -> VectorStore:
        """Lazy initialization of vector store"""
        if self._vector_store is None:
            embedding_service = get_embedding_service()
            self._vector_store = Milvus(
                embedding_function=embedding_service.embeddings,
                collection_name=self.collection_name,
                connection_args=self.connection_args,
                drop_old=False,
            )
        return self._vector_store
    
    def search(
        self, 
        query: str, 
        top_k: int = 5,
        filter: str = None
    ) -> List[Tuple[Document, float]]:
        """Search for similar products"""
        return self.vector_store.similarity_search_with_score(
            query=query,
            k=top_k,
            filter=filter
        )
    
    def search_by_vector(
        self,
        embedding: List[float],
        top_k: int = 5
    ) -> List[Tuple[Document, float]]:
        """Search by vector embedding"""
        return self.vector_store.similarity_search_by_vector_with_score(
            embedding=embedding,
            k=top_k
        )
    
    def add_documents(self, documents: List[Document]) -> List[str]:
        """Add documents to vector store"""
        return self.vector_store.add_documents(documents)
    
    def delete_collection(self):
        """Delete the collection"""
        self.vector_store.delete_collection()
        self._vector_store = None


class ProductVectorStore:
    """Specialized vector store for products"""
    
    def __init__(self):
        self.service = VectorStoreService(
            collection_name="product_embeddings"
        )
    
    def build_product_context(self, query: str, top_k: int = 5) -> str:
        """Build context string from product search results"""
        results = self.service.search(query, top_k=top_k)
        
        context_parts = []
        for doc, score in results:
            name = doc.metadata.get("name", "Unknown")
            description = doc.page_content
            context_parts.append(f"- {name}: {description}")
        
        return "\n".join(context_parts)
    
    def get_product_ids(self, query: str, top_k: int = 5) -> List[int]:
        """Get product IDs from search results"""
        results = self.service.search(query, top_k=top_k)
        
        product_ids = []
        for doc, _ in results:
            product_id = doc.metadata.get("product_id")
            if product_id:
                product_ids.append(int(product_id))
        
        return product_ids


# Global instance
_product_vector_store: Optional[ProductVectorStore] = None


def get_product_vector_store() -> ProductVectorStore:
    """Get global product vector store instance"""
    global _product_vector_store
    if _product_vector_store is None:
        _product_vector_store = ProductVectorStore()
    return _product_vector_store


def init_vector_store_with_sample_data():
    """Initialize vector store with sample product data"""
    from langchain.schema import Document
    
    sample_products = [
        {
            "id": 1,
            "text": "iPhone 15 Pro - Apple最新款智能手机，A17 Pro芯片，钛金属设计，适合科技爱好者",
            "name": "iPhone 15 Pro"
        },
        {
            "id": 2,
            "text": "MacBook Air M3 - 轻薄便携笔记本电脑，M3芯片，续航超长，适合办公族",
            "name": "MacBook Air M3"
        },
        {
            "id": 3,
            "text": "Dior 999 口红 - 经典正红色，滋润不干，适合送女生",
            "name": "Dior 999 口红"
        },
        {
            "id": 4,
            "text": "SK-II 神仙水 - 护肤精华水，改善肌肤，适合敏感肌",
            "name": "SK-II 神仙水"
        },
        {
            "id": 5,
            "text": "AirPods Pro 2 - 主动降噪无线耳机，适合音乐爱好者",
            "name": "AirPods Pro 2"
        },
    ]
    
    documents = [
        Document(
            page_content=p["text"],
            metadata={"product_id": p["id"], "name": p["name"]}
        )
        for p in sample_products
    ]
    
    service = VectorStoreService(
        collection_name="product_embeddings",
        connection_args={
            "host": os.getenv("MILVUS_HOST", "milvus"),
            "port": "19530"
        }
    )
    
    # Delete old collection and create new one
    service.delete_collection()
    service.add_documents(documents)
    
    return "Vector store initialized with sample data"
