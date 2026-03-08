package repository

import (
	"spark/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ConfigRepository struct {
	DB *gorm.DB
}

func NewConfigRepository(db *gorm.DB) *ConfigRepository {
	return &ConfigRepository{DB: db}
}

func (r *ConfigRepository) GetConfig(key string) (*models.Config, error) {
	var config models.Config
	err := r.DB.Where("key = ?", key).First(&config).Error
	return &config, err
}

func (r *ConfigRepository) UpsertConfig(config *models.Config) error {
	return r.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		UpdateAll: true,
	}).Create(config).Error
}
