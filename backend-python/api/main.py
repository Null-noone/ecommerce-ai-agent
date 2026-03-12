import os
import asyncio
from typing import Optional
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from sse_starlette.sse import EventSourceResponse

app = FastAPI(title="E-commerce AI Agent")


class SemanticSearchRequest(BaseModel):
    query: str
    top_k: int = 5


class ChatRequest(BaseModel):
    session_id: str
    query: str
    context: Optional[list] = None


@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {"status": "healthy", "service": "python-ai-service"}


@app.post("/agent/semantic_search")
async def semantic_search(req: SemanticSearchRequest):
    """Semantic search for products using vector similarity"""
    try:
        from core.vector_store import get_vector_store
        
        vector_store = get_vector_store()
        
        # Search similar products
        results = vector_store.similarity_search_with_score(
            query=req.query,
            k=req.top_k
        )
        
        product_ids = []
        for doc, score in results:
            product_ids.append(doc.metadata.get("product_id"))
        
        return {"product_ids": product_ids, "query": req.query}
        
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/agent/chat")
async def chat(req: ChatRequest):
    """AI chatbot with product context (streaming)"""
    async def generate():
        try:
            from core.llm import get_llm
            from core.vector_store import get_vector_store
            
            # Get product context
            vector_store = get_vector_store()
            docs = vector_store.similarity_search_with_score(
                query=req.query,
                k=3
            )
            
            # Build context
            context = "\n".join([
                f"- {doc.metadata.get('name')}: {doc.page_content}"
                for doc, _ in docs
            ])
            
            # Build prompt
            prompt = f"""你是一个专业的电商导购助手。请根据以下【商品信息】回答用户问题。
如果不知道，请说"抱歉，我不太清楚"，不要编造。

【商品信息】:
{context}

用户问题: {req.query}

回答:"""
            
            # Stream response
            llm = get_llm()
            
            async for chunk in llm.astream(prompt):
                if chunk.content:
                    yield {"event": "message", "data": chunk.content}
                    
        except Exception as e:
            yield {"event": "error", "data": str(e)}
    
    return EventSourceResponse(generate())


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
