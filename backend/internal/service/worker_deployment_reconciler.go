package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/insmtx/Leros/backend/config"
	"github.com/insmtx/Leros/backend/internal/api/auth"
	"github.com/insmtx/Leros/backend/internal/api/contract"
	"github.com/insmtx/Leros/backend/internal/infra/db"
	"github.com/insmtx/Leros/backend/internal/worker"
	"github.com/insmtx/Leros/backend/types"
	"github.com/ygpkg/yg-go/logs"
)

const (
	defaultWorkerDeploymentReconcileInterval = 10 * time.Second
	defaultWorkerDeploymentProvisionTimeout  = 5 * time.Minute
)

// StartWorkerDeploymentReconciler keeps WorkerDeployment records aligned with runtime workers.
func StartWorkerDeploymentReconciler(
	ctx context.Context,
	database *gorm.DB,
	workerScheduler worker.WorkerScheduler,
	schedulerConfig *config.SchedulerConfig,
) {
	if database == nil || workerScheduler == nil {
		return
	}

	ticker := time.NewTicker(defaultWorkerDeploymentReconcileInterval)
	defer ticker.Stop()

	reconcileWorkerDeployments(ctx, database, workerScheduler, schedulerConfig)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			reconcileWorkerDeployments(ctx, database, workerScheduler, schedulerConfig)
		}
	}
}

func reconcileWorkerDeployments(
	ctx context.Context,
	database *gorm.DB,
	workerScheduler worker.WorkerScheduler,
	schedulerConfig *config.SchedulerConfig,
) {
	statuses := []string{
		string(types.WorkerDeploymentStatusPending),
		string(types.WorkerDeploymentStatusProvisioning),
		string(types.WorkerDeploymentStatusFailed),
		string(types.WorkerDeploymentStatusReady),
	}
	deployments, err := db.ListWorkerDeploymentsByStatuses(ctx, database, statuses)
	if err != nil {
		logs.WarnContextf(ctx, "list worker deployments failed: %v", err)
		return
	}
	for _, deployment := range deployments {
		if err := reconcileWorkerDeployment(ctx, database, workerScheduler, schedulerConfig, deployment); err != nil {
			logs.WarnContextf(ctx, "reconcile worker deployment %s failed: %v", deployment.DeploymentName, err)
		}
	}
}

func reconcileWorkerDeployment(
	ctx context.Context,
	database *gorm.DB,
	workerScheduler worker.WorkerScheduler,
	schedulerConfig *config.SchedulerConfig,
	deployment *types.WorkerDeployment,
) error {
	if deployment == nil {
		return nil
	}
	assistant, err := db.GetDigitalAssistantByID(ctx, database, deployment.DigitalAssistantID)
	if err != nil {
		return err
	}
	if assistant == nil || assistant.Status != string(contract.DigitalAssistantStatusActive) {
		if err := workerScheduler.Stop(ctx, deployment.DeploymentName); err != nil {
			return fmt.Errorf("stop inactive worker: %w", err)
		}
		return db.MarkWorkerDeploymentStatus(ctx, database, deployment.ID, string(types.WorkerDeploymentStatusStopped), "")
	}

	if deployment.Status == string(types.WorkerDeploymentStatusReady) {
		if err := workerScheduler.Health(ctx, deployment.DeploymentName); err != nil {
			return db.MarkWorkerDeploymentStatus(ctx, database, deployment.ID, string(types.WorkerDeploymentStatusFailed), err.Error())
		}
		return db.MarkWorkerDeploymentStatus(ctx, database, deployment.ID, string(types.WorkerDeploymentStatusReady), "")
	}

	if deployment.Status == string(types.WorkerDeploymentStatusProvisioning) {
		return reconcileProvisioningWorkerDeployment(ctx, database, workerScheduler, deployment)
	}

	return startWorkerDeployment(ctx, database, workerScheduler, schedulerConfig, deployment, assistant)
}

func reconcileProvisioningWorkerDeployment(
	ctx context.Context,
	database *gorm.DB,
	workerScheduler worker.WorkerScheduler,
	deployment *types.WorkerDeployment,
) error {
	if err := workerScheduler.Health(ctx, deployment.DeploymentName); err == nil {
		return db.MarkWorkerDeploymentStatus(ctx, database, deployment.ID, string(types.WorkerDeploymentStatusReady), "")
	} else {
		startedAt := deployment.LastStartedAt
		if startedAt == nil {
			return db.MarkWorkerDeploymentStatus(ctx, database, deployment.ID, string(types.WorkerDeploymentStatusFailed), err.Error())
		}
		if time.Since(*startedAt) > defaultWorkerDeploymentProvisionTimeout {
			return db.MarkWorkerDeploymentStatus(ctx, database, deployment.ID, string(types.WorkerDeploymentStatusFailed), err.Error())
		}
		return db.MarkWorkerDeploymentStatus(ctx, database, deployment.ID, string(types.WorkerDeploymentStatusProvisioning), err.Error())
	}
}

func startWorkerDeployment(
	ctx context.Context,
	database *gorm.DB,
	workerScheduler worker.WorkerScheduler,
	schedulerConfig *config.SchedulerConfig,
	deployment *types.WorkerDeployment,
	assistant *types.DigitalAssistant,
) error {
	bootstrapToken, err := auth.GenerateBootstrapToken()
	if err != nil {
		return err
	}
	if err := db.MarkWorkerDeploymentStarted(ctx, database, deployment.ID, auth.HashBootstrapToken(bootstrapToken)); err != nil {
		return err
	}

	spec := &worker.WorkerSpec{
		ID:             deployment.DeploymentName,
		OrgID:          deployment.OrgID,
		WorkerID:       deployment.WorkerID,
		Name:           assistant.Name,
		BootstrapToken: bootstrapToken,
		ServerAddr:     schedulerServerAddr(schedulerConfig),
		EnvType:        worker.WorkerEnvProcess,
	}
	if _, err := workerScheduler.Start(ctx, spec); err != nil {
		_ = db.MarkWorkerDeploymentStatus(ctx, database, deployment.ID, string(types.WorkerDeploymentStatusFailed), err.Error())
		return err
	}
	if err := workerScheduler.Health(ctx, deployment.DeploymentName); err != nil {
		return db.MarkWorkerDeploymentStatus(ctx, database, deployment.ID, string(types.WorkerDeploymentStatusProvisioning), err.Error())
	}
	return db.MarkWorkerDeploymentStatus(ctx, database, deployment.ID, string(types.WorkerDeploymentStatusReady), "")
}

func schedulerServerAddr(schedulerConfig *config.SchedulerConfig) string {
	if schedulerConfig == nil {
		return ""
	}
	return strings.TrimSpace(schedulerConfig.ServerAddr)
}
