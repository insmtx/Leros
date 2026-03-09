
* AI OS 架构哲学
* DigitalEmployee / Agent / Skill / Workflow 体系
* 事件驱动架构
* Skills 体系（类 OpenClaw 但更先进）
* 多交互渠道（GitHub / GitLab / 企业微信 / 飞书 / App）
* 权限与安全
* Golang 工程结构
* 代码助手 DigitalEmployee 的落地方案
* 研发阶段路线图

---


# AI OS 架构设计文档

## 1. 项目愿景

本项目旨在构建一个 **企业级 AI 操作系统（AI OS）**，用于管理和运行 **AI Digital Employees（AI数字员工）**。

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

## 2.2 DigitalEmployee 是最高抽象

系统核心结构：

```
DigitalEmployee
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
| DigitalEmployee | 数字员工  |
| Agent           | AI 决策 |
| Skill           | 能力    |
| Tool            | 外部系统  |

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
                  | Event Gateway  |
                  +--------+-------+
                           |
                           v
                   +--------------+
                   | Event Bus    |
                   +------+-------+
                          |
                          v
                  +---------------+
                  | Orchestrator  |
                  +-------+-------+
                          |
        +-----------------+----------------+
        |                                  |
        v                                  v

 +-------------+                   +---------------+
 | Agent Engine|                   | Workflow Engine|
 +------+------ +                  +-------+-------+
        |                                  |
        v                                  v
  +------------+                    +-------------+
  | Skill Exec |                    | Skill Exec  |
  +-----+------+                    +------+------+ 
        |                                  |
        v                                  v
     +---------+                     +----------+
     | Tools   |                     | Tools    |
     +---------+                     +----------+
```

---

# 4. 核心组件

---

# 4.1 Event Gateway

负责接收所有外部事件。

来源：

```
GitHub Webhook
GitLab Webhook
企业微信
飞书
APP
Webhook API
```

统一转换为：

```
Internal Event
```

示例：

```
event:
  type: git.pr.opened
  source: github
  repo: org/repo
  pr: 123
```

---

# 4.2 Event Bus

```
RabbitMQ | Redis Stream
```

作用：

* 解耦系统
* 支持高并发
* 支持异步

---

# 4.3 Orchestrator

AI OS 的核心调度器。

职责：

```
事件 → 找到匹配的 DigitalEmployee
      → 找到 Agent
      → 执行 Workflow
```

例：

```
git.pr.opened
      ↓
CodeReviewerEmployee
      ↓
ReviewAgent
      ↓
CodeReviewWorkflow
```

---

# 5. DigitalEmployee 设计

## 5.1 DigitalEmployee 结构

```
DigitalEmployee
  id
  name
  description
  agents[]
  permissions
  knowledge_base
  config
```

例：

```
CodeAssistantEmployee
```

职责：

```
代码审查
代码生成
PR 讨论
issue 回复
```

---

# 6. Agent 设计

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

# 7. Workflow Engine

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

# 8. Skills 系统（核心）

Skills 是 AI OS 的核心能力。

设计原则：

```
Skill = 可复用能力
```

---

## 8.1 Skill 结构

```
Skill
  id
  name
  description
  input_schema
  output_schema
  permissions
  executor
```

例：

```
skill: git.get_diff
skill: github.comment_pr
skill: opencode.generate_patch
```

---

## 8.2 Skill 分类

### Integration Skill

外部系统能力：

```
github
gitlab
wechat
feishu
jira
```

---

### AI Skill

AI 推理能力：

```
code_review
summarize
classification
```

---

### Tool Skill

工具能力：

```
run_shell
execute_python
http_request
```

---

### Workflow Skill

组合能力：

```
pr_review_workflow
bug_triage_workflow
```

---

# 9. Skill 执行架构

Skill Runner 设计：

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
GitHub
GitLab
企业微信
飞书
App
Webhook
```

---

## 10.1 Interaction Channel

统一抽象：

```
Channel
```

结构：

```
Channel
  id
  type
  auth
  skill_adapter
```

---

## 10.2 Channel Adapter

例如：

```
github_adapter
wechat_adapter
feishu_adapter
```

---

## 10.3 Interaction Skill

例如：

```
skill.github.comment
skill.wechat.reply
skill.feishu.reply
```

---

# 11. 权限系统

权限控制粒度：

```
DigitalEmployee
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
CodeAssistantEmployee
    可以：
        read repo
        comment pr
    不能：
        push code
```

---

# 12. Golang 工程结构

推荐结构：

```
aios/
│
├── cmd/
│   ├── api
│   ├── worker
│   └── scheduler
│
├── internal/
│
│   ├── core/
│   │   ├── employee
│   │   ├── agent
│   │   ├── workflow
│   │   ├── skill
│   │   └── event
│
│   ├── orchestrator/
│   │   └── orchestrator.go
│
│   ├── engine/
│   │   ├── agent_engine
│   │   ├── workflow_engine
│   │   └── skill_engine
│
│   ├── integrations/
│   │   ├── github
│   │   ├── gitlab
│   │   ├── wechat
│   │   └── feishu
│
│   ├── skills/
│   │   ├── git
│   │   ├── opencode
│   │   ├── llm
│   │   └── messaging
│
│   ├── storage/
│   │   ├── postgres
│   │   ├── redis
│   │   └── vector
│
│   ├── eventbus/
│   │   └── nats
│
│   ├── auth/
│   │   └── permissions
│
│   └── config/
│
├── pkg/
│
│   ├── sdk
│   └── client
│
├── api/
│   └── proto
│
├── deployments/
│   └── docker
│
└── docs/
```

---

# 13. 第一个 DigitalEmployee

## CodeAssistantEmployee

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
CodeAssistantEmployee
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

# 14. opencode Skill

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

# 15. GitHub / GitLab Hook

触发：

```
pull_request
issue
comment
push
```

事件：

```
git.pr.opened
git.pr.commented
git.issue.created
```

---

# 16. 企业微信 / 飞书交互

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
1 Event Gateway
2 Orchestrator
3 Skill System
4 GitHub Integration
5 CodeAssistantEmployee
```

---

# 18. 第一阶段功能

MVP：

```
PR 自动 Review
PR 自动总结
Issue 自动回复
代码解释
```

---

# 19. 技术栈

推荐：

```
语言
Golang

消息系统
NATS

数据库
Postgres

缓存
Redis

向量库
Qdrant

LLM
OpenAI / Claude / DeepSeek
```

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
DigitalEmployee
```

系统特点：

```
高度模块化
企业级权限
多渠道交互
可扩展 AI 能力
```
