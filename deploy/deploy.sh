#!/bin/bash

# E-commerce AI Agent Deployment Script

set -e

echo "🛒 Deploying E-commerce AI Agent..."

# ==================== Configuration ====================
export COMPOSE_PROJECT_NAME=ecommerce-ai
export COMPOSE_FILE=docker-compose.yml

# ==================== Build ====================
echo "📦 Building Docker images..."
docker-compose build

# ==================== Start Services ====================
echo "🚀 Starting services..."
docker-compose up -d

# ==================== Wait for services ====================
echo "⏳ Waiting for services to be ready..."

# Wait for MySQL
echo "  - MySQL..."
sleep 10

# Wait for Redis
echo "  - Redis..."
sleep 2

# Wait for Kafka
echo "  - Kafka..."
sleep 5

# Wait for Milvus
echo "  - Milvus..."
sleep 10

# ==================== Check Status ====================
echo "📊 Checking service status..."
docker-compose ps

# ==================== Show URLs ====================
echo ""
echo "✅ Deployment complete!"
echo ""
echo "🌐 Access URLs:"
echo "  - Frontend: http://localhost"
echo "  - Go API:   http://localhost:8080"
echo "  - Python AI: http://localhost:8000"
echo ""
echo "📝 To check logs: docker-compose logs -f"
echo "📝 To stop: docker-compose down"
