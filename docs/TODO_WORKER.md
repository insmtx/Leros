# TODO_WORKER - Worker端详细任务清单

> **最后更新：2026-04-25**
>
> 本文档定义Worker端的具体开发任务，按模块和优先级划分。

---

## Worker职责定义

根据v3.1架构文档，Worker端负责：

1. **控制面（Control Plane）**
   - Event Engine：事件路由与分发
   - Memory：会话上下文管理
   - Policy Engine：权限控制

2. **执行面（Execution Plane）**
   - Execution Engine：任务执行调度
   - Agent Runtime：LLM推理和工具调用
   - Skill执行：技能调用和执行

**注意：** 当前代码中Orchestrator（Event Engine）在Server中运行，Phase 2阶段需要逐步分离到Worker进程。

---

## 1. NATS JetStream优化

**优先级：P0-2**

### 目标

实现基于Session的事件路由和隔离，优化NATS JetStream的Stream策略。

### 需要完成的工作

1. **Stream策略设计**
   - 设计Session-based Stream策略
   - 设计Events Stream策略

2. **Session Stream管理器实现**
   - 实现为Session创建Stream的能力
   - 实现删除Session Stream的能力
   - 实现发布用户消息的能力
   - 实现发布助手消息的能力
   - 实现发布流式事件的能力
   - 实现订阅Session事件的能力

3. **Consumer Group管理**
   - 定义持久化Consumer命名规范
   - 实现Consumer配置管理
   - 实现Consumer健康检查

4. **消息TTL和清理策略**
   - 实现Stream级别Retention策略
   - 实现定期清理过期Session
   - 实现消息压缩（可选）

5. **集成**
   - Session创建时初始化Stream
   - Session关闭时清理资源

6. **测试**
   - 编写单元测试

---

## 2. 会话上下文管理

**优先级：P0-3**

### 目标

实现会话上下文的存储和管理，支持Agent在运行时注入会话历史。

### 需要完成的工作

1. **会话存储接口定义**
   - 定义会话存储接口
   - 定义会话上下文结构

2. **基于Redis的短期记忆实现**
   - 实现Redis存储
   - 实现会话数据序列化
   - 实现TTL管理
   - 实现消息历史限制

3. **Agent集成**
   - Agent Run时加载Session历史
   - 构建Messages列表
   - 注入到请求上下文

4. **对话轮次限制**
   - 实现对话轮次限制配置
   - 实现超出限制时的截断逻辑
   - 可选：实现摘要压缩

5. **测试**
   - 编写单元测试
   - 进行集成测试（需要Redis）

---

## 3. Execution Engine实现

**优先级：P1-1（架构核心）**

**重要性：** 这是v3.1架构文档要求的关键模块，当前代码中缺失。Execution Engine负责将Event Engine与Agent Runtime解耦。

### 目标

实现Execution Engine，负责任务执行调度，与Event Engine解耦。

### 需要完成的工作

1. **目录结构创建**
   - 创建execution目录和相关子目录

2. **接口定义**
   - 定义Execution Engine接口
   - 定义Executor接口
   - 定义Task类型

3. **Engine核心实现**
   - 实现执行器注册表
   - 实现任务分发逻辑
   - 实现并发控制
   - 实现执行历史追踪（可选）

4. **同步执行器实现**
   - 实现同步执行任务（阻塞直到完成）
   - 集成超时控制
   - 集成错误处理

5. **异步执行器实现**
   - 实现异步执行任务（立即返回，后台执行）
   - 实现任务队列
   - 实现任务状态追踪

6. **重试控制实现**
   - 实现重试策略配置
   - 实现指数退避算法
   - 实现重试条件判断

7. **超时控制实现**
   - 实现超时Context创建
   - 实现超时自动取消
   - 实现超时错误返回

8. **与Orchestrator集成**
   - 重构Orchestrator使用Execution Engine
   - 创建AgentRun执行器
   - 创建Skill执行执行器

9. **测试**
   - 编写单元测试
   - 进行集成测试

---

## 4. Event Engine完善

**优先级：P1-2**

### 目标

完善Event Engine，实现事件路由插件化，去除硬编码。

### 需要完成的工作

1. **独立Router提取**
   - 实现事件路由器
   - 实现注册Handler能力
   - 实现获取Handler能力
   - 实现路由事件到Handler能力

2. **内置Handler目录和实现**
   - 创建builtins目录
   - 定义Handler接口

3. **PR Handler实现**
   - 实现解析PR事件（opened、updated、closed）
   - 实现提取PR信息
   - 实现创建Agent Run任务
   - 实现调用Execution Engine执行

4. **Issue Handler实现**
   - 实现解析Issue事件
   - 实现提取Issue信息
   - 实现创建Agent Run任务
   - 实现调用Execution Engine执行

5. **Push Handler实现**
   - 实现解析Push事件
   - 实现提取commits信息
   - 可选：实现触发代码审查或CI

6. **Orchestrator重构**
   - 重构Orchestrator使用Router
   - 注册内置Handler
   - 删除硬编码switch-case

7. **测试**
   - 编写单元测试
   - 进行集成测试

---

## 5. Skill实现

**优先级：P2-2**

### 目标

将SKILL.md转换为实际可执行的Go代码，实现核心Skills。

### 需要完成的工作

1. **code-review Skill实现**
   - 实现代码审查Skill
   - 实现Prompt构建
   - 实现结果解析

2. **weather Skill实现（示例）**
   - 实现天气查询Skill
   - 集成天气API

3. **commit-conventions Skill实现**
   - 实现commit message格式验证Skill
   - 实现格式验证逻辑

4. **humanizer-zh Skill实现**
   - 实现中文表达优化Skill
   - 实现文本优化逻辑

5. **Skill注册**
   - 实现Skill注册功能
   - 实现自动加载skills目录下的所有Skill

6. **测试**
   - 编写单元测试
   - 进行集成测试

---

## 6. GitHub PR Review完整链路

**优先级：P1-3（MVP核心）**

### 目标

端到端验证GitHub PR自动审查流程，确保从Webhook接收到Review发布的完整链路正常工作。

### 端到端流程

1. GitHub Webhook → Server接收
2. 事件标准化 → 发布到NATS
3. Event Engine消费 → Router路由到PRHandler
4. PRHandler提取信息 → 创建Task
5. Execution Engine执行 → AgentRunExecutor
6. Agent Run → 调用code-review Skill或LLM
7. Skill执行 → 调用获取PR diff的能力
8. LLM分析 → 生成审查意见
9. Skill执行 → 调用发布PR Review的能力
10. 结果返回 → 发布Review到GitHub

### 需要验证的点

- [ ] Webhook签名验证正确
- [ ] 事件正确发布到NATS
- [ ] Event Engine正确消费事件
- [ ] PRHandler正确提取PR信息
- [ ] Agent正确执行code-review
- [ ] 能力调用成功（需要GitHub Token）
- [ ] Review正确发布到GitHub
- [ ] 审查意见质量可接受

### 需要完成的工作

1. **配置准备**
   - 配置GitHub Webhook订阅
   - 准备测试仓库
   - 配置OAuth Token

2. **验证各组件**
   - 验证Webhook签名验证
   - 验证事件发布
   - 验证PRHandler执行
   - 验证code-review Skill
   - 验证PR Diff能力
   - 验证Publish Review能力

3. **端到端测试**
   - 创建测试PR进行端到端测试
   - 验证Review发布结果
   - 检查日志无错误
   - 进行性能测试（延迟<30秒）

---

## 7. Worker进程分离（后续）

**优先级：P3**

### 目标

将Orchestrator从Server进程分离到独立Worker进程。

### 需要完成的工作

1. **Worker启动逻辑实现**
   - 实现Worker启动逻辑
   - 加载配置
   - 初始化NATS Subscriber
   - 初始化Execution Engine
   - 初始化Orchestrator

2. **Server端调整**
   - Server端移除Orchestrator
   - Server端保留Publisher

3. **部署配置更新**
   - 更新部署配置
   - 更新Docker Compose

4. **测试**
   - 测试独立Worker

---

## 8. 测试和文档

**优先级：P2**

### 目标

为Worker端各模块编写充分的测试和文档。

### 需要完成的工作

1. **单元测试**
   - Execution Engine单元测试
   - Event Engine单元测试
   - Session Store单元测试
   - Skill单元测试

2. **集成测试**
   - NATS集成测试
   - Execution Engine集成测试
   - Event Engine集成测试
   - PR Review链路集成测试

3. **文档**
   - Worker架构说明
   - Execution Engine设计文档
   - Event Engine设计文档
   - NATS Stream设计规范
   - 部署指南

---

## 依赖关系

- NATS Stream管理 → Session上下文 → Agent Run注入历史
- Execution Engine → SyncExecutor/AsyncExecutor → AgentRunExecutor/SkillExecutor
- Event Engine Router → PR/Issue/Push Handler → AgentRunExecutor → Agent Runtime → 能力调用
- Session上下文 → Agent Runtime

---

## 与Server端的协作点

| Worker任务 | 依赖的Server任务 | 说明 |
|-----------|-----------------|------|
| Session上下文管理 | Session数据模型 | 共用Session ID |
| NATS Stream管理 | Session创建API | Session创建时初始化Stream |
| Agent Run注入历史 | SendMessage API | SendMessage触发Agent Run |
| Execution Engine | Assistant Service | Execute调用Agent Runner |
| PR Handler | GitHub Webhook | Webhook发布事件到NATS |
| PR Review链路 | GitHub PR能力 | 能力实现在backend/tools/ |

---

*最后更新：2026-04-25*
