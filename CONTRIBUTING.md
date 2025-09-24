# 🤝 贡献指南

欢迎为 WhaleHire 项目做贡献！我们非常感谢社区的每一份贡献。

## 📋 快速开始

### 1. 准备工作

在开始贡献之前，请确保你已经：

- 阅读了 [项目 README](./README.md) 了解项目概况
- 查看了 [后端开发文档](./backend/README.md) 和 [前端开发文档](./ui/README.md)
- 搭建好了本地开发环境

### 2. 开发环境要求

- **Go**: >= 1.25.0
- **Node.js**: >= 18.0.0
- **Docker**: >= 20.0.0
- **PostgreSQL**: >= 17.0
- **Redis**: >= 7.0

详细的环境搭建步骤请参考 [项目 README](./README.md#🚀-快速开始)。

## 🔄 贡献流程

### 1. Fork 和 Clone

```bash
# Fork 项目到你的 GitHub 账户
# 然后 clone 到本地
git clone https://github.com/your-username/WhaleHire.git
cd WhaleHire
```

### 2. 创建功能分支

```bash
git checkout -b feature/your-feature-name
# 或者修复 bug
git checkout -b fix/your-bug-fix
```

### 3. 开发和测试

- **后端开发**: 参考 [后端开发文档](./backend/README.md)
- **前端开发**: 参考 [前端开发文档](./ui/README.md)
- 确保代码通过所有测试和代码检查

### 4. 提交代码

```bash
git add .
git commit -m "feat: add your feature description"
git push origin feature/your-feature-name
```

### 5. 创建 Pull Request

- 在 GitHub 上创建 Pull Request
- 填写详细的 PR 描述
- 等待代码审查和合并

## 📝 提交规范

我们使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### 提交类型

- `feat`: 新功能
- `fix`: 修复 bug
- `docs`: 文档更新
- `style`: 代码格式调整（不影响功能）
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 构建工具或辅助工具的变动

### 示例

```bash
feat(auth): add OAuth login support
fix(api): resolve user registration validation issue
docs(readme): update installation instructions
```

## 🛠️ 代码质量

### Git Hooks

项目配置了 Git Hooks 来确保代码质量：

```bash
# 安装 Git Hooks
make install-hooks
```

### 代码检查

**后端**:
- 使用 `go fmt` 格式化代码
- 使用 `go vet` 进行静态分析
- 使用 `golangci-lint` 进行代码质量检查

**前端**:
- 使用 ESLint 进行代码检查
- 使用 TypeScript 严格模式
- 遵循 Next.js 最佳实践

## 📚 开发指南

### 技术栈和架构

详细的技术栈信息请查看：
- [项目 README - 技术架构](./README.md#🏗️-技术架构)
- [后端技术栈](./backend/README.md#技术栈)
- [前端技术栈](./ui/README.md#🚀-技术栈)

### 编码规范

- **后端编码规范**: 详见 [后端 README](./backend/README.md#代码规范)
- **前端编码规范**: 详见 [前端 README](./ui/README.md#🔧-开发指南)

### API 设计

- 遵循 RESTful API 设计原则
- 使用统一的响应格式
- API 文档使用 Swagger 自动生成
- 详细规范请参考 [后端 API 文档](./backend/README.md#api-接口)

## 🐛 问题报告

### 报告 Bug

1. 搜索 [Issues](../../issues) 确认问题未被报告
2. 创建新的 Issue 并包含：
   - 问题描述
   - 复现步骤
   - 期望行为
   - 实际行为
   - 环境信息（操作系统、浏览器版本等）

### 功能请求

1. 搜索现有的 Issues 和 Pull Requests
2. 创建新的 Issue 描述：
   - 功能需求
   - 使用场景
   - 预期收益

## 💬 社区交流

- **GitHub Issues**: 技术问题讨论
- **Pull Requests**: 代码审查和讨论
- **GitHub Discussions**: 功能建议和一般讨论

## 📄 许可证

通过贡献代码，你同意你的贡献将在 [MIT 许可证](./LICENSE) 下发布。

## 🙏 致谢

感谢所有为 WhaleHire 项目做出贡献的开发者！

---

**让我们一起用 AI 重塑招聘管理！** 🚀
