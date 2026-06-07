package database

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB initializes database connection with MySQL
func InitDB(dsn string) error {
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connected successfully")
	return nil
}

// InitDBWithDSN initializes database connection with specified driver
func InitDBWithDSN(dsn string, dbType string) error {
	var err error

	switch dbType {
	case "mysql":
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Printf("Database connected successfully (type: %s)", dbType)
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
