import os
from langchain_openai import ChatOpenAI
from langchain_openai import OpenAIEmbeddings
from dotenv import load_dotenv

load_dotenv()


def get_llm():
    """Get ChatOpenAI instance for Doubao (Volcengine)"""
    return ChatOpenAI(
        model="doubao-pro-32k",
        openai_api_key=os.getenv("VOLC_API_KEY"),
        openai_api_base=os.getenv("VOLC_ENDPOINT", "https://ark.cn-beijing.volces.com/api/v3"),
        temperature=0.7,
        streaming=True,
    )


def get_embeddings():
    """Get OpenAIEmbeddings instance for Doubao"""
    return OpenAIEmbeddings(
        openai_api_key=os.getenv("VOLC_API_KEY"),
        openai_api_base=os.getenv("VOLC_ENDPOINT", "https://ark.cn-beijing.volces.com/api/v3"),
        model="doubao-embedding-text-256",
    )
