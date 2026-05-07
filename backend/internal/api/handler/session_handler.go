package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/insmtx/SingerOS/backend/internal/api/contract"
	"github.com/insmtx/SingerOS/backend/internal/api/dto"
)

type SessionHandler struct {
	service contract.SessionService
}

func NewSessionHandler(service contract.SessionService) *SessionHandler {
	return &SessionHandler{
		service: service,
	}
}

func (h *SessionHandler) RegisterRoutes(r gin.IRouter) {
	r.POST("/CreateSession", h.CreateSession)
	r.POST("/GetSession", h.GetSession)
	r.POST("/UpdateSession", h.UpdateSession)
	r.POST("/DeleteSession", h.DeleteSession)
	r.POST("/ListSessions", h.ListSessions)
	r.POST("/ActivateSession", h.ActivateSession)
	r.POST("/PauseSession", h.PauseSession)
	r.POST("/EndSession", h.EndSession)
	r.POST("/ResumeSession", h.ResumeSession)
	r.POST("/AddMessage", h.AddMessage)
	r.POST("/GetSessionMessages", h.GetSessionMessages)
	r.POST("/DeleteMessage", h.DeleteMessage)
	r.POST("/ClearSessionMessages", h.ClearSessionMessages)
}

func RegisterSessionRoutes(r gin.IRouter, service contract.SessionService) {
	h := NewSessionHandler(service)
	h.RegisterRoutes(r)
}

type CreateSessionRequest struct {
	contract.CreateSessionRequest
}

func (h *SessionHandler) CreateSession(ctx *gin.Context) {
	var req contract.CreateSessionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	result, err := h.service.CreateSession(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, dto.Success(result))
}

type GetSessionRequest struct {
	ID        *uint   `json:"id,omitempty"`
	SessionID *string `json:"session_id,omitempty"`
}

func (h *SessionHandler) GetSession(ctx *gin.Context) {
	var req GetSessionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	var id uint
	var sessionID string

	if req.ID != nil {
		id = *req.ID
	}
	if req.SessionID != nil {
		sessionID = *req.SessionID
	}

	result, err := h.service.GetSession(ctx, id, sessionID)
	if err != nil {
		if err.Error() == "session not found" {
			ctx.JSON(http.StatusNotFound, dto.Error(dto.CodeNotFound, err.Error()))
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, dto.Success(result))
}

type UpdateSessionRequest struct {
	ID uint `json:"id" binding:"required"`
	contract.UpdateSessionRequest
}

func (h *SessionHandler) UpdateSession(ctx *gin.Context) {
	var req UpdateSessionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	result, err := h.service.UpdateSession(ctx, req.ID, &req.UpdateSessionRequest)
	if err != nil {
		if err.Error() == "session not found" {
			ctx.JSON(http.StatusNotFound, dto.Error(dto.CodeNotFound, err.Error()))
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, dto.Success(result))
}

type DeleteSessionRequest struct {
	ID uint `json:"id" binding:"required"`
}

func (h *SessionHandler) DeleteSession(ctx *gin.Context) {
	var req DeleteSessionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	err := h.service.DeleteSession(ctx, req.ID)
	if err != nil {
		if err.Error() == "session not found" {
			ctx.JSON(http.StatusNotFound, dto.Error(dto.CodeNotFound, err.Error()))
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, dto.Success(nil))
}

func (h *SessionHandler) ListSessions(ctx *gin.Context) {
	var req contract.ListSessionsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	result, err := h.service.ListSessions(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, dto.Success(result))
}

type SessionIDRequest struct {
	ID uint `json:"id" binding:"required"`
}

func (h *SessionHandler) ActivateSession(ctx *gin.Context) {
	var req SessionIDRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	err := h.service.ActivateSession(ctx, req.ID)
	if err != nil {
		if err.Error() == "session not found" {
			ctx.JSON(http.StatusNotFound, dto.Error(dto.CodeNotFound, err.Error()))
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, dto.Success(nil))
}

func (h *SessionHandler) PauseSession(ctx *gin.Context) {
	var req SessionIDRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	err := h.service.PauseSession(ctx, req.ID)
	if err != nil {
		if err.Error() == "session not found" {
			ctx.JSON(http.StatusNotFound, dto.Error(dto.CodeNotFound, err.Error()))
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, dto.Success(nil))
}

func (h *SessionHandler) EndSession(ctx *gin.Context) {
	var req SessionIDRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	err := h.service.EndSession(ctx, req.ID)
	if err != nil {
		if err.Error() == "session not found" {
			ctx.JSON(http.StatusNotFound, dto.Error(dto.CodeNotFound, err.Error()))
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, dto.Success(nil))
}

func (h *SessionHandler) ResumeSession(ctx *gin.Context) {
	var req SessionIDRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	err := h.service.ResumeSession(ctx, req.ID)
	if err != nil {
		if err.Error() == "session not found" {
			ctx.JSON(http.StatusNotFound, dto.Error(dto.CodeNotFound, err.Error()))
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, dto.Success(nil))
}

type AddMessageRequest struct {
	SessionID uint `json:"session_id" binding:"required"`
	contract.AddMessageRequest
}

func (h *SessionHandler) AddMessage(ctx *gin.Context) {
	var req AddMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	result, err := h.service.AddMessage(ctx, req.SessionID, &req.AddMessageRequest)
	if err != nil {
		if err.Error() == "session not found" {
			ctx.JSON(http.StatusNotFound, dto.Error(dto.CodeNotFound, err.Error()))
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, dto.Success(result))
}

type GetSessionMessagesRequest struct {
	SessionID uint `json:"session_id" binding:"required"`
	Page      int  `json:"page,omitempty"`
	PerPage   int  `json:"per_page,omitempty"`
}

func (h *SessionHandler) GetSessionMessages(ctx *gin.Context) {
	var req GetSessionMessagesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	page := req.Page
	if page == 0 {
		page = 1
	}
	perPage := req.PerPage
	if perPage == 0 {
		perPage = 20
	}

	result, err := h.service.GetSessionMessages(ctx, req.SessionID, page, perPage)
	if err != nil {
		if err.Error() == "session not found" {
			ctx.JSON(http.StatusNotFound, dto.Error(dto.CodeNotFound, err.Error()))
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, dto.Success(result))
}

type DeleteMessageRequest struct {
	MessageID uint `json:"message_id" binding:"required"`
}

func (h *SessionHandler) DeleteMessage(ctx *gin.Context) {
	var req DeleteMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	err := h.service.DeleteMessage(ctx, req.MessageID)
	if err != nil {
		if err.Error() == "message not found" {
			ctx.JSON(http.StatusNotFound, dto.Error(dto.CodeNotFound, err.Error()))
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, dto.Success(nil))
}

type ClearSessionMessagesRequest struct {
	SessionID uint `json:"session_id" binding:"required"`
}

func (h *SessionHandler) ClearSessionMessages(ctx *gin.Context) {
	var req ClearSessionMessagesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	err := h.service.ClearSessionMessages(ctx, req.SessionID)
	if err != nil {
		if err.Error() == "session not found" {
			ctx.JSON(http.StatusNotFound, dto.Error(dto.CodeNotFound, err.Error()))
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, dto.Success(nil))
}
