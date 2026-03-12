"""
Database module for Python AI Service
Provides MySQL connection and product data access
"""

import os
from typing import Optional, List, Dict, Any
from contextlib import contextmanager

from sqlalchemy import create_engine, text
from sqlalchemy.orm import sessionmaker, Session
from sqlalchemy.pool import QueuePool


class Database:
    """Database connection manager"""
    
    def __init__(
        self,
        host: str = None,
        port: int = 3306,
        user: str = None,
        password: str = None,
        database: str = "ecommerce_db"
    ):
        self.host = host or os.getenv("MYSQL_HOST", "mysql")
        self.port = port
        self.user = user or os.getenv("MYSQL_USER", "ecom_user")
        self.password = password or os.getenv("MYSQL_PASSWORD", "EcomPass456!")
        self.database = database
        
        self.url = f"mysql+pymysql://{self.user}:{self.password}@{self.host}:{self.port}/{self.database}"
        self.engine = None
        self.SessionLocal = None
    
    def init(self):
        """Initialize database connection"""
        self.engine = create_engine(
            self.url,
            poolclass=QueuePool,
            pool_size=5,
            max_overflow=10,
            pool_pre_ping=True,
            echo=False
        )
        self.SessionLocal = sessionmaker(
            autocommit=False,
            autoflush=False,
            bind=self.engine
        )
    
    @contextmanager
    def get_session(self) -> Session:
        """Get database session"""
        if self.SessionLocal is None:
            self.init()
        
        session = self.SessionLocal()
        try:
            yield session
            session.commit()
        except Exception:
            session.rollback()
            raise
        finally:
            session.close()
    
    def execute(self, query: str, params: dict = None) -> List[Dict[str, Any]]:
        """Execute raw SQL query"""
        if self.engine is None:
            self.init()
        
        with self.engine.connect() as conn:
            result = conn.execute(text(query), params or {})
            return [dict(row._mapping) for row in result]


class ProductRepository:
    """Product data access"""
    
    def __init__(self, db: Database):
        self.db = db
    
    def get_by_id(self, product_id: int) -> Optional[Dict[str, Any]]:
        """Get product by ID"""
        query = "SELECT * FROM products WHERE id = :id"
        results = self.db.execute(query, {"id": product_id})
        return results[0] if results else None
    
    def get_by_ids(self, product_ids: List[int]) -> List[Dict[str, Any]]:
        """Get products by ID list"""
        if not product_ids:
            return []
        
        placeholders = ",".join([":id" + str(i) for i in range(len(product_ids))])
        params = {f"id{i}": pid for i, pid in enumerate(product_ids)}
        query = f"SELECT * FROM products WHERE id IN ({placeholders})"
        
        return self.db.execute(query, params)
    
    def search(self, keyword: str, limit: int = 10) -> List[Dict[str, Any]]:
        """Basic keyword search"""
        query = """
            SELECT * FROM products 
            WHERE name LIKE :keyword OR description LIKE :keyword
            LIMIT :limit
        """
        keyword_pattern = f"%{keyword}%"
        return self.db.execute(query, {"keyword": keyword_pattern, "limit": limit})
    
    def get_all(self, limit: int = 100, offset: int = 0) -> List[Dict[str, Any]]:
        """Get all products"""
        query = "SELECT * FROM products LIMIT :limit OFFSET :offset"
        return self.db.execute(query, {"limit": limit, "offset": offset})
    
    def get_categories(self) -> List[Dict[str, Any]]:
        """Get all categories"""
        query = "SELECT * FROM categories"
        return self.db.execute(query)


# Global database instance
_db: Optional[Database] = None


def get_database() -> Database:
    """Get global database instance"""
    global _db
    if _db is None:
        _db = Database()
        _db.init()
    return _db


def get_product_repository() -> ProductRepository:
    """Get product repository"""
    return ProductRepository(get_database())
