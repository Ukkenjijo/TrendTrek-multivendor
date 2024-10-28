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

		privateadmin.Post("/coupons/add",controllers.CreateCoupon)
		privateadmin.Get("/coupons",controllers.GetAllCoupons)
		

		
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
	user.Get("/stores",controllers.GetAllStores)
	user.Get("/products/category/:id",controllers.GetProductsByCategory)
	user.Get("/product/:id",controllers.GetProductbyId)
	user.Get("/search",controllers.SearchProducts)
	user.Get("google/login",controllers.GoogleLogin)
	user.Get("google/callback",controllers.GoogleCallback)
    
	privateuser:=app.Group("/api/v1/user")
	privateuser.Use(middleware.JWTMiddleware())
	{
		privateuser.Get("myaccount/profile",controllers.GetProfile)
		privateuser.Patch("myaccount/profile/update",controllers.UpdateProfile)
		privateuser.Get("myaccount/addresses",controllers.ListAddresses)
		privateuser.Post("myaccount/addresses",controllers.AddAddress)
		privateuser.Patch("myaccount/addresses/:id",controllers.EditAddress)
		privateuser.Delete("myaccount/addresses/:id",controllers.DeleteAddress)
		privateuser.Get("myaccount/getrefferallink",controllers.GenerateReferralLink)
		privateuser.Get("myaccount/wallet",controllers.GetWalletBallance)
		privateuser.Get("myaccount/wallet/history",controllers.GetWalletHistory)
		privateuser.Post("cart/add",controllers.AddToCart)
		privateuser.Get("cart",controllers.ListCartItems)
		privateuser.Put("cart/update/:id",controllers.UpdateCartQuantity)
		privateuser.Delete("cart/remove/:id",controllers.RemoveFromCart)
		privateuser.Post("/wishlist/add/:product_id",controllers.AddToWishlist)
		privateuser.Delete("wishlist/remove/:product_id",controllers.RemoveFromWishlist)
		privateuser.Get("wishlist",controllers.GetWishlist)
		privateuser.Post("checkout/orders",controllers.PlaceOrder)
		privateuser.Get("orders",controllers.ListOrders)
		privateuser.Get("orders/:id",controllers.GetOrderDetails)
		privateuser.Put("orders/cancel/:id",controllers.CancelOrder)
		privateuser.Patch("orders/return/:id", controllers.ReturnOrderItem)
		privateuser.Post("coupons/apply",controllers.ApplyCoupon)
		privateuser.Put("coupons/remove",controllers.RemoveCoupon)
		privateuser.Put("orders/cancel/:order_id/:item_id",controllers.CancelOrderItem)
		privateuser.Post("payments/razorpay/verify", controllers.VerfyRazorpayPayment)
		


		

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
		privatestore.Put("/products/updatestock/:id",controllers.UpdateProductStock)
		privatestore.Get("/products",controllers.GetProducts)
		privatestore.Post("/products/:product_id/offer",controllers.CreateOrUpdateOffer)	
		privatestore.Delete("/products/:product_id/offer",controllers.DeleteOffer)
		privatestore.Get("/products/offers",controllers.ListOffers)
		privatestore.Get("myaccount/seller/profile",controllers.GetProfile)
		privatestore.Patch("myaccount/seller/profile/update",controllers.UpdateProfile)
		privatestore.Get("myaccount/store/profile",controllers.GetStoreProfile)
		privatestore.Patch("myaccount/store/profile/update",controllers.UpdateStoreProfile)
		privatestore.Get("orders",controllers.ListSellerOrders)
		privatestore.Put("orders/:order_id/:item_id/status",controllers.UpdateOrderItemStatus)
		

	}
	
}
