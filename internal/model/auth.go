package model

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Remember bool   `json:"remember"`
}

type LoginResponse struct {
	Token string   `json:"token"`
	User  UserInfo `json:"user"`
}

type MeRequest struct {}

type MeResponse struct {
	User  UserInfo `json:"user"`
}

type UserInfo struct {
	ID       UserId `json:"id"`
	Name     string `json:"name"`
	Role     int    `json:"role"`
	Register string `json:"register"`
	Avatar   string `json:"avatar,omitempty"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterResponse struct{}