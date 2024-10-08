// /config/config.go
package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Config struct holds the database configuration
type Config struct {
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	DBPort     string
}

// LoadEnv loads environment variables from the .env file
func LoadEnv() {
	err := godotenv.Load()
	fmt.Println(err)
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

// GetConfig reads the environment variables and returns a Config object
func GetConfig() *Config {
	return &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBPort:     os.Getenv("DB_PORT"),
	}
	
}



// ConnectDB establishes a connection to the database and returns a GORM DB instance
func ConnectDB(config *Config) (*gorm.DB, error) {
	fmt.Println(os.Getenv("GOOGLE_CLIENT_ID"))
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		config.DBHost, config.DBUser, config.DBPassword, config.DBName, config.DBPort)
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

