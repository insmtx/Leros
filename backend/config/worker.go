package config

type WorkerConfig struct {
	OrgID    uint `yaml:"org_id" json:"org_id"`
	WorkerID uint `yaml:"worker_id" json:"worker_id"`

	ServerAddr   string `yaml:"server_addr,omitempty" json:"server_addr,omitempty"`
	SkillsDir    string `yaml:"skills_dir,omitempty" json:"skills_dir,omitempty"`
	ToolsEnabled bool   `yaml:"tools_enabled,omitempty" json:"tools_enabled,omitempty"`

	NATS     *NATSConfig       `yaml:"nats,omitempty"`
	Database *DatabaseConfig   `yaml:"database,omitempty"`
	LLM      *LLMConfig        `yaml:"llm,omitempty" json:"llm,omitempty"`
	CLI      *CLIEnginesConfig `yaml:"cli,omitempty"`
}

// CLIEnginesConfig is the configuration for external AI coding CLIs.
type CLIEnginesConfig struct {
	Default string `yaml:"default,omitempty" json:"default,omitempty"`
}
