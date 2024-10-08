package controllers

import (
	"context"
	"encoding/json"
	"io"
	"os"

	// "os"

	"github.com/Ukkenjijo/trendtrek/database"
	"github.com/Ukkenjijo/trendtrek/models"
	"github.com/Ukkenjijo/trendtrek/utils"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var ClientID = os.Getenv("GOOGLE_CLIENT_ID")
var ClientSecret=os.Getenv("GOOGLE_CLIENT_SECRET")

// Set up the OAuth2 configuration using Google credentials
var googleOauthConfig = &oauth2.Config{
	ClientID:     ClientID,                                            // Get from environment
	ClientSecret: ClientSecret,               // Get from environment
	RedirectURL:  "http://localhost:3000/api/v1/user/google/callback", // The URL to redirect after login
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint:     google.Endpoint,
}

// GoogleLogin redirects the user to the Google OAuth 2.0 login page
func GoogleLogin(c *fiber.Ctx) error {

	// Generate a URL to Google's OAuth 2.0 consent screen
	authURL := googleOauthConfig.AuthCodeURL("hjdfyuhadVFYU6781235")

	// Redirect the user to Google's OAuth 2.0 consent screen
	return c.Redirect(authURL, fiber.StatusTemporaryRedirect)
}

// GoogleCallback handles the callback from Google OAuth 2.0
func GoogleCallback(c *fiber.Ctx) error {
	// Get the authorization code from the URL
	code := c.Query("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Authorization code not found",
		})
	}

	// Exchange the authorization code for an access token
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to exchange token",
		})
	}

	// Use the access token to get the user's profile information
	client := googleOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user info",
		})
	}
	defer resp.Body.Close()

	// Parse the user information
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read response body",
		})
	}

	// Parse the user info JSON into a struct
	var userInfo struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Picture string `json:"picture"`
		Name    string `json:"name"`
	}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to unmarshal user info",
		})
	}

	// Check if the user already exists in your database, if not create a new user
	var user models.User
	if err := database.DB.Where("email = ?", userInfo.Email).First(&user).Error; err != nil {
		// Create a new user if not found
		user = models.User{
			Email:          userInfo.Email,
			Name:           userInfo.Name,
			ProfilePicture: userInfo.Picture,
			Verified:       true,
		}
		database.DB.Create(&user)
	} else {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch users"})
	}

	// Generate a JWT token for the user
	jwtToken, err := utils.GenerateJWT(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate JWT",
		})
	}

	// Return the JWT token to the client
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User authenticated successfully",
		"token":   jwtToken,
		"user":    user,
	})
}
