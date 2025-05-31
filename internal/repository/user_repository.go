package repository

import (
	"reisen-be/internal/model"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("name = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) Update(user *model.User) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 更新基础字段
		if err := tx.Model(user).Updates(user).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *UserRepository) GetByID(id model.UserId) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) List(filter *model.UserFilterParams, page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := r.db.Model(&model.User{})

	// 应用过滤条件
	if filter != nil {
		if filter.User != nil {
			query = query.Where("id = ?", *filter.User)
		} else
		if filter.Keyword != nil {
			query = query.Where("name LIKE ?", "%" + *filter.Keyword + "%")
		}
		if filter.Role != nil {
			query = query.Where("role = ?", *filter.Role)
		}
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := query.Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *UserRepository) UpdatePassword(userID model.UserId, hashedPwd string) error {
	return r.db.Model(&model.User{}).
		Where("id = ?", userID).
		Update("password", hashedPwd).Error
}

func (r *UserRepository) Delete(userID model.UserId) error {
	return r.db.Delete(&model.User{}, userID).Error
}
