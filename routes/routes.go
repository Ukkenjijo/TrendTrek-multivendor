package routes

import (
	"github.com/Ukkenjijo/trendtrek/controllers"
	"github.com/gofiber/fiber/v2"
)

func SetUpRoutes(app *fiber.App) {
	user := app.Group("/api/v1/user")
	user.Post("/signup", controllers.Signup)        // Signup route
	user.Post("/verify-otp", controllers.VerifyOTP) // OTP verification route
	user.Post("/resend-otp",controllers.ResendOTP)
}
