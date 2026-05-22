package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/insmtx/Leros/backend/internal/api/contract"
	"github.com/insmtx/Leros/backend/internal/infra/db"
	"github.com/insmtx/Leros/backend/pkg/dm"
	"github.com/insmtx/Leros/backend/types"
	"github.com/ygpkg/yg-go/encryptor/snowflake"
	"github.com/ygpkg/yg-go/logs"
)

type workMessageWrapper struct {
	s      *workService
	ctx    context.Context
	req    *contract.NewMessageRequest
	caller *types.Caller

	project     *types.Project
	task        *types.Task
	taskSession *types.Session
	message     *types.SessionMessage
}

func (w *workMessageWrapper) resolveOrCreateProject() error {
	if w.req.ProjectID != "" {
		p, err := db.GetProjectByPublicID(w.ctx, w.s.db, w.caller.OrgID, w.req.ProjectID)
		if err != nil {
			return err
		}
		if p == nil {
			return errors.New("project not found")
		}
		w.project = p
		return nil
	}

	runes := []rune(w.req.Content)
	title := string(runes)
	if len(runes) > 50 {
		title = string(runes[:50])
	}

	projectID := fmt.Sprintf("prj_%s", snowflake.GenerateIDBase58())
	w.project = &types.Project{
		PublicID:    projectID,
		OrgID:       w.caller.OrgID,
		OwnerID:     w.caller.Uin,
		Name:        title,
		Description: "",
		Status:      string(types.ProjectStatusActive),
	}
	if err := db.CreateProject(w.ctx, w.s.db, w.project); err != nil {
		return fmt.Errorf("create project: %w", err)
	}

	if err := db.CreateProjectMember(w.ctx, w.s.db, &types.ProjectMember{
		ProjectID:  w.project.ID,
		MemberID:   w.caller.Uin,
		MemberType: types.MemberTypeUser,
		MemberRole: types.MemberRoleOwner,
	}); err != nil {
		logs.WarnContextf(w.ctx, "create project member failed: %v", err)
	}

	return nil
}

func (w *workMessageWrapper) ensureProjectSession() error {
	projectSession, err := db.GetProjectSession(w.ctx, w.s.db, w.project.ID)
	if err != nil {
		return fmt.Errorf("get project session: %w", err)
	}
	if projectSession != nil {
		return nil
	}

	projectSessionID := fmt.Sprintf("sess_%s", snowflake.GenerateIDBase58())
	projectSession = &types.Session{
		PublicID:             projectSessionID,
		Type:                 types.SessionTypeProject,
		Uin:                  w.caller.Uin,
		OrgID:                w.caller.OrgID,
		AssistantID:          w.req.AssistantID,
		AllocatedAssistantID: w.req.AssistantID,
		ProjectID:            &w.project.ID,
		Status:               string(types.SessionStatusActive),
		Title:                "项目协作",
	}
	if err := db.CreateSession(w.ctx, w.s.db, projectSession); err != nil {
		return fmt.Errorf("create project session: %w", err)
	}
	return nil
}

func (w *workMessageWrapper) resolveOrCreateTask() error {
	if w.req.TaskID != "" {
		t, err := db.GetTaskByPublicID(w.ctx, w.s.db, w.req.TaskID)
		if err != nil {
			return err
		}
		if t == nil {
			return errors.New("task not found")
		}
		w.task = t
		return nil
	}

	runes := []rune(w.req.Content)
	taskTitle := string(runes)
	if len(runes) > 50 {
		taskTitle = string(runes[:50])
	}

	taskID := fmt.Sprintf("task_%s", snowflake.GenerateIDBase58())
	w.task = &types.Task{
		PublicID:    taskID,
		OrgID:       w.caller.OrgID,
		OwnerID:     w.caller.Uin,
		ProjectID:   w.project.ID,
		TaskType:    types.TaskTypeGeneral,
		Title:       taskTitle,
		Description: w.req.Content,
		Status:      string(types.TaskStatusCreated),
	}
	if err := db.CreateTask(w.ctx, w.s.db, w.task); err != nil {
		return fmt.Errorf("create task: %w", err)
	}

	return nil
}

func (w *workMessageWrapper) createTaskSession() error {
	taskSessionID := fmt.Sprintf("sess_%s", snowflake.GenerateIDBase58())
	w.taskSession = &types.Session{
		PublicID:             taskSessionID,
		Type:                 types.SessionTypeTask,
		Uin:                  w.caller.Uin,
		OrgID:                w.caller.OrgID,
		AssistantID:          w.req.AssistantID,
		AllocatedAssistantID: w.req.AssistantID,
		ProjectID:            &w.project.ID,
		TaskID:               &w.task.ID,
		Status:               string(types.SessionStatusActive),
		Title:                w.task.Title,
	}
	if err := db.CreateSession(w.ctx, w.s.db, w.taskSession); err != nil {
		return fmt.Errorf("create task session: %w", err)
	}

	w.task.SessionID = &w.taskSession.ID
	if err := w.s.db.WithContext(w.ctx).Model(w.task).Update("session_id", w.taskSession.ID).Error; err != nil {
		logs.WarnContextf(w.ctx, "update task session_id failed: %v", err)
	}

	return nil
}

func (w *workMessageWrapper) createMessage() error {
	sequence, err := db.GetNextSequence(w.ctx, w.s.db, w.taskSession.ID)
	if err != nil {
		return err
	}

	msgType := w.req.MessageType
	if msgType == "" {
		msgType = string(types.MessageTypeText)
	}

	w.message = &types.SessionMessage{
		SessionID:   w.taskSession.ID,
		Role:        string(types.MessageRoleUser),
		Content:     w.req.Content,
		MessageType: msgType,
		Status:      string(types.MessageStatusPending),
		Sequence:    sequence,
		Timestamp:   time.Now().UnixMilli(),
	}
	if err := db.CreateMessage(w.ctx, w.s.db, w.message); err != nil {
		return fmt.Errorf("create message: %w", err)
	}

	return nil
}

func (w *workMessageWrapper) publishMessageEvents() error {
	now := time.Now()
	if err := db.IncrementMessageCount(w.ctx, w.s.db, w.taskSession.ID); err != nil {
		return err
	}
	if err := db.UpdateLastMessageAt(w.ctx, w.s.db, w.taskSession.ID, now); err != nil {
		return err
	}

	if w.taskSession.OrgID > 0 {
		topic, err := dm.SessionMessageRequestSubject(w.taskSession.OrgID, w.taskSession.PublicID)
		if err != nil {
			logs.WarnContextf(w.ctx, "failed to build message request subject: %v", err)
		} else {
			if err := w.s.eventbus.Publish(w.ctx, topic, w.message); err != nil {
				logs.WarnContextf(w.ctx, "failed to publish message to eventbus: %v", err)
			}
		}
	}

	return nil
}

func (w *workMessageWrapper) publishWorkerTask() error {
	return w.s.publishWorkerTask(w.ctx, w.taskSession, w.message)
}
