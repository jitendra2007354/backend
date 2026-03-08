package services

import "gorm.io/gorm"

// DB is the global database connection used by all services.
// It is initialized in main.go via dependency injection.
var DB *gorm.DB
