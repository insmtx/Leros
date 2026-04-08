# ReAct Agent Runtime for SingerOS

该项目实现了SingerOS平台的ReAct（Reasoning + Acting）代理运行时。

## 概述

ReAct Agent采用"Reasoning + Acting"范式，结合了推理和行动能力来解决问题。它通过连续执行以下循环来工作：
1. **推理 (Reason)**：分析当前状态和目标
2. **行动 (Act)**：执行特定技能或操作
3. **观察 (Observe)**：获取行动结果  
4. 迭代直到任务完成

## 架构

```
┌─────────────────┐    ┌─────────────┐    ┌──────────────┐
│   ReActAgent    │───▶│  LLM调用    │───▶│  技能执行    │
│                 │    │             │    │              │
└─────────────────┘    └─────────────┘    └──────────────┘
         ▲                       │
         │               ┌───────▼───────┐
         └───────────────│ 状态管理器    │
                         │               │
                         └───────────────┘
```

### 核心组件

- **`ReActAgent`**：主代理执行器，管理整个推理-行动循环
- **`State`**：表示单个迭代的状态（思考、行动、观察结果）
- **`Config`**：配置代理行为的参数集合
- **`Input`**：用户查询输入封装
- **`Output`**：最终执行结果封装

## 关键功能

### 1. 推理循环
- 基于LLM的智能推理
- 状态历史跟踪
- 动态行动选择

### 2. 技能集成
- 与SingerOS技能系统的深度集成
- 技能白名单控制
- 错误处理与降级

### 3. 状态管理
- 完整的迭代历史
- 内存限制保护
- 并发访问安全

### 4. 异常处理
- 超时管理
- 错误恢复机制
- 可靠性保证

## 配置选项

```go
config := &Config{
    MaxIterations:  10,              // 最大迭代次数
    Timeout:        5 * time.Minute, // 执行超时时间
    Model:          "gpt-4",         // 使用的LLM模型
    SkillWhitelist: []string{        // 允许调用的技能白名单
        "echo.simple_echo",
        "http.request",
    },
}
```

## 使用示例

```go
// 创建代理实例
agent := NewReActAgent(config, llmProvider, skillManager, digitalAssistant)

// 准备输入
input := &Input{
    Query: "帮我总结PR#123的主要变更",
    SessionID: "session-abc-123",
    Context: map[string]interface{}{
        "repo_url": "https://github.com/...",
        "pr_id": 123,
    },
}

// 执行代理
output, err := agent.Run(ctx, input)
if err != nil {
    log.Printf("Agent execution failed: %v", err)
}
```

## 安全与控制

### 技能白名单
通过`Config.SkillWhitelist`控制可调用的技能集合。

### 迭代限制
通过`Config.MaxIterations`限制最大执行步数。

### 超时控制
使用Go context和定时器确保执行时间限制。

## 测试覆盖

本实现在设计时充分考虑了测试要求，提供了全面的测试覆盖：

### 单元测试（unit_test.go）
- [x] 核心代理创建与配置
- [x] 基本运行流程测试
- [x] 边界条件处理
- [x] 错误场景模拟

### 集成测试（integration_test.go）
- [x] 与真实LLM组件集成
- [x] 多技能协调测试
- [x] 状态传递验证

### 端到端测试（e2e_test.go）
- [x] 全流程端到端验证
- [x] 性能基准测试
- [x] 异常处理场景测试

### 状态管理测试（state_manager_test.go）
- [x] 状态历史完整性
- [x] 内存使用限制
- [x] 并发访问安全性

## 性能优化

- [x] 状态历史滚动窗口（超过`MaxStateHistorySize`时）
- [x] 无锁读取优化
- [x] JSON解析缓存机制

## 未来扩展

- [ ] 多模型混合推理支持
- [ ] 高级记忆机制
- [ ] 自适应提示工程
- [ ] 模型无关性API

## API规范

### State 结构
- `Thought`: 推理内容
- `Action`: 动作名称
- `ActionArgs`: 动作参数  
- `Observed`: 观察结果
- `Completed`: 完成标志
- `CreatedAt`: 时间戳

### 提示模板  
系统使用标准的JSON格式提示模板引导LLM输出：

```
请严格遵循以下JSON格式响应：
{
  "thought": "你的思考过程",
  "action": "动作或'finish'",
  "action_args": { ... }
}
```

遵循ReAct范式原则，平衡深度推理与实际行动，创建强大而可靠的AI智能体系统。