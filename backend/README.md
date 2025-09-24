# WhaleHire Backend

WhaleHire 招聘平台后端服务，基于 Go 语言开发，采用清洁架构设计。

## 项目架构

### 技术栈

- **语言**: Go 1.21+
- **Web 框架**: Echo v4
- **数据库**: PostgreSQL
- **缓存**: Redis
- **依赖注入**: Google Wire
- **配置管理**: Viper
- **日志**: slog

### 目录结构

```
backend/
├── Makefile            # 构建脚本
├── README.md           # 项目说明文档
├── go.mod              # Go模块依赖
├── go.sum              # Go模块校验和
├── build/              # 构建相关
│   └── Dockerfile      # Docker构建文件
├── cmd/                # 应用入口
│   ├── main.go         # 主程序
│   ├── wire.go         # 依赖注入配置
│   └── wire_gen.go     # 生成的依赖注入代码
├── config/             # 配置文件
│   ├── config.go       # 配置结构定义
│   ├── config.yaml     # 配置文件
│   └── provider.go     # 配置提供者
├── consts/             # 常量定义
│   ├── admin.go        # 管理员相关常量
│   ├── os.go           # 操作系统相关常量
│   └── user.go         # 用户相关常量
├── db/                 # 数据库模型(Ent生成)
│   ├── client.go       # 数据库客户端
│   ├── ent.go          # Ent入口文件
│   ├── mutation.go     # 变更操作
│   ├── tx.go           # 事务处理
│   ├── page.go         # 分页处理
│   ├── admin/          # 管理员模型
│   ├── user/           # 用户模型
│   ├── conversation/   # 对话模型
│   ├── message/        # 消息模型
│   ├── attachment/     # 附件模型
│   ├── role/           # 角色模型
│   ├── setting/        # 设置模型
│   ├── useridentity/   # 用户身份模型
│   ├── userloginhistory/    # 用户登录历史
│   ├── adminrole/      # 管理员角色模型
│   ├── adminloginhistory/   # 管理员登录历史
│   ├── enttest/        # 测试工具
│   ├── hook/           # 钩子函数
│   ├── intercept/      # 拦截器
│   ├── migrate/        # 迁移工具
│   ├── predicate/      # 查询谓词
│   └── runtime/        # 运行时
├── docs/               # API文档
│   ├── docs.go         # Swagger文档生成
│   └── swagger.json    # Swagger JSON文档
├── domain/             # 领域层
│   ├── admin.go        # 管理员领域模型
│   ├── user.go         # 用户领域模型
│   ├── general_agent.go # 通用智能体模型
│   ├── ip.go           # IP相关模型
│   └── oauth.go        # OAuth模型
├── ent/                # Ent Schema定义
│   ├── entc.go         # Ent代码生成配置
│   ├── generate.go     # 生成脚本
│   ├── rule/           # 验证规则
│   ├── schema/         # 数据库Schema
│   │   ├── admin.go    # 管理员Schema
│   │   ├── user.go     # 用户Schema
│   │   ├── conversation.go # 对话Schema
│   │   ├── message.go  # 消息Schema
│   │   ├── attachment.go # 附件Schema
│   │   ├── role.go     # 角色Schema
│   │   ├── setting.go  # 设置Schema
│   │   ├── useridentity.go # 用户身份Schema
│   │   ├── userloginhistory.go # 用户登录历史Schema
│   │   ├── adminrole.go # 管理员角色Schema
│   │   └── adminloginhistory.go # 管理员登录历史Schema
│   └── types/          # 自定义类型
├── errcode/            # 错误码定义
│   ├── errcode.go      # 错误码结构
│   ├── locale.en.toml  # 英文错误信息
│   └── locale.zh.toml  # 中文错误信息
├── internal/           # 内部模块
│   ├── handlers.go     # 通用处理器
│   ├── provider.go     # 内部依赖提供者
│   ├── third.go        # 第三方服务
│   ├── middleware/     # 中间件
│   │   ├── active.go   # 活跃状态中间件
│   │   ├── auth.go     # 认证中间件
│   │   ├── logger.go   # 日志中间件
│   │   └── readonly.go # 只读模式中间件
│   ├── user/           # 用户模块
│   │   ├── handler/    # HTTP处理层
│   │   ├── usecase/    # 业务逻辑层
│   │   └── repo/       # 数据访问层
│   └── general_agent/  # 通用智能体模块
│       ├── handler/    # HTTP处理层
│       ├── usecase/    # 业务逻辑层
│       └── repo/       # 数据访问层
├── migration/          # 数据库迁移文件
│   ├── 000001_add_user_tables.up.sql     # 用户表创建
│   ├── 000001_add_user_tables.down.sql   # 用户表回滚
│   ├── 000002_add_conversation_tables.up.sql   # 对话表创建
│   ├── 000002_add_conversation_tables.down.sql # 对话表回滚
│   ├── 000003_add_agent_name_to_conversations.up.sql   # 智能体名称字段
│   ├── 000003_add_agent_name_to_conversations.down.sql # 智能体名称字段回滚
│   ├── 000004_add_auth.up.sql    # 认证相关表
│   └── 000004_add_auth.down.sql  # 认证相关表回滚
├── pkg/                # 公共包
│   ├── provider.go     # 依赖提供者
│   ├── cvt/            # 类型转换工具
│   ├── eino/           # AI引擎相关
│   │   ├── chains/     # 链式处理
│   │   ├── graphs/     # 图处理
│   │   ├── models/     # 模型定义
│   │   └── tools/      # 工具集
│   ├── entx/           # Ent扩展
│   │   ├── entx.go     # 扩展功能
│   │   ├── softdelete.go # 软删除
│   │   └── tx.go       # 事务扩展
│   ├── ipdb/           # IP数据库
│   │   ├── ip2region.xdb # IP地址库
│   │   └── ipdb.go     # IP查询工具
│   ├── logger/         # 日志组件
│   │   ├── context.go  # 上下文日志
│   │   └── logger.go   # 日志实现
│   ├── oauth/          # OAuth认证
│   ├── request/        # 请求处理
│   ├── service/        # 服务层
│   ├── session/        # 会话管理
│   ├── store/          # 数据存储
│   └── version/        # 版本管理
├── script/             # 脚本文件
│   ├── README.md       # 脚本说明
│   └── create_migration.go # 迁移文件创建脚本
└── templates/          # 模板文件
    └── page.tmpl       # 页面模板
```

### 架构设计

采用清洁架构(Clean Architecture)设计原则：

1. **Handler 层**: 处理 HTTP 请求，参数验证，响应格式化
2. **UseCase 层**: 业务逻辑处理，事务管理
3. **Repository 层**: 数据访问，数据库操作
4. **Domain 层**: 领域模型，业务规则定义

## 快速开始

### 环境要求

- Go 1.25+
- PostgreSQL 12+
- Redis 7+

### 安装依赖

```bash
# 安装wire工具
go install github.com/google/wire/cmd/wire@latest
```

### 开发运行应用

#### 1. 设置环境变量

```bash
cd ..
vim .env

# Database Configuration
POSTGRES_DB=whalehire
POSTGRES_USER=whalehire
POSTGRES_PASSWORD=whalehire123

# Redis Configuration
REDIS_PASSWORD=redis123

# Application Database Configuration
WHALEHIRE_DATABASE_MASTER="postgres://whalehire:whalehire123@localhost:5432/whalehire?sslmode=disable&timezone=Asia/Shanghai"
WHALEHIRE_DATABASE_SLAVE="postgres://whalehire:whalehire123@localhost:5432/whalehire?sslmode=disable&timezone=Asia/Shanghai"
WHALEHIRE_REDIS_HOST=localhost
WHALEHIRE_REDIS_PORT=6379
WHALEHIRE_REDIS_PASS=redis123

# Admin Configuration
WHALEHIRE_ADMIN_USER=admin
WHALEHIRE_ADMIN_PASSWORD=admin123
```

#### 2.启动数据库容器

```bash
cd .. # compose 文件在项目根目录
docker-compose -f docker-compose.dev.yml up -d # 安装数据库和 Redis 镜像
```

#### 3. 运行后端服务

```bash
go mod tidy
go run cmd/main.go cmd/wire_gen.go
```

## API 接口

### 用户管理 API

#### 用户认证相关

- `POST /api/v1/user/register` - 用户注册
- `POST /api/v1/user/login` - 用户登录
- `POST /api/v1/user/logout` - 用户登出
- `GET /api/v1/user/oauth/signup-or-in` - OAuth 注册或登录
- `GET /api/v1/user/oauth/callback` - OAuth 回调处理

#### 用户信息管理

- `GET /api/v1/user/profile` - 获取用户资料
- `PUT /api/v1/user/profile` - 更新用户资料

#### 用户管理（管理员权限）

- `GET /api/v1/user/list` - 获取用户列表
- `PUT /api/v1/user/update` - 更新用户信息
- `DELETE /api/v1/user/delete` - 删除用户
- `GET /api/v1/user/login-history` - 获取用户登录历史

### 管理员 API

#### 管理员认证

- `POST /api/v1/admin/login` - 管理员登录
- `POST /api/v1/admin/logout` - 管理员登出
- `GET /api/v1/admin/profile` - 获取管理员资料

#### 管理员管理

- `POST /api/v1/admin/create` - 创建管理员
- `DELETE /api/v1/admin/delete` - 删除管理员
- `GET /api/v1/admin/list` - 获取管理员列表
- `GET /api/v1/admin/login-history` - 获取管理员登录历史

#### 角色与权限管理

- `GET /api/v1/admin/role` - 获取角色列表
- `POST /api/v1/admin/role` - 分配角色权限

#### 系统设置

- `GET /api/v1/admin/setting` - 获取系统设置
- `PUT /api/v1/admin/setting` - 更新系统设置

### 通用智能体 API

#### AI 对话生成

- `POST /api/v1/general-agent/generate` - 生成 AI 回复
- `POST /api/v1/general-agent/generate-stream` - 流式生成 AI 回复

#### 对话管理

- `POST /api/v1/general-agent/conversations` - 创建对话
- `GET /api/v1/general-agent/conversations` - 获取对话列表
- `POST /api/v1/general-agent/conversations/history` - 获取对话历史
- `DELETE /api/v1/general-agent/conversations` - 删除对话
- `POST /api/v1/general-agent/conversations/{id}/addmessage` - 向对话添加消息

## 开发指南

### 添加新模块

1. 在 `domain/` 中定义领域模型和接口
2. 在 `internal/` 中创建模块目录
3. 实现 Repository、UseCase、Handler 三层
4. 在 `wire.go` 中注册依赖
5. 运行 `make wire` 生成依赖注入代码

### 代码规范

- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 接口定义在 `domain` 包中
- 错误处理使用自定义错误码
- 日志使用结构化日志

### 测试

```bash
# 运行所有测试
make test

# 运行特定包的测试
go test ./internal/user/...
```

## 部署

### Docker 部署

```bash
# 构建镜像
make image
```
