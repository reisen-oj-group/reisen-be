package repository

import (
	"reisen-be/internal/model"

	"gorm.io/gorm"
)

type JudgementRepository struct {
	db *gorm.DB
}

func NewJudgementRepository(db *gorm.DB) *JudgementRepository {
	return &JudgementRepository{db: db}
}

func (r *JudgementRepository) Create(judgement *model.Judgement) error {
	return r.db.Create(judgement).Error
}

func (r *JudgementRepository) Update(judgement *model.Judgement) error {
	return r.db.Save(judgement).Error
}

func (r *JudgementRepository) Delete(problemID model.ProblemId, userID model.UserId) error {
	return r.db.Delete(&model.Judgement{}, problemID, userID).Error
}

func (r *JudgementRepository) GetByID(problemID model.ProblemId, userID model.UserId) (*model.Judgement, error) {
	var judgement model.Judgement
	if err := r.db.First(&judgement, problemID, userID).Error; err != nil {
		return nil, err
	}
	return &judgement, nil
}

func (r *JudgementRepository) GetByUser(userID model.UserId) ([]model.Judgement, error) {
	var judgements []model.Judgement
	err := r.db.Where("user_id = ?", userID).Order("problem_id ASC").Find(&judgements).Error
	return judgements, err
}
