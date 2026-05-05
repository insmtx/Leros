package dm

// TaskType 表示请求 Worker 执行的任务类型。
type TaskType string

const (
	// TaskTypeAgentRun 表示请求 Worker 执行一次 Agent 运行。
	TaskTypeAgentRun TaskType = "agent.run"
)

// InputType 表示任务输入的主要形态。
type InputType string

const (
	// InputTypeMessage 表示普通对话消息输入。
	InputTypeMessage InputType = "message"
	// InputTypeTaskInstruction 表示直接任务指令输入。
	InputTypeTaskInstruction InputType = "task_instruction"
)

// MessageRole 表示对话或流式消息的产生者角色。
type MessageRole string

const (
	// MessageRoleUser 表示人类用户或外部用户消息。
	MessageRoleUser MessageRole = "user"
	// MessageRoleAssistant 表示助手消息。
	MessageRoleAssistant MessageRole = "assistant"
	// MessageRoleSystem 表示系统消息。
	MessageRoleSystem MessageRole = "system"
	// MessageRoleTool 表示工具结果消息。
	MessageRoleTool MessageRole = "tool"
)

// WorkerTaskMessage 是 Server 发送给 Worker 的任务消息协议。
type WorkerTaskMessage = Envelope[WorkerTaskBody]

// WorkerTaskBody 是 Server 发送给 Worker 的任务消息载荷。
type WorkerTaskBody struct {
	TaskType TaskType `json:"task_type"`

	Actor     ActorContext    `json:"actor"`
	Execution ExecutionTarget `json:"execution"`
	Input     TaskInput       `json:"input"`

	Runtime RuntimeOptions `json:"runtime,omitempty"`
	Policy  TaskPolicy     `json:"policy,omitempty"`
}

// ActorContext 描述任务发起方身份。
type ActorContext struct {
	UserID      string `json:"user_id,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Channel     string `json:"channel,omitempty"`
	ExternalID  string `json:"external_id,omitempty"`
	AccountID   string `json:"account_id,omitempty"`
}

// ExecutionTarget 描述本次任务选择的执行目标和能力范围。
type ExecutionTarget struct {
	AssistantID string   `json:"assistant_id,omitempty"`
	AgentID     string   `json:"agent_id,omitempty"`
	Skills      []string `json:"skills,omitempty"`
	Tools       []string `json:"tools,omitempty"`
}

// TaskInput 是 Worker Runtime 消费的标准化任务输入。
type TaskInput struct {
	Type        InputType      `json:"type"`
	Text        string         `json:"text,omitempty"`
	Messages    []ChatMessage  `json:"messages,omitempty"`
	Attachments []Attachment   `json:"attachments,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// ChatMessage 是紧凑的对话消息快照。
type ChatMessage struct {
	Role    MessageRole `json:"role"`
	Content string      `json:"content"`
}

// Attachment 描述任务输入中可用的附件。
type Attachment struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
	URL      string `json:"url,omitempty"`
}

// RuntimeOptions 控制 Worker Runtime 的执行参数。
type RuntimeOptions struct {
	Kind    string `json:"kind,omitempty"`
	WorkDir string `json:"work_dir,omitempty"`
	MaxStep int    `json:"max_step,omitempty"`
}

// TaskPolicy 承载 Worker 任务需要遵守的策略开关。
type TaskPolicy struct {
	RequireApproval bool `json:"require_approval,omitempty"`
}
