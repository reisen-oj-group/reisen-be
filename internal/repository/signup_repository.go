package repository

import (
	"reisen-be/internal/model"
	"time"

	"gorm.io/gorm"
)

type SignupRepository struct {
	db *gorm.DB
}

func NewSignupRepository(db *gorm.DB) *SignupRepository {
	return &SignupRepository{db: db}
}

func (r *SignupRepository) Signup(userID model.UserId, contestID model.ContestId) error {
	signup := model.Signup{
		ContestID: contestID,
		UserID:    userID,
		Register:  time.Now(),
	}
	return r.db.Create(&signup).Error
}

func (r *SignupRepository) Signout(userID model.UserId, contestID model.ContestId) error {
	return r.db.Where("contest_id = ? AND user_id = ?", contestID, userID).
		Delete(&model.Signup{}).Error
}

func (r *SignupRepository) GetSignup(userID model.UserId, contestID model.ContestId) (*model.Signup, error) {
	var signup model.Signup
	if err := r.db.Where("contest_id = ? AND user_id = ?", contestID, userID).
		Take(&signup).Error; err != nil {
		return nil, err
	}
	return &signup, nil
}

func (r *SignupRepository) GetSignupCount(contestID model.ContestId) (int64, error) {
	var count int64
	err := r.db.Model(&model.Signup{}).
		Where("contest_id = ?", contestID).
		Count(&count).Error
	return count, err
}

func (r *SignupRepository) GetSignupsAll(contestID model.ContestId) ([]model.Signup, error) {
	var signups []model.Signup

	// 分页查询
	if err := r.db.Where("contest_id = ?", contestID).
		Order("register ASC").
		Find(&signups).Error; err != nil {
		return nil, err
	}
	return signups, nil
}

func (r *SignupRepository) GetSignups(contestID model.ContestId, page, pageSize int) ([]model.Signup, int64, error) {
	var signups []model.Signup
	var total int64

	// 获取总数
	if err := r.db.Model(&model.Signup{}).
		Where("contest_id = ?", contestID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := r.db.Where("contest_id = ?", contestID).
		Order("register ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&signups).Error; err != nil {
		return nil, 0, err
	}

	return signups, total, nil
}

func (r *SignupRepository) GetUserRegisteredContests(userID model.UserId, page, pageSize int) ([]model.Contest, int64, error) {
	var contests []model.Contest
	var total int64

	// 获取总数
	if err := r.db.Model(&model.Signup{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := r.db.Model(&model.Contest{}).
		Joins("JOIN signups ON signups.contest_id = contests.id").
		Where("signups.user_id = ?", userID).
		Order("contests.start_time DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&contests).Error; err != nil {
		return nil, 0, err
	}
	return contests, total, nil
}
