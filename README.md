# CollieOS 🐕

**CollieOS** 是一个面向复杂任务的多智能体编排框架（Multi-Agent Orchestration Framework），提供操作员驱动的目标输入、编排层规划与执行层协作三层架构。

**CollieOS** is a multi-agent orchestration framework for complex task execution, featuring a three-tier architecture: operator-driven goal input, an orchestration layer for planning, and an execution layer for parallel agent collaboration.

---

## 架构概览 / Architecture Overview

```
  Operator  ──── (goal / constraints) ────▶  Orchestrator Agent (Collie)
                                                        │
                              ┌─────────────────────────┼──────────────────────┐
                              ▼                         ▼                      ▼
                        Executor Agent A        Executor Agent B        Executor Agent C
                         (sub-task)              (sub-task)              (sub-task)
```

| 层级 / Layer | 角色 / Role | 职责 / Responsibility |
|---|---|---|
| **操作员层 Operator** | 人类用户或上游系统 | 定义目标、约束与验收标准 / Define goals, constraints, and acceptance criteria |
| **编排层 Orchestrator** | Collie Agent | 任务分解、执行规划、状态监控、结果聚合 / Task decomposition, planning, monitoring, result aggregation |
| **执行层 Executor** | Worker Agents | 执行具体子任务，上报状态与结果 / Execute sub-tasks and report status/results |

---

## 核心特性 / Core Features

- **目标驱动的任务分解** — 编排层将高层目标自动拆解为可执行的子任务图（DAG），动态调度执行顺序
- **人机协同（Human-in-the-Loop）** — 操作员可在任意检查点注入反馈、修正约束或中止流程
- **多 Agent 并行协作** — 执行层支持多个独立 Agent 并发运行，由编排层统一管理依赖与状态
- **可审计的决策链** — 编排层的每次规划、调度与状态变更均记录可追溯日志

---

- **Goal-driven task decomposition** — The Orchestrator decomposes high-level goals into a directed acyclic graph (DAG) of sub-tasks and dynamically schedules their execution order
- **Human-in-the-Loop** — Operators can inject feedback, modify constraints, or abort execution at any checkpoint
- **Parallel multi-agent collaboration** — The execution layer supports concurrent independent agents with dependency and state management handled by the Orchestrator
- **Auditable decision trail** — Every planning step, dispatch decision, and state transition is logged for traceability

---

## 快速开始 / Getting Started

> 🚧 项目正在积极开发中，更多文档和示例即将到来。
> 🚧 The project is under active development. More documentation and examples coming soon.

---

## 许可证 / License

本项目遵循 [GNU General Public License v3.0](LICENSE) 开源协议。

This project is licensed under the [GNU General Public License v3.0](LICENSE).