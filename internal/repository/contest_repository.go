package repository

import (
	"reisen-be/internal/model"

	"gorm.io/gorm"
)

type ContestRepository struct {
	db *gorm.DB
}

func NewContestRepository(db *gorm.DB) *ContestRepository {
	return &ContestRepository{db: db}
}

func (r *ContestRepository) Create(contest *model.Contest) error {
	return r.db.Create(contest).Error
}

func (r *ContestRepository) Update(contest *model.Contest) error {
	return r.db.Save(contest).Error
}

func (r *ContestRepository) GetByID(id model.ContestId) (*model.Contest, error) {
	var contest model.Contest
	if err := r.db.First(&contest, id).Error; err != nil {
		return nil, err
	}
	return &contest, nil
}

func (r *ContestRepository) Delete(contestID model.ContestId) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 先删除相关报名信息
		if err := tx.Where("contest_id = ?", contestID).Delete(&model.Signup{}).Error; err != nil {
			return err
		}
		// 再删除排名信息
		if err := tx.Where("contest_id = ?", contestID).Delete(&model.Ranking{}).Error; err != nil {
			return err
		}
		// 最后删除比赛
		return tx.Delete(&model.Contest{}, contestID).Error
	})
}

func (r *ContestRepository) Register(userID model.UserId, contestID model.ContestId) error {
	registration := model.Signup{
		ContestID: contestID,
		UserID:    userID,
	}
	return r.db.Create(&registration).Error
}

func (r *ContestRepository) Unregister(userID model.UserId, contestID model.ContestId) error {
	return r.db.Where("contest_id = ? AND user_id = ?", contestID, userID).
		Delete(&model.Signup{}).Error
}

func (r *ContestRepository) IsRegistered(userID model.UserId, contestID model.ContestId) (bool, error) {
	var count int64
	err := r.db.Model(&model.Signup{}).
		Where("contest_id = ? AND user_id = ?", contestID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *ContestRepository) GetSignups(contestID model.ContestId) ([]model.Signup, error) {
	var registrations []model.Signup
	err := r.db.Where("contest_id = ?", contestID).Find(&registrations).Error
	return registrations, err
}

func (r *ContestRepository) UpdateRanking(ranking *model.Ranking) error {
	return r.db.Save(ranking).Error
}

func (r *ContestRepository) GetRanking(contestID model.ContestId, userID model.UserId) (*model.Ranking, error) {
	var ranking model.Ranking
	err := r.db.Where("contest_id = ? AND user_id = ?", contestID, userID).First(&ranking).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &ranking, nil
}

func (r *ContestRepository) ListRankings(contestID model.ContestId, page, pageSize int) ([]model.Ranking, int64, error) {
	var rankings []model.Ranking
	var total int64

	// 获取总数
	if err := r.db.Model(&model.Ranking{}).
		Where("contest_id = ?", contestID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := r.db.Where("contest_id = ?", contestID).
		Order("rank ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&rankings).Error; err != nil {
		return nil, 0, err
	}

	return rankings, total, nil
}

func (r *ContestRepository) List(filter *model.ContestFilter, page, pageSize int) ([]model.Contest, int64, error) {
	var contests []model.Contest
	var total int64

	query := r.db.Model(&model.Contest{})

	// 应用过滤条件
	if filter != nil {
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
		if filter.Rule != nil {
			query = query.Where("rule = ?", *filter.Rule)
		}
		if filter.Keyword != nil {
			query = query.Where("title LIKE ?", "%"+*filter.Keyword+"%")
		}
		if filter.Before != nil {
			query = query.Where("start_time <= ?", *filter.Before)
		}
		if filter.After != nil {
			query = query.Where("start_time >= ?", *filter.After)
		}
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := query.Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("start_time DESC").
		Find(&contests).Error; err != nil {
		return nil, 0, err
	}
	return contests, total, nil
}