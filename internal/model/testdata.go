package model

// TestdataConfig 测试数据配置文件结构
type TestdataConfig struct {
	TimeLimit   int      `yaml:"time_limit"`    // 时间限制(ms)
	MemoryLimit int      `yaml:"memory_limit"`  // 内存限制(KB)
	TestCases   []TestCase `yaml:"test_cases"`   // 测试用例列表
}

// TestCase 单个测试用例配置
type TestCase struct {
	Input       string `yaml:"input"`        // 输入文件路径
	Output      string `yaml:"output"`       // 输出文件路径
	Score       int    `yaml:"score"`        // 该测试用例分值
}

type TestdataUploadRequest struct {
	ProblemID ProblemId `form:"problem"`
}

type TestdataUploadResponse struct {
	Valid       bool 	         `json:"valid"`
	Message     string         `json:"message"`
	Config      TestdataConfig `json:"config"`
}

type TestdataDownloadRequest struct {
	ProblemID ProblemId `json:"problem"`
}

type TestdataDownloadResponse struct {
	URL string `json:"url"`
}

type TestdataDeleteRequest struct {
	ProblemID ProblemId `json:"problem"`
}

type TestdataDeleteResponse struct {
	
}

type TestdataConfigUploadRequest struct {
	ProblemID ProblemId      `json:"problem"`
	Config    TestdataConfig `json:"config"`
}

type TestdataConfigUploadResponse struct {
	Valid       bool 	       `json:"valid"`
	Message     string       `json:"message"`
}

type TestdataConfigRequest struct {
	ProblemID ProblemId `json:"problem"`
}

type TestdataConfigResponse struct {
	Config TestdataConfig `json:"config"`
}
