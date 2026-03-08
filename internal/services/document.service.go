package services

import (
	"errors"
	"fmt"
	"mime/multipart"
	"spark/internal/models"
)

func UploadDocument(userID uint, docType, fileName string, file multipart.File) (map[string]string, error) {
	folder := fmt.Sprintf("documents/%d/%s", userID, docType)
	url, err := UploadFile(file, fileName, folder)
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := DB.First(&user, userID).Error; err != nil {
		return nil, errors.New("user not found")
	}

	var driver models.Driver
	DB.Where("user_id = ?", userID).First(&driver)

	switch docType {
	case "driverPic":
		user.PFP = &url
		DB.Save(&user)
	case "license":
		driver.DriverLicensePhotoURL = url
		DB.Save(&driver)
	case "rc":
		driver.RCPhotoURL = url
		DB.Save(&driver)
	}
	return map[string]string{"url": url}, nil
}

func VerifyDocument(driverID uint, docType string) error {
	// Logic to set verified flags
	return nil
}
