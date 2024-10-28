package controllers

import (
	"github.com/Ukkenjijo/trendtrek/database"
	"github.com/Ukkenjijo/trendtrek/models"
	"github.com/gofiber/fiber/v2"
	
)

// AddToWishlist adds a product to the user's wishlist
func AddToWishlist(c *fiber.Ctx) error {
	db := database.DB
	userID := c.Locals("user_id")

	productID, err := c.ParamsInt("product_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Invalid input",
		})
	}
	//check if the product already exists in the wishlist

	var wishlistItem models.WishlistItem
	if err := db.Where("user_id = ? AND product_id = ?", userID, productID).First(&wishlistItem).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Product already exists in wishlist",
		})
	}

	if err := db.Create(&models.WishlistItem{
		UserID:    uint(userID.(float64)),
		ProductID: uint(productID),
	}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to add item to wishlist",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  true,
		"message": "Item added to wishlist successfully",
	})
}

// RemoveFromWishlist removes a product from the user's wishlist
func RemoveFromWishlist(c *fiber.Ctx) error {
	db := database.DB
	userID, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  false,
			"message": "Unauthorized access",
		})
	}

	productID, err := c.ParamsInt("product_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Invalid product ID",
		})
	}

	var wishlistItem models.WishlistItem
	if err := db.Where("user_id = ? AND product_id = ?", uint(userID), uint(productID)).First(&wishlistItem).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  false,
			"message": "Item not found in wishlist",
		})
	}

	if err := db.Delete(&wishlistItem).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to remove item from wishlist",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  true,
		"message": "Item removed from wishlist successfully",
	})
}

// GetWishlist retrieves the user's wishlist
// GetWishlist fetches all wishlist items for a user
// GetWishlist retrieves the user's wishlist
// GetWishlist fetches all wishlist items for a user
func GetWishlist(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64) // Retrieve user ID from JWT or session context

	// Fetch wishlist items for the user
	var wishlist []models.WishlistItem
	if err := database.DB.Preload("Product").Preload("Product.Images").Preload("Product.Category").Preload("Product.Store").Where("user_id = ?", uint(userID)).Find(&wishlist).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to fetch wishlist",
		})
	}

	// If the wishlist is empty, return an empty array
	if len(wishlist) == 0 {
		return c.Status(fiber.StatusOK).JSON(models.WishlistResponse{
			Status:  true,
			Message: "Add something to the wishlist",
			Wishlist: []models.ProductResponse{},
		})
	}

	// Map wishlist items to the custom response struct
	var productResponses []models.ProductResponse
	for _, item := range wishlist {
		productResponse := models.ProductResponse{
			ID:            item.Product.ID,
			Name:          item.Product.Name,
			Description:   item.Product.Description,
			Price:         item.Product.Price,
			StockQuantity: item.Product.StockQuantity,
			IsActive:      item.Product.IsActive,
			Category: models.CategoryResponse{
				ID:   item.Product.Category.ID,
				Name: item.Product.Category.Name,
			},
			Store: models.StoreResponse{
				ID:   item.Product.Store.ID,
				Name: item.Product.Store.Name,
			},
			Images: make([]string, len(item.Product.Images)),
		}

		for i, image := range item.Product.Images {
			productResponse.Images[i] = image.URL
		}

		productResponses = append(productResponses, productResponse)
	}

	return c.Status(fiber.StatusOK).JSON(models.WishlistResponse{
		Status:  true,
		Message: "Wishlist fetched successfully",
		Wishlist: productResponses,
	})

	
}
