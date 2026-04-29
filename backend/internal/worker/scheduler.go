package worker

import (
	"context"
	"time"
)

type WorkerEnvType string

const (
	WorkerEnvProcess  WorkerEnvType = "process"
	WorkerEnvDocker   WorkerEnvType = "docker"
	WorkerEnvKubeVirt WorkerEnvType = "kubevirt"
)

type WorkerScheduler interface {
	Start(ctx context.Context, spec *WorkerSpec) (*WorkerInstance, error)
	Stop(ctx context.Context, workerID string) error
	Health(ctx context.Context, workerID string) error
	List(ctx context.Context) ([]*WorkerInstance, error)
}

type WorkerSpec struct {
	ID          string
	Name        string
	Labels      map[string]string
	Annotations map[string]string
	EnvType     WorkerEnvType
	Image       string
	Command     []string
	Args        []string
	Env         map[string]string
	WorkingDir  string
}

type WorkerInstance struct {
	ID        string
	WorkerID  string
	Status    string
	PID       int
	StartedAt time.Time
	Endpoint  string
}
