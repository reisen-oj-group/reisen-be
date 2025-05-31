package service

import (
	"errors"
	"reisen-be/internal/model"
	"reisen-be/internal/repository"
	"strconv"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// 查询用户信息，用于访问用户主页
func (s *UserService) GetUser(userID model.UserId) (*model.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	return user, nil
}

func (s *UserService) EditUser(profile *model.User) (*model.User, error) {
	user, err := s.userRepo.GetByID(profile.ID)
	if err != nil {
		return nil, err
	}
	user.Name = profile.Name
	user.Avatar = profile.Avatar
	return user, s.userRepo.Update(user)
}

func (s *UserService) DeleteUser(userID model.UserId) error {
	return s.userRepo.Delete(userID)
}

// 查询用户列表，用于管理后台管理
func (s *UserService) ListUsers(filterRaw *model.UserFilterParamsRaw, page, pageSize int) ([]model.User, int64, error) {
	filter, err := s.ConvertFilterParamsRaw(filterRaw)
	if err != nil {
		return nil, 0, err
	}
	users, total, err := s.userRepo.List(filter, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// 处理原始筛选条件（将 user 字段解析成用户 ID 或是用户名）
func (s *UserService) ConvertFilterParamsRaw(raw *model.UserFilterParamsRaw) (*model.UserFilterParams, error) {
	if raw == nil {
		return nil, nil
	}

	params := &model.UserFilterParams{
			Role: raw.Role,
	}

	// 处理 User 字段转换
	if raw.User != nil {
		// 尝试解析为数字 ID
		if userID, err := strconv.Atoi(*raw.User); err == nil {
			params.User = (*model.UserId)(&userID)
		} else {
			// 如果是字符串，进行模糊匹配
			params.Keyword = raw.User
		}
	}
	return params, nil
}
