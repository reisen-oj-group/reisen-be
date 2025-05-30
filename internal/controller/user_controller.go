package controller

import (
	"net/http"
	"reisen-be/internal/model"
	"reisen-be/internal/service"
	"time"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService *service.UserService
}

func NewUserController(userService *service.UserService) *UserController {
	return &UserController{userService: userService}
}


func (s *UserController) GetUser(ctx *gin.Context) {
	var req model.UserRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := s.userService.GetUser(req.User)
	if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
	}

	ctx.JSON(http.StatusOK, model.UserResponse{
		User: model.UserInfo{
			ID:       user.ID,
			Name:     user.Name,
			Role:     int(user.Role),
			Register: user.Register.Format(time.RFC3339),
			Avatar:   user.Avatar,
		},
	})
}