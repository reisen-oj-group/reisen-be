package query

import (
	"reisen-be/internal/model"

	"gorm.io/gorm"
)

type ProblemListQuery struct {
	db *gorm.DB
}

func NewProblemListQuery(db *gorm.DB) *ProblemListQuery {
	return &ProblemListQuery{db: db}
}

func (r *ProblemListQuery) List(filter *model.ProblemFilter, userID *model.UserId, page, pageSize int) ([]model.ProblemCoreWithJudgements, int64, error) {
	var problems []model.ProblemCoreWithJudgements
	var total int64

	query := r.db.Model(&model.ProblemCoreWithJudgements{})

	if userID != nil {
		query = query.Preload("Judgements", "user_id = ?", userID)
	}

	// 应用过滤条件
	if filter != nil {
		if filter.MinDifficulty != nil && *filter.MinDifficulty > 0 {
			query = query.Where("difficulty >= ?", filter.MinDifficulty)
		}
		if filter.MaxDifficulty != nil && *filter.MaxDifficulty > 0 {
			query = query.Where("difficulty <= ?", filter.MaxDifficulty)
		}
		if filter.Provider != nil && *filter.Provider > 0 {
			query = query.Where("provider = ?", filter.Provider)
		}
		if len(filter.Tags) > 0 {
			query = query.Joins("JOIN problem_tags ON problem_tags.problem_id = problems.id").
				Where("problem_tags.tag_id IN ?", filter.Tags)
		}
		if filter.Keywords != nil && *filter.Keywords != "" {
			query = query.Where("JSON_SEARCH(title, 'one', ?) IS NOT NULL", "%"+*filter.Keywords+"%")
		}
		if filter.Status != nil && *filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := query.Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&problems).Error; err != nil {
		return nil, 0, err
	}

	for i := range problems {
		if problems[i].Judgements == nil {
			problems[i].Judgements = make([]model.Judgement, 0)
		}
	}

	return problems, total, nil
}
