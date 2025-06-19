package controller

import (
	"net/http"
	"path/filepath"
	"strings"

	"reisen-be/internal/model"
	"reisen-be/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	maxAvatarSize = 2 << 20 // 2MB
	maxBannerSize = 5 << 20 // 5MB
)

type ImageController struct {
	imageService *service.ImageService
}

func NewImageController(imageService *service.ImageService) *ImageController {
	return &ImageController{
		imageService: imageService,
	}
}

func (c *ImageController) UploadAvatar(ctx *gin.Context) {
	// 获取上传的文件
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No image uploaded"})
		return
	}
	// 检查文件大小
	if file.Size > maxAvatarSize {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "File too large."})
		return
	}
	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
	}
	if !allowedExts[ext] {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only jpg, jpeg, png and gif are allowed"})
		return
	}
	
	user := ctx.MustGet("user").(*model.User)
	path, err := c.imageService.SaveAvatar(file, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	ctx.JSON(http.StatusOK, model.AvatarResponse{
		Path: *path,
	})
}

func (c *ImageController) UploadBanner(ctx *gin.Context) {
	// 获取上传的文件
	file, err := ctx.FormFile("image")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No image uploaded"})
		return
	}
	// 检查文件大小
	if file.Size > maxBannerSize {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "File too large."})
		return
	}
	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
	}
	if !allowedExts[ext] {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only jpg, jpeg, png and gif are allowed"})
		return
	}
	
	user := ctx.MustGet("user").(*model.User)
	path, err := c.imageService.SaveAvatar(file, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	ctx.JSON(http.StatusOK, model.AvatarResponse{
		Path: *path,
	})
}