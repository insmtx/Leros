package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/insmtx/Leros/backend/internal/infra/filestore"
)

const presignQueryParam = "presign"

// RegisterStaticRoutes registers static resource presign routes
func RegisterStaticRoutes(r gin.IRouter) {
	r.PUT("/:bucket/*key", handlePresignUpload)
	r.GET("/:bucket/*key", handlePresignDownload)
}

func handlePresignUpload(ctx *gin.Context) {
	if !isPresignRequest(ctx) {
		ctx.String(http.StatusBadRequest, "missing presign query parameter")
		return
	}

	bucket := strings.TrimSpace(ctx.Param("bucket"))
	key := strings.TrimPrefix(ctx.Param("key"), "/")

	if bucket == "" || key == "" {
		ctx.String(http.StatusBadRequest, "bucket and key are required")
		return
	}

	url, expiresAt, err := filestore.PresignUpload(ctx.Request.Context(), bucket, key)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "failed to generate presigned upload URL")
		return
	}

	ctx.Header("X-Presign-Expires-At", expiresAt.Format(time.RFC3339))
	ctx.String(http.StatusOK, url)
}

func handlePresignDownload(ctx *gin.Context) {
	if !isPresignRequest(ctx) {
		ctx.String(http.StatusBadRequest, "missing presign query parameter")
		return
	}

	bucket := strings.TrimSpace(ctx.Param("bucket"))
	key := strings.TrimPrefix(ctx.Param("key"), "/")

	if bucket == "" || key == "" {
		ctx.String(http.StatusBadRequest, "bucket and key are required")
		return
	}

	url, expiresAt, err := filestore.PresignDownload(ctx.Request.Context(), bucket, key)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "failed to generate presigned download URL")
		return
	}

	ctx.Header("X-Presign-Expires-At", expiresAt.Format(time.RFC3339))
	ctx.String(http.StatusOK, url)
}

func isPresignRequest(ctx *gin.Context) bool {
	return strings.TrimSpace(ctx.Query(presignQueryParam)) != ""
}
