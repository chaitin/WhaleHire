# WhaleHire Backend

WhaleHire 招聘平台后端服务，基于 Go 语言开发，采用清洁架构设计。

## 项目架构

### 技术栈

- **语言**: Go 1.21+
- **Web框架**: Echo v4
- **数据库**: PostgreSQL
- **缓存**: Redis
- **依赖注入**: Google Wire
- **配置管理**: Viper
- **日志**: slog

### 目录结构

```
backend/
├── cmd/server/          # 应用入口
│   ├── main.go         # 主程序
│   ├── wire.go         # 依赖注入配置
│   └── wire_gen.go     # 生成的依赖注入代码
├── config/             # 配置文件
│   ├── config.go       # 配置结构定义
│   └── config.yaml     # 配置文件
├── domain/             # 领域层
│   └── user.go         # 用户领域模型
├── internal/           # 内部模块
│   └── user/           # 用户模块
│       ├── handler/    # HTTP处理层
│       ├── usecase/    # 业务逻辑层
│       └── repo/       # 数据访问层
├── pkg/                # 公共包
│   ├── logger/         # 日志组件
│   ├── session/        # 会话管理
│   ├── store/          # 数据存储
│   └── provider.go     # 依赖提供者
├── consts/             # 常量定义
├── errcode/            # 错误码定义
└── ent/                # 数据库实体(预留)
```

### 架构设计

采用清洁架构(Clean Architecture)设计原则：

1. **Handler层**: 处理HTTP请求，参数验证，响应格式化
2. **UseCase层**: 业务逻辑处理，事务管理
3. **Repository层**: 数据访问，数据库操作
4. **Domain层**: 领域模型，业务规则定义

## 快速开始

### 环境要求

- Go 1.21+
- PostgreSQL 12+
- Redis 6+

### 安装依赖

```bash
# 安装项目依赖
make deps

# 安装wire工具
go install github.com/google/wire/cmd/wire@latest
```

### 配置数据库

1. 创建PostgreSQL数据库:
```sql
CREATE DATABASE whalehire;
```

2. 修改配置文件 `config/config.yaml` 中的数据库连接信息

### 运行应用

```bash
# 开发模式运行
make dev

# 或者构建后运行
make build
make run
```

### 可用命令

```bash
make help      # 查看所有可用命令
make deps      # 安装依赖
make wire      # 生成依赖注入代码
make build     # 构建应用
make run       # 运行应用
make dev       # 开发模式运行
make test      # 运行测试
make clean     # 清理构建文件
make fmt       # 格式化代码
```

## API 接口

### 用户管理 API

- `POST /api/v1/users/register` - 用户注册
- `POST /api/v1/users/login` - 用户登录
- `POST /api/v1/users/logout` - 用户登出
- `GET /api/v1/users/profile` - 获取用户资料
- `PUT /api/v1/users/profile` - 更新用户资料
- `PUT /api/v1/users/password` - 修改密码
- `GET /api/v1/users` - 获取用户列表(管理员)
- `GET /api/v1/users/{id}` - 根据ID获取用户
- `DELETE /api/v1/users/{id}` - 删除用户(管理员)

### 健康检查

- `GET /health` - 服务健康检查

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
docker build -t whalehire-backend .

# 运行容器
docker run -p 8080:8080 whalehire-backend
```

### 生产环境配置

1. 修改 `config/config.yaml` 中的生产环境配置
2. 设置环境变量覆盖敏感配置
3. 配置反向代理(Nginx)
4. 设置日志收集和监控

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交代码
4. 创建 Pull Request

## 许可证

MIT License