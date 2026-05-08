package runtime

import (
	"context"
	"fmt"
	"strings"

	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/internal/agent"
	"github.com/insmtx/SingerOS/backend/internal/agent/externalcli"
	"github.com/insmtx/SingerOS/backend/internal/agent/lifecycle"
	"github.com/insmtx/SingerOS/backend/internal/agent/runtimeenv"
	"github.com/insmtx/SingerOS/backend/internal/agent/singeros"
	"github.com/insmtx/SingerOS/backend/runtime/engines/builtin"
	"github.com/ygpkg/yg-go/logs"
)

type Options struct {
	LLMConfig      *config.LLMConfig
	CLIConfig      *config.CLIEnginesConfig
	ToolsEnabled   bool
	DefaultRuntime string
}

type Service struct {
	env    *runtimeenv.Environment
	router agent.Runner
}

func NewService(ctx context.Context, opts Options) (*Service, error) {
	env, err := runtimeenv.New(ctx, runtimeenv.Options{
		ToolsEnabled: opts.ToolsEnabled,
	})
	if err != nil {
		return nil, fmt.Errorf("create runtime environment: %w", err)
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

// Run executes a request through the configured runtime router.
func (s *Service) Run(ctx context.Context, req *agent.RequestContext) (*agent.RunResult, error) {
	if s == nil || s.router == nil {
		return nil, fmt.Errorf("agent runtime service is not initialized")
	}
	return s.router.Run(ctx, req)
}

func (s *Service) Environment() *runtimeenv.Environment {
	return s.env
}

func (s *Service) buildRouter(ctx context.Context, opts Options) (agent.Runner, error) {
	lifecycleBuilder := lifecycle.NewContextBuilder(lifecycle.ContextBuilder{
		BaseSystemPrompt: singeros.DefaultSystemPrompt(),
		Runtime:          s.env,
	})
	router := lifecycle.NewRuntimeRouter(agent.RuntimeKindSingerOS, lifecycleBuilder, s.env)

	registered := 0
	registeredKinds := make(map[string]struct{})
	cliNames := []string{}

	if opts.LLMConfig != nil && opts.LLMConfig.APIKey != "" {
		switch opts.LLMConfig.Provider {
		case "", "openai":
			logs.Info("Registering SingerOS agent runtime")
			singerRunner, err := singeros.NewRunner(ctx, opts.LLMConfig, s.env)
			if err != nil {
				return nil, err
			}
			if err := router.Register(agent.RuntimeKindSingerOS, singerRunner); err != nil {
				return nil, err
			}
			registered++
			registeredKinds[agent.RuntimeKindSingerOS] = struct{}{}
		default:
			logs.Warnf("Skipping SingerOS agent runtime for unsupported Eino chat model provider: %s", opts.LLMConfig.Provider)
		}
	}

	if opts.CLIConfig != nil {
		cliRegistry, err := builtin.NewRegistryFromConfig(opts.CLIConfig)
		if err != nil {
			return nil, fmt.Errorf("create CLI engine registry: %w", err)
		}
		cliNames = cliRegistry.Names()
		for _, name := range cliNames {
			engine, ok := cliRegistry.Get(name)
			if !ok {
				continue
			}
			runner, err := externalcli.NewRunner(name, engine, opts.LLMConfig)
			if err != nil {
				return nil, err
			}
			if err := router.Register(name, runner); err != nil {
				return nil, err
			}
			registered++
			registeredKinds[strings.ToLower(strings.TrimSpace(name))] = struct{}{}
			logs.Infof("Registering external agent CLI runtime: %s", name)
		}
	}

	if registered == 0 {
		return nil, fmt.Errorf("no agent runtime is available")
	}

	selectedDefault := s.selectDefaultRuntime(opts.DefaultRuntime, opts, cliNames)
	if selectedDefault == "" {
		selectedDefault = agent.RuntimeKindSingerOS
	}
	normalizedDefault := strings.ToLower(strings.TrimSpace(selectedDefault))
	if _, ok := registeredKinds[normalizedDefault]; !ok {
		return nil, fmt.Errorf("default agent runtime %q is not available", selectedDefault)
	}
	router.SetDefault(selectedDefault)

	return router, nil
}

var _ agent.Runner = (*Service)(nil)

func (s *Service) selectDefaultRuntime(defaultRuntime string, opts Options, cliNames []string) string {
	if strings.TrimSpace(defaultRuntime) != "" {
		return defaultRuntime
	}
	if opts.CLIConfig != nil && strings.TrimSpace(opts.CLIConfig.Default) != "" {
		return opts.CLIConfig.Default
	}
	if opts.LLMConfig != nil && opts.LLMConfig.APIKey != "" {
		return agent.RuntimeKindSingerOS
	}
	if len(cliNames) > 0 {
		return cliNames[0]
	}
	return ""
}
