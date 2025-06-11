package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reisen-be/internal/model"
	"reisen-be/internal/repository"
	"strconv"
	"time"
)

type JudgeService struct {
	submissionRepo *repository.SubmissionRepository
	problemRepo    *repository.ProblemRepository
	userRepo       *repository.UserRepository
	dataDir       string
}

func NewJudgeService(
	submissionRepo *repository.SubmissionRepository,
	problemRepo *repository.ProblemRepository,
	userRepo *repository.UserRepository,
	dataDir string,
) *JudgeService {
	return &JudgeService{
		submissionRepo: submissionRepo,
		problemRepo:    problemRepo,
		userRepo:       userRepo,
		dataDir:        dataDir,
	}
}

// TestCase 表示一个测试用例
type TestCase struct {
	Input  string
	Output string
	Score  int
}

// getTestCases 从文件系统加载测试用例
func (s *JudgeService) getTestCases(problemID model.ProblemId) ([]TestCase, error) {
	problemDir := filepath.Join(s.dataDir, fmt.Sprint(problemID), "data")
	var testCases []TestCase

	// 遍历数据目录，查找所有 .in 文件
	files, err := os.ReadDir(problemDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read problem data directory: %v", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".in" {
			// 获取测试用例编号（例如 1.in -> 1）
			testCaseNum := file.Name()[0 : len(file.Name())-len(filepath.Ext(file.Name()))]

			// 读取输入文件
			inputPath := filepath.Join(problemDir, file.Name())
			input, err := os.ReadFile(inputPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read input file %s: %v", inputPath, err)
			}

			// 读取输出文件
			outputPath := filepath.Join(problemDir, testCaseNum+".out")
			output, err := os.ReadFile(outputPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read output file %s: %v", outputPath, err)
			}

			// 目前所有测试用例分数相同，后续可以从 config.yml 读取
			testCases = append(testCases, TestCase{
				Input:  string(input),
				Output: string(output),
				Score:  100 / len(files), // 平均分配分数
			})
		}
	}

	if len(testCases) == 0 {
		return nil, fmt.Errorf("no test cases found for problem %d", problemID)
	}

	return testCases, nil
}

// compileCode 编译用户代码（与之前相同，略作调整）
func compileCode(sourceCode string, lang model.CodeLangId) (string, *model.CompileInfo, error) {
	var args []string
	var env []string
	var sourceFile string

	switch lang {
	case "cpp":
		args = []string{"/usr/bin/g++", "a.cc", "-o", "a", "-O2", "-std=c++11"}
		env = []string{"PATH=/usr/bin:/bin"}
		sourceFile = "a.cc"
	case "c":
		args = []string{"/usr/bin/gcc", "a.c", "-o", "a", "-O2", "-std=c11"}
		env = []string{"PATH=/usr/bin:/bin"}
		sourceFile = "a.c"
	case "go":
		args = []string{"/usr/bin/go", "build", "-o", "a", "main.go"}
		env = []string{"PATH=/usr/bin:/bin", "GOPATH=/go"}
		sourceFile = "main.go"
	default:
		return "", nil, fmt.Errorf("unsupported language: %s", lang)
	}

	payload := map[string]interface{}{
		"cmd": []map[string]interface{}{{
			"args": args,
			"env":  env,
			"files": []map[string]interface{}{
				{"content": ""},
				{"name": "stdout", "max": 10240},
				{"name": "stderr", "max": 10240},
			},
			"cpuLimit":    10_000_000_000,
			"memoryLimit": 512 * 1024 * 1024,
			"procLimit":   50,
			"copyIn": map[string]interface{}{
				sourceFile: map[string]interface{}{
					"content": sourceCode,
				},
			},
			"copyOut":       []string{"stdout", "stderr"},
			"copyOutCached": []string{"a"},
		}},
	}

	data, _ := json.Marshal(payload)
	resp, err := http.Post("http://localhost:5050/run", "application/json", bytes.NewReader(data))
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	var results []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return "", nil, err
	}

	result := results[0]
	exitStatus := int(result["exitStatus"].(float64))

	compileInfo := &model.CompileInfo{
		Success: exitStatus == 0,
		Message: result["files"].(map[string]interface{})["stderr"].(string),
	}

	if exitStatus != 0 {
		return "", compileInfo, fmt.Errorf("compile failed with exit status %d", exitStatus)
	}

	fileId := result["fileIds"].(map[string]interface{})["a"].(string)
	return fileId, compileInfo, nil
}

// runCode 运行编译后的程序（与之前相同）
func runCode(fileId string, input string, timeLimit int, memoryLimit int) (string, string, int, int, error) {
	payload := map[string]interface{}{
		"cmd": []map[string]interface{}{{
			"args": []string{"./a"},
			"env":  []string{"PATH=/usr/bin:/bin"},
			"files": []map[string]interface{}{
				{"content": input},
				{"name": "stdout", "max": 10240},
				{"name": "stderr", "max": 10240},
			},
			"cpuLimit":    uint64(timeLimit) * 1_000_000,
			"memoryLimit": uint64(memoryLimit) * 1024 * 1024,
			"procLimit":   50,
			"copyIn": map[string]interface{}{
				"a": map[string]interface{}{
					"fileId": fileId,
				},
			},
			"copyOut": []string{"stdout", "stderr"},
		}},
	}

	data, _ := json.Marshal(payload)
	resp, err := http.Post("http://localhost:5050/run", "application/json", bytes.NewReader(data))
	if err != nil {
		return "", "", 0, 0, err
	}
	defer resp.Body.Close()

	var results []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return "", "", 0, 0, err
	}

	result := results[0]
	exitStatus := int(result["exitStatus"].(float64))
	timeUsed := int(result["time"].(float64) / 1_000_000)
	memoryUsed := int(result["memory"].(float64) / 1024)

	stdout := result["files"].(map[string]interface{})["stdout"].(string)
	stderr := result["files"].(map[string]interface{})["stderr"].(string)

	// 检查是否因超时或内存超出而终止
	if exitStatus != 0 {
		if timeUsed >= timeLimit {
			return stdout, stderr, timeUsed, memoryUsed, fmt.Errorf("time limit exceeded")
		}
		if memoryUsed >= memoryLimit {
			return stdout, stderr, timeUsed, memoryUsed, fmt.Errorf("memory limit exceeded")
		}
		return stdout, stderr, timeUsed, memoryUsed, fmt.Errorf("runtime error")
	}

	return stdout, stderr, timeUsed, memoryUsed, nil
}

// deleteFile 删除临时文件（与之前相同）
func deleteFile(fileId string) error {
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("http://localhost:5050/file/%s", fileId), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete file")
	}
	return nil
}

// SubmitCode 处理代码提交和评测
func (s *JudgeService) SubmitCode(req *model.JudgeRequest, userID model.UserId) (*model.SubmissionFull, error) {
	// 1. 创建初始提交记录
	now := time.Now()
	submission := &model.Submission{
		ProblemID:      req.Problem,
		UserID:         userID,
		ContestID:      req.Contest,
		SubmissionTime: now,
		Lang:           req.Lang,
		Code:           req.Code,
		CodeLength:     len(req.Code),
		Testcases:      make([]model.Testcase, 0),
		Verdict:        model.VerdictPD, // Pending
	}

	// 2. 获取题目信息
	problem, err := s.problemRepo.GetByID(req.Problem)
	if err != nil {
		return nil, err
	}

	// 3. 获取测试用例
	testCases, err := s.getTestCases(req.Problem)
	if err != nil {
		return nil, fmt.Errorf("failed to get test cases: %v", err)
	}

	// 4. 编译代码
	fileId, compileInfo, err := compileCode(req.Code, req.Lang)
	submission.CompileInfo = compileInfo

	if err != nil {
		// 编译失败
		evalTime := time.Now()
		submission.EvaluationTime = &evalTime
		submission.Verdict = model.VerdictCE // Compile Error
		submission.Score = new(int)
		*submission.Score = 0
	} else {
		defer deleteFile(fileId)

		// 5. 运行测试用例
		var maxTimeUsed int
		var maxMemoryUsed int
		allPassed := true
		totalScore := 0

		for i, tc := range testCases {
			output, stderr, timeUsed, memoryUsed, err := runCode(
				fileId,
				tc.Input,
				problem.LimitTime,
				problem.LimitMemory,
			)

			// 记录测试用例结果
			testResult := model.Testcase{
				ID:      i + 1,
				Input:   &tc.Input,
				Output:  &output,
				Checker: &stderr,
			}

			if err != nil {
				if err.Error() == "time limit exceeded" {
					testResult.Verdict = model.VerdictTLE
				} else if err.Error() == "memory limit exceeded" {
					testResult.Verdict = model.VerdictMLE
				} else {
					testResult.Verdict = model.VerdictRE
				}
				allPassed = false
			} else if output != tc.Output {
				testResult.Verdict = model.VerdictWA
				allPassed = false
			} else {
				testResult.Verdict = model.VerdictAC
				testResult.Score = &tc.Score
				totalScore += tc.Score
			}

			testResult.Time = &timeUsed
			testResult.Memory = &memoryUsed

			if timeUsed > maxTimeUsed {
				maxTimeUsed = timeUsed
			}
			if memoryUsed > maxMemoryUsed {
				maxMemoryUsed = memoryUsed
			}

			submission.Testcases = append(submission.Testcases, testResult)
		}

		// 6. 确定最终评测结果
		evalTime := time.Now()
		submission.EvaluationTime = &evalTime
		submission.TimeUsed = &maxTimeUsed
		submission.MemoryUsed = &maxMemoryUsed
		submission.Score = &totalScore

		if allPassed {
			submission.Verdict = model.VerdictAC
		} else {
			// 检查是否有运行时错误
			for _, tc := range submission.Testcases {
				if tc.Verdict == model.VerdictRE || tc.Verdict == model.VerdictTLE || tc.Verdict == model.VerdictMLE {
					submission.Verdict = tc.Verdict
					break
				}
			}
			if submission.Verdict == model.VerdictPD {
				submission.Verdict = model.VerdictWA
			}
		}
	}

	// 7. 保存提交记录
	if err := s.submissionRepo.Create(submission); err != nil {
		return nil, err
	}

	// 8. 更新题目统计信息
	if submission.Verdict == model.VerdictAC {
		if err := s.problemRepo.IncreaseSubmitCorrect(submission.ProblemID); err != nil {
			return nil, err
		}
	}
	if err := s.problemRepo.IncreaseSubmitTotal(submission.ProblemID); err != nil {
		return nil, err
	}

	// 9. 获取用户信息
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// 10. 构建响应
	return &model.SubmissionFull{
		SubmissionCore: model.SubmissionCore{
			ID:             submission.ID,
			ProblemID:      submission.ProblemID,
			UserID:         submission.UserID,
			ContestID:      submission.ContestID,
			SubmissionTime: submission.SubmissionTime,
			EvaluationTime: submission.EvaluationTime,
			Lang:           submission.Lang,
			Verdict:        submission.Verdict,
			Score:          submission.Score,
			TimeUsed:       submission.TimeUsed,
			MemoryUsed:     submission.MemoryUsed,
			CodeLength:     submission.CodeLength,
		},
		Code:        submission.Code,
		CompileInfo: submission.CompileInfo,
		Testcases:   submission.Testcases,
		Problem: model.ProblemCore{
			ID:           problem.ID,
			Type:         problem.Type,
			Status:       problem.Status,
			LimitTime:    problem.LimitTime,
			LimitMemory:  problem.LimitMemory,
			CountCorrect: problem.CountCorrect,
			CountTotal:   problem.CountTotal,
			Difficulty:   problem.Difficulty,
			Title:        problem.Title,
		},
		User: model.User{
			ID:       user.ID,
			Name:     user.Name,
			Role:     user.Role,
			Register: user.Register,
			Avatar:   user.Avatar,
		},
	}, nil
}

// GetSubmissionDetail 获取提交详情
func (s *JudgeService) GetSubmissionDetail(id int64) (*model.SubmissionFull, error) {
	submission, err := s.submissionRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	problem, err := s.problemRepo.GetByID(submission.ProblemID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(submission.UserID)
	if err != nil {
		return nil, err
	}

	return &model.SubmissionFull{
		SubmissionCore: model.SubmissionCore{
			ID:             submission.ID,
			ProblemID:      submission.ProblemID,
			UserID:         submission.UserID,
			ContestID:      submission.ContestID,
			SubmissionTime: submission.SubmissionTime,
			EvaluationTime: submission.EvaluationTime,
			Lang:           submission.Lang,
			Verdict:        submission.Verdict,
			Score:          submission.Score,
			TimeUsed:       submission.TimeUsed,
			MemoryUsed:     submission.MemoryUsed,
			CodeLength:     submission.CodeLength,
		},
		Code:        submission.Code,
		CompileInfo: submission.CompileInfo,
		Testcases:   submission.Testcases,
		Problem: model.ProblemCore{
			ID:           problem.ID,
			Type:         problem.Type,
			Status:       problem.Status,
			LimitTime:    problem.LimitTime,
			LimitMemory:  problem.LimitMemory,
			CountCorrect: problem.CountCorrect,
			CountTotal:   problem.CountTotal,
			Difficulty:   problem.Difficulty,
			Title:        problem.Title,
		},
		User: model.User{
			ID:       user.ID,
			Name:     user.Name,
			Role:     user.Role,
			Register: user.Register,
			Avatar:   user.Avatar,
		},
	}, nil
}

// ConvertFilterParamsRaw 将原始参数转换为处理后的参数
func (s *JudgeService) ConvertFilterParamsRaw(raw *model.SubmissionFilterParamsRaw) (*model.SubmissionFilterParams, error) {
	if raw == nil {
		return nil, nil
	}

	params := &model.SubmissionFilterParams{
			Problem: raw.Problem,
			Lang:    raw.Lang,
			Verdict: raw.Verdict,
	}

	// 处理 User 字段转换
	if raw.User != nil {
		// 尝试解析为数字 ID
		if userID, err := strconv.Atoi(*raw.User); err == nil {
			params.User = (*model.UserId)(&userID)
		} else {
			// 如果是字符串，查询用户 ID
			user, err := s.userRepo.FindByUsername(*raw.User)
			if err != nil {
					return nil, fmt.Errorf("failed to find user by name: %v", err)
			}
			params.User = &user.ID
		}
	}

	return params, nil
}

// ListSubmissions 获取提交列表
func (s *JudgeService) ListSubmissions(filterRaw *model.SubmissionFilterParamsRaw, page, pageSize int) ([]model.SubmissionLite, int64, error) {
	filter, err := s.ConvertFilterParamsRaw(filterRaw)
	if err != nil {
		return nil, 0, err
	}

	submissions, total, err := s.submissionRepo.List(filter, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	var lites []model.SubmissionLite
	for _, sub := range submissions {
		problem, err := s.problemRepo.GetByID(sub.ProblemID)
		if err != nil {
			return nil, 0, err
		}

		user, err := s.userRepo.GetByID(sub.UserID)
		if err != nil {
			return nil, 0, err
		}

		lites = append(lites, model.SubmissionLite{
			SubmissionCore: model.SubmissionCore{
				ID:             sub.ID,
				ProblemID:      sub.ProblemID,
				UserID:         sub.UserID,
				ContestID:      sub.ContestID,
				SubmissionTime: sub.SubmissionTime,
				EvaluationTime: sub.EvaluationTime,
				Lang:           sub.Lang,
				Verdict:        sub.Verdict,
				Score:          sub.Score,
				TimeUsed:       sub.TimeUsed,
				MemoryUsed:     sub.MemoryUsed,
				CodeLength:     sub.CodeLength,
			},
			Problem: model.ProblemCore{
				ID:           problem.ID,
				Type:         problem.Type,
				Status:       problem.Status,
				LimitTime:    problem.LimitTime,
				LimitMemory:  problem.LimitMemory,
				CountCorrect: problem.CountCorrect,
				CountTotal:   problem.CountTotal,
				Difficulty:   problem.Difficulty,
				Title:        problem.Title,
			},
			User: model.User{
				ID:       user.ID,
				Name:     user.Name,
				Role:     user.Role,
				Register: user.Register,
				Avatar:   user.Avatar,
			},
		})
	}

	return lites, total, nil
}