package runtime

import (
	"context"
	"fmt"
	"strings"

	"github.com/insmtx/Leros/backend/config"
	"github.com/insmtx/Leros/backend/engines"
	"github.com/insmtx/Leros/backend/engines/builtin"
	"github.com/insmtx/Leros/backend/internal/agent"
	"github.com/insmtx/Leros/backend/internal/runtime/deps"
	"github.com/insmtx/Leros/backend/internal/runtime/drivers/externalcli"
	"github.com/insmtx/Leros/backend/internal/runtime/lifecycle"
	lifecyclecontext "github.com/insmtx/Leros/backend/internal/runtime/lifecycle/context"
	"github.com/insmtx/Leros/backend/internal/runtime/lifecycle/steps"
	"github.com/ygpkg/yg-go/logs"
)

type Options struct {
	LLMConfig      *config.LLMConfig
	CLIConfig      *config.CLIEnginesConfig
	DefaultRuntime string
	CLISkillDirs   []string
}

type Service struct {
	env    *deps.Container
	router agent.Runner
}

func NewService(ctx context.Context, opts Options) (*Service, error) {
	env, err := deps.New(ctx, deps.Options{
		CLISkillDirs: opts.CLISkillDirs,
	})
	if err != nil {
		return nil, fmt.Errorf("create runtime dependencies: %w", err)
	}

	s := &Service{env: env}

	router, err := s.buildRouter(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("build runtime router: %w", err)
	}

	s.router = router
	return s, nil
}

func (s *Service) Router() agent.Runner {
	return s.router
}

// Run 通过配置的运行时路由器执行请求。
func (s *Service) Run(ctx context.Context, req *agent.RequestContext) (*agent.RunResult, error) {
	if s == nil || s.router == nil {
		return nil, fmt.Errorf("agent runtime service is not initialized")
	}
	return s.router.Run(ctx, req)
}

func (s *Service) Environment() *deps.Container {
	return s.env
}

func (s *Service) buildRouter(ctx context.Context, opts Options) (agent.Runner, error) {
	lifecycleBuilder := lifecyclecontext.NewContextBuilder(lifecyclecontext.ContextBuilder{
		SessionMessages: lifecyclecontext.NewPassthroughSessionMessageProvider(),
	})
	router := agent.NewRuntimeRouter(agent.RuntimeKindLeros)

	registered := 0
	registeredKinds := make(map[string]struct{})

	// 统一创建引擎注册表（始终包含 native，可选包含 CLI）。
	engineRegistry, err := builtin.NewRegistryFromConfig(opts.CLIConfig, s.env)
	if err != nil {
		return nil, fmt.Errorf("create engine registry: %w", err)
	}

	for _, name := range engineRegistry.Names() {
		engine, ok := engineRegistry.Get(name)
		if !ok {
			continue
		}

		runner, err := externalcli.NewRunner(name, engine)
		if err != nil {
			return nil, err
		}
		// 仅外部 CLI 引擎需要审批路由器；native 的 Approver 是 noop。
		if name != engines.EngineNative {
			runner.SetApprovalHandler(engines.DefaultApprovalRouter)
		}
		logs.Infof("Registering agent runtime: %s", name)

		if err := router.Register(name, runner); err != nil {
			return nil, err
		}
		registered++
		registeredKinds[strings.ToLower(strings.TrimSpace(name))] = struct{}{}
	}

	if registered == 0 {
		return nil, fmt.Errorf("no agent runtime is available")
	}

	engineNames := engineRegistry.Names()
	selectedDefault := s.selectDefaultRuntime(opts.DefaultRuntime, opts, engineNames)
	if selectedDefault == "" {
		selectedDefault = engines.EngineNative
	}
	normalizedDefault := strings.ToLower(strings.TrimSpace(selectedDefault))
	if _, ok := registeredKinds[normalizedDefault]; !ok {
		return nil, fmt.Errorf("default agent runtime %q is not available", selectedDefault)
	}
	router.SetDefault(selectedDefault)

	runner := lifecycle.NewRunner(router, lifecycleBuilder, s.env)
	runner.SetArtifactRecorder(steps.NewWorkspaceArtifactRecorder())
	return runner, nil
}

var _ agent.Runner = (*Service)(nil)

func (s *Service) selectDefaultRuntime(defaultRuntime string, opts Options, cliNames []string) string {
	if strings.TrimSpace(defaultRuntime) != "" {
		return defaultRuntime
	}
	if opts.CLIConfig != nil && strings.TrimSpace(opts.CLIConfig.Default) != "" {
		return opts.CLIConfig.Default
	}
	return agent.RuntimeKindLeros
}
