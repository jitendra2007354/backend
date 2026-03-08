package repository

import (
	"spark/internal/models"

	"gorm.io/gorm"
)

type WalletRepository struct {
	DB *gorm.DB
}

func NewWalletRepository(db *gorm.DB) *WalletRepository {
	return &WalletRepository{DB: db}
}

// GetUserBalance retrieves the wallet balance from the User model
func (r *WalletRepository) GetUserBalance(userID uint) (float64, error) {
	var user models.User
	if err := r.DB.Select("wallet_balance").First(&user, userID).Error; err != nil {
		return 0, err
	}
	return user.WalletBalance, nil
}

// UpdateUserBalance updates the wallet balance for a user
func (r *WalletRepository) UpdateUserBalance(userID uint, amount float64) error {
	// This adds the amount to the existing balance
	return r.DB.Model(&models.User{}).Where("id = ?", userID).Update("wallet_balance", gorm.Expr("wallet_balance + ?", amount)).Error
}

// UpdateDriverPlatformFee updates the outstanding platform fee for a driver
func (r *WalletRepository) UpdateDriverPlatformFee(driverID uint, amount float64) error {
	// This adds the amount to the existing fee
	return r.DB.Model(&models.Driver{}).Where("id = ?", driverID).Update("outstanding_platform_fee", gorm.Expr("outstanding_platform_fee + ?", amount)).Error
}
