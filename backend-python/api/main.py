"""
E-commerce AI Agent API
FastAPI application for semantic search and AI chat
"""

import os
import logging
from contextlib import asynccontextmanager
from typing import Optional

from fastapi import FastAPI, HTTPException, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from pydantic import BaseModel
from sse_starlette.sse import EventSourceResponse

from core.llm import get_llm_service, get_embedding_service
from core.vector_store import get_product_vector_store, init_vector_store_with_sample_data

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


# ==================== Lifespan Events ====================

@asynccontextmanager
async def lifespan(app: FastAPI):
    """Startup and shutdown events"""
    # Startup
    logger.info("Starting AI Agent Service...")
    
    # Initialize vector store with sample data
    try:
        init_vector_store_with_sample_data()
        logger.info("Vector store initialized")
    except Exception as e:
        logger.warning(f"Failed to initialize vector store: {e}")
    
    yield
    
    # Shutdown
    logger.info("Shutting down AI Agent Service...")


# ==================== FastAPI App ====================

app = FastAPI(
    title="E-commerce AI Agent",
    description="AI-powered e-commerce assistant with semantic search",
    version="1.0.0",
    lifespan=lifespan,
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


# ==================== Request Models ====================

class SemanticSearchRequest(BaseModel):
    query: str
    top_k: int = 5


class ChatRequest(BaseModel):
    session_id: str
    query: str
    context: Optional[list] = None


# ==================== Health Check ====================

@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "service": "python-ai-service",
        "version": "1.0.0"
    }


# ==================== Semantic Search ====================

@app.post("/agent/semantic_search")
async def semantic_search(req: SemanticSearchRequest):
    """
    Semantic search for products using vector similarity
    
    - **query**: Natural language search query
    - **top_k**: Number of results to return
    """
    try:
        vector_store = get_product_vector_store()
        
        # Get product IDs from vector search
        product_ids = vector_store.get_product_ids(
            query=req.query,
            top_k=req.top_k
        )
        
        logger.info(f"Semantic search: '{req.query}' -> {len(product_ids)} results")
        
        return {
            "product_ids": product_ids,
            "query": req.query,
            "total": len(product_ids)
        }
        
    except Exception as e:
        logger.error(f"Semantic search error: {e}")
        raise HTTPException(status_code=500, detail=str(e))


# ==================== AI Chat ====================

@app.post("/agent/chat")
async def chat(req: ChatRequest):
    """
    AI chatbot with product context (streaming)
    
    - **session_id**: Chat session ID
    - **query**: User question
    - **context**: Optional additional context
    """
    async def generate():
        try:
            llm_service = get_llm_service()
            vector_store = get_product_vector_store()
            
            # Get product context
            context = vector_store.build_product_context(req.query, top_k=3)
            
            # Build prompt
            prompt = f"""你是一个专业的电商导购助手。请根据以下【商品信息】回答用户问题。
如果不知道，请说"抱歉，我不太清楚"，不要编造。

【商品信息】:
{context}

用户问题: {req.query}

回答:"""
            
            # Stream response
            async for chunk in llm_service.chat_stream(prompt):
                if chunk:
                    yield {"event": "message", "data": chunk}
                    
        except Exception as e:
            logger.error(f"Chat error: {e}")
            yield {"event": "error", "data": str(e)}
    
    return EventSourceResponse(generate())


# ==================== Non-streaming Chat ====================

@app.post("/agent/chat/simple")
async def chat_simple(req: ChatRequest):
    """Simple non-streaming chat"""
    try:
        llm_service = get_llm_service()
        vector_store = get_product_vector_store()
        
        # Get product context
        context = vector_store.build_product_context(req.query, top_k=3)
        
        # Build prompt
        prompt = f"""你是一个专业的电商导购助手。请根据以下【商品信息】回答用户问题。
如果不知道，请说"抱歉，我不太清楚"，不要编造。

【商品信息】:
{context}

用户问题: {req.query}

回答:"""
        
        # Get response
        response = llm_service.chat(prompt)
        
        return {
            "response": response,
            "session_id": req.session_id
        }
        
    except Exception as e:
        logger.error(f"Chat error: {e}")
        raise HTTPException(status_code=500, detail=str(e))


# ==================== Embedding ====================

@app.post("/agent/embed")
async def embed_text(request: Request):
    """Get text embedding"""
    try:
        body = await request.json()
        text = body.get("text", "")
        
        if not text:
            raise HTTPException(status_code=400, detail="Text is required")
        
        embedding_service = get_embedding_service()
        embedding = embedding_service.embed_text(text)
        
        return {
            "embedding": embedding,
            "model": "doubao-embedding-text-256"
        }
        
    except Exception as e:
        logger.error(f"Embedding error: {e}")
        raise HTTPException(status_code=500, detail=str(e))


# ==================== Error Handlers ====================

@app.exception_handler(Exception)
async def global_exception_handler(request: Request, exc: Exception):
    """Global exception handler"""
    logger.error(f"Unhandled exception: {exc}")
    return JSONResponse(
        status_code=500,
        content={
            "error": "Internal server error",
            "detail": str(exc)
        }
    )


# ==================== Main ====================

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        app,
        host="0.0.0.0",
        port=8000,
        log_level="info"
    )
