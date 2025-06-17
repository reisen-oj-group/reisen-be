package controller

import (
	"net/http"
	"reisen-be/internal/model"
	"reisen-be/internal/service"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService *service.UserService
	judgeService *service.JudgeService
	contestService *service.ContestService
}

func NewUserController(
	userService *service.UserService,
	judgeService *service.JudgeService,
	contestService *service.ContestService,
) *UserController {
	return &UserController{
		userService: userService,
		judgeService: judgeService,
		contestService: contestService,
	}
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
		User: *user,
	})
}

// 获取用户列表
func (c *UserController) AllUsers(ctx *gin.Context) {
	var req model.UserListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pageSize := 50
	users, total, err := c.userService.ListUsers(&req.UserFilterParamsRaw, req.Page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, model.UserListResponse{
		Total:   int(total),
		Users:   users,
	})
}

// 编辑基本信息（不涉及密码修改）
func (c *UserController) EditUser(ctx *gin.Context) {
	var req model.UserEditRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// 从 JWT 获取当前操作角色
	operator := ctx.MustGet("user").(*model.User)

	if operator.Role < model.RoleAdmin && operator.ID != req.User.ID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "No permission"})
		return
	}

	user, err := c.userService.EditUser(&req.User)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, user)
}

// 删除用户帐号
func (c *UserController) DeleteUser(ctx *gin.Context) {
	var req model.UserDeleteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 从 JWT 获取当前操作角色
	operator := ctx.MustGet("user").(*model.User)

	if operator.Role < model.RoleSuper && operator.ID != req.User {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "No permission"})
		return
	}

	err := c.userService.DeleteUser(req.User)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusOK)
}


// 删除用户帐号
func (c *UserController) GetPractice(ctx *gin.Context) {
	var req model.UserPracticeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rankings, err := c.contestService.ListPractice(req.User)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	judgements, err := c.judgeService.ListPractice(req.User)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, model.UserPracticeResponse{
		Rankings: rankings,
		Judgements: judgements,
	})
}
