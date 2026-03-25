package llm

import "context"

type Provider interface {
	Name() string
	Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
	GenerateStream(ctx context.Context, req *GenerateRequest) (<-chan StreamChunk, error)
	CountTokens(text string) int
	Models() []string
}

type GenerateRequest struct {
	Messages    []Message
	Model       string
	MaxTokens   int
	Temperature float64
	Stop        []string
}

type GenerateResponse struct {
	Content      string
	Usage        TokenUsage
	FinishReason string
}

type Message struct {
	Role    string
	Content string
}

type StreamChunk struct {
	Content string
	Done    bool
}

type TokenUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}
