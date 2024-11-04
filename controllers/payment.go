package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"
	

	"github.com/Ukkenjijo/trendtrek/database"
	"github.com/Ukkenjijo/trendtrek/models"
	"github.com/gofiber/fiber/v2"
)

func VerfyRazorpayPayment(c *fiber.Ctx) error {
	userID:=c.Locals("user_id")

	var payload models.RAZORPAY_Payment
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Review your payload", "data": err})
	}
	//Verify Razorypay signature
	secret := os.Getenv("RAZORPAY_SECRET")
	body := payload.RazorpayOrderID + "|" + payload.RazorpayPaymentID
	computedSignature := hmac.New(sha256.New, []byte(secret))
	computedSignature.Write([]byte(body))
	expectedSignature := hex.EncodeToString(computedSignature.Sum(nil))

	if payload.RazorpaySignature != expectedSignature {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Signature mismatch"})
	}
	//set the payment status to success
	var payment models.Payment
	if err := database.DB.Where("razorpay_payment_id = ?", payload.RazorpayOrderID).First(&payment).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Payment not found"})
	}
	payment.PaymentStatus = "paid"
	database.DB.Save(&payment)

	//get the users cart and clear it after payment
	var cart models.Cart
	if err := database.DB.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Cart not found"})
	}
	if err:=ReducestockandDeleteCart(database.DB,&cart);err!=nil{
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Cart not found"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Order paid successfully"})

}


