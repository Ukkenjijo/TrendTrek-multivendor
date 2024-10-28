package controllers

import (
	"fmt"
	"log"

	"github.com/Ukkenjijo/trendtrek/database"
	"github.com/Ukkenjijo/trendtrek/models"
	"github.com/Ukkenjijo/trendtrek/utils"
	"github.com/gofiber/fiber/v2"
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
        CartID:    cart.ID,
        ProductID: cartItemRequest.ProductID,
        Quantity:  cartItemRequest.Quantity,
        Price:     product.Price,
        DiscountedPrice: discountedPrice,
        DiscountPercentage: discountPercentage,
        TotalPrice: totalPrice,
    }
    if err := database.DB.Create(&newCartItem).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add to cart"})
    }

    var newcart models.Cart
    if err := database.DB.Preload("Items").Where("user_id = ?", userId).First(&newcart).Error; err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Cart not found"})
    }
    
    // Calculate the total cart value
    var cartTotal float64 = 0
    for _, item := range newcart.Items {
        cartTotal += item.TotalPrice
        log.Println("total1",cartTotal)
    }
    newcart.CartTotal = cartTotal-cart.CouponDiscount
    log.Println("total",newcart.CartTotal,cartTotal)

    if err := database.DB.Save(&newcart).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create total of cart"})
    }


    return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Product added to cart"})
}

func ListCartItems(c *fiber.Ctx) error {
    userId := c.Locals("user_id")

    var cart models.Cart
    if err := database.DB.Preload("Items").Where("user_id = ?", userId).First(&cart).Error; err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Cart not found"})
    }

    // Calculate the total cart value
    var TotalDiscount,cartOrginal float64 = 0.0,0.0
    for _, item := range cart.Items {
        cartOrginal += item.Price*float64(item.Quantity)
        TotalDiscount+= (item.Price- item.DiscountedPrice)
    }
    order_total:=cart.CartTotal-cart.CouponDiscount

    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "message": "Cart retrieved successfully",
        "data":    cart.Items,    // Send just the items to the client
        "cart_total": cartOrginal,  // Send the total value of the cart
        "coupon_discount": cart.CouponDiscount,
        "product_discount": TotalDiscount,
        "order_total": order_total,
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

    return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Product removed from cart"})
}

