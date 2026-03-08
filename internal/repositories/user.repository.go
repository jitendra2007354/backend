package repository

import (
	"spark/internal/models"

	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) Create(user *models.User) error {
	return r.DB.Create(user).Error
}

func (r *UserRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.DB.First(&user, id).Error
	return &user, err
}

func (r *UserRepository) FindByPhone(phone string) (*models.User, error) {
	var user models.User
	err := r.DB.Where("phone_number = ?", phone).First(&user).Error
	return &user, err
}

func (r *UserRepository) Update(user *models.User) error {
	return r.DB.Save(user).Error
}

func (r *UserRepository) FindByRole(role string) ([]models.User, error) {
	var users []models.User
	err := r.DB.Where("user_type = ?", role).Find(&users).Error
	return users, err
}

func (r *UserRepository) SetOnlineStatus(userID uint, isOnline bool) error {
	return r.DB.Model(&models.User{}).Where("id = ?", userID).Update("is_online", isOnline).Error
}
