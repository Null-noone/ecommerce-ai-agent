import os
from langchain_community.vectorstores import Milvus
from core.llm import get_embeddings


def get_vector_store():
    """Get Milvus vector store instance"""
    return Milvus(
        embedding_function=get_embeddings(),
        collection_name="product_embeddings",
        connection_args={
            "host": os.getenv("MILVUS_HOST", "milvus"),
            "port": "19530"
        },
        drop_old=False,
    )


def init_vector_store():
    """Initialize vector store with sample data (for testing)"""
    from core.llm import get_embeddings
    from langchain_community.docstore.in_memory import InMemoryDocstore
    from langchain.schema import Document
    import faiss
    
    # Sample product data for embedding
    sample_products = [
        {"id": 1, "text": "iPhone 15 Pro - Apple最新款智能手机，A17 Pro芯片，钛金属设计，适合科技爱好者", "name": "iPhone 15 Pro"},
        {"id": 2, "text": "MacBook Air M3 - 轻薄便携笔记本电脑，M3芯片，续航超长，适合办公族", "name": "MacBook Air M3"},
        {"id": 3, "text": "Dior 999 口红 - 经典正红色，滋润不干，适合送女生", "name": "Dior 999 口红"},
        {"id": 4, "text": "SK-II 神仙水 - 护肤精华水，改善肌肤，适合敏感肌", "name": "SK-II 神仙水"},
        {"id": 5, "text": "AirPods Pro 2 - 主动降噪无线耳机，适合音乐爱好者", "name": "AirPods Pro 2"},
    ]
    
    docs = [
        Document(
            page_content=p["text"],
            metadata={"product_id": p["id"], "name": p["name"]}
        ) for p in sample_products
    ]
    
    return Milvus.from_documents(
        docs,
        get_embeddings(),
        collection_name="product_embeddings",
        connection_args={
            "host": os.getenv("MILVUS_HOST", "milvus"),
            "port": "19530"
        },
        drop_old=True,
    )
