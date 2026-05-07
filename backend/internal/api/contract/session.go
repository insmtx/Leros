package contract

import "context"

// SessionService 定义会话服务接口
type SessionService interface {
	// Session CRUD
	CreateSession(ctx context.Context, req *CreateSessionRequest) (*Session, error)
	GetSession(ctx context.Context, id uint, sessionID string) (*Session, error)
	UpdateSession(ctx context.Context, id uint, req *UpdateSessionRequest) (*Session, error)
	DeleteSession(ctx context.Context, id uint) error
	ListSessions(ctx context.Context, req *ListSessionsRequest) (*SessionList, error)

	// Lifecycle management
	ActivateSession(ctx context.Context, id uint) error
	PauseSession(ctx context.Context, id uint) error
	EndSession(ctx context.Context, id uint) error
	ResumeSession(ctx context.Context, id uint) error

	// Message management
	AddMessage(ctx context.Context, sessionID uint, req *AddMessageRequest) (*SessionMessage, error)
	GetSessionMessages(ctx context.Context, sessionID uint, page, perPage int) (*MessageList, error)
	DeleteMessage(ctx context.Context, messageID uint) error
	ClearSessionMessages(ctx context.Context, sessionID uint) error
}
