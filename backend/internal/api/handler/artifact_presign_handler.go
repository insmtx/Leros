package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/insmtx/Leros/backend/internal/api/contract"
	"github.com/insmtx/Leros/backend/internal/api/dto"
	"github.com/ygpkg/yg-go/logs"
)

// PresignArtifactUpload handles Worker artifact presigned upload URL requests.
func (h *ProjectFileHandler) PresignArtifactUpload(ctx *gin.Context) {
	var req contract.PresignArtifactUploadRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	resp, err := h.service.PresignArtifactUpload(ctx.Request.Context(), &req)
	if err != nil {
		logs.ErrorContextf(ctx, "presign artifact upload failed: %v", err)
		ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, dto.Success(resp))
}
