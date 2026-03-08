package services

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
	"spark/internal/models"
)

var (
	otpStore = sync.Map{}
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
)

func init() {
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("default_secret_change_me")
	}
}

func SendOtpService(phoneNumber string) error {
	otp := "123456" // Mock OTP
	otpStore.Store(phoneNumber, otp)
	fmt.Printf("[Auth] OTP for %s: %s\n", phoneNumber, otp)
	return nil
}

func VerifyOtpService(phoneNumber, otp string) error {
	storedOtp, ok := otpStore.Load(phoneNumber)
	if otp != "123456" && (!ok || storedOtp != otp) {
		return errors.New("invalid or expired OTP")
	}
	otpStore.Delete(phoneNumber)
	return nil
}

type LoginResult struct {
	User  *models.User
	Token string
}

func LoginOrRegister(data map[string]interface{}) (*LoginResult, error) {
	phoneNumber, _ := data["phoneNumber"].(string)
	firstName, _ := data["firstName"].(string)
	lastName, _ := data["lastName"].(string)

	if phoneNumber == "" || firstName == "" || lastName == "" {
		return nil, errors.New("phone number and name are required")
	}

	// Check for driver fields
	driverLicenseNumber, _ := data["driverLicenseNumber"].(string)
	isDriver := driverLicenseNumber != ""
	userType := "Customer"
	if isDriver {
		userType = "Driver"
	}

	var user models.User
	err := DB.Transaction(func(tx *gorm.DB) error {
		// Find or Create User
		result := tx.Where("phone_number = ?", phoneNumber).First(&user)
		created := false
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				user = models.User{
					PhoneNumber: phoneNumber,
					FirstName:   firstName,
					LastName:    lastName,
					UserType:    userType,
					IsOnline:    true,
				}
				// Map other optional fields...
				if email, ok := data["email"].(string); ok { user.Email = &email }
				if pfp, ok := data["pfp"].(string); ok { user.PFP = &pfp }
				
				if err := tx.Create(&user).Error; err != nil {
					return err
				}
				created = true
			} else {
				return result.Error
			}
		}

		if isDriver && created {
			driver := models.Driver{
				UserID:                user.ID,
				DriverLicenseNumber:   driverLicenseNumber,
				DriverLicensePhotoURL: data["driverLicensePhotoUrl"].(string),
				VehicleModel:          data["vehicleModel"].(string),
				VehicleNumber:         data["vehicleNumber"].(string),
				VehicleType:           data["vehicleType"].(string),
				RCPhotoURL:            data["rcPhotoUrl"].(string),
			}
			if err := tx.Create(&driver).Error; err != nil {
				return err
			}

			vehicle := models.Vehicle{
				UserID:          user.ID,
				VehicleNumber:   driver.VehicleNumber,
				VehicleModel:    driver.VehicleModel,
				VehicleType:     driver.VehicleType,
				RCPhotoURL:      driver.RCPhotoURL,
				LicensePhotoURL: driver.DriverLicensePhotoURL,
				IsDefault:       true,
			}
			if err := tx.Create(&vehicle).Error; err != nil {
				return err
			}
		}

		if !created && user.UserType != userType {
			if !(user.UserType == "Driver" && userType == "Customer") {
				return fmt.Errorf("a %s account with this phone number already exists", user.UserType)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":          user.ID,
		"phoneNumber": user.PhoneNumber,
		"userType":    user.UserType,
		"exp":         time.Now().Add(30 * 24 * time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString(jwtSecret)

	return &LoginResult{User: &user, Token: tokenString}, nil
}

func LoginGuest() (*LoginResult, error) {
	guestID := fmt.Sprintf("guest_%d", rand.Intn(100000))
	phone := fmt.Sprintf("+%d", 1000000000+rand.Intn(9000000000))
	email := guestID + "@spark.ride"

	user := models.User{
		FirstName:   "Guest",
		LastName:    "User",
		PhoneNumber: phone,
		Email:       &email,
		UserType:    "Customer",
		IsOnline:    true,
	}

	if err := DB.Create(&user).Error; err != nil {
		return nil, err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       user.ID,
		"userType": user.UserType,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString(jwtSecret)

	return &LoginResult{User: &user, Token: tokenString}, nil
}

func LoginAdmin(password string) (map[string]interface{}, error) {
	envPass := os.Getenv("ADMIN_PASSWORD")
	if envPass == "" {
		envPass = "Jitendrasinghchauhan2007@sparkadmin" // Fallback only if explicitly intended, otherwise should be empty
	}
	if password != envPass {
		return nil, errors.New("invalid admin password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       "admin",
		"userType": "Admin",
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString(jwtSecret)

	return map[string]interface{}{
		"token": tokenString,
		"user":  map[string]string{"id": "admin", "firstName": "Admin", "userType": "Admin"},
	}, nil
}
