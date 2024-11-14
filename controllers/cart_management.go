package controllers

import (
	"fmt"
	"log"
	"math"
	

	"github.com/Ukkenjijo/trendtrek/database"
	"github.com/Ukkenjijo/trendtrek/models"
	"github.com/Ukkenjijo/trendtrek/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AddToCart(c *fiber.Ctx) error {
	userId := c.Locals("user_id")

	// Parse request body
	cartItemRequest := new(struct {
		ProductID uint `json:"product_id" validate:"required"`
		Quantity  int  `json:"quantity" validate:"required,gte=1"`
	})
	log.Println(cartItemRequest)
	if err := c.BodyParser(cartItemRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	log.Println(cartItemRequest)

	// Validate request
	if err := utils.ValidateStruct(cartItemRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Check if the user has an existing cart
	var cart models.Cart
	if err := database.DB.Where("user_id = ?", userId).First(&cart).Error; err != nil {
		// If no cart found, create a new one
		cart = models.Cart{UserID: uint(userId.(float64))}
		if err := database.DB.Create(&cart).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create cart"})
		}
	}

	// Find the product in the database with the offers
	var product models.Product
	if err := database.DB.Preload("Offer").First(&product, cartItemRequest.ProductID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
	}
	//Calculate the offer based on the product offer
	originalPrice := product.Price
	discountedPrice := originalPrice
	var discountPercentage *float64

	if product.Offer != nil && product.Offer.DiscountPercentage > 0 {

		discountPercentage = &product.Offer.DiscountPercentage
		discountedPrice = originalPrice * (1 - *discountPercentage/100)
	}

	// Check stock availability
	if cartItemRequest.Quantity > product.StockQuantity {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Not enough stock available",
		})
	}

	// Check maximum quantity per user
	maxQuantityPerUser := 5 // For example, limit each user to 5 units of each product
	if cartItemRequest.Quantity > maxQuantityPerUser {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Maximum %d of this product can be added to cart", maxQuantityPerUser),
		})
	}

	// Calculate total price
	totalPrice := float64(cartItemRequest.Quantity) * discountedPrice

	// Check if the product is already in the user's cart
	var existingCartItem models.CartItem
	if err := database.DB.Where("cart_id = ? AND product_id = ?", cart.ID, cartItemRequest.ProductID).First(&existingCartItem).Error; err == nil {
		// Update the quantity of the existing cart item
		existingCartItem.Quantity += cartItemRequest.Quantity
		if existingCartItem.Quantity > maxQuantityPerUser {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Cannot exceed %d of this product in your cart", maxQuantityPerUser),
			})
		}
		if existingCartItem.Quantity > product.StockQuantity {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Not enough stock available"})
		}

		// Update the total price for the item
		existingCartItem.TotalPrice = float64(existingCartItem.Quantity) * discountedPrice

		// Save the updated cart item
		if err := database.DB.Save(&existingCartItem).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update cart item"})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Cart item updated successfully"})
	}

	// Create a new cart item
	newCartItem := models.CartItem{
		CartID:             cart.ID,
		ProductID:          cartItemRequest.ProductID,
		Quantity:           cartItemRequest.Quantity,
		Price:              product.Price,
		DiscountedPrice:    discountedPrice,
		DiscountPercentage: discountPercentage,
		TotalPrice:         totalPrice,
	}
	if err := database.DB.Create(&newCartItem).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add to cart"})
	}

	var newcart models.Cart
	if err := database.DB.Preload("Items").Where("user_id = ?", userId).First(&newcart).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Cart not found"})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Product added to cart"})
}

func ListCartItems(c *fiber.Ctx) error {
	userId := c.Locals("user_id")

	var cart models.Cart
	if err := database.DB.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Preload("Product",func(db *gorm.DB) *gorm.DB {
			return db.Preload("Images")
		})
	}).Where("user_id = ?", userId).First(&cart).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Cart not found"})
	}

	//calculate the total price before discount
	var totalAmount, product_discount float64
	for _, item := range cart.Items {
		totalAmount += item.TotalPrice
		product_discount += (item.Price - item.DiscountedPrice) * float64(item.Quantity)
	}

	cart.CartTotal = totalAmount

	if err := database.DB.Save(&cart).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update cart"})
	}

	//Initalize the discount and total amount
	var discount float64
	finalamount := totalAmount
	//Apply a coupon if there is one
	if cart.CouponID != nil {
		var coupon models.Coupon
		if err := database.DB.First(&coupon, cart.CouponID).Error; err == nil {
			//Check minimum purchase requirment
			if totalAmount >= coupon.MinPurchaseAmount {
				discount = (totalAmount * coupon.Discount / 100)

			} else {
				discount = coupon.MaxDiscountAmount
			}
			finalamount = totalAmount - discount

			cart.CartTotal = finalamount
			cart.CouponDiscount = discount
			if err := database.DB.Save(&cart).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update cart"})
			}
		} else {
			//If minimum purchse amount is not met ignore coupon
			cart.CouponID = nil
			cart.CartTotal = totalAmount
			if err := database.DB.Save(&cart).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update cart"})
			}
		}
	}

	//create a response struct to display cart
	var itemsResponse []fiber.Map
	for _, item := range cart.Items {

		itemsResponse = append(itemsResponse, fiber.Map{
			"id":            item.ID,
			"product_id":    item.ProductID,
			"quantity":      item.Quantity,
			"price":         fmt.Sprintf("%.2f", item.Price),
			"total_price":   fmt.Sprintf("%.2f", item.TotalPrice),
			"total_discount": math.RoundToEven(item.Price-item.DiscountedPrice) * float64(item.Quantity),
			"product_name":  item.Product.Name,
			"product_image": item.Product.Images[0].URL,
		})

	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"items":            itemsResponse,
		"total_amount":     fmt.Sprintf("%.2f", cart.CartTotal),
		"coupon_discount":  fmt.Sprintf("%.2f", cart.CouponDiscount),
		"toatl_product_discounts": fmt.Sprintf("%.2f", product_discount),
		"total_items":      len(cart.Items),
	})
}

func UpdateCartQuantity(c *fiber.Ctx) error {
	userId := c.Locals("user_id")
	productId := c.Params("id")

	// Parse the request body
	cartItemRequest := new(struct {
		Quantity int `json:"quantity" validate:"required,gte=1"`
	})
	if err := c.BodyParser(cartItemRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate the request
	if err := utils.ValidateStruct(cartItemRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Find the user's cart
	var cart models.Cart
	if err := database.DB.Where("user_id = ?", userId).First(&cart).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Cart not found"})
	}

	// Find the product and cart item
	var product models.Product
	if err := database.DB.Preload("Offer").First(&product, productId).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
	}

	var cartItem models.CartItem
	if err := database.DB.Where("cart_id = ? AND product_id = ?", cart.ID, productId).First(&cartItem).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found in cart"})
	}

	// Check if the new quantity exceeds stock or max quantity per user
	if cartItemRequest.Quantity > product.StockQuantity {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Not enough stock available"})
	}
	maxQuantityPerUser := 5
	if cartItemRequest.Quantity > maxQuantityPerUser {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Maximum %d of this product can be added to cart", maxQuantityPerUser),
		})
	}

	//Calculate the offer based on the product offer
	originalPrice := product.Price
	discountedPrice := originalPrice
	var discountPercentage *float64

	if product.Offer != nil && product.Offer.DiscountPercentage > 0 {

		discountPercentage = &product.Offer.DiscountPercentage
		discountedPrice = originalPrice * (1 - *discountPercentage/100)
	}

	// Update the cart item's quantity and total price
	cartItem.Quantity = cartItemRequest.Quantity
	cartItem.TotalPrice = float64(cartItem.Quantity) * discountedPrice

	if err := database.DB.Save(&cartItem).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update cart item"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Cart item quantity updated successfully"})
}

func RemoveFromCart(c *fiber.Ctx) error {
	userId := c.Locals("user_id")
	productId := c.Params("id")

	// Find the user's cart
	var cart models.Cart
	if err := database.DB.Where("user_id = ?", userId).First(&cart).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Cart not found"})
	}

	// Find the cart item in the database
	var cartItem models.CartItem
	if err := database.DB.Where("cart_id = ? AND product_id = ?", cart.ID, productId).First(&cartItem).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found in cart"})
	}

	// Remove the cart item
	if err := database.DB.Delete(&cartItem).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to remove product from cart"})
	}
	//check if there are any more items in the cart
	var remainingItems int64
	if err := database.DB.Model(&models.CartItem{}).Where("cart_id = ?", cart.ID).Count(&remainingItems).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to remove product from cart"})
	}
	if remainingItems == 0 {
		if err := database.DB.Delete(&cart).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to remove product from cart"})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Product removed from cart"})
}
