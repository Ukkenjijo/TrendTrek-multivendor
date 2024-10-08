package controllers

import (
	"log"
	"time"

	"github.com/Ukkenjijo/trendtrek/database"
	"github.com/Ukkenjijo/trendtrek/models"
	"github.com/Ukkenjijo/trendtrek/utils"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type OTPVerificationRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

func Signup(c *fiber.Ctx) error {
	req := new(models.EmailSignupRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Check if user already exists
	var existingUser models.User
	if err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Email already registered"})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	newUser := models.User{
		Name:           req.Name,
		Email:          req.Email,
		PhoneNumber:    string(req.PhoneNumber),
		HashedPassword: string(hashedPassword),
	}
	if err := database.DB.Create(&newUser).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user"})
	}

	// Generate OTP and send to user's email
	otp, err := utils.GenerateOTP()
	log.Println(otp)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to send OTP"})
	}
	if err := utils.SendOTPEmail(req.Email, otp); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to send OTP"})
	}

	// Store OTP with expiration of 5 minutes
	utils.StoreOTP(req.Email, otp, 5*time.Minute)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "OTP sent to email"})
}

var otpCooldown = make(map[string]time.Time) // To track the cooldown of each email

// ResendOTP handles resending the OTP to the user's email
func ResendOTP(c *fiber.Ctx) error {
	type ResendOTPRequest struct {
		Email string `json:"email"`
	}

	req := new(ResendOTPRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Check if the email is on cooldown
	if lastSentTime, exists := otpCooldown[req.Email]; exists {
		timeSinceLastSent := time.Since(lastSentTime)
		if timeSinceLastSent < 1*time.Minute {
			remainingTime := 1*time.Minute - timeSinceLastSent
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":         "Please wait before requesting another OTP",
				"cooldown_time": remainingTime.Seconds(),
			})
		}
	}

	// Generate a new OTP
	otp, err := utils.GenerateOTP()
	log.Println("Generated OTP:", otp) // Log the generated OTP (for debugging)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate OTP"})
	}

	// Send the OTP via email
	if err := utils.SendOTPEmail(req.Email, otp); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to send OTP"})
	}

	// Store the OTP with an expiration of 5 minutes
	utils.StoreOTP(req.Email, otp, 5*time.Minute)

	// Update the last sent time in cooldown map (start the cooldown)
	otpCooldown[req.Email] = time.Now()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "OTP resent to email"})
}

func VerifyOTP(c *fiber.Ctx) error {
	// Parse incoming request
	req := new(OTPVerificationRequest)
	if err := c.BodyParser(req); err != nil {
		log.Println(c.BodyParser(req))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Verify OTP
	if !utils.VerifyOTP(req.Email, req.OTP) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired OTP"})
	}

	// Find the user by email
	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	// Check if the user is already verified
	if user.Verified {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "User is already verified"})
	}

	// Mark the user as verified
	user.Verified = true

	// Update the user in the database
	if err := database.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update user verification status"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "User verified successfully"})
}
