package repository

import (
	"spark/internal/models"

	"gorm.io/gorm"
)

type RideRepository struct {
	DB *gorm.DB
}

func NewRideRepository(db *gorm.DB) *RideRepository {
	return &RideRepository{DB: db}
}

func (r *RideRepository) Create(ride *models.Ride) error {
	return r.DB.Create(ride).Error
}

func (r *RideRepository) FindByID(id uint) (*models.Ride, error) {
	var ride models.Ride
	err := r.DB.First(&ride, id).Error
	return &ride, err
}

func (r *RideRepository) FindWithDetails(id uint) (*models.Ride, error) {
	var ride models.Ride
	err := r.DB.Preload("Customer").Preload("Driver.User").First(&ride, id).Error
	return &ride, err
}

func (r *RideRepository) FindPending() ([]models.Ride, error) {
	var rides []models.Ride
	err := r.DB.Where("status = ?", "pending").Preload("Customer").Order("created_at DESC").Find(&rides).Error
	return rides, err
}

func (r *RideRepository) UpdateStatus(id uint, status string) error {
	return r.DB.Model(&models.Ride{}).Where("id = ?", id).Update("status", status).Error
}

func (r *RideRepository) FindByCustomer(customerID uint) ([]models.Ride, error) {
	var rides []models.Ride
	err := r.DB.Where("customer_id = ?", customerID).Order("created_at DESC").Find(&rides).Error
	return rides, err
}

func (r *RideRepository) Save(ride *models.Ride) error {
	return r.DB.Save(ride).Error
}
