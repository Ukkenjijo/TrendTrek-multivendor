package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/Ukkenjijo/trendtrek/database"
	"github.com/Ukkenjijo/trendtrek/models"
	"github.com/gofiber/fiber/v2"
	"github.com/jung-kurt/gofpdf"
	"gorm.io/gorm"
)

func VerfyRazorpayPayment(c *fiber.Ctx) error {
	userID := c.Locals("user_id")

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
	if err := ReducestockandDeleteCart(database.DB, &cart); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Cart not found"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Order paid successfully"})

}

func RetryPayment(c *fiber.Ctx) error {
	InitRazorpay()
	orderID:= c.Params("order_id")
	var payment models.Payment
	if err:=database.DB.Where("order_id = ?",orderID).First(&payment).Error;err!=nil{
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Payment not found"})
	}
	if payment.PaymentStatus=="paid"{
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Payment already completed"})
	}

	//Initalize the payment for the order retry
	amount:=float64(payment.Amount)
	amount=amount*100
	options:=map[string]interface{}{
		"amount":amount,
		"currency":"INR",
		"receipt":fmt.Sprintf("order_%d",payment.OrderID),
		"payment_capture":1,
	}
	razorpayOrder, err := razorpayClient.Order.Create(options,nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create Razorpay order"})
	}
	payment.RazorpayPaymentID=razorpayOrder["id"].(string)
	if err:=database.DB.Save(&payment).Error;err!=nil{
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update payment"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":"Order placed successfully",
		"order_id":payment.OrderID,
		"amount":payment.Amount,
		"razorpay_order_id": razorpayOrder["id"],
		"currency":"INR",
	})

	
}






func GenerateInvoicePdf(c *fiber.Ctx) error {
	orderID := c.Params("order_id")

	//Retrieve the order infromation from the database
	var order models.Order
	if err := database.DB.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Preload("Product", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Images")
		})
	}).First(&order, orderID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order not found"})
	}
	if order.Status != "completed" {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Invoice cannot be generated for incomplited status"})
	}
	var orderpaymentdetails models.OrderPaymentDetail
	if err := database.DB.Where("order_id = ?", orderID).First(&orderpaymentdetails).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order payment details not found"})
	}
	var name string
	if err := database.DB.Model(&models.User{}).Select("name").Where("id = ?", order.UserID).First(&name).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}
	ShippingAddress := fmt.Sprintf("%s, %s, %s, %s, %s", order.ShippingStreet, order.ShippingCity, order.ShippingState, order.ShippingCountry, order.ShippingZipCode)
	//Initalize the pdf generator
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Company Header Section
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(95, 5, "TRENDTREK RETAIL LIMITED")
	pdf.Cell(95, 5, "Customer Support: 1800-889-9991")
	pdf.Ln(6)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(95, 5, "Gr. Floor, Reliance Corporate, IT Park Ltd")
	pdf.Cell(95, 5, "Email: customercare@trendtrek.com")
	pdf.Ln(6)
	pdf.Cell(95, 5, "Navi Mumbai, MAH 400601")
	pdf.Cell(95, 5, "GSTIN: 27AABCR1718E1ZP")
	pdf.Ln(10)

	// Invoice Title
	pdf.SetFont("Arial", "B", 20)
	pdf.Cell(0, 10, "TAX INVOICE")
	pdf.Ln(12)

	// Order and Customer Details Section
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(100, 10, fmt.Sprintf("Invoice Number: INV-%s", orderID))
	pdf.Cell(90, 10, fmt.Sprintf("Date: %s", time.Now().Format("2006-01-02")))
	pdf.Ln(8)

	// Customer Details
	pdf.Cell(100, 10, fmt.Sprintf("Customer: %s", name))
	pdf.Ln(6)
	pdf.Cell(100, 10, fmt.Sprintf("Shipping Address: %s", ShippingAddress))
	pdf.Ln(6)

	pdf.Cell(100, 10, "GSTIN: UNREGISTERED")

	pdf.Ln(6)
	pdf.Cell(100, 10, fmt.Sprintf("State Code: %s", "KL"))
	pdf.Ln(10)

	// Shipping Details
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(100, 6, fmt.Sprintf("Order ID: %s", orderID))
	pdf.Cell(90, 6, fmt.Sprintf("Payment Mode: %s", order.PaymentMode))
	pdf.Ln(6)
	pdf.Cell(100, 6, fmt.Sprintf("Carrier: %s", "DELHIVERY"))
	pdf.Cell(90, 6, fmt.Sprintf("AWB Number: %s", "123456789"))
	pdf.Ln(10)

	// Table Headers for Items
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(240, 240, 240)
	// Adjusted column widths to accommodate new fields
	pdf.CellFormat(60, 10, "Product", "1", 0, "C", true, 0, "")
	pdf.CellFormat(20, 10, "HSN", "1", 0, "C", true, 0, "")
	pdf.CellFormat(20, 10, "Qty", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 10, "MRP", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 10, "Discount", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 10, "Unit Price", "1", 0, "C", true, 0, "")
	pdf.CellFormat(20, 10, "Total Price", "1", 1, "C", true, 0, "")

	// Populate Table with Item Data
	pdf.SetFont("Arial", "", 10)
	for _, item := range order.Items {
		// First line: Product name
		currentY := pdf.GetY()
		pdf.MultiCell(60, 10, item.Product.Name, "1", "L", false)
		newY := pdf.GetY()
		pdf.SetY(currentY)
		pdf.SetX(70) // Start X position for remaining cells

		// Remaining cells
		pdf.CellFormat(20, newY-currentY, "12345", "1", 0, "C", false, 0, "")
		pdf.CellFormat(20, newY-currentY, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, newY-currentY, fmt.Sprintf("$%.2f", item.Product.Price), "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, newY-currentY, fmt.Sprintf("$%.2f", (item.Product.Price-item.Price)), "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, newY-currentY, fmt.Sprintf("$%.2f", item.Price), "1", 0, "C", false, 0, "")
		pdf.CellFormat(20, newY-currentY, fmt.Sprintf("$%.2f",item.TotalPrice ), "1", 1, "C", false, 0, "")
	}

	// Add totals section with more details
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(150, 10, "Subtotal:", "", 0, "R", false, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("$%.2f", orderpaymentdetails.OrderAmount), "", 1, "C", false, 0, "")

	pdf.CellFormat(150, 10, "Discount:", "", 0, "R", false, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("$%.2f", orderpaymentdetails.OrderDiscount), "", 1, "C", false, 0, "")

	pdf.CellFormat(150, 10, "Coupon Savings:", "", 0, "R", false, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("$%.2f", orderpaymentdetails.CouponSavings), "", 1, "C", false, 0, "")

	pdf.CellFormat(150, 10, "Shipping Charge:", "", 0, "R", false, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("$%.2f", orderpaymentdetails.ShippingCost), "", 1, "C", false, 0, "")

	pdf.CellFormat(150, 10, "GST:", "", 0, "R", false, 0, "")
	pdf.CellFormat(40, 10, "NIL", "", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(150, 10, "Final Total:", "", 0, "R", false, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("$%.2f", orderpaymentdetails.FinalOrderAmount), "", 1, "C", false, 0, "")

	// Footer notes
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 8)
	pdf.MultiCell(0, 5, "1. Products being sent under this invoice are for personal consumption of the customer and not for re-sale or commercial purposes.", "", "", false)
	pdf.MultiCell(0, 5, "E. & O.E.", "", "", false)
	pdf.MultiCell(0, 5, "An Electronic document issued in accordance with the provisions of the Information Technology Act, 2000", "", "", false)

	// Save the PDF to a file
	filename := fmt.Sprintf("invoice_%s.pdf", orderID)
	if err := pdf.OutputFileAndClose(filename); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate invoice PDF"})
	}
	// Serve the PDF file as a downloadable response
	return c.SendFile(filename)

}
