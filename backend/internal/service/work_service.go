package service

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/insmtx/Leros/backend/internal/api/auth"
	"github.com/insmtx/Leros/backend/internal/api/contract"
	eventbus "github.com/insmtx/Leros/backend/internal/infra/mq"
)

var _ contract.WorkService = (*workService)(nil)

type workService struct {
	db       *gorm.DB
	eventbus eventbus.EventBus
	inferrer AssistantInferrer
}

func NewWorkService(database *gorm.DB, eventbus eventbus.EventBus, inferrer AssistantInferrer) contract.WorkService {
	return &workService{
		db:       database,
		eventbus: eventbus,
		inferrer: inferrer,
	}
}

func (s *workService) NewMessage(ctx context.Context, req *contract.NewMessageRequest) (*contract.NewMessageResponse, error) {
	if req.Content == "" {
		return nil, errors.New("content is required")
	}

	caller, _ := auth.FromContext(ctx)
	if caller == nil || caller.Uin == 0 || caller.OrgID == 0 {
		return nil, errors.New("user not authenticated or org not set")
	}

	p := NewMessagePoster(s.db, s.eventbus, s.inferrer)
	return p.RunNewMessage(ctx, req, caller)
}
