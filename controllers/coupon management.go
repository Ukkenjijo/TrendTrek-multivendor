package controllers

import (
	"time"

	"github.com/Ukkenjijo/trendtrek/database"
	"github.com/Ukkenjijo/trendtrek/models"
	"github.com/Ukkenjijo/trendtrek/utils"
	"github.com/gofiber/fiber/v2"
)

func CreateCoupon(c *fiber.Ctx) error {
	coupon := new(models.Coupon)
	couponreq := new(models.CouponRequest)
	if err := c.BodyParser(couponreq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Review your input", "data": err})
	}
	if err := utils.ValidateStruct(couponreq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Review your input", "data": err})
	}
	//parse the expiration date
	expirationDate, err := time.Parse(time.RFC3339, couponreq.ExpiresAt)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Review your input", "data": err})
	}
	coupon.Name = couponreq.Name
	coupon.Code = couponreq.Code
	coupon.ExpiresAt = expirationDate
	coupon.Discount = couponreq.Discount
	coupon.MaxUsage = couponreq.MaxUsage
	coupon.MinPurchaseAmount = couponreq.MinPurchaseAmount
	coupon.MaxDiscountAmount = couponreq.MaxDiscountAmount
	coupon.UsageCount = 0

	if err := database.DB.Create(&coupon).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Couldn't create coupon", "data": err})
	}
	return c.JSON(fiber.Map{"status": "success", "message": "Created coupon", "data": coupon})
}

func GetAllCoupons(c *fiber.Ctx) error {
	var coupons []models.Coupon
	if err := database.DB.Find(&coupons).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Couldn't get coupons", "data": err})
	}
	return c.JSON(fiber.Map{"status": "success", "message": "Fetched coupons", "data": coupons})
}

func ApplyCoupon(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	var req struct {
		CouponCode string `json:"coupon_code" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Review your input", "data": err})
	}
	//Find the coupon by code
	var coupon models.Coupon
	if err := database.DB.Where("code = ?", req.CouponCode).First(&coupon).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "error", "message": "Coupon not found", "data": err})
	}
	//Check if the coupon is expired or has exceeded its usage limit
	if coupon.ExpiresAt.Before(time.Now()) || coupon.MaxUsage <= coupon.UsageCount {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "Coupon expired or has exceeded its usage limit",})
	}
	//Get the users cart
	var cart models.Cart
	if err := database.DB.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "error", "message": "Cart not found", "data": err})
	}
	cart.CouponID=&coupon.ID
	if err := database.DB.Save(&cart).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Couldn't apply coupon", "data": err})
	}
	return c.JSON(fiber.Map{"status": "success", "message": "Coupon applied successfully", "data": cart})
}

func RemoveCoupon(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	//Get the users cart
	var cart models.Cart
	if err := database.DB.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "error", "message": "Cart not found", "data": err})
	}
	//Reset the coupon discount
	cart.CouponID = nil
	cart.CouponDiscount = 0
	//Update the cart
	if err := database.DB.Save(&cart).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Couldn't remove coupon", "data": err})
	}
	return c.JSON(fiber.Map{"status": "success", "message": "Coupon removed successfully", "data": nil})
}


