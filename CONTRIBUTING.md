# 贡献指南

欢迎为 WhaleHire 项目做贡献！本指南将帮助你开始贡献代码。

# WhaleHire 项目开发规范

## 项目概述

WhaleHire 是一个基于 AI 的招聘管理系统，采用前后端分离架构。

## 技术栈

### 后端技术栈

- **语言**: Go 1.25.1
- **Web 框架**: Echo v4.13.4
- **ORM**: Ent v0.14.4
- **数据库**: PostgreSQL
- **缓存**: Redis v9.13.0
- **依赖注入**: Google Wire v0.7.0
- **配置管理**: Viper v1.20.1
- **数据库迁移**: golang-migrate v4.19.0
- **API 文档**: Swagger
- **容器化**: Docker

### 前端技术栈

- **框架**: Next.js 15.5.2 (App Router)
- **语言**: TypeScript 5.x
- **UI 库**: Radix UI + shadcn/ui
- **样式**: Tailwind CSS v4
- **表单**: React Hook Form + Zod
- **图标**: Lucide React
- **构建工具**: Turbopack
- **代码规范**: ESLint + Prettier

## 开发规范

### 1. 代码结构规范

#### 后端目录结构

```
backend/
├── cmd/                 # 应用入口
├── config/             # 配置文件
├── internal/           # 内部业务逻辑
├── pkg/                # 公共包
├── db/                 # 数据库模型(Ent生成)
├── ent/                # Ent schema定义
├── domain/             # 领域模型
├── errcode/            # 错误码定义
├── docs/               # API文档
└── migration/          # 数据库迁移文件
```

#### 前端目录结构

```
ui/
├── src/
│   ├── app/            # Next.js App Router页面
│   ├── components/     # 组件目录
│   │   ├── ui/         # shadcn/ui组件
│   │   └── ...         # 业务组件
│   └── lib/            # 工具函数
├── public/             # 静态资源
└── 配置文件
```

### 2. 编码规范

#### 后端编码规范

**命名规范**:

- 包名: 小写，简洁明了
- 函数名: 驼峰命名，公开函数首字母大写
- 变量名: 驼峰命名，私有变量首字母小写
- 常量名: 全大写，下划线分隔
- 接口名: 以"er"结尾（如 Servicer）

**代码组织**:

- 使用依赖注入（Wire）管理依赖
- 遵循 Clean Architecture 原则
- 错误处理必须显式处理，不允许忽略
- 使用 context.Context 传递请求上下文

**数据库规范**:

- 使用 Ent 进行 ORM 操作
- 数据库迁移文件必须有 up 和 down 版本
- 所有数据库操作必须在事务中进行
- 敏感数据必须加密存储

#### 前端编码规范

**组件规范**:

- 使用函数式组件和 React Hooks
- 组件名使用 PascalCase
- 文件名使用 kebab-case
- 优先使用 shadcn/ui 组件
- 自定义组件放在 components 目录下

**TypeScript 规范**:

- 严格模式开启（strict: true）
- 所有函数参数和返回值必须有类型注解
- 使用 interface 定义对象类型
- 使用泛型提高代码复用性

**样式规范**:

- 使用 Tailwind CSS 进行样式开发
- 遵循移动端优先的响应式设计
- 使用 CSS 变量管理主题色彩
- 组件样式使用 class-variance-authority 管理变体

### 3. 代码质量保证

#### 代码检查工具

**后端**:

- 使用`go fmt`格式化代码
- 使用`go vet`进行静态分析
- 使用`golangci-lint`进行代码质量检查
- 单元测试覆盖率不低于 80%

**前端**:

- ESLint 配置基于 Next.js 推荐规则
- 使用 Prettier 进行代码格式化
- TypeScript 严格模式检查
- 组件必须通过 React Testing Library 测试

#### Git 提交规范

**提交信息格式**:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**类型说明**:

- `feat`: 新功能
- `fix`: 修复 bug
- `docs`: 文档更新
- `style`: 代码格式调整
- `refactor`: 重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动

### 4. API 设计规范

#### RESTful API 规范

- 使用 HTTP 动词表示操作（GET, POST, PUT, DELETE）
- URL 使用名词，避免动词
- 统一的响应格式
- 使用 HTTP 状态码表示请求结果
- API 版本控制（/api/v1/）

#### 响应格式

```json
{
  "code": 200,
  "message": "success",
  "data": {},
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 5. 安全规范

- 所有 API 接口必须进行身份验证
- 敏感数据传输使用 HTTPS
- 输入数据必须进行验证和过滤
- 使用 JWT 进行用户认证
- 实施 CORS 策略
- 定期更新依赖包，修复安全漏洞

### 6. 性能优化规范

#### 后端性能

- 数据库查询优化，避免 N+1 问题
- 使用 Redis 缓存热点数据
- 实施数据库连接池
- API 响应时间不超过 500ms

#### 前端性能

- 使用 Next.js 的 SSR/SSG 优化首屏加载
- 图片资源使用 WebP 格式
- 实施代码分割和懒加载
- 使用 React.memo 优化组件渲染
- Core Web Vitals 指标达到 Good 标准

### 7. 测试规范

#### 后端测试

- 单元测试：使用 Go 标准 testing 包
- 集成测试：测试 API 端点
- 数据库测试：使用测试数据库

#### 前端测试

- 组件测试：React Testing Library
- E2E 测试：Playwright 或 Cypress
- 可访问性测试：遵循 WCAG 2.1 AA 标准

### 8. 部署规范

- 使用 Docker 容器化部署
- 环境变量管理敏感配置
- 实施 CI/CD 流水线
- 生产环境使用 HTTPS
- 实施健康检查和监控

### 9. 文档规范

- API 文档使用 Swagger 自动生成
- 代码注释使用英文
- README 文档包含项目启动说明
- 重要功能必须有设计文档

### 10. 开发流程

1. **需求分析** → 技术方案设计
2. **数据库设计** → Ent Schema 定义
3. **API 设计** → Swagger 文档
4. **后端开发** → 单元测试
5. **前端开发** → 组件测试
6. **集成测试** → E2E 测试
7. **代码审查** → 部署上线

## 工具配置

### 开发环境要求

- Go 1.25+
- Node.js 18+
- PostgreSQL 17+
- Redis 7+
- Docker & Docker Compose

### IDE 推荐配置

- VS Code + Go 扩展
- GoLand
- 配置 ESLint 和 Prettier 插件

## 其他指南

- 提交消息应清晰且有意义
- 大功能实现应先创建设计文档
- 问题讨论可以在 GitHub Issues 中进行
- 遇到问题随时提问
