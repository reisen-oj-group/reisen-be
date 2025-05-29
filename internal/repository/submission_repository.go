package repository

import (
	"reisen-be/internal/model"
	// "time"

	"gorm.io/gorm"
)

type SubmissionRepository struct {
	db *gorm.DB
}

func NewSubmissionRepository(db *gorm.DB) *SubmissionRepository {
	return &SubmissionRepository{db: db}
}

func (r *SubmissionRepository) Create(submission *model.Submission) error {
	return r.db.Create(submission).Error
}

func (r *SubmissionRepository) Update(submission *model.Submission) error {
	return r.db.Save(submission).Error
}

func (r *SubmissionRepository) GetByID(id int64) (*model.Submission, error) {
	var submission model.Submission
	if err := r.db.First(&submission, id).Error; err != nil {
		return nil, err
	}
	return &submission, nil
}

func (r *SubmissionRepository) List(filter *model.RecordFilterParams, page, pageSize int) ([]model.Submission, int64, error) {
	var submissions []model.Submission
	var total int64

	query := r.db.Model(&model.Submission{})

	// 应用过滤条件
	if filter != nil {
		if filter.User != nil {
			query = query.Where("user_id = ?", *filter.User)
		}
		if filter.Problem != nil {
			query = query.Where("problem_id = ?", *filter.Problem)
		}
		if filter.Lang != nil {
			query = query.Where("lang = ?", *filter.Lang)
		}
		if filter.Verdict != nil {
			query = query.Where("verdict = ?", *filter.Verdict)
		}
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := query.Order("submission_time DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&submissions).Error; err != nil {
		return nil, 0, err
	}

	return submissions, total, nil
}