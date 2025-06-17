package repository

import (
	"reisen-be/internal/model"

	"gorm.io/gorm"
)

type ProblemRepository struct {
	db *gorm.DB
}

func NewProblemRepository(db *gorm.DB) *ProblemRepository {
	return &ProblemRepository{db: db}
}

func (r *ProblemRepository) Create(problem *model.Problem) error {
	return r.db.Create(problem).Error;
}

func (r *ProblemRepository) Update(problem *model.Problem) error {
	return r.db.Save(problem).Error;
}

func (r *ProblemRepository) GetByID(id model.ProblemId) (*model.Problem, error) {
	var problem model.Problem
	if err := r.db.First(&problem, id).Error; err != nil {
		return nil, err
	}
	return &problem, nil
}

func (r *ProblemRepository) List(filter *model.ProblemFilter, page, pageSize int) ([]model.ProblemCore, int64, error) {
	var problems []model.ProblemCore
	var total int64

	query := r.db.Model(&model.ProblemCore{})

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

	return problems, total, nil
}

func (r *ProblemRepository) Delete(problemID model.ProblemId) error {
	return r.db.Delete(&model.Problem{}, problemID).Error
}

func (r *ProblemRepository) IncreaseSubmitTotal(problemID model.ProblemId) error {
	return r.db.
		Model(&model.Problem{}).
		Where("id = ?", problemID).
		Update("count_total", gorm.Expr("count_total + ?", 1)).
		Error
}

func (r *ProblemRepository) IncreaseSubmitCorrect(problemID model.ProblemId) error {
	return r.db.
		Model(&model.Problem{}).
		Where("id = ?", problemID).
		Update("count_correct", gorm.Expr("count_correct + ?", 1)).
		Error
}

func (r *ProblemRepository) UpdateTestdataStatus(problemID model.ProblemId, hasData, hasConfig bool) error {
	return r.db.Model(&model.Problem{}).
		Where("id = ?", problemID).
		Updates(map[string]interface{}{
			"has_testdata": hasData,
			"has_config":   hasConfig,
		}).Error
}

func (r *ProblemRepository) GetTestdataStatus(problemID model.ProblemId) (bool, bool, error) {
	var problem model.Problem
	if err := r.db.Select("has_testdata, has_config").
		First(&problem, problemID).Error; err != nil {
		return false, false, err
	}
	return problem.HasTestdata, problem.HasConfig, nil
}
