package controllers

import (
	"fmt"

	"github.com/Ukkenjijo/trendtrek/database"
	"github.com/Ukkenjijo/trendtrek/models"
	"github.com/gofiber/fiber/v2"
)

func ListSellerOrders(c *fiber.Ctx) error {
	// Get the user id from the token
	userId := c.Locals("user_id")
	// Get store id from user id
	storeId, _ := GetStoreIDByUserID(uint(userId.(float64)))

	// Fetch the orders that contain the store id
	var orders []models.Order
	if err := database.DB.Preload("Items", "product_id IN (SELECT id FROM products WHERE store_id = ?)", storeId).Find(&orders).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve orders"})
	}

	// Create a response for each order including its items
	var orderResponses []fiber.Map
	for _, order := range orders {
		orderResponse := fiber.Map{
			"order_id":       order.ID,
			"user_id":        order.UserID,
			"total_amount":   fmt.Sprintf("%.2f", order.TotalAmount),
			"status":         order.Status,
			"shipping_city":  order.ShippingCity,
			"shipping_state": order.ShippingState,
			"payment_mode":   order.PaymentMode,
			"items":          make([]fiber.Map, len(order.Items)),
			"payment_status": make([]fiber.Map, len(orders)),
		}

		// Add the payment status of each order
		var payment models.Payment
		if err := database.DB.Where("order_id = ?", order.ID).First(&payment).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve payment"})
		}
		orderResponse["payment_status"] = payment.PaymentStatus

		for i, item := range order.Items {
			orderResponse["items"].([]fiber.Map)[i] = fiber.Map{
				"product_id":  item.ProductID,
				"quantity":    item.Quantity,
				"total_price": fmt.Sprintf("%.2f", item.TotalPrice),
				"status":      item.Status,
			}
		}
		orderResponses = append(orderResponses, orderResponse)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Orders retrieved successfully",
		"orders":  orderResponses,
	})
}

func UpdateOrderItemStatus(c *fiber.Ctx) error {
	//get the order id
	orderId := c.Params("order_id")
	//get the item id
	itemId := c.Params("item_id")
	//degine the struct for status request

	//get store id from user id
	storeId, _ := GetStoreIDByUserID(uint(c.Locals("user_id").(float64)))
	req := new(models.StatusRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	//Fetch the order item from the order
	var orderItem models.OrderItem
	if err := database.DB.Where("order_id = ? AND id = ?", orderId, itemId).Where("product_id IN (SELECT id FROM products WHERE store_id = ?)", storeId).First(&orderItem).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve order item"})
	}

	//Update the order item status
	orderItem.Status = req.Status
	if err := database.DB.Save(&orderItem).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update order item"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Order item status updated successfully"})

}
