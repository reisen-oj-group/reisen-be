package model

import (
	"time"
)

type Role int

const (
	RoleUser  Role = 1
	RoleJury  Role = 2
	RoleAdmin Role = 3
	RoleSuper Role = 4
)

type User struct {
	ID       UserId    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name     string    `gorm:"size:50;not null;unique"  json:"name"`
	Password string    `gorm:"size:100;not null"        json:"-"`
	Role     Role      `gorm:"default:0"                json:"role"`
	Register time.Time `gorm:"autoCreateTime"           json:"register"`
	Avatar   string    `gorm:"size:200"                 json:"avatar"`
}

func (User) TableName() string {
	return "users"
}

// UserFilterParams 记录过滤参数
// User 和 Keyword 不会同时存在，根据传过来的 User 决定是数字还是模糊匹配
type UserFilterParams struct {
	User    *UserId
	Keyword *string
	Role    *Role
}

// UserFilterParamsRaw 记录传递过来的过滤参数
type UserFilterParamsRaw struct {
	User    *string     `json:"user,omitempty"`
	Role    *Role       `json:"role,omitempty"`
}

type UserRequest struct {
	User     UserId   `json:"user"`
}

type UserResponse struct {
	User     User     `json:"user"`
}

// 用户列表请求
type UserListRequest struct {
	UserFilterParamsRaw
	Page int `json:"page"`
}

// 用户列表响应
type UserListResponse struct {
	Total   int           `json:"total"`
	Users   []User        `json:"users"`
}

// 编辑用户请求，若 User.ID 为 0 则新建，User 不包含 Password
// 用户有权限修改名称、密码和头像，
type UserEditRequest struct {
	User     User         `json:"user"`
	Password string       `json:"password,omitempty"`
}

// 编辑用户响应（补全诸如注册时间等信息）
type UserEditResponse struct {
	User     User         `json:"user"`
}

// 删除用户请求
type UserDeleteRequest struct {
	User    UserId        `json:"user"`
}

// 删除用户响应（空，根据状态码确定是否成功）
type UserDeleteResponse struct {
	
}
