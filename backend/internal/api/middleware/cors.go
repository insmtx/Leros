package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool { return true },
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		AllowHeaders: []string{
			"Accept",
			"Origin",
			"Accept-Encoding",
			"Accept-Language",
			"Access-Control-Request-Headers",
			"Access-Control-Request-Method",
			"Host",
			"Proxy-Connection",
			"Referer",
			"Sec-Fetch-Mode",
			"User-Agent",
			"Content-Type",
			"Env",
			"Authorization",
			"Upgrade",
			"Connection",
			"Cache-Control",
			"X-Request-ID",
			"X-Trace-ID",
			"X-Requested-With",
			"X-Leros-Client-App",
			"X-Leros-Client-Version",
			"X-Leros-Client-Platform",
			"X-Leros-Client-Arch",
		},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
