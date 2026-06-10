package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/insmtx/Leros/backend/internal/api/auth"
	"github.com/insmtx/Leros/backend/internal/api/contract"
	"github.com/insmtx/Leros/backend/internal/api/dto"
)

type FileHandler struct {
	service contract.FileService
}

func NewFileHandler(service contract.FileService) *FileHandler {
	return &FileHandler{service: service}
}

func (h *FileHandler) RegisterRoutes(r gin.IRouter) {
	r.POST("/files/upload", h.UploadFile)
	r.GET("/files/:id/download", h.DownloadFile)
}

func (h *FileHandler) UploadFile(ctx *gin.Context) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, "file is required"))
		return
	}

	purpose := strings.TrimSpace(ctx.PostForm("purpose"))
	if purpose == "" {
		purpose = "attachment"
	}

	file, err := fileHeader.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, "failed to open file"))
		return
	}
	defer file.Close()

	caller, _ := auth.FromGinContext(ctx)
	if caller == nil || caller.OrgID == 0 {
		ctx.JSON(http.StatusUnauthorized, dto.Error(dto.CodeInternalError, "not authenticated"))
		return
	}

	result, err := h.service.UploadFile(ctx, &contract.UploadFileRequest{
		OrgID:    caller.OrgID,
		OwnerID:  caller.Uin,
		File:     file,
		Filename: fileHeader.Filename,
		FileSize: fileHeader.Size,
		MimeType: fileHeader.Header.Get("Content-Type"),
		Purpose:  purpose,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, dto.Success(result))
}

func (h *FileHandler) DownloadFile(ctx *gin.Context) {
	fileID := strings.TrimSpace(ctx.Param("id"))
	if fileID == "" {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, "file id is required"))
		return
	}

	caller, _ := auth.FromGinContext(ctx)
	if caller == nil || caller.OrgID == 0 {
		ctx.JSON(http.StatusUnauthorized, dto.Error(dto.CodeInternalError, "not authenticated"))
		return
	}

	downloadURL, err := h.service.GetFileDownloadURL(ctx, caller.OrgID, fileID)
	if err != nil {
		if err.Error() == "file not found" {
			ctx.JSON(http.StatusNotFound, dto.Error(dto.CodeNotFound, "file not found"))
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}

	ctx.Redirect(http.StatusFound, downloadURL.URL)
}
