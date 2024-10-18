package controllers

import (
	"github.com/Ukkenjijo/trendtrek/database"
	"github.com/Ukkenjijo/trendtrek/models"
	"github.com/Ukkenjijo/trendtrek/utils"
	"github.com/gofiber/fiber/v2"
)

func ListAddresses(c *fiber.Ctx) error {
	userId := c.Locals("user_id") // Get user ID from the context

	var addresses []models.Address
	if err := database.DB.Where("user_id = ?", userId).Find(&addresses).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve addresses"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Addresses retrieved successfully",
		"data":    addresses,
	})
}

func AddAddress(c *fiber.Ctx) error {
	userId := c.Locals("user_id") // Get user ID from the context

	// Parse request body
	address := new(models.Address)
	if err := c.BodyParser(address); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate request
	if err := utils.ValidateStruct(address); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// If this is the default address, set all other addresses to non-default
	if address.IsDefault {
		if err := database.DB.Model(&models.Address{}).Where("user_id = ?", userId).Update("is_default", false).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update default status"})
		}
	}

	// Assign the user ID to the new address
	address.UserID = uint(userId.(float64))

	// Save the address to the database
	if err := database.DB.Create(&address).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add address"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Address added successfully",
		"data":    address,
	})
}

func EditAddress(c *fiber.Ctx) error {
    userId := c.Locals("user_id") // Get user ID from the context
    addressId := c.Params("id") // Get the address ID from the URL

    // Parse request body
    updatedAddress := new(models.Address)
    if err := c.BodyParser(updatedAddress); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
    }

    // Find the address in the database
    var address models.Address
    if err := database.DB.Where("id = ? AND user_id = ?", addressId, userId).First(&address).Error; err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Address not found"})
    }

    // Update the address fields only if they are provided in the request body
    if updatedAddress.Street != "" {
        address.Street = updatedAddress.Street
    }
    if updatedAddress.City != "" {
        address.City = updatedAddress.City
    }
    if updatedAddress.State != "" {
        address.State = updatedAddress.State
    }
    if updatedAddress.Country != "" {
        address.Country = updatedAddress.Country
    }
    if updatedAddress.ZipCode != "" {
        address.ZipCode = updatedAddress.ZipCode
    }
    // Handle default address change
    if updatedAddress.IsDefault {
        if err := database.DB.Model(&models.Address{}).Where("user_id = ?", userId).Update("is_default", false).Error; err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update default status"})
        }
        address.IsDefault = true
    }

    // Save the updated address
    if err := database.DB.Save(&address).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update address"})
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "message": "Address updated successfully",
        "data":    address,
    })
}

func DeleteAddress(c *fiber.Ctx) error {
    userId := c.Locals("user_id") // Get user ID from the context
    addressId := c.Params("id") // Get the address ID from the URL

    // Find the address in the database
    var address models.Address
    if err := database.DB.Where("id = ? AND user_id = ?", addressId, userId).First(&address).Error; err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Address not found"})
    }

    // Delete the address
    if err := database.DB.Delete(&address).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete address"})
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "message": "Address deleted successfully",
    })
}

