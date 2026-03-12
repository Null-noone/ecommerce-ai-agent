# 🛒 云原生电商平台 + AI Agent

> 基于 Go + Python + Vue 的云原生电商系统，深度集成 AI 智能导购

## 🏗️ 系统架构

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Vue 3    │────▶│  Go Gin    │────▶│   MySQL    │
│  Frontend  │     │  Backend   │     │  Database  │
└─────────────┘     └──────┬──────┘     └─────────────┘
                          │
                    ┌─────┴─────┐
                    ▼           ▼
              ┌─────────┐ ┌─────────┐
              │  Kafka  │ │  Redis  │
              └─────────┘ └─────────┘
                    │           │
                    ▼           ▼
              ┌─────────────────────────────────┐
              │      Python FastAPI + LangChain │
              │           (AI Agent)             │
              └──────────────┬──────────────────┘
                             │
                             ▼
              ┌───────────────────────────────┐
              │      Milvus (向量数据库)       │
              └───────────────────────────────┘
```

## 🚀 快速开始

### 前置要求

- Docker & Docker Compose
- 火山引擎 API Key（或替换为 OpenAI）

### 1. 克隆项目

```bash
git clone https://github.com/Null-noone/ecommerce-ai-agent.git
cd ecommerce-ai-agent
```

### 2. 配置环境变量

```bash
cp deploy/.env.example deploy/.env
# 编辑 deploy/.env，填入你的 API Key
```

### 3. 启动服务

```bash
cd deploy
docker-compose up -d
```

### 4. 访问

- 前端: http://localhost
- Go API: http://localhost:8080
- Python AI: http://localhost:8000

## 📦 项目结构

```
ecommerce-ai-agent/
├── deploy/                  # Docker 部署配置
│   ├── docker-compose.yml
│   ├── .env
│   └── init/mysql/schema.sql
├── backend-go/             # Go 后端 (Gin + go-zero)
│   ├── cmd/api/
│   ├── internal/
│   └── model/
├── backend-python/        # Python AI (FastAPI + LangChain)
│   ├── api/
│   └── core/
└── frontend/             # Vue 3 前端
    └── src/
```

## 🔌 API 接口

| 接口 | 方法 | 描述 |
|------|------|------|
| `/api/v1/auth/register` | POST | 用户注册 |
| `/api/v1/auth/login` | POST | 用户登录 |
| `/api/v1/products` | GET | 商品列表 |
| `/api/v1/search/semantic` | GET | 语义搜索 |
| `/api/v1/orders` | POST | 创建订单 |

## 🤖 AI 功能

- **语义搜索**: 输入自然语言搜索商品（如"适合送女生的口红"）
- **智能客服**: 基于商品信息的 AI 问答

## 🛠️ 技术栈

- **前端**: Vue 3 + Vite + Pinia + Element Plus
- **Go 后端**: Gin + go-zero + GORM + Kafka
- **Python AI**: FastAPI + LangChain + Milvus
- **基础设施**: MySQL + Redis + Kafka + Milvus

## 📄 License

MIT
