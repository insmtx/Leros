package engines

import (
	"fmt"
	"sync"
)

// Registry 存储按名称启用的外部 CLI 引擎。
type Registry struct {
	mu      sync.RWMutex
	engines map[string]Engine
}

// NewRegistry 创建一个空的引擎注册表。
func NewRegistry() *Registry {
	return &Registry{engines: make(map[string]Engine)}
}

// Register 添加或替换引擎。
func (r *Registry) Register(name string, engine Engine) error {
	if name == "" {
		return fmt.Errorf("engine name is required")
	}
	if engine == nil {
		return fmt.Errorf("engine %q is nil", name)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.engines[name] = engine
	return nil
}

// Get 按名称获取引擎。
func (r *Registry) Get(name string) (Engine, bool) {
	if r == nil {
		return nil, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	engine, ok := r.engines[name]
	return engine, ok
}

// Names 返回已启用引擎的名称列表。
func (r *Registry) Names() []string {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.engines))
	for name := range r.engines {
		names = append(names, name)
	}
	return names
}
