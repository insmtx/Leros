package config

type RabbitMQConfig struct {
	URL      string `yaml:"url,omitempty" json:"url,omitempty"`
	Host     string `yaml:"host,omitempty" json:"host,omitempty"`
	Port     int    `yaml:"port,omitempty" json:"port,omitempty"`
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	Password string `yaml:"password,omitempty" json:"password,omitempty"`
}

type Config struct {
	Github   *GithubAppConfig `yaml:"github,omitempty"`
	Gitlab   *GitlabAppConfig `yaml:"gitlab,omitempty"`
	RabbitMQ *RabbitMQConfig  `yaml:"rabbitmq,omitempty"`
}
