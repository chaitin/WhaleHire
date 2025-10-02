# 🐳 WhaleHire

**用 AI 重塑招聘管理**

WhaleHire 是一个基于人工智能的现代化招聘管理平台，旨在通过智能化技术提升招聘效率，为 HR 和求职者提供更好的招聘体验。

## ✨ 项目特色

- 🤖 **AI 智能助手**: 集成先进的 AI 对话系统，提供智能化招聘建议
- 🎯 **精准匹配**: 基于 AI 算法的简历与职位智能匹配
- 📊 **数据驱动**: 全面的招聘数据分析和可视化报表
- 🔒 **安全可靠**: 企业级安全保障，保护用户隐私和数据安全
- 🌐 **现代化界面**: 响应式设计，支持多设备访问

## 🏗️ 技术架构

### 整体架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端 (UI)      │    │   后端 (API)    │    │   数据存储        │
│                 │    │                 │    │                 │
│ Next.js 15      │◄──►│ Go + Echo       │◄──►│ PostgreSQL      │
│ TypeScript      │    │ Clean Arch      │    │ Redis           │
│ Tailwind CSS    │    │ Wire DI         │    │ Vector DB       │
│ shadcn/ui       │    │ Ent ORM         │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 技术栈

#### 后端技术栈

- **语言**: Go 1.25+
- **Web 框架**: Echo v4.13.4
- **ORM**: Ent v0.14.4
- **数据库**: PostgreSQL + pgvector
- **缓存**: Redis v7
- **依赖注入**: Google Wire
- **配置管理**: Viper
- **API 文档**: Swagger

#### 前端技术栈

- **框架**: React 18.2.0 + Vite 4.4.5
- **语言**: TypeScript 5.0.2
- **路由**: React Router DOM 6.15.0
- **UI 库**: Radix UI + shadcn/ui
- **样式**: Tailwind CSS 3.3.3
- **图标**: Lucide React 0.263.1
- **图表**: ECharts + echarts-for-react
- **构建工具**: Vite

## 🚀 快速开始

### 环境要求

- **Node.js**: >= 18.0.0
- **Go**: >= 1.25.0
- **Docker**: >= 20.0.0
- **Docker Compose**: >= 2.0.0

### 1. 克隆项目

```bash
git clone <repository-url>
cd WhaleHire
```

### 2. 环境配置

复制环境变量文件并配置：

```bash
cp .env.example .env
vim .env
```

配置以下环境变量：

```bash
# Database Configuration
POSTGRES_DB=whalehire
POSTGRES_USER=whalehire
POSTGRES_PASSWORD=whalehire123

# Redis Configuration
REDIS_PASSWORD=redis123

# Application Configuration
WHALEHIRE_DATABASE_MASTER="postgres://whalehire:whalehire123@localhost:5432/whalehire?sslmode=disable&timezone=Asia/Shanghai"
WHALEHIRE_DATABASE_SLAVE="postgres://whalehire:whalehire123@localhost:5432/whalehire?sslmode=disable&timezone=Asia/Shanghai"
WHALEHIRE_REDIS_HOST=localhost
WHALEHIRE_REDIS_PORT=6379
WHALEHIRE_REDIS_PASS=redis123

# Admin Configuration
WHALEHIRE_ADMIN_USER=admin
WHALEHIRE_ADMIN_PASSWORD=admin123
```

### 3. 启动基础服务

启动数据库和缓存服务：

```bash
# 启动 PostgreSQL 数据库
docker-compose up -d whalehire-db

# 启动 Redis 缓存
docker-compose up -d whalehire-redis
```

### 4. 启动后端服务

```bash
cd backend
go mod tidy
go run cmd/main.go cmd/wire_gen.go
```

后端服务将在 `http://localhost:8888` 启动

### 5. 启动前端服务

```bash
cd frontend
npm install
npm run dev
```

前端服务将在 `http://localhost:5175` 启动

### 6. 访问应用

- **前端应用**: http://localhost:5175
- **后端 API**: http://localhost:8888

## 📁 项目结构

```
WhaleHire/
├── backend/                 # 后端服务
│   ├── cmd/                # 应用入口
│   ├── internal/           # 业务逻辑
│   ├── pkg/                # 公共包
│   ├── db/                 # 数据模型
│   ├── ent/                # Schema 定义
│   └── README.md           # 后端详细文档
├── frontend/               # 前端应用 (React + Vite)
│   ├── src/                # 源代码
│   │   ├── components/     # 组件目录
│   │   │   ├── ui/         # shadcn/ui 组件
│   │   │   ├── auth/       # 认证相关组件
│   │   │   ├── layout/     # 布局组件
│   │   │   └── ...         # 其他业务组件
│   │   ├── pages/          # 页面组件
│   │   ├── hooks/          # 自定义 Hooks
│   │   ├── services/       # API 服务
│   │   ├── types/          # TypeScript 类型定义
│   │   ├── lib/            # 工具函数
│   │   └── styles/         # 样式文件
│   ├── public/             # 静态资源
│   └── README.md           # 前端详细文档
├── docs/                   # 项目文档
├── docker-compose.yml      # 生产环境容器编排
├── docker-compose.dev.yml  # 开发环境容器编排
└── Makefile               # 项目管理脚本
```

## 📚 详细文档

- **[后端开发文档](./backend/README.md)** - Go 后端服务的详细开发指南
- **[前端开发文档](./frontend/README.md)** - React + Vite 前端应用的详细开发指南
- **[架构设计文档](./docs/specialized-agents-architecture.md)** - 系统架构设计说明
- **[智能体设计文档](./docs/agent-design.md)** - AI 智能体设计方案

## 🛠️ 开发工具

### Git Hooks

项目配置了 Git Hooks 来确保代码质量：

```bash
# 安装 Git Hooks
make install-hooks

# 卸载 Git Hooks
make uninstall-hooks
```

### 代码规范

- **后端**: 遵循 Go 官方代码规范，使用 `gofmt`、`go vet`、`golangci-lint`
**前端**: 使用 ESLint + TypeScript 严格模式，遵循 React + Vite 最佳实践

## 🐳 Docker 部署

### 开发环境

```bash
# 启动开发环境（数据库 + Redis）
docker-compose -f docker-compose.dev.yml up -d
```

### 生产环境

```bash
# 启动完整服务栈
docker-compose up -d
```

## 🤝 贡献指南

我们欢迎所有形式的贡献！请阅读以下指南：

### 开发流程

1. **Fork** 项目到你的 GitHub 账户
2. **Clone** 你的 Fork 到本地
3. **创建功能分支**: `git checkout -b feature/your-feature`
4. **开发功能** 并确保通过所有测试
5. **提交代码**: `git commit -m "feat: add your feature"`
6. **推送分支**: `git push origin feature/your-feature`
7. **创建 Pull Request**

### 提交规范

使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

- `feat:` 新功能
- `fix:` 修复 bug
- `docs:` 文档更新
- `style:` 代码格式调整
- `refactor:` 代码重构
- `test:` 测试相关
- `chore:` 构建工具或辅助工具的变动

## 📄 许可证

本项目采用 [AGPL-3.0 license](./LICENSE)。

## 🆘 问题反馈

如果您在使用过程中遇到问题，请：

1. 查看相关文档和 FAQ
2. 搜索 [Issues](../../issues) 是否有相似问题
3. 创建新的 Issue 并详细描述问题
4. 提供复现步骤和环境信息

## 🔗 相关链接

**敬请期待**

- [项目官网](https://whalehire.com)
- [在线演示](https://demo.whalehire.com)
- [API 文档](https://api.whalehire.com/docs)
- [开发者社区](https://community.whalehire.com)

---

**让 AI 赋能招聘，让招聘更智能！** 🚀
