package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/SingerOS/backend/internal/api/auth"
	"github.com/ygpkg/yg-go/apis/constants"
	"github.com/ygpkg/yg-go/encryptor/snowflake"
)

// CallerMiddleware .
func CallerMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		reqID := ctx.Request.Header.Get(constants.HeaderKeyRequestID)
		if reqID == "" {
			reqID = snowflake.GenerateIDBase58()
		}
		traceID := ctx.Request.Header.Get(constants.HeaderKeyTraceID)
		if traceID == "" {
			traceID = reqID
		}
		auth.WithContext(ctx, &auth.Caller{
			Uin:   0,
			OrgID: 0,
		}, &auth.Trace{
			RequestID: reqID,
			TraceID:   traceID,
			SpanID:    []string{},
		})
		ctx.Next()
	}
}
