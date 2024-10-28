package controllers

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Ukkenjijo/trendtrek/database"
	"github.com/Ukkenjijo/trendtrek/models"
	"github.com/Ukkenjijo/trendtrek/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/razorpay/razorpay-go"
	"gorm.io/gorm"
)

// Initialize Razorpay Client
var razorpayClient *razorpay.Client

func InitRazorpay() {
	razorpayClient = razorpay.NewClient(os.Getenv("RAZORPAY_KEY"), os.Getenv("RAZORPAY_SECRET"))
}

// PlaceOrder function with Razorpay payment integration
func PlaceOrder(c *fiber.Ctx) error {
	userId := c.Locals("user_id")

	// Parse the request body
	req := new(models.OrderRequest)

	if err := c.BodyParser(req); err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}
	log.Println(req)

	// Validate the request
	if err := utils.ValidateStruct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Get the user's cart
	var cart models.Cart
	if err := database.DB.Preload("Items").Where("user_id = ?", userId).First(&cart).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Cart not found"})
	}
	addressID, _ := strconv.ParseUint(req.AddressID, 10, 32)

	// Get the address snapshot
	var address models.Address
	if err := database.DB.Where("id = ? AND user_id = ?", addressID, userId).First(&address).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Address not found"})
	}

	// Ensure the cart is not empty
	if len(cart.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cart is empty"})
	}

	// Begin transaction
	tx := database.DB.Begin()
	defer tx.Rollback()

	// Check stock availability and calculate total amount
	var totalAmount float64=cart.CartTotal-cart.CouponDiscount
	// for _, item := range cart.Items {
	// 	var product models.Product
	// 	if err := tx.First(&product, item.ProductID).Error; err != nil {
	// 		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
	// 	}

	// 	if product.StockQuantity < item.Quantity {
	// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("Not enough stock for product %s", product.Name)})
	// 	}
	// 	totalAmount += item.TotalPrice
	// 	product.StockQuantity -= item.Quantity

	// 	if err := tx.Save(&product).Error; err != nil {
	// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update stock"})
	// 	}
	// }
	
	

	// Create the order in the database
	order := models.Order{
		UserID:          uint(userId.(float64)),
		TotalAmount:     totalAmount,
		PaymentMode:     req.PaymentMode,
		Status:          "pending",
		ShippingStreet:  address.Street,
		ShippingCity:    address.City,
		ShippingState:   address.State,
		ShippingCountry: address.Country,
		ShippingZipCode: address.ZipCode,
	}

	if err := tx.Create(&order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create order"})
	}
     cartOrginal,TotalDiscount:=0.0,0.0
	// Create the order items
	for _, item := range cart.Items {
		orderItem := models.OrderItem{
			OrderID:    order.ID,
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			Price:      item.Price,
			TotalPrice: item.TotalPrice,
		}
		cartOrginal += item.Price
        TotalDiscount+= (item.Price- item.DiscountedPrice)
		if err := tx.Create(&orderItem).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create order item"})
		}
	}
	//Create the order details
	var orderPaymentDetail models.OrderPaymentDetail
	orderPaymentDetail.OrderID = order.ID
	orderPaymentDetail.PaymentType = req.PaymentMode
	orderPaymentDetail.OrderAmount = cartOrginal
	orderPaymentDetail.OrderDiscount = TotalDiscount
	orderPaymentDetail.CouponSavings = cart.CouponDiscount
	orderPaymentDetail.FinalOrderAmount = totalAmount

	// Clear the cart
	if err := tx.Delete(&cart).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to clear cart"})
	}

	// Create the payment
	var payment models.Payment
	payment.OrderID = order.ID
	payment.PaymentType = req.PaymentMode
	payment.Amount = totalAmount
	payment.PaymentStatus = "pending"
	

	

	InitRazorpay()

	// If PaymentMode is Razorpay, create a Razorpay order
	if req.PaymentMode == "razorpay" {
		// Razorpay expects the amount in paise (so multiply by 100)
		amount := int(totalAmount * 100)

		// Prepare Razorpay order options
		options := map[string]interface{}{
			"amount":          amount,
			"currency":        "INR",
			"receipt":         fmt.Sprintf("order_%d", order.ID),
			"payment_capture": 1,
		}

		// Create Razorpay order
		razorpayOrder, err := razorpayClient.Order.Create(options, nil)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create Razorpay order"})
		}
		//set the payment status for razorpay
		payment.RazorpayPaymentID = razorpayOrder["id"].(string)

		if err := tx.Create(&payment).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create payment"})
		}
		
		if err := tx.Create(&orderPaymentDetail).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create payment"})
		}

		// Commit the transaction and return the Razorpay order ID
		if err := tx.Commit().Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to commit transaction"})
		}

		// Return Razorpay order details
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":           "Order placed successfully",
			"order_id":          order.ID,
			"razorpay_order_id": razorpayOrder["id"],
			"amount":            totalAmount,
			"currency":          "INR",
		})
	}
	if err := tx.Create(&payment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create payment"})
	}
	

	// Commit the transaction for non-Razorpay payments
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to commit transaction"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "Order placed successfully",
		"order_id": order.ID,
	})
}

func ListOrders(c *fiber.Ctx) error {
	userId := c.Locals("user_id")
	var orders []models.Order
	if err := database.DB.Preload("Items").Where("user_id = ?", userId).Find(&orders).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve orders"})
	}

	var orderResponses []fiber.Map
	for _, order := range orders {
		orderResponse := fiber.Map{
			"order_id":       order.ID,
			"total_amount":   order.TotalAmount,
			"status":         order.Status,
			"shipping_city":  order.ShippingCity,
			"shipping_state": order.ShippingState,
			"payment_mode":   order.PaymentMode,
			"items":          make([]fiber.Map, len(order.Items)),
		}
		for i, item := range order.Items {
			orderResponse["items"].([]fiber.Map)[i] = fiber.Map{
				"product_id":  item.ProductID,
				"quantity":    item.Quantity,
				"total_price": item.TotalPrice,
			}
		}
		orderResponses = append(orderResponses, orderResponse)
	}
	if len(orderResponses)==0{
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "No orders found"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Orders retrieved successfully",
		"orders":  orderResponses,
	})
}

func CancelOrder(c *fiber.Ctx) error {
	userId := c.Locals("user_id")
	orderId, _ := c.ParamsInt("id")

	// Find the order
	var order models.Order
	tx := database.DB.Begin()
	defer tx.Rollback()

	if err := tx.Where("id = ? AND user_id = ?", orderId, userId).First(&order).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order not found"})
	}

	// Ensure the order can still be canceled (e.g., only "pending" orders can be canceled)
	if order.Status != "pending" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Order cannot be canceled"})
	}

	// Update the order status to "canceled"
	order.Status = "canceled"
	if err := tx.Save(&order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to cancel order"})
	}

	// Return stock back to the products
	for _, item := range order.Items {
		var product models.Product
		if err := tx.First(&product, item.ProductID).Error; err == nil {
			product.StockQuantity += item.Quantity
			if err := tx.Save(&product).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to return stock"})
			}
		}
		//set status to canceled
		item.Status = "canceled"
		if err := tx.Save(&item).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to cancel order item"})
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Transaction failed"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "Order canceled successfully",
		"order_id": order.ID,
	})
}

func CancelOrderItem(c *fiber.Ctx) error {
	userId := c.Locals("user_id")
	orderId, _ := c.ParamsInt("order_id")
	itemId, _ := c.ParamsInt("item_id")

	// Find the order
	var order models.Order
	tx := database.DB.Begin()
	defer tx.Rollback()

	if err := tx.Where("id = ? AND user_id = ?", orderId, userId).First(&order).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order not found"})
	}

	// Find the order item
	var orderItem models.OrderItem
	if err := tx.Where("order_id = ? AND product_id = ?", orderId, itemId).First(&orderItem).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order item not found"})
	}
	//Only pending items can be cancelled
	if orderItem.Status != "pending" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Order item cannot be canceled"})
	}

	// Update the order item status to "canceled"
	orderItem.Status = "canceled"
	if err := tx.Save(&orderItem).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to cancel order item"})
	}

	// Return stock back to the product
	var product models.Product
	if err := tx.First(&product, orderItem.ProductID).Error; err == nil {
		product.StockQuantity += orderItem.Quantity
		if err := tx.Save(&product).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to return stock"})
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Transaction failed"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "Order item canceled successfully",
		"order_id": order.ID,
	})

}

func GetOrderDetails(c *fiber.Ctx) error {
	orderId := c.Params("id")
	var order models.Order
	tx := database.DB.Begin()
	defer tx.Rollback()

	if err := tx.Preload("Items").First(&order, orderId).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order not found"})
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Transaction failed"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Order details retrieved successfully",
		"order":   order,
	})
}

type ReturnRequest struct {
	Reason string `json:"reason" validate:"required"` // Reason for returning the item
}

func ReturnOrderItem(c *fiber.Ctx) error {
	// Step 1: Get the order item ID from the URL path
	orderItemID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid order item ID"})
	}

	// Step 2: Parse the request body to get the return reason
	req := new(ReturnRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Step 3: Find the order item by ID in the database
	var orderItem models.OrderItem
	if err := database.DB.First(&orderItem, uint(orderItemID)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order item not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to query order item"})
	}

	// Step 4: Check if the order item is eligible for return (e.g., within return window)
	var order models.Order
	if err := database.DB.First(&order, orderItem.OrderID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to query order"})
	}

	// Assume the return window is 30 days from order creation
	returnWindow := order.CreatedAt.AddDate(0, 0, 30)
	if time.Now().After(returnWindow) {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "Return window has expired"})
	}

	// Step 5: Update the order item's status to "returned" and set the return reason
	orderItem.Status = "returned"
	orderItem.ReturnReason = req.Reason
	orderItem.ReturnedAt = time.Now()

	if err := database.DB.Save(&orderItem).Error; err != nil {
		log.Printf("Error updating order item: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update order item"})
	}

	// Step 6: Respond with success message and updated order item details
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":    "Item returned successfully",
		"order_item": orderItem,
	})
}
