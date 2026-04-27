# SingerOS 总体待办清单

> **最后更新：2026-04-25**
> **架构版本：v3.1（三核架构 + 领域驱动设计）**
>
> 本文档是SingerOS项目的总体待办清单，按优先级和阶段划分。
> 详细的Server和Worker任务分别参见：
> - [TODO_SERVER.md](./TODO_SERVER.md) - Server端具体任务
> - [TODO_WORKER.md](./TODO_WORKER.md) - Worker端具体任务

---

## 当前关注重点

1. **API接口规范** - 需要完善API接口定义，为前端开发提供明确的接口契约
2. **会话管理体系** - 需要建立Session/Conversation数据模型和管理API，支持多轮对话场景
3. **事件流转验证** - 需要验证Webhook→Event→Agent→Response完整链路的可行性
4. **架构模块完善** - 需要根据v3.1架构文档补充Execution Engine等关键模块

---

## P0 - 紧急任务（1-2周内必须完成）

### 目标：让前端可以开始开发，同时打通核心事件链路

#### P0-1: API接口定义（立即开始）

**优先级：★★★★★**  
**目的：前端不再闲着，可以基于接口定义和Mock数据开发**

- [ ] 数字员工管理API
  - 创建、查询、更新、删除、启用/停用数字员工
  - 支持分页列表查询
  - [详细任务 → TODO_SERVER.md](./TODO_SERVER.md#1-数字员工管理)
  
- [ ] 会话管理API
  - 创建会话、发送消息、查询会话历史
  - 支持流式响应（SSE）
  - [详细任务 → TODO_SERVER.md](./TODO_SERVER.md#2-会话管理)

- [ ] 项目绑定API
  - 项目CRUD、GitHub OAuth授权、账号管理
  - [详细任务 → TODO_SERVER.md](./TODO_SERVER.md#3-项目绑定)

#### P0-2: NATS Stream会话控制（第2周）

**优先级：★★★★**  
**目的：实现基于SessionID的事件路由和隔离**

- [ ] Session-based Stream策略设计
- [ ] Session Stream生命周期管理
- [ ] Consumer Group管理
- [ ] 消息TTL和清理策略
- [ ] [详细任务 → TODO_WORKER.md](./TODO_WORKER.md#1-nats-jetstream优化)

#### P0-3: 会话上下文管理（第2周）

**优先级：★★★★**  
**目的：Agent支持会话历史和上下文**

- [ ] SessionContext存储接口
- [ ] 基于Redis的短期记忆
- [ ] Agent Run时注入会话历史
- [ ] 对话轮次限制
- [ ] [详细任务 → TODO_WORKER.md](./TODO_WORKER.md#2-会话上下文管理)

---

## P1 - 重要任务（3-4周完成）

### 目标：完善架构模块，打通事件流转

#### P1-1: Execution Engine实现（架构要求）

**优先级：★★★★**  
**目的：符合v3.1架构文档要求，Event Engine与Execution Engine解耦**

- [ ] Execution Engine核心逻辑
- [ ] 任务分发器
- [ ] 同步/异步执行器
- [ ] 重试和超时控制
- [ ] Orchestrator重构使用Execution Engine
- [ ] [详细任务 → TODO_WORKER.md](./TODO_WORKER.md#3-execution-engine实现)

#### P1-2: Event Engine完善

**优先级：★★★**  
**目的：事件路由插件化，去除硬编码**

- [ ] 独立Router提取
- [ ] 内置事件处理器（PR/Issue/Push）
- [ ] [详细任务 → TODO_WORKER.md](./TODO_WORKER.md#4-event-engine完善)

#### P1-3: GitHub PR Review完整链路

**优先级：★★★★★**  
**目的：端到端验证MVP核心功能**

- [ ] PR Diff获取能力
- [ ] PR Review发布能力
- [ ] code-review Skill实现
- [ ] Webhook配置
- [ ] 端到端测试
- [ ] [详细任务 → TODO_WORKER.md](./TODO_WORKER.md#6-github-pr-review完整链路)

---

## P2 - 次要任务（5-6周完成）

### 目标：前端功能完善，Skill系统落地

#### P2-1: 前端功能开发

**优先级：★★★**  
**目的：完整的用户界面**

- [ ] 数字员工管理页面
- [ ] 会话页面（消息气泡、流式打字机）
- [ ] 会话历史侧边栏
- [ ] 项目绑定界面
- [ ] [详细任务 → TODO_SERVER.md](./TODO_SERVER.md#5-前端协同)

#### P2-2: Skill实现

**优先级：★★**  
**目的：将SKILL.md转换为可执行的Go代码**

- [ ] code-review Skill
- [ ] weather Skill
- [ ] commit-conventions Skill
- [ ] humanizer-zh Skill
- [ ] [详细任务 → TODO_WORKER.md](./TODO_WORKER.md#5-skill实现)

---

## P3 - 扩展任务（后续规划）

### 长期能力建设

- [ ] Memory系统（长期记忆/RAG）
- [ ] Workflow Engine（DAG编排）
- [ ] Policy Engine（权限控制）
- [ ] Model Router（多模型管理）
- [ ] 多租户隔离
- [ ] RBAC权限
- [ ] 成本管理
- [ ] 其他渠道Connector（GitLab、飞书、企业微信）
- [ ] Edge Runtime
- [ ] Remote Runtime
- [ ] Runtime Manager

---

## 里程碑

### Phase 1（Week 1）- 接口定义完成 ✅ 进行中
- [ ] 数字员工API定义完成
- [ ] 会话管理API定义完成
- [ ] 项目绑定API定义完成
- [ ] DTO和数据模型定义完成
- [ ] 前端可以基于Mock开始开发

**交付物：**
- TODO_SERVER.md 完整定义
- API文档
- Mock数据和服务

### Phase 2（Week 2）- 会话控制完成
- [ ] NATS Stream会话管理实现
- [ ] Session上下文持久化
- [ ] Agent支持会话历史
- [ ] 前端会话页面对接真实API

**交付物：**
- Session管理API可用
- SSE流式响应可用
- 前端可以发送消息并收到流式响应

### Phase 3（Week 3）- 事件流转打通
- [ ] Execution Engine实现
- [ ] Event Engine完善
- [ ] GitHub PR Review端到端跑通
- [ ] 前端会话功能完整

**交付物：**
- 测试仓库创建PR后自动收到AI Review
- Review包含具体的代码分析
- 无错误日志

### Phase 4（Week 4）- MVP完善
- [ ] 所有P0/P1任务完成
- [ ] Issue自动回复可用
- [ ] 前端功能完善
- [ ] 稳定性优化

**交付物：**
- MVP功能完整可用
- 系统稳定性 > 99%
- 端到端延迟 < 30秒

---

## 任务分配建议

### 后端团队（2-3人）

**后端开发A（API方向）：**
- 数字员工管理API
- 会话管理API
- 项目绑定API

**后端开发B（事件方向）：**
- Execution Engine实现
- Event Engine完善
- NATS Stream管理
- GitHub工具实现

**前端开发（1人）：**
- 数字员工管理页面
- 会话页面
- 项目绑定页面
- API对接和WebSocket/SSE

---

## 风险控制

### 关键风险

1. **API设计不合理** - 需要反复调整
   - 缓解：先出接口文档评审，再实现

2. **NATS Stream策略复杂** - 可能影响性能
   - 缓解：先实现基础版，后续优化

3. **Execution Engine与Orchestrator解耦困难**
   - 缓解：先定义清晰接口，逐步迁移

### 不允许做的事情

- ❌ 不要一开始搞Skill Proxy远程化
- ❌ 不要设计复杂Agent（Plan/Reflect）
- ❌ 不要做Memory系统（短期除外）
- ❌ 不要做多租户
- ❌ 不要做过度的Workflow Engine

---

## 成功标准

### MVP验收（Phase 3完成后）

- [ ] 可以创建一个代码助手数字员工
- [ ] 可以绑定GitHub仓库
- [ ] 可以通过GitHub PR事件触发任务
- [ ] 可以自动分析代码并生成Review
- [ ] 可以回写PR审查评论到GitHub
- [ ] 可以查看任务执行记录
- [ ] 审查质量达到人工审查的60%以上
- [ ] 端到端延迟小于30秒
- [ ] 系统稳定性大于99%

---

## 相关文档

- [ARCHITECTURE.md](./ARCHITECTURE.md) - 架构设计文档（v3.1）
- [PRD.md](./PRD.md) - 产品需求文档（v2.0）
- [TODO_SERVER.md](./TODO_SERVER.md) - Server端详细任务
- [TODO_WORKER.md](./TODO_WORKER.md) - Worker端详细任务

---

*最后更新：2026-04-25*
