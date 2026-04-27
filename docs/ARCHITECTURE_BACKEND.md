# SingerOS 后端架构设计文档

> 面向 AI OS 的 Golang 包结构指南
>
> **版本：3.1** | **最后更新：2026-04-23**

## 1. 概述

本文档提供 SingerOS 后端的 **Golang 包结构设计**，与 `ARCHITECTURE.md` 配合使用。

- `ARCHITECTURE.md` - 高层架构设计、模块划分、执行链路
- `ARCHITECTURE_BACKEND.md` - **本文档** - Go 包结构、目录组织

## 2. 设计原则

### 2.1 架构定位：契约驱动服务架构

> **Contract-driven Service Architecture**

SingerOS 采用**契约驱动的服务架构**，而不是：
- ❌ Clean Architecture
- ❌ DDD
- ❌ Hexagonal Architecture

**特点：**
- ✔ 类 RPC 风格（类 RPC，但不绑定 RPC 实现）
- ✔ 轻抽象（无 Repository 层）
- ✔ 高工程效率
- ✔ 适合 Agent / Workflow / OS 类系统扩展

**核心原则：**
1. **contract 定义能力** - 系统能力的"语言"
2. **service 实现能力** - 直接操作 DB
3. **handler 适配输入输出** - HTTP 适配
4. **db 提供执行能力** - 真实数据库操作

### 2.2 按"领域分层"，不是按技术分层

> ❌ 旧模式：controller / service / dao / model
> ✅ 新模式：event / execution / agent / skill

**原因：**
- 技术分层导致模块间耦合严重
- 领域分层让每个模块职责清晰、可独立演进

### 2.3 核心引擎必须"内聚 + 可替换"

- Event Engine 可以单独部署
- Execution Engine 可以替换
- Agent Runtime 可扩展

### 2.4 强制隔离（Enforced Isolation）

| 目录 | 用途 | 访问控制 |
|------|------|----------|
| `internal/` | 私有核心代码 | Go 编译器强制隔离，只能被本项目内部引用 |
| `pkg/` | 对外公开接口 | 其他项目可安全导入 |

## 3. 推荐的 Golang 包结构

### 3.1 完整目录结构（推荐版本）

```bash
backend/
│
├── cmd/
│   └── singer/                # 主后端服务（HTTP + 事件网关）
│       └── main.go
│
├── internal/                  # 私有核心代码（强制隔离）
│
│   ├── api/                   # HTTP 适配层
│   │   ├── handler/
│   │   │   ├── event_handler.go
│   │   │   ├── agent_handler.go
│   │   │   └── session_handler.go
│   │   ├── dto/
│   │   │   ├── event.go
│   │   │   ├── agent.go
│   │   │   └── session.go
│   │   ├── contract/          # ⭐ 系统能力定义（核心）
│   │   │   ├── event.go
│   │   │   ├── task.go
│   │   │   ├── agent.go
│   │   │   └── session.go
│   │   └── router.go          # 路由注册

│   │
│   ├── eventengine/           # ⭐ 事件引擎
│   │   ├── engine.go              # Event Engine 核心
│   │   ├── router.go              # 事件路由
│   │   ├── registry.go            # Handler 注册中心
│   │   └── builtins/              # 内置事件处理器
│   │
│   ├── execution/             # ⭐ 执行引擎
│   │   ├── engine.go              # Execution Engine 核心
│   │   ├── dispatcher.go          # 调度器
│   │   ├── executor.go            # 执行器接口
│   │   ├── context/               # 执行上下文
│   │   └── retry.go               # 重试控制
│   │
│   ├── agent/                 # ⭐ Agent Runtime
│   │   ├── runtime.go             # Agent Runtime 接口
│   │   ├── lifecycle.go           # 生命周期管理
│   │   ├── context.go             # 上下文管理
│   │   └── eino/                  # Eino 具体实现
│   │
│   ├── skill/                 # ⭐ Skill 体系
│   │   ├── registry.go            # Skill 注册中心
│   │   ├── executor.go            # Skill 执行器
│   │   ├── base_skill.go          # 基础 Skill
│   │   └── builtin/               # 内置技能
│   │
│   ├── service/               # ⭐ 业务逻辑层（直接操作 DB）
│   │   ├── event_service.go
│   │   ├── agent_service.go
│   │   └── session_service.go
│   │
│   ├── db/                    # 数据访问层
│   │   ├── client.go              # GORM client
│   │   ├── event_repo.go          # 直接 DB 操作
│   │   ├── agent_repo.go
│   │   └── model/                 # 数据库模型
│   │       ├── event_model.go
│   │       └── agent_model.go
│   │
│   ├── convert/               # 结构转换层
│   │   ├── event_convert.go
│   │   └── agent_convert.go
│   │
│   ├── wire/                  # 依赖组装层
│   │   ├── event_wire.go
│   │   └── agent_wire.go
│   │
│   ├── connectors/            # 外部接入
│   │   ├── connector.go
│   │   ├── github/
│   │   ├── gitlab/
│   │   └── wework/
│   │
│   └── infra/                 # 基础设施
│       ├── mq/                    # 消息队列（NATS）
│       ├── db/                    # 数据库连接
│       └── logger/                # 日志
│
├── pkg/                       # 对外公开接口
│   ├── event/                 # Event 定义
│   ├── client/                # SDK
│   └── types/                 # 公开类型
│
├── types/                     # 核心类型定义
├── config/                    # 配置管理
├── auth/                      # 认证系统
└── skills/                    # 技能体系（旧版，逐步迁移）
```

## 4. 核心模块说明

### 4.1 `internal/api/` - HTTP 适配层 ⭐⭐

包含：handler / dto / contract / router

**定位：** HTTP → Command → service

#### contract 示例（在 api/contract/ 中）

```go
package contract

type EventService interface {
    CreateEvent(ctx context.Context, cmd CreateEventCommand) error
    GetEvent(ctx context.Context, id string) (*Event, error)
}

type CreateEventCommand struct {
    ChannelName string
    EventType   string
    Payload     map[string]interface{}
    Timestamp   int64
}
```

**特点：**
* ✔ 不依赖 Gin
* ✔ 不依赖 DB
* ✔ 是系统"能力语言"

---

#### handler 示例（在 api/handler/ 中）

```go
func (h *EventHandler) Create(c *gin.Context) {
    var req dto.CreateEventRequest
    _ = c.ShouldBindJSON(&req)

    cmd := convert.ToCommand(req)

    err := h.service.CreateEvent(c.Request.Context(), cmd)

    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"ok": true})
}
```

**特点：**
* ❌ 不写业务逻辑
* ❌ 不操作 DB
* ✔ 只做转换 + 调用

---

#### router 示例（在 api/ 根目录）

```go
func Register(r *gin.Engine, h *EventHandler) {
    r.POST("/events", h.Create)
    r.GET("/events/:id", h.Get)
}
```

---

### 4.2 `internal/eventengine/` - 事件引擎 ⭐⭐⭐

### 4.3 `internal/execution/` - 执行引擎

**职责：** 任务调度、执行控制、重试/超时管理

**子目录：**
- `engine.go` - Execution Engine 核心
- `dispatcher.go` - 调度器（任务分发）
- `executor.go` - 执行器接口
- `sync_executor.go` / `async_executor.go` - 同步/异步执行器
- `retry.go` / `timeout.go` - 重试和超时控制
- `context/` - 执行上下文

**关键点：**
- 支持同步/异步执行
- 支持重试和降级
- 支持超时控制

---

### 4.4 `internal/agent/` - Agent Runtime

**职责：** Agent 生命周期管理、LLM 调用、上下文维护

**子目录：**
- `runtime.go` - Agent Runtime 接口
- `lifecycle.go` - 生命周期管理
- `context.go` - 上下文管理
- `reasoning.go` - 推理循环
- `eino/` - Eino 具体实现

**⚠️ 常见错误：**
- ❌ Agent Runtime 直接调用 MQ / DB
- ✅ 必须通过 contract.AgentService

---

### 4.5 `internal/service/` - 业务逻辑层 ⭐⭐⭐

**定位：** 业务编排 + 直接 DB 操作

#### 实现 contract

```go
type eventService struct {
    db *db.Client
}

var _ contract.EventService = (*eventService)(nil)
```

#### 核心逻辑

```go
func (s *eventService) CreateEvent(
    ctx context.Context,
    cmd contract.CreateEventCommand,
) error {
    event := &db.EventModel{
        ID:          uuid.New().String(),
        ChannelName: cmd.ChannelName,
        EventType:   cmd.EventType,
        Payload:     cmd.Payload,
        Timestamp:   cmd.Timestamp,
    }

    return s.db.CreateEvent(event)
}
```

---

### 4.6 `internal/db/` - 数据访问层 ⭐⭐

**定位：** 真实数据库操作（只是 SQL/ORM wrapper）

#### 示例

```go
type Client struct {
    db *gorm.DB
}

func (c *Client) CreateEvent(e *EventModel) error {
    return c.db.Create(e).Error
}
```

---

### 4.7 `internal/convert/` - 结构转换层 ⭐

**定位：** DTO ↔ Command ↔ Model

#### 示例

```go
func ToCommand(req dto.CreateEventRequest) contract.CreateEventCommand {
    return contract.CreateEventCommand{
        ChannelName: req.ChannelName,
        EventType:   req.EventType,
        Payload:     req.Payload,
        Timestamp:   req.Timestamp,
    }
}
```

**特点：**
* ✔ 避免 DTO 污染 service
* ✔ 保持层隔离

---

### 4.8 `internal/wire/` - 依赖组装层 ⭐

**定位：** 依赖组装，避免 main 爆炸

#### 示例

```go
func NewEventModule(db *db.Client) contract.EventService {
    svc := &eventService{db: db}
    return svc
}
```

---

### 4.9 `internal/skill/` - Skill 体系

**职责：** 技能注册、执行、管理

**子目录：**
- `registry.go` - Skill 注册中心（必须动态注册）
- `executor.go` - Skill 执行器
- `base_skill.go` - 基础 Skill 实现
- `builtin/` - 内置技能

**⚠️ 常见错误：**
- ❌ Skill 写死在代码中
- ✅ 必须 Registry 化，支持动态注册

### 4.10 `internal/connectors/` - 连接器

**职责：** 外部系统接入（GitHub、GitLab、飞书等）

**子目录：**
- `connector.go` - Connector 接口
- `github/` - GitHub 连接器
- `gitlab/` - GitLab 连接器
- `wework/` - 企业微信连接器

### 4.11 `internal/infra/` - 基础设施

**职责：** 统一基础设施访问

**子目录：**
- `mq/` - 消息队列（NATS Publisher / Subscriber）
- `db/` - 数据库连接
- `logger/` - 日志

### 4.12 `pkg/` - 对外公开接口

**职责：** 对外共享的类型和 SDK

**子目录：**
- `event/` - Event 定义（对外共享）
- `client/` - SingerOS SDK
- `types/` - 公开类型

## 5. 进程拆分建议

### 5.1 为什么需要进程拆分？

| 优势 | 说明 |
|------|------|
| 水平扩展 | 不同组件独立扩缩容 |
| 解耦 | 故障隔离 |
| 负载分离 | 不同负载类型分开处理 |

### 5.2 推荐的进程拆分方案

#### Phase 1（当前）：单进程

```bash
cmd/singer/               # 主服务（所有功能）
```

#### Phase 2：分离执行节点

```bash
cmd/server/               # API 服务（HTTP + Event Engine）
cmd/worker/               # 执行节点（Execution Engine + Agent Runtime）
```

#### Phase 3：分离连接器

```bash
cmd/connector/            # 连接器进程（Connectors + Event Bus Publisher）
```

### 5.3 进程间通信

所有进程间通过 **Event Bus** 通信：

```
Connector Process → Event Bus → Event Engine Process → Execution Engine Process → Agent Runtime Worker
```

## 6. 常见错误与最佳实践

### 6.1 完整调用链（最重要）

```
HTTP Request
   ↓
handler (DTO → Command)
   ↓
contract.Service (interface)
   ↓
service (business logic)
   ↓
db client (SQL/ORM)
   ↓
Database
```

### 6.2 核心设计原则（架构本质）

#### ✔ 1️⃣ contract 是系统语言

不是 RPC，但类似 RPC 的"能力定义"

#### ✔ 2️⃣ service 是唯一业务执行点

不允许业务散落

#### ✔ 3️⃣ DB 不抽象 interface

直接调用

#### ✔ 4️⃣ handler 只做适配

不参与业务

#### ✔ 5️⃣ convert 保持隔离

避免 DTO 污染 service

### 6.3 常见错误

| ❌ 错误做法 | ✅ 正确做法 |
|------------|------------|
| 把所有逻辑写进 Event Handler | Handler → 调用 contract.Service |
| Event Handler 使用 `switch` 硬编码路由 | Router 独立 + Handler 插件化 |
| Agent Runtime 直接调 MQ / DB | 通过 contract.AgentService |
| Skill 写死在代码中 | 必须 Registry 化，支持动态注册 |
| 按技术分层（controller/service/model） | 按领域分层（contract/handler/service/db） |
| 添加 Repository 抽象 | service 直接调用 db.client |
| 缺少接口定义，直接依赖实现 | contract 定义所有能力 |

### 6.4 最佳实践

1. **contract 定义系统能力** - 每层定义 interface，支持替换
2. **service 直接操作 DB** - 无 Repository 抽象
3. **handler 只做转换 + 调用** - 不写业务逻辑
4. **convert 保持隔离** - DTO ↔ Command ↔ Model
5. **wire 组装依赖** - 避免 main 爆炸
6. **每个包只暴露必要的接口**
7. **使用 `internal/` 强制隔离核心实现**
8. **使用 `pkg/` 对外公开稳定接口**
9. **Handler 必须插件化，不写死 `switch`**
10. **Skill 必须 Registry 化**
11. **依赖注入，避免全局变量**



## 8. 下一步行动

### 立即可做（低成本高收益）

1. 创建 `internal/` 目录结构
2. 移动现有代码到对应领域目录
3. 定义各层 interface
4. Event Engine Handler 插件化

### 中期优化

1. 实现 Execution Engine 的重试/超时控制
2. 完善 Skill Registry
3. 添加 Policy Engine 基础框架

### 长期规划

1. 进程拆分（Server / Worker / Connector）
2. 分布式部署
3. 水平扩展

## 9. 总结

SingerOS 后端应该从：

```
MVC / service-based
```

升级为：

```
Contract-driven Service Architecture
```

**核心原则：**

- **contract 定义能力** - 系统能力的"语言"
- **service 实现能力** - 直接操作 DB
- **handler 适配输入输出** - HTTP 适配
- **db 提供执行能力** - 真实数据库操作
- **convert 保持隔离** - DTO ↔ Command ↔ Model
- **wire 组装依赖** - 避免 main 爆炸
- **无 Repository 抽象** - service 直接调用 db
