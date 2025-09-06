# WhaleHire Docker Compose 部署指南

本文档说明如何使用 Docker Compose 快速启动 WhaleHire 完整系统，包括前端、后端、PostgreSQL 数据库和 Redis 缓存。

## 系统架构

- **前端**: Next.js 应用 (端口 3000)
- **后端**: Go 应用 (端口 8888)
- **数据库**: PostgreSQL with pgvector 扩展 (端口 5432)
- **缓存**: Redis (端口 6379)

## 快速启动

### 1. 环境准备

确保已安装以下软件：

- Docker
- Docker Compose
- Git

### 2. 克隆项目

```bash
git clone <repository-url>
cd WhaleHire
```

### 3. 环境变量配置

复制环境变量模板文件：

```bash
cp .env.example .env
```

根据需要修改 `.env` 文件中的配置。

### 4. 启动服务

启动所有服务：

```bash
docker-compose up -d
```

查看服务状态：

```bash
docker-compose ps
```

查看日志：

```bash
# 查看所有服务日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f whalehire-backend
docker-compose logs -f whalehire-frontend
```

### 5. 访问应用

- **前端应用**: http://localhost:3000
- **后端 API**: http://localhost:8888
- **数据库**: localhost:5432 (用户名: whalehire, 密码: whalehire123)
- **Redis**: localhost:6379 (密码: redis123)

## 服务管理

### 停止服务

```bash
docker-compose down
```

### 重启服务

```bash
docker-compose restart
```

### 重新构建并启动

```bash
docker-compose up --build -d
```

### 清理数据卷

```bash
docker-compose down -v
```

## 开发模式

### 方式一：使用开发专用 docker-compose 文件（推荐）

```bash
# 启动开发环境的数据库和 Redis（映射到宿主机端口）
docker-compose -f docker-compose.dev.yml up -d

# 在本地启动后端（使用开发配置）
cd backend
go run cmd/main.go cmd/wire_gen.go -config=config/config.dev.yaml

# 在本地启动前端
cd ui
npm install
npm run dev
```

### 方式二：从完整环境中选择服务

```bash
# 只启动数据库和 Redis
docker-compose up whalehire-db whalehire-redis -d

# 在本地启动后端（需要修改配置文件中的主机名为 localhost）
cd backend
go run cmd/main.go cmd/wire_gen.go

# 在本地启动前端
cd ui
npm install
npm run dev
```

### 开发环境说明

- `docker-compose.dev.yml`: 专为开发环境设计，只包含数据库和 Redis
- `config/config.dev.yaml`: 开发环境配置，数据库和 Redis 主机名设为 localhost
- 数据库端口 5432 和 Redis 端口 6379 直接映射到宿主机
- 本地运行的后端可以直接通过 localhost 访问容器中的服务
- 开发环境使用独立的数据卷，不会影响生产环境数据

## 故障排除

### 常见问题

1. **端口冲突**

   - 确保端口 3000, 8888, 5432, 6379 未被占用
   - 可以在 docker-compose.yml 中修改端口映射

2. **数据库连接失败**

   - 检查数据库健康检查状态
   - 确保后端等待数据库完全启动

3. **前端构建失败**
   - 检查 Node.js 版本兼容性
   - 清理 node_modules 并重新安装依赖

### 查看容器状态

```bash
# 查看运行中的容器
docker ps

# 进入容器调试
docker exec -it whalehire-backend sh
docker exec -it whalehire-postgres psql -U whalehire -d whalehire
```

## 生产部署注意事项

1. **安全配置**

   - 修改默认密码
   - 使用强密码
   - 配置防火墙规则

2. **性能优化**

   - 调整数据库连接池大小
   - 配置 Redis 内存限制
   - 启用 Gzip 压缩

3. **监控和日志**
   - 配置日志轮转
   - 设置监控告警
   - 定期备份数据

## 技术栈

- **前端**: Next.js 15, React 19, TypeScript, Tailwind CSS
- **后端**: Go, Echo, Ent ORM, Wire DI
- **数据库**: PostgreSQL 17 with pgvector
- **缓存**: Redis 7
- **容器化**: Docker, Docker Compose
