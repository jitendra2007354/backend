package repository

import (
	"spark/internal/models"

	"gorm.io/gorm"
)

type VehicleRepository struct {
	DB *gorm.DB
}

func NewVehicleRepository(db *gorm.DB) *VehicleRepository {
	return &VehicleRepository{DB: db}
}

func (r *VehicleRepository) Create(vehicle *models.Vehicle) error {
	return r.DB.Create(vehicle).Error
}

func (r *VehicleRepository) FindByDriverID(driverID uint) ([]models.Vehicle, error) {
	var vehicles []models.Vehicle
	err := r.DB.Where("user_id = ? AND is_deleted = ?", driverID, false).Find(&vehicles).Error
	return vehicles, err
}

func (r *VehicleRepository) FindByID(id uint) (*models.Vehicle, error) {
	var vehicle models.Vehicle
	err := r.DB.Where("id = ? AND is_deleted = ?", id, false).First(&vehicle).Error
	return &vehicle, err
}

func (r *VehicleRepository) SetDefault(driverID, vehicleID uint) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		// Unset current default
		if err := tx.Model(&models.Vehicle{}).Where("user_id = ?", driverID).Update("is_default", false).Error; err != nil {
			return err
		}
		// Set new default
		return tx.Model(&models.Vehicle{}).Where("id = ? AND user_id = ?", vehicleID, driverID).Update("is_default", true).Error
	})
}

func (r *VehicleRepository) SoftDelete(id uint) error {
	return r.DB.Model(&models.Vehicle{}).Where("id = ?", id).Update("is_deleted", true).Error
}
