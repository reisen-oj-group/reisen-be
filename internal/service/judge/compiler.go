package judge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reisen-be/internal/model"
	"time"
)

type Compiler struct {
	client *http.Client
}

func NewCompiler() *Compiler {
	return &Compiler{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Compiler) Compile(task *model.JudgeTask) (string, *model.CompileInfo, error) {
	langConfig := getLangConfig(task.Lang)
	if langConfig == nil {
		return "", nil, fmt.Errorf("unsupported language: %s", task.Lang)
	}

	payload := model.RunRequestPayload{
		Cmd: []model.Cmd{
			{
				Args:        langConfig.CompileArgs,
				Env:         langConfig.CompileEnv,
				CPULimit:    10_000_000_000,    // 10s
				MemoryLimit: 512 * 1024 * 1024, // 512MB
				ProcLimit:   50,
				Files: []any{
					map[string]any{"content": ""},
					map[string]any{"name": "stdout", "max": 10240},
					map[string]any{"name": "stderr", "max": 10240},
				},
				CopyIn: map[string]any{
					langConfig.SourceFile: map[string]interface{}{
						"content": task.Code,
					},
				},
				CopyOut:       []string{"stdout", "stderr"},
				CopyOutCached: []string{langConfig.OutputFile},
			},
		},
	}

	data, _ := json.Marshal(payload)
	resp, err := c.client.Post("http://localhost:5050/run", "application/json", bytes.NewReader(data))
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	var results []GoJudgeResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return "", nil, err
	}

	result := results[0]
	exitStatus := result.ExitStatus

    info := "No information."

    if message, ok := result.Files["stderr"]; ok {
        info = message
    }

	compileInfo := &model.CompileInfo{
		Success: exitStatus == 0,
		Message: info,
	}

	if exitStatus != 0 {
		return "", compileInfo, fmt.Errorf("compile failed with exit status %d", exitStatus)
	}

	fileId := result.FileIds[langConfig.OutputFile]
	return fileId, compileInfo, nil
}

func (c *Compiler) DeleteFile(fileId string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("http://localhost:5050/file/%s", fileId), nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete file")
	}
	return nil
}

type LanguageConfig struct {
	SourceFile  string
	OutputFile  string
	CompileArgs []string
	CompileEnv  []string
	RunArgs     []string
	RunEnv      []string
}

func getLangConfig(lang model.CodeLangId) *LanguageConfig {
	configs := map[model.CodeLangId]*LanguageConfig{
		// support GCC 9
		"c": {
			SourceFile:  "a.c",
			OutputFile:  "a",
			CompileArgs: []string{"/opt/rh/devtoolset-9/root/usr/bin/gcc", "a.c", "-DONLINE_JUDGE", "-Wall", "-fno-asm", "-lm", "-march=native", "-o", "a", "-O2", "-std=c11"},
			CompileEnv:  []string{
				"PATH=/opt/rh/devtoolset-9/root/usr/bin:/usr/bin:/bin",
        "LD_LIBRARY_PATH=/opt/rh/devtoolset-9/root/usr/lib64:/lib64:/lib",
			},
			RunArgs:     []string{"./a"},
			RunEnv:      []string{
				"PATH=/opt/rh/devtoolset-9/root/usr/bin:/usr/bin:/bin",
        "LD_LIBRARY_PATH=/opt/rh/devtoolset-9/root/usr/lib64:/lib64:/lib",
			},
		},
		"cpp98": {
			SourceFile:  "a.cc",
			OutputFile:  "a",
			CompileArgs: []string{"/opt/rh/devtoolset-9/root/usr/bin/g++", "a.cc", "-DONLINE_JUDGE", "-Wall", "-fno-asm", "-lm", "-march=native", "-o", "a", "-O2", "-std=c++98"},
			CompileEnv:  []string{
				"PATH=/opt/rh/devtoolset-9/root/usr/bin:/usr/bin:/bin",
        "LD_LIBRARY_PATH=/opt/rh/devtoolset-9/root/usr/lib64:/lib64:/lib",
			},
			RunArgs:     []string{"./a"},
			RunEnv:      []string{
				"PATH=/opt/rh/devtoolset-9/root/usr/bin:/usr/bin:/bin",
        "LD_LIBRARY_PATH=/opt/rh/devtoolset-9/root/usr/lib64:/lib64:/lib",
			},
		},
		"cpp11": {
			SourceFile:  "a.cc",
			OutputFile:  "a",
			CompileArgs: []string{"/opt/rh/devtoolset-9/root/usr/bin/g++", "a.cc", "-DONLINE_JUDGE", "-Wall", "-fno-asm", "-lm", "-march=native", "-o", "a", "-O2", "-std=c++11"},
			CompileEnv:  []string{
				"PATH=/opt/rh/devtoolset-9/root/usr/bin:/usr/bin:/bin",
        "LD_LIBRARY_PATH=/opt/rh/devtoolset-9/root/usr/lib64:/lib64:/lib",
			},
			RunArgs:     []string{"./a"},
			RunEnv:      []string{
				"PATH=/opt/rh/devtoolset-9/root/usr/bin:/usr/bin:/bin",
        "LD_LIBRARY_PATH=/opt/rh/devtoolset-9/root/usr/lib64:/lib64:/lib",
			},
		},
		"cpp14": {
			SourceFile:  "a.cc",
			OutputFile:  "a",
			CompileArgs: []string{"/opt/rh/devtoolset-9/root/usr/bin/g++", "a.cc", "-DONLINE_JUDGE", "-Wall", "-fno-asm", "-lm", "-march=native", "-o", "a", "-O2", "-std=c++14"},
			CompileEnv:  []string{
				"PATH=/opt/rh/devtoolset-9/root/usr/bin:/usr/bin:/bin",
        "LD_LIBRARY_PATH=/opt/rh/devtoolset-9/root/usr/lib64:/lib64:/lib",
			},
			RunArgs:     []string{"./a"},
			RunEnv:      []string{
				"PATH=/opt/rh/devtoolset-9/root/usr/bin:/usr/bin:/bin",
        "LD_LIBRARY_PATH=/opt/rh/devtoolset-9/root/usr/lib64:/lib64:/lib",
			},
		},
		"pas": {
			SourceFile:  "a.pas",
			OutputFile:  "a",
			CompileArgs: []string{"/usr/bin/fpc", "a.pas", "-dONLINE_JUDGE", "-vnw", "-o", "a", "-O2", "-std=c11"},
			CompileEnv:  []string{"PATH=/usr/bin:/bin"},
			RunArgs:     []string{"./a"},
			RunEnv:      []string{"PATH=/usr/bin:/bin"},
		},
		// "go": {
		// 	SourceFile:  "main.go",
		// 	OutputFile:  "a",
		// 	CompileArgs: []string{"/usr/bin/go", "build", "-o", "a", "main.go"},
		// 	CompileEnv:  []string{"PATH=/usr/bin:/bin", "GOPATH=/go"},
		// 	RunArgs:     []string{"./a"},
		// 	RunEnv:      []string{"PATH=/usr/bin:/bin"},
		// },
		"python": {
			SourceFile:  "main.py",
			OutputFile:  "main.py",
			CompileArgs: nil, // Python 不需要编译
			RunArgs:     []string{"/usr/bin/python3", "main.py"},
			RunEnv:      []string{"PATH=/usr/bin:/bin"},
		},
	}

	return configs[lang]
}
