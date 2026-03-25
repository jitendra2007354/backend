package services

import (
	cryptorand "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"spark/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

var (
	sessionStore = sync.Map{} // To store session IDs (JWTs)
)

func getJWTSecret() []byte {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default_secret_change_me"
	}
	return []byte(jwtSecret)
}

type LoginResult struct {
	User  *models.User
	Token string
}

// generateSessionID generates a random session ID.
func generateSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := cryptorand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}

type SessionData struct {
	UserID    uint
	SessionID string
	LoginTime time.Time
	// Other session-related data
}

// getStr is a safe helper to extract a string from the map.
// It returns an empty string if the key is missing or the value is not a string.
func getStr(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

// getStrPtr is a safe helper that returns a *string, or nil if the string is empty.
// This is useful for model fields that are pointers to strings.
func getStrPtr(data map[string]interface{}, key string) *string {
	val := getStr(data, key)
	if val == "" {
		return nil
	}
	return &val
}

// Deprecated: replaced by the  customerLoginService and DriverLoginService
func LoginOrRegister(data map[string]interface{}) (*LoginResult, error) {
	// --- 1. Safely extract and validate required data ---
	fmt.Printf("Raw Data received in LoginOrRegister: %+v\n", data)

	for key, value := range data {
		fmt.Printf("Key: %s, Value: %v, Type: %T\n", key, value, value)
	}

	action := getStr(data, "action")
	phoneNumber := getStr(data, "phoneNumber")

	if phoneNumber == "" {
		return nil, errors.New("phone number is required")
	}

	firstName := getStr(data, "firstName")
	lastName := getStr(data, "lastName")

	if action != "login" && (firstName == "" || lastName == "") {
		return nil, errors.New("first name and last name are required")
	}

	driverLicenseNumber := getStr(data, "driverLicenseNumber")
	fmt.Printf("DEBUG: driverLicenseNumber value: %s\n", driverLicenseNumber)
	if driverLicenseNumber != "" {
		vehicleModel := getStr(data, "vehicleModel")
		vehicleNumber := getStr(data, "vehicleNumber")
		vehicleType := getStr(data, "vehicleType")
		rcNumber := getStr(data, "rcNumber")
		fmt.Printf("DEBUG: vehicle information vehicleModel: %s, vehicleNumber: %s, vehicleType: %s, rcNumber: %s\n", vehicleModel, vehicleNumber, vehicleType, rcNumber)
	}

	isDriver := driverLicenseNumber != "" || getStr(data, "userType") == "Driver"
	userType := "Customer"
	if isDriver {
		userType = "Driver"
	}

	// --- 2. Perform database operations in a transaction ---
	var user models.User
	err := DB.Transaction(func(tx *gorm.DB) error {

		// Find existing user strictly by phone number.
		// Using OR email causes different phone numbers to merge into the same account during testing!
		result := tx.Where("phone_number = ?", phoneNumber).First(&user)

		if action == "login" {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return errors.New("account not found, please sign up")
			} else if result.Error != nil {
				return result.Error
			}
			if isDriver && user.UserType != "Driver" {
				return errors.New("account not registered as a partner")
			}

			// Refresh session
			sessionID, err := generateSessionID()
			if err == nil {
				sessionStore.Store(user.ID, SessionData{UserID: user.ID, SessionID: sessionID, LoginTime: time.Now()})
			}
			return nil
		}

		isNewUser := false
		if result.Error == nil {
			// Existing user found, refresh session
			sessionID, err := generateSessionID()
			if err != nil {
				return fmt.Errorf("failed to generate session ID: %w", err)
			}
			sessionStore.Store(user.ID, SessionData{UserID: user.ID, SessionID: sessionID, LoginTime: time.Now()})
		} else if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			isNewUser = true
		} else {
			return result.Error
		}

		if isNewUser {
			user = models.User{
				PhoneNumber: phoneNumber,
				FirstName:   firstName,
				LastName:    lastName,
				UserType:    userType,
				IsOnline:    true,
				Email:       getStrPtr(data, "email"),
				PFP:         getStrPtr(data, "pfpUrl"),
				City:        getStrPtr(data, "city"),
				State:       getStrPtr(data, "state"),
			}

			if err := tx.Create(&user).Error; err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
		}

		// Update existing user information
		user.FirstName = firstName
		user.LastName = lastName
		if emailPtr := getStrPtr(data, "email"); emailPtr != nil {
			user.Email = emailPtr
		}
		if pfpPtr := getStrPtr(data, "pfpUrl"); pfpPtr != nil {
			user.PFP = pfpPtr
		}
		if cityPtr := getStrPtr(data, "city"); cityPtr != nil {
			user.City = cityPtr
		}
		if statePtr := getStrPtr(data, "state"); statePtr != nil {
			user.State = statePtr
		}
		if isDriver && user.UserType != "Driver" {
			user.UserType = "Driver"
		}

		if err := tx.Save(&user).Error; err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		// If the user is a driver (based on input), create the associated Driver and Vehicle records if not already present.
		if isDriver && userType == "Driver" {
			// Validate required driver fields
			vehicleModel := strings.TrimSpace(getStr(data, "vehicleModel"))
			vehicleNumber := strings.TrimSpace(getStr(data, "vehicleNumber"))
			vehicleType := strings.TrimSpace(getStr(data, "vehicleType"))
			rcNumber := strings.TrimSpace(getStr(data, "rcNumber"))
			dlPhoto := strings.TrimSpace(getStr(data, "driverLicensePhotoUrl"))
			rcPhoto := strings.TrimSpace(getStr(data, "rcPhotoUrl"))

			if driverLicenseNumber == "" || vehicleModel == "" || vehicleNumber == "" || vehicleType == "" || rcNumber == "" {
				return errors.New("missing required driver information: license number, vehicle model/number, vehicle type and RC number are all required")
			}

			var existingDriver models.Driver
			driverErr := tx.Where("user_id = ?", user.ID).First(&existingDriver).Error
			if errors.Is(driverErr, gorm.ErrRecordNotFound) {
				driver := models.Driver{
					UserID:                user.ID,
					DriverLicensePhotoURL: dlPhoto,
					RCPhotoURL:            rcPhoto,
					DriverLicenseNumber:   driverLicenseNumber,
					VehicleNumber:         vehicleNumber,
					VehicleModel:          vehicleModel,
					VehicleType:           vehicleType,
					RCNumber:              rcNumber,
				}
				if err := tx.Create(&driver).Error; err != nil {
					return fmt.Errorf("failed to create driver record: %w", err)
				}

				vehicle := models.Vehicle{
					DriverID:        driver.ID,
					UserID:          user.ID,
					VehicleNumber:   vehicleNumber,
					VehicleModel:    vehicleModel,
					VehicleType:     vehicleType,
					RCNumber:        rcNumber,
					RCPhotoURL:      rcPhoto,
					LicensePhotoURL: dlPhoto,
					IsDefault:       true,
				}
				if err := tx.Create(&vehicle).Error; err != nil {
					return fmt.Errorf("failed to create vehicle record: %w", err)
				}
			} else if driverErr != nil {
				return driverErr
			} else {
				// Driver and/or Vehicle may already exist; ensure the vehicle is present.
				var existingVehicle models.Vehicle
				vehicleErr := tx.Where("user_id = ?", user.ID).First(&existingVehicle).Error
				if errors.Is(vehicleErr, gorm.ErrRecordNotFound) {
					vehicle := models.Vehicle{
						DriverID:        existingDriver.ID,
						UserID:          user.ID,
						VehicleNumber:   vehicleNumber,
						VehicleModel:    vehicleModel,
						VehicleType:     vehicleType,
						RCNumber:        rcNumber,
						RCPhotoURL:      rcPhoto,
						LicensePhotoURL: dlPhoto,
						IsDefault:       true,
					}
					if err := tx.Create(&vehicle).Error; err != nil {
						return fmt.Errorf("failed to create vehicle record: %w", err)
					}
				} else if vehicleErr != nil {
					return vehicleErr
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// --- 3. Generate JWT Token ---
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":          user.ID,
		"phoneNumber": user.PhoneNumber,
		"userType":    user.UserType,
		"exp":         time.Now().Add(30 * 24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString(getJWTSecret())
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

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
	tokenString, err := token.SignedString(getJWTSecret())
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

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
	tokenString, _ := token.SignedString(getJWTSecret())

	return map[string]interface{}{
		"token": tokenString,
		"user":  map[string]string{"id": "admin", "firstName": "Admin", "userType": "Admin"},
	}, nil
}

func CustomerLoginService(data map[string]interface{}) (*LoginResult, error) {
	// Internal redirect: if driver info is present, use DriverLoginService to preserve Driver user type.
	if dl, ok := data["driverLicenseNumber"].(string); ok && dl != "" {
		return DriverLoginService(data)
	}
	if vn, ok := data["vehicleNumber"].(string); ok && vn != "" {
		return DriverLoginService(data)
	}

	// --- 1. Safely extract and validate required data ---
	fmt.Printf("Raw Data received in CustomerLoginService: %+v\n", data)

	action := getStr(data, "action")
	phoneNumber := getStr(data, "phoneNumber")
	firstName := getStr(data, "firstName")
	lastName := getStr(data, "lastName")

	if phoneNumber == "" {
		return nil, errors.New("phone number is required")
	}
	if action != "login" && (firstName == "" || lastName == "") {
		return nil, errors.New("first name and last name are required")
	}

	// --- 2. Perform database operations in a transaction ---
	var user models.User
	err := DB.Transaction(func(tx *gorm.DB) error {

		// Find existing user strictly by phone number
		result := tx.Where("phone_number = ?", phoneNumber).First(&user)

		if action == "login" {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return errors.New("account not found, please sign up")
			} else if result.Error != nil {
				return result.Error
			}
			return nil
		}

		if result.Error != nil {
			// Check if the error is because the record wasn't found
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				// User not found, prepare to create a new one.
				user = models.User{
					PhoneNumber: phoneNumber,
					FirstName:   firstName,
					LastName:    lastName,
					UserType:    "Customer", // Hardcode Customer
					IsOnline:    true,
					Email:       getStrPtr(data, "email"),
					PFP:         getStrPtr(data, "pfpUrl"),
					City:        getStrPtr(data, "city"),
					State:       getStrPtr(data, "state"),
				}

				if err := tx.Create(&user).Error; err != nil {
					return fmt.Errorf("failed to create user: %w", err)
				}
			} else {
				return result.Error
			}
		} else {
			// Existing user found, update fields if changed
			user.FirstName = firstName
			user.LastName = lastName
			if emailPtr := getStrPtr(data, "email"); emailPtr != nil {
				user.Email = emailPtr
			}
			if pfpPtr := getStrPtr(data, "pfpUrl"); pfpPtr != nil {
				user.PFP = pfpPtr
			}
			if cityPtr := getStrPtr(data, "city"); cityPtr != nil {
				user.City = cityPtr
			}
			if statePtr := getStrPtr(data, "state"); statePtr != nil {
				user.State = statePtr
			}

			if err := tx.Save(&user).Error; err != nil {
				return fmt.Errorf("failed to update user: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// --- 3. Generate JWT Token ---
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":          user.ID,
		"phoneNumber": user.PhoneNumber,
		"userType":    user.UserType,
		"exp":         time.Now().Add(30 * 24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString(getJWTSecret())
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	return &LoginResult{User: &user, Token: tokenString}, nil
}

func DriverLoginService(data map[string]interface{}) (*LoginResult, error) {
	return LoginOrRegister(data)
}

func UpdateDriverProfile(userID uint, data map[string]interface{}) (*models.User, error) {
	var user models.User
	if err := DB.First(&user, userID).Error; err != nil {
		return nil, errors.New("user not found")
	}

	// This function can be expanded to update other profile fields.
	// For now, it only handles 'activeVehicleTypes'.
	if types, ok := data["activeVehicleTypes"]; ok {
		// We assume the 'User' model has a field like 'ActiveVehicleTypes datatypes.JSON'
		// and the database column can store JSON (like TEXT or JSON type).
		jsonTypes, err := json.Marshal(types)
		if err != nil {
			return nil, errors.New("invalid format for activeVehicleTypes")
		}

		// Using .Model().Update() is safer for partial updates.
		if err := DB.Model(&user).Update("active_vehicle_types", jsonTypes).Error; err != nil {
			return nil, err
		}
	}

	// Reload the user to return the updated record.
	DB.First(&user, userID)
	return &user, nil
}
