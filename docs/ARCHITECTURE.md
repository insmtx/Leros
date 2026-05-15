# Leros 架构设计文档

> 基于 **Event Engine + Execution Engine + Agent Runtime 三核架构** 构建的企业级 AI 操作系统
>
> **版本：3.2** | **最后更新：2026-05-13**

## 1. 核心愿景

构建一个企业级数字员工平台，让企业可以像管理真实员工一样，创建、配置、授权、调度和审计 AI 数字员工，并实现：

* **多 Agent 协作** — 多个智能体协同工作
* **多运行时执行** — 支持不同 Agent 引擎并存
* **本地 + 云端协同** — Edge 与 Remote Runtime 分工
* **可控、安全、可审计** — 企业级安全控制

数字员工不是单纯的聊天机器人。它需要有独立身份、接收任务的入口、真实执行工作的环境，以及模型、工具、技能、知识库等基础能力。

## 设计原则

* **事件驱动（Event-Driven First）** — 所有行为统一抽象为 Event，通过 Event Bus 传播
* **控制面 / 执行面分离（Control vs Execution）** — 决策与执行彻底解耦
* **三核架构（Three-Core Architecture）** — Event Engine + Execution Engine + Agent Runtime 职责分离
* **领域驱动设计（Domain-Driven Design）** — 按领域分层（event / execution / agent / skill），而非按技术分层（controller / service / model）
* **接口优先（Interface-Driven）** — 每一层都必须定义 interface，而不是直接依赖实现
* **核心引擎内聚可替换** — Event Engine、Execution Engine、Agent Runtime 必须可独立替换和部署
* **分层命名（Layered Naming）** — Engine = 执行能力 | Runtime = 运行时容器 | Service = 对外能力 | Connector = 外部接入
* **边缘优先（Edge-First）** — 本地能力（文件 / GUI）优先由 Edge Runtime 执行
* **安全优先（Security by Design）** — 明确本地与远程执行边界
* **数字助手是最高抽象（Digital Assistant First）** — 代表完整的 AI 数字员工实例
* **强制隔离（Enforced Isolation）** — 使用 internal 目录强制隔离核心实现，pkg 对外公开接口

## 2. 分层架构（四平面模型）

### 2.1 架构总览

```
┌───────────────────────────────────────────┐
│                Client / Edge               │
│      App / CLI / 本地 Agent Runtime       │
└────────────────────┬───────────────────────┘
                      │
                      ▼
┌────────────────────────────────────────────┐
│            Interface Layer（接口层）        │
│         Assistant Service / Connector      │
└────────────────────┬───────────────────────┘
                      │
                      ▼
┌────────────────────────────────────────────┐
│          Control Plane（控制面）            │
│  Event Engine / Memory / Policy Engine    │
└────────────────────┬───────────────────────┘
                      │
                      ▼
┌────────────────────────────────────────────┐
│          Execution Plane（执行面）          │
│  Execution Engine / Agent Runtime / Skill  │
└────────────────────────────────────────────┘
```

### 2.2 四平面职责

| 平面 | 组件 | 职责 |
|------|------|------|
| **Edge Plane** | Edge Runtime / Client | 本地文件访问、GUI 自动化、用户环境交互 |
| **Interface Layer** | Assistant Service / Connector | 对外 API / 渠道接入 / 事件标准化 |
| **Control Plane** | Event Engine / Memory / Policy Engine | 事件路由、上下文构建、权限控制 |
| **Execution Plane** | Execution Engine / Agent Runtime / Skill | Agent 推理、Skill 调用、Workflow 编排 |

### 2.3 核心数据通道（统一事件流）

```
External Event / User Input
         │
         ▼
Connector（事件标准化）
         │
         ▼
Event Bus（统一事件模型）
         │
         ▼
Event Engine（事件路由）
         │
         ▼
Execution Engine（执行调度）
         │
         ▼
Agent Runtime / Workflow Engine / Skill（执行单元）
         │
         ▼
Event Bus（响应流）
         │
         ▼
Assistant Service → Client / UI
```

> **核心原则**：所有模块之间只能通过 Event Bus 通信

## 3. 核心模块划分

### 3.1 Connector（连接器）

**职责：**

* 接收外部系统事件（Webhook / API / 用户输入）
* 标准化为内部 Event
* 发布到 Event Bus

**支持渠道：**

GitHub / GitLab / 企业微信 / 飞书 / CLI / Web UI

**关键能力：**

签名验证、多协议适配、事件换

**包路径：**

```
internal/api/connectors/
├── github/     GitHub webhook 连接器
├── gitlab/     GitLab 连接器（存根）
└── wework/     企业微信连接器（存根）
```

Connector 通过 internal/api 路由注册到 Gin，不再使用独立的 Event Gateway。

### 3.2 Event Bus（事件总线）

系统唯一通信通道，所有模块之间只能通过 Event Bus 通信。

**实现：** NATS JetStream

**标准 Event 模型核心字段：**

* **ID** — 事件唯一标识
* **Type** — 事件类型（command / response / stream / state / system 等）
* **Source** — 事件来源
* **Target** — 事件目标
* **SessionID** — 会话标识
* **Payload** — 事件载荷
* **Timestamp** — 事件戳

**Event 分类：**

* command — 指令事件
* response — 响应事件
* stream — 流式事件
* state — 状态事件
* system — 系统事件

### 3.3 Assistant Service（助手服务）【原 Gateway】

对外统一 API 入口，多渠道统一访问，用户请求处理。

### 3.4 Event Engine（事件引擎）【原 Orchestrator】

**职责：** 订阅事件总线事件、事件路由与分发、调用 Handler 处理、触发执行流程。

**核心能力：** 事件过滤与路由规则、事件聚合与防抖、事件优先级调度。

**包：** `internal/eventengine`

**当前状态（Phase 1.5）：**

* ✅ 已实现事件订阅和路由
* ✅ 已实现 Handler 注册机制
* ⚠️ Router 未独立提取（Phase 2）
* ⚠️ Handler 未插件化（Phase 2）
* ⚠️ Execution Engine 未独立（Phase 2）

**本质：** 系统的响应中心 — 负责响应外部事件并启动执行流程。

### 3.5 Execution Engine（执行引擎）【Phase 2 计划中】

**职责：** 调用 Skill / Workflow / Agent；控制执行流程（同步 / 异步 / 重试）。

**核心能力：** 同步/异步执行控制、重试与降级机制、执行超时控制、并发执行管理。

**当前状态：** 尚未独立实现。当前流程：Event Handler → 直接调用 Agent Runtime。Phase 2 计划从 Orchestrator 中抽离。

**核心原则：** Event Engine 与 Execution Engine 必须解耦。

### 3.6 Agent Runtime（智能体运行时）

**职责：** 管理 Agent 生命周期、调用 LLM、管理 Memory / Context、工具调用（Tool / Skill）、多引擎调度（Leros / Claude / Codex）、会话级上下文构建与自我学习。

**核心能力：**

* Agent 状态管理、上下文维护（Skill 目录、Tool 注、会话记忆）
* 推理循环（Reasoning Loop）、工具调用协调
* Session 上下文注入、流式事件输出与 Trace 记录
* 运行后自我学习、多引擎注册与默认选择

**包结构：**

```
internal/agent/
├── events/           Agent 流式事件系统
├── eventtypes/       NATS 领域消息协议
├── eino/             Cloudwego Eino 框架集成
├── leros/            Leros 原生运行时
├── externalcli/      外部 CLI 引擎适配器（Claude Code, Codex）
├── runtime/          服务层：构建 Environment + Router
│   ├── env/          运行时环境（Skill 目录 + Tool 注册表）
│   └── lifecycle/    生命周期包装（上下文构建 + Trace + 自我学习）
└── simplechat/       轻量级简单聊天实现
```

**已实现能力：**

Session 上下文注入、事件流式输出、Leros 原生运行时 + Eino 集成、外部 CLI 引擎适配（Claude Code, Codex）、多引擎注册与路由分发、Lifecycle 层（系统提示词构建、Skill 注入、Trace 记录、自我学习）。

### 3.7 Workflow Engine（工作流引擎）【规划中】

多步骤任务编排、DAG / 状态机执行、长任务执行管理。

### 3.8 Runtime Manager（运行时调度器）

管理所有 Runtime 实例、能力注册（Skill / GPU / Browser）、负载均衡、健康检查。

### 3.9 Memory（记忆系统）

会话上下文（短期记忆）、长期记忆（向量）、知识检索（RAG）。

### 3.10 Model Router（模型调度）

多模型管理、fallback 降级、成本控制。

### 3.11 Policy Engine（策略引擎）

**职责：** Agent 行为控制、Skill 调用权限、审计日志。

**强制规则：** Remote Runtime 不得直接访问本地资源。所有高权限操作必须经过 Policy Engine。

### 3.12 Skills 能力系统

Skill 是可复用的 AI 能力单元，是 Leros 的核心构建块。

**Skill 分类：**

* **集成类 Skills** — 外部系统集成（GitHub、GitLab、飞书等）
* **AI 类 Skills** — 基于大模型的推理能力（代码审查、摘要生成、分类等）
* **工具类 Skills** — 底层工具能力（Shell 执行、Python 脚本、HTTP 请求等）
* **工作流类 Skills** — 复杂编排能力（PR 审查工作流、Bug 分类工作流等）

**技能加载方式：** 文件系统当前主要方式、代码嵌入编译时打包、远程加载规划中。

**包结构：**

```
backend/skills/         Skill 定义文件（SKILL.md）
├── code-review/
├── commit-conventions/
├── humanizer-zh/
└── weather/

backend/internal/skill/ Skill 运行时管理系统
├── catalog/           Skill 目录（文件加载 + 静态目录）
├── runtime/           Skill 运行时（Manager + PostProcessor）
└── store/             Skill 持久化存储

backend/tools/         Tool 执行代码
├── skill_manage/      Skill 管理工具
├── skill_use/        Skill 使用工具
├── memory/           内存工具
└── node/             Node.js 工具运行时
```

**当前状态（Phase 1.5）：**

Skill Registry 已完成；`internal/skill/` 运行时系统已实现；`internal/agent/runtime/env` 提供 Runtime 级别的 Skill 目录 + Tool 注册表；Tool 执行在 `backend/tools/`。

### 3.13 Tools 工具系统

Tool 是底层原子能力。与 Skills 的区别：

| 维度 | Tools | Skills |
|------|-------|--------|
| 粒度 | 原子操作 | 可组合 |
| 注册 | 系统注册 | 用户可创建 |
| 侧重 | 执行 | 智能决策 |

关系：Agent → Skill → Tool

**内置 Tools：** HTTP 请求、Shell 命令执行、Python 脚本执行、文件读写操作、数据库查询工具。

## 4. 数字助手（核心抽象）

数字助手是企业中的"AI 员工"。

**组成：** 身份信息 / 运行时配置 / 模型配置 / Skills 集合 / 渠道绑定 / Memory / Policy

**助手状态：**

* **草稿** — 配置中，未启用
* **激活** — 正常运行，可接收事件
* **停用** — 临时禁用
* **归档** — 历史版本归档

## 5. 执行面组件

### 5.1 Agent Runtime（远程执行节点）

**职责：** 消费任务 Event、执行 Agent 推理、调用 Skill。

**特性：** 无状态（或弱状态）、Worker 模式、不暴露 API。

### 5.2 Edge Runtime（本地执行节点）

**职责：** 本地文件访问、GUI 自动化（AX / UIA）、本地模型、用户环境交互。

| 能力 | Edge | Remote |
|------|------|--------|
| 本地文件 | 是 | 否 |
| GUI 操作 | 是 | 否 |
| 云执行 | 否 | 是 |

**安全原则：** Edge Runtime 是唯一可操作用户环境的组件。

## 6. 关键执行链路（统一模型）

### 6.1 标准执行链路

```
User / Webhook
  │
  ▼
Connector（事件标准化）
  │
  ▼
Event Bus
  │
  ▼
Event Engine（事件路由）
  │
  ▼
Execution Engine（执行调度）
  │
  ▼
┌─────────────────────────────────┐
│  Agent Runtime / Workflow       │
│  Engine / Direct Skill Call    │
└─────────────────────────────────┘
  │
  ▼
Skill / Tool 执行
  │
  ▼
Event Bus（流式返回）
  │
  ▼
Assistant Service → Client
```

### 6.2 示例：GitHub PR 自动审查流程

1. **事件触发** — 开发者创建 PR，GitHub 发送 Webhook
2. **事件接收** — GitHub Connector 接收请求
3. **签名验证** — 验证 Webhook 签名确保来源合法
4. **事件标准化** — 转换为内部 Event 格式
5. **事件发布** — 发布到 Event Bus
6. **事件消费** — Event Engine 订阅并处理事件
7. **路由匹配** — Event Engine 根据事件类型选择 Handler
8. **执行触发** — Event Engine 调用 Execution Engine
9. **执行调度** — Execution Engine 决定执行策略（同步/异步/重试）
10. **节点选择** — Runtime Manager 选择合适的 Runtime 节点
11. **配置加载** — Agent Runtime 加载目标数字助手的配置
12. **上下文构建** — 获取 PR 差异内容，构建提示词
13. **能力注入** — 注入代码审查 Skills 和 GitHub Tools
14. **大模型推理** — Agent Runtime 调用 LLM 分析代码并生成审查意见
15. **工具执行** — Execution Engine 调用 GitHub API 发布 Review 评论
16. **结果返回** — 通过 Event Bus 流式返回执行结果
17. **结果记录** — 持久化到事件表

## 7. 安全模型

### 三层权限模型

```
Edge Runtime      → 高权限（本地）
Control Plane     → 中权限（调度）
Remote Runtime    → 低权限（执行）
```

### 核心规则

* Remote 不能访问本地
* 所有敏感操作必须经过 Policy Engine
* 全链路审计

### 安全边界

| 组件 | 权限级别 | 可访问资源 |
|------|----------|------------|
| Edge Runtime | 高 | 本地文件、GUI、用户环境 |
| Control Plane | 中 | 调度、路由、配置 |
| Remote Runtime | 低 | 云端资源、API |
| Policy Engine | 最高 | 权限决策、审计 |

## 8. Go 包结构（领域驱动设计）

### 8.1 设计原则

* **按"领域分层"，不是按技术分层** — 按 event / execution / agent / skill 分层，而非 controller / service / dao

### 8.2 目录结构

```
backend/
│
├── cmd/                        启动入口
│   └── leros/                 leros 二进制
│       ├── server              server 子命令
│       └── worker              worker 子命令
│           ├── simplechat     内置简单聊天运行时
│           └── claude-worker  Claude Code 引擎运行时
│
├── internal/                  私有核心代码（强制隔离）
│   ├── api/                   HTTP 适配层（契约驱动）
│   │   ├── handler/           HTTP 处理器
│   │   ├── dto/               数据传输对象
│   │   ├── contract/          系统能力定义（Service 接口 + DTO）
│   │   ├── middleware/        HTTP 中间件
│   │   ├── auth/              认证上下文
│   │   └── connectors/        渠道连接器
│   │       ├── github/
│   │       ├── gitlab/
│   │       └── wework/
│   │
│   ├── eventengine/          事件引擎（GitHub 事件 Orchestrator）
│   │
│   ├── agent/                Agent Runtime（多引擎运行时系统）
│   │   ├── events/           Agent 流式事件（EventSink, Emitter）
│   │   ├── eventtypes/       NATS 领域消息协议
│   │   ├── eino/             Cloudwego Eino 集成
│   │   ├── leros/            内置 Leros 原生运行时
│   │   ├── externalcli/      外部 CLI 引擎适配（Claude, Codex）
│   │   ├── runtime/          服务层：构建 Environment + Router
│   │   │   ├── env/          运行时环境（Skill 目录 + Tool 注册表）
│   │   │   └── lifecycle/    生命周期包装（上下文 + Trace + 学习）
│   │   └── simplechat/       轻量级简单聊天实现
│   │
│   ├── service/              业务逻辑层
│   │
│   ├── skill/                Skill 系统（运行时管理）
│   │   ├── catalog/          Skill 目录（文件加载 + 静态目录）
│   │   ├── runtime/          Skill 运行时（Manager + PostProcessor）
│   │   └── store/            Skill 持久化存储
│   │
│   ├── worker/               Worker 管理系统
│   │   ├── client/           Worker 客户端（WebSocket + 任务执行）
│   │   ├── server/           Worker 管理服务（HTTP + WS 服务器）
│   │   ├── scheduler/        Worker 调度器（进程/Docker 容器）
│   │   ├── taskconsumer/     NATS 任务消费者 → agent.Runner
│   │   └── wsproto/          Worker-Server WebSocket 协议
│   │
│   ├── memory/               记忆系统
│   │   └── local/            本地内存存储
│   │
│   └── infra/                基础设施
│       ├── mq/               NATS JetStream 消息队列
│       ├── db/               数据库访问
│       ├── providers/        第三方服务 Provider
│       │   └── github/
│       └── websocket/        WebSocket 工具
│
├── pkg/                      对外公开接口
│   ├── dm/                   领域消息协议（NATS Topic 构建）
│   ├── event/               交互事件常量
│   └── leros/               Leros 工具函数
│
├── runtime/                  运行时层（独立于 internal）
│   ├── engines/             外部 AI CLI 引擎抽象
│   │   ├── builtin/         内置引擎工厂
│   │   ├── claude/          Claude Code 引擎适配
│   │   └── codex/           Codex CLI 引擎适配
│   └── events/              共享事件契约
│
├── types/                    核心类型定义
├── config/                   配置管理
├── skills/                   Skill 定义文件（SKILL.md）
│   ├── code-review/
│   ├── commit-conventions/
│   ├── humanizer-zh/
│   └── weather/
├── tools/                    Tool 实现（注册 + 执行）
├── mcp/                      MCP 服务器（Worker 运行时引导）
└── tests/                    测试工具
```

### 8.3 internal 目录说明

* Go 编译器强制保证只能被本项目内部引用
* 明确"内部实现"与"对外接口"的边界
* 为后续拆分多进程/微服务做准备

### 8.4 pkg 目录说明

* 对外公开的类型和 SDK
* 其他项目可以安全导入

### 8.5 进程阶段

* **Phase 1.5（当前）** — leros 二进制通过 server/worker 子命令区分运行模式
* **Phase 2（计划）** — NATS 任务消费者 + Agent Runner + MCP Server；支持 worker simplechat / worker claude-worker 子命令
* **Phase 3（远期）** — 独立进程部署（Server / Worker / Connector）

## 9. 技术栈

| 类别     | 技术                                 |
| -------- | ------------------------------------ |
| 语言     | Golang                               |
| 网关     | Gin                                  |
| 事件总线 | NATS JetStream                       |
| 数据库   | PostgreSQL                           |
| 缓存     | Redis                                |
| 向量库   | Qdrant                               |
| LLM      | 多模型（OpenAI / Claude / DeepSeek）|
| 容器化   | Docker + Compose                     |

## 10. 架构演进路径

### Phase 1.5（当前实际）

* `leros` 单进程服务（Server + Worker + Orchestrator 合一）
* GitHub 自动化闭环（Webhook → NATS → Event Engine → Agent Runtime）
* 事件总线（NATS JetStream）
* Connector 层完成（GitHub / GitLab / WeWork）
* Agent Runtime 完整实现：Leros 原生运行时 + 外部 CLI 引擎适配（Claude Code, Codex）+ 多引擎路由器 + Lifecycle 层
* Worker 管理系统：进程/Docker 容器调度 + WebSocket 通信（server/client）+ 任务消费者（taskconsumer）
* Session 管理（消息增删、状态流转、NATS topic 构建）
* Skill 系统三层分离：`internal/skill/`（目录/运行时/存储）+ `backend/skills/`（SKILL.md 文件）+ `backend/tools/`（Tool 执行）
* MCP 服务器集成（Worker 运行时引导）
* ⚠️ Event Engine 与 Execution Engine 未完全分离（Phase 2）

### Phase 1（原始计划）

* 单运行时
* GitHub 自动化闭环
* 基础 Event Bus
* Connector 层完成
* ~~Event Engine 与 Execution Engine 分离~~ → 延期至 Phase 2

### Phase 2

* Execution Engine 独立（从 Orchestrator 中抽离执行逻辑）
* Event Engine Handler 插件化
* 流式事件完善
* Runtime Manager（多运行时管理/健康检查/负载均衡）
* Worker 进程完善（Docker 容器调度）

### Phase 3

* Workflow Engine
* Memory + RAG
* Policy Engine 完整落地

### Phase 4

* 多租户
* Skill Marketplace
* 企业级治理能力

### Phase 5

* 进程拆分（Server / Worker / Connector）
* 分布式部署
* 水平扩展

## 11. 附录：架构演进历史

### v3.2 (2026-05-13) — Agent Runtime 重构 + Worker 系统完成

* 统一 Runner 接口，消除 AgentRuntime 与 Runner 重复定义
* Agent Runtime 包结构大幅演进：新增 runtime 服务层（Environment + Router）、lifecycle 生命周期管理、leros 原生运行时、externalcli 外部 CLI 适配
* Worker 管理系统完整实现：server/client/scheduler/taskconsumer/wsproto
* 新增 MCP 服务器集成、pkg/dm 领域消息协议、pkg/leros 工具函数
* 移除废弃的 backend/gateway 和 backend/interaction 模块，Connector 并入 internal/api/connectors

### v3.1.1 (2026-04-27) — 架构实现状态更新

* ⚠️ Event Engine：Orchestrator 已实现，但 Router 未独立、Handler 未插件化（Phase 2）
* ⚠️ Execution Engine：尚未独立实现，执行逻辑在 Orchestrator 中（Phase 2）
* ✅ Agent Runtime：完整实现（SimpleChat + Eino + Session 上下文）
* ✅ API 层：契约驱动服务架构（handler/dto/contract/middleware）
* ✅ Skill System：Registry 化完成

### v3.1 (2026-04-23) — Go 包结构优化

* 使用 `internal/` 实现核心代码隔离
* 使用 `pkg/` 对外公开接口
* Skill Registry 化
* 接口优先设计（每层定义 interface）

### v3.0 (2026-04-23) — 三核架构重构

* Orchestrator → Event Engine（专注事件处理）
* 新增 Execution Engine（专注执行控制）
* Agent Runtime 职责明确（专注 Agent 推理）
* Gateway → Assistant Service（明确对外服务定位）

### 命名演变

| 版本 | 核心模块命名 |
|------|-------------|
| v1.0 | Gateway / Orchestrator / Agent Runtime |
| v2.0 | Gateway / Orchestrator / Agent Runtime（细化职责） |
| v3.0 | Assistant Service / Event Engine / Execution Engine / Agent Runtime（三核架构） |
| v3.1 | 引入 `internal/` 和 `pkg/` 强制隔离 |
| v3.2 | 移除 gateway/interaction，Connector 并入 api，Agent Runtime 多引擎架构，Worker 系统完整实现 |

## 12. 总结

### Leros 的本质：

一个 **事件驱动的分布式 Agent 操作系统**

### 核心能力：

* 多 Agent 编排
* 多 Runtime 执行
* 本地 + 云协同
* 企业级安全控制

### 架构关键词：

```
Event-Driven
Three-Core Architecture
Domain-Driven Design
Interface-First
Control / Execution Separation
Multi-Runtime
Edge + Cloud
Policy-Driven
Enforced Isolation (internal)
```

### 核心架构公式：

```
Connector → Event → Event Engine → Execution Engine → Capability → Service
                                                  ↓
                                      Agent Runtime / Workflow / Skill
```

### 常见错误清单（务必避免）

| ❌ 错误做法 | ✅ 正确做法 |
|------------|------------|
| 把所有逻辑写进 Event Handler | Handler → 调 Execution Engine |
| Event Handler 使用 switch 硬编码路由 | Router 独立 + Handler 插件化 |
| Agent Runtime 直接调 MQ / DB | 通过 Execution Engine / Skill / Infra |
| Skill 写死在代码中 | 必须 Registry 化，支持动态注册 |
| 按技术分层（controller/service/model）| 按领域分层（event/execution/agent/skill） |
| 缺少接口定义，直接依赖实现 | 每层定义 interface，支持替换 |
