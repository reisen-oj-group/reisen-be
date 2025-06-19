package judge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reisen-be/internal/model"
	"sync"
	"time"
)

type GoJudgeResult struct {
	Status     string             `json:"status"`
	Error      string             `json:"error,omitempty"`
	ExitStatus int                `json:"exitStatus"`
	Time       int64              `json:"time"`
	Memory     int64              `json:"memory"`
	RunTime    int64              `json:"runTime"`
	Files      map[string]string  `json:"files"`
	FileIds    map[string]string  `json:"fileIds,omitempty"`
	FileError  []GoJudgeFileError `json:"fileError,omitempty"`
}

type GoJudgeFileError struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Message string `json:"message,omitempty"`
}

type Runner struct {
	client *http.Client
	mu     sync.Mutex
}

func NewRunner() *Runner {
	return &Runner{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// 运行得到测试点结果
func (r *Runner) Run(task *model.JudgeTask, fileId string, testCase model.TestCaseConfig, root string) (*model.Testcase, error) {
	langConfig := getLangConfig(task.Lang)
	if langConfig == nil {
		return nil, fmt.Errorf("unsupported language: %s", task.Lang)
	}

	inputFile, err := os.Open(filepath.Join(root, testCase.InputFile))
	if err != nil {
		return nil, err
	}
	defer inputFile.Close()

	// 创建一个最多容纳 256 字节的缓冲区
	buffer := make([]byte, 256)

	// 从输入文件中读取数据到缓冲区
	inputLen, err := inputFile.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, err
	}
	stdin := string(buffer[:inputLen])

	payload := model.RunRequestPayload{
		Cmd: []model.Cmd{
			{
				Args:        langConfig.RunArgs,
				Env:         langConfig.RunEnv,
				CPULimit:    uint64(task.Config.TimeLimit) * 1_000_000,
				MemoryLimit: uint64(task.Config.MemoryLimit) * 1024 * 1024,
				ProcLimit:   50,
				Files: []any{
					map[string]any{"src": filepath.Join(root, testCase.InputFile)}, // input_file -> stdin
					map[string]any{"name": "stdout", "max": 10240},                 // stdout -> stdout
					map[string]any{"name": "stderr", "max": 10240},                 // stderr -> stderr
				},
				CopyIn: map[string]any{
					langConfig.OutputFile: map[string]any{
						"fileId": fileId,
					},
				},
				CopyOut: []string{"stdout", "stderr"},
			},
		},
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	data, _ := json.Marshal(payload)
	resp, err := r.client.Post("http://localhost:5050/run", "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var results []GoJudgeResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}
	result := results[0]
	stdout := ""
	stderr := ""

	if message, ok := result.Files["stdout"]; ok {
		stdout = message
	}
	if message, ok := result.Files["stderr"]; ok {
		stderr = message
	}

	timeUsed := int(result.Time / 1_000_000)
	memoryUsed := int(result.Memory / 1024)

	testResult := &model.Testcase{
		ID:      testCase.ID,
		Time:    &timeUsed,
		Memory:  &memoryUsed,
		Input:   &stdin,
		Output:  &stdout,
		Checker: &stderr,
	}

	switch result.Status {
	case model.StatusAccepted:
		testResult.Verdict = model.VerdictAC
	case model.StatusMemoryLimitExceeded:
		testResult.Verdict = model.VerdictMLE
	case model.StatusTimeLimitExceeded:
		testResult.Verdict = model.VerdictTLE
	case model.StatusOutputLimitExceeded:
		testResult.Verdict = model.VerdictOLE
	case model.StatusFileError:
		testResult.Verdict = model.VerdictUKE
		if len(result.FileError) > 0 {
			msg := result.FileError[0].Message
			testResult.Checker = &msg
		}
	case model.StatusNonzeroExitStatus:
		msg := fmt.Sprintf("Program exited with code %d", result.ExitStatus)
		testResult.Verdict = model.VerdictRE
		testResult.Checker = &msg
	case model.StatusSignalled:
		msg := "Program terminated by signal"
		testResult.Verdict = model.VerdictRE
		testResult.Checker = &msg
	case model.StatusInternalError:
		testResult.Verdict = model.VerdictUKE
	default:
		testResult.Verdict = model.VerdictUKE
	}

	return testResult, nil
}
