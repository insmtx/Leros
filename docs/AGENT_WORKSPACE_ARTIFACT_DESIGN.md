# Agent Workspace 与最终产物设计

## 1. 背景与目标

本文档定义 Agent 运行时工作区、任务多轮对话产物归属、最终产物声明和附件访问的 v1 方案。

v1 目标：

- 为每个 project/task 提供隔离的 Agent 工作区。
- 支持同一个 task 内多轮对话，并能识别每一轮生成了哪些最终产物。
- 最终产物通过 manifest 显式声明，避免扫描目录后猜测。
- 产物记录可持久化，并通过任务或消息维度接口返回给前端。
- 下载接口只暴露 artifact id，不暴露容器内真实路径。

v1 不做：

- 不实现正文中的 inline 附件展示。
- 不实现本地 artifact 文件快照目录。
- 不通过 `RuntimeOptions` 传递 `ProjectDir`、`OutputDir`、`TmpDir`、`ArtifactManifestPath` 等内部路径。
- 不设计大量 workspace 环境变量。

## 2. 核心原则

### 2.1 工作区按 project/task 隔离

项目之间必须物理隔离。不同 `project_id` 的 Agent 运行目录不能互相访问。

同一个 task 可能发生多轮对话，因此 task 工作区需要保持稳定，承载连续上下文、文件修改和项目资料。

### 2.2 产物归属按 request/turn 区分

同一个 task 内每一轮用户请求都可能生成产物。为了回答“哪一轮生成了什么”，每轮执行必须有独立 turn 目录。

v1 约定：

```text
turn_id = request_id
```

后续 assistant message 创建完成后，artifact 记录再绑定最终 `message_id`。

### 2.3 路径由内部 resolver 统一计算

workspace 路径是系统内部状态，不应作为 runtime API 参数层层传递。

统一由内部 workspace resolver 根据以下业务标识计算：

```text
org_id
project_id
task_id
request_id
```

`RuntimeOptions` 只保留运行控制参数，例如：

```go
type RuntimeOptions struct {
    Kind    string `json:"kind,omitempty"`
    WorkDir string `json:"work_dir,omitempty"`
    MaxStep int    `json:"max_step,omitempty"`
}
```

### 2.4 最终产物必须显式声明

系统不通过扫描目录自动判断最终产物。扫描只能用于校验和补充元数据。

只有写入本轮 `artifacts.jsonl` 且 `is_final: true` 的文件才进入最终产物列表。

## 3. 目录结构

workspace root 沿用现有 `LEROS_WORKSPACE_ROOT`。

v1 目录结构：

```text
{LEROS_WORKSPACE_ROOT}/
  projects/
    {org_id}/
      {project_id}/
        repo/
          .git/
          .leros/
            tasks/
              {task_id}/
                turns/
                  {request_id}/
                    tmp/
                    logs/
                    artifacts.jsonl
```

路径职责：

| 路径 | 职责 |
| --- | --- |
| `repo/` | Agent CLI 默认工作目录，也是项目 Git 工作区 |
| `repo/.git/` | 项目 Git 管理目录 |
| `repo/.leros/tasks/{task_id}/turns/{request_id}/tmp/` | 本轮临时文件 |
| `repo/.leros/tasks/{task_id}/turns/{request_id}/logs/` | 本轮日志 |
| `repo/.leros/tasks/{task_id}/turns/{request_id}/artifacts.jsonl` | 本轮最终产物声明 |

`.leros/` 是运行态目录，必须写入 `repo/.git/info/exclude`，不进入项目 Git。

v1 不创建也不依赖以下本地快照目录：

```text
repo/.leros/tasks/{task_id}/artifacts/{artifact_id}/
```

该能力仅作为后续扩展预留。

## 4. Workspace Resolver

建议由一个内部包统一负责 workspace 计算与准备，例如：

```text
backend/internal/workspace
```

核心输入：

```go
type TaskWorkspaceRequest struct {
    OrgID            uint
    ProjectID        string
    TaskID           string
    RequestID        string
    RequestedWorkDir string
}
```

核心输出：

```go
type TaskWorkspace struct {
    WorkspaceRoot        string
    ProjectRoot          string
    RepoDir              string
    TaskDir              string
    TurnDir              string
    TurnTmpDir           string
    TurnLogDir           string
    ArtifactManifestPath string
    EffectiveWorkDir     string
}
```

resolver 职责：

- 创建 project repo、task 目录、turn 目录。
- 初始化 `repo/.git`。
- 将 `.leros/` 写入 `repo/.git/info/exclude`。
- 校验 `runtime.work_dir`。
- 返回安全后的 `EffectiveWorkDir`。
- 提供 manifest 路径、turn tmp/log 路径等内部派生路径。

## 5. work_dir 约束

`runtime.work_dir` 是用户或上游请求指定的期望工作目录，但不能直接信任。

规则：

- 空值：使用 `repo/`。
- 相对路径：解析为 `repo/` 内子目录。
- 绝对路径：必须位于当前 project repo 内。
- 禁止 `..` 逃逸。
- 禁止软链逃逸。
- 禁止跨 project workspace。
- 禁止把工作目录指向 `.leros/tasks/{task_id}/turns/{request_id}/tmp`、logs 或其他运行态目录。

Agent CLI 实际执行目录使用 resolver 返回的 `EffectiveWorkDir`。

## 6. 执行流程

1. Server 收到用户消息，创建 request。
2. Server 创建或定位 project、task、session，并生成 `request_id`。
3. Server 发布 worker task，携带 `org_id`、`project_id`、`task_id`、`request_id`。
4. Worker 收到 task 后调用 workspace resolver，准备 project repo、task 目录和 turn 目录。
5. Worker 校验并解析 `runtime.work_dir`，得到 `EffectiveWorkDir`。
6. Agent CLI 在 `repo/` 或 repo 内子目录执行。
7. Agent 将需要交付的最终文件写入 project repo 内。
8. Agent 将本轮最终产物声明写入 `turns/{request_id}/artifacts.jsonl`。
9. Worker 在完成事件前读取本轮 manifest。
10. Worker 校验产物路径、文件存在性、mime type、file size 和 sha256。
11. Worker 将 final artifacts 放入 completed payload。
12. Server 创建最终 assistant message 时持久化 artifact 记录，并绑定 `task_id`、`session_id`、`message_id`、`request_id`。
13. 前端通过 task/message artifact 接口查询产物列表，通过 artifact download 接口下载。

## 7. Agent 产物声明

v1 采用 JSON Lines manifest。每一行表示一个产物声明。

示例：

```json
{"path":"HZSM_case_knowledgebase_reordered_v6_no_h3c.pptx","title":"HZSM_case_knowledgebase_reordered_v6_no_h3c.pptx","description":"最终修改后的 PPT 文件","mime_type":"application/vnd.openxmlformats-officedocument.presentationml.presentation","artifact_type":"file","is_final":true}
```

字段：

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| `path` | 是 | 相对 project repo 的文件路径 |
| `title` | 否 | 前端展示名，空值时使用文件名 |
| `description` | 否 | 产物说明 |
| `mime_type` | 否 | 空值时由系统探测 |
| `artifact_type` | 否 | 默认 `file` |
| `is_final` | 是 | 只有 `true` 才进入最终产物列表 |

manifest 路径由内部 resolver 计算。系统提示或 wrapper 可以告知 Agent “本轮最终产物声明文件在哪里”，但不应暴露整套 workspace 路径参数。

如后续希望完全避免环境变量，可提供内部工具：

```text
artifact_declare(path, title, description, mime_type)
```

工具内部写入本轮 manifest。

## 8. 产物收集规则

Worker 完成任务时读取当前 turn 的 `artifacts.jsonl`。

收集规则：

- 只处理 `is_final: true` 的记录。
- `path` 必须是相对 project repo 的路径。
- 不允许绝对路径。
- 不允许 `..`。
- 不允许软链逃逸。
- 不允许指向 `.git`、`.leros`、tmp、logs 等运行态目录。
- 文件必须真实存在。
- 文件不能是目录。
- 系统补充 `mime_type`、`file_size`、`sha256`。
- 同一路径重复声明时，保留最后一次有效声明。

未声明文件：

- 不进入最终产物列表。
- 不展示给用户。
- 可作为日志或调试信息记录，但不作为 artifact 持久化。

## 9. 持久化模型

Artifact 记录需要支持任务维度、消息维度和请求轮次维度。

建议字段：

```text
artifact_id
org_id
project_id
task_id
session_id
message_id
request_id
relative_path
mime_type
file_size
sha256
source
status
created_at
```

字段说明：

| 字段 | 说明 |
| --- | --- |
| `artifact_id` | 对外稳定 ID |
| `task_id` | 所属 task |
| `session_id` | 所属 session |
| `message_id` | 所属 assistant message |
| `request_id` | 所属 turn |
| `relative_path` | 相对 project repo 的路径 |
| `sha256` | 文件内容 hash |
| `source` | v1 默认为 `agent_declared` |
| `status` | `completed`、`failed` 等 |

同一个 task 的多轮产物通过 `request_id` 或 `message_id` 区分。

## 10. API 设计

v1 预留接口：

```text
GET /v1/tasks/{task_id}/artifacts
GET /v1/tasks/{task_id}/artifacts?group_by=turn
GET /v1/messages/{message_id}/artifacts
GET /v1/artifacts/{artifact_id}/download
```

任务产物列表返回示例：

```json
[
  {
    "artifact_id": "art_xxx",
    "task_id": "task_xxx",
    "message_id": "123",
    "request_id": "req_xxx",
    "title": "HZSM_case_knowledgebase_reordered_v6_no_h3c.pptx",
    "artifact_type": "file",
    "mime_type": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
    "file_size": 123456,
    "sha256": "...",
    "download_url": "/v1/artifacts/art_xxx/download",
    "created_at": "2026-05-26T12:00:00Z"
  }
]
```

下载接口 v1：

- 根据 `artifact_id` 查询 artifact 记录。
- 校验当前用户是否可访问对应 org/project/task。
- 根据 project repo 和 `relative_path` 解析真实文件路径。
- 校验路径仍在当前 project repo 内。
- 返回文件内容。
- 不向前端暴露容器绝对路径。

## 11. 多轮对话产物归属

同一个 task 内可能发生多轮对话：

```text
task_1
  request_1 -> assistant message A -> artifacts: a.pptx
  request_2 -> assistant message B -> artifacts: b.xlsx
  request_3 -> assistant message C -> artifacts: a.pptx
```

每轮产物记录都绑定：

```text
task_id + request_id + message_id
```

因此系统可以回答：

- 某个 task 一共生成过哪些产物。
- 某一轮 request 生成了哪些产物。
- 某条 assistant message 底部应该展示哪些附件。
- 同名文件在不同轮次分别由哪次生成。

## 12. v1 历史文件限制

v1 不实现本地文件快照目录，因此如果后续轮次覆盖了同名文件，历史 artifact 的 `relative_path` 可能指向被覆盖后的当前文件。

v1 仍记录：

```text
request_id
message_id
relative_path
sha256
created_at
```

这些字段用于识别产物归属和检测内容是否已变化。

如果需要冻结历史文件内容，应进入后续版本，通过 Git、S3 或本地快照目录实现。

## 13. 后续扩展

### 13.1 Git 历史文件

后续可在任务完成时提交或记录 blob。

artifact 记录增加：

```text
git_commit
git_blob
```

下载时按 commit/blob 读取历史版本，避免同名文件覆盖问题。

### 13.2 S3 / MinIO

后续可在任务完成后上传 artifact 文件到对象存储。

artifact 记录增加或使用：

```text
storage_backend = s3
storage_key = ...
```

下载接口保持不变，前端无感知。

### 13.3 本地文件快照目录

如需本地冻结历史文件，可后续引入：

```text
repo/.leros/tasks/{task_id}/artifacts/{artifact_id}/
```

该目录不属于 v1 实现范围。

### 13.4 正文 inline artifact

后续可通过结构化 message chunk 或 annotation，在 assistant 正文中展示文件链接。

v1 只要求最终产物列表可查询和下载。

### 13.5 attempt / retry

如果同一个 request 需要多次执行，可扩展：

```text
turns/{request_id}/attempts/{attempt_id}/
```

v1 暂不引入 attempt 维度。

## 14. 验收标准

- 工作区目录使用 `projects/{org_id}/{project_id}/repo`。
- 多轮产物声明目录使用 `tasks/{task_id}/turns/{request_id}`。
- 文档不要求 v1 实现本地 artifact 快照目录。
- 文档明确内部路径不通过 `RuntimeOptions` 传递。
- 文档明确最终产物来自 manifest 显式声明。
- 文档明确 v1 历史文件冻结由后续 Git/S3/本地快照方案解决。
