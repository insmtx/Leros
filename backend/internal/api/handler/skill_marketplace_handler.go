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

// RegisterSkillMarketplaceRoutes 注册 Skill 市场相关路由。
func RegisterSkillMarketplaceRoutes(r gin.IRouter, service contract.SkillMarketplaceService) {
	r.GET("/skill-marketplace/search", searchSkillMarketplace(service))
	r.GET("/skill-marketplace/skills/:skill_id/download", downloadBuiltinSkill(service))
	r.POST("/skill-marketplace/install", installSkill(service))
}

func downloadBuiltinSkill(service contract.SkillMarketplaceService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		skillID := strings.TrimSpace(ctx.Param("skill_id"))
		if skillID == "" {
			ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, "skill_id is required"))
			return
		}

		download, err := service.DownloadBuiltinSkill(ctx, skillID)
		if err != nil {
			if err.Error() == "skill not found" {
				ctx.JSON(http.StatusNotFound, dto.Error(dto.CodeNotFound, err.Error()))
				return
			}
			ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
			return
		}
		defer download.Reader.Close()

		ctx.Header("Content-Type", "application/zip")
		ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, download.FileName))
		ctx.Status(http.StatusOK)
		if _, err := io.Copy(ctx.Writer, download.Reader); err != nil {
			ctx.Error(err)
		}
	}
}

func searchSkillMarketplace(service contract.SkillMarketplaceService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req contract.SearchSkillMarketplaceRequest
		if err := ctx.ShouldBindQuery(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
			return
		}

		result, err := service.SearchSkillMarketplace(ctx, &req)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
			return
		}

		ctx.JSON(http.StatusOK, dto.Success(result))
	}
}

func installSkill(service contract.SkillMarketplaceService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req contract.InstallSkillRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
			return
		}

		if strings.TrimSpace(req.Source) == "" {
			ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, "source is required"))
			return
		}
		if strings.TrimSpace(req.SkillID) == "" {
			ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, "skill_id is required"))
			return
		}

		result, err := service.InstallSkill(ctx, &req)
		if err != nil {
			if err.Error() == "user not authenticated or org not set" {
				ctx.JSON(http.StatusUnauthorized, dto.Error(dto.CodeInternalError, err.Error()))
				return
			}
			ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
			return
		}

		ctx.JSON(http.StatusOK, dto.Success(result))
	}
}
