package todo

import "github.com/insmtx/Leros/backend/internal/runtime/events"

// RuntimeTodoItem 是标准的运行时待办事项结构。
type RuntimeTodoItem = events.RuntimeTodoItem

// Status 表示待办事项的状态。
type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
	StatusCancelled  Status = "cancelled"
)

// Mode 描述解析后的待办输出是快照还是更新。
type Mode string

const (
	ModeSnapshot Mode = "snapshot"
	ModeUpdate   Mode = "update"
)

// ParseResult 是运行时待办更新的标准化解析结果。
type ParseResult struct {
	Items []RuntimeTodoItem
	Mode  Mode
	Merge bool
}
