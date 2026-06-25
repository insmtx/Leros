package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/insmtx/Leros/backend/internal/api/contract"
	"github.com/insmtx/Leros/backend/internal/api/dto"
	"github.com/insmtx/Leros/backend/types"
)

// RegisterSkillRoutes registers skill management routes.
func RegisterSkillRoutes(r gin.IRouter, service contract.SkillService) {
	r.PATCH("/skills/:code/status", toggleSkillStatus(service))
	r.GET("/skills/recent", listRecentUsedSkills(service))
	r.GET("/skills/statuses", getSkillStatuses(service))
}

func toggleSkillStatus(service contract.SkillService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		code := strings.TrimSpace(ctx.Param("code"))
		if code == "" {
			ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, "code is required"))
			return
		}

		var req contract.ToggleSkillStatusRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
			return
		}

		result, err := service.ToggleSkillStatus(ctx, code, &req)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				ctx.JSON(http.StatusNotFound, dto.Error(dto.CodeNotFound, err.Error()))
				return
			}
			if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "required") {
				ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
				return
			}
			ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
			return
		}

		ctx.JSON(http.StatusOK, dto.Success(result))
	}
}

func listRecentUsedSkills(service contract.SkillService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limit := 10
		if l := ctx.Query("limit"); l != "" {
			if v, err := strconv.Atoi(l); err == nil && v > 0 {
				limit = v
			}
		}

		skills, err := service.ListRecentUsedSkills(ctx, limit)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
			return
		}
		if skills == nil {
			skills = []types.Skill{}
		}

		ctx.JSON(http.StatusOK, dto.Success(skills))
	}
}

func getSkillStatuses(service contract.SkillService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		codesStr := ctx.Query("codes")
		if codesStr == "" {
			ctx.JSON(http.StatusOK, dto.Success(map[string]string{}))
			return
		}

		codes := strings.Split(codesStr, ",")
		filtered := make([]string, 0, len(codes))
		for _, c := range codes {
			if trimmed := strings.TrimSpace(c); trimmed != "" {
				filtered = append(filtered, trimmed)
			}
		}

		statuses, err := service.GetSkillStatuses(ctx, filtered)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
			return
		}
		if statuses == nil {
			statuses = map[string]string{}
		}

		ctx.JSON(http.StatusOK, dto.Success(statuses))
	}
}
