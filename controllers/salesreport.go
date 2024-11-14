package controllers

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/Ukkenjijo/trendtrek/database"
	"github.com/Ukkenjijo/trendtrek/models"
	"github.com/gofiber/fiber/v2"
	"github.com/jung-kurt/gofpdf"
	chart "github.com/wcharczuk/go-chart"
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

type PaymentData struct {
	COD      int64
	Razorpay int64
	Wallet   int64
}

type FinancialData struct {
	TotalOrderAmount float64
	FinalOrderAmount float64
	TotalDiscounts   float64
	CouponDeductions float64
}

type OrderStatusData struct {
	Pending   int64
	Completed int64
	Returned  int64
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
	// Create PDF with custom styling
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)

	// Add first page
	pdf.AddPage()
	// Add company logo (assuming you have a logo.png in your assets)

	pdf.ImageOptions("uploads/logo.jpg", 10, 10, 30, 0, false, gofpdf.ImageOptions{ImageType: "jpg"}, 0, "")

	// Header section
	pdf.SetFont("Arial", "B", 20)
	pdf.SetTextColor(44, 62, 80) // Dark blue color
	pdf.CellFormat(190, 10, "Sales Report", "", 1, "C", false, 0, "")
	pdf.Ln(8)

	// Subheader with date range
	pdf.SetFont("Arial", "I", 12)
	pdf.SetTextColor(127, 140, 141) // Gray color
	pdf.CellFormat(190, 10, fmt.Sprintf("Period: %s to %s",
		startDate.Format("January 2, 2006"),
		endDate.Format("January 2, 2006")), "", 1, "C", false, 0, "")
	pdf.Ln(20)

	// Summary section
	addSummarySection(pdf, SummaryData{
		TotalSales:      totalSalesCount,
		Revenue:         finalOrderAmount,
		TotalDiscounts:  totalDiscounts,
		CompletedOrders: completedCount,
	})
	pdf.Ln(15)

	addPaymentMethodsSection(pdf, PaymentData{
		COD:      codCount,
		Razorpay: razorpayCount,
		Wallet:   walletCount,
	})
	pdf.Ln(15)

	// Order Status Breakdown
	addOrderStatusSection(pdf, OrderStatusData{
		Pending:   pendingCount,
		Completed: completedCount,
		Returned:  returnedCount,
	})
	pdf.Ln(15)

	// Financial Details
	addFinancialDetailsSection(pdf, FinancialData{
		TotalOrderAmount: totalOrderAmount,
		FinalOrderAmount: finalOrderAmount,
		TotalDiscounts:   totalDiscounts,
		CouponDeductions: couponDeductions,
	})
	pdf.Ln(15)

	// Generate and embed charts in the PDF
	err := generateAndEmbedCharts(pdf, codCount, razorpayCount, walletCount, pendingCount, completedCount, returnedCount)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate charts"})
	}

	// Footer
	addFooter(pdf)

	// Set a filename for the PDF
	filename := fmt.Sprintf("sales_report_%s.pdf", time.Now().Format("20060102150405"))

	// Output PDF to file
	if err := pdf.OutputFileAndClose(filename); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate PDF"})
	}

	// Serve the PDF file as a downloadable response
	return c.SendFile(filename)
}

type SummaryData struct {
	TotalSales      int64
	Revenue         float64
	TotalDiscounts  float64
	CompletedOrders int64
}

func addSectionHeader(pdf *gofpdf.Fpdf, title string) {
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(44, 62, 80)
	pdf.CellFormat(190, 10, title, "", 1, "L", false, 0, "")
	pdf.Ln(5)
}

func addSummarySection(pdf *gofpdf.Fpdf, data SummaryData) {
	addSectionHeader(pdf, "Summary")

	// Create a grid for key metrics
	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(249, 249, 249)

	// Row 1
	pdf.CellFormat(95, 20, fmt.Sprintf("Total Sales: %d", data.TotalSales), "1", 0, "L", true, 0, "")
	pdf.CellFormat(95, 20, fmt.Sprintf("Revenue: $%.2f", data.Revenue), "1", 1, "L", true, 0, "")

	// Row 2
	pdf.CellFormat(95, 20, fmt.Sprintf("Total Discounts: $%.2f", data.TotalDiscounts), "1", 0, "L", true, 0, "")
	pdf.CellFormat(95, 20, fmt.Sprintf("Completed Orders: %d", data.CompletedOrders), "1", 1, "L", true, 0, "")
}

func addPaymentMethodsSection(pdf *gofpdf.Fpdf, data PaymentData) {
	addSectionHeader(pdf, "Payment Methods")

	// Table header
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(240, 240, 240)

	// Header row
	pdf.CellFormat(47.5, 10, "COD", "1", 0, "C", true, 0, "")
	pdf.CellFormat(47.5, 10, "Razorpay", "1", 0, "C", true, 0, "")
	pdf.CellFormat(47.5, 10, "Wallet", "1", 0, "C", true, 0, "")
	pdf.CellFormat(47.5, 10, "Total", "1", 1, "C", true, 0, "")

	// Data row
	pdf.SetFont("Arial", "", 10)
	total := data.COD + data.Razorpay + data.Wallet
	pdf.CellFormat(47.5, 10, fmt.Sprintf("%d", data.COD), "1", 0, "C", false, 0, "")
	pdf.CellFormat(47.5, 10, fmt.Sprintf("%d", data.Razorpay), "1", 0, "C", false, 0, "")
	pdf.CellFormat(47.5, 10, fmt.Sprintf("%d", data.Wallet), "1", 0, "C", false, 0, "")
	pdf.CellFormat(47.5, 10, fmt.Sprintf("%d", total), "1", 1, "C", false, 0, "")
}

func addOrderStatusSection(pdf *gofpdf.Fpdf, data OrderStatusData) {
	addSectionHeader(pdf, "Order Status Breakdown")

	total := float64(data.Pending + data.Completed + data.Returned)

	// Table header
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(63.3, 10, "Status", "1", 0, "C", true, 0, "")
	pdf.CellFormat(63.3, 10, "Count", "1", 0, "C", true, 0, "")
	pdf.CellFormat(63.3, 10, "Percentage", "1", 1, "C", true, 0, "")

	// Data rows
	pdf.SetFont("Arial", "", 10)
	addOrderStatusRow(pdf, "Pending", data.Pending, total)
	addOrderStatusRow(pdf, "Completed", data.Completed, total)
	addOrderStatusRow(pdf, "Returned", data.Returned, total)
}

func addOrderStatusRow(pdf *gofpdf.Fpdf, status string, count int64, total float64) {
	percentage := (float64(count) / total) * 100
	pdf.CellFormat(63.3, 10, status, "1", 0, "C", false, 0, "")
	pdf.CellFormat(63.3, 10, fmt.Sprintf("%d", count), "1", 0, "C", false, 0, "")
	pdf.CellFormat(63.3, 10, fmt.Sprintf("%.1f%%", percentage), "1", 1, "C", false, 0, "")
}

func addFinancialDetailsSection(pdf *gofpdf.Fpdf, data FinancialData) {
	addSectionHeader(pdf, "Financial Details")

	pdf.SetFont("Arial", "", 12)
	pdf.SetFillColor(249, 249, 249)

	// Add financial rows
	addFinancialRow(pdf, "Total Order Amount:", fmt.Sprintf("$%.2f", data.TotalOrderAmount))
	addFinancialRow(pdf, "Total Discounts:", fmt.Sprintf("$%.2f", data.TotalDiscounts))
	addFinancialRow(pdf, "Coupon Deductions:", fmt.Sprintf("$%.2f", math.Abs(data.CouponDeductions)))

	// Final amount in bold
	pdf.SetFont("Arial", "B", 12)
	addFinancialRow(pdf, "Final Order Amount:", fmt.Sprintf("$%.2f", data.FinalOrderAmount))
}

func addFinancialRow(pdf *gofpdf.Fpdf, label, value string) {
	pdf.CellFormat(95, 10, label, "1", 0, "L", true, 0, "")
	pdf.CellFormat(95, 10, value, "1", 1, "L", true, 0, "")
}

func addFooter(pdf *gofpdf.Fpdf) {
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(127, 140, 141)
	pdf.CellFormat(190, 200,
		fmt.Sprintf("Generated on %s", time.Now().Format("January 2, 2006 15:04 MST")),
		"", 1, "L", false, 0, "")
}

// Generates and embeds pie and bar charts
func generateAndEmbedCharts(pdf *gofpdf.Fpdf, codCount, razorpayCount, walletCount, pendingCount, completedCount, returnedCount int64) error {
	// Generate Payment Method Pie Chart
	paymentChart := chart.PieChart{
		Width:  256,
		Height: 256,
		Values: []chart.Value{
			{Value: float64(codCount), Label: "COD"},
			{Value: float64(razorpayCount), Label: "Razorpay"},
			{Value: float64(walletCount), Label: "Wallet"},
		},
	}

	// Generate Order Status Pie Chart
	orderStatusChart := chart.PieChart{
		Width:  256,
		Height: 256,
		Values: []chart.Value{
			{Value: float64(pendingCount), Label: "Pending"},
			{Value: float64(completedCount), Label: "Completed"},
			{Value: float64(returnedCount), Label: "Returned"},
		},
	}

	// Encode and embed payment chart image
	buffer := new(bytes.Buffer)
	err := paymentChart.Render(chart.PNG, buffer)
	if err != nil {
		return err
	}
	pdf.RegisterImageOptionsReader("payment_chart", gofpdf.ImageOptions{ImageType: "PNG"}, buffer)
	pdf.ImageOptions("payment_chart", 15, 60, 80, 80, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")

	// Encode and embed order status chart image
	buffer.Reset()
	err = orderStatusChart.Render(chart.PNG, buffer)
	if err != nil {
		return err
	}
	pdf.RegisterImageOptionsReader("order_status_chart", gofpdf.ImageOptions{ImageType: "PNG"}, buffer)
	pdf.ImageOptions("order_status_chart", 110, 60, 80, 80, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")

	return nil
}
