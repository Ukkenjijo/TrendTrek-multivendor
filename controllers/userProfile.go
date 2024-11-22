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

func GetProfile(c *fiber.Ctx) error {
	user := new(models.UserProfileResponse)
	// Get the user ID from the context
	userID := c.Locals("user_id")
	// Get the user profile from the database
	if err := database.DB.Model(&models.User{}).Select([]string{"id", "name", "email", "phone_number", "profile_picture"}).First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Return 404 Not Found
			c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found in database"})
		} else {
			// Return 500 Internal Server Error
			c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database Error"})
		}
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Userprofile retrived Succesfully", "data": user})

}

func UpdateProfile(c *fiber.Ctx) error {
	// Get the user ID from the context
	userID := c.Locals("user_id")

	// Get the user profile from the database
	var dbUser models.User
	if err := database.DB.First(&dbUser, userID).Error; err != nil {
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
	email := form.Value["email"]
	phone := form.Value["phone_number"]

	// Update the fields in the database user object
	if len(name) > 0 {
		dbUser.Name = name[0]
	}
	if len(email) > 0 {
		dbUser.Email = email[0]
	}
	if len(phone) > 0 {
		dbUser.PhoneNumber = phone[0]
	}

	// Handle the image upload
	file, err := c.FormFile("profile_picture")
	if err == nil {
		// Generate a unique file name for the image
		timestamp := time.Now().Unix()
		fileName := fmt.Sprintf("%d_%d_%s", int(userID.(float64)), timestamp, file.Filename)

		// Define the path to save the image (adjust the path as necessary)
		filePath := "./uploads/profile_pictures/" + fileName

		// Save the file to the specified directory
		if err := c.SaveFile(file, filePath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save profile picture"})
		}

		// Store the file path in the database
		dbUser.ProfilePicture = fmt.Sprintf("https://jijoshibuukken.website/uploads/profile_pictures/%s", fileName)
	}

	// Save the updated user to the database
	if err := database.DB.Save(&dbUser).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update user profile"})
	}

	// Return 200 OK with the updated user profile
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User profile updated successfully",
		"data": fiber.Map{
			"name":            dbUser.Name,
			"email":           dbUser.Email,
			"phone_number":    dbUser.PhoneNumber,
			"profile_picture": dbUser.ProfilePicture,
		},
	})
}
