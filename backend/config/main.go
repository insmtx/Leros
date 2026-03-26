package config

type RabbitMQConfig struct {
	URL string `yaml:"url,omitempty" json:"url,omitempty"`
}

type Config struct {
	Github   *GithubAppConfig `yaml:"github,omitempty"`
	Gitlab   *GitlabAppConfig `yaml:"gitlab,omitempty"`
	RabbitMQ *RabbitMQConfig  `yaml:"rabbitmq,omitempty"`
}
