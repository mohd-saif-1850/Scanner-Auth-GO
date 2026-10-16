package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"scannerAuth/internal/domain"
	"scannerAuth/internal/dto"
	"scannerAuth/internal/service"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "invalid request body",
			Details: err.Error(),
		})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		if err == domain.ErrEmailTaken {
			c.JSON(http.StatusConflict, dto.ErrorResponse{Code: "EMAIL_TAKEN", Message: "email already in use"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: "INTERNAL_ERROR", Message: "internal server error"})
		return
	}

	resp := service.DomainUserToResponse(user)
	c.JSON(http.StatusCreated, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "invalid request body",
			Details: err.Error(),
		})
		return
	}

	meta := dto.ClientMeta{
		UserAgent: c.Request.UserAgent(),
		IPAddress: c.ClientIP(),
	}

	resp, err := h.authService.Login(c.Request.Context(), req, meta)
	if err != nil {
		switch err {
		case domain.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Code: "INVALID_CREDENTIALS", Message: "invalid email or password"})
		case domain.ErrAccountLocked:
			c.JSON(http.StatusLocked, dto.ErrorResponse{Code: "ACCOUNT_LOCKED", Message: "account is temporarily locked"})
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: "INTERNAL_ERROR", Message: "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "refresh_token is required",
		})
		return
	}

	resp, err := h.authService.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Code: "UNAUTHORIZED", Message: "invalid or expired refresh token"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "refresh_token is required",
		})
		return
	}

	_ = h.authService.Logout(c.Request.Context(), req.RefreshToken)
	c.Status(http.StatusNoContent)
}
