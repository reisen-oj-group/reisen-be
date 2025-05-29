package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

const (
	VerdictPD  VerdictId = "PD"
	VerdictAC  VerdictId = "AC"
	VerdictWA  VerdictId = "WA"
	VerdictRE  VerdictId = "RE"
	VerdictTLE VerdictId = "TLE"
	VerdictMLE VerdictId = "MLE"
	VerdictCE  VerdictId = "CE"
	VerdictUKE VerdictId = "UKE"
)

// Testcase 测试点详情
type Testcase struct {
	ID      int       `json:"id"`
	Verdict VerdictId `json:"verdict"`
	Time    *int      `json:"time,omitempty"`
	Memory  *int      `json:"memory,omitempty"`
	Score   *int      `json:"score,omitempty"`
	Input   *string   `json:"input,omitempty"`
	Output  *string   `json:"output,omitempty"`
	Checker *string   `json:"checker,omitempty"`
}

type TestcaseList []Testcase

func (t *TestcaseList) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, t)
}

func (t TestcaseList) Value() (driver.Value, error) {
	return json.Marshal(t)
}

// CompileInfo 编译信息
type CompileInfo struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (t *CompileInfo) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, t)
}

func (t CompileInfo) Value() (driver.Value, error) {
	return json.Marshal(t)
}

// Submission 提交记录
type Submission struct {
	ID             SubmissionId `gorm:"primaryKey;autoIncrement" json:"id"`
	ProblemID      ProblemId    `gorm:"index" json:"problem"`
	UserID         UserId       `gorm:"index" json:"user"`
	ContestID      *ContestId   `gorm:"index" json:"contest,omitempty"`
	SubmissionTime time.Time    `gorm:"not null" json:"submissionTime"`
	EvaluationTime *time.Time   `gorm:"index" json:"evaluationTime,omitempty"`
	Lang           CodeLangId   `gorm:"type:varchar(10);not null" json:"lang"`
	Verdict        VerdictId    `gorm:"type:varchar(10)" json:"verdict,omitempty"`
	Score          *int         `json:"score,omitempty"`
	TimeUsed       *int         `json:"timeUsed,omitempty"`
	MemoryUsed     *int         `json:"memoryUsed,omitempty"`
	CodeLength     int          `gorm:"not null" json:"codeLength"`
	Code           string       `gorm:"type:text;not null" json:"code"`
	CompileInfo    *CompileInfo `gorm:"type:json" json:"compile,omitempty"`
	Testcases      TestcaseList `gorm:"type:json" json:"testcases"`
}

func (Submission) TableName() string {
	return "submission"
}

// SubmissionCore 提交记录核心信息
type SubmissionCore struct {
	ID             SubmissionId `json:"id"`
	ProblemID      ProblemId    `json:"problem"`
	UserID         UserId       `json:"user"`
	ContestID      *ContestId   `json:"contest,omitempty"`
	SubmissionTime time.Time    `json:"submission"`
	EvaluationTime *time.Time   `json:"evaluation,omitempty"`
	Lang           CodeLangId   `json:"lang"`
	Verdict        VerdictId    `json:"verdict,omitempty"`
	Score          *int         `json:"score,omitempty"`
	TimeUsed       *int         `json:"time,omitempty"`
	MemoryUsed     *int         `json:"memory,omitempty"`
	CodeLength     int          `json:"length"`
}

// SubmissionLite 轻量级提交信息
type SubmissionLite struct {
	SubmissionCore
	Problem ProblemCore `json:"problem"`
	User    User        `json:"user"`
}

// SubmissionFull 完整提交信息
type SubmissionFull struct {
	SubmissionCore
	Code        string      `json:"code"`
	CompileInfo *CompileInfo `json:"compile,omitempty"`
	Testcases   TestcaseList `json:"detail"`
	Problem     ProblemCore `json:"problem"`
	User        User        `json:"user"`
}

// JudgeRequest 评测请求
type JudgeRequest struct {
	Problem  ProblemId    `json:"problem"`
	Lang     CodeLangId   `json:"lang"`
	Code     string       `json:"code"`
	Contest  *ContestId   `json:"contest,omitempty"`
}

// JudgeResponse 评测响应
type JudgeResponse struct {
	Record SubmissionId `json:"record"`
}

// RecordFilterParams 记录过滤参数
type RecordFilterParams struct {
	User    *int      `json:"user,omitempty"`
	Problem *int      `json:"problem,omitempty"`
	Lang    *string   `json:"lang,omitempty"`
	Verdict *VerdictId `json:"verdict,omitempty"`
}

// RecordListRequest 记录列表请求
type RecordListRequest struct {
	RecordFilterParams
	Page int `json:"page"`
}

// RecordListResponse 记录列表响应
type RecordListResponse struct {
	Total   int             `json:"total"`
	Records []SubmissionLite `json:"records"`
}

// RecordDetailRequest 记录详情请求
type RecordDetailRequest struct {
	ID int64 `json:"id"`
}

// RecordDetailResponse 记录详情响应
type RecordDetailResponse struct {
	Record SubmissionFull `json:"record"`
}

// Result 题目结果
type Result struct {
	Problem ProblemId `json:"problem"`
	Contest ContestId `json:"contest"`
	User    UserId    `json:"user"`

	Judge   interface{} `json:"judge"`
	Attempt int         `json:"attempt"`
	Penalty int         `json:"penalty"`
	Time    int         `json:"time"`
}
