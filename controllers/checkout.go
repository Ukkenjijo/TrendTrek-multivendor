package controllers

import (
	"fmt"
	"log"

	"github.com/Ukkenjijo/trendtrek/database"
	"github.com/Ukkenjijo/trendtrek/models"
	"github.com/Ukkenjijo/trendtrek/utils"
	"github.com/gofiber/fiber/v2"
)

func PlaceOrder(c *fiber.Ctx) error {
	userId := c.Locals("user_id")

	//parse the request body
	req := new(models.OrderRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}
	log.Println(req)

	//Validate the request
	if err := utils.ValidateStruct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	//Get the users cart
	var cart models.Cart
	if err := database.DB.Preload("Items").Where("user_id = ?", userId).First(&cart).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Cart not found"})
	}

	//get the address sanpshot
	var address models.Address
	if err := database.DB.Where("id = ? AND user_id = ?", req.AddressID, userId).First(&address).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Address not found"})
	}

	//Ensure the cart is not empty
	if len(cart.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cart is empty"})
	}
	log.Println(cart)
	//Begin transaction
	tx := database.DB.Begin()
	defer tx.Rollback()

	//Check stock availability for each item
	var TotalAmount float64
	for _, item := range cart.Items {
		var product models.Product
		if err := tx.First(&product, item.ProductID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
		}
		log.Println("Stock Quantity:", product.StockQuantity)
		if product.StockQuantity < item.Quantity {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("Not enough stock for product %s", product.Name)})
		}
		//calculate the total amount for the order
		TotalAmount += item.TotalPrice
		//Deduct the quantity from the stock
		product.StockQuantity -= item.Quantity

		if err := tx.Save(&product).Error; err != nil {
			log.Println("New Stock Quantity:", product.StockQuantity)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update stock"})
		}

	}
	log.Println(TotalAmount)

	
	// Create an order with the user's details and order information
	order := models.Order{
		UserID:      uint(userId.(float64)),
		TotalAmount: TotalAmount,
		PaymentMode: req.PaymentMode,
		Status:      "pending",
		ShippingStreet:  address.Street,
		ShippingCity:    address.City,
		ShippingState:   address.State,
		ShippingCountry: address.Country,
		ShippingZipCode: address.ZipCode,
	}

	if err := tx.Create(&order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create order"})
	}
	//create the order items
	for _, item := range cart.Items {
		orderItem := models.OrderItem{
			OrderID:    order.ID,
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			TotalPrice: item.TotalPrice,
		}
		if err := tx.Create(&orderItem).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create order item"})
		}
	}
	// Clear the user's cart when the order is placed
	if err := tx.Where("id = ?", cart.ID).Delete(&models.Cart{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to clear cart"})
	}

	//Commit the transaction
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
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Orders retrieved successfully",
		"orders":  orders,
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
