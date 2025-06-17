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

// 测试点详情
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

// 编译信息
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

// 提交记录核心信息
type SubmissionCore struct {
	BaseModel
	ID          SubmissionId `json:"id"`
	ProblemID   ProblemId    `json:"problem"`
	UserID      UserId       `json:"user"`
	ContestID   *ContestId   `json:"contest,omitempty"`
	SubmittedAt time.Time    `json:"submittedAt"`
	ProcessedAt time.Time    `json:"processedAt"`
	Lang        CodeLangId   `json:"lang"`
	Verdict     VerdictId    `json:"verdict"`
	Score       *int         `json:"score,omitempty"`
	TimeUsed    *int         `json:"time,omitempty"`
	MemoryUsed  *int         `json:"memory,omitempty"`
	CodeLength  int          `json:"length"`
}

// 提交记录
type Submission struct {
	SubmissionCore
	Code        string       `gorm:"type:text" json:"code"`
	CompileInfo *CompileInfo `gorm:"type:json" json:"compile,omitempty"`
	Testcases   TestcaseList `gorm:"type:json" json:"detail"`
}

// 轻量提交记录（用于记录列表）
type SubmissionLite struct {
	SubmissionCore
	Problem ProblemCore `json:"problem"`
	User    User        `json:"user"`
}

// 完整提交记录（用于记录详情）
type SubmissionFull struct {
	Submission
	Problem ProblemCore `json:"problem"`
	User    User        `json:"user"`
}

// 评测请求
type JudgeRequest struct {
	Problem ProblemId  `json:"problem"`
	Lang    CodeLangId `json:"lang"`
	Code    string     `json:"code"`
	Contest *ContestId `json:"contest,omitempty"`
}

// 评测响应
type JudgeResponse struct {
	Submission SubmissionId `json:"submission"`
}

// 记录过滤参数
type SubmissionFilter struct {
	User    *UserId     `json:"user,omitempty"`
	Problem *ProblemId  `json:"problem,omitempty"`
	Lang    *CodeLangId `json:"lang,omitempty"`
	Verdict *VerdictId  `json:"verdict,omitempty"`
}

// 记录传递过来的过滤参数
type SubmissionFilterRaw struct {
	User    *string     `json:"user,omitempty"`
	Problem *ProblemId  `json:"problem,omitempty"`
	Lang    *CodeLangId `json:"lang,omitempty"`
	Verdict *VerdictId  `json:"verdict,omitempty"`
}

// 记录列表请求
type SubmissionListRequest struct {
	SubmissionFilterRaw
	Page int `json:"page"`
}

// 记录列表响应
type SubmissionListResponse struct {
	Total       int64            `json:"total"`
	Submissions []SubmissionLite `json:"submissions"`
}

// 记录详情请求
type SubmissionDetailRequest struct {
	ID int64 `json:"id"`
}

// 记录详情响应
type SubmissionDetailResponse struct {
	Submission SubmissionFull `json:"submission"`
}

// 题目结果（用于个人练习数据展示和题目列表）
type Judgement struct {
	ProblemID  ProblemId  `gorm:"primaryKey" json:"problem"`
	UserID     UserId     `gorm:"primaryKey" json:"user"`
	Judge      string     `json:"judge"`
	Difficulty int        `json:"difficulty"`	// 防止练习数据大量查询
	Stamp      *time.Time `json:"stamp,omitempty"` // 通过时间
}

type JudgementList []Judgement

func (r *JudgementList) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, r)
}

func (r JudgementList) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// 结果过滤参数
type JudgementFilter struct {
	User    *UserId    `json:"user,omitempty"`
	Problem *ProblemId `json:"problem,omitempty"`
}

// 结果更新请求
type JudgementUpdateRequest struct {
	Contest   ContestId
	User      UserId
	Problem   ProblemId
	Judgement Judgement
}

// 结果更新响应
type JudgementUpdateResponse struct {
	Ranking Ranking
}
