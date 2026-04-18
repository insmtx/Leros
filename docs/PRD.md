# 产品需求文档 (PRD) - SingerOS

> **最后更新：2026-04-18**
> **当前阶段：MVP（最小可行产品）**

---

## 📋 MVP 产品愿景

构建一个基于事件驱动的 AI 操作系统，让企业可以像管理员工一样管理 AI 数字助手。

**MVP 核心场景：**

- GitHub PR 自动审查
- Issue 自动回复
- 代码解释和总结

---

# 1. 员工视图 - AI 工作台

## 1.1 MVP 范围（当前重点）

员工当前通过与外部平台的自然交互使用 AI：

**交互渠道：**

- ✅ GitHub (PR, Issues, Comments)
- 🔄 GitLab (规划中)
- 🔄 企业微信 (规划中)
- 🔄 飞书 (规划中)

**交互方式：**

- ✅ 自然语言评论/回复
- ✅ PR/Issue 事件触发
- ❌ 独立工作台界面（后续版本）

## 1.2 MVP 用户体验

**场景：GitHub PR 审查**

1. 开发者提交 PR 到 GitHub
2. SingerOS 自动检测 PR opened 事件
3. AI Agent 分析代码变更
4. 自动发布 Review 评论
5. 开发者在 GitHub 查看结果

**用户看到的：**

- ✅ PR 中的 AI Review 评论
- ✅ 具体的代码建议和发现
- ✅ 结构化的审查结果

---

# 2. 流程设计视图 - Workflow Studio

## 2.1 MVP 范围（❌ 后续版本）

当前通过 **Event → Orchestrator → Agent Runtime** 的自动化流程实现，无需可视化配置。

**规划中的功能：**

- ❌ 拖拽式流程图
- ❌ Agent 节点配置
- ❌ Skill 节点编排
- ❌ 条件判断和分支
- ❌ 人工审批节点
- ❌ 版本管理和回滚

**当前实现：**

- ✅ 基于代码的事件路由 (`backend/orchestrator/orchestrator.go`)
- ✅ Skills 通过文件化配置 (`backend/skills/bundled/`)
- ✅ LLM Agent 自动决策和工具调用

---

# 3. 控制中心

## 3.1 MVP 范围（🔄 部分实现）

### ✅ 已实现的管理能力：

**运行时监控：**

- ✅ 事件处理日志 (`yg-go/logs`)
- ✅ 工具调用追踪
- ✅ LLM 执行日志

**配置管理：**

- ✅ 配置文件 (`config.yaml`)
- ✅ LLM 配置（模型、API Key、参数）
- ✅ GitHub OAuth 配置
- ✅ RabbitMQ 连接配置

**数据库：**

- ✅ DigitalAssistant 模型（类型定义完成）
- ✅ Event 持久化（模型定义完成）
- ✅ User 模型（模型定义完成）
- 🔄 CRUD API（待实现）

### ❌ 规划中的功能（后续版本）：

**Agent 管理：**

- ❌ Agent 注册和配置
- ❌ Agent 运行时状态
- ❌ Agent 性能监控

**Skill 管理：**

- ❌ Skill 市场/商店
- ❌ Skill 安装和卸载
- ❌ Skill 版本管理
- ✅ Skill Catalog（当前通过代码嵌入）

**模型管理：**

- ❌ 多模型切换界面
- ❌ 模型性能对比
- ✅ LLM Router（代码实现）

**成本统计：**

- ❌ Token 用量统计
- ❌ 成本分析报表
- ❌ 预算和告警

**租户和权限：**

- ❌ 多租户管理
- ❌ RBAC 权限系统
- ✅ 基础 Auth 系统（OAuth + 账户管理）

**审计和监控：**

- ❌ 审计日志
- ❌ 执行回放
- ❌ 资源监控面板

---

# 4. MVP 核心功能清单

## 4.1 必须有的功能（✅ 已完成）

- ✅ GitHub Webhook 接收
- ✅ 事件标准化和路由
- ✅ LLM Agent 执行
- ✅ GitHub API 工具集
- ✅ OAuth 账户管理
- ✅ PR 审查技能

## 4.2 应该有的功能（🔄 进行中）

- 🔄 完整的 PR Review 流程验证
- 🔄 Issue 自动回复
- 🔄 错误处理和重试
- 🔄 详细的执行日志

## 4.3 可以有的功能（❌ 后续版本）

- ❌ 多 GitHub App 支持
- ❌ 自定义审查规则
- ❌ Review 模板配置
- ❌ 执行历史记录
- ❌ 性能指标监控

---

# 5. 技术架构

详见 [ARCHITECTURE.md](ARCHITECTURE.md)

**核心组件：**

- Event Gateway: 接收外部事件
- Event Bus: RabbitMQ 消息队列
- Orchestrator: 事件路由和分发
- Agent Runtime: Eino LLM 执行引擎
- Tools: 外部系统交互能力
- Skills: 可复用 AI 能力

---

# 6. 成功标准

## 6.1 MVP 验收标准

- [ ] 开发者创建 PR 后自动收到 AI Review
- [ ] Review 质量达到人工审查的 60% 以上
- [ ] 系统稳定性 > 99%
- [ ] 端到端延迟 < 30 秒
- [ ] 零安全事故

## 6.2 用户体验指标

- [ ] PR 审查覆盖率 > 80%
- [ ] 开发者满意度 > 4/5
- [ ] 平均响应时间 < 20 秒
- [ ] 误报率 < 10%

---

# 7. 路线图

## Phase 1: MVP（当前）

- 时间：2 周
- 目标：PR 自动审查端到端流程
- 状态：🔄 进行中

## Phase 2: 多场景扩展

- 时间：4 周
- 目标：
  - Issue 自动回复
  - 代码解释和总结
  - GitLab 集成
- 状态：❌ 规划中

## Phase 3: 企业级能力

- 时间：8 周
- 目标：
  - 多租户支持
  - 权限管理
  - 成本统计
  - 审计日志
- 状态：❌ 规划中

## Phase 4: 高级功能

- 时间：12 周+
- 目标：
  - Workflow Studio
  - AI Agent 市场
  - 自定义技能开发平台
- 状态：❌ 规划中

---

# 8. 风险和假设

## 8.1 技术风险

- LLM API 可用性和延迟
- GitHub API 速率限制
- 大规模事件处理能力

## 8.2 产品假设

- 开发者愿意接受 AI 审查
- AI Review 质量可接受
- 集成到现有工作流程无障碍

---

*最后更新：2026-04-18*
