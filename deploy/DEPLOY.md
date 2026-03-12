# 🚀 云服务器部署指南

## 服务器信息

- **公网 IP**: 111.230.175.98
- **SSH 端口**: 22

## 部署步骤

### 1. 连接服务器

```bash
ssh root@111.230.175.98
```

### 2. 安装 Docker

```bash
# 安装 Docker
curl -fsSL https://get.docker.com | sh

# 启动 Docker
systemctl start docker
systemctl enable docker

# 安装 Docker Compose
curl -L "https://github.com/docker/compose/releases/download/v2.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
```

### 3. 克隆项目

```bash
cd /opt
git clone https://github.com/Null-noone/ecommerce-ai-agent.git
cd ecommerce-ai-agent
```

### 4. 配置环境变量

```bash
cd deploy
cp .env.example .env
# 编辑 .env，填入你的 VOLC_API_KEY
```

### 5. 启动服务

```bash
chmod +x deploy.sh
./deploy.sh
```

### 6. 开放端口

```bash
# 开放防火墙端口
firewall-cmd --permanent --add-port=80/tcp
firewall-cmd --permanent --add-port=8080/tcp
firewall-cmd --reload
```

## 访问地址

| 服务 | 地址 |
|------|------|
| 前端 | http://111.230.175.98 |
| Go API | http://111.230.175.98:8080 |
| Python AI | http://111.230.175.98:8000 |

## 常用命令

```bash
# 查看日志
docker-compose logs -f

# 重启服务
docker-compose restart

# 停止服务
docker-compose down
```

## 配置 HTTPS (可选)

建议使用 Nginx 反向代理 + Let's Encrypt：
