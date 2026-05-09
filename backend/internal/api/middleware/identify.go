package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/ygpkg/yg-go/encryptor/snowflake"
	ygauth "github.com/ygpkg/yg-go/apis/runtime/auth"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"

	localauth "github.com/insmtx/SingerOS/backend/internal/api/auth"
	"github.com/insmtx/SingerOS/backend/internal/infra/db"
)

const (
	headerKeyRequestID = "X-Request-ID"
	headerKeyTraceID   = "X-Trace-ID"
)

// CallerMiddleware .
func CallerMiddleware(jwtSecret string, database *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		reqID := ctx.Request.Header.Get(headerKeyRequestID)
		if reqID == "" {
			reqID = snowflake.GenerateIDBase58()
		}
		traceID := ctx.Request.Header.Get(headerKeyTraceID)
		if traceID == "" {
			traceID = reqID
		}

		caller := parseCallerFromRequest(ctx, jwtSecret, database, reqID)

		localauth.WithGinContext(ctx, caller, &localauth.Trace{
			RequestID: reqID,
			TraceID:   traceID,
			SpanID:    []string{},
		})
		ctx.Next()
	}
}

func parseCallerFromRequest(ctx *gin.Context, jwtSecret string, database *gorm.DB, reqID string) *localauth.Caller {
	authHeader := ctx.Request.Header.Get("Authorization")
	if authHeader == "" {
		return &localauth.Caller{
			Uin:   0,
			OrgID: 0,
			State: localauth.AuthStateNil,
		}
	}

	tokenStr := extractTokenFromHeader(authHeader)
	if tokenStr == "" {
		logs.Debugw("no valid token found in request", "authHeader", authHeader, "reqID", reqID)
		return &localauth.Caller{
			Uin:   0,
			OrgID: 0,
			State: localauth.AuthStateNil,
		}
	}

	userClaims, err := parseJWTToken(tokenStr, jwtSecret)
	if err != nil {
		logs.Warnw("parse jwt token failed", "error", err)
		return &localauth.Caller{
			Uin:   0,
			OrgID: 0,
			State: localauth.AuthStateFailed,
		}
	}

	if userClaims.Uin == 0 {
		return &localauth.Caller{
			Uin:   0,
			OrgID: 0,
			State: localauth.AuthStateFailed,
		}
	}

	queryCtx, cancel := context.WithTimeout(ctx.Request.Context(), 3*time.Second)
	defer cancel()

	userOrg, err := db.GetUserOrgByUin(queryCtx, database, userClaims.Uin)
	if err != nil {
		logs.Warnw("get user org by uin failed, db error", "error", err, "uin", userClaims.Uin, "reqID", ctx.Request.Header.Get(headerKeyRequestID))
		return &localauth.Caller{
			Uin:   userClaims.Uin,
			OrgID: 0,
			State: localauth.AuthStateFailed,
		}
	}

	if userOrg == nil {
		logs.Warnw("user org not found", "uin", userClaims.Uin)
		return &localauth.Caller{
			Uin:   userClaims.Uin,
			OrgID: 0,
			State: localauth.AuthStateFailed,
		}
	}

	return &localauth.Caller{
		Uin:   userClaims.Uin,
		OrgID: userOrg.OrgID,
		State: localauth.AuthStateSucc,
	}
}

func extractTokenFromHeader(authHeader string) string {
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	}
	return strings.TrimSpace(authHeader)
}

func parseJWTToken(tokenStr, jwtSecret string) (*ygauth.UserClaims, error) {
	claims := &ygauth.UserClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}
