package runtime

import (
	"context"
	"sync"
	"time"

	skillcatalog "github.com/insmtx/SingerOS/backend/internal/skill/catalog"
	skillstore "github.com/insmtx/SingerOS/backend/internal/skill/store"
	"github.com/insmtx/SingerOS/backend/runtime/engines"
	"github.com/ygpkg/yg-go/logs"
)

const defaultPostProcessDelay = 500 * time.Millisecond

// PostProcessor 在 Skill 变更后执行非阻塞后处理。
type PostProcessor struct {
	sourceDir string
	catalog   skillcatalog.CatalogReloader
	delay     time.Duration

	mu    sync.Mutex
	timer *time.Timer
}

// NewPostProcessor 创建 Skill 变更后处理器。
func NewPostProcessor(sourceDir string, catalog skillcatalog.CatalogReloader) *PostProcessor {
	return &PostProcessor{
		sourceDir: sourceDir,
		catalog:   catalog,
		delay:     defaultPostProcessDelay,
	}
}

// AfterMutation 调度异步 CLI 同步和 Catalog reload 工作。
func (p *PostProcessor) AfterMutation(result *skillstore.Result) {
	if p == nil || result == nil || !result.Success {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.timer != nil {
		p.timer.Stop()
	}
	p.timer = time.AfterFunc(p.delay, func() {
		p.run(result.Action)
	})
}

func (p *PostProcessor) run(action string) {
	if p == nil {
		return
	}
	if p.sourceDir != "" {
		if err := engines.SyncSingerOSSkillsFrom(p.sourceDir, nil); err != nil {
			logs.Warnf("Sync SingerOS skills after %s failed: %v", action, err)
		}
	}
	if p.catalog != nil {
		if err := p.catalog.Reload(context.Background()); err != nil {
			logs.Warnf("Reload SingerOS skill catalog after %s failed: %v", action, err)
		}
	}
}
