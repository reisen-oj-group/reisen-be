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
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(problem).Error; err != nil {
			return err
		}
		
		// 处理标签关联
		if len(problem.Tags) > 0 {
			if err := tx.Model(problem).Association("Tags").Replace(problem.Tags); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ProblemRepository) Update(problem *model.Problem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 更新基础字段
		if err := tx.Model(problem).Updates(problem).Error; err != nil {
			return err
		}
		
		// 更新标签关联
		if err := tx.Model(problem).Association("Tags").Replace(problem.Tags); err != nil {
			return err
		}
		return nil
	})
}

func (r *ProblemRepository) GetByID(id model.ProblemId) (*model.Problem, error) {
	var problem model.Problem
	if err := r.db.Preload("Tags").First(&problem, id).Error; err != nil {
		return nil, err
	}
	return &problem, nil
}

func (r *ProblemRepository) List(filter *model.ProblemFilter, page, pageSize int) ([]model.ProblemCore, int64, error) {
	var problems []model.Problem
	var total int64

	query := r.db.Model(&model.Problem{})

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

	// 转换为ProblemCore格式
	cores := make([]model.ProblemCore, 0)
	for _, p := range problems {
		var tagIDs []model.TagId
		for _, tag := range p.Tags {
			tagIDs = append(tagIDs, tag.TagID)
		}
		
		cores = append(cores, model.ProblemCore{
			ID:           p.ID,
			Type:         p.Type,
			Status:       p.Status,
			LimitTime:    p.LimitTime,
			LimitMemory:  p.LimitMemory,
			CountCorrect: p.CountCorrect,
			CountTotal:   p.CountTotal,
			Difficulty:   p.Difficulty,
			Title:        p.Title,
			Tags:         tagIDs,
		})
	}

	return cores, total, nil
}

func (r *ProblemRepository) Delete(problemID model.ProblemId) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 先删除标签关联
		if err := tx.Where("problem_id = ?", problemID).Delete(&model.ProblemTag{}).Error; err != nil {
			return err
		}
		// 再删除题目
		return tx.Delete(&model.Problem{}, problemID).Error
	})
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