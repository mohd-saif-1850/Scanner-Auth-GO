package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"scannerAuth/internal/domain"
	"scannerAuth/internal/dto"
	"scannerAuth/internal/service"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "VALIDATION_ERROR",
			"message": "invalid request body",
			"details": err.Error(),
		})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		if err == domain.ErrEmailTaken {
			c.JSON(http.StatusConflict, gin.H{
				"code":    "EMAIL_TAKEN",
				"message": "email already exists",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "internal server error",
		})
		return
	}

	response := dto.UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		FullName:  user.FullName,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}

	if user.Phone != nil {
		response.Phone = *user.Phone
	}

	c.JSON(http.StatusCreated, response)
}
