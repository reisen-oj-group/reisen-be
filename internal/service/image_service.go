package service

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"path/filepath"
	"reisen-be/internal/repository"
	"reisen-be/internal/filesystem"
	"reisen-be/internal/model"
	"strings"
)

type ImageService struct {
	userRepo          *repository.UserRepository
	ImageFilesystem *filesystem.ImageFilesystem
}

func NewImageService(userRepo *repository.UserRepository, ImageFilesystem *filesystem.ImageFilesystem) *ImageService {
	return &ImageService{
		userRepo: userRepo,
		ImageFilesystem: ImageFilesystem,
	}
}

func (s *ImageService) decodeImage(file *multipart.FileHeader) (image.Image, error) {
	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	ext := strings.ToLower(filepath.Ext(file.Filename))
	
	switch ext {
	case ".jpg", ".jpeg":
		return jpeg.Decode(src)
	case ".png":
		return png.Decode(src)
	case ".gif":
		return gif.Decode(src)
	default:
		return nil, err
	}
}

func (s *ImageService) SaveAvatar(file *multipart.FileHeader, userID model.UserId) (*string, error) {
	img, err := s.decodeImage(file)
	if err != nil {
		return nil, err
	}
	filename := fmt.Sprintf("avatar%d.png", userID)
	virtpath := fmt.Sprintf("/images/avatar/%s", filename)

	// 调用文件系统保存转换后的图片
	_, err = s.ImageFilesystem.SaveAsPNG(img, "avatar", filename)
	if err != nil {
		return nil, err
	}

	s.userRepo.UpdateAvatar(userID, virtpath)

	// 返回前端路径而非真实路径给前端
	return &virtpath, nil
}