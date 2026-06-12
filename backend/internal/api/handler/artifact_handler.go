package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/insmtx/Leros/backend/internal/api/contract"
	"github.com/insmtx/Leros/backend/internal/api/dto"
)

type ArtifactHandler struct {
	service contract.ArtifactService
}

func NewArtifactHandler(service contract.ArtifactService) *ArtifactHandler {
	return &ArtifactHandler{service: service}
}

func (h *ArtifactHandler) RegisterRoutes(r gin.IRouter) {
	r.POST("/ListTaskArtifacts", h.ListTaskArtifacts)
	r.POST("/GetArtifact", h.GetArtifact)
	r.GET("/artifacts/:artifact_id/download", h.DownloadArtifact)
}

func RegisterArtifactRoutes(r gin.IRouter, service contract.ArtifactService) {
	h := NewArtifactHandler(service)
	h.RegisterRoutes(r)
}

type ListTaskArtifactsRequest struct {
	TaskID string `json:"task_id"`
}

type GetArtifactRequest struct {
	ArtifactID string `json:"artifact_id"`
}

func (h *ArtifactHandler) ListTaskArtifacts(ctx *gin.Context) {
	var req ListTaskArtifactsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}
	taskID := strings.TrimSpace(req.TaskID)
	if taskID == "" {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, "task_id is required"))
		return
	}
	result, err := h.service.ListTaskArtifacts(ctx, taskID)
	if err != nil {
		handleArtifactServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.Success(result))
}

func (h *ArtifactHandler) GetArtifact(ctx *gin.Context) {
	var req GetArtifactRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}
	artifactID := strings.TrimSpace(req.ArtifactID)
	if artifactID == "" {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, "artifact_id is required"))
		return
	}
	result, err := h.service.GetArtifact(ctx, artifactID)
	if err != nil {
		handleArtifactServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.Success(result))
}

func (h *ArtifactHandler) DownloadArtifact(ctx *gin.Context) {
	artifactID := strings.TrimSpace(ctx.Param("artifact_id"))
	download, err := h.service.GetArtifactDownload(ctx, artifactID)
	if err != nil {
		handleArtifactServiceError(ctx, err)
		return
	}
	defer download.Reader.Close()

	mimeType := download.MimeType
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	ctx.Header("Content-Type", mimeType)
	ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, sanitizeDownloadName(download.FileName)))
	if download.Size > 0 {
		ctx.Header("Content-Length", fmt.Sprintf("%d", download.Size))
	}
	ctx.Status(http.StatusOK)
	if _, err := io.Copy(ctx.Writer, download.Reader); err != nil {
		ctx.Error(err)
	}
}

func sanitizeDownloadName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "artifact"
	}
	name = strings.ReplaceAll(name, `"`, "")
	name = strings.ReplaceAll(name, "\r", "")
	name = strings.ReplaceAll(name, "\n", "")
	return name
}

func handleArtifactServiceError(ctx *gin.Context, err error) {
	errMsg := err.Error()
	switch errMsg {
	case "user not authenticated or org not set":
		ctx.JSON(http.StatusUnauthorized, dto.Error(dto.CodeInternalError, errMsg))
	case "task_id is required", "artifact_id is required":
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, errMsg))
	case "task not found", "artifact not found":
		ctx.JSON(http.StatusNotFound, dto.Error(dto.CodeNotFound, errMsg))
	default:
		ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, errMsg))
	}
}
