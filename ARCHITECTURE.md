
* AI OS 架构哲学
* DigitalAssistant / Agent / Skill / Workflow 体系
* 事件驱动架构
* Skills 体系（类 OpenClaw 但更先进）
* 多交互渠道（GitHub / GitLab / 企业微信 / 飞书 / App）
* 权限与安全
* Golang 工程结构（反映实际代码实现）
* 代码助手 DigitalAssistant 的落地方案
* 研发阶段路线图

---


# AI OS 架构设计文档

## 1. 项目愿景

本项目旨在构建一个 **企业级 AI 操作系统（AI OS）**，用于管理和运行 **AI Digital Assistants（AI数字员工）**。

AI OS 提供：

* AI 数字员工管理
* Agent 能力编排
* Skills / Tools / Workflow 执行
* 多渠道交互
* 企业级权限控制
* 事件驱动任务执行

核心目标：

> 让企业可以像组织员工一样管理 AI。

---

# 2. 核心设计原则

## 2.1 Event Driven Architecture

AI OS 的核心是 **事件驱动系统**：

```
Webhook
   ↓
Event
   ↓
Agent Decision
   ↓
Workflow
   ↓
Skills
   ↓
Result
```

所有 AI 行为都由 **事件触发**。

典型事件：

```
git.pr.opened
git.pr.commented
git.issue.created
wechat.message.received
feishu.message.received
app.task.created
```

---

## 2.2 DigitalAssistant 是最高抽象

系统核心结构：

```
DigitalAssistant
    ↓
Agent
    ↓
Skill
    ↓
Tool
```

职责划分：

| 层级              | 职责    |
| --------------- | ----- |
| DigitalAssistant | 数字助手  |
| Agent           | AI 决策 |
| Skill           | 能力    |
| Tool            | 外部系统  |

---

## 2.3 控制平面 vs 数据平面

SingerOS 严格分离了：

* **控制平面**（治理与管理）：Agent 注册、Skill 注册、工作流存储、租户管理、策略引擎
* **数据平面**（运行时执行）：Orchestrator、Agent Runtime、Skill Proxy、Model Router、Memory Engine、Scheduler
* **基础设施层**：数据库、消息队列、向量存储、缓存

---

# 3. 核心架构

## 3.1 架构总览

```
                +----------------------+
                |   External Systems   |
                | GitHub / GitLab     |
                | WeChat / Feishu     |
                | App / Webhook       |
                +----------+-----------+
                           |
                           v
                  +----------------+
                  | Event Gateway  |  ✅ 已实现
                  +--------+-------+
                           |
                           v
                   +--------------+
                   | Event Bus    |  ✅ RabbitMQ 已实现
                   +------+-------+
                          |
                          v
                  +---------------+
                  | Orchestrator  |  🔄 规划中
                  +-------+-------+
                          |
        +-----------------+----------------+
        |                                  |
        v                                  v

 +-------------+                   +---------------+
 | Agent Engine|  🔄 规划中         | Workflow Engine|  🔄 规划中
 +------+------+                   +-------+-------+
        |                                  |
        v                                  v
  +------------+                    +-------------+
  | Skill Proxy|  ✅ 骨架已实现      | Skill Proxy |  ✅ 骨架已实现
  +-----+------+                    +------+------+
        |                                  |
        v                                  v
     +---------+                     +----------+
     | Tools   |  🔄 规划中           | Tools    |  🔄 规划中
     +---------+                     +----------+
```

---

# 4. 核心组件

---

# 4.1 Event Gateway（✅ 已实现）

负责接收所有外部事件。

来源：

```
GitHub Webhook      ✅ 已实现（含签名验证、事件解析）
GitLab Webhook      🔄 Stub
企业微信/WeWork      🔄 Stub
飞书                 🔄 规划中
APP                  🔄 规划中
Webhook API          🔄 规划中
```

统一转换为 `interaction.Event`：

```go
type Event struct {
    EventID    string
    TraceID    string
    Channel    string
    EventType  string
    Actor      string
    Repository string
    Context    map[string]interface{}
    Payload    interface{}
    CreatedAt  time.Time
}
```

---

# 4.2 Event Bus（✅ RabbitMQ 已实现）

```
RabbitMQ（当前实现）
```

作用：

* 解耦系统
* 支持高并发
* 支持异步

Publisher/Subscriber 接口定义在 `backend/interaction/eventbus/bus.go`，RabbitMQ 实现在 `backend/interaction/eventbus/rabbitmq/`。

---

# 4.3 Orchestrator（🔄 规划中）

AI OS 的核心调度器。

职责：

```
事件 → 找到匹配的 DigitalAssistant
      → 找到 Agent
      → 执行 Workflow
```

例：

```
git.pr.opened
      ↓
CodeReviewerAssistant
      ↓
ReviewAgent
      ↓
CodeReviewWorkflow
```

---

# 5. DigitalAssistant 设计（✅ 类型已定义）

## 5.1 DigitalAssistant 结构

定义于 `backend/types/digital_assistant.go`：

```go
type DigitalAssistant struct {
    gorm.Model
    Code        string          // 助手唯一标识符
    OrgID       uint            // 所属组织
    OwnerID     uint            // 拥有者
    Name        string          // 助手名称
    Description string          // 描述
    Avatar      string          // 头像URL
    Status      string          // 状态
    Version     int             // 版本号
    Config      AssistantConfig // 配置（Runtime、LLM、Skills、Channels、Memory、Policies）
}
```

`AssistantConfig` 包含：
- `RuntimeConfig` - 执行环境类型
- `LLMConfig` - 大型语言模型配置
- `Skills []SkillRef` - 技能引用列表
- `Channels []ChannelRef` - 渠道引用列表
- `Knowledge []KnowledgeRef` - 知识库引用列表
- `MemoryConfig` - 记忆配置
- `PolicyConfig` - 策略配置

例：

```
CodeAssistantDigitalAssistant
```

职责：

```
代码审查
代码生成
PR 讨论
issue 回复
```

---

# 6. Agent 设计（🔄 规划中）

Agent 负责：

```
理解任务
规划执行
调用 skills
```

结构：

```
Agent
  id
  prompt
  model
  memory
  skillset
```

---

# 7. Workflow Engine（🔄 规划中）

Workflow 用于执行复杂流程。

示例：

```
PR Review Workflow
```

流程：

```
1 获取 PR diff
2 调用 LLM 分析
3 生成 Review Comment
4 发布 comment
```

---

# 8. Skills 系统（✅ 接口已实现）

Skills 是 AI OS 的核心能力。

设计原则：

```
Skill = 可复用能力
```

---

## 8.1 Skill 接口（backend/skills/skill.go）

```go
type Skill interface {
    Info() *SkillInfo
    Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)
    Validate(input map[string]interface{}) error
    GetID() string
    GetName() string
    GetDescription() string
}
```

`SkillInfo` 结构：

```
Skill
  id
  name
  description
  version
  category
  author
  skill_type     // local | remote
  input_schema
  output_schema
  permissions
```

嵌入 `BaseSkill` 可减少样板代码。参见 `backend/skills/examples/` 中的示例实现。

---

## 8.2 Skill 分类

### Integration Skill

外部系统能力：

```
github      ✅ webhook connector 已实现
gitlab      🔄 stub
wechat      🔄 stub
feishu      🔄 规划中
jira        🔄 规划中
```

---

### AI Skill（🔄 规划中）

AI 推理能力：

```
code_review
summarize
classification
```

---

### Tool Skill（🔄 规划中）

工具能力：

```
run_shell
execute_python
http_request
```

---

### Workflow Skill（🔄 规划中）

组合能力：

```
pr_review_workflow
bug_triage_workflow
```

---

# 9. Skill Proxy（✅ 服务骨架已实现）

Skill Proxy 提供独立的技能执行隔离。

入口：`backend/cmd/skill-proxy/main.go`

Skill Runner 设计（规划中）：

```
Skill Request
    ↓
Skill Registry
    ↓
Skill Executor
    ↓
Tool
```

---

# 10. 多交互渠道设计

系统必须支持：

```
GitHub       ✅ 已实现
GitLab       🔄 stub
企业微信      🔄 stub
飞书          🔄 规划中
App          🔄 规划中
Webhook      🔄 规划中
```

---

## 10.1 Connector 接口（✅ 已实现）

统一抽象，定义于 `backend/interaction/connectors/connector.go`：

```go
type Connector interface {
    ChannelCode() string
    RegisterRoutes(r gin.IRouter)
}
```

---

## 10.2 已实现的 Connector

**GitHub Connector**（`backend/interaction/connectors/github/`）：

- Webhook 接收（`POST /github/webhook`）
- HMAC-SHA256 签名验证
- 事件解析（目前支持 `issue_comment`）
- 事件发布至 RabbitMQ topic `interaction.github.issue_comment`

---

# 11. 权限系统（🔄 规划中）

权限控制粒度：

```
DigitalAssistant
Agent
Skill
Tool
```

权限模型：

```
RBAC + Capability
```

例：

```
CodeAssistantDigitalAssistant
    可以：
        read repo
        comment pr
    不能：
        push code
```

---

# 12. Golang 工程结构（实际代码）

```
SingerOS/
│
├── backend/
│   ├── cmd/
│   │   ├── singer/          # 主服务（HTTP + Event Gateway）
│   │   └── skill-proxy/     # Skill Proxy 服务
│   │
│   ├── config/              # 配置加载与类型定义
│   │
│   ├── gateway/             # HTTP gateway（预留扩展）
│   │
│   ├── interaction/         # 事件驱动交互层
│   │   ├── connectors/
│   │   │   ├── github/      # GitHub Webhook connector ✅
│   │   │   ├── gitlab/      # GitLab connector 🔄 stub
│   │   │   └── wework/      # 企业微信 connector 🔄 stub
│   │   ├── eventbus/
│   │   │   └── rabbitmq/    # RabbitMQ Publisher ✅
│   │   └── gateway/         # Event Gateway 路由注册
│   │
│   ├── skills/              # Skill 接口、BaseSkill、SkillManager ✅
│   │   └── examples/        # 示例 Skill 实现
│   │
│   └── types/               # 核心领域类型
│       ├── digital_assistant.go          # DigitalAssistant, AssistantConfig ✅
│       ├── digital_assistant_instance.go # DigitalAssistantInstance
│       ├── event.go                      # Event（持久化）✅
│       └── tables.go                    # DB 表名常量
│
├── proto/                   # Protobuf 定义
├── gen/                     # 生成的 proto 代码
├── frontend/                # 前端应用
├── deployments/             # Docker 构建配置
└── docs/                    # 文档
```

---

# 13. 第一个 DigitalAssistant（🔄 规划中）

## CodeAssistantDigitalAssistant

职责：

```
PR review
代码生成
issue 回复
代码解释
```

---

## 架构

```
CodeAssistantDigitalAssistant
    |
    +-- ReviewAgent
    |
    +-- CodingAgent
```

---

## Workflow

PR Review：

```
PR Opened
   ↓
Fetch Diff
   ↓
LLM Review
   ↓
Comment
```

---

# 14. opencode Skill（🔄 规划中）

opencode 用作：

```
code generation
code patch
refactor
```

Skill：

```
skill.opencode.generate_patch
skill.opencode.review
skill.opencode.fix_bug
```

---

# 15. GitHub / GitLab Hook（✅ GitHub 已实现）

触发：

```
pull_request
issue
comment
push
```

当前已支持的事件：

```
interaction.github.issue_comment   ✅
```

规划中的事件：

```
git.pr.opened
git.pr.commented
git.issue.created
```

---

# 16. 企业微信 / 飞书交互（🔄 规划中）

交互流程：

```
用户
   ↓
企业微信
   ↓
Webhook
   ↓
Event Gateway
   ↓
Agent
   ↓
Skill
   ↓
Reply
```

---

# 17. AI OS 最小 MVP

第一阶段建议：

```
1 Event Gateway          ✅ 已完成
2 Event Bus (RabbitMQ)   ✅ 已完成
3 Skill System 接口      ✅ 已完成
4 GitHub Integration     ✅ Webhook + 事件解析已完成
5 Skill Proxy 服务       ✅ 服务骨架已完成
6 Orchestrator           🔄 规划中
7 Agent Engine           🔄 规划中
8 CodeAssistantDigitalAssistant  🔄 规划中
```

---

# 18. 第一阶段功能

MVP：

```
PR 自动 Review      🔄 规划中
PR 自动总结         🔄 规划中
Issue 自动回复      🔄 规划中（GitHub issue_comment 事件已接入）
代码解释            🔄 规划中
```

---

# 19. 技术栈

| 组件 | 技术 | 状态 |
|------|------|------|
| 语言 | Golang | ✅ 已使用 |
| HTTP 框架 | Gin | ✅ 已使用 |
| CLI 框架 | Cobra | ✅ 已使用 |
| 消息队列 | RabbitMQ | ✅ 已实现 |
| ORM | GORM | ✅ 已使用（类型定义） |
| 数据库 | Postgres | 🔄 规划中 |
| 缓存 | Redis | 🔄 规划中 |
| 向量库 | Qdrant | 🔄 规划中 |
| LLM | OpenAI / Claude / DeepSeek | 🔄 规划中 |

---

# 20. 未来扩展

未来 AI OS 可以支持：

```
AI Product Manager
AI QA
AI DevOps
AI Support
AI Sales
```

---

# 总结

本 AI OS 架构的核心：

```
Event Driven
Skill Based
Agent Orchestration
DigitalAssistant
```

系统特点：

```
高度模块化
企业级权限
多渠道交互
可扩展 AI 能力
```

