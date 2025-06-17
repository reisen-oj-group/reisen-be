package service

import (
	"archive/zip"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reisen-be/internal/model"
	"reisen-be/internal/query"
	"reisen-be/internal/repository"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

type ProblemService struct {
	problemListQuery *query.ProblemListQuery
	problemRepo      *repository.ProblemRepository
}

func NewProblemService(problemListQuery *query.ProblemListQuery, problemRepo *repository.ProblemRepository) *ProblemService {
	return &ProblemService{
		problemListQuery: problemListQuery,
		problemRepo:      problemRepo,
	}
}

func (s *ProblemService) CreateProblem(problem *model.Problem) error {
	return s.problemRepo.Create(problem)
}

func (s *ProblemService) UpdateProblem(problem *model.Problem) error {
	return s.problemRepo.Update(problem)
}

func (s *ProblemService) GetProblem(id model.ProblemId) (*model.Problem, error) {
	return s.problemRepo.GetByID(id)
}

func (s *ProblemService) AllProblems(filter *model.ProblemFilter, page, pageSize int) ([]model.ProblemCore, int64, error) {
	return s.problemRepo.List(filter, page, pageSize)
}

func (s *ProblemService) ListProblems(filter *model.ProblemFilter, userID *model.UserId, page, pageSize int) ([]model.ProblemCoreWithJudgements, int64, error) {
	return s.problemListQuery.List(filter, userID, page, pageSize)
}

func (s *ProblemService) DeleteProblem(id model.ProblemId) error {
	return s.problemRepo.Delete(id)
}

func GenerateConfig(dataDir, configPath string) error {
	entries, _ := os.ReadDir(dataDir)
	inputMap := map[string]string{}
	outputMap := map[string]string{}

	for _, f := range entries {
		name := f.Name()
		base := strings.TrimSuffix(name, filepath.Ext(name))
		if strings.HasSuffix(name, ".in") {
			inputMap[base] = name
		} else if strings.HasSuffix(name, ".out") || strings.HasSuffix(name, ".ans") {
			outputMap[base] = name
		}
	}

	var testcases []map[string]string
	for base, in := range inputMap {
		if out, ok := outputMap[base]; ok {
			testcases = append(testcases, map[string]string{
				"input":  "data/" + in,
				"output": "data/" + out,
			})
		}
	}

	cfg := map[string]interface{}{
		"spj":          false,
		"time_limit":   1000,
		"memory_limit": 262144,
		"testcases":    testcases,
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func UnzipAndValidate(zipPath, dest string) error {
    r, err := zip.OpenReader(zipPath)
    if err != nil {
        return err
    }
    defer r.Close()

    for _, f := range r.File {
        if strings.Contains(f.Name, "..") {
            return errors.New("invalid path in zip")
        }
        if strings.HasPrefix(f.Name, "data/") || strings.HasPrefix(f.Name, "meta/") {
            // only extract .in, .out, .ans, .cpp, .h, etc.
            // write file to disk under dest
        } else {
            return fmt.Errorf("unsupported root-level file: %s", f.Name)
        }
    }
    return nil
}

func UploadDataset(c *gin.Context) {
	id := c.Param("id")
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	// 创建目标目录
	basePath := filepath.Join("data", "problems", id)
	os.MkdirAll(basePath, 0755)

	// 保存 zip 临时文件并解压
	tmpPath := filepath.Join(os.TempDir(), "upload_"+uuid.New().String()+".zip")
	c.SaveUploadedFile(file, tmpPath)
	err = UnzipAndValidate(tmpPath, basePath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 自动生成 config.yml
	err = GenerateConfig(filepath.Join(basePath, "data"), filepath.Join(basePath, "config.yml"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Uploaded and processed"})
}
func DeleteDataset(c *gin.Context) {
    id := c.Param("id")
    basePath := filepath.Join("data", "problems", id)
    if err := os.RemoveAll(basePath); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "Deleted"})
}
