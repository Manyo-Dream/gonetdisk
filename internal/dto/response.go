package dto

type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type UserResponse struct {
	Username   string `json:"username"`
	Email      string `json:"email"`
	AvatarUrl  string `json:"avatar_url"`
	UsedSpace  uint64 `json:"used_space"`
	TotalSpace uint64 `json:"total_space"`
	Status     int    `json:"status"`
}
