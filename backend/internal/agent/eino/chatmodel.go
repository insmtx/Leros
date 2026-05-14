package eino

import (
	"context"
	"fmt"
	"strings"

	einoark "github.com/cloudwego/eino-ext/components/model/ark"
	einoclaude "github.com/cloudwego/eino-ext/components/model/claude"
	einodeepseek "github.com/cloudwego/eino-ext/components/model/deepseek"
	einogemini "github.com/cloudwego/eino-ext/components/model/gemini"
	einoopenai "github.com/cloudwego/eino-ext/components/model/openai"
	einoopenrouter "github.com/cloudwego/eino-ext/components/model/openrouter"
	einoqwen "github.com/cloudwego/eino-ext/components/model/qwen"
	einomodel "github.com/cloudwego/eino/components/model"
	"google.golang.org/genai"

	"github.com/insmtx/Leros/backend/config"
	"github.com/insmtx/Leros/backend/types"
)

// NewChatModel creates an Eino chat model using the adapter mapped from provider.
func NewChatModel(ctx context.Context, cfg *config.LLMConfig) (einomodel.ToolCallingChatModel, error) {
	if cfg == nil {
		return nil, fmt.Errorf("llm config is required")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("llm api key is required")
	}
	if strings.TrimSpace(cfg.Model) == "" {
		return nil, fmt.Errorf("llm model is required")
	}

	provider := types.LLMProviderType(strings.TrimSpace(cfg.Provider))
	switch provider {
	case types.LLMProviderOpenAI, types.LLMProviderCustom:
		return newOpenAICompatibleChatModel(ctx, cfg)
	case types.LLMProviderAnthropic:
		return newClaudeChatModel(ctx, cfg)
	case types.LLMProviderQwen:
		return newQwenChatModel(ctx, cfg)
	case types.LLMProviderDeepSeek:
		return newDeepSeekChatModel(ctx, cfg)
	case types.LLMProviderGemini:
		return newGeminiChatModel(ctx, cfg)
	case types.LLMProviderArk:
		return newArkChatModel(ctx, cfg)
	case types.LLMProviderOpenRouter:
		return newOpenRouterChatModel(ctx, cfg)
	default:
		return nil, fmt.Errorf("llm provider %q is not supported", cfg.Provider)
	}
}

func newOpenAICompatibleChatModel(ctx context.Context, cfg *config.LLMConfig) (einomodel.ToolCallingChatModel, error) {
	chatModel, err := einoopenai.NewChatModel(ctx, &einoopenai.ChatModelConfig{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		Model:   cfg.Model,
	})
	if err != nil {
		return nil, fmt.Errorf("create eino openai chat model: %w", err)
	}

	return chatModel, nil
}

func newClaudeChatModel(ctx context.Context, cfg *config.LLMConfig) (einomodel.ToolCallingChatModel, error) {
	var baseURL *string
	if strings.TrimSpace(cfg.BaseURL) != "" {
		baseURL = &cfg.BaseURL
	}
	chatModel, err := einoclaude.NewChatModel(ctx, &einoclaude.Config{
		APIKey:    cfg.APIKey,
		BaseURL:   baseURL,
		Model:     cfg.Model,
		MaxTokens: 4096,
	})
	if err != nil {
		return nil, fmt.Errorf("create eino claude chat model: %w", err)
	}
	return chatModel, nil
}

func newQwenChatModel(ctx context.Context, cfg *config.LLMConfig) (einomodel.ToolCallingChatModel, error) {
	chatModel, err := einoqwen.NewChatModel(ctx, &einoqwen.ChatModelConfig{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		Model:   cfg.Model,
	})
	if err != nil {
		return nil, fmt.Errorf("create eino qwen chat model: %w", err)
	}
	return chatModel, nil
}

func newDeepSeekChatModel(ctx context.Context, cfg *config.LLMConfig) (einomodel.ToolCallingChatModel, error) {
	chatModel, err := einodeepseek.NewChatModel(ctx, &einodeepseek.ChatModelConfig{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		Model:   cfg.Model,
	})
	if err != nil {
		return nil, fmt.Errorf("create eino deepseek chat model: %w", err)
	}
	return chatModel, nil
}

func newGeminiChatModel(ctx context.Context, cfg *config.LLMConfig) (einomodel.ToolCallingChatModel, error) {
	clientConfig := &genai.ClientConfig{
		APIKey: cfg.APIKey,
	}
	if strings.TrimSpace(cfg.BaseURL) != "" {
		clientConfig.HTTPOptions = genai.HTTPOptions{
			BaseURL: cfg.BaseURL,
		}
	}
	client, err := genai.NewClient(ctx, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("create gemini client: %w", err)
	}
	chatModel, err := einogemini.NewChatModel(ctx, &einogemini.Config{
		Client: client,
		Model:  cfg.Model,
	})
	if err != nil {
		return nil, fmt.Errorf("create eino gemini chat model: %w", err)
	}
	return chatModel, nil
}

func newArkChatModel(ctx context.Context, cfg *config.LLMConfig) (einomodel.ToolCallingChatModel, error) {
	chatModel, err := einoark.NewChatModel(ctx, &einoark.ChatModelConfig{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		Model:   cfg.Model,
	})
	if err != nil {
		return nil, fmt.Errorf("create eino ark chat model: %w", err)
	}
	return chatModel, nil
}

func newOpenRouterChatModel(ctx context.Context, cfg *config.LLMConfig) (einomodel.ToolCallingChatModel, error) {
	chatModel, err := einoopenrouter.NewChatModel(ctx, &einoopenrouter.Config{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		Model:   cfg.Model,
	})
	if err != nil {
		return nil, fmt.Errorf("create eino openrouter chat model: %w", err)
	}
	return chatModel, nil
}
