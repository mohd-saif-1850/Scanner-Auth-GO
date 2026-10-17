package middleware

import (
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Logger() gin.HandlerFunc {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	return func(c *gin.Context) {
		requestID := uuid.NewString()
		c.Set("X-Request-ID", requestID)
		c.Header("X-Request-ID", requestID)

		start := time.Now()
		c.Next()
		duration := time.Since(start)

		logger.Info("http_request",
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", duration.Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}
