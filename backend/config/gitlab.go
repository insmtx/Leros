package config

type GitlabAppConfig struct {
	AppID         int64  `yaml:"app_id,omitempty"`
	PrivateKey    string `yaml:"private_key,omitempty"`
	WebhookSecret string `yaml:"webhook_secret,omitempty"`
	BaseURL       string `yaml:"base_url,omitempty"`
}
