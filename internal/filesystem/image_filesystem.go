package filesystem

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
)

type ImageFilesystem struct {
	dataDir string
}

func NewImageFilesystem(dataDir string) *ImageFilesystem {
	return &ImageFilesystem{
		dataDir: dataDir,
	}
}

func (f *ImageFilesystem) GetImagePath() string {
	return f.dataDir
}

func (f *ImageFilesystem) GetCatagoryPath(catagory string) string {
	return filepath.Join(f.GetImagePath(), catagory)
}

func (f *ImageFilesystem) SaveAsPNG(img image.Image, catagory string, filename string) (*string, error) {
	path := filepath.Join(f.GetCatagoryPath(catagory), filename)
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return nil, err
	}

	// 创建目标文件
	out, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	// 保存为 PNG 格式
	return &path, png.Encode(out, img)
}
