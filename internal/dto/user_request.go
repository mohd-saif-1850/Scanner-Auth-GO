package dto

type UpdateUserRequest struct {
	FullName string `json:"full_name" binding:"omitempty,min=2,max=120"`
	Phone    string `json:"phone" binding:"omitempty"`
	Email    string `json:"email" binding:"omitempty,email"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=72"`
}

type DeleteRequest struct {
	Password string `json:"password" binding:"required"`
}
