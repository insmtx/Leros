package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/insmtx/SingerOS/backend/internal/api/contract"
	"github.com/insmtx/SingerOS/backend/internal/infra/db"
	"github.com/insmtx/SingerOS/backend/types"
)

var _ contract.SessionService = (*sessionService)(nil)

type sessionService struct {
	db *gorm.DB
}

func NewSessionService(db *gorm.DB) contract.SessionService {
	return &sessionService{
		db: db,
	}
}

func (s *sessionService) CreateSession(ctx context.Context, req *contract.CreateSessionRequest) (*contract.Session, error) {
	if req.Type == "" {
		return nil, errors.New("type is required")
	}

	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = fmt.Sprintf("sess_%d", time.Now().UnixNano())
	}

	exists, err := db.SessionIDExists(ctx, s.db, sessionID, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("session with this session_id already exists")
	}

	session := &types.Session{
		SessionID:     sessionID,
		Type:          req.Type,
		UserID:        req.UserID,
		AssistantID:   req.AssistantID,
		AssistantCode: req.AssistantCode,
		Status:        string(types.SessionStatusActive),
		Title:         req.Title,
		MessageCount:  0,
		ExpiredAt:     req.ExpiredAt,
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
	sessions, total, err := db.ListSessions(
		ctx,
		s.db,
		req.Type,
		req.Status,
		req.UserID,
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
		Sequence:    sequence,
	}

	if req.Metadata != nil {
		message.Metadata = *req.Metadata
	} else {
		message.Metadata = types.MessageMetadata{}
	}

	if message.MessageType == "" {
		message.MessageType = string(types.MessageTypeText)
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

	session, err := db.GetSessionBySessionID(ctx, s.db, message.SessionID)
	if err != nil {
		return err
	}

	if err := db.DeleteMessage(ctx, s.db, messageID); err != nil {
		return err
	}

	if session != nil {
		if err := db.IncrementMessageCount(ctx, s.db, session.ID); err != nil {
			return err
		}
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

func convertToContractSession(session *types.Session) *contract.Session {
	result := &contract.Session{
		ID:            session.ID,
		SessionID:     session.SessionID,
		Type:          session.Type,
		UserID:        session.UserID,
		AssistantID:   session.AssistantID,
		AssistantCode: session.AssistantCode,
		Status:        session.Status,
		Title:         session.Title,
		MessageCount:  session.MessageCount,
		CreatedAt:     session.CreatedAt,
		UpdatedAt:     session.UpdatedAt,
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
		ID:          message.ID,
		SessionID:   message.SessionID,
		Role:        message.Role,
		Content:     message.Content,
		MessageType: message.MessageType,
		Sequence:    message.Sequence,
		CreatedAt:   message.CreatedAt,
	}

	if message.Metadata.ImageURL != "" || message.Metadata.Language != "" || message.Metadata.FileURL != "" || message.Metadata.FileName != "" || message.Metadata.Extra != nil {
		result.Metadata = &message.Metadata
	}

	return result
}
