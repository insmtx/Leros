---
name: find-skills
description: Helps users discover and install agent skills when they ask questions like "how do I do X", "find a skill for X", "is there a skill that can...", or express interest in extending capabilities. This skill should be used when the user is looking for functionality that might exist as an installable skill.
---

# 查找与安装 Skill

此技能可帮助你从开放的 Agent 技能生态系统中发现并安装技能。

## 何时触发

当用户出现以下情况时：

- 询问"怎么实现 X"，而 X 可能是某个已有 skill 的常见任务
- 说"找一个能做 X 的 skill"或"有没有 skill 可以……"
- 问"你能不能做 X"，而 X 属于某个专门的 skill 能力
- 表达想要扩展 agent 能力的意愿
- 想要搜索工具、模板或工作流
- 提到希望在某个领域（设计、测试、部署等）获得帮助

## Skill 管理

Leros 命令行工具提供了 skill 管理功能：

```
leros skill install <identifier>  安装 skill。支持三种标识符格式：
                                  - 短名称（如 code-review）
                                  - GitHub 路径（如 owner/repo/path）
                                  - 直接 URL（如 https://.../SKILL.md）
leros skill search <query>        搜索远程 skill
```

| Flag      | 适用范围         | 说明                  |
| --------- | ---------------- | --------------------- |
| `--json`  | install / search | JSON 格式输出         |
| `--force` | install          | 覆盖已有 skill        |
| `--yes`   | install          | 跳过确认提示          |
| `--limit` | search           | 最大结果数（默认 10） |

## 如何帮助用户查找 Skill

### 第一步：理解需求

当用户请求帮助时，识别出：

1. **领域**（如 React、测试、设计、部署）
2. **具体任务**（如编写测试、创建动画、审查 PR）
3. **该任务是否常见**，以至于很可能存在现成的 skill

### 第二步：查看排行榜

在搜索之前，先查看排行榜看是否有该领域的知名 skill：

- [clawhub.ai](https://clawhub.ai/) — OpenClaw 生态的 skill 与插件注册中心
- [skills.sh leaderboard](https://skills.sh/) — 按总安装量排名的全局排行榜

排行榜展示了最受欢迎、经过实战检验的 skill，例如 Web 开发领域的热门 skill：

- `vercel-labs/agent-skills` — React、Next.js、Web 设计（10 万+ 安装）
- `anthropics/skills` — 前端设计、文档处理（10 万+ 安装）

### 第三步：搜索 Skill（主要来源）

使用 `leros skill search` 在 skill 市场中搜索：

```bash
leros skill search <query>
```

该命令会自动搜索 **多个来源**：

1. **SingerOS 服务器** — 搜索本地部署的 skill 市场（预置 skill + 已发布的 skill）
2. **skills.sh** — 搜索全局开源 skill 生态

示例：

- 用户问 "怎么让我的 React 应用更快？" → `leros skill search react performance`
- 用户问 "能帮我做 PR 审查吗？" → `leros skill search pr review`
- 用户问 "我需要生成 changelog" → `leros skill search changelog`

搜索结果的 JSON 格式包含以下字段：

- `name` — skill 名称
- `description` — 功能描述
- `source` — 来源（`Leros` 或 `SkillsSh`）
- `trust` — 可信度评分

使用 `--json` 和 `--limit` 参数可以控制输出格式和结果数量：

```bash
leros skill search --json --limit 20 <query>
```

### 第四步：验证质量

**不要仅凭搜索结果就推荐一个 skill**。始终检查：

1. **安装量** — 优先选择安装量高的 skill
2. **来源可靠性** — 来自可信来源（如 SingerOS 官方市场）的 skill 更可靠
3. **GitHub stars** — 对于来自 GitHub 的 skill，检查源仓库的 stars 数量

### 第五步：向用户展示选项

找到相关 skill 后，向用户展示：

1. Skill 名称和功能描述
2. 安装量和来源
3. 可执行的安装命令

示例回复：

```
我找到了一个可能对你有帮助的 skill！

**code-review** — AI 驱动的代码审查工具，自动发现 bug 和性能问题。
  来源：SingerOS 市场 | 安装量：5K+

安装命令：
leros skill install code-review
```

如果多个结果都相关，列出前 3-5 个让用户选择。

### 第六步：安装 Skill

如果用户决定安装，执行：

```bash
leros skill install --yes <identifier>
```

`--yes` 跳过确认提示。

**安装后验证**：确认 skill 已正确安装到 workspace：

```bash
ls <workspace_root>/.leros/skills/<name>/
```

如果安装成功，应该能看到 `SKILL.md` 文件。你也可以使用以下命令查看已安装的 skill 列表：

```bash
ls <workspace_root>/.leros/skills/
```

## 常见 Skill 分类

搜索时可参考以下常见分类：

| 分类     | 示例关键词                               |
| -------- | ---------------------------------------- |
| Web 开发 | react, nextjs, typescript, css, tailwind |
| 测试     | testing, jest, playwright, e2e           |
| DevOps   | deploy, docker, kubernetes, ci-cd        |
| 文档     | docs, readme, changelog, api-docs        |
| 代码质量 | review, lint, refactor, best-practices   |
| 设计     | ui, ux, design-system, accessibility     |
| 效率工具 | workflow, automation, git                |

## 搜索技巧

1. **使用具体的关键词**："react testing" 比只搜 "testing" 效果好
2. **同时尝试中英文搜索**：Skill 市场支持中英文语义搜索，如果一种语言没结果，换另一种语言试试
3. **尝试同义词**：如果 "deploy" 没结果，试试 "deployment" 或 "ci-cd"

## 未找到 Skill 时的处理

如果没有找到相关的 skill：

1. 告知用户没有找到现成的 skill
2. 表示可以直接用通用能力帮他们完成该任务
3. 建议用户创建一个自定义 skill（参见 `skill-creator` skill）

示例：

```
我搜索了与 "xxx" 相关的 skill，但没有找到匹配项。
不过我仍然可以直接帮你完成这个任务！需要我继续吗？

如果你经常做这件事，也可以考虑创建一个自定义 skill。
关于如何创建，请参考 skill-creator。
```
