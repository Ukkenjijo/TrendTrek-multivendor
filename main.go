package main

import (
	"log"

	"github.com/Ukkenjijo/trendtrek/config"
	"github.com/Ukkenjijo/trendtrek/database"

	"github.com/Ukkenjijo/trendtrek/routes"
	"gorm.io/gorm"
	"github.com/gofiber/fiber/v2/middleware/logger"

	
	"github.com/gofiber/fiber/v2"
)

var DB *gorm.DB


func init(){
	config.LoadEnv()


}

func main() {
	app := fiber.New()
	app.Static("/uploads", "./uploads")
	
	app.Use(logger.New())

	database.ConnectToDB()

	// Setup routes
	routes.SetUpRoutes(app)

	log.Fatal(app.Listen(":3000"))
}
