package contract

import (
	"time"

	"github.com/insmtx/SingerOS/backend/types"
)

// CreateSessionRequest 创建会话请求
type CreateSessionRequest struct {
	SessionID     string                       `json:"session_id,omitempty"`
	Type          string                       `json:"type" binding:"required"`
	UserID        uint                         `json:"user_id,omitempty"`
	AssistantID   uint                         `json:"assistant_id,omitempty"`
	AssistantCode string                       `json:"assistant_code,omitempty"`
	Title         string                       `json:"title,omitempty"`
	Metadata      *types.SessionMetadata       `json:"metadata,omitempty"`
	ExpiredAt     *time.Time                   `json:"expired_at,omitempty"`
}

// UpdateSessionRequest 更新会话请求
type UpdateSessionRequest struct {
	Title         string                       `json:"title,omitempty"`
	Metadata      *types.SessionMetadata       `json:"metadata,omitempty"`
	ExpiredAt     *time.Time                   `json:"expired_at,omitempty"`
}

// ListSessionsRequest 查询会话列表请求
type ListSessionsRequest struct {
	Type          *string `form:"type,omitempty"`
	Status        *string `form:"status,omitempty"`
	UserID        *uint   `form:"user_id,omitempty"`
	AssistantID   *uint   `form:"assistant_id,omitempty"`
	AssistantCode *string `form:"assistant_code,omitempty"`
	Keyword       *string `form:"keyword,omitempty"`
	Page          int     `form:"page,default=1"`
	PerPage       int     `form:"per_page,default=20"`
}

// AddMessageRequest 添加消息请求
type AddMessageRequest struct {
	Role        string                     `json:"role" binding:"required"`
	Content     string                     `json:"content" binding:"required"`
	MessageType string                     `json:"message_type,omitempty"`
	Metadata    *types.MessageMetadata     `json:"metadata,omitempty"`
}

// Session 会话响应结构
type Session struct {
	ID            uint                      `json:"id"`
	SessionID     string                    `json:"session_id"`
	Type          string                    `json:"type"`
	UserID        uint                      `json:"user_id"`
	AssistantID   uint                      `json:"assistant_id"`
	AssistantCode string                    `json:"assistant_code"`
	Status        string                    `json:"status"`
	Title         string                    `json:"title"`
	Metadata      *types.SessionMetadata    `json:"metadata,omitempty"`
	MessageCount  int                       `json:"message_count"`
	LastMessageAt *time.Time                `json:"last_message_at,omitempty"`
	ExpiredAt     *time.Time                `json:"expired_at,omitempty"`
	CreatedAt     time.Time                 `json:"created_at"`
	UpdatedAt     time.Time                 `json:"updated_at"`
}

// SessionMessage 消息响应结构
type SessionMessage struct {
	ID          uint                      `json:"id"`
	SessionID   string                    `json:"session_id"`
	Role        string                    `json:"role"`
	Content     string                    `json:"content"`
	MessageType string                    `json:"message_type"`
	Metadata    *types.MessageMetadata    `json:"metadata,omitempty"`
	Sequence    int64                     `json:"sequence"`
	CreatedAt   time.Time                 `json:"created_at"`
}

// SessionList 会话列表响应
type SessionList struct {
	Total int64     `json:"total"`
	Page  int       `json:"page"`
	Items []Session `json:"items"`
}

// MessageList 消息列表响应
type MessageList struct {
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Items []SessionMessage `json:"items"`
}
