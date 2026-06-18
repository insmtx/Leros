package handler

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/insmtx/Leros/backend/config"
	"github.com/insmtx/Leros/backend/internal/api/auth"
	"github.com/insmtx/Leros/backend/internal/api/dto"
	"github.com/insmtx/Leros/backend/internal/infra/db"
)

const headerWorkerBootstrapToken = "X-Worker-Bootstrap-Token"

const defaultWorkerTokenTTL = 24 * time.Hour

// WorkerAuthHandler signs short-lived worker access tokens from bootstrap tokens.
type WorkerAuthHandler struct {
	cfg       *config.WorkerAuthConfig
	jwtSecret string
	db        *gorm.DB
}

type issueWorkerTokenRequest struct {
	OrgID          uint   `json:"org_id" binding:"required"`
	WorkerID       uint   `json:"worker_id" binding:"required"`
	BootstrapToken string `json:"bootstrap_token,omitempty"`
}

type issueWorkerTokenResponse struct {
	AuthToken string `json:"auth_token"`
	ExpiredAt int64  `json:"expired_at"`
	TokenType string `json:"token_type"`
}

// NewWorkerAuthHandler creates a worker auth handler.
func NewWorkerAuthHandler(cfg *config.WorkerAuthConfig, jwtSecret string, database *gorm.DB) *WorkerAuthHandler {
	return &WorkerAuthHandler{
		cfg:       cfg,
		jwtSecret: strings.TrimSpace(jwtSecret),
		db:        database,
	}
}

// RegisterWorkerAuthRoutes registers worker auth routes.
func RegisterWorkerAuthRoutes(r gin.IRouter, cfg *config.WorkerAuthConfig, jwtSecret string, database *gorm.DB) {
	h := NewWorkerAuthHandler(cfg, jwtSecret, database)
	r.POST("/workers/token", h.IssueToken)
}

// IssueToken exchanges a worker bootstrap token for a short-lived worker access token.
func (h *WorkerAuthHandler) IssueToken(ctx *gin.Context) {
	var req issueWorkerTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	bootstrapToken := strings.TrimSpace(ctx.GetHeader(headerWorkerBootstrapToken))
	if bootstrapToken == "" {
		bootstrapToken = strings.TrimSpace(req.BootstrapToken)
	}
	if bootstrapToken == "" {
		ctx.JSON(http.StatusUnauthorized, dto.Error(dto.CodeInternalError, "worker bootstrap token is required"))
		return
	}

	if err := h.validateBootstrapToken(ctx, req.OrgID, req.WorkerID, bootstrapToken); err != nil {
		ctx.JSON(http.StatusUnauthorized, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}
	if err := h.validateWorker(ctx, req.OrgID, req.WorkerID); err != nil {
		ctx.JSON(http.StatusForbidden, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}

	token, expiredAt, err := auth.GenerateWorkerToken(req.OrgID, req.WorkerID, h.jwtSecret, h.tokenTTL())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, dto.Success(issueWorkerTokenResponse{
		AuthToken: token,
		ExpiredAt: expiredAt,
		TokenType: "Bearer",
	}))
}

func (h *WorkerAuthHandler) validateBootstrapToken(ctx *gin.Context, orgID, workerID uint, token string) error {
	if h.cfg != nil {
		for _, item := range h.cfg.BootstrapTokens {
			if item.OrgID != orgID || item.WorkerID != workerID {
				continue
			}
			expected := strings.TrimSpace(item.Token)
			if expected != "" && subtle.ConstantTimeCompare([]byte(token), []byte(expected)) == 1 {
				return nil
			}
		}
	}

	if h.db != nil {
		deployment, err := db.GetWorkerDeploymentByOrgWorkerID(ctx.Request.Context(), h.db, orgID, workerID)
		if err != nil {
			return err
		}
		if deployment != nil && deployment.BootstrapTokenHash != "" {
			got := auth.HashBootstrapToken(token)
			if subtle.ConstantTimeCompare([]byte(got), []byte(deployment.BootstrapTokenHash)) == 1 {
				return nil
			}
		}
	}

	return errors.New("invalid worker bootstrap token")
}

func (h *WorkerAuthHandler) validateWorker(ctx *gin.Context, orgID, workerID uint) error {
	if h.db == nil {
		return nil
	}
	deployment, err := db.GetWorkerDeploymentByOrgWorkerID(ctx.Request.Context(), h.db, orgID, workerID)
	if err != nil {
		return err
	}
	assistantID := workerID
	if deployment != nil {
		if deployment.OrgID != orgID {
			return errors.New("worker organization mismatch")
		}
		assistantID = deployment.DigitalAssistantID
	}
	assistant, err := db.GetDigitalAssistantByID(ctx.Request.Context(), h.db, assistantID)
	if err != nil {
		return err
	}
	if assistant == nil {
		return errors.New("worker not found")
	}
	if assistant.OrgID != orgID {
		return errors.New("worker organization mismatch")
	}
	if assistant.Status != "active" {
		return errors.New("worker is not active")
	}
	return nil
}

func (h *WorkerAuthHandler) tokenTTL() time.Duration {
	if h.cfg != nil && h.cfg.TokenTTLSeconds > 0 {
		return time.Duration(h.cfg.TokenTTLSeconds) * time.Second
	}
	return defaultWorkerTokenTTL
}
