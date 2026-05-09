package middleware

import (
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	localauth "github.com/insmtx/SingerOS/backend/internal/api/auth"
	ygauth "github.com/ygpkg/yg-go/apis/runtime/auth"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/encryptor/snowflake"
)

const (
	headerKeyRequestID = "X-Request-ID"
	headerKeyTraceID   = "X-Trace-ID"
)

// CallerMiddleware .
func CallerMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		reqID := ctx.Request.Header.Get(headerKeyRequestID)
		if reqID == "" {
			reqID = snowflake.GenerateIDBase58()
		}
		traceID := ctx.Request.Header.Get(headerKeyTraceID)
		if traceID == "" {
			traceID = reqID
		}

		caller := parseCallerFromRequest(ctx, jwtSecret)

		localauth.WithGinContext(ctx, caller, &localauth.Trace{
			RequestID: reqID,
			TraceID:   traceID,
			SpanID:    []string{},
		})
		ctx.Next()
	}
}

func parseCallerFromRequest(ctx *gin.Context, jwtSecret string) *localauth.Caller {
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
		return &localauth.Caller{
			Uin:   0,
			OrgID: 0,
			State: localauth.AuthStateNil,
		}
	}

	claims, err := parseJWTToken(tokenStr, jwtSecret)
	if err != nil {
		logs.Warnw("parse jwt token failed", "error", err)
		return &localauth.Caller{
			Uin:   0,
			OrgID: 0,
			State: localauth.AuthStateFailed,
		}
	}

	if claims.Uin == 0 {
		return &localauth.Caller{
			Uin:   0,
			OrgID: 0,
			State: localauth.AuthStateFailed,
		}
	}

	return &localauth.Caller{
		Uin:   claims.Uin,
		OrgID: 0,
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
