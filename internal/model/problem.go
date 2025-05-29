package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
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

type TitleMap map[string]string

func (t *TitleMap) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, t)
}

func (t TitleMap) Value() (driver.Value, error) {
	return json.Marshal(t)
}


type Problem struct {
	ID           ProblemId              `gorm:"primaryKey;autoIncrement" json:"id"`
	Type         ProblemType            `gorm:"type:varchar(20)" json:"type"`
	Status       ProblemStatus          `gorm:"type:varchar(10)" json:"status"`
	LimitTime    int                    `gorm:"not null" json:"limitTime"`
	LimitMemory  int                    `gorm:"not null" json:"limitMemory"`
	Statements   StatementsMap          `gorm:"type:json" json:"statements"`
	Title        TitleMap               `gorm:"type:json" json:"title"`
	CountCorrect int                    `gorm:"default:0" json:"countCorrect"`
	CountTotal   int                    `gorm:"default:0" json:"countTotal"`
	Difficulty   int                    `gorm:"not null" json:"difficulty"`
	Tags         []ProblemTag           `gorm:"foreignKey:ProblemID" json:"tags,omitempty"`
	
	// 新增测试数据相关字段
	HasTestdata bool      `gorm:"default:false" json:"hasTestdata"`
	HasConfig   bool      `gorm:"default:false" json:"hasConfig"`

	// 创建时间和更新时间
	CreatedAt    time.Time              `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time              `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Problem) TableName() string {
	return "problem"
}

type ProblemCore struct {
	ID           ProblemId     `json:"id"`
	Type         ProblemType   `json:"type"`
	Status       ProblemStatus `json:"status"`
	LimitTime    int           `json:"limitTime"`
	LimitMemory  int           `json:"limitMemory"`
	CountCorrect int           `json:"countCorrect"`
	CountTotal   int           `json:"countTotal"`
	Difficulty   int           `json:"difficulty"`
	Title        map[string]string `json:"title"`
	Tags         []TagId       `json:"tags"`
}

// ProblemTag 题目标签关联表
type ProblemTag struct {
	ProblemID ProblemId `gorm:"primaryKey" json:"problemId"`
	TagID     TagId     `gorm:"primaryKey" json:"tagId"`
}

func (ProblemTag) TableName() string {
	return "problem_tag"
}

// 实现自定义JSON序列化/反序列化
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
	MinDifficulty *int
	MaxDifficulty *int
	Tags          []int
	Keywords      *string
	Status        *ProblemStatus
}

type ProblemEditRequest struct {
	Problem Problem `json:"problem"`
}

type ProblemEditResponse struct {
	Problem Problem `json:"problem"`
}

type ProblemRequest struct {
	Problem  ProblemId `json:"problem"`
	User    *UserId    `json:"user,omitempty"`
}

type ProblemResponse struct {
	Problem Problem `json:"problem"`
	Result  *Result `json:"result,omitempty"`
}

type ProblemListRequest struct {
	Page          *int    `json:"page,omitempty"`
	User          *UserId `json:"user,omitempty"`
	MinDifficulty *int    `json:"minDifficulty,omitempty"`
	MaxDifficulty *int    `json:"maxDifficulty,omitempty"`
	Tags          []int   `json:"tags,omitempty"`
	Keywords      *string `json:"keywords,omitempty"`
}

type ProblemListResponse struct {
	Total    int           `json:"total"`
	Problems []ProblemCore `json:"problems"`
	Results  []Result      `json:"results,omitempty"`
}
