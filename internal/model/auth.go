package model

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

type LoginResponse struct {
	Token string   `json:"token"`
	User  User     `json:"user"`
}

type MeRequest struct {}

type MeResponse struct {
	User  User     `json:"user"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponse struct{}

type CreateRequest struct {
	User User       `json:"user"`
	Password string `json:"password"`
}

type CreateResponse struct{}

type SetPasswordRequest struct {
	User        UserId `json:"user"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type SetPasswordResponse struct{}