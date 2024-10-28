package controllers

import (
	"github.com/Ukkenjijo/trendtrek/database"
	"github.com/Ukkenjijo/trendtrek/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func CreateOrUpdateOffer(c *fiber.Ctx) error {
	// Extract product ID from the URL
	productID := c.Params("product_id")

	// Retrieve seller ID from JWT or session context
	userID := c.Locals("user_id")
	sellerID, _ := GetStoreIDByUserID(uint(userID.(float64)))

	// Parse the discount percentage from the request body
	var req struct {
		DiscountPercentage float64 `json:"discount_percentage" validate:"required,gte=0,lte=100"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Fetch the product and ensure the seller owns it
	var product models.Product
	if err := database.DB.Where("id = ? AND store_id = ?", productID, sellerID).First(&product).Error; err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Product not found or not authorized"})
	}

	// Check if an offer already exists for this product
	var offer models.Offer
	if err := database.DB.Where("product_id = ?", product.ID).First(&offer).Error; err != nil {
		// If no existing offer, create a new one
		if err == gorm.ErrRecordNotFound {
			offer = models.Offer{
				ProductID:          product.ID,
				DiscountPercentage: req.DiscountPercentage,
			}
			if err := database.DB.Create(&offer).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create offer"})
			}
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
		}
	} else {
		// Update existing offer
		offer.DiscountPercentage = req.DiscountPercentage
		if err := database.DB.Save(&offer).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update offer"})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Offer applied successfully", "offer": offer})
}

// DeleteOffer removes an offer from a product
func DeleteOffer(c *fiber.Ctx) error {
	// Extract product ID and seller ID
	productID := c.Params("product_id")
	userId:= c.Locals("user_id")
	sellerID, _ := GetStoreIDByUserID(uint(userId.(float64)))

	// Fetch product to confirm seller ownership
	var product models.Product
	if err := database.DB.Where("id = ? AND store_id = ?", productID, sellerID).First(&product).Error; err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Product not found or not authorized"})
	}

	// Delete the offer associated with this product
	if err := database.DB.Where("product_id = ?", product.ID).Delete(&models.Offer{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete offer"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Offer removed successfully"})
}
// ListOffers returns all offers created by a seller
func ListOffers(c *fiber.Ctx) error {
	//get seller id form user id
	sellerID, _ := GetStoreIDByUserID(uint(c.Locals("user_id").(float64)))

	var offers []models.Offer
	if err := database.DB.Joins("JOIN products ON products.id = offers.product_id").
		Where("products.store_id = ?", sellerID).
		Find(&offers).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve offers"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"offers": offers})
}
