package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	localauth "github.com/insmtx/Leros/backend/internal/api/auth"
	"github.com/insmtx/Leros/backend/types"
)

const headerStaticAPIKey = "X-Static-Api-Key"

func StaticAuth(staticAPIKey, jwtSecret string, db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if staticAPIKey != "" {
			apiKey := strings.TrimSpace(ctx.GetHeader(headerStaticAPIKey))
			if apiKey != "" && subtle.ConstantTimeCompare([]byte(apiKey), []byte(staticAPIKey)) == 1 {
				localauth.WithGinContext(ctx, types.SystemIdentity(), &types.Trace{})
				ctx.Next()
				return
			}
		}

		caller := parseCallerFromRequest(ctx, jwtSecret, db, "")
		if caller.State == types.AuthStateSucc {
			ctx.Next()
			return
		}

		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}
}
