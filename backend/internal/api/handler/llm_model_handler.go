package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/insmtx/Leros/backend/internal/api/contract"
	"github.com/insmtx/Leros/backend/internal/api/dto"
)

type LLMModelHandler struct {
	service contract.LLMModelService
}

func NewLLMModelHandler(service contract.LLMModelService) *LLMModelHandler {
	return &LLMModelHandler{service: service}
}

func (h *LLMModelHandler) RegisterRoutes(r gin.IRouter) {
	r.POST("/CreateLLMModel", h.CreateLLMModel)
	r.POST("/GetLLMModel", h.GetLLMModel)
	r.POST("/GetDefaultLLMModel", h.GetDefaultLLMModel)
	r.POST("/UpdateLLMModel", h.UpdateLLMModel)
	r.POST("/DeleteLLMModel", h.DeleteLLMModel)
	r.POST("/ListLLMModels", h.ListLLMModels)
	r.POST("/TestLLMModel", h.TestLLMModel)
}

func RegisterLLMModelRoutes(r gin.IRouter, service contract.LLMModelService) {
	h := NewLLMModelHandler(service)
	h.RegisterRoutes(r)
}

func (h *LLMModelHandler) CreateLLMModel(ctx *gin.Context) {
	var req contract.CreateLLMModelRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	result, err := h.service.CreateLLMModel(ctx, &req)
	if err != nil {
		handleLLMModelServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.Success(result))
}

type GetLLMModelRequest struct {
	ID   *uint   `json:"id,omitempty"`
	Code *string `json:"code,omitempty"`
}

func (h *LLMModelHandler) GetLLMModel(ctx *gin.Context) {
	var req GetLLMModelRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}
	if req.ID == nil && req.Code == nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, "id or code is required"))
		return
	}

	var id uint
	var code string
	if req.ID != nil {
		id = *req.ID
	}
	if req.Code != nil {
		code = *req.Code
	}

	result, err := h.service.GetLLMModel(ctx, id, code)
	if err != nil {
		handleLLMModelServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.Success(result))
}

func (h *LLMModelHandler) GetDefaultLLMModel(ctx *gin.Context) {
	result, err := h.service.GetDefaultLLMModel(ctx)
	if err != nil {
		handleLLMModelServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.Success(result))
}

type UpdateLLMModelRequest struct {
	ID uint `json:"id" binding:"required"`
	contract.UpdateLLMModelRequest
}

func (h *LLMModelHandler) UpdateLLMModel(ctx *gin.Context) {
	var req UpdateLLMModelRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	result, err := h.service.UpdateLLMModel(ctx, req.ID, &req.UpdateLLMModelRequest)
	if err != nil {
		handleLLMModelServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.Success(result))
}

type DeleteLLMModelRequest struct {
	ID uint `json:"id" binding:"required"`
}

func (h *LLMModelHandler) DeleteLLMModel(ctx *gin.Context) {
	var req DeleteLLMModelRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	if err := h.service.DeleteLLMModel(ctx, req.ID); err != nil {
		handleLLMModelServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.Success(nil))
}

func (h *LLMModelHandler) ListLLMModels(ctx *gin.Context) {
	var req contract.ListLLMModelsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	result, err := h.service.ListLLMModels(ctx, &req)
	if err != nil {
		handleLLMModelServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.Success(result))
}

func (h *LLMModelHandler) TestLLMModel(ctx *gin.Context) {
	var req contract.TestLLMModelRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	result, err := h.service.TestLLMModel(ctx, &req)
	if err != nil {
		handleLLMModelServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.Success(result))
}

func handleLLMModelServiceError(ctx *gin.Context, err error) {
	if err.Error() == "user not authenticated or org not set" {
		ctx.JSON(http.StatusUnauthorized, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}
	if err.Error() == "permission denied" {
		ctx.JSON(http.StatusForbidden, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}
	if err.Error() == "llm model not found" {
		ctx.JSON(http.StatusNotFound, dto.Error(dto.CodeNotFound, err.Error()))
		return
	}
	if err.Error() == "id or code is required" ||
		err.Error() == "provider is required" ||
		err.Error() == "model is required" ||
		err.Error() == "base_url is required" ||
		err.Error() == "api_key is required" ||
		err.Error() == "llm model with this code already exists" {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}
	ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
}
