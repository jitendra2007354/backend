package repository

import (
	"spark/internal/models"

	"gorm.io/gorm"
)

type DriverRepository struct {
	DB *gorm.DB
}

func NewDriverRepository(db *gorm.DB) *DriverRepository {
	return &DriverRepository{DB: db}
}

func (r *DriverRepository) FindByUserID(userID uint) (*models.Driver, error) {
	var driver models.Driver
	err := r.DB.Where("user_id = ?", userID).First(&driver).Error
	return &driver, err
}

func (r *DriverRepository) FindByID(id uint) (*models.Driver, error) {
	var driver models.Driver
	err := r.DB.First(&driver, id).Error
	return &driver, err
}

func (r *DriverRepository) Update(driver *models.Driver) error {
	return r.DB.Save(driver).Error
}

func (r *DriverRepository) FindNearby(lat, lng float64, radius float64, vehicleType string) ([]models.Driver, error) {
	// Placeholder for spatial query logic
	return []models.Driver{}, nil
}
