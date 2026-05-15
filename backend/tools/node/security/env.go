package security

import (
	"os"
	"strings"
)

// SanitizedEnv 净化环境变量，移除敏感凭证
func SanitizedEnv(extra map[string]string) []string {
	blocked := map[string]struct{}{
		"ANTHROPIC_API_KEY":             {},
		"CLAUDE_CODE_OAUTH_TOKEN":       {},
		"DEEPSEEK_API_KEY":              {},
		"FEISHU_APP_SECRET":             {},
		"FIRECRAWL_API_KEY":             {},
		"GH_TOKEN":                      {},
		"GITHUB_APP_PRIVATE_KEY":        {},
		"GITHUB_APP_PRIVATE_KEY_PATH":   {},
		"GITHUB_TOKEN":                  {},
		"GOOGLE_API_KEY":                {},
		"OPENAI_API_KEY":                {},
		"OPENROUTER_API_KEY":            {},
		"SLACK_BOT_TOKEN":               {},
		"SLACK_SIGNING_SECRET":          {},
		"WEWORK_CORP_SECRET":            {},
		"WEWORK_WEBHOOK_SIGNING_SECRET": {},
	}

	env := make(map[string]string)
	for _, item := range os.Environ() {
		key, value, ok := strings.Cut(item, "=")
		if !ok {
			continue
		}
		if _, denied := blocked[strings.ToUpper(key)]; denied {
			continue
		}
		env[key] = value
	}
	for key, value := range extra {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if _, denied := blocked[strings.ToUpper(key)]; denied {
			continue
		}
		env[key] = value
	}

	result := make([]string, 0, len(env))
	for key, value := range env {
		result = append(result, key+"="+value)
	}
	return result
}
