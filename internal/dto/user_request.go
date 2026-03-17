package dto

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type UserInfoUpdateRequest struct {
	Username  *string `json:"username" binding:"omitempty,min=3,max=32"`
	AvatarURL *string `json:"avatar_url" binding:"omitempty"`
}
