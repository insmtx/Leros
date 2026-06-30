package service

import (
	"testing"
	"time"

	"github.com/insmtx/Leros/backend/internal/api/contract"
	dbpkg "github.com/insmtx/Leros/backend/internal/infra/db"
	"github.com/insmtx/Leros/backend/types"
)

func TestCreateTask_TouchesProjectUpdatedAt(t *testing.T) {
	database := setupTestDB(t)
	service := NewTaskService(database)
	ctx := setupTestContextWithCaller(t)

	project := &types.Project{
		PublicID: "prj_test_create_task_touch",
		OrgID:    1,
		OwnerID:  1,
		Name:     "Task Project",
		Status:   string(types.ProjectStatusActive),
	}
	if err := dbpkg.CreateProject(ctx, database, project); err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	oldUpdatedAt := time.Now().Add(-time.Hour).UTC()
	if err := database.Model(&types.Project{}).
		Where("id = ?", project.ID).
		Update("updated_at", oldUpdatedAt).Error; err != nil {
		t.Fatalf("set old project updated_at: %v", err)
	}

	_, err := service.CreateTask(ctx, &contract.CreateTaskRequest{
		ProjectID: project.PublicID,
		Title:     "新建任务",
	})
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	refreshedProject, err := dbpkg.GetProjectByID(ctx, database, project.ID)
	if err != nil {
		t.Fatalf("GetProjectByID failed: %v", err)
	}
	if refreshedProject == nil {
		t.Fatal("expected project to exist after CreateTask")
	}
	if !refreshedProject.UpdatedAt.After(oldUpdatedAt) {
		t.Fatalf("expected project updated_at after %v, got %v", oldUpdatedAt, refreshedProject.UpdatedAt)
	}
}
