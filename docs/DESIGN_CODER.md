# Leros 代码助手架构设计

本文档描述 Leros 中代码助手（CodeAssistantDigitalAssistant）相关的业务架构设计。

## 1. CodeAssistantDigitalAssistant

职责：
- PR 自动审查
- 代码生成
- Issue 自动回复
- 代码解释

### 1.1 架构

```
CodeAssistantDigitalAssistant
    |
    +-- ReviewAgent
    |
    +-- CodingAgent
```

### 1.2 Workflow

PR Review 流程：

```
PR Opened
    ↓
Fetch Diff
    ↓
LLM Review
    ↓
Comment
```

## 2. opencode Skill

opencode 用作：
- 代码生成
- 代码补丁
- 代码重构

Skill 定义：
- `skill.opencode.generate_patch`
- `skill.opencode.review`
- `skill.opencode.fix_bug`

## 3. GitHub / GitLab Hook 触发

触发事件：
- `pull_request`
- `issue`
- `comment`
- `push`

当前已支持的事件：
- `interaction.github.issue_comment` ✅
- `interaction.github.pull_request` ✅
- `interaction.github.push` ✅

规划中的事件：
- `git.pr.opened`
- `git.pr.commented`
- `git.issue.created`

## 4. 第一阶段 MVP 功能

- PR 自动 Review
- PR 自动总结
- Issue 自动回复
- 代码解释
