package config

import (
	"fmt"
)

// AppConfig holds the application configuration
type AppConfig struct {
	Port        string
	DatabaseDSN string
	JWTSecret   string
	Env         string
}

// App is the global configuration instance
var App *AppConfig

// InitConfig initializes the application configuration
func InitConfig() {
	LoadEnv()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		GetEnv("DB_USER", "root"),
		GetEnv("DB_PASSWORD", ""),
		GetEnv("DB_HOST", "127.0.0.1"),
		GetEnv("DB_PORT", "3306"),
		GetEnv("DB_NAME", "spark_db"),
	)

	App = &AppConfig{
		Port:        GetEnv("PORT", "8000"),
		DatabaseDSN: dsn,
		JWTSecret:   GetEnv("JWT_SECRET", "default_secret"),
		Env:         GetEnv("NODE_ENV", "development"),
	}
}
