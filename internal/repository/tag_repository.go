package repository

import (
	"reisen-be/internal/model"

	"gorm.io/gorm"
)

type TagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) GetClassifies() ([]model.TagClassify, error) {
	var classifies []model.TagClassify
	err := r.db.Find(&classifies).Error
	return classifies, err
}

func (r *TagRepository) GetTags(classifyID *model.TagClassifyId) ([]model.Tag, error) {
	var tags []model.Tag
	query := r.db.Model(&model.Tag{})
	if classifyID != nil {
		query = query.Where("classify = ?", *classifyID)
	}
	err := query.Find(&tags).Error
	return tags, err
}

func (r *TagRepository) GetTagByID(id model.TagId) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.First(&tag, id).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *TagRepository) CreateTag(tag *model.Tag) error {
	return r.db.Create(tag).Error
}

func (r *TagRepository) UpdateTag(tag *model.Tag) error {
	return r.db.Save(tag).Error
}

func (r *TagRepository) DeleteTag(id model.TagId) error {
	return r.db.Delete(&model.Tag{}, id).Error
}