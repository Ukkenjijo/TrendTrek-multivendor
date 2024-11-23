package controllers

import (
	"fmt"
	"log"
	"time"

	"github.com/Ukkenjijo/trendtrek/database"
	"github.com/Ukkenjijo/trendtrek/models"
	"github.com/Ukkenjijo/trendtrek/utils"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func StoreSignup(c *fiber.Ctx) error {
	// Parse text fields from the multipart form
	storeName := c.FormValue("store_name")
	description := c.FormValue("description")
	address := c.FormValue("address")
	city := c.FormValue("city")
	state := c.FormValue("state")
	country := c.FormValue("country")
	sellerName := c.FormValue("name")
	email := c.FormValue("email")
	phoneNumber := c.FormValue("phone_number")
	password := c.FormValue("password")

	log.Println("Store Name:", storeName)
	log.Println("Description:", description)
	log.Println("Address:", address)
	log.Println("City:", city)
	log.Println("State:", state)
	log.Println("Country:", country)
	log.Println("Seller Name:", sellerName)
	log.Println("Email:", email)
	log.Println("Phone Number:", phoneNumber)
	log.Println("Password:", password)

	// Check if all required fields are provided
	if storeName == "" || description == "" || address == "" || city == "" || state == "" || country == "" ||
		sellerName == "" || email == "" || phoneNumber == "" || password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "All fields are required"})
	}

	// Validate the request input
	// if err := utils.ValidateStruct(password); err != nil {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	// }

	

	// Handle file upload (certificate image)
	file, err := c.FormFile("certificate")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Certificate image is required"})
	}

	// Generate a unique filename for the uploaded certificate
	filename := fmt.Sprintf("%d-%s", time.Now().Unix(), file.Filename)

	// Define the upload directory (you can adjust this based on your environment)
	uploadDir := "./uploads/certificates/"
	
	
    // Store the image URL in the database
    imageURL := fmt.Sprintf("https://jijoshibuukken.website/uploads/certificates/%s", filename)

	// Save the file on the server
	if err := c.SaveFile(file, uploadDir+filename); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save certificate"})
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	newseller := models.User{
		Name:           sellerName,
		Email:          email,
		PhoneNumber:    string(phoneNumber),
		HashedPassword: string(hashedPassword),
		Role:           models.RoleSeller,
	}

	if err := database.DB.Create(&newseller).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user"})
	}

	// Create the new store record with the file path of the uploaded certificate
	newStore := models.Store{
		Name:        storeName,
		Description: description,
		Address:     address,
		City:        city,
		State:       state,
		Country:     country,
		Certificate: imageURL, // Save the file path in the Certificate field
		UserID:      newseller.ID,         // Associate the store with the user
	}

	// Save the store in the database
	if err := database.DB.Create(&newStore).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create store"})
	}

	// Generate OTP and send to user's email
	otp, err := utils.GenerateOTP()
	log.Println(otp)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to send OTP"})
	}
	if err := utils.SendOTPEmail(email, otp); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to send OTP"})
	}

	// Store OTP with expiration of 5 minutes
	utils.StoreOTP(email, otp, 5*time.Minute)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "OTP sent to email"})

}


