package middleware

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"scannerAuth/internal/dto"
	"scannerAuth/internal/domain"
)

func ErrorHandler() gin.HandlerFunc {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		statusCode := http.StatusInternalServerError
		errResp := dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "internal server error",
		}

		switch err {
		case domain.ErrUserNotFound:
			statusCode = http.StatusNotFound
			errResp.Code = "NOT_FOUND"
			errResp.Message = "user not found"

		case domain.ErrEmailTaken:
			statusCode = http.StatusConflict
			errResp.Code = "EMAIL_TAKEN"
			errResp.Message = "email already in use"

		case domain.ErrInvalidCredentials:
			statusCode = http.StatusUnauthorized
			errResp.Code = "INVALID_CREDENTIALS"
			errResp.Message = "invalid credentials"

		case domain.ErrAccountLocked:
			statusCode = http.StatusLocked
			errResp.Code = "ACCOUNT_LOCKED"
			errResp.Message = "account is temporarily locked"

		default:
			logger.Error("unhandled error", "error", err.Error(), "request_id", c.GetString("X-Request-ID"))
		}

		c.JSON(statusCode, errResp)
	}
}
