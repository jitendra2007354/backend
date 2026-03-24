package services

import (
	"errors"

	"spark/internal/models"

	"gorm.io/gorm"
)

func GetConfig(key string) (*models.Config, error) {
	var config models.Config
	if err := DB.Where("`key` = ?", key).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func GetApplicableConfig() (*models.Config, error) {
	config, err := GetConfig("global")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			defaultConfig := models.Config{
				Key:                     "global",
				BaseFare:                50,
				BaseFareBike:            25,
				BaseFareAuto:            35,
				BaseFareCar:             50,
				BaseFareSUV:             65,
				BaseFareLuxury:          100,
				PerKmRate:               12,
				PerMinuteRate:           1,
				CommissionRate:          12,
				CancellationFee:         50,
				DriverSearchRadius:      5,
				RideAcceptTime:          30,
				WalletMinBalance:        0,
				MaxBidPerKm:             30,
				MinBidPerKm:             8,
				MinBidPerKmBike:         4.8,
				MinBidPerKmAuto:         6.4,
				MinBidPerKmCar:          8.0,
				MinBidPerKmSUV:          9.6,
				MinBidPerKmLuxury:       14.4,
				AutoBlockHours:          24,
				SurgeMultiplier:         1,
				TaxRate:                 5,
				CancellationGracePeriod: 60,
			}
			if err := DB.Create(&defaultConfig).Error; err != nil {
				return nil, err
			}
			return &defaultConfig, nil
		}
		return nil, err
	}
	return config, nil
}

func SetConfig(data map[string]interface{}) error {
	key, ok := data["key"].(string)
	if !ok || key == "" {
		return errors.New("key required")
	}

	config := models.Config{Key: key}

	if v, ok := data["baseFare"].(float64); ok {
		config.BaseFare = v
	}
	if v, ok := data["baseFareBike"].(float64); ok {
		config.BaseFareBike = v
	}
	if v, ok := data["baseFareAuto"].(float64); ok {
		config.BaseFareAuto = v
	}
	if v, ok := data["baseFareCar"].(float64); ok {
		config.BaseFareCar = v
	}
	if v, ok := data["baseFareSuv"].(float64); ok {
		config.BaseFareSUV = v
	}
	if v, ok := data["baseFareLuxury"].(float64); ok {
		config.BaseFareLuxury = v
	}
	if v, ok := data["perKmRate"].(float64); ok {
		config.PerKmRate = v
	}
	if v, ok := data["perMinuteRate"].(float64); ok {
		config.PerMinuteRate = v
	}
	if v, ok := data["commissionRate"].(float64); ok {
		config.CommissionRate = v
	}
	if v, ok := data["cancellationFee"].(float64); ok {
		config.CancellationFee = v
	}
	if v, ok := data["driverSearchRadius"].(float64); ok {
		config.DriverSearchRadius = v
	}
	if v, ok := data["rideAcceptTime"].(float64); ok {
		config.RideAcceptTime = int(v)
	}
	if v, ok := data["walletMinBalance"].(float64); ok {
		config.WalletMinBalance = v
	}
	if v, ok := data["maxBidPerKm"].(float64); ok {
		config.MaxBidPerKm = v
	}
	if v, ok := data["minBidPerKm"].(float64); ok {
		config.MinBidPerKm = v
	}
	if v, ok := data["minBidPerKmBike"].(float64); ok {
		config.MinBidPerKmBike = v
	}
	if v, ok := data["minBidPerKmAuto"].(float64); ok {
		config.MinBidPerKmAuto = v
	}
	if v, ok := data["minBidPerKmCar"].(float64); ok {
		config.MinBidPerKmCar = v
	}
	if v, ok := data["minBidPerKmSuv"].(float64); ok {
		config.MinBidPerKmSUV = v
	}
	if v, ok := data["minBidPerKmLuxury"].(float64); ok {
		config.MinBidPerKmLuxury = v
	}
	if v, ok := data["autoBlockHours"].(float64); ok {
		config.AutoBlockHours = int(v)
	}
	if v, ok := data["surgeMultiplier"].(float64); ok {
		config.SurgeMultiplier = v
	}
	if v, ok := data["taxRate"].(float64); ok {
		config.TaxRate = v
	}
	if v, ok := data["cancellationGracePeriod"].(float64); ok {
		config.CancellationGracePeriod = int(v)
	}

	return DB.Where("`key` = ?", key).Assign(config).FirstOrCreate(&config).Error
}
