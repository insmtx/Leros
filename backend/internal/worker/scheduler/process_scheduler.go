package scheduler

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/internal/worker"
	"github.com/ygpkg/yg-go/logs"
)

type ProcessScheduler struct {
	config    *config.SchedulerConfig
	instances map[string]*ProcessInstance
	mu        sync.RWMutex
}

type ProcessInstance struct {
	ID        string
	WorkerID  string
	Cmd       *exec.Cmd
	Process   *os.Process
	Status    string
	StartedAt time.Time
	LastSeen  time.Time
	mu        sync.RWMutex
}

func NewProcessScheduler(config *config.SchedulerConfig) worker.WorkerScheduler {
	return &ProcessScheduler{
		config:    config,
		instances: make(map[string]*ProcessInstance),
	}
}

func (ps *ProcessScheduler) Start(ctx context.Context, spec *worker.WorkerSpec) (*worker.WorkerInstance, error) {
	if spec.EnvType != "" && spec.EnvType != worker.WorkerEnvProcess {
		return nil, fmt.Errorf("unsupported env type: %s, ProcessScheduler only supports process runtime", spec.EnvType)
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()

	workerID := spec.ID
	if workerID == "" {
		workerID = fmt.Sprintf("worker_%d", time.Now().UnixNano())
	}

	instance := &ProcessInstance{
		ID:        workerID,
		WorkerID:  workerID,
		Status:    "initializing",
		StartedAt: time.Now(),
		LastSeen:  time.Now(),
	}

	if err := ps.startProcess(instance, spec); err != nil {
		return nil, fmt.Errorf("failed to start process: %w", err)
	}

	ps.instances[workerID] = instance

	return &worker.WorkerInstance{
		ID:        instance.ID,
		WorkerID:  instance.WorkerID,
		Status:    instance.Status,
		PID:       instance.Process.Pid,
		StartedAt: instance.StartedAt,
	}, nil
}

func (ps *ProcessScheduler) Stop(ctx context.Context, workerID string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	instance, ok := ps.instances[workerID]
	if !ok {
		return fmt.Errorf("worker %s not found", workerID)
	}

	if err := ps.stopProcess(instance); err != nil {
		return fmt.Errorf("failed to stop process: %w", err)
	}

	instance.Status = "stopped"
	delete(ps.instances, workerID)
	return nil
}

func (ps *ProcessScheduler) Health(ctx context.Context, workerID string) error {
	ps.mu.RLock()
	instance, ok := ps.instances[workerID]
	ps.mu.RUnlock()

	if !ok {
		return fmt.Errorf("worker %s not found", workerID)
	}

	return ps.healthCheck(instance)
}

func (ps *ProcessScheduler) List(ctx context.Context) ([]*worker.WorkerInstance, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	result := make([]*worker.WorkerInstance, 0, len(ps.instances))
	for _, instance := range ps.instances {
		instance.mu.RLock()
		result = append(result, &worker.WorkerInstance{
			ID:        instance.ID,
			WorkerID:  instance.WorkerID,
			Status:    instance.Status,
			PID:       instance.Process.Pid,
			StartedAt: instance.StartedAt,
		})
		instance.mu.RUnlock()
	}
	return result, nil
}

func (ps *ProcessScheduler) startProcess(instance *ProcessInstance, spec *worker.WorkerSpec) error {
	cmdPath := ps.config.WorkerBinary
	if cmdPath == "" {
		cmdPath = "./bundles/singer"
	}

	if _, err := os.Stat(cmdPath); os.IsNotExist(err) {
		return fmt.Errorf("worker binary not found: %s", cmdPath)
	}

	env := os.Environ()
	for key, value := range ps.config.Env {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	workDir := ps.config.WorkingDir
	if workDir == "" {
		workDir = filepath.Dir(cmdPath)
	}

	args := []string{cmdPath, "worker"}
	if spec.ID != "" {
		args = append(args, "--assistant-code", spec.ID)
	}
	args = append(args, "--server-addr", ps.config.ServerAddr)

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = workDir
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	instance.mu.Lock()
	instance.Cmd = cmd
	instance.Process = cmd.Process
	instance.Status = "running"
	instance.LastSeen = time.Now()
	instance.mu.Unlock()

	logs.Infof("Worker process for assistant %s started with PID %d", spec.ID, cmd.Process.Pid)

	go ps.monitorProcess(instance)

	return nil
}

func (ps *ProcessScheduler) stopProcess(instance *ProcessInstance) error {
	instance.mu.RLock()
	defer instance.mu.RUnlock()

	if instance.Process == nil {
		return nil
	}

	if err := instance.Process.Signal(os.Interrupt); err != nil {
		logs.Warnf("Failed to send interrupt signal to %s: %v", instance.WorkerID, err)
		if err := instance.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
	}

	logs.Infof("Sent interrupt signal to worker %s", instance.WorkerID)
	return nil
}

func (ps *ProcessScheduler) healthCheck(instance *ProcessInstance) error {
	instance.mu.RLock()
	defer instance.mu.RUnlock()

	if instance.Process == nil {
		return fmt.Errorf("process not started")
	}

	process, err := os.FindProcess(instance.Process.Pid)
	if err != nil {
		return fmt.Errorf("process not found: %w", err)
	}

	if err := process.Signal(os.Interrupt); err != nil {
		return fmt.Errorf("process is not responding: %w", err)
	}

	instance.LastSeen = time.Now()
	return nil
}

func (ps *ProcessScheduler) monitorProcess(instance *ProcessInstance) {
	instance.mu.RLock()
	cmd := instance.Cmd
	instance.mu.RUnlock()

	if cmd == nil {
		return
	}

	err := cmd.Wait()

	instance.mu.Lock()
	defer instance.mu.Unlock()

	if err != nil {
		logs.Errorf("Worker process %s exited with error: %v", instance.WorkerID, err)
		instance.Status = "error"
	} else {
		logs.Infof("Worker process %s exited normally", instance.WorkerID)
		instance.Status = "stopped"
	}

	ps.removeInstance(instance.WorkerID)
}

func (ps *ProcessScheduler) removeInstance(workerID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	delete(ps.instances, workerID)
}
