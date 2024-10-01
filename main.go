package main

import (
	
	"log"
	

	"github.com/Ukkenjijo/trendtrek/config"
	"github.com/Ukkenjijo/trendtrek/database"

	"github.com/Ukkenjijo/trendtrek/routes"
	"gorm.io/gorm"

	"github.com/gofiber/fiber/v2"
)

var DB *gorm.DB

func main() {
	app := fiber.New()
	config.LoadEnv()

	database.ConnectToDB()

	// Setup routes
	routes.SetUpRoutes(app)
	
	

	log.Fatal(app.Listen(":3000"))
}
