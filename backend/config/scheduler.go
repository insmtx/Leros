package config

type SchedulerConfig struct {
	Mode         string            `yaml:"mode,omitempty" json:"mode,omitempty"` // 调度模式，支持 "process","docker-cli","docker-api","k8s"
	WorkerBinary string            `yaml:"worker_binary,omitempty" json:"worker_binary,omitempty"`
	WorkingDir   string            `yaml:"working_dir,omitempty" json:"working_dir,omitempty"`
	Env          map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	ServerAddr   string            `yaml:"server_addr,omitempty" json:"server_addr,omitempty"`
}
