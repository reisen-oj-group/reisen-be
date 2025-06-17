package query

import (
	"reisen-be/internal/model"

	"gorm.io/gorm"
)

type ContestListQuery struct {
	db *gorm.DB
}

func NewContestListQuery(db *gorm.DB) *ContestListQuery {
	return &ContestListQuery{db: db}
}

func (q *ContestListQuery) List(filter *model.ContestFilter, userID *model.UserId, page, pageSize int) ([]model.ContestWithSignups, int64, error) {
	var contests []model.ContestWithSignups
	var total int64

	query := q.db.Model(&model.ContestWithSignups{})

	if userID != nil {
		query = query.Preload("Signups", "user_id = ?", userID)
	}

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
