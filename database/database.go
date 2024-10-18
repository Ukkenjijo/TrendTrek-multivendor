package database

import (
	"fmt"
	"log"

	"github.com/Ukkenjijo/trendtrek/config"
	"github.com/Ukkenjijo/trendtrek/models"
	"gorm.io/gorm"
)
var DB *gorm.DB
func ConnectToDB(){
	// Get the config object
	cfg := config.GetConfig()

	// Connect to the database
	var err error
	DB, err = config.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}

	// Run database migrations (example)
	err = DB.AutoMigrate(&models.User{},&models.Store{},&models.Category{},&models.Product{},&models.Image{},&models.Address{},&models.Cart{},&models.CartItem{},&models.Order{},&models.OrderItem{})
	if err != nil {
		fmt.Printf("Error during migration: %v\n", err)
	}

	fmt.Println("Database connection successful!")

}