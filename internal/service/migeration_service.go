// internal/service/migration_service.go
package service

import (
	"gorm.io/gorm"
)

type GormMigrationService struct {
	db *gorm.DB
}

func NewGormMigrationService(db *gorm.DB) *GormMigrationService {
	return &GormMigrationService{db: db}
}

func (s *GormMigrationService) RunMigrations() error {
	

	return nil
}