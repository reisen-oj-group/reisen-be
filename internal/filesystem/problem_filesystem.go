package filesystem

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reisen-be/internal/model"
	"reisen-be/internal/utils"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type ProblemFilesystem struct {
	dataDir string
}

func NewProblemFilesystem(dataDir string) *ProblemFilesystem {
	return &ProblemFilesystem{
		dataDir: dataDir,
	}
}

func (f *ProblemFilesystem) GetProblemPath(problemID model.ProblemId) string {
	return filepath.Join(f.dataDir, fmt.Sprint(problemID))
}

func (f *ProblemFilesystem) GetDataPath(problemID model.ProblemId) string {
	return filepath.Join(f.GetProblemPath(problemID), "tests")
}

func (f *ProblemFilesystem) GetConfigPath(problemID model.ProblemId) string {
	return filepath.Join(f.GetProblemPath(problemID), "config.yml")
}

func (f *ProblemFilesystem) GetJudgeConfig(problemID model.ProblemId) (*model.JudgeConfig, error) {
	configPath := f.GetConfigPath(problemID)

	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config model.JudgeConfig
	if err := yaml.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}

// 从数据文件直接生成配置文件
func (f *ProblemFilesystem) GenerateConfig(problem model.ProblemCore) error {
	// 确保数据目录存在
	dataPath := f.GetDataPath(problem.ID)
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		return errors.New("暂无数据")
	}

	entries, _ := os.ReadDir(dataPath)
	IMap := map[string]string{}
	OMap := map[string]string{}

	for _, file := range entries {
		name := file.Name()
		base := strings.TrimSuffix(name, filepath.Ext(name))

		if strings.HasSuffix(name, ".in") {
			IMap[base] = name
		} else if strings.HasSuffix(name, ".out") || strings.HasSuffix(name, ".ans") {
			OMap[base] = name
		}
	}

	var count = 0

	var testcases []model.TestCaseConfig
	for base, iFile := range IMap {
		if oFile, ok := OMap[base]; ok {
			count = count + 1
			testcases = append(testcases, model.TestCaseConfig{
				ID:         count,
				InputFile:  "tests/" + iFile,
				OutputFile: "tests/" + oFile,
				Score:      10,
			})
		}
	}

	config := model.JudgeConfig{
		TimeLimit:   problem.LimitTime,
		MemoryLimit: problem.LimitMemory,
		TestCases:   testcases,
		CheckerType: "strict",
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	configPath := f.GetConfigPath(problem.ID)
	return os.WriteFile(configPath, data, 0644)
}


func (f *ProblemFilesystem) UploadTestdata(problemID model.ProblemId, filePath string) error {
	// 确保问题目录存在
	problemPath := f.GetProblemPath(problemID)
	if err := os.MkdirAll(problemPath, 0755); err != nil {
		return err
	}

	// 确保数据目录为空
	dataPath := f.GetDataPath(problemID)
	if err := os.RemoveAll(dataPath); err != nil {
		return err
	}
	// 创建数据目录
	if err := os.Mkdir(dataPath, 0755); err != nil {
		return err
	}
	// 解压数据到数据目录
	if err := utils.Unzip(filePath, dataPath); err != nil {
		return err
	}
	return nil
}

func (f *ProblemFilesystem) DownloadTestdata(problemID model.ProblemId) (*string, error) {
	// 确保数据目录存在
	dataPath := f.GetDataPath(problemID)
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		return nil, errors.New("暂无数据")
	}
	// 创建数据 zip 压缩包
	tempDir := os.TempDir()
	zipPath := filepath.Join(tempDir, "problem_"+fmt.Sprint(problemID)+"_data.zip")
	
	if err := utils.ZipFolder(dataPath, zipPath); err != nil {
		return nil, err
	}
	// 文件发送完成后删除临时文件
	go func() {
		time.Sleep(60 * time.Second) // 等待足够时间确保文件已发送
		os.Remove(zipPath)
	}()
	return &zipPath, nil
}

func (f *ProblemFilesystem) DeleteTestdata(problemID model.ProblemId) error {
	dataPath := f.GetDataPath(problemID)
	if err := os.RemoveAll(dataPath); err != nil {
		return err
	}
	return nil
}

func (f *ProblemFilesystem) UploadConfig(problemID model.ProblemId, config *model.TestdataConfig) error {
	// 确保试题目录存在
	problemPath := f.GetProblemPath(problemID)
	if err := os.MkdirAll(problemPath, 0755); err != nil {
		return err
	}
	// 写入 YAML 文件
	configPath := f.GetConfigPath(problemID)
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		return err
	}
	return nil
}

func (f *ProblemFilesystem) GetConfig(problemID model.ProblemId) (*model.TestdataConfig, error) {
	configPath := f.GetConfigPath(problemID)
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
