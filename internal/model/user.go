package model

import (
	"time"
	"gorm.io/gorm"
)

type Role int

const (
	RoleUser  Role = 0
	RoleJury  Role = 1
	RoleAdmin Role = 2
)

type User struct {
	gorm.Model
	ID       UserId `gorm:"primaryKey;autoIncrement" json:"id"`
	Name     string `gorm:"size:50;not null;unique" json:"name"`
	Password string `gorm:"size:100;not null"`
	Role     Role   `gorm:"default:0" json:"role"`
	Register time.Time ` json:"register"`
	Avatar   string `gorm:"size:200"  json:"avatar"`
}

func (User) TableName() string {
	return "users"
}