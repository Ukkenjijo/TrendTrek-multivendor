package controllers

import (
	"errors"
	"fmt"
	"time"

	"github.com/Ukkenjijo/trendtrek/database"
	"github.com/Ukkenjijo/trendtrek/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func GetStoreProfile(c *fiber.Ctx) error {
	user := new(models.StoreProfileResponse)
	// Get the user ID from the context
	userID := c.Locals("user_id").(float64)
	sellerID, _ := GetStoreIDByUserID(uint(userID))
	// Get the user profile from the database
	if err := database.DB.Model(&models.Store{}).Select([]string{"id", "name", "description", "address", "city", "state", "country", "store_image"}).First(&user, float64(sellerID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Return 404 Not Found
			c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Store not found in database"})
		} else {
			// Return 500 Internal Server Error
			c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database Error"})
		}
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Storeprofile retrived Succesfully", "data": user})

}

func UpdateStoreProfile(c *fiber.Ctx) error {
	// Get the user ID from the context
	userID := c.Locals("user_id").(float64)
	sellerID, _ := GetStoreIDByUserID(uint(userID))
	// Get the user profile from the database
	var dbUser models.Store
	if err := database.DB.First(&dbUser, float64(sellerID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found in database"})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database Error"})
		}
	}

	// Parse multipart form data
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to parse multipart form"})
	}

	// Update user fields from the form
	name := form.Value["name"]
	description := form.Value["description"]
	Address := form.Value["address"]
	city := form.Value["city"]
	state := form.Value["state"]
	country := form.Value["country"]

	// Update the fields in the database user object
	if len(name) > 0 {
		dbUser.Name = name[0]
	}
	if len(description) > 0 {
		dbUser.Description = description[0]
	}
	if len(Address) > 0 {
		dbUser.Address = Address[0]
	}
	if len(city) > 0 {
		dbUser.City = city[0]
	}

	if len(state) > 0 {
		dbUser.State = state[0]
	}
	if len(country) > 0 {
		dbUser.Country = country[0]
	}

	// Handle the store image upload
	storeImage, err := c.FormFile("store_image")
	if err == nil {
		// Generate a unique file name for the image
		timestamp := time.Now().Unix()
		storeImageName := fmt.Sprintf("%d_%d_%s", int(sellerID), timestamp, storeImage.Filename)

		// Define the path to save the image (adjust the path as necessary)
		storeImagePath := "./uploads/store_images/" + storeImageName

		// Save the file to the specified directory
		if err := c.SaveFile(storeImage, storeImagePath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save store image"})
		}

		// Store the file path in the database
		dbUser.StoreImage = fmt.Sprintf("http://localhost:3000/uploads/store_images/%s", storeImageName)
	}

	// Save the updated user to the database
	if err := database.DB.Save(&dbUser).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update user profile"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Store profile updated successfully",
		"data": fiber.Map{
			"name":        dbUser.Name,
			"description": dbUser.Description,
			"address":     dbUser.Address,
			"city":        dbUser.City,
			"state":       dbUser.State,
			"country":     dbUser.Country,
			"store_image": dbUser.StoreImage,
		},
	})
}


func GetAllStores(c *fiber.Ctx) error {
	var stores []models.StoreProfileResponse
	if err := database.DB.Model(&models.Store{}).Select([]string{"id", "name", "description", "address", "city", "state", "country", "store_image"}).Find(&stores).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "No stores found in database"})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database Error"})
		}
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Stores retrieved successfully", "data": stores})
}