package controller

import (
	"net/http"
	"reisen-be/internal/model"
	"reisen-be/internal/repository"
	"reisen-be/internal/service"

	"github.com/gin-gonic/gin"
)

type AvatarController struct {
	fileService *service.FileService
	userRepo    *repository.UserRepository
}

func NewAvatarController(fileService *service.FileService, userRepo *repository.UserRepository) *AvatarController {
	return &AvatarController{
		fileService: fileService,
		userRepo:    userRepo,
	}
}

// 上传用户头像
func (c *AvatarController) UploadAvatar(ctx *gin.Context) {
	// 从JWT中获取用户ID
	user := ctx.MustGet("user").(model.User)

	// 获取上传文件
	file, err := ctx.FormFile("avatar")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请选择头像文件"})
		return
	}

	// 保存头像文件
	filename, err := c.fileService.SaveAvatar(user.ID, file)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 更新用户头像信息
	user.Avatar = filename
	if err := c.userRepo.Update(&user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"avatar": filename})
}
