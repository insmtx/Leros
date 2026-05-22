package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/insmtx/Leros/backend/internal/api/auth"
	"github.com/insmtx/Leros/backend/internal/api/contract"
	"github.com/insmtx/Leros/backend/internal/infra/db"
	eventbus "github.com/insmtx/Leros/backend/internal/infra/mq"
	"github.com/insmtx/Leros/backend/internal/worker/protocol"
	"github.com/insmtx/Leros/backend/pkg/dm"
	"github.com/insmtx/Leros/backend/types"
	"github.com/ygpkg/yg-go/logs"
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

	w := &workMessageWrapper{s: s, ctx: ctx, req: req, caller: caller}

	if err := w.resolveOrCreateProject(); err != nil {
		return nil, err
	}
	if err := w.ensureProjectSession(); err != nil {
		return nil, err
	}
	if err := w.resolveOrCreateTask(); err != nil {
		return nil, err
	}
	if err := w.createTaskSession(); err != nil {
		return nil, err
	}
	if err := w.createMessage(); err != nil {
		return nil, err
	}
	if err := w.publishMessageEvents(); err != nil {
		return nil, err
	}
	if err := w.publishWorkerTask(); err != nil {
		return nil, err
	}

	return &contract.NewMessageResponse{
		ProjectID:   w.project.PublicID,
		TaskID:      w.task.PublicID,
		SessionID:   w.taskSession.PublicID,
		MessageID:   fmt.Sprintf("%d", w.message.ID),
		AssistantID: w.taskSession.AllocatedAssistantID,
	}, nil
}

func (s *workService) publishWorkerTask(ctx context.Context, session *types.Session, message *types.SessionMessage) error {
	caller, _ := auth.FromContext(ctx)
	orgID := session.OrgID
	if orgID == 0 && caller != nil {
		orgID = caller.OrgID
	}

	if session.AssistantID == 0 && session.AllocatedAssistantID == 0 && s.inferrer != nil {
		assignedAssistantID := s.inferrer.InferAssignedAssistantID(ctx, orgID, string(session.Type))
		if assignedAssistantID > 0 {
			session.AllocatedAssistantID = assignedAssistantID
			if err := db.UpdateAllocatedAssistantID(ctx, s.db, session.ID, assignedAssistantID); err != nil {
				return fmt.Errorf("failed to update allocated_assistant_id: %w", err)
			}
		}
	}

	if session.AllocatedAssistantID == 0 {
		logs.DebugContextf(ctx, "Skipping task publish: no worker allocated for session %s", session.PublicID)
		return nil
	}

	topic, err := dm.WorkerTaskSubject(orgID, session.AllocatedAssistantID)
	if err != nil {
		return fmt.Errorf("failed to construct worker task topic: %w", err)
	}

	messagePayload := protocol.WorkerTaskMessage{
		ID:        fmt.Sprintf("msg_%d_%d", session.ID, message.Sequence),
		Type:      protocol.MessageTypeWorkerTask,
		CreatedAt: time.Now().UTC(),
		Trace: protocol.TraceContext{
			TraceID:   session.PublicID,
			RequestID: fmt.Sprintf("req_%d", message.ID),
			TaskID:    fmt.Sprintf("task_%d", message.ID),
		},
		Route: protocol.RouteContext{
			OrgID:     orgID,
			SessionID: session.PublicID,
			WorkerID:  session.AllocatedAssistantID,
		},
		Body: protocol.WorkerTaskBody{
			TaskType: protocol.TaskTypeAgentRun,
			Actor: protocol.ActorContext{
				UserID:      fmt.Sprintf("%d", session.Uin),
				DisplayName: "",
				Channel:     "session",
			},
			Input: protocol.TaskInput{
				Type: protocol.InputTypeMessage,
			},
		},
		Metadata: map[string]any{
			"session_id":   session.PublicID,
			"message_type": message.MessageType,
			"sequence":     message.Sequence,
			"timestamp":    message.Timestamp,
		},
	}

	if err := s.eventbus.Publish(ctx, topic, messagePayload); err != nil {
		logs.ErrorContextf(ctx, "Failed to publish message to assistant %d: %v", session.AllocatedAssistantID, err)
		return fmt.Errorf("failed to publish message to assistant: %w", err)
	}
	logs.DebugContextf(ctx, "Published message to topic %s: session_id=%s sequence=%d", topic, session.PublicID, message.Sequence)
	return nil
}
