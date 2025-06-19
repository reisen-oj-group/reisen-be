package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type ProblemType string
type ProblemStatus string

const (
	ProblemTypeTraditional ProblemType = "traditional"
	ProblemTypeInteractive ProblemType = "interactive"

	ProblemStatusPublic  ProblemStatus = "public"
	ProblemStatusPrivate ProblemStatus = "private"
	ProblemStatusContest ProblemStatus = "contest"
)

type Statement struct {
	Background string `json:"background,omitempty"`
	Legend     string `json:"legend,omitempty"`
	FormatI    string `json:"formatI,omitempty"`
	FormatO    string `json:"formatO,omitempty"`
	Examples   []struct {
		DataI string `json:"dataI"`
		DataO string `json:"dataO"`
	} `json:"examples"`
	Hint string `json:"hint,omitempty"`
	Note string `json:"note,omitempty"`
}

type StatementsMap map[string]Statement

func (s *StatementsMap) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, s)
}

func (s StatementsMap) Value() (driver.Value, error) {
	return json.Marshal(s)
}

type TitlesMap map[string]string

func (t *TitlesMap) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, t)
}

func (t TitlesMap) Value() (driver.Value, error) {
	return json.Marshal(t)
}

type ProblemCore struct {
	BaseModel
	ID           ProblemId     `json:"id"`
	Type         ProblemType   `json:"type"`
	Status       ProblemStatus `json:"status"`
	LimitTime    int           `json:"limitTime"`    // 以 ms 为单位
	LimitMemory  int           `json:"limitMemory"`  // 以 kB 为单位
	CountCorrect int64         `json:"countCorrect"` // 通过提交记录个数
	CountTotal   int64         `json:"countTotal"`   // 全部提交记录个数
	Difficulty   int           `json:"difficulty"`   // Codeforces 难度评级，范围 800~3500
	Title        TitlesMap     `json:"title" gorm:"type:json"`
	Tags         []TagId       `json:"tags"  gorm:"type:json"`
	Provider     UserId        `json:"provider"`
}

func (ProblemCore) TableName() string {
	return "problems"
}

// 用于题目列表查询
type ProblemCoreWithJudgements struct {
	ProblemCore
	Judgements []Judgement `gorm:"foreignKey:ProblemID" json:"judgements"` // 反向关联
}

func (ProblemCoreWithJudgements) TableName() string {
	return "problems"
}

type Problem struct {
	ProblemCore
	Statements StatementsMap `json:"statements" gorm:"type:json"`

	// 测试数据相关字段
	HasTestdata bool `json:"hasTestdata" gorm:"default:false"`
	HasConfig   bool `json:"hasConfig"   gorm:"default:false"`
}

// 实现自定义 JSON 序列化/反序列化
func (p *Problem) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	return json.Unmarshal(value.([]byte), p)
}

func (p Problem) Value() ([]byte, error) {
	return json.Marshal(p)
}

type ProblemFilter struct {
	MinDifficulty *int           `json:"minDifficulty"`
	MaxDifficulty *int           `json:"maxDifficulty"`
	Tags          []int          `json:"tags"`
	Keywords      *string        `json:"keywords"`
	Provider      *UserId        `json:"provider"`
	Status        *ProblemStatus `json:"status"`
}

// 题目编辑请求
type ProblemEditRequest struct {
	Problem Problem `json:"problem"`
}

// 题目编辑响应
type ProblemEditResponse struct {
	Problem Problem `json:"problem"`
}

// 题目删除请求
type ProblemDeleteRequest struct {
	Problem ProblemId `json:"problem"`
}

// 题目删除响应
type ProblemDeleteResponse struct {
}

// 题目信息请求（携带用户信息用于查询用户本题提交信息）
type ProblemRequest struct {
	Problem ProblemId `json:"problem"`
	User    *UserId   `json:"user,omitempty"`
}

// 题目信息响应
type ProblemResponse struct {
	Problem   Problem    `json:"problem"`
	Judgement *Judgement `json:"judgement,omitempty"`
}

// 主题目列表查询
type ProblemListRequest struct {
	ProblemFilter
	Page *int `json:"page,omitempty"`
}

// 主题目列表响应
type ProblemListResponse struct {
	Total    int64                       `json:"total"`
	Problems []ProblemCoreWithJudgements `json:"problems"`
}

// 题目列表查询
type ProblemAllRequest struct {
	ProblemFilter
	Page *int `json:"page,omitempty"`
	Size *int `json:"size,omitempty"`
}

// 题目列表响应
type ProblemAllResponse struct {
	Total    int64         `json:"total"`
	Problems []ProblemCore `json:"problems"`
}
