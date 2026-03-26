package config

type GithubAppConfig struct {
	AppID         int64  `yaml:"app_id"`
	PrivateKey    string `yaml:"private_key"`
	WebhookSecret string `yaml:"webhook_secret"`
	BaseURL       string `yaml:"base_url"`
}
