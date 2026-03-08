package services

import (
	"errors"
	"spark/internal/models"
)

func GetConfig(key string) (*models.Config, error) {
	var config models.Config
	if err := DB.Where("key = ?", key).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func GetApplicableConfig() (*models.Config, error) {
	return GetConfig("global")
}

func SetConfig(data map[string]interface{}) error {
	_, ok := data["key"].(string)
	if !ok {
		return errors.New("key required")
	}
	// Upsert logic
	// DB.Clauses(clause.OnConflict{UpdateAll: true}).Create(&config)
	return nil
}
