package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/datatypes"
)

type ContestDifficulty int
type ContestStatus string
type ContestRule string

const (
	ContestDifficulty1 ContestDifficulty = 1
	ContestDifficulty2 ContestDifficulty = 2
	ContestDifficulty3 ContestDifficulty = 3
	ContestDifficulty4 ContestDifficulty = 4
	ContestDifficulty5 ContestDifficulty = 5

	ContestStatusPrivate ContestStatus = "private"
	ContestStatusPublic  ContestStatus = "public"
	ContestStatusDeleted ContestStatus = "deleted"

	ContestRuleOI  ContestRule = "OI"
	ContestRuleACM ContestRule = "ACM"
	ContestRuleIOI ContestRule = "IOI"
)

type ContestProblems map[ProblemLabel]ProblemId

func (c *ContestProblems) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, c)
}

func (c ContestProblems) Value() (driver.Value, error) {
	return json.Marshal(c)
}

type ContestProblemStatus struct {
	FirstBloodUserID *UserId    `json:"firstBloodUserId,omitempty"`
	FirstBloodTime   *time.Time `json:"firstBloodTime,omitempty"`
	SolvedCount      int        `json:"solvedCount"`
	TotalCount       int        `json:"totalCount"`
}

type ContestProblemStatuses map[ProblemId]ContestProblemStatus
func (c *ContestProblemStatuses) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, c)
}

func (c ContestProblemStatuses) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// 比赛基本信息
type Contest struct {
	BaseModel
	ID            ContestId         `json:"id"`
	Title         string            `gorm:"size:100"         json:"title"`
	Banner        string            `gorm:"size:200"         json:"banner,omitempty"`
	Summary       string            `gorm:"type:text"        json:"summary"`
	Description   string            `gorm:"type:text"        json:"description"`
	Difficulty    ContestDifficulty `gorm:"not null"         json:"difficulty"`
	Status        ContestStatus     `gorm:"type:varchar(10)" json:"status"`
	StartTime     time.Time         `gorm:"not null"         json:"startTime"`
	EndTime       time.Time         `gorm:"not null"         json:"endTime"`
	Rule          ContestRule       `gorm:"type:varchar(10)" json:"rule"`
	Problems      ContestProblems   `gorm:"type:json"        json:"problems"`
	ProblemStatus ContestProblemStatuses `gorm:"type:json" json:"problemStatus,omitempty"`
}

// 比赛报名信息
type Signup struct {
	ContestID ContestId `gorm:"primaryKey"     json:"contest"`
	UserID    UserId    `gorm:"primaryKey"     json:"user"`
	Stamp     time.Time `gorm:"autoCreateTime" json:"register"`
}

func (Signup) TableName() string {
	return "signups"
}

// 用于比赛列表查询
type ContestWithSignups struct {
	Contest
	Signups []Signup `gorm:"foreignKey:ContestID" json:"signups"` // 反向关联
}

func (ContestWithSignups) TableName() string {
	return "contests"
}

// 比赛排名信息
type Ranking struct {
	ContestID ContestId      `gorm:"primaryKey"  json:"contest"`
	UserID    UserId         `gorm:"primaryKey"  json:"user"`
	Team      string         `gorm:"size:50"     json:"team"`
	Ranking   int            `                   json:"ranking"`
	Detail    datatypes.JSON `gorm:"type:json"   json:"detail"`
}

func (Ranking) TableName() string {
	return "rankings"
}

// ACM problem cell data
type ACMCell struct {
	IsFirst   bool `json:"isFirst"`   // 是否为一血
	IsSolved  bool `json:"isSolved"`  // 是否已通过
	AttemptBF int  `json:"attemptBF"` // 封榜前尝试次数
	AttemptAF int  `json:"attemptAF"` // 封榜后尝试次数
	Penalty   int  `json:"penalty"`   // 罚时
}

// ACM ranking detail
type ACMDetail struct {
	Type         string                `json:"type"`         // "ACM"
	TotalPenalty int                   `json:"totalPenalty"` // 总罚时
	TotalSolved  int                   `json:"totalSolved"`  // 总通过数
	Problems     map[ProblemId]ACMCell `json:"problems"`
}

// OI problem cell data
type OIProblem struct {
	Score int `json:"score"`
}

// OI ranking detail
type OIDetail struct {
	Type       string                  `json:"type"`       // "OI"
	TotalScore int                     `json:"totalScore"` // 总分
	Problems   map[ProblemId]OIProblem `json:"problems"`
}

// IOI problem cell data (same as OI for now)
type IOIProblem struct {
	Score int `json:"score"`
}

// IOI ranking detail
type IOIDetail struct {
	Type       string                   `json:"type"`       // "IOI"
	TotalScore int                      `json:"totalScore"` // 总分
	Problems   map[ProblemId]IOIProblem `json:"problems"`
}

// 比赛过滤条件
type ContestFilter struct {
	Status     *ContestStatus
	Rule       *ContestRule
	Difficulty *ContestDifficulty
	UserID     *UserId // 用户是否已报名
	Keyword    *string
	Before     *time.Time // 开始时间之前
	After      *time.Time // 开始时间之后
}

// 原始过滤参数
type ContestFilterRaw struct {
	Status     *ContestStatus     `json:"status,omitempty"`
	Rule       *ContestRule       `json:"rule,omitempty"`
	Difficulty *ContestDifficulty `json:"difficulty,omitempty"`
	User       *string            `json:"user,omitempty"`
	Keyword    *string            `json:"keyword,omitempty"`
	Before     *time.Time         `json:"before,omitempty"`
	After      *time.Time         `json:"after,omitempty"`
}

// 比赛详情请求
type ContestRequest struct {
	Contest ContestId `json:"contest"`
}

// 比赛详情响应
type ContestResponse struct {
	Contest Contest  `json:"contest"`
	Signup  *Signup  `json:"signup,omitempty"`
	Ranking *Ranking `json:"ranking,omitempty"`
}

// 比赛试题请求
type ContestProblemsRequest struct {
	Contest ContestId `json:"contest"`
}

// 比赛试题响应
type ContestProblemsResponse struct {
	Problems []ProblemCore `json:"problems"`
}

// 比赛列表请求
type ContestListRequest struct {
	ContestFilterRaw
	Page *int `json:"page,omitempty"`
}

// 比赛列表响应
type ContestListResponse struct {
	Total    int64                `json:"total"`
	Contests []ContestWithSignups `json:"contests"`
}

// 比赛列表请求
type ContestAllRequest struct {
	ContestFilterRaw
	Page *int `json:"page,omitempty"`
	Size *int `json:"size,omitempty"`
}

// 比赛列表响应
type ContestAllResponse struct {
	Total    int64     `json:"total"`
	Contests []Contest `json:"contests"`
}

// 比赛编辑请求
type ContestEditRequest struct {
	Contest Contest `json:"contest"`
}

// 比赛编辑响应
type ContestEditResponse struct {
	Contest Contest `json:"contest"`
}

// 比赛删除请求
type ContestDeleteRequest struct {
	Contest ContestId `json:"contest"`
}

// 比赛删除响应
type ContestDeleteResponse struct {
}

// 比赛报名请求
type ContestSignupRequest struct {
	Contest ContestId `json:"contest"`
}

// 比赛报名响应
type ContestSignupResponse struct {
}

// 比赛取消报名请求
type ContestSignoutRequest struct {
	Contest ContestId `json:"contest"`
}

// 比赛取消报名响应
type ContestSignoutResponse struct {
}

// 比赛排名请求
type ContestRankingRequest struct {
	Contest ContestId `json:"contest"`
}

// 比赛排名响应
type ContestRankingResponse struct {
	Ranking *Ranking `json:"ranking,omitempty"`
}

// 比赛排行榜请求
type ContestRanklistRequest struct {
	Contest ContestId `json:"contest"`
}

// 比赛排行榜响应
type ContestRanklistResponse struct {
	Rankings []Ranking `json:"rankings"`
}
