package routes

import (
	"github.com/Ukkenjijo/trendtrek/controllers"
	"github.com/Ukkenjijo/trendtrek/middleware"
	"github.com/gofiber/fiber/v2"
)

func SetUpRoutes(app *fiber.App) {

	admin:=app.Group("/api/v1/admin")
	admin.Post("/login",controllers.AdminLogin)
	privateadmin:=app.Group("/api/v1/admin")
	privateadmin.Use(middleware.JWTMiddleware(),middleware.AdminRoleMiddleware())
	{
		privateadmin.Get("/users",controllers.GetAllUsers)
		privateadmin.Patch("/users/block",controllers.BlockUser)
		privateadmin.Patch("/users/unblock",controllers.UnblockUser)
		privateadmin.Get("/users/blocked",controllers.ListBlockedUsers)
		privateadmin.Patch("/vendor/unblock",controllers.UnblockUser)
		privateadmin.Get("/vendor/blocked",controllers.ListBlockedUsers)

		privateadmin.Post("/categories/add",controllers.AddCategory)
		privateadmin.Patch("/categories/edit/:id",controllers.EditCategory)
		privateadmin.Delete("/categories/delete/:id",controllers.DeleteCategory)
		
	}
	




	user := app.Group("/api/v1/user")
	user.Post("/signup", controllers.Signup)        // Signup route
	user.Post("/forget-password",controllers.ForgetPassword)
	user.Patch("/forget-password/reset",controllers.ResetPassword)
	user.Post("/verify-otp", controllers.VerifyOTP) // OTP verification route
	user.Post("/resend-otp", controllers.ResendOTP)
	user.Post("/login",controllers.Login)
	user.Get("/categories",controllers.GetAllCategories)
	user.Get("/category/:id",controllers.GetCategoryByID)
	user.Get("/products",controllers.GetAllProducts)
	user.Get("/products/category/:id",controllers.GetProductsByCategory)
	user.Get("/product/:id",controllers.GetProductbyId)
	user.Get("google/login",controllers.GoogleLogin)
	user.Get("google/callback",controllers.GoogleCallback)
    
	privateuser:=app.Group("/api/v1/user")
	privateuser.Use(middleware.JWTMiddleware())
	{
		privateuser.Get("myaccount/profile",controllers.GetProfile)
		privateuser.Patch("myaccount/profile/update",controllers.UpdateProfile)
		

	}


	store := app.Group("/api/v1/vendor")
	store.Post("/signup", controllers.StoreSignup)
	store.Post("/verify-otp", controllers.VerifyOTP)
	store.Post("/resend-otp", controllers.ResendOTP)
	store.Post("/login",controllers.Login)

	privatestore := app.Group("/api/v1/vendor")
	privatestore.Use(middleware.JWTMiddleware(),middleware.SellerRoleMiddleware())
	{
		privatestore.Post("/products/add",controllers.AddProduct)
		privatestore.Post("/products/edit/:id",controllers.EditProduct)
		privatestore.Delete("/products/delete/:id",controllers.DeleteProduct)
		privatestore.Get("/products",controllers.GetProducts)	

	}
	
}
