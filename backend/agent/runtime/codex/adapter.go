// Package codex 将 Codex CLI 适配到 Leros 外部 CLI 引擎接口。
// 使用 codex app-server --listen stdio:// 模式进行通信。
package codex

import (
	"context"
	"os"
	"path/filepath"

	engines "github.com/insmtx/Leros/backend/agent/runtime/provider"
)

// Adapter 通过 Codex CLI app-server 模式执行提示。
type Adapter struct {
	invoker *AppServerInvoker
}

// NewAdapter 创建 Codex CLI 引擎适配器（app-server 模式）。
func NewAdapter(binary string, extraEnv map[string]string) *Adapter {
	if binary == "" {
		binary = "codex"
	}
	return &Adapter{invoker: NewAppServerInvoker(binary, extraEnv)}
}

// Prepare 执行 Codex 工作区设置（当前为空实现）。
func (a *Adapter) Prepare(_ context.Context, _ engines.PrepareRequest) error {
	return nil
}

// Run 启动 Codex CLI 并返回进程句柄。
func (a *Adapter) Run(ctx context.Context, req engines.RunRequest) (*engines.RunHandle, error) {
	handle, err := a.invoker.Run(ctx, req)
	if err != nil {
		return nil, err
	}
	return handle, nil
}

// GetSkillDir returns the skill directory path for Codex CLI.
func (a *Adapter) GetSkillDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".agents", "skills")
}

var _ engines.Engine = (*Adapter)(nil)
