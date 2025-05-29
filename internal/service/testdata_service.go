package service

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reisen-be/internal/model"
	"reisen-be/internal/repository"
	// "time"

	// "gorm.io/gorm"
	"gopkg.in/yaml.v3"
	"fmt"
)

type TestdataService struct {
	problemRepo *repository.ProblemRepository
	dataDir      string
}

func NewTestdataService(problemRepo *repository.ProblemRepository, dataDir string) *TestdataService {
	return &TestdataService{
		problemRepo: problemRepo,
		dataDir:     dataDir,
	}
}

func (s *TestdataService) getProblemPath(problemID model.ProblemId) string {
	return filepath.Join(s.dataDir, fmt.Sprint(problemID))
}

func (s *TestdataService) getDataPath(problemID model.ProblemId) string {
	return filepath.Join(s.getProblemPath(problemID), "data")
}

func (s *TestdataService) getConfigPath(problemID model.ProblemId) string {
	return filepath.Join(s.getProblemPath(problemID), "config.yml")
}

func (s *TestdataService) UploadTestdata(problemID model.ProblemId, filePath string) error {
	// 确保问题目录存在
	problemPath := s.getProblemPath(problemID)
	if err := os.MkdirAll(problemPath, 0755); err != nil {
		return err
	}

	// 解压ZIP文件到data目录
	dataPath := s.getDataPath(problemID)
	if err := os.RemoveAll(dataPath); err != nil {
		return err
	}
	if err := os.Mkdir(dataPath, 0755); err != nil {
		return err
	}

	if err := unzip(filePath, dataPath); err != nil {
		return err
	}

	// 更新数据库记录
	return s.problemRepo.UpdateTestdataStatus(problemID, true, false)
}

func (s *TestdataService) DownloadTestdata(problemID model.ProblemId) (string, error) {
	dataPath := s.getDataPath(problemID)
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		return "", errors.New("test data not found")
	}

	// 创建临时ZIP文件
	tempDir := os.TempDir()
	zipPath := filepath.Join(tempDir, "problem_"+fmt.Sprint(problemID)+"_data.zip")
	
	if err := zipDir(dataPath, zipPath); err != nil {
		return "", err
	}

	return zipPath, nil
}

func (s *TestdataService) DeleteTestdata(problemID model.ProblemId) error {
	dataPath := s.getDataPath(problemID)
	if err := os.RemoveAll(dataPath); err != nil {
		return err
	}
	return s.problemRepo.UpdateTestdataStatus(problemID, false, false)
}

func (s *TestdataService) UploadConfig(problemID model.ProblemId, config *model.TestdataConfig) error {
	// 确保试题目录存在
	problemPath := s.getProblemPath(problemID)
	if err := os.MkdirAll(problemPath, 0755); err != nil {
		return err
	}

	// 写入YAML文件
	configPath := s.getConfigPath(problemID)
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		return err
	}

	// 更新数据库记录
	return s.problemRepo.UpdateTestdataStatus(problemID, true, true)
}

func (s *TestdataService) GetConfig(problemID model.ProblemId) (*model.TestdataConfig, error) {
	configPath := s.getConfigPath(problemID)
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config model.TestdataConfig
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// 辅助函数: 解压ZIP文件
func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// 辅助函数: 压缩目录为ZIP
func zipDir(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		header.Name = relPath

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
			return err
		}
		return nil
	})

	return err
}
