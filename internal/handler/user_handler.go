package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"scannerAuth/internal/domain"
	"scannerAuth/internal/dto"
	"scannerAuth/internal/service"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID := mustUserID(c)

	resp, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		if err == domain.ErrUserNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: "NOT_FOUND", Message: "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: "INTERNAL_ERROR", Message: "internal server error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) Update(c *gin.Context) {
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "invalid request body",
			Details: err.Error(),
		})
		return
	}

	userID := mustUserID(c)

	resp, err := h.userService.Update(c.Request.Context(), userID, req)
	if err != nil {
		switch err {
		case domain.ErrEmailTaken:
			c.JSON(http.StatusConflict, dto.ErrorResponse{Code: "EMAIL_TAKEN", Message: "email already in use"})
		case domain.ErrUserNotFound:
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: "NOT_FOUND", Message: "user not found"})
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: "INTERNAL_ERROR", Message: "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "invalid request body",
			Details: err.Error(),
		})
		return
	}

	userID := mustUserID(c)

	if err := h.userService.ChangePassword(c.Request.Context(), userID, req); err != nil {
		if err == domain.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Code: "INVALID_CREDENTIALS", Message: "current password is incorrect"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: "INTERNAL_ERROR", Message: "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *UserHandler) Delete(c *gin.Context) {
	var req dto.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "password is required",
		})
		return
	}

	userID := mustUserID(c)

	if err := h.userService.Delete(c.Request.Context(), userID, req.Password); err != nil {
		if err == domain.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Code: "INVALID_CREDENTIALS", Message: "incorrect password"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: "INTERNAL_ERROR", Message: "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}

// mustUserID reads the authenticated user's UUID from Gin context (set by auth middleware).
func mustUserID(c *gin.Context) uuid.UUID {
	return c.MustGet("userID").(uuid.UUID)
}
