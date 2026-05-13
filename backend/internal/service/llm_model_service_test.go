package service

import (
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/insmtx/Leros/backend/internal/api/contract"
	dbrepo "github.com/insmtx/Leros/backend/internal/infra/db"
	"github.com/insmtx/Leros/backend/types"
)

func setupLLMModelServiceDB(t *testing.T) *gorm.DB {
	t.Helper()

	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	if err := database.AutoMigrate(&types.LLMModel{}); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return database
}

func setupLLMModelService(t *testing.T) (contract.LLMModelService, *gorm.DB) {
	t.Helper()

	database := setupLLMModelServiceDB(t)
	return NewLLMModelService(database), database
}

func countDefaultLLMModels(t *testing.T, database *gorm.DB, orgID uint) int64 {
	t.Helper()

	var count int64
	if err := database.Model(&types.LLMModel{}).
		Where("org_id = ? AND is_default = ?", orgID, true).
		Count(&count).Error; err != nil {
		t.Fatalf("count default llm models failed: %v", err)
	}
	return count
}

func TestCreateLLMModelGeneratesCodeDefaultsNameAndMasksAPIKey(t *testing.T) {
	service, database := setupLLMModelService(t)
	ctx := setupTestContextWithCaller(t)

	model, err := service.CreateLLMModel(ctx, &contract.CreateLLMModelRequest{
		Provider: string(types.LLMProviderOpenAI),
		Model:    "gpt-4o-mini",
		BaseURL:  "https://api.openai.com/v1",
		APIKey:   "sk-test-1234567890",
	})
	if err != nil {
		t.Fatalf("CreateLLMModel failed: %v", err)
	}

	if !strings.HasPrefix(model.Code, "llm_") {
		t.Fatalf("expected generated llm code, got %q", model.Code)
	}
	if model.Name != "gpt-4o-mini" {
		t.Fatalf("expected name to default to model, got %q", model.Name)
	}
	if model.BaseURL != "https://api.openai.com/v1" {
		t.Fatalf("expected normalized base_url, got %q", model.BaseURL)
	}
	if model.APIKey != "sk-***7890" {
		t.Fatalf("expected masked api key, got %q", model.APIKey)
	}
	if model.MaxTokens != 4096 || model.Temperature != 0.7 || model.TimeoutSec != 120 {
		t.Fatalf("unexpected defaults: max_tokens=%d temperature=%v timeout_sec=%d", model.MaxTokens, model.Temperature, model.TimeoutSec)
	}

	stored, err := dbrepo.GetLLMModelByID(ctx, database, model.ID)
	if err != nil {
		t.Fatalf("GetLLMModelByID failed: %v", err)
	}
	if stored.APIKeyEncrypted != "sk-test-1234567890" {
		t.Fatalf("expected stored api key to match input, got %q", stored.APIKeyEncrypted)
	}
	if stored.APIKeyMasked != "sk-***7890" {
		t.Fatalf("expected stored masked api key, got %q", stored.APIKeyMasked)
	}
}

func TestCreateLLMModelRequiresAPIKey(t *testing.T) {
	service, _ := setupLLMModelService(t)
	ctx := setupTestContextWithCaller(t)

	_, err := service.CreateLLMModel(ctx, &contract.CreateLLMModelRequest{
		Provider: string(types.LLMProviderOpenAI),
		Model:    "gpt-4o-mini",
		BaseURL:  "https://api.openai.com/v1",
	})
	if err == nil {
		t.Fatal("expected error for missing api_key")
	}
	if err.Error() != "api_key is required" {
		t.Fatalf("expected api_key required error, got %q", err.Error())
	}
}

func TestCreateLLMModelRequiresBaseURL(t *testing.T) {
	service, _ := setupLLMModelService(t)
	ctx := setupTestContextWithCaller(t)

	_, err := service.CreateLLMModel(ctx, &contract.CreateLLMModelRequest{
		Provider: string(types.LLMProviderOpenAI),
		Model:    "gpt-4o-mini",
		APIKey:   "sk-test-1234567890",
	})
	if err == nil {
		t.Fatal("expected error for missing base_url")
	}
	if err.Error() != "base_url is required" {
		t.Fatalf("expected base_url required error, got %q", err.Error())
	}
}

func TestCreateLLMModelTrimsChatCompletionsPath(t *testing.T) {
	service, database := setupLLMModelService(t)
	ctx := setupTestContextWithCaller(t)

	model, err := service.CreateLLMModel(ctx, &contract.CreateLLMModelRequest{
		Provider: string(types.LLMProviderOpenAI),
		Model:    "gpt-4o-mini",
		BaseURL:  "https://api.openai.com/v1/chat/completions",
		APIKey:   "sk-test-1234567890",
	})
	if err != nil {
		t.Fatalf("CreateLLMModel failed: %v", err)
	}
	if model.BaseURL != "https://api.openai.com/v1" {
		t.Fatalf("expected normalized base_url in response, got %q", model.BaseURL)
	}

	stored, err := dbrepo.GetLLMModelByID(ctx, database, model.ID)
	if err != nil {
		t.Fatalf("GetLLMModelByID failed: %v", err)
	}
	if stored.BaseURL != "https://api.openai.com/v1" {
		t.Fatalf("expected normalized base_url in database, got %q", stored.BaseURL)
	}
}

func TestCreateLLMModelKeepsSingleDefault(t *testing.T) {
	service, database := setupLLMModelService(t)
	ctx := setupTestContextWithCaller(t)

	first, err := service.CreateLLMModel(ctx, &contract.CreateLLMModelRequest{
		Provider:  string(types.LLMProviderOpenAI),
		Model:     "gpt-4o-mini",
		BaseURL:   "https://api.openai.com/v1",
		APIKey:    "sk-test-1234567890",
		IsDefault: true,
	})
	if err != nil {
		t.Fatalf("first CreateLLMModel failed: %v", err)
	}
	second, err := service.CreateLLMModel(ctx, &contract.CreateLLMModelRequest{
		Provider:  string(types.LLMProviderDeepSeek),
		Model:     "deepseek-chat",
		BaseURL:   "https://api.deepseek.com/v1",
		APIKey:    "sk-test-abcdefgh",
		IsDefault: true,
	})
	if err != nil {
		t.Fatalf("second CreateLLMModel failed: %v", err)
	}

	if count := countDefaultLLMModels(t, database, 1); count != 1 {
		t.Fatalf("expected one default llm model, got %d", count)
	}
	storedFirst, err := dbrepo.GetLLMModelByID(ctx, database, first.ID)
	if err != nil {
		t.Fatalf("GetLLMModelByID failed: %v", err)
	}
	if storedFirst.IsDefault {
		t.Fatal("expected first model default flag to be cleared")
	}
	storedSecond, err := dbrepo.GetLLMModelByID(ctx, database, second.ID)
	if err != nil {
		t.Fatalf("GetLLMModelByID failed: %v", err)
	}
	if !storedSecond.IsDefault {
		t.Fatal("expected second model to be default")
	}
}

func TestUpdateLLMModelKeepsAPIKeyWhenOmitted(t *testing.T) {
	service, database := setupLLMModelService(t)
	ctx := setupTestContextWithCaller(t)

	model, err := service.CreateLLMModel(ctx, &contract.CreateLLMModelRequest{
		Name:     "主模型",
		Provider: string(types.LLMProviderOpenAI),
		Model:    "gpt-4o-mini",
		BaseURL:  "https://api.openai.com/v1",
		APIKey:   "sk-test-1234567890",
	})
	if err != nil {
		t.Fatalf("CreateLLMModel failed: %v", err)
	}

	updated, err := service.UpdateLLMModel(ctx, model.ID, &contract.UpdateLLMModelRequest{
		Name: "更新后的主模型",
	})
	if err != nil {
		t.Fatalf("UpdateLLMModel failed: %v", err)
	}
	if updated.APIKey != "sk-***7890" {
		t.Fatalf("expected response to keep masked api key, got %q", updated.APIKey)
	}

	stored, err := dbrepo.GetLLMModelByID(ctx, database, model.ID)
	if err != nil {
		t.Fatalf("GetLLMModelByID failed: %v", err)
	}
	if stored.APIKeyEncrypted != "sk-test-1234567890" {
		t.Fatalf("expected api key to remain unchanged, got %q", stored.APIKeyEncrypted)
	}
	if stored.APIKeyMasked != "sk-***7890" {
		t.Fatalf("expected masked api key to remain unchanged, got %q", stored.APIKeyMasked)
	}
}

func TestUpdateLLMModelKeepsSingleDefault(t *testing.T) {
	service, database := setupLLMModelService(t)
	ctx := setupTestContextWithCaller(t)

	first, err := service.CreateLLMModel(ctx, &contract.CreateLLMModelRequest{
		Provider:  string(types.LLMProviderOpenAI),
		Model:     "gpt-4o-mini",
		BaseURL:   "https://api.openai.com/v1",
		APIKey:    "sk-test-1234567890",
		IsDefault: true,
	})
	if err != nil {
		t.Fatalf("first CreateLLMModel failed: %v", err)
	}
	second, err := service.CreateLLMModel(ctx, &contract.CreateLLMModelRequest{
		Provider: string(types.LLMProviderDeepSeek),
		Model:    "deepseek-chat",
		BaseURL:  "https://api.deepseek.com/v1",
		APIKey:   "sk-test-abcdefgh",
	})
	if err != nil {
		t.Fatalf("second CreateLLMModel failed: %v", err)
	}

	isDefault := true
	if _, err := service.UpdateLLMModel(ctx, second.ID, &contract.UpdateLLMModelRequest{
		IsDefault: &isDefault,
	}); err != nil {
		t.Fatalf("UpdateLLMModel failed: %v", err)
	}

	if count := countDefaultLLMModels(t, database, 1); count != 1 {
		t.Fatalf("expected one default llm model, got %d", count)
	}
	storedFirst, err := dbrepo.GetLLMModelByID(ctx, database, first.ID)
	if err != nil {
		t.Fatalf("GetLLMModelByID failed: %v", err)
	}
	if storedFirst.IsDefault {
		t.Fatal("expected first model default flag to be cleared")
	}
	storedSecond, err := dbrepo.GetLLMModelByID(ctx, database, second.ID)
	if err != nil {
		t.Fatalf("GetLLMModelByID failed: %v", err)
	}
	if !storedSecond.IsDefault {
		t.Fatal("expected second model to be default")
	}
}

func TestDeleteLLMModelDoesNotLeaveMultipleDefaults(t *testing.T) {
	service, database := setupLLMModelService(t)
	ctx := setupTestContextWithCaller(t)

	model, err := service.CreateLLMModel(ctx, &contract.CreateLLMModelRequest{
		Provider:  string(types.LLMProviderOpenAI),
		Model:     "gpt-4o-mini",
		BaseURL:   "https://api.openai.com/v1",
		APIKey:    "sk-test-1234567890",
		IsDefault: true,
	})
	if err != nil {
		t.Fatalf("CreateLLMModel failed: %v", err)
	}

	if err := service.DeleteLLMModel(ctx, model.ID); err != nil {
		t.Fatalf("DeleteLLMModel failed: %v", err)
	}
	if count := countDefaultLLMModels(t, database, 1); count != 0 {
		t.Fatalf("expected no default llm model after deleting default, got %d", count)
	}
}

func TestUpdateLLMModelTrimsChatCompletionsPath(t *testing.T) {
	service, database := setupLLMModelService(t)
	ctx := setupTestContextWithCaller(t)

	model, err := service.CreateLLMModel(ctx, &contract.CreateLLMModelRequest{
		Provider: string(types.LLMProviderOpenAI),
		Model:    "gpt-4o-mini",
		BaseURL:  "https://api.openai.com/v1",
		APIKey:   "sk-test-1234567890",
	})
	if err != nil {
		t.Fatalf("CreateLLMModel failed: %v", err)
	}

	baseURL := "https://example.com/v1/chat/completions/"
	updated, err := service.UpdateLLMModel(ctx, model.ID, &contract.UpdateLLMModelRequest{
		BaseURL: &baseURL,
	})
	if err != nil {
		t.Fatalf("UpdateLLMModel failed: %v", err)
	}
	if updated.BaseURL != "https://example.com/v1" {
		t.Fatalf("expected normalized base_url in response, got %q", updated.BaseURL)
	}

	stored, err := dbrepo.GetLLMModelByID(ctx, database, model.ID)
	if err != nil {
		t.Fatalf("GetLLMModelByID failed: %v", err)
	}
	if stored.BaseURL != "https://example.com/v1" {
		t.Fatalf("expected normalized base_url in database, got %q", stored.BaseURL)
	}
}

func TestNormalizeLLMBaseURLTrimsKnownEndpointSuffixes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		baseURL string
		want    string
	}{
		{name: "openai chat completions", baseURL: "https://api.example.com/v1/chat/completions", want: "https://api.example.com"},
		{name: "openai completions", baseURL: "https://api.example.com/v1/completions", want: "https://api.example.com"},
		{name: "openai responses", baseURL: "https://api.example.com/v1/responses", want: "https://api.example.com"},
		{name: "anthropic messages", baseURL: "https://api.anthropic.com/v1/messages", want: "https://api.anthropic.com"},
		{name: "ollama chat", baseURL: "http://localhost:11434/api/chat", want: "http://localhost:11434"},
		{name: "ollama generate", baseURL: "http://localhost:11434/api/generate", want: "http://localhost:11434"},
		{name: "gemini generate content", baseURL: "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent", want: "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro"},
		{name: "gemini stream generate content", baseURL: "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:streamGenerateContent", want: "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro"},
		{name: "trailing slash", baseURL: "https://api.example.com/v1/chat/completions/", want: "https://api.example.com"},
		{name: "v1 suffix", baseURL: "https://api.example.com/v1", want: "https://api.example.com"},
		{name: "base url unchanged", baseURL: "https://api.example.com/", want: "https://api.example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := normalizeLLMBaseURL(tt.baseURL); got != tt.want {
				t.Fatalf("normalizeLLMBaseURL(%q) = %q, want %q", tt.baseURL, got, tt.want)
			}
		})
	}
}

func TestUpdateLLMModelUpdatesMaskedAPIKeyWhenProvided(t *testing.T) {
	service, database := setupLLMModelService(t)
	ctx := setupTestContextWithCaller(t)

	model, err := service.CreateLLMModel(ctx, &contract.CreateLLMModelRequest{
		Provider: string(types.LLMProviderOpenAI),
		Model:    "gpt-4o-mini",
		BaseURL:  "https://api.openai.com/v1",
		APIKey:   "sk-test-1234567890",
	})
	if err != nil {
		t.Fatalf("CreateLLMModel failed: %v", err)
	}

	newAPIKey := "sk-new-abcdefgh"
	updated, err := service.UpdateLLMModel(ctx, model.ID, &contract.UpdateLLMModelRequest{
		APIKey: &newAPIKey,
	})
	if err != nil {
		t.Fatalf("UpdateLLMModel failed: %v", err)
	}
	if updated.APIKey != "sk-***efgh" {
		t.Fatalf("expected response to use new masked api key, got %q", updated.APIKey)
	}

	stored, err := dbrepo.GetLLMModelByID(ctx, database, model.ID)
	if err != nil {
		t.Fatalf("GetLLMModelByID failed: %v", err)
	}
	if stored.APIKeyEncrypted != "sk-new-abcdefgh" {
		t.Fatalf("expected api key to update, got %q", stored.APIKeyEncrypted)
	}
	if stored.APIKeyMasked != "sk-***efgh" {
		t.Fatalf("expected masked api key to update, got %q", stored.APIKeyMasked)
	}
}
