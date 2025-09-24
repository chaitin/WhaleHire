# WhaleHire 前端项目

WhaleHire 是一个基于 AI 的智能招聘平台前端应用，采用现代化的技术栈构建，为 HR 和求职者提供高效的招聘解决方案。

## 🚀 技术栈

- **框架**: [Next.js 15.5.2](https://nextjs.org/) (App Router)
- **语言**: [TypeScript 5.x](https://www.typescriptlang.org/)
- **UI 库**: [Radix UI](https://www.radix-ui.com/) + [shadcn/ui](https://ui.shadcn.com/)
- **样式**: [Tailwind CSS v4](https://tailwindcss.com/)
- **表单**: [React Hook Form](https://react-hook-form.com/) + [Zod](https://zod.dev/)
- **图标**: [Lucide React](https://lucide.dev/)
- **构建工具**: [Turbopack](https://turbo.build/pack)
- **代码规范**: ESLint + TypeScript

## 📁 项目结构

```
ui/
├── src/
│   ├── app/                    # Next.js App Router 页面
│   │   ├── admin/             # 管理员页面
│   │   │   ├── dashboard/     # 管理员仪表板
│   │   │   └── settings/      # 系统设置
│   │   ├── api/               # API 路由
│   │   ├── auth/              # 认证相关页面
│   │   ├── chat/              # 聊天功能页面
│   │   ├── dashboard/         # 用户仪表板
│   │   ├── globals.css        # 全局样式
│   │   ├── layout.tsx         # 根布局
│   │   └── page.tsx           # 首页（登录/注册）
│   ├── components/            # 组件目录
│   │   ├── admin/             # 管理员组件
│   │   ├── business/          # 业务组件
│   │   ├── custom/            # 自定义组件
│   │   └── ui/                # shadcn/ui 基础组件
│   └── lib/                   # 工具库
│       ├── api/               # API 接口封装
│       └── utils.ts           # 工具函数
├── public/                    # 静态资源
├── components.json            # shadcn/ui 配置
├── next.config.ts             # Next.js 配置
├── tailwind.config.ts         # Tailwind CSS 配置
├── tsconfig.json              # TypeScript 配置
└── package.json               # 项目依赖
```

## 🛠️ 开发环境要求

- **Node.js**: >= 18.0.0
- **npm**: >= 8.0.0 (或 yarn/pnpm)
- **操作系统**: macOS, Linux, Windows

## 🚀 快速开始

### 1. 克隆项目

```bash
git clone <repository-url>
cd WhaleHire/ui
```

### 2. 安装依赖

```bash
npm install
```

### 3. 环境配置

创建 `.env.local` 文件（如需要）：

```bash
# 后端 API 地址（默认: http://localhost:8888）
BACKEND_HOST=http://localhost:8888
```

### 4. 启动开发服务器

```bash
npm run dev
```

打开 [http://localhost:3000](http://localhost:3000) 查看应用。

## 📝 可用脚本

- `npm run dev` - 启动开发服务器
- `npm run build` - 构建生产版本
- `npm run start` - 启动生产服务器
- `npm run lint` - 运行 ESLint 代码检查

## 🏗️ 构建部署

### 开发环境构建

```bash
npm run build
npm run start
```

### Docker 部署

项目包含 Dockerfile，支持容器化部署：

```bash
# 构建镜像
docker build -t whalehire-ui .

# 运行容器
docker run -p 3000:3000 whalehire-ui
```

### 生产环境配置

项目配置了 `output: "standalone"` 模式，构建后会生成独立的部署包，便于在各种环境中部署。

## 🎨 UI 组件系统

项目使用 shadcn/ui 组件系统，基于 Radix UI 构建：

- **设计风格**: New York
- **主题色**: Neutral
- **图标库**: Lucide React
- **CSS 变量**: 支持主题切换

### 添加新组件

```bash
npx shadcn@latest add <component-name>
```

## 🔧 开发指南

### 代码规范

- 使用 TypeScript 严格模式
- 遵循 ESLint 规则
- 组件使用函数式组件 + Hooks
- 样式使用 Tailwind CSS 类名
- 表单验证使用 React Hook Form + Zod

### 文件命名规范

- 组件文件：`kebab-case.tsx`
- 页面文件：`page.tsx`
- 布局文件：`layout.tsx`
- 工具文件：`kebab-case.ts`

### 组件开发规范

```tsx
// 示例组件结构
"use client"; // 客户端组件标记

import { useState } from "react";
import { Button } from "@/components/ui/button";

interface ComponentProps {
  title: string;
  onAction?: () => void;
}

export function ExampleComponent({ title, onAction }: ComponentProps) {
  const [state, setState] = useState(false);

  return (
    <div className="p-4">
      <h2 className="text-lg font-semibold">{title}</h2>
      <Button onClick={onAction}>Action</Button>
    </div>
  );
}
```

## 🤝 贡献指南

### 开发流程

1. **Fork 项目** 并克隆到本地
2. **创建功能分支**: `git checkout -b feature/your-feature`
3. **开发功能** 并确保代码质量
4. **运行测试** 和代码检查
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

### 代码审查要点

- [ ] TypeScript 类型定义完整
- [ ] 组件可复用性良好
- [ ] 样式响应式设计
- [ ] 无控制台错误或警告
- [ ] 代码符合项目规范

## 🔗 相关链接

- [Next.js 文档](https://nextjs.org/docs)
- [Tailwind CSS 文档](https://tailwindcss.com/docs)
- [shadcn/ui 文档](https://ui.shadcn.com/)
- [React Hook Form 文档](https://react-hook-form.com/)
- [Zod 文档](https://zod.dev/)

## 📄 许可证

本项目采用 [AGPL-3.0 license](../LICENSE)。

## 🆘 问题反馈

如果您在使用过程中遇到问题，请：

1. 查看 [Issues](../../issues) 是否有相似问题
2. 创建新的 Issue 并详细描述问题
3. 提供复现步骤和环境信息

---

**Happy Coding! 🎉**
