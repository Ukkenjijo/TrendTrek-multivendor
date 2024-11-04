package controllers

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/Ukkenjijo/trendtrek/database"
	"github.com/Ukkenjijo/trendtrek/models"
	"github.com/gofiber/fiber/v2"
	"github.com/jung-kurt/gofpdf"
	"gorm.io/gorm"
)

func GetSalesReport(c *fiber.Ctx) error {
	userId := c.Locals("user_id")
	sellerId, _ := GetStoreIDByUserID(uint(userId.(float64)))

	// Parse the date range filter
	dateRange := c.Query("range")           // Options: "daily", "weekly", "monthly", "custom"
	startDateParam := c.Query("start_date") // For custom date range
	endDateParam := c.Query("end_date")

	// Define start and end dates
	var startDate, endDate time.Time
	now := time.Now()

	var pendingCount int64
	var returnedCount int64
	var completedCount int64

	// Set start and end dates based on date range
	switch dateRange {
	case "daily":
		startDate = now.Truncate(24 * time.Hour)
		endDate = startDate.Add(24 * time.Hour)
	case "weekly":
		startDate = now.AddDate(0, 0, -int(now.Weekday())) // Start of the week
		endDate = startDate.AddDate(0, 0, 7)               // End of the week
	case "monthly":
		startDate = now.AddDate(0, 0, -now.Day()+1) // Start of the month
		endDate = startDate.AddDate(0, 1, 0)        // End of the month
	case "custom":
		var err error
		startDate, err = time.Parse("2006-01-02", startDateParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid date format"})
		}
		endDate, err = time.Parse("2006-01-02", endDateParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid date format"})
		}
		endDate = endDate.Add(24 * time.Hour) // Include the full end date
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid range option"})
	}

	//Initalize report metrics
	var totalSalesCount int64
	var totalOrderAmount, totalDiscounts float64

	//Query all the orders from order items table of the particular seller
	var orders []models.OrderItem
	if err := database.DB.Preload("Product", func(db *gorm.DB) *gorm.DB { return db.Preload("Offer") }).Where("product_id in (SELECT id FROM products WHERE store_id = ?)", sellerId).Where("created_at BETWEEN ? AND ?", startDate, endDate).Find(&orders).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve orders"})
	}
	//calculate total sales count
	for _, order := range orders {
		totalSalesCount++
		totalOrderAmount += math.Round(order.TotalPrice)
		log.Println(order.Product.OfferID)
		if order.Product.Offer != nil {
			discount := math.Round(order.TotalPrice * order.Product.Offer.DiscountPercentage / 100)
			log.Println(discount, " ", order.Product.Offer.DiscountPercentage)
			totalDiscounts += discount
		}
		if order.Status == "pending" {
			pendingCount++
		} else if order.Status == "returned" {
			returnedCount++
		} else if order.Status == "completed" {
			completedCount++
		}

	}

	salesreport := models.SalesReportSeller{
		TotalSalesCount:  totalSalesCount,
		TotalOrderAmount: totalOrderAmount,
		TotalDiscounts:   totalDiscounts,
		PendingCount:     pendingCount,
		ReturnedCount:    returnedCount,
		CompletedCount:   completedCount,
	}
	//return the sales report
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": salesreport})

}

func GetSalesReportAdmin(c *fiber.Ctx) error {
	// Parse the date range filter
	dateRange := c.Query("range")           // Options: "daily", "weekly", "monthly", "custom"
	startDateParam := c.Query("start_date") // For custom date range
	endDateParam := c.Query("end_date")

	// Define start and end dates
	var startDate, endDate time.Time
	now := time.Now()

	// Set start and end dates based on date range
	switch dateRange {
	case "daily":
		startDate = now.Truncate(24 * time.Hour)
		endDate = startDate.Add(24 * time.Hour)
	case "weekly":
		startDate = now.AddDate(0, 0, -int(now.Weekday())) // Start of the week
		endDate = startDate.AddDate(0, 0, 7)               // End of the week
	case "monthly":
		startDate = now.AddDate(0, 0, -now.Day()+1) // Start of the month
		endDate = startDate.AddDate(0, 1, 0)        // End of the month
	case "custom":
		var err error
		startDate, err = time.Parse("2006-01-02", startDateParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid date format"})
		}
		endDate, err = time.Parse("2006-01-02", endDateParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid date format"})
		}
		endDate = endDate.Add(24 * time.Hour) // Include the full end date
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid range option"})
	}
	// Initialize report metrics
	var totalSalesCount, codCount, razorpayCount, walletCount, pendingCount, returnedCount, completedCount int64
	var totalOrderAmount, FinalOrderAmount, totalDiscounts, couponDeductions float64

	//fetch all the orders
	var orders []models.Order
	if err := database.DB.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Preload("Product", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Images")
		})
	}).Where("created_at BETWEEN ? AND ?", startDate, endDate).Find(&orders).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve orders"})
	}
	//calculate metrics from the orders
	for _, order := range orders {
		var orderpaymentdetails models.OrderPaymentDetail
		if err := database.DB.Where("order_id = ?", order.ID).First(&orderpaymentdetails).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve order payment details"})
		}

		totalSalesCount++
		FinalOrderAmount += math.Round(orderpaymentdetails.FinalOrderAmount)
		totalOrderAmount += math.Round(orderpaymentdetails.OrderAmount)
		totalDiscounts += math.Round(orderpaymentdetails.OrderDiscount)
		couponDeductions += math.Round(orderpaymentdetails.CouponSavings)
		if orderpaymentdetails.PaymentType == "COD" {
			codCount++
		}
		if orderpaymentdetails.PaymentType == "Razorpay" {
			razorpayCount++
		}
		if orderpaymentdetails.PaymentType == "Wallet" {
			walletCount++
		}
		if order.Status == "pending" {
			pendingCount++
		} else if order.Status == "returned" {
			returnedCount++
		} else if order.Status == "completed" {
			completedCount++
		}
	}
	//create a custom response for the sales report
	salesreport := fiber.Map{
		"totalSalesCount":                  totalSalesCount,
		"totalOrderAmount":                 totalOrderAmount,
		"totalDiscounts":                   totalDiscounts,
		"FinalOrderAmountbefore descounts": FinalOrderAmount,
		"No. of cod transactions":          codCount,
		"No. of razorpay transactions":     razorpayCount,
		"No. of wallet transactions":       walletCount,
		"Coupon discounts":                 math.Abs(couponDeductions),
		"No. of pending orders":            pendingCount,
		"No. of returned orders":           returnedCount,
		"No. of completed orders":          completedCount,
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": salesreport})

}

// GenerateSalesReportPDF generates a PDF for the sales report
func GenerateSalesReportPDF(c *fiber.Ctx) error {
	// Retrieve the sales report data based on your GetSalesReport function
	// The code below is based on `GetSalesReportAdmin` for example

	// Example of setting the date range (daily, weekly, monthly, custom)
	// Parse date range as in `GetSalesReportAdmin`
	// This is just a sample; you may replace it with actual data

	dateRange := c.Query("range")           // Options: "daily", "weekly", "monthly", "custom"
	startDateParam := c.Query("start_date") // For custom date range
	endDateParam := c.Query("end_date")     // For custom date range

	var startDate, endDate time.Time
	now := time.Now()

	// Set start and end dates based on range
	switch dateRange {
	case "daily":
		startDate = now.Truncate(24 * time.Hour)
		endDate = startDate.Add(24 * time.Hour)
	case "weekly":
		startDate = now.AddDate(0, 0, -int(now.Weekday()))
		endDate = startDate.AddDate(0, 0, 7)
	case "monthly":
		startDate = now.AddDate(0, 0, -now.Day()+1)
		endDate = startDate.AddDate(0, 1, 0)
	case "custom":
		var err error
		startDate, err = time.Parse("2006-01-02", startDateParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid date format"})
		}
		endDate, err = time.Parse("2006-01-02", endDateParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid date format"})
		}
		endDate = endDate.Add(24 * time.Hour)
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid range option"})
	}

	// Initialize the metrics as in GetSalesReportAdmin
	var totalSalesCount, codCount, razorpayCount, walletCount, pendingCount, returnedCount, completedCount int64
	var totalOrderAmount, finalOrderAmount, totalDiscounts, couponDeductions float64

	// Fetch and calculate report metrics
	var orders []models.Order
	if err := database.DB.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Preload("Product", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Images")
		})
	}).Where("created_at BETWEEN ? AND ?", startDate, endDate).Find(&orders).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve orders"})
	}

	// Calculate totals
	for _, order := range orders {
		var orderPaymentDetails models.OrderPaymentDetail
		if err := database.DB.Where("order_id = ?", order.ID).First(&orderPaymentDetails).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve order payment details"})
		}
		totalSalesCount++
		finalOrderAmount += math.Round(orderPaymentDetails.FinalOrderAmount)
		totalOrderAmount += math.Round(orderPaymentDetails.OrderAmount)
		totalDiscounts += math.Round(orderPaymentDetails.OrderDiscount)
		couponDeductions += math.Round(orderPaymentDetails.CouponSavings)
		switch orderPaymentDetails.PaymentType {
		case "COD":
			codCount++
		case "Razorpay":
			razorpayCount++
		case "Wallet":
			walletCount++
		}
		if order.Status == "pending" {
			pendingCount++
		} else if order.Status == "returned" {
			returnedCount++
		} else if order.Status == "completed" {
			completedCount++
		}
	}

	// Create a new PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle("Sales Report", false)
	pdf.AddPage()

	// Set header font and title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "Sales Report")
	pdf.Ln(20) // Line break

	// Set regular font for the rest of the report
	pdf.SetFont("Arial", "", 12)

	// Add report details
	pdf.Cell(190, 10, fmt.Sprintf("Date Range: %s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")))
	pdf.Ln(10)

	reportData := []string{
		fmt.Sprintf("Total Sales Count: %d", totalSalesCount),
		fmt.Sprintf("Total Order Amount: $%.2f", totalOrderAmount),
		fmt.Sprintf("Total Discounts: $%.2f", totalDiscounts),
		fmt.Sprintf("Final Order Amount before Discounts: $%.2f", finalOrderAmount),
		fmt.Sprintf("No. of COD Transactions: %d", codCount),
		fmt.Sprintf("No. of Razorpay Transactions: %d", razorpayCount),
		fmt.Sprintf("No. of Wallet Transactions: %d", walletCount),
		fmt.Sprintf("Coupon Deductions: $%.2f", math.Abs(couponDeductions)),
		fmt.Sprintf("No. of Pending Orders: %d", pendingCount),
		fmt.Sprintf("No. of Returned Orders: %d", returnedCount),
		fmt.Sprintf("No. of Completed Orders: %d", completedCount),
	}

	// Add report data to PDF
	for _, line := range reportData {
		pdf.Cell(190, 10, line)
		pdf.Ln(10) // Line break
	}

	// Set a filename for the PDF
	filename := fmt.Sprintf("sales_report_%s.pdf", time.Now().Format("20060102150405"))

	// Output PDF to file
	if err := pdf.OutputFileAndClose(filename); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate PDF"})
	}

	// Serve the PDF file as a downloadable response
	return c.SendFile(filename)
}
