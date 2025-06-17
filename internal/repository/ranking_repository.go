package repository

import (
	"reisen-be/internal/model"
	"gorm.io/gorm"
)

type RankingRepository struct {
	db *gorm.DB
}

func NewRankingRepository(db *gorm.DB) *RankingRepository {
	return &RankingRepository{db: db}
}

func (r *RankingRepository) Create(ranking *model.Ranking) error {
	return r.db.Create(ranking).Error
}

func (r *RankingRepository) Update(ranking *model.Ranking) error {
	return r.db.Save(ranking).Error
}

func (r *RankingRepository) Delete(contestID model.ContestId, userID model.UserId) error {
	return r.db.Delete(&model.Ranking{}, contestID, userID).Error
}

func (r *RankingRepository) GetByID(contestID model.ContestId, userID model.UserId) (*model.Ranking, error) {
	var ranking model.Ranking
	if err := r.db.First(&ranking, contestID, userID).Error; err != nil {
		return nil, err
	}
	return &ranking, nil
}

func (r *RankingRepository) GetByUser(userID model.UserId) ([]model.Ranking, error) {
	var rankings []model.Ranking
	err := r.db.Where("user_id = ?", userID).Order("contest_id ASC").Find(&rankings).Error
	return rankings, err
}

func (r *RankingRepository) GetByContest(contestID model.ContestId) ([]model.Ranking, error) {
	var rankings []model.Ranking
	err := r.db.Where("contest_id = ?", contestID).Order("ranking ASC").Find(&rankings).Error
	return rankings, err
}
