package runtime

import (
	"context"
	"fmt"

	skillstore "github.com/insmtx/SingerOS/backend/internal/skill/store"
)

// Manager 执行 Skill 变更，并调度非阻塞的运行时后处理。
type Manager struct {
	store *skillstore.SkillStore
	post  *PostProcessor
}

// NewManager 基于 SkillStore 创建 Manager。
func NewManager(store *skillstore.SkillStore, post *PostProcessor) (*Manager, error) {
	if store == nil {
		return nil, fmt.Errorf("skill store is required")
	}
	return &Manager{store: store, post: post}, nil
}

// RootDir 返回当前管理的 Skill 根目录。
func (m *Manager) RootDir() string {
	if m == nil || m.store == nil {
		return ""
	}
	return m.store.RootDir()
}

// Create 写入新 Skill，并在成功后调度后处理。
func (m *Manager) Create(ctx context.Context, req skillstore.CreateRequest) (*skillstore.Result, error) {
	if err := m.validate(); err != nil {
		return nil, err
	}
	result, err := m.store.Create(ctx, req)
	m.after(result, err)
	return result, err
}

// Patch 替换 Skill 文件中的文本，并在成功后调度后处理。
func (m *Manager) Patch(ctx context.Context, req skillstore.PatchRequest) (*skillstore.Result, error) {
	if err := m.validate(); err != nil {
		return nil, err
	}
	result, err := m.store.Patch(ctx, req)
	m.after(result, err)
	return result, err
}

// WriteFile 写入 Skill supporting file，并在成功后调度后处理。
func (m *Manager) WriteFile(ctx context.Context, req skillstore.WriteFileRequest) (*skillstore.Result, error) {
	if err := m.validate(); err != nil {
		return nil, err
	}
	result, err := m.store.WriteFile(ctx, req)
	m.after(result, err)
	return result, err
}

// RemoveFile 删除 Skill supporting file，并在成功后调度后处理。
func (m *Manager) RemoveFile(ctx context.Context, req skillstore.RemoveFileRequest) (*skillstore.Result, error) {
	if err := m.validate(); err != nil {
		return nil, err
	}
	result, err := m.store.RemoveFile(ctx, req)
	m.after(result, err)
	return result, err
}

func (m *Manager) validate() error {
	if m == nil || m.store == nil {
		return fmt.Errorf("skill manager is not initialized")
	}
	return nil
}

func (m *Manager) after(result *skillstore.Result, err error) {
	if err != nil || m == nil || m.post == nil {
		return
	}
	m.post.AfterMutation(result)
}
