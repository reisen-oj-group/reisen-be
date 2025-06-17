package model

import (
	"time"

	"gorm.io/gorm"
)

// 数据库实体相关类型
type ContestId uint
type ProblemId uint
type SubmissionId uint
type TagId uint
type UserId uint
type TagClassifyId uint

// 配置文件相关类型
type UserLangId string
type CodeLangId string
type VerdictId string
type ProblemLabel string

type BaseModel struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
