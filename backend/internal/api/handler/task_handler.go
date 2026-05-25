package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/insmtx/Leros/backend/internal/api/contract"
	"github.com/insmtx/Leros/backend/internal/api/dto"
)

type TaskHandler struct {
	service contract.TaskService
}

func NewTaskHandler(service contract.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

// ================================================================
// Route Registration
// ================================================================

func (h *TaskHandler) RegisterRoutes(r gin.IRouter) {
	r.POST("/CreateTask", h.CreateTask)
	r.POST("/GetTask", h.GetTask)
	r.POST("/UpdateTask", h.UpdateTask)
	r.POST("/DeleteTask", h.DeleteTask)
	r.POST("/ListTasks", h.ListTasks)
}

func RegisterTaskRoutes(r gin.IRouter, service contract.TaskService) {
	h := NewTaskHandler(service)
	h.RegisterRoutes(r)
}

// ================================================================
// Handler Methods
// ================================================================

// @Summary 创建任务
// @Description 创建一个新任务
// @Tags Task
// @Accept json
// @Produce json
// @Param body body contract.CreateTaskRequest true "创建任务请求"
// @Success 200 {object} dto.Response "成功响应"
// @Failure 400 {object} dto.ErrorResponse "请求参数错误"
// @Failure 401 {object} dto.ErrorResponse "未认证"
// @Failure 500 {object} dto.ErrorResponse "内部服务器错误"
// @Router /CreateTask [post]
func (h *TaskHandler) CreateTask(ctx *gin.Context) {
	var req contract.CreateTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, "title is required"))
		return
	}

	result, err := h.service.CreateTask(ctx, &req)
	if err != nil {
		handleTaskServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.Success(result))
}

type GetTaskRequest struct {
	PublicID *string `json:"public_id,omitempty"`
}

// @Summary 获取任务详情
// @Description 根据PublicId获取任务详情
// @Tags Task
// @Accept json
// @Produce json
// @Param body body GetTaskRequest true "获取任务请求"
// @Success 200 {object} dto.Response "成功响应"
// @Failure 400 {object} dto.ErrorResponse "请求参数错误"
// @Failure 401 {object} dto.ErrorResponse "未认证"
// @Failure 404 {object} dto.ErrorResponse "资源不存在"
// @Failure 500 {object} dto.ErrorResponse "内部服务器错误"
// @Router /GetTask [post]
func (h *TaskHandler) GetTask(ctx *gin.Context) {
	var req GetTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}
	if req.PublicID == nil || *req.PublicID == "" {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, "public_id is required"))
		return
	}

	result, err := h.service.GetTask(ctx, *req.PublicID)
	if err != nil {
		handleTaskServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.Success(result))
}

type UpdateTaskRequest struct {
	PublicID string `json:"public_id" binding:"required"`
	contract.UpdateTaskRequest
}

// @Summary 更新任务
// @Description 更新任务信息
// @Tags Task
// @Accept json
// @Produce json
// @Param body body UpdateTaskRequest true "更新任务请求"
// @Success 200 {object} dto.Response "成功响应"
// @Failure 400 {object} dto.ErrorResponse "请求参数错误"
// @Failure 401 {object} dto.ErrorResponse "未认证"
// @Failure 404 {object} dto.ErrorResponse "资源不存在"
// @Failure 500 {object} dto.ErrorResponse "内部服务器错误"
// @Router /UpdateTask [post]
func (h *TaskHandler) UpdateTask(ctx *gin.Context) {
	var req UpdateTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	result, err := h.service.UpdateTask(ctx, req.PublicID, &req.UpdateTaskRequest)
	if err != nil {
		handleTaskServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.Success(result))
}

type DeleteTaskRequest struct {
	PublicID string `json:"public_id" binding:"required"`
}

// @Summary 删除任务
// @Description 根据PublicId删除任务（软删除）
// @Tags Task
// @Accept json
// @Produce json
// @Param body body DeleteTaskRequest true "删除任务请求"
// @Success 200 {object} dto.Response "成功响应"
// @Failure 400 {object} dto.ErrorResponse "请求参数错误"
// @Failure 401 {object} dto.ErrorResponse "未认证"
// @Failure 404 {object} dto.ErrorResponse "资源不存在"
// @Failure 500 {object} dto.ErrorResponse "内部服务器错误"
// @Router /DeleteTask [post]
func (h *TaskHandler) DeleteTask(ctx *gin.Context) {
	var req DeleteTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	if err := h.service.DeleteTask(ctx, req.PublicID); err != nil {
		handleTaskServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.Success(nil))
}

// @Summary 查询任务列表
// @Description 分页查询任务列表
// @Tags Task
// @Accept json
// @Produce json
// @Param body body contract.ListTasksRequest true "查询列表请求"
// @Success 200 {object} dto.Response "成功响应"
// @Failure 400 {object} dto.ErrorResponse "请求参数错误"
// @Failure 401 {object} dto.ErrorResponse "未认证"
// @Failure 500 {object} dto.ErrorResponse "内部服务器错误"
// @Router /ListTasks [post]
func (h *TaskHandler) ListTasks(ctx *gin.Context) {
	var req contract.ListTasksRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	req.Fill()

	result, err := h.service.ListTasks(ctx, &req)
	if err != nil {
		handleTaskServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.Success(result))
}

// ================================================================
// Error Handling
// ================================================================

func handleTaskServiceError(ctx *gin.Context, err error) {
	errMsg := err.Error()

	switch errMsg {
	case "user not authenticated or org not set":
		ctx.JSON(http.StatusUnauthorized, dto.Error(dto.CodeInternalError, errMsg))
		return
	}

	switch errMsg {
	case "task not found",
		"project not found":
		ctx.JSON(http.StatusNotFound, dto.Error(dto.CodeNotFound, errMsg))
	case "title is required",
		"title cannot be empty",
		"public_id is required",
		"project_id is required":
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, errMsg))
	default:
		ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, errMsg))
	}
}
