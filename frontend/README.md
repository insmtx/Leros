# SingerOS Frontend

SingerOS 前端工程 — 基于 pnpm + Turborepo 的 monorepo 架构，包含 Web 应用和 Electron 桌面应用，共享 UI 组件库与状态管理包。

## 项目结构

```
frontend/
├── apps/                        # 应用入口
│   ├── web/                     # Web 应用（Next.js 16）
│   │   ├── app/                 # App Router 页面
│   │   │   ├── layout.tsx       # RootLayout (ThemeProvider + Toaster)
│   │   │   ├── page.tsx         # 主页 (→ Shell)
│   │   │   └── globals.css      # 全局样式 + TailwindCSS
│   │   ├── components/          # Web 专属业务组件
│   │   │   ├── chat/            # 聊天界面组件
│   │   │   ├── input/           # 输入区组件
│   │   │   └── layout/          # 布局组件 (Shell, LeftRail, etc.)
│   │   ├── public/              # 静态资源
│   │   ├── next.config.ts       # Next.js 配置 (transpilePackages)
│   │   ├── postcss.config.mjs   # PostCSS 配置
│   │   └── tsconfig.json        # TypeScript 配置
│   │
│   └── desktop/                 # 桌面应用（Electron + Vite）
│   │   ├── src/
│   │   │   ├── main/            # Electron 主进程
│   │   │   ├── preload/         # Preload 脚本
│   │   │   └── renderer/        # 渲染进程（React SPA）
│   │   │       ├── src/
│   │   │       │   ├── App.tsx  # 根组件 (BrowserRouter + ThemeProvider)
│   │   │       │   ├── routes.tsx  # 路由定义
│   │   │       │   ├── main.tsx    # 渲染入口
│   │   │       │   ├── globals.css # 全局样式
│   │   │       │   ├── components/ # 桌面端业务组件 (chat/input/layout)
│   │   │       │   └── platform/   # 桌面端平台适配（预留）
│   │   │       └── index.html     # 渲染进程 HTML
│   │   ├── resources/           # 打包资源
│   │   ├── electron.vite.config.ts  # Electron-Vite 配置
│   │   ├── electron-builder.yml     # 打包配置 (mac/win/linux)
│   │   └── tsconfig.web.json        # 渲染进程 TS 配置
│   │   └── tsconfig.node.json       # 主进程 TS 配置
│   │
├── packages/                    # 共享包
│   ├── ui/                      # @singeros/ui — UI 组件库
│   │   ├── components/
│   │   │   ├── ui/              # 54 个基础 UI 原语组件 (button, dialog, etc.)
│   │   │   └── common/          # 通用业务组件 (theme-provider)
│   │   ├── hooks/               # 共享 Hooks (use-mobile, use-sse, use-websocket)
│   │   ├── lib/                 # 工具库 (request, sse, websocket, utils)
│   │   ├── styles/              # 设计系统样式 (tokens.css, base.css)
│   │   ├── components.json      # shadcn 配置
│   │   └── package.json         # 导出路径映射
│   │
│   ├── store/                   # @singeros/store — Zustand 状态管理
│   │   ├── appStore.ts          # Store 入口 (合并 layoutSlice + topicSlice + chatSlice)
│   │   ├── slices/              # 状态切片
│   │   │   ├── layoutSlice.ts   # 布局状态 (工作区、导航、会话列开关)
│   │   │   ├── topicSlice.ts    # Topic CRUD (乐观更新)
│   │   │   └── chatSlice.ts     # 聊天核心 (消息流、工具调用、输入状态)
│   │   ├── types/               # 领域类型 (api.ts, chat.ts)
│   │   ├── mocks/               # Mock 数据 (chatMocks, streamSimulator)
│   │   ├── utils/               # 工具函数 (flattenActions, format)
│   │   └── package.json         # 导出路径映射
│   │
│   ├── tsconfig/                # @singeros/tsconfig — 共享 TS 配置
│   │   ├── base.json            # 基础配置 (strict, ESNext, bundler)
│   │   ├── next.json            # Next.js 配置 (继承 base)
│   │   └── react-library.json   # React 库配置 (继承 base)
│   │
│   ├── biome/                   # @singeros/biome — 共享 Biome lint 配置
│   │   └── biome.json           # 规则定义 (recommended + 自定义)
│   │
├── pnpm-workspace.yaml          # 工作空间定义 + catalog 依赖版本管理
├── turbo.json                   # Turborepo 任务配置 (build/dev/typecheck/test/lint)
├── package.json                 # 根 package.json (monorepo 脚本)
├── biome.json                   # Biome 入口 (extends @singeros/biome)
├── .npmrc                       # pnpm 配置 (shamefully-hoist)
└── .gitignore                   # Git 忽略规则
```

## 包说明

| 包名 | 位置 | 技术栈 | 用途 |
|------|------|--------|------|
| `@singeros/web` | `apps/web` | Next.js 16 + React 19 + TailwindCSS 4 | Web 端应用 |
| `@singeros/desktop` | `apps/desktop` | Electron 39 + React 19 + Vite 8 | 桌面端应用 |
| `@singeros/ui` | `packages/ui` | React 19 + @base-ui/react + CVA + TailwindCSS 4 | 共享 UI 组件库 |
| `@singeros/store` | `packages/store` | Zustand 5 + React 19 | 共享状态管理 |
| `@singeros/tsconfig` | `packages/tsconfig` | — | 共享 TypeScript 配置 |
| `@singeros/biome` | `packages/biome` | Biome 2 | 共享代码检查配置 |

## 本地开发

### 前置要求

- **Node.js** >= 20
- **pnpm** >= 10（项目锁定 `pnpm@10.28.2`）

### 安装依赖

```bash
cd frontend
pnpm install
```

### 启动开发

```bash
# Web 应用 (端口 3005)
pnpm dev:web

# 桌面应用 (Electron)
pnpm dev:desktop

# 同时启动所有
pnpm dev
```

### 构建

```bash
# 构建所有包
pnpm build

# 构建指定应用
turbo build --filter=@singeros/web
turbo build --filter=@singeros/desktop
```

### 其他命令

```bash
# 类型检查
pnpm typecheck

# 运行测试
pnpm test

# 代码检查
pnpm lint

# 清理构建产物和依赖
pnpm clean
```

### 添加 UI 组件

使用 shadcn CLI 向 `@singeros/ui` 包添加新组件：

```bash
pnpm ui:add
# 然后在交互界面中选择要添加的组件
```

新增组件后，需在 `packages/ui/package.json` 的 `exports` 中添加导出路径，以便其他包可通过 `@singeros/ui/components/ui/<name>` 引入。

## 依赖版本管理

项目使用 `pnpm-workspace.yaml` 的 `catalog` 功能统一管理核心依赖版本：

- React / React-DOM: `19.2.3`
- TypeScript: `^5.9.3`
- TailwindCSS: `^4`
- Vitest: `^4.1.0`

各子包通过 `catalog:` 引用统一版本，避免版本不一致问题。

## 架构说明

### 双端共享策略

Web 和 Desktop 应用共享同一套 UI 组件库（`@singeros/ui`）和状态管理（`@singeros/store`），但使用不同的框架和路由：

| 维度 | Web (Next.js) | Desktop (Electron) |
|------|---------------|---------------------|
| 渲染 | SSR + CSR | CSR (本地渲染) |
| 路由 | App Router (Next.js) | BrowserRouter (react-router-dom) |
| 样式 | PostCSS + TailwindCSS | Vite + @tailwindcss/vite |
| 端口 | 3005 | 5175 (renderer) |
| 构建输出 | `.next/` | `out/` + `build/` |

### Next.js 转译配置

Web 应用需要显式配置 `transpilePackages` 以正确引用 workspace 包：

```ts
// apps/web/next.config.ts
const nextConfig: NextConfig = {
  transpilePackages: ["@singeros/ui", "@singeros/store"],
};
```

### Electron-Vite 配置

Desktop 应用使用 `electron-vite` 构建，renderer 进程需要 `dedupe` 配置避免 React 双实例：

```ts
// apps/desktop/electron.vite.config.ts
renderer: {
  resolve: {
    dedupe: ["react", "react-dom"],
  },
}
```
