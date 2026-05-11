package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/insmtx/Leros/backend/internal/agent/eventtypes"
	"github.com/insmtx/Leros/backend/internal/api/auth"
	"github.com/insmtx/Leros/backend/internal/api/contract"
	"github.com/insmtx/Leros/backend/internal/api/dto"
	"github.com/insmtx/Leros/backend/internal/infra/db"
	eventbus "github.com/insmtx/Leros/backend/internal/infra/mq"
	"github.com/insmtx/Leros/backend/pkg/dm"
	"github.com/insmtx/Leros/backend/runtime/events"
	"github.com/insmtx/Leros/backend/types"
	"github.com/ygpkg/yg-go/encryptor/snowflake"
	"github.com/ygpkg/yg-go/logs"
)

var _ contract.SessionService = (*sessionService)(nil)

type sessionService struct {
	db         *gorm.DB
	subscriber eventbus.Subscriber
	publisher  eventbus.Publisher
	inferrer   AssistantInferrer
}

func NewSessionService(db *gorm.DB, subscriber eventbus.Subscriber, publisher eventbus.Publisher, inferrer AssistantInferrer) contract.SessionService {
	return &sessionService{
		db:         db,
		subscriber: subscriber,
		publisher:  publisher,
		inferrer:   inferrer,
	}
}

func (s *sessionService) CreateSession(ctx context.Context, req *contract.CreateSessionRequest) (*contract.Session, error) {
	if req.Type == "" {
		return nil, errors.New("type is required")
	}

	caller, _ := auth.FromContext(ctx)
	if caller == nil || caller.Uin == 0 || caller.OrgID == 0 {
		return nil, errors.New("user not authenticated or org not set")
	}

	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = fmt.Sprintf("sess_%s", snowflake.GenerateIDBase58())
	}

	exists, err := db.SessionIDExists(ctx, s.db, sessionID, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("session with this session_id already exists")
	}

	session := &types.Session{
		SessionID:            sessionID,
		Type:                 req.Type,
		Uin:                  caller.Uin,
		OrgID:                caller.OrgID,
		AssistantID:          req.AssistantID,
		AllocatedAssistantID: req.AssistantID,
		Status:               string(types.SessionStatusActive),
		Title:                req.Title,
		MessageCount:         0,
		ExpiredAt:            req.ExpiredAt,
	}

	if req.Metadata != nil {
		session.Metadata = *req.Metadata
	}

	if err := db.CreateSession(ctx, s.db, session); err != nil {
		return nil, err
	}

	return convertToContractSession(session), nil
}

func (s *sessionService) GetSession(ctx context.Context, id uint, sessionID string) (*contract.Session, error) {
	var session *types.Session
	var err error

	if id > 0 {
		session, err = db.GetSessionByID(ctx, s.db, id)
	} else if sessionID != "" {
		session, err = db.GetSessionBySessionID(ctx, s.db, sessionID)
	} else {
		return nil, errors.New("id or session_id is required")
	}

	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, errors.New("session not found")
	}

	return convertToContractSession(session), nil
}

func (s *sessionService) UpdateSession(ctx context.Context, id uint, req *contract.UpdateSessionRequest) (*contract.Session, error) {
	session, err := db.GetSessionByID(ctx, s.db, id)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, errors.New("session not found")
	}

	if req.Title != "" {
		session.Title = req.Title
	}
	if req.Metadata != nil {
		session.Metadata = *req.Metadata
	}
	if req.ExpiredAt != nil {
		session.ExpiredAt = req.ExpiredAt
	}

	session.UpdatedAt = time.Now()

	if err := db.UpdateSession(ctx, s.db, session); err != nil {
		return nil, err
	}

	return convertToContractSession(session), nil
}

func (s *sessionService) DeleteSession(ctx context.Context, id uint) error {
	session, err := db.GetSessionByID(ctx, s.db, id)
	if err != nil {
		return err
	}
	if session == nil {
		return errors.New("session not found")
	}

	return db.DeleteSession(ctx, s.db, id)
}

func (s *sessionService) ListSessions(ctx context.Context, req *contract.ListSessionsRequest) (*contract.SessionList, error) {
	caller, _ := auth.FromContext(ctx)

	var uin *uint
	var orgID *uint
	if caller != nil && caller.Uin > 0 {
		uin = &caller.Uin
		orgID = &caller.OrgID
	}

	sessions, total, err := db.ListSessions(
		ctx,
		s.db,
		req.Type,
		req.Status,
		uin,
		orgID,
		req.AssistantID,
		req.AssistantCode,
		req.Keyword,
		req.Page,
		req.PerPage,
	)
	if err != nil {
		return nil, err
	}

	items := make([]contract.Session, 0, len(sessions))
	for _, session := range sessions {
		items = append(items, *convertToContractSession(session))
	}

	return &contract.SessionList{
		Total: total,
		Page:  req.Page,
		Items: items,
	}, nil
}

func (s *sessionService) ActivateSession(ctx context.Context, id uint) error {
	session, err := db.GetSessionByID(ctx, s.db, id)
	if err != nil {
		return err
	}
	if session == nil {
		return errors.New("session not found")
	}

	if session.Status == string(types.SessionStatusEnded) {
		return errors.New("cannot activate from ended state")
	}

	return db.ActivateSession(ctx, s.db, id)
}

func (s *sessionService) PauseSession(ctx context.Context, id uint) error {
	session, err := db.GetSessionByID(ctx, s.db, id)
	if err != nil {
		return err
	}
	if session == nil {
		return errors.New("session not found")
	}

	if session.Status == string(types.SessionStatusEnded) || session.Status == string(types.SessionStatusExpired) {
		return fmt.Errorf("cannot pause from %s state", session.Status)
	}

	return db.PauseSession(ctx, s.db, id)
}

func (s *sessionService) EndSession(ctx context.Context, id uint) error {
	session, err := db.GetSessionByID(ctx, s.db, id)
	if err != nil {
		return err
	}
	if session == nil {
		return errors.New("session not found")
	}

	if session.Status == string(types.SessionStatusEnded) {
		return errors.New("session already ended")
	}

	return db.EndSession(ctx, s.db, id)
}

func (s *sessionService) ResumeSession(ctx context.Context, id uint) error {
	session, err := db.GetSessionByID(ctx, s.db, id)
	if err != nil {
		return err
	}
	if session == nil {
		return errors.New("session not found")
	}

	if session.Status != string(types.SessionStatusPaused) {
		return errors.New("can only resume from paused state")
	}

	return db.ResumeSession(ctx, s.db, id)
}

func (s *sessionService) AddMessage(ctx context.Context, sessionID uint, req *contract.AddMessageRequest) (*contract.SessionMessage, error) {
	if req.Role == "" {
		return nil, errors.New("role is required")
	}
	if req.Content == "" {
		return nil, errors.New("content is required")
	}

	session, err := db.GetSessionByID(ctx, s.db, sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, errors.New("session not found")
	}

	sequence, err := db.GetNextSequence(ctx, s.db, session.SessionID)
	if err != nil {
		return nil, err
	}

	message := &types.SessionMessage{
		SessionID:   session.SessionID,
		Role:        req.Role,
		Content:     req.Content,
		MessageType: req.MessageType,
		Status:      req.Status,
		Sequence:    sequence,
		Timestamp:   time.Now().UnixMilli(), // Unix 毫秒时间戳
	}

	if req.Chunks != nil && len(req.Chunks) > 0 {
		message.Chunks = req.Chunks
	}

	if req.Thinking != "" {
		message.Thinking = req.Thinking
	}

	if req.ToolCalls != nil && len(req.ToolCalls) > 0 {
		message.ToolCalls = req.ToolCalls
	}

	if req.Metadata != nil {
		message.Metadata = *req.Metadata
	} else {
		message.Metadata = types.MessageMetadata{}
	}

	if message.MessageType == "" {
		message.MessageType = string(types.MessageTypeText)
	}

	if message.Status == "" {
		message.Status = string(types.MessageStatusComplete)
	}

	if err := db.CreateMessage(ctx, s.db, message); err != nil {
		return nil, err
	}

	now := time.Now()
	if err := db.IncrementMessageCount(ctx, s.db, sessionID); err != nil {
		return nil, err
	}
	if err := db.UpdateLastMessageAt(ctx, s.db, sessionID, now); err != nil {
		return nil, err
	}

	caller, _ := auth.FromContext(ctx)
	orgID := session.OrgID
	if orgID == 0 && caller != nil {
		orgID = caller.OrgID
	}

	if session.AssistantID == 0 && session.AllocatedAssistantID == 0 && s.inferrer != nil {
		assignedAssistantID := s.inferrer.InferAssignedAssistantID(ctx, orgID, session.Type)
		if assignedAssistantID > 0 {
			session.AllocatedAssistantID = assignedAssistantID
			session.UpdatedAt = time.Now()
			if err := db.UpdateSession(ctx, s.db, session); err != nil {
				return nil, fmt.Errorf("failed to update allocated_assistant_id: %w", err)
			}
		}
	}

	topic, err := dm.WorkerTaskTopic(orgID, session.AllocatedAssistantID)
	if err != nil {
		return nil, fmt.Errorf("failed to construct worker task topic: %w", err)
	}

	messagePayload := map[string]interface{}{
		"session_id":   session.SessionID,
		"role":         message.Role,
		"content":      message.Content,
		"message_type": message.MessageType,
		"sequence":     message.Sequence,
		"timestamp":    message.Timestamp,
	}

	if err := s.publisher.Publish(ctx, topic, messagePayload); err != nil {
		logs.ErrorContextf(ctx, "Failed to publish message to assistant %d: %v", session.AllocatedAssistantID, err)
		return nil, fmt.Errorf("failed to publish message to assistant: %w", err)
	}
	logs.DebugContextf(ctx, "Published message to topic %s: %v", topic, messagePayload)

	return convertToContractSessionMessage(message), nil
}

func (s *sessionService) GetSessionMessages(ctx context.Context, sessionID uint, page, perPage int) (*contract.MessageList, error) {
	session, err := db.GetSessionByID(ctx, s.db, sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, errors.New("session not found")
	}

	messages, total, err := db.GetSessionMessages(ctx, s.db, session.SessionID, page, perPage)
	if err != nil {
		return nil, err
	}

	items := make([]contract.SessionMessage, 0, len(messages))
	for _, message := range messages {
		items = append(items, *convertToContractSessionMessage(message))
	}

	return &contract.MessageList{
		Total: total,
		Page:  page,
		Items: items,
	}, nil
}

func (s *sessionService) DeleteMessage(ctx context.Context, messageID uint) error {
	message, err := db.GetMessageByID(ctx, s.db, messageID)
	if err != nil {
		return err
	}
	if message == nil {
		return errors.New("message not found")
	}

	if err := db.DeleteMessage(ctx, s.db, messageID); err != nil {
		return err
	}

	return nil
}

func (s *sessionService) ClearSessionMessages(ctx context.Context, sessionID uint) error {
	session, err := db.GetSessionByID(ctx, s.db, sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return errors.New("session not found")
	}

	if err := db.ClearSessionMessages(ctx, s.db, session.SessionID); err != nil {
		return err
	}

	session.MessageCount = 0
	session.LastMessageAt = nil
	session.UpdatedAt = time.Now()

	return db.UpdateSession(ctx, s.db, session)
}

func toJSONString(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func (s *sessionService) StreamSessionEvents(ctx context.Context, sessionID string, lastSequence int64, sink events.Sink) error {
	caller, _ := auth.FromContext(ctx)
	if caller == nil || caller.OrgID == 0 {
		return errors.New("user not authenticated or org not set")
	}

	topic, err := dm.SessionResultStreamTopic(caller.OrgID, sessionID)
	if err != nil {
		return fmt.Errorf("failed to construct session result stream topic: %w", err)
	}
	return s.subscriber.Subscribe(ctx, topic, func(event any) {
		switch msg := event.(type) {
		case eventtypes.MessageStreamMessage:
			if msg.Body.Seq <= lastSequence {
				return
			}

			se := dto.SessionEvent{
				SessionID: msg.Route.SessionID,
				Sequence:  msg.Body.Seq,
				Timestamp: msg.CreatedAt.UnixMilli(),
			}

			switch msg.Body.Event {
			case eventtypes.StreamEventMessageDelta:
				se.Type = dto.SessionEventTypeMessageDelta
				se.Payload = dto.MessageDeltaPayload{
					Role:    string(msg.Body.Payload.Role),
					Content: msg.Body.Payload.Content,
				}
			case eventtypes.StreamEventToolCallStarted:
				se.Type = dto.SessionEventTypeToolCallStarted
				if tc := msg.Body.Payload.ToolCall; tc != nil {
					se.Payload = dto.ToolCallDeltaPayload{
						ID:   tc.ID,
						Name: tc.Name,
					}
				}
			case eventtypes.StreamEventRunStarted:
				se.Type = dto.SessionEventTypeRunStarted
			case eventtypes.StreamEventRunCompleted:
				se.Type = dto.SessionEventTypeRunCompleted
			case eventtypes.StreamEventRunFailed:
				se.Type = dto.SessionEventTypeRunFailed
			default:
				logs.Warnf("unknown stream event type: %v", msg.Body.Event)
				return
			}
			_ = sink.Emit(ctx, &events.Event{
				Type:    events.EventType(se.Type),
				Content: toJSONString(se),
			})
		}
	})
}

func convertToContractSession(session *types.Session) *contract.Session {
	result := &contract.Session{
		ID:                   session.ID,
		SessionID:            session.SessionID,
		Type:                 session.Type,
		Uin:                  session.Uin,
		OrgID:                session.OrgID,
		AssistantID:          session.AssistantID,
		AllocatedAssistantID: session.AllocatedAssistantID,
		Status:               session.Status,
		Title:                session.Title,
		MessageCount:         session.MessageCount,
		CreatedAt:            session.CreatedAt,
		UpdatedAt:            session.UpdatedAt,
	}

	if session.Metadata.Tags != nil || session.Metadata.Extra != nil || session.Metadata.UserAgent != "" || session.Metadata.IPAddress != "" {
		result.Metadata = &session.Metadata
	}
	if session.LastMessageAt != nil {
		result.LastMessageAt = session.LastMessageAt
	}
	if session.ExpiredAt != nil {
		result.ExpiredAt = session.ExpiredAt
	}

	return result
}

func convertToContractSessionMessage(message *types.SessionMessage) *contract.SessionMessage {
	result := &contract.SessionMessage{
		ID:          fmt.Sprintf("%d", message.ID), // 转换为 string 以匹配前端
		SessionID:   message.SessionID,
		Role:        message.Role,
		Content:     message.Content,
		MessageType: message.MessageType,
		Status:      message.Status,
		Timestamp:   message.Timestamp,
		Sequence:    message.Sequence,
		CreatedAt:   message.CreatedAt,
	}

	if message.Chunks != nil && len(message.Chunks) > 0 {
		result.Chunks = message.Chunks
	}

	if message.Thinking != "" {
		result.Thinking = message.Thinking
	}

	if message.ToolCalls != nil && len(message.ToolCalls) > 0 {
		result.ToolCalls = message.ToolCalls
	}

	if message.Metadata.ImageURL != "" || message.Metadata.Language != "" || message.Metadata.FileURL != "" || message.Metadata.FileName != "" || message.Metadata.Model != "" || message.Metadata.Extra != nil {
		result.Metadata = &message.Metadata
	}

	return result
}
