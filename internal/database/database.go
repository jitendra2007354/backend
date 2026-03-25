package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() {
	var dsn string

	if os.Getenv("TIDB_HOST") != "" {
		// Production TiDB
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=true",
			os.Getenv("TIDB_USER"),
			os.Getenv("TIDB_PASSWORD"),
			os.Getenv("TIDB_HOST"),
			os.Getenv("TIDB_PORT"),
			os.Getenv("TIDB_DATABASE"),
		)
	} else {
		// Standard MySQL (Local or External)
		tls := "false"
		if mode := os.Getenv("DB_SSL_MODE"); mode != "" {
			tls = mode
		}

		port := os.Getenv("DB_PORT")
		if port == "" {
			port = "3306"
		}

		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=%s",
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_HOST"),
			port,
			os.Getenv("DB_NAME"),
			tls,
		)
	}

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	log.Println("✅ Database connection established")
}
