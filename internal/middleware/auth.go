package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"scannerAuth/internal/dto"
	"scannerAuth/internal/pkg/token"
)

func AuthRequired(tokenMgr token.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Code:    "UNAUTHORIZED",
				Message: "missing authorization header",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Code:    "UNAUTHORIZED",
				Message: "invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenStr := parts[1]
		claims, err := tokenMgr.ValidateAccessToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Code:    "UNAUTHORIZED",
				Message: "invalid or expired token",
			})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Next()
	}
}
