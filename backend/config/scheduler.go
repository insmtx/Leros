package config

type SchedulerConfig struct {
	WorkerBinary string
	WorkingDir   string
	Env          map[string]string
	ServerAddr   string
}
