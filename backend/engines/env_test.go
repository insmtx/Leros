package engines

import "testing"

func TestBuildRunEnvInjectsAnthropicModelEnv(t *testing.T) {
	env := BuildRunEnv([]string{"PATH=/bin"}, []string{"CUSTOM=1"}, ModelConfig{
		Provider: "anthropic",
		APIKey:   "key",
		BaseURL:  "https://example.test",
	})

	want := map[string]bool{
		"PATH=/bin":                false,
		"CUSTOM=1":                 false,
		"ANTHROPIC_API_KEY=key":    false,
		"ANTHROPIC_AUTH_TOKEN=key": false,
		"ANTHROPIC_BASE_URL=https://example.test": false,
	}
	for _, item := range env {
		if _, ok := want[item]; ok {
			want[item] = true
		}
	}
	for item, found := range want {
		if !found {
			t.Fatalf("missing env entry %q in %#v", item, env)
		}
	}
}

func TestBuildRunEnvInjectsOpenAIModelEnv(t *testing.T) {
	env := BuildRunEnv(nil, nil, ModelConfig{
		Provider: "openai",
		APIKey:   "key",
		BaseURL:  "https://openai.test/v1",
	})

	want := map[string]bool{
		"OPENAI_API_KEY=key":                     false,
		"OPENAI_API_BASE=https://openai.test/v1": false,
		"OPENAI_BASE_URL=https://openai.test/v1": false,
	}
	for _, item := range env {
		if _, ok := want[item]; ok {
			want[item] = true
		}
	}
	for item, found := range want {
		if !found {
			t.Fatalf("missing env entry %q in %#v", item, env)
		}
	}
}
