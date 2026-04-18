
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

## 2. 当前实现状态（截至 2026-04-18）

### 已完成的核心组件

| 组件 | 状态 | 说明 |
|------|------|------|
| Event Gateway | ✅ 已完成 | GitHub webhook 接收、签名验证、事件解析 |
| Event Bus | ✅ 已完成 | RabbitMQ 集成，Publisher/Subscriber 模式 |
| Orchestrator | ✅ 已完成 | 事件路由到 Agent Runtime |
| Agent Runtime (Eino) | ✅ 已完成 | Eino-based LLM 代理执行引擎 |
| Tools 系统 | ✅ 已完成 | 工具注册表、执行运行时、权限解析 |
| Skills 系统 | ✅ 已完成 | Skill 接口、文件化技能目录、执行框架 |
| Auth 系统 | ✅ 已完成 | OAuth 流程、账户管理、凭证解析 |
| GitHub 集成 | ✅ 已完成 | Webhook、OAuth、PR 读写工具 |
| 服务集成 | ✅ 已完成 | Singer 主服务、Skill Proxy 框架 |

### 进行中的组件

| 组件 | 状态 | 说明 |
|------|------|------|
| PR 自动 Review | 🔄 进行中有 | 技能已定义，工具已实现，需要完整流程验证 |
| 多渠道扩展 | 🔄 桩代码 | GitLab、企业微信连接器框架存在 |
| DigitalAssistant 管理 | 🔄 类型定义 | 数据库模型完成，API 层待实现 |

---

# 2. 核心设计原则

## 2.1 Event Driven Architecture

AI OS 的核心是 **事件驱动系统**：

```
Webhook (GitHub/GitLab/WeChat/Feishu)
    ↓
Event Gateway (标准化事件)
    ↓
Event (统一事件结构)
    ↓
Event Bus (RabbitMQ)
    ↓
Orchestrator (事件路由)
    ↓
Agent Runtime (Eino - LLM 决策)
    ↓
Tools/Skills (能力执行)
    ↓
Result (外部系统交互)
```

所有 AI 行为都由 **事件触发**。

典型事件：

```
github.pull_request.opened       ✅ 已支持
github.pull_request.synchronize  ✅ 已支持
github.issue_comment.created      ✅ 已支持
github.push                       ✅ 已支持
gitlab.merge_request.opened      🔄 规划中
wechat.message.received          🔄 规划中
feishu.message.received          🔄 规划中
```

### 事件处理流程

实际实现的事件流转：

1. **Webhook 接收** (`backend/interaction/connectors/github/webhook.go`)
   - HMAC-SHA256 签名验证
   - 事件类型解析
   - 转换为统一 Event 结构

2. **事件发布** (`backend/interaction/eventbus/rabbitmq/`)
   - 发布到 RabbitMQ topic (如 `interaction.github.pull_request`)
   - 支持异步处理和重试

3. **事件消费** (`backend/orchestrator/orchestrator.go`)
   - 订阅对应 topic
   - 路由到 Agent Runtime

4. **Agent 执行** (`backend/runtime/eino_runner.go`)
   - 构建系统提示词（包含技能上下文）
   - LLM 决策和工具调用
   - 权限解析和凭证管理

5. **工具执行** (`backend/toolruntime/runtime.go`)
   - 根据工具名调用具体实现
   - 注入认证上下文
   - 返回执行结果

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
                   | Event Gateway  |  ✅ 已实现 (backend/interaction/)
                   | - webhook auth |     GitHub: 完整支持
                   | - event parse  |     GitLab/WeWork: Stub
                   +--------+-------+
                            |
                            v
                    +--------------+
                    | Event Bus    |  ✅ RabbitMQ (backend/interaction/eventbus/)
                    | - Publisher  |     - Publisher/Subscriber 接口
                    | - Topic      |     - Topic-based 路由
                    +------+-------+
                           |
                           v
                   +---------------+
                   | Orchestrator  |  ✅ 已实现 (backend/orchestrator/)
                   | - topic route |     - 事件类型路由
                   | - runtime disp|     - 分发到 Agent Runtime
                   +-------+-------+
                           |
         +-----------------+----------------+
         |                                  |
         v                                  v

  +-------------+                   +---------------+
  | Eino Runtime|  ✅ 已实现         | Skills Catalog|  ✅ 已实现
  | - LLM Agent |                   | - Skill Files |
  | - Tool Call |                   | - YAML Front  |
  +------+------+                   +-------+-------+
         |                                  |
         v                                  v
   +------------+                    +-------------+
   | Tools Exec |  ✅ 已实现         | Tool Registry|  ✅ 已实现
   | - Auth Res |                    | - GH Tools  |
   | - Context  |                    | - Validation|
   +-----+------+                    +------+------+
         |                                  |
         v                                  v
      +---------+                     +----------+
      | GitHub  |  ✅ 完整工具集       | External |
      | API     |  - PR Read/Write   | Systems  |
      | Tools   |  - Compare/Files   +----------+
      +---------+
```

### 核心数据流

```
GitHub Webhook PR Opened
    ↓
/github/webhook → github.Connector.handleWebhook()
    ↓
验证签名 → 解析 payload → 构建 interaction.Event
    ↓
发布到 RabbitMQ topic: interaction.github.pull_request
    ↓
Orchestrator 订阅 topic → 调用 runner.HandleEvent()
    ↓
EinoRunner 接收事件 → 构建 AgentRunner
    ↓
Agent 执行 LLM → 决定调用工具 (如 github.pr.get_metadata)
    ↓
ToolRuntime 解析账户 → 创建 GitHub Client → 执行 API 调用
    ↓
LLM 分析结果 → 生成 Review → 调用 github.pr.publish_review
    ↓
发布 PR Review 到 GitHub
```

---

# 4. 核心组件

---

# 4.1 Event Gateway（✅ 已实现）

负责接收所有外部事件并标准化为内部事件格式。

**来源支持：**

```
GitHub Webhook      ✅ 已实现（含签名验证、事件解析、OAuth）
GitLab Webhook      🔄 Stub (backend/interaction/connectors/gitlab/)
企业微信/WeWork      🔄 Stub (backend/interaction/connectors/wework/)
飞书                 ❌ 规划中
APP                  ❌ 规划中
Webhook API          ❌ 规划中
```

**统一事件结构** (`backend/interaction/event.go`)：

```go
type Event struct {
    EventID    string                 // 事件唯一标识符
    TraceID    string                 // 分布式追踪 ID
    Channel    string                 // 事件来源渠道（如 "github"）
    EventType  string                 // 事件类型（如 "pull_request"）
    Actor      string                 // 事件触发者
    Repository string                 // 关联的代码仓库
    Context    map[string]interface{} // 事件上下文信息
    Payload    interface{}            // 事件原始负载数据
    CreatedAt  time.Time              // 事件创建时间
}
```

**GitHub Connector 实现** (`backend/interaction/connectors/github/`)：

```
POST /github/webhook     - Webhook 接收端点
GET  /github/auth        - OAuth 授权发起
GET  /github/callback    - OAuth 回调处理
```

**已支持事件类型：**
- `issue_comment` (opened, created)
- `pull_request` (opened, synchronize, reopened, ready_for_review)
- `push` (commits pushed)

每个事件类型对应独立的 RabbitMQ topic，如 `interaction.github.pull_request`。

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

# 4.3 Orchestrator（✅ 已实现）

AI OS 的事件路由和调度器 (`backend/orchestrator/orchestrator.go`)。

**职责：**

```
事件消费 → 找到匹配的 Handler → 调用 Agent Runtime → 处理结果
```

**实现机制：**

```go
type Orchestrator struct {
    subscriber eventbus.Subscriber
    runner     agentruntime.Runner
    handlers   map[string]EventHandlerFunc  // topic → handler 映射
}
```

**已注册的默认处理器：**

```go
handlers[interaction.TopicGithubIssueComment] = handleIssueComment
handlers[interaction.TopicGithubPullRequest]  = handlePullRequest
handlers[interaction.TopicGithubPush]         = handlePush
```

**事件处理流程示例：**

```
git.pr.opened
    ↓
Orchestrator 根据 topic 找到 handlePullRequest
    ↓
调用 runner.HandleEvent(event)
    ↓
EinoRunner 接收事件 → 执行 LLM Agent
    ↓
LLM 决策调用工具 → 生成结果
```

---

# 5. Agent Runtime 设计（✅ Eino 实现完成）

## 5.1 Runner 接口

定义于 `backend/runtime/runner.go`：

```go
type Runner interface {
    HandleEvent(ctx context.Context, event *interaction.Event) error
}
```

这是 Orchestrator 和具体实现之间的抽象边界。

## 5.2 EinoRunner 实现

`EinoRunner` (`backend/runtime/eino_runner.go`) 是基于 CloudWeGo Eino 框架的 LLM Agent 运行时。

**核心组件：**

```go
type EinoRunner struct {
    chatModel    einomodel.ToolCallingChatModel  // LLM 模型
    toolAdapter  *runtimeeino.ToolAdapter        // 工具适配器
    skills       *runtimeprompt.SkillsContext    // 技能上下文
    tools        *runtimeprompt.ToolsContext     // 工具上下文
    systemPrompt string                          // 系统提示词
}
```

**执行流程：**

1. **事件接收** - 从 Orchestrator 接收标准化事件
2. **提示构建** - 根据事件类型构建系统提示词
3. **技能注入** - 从 Skills Catalog 加载相关技能
4. **工具注入** - 注册可用工具到 LLM
5. **LLM 执行** - 调用 LLM 进行决策和工具调用
6. **结果处理** - 处理工具输出和最终响应

**系统提示词定制：**

```go
// 根据事件类型定制提示词
switch event.EventType {
case "pull_request":
    // PR Review 专用指令
case "push":
    // Push 事件代码审查指令
default:
    // 通用指令
}
```

**LLM 配置** (`backend/config/config.go`)：

```go
type LLMConfig struct {
    Provider   string  // "openai", "claude", "deepseek"
    APIKey     string
    Model      string
    BaseURL    string
    MaxTokens  int
    Temperature float64
}
```

当前支持 OpenAI 兼容的 API（通过 `backend/runtime/eino/chatmodel.go`）。

---

# 5.5 DigitalAssistant 设计（✅ 类型已定义，🔄 API 待实现）

## 5.5.1 DigitalAssistant 结构

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
    Status      string          // 状态：active, inactive, training
    Version     int             // 版本号
    Config      AssistantConfig // 配置
}
```

`AssistantConfig` 包含：

- `RuntimeConfig` - 执行环境类型（docker, process）
- `LLMConfig` - 大型语言模型配置
- `Skills []SkillRef` - 技能引用列表
- `Channels []ChannelRef` - 渠道引用列表
- `Knowledge []KnowledgeRef` - 知识库引用列表
- `MemoryConfig` - 记忆配置（短期/长期）
- `Policies PolicyConfig` - 策略配置

**当前状态：**

- ✅ 数据库模型和类型定义完成
- ✅ GORM 自动迁移支持
- 🔄 REST API 和管理界面待实现
- 🔄 实例化和运行时绑定待实现

---

# 5.6 Auth & Account Management（✅ 已实现）

SingerOS 的认证和授权系统 (`backend/auth/`) 提供多平台账户管理能力。

## 5.6.1 核心组件

**Service** (`backend/auth/service.go`)：

```go
type Service struct {
    store           Store                    // 账户存储
    resolver        *AccountResolver         // 账户解析器
    providers       map[string]AuthorizationProvider  // OAuth providers
    authResolvers   map[string][]ProviderAuthResolver // 运行时授权解析器
}
```

**AuthorizedAccount** - 授权账户模型：

```go
type AuthorizedAccount struct {
    ID                  string
    UserID              string              // 内部用户 ID
    Provider            string              // "github", "gitlab"
    ExternalAccountID   string              // 外部平台账户 ID
    Credential          *Credential         // OAuth 凭证（加密存储）
    Metadata            map[string]string   // 额外元数据
    CreatedAt           time.Time
}
```

## 5.6.2 OAuth 流程

**1. 发起授权** (`StartAuthorization`)：

```
GET /github/auth?user_id=xxx&redirect_uri=xxx
    ↓
生成 OAuth State → 保存状态 → 构建授权 URL
    ↓
重定向到 GitHub OAuth
```

**2. 处理回调** (`HandleAuthorizationCallback`)：

```
GET /github/callback?code=xxx&state=xxx
    ↓
验证 State → 交换 Access Token → 获取用户信息
    ↓
创建/更新 AuthorizedAccount → 保存凭证
    ↓
设置默认账户绑定
```

## 5.6.3 运行时账户解析

工具执行时通过 `AuthSelector` 解析合适的账户：

```go
type AuthSelector struct {
    Provider          string            // 目标平台
    SubjectType       string            // "user", "org"
    SubjectID         string            // 主体 ID
    ScopeType         string            // 范围类型
    ScopeID           string            // 范围 ID
    ExternalRefs      map[string]string // 外部引用（如 GitHub installation_id）
}
```

解析策略：
1. 检查 ExplicitProfileID（显式指定账户）
2. 检查 ExternalRefs（如 installation_id → 关联账户）
3. 检查 SubjectID（用户默认账户）
4. 降级到系统默认账户

## 5.6.4 GitHub OAuth Provider

实现于 `backend/auth/providers/github/oauth_provider.go`：

```go
type OAuthProvider struct {
    cfg config.GithubAppConfig
}
```

- 支持 GitHub App 模式
- 自动处理 token 刷新
- 提取用户元数据（login, email, avatar）

---

---

# 6. Skills 系统（✅ 接口和实现完成）

Skills 是 SingerOS 的可复用能力单元。

## 6.1 Skill 接口

定义于 `backend/skills/skill.go`：

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

```go
type SkillInfo struct {
    ID           string
    Name         string
    Description  string
    Version      string
    Category     string
    SkillsType   string  // "local" | "remote"
    InputSchema  map[string]interface{}
    OutputSchema map[string]interface{}
    Permissions  []string
}
```

嵌入式 `BaseSkill` 可减少样板代码实现。参见 `backend/skills/examples/` 中的示例。

## 6.2 文件化 Skills（✅ 已实现）

Skills 可以定义为文件系统目录中的 `SKILL.md` 文件，使用 YAML frontmatter 定义元数据。

**文件结构** (`backend/skills/bundled/`)：

```
backend/skills/bundled/
├── github-pr-review/
│   └── SKILL.md
├── issue-triage/
│   └── SKILL.md
└── bundled.go    // embed.FS 嵌入
```

**SKILL.md 格式：**

```yaml
---
name: github-pr-review
description: Review GitHub pull requests and push events.
version: 0.1.0
metadata:
  singeros:
    category: github
    tags: [github, pr, review]
    always: true
    requires_tools:
      - github.pr.get_metadata
      - github.pr.get_files
      - github.pr.publish_review
---
# Skill 详细描述

## When to Use
...

## Operating Mode
...
```

**Catalog 系统** (`backend/skills/catalog/`)：

```go
type Catalog struct {
    fs      fs.FS
    entries map[string]*Entry
}

// 扫描嵌入的文件系统，解析所有 SKILL.md
func New(skillFS fs.FS) (*Catalog, error)
```

Skills Catalog 在启动时加载到 `EinoRunner`，用于构建 LLM 提示词上下文。

## 6.3 Skill 分类

### Integration Skills

外部系统能力：

```
github          ✅ webhook connector + OAuth 已实现
gitlab          🔄 stub connector
wechat          🔄 stub connector
feishu          ❌ 规划中
jira            ❌ 规划中
```

### AI Skills（🔄 文件化技能已实现）

AI 推理能力（通过 SKILL.md 定义）：

```
github-pr-review    ✅ 已实现（文件化 skill）
code-summarize      ❌ 待实现
issue-classify      ❌ 待实现
```

### Tool Skills（✅ 已实现）

底层工具能力（通过 `backend/tools/` 实现）：

```
github.pr.get_metadata      ✅ PR 元数据读取
github.pr.get_files         ✅ PR 文件列表
github.pr.compare_commits   ✅ Commit 对比
github.pr.publish_review    ✅ PR Review 发布
github.repo.get_file        ✅ 文件内容读取
```

### Workflow Skills（❌ 规划中）

组合能力：

```
pr_review_workflow      ❌ 规划中
bug_triage_workflow     ❌ 规划中
```

---

# 7. Tools 系统（✅ 已实现）

Tools 是 SingerOS 的底层原子能力，提供与外部系统交互的具体实现。

## 7.1 Tool 接口

定义于 `backend/tools/tool.go`：

```go
type Tool interface {
    Info() *ToolInfo
    Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)
    Validate(input map[string]interface{}) error
}

// RuntimeTool 支持带执行上下文的工具调用
type RuntimeTool interface {
    Tool
    ExecuteWithContext(ctx context.Context, execCtx *ExecutionContext, input map[string]interface{}) (map[string]interface{}, error)
}
```

`ToolInfo` 结构：

```go
type ToolInfo struct {
    Name        string       // 工具名称（唯一标识）
    Description string       // 工具描述
    Provider    string       // 提供者（如 "github"）
    ReadOnly    bool         // 是否只读（不改外部状态）
    InputSchema *Schema      // 输入 Schema（JSON Schema 格式）
}
```

## 7.2 Tool Registry

注册表 (`backend/tools/registry.go`) 管理所有工具：

```go
type Registry struct {
    mu    sync.RWMutex
    tools map[string]Tool
}

func (r *Registry) Register(tool Tool) error
func (r *Registry) Get(name string) (Tool, error)
func (r *Registry) List() []Tool
```

## 7.3 Tool Runtime

运行时 (`backend/toolruntime/runtime.go`) 处理工具执行：

```go
type Runtime struct {
    registry            *tools.Registry
    githubClientFactory *githubprovider.ClientFactory
}

func (r *Runtime) Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResult, error)
```

**执行流程：**

1. **查找工具** - 根据名称从 Registry 获取 Tool
2. **构建上下文** - 解析 Auth Account，创建 GitHub Client
3. **验证输入** - 调用 Tool.Validate()
4. **执行工具** - 调用 Execute 或 ExecuteWithContext
5. **返回结果** - 包含输出和解析的账户信息

## 7.4 GitHub Tools

实现于 `backend/tools/github/`：

| 工具名称 | 功能 | 只读 |
|----------|------|------|
| `github.account_info` | 获取当前账户信息 | ✅ |
| `github.pr.get_metadata` | 获取 PR 元数据 | ✅ |
| `github.pr.get_files` | 获取 PR 变更文件 | ✅ |
| `github.repo.compare_commits` | 对比两个 commit | ✅ |
| `github.repo.get_file` | 读取文件内容 | ✅ |
| `github.pr.publish_review` | 发布 PR Review | ❌ |

**工具使用示例**：

```go
// 在工具调用中，auth 上下文自动注入
input := map[string]interface{}{
    "repo": "owner/repo",
    "pr_number": 123,
}
result, err := runtime.Execute(ctx, &ExecuteRequest{
    ToolName: "github.pr.get_metadata",
    Input:    input,
    Selector: &auth.AuthSelector{...},
})
```

---

# 8. Skill Proxy（✅ 服务骨架已实现）

Skill Proxy (`backend/cmd/skill-proxy/`) 提供独立的技能执行隔离环境。

**用途：**

- 隔离高风险技能的执行（如 shell 执行）
- 支持多语言技能实现（Python, Node.js 等）
- 资源限制和沙箱化

**当前状态：**

- ✅ 服务框架搭建完成
- ✅ 独立进程启动
- 🔄 具体技能执行器待实现

---

# 9. 多交互渠道设计

系统必须支持多渠道交互：

```
GitHub       ✅ 完整实现（Webhook + OAuth + API Tools）
GitLab       🔄 Stub（backend/interaction/connectors/gitlab/）
企业微信      🔄 Stub（backend/interaction/connectors/wework/）
飞书          ❌ 规划中
App           ❌ 规划中
Webhook API   ❌ 规划中
```

---

## 9.1 Connector 接口（✅ 已实现）

统一抽象，定义于 `backend/interaction/connectors/connector.go`：

```go
type Connector interface {
    ChannelCode() string                    // 渠道代码，如 "github"
    RegisterRoutes(r gin.IRouter)           // 注册 HTTP 路由
}
```

---

## 9.2 GitHub Connector（✅ 完整实现）

实现于 `backend/interaction/connectors/github/`：

**组成部分：**

| 文件 | 功能 |
|------|------|
| `github.go` | Connector 主结构，路由注册 |
| `webhook.go` | Webhook 接收和签名验证 |
| `events.go` | 事件解析和转换 |
| `converter.go` | GitHub 事件 → interaction.Event |
| `client.go` | GitHub Client 封装 |

**支持的事件：**

```
interaction.github.issue_comment     ✅ issue 评论
interaction.github.pull_request      ✅ PR opened/synchronize/reopened
interaction.github.push              ✅ push 事件
```

**OAuth 集成：**

```
GET /github/auth?user_id=xxx         → 发起 GitHub OAuth
GET /github/callback?code=xxx        → 处理 OAuth 回调
```

---

# 10. 权限系统（✅ 基础实现完成）

## 10.1 权限控制粒度

```
DigitalAssistant    🔄 配置级别
Agent               🔄 执行级别
Skill               ✅ 通过 requires_tools 控制
Tool                ✅ AuthSelector 解析
```

## 10.2 AuthSelector

工具执行时的授权选择器 (`backend/auth/types.go`)：

```go
type AuthSelector struct {
    Provider          string            // 目标平台 ("github")
    SubjectType       string            // "user" | "org"
    SubjectID         string            // 主体 ID
    ScopeType         string            // "event" | "skill" | "tool"
    ScopeID           string            // 范围 ID
    ExplicitProfileID string            // 显式指定的账户
    ExternalRefs      map[string]string // 外部引用
}
```

**解析策略：**

1. 显式账户优先（ExplicitProfileID）
2. 外部引用匹配（如 installation_id → Account）
3. SubjectID 关联账户
4. 系统默认账户

## 10.3 权限模型

当前实现：**Capability-based + Account Resolution**

- 每个工具调用需要解析有效的 OAuth 凭证
- 通过 ExternalRefs 匹配安装级别的权限
- 支持多账户切换和降级

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

